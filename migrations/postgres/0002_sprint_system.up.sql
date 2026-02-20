-- Sprint system: model registry, sprints, tasks, executions

-- Model registry: replaces hardcoded model tier enum.
-- All models stored in DB — add new models without code changes.
CREATE TABLE model_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    backend TEXT NOT NULL,              -- 'messages-api' | 'ollama' | 'openrouter' | 'opencode'
    tier TEXT NOT NULL,                 -- 'lex' | 'haiku45' | 'haiku3' | 'ollama' | 'custom'
    cost_per_million_input DECIMAL(10,6),
    cost_per_million_output DECIMAL(10,6),
    context_window INTEGER,
    max_output_tokens INTEGER,
    backend_config JSONB DEFAULT '{}',  -- endpoint URL, API key path, model alias, etc.
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_model_registry_backend ON model_registry(backend);
CREATE INDEX idx_model_registry_tier ON model_registry(tier);
CREATE INDEX idx_model_registry_enabled ON model_registry(enabled) WHERE enabled = true;

-- Seed: LLM API models (messages-api backend)
INSERT INTO model_registry (name, display_name, backend, tier, cost_per_million_input, cost_per_million_output, context_window, max_output_tokens) VALUES
    ('claude-opus-4-6',            'Opus 4.6',    'messages-api', 'lex',     15.000000,  75.000000, 200000, 32000),
    ('claude-sonnet-4-5-20250929', 'Sonnet 4.5',  'messages-api', 'lex',      3.000000,  15.000000, 200000, 16000),
    ('claude-haiku-4-5-20251001',  'Haiku 4.5',   'messages-api', 'haiku45',  0.800000,   4.000000, 200000, 16000),
    ('claude-3-haiku-20240307',    'Haiku 3',      'messages-api', 'haiku3',   0.250000,   1.250000, 200000,  4096),
    ('claude-3-5-sonnet-20241022', 'Sonnet 3.5',   'messages-api', 'lex',      3.000000,  15.000000, 200000, 16000),
    ('claude-3-opus-20240229',     'Opus 3',       'messages-api', 'lex',     15.000000,  75.000000, 200000,  4096);

-- Seed: Ollama models (8GB VRAM)
INSERT INTO model_registry (name, display_name, backend, tier, backend_config) VALUES
    ('llama3.2:3b',                 'Llama 3.2 3B',      'ollama', 'ollama', '{"endpoint": "http://localhost:11434"}'),
    ('llama3.1:8b-instruct-q4_0',  'Llama 3.1 8B Q4',   'ollama', 'ollama', '{"endpoint": "http://localhost:11434"}'),
    ('mistral:7b',                  'Mistral 7B',        'ollama', 'ollama', '{"endpoint": "http://localhost:11434"}'),
    ('phi3:mini',                   'Phi-3 Mini',        'ollama', 'ollama', '{"endpoint": "http://localhost:11434"}'),
    ('qwen2.5:7b',                  'Qwen 2.5 7B',       'ollama', 'ollama', '{"endpoint": "http://localhost:11434"}'),
    ('codellama:7b',                'CodeLlama 7B',      'ollama', 'ollama', '{"endpoint": "http://localhost:11434"}'),
    ('granite-embedding:278m',      'Granite Embedding', 'ollama', 'ollama', '{"endpoint": "http://localhost:11434"}');

-- Sprints: top-level unit of work dispatched by Commander or autonomous Lex
CREATE TABLE sprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    project_name TEXT REFERENCES projects(name) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending'   -- 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
        CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    created_by TEXT NOT NULL DEFAULT 'lex',  -- 'lex' | 'commander' | agent id
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sprints_status ON sprints(status);
CREATE INDEX idx_sprints_project ON sprints(project_name);
CREATE INDEX idx_sprints_created ON sprints(created_at DESC);

-- Sprint tasks: individual units of work within a sprint
CREATE TABLE sprint_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    sequence_number INTEGER NOT NULL DEFAULT 0,  -- execution order within sprint
    depends_on UUID[] DEFAULT '{}',              -- task IDs that must complete first
    name TEXT NOT NULL,
    description TEXT,
    model_name TEXT NOT NULL REFERENCES model_registry(name),
    system_prompt TEXT,
    user_prompt TEXT NOT NULL,
    max_tokens INTEGER DEFAULT 4096,
    temperature DECIMAL(3,2) DEFAULT 0.7,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'completed', 'failed', 'skipped')),
    result_summary TEXT,
    result_data JSONB DEFAULT '{}',
    tokens_input BIGINT DEFAULT 0,
    tokens_output BIGINT DEFAULT 0,
    cost_usd DECIMAL(10,6) DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sprint_tasks_sprint ON sprint_tasks(sprint_id);
CREATE INDEX idx_sprint_tasks_status ON sprint_tasks(status);
CREATE INDEX idx_sprint_tasks_sequence ON sprint_tasks(sprint_id, sequence_number);
CREATE INDEX idx_sprint_tasks_model ON sprint_tasks(model_name);

-- Sprint executions: audit log of sprint runs (a sprint may be re-run)
CREATE TABLE sprint_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    executed_by TEXT NOT NULL DEFAULT 'lex',
    status TEXT NOT NULL DEFAULT 'running'
        CHECK (status IN ('running', 'completed', 'failed')),
    tasks_total INTEGER DEFAULT 0,
    tasks_completed INTEGER DEFAULT 0,
    tasks_failed INTEGER DEFAULT 0,
    total_tokens_input BIGINT DEFAULT 0,
    total_tokens_output BIGINT DEFAULT 0,
    total_cost_usd DECIMAL(10,6) DEFAULT 0,
    duration_ms BIGINT,
    notes TEXT,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_sprint_executions_sprint ON sprint_executions(sprint_id);
CREATE INDEX idx_sprint_executions_status ON sprint_executions(status);
