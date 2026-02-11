package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/meridian/stratavore/internal/notifications"
	"github.com/meridian/stratavore/internal/storage"
	"github.com/meridian/stratavore/pkg/types"
	"go.uber.org/zap"
)

// Manager handles token budget tracking and enforcement
type Manager struct {
	db       *storage.PostgresClient
	notifier *notifications.Client
	logger   *zap.Logger
}

// NewManager creates a new budget manager
func NewManager(db *storage.PostgresClient, notifier *notifications.Client, logger *zap.Logger) *Manager {
	return &Manager{
		db:       db,
		notifier: notifier,
		logger:   logger,
	}
}

// CheckBudget checks if a runner can be launched within budget
func (m *Manager) CheckBudget(ctx context.Context, projectName string, estimatedTokens int64) error {
	// Check global budget
	globalBudget, err := m.db.GetTokenBudget(ctx, "global", "")
	if err == nil && globalBudget != nil {
		if globalBudget.UsedTokens+estimatedTokens > globalBudget.LimitTokens {
			return fmt.Errorf("global token budget exceeded: %d/%d tokens used",
				globalBudget.UsedTokens, globalBudget.LimitTokens)
		}
	}

	// Check project budget
	projectBudget, err := m.db.GetTokenBudget(ctx, "project", projectName)
	if err == nil && projectBudget != nil {
		if projectBudget.UsedTokens+estimatedTokens > projectBudget.LimitTokens {
			return fmt.Errorf("project token budget exceeded: %d/%d tokens used",
				projectBudget.UsedTokens, projectBudget.LimitTokens)
		}
	}

	return nil
}

// RecordUsage records token usage and checks for warnings
func (m *Manager) RecordUsage(ctx context.Context, scope, scopeID string, tokens int64) error {
	// Update usage
	err := m.db.IncrementTokenUsage(ctx, scope, scopeID, tokens)
	if err != nil {
		return fmt.Errorf("record usage: %w", err)
	}

	// Check for warnings
	budget, err := m.db.GetTokenBudget(ctx, scope, scopeID)
	if err != nil {
		return nil // No budget configured
	}

	if budget == nil {
		return nil
	}

	percent := int((float64(budget.UsedTokens) / float64(budget.LimitTokens)) * 100)

	// Send notifications at thresholds
	if percent >= 90 && m.notifier != nil {
		m.notifier.TokenBudgetWarning(fmt.Sprintf("%s:%s", scope, scopeID), percent)
		m.logger.Warn("token budget critical",
			zap.String("scope", scope),
			zap.String("scope_id", scopeID),
			zap.Int("percent", percent))
	} else if percent >= 75 && m.notifier != nil {
		m.notifier.TokenBudgetWarning(fmt.Sprintf("%s:%s", scope, scopeID), percent)
		m.logger.Warn("token budget warning",
			zap.String("scope", scope),
			zap.String("scope_id", scopeID),
			zap.Int("percent", percent))
	}

	return nil
}

// CreateBudget creates a new token budget
func (m *Manager) CreateBudget(ctx context.Context, budget *types.TokenBudget) error {
	err := m.db.CreateTokenBudget(ctx, budget)
	if err != nil {
		return fmt.Errorf("create budget: %w", err)
	}

	m.logger.Info("token budget created",
		zap.String("scope", budget.Scope),
		zap.String("scope_id", budget.ScopeID),
		zap.Int64("limit", budget.LimitTokens),
		zap.String("granularity", budget.PeriodGranularity))

	return nil
}

// RolloverBudgets rolls over expired budgets to new period
func (m *Manager) RolloverBudgets(ctx context.Context) error {
	now := time.Now()

	budgets, err := m.db.GetExpiredBudgets(ctx, now)
	if err != nil {
		return fmt.Errorf("get expired budgets: %w", err)
	}

	for _, budget := range budgets {
		// Calculate new period
		var newStart, newEnd time.Time

		switch budget.PeriodGranularity {
		case "hourly":
			newStart = budget.PeriodEnd
			newEnd = newStart.Add(time.Hour)
		case "daily":
			newStart = budget.PeriodEnd
			newEnd = newStart.Add(24 * time.Hour)
		case "weekly":
			newStart = budget.PeriodEnd
			newEnd = newStart.Add(7 * 24 * time.Hour)
		case "monthly":
			newStart = budget.PeriodEnd
			newEnd = newStart.AddDate(0, 1, 0)
		default:
			continue
		}

		// Create new budget period
		newBudget := &types.TokenBudget{
			Scope:             budget.Scope,
			ScopeID:           budget.ScopeID,
			LimitTokens:       budget.LimitTokens,
			UsedTokens:        0,
			PeriodGranularity: budget.PeriodGranularity,
			PeriodStart:       newStart,
			PeriodEnd:         newEnd,
		}

		err = m.db.CreateTokenBudget(ctx, newBudget)
		if err != nil {
			m.logger.Error("failed to rollover budget",
				zap.String("scope", budget.Scope),
				zap.String("scope_id", budget.ScopeID),
				zap.Error(err))
			continue
		}

		m.logger.Info("budget rolled over",
			zap.String("scope", budget.Scope),
			zap.String("scope_id", budget.ScopeID),
			zap.Time("new_start", newStart),
			zap.Time("new_end", newEnd))
	}

	return nil
}

// GetBudgetStatus returns current budget status
func (m *Manager) GetBudgetStatus(ctx context.Context, scope, scopeID string) (*BudgetStatus, error) {
	budget, err := m.db.GetTokenBudget(ctx, scope, scopeID)
	if err != nil {
		return nil, fmt.Errorf("get budget: %w", err)
	}

	if budget == nil {
		return &BudgetStatus{
			Scope:      scope,
			ScopeID:    scopeID,
			HasBudget:  false,
			Unlimited:  true,
		}, nil
	}

	remaining := budget.LimitTokens - budget.UsedTokens
	if remaining < 0 {
		remaining = 0
	}

	percent := 0
	if budget.LimitTokens > 0 {
		percent = int((float64(budget.UsedTokens) / float64(budget.LimitTokens)) * 100)
	}

	return &BudgetStatus{
		Scope:           scope,
		ScopeID:         scopeID,
		HasBudget:       true,
		Unlimited:       false,
		LimitTokens:     budget.LimitTokens,
		UsedTokens:      budget.UsedTokens,
		RemainingTokens: remaining,
		PercentUsed:     percent,
		PeriodStart:     budget.PeriodStart,
		PeriodEnd:       budget.PeriodEnd,
	}, nil
}

// BudgetStatus represents current budget state
type BudgetStatus struct {
	Scope           string
	ScopeID         string
	HasBudget       bool
	Unlimited       bool
	LimitTokens     int64
	UsedTokens      int64
	RemainingTokens int64
	PercentUsed     int
	PeriodStart     time.Time
	PeriodEnd       time.Time
}

// StartRolloverLoop starts periodic budget rollover
func (m *Manager) StartRolloverLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	m.logger.Info("budget rollover loop started",
		zap.Duration("interval", interval))

	for {
		select {
		case <-ticker.C:
			if err := m.RolloverBudgets(ctx); err != nil {
				m.logger.Error("budget rollover error", zap.Error(err))
			}
		case <-ctx.Done():
			m.logger.Info("budget rollover loop stopped")
			return
		}
	}
}
