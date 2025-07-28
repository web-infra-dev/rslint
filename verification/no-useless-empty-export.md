## Rule: no-useless-empty-export

### Test File: no-useless-empty-export.test.ts

### Validation Summary
- ✅ **CORRECT**: Error messages match, definition file handling, basic empty export detection, fix generation
- ⚠️ **POTENTIAL ISSUES**: AST node type mapping differences, TSModuleDeclaration handling, complex type-only logic
- ❌ **INCORRECT**: Missing module declaration traversal, export default detection gaps, potential AST structure mismatches

### Discrepancies Found

#### 1. Missing TSModuleDeclaration Traversal
**TypeScript Implementation:**
```typescript
return {
  Program: checkNode,
  TSModuleDeclaration: checkNode,
};
```

**Go Implementation:**
```go
// Only processes SourceFile.Statements directly
for _, statement := range ctx.SourceFile.Statements.Nodes {
  // ...
}
// Return empty listeners since we already processed everything
return rule.RuleListeners{}
```

**Issue:** The TypeScript version registers listeners for both `Program` and `TSModuleDeclaration` nodes, allowing it to check for empty exports within module declarations. The Go version only processes top-level statements and returns empty listeners, missing nested module declarations.

**Impact:** Empty exports within `declare module` or `namespace` declarations would not be detected by the Go implementation.

**Test Coverage:** This affects any test cases with nested module declarations (though none are present in the current test suite).

#### 2. AST Node Type Mapping Inconsistencies
**TypeScript Implementation:**
```typescript
const exportOrImportNodeTypes = new Set([
  AST_NODE_TYPES.ExportAllDeclaration,
  AST_NODE_TYPES.ExportDefaultDeclaration,
  AST_NODE_TYPES.ExportNamedDeclaration,
  AST_NODE_TYPES.ExportSpecifier,
  AST_NODE_TYPES.ImportDeclaration,
  AST_NODE_TYPES.TSExportAssignment,
  AST_NODE_TYPES.TSImportEqualsDeclaration,
]);
```

**Go Implementation:**
```go
case ast.KindExportDeclaration, ast.KindExportAssignment:
  return true 
case ast.KindImportDeclaration:
  // ...
case ast.KindImportEqualsDeclaration:
  return true
```

**Issue:** The TypeScript version explicitly includes `ExportDefaultDeclaration` and `ExportSpecifier` in its export/import detection, while the Go version may not handle these correctly. The Go version uses `KindExportDeclaration` which may not map directly to TypeScript's `ExportNamedDeclaration`.

**Impact:** Export default statements and individual export specifiers might not be properly detected as "other exports" in the Go version.

**Test Coverage:** This affects test cases with `export default` statements.

#### 3. Empty Export Detection Logic Differences
**TypeScript Implementation:**
```typescript
function isEmptyExport(node: TSESTree.Node): node is TSESTree.ExportNamedDeclaration {
  return (
    node.type === AST_NODE_TYPES.ExportNamedDeclaration &&
    node.specifiers.length === 0 &&
    !node.declaration
  );
}
```

**Go Implementation:**
```go
func isEmptyExport(node *ast.Node) bool {
  if node.Kind != ast.KindExportDeclaration {
    return false
  }
  
  exportDecl := node.AsExportDeclaration()
  if exportDecl.ModuleSpecifier != nil {
    return false
  }
  
  if exportDecl.ExportClause == nil {
    return false
  }
  
  if exportDecl.ExportClause.Kind == ast.KindNamedExports {
    namedExports := exportDecl.ExportClause.AsNamedExports()
    return len(namedExports.Elements.Nodes) == 0
  }
  
  return false
}
```

**Issue:** The TypeScript version checks for `!node.declaration` to ensure it's not a declaration export (like `export const x = 1`), while the Go version doesn't have this check. The Go version also has more complex module specifier handling.

**Impact:** The Go version might incorrectly identify some declaration exports as empty exports, or miss some edge cases.

**Test Coverage:** This could affect the behavior with mixed export types.

#### 4. Type-Only Import/Export Filtering Complexity
**TypeScript Implementation:**
```typescript
// Simple node type checking without explicit type-only filtering
exportOrImportNodeTypes.has(statement.type)
```

**Go Implementation:**
```go
case ast.KindExportDeclaration:
  exportDecl := node.AsExportDeclaration()
  // Check if it's a type-only export
  if exportDecl.IsTypeOnly {
    return false
  }
  // ... complex logic for different export types

case ast.KindImportDeclaration:
  importDecl := node.AsImportDeclaration()
  // Skip type-only imports
  if importDecl.ImportClause != nil && importDecl.ImportClause.IsTypeOnly() {
    return false
  }
```

**Issue:** The Go version has extensive type-only filtering logic that the TypeScript version doesn't seem to have. This could lead to different behavior for type-only imports/exports.

**Impact:** Type-only imports and exports might be handled differently between the two implementations, potentially affecting when empty exports are considered "useless".

**Test Coverage:** The test suite includes type-only cases in definition files, which might reveal discrepancies.

#### 5. Missing Support for Export Specifiers
**TypeScript Implementation:**
```typescript
AST_NODE_TYPES.ExportSpecifier, // Included in exportOrImportNodeTypes
```

**Go Implementation:**
```go
// No explicit handling for individual export specifiers
```

**Issue:** The TypeScript version includes `ExportSpecifier` as a type that indicates the presence of other exports, but the Go version doesn't have equivalent logic.

**Impact:** Files with only export specifiers (like re-exports) might not be properly detected as having "other exports".

**Test Coverage:** This affects test cases with `export { _ }` syntax.

### Recommendations
- **Add TSModuleDeclaration traversal**: Implement proper AST traversal to handle nested module declarations
- **Fix AST node type mapping**: Ensure Go AST node types properly correspond to TypeScript ESLint node types
- **Add declaration export filtering**: Include logic to distinguish between empty exports and declaration exports
- **Simplify type-only logic**: Align type-only import/export filtering with the TypeScript implementation
- **Add export specifier detection**: Ensure individual export specifiers are properly detected as "other exports"
- **Add comprehensive test cases**: Include tests for nested modules, export defaults, and mixed export scenarios

---