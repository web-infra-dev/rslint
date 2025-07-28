## Rule: no-this-alias

### Test File: no-this-alias.test.ts

### Validation Summary
- ✅ **CORRECT**: Default options matching, message text consistency, basic identifier checking, allowedNames filtering
- ⚠️ **POTENTIAL ISSUES**: AST listener pattern differences, variable declaration handling approach
- ❌ **INCORRECT**: Missing VariableDeclarator handling, incorrect assignment expression detection, incomplete AST pattern coverage

### Discrepancies Found

#### 1. Incorrect AST Pattern Matching
**TypeScript Implementation:**
```typescript
"VariableDeclarator[init.type='ThisExpression'], AssignmentExpression[right.type='ThisExpression']"(
  node: TSESTree.AssignmentExpression | TSESTree.VariableDeclarator,
): void {
  const id = node.type === AST_NODE_TYPES.VariableDeclarator ? node.id : node.left;
  // ...
}
```

**Go Implementation:**
```go
return rule.RuleListeners{
  ast.KindVariableDeclaration: func(node *ast.Node) {
    decl := node.AsVariableDeclaration()
    checkNode(node, decl.Name(), decl.Initializer)
  },
  ast.KindBinaryExpression: func(node *ast.Node) {
    expr := node.AsBinaryExpression()
    if expr.OperatorToken.Kind == ast.KindEqualsToken {
      checkNode(node, expr.Left, expr.Right)
    }
  },
}
```

**Issue:** The Go implementation listens to `KindVariableDeclaration` instead of `KindVariableDeclarator`. The TypeScript version specifically targets the individual declarators within a variable declaration, not the declaration itself.

**Impact:** This causes the rule to miss individual variable declarators and may not handle cases like `const a = this, b = other;` correctly.

**Test Coverage:** This affects most test cases with variable declarations.

#### 2. Assignment Expression Detection Mismatch
**TypeScript Implementation:**
```typescript
// Uses CSS-style selector to match AssignmentExpression nodes
"AssignmentExpression[right.type='ThisExpression']"
```

**Go Implementation:**
```go
ast.KindBinaryExpression: func(node *ast.Node) {
  expr := node.AsBinaryExpression()
  if expr.OperatorToken.Kind == ast.KindEqualsToken {
    checkNode(node, expr.Left, expr.Right)
  }
},
```

**Issue:** Assignment expressions in TypeScript AST are represented as `AssignmentExpression` nodes, but the Go implementation looks for `BinaryExpression` with `EqualsToken`. This is a fundamental AST node type mismatch.

**Impact:** Assignment patterns like `that = this;` may not be detected correctly, causing the rule to miss violations.

**Test Coverage:** Test case `let that; that = this;` would fail to trigger the rule.

#### 3. Variable Declaration Structure Handling
**TypeScript Implementation:**
```typescript
const id = node.type === AST_NODE_TYPES.VariableDeclarator ? node.id : node.left;
```

**Go Implementation:**
```go
ast.KindVariableDeclaration: func(node *ast.Node) {
  decl := node.AsVariableDeclaration()
  checkNode(node, decl.Name(), decl.Initializer)
}
```

**Issue:** The Go implementation assumes a single declarator per declaration by calling `decl.Name()` and `decl.Initializer`, but variable declarations can have multiple declarators (e.g., `const a = this, b = other;`).

**Impact:** Only the first declarator in a multi-declarator statement would be checked.

**Test Coverage:** Would miss violations in complex variable declaration statements.

#### 4. Missing AST Node Kind Mappings
**TypeScript Implementation:**
```typescript
// Handles both VariableDeclarator and AssignmentExpression
node.type === AST_NODE_TYPES.VariableDeclarator ? node.id : node.left
```

**Go Implementation:**
```go
// Only checks if id.Kind != ast.KindIdentifier for destructuring
if id.Kind != ast.KindIdentifier
```

**Issue:** The Go implementation doesn't properly map TypeScript AST node types to Go AST node kinds. Destructuring patterns might not be detected correctly.

**Impact:** Object and array destructuring patterns (`const { props } = this` or `const [foo] = this`) may not be properly identified.

**Test Coverage:** Destructuring test cases may produce incorrect results or miss violations.

#### 5. Default Option Values Mismatch
**TypeScript Implementation:**
```typescript
defaultOptions: [
  {
    allowDestructuring: true,
    allowedNames: [],
  },
],
```

**Go Implementation:**
```go
opts := NoThisAliasOptions{
  AllowDestructuring: true,
  AllowedNames:       []string{},
}
```

**Issue:** While the default values match, this is correct behavior.

**Impact:** No impact - this is implemented correctly.

**Test Coverage:** Default behavior tests should pass.

### Recommendations
- Replace `ast.KindVariableDeclaration` listener with `ast.KindVariableDeclarator` to match individual declarators
- Replace `ast.KindBinaryExpression` listener with the appropriate AST node kind for assignment expressions
- Implement proper handling of multiple declarators within a single variable declaration
- Verify AST node kind mappings for destructuring patterns (ObjectPattern, ArrayPattern equivalents)
- Add comprehensive AST debugging to understand the actual node structure in the Go typescript-go library
- Consider using selector-style pattern matching similar to the TypeScript implementation if available in the Go rule framework

---