## Rule: no-non-null-asserted-nullish-coalescing

### Test File: no-non-null-asserted-nullish-coalescing.test.ts

### Validation Summary
- ✅ **CORRECT**: Rule message IDs and descriptions match, basic AST pattern matching for nullish coalescing with non-null assertion, suggestion mechanism with fix to remove exclamation token
- ⚠️ **POTENTIAL ISSUES**: Assignment detection logic implementation differs significantly, scope analysis approach varies between implementations, exclamation token positioning logic is simplified
- ❌ **INCORRECT**: Missing proper scope-based variable tracking, simplified string-based assignment detection may miss complex cases, AST node listener targets wrong node type

### Discrepancies Found

#### 1. AST Node Listener Target Mismatch
**TypeScript Implementation:**
```typescript
'LogicalExpression[operator = "??"] > TSNonNullExpression.left'(
  node: TSESTree.TSNonNullExpression,
): void {
```

**Go Implementation:**
```go
ast.KindBinaryExpression: func(node *ast.Node) {
  // ... checks for nullish coalescing operator
  // ... checks for non-null assertion on left operand
```

**Issue:** The TypeScript version uses a specific CSS-like selector to target non-null expressions that are the left operand of nullish coalescing operators. The Go version listens to all binary expressions and manually checks conditions.

**Impact:** The Go approach is functionally equivalent but less precise in targeting. It may process more nodes unnecessarily.

**Test Coverage:** All test cases should still pass, but performance may be impacted.

#### 2. Variable Assignment Detection Logic
**TypeScript Implementation:**
```typescript
function hasAssignmentBeforeNode(
  variable: TSESLint.Scope.Variable,
  node: TSESTree.Node,
): boolean {
  return (
    variable.references.some(
      ref => ref.isWrite() && ref.identifier.range[1] < node.range[1],
    ) ||
    variable.defs.some(
      def =>
        isDefinitionWithAssignment(def) && def.node.range[1] < node.range[1],
    )
  );
}
```

**Go Implementation:**
```go
hasAssignmentBeforeNode := func(identifier *ast.Identifier, node *ast.Node) bool {
  // Uses TypeChecker.GetSymbolAtLocation()
  // Checks symbol declarations for initializers
  // Falls back to simple string-based pattern matching
  // Look for patterns like "varName =" or "varName=" in the code before this node
  assignmentPattern := varName + " ="
  assignmentPattern2 := varName + "="
  
  if strings.Contains(beforeCode, assignmentPattern) || strings.Contains(beforeCode, assignmentPattern2) {
    return true
  }
}
```

**Issue:** The Go implementation lacks proper scope analysis and reference tracking. It uses a simplified string-based approach that may produce false positives/negatives.

**Impact:** May incorrectly identify assignments, leading to missed violations or false positives. The string-based matching could match variable names in comments, strings, or unrelated contexts.

**Test Coverage:** Valid test cases like `let x: string; x! ?? '';` depend on this logic working correctly.

#### 3. Scope Analysis Implementation
**TypeScript Implementation:**
```typescript
const scope = context.sourceCode.getScope(node);
const variable = ASTUtils.findVariable(scope, identifier.name);
```

**Go Implementation:**
```go
symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier.AsNode())
if symbol == nil {
  // If we can't find the symbol, assume it's assigned (external variable, etc.)
  return true
}
```

**Issue:** The TypeScript version uses proper scope analysis to find variables within the current scope. The Go version uses TypeScript's symbol resolution but lacks the sophisticated scope traversal.

**Impact:** May not correctly handle variables in different scopes, leading to incorrect behavior for nested functions or block scopes.

**Test Coverage:** Test cases with function scopes and variable shadowing may be affected.

#### 4. Definition Type Checking
**TypeScript Implementation:**
```typescript
function isDefinitionWithAssignment(definition: Definition): boolean {
  if (definition.type !== DefinitionType.Variable) {
    return false;
  }
  const variableDeclarator = definition.node;
  return variableDeclarator.definite || variableDeclarator.init != null;
}
```

**Go Implementation:**
```go
if ast.IsVariableDeclaration(decl) {
  varDecl := decl.AsVariableDeclaration()
  // Check if it has an initializer or is definitely assigned
  if varDecl.Initializer != nil || (varDecl.ExclamationToken != nil && varDecl.ExclamationToken.Kind == ast.KindExclamationToken) {
```

**Issue:** The Go version attempts to replicate the definite assignment check but the logic for checking the exclamation token may not be equivalent to the TypeScript `definite` property.

**Impact:** May not correctly identify definitely assigned variables (like `let x!: string;`).

**Test Coverage:** Test case `let x!: string; x ?? '';` specifically tests this logic.

#### 5. Reference Tracking Missing
**TypeScript Implementation:**
```typescript
variable.references.some(
  ref => ref.isWrite() && ref.identifier.range[1] < node.range[1],
)
```

**Go Implementation:**
```go
// No equivalent reference tracking implementation
// Falls back to string-based pattern matching
```

**Issue:** The Go implementation completely lacks reference tracking to find write operations to variables before the current node.

**Impact:** Cannot properly detect assignments like `x = foo(); x! ?? '';` without relying on the unreliable string matching.

**Test Coverage:** Multiple test cases involve variable assignments that this logic should detect.

#### 6. Exclamation Token Position Detection
**TypeScript Implementation:**
```typescript
const exclamationMark = nullThrows(
  context.sourceCode.getLastToken(
    node,
    ASTUtils.isNonNullAssertionPunctuator,
  ),
  NullThrowsReasons.MissingToken('!', 'Non-null Assertion'),
);
return fixer.remove(exclamationMark);
```

**Go Implementation:**
```go
// Find the '!' character position by scanning from the expression end
exclamationStart := exprEnd
exclamationEnd := leftEnd

// Scan to find the actual '!' character (skip whitespace)
for i := exprEnd; i < leftEnd; i++ {
  if sourceText[i] == '!' {
    exclamationStart = i
    exclamationEnd = i + 1
    break
  }
}
```

**Issue:** The Go implementation uses manual character scanning instead of proper token-based detection.

**Impact:** May not handle all edge cases with whitespace or complex expressions. The TypeScript version is more robust.

**Test Coverage:** The test case with spacing `x  ! ?? '';` specifically tests this logic.

### Recommendations
- Implement proper scope analysis and variable reference tracking instead of string-based pattern matching
- Add comprehensive symbol resolution to handle variable assignments across different scopes
- Improve exclamation token detection using proper AST token navigation
- Consider using a more targeted AST listener approach similar to the TypeScript selector pattern
- Add proper handling of definite assignment assertions (`let x!: string;`)
- Implement reference tracking to detect write operations to variables before the current node
- Replace string-based assignment detection with proper AST traversal and analysis

---