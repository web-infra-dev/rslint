## Rule: no-empty-function

### Test File: no-empty-function.test.ts

### Validation Summary
- ✅ **CORRECT**: Core empty function detection logic, basic configuration option parsing, main function types handled (functions, arrow functions, methods, constructors, getters, setters), parameter properties detection for constructors, basic async/generator function detection
- ⚠️ **POTENTIAL ISSUES**: AST node type mapping differences, decorator detection implementation, override method detection, error positioning accuracy
- ❌ **INCORRECT**: Constructor accessibility modifier detection logic, decorator detection method, some AST navigation patterns

### Discrepancies Found

#### 1. Constructor Accessibility Detection
**TypeScript Implementation:**
```typescript
function isAllowedEmptyConstructor(
  node: TSESTree.FunctionDeclaration | TSESTree.FunctionExpression,
): boolean {
  const parent = node.parent;
  if (
    isBodyEmpty(node) &&
    parent.type === AST_NODE_TYPES.MethodDefinition &&
    parent.kind === 'constructor'
  ) {
    const { accessibility } = parent;

    return (
      // allow protected constructors
      (accessibility === 'protected' && isAllowedProtectedConstructors) ||
      // allow private constructors
      (accessibility === 'private' && isAllowedPrivateConstructors) ||
      // allow constructors which have parameter properties
      hasParameterProperties(node)
    );
  }

  return false;
}
```

**Go Implementation:**
```go
// In the constructor case within checkFunction:
if node.Kind == ast.KindConstructor {
    // Check accessibility modifiers for constructors
    hasPrivate := ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate)
    hasProtected := ast.HasSyntacticModifier(node, ast.ModifierFlagsProtected)

    if isAllowed("constructors") {
        return
    }
    if hasPrivate && isAllowed("private-constructors") {
        return
    }
    if hasProtected && isAllowed("protected-constructors") {
        return
    }

    // Constructors with parameter properties are allowed
    if hasParameterProperties(node) {
        return
    }
}
```

**Issue:** The TypeScript version checks accessibility on the parent MethodDefinition node, while the Go version checks modifiers directly on the constructor node. This may not correctly identify private/protected constructors.

**Impact:** Private and protected constructors may not be properly detected, causing the rule to incorrectly report or allow empty constructors.

**Test Coverage:** Test cases for private and protected constructors would reveal this issue.

#### 2. Decorator Detection Method
**TypeScript Implementation:**
```typescript
function isAllowedEmptyDecoratedFunctions(
  node: TSESTree.FunctionDeclaration | TSESTree.FunctionExpression,
): boolean {
  if (isAllowedDecoratedFunctions && isBodyEmpty(node)) {
    const decorators =
      node.parent.type === AST_NODE_TYPES.MethodDefinition
        ? node.parent.decorators
        : undefined;
    return !!decorators && !!decorators.length;
  }

  return false;
}
```

**Go Implementation:**
```go
// Decorated function check
if ast.GetCombinedModifierFlags(node)&ast.ModifierFlagsDecorator != 0 && isAllowed("decoratedFunctions") {
    return
}
```

**Issue:** The TypeScript version checks for decorators on the parent MethodDefinition, while the Go version checks modifier flags directly on the node. The decorator detection approach is fundamentally different.

**Impact:** Decorated functions may not be properly identified, causing incorrect rule behavior for functions with decorators.

**Test Coverage:** The decorator test case would likely fail with the current Go implementation.

#### 3. Override Method Detection Inconsistency
**TypeScript Implementation:**
```typescript
function isAllowedEmptyOverrideMethod(
  node: TSESTree.FunctionExpression,
): boolean {
  return (
    isAllowedOverrideMethods &&
    isBodyEmpty(node) &&
    node.parent.type === AST_NODE_TYPES.MethodDefinition &&
    node.parent.override
  );
}
```

**Go Implementation:**
```go
// Override method check
if ast.HasSyntacticModifier(node, ast.ModifierFlagsOverride) && isAllowed("overrideMethods") {
    return
}
```

**Issue:** The TypeScript version checks the `override` property on the parent MethodDefinition, while the Go version checks the modifier flag directly on the node. This mirrors the decorator issue.

**Impact:** Override methods may not be correctly identified, causing the rule to behave incorrectly for overridden methods.

**Test Coverage:** The override method test case would reveal this discrepancy.

#### 4. AST Node Structure Differences
**TypeScript Implementation:**
```typescript
// Works with FunctionExpression nodes inside MethodDefinition parents
FunctionExpression(node): void {
  if (
    isAllowedEmptyConstructor(node) ||
    isAllowedEmptyDecoratedFunctions(node) ||
    isAllowedEmptyOverrideMethod(node)
  ) {
    return;
  }

  rules.FunctionExpression(node);
}
```

**Go Implementation:**
```go
// Handles method declarations directly without parent-child relationship
return rule.RuleListeners{
    ast.KindFunctionDeclaration: checkFunction,
    ast.KindFunctionExpression:  checkFunction,
    ast.KindArrowFunction:       checkFunction,
    ast.KindConstructor:         checkFunction,
    ast.KindMethodDeclaration:   checkFunction,
    ast.KindGetAccessor:         checkFunction,
    ast.KindSetAccessor:         checkFunction,
}
```

**Issue:** The AST structure is different between TypeScript-ESLint and typescript-go. The TypeScript version primarily handles FunctionExpression nodes with parent context, while the Go version handles various node types directly.

**Impact:** This fundamental difference could lead to missed or incorrectly handled cases, especially for method-related functionality.

**Test Coverage:** All method-related test cases could be affected.

#### 5. Error Position Reporting
**TypeScript Implementation:**
```typescript
// Uses the base ESLint rule's error reporting which targets the opening brace
rules.FunctionExpression(node);
```

**Go Implementation:**
```go
// Custom brace position detection with fallback
getOpenBracePosition := func(node *ast.Node) (core.TextRange, bool) {
    // Complex logic to find opening brace position
    // ...
    // Fallback: use the body's start position
    return core.TextRange{}.WithPos(bodyStart).WithEnd(bodyStart + 1), true
}
```

**Issue:** The Go implementation has custom logic for finding the opening brace position, which may not exactly match the TypeScript-ESLint behavior, especially in edge cases.

**Impact:** Error column positions may not match expected test results exactly.

**Test Coverage:** All invalid test cases check specific column positions which could reveal mismatches.

### Recommendations
- Fix constructor accessibility detection by checking the correct AST node or parent node for modifiers
- Implement proper decorator detection that matches the TypeScript-ESLint approach
- Fix override method detection to check the appropriate node/parent for the override flag
- Review and align the AST node type handling with the expected TypeScript-ESLint patterns
- Verify error position reporting matches expected column numbers in tests
- Add comprehensive test coverage for edge cases involving nested function contexts
- Consider adding debug logging to verify which code paths are being taken during rule execution

---