## Rule: no-unused-expressions

### Test File: no-unused-expressions.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic expression statement validation, TypeScript-specific node unwrapping (as expressions, type assertions, non-null expressions), configuration option support for allowShortCircuit/allowTernary/allowTaggedTemplates
- ⚠️ **POTENTIAL ISSUES**: ImportExpression detection logic, directive prologue detection, optional chaining detection, binary expression handling for logical operators
- ❌ **INCORRECT**: Missing base ESLint rule integration, incorrect AST node type mapping for import expressions, complex directive prologue logic that may not match ESLint behavior

### Discrepancies Found

#### 1. Base Rule Integration Missing
**TypeScript Implementation:**
```typescript
const baseRule = getESLintCoreRule('no-unused-expressions');
// ... delegates to baseRule.create(context) and only adds TS-specific handling
rules.ExpressionStatement(node);
```

**Go Implementation:**
```go
// Complete reimplementation without base rule delegation
return rule.RuleListeners{
    ast.KindExpressionStatement: func(node *ast.Node) {
        // Custom logic for all cases
    },
}
```

**Issue:** The TypeScript version extends the base ESLint rule and only adds TypeScript-specific handling, while the Go version is a complete rewrite. This could lead to missing edge cases that the base ESLint rule handles.

**Impact:** May miss various expression types and edge cases that the original ESLint rule covers.

**Test Coverage:** All test cases could potentially be affected by missing base rule logic.

#### 2. ImportExpression Detection
**TypeScript Implementation:**
```typescript
node.type === AST_NODE_TYPES.ImportExpression
```

**Go Implementation:**
```go
if node.Kind == ast.KindImportKeyword {
    return true
}
```

**Issue:** The Go version checks for `KindImportKeyword` instead of import expressions/calls. Import expressions like `import('./foo')` are call-like expressions, not just keywords.

**Impact:** May incorrectly identify import keywords in other contexts as valid expressions.

**Test Coverage:** Test cases with `import('./foo')` may not work correctly.

#### 3. Directive Prologue Detection
**TypeScript Implementation:**
```typescript
if (node.directive || isValidExpression(node.expression)) {
  return;
}
```

**Go Implementation:**
```go
if ast.IsPrologueDirective(node) {
    // Complex custom logic to detect 'use strict' and other directives
    // Checks literal text, parent context, and position in block
}
```

**Issue:** The TypeScript version relies on a simple `directive` property, while the Go version has complex custom detection logic that may not match ESLint's behavior exactly.

**Impact:** May incorrectly allow or disallow directive statements in certain contexts.

**Test Coverage:** Test cases with 'use strict' in modules, namespaces, and functions.

#### 4. Optional Chaining Detection
**TypeScript Implementation:**
```typescript
(node.type === AST_NODE_TYPES.ChainExpression &&
  node.expression.type === AST_NODE_TYPES.CallExpression)
```

**Go Implementation:**
```go
if node.Kind == ast.KindCallExpression {
    callExpr := node.AsCallExpression()
    if callExpr.QuestionDotToken != nil {
        return true
    }
}
// Plus complex parent traversal logic for property/element access
```

**Issue:** The TypeScript version checks for `ChainExpression` wrapping `CallExpression`, while the Go version checks for `QuestionDotToken` and does parent traversal. These approaches may not be equivalent.

**Impact:** May incorrectly validate or invalidate optional chaining expressions.

**Test Coverage:** Test cases with `foo?.()`, `a?.['b']?.c()`, etc.

#### 5. Binary Expression Logic
**TypeScript Implementation:**
```typescript
if (allowShortCircuit && node.type === AST_NODE_TYPES.LogicalExpression) {
  return isValidExpression(node.right);
}
```

**Go Implementation:**
```go
if node.Kind == ast.KindBinaryExpression {
    binaryExpr := node.AsBinaryExpression()
    if binaryExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken || 
       binaryExpr.OperatorToken.Kind == ast.KindBarBarToken {
        if opts.AllowShortCircuit {
            return isValidExpression(binaryExpr.Right)
        }
        // Even without allowShortCircuit, allow if right side has side effects
        return isValidExpression(binaryExpr.Right)
    }
}
```

**Issue:** The Go version allows logical expressions with side effects even when `allowShortCircuit` is false, while the TypeScript version only allows them when the option is enabled.

**Impact:** May incorrectly allow short-circuit expressions when they should be flagged.

**Test Coverage:** Test case `'foo && foo?.bar;'` with `allowShortCircuit: true` should flag this, but Go version might not.

#### 6. TypeScript Node Type Mappings
**TypeScript Implementation:**
```typescript
expressionType === TSTree.AST_NODE_TYPES.TSInstantiationExpression ||
expressionType === TSTree.AST_NODE_TYPES.TSAsExpression ||
expressionType === TSTree.AST_NODE_TYPES.TSNonNullExpression ||
expressionType === TSTree.AST_NODE_TYPES.TSTypeAssertion
```

**Go Implementation:**
```go
case ast.KindAsExpression:
case ast.KindTypeAssertionExpression:
case ast.KindNonNullExpression:
case ast.KindSatisfiesExpression:  // Extra node type
// Missing: TSInstantiationExpression equivalent
case ast.KindExpressionWithTypeArguments:  // Different mapping
```

**Issue:** The Go version maps to different AST node types and includes `SatisfiesExpression` not in TypeScript version, while missing proper handling for instantiation expressions.

**Impact:** May not correctly handle all TypeScript-specific expression types.

**Test Coverage:** Test cases with `Foo<string>;` and type assertions.

### Recommendations
- Implement base ESLint rule logic or ensure all edge cases from base rule are covered
- Fix ImportExpression detection to properly identify dynamic import calls
- Simplify directive prologue detection to match ESLint behavior or verify current logic is equivalent
- Review optional chaining detection logic and ensure it matches TypeScript-ESLint behavior
- Fix binary expression logic to respect allowShortCircuit option correctly
- Verify TypeScript node type mappings are correct and complete
- Add comprehensive test coverage for edge cases identified in base ESLint rule

---