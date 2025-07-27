## Rule: consistent-type-assertions

### Test File: consistent-type-assertions.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic assertion style enforcement, object/array literal detection, const assertion special handling, message ID structure
- ⚠️ **POTENTIAL ISSUES**: AST node kind mapping, parameter detection logic, qualified name handling, fix generation for style conversion
- ❌ **INCORRECT**: Missing auto-fix functionality for style conversion, incomplete parameter context detection, suggestion implementation gaps

### Discrepancies Found

#### 1. Missing Auto-Fix for Style Conversion
**TypeScript Implementation:**
```typescript
fix:
  messageId === 'as'
    ? (fixer): TSESLint.RuleFix => {
        // Complex precedence-aware fix logic
        const tsNode = getParserServices(context, true).esTreeNodeToTSNodeMap.get(node);
        const expressionCode = context.sourceCode.getText(node.expression);
        const typeAnnotationCode = context.sourceCode.getText(node.typeAnnotation);
        
        const asPrecedence = getOperatorPrecedence(ts.SyntaxKind.AsExpression, ts.SyntaxKind.Unknown);
        const parentPrecedence = getOperatorPrecedence(tsNode.parent.kind, ...);
        
        const expressionCodeWrapped = getWrappedCode(expressionCode, expressionPrecedence, asPrecedence);
        const text = `${expressionCodeWrapped} as ${typeAnnotationCode}`;
        return fixer.replaceText(node, isParenthesized(node, context.sourceCode) ? text : getWrappedCode(text, asPrecedence, parentPrecedence));
      }
    : undefined,
```

**Go Implementation:**
```go
func reportIncorrectAssertionType(ctx rule.RuleContext, node *ast.Node, options Options, isAsExpression bool) {
    // Reports message but no fix implementation
    switch options.AssertionStyle {
    case "as":
        cast := getTypeAnnotationText(ctx, typeAnnotation)
        // For angle-bracket to as conversion, we'd need complex fix logic
        // For now, just report without fix
        ctx.ReportNode(node, buildAsMessage(cast))
    }
}
```

**Issue:** The Go implementation lacks the sophisticated auto-fix functionality for converting between assertion styles, particularly the precedence-aware parentheses handling.

**Impact:** Users won't get automatic fixes when using wrong assertion style, reducing developer experience.

**Test Coverage:** All invalid test cases with `output` property expecting auto-fixes.

#### 2. Incomplete Parameter Context Detection
**TypeScript Implementation:**
```typescript
function isAsParameter(node: AsExpressionOrTypeAssertion): boolean {
  return (
    node.parent.type === AST_NODE_TYPES.NewExpression ||
    node.parent.type === AST_NODE_TYPES.CallExpression ||
    node.parent.type === AST_NODE_TYPES.ThrowStatement ||
    node.parent.type === AST_NODE_TYPES.AssignmentPattern ||
    node.parent.type === AST_NODE_TYPES.JSXExpressionContainer ||
    (node.parent.type === AST_NODE_TYPES.TemplateLiteral &&
      node.parent.parent.type === AST_NODE_TYPES.TaggedTemplateExpression)
  );
}
```

**Go Implementation:**
```go
func isAsParameter(node *ast.Node) bool {
    // Missing AssignmentPattern and JSXExpressionContainer checks
    switch parent.Kind {
    case ast.KindNewExpression, ast.KindCallExpression, ast.KindThrowStatement:
        return true
    case ast.KindJsxExpression:
        return true
    // Missing: AssignmentPattern equivalent
    // Missing: Proper TemplateLiteral + TaggedTemplate check
    }
}
```

**Issue:** Go implementation doesn't detect all parameter contexts, particularly assignment patterns (default parameters) and has incomplete template literal handling.

**Impact:** Rules with `allow-as-parameter` option may incorrectly flag valid parameter usages.

**Test Coverage:** Test cases with default parameters like `function b(x = {} as Foo.Bar) {}`.

#### 3. AST Node Kind Mapping Issues
**TypeScript Implementation:**
```typescript
return {
  TSAsExpression(node): void { /* ... */ },
  TSTypeAssertion(node): void { /* ... */ }
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
    ast.KindAsExpression: func(node *ast.Node) { /* ... */ },
    ast.KindTypeAssertionExpression: func(node *ast.Node) { /* ... */ },
}
```

**Issue:** Need to verify that `ast.KindTypeAssertionExpression` correctly maps to TypeScript's angle-bracket assertions (`<Type>expr`).

**Impact:** Rule might not trigger on angle-bracket assertions if the AST kind mapping is incorrect.

**Test Coverage:** All test cases using angle-bracket syntax like `<Foo>expr`.

#### 4. Qualified Name Handling in checkType Function
**TypeScript Implementation:**
```typescript
function checkType(node: TSESTree.TypeNode): boolean {
  switch (node.type) {
    case AST_NODE_TYPES.TSTypeReference:
      return (
        // Ignore `as const` and `<const>`
        !isConst(node) ||
        // Allow qualified names which have dots between identifiers, `Foo.Bar`
        node.typeName.type === AST_NODE_TYPES.TSQualifiedName
      );
  }
}
```

**Go Implementation:**
```go
func checkType(node *ast.Node) bool {
    switch node.Kind {
    case ast.KindTypeReference:
        // For type references, check if it's `const`
        if isConst(node) {
            return false
        }
        // Also check for qualified names with dots (e.g., Foo.Bar)
        typeRef := node.AsTypeReferenceNode()
        if typeRef != nil && typeRef.TypeName != nil && ast.IsQualifiedName(typeRef.TypeName) {
            return true
        }
        return true
    }
}
```

**Issue:** The logic differs - TypeScript uses OR (`!isConst(node) || node.typeName.type === AST_NODE_TYPES.TSQualifiedName`) while Go uses separate if statements that change the behavior.

**Impact:** Qualified const types like `Foo.Bar.const` might be handled differently.

**Test Coverage:** Test cases with qualified type names in assertions.

#### 5. Suggestion Implementation Gaps
**TypeScript Implementation:**
```typescript
function getSuggestions(node, annotationMessageId, satisfiesMessageId): TSESLint.ReportSuggestionArray<MessageIds> {
  const suggestions: TSESLint.ReportSuggestionArray<MessageIds> = [];
  if (node.parent.type === AST_NODE_TYPES.VariableDeclarator && !node.parent.id.typeAnnotation) {
    // Add annotation suggestion with proper fix logic
    suggestions.push({
      messageId: annotationMessageId,
      data: { cast: context.sourceCode.getText(node.typeAnnotation) },
      fix: fixer => [
        fixer.insertTextAfter(parent.id, `: ${context.sourceCode.getText(node.typeAnnotation)}`),
        fixer.replaceText(node, getTextWithParentheses(context.sourceCode, node.expression)),
      ],
    });
  }
  // Always add satisfies suggestion
  suggestions.push({...});
  return suggestions;
}
```

**Go Implementation:**
```go
func getSuggestions(ctx rule.RuleContext, node *ast.Node, isAsExpression bool, annotationMessageId, satisfiesMessageId string) []rule.RuleSuggestion {
    // Check if this is a variable declarator that can have type annotation
    parent := node.Parent
    if parent != nil && parent.Kind == ast.KindVariableDeclaration {
        // Missing proper type annotation check (!node.parent.id.typeAnnotation)
        // Fix implementation is basic compared to TypeScript version
    }
}
```

**Issue:** Go implementation doesn't properly check if variable already has type annotation and the fix logic is incomplete.

**Impact:** Suggestions might be offered even when not appropriate, or fixes might not work correctly.

**Test Coverage:** Test cases expecting suggestions for variable declaration conversions.

#### 6. Missing `getTextWithParentheses` Utility
**TypeScript Implementation:**
```typescript
fixer.replaceText(node, getTextWithParentheses(context.sourceCode, node.expression))
```

**Go Implementation:**
```go
rule.RuleFixReplace(ctx.SourceFile, node, getExpressionText(ctx, expression))
```

**Issue:** Go implementation uses simple text extraction without considering when parentheses are needed around expressions.

**Impact:** Generated fixes might produce syntactically incorrect code when parentheses are required.

**Test Coverage:** Complex expression cases in suggestions.

### Recommendations
- Implement comprehensive auto-fix functionality with proper precedence handling
- Complete the `isAsParameter` function to detect all parameter contexts including assignment patterns
- Verify and fix AST node kind mappings for angle-bracket assertions
- Correct the `checkType` function logic to match TypeScript behavior exactly
- Enhance suggestion implementation with proper type annotation checking
- Add parentheses-aware text extraction utilities for fix generation
- Add comprehensive test coverage for edge cases around parameter detection
- Implement proper JSX support detection and handling

---