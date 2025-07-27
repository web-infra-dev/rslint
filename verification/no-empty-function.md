## Rule: no-empty-function

### Test File: no-empty-function.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic empty function detection logic
  - Constructor parameter properties detection
  - Private/protected constructor accessibility checks
  - Override method detection
  - Configuration option parsing structure
  - Core AST pattern matching for function types

- ⚠️ **POTENTIAL ISSUES**: 
  - Method kind detection for constructors, getters, and setters
  - Decorator detection mechanism
  - Arrow function body type checking
  - Function name extraction logic complexity

- ❌ **INCORRECT**: 
  - AST node kind comparisons for method types
  - Decorator flag checking mechanism
  - Generator function detection approach

### Discrepancies Found

#### 1. Method Kind Detection for Constructors, Getters, and Setters

**TypeScript Implementation:**
```typescript
parent.type === AST_NODE_TYPES.MethodDefinition &&
parent.kind === 'constructor'

// And for getters/setters:
parent.kind === 'constructor'
// method.Kind == ast.KindGetAccessor
// method.Kind == ast.KindSetAccessor
```

**Go Implementation:**
```go
parent.Kind == ast.KindMethodDeclaration
method := parent.AsMethodDeclaration()
if method.Kind == ast.KindConstructor {
    // ...
}
if method.Kind == ast.KindGetAccessor && isAllowed("getters") {
    return
}
if method.Kind == ast.KindSetAccessor && isAllowed("setters") {
    return
}
```

**Issue:** The Go implementation is checking `method.Kind` against constructor/getter/setter AST kinds, but these should likely be string comparisons or different AST node properties. In TypeScript AST, `kind` is a string property, while the Go version seems to be treating it as an AST node kind enum.

**Impact:** Constructor, getter, and setter detection may fail, causing the rule to incorrectly flag allowed empty functions.

**Test Coverage:** Test cases for constructors, getters, setters, and private/protected constructors would reveal this issue.

#### 2. Decorator Detection Mechanism

**TypeScript Implementation:**
```typescript
const decorators =
  node.parent.type === AST_NODE_TYPES.MethodDefinition
    ? node.parent.decorators
    : undefined;
return !!decorators && !!decorators.length;
```

**Go Implementation:**
```go
if ast.GetCombinedModifierFlags(parent)&ast.ModifierFlagsDecorator != 0 && isAllowed("decoratedFunctions") {
    return
}
```

**Issue:** The Go implementation uses `ast.ModifierFlagsDecorator` to detect decorators, but this may not be the correct way to detect decorators in the typescript-go AST. The TypeScript version explicitly checks for a `decorators` array property.

**Impact:** Decorated functions may not be properly detected as allowed empty functions.

**Test Coverage:** The decorator test case `@decorator() foo() {}` would fail.

#### 3. Generator Function Detection

**TypeScript Implementation:**
```typescript
function hasParameterProperties(
  node: TSESTree.FunctionDeclaration | TSESTree.FunctionExpression,
): boolean {
  return node.params.some(
    param => param.type === AST_NODE_TYPES.TSParameterProperty,
  );
}
```

**Go Implementation:**
```go
isGenerator = fn.AsteriskToken != nil
```

**Issue:** The Go implementation checks for `AsteriskToken` to detect generator functions, but this approach may not work correctly across all function types. The parameter properties detection also uses modifier flags instead of checking for `TSParameterProperty` node types.

**Impact:** Generator functions and parameter properties may not be detected correctly.

**Test Coverage:** Generator function tests and constructor parameter property tests would reveal this.

#### 4. Arrow Function Body Type Checking

**TypeScript Implementation:**
```typescript
// Base rule handles arrow functions
// No explicit arrow function body checking in the extended logic
```

**Go Implementation:**
```go
if fn.Body == nil {
    return false
}
if fn.Body.Kind != ast.KindBlock {
    return false // Expression body, not empty
}
block := fn.Body.AsBlock()
return len(block.Statements.Nodes) == 0
```

**Issue:** The Go implementation has complex arrow function body checking that distinguishes between block and expression bodies, but this logic may be overly complex compared to the TypeScript version which relies on the base rule.

**Impact:** Arrow functions with expression bodies might be incorrectly handled.

**Test Coverage:** Arrow function test cases would reveal discrepancies.

#### 5. Method Definition vs Method Declaration Confusion

**TypeScript Implementation:**
```typescript
parent.type === AST_NODE_TYPES.MethodDefinition
```

**Go Implementation:**
```go
parent.Kind == ast.KindMethodDeclaration
```

**Issue:** The TypeScript uses `MethodDefinition` while Go uses `MethodDeclaration`. These may not be equivalent AST node types in the respective parsers.

**Impact:** Method detection may fail entirely, causing all method-related logic to not work.

**Test Coverage:** All method-related test cases would be affected.

#### 6. Parameter Properties Detection Logic

**TypeScript Implementation:**
```typescript
return node.params.some(
  param => param.type === AST_NODE_TYPES.TSParameterProperty,
);
```

**Go Implementation:**
```go
if ast.GetCombinedModifierFlags(param)&(ast.ModifierFlagsPublic|ast.ModifierFlagsPrivate|ast.ModifierFlagsProtected|ast.ModifierFlagsReadonly) != 0 {
    return true
}
```

**Issue:** The TypeScript version checks for `TSParameterProperty` node type, while the Go version checks for modifier flags. These are different approaches that may not be equivalent.

**Impact:** Constructor parameter properties may not be detected correctly, causing valid empty constructors to be flagged.

**Test Coverage:** The `constructor(private name: string) {}` test case would reveal this.

### Recommendations
- Verify the correct AST node types and properties for method definitions in typescript-go
- Implement proper decorator detection using the correct AST properties
- Fix the method kind detection to use appropriate string or enum comparisons
- Simplify arrow function body checking to match TypeScript behavior
- Correct parameter properties detection to check for the right node types or properties
- Add more comprehensive test cases to validate all allowed function types
- Test edge cases around async/generator combinations with different method types
- Verify that the base rule integration works correctly for the core empty function detection

---