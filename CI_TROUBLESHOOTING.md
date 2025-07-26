# CI Troubleshooting Guide

This document outlines common CI failures in the rslint project and how to fix them.

## Common CI Failures and Solutions

### 1. Git Submodule Issues

**Symptom:** `go: cannot load module typescript-go listed in go.work file: open typescript-go/go.mod: no such file or directory`

**Cause:** The `typescript-go` submodule is not initialized.

**Solution:**
```bash
git submodule update --init --recursive
```

**Prevention:** Ensure CI workflow includes submodule initialization:
```yaml
- name: Checkout code
  uses: actions/checkout@v4
  with:
    submodules: true  # or 'recursive' for nested submodules
```

### 2. Node.js Snapshot Testing Issues

**Symptom:** `TypeError [Error]: t.assert.snapshot is not a function`

**Cause:** Node.js snapshot testing requires Node.js 24+ with experimental features.

**Solution:** The codebase now includes a backward-compatible fallback for older Node.js versions. The test file automatically detects if native snapshot support is available and falls back to manual JSON comparison.

**For Future Changes:**
- Use the `assertSnapshot()` helper function instead of direct `t.assert.snapshot()`
- Test with both Node.js 20.x and 24.x to ensure compatibility

### 3. Formatting Issues

**Symptom:** `Code style issues found in the above file. Run Prettier with --write to fix.`

**Cause:** Code doesn't match Prettier configuration.

**Solution:**
```bash
# Fix all formatting issues
npx prettier --write .

# Check formatting
pnpm format:check
```

**Prevention:** Run formatting checks before committing:
```bash
pnpm format:check
```

### 4. Test Script Glob Pattern Issues

**Symptom:** `Could not find '/path/to/tests/**.test.mjs'`

**Cause:** Incorrect glob patterns in package.json test scripts.

**Solution:** Use proper glob patterns:
```json
{
  "scripts": {
    "test": "node --test tests/*.test.mjs",
    "test:update": "node --test --test-update-snapshots tests/*.test.mjs"
  }
}
```

### 5. IDE Configuration Files

**Symptom:** Large diffs with IDE-specific files (`.idea/`, etc.)

**Cause:** IDE configuration files being committed to repository.

**Solution:** These files are now excluded in `.gitignore`. Remove them if already committed:
```bash
git rm -r .idea/
git commit -m "Remove IDE configuration files"
```

**Prevention:** The `.gitignore` now includes:
```
# IDE files
.idea/
*.iml
```

## Running CI Locally

To simulate the CI environment locally:

### Prerequisites
```bash
# Install dependencies
pnpm install

# Initialize submodules
git submodule update --init --recursive
```

### Full CI Simulation
```bash
# 1. Type checking
pnpm typecheck

# 2. Format checking
pnpm format:check

# 3. Build
pnpm build

# 4. Lint
pnpm run lint

# 5. Node.js tests
pnpm -r test

# 6. Go tests (sample - full test suite takes longer)
go test ./internal/utils ./internal/rules/no_array_delete
```

### Go Test Considerations

- **Full test suite:** `go test ./internal/...` can take several minutes
- **Individual tests:** Run specific packages to speed up development
- **Timeout issues:** Some tests require TypeScript compilation and can be slow

### Node.js Version Compatibility

- **CI Environment:** Node.js 24.x
- **Local Development:** Node.js 20.x+ supported with fallback compatibility
- **Snapshot Testing:** Native support in Node.js 24+, fallback for older versions

## Debugging Tips

### 1. Check Dependencies
```bash
# Verify Go modules
go mod tidy
go mod verify

# Check pnpm dependencies
pnpm install --frozen-lockfile
```

### 2. Clean Build
```bash
# Clean and rebuild
rm -rf node_modules dist packages/*/dist packages/*/bin
pnpm install
pnpm build
```

### 3. Test Environment
```bash
# Check versions
node --version
go version
pnpm --version

# Verify submodules
git submodule status
```

### 4. Isolate Failures
```bash
# Test individual components
cd packages/rslint && node --test tests/api.test.mjs
go test -v ./internal/utils
```

## CI Workflow Overview

The CI pipeline consists of two main jobs:

### 1. `test-go`
- Checkout with submodules
- Setup Go 1.24.1
- Run `go test ./internal/...`

### 2. `test-node`
- Checkout with submodules
- Setup Node.js 24 and Go 1.24.1
- Install pnpm and dependencies
- TypeScript type checking
- Format checking
- Build project
- Run tests with xvfb (Linux) or directly (other OS)
- Lint checking

Both jobs must pass for CI to succeed.

## Best Practices

1. **Always test locally** before pushing
2. **Run the full CI simulation** for major changes
3. **Check submodules** are properly initialized
4. **Avoid committing IDE files** (use .gitignore)
5. **Keep snapshot files** in version control but update them appropriately
6. **Test with multiple Node.js versions** if modifying test infrastructure

## Getting Help

If you encounter CI issues not covered here:

1. Check the specific CI logs for error details
2. Run the failing step locally to reproduce
3. Verify all prerequisites are met
4. Consider if your changes introduced new dependencies or requirements
5. Create an issue with the full error log and reproduction steps