# RSLint Automated Rule Porting

This document describes the `automated-port.js` script that automatically ports missing TypeScript-ESLint rules to Go for RSLint.

## Overview

The `automated-port.js` script is designed to automatically discover and port TypeScript-ESLint rules that haven't been implemented in RSLint yet. It uses Claude CLI to analyze the original TypeScript implementations and create equivalent Go versions following RSLint's patterns.

**IMPORTANT**: This script **ONLY creates and edits files**. It does **NOT run tests or builds**. After porting rules, use `/Users/bytedance/dev/rslint/automate-build-test.js` to test the newly created rules.

## Features

- **Automatic Rule Discovery**: Compares TypeScript-ESLint's rules with existing RSLint rules to identify missing ones
- **Source Context Fetching**: Downloads original TypeScript rule and test files for accurate porting
- **Concurrent Processing**: Supports parallel porting with configurable worker count
- **Claude CLI Integration**: Uses Claude to analyze and port rules with proper context
- **Progress Tracking**: Real-time progress monitoring and detailed logging
- **Retry Logic**: Automatic retry on failures with exponential backoff
- **File Locking**: Prevents conflicts during concurrent execution

## Prerequisites

1. **Claude CLI**: Must be installed and configured
2. **Settings File**: Uses `.claude/settings.local.json` for configuration
3. **Node.js**: Script requires Node.js runtime
4. **Network Access**: Needs access to GitHub for fetching TypeScript-ESLint sources

## Usage

### Basic Commands

```bash
# List missing rules (no porting)
node automated-port.js --list

# Show current porting status
node automated-port.js --status

# Port all missing rules sequentially
node automated-port.js

# Port with concurrent processing (3 workers)
node automated-port.js --concurrent --workers=3
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--list` | List missing rules only (no porting) | - |
| `--status` | Show porting status and progress | - |
| `--concurrent` | Enable parallel processing | Sequential |
| `--workers=N` | Number of concurrent workers | 3 |
| `--help, -h` | Show help message | - |

### Examples

```bash
# Quick status check
node automated-port.js --status

# See what needs to be ported
node automated-port.js --list

# Start porting (recommended for first run)
node automated-port.js

# Fast parallel porting (for experienced users)
node automated-port.js --concurrent --workers=5
```

## How It Works

### 1. Rule Discovery Process

The script automatically:
1. Fetches the complete list of TypeScript-ESLint rules from GitHub
2. Scans existing RSLint rules in `internal/rules/`
3. Identifies missing rules that need to be ported
4. Prioritizes rules for porting

### 2. Porting Process

For each missing rule, the script:

1. **Fetches Sources**: Downloads the original TypeScript rule and test files
2. **Analyzes Rule**: Uses Claude to understand the rule's behavior and requirements
3. **Creates Go Implementation**: Ports the rule to Go following RSLint patterns
4. **Transforms Tests**: Adapts TypeScript tests to RSLint's test framework
5. **Registers Rule**: Adds the rule to appropriate configuration files

**Note**: The script stops here - it does NOT run tests or builds. Testing is handled by the separate `automate-build-test.js` script.

### 3. File Structure

The script creates files in these locations:

```
/Users/bytedance/dev/rslint/
├── internal/rules/
│   └── rule_name/
│       ├── rule_name.go          # Go rule implementation
│       └── rule_name_test.go     # Go tests
└── packages/rslint-test-tools/tests/typescript-eslint/rules/
    └── rule-name.test.ts         # TypeScript test file
```

## Progress Monitoring

### Sequential Mode
```
[14:30:15] → Starting rule porting with 42 missing rules
[14:30:16] 🔄 Porting adjacent-overload-signatures (attempt 1/3)
[14:30:17] → Fetching original sources from GitHub...
[14:30:18] ✓ Original sources fetched (rule: yes, test: yes)
[14:30:19] 🔄 Claude: Analyzing TypeScript-ESLint rule structure...
[14:30:45] ✓ Successfully ported adjacent-overload-signatures in 29s
```

### Concurrent Mode
```
[14:30:15] → Starting rule porting with 42 missing rules
[14:30:15] → Starting worker 1: porter_0_a1b2c3d4
[14:30:15] → Starting worker 2: porter_1_e5f6g7h8
[14:30:15] → Starting worker 3: porter_2_i9j0k1l2
[14:30:30] ◆ Progress: 3/42 (7% success) - 2 ported, 0 failed, 1 in progress
```

## Configuration

### Claude Settings
The script uses the existing `.claude/settings.local.json` configuration:
- Same permissions as `automate-build-test.js`
- File locking hooks for concurrent safety
- Streaming JSON output format

### Timeout Settings
- **Per Rule**: 10 minutes maximum
- **Retry Attempts**: 3 attempts per rule
- **Retry Delay**: 10 seconds between attempts
- **Rate Limiting**: 3 seconds between rules

## Error Handling

### Common Issues and Solutions

#### 1. Network Errors
```
✗ Failed to fetch rule sources from GitHub
```
**Solution**: Check internet connection and GitHub availability

#### 2. Claude CLI Errors
```
✗ Claude CLI failed (exit code 1)
```
**Solution**: Verify Claude CLI is installed and configured properly

#### 3. Permission Errors
```
✗ Could not acquire file lock
```
**Solution**: Ensure no other instances are running, or reduce worker count

#### 4. Timeout Errors
```
✗ Rule timed out after 10 minutes
```
**Solution**: Rule may be complex; will retry automatically

### Retry Logic
- Rules are retried up to 3 times on failure
- Each retry includes fresh source fetching
- Exponential backoff prevents API rate limiting
- Failed rules are reported in final summary

## Output and Results

### Success Indicators
- ✅ New Go rule files created in `internal/rules/`
- ✅ Test files created in `packages/rslint-test-tools/tests/`
- ✅ Rules registered in configuration files
- ✅ Files created without errors

**Note**: Build and test success must be verified separately using `automate-build-test.js`

### Final Summary
```
=== Porting Summary ===

✓ Successfully ported 38 rules:
  - adjacent-overload-signatures
  - array-type
  - await-thenable
  ...

✗ Failed to port 4 rules:
  - complex-rule-1: Parse error in TypeScript source
  - complex-rule-2: Timeout after 3 attempts
  ...
```

## Best Practices

### When to Use Sequential Mode
- **First time** running the script
- **Debugging** specific rule issues
- **Limited resources** or slow network
- **Learning** how the porting process works

### When to Use Concurrent Mode
- **Large batches** of rules to port
- **Fast network** and powerful machine
- **Experienced** with the porting process
- **Time-sensitive** porting needs

### Recommended Workflow
1. Start with `--status` to see scope
2. Use `--list` to review missing rules
3. Run sequentially first time: `node automated-port.js`
4. Use concurrent for subsequent runs: `node automated-port.js --concurrent`
5. **Test the ported rules**: `node automate-build-test.js` (separate script)
6. **Fix any issues**: Use `automate-build-test.js` with Claude to fix test failures

## Troubleshooting

### Script Won't Start
1. Check Node.js is installed: `node --version`
2. Verify Claude CLI: `claude --version`
3. Check settings file exists: `.claude/settings.local.json`

### Rules Failing to Port
1. Check GitHub connectivity
2. Verify Claude CLI authentication
3. Review error messages in output
4. Try single rule porting first

### Concurrent Issues
1. Reduce worker count: `--workers=1`
2. Check file system permissions
3. Ensure no other automation running
4. Clear tmp directory if needed

### Performance Issues
1. Use `--concurrent` for speed
2. Increase worker count: `--workers=5`
3. Check network bandwidth
4. Monitor system resources

## Integration with Existing Workflow

This script complements the existing `automate-build-test.js`:

1. **Port Rules**: Use `automated-port.js` to create new rules (file creation only)
2. **Test Rules**: Use `automate-build-test.js` to verify they work (builds and tests)
3. **Fix Issues**: Use `automate-build-test.js` with Claude to fix any test failures
4. **Build System**: Integrate both into CI/CD pipeline

**Key Separation**: `automated-port.js` creates files, `automate-build-test.js` runs tests and builds.

## Contributing

When modifying the script:
1. Follow same patterns as `automate-build-test.js`
2. Test with both sequential and concurrent modes
3. Verify error handling works correctly
4. Update this documentation

## Support

For issues or questions:
1. Check existing RSLint documentation
2. Review Claude CLI documentation
3. Examine `automate-build-test.js` for patterns
4. Test with simple rules first

---

**Related Files:**
- `automate-build-test.js` - Test automation script
- `.claude/settings.local.json` - Claude CLI configuration
- `eslint-to-go-porter/` - Alternative porting tool
- `internal/rules/` - Existing Go rule implementations