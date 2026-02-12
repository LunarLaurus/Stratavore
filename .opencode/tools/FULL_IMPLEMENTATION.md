# Stratavore Custom Tool - Full Implementation

## Status: DEPLOYED ✅

The complete Stratavore agent/task system management tool is now fully operational. The simplified tool has been deprecated in favor of the full-featured implementation.

## Quick Reference

### Core Commands
```
stratavore spawn cadet "task-id"           # Create new agent
stratavore assign agent-123 "task-id"      # Assign task
stratavore complete agent-123 true "done" # Complete task
stratavore status agent-123 working        # Update status
stratavore list                           # List all agents
stratavore available                      # Show available agents
stratavore summary                        # System overview
stratavore personalities                 # Agent types
stratavore job-status                    # Job queue status
```

### Advanced Operations
```
stratavore spawn-batch specialist 5 "batch-task"  # Create team
stratavore list cadet                         # Filter by type
stratavore available debugger                  # Available by type
```

## Agent Personalities

| Type | Role | Strengths | Best For |
|------|------|-----------|----------|
| **CADET** | Junior Developer | Speed, enthusiasm, learning | Quick fixes, documentation, testing |
| **SENIOR** | Experienced Dev | Architecture, problem solving | System design, refactoring, optimization |
| **SPECIALIST** | Domain Expert | Deep expertise | Databases, security, performance |
| **RESEARCHER** | Investigator | Analysis, exploration | Feature research, prototyping |
| **DEBUGGER** | Troubleshooter | Precision, persistence | Bug fixing, root cause analysis |
| **OPTIMIZER** | Performance Expert | Efficiency, scalability | Optimization, benchmarking |

## Workflow Examples

### 1. Basic Task Assignment
```
# Check current system
stratavore summary

# Find available specialist
stratavore available specialist

# Assign task
stratavore assign specialist_1770833235 "database-optimize"

# Monitor progress
stratavore list specialist
```

### 2. Team Formation
```
# Create specialist team
stratavore spawn-batch specialist 3 "project-alpha"

# Check team composition
stratavore list specialist
stratavore summary

# Monitor workload
stratavore available
```

### 3. Task Completion
```
# Mark task complete
stratavore complete specialist_1770833235 true "Optimized queries 50% faster"

# Update agent status
stratavore status specialist_1770833235 idle "Ready for next task"

# Review system state
stratavore summary
```

### 4. System Monitoring
```
# Overall health check
stratavore summary

# Job queue status
stratavore job-status

# Agent availability
stratavore available

# Detailed breakdown
stratavore list
stratavore personalities
```

## Technical Details

### Data Files
- `agents/active_agents.jsonl` - Live agent registry
- `agents/agent_todos.jsonl` - Activity audit trail
- `jobs/jobs.jsonl` - Task definitions

### Agent States
- `IDLE` - Available for tasks
- `WORKING` - Actively processing
- `SPAWNING` - Initialization phase
- `COMPLETED` - Finished current task
- `ERROR` - Failed or terminated

### Integration Points
- REST API at `localhost:8080`
- Direct file access via `read` tool
- Command-line interface via wrapper

## Error Handling

All commands provide clear error messages:
- Invalid personalities
- Missing agents/tasks
- Unavailable agents
- System issues

## Best Practices

1. **Check availability before spawning**
   ```
   stratavore available cadet || stratavore spawn cadet
   ```

2. **Monitor system load**
   ```
   stratavore summary
   ```

3. **Use appropriate personalities**
   - CADET for simple tasks
   - SPECIALIST for technical challenges
   - DEBUGGER for troubleshooting

4. **Track task completion**
   - Always mark tasks complete
   - Include descriptive notes
   - Update agent status appropriately

## Migration Complete

- ✅ Full tool deployed and tested
- ✅ Simple tool deprecated
- ✅ Encoding issues resolved
- ✅ All commands functional
- ✅ Documentation updated

The Stratavore agent/task system is now fully integrated with OpenCode's custom tool framework and ready for production use.

---

*All systems nominal. Full operational capability achieved.*