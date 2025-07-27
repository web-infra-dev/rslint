## Rule: no-unnecessary-type-parameters

### Test File: no-unnecessary-type-parameters.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic type parameter detection, simple usage counting, error message structure, support for function and class type parameters
- ⚠️ **POTENTIAL ISSUES**: Type-aware analysis implementation, scope management, AST traversal completeness, fix generation functionality
- ❌ **INCORRECT**: Missing sophisticated type analysis, incomplete scope-based reference tracking, simplified type parameter usage counting

### Discrepancies Found

#### 1. Type-Aware Analysis Implementation
**TypeScript Implementation:**
```typescript
function countTypeParameterUsage(
  checker: ts.TypeChecker,
  node: NodeWithTypeParameters,
): Map<ts.Identifier, number> {
  const counts = new Map<ts.Identifier, number>();
  // Comprehensive type analysis with TypeScript compiler APIs
  if (ts.isClassLike(node)) {
    for (const typeParameter of node.typeParameters) {
      collectTypeParameterUsageCounts(checker, typeParameter, counts, true);
    }
    for (const member of node.members) {
      collectTypeParameterUsageCounts(checker, member, counts, true);
    }
  } else {
    collectTypeParameterUsageCounts(checker, node, counts, false);
  }
}
```

**Go Implementation:**
```go
func countTypeParameterUsages(ctx rule.RuleContext, node *ast.Node, typeParamName string, typeParamNode *ast.Node) int {
  // Text-based approach to count meaningful occurrences
  nodeText := string(ctx.SourceFile.Text()[node.Pos():node.End()])
  count := 0
  // Simple string matching without deep type analysis
}
```

**Issue:** The Go implementation uses basic text-based counting instead of leveraging the TypeScript type checker for sophisticated type analysis. The TypeScript version performs deep type traversal including inferred types, mapped types, conditional types, and other complex type constructs.

**Impact:** May miss complex type parameter usages in inferred return types, mapped types, conditional types, and other advanced TypeScript constructs. Could result in false positives or negatives.

**Test Coverage:** Tests involving complex generic types, mapped types, conditional types, and inferred types may fail.

#### 2. Scope-Based Reference Tracking
**TypeScript Implementation:**
```typescript
function isTypeParameterRepeatedInAST(
  node: TSESTree.TSTypeParameter,
  references: Reference[],
  startOfBody = Infinity,
): boolean {
  // Uses ESLint scope manager for precise reference tracking
  for (const reference of references) {
    if (!reference.isTypeReference || 
        reference.identifier.name !== node.name.name) {
      continue;
    }
    // Sophisticated reference analysis
  }
}
```

**Go Implementation:**
```go
func isTypeParameterRepeatedInAST(typeParam *ast.Node, references []*ast.Node, startOfBody int) bool {
  // Simplified implementation without scope manager
  count := 0
  typeParamName := typeParam.AsTypeParameter().Name().Text()
  // Basic reference counting without scope analysis
}
```

**Issue:** The Go version lacks the sophisticated scope management system that the TypeScript version uses. It cannot distinguish between type references and value references, or properly track variable scoping.

**Impact:** May incorrectly count references that are in different scopes or incorrectly categorize value references as type references.

**Test Coverage:** Tests with shadowed type parameters or complex scoping scenarios may produce incorrect results.

#### 3. AST Node Type Coverage
**TypeScript Implementation:**
```typescript
return {
  [[
    'ArrowFunctionExpression[typeParameters]',
    'FunctionDeclaration[typeParameters]',
    'FunctionExpression[typeParameters]',
    'TSCallSignatureDeclaration[typeParameters]',
    'TSConstructorType[typeParameters]',
    'TSDeclareFunction[typeParameters]',
    'TSEmptyBodyFunctionExpression[typeParameters]',
    'TSFunctionType[typeParameters]',
    'TSMethodSignature[typeParameters]',
  ].join(', ')](node: TSESTree.FunctionLike): void {
    checkNode(node, 'function');
  },
```

**Go Implementation:**
```go
return rule.RuleListeners{
  ast.KindArrowFunction: func(node *ast.Node) {
    if node.TypeParameters() != nil {
      checkNode(node, "function")
    }
  },
  // Missing: ast.KindDeclareFunction, ast.KindTSEmptyBodyFunctionExpression
  // Commented out due to compilation issues
}
```

**Issue:** The Go implementation is missing support for some AST node types that the TypeScript version handles, particularly `DeclareFunction` and `TSEmptyBodyFunctionExpression`.

**Impact:** May miss type parameter violations in declare functions and certain function expression types.

**Test Coverage:** Tests with declare functions may not trigger the rule appropriately.

#### 4. Type Parameter Constraint Analysis
**TypeScript Implementation:**
```typescript
collectTypeParameterUsageCounts(checker, node, counts, false);

// Deep analysis of constraints
if (declaration.constraint && !visitedConstraints.has(declaration.constraint)) {
  visitedConstraints.add(declaration.constraint);
  visitType(checker.getTypeAtLocation(declaration.constraint), false);
}
```

**Go Implementation:**
```go
// Check if this type parameter has a constraint
if typeParamDecl.Constraint != nil {
  constraintText := string(ctx.SourceFile.Text()[typeParamDecl.Constraint.Pos():typeParamDecl.Constraint.End()])
  // If constraint involves another type parameter, this is meaningful
  for _, otherTypeParam := range node.TypeParameters() {
    // Simple text-based constraint analysis
  }
}
```

**Issue:** The Go version uses simple text-based analysis for constraints instead of leveraging type checker information to understand constraint relationships properly.

**Impact:** May incorrectly evaluate complex constraint relationships and miss sophisticated type parameter interdependencies.

**Test Coverage:** Tests with complex generic constraints may not be handled correctly.

#### 5. Fix Generation and Suggestions
**TypeScript Implementation:**
```typescript
suggest: [
  {
    messageId: 'replaceUsagesWithConstraint',
    *fix(fixer): Generator<TSESLint.RuleFix> {
      // Comprehensive fix generation with proper constraint handling
      const constraint = esTypeParameter.constraint;
      const constraintText = constraint != null &&
        constraint.type !== AST_NODE_TYPES.TSAnyKeyword
        ? context.sourceCode.getText(constraint)
        : 'unknown';
      
      // Replace all usages and remove type parameter
      for (const reference of smTypeParameterVariable.references) {
        // Complex fix logic with parentheses handling
      }
    },
  },
],
```

**Go Implementation:**
```go
// Report without suggestions to match test expectations
// Use the full type parameter position instead of just the name
ctx.ReportRange(
  core.NewTextRange(startPos, endPos),
  message,
)
```

**Issue:** The Go implementation completely omits fix generation and suggestions, which are a key feature of the TypeScript version.

**Impact:** Users won't get automatic fixes for type parameter issues, reducing the rule's usefulness.

**Test Coverage:** All test cases expecting suggestions will fail.

#### 6. Type Checker Integration
**TypeScript Implementation:**
```typescript
function visitType(type: ts.Type | undefined, assumeMultipleUses: boolean, isReturnType = false): void {
  // Deep integration with TypeScript type checker
  if (tsutils.isTypeParameter(type)) {
    // Handle type parameters with full type information
  } else if (type.aliasTypeArguments) {
    // Handle generic type aliases
  } else if (tsutils.isUnionOrIntersectionType(type)) {
    // Handle union/intersection types
  }
  // ... many more type-specific handlers
}
```

**Go Implementation:**
```go
func (c *typeParameterCounter) visitType(t *checker.Type, assumeMultipleUses, isReturnType bool) {
  if t == nil || c.incrementTypeUsage(t) > 9 {
    return
  }
  
  // Simplified type checking using available methods
  if c.checker.IsArrayLikeType(t) {
    // Handle array-like types
    return
  }
  // Basic type checking without full TypeScript utilities
}
```

**Issue:** The Go version lacks access to the comprehensive TypeScript type utilities (tsutils) and has a much simpler type analysis system.

**Impact:** Cannot perform the sophisticated type analysis that the TypeScript version relies on for accurate type parameter usage detection.

**Test Coverage:** Tests involving complex type structures may produce different results.

#### 7. Message ID and Data Handling
**TypeScript Implementation:**
```typescript
context.report({
  node: esTypeParameter,
  messageId: 'sole',
  data: {
    name: typeParameter.name.text,
    descriptor,
    uses: identifierCounts === 1 ? 'never used' : 'used only once',
  },
  suggest: [/* fix suggestions */]
});
```

**Go Implementation:**
```go
message := rule.RuleMessage{
  Id: "sole",
  Description: fmt.Sprintf("Type parameter %s is %s in the %s signature.", typeParamName, uses, descriptor),
}
```

**Issue:** The Go version embeds the data directly into the description instead of using a template-based message system with separate data fields.

**Impact:** Less flexible message formatting and potential inconsistencies in message presentation.

**Test Coverage:** Tests expecting specific data fields in error messages may fail.

### Recommendations
- Implement proper TypeScript type checker integration for sophisticated type analysis
- Add scope-based reference tracking system similar to ESLint's scope manager
- Complete AST node type coverage including declare functions and other missing types
- Implement fix generation and suggestion system
- Add support for complex type constructs (mapped types, conditional types, etc.)
- Enhance constraint analysis with proper type checking
- Implement template-based message system with data interpolation
- Add comprehensive type parameter usage detection for inferred types
- Include support for type predicates and advanced TypeScript features
- Test against all original TypeScript-ESLint test cases to ensure compatibility

---