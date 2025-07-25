# Commit Rule Implementation

You are tasked with creating a git commit for the newly ported rule.

## Context
- Rule name: {{RULE_NAME}}
- The rule has been successfully ported, tested, and cross-validated
- All necessary files have been created and updated

## Your Task

1. **Navigate to the repository root**:
   ```bash
   cd /Users/bytedance/dev/rslint
   ```

2. **Check git status to see all changes**:
   ```bash
   git status
   ```

3. **Add all relevant files**:
   - Rule implementation: `internal/rules/{{RULE_NAME_UNDERSCORED}}/`
   - Test file: `packages/rslint-test-tools/tests/typescript-eslint/rules/{{RULE_NAME}}.test.ts`
   - Snapshot file: `packages/rslint-test-tools/tests/typescript-eslint/rules/{{RULE_NAME}}.test.ts.snapshot`
   - Command files: `cmd/rslint/cmd.go`, `cmd/rslint/api.go`, `cmd/rslint/lsp.go`
   - Any other modified files (e.g., main test snapshots)

4. **Create a descriptive commit**:
   ```bash
   git add internal/rules/{{RULE_NAME_UNDERSCORED}}/
   git add packages/rslint-test-tools/tests/typescript-eslint/rules/{{RULE_NAME}}*
   git add cmd/rslint/*.go
   git add packages/rslint/tests/*.snapshot
   git commit -m "feat: add {{RULE_NAME}} rule

   - Implement {{RULE_NAME}} rule with all TypeScript ESLint functionality
   - Add comprehensive test coverage matching original TypeScript tests
   - Register rule in cmd files for CLI, API, and LSP usage
   - Add cross-validation tests in rslint-test-tools
   - Update snapshots for new rule count"
   ```

## Important Notes
- DO NOT push the commit - only create it locally
- Ensure all relevant files are included in the commit
- Use conventional commit format: `feat:` for new features
- Include a descriptive commit message that explains what was added
- If there are any unrelated changes, do not include them in the commit

Begin by navigating to the repository root and checking the git status.