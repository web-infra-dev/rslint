## Rule: consistent-type-definitions

### Test File: consistent-type-definitions.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core rule logic structure (interface vs type modes)
  - Basic AST pattern matching for TypeAliasDeclaration and InterfaceDeclaration
  - Error message consistency
  - Option parsing (default "interface", accepts string or array)
  - Parenthesized type unwrapping logic
  - Basic fix generation structure
  - Export default interface handling
  - Declare global module detection
  
- ⚠️ **POTENTIAL ISSUES**: 
  - Complex text manipulation in fixes may not handle all whitespace/formatting edge cases
  - Heritage clause processing (extends) might have subtle differences
  - AST selector pattern differences between implementations
  
- ❌ **INCORRECT**: 
  - TypeScript uses AST selector pattern `"TSTypeAliasDeclaration[typeAnnotation.type='TSTypeLiteral']"` while Go manually checks node types
  - Missing export declare handling in Go implementation
  - Potential differences in fix text generation for complex cases

### Discrepancies Found

#### 1. AST Pattern Matching Strategy
**TypeScript Implementation:**
```typescript
"TSTypeAliasDeclaration[typeAnnotation.type='TSTypeLiteral']"(
  node: TSESTree.TSTypeAliasDeclaration,
): void {
  // Automatically filtered to only type aliases with type literals
}
```

**Go Implementation:**
```go
listeners[ast.KindTypeAliasDeclaration] = func(node *ast.Node) {
  typeAlias := node.AsTypeAliasDeclaration()
  if typeAlias.Type != nil {
    actualType := unwrapParentheses(typeAlias.Type)
    if actualType != nil && actualType.Kind == ast.KindTypeLiteral {
      // Manual filtering required
    }
  }
}
```

**Issue:** The TypeScript version uses a sophisticated AST selector that automatically filters to only type alias declarations with type literal annotations, while the Go version manually checks the type after receiving all type alias declarations.

**Impact:** Both approaches should work correctly, but the Go version processes more nodes than necessary.

**Test Coverage:** All `type T = { ... }` test cases rely on this pattern matching.

#### 2. Export Declare Handling
**TypeScript Implementation:**
```typescript
// Has explicit test case for:
export declare type Test = {
  foo: string;
  bar: string;
};
```

**Go Implementation:**
```go
// No explicit handling for 'declare' keyword in export scenarios
// May not properly detect and handle "export declare" patterns
```

**Issue:** The Go implementation doesn't appear to have specific logic for handling the `declare` keyword in export scenarios, which is tested in the TypeScript version.

**Impact:** Test cases with `export declare type` or `export declare interface` may not be handled correctly.

**Test Coverage:** Test cases with `export declare type Test = {...}` and `export declare interface Test {...}`

#### 3. Fix Generation Text Manipulation
**TypeScript Implementation:**
```typescript
const beforeEqualsToken = nullThrows(
  context.sourceCode.getTokenBefore(equalsToken, {
    includeComments: true,
  }),
  NullThrowsReasons.MissingToken('before =', 'type alias'),
);

return [
  fixer.replaceText(typeToken, 'interface'),
  fixer.replaceTextRange(
    [beforeEqualsToken.range[1], node.typeAnnotation.range[0]],
    ' ',
  ),
  fixer.removeRange([
    node.typeAnnotation.range[1],
    node.range[1],
  ]),
];
```

**Go Implementation:**
```go
// Replace equals and everything up to the actual type literal
if equalsStart >= 0 {
  fixes = append(fixes, rule.RuleFix{
    Text:  " " + getNodeText(ctx, actualTypeLiteral),
    Range: core.TextRange{}.WithPos(equalsStart).WithEnd(int(typeAlias.Type.End())),
  })
}
```

**Issue:** The TypeScript version has more sophisticated token-based manipulation that preserves comments and handles edge cases better. The Go version uses simpler text range replacement that might not handle all formatting scenarios.

**Impact:** Edge cases with comments or complex formatting might not produce identical output.

**Test Coverage:** Test case with `type T /* comment */={ x: number; };`

#### 4. Heritage Clause Processing Detail
**TypeScript Implementation:**
```typescript
node.extends.forEach(heritage => {
  const typeIdentifier = context.sourceCode.getText(heritage);
  fixes.push(
    fixer.insertTextAfter(node.body, ` & ${typeIdentifier}`),
  );
});
```

**Go Implementation:**
```go
if interfaceDecl.HeritageClauses != nil {
  for _, clause := range interfaceDecl.HeritageClauses.Nodes {
    if clause.Kind == ast.KindHeritageClause {
      heritageClause := clause.AsHeritageClause()
      if heritageClause.Token == ast.KindExtendsKeyword && len(heritageClause.Types.Nodes) > 0 {
        for _, heritageType := range heritageClause.Types.Nodes {
          typeText := getNodeText(ctx, heritageType)
          intersectionText += fmt.Sprintf(" & %s", typeText)
        }
      }
    }
  }
}
```

**Issue:** The Go version has more complex logic for finding heritage clauses and checking for extends keyword, while TypeScript directly accesses `node.extends`. This suggests potential differences in AST structure handling.

**Impact:** Should work correctly but represents different approaches to the same functionality.

**Test Coverage:** Test cases with `interface A extends B, C { x: number; }`

#### 5. Export Default Interface Detection
**TypeScript Implementation:**
```typescript
if (
  node.parent.type === AST_NODE_TYPES.ExportDefaultDeclaration
) {
  fixes.push(
    fixer.removeRange([node.parent.range[0], node.range[0]]),
    fixer.insertTextAfter(
      node.body,
      `\nexport default ${node.id.name}`,
    ),
  );
}
```

**Go Implementation:**
```go
// Look backwards from interface keyword to see if we have "export default"
textBefore := text[searchStart:interfaceKeywordStart]
if strings.Contains(textBefore, "export") && strings.Contains(textBefore, "default") {
  isExportDefault = true
}
```

**Issue:** The TypeScript version uses proper AST parent node checking, while the Go version uses text-based detection by looking backwards in the source text. This is less robust.

**Impact:** Edge cases with complex formatting or comments between export/default/interface might not be detected correctly.

**Test Coverage:** Test case with `export default interface Test { ... }`

### Recommendations
- **Fix AST selector approach**: Consider implementing a more efficient filtering mechanism for type aliases with type literals to avoid processing unnecessary nodes
- **Add export declare support**: Implement proper handling for `export declare type` and `export declare interface` patterns
- **Improve fix generation**: Enhance text manipulation to better handle comments, whitespace, and edge cases similar to the TypeScript version
- **Strengthen export default detection**: Use AST-based parent node checking instead of text-based detection for export default interfaces
- **Add comprehensive edge case testing**: Ensure all complex formatting scenarios (comments, whitespace, parentheses) are properly tested

---