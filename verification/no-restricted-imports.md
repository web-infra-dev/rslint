## Rule: no-restricted-imports

### Test File: no-restricted-imports.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic import/export restriction logic, path and pattern matching structure, type-only import handling concept, configuration option parsing structure
- ⚠️ **POTENTIAL ISSUES**: ExportAllDeclaration handling, inline type import detection (`import { type Bar }`), regex flag handling, glob pattern compilation differences
- ❌ **INCORRECT**: Missing ExportAllDeclaration listener, incorrect type-only detection logic, missing individual type specifier checking, potential regex compilation differences

### Discrepancies Found

#### 1. Missing ExportAllDeclaration Support
**TypeScript Implementation:**
```typescript
return {
  ExportAllDeclaration: rules.ExportAllDeclaration,
  // ... other listeners
};
```

**Go Implementation:**
```go
// ExportAllDeclaration listener is completely missing
return rule.RuleListeners{
  ast.KindImportDeclaration: func(node *ast.Node) { ... },
  ast.KindExportDeclaration: func(node *ast.Node) { ... },
  ast.KindImportEqualsDeclaration: func(node *ast.Node) { ... },
}
```

**Issue:** The Go implementation lacks support for `export * from 'module'` statements, which are explicitly handled in the TypeScript version.

**Impact:** Test case `export * from 'import1'` will not be caught by the Go implementation.

**Test Coverage:** Test case with `export * from 'import1'` with options `['import1']` expects an error but would pass in Go implementation.

#### 2. Incorrect Type-Only Import Detection
**TypeScript Implementation:**
```typescript
function checkImportNode(node: TSESTree.ImportDeclaration): void {
  if (
    node.importKind === 'type' ||
    (node.specifiers.length > 0 &&
      node.specifiers.every(
        specifier =>
          specifier.type === AST_NODE_TYPES.ImportSpecifier &&
          specifier.importKind === 'type',
      ))
  ) {
    // Handle type-only imports
  }
}
```

**Go Implementation:**
```go
isTypeOnlyImportExport := func(node *ast.Node) bool {
  switch node.Kind {
  case ast.KindImportDeclaration:
    importDecl := node.AsImportDeclaration()
    if importDecl.ImportClause == nil {
      return false
    }
    
    importClause := importDecl.ImportClause.AsImportClause()
    if importClause.IsTypeOnly {
      return true
    }
    
    // Check if all specifiers are type-only
    if importClause.NamedBindings != nil && ast.IsNamedImports(importClause.NamedBindings) {
      namedImports := importClause.NamedBindings.AsNamedImports()
      if namedImports.Elements != nil {
        allTypeOnly := true
        for _, elem := range namedImports.Elements.Nodes {
          if !elem.AsImportSpecifier().IsTypeOnly {
            allTypeOnly = false
            break
          }
        }
        return allTypeOnly
      }
    }
```

**Issue:** The Go implementation doesn't properly handle inline type imports like `import { type Bar }` where only some specifiers are type-only. The TypeScript version checks `specifier.importKind === 'type'` but the Go version checks `elem.AsImportSpecifier().IsTypeOnly` which may not be equivalent.

**Impact:** Test cases with mixed type/value imports like `import { Bar, type Baz }` may not be handled correctly.

**Test Coverage:** Test case `import { type Bar } from 'import-foo'` and `import { Bar, type Baz } from 'import-foo'` may behave differently.

#### 3. Regex Flag Handling Differences
**TypeScript Implementation:**
```typescript
if (restrictedPattern.regex) {
  allowedImportTypeRegexMatchers.push(
    new RegExp(
      restrictedPattern.regex,
      restrictedPattern.caseSensitive ? 'u' : 'iu',
    ),
  );
}
```

**Go Implementation:**
```go
if pr.Regex != "" {
  flags := "(?s)" // s flag for . to match newlines
  if !pr.CaseSensitive {
    flags += "(?i)" // i flag for case insensitive
  }
  re, err := regexp.Compile(flags + pr.Regex)
}
```

**Issue:** The TypeScript version uses 'u' (Unicode) flag and 'iu' (case insensitive + Unicode) flags, while the Go version uses inline flags `(?s)` (dotall) and `(?i)` (case insensitive). The 's' flag in Go makes `.` match newlines, which may not be the intended behavior.

**Impact:** Regex patterns may match differently between implementations, especially with Unicode characters and newline handling.

**Test Coverage:** Regex pattern tests may pass/fail differently between implementations.

#### 4. ImportNames Logic Complexity
**TypeScript Implementation:**
```typescript
// The TypeScript version delegates much of the logic to the base ESLint rule
const rules = baseRule.create(context);
// Then wraps and filters the results
```

**Go Implementation:**
```go
// Check if specific import names are restricted
if len(pr.ImportNames) > 0 && len(importedNames) > 0 {
  for _, importedName := range importedNames {
    for _, restrictedName := range pr.ImportNames {
      if importedName == restrictedName {
        // Report error for each match
      }
    }
  }
}
```

**Issue:** The Go implementation reports multiple errors for each restricted import name, while the TypeScript implementation may handle this differently through the base rule delegation.

**Impact:** Test cases with multiple restricted import names may generate different numbers of errors.

**Test Coverage:** Test case `import { Bar, type Baz } from 'import-foo'` expects exactly 2 errors, which needs verification.

#### 5. Message ID Inconsistencies
**TypeScript Implementation:**
```typescript
// Uses base rule's message IDs: 'path', 'pathWithCustomMessage', 'importNameWithCustomMessage', 'patterns', 'patternWithCustomMessage'
```

**Go Implementation:**
```go
messageId := "path"
if pr.Message != "" {
  messageId = "pathWithCustomMessage"
}
// Similar logic for patterns
```

**Issue:** The Go implementation hardcodes message IDs rather than using a consistent mapping from the base rule.

**Impact:** Error messages may not match exactly between implementations.

**Test Coverage:** All invalid test cases specify expected messageId values.

#### 6. Export Named Declaration with Source Logic
**TypeScript Implementation:**
```typescript
'ExportNamedDeclaration[source]'(
  node: {
    source: NonNullable<TSESTree.ExportNamedDeclaration['source']>;
  } & TSESTree.ExportNamedDeclaration,
): void {
  if (
    node.exportKind === 'type' ||
    (node.specifiers.length > 0 &&
      node.specifiers.every(specifier => specifier.exportKind === 'type'))
  ) {
    // Handle type-only exports
  }
}
```

**Go Implementation:**
```go
ast.KindExportDeclaration: func(node *ast.Node) {
  exportDecl := node.AsExportDeclaration()
  if exportDecl.ModuleSpecifier == nil {
    return
  }
  // ... rest of logic
}
```

**Issue:** The TypeScript version uses a selector `'ExportNamedDeclaration[source]'` to only handle exports with a source, while the Go version checks `ModuleSpecifier == nil` inside the handler. The logic should be equivalent but the structure is different.

**Impact:** Should work correctly but the approach differs.

**Test Coverage:** Export re-export test cases should validate this behavior.

### Recommendations
- Add `ast.KindExportAllDeclaration` listener to handle `export * from 'module'` statements
- Fix type-only import detection to properly handle inline type specifiers (`import { type Bar }`)
- Review regex flag usage - remove `(?s)` flag unless dotall behavior is specifically needed
- Verify message ID mapping matches the base ESLint rule exactly
- Test the importNames logic thoroughly to ensure error count matches expectations
- Add comprehensive tests for mixed type/value import scenarios
- Validate that export re-export logic works correctly for all test cases

---