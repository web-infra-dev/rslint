# Run All Tests and Ensure 100% Pass Rate

You are tasked with running all tests in the rslint workspace and ensuring they all pass. This is a critical final step after rule porting or any changes to the codebase.

## Context
- Rule name: {{RULE_NAME}} (if applicable)
- All development work should be complete
- The rule should be properly registered in config.go
- The project should be built successfully

## Your Task

### 1. Navigate to Repository Root
```bash
cd /Users/bytedance/dev/rslint
```

### 2. Run Full Test Suite
Execute the complete test suite across all workspaces:
```bash
pnpm test -w
```

**Timeout**: Maximum 120 seconds per test run attempt.

### 3. Analyze Results
- If ALL tests pass: Report success and stop
- If ANY tests fail: Analyze the failures and fix them

### 4. Fix Test Failures (if needed)
When tests fail, you MUST:

a) **Identify the root cause**:
   - Configuration issues (rule not registered)
   - Implementation bugs in Go code
   - Snapshot mismatches
   - Build issues

b) **Apply fixes systematically**:
   - Rule registration: Ensure both namespaced and non-namespaced registration in `internal/config/config.go`
   - Go implementation: Fix logic errors, infinite loops, AST issues
   - Snapshots: Update if rule count or output format changed
   - Build: Rebuild after configuration changes

c) **Rebuild if necessary**:
   ```bash
   pnpm build
   ```

d) **Re-run tests**:
   ```bash
   pnpm test -w
   ```

### 5. Repeat Until Success
Continue the fix-test cycle until ALL tests pass. You MUST achieve 100% pass rate.

## Common Test Failure Patterns

### A. Rule Registration Issues
**Symptoms**: "Expected diagnostics for invalid case" but no diagnostics generated
**Fix**: Check **ONLY** `internal/config/config.go` has both registrations:
```go
GlobalRuleRegistry.Register("@typescript-eslint/rule-name", package.RuleNameRule)
GlobalRuleRegistry.Register("rule-name", package.RuleNameRule)  // CRITICAL for tests!
```
**IMPORTANT**: DO NOT register in cmd/ files - only config.go. Both registrations are absolutely required.

### B. Snapshot Mismatches  
**Symptoms**: "Expected values to be strictly equal" with ruleCount differences (rule count increased from 48 to 54+)
**Fix**: Update snapshots in packages/rslint:
```bash
cd packages/rslint && npm test -- --update-snapshots
```

### C. Go Implementation Bugs
**Symptoms**: Logic errors, crashes, infinite loops
**Fix**: Review and fix the Go rule implementation, then rebuild

### D. Build Issues
**Symptoms**: "module not found", compilation errors
**Fix**: Run `pnpm build` and check for syntax errors

## Critical Requirements

1. **100% Pass Rate**: ALL tests must pass before completion
2. **No Debug Output**: Remove any fmt.Printf, console.log, or "Sending ruleOptions:" statements
3. **Clean State**: No temporary files or uncommitted debug code
4. **Timeout Management**: If a single test run exceeds 120s, investigate blocking issues
5. **Rule Count**: Rule count has increased from 48 to 54+ - snapshot updates may be needed

## Output Format

Provide clear status updates:
- ✅ "All tests passing - SUCCESS"
- ❌ "Test failures detected - investigating..."
- 🔧 "Applying fix: [description]"
- 🔄 "Re-running tests after fixes..."

## Important Notes

- **Never skip failing tests** - they indicate real issues that must be fixed
- **Always rebuild after config changes** - registration changes require rebuild
- **Be systematic** - fix one category of issues at a time
- **Document fixes** - explain what was wrong and how it was fixed
- **Verify completeness** - ensure ALL workspace tests pass, not just specific packages

Begin by navigating to the repository root and running the full test suite.