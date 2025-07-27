## Rule: no-unsafe-function-type

### Test File: no-unsafe-function-type.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core rule logic correctly identifies `Function` type references
  - Error message content matches TypeScript implementation
  - Handles type references (`let value: Function`)
  - Handles class implements clauses (`class Foo implements Function`)
  - Handles interface extends clauses (`interface Foo extends Function`)
  - Correctly uses `isReferenceToGlobalFunction` utility for type checking
  - Proper scope handling to distinguish global Function from local Function types

- ⚠️ **POTENTIAL ISSUES**: 
  - AST node traversal patterns differ but appear functionally equivalent
  - Heritage clause handling is more complex in Go but covers same cases

- ❌ **INCORRECT**: No definitive discrepancies found

### Discrepancies Found

#### 1. AST Node Pattern Differences
**TypeScript Implementation:**
```typescript
return {
  TSClassImplements(node): void {
    checkBannedTypes(node.expression);
  },
  TSInterfaceHeritage(node): void {
    checkBannedTypes(node.expression);
  },
  TSTypeReference(node): void {
    checkBannedTypes(node.typeName);
  },
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
  ast.KindTypeReference: func(node *ast.Node) {
    typeRef := node.AsTypeReference()
    checkBannedTypes(typeRef.TypeName)
  },
  ast.KindHeritageClause: func(node *ast.Node) {
    // Complex logic to handle both implements and extends
  },
}
```

**Issue:** The Go implementation uses a single `KindHeritageClause` listener instead of separate `TSClassImplements` and `TSInterfaceHeritage` listeners, but this is acceptable as it achieves the same result through conditional logic.

**Impact:** No functional impact - both approaches correctly identify the target AST patterns.

**Test Coverage:** All test cases should pass with both implementations.

#### 2. Heritage Clause Processing Complexity
**TypeScript Implementation:**
```typescript
TSClassImplements(node): void {
  checkBannedTypes(node.expression);
},
TSInterfaceHeritage(node): void {
  checkBannedTypes(node.expression);
},
```

**Go Implementation:**
```go
ast.KindHeritageClause: func(node *ast.Node) {
  heritageClause := node.AsHeritageClause()
  
  // Only check implements and extends clauses
  if heritageClause.Token != ast.KindImplementsKeyword && heritageClause.Token != ast.KindExtendsKeyword {
    return
  }

  // Check if this is a class implements or interface extends
  parent := node.Parent
  if parent == nil {
    return
  }

  isClassImplements := ast.IsClassDeclaration(parent) && heritageClause.Token == ast.KindImplementsKeyword
  isInterfaceExtends := ast.IsInterfaceDeclaration(parent) && heritageClause.Token == ast.KindExtendsKeyword

  if !isClassImplements && !isInterfaceExtends {
    return
  }

  // Check each type in the heritage clause
  for _, heritageType := range heritageClause.Types.Nodes {
    if heritageType.AsExpressionWithTypeArguments().Expression != nil {
      checkBannedTypes(heritageType.AsExpressionWithTypeArguments().Expression)
    }
  }
},
```

**Issue:** The Go implementation is more verbose but functionally equivalent. It manually filters for the correct heritage clause types and parent contexts.

**Impact:** No functional impact - both approaches achieve the same filtering result.

**Test Coverage:** The test cases `class Weird implements Function` and `interface Weird extends Function` validate this behavior.

#### 3. Global Function Type Detection
**TypeScript Implementation:**
```typescript
node.type === AST_NODE_TYPES.Identifier &&
node.name === 'Function' &&
isReferenceToGlobalFunction('Function', node, context.sourceCode)
```

**Go Implementation:**
```go
func isReferenceToGlobalFunction(ctx rule.RuleContext, node *ast.Node) bool {
  if !ast.IsIdentifier(node) || node.AsIdentifier().Text != "Function" {
    return false
  }

  // Get the type at this location to check if it's the global Function
  nodeType := ctx.TypeChecker.GetTypeAtLocation(node)
  if nodeType == nil {
    return false
  }

  // Multiple approaches to detect built-in Function type
  if utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, nodeType, "Function") {
    return true
  }
  
  if utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, nodeType, "FunctionConstructor") {
    return true
  }
  
  // Check symbol declarations to distinguish global vs local Function
  symbol := checker.Type_symbol(nodeType)
  if symbol != nil && symbol.Name == "Function" {
    hasDefaultLibDeclaration := false
    hasUserCodeDeclaration := false
    
    for _, declaration := range symbol.Declarations {
      sourceFile := ast.GetSourceFileOfNode(declaration)
      if sourceFile != nil {
        if utils.IsSourceFileDefaultLibrary(ctx.Program, sourceFile) {
          hasDefaultLibDeclaration = true
        } else {
          hasUserCodeDeclaration = true
        }
      }
    }
    
    if hasUserCodeDeclaration {
      return false
    }
    
    if hasDefaultLibDeclaration {
      return true
    }
  }
  
  return false
}
```

**Issue:** The Go implementation is more comprehensive in detecting the global Function type, using multiple strategies including built-in symbol detection and source file analysis.

**Impact:** Positive impact - the Go implementation may be more robust at distinguishing global Function from local Function types.

**Test Coverage:** The valid test case with local `type Function = () => void` specifically tests this distinction.

### Recommendations
- ✅ **No fixes needed** - The Go implementation appears functionally correct and equivalent to the TypeScript version
- ✅ **Test coverage is adequate** - All original test cases are ported and should pass
- ✅ **Enhanced type detection** - The Go implementation may actually be more robust in distinguishing global vs local Function types

### Overall Assessment
The Go port of `no-unsafe-function-type` is **functionally correct** and properly implements the same rule logic as the TypeScript version. The implementation differences are due to language-specific AST handling patterns but achieve equivalent results. The Go version may even be more robust in some edge cases related to global type detection.

---