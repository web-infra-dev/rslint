# Rule Validation: explicit-module-boundary-types

## Rule: explicit-module-boundary-types

### Test File: explicit-module-boundary-types.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic export checking, function type checking, class member handling, option parsing structure
- ⚠️ **POTENTIAL ISSUES**: AST navigation patterns, parameter type checking logic, scope resolution, higher-order function detection
- ❌ **INCORRECT**: Missing scope-based reference following, incomplete parameter type annotation detection, simplified export tracking

### Discrepancies Found

#### 1. Missing Scope-Based Reference Following
**TypeScript Implementation:**
```typescript
function followReference(node: TSESTree.Identifier): void {
  const scope = context.sourceCode.getScope(node);
  const variable = scope.set.get(node.name);
  
  // check all of the definitions
  for (const definition of variable.defs) {
    checkNode(definition.node);
  }
  
  // follow references to find writes to the variable
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

**Issue:** The Go implementation completely lacks scope analysis and variable reference tracking, which is crucial for following exported identifiers to their definitions.

**Impact:** The rule will miss many cases where functions are exported indirectly through variable assignments or object properties.

**Test Coverage:** Test cases like `export default foo;` and `export { test };` will not work correctly.

#### 2. Incomplete Parameter Type Annotation Detection
**TypeScript Implementation:**
```typescript
function checkParameter(param: TSESTree.Parameter): void {
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
      return;
  }
}
```

**Go Implementation:**
```go
checkParameter = func(param *ast.Node) {
  // Check if parameter has type annotation
  if param.Kind == ast.KindIdentifier {
    // For identifiers, check parent for type annotation
    parent := param.Parent
    if parent != nil && parent.Kind == ast.KindParameter {
      paramNode := parent.AsParameterDeclaration()
      if paramNode.Type != nil {
        hasType = true
        // Check if it's any type
        if paramNode.Type.Kind == ast.KindAnyKeyword {
          isAnyType = true
        }
      }
    }
  }
}
```

**Issue:** The Go implementation has incomplete parameter type checking logic. It doesn't properly handle destructuring patterns, rest parameters, or the full range of parameter types.

**Impact:** Many parameter type validation cases will be missed, particularly for destructuring parameters and rest parameters.

**Test Coverage:** Tests with `{ foo }: any`, `[bar]: any`, and `...bar: any` parameters will fail.

#### 3. Simplified Higher-Order Function Detection
**TypeScript Implementation:**
```typescript
function isExportedHigherOrderFunction({ node }: FunctionInfo<FunctionNode>): boolean {
  let current: TSESTree.Node | undefined = node.parent;
  while (current) {
    if (current.type === AST_NODE_TYPES.ReturnStatement) {
      // the parent of a return will always be a block statement, so we can skip over it
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
isExportedHigherOrderFunction := func(info functionInfo) bool {
  current := info.node.Parent
  for current != nil {
    if current.Kind == ast.KindReturnStatement {
      // Skip block statement parent
      current = current.Parent.Parent
      continue
    }
    
    if !isFunction(current) {
      return false
    }
    
    returns := getReturnsInFunction(current)
    funcInfo := functionInfo{node: current, returns: returns}
    if !doesImmediatelyReturnFunctionExpression(funcInfo) {
      return false
    }
    
    if checkedFunctions[current] {
      return true
    }
    
    current = current.Parent
  }
  return false
}
```

**Issue:** While the Go implementation follows a similar pattern, it doesn't account for the complexity of AST traversal in the Go TypeScript AST, which may have different parent-child relationships.

**Impact:** Higher-order function detection may not work correctly in all cases.

**Test Coverage:** Complex nested higher-order function tests may fail.

#### 4. Missing Export Declaration Handling
**TypeScript Implementation:**
```typescript
'ExportNamedDeclaration:not([source]):exit'(
  node: TSESTree.ExportNamedDeclarationWithoutSource,
): void {
  if (node.declaration) {
    checkNode(node.declaration);
  } else {
    for (const specifier of node.specifiers) {
      followReference(specifier.local);
    }
  }
},
```

**Go Implementation:**
```go
ast.KindExportDeclaration: func(node *ast.Node) {
  exportDecl := node.AsExportDeclaration()
  if exportDecl.ModuleSpecifier == nil { // Not re-export
    if exportDecl.ExportClause != nil {
      // export { foo, bar }
      if exportDecl.ExportClause.Kind == ast.KindNamedExports {
        for _, spec := range exportDecl.ExportClause.AsNamedExports().Elements.Nodes {
          if spec.Kind == ast.KindExportSpecifier {
            specNode := spec.AsExportSpecifier()
            nameNode := specNode.Name()
            followReference(nameNode)
          }
        }
      }
    }
    // Note: Declaration field handling removed as API changed
  }
},
```

**Issue:** The Go implementation has incomplete export declaration handling and lacks the declaration field processing that the TypeScript version has.

**Impact:** Direct exports like `export function foo() {}` may not be properly detected.

**Test Coverage:** Many basic export tests will fail.

#### 5. Incomplete AST Node Kind Mapping
**TypeScript Implementation:**
```typescript
switch (node.type) {
  case AST_NODE_TYPES.ArrowFunctionExpression:
  case AST_NODE_TYPES.FunctionExpression:
  case AST_NODE_TYPES.ArrayExpression:
  case AST_NODE_TYPES.PropertyDefinition:
  case AST_NODE_TYPES.AccessorProperty:
  case AST_NODE_TYPES.MethodDefinition:
  case AST_NODE_TYPES.TSAbstractMethodDefinition:
  // ... many more cases
}
```

**Go Implementation:**
```go
switch node.Kind {
  case ast.KindArrowFunction, ast.KindFunctionExpression:
  case ast.KindArrayLiteralExpression:
  case ast.KindPropertyDeclaration, ast.KindMethodDeclaration:
  // ... fewer cases handled
}
```

**Issue:** The Go implementation doesn't handle all the AST node types that the TypeScript version does, particularly accessor properties and abstract methods.

**Impact:** Some class member types won't be properly checked.

**Test Coverage:** Tests with accessor properties and abstract methods may fail.

#### 6. Missing Direct Const Assertion Detection
**TypeScript Implementation:**
```typescript
// Complex logic for detecting as const patterns in arrow functions
// Handles nested satisfies expressions and complex type assertions
```

**Go Implementation:**
```go
func hasDirectConstAssertion(node *ast.Node) bool {
  // Simplified implementation that may miss edge cases
  // Missing proper handling of complex satisfies expressions
}
```

**Issue:** The Go implementation has a simplified version of const assertion detection that may miss complex patterns involving satisfies expressions.

**Impact:** Some const assertion tests may fail, particularly complex nested patterns.

**Test Coverage:** Tests with `satisfies` expressions and complex const assertions may not work.

#### 7. Overload Function Detection Issues
**TypeScript Implementation:**
```typescript
function hasOverloadSignatures(
  node: FunctionDeclaration | MethodDefinition,
  context: RuleContext,
): boolean {
  // Uses sophisticated scope analysis and sibling checking
}
```

**Go Implementation:**
```go
func hasOverloadSignatures(node *ast.Node, ctx rule.RuleContext) bool {
  // Simplified implementation that may not correctly identify overloads
  // Missing proper scope context analysis
}
```

**Issue:** The Go implementation has incomplete overload detection logic that doesn't properly use scope analysis.

**Impact:** Function overload tests may not work correctly.

**Test Coverage:** Tests with function overloads will likely fail.

### Recommendations
- **Implement proper scope analysis**: Add TypeScript type checker integration for variable reference tracking
- **Complete parameter type checking**: Implement full support for all parameter patterns (destructuring, rest, etc.)
- **Fix export declaration handling**: Ensure all export patterns are properly detected and followed
- **Enhance AST node coverage**: Add support for all missing AST node types (accessor properties, abstract methods, etc.)
- **Improve const assertion detection**: Implement complete detection for complex const assertion patterns
- **Add overload function detection**: Implement proper overload signature detection using scope analysis
- **Add comprehensive test coverage**: Ensure all TypeScript test cases pass with the Go implementation
- **Implement missing utility functions**: Add proper implementations for functions like `isStaticMemberAccessOfValue`

### Critical Missing Features
1. **Scope-based variable tracking**: Essential for following exported references
2. **Complete parameter validation**: Required for proper argument type checking
3. **Full export pattern support**: Needed for detecting all exported functions
4. **Type checker integration**: Required for proper TypeScript-aware analysis

---