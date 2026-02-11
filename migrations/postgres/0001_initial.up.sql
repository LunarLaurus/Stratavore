-- Projects table
CREATE TABLE projects (
    name TEXT PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    status project_status NOT NULL DEFAULT 'idle',
    
    -- Metadata
    description TEXT,
    tags TEXT[],
    
    -- Statistics
    total_runners INTEGER DEFAULT 0,
    active_runners INTEGER DEFAULT 0,
    total_sessions INTEGER DEFAULT 0,
    total_tokens BIGINT DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ,
    archived_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_last_accessed ON projects(last_accessed_at) WHERE status = 'active';

-- Runners table (updated with runtime_type instead of just PID)
CREATE TABLE runners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Runtime identification
    runtime_type runtime_type NOT NULL DEFAULT 'process',
    runtime_id TEXT NOT NULL,  -- PID, container_id, or remote identifier
    node_id TEXT,  -- For remote runners
    
    project_name TEXT NOT NULL,
    project_path TEXT NOT NULL,
    status runner_status NOT NULL DEFAULT 'starting',
    
    -- Configuration
    flags JSONB DEFAULT '[]',
    capabilities JSONB DEFAULT '[]',
    environment JSONB DEFAULT '{}',
    
    -- Session tracking
    session_id TEXT,
    conversation_mode conversation_mode,
    
    -- Resource tracking
    tokens_used BIGINT DEFAULT 0,
    cpu_percent DECIMAL(5,2),
    memory_mb BIGINT,
    
    -- Restart tracking
    restart_attempts INTEGER DEFAULT 0,
    max_restart_attempts INTEGER DEFAULT 3,
    
    -- Timestamps
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_heartbeat TIMESTAMPTZ,
    heartbeat_ttl_seconds INTEGER DEFAULT 30,
    terminated_at TIMESTAMPTZ,
    exit_code INTEGER,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    FOREIGN KEY (project_name) REFERENCES projects(name) ON DELETE CASCADE
);

-- Indexes for active runners
CREATE INDEX idx_runners_status ON runners(status) WHERE status IN ('running', 'paused', 'starting');
CREATE INDEX idx_runners_project ON runners(project_name);
CREATE INDEX idx_runners_heartbeat ON runners(last_heartbeat) WHERE status IN ('running', 'starting');
CREATE INDEX idx_runners_runtime ON runners(runtime_type, runtime_id);
CREATE INDEX idx_runners_stale ON runners(last_heartbeat, status) 
    WHERE status IN ('running', 'starting');

-- Project capabilities/features
CREATE TABLE project_capabilities (
    project_name TEXT NOT NULL,
    capability_name TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    version TEXT,
    config JSONB DEFAULT '{}',
    
    installed_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (project_name, capability_name),
    FOREIGN KEY (project_name) REFERENCES projects(name) ON DELETE CASCADE
);

-- Sessions table (conversation history)
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    runner_id UUID NOT NULL,
    project_name TEXT NOT NULL,
    
    -- Session metadata
    started_at TIMESTAMPTZ DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    last_message_at TIMESTAMPTZ,
    message_count INTEGER DEFAULT 0,
    tokens_used BIGINT DEFAULT 0,
    
    -- Resume capability
    resumable BOOLEAN DEFAULT true,
    resumed_from TEXT,
    
    -- Session data
    summary TEXT,
    context_vector vector(1536),
    
    -- Blob storage reference (for large transcripts)
    transcript_s3_key TEXT,
    transcript_size_bytes BIGINT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    FOREIGN KEY (runner_id) REFERENCES runners(id) ON DELETE CASCADE,
    FOREIGN KEY (project_name) REFERENCES projects(name) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_project ON sessions(project_name);
CREATE INDEX idx_sessions_runner ON sessions(runner_id);
CREATE INDEX idx_sessions_resumable ON sessions(resumable) WHERE resumable = true;
CREATE INDEX idx_sessions_ended ON sessions(ended_at) WHERE ended_at IS NULL;

-- Session blob storage (for transcripts in object store)
CREATE TABLE session_blobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id TEXT NOT NULL,
    blob_type TEXT NOT NULL,  -- 'transcript', 'attachment', etc.
    storage_key TEXT NOT NULL,  -- S3 key or file path
    size_bytes BIGINT,
    checksum TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX idx_session_blobs_session ON session_blobs(session_id);

-- Outbox table for reliable event delivery (transactional outbox pattern)
CREATE TABLE outbox (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered BOOLEAN NOT NULL DEFAULT FALSE,
    delivered_at TIMESTAMPTZ,
    
    -- Event identification
    event_id UUID NOT NULL DEFAULT gen_random_uuid(),
    service_name TEXT NOT NULL DEFAULT 'stratavore',
    aggregate_type TEXT,
    aggregate_id TEXT,
    event_type TEXT NOT NULL,
    
    -- Event data
    payload JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    
    -- Routing
    routing_key TEXT NOT NULL,
    
    -- Retry management
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 5,
    last_attempt_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ,
    error TEXT,
    
    -- Trace context
    trace_id TEXT,
    span_id TEXT
);

CREATE INDEX idx_outbox_undelivered ON outbox(delivered, next_retry_at) 
    WHERE delivered = false;
CREATE INDEX idx_outbox_event_id ON outbox(event_id);
CREATE INDEX idx_outbox_created ON outbox(created_at);

-- Events table (audit log + event sourcing - no FK to preserve immutability)
CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    
    event_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    
    data JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    
    -- Context
    user_id TEXT,
    hostname TEXT,
    trace_id TEXT,
    
    -- Signature for tamper detection (HMAC of event data)
    signature TEXT
);

CREATE INDEX idx_events_entity ON events(entity_type, entity_id);
CREATE INDEX idx_events_type ON events(event_type);
CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_trace ON events(trace_id) WHERE trace_id IS NOT NULL;

-- Token budget tracking
CREATE TABLE token_budgets (
    id SERIAL PRIMARY KEY,
    scope TEXT NOT NULL,  -- 'global', 'project', 'runner'
    scope_id TEXT,
    
    limit_tokens BIGINT NOT NULL,
    used_tokens BIGINT DEFAULT 0,
    
    -- Period management
    period_granularity TEXT NOT NULL,  -- 'hourly', 'daily', 'weekly', 'monthly'
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(scope, scope_id, period_start)
);

CREATE INDEX idx_token_budgets_scope ON token_budgets(scope, scope_id);
CREATE INDEX idx_token_budgets_period ON token_budgets(period_start, period_end);

-- Resource quotas
CREATE TABLE resource_quotas (
    project_name TEXT PRIMARY KEY,
    
    max_concurrent_runners INTEGER DEFAULT 5,
    max_memory_mb BIGINT,
    max_cpu_percent INTEGER,
    max_tokens_per_day BIGINT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    FOREIGN KEY (project_name) REFERENCES projects(name) ON DELETE CASCADE
);

-- Daemon state tracking
CREATE TABLE daemon_state (
    singleton BOOLEAN PRIMARY KEY DEFAULT TRUE,
    daemon_id UUID NOT NULL DEFAULT gen_random_uuid(),
    hostname TEXT NOT NULL,
    version TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Configuration snapshot
    config JSONB DEFAULT '{}',
    
    CONSTRAINT singleton_check CHECK (singleton = true)
);

-- Agent tokens for authentication (join tokens)
CREATE TABLE agent_tokens (
    token TEXT PRIMARY KEY,
    runner_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    
    -- Metadata
    issued_by TEXT,
    agent_version TEXT,
    
    FOREIGN KEY (runner_id) REFERENCES runners(id) ON DELETE CASCADE
);

CREATE INDEX idx_agent_tokens_expires ON agent_tokens(expires_at) 
    WHERE used_at IS NULL AND revoked_at IS NULL;

-- Triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER runners_updated_at BEFORE UPDATE ON runners
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER project_capabilities_updated_at BEFORE UPDATE ON project_capabilities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER token_budgets_updated_at BEFORE UPDATE ON token_budgets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER resource_quotas_updated_at BEFORE UPDATE ON resource_quotas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Function to clean up stale runners (for reconciliation)
CREATE OR REPLACE FUNCTION reconcile_stale_runners(ttl_seconds INTEGER DEFAULT 30)
RETURNS TABLE(runner_id UUID, project_name TEXT) AS $$
BEGIN
    RETURN QUERY
    UPDATE runners
    SET status = 'failed',
        terminated_at = NOW()
    WHERE status IN ('starting', 'running')
      AND last_heartbeat < NOW() - (ttl_seconds || ' seconds')::INTERVAL
    RETURNING id, runners.project_name;
END;
$$ LANGUAGE plpgsql;

-- Function to generate normalized period_start for token budgets
CREATE OR REPLACE FUNCTION normalize_period_start(
    granularity TEXT,
    at_time TIMESTAMPTZ DEFAULT NOW()
) RETURNS TIMESTAMPTZ AS $$
BEGIN
    RETURN CASE granularity
        WHEN 'hourly' THEN date_trunc('hour', at_time)
        WHEN 'daily' THEN date_trunc('day', at_time)
        WHEN 'weekly' THEN date_trunc('week', at_time)
        WHEN 'monthly' THEN date_trunc('month', at_time)
        ELSE date_trunc('day', at_time)
    END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to hash project name for advisory locks
CREATE OR REPLACE FUNCTION hash_project(project_name TEXT)
RETURNS BIGINT AS $$
BEGIN
    RETURN ('x' || substr(md5(project_name), 1, 15))::bit(60)::bigint;
END;
$$ LANGUAGE plpgsql IMMUTABLE;
