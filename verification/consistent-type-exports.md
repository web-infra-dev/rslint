# Rule: consistent-type-exports

## Test File: consistent-type-exports.test.ts

### Validation Summary
- ✅ **CORRECT**: Core export detection logic, type-based specifier identification, basic configuration handling, error reporting structure, AST pattern matching for ExportDeclaration
- ⚠️ **POTENTIAL ISSUES**: Export * handling disabled due to timeout concerns, getExportSpecifierName function logic differs from TypeScript version, incomplete validation of external module filtering
- ❌ **INCORRECT**: Export * from module type-only detection missing, export specifier name formatting differs, external module detection logic may be overly restrictive

### Discrepancies Found

#### 1. Export All Declaration Handling
**TypeScript Implementation:**
```typescript
ExportAllDeclaration(node): void {
  if (node.exportKind === 'type') {
    return;
  }
  
  const sourceModule = ts.resolveModuleName(
    node.source.value,
    context.filename,
    services.program.getCompilerOptions(),
    ts.sys,
  );
  // Complex logic to determine if all exports are type-only
  const isThereAnyExportedValue = checker
    .getPropertiesOfType(sourceFileType)
    .some(propertyTypeSymbol =>
      checker.getPropertyOfType(
        sourceFileType,
        propertyTypeSymbol.escapedName.toString(),
      ) != null,
    );
  if (isThereAnyExportedValue) {
    return;
  }
  
  context.report({
    node,
    messageId: 'typeOverValue',
    // ... fix logic
  });
}
```

**Go Implementation:**
```go
// Skip export * from '...' and export * as name from '...' declarations
// These require complex module resolution which can cause issues
if exportDecl.ModuleSpecifier != nil && (exportDecl.ExportClause == nil || ast.IsNamespaceExport(exportDecl.ExportClause)) {
  // Skip export * declarations to avoid timeouts
  return
}
```

**Issue:** The Go implementation completely skips export * declarations to avoid timeout issues, while the TypeScript version has sophisticated logic to detect when all exports from a module are type-only.

**Impact:** Test cases for `export * from './type-only-module'` will not be caught by the Go implementation, leading to false negatives.

**Test Coverage:** Multiple test cases are disabled in the RSLint test file due to this limitation.

#### 2. Export Specifier Name Formatting
**TypeScript Implementation:**
```typescript
function getSpecifierText(specifier: TSESTree.ExportSpecifier): string {
  const exportedName =
    specifier.exported.type === AST_NODE_TYPES.Literal
      ? specifier.exported.raw
      : specifier.exported.name;
  const localName =
    specifier.local.type === AST_NODE_TYPES.Literal
      ? specifier.local.raw
      : specifier.local.name;

  return `${localName}${
    exportedName !== localName ? ` as ${exportedName}` : ''
  }`;
}
```

**Go Implementation:**
```go
func getExportSpecifierName(specifier *ast.ExportSpecifier) string {
  // In TypeScript AST:
  // - Name returns the exported name (what shows up after 'as' or the identifier if no 'as')
  // - PropertyName returns the local name (what appears before 'as', if present)

  exported := specifier.Name()
  local := specifier.PropertyName

  // If no propertyName, then local and exported are the same
  if local == nil {
    return getIdentifierName(exported)
  }

  exportedName := getIdentifierName(exported)
  localName := getIdentifierName(local)
  
  return fmt.Sprintf("%s as %s", localName, exportedName)
}
```

**Issue:** The Go implementation's understanding of TypeScript AST structure differs. The TypeScript version uses `specifier.local` as the original name and `specifier.exported` as the exported name, while Go implementation assumes the opposite relationship.

**Impact:** Export specifier names in error messages and fixes may be incorrectly formatted, especially for aliased exports.

**Test Coverage:** Tests with aliased exports like `Type2 as Foo` may show incorrect formatting.

#### 3. External Module Detection
**TypeScript Implementation:**
```typescript
function isSymbolTypeBased(
  symbol: ts.Symbol | undefined,
): boolean | undefined {
  if (!symbol) {
    return undefined;
  }

  const aliasedSymbol = tsutils.isSymbolFlagSet(
    symbol,
    ts.SymbolFlags.Alias,
  )
    ? checker.getAliasedSymbol(symbol)
    : symbol;

  if (checker.isUnknownSymbol(aliasedSymbol)) {
    return undefined;
  }

  return !(aliasedSymbol.flags & ts.SymbolFlags.Value);
}
```

**Go Implementation:**
```go
isSymbolTypeBased := func(symbol *ast.Symbol) (bool, bool) {
  if symbol == nil {
    return false, false
  }

  // Check if this symbol is from an external module first
  if symbol.Declarations != nil && len(symbol.Declarations) > 0 {
    decl := symbol.Declarations[0]
    if decl != nil {
      sourceFile := ast.GetSourceFileOfNode(decl)
      if sourceFile != nil {
        fileName := sourceFile.FileName()
        // Skip external modules completely
        if strings.Contains(fileName, "node_modules") ||
           strings.HasPrefix(fileName, "/usr/") ||
           strings.HasPrefix(fileName, "C:\\") ||
           !strings.HasSuffix(fileName, ".ts") && !strings.HasSuffix(fileName, ".tsx") && !strings.HasSuffix(fileName, ".js") && !strings.HasSuffix(fileName, ".jsx") {
          return false, false
        }
      }
    }
  }
  // ... rest of logic
}
```

**Issue:** The Go implementation has additional external module filtering logic that the TypeScript version doesn't have. This could lead to different behavior for symbols from certain modules.

**Impact:** Symbols from modules that don't match the Go version's file extension criteria may be incorrectly ignored.

**Test Coverage:** The first test case `export { Foo } from 'foo';` relies on this behavior but may behave differently than expected.

#### 4. Error Message Formatting
**TypeScript Implementation:**
```typescript
messages: {
  multipleExportsAreTypes:
    'Type exports {{exportNames}} are not values and should be exported using `export type`.',
  singleExportIsType:
    'Type export {{exportNames}} is not a value and should be exported using `export type`.',
  typeOverValue:
    'All exports in the declaration are only used as types. Use `export type`.',
}
```

**Go Implementation:**
```go
message := fmt.Sprintf("Type export %s is not a value and should be exported using `export type`.", exportNames)
message := fmt.Sprintf("Type exports %s are not values and should be exported using `export type`.", exportNames)
ctx.ReportNodeWithFixes(report.Node, rule.RuleMessage{
  Id:          "typeOverValue",
  Description: "All exports in the declaration are only used as types. Use `export type`.",
})
```

**Issue:** The Go implementation constructs error messages manually instead of using templated messages, and the message ID usage differs from the TypeScript version.

**Impact:** Error messages may not match exactly, potentially causing test failures or inconsistent user experience.

**Test Coverage:** All test cases expect specific messageId values that may not match the Go implementation's approach.

#### 5. Program Exit Processing
**TypeScript Implementation:**
```typescript
'Program:exit'(): void {
  for (const sourceExports of Object.values(sourceExportsMap)) {
    // Process all collected exports at the end
  }
}
```

**Go Implementation:**
```go
ast.KindEndOfFile: func(node *ast.Node) {
  processExports()
}
```

**Issue:** The Go implementation uses `KindEndOfFile` instead of a program exit event, which may have different timing or behavior.

**Impact:** The processing timing might differ, potentially affecting the order of error reporting or missing some export declarations.

**Test Coverage:** All tests depend on proper end-of-program processing to generate errors.

### Recommendations
- **Critical**: Implement proper export * declaration handling with module resolution to match TypeScript behavior
- **High Priority**: Fix export specifier name formatting to match TypeScript AST understanding
- **Medium Priority**: Align error message formatting and message ID usage with TypeScript version
- **Medium Priority**: Review external module detection logic to ensure it matches TypeScript behavior
- **Low Priority**: Consider using Program exit events instead of EndOfFile if available in the Go AST

### Test Cases Needing Enhancement
- Re-enable export * test cases once proper module resolution is implemented
- Add specific tests for export specifier name formatting edge cases
- Verify error message formatting matches exactly with expected messageId values
- Test external module detection behavior with various file patterns

---