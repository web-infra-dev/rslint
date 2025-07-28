## Rule: consistent-generic-constructors

### Test File: consistent-generic-constructors.test.ts

### Validation Summary
- ✅ **CORRECT**: Core rule logic for detecting generic constructor patterns, basic AST pattern matching for variable declarations and property declarations, error message structure and IDs, basic option parsing for constructor vs type-annotation modes
- ⚠️ **POTENTIAL ISSUES**: Comment preservation during fixes may not be as sophisticated as TypeScript version, handling of computed property names in type annotation attachment, binding element processing logic differs from TypeScript selector approach
- ❌ **INCORRECT**: Missing support for AssignmentPattern nodes (destructuring with defaults in function parameters), AST node selector logic doesn't match TypeScript's sophisticated selector, isolatedDeclarations option access may not work correctly

### Discrepancies Found

#### 1. Missing AssignmentPattern Support
**TypeScript Implementation:**
```typescript
'VariableDeclarator,PropertyDefinition,AccessorProperty,:matches(FunctionDeclaration,FunctionExpression) > AssignmentPattern'(
  node: TSESTree.AccessorProperty | TSESTree.AssignmentPattern | TSESTree.PropertyDefinition | TSESTree.VariableDeclarator,
): void {
  // Handles AssignmentPattern nodes which are function parameter defaults with destructuring
```

**Go Implementation:**
```go
return rule.RuleListeners{
    ast.KindVariableDeclaration: handleNode,
    ast.KindPropertyDeclaration: handleNode, // Includes accessor properties
    ast.KindParameter:           handleNode,
    ast.KindBindingElement:      handleNode, // Support for assignment patterns
}
```

**Issue:** The TypeScript version uses a sophisticated CSS-like selector that specifically targets AssignmentPattern nodes that are children of FunctionDeclaration or FunctionExpression. The Go version attempts to handle this with KindBindingElement but the logic is different.

**Impact:** Test cases involving destructuring parameters with defaults (like `function foo({ a }: Foo<string> = new Foo()) {}`) may not be handled correctly.

**Test Coverage:** Tests with function parameter destructuring patterns will reveal this issue.

#### 2. Selector Logic Mismatch
**TypeScript Implementation:**
```typescript
':matches(FunctionDeclaration,FunctionExpression) > AssignmentPattern'
```

**Go Implementation:**
```go
if node.Kind == ast.KindBindingElement {
    // Only process binding elements that are function parameters
    current := node.Parent
    for current != nil {
        if current.Kind == ast.KindParameter {
            break
        }
        // If we find a variable declaration or other non-parameter context, skip
        if current.Kind == ast.KindVariableDeclaration ||
            current.Kind == ast.KindVariableStatement ||
            current.Kind == ast.KindVariableDeclarationList {
            return
        }
        current = current.Parent
    }
    // If we didn't find a parameter parent, skip this binding element
    if current == nil {
        return
    }
}
```

**Issue:** The Go version tries to mimic the selector behavior but may miss edge cases or have different traversal logic than the TypeScript selector engine.

**Impact:** Some valid cases might be skipped or invalid cases might be processed.

**Test Coverage:** Function parameter destructuring tests may fail.

#### 3. IsolatedDeclarations Option Access
**TypeScript Implementation:**
```typescript
const isolatedDeclarations = context.parserOptions.isolatedDeclarations;
```

**Go Implementation:**
```go
isolatedDeclarations := ctx.Program.Options().IsolatedDeclarations.IsTrue()
```

**Issue:** The Go version accesses isolatedDeclarations through the Program options, but it's unclear if this correctly corresponds to the TypeScript parser options.

**Impact:** The rule behavior when isolatedDeclarations is enabled may not match TypeScript ESLint.

**Test Coverage:** The commented-out test case for isolatedDeclarations in the Go test file indicates this feature isn't fully working.

#### 4. Comment Preservation Complexity
**TypeScript Implementation:**
```typescript
const extraComments = new Set(
  context.sourceCode.getCommentsInside(lhs.parent),
);
context.sourceCode
  .getCommentsInside(lhs.typeArguments)
  .forEach(c => extraComments.delete(c));
```

**Go Implementation:**
```go
// Basic comment preservation - get any comments within the type arguments
// This is a simplified approach compared to the sophisticated TypeScript version
typeAnnotation := calleeText + typeArgsText
```

**Issue:** The Go version has a much simpler comment preservation strategy and may not handle complex comment scenarios as well as the TypeScript version.

**Impact:** Comments in type arguments may not be preserved correctly during fixes.

**Test Coverage:** Tests with comments in type arguments may produce different output.

#### 5. getLHSRHS Function Logic Differences
**TypeScript Implementation:**
```typescript
function getLHSRHS(): [
  (TSESTree.AccessorProperty | TSESTree.BindingName | TSESTree.PropertyDefinition),
  TSESTree.Expression | null,
] {
  switch (node.type) {
    case AST_NODE_TYPES.VariableDeclarator:
      return [node.id, node.init];
    case AST_NODE_TYPES.PropertyDefinition:
    case AST_NODE_TYPES.AccessorProperty:
      return [node, node.value];
    case AST_NODE_TYPES.AssignmentPattern:
      return [node.left, node.right];
    default:
      throw new Error(`Unhandled node type: ${(node as { type: string }).type}`);
  }
}
```

**Go Implementation:**
```go
func getLHSRHS(node *ast.Node) *lhsRhsPair {
    switch node.Kind {
    case ast.KindVariableDeclaration:
        // Returns name and initializer
    case ast.KindPropertyDeclaration:
        // Returns node and initializer
    case ast.KindParameter:
        // Returns name and initializer
    case ast.KindBindingElement:
        // Returns name and initializer
    default:
        return &lhsRhsPair{lhs: nil, rhs: nil}
    }
}
```

**Issue:** The Go version handles different AST node types than TypeScript and doesn't have an exact equivalent for AssignmentPattern handling.

**Impact:** The extraction of left-hand side and right-hand side values may be incorrect for some node types.

**Test Coverage:** Various declaration patterns in tests will reveal mismatches.

#### 6. getIDToAttachAnnotation Computed Property Handling
**TypeScript Implementation:**
```typescript
if (!node.computed) {
  return node.key;
}
// If the property's computed, we have to attach the
// annotation after the square bracket, not the enclosed expression
return nullThrows(
  context.sourceCode.getTokenAfter(node.key),
  NullThrowsReasons.MissingToken(']', 'key'),
);
```

**Go Implementation:**
```go
if propDecl.Name().Kind == ast.KindComputedPropertyName {
    // For computed properties, find the closing bracket token to match TypeScript behavior
    computed := propDecl.Name().AsComputedPropertyName()
    if computed.Expression != nil {
        // Use scanner to find the closing bracket after the expression
        s := scanner.GetScannerForSourceFile(ctx.SourceFile, computed.Expression.End())
        for s.TokenStart() < ctx.SourceFile.End() {
            if s.Token() == ast.KindCloseBracketToken {
                // For now, use the computed property node
                // TODO: Better position handling for after closing bracket
                return computed.AsNode() 
            }
            // ... scanner logic
        }
    }
    return propDecl.Name()
}
```

**Issue:** The Go version has a TODO comment indicating incomplete handling of computed property annotation attachment.

**Impact:** Type annotation attachment for computed properties like `[key]: Type = new Constructor()` may not work correctly.

**Test Coverage:** Tests with computed properties will reveal this issue.

### Recommendations
- Implement proper AssignmentPattern support to match TypeScript's selector behavior exactly
- Fix the isolatedDeclarations option access to properly read from parser options
- Complete the computed property name handling in getIDToAttachAnnotation
- Enhance comment preservation logic to match TypeScript's sophisticated approach
- Add comprehensive test coverage for edge cases like complex destructuring patterns
- Verify that the binding element processing correctly handles all function parameter scenarios
- Consider adding debug logging to compare AST traversal with TypeScript version

---