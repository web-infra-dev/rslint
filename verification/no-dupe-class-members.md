## Rule: no-dupe-class-members

### Test File: no-dupe-class-members.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic duplicate detection for methods and properties
  - Static vs instance member separation
  - Getter/setter pair handling
  - Computed property exclusion
  - Class declaration and expression support
  - Numeric literal normalization (10 vs 1e1)
  - Multiple duplicate detection

- ⚠️ **POTENTIAL ISSUES**: 
  - Debug logging left in production code
  - Complex numeric literal handling may not cover all edge cases
  - Error message format may not match ESLint exactly

- ❌ **INCORRECT**: 
  - Missing TypeScript empty body function expression handling
  - No method overload support (function signatures)
  - Potential issues with string literal method names

### Discrepancies Found

#### 1. Missing TypeScript Empty Body Function Expression Handling
**TypeScript Implementation:**
```typescript
function wrapMemberDefinitionListener<
  N extends TSESTree.MethodDefinition | TSESTree.PropertyDefinition,
>(coreListener: (node: N) => void): (node: N) => void {
  return (node: N): void => {
    if (node.computed) {
      return;
    }

    if (
      node.value &&
      node.value.type === AST_NODE_TYPES.TSEmptyBodyFunctionExpression
    ) {
      return;
    }

    return coreListener(node);
  };
}
```

**Go Implementation:**
```go
// No equivalent handling for TypeScript empty body function expressions
```

**Issue:** The Go implementation doesn't skip TypeScript method overloads with empty bodies (function signatures), which could lead to false positives.

**Impact:** Method overload signatures like `foo(a: string): string;` followed by implementation `foo(a: any): any {}` may be incorrectly flagged as duplicates.

**Test Coverage:** The test case with method overloads may fail or pass incorrectly:
```typescript
class Foo {
  foo(a: string): string;
  foo(a: number): number;  
  foo(a: any): any {}
}
```

#### 2. Debug Code in Production
**TypeScript Implementation:**
```typescript
// No debug logging
```

**Go Implementation:**
```go
// Debug info
if memberName == "foo" {
    ctx.ReportNode(memberNode, buildUnexpectedMessage(fmt.Sprintf("DEBUG: Processing %s (kind: %s, static: %v, existing: %d)", memberName, memberKind, memberIsStatic, len(existingMembers))))
}
```

**Issue:** Debug logging code left in the production implementation.

**Impact:** Will generate spurious error messages for any member named "foo", causing test failures.

**Test Coverage:** Any test with a method/property named "foo" will produce unexpected debug output.

#### 3. Numeric Literal Parsing Edge Cases
**TypeScript Implementation:**
```typescript
// Relies on base ESLint rule for proper numeric literal handling
```

**Go Implementation:**
```go
// Check if it's a numeric literal and evaluate it
if nameNode != nil && nameNode.Kind == ast.KindNumericLiteral {
    numLit := nameNode.AsNumericLiteral()
    // Parse the numeric literal text to get its actual value
    // This will convert both "10" and "1e1" to "10"
    var val float64
    fmt.Sscanf(numLit.Text, "%g", &val)
    return fmt.Sprintf("%g", val)
}
```

**Issue:** Custom numeric parsing may not handle all JavaScript numeric literal formats correctly (hex, octal, binary literals).

**Impact:** Methods like `0x10() {}` and `16() {}` might not be detected as duplicates when they should be.

**Test Coverage:** Current tests only cover decimal and scientific notation.

#### 4. Error Message Format Inconsistency
**TypeScript Implementation:**
```typescript
// Uses base ESLint rule messages
messages: baseRule.meta.messages,
```

**Go Implementation:**
```go
func buildUnexpectedMessage(name string) rule.RuleMessage {
    return rule.RuleMessage{
        Id:          "unexpected",
        Description: fmt.Sprintf("Duplicate name '%s'.", name),
    }
}
```

**Issue:** Error message format may not exactly match the base ESLint rule format.

**Impact:** Test snapshots expecting specific message formats may fail.

**Test Coverage:** All invalid test cases expect `messageId: 'unexpected'` with data containing the member name.

### Recommendations
- **CRITICAL**: Remove debug logging code for "foo" members
- **HIGH**: Implement TypeScript empty body function expression detection to handle method overloads
- **MEDIUM**: Verify error message format matches ESLint exactly
- **LOW**: Enhance numeric literal parsing to handle hex/octal/binary literals
- **LOW**: Add test cases for method overloads with different signatures
- **LOW**: Add test cases for various numeric literal formats (hex, octal, binary)

---