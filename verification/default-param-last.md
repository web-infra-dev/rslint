## Rule: default-param-last

### Test File: default-param-last.test.ts

### Validation Summary
- ✅ **CORRECT**: Core rule logic matches, right-to-left parameter iteration, error message consistency, basic parameter classification
- ⚠️ **POTENTIAL ISSUES**: TypeScript parameter property handling, destructuring parameter detection with defaults, expanded function type coverage
- ❌ **INCORRECT**: Missing TSParameterProperty unwrapping, incomplete destructuring pattern support for default detection

### Discrepancies Found

#### 1. Missing TSParameterProperty Unwrapping
**TypeScript Implementation:**
```typescript
const param =
  current.type === AST_NODE_TYPES.TSParameterProperty
    ? current.parameter
    : current;

if (isPlainParam(param)) {
  hasSeenPlainParam = true;
  continue;
}

if (hasSeenPlainParam &&
    (isOptionalParam(param) || param.type === AST_NODE_TYPES.AssignmentPattern)) {
  context.report({ node: current, messageId: 'shouldBeLast' });
}
```

**Go Implementation:**
```go
// No TSParameterProperty handling - processes parameters directly
for i := len(params) - 1; i >= 0; i-- {
    current := params[i]
    if current == nil {
        continue
    }

    if isPlainParam(current) {
        hasSeenPlainParam = true
        continue
    }
    
    if hasSeenPlainParam && (isOptionalParam(current) || isDefaultParam(current)) {
        violatingParams = append(violatingParams, current)
    }
}
```

**Issue:** The TypeScript version handles constructor parameter properties (like `public a = 0, private b: number`) by unwrapping the inner parameter from the TSParameterProperty node before analysis. The Go version processes parameters directly without this unwrapping.

**Impact:** Constructor parameters with access modifiers will not be analyzed correctly, causing failures in test cases like the constructor tests with `public`, `private`, `protected` modifiers.

**Test Coverage:** Multiple constructor test cases will fail: `public a = 0, private b: number` scenarios.

#### 2. Destructuring Parameter Default Detection
**TypeScript Implementation:**
```typescript
// Detects AssignmentPattern directly as default parameter
function isPlainParam(node: TSESTree.Parameter): boolean {
  return !(
    node.type === AST_NODE_TYPES.AssignmentPattern ||
    node.type === AST_NODE_TYPES.RestElement ||
    isOptionalParam(node)
  );
}

// Also checks for AssignmentPattern in violation detection
if (hasSeenPlainParam &&
    (isOptionalParam(param) || param.type === AST_NODE_TYPES.AssignmentPattern)) {
  context.report({ node: current, messageId: 'shouldBeLast' });
}
```

**Go Implementation:**
```go
func isDefaultParam(node *ast.Node) bool {
    if node == nil || !ast.IsParameter(node) {
        return false
    }
    
    param := node.AsParameterDeclaration()
    return param.Initializer != nil
}
```

**Issue:** The TypeScript version explicitly checks for `AssignmentPattern` nodes (which represent destructuring with defaults like `{ a } = {}`), while the Go version only checks `Initializer` on `ParameterDeclaration` nodes.

**Impact:** Destructuring parameters with defaults like `function foo({ a } = {}, b) {}` and `function foo([a] = [], b) {}` may not be detected as default parameters.

**Test Coverage:** Test cases with destructuring defaults will likely fail.

#### 3. Optional Parameter Pattern Support
**TypeScript Implementation:**
```typescript
function isOptionalParam(node: TSESTree.Parameter): boolean {
  return (
    (node.type === AST_NODE_TYPES.ArrayPattern ||
      node.type === AST_NODE_TYPES.AssignmentPattern ||
      node.type === AST_NODE_TYPES.Identifier ||
      node.type === AST_NODE_TYPES.ObjectPattern ||
      node.type === AST_NODE_TYPES.RestElement) &&
    node.optional
  );
}
```

**Go Implementation:**
```go
func isOptionalParam(node *ast.Node) bool {
    if node == nil || !ast.IsParameter(node) {
        return false
    }
    
    param := node.AsParameterDeclaration()
    return param.QuestionToken != nil
}
```

**Issue:** The TypeScript version checks for `optional` property on various parameter pattern types, while the Go version only checks `QuestionToken` on parameter declarations.

**Impact:** Optional destructuring parameters or other parameter patterns may not be detected correctly.

**Test Coverage:** Edge cases with optional parameter patterns may fail.

#### 4. Expanded Function Type Coverage
**TypeScript Implementation:**
```typescript
return {
  ArrowFunctionExpression: checkDefaultParamLast,
  FunctionDeclaration: checkDefaultParamLast,
  FunctionExpression: checkDefaultParamLast,
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
  ast.KindArrowFunction: func(node *ast.Node) { checkDefaultParamLast(ctx, node) },
  ast.KindFunctionDeclaration: func(node *ast.Node) { checkDefaultParamLast(ctx, node) },
  ast.KindFunctionExpression: func(node *ast.Node) { checkDefaultParamLast(ctx, node) },
  ast.KindMethodDeclaration: func(node *ast.Node) { checkDefaultParamLast(ctx, node) },
  ast.KindConstructor: func(node *ast.Node) { checkDefaultParamLast(ctx, node) },
  ast.KindGetAccessor: func(node *ast.Node) { checkDefaultParamLast(ctx, node) },
  ast.KindSetAccessor: func(node *ast.Node) { checkDefaultParamLast(ctx, node) },
}
```

**Issue:** The Go implementation covers additional function types (methods, constructors, accessors) beyond the original TypeScript-ESLint rule scope.

**Impact:** The rule will trigger on more function types than the original rule, which could be seen as enhanced coverage or unwanted deviation depending on intent.

**Test Coverage:** Constructor test cases are included in the test suite, suggesting this expansion is intentional.

#### 5. Parameter Extraction Complexity
**TypeScript Implementation:**
```typescript
// Simple parameter access
for (let i = node.params.length - 1; i >= 0; i--) {
  const current = node.params[i];
}
```

**Go Implementation:**
```go
// Complex parameter extraction with switch statement
switch functionNode.Kind {
case ast.KindArrowFunction:
    if functionNode.AsArrowFunction().Parameters != nil {
        params = functionNode.AsArrowFunction().Parameters.Nodes
    }
case ast.KindFunctionDeclaration:
    if functionNode.AsFunctionDeclaration().Parameters != nil {
        params = functionNode.AsFunctionDeclaration().Parameters.Nodes
    }
// ... more cases
}
```

**Issue:** The Go version requires explicit parameter extraction for each function type, which is more complex but necessary due to the type system.

**Impact:** Risk of missing parameter extraction for new function types, but provides type safety.

**Test Coverage:** All function types need verification that parameter extraction works correctly.

### Recommendations
- **CRITICAL**: Implement TSParameterProperty unwrapping logic to handle constructor parameter properties correctly
- **HIGH**: Add proper destructuring parameter detection for AssignmentPattern, ObjectPattern, and ArrayPattern nodes
- **HIGH**: Verify that parameter extraction works correctly for all supported function types
- **MEDIUM**: Consider whether expanded function type coverage (methods, accessors) is desired or should match original rule
- **LOW**: Add comprehensive test coverage for edge cases with complex parameter patterns
- **LOW**: Consider adding debug logging to verify parameter classification logic

---