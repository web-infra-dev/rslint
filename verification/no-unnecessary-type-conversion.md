## Rule: no-unnecessary-type-conversion

### Test File: no-unnecessary-type-conversion.test.ts

### Validation Summary
- ✅ **CORRECT**: Call expression handling for String(), Number(), Boolean(), BigInt(); Basic unary plus operator handling; Core type checking logic structure
- ⚠️ **POTENTIAL ISSUES**: Missing assignment expression handling (+=); Incomplete unary operator coverage; Missing toString() method call detection; Incomplete AST listener registration; Shadow variable detection not implemented
- ❌ **INCORRECT**: Missing double negation (!!) detection; Missing double tilde (~~) detection; Missing string concatenation with empty string detection; Incomplete error reporting locations

### Discrepancies Found

#### 1. Missing Assignment Expression Handler
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
    // Handle str += '' cases
  }
}
```

**Go Implementation:**
```go
// Missing: No listener for assignment expressions in RuleListeners
```

**Issue:** The Go implementation doesn't handle assignment expressions like `str += ''` which should be flagged as unnecessary type conversions.

**Impact:** Test cases like `str += '';` will not be detected, causing false negatives.

**Test Coverage:** Tests with `str += '';` patterns will fail.

#### 2. Missing String Concatenation Detection
**TypeScript Implementation:**
```typescript
'BinaryExpression[operator = "+"]'(
  node: TSESTree.BinaryExpression,
): void {
  if (
    node.right.type === AST_NODE_TYPES.Literal &&
    node.right.value === '' &&
    doesUnderlyingTypeMatchFlag(
      services.getTypeAtLocation(node.left),
      ts.TypeFlags.StringLike,
    )
  ) {
    // Handle "string" + "" cases
  }
  // Also handles "" + "string" cases
}
```

**Go Implementation:**
```go
ast.KindBinaryExpression: func(node *ast.Node) {
  binExpr := node.AsBinaryExpression()
  if binExpr.OperatorToken.Kind == ast.KindPlusToken {
    handleStringConcatenation(ctx, node)
  }
},
```

**Issue:** The Go implementation has the `handleStringConcatenation` function but it's checking `strLit.Text` instead of the actual string value, and the logic may not correctly identify empty string literals.

**Impact:** Cases like `"string" + ""` and `"" + "string"` may not be properly detected.

**Test Coverage:** String concatenation test cases will likely fail.

#### 3. Missing Double Negation (!!) Detection
**TypeScript Implementation:**
```typescript
'UnaryExpression[operator = "!"] > UnaryExpression[operator = "!"]'(
  node: TSESTree.UnaryExpression,
): void {
  handleUnaryOperator(
    node,
    ts.TypeFlags.BooleanLike,
    'boolean',
    'Using !! on a boolean',
    true,
  );
}
```

**Go Implementation:**
```go
// Missing: No listener for double negation pattern
```

**Issue:** The Go implementation doesn't detect the double negation pattern `!!expr` when used on boolean types.

**Impact:** Test cases with `!!true` will not be flagged as unnecessary type conversions.

**Test Coverage:** Tests like `!!true;` will fail.

#### 4. Missing Double Tilde (~~) Detection
**TypeScript Implementation:**
```typescript
'UnaryExpression[operator = "~"] > UnaryExpression[operator = "~"]'(
  node: TSESTree.UnaryExpression,
): void {
  handleUnaryOperator(
    node,
    ts.TypeFlags.NumberLike,
    'number',
    'Using ~~ on a number',
    true,
  );
}
```

**Go Implementation:**
```go
// Missing: No listener for double tilde pattern
```

**Issue:** The Go implementation doesn't detect the double tilde pattern `~~expr` when used on number types.

**Impact:** Test cases with `~~123` will not be flagged as unnecessary type conversions.

**Test Coverage:** Tests like `~~123;` will fail.

#### 5. Missing toString() Method Call Detection
**TypeScript Implementation:**
```typescript
'CallExpression > MemberExpression.callee > Identifier[name = "toString"].property'(
  node: TSESTree.Expression,
): void {
  const memberExpr = node.parent as TSESTree.MemberExpression;
  const type = getConstrainedTypeAtLocation(services, memberExpr.object);
  if (doesUnderlyingTypeMatchFlag(type, ts.TypeFlags.StringLike)) {
    // Handle string.toString() cases
  }
}
```

**Go Implementation:**
```go
// The handleToStringCall function exists but is never called
// Missing: No listener registration for toString() method calls
```

**Issue:** The Go implementation has the `handleToStringCall` function but doesn't register a listener to detect toString() method calls on strings.

**Impact:** Test cases like `'asdf'.toString()` will not be detected.

**Test Coverage:** Tests with `.toString()` method calls will fail.

#### 6. Incomplete Type Flag Checking
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
  if typ == nil {
    return false
  }
  
  return utils.Every(utils.UnionTypeParts(typ), func(t *checker.Type) bool {
    return utils.Some(utils.IntersectionTypeParts(t), func(t *checker.Type) bool {
      return utils.IsTypeFlagSet(t, typeFlag)
    })
  })
}
```

**Issue:** The Go implementation checks intersection types with `utils.Some` instead of checking the type flags directly on union constituents. This logic doesn't match the TypeScript version.

**Impact:** Type checking may produce incorrect results, leading to false positives or negatives.

**Test Coverage:** Complex type scenarios may not work correctly.

#### 7. Shadow Variable Detection Missing
**TypeScript Implementation:**
```typescript
const scope = context.sourceCode.getScope(node);
const variable = scope.set.get(nodeCallee.name);
if (
  !!variable?.defs.length ||
  !doesUnderlyingTypeMatchFlag(
    getConstrainedTypeAtLocation(services, node.arguments[0]),
    typeFlag,
  )
) {
  return;
}
```

**Go Implementation:**
```go
// For now, skip symbol checking to get basic functionality working
// TODO: Add proper shadowing detection later
_ = ctx.TypeChecker.GetSymbolAtLocation(callee)
```

**Issue:** The Go implementation explicitly skips shadow variable detection, which is needed to avoid flagging cases where built-in constructors are shadowed by user-defined functions.

**Impact:** Cases where `String`, `Number`, etc. are redefined by user code may still be flagged incorrectly.

**Test Coverage:** Tests with redefined constructor functions may produce false positives.

#### 8. Incorrect String Literal Value Access
**TypeScript Implementation:**
```typescript
node.right.value === ''
```

**Go Implementation:**
```go
if strLit.Text == "" {
```

**Issue:** The Go implementation checks `strLit.Text` which includes quotes, but should check the actual string value without quotes.

**Impact:** Empty string detection will fail because `strLit.Text` for `""` would be `'""'` not `""`.

**Test Coverage:** All string concatenation tests will fail.

### Recommendations
- Add missing AST listeners for assignment expressions, double negation, double tilde, and toString() method calls
- Fix the string literal value checking to use the actual string content, not the raw text with quotes
- Implement proper shadow variable detection for built-in constructors
- Correct the type flag checking logic to match the TypeScript implementation
- Add comprehensive AST pattern matching for all the cases handled in the TypeScript version
- Fix error reporting locations to match the expected column positions from tests
- Implement proper parentheses preservation logic for fix suggestions

---