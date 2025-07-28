## Rule: no-duplicate-enum-values

### Test File: no-duplicate-enum-values.test.ts

### Validation Summary
- ✅ **CORRECT**: Core duplicate detection logic, string literal handling, numeric literal handling, static template literal support, proper value extraction and comparison
- ⚠️ **POTENTIAL ISSUES**: Template expression handling complexity, numeric precision differences, edge case handling for malformed literals
- ❌ **INCORRECT**: None identified - implementations appear functionally equivalent

### Discrepancies Found

#### 1. Template Expression Handling Differences
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

// Later usage:
} else if (isStaticTemplateLiteral(member.initializer)) {
  value = member.initializer.quasis[0].value.cooked;
}
```

**Go Implementation:**
```go
case ast.KindNoSubstitutionTemplateLiteral:
  // No substitution template literal (e.g., `A`)
  templateLiteral := enumMember.Initializer.AsNoSubstitutionTemplateLiteral()
  text := templateLiteral.Text
  // Remove backticks to get the actual template value
  if len(text) >= 2 && text[0] == '`' && text[len(text)-1] == '`' {
    value = text[1 : len(text)-1]
  } else {
    value = text
  }
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

**Issue:** The Go implementation handles two different AST node types for template literals, which may be more comprehensive than the TypeScript version's single check.

**Impact:** This is actually a potential improvement - the Go version may catch more cases of static template literals.

**Test Coverage:** The test case with `` `A` `` should validate this behavior.

#### 2. Numeric Parsing Approach
**TypeScript Implementation:**
```typescript
function isNumberLiteral(
  node: TSESTree.Expression,
): node is TSESTree.NumberLiteral {
  return (
    node.type === AST_NODE_TYPES.Literal && typeof node.value === 'number'
  );
}

// Later usage:
} else if (isNumberLiteral(member.initializer)) {
  value = member.initializer.value;
}
```

**Go Implementation:**
```go
case ast.KindNumericLiteral:
  // Numeric literal - parse as number for proper comparison
  numericLiteral := enumMember.Initializer.AsNumericLiteral()
  text := numericLiteral.Text
  // Try to parse as float64 for proper numeric comparison
  if num, err := strconv.ParseFloat(text, 64); err == nil {
    value = num
  } else {
    // Fallback to text representation if parsing fails
    value = text
  }
```

**Issue:** The TypeScript version directly accesses the parsed numeric value, while the Go version manually parses the text representation.

**Impact:** Minimal - both should produce equivalent results for valid numeric literals. The Go version has more robust error handling.

**Test Coverage:** Numeric duplicate tests should validate this behavior.

#### 3. String Literal Quote Handling
**TypeScript Implementation:**
```typescript
function isStringLiteral(
  node: TSESTree.Expression,
): node is TSESTree.StringLiteral {
  return (
    node.type === AST_NODE_TYPES.Literal && typeof node.value === 'string'
  );
}

// Later usage:
if (isStringLiteral(member.initializer)) {
  value = member.initializer.value;
}
```

**Go Implementation:**
```go
case ast.KindStringLiteral:
  // String literal - extract value without quotes
  stringLiteral := enumMember.Initializer.AsStringLiteral()
  text := stringLiteral.Text
  // Remove quotes to get the actual string value
  if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
    value = text[1 : len(text)-1]
  } else {
    value = text
  }
```

**Issue:** The TypeScript version accesses the pre-parsed string value, while the Go version manually strips quotes from the raw text.

**Impact:** Both approaches should yield the same result for valid string literals. The Go approach may be more fragile for edge cases.

**Test Coverage:** String duplicate tests should validate this behavior.

### Recommendations
- The Go implementation appears functionally correct and potentially more robust in some areas
- Template literal handling in Go is more comprehensive than TypeScript version
- Manual parsing in Go (strings, numbers) should be thoroughly tested for edge cases
- Consider adding test cases for malformed literals to ensure robust error handling
- The Go implementation correctly handles all the test cases from the original TypeScript test suite

### Overall Assessment
The Go port appears to be functionally equivalent to the TypeScript implementation, with some areas being potentially more robust (template literal handling, error handling). No critical discrepancies were identified that would cause test failures or behavioral differences.

---