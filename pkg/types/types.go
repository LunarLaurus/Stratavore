package types

import "time"

// RunnerStatus represents the state of a runner
type RunnerStatus string

const (
	StatusStarting   RunnerStatus = "starting"
	StatusRunning    RunnerStatus = "running"
	StatusPaused     RunnerStatus = "paused"
	StatusTerminated RunnerStatus = "terminated"
	StatusFailed     RunnerStatus = "failed"
)

// RuntimeType represents how the runner is executed
type RuntimeType string

const (
	RuntimeProcess   RuntimeType = "process"
	RuntimeContainer RuntimeType = "container"
	RuntimeRemote    RuntimeType = "remote"
)

// ProjectStatus represents project state
type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "active"
	ProjectIdle     ProjectStatus = "idle"
	ProjectArchived ProjectStatus = "archived"
)

// ConversationMode represents session continuation
type ConversationMode string

const (
	ModeNew      ConversationMode = "new"
	ModeContinue ConversationMode = "continue"
	ModeResume   ConversationMode = "resume"
)

// Runner represents a Meridian Lex instance
type Runner struct {
	ID           string       `json:"id"`
	RuntimeType  RuntimeType  `json:"runtime_type"`
	RuntimeID    string       `json:"runtime_id"`
	NodeID       string       `json:"node_id,omitempty"`
	ProjectName  string       `json:"project_name"`
	ProjectPath  string       `json:"project_path"`
	Status       RunnerStatus `json:"status"`
	Flags        []string     `json:"flags"`
	Capabilities []string     `json:"capabilities"`
	Environment  map[string]string `json:"environment"`
	
	SessionID        string           `json:"session_id,omitempty"`
	ConversationMode ConversationMode `json:"conversation_mode,omitempty"`
	
	TokensUsed       int64   `json:"tokens_used"`
	CPUPercent       float64 `json:"cpu_percent"`
	MemoryMB         int64   `json:"memory_mb"`
	
	RestartAttempts    int `json:"restart_attempts"`
	MaxRestartAttempts int `json:"max_restart_attempts"`
	
	StartedAt      time.Time  `json:"started_at"`
	LastHeartbeat  *time.Time `json:"last_heartbeat,omitempty"`
	HeartbeatTTL   int        `json:"heartbeat_ttl_seconds"`
	TerminatedAt   *time.Time `json:"terminated_at,omitempty"`
	ExitCode       *int       `json:"exit_code,omitempty"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Project represents a development project
type Project struct {
	Name        string        `json:"name"`
	Path        string        `json:"path"`
	Status      ProjectStatus `json:"status"`
	Description string        `json:"description,omitempty"`
	Tags        []string      `json:"tags"`
	
	TotalRunners  int   `json:"total_runners"`
	ActiveRunners int   `json:"active_runners"`
	TotalSessions int   `json:"total_sessions"`
	TotalTokens   int64 `json:"total_tokens"`
	
	CreatedAt      time.Time  `json:"created_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
	ArchivedAt     *time.Time `json:"archived_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Session represents a conversation session
type Session struct {
	ID          string    `json:"id"`
	RunnerID    string    `json:"runner_id"`
	ProjectName string    `json:"project_name"`
	
	StartedAt      time.Time  `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at,omitempty"`
	LastMessageAt  *time.Time `json:"last_message_at,omitempty"`
	MessageCount   int        `json:"message_count"`
	TokensUsed     int64      `json:"tokens_used"`
	
	Resumable    bool   `json:"resumable"`
	ResumedFrom  string `json:"resumed_from,omitempty"`
	Summary      string `json:"summary,omitempty"`
	
	TranscriptS3Key   string `json:"transcript_s3_key,omitempty"`
	TranscriptSizeBytes int64 `json:"transcript_size_bytes,omitempty"`
	
	CreatedAt time.Time `json:"created_at"`
}

// Heartbeat represents agent health status
type Heartbeat struct {
	RunnerID   string       `json:"runner_id"`
	Status     RunnerStatus `json:"status"`
	Timestamp  time.Time    `json:"timestamp"`
	CPUPercent float64      `json:"cpu_percent"`
	MemoryMB   int64        `json:"memory_mb"`
	TokensUsed int64        `json:"tokens_used"`
	SessionID  string       `json:"session_id,omitempty"`
	
	// Agent metadata
	AgentVersion string `json:"agent_version"`
	Hostname     string `json:"hostname"`
}

// Event represents a system event for audit/event sourcing
type Event struct {
	ID         int64                  `json:"id"`
	EventID    string                 `json:"event_id"`
	Timestamp  time.Time              `json:"timestamp"`
	EventType  string                 `json:"event_type"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Data       map[string]interface{} `json:"data"`
	Metadata   map[string]interface{} `json:"metadata"`
	UserID     string                 `json:"user_id,omitempty"`
	Hostname   string                 `json:"hostname,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	Signature  string                 `json:"signature,omitempty"`
}

// OutboxEntry represents an event pending delivery
type OutboxEntry struct {
	ID            int64                  `json:"id"`
	CreatedAt     time.Time              `json:"created_at"`
	Delivered     bool                   `json:"delivered"`
	DeliveredAt   *time.Time             `json:"delivered_at,omitempty"`
	
	EventID       string                 `json:"event_id"`
	ServiceName   string                 `json:"service_name"`
	AggregateType string                 `json:"aggregate_type,omitempty"`
	AggregateID   string                 `json:"aggregate_id,omitempty"`
	EventType     string                 `json:"event_type"`
	
	Payload       map[string]interface{} `json:"payload"`
	Metadata      map[string]interface{} `json:"metadata"`
	RoutingKey    string                 `json:"routing_key"`
	
	Attempts      int        `json:"attempts"`
	MaxAttempts   int        `json:"max_attempts"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty"`
	Error         string     `json:"error,omitempty"`
	
	TraceID string `json:"trace_id,omitempty"`
	SpanID  string `json:"span_id,omitempty"`
}

// LaunchRequest represents a request to start a runner
type LaunchRequest struct {
	ProjectName      string           `json:"project_name"`
	ProjectPath      string           `json:"project_path"`
	Flags            []string         `json:"flags"`
	Capabilities     []string         `json:"capabilities"`
	Environment      map[string]string `json:"environment"`
	ConversationMode ConversationMode `json:"conversation_mode"`
	SessionID        string           `json:"session_id,omitempty"`
	RuntimeType      RuntimeType      `json:"runtime_type"`
}

// ResourceQuota represents project resource limits
type ResourceQuota struct {
	ProjectName         string `json:"project_name"`
	MaxConcurrentRunners int   `json:"max_concurrent_runners"`
	MaxMemoryMB         int64  `json:"max_memory_mb,omitempty"`
	MaxCPUPercent       int    `json:"max_cpu_percent,omitempty"`
	MaxTokensPerDay     int64  `json:"max_tokens_per_day,omitempty"`
}

// TokenBudget represents token usage limits
type TokenBudget struct {
	ID                int       `json:"id"`
	Scope             string    `json:"scope"`
	ScopeID           string    `json:"scope_id,omitempty"`
	LimitTokens       int64     `json:"limit_tokens"`
	UsedTokens        int64     `json:"used_tokens"`
	PeriodGranularity string    `json:"period_granularity"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
}

// DaemonInfo represents daemon state
type DaemonInfo struct {
	DaemonID      string                 `json:"daemon_id"`
	Hostname      string                 `json:"hostname"`
	Version       string                 `json:"version"`
	StartedAt     time.Time              `json:"started_at"`
	LastHeartbeat time.Time              `json:"last_heartbeat"`
	Config        map[string]interface{} `json:"config"`
}

// Metrics represents global metrics
type Metrics struct {
	ActiveRunners  int   `json:"active_runners"`
	ActiveProjects int   `json:"active_projects"`
	TotalSessions  int   `json:"total_sessions"`
	TokensUsed     int64 `json:"tokens_used"`
	TokenLimit     int64 `json:"token_limit"`
}

// ===== Sprint System Types =====

// ModelRegistry represents a registered LLM backend and model.
type ModelRegistry struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	DisplayName          string   `json:"display_name"`
	Backend              string   `json:"backend"`               // 'messages-api' | 'ollama' | 'openrouter' | 'opencode'
	Tier                 string   `json:"tier"`                  // 'lex' | 'haiku45' | 'haiku3' | 'ollama' | 'custom'
	CostPerMillionInput  float64  `json:"cost_per_million_input"`
	CostPerMillionOutput float64  `json:"cost_per_million_output"`
	ContextWindow        int      `json:"context_window"`
	MaxOutputTokens      int      `json:"max_output_tokens"`
	BackendConfig        map[string]interface{} `json:"backend_config"`
	Enabled              bool     `json:"enabled"`
	CreatedAt            time.Time `json:"created_at"`
}

// SprintStatus represents sprint lifecycle state.
type SprintStatus string

const (
	SprintPending   SprintStatus = "pending"
	SprintRunning   SprintStatus = "running"
	SprintCompleted SprintStatus = "completed"
	SprintFailed    SprintStatus = "failed"
	SprintCancelled SprintStatus = "cancelled"
)

// Sprint is a top-level unit of work dispatched by Commander or autonomous Lex.
type Sprint struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	ProjectName string       `json:"project_name,omitempty"`
	Status      SprintStatus `json:"status"`
	CreatedBy   string       `json:"created_by"`
	Tags        []string     `json:"tags"`
	CreatedAt   time.Time    `json:"created_at"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	UpdatedAt   time.Time    `json:"updated_at"`

	Tasks []SprintTask `json:"tasks,omitempty"`
}

// SprintTaskStatus represents sprint task lifecycle state.
type SprintTaskStatus string

const (
	TaskPending   SprintTaskStatus = "pending"
	TaskRunning   SprintTaskStatus = "running"
	TaskCompleted SprintTaskStatus = "completed"
	TaskFailed    SprintTaskStatus = "failed"
	TaskSkipped   SprintTaskStatus = "skipped"
)

// SprintTask is an individual unit of work within a sprint.
type SprintTask struct {
	ID             string           `json:"id"`
	SprintID       string           `json:"sprint_id"`
	SequenceNumber int              `json:"sequence_number"`
	DependsOn      []string         `json:"depends_on"`
	Name           string           `json:"name"`
	Description    string           `json:"description,omitempty"`
	ModelName      string           `json:"model_name"`
	SystemPrompt   string           `json:"system_prompt,omitempty"`
	UserPrompt     string           `json:"user_prompt"`
	MaxTokens      int              `json:"max_tokens"`
	Temperature    float64          `json:"temperature"`
	Status         SprintTaskStatus `json:"status"`
	ResultSummary  string           `json:"result_summary,omitempty"`
	ResultData     map[string]interface{} `json:"result_data,omitempty"`
	TokensInput    int64            `json:"tokens_input"`
	TokensOutput   int64            `json:"tokens_output"`
	CostUSD        float64          `json:"cost_usd"`
	StartedAt      *time.Time       `json:"started_at,omitempty"`
	CompletedAt    *time.Time       `json:"completed_at,omitempty"`
	ErrorMessage   string           `json:"error_message,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// SprintExecution is an audit log entry for a sprint run.
type SprintExecution struct {
	ID              string     `json:"id"`
	SprintID        string     `json:"sprint_id"`
	ExecutedBy      string     `json:"executed_by"`
	Status          string     `json:"status"` // 'running' | 'completed' | 'failed'
	TasksTotal      int        `json:"tasks_total"`
	TasksCompleted  int        `json:"tasks_completed"`
	TasksFailed     int        `json:"tasks_failed"`
	TotalTokensIn   int64      `json:"total_tokens_input"`
	TotalTokensOut  int64      `json:"total_tokens_output"`
	TotalCostUSD    float64    `json:"total_cost_usd"`
	DurationMs      int64      `json:"duration_ms,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}
