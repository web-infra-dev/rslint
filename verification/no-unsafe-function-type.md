## Rule: no-unsafe-function-type

### Test File: no-unsafe-function-type.test.ts

### Validation Summary
- ✅ **CORRECT**: Core message handling, global Function type detection logic, basic type reference checking
- ⚠️ **POTENTIAL ISSUES**: Heritage clause handling is more complex than TypeScript version, AST node kind mappings may not be 1:1
- ❌ **INCORRECT**: AST visitor pattern mismatch - Go listens to different node types than TypeScript

### Discrepancies Found

#### 1. AST Visitor Pattern Mismatch
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
  // Check type references like: let value: Function;
  ast.KindTypeReference: func(node *ast.Node) {
    typeRef := node.AsTypeReferenceNode()
    checkBannedTypes(typeRef.TypeName)
  },

  // Check class implements clauses like: class Foo implements Function {}
  ast.KindHeritageClause: func(node *ast.Node) {
    // Complex logic to filter heritage clauses...
  },
}
```

**Issue:** The TypeScript version has specific visitors for `TSClassImplements` and `TSInterfaceHeritage` nodes, while the Go version uses a single `KindHeritageClause` visitor with complex filtering logic. This creates a structural mismatch.

**Impact:** The Go implementation may miss some cases or incorrectly process heritage clauses that aren't related to class implements or interface extends.

**Test Coverage:** Test cases for `class Weird implements Function` and `interface Weird extends Function` may not work correctly.

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

**Issue:** The Go implementation has much more complex logic to determine if a heritage clause should be processed, while the TypeScript version relies on specific AST node types to filter appropriately.

**Impact:** The additional complexity increases the chance of bugs and makes the code harder to maintain. The filtering logic may not perfectly match TypeScript's behavior.

**Test Coverage:** Both class implements and interface extends test cases need careful verification.

#### 3. Global Function Detection Implementation
**TypeScript Implementation:**
```typescript
isReferenceToGlobalFunction('Function', node, context.sourceCode)
```

**Go Implementation:**
```go
func isReferenceToGlobalFunction(ctx rule.RuleContext, node *ast.Node) bool {
  if !ast.IsIdentifier(node) || node.AsIdentifier().Text != "Function" {
    return false
  }

  // Get the symbol for the identifier
  symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
  if symbol == nil {
    return false
  }

  // Check if this symbol is from the default library (lib.*.d.ts)
  for _, declaration := range symbol.Declarations {
    if declaration == nil {
      continue
    }
    
    sourceFile := ast.GetSourceFileOfNode(declaration)
    if sourceFile == nil {
      continue
    }
    
    // If any declaration is NOT from the default library, this is user-defined
    if !utils.IsSourceFileDefaultLibrary(ctx.Program, sourceFile) {
      return false
    }
  }
  
  // If we have declarations and they're all from the default library, this is the global Function
  return len(symbol.Declarations) > 0
}
```

**Issue:** While the Go implementation looks functionally correct, it doesn't use the same utility function as the TypeScript version. The logic for checking if a symbol is from the default library may have subtle differences.

**Impact:** This could potentially cause differences in edge cases, particularly around shadowed `Function` identifiers.

**Test Coverage:** The test case with the locally scoped `Function` type alias should verify this works correctly.

### Recommendations
- **Fix AST visitor pattern**: Research the correct Go AST node kinds that correspond to TypeScript's `TSClassImplements` and `TSInterfaceHeritage` to simplify the heritage clause handling
- **Simplify heritage clause logic**: If possible, find more direct AST node types to listen to rather than using complex filtering in `KindHeritageClause`
- **Verify global Function detection**: Ensure the `IsSourceFileDefaultLibrary` utility correctly identifies default library files in all edge cases
- **Add comprehensive testing**: Create additional test cases to verify the heritage clause processing works correctly in complex inheritance scenarios
- **Review AST mapping**: Double-check that `ast.KindTypeReference` correctly corresponds to TypeScript's `TSTypeReference` nodes

---