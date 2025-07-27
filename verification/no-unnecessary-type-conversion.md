## Rule: no-unnecessary-type-conversion

### Test File: no-unnecessary-type-conversion.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic type conversion detection for String(), Number(), Boolean(), BigInt() calls
- ✅ **CORRECT**: Type flag checking for primitive types
- ✅ **CORRECT**: Basic string concatenation handling
- ⚠️ **POTENTIAL ISSUES**: Several critical AST pattern matching issues and missing functionality
- ❌ **INCORRECT**: Missing key features including toString calls, assignment operators, double operators, symbol shadowing

### Discrepancies Found

#### 1. Missing toString() Method Call Detection
**TypeScript Implementation:**
```typescript
'CallExpression > MemberExpression.callee > Identifier[name = "toString"].property'(
  node: TSESTree.Expression,
): void {
  const memberExpr = node.parent as TSESTree.MemberExpression;
  const type = getConstrainedTypeAtLocation(services, memberExpr.object);
  if (doesUnderlyingTypeMatchFlag(type, ts.TypeFlags.StringLike)) {
    // Report unnecessary toString() call
  }
}
```

**Go Implementation:**
```go
// handleToStringCall function exists but is never called
func handleToStringCall(ctx rule.RuleContext, node *ast.Node) {
  // Implementation exists but no listener registered
}
```

**Issue:** The Go implementation has the `handleToStringCall` function but no AST listener is registered to detect `obj.toString()` calls.

**Impact:** Test cases like `"'asdf'.toString();"` will not be caught by the Go version.

**Test Coverage:** Affects test cases checking for unnecessary `.toString()` calls on strings.

#### 2. Missing Assignment Expression Handler (`+=` operator)
**TypeScript Implementation:**
```typescript
'AssignmentExpression[operator = "+="]'(
  node: TSESTree.AssignmentExpression,
): void {
  if (
    node.right.type === AST_NODE_TYPES.Literal &&
    node.right.value === '' &&
    doesUnderlyingTypeMatchFlag(
      services.getTypeAtLocation(node.left),
      ts.TypeFlags.StringLike,
    )
  ) {
    // Report unnecessary string concatenation assignment
  }
}
```

**Go Implementation:**
```go
// handleStringConcatenationAssignment exists but no listener is registered
// Only handles binary expressions, not assignment expressions
```

**Issue:** The Go version doesn't register a listener for assignment expressions (`str += ''`).

**Impact:** Test cases like `str += '';` will not be detected.

**Test Coverage:** Affects assignment expression test cases.

#### 3. Missing Double Operator Detection (`!!` and `~~`)
**TypeScript Implementation:**
```typescript
'UnaryExpression[operator = "!"] > UnaryExpression[operator = "!"]'(
  node: TSESTree.UnaryExpression,
): void {
  handleUnaryOperator(node, ts.TypeFlags.BooleanLike, 'boolean', 'Using !! on a boolean', true);
}

'UnaryExpression[operator = "~"] > UnaryExpression[operator = "~"]'(
  node: TSESTree.UnaryExpression,
): void {
  handleUnaryOperator(node, ts.TypeFlags.NumberLike, 'number', 'Using ~~ on a number', true);
}
```

**Go Implementation:**
```go
// handleDoubleNegation and handleDoubleTilde exist but no listeners registered
// Only handles single unary plus operator
ast.KindPrefixUnaryExpression: func(node *ast.Node) {
  unaryExpr := node.AsPrefixUnaryExpression()
  if unaryExpr.Operator == ast.KindPlusToken {
    handleUnaryPlus(ctx, node)
  }
},
```

**Issue:** Missing listeners for double negation (`!!`) and double tilde (`~~`) operators.

**Impact:** Test cases like `!!true;` and `~~123;` will not be detected.

**Test Coverage:** Affects double operator test cases.

#### 4. Incorrect Type Flag Checking Logic
**TypeScript Implementation:**
```typescript
function doesUnderlyingTypeMatchFlag(
  type: ts.Type,
  typeFlag: ts.TypeFlags,
): boolean {
  return tsutils
    .unionConstituents(type)
    .every(t => isTypeFlagSet(t, typeFlag));
}
```

**Go Implementation:**
```go
func doesUnderlyingTypeMatchFlag(ctx rule.RuleContext, typ *checker.Type, typeFlag checker.TypeFlags) bool {
  return utils.Every(utils.UnionTypeParts(typ), func(t *checker.Type) bool {
    return utils.Some(utils.IntersectionTypeParts(t), func(t *checker.Type) bool {
      return utils.IsTypeFlagSet(t, typeFlag)
    })
  })
}
```

**Issue:** The Go version uses `utils.Some(utils.IntersectionTypeParts(...))` which changes the logic. The TypeScript version directly checks if each union constituent has the type flag, while Go checks if any intersection part has the flag.

**Impact:** This could lead to false positives or negatives in type checking.

**Test Coverage:** May affect all type-sensitive test cases.

#### 5. Missing Symbol Shadowing Detection
**TypeScript Implementation:**
```typescript
const scope = context.sourceCode.getScope(node);
const variable = scope.set.get(nodeCallee.name);
if (!!variable?.defs.length) {
  return; // Skip if the global is shadowed
}
```

**Go Implementation:**
```go
// For now, skip symbol checking to get basic functionality working
// TODO: Add proper shadowing detection later
_ = ctx.TypeChecker.GetSymbolAtLocation(callee)
```

**Issue:** The Go version has a TODO comment and skips symbol shadowing detection entirely.

**Impact:** Will incorrectly flag user-defined functions named `String`, `Number`, etc. as unnecessary type conversions.

**Test Coverage:** Affects test cases with shadowed global constructors.

#### 6. Missing `suggestSatisfies` Suggestion
**TypeScript Implementation:**
```typescript
suggest: [
  {
    messageId: 'suggestRemove',
    fix: getWrappingFixer(wrappingFixerParams),
  },
  {
    messageId: 'suggestSatisfies',
    data: { type: typeString },
    fix: getWrappingFixer({
      ...wrappingFixerParams,
      wrap: expr => `${expr} satisfies ${typeString}`,
    }),
  },
],
```

**Go Implementation:**
```go
ctx.ReportNodeWithSuggestions(callee, rule.RuleMessage{
  Id:          "unnecessaryTypeConversion",
  Description: message,
}, rule.RuleSuggestion{
  Message: rule.RuleMessage{
    Id:          "suggestRemove",
    Description: "Remove the type conversion.",
  },
  // Only provides "remove" suggestion, missing "satisfies" suggestion
})
```

**Issue:** The Go version only provides one suggestion (remove) but the TypeScript version provides two (remove and satisfies).

**Impact:** Users won't see the type assertion suggestion alternative.

**Test Coverage:** All test cases expect both suggestions.

#### 7. Wrong Type Flag Used for String Detection
**TypeScript Implementation:**
```typescript
if (doesUnderlyingTypeMatchFlag(type, ts.TypeFlags.StringLike)) {
```

**Go Implementation:**
```go
if !doesUnderlyingTypeMatchFlag(ctx, objType, checker.TypeFlagsString) {
```

**Issue:** The Go version uses `TypeFlagsString` instead of `TypeFlagsStringLike`, which may not include string literal types.

**Impact:** May miss some string types in type checking.

**Test Coverage:** Could affect string-related test cases.

#### 8. Incorrect String Literal Check
**TypeScript Implementation:**
```typescript
if (
  node.right.type === AST_NODE_TYPES.Literal &&
  node.right.value === '' &&
  // ...
) {
```

**Go Implementation:**
```go
if right.Kind == ast.KindStringLiteral {
  strLit := right.AsStringLiteral()
  if strLit.Text == "" {
```

**Issue:** The Go version checks `strLit.Text` instead of the literal value. `Text` may include quotes.

**Impact:** May not correctly identify empty string literals.

**Test Coverage:** Affects empty string concatenation test cases.

### Recommendations
- Add missing AST listeners for toString calls, assignment expressions, and double operators
- Implement proper symbol shadowing detection to avoid false positives
- Add the missing "suggestSatisfies" suggestions for all rule violations
- Fix the type flag checking logic to match TypeScript behavior exactly
- Use `TypeFlagsStringLike` instead of `TypeFlagsString` for consistency
- Correct string literal value checking logic
- Add comprehensive test coverage for all missing functionality
- Review the `doesUnderlyingTypeMatchFlag` implementation for logical correctness

---