## Rule: no-misused-new

### Test File: no-misused-new.test.ts

### Validation Summary
- ✅ **CORRECT**: Error message definitions, basic structure, and type reference name extraction
- ⚠️ **POTENTIAL ISSUES**: AST node type handling, parent traversal logic, method body detection
- ❌ **INCORRECT**: CSS selector pattern matching not properly translated, missing TSEmptyBodyFunctionExpression detection, incomplete interface body validation

### Discrepancies Found

#### 1. CSS Selector Pattern Not Properly Translated
**TypeScript Implementation:**
```typescript
"ClassBody > MethodDefinition[key.name='new']"(node: TSESTree.MethodDefinition): void {
  if (
    node.value.type === AST_NODE_TYPES.TSEmptyBodyFunctionExpression &&
    isMatchingParentType(node.parent.parent, node.value.returnType)
  ) {
    context.report({
      node,
      messageId: 'errorMessageClass',
    });
  }
}
```

**Go Implementation:**
```go
ast.KindMethodDeclaration: func(node *ast.Node) {
  methodDecl := node.AsMethodDeclaration()
  
  // Check if the method name is 'new'
  methodName, _ := utils.GetNameFromMember(ctx.SourceFile, &methodDecl.Node)
  if methodName != "new" {
    return
  }

  // Check if it's in a class body
  if node.Parent == nil || node.Parent.Kind != ast.KindBlock {
    return
  }
  // ... rest of logic
}
```

**Issue:** The TypeScript version uses a CSS selector that specifically targets `MethodDefinition` nodes inside `ClassBody` with `key.name='new'`. The Go version listens to all `MethodDeclaration` nodes and then filters, which may catch different AST structures.

**Impact:** May miss or incorrectly flag methods depending on AST structure differences between TypeScript-ESLint and typescript-go.

**Test Coverage:** This affects test cases with class methods named 'new'.

#### 2. Missing TSEmptyBodyFunctionExpression Detection
**TypeScript Implementation:**
```typescript
if (
  node.value.type === AST_NODE_TYPES.TSEmptyBodyFunctionExpression &&
  isMatchingParentType(node.parent.parent, node.value.returnType)
)
```

**Go Implementation:**
```go
// Check if the method has an empty body (TSEmptyBodyFunctionExpression)
if methodDecl.Body != nil {
  // Method has a body, so it's OK
  return
}
```

**Issue:** The TypeScript version specifically checks for `TSEmptyBodyFunctionExpression` type, while the Go version only checks if `Body` is nil. This is a critical difference as it affects when the rule triggers.

**Impact:** The Go version may not properly distinguish between different types of empty methods (abstract methods, method signatures, etc.).

**Test Coverage:** This affects the valid test case `class C { new() {} }` which should be OK because it has a body.

#### 3. Interface Constructor Signature Pattern Mismatch
**TypeScript Implementation:**
```typescript
'TSInterfaceBody > TSConstructSignatureDeclaration'(node: TSESTree.TSConstructSignatureDeclaration): void {
  if (
    isMatchingParentType(
      node.parent.parent as TSESTree.TSInterfaceDeclaration,
      node.returnType,
    )
  ) {
    context.report({
      node,
      messageId: 'errorMessageInterface',
    });
  }
}
```

**Go Implementation:**
```go
ast.KindConstructSignature: func(node *ast.Node) {
  constructSig := node.AsConstructSignatureDeclaration()

  // Check if it's in an interface body
  if node.Parent == nil || node.Parent.Kind != ast.KindBlock {
    return
  }

  interfaceBody := node.Parent
  if interfaceBody.Parent == nil || interfaceBody.Parent.Kind != ast.KindInterfaceDeclaration {
    return
  }
  // ... rest of logic
}
```

**Issue:** The TypeScript version uses specific CSS selector targeting, while Go version manually traverses parent hierarchy. The AST structure assumptions may not match.

**Impact:** May miss construct signatures in interfaces or flag incorrect nodes.

**Test Coverage:** This affects interface test cases with `new (): I` patterns.

#### 4. Method Signature Constructor Detection Scope
**TypeScript Implementation:**
```typescript
"TSMethodSignature[key.name='constructor']"(node: TSESTree.TSMethodSignature): void {
  context.report({
    node,
    messageId: 'errorMessageInterface',
  });
}
```

**Go Implementation:**
```go
ast.KindMethodSignature: func(node *ast.Node) {
  methodSig := node.AsMethodSignatureDeclaration()

  // Check if the method name is 'constructor'
  methodName, _ := utils.GetNameFromMember(ctx.SourceFile, &methodSig.Node)
  if methodName != "constructor" {
    return
  }

  // Report error for any method signature named 'constructor' in interfaces
  ctx.ReportNode(node, buildErrorMessageInterfaceMessage())
}
```

**Issue:** The TypeScript version reports ALL method signatures named 'constructor' without context checking, while the Go implementation has the same logic but may have different AST node coverage.

**Impact:** Should work similarly, but AST node type differences might affect coverage.

**Test Coverage:** This affects the test case with `constructor(): void;` in type literals and interfaces.

#### 5. Type Reference Name Extraction Logic Gap
**TypeScript Implementation:**
```typescript
function getTypeReferenceName(node): string | null {
  if (node) {
    switch (node.type) {
      case AST_NODE_TYPES.TSTypeAnnotation:
        return getTypeReferenceName(node.typeAnnotation);
      case AST_NODE_TYPES.TSTypeReference:
        return getTypeReferenceName(node.typeName);
      case AST_NODE_TYPES.Identifier:
        return node.name;
      default:
        break;
    }
  }
  return null;
}
```

**Go Implementation:**
```go
func getTypeReferenceName(node *ast.Node) string {
  if node == nil {
    return ""
  }

  switch node.Kind {
  case ast.KindTypeReference:
    typeRef := node.AsTypeReferenceNode()
    return getTypeReferenceName(typeRef.TypeName)
  case ast.KindIdentifier:
    return node.AsIdentifier().Text
  default:
    return ""
  }
}
```

**Issue:** The Go version is missing the `TSTypeAnnotation` case, which is crucial for extracting type names from annotated return types.

**Impact:** May fail to properly match return types in methods with type annotations, leading to missed violations.

**Test Coverage:** This affects test cases where methods have explicit return type annotations like `new(): C`.

### Recommendations
- Add handling for `TSTypeAnnotation` equivalent in Go's `getTypeReferenceName` function
- Investigate the correct AST node types in typescript-go that correspond to TypeScript-ESLint's `TSEmptyBodyFunctionExpression`
- Verify the AST structure for method declarations in classes vs. method signatures in interfaces
- Add more specific parent traversal validation to ensure correct context detection
- Test the rule with actual TypeScript code to verify AST node type mappings
- Consider adding debug output to compare AST structures between the two implementations

---