## Rule: no-import-type-side-effects

### Test File: no-import-type-side-effects.test.ts

### Validation Summary
- ✅ **CORRECT**: Core logic pattern matching for import declarations, checking for inline type qualifiers, skipping type-only imports, handling named imports correctly
- ⚠️ **POTENTIAL ISSUES**: Fix generation logic may have positioning issues, AST node access patterns differ significantly
- ❌ **INCORRECT**: Fix implementation has fundamental flaws in text range calculation and keyword positioning

### Discrepancies Found

#### 1. Fix Generation - Type Keyword Removal Logic
**TypeScript Implementation:**
```typescript
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
```

**Go Implementation:**
```go
for _, specifier := range specifiers {
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
    
    if removeEnd > removeStart {
      fixes = append(fixes, rule.RuleFix{
        Range: core.NewTextRange(removeStart, removeEnd),
        Text: "",
      })
    }
  }
}
```

**Issue:** The Go implementation uses a crude approach to find the "type" keyword by calculating the range between specifier start and identifier start. The TypeScript version uses `getFirstToken(specifier, isTypeKeyword)` to precisely locate the type keyword token.

**Impact:** The Go version may remove incorrect text ranges, potentially including commas, whitespace, or other tokens that should be preserved.

**Test Coverage:** All invalid test cases would be affected by incorrect fix generation.

#### 2. Import Keyword Positioning for Type Insertion
**TypeScript Implementation:**
```typescript
const importKeyword = nullThrows(
  context.sourceCode.getFirstToken(node, isImportKeyword),
  NullThrowsReasons.MissingToken('import keyword', 'import'),
);
fixes.push(fixer.insertTextAfter(importKeyword, ' type'));
```

**Go Implementation:**
```go
importStart := importNode.Pos()
insertPos := importStart + 6  // "import" is 6 characters long
fixes = append(fixes, rule.RuleFix{
  Range: core.NewTextRange(insertPos, insertPos),
  Text: " type",
})
```

**Issue:** The Go implementation hardcodes the import keyword length as 6 characters and adds it to the node position. This is fragile and may not account for whitespace or other tokens between "import" and the import clause.

**Impact:** The "type" keyword may be inserted at the wrong position, potentially breaking the syntax.

**Test Coverage:** All invalid test cases rely on correct insertion of the "type" keyword.

#### 3. AST Node Access Pattern Differences
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

**Issue:** The property names differ: TypeScript uses `importKind !== 'type'` while Go uses `!specifier.IsTypeOnly`. While functionally equivalent, this needs verification that the Go AST property correctly maps to the TypeScript equivalent.

**Impact:** Potential mismatch in detecting type-only specifiers, though this appears to be a correct mapping.

**Test Coverage:** Valid test cases like `"import { type T, U } from 'mod';"` would reveal issues here.

#### 4. Edge Case Handling - Mixed Import Types
**TypeScript Implementation:**
```typescript
if (
  specifier.type !== AST_NODE_TYPES.ImportSpecifier ||
  specifier.importKind !== 'type'
) {
  return;  // Early return if ANY specifier is not type-only
}
```

**Go Implementation:**
```go
if !specifier.IsTypeOnly {
  allTypeOnly = false
  break  // Continue checking but mark as not all type-only
}
```

**Issue:** Both implementations correctly handle the case where not all specifiers are type-only, but the Go version continues processing while TypeScript returns immediately.

**Impact:** Minimal - both achieve the correct result of not reporting when mixed import types exist.

**Test Coverage:** Valid test case `"import { type T, U } from 'mod';"` verifies this behavior.

### Recommendations
- **Critical Fix Required**: Rewrite the fix generation logic to properly locate and remove type keywords using token-based positioning rather than crude range calculations
- **Import Keyword Detection**: Implement proper token scanning to find the import keyword position instead of hardcoding offsets
- **AST Property Verification**: Verify that `specifier.IsTypeOnly` in Go correctly maps to `specifier.importKind === 'type'` in TypeScript
- **Token-Based Text Manipulation**: Consider implementing utility functions similar to TypeScript's `getFirstToken()` for precise token location
- **Test Enhanced Validation**: Add specific test cases to verify fix output matches expected results exactly

---