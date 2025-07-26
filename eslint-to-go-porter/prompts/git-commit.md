# Commit Rule Implementation

You are tasked with creating a git commit for the newly ported rule.

## Context
- Rule name: {{RULE_NAME}}
- The rule has been successfully ported, tested, and cross-validated
- All necessary files have been created and updated

## Your Task

1. **Working in repository root** (already positioned correctly)

2. **CRITICAL: Verify rule registration**:
   The rule should already be registered ONLY in config.go. Verify:

   ```bash
   grep -n "{{RULE_NAME_UNDERSCORED}}" internal/config/config.go
   ```

   You should see BOTH registrations:
   - `GlobalRuleRegistry.Register("@typescript-eslint/{{RULE_NAME}}", {{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule)`
   - `GlobalRuleRegistry.Register("{{RULE_NAME}}", {{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule)`

   **IMPORTANT**: 
   - Registration should be ONLY in internal/config/config.go
   - DO NOT register in any cmd/ files - this is incorrect
   - Both registrations are absolutely required for test compatibility
   - If NOT registered, this must be fixed before committing.

3. **Ensure all tests pass**:
   Before committing, ALL tests must pass:
   ```bash
   pnpm test -w
   ```
   
   If any tests fail, use the `run-all-tests.md` prompt to fix all issues.
   **DO NOT COMMIT with failing tests.**

4. **Clean up any temporary files and debug output**:
   Before committing, ensure no temporary debug files were created and no debug output remains:
   ```bash
   # Check for any unwanted files
   git status --porcelain | grep -E '(\.tmp|\.log|\.debug|\.trace|\.prof)$'
   
   # Check for debug output in Go files
   grep -r "fmt.Printf" internal/rules/{{RULE_NAME_UNDERSCORED}}/
   
   # Check for debug output in service files
   grep -r "console.log\|Sending ruleOptions" packages/rslint/src/
   
   # If any temp files or debug output exist, remove them
   # git clean -f <files>
   ```

5. **Check git status to see all changes**:
   ```bash
   git status
   ```

6. **Add all files and commit**:
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
- Ensure NO temporary files (*.tmp, *.log, *.debug, *.trace, *.prof) are included
- Ensure NO debug output (fmt.Printf, console.log, "Sending ruleOptions:") remains in any files
- The commit should only contain the rule implementation, test file, and config registration
- Rule count may have increased from 48 to 54+ rules - this is expected

Begin by verifying the rule is registered and checking for any temporary files.