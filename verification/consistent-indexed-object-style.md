## Rule: consistent-indexed-object-style

### Test File: consistent-indexed-object-style.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic record/index-signature mode switching, interface to type conversion, readonly modifier handling, generic parameter preservation, type literal conversion, basic circular reference detection
- ⚠️ **POTENTIAL ISSUES**: Mapped type handling not implemented, suggestion system missing, complex circular reference detection may be incomplete
- ❌ **INCORRECT**: Missing TSMappedType handler, missing suggestion support for complex Record types, missing comprehensive scope-based circular reference detection

### Discrepancies Found

#### 1. Missing TSMappedType Handler
**TypeScript Implementation:**
```typescript
TSMappedType(node): void {
  const key = node.key;
  const scope = context.sourceCode.getScope(key);
  
  const scopeManagerKey = nullThrows(
    scope.variables.find(
      value => value.name === key.name && value.isTypeVariable,
    ),
    'key type parameter must be a defined type variable in its scope',
  );

  // If the key is used to compute the value, we can't convert to a Record.
  if (
    scopeManagerKey.references.some(
      reference => reference.isTypeReference,
    )
  ) {
    return;
  }
  // ... rest of mapped type logic
}
```

**Go Implementation:**
```go
// Missing entirely - no handler for mapped types like { [K in keyof T]: V }
```

**Issue:** The Go implementation doesn't handle mapped types at all. The TypeScript version has comprehensive logic for converting mapped types to Record when appropriate.

**Impact:** Test cases with mapped types like `{ [key in string]: number }` won't be processed correctly.

**Test Coverage:** Multiple test cases use mapped types, including ones with readonly, optional, and required modifiers.

#### 2. Missing Suggestion System
**TypeScript Implementation:**
```typescript
context.report({
  node,
  messageId: 'preferIndexSignature',
  ...getFixOrSuggest({
    fixOrSuggest: shouldFix ? 'fix' : 'suggest',
    suggestion: {
      messageId: 'preferIndexSignatureSuggestion',
      fix: fixer => {
        const key = context.sourceCode.getText(params[0]);
        const type = context.sourceCode.getText(params[1]);
        return fixer.replaceText(node, `{ [key: ${key}]: ${type} }`);
      },
    },
  }),
});
```

**Go Implementation:**
```go
if shouldFix {
  ctx.ReportNodeWithFixes(node, rule.RuleMessage{
    Id:          "preferIndexSignature",
    Description: "An index signature is preferred over a record.",
  }, rule.RuleFix{...})
} else {
  // For complex key types, just report without fix/suggestion 
  // (test framework doesn't support suggestions)
  ctx.ReportNode(node, rule.RuleMessage{...})
}
```

**Issue:** The Go implementation doesn't provide suggestions for complex Record types, only reports errors without fixes.

**Impact:** Test cases expecting suggestions for complex types like `Record<string | number, any>` will fail.

**Test Coverage:** Tests with `suggestions` array in error objects will not pass.

#### 3. Incomplete Circular Reference Detection
**TypeScript Implementation:**
```typescript
function isDeeplyReferencingType(
  node: TSESTree.Node,
  superVar: ScopeVariable,
  visited: Set<TSESTree.Node>,
): boolean {
  if (visited.has(node)) {
    // something on the chain is circular but it's not the reference being checked
    return false;
  }

  visited.add(node);
  // ... comprehensive AST traversal with scope analysis
}
```

**Go Implementation:**
```go
func isDeeplyReferencingType(node *ast.Node, superTypeName string, visited map[*ast.Node]bool) bool {
  // ... simpler traversal without scope analysis
  // Only checks type names, not variable scope references
}
```

**Issue:** The Go implementation uses simple string-based type name matching instead of proper scope-based variable reference analysis.

**Impact:** May miss some circular references or incorrectly identify others, especially in complex nested scenarios.

**Test Coverage:** Some circular reference test cases may not behave identically.

#### 4. Missing Conditional Type Handling
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TSConditionalType:
  return [
    node.checkType,
    node.extendsType,
    node.falseType,
    node.trueType,
  ].some(type => isDeeplyReferencingType(type, superVar, visited));
```

**Go Implementation:**
```go
case ast.KindConditionalType:
  // For conditional types like "Foo extends T ? string : number"
  // Check only the true and false types, not the check or extends types
  conditionalType := node.AsConditionalTypeNode()
  if conditionalType.TrueType != nil && containsTypeReference(conditionalType.TrueType, typeName) {
    return true
  }
  if conditionalType.FalseType != nil && containsTypeReference(conditionalType.FalseType, typeName) {
    return true
  }
  // Note: We don't check CheckType or ExtendsType as these are type predicates, not circular refs
```

**Issue:** Different handling of conditional types - the Go version explicitly excludes checkType and extendsType while the TypeScript version includes all parts.

**Impact:** May lead to different circular reference detection behavior for conditional types.

**Test Coverage:** Tests with conditional types in circular scenarios may behave differently.

#### 5. Missing Index Signature Parameter Validation
**TypeScript Implementation:**
```typescript
const parameter = member.parameters.at(0);
if (parameter?.type !== AST_NODE_TYPES.Identifier) {
  return;
}

const keyType = parameter.typeAnnotation;
if (!keyType) {
  return;
}
```

**Go Implementation:**
```go
param := indexSig.Parameters.Nodes[0]
if param.Kind != ast.KindParameter {
  return
}

paramDecl := param.AsParameterDeclaration()
keyType := paramDecl.Type
if keyType == nil {
  return
}
```

**Issue:** The Go version doesn't validate that the parameter is an Identifier before proceeding, only checks if it's a Parameter.

**Impact:** May process invalid index signatures that should be ignored.

**Test Coverage:** Edge cases with malformed index signatures may behave differently.

#### 6. Missing Parent Declaration Context Handling
**TypeScript Implementation:**
```typescript
function findParentDeclaration(
  node: TSESTree.Node,
): TSESTree.TSTypeAliasDeclaration | undefined {
  if (node.parent && node.parent.type !== AST_NODE_TYPES.TSTypeAnnotation) {
    if (node.parent.type === AST_NODE_TYPES.TSTypeAliasDeclaration) {
      return node.parent;
    }
    return findParentDeclaration(node.parent);
  }
  return undefined;
}
```

**Go Implementation:**
```go
func findParentDeclaration(node *ast.Node) *ast.Node {
  parent := node.Parent
  for parent != nil {
    if parent.Kind == ast.KindTypeAliasDeclaration {
      return parent
    }
    parent = parent.Parent
  }
  return nil
}
```

**Issue:** The Go version doesn't handle the TSTypeAnnotation exclusion logic that prevents finding parent declarations across type annotation boundaries.

**Impact:** May incorrectly identify parent declarations in some contexts.

**Test Coverage:** Complex nested type scenarios may behave differently.

### Recommendations
- Implement TSMappedType handler with proper key usage detection and modifier handling (readonly, optional, required)
- Add suggestion system support for complex Record types that can't be auto-fixed
- Enhance circular reference detection with proper scope analysis instead of simple string matching
- Add comprehensive parameter validation for index signatures
- Implement proper parent declaration finding with TSTypeAnnotation boundary handling
- Add support for all mapped type modifiers (+readonly, -readonly, +?, -?, etc.)
- Ensure all AST node types are handled in circular reference detection
- Add proper scope-based variable reference tracking

---