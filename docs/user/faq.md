# Frequently Asked Questions

## General Questions

### What is Stratavore?
Stratavore is a comprehensive workspace orchestration system for AI-assisted development. It manages Claude Code sessions as first-class resources, tracks global state, enables multi-project parallel workflows, and integrates with Docker infrastructure for persistence, messaging, and observability.

### What problem does Stratavore solve?
Stratavore solves several common problems in AI development workflows:
- **Session Management**: Resume work instantly from anywhere with full context preservation
- **Resource Management**: Track tokens, manage quotas, prevent overruns
- **Multi-Project Workflows**: Work on multiple AI-assisted projects simultaneously
- **Global Visibility**: Always know what's running where with comprehensive dashboards
- **Reliability**: Transactional outbox pattern ensures no message loss

### Who is Stratavore for?
Stratavore is designed for:
- Individual developers working on multiple AI-assisted projects
- Teams coordinating AI development workflows
- Organizations managing AI development at scale
- DevOps engineers orchestrating AI infrastructure

## Installation and Setup

### How do I install Stratavore?
See the [Quick Start Guide](quick-start.md) for detailed installation instructions. The basic steps are:
1. Clone the repository
2. Install dependencies
3. Build binaries
4. Setup infrastructure (PostgreSQL, RabbitMQ)
5. Configure and start the daemon

### What are the system requirements?
**Minimum:**
- Go 1.22+
- PostgreSQL 14+ with pgvector
- RabbitMQ 3.12+
- 2 CPU cores, 4GB RAM, 20GB storage

**Recommended:**
- 4+ CPU cores, 8GB+ RAM, 100GB+ SSD storage

### Can I run Stratavore on Windows?
Yes! See the [Windows Support Guide](../operations/windows.md) for Windows-specific installation and configuration instructions.

### Do I need Docker?
Docker is optional but recommended for:
- Development infrastructure (PostgreSQL, RabbitMQ)
- Containerized deployments
- Integration testing

## Usage

### How do I start using Stratavore?
```bash
# Start the daemon
stratavored

# Create a project
stratavore new my-project

# Launch a runner
stratavore my-project
```

### Can I work on multiple projects simultaneously?
Yes! Stratavore is designed for multi-project workflows:
```bash
# Terminal 1: Project A
stratavore project-a

# Terminal 2: Project B
stratavore project-b

# Terminal 3: Monitor all
watch -n 2 stratavore status
```

### How do I resume a session?
Sessions are automatically tracked. Simply run:
```bash
stratavore my-project
# Will attach to existing runner if still active
```

### What is "God Mode"?
God mode bypasses certain restrictions for administrative tasks:
```bash
# Launch with god mode
stratavore my-project --god

# Run commands with elevated privileges
stratavore --god kill all
```

## Configuration

### Where is the configuration file?
Stratavore looks for configuration in this order:
1. `--config` flag
2. `~/.config/stratavore/stratavore.yaml`
3. `/etc/stratavore/stratavore.yaml`
4. Environment variables with `STRATAVORE_` prefix

### How do I configure resource quotas?
Set quotas in your configuration or per-project:
```yaml
quotas:
  max_concurrent_runners: 5
  max_memory_mb: 2048
  max_tokens_per_day: 100000
```

Or via CLI:
```bash
stratavore quotas set my-project max-runners 10
```

### How do I secure my deployment?
See the [Deployment Guide](../operations/deployment.md) for security configuration including:
- TLS/SSL setup
- Firewall configuration
- Database security
- API authentication

## Troubleshooting

### The daemon won't start
Check the following:
```bash
# Check logs
journalctl -u stratavored -f

# Test database connection
psql -h localhost -U stratavore -d stratavore_state

# Verify configuration
stratavored --config-check
```

### Runners aren't responding
```bash
# Check runner status
stratavore runners

# Manual reconciliation
stratavore reconcile

# Check network connectivity
stratavore health --verbose
```

### Database connection issues
```bash
# Test database connection
stratavore test database

# Check PostgreSQL is running
pg_isready -h localhost

# Verify credentials
psql -h localhost -U stratavore -d stratavore_state
```

### Events not being published
```bash
# Check outbox table
psql stratavore_state -c "SELECT count(*) FROM outbox WHERE delivered = false;"

# Test RabbitMQ connection
curl http://localhost:15672/api/overview
```

## Performance and Scaling

### How many runners can Stratavore handle?
- **Single node**: ~1000 concurrent runners
- **HA deployment**: 3000+ concurrent runners with proper scaling

### How do I improve performance?
- Optimize PostgreSQL configuration
- Use read replicas for reporting
- Implement proper indexing
- Monitor resource usage

### Can I run multiple daemon instances?
Yes! Multiple daemons can run concurrently, coordinating through the database using advisory locks.

## Monitoring and Observability

### How do I monitor Stratavore?
Stratavore provides built-in Prometheus metrics:
```bash
# View metrics
curl http://localhost:9091/metrics
```

Key metrics to monitor:
- `stratavore_runners_total{status="running|paused|terminated"}`
- `stratavore_tokens_used_total`
- `stratavore_heartbeat_latency_seconds`

### How do I set up alerts?
See the [Deployment Guide](../operations/deployment.md#monitoring-and-observability) for Prometheus alerting rules and Grafana dashboards.

### Can I integrate with existing monitoring?
Yes! Stratavore exports standard Prometheus metrics and structured logs that integrate with existing monitoring stacks.

## Development

### How do I contribute?
See the [Contributing Guide](../developer/contributing.md) for detailed instructions on:
- Setting up development environment
- Code standards and testing
- Submitting changes
- Review process

### What's the architecture?
See the [Architecture Documentation](../developer/architecture.md) for detailed information about:
- System components and data flow
- Storage layer and event system
- Reliability guarantees
- Security model

### How do I run tests?
```bash
# Unit tests
make test

# Integration tests
make test-integration

# With coverage
make test-coverage
```

## Production Deployment

### How do I deploy to production?
See the [Deployment Guide](../operations/deployment.md) for:
- Production configuration
- High availability setup
- Security hardening
- Backup and recovery

### What about high availability?
Stratavore supports HA deployments with:
- Multiple daemon instances behind load balancer
- PostgreSQL replication
- RabbitMQ clustering
- Automated failover

### How do I back up data?
```bash
# Database backup
pg_dump -h localhost -U stratavore -d stratavore_prod > backup.sql

# Automated backup script
# See Deployment Guide for complete backup strategy
```

## Integration

### Does Stratavore work with Claude Code?
Yes! Stratavore is specifically designed to orchestrate Claude Code sessions and provides seamless integration.

### Can I integrate with other AI services?
While optimized for Claude Code, Stratavore's architecture is modular and can be extended to support other AI services through the agent wrapper system.

### How does it integrate with Docker?
Stratavore integrates with Docker for:
- Infrastructure services (PostgreSQL, RabbitMQ)
- Containerized deployments
- Resource isolation and management

## Licensing and Support

### What is the license?
Stratavore is licensed under the MIT License. See the LICENSE file for details.

### How do I get support?
- **Documentation**: Check the [docs](../README.md) folder
- **Issues**: Report bugs on GitHub
- **Discussions**: Join community discussions
- **Contributing**: See the [Contributing Guide](../developer/contributing.md)

### Is there commercial support?
Commercial support options are available. Please contact the maintainers for more information.

---

Have a question not answered here? Please:
1. Check the [documentation](../README.md)
2. Search [GitHub issues](https://github.com/meridian/stratavore/issues)
3. Start a [discussion](https://github.com/meridian/stratavore/discussions)
4. Open a new issue if needed