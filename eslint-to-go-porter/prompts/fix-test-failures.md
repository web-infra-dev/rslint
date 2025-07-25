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

Provide the complete corrected Go code.

## IMPORTANT: Common Issues to Check

1. **Rule Not Registered**: If tests show no diagnostics being generated, ensure the rule is registered in:
   - `/Users/bytedance/dev/rslint/cmd/rslint/cmd.go`
   - `/Users/bytedance/dev/rslint/cmd/rslint/api.go`
   - `/Users/bytedance/dev/rslint/cmd/rslint/lsp.go`
   
   Each file needs:
   - Import: `"github.com/typescript-eslint/rslint/internal/rules/{{RULE_NAME_UNDERSCORED}}"`
   - In rules array: `{{RULE_NAME_UNDERSCORED}}.{{RULE_NAME_PASCAL}}Rule,`

2. **Build Required**: After registration, the project must be rebuilt:
   ```bash
   cd /Users/bytedance/dev/rslint && pnpm build
   ```

3. **API Mismatches**: Check for incorrect method names or missing imports
4. **Test Structure**: Ensure test follows the `rule_tester.RunRuleTester` pattern
5. **Message IDs**: In Go rules, use descriptive message IDs in `RuleMessage.Id` field (e.g., "preferConstAssertion", "variableConstAssertion")
6. **Rule Variable Export**: Ensure the rule variable is exported with proper naming: `var RuleNamePascalRule = rule.Rule{...}`