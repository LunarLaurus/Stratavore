# Job Management System

This directory contains Stratavore's job tracking system in JSONL (JSON Lines) format. Each line represents a single job with complete metadata.

## File Structure

### `jobs.jsonl`
The primary job database in JSONL format. Each line contains a JSON object with these fields:

```json
{
  "id": "job-YYYY-MM-DD-XXX",
  "title": "Job title",
  "description": "Detailed description of work",
  "status": "pending|in_progress|completed|cancelled|blocked",
  "priority": "low|medium|high|critical",
  "created_at": "ISO 8601 timestamp",
  "updated_at": "ISO 8601 timestamp",
  "assignee": "agent-name or null",
  "labels": ["tag1", "tag2"],
  "estimated_hours": number,
  "actual_hours": number or null,
  "dependencies": ["job-id-1", "job-id-2"],
  "deliverables": ["deliverable1", "deliverable2"]
}
```

### `progress.json`
Real-time progress tracking file that is automatically updated as work progresses.

## Job Status Values

- **pending**: Not started yet
- **in_progress**: Currently being worked on
- **completed**: Successfully finished
- **cancelled**: Cancelled or abandoned
- **blocked**: Blocked by dependencies

## Priority Levels

- **low**: Nice to have, can be deferred
- **medium**: Important but not urgent
- **high**: Important and should be done soon
- **critical**: Urgent, blocks other work

## Working with Jobs

### Adding a New Job
```bash
# Append new job to jobs.jsonl
echo '{"id": "job-2025-02-11-007", "title": "New job", "status": "pending", ...}' >> jobs/jobs.jsonl
```

### Updating Job Status
```bash
# Use job management scripts or manually update
# Progress tracker will handle updates automatically
```

### Querying Jobs
```bash
# List all jobs
cat jobs/jobs.jsonl | jq '.'

# Filter by status
cat jobs/jobs.jsonl | jq 'select(.status == "pending")'

# Filter by assignee
cat jobs/jobs.jsonl | jq 'select(.assignee == "agent-name")'

# Filter by priority
cat jobs/jobs.jsonl | jq 'select(.priority == "high")'
```

## Progress Tracking

The `progress.json` file is automatically updated by the job management system and contains:

- Current active jobs
- Time tracking data
- Completion statistics
- Agent workload distribution

## Integration with Agents

When an agent starts working on a job:

1. Job status changes from "pending" to "in_progress"
2. `assignee` field is set to agent name
3. Progress tracker updates automatically
4. Time tracking begins

When job is completed:

1. Status changes to "completed"
2. `actual_hours` field is populated
3. Deliverables are marked as complete
4. Progress tracker updates statistics

## Dependencies

Jobs can depend on other jobs. The `dependencies` array contains job IDs that must be completed before this job can start.

## Labels

Labels help categorize and filter jobs:
- `documentation`: Documentation-related tasks
- `features`: New feature development
- `bug-fixes`: Bug fixes and improvements
- `testing`: Testing and quality assurance
- `performance`: Performance optimizations
- `security`: Security improvements
- `monitoring`: Monitoring and observability

## Automation

The job system integrates with:
- CI/CD pipelines for automated testing
- Documentation generation
- Release management
- Progress reporting to stakeholders

## File Formats

### JSONL Format
JSONL (JSON Lines) is used for:
- Stream processing
- Easy append operations
- Simple git diff tracking
- Backup and restore capabilities

Each line is a complete JSON object, making it easy to:
- Process line by line
- Filter with standard Unix tools
- Parse with jq or other JSON tools
- Maintain history with git

### Progress Tracking
The progress.json file contains aggregated data for:
- Dashboard displays
- Reporting and analytics
- Workload balancing
- Performance metrics