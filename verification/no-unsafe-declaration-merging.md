## Rule: no-unsafe-declaration-merging

### Test File: no-unsafe-declaration-merging.test.ts

### Validation Summary
- ✅ **CORRECT**: Rule structure, AST node listeners for classes and interfaces, error message format
- ⚠️ **POTENTIAL ISSUES**: Scope handling differences, symbol resolution approach
- ❌ **INCORRECT**: Core symbol checking logic is not implemented (always returns false)

### Discrepancies Found

#### 1. Missing Core Implementation
**TypeScript Implementation:**
```typescript
function checkUnsafeDeclaration(
  scope: Scope,
  node: TSESTree.Identifier,
  unsafeKind: AST_NODE_TYPES,
): void {
  const variable = scope.set.get(node.name);
  if (!variable) {
    return;
  }

  const defs = variable.defs;
  if (defs.length <= 1) {
    return;
  }

  if (defs.some(def => def.node.type === unsafeKind)) {
    context.report({
      node,
      messageId: 'unsafeMerging',
    });
  }
}
```

**Go Implementation:**
```go
hasDeclarationOfKind := func(symbol any, kind ast.Kind) bool {
  if symbol == nil {
    return false
  }
  // Note: This is a simplified check - in a real implementation,
  // we would need to access the symbol's declarations
  return false
}
```

**Issue:** The Go implementation contains a placeholder that always returns false, making the rule non-functional.

**Impact:** Rule will never trigger on any unsafe declaration merging cases, causing all invalid test cases to fail.

**Test Coverage:** All invalid test cases will fail because the rule never reports violations.

#### 2. Scope vs Symbol-based Approach
**TypeScript Implementation:**
```typescript
// by default eslint returns the inner class scope for the ClassDeclaration node
// but we want the outer scope within which merged variables will sit
const currentScope = context.sourceCode.getScope(node).upper;
if (currentScope == null) {
  return;
}

checkUnsafeDeclaration(
  currentScope,
  node.id,
  AST_NODE_TYPES.TSInterfaceDeclaration,
);
```

**Go Implementation:**
```go
// Get the symbol for this class name
symbol := ctx.TypeChecker.GetSymbolAtLocation(className)
if symbol == nil {
  return
}
```

**Issue:** The TypeScript version uses scope-based variable resolution while the Go version attempts to use TypeScript symbols. The scope management is crucial for handling nested scopes correctly.

**Impact:** May miss cases where declarations exist in different scopes or handle scope boundaries incorrectly.

**Test Coverage:** Test cases with nested scopes or global declarations may behave differently.

#### 3. Symbol Declaration Analysis Missing
**TypeScript Implementation:**
```typescript
const defs = variable.defs;
if (defs.length <= 1) {
  return;
}

if (defs.some(def => def.node.type === unsafeKind)) {
  context.report({
    node,
    messageId: 'unsafeMerging',
  });
}
```

**Go Implementation:**
```go
// Check if this symbol also has interface declarations
if hasDeclarationOfKind(symbol, ast.KindInterfaceDeclaration) {
  reportUnsafeMerging(className)
}
```

**Issue:** The Go version needs to implement actual symbol declaration analysis. It should check if a symbol has multiple declarations and if any are of the conflicting kind.

**Impact:** Without proper declaration analysis, the rule cannot detect when classes and interfaces with the same name exist in the same scope.

**Test Coverage:** All invalid test cases depend on this functionality.

#### 4. Missing Scope Boundary Handling
**TypeScript Implementation:**
```typescript
// Uses scope.upper to get the containing scope for class declarations
const currentScope = context.sourceCode.getScope(node).upper;
```

**Go Implementation:**
```go
// No equivalent scope boundary handling
```

**Issue:** The TypeScript version carefully handles scope boundaries, especially for class declarations where it needs to check the outer scope. The Go version doesn't implement this logic.

**Impact:** May incorrectly handle cases where classes and interfaces are declared in different nested scopes.

**Test Coverage:** Cases with function-scoped classes or global declarations may behave incorrectly.

### Recommendations
- **Implement hasDeclarationOfKind function**: Replace the placeholder with actual logic to check symbol declarations using the TypeScript checker API
- **Add proper symbol declaration iteration**: Access symbol.valueDeclaration and symbol.declarations to check for multiple declarations of different kinds
- **Consider scope handling**: Evaluate if the symbol-based approach adequately handles scope boundaries or if additional scope checks are needed
- **Test the implementation**: Run the test suite to verify that all invalid cases are properly detected
- **Add debug logging**: Temporarily add logging to understand how symbols and declarations are accessed in the Go TypeScript bindings

---