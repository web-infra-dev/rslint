## Rule: no-non-null-assertion

### Test File: no-non-null-assertion.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core rule logic correctly identifies non-null assertions (`!`) and reports them
  - Message IDs and descriptions match the TypeScript implementation
  - Basic suggestion logic for property access, element access, and call expressions
  - Proper handling of optional chaining detection for different parent node types
  - Correct suggestion generation for cases with and without existing optional chaining

- ⚠️ **POTENTIAL ISSUES**: 
  - Text range calculation for the exclamation mark position may not account for all edge cases
  - Multi-line formatting and comment preservation in suggestions needs verification
  - Complex nested scenarios with multiple non-null assertions may not be handled identically

- ❌ **INCORRECT**: 
  - No critical functional discrepancies identified in core logic

### Discrepancies Found

#### 1. Token Position Calculation Approach
**TypeScript Implementation:**
```typescript
const nonNullOperator = nullThrows(
  context.sourceCode.getTokenAfter(
    node.expression,
    isNonNullAssertionPunctuator,
  ),
  NullThrowsReasons.MissingToken('!', 'expression'),
);
```

**Go Implementation:**
```go
nonNullEnd := node.End()
exclamationStart := nonNullEnd - 1
exclamationRange := core.NewTextRange(exclamationStart, nonNullEnd)
```

**Issue:** The TypeScript implementation uses AST token navigation to find the exact `!` token, while the Go implementation assumes the `!` is the last character before `node.End()`. This could potentially cause issues if there are whitespace or formatting differences.

**Impact:** May affect the precision of fix suggestions, especially in complex formatting scenarios.

**Test Coverage:** Multi-line test cases like the ones with comments between `x!` and `.y` may reveal positioning issues.

#### 2. Multi-line and Comment Handling
**TypeScript Implementation:**
```typescript
fix(fixer) {
  const punctuator = nullThrows(
    context.sourceCode.getTokenAfter(nonNullOperator),
    NullThrowsReasons.MissingToken('.', '!'),
  );
  return [
    fixer.remove(nonNullOperator),
    fixer.insertTextBefore(punctuator, '?'),
  ];
}
```

**Go Implementation:**
```go
case ast.KindPropertyAccessExpression:
  // x!.y -> x?.y (replace ! with ? since . is already there)
  return rule.RuleFixReplaceRange(exclamationRange, "?")
```

**Issue:** The TypeScript implementation specifically handles the case where it needs to find the next punctuator (`.`) and insert `?` before it, preserving any comments or whitespace between `!` and `.`. The Go implementation assumes it can simply replace `!` with `?`, which may not work correctly when there are comments or multi-line formatting between the `!` and `.`.

**Impact:** Fix suggestions may not preserve formatting and comments correctly in multi-line scenarios.

**Test Coverage:** Test cases with comments between `x!` and `.y` will likely fail with incorrect fix suggestions.

#### 3. Element Access vs Property Access Fix Logic
**TypeScript Implementation:**
```typescript
if (node.parent.computed) {
  // it is x![y]?.z
  suggest.push({
    messageId: 'suggestOptionalChain',
    fix: replaceTokenWithOptional(),
  });
} else {
  // it is x!.y?.z
  suggest.push({
    messageId: 'suggestOptionalChain',
    fix(fixer) {
      const punctuator = nullThrows(
        context.sourceCode.getTokenAfter(nonNullOperator),
        NullThrowsReasons.MissingToken('.', '!'),
      );
      return [
        fixer.remove(nonNullOperator),
        fixer.insertTextBefore(punctuator, '?'),
      ];
    },
  });
}
```

**Go Implementation:**
```go
switch parent.Kind {
case ast.KindPropertyAccessExpression:
  // x!.y -> x?.y (replace ! with ? since . is already there)
  return rule.RuleFixReplaceRange(exclamationRange, "?")
default:
  // x![y] -> x?.[y] or x!() -> x?.() (replace ! with ?.)
  return rule.RuleFixReplaceRange(exclamationRange, "?.")
}
```

**Issue:** The TypeScript implementation distinguishes between computed (`x![y]`) and non-computed (`x!.y`) member expressions within the same parent type, while the Go implementation handles them as separate AST node types. This difference in approach could lead to different fix suggestions.

**Impact:** The logic should still work correctly, but the approach is fundamentally different.

**Test Coverage:** Both property access (`x!.y`) and element access (`x![y]`) test cases should verify the correctness.

### Recommendations
- Verify that the simple character position calculation (`node.End() - 1`) correctly identifies the `!` token in all formatting scenarios
- Test multi-line cases thoroughly to ensure fix suggestions preserve comments and whitespace
- Consider implementing more robust token finding similar to the TypeScript implementation if positioning issues arise
- Add specific test cases for edge formatting scenarios to validate fix accuracy
- Ensure the AST node type mapping between TypeScript's `computed` property and Go's separate node types produces equivalent results

---