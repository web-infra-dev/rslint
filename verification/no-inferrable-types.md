## Rule: no-inferrable-types

### Test File: no-inferrable-types.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic type inference detection for primitive types (string, number, boolean, bigint)
  - Function call detection (BigInt, Number, String, Boolean, Symbol, RegExp)
  - Unary operator handling for +, -, !, void
  - Optional chaining support with skipChainExpression logic
  - Configuration options (ignoreParameters, ignoreProperties)
  - Core message ID and structure

- ⚠️ **POTENTIAL ISSUES**:
  - AST node type mappings between TypeScript and Go AST
  - Void expression handling implementation differences
  - Template literal detection logic
  - Parameter property handling in constructors
  - Error reporting position accuracy

- ❌ **INCORRECT**:
  - Missing literal type support (ast.KindLiteralType handling is incomplete)
  - Incorrect void expression detection
  - Missing proper accessor property detection
  - Incomplete template literal handling
  - Parameter visitor logic doesn't handle all function types correctly

### Discrepancies Found

#### 1. Void Expression Detection
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TSUndefinedKeyword:
  return (
    hasUnaryPrefix(init, 'void') || isIdentifier(init, 'undefined')
  );
```

**Go Implementation:**
```go
case ast.KindUndefinedKeyword:
  // Check for void expressions (void someValue)
  isVoidExpr := init.Kind == ast.KindVoidExpression
  // Check for undefined literals
  literalResult := isLiteral(init, "undefined")
  return isVoidExpr || literalResult
```

**Issue:** Go implementation checks for `ast.KindVoidExpression` directly, but should use `hasUnaryPrefix(init, "void")` to match TypeScript logic that checks for prefix unary expressions with void operator.

**Impact:** May not correctly identify `void someValue` expressions as inferrable for undefined type.

**Test Coverage:** Test case `'const a: undefined = void someValue;'` may fail.

#### 2. Template Literal Handling
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TSStringKeyword:
  return (
    isFunctionCall(init, 'String') ||
    isLiteral(init, 'string') ||
    init.type === AST_NODE_TYPES.TemplateLiteral
  );
```

**Go Implementation:**
```go
case ast.KindStringKeyword:
  return isFunctionCall(init, "String") ||
    isLiteral(init, "string") ||
    init.Kind == ast.KindTemplateExpression ||
    init.Kind == ast.KindNoSubstitutionTemplateLiteral
```

**Issue:** Go checks for both `KindTemplateExpression` and `KindNoSubstitutionTemplateLiteral`, while TypeScript only checks for `TemplateLiteral`. The mapping may be incorrect.

**Impact:** Potential false positives or negatives for template literal detection.

**Test Coverage:** Test case `'const a: string = \`str\`;'` needs verification.

#### 3. Literal Type Support
**TypeScript Implementation:**
```typescript
// No explicit literal type handling in isInferrable function
```

**Go Implementation:**
```go
case ast.KindLiteralType:
  // Handle literal types like `null`, `undefined`, boolean literals, etc.
  literalType := annotation.AsLiteralTypeNode()
  if literalType.Literal != nil {
    switch literalType.Literal.Kind {
    case ast.KindNullKeyword:
      return init.Kind == ast.KindNullKeyword
    case ast.KindTrueKeyword, ast.KindFalseKeyword:
      return init.Kind == ast.KindTrueKeyword || init.Kind == ast.KindFalseKeyword
    case ast.KindNumericLiteral:
      return init.Kind == ast.KindNumericLiteral
    case ast.KindStringLiteral:
      return init.Kind == ast.KindStringLiteral
    }
  }
  return false
```

**Issue:** Go implementation adds literal type handling that doesn't exist in the TypeScript version. This could be over-implementation or the TypeScript version handles this differently.

**Impact:** May report additional cases that TypeScript version doesn't catch.

**Test Coverage:** Need to verify if literal types like `const a: 5 = 5;` should be reported.

#### 4. Parameter Visitor Function Coverage
**TypeScript Implementation:**
```typescript
function inferrableParameterVisitor(
  node:
    | TSESTree.ArrowFunctionExpression
    | TSESTree.FunctionDeclaration
    | TSESTree.FunctionExpression,
): void {
  // Only handles the three function types
}
```

**Go Implementation:**
```go
case ast.KindArrowFunction:        inferrableParameterVisitor,
case ast.KindFunctionDeclaration:  inferrableParameterVisitor,
case ast.KindFunctionExpression:   inferrableParameterVisitor,
case ast.KindConstructor:          inferrableParameterVisitor,
case ast.KindMethodDeclaration:    inferrableParameterVisitor,
```

**Issue:** Go implementation handles more node types (Constructor, MethodDeclaration) than TypeScript version.

**Impact:** May check parameters in constructors and methods that TypeScript version ignores.

**Test Coverage:** Constructor parameter test case shows this should work, so Go implementation may be more complete.

#### 5. Accessor Property Detection
**TypeScript Implementation:**
```typescript
return {
  AccessorProperty: inferrablePropertyVisitor,
  // ...
};
```

**Go Implementation:**
```go
case ast.KindPropertySignature:
  // For accessor properties, check if it has the accessor keyword
  propSig := node.AsPropertySignatureDeclaration()
  if propSig.Modifiers() != nil {
    for _, mod := range propSig.Modifiers().Nodes {
      if mod.Kind == ast.KindAccessorKeyword {
        // This is an accessor property
```

**Issue:** Go implementation tries to detect accessor properties within PropertySignature handling, but TypeScript has a separate AccessorProperty node type. The detection logic may be incomplete.

**Impact:** May not properly handle accessor properties like `accessor a: number = 5;`.

**Test Coverage:** Test case with accessor properties needs verification.

#### 6. Null Value Detection
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TSNullKeyword:
  return init.type === AST_NODE_TYPES.Literal && init.value == null;
```

**Go Implementation:**
```go
case ast.KindNullKeyword:
  return isLiteral(init, "null")

// In isLiteral function:
case "null":
  return init.Kind == ast.KindNullKeyword
```

**Issue:** TypeScript checks for `Literal` node with `null` value, while Go checks for `NullKeyword` directly. Different AST representations may cause mismatches.

**Impact:** Potential mismatch in null literal detection.

**Test Coverage:** Test case `'const a: null = null;'` needs verification.

### Recommendations
- Fix void expression detection to use hasUnaryPrefix logic consistently
- Verify template literal AST node mappings between TypeScript and Go
- Remove or verify literal type handling if it's not in the original TypeScript rule
- Ensure null literal detection matches TypeScript behavior
- Verify accessor property detection works correctly with Go AST
- Add comprehensive test coverage for edge cases like optional chaining
- Consider adding debug logging to compare behavior between implementations

---