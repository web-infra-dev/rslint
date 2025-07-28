## Rule: no-inferrable-types

### Test File: no-inferrable-types.test.ts

### Validation Summary
- ✅ **CORRECT**: Core rule concept, basic type inference detection, configuration option handling, most AST pattern matching
- ⚠️ **POTENTIAL ISSUES**: Void expression handling differences, optional chaining implementation, literal type handling inconsistencies
- ❌ **INCORRECT**: Missing TSNullKeyword handling in null literal detection, incomplete skipChainExpression logic, parameter property detection gaps

### Discrepancies Found

#### 1. Null Literal Detection Mismatch
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TSNullKeyword:
  return init.type === AST_NODE_TYPES.Literal && init.value == null;
```

**Go Implementation:**
```go
case ast.KindNullKeyword:
  return isLiteral(init, "null")

// isLiteral function:
case "null":
  return init.Kind == ast.KindNullKeyword
```

**Issue:** The TypeScript version checks for `Literal` nodes with null values, while Go only checks for `KindNullKeyword`. This misses cases where `null` appears as a literal value.

**Impact:** May fail to detect some null literal patterns that should trigger the rule.

**Test Coverage:** Test case `const a: null = null;` may not behave consistently.

#### 2. Void Expression Handling Inconsistency
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TSUndefinedKeyword:
  return (
    hasUnaryPrefix(init, 'void') || isIdentifier(init, 'undefined')
  );
```

**Go Implementation:**
```go
case ast.KindUndefinedKeyword:
  isVoidExpr := init.Kind == ast.KindVoidExpression
  literalResult := isLiteral(init, "undefined")
  return isVoidExpr || literalResult
```

**Issue:** TypeScript checks for unary prefix with 'void' operator, while Go directly checks for `KindVoidExpression`. These may not be equivalent AST representations.

**Impact:** May miss or incorrectly flag void expressions like `void someValue`.

**Test Coverage:** Test case `const a: undefined = void someValue;` needs verification.

#### 3. Optional Chaining Logic Gaps
**TypeScript Implementation:**
```typescript
function skipChainExpression(init: TSESTree.Expression): TSESTree.Expression {
  const node = skipChainExpression(init);
  // Comprehensive handling of chain expressions
}
```

**Go Implementation:**
```go
skipChainExpression := func(node *ast.Node) *ast.Node {
  // Only handles specific optional chaining cases
  // Missing comprehensive chain expression unwrapping
}
```

**Issue:** The Go version's `skipChainExpression` is less comprehensive and may not handle all chain expression patterns that the TypeScript version covers.

**Impact:** May fail to detect inferrable types in complex optional chaining scenarios.

**Test Coverage:** Test cases with `BigInt?.()`, `Number?.()` etc. may not work correctly.

#### 4. Parameter Property Detection Missing
**TypeScript Implementation:**
```typescript
node.params.forEach(param => {
  if (param.type === AST_NODE_TYPES.TSParameterProperty) {
    param = param.parameter;
  }
  if (param.type === AST_NODE_TYPES.AssignmentPattern) {
    reportInferrableType(param, param.left.typeAnnotation, param.right);
  }
});
```

**Go Implementation:**
```go
// Missing explicit TSParameterProperty handling
// Only handles basic Parameter nodes
```

**Issue:** The Go version doesn't explicitly handle TypeScript parameter properties (constructor parameters with visibility modifiers).

**Impact:** May miss constructor parameter properties that should be flagged.

**Test Coverage:** Test case `constructor(public a: boolean = true)` may not be detected properly.

#### 5. Template Literal AST Node Mismatch
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TSStringKeyword:
  return (
    isFunctionCall(init, 'String') ||
    isLiteral(init, 'string') ||
    init.type === AST_NODE_TYPES.TemplateLiteral
  );
```

**Go Implementation:**
```go
case ast.KindStringKeyword:
  return isFunctionCall(init, "String") ||
    isLiteral(init, "string") ||
    init.Kind == ast.KindTemplateExpression ||
    init.Kind == ast.KindNoSubstitutionTemplateLiteral
```

**Issue:** TypeScript checks for `TemplateLiteral` while Go checks for both `TemplateExpression` and `NoSubstitutionTemplateLiteral`. This may be correct but needs verification.

**Impact:** Potential differences in template literal detection.

**Test Coverage:** Test case with template literals `const a: string = \`str\`;` needs verification.

#### 6. Literal Type Handling Addition
**TypeScript Implementation:**
```typescript
// No explicit KindLiteralType handling
```

**Go Implementation:**
```go
case ast.KindLiteralType:
  // Handle literal types like `null`, `undefined`, boolean literals, etc.
  literalType := annotation.AsLiteralTypeNode()
  // ... additional logic
```

**Issue:** The Go version adds `KindLiteralType` handling that doesn't exist in the TypeScript version. This may be necessary for Go's AST but could lead to behavioral differences.

**Impact:** May flag additional cases that TypeScript version doesn't catch.

**Test Coverage:** May affect literal type annotations.

#### 7. Assignment Pattern vs Variable Declaration Confusion
**TypeScript Implementation:**
```typescript
function inferrableVariableVisitor(node: TSESTree.VariableDeclarator): void {
  reportInferrableType(node, node.id.typeAnnotation, node.init);
}
```

**Go Implementation:**
```go
inferrableVariableVisitor := func(node *ast.Node) {
  varDecl := node.AsVariableDeclaration()
  if varDecl.Type != nil && varDecl.Initializer != nil {
    reportTarget := varDecl.Name()
    reportInferrableType(node, varDecl.Type, varDecl.Initializer, reportTarget.AsNode())
  }
}
```

**Issue:** TypeScript operates on `VariableDeclarator` nodes while Go operates on `VariableDeclaration` nodes. This is a fundamental AST structure difference.

**Impact:** May miss or incorrectly process variable declarations.

**Test Coverage:** All variable declaration test cases need verification.

### Recommendations
- Fix null literal detection to match TypeScript behavior by checking for literal nodes with null values
- Verify and align void expression handling between implementations
- Enhance skipChainExpression logic to be more comprehensive
- Add explicit TSParameterProperty handling for constructor parameters
- Verify template literal AST node mapping is correct
- Ensure variable declaration vs declarator handling is appropriate for the Go AST
- Test all edge cases thoroughly, especially optional chaining and parameter properties
- Consider removing or verifying the necessity of KindLiteralType handling if it diverges from TypeScript behavior

---