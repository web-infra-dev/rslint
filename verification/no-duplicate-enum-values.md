## Rule: no-duplicate-enum-values

### Test File: no-duplicate-enum-values.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core duplicate detection logic structure
  - Skip behavior for members without initializers 
  - Enum member iteration pattern
  - Basic string/numeric literal handling pattern
  - Error reporting with messageId structure

- ⚠️ **POTENTIAL ISSUES**: 
  - Value extraction methodology differs significantly from TypeScript
  - Template literal AST node type handling
  - Type consistency in value storage and comparison

- ❌ **INCORRECT**: 
  - String literal value extraction (uses .Text instead of parsed value)
  - Numeric literal value extraction (uses .Text instead of numeric value)
  - Template literal AST node kind may be incorrect
  - Missing proper value normalization for comparison

### Discrepancies Found

#### 1. Template Literal Node Kind Mismatch
**TypeScript Implementation:**
```typescript
function isStaticTemplateLiteral(
  node: TSESTree.Expression,
): node is TSESTree.TemplateLiteral {
  return (
    node.type === AST_NODE_TYPES.TemplateLiteral &&
    node.expressions.length === 0 &&
    node.quasis.length === 1
  );
}
```

**Go Implementation:**
```go
case ast.KindTemplateExpression:
  // Template literal - only handle static templates (no expressions)
  templateExpr := enumMember.Initializer.AsTemplateExpression()
  if templateExpr.TemplateSpans == nil || len(templateExpr.TemplateSpans.Nodes) == 0 {
    // Static template literal with no expressions
    value = templateExpr.Head.Text
  } else {
    // Skip template literals with expressions
    continue
  }
```

**Issue:** The Go implementation checks for `ast.KindTemplateExpression` but should likely check for `ast.KindTemplateLiteral` or similar. Template expressions typically contain interpolations, while template literals are the base form.

**Impact:** This could cause the rule to miss static template literals like `` `A` `` or incorrectly handle template expressions with interpolations.

**Test Coverage:** Test cases with `` `A` `` template literals may not be properly detected.

#### 2. Numeric Value Comparison Using Text
**TypeScript Implementation:**
```typescript
function isNumberLiteral(
  node: TSESTree.Expression,
): node is TSESTree.NumberLiteral {
  return (
    node.type === AST_NODE_TYPES.Literal && typeof node.value === 'number'
  );
}
// Later: value = member.initializer.value;
```

**Go Implementation:**
```go
case ast.KindNumericLiteral:
  // Numeric literal
  numericLiteral := enumMember.Initializer.AsNumericLiteral()
  value = numericLiteral.Text
```

**Issue:** The Go implementation uses the text representation of numeric literals instead of the parsed numeric value. This means `1` and `1.0` would be treated as different values, when they should be considered the same.

**Impact:** False negatives where numerically equivalent values with different text representations (e.g., `1` vs `1.0`, `0` vs `-0`) are not detected as duplicates.

**Test Coverage:** Test cases with `A = 0, B = -0` might not behave correctly, though the TypeScript rule mentions this as a valid case, suggesting `-0` and `0` should be treated as different.

#### 3. String Literal Text vs Value Extraction
**TypeScript Implementation:**
```typescript
if (isStringLiteral(member.initializer)) {
  value = member.initializer.value;
}
```

**Go Implementation:**
```go
case ast.KindStringLiteral:
  // String literal
  stringLiteral := enumMember.Initializer.AsStringLiteral()
  value = stringLiteral.Text
```

**Issue:** The Go implementation uses `.Text` while TypeScript uses `.value`. The `.Text` property might include quotes or escape sequences, whereas `.value` is the parsed string value.

**Impact:** String comparison might fail for strings with escape sequences or special characters.

**Test Coverage:** Test cases with escaped strings like `"quote\"here"` may not work correctly.

#### 4. Template Literal Value Extraction
**TypeScript Implementation:**
```typescript
} else if (isStaticTemplateLiteral(member.initializer)) {
  value = member.initializer.quasis[0].value.cooked;
}
```

**Go Implementation:**
```go
value = templateExpr.Head.Text
```

**Issue:** The Go implementation uses `.Text` from the head, while TypeScript uses `.value.cooked` from the quasi. The cooked value represents the processed template string content, while `.Text` might include backticks.

**Impact:** Template literal values might not be properly extracted, leading to incorrect duplicate detection.

**Test Coverage:** Test cases comparing template literals with string literals (like `A = 'A', B = \`A\``) may fail.

### Recommendations
- Verify the correct AST node kind for template literals in the typescript-go AST
- Use parsed numeric values instead of text representation for numeric literals
- Investigate proper string value extraction (without quotes) for string literals
- Ensure template literal content extraction matches the processed/cooked value
- Add debug logging to verify that values are being extracted correctly during testing
- Consider adding test cases for edge cases like escaped strings and different numeric representations

---