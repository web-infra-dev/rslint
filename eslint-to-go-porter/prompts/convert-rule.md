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

1. Look at 2-3 existing rule implementations to understand the `rule.Rule` structure
2. Examine the `internal/rule` package to understand:
   - `RuleMessage`, `RuleMeta`, `RuleDocs` types
   - `ReportNode`, `ReportNodeWithFixes`, `ReportNodeWithSuggestions` methods
   - `RuleFix` and `RuleSuggestion` constructors (e.g., `RuleFixReplaceNode`, `RuleFixRemove`, etc.)
   - `RuleContext` and `RuleListeners` interfaces
3. Check the `internal/utils` package for helper functions:
   - `GetNodeText` for getting node text
   - `GetNameFromMember` for extracting property/method names
   - `TrimNodeTextRange` for getting accurate node ranges
4. Understand the AST node types and methods available in the `ast` package
5. Study TypeScript rule patterns and common ESLint rule structures

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
   - Rule message builder functions (use descriptive IDs like "preferConstAssertion", not just the rule name)
   - Main rule implementation with exported variable: `var RuleNamePascalRule = rule.Rule{...}`

5. Ensure the Go code is idiomatic and follows Go conventions
6. Handle all edge cases from the original TypeScript implementation
7. Preserve the rule's logic and behavior exactly
8. IMPORTANT: Use the original TypeScript test as a comprehensive reference to ensure your Go test covers:
   - All valid test cases that should not trigger the rule
   - All invalid test cases that should trigger the rule with correct error messages
   - All edge cases and complex scenarios from the original test
   - Proper error positioning (line, column, endLine, endColumn)
9. IMPORTANT: For rule messages, do NOT use template strings with {{placeholders}}. Instead, format the message directly in Go code using fmt.Sprintf or string concatenation. The rslint framework does not support template string interpolation in messages.
10. IMPORTANT: Ensure error positions (line, column, endLine, endColumn) match the TypeScript rule behavior exactly
11. IMPORTANT: Do NOT attempt to run, compile, or execute the generated Go code
12. IMPORTANT: Create both the rule implementation (.go) and test file (_test.go)
13. IMPORTANT: Follow Go naming conventions and ensure exported identifiers are properly capitalized
14. **CRITICAL**: Avoid recursive calls that could cause infinite loops (e.g., checking if a type reference is simple by recursively calling the same function)
15. **CRITICAL**: Use camelCase for message IDs (e.g., "preferConstAssertion" not "prefer-const-assertion")
16. **CRITICAL**: Remove ALL debug output (fmt.Printf, console.log) from final code - no debug statements should remain
17. **CRITICAL**: When testing, remember that:
    - RSLint uses 1-based line and column numbers (not 0-based)
    - Tests often use the non-namespaced version of rule names (e.g., "array-type" not "@typescript-eslint/array-type")
    - Use `utils.GetNameFromMember()` for robust property name extraction
    - Class members are accessed via `node.Members()` which returns `[]*ast.Node` directly
    - Accessor properties (`accessor a = ...`) are `PropertyDeclaration` nodes with the accessor modifier
18. **RULE REGISTRATION**: The rule MUST be registered in `internal/config/config.go` with BOTH:
    - Namespaced: `@typescript-eslint/rule-name` (for production use)
    - Non-namespaced: `rule-name` (REQUIRED for test compatibility)
    - Missing dual registration causes test failures where no diagnostics are generated

## Output

**CRITICAL: Always work from project root directory (`/Users/bytedance/dev/rslint`)**

IMPORTANT: You are working in the repository root. When examining existing rules, look in the `internal/rules` directory for examples.

Please create ONLY these TWO files:

1. The rule implementation: `internal/rules/{{RULE_NAME_UNDERSCORED}}/{{RULE_NAME_UNDERSCORED}}.go`
2. The test file: `internal/rules/{{RULE_NAME_UNDERSCORED}}/{{RULE_NAME_UNDERSCORED}}_test.go`

Look at existing rule test files for the correct test structure using `rule_tester.RunRuleTester`.

IMPORTANT: 
- Do NOT attempt to run, compile, or execute any of the generated Go code
- Do NOT create any temporary files, debug files, or log files
- Do NOT create any additional files beyond the two specified above
- **CRITICAL**: Do NOT use fmt.Printf or other debug output in the final code
- **CRITICAL**: Remove ANY and ALL debug logging or temporary code before creating the final files
- Do NOT add debug output like "Sending ruleOptions:" or similar console logging
- Final code must be completely clean of debug statements

Create both files using the Write tool. The files should contain complete, working Go code with no explanations or comments outside the code.

