# Cross-Validate Go Rule Implementation

You are tasked with running the cross-validation tests in rslint-test-tools to ensure that the ported Go rule behaves identically to the original TypeScript ESLint rule.

## Context
- Rule name: {{RULE_NAME}}
- The Go rule has been successfully ported and its Go tests are passing
- Cross-validation tests exist at: `packages/rslint-test-tools/tests/typescript-eslint/rules/{{RULE_NAME}}.test.ts`
- These tests compare the Go rule output against the original TypeScript ESLint rule

## Your Task
1. **Navigate to the test directory**:
   ```bash
   cd packages/rslint-test-tools
   ```

2. **Run the cross-validation test**:
   - First try: `node --import=tsx/esm --test tests/typescript-eslint/rules/{{RULE_NAME}}.test.ts`
   - If snapshot errors occur indicating missing snapshots, run: `node --import=tsx/esm --test --test-update-snapshots tests/typescript-eslint/rules/{{RULE_NAME}}.test.ts`
   - You can also use the npm scripts: `npm test` (runs all tests) or `npm run test:update` (updates snapshots)

3. **Fix any issues**:
   - If tests fail due to behavioral differences, you may need to examine the Go rule implementation
   - If there are snapshot mismatches, verify the differences are expected
   - If there are test framework issues, debug the test runner configuration
   - **CRITICAL**: If tests show "Expected diagnostics for invalid case" but no diagnostics are generated, the rule may not be registered. Check:
     - **ONLY** `internal/config/config.go` - Must have BOTH registrations:
       ```go
       GlobalRuleRegistry.Register("@typescript-eslint/{{RULE_NAME}}", {{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule)
       GlobalRuleRegistry.Register("{{RULE_NAME}}", {{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule)  // CRITICAL for tests!
       ```
     - DO NOT register in cmd/ files - only config.go
     - The dual registration is absolutely required for test compatibility
     - After updating registration, rebuild: `pnpm build`

4. **Report Results**:
   - Report whether the cross-validation tests pass or fail
   - If they fail, provide details about the failures
   - For comprehensive testing across all workspaces, use the `run-all-tests.md` prompt

## Important Notes
- NEVER modify the Go rule implementation directly - only run and analyze the tests
- The test framework uses Node.js built-in test runner with snapshot testing
- Focus on ensuring behavioral equivalence between Go and TypeScript implementations
- **Note about RuleTester**: The RuleTester has been modified to skip message ID comparison because rslint only exposes rule names, not message IDs. The important validation is that diagnostics are generated at the correct locations with the correct ranges.
- **Snapshot updates**: When running tests for the first time or after changes, snapshots may need to be created/updated using the `--test-update-snapshots` flag
- **DO NOT create any temporary files, debug logs, or trace files during testing**
- **DO NOT add debug output** like "Sending ruleOptions:" or similar console logging
- All implementations should be clean of debug statements

Begin by navigating to the test directory and running the cross-validation tests.