# Rule Validation: no-misused-new

## Rule: no-misused-new

### Test File: no-misused-new.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Error message definitions match exactly
  - Basic AST pattern matching for method declarations with name 'new'
  - Constructor signature detection in interfaces
  - Method signature detection with name 'constructor'
  - Type reference name extraction logic

- ⚠️ **POTENTIAL ISSUES**: 
  - Parent navigation logic differs significantly from TypeScript version
  - AST structure assumptions may not match typescript-go's actual structure
  - Missing handling of TSEmptyBodyFunctionExpression equivalent

- ❌ **INCORRECT**: 
  - Selector pattern matching doesn't align with TypeScript's CSS-like selectors
  - Interface body detection logic is incorrect
  - Class body detection logic is incorrect
  - Return type handling for generic types may be incomplete

### Discrepancies Found

#### 1. CSS Selector vs Manual AST Navigation
**TypeScript Implementation:**
```typescript
return {
  "ClassBody > MethodDefinition[key.name='new']"(node: TSESTree.MethodDefinition): void {
    // Automatically gets methods named 'new' in class bodies
  },
  'TSInterfaceBody > TSConstructSignatureDeclaration'(node: TSESTree.TSConstructSignatureDeclaration): void {
    // Automatically gets construct signatures in interface bodies
  },
  "TSMethodSignature[key.name='constructor']"(node: TSESTree.TSMethodSignature): void {
    // Automatically gets method signatures named 'constructor'
  }
}
```

**Go Implementation:**
```go
ast.KindMethodDeclaration: func(node *ast.Node) {
  // Manual checks for parent structure and method name
  if node.Parent == nil || node.Parent.Kind != ast.KindBlock {
    return
  }
  // ... more manual navigation
}
```

**Issue:** The Go version manually navigates the AST tree and makes assumptions about parent-child relationships that may not be correct for the typescript-go AST structure.

**Impact:** The rule may miss valid cases or trigger on invalid cases due to incorrect parent detection.

**Test Coverage:** Multiple test cases would be affected, particularly those involving class expressions and interface declarations.

#### 2. Empty Body Function Detection
**TypeScript Implementation:**
```typescript
if (
  node.value.type === AST_NODE_TYPES.TSEmptyBodyFunctionExpression &&
  isMatchingParentType(node.parent.parent, node.value.returnType)
) {
  // Only report if method has empty body AND return type matches class
}
```

**Go Implementation:**
```go
if methodDecl.Body != nil {
  // Method has a body, so it's OK
  return
}
// Always check return type if no body
```

**Issue:** The Go version doesn't explicitly check for the equivalent of `TSEmptyBodyFunctionExpression`. It only checks if `Body != nil`, which may not be the same condition.

**Impact:** May incorrectly flag methods that have bodies or miss methods that should be flagged.

**Test Coverage:** The valid test case `class C { new() {} }` expects no error because it has a body.

#### 3. Interface Body Detection
**TypeScript Implementation:**
```typescript
'TSInterfaceBody > TSConstructSignatureDeclaration'(node) {
  // CSS selector automatically ensures we're in an interface body
}
```

**Go Implementation:**
```go
if node.Parent == nil || node.Parent.Kind != ast.KindBlock {
  return
}
interfaceBody := node.Parent
if interfaceBody.Parent == nil || interfaceBody.Parent.Kind != ast.KindInterfaceDeclaration {
  return
}
```

**Issue:** The Go version assumes interface bodies are represented as `ast.KindBlock`, but typescript-go may use a different AST node kind for interface bodies.

**Impact:** Constructor signatures in interfaces may not be detected at all.

**Test Coverage:** All interface-related test cases would fail: `interface I { new (): I; constructor(): void; }`

#### 4. Generic Type Handling
**TypeScript Implementation:**
```typescript
function getTypeReferenceName(node): string | null {
  switch (node.type) {
    case AST_NODE_TYPES.TSTypeAnnotation:
      return getTypeReferenceName(node.typeAnnotation);
    case AST_NODE_TYPES.TSTypeReference:
      return getTypeReferenceName(node.typeName);
    case AST_NODE_TYPES.Identifier:
      return node.name;
  }
  return null;
}
```

**Go Implementation:**
```go
func getTypeReferenceName(node *ast.Node) string {
  switch node.Kind {
  case ast.KindTypeReference:
    typeRef := node.AsTypeReferenceNode()
    return getTypeReferenceName(typeRef.TypeName)
  case ast.KindIdentifier:
    return node.AsIdentifier().Text
  }
  return ""
}
```

**Issue:** The Go version doesn't handle `TSTypeAnnotation` equivalent, which may be needed for proper type extraction.

**Impact:** Generic types like `G<T>` in return types may not be properly matched to interface names.

**Test Coverage:** The test case `interface G { new <T>(): G<T>; }` expects an error.

#### 5. Type Literal Handling
**TypeScript Implementation:**
```typescript
// The CSS selectors naturally exclude type literals since they target specific parent types
```

**Go Implementation:**
```go
// Manual parent checking may incorrectly include type literals
```

**Issue:** The Go version's manual parent navigation might incorrectly flag constructor signatures in type literals, which should be allowed.

**Impact:** False positives on valid code like `type T = { constructor(): void; }`.

**Test Coverage:** The test case `type T = { constructor(): void; };` expects an error, but `type T = { new (): T };` should be valid.

### Recommendations
- **Fix AST Navigation**: Research the actual typescript-go AST structure for interfaces and classes to correct parent-child relationships
- **Implement Proper Body Detection**: Find the typescript-go equivalent of `TSEmptyBodyFunctionExpression` checking
- **Add TSTypeAnnotation Handling**: Extend `getTypeReferenceName` to handle type annotations
- **Verify Selector Equivalence**: Ensure the manual AST navigation truly matches the behavior of the TypeScript CSS selectors
- **Add Debug Logging**: Temporarily add logging to understand the actual AST structure typescript-go produces
- **Test with Real AST**: Run the Go implementation against the test cases and examine what AST nodes are actually produced

---