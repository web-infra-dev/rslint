## Rule: no-confusing-non-null-assertion

### Test File: no-confusing-non-null-assertion.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic operator identification, message structure and IDs, general rule intention
- ⚠️ **POTENTIAL ISSUES**: Token scanning approach, fix generation logic for wrapping
- ❌ **INCORRECT**: Missing AssignmentExpression handling, flawed AST node processing, incorrect token analysis, broken fix generation

### Discrepancies Found

#### 1. Missing AssignmentExpression Support
**TypeScript Implementation:**
```typescript
'BinaryExpression, AssignmentExpression'(
  node: TSESTree.AssignmentExpression | TSESTree.BinaryExpression,
): void {
```

**Go Implementation:**
```go
return rule.RuleListeners{
    ast.KindBinaryExpression: checkNode,
}
```

**Issue:** The Go implementation only listens for `BinaryExpression` nodes but completely ignores `AssignmentExpression` nodes, which are crucial for detecting patterns like `a! = b`.

**Impact:** Assignment expressions with confusing non-null assertions will not be detected, causing the rule to miss critical cases.

**Test Coverage:** Test cases like `'a! = b;'` and `'(obj = new new OuterObj().InnerObj).Name! = c;'` will fail.

#### 2. Flawed AST Node Processing Logic
**TypeScript Implementation:**
```typescript
if (node.Kind == ast.KindBinaryExpression) {
    binaryExpr := node.AsBinaryExpression()
    operator = binaryExpr.OperatorToken.Kind
    left = binaryExpr.Left
    operatorToken = binaryExpr.OperatorToken.Kind
} else if node.Kind == ast.KindBinaryExpression {
    assignExpr := node.AsBinaryExpression()
    operator = assignExpr.OperatorToken.Kind
    left = assignExpr.Left
    operatorToken = assignExpr.OperatorToken.Kind
}
```

**Go Implementation:**
```go
// Both conditions check for the same node type!
if (node.Kind == ast.KindBinaryExpression) {
    // ...
} else if node.Kind == ast.KindBinaryExpression {
    // This branch will never execute
}
```

**Issue:** The second condition should check for `ast.KindBinaryExpression` but uses the wrong type, and both branches incorrectly call `AsBinaryExpression()`.

**Impact:** Assignment expressions will never be processed even if the listener was added.

**Test Coverage:** All assignment-related test cases will fail.

#### 3. Primitive Token Analysis
**TypeScript Implementation:**
```typescript
const leftHandFinalToken = context.sourceCode.getLastToken(node.left);
const tokenAfterLeft = context.sourceCode.getTokenAfter(node.left);
if (
  leftHandFinalToken?.type === AST_TOKEN_TYPES.Punctuator &&
  leftHandFinalToken.value === '!' &&
  tokenAfterLeft?.value !== ')'
) {
```

**Go Implementation:**
```go
leftText := string(sourceFile.Text()[leftRange.Pos():leftRange.End()])
if !strings.HasSuffix(leftText, "!") {
    return
}
```

**Issue:** The Go implementation uses naive string suffix checking instead of proper token analysis. This cannot distinguish between a non-null assertion and an exclamation mark that might be part of a string or comment.

**Impact:** False positives on strings containing "!" and false negatives on complex expressions.

**Test Coverage:** Could cause issues with edge cases involving strings or comments.

#### 4. Incorrect Fix Generation for Wrapping
**TypeScript Implementation:**
```typescript
function wrapUpLeftFixer(
  node: TSESTree.AssignmentExpression | TSESTree.BinaryExpression,
): TSESLint.ReportFixFunction {
  return (fixer): TSESLint.RuleFix[] => [
    fixer.insertTextBefore(node.left, '('),
    fixer.insertTextAfter(node.left, ')'),
  ];
}
```

**Go Implementation:**
```go
func wrapUpLeftFixes(sourceFile *ast.SourceFile, left *ast.Node, exclamationPos int) []rule.RuleFix {
    return []rule.RuleFix{
        rule.RuleFixInsertBefore(sourceFile, left, "("),
        rule.RuleFixReplaceRange(core.NewTextRange(exclamationPos, exclamationPos+1), ")!"),
    }
}
```

**Issue:** The Go implementation incorrectly replaces the exclamation mark with ")!" instead of simply inserting parentheses around the left expression.

**Impact:** Suggested fixes will be incorrect, potentially changing code semantics.

**Test Coverage:** All suggestion outputs will be wrong for wrapping fixes.

#### 5. Incomplete Parenthesis Check
**TypeScript Implementation:**
```typescript
const tokenAfterLeft = context.sourceCode.getTokenAfter(node.left);
if (
  // ... &&
  tokenAfterLeft?.value !== ')'
) {
```

**Go Implementation:**
```go
s := scanner.GetScannerForSourceFile(sourceFile, leftRange.End())
s.Scan()
if s.Token() == ast.KindCloseParenToken {
    return
}
```

**Issue:** The Go implementation checks for closing parenthesis but the scanning logic may not correctly position the scanner after the left expression.

**Impact:** May miss or incorrectly identify parenthesis patterns.

**Test Coverage:** Could affect valid cases that should not trigger the rule.

#### 6. Missing Non-Null Expression Detection
**TypeScript Implementation:**
```typescript
if (node.left.type === AST_NODE_TYPES.TSNonNullExpression) {
  // Provide specific suggestions for removing the assertion
}
```

**Go Implementation:**
```go
if left.Kind == ast.KindNonNullExpression {
  // Similar logic but with the flawed token analysis
}
```

**Issue:** While the Go implementation attempts to check for non-null expressions, the underlying token analysis is flawed, making this check unreliable.

**Impact:** Suggestions may not be correctly categorized between removal and wrapping.

**Test Coverage:** Affects the quality of suggested fixes in test cases.

### Recommendations
- **Fix AST node type handling**: Correct the condition to check for `AssignmentExpression` and use appropriate casting methods
- **Add AssignmentExpression listener**: Include `ast.KindAssignmentExpression` in the rule listeners
- **Implement proper token analysis**: Use the scanner API correctly to identify the last token and check for punctuators
- **Fix the wrapping fix generation**: Correct the logic to only insert parentheses without modifying the exclamation mark
- **Improve parenthesis detection**: Ensure the scanner is positioned correctly after the left expression
- **Add comprehensive token validation**: Verify that the detected "!" is actually a non-null assertion token, not part of other syntax

---