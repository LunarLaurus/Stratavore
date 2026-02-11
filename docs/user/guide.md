# Stratavore User Guide

This guide covers everything you need to know to use Stratavore effectively for managing your AI development workflows.

## Table of Contents

1. [Concepts](#concepts)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Basic Usage](#basic-usage)
5. [Project Management](#project-management)
6. [Runner Management](#runner-management)
7. [Session Management](#session-management)
8. [Monitoring and Status](#monitoring-and-status)
9. [Advanced Features](#advanced-features)
10. [Troubleshooting](#troubleshooting)

## Concepts

### Projects
Projects are the top-level organizational unit in Stratavore. Each project represents a development workspace with its own:
- Runner instances
- Session history
- Resource quotas
- Configuration

### Runners
Runners are individual Claude Code instances managed by Stratavore. Each runner:
- Executes in an isolated environment
- Sends periodic heartbeats to report status
- Maintains session context
- Can be started, stopped, and monitored

### Sessions
Sessions represent conversation history and context within runners. They:
- Store the complete conversation transcript
- Support resumption from any point
- Track token usage
- Can be searched and referenced

### Events
Stratavore uses an event-driven architecture where:
- All state changes emit events
- Events are published to RabbitMQ
- Components react to events asynchronously
- Full audit trail is maintained

## Installation

See the [Quick Start Guide](quick-start.md) for detailed installation instructions.

### Prerequisites
- Go 1.22+ (for building from source)
- PostgreSQL 14+ with pgvector extension
- RabbitMQ 3.12+
- Docker (optional, for infrastructure)

### Binary Installation
```bash
# Install pre-built binaries
sudo make install

# Or install manually
sudo cp bin/stratavore /usr/local/bin/
sudo cp bin/stratavored /usr/local/bin/
sudo cp bin/stratavore-agent /usr/local/bin/
```

### Systemd Service
```bash
# Install and enable systemd service
sudo make systemd-install
sudo systemctl enable stratavored
sudo systemctl start stratavored
```

## Configuration

Stratavore uses a hierarchical configuration system:

### Configuration Files
1. `--config` flag (highest precedence)
2. `~/.config/stratavore/stratavore.yaml`
3. `/etc/stratavore/stratavore.yaml`
4. Environment variables with `STRATAVORE_` prefix

### Example Configuration
```yaml
# Database configuration
database:
  postgresql:
    host: localhost
    port: 5432
    database: stratavore_state
    user: stratavore
    password: "${STRATAVORE_DB_PASSWORD}"  # Environment variable

# RabbitMQ configuration
messaging:
  rabbitmq:
    host: localhost
    port: 5672
    exchange: stratavore.events
    publisher_confirms: true

# Daemon settings
daemon:
  grpc_port: 50051
  heartbeat_interval_seconds: 10
  reconcile_interval_seconds: 30

# Metrics configuration
metrics:
  prometheus:
    enabled: true
    port: 9091

# Security settings
security:
  mtls:
    enabled: false
    cert_file: ""
    key_file: ""
    ca_file: ""
```

### Environment Variables
```bash
export STRATAVORE_DB_PASSWORD="your_password"
export STRATAVORE_RABBITMQ_PASSWORD="rabbitmq_password"
export STRATAVORE_LOG_LEVEL="info"
```

## Basic Usage

### Starting the Daemon
```bash
# Start in foreground (development)
stratavored

# Start as systemd service (production)
sudo systemctl start stratavored
```

### Creating Your First Project
```bash
# Create a new project
stratavore new my-project

# List all projects
stratavore projects
```

### Launching a Runner
```bash
# Smart launch (creates if doesn't exist, attaches if it does)
stratavore my-project

# Launch with specific options
stratavore my-project --god --token-limit 100000
```

## Project Management

### Creating Projects
```bash
# Create with default settings
stratavore new project-name

# Create with specific directory
stratavore new project-name --path /path/to/project

# Create with resource quotas
stratavore new project-name --max-runners 5 --memory-limit 4096
```

### Listing Projects
```bash
# Show all projects
stratavore projects

# Show detailed project info
stratavore projects --verbose

# Filter by status
stratavore projects --status active
```

### Project Configuration
Projects can have per-project configuration in `.stratavore.yaml`:

```yaml
# .stratavore.yaml in project directory
project:
  name: "my-project"
  description: "My awesome AI project"

quotas:
  max_concurrent_runners: 3
  max_memory_mb: 2048
  max_tokens_per_day: 100000

claude:
  model: "claude-3-sonnet"
  temperature: 0.7
  max_tokens: 4096
```

## Runner Management

### Launching Runners
```bash
# Basic launch
stratavore project-name

# Launch multiple runners
stratavore project-name --count 3

# Launch with specific configuration
stratavore project-name --model claude-3-opus --temperature 0.3
```

### Listing Runners
```bash
# Show all active runners
stratavore runners

# Show runners for specific project
stratavore runners --project my-project

# Show detailed runner information
stratavore runners --verbose
```

### Controlling Runners
```bash
# Stop a specific runner
stratavore kill runner-abc123

# Stop all runners for a project
stratavore kill --project my-project

# Restart a runner
stratavore restart runner-abc123

# Attach to a running runner
stratavore attach runner-abc123
```

### Runner Status
Each runner has a status:
- `starting` - Being initialized
- `running` - Active and healthy
- `paused` - Temporarily suspended
- `stopping` - Graceful shutdown in progress
- `terminated` - Stopped (can be restarted)
- `failed` - Crashed or error state

## Session Management

### Viewing Sessions
```bash
# List sessions for a project
stratavore sessions --project my-project

# List all sessions globally
stratavore sessions

# Show session details
stratavore sessions --show session-xyz789
```

### Resuming Sessions
```bash
# Resume last session for project
stratavore my-project

# Resume specific session
stratavore resume session-xyz789
```

### Session Context
Sessions automatically preserve:
- Conversation history
- File context and working directory
- Environment variables
- Tool configurations

## Monitoring and Status

### Global Status
```bash
# Show overall system status
stratavore status

# Show status with metrics
stratavore status --metrics

# Continuous monitoring
watch -n 2 stratavore status
```

### Resource Monitoring
```bash
# Show resource usage
stratavore resources

# Show token usage
stratavore tokens

# Show quotas and limits
stratavore quotas
```

### Metrics and Alerts
```bash
# View Prometheus metrics
curl http://localhost:9091/metrics

# Subscribe to real-time notifications
ntfy subscribe stratavore-status
```

## Advanced Features

### Resource Quotas
Set limits to control resource usage:

```yaml
quotas:
  max_concurrent_runners: 5
  max_memory_mb: 4096
  max_cpu_percent: 80
  max_tokens_per_day: 500000
```

### Token Budgets
Track and limit token usage:

```bash
# Set daily token budget
stratavore budget set daily 100000

# Check current usage
stratavore budget status

# Reset budget
stratavore budget reset
```

### Multi-Project Workflows
Work on multiple projects simultaneously:

```bash
# Terminal 1: Project A
stratavore project-a

# Terminal 2: Project B  
stratavore project-b

# Terminal 3: Project C
stratavore project-c

# Terminal 4: Monitor all
watch -n 2 stratavore status
```

### God Mode
Bypass restrictions for administrative tasks:

```bash
# Launch with god mode
stratavore my-project --god

# Run commands with elevated privileges
stratavore --god status
stratavore --god kill all
```

### Event Subscriptions
Subscribe to specific events:

```bash
# Subscribe to all events
stratavore events subscribe

# Subscribe to project-specific events
stratavore events subscribe --project my-project

# Subscribe to alert events only
stratavore events subscribe --type alerts
```

## Troubleshooting

### Common Issues

#### Daemon Won't Start
```bash
# Check logs
journalctl -u stratavored -f

# Verify database connection
psql -h localhost -U stratavore -d stratavore_state

# Check configuration
stratavored --config-check
```

#### Runner Not Responding
```bash
# Check runner status
stratavore runners

# Manual reconciliation
stratavore reconcile

# Check agent logs
stratavore logs --runner runner-abc123
```

#### Connection Issues
```bash
# Test database connection
stratavore test database

# Test RabbitMQ connection
stratavore test messaging

# Test gRPC connection
stratavore test api
```

### Debug Mode
Enable debug logging for troubleshooting:

```bash
# Run daemon in debug mode
STRATAVORE_LOG_LEVEL=debug stratavored

# Run CLI with debug output
STRATAVORE_LOG_LEVEL=debug stratavore --debug status
```

### Health Checks
Run comprehensive health checks:

```bash
# Full system health check
stratavore health

# Component-specific checks
stratavore health database
stratavore health messaging
stratavore health api
```

## Best Practices

### Project Organization
- Use descriptive project names
- Set appropriate resource quotas
- Regularly review session history
- Clean up unused runners

### Resource Management
- Monitor token usage regularly
- Set reasonable quotas per project
- Use god mode sparingly
- Implement backup strategies

### Security
- Use environment variables for secrets
- Enable TLS in production
- Regularly rotate credentials
- Monitor audit logs

### Performance
- Limit concurrent runners per project
- Use SQLite cache for frequent queries
- Optimize database indexes
- Monitor system metrics

---

For more detailed information, see the [Architecture Documentation](../developer/architecture.md) or [API Reference](../api/).