You are reviewing a Go rule that was converted from TypeScript. Please verify the implementation is correct.

## Original TypeScript Rule

```typescript
{{RULE_SOURCE}}
```

## Converted Go Rule Path

The Go rule has been written to: `{{GO_RULE_PATH}}`

## Instructions

1. Read the Go rule file at the path above
2. Verify the conversion is accurate and complete
3. Check that all edge cases from the TypeScript version are handled
4. Ensure the Go code follows idiomatic patterns
5. Verify imports are correct
6. Check that the rule structure matches rslint conventions
7. Verify the rule variable is exported (e.g., `var PreferAsConstRule = rule.Rule{...}`)
8. Ensure NO debug output (fmt.Printf, console.log) or temporary code remains in the implementation
8a. Check for accidental debug logging like "Sending ruleOptions:" or similar output
8b. Final code must be completely clean of ALL debug statements
9. IMPORTANT: Do NOT attempt to run, compile, or execute the Go code
10. IMPORTANT: Do NOT create any temporary files during verification

## Rule Registration Reminder

Remember that after creating the rule, it must be registered in:
- **ONLY** `internal/config/config.go` - BOTH with namespace AND without namespace:
  ```go
  GlobalRuleRegistry.Register("@typescript-eslint/rule-name", rule_name.RuleNameRule)
  GlobalRuleRegistry.Register("rule-name", rule_name.RuleNameRule)  // CRITICAL for tests!
  ```

**IMPORTANT**: 
- The rule is automatically loaded via the config system
- DO NOT register in cmd/ files - this is incorrect and outdated
- Both registrations (namespaced + non-namespaced) are absolutely required
- Missing the non-namespaced registration causes "Expected diagnostics but got none" test failures

## Common Issues to Verify
- Message IDs in `RuleMessage.Id` should be descriptive (e.g., "preferConstAssertion"), not just the rule name
- The rule variable must be exported and follow naming convention: `var PreferAsConstRule = rule.Rule{...}`
- Imports should include all necessary packages (ast, core, scanner, rule, utils, etc.)

If you find any issues, fix them. Otherwise, confirm the rule is correctly implemented.

Respond with either:
- "VERIFIED" if the rule is correct
- The corrected Go code if fixes are needed

IMPORTANT: Do NOT run or execute any code during verification.