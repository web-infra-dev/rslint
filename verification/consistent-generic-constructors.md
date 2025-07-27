## Rule: consistent-generic-constructors

### Test File: consistent-generic-constructors.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core rule logic for detecting type annotation vs constructor generic preferences
  - Basic AST pattern matching for variable declarations, property declarations, and parameters
  - Error message IDs and descriptions match exactly
  - Options parsing supports both string and array formats
  - Type reference matching logic with same constructor name validation
  - Basic fix generation for moving generics between locations
  - Support for accessor properties and binding elements

- ⚠️ **POTENTIAL ISSUES**: 
  - Comment preservation during fixes is simplified compared to TypeScript version
  - Computed property handling may not match exact positioning for type annotation attachment
  - Error reporting node selection differs for some cases (variable statements vs declarations)
  - isolatedDeclarations option checking relies on Program.Options() which may not match parser options exactly

- ❌ **INCORRECT**: 
  - Missing proper handling of assignment patterns in destructuring contexts outside function parameters
  - AST node kind mapping doesn't fully match TypeScript ESLint's selector pattern
  - Binding element filtering logic is too restrictive and may miss valid cases

### Discrepancies Found

#### 1. AST Node Selector Mismatch
**TypeScript Implementation:**
```typescript
'VariableDeclarator,PropertyDefinition,AccessorProperty,:matches(FunctionDeclaration,FunctionExpression) > AssignmentPattern'
```

**Go Implementation:**
```go
return rule.RuleListeners{
    ast.KindVariableDeclaration: handleNode,
    ast.KindPropertyDeclaration: handleNode,
    ast.KindParameter:           handleNode,
    ast.KindBindingElement:      handleNode,
}
```

**Issue:** The TypeScript version uses a complex CSS-like selector that includes `AssignmentPattern` specifically within function contexts, while the Go version uses `BindingElement` with custom filtering logic that may be too restrictive.

**Impact:** May miss assignment patterns in destructuring that should be checked, or may incorrectly process binding elements outside intended contexts.

**Test Coverage:** Test cases with destructuring assignments like `const [a = new Foo<string>()] = [];` may not work correctly.

#### 2. Comment Preservation in Fixes
**TypeScript Implementation:**
```typescript
const extraComments = new Set(
  context.sourceCode.getCommentsInside(lhs.parent),
);
context.sourceCode
  .getCommentsInside(lhs.typeArguments)
  .forEach(c => extraComments.delete(c));
context.report({
  *fix(fixer) {
    for (const comment of extraComments) {
      yield fixer.insertTextAfter(
        rhs.callee,
        context.sourceCode.getText(comment),
      );
    }
  }
});
```

**Go Implementation:**
```go
typeArgsText := getNodeListTextWithBrackets(ctx, lhsTypeArgs)
// Basic comment preservation - the type args text already includes any comments
```

**Issue:** The Go version has simplified comment preservation that just includes comments within the type arguments, but doesn't handle the sophisticated comment redistribution logic of the TypeScript version.

**Impact:** Comments may not be preserved correctly when moving generics between type annotation and constructor, leading to lost or misplaced comments.

**Test Coverage:** Test cases with comments like `const a: /* comment */ Foo/* another */ <string> = new Foo();` may not produce identical output.

#### 3. Computed Property Type Annotation Attachment
**TypeScript Implementation:**
```typescript
function getIDToAttachAnnotation(): TSESTree.Node | TSESTree.Token {
  if (node.computed) {
    return nullThrows(
      context.sourceCode.getTokenAfter(node.key),
      NullThrowsReasons.MissingToken(']', 'key'),
    );
  }
  return node.key;
}
```

**Go Implementation:**
```go
if propDecl.Name().Kind == ast.KindComputedPropertyName {
    // For computed properties, find the closing bracket token to match TypeScript behavior
    computed := propDecl.Name().AsComputedPropertyName()
    // Use scanner to find the closing bracket after the expression
    // TODO: Better position handling for after closing bracket
    return computed.AsNode() 
}
```

**Issue:** The Go version has incomplete handling for computed properties and doesn't properly find the closing bracket token position for type annotation attachment.

**Impact:** Type annotations may be attached at the wrong position for computed properties, leading to invalid syntax.

**Test Coverage:** Test cases like `class Foo { [a]: Foo<string> = new Foo(); }` may produce incorrect fix output positioning.

#### 4. Error Reporting Node Selection
**TypeScript Implementation:**
```typescript
context.report({
  node, // Reports on the original node (VariableDeclarator, PropertyDefinition, etc.)
```

**Go Implementation:**
```go
// Determine what node to report the error on
reportNode := node
if node.Kind == ast.KindVariableDeclaration && node.Parent != nil {
    reportNode = node.Parent // For variable declarations, report on the statement
}
```

**Issue:** The Go version changes the reporting node for variable declarations to the parent statement, while TypeScript reports on the declarator itself.

**Impact:** Error positioning and highlighting may differ from TypeScript ESLint output.

**Test Coverage:** Error location for variable declaration cases may not match expected positions.

#### 5. isolatedDeclarations Option Handling
**TypeScript Implementation:**
```typescript
const isolatedDeclarations = context.parserOptions.isolatedDeclarations;
```

**Go Implementation:**
```go
isolatedDeclarations := ctx.Program.Options().IsolatedDeclarations.IsTrue()
```

**Issue:** The Go version reads from Program options while TypeScript reads from parser options, which may not be the same value depending on how options are passed through.

**Impact:** The isolatedDeclarations feature gate may not work correctly, causing the rule to flag cases it should ignore.

**Test Coverage:** The commented-out test case for isolatedDeclarations cannot be validated.

#### 6. Assignment Pattern Context Filtering
**TypeScript Implementation:**
```typescript
':matches(FunctionDeclaration,FunctionExpression) > AssignmentPattern'
```

**Go Implementation:**
```go
if node.Kind == ast.KindBindingElement {
    // Only process binding elements that are function parameters
    current := node.Parent
    for current != nil {
        if current.Kind == ast.KindParameter {
            break
        }
        if current.Kind == ast.KindVariableDeclaration ||
            current.Kind == ast.KindVariableStatement ||
            current.Kind == ast.KindVariableDeclarationList {
            return
        }
        current = current.Parent
    }
    if current == nil {
        return
    }
}
```

**Issue:** The Go version's filtering logic is more complex but may still miss valid assignment patterns or incorrectly filter out cases that should be processed.

**Impact:** Some destructuring assignment patterns with defaults may not be processed correctly.

**Test Coverage:** Complex destructuring scenarios may behave differently.

### Recommendations
- Fix AST node selection to more accurately match TypeScript ESLint's selector behavior
- Improve comment preservation logic to handle sophisticated comment redistribution
- Complete the computed property type annotation attachment positioning
- Align error reporting node selection with TypeScript behavior
- Verify isolatedDeclarations option is correctly passed through from parser options
- Simplify and correct binding element filtering to match intended selector semantics
- Add more comprehensive test cases for edge cases like complex destructuring patterns
- Consider implementing a more direct mapping from TypeScript ESLint's CSS selectors to Go AST listeners

---