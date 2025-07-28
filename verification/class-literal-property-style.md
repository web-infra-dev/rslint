## Rule: class-literal-property-style

### Test File: class-literal-property-style.test.ts

### Validation Summary
- ✅ **CORRECT**: Core rule logic structure, option parsing, literal detection, modifier handling, basic getter-to-field and field-to-getter conversions
- ⚠️ **POTENTIAL ISSUES**: Constructor detection pattern differences, nested class handling, abstract method detection 
- ❌ **INCORRECT**: Setter duplicate detection logic, computed property name handling, AST node type checking

### Discrepancies Found

#### 1. Setter Duplicate Detection Logic
**TypeScript Implementation:**
```typescript
const hasDuplicateKeySetter =
  name &&
  node.parent.body.some(element => {
    return (
      element.type === AST_NODE_TYPES.MethodDefinition &&
      element.kind === 'set' &&
      isStaticMemberAccessOfValue(element, context, name)
    );
  });
```

**Go Implementation:**
```go
if name != "" && node.Parent != nil {
  members := node.Parent.Members()
  if members != nil {
    for _, member := range members {
      if ast.IsSetAccessorDeclaration(member) && isStaticMemberAccessOfValue(ctx, member, name) {
        return // Skip if there's a setter with the same name
      }
    }
  }
}
```

**Issue:** The Go version calls `node.Parent.Members()` directly, but `node.Parent` may not be the class declaration. The TypeScript version correctly accesses `node.parent.body` which is specifically the class body.

**Impact:** May miss setter detection when the getter's parent is not directly the class, leading to false positives.

**Test Coverage:** Test cases with getters that have corresponding setters may not be handled correctly.

#### 2. Abstract Method Detection
**TypeScript Implementation:**
```typescript
if (
  node.kind !== 'get' ||
  node.override ||
  !node.value.body ||
  node.value.body.body.length === 0
) {
  return;
}
```

**Go Implementation:**
```go
if getter.Body == nil {
  return
}

if !ast.IsBlock(getter.Body) {
  return
}
```

**Issue:** The Go version doesn't explicitly check for abstract getters (those without body implementation). The TypeScript version checks `!node.value.body` which catches abstract methods.

**Impact:** May incorrectly flag abstract getters for conversion to fields.

**Test Coverage:** The valid test case with `abstract get p1(): string;` may fail.

#### 3. Constructor Detection Pattern
**TypeScript Implementation:**
```typescript
'MethodDefinition[kind="constructor"] ThisExpression'(
  node: TSESTree.ThisExpression,
): void {
  // Specific selector for this expressions inside constructors
}
```

**Go Implementation:**
```go
listeners[ast.KindThisKeyword] = func(node *ast.Node) {
  // Broader listener that checks all this expressions
  // Then walks up to find if we're in a constructor
}
```

**Issue:** The Go version uses a broader approach that may catch `this` expressions outside of constructors, then tries to filter. The TypeScript version uses a specific selector that only matches `this` expressions directly inside constructor method definitions.

**Impact:** May have different behavior for nested functions or complex constructor patterns.

**Test Coverage:** Test cases with nested classes or functions inside constructors may behave differently.

#### 4. Computed Property Name Handling
**TypeScript Implementation:**
```typescript
// Uses getStaticMemberAccessValue utility from @typescript-eslint/utils
// which has sophisticated computed property handling
```

**Go Implementation:**
```go
func extractPropertyName(ctx rule.RuleContext, nameNode *ast.Node) string {
  if nameNode.Kind == ast.KindComputedPropertyName {
    computed := nameNode.AsComputedPropertyName()
    return extractPropertyNameFromExpression(ctx, computed.Expression)
  }
  // ...
}
```

**Issue:** The Go version's computed property handling may not match the TypeScript-ESLint utility's behavior exactly, particularly for complex expressions.

**Impact:** Properties with computed names may not be properly identified or excluded.

**Test Coverage:** Test cases with computed property names like `[myValue]` may behave differently.

#### 5. Template Literal Support
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TaggedTemplateExpression:
  return node.quasi.quasis.length === 1;

case AST_NODE_TYPES.TemplateLiteral:
  return node.quasis.length === 1;
```

**Go Implementation:**
```go
case ast.KindTemplateExpression:
  template := node.AsTemplateExpression()
  return template != nil && len(template.TemplateSpans.Nodes) == 0
case ast.KindTaggedTemplateExpression:
  // Support tagged template expressions only with no interpolation
  tagged := node.AsTaggedTemplateExpression()
  if tagged.Template.Kind == ast.KindNoSubstitutionTemplateLiteral {
    return true
  }
  // ...
```

**Issue:** Different AST node kinds and property names for template literal checking. The logic may not be equivalent.

**Impact:** Template literals with or without interpolation may be handled differently.

**Test Coverage:** Test cases with template literals and tagged template expressions may have different behavior.

#### 6. Class Body Event Handling
**TypeScript Implementation:**
```typescript
ClassBody: enterClassBody,
'ClassBody:exit': exitClassBody,
```

**Go Implementation:**
```go
listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
  enterClassBody()
}
listeners[rule.ListenerOnExit(ast.KindClassDeclaration)] = func(node *ast.Node) {
  exitClassBody()
}
```

**Issue:** The TypeScript version specifically listens to ClassBody enter/exit events, while the Go version listens to class declaration events. These may not be equivalent in all cases.

**Impact:** Stack management for properties may not work correctly with nested classes or class expressions.

**Test Coverage:** Nested class scenarios may behave differently.

### Recommendations
- **Fix setter detection**: Access the class body correctly by walking up to find the class declaration/expression, then access its members
- **Add abstract method check**: Verify that getters have actual bodies before suggesting conversion to fields
- **Improve constructor detection**: Make the `this` expression filtering more precise to match TypeScript-ESLint's selector behavior
- **Enhance computed property handling**: Ensure the property name extraction matches TypeScript-ESLint's utilities exactly
- **Verify template literal logic**: Double-check the AST node type mappings and property access for template literals
- **Fix class body event handling**: Consider listening to class body-specific events if available, or ensure proper class member access
- **Add comprehensive test coverage**: Test all edge cases mentioned, particularly nested classes, abstract methods, and computed properties

---