## Rule: default-param-last

### Test File: default-param-last.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core logic for detecting default/optional parameters before required parameters
  - Support for function declarations, expressions, and arrow functions
  - Support for class constructors and methods
  - Error message consistency ("Default parameters should be last.")
  - Proper iteration through parameters from right to left
  - Basic parameter type detection (optional, default, rest)

- ⚠️ **POTENTIAL ISSUES**: 
  - TypeScript parameter property handling may differ from Go implementation
  - Destructuring pattern handling needs verification
  - Rest parameter detection logic differs between implementations

- ❌ **INCORRECT**: 
  - Missing support for TSParameterProperty unwrapping
  - Destructuring patterns (ArrayPattern, ObjectPattern) not handled in Go
  - AssignmentPattern detection logic is incomplete

### Discrepancies Found

#### 1. TSParameterProperty Handling
**TypeScript Implementation:**
```typescript
const param =
  current.type === AST_NODE_TYPES.TSParameterProperty
    ? current.parameter
    : current;
```

**Go Implementation:**
```go
// No equivalent handling for TSParameterProperty
// Go code directly processes parameters without unwrapping
```

**Issue:** The TypeScript version unwraps TSParameterProperty nodes to get the actual parameter, but the Go version doesn't handle this case. TSParameterProperty represents constructor parameters with accessibility modifiers (public, private, protected).

**Impact:** Constructor parameters with accessibility modifiers may not be properly analyzed, leading to missed violations in class constructors.

**Test Coverage:** Constructor test cases with `public a = 0`, `protected b?: number`, etc. may fail.

#### 2. Destructuring Pattern Support
**TypeScript Implementation:**
```typescript
function isOptionalParam(node: TSESTree.Parameter): boolean {
  return (
    (node.type === AST_NODE_TYPES.ArrayPattern ||
      node.type === AST_NODE_TYPES.AssignmentPattern ||
      node.type === AST_NODE_TYPES.Identifier ||
      node.type === AST_NODE_TYPES.ObjectPattern ||
      node.type === AST_NODE_TYPES.RestElement) &&
    node.optional
  );
}
```

**Go Implementation:**
```go
func isOptionalParam(node *ast.Node) bool {
	if node == nil || !ast.IsParameter(node) {
		return false
	}

	param := node.AsParameterDeclaration()
	return param.QuestionToken != nil
}
```

**Issue:** The Go version only checks for QuestionToken on ParameterDeclaration, but doesn't handle destructuring patterns (ArrayPattern, ObjectPattern) that can also have optional markers.

**Impact:** Test cases with destructuring patterns like `function foo({ a } = {}, b) {}` or `function foo([a] = [], b) {}` may not be correctly identified as having default parameters.

**Test Coverage:** Tests with destructuring patterns will likely fail to report violations.

#### 3. AssignmentPattern Detection
**TypeScript Implementation:**
```typescript
function isPlainParam(node: TSESTree.Parameter): boolean {
  return !(
    node.type === AST_NODE_TYPES.AssignmentPattern ||
    node.type === AST_NODE_TYPES.RestElement ||
    isOptionalParam(node)
  );
}
```

**Go Implementation:**
```go
func isPlainParam(node *ast.Node) bool {
	if node == nil {
		return false
	}

	return !isOptionalParam(node) && !isDefaultParam(node) && !isRestParam(node)
}
```

**Issue:** The TypeScript version explicitly checks for AssignmentPattern in isPlainParam, but the Go version relies on isDefaultParam which only checks for Initializer on ParameterDeclaration. This may not cover all cases where AssignmentPattern appears.

**Impact:** Some destructuring with defaults may not be properly detected as non-plain parameters.

**Test Coverage:** Complex destructuring cases may not trigger violations when they should.

#### 4. Parameter Extraction Logic
**TypeScript Implementation:**
```typescript
// Directly accesses node.params for all function types
```

**Go Implementation:**
```go
// Different parameter extraction for each function type
switch functionNode.Kind {
case ast.KindArrowFunction:
    if functionNode.AsArrowFunction().Parameters != nil {
        params = functionNode.AsArrowFunction().Parameters.Nodes
    }
case ast.KindFunctionDeclaration:
    if functionNode.AsFunctionDeclaration().Parameters != nil {
        params = functionNode.AsFunctionDeclaration().Parameters.Nodes
    }
// ... more cases
```

**Issue:** The Go implementation manually extracts parameters for each function type, while TypeScript has a unified approach. This is more verbose but functionally equivalent if all cases are covered.

**Impact:** Potential for missing function types if not all cases are handled in the switch statement.

**Test Coverage:** Should verify all function types are properly supported.

### Recommendations
- Add TSParameterProperty unwrapping logic to handle constructor parameters with accessibility modifiers
- Implement proper destructuring pattern detection for ArrayPattern and ObjectPattern
- Enhance AssignmentPattern detection to cover destructuring with defaults
- Verify that all function types are properly handled in the parameter extraction switch statement
- Add specific test cases for TypeScript-specific constructs like parameter properties
- Consider adding debug logging to verify parameter detection is working correctly for edge cases

---