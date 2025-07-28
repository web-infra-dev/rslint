## Rule: class-methods-use-this

### Test File: class-methods-use-this.test.ts

### Validation Summary
- ✅ **CORRECT**: Core rule logic structure and flow, basic AST pattern matching for methods/getters/setters, static method exclusion, constructor exclusion, `this` and `super` detection, configuration option parsing
- ⚠️ **POTENTIAL ISSUES**: Stack context management differs from TypeScript implementation, accessor property handling implementation, complex nested function scenarios
- ❌ **INCORRECT**: Accessor property AST node type mismatch, `enforceForClassFields` arrow function handling, parent-child relationship traversal for class member detection

### Discrepancies Found

#### 1. Accessor Property AST Node Type Mismatch
**TypeScript Implementation:**
```typescript
| TSESTree.AccessorProperty
| TSESTree.MethodDefinition
| TSESTree.PropertyDefinition;

'AccessorProperty > ArrowFunctionExpression.value'(node): void {
  enterFunction(node);
},
```

**Go Implementation:**
```go
// Note: AccessorProperty doesn't exist in current AST, handled via PropertyDeclaration with accessor modifier
if ast.HasAccessorModifier(node) && property.Initializer == nil {
    // This is an accessor property without initializer - treat it like a method
    pushContext(node)
}
```

**Issue:** The TypeScript implementation uses a specific `AccessorProperty` AST node type that doesn't exist in the Go AST. The Go implementation attempts to handle this through `PropertyDeclaration` with accessor modifiers, but this may not capture all accessor property scenarios correctly.

**Impact:** May miss or incorrectly handle accessor properties like `accessor method = () => {};`

**Test Coverage:** Tests with `accessor method = () => {};` patterns may fail

#### 2. Stack Context Management Differences
**TypeScript Implementation:**
```typescript
function pushContext(member?: TSESTree.AccessorProperty | TSESTree.MethodDefinition | TSESTree.PropertyDefinition): void {
  if (member?.parent.type === AST_NODE_TYPES.ClassBody) {
    stack = {
      class: member.parent.parent,
      member,
      parent: stack,
      usesThis: false,
    };
  } else {
    stack = {
      class: null,
      member: null,
      parent: stack,
      usesThis: false,
    };
  }
}
```

**Go Implementation:**
```go
pushContext := func(member *ast.Node) {
    if member != nil && member.Parent != nil {
        // Check if the parent is a class declaration or expression
        classNode := member.Parent
        if classNode != nil && (classNode.Kind == ast.KindClassDeclaration || classNode.Kind == ast.KindClassExpression) {
            stack = &StackInfo{
                Class:    classNode,
                Member:   member,
                Parent:   stack,
                UsesThis: false,
            }
        } else {
            stack = &StackInfo{
                Class:    nil,
                Member:   nil,
                Parent:   stack,
                UsesThis: false,
            }
        }
    } else {
        stack = &StackInfo{
            Class:    nil,
            Member:   nil,
            Parent:   stack,
            UsesThis: false,
        }
    }
}
```

**Issue:** The Go implementation checks if `member.Parent` is directly a class, but TypeScript checks if `member.parent.type === AST_NODE_TYPES.ClassBody` and then uses `member.parent.parent` as the class. This suggests the Go version may be missing the intermediate ClassBody node.

**Impact:** May incorrectly identify class context or fail to associate methods with their containing class.

**Test Coverage:** All test cases involving class methods could be affected

#### 3. Arrow Function in Property Handling
**TypeScript Implementation:**
```typescript
'PropertyDefinition > ArrowFunctionExpression.value'(node: TSESTree.ArrowFunctionExpression): void {
  enterFunction(node);
},
'PropertyDefinition > ArrowFunctionExpression.value:exit'(node: TSESTree.ArrowFunctionExpression): void {
  exitFunction(node);
},
```

**Go Implementation:**
```go
listeners[ast.KindPropertyDeclaration] = func(node *ast.Node) {
    property := node.AsPropertyDeclaration()
    if property.Initializer != nil && property.Initializer.Kind == ast.KindArrowFunction {
        // This is a property with arrow function initializer - treat it like a method
        pushContext(node)
    }
}
```

**Issue:** The TypeScript implementation uses specific selectors to target arrow functions that are direct values of property definitions, while the Go implementation manually checks in the property declaration listener. The Go approach handles this in enter/exit of PropertyDeclaration rather than specifically targeting the arrow function.

**Impact:** May not properly track `this` usage in arrow function property initializers.

**Test Coverage:** Tests with `property = () => {};` patterns

#### 4. Function Name Extraction Inconsistency
**TypeScript Implementation:**
```typescript
getFunctionNameWithKind(node)
// Uses utility function to extract consistent names
```

**Go Implementation:**
```go
func getFunctionNameWithKind(ctx rule.RuleContext, node *ast.Node) string {
    switch node.Kind {
    case ast.KindFunctionExpression:
        if node.AsFunctionExpression().Name() != nil {
            return "function '" + node.AsFunctionExpression().Name().AsIdentifier().Text + "'"
        }
        return "function"
    // ... more cases
}
```

**Issue:** The Go implementation has custom logic for name extraction that may not match the TypeScript utility's behavior exactly, particularly for edge cases like computed property names or complex method definitions.

**Impact:** Error messages may differ in format or content from expected TypeScript-ESLint output.

**Test Coverage:** All error message assertions in tests

#### 5. Class Implementation Detection
**TypeScript Implementation:**
```typescript
if (ignoreClassesThatImplementAnInterface === true &&
    stackContext.class.implements.length > 0) {
  return;
}
```

**Go Implementation:**
```go
if classDecl.HeritageClauses != nil && len(classDecl.HeritageClauses.Nodes) > 0 {
    for _, clause := range classDecl.HeritageClauses.Nodes {
        if clause.AsHeritageClause().Token == ast.KindImplementsKeyword {
            hasImplements = true
            break
        }
    }
}
```

**Issue:** The TypeScript implementation directly accesses `class.implements`, while the Go implementation manually traverses `HeritageClauses` to find implements keywords. This difference in AST structure access could lead to different behavior.

**Impact:** May not correctly identify classes that implement interfaces, affecting the `ignoreClassesThatImplementAnInterface` option.

**Test Coverage:** Tests with `implements Bar` clauses

### Recommendations
- **Fix accessor property handling**: Investigate the correct AST node types for accessor properties in the Go AST and update the implementation accordingly
- **Correct stack context management**: Ensure the parent-child relationship traversal matches the TypeScript implementation's ClassBody intermediate node logic
- **Improve arrow function property handling**: Align the Go implementation's approach to match TypeScript's specific selector-based targeting of arrow functions in property initializers
- **Standardize function name extraction**: Use or port the TypeScript utility functions for consistent name extraction
- **Verify class implementation detection**: Ensure the Go AST traversal for heritage clauses correctly identifies implementing classes
- **Add missing AST pattern coverage**: Review test cases to identify any AST patterns not properly handled by the current Go implementation
- **Test error message formatting**: Verify that error messages match exactly with TypeScript-ESLint output format

---