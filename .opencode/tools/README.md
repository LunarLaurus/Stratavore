# Stratavore Agent/Task System Tool

## Overview

The `stratavore` custom tool provides comprehensive management of the Stratavore agent/task system through OpenCode. It allows you to spawn agents, assign tasks, monitor status, and manage jobs programmatically.

## Installation

The tool is already installed in `.opencode/tools/stratavore.ts` and ready to use.

## Usage

### Basic Syntax
```
stratavore <action> [parameters]
```

## Actions Available

### 1. **Spawn Agent**
Creates a new agent with specified personality.

**Parameters:**
- `action`: "spawn"
- `personality`: Agent personality (required)
- `task_id`: Optional task to assign immediately

**Personalities:**
- `cadet` - Quick with simple tasks, rapid iteration
- `senior` - Methodical, system design focused
- `specialist` - Domain expert (databases, security, performance)
- `researcher` - Investigative, exploratory analysis
- `debugger` - Troubleshooting and bug fixing specialist
- `optimizer` - Performance and efficiency expert

**Examples:**
```
stratavore spawn cadet
stratavore spawn specialist "fix-database-issue"
stratavore spawn debugger task-123
```

### 2. **Spawn Batch Agents**
Creates multiple agents of the same personality.

**Parameters:**
- `action`: "spawn-batch"
- `personality`: Agent personality (required)
- `count`: Number of agents to spawn (required)
- `task_id`: Optional task for all agents

**Examples:**
```
stratavore spawn-batch cadet 3
stratavore spawn-batch specialist 2 "performance-analysis"
```

### 3. **Assign Task**
Assigns a task to an available agent.

**Parameters:**
- `action`: "assign"
- `agent_id`: Target agent ID (required)
- `task_id`: Task/job ID (required)

**Examples:**
```
stratavore assign cadet_1770826427 job-2025-02-11-016
stratavore assign specialist_1770827623 debug-api-endpoint
```

### 4. **Complete Task**
Marks an agent's current task as completed.

**Parameters:**
- `action`: "complete"
- `agent_id`: Agent ID (required)
- `success`: Completion status (default: true)
- `notes`: Optional completion notes

**Examples:**
```
stratavore complete cadet_1770826427
stratavore complete specialist_1770827623 false "API endpoint not found"
stratavore complete debugger_1770833258 true "Fixed memory leak in parser"
```

### 5. **Update Agent Status**
Manually updates an agent's status and logs a thought.

**Parameters:**
- `action`: "status"
- `agent_id`: Agent ID (required)
- `status`: New status (required)
- `thought`: Optional thought log entry

**Valid Statuses:**
- `idle` - Available for tasks
- `working` - Actively processing
- `paused` - Temporarily suspended
- `completed` - Finished current task
- `error` - Failed or stopped

**Examples:**
```
stratavore status cadet_1770826427 idle "Ready for next task"
stratavore status specialist_1770827623 working "Starting database optimization"
stratavore status debugger_1770833258 error "Unexpected system crash"
```

### 6. **List Agents**
Shows all agents with optional personality filter.

**Parameters:**
- `action`: "list"
- `personality`: Optional personality filter

**Examples:**
```
stratavore list
stratavore list cadet
stratavore list specialist
```

### 7. **Get Summary**
Shows comprehensive system summary with agent statistics.

**Parameters:**
- `action`: "summary"

**Example:**
```
stratavore summary
```

### 8. **Job Status**
Displays job system status and metrics.

**Parameters:**
- `action`: "job-status"
- `job_filter`: Optional status filter (pending, in_progress, completed, cancelled)

**Examples:**
```
stratavore job-status
stratavore job-status pending
stratavore job-status in_progress
```

## Workflow Examples

### Basic Agent Workflow
```
# Spawn a specialist agent
stratavore spawn specialist

# List available agents to get ID
stratavore list specialist

# Assign a task
stratavore assign specialist_1770827623 optimize-database

# Check progress
stratavore summary

# Mark task complete
stratavore complete specialist_1770827623 true "Database optimized, 50% performance improvement"
```

### Parallel Task Distribution
```
# Spawn multiple agents for different tasks
stratavore spawn cadet "write-documentation"
stratavore spawn debugger "fix-bug-123"
stratavore spawn optimizer "improve-performance"

# Check system status
stratavore summary

# Monitor job queue
stratavore job-status pending
```

### Batch Processing
```
# Spawn team of specialists for large project
stratavore spawn-batch specialist 5 "project-alpha"

# Monitor workload distribution
stratavore list
stratavore summary
```

## Error Handling

The tool provides clear error messages for:
- Invalid personalities
- Missing required parameters
- Agent not found
- Agent not available for task assignment
- System execution failures

## Integration with OpenCode

This tool integrates seamlessly with OpenCode's workflow:
- Use `read` to examine agent data files
- Use `bash` for direct system commands
- Use `stratavore` for high-level operations
- All operations maintain data consistency

## Data Persistence

All operations persist to:
- `agents/active_agents.jsonl` - Agent registry
- `agents/agent_todos.jsonl` - Activity audit trail
- `jobs/jobs.jsonl` - Job definitions

Files use JSONL format for easy parsing and corruption resistance.

## Troubleshooting

If operations fail:
1. Check Python 3 is installed and accessible
2. Verify agent_manager.py exists in agents/ directory
3. Ensure proper file permissions
4. Check for stuck spawning agents with `stratavore summary`

## Advanced Usage

Combine with other OpenCode tools:
```
# Read agent thoughts for debugging
read agents/agent_todos.jsonl

# Check system processes
bash ps aux | grep python

# Validate system integrity
stratavore summary && stratavore job-status
```