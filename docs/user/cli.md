# CLI Reference

This document provides a complete reference for all Stratavore CLI commands and options.

## Global Options

These options can be used with any command:

```bash
--config string          Path to configuration file (default: ~/.config/stratavore/stratavore.yaml)
--debug                  Enable debug logging
--god                    Enable god mode (bypass restrictions)
--help                   Show help for command
--timeout duration       Command timeout (default: 30s)
--version                Show version information
```

## Commands

### project

Manage projects.

#### `new`
Create a new project.

```bash
stratavore new <project-name> [flags]
```

**Flags:**
```bash
--description string    Project description
--path string           Project directory path (default: current directory)
--max-runners int       Maximum concurrent runners (default: 5)
--memory-limit int      Memory limit in MB (default: 2048)
--token-limit int       Daily token limit (default: 100000)
```

**Examples:**
```bash
# Create project in current directory
stratavore new my-project

# Create with custom settings
stratavore new big-project --max-runners 10 --memory-limit 4096

# Create in specific directory
stratavore new remote-project --path /home/user/projects/remote
```

#### `list`
List all projects.

```bash
stratavore projects [flags]
```

**Flags:**
```bash
--status string     Filter by status (active, idle, archived)
--verbose          Show detailed information
--format string    Output format (table, json, yaml) (default: table)
```

**Examples:**
```bash
# List all projects
stratavore projects

# List active projects with details
stratavore projects --status active --verbose

# Output as JSON
stratavore projects --format json
```

#### `show`
Show detailed information about a project.

```bash
stratavore projects show <project-name> [flags]
```

**Flags:**
```bash
--include-metrics    Include usage metrics
--include-quota     Show quota information
```

**Examples:**
```bash
# Show project details
stratavore projects show my-project

# Show with metrics and quotas
stratavore projects show my-project --include-metrics --include-quota
```

#### `delete`
Delete a project.

```bash
stratavore projects delete <project-name> [flags]
```

**Flags:**
```bash
--force        Skip confirmation prompt
--keep-data    Keep session data (only removes project metadata)
```

**Examples:**
```bash
# Delete project (with confirmation)
stratavore projects delete old-project

# Force delete without confirmation
stratavore projects delete old-project --force
```

### runners

Manage runners.

#### `list`
List all runners.

```bash
stratavore runners [flags]
```

**Flags:**
```bash
--project string       Filter by project name
--status string        Filter by status (starting, running, paused, stopping, terminated, failed)
--verbose             Show detailed information
--format string       Output format (table, json, yaml) (default: table)
--limit int           Limit number of results (default: 50)
```

**Examples:**
```bash
# List all runners
stratavore runners

# List running runners for specific project
stratavore runners --project my-project --status running

# Show detailed information in JSON format
stratavore runners --verbose --format json
```

#### `show`
Show detailed information about a specific runner.

```bash
stratavore runners show <runner-id> [flags]
```

**Flags:**
```bash
--include-logs        Include recent log entries
--include-metrics     Include performance metrics
--include-session     Include session information
```

**Examples:**
```bash
# Show runner details
stratavore runners show runner_abc123

# Show with logs and metrics
stratavore runners show runner_abc123 --include-logs --include-metrics
```

#### `kill`
Stop one or more runners.

```bash
stratavore kill <runner-id|project-name|all> [flags]
```

**Flags:**
```bash
--force              Force kill (don't wait for graceful shutdown)
--timeout duration   Graceful shutdown timeout (default: 30s)
--project           Interpret argument as project name
```

**Examples:**
```bash
# Stop specific runner
stratavore kill runner_abc123

# Stop all runners for a project
stratavore kill my-project --project

# Force kill immediately
stratavore kill runner_abc123 --force

# Stop all runners
stratavore kill all
```

#### `restart`
Restart a runner.

```bash
stratavore restart <runner-id> [flags]
```

**Flags:**
```bash
--timeout duration   Graceful shutdown timeout (default: 30s)
--preserve-session   Keep current session
```

**Examples:**
```bash
# Restart runner
stratavore restart runner_abc123

# Restart keeping session
stratavore restart runner_abc123 --preserve-session
```

### launch

Launch a new runner or attach to existing one.

```bash
stratavore launch <project-name> [flags]
```

**Flags:**
```bash
--model string        Claude model (claude-3-sonnet, claude-3-opus, claude-3-haiku)
--temperature float    Temperature setting (0.0-1.0)
--max-tokens int       Maximum tokens per response
--count int            Number of runners to launch (default: 1)
--attach               Attach to first runner after launch
--no-wait             Don't wait for runner to be ready
```

**Examples:**
```bash
# Launch runner with default settings
stratavore launch my-project

# Launch with specific model and settings
stratavore launch my-project --model claude-3-opus --temperature 0.3

# Launch multiple runners
stratavore launch my-project --count 3

# Launch and attach immediately
stratavore launch my-project --attach
```

### attach

Attach to an existing runner.

```bash
stratavore attach <runner-id> [flags]
```

**Flags:**
```bash
--read-only          Attach in read-only mode
--session string     Attach to specific session
```

**Examples:**
```bash
# Attach to runner
stratavore attach runner_abc123

# Attach in read-only mode
stratavore attach runner_abc123 --read-only
```

### status

Show system status and metrics.

```bash
stratavore status [flags]
```

**Flags:**
```bash
--detailed           Show detailed metrics
--watch              Watch mode (update every 2 seconds)
--format string      Output format (table, json) (default: table)
--component string   Show specific component (database, messaging, metrics)
```

**Examples:**
```bash
# Show basic status
stratavore status

# Show detailed metrics
stratavore status --detailed

# Watch mode (continuous updates)
stratavore status --watch

# Show database component status
stratavore status --component database
```

### sessions

Manage sessions.

#### `list`
List sessions.

```bash
stratavore sessions [flags]
```

**Flags:**
```bash
--project string      Filter by project name
--runner string       Filter by runner ID
--limit int           Limit number of results (default: 20)
--format string       Output format (table, json, yaml) (default: table)
```

**Examples:**
```bash
# List recent sessions
stratavore sessions

# List sessions for project
stratavore sessions --project my-project

# List sessions for specific runner
stratavore sessions --runner runner_abc123
```

#### `show`
Show session details.

```bash
stratavore sessions show <session-id> [flags]
```

**Flags:**
```bash
--include-transcript   Include full conversation transcript
--include-metrics      Include usage metrics
```

**Examples:**
```bash
# Show session details
stratavore sessions show session_xyz789

# Show with transcript
stratavore sessions show session_xyz789 --include-transcript
```

#### `resume`
Resume a session.

```bash
stratavore resume <session-id> [flags]
```

**Flags:**
```bash
--new-runner         Resume in new runner instead of attaching to existing
```

**Examples:**
```bash
# Resume session
stratavore resume session_xyz789

# Resume in new runner
stratavore resume session_xyz789 --new-runner
```

### daemon

Manage the Stratavore daemon.

#### `start`
Start the daemon.

```bash
stratavore daemon start [flags]
```

**Flags:**
```bash
--foreground         Run in foreground (don't daemonize)
--config string      Configuration file path
--log-level string   Log level (debug, info, warn, error) (default: info)
```

**Examples:**
```bash
# Start daemon
stratavore daemon start

# Start in foreground with debug logging
stratavore daemon start --foreground --log-level debug
```

#### `stop`
Stop the daemon.

```bash
stratavore daemon stop [flags]
```

**Flags:**
```bash
--force        Force kill without graceful shutdown
```

**Examples:**
```bash
# Stop daemon gracefully
stratavore daemon stop

# Force stop
stratavore daemon stop --force
```

#### `restart`
Restart the daemon.

```bash
stratavore daemon restart [flags]
```

**Examples:**
```bash
# Restart daemon
stratavore daemon restart
```

#### `status`
Show daemon status.

```bash
stratavore daemon status [flags]
```

**Flags:**
```bash
--detailed       Show detailed health checks
--config         Show configuration information
```

**Examples:**
```bash
# Show daemon status
stratavore daemon status

# Show with detailed health checks
stratavore daemon status --detailed
```

### config

Manage configuration.

#### `show`
Show current configuration.

```bash
stratavore config show [flags]
```

**Flags:**
```bash
--source         Show configuration source (file, env, defaults)
--format string   Output format (yaml, json) (default: yaml)
```

**Examples:**
```bash
# Show configuration
stratavore config show

# Show as JSON with sources
stratavore config show --format json --source
```

#### `validate`
Validate configuration.

```bash
stratavore config validate [flags]
```

**Flags:**
```bash
--config string      Configuration file to validate
```

**Examples:**
```bash
# Validate current configuration
stratavore config validate

# Validate specific file
stratavore config validate --config /path/to/config.yaml
```

### health

Check system health.

```bash
stratavore health [flags]
```

**Flags:**
```bash
--component string   Check specific component (database, messaging, api)
--timeout duration    Health check timeout (default: 10s)
--verbose            Show detailed results
```

**Examples:**
```bash
# Check all components
stratavore health

# Check database only
stratavore health --component database

# Show detailed results
stratavore health --verbose
```

### logs

View logs.

```bash
stratavore logs [flags]
```

**Flags:**
```bash
--component string    Component to view logs for (daemon, agent, all)
--runner string       View logs for specific runner
--follow             Follow log output (like tail -f)
--since string       Show logs since timestamp (e.g., "1h", "30m")
--limit int          Limit number of lines (default: 100)
--level string       Filter by log level (debug, info, warn, error)
```

**Examples:**
```bash
# Show recent daemon logs
stratavore logs

# Follow logs in real-time
stratavore logs --follow

# Show logs for specific runner
stratavore logs --runner runner_abc123

# Show error logs from last hour
stratavore logs --level error --since 1h
```

### metrics

View and manage metrics.

#### `show`
Show current metrics.

```bash
stratavore metrics show [flags]
```

**Flags:**
```bash
--format string     Output format (prometheus, json) (default: prometheus)
--component string  Show metrics for specific component
```

**Examples:**
```bash
# Show Prometheus metrics
stratavore metrics show

# Show as JSON
stratavore metrics show --format json
```

### quotas

Manage resource quotas.

#### `show`
Show current quotas.

```bash
stratavore quotas show [flags]
```

**Flags:**
```bash
--project string    Show quotas for specific project
--global           Show global quotas
```

**Examples:**
```bash
# Show all quotas
stratavore quotas show

# Show for specific project
stratavore quotas show --project my-project
```

#### `set`
Set quota values.

```bash
stratavore quotas set <project-name> <quota-name> <value> [flags]
```

**Examples:**
```bash
# Set max runners for project
stratavore quotas set my-project max-runners 10

# Set memory limit
stratavore quotas set my-project memory-limit 4096
```

### budget

Manage token budgets.

#### `show`
Show budget status.

```bash
stratavore budget show [flags]
```

**Flags:**
```bash
--project string    Show budget for specific project
--scope string       Budget scope (global, project, runner)
```

**Examples:**
```bash
# Show all budgets
stratavore budget show

# Show project budget
stratavore budget show --project my-project
```

#### `set`
Set budget limits.

```bash
stratavore budget set <scope> <period> <amount> [flags]
```

**Examples:**
```bash
# Set daily token limit
stratavore budget set project daily 100000

# Set global hourly limit
stratavore budget set global hourly 1000000
```

### events

Subscribe to and manage events.

#### `subscribe`
Subscribe to event stream.

```bash
stratavore events subscribe [flags]
```

**Flags:**
```bash
--project string      Filter by project name
--type string         Filter by event type (runner, session, system)
--format string       Output format (json, pretty) (default: pretty)
```

**Examples:**
```bash
# Subscribe to all events
stratavore events subscribe

# Subscribe to project events
stratavore events subscribe --project my-project

# Subscribe to runner events only
stratavore events subscribe --type runner
```

### version

Show version information.

```bash
stratavore version [flags]
```

**Flags:**
```bash
--short        Show only version number
--detailed     Show detailed build information
```

**Examples:**
```bash
# Show version
stratavore version

# Show short version
stratavore version --short

# Show detailed build info
stratavore version --detailed
```

## Examples

### Common Workflows

**Multi-project development:**
```bash
# Terminal 1: Start working on project A
stratavore launch project-a --attach

# Terminal 2: Work on project B
stratavore launch project-b --attach

# Terminal 3: Monitor all runners
watch -n 2 stratavore runners --verbose
```

**Resource management:**
```bash
# Check current usage
stratavore status --detailed

# Set project quotas
stratavore quotas set big-project max-runners 10
stratavore quotas set big-project memory-limit 8192

# Monitor token usage
stratavore budget show --project big-project
```

**Troubleshooting:**
```bash
# Check system health
stratavore health --verbose

# View recent logs
stratavore logs --follow --since 10m

# Check specific runner
stratavore runners show runner_abc123 --include-logs
```

---

For more information, see the [User Guide](guide.md) or [Development Guide](../developer/development.md).