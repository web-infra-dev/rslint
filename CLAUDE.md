# RSLint Development Guide for AI Agents

This guide helps AI agents understand and work with the RSLint codebase effectively.

## Core Concepts

RSLint is a TypeScript/JavaScript linter written in Go that implements TypeScript-ESLint rules. It uses the TypeScript compiler API through a Go shim and provides diagnostics via CLI and Language Server Protocol (LSP).

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

   - Add import: `"github.com/web-infra-dev/rslint/internal/rules/rule_name"`
   - In `RegisterAllTypeSriptEslintPluginRules()`, add:
     ```go
     GlobalRuleRegistry.Register("@typescript-eslint/rule-name", rule_name.RuleNameRule)
     ```

4. **Add struct field to TypedRules if the rule needs configuration**

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

### Unit Tests

Create test file: `packages/rslint-test-tools/tests/typescript-eslint/rules/<rule-name>.test.ts`

```typescript
import { describe } from 'vitest';
import { createTester } from '../../utils';

describe('rule-name', () => {
  const { testRule } = createTester({ options: [] });

  testRule({
    valid: ['valid code examples'],
    invalid: [
      {
        code: 'invalid code',
        errors: [{ messageId: 'messageId' }],
      },
    ],
  });
});
```

### Manual Testing

```bash
# Build the project
pnpm build

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
# All tests
pnpm test

# Just Go tests
go test ./...

# Specific rule tests
cd packages/rslint-test-tools
pnpm test <rule-name>
```

### CI Requirements

Your changes must pass:

- `golangci-lint` - Go code quality
- `go fmt` - Go formatting
- `go vet` - Go static analysis
- Unit tests
- VSCode extension tests (may have timing issues, focus on Go tests)

## Debugging Tips

1. **Check rule registration**: Ensure the rule is in `RegisterAllTypeSriptEslintPluginRules()`
2. **Verify configuration**: Rule must be in `rslint.json` to generate diagnostics
3. **Test CLI directly**: Use `rslint` binary to verify rule works
4. **Add logging**: Use `log.Printf()` in LSP code (outputs to stderr)

## Common Pitfalls to Avoid

1. **Don't modify** `getAllTypeScriptEslintPluginRules()` - it must match main branch
2. **Don't change** core infrastructure without understanding impacts
3. **Always handle nil** from type assertions
4. **Test with real TypeScript code** to ensure rule behaves correctly

## When You're Done

1. Ensure all tests pass: `pnpm test`
2. Verify no linting errors: `pnpm build`
3. Test your rule manually with the CLI
4. Document any special behavior or options in the rule implementation
