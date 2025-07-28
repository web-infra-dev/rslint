## Rule: explicit-module-boundary-types

### Test File: explicit-module-boundary-types.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic function/method checking, return type validation, parameter type checking, message ID consistency, most configuration options
- ⚠️ **POTENTIAL ISSUES**: Higher-order function detection logic, typed function expression detection, overload signature detection, reference following for exported identifiers
- ❌ **INCORRECT**: AST traversal pattern differences, parameter type annotation detection, private member handling, export tracking mechanism

### Discrepancies Found

#### 1. AST Traversal and Export Detection Pattern
**TypeScript Implementation:**
```typescript
return {
  'ArrowFunctionExpression, FunctionDeclaration, FunctionExpression': enterFunction,
  'ExportDefaultDeclaration:exit'(node): void {
    checkNode(node.declaration);
  },
  'ExportNamedDeclaration:not([source]):exit'(node): void {
    if (node.declaration) {
      checkNode(node.declaration);
    } else {
      for (const specifier of node.specifiers) {
        followReference(specifier.local);
      }
    }
  },
  'Program:exit'(): void {
    for (const [node, returns] of functionReturnsMap) {
      if (isExportedHigherOrderFunction({ node, returns })) {
        checkNode(node);
      }
    }
  },
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
  ast.KindArrowFunction: func(node *ast.Node) {
    functionStack = append(functionStack, node)
    functionReturnsMap[node] = []*ast.Node{}
  },
  ast.KindExportAssignment: func(node *ast.Node) {
    exportDefault := node.AsExportAssignment()
    markExportedNodes(exportDefault.Expression, exportedFunctions)
  },
  rule.ListenerOnExit(ast.KindSourceFile): func(node *ast.Node) {
    // Check all functions that were tracked
    for funcNode := range exportedFunctions {
      // Process exported functions
    }
  },
}
```

**Issue:** The Go implementation uses a different AST traversal pattern that may miss export patterns handled by the TypeScript version. The TypeScript version uses specific exit listeners for export declarations, while Go uses a manual export tracking approach.

**Impact:** May miss some exported functions or incorrectly identify non-exported functions as needing type annotations.

**Test Coverage:** Tests involving `export default`, `export { ... }`, and complex export patterns may fail.

#### 2. Parameter Type Annotation Detection
**TypeScript Implementation:**
```typescript
switch (param.type) {
  case AST_NODE_TYPES.ArrayPattern:
  case AST_NODE_TYPES.Identifier:
  case AST_NODE_TYPES.ObjectPattern:
  case AST_NODE_TYPES.RestElement:
    if (!param.typeAnnotation) {
      report('missingArgType', 'missingArgTypeUnnamed');
    } else if (
      options.allowArgumentsExplicitlyTypedAsAny !== true &&
      param.typeAnnotation.typeAnnotation.type === AST_NODE_TYPES.TSAnyKeyword
    ) {
      report('anyTypedArg', 'anyTypedArgUnnamed');
    }
```

**Go Implementation:**
```go
hasType := false
isAnyType := false

if param.Kind == ast.KindIdentifier {
  parent := param.Parent
  if parent != nil && parent.Kind == ast.KindParameter {
    paramNode := parent.AsParameterDeclaration()
    if paramNode.Type != nil {
      hasType = true
      if paramNode.Type.Kind == ast.KindAnyKeyword {
        isAnyType = true
      }
    }
  }
}
```

**Issue:** The Go implementation's parameter type detection logic is incomplete and may not correctly identify type annotations on different parameter patterns (destructuring, rest parameters, etc.).

**Impact:** May incorrectly report missing type annotations or fail to detect `any` types in parameters.

**Test Coverage:** Tests with destructuring parameters, rest parameters, and `any` typed parameters may produce incorrect results.

#### 3. Higher-Order Function Detection
**TypeScript Implementation:**
```typescript
function isExportedHigherOrderFunction({
  node,
}: FunctionInfo<FunctionNode>): boolean {
  let current: TSESTree.Node | undefined = node.parent;
  while (current) {
    if (current.type === AST_NODE_TYPES.ReturnStatement) {
      current = current.parent.parent;
      continue;
    }
    if (!isFunction(current)) {
      return false;
    }
    const returns = getReturnsInFunction(current);
    if (!doesImmediatelyReturnFunctionExpression({ node: current, returns })) {
      return false;
    }
    if (checkedFunctions.has(current)) {
      return true;
    }
    current = current.parent;
  }
  return false;
}
```

**Go Implementation:**
```go
func doesImmediatelyReturnFunctionExpression(info functionInfo) bool {
  node := info.node
  returns := info.returns
  
  // For arrow functions, check if body is directly a function
  if node.Kind == ast.KindArrowFunction {
    arrowFunc := node.AsArrowFunction()
    if arrowFunc.Body != nil && arrowFunc.Body.Kind != ast.KindBlock {
      return isFunction(arrowFunc.Body)
    }
  }
  
  // For functions with block bodies, check return statements
  if len(returns) != 1 {
    return false
  }
  
  returnStatement := returns[0]
  if returnStatement.AsReturnStatement().Expression == nil {
    return false
  }
  
  expr := returnStatement.AsReturnStatement().Expression
  return isFunction(expr)
}
```

**Issue:** The Go implementation lacks the recursive parent traversal logic that determines if a function is part of an exported higher-order function chain. It only checks immediate return behavior.

**Impact:** May incorrectly apply or skip higher-order function allowances, leading to false positives or negatives.

**Test Coverage:** Complex higher-order function tests may fail.

#### 4. Typed Function Expression Detection
**TypeScript Implementation:**
```typescript
function isTypedFunctionExpression(node: TSESTree.Node, options: Options): boolean {
  if (!options.allowTypedFunctionExpressions) {
    return false;
  }
  
  const parent = node.parent;
  if (!parent) {
    return false;
  }
  
  // Comprehensive checks for various typed contexts
  // Variable declarator, as expressions, property assignments, etc.
}
```

**Go Implementation:**
```go
func isTypedFunctionExpression(node *ast.Node, options ExplicitModuleBoundaryTypesOptions) bool {
  if !options.AllowTypedFunctionExpressions {
    return false
  }
  
  parent := node.Parent
  if parent == nil {
    return false
  }
  
  // Basic checks for typed contexts
  if parent.Kind == ast.KindVariableDeclaration {
    // ...
  }
}
```

**Issue:** The Go implementation has incomplete logic for detecting typed function expressions and may miss many valid typed contexts.

**Impact:** May incorrectly require type annotations on functions that are already in typed contexts.

**Test Coverage:** Tests with typed function expressions in various contexts may fail.

#### 5. Private Member Handling
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.PropertyDefinition:
case AST_NODE_TYPES.AccessorProperty:
case AST_NODE_TYPES.MethodDefinition:
case AST_NODE_TYPES.TSAbstractMethodDefinition:
  if (
    node.accessibility === 'private' ||
    node.key.type === AST_NODE_TYPES.PrivateIdentifier
  ) {
    return;
  }
  return checkNode(node.value);
```

**Go Implementation:**
```go
// Skip private members  
if node.Kind == ast.KindPropertyDeclaration {
  prop := node.AsPropertyDeclaration()
  if prop.Modifiers() != nil {
    for _, mod := range prop.Modifiers().Nodes {
      if mod.Kind == ast.KindPrivateKeyword {
        return
      }
    }
  }
}
```

**Issue:** The Go implementation may not correctly handle all forms of private members (private identifiers with #, accessibility modifiers) and the logic is embedded in different places rather than being centralized.

**Impact:** May incorrectly check private members that should be ignored.

**Test Coverage:** Tests with private properties, private identifiers, and accessibility modifiers may fail.

#### 6. Reference Following for Export Tracking
**TypeScript Implementation:**
```typescript
function followReference(node: TSESTree.Identifier): void {
  const scope = context.sourceCode.getScope(node);
  const variable = scope.set.get(node.name);
  if (!variable) {
    return;
  }

  for (const definition of variable.defs) {
    // Check definitions
    checkNode(definition.node);
  }

  for (const reference of variable.references) {
    if (!reference.init && reference.writeExpr) {
      checkNode(reference.writeExpr);
    }
  }
}
```

**Go Implementation:**
```go
followReference := func(node *ast.Node) {
  if node.Kind != ast.KindIdentifier {
    return
  }
  
  // In a real implementation, we would use the type checker to resolve references
  // For now, we'll do a simplified version
  // This would need proper scope analysis
}
```

**Issue:** The Go implementation has a placeholder for reference following and lacks proper scope analysis to track exported identifiers.

**Impact:** May miss functions that are exported through identifier references.

**Test Coverage:** Tests involving `export { functionName }` patterns may fail.

### Recommendations
- Implement proper export declaration handling with specific listeners for different export patterns
- Complete the parameter type annotation detection logic for all parameter patterns (destructuring, rest, etc.)
- Add recursive parent traversal logic for higher-order function detection
- Enhance typed function expression detection to cover all valid typed contexts
- Centralize and improve private member detection logic
- Implement proper scope analysis and reference following for export tracking
- Add comprehensive test coverage for edge cases around exports, parameter patterns, and private members
- Consider using the TypeScript compiler's scope and reference resolution capabilities more extensively

---