## Rule: prefer-as-const

### Test File: prefer-as-const.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic literal type assertion detection, `as` and type assertion expression handling, core message structure
- ⚠️ **POTENTIAL ISSUES**: Template literal handling, fix application logic, suggestion implementation complexity
- ❌ **INCORRECT**: Variable declarator AST mapping, property definition AST mapping, fix range calculation

### Discrepancies Found

#### 1. Incorrect AST Node Mapping for Variable Declarations
**TypeScript Implementation:**
```typescript
VariableDeclarator(node): void {
  if (node.init && node.id.typeAnnotation) {
    compareTypes(node.init, node.id.typeAnnotation.typeAnnotation, false);
  }
}
```

**Go Implementation:**
```go
ast.KindVariableDeclaration: func(node *ast.Node) {
  if node.Kind != ast.KindVariableDeclaration {
    return
  }
  varDecl := node.AsVariableDeclaration()
  if varDecl.Initializer != nil && varDecl.Type != nil {
    compareTypes(varDecl.Initializer, varDecl.Type, false)
  }
}
```

**Issue:** The TypeScript version listens to `VariableDeclarator` nodes (individual variable bindings within a declaration), while the Go version listens to `VariableDeclaration` nodes (the entire declaration statement). This means the Go version might miss cases with multiple declarators or have different access patterns to the type annotation.

**Impact:** Test cases like `let foo: 'bar' = 'bar';` may not be properly detected or may fail due to incorrect AST navigation.

**Test Coverage:** This affects multiple test cases including basic variable declarations with type annotations.

#### 2. Incorrect AST Node Mapping for Class Properties
**TypeScript Implementation:**
```typescript
PropertyDefinition(node): void {
  if (node.value && node.typeAnnotation) {
    compareTypes(node.value, node.typeAnnotation.typeAnnotation, false);
  }
}
```

**Go Implementation:**
```go
ast.KindPropertyDeclaration: func(node *ast.Node) {
  if node.Kind != ast.KindPropertyDeclaration {
    return
  }
  propDecl := node.AsPropertyDeclaration()
  if propDecl.Initializer != nil && propDecl.Type != nil {
    compareTypes(propDecl.Initializer, propDecl.Type, false)
  }
}
```

**Issue:** The TypeScript version accesses `node.typeAnnotation.typeAnnotation` (nested structure), while the Go version accesses `propDecl.Type` directly. This suggests a difference in how type annotations are represented in the AST structures.

**Impact:** Class property test cases like `class foo { bar: 'baz' = 'baz'; }` may not work correctly.

**Test Coverage:** This affects all class property test cases in the invalid array.

#### 3. Template Literal Exclusion Logic
**TypeScript Implementation:**
```typescript
// No explicit template literal exclusion in compareTypes
```

**Go Implementation:**
```go
// Skip template literal types - they are different from regular literal types
if literalNode.Kind == ast.KindNoSubstitutionTemplateLiteral {
  return
}
```

**Issue:** The Go version explicitly excludes template literals, but the TypeScript version doesn't show this exclusion. This could lead to different behavior for template literal cases.

**Impact:** Template literal test cases like `let foo = \`bar\` as \`bar\`;` might behave differently between implementations.

**Test Coverage:** Valid test cases with template literals may be incorrectly flagged or vice versa.

#### 4. Complex Fix Range Calculation for Suggestions
**TypeScript Implementation:**
```typescript
suggest: [
  {
    messageId: 'variableSuggest',
    fix: (fixer): TSESLint.RuleFix[] => [
      fixer.remove(typeNode.parent),
      fixer.insertTextAfter(valueNode, ' as const'),
    ],
  },
]
```

**Go Implementation:**
```go
s := scanner.GetScannerForSourceFile(ctx.SourceFile, parent.Pos())
colonStart := -1
for s.TokenStart() < typeNode.Pos() {
  if s.Token() == ast.KindColonToken {
    colonStart = s.TokenStart()
  }
  s.Scan()
}

if colonStart != -1 {
  ctx.ReportNodeWithSuggestions(literalNode, buildVariableConstAssertionMessage(),
    rule.RuleSuggestion{
      Message: buildVariableSuggestMessage(),
      FixesArr: []rule.RuleFix{
        rule.RuleFixReplaceRange(
          core.NewTextRange(colonStart, typeNode.End()),
          "",
        ),
        rule.RuleFixInsertAfter(valueNode, " as const"),
      },
    })
}
```

**Issue:** The TypeScript version uses `fixer.remove(typeNode.parent)` to remove the entire type annotation, while the Go version manually scans for the colon token. This could lead to different removal ranges and potentially incorrect fixes.

**Impact:** Suggestion fixes for variable declarations may not work correctly or may produce malformed code.

**Test Coverage:** Test cases with suggestions like `let foo: 'bar' = 'bar';` may have incorrect output.

#### 5. Fix Reporting Target Inconsistency
**TypeScript Implementation:**
```typescript
context.report({
  node: typeNode,
  messageId: 'preferConstAssertion',
  fix: fixer => fixer.replaceText(typeNode, 'const'),
});
```

**Go Implementation:**
```go
ctx.ReportNodeWithFixes(literalNode, buildPreferConstAssertionMessage(),
  rule.RuleFixReplace(ctx.SourceFile, typeNode, "const"))
```

**Issue:** The TypeScript version reports on `typeNode` and replaces `typeNode`, while the Go version reports on `literalNode` but replaces `typeNode`. This inconsistency could affect error positioning and highlighting.

**Impact:** Error messages may appear at incorrect locations in the code.

**Test Coverage:** Column positions in test expectations may not match actual error positions.

### Recommendations
- Fix AST node type mapping: Use the correct Go AST equivalents for `VariableDeclarator` and `PropertyDefinition`
- Verify type annotation access patterns: Ensure the Go version correctly navigates nested type annotation structures
- Align template literal handling: Verify whether template literal exclusion is needed and implement consistently
- Simplify fix range calculation: Use more reliable methods for determining type annotation ranges for removal
- Standardize error reporting targets: Ensure consistent node targeting between error reporting and fix application
- Add comprehensive AST debugging: Use debug output to verify correct node types and structures are being processed

---