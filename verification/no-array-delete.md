## Rule: no-array-delete

### Test File: no-array-delete.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core rule logic correctly identifies delete expressions on array-typed values
  - Type checking logic properly handles union types (all members must be arrays)
  - Type checking logic properly handles intersection types (any member can be array)
  - Basic AST pattern matching for DeleteExpression -> ElementAccessExpression
  - Error message IDs and descriptions match exactly
  - Suggestion message matches exactly
  - Basic fix generation structure is present

- ⚠️ **POTENTIAL ISSUES**: 
  - Comment handling approach differs significantly from TypeScript version
  - Parentheses handling for complex expressions may differ
  - Token range calculation approach is different

- ❌ **INCORRECT**: 
  - Missing support for MemberExpression (property access) patterns
  - Fix generation logic is fundamentally different and likely incorrect
  - No handling of SequenceExpression parentheses requirement
  - Missing comment preservation in fixes

### Discrepancies Found

#### 1. AST Pattern Matching Scope
**TypeScript Implementation:**
```typescript
'UnaryExpression[operator="delete"]'(node: TSESTree.UnaryExpression): void {
  const { argument } = node;
  if (argument.type !== AST_NODE_TYPES.MemberExpression) {
    return;
  }
  // ... continue processing
}
```

**Go Implementation:**
```go
ast.KindDeleteExpression: func(node *ast.Node) {
  // ...
  if !ast.IsElementAccessExpression(deleteExpression) {
    return;
  }
  // ... continue processing
}
```

**Issue:** The Go version only handles ElementAccessExpression (arr[index]) but the TypeScript version handles all MemberExpression types, which includes both ElementAccess and PropertyAccess patterns.

**Impact:** This is actually correct behavior - the rule should only trigger on bracket notation (arr[0]) not dot notation (obj.prop), as evidenced by the test cases.

**Test Coverage:** All test cases use bracket notation, confirming this is correct.

#### 2. Fix Generation Strategy
**TypeScript Implementation:**
```typescript
fix(fixer): TSESLint.RuleFix | null {
  const { object, property } = argument;
  const shouldHaveParentheses = property.type === AST_NODE_TYPES.SequenceExpression;
  const nodeMap = services.esTreeNodeToTSNodeMap;
  const target = nodeMap.get(object).getText();
  const rawKey = nodeMap.get(property).getText();
  const key = shouldHaveParentheses ? `(${rawKey})` : rawKey;
  let suggestion = `${target}.splice(${key}, 1)`;
  // ... comment handling
  return fixer.replaceText(node, suggestion);
}
```

**Go Implementation:**
```go
ctx.ReportNodeWithSuggestions(node, buildNoArrayDeleteMessage(), rule.RuleSuggestion{
  Message: buildUseSpliceMessage(),
  FixesArr: []rule.RuleFix{
    rule.RuleFixRemoveRange(deleteTokenRange),
    rule.RuleFixReplaceRange(leftBracketTokenRange, ".splice("),
    rule.RuleFixReplaceRange(rightBracketTokenRange, ", 1)"),
  },
})
```

**Issue:** The Go version uses multiple targeted range replacements instead of replacing the entire expression. This approach is fundamentally different and may not handle complex expressions correctly.

**Impact:** The piecemeal replacement approach may fail for complex expressions, nested parentheses, or expressions with comments.

**Test Coverage:** Test cases with complex expressions like `delete a[(b + 1) * (b + 2)]` and `delete arr[(doWork(), 1)]` would reveal this issue.

#### 3. SequenceExpression Parentheses Handling
**TypeScript Implementation:**
```typescript
const shouldHaveParentheses = property.type === AST_NODE_TYPES.SequenceExpression;
const key = shouldHaveParentheses ? `(${rawKey})` : rawKey;
```

**Go Implementation:**
```go
// No equivalent logic for SequenceExpression parentheses
```

**Issue:** The Go version doesn't check if the array index is a SequenceExpression that needs parentheses preservation.

**Impact:** For code like `delete arr[(doWork(), 1)]`, the fix would generate `arr.splice(doWork(), 1, 1)` instead of the correct `arr.splice((doWork(), 1), 1)`.

**Test Coverage:** The test case with `delete arr[(doWork(), 1)]` expects the output `arr.splice((doWork(), 1), 1)` which the Go version cannot produce correctly.

#### 4. Comment Preservation
**TypeScript Implementation:**
```typescript
const comments = context.sourceCode.getCommentsInside(node);
if (comments.length > 0) {
  const indentationCount = node.loc.start.column;
  const indentation = ' '.repeat(indentationCount);
  const commentsText = comments
    .map(comment => {
      return comment.type === AST_TOKEN_TYPES.Line
        ? `//${comment.value}`
        : `/*${comment.value}*/`;
    })
    .join(`\n${indentation}`);
  suggestion = `${commentsText}\n${indentation}${suggestion}`;
}
```

**Go Implementation:**
```go
// No comment preservation logic
```

**Issue:** The Go version completely lacks comment preservation during fix generation.

**Impact:** Comments within delete expressions will be lost during fixes, violating the expected behavior shown in test cases.

**Test Coverage:** The complex comment test case demonstrates this requirement clearly.

#### 5. Parentheses Stripping Logic
**TypeScript Implementation:**
```typescript
// Uses ESTree AST directly without explicit parentheses stripping
const { object, property } = argument;
```

**Go Implementation:**
```go
deleteExpression := ast.SkipParentheses(node.AsDeleteExpression().Expression)
```

**Issue:** The Go version strips parentheses from the entire expression, which may affect the fix generation for complex nested parentheses.

**Impact:** May not handle cases like `delete ((a[((b))]))` correctly in fix generation.

**Test Coverage:** The test case `delete ((a[((b))]))` expects output `a.splice(b, 1)` which tests this behavior.

### Recommendations
- **Fix the fix generation strategy**: Replace the multi-range approach with a single text replacement that properly reconstructs the splice call
- **Add SequenceExpression detection**: Implement logic to detect when parentheses are needed around the array index
- **Implement comment preservation**: Add logic to extract and preserve comments within the delete expression
- **Improve complex expression handling**: Ensure the fix generation works correctly for nested expressions and parentheses
- **Add comprehensive fix testing**: The current test approach validates the core logic but may miss fix generation edge cases

---