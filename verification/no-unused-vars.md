## Rule: no-unused-vars

### Test File: no-unused-vars.test.ts

### Validation Summary
- ⚠️ **POTENTIAL ISSUES**: Basic structure in place, some configuration handling
- ❌ **INCORRECT**: 
  - Missing core scope analysis functionality
  - No ambient declaration handling
  - Incomplete AST traversal
  - Missing export statement analysis
  - No type-only usage detection
  - Simplified variable collection logic
  - Missing complex visitor patterns
  - No proper scope chain tracking

### Discrepancies Found

#### 1. Missing Scope Analysis Engine
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

**Issue:** The Go implementation lacks the sophisticated scope analysis that the TypeScript version relies on. The TypeScript version uses `collectVariables(context)` which leverages ESLint's scope-manager to properly track variable definitions, references, and scopes.

**Impact:** This will cause the rule to miss many unused variables and incorrectly flag used variables as unused.

**Test Coverage:** All test cases will likely fail due to incorrect variable analysis.

#### 2. Missing Ambient Declaration Handling
**TypeScript Implementation:**
```typescript
// top-level declaration file handling
[ambientDeclarationSelector(AST_NODE_TYPES.Program)](
  node: DeclarationSelectorNode,
): void {
  if (!isDefinitionFile(context.filename)) {
    return;
  }
  const moduleDecl = nullThrows(
    node.parent,
    NullThrowsReasons.MissingParent,
  ) as TSESTree.Program;
  if (checkForOverridingExportStatements(moduleDecl)) {
    return;
  }
  markDeclarationChildAsUsed(node);
},
```

**Go Implementation:**
```go
// No equivalent handling for ambient declarations
```

**Issue:** The Go version completely lacks handling for ambient declarations in TypeScript definition files (.d.ts), which should be automatically marked as used.

**Impact:** Definition files will incorrectly report unused variables for ambient declarations.

**Test Coverage:** Test cases with `declare` statements and `.d.ts` files will fail.

#### 3. Missing Type-Only Usage Detection
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

**Issue:** The Go implementation has a basic type-only detection but lacks the sophisticated logic for handling type imports and the `referenceContainsTypeQuery` functionality.

**Impact:** Type-only imports may be incorrectly flagged as unused, and variables used only in type positions may not be properly detected.

**Test Coverage:** Tests with type-only imports and `typeof` usage will fail.

#### 4. Incomplete AST Traversal
**TypeScript Implementation:**
```typescript
function visitPattern(
  node: TSESTree.Node,
  cb: (node: TSESTree.Identifier) => void,
): void {
  const visitor = new PatternVisitor({}, node, cb);
  visitor.visit(node);
}
```

**Go Implementation:**
```go
func collectVariableInfo(ctx rule.RuleContext, node *ast.Node, variables map[*ast.Node]*VariableInfo) {
	// Handle variable declarations
	switch node.Kind {
	case ast.KindVariableStatement:
		// Limited handling
	}
	// Recursively process children - simplified traversal
	// In a real implementation, we would need proper AST traversal
}
```

**Issue:** The Go implementation has incomplete AST traversal that doesn't properly walk all node types or handle complex patterns like destructuring.

**Impact:** Many variable declarations and references will be missed, leading to incorrect results.

**Test Coverage:** Tests with destructuring patterns, nested scopes, and complex expressions will fail.

#### 5. Missing Export Statement Analysis
**TypeScript Implementation:**
```typescript
function hasOverridingExportStatement(
  body: TSESTree.ProgramStatement[],
): boolean {
  for (const statement of body) {
    if (
      (statement.type === AST_NODE_TYPES.ExportNamedDeclaration &&
        statement.declaration == null) ||
      statement.type === AST_NODE_TYPES.ExportAllDeclaration ||
      statement.type === AST_NODE_TYPES.TSExportAssignment
    ) {
      return true;
    }
    // ... more export checks
  }
  return false;
}
```

**Go Implementation:**
```go
// No equivalent export statement analysis
```

**Issue:** The Go version doesn't check for export statements that would make variables used implicitly.

**Impact:** Exported variables may be incorrectly flagged as unused.

**Test Coverage:** Tests with export statements will fail.

#### 6. Simplified Reference Analysis
**TypeScript Implementation:**
```typescript
const writeReferences = unusedVar.references.filter(
  ref =>
    ref.isWrite() &&
    ref.from.variableScope === unusedVar.scope.variableScope,
);
```

**Go Implementation:**
```go
func isWriteReference(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	// Check if this is a write reference
	if ast.IsBinaryExpression(parent) {
		binExpr := parent.AsBinaryExpression()
		if binExpr.Left == node && isAssignmentOperator(binExpr.OperatorToken.Kind) {
			return true
		}
	}
	return false
}
```

**Issue:** The Go implementation has overly simplified write reference detection that doesn't handle all assignment contexts or scope considerations.

**Impact:** The distinction between "defined" and "assigned a value" in error messages will be incorrect.

**Test Coverage:** Tests expecting different messages for assigned vs defined variables will fail.

#### 7. Missing Complex Option Handling
**TypeScript Implementation:**
```typescript
function isAfterLastUsedArg(variable: ScopeVariable): boolean {
  const def = variable.defs[0];
  const params = context.sourceCode.getDeclaredVariables(def.node);
  const posteriorParams = params.slice(params.indexOf(variable) + 1);
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

**Issue:** The Go implementation doesn't properly implement the "after-used" parameter logic.

**Impact:** The `args: "after-used"` option will not work correctly.

**Test Coverage:** Tests with `args: "after-used"` option will fail.

#### 8. Missing Rest Sibling Analysis
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

**Issue:** The Go implementation doesn't properly detect rest siblings in destructuring patterns.

**Impact:** The `ignoreRestSiblings` option will not work correctly.

**Test Coverage:** Tests with rest sibling destructuring will fail.

### Recommendations
- **Critical**: Implement proper scope analysis equivalent to ESLint's scope-manager
- **Critical**: Add comprehensive AST traversal to collect all variable declarations and references
- **Critical**: Implement ambient declaration handling for TypeScript definition files
- **High**: Add sophisticated type-only usage detection using TypeScript's type checker
- **High**: Implement export statement analysis to mark exported variables as used
- **High**: Add proper "after-used" parameter analysis
- **Medium**: Implement rest sibling detection for destructuring patterns
- **Medium**: Add comprehensive write reference detection
- **Low**: Enhance error message generation to match TypeScript implementation exactly

The Go implementation needs a fundamental rewrite to match the TypeScript version's functionality. The current implementation is too simplified and will not pass the majority of test cases.

---