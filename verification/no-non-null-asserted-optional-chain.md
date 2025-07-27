# Rule Validation: no-non-null-asserted-optional-chain

## Rule: no-non-null-asserted-optional-chain

### Test File: no-non-null-asserted-optional-chain.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic non-null assertion detection on optional chains, message IDs and descriptions match, suggestion-based fixes (not auto-fixes), terminal assertion logic for TypeScript 3.9+ compatibility
- ⚠️ **POTENTIAL ISSUES**: Complex AST pattern matching approach differs significantly, reliance on `ast.IsOptionalChain()` without validation, position calculation for fix suggestions may be inaccurate
- ❌ **INCORRECT**: Missing selector-based pattern matching from TypeScript implementation, overly complex logic that may miss edge cases, potential issues with parenthesized expression handling

### Discrepancies Found

#### 1. Fundamental Pattern Matching Approach Difference
**TypeScript Implementation:**
```typescript
return {
  // Pattern 1: (x?.y)! - non-nulling a wrapped chain
  'TSNonNullExpression > ChainExpression'(node: TSESTree.ChainExpression): void {
    // selector guarantees this assertion
    const parent = node.parent as TSESTree.TSNonNullExpression;
    // ... report error
  },

  // Pattern 2: x?.y! - non-nulling at the end of a chain  
  'ChainExpression > TSNonNullExpression'(node: TSESTree.TSNonNullExpression): void {
    // ... report error
  },
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
  ast.KindNonNullExpression: func(node *ast.Node) {
    nonNullExpr := node.AsNonNullExpression()
    expression := nonNullExpr.Expression
    
    if isDirectOptionalChainAssertion(expression, node) {
      reportError(ctx, node, node)
    }
  },
}
```

**Issue:** The TypeScript implementation uses two specific CSS-like selectors to target exact AST patterns: `TSNonNullExpression > ChainExpression` and `ChainExpression > TSNonNullExpression`. The Go implementation uses a single listener on `KindNonNullExpression` and tries to recreate the selector logic manually.

**Impact:** The Go approach may miss cases or incorrectly flag valid code because it doesn't precisely match the TypeScript selector semantics.

**Test Coverage:** This affects all test cases, particularly the parenthesized cases like `(foo?.bar)!`

#### 2. Parenthesized Expression Handling
**TypeScript Implementation:**
```typescript
// Directly targets: TSNonNullExpression > ChainExpression
// This automatically handles (foo?.bar)! because the ChainExpression is directly under TSNonNullExpression
'TSNonNullExpression > ChainExpression'(node: TSESTree.ChainExpression): void
```

**Go Implementation:**
```go
// Pattern 1: NonNullExpression > ChainExpression (parenthesized optional chain)
if expression.Kind == ast.KindParenthesizedExpression {
  parenExpr := expression.AsParenthesizedExpression()
  return hasOptionalChaining(parenExpr.Expression)
}
```

**Issue:** The Go implementation assumes parenthesized expressions are the way to handle `(foo?.bar)!`, but this may not be correct. The TypeScript selector `TSNonNullExpression > ChainExpression` suggests the ChainExpression is a direct child, not necessarily wrapped in parentheses.

**Impact:** May miss or incorrectly handle parenthesized optional chain assertions.

**Test Coverage:** Affects test cases like `(foo?.bar)!`, `(foo?.bar)!()`, `(foo?.bar!)`, `(foo?.bar!)()`

#### 3. Terminal Assertion Logic Complexity
**TypeScript Implementation:**
```typescript
// Simple and direct - uses two separate selectors
// No complex "terminal" logic needed because selectors handle the patterns
```

**Go Implementation:**
```go
func isTerminalAssertion(nonNullNode *ast.Node) bool {
  if nonNullNode.Parent == nil {
    return true
  }
  
  parent := nonNullNode.Parent
  switch parent.Kind {
  case ast.KindPropertyAccessExpression:
    propAccess := parent.AsPropertyAccessExpression()
    return propAccess.Expression != nonNullNode
  case ast.KindElementAccessExpression:
    elemAccess := parent.AsElementAccessExpression()
    return elemAccess.Expression != nonNullNode
  case ast.KindCallExpression:
    callExpr := parent.AsCallExpression()
    return callExpr.Expression != nonNullNode
  default:
    return true
  }
}
```

**Issue:** The Go implementation introduces complex "terminal assertion" logic that doesn't exist in the TypeScript version. This may be unnecessary and could introduce bugs.

**Impact:** May incorrectly flag valid TypeScript 3.9+ patterns like `foo?.bar!.baz` as invalid.

**Test Coverage:** Could affect valid test cases like `foo?.bar!.baz`, `foo?.bar!()`, `foo?.['bar']!.baz`

#### 4. Missing Direct Selector Equivalent
**TypeScript Implementation:**
```typescript
// Pattern 2: ChainExpression > TSNonNullExpression
// This catches cases like foo?.bar! where the NonNull is inside the chain
'ChainExpression > TSNonNullExpression'(node: TSESTree.TSNonNullExpression): void
```

**Go Implementation:**
```go
// Only listens to KindNonNullExpression, no equivalent to the reverse pattern
// where the NonNull is a child of the ChainExpression
```

**Issue:** The Go implementation doesn't have an equivalent to the `ChainExpression > TSNonNullExpression` pattern, which may miss certain cases.

**Impact:** Could miss violations where the non-null assertion is embedded within the chain expression.

**Test Coverage:** May affect test cases like `(foo?.bar!)` where the assertion is inside parentheses

#### 5. Fix Position Calculation
**TypeScript Implementation:**
```typescript
// For TSNonNullExpression > ChainExpression
fix(fixer): TSESLint.RuleFix {
  return fixer.removeRange([
    parent.range[1] - 1,  // parent is the TSNonNullExpression
    parent.range[1],
  ]);
},

// For ChainExpression > TSNonNullExpression  
fix(fixer): TSESLint.RuleFix {
  return fixer.removeRange([node.range[1] - 1, node.range[1]]);
},
```

**Go Implementation:**
```go
nonNullEnd := nonNullNode.End()
exclamationStart := nonNullEnd - 1
exclamationRange := core.NewTextRange(exclamationStart, nonNullEnd)
```

**Issue:** The Go implementation always calculates the fix position based on the nonNullNode, but the TypeScript implementation uses different calculation strategies depending on which pattern matched.

**Impact:** Fix suggestions may target incorrect positions in the source code.

**Test Coverage:** Could affect the accuracy of suggested fixes in all invalid test cases.

### Recommendations
- **Implement equivalent selector patterns**: Create separate handlers that mimic `TSNonNullExpression > ChainExpression` and `ChainExpression > TSNonNullExpression` patterns
- **Simplify logic**: Remove the complex terminal assertion logic and rely on the selector patterns like the TypeScript version
- **Validate AST pattern assumptions**: Test the assumptions about how parenthesized expressions and chain expressions are represented in the Go AST
- **Fix position calculation**: Use pattern-specific fix calculations that match the TypeScript implementation
- **Add comprehensive testing**: Verify each test case individually to ensure the Go port produces identical results

### Critical Issues Requiring Immediate Attention
1. The Go implementation may fail to catch violations that the TypeScript version would catch due to missing the `ChainExpression > TSNonNullExpression` pattern
2. The terminal assertion logic could incorrectly allow valid TypeScript 3.9+ code patterns
3. Fix suggestions may be positioned incorrectly, breaking the user experience

---