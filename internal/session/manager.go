package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/meridian-lex/stratavore/internal/storage"
	"github.com/meridian-lex/stratavore/pkg/types"
	"go.uber.org/zap"
)

// Manager handles session tracking and resumption
type Manager struct {
	db     *storage.PostgresClient
	logger *zap.Logger
}

// NewManager creates a new session manager
func NewManager(db *storage.PostgresClient, logger *zap.Logger) *Manager {
	return &Manager{
		db:     db,
		logger: logger,
	}
}

// CreateSession creates a new session for a runner
func (m *Manager) CreateSession(ctx context.Context, runnerID, projectName string) (*types.Session, error) {
	sessionID := uuid.New().String()

	session := &types.Session{
		ID:          sessionID,
		RunnerID:    runnerID,
		ProjectName: projectName,
		StartedAt:   time.Now(),
		Resumable:   true,
		CreatedAt:   time.Now(),
	}

	// Insert into database
	err := m.db.CreateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	m.logger.Info("session created",
		zap.String("session_id", sessionID),
		zap.String("runner_id", runnerID),
		zap.String("project", projectName))

	return session, nil
}

// EndSession marks a session as ended
func (m *Manager) EndSession(ctx context.Context, sessionID string) error {
	now := time.Now()

	err := m.db.EndSession(ctx, sessionID, now)
	if err != nil {
		return fmt.Errorf("end session: %w", err)
	}

	m.logger.Info("session ended", zap.String("session_id", sessionID))
	return nil
}

// UpdateSessionMessage records a message in the session
func (m *Manager) UpdateSessionMessage(ctx context.Context, sessionID string, tokensUsed int64) error {
	now := time.Now()

	err := m.db.UpdateSessionMessage(ctx, sessionID, now, tokensUsed)
	if err != nil {
		return fmt.Errorf("update session message: %w", err)
	}

	return nil
}

// GetResumableSessions returns sessions that can be resumed for a project
func (m *Manager) GetResumableSessions(ctx context.Context, projectName string) ([]*types.Session, error) {
	sessions, err := m.db.GetResumableSessions(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("get resumable sessions: %w", err)
	}

	return sessions, nil
}

// ResumeSession prepares a session for resumption
func (m *Manager) ResumeSession(ctx context.Context, sessionID string) (*ResumeInfo, error) {
	session, err := m.db.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	if !session.Resumable {
		return nil, fmt.Errorf("session %s is not resumable", sessionID)
	}

	// Check if runner is still active
	runner, err := m.db.GetRunner(ctx, session.RunnerID)
	if err == nil && runner.Status == types.StatusRunning {
		// Runner still active - can attach directly
		return &ResumeInfo{
			Session:      session,
			RunnerActive: true,
			RunnerID:     runner.ID,
		}, nil
	}

	// Runner dead - need to start new runner with resume flag
	return &ResumeInfo{
		Session:        session,
		RunnerActive:   false,
		NeedsNewRunner: true,
	}, nil
}

// ResumeInfo contains information about session resumption
type ResumeInfo struct {
	Session        *types.Session
	RunnerActive   bool
	RunnerID       string
	NeedsNewRunner bool
}

// MarkSessionNonResumable marks a session as not resumable
func (m *Manager) MarkSessionNonResumable(ctx context.Context, sessionID string, reason string) error {
	err := m.db.MarkSessionNonResumable(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("mark session non-resumable: %w", err)
	}

	m.logger.Info("session marked non-resumable",
		zap.String("session_id", sessionID),
		zap.String("reason", reason))

	return nil
}

// SaveTranscript saves conversation transcript to storage
func (m *Manager) SaveTranscript(ctx context.Context, sessionID string, transcript []byte) error {
	// In production, this would upload to S3/object storage
	// For now, just store metadata

	storageKey := fmt.Sprintf("sessions/%s/transcript.json", sessionID)
	sizeBytes := int64(len(transcript))

	err := m.db.SaveTranscriptMetadata(ctx, sessionID, storageKey, sizeBytes)
	if err != nil {
		return fmt.Errorf("save transcript metadata: %w", err)
	}

	m.logger.Info("transcript saved",
		zap.String("session_id", sessionID),
		zap.Int64("size_bytes", sizeBytes))

	// TODO: Actually upload to S3
	// err = m.s3Client.Upload(storageKey, transcript)

	return nil
}

// LoadTranscript loads conversation transcript from storage
func (m *Manager) LoadTranscript(ctx context.Context, sessionID string) ([]byte, error) {
	session, err := m.db.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	if session.TranscriptS3Key == "" {
		return nil, fmt.Errorf("no transcript available for session %s", sessionID)
	}

	// TODO: Download from S3
	// transcript, err := m.s3Client.Download(session.TranscriptS3Key)

	m.logger.Info("transcript loaded",
		zap.String("session_id", sessionID))

	// Placeholder
	return []byte{}, nil
}

// GetSessionStats returns statistics for a session
func (m *Manager) GetSessionStats(ctx context.Context, sessionID string) (*SessionStats, error) {
	session, err := m.db.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	duration := time.Duration(0)
	if session.EndedAt != nil {
		duration = session.EndedAt.Sub(session.StartedAt)
	} else {
		duration = time.Since(session.StartedAt)
	}

	stats := &SessionStats{
		SessionID:    sessionID,
		MessageCount: session.MessageCount,
		TokensUsed:   session.TokensUsed,
		Duration:     duration,
		Active:       session.EndedAt == nil,
	}

	return stats, nil
}

// SessionStats contains session statistics
type SessionStats struct {
	SessionID    string
	MessageCount int
	TokensUsed   int64
	Duration     time.Duration
	Active       bool
}
