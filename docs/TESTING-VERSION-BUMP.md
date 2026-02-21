# Testing Version Bump Workflow

This guide covers how to test the automatic version bump workflow locally and in CI.

## Understanding the Workflow

The version bump workflow is triggered on every push to main (except for VERSION file changes).
It will:
1. Parse the commit message
2. Determine bump type (major/minor/patch)
3. Update VERSION file and Go version constants
4. Create and push git tag
5. Push updated version commit

## Manual Testing Strategy

### Test 1: Verify Workflow File Syntax

```bash
cd /path/to/Stratavore

# Validate YAML syntax
python3 << 'EOF'
import yaml
with open('.github/workflows/version-bump.yml', 'r') as f:
    config = yaml.safe_load(f)
    print("Valid YAML")
    print(f"Name: {config['name']}")
    print(f"Jobs: {list(config['jobs'].keys())}")
EOF

# Expected output: Valid YAML, name "Version Bump", job "bump-version"
```

### Test 2: Version Parsing Logic

Test the version bumping algorithm:

```bash
# Create test script
cat > test_version_bump.sh << 'EOF'
#!/bin/bash

bump_version() {
    local current=$1
    local bump_type=$2

    IFS='.' read -r MAJOR MINOR PATCH <<< "$current"

    case "$bump_type" in
        major)
            NEW_MAJOR=$((MAJOR + 1))
            NEW_MINOR=0
            NEW_PATCH=0
            ;;
        minor)
            NEW_MAJOR=$MAJOR
            NEW_MINOR=$((MINOR + 1))
            NEW_PATCH=0
            ;;
        patch)
            NEW_MAJOR=$MAJOR
            NEW_MINOR=$MINOR
            NEW_PATCH=$((PATCH + 1))
            ;;
    esac

    echo "${NEW_MAJOR}.${NEW_MINOR}.${NEW_PATCH}"
}

# Test cases
echo "Testing version bumping logic:"
echo "1.4.0 + patch = $(bump_version 1.4.0 patch)"  # Should be 1.4.1
echo "1.4.0 + minor = $(bump_version 1.4.0 minor)"  # Should be 1.5.0
echo "1.4.0 + major = $(bump_version 1.4.0 major)"  # Should be 2.0.0
echo "2.3.5 + patch = $(bump_version 2.3.5 patch)"  # Should be 2.3.6
echo "2.3.5 + minor = $(bump_version 2.3.5 minor)"  # Should be 2.4.0
echo "2.3.5 + major = $(bump_version 2.3.5 major)"  # Should be 3.0.0
EOF

chmod +x test_version_bump.sh
./test_version_bump.sh
```

### Test 3: Commit Message Parsing

Test the conventional commit parsing logic:

```bash
# Create test script
cat > test_commit_parse.sh << 'EOF'
#!/bin/bash

determine_bump() {
    local commit_msg=$1
    local first_line=$(echo "$commit_msg" | head -1)

    if echo "$commit_msg" | grep -qiE "^BREAKING CHANGE:|breaking change:"; then
        echo "major"
    elif echo "$first_line" | grep -qiE "^feat(\(.+\))?:|^feature:"; then
        echo "minor"
    elif echo "$first_line" | grep -qiE "^fix(\(.+\))?:|^bugfix:"; then
        echo "patch"
    elif echo "$first_line" | grep -qiE "^major:"; then
        echo "major"
    else
        echo "patch"
    fi
}

# Test cases
echo "Testing commit message parsing:"

# MINOR bumps
echo "feat: add new feature = $(determine_bump 'feat: add new feature')"  # minor
echo "feat(api): add endpoint = $(determine_bump 'feat(api): add endpoint')"  # minor
echo "feature: new system = $(determine_bump 'feature: new system')"  # minor

# PATCH bumps
echo "fix: fix bug = $(determine_bump 'fix: fix bug')"  # patch
echo "fix(api): resolve issue = $(determine_bump 'fix(api): resolve issue')"  # patch
echo "bugfix: correction = $(determine_bump 'bugfix: correction')"  # patch
echo "docs: update readme = $(determine_bump 'docs: update readme')"  # patch
echo "chore: cleanup = $(determine_bump 'chore: cleanup')"  # patch

# MAJOR bumps
echo "major: redesign = $(determine_bump 'major: redesign')"  # major
echo "feat: breaking

BREAKING CHANGE: API changed = $(determine_bump $'feat: breaking\n\nBREAKING CHANGE: API changed')"  # major
EOF

chmod +x test_commit_parse.sh
./test_commit_parse.sh
```

### Test 4: File Update Logic

Test the sed commands used to update version constants:

```bash
# Create temporary test files
cp cmd/stratavored/main.go cmd/stratavored/main.go.backup
cp cmd/stratavore/main.go cmd/stratavore/main.go.backup

# Test version update
NEW_VERSION="1.5.0"
sed -i "s/Version   = \"[^\"]*\"/Version   = \"$NEW_VERSION\"/" cmd/stratavored/main.go
sed -i "s/Version   = \"[^\"]*\"/Version   = \"$NEW_VERSION\"/" cmd/stratavore/main.go

# Verify changes
echo "Updated stratavored:"
grep 'Version   =' cmd/stratavored/main.go

echo "Updated stratavore:"
grep 'Version   =' cmd/stratavore/main.go

# Restore backups
mv cmd/stratavored/main.go.backup cmd/stratavored/main.go
mv cmd/stratavore/main.go.backup cmd/stratavore/main.go
```

## Integration Testing (Real Workflow)

### Prerequisite: Test on a Feature Branch

The workflow only triggers on pushes to main. For safe testing, create a test branch:

```bash
git checkout -b test/version-bump-ci
```

Modify the workflow to run on this branch temporarily:

```yaml
# .github/workflows/version-bump.yml
on:
  push:
    branches:
      - main
      - test/version-bump-ci  # Add for testing
```

### Test Case 1: PATCH Bump

```bash
# Make a bugfix commit
echo "// test fix" >> internal/daemon/runner.go

git add internal/daemon/runner.go
git commit -m "fix: correct database connection pool handling"
git push origin test/version-bump-ci

# Monitor GitHub Actions in browser:
# https://github.com/Meridian-Lex/Stratavore/actions

# Check results:
git fetch origin
git describe --tags --abbrev=0
```

**Expected results:**
- Version bumps from 1.4.0 to 1.4.1
- New tag created: v1.4.1
- VERSION file updated
- Go constants updated
- Version bump commit with [skip ci]

### Test Case 2: MINOR Bump

```bash
# Make a feature commit
echo "// new feature" >> internal/api/handler.go

git add internal/api/handler.go
git commit -m "feat(api): add new health check endpoint"
git push origin test/version-bump-ci

# Check results:
git fetch origin
git describe --tags --abbrev=0
```

**Expected results:**
- Version bumps from 1.4.1 to 1.5.0

### Test Case 3: MAJOR Bump

```bash
# Make a breaking change commit
echo "// breaking change" >> cmd/stratavored/main.go

git add cmd/stratavored/main.go
git commit -m "feat: redesign configuration format

BREAKING CHANGE: Old YAML configuration no longer supported"
git push origin test/version-bump-ci

# Check results:
git fetch origin
git describe --tags --abbrev=0
```

**Expected results:**
- Version bumps from 1.5.0 to 2.0.0

### Test Case 4: Skip CI Flag

```bash
# Make a commit with [skip ci]
echo "// minor change" >> README.md

git add README.md
git commit -m "docs: update installation instructions [skip ci]"
git push origin test/version-bump-ci

# Check GitHub Actions - should skip the workflow
```

**Expected results:**
- Workflow does not trigger
- No new version tag created
- Version remains unchanged

### Test Case 5: Version Bump Commit Itself

```bash
# Push another commit after version bump
echo "// another change" >> internal/storage/postgres.go

git add internal/storage/postgres.go
git commit -m "fix: improve connection timeout handling"
git push origin test/version-bump-ci

# Wait for automated version bump to complete
git fetch origin

# Make another commit
echo "// yet another" >> internal/session/manager.go

git add internal/session/manager.go
git commit -m "refactor: simplify session lifecycle"
git push origin test/version-bump-ci

# Verify no infinite loops - only one version bump per real commit
```

**Expected results:**
- Version bump commit includes "chore: bump version"
- Subsequent commits trigger new version bumps
- No infinite loop of version bump commits

## Cleanup After Testing

```bash
# Delete test branch
git push origin --delete test/version-bump-ci
git branch -d test/version-bump-ci

# Remove temporary test files
rm -f test_version_bump.sh test_commit_parse.sh

# Revert workflow to main-only trigger (if modified)
git checkout .github/workflows/version-bump.yml
```

## Production Validation

Once confident in testing, validate on main:

```bash
# Create a proper feature commit
git checkout main
git pull origin main

# Make real feature commit
git commit -m "feat(daemon): add metrics export endpoint"
git push origin main

# Monitor version bump in GitHub Actions
# Verify version incremented correctly
git fetch origin
git describe --tags --abbrev=0

# Check files updated
git log -1 --stat  # Should show VERSION, main.go updates
```

## Monitoring and Debugging

### Check Workflow Logs
- Repository: https://github.com/Meridian-Lex/Stratavore
- Actions tab → Version Bump workflow
- Click on workflow run to see detailed logs
- Each step has a collapsible log section

### Common Issues and Solutions

**Issue: Workflow not triggering**
- Verify commit is pushed to `main` branch
- Check if commit message contains `[skip ci]`
- Look for path-ignore rules in workflow (VERSION file changes skip it)

**Issue: Version not updating in one location**
- Check sed regex matches the current format
- Verify file hasn't changed significantly
- Review workflow logs for exact error

**Issue: Git push fails**
- Verify SSH key is configured
- Check GitHub Actions permissions
- Review branch protection rules

**Issue: Tag already exists**
- Indicates version calculation error
- Delete problematic tag: `git push origin :v1.4.1`
- Investigate commit message that triggered it

## Success Criteria

A successful version bump should:

1. Increment VERSION file to new semantic version
2. Update both Go version constants identically
3. Create annotated git tag with `v` prefix
4. Create commit with `chore: bump version` message
5. Include `[skip ci]` in commit message
6. Push both commit and tag to origin
7. Not trigger another workflow run (infinite loop prevention)
8. Preserve all other file changes from original commit

## Additional Resources

- GitHub Actions documentation: https://docs.github.com/en/actions
- Conventional Commits: https://www.conventionalcommits.org/
- Semantic Versioning: https://semver.org/
- Git Tags: https://git-scm.com/book/en/v2/Git-Basics-Tagging
