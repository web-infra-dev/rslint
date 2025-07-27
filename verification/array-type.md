# Rule Validation: array-type

## Rule: array-type

### Test File: array-type.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core rule logic structure and flow
  - Basic AST node listeners (KindArrayType, KindTypeReference)
  - Simple type detection logic for primitive types
  - Configuration option handling (default, readonly)
  - Message building functions and error reporting
  - Basic parentheses detection for simple cases

- ⚠️ **POTENTIAL ISSUES**: 
  - Complex parentheses handling differences between TypeScript and Go AST
  - Type argument extraction and validation logic
  - Edge cases in qualified name handling
  - Readonly<T[]> pattern detection and handling
  - Text range extraction and preservation in fixes

- ❌ **INCORRECT**: 
  - Missing advanced parentheses logic from TypeScript's `isParenthesized` utility
  - Incomplete handling of `Readonly<T[]>` case transformation
  - Potential differences in AST structure handling between typescript-go and TypeScript

### Discrepancies Found

#### 1. Parentheses Detection Logic
**TypeScript Implementation:**
```typescript
function isParenthesized(node: TSESTree.Node): boolean {
  // Complex logic using context.sourceCode utilities
  // Handles multiple levels of parenthesization
}
```

**Go Implementation:**
```go
func isParenthesized(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	
	// Simple check - if the parent is a parenthesized type expression
	return ast.IsParenthesizedTypeNode(parent)
}
```

**Issue:** The Go implementation has oversimplified parentheses detection. The TypeScript version uses ESLint's sophisticated `isParenthesized` utility that can detect multiple levels of parenthesization and complex contexts, while the Go version only checks for direct parent being a parenthesized type node.

**Impact:** This could lead to incorrect fix suggestions where parentheses are incorrectly added or removed, especially in complex nested type expressions.

**Test Coverage:** Tests like `'const foo: Array<new (...args: any[]) => void> = []'` and complex union/intersection types may be affected.

#### 2. Readonly<T[]> Pattern Handling
**TypeScript Implementation:**
```typescript
if (node.typeName.name === 'Readonly' &&
    node.typeArguments?.params[0].type !== AST_NODE_TYPES.TSArrayType) {
  return;
}

const isReadonlyWithGenericArrayType =
  node.typeName.name === 'Readonly' &&
  node.typeArguments?.params[0].type === AST_NODE_TYPES.TSArrayType;
```

**Go Implementation:**
```go
// Handle Readonly<T[]> case
if typeName == "Readonly" {
	if typeRef.TypeArguments == nil || len(typeRef.TypeArguments.Nodes) == 0 {
		return
	}
	if typeRef.TypeArguments.Nodes[0].Kind != ast.KindArrayType {
		return
	}
}

isReadonlyWithGenericArrayType := typeName == "Readonly" &&
	typeRef.TypeArguments != nil &&
	len(typeRef.TypeArguments.Nodes) > 0 &&
	typeRef.TypeArguments.Nodes[0].Kind == ast.KindArrayType
```

**Issue:** The Go implementation correctly identifies the `Readonly<T[]>` pattern, but the fix generation for this case may not handle the transformation correctly, particularly in the `end` string construction where it checks `!isReadonlyWithGenericArrayType` to avoid adding `[]`.

**Impact:** Incorrect transformation of `Readonly<string[]>` to `readonly string[]` instead of the expected output.

**Test Coverage:** Test case `"const x: Readonly<string[]> = ['a', 'b'];"` with expected output `"const x: readonly string[] = ['a', 'b'];"`

#### 3. Text Range Preservation in Fixes
**TypeScript Implementation:**
```typescript
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
```

**Go Implementation:**
```go
// Get the exact text of the element type to preserve formatting
elementTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, arrayType.ElementType)
elementTypeText := string(ctx.SourceFile.Text()[elementTypeRange.Pos():elementTypeRange.End()])

// When converting T[] -> Array<T>, remove unnecessary parentheses
if ast.IsParenthesizedTypeNode(arrayType.ElementType) {
	// For parenthesized types, get the inner type to avoid double parentheses
	innerType := arrayType.ElementType.AsParenthesizedTypeNode().Type
	innerTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, innerType)
	elementTypeText = string(ctx.SourceFile.Text()[innerTypeRange.Pos():innerTypeRange.End()])
}

newText := fmt.Sprintf("%s<%s>", className, elementTypeText)
ctx.ReportNodeWithFixes(errorNode, message,
	rule.RuleFixReplace(ctx.SourceFile, errorNode, newText))
```

**Issue:** The Go implementation uses a different approach for text extraction and replacement. While it attempts to preserve formatting by extracting exact text, it uses a single replacement instead of the TypeScript's two-part replacement strategy. This could lead to issues with preserving whitespace and comments.

**Impact:** Potential loss of formatting, whitespace, or comments in the transformed code.

**Test Coverage:** Complex nested cases and cases with unusual formatting may reveal issues.

#### 4. Type Parameter Validation Logic
**TypeScript Implementation:**
```typescript
if (
  typeParams.length !== 1 ||
  (currentOption === 'array-simple' && !isSimpleType(typeParams[0]))
) {
  return;
}
```

**Go Implementation:**
```go
if len(typeParams.Nodes) != 1 {
	return
}

// For array-simple mode, determine if we have type parameters to check
// 'any' (no type params) is considered simple
isSimple := typeParams == nil || len(typeParams.Nodes) == 0 || 
	(len(typeParams.Nodes) == 1 && isSimpleType(typeParams.Nodes[0]))

// For array-simple mode, only report errors if the type is simple
if !isSimple {
	return
}
```

**Issue:** The Go implementation has logic that seems to contradict itself. First it returns if there isn't exactly 1 type parameter, but then in the `isSimple` calculation, it includes cases with 0 parameters. This logic should be restructured to match the TypeScript flow.

**Impact:** Could miss reporting errors for `Array<>` or `Array` cases in array-simple mode.

**Test Coverage:** Test cases like `'let x: Array;'` and `'let x: Array<>;'` may not be handled correctly.

#### 5. Message ID Selection Logic
**TypeScript Implementation:**
```typescript
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
var messageId string
if currentOption == "array" {
	if isReadonlyWithGenericArrayType {
		messageId = "errorStringArrayReadonly"
	} else {
		messageId = "errorStringArray"
	}
} else if currentOption == "array-simple" {
	// For array-simple mode, determine if we have type parameters to check
	// 'any' (no type params) is considered simple
	isSimple := typeParams == nil || len(typeParams.Nodes) == 0 || 
		(len(typeParams.Nodes) == 1 && isSimpleType(typeParams.Nodes[0]))
	
	// For array-simple mode, only report errors if the type is simple
	if !isSimple {
		return
	}
	
	if isReadonlyArrayType && typeName != "ReadonlyArray" {
		messageId = "errorStringArraySimpleReadonly"
	} else {
		messageId = "errorStringArraySimple"
	}
}
```

**Issue:** The Go implementation has the early return for non-simple types in the wrong place. It should check for simplicity before setting the message ID, and the message ID selection logic is embedded within the array-simple condition, making it incomplete for other modes.

**Impact:** Incorrect message IDs may be used, leading to wrong error messages.

**Test Coverage:** Various test cases with different option combinations may show wrong message IDs.

### Recommendations
- **Fix parentheses detection**: Implement a more robust `isParenthesized` function that matches TypeScript-ESLint's behavior
- **Restructure type parameter validation**: Move the simplicity check to the correct location in the control flow
- **Improve message ID logic**: Extract message ID selection to match the TypeScript implementation's structure
- **Enhance text range handling**: Consider implementing two-part text replacement similar to TypeScript version
- **Add comprehensive test coverage**: Ensure all edge cases from TypeScript tests are properly handled
- **Validate Readonly<T[]> transformation**: Ensure the fix generation correctly handles this special case
- **Review AST structure assumptions**: Verify that typescript-go AST structure matches expected patterns

---