## Rule: no-invalid-void-type

### Test File: no-invalid-void-type.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic rule structure and option parsing
  - Generic type argument checking with allowlist support
  - Message ID generation and error reporting
  - Core void keyword detection pattern

- ⚠️ **POTENTIAL ISSUES**: 
  - Union type validation logic is simplified/incomplete
  - Type parameter default checking is commented out
  - Function overload signature detection is missing
  - `this` parameter void checking is not implemented
  - AST node kind mappings may be incorrect

- ❌ **INCORRECT**: 
  - Missing comprehensive union type validation
  - Incomplete implementation of several key features
  - AST traversal patterns don't match TypeScript version exactly

### Discrepancies Found

#### 1. Union Type Validation Logic Missing
**TypeScript Implementation:**
```typescript
function isValidUnionType(node: TSESTree.TSUnionType): boolean {
  return node.types.every(
    member =>
      validUnionMembers.includes(member.type) ||
      // allows any T<..., void, ...> here, checked by checkGenericTypeArgument
      (member.type === AST_NODE_TYPES.TSTypeReference &&
        member.typeArguments?.type ===
          AST_NODE_TYPES.TSTypeParameterInstantiation &&
        member.typeArguments.params
          .map(param => param.type)
          .includes(AST_NODE_TYPES.TSVoidKeyword)),
  );
}
```

**Go Implementation:**
```go
// Check if a union containing void is valid
// TODO: Reimplement when UnionType API is clarified
_ = func(node *ast.UnionType) bool {
  // Simplified check - just return true for now
  // The Types() method is not available in current API
  return true
}
```

**Issue:** The Go implementation completely skips union type validation, always returning true

**Impact:** Critical test cases will fail, including:
- `type UnionType2 = string | number | void;` (should error with `invalidVoidUnionConstituent`)
- Union type validation in function return types
- Overload signature union type checking

**Test Coverage:** Multiple test cases in the invalid arrays test union types with void

#### 2. Type Parameter Default Checking Not Implemented
**TypeScript Implementation:**
```typescript
function checkDefaultVoid(
  node: TSESTree.TSVoidKeyword,
  parentNode: TSESTree.TSTypeParameter,
): void {
  if (parentNode.default !== node) {
    context.report({
      node,
      messageId: getNotReturnOrGenericMessageId(node),
    });
  }
}
```

**Go Implementation:**
```go
// Check if generic type parameter defaults to void
// TODO: Reimplement when TypeParameterDeclaration API is clarified
_ = func(node *ast.Node, parentNode *ast.TypeParameterDeclaration) {
  // Skip check for now
}
```

**Issue:** Type parameter default checking is completely disabled

**Impact:** Test cases like `<T extends void = void>` will not be properly validated

**Test Coverage:** Tests with generic type parameter defaults will fail

#### 3. Function Overload Signature Detection Missing
**TypeScript Implementation:**
```typescript
if (node.parent.type === AST_NODE_TYPES.TSUnionType) {
  const declaringFunction = getParentFunctionDeclarationNode(
    node.parent,
  );

  if (
    declaringFunction &&
    hasOverloadSignatures(declaringFunction, context)
  ) {
    return;
  }
}
```

**Go Implementation:**
```go
// Using void as part of function overloading implementation
// Skip overload check for now as HasOverloadSignatures is not available
// TODO: Reimplement overload signature detection
```

**Issue:** The Go version doesn't check for function overload signatures, which is critical for allowing void in union return types for overloaded functions

**Impact:** Valid overload signatures will be incorrectly flagged as errors

**Test Coverage:** Multiple test cases with function overloads will fail

#### 4. This Parameter Void Checking Not Implemented
**TypeScript Implementation:**
```typescript
// this parameter is ok to be void.
if (
  allowAsThisParameter &&
  node.parent.type === AST_NODE_TYPES.TSTypeAnnotation &&
  node.parent.parent.type === AST_NODE_TYPES.Identifier &&
  node.parent.parent.name === 'this'
) {
  return;
}
```

**Go Implementation:**
```go
// This parameter is ok to be void
// Skip this parameter check for now - needs proper type node detection
// TODO: Reimplement this parameter void check
```

**Issue:** The `allowAsThisParameter` option is not properly implemented

**Impact:** Valid `this` parameter void types will be incorrectly flagged when the option is enabled

**Test Coverage:** Tests with `allowAsThisParameter: true` will fail

#### 5. AST Node Kind Mappings Issues
**TypeScript Implementation:**
```typescript
const invalidGrandParents: AST_NODE_TYPES[] = [
  AST_NODE_TYPES.TSPropertySignature,
  AST_NODE_TYPES.CallExpression,
  AST_NODE_TYPES.PropertyDefinition,
  AST_NODE_TYPES.AccessorProperty,
  AST_NODE_TYPES.Identifier,
];
```

**Go Implementation:**
```go
invalidGrandParents := []ast.Kind{
  ast.KindPropertySignature,
  ast.KindCallExpression,
  ast.KindPropertyDeclaration,
  ast.KindPropertyDeclaration, // No accessor property declaration kind, using PropertyDeclaration
  ast.KindIdentifier,
}
```

**Issue:** AccessorProperty is mapped to PropertyDeclaration, and there's a duplicate entry

**Impact:** May not correctly handle accessor properties differently from regular properties

**Test Coverage:** Tests with accessor properties may not behave correctly

#### 6. Valid Parents Logic Incomplete
**TypeScript Implementation:**
```typescript
const validParents: AST_NODE_TYPES[] = [
  AST_NODE_TYPES.TSTypeAnnotation, //
];
// ...
if (allowInGenericTypeArguments === true) {
  validParents.push(AST_NODE_TYPES.TSTypeParameterInstantiation);
}
```

**Go Implementation:**
```go
validParents := []ast.Kind{}

// If allowInGenericTypeArguments is true, add to valid parents
if allowGeneric, ok := opts.AllowInGenericTypeArguments.(bool); ok && allowGeneric {
  validParents = append(validParents, ast.KindExpressionWithTypeArguments)
}
```

**Issue:** The Go version doesn't include `TSTypeAnnotation` by default and uses `ExpressionWithTypeArguments` instead of `TSTypeParameterInstantiation`

**Impact:** Basic type annotations may not be handled correctly

**Test Coverage:** Basic function parameter type annotations will likely fail

#### 7. Generic Type Reference Name Extraction
**TypeScript Implementation:**
```typescript
const fullyQualifiedName = context.sourceCode
  .getText(node.parent.parent.typeName)
  .replaceAll(' ', '');
```

**Go Implementation:**
```go
func getTypeReferenceName(ctx rule.RuleContext, typeRef *ast.TypeReferenceNode) string {
  textRange := utils.TrimNodeTextRange(ctx.SourceFile, typeRef.TypeName)
  return string(ctx.SourceFile.Text()[textRange.Pos():textRange.End()])
}
```

**Issue:** The Go version doesn't remove spaces from the type name, which is crucial for matching against the allowlist

**Impact:** Allowlist matching for spaced type names like `Ex . Mx . Tx` will fail

**Test Coverage:** Whitelist tests with spaced type names will fail

### Recommendations
- **CRITICAL**: Implement proper union type validation using the AST API
- **CRITICAL**: Add function overload signature detection functionality
- **HIGH**: Implement `this` parameter void checking when `allowAsThisParameter` is enabled
- **HIGH**: Fix type parameter default checking
- **MEDIUM**: Correct AST node kind mappings for accessor properties
- **MEDIUM**: Add proper space removal in type name extraction for allowlist matching
- **MEDIUM**: Include `TSTypeAnnotation` equivalent in valid parents by default
- **LOW**: Remove duplicate entries in invalidGrandParents array
- **TEST**: Add comprehensive test coverage for all the missing functionality
- **TEST**: Verify AST node traversal patterns match TypeScript behavior exactly

---