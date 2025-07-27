## Rule: no-unused-expressions

### Test File: no-unused-expressions.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic expression statement handling, TypeScript-specific node unwrapping (as expressions, type assertions, non-null expressions), configuration options parsing, directive prologue handling concept
- ⚠️ **POTENTIAL ISSUES**: Chain expression logic implementation differs significantly, import expression detection may be incorrect, directive prologue implementation is complex and may not match ESLint behavior exactly
- ❌ **INCORRECT**: Missing base rule delegation, incorrect AST node type mapping for several cases, logical expression handling logic differs from TypeScript implementation

### Discrepancies Found

#### 1. Missing Base Rule Delegation
**TypeScript Implementation:**
```typescript
const baseRule = getESLintCoreRule('no-unused-expressions');
const rules = baseRule.create(context);
// Later calls rules.ExpressionStatement(node)
```

**Go Implementation:**
```go
// Complete reimplementation without base rule
var NoUnusedExpressionsRule = rule.Rule{
    Name: "no-unused-expressions",
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        // Custom logic only
    },
}
```

**Issue:** The TypeScript implementation extends ESLint's base rule and only adds TypeScript-specific handling, while the Go version completely reimplements the rule logic. This means the Go version may miss edge cases and behaviors that the base ESLint rule handles.

**Impact:** Missing coverage for complex expressions that the base ESLint rule would catch, potentially leading to false negatives.

**Test Coverage:** Many of the failing test cases likely stem from this fundamental difference.

#### 2. Incorrect Chain Expression Handling
**TypeScript Implementation:**
```typescript
return (
  (node.type === AST_NODE_TYPES.ChainExpression &&
    node.expression.type === AST_NODE_TYPES.CallExpression) ||
  node.type === AST_NODE_TYPES.ImportExpression
);
```

**Go Implementation:**
```go
// ChainExpression with CallExpression (e.g., foo?.())
if node.Kind == ast.KindCallExpression {
    callExpr := node.AsCallExpression()
    if callExpr.QuestionDotToken != nil {
        return true
    }
}
```

**Issue:** The TypeScript version looks for ChainExpression nodes containing CallExpression, while the Go version checks for CallExpression with optional chaining tokens. The AST structure mapping is incorrect.

**Impact:** Optional chaining expressions like `a?.b?.c?.()` may not be handled correctly.

**Test Coverage:** Tests like `test.age?.toLocaleString();` and `one[2]?.[3][4]?.();` reveal this issue.

#### 3. Import Expression Detection
**TypeScript Implementation:**
```typescript
node.type === AST_NODE_TYPES.ImportExpression
```

**Go Implementation:**
```go
// ImportExpression (e.g., import('./foo'))
if node.Kind == ast.KindImportKeyword {
    return true
}
```

**Issue:** The Go version checks for `ast.KindImportKeyword` instead of a proper import expression. Import keywords and import expressions are different AST nodes.

**Impact:** Dynamic imports like `import('./foo')` may not be recognized as valid expressions.

**Test Coverage:** Tests with `import('./foo');` may fail.

#### 4. Logical Expression Handling Logic Difference
**TypeScript Implementation:**
```typescript
if (allowShortCircuit && node.type === AST_NODE_TYPES.LogicalExpression) {
  return isValidExpression(node.right);
}
```

**Go Implementation:**
```go
if binaryExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken || 
   binaryExpr.OperatorToken.Kind == ast.KindBarBarToken {
    // Allow if allowShortCircuit is true, or if right side has side effects
    if opts.AllowShortCircuit {
        return isValidExpression(binaryExpr.Right)
    }
    // Even without allowShortCircuit, allow if right side has side effects
    return isValidExpression(binaryExpr.Right)
}
```

**Issue:** The Go version always checks the right side for side effects even when `allowShortCircuit` is false, which differs from the TypeScript behavior. The TypeScript version only checks the right side when `allowShortCircuit` is true.

**Impact:** Short-circuit expressions may be incorrectly allowed when they should be flagged.

**Test Coverage:** Tests with `allowShortCircuit: true` options may behave differently.

#### 5. Missing Satisfies Expression Handling
**TypeScript Implementation:**
```typescript
// Handles TSInstantiationExpression, TSAsExpression, TSNonNullExpression, TSTypeAssertion
```

**Go Implementation:**
```go
case ast.KindSatisfiesExpression:
    expression = expression.AsSatisfiesExpression().Expression
```

**Issue:** The Go version includes `SatisfiesExpression` handling, but the TypeScript version doesn't explicitly mention it. This might be correct for newer TypeScript versions, but could cause inconsistency.

**Impact:** `satisfies` expressions may be handled differently between implementations.

**Test Coverage:** No test cases cover `satisfies` expressions in the provided tests.

#### 6. Directive Prologue Implementation Complexity
**TypeScript Implementation:**
```typescript
if (node.directive || isValidExpression(node.expression)) {
  return;
}
```

**Go Implementation:**
```go
// Skip directive prologues (e.g., 'use strict')
if ast.IsPrologueDirective(node) {
    // Complex logic to check directive position and content
    // Multiple conditions and parent traversal
}
```

**Issue:** The Go implementation has a much more complex directive prologue detection than the TypeScript version, which simply checks `node.directive`. This complexity may introduce bugs or inconsistencies.

**Impact:** Directive statements like `'use strict'` may be handled inconsistently.

**Test Coverage:** Tests with `'use strict'` in various positions test this behavior.

#### 7. Missing Expression Types in Valid Expression Check
**TypeScript Implementation:**
```typescript
// Base rule handles many expression types automatically
```

**Go Implementation:**
```go
return node.Kind == ast.KindCallExpression || 
       node.Kind == ast.KindNewExpression ||
       node.Kind == ast.KindPostfixUnaryExpression ||
       node.Kind == ast.KindDeleteExpression ||
       node.Kind == ast.KindAwaitExpression ||
       node.Kind == ast.KindYieldExpression
```

**Issue:** The Go version only explicitly handles a limited set of expression types as valid, while the base ESLint rule would handle many more cases automatically.

**Impact:** Valid expressions may be incorrectly flagged as unused expressions.

**Test Coverage:** Edge cases with assignment expressions, update expressions, etc. may fail.

### Recommendations
- **Implement base rule logic**: Either port the complete ESLint base rule logic or ensure all expression types are properly handled
- **Fix chain expression detection**: Properly map TypeScript's ChainExpression AST nodes to Go's AST structure
- **Correct import expression handling**: Use the proper AST node type for import expressions
- **Align logical expression behavior**: Match the TypeScript implementation's short-circuit logic exactly
- **Simplify directive prologue detection**: Use a simpler, more reliable method that matches the TypeScript behavior
- **Add comprehensive expression type coverage**: Ensure all valid expression types are recognized
- **Add missing test cases**: Include tests for satisfies expressions, complex chain expressions, and edge cases
- **Consider AST structure differences**: Carefully map TypeScript AST node types to Go AST equivalents

---