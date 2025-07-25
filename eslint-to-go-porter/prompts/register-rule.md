# Register Rule in Command Files

You are tasked with registering a newly ported rule in the rslint command files.

## Context
- Rule name: {{RULE_NAME}}
- The rule has been successfully created and tested
- The rule implementation exists at: `internal/rules/{{RULE_NAME_UNDERSCORED}}/{{RULE_NAME_UNDERSCORED}}.go`

## Your Task

1. **Check if the rule is already registered**:
   ```bash
   grep -n "{{RULE_NAME_UNDERSCORED}}" cmd/rslint/cmd.go
   grep -n "{{RULE_NAME_UNDERSCORED}}" cmd/rslint/api.go
   grep -n "{{RULE_NAME_UNDERSCORED}}" cmd/rslint/lsp.go
   ```

2. **If NOT registered, update each file**:
   
   For each of the three files (`cmd/rslint/cmd.go`, `cmd/rslint/api.go`, `cmd/rslint/lsp.go`):
   
   a. **Add the import** (alphabetically with other rule imports):
   ```go
   "github.com/typescript-eslint/rslint/internal/rules/{{RULE_NAME_UNDERSCORED}}"
   ```
   
   b. **Add to the rules array** (alphabetically):
   ```go
   {{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule,
   ```

3. **Build the project to verify**:
   ```bash
   pnpm build
   ```

4. **Update test snapshots if needed**:
   ```bash
   cd packages/rslint && node --test --test-update-snapshots 'tests/**.test.mjs'
   ```

## Important Notes
- Add imports and rule entries alphabetically
- Ensure proper indentation and formatting
- The build must succeed after registration
- If build fails, check for syntax errors or missing imports

Begin by checking if the rule is already registered.