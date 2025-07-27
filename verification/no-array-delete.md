## Rule: no-array-delete

### Test File: no-array-delete.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core array/tuple type detection logic matches
  - Union and intersection type handling is equivalent
  - Basic delete expression detection works
  - Element access expression validation
  - Error message IDs are consistent
  - Suggestion mechanism is implemented
- ⚠️ **POTENTIAL ISSUES**: 
  - Fix generation approach differs significantly (token-based vs text replacement)
  - Comment preservation logic appears incomplete in Go version
  - Parentheses handling for complex expressions may differ
- ❌ **INCORRECT**: 
  - Missing member expression detection (TypeScript checks `AST_NODE_TYPES.MemberExpression`, Go only checks element access)
  - Go version doesn't handle property access syntax (`obj.prop`) deletion
  - Sequence expression parentheses logic not implemented in Go

### Discrepancies Found

#### 1. Missing Member Expression Support
**TypeScript Implementation:**
```typescript
if (argument.type !== AST_NODE_TYPES.MemberExpression) {
  return;
}
```

**Go Implementation:**
```go
if !ast.IsElementAccessExpression(deleteExpression) {
  return;
}
```

**Issue:** The TypeScript version checks for any `MemberExpression` (which includes both property access `obj.prop` and element access `obj[key]`), while the Go version only checks for `ElementAccessExpression` (bracket notation).

**Impact:** The Go version would miss cases like `delete arr.length` or other property access patterns that should be caught.

**Test Coverage:** This discrepancy would affect edge cases involving property access on arrays, though the current test suite focuses on element access patterns.

#### 2. Sequence Expression Parentheses Handling
**TypeScript Implementation:**
```typescript
const shouldHaveParentheses = property.type === AST_NODE_TYPES.SequenceExpression;
const key = shouldHaveParentheses ? `(${rawKey})` : rawKey;
```

**Go Implementation:**
```go
// No equivalent logic for sequence expression parentheses
```

**Issue:** The Go version doesn't check if the argument expression is a sequence expression that needs parentheses in the fix suggestion.

**Impact:** For test cases like `delete arr[(doWork(), 1)]`, the Go version might generate incorrect fix suggestions without proper parentheses.

**Test Coverage:** Test case with `delete arr[(doWork(), 1)]` expects output `arr.splice((doWork(), 1), 1)` - the parentheses must be preserved.

#### 3. Comment Preservation Logic Missing
**TypeScript Implementation:**
```typescript
const comments = context.sourceCode.getCommentsInside(node);
if (comments.length > 0) {
  const indentationCount = node.loc.start.column;
  const indentation = ' '.repeat(indentationCount);
  const commentsText = comments.map(comment => {
    return comment.type === AST_TOKEN_TYPES.Line
      ? `//${comment.value}`
      : `/*${comment.value}*/`;
  }).join(`\n${indentation}`);
  suggestion = `${commentsText}\n${indentation}${suggestion}`;
}
```

**Go Implementation:**
```go
// No comment preservation logic implemented
```

**Issue:** The Go version doesn't preserve comments when generating fix suggestions, while the TypeScript version has sophisticated comment handling.

**Impact:** Complex test cases with embedded comments will fail, such as the test with `/* multi line */` and `// single-line` comments.

**Test Coverage:** The `noFormat` test case with extensive comments expects all comments to be preserved and properly formatted in the output.

#### 4. Token-Based vs Text-Based Fix Generation
**TypeScript Implementation:**
```typescript
return fixer.replaceText(node, suggestion);
```

**Go Implementation:**
```go
FixesArr: []rule.RuleFix{
  rule.RuleFixRemoveRange(deleteTokenRange),
  rule.RuleFixReplaceRange(leftBracketTokenRange, ".splice("),
  rule.RuleFixReplaceRange(rightBracketTokenRange, ", 1)"),
}
```

**Issue:** The approaches differ fundamentally - TypeScript builds a complete replacement string, while Go uses multiple targeted token replacements.

**Impact:** The Go approach may be more fragile with complex expressions, parentheses, and whitespace preservation.

**Test Coverage:** Complex expressions with nested parentheses like `delete ((a[((b))]))` may not generate clean output.

#### 5. AST Node Navigation Differences
**TypeScript Implementation:**
```typescript
const { object, property } = argument;
```

**Go Implementation:**
```go
expression := deleteExpression.AsElementAccessExpression()
// Uses expression.Expression and expression.ArgumentExpression
```

**Issue:** The property access patterns differ between the implementations, potentially affecting how complex member expressions are handled.

**Impact:** Edge cases with deeply nested property access might behave differently.

**Test Coverage:** Test cases like `delete obj.a.b.c[0]` rely on proper AST navigation.

### Recommendations
- Add support for `MemberExpression` detection in addition to `ElementAccessExpression`
- Implement sequence expression parentheses logic for fix suggestions
- Add comprehensive comment preservation during fix generation
- Consider switching to text-based fix generation for better consistency with TypeScript-ESLint
- Add test cases specifically for property access deletion patterns
- Validate that all complex expression test cases generate correct fix suggestions
- Ensure whitespace and formatting preservation matches TypeScript-ESLint output exactly

---