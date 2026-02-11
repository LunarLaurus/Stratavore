package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/meridian-lex/stratavore/pkg/types"
	"go.uber.org/zap"
)

// Manager wraps RedisCache and provides a cache-aside pattern with
// transparent fallback when Redis is unavailable.
type Manager struct {
	redis  *RedisCache
	logger *zap.Logger

	// Metrics
	hits   int64
	misses int64
}

// NewManager creates a CacheManager. If cfg is nil or Redis is unreachable
// the manager operates in pass-through mode (no-op cache).
func NewManager(cfg *Config, logger *zap.Logger) (*Manager, error) {
	if cfg == nil {
		logger.Info("cache disabled: no config provided, operating in pass-through mode")
		return &Manager{logger: logger}, nil
	}

	rc, err := NewRedisCache(*cfg, logger)
	if err != nil {
		logger.Warn("redis unavailable, cache disabled",
			zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
			zap.Error(err))
		// Non-fatal: return a no-op manager
		return &Manager{logger: logger}, nil
	}

	return &Manager{redis: rc, logger: logger}, nil
}

// Enabled reports whether the backing Redis cache is active.
func (m *Manager) Enabled() bool { return m.redis != nil }

// Close shuts down the Redis connection if one exists.
func (m *Manager) Close() error {
	if m.redis == nil {
		return nil
	}
	return m.redis.Close()
}

// ---------------------------------------------------------------------------
// Project helpers
// ---------------------------------------------------------------------------

// GetProject returns a cached project or nil on miss / disabled cache.
func (m *Manager) GetProject(ctx context.Context, name string) *types.Project {
	if m.redis == nil {
		return nil
	}
	p, err := m.redis.GetProject(ctx, name)
	if err != nil {
		m.logger.Debug("cache get error", zap.String("key", "project:"+name), zap.Error(err))
		return nil
	}
	if p != nil {
		m.hits++
	} else {
		m.misses++
	}
	return p
}

// SetProject stores a project in the cache. Errors are logged but not returned.
func (m *Manager) SetProject(ctx context.Context, project *types.Project) {
	if m.redis == nil || project == nil {
		return
	}
	if err := m.redis.SetProject(ctx, project); err != nil {
		m.logger.Debug("cache set error", zap.String("key", "project:"+project.Name), zap.Error(err))
	}
}

// InvalidateProject removes a project entry from the cache.
func (m *Manager) InvalidateProject(ctx context.Context, name string) {
	if m.redis == nil {
		return
	}
	if err := m.redis.InvalidateProject(ctx, name); err != nil {
		m.logger.Debug("cache invalidate error", zap.String("key", "project:"+name), zap.Error(err))
	}
}

// ---------------------------------------------------------------------------
// Runner helpers
// ---------------------------------------------------------------------------

// GetRunner returns a cached runner or nil on miss / disabled cache.
func (m *Manager) GetRunner(ctx context.Context, id string) *types.Runner {
	if m.redis == nil {
		return nil
	}
	r, err := m.redis.GetRunner(ctx, id)
	if err != nil {
		m.logger.Debug("cache get error", zap.String("key", "runner:"+id), zap.Error(err))
		return nil
	}
	if r != nil {
		m.hits++
	} else {
		m.misses++
	}
	return r
}

// SetRunner stores a runner in the cache.
func (m *Manager) SetRunner(ctx context.Context, runner *types.Runner) {
	if m.redis == nil || runner == nil {
		return
	}
	if err := m.redis.SetRunner(ctx, runner); err != nil {
		m.logger.Debug("cache set error", zap.String("key", "runner:"+runner.ID), zap.Error(err))
	}
}

// InvalidateRunner removes a runner entry from the cache.
func (m *Manager) InvalidateRunner(ctx context.Context, id string) {
	if m.redis == nil {
		return
	}
	if err := m.redis.InvalidateRunner(ctx, id); err != nil {
		m.logger.Debug("cache invalidate error", zap.String("key", "runner:"+id), zap.Error(err))
	}
}

// GetRunnerList returns a cached runner list for a project.
func (m *Manager) GetRunnerList(ctx context.Context, projectName string) []*types.Runner {
	if m.redis == nil {
		return nil
	}
	runners, err := m.redis.GetRunnerList(ctx, projectName)
	if err != nil {
		m.logger.Debug("cache get error", zap.String("key", "runners:project:"+projectName), zap.Error(err))
		return nil
	}
	if runners != nil {
		m.hits++
	} else {
		m.misses++
	}
	return runners
}

// SetRunnerList stores a runner list in the cache.
func (m *Manager) SetRunnerList(ctx context.Context, projectName string, runners []*types.Runner) {
	if m.redis == nil {
		return
	}
	if err := m.redis.SetRunnerList(ctx, projectName, runners); err != nil {
		m.logger.Debug("cache set error", zap.String("key", "runners:project:"+projectName), zap.Error(err))
	}
}

// InvalidateRunnerList removes the runner list for a project from the cache.
// Should be called whenever a runner is added, removed, or its status changes.
func (m *Manager) InvalidateRunnerList(ctx context.Context, projectName string) {
	if m.redis == nil {
		return
	}
	if err := m.redis.InvalidateRunnerList(ctx, projectName); err != nil {
		m.logger.Debug("cache invalidate error", zap.String("key", "runners:project:"+projectName), zap.Error(err))
	}
}

// ---------------------------------------------------------------------------
// Warm-up
// ---------------------------------------------------------------------------

// Warm pre-populates the cache with the provided data. Safe to call at
// daemon startup to minimise cold-start latency.
func (m *Manager) Warm(ctx context.Context, projects []*types.Project, runners []*types.Runner) {
	if m.redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := m.redis.Warm(ctx, projects, runners); err != nil {
		m.logger.Warn("cache warm failed", zap.Error(err))
	}
}

// ---------------------------------------------------------------------------
// Stats
// ---------------------------------------------------------------------------

// Stats returns hit/miss counters and, if Redis is active, backend info.
func (m *Manager) Stats(ctx context.Context) map[string]interface{} {
	out := map[string]interface{}{
		"enabled": m.Enabled(),
		"hits":    m.hits,
		"misses":  m.misses,
	}
	if m.hits+m.misses > 0 {
		out["hit_ratio"] = float64(m.hits) / float64(m.hits+m.misses)
	}
	if m.redis != nil {
		if s, err := m.redis.GetStats(ctx); err == nil {
			out["backend_keys"] = s.Keys
		}
	}
	return out
}
