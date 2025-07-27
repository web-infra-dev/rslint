## Rule: no-require-imports

### Test File: no-require-imports.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic require() call detection, external module reference handling, allow patterns with regex, allowAsImport option for import equals declarations, string and template literal argument handling
- ⚠️ **POTENTIAL ISSUES**: Global require detection logic complexity, optional chaining handling approach, template literal with expressions edge case
- ❌ **INCORRECT**: None identified that would cause test failures

### Discrepancies Found

#### 1. Global Require Detection Approach
**TypeScript Implementation:**
```typescript
const variable = ASTUtils.findVariable(
  context.sourceCode.getScope(node),
  'require',
);
// ignore non-global require usage as it's something user-land custom instead
// of the commonjs standard
if (!variable?.identifiers.length) {
  context.report({
    node,
    messageId: 'noRequireImports',
  });
}
```

**Go Implementation:**
```go
// isGlobalRequire checks if the require is the global require function
func isGlobalRequire(ctx rule.RuleContext, node *ast.Node) bool {
	// Walk up to find the source file and traverse all variable declarations
	sourceFile := ctx.SourceFile
	if sourceFile == nil {
		return true
	}
	
	// Check if 'require' is defined anywhere in the current scope context
	return !isRequireLocallyDefined(sourceFile, node)
}
```

**Issue:** The Go implementation uses a custom scope traversal algorithm that manually walks through AST nodes to find local require declarations, while TypeScript uses the ESLint scope analysis utilities. The Go approach may miss complex scoping scenarios or block-scoped declarations.

**Impact:** Could potentially miss cases where require is redefined in nested scopes or through more complex declaration patterns.

**Test Coverage:** Tests with local require redefinition should reveal any issues, particularly the cases with `let require = bazz;`.

#### 2. Optional Chaining Detection
**TypeScript Implementation:**
```typescript
'CallExpression[callee.name="require"]'(
  node: TSESTree.CallExpression,
): void {
  // Handles both require() and require?.() through selector matching
}
```

**Go Implementation:**
```go
// Check if this is a require call or require?.() call
var isRequireCall bool

if ast.IsIdentifier(callExpr.Expression) {
	identifier := callExpr.Expression.AsIdentifier()
	if identifier.Text == "require" {
		isRequireCall = true
	}
} else if callExpr.QuestionDotToken != nil {
	// Handle optional chaining: require?.()
	// The expression should be require for require?.()
	if ast.IsIdentifier(callExpr.Expression) {
		identifier := callExpr.Expression.AsIdentifier()
		if identifier.Text == "require" {
			isRequireCall = true
		}
	}
}
```

**Issue:** The Go implementation explicitly checks for optional chaining tokens, which should work correctly but uses a different detection strategy than the TypeScript selector-based approach.

**Impact:** Should function equivalently for the test cases, but the logic is more explicit in Go.

**Test Coverage:** Test cases with `require?.()` should validate this works correctly.

#### 3. Template Literal Handling
**TypeScript Implementation:**
```typescript
function isStringOrTemplateLiteral(node: TSESTree.Node): boolean {
  return (
    (node.type === AST_NODE_TYPES.Literal &&
      typeof node.value === 'string') ||
    node.type === AST_NODE_TYPES.TemplateLiteral
  );
}
```

**Go Implementation:**
```go
func isStringOrTemplateLiteral(node *ast.Node) bool {
	return (node.Kind == ast.KindStringLiteral) ||
		(node.Kind == ast.KindTemplateExpression && node.AsTemplateExpression().TemplateSpans == nil) ||
		(node.Kind == ast.KindNoSubstitutionTemplateLiteral)
}
```

**Issue:** The Go implementation has more specific handling for template expressions, checking for empty TemplateSpans, while TypeScript accepts all template literals. However, the getStaticStringValue function in Go only extracts values from simple templates without expressions, which aligns with the intent.

**Impact:** Go implementation may be more restrictive but correctly handles the test cases.

**Test Coverage:** Template literal test cases like `require(\`./package.json\`)` should validate this.

#### 4. Message ID Consistency
**TypeScript Implementation:**
```typescript
messages: {
  noRequireImports: 'A `require()` style import is forbidden.',
},
```

**Go Implementation:**
```go
ctx.ReportNode(node, rule.RuleMessage{
	Id:          "noRequireImports",
	Description: "A `require()` style import is forbidden.",
})
```

**Issue:** Both use the same message ID and description, ensuring consistency.

**Impact:** No impact - correctly implemented.

**Test Coverage:** All test cases expecting `messageId: 'noRequireImports'` should pass.

### Recommendations
- **Verify scope detection**: The manual scope traversal in the Go implementation should be tested against complex scoping scenarios to ensure it matches ESLint's scope analysis behavior
- **Validate optional chaining**: Ensure the explicit optional chaining detection works for all variations of `require?.()` calls
- **Test template literal edge cases**: Verify that template literals with expressions are handled consistently between implementations
- **Consider performance**: The Go implementation's scope traversal may be less efficient than using dedicated scope analysis utilities

---