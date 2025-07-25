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
8. IMPORTANT: Do NOT attempt to run, compile, or execute the Go code

## Rule Registration Reminder

Remember that after creating the rule, it must be registered in:
- `/Users/bytedance/dev/rslint/cmd/rslint/cmd.go`
- `/Users/bytedance/dev/rslint/cmd/rslint/api.go`
- `/Users/bytedance/dev/rslint/cmd/rslint/lsp.go`

Each file needs:
- Import: `"github.com/typescript-eslint/rslint/internal/rules/RULE_NAME_UNDERSCORED"`
- In rules array: `RULE_NAME_UNDERSCORED.RuleNamePascalRule,`

## Common Issues to Verify
- Message IDs in `RuleMessage.Id` should be descriptive (e.g., "preferConstAssertion"), not just the rule name
- The rule variable must be exported and follow naming convention: `var PreferAsConstRule = rule.Rule{...}`
- Imports should include all necessary packages (ast, core, scanner, rule, utils, etc.)

If you find any issues, fix them. Otherwise, confirm the rule is correctly implemented.

Respond with either:
- "VERIFIED" if the rule is correct
- The corrected Go code if fixes are needed

IMPORTANT: Do NOT run or execute any code during verification.