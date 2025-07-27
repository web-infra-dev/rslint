## Rule: class-methods-use-this

### Test File: class-methods-use-this.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic method detection, static method exclusion, constructor exclusion, `this` and `super` usage tracking, override modifier handling, implements interface handling, accessor property support, private/protected/public field handling
- ⚠️ **POTENTIAL ISSUES**: AST traversal patterns for arrow functions in properties, context stack management complexity, error reporting positioning
- ❌ **INCORRECT**: Accessor property detection logic, arrow function handling in property initializers, computed property name handling

### Discrepancies Found

#### 1. Accessor Property Detection Logic
**TypeScript Implementation:**
```typescript
node.type === AST_NODE_TYPES.AccessorProperty
```

**Go Implementation:**
```go
if ast.HasAccessorModifier(node) && (property.Initializer == nil || property.Initializer.Kind != ast.KindArrowFunction) {
    // This is an accessor property without arrow function - treat it like a method
    pushContext(node)
}
```

**Issue:** The Go implementation tries to detect accessor properties via `ast.HasAccessorModifier()` on `PropertyDeclaration` nodes, but the TypeScript implementation expects a distinct `AccessorProperty` AST node type. This suggests the Go AST may not have the same node structure for accessor properties.

**Impact:** Tests with `accessor method = () => {};` may not be handled correctly if the AST structure differs.

**Test Coverage:** Test cases with `accessor method = () => {};` syntax.

#### 2. Arrow Function Context Management
**TypeScript Implementation:**
```typescript
'PropertyDefinition > ArrowFunctionExpression.value'(node): void {
  enterFunction(node);
},
'PropertyDefinition > ArrowFunctionExpression.value:exit'(node): void {
  exitFunction(node);
}
```

**Go Implementation:**
```go
listeners[ast.KindArrowFunction] = func(node *ast.Node) {
    // Check if this arrow function is a property initializer
    if node.Parent != nil && node.Parent.Kind == ast.KindPropertyDeclaration {
        propDecl := node.Parent.AsPropertyDeclaration()
        if propDecl.Initializer == node {
            // This is handled by PropertyDeclaration logic, skip here
            return
        }
    }
    enterFunction(node)
}
```

**Issue:** The TypeScript implementation uses specific CSS-like selectors to handle arrow functions within property definitions, while the Go implementation uses a more complex conditional logic that may not perfectly match the TypeScript behavior.

**Impact:** Arrow functions in property initializers might be double-processed or missed entirely.

**Test Coverage:** Test cases with `property = () => {};` syntax.

#### 3. Computed Property Name Handling
**TypeScript Implementation:**
```typescript
if (node.computed || exceptMethods.size === 0) {
  return true;
}
```

**Go Implementation:**
```go
// Skip methods with computed property names
if hasComputedPropertyName(node) {
    return false
}
```

**Issue:** The TypeScript implementation returns `true` (include the method) when the property is computed, while the Go implementation returns `false` (exclude the method) for computed properties.

**Impact:** Methods with computed property names will behave differently between implementations.

**Test Coverage:** Test cases with `[computed]() {}` method syntax.

#### 4. Class Member Parent Detection
**TypeScript Implementation:**
```typescript
if (member?.parent.type === AST_NODE_TYPES.ClassBody) {
  stack = {
    class: member.parent.parent,
    member,
    parent: stack,
    usesThis: false,
  };
}
```

**Go Implementation:**
```go
// Check if the parent is a class declaration or expression
classNode := member.Parent
if classNode != nil && (classNode.Kind == ast.KindClassDeclaration || classNode.Kind == ast.KindClassExpression) {
    stack = &StackInfo{
        Class:    classNode,
        Member:   member,
        Parent:   stack,
        UsesThis: false,
    }
}
```

**Issue:** The TypeScript implementation expects the member's parent to be a `ClassBody` and then takes the parent's parent as the class. The Go implementation expects the member's direct parent to be the class. This suggests different AST structures.

**Impact:** Context detection for class members may fail, causing rules to not apply correctly.

**Test Coverage:** All class method test cases are affected.

#### 5. Private Identifier Handling
**TypeScript Implementation:**
```typescript
const hashIfNeeded = node.key.type === AST_NODE_TYPES.PrivateIdentifier ? '#' : '';
const name = getStaticMemberAccessValue(node, context);
return (typeof name !== 'string' || !exceptMethods.has(hashIfNeeded + name));
```

**Go Implementation:**
```go
// For private identifiers, the name already includes "#"
for _, exceptMethod := range options.ExceptMethods {
    if exceptMethod == name {
        return false
    }
}
```

**Issue:** The TypeScript implementation manually adds "#" prefix for private identifiers, while the Go implementation assumes the name already includes it. This could cause mismatches in exception handling.

**Impact:** Private methods in `exceptMethods` configuration may not be properly excluded.

**Test Coverage:** Test cases with `#method()` syntax and `exceptMethods` configuration.

#### 6. Error Reporting Position
**TypeScript Implementation:**
```typescript
context.report({
  loc: getFunctionHeadLoc(node, context.sourceCode),
  node,
  messageId: 'missingThis',
  data: { name: getFunctionNameWithKind(node) },
});
```

**Go Implementation:**
```go
// For methods, getters, and setters, report on the name node for better error positioning
var reportNode *ast.Node
switch node.Kind {
case ast.KindMethodDeclaration:
    if node.AsMethodDeclaration().Name() != nil {
        reportNode = node.AsMethodDeclaration().Name()
    }
}
if reportNode == nil {
    reportNode = node
}
ctx.ReportNode(reportNode, buildMissingThisMessage(functionName))
```

**Issue:** The TypeScript implementation uses `getFunctionHeadLoc` to get a specific location within the function, while the Go implementation reports on either the method name or the entire node.

**Impact:** Error positioning may differ between implementations, affecting IDE experience.

**Test Coverage:** All test cases will show this difference in error positioning.

### Recommendations
- **Fix Accessor Property Detection**: Investigate the actual AST structure for accessor properties in the Go parser and adjust detection logic accordingly
- **Simplify Arrow Function Handling**: Align the Go implementation's arrow function context management with the TypeScript selector-based approach
- **Correct Computed Property Logic**: Change the Go implementation to include (not exclude) computed property methods, matching TypeScript behavior
- **Fix Class Member Context**: Investigate the AST structure differences and adjust parent/class detection to match TypeScript's ClassBody -> Class pattern
- **Standardize Private Identifier Handling**: Ensure consistent "#" prefix handling between implementations
- **Improve Error Positioning**: Consider implementing `getFunctionHeadLoc` equivalent for more precise error reporting
- **Add Missing Test Cases**: Include tests for computed properties, complex accessor scenarios, and edge cases around AST structure differences

---