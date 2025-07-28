## Rule: no-unnecessary-type-constraint

### Test File: no-unnecessary-type-constraint.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic unnecessary constraint detection for `any` and `unknown` keywords, file extension-based disambiguation logic, core fix generation logic
- ⚠️ **POTENTIAL ISSUES**: Trailing comma detection logic complexity, parent traversal for arrow function detection, default type parameter handling
- ❌ **INCORRECT**: Message structure and data interpolation, suggestion vs fix mechanism, incomplete edge case handling for default type parameters

### Discrepancies Found

#### 1. Message Structure and Data Interpolation
**TypeScript Implementation:**
```typescript
context.report({
  node,
  messageId: 'unnecessaryConstraint',
  data: {
    name: node.name.name,
    constraint,
  },
  suggest: [
    {
      messageId: 'removeUnnecessaryConstraint',
      data: {
        constraint,
      },
      fix(fixer): TSESLint.RuleFix | null {
        // fix logic
      },
    },
  ],
});
```

**Go Implementation:**
```go
message := rule.RuleMessage{
  Id:          "unnecessaryConstraint",
  Description: fmt.Sprintf("Constraining the generic type `%s` to `%s` does nothing and is unnecessary.", typeName, constraint),
}

ctx.ReportNodeWithFixes(node, message, fix)
```

**Issue:** The Go version constructs error messages directly instead of using message IDs with data interpolation, and provides fixes directly instead of suggestions.

**Impact:** This affects consistency with TypeScript-ESLint's error reporting format and may not match expected test outputs.

**Test Coverage:** All test cases expect specific message structure with data interpolation.

#### 2. Default Type Parameter Handling
**TypeScript Implementation:**
```typescript
function shouldAddTrailingComma(): boolean {
  return (
    (node.parent as TSESTree.TSTypeParameterDeclaration).params.length === 1 &&
    context.sourceCode.getTokensAfter(node)[0].value !== ',' &&
    !node.default  // Critical check for default type parameters
  );
}
```

**Go Implementation:**
```go
func shouldAddTrailingComma(node *ast.Node, inArrowFunction bool, requiresDisambiguation bool, ctx rule.RuleContext) bool {
  // Missing check for default type parameters
  if !inArrowFunction || !requiresDisambiguation {
    return false
  }
  // ... rest of logic without default parameter check
}
```

**Issue:** The Go version doesn't check for default type parameters (`<T extends any = unknown>`), which should prevent trailing comma addition.

**Impact:** May incorrectly add trailing commas when type parameters have defaults, leading to invalid syntax.

**Test Coverage:** Test case `'const data = <T extends any = unknown>() => {};'` expects proper handling of default parameters.

#### 3. Trailing Comma Detection Logic
**TypeScript Implementation:**
```typescript
context.sourceCode.getTokensAfter(node)[0].value !== ','
```

**Go Implementation:**
```go
// Find the next token after the type parameter
for i := nodeEnd; i < len(ctx.SourceFile.Text()); i++ {
  char := ctx.SourceFile.Text()[i]
  if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
    if char == ',' {
      return false // Already has trailing comma
    }
    break
  }
}
```

**Issue:** The Go version uses manual character scanning instead of proper token analysis, which may be less reliable for complex whitespace scenarios.

**Impact:** May not correctly detect existing commas in all whitespace configurations.

**Test Coverage:** Multiple test cases with various whitespace patterns around commas.

#### 4. Parent Traversal for Arrow Function Detection
**TypeScript Implementation:**
```typescript
return {
  ':not(ArrowFunctionExpression) > TSTypeParameterDeclaration > TSTypeParameter[constraint]'(
    node: TypeParameterWithConstraint,
  ): void {
    checkNode(node, false);
  },
  'ArrowFunctionExpression > TSTypeParameterDeclaration > TSTypeParameter[constraint]'(
    node: TypeParameterWithConstraint,
  ): void {
    checkNode(node, true);
  },
};
```

**Go Implementation:**
```go
// Walk up the tree to find if we're in an arrow function
current := parent
for current != nil {
  if current.Kind == ast.KindArrowFunction {
    inArrowFunction = true
    break
  }
  // Stop if we hit a non-arrow function declaration
  if current.Kind == ast.KindFunctionDeclaration ||
     current.Kind == ast.KindFunctionExpression ||
     // ... other function types
  {
    break
  }
  current = current.Parent
}
```

**Issue:** The Go version uses manual parent traversal which may not correctly handle all AST structures, while TypeScript uses precise CSS-like selectors.

**Impact:** May incorrectly classify type parameters as being in or not in arrow functions, affecting trailing comma logic.

**Test Coverage:** Arrow function test cases expect specific trailing comma behavior.

#### 5. AST Node Access Pattern
**TypeScript Implementation:**
```typescript
(node.parent as TSESTree.TSTypeParameterDeclaration).params.length === 1
```

**Go Implementation:**
```go
typeParams := current.TypeParameters()
if typeParams != nil && len(typeParams) == 1 {
  // logic
}
```

**Issue:** The Go version searches for any parent with TypeParameters, while TypeScript directly accesses the immediate parent's params.

**Impact:** May count type parameters from the wrong scope or context.

**Test Coverage:** Test cases with multiple type parameters need accurate counting.

### Recommendations
- Implement proper message structure with ID-based messages and data interpolation to match TypeScript-ESLint format
- Add default type parameter detection to prevent incorrect trailing comma insertion
- Replace manual character scanning with proper token-based comma detection
- Use more precise AST pattern matching instead of manual parent traversal
- Add proper handling for all edge cases involving whitespace and existing commas
- Ensure message IDs match the expected test outputs exactly
- Consider implementing suggestion mechanism instead of direct fixes to match TypeScript behavior

---