## Rule: no-empty-object-type

### Test File: no-empty-object-type.test.ts

### Validation Summary
- ✅ **CORRECT**: Core empty interface detection, empty object type detection, allowInterfaces/allowObjectTypes options, allowWithName regex matching, class declaration merging detection, suggestion generation for fixes
- ⚠️ **POTENTIAL ISSUES**: AST node kind checking patterns, type parameter handling in suggestions, export modifier preservation in fixes
- ❌ **INCORRECT**: Class expression vs class declaration distinction in merged interface detection

### Discrepancies Found

#### 1. Class Expression vs Class Declaration Distinction

**TypeScript Implementation:**
```typescript
const mergedWithClassDeclaration = scope.set
  .get(node.id.name)
  ?.defs.some(
    def => def.node.type === AST_NODE_TYPES.ClassDeclaration,
  );
```

**Go Implementation:**
```go
// Check if merged with class declaration (not class expression)
mergedWithClass := false
if interfaceDecl.Name() != nil {
    symbol := ctx.TypeChecker.GetSymbolAtLocation(interfaceDecl.Name())
    if symbol != nil && symbol.Declarations != nil {
        for _, decl := range symbol.Declarations {
            // Only count class declarations, not class expressions
            if decl.Kind == ast.KindClassDeclaration {
                mergedWithClass = true
                break
            }
        }
    }
}
```

**Issue:** The TypeScript implementation specifically excludes class expressions from being considered as "merged" interfaces, while the Go implementation only checks for `KindClassDeclaration`. However, the Go comment indicates awareness of this distinction, so this may be correct.

**Impact:** Test case with `const derived = class Derived {};` should still show suggestions, which the current Go implementation should handle correctly.

**Test Coverage:** The test case `const derived = class Derived {};` validates this behavior.

#### 2. AST Node Access Patterns

**TypeScript Implementation:**
```typescript
TSInterfaceDeclaration(node): void {
  const extend = node.extends;
  if (node.body.body.length !== 0 || ...)
}
```

**Go Implementation:**
```go
listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
    interfaceDecl := node.AsInterfaceDeclaration()
    
    var extendsList []*ast.Node
    if interfaceDecl.HeritageClauses != nil {
        for _, clause := range interfaceDecl.HeritageClauses.Nodes {
            if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
                extendsList = clause.AsHeritageClause().Types.Nodes
                break
            }
        }
    }
    
    // Check if interface has members
    if interfaceDecl.Members != nil && len(interfaceDecl.Members.Nodes) > 0 {
        return
    }
}
```

**Issue:** The Go implementation uses `HeritageClauses` and searches for `ExtendsKeyword`, which is more verbose but functionally equivalent to the TypeScript `node.extends` direct access.

**Impact:** No functional impact - both approaches correctly identify extends clauses.

**Test Coverage:** All interface extension test cases validate this.

#### 3. Type Parameter Text Extraction

**TypeScript Implementation:**
```typescript
const typeParam = node.typeParameters
  ? context.sourceCode.getText(node.typeParameters)
  : '';
```

**Go Implementation:**
```go
// Get type parameters if any
typeParamsText := ""
if interfaceDecl.TypeParameters != nil {
    typeParamsText = getNodeListTextWithBrackets(ctx, interfaceDecl.TypeParameters)
}
```

**Issue:** The Go implementation uses a custom `getNodeListTextWithBrackets` function that manually reconstructs angle brackets, while TypeScript gets the text directly.

**Impact:** Potential formatting differences in the generated fix text, especially around whitespace and bracket positioning.

**Test Coverage:** The test case `interface Base<T> extends Derived<T> {}` validates this.

#### 4. Export Modifier Detection

**TypeScript Implementation:**
```typescript
// Not explicitly shown in the provided code, but TypeScript ESLint automatically handles export modifiers
```

**Go Implementation:**
```go
// Check for export modifier
exportText := ""
if interfaceDecl.Modifiers() != nil {
    for _, mod := range interfaceDecl.Modifiers().Nodes {
        if mod.Kind == ast.KindExportKeyword {
            exportText = "export "
            break
        }
    }
}
```

**Issue:** The Go implementation explicitly searches for export modifiers, which is good, but the TypeScript implementation might handle this automatically.

**Impact:** Should correctly preserve export modifiers in fix suggestions.

**Test Coverage:** The test case with `export interface Derived extends Base {}` validates this.

#### 5. Intersection Type Detection

**TypeScript Implementation:**
```typescript
if (
  node.members.length ||
  node.parent.type === AST_NODE_TYPES.TSIntersectionType ||
  ...
) {
  return;
}
```

**Go Implementation:**
```go
// Don't report if part of intersection type
if node.Parent != nil && ast.IsIntersectionTypeNode(node.Parent) {
    return
}
```

**Issue:** Both implementations correctly skip empty object types in intersection types, using different but equivalent approaches.

**Impact:** No functional impact.

**Test Coverage:** The test case `type MyNonNullable<T> = T & {};` validates this.

### Recommendations
- **Verify bracket positioning**: Test the `getNodeListTextWithBrackets` function to ensure it produces the same output format as TypeScript ESLint
- **Validate export handling**: Ensure the export modifier detection works correctly with all export syntax variations
- **Test class expression edge cases**: Verify that class expressions (not declarations) don't suppress suggestions as intended
- **Check whitespace handling**: Ensure the fix suggestions maintain proper spacing and formatting

### Overall Assessment
The Go implementation appears to be functionally correct and handles the same core logic as the TypeScript version. The main differences are in implementation details rather than behavioral differences. The rule should work equivalently to the TypeScript-ESLint version.

---