## Rule: no-non-null-asserted-nullish-coalescing

### Test File: no-non-null-asserted-nullish-coalescing.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic nullish coalescing detection, message IDs match, suggestion mechanism exists, exclamation token removal logic
- ⚠️ **POTENTIAL ISSUES**: Assignment detection logic differs significantly, scope analysis approach varies, string-based assignment detection is fragile
- ❌ **INCORRECT**: Variable assignment detection implementation is overly simplistic and unreliable

### Discrepancies Found

#### 1. Variable Assignment Detection Logic
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
hasAssignmentBeforeNode := func(identifier *ast.Identifier, node *ast.Node) bool {
    // ... symbol checking logic ...
    
    // Check for assignment expressions with string matching
    sourceFile := ctx.SourceFile
    sourceText := sourceFile.Text()
    nodeStart := node.Pos()
    
    // Look for assignment patterns like "x = " before this node
    varName := identifier.Text
    beforeCode := sourceText[:nodeStart]
    
    // Simple regex-like check for assignment pattern
    assignmentPattern := varName + " ="
    assignmentPattern2 := varName + "="
    
    if strings.Contains(beforeCode, assignmentPattern) || strings.Contains(beforeCode, assignmentPattern2) {
        return true
    }
    
    return false
}
```

**Issue:** The Go implementation uses primitive string matching to detect assignments instead of proper scope analysis. This approach will produce false positives and miss complex assignment patterns.

**Impact:** 
- False positives: `let foo = 5; let otherFoo = 10; foo! ?? bar` would incorrectly match "foo =" from `otherFoo =`
- False negatives: Complex assignments like `[foo] = arr`, `({foo} = obj)`, or `foo += bar` won't be detected
- Incorrect handling of scoped variables with same names

**Test Coverage:** This affects test cases with variable assignments, particularly:
- `let x: string; x = foo(); x! ?? '';` (should trigger rule)
- `let x: string; x! ?? '';` (should not trigger rule)

#### 2. Scope Analysis Implementation
**TypeScript Implementation:**
```typescript
const scope = context.sourceCode.getScope(node);
const identifier = node.expression;
const variable = ASTUtils.findVariable(scope, identifier.name);
if (variable && !hasAssignmentBeforeNode(variable, node)) {
  return;
}
```

**Go Implementation:**
```go
// Get the symbol for the identifier
symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier.AsNode())
if symbol == nil {
    // If we can't find the symbol, assume it's assigned (external variable, etc.)
    return true
}
```

**Issue:** The Go version uses TypeScript's symbol system instead of ESLint's scope analysis. While functionally similar, the fallback behavior differs - TypeScript version continues analysis when variable is found, Go assumes assignment when symbol is missing.

**Impact:** Different behavior for edge cases involving undeclared variables or complex scoping scenarios.

**Test Coverage:** May affect test cases with undeclared variables or complex scoping.

#### 3. Definite Assignment Assertion Detection
**TypeScript Implementation:**
```typescript
const variableDeclarator = definition.node;
return variableDeclarator.definite || variableDeclarator.init != null;
```

**Go Implementation:**
```go
// Check if it has an initializer or is definitely assigned
if varDecl.Initializer != nil || (varDecl.ExclamationToken != nil && varDecl.ExclamationToken.Kind == ast.KindExclamationToken) {
```

**Issue:** The Go implementation checks for `ExclamationToken` to detect definite assignment assertions (`let x!: string`), which appears correct but the logic structure differs from TypeScript's `definite` property check.

**Impact:** Should work correctly but implementation approach differs. May have edge cases with complex declaration patterns.

**Test Coverage:** Affects test case `let x!: string; x! ?? '';` (should trigger rule).

#### 4. AST Traversal Pattern
**TypeScript Implementation:**
```typescript
'LogicalExpression[operator = "??"] > TSNonNullExpression.left'(
  node: TSESTree.TSNonNullExpression,
): void {
```

**Go Implementation:**
```go
ast.KindBinaryExpression: func(node *ast.Node) {
    // Manual checks for nullish coalescing and non-null assertion
    if binaryExpr.OperatorToken.Kind != ast.KindQuestionQuestionToken {
        return
    }
    if !ast.IsNonNullExpression(leftOperand) {
        return
    }
```

**Issue:** The Go version uses a broader AST listener (`BinaryExpression`) and manually filters, while TypeScript uses a specific selector. Both should work but performance characteristics differ.

**Impact:** Minimal functional impact, but Go version may process more nodes unnecessarily.

**Test Coverage:** Should handle all test cases correctly.

### Recommendations
- **Critical**: Replace string-based assignment detection with proper AST traversal to find assignment expressions and variable references
- **Important**: Implement proper scope analysis or ensure TypeScript symbol analysis correctly handles all edge cases
- **Minor**: Consider optimizing AST traversal pattern to match TypeScript's specificity
- **Testing**: Add test cases that would expose the string-matching weaknesses:
  - Variables with similar names: `let foo = 1; let fooBar = 2; foo! ?? bar`
  - Complex assignment patterns: `[x] = arr; x! ?? bar`
  - Nested scopes with same variable names

### Missing Functionality
- Proper reference analysis for write operations
- Handling of destructuring assignments
- Complex assignment operator support (+=, *=, etc.)
- Robust scope-aware variable tracking

---