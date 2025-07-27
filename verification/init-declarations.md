## Rule: init-declarations

### Test File: init-declarations.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic mode handling ("always" vs "never")
  - Option parsing for arrays and strings
  - Variable declaration identification
  - TypeScript ambient declaration handling
  - For-in/for-of loop special cases
  - Error message formatting with variable names
  - Report location handling for identifiers
  - `ignoreForLoopInit` option support

- ⚠️ **POTENTIAL ISSUES**: 
  - Context override logic differs significantly from TypeScript implementation
  - For loop variable declaration handling may have edge cases
  - Namespace and module detection may need refinement

- ❌ **INCORRECT**: 
  - Missing `VariableDeclaration:exit` pattern matching from TypeScript implementation
  - Different AST traversal approach may cause subtle behavioral differences

### Discrepancies Found

#### 1. AST Pattern Matching Strategy Difference
**TypeScript Implementation:**
```typescript
return {
  'VariableDeclaration:exit'(node: TSESTree.VariableDeclaration): void {
    if (mode === 'always') {
      if (node.declare) {
        return;
      }
      if (isAncestorNamespaceDeclared(node)) {
        return;
      }
    }
    rules['VariableDeclaration:exit'](node);
  },
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
  ast.KindVariableStatement: func(node *ast.Node) {
    varStmt := node.AsVariableStatement()
    if varStmt.DeclarationList == nil {
      return
    }
    varDeclList := varStmt.DeclarationList.AsVariableDeclarationList()
    handleVarDeclList(varDeclList, node)
  },
  ast.KindVariableDeclarationList: func(node *ast.Node) {
    // Handle for loop cases
  },
}
```

**Issue:** The TypeScript version listens to `VariableDeclaration:exit` events and uses the base ESLint rule with a custom context override, while the Go version directly processes `VariableStatement` and `VariableDeclarationList` nodes. This fundamental difference could lead to different behavior in edge cases.

**Impact:** The Go implementation may miss or double-process certain variable declarations that the TypeScript version handles through the base rule delegation.

**Test Coverage:** This affects all test cases, particularly complex nested declarations.

#### 2. Context Override vs Direct Implementation
**TypeScript Implementation:**
```typescript
function getBaseContextOverride(): typeof context {
  const reportOverride: typeof context.report = descriptor => {
    if ('node' in descriptor && descriptor.loc == null) {
      const { node, ...rest } = descriptor;
      if (
        node.type === AST_NODE_TYPES.VariableDeclarator &&
        node.init == null
      ) {
        context.report({
          ...rest,
          loc: getReportLoc(node),
        });
        return;
      }
    }
    context.report(descriptor);
  };
  // Custom proxy logic for context override
}
```

**Go Implementation:**
```go
getReportLoc := func(node *ast.Node) core.TextRange {
  declarator := node.AsVariableDeclaration()
  if declarator.Name().Kind == ast.KindIdentifier {
    identifier := declarator.Name()
    return utils.TrimNodeTextRange(ctx.SourceFile, identifier)
  }
  return utils.TrimNodeTextRange(ctx.SourceFile, node)
}
```

**Issue:** The TypeScript version overrides the context to adjust reporting locations specifically for uninitialized variable declarators, while the Go version implements its own location calculation. The TypeScript approach is more sophisticated and handles edge cases with type annotations.

**Impact:** Different reporting locations, especially for variables with type annotations.

**Test Coverage:** All error cases where location precision matters, particularly TypeScript-specific syntax.

#### 3. Base Rule Delegation vs Full Reimplementation
**TypeScript Implementation:**
```typescript
const baseRule = getESLintCoreRule('init-declarations');
// ... uses baseRule.create(getBaseContextOverride())
```

**Go Implementation:**
```go
// Full custom implementation without base rule delegation
```

**Issue:** The TypeScript version leverages the existing ESLint core rule and only customizes behavior for TypeScript-specific features, while the Go version is a complete reimplementation. This could lead to missing edge cases handled by the base ESLint rule.

**Impact:** Potential missing functionality that exists in the base ESLint rule but isn't reimplemented in Go.

**Test Coverage:** Complex JavaScript patterns that the base ESLint rule handles but may not be covered in current tests.

#### 4. For Loop Handling Differences
**TypeScript Implementation:**
```typescript
// Relies on base rule for for-loop handling with TypeScript-specific overrides
```

**Go Implementation:**
```go
isInForLoopInit := func(node *ast.Node) bool {
  // Complex logic to detect for loop contexts
  // Handles both direct parent and VariableDeclarationList cases
}
```

**Issue:** The Go implementation has custom logic for detecting for-loop contexts, which may not perfectly match the base ESLint rule's behavior.

**Impact:** Different behavior for for-loop variable declarations, especially with the `ignoreForLoopInit` option.

**Test Coverage:** Test cases with `ignoreForLoopInit` option and complex for-loop patterns.

#### 5. Ambient Declaration Detection
**TypeScript Implementation:**
```typescript
if (isAncestorNamespaceDeclared(node)) {
  return;
}
// Only applies in "always" mode
```

**Go Implementation:**
```go
if isAncestorNamespaceDeclared(parentNode) {
  return;
}
// Applies regardless of mode
```

**Issue:** The TypeScript version only checks for ancestor namespace declarations in "always" mode, while the Go version checks it regardless of mode.

**Impact:** Different behavior for variables in declare namespaces when using "never" mode.

**Test Coverage:** Test cases with declare namespaces in both "always" and "never" modes.

### Recommendations
- Consider adopting the TypeScript approach of delegating to a base rule implementation for better compatibility
- Implement proper context override logic to match TypeScript's location reporting behavior
- Ensure ambient declaration detection logic matches the TypeScript version's mode-specific behavior
- Add more comprehensive test cases for edge cases around for-loop handling
- Verify that all base ESLint rule functionality is properly reimplemented in the Go version
- Test complex nested scenarios and TypeScript-specific syntax more thoroughly

---