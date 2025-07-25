# Commit Rule Implementation

You are tasked with creating a git commit for the newly ported rule.

## Context
- Rule name: {{RULE_NAME}}
- The rule has been successfully ported, tested, and cross-validated
- All necessary files have been created and updated

## Your Task

1. **Working in repository root** (already positioned correctly)

2. **CRITICAL: Register the rule in cmd files**:
   The rule MUST be registered in three command files. Follow these steps:

   a) **Check current registration status**:
   ```bash
   grep -n "{{RULE_NAME_UNDERSCORED}}" cmd/rslint/cmd.go || echo "NOT FOUND in cmd.go"
   grep -n "{{RULE_NAME_UNDERSCORED}}" cmd/rslint/api.go || echo "NOT FOUND in api.go"
   grep -n "{{RULE_NAME_UNDERSCORED}}" cmd/rslint/lsp.go || echo "NOT FOUND in lsp.go"
   ```

   b) **For each file where the rule is NOT found, you MUST update it**:
   
   **File: cmd/rslint/cmd.go**
   - Add import: `"github.com/typescript-eslint/rslint/internal/rules/{{RULE_NAME_UNDERSCORED}}"`
   - Add to rules array: `{{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule,`
   
   **File: cmd/rslint/api.go**
   - Add import: `"github.com/typescript-eslint/rslint/internal/rules/{{RULE_NAME_UNDERSCORED}}"`
   - Add to rules array: `{{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule,`
   
   **File: cmd/rslint/lsp.go**
   - Add import: `"github.com/typescript-eslint/rslint/internal/rules/{{RULE_NAME_UNDERSCORED}}"`
   - Add to rules array: `{{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule,`

   c) **Where to add**:
   - Imports: Add alphabetically with other rule imports
   - Rules array: Add alphabetically in the `var rules = []rule.Rule{` array

   d) **Verify the changes**:
   ```bash
   # After making changes, verify they were added correctly
   grep -n "{{RULE_NAME_UNDERSCORED}}" cmd/rslint/cmd.go
   grep -n "{{RULE_NAME_UNDERSCORED}}" cmd/rslint/api.go
   grep -n "{{RULE_NAME_UNDERSCORED}}" cmd/rslint/lsp.go
   ```

3. **Check git status to see all changes**:
   ```bash
   git status
   ```

4. **Add all files and commit**:
   ```bash
   # Add all changes in the repository
   git add -A
   
   # Create the commit
   git commit -m "feat: add {{RULE_NAME}} rule

   - Implement {{RULE_NAME}} rule with all TypeScript ESLint functionality
   - Add comprehensive test coverage matching original TypeScript tests
   - All tests passing successfully"
   ```

## Important Notes
- DO NOT push the commit - only create it locally
- Ensure all relevant files are included in the commit
- Use conventional commit format: `feat:` for new features
- Include a descriptive commit message that explains what was added
- If there are any unrelated changes, do not include them in the commit

Begin by navigating to the repository root and checking the git status.