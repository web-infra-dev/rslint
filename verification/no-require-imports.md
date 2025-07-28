## Rule: no-require-imports

### Test File: no-require-imports.test.ts

### Validation Summary
- ✅ **CORRECT**: Core rule logic structure, TSExternalModuleReference handling, configuration options parsing, basic string literal processing
- ⚠️ **POTENTIAL ISSUES**: Optional chaining syntax handling, scope resolution complexity, template literal processing edge cases
- ❌ **INCORRECT**: Global require detection logic has fundamental differences that could cause false positives/negatives

### Discrepancies Found

#### 1. Global Require Detection Logic Complexity
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
	sourceFile := ctx.SourceFile
	if sourceFile == nil {
		return true
	}
	
	// Check if 'require' is defined anywhere in the current scope context
	return !isRequireLocallyDefined(sourceFile, node)
}

// Complex multi-function implementation with scope traversal
```

**Issue:** The Go implementation uses a complex custom scope traversal mechanism instead of leveraging TypeScript's built-in scope analysis. The TypeScript version uses ESLint's `findVariable` utility which properly handles JavaScript/TypeScript scoping rules, while the Go version manually walks the AST which may miss edge cases.

**Impact:** This could lead to incorrect behavior in complex scoping scenarios, particularly with nested functions, closures, and different types of declarations.

**Test Coverage:** Test cases like the one with `createRequire` from 'module' and local require reassignment scenarios rely heavily on this logic.

#### 2. Optional Chaining Implementation Inconsistency
**TypeScript Implementation:**
```typescript
'CallExpression[callee.name="require"]'(
  node: TSESTree.CallExpression,
): void {
  // Automatically matches both require() and require?.() calls
}
```

**Go Implementation:**
```go
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

**Issue:** The TypeScript selector pattern automatically handles both `require()` and `require?.()` calls, while the Go implementation has separate logic paths. The Go logic for optional chaining appears to check the same condition twice (identifier.Text == "require") which may be redundant.

**Impact:** Potential for different handling of optional chaining edge cases.

**Test Coverage:** Tests with `require?.()` syntax rely on this logic.

#### 3. Template Literal Processing Differences
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

**Issue:** The TypeScript version accepts any `TemplateLiteral` node, while the Go version has additional conditions checking for `TemplateSpans == nil`. This could cause discrepancies in how complex template literals are handled.

**Impact:** May affect validation of require calls with template literal arguments.

**Test Coverage:** Tests with template literal syntax like `require(\`./package.json\`)` depend on this.

#### 4. Scope Resolution Accuracy
**TypeScript Implementation:**
```typescript
// Uses ESLint's built-in scope analysis
const variable = ASTUtils.findVariable(
  context.sourceCode.getScope(node),
  'require',
);
```

**Go Implementation:**
```go
// Custom implementation traversing parent nodes and checking statements
func isRequireLocallyDefined(sourceFile *ast.SourceFile, callNode *ast.Node) bool {
	currentNode := callNode
	
	for currentNode != nil {
		if hasLocalRequireInContext(currentNode) {
			return true
		}
		currentNode = currentNode.Parent
	}
	
	return hasLocalRequireInStatements(sourceFile.Statements)
}
```

**Issue:** The Go implementation may not correctly handle all JavaScript/TypeScript scoping rules, particularly for function scopes, block scopes, and complex nested structures. The manual traversal might miss declarations in sibling scopes or misunderstand scope boundaries.

**Impact:** Critical for test cases where `require` is locally defined but should not trigger the rule.

**Test Coverage:** Several test cases depend on this, especially those with local `require` declarations.

### Recommendations
- **HIGH PRIORITY**: Review and test the global require detection logic extensively against the TypeScript test cases, particularly the ones involving local require declarations
- **MEDIUM PRIORITY**: Simplify the optional chaining logic to match the TypeScript behavior more closely
- **MEDIUM PRIORITY**: Verify template literal handling matches ESLint's `getStaticStringValue` utility behavior
- **LOW PRIORITY**: Consider leveraging TypeScript-Go's scope analysis if available instead of manual traversal
- **TEST ENHANCEMENT**: Add more edge case tests for complex scoping scenarios, nested functions, and different declaration types

---