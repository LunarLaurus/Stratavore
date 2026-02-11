# Agents Guide

This guide provides information for AI coding agents working on the Stratavore codebase.

## Project Structure

See AGENTS.md for comprehensive guidelines and commands.

## Job Management

Stratavore uses a structured job management system to track all work. Jobs are stored in [JSONL format](../../jobs/README.md) in the `jobs/` directory.

### Job Status and Progress

The [progress tracker](../../jobs/progress.json) provides real-time updates on:
- Current active jobs and their status
- Agent workload distribution
- Dependency tracking and blocking issues
- Time tracking and completion metrics

### Working with Jobs

When assigned to a job:
1. Check current status in `jobs/progress.json`
2. Update job status from "pending" to "in_progress"
3. Log work hours regularly
4. Update deliverables as they're completed
5. Mark job as "completed" when finished

### Example Job Update Process
```bash
# Check your assigned jobs
cat jobs/jobs.jsonl | jq 'select(.assignee == "agent-name" and .status == "pending")'

# Update job status (automated through job management system)
# Progress tracker handles updates automatically as you work
```

### Key Job Categories

- **Database & Infrastructure**: Core system components
- **Features & Extensions**: New functionality
- **Testing & Quality**: Automated testing and code quality
- **Documentation**: User and developer documentation
- **UI/UX**: Interactive interfaces and user experience
- **Performance**: Optimization and scaling

## Current Priorities

Based on the [progress tracker](../../jobs/progress.json), current priorities are:

1. **Database schema documentation** (job-2025-02-11-001) - Blocks multiple features
2. **Automated test suite** (job-2025-02-11-005) - Currently in progress
3. **Multi-node runner support** (job-2025-02-11-004) - High priority scaling feature
4. **OpenCode support** (job-2025-02-11-007) - Extensibility requirement
5. **Interactive CLI menu** (job-2025-02-11-008) - User experience improvement

## Development Workflow

See Development Guide for detailed workflow information.

### Agent-Specific Guidelines

1. **Before Starting**: Check `jobs/progress.json` for current status and blockers
2. **During Work**: Regularly update job progress and log hours
3. **Completing Jobs**: Mark as completed and update actual hours
4. **Dependencies**: Check if your job blocks others and communicate delays
5. **Quality**: Ensure tests pass and documentation is updated

### Communication

- Job status updates are reflected in `jobs/progress.json`
- Use job labels to track work categories
- Check dependencies before starting work
- Update blockers immediately when encountered

### Integration with Existing Systems

The job management system integrates with:
- Code reviews and PR validation
- Test suite execution
- Documentation generation
- Release management

## Best Practices for Agents

1. **Always check the progress tracker before starting work**
2. **Update job status promptly when starting/completing tasks**
3. **Log accurate time estimates and actual hours**
4. **Identify and document blockers early**
5. **Communicate dependencies clearly**
6. **Follow the established job status workflow**

## Job Status Workflow

```
pending → in_progress → completed
    ↓         ↓           ↓
  cancelled  blocked   → delivered
```

- **pending**: Ready to start, no blockers
- **in_progress**: Currently being worked on
- **completed**: Work done, ready for review
- **blocked**: Waiting on dependencies
- **cancelled**: No longer needed

## Resource Management

- Each agent should have 1-2 active jobs maximum
- Balance high, medium, and low priority work
- Communicate capacity constraints early
- Update progress at least daily

---

For job-specific questions, check the [jobs directory](../../jobs/) and current progress tracker.