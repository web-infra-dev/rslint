# Rule Validation: no-loop-func

## Rule: no-loop-func

### Test File: no-loop-func.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic loop detection, function type handling, IIFE detection pattern, error message structure
- ⚠️ **POTENTIAL ISSUES**: Simplified variable safety analysis, incomplete scope reference handling, missing async/generator checks
- ❌ **INCORRECT**: Variable declaration kind detection, reference analysis implementation, IIFE reference checking

### Discrepancies Found

#### 1. Variable Declaration Kind Detection
**TypeScript Implementation:**
```typescript
const kind =
  declaration?.type === AST_NODE_TYPES.VariableDeclaration
    ? declaration.kind
    : '';
```

**Go Implementation:**
```go
kind := ""
if declaration.Parent != nil && declaration.Parent.Kind == ast.KindVariableDeclaration {
    varDecl := declaration.Parent.AsVariableDeclaration()
    flags := varDecl.Flags
    if flags&ast.NodeFlagsConst != 0 {
        kind = "const"
    } else if flags&ast.NodeFlagsLet != 0 {
        kind = "let"
    } else {
        kind = "var"
    }
}
```

**Issue:** The Go implementation extracts declaration kind from flags, but TypeScript-ESLint uses the `kind` property directly. The flag-based approach may not correctly identify the declaration type.

**Impact:** Variables may be incorrectly classified as unsafe when they should be safe (const variables) or vice versa.

**Test Coverage:** Test cases with `const` and `let` declarations will reveal this issue.

#### 2. Reference Safety Analysis
**TypeScript Implementation:**
```typescript
function isSafe(
  loopNode: TSESTree.Node,
  reference: TSESLint.Scope.Reference,
): boolean {
  // Complex analysis including:
  // - Type reference checking
  // - Variable scope analysis
  // - Write reference border checking
  // - Reference timing analysis
  return variable?.references.every(isSafeReference) ?? false;
}
```

**Go Implementation:**
```go
func isSafe(loopNode *ast.Node, reference *ast.Symbol, variable *ast.Symbol, ctx rule.RuleContext) bool {
    // Simplified analysis that only checks:
    // - const variables (safe)
    // - let variables declared in loop (safe)
    // - everything else (unsafe)
    return false
}
```

**Issue:** The Go implementation is overly conservative and missing the sophisticated reference analysis from TypeScript. It doesn't check for write references after borders or perform proper scope analysis.

**Impact:** Many valid cases will be flagged as unsafe, causing false positives.

**Test Coverage:** Most test cases involving complex variable references will fail.

#### 3. Async/Generator Function Handling
**TypeScript Implementation:**
```typescript
if (!(node.async || node.generator) && isIIFE(node)) {
  // IIFE handling logic
}
```

**Go Implementation:**
```go
isGenerator := false
switch node.Kind {
case ast.KindFunctionDeclaration:
    fn := node.AsFunctionDeclaration()
    isGenerator = fn.AsteriskToken != nil
case ast.KindFunctionExpression:
    fn := node.AsFunctionExpression()
    isGenerator = fn.AsteriskToken != nil
}

if !isGenerator && isIIFE(node) {
    // Missing async check
}
```

**Issue:** The Go implementation doesn't check for async functions, only generator functions. Async functions should also be excluded from IIFE skipping.

**Impact:** Async IIFE functions may be incorrectly skipped when they should be analyzed.

**Test Coverage:** Test cases with async IIFE functions would reveal this issue.

#### 4. IIFE Reference Checking
**TypeScript Implementation:**
```typescript
const isFunctionReferenced =
  isFunctionExpression && node.id
    ? references.some(r => r.identifier.name === node.id?.name)
    : false;
```

**Go Implementation:**
```go
isFunctionReferenced := false
if isFunctionExpression {
    funcExpr := node.AsFunctionExpression()
    if funcExpr.Name() != nil && ast.IsIdentifier(funcExpr.Name()) {
        // For simplicity, we'll assume named function expressions might be referenced
        // A more accurate check would require full scope analysis
        isFunctionReferenced = true
    }
}
```

**Issue:** The Go implementation uses a simplistic assumption instead of actually checking if the function is referenced elsewhere in the scope.

**Impact:** Named function expressions in IIFE patterns may be incorrectly analyzed.

**Test Coverage:** Test cases with named function expressions in IIFE patterns.

#### 5. Variable Reference Collection
**TypeScript Implementation:**
```typescript
const references = context.sourceCode.getScope(node).through;
const unsafeRefs = references
  .filter(r => r.resolved && !isSafe(loopNode, r))
  .map(r => r.identifier.name);
```

**Go Implementation:**
```go
func getUnsafeRefs(node *ast.Node, loopNode *ast.Node, ctx rule.RuleContext) []string {
    // Manual AST traversal to find identifiers
    // Missing proper scope analysis
    // Using simplified variable checking
}
```

**Issue:** The Go implementation manually traverses the AST to find variable references instead of using proper scope analysis. This misses the sophisticated reference resolution from TypeScript-ESLint.

**Impact:** Variable references may be missed or incorrectly identified, leading to both false positives and false negatives.

**Test Coverage:** Complex test cases with nested scopes and variable shadowing will fail.

#### 6. Top Loop Analysis
**TypeScript Implementation:**
```typescript
function getTopLoopNode(
  node: TSESTree.Node,
  excludedNode: TSESTree.Node | null | undefined,
): TSESTree.Node {
  const border = excludedNode ? excludedNode.range[1] : 0;
  let retv = node;
  let containingLoopNode: TSESTree.Node | null = node;

  while (containingLoopNode && containingLoopNode.range[0] >= border) {
    retv = containingLoopNode;
    containingLoopNode = getContainingLoopNode(containingLoopNode);
  }

  return retv;
}
```

**Go Implementation:**
```go
func getTopLoopNode(node *ast.Node, excludedNode *ast.Node) *ast.Node {
    // Note: topLoop analysis removed for simplicity
    // This functionality is not fully implemented
}
```

**Issue:** The Go implementation has removed the top loop analysis, which is crucial for determining the correct border for reference safety checking.

**Impact:** Complex nested loop scenarios may not be handled correctly.

**Test Coverage:** Test cases with nested loops and complex variable scoping.

#### 7. Message ID Handling
**TypeScript Implementation:**
```typescript
context.report({
  node,
  messageId: 'unsafeRefs',
  data: { varNames: `'${unsafeRefs.join("', '")}'` },
});
```

**Go Implementation:**
```go
func buildUnsafeRefsMessage(varNames []string) rule.RuleMessage {
    quotedNames := make([]string, len(varNames))
    for i, name := range varNames {
        quotedNames[i] = fmt.Sprintf("'%s'", name)
    }
    return rule.RuleMessage{
        Id:          "unsafeRefs",
        Description: fmt.Sprintf("Function declared in a loop contains unsafe references to variable(s) %s.", strings.Join(quotedNames, ", ")),
    }
}
```

**Issue:** The Go implementation hardcodes the message description instead of using message IDs with interpolation like the TypeScript version.

**Impact:** Message formatting may not match exactly, and localization support is reduced.

**Test Coverage:** Error message format validation in tests.

### Recommendations
- Implement proper scope analysis to match TypeScript-ESLint's reference resolution
- Fix variable declaration kind detection to use proper AST properties instead of flags
- Add async function detection alongside generator function detection
- Implement proper IIFE reference checking using scope analysis
- Restore and implement the top loop analysis for complex nested scenarios
- Add comprehensive variable safety analysis including write reference timing
- Implement proper message ID system with interpolation
- Add test cases specifically for the edge cases identified above
- Consider using a more sophisticated AST traversal approach that matches TypeScript-ESLint's scope management

---