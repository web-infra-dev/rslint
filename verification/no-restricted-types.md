## Rule: no-restricted-types

### Test File: no-restricted-types.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic type keyword detection (string, number, boolean, etc.)
  - Empty tuple `[]` detection
  - Empty type literal `{}` detection
  - Type reference checking with qualified names (NS.Banned)
  - Space normalization in type names
  - Basic fix and suggestion support
  - Class implements clause checking
  - Interface extends clause checking
  - Configuration parsing for different ban config formats

- ⚠️ **POTENTIAL ISSUES**:
  - AST node listener registration pattern may miss dynamic cases
  - Type argument checking logic might not match all cases

- ❌ **INCORRECT**:
  - Missing support for type assertions (`1 as Banned`)
  - Missing support for union types (`Banned | {}`)
  - Missing support for intersection types (`Banned & {}`)
  - Missing support for array types (`Banned[]`)
  - Missing support for tuple element types (`[Banned]`)
  - Missing support for property types in type literals (`{ c: Banned }`)

### Discrepancies Found

#### 1. Type Assertion Support Missing
**TypeScript Implementation:**
```typescript
return {
  ...keywordSelectors,
  // ... other listeners that catch all type nodes
};
```

**Go Implementation:**
```go
listeners := rule.RuleListeners{}
// Only has specific listeners for certain node types
```

**Issue:** The Go implementation doesn't listen for type assertion nodes (`1 as Banned`), which means it won't catch banned types used in type assertions.

**Impact:** Test case `'1 as Banned;'` would not be caught by the Go implementation.

**Test Coverage:** Test case with `code: '1 as Banned;'` would fail.

#### 2. Union and Intersection Type Support Missing
**TypeScript Implementation:**
```typescript
// The TypeScript implementation catches all type nodes through comprehensive listeners
```

**Go Implementation:**
```go
// Missing listeners for:
// - ast.KindUnionType
// - ast.KindIntersectionType
```

**Issue:** The Go implementation doesn't check banned types within union (`Banned | {}`) or intersection (`Banned & {}`) types.

**Impact:** Test cases like `'type Union = Banned | {};'` and `'type Intersection = Banned & {};'` would not be caught.

**Test Coverage:** Multiple test cases with union and intersection types would fail.

#### 3. Array Type Support Missing
**TypeScript Implementation:**
```typescript
// Catches all type references including array types
TSTypeReference(node): void {
  checkBannedTypes(node.typeName);
  // ...
}
```

**Go Implementation:**
```go
// Missing listener for ast.KindArrayType
```

**Issue:** The Go implementation doesn't check for banned types used as array element types (`Banned[]`).

**Impact:** Test case `'let value: Banned[];'` would not be caught.

**Test Coverage:** Test case with `code: 'let value: Banned[];'` would fail.

#### 4. Tuple Element Type Support Missing
**TypeScript Implementation:**
```typescript
// The comprehensive listener approach catches tuple element types
```

**Go Implementation:**
```go
listeners[ast.KindTupleType] = func(node *ast.Node) {
  // Only checks if tuple is empty, doesn't check element types
}
```

**Issue:** The Go implementation only checks for empty tuples but doesn't recursively check tuple element types for banned types.

**Impact:** Test case `'let value: [Banned];'` would not be caught.

**Test Coverage:** Test case with `code: 'let value: [Banned];'` would fail.

#### 5. Property Type Support Missing
**TypeScript Implementation:**
```typescript
// Catches all type nodes including property types in type literals
```

**Go Implementation:**
```go
listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
  // Only checks if type literal is empty, doesn't check member types
}
```

**Issue:** The Go implementation doesn't recursively check property types within type literals.

**Impact:** Test case `'let b: { c: Banned };'` would not be caught.

**Test Coverage:** Test case with `code: 'let b: { c: Banned };'` would fail.

#### 6. Incomplete AST Coverage Strategy
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

return {
  ...keywordSelectors,
  TSClassImplements(node): void { checkBannedTypes(node); },
  TSInterfaceHeritage(node): void { checkBannedTypes(node); },
  TSTupleType(node): void { /* ... */ },
  TSTypeLiteral(node): void { /* ... */ },
  TSTypeReference(node): void { /* ... */ },
};
```

**Go Implementation:**
```go
listeners := rule.RuleListeners{}

// Add listeners for keyword types
for keyword, kind := range typeKeywords {
  if _, exists := opts.Types[keyword]; exists {
    listeners[kind] = func(node *ast.Node) { /* ... */ }
  }
}
```

**Issue:** The Go implementation only adds keyword listeners if they exist in the banned types configuration, but the TypeScript version has a more comprehensive approach that catches type nodes in various contexts.

**Impact:** Many test cases involving banned types in complex type expressions would not be caught.

**Test Coverage:** Multiple test cases would fail due to incomplete AST coverage.

### Recommendations
- Add listeners for missing AST node types:
  - `ast.KindTypeAssertion` for type assertions
  - `ast.KindUnionType` for union types  
  - `ast.KindIntersectionType` for intersection types
  - `ast.KindArrayType` for array types
- Implement recursive type checking for complex type expressions
- Add comprehensive traversal of tuple elements and type literal members
- Consider a more general approach that visits all type nodes rather than specific node types
- Add proper handling for all contexts where types can appear
- Ensure the keyword listener registration works for all cases, not just when types are pre-configured

---