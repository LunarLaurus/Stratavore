package messaging

import (
	"context"
	"time"

	"github.com/meridian-lex/stratavore/internal/storage"
	"github.com/meridian-lex/stratavore/pkg/types"
	"go.uber.org/zap"
)

// OutboxPublisher polls the outbox table and publishes events
type OutboxPublisher struct {
	db        *storage.PostgresClient
	client    *Client
	interval  time.Duration
	batchSize int
	logger    *zap.Logger
	stopCh    chan struct{}
}

// NewOutboxPublisher creates a new outbox publisher
func NewOutboxPublisher(
	db *storage.PostgresClient,
	client *Client,
	interval time.Duration,
	batchSize int,
	logger *zap.Logger,
) *OutboxPublisher {
	return &OutboxPublisher{
		db:        db,
		client:    client,
		interval:  interval,
		batchSize: batchSize,
		logger:    logger,
		stopCh:    make(chan struct{}),
	}
}

// Start begins polling and publishing
func (p *OutboxPublisher) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.logger.Info("outbox publisher started",
		zap.Duration("interval", p.interval),
		zap.Int("batch_size", p.batchSize))

	for {
		select {
		case <-ticker.C:
			p.processBatch(ctx)
		case <-p.stopCh:
			p.logger.Info("outbox publisher stopped")
			return
		case <-ctx.Done():
			p.logger.Info("outbox publisher context cancelled")
			return
		}
	}
}

// Stop stops the publisher
func (p *OutboxPublisher) Stop() {
	close(p.stopCh)
}

// processBatch retrieves and publishes a batch of outbox entries
func (p *OutboxPublisher) processBatch(ctx context.Context) {
	entries, err := p.db.GetPendingOutboxEntries(ctx, p.batchSize)
	if err != nil {
		p.logger.Error("failed to get pending outbox entries", zap.Error(err))
		return
	}

	if len(entries) == 0 {
		return
	}

	p.logger.Debug("processing outbox batch", zap.Int("count", len(entries)))

	for _, entry := range entries {
		p.processEntry(ctx, entry)
	}
}

// processEntry publishes a single outbox entry
func (p *OutboxPublisher) processEntry(ctx context.Context, entry *types.OutboxEntry) {
	// Check if max attempts exceeded
	if entry.Attempts >= entry.MaxAttempts {
		p.logger.Warn("outbox entry exceeded max attempts",
			zap.Int64("id", entry.ID),
			zap.String("event_type", entry.EventType),
			zap.Int("attempts", entry.Attempts))

		// Could move to DLQ here instead of just logging
		return
	}

	// Try to publish
	err := p.client.Publish(ctx, entry.RoutingKey, entry.Payload)
	if err != nil {
		p.logger.Error("failed to publish outbox entry",
			zap.Int64("id", entry.ID),
			zap.String("event_type", entry.EventType),
			zap.Error(err))

		// Increment attempts and schedule retry with exponential backoff
		errMsg := err.Error()
		if err := p.db.IncrementOutboxAttempts(ctx, entry.ID, errMsg); err != nil {
			p.logger.Error("failed to increment outbox attempts", zap.Error(err))
		}

		return
	}

	// Mark as delivered
	if err := p.db.MarkOutboxDelivered(ctx, entry.ID); err != nil {
		p.logger.Error("failed to mark outbox delivered",
			zap.Int64("id", entry.ID),
			zap.Error(err))
		return
	}

	p.logger.Debug("published outbox entry",
		zap.Int64("id", entry.ID),
		zap.String("event_id", entry.EventID),
		zap.String("event_type", entry.EventType),
		zap.String("routing_key", entry.RoutingKey))
}

// GetStats returns current outbox statistics
func (p *OutboxPublisher) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Could query database for stats like pending count, oldest pending, etc.
	return map[string]interface{}{
		"interval_seconds": p.interval.Seconds(),
		"batch_size":       p.batchSize,
	}, nil
}
