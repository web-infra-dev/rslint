# Rule Validation: consistent-type-definitions

## Rule: consistent-type-definitions

### Test File: consistent-type-definitions.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic rule logic for detecting type aliases with type literals vs interfaces
  - Option parsing for "interface" vs "type" preferences
  - Core AST pattern matching for TypeAliasDeclaration and InterfaceDeclaration
  - Error message IDs and descriptions match exactly
  - Basic fix generation for simple cases
  - Handling of parentheses around type literals
  - Support for type parameters in both directions

- ⚠️ **POTENTIAL ISSUES**:
  - Export default interface handling complexity may have edge cases
  - Declare global module detection logic differences
  - Text replacement range calculations might differ in edge cases
  - Heritage clause handling for multiple extends might not preserve exact formatting

- ❌ **INCORRECT**:
  - Missing proper handling of `declare` keyword in export declarations
  - Incorrect detection of declare global modules (checking for wrong AST structure)
  - Export default conversion missing proper newline handling

### Discrepancies Found

#### 1. Declare Global Module Detection

**TypeScript Implementation:**
```typescript
function isCurrentlyTraversedNodeWithinModuleDeclaration(
  node: TSESTree.Node,
): boolean {
  return context.sourceCode
    .getAncestors(node)
    .some(
      node =>
        node.type === AST_NODE_TYPES.TSModuleDeclaration &&
        node.declare &&
        node.kind === 'global',
    );
}
```

**Go Implementation:**
```go
func isWithinDeclareGlobalModule(ctx rule.RuleContext, node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindModuleDeclaration {
			moduleDecl := current.AsModuleDeclaration()
			// Check if this is a global module declaration with declare modifier
			if moduleDecl.Name() != nil &&
			   ast.IsIdentifier(moduleDecl.Name()) &&
			   moduleDecl.Name().AsIdentifier().Text == "global" {
				// Check for declare modifier
				if moduleDecl.Modifiers() != nil {
					for _, modifier := range moduleDecl.Modifiers().Nodes {
						if modifier.Kind == ast.KindDeclareKeyword {
							return true
						}
					}
				}
			}
		}
		current = current.Parent
	}
	return false
}
```

**Issue:** The Go implementation checks for `moduleDecl.Name().AsIdentifier().Text == "global"` and looks for a declare modifier, but the TypeScript version checks `node.declare && node.kind === 'global'`. These are different properties - the TypeScript version checks if the module declaration itself has a declare flag and if its kind is 'global', while the Go version checks the name text and modifiers separately.

**Impact:** This affects test cases with `declare global {}` blocks where interfaces should not have fixes applied.

**Test Coverage:** Test cases with "declare global" should fail to provide fixes, but the Go version might incorrectly provide fixes.

#### 2. Export Declare Handling

**TypeScript Implementation:**
```typescript
// The TypeScript version properly handles export declare statements through ESLint's AST traversal
```

**Go Implementation:**
```go
// Missing explicit handling of export declare modifiers in type alias conversion
```

**Issue:** The Go implementation doesn't explicitly check for and preserve `declare` keywords when converting between types and interfaces, particularly in export declare scenarios.

**Impact:** Test cases like `export declare type Test = {...}` → `export declare interface Test {...}` may not preserve the `declare` keyword correctly.

**Test Coverage:** The test case with "export declare type" and "export declare interface" may not produce correct output.

#### 3. Export Default Newline Handling

**TypeScript Implementation:**
```typescript
if (node.parent.type === AST_NODE_TYPES.ExportDefaultDeclaration) {
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
fixes = append(fixes, rule.RuleFix{
	Text:  fmt.Sprintf("\nexport default %s", interfaceName),
	Range: core.TextRange{}.WithPos(insertPos).WithEnd(insertPos),
})
```

**Issue:** The Go implementation adds a newline before "export default" but the expected test output shows no leading newline in the export default test case.

**Impact:** The export default test case expects output without a leading newline before "export default".

**Test Coverage:** The export default interface test case will likely fail due to extra newline.

#### 4. Selector-based Filtering in TypeScript

**TypeScript Implementation:**
```typescript
"TSTypeAliasDeclaration[typeAnnotation.type='TSTypeLiteral']"(
  node: TSESTree.TSTypeAliasDeclaration,
): void {
```

**Go Implementation:**
```go
listeners[ast.KindTypeAliasDeclaration] = func(node *ast.Node) {
	typeAlias := node.AsTypeAliasDeclaration()
	// Check if the type is a type literal (object type), potentially wrapped in parentheses
	if typeAlias.Type != nil {
		actualType := unwrapParentheses(typeAlias.Type)
		if actualType != nil && actualType.Kind == ast.KindTypeLiteral {
```

**Issue:** The TypeScript version uses a CSS-like selector to only match type aliases with type literal annotations, while the Go version manually checks this condition. The logic appears equivalent but the approach differs.

**Impact:** Minimal - both should catch the same cases, but the manual checking in Go might be slightly less precise in edge cases.

**Test Coverage:** Should not affect test results significantly.

#### 5. Text Range Calculation Differences

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
	// Replace the equals and everything up to the type (including parentheses) with just a space,
	// then insert the type literal content
	fixes = append(fixes, rule.RuleFix{
		Text:  " " + getNodeText(ctx, actualTypeLiteral),
		Range: core.TextRange{}.WithPos(equalsStart).WithEnd(int(typeAlias.Type.End())),
	})
}
```

**Issue:** The TypeScript version carefully calculates token positions and handles the replacement in three separate fixes, while the Go version tries to do it in fewer operations. The Go approach might not handle whitespace and comments as precisely.

**Impact:** Could affect output formatting, particularly with comments between tokens.

**Test Coverage:** Test cases with comments like `type T /* comment */={ x: number; };` might not preserve formatting correctly.

### Recommendations
- Fix the declare global module detection to match TypeScript's logic (check `node.declare` and `node.kind` properties)
- Add explicit handling for `declare` keyword preservation in export statements
- Remove the leading newline in export default conversion to match expected test output
- Improve text range calculation to handle comments and whitespace more precisely
- Add comprehensive testing for edge cases around declare global scenarios
- Consider implementing token-based text manipulation similar to the TypeScript version for more accurate fixes

---