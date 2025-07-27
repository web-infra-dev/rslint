# Rule Validation: no-invalid-void-type

## Rule: no-invalid-void-type

### Test File: no-invalid-void-type.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic rule structure, options parsing, message ID definitions, generic type argument checking with allowlist
- ⚠️ **POTENTIAL ISSUES**: Simplified union type validation, missing overload signature detection, incomplete this parameter validation, AST node type mapping differences
- ❌ **INCORRECT**: Missing TSTypeAnnotation validation, incomplete type parameter default checking, missing HasOverloadSignatures utility, incomplete AST traversal patterns

### Discrepancies Found

#### 1. Missing TSTypeAnnotation Parent Validation
**TypeScript Implementation:**
```typescript
const validParents: AST_NODE_TYPES[] = [
  AST_NODE_TYPES.TSTypeAnnotation, //
];
```

**Go Implementation:**
```go
// Valid parent node types for void
validParents := []ast.Kind{}
```

**Issue:** The Go implementation doesn't include TSTypeAnnotation as a valid parent, which is essential for allowing void in return types.

**Impact:** Valid void usage in return types may be incorrectly flagged as errors.

**Test Coverage:** Tests like `'function func(): void {}'` would fail without this.

#### 2. Incomplete AST Node Type Mapping
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

**Issue:** AccessorProperty is mapped to PropertyDeclaration (duplicated), which may not be semantically equivalent.

**Impact:** Accessor properties may not be handled correctly.

**Test Coverage:** Tests with accessor properties may pass when they should fail.

#### 3. Missing Union Type Validation Logic
**TypeScript Implementation:**
```typescript
function isValidUnionType(node: TSESTree.TSUnionType): boolean {
  return node.types.every(
    member =>
      validUnionMembers.includes(member.type) ||
      (member.type === AST_NODE_TYPES.TSTypeReference &&
        member.typeArguments?.type === AST_NODE_TYPES.TSTypeParameterInstantiation &&
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

**Issue:** Union type validation is completely missing, always returning true.

**Impact:** Invalid void usage in unions will not be detected.

**Test Coverage:** Tests like `'type UnionType2 = string | number | void;'` would incorrectly pass.

#### 4. Missing This Parameter Validation
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

**Issue:** This parameter validation is completely skipped.

**Impact:** Valid void usage in this parameters may be flagged as errors when allowAsThisParameter is true.

**Test Coverage:** Tests like `'function f(this: void) {}'` would fail.

#### 5. Incomplete Type Parameter Default Checking
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

**Issue:** Type parameter default validation is not implemented.

**Impact:** Invalid void usage in type parameter defaults may not be caught.

**Test Coverage:** Tests with `<T extends void = void>` patterns may incorrectly pass.

#### 6. Missing Overload Signature Detection
**TypeScript Implementation:**
```typescript
// using `void` as part of the return type of function overloading implementation
if (node.parent.type === AST_NODE_TYPES.TSUnionType) {
  const declaringFunction = getParentFunctionDeclarationNode(node.parent);
  
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

**Issue:** Overload signature detection is completely missing.

**Impact:** Valid void usage in function overloads may be incorrectly flagged.

**Test Coverage:** Multiple overload test cases would fail without this logic.

#### 7. Wrong AST Node Kind for Generic Type Arguments
**TypeScript Implementation:**
```typescript
if (allowInGenericTypeArguments === true) {
  validParents.push(AST_NODE_TYPES.TSTypeParameterInstantiation);
}

// Check if parent is TSTypeParameterInstantiation
if (
  node.parent.type === AST_NODE_TYPES.TSTypeParameterInstantiation &&
  node.parent.parent.type === AST_NODE_TYPES.TSTypeReference
) {
  checkGenericTypeArgument(node);
  return;
}
```

**Go Implementation:**
```go
// If allowInGenericTypeArguments is true, add to valid parents
if allowGeneric, ok := opts.AllowInGenericTypeArguments.(bool); ok && allowGeneric {
  validParents = append(validParents, ast.KindExpressionWithTypeArguments)
}

// Check T<..., void, ...> against allowInGenericArguments option
if node.Parent != nil &&
  node.Parent.Kind == ast.KindExpressionWithTypeArguments &&
  node.Parent.Parent != nil &&
  node.Parent.Parent.Kind == ast.KindTypeReference {
  checkGenericTypeArgument(node)
  return
}
```

**Issue:** Uses `ast.KindExpressionWithTypeArguments` instead of the equivalent of `TSTypeParameterInstantiation`.

**Impact:** May not correctly identify generic type argument contexts.

**Test Coverage:** Generic type argument tests may not behave as expected.

#### 8. Simplified Message ID Logic
**TypeScript Implementation:**
```typescript
context.report({
  node,
  messageId:
    allowInGenericTypeArguments && allowAsThisParameter
      ? 'invalidVoidNotReturnOrThisParamOrGeneric'
      : allowInGenericTypeArguments
        ? getNotReturnOrGenericMessageId(node)
        : allowAsThisParameter
          ? 'invalidVoidNotReturnOrThisParam'
          : 'invalidVoidNotReturn',
});
```

**Go Implementation:**
```go
// Determine message ID based on options
messageId := "invalidVoidNotReturn"

allowInGeneric := false
if allowGenericBool, ok := opts.AllowInGenericTypeArguments.(bool); ok {
  allowInGeneric = allowGenericBool
} else if _, ok := opts.AllowInGenericTypeArguments.([]interface{}); ok {
  allowInGeneric = true
}

if allowInGeneric && opts.AllowAsThisParameter {
  messageId = "invalidVoidNotReturnOrThisParamOrGeneric"
} else if allowInGeneric {
  messageId = getNotReturnOrGenericMessageId(node, opts)
} else if opts.AllowAsThisParameter {
  messageId = "invalidVoidNotReturnOrThisParam"
}
```

**Issue:** The message ID selection logic is less nuanced and may not handle all cases correctly.

**Impact:** Error messages may not be as specific as in the TypeScript version.

**Test Coverage:** Message ID assertions in tests may fail.

### Recommendations

1. **Add TSTypeAnnotation Support**: Include the equivalent of TSTypeAnnotation in validParents to allow void in return types
2. **Implement Union Type Validation**: Add proper union type checking to detect invalid void constituents
3. **Add This Parameter Detection**: Implement proper this parameter validation when allowAsThisParameter is true
4. **Fix AST Node Mapping**: Ensure correct mapping between TypeScript AST node types and Go AST kinds
5. **Implement Type Parameter Checking**: Add support for checking type parameter defaults and constraints
6. **Add Overload Signature Detection**: Implement the HasOverloadSignatures utility and related logic
7. **Enhance Message ID Logic**: Make message ID selection more accurate to match TypeScript behavior
8. **Complete TODO Items**: Address all the TODO comments in the Go implementation

### Critical Missing Functionality

The Go implementation is missing several core features that are essential for the rule to work correctly:
- Union type validation (completely skipped)
- This parameter validation (not implemented) 
- Overload signature detection (missing utility)
- Type parameter constraint/default checking (incomplete)
- Proper AST node parent validation (missing TSTypeAnnotation equivalent)

These missing features would cause many test cases to fail and the rule to behave incorrectly compared to the TypeScript-ESLint version.

---