# Contributing to Stratavore

We welcome contributions to Stratavore! This guide will help you get started.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Development Workflow](#development-workflow)
3. [Code Standards](#code-standards)
4. [Testing Guidelines](#testing-guidelines)
5. [Documentation](#documentation)
6. [Submitting Changes](#submitting-changes)
7. [Review Process](#review-process)
8. [Community](#community)

## Getting Started

### Prerequisites
- Go 1.22 or later
- Git
- PostgreSQL 14+ with pgvector
- RabbitMQ 3.12+
- Docker (for development environment)

### Initial Setup
```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/stratavore.git
cd stratavore

# Add upstream remote
git remote add upstream https://github.com/meridian/stratavore.git

# Install dependencies
make deps

# Set up development environment
make dev-up

# Run tests to verify setup
make test
```

## Development Workflow

### 1. Create a Branch
```bash
# Sync with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

### 2. Make Your Changes
- Write clean, well-commented code
- Follow existing code patterns and style
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes
```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run linting
make lint

# Format code
make format

# Build to verify
make build
```

### 4. Commit Your Changes
```bash
# Stage your changes
git add .

# Commit with conventional commit format
git commit -m "feat: add new feature description"

# Or for bug fixes
git commit -m "fix: resolve issue with runner startup"

# Push to your fork
git push origin feature/your-feature-name
```

## Code Standards

### Go Conventions
- Follow standard Go formatting (`gofmt`)
- Use `goimports` for import organization
- Exported types and functions should have documentation
- Use meaningful variable and function names

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
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create runner: %w", err)
}

// Use structured logging for errors
logger.Error("runner creation failed",
    zap.String("project", projectName),
    zap.Error(err),
)
```

### Documentation
```go
// RunnerManager manages Claude Code runners
type RunnerManager struct {
    storage  *PostgresClient
    config   *Config
    logger   *zap.Logger
}

// CreateRunner creates a new runner for the specified project.
// Returns the created runner or an error if creation fails.
func (rm *RunnerManager) CreateRunner(ctx context.Context, req *CreateRunnerRequest) (*Runner, error) {
    // Implementation...
}
```

## Testing Guidelines

### Unit Tests
- Test both success and error paths
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Target 80%+ code coverage

### Example Unit Test
```go
func TestRunnerManager_CreateRunner(t *testing.T) {
    tests := []struct {
        name        string
        request     *CreateRunnerRequest
        setupMock   func(*mock.MockStorage)
        want        *Runner
        wantErr     bool
        errContains string
    }{
        {
            name: "successful runner creation",
            request: &CreateRunnerRequest{
                ProjectName: "test-project",
                Model:       "claude-3-sonnet",
            },
            setupMock: func(m *mock.MockStorage) {
                m.EXPECT().CreateRunner(gomock.Any(), gomock.Any()).
                    Return(&Runner{ID: "test-id", ProjectName: "test-project"}, nil)
                m.EXPECT().CreateOutboxEvent(gomock.Any(), gomock.Any()).
                    Return(nil)
            },
            want: &Runner{ID: "test-id", ProjectName: "test-project"},
        },
        {
            name: "quota exceeded",
            request: &CreateRunnerRequest{
                ProjectName: "test-project",
                Model:       "claude-3-sonnet",
            },
            setupMock: func(m *mock.MockStorage) {
                m.EXPECT().CreateRunner(gomock.Any(), gomock.Any()).
                    Return(nil, ErrQuotaExceeded)
            },
            wantErr:     true,
            errContains: "quota exceeded",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockStorage := mock.NewMockStorage(ctrl)
            tt.setupMock(mockStorage)

            rm := &RunnerManager{
                storage: mockStorage,
                config:  &Config{},
                logger:  zap.NewNop(),
            }

            got, err := rm.CreateRunner(context.Background(), tt.request)

            if tt.wantErr {
                require.Error(t, err)
                if tt.errContains != "" {
                    assert.Contains(t, err.Error(), tt.errContains)
                }
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### Integration Tests
- Use testcontainers for real services
- Test complete workflows
- Verify event delivery
- Test error scenarios

### Integration Test Example
```go
func TestRunnerLifecycle_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    ctx := context.Background()
    
    // Setup test infrastructure
    infra := testutils.NewInfrastructure(t)
    defer infra.Cleanup()

    // Create client
    client := infra.NewClient()

    // Test runner creation
    runner, err := client.CreateRunner(ctx, "test-project")
    require.NoError(t, err)
    assert.NotEmpty(t, runner.ID)

    // Wait for runner to be healthy
    assert.Eventually(t, func() bool {
        r, _ := client.GetRunner(ctx, runner.ID)
        return r.Status == RunnerStatusRunning
    }, 30*time.Second, time.Second)

    // Test runner stop
    err = client.StopRunner(ctx, runner.ID)
    require.NoError(t, err)

    // Verify events were published
    events := infra.GetEvents()
    assert.Len(t, events, 2) // created + stopped
}
```

## Documentation

### Code Documentation
- All exported types, functions, and methods need godoc comments
- Include examples for complex operations
- Document error conditions

### User Documentation
- Update user guide for new features
- Update CLI reference for new commands
- Include examples in documentation

### API Documentation
- Update gRPC service definitions
- Include example requests/responses
- Document error responses

## Submitting Changes

### Pull Request Template
```markdown
## Description
Brief description of the change.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] CHANGELOG updated
```

### Commit Message Format
Use conventional commits:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code formatting changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Maintenance tasks

**Examples:**
```
feat: add resource quota enforcement

Implement project-level quotas for max concurrent runners
and memory usage. Includes database migration and API changes.

Closes #123

fix: resolve race condition in runner creation

Use advisory locks to prevent concurrent runner creation
from exceeding quota limits.

Closes #124
```

## Review Process

### Review Criteria
1. **Correctness**: Does the code work as intended?
2. **Testing**: Are there adequate tests?
3. **Documentation**: Is the documentation updated?
4. **Performance**: Any performance implications?
5. **Security**: Any security considerations?
6. **Style**: Does it follow project conventions?

### Review Guidelines
- Be constructive and respectful
- Provide specific feedback
- Suggest improvements
- Ask questions if anything is unclear

### Merge Process
- All PRs require at least one approval
- Maintain CI green status
- Resolve all review comments
- Update based on feedback

## Areas for Contribution

### High Priority
- [ ] Web UI for dashboard and management
- [ ] Enhanced monitoring and alerting
- [ ] Performance optimizations
- [ ] Additional documentation

### Feature Ideas
- [ ] Session similarity search
- [ ] Team collaboration features
- [ ] Workflow automation
- [ ] Advanced scheduling policies

### Improvements
- [ ] Better error messages
- [ ] Enhanced CLI experience
- [ ] More comprehensive tests
- [ ] Performance benchmarking

## Development Tools

### IDE Configuration
**VS Code Extensions:**
- Go (golang.go)
- Test Explorer (hbenl.vscode-test-explorer)
- Docker (ms-azuretools.vscode-docker)

**VS Code Settings:**
```json
{
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    }
}
```

### Useful Commands
```bash
# Generate mocks
make generate

# Run benchmarks
make benchmark

# Run security scan
gosec ./...

# Check for outdated dependencies
go list -u -m all

# Update dependencies
make update-deps
```

## Getting Help

### Questions
- Join our GitHub Discussions
- Check existing issues and PRs
- Read the documentation

### Bug Reports
- Use GitHub issue templates
- Include reproduction steps
- Provide system information
- Include relevant logs

### Feature Requests
- Open an issue with "enhancement" label
- Describe use case
- Suggest implementation approach
- Consider contributing the feature

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please:

- Be respectful and considerate
- Use inclusive language
- Focus on constructive feedback
- Help others learn and grow

## License

By contributing to Stratavore, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Stratavore! Your contributions help make AI development better for everyone.