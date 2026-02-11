# gRPC API Reference

This document provides a comprehensive reference for Stratavore's gRPC API.

## Overview

Stratavore exposes its functionality through a gRPC API that the CLI, daemon, and external clients use to communicate. The API follows standard gRPC patterns and supports streaming for real-time updates.

## Connection Details

**Default Endpoint:** `localhost:50051`
**Protocol:** gRPC over HTTP/2
**TLS:** Optional (configurable)
**Authentication:** mTLS or JWT tokens

## Service Definition

### Runner Service

Manages runner lifecycle and operations.

```protobuf
service RunnerService {
  rpc CreateRunner(CreateRunnerRequest) returns (CreateRunnerResponse);
  rpc GetRunner(GetRunnerRequest) returns (GetRunnerResponse);
  rpc ListRunners(ListRunnersRequest) returns (ListRunnersResponse);
  rpc UpdateRunner(UpdateRunnerRequest) returns (UpdateRunnerResponse);
  rpc DeleteRunner(DeleteRunnerRequest) returns (DeleteRunnerResponse);
  rpc StopRunner(StopRunnerRequest) returns (StopRunnerResponse);
  rpc RestartRunner(RestartRunnerRequest) returns (RestartRunnerResponse);
  rpc StreamRunnerEvents(StreamRunnerEventsRequest) returns (stream RunnerEvent);
}
```

#### CreateRunner
Creates a new runner for the specified project.

**Request:**
```protobuf
message CreateRunnerRequest {
  string project_name = 1;
  string model = 2;                    // claude-3-sonnet, claude-3-opus, claude-3-haiku
  double temperature = 3;
  int32 max_tokens = 4;
  map<string, string> metadata = 5;
  string runtime_id = 6;              // Optional: pre-assigned runtime ID
  ResourceQuota quota = 7;            // Optional: custom quota
}
```

**Response:**
```protobuf
message CreateRunnerResponse {
  Runner runner = 1;
  repeated string warnings = 2;
}
```

**Example:**
```go
req := &pb.CreateRunnerRequest{
    ProjectName: "my-project",
    Model:       "claude-3-sonnet",
    Temperature: 0.7,
    MaxTokens:   4096,
    Metadata: map[string]string{
        "environment": "development",
        "user":        "john.doe",
    },
}

resp, err := client.CreateRunner(ctx, req)
if err != nil {
    return fmt.Errorf("failed to create runner: %w", err)
}

fmt.Printf("Created runner: %s\n", resp.Runner.Id)
```

#### GetRunner
Retrieves detailed information about a specific runner.

**Request:**
```protobuf
message GetRunnerRequest {
  string runner_id = 1;
  bool include_metrics = 2;
  bool include_session = 3;
}
```

**Response:**
```protobuf
message GetRunnerResponse {
  Runner runner = 1;
  RunnerMetrics metrics = 2;      // Optional
  Session session = 3;            // Optional
}
```

#### ListRunners
Lists runners with optional filtering.

**Request:**
```protobuf
message ListRunnersRequest {
  string project_name = 1;          // Filter by project
  RunnerStatus status = 2;          // Filter by status
  int32 limit = 3;                  // Max results
  int32 offset = 4;                 // Pagination offset
  bool include_metrics = 5;          // Include performance metrics
  map<string, string> metadata_filter = 6; // Filter by metadata
}
```

**Response:**
```protobuf
message ListRunnersResponse {
  repeated Runner runners = 1;
  int32 total_count = 2;
  bool has_more = 3;
}
```

#### StreamRunnerEvents
Streams real-time runner events.

**Request:**
```protobuf
message StreamRunnerEventsRequest {
  string project_name = 1;          // Optional: filter by project
  repeated EventType event_types = 2; // Optional: filter by event type
  string runner_id = 3;             // Optional: filter by runner
}
```

**Response (Stream):**
```protobuf
message RunnerEvent {
  string event_id = 1;
  EventType event_type = 2;
  google.protobuf.Timestamp timestamp = 3;
  string runner_id = 4;
  string project_name = 5;
  google.protobuf.Any payload = 6;   // Event-specific data
  map<string, string> metadata = 7;
}
```

**Example:**
```go
stream, err := client.StreamRunnerEvents(ctx, &pb.StreamRunnerEventsRequest{
    ProjectName: "my-project",
})
if err != nil {
    return err
}

for {
    event, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        return err
    }
    
    fmt.Printf("Event: %s - %s\n", event.EventType, event.RunnerId)
}
```

### Project Service

Manages projects and their configuration.

```protobuf
service ProjectService {
  rpc CreateProject(CreateProjectRequest) returns (CreateProjectResponse);
  rpc GetProject(GetProjectRequest) returns (GetProjectResponse);
  rpc ListProjects(ListProjectsRequest) returns (ListProjectsResponse);
  rpc UpdateProject(UpdateProjectRequest) returns (UpdateProjectResponse);
  rpc DeleteProject(DeleteProjectRequest) returns (DeleteProjectResponse);
}
```

#### CreateProject
Creates a new project with specified configuration.

**Request:**
```protobuf
message CreateProjectRequest {
  string name = 1;
  string description = 2;
  string path = 3;                    // Working directory
  ResourceQuota quota = 4;
  ProjectConfig config = 5;
  map<string, string> metadata = 6;
}
```

**Response:**
```protobuf
message CreateProjectResponse {
  Project project = 1;
  repeated string warnings = 2;
}
```

#### ListProjects
Lists projects with optional filtering.

**Request:**
```protobuf
message ListProjectsRequest {
  ProjectStatus status = 1;           // Filter by status
  bool include_stats = 2;             // Include usage statistics
  int32 limit = 3;
  int32 offset = 4;
}
```

**Response:**
```protobuf
message ListProjectsResponse {
  repeated Project projects = 1;
  int32 total_count = 2;
  bool has_more = 3;
}
```

### Session Service

Manages conversation sessions.

```protobuf
service SessionService {
  rpc CreateSession(CreateSessionRequest) returns (CreateSessionResponse);
  rpc GetSession(GetSessionRequest) returns (GetSessionResponse);
  rpc ListSessions(ListSessionsRequest) returns (ListSessionsResponse);
  rpc ResumeSession(ResumeSessionRequest) returns (ResumeSessionResponse);
  rpc GetSessionTranscript(GetSessionTranscriptRequest) returns (GetSessionTranscriptResponse);
}
```

#### GetSessionTranscript
Retrieves the full conversation transcript for a session.

**Request:**
```protobuf
message GetSessionTranscriptRequest {
  string session_id = 1;
  bool include_metadata = 2;
  TranscriptFormat format = 3;        // json, text, markdown
}
```

**Response:**
```protobuf
message GetSessionTranscriptResponse {
  string session_id = 1;
  repeated TranscriptMessage messages = 2;
  SessionMetadata metadata = 3;
  int64 total_tokens = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
}

message TranscriptMessage {
  string role = 1;                    // user, assistant, system
  string content = 2;
  google.protobuf.Timestamp timestamp = 3;
  map<string, string> metadata = 4;
}
```

### Daemon Service

Daemon management and health monitoring.

```protobuf
service DaemonService {
  rpc GetStatus(GetStatusRequest) returns (GetStatusResponse);
  rpc GetMetrics(GetMetricsRequest) returns (GetMetricsResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
  rpc Reconcile(ReconcileRequest) returns (ReconcileResponse);
}
```

#### GetStatus
Returns comprehensive daemon and system status.

**Request:**
```protobuf
message GetStatusRequest {
  bool include_detailed = 1;
  ComponentType component = 2;        // Optional: specific component
}
```

**Response:**
```protobuf
message GetStatusResponse {
  DaemonStatus daemon_status = 1;
  map<string, ComponentStatus> components = 2;
  SystemMetrics system_metrics = 3;
  google.protobuf.Timestamp last_updated = 4;
}

message DaemonStatus {
  bool healthy = 1;
  int32 active_runners = 2;
  int32 total_projects = 3;
  google.protobuf.Timestamp uptime = 4;
  string version = 5;
}
```

## Data Types

### Runner
```protobuf
message Runner {
  string id = 1;
  string project_name = 2;
  RunnerStatus status = 3;
  string model = 4;
  double temperature = 5;
  int32 max_tokens = 6;
  string runtime_id = 7;
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
  google.protobuf.Timestamp last_heartbeat = 10;
  map<string, string> metadata = 11;
  ResourceUsage resource_usage = 12;
}

enum RunnerStatus {
  RUNNER_STATUS_UNSPECIFIED = 0;
  RUNNER_STATUS_STARTING = 1;
  RUNNER_STATUS_RUNNING = 2;
  RUNNER_STATUS_PAUSED = 3;
  RUNNER_STATUS_STOPPING = 4;
  RUNNER_STATUS_TERMINATED = 5;
  RUNNER_STATUS_FAILED = 6;
}
```

### Project
```protobuf
message Project {
  string name = 1;
  string description = 2;
  string path = 3;
  ProjectStatus status = 4;
  ResourceQuota quota = 5;
  ProjectConfig config = 6;
  ProjectStatistics statistics = 7;
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
  map<string, string> metadata = 10;
}

message ProjectStatistics {
  int32 active_runners = 1;
  int32 total_runners = 2;
  int64 total_sessions = 3;
  int64 total_tokens_used = 4;
  int64 tokens_used_today = 5;
}
```

### Session
```protobuf
message Session {
  string id = 1;
  string runner_id = 2;
  string project_name = 3;
  SessionStatus status = 4;
  int64 token_count = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
  google.protobuf.Timestamp last_activity = 8;
  string model = 9;
  map<string, string> metadata = 10;
}
```

### Resource Metrics
```protobuf
message ResourceUsage {
  double cpu_percent = 1;
  int64 memory_mb = 2;
  int64 tokens_used = 3;
  google.protobuf.Timestamp last_updated = 4;
}

message ResourceQuota {
  int32 max_concurrent_runners = 1;
  int64 max_memory_mb = 2;
  int32 max_cpu_percent = 3;
  int64 max_tokens_per_day = 4;
}
```

## Authentication

### mTLS Authentication
Configure mutual TLS for client authentication:

```go
creds, err := credentials.NewClientTLSFromFile("client.crt", "client.key", "ca.crt")
if err != nil {
    return err
}

conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
```

### JWT Authentication
Use JWT tokens for authentication:

```go
token := "your-jwt-token"
ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)

resp, err := client.CreateRunner(ctx, req)
```

## Error Handling

### Standard gRPC Status Codes

- **OK (0)**: Success
- **INVALID_ARGUMENT (3)**: Invalid request parameters
- **NOT_FOUND (5)**: Resource not found
- **ALREADY_EXISTS (6)**: Resource already exists
- **PERMISSION_DENIED (7)**: Insufficient permissions
- **RESOURCE_EXHAUSTED (8)**: Quota exceeded
- **FAILED_PRECONDITION (9)**: System not in required state
- **ABORTED (10)**: Operation aborted
- **UNAVAILABLE (14)**: Service unavailable
- **DEADLINE_EXCEEDED (4)**: Operation timeout

### Error Response Format
```protobuf
message ErrorInfo {
  string code = 1;
  string message = 2;
  map<string, string> details = 3;
  string request_id = 4;
}
```

## Streaming Examples

### Real-time Runner Monitoring
```go
func monitorRunners(client pb.RunnerServiceClient, projectName string) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    stream, err := client.StreamRunnerEvents(ctx, &pb.StreamRunnerEventsRequest{
        ProjectName: projectName,
        EventTypes: []pb.EventType{
            pb.EventType_EVENT_TYPE_RUNNER_STARTED,
            pb.EventType_EVENT_TYPE_RUNNER_STOPPED,
            pb.EventType_EVENT_TYPE_RUNNER_FAILED,
        },
    })

    for {
        event, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Printf("Stream error: %v", err)
            break
        }

        switch event.EventType {
        case pb.EventType_EVENT_TYPE_RUNNER_STARTED:
            log.Printf("Runner started: %s", event.RunnerId)
        case pb.EventType_EVENT_TYPE_RUNNER_STOPPED:
            log.Printf("Runner stopped: %s", event.RunnerId)
        case pb.EventType_EVENT_TYPE_RUNNER_FAILED:
            log.Printf("Runner failed: %s", event.RunnerId)
        }
    }
}
```

### Session Transcript Streaming
```go
func streamSessionTranscript(client pb.SessionServiceClient, sessionId string) error {
    req := &pb.GetSessionTranscriptRequest{
        SessionId: sessionId,
        Format:    pb.TranscriptFormat_TRANSCRIPT_FORMAT_JSON,
    }

    resp, err := client.GetSessionTranscript(context.Background(), req)
    if err != nil {
        return err
    }

    for _, msg := range resp.Messages {
        fmt.Printf("[%s] %s: %s\n", 
            msg.Timestamp.Format("2006-01-02 15:04:05"), 
            msg.Role, 
            msg.Content)
    }

    return nil
}
```

## Client Libraries

### Go Client
```go
// Install: go get github.com/meridian/stratavore/pkg/client
import "github.com/meridian/stratavore/pkg/client"

func main() {
    // Create client
    c, err := client.New(&client.Config{
        Address: "localhost:50051",
        TLS: &client.TLSConfig{
            CertFile: "client.crt",
            KeyFile:  "client.key",
            CAFile:   "ca.crt",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    // Create runner
    runner, err := c.Runner().Create(context.Background(), &client.CreateRunnerRequest{
        ProjectName: "my-project",
        Model:       "claude-3-sonnet",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created runner: %s\n", runner.ID)
}
```

### Python Client (Future)
```python
# Future: pip install stratavore-client
from stratavore import Client

def main():
    # Create client
    client = Client(address="localhost:50051")
    
    # Create runner
    runner = client.runner.create(
        project_name="my-project",
        model="claude-3-sonnet"
    )
    
    print(f"Created runner: {runner.id}")
```

---

For more information, see the [Protocol Buffers Documentation](protobuf.md) or [Architecture Guide](../developer/architecture.md).