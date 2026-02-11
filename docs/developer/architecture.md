# Stratavore Architecture

This document provides a deep dive into Stratavore's architecture, design decisions, and implementation details.

## Table of Contents

1. [System Overview](#system-overview)
2. [Component Architecture](#component-architecture)
3. [Data Flow](#data-flow)
4. [Storage Layer](#storage-layer)
5. [Event System](#event-system)
6. [Reliability Guarantees](#reliability-guarantees)
7. [Security Model](#security-model)
8. [Scalability](#scalability)

## System Overview

Stratavore is a control plane for managing Claude Code instances (runners) across multiple projects. It provides:

- **Centralized State Management**: Single source of truth for all runners and sessions
- **Event-Driven Coordination**: RabbitMQ for real-time events and asynchronous processing
- **Reliable Delivery**: Transactional outbox pattern ensures zero message loss
- **Resource Management**: Quota enforcement and token budget tracking
- **Observability**: Comprehensive metrics, logging, and distributed tracing

### Key Components

```
┌─────────────────────────────────────────────────────────────┐
│                     STRATAVORE SYSTEM                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │     CLI      │  │   Daemon     │  │    Agent     │     │
│  │ (stratavore) │  │(stratavored) │  │ (wrapper)    │     │
│  └───────┬──────┘  └───────┬──────┘  └───────┬──────┘     │
│          │                 │                  │            │
│          └────────┬────────┴──────────────────┘            │
│                   │                                        │
│  ┌────────────────▼─────────────────────────────┐          │
│  │           gRPC API Layer                     │          │
│  │   (Authentication, Rate Limiting, Logging)   │          │
│  └────────────────┬─────────────────────────────┘          │
│                   │                                        │
│  ┌────────────────▼─────────────────────────────┐          │
│  │         Runner Manager                       │          │
│  │  - Lifecycle management                      │          │
│  │  - Heartbeat processing                      │          │
│  │  - Reconciliation                            │          │
│  └────────┬──────────────────┬──────────────────┘          │
│           │                  │                             │
│  ┌────────▼────────┐  ┌──────▼──────────┐                 │
│  │ Storage Layer   │  │  Event System   │                 │
│  │  - PostgreSQL   │  │  - RabbitMQ     │                 │
│  │  - SQLite Cache │  │  - Outbox       │                 │
│  │  - Transactions │  │  - Publisher    │                 │
│  └─────────────────┘  └─────────────────┘                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. CLI (stratavore)

**Purpose**: User-facing command-line interface

**Responsibilities**:
- Parse user commands and flags
- Communicate with daemon via gRPC
- Display status information (TUI planned)
- Attach to running sessions (PTY handling)

**Implementation**:
- Built with Cobra for command structure
- Viper for configuration management
- Bubble Tea for interactive TUI (future)

### 2. Daemon (stratavored)

**Purpose**: Control plane process managing all runners

**Responsibilities**:
- Runner lifecycle management (create, monitor, terminate)
- Heartbeat processing and health checking
- Reconciliation of stale runners
- Event publishing via outbox pattern
- gRPC API server for CLI/agent communication
- Metrics exposition for Prometheus

**Key Modules**:

```go
type Daemon struct {
    runnerManager   *RunnerManager      // Manages runners
    storage         *PostgresClient     // Database access
    messaging       *MessagingClient    // RabbitMQ
    outboxPublisher *OutboxPublisher    // Event delivery
    grpcServer      *GRPCServer         // API server
    reconciler      *Reconciler         // Stale runner cleanup
    metricsServer   *MetricsServer      // Prometheus
}
```

**Goroutines**:
1. **Main**: gRPC server event loop
2. **Outbox Publisher**: Polls outbox table, publishes events
3. **Reconciler**: Periodic cleanup of stale runners
4. **Metrics Server**: HTTP server for Prometheus

### 3. Agent (stratavore-agent)

**Purpose**: Wrapper around Claude Code process

**Responsibilities**:
- Launch Claude Code with correct flags
- Send periodic heartbeats to daemon
- Monitor process lifecycle
- Forward signals (SIGTERM, SIGINT)
- Report resource usage (CPU, memory, tokens)

**Lifecycle**:
```
Start -> Launch Claude -> Monitor -> Heartbeat Loop -> Exit
                              │
                              ├─> Process Exit -> Report Exit Code
                              ├─> Signal Received -> Forward -> Wait -> Exit
                              └─> Daemon Request -> Graceful Shutdown
```

## Data Flow

### Launch Flow (Detailed)

```
User                CLI              Daemon           Agent           Claude Code
 │                   │                 │               │                  │
 │─ stratavore proj ─┤                 │               │                  │
 │                   │─── LaunchReq ───┤               │                  │
 │                   │                 │               │                  │
 │                   │            ┌────▼─────┐         │                  │
 │                   │            │Acquire   │         │                  │
 │                   │            │Advisory  │         │                  │
 │                   │            │Lock      │         │                  │
 │                   │            └────┬─────┘         │                  │
 │                   │                 │               │                  │
 │                   │            ┌────▼─────┐         │                  │
 │                   │            │Check     │         │                  │
 │                   │            │Quota     │         │                  │
 │                   │            └────┬─────┘         │                  │
 │                   │                 │               │                  │
 │                   │            ┌────▼─────────┐     │                  │
 │                   │            │BEGIN TX      │     │                  │
 │                   │            │INSERT runner │     │                  │
 │                   │            │INSERT outbox │     │                  │
 │                   │            │COMMIT        │     │                  │
 │                   │            └────┬─────────┘     │                  │
 │                   │                 │               │                  │
 │                   │                 │─── spawn ────┤                  │
 │                   │                 │               │                  │
 │                   │                 │               │─── exec ────────┤
 │                   │                 │               │                  │
 │                   │                 │          ┌────▼──────┐      ┌────▼────┐
 │                   │                 │          │Register   │      │Running  │
 │                   │                 │          │Runtime ID │      │         │
 │                   │                 │          └────┬──────┘      └─────────┘
 │                   │                 │               │                  │
 │                   │            ┌────▼─────────┐     │                  │
 │                   │            │Update        │     │                  │
 │                   │            │RuntimeID     │     │                  │
 │                   │            └────┬─────────┘     │                  │
 │                   │                 │               │                  │
 │                   │◄─── Response ───┤               │                  │
 │◄─ Runner ID ─────┤                 │               │                  │
 │                   │                 │          ┌────▼──────┐           │
 │                   │                 │◄─Heartbeat─┤Every 10s │           │
 │                   │                 │          └───────────┘           │
```

### Heartbeat Processing

```
Agent                     Daemon                  Database
  │                         │                        │
  │─── Heartbeat(metrics)──┤                        │
  │                         │                        │
  │                    ┌────▼─────┐                  │
  │                    │Validate  │                  │
  │                    │Runner    │                  │
  │                    └────┬─────┘                  │
  │                         │                        │
  │                         │─── UPDATE runners ────┤
  │                         │    SET last_heartbeat │
  │                         │        cpu_percent    │
  │                         │        memory_mb      │
  │                         │        tokens_used    │
  │                         │◄──────────────────────┤
  │                         │                        │
  │◄──────── ACK ──────────┤                        │
  │                         │                        │
```

### Event Publishing (Outbox Pattern)

```
Application          Outbox Table         Publisher        RabbitMQ
     │                    │                    │              │
     │─── BEGIN TX ──────┤                    │              │
     │                    │                    │              │
     │─ INSERT runner ────┤                    │              │
     │                    │                    │              │
     │─ INSERT outbox ────┤                    │              │
     │   event            │                    │              │
     │                    │                    │              │
     │─── COMMIT ─────────┤                    │              │
     │                    │                    │              │
     │                    │                    │              │
     │              (Background Poll)          │              │
     │                    │                    │              │
     │                    │◄── SELECT FOR ────┤              │
     │                    │    UPDATE SKIP     │              │
     │                    │    LOCKED          │              │
     │                    │                    │              │
     │                    │─── pending rows ───┤              │
     │                    │                    │              │
     │                    │                    │── Publish ───┤
     │                    │                    │              │
     │                    │                    │◄── Confirm ──┤
     │                    │                    │              │
     │                    │◄── UPDATE outbox ──┤              │
     │                    │    SET delivered   │              │
     │                    │                    │              │
```

## Storage Layer

### PostgreSQL Schema Design

#### Core Tables

**projects**:
- Metadata for development projects
- Statistics (runners, sessions, tokens)
- Status tracking (active, idle, archived)

**runners**:
- Running Claude Code instances
- Runtime identification (process, container, remote)
- Resource metrics (CPU, memory, tokens)
- Heartbeat tracking with TTL
- Restart policy

**sessions**:
- Conversation history
- Resumption metadata
- Token usage tracking
- Vector embeddings for similarity (future)

**outbox**:
- Pending events for reliable delivery
- Retry management with exponential backoff
- Trace context propagation

**events**:
- Immutable audit log
- Event sourcing
- HMAC signatures for tamper detection

#### Indexes

Critical indexes for performance:

```sql
-- Active runner lookups
CREATE INDEX idx_runners_status ON runners(status) 
  WHERE status IN ('running', 'paused', 'starting');

-- Stale runner detection
CREATE INDEX idx_runners_stale ON runners(last_heartbeat, status) 
  WHERE status IN ('running', 'starting');

-- Outbox processing
CREATE INDEX idx_outbox_undelivered ON outbox(delivered, next_retry_at) 
  WHERE delivered = false;

-- Event queries
CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_entity ON events(entity_type, entity_id);
```

### Transaction Patterns

#### Runner Creation with Quota Check

```sql
BEGIN;

-- Acquire project-level advisory lock
SELECT pg_advisory_xact_lock(hash_project('myproject'));

-- Check current active count
SELECT count(*) FROM runners 
WHERE project_name = 'myproject' 
  AND status IN ('starting', 'running');

-- Verify against quota
-- If OK, insert runner and outbox event

COMMIT;
```

This pattern guarantees:
- No race conditions on quota enforcement
- Atomic runner creation + event emission
- Lock released automatically on commit

#### Outbox Processing

```sql
-- Fetch pending events (multiple publishers can run concurrently)
SELECT id, payload, routing_key 
FROM outbox
WHERE delivered = false 
  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
ORDER BY created_at
LIMIT 50
FOR UPDATE SKIP LOCKED;

-- Publish each event to RabbitMQ with confirms
-- On success:
UPDATE outbox SET delivered = true, delivered_at = NOW() WHERE id = $1;

-- On failure:
UPDATE outbox SET 
  attempts = attempts + 1,
  last_attempt_at = NOW(),
  next_retry_at = NOW() + (POWER(2, attempts) || ' seconds')::INTERVAL,
  error = $2
WHERE id = $1;
```

`SKIP LOCKED` allows multiple outbox publishers to run concurrently without blocking.

## Event System

### RabbitMQ Topology

**Exchange**: `stratavore.events` (type: topic, durable)

**Queues**:
- `stratavore.daemon.events` - Consumed by daemon (all events)
- `stratavore.metrics` - Consumed by metrics exporter
- `stratavore.alerts` - Consumed by alert manager → ntfy

**Routing Key Patterns**:
```
runner.started.<project_name>
runner.stopped.<project_name>
runner.failed.<project_name>
runner.heartbeat.<runner_id>

session.created.<project_name>
session.resumed.<session_id>

system.alert.critical
system.alert.warning
system.alert.info

metrics.tokens.<scope>
metrics.resources.<runner_id>
```

### Publisher Confirms

All event publishes use RabbitMQ publisher confirms:

```go
// Enable confirms
channel.Confirm(false)
confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

// Publish with confirm wait
err := channel.PublishWithContext(ctx, exchange, routingKey, false, false, msg)
if err != nil {
    return err
}

// Wait for confirmation
select {
case confirm := <-confirms:
    if !confirm.Ack {
        return errors.New("message not acknowledged")
    }
case <-time.After(5 * time.Second):
    return errors.New("confirmation timeout")
}
```

This guarantees the broker received and persisted the message.

## Reliability Guarantees

### Transactional Outbox Pattern

**Problem**: How to update database AND publish event atomically?

**Solution**:
1. Write both DB record and outbox event in single transaction
2. Background worker polls outbox and publishes to RabbitMQ
3. Mark delivered only after broker confirms receipt
4. Retry with exponential backoff on failure

**Guarantees**:
- At-least-once delivery (may have duplicates)
- No message loss even if RabbitMQ is down
- Eventually consistent

**Idempotency**:
- Each event has unique `event_id`
- Consumers should deduplicate by `event_id`

### Heartbeat and Reconciliation

**Problem**: Detect and cleanup stale/crashed runners

**Solution**:
1. Agents send heartbeat every 10 seconds
2. Daemon reconciler runs every 30 seconds
3. Runners with `last_heartbeat` > TTL (30s) marked failed
4. Failed event published for cleanup

**Edge Cases**:
- Network partition: Runner may be alive but appear stale
- Clock skew: Use server timestamps, not agent timestamps
- Race condition: Reconciler uses WHERE clause with timestamp check

### Database Resilience

**Connection Pooling**:
- pgxpool with min/max connections
- Connection health checks
- Automatic reconnection on failure

**Query Timeouts**:
- All queries use context with timeout
- Prevent hanging on slow queries

**Prepared Statements**:
- Hot paths use prepared statements
- Reduces parse overhead

## Security Model

### Authentication

**gRPC API**:
- Optional mTLS with client certificates
- JWT tokens for session authentication
- API keys for programmatic access

**Agent Join Tokens**:
- Ephemeral tokens generated by daemon
- Short TTL (5 minutes default)
- Single-use, revoked after first use
- Stored in `agent_tokens` table

**Database**:
- Separate user accounts (read-only vs write)
- Credential rotation via secret management
- SSL/TLS connections in production

### Authorization

**Role-Based Access**:
- Admin: Full control (God mode)
- User: Own projects only
- Service: Specific operations

**Audit Logging**:
- All actions logged to `events` table
- Immutable records with HMAC signatures
- User ID and hostname tracked

### Data Protection

**Secrets**:
- Never logged or exposed in errors
- Read from files or environment
- Support for Vault/Kubernetes secrets

**Encryption**:
- TLS for all network traffic
- At-rest encryption for sensitive DB fields (future)

## Scalability

### Horizontal Scaling

**Daemon**:
- Multiple daemon instances can run concurrently
- Coordination via database (advisory locks)
- Each daemon can handle ~1000 runners

**Outbox Publishers**:
- Multiple publishers use `SKIP LOCKED`
- No coordination needed
- Linear scalability

**Database**:
- PostgreSQL read replicas for queries
- Write queries to primary
- Connection pooling prevents exhaustion

### Performance Characteristics

**Latency**:
- Runner launch: ~200ms (database + process spawn)
- Heartbeat processing: <10ms (single UPDATE query)
- State query: <5ms (from SQLite cache)
- Event delivery: ~50ms (outbox + publish + confirm)

**Throughput**:
- Heartbeats: 10,000/sec per daemon
- Runner launches: 100/sec per daemon
- Event publishes: 5,000/sec (limited by RabbitMQ)

**Resource Usage**:
- Daemon: ~50MB memory baseline + 5MB per active runner
- Database: ~1KB per runner record
- RabbitMQ: ~500 bytes per event message

### Resource Management

**Quotas**:
```go
type ResourceQuota struct {
    MaxConcurrentRunners int
    MaxMemoryMB         int64
    MaxCPUPercent       int
    MaxTokensPerDay     int64
}
```

Enforced at:
1. Runner creation (advisory lock + count check)
2. Heartbeat processing (update metrics)
3. Budget reconciliation (periodic check)

**Token Budgets**:
- Scope: global, project, or runner
- Period: hourly, daily, weekly, monthly
- Automatic rollover with `period_start` normalization

---

## Implementation Notes

### Critical Code Paths

**Runner Creation** (`internal/daemon/runner_manager.go`):
- Must use transaction with advisory lock
- Must insert outbox event in same transaction
- Must handle quota exceeded gracefully

**Heartbeat Processing** (`internal/daemon/runner_manager.go`):
- Single UPDATE query for performance
- Non-blocking channel send to avoid backpressure
- Log errors but don't fail on heartbeat loss

**Outbox Publishing** (`internal/messaging/outbox.go`):
- Use `SKIP LOCKED` for concurrency
- Implement exponential backoff: `POWER(2, attempts)`
- Respect `max_attempts` limit

### Testing Strategy

**Unit Tests**:
- Mock database and RabbitMQ clients
- Test business logic in isolation
- Use table-driven tests

**Integration Tests**:
- Use testcontainers for PostgreSQL and RabbitMQ
- Test full workflows (launch, heartbeat, cleanup)
- Verify event delivery

**Chaos Tests**:
- Kill processes randomly
- Network partitions
- Database failover
- RabbitMQ unavailability

---

This architecture provides a solid foundation for reliable, scalable AI development workspace orchestration.
