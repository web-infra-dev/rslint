You are tasked with converting a TypeScript ESLint rule to Go for the rslint project.

## TypeScript Rule to Convert

```typescript
{{RULE_SOURCE}}
```

## Instructions

1. Convert the TypeScript ESLint rule above to Go, following the rslint rule structure
2. Use the same package naming convention: `package {{RULE_NAME_UNDERSCORED}}`
3. Implement the rule following these patterns:
   - Create a `{{RULE_NAME_PASCAL}}Rule` variable of type `rule.Rule`
   - Use the same rule name in kebab-case for the Name field
   - Implement the Run function that returns `rule.RuleListeners`
   - Use appropriate AST node kinds from the `ast` package
   - Use the `checker` package for type checking operations
   - Use the `utils` package for common utilities

4. Reference these imports as needed:
   ```go
   import (
       "github.com/microsoft/typescript-go/shim/ast"
       "github.com/microsoft/typescript-go/shim/checker"
       "github.com/microsoft/typescript-go/shim/core"
       "github.com/typescript-eslint/rslint/internal/rule"
       "github.com/typescript-eslint/rslint/internal/utils"
   )
   ```

5. Structure the output file as:
   - Package declaration
   - Imports
   - Helper functions (if any)
   - Rule message builder function
   - Main rule implementation

6. Ensure the Go code is idiomatic and follows Go conventions
7. Handle all edge cases from the original TypeScript implementation
8. Preserve the rule's logic and behavior exactly

## Output

Please provide the complete Go implementation for the file:
`/Users/bytedance/dev/rslint/internal/rules/{{RULE_NAME_KEBAB}}/{{RULE_NAME_KEBAB}}.go`

Provide ONLY the Go code without any explanations or comments outside the code.