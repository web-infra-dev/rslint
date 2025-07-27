# Rule Validation: no-empty-interface

## Rule: no-empty-interface

### Test File: no-empty-interface.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic empty interface detection
  - Single extends detection and allowSingleExtends option handling
  - Type parameter preservation in fix replacements
  - Merged class declaration detection logic
  - Auto-fix generation for interface-to-type conversion
  - Message IDs and descriptions match the original

- ⚠️ **POTENTIAL ISSUES**: 
  - Ambient declaration detection may be incomplete
  - Scope-based checking differences between TypeScript and Go implementations
  - Missing suggestion support for ambient declarations

- ❌ **INCORRECT**: 
  - No significant logic errors found, but some implementation details differ

### Discrepancies Found

#### 1. Ambient Declaration Detection Logic
**TypeScript Implementation:**
```typescript
const isInAmbientDeclaration =
  isDefinitionFile(context.filename) &&
  scope.type === ScopeType.tsModule &&
  scope.block.declare;
```

**Go Implementation:**
```go
isInAmbientDeclaration := false
if strings.HasSuffix(ctx.SourceFile.FileName(), ".d.ts") {
  // Check if we're inside a declared module
  parent := node.Parent
  for parent != nil {
    if parent.Kind == ast.KindModuleDeclaration {
      moduleDecl := parent.AsModuleDeclaration()
      modifiers := moduleDecl.Modifiers()
      if modifiers != nil {
        for _, modifier := range modifiers.Nodes {
          if modifier.Kind == ast.KindDeclareKeyword {
            isInAmbientDeclaration = true
            break
          }
        }
      }
    }
    if isInAmbientDeclaration {
      break
    }
    parent = parent.Parent
  }
}
```

**Issue:** The Go implementation manually traverses parent nodes to find declared modules, while the TypeScript version uses scope information. The Go approach may miss some ambient declaration scenarios or be less precise about scope types.

**Impact:** May incorrectly apply auto-fixes in ambient contexts where suggestions should be used instead.

**Test Coverage:** The test case with `declare module FooBar` should reveal this - it expects suggestions instead of auto-fixes.

#### 2. Scope-based Merged Declaration Detection
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
mergedWithClassDeclaration := false
if ctx.TypeChecker != nil {
  symbol := ctx.TypeChecker.GetSymbolAtLocation(interfaceDecl.Name())
  if symbol != nil {
    // Check if this symbol has a class declaration
    for _, decl := range symbol.Declarations {
      if decl.Kind == ast.KindClassDeclaration {
        mergedWithClassDeclaration = true
        break
      }
    }
  }
}
```

**Issue:** The Go implementation uses TypeScript symbol information while the TypeScript ESLint version uses scope manager data. This could lead to different results in edge cases, though the fundamental logic is similar.

**Impact:** May affect whether auto-fixes or suggestions are provided for merged interface/class declarations.

**Test Coverage:** Test cases with merged class declarations should validate this behavior.

#### 3. Missing Suggestion Support
**TypeScript Implementation:**
```typescript
context.report({
  node: node.id,
  messageId: 'noEmptyWithSuper',
  ...(useAutoFix
    ? { fix }
    : !mergedWithClassDeclaration
      ? {
          suggest: [
            {
              messageId: 'noEmptyWithSuper',
              fix,
            },
          ],
        }
      : null),
});
```

**Go Implementation:**
```go
if isInAmbientDeclaration || mergedWithClassDeclaration {
  // Just report without fix or suggestion for ambient declarations or merged class declarations
  ctx.ReportNode(interfaceDecl.Name(), message)
} else {
  // Use auto-fix for non-ambient, non-merged cases
  ctx.ReportNodeWithFixes(interfaceDecl.Name(), message,
    rule.RuleFixReplace(ctx.SourceFile, node, replacement))
}
```

**Issue:** The Go implementation doesn't provide suggestions for ambient declarations where merged class declarations are not involved. The TypeScript version provides suggestions in ambient contexts when there's no class merging.

**Impact:** The `.d.ts` test case expects suggestions but the Go implementation may not provide them.

**Test Coverage:** The `declare module FooBar` test case expects suggestions to be provided.

#### 4. Heritage Clause Processing
**TypeScript Implementation:**
```typescript
const extend = node.extends;
if (extend.length === 0) {
  // ... handle no extends case
} else if (extend.length === 1 && !allowSingleExtends) {
  // ... handle single extends case
}
```

**Go Implementation:**
```go
extendCount := 0
var extendClause *ast.HeritageClause
if interfaceDecl.HeritageClauses != nil {
  for _, clause := range interfaceDecl.HeritageClauses.Nodes {
    heritageClause := clause.AsHeritageClause()
    if heritageClause.Token == ast.KindExtendsKeyword {
      extendClause = heritageClause
      extendCount = len(heritageClause.Types.Nodes)
      break
    }
  }
}
```

**Issue:** The Go implementation correctly processes heritage clauses but uses a different traversal pattern. This should work correctly but the approach differs from the direct access in TypeScript.

**Impact:** Should have no functional impact, but the logic paths are different.

**Test Coverage:** All test cases with extends clauses should work correctly.

### Recommendations
- **Fix suggestion support**: Implement suggestion reporting in the Go version for ambient declarations that don't have merged class declarations
- **Verify ambient declaration detection**: Ensure the parent traversal logic correctly identifies all ambient declaration contexts
- **Add scope-based validation**: Consider whether additional scope-based checks are needed to match TypeScript ESLint's behavior more precisely
- **Test edge cases**: Add more test cases for complex ambient declaration scenarios and nested module declarations

### Test Cases That Need Special Attention
1. The `declare module FooBar` test case - should provide suggestions, not just errors
2. Merged interface/class declaration test cases - ensure proper detection
3. Complex type parameter scenarios - verify text extraction accuracy
4. Nested ambient declaration contexts - ensure proper detection

---