# Cross-Validate Go Rule Implementation

You are tasked with running the cross-validation tests in rslint-test-tools to ensure that the ported Go rule behaves identically to the original TypeScript ESLint rule.

## Context
- Rule name: {{RULE_NAME}}
- The Go rule has been successfully ported and its Go tests are passing
- Cross-validation tests exist at: `/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/rules/{{RULE_NAME}}.test.ts`
- These tests compare the Go rule output against the original TypeScript ESLint rule

## Your Task
1. **Navigate to the test directory**:
   ```bash
   cd /Users/bytedance/dev/rslint/packages/rslint-test-tools
   ```

2. **Run the cross-validation test**:
   - First try: `node --import=tsx/esm --test tests/typescript-eslint/rules/{{RULE_NAME}}.test.ts`
   - If snapshot errors occur indicating missing snapshots, run: `node --import=tsx/esm --test --test-update-snapshots tests/typescript-eslint/rules/{{RULE_NAME}}.test.ts`
   - You can also use the npm scripts: `npm test` (runs all tests) or `npm run test:update` (updates snapshots)

3. **Fix any issues**:
   - If tests fail due to behavioral differences, you may need to examine the Go rule implementation
   - If there are snapshot mismatches, verify the differences are expected
   - If there are test framework issues, debug the test runner configuration
   - **IMPORTANT**: If tests show "Expected diagnostics for invalid case" but no diagnostics are generated, the rule may not be registered. Check:
     - `/Users/bytedance/dev/rslint/cmd/rslint/cmd.go`
     - `/Users/bytedance/dev/rslint/cmd/rslint/api.go`
     - `/Users/bytedance/dev/rslint/cmd/rslint/lsp.go`
     Each file needs the rule import and entry in the rules array.

4. **Report Results**:
   - Report whether the cross-validation tests pass or fail
   - If they fail, provide details about the failures
   - Suggest next steps if fixes are needed

## Important Notes
- NEVER modify the Go rule implementation directly - only run and analyze the tests
- The test framework uses Node.js built-in test runner with snapshot testing
- Focus on ensuring behavioral equivalence between Go and TypeScript implementations

Begin by navigating to the test directory and running the cross-validation tests.