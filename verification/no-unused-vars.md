## Rule: no-unused-vars

### Test File: no-unused-vars.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic option parsing structure, message ID definitions, regex pattern compilation
- ⚠️ **POTENTIAL ISSUES**: Incomplete AST traversal, simplified scope analysis, missing type-aware features
- ❌ **INCORRECT**: Missing scope manager integration, incomplete variable collection, missing ambient declaration handling

### Discrepancies Found

#### 1. Missing Scope Manager Integration
**TypeScript Implementation:**
```typescript
const analysisResults = collectVariables(context);
const variables = [
  ...Array.from(analysisResults.unusedVariables, variable => ({
    used: false,
    variable,
  })),
  ...Array.from(analysisResults.usedVariables, variable => ({
    used: true,
    variable,
  })),
];
```

**Go Implementation:**
```go
func collectVariables(ctx rule.RuleContext, sourceFile *ast.Node) map[*ast.Node]*VariableInfo {
	variables := make(map[*ast.Node]*VariableInfo)
	// This is a simplified version - in a real implementation, we would need to:
	// 1. Walk the entire AST
	// 2. Track all variable declarations
	// 3. Track all variable references
	// 4. Determine if references are type-only
	// 5. Handle scope correctly
	collectVariableInfo(ctx, sourceFile, variables)
	return variables
}
```

**Issue:** The Go implementation lacks proper scope analysis using a scope manager. The TypeScript version uses `collectVariables(context)` which provides comprehensive scope analysis, while the Go version has a placeholder with manual AST traversal.

**Impact:** This will cause incorrect reporting of unused variables because scope boundaries, variable shadowing, and reference resolution won't work properly.

**Test Coverage:** Most test cases will fail due to incorrect variable collection and scope analysis.

#### 2. Incomplete AST Traversal
**TypeScript Implementation:**
```typescript
return {
  // Multiple AST selectors for ambient declarations
  [ambientDeclarationSelector(AST_NODE_TYPES.Program)](node: DeclarationSelectorNode): void {
    // Ambient declaration handling
  },
  [ambientDeclarationSelector('TSModuleDeclaration[declare = true] > TSModuleBlock')](node: DeclarationSelectorNode): void {
    // Namespace handling
  },
  'Program:exit'(programNode): void {
    const unusedVars = collectUnusedVariables();
    // Process all variables at program exit
  },
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
	ast.KindSourceFile: func(node *ast.Node) {
		// Process all variables at the end of the file
		collectAndReportUnusedVariables(ctx, opts, node)
	},
}
```

**Issue:** The Go implementation only listens to `KindSourceFile` and doesn't handle ambient declarations, module declarations, or use the proper program exit pattern.

**Impact:** Ambient declarations, TypeScript modules, and definition files won't be handled correctly.

**Test Coverage:** Tests involving ambient declarations and module scoping will fail.

#### 3. Missing Type-Only Reference Detection
**TypeScript Implementation:**
```typescript
const usedOnlyAsType = unusedVar.references.some(ref =>
  referenceContainsTypeQuery(ref.identifier),
);

const isImportUsedOnlyAsType =
  usedOnlyAsType &&
  unusedVar.defs.some(
    def => def.type === DefinitionType.ImportBinding,
  );
if (isImportUsedOnlyAsType) {
  continue;
}
```

**Go Implementation:**
```go
func isTypeOnlyUsage(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Check if the identifier is used in a type context
	switch parent.Kind {
	case ast.KindTypeReference:
		return true
	case ast.KindTypeQuery:
		return true
	case ast.KindQualifiedName:
		return isTypeOnlyUsage(parent)
	}

	return false
}
```

**Issue:** The Go implementation has a simplified type-only detection that doesn't use the sophisticated `referenceContainsTypeQuery` function and doesn't properly handle import bindings.

**Impact:** Type-only imports and references won't be handled correctly, causing false positives for unused variables that are only used in type positions.

**Test Coverage:** Tests for type-only imports and `typeof` references will fail.

#### 4. Missing Parameter Position Analysis
**TypeScript Implementation:**
```typescript
function isAfterLastUsedArg(variable: ScopeVariable): boolean {
  const def = variable.defs[0];
  const params = context.sourceCode.getDeclaredVariables(def.node);
  const posteriorParams = params.slice(params.indexOf(variable) + 1);

  // If any used parameters occur after this parameter, do not report.
  return !posteriorParams.some(
    v => v.references.length > 0 || v.eslintUsed,
  );
}
```

**Go Implementation:**
```go
func isAfterLastUsedParam(ctx rule.RuleContext, varInfo *VariableInfo) bool {
	// Check if this parameter comes after the last used parameter
	// This requires analyzing all parameters in the function
	return true // Simplified for now
}
```

**Issue:** The Go implementation doesn't properly implement the "after-used" parameter logic, which requires analyzing parameter order and usage.

**Impact:** The `args: "after-used"` option won't work correctly.

**Test Coverage:** Tests with `args: "after-used"` will fail.

#### 5. Missing Rest Sibling Detection
**TypeScript Implementation:**
```typescript
function hasRestSibling(node: TSESTree.Node): boolean {
  return (
    node.type === AST_NODE_TYPES.Property &&
    node.parent.type === AST_NODE_TYPES.ObjectPattern &&
    node.parent.properties[node.parent.properties.length - 1].type ===
      AST_NODE_TYPES.RestElement
  );
}

function hasRestSpreadSibling(variable: ScopeVariable): boolean {
  if (options.ignoreRestSiblings) {
    const hasRestSiblingDefinition = variable.defs.some(def =>
      hasRestSibling(def.name.parent),
    );
    const hasRestSiblingReference = variable.references.some(ref =>
      hasRestSibling(ref.identifier.parent),
    );

    return hasRestSiblingDefinition || hasRestSiblingReference;
  }

  return false;
}
```

**Go Implementation:**
```go
func hasRestSibling(varInfo *VariableInfo) bool {
	// Check if the variable has a rest sibling in object destructuring
	if varInfo.Definition == nil {
		return false
	}
	
	parent := varInfo.Definition.Parent
	if parent != nil && parent.Kind == ast.KindObjectBindingPattern {
		// Check if there's a rest element in the pattern
		// This is simplified - would need proper implementation
		return false
	}
	
	return false
}
```

**Issue:** The Go implementation doesn't properly detect rest siblings in destructuring patterns and doesn't check both definitions and references.

**Impact:** The `ignoreRestSiblings` option won't work correctly.

**Test Coverage:** Tests with rest sibling destructuring will fail.

#### 6. Missing Static Init Block Detection
**TypeScript Implementation:**
```typescript
if (def.type === TSESLint.Scope.DefinitionType.ClassName) {
  const hasStaticBlock = def.node.body.body.some(
    node => node.type === AST_NODE_TYPES.StaticBlock,
  );

  if (options.ignoreClassWithStaticInitBlock && hasStaticBlock) {
    continue;
  }
}
```

**Go Implementation:**
```go
func isClassWithStaticInitBlock(definition *ast.Node) bool {
	if definition == nil || definition.Kind != ast.KindClassDeclaration {
		return false
	}
	
	classDecl := definition.AsClassDeclaration()
	if classDecl.Members != nil {
		for _, member := range classDecl.Members.Nodes {
			// Check for static blocks - using a different approach since KindStaticBlock may not be available
			if member.Kind == ast.KindMethodDeclaration {
				// This is a simplified check
				return false
			}
		}
	}
	
	return false
}
```

**Issue:** The Go implementation doesn't properly detect static init blocks (`static {}`) in classes.

**Impact:** The `ignoreClassWithStaticInitBlock` option won't work correctly.

**Test Coverage:** Tests with static init blocks will fail.

#### 7. Missing Global Directive Comment Handling
**TypeScript Implementation:**
```typescript
// If there are no regular declaration, report the first `/*globals*/` comment directive.
} else if (
  'eslintExplicitGlobalComments' in unusedVar &&
  unusedVar.eslintExplicitGlobalComments
) {
  const directiveComment = unusedVar.eslintExplicitGlobalComments[0];

  context.report({
    loc: getNameLocationInGlobalDirectiveComment(
      context.sourceCode,
      directiveComment,
      unusedVar.name,
    ),
    node: programNode,
    messageId: 'unusedVar',
    data: getDefinedMessageData(unusedVar),
  });
}
```

**Go Implementation:**
```go
// No equivalent handling for global directive comments
```

**Issue:** The Go implementation doesn't handle variables declared via `/* global */` comments.

**Impact:** Global directive comments won't be processed for unused variable detection.

**Test Coverage:** Tests with global directive comments will fail.

#### 8. Incomplete Variable Type Detection
**TypeScript Implementation:**
```typescript
function defToVariableType(def: Definition): VariableType {
  if (
    options.destructuredArrayIgnorePattern &&
    def.name.parent.type === AST_NODE_TYPES.ArrayPattern
  ) {
    return 'array-destructure';
  }

  switch (def.type) {
    case DefinitionType.CatchClause:
      return 'catch-clause';
    case DefinitionType.Parameter:
      return 'parameter';
    default:
      return 'variable';
  }
}
```

**Go Implementation:**
```go
func getVariableType(definition *ast.Node) VariableType {
	if definition == nil {
		return VariableTypeVariable
	}

	parent := definition.Parent
	if parent == nil {
		return VariableTypeVariable
	}

	// Check for array destructuring
	if parent.Kind == ast.KindArrayBindingPattern {
		return VariableTypeArrayDestructure
	}

	// Check for catch clause
	if parent.Kind == ast.KindCatchClause {
		return VariableTypeCatchClause
	}

	// Check for parameter
	if definition.Kind == ast.KindParameter {
		return VariableTypeParameter
	}

	return VariableTypeVariable
}
```

**Issue:** The Go implementation doesn't consider the destructured array ignore pattern option when determining variable type, and the logic for detecting destructuring patterns may be incomplete.

**Impact:** Array destructuring pattern detection won't align with the ignore pattern logic.

**Test Coverage:** Tests with destructured array patterns will have inconsistent behavior.

### Recommendations
- Implement proper scope manager integration to track variable declarations and references across scopes
- Add comprehensive AST traversal to handle all node types and patterns
- Implement sophisticated type-only reference detection using TypeScript type information
- Add proper parameter position analysis for "after-used" args option
- Implement complete rest sibling detection for object/array destructuring
- Add static init block detection for classes
- Handle global directive comments (/* global */ declarations)
- Ensure variable type detection aligns with all configuration options
- Add comprehensive test coverage for all edge cases mentioned in the TypeScript implementation comments
- Consider implementing the ambient declaration selectors for TypeScript module handling
- Add proper write reference detection for assignment vs definition reporting

---