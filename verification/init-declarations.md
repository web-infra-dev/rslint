## Rule: init-declarations

### Test File: init-declarations.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic rule structure and configuration parsing
  - Handling of "always" and "never" modes
  - Support for `ignoreForLoopInit` option
  - Ambient/declare context detection
  - Identifier-only reporting (skips destructuring patterns)
  - For-in/for-of loop variable handling
  - TypeScript-specific features (namespaces, declare statements)

- ⚠️ **POTENTIAL ISSUES**:
  - Different AST event handling pattern may cause edge cases
  - Report location calculation differences
  - Debug output left in production code

- ❌ **INCORRECT**:
  - Missing base rule context override functionality
  - Inconsistent message IDs with TypeScript-ESLint

### Discrepancies Found

#### 1. Base Rule Context Override Missing
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
  // ... proxy setup
}
```

**Go Implementation:**
```go
// Missing equivalent functionality
getReportLoc := func(node *ast.Node) core.TextRange {
  // Basic implementation without override pattern
  declarator := node.AsVariableDeclaration()
  if declarator.Name().Kind == ast.KindIdentifier {
    identifier := declarator.Name()
    return utils.TrimNodeTextRange(ctx.SourceFile, identifier)
  }
  return utils.TrimNodeTextRange(ctx.SourceFile, node)
}
```

**Issue:** The Go implementation lacks the sophisticated context override mechanism that TypeScript uses to customize report locations specifically for uninitialized variables.

**Impact:** May lead to different highlighting/error positioning in the editor, especially for variables with type annotations.

**Test Coverage:** All test cases with location expectations rely on this functionality.

#### 2. Message ID Mismatch
**TypeScript Implementation:**
```typescript
// Uses base rule's message IDs from ESLint core
messages: baseRule.meta.messages,
// Expected message IDs: 'initialized', 'notInitialized'
```

**Go Implementation:**
```go
ctx.ReportRange(getReportLoc(decl), rule.RuleMessage{
  Id:          "initialized",
  Description: fmt.Sprintf("Variable '%s' should be initialized at declaration.", idName),
})
// ...
ctx.ReportRange(getReportLoc(decl), rule.RuleMessage{
  Id:          "notInitialized", 
  Description: fmt.Sprintf("Variable '%s' should not be initialized.", idName),
})
```

**Issue:** The Go implementation hardcodes message IDs and descriptions instead of using the base ESLint rule's messages, which may lead to inconsistencies.

**Impact:** Different error messages and message IDs compared to ESLint/TypeScript-ESLint.

**Test Coverage:** All test cases expect specific message IDs that may not match.

#### 3. Event Handler Pattern Differences
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
    // Handle VariableStatement
  },
  ast.KindVariableDeclarationList: func(node *ast.Node) {
    // Handle VariableDeclarationList in for loops
  },
}
```

**Issue:** The TypeScript version uses `:exit` event and delegates to base rule, while Go implementation processes AST nodes directly with different node types.

**Impact:** May miss edge cases or handle node traversal differently than the base rule.

**Test Coverage:** Complex nested scenarios and edge cases may behave differently.

#### 4. Debug Code in Production
**TypeScript Implementation:**
```typescript
// No debug output in production code
```

**Go Implementation:**
```go
// Debug info
if varDeclList.Parent != nil && varDeclList.Parent.Kind == ast.KindForStatement {
  fmt.Printf("DEBUG: In for loop, ignoreForLoopInit=%v, isForLoopInit=%v, mode=%s\n", opts.IgnoreForLoopInit, isForLoopInit, opts.Mode)
}
```

**Issue:** Debug print statements are left in production code.

**Impact:** Unwanted console output in production usage.

**Test Coverage:** Any for-loop test cases will produce debug output.

#### 5. Type Annotation Handling in Report Location
**TypeScript Implementation:**
```typescript
function getReportLoc(node: TSESTree.VariableDeclarator): TSESTree.SourceLocation {
  const start: TSESTree.Position = structuredClone(node.loc.start);
  const end: TSESTree.Position = {
    line: node.loc.start.line,
    column: node.loc.start.column + (node.id as TSESTree.Identifier).name.length,
  };
  return { start, end };
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

**Issue:** The TypeScript version specifically calculates the range to exclude type annotations, while the Go version relies on `utils.TrimNodeTextRange` which may not handle this precisely.

**Impact:** Different highlighting ranges, especially for variables with type annotations like `let arr: string;`.

**Test Coverage:** Test cases with type annotations may show different column positions.

### Recommendations
- Remove debug print statements from production code (lines 145-148)
- Implement proper message ID compatibility with base ESLint rule
- Consider implementing a context override pattern similar to TypeScript version
- Verify that `utils.TrimNodeTextRange` properly handles type annotation exclusion
- Add comprehensive tests for edge cases around AST node handling differences
- Ensure report location calculation matches TypeScript-ESLint exactly for type-annotated variables

---