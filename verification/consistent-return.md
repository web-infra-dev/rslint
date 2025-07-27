# Rule: consistent-return

## Test File: consistent-return.test.ts

## Validation Summary
- ✅ **CORRECT**: 
  - Basic return consistency checking
  - Function stack management for nested functions
  - Support for various function types (declaration, expression, arrow, method, accessor)
  - treatUndefinedAsUnspecified option handling
  - Error message formatting with function names
  - Async function detection and naming

- ⚠️ **POTENTIAL ISSUES**:
  - Complex Promise<void> type detection logic may have subtle differences
  - Union type handling for Promise return types needs verification
  - Edge cases with deeply nested Promise types

- ❌ **INCORRECT**:
  - Missing proper handling of union types that include void (non-Promise cases)
  - Incomplete implementation of the return policy system
  - Error reporting location differs from TypeScript implementation

## Discrepancies Found

### 1. Union Type Handling for Non-Promise Functions

**TypeScript Implementation:**
```typescript
function isReturnVoidOrThenableVoid(node: FunctionNode): boolean {
  const functionType = services.getTypeAtLocation(node);
  const tsNode = services.esTreeNodeToTSNodeMap.get(node);
  const callSignatures = functionType.getCallSignatures();

  return callSignatures.some(signature => {
    const returnType = signature.getReturnType();
    if (node.async) {
      return isPromiseVoid(tsNode, returnType);
    }
    return isTypeFlagSet(returnType, ts.TypeFlags.Void);
  });
}
```

**Go Implementation:**
```go
getReturnPolicy := func(funcNode *ast.Node) int {
  // ... complex logic with 3 return policies ...
  // Policy 2 (mixed returns) only handles Promise union types
  // Missing direct union type handling for sync functions
}
```

**Issue:** The TypeScript version has a simpler boolean check for void/thenable-void returns, while the Go version implements a more complex 3-tier policy system. However, the Go version doesn't properly handle sync functions with union return types that include void (e.g., `string | void`).

**Impact:** Test cases with union types like `number | void` may not behave correctly - the rule should allow mixed returns but might enforce strict consistency instead.

**Test Coverage:** This affects the valid test case: `function foo(flag?: boolean): number | void { ... }`

### 2. Error Reporting Location Differences

**TypeScript Implementation:**
```typescript
ReturnStatement(node): void {
  // ... logic ...
  rules.ReturnStatement(node); // Delegates to base ESLint rule for reporting
}
```

**Go Implementation:**
```go
if hasArgument {
  ctx.ReportNode(returnStmt.Expression, buildUnexpectedReturnValueMessage(funcInfo.functionName))
} else {
  // Complex range calculation for return keyword
  ctx.ReportRange(core.NewTextRange(nodeRange.Pos(), returnKeywordEnd), buildMissingReturnValueMessage(funcInfo.functionName))
}
```

**Issue:** The TypeScript version delegates error reporting to the base ESLint rule, which likely has different reporting logic. The Go version implements custom range calculation that may not match the expected column positions in test cases.

**Impact:** Error locations (line/column) in test output may not match expected values.

**Test Coverage:** All invalid test cases specify exact column positions that may not align.

### 3. Promise Union Type Detection

**TypeScript Implementation:**
```typescript
function isPromiseVoid(node: ts.Node, type: ts.Type): boolean {
  if (
    tsutils.isThenableType(checker, node, type) &&
    tsutils.isTypeReference(type)
  ) {
    const awaitedType = type.typeArguments?.[0];
    if (awaitedType) {
      if (isTypeFlagSet(awaitedType, ts.TypeFlags.Void)) {
        return true;
      }
      return isPromiseVoid(node, awaitedType);
    }
  }
  return false;
}
```

**Go Implementation:**
```go
isPromiseUnionWithVoid := func(funcNode *ast.Node, t *checker.Type) bool {
  // Complex nested logic for Promise<union> types
  // Includes special handling for Promise<Promise<void | undefined>>
  // May be over-engineered compared to TypeScript version
}
```

**Issue:** The Go version has more complex logic for handling Promise union types, including special cases that may not exist in the TypeScript version. This could lead to different behavior for edge cases.

**Impact:** Complex Promise types with unions may behave differently between implementations.

**Test Coverage:** The test case `Promise<Promise<void | undefined>>` exercises this logic.

### 4. Base Rule Integration Missing

**TypeScript Implementation:**
```typescript
const baseRule = getESLintCoreRule('consistent-return');
// ... 
const rules = baseRule.create(context);
// Delegates most logic to base ESLint rule
```

**Go Implementation:**
```go
// Implements entire rule logic from scratch
// No integration with base ESLint rule behavior
```

**Issue:** The TypeScript version extends the base ESLint `consistent-return` rule and only adds TypeScript-specific type checking. The Go version reimplements the entire rule, which may miss some edge cases or behaviors from the original ESLint rule.

**Impact:** Subtle differences in rule behavior for edge cases not covered by TypeScript-specific logic.

**Test Coverage:** Base JavaScript functionality that would normally be handled by the original ESLint rule.

### 5. Function Name Extraction Inconsistencies

**TypeScript Implementation:**
```typescript
// Relies on base rule for function name extraction and formatting
```

**Go Implementation:**
```go
getFunctionName := func(node *ast.Node) string {
  // Custom implementation with specific formatting
  // Returns "Function 'name'", "Async function 'name'", etc.
}
```

**Issue:** The custom function name extraction in Go may not match the exact formatting used by the base ESLint rule, potentially causing test failures due to message differences.

**Impact:** Error messages may have different formatting for function names.

**Test Coverage:** All error test cases specify exact function names in error messages.

## Recommendations

1. **Fix Union Type Handling**: Implement proper detection of sync functions with union return types that include void. The `getReturnPolicy` function should return policy 2 (mixed returns allowed) for cases like `string | void`.

2. **Simplify Return Policy Logic**: Consider simplifying the 3-tier policy system to match the TypeScript implementation's simpler boolean approach for void detection.

3. **Align Error Reporting**: Ensure error reporting locations match the expected test output by either adjusting the range calculation or updating test expectations.

4. **Review Promise Type Detection**: Simplify the `isPromiseUnionWithVoid` logic to more closely match the TypeScript version's approach.

5. **Add Missing Test Cases**: Consider adding test cases that specifically validate the edge cases in Promise union type handling.

6. **Verify Function Name Formatting**: Ensure the custom function name extraction produces messages that match the expected test output format.

7. **Document Behavioral Differences**: If the Go implementation intentionally differs from the TypeScript version (e.g., more sophisticated Promise handling), document these differences and update tests accordingly.

---