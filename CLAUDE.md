# Rslint Development Guide for AI Agents

This guide helps AI agents understand and work with the Rslint codebase effectively.

## Core Concepts

Rslint is a TypeScript/JavaScript linter written in Go that implements TypeScript-ESLint rules. It uses the TypeScript compiler API through a Go shim and provides diagnostics via CLI and Language Server Protocol (LSP).

## Rule Implementation Guide

### Creating a New TypeScript-ESLint Rule

1. **Create the rule file**: `internal/rules/<rule_name>/<rule_name>.go`
2. **Define the rule structure**:

```go
package rule_name

import (
    "github.com/microsoft/typescript-go/shim/ast"
    "github.com/web-infra-dev/rslint/internal/rule"
)

var RuleNameRule = rule.Rule{
    Name: "rule-name",  // Use short name WITHOUT @typescript-eslint/ prefix
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        return rule.RuleListeners{
            ast.KindSomeNode: func(node *ast.Node) {
                // Rule logic here
            },
        }
    },
}
```

3. **Register the rule in `internal/config/config.go`**:

   - Add import: `"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/rule_name"`
   - In `RegisterAllTypeScriptEslintPluginRules()`, add:
     ```go
     GlobalRuleRegistry.Register("@typescript-eslint/rule-name", rule_name.RuleNameRule)
     ```

4. **Add the rule to the API hardcoded list in `cmd/rslint/api.go`**:

   - Add import: `"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/rule_name"`
   - In the `origin_rules` slice (around line 100), add:
     ```go
     rule_name.RuleNameRule,
     ```
   - **IMPORTANT**: The API uses a hardcoded list for the test runner. If you don't add your rule here, tests will fail with "Expected diagnostics for invalid case" errors.

5. **Add struct field to TypedRules if the rule needs configuration**

### Critical Safety Requirements

**Always check for nil pointers** when working with AST nodes:

```go
// ALWAYS do this:
typeRef := node.AsTypeReference()
if typeRef == nil {
    return
}

// Check nested properties:
if typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0 {
    // safe to access
}
```

Common nil-check patterns:

- `node.AsXXX()` methods can return nil
- `node.Parent` can be nil for root nodes
- `nodeList.Nodes` - check nodeList isn't nil first
- Check each level when accessing nested properties

### Reporting Diagnostics

Use `ctx.ReportNode()` to emit diagnostics:

```go
ctx.ReportNode(node, rule.RuleMessage{
    Id:          "messageId",
    Description: "Clear description of the violation",
})

// With auto-fix:
ctx.ReportNodeWithFixes(node, message,
    rule.RuleFixReplace(ctx.SourceFile, node, "replacement text"))
```

## Testing Rules

### Primary Focus: Go Tests

**IMPORTANT**: TypeScript test files are maintained from the main branch and should NOT be modified. Focus on Go implementation and Go tests.

### Go Unit Tests

Write comprehensive Go tests in `internal/rules/<rule_name>/<rule_name>_test.go`:

```go
package rule_name

import (
    "testing"
    "github.com/web-infra-dev/rslint/internal/rule"
)

func TestRuleNameRule(t *testing.T) {
    rule.RunTest(t, RuleNameRule, []rule.TestCase{
        // Valid cases
        {
            Code: "valid code example",
        },
        // Invalid cases
        {
            Code: "invalid code example",
            Errors: []rule.ExpectedError{
                {
                    MessageId: "messageId",
                    Line:      1,
                    Column:    1,
                },
            },
        },
    })
}
```

### TypeScript Test Files

**DO NOT CREATE OR MODIFY** TypeScript test files. They exist in `packages/rslint-test-tools/tests/typescript-eslint/rules/` and are maintained from the main branch. If a TypeScript test file exists for your rule, it should work automatically once your Go implementation is complete and registered properly.

### Manual Testing

```bash
# Build the Go binary
go build ./cmd/rslint

# Test directly
cd packages/rslint/fixtures
../bin/rslint src/test.ts

# Add rule to fixtures/rslint.json to enable it
```

## Important Implementation Details

### Rule Naming Convention

- **Rule implementation**: Use short name (e.g., `"array-type"`)
- **Registration**: Use full name (e.g., `"@typescript-eslint/array-type"`)
- **Configuration files**: Use full name with prefix

### AST Navigation

Use the TypeScript AST through the Go shim:

- `ast.KindXXX` constants for node types
- `node.AsXXX()` methods for type assertions (always check for nil)
- `ast.IsXXX(node)` helper functions for type checking

### Running Tests

```bash
# Go tests (primary focus)
go test ./internal/rules/<rule_name>/

# All Go tests
go test ./...

# TypeScript integration tests (after Go implementation is complete)
cd packages/rslint-test-tools
pnpm test <rule-name>

# All tests (run this before finalizing)
pnpm test
```

### CI Requirements

Your changes must pass:

- `golangci-lint` - Go code quality
- `go fmt` - Go formatting
- `go vet` - Go static analysis
- `go test -parallel 4 ./internal/...` - Go unit tests
- `pnpm tsc -b tsconfig.json` - TypeScript type checking
- `pnpm test` - TypeScript/JavaScript unit tests
- `pnpm run lint` - ESLint and other linting checks
- VSCode extension tests (may have timing issues, focus on Go tests)

## Debugging Tips

1. **Check rule registration**: Ensure the rule is in `RegisterAllTypeScriptEslintPluginRules()`
2. **Verify configuration**: Rule must be in `rslint.json` to generate diagnostics
3. **Test CLI directly**: Use `rslint` binary to verify rule works
4. **VSCode extension tests**: If failing with "Expected diagnostics but got 0", this is usually due to missing rule registration (not LSP issues)
5. **Add logging**: Use `log.Printf()` in LSP code (outputs to stderr)

## Common Pitfalls to Avoid

1. **Don't modify** `getAllTypeScriptEslintPluginRules()` - it must match main branch
2. **Don't create or modify TypeScript test files** - they are maintained from main branch
3. **Don't change** core infrastructure without understanding impacts
4. **Always handle nil** from type assertions
5. **Focus on Go tests first** - ensure your Go implementation passes before running TypeScript tests
6. **Missing API registration** - Always add new rules to the hardcoded list in `cmd/rslint/api.go`
7. **Test failures** - "Expected diagnostics for invalid case" usually means the rule isn't registered in the API
8. **Column position mismatches** - TypeScript-ESLint and Go implementation may calculate positions differently, focus on Go test compatibility

## Complete Checklist for Adding a New Rule

### Core Implementation (Primary Focus)

1. [ ] Create rule implementation in `internal/rules/<rule_name>/<rule_name>.go`
2. [ ] Add nil checks for all AST node type assertions
3. [ ] Create comprehensive Go tests in `internal/rules/<rule_name>/<rule_name>_test.go`
4. [ ] Register in `internal/config/config.go` with full @typescript-eslint/ prefix
5. [ ] Add to hardcoded list in `cmd/rslint/api.go`

### Testing & Validation

6. [ ] Run Go tests: `go test ./internal/rules/<rule_name>/` - **MUST PASS**
7. [ ] Run Go quality checks: `go vet ./cmd/... ./internal/...` and `go fmt ./cmd/... ./internal/...`
8. [ ] Build binary: `go build ./cmd/rslint`
9. [ ] Test manually with CLI: `cd packages/rslint/fixtures && ../bin/rslint src/test.ts`
10. [ ] Run all Go tests: `go test -parallel 4 ./internal/...`

### Integration Testing (After Go Implementation Complete)

11. [ ] **DO NOT** create/modify TypeScript test files - they exist from main branch
12. [ ] Run TypeScript integration test: `cd packages/rslint-test-tools && pnpm test <rule-name>`
13. [ ] If TypeScript tests fail due to position mismatches, prioritize Go test compatibility
14. [ ] Run TypeScript type checking: `pnpm tsc -b tsconfig.json`
15. [ ] Run linting checks: `pnpm run lint`
16. [ ] Run all tests: `pnpm test`
17. [ ] Ensure CI passes (focus on Go tests, golangci-lint)

## When You're Done

**Primary Validation (Must Pass)**:

1. Go tests pass: `go test ./internal/rules/<rule_name>/`
2. Go quality checks: `go vet ./cmd/... ./internal/...` and `go fmt ./cmd/... ./internal/...`
3. All Go tests pass: `go test -parallel 4 ./internal/...`
4. Manual CLI testing works

**Secondary Validation**: 5. TypeScript type checking: `pnpm tsc -b tsconfig.json` 6. Linting checks: `pnpm run lint` 7. Integration tests: `pnpm test` (note: some TypeScript test position mismatches are acceptable if Go tests pass) 8. Document any special behavior or options in the rule implementation

**Key Principle**: Go implementation and tests are the source of truth. TypeScript tests are integration tests that may have minor position differences.

## Go Test Infrastructure Notes

### AST Node Kinds

- CallExpression nodes use `ast.KindCallExpression` (value: 213)
- Node kind definitions are in `typescript-go/internal/ast/kind.go`
- The TypeScript Go shim properly maps all AST node types from the TypeScript compiler

### Common Go Test Issues

1. **File Path Comparison**: The test infrastructure uses `sourceFile.FileName()` for file paths. Ensure consistency between test and production code.
2. **Options Handling**: Go tests may pass options as `map[string]interface{}` while TypeScript tests use arrays. Handle both formats:
   ```go
   // Handle array format: [{ option: value }]
   if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
       optsMap, ok = optArray[0].(map[string]interface{})
   }
   ```
3. **Regex Patterns**: Go regex patterns differ from JavaScript. In Go, `/foo/` is literal, use `"foo"` for matching substrings.

### AST Traversal

- The linter uses a dual-visitor pattern: `childVisitor` for main traversal and `patternVisitor` for destructuring patterns
- All nodes are visited via `ForEachChild` which respects TypeScript's AST structure
- Listeners are registered by node kind and executed during traversal
- Use `isInVariableDeclaration` helper to check if a node is within any variable declaration ancestor
