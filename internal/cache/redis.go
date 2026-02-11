package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/meridian/stratavore/pkg/types"
	"go.uber.org/zap"
)

// RedisCache provides caching for frequently accessed data
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
	ttl    map[string]time.Duration
}

// Config for Redis cache
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(cfg Config, logger *zap.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	cache := &RedisCache{
		client: client,
		logger: logger,
		ttl: map[string]time.Duration{
			"project":      5 * time.Minute,
			"runner":       30 * time.Second,
			"runner_list":  10 * time.Second,
			"project_list": 1 * time.Minute,
			"status":       5 * time.Second,
		},
	}

	logger.Info("redis cache connected", zap.String("addr", client.Options().Addr))
	return cache, nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// GetProject retrieves cached project
func (c *RedisCache) GetProject(ctx context.Context, name string) (*types.Project, error) {
	key := fmt.Sprintf("project:%s", name)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	var project types.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, err
	}

	c.logger.Debug("cache hit", zap.String("key", key))
	return &project, nil
}

// SetProject caches a project
func (c *RedisCache) SetProject(ctx context.Context, project *types.Project) error {
	key := fmt.Sprintf("project:%s", project.Name)
	data, err := json.Marshal(project)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl["project"]).Err()
}

// GetRunner retrieves cached runner
func (c *RedisCache) GetRunner(ctx context.Context, runnerID string) (*types.Runner, error) {
	key := fmt.Sprintf("runner:%s", runnerID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var runner types.Runner
	if err := json.Unmarshal(data, &runner); err != nil {
		return nil, err
	}

	c.logger.Debug("cache hit", zap.String("key", key))
	return &runner, nil
}

// SetRunner caches a runner
func (c *RedisCache) SetRunner(ctx context.Context, runner *types.Runner) error {
	key := fmt.Sprintf("runner:%s", runner.ID)
	data, err := json.Marshal(runner)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl["runner"]).Err()
}

// GetRunnerList retrieves cached runner list for project
func (c *RedisCache) GetRunnerList(ctx context.Context, projectName string) ([]*types.Runner, error) {
	key := fmt.Sprintf("runners:project:%s", projectName)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var runners []*types.Runner
	if err := json.Unmarshal(data, &runners); err != nil {
		return nil, err
	}

	c.logger.Debug("cache hit", zap.String("key", key))
	return runners, nil
}

// SetRunnerList caches runner list for project
func (c *RedisCache) SetRunnerList(ctx context.Context, projectName string, runners []*types.Runner) error {
	key := fmt.Sprintf("runners:project:%s", projectName)
	data, err := json.Marshal(runners)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl["runner_list"]).Err()
}

// InvalidateProject removes project from cache
func (c *RedisCache) InvalidateProject(ctx context.Context, name string) error {
	key := fmt.Sprintf("project:%s", name)
	return c.client.Del(ctx, key).Err()
}

// InvalidateRunner removes runner from cache
func (c *RedisCache) InvalidateRunner(ctx context.Context, runnerID string) error {
	key := fmt.Sprintf("runner:%s", runnerID)
	return c.client.Del(ctx, key).Err()
}

// InvalidateRunnerList removes runner list from cache
func (c *RedisCache) InvalidateRunnerList(ctx context.Context, projectName string) error {
	key := fmt.Sprintf("runners:project:%s", projectName)
	return c.client.Del(ctx, key).Err()
}

// Warm preloads frequently accessed data
func (c *RedisCache) Warm(ctx context.Context, projects []*types.Project, runners []*types.Runner) error {
	pipe := c.client.Pipeline()

	// Cache all projects
	for _, p := range projects {
		data, _ := json.Marshal(p)
		key := fmt.Sprintf("project:%s", p.Name)
		pipe.Set(ctx, key, data, c.ttl["project"])
	}

	// Cache all active runners
	for _, r := range runners {
		data, _ := json.Marshal(r)
		key := fmt.Sprintf("runner:%s", r.ID)
		pipe.Set(ctx, key, data, c.ttl["runner"])
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	c.logger.Info("cache warmed",
		zap.Int("projects", len(projects)),
		zap.Int("runners", len(runners)))

	return nil
}

// GetStats returns cache statistics
func (c *RedisCache) GetStats(ctx context.Context) (*CacheStats, error) {
	info, err := c.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	dbsize, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &CacheStats{
		Keys:     dbsize,
		RawStats: info,
	}, nil
}

// CacheStats contains cache statistics
type CacheStats struct {
	Keys     int64
	RawStats string
}

// Flush clears all cached data (use with caution)
func (c *RedisCache) Flush(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}
