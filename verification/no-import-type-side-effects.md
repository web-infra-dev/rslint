## Rule: no-import-type-side-effects

### Test File: no-import-type-side-effects.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic rule logic pattern, correct AST node targeting, proper message ID and description
- ⚠️ **POTENTIAL ISSUES**: Fix generation logic may not handle text positioning accurately, missing utility functions for token detection
- ❌ **INCORRECT**: Fix implementation uses hardcoded positions instead of proper token detection, may produce malformed fixes

### Discrepancies Found

#### 1. Fix Generation Logic - Token Detection vs Hardcoded Positions

**TypeScript Implementation:**
```typescript
fix(fixer) {
  const fixes: TSESLint.RuleFix[] = [];
  for (const specifier of specifiers) {
    const qualifier = nullThrows(
      context.sourceCode.getFirstToken(specifier, isTypeKeyword),
      NullThrowsReasons.MissingToken('type keyword', 'import specifier'),
    );
    fixes.push(
      fixer.removeRange([
        qualifier.range[0],
        specifier.imported.range[0],
      ]),
    );
  }

  const importKeyword = nullThrows(
    context.sourceCode.getFirstToken(node, isImportKeyword),
    NullThrowsReasons.MissingToken('import keyword', 'import'),
  );
  fixes.push(fixer.insertTextAfter(importKeyword, ' type'));

  return fixes;
}
```

**Go Implementation:**
```go
func createFix(ctx rule.RuleContext, importNode *ast.Node, specifiers []*ast.ImportSpecifier) []rule.RuleFix {
  // ...
  for _, specifier := range specifiers {
    // Gets the entire specifier range and identifier range
    specifierRange := utils.TrimNodeTextRange(ctx.SourceFile, specifier.AsNode())
    
    var identifierNode *ast.Node
    if specifier.PropertyName != nil {
      identifierNode = specifier.PropertyName
    } else {
      identifierNode = specifier.Name()
    }
    
    if identifierNode != nil {
      identifierRange := utils.TrimNodeTextRange(ctx.SourceFile, identifierNode)
      removeStart := specifierRange.Pos()
      removeEnd := identifierRange.Pos()
      // ...
    }
  }
  
  // Hardcoded position calculation
  importStart := importNode.Pos()
  insertPos := importStart + 6  // "import" is 6 characters long
  // ...
}
```

**Issue:** The Go implementation uses hardcoded position arithmetic (`importStart + 6`) instead of properly detecting the import keyword token, and calculates remove ranges based on node positions rather than finding the actual "type" keyword tokens.

**Impact:** This could produce incorrect fixes if there are comments, whitespace, or other tokens between elements. The TypeScript version uses proper token detection to find exact keyword positions.

**Test Coverage:** All invalid test cases rely on fix functionality, so this affects all of them.

#### 2. Missing Utility Functions for Token Detection

**TypeScript Implementation:**
```typescript
// Uses utility functions for precise token detection
isImportKeyword, isTypeKeyword, nullThrows
context.sourceCode.getFirstToken(specifier, isTypeKeyword)
context.sourceCode.getFirstToken(node, isImportKeyword)
```

**Go Implementation:**
```go
// No equivalent token detection utilities used
// Relies on AST node ranges and hardcoded offsets
```

**Issue:** The Go implementation lacks the token-level precision of the TypeScript version. It doesn't have equivalents to `isImportKeyword`, `isTypeKeyword`, or `getFirstToken` utilities.

**Impact:** Less robust fix generation that may fail with unusual formatting or syntax variations.

**Test Coverage:** Could cause issues with any test case involving fixes, especially with complex formatting.

#### 3. Specifier Type Checking Logic

**TypeScript Implementation:**
```typescript
for (const specifier of node.specifiers) {
  if (
    specifier.type !== AST_NODE_TYPES.ImportSpecifier ||
    specifier.importKind !== 'type'
  ) {
    return;
  }
  specifiers.push(specifier);
}
```

**Go Implementation:**
```go
for _, element := range namedImports.Elements.Nodes {
  if !ast.IsImportSpecifier(element) {
    allTypeOnly = false
    break
  }
  
  specifier := element.AsImportSpecifier()
  if !specifier.IsTypeOnly {
    allTypeOnly = false
    break
  }
  
  typeOnlySpecifiers = append(typeOnlySpecifiers, specifier)
}
```

**Issue:** The logic structure is slightly different. TypeScript checks each specifier individually and returns early if any non-type specifier is found. Go uses a flag-based approach but the core logic should be equivalent.

**Impact:** Minimal - both approaches should catch the same cases, but the Go version is more verbose.

**Test Coverage:** Should work correctly for all test cases.

#### 4. Property Name vs Imported Name Handling

**TypeScript Implementation:**
```typescript
// Uses specifier.imported.range[0] directly
fixer.removeRange([
  qualifier.range[0],
  specifier.imported.range[0],
])
```

**Go Implementation:**
```go
var identifierNode *ast.Node
if specifier.PropertyName != nil {
  identifierNode = specifier.PropertyName
} else {
  identifierNode = specifier.Name()
}
```

**Issue:** The Go implementation checks for PropertyName vs Name, but the TypeScript version uses `specifier.imported` which may map differently in the AST structure.

**Impact:** Could affect handling of aliased imports (`import { type A as B }`).

**Test Coverage:** Test cases with aliases (`type A as AA`) may reveal discrepancies.

### Recommendations
- Implement proper token detection utilities similar to `isImportKeyword` and `isTypeKeyword`
- Replace hardcoded position arithmetic with token-based positioning
- Add utility functions to find specific tokens within nodes (equivalent to `getFirstToken`)
- Verify the PropertyName vs Name logic matches TypeScript's `imported` property semantics
- Add error handling for cases where expected tokens are not found
- Consider implementing a source code utility similar to TypeScript's `sourceCode` for token-level operations

---