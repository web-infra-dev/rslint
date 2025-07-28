## Rule: no-magic-numbers

### Test File: no-magic-numbers.test.ts

### Validation Summary
- ✅ **CORRECT**: Core AST pattern matching, TypeScript-specific features (enums, readonly properties, type indexes, numeric literal types), numeric format handling (hex, binary, octal, scientific notation), bigint support, negative number handling via unary expressions
- ⚠️ **POTENTIAL ISSUES**: Missing base rule fallback behavior, extra configuration options not in original, message ID vs hardcoded message format, complex value comparison logic may have edge cases
- ❌ **INCORRECT**: No delegation to base ESLint rule, hardcoded error messages instead of using messageId system

### Discrepancies Found

#### 1. Base Rule Integration Missing
**TypeScript Implementation:**
```typescript
const baseRule = getESLintCoreRule('no-magic-numbers');
// ... 
// Let the base rule deal with the rest
rules.Literal(node);
```

**Go Implementation:**
```go
// No base rule integration - implements everything from scratch
```

**Issue:** The TypeScript version extends the base ESLint `no-magic-numbers` rule and falls back to it for non-TypeScript-specific cases. The Go version implements the entire rule logic without this fallback.

**Impact:** May miss some edge cases or behaviors that the base ESLint rule handles, particularly for JavaScript-specific scenarios.

**Test Coverage:** All test cases rely on complete rule implementation rather than base rule delegation.

#### 2. Message ID System vs Hardcoded Messages
**TypeScript Implementation:**
```typescript
context.report({
  node: fullNumberNode,
  messageId: 'noMagic',
  data: { raw },
});
```

**Go Implementation:**
```go
message := rule.RuleMessage{
  Id:          "noMagic",
  Description: fmt.Sprintf("No magic number: %s.", raw),
}
```

**Issue:** TypeScript uses the message ID system with data interpolation, while Go hardcodes the message format. This could cause test failures if the test framework expects specific message formats.

**Impact:** Tests expecting `messageId: 'noMagic'` with data objects may fail with the Go implementation's hardcoded message.

**Test Coverage:** All invalid test cases expect the `messageId: 'noMagic'` format.

#### 3. Extra Configuration Options
**TypeScript Implementation:**
```typescript
// Standard options only:
detectObjects, enforceConst, ignore, ignoreArrayIndexes, 
ignoreEnums, ignoreNumericLiteralTypes, ignoreReadonlyClassProperties, ignoreTypeIndexes
```

**Go Implementation:**
```go
// Additional options not in TypeScript:
IgnoreDefaultValues          bool `json:"ignoreDefaultValues"`
IgnoreClassFieldInitialValues bool `json:"ignoreClassFieldInitialValues"`
```

**Issue:** The Go implementation includes extra configuration options that don't exist in the original TypeScript rule.

**Impact:** These extra options may provide additional functionality but could cause configuration incompatibility.

**Test Coverage:** No test cases use these extra options, so they're untested functionality.

#### 4. Complex Value Comparison Logic
**TypeScript Implementation:**
```typescript
function normalizeIgnoreValue(value: bigint | number | string): bigint | number {
  if (typeof value === 'string') {
    return BigInt(value.slice(0, -1));
  }
  return value;
}
```

**Go Implementation:**
```go
// Complex valuesEqual function with multiple type conversions
func valuesEqual(a, b any) bool {
  // 50+ lines of comparison logic
}
```

**Issue:** The Go implementation has much more complex value comparison logic that may not match the TypeScript behavior exactly.

**Impact:** Could lead to differences in which ignore values match which literals, particularly with type coercion.

**Test Coverage:** Test cases with various ignore value formats (numbers, bigints, scientific notation) exercise this logic.

#### 5. Option Parsing Complexity
**TypeScript Implementation:**
```typescript
// Standard ESLint option parsing via rule framework
```

**Go Implementation:**
```go
// Custom dual-format parsing (array vs object)
if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
  optsMap, ok = optArray[0].(map[string]interface{})
} else {
  optsMap, ok = options.(map[string]interface{})
}
```

**Issue:** The Go implementation includes complex option parsing logic to handle both array and object formats, which may not be necessary.

**Impact:** Could introduce parsing errors or inconsistencies with the standard ESLint configuration format.

**Test Coverage:** Test cases use array format `[{ option: value }]` which exercises this parsing logic.

### Recommendations
- Consider implementing base rule delegation for non-TypeScript-specific cases
- Align error message system with TypeScript implementation (use messageId + data pattern)
- Remove extra configuration options not present in original rule or document them as extensions
- Simplify value comparison logic to match TypeScript behavior more closely
- Simplify option parsing to standard ESLint format
- Add test cases specifically for the extra configuration options if they're kept
- Verify that all numeric format edge cases (hex, binary, octal, scientific notation) work identically

---