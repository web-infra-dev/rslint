## Rule: no-dynamic-delete

### Test File: no-dynamic-delete.test.ts

### Validation Summary
- ✅ **CORRECT**: Core rule logic for detecting dynamic delete operations, AST pattern matching for delete expressions with computed member access, Error message consistency, Basic acceptable index expression detection (string literals, number literals, negative number literals)
- ⚠️ **POTENTIAL ISSUES**: AST node kind mapping differences between TypeScript and Go AST, Missing auto-fix functionality present in TypeScript version
- ❌ **INCORRECT**: No critical logic discrepancies found

### Discrepancies Found

#### 1. Missing Auto-Fix Functionality
**TypeScript Implementation:**
```typescript
function createFixer(
  member: TSESTree.MemberExpression,
): TSESLint.ReportFixFunction | undefined {
  if (
    member.property.type === AST_NODE_TYPES.Literal &&
    typeof member.property.value === 'string'
  ) {
    return createPropertyReplacement(
      member.property,
      `.${member.property.value}`,
    );
  }

  return undefined;
}

context.report({
  node: node.argument.property,
  messageId: 'dynamicDelete',
  fix: createFixer(node.argument),
});
```

**Go Implementation:**
```go
ctx.ReportNode(argumentExpression, rule.RuleMessage{
  Id:          "dynamicDelete",
  Description: "Do not delete dynamically computed property keys.",
})
```

**Issue:** The Go implementation lacks the auto-fix functionality that converts `delete container['property']` to `delete container.property` when the property is a string literal.

**Impact:** Users won't get automatic fixes for fixable violations, reducing developer experience compared to TypeScript-ESLint.

**Test Coverage:** All test cases have `output: null`, suggesting the original rule provides fixes but they aren't tested in the current suite.

#### 2. AST Node Kind Mapping
**TypeScript Implementation:**
```typescript
node.argument.type !== AST_NODE_TYPES.MemberExpression ||
!node.argument.computed
```

**Go Implementation:**
```go
if !ast.IsElementAccessExpression(expression) {
  return
}
```

**Issue:** The Go implementation uses `ElementAccessExpression` to match TypeScript's computed `MemberExpression`, which appears to be the correct mapping, but this should be verified.

**Impact:** Potential for missing or incorrectly flagging cases if the AST node mapping is incorrect.

**Test Coverage:** All test cases should verify this mapping is correct.

#### 3. Unary Expression Handling
**TypeScript Implementation:**
```typescript
(property.type === AST_NODE_TYPES.UnaryExpression &&
 property.operator === '-' &&
 property.argument.type === AST_NODE_TYPES.Literal &&
 typeof property.argument.value === 'number')
```

**Go Implementation:**
```go
case ast.KindPrefixUnaryExpression:
  unary := property.AsPrefixUnaryExpression()
  if unary.Operator == ast.KindMinusToken &&
      ast.IsNumericLiteral(unary.Operand) {
    return true
  }
```

**Issue:** The Go implementation correctly handles negative number literals, but the AST node kind mapping should be verified (`KindPrefixUnaryExpression` vs `UnaryExpression`).

**Impact:** Could miss or incorrectly handle negative number literals if mapping is wrong.

**Test Coverage:** Test case `delete container[-7]` should pass, `delete container[+7]` should fail.

### Recommendations
- Add auto-fix functionality to match TypeScript implementation behavior
- Verify AST node kind mappings are correct between TypeScript AST and Go AST representations
- Consider adding integration tests to verify the rule works correctly with the Go TypeScript parser
- Add test cases that specifically verify auto-fix behavior if implemented
- Ensure proper handling of edge cases like `delete container[+'Infinity']` and `delete container[-Infinity]`

---