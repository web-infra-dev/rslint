# Rule: consistent-type-exports

## Test File: consistent-type-exports.test.ts

## Validation Summary
- ✅ **CORRECT**: 
  - Core export declaration handling for named exports
  - Type-based symbol checking logic
  - Configuration option handling (`fixMixedExportsWithInlineTypeSpecifier`)
  - Basic AST pattern matching for export specifiers
  - Message generation for single/multiple type exports
  - Source module extraction from export declarations

- ⚠️ **POTENTIAL ISSUES**:
  - Export * declarations are explicitly disabled in Go version (timeout prevention)
  - Symbol resolution edge cases may differ between TypeScript and Go implementations
  - Complex module resolution behavior may not be identical

- ❌ **INCORRECT**:
  - Missing ExportAllDeclaration handling completely
  - Incorrect AST navigation for export specifier names
  - Missing regex pattern usage for source text manipulation in fixes
  - Program:exit pattern not implemented (using KindEndOfFile instead)

## Discrepancies Found

### 1. Missing ExportAllDeclaration Handler
**TypeScript Implementation:**
```typescript
ExportAllDeclaration(node): void {
  if (node.exportKind === 'type') {
    return;
  }
  // Complex module resolution logic to determine if all exports are types
  // Reports 'typeOverValue' error with asterisk fix
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

**Issue:** The Go implementation completely skips `export *` declarations, while TypeScript handles them with complex module resolution to determine if all re-exported symbols are types.

**Impact:** Missing functionality for `export * from './type-only-module'` cases that should be converted to `export type * from './type-only-module'`.

**Test Coverage:** Multiple test cases are disabled in Go version that test export * scenarios.

### 2. Incorrect Export Specifier Name Handling
**TypeScript Implementation:**
```typescript
function getSpecifierText(specifier: TSESTree.ExportSpecifier): string {
  const exportedName = specifier.exported.type === AST_NODE_TYPES.Literal
    ? specifier.exported.raw
    : specifier.exported.name;
  const localName = specifier.local.type === AST_NODE_TYPES.Literal
    ? specifier.local.raw
    : specifier.local.name;

  return `${localName}${exportedName !== localName ? ` as ${exportedName}` : ''}`;
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

**Issue:** The Go implementation has the AST navigation backwards. In TypeScript AST, `exported` is the right-hand side (after `as`) and `local` is the left-hand side (before `as`). The Go version incorrectly treats `Name()` as exported and `PropertyName` as local.

**Impact:** Export specifier text generation will be incorrect for aliased exports like `export { foo as bar }`, potentially breaking fix generation.

**Test Coverage:** Tests with aliased exports like `export { Type2 as Foo }` may produce incorrect fixes.

### 3. Missing Event Pattern Implementation
**TypeScript Implementation:**
```typescript
return {
  ExportAllDeclaration(node): void { /* ... */ },
  ExportNamedDeclaration(node): void { /* ... */ },
  'Program:exit'(): void {
    // Process all collected exports at the end
    for (const sourceExports of Object.values(sourceExportsMap)) {
      // Report violations
    }
  }
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
    ast.KindExportDeclaration: func(node *ast.Node) { /* ... */ },
    ast.KindEndOfFile: func(node *ast.Node) {
        processExports()
    },
}
```

**Issue:** The Go version uses `KindEndOfFile` instead of the standard `Program:exit` pattern. While functionally similar, this may not be the exact equivalent timing.

**Impact:** Minimal - processing should still occur at the right time, but timing differences could affect edge cases.

**Test Coverage:** All tests should still pass, but subtle timing differences might exist.

### 4. Source Text Manipulation in Fixes
**TypeScript Implementation:**
```typescript
function* fixExportInsertType(fixer, sourceCode, node) {
  const exportToken = nullThrows(sourceCode.getFirstToken(node));
  yield fixer.insertTextAfter(exportToken, ' type');
  
  for (const specifier of node.specifiers) {
    if (specifier.exportKind === 'type') {
      const kindToken = nullThrows(sourceCode.getFirstToken(specifier));
      const firstTokenAfter = nullThrows(sourceCode.getTokenAfter(kindToken));
      yield fixer.removeRange([kindToken.range[0], firstTokenAfter.range[0]]);
    }
  }
}
```

**Go Implementation:**
```go
func fixExportInsertType(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
    // Insert "type" after "export"
    sourceText := string(sourceFile.Text())
    nodeText := sourceText[nodeStart:nodeEnd]
    
    match := exportPattern.FindStringIndex(nodeText)
    if match != nil {
        exportEndPos := nodeStart + match[1]
        fixes = append(fixes, rule.RuleFixReplaceRange(
            core.NewTextRange(exportEndPos, exportEndPos),
            " type",
        ))
    }
    
    // Remove inline "type" specifiers using regex
    typeMatch := typePattern.FindStringIndex(specifierText)
}
```

**Issue:** The Go implementation uses regex patterns for source manipulation instead of proper token-based navigation. This is less precise and may fail with complex formatting or comments.

**Impact:** Fix generation may be less robust, especially with unusual formatting, comments, or edge cases.

**Test Coverage:** The `noFormat` test cases specifically test complex formatting scenarios that the regex approach might not handle correctly.

### 5. Symbol Type Checking Logic Differences
**TypeScript Implementation:**
```typescript
function isSymbolTypeBased(symbol: ts.Symbol | undefined): boolean | undefined {
  if (!symbol) {
    return undefined;
  }

  const aliasedSymbol = tsutils.isSymbolFlagSet(symbol, ts.SymbolFlags.Alias)
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

    aliasedSymbol := symbol
    if utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsAlias) {
        aliasedSymbol = ctx.TypeChecker.GetAliasedSymbol(symbol)
    }

    if ctx.TypeChecker.IsUnknownSymbol(aliasedSymbol) {
        return false, false
    }

    isType := !utils.IsSymbolFlagSet(aliasedSymbol, ast.SymbolFlagsValue)
    return isType, true
}
```

**Issue:** The Go version returns a tuple `(bool, bool)` instead of `bool | undefined`. This changes the logic flow since the TypeScript version uses tri-state logic (true/false/undefined) while Go uses a success flag pattern.

**Impact:** May affect handling of unknown symbols or symbols that can't be resolved, potentially causing different behavior in edge cases.

**Test Coverage:** Tests involving complex symbol resolution scenarios might behave differently.

## Recommendations

### Critical Fixes Needed:
1. **Fix Export Specifier Name Logic**: Correct the AST navigation in `getExportSpecifierName()` to properly handle `local` vs `exported` names.
2. **Implement ExportAllDeclaration Handling**: Either implement the complex module resolution logic or document why it's disabled.
3. **Replace Regex-based Fixes**: Use proper token-based source manipulation instead of regex patterns for more robust fix generation.

### Enhancements:
1. **Symbol Resolution Edge Cases**: Review and test the tri-state vs success-flag pattern difference for symbol type checking.
2. **Token-based Text Manipulation**: Implement proper token navigation for all fix functions.
3. **Complex Formatting Support**: Ensure fixes work correctly with comments and unusual formatting.

### Test Coverage:
1. **Re-enable Export * Tests**: Once ExportAllDeclaration is properly implemented, re-enable the commented test cases.
2. **Add Edge Case Tests**: Test complex formatting scenarios and aliased export edge cases.
3. **Symbol Resolution Tests**: Add tests for unknown symbols and complex type resolution scenarios.

---