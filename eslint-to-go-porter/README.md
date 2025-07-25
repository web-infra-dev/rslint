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
2. **Adapt Test**: Creates cross-validation test file in rslint-test-tools
3. **Convert**: Uses Claude Code (claude-sonnet-4-20250514) to convert the rule to Go
4. **Verify**: Second Claude pass to verify the conversion is correct
5. **Test**: Runs Go tests on the converted rule
6. **Fix**: If tests fail, uses Claude to fix the implementation based on test output
7. **Cross-Validate**: Runs rslint-test-tools tests to ensure Go rule matches TypeScript behavior
8. **Git Commit**: Creates a git commit with all changes (does not push)

## IMPORTANT: Manual Rule Registration

After porting a rule, you MUST manually register it in:

`/Users/bytedance/dev/rslint/internal/config/config.go`

In the `RegisterAllTypeSriptEslintPluginRules()` function:
- Add import: `"github.com/typescript-eslint/rslint/internal/rules/[rule_name_underscored]"`
- Add BOTH registrations (namespaced and non-namespaced):
  ```go
  GlobalRuleRegistry.Register("@typescript-eslint/[rule-name]", [rule_name_underscored].[RuleNamePascal]Rule)
  GlobalRuleRegistry.Register("[rule-name]", [rule_name_underscored].[RuleNamePascal]Rule)
  ```

Example for `prefer-as-const`:
```go
// Import section
import (
    // ... other imports ...
    "github.com/typescript-eslint/rslint/internal/rules/prefer_as_const"
)

// In RegisterAllTypeSriptEslintPluginRules()
GlobalRuleRegistry.Register("@typescript-eslint/prefer-as-const", prefer_as_const.PreferAsConstRule)
GlobalRuleRegistry.Register("prefer-as-const", prefer_as_const.PreferAsConstRule)
```

After registration:
1. Rebuild: `cd /Users/bytedance/dev/rslint && pnpm build`
2. Update test snapshots if needed: `cd packages/rslint && npm test -- --update-snapshots`

## Testing Notes

- **Cross-validation**: The RuleTester compares diagnostic locations but skips message ID comparison (rslint only exposes rule names)
- **Snapshots**: Tests use Node.js snapshot testing. Run with `--test-update-snapshots` flag when creating/updating snapshots
- **Main tests**: After adding a new rule, update main test snapshots if rule count changes:
  ```bash
  cd packages/rslint && node --test --test-update-snapshots 'tests/**.test.mjs'
  ```

## Features

- **Smart duplicate detection**: Automatically skips already ported rules
- **Multi-pass conversion**: Initial conversion → verification → test → fix cycle
- **Automatic test execution**: Runs Go tests and fixes failures
- **JSON streaming progress**: Debug mode shows Claude's responses
- **Batch processing**: Port multiple or all rules with rate limiting
- **Force mode**: Re-port existing rules with `--force` flag
- **Status tracking**: Check overall porting progress
- **Automatic git commits**: Creates descriptive commits for successfully ported rules

## Output

- Go rules: `/Users/bytedance/dev/rslint/internal/rules/[rule_name]/[rule_name].go`
- Test files: `/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/rules/[rule_name].test.ts`

## Requirements

- Node.js 18+
- Go 1.21+
- Claude CLI installed and authenticated
- Internet connection for fetching files from GitHub