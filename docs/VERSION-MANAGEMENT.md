# Version Management - Stratavore

This document describes the automatic version bumping system for Stratavore.

## Overview

Stratavore uses **semantic versioning** (MAJOR.MINOR.PATCH) with automatic version bumping triggered on commits to the main branch. Version changes are determined by parsing conventional commit message prefixes.

## Current Version

The current version is maintained in:
- `/VERSION` file (source of truth)
- `cmd/stratavored/main.go` (Version constant)
- `cmd/stratavore/main.go` (Version constant)
- Git tags (v1.4.0, v1.5.0, etc.)

**Current version: 1.4.0**

## Version Bumping Rules

The GitHub Actions workflow analyzes the commit message to determine the version bump type:

### MAJOR Version Bump (X.0.0)
Triggered when commit message contains:
- `BREAKING CHANGE:` anywhere in the message
- `breaking change:` (case-insensitive)
- `major:` prefix

**Example:**
```
feat: remove deprecated config format

BREAKING CHANGE: Old YAML format no longer supported, use new format
```
Result: `1.4.0` → `2.0.0`

### MINOR Version Bump (X.Y.0)
Triggered when commit message starts with:
- `feat:` or `feat(...):` (conventional commits)
- `feature:` or `feature(...):` (alternative format)

**Example:**
```
feat(api): add new gRPC endpoint for batch operations
```
Result: `1.4.0` → `1.5.0`

### PATCH Version Bump (X.Y.Z)
Triggered when commit message starts with:
- `fix:` or `fix(...):` (conventional commits)
- `bugfix:` or `bugfix(...):` (alternative format)
- Any other commit (default behavior)

**Example:**
```
fix: correct database migration race condition
```
Result: `1.4.0` → `1.4.1`

## Workflow Details

The version bump workflow (`.github/workflows/version-bump.yml`):

1. **Trigger**: Automatically runs on push to `main` branch (excluding VERSION file changes)
2. **Skip Conditions**: Skips if commit contains `[skip ci]` or `chore: bump version`
3. **Process**:
   - Reads current version from VERSION file
   - Parses commit message to determine bump type
   - Calculates new semantic version
   - Updates VERSION file
   - Updates version constants in:
     - `cmd/stratavored/main.go`
     - `cmd/stratavore/main.go`
   - Creates commit: `chore: bump version to X.Y.Z [skip ci]`
   - Creates git tag: `vX.Y.Z`
   - Pushes both to origin

4. **Prevention of Infinite Loops**: Version bump commits include `[skip ci]` flag to prevent re-triggering the workflow

## Manual Version Updates

To manually update the version (not recommended in normal operation):

```bash
# Update VERSION file
echo "1.5.0" > VERSION

# Update Go version constants
sed -i 's/Version   = ".*"/Version   = "1.5.0"/' cmd/stratavored/main.go
sed -i 's/Version   = ".*"/Version   = "1.5.0"/' cmd/stratavore/main.go

# Create tag and commit
git add VERSION cmd/stratavored/main.go cmd/stratavore/main.go
git commit -m "chore: bump version to 1.5.0"
git tag v1.5.0
git push origin main --tags
```

## Conventional Commit Format

To ensure proper version bumping, use conventional commits:

### Format
```
type(scope): description

[optional body]

[optional footer]
```

### Types
- `feat`: A new feature (triggers MINOR bump)
- `fix`: A bug fix (triggers PATCH bump)
- `docs`: Documentation only (triggers PATCH bump)
- `style`: Code style changes (triggers PATCH bump)
- `refactor`: Code refactoring (triggers PATCH bump)
- `test`: Test additions/changes (triggers PATCH bump)
- `chore`: Build/tooling changes (triggers PATCH bump)

### Examples

**Feature (MINOR bump):**
```
feat(daemon): add session persistence across restarts
```

**Bugfix (PATCH bump):**
```
fix(api): resolve gRPC connection timeout
```

**Breaking Change (MAJOR bump):**
```
feat(config)!: redesign configuration schema

BREAKING CHANGE: Configuration file format has changed. See migration guide.
```

Or alternatively:
```
feat: redesign configuration schema

BREAKING CHANGE: Configuration file format has changed. See migration guide.
```

## Accessing Version Information

### From CLI
```bash
stratavore --version
stratavored --help  # Shows version in help output

# Both output: "1.4.0 (built <timestamp>, commit <hash>)"
```

### From Code
The version is available as a package constant:
```go
package main

import "flag"

var (
    Version   = "1.4.0"
    BuildTime = "unknown"
    Commit    = "unknown"
)
```

### From Git
```bash
# List all version tags
git tag -l

# Get latest version tag
git describe --tags --abbrev=0
```

## Release Process

### Automatic Release
1. Developer commits to main with conventional commit message
2. GitHub Actions runs version bump workflow
3. Workflow automatically:
   - Calculates new version
   - Updates VERSION file and Go constants
   - Creates git commit and tag
   - Pushes to origin

### Manual Release (if needed)
```bash
# Commit normally
git commit -m "feat(api): add new endpoint"
git push origin main

# Wait for automatic workflow to complete
# Verify version was bumped
git describe --tags
```

## Troubleshooting

### Version bump workflow not running
- Check if commit message contains `[skip ci]` or `chore: bump version`
- Verify workflow file syntax in `.github/workflows/version-bump.yml`
- Check GitHub Actions logs in repository settings

### Version mismatch across files
- VERSION file is source of truth
- If out of sync, check git history for incomplete commits
- Manually synchronize using sed commands above

### Tag already exists
- GitHub will fail if tag exists (intentional safety feature)
- Use unique version numbers
- Delete old tag if needed: `git push origin :vX.Y.Z`

## Git Tag Naming

- Tags follow format: `vX.Y.Z` (e.g., v1.4.0, v2.0.0)
- Annotated tags with release message for changelog
- Include commit SHA in push to ensure atomic updates

## Integration with Build System

The Makefile and build scripts reference VERSION file:

```bash
# In Makefile or build scripts:
VERSION=$(cat VERSION)
go build -ldflags="-X main.Version=${VERSION}" ./cmd/stratavored
```

This ensures compiled binaries include correct version string.

## Monitoring

To track version changes over time:

```bash
# View version history
git log --oneline -- VERSION | head -20

# View all version tags
git tag -l 'v*' | sort -V

# Compare versions between branches
git show origin/main:VERSION
git show origin/develop:VERSION
```

## Related Files

- `.github/workflows/version-bump.yml` - Automatic version bump workflow
- `VERSION` - Current version (source of truth)
- `cmd/stratavored/main.go` - Daemon version constant
- `cmd/stratavore/main.go` - CLI version constant
- `Makefile` - Build configuration
- `go.mod` - Go module configuration
