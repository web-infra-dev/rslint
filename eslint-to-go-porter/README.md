# ESLint to Go Porter

Automated tool to port TypeScript ESLint rules to Go using Claude Code.

## Installation

```bash
cd eslint-to-go-porter
npm install
```

## Usage

### Check porting status
```bash
npm start status
```

### List available rules
```bash
npm start list
```

### Port specific rules
```bash
npm start port no-array-delete no-unsafe-call
```

### Port all rules (skips already ported)
```bash
npm start port --all
```

### Force re-port existing rules
```bash
npm start port --force no-array-delete
```

### Show JSON streaming progress
```bash
npm start port --progress no-array-delete
```

## How it works

1. **Fetch**: Downloads TypeScript rule source from typescript-eslint repository
2. **Test Download**: Saves corresponding test files for reference
3. **Convert**: Uses Claude Code (claude-sonnet-4-20250514) to convert the rule to Go
4. **Verify**: Second Claude pass to verify the conversion is correct
5. **Test**: Runs Go tests on the converted rule
6. **Fix**: If tests fail, uses Claude to fix the implementation based on test output
7. **Retest**: Runs tests again after fixes

## Features

- **Smart duplicate detection**: Automatically skips already ported rules
- **Multi-pass conversion**: Initial conversion → verification → test → fix cycle
- **Automatic test execution**: Runs Go tests and fixes failures
- **JSON streaming progress**: Debug mode shows Claude's responses
- **Batch processing**: Port multiple or all rules with rate limiting
- **Force mode**: Re-port existing rules with `--force` flag
- **Status tracking**: Check overall porting progress

## Output

- Go rules: `/Users/bytedance/dev/rslint/internal/rules/[rule_name]/[rule_name].go`
- Test files: `/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/rules/[rule_name].test.ts`

## Requirements

- Node.js 18+
- Go 1.21+
- Claude CLI installed and authenticated
- Internet connection for fetching files from GitHub