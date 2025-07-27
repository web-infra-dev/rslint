# Validation Report: class-literal-property-style

## Rule: class-literal-property-style

### Test File: class-literal-property-style.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic rule structure and listener pattern
  - Option parsing for 'fields' vs 'getters' modes
  - Message IDs and descriptions match exactly
  - Supported literal detection covers most cases
  - Basic getter-to-field and field-to-getter conversion logic
  - Override modifier handling
  - Static modifier handling
  - Accessibility modifier handling
  - Computed property name support
  - Constructor assignment exclusion logic
  - Setter duplicate detection

- ⚠️ **POTENTIAL ISSUES**: 
  - Template literal handling may differ slightly between implementations
  - AST traversal patterns might have subtle differences
  - Error positioning/reporting nodes may not be identical

- ❌ **INCORRECT**: 
  - Missing support for method declarations in AST pattern matching
  - Template expression validation logic is incomplete
  - Nested class constructor exclusion logic may be flawed

### Discrepancies Found

#### 1. Method Declaration Support Missing

**TypeScript Implementation:**
```typescript
interface NodeWithModifiers {
  accessibility?: TSESTree.Accessibility;
  static: boolean;
}

const printNodeModifiers = (
  node: NodeWithModifiers,
  final: 'get' | 'readonly',
): string =>
  `${node.accessibility ?? ''}${
    node.static ? ' static' : ''
  } ${final} `.trimStart();
```

**Go Implementation:**
```go
func getStaticMemberAccessValue(ctx rule.RuleContext, node *ast.Node) string {
	var nameNode *ast.Node

	if ast.IsPropertyDeclaration(node) {
		nameNode = node.AsPropertyDeclaration().Name()
	} else if ast.IsMethodDeclaration(node) {
		nameNode = node.AsMethodDeclaration().Name()
	} else if ast.IsGetAccessorDeclaration(node) {
		nameNode = node.AsGetAccessorDeclaration().Name()
	} else if ast.IsSetAccessorDeclaration(node) {
		nameNode = node.AsSetAccessorDeclaration().Name()
	} else {
		return ""
	}
}
```

**Issue:** The Go implementation includes method declarations in `getStaticMemberAccessValue` but the TypeScript version focuses on class members with specific patterns. This might lead to different behavior.

**Impact:** Could cause the rule to incorrectly process method declarations that shouldn't be considered for this rule.

**Test Coverage:** This might affect test cases with regular methods vs getters.

#### 2. Template Expression Validation Logic

**TypeScript Implementation:**
```typescript
const isSupportedLiteral = (
  node: TSESTree.Node,
): node is TSESTree.LiteralExpression => {
  switch (node.type) {
    case AST_NODE_TYPES.TaggedTemplateExpression:
      return node.quasi.quasis.length === 1;

    case AST_NODE_TYPES.TemplateLiteral:
      return node.quasis.length === 1;
  }
};
```

**Go Implementation:**
```go
case ast.KindTemplateExpression:
	// Only support template literals with no interpolation
	template := node.AsTemplateExpression()
	return template != nil && len(template.TemplateSpans.Nodes) == 0
case ast.KindTaggedTemplateExpression:
	// Support tagged template expressions only with no interpolation
	tagged := node.AsTaggedTemplateExpression()
	if tagged.Template.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return true
	}
	if tagged.Template.Kind == ast.KindTemplateExpression {
		template := tagged.Template.AsTemplateExpression()
		return template != nil && len(template.TemplateSpans.Nodes) == 0
	}
	return false
```

**Issue:** The validation logic for template expressions differs. TypeScript checks `quasis.length === 1` while Go checks `len(template.TemplateSpans.Nodes) == 0`. These are checking different aspects of template literals.

**Impact:** May incorrectly accept or reject template literals with different interpolation patterns.

**Test Coverage:** Test cases with template literals and tagged template expressions may behave differently.

#### 3. Constructor Assignment Exclusion Logic

**TypeScript Implementation:**
```typescript
'MethodDefinition[kind="constructor"] ThisExpression'(
  node: TSESTree.ThisExpression,
): void {
  if (node.parent.type === AST_NODE_TYPES.MemberExpression) {
    let parent: TSESTree.Node | undefined = node.parent;

    while (!isFunction(parent)) {
      parent = parent.parent;
    }

    if (
      parent.parent.type === AST_NODE_TYPES.MethodDefinition &&
      parent.parent.kind === 'constructor'
    ) {
      excludeAssignedProperty(node.parent);
    }
  }
}
```

**Go Implementation:**
```go
listeners[ast.KindThisKeyword] = func(node *ast.Node) {
	// Check if this is inside a member expression (this.property or this['property'])
	if node.Parent == nil || (!ast.IsPropertyAccessExpression(node.Parent) && !ast.IsElementAccessExpression(node.Parent)) {
		return
	}

	// Walk up to find the containing function
	parent := memberExpr.Parent
	for parent != nil && !isFunction(parent) {
		parent = parent.Parent
	}

	// Check if this function is a constructor by checking its parent
	if parent != nil && parent.Parent != nil {
		if ast.IsMethodDeclaration(parent.Parent) {
			method := parent.Parent.AsMethodDeclaration()
			if method.Kind == ast.KindConstructorKeyword {
				// We're in a constructor - exclude this property
				if len(propertiesInfoStack) > 0 {
					info := propertiesInfoStack[len(propertiesInfoStack)-1]
					info.excludeSet[propName] = true
				}
			}
		} else if ast.IsConstructorDeclaration(parent.Parent) {
			// Direct constructor declaration
			if len(propertiesInfoStack) > 0 {
				info := propertiesInfoStack[len(propertiesInfoStack)-1]
				info.excludeSet[propName] = true
			}
		}
	}
}
```

**Issue:** The Go implementation has a more complex logic for detecting constructor assignments, but it doesn't match the precise TypeScript pattern. The TypeScript version uses a selector pattern while Go manually traverses the AST.

**Impact:** May incorrectly exclude or include properties that are assigned in constructors, especially in nested class scenarios.

**Test Coverage:** Test cases with nested classes and constructor assignments may fail.

#### 4. Property Name Extraction for String Literals

**TypeScript Implementation:**
```typescript
const name = getStaticMemberAccessValue(node, context);
```

**Go Implementation:**
```go
func extractPropertyName(ctx rule.RuleContext, nameNode *ast.Node) string {
	// Handle string literals as property names
	if ast.IsLiteralExpression(nameNode) {
		text := nameNode.Text()
		// Remove quotes for string literals to normalize the name
		if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
			return text[1 : len(text)-1]
		}
		return text
	}
}
```

**Issue:** The Go implementation manually handles quote removal for string literals, but the TypeScript implementation may handle this differently through the ESLint utilities.

**Impact:** Could cause mismatches in property name comparison when dealing with quoted property names.

**Test Coverage:** Test cases with quoted property names like `['foo']` vs `foo` may behave inconsistently.

#### 5. Class Body Listener Pattern

**TypeScript Implementation:**
```typescript
return {
  ...(style === 'getters' && {
    ClassBody: enterClassBody,
    'ClassBody:exit': exitClassBody,
  }),
};
```

**Go Implementation:**
```go
listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
	enterClassBody()
}
listeners[rule.ListenerOnExit(ast.KindClassDeclaration)] = func(node *ast.Node) {
	exitClassBody()
}
listeners[ast.KindClassExpression] = func(node *ast.Node) {
	enterClassBody()
}
listeners[rule.ListenerOnExit(ast.KindClassExpression)] = func(node *ast.Node) {
	exitClassBody()
}
```

**Issue:** TypeScript uses `ClassBody` as the node type while Go uses `ClassDeclaration` and `ClassExpression`. This is a fundamental difference in AST structure.

**Impact:** The scope tracking for properties might be incorrect if the AST traversal doesn't match exactly.

**Test Coverage:** All test cases with the 'getters' option could be affected.

### Recommendations
- Verify template literal validation logic matches exactly between implementations
- Review constructor assignment exclusion to ensure nested class scenarios work correctly
- Ensure property name normalization is consistent between quoted and unquoted names
- Validate that class body traversal correctly tracks property scope
- Add specific test cases for edge cases around template literals and constructor assignments
- Review AST node type mappings between TypeScript ESTree and Go typescript-go
- Consider adding debug logging to compare behavior on specific test cases

---