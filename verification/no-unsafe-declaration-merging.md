## Rule: no-unsafe-declaration-merging

### Test File: no-unsafe-declaration-merging.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic rule structure, listener setup for ClassDeclaration and InterfaceDeclaration, error message ID and description
- ⚠️ **POTENTIAL ISSUES**: Incomplete symbol checking implementation, missing scope analysis
- ❌ **INCORRECT**: Core detection logic not implemented, missing TypeScript symbol API usage

### Discrepancies Found

#### 1. Incomplete Symbol Declaration Checking
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

**Issue:** The Go implementation has a placeholder function that always returns false, completely disabling the rule's core functionality.

**Impact:** The rule will never detect unsafe declaration merging, causing all test cases to fail.

**Test Coverage:** All invalid test cases will fail since no errors will be reported.

#### 2. Missing Scope Analysis for Class Declarations
**TypeScript Implementation:**
```typescript
ClassDeclaration(node): void {
  if (node.id) {
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
  }
}
```

**Go Implementation:**
```go
ast.KindClassDeclaration: func(node *ast.Node) {
  classDecl := node.AsClassDeclaration()
  className := classDecl.Name()
  if className == nil {
    return
  }

  // Get the symbol for this class name
  symbol := ctx.TypeChecker.GetSymbolAtLocation(className)
  // ... rest of logic
}
```

**Issue:** The Go implementation doesn't handle scope correctly. The TypeScript version explicitly uses the outer scope for class declarations, while the Go version doesn't consider scope at all.

**Impact:** May miss cases where class and interface declarations exist in nested scopes or incorrectly flag valid cases.

**Test Coverage:** Test cases with nested scopes or global declarations may behave differently.

#### 3. Missing Symbol Declaration Analysis
**TypeScript Implementation:**
```typescript
const variable = scope.set.get(node.name);
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
symbol := ctx.TypeChecker.GetSymbolAtLocation(className)
if symbol == nil {
  return
}

// Check if this symbol also has interface declarations
if hasDeclarationOfKind(symbol, ast.KindInterfaceDeclaration) {
  reportUnsafeMerging(className)
}
```

**Issue:** The Go implementation needs to access the symbol's declarations to check if it has multiple declaration types. The TypeScript API provides `variable.defs` while Go needs to use TypeScript's symbol API properly.

**Impact:** Cannot detect when the same identifier is used for both class and interface declarations.

**Test Coverage:** All invalid test cases require this functionality to work.

#### 4. Missing Implementation of Symbol Declaration Inspection
**Issue:** The Go version needs to implement proper access to TypeScript symbol declarations. The TypeScript compiler API in Go should provide access to symbol declarations similar to the TypeScript version's `variable.defs`.

**Expected Go Implementation Pattern:**
```go
// Pseudo-code for what should be implemented
symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
if symbol != nil {
  declarations := symbol.GetDeclarations() // Need to access declarations
  if len(declarations) > 1 {
    // Check if declarations include both class and interface types
    hasClass := false
    hasInterface := false
    for _, decl := range declarations {
      if decl.Kind() == ast.KindClassDeclaration {
        hasClass = true
      } else if decl.Kind() == ast.KindInterfaceDeclaration {
        hasInterface = true
      }
    }
    if hasClass && hasInterface {
      reportUnsafeMerging(node)
    }
  }
}
```

### Recommendations
- Implement proper symbol declaration analysis using TypeScript's symbol API in Go
- Add correct scope handling for class declarations (use outer scope)
- Replace the placeholder `hasDeclarationOfKind` function with actual symbol inspection
- Verify that the TypeScript symbol API bindings in typescript-go support accessing symbol declarations
- Add comprehensive tests to ensure scope-based detection works correctly
- Consider edge cases like global declarations and nested scopes

### Critical Implementation Gap
The current Go implementation is essentially non-functional due to the placeholder logic. The core detection mechanism must be implemented using the TypeScript compiler's symbol table to properly identify declaration merging scenarios.

---