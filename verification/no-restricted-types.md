## Rule: no-restricted-types

### Test File: no-restricted-types.test.ts

### Validation Summary
- ✅ **CORRECT**: Core banned type checking logic, keyword type listeners, option parsing structure, error message formatting, fix and suggestion handling
- ⚠️ **POTENTIAL ISSUES**: Heritage clause handling implementation differs, AST navigation patterns may not be equivalent
- ❌ **INCORRECT**: Missing union and intersection type checking, missing type assertion handling, incomplete heritage clause coverage

### Discrepancies Found

#### 1. Missing Union and Intersection Type Support
**TypeScript Implementation:**
```typescript
// No explicit union/intersection handlers, but the generic node checking
// and TSTypeReference handler would catch these patterns
checkBannedTypes(node.typeName); // Checks type references in unions/intersections
```

**Go Implementation:**
```go
// No listeners for union or intersection types
// Missing: ast.KindUnionType and ast.KindIntersectionType listeners
```

**Issue:** The Go implementation lacks explicit handling for union types (`Banned | {}`) and intersection types (`Banned & {}`), which are covered by test cases.

**Impact:** Test cases like `'type Union = Banned | {};'` and `'type Intersection = Banned & {};'` may not trigger the rule properly.

**Test Coverage:** Tests for union and intersection types would fail.

#### 2. Missing Type Assertion Support
**TypeScript Implementation:**
```typescript
// Implicit coverage through generic AST traversal and TSTypeReference
// Type assertions like "1 as Banned" are caught when visiting type nodes
```

**Go Implementation:**
```go
// No listener for ast.KindTypeAssertion or ast.KindAsExpression
// Missing: type assertion handling
```

**Issue:** Type assertions like `1 as Banned;` are not explicitly handled in the Go implementation.

**Impact:** The test case `'1 as Banned;'` may not be caught by the rule.

**Test Coverage:** Type assertion test case would fail.

#### 3. Heritage Clause Implementation Differences
**TypeScript Implementation:**
```typescript
TSClassImplements(node): void {
  checkBannedTypes(node);
},
TSInterfaceHeritage(node): void {
  checkBannedTypes(node);
}
```

**Go Implementation:**
```go
// Manual traversal of heritage clauses in class and interface declarations
listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
  // Complex manual parsing of heritage clauses
}
listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
  // Complex manual parsing of heritage clauses  
}
```

**Issue:** The TypeScript version uses specific AST node types for heritage clauses, while Go manually traverses class/interface declarations. This could miss edge cases or have different behavior.

**Impact:** May affect reliability of catching banned types in implements/extends clauses.

**Test Coverage:** Heritage clause tests might pass but with different execution paths.

#### 4. Dynamic Listener Registration Issue
**TypeScript Implementation:**
```typescript
const keywordSelectors = objectReduceKey(
  TYPE_KEYWORDS,
  (acc: TSESLint.RuleListener, keyword) => {
    if (bannedTypes.has(keyword)) {
      acc[TYPE_KEYWORDS[keyword]] = (node: TSESTree.Node): void =>
        checkBannedTypes(node, keyword);
    }
    return acc;
  },
  {},
);
```

**Go Implementation:**
```go
for keyword, kind := range typeKeywords {
  if _, exists := opts.Types[keyword]; exists {
    listeners[kind] = func(node *ast.Node) {
      // Closure variable capture issue - all closures capture final keyword value
      for k, v := range typeKeywords {
        if v == node.Kind {
          checkBannedTypes(node, k)
          break
        }
      }
    }
  }
}
```

**Issue:** The Go implementation has a potential closure variable capture problem and uses inefficient reverse lookup instead of capturing the keyword directly.

**Impact:** May cause incorrect keyword identification or performance issues.

**Test Coverage:** Keyword type tests may behave unexpectedly.

#### 5. String-based Configuration vs Structured Configuration
**TypeScript Implementation:**
```typescript
// Handles both string and object configurations cleanly
if (typeof bannedType === 'string') {
  return ` ${bannedType}`;
}
if (bannedType.message) {
  return ` ${bannedType.message}`;
}
```

**Go Implementation:**
```go
// Complex type switching with potential edge cases
switch v := bannedConfig.(type) {
case bool:
  if v { return "" }
case string:
  return " " + v
case map[string]interface{}:
  // Complex nested type assertions
}
```

**Issue:** The Go implementation's type switching is more complex and may not handle all edge cases that the TypeScript version handles gracefully.

**Impact:** Configuration parsing edge cases might behave differently.

**Test Coverage:** Complex configuration test cases might fail.

#### 6. Missing Support for Generic Type Parameters
**TypeScript Implementation:**
```typescript
TSTypeReference(node): void {
  checkBannedTypes(node.typeName);
  if (node.typeArguments) {
    checkBannedTypes(node); // Checks full generic type
  }
}
```

**Go Implementation:**
```go
listeners[ast.KindTypeReference] = func(node *ast.Node) {
  // Checks type name and full type with arguments
  // But may not handle complex nested generics properly
}
```

**Issue:** The Go implementation may not properly handle complex generic type patterns like nested type arguments.

**Impact:** Test cases with complex generics like `Banned<A,B>` might not work correctly.

**Test Coverage:** Generic type parameter tests could fail.

### Recommendations
- Add explicit listeners for `ast.KindUnionType` and `ast.KindIntersectionType`
- Add listener for `ast.KindTypeAssertion` to handle type assertions
- Fix the closure variable capture issue in keyword listener registration
- Simplify heritage clause handling to match TypeScript patterns more closely
- Add comprehensive test coverage for edge cases in configuration parsing
- Enhance generic type parameter handling for complex nested cases
- Consider using dedicated AST node types for heritage clauses if available in the Go AST

---