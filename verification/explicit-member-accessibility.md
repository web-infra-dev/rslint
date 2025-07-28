# Rule: explicit-member-accessibility

## Test File: explicit-member-accessibility.test.ts

## Validation Summary
- ✅ **CORRECT**: Basic accessibility checking, configuration parsing, modifier detection, AST node handling for methods/properties/constructors, parameter property handling
- ⚠️ **POTENTIAL ISSUES**: Abstract member handling, accessor property detection, decorator positioning, fix suggestions, public keyword removal logic
- ❌ **INCORRECT**: Message ID consistency, fix/suggestion generation, parameter property detection logic, computed property names

## Discrepancies Found

### 1. Message ID Structure Mismatch
**TypeScript Implementation:**
```typescript
export type MessageIds =
  | 'addExplicitAccessibility'
  | 'missingAccessibility'
  | 'unwantedPublicAccessibility';

messages: {
  addExplicitAccessibility: "Add '{{ type }}' accessibility modifier",
  missingAccessibility: 'Missing accessibility modifier on {{type}} {{name}}.',
  unwantedPublicAccessibility: 'Public accessibility modifier on {{type}} {{name}}.',
},
```

**Go Implementation:**
```go
ctx.ReportNode(node, rule.RuleMessage{
    Id:          "missingAccessibility",
    Description: fmt.Sprintf("Missing accessibility modifier on %s %s.", nodeType, methodName),
})
```

**Issue:** The Go implementation uses hardcoded descriptions instead of message IDs with data placeholders. The TypeScript version uses template variables like `{{type}}` and `{{name}}`.

**Impact:** Error messages may not match exactly, and internationalization/customization is not supported.

**Test Coverage:** All test cases that check messageId values will fail in Go implementation.

### 2. Parameter Property Detection Logic
**TypeScript Implementation:**
```typescript
function checkParameterPropertyAccessibilityModifier(
  node: TSESTree.TSParameterProperty,
): void {
  const nodeType = 'parameter property';
  // HAS to be an identifier or assignment or TSC will throw
  if (
    node.parameter.type !== AST_NODE_TYPES.Identifier &&
    node.parameter.type !== AST_NODE_TYPES.AssignmentPattern
  ) {
    return;
  }
```

**Go Implementation:**
```go
checkParameterPropertyAccessibilityModifier := func(node *ast.Node) {
    if node.Kind != ast.KindParameter {
        return
    }
    param := node.AsParameterDeclaration()
    
    // Check if it's a parameter property (has modifiers)
    if param.Modifiers() == nil {
        return
    }

    // Check if it has readonly or accessibility modifiers
    hasReadonly := false
    hasAccessibility := false
    // ... modifier checking logic
    
    // A parameter property must have readonly OR accessibility modifier
    if !hasReadonly && !hasAccessibility {
        return
    }
```

**Issue:** The Go implementation checks for different AST node types and uses different logic to determine if a parameter is a parameter property. The TypeScript version specifically handles `TSParameterProperty` nodes, while Go checks regular parameters with modifiers.

**Impact:** Parameter properties may not be detected correctly in all cases, especially complex destructuring patterns.

**Test Coverage:** Tests with parameter properties like `constructor(readonly foo: string)` may not work correctly.

### 3. Fix and Suggestion Generation Missing
**TypeScript Implementation:**
```typescript
function findPublicKeyword(node): { range: TSESLint.AST.Range; rangeToRemove: TSESLint.AST.Range } {
  const tokens = context.sourceCode.getTokens(node);
  let rangeToRemove!: TSESLint.AST.Range;
  let keywordRange!: TSESLint.AST.Range;
  for (let i = 0; i < tokens.length; i++) {
    const token = tokens[i];
    if (token.type === AST_TOKEN_TYPES.Keyword && token.value === 'public') {
      keywordRange = structuredClone(token.range);
      const commensAfterPublicKeyword = context.sourceCode.getCommentsAfter(token);
      if (commensAfterPublicKeyword.length) {
        rangeToRemove = [token.range[0], commensAfterPublicKeyword[0].range[0]];
        break;
      } else {
        rangeToRemove = [token.range[0], tokens[i + 1].range[0]];
        break;
      }
    }
  }
  return { range: keywordRange, rangeToRemove };
}
```

**Go Implementation:**
```go
func findPublicKeywordRange(ctx rule.RuleContext, node *ast.Node) (core.TextRange, core.TextRange) {
    // ... get modifiers from different node types
    for i, mod := range modifiers.NodeList.Nodes {
        if mod.Kind == ast.KindPublicKeyword {
            keywordRange := core.NewTextRange(mod.Pos(), mod.End())
            
            // Calculate range to remove (including following whitespace)
            removeEnd := mod.End()
            if i+1 < len(modifiers.NodeList.Nodes) {
                removeEnd = modifiers.NodeList.Nodes[i+1].Pos()
            } else {
                // Find next token after public keyword
                text := string(ctx.SourceFile.Text())
                for removeEnd < len(text) && (text[removeEnd] == ' ' || text[removeEnd] == '\t') {
                    removeEnd++
                }
            }
            
            removeRange := core.NewTextRange(mod.Pos(), removeEnd)
            return keywordRange, removeRange
        }
    }
    return core.NewTextRange(0, 0), core.NewTextRange(0, 0)
}
```

**Issue:** The Go implementation has the keyword finding logic but doesn't use it in actual fix generation. The tests show `output` properties with fixed code, but the Go version doesn't provide these fixes.

**Impact:** Auto-fixes for removing unwanted public modifiers don't work, causing test output mismatches.

**Test Coverage:** All test cases with `output` properties showing removed `public` keywords.

#### 4. Decorator Handling
**TypeScript Implementation:**
```typescript
function getMissingAccessibilitySuggestions(node) {
  function fix(accessibility, fixer) {
    if (node.decorators.length) {
      const lastDecorator = node.decorators[node.decorators.length - 1];
      const nextToken = nullThrows(context.sourceCode.getTokenAfter(lastDecorator));
      return fixer.insertTextBefore(nextToken, `${accessibility} `);
    }
    return fixer.insertTextBefore(node, `${accessibility} `);
  }
}
```

**Go Implementation:**
```go
func hasDecorators(node *ast.Node) bool {
    // Check if node has decorator modifiers
    return ast.GetCombinedModifierFlags(node)&ast.ModifierFlagsDecorator != 0
}
```

**Issue:** The Go implementation has a basic decorator detection function but doesn't use it properly in fix positioning. The TypeScript version carefully positions fixes after decorators.

**Impact:** Incorrect positioning of accessibility modifiers when decorators are present.

**Test Coverage:** Test cases with `@foo @bar()` decorators show this issue.

#### 5. Abstract Member Handling
**TypeScript Implementation:**
```typescript
// Handles both TSAbstractMethodDefinition and TSAbstractPropertyDefinition
'MethodDefinition, TSAbstractMethodDefinition': checkMethodAccessibilityModifier,
'PropertyDefinition, TSAbstractPropertyDefinition, AccessorProperty, TSAbstractAccessorProperty': checkPropertyAccessibilityModifier,
```

**Go Implementation:**
```go
return rule.RuleListeners{
    ast.KindMethodDeclaration: checkMethodAccessibilityModifier,
    ast.KindConstructor: checkMethodAccessibilityModifier,
    ast.KindGetAccessor: checkMethodAccessibilityModifier,
    ast.KindSetAccessor: checkMethodAccessibilityModifier,
    ast.KindPropertyDeclaration: checkPropertyAccessibilityModifier,
    ast.KindParameter: checkParameterPropertyAccessibilityModifier,
}
```

**Issue:** The Go implementation doesn't explicitly handle abstract method and property declarations as separate cases, relying on the `isAbstract()` helper function instead.

**Impact:** May miss some abstract member cases or handle them incorrectly compared to the TypeScript version.

**Test Coverage:** Test cases with `abstract class` and `abstract method()` patterns.

#### 6. Node Type Determination
**TypeScript Implementation:**
```typescript
let nodeType = 'method definition';
let check = baseCheck;
switch (methodDefinition.kind) {
  case 'method': check = methodCheck; break;
  case 'constructor': check = ctorCheck; break;
  case 'get':
  case 'set':
    check = accessorCheck;
    nodeType = `${methodDefinition.kind} property accessor`;
    break;
}
```

**Go Implementation:**
```go
func getNodeType(node *ast.Node, memberKind string) string {
    switch memberKind {
    case "constructor":
        return "constructor"
    case "get", "set":
        return fmt.Sprintf("%s property accessor", memberKind)
    default:
        if node.Kind == ast.KindPropertyDeclaration {
            return "class property"
        }
        return "method definition"
    }
}
```

**Issue:** The logic is similar but the Go version has a different flow and may not handle all edge cases the same way.

**Impact:** Error messages might have slightly different node type descriptions.

**Test Coverage:** Test cases checking specific error message content with node types.

### Recommendations
- Implement proper fix/suggestion integration with the reporting mechanism
- Add auto-fix functionality for removing unwanted public modifiers  
- Improve parameter property detection to match TypeScript AST behavior exactly
- Enhance decorator handling in fix positioning
- Add comprehensive abstract member support
- Verify error message consistency with original implementation
- Add missing test coverage for edge cases like computed property names and complex decorator scenarios

---