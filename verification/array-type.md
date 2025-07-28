# Array-Type Rule Validation

## Rule: array-type

### Test File: array-type.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core logic for handling Array<T> vs T[] preferences
  - Basic readonly array handling with ReadonlyArray<T> vs readonly T[]
  - Simple vs complex type detection with isSimpleType function covering all basic TypeScript keywords
  - Options parsing for default and readonly configurations with dual-format support
  - Message ID mapping and error message construction for all 6 message types
  - Fix generation for both array-to-generic and generic-to-array conversions
  - Empty array type handling (Array and Array<>) with 'any' fallback
  - Readonly<T[]> pattern detection and handling

- ⚠️ **POTENTIAL ISSUES**:
  - Parentheses detection logic is significantly simplified in Go version
  - Fix generation uses different strategy (whole node replacement vs range replacement)
  - Array-simple mode logic flow differs but should produce same results
  - Edge case handling for malformed or unusual type references

- ❌ **INCORRECT**: 
  - QualifiedName handling in isSimpleType has inconsistent logic paths
  - Parentheses handling for complex nested types will likely differ
  - Message ID assignment logic for array-simple mode has subtle differences

### Discrepancies Found

#### 1. QualifiedName Handling in isSimpleType

**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TSQualifiedName:
  return true;
case AST_NODE_TYPES.TSTypeReference:
  if (
    node.typeName.type === AST_NODE_TYPES.Identifier &&
    node.typeName.name === 'Array'
  ) {
    // Array type logic
  } else {
    if (node.typeArguments) {
      return false;
    }
    return isSimpleType(node.typeName);
  }
```

**Go Implementation:**
```go
case ast.KindQualifiedName:
  return true
case ast.KindTypeReference:
  // ... Array handling logic
  } else if ast.IsQualifiedName(typeRef.TypeName) {
    // TypeReference with a QualifiedName is simple if it has no type arguments
    if typeRef.TypeArguments != nil {
      return false
    }
    return true
  }
```

**Issue:** The Go implementation has inconsistent logic: direct QualifiedName always returns true, but QualifiedName within TypeReference returns false if type arguments exist. The TypeScript version delegates to isSimpleType(node.typeName) for the TypeReference case.

**Impact:** Types like `fooName.BarType<T>` will be handled differently - Go will consider them not simple (correct), but direct qualified names might be handled inconsistently.

**Test Coverage:** Test cases `"let v: Array<fooName.BarType>"` and `"let w: fooName.BazType<string>[]"` will reveal this inconsistency.

#### 2. Parentheses Detection and Application

**TypeScript Implementation:**
```typescript
function typeNeedsParentheses(node: TSESTree.Node): boolean {
  switch (node.type) {
    case AST_NODE_TYPES.TSUnionType:
    case AST_NODE_TYPES.TSFunctionType:
    case AST_NODE_TYPES.TSIntersectionType:
    case AST_NODE_TYPES.TSTypeOperator:
    case AST_NODE_TYPES.TSInferType:
    case AST_NODE_TYPES.TSConstructorType:
    case AST_NODE_TYPES.TSConditionalType:
      return true;
    // ... other cases
  }
}

// Uses context.sourceCode for parentheses detection
!isParenthesized(node.parent.elementType, context.sourceCode)
```

**Go Implementation:**
```go
func typeNeedsParentheses(node *ast.Node) bool {
  switch node.Kind {
    case ast.KindUnionType,
      ast.KindFunctionType,
      ast.KindIntersectionType,
      ast.KindTypeOperator,
      ast.KindInferType,
      ast.KindConstructorType,
      ast.KindConditionalType:
      return true
    // ... other cases
  }
}

func isParenthesized(node *ast.Node) bool {
  // Much simpler - only checks immediate parent
  return ast.IsParenthesizedTypeNode(parent)
}
```

**Issue:** The Go version's parentheses detection is significantly simplified and doesn't use source code context. This will lead to different parenthesization decisions.

**Impact:** Complex types may have incorrect parentheses added or removed, particularly in nested scenarios like `readonly (string | number)[]`.

**Test Coverage:** All test cases with union types, function types, and complex nested arrays will be affected.

#### 3. Array-Simple Mode Message ID Selection

**TypeScript Implementation:**
```typescript
if (
  typeParams.length !== 1 ||
  (currentOption === 'array-simple' && !isSimpleType(typeParams[0]))
) {
  return;
}

// Later...
const messageId =
  currentOption === 'array'
    ? isReadonlyWithGenericArrayType
      ? 'errorStringArrayReadonly'
      : 'errorStringArray'
    : isReadonlyArrayType && node.typeName.name !== 'ReadonlyArray'
      ? 'errorStringArraySimpleReadonly'
      : 'errorStringArraySimple';
```

**Go Implementation:**
```go
// For array-simple mode, determine if we have type parameters to check
isSimple := typeParams == nil || len(typeParams.Nodes) == 0 || 
    (len(typeParams.Nodes) == 1 && isSimpleType(typeParams.Nodes[0]))

// For array-simple mode, only report errors if the type is simple
if !isSimple {
  return
}

// Later...
if currentOption == "array" {
  if isReadonlyWithGenericArrayType {
    messageId = "errorStringArrayReadonly"
  } else {
    messageId = "errorStringArray"
  }
} else if currentOption == "array-simple" {
  if isReadonlyArrayType && typeName != "ReadonlyArray" {
    messageId = "errorStringArraySimpleReadonly"
  } else {
    messageId = "errorStringArraySimple"
  }
}
```

**Issue:** The Go version pre-calculates simplicity and uses different conditional logic for message ID selection. The conditions should be equivalent but the flow is different.

**Impact:** Edge cases in array-simple mode might select different message IDs, affecting error messages.

**Test Coverage:** Array-simple mode test cases with complex types like `ReadonlyArray<string | number>` will test this logic.

#### 4. Fix Generation Strategy Differences

**TypeScript Implementation:**
```typescript
fix(fixer) {
  const typeNode = node.elementType;
  const arrayType = isReadonly ? 'ReadonlyArray' : 'Array';

  return [
    fixer.replaceTextRange(
      [errorNode.range[0], typeNode.range[0]],
      `${arrayType}<`,
    ),
    fixer.replaceTextRange(
      [typeNode.range[1], errorNode.range[1]],
      '>',
    ),
  ];
}
```

**Go Implementation:**
```go
// Get the exact text of the element type to preserve formatting
elementTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, arrayType.ElementType)
elementTypeText := string(ctx.SourceFile.Text()[elementTypeRange.Pos():elementTypeRange.End()])

// When converting T[] -> Array<T>, remove unnecessary parentheses
if ast.IsParenthesizedTypeNode(arrayType.ElementType) {
  innerType := arrayType.ElementType.AsParenthesizedTypeNode().Type
  innerTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, innerType)
  elementTypeText = string(ctx.SourceFile.Text()[innerTypeRange.Pos():innerTypeRange.End()])
}

newText := fmt.Sprintf("%s<%s>", className, elementTypeText)
ctx.ReportNodeWithFixes(errorNode, message,
  rule.RuleFixReplace(ctx.SourceFile, errorNode, newText))
```

**Issue:** TypeScript uses precise range replacements to preserve formatting, while Go replaces the entire node. Go also has special logic to remove parentheses from element types.

**Impact:** May produce different formatting in fixes, especially regarding whitespace and parentheses preservation.

**Test Coverage:** All invalid test cases with `output` properties will test fix generation accuracy.

#### 5. Readonly<T[]> End Bracket Handling

**TypeScript Implementation:**
```typescript
const end = `${typeParens ? ')' : ''}${isReadonlyWithGenericArrayType ? '' : `[]`}${parentParens ? ')' : ''}`;
```

**Go Implementation:**
```go
if !isReadonlyWithGenericArrayType {
  end += "[]"
}
```

**Issue:** Both handle the case where `Readonly<T[]>` shouldn't get additional `[]` brackets, but the logic structure differs.

**Impact:** Should produce the same results but different code organization could hide edge case bugs.

**Test Coverage:** Test cases like `"const x: Readonly<string[]> = ['a', 'b'];"` test this logic.

#### 6. Empty Array Type Handling

**TypeScript Implementation:**
```typescript
if (!typeParams || typeParams.length === 0) {
  // Create an 'any' array
  context.report({
    node,
    messageId,
    data: {
      type: 'any',
      className: isReadonlyArrayType ? 'ReadonlyArray' : 'Array',
      readonlyPrefix,
    },
    fix(fixer) {
      return fixer.replaceText(node, `${readonlyPrefix}any[]`);
    },
  });
  return;
}
```

**Go Implementation:**
```go
if typeParams == nil || len(typeParams.Nodes) == 0 {
  // Create an 'any' array
  className := "Array"
  if isReadonlyArrayType {
    className = "ReadonlyArray"
  }

  var message rule.RuleMessage
  switch messageId {
  case "errorStringArray":
    message = buildErrorStringArrayMessage(className, readonlyPrefix, "any")
  // ... other cases
  }

  ctx.ReportNodeWithFixes(node, message,
    rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("%sany[]", readonlyPrefix)))
  return
}
```

**Issue:** Both implementations handle empty array types correctly by converting them to `any[]`, but Go uses separate message construction while TypeScript uses data interpolation.

**Impact:** Should produce identical results but message formatting consistency needs verification.

**Test Coverage:** Test cases `"let x: Array;"` and `"let x: Array<>;"` test this functionality.

### Recommendations

- **Fix QualifiedName consistency**: Ensure isSimpleType handles QualifiedName the same way whether it appears directly or within a TypeReference
- **Enhance parentheses detection**: Implement more sophisticated parentheses detection that considers source code context, similar to the TypeScript version
- **Verify message ID logic**: Ensure array-simple mode message selection produces identical results to TypeScript version
- **Validate fix generation**: Test that fix outputs exactly match expected results, particularly for complex type conversions and formatting
- **Add comprehensive test coverage**: Include specific test cases for:
  - Qualified names with and without type parameters
  - Complex nested readonly arrays
  - Parenthesized types in various contexts
  - All combinations of array-simple mode with complex types
- **Standardize message formatting**: Ensure all message construction produces identical strings to TypeScript version

### Test Cases Requiring Special Attention

1. `"let v: Array<fooName.BarType> = [{ bar: 'bar' }];"` - Tests QualifiedName handling
2. `"let w: fooName.BazType<string>[] = [['baz']];"` - Tests QualifiedName with type parameters
3. `"type barUnion = (string | number | boolean)[];"` - Tests union type parentheses
4. `"const foo: ReadonlyArray<new (...args: any[]) => void> = [];"` - Tests complex type parentheses
5. `"const x: Readonly<string[]> = ['a', 'b'];"` - Tests Readonly<T[]> pattern
6. `"let x: Array;"` and `"let x: Array<>;"` - Tests empty array type handling
7. All multi-pass fix scenarios for deeply nested arrays

### Overall Assessment

The Go implementation captures the core functionality correctly but has several areas where the logic differs from the TypeScript version. Most critically, the parentheses handling and QualifiedName logic need attention to ensure identical behavior. The fix generation strategy difference (range replacement vs whole node replacement) may be acceptable if it produces the same final results, but should be thoroughly tested.

---