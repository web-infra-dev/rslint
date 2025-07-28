## Rule: max-params

### Test File: max-params.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core parameter counting logic
  - Void `this` parameter detection and handling
  - Default options (max: 3, countVoidThis: false)
  - Configuration option parsing (max, maximum, countVoidThis)
  - Support for all major function-like AST nodes
  - Error message structure with messageId: "exceed"

- ⚠️ **POTENTIAL ISSUES**: 
  - AST node kind mappings between TypeScript and Go may differ
  - Type annotation checking for void `this` parameters
  - Function name extraction for error messages

- ❌ **INCORRECT**: 
  - Missing `TSDeclareFunction` node type handling
  - Inconsistent function name extraction logic

### Discrepancies Found

#### 1. Missing TSDeclareFunction Support
**TypeScript Implementation:**
```typescript
return {
  ArrowFunctionExpression: wrapListener(baseRules.ArrowFunctionExpression),
  FunctionDeclaration: wrapListener(baseRules.FunctionDeclaration),
  FunctionExpression: wrapListener(baseRules.FunctionExpression),
  TSDeclareFunction: wrapListener(baseRules.FunctionDeclaration),
  TSFunctionType: wrapListener(baseRules.FunctionDeclaration),
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
  ast.KindFunctionDeclaration: checkFunction,
  ast.KindFunctionExpression:  checkFunction,
  ast.KindArrowFunction:       checkFunction,
  ast.KindMethodDeclaration:   checkFunction,
  ast.KindConstructor:         checkFunction,
  ast.KindGetAccessor:         checkFunction,
  ast.KindSetAccessor:         checkFunction,
  ast.KindFunctionType:        checkFunction,
}
```

**Issue:** The Go implementation is missing a listener for `TSDeclareFunction` (declare function statements), which the TypeScript version explicitly handles.

**Impact:** The rule won't trigger on TypeScript declare function statements like `declare function makeDate(m: number, d: number, y: number): Date;`

**Test Coverage:** The test case with `declare function makeDate(m: number, d: number, y: number): Date;` should fail in the Go implementation.

#### 2. Function Name Extraction Inconsistency
**TypeScript Implementation:**
```typescript
// Uses base ESLint rule which has standardized function name extraction
const baseRules = baseRule.create(context);
```

**Go Implementation:**
```go
func getFunctionName(node *ast.Node) string {
  switch node.Kind {
  case ast.KindFunctionDeclaration:
    funcDecl := node.AsFunctionDeclaration()
    if funcDecl.Name() != nil {
      return "Function '" + funcDecl.Name().AsIdentifier().Text + "'"
    }
    return "Function"
  // ... more cases
  }
}
```

**Issue:** The Go implementation includes detailed function name extraction logic that may not match the base ESLint rule's behavior exactly. The TypeScript version delegates this to the base rule.

**Impact:** Error messages may have slightly different formatting or function name identification.

**Test Coverage:** Error message format differences may not be caught by current tests that only check for `messageId: 'exceed'`.

#### 3. AST Node Kind Mapping Uncertainty
**TypeScript Implementation:**
```typescript
type FunctionLike =
  | TSESTree.ArrowFunctionExpression
  | TSESTree.FunctionDeclaration
  | TSESTree.FunctionExpression
  | TSESTree.TSDeclareFunction
  | TSESTree.TSFunctionType;
```

**Go Implementation:**
```go
// Uses ast.Kind constants like:
// ast.KindFunctionDeclaration, ast.KindArrowFunction, etc.
```

**Issue:** The mapping between TypeScript ESTree node types and Go AST kinds needs verification. For example, `ArrowFunctionExpression` in TypeScript maps to `ast.KindArrowFunction` in Go.

**Impact:** Rule may not trigger on certain function types if AST kind mapping is incorrect.

**Test Coverage:** All function type test cases should verify this mapping works correctly.

#### 4. Void This Parameter Type Checking
**TypeScript Implementation:**
```typescript
node.params[0].typeAnnotation?.typeAnnotation.type !== AST_NODE_TYPES.TSVoidKeyword
```

**Go Implementation:**
```go
func isVoidThisParam(param *ast.Node) bool {
  // ... checks for parameter name "this"
  if paramNode.Type == nil {
    return false
  }
  return paramNode.Type.Kind == ast.KindVoidKeyword
}
```

**Issue:** The TypeScript version uses `typeAnnotation?.typeAnnotation.type` (nested structure) while Go uses `paramNode.Type.Kind`. The type annotation structure may differ between implementations.

**Impact:** Void `this` parameter detection may fail, causing incorrect parameter counting when `countVoidThis: false`.

**Test Coverage:** Test cases with `method(this: void, ...)` should verify this works correctly.

### Recommendations
- Add support for `TSDeclareFunction` AST node type (likely maps to a specific Go AST kind)
- Verify AST node kind mappings are correct for all function-like constructs
- Test void `this` parameter type annotation checking with actual TypeScript code
- Ensure error message formatting matches the base ESLint rule behavior
- Add integration tests that compare error messages between TypeScript-ESLint and RSLint implementations
- Consider adding debug logging to verify AST node types are being matched correctly

---