## Rule: no-non-null-asserted-optional-chain

### Test File: no-non-null-asserted-optional-chain.test.ts

### Validation Summary
- ✅ **CORRECT**: Core detection of non-null assertions on optional chains, message IDs and descriptions, suggestion-based fixes with range removal, terminal assertion detection for TypeScript 3.9+ compatibility
- ⚠️ **POTENTIAL ISSUES**: AST traversal approach differs significantly, reliance on `ast.IsOptionalChain()` without verification of its behavior
- ❌ **INCORRECT**: Missing detection of parenthesized optional chains `(foo?.bar)!`, incomplete pattern matching compared to TypeScript selectors

### Discrepancies Found

#### 1. Missing Parenthesized Optional Chain Detection
**TypeScript Implementation:**
```typescript
// Selector specifically targets: TSNonNullExpression > ChainExpression
'TSNonNullExpression > ChainExpression'(
  node: TSESTree.ChainExpression,
): void {
  // This catches cases like (foo?.bar)! where the chain is wrapped
  const parent = node.parent as TSESTree.TSNonNullExpression;
  context.report({
    node,
    messageId: 'noNonNullOptionalChain',
    // ...
  });
}
```

**Go Implementation:**
```go
// Only checks if expression is ParenthesizedExpression, but doesn't handle the case
// where a ChainExpression is directly the child of NonNullExpression
if expression.Kind == ast.KindParenthesizedExpression {
    parenExpr := expression.AsParenthesizedExpression()
    return hasOptionalChaining(parenExpr.Expression)
}
```

**Issue:** The Go implementation attempts to handle parenthesized expressions but doesn't correctly match the TypeScript selector pattern `TSNonNullExpression > ChainExpression`. The TypeScript version has two separate handlers - one for when a ChainExpression is directly under a NonNullExpression, and another for the reverse.

**Impact:** Test cases like `(foo?.bar)!`, `(foo?.bar)!()`, and `(foo?.bar)!.baz` may not be detected correctly.

**Test Coverage:** Invalid test cases 5-8 in the test suite cover this pattern.

#### 2. Selector Pattern Mismatch
**TypeScript Implementation:**
```typescript
// Two complementary selectors:
'TSNonNullExpression > ChainExpression' // Catches (chain)!
'ChainExpression > TSNonNullExpression' // Catches chain!
```

**Go Implementation:**
```go
// Single listener on NonNullExpression only
ast.KindNonNullExpression: func(node *ast.Node) {
    // Tries to handle both patterns in one function
}
```

**Issue:** The TypeScript implementation uses CSS-like selectors that precisely target the parent-child relationships. The Go version uses a single listener approach that may miss certain AST structures.

**Impact:** The dual-selector approach in TypeScript ensures comprehensive coverage of both patterns, while the Go single-listener approach may have gaps.

**Test Coverage:** All invalid test cases rely on this pattern detection.

#### 3. AST Navigation Assumptions
**TypeScript Implementation:**
```typescript
// Direct selector guarantees the relationship
const parent = node.parent as TSESTree.TSNonNullExpression;
```

**Go Implementation:**
```go
// Manual parent traversal with assumptions
func isTerminalAssertion(nonNullNode *ast.Node) bool {
    if nonNullNode.Parent == nil {
        return true
    }
    parent := nonNullNode.Parent
    // Complex logic to determine if assertion is terminal
}
```

**Issue:** The Go version makes assumptions about AST structure and parent relationships that may not hold in all cases. The TypeScript selectors provide guaranteed structural relationships.

**Impact:** May incorrectly classify some assertions as terminal or non-terminal, affecting the rule's accuracy.

**Test Coverage:** Valid test cases like `foo?.bar!.baz` and `foo?.bar!()` test the terminal detection logic.

#### 4. Reliance on `ast.IsOptionalChain()`
**TypeScript Implementation:**
```typescript
// Direct pattern matching through selectors - no need for helper functions
'ChainExpression > TSNonNullExpression'
```

**Go Implementation:**
```go
// Relies on utility function
return ast.IsOptionalChain(node)
```

**Issue:** The Go implementation depends on `ast.IsOptionalChain()` which is not defined in the provided code. Its behavior and reliability are unknown.

**Impact:** If `ast.IsOptionalChain()` doesn't correctly identify all optional chain patterns, the rule will miss violations.

**Test Coverage:** All test cases involving optional chains depend on this function.

### Recommendations
- Implement dual AST node listeners to match TypeScript's selector pattern: one for `NonNullExpression` and one for `ChainExpression`
- Verify the behavior of `ast.IsOptionalChain()` and potentially implement custom optional chain detection
- Add specific handling for parenthesized optional chains that matches the TypeScript selector `TSNonNullExpression > ChainExpression`
- Simplify the terminal assertion detection logic by leveraging AST structure more directly
- Add comprehensive test cases to verify all parenthesized optional chain patterns are caught
- Consider implementing the exact same logic flow as TypeScript: separate handlers for the two main patterns rather than trying to unify them

---