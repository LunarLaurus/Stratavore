# Stratavore Custom Tool - Usage Guide

## Tool Overview

Two custom tools have been created for Stratavore agent/task system management:

1. **stratavore.ts** - Full-featured tool (requires Python environment fixes)
2. **stratavore_simple.ts** - Simplified tool with basic functionality (recommended)

## Quick Start Commands

### Check System Status
```
stratavore_simple check-system
```

### List All Agents
```
stratavore_simple list
```

### Count Agents by Status and Personality
```
stratavore_simple count
```

### Find Available Agents
```
stratavore_simple available
```

### Filter by Personality
```
stratavore_simple list cadet
stratavore_simple available specialist
```

### Spawn New Agent
```
stratavore_simple spawn-agent cadet
stratavore_simple spawn-agent debugger "task-123"
```

## Advanced Usage Examples

### Agent Workflow Management

1. **Check current system state:**
   ```
   stratavore_simple check-system
   stratavore_simple count
   ```

2. **Find available workers:**
   ```
   stratavore_simple available
   ```

3. **Spawn specialized team:**
   ```
   stratavore_simple spawn-agent specialist
   stratavore_simple spawn-agent debugger  
   stratavore_simple spawn-agent optimizer
   ```

4. **Monitor team composition:**
   ```
   stratavore_simple list
   stratavore_simple count
   ```

### Task Assignment Strategy

1. **Check for idle cadets for quick tasks:**
   ```
   stratavore_simple available cadet
   ```

2. **Check for senior agents for complex tasks:**
   ```
   stratavore_simple available senior
   ```

3. **Review overall workload:**
   ```
   stratavore_simple count
   ```

## Integration with Other OpenCode Tools

### Complete Workflow Example

```
# 1. Check system health
stratavore_simple check-system

# 2. Review current agents
stratavore_simple list

# 3. Read detailed agent data if needed
read agents/active_agents.jsonl

# 4. Check job queue
read jobs/jobs.jsonl

# 5. Spawn appropriate agent for task
stratavore_simple spawn-agent specialist "database-optimization"

# 6. Monitor agent status changes
stratavore_simple list specialist

# 7. Review agent thoughts/metrics
read agents/agent_todos.jsonl

# 8. Complete task workflow (manual process)
# Use bash to interact with agent manager directly if needed
```

### Data File Monitoring

```
# Monitor agent activity
read agents/agent_todos.jsonl

# Check job assignments  
read jobs/jobs.jsonl

# Review agent metrics
read agents/active_agents.jsonl
```

## Troubleshooting

### If Commands Fail

1. **Check system files:**
   ```
   stratavore_simple check-system
   ```

2. **Verify data integrity:**
   ```
   read agents/active_agents.jsonl
   ```

3. **Use fallback Python commands:**
   ```
   bash python .opencode/tools/stratavore_wrapper.py summary
   ```

### Common Issues

- **File not found**: Use `check-system` to verify all required files exist
- **Encoding errors**: Use simplified tool instead of full version
- **Permission denied**: Check file permissions on agents/ and jobs/ directories

## File Structure Reference

```
Stratavore-git/
├── .opencode/tools/
│   ├── stratavore_simple.ts     # Recommended tool
│   ├── stratavore.ts           # Full-featured (experimental)
│   └── stratavore_wrapper.py   # Python wrapper
├── agents/
│   ├── active_agents.jsonl     # Agent registry
│   ├── agent_todos.jsonl       # Activity log
│   └── agent_manager.py        # Management script
└── jobs/
    └── jobs.jsonl             # Task definitions
```

## API Integration Notes

The Stratavore system also provides REST APIs at `localhost:8080`:

- `/api/agents` - List all agents
- `/api/status` - Full system status
- `/api/spawn-agent` - Create new agent
- `/api/assign-agent` - Assign task to agent

These can be used alongside the custom tool for comprehensive system management.

## Best Practices

1. **Use simple tool first**: Start with `stratavore_simple` for reliability
2. **Monitor system health**: Run `check-system` before operations
3. **Review agent status**: Use `count` to understand workload distribution
4. **Check available agents**: Use `available` before spawning new ones
5. **Read data files directly**: Use `read` tool for detailed analysis
6. **Combine tools**: Use both stratavore tools and standard OpenCode tools

## Migration from Full Tool

If you need advanced functionality not available in the simple tool:

1. Fix Python environment encoding issues
2. Use REST API endpoints directly
3. Use bash commands with the Python wrapper
4. Edit agent files directly using `read`/`write` tools

The simple tool covers the most common use cases and provides a stable foundation for Stratavore system management.