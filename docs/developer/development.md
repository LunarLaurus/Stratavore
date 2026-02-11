# Development Guide

This guide covers setting up a development environment for working on Stratavore.

## Prerequisites

- Go 1.22 or later
- PostgreSQL 14+ with pgvector extension
- RabbitMQ 3.12+
- Docker and Docker Compose
- Git
- Make

## Development Setup

### 1. Clone Repository
```bash
git clone https://github.com/meridian/stratavore.git
cd stratavore
```

### 2. Install Dependencies
```bash
# Download Go dependencies
make deps

# Install development tools
make install-tools
```

### 3. Setup Development Infrastructure
```bash
# Start development services
make dev-up

# This starts:
# - PostgreSQL with test database
# - RabbitMQ with management UI
# - Redis for caching (if needed)
# - Prometheus for metrics
# - Grafana for dashboards
```

### 4. Run Database Migrations
```bash
# Run all migrations
make migration-up

# Or run specific migration
make migration-up VERSION=0002
```

### 5. Build and Test
```bash
# Build all binaries
make build

# Run tests
make test

# Run integration tests
make test-integration
```

## Development Workflow

### Code Organization

```
Stratavore/
├── cmd/                    # Main entry points
│   ├── cli/               # CLI binary
│   ├── daemon/            # Daemon binary
│   └── agent/             # Agent binary
├── internal/              # Private application code
│   ├── cli/               # CLI implementation
│   ├── daemon/            # Daemon implementation
│   ├── agent/             # Agent implementation
│   ├── storage/           # Database operations
│   ├── messaging/         # RabbitMQ operations
│   ├── metrics/           # Prometheus metrics
│   └── config/            # Configuration management
├── pkg/                   # Public/shared packages
│   ├── api/               # gRPC definitions
│   ├── types/             # Common types
│   └── utils/             # Utility functions
├── migrations/            # Database migrations
├── configs/               # Configuration files
├── scripts/               # Build and deployment scripts
└── test/                  # Test utilities
```

### Making Changes

1. **Create a feature branch**
```bash
git checkout -b feature/my-new-feature
```

2. **Make your changes**
- Follow existing code patterns
- Add tests for new functionality
- Update documentation as needed

3. **Run tests and linting**
```bash
make test
make lint
make format
```

4. **Run integration tests**
```bash
make test-integration
```

5. **Commit and push**
```bash
git add .
git commit -m "feat: add my new feature"
git push origin feature/my-new-feature
```

## Development Commands

### Building
```bash
# Build all binaries
make build

# Build specific component
make build-cli
make build-daemon
make build-agent

# Build with race detector
make build-race
```

### Testing
```bash
# Run all tests
make test

# Run tests for specific package
go test -v ./internal/storage/...

# Run with coverage
make test-coverage

# Run integration tests (requires infrastructure)
make test-integration

# Run specific test
go test -v ./internal/storage -run TestCreateRunner
```

### Code Quality
```bash
# Run all linters
make lint

# Format code
make format

# Check for security issues
gosec ./...

# Run static analysis
staticcheck ./...
```

### Database Operations
```bash
# Create new migration
make migration-create NAME=add_new_table

# Run migrations
make migration-up

# Rollback migration
make migration-down

# Reset database
make migration-reset
```

### Development Services
```bash
# Start development services
make dev-up

# Stop development services
make dev-down

# View logs
make dev-logs

# Restart services
make dev-restart
```

## Debugging

### Debugging the Daemon
```bash
# Run daemon in debug mode
go run cmd/daemon/main.go --debug

# Run with delve debugger
dlv debug cmd/daemon/main.go

# Add debug logging
export STRATAVORE_LOG_LEVEL=debug
go run cmd/daemon/main.go
```

### Debugging the CLI
```bash
# Run CLI with debug output
go run cmd/cli/main.go --debug status

# Test against local daemon
export STRATAVORE_GRPC_ADDRESS=localhost:50051
go run cmd/cli/main.go status
```

### Database Debugging
```bash
# Connect to development database
psql -h localhost -U stratavore -d stratavore_state_dev

# View recent events
SELECT * FROM events ORDER BY timestamp DESC LIMIT 10;

# Check runner status
SELECT * FROM runners WHERE status = 'running';
```

### Message Queue Debugging
```bash
# Access RabbitMQ management UI
open http://localhost:15672

# View queues
curl -u guest:guest http://localhost:15672/api/queues

# View exchanges
curl -u guest:guest http://localhost:15672/api/exchanges
```

## Testing Strategy

### Unit Tests
- Mock external dependencies (database, RabbitMQ)
- Test business logic in isolation
- Use table-driven tests for multiple scenarios
- Target 80%+ code coverage

### Integration Tests
- Use testcontainers for real services
- Test complete workflows
- Verify event delivery
- Test error scenarios

### Performance Tests
- Benchmark critical paths
- Load test with simulated runners
- Memory profiling
- Database query optimization

### Test Examples

#### Unit Test
```go
func TestRunnerManager_CreateRunner(t *testing.T) {
    // Setup
    db := mock.NewDatabase(t)
    rm := NewRunnerManager(db, config.DefaultConfig())
    
    // Test
    runner, err := rm.CreateRunner(context.Background(), &CreateRunnerRequest{
        ProjectName: "test-project",
        Model:       "claude-3-sonnet",
    })
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, "test-project", runner.ProjectName)
    assert.Equal(t, RunnerStatusStarting, runner.Status)
}
```

#### Integration Test
```go
func TestRunnerLifecycle_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test infrastructure
    ctx := context.Background()
    infra := testutils.NewInfrastructure(t)
    defer infra.Cleanup()
    
    // Test complete workflow
    client := infra.NewClient()
    
    // Create runner
    runner, err := client.CreateRunner(ctx, "test-project")
    require.NoError(t, err)
    
    // Wait for runner to be healthy
    assert.Eventually(t, func() bool {
        r, _ := client.GetRunner(ctx, runner.ID)
        return r.Status == RunnerStatusRunning
    }, 30*time.Second, time.Second)
    
    // Stop runner
    err = client.StopRunner(ctx, runner.ID)
    require.NoError(t, err)
}
```

## Code Style Guidelines

### Formatting
- Use `gofmt` for basic formatting
- Use `goimports` for import organization
- Max line length: 100 characters
- Use meaningful variable names

### Import Organization
```go
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // External dependencies
    "github.com/spf13/cobra"
    "go.uber.org/zap"
    
    // Internal packages
    "github.com/meridian/stratavore/internal/storage"
    "github.com/meridian/stratavore/pkg/types"
)
```

### Error Handling
```go
// Always wrap errors
if err != nil {
    return fmt.Errorf("failed to create runner: %w", err)
}

// Use structured logging
logger.Error("runner creation failed",
    zap.String("project", projectName),
    zap.Error(err),
)
```

### Testing Guidelines
- Use testify for assertions
- Table-driven tests for multiple cases
- Setup and teardown with defer
- Mock external dependencies

### Documentation
- Public functions must have godoc comments
- Complex logic should have inline comments
- Package-level documentation
- Example usage in godoc

## Contributing

### Before Contributing
1. Read this development guide
2. Review existing code patterns
3. Set up development environment
4. Run existing tests to ensure setup works

### Making a Contribution
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Ensure all tests pass
5. Update documentation
6. Submit a pull request

### Pull Request Checklist
- [ ] Code follows style guidelines
- [ ] All tests pass
- [ ] New functionality has tests
- [ ] Documentation is updated
- [ ] CHANGELOG is updated (if applicable)
- [ ] No breaking changes without version bump

### Release Process
1. Update version in `VERSION` file
2. Update CHANGELOG.md
3. Create git tag
4. Build and publish binaries
5. Update documentation

## Performance Considerations

### Database
- Use prepared statements for hot paths
- Add indexes for query performance
- Use connection pooling
- Batch operations when possible

### Memory Management
- Reuse buffers and objects
- Avoid allocations in hot paths
- Use object pools for frequently allocated types
- Profile with `pprof`

### Concurrency
- Use channels for communication
- Avoid shared mutable state
- Use sync primitives correctly
- Test with race detector

## Debugging Common Issues

### Deadlocks
- Run with race detector
- Check for lock ordering
- Use timeout on channel operations
- Avoid blocking operations in critical sections

### Memory Leaks
- Use `pprof` heap profiling
- Check for goroutine leaks
- Ensure resources are closed
- Monitor with `runtime/pprof`

### Performance Issues
- Profile with `pprof` CPU profiling
- Check database query performance
- Monitor memory allocations
- Use benchmarking

## Tooling

### Required Tools
```bash
# Install development tools
make install-tools

# This installs:
# - golangci-lint (linting)
# - goimports (import formatting)
# - mockery (mock generation)
# - testcontainers (integration testing)
# - delve (debugging)
```

### IDE Configuration
**VS Code:**
- Install Go extension
- Configure gopls
- Enable linting on save
- Set up debugging

**GoLand:**
- Configure Go SDK
- Enable code inspection
- Set up test runner
- Configure debugger

### Useful Aliases
```bash
# Add to ~/.bashrc or ~/.zshrc
alias stest='make test'
alias slint='make lint'
alias sdev='make dev-up'
alias sbuild='make build'
```

---

For more information, see the [Architecture Documentation](architecture.md) or [Testing Guide](testing.md).