## Rule: no-unnecessary-type-parameters

### Test File: no-unnecessary-type-parameters.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic AST node type handling, message structure, rule registration pattern
- ⚠️ **POTENTIAL ISSUES**: Scope/reference tracking implementation, type parameter usage counting methodology, type-aware analysis depth
- ❌ **INCORRECT**: Complete absence of scope manager functionality, oversimplified type parameter usage counting, missing complex type analysis

### Discrepancies Found

#### 1. **Missing Scope Manager Integration**
**TypeScript Implementation:**
```typescript
// Get the scope in which the type parameters are declared.
const scope = context.sourceCode.getScope(node);

const smTypeParameterVariable = nullThrows(
  (() => {
    const variable = scope.set.get(esTypeParameter.name.name);
    return variable?.isTypeVariable ? variable : undefined;
  })(),
  "Type parameter should be present in scope's variables.",
);

// Quick path: if the type parameter is used multiple times in the AST,
// we don't need to dip into types to know it's repeated.
if (
  isTypeParameterRepeatedInAST(
    esTypeParameter,
    smTypeParameterVariable.references,
    node.body?.range[0] ?? node.returnType?.range[1],
  )
) {
  continue;
}
```

**Go Implementation:**
```go
// Scope functionality not available - simplified implementation
counter := newTypeParameterCounter(checker, descriptor == "class")

// Count type parameter usages
usageCount := countTypeParameterUsages(ctx, node, typeParamName, typeParam)
```

**Issue:** The Go implementation completely lacks scope manager functionality, which is crucial for accurately tracking type parameter references and distinguishing between different scopes.

**Impact:** Cannot properly detect type parameter shadowing, cannot track references across scopes, may produce false positives/negatives.

**Test Coverage:** All test cases that involve nested functions or complex scoping scenarios.

#### 2. **Oversimplified Type Parameter Usage Counting**
**TypeScript Implementation:**
```typescript
// For any inferred types, we have to dip into type checking.
counts ??= countTypeParameterUsage(checker, tsNode);
const identifierCounts = counts.get(typeParameter.name);
if (!identifierCounts || identifierCounts > 2) {
  continue;
}

function countTypeParameterUsage(
  checker: ts.TypeChecker,
  node: NodeWithTypeParameters,
): Map<ts.Identifier, number> {
  // Complex type-aware counting with recursive type analysis
}
```

**Go Implementation:**
```go
func countTypeParameterUsages(ctx rule.RuleContext, node *ast.Node, typeParamName string, typeParamNode *ast.Node) int {
  // Use text-based approach to count meaningful occurrences
  nodeText := string(ctx.SourceFile.Text()[node.Pos():node.End()])
  
  // Count occurrences of the type parameter name in the node text
  count := 0
  // ... simple string matching
}
```

**Issue:** Go implementation uses basic string matching instead of proper type-aware analysis, missing complex type relationships.

**Impact:** Cannot detect type parameter usage in inferred return types, mapped types, conditional types, or other complex TypeScript constructs.

**Test Coverage:** Test cases involving complex generic types, inferred types, and type transformations will fail.

#### 3. **Missing Complex Type Analysis**
**TypeScript Implementation:**
```typescript
function collectTypeParameterUsageCounts(
  checker: ts.TypeChecker,
  node: ts.Node,
  foundIdentifierUsages: Map<ts.Identifier, number>,
  fromClass: boolean,
): void {
  // Comprehensive type analysis including:
  // - Union/intersection types
  // - Index access types  
  // - Template literal types
  // - Conditional types
  // - Mapped types
  // - Object type properties
  // - Call/construct signatures
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
  // ... very basic type handling
}
```

**Issue:** Go implementation lacks comprehensive type analysis capabilities present in the TypeScript version.

**Impact:** Cannot properly analyze complex type relationships, missing many legitimate use cases.

**Test Coverage:** Tests involving complex type operations, mapped types, conditional types will produce incorrect results.

#### 4. **Inadequate AST Reference Tracking**
**TypeScript Implementation:**
```typescript
function isTypeParameterRepeatedInAST(
  node: TSESTree.TSTypeParameter,
  references: Reference[],
  startOfBody = Infinity,
): boolean {
  // Sophisticated reference analysis with:
  // - Reference scope checking
  // - Type vs value reference distinction
  // - Parent node analysis for type arguments
  // - Special handling for Array/ReadonlyArray
}
```

**Go Implementation:**
```go
func isTypeParameterRepeatedInAST(typeParam *ast.Node, references []*ast.Node, startOfBody int) bool {
  count := 0
  typeParamName := typeParam.AsTypeParameter().Name().Text()
  
  for _, ref := range references {
    // Basic position and name checking
    // Missing sophisticated reference analysis
  }
}
```

**Issue:** Go implementation lacks proper reference tracking and analysis.

**Impact:** Cannot distinguish between type and value references, missing context-aware reference counting.

**Test Coverage:** Tests with complex reference patterns will be affected.

#### 5. **Missing Fix/Suggestion Generation**
**TypeScript Implementation:**
```typescript
suggest: [
  {
    messageId: 'replaceUsagesWithConstraint',
    *fix(fixer): Generator<TSESLint.RuleFix> {
      // Complex fix generation including:
      // - Constraint text handling
      // - Reference replacement with proper parentheses
      // - Type parameter removal from declaration
      // - Comma handling for multiple parameters
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

**Issue:** Go implementation completely omits fix/suggestion generation.

**Impact:** Users don't get automated fixes, reducing rule utility.

**Test Coverage:** All test cases expect suggestions, so this is a major functionality gap.

#### 6. **Incomplete AST Node Kind Coverage**
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
  ast.KindArrowFunction: func(node *ast.Node) { /* ... */ },
  ast.KindFunctionDeclaration: func(node *ast.Node) { /* ... */ },
  // Missing: ast.KindDeclareFunction, ast.KindTSEmptyBodyFunctionExpression
  // Commented out due to compilation issues
}
```

**Issue:** Go implementation doesn't cover all AST node types that the TypeScript version handles.

**Impact:** Some function-like constructs won't be analyzed by the rule.

**Test Coverage:** Tests involving declare functions and other specific AST constructs.

### Recommendations
- **Implement proper scope tracking**: Add scope manager functionality or equivalent reference tracking system
- **Enhance type analysis**: Implement comprehensive type-aware analysis using available checker methods
- **Add fix generation**: Implement suggestion/fix generation to match TypeScript functionality
- **Complete AST coverage**: Add support for all relevant AST node kinds
- **Improve reference tracking**: Implement sophisticated reference analysis similar to TypeScript version
- **Add constraint handling**: Properly handle type parameter constraints and their relationships
- **Implement type-aware counting**: Replace string-based counting with proper type system integration

---