# Rule Validation: no-useless-empty-export

## Rule: no-useless-empty-export

### Test File: no-useless-empty-export.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic empty export detection logic (`export {}`)
  - Definition file handling (`.d.ts` files are excluded)
  - Core rule message and messageId
  - Fix functionality (removal of empty exports)
  - Import and export assignment detection

- ⚠️ **POTENTIAL ISSUES**: 
  - Immediate processing approach may miss some edge cases
  - Debug statements left in production code
  - Handling of module declarations might be incomplete

- ❌ **INCORRECT**: 
  - Missing TSModuleDeclaration support
  - Incomplete AST node type coverage
  - Different processing logic that may affect correctness

### Discrepancies Found

#### 1. Missing TSModuleDeclaration Support
**TypeScript Implementation:**
```typescript
return {
  Program: checkNode,
  TSModuleDeclaration: checkNode,
};
```

**Go Implementation:**
```go
// Only listens to individual node types, no TSModuleDeclaration equivalent
return rule.RuleListeners{
  ast.KindExportDeclaration: func(node *ast.Node) { ... },
  // ... other listeners
}
```

**Issue:** The TypeScript version checks both Program and TSModuleDeclaration nodes, but the Go version only processes individual export/import nodes without considering module declarations.

**Impact:** Empty exports within TypeScript module declarations may not be detected.

**Test Coverage:** May miss cases with `declare module` blocks containing empty exports.

#### 2. Different Processing Logic
**TypeScript Implementation:**
```typescript
function checkNode(node: TSESTree.Program | TSESTree.TSModuleDeclaration): void {
  const emptyExports: TSESTree.ExportNamedDeclaration[] = [];
  let foundOtherExport = false;

  for (const statement of node.body) {
    if (isEmptyExport(statement)) {
      emptyExports.push(statement);
    } else if (exportOrImportNodeTypes.has(statement.type)) {
      foundOtherExport = true;
    }
  }

  if (foundOtherExport) {
    // Report all empty exports at once
  }
}
```

**Go Implementation:**
```go
// Immediate processing per node type
ast.KindExportDeclaration: func(node *ast.Node) {
  if isEmptyExport(node) {
    emptyExports = append(emptyExports, node)
  } else {
    hasOtherExportsOrImports = true
  }
  
  // Process immediately
  if hasOtherExportsOrImports {
    // Report and clear
  }
}
```

**Issue:** The TypeScript version processes all statements in a module/program at once, while the Go version processes nodes individually and immediately reports when other exports are found.

**Impact:** The Go version may miss some empty exports or report them in different order, and may not handle all cases where multiple empty exports exist.

**Test Coverage:** Test cases with multiple empty exports may behave differently.

#### 3. Incomplete AST Node Type Coverage
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
// Only handles specific cases:
ast.KindExportDeclaration: // covers ExportNamedDeclaration
ast.KindExportAssignment: // covers TSExportAssignment  
ast.KindImportDeclaration: // covers ImportDeclaration
ast.KindImportEqualsDeclaration: // covers TSImportEqualsDeclaration
// Missing: ExportAllDeclaration, ExportDefaultDeclaration, ExportSpecifier
```

**Issue:** The Go version doesn't listen for all the node types that the TypeScript version considers as "other exports/imports".

**Impact:** `export * from 'module'` and `export default` statements may not trigger the rule to report empty exports.

**Test Coverage:** Test cases like `export * from '_'; export {};` may fail.

#### 4. Empty Export Detection Logic Difference
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
  // Empty export is when there's no export clause and no module specifier
  // This represents `export {}`
  return exportDecl.ExportClause == nil && exportDecl.ModuleSpecifier == nil
}
```

**Issue:** The logic is checking different properties. TypeScript checks `specifiers.length === 0 && !node.declaration`, while Go checks `ExportClause == nil && ModuleSpecifier == nil`.

**Impact:** May not correctly identify all forms of empty exports.

**Test Coverage:** Need to verify this works for all test cases with `export {}`.

#### 5. Debug Statements in Production Code
**TypeScript Implementation:**
```typescript
// No debug statements
```

**Go Implementation:**
```go
fmt.Printf("DEBUG: Found ExportDeclaration\n")
fmt.Printf("DEBUG: Found empty export\n")
fmt.Printf("DEBUG: Found non-empty export\n")
fmt.Printf("DEBUG: Reporting %d empty exports\n", len(emptyExports))
```

**Issue:** Debug print statements are left in the production code.

**Impact:** Will produce unwanted output during rule execution.

**Test Coverage:** All test cases will produce debug output.

### Recommendations
- Remove all debug print statements
- Add listeners for missing AST node types: ExportAllDeclaration (likely `ast.KindExportDeclaration` with different structure), ExportDefaultDeclaration  
- Implement proper module-level processing similar to TypeScript's `checkNode` function
- Add support for TSModuleDeclaration equivalent if available in Go AST
- Review and correct the empty export detection logic to match TypeScript behavior
- Restructure the processing logic to collect all statements first, then process them together
- Add comprehensive test cases to verify all edge cases work correctly

---