## Rule: no-loop-func

### Test File: no-loop-func.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic loop detection (for, while, do-while, for-in, for-of)
  - IIFE tracking mechanism
  - AST node kind matching for function types
  - Type reference safety checking
  - Const variable safety handling
  - Let variable in-loop safety checking
  - Error message formatting with variable names

- ⚠️ **POTENTIAL ISSUES**: 
  - Overly conservative `isSafe` implementation
  - Missing async function detection
  - Simplified IIFE reference checking
  - Incomplete scope analysis for write references

- ❌ **INCORRECT**: 
  - Missing critical scope-based safety analysis
  - Incomplete variable reference tracking
  - Missing write reference analysis after border
  - Typo in struct name (`iifeTacker` instead of `iifeTracker`)

### Discrepancies Found

#### 1. Incomplete Safety Analysis in `isSafe` Function
**TypeScript Implementation:**
```typescript
function isSafe(
  loopNode: TSESTree.Node,
  reference: TSESLint.Scope.Reference,
): boolean {
  // ... complex analysis including:
  // - Variable scope checking
  // - Write reference analysis after border
  // - Top loop node border calculations
  return variable?.references.every(isSafeReference) ?? false;
}
```

**Go Implementation:**
```go
func isSafe(loopNode *ast.Node, reference *ast.Symbol, variable *ast.Symbol, ctx rule.RuleContext) bool {
  // ... basic checks then:
  // Note: topLoop analysis removed for simplicity
  // For now, we'll assume any non-const variable referenced in a loop function is potentially unsafe
  // This is a conservative approach
  return false
}
```

**Issue:** The Go implementation has a placeholder that always returns `false` for non-const/non-local-let variables, missing the sophisticated scope and write reference analysis.

**Impact:** Will produce false positives for safe variable references, flagging legitimate cases as unsafe.

**Test Coverage:** Many valid test cases will likely fail, particularly those involving variables declared outside loops that are never modified.

#### 2. Missing Async Function Detection
**TypeScript Implementation:**
```typescript
if (!(node.async || node.generator) && isIIFE(node)) {
  // IIFE logic
}
```

**Go Implementation:**
```go
// Check if this is an IIFE
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

**Issue:** The Go version doesn't check for async functions, only generators.

**Impact:** Async IIFEs might be incorrectly skipped when they should be analyzed.

**Test Coverage:** Any test cases with async IIFEs may behave differently.

#### 3. Simplified IIFE Reference Checking
**TypeScript Implementation:**
```typescript
const references = context.sourceCode.getScope(node).through;
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

**Issue:** The Go version uses a simplistic assumption rather than actual scope analysis to determine if a function is referenced.

**Impact:** May incorrectly treat some IIFEs as referenced when they're not, leading to false positives.

**Test Coverage:** IIFE-related test cases may produce different results.

#### 4. Missing Write Reference Analysis
**TypeScript Implementation:**
```typescript
const border = getTopLoopNode(
  loopNode,
  kind === 'let' ? declaration : null,
).range[0];

function isSafeReference(upperRef: TSESLint.Scope.Reference): boolean {
  const id = upperRef.identifier;
  return (
    !upperRef.isWrite() ||
    (variable?.scope.variableScope === upperRef.from.variableScope &&
      id.range[0] < border)
  );
}

return variable?.references.every(isSafeReference) ?? false;
```

**Go Implementation:**
```go
// Check for write references after the border
// Note: topLoop analysis removed for simplicity
// Check if there are any write references to this variable after the border
// For now, we'll assume any non-const variable referenced in a loop function is potentially unsafe
// This is a conservative approach
return false
```

**Issue:** Complete absence of write reference analysis, which is crucial for determining if a variable is actually unsafe.

**Impact:** Will flag many safe variables as unsafe, causing numerous false positives.

**Test Coverage:** Most test cases involving variables that are read-only or not modified after the loop border will fail.

#### 5. Struct Name Typo
**Go Implementation:**
```go
type iifeTacker struct {
  skippedIIFENodes map[*ast.Node]bool
}
```

**Issue:** Typo in struct name (`iifeTacker` should be `iifeTracker`).

**Impact:** While functional, this is inconsistent naming that could cause confusion.

#### 6. Incomplete Variable Declaration Analysis
**TypeScript Implementation:**
```typescript
const definition = variable?.defs[0];
const declaration = definition?.parent;
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

**Issue:** The Go version uses a different approach for determining variable declaration kind using flags instead of direct properties.

**Impact:** May not correctly identify all variable declaration types, particularly in edge cases.

### Recommendations
- Implement proper scope analysis for write reference checking
- Add async function detection alongside generator detection
- Implement accurate IIFE reference analysis using scope information
- Complete the `isSafe` function with proper border and reference analysis
- Fix the typo in `iifeTacker` struct name
- Add comprehensive test coverage for edge cases involving variable scope
- Implement the missing `getTopLoopNode` functionality properly
- Add proper variable reference tracking across scopes

---