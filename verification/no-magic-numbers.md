# Rule Validation: no-magic-numbers

## Rule: no-magic-numbers

### Test File: no-magic-numbers.test.ts

### Validation Summary
- ✅ **CORRECT**: Core magic number detection, TypeScript-specific patterns (enums, literal types, type indexes, readonly properties), bigint support, unary expression handling, ignore value matching
- ⚠️ **POTENTIAL ISSUES**: Base ESLint rule fallback behavior, some configuration options missing, numeric literal parsing edge cases
- ❌ **INCORRECT**: Missing base ESLint rule integration, incomplete option set, potential issues with complex numeric formats

### Discrepancies Found

#### 1. Missing Base ESLint Rule Integration
**TypeScript Implementation:**
```typescript
const baseRule = getESLintCoreRule('no-magic-numbers');
// ...
const rules = baseRule.create(context);
// ...
// Let the base rule deal with the rest
rules.Literal(node);
```

**Go Implementation:**
```go
// No equivalent base rule fallback mechanism
// All logic is implemented directly without fallback
```

**Issue:** The TypeScript implementation extends the base ESLint rule and falls back to it for cases not covered by TypeScript-specific logic. The Go implementation reimplements everything from scratch without this fallback.

**Impact:** May miss edge cases or behaviors that the base ESLint rule handles, potentially causing different behavior for complex numeric literal scenarios.

**Test Coverage:** This could affect any test case that relies on base ESLint behavior not explicitly covered by TypeScript-specific options.

#### 2. Missing Configuration Options
**TypeScript Implementation:**
```typescript
// Includes all base rule options plus TypeScript extensions
{
  detectObjects: false,
  enforceConst: false,
  ignore: [],
  ignoreArrayIndexes: false,
  ignoreEnums: false,
  ignoreNumericLiteralTypes: false,
  ignoreReadonlyClassProperties: false,
  ignoreTypeIndexes: false,
}
```

**Go Implementation:**
```go
// Missing some base ESLint options
type NoMagicNumbersOptions struct {
    // Has ignoreDefaultValues and ignoreClassFieldInitialValues
    // but these are not in the TypeScript version being ported
    IgnoreDefaultValues          bool `json:"ignoreDefaultValues"`
    IgnoreClassFieldInitialValues bool `json:"ignoreClassFieldInitialValues"`
}
```

**Issue:** The Go implementation includes some options that aren't in the TypeScript-ESLint version being ported, and may be missing some base ESLint options.

**Impact:** Configuration mismatch could cause different behavior when users expect consistent options.

**Test Coverage:** Test cases using base ESLint options might fail.

#### 3. Numeric Literal Parsing Complexity
**TypeScript Implementation:**
```typescript
// Uses node.raw directly from AST
if (
  node.parent.type === AST_NODE_TYPES.UnaryExpression &&
  node.parent.operator === '-'
) {
  fullNumberNode = node.parent;
  raw = `${node.parent.operator}${node.raw}`;
}
```

**Go Implementation:**
```go
// Implements custom numeric parsing
func parseNumericValue(text string) any {
    // Handle different number formats manually
    if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
        // Hexadecimal
    } else if strings.HasPrefix(text, "0b") || strings.HasPrefix(text, "0B") {
        // Binary
    }
    // ... more format handling
}
```

**Issue:** The Go implementation reimplements numeric parsing that TypeScript handles natively, potentially missing edge cases or format variations.

**Impact:** Could cause inconsistent behavior with complex numeric formats like scientific notation edge cases.

**Test Coverage:** Test cases with various numeric formats (0x, 0b, 0o, scientific notation) should be verified carefully.

#### 4. Error Message Format Differences
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

**Issue:** The Go implementation hardcodes the message format instead of using a message template system like the TypeScript version.

**Impact:** Error messages may not match exactly, affecting user experience and tooling that relies on specific message formats.

**Test Coverage:** All invalid test cases expect specific message formats that may not match.

#### 5. AST Node Kind Mappings
**TypeScript Implementation:**
```typescript
// Uses TSESTree AST node types
AST_NODE_TYPES.UnaryExpression
AST_NODE_TYPES.TSEnumMember
AST_NODE_TYPES.TSLiteralType
AST_NODE_TYPES.TSIndexedAccessType
```

**Go Implementation:**
```go
// Uses typescript-go AST kinds
ast.KindPrefixUnaryExpression
ast.KindEnumMember
ast.KindLiteralType
ast.KindIndexedAccessType
```

**Issue:** Need to verify that AST node kind mappings are correct between the two AST representations.

**Impact:** Incorrect mappings could cause rules to miss or incorrectly flag patterns.

**Test Coverage:** All TypeScript-specific test cases need verification.

#### 6. Bigint Handling Edge Cases
**TypeScript Implementation:**
```typescript
function normalizeIgnoreValue(value: bigint | number | string): bigint | number {
  if (typeof value === 'string') {
    return BigInt(value.slice(0, -1)); // Remove 'n' suffix
  }
  return value;
}
```

**Go Implementation:**
```go
func normalizeIgnoreValue(value any) any {
    if strVal, ok := value.(string); ok {
        if strings.HasSuffix(strVal, "n") {
            numStr := strVal[:len(strVal)-1]
            if bigInt, ok := new(big.Int).SetString(numStr, 10); ok {
                return bigInt
            }
        }
    }
    return value
}
```

**Issue:** The Go implementation has more error handling but different type handling for bigint values.

**Impact:** Could cause different behavior when parsing bigint ignore values.

**Test Coverage:** Test cases with bigint ignore values need careful verification.

### Recommendations
- Investigate whether base ESLint rule fallback behavior is needed or if full reimplementation is sufficient
- Remove or properly document the extra configuration options (ignoreDefaultValues, ignoreClassFieldInitialValues) if they're not part of the TypeScript-ESLint rule
- Verify AST node kind mappings are correct for all TypeScript-specific patterns
- Implement message templating system to match TypeScript error message formats exactly
- Add comprehensive test coverage for numeric format edge cases
- Verify bigint handling matches TypeScript behavior exactly
- Consider implementing the schema validation that the TypeScript version uses

---