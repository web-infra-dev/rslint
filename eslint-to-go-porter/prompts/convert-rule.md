You are tasked with converting a TypeScript ESLint rule to Go for the rslint project.

## TypeScript Rule to Convert

```typescript
{{RULE_SOURCE}}
```

## Original TypeScript Test (for reference)

```typescript
{{TEST_SOURCE}}
```

Use this test as a reference to ensure your Go test covers all the same test cases, edge cases, and scenarios. Convert the test cases from TypeScript ESLint format to Go rslint format.

## Instructions

**STEP 1: Research the API**
Before writing any code, examine the existing rslint codebase to understand:

Note: The original TypeScript rule source is available at `{{RULE_NAME_UNDERSCORED}}/{{RULE_NAME_UNDERSCORED}}.ts` in the current directory for reference.
1. Look at 2-3 existing rule implementations to understand the `rule.Rule` structure
2. Examine the `internal/rule` package to understand:
   - `RuleMessage`, `RuleMeta`, `RuleDocs` types
   - `ReportNode`, `ReportNodeWithFixes`, `ReportNodeWithSuggestions` methods
   - `RuleFix` and `RuleSuggestion` constructors (e.g., `RuleFixReplaceNode`, `RuleFixRemove`, etc.)
   - `RuleContext` and `RuleListeners` interfaces
3. Check the `internal/utils` package for helper functions like `GetNodeText`
4. Understand the AST node types and methods available in the `ast` package

**STEP 2: Convert the Rule**
Only after understanding the API structure, convert the TypeScript ESLint rule above to Go, following the rslint rule structure:

1. Use the same package naming convention: `package {{RULE_NAME_UNDERSCORED}}`
2. Implement the rule following these patterns:
   - Create a `{{RULE_NAME_PASCAL}}Rule` variable of type `rule.Rule`
   - Use the same rule name in kebab-case for the Name field
   - Implement the Run function that returns `rule.RuleListeners`
   - Use appropriate AST node kinds from the `ast` package
   - Use the `checker` package for type checking operations
   - Use the `utils` package for common utilities

3. Reference these imports as needed:
   ```go
   import (
       "github.com/microsoft/typescript-go/shim/ast"
       "github.com/microsoft/typescript-go/shim/checker"
       "github.com/microsoft/typescript-go/shim/core"
       "github.com/typescript-eslint/rslint/internal/rule"
       "github.com/typescript-eslint/rslint/internal/utils"
   )
   ```

4. Structure the output file as:
   - Package declaration
   - Imports
   - Helper functions (if any)
   - Rule message builder function
   - Main rule implementation

5. Ensure the Go code is idiomatic and follows Go conventions
6. Handle all edge cases from the original TypeScript implementation
7. Preserve the rule's logic and behavior exactly
8. IMPORTANT: Use the original TypeScript test as a comprehensive reference to ensure your Go test covers:
   - All valid test cases that should not trigger the rule
   - All invalid test cases that should trigger the rule with correct error messages
   - All edge cases and complex scenarios from the original test
   - Proper error positioning (line, column, endLine, endColumn)
9. IMPORTANT: Do NOT attempt to run, compile, or execute the generated Go code
10. IMPORTANT: Create both the rule implementation (.go) and test file (_test.go)

## Output

IMPORTANT: You are working in the /Users/bytedance/dev/rslint/internal/rules directory. Only examine existing rules within this directory as examples. Do not navigate to parent directories.

Please create TWO files:

1. The rule implementation: `{{RULE_NAME_UNDERSCORED}}/{{RULE_NAME_UNDERSCORED}}.go`
2. The test file: `{{RULE_NAME_UNDERSCORED}}/{{RULE_NAME_UNDERSCORED}}_test.go`

Look at existing rule test files for the correct test structure using `rule_tester.RunRuleTester`.

IMPORTANT: Do NOT attempt to run, compile, or execute any of the generated Go code.

Create both files using the Write tool. The files should contain complete, working Go code with no explanations or comments outside the code.