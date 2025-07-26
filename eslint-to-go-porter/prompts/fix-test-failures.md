The Go rule you created is failing tests. Please fix the implementation.

## Test Output

```
{{TEST_OUTPUT}}
```

## Current Go Rule Path

The failing Go rule is at: `{{GO_RULE_PATH}}`

## Original TypeScript Rule

```typescript
{{RULE_SOURCE}}
```

## Instructions

1. Read the current Go rule implementation
2. Analyze the test failures
3. Fix the implementation to pass all tests
4. Ensure the fixes maintain compatibility with the original TypeScript behavior
5. IMPORTANT: Do NOT create any temporary debug files or add fmt.Printf statements
6. Clean up any debug code before providing the final implementation

Provide the complete corrected Go code without any debug statements or temporary code.

## IMPORTANT: Common Issues to Check

1. **Rule Not Registered**: If tests show no diagnostics being generated, ensure the rule is registered in:
   - **ONLY** `internal/config/config.go` - BOTH with namespace AND without namespace:
     ```go
     GlobalRuleRegistry.Register("@typescript-eslint/rule-name", rule_name.RuleNameRule)
     GlobalRuleRegistry.Register("rule-name", rule_name.RuleNameRule)  // CRITICAL for tests!
     ```
   
   **IMPORTANT**: Rules are loaded automatically via the config system.
   - DO NOT manually add imports or registrations to cmd/ files
   - The dual registration (namespaced + non-namespaced) is absolutely required
   - Missing the non-namespaced version causes "Expected diagnostics but got none" failures

2. **Build Required**: After registration, the project must be rebuilt:
   ```bash
   pnpm build
   ```

3. **API Mismatches**: Check for incorrect method names or missing imports
4. **Test Structure**: Ensure test follows the `rule_tester.RunRuleTester` pattern
5. **Message IDs**: In Go rules, use descriptive message IDs in `RuleMessage.Id` field (e.g., "preferConstAssertion", "variableConstAssertion")
6. **Rule Variable Export**: Ensure the rule variable is exported with proper naming: `var RuleNamePascalRule = rule.Rule{...}`
7. **Message Formatting**: Do NOT use template strings with {{placeholders}} in messages. Format messages directly using fmt.Sprintf or string concatenation. The rslint framework does not support template string interpolation.
8. **Position Debugging**: RSLint uses 1-based line and column numbers. When reporting on nodes, consider which part should be highlighted
9. **AST Issues**: 
   - Use `utils.GetNameFromMember()` for property names instead of custom implementations
   - Class members are `node.Members()` not `node.Members`
   - Accessor properties are `PropertyDeclaration` nodes with accessor modifier
10. **Infinite Loops**: Check for recursive functions without proper base cases (e.g., isSimpleType calling itself)
11. **Test Snapshot Updates**: After fixing rules, update test snapshots:
    ```bash
    cd packages/rslint && npm test -- --update-snapshots
    ```
12. **Debug Output**: 
    - While debugging, you may use fmt.Printf temporarily, but MUST remove ALL debug output before providing final code
    - Do NOT add debug logging like "Sending ruleOptions:" or similar console output
    - Clean implementations should have zero debug statements
    - Check service.ts and other files for accidental debug output
13. **Comprehensive Testing**: After fixes, use the `run-all-tests.md` prompt to ensure ALL workspace tests pass, not just the failing ones
14. **Rule Count Updates**: When new rules are added, the total rule count increases from 48 to 54+ and may require snapshot updates