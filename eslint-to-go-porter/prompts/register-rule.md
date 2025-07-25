# Register Rule in Command Files

You are tasked with registering a newly ported rule in the rslint command files.

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
   
   b. **Add BOTH registrations to support namespaced and non-namespaced usage**:
   ```go
   GlobalRuleRegistry.Register("@typescript-eslint/{{RULE_NAME}}", {{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule)
   GlobalRuleRegistry.Register("{{RULE_NAME}}", {{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule)
   ```
   
   Note: The dual registration is critical for test compatibility!

3. **Build the project to verify**:
   ```bash
   pnpm build
   ```

4. **Update test snapshots if rule count changed**:
   Since adding a new rule increases the total rule count, you may need to update test snapshots:
   ```bash
   cd packages/rslint && npm test
   ```
   
   If tests fail due to rule count mismatch, update snapshots:
   ```bash
   cd packages/rslint && npm test -- --update-snapshots
   ```

## Important Notes
- Add imports and rule entries alphabetically
- Ensure proper indentation and formatting
- The build must succeed after registration
- Always register rules with BOTH the namespaced (@typescript-eslint/) and non-namespaced versions
- The non-namespaced version is required for test compatibility
- If build fails, check for syntax errors or missing imports

Begin by checking if the rule is already registered in config.go.