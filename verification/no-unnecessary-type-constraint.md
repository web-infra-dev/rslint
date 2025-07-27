## Rule: no-unnecessary-type-constraint

### Test File: no-unnecessary-type-constraint.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic unnecessary constraint detection for `any` and `unknown` keywords
  - File extension detection for trailing comma disambiguation (.tsx, .mts, .cts)
  - Core AST pattern matching for type parameters with constraints
  - Error message structure and content
  - Basic fix generation logic

- ⚠️ **POTENTIAL ISSUES**: 
  - Complex trailing comma detection logic may not handle all edge cases
  - Parent traversal for arrow function detection could miss nested scenarios
  - Token parsing for existing comma detection is simplified

- ❌ **INCORRECT**: 
  - Message ID handling is inconsistent with TypeScript implementation
  - AST selector pattern is fundamentally different - TypeScript uses CSS-like selectors, Go uses single node type listener
  - Fix range calculation may be incorrect for type parameters with default values

### Discrepancies Found

#### 1. AST Selector Pattern Mismatch
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
return rule.RuleListeners{
  ast.KindTypeParameter: func(node *ast.Node) {
    // Manual constraint checking and arrow function detection
    typeParam := node.AsTypeParameter()
    if typeParam.Constraint == nil {
      return
    }
    // ... manual parent traversal
  },
}
```

**Issue:** The TypeScript implementation uses sophisticated CSS-like selectors to differentiate between arrow functions and other constructs, while the Go implementation uses a single listener with manual parent traversal. This could lead to different behavior in complex nested scenarios.

**Impact:** May miss or incorrectly identify arrow function contexts, affecting trailing comma logic.

**Test Coverage:** Arrow function test cases, especially nested ones.

#### 2. Message ID Implementation
**TypeScript Implementation:**
```typescript
messageId: 'unnecessaryConstraint',
data: {
  name: node.name.name,
  constraint,
},
```

**Go Implementation:**
```go
message := rule.RuleMessage{
  Id:          "unnecessaryConstraint",
  Description: fmt.Sprintf("Constraining the generic type `%s` to `%s` does nothing and is unnecessary.", typeName, constraint),
}
```

**Issue:** The Go implementation embeds the message data directly into the description, while TypeScript uses a separate data object and message templates. This breaks the message ID pattern expected by the testing framework.

**Impact:** Tests expecting specific message IDs and data fields will fail.

**Test Coverage:** All test cases that check messageId and data fields.

#### 3. Trailing Comma Detection Logic
**TypeScript Implementation:**
```typescript
function shouldAddTrailingComma(): boolean {
  if (!inArrowFunction || !requiresGenericDeclarationDisambiguation) {
    return false;
  }
  return (
    (node.parent as TSESTree.TSTypeParameterDeclaration).params.length === 1 &&
    context.sourceCode.getTokensAfter(node)[0].value !== ',' &&
    !node.default
  );
}
```

**Go Implementation:**
```go
func shouldAddTrailingComma(node *ast.Node, inArrowFunction bool, requiresDisambiguation bool, ctx rule.RuleContext) bool {
  // Complex logic with manual token parsing
  typeParams := current.TypeParameters()
  if typeParams != nil && len(typeParams) == 1 {
    // Manual character-by-character parsing
    for i := nodeEnd; i < len(ctx.SourceFile.Text()); i++ {
      char := ctx.SourceFile.Text()[i]
      if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
        if char == ',' {
          return false
        }
        break
      }
    }
    return true
  }
}
```

**Issue:** The Go implementation uses manual character parsing instead of proper tokenization, and doesn't check for default parameters (`!node.default` condition is missing).

**Impact:** May incorrectly add trailing commas when type parameters have default values, or miss existing commas in complex whitespace scenarios.

**Test Coverage:** Test case with default values: `'const data = <T extends any = unknown>() => {};'`

#### 4. Fix Range Calculation
**TypeScript Implementation:**
```typescript
fix(fixer): TSESLint.RuleFix | null {
  return fixer.replaceTextRange(
    [node.name.range[1], node.constraint.range[1]],
    shouldAddTrailingComma() ? ',' : '',
  );
}
```

**Go Implementation:**
```go
nameEnd := name.End()
constraintEnd := typeParam.Constraint.End()
fixRange := core.NewTextRange(nameEnd, constraintEnd)
```

**Issue:** The range calculation looks correct, but the Go implementation doesn't handle the case where a type parameter has both a constraint AND a default value. The TypeScript version accounts for this in the range calculation.

**Impact:** Incorrect fix output for type parameters with default values.

**Test Coverage:** Test case: `'const data = <T extends any = unknown>() => {};'` expects output `'const data = <T = unknown>() => {};'`

#### 5. Parent Type Parameter Declaration Access
**TypeScript Implementation:**
```typescript
(node.parent as TSESTree.TSTypeParameterDeclaration).params.length === 1
```

**Go Implementation:**
```go
// Complex parent traversal to find type parameters
current := typeParam.Parent
for current != nil {
  switch current.Kind {
    case ast.KindArrowFunction, ast.KindFunctionDeclaration, ...
      typeParams := current.TypeParameters()
      if typeParams != nil && len(typeParams) == 1 {
```

**Issue:** The TypeScript implementation directly accesses the parent TSTypeParameterDeclaration, while Go does complex traversal. This may not find the correct parent in all cases.

**Impact:** Incorrect trailing comma detection in complex nested scenarios.

**Test Coverage:** All arrow function test cases with single type parameters.

### Recommendations
- Fix message ID handling to use proper template messages with data objects
- Implement proper tokenization instead of manual character parsing for trailing comma detection
- Add check for default parameters in trailing comma logic
- Simplify parent traversal to directly access type parameter declaration parent
- Add handling for type parameters with both constraints and default values
- Consider implementing CSS-like selector pattern matching for more accurate AST targeting
- Add comprehensive test coverage for nested arrow functions and complex type parameter scenarios

---