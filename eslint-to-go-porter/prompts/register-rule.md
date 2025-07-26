# Register Rule in Configuration

You are tasked with registering a newly ported rule in the rslint configuration system.

## Context
- Rule name: {{RULE_NAME}}
- The rule has been successfully created and tested
- The rule implementation exists at: `internal/rules/{{RULE_NAME_UNDERSCORED}}/{{RULE_NAME_UNDERSCORED}}.go`

## Your Task

1. **Check if the rule is already registered**:
   ```bash
   grep -n "{{RULE_NAME_UNDERSCORED}}" internal/config/config.go
   ```

2. **If NOT registered, update the global registry**:
   
   In `internal/config/config.go`, find the `RegisterAllTypeSriptEslintPluginRules()` function:
   
   a. **Add the import** (alphabetically with other rule imports):
   ```go
   "github.com/typescript-eslint/rslint/internal/rules/{{RULE_NAME_UNDERSCORED}}"
   ```
   
   b. **CRITICAL: Add BOTH registrations to support namespaced and non-namespaced usage**:
   ```go
   GlobalRuleRegistry.Register("@typescript-eslint/{{RULE_NAME}}", {{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule)
   GlobalRuleRegistry.Register("{{RULE_NAME}}", {{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule)  // REQUIRED for tests!
   ```
   
   **IMPORTANT**: The dual registration is absolutely critical:
   - Namespaced version (@typescript-eslint/rule-name) for production use
   - Non-namespaced version (rule-name) is REQUIRED for test compatibility
   - Missing the non-namespaced registration will cause test failures
   
   **NOTE**: Rules are automatically loaded via the config system. Manual registration in cmd/ files is NOT needed and is incorrect.

3. **Build the project to verify**:
   ```bash
   pnpm build
   ```

4. **Run comprehensive tests**:
   Since adding a new rule increases the total rule count, you need to verify all tests pass.
   Use the `run-all-tests.md` prompt to run the complete test suite and fix any issues.
   
   This will automatically handle:
   - Rule count snapshots updates
   - Cross-validation test verification  
   - Any other test dependencies

## Important Notes
- **ONLY register in internal/config/config.go** - DO NOT add to cmd/ files
- Add imports and rule entries alphabetically
- Ensure proper indentation and formatting
- The build must succeed after registration
- **CRITICAL**: Always register rules with BOTH versions:
  - Namespaced: `@typescript-eslint/rule-name` (for production)
  - Non-namespaced: `rule-name` (REQUIRED for test compatibility)
- Missing the non-namespaced registration will cause "Expected diagnostics but got none" test failures
- If build fails, check for syntax errors or missing imports
- DO NOT create any temporary files during this process
- Clean up any debug output before finalizing changes
- Rules are loaded automatically via config system - no manual cmd/ file changes needed

Begin by checking if the rule is already registered in config.go.