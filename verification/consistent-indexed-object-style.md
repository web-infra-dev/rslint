# Rule Validation: consistent-indexed-object-style

## Test File: consistent-indexed-object-style.test.ts

## Validation Summary
- ✅ **CORRECT**: Basic rule structure, option handling, core listeners, mapped type handling, readonly modifier support
- ⚠️ **POTENTIAL ISSUES**: Circular reference detection logic, qualified name handling, nested type literal processing
- ❌ **INCORRECT**: Critical circular reference prevention logic is overly permissive

## Discrepancies Found

### 1. **CRITICAL: Circular Reference Detection Logic**

**TypeScript Implementation:**
```typescript
function isDeeplyReferencingType(
  node: TSESTree.Node,
  superVar: ScopeVariable,
  visited: Set<TSESTree.Node>,
): boolean {
  // Complex recursive traversal with scope-aware variable checking
  // Uses TypeScript-ESLint's scope manager to track variable references
  if (superVar.references.some(ref => isNodeEqual(ref.identifier, node))) {
    return true;
  }
}
```

**Go Implementation:**
```go
func hasProblematicSelfReference(node *ast.Node, typeName string) bool {
  // Always returns false - completely disabled
  return false  // Never block conversions due to self-references
}

func hasCircularReference(sourceFile *ast.SourceFile, typeName string) bool {
  // Always returns false - completely disabled  
  return false  // Never block conversions due to circular references
}
```

**Issue:** The Go implementation has completely disabled circular reference detection, making it overly permissive compared to TypeScript-ESLint.

**Impact:** The rule will convert types that should remain as index signatures due to problematic circular references, potentially breaking TypeScript compilation.

**Test Coverage:** Several valid test cases rely on circular reference detection:
- `'type Foo = { [key: string]: Foo };'`
- Complex circular chains with interfaces
- Indirect circular references

### 2. **Scope-Aware Variable Tracking Missing**

**TypeScript Implementation:**
```typescript
const scope = context.sourceCode.getScope(parentId);
const superVar = ASTUtils.findVariable(scope, parentId.name);
if (superVar && isDeeplyReferencingType(node, superVar, new Set([parentId]))) {
  return;
}
```

**Go Implementation:**
```go
// No scope-aware variable tracking implemented
// Uses simple string-based type name matching instead
func containsSelfReference(node *ast.Node, typeName string) bool {
  // Basic string matching without scope awareness
}
```

**Issue:** The Go implementation lacks TypeScript-ESLint's sophisticated scope-aware variable tracking, using only basic string matching.

**Impact:** May incorrectly identify or miss circular references in complex scoping scenarios.

### 3. **SCC-Based Circular Detection Incorrect**

**TypeScript Implementation:**
```typescript
// TypeScript-ESLint uses deep recursive traversal with visited tracking
// Checks actual variable references and definitions
```

**Go Implementation:**
```go
func getCircularChainTypes(sourceFile *ast.SourceFile, typeName string) []string {
  // Uses Strongly Connected Components (SCC) algorithm
  // But then blocks ALL types in circular chains with len(circularChain) > 1
  if len(circularChain) > 1 {
    return // Don't convert any types in circular chains
  }
}
```

**Issue:** The SCC approach is more aggressive than TypeScript-ESLint and blocks conversions that should be allowed.

**Impact:** Will incorrectly prevent valid conversions in complex type dependency graphs.

### 4. **Qualified Name Handling**

**TypeScript Implementation:**
```typescript
if (typeName.type !== AST_NODE_TYPES.Identifier) {
  return; // Skip qualified names like A.B
}
```

**Go Implementation:**
```go
} else if typeRef.TypeName != nil && ast.IsQualifiedName(typeRef.TypeName) {
  // For qualified names like A.Foo, we don't treat them as references to Foo
  // since they are in a different namespace. These should NOT prevent conversion.
}
```

**Issue:** The Go implementation correctly identifies qualified names but the comment suggests they should NOT prevent conversion, while TypeScript-ESLint simply skips them.

**Impact:** Potentially allows conversions that TypeScript-ESLint would prevent.

### 5. **Missing Index Signature Validation**

**TypeScript Implementation:**
```typescript
const parameter = member.parameters.at(0);
if (parameter?.type !== AST_NODE_TYPES.Identifier) {
  return;
}
```

**Go Implementation:**
```go
parameter := indexSig.Parameters.Nodes[0]
if parameter.Kind != ast.KindParameter {
  return;
}
// Missing validation that parameter name is an Identifier
```

**Issue:** The Go implementation doesn't validate that the index signature parameter is an Identifier node.

**Impact:** May attempt to process malformed index signatures that TypeScript-ESLint would skip.

### 6. **Conditional Type Detection Logic**

**TypeScript Implementation:**
```typescript
// Specific handling in isDeeplyReferencingType for conditional types
case AST_NODE_TYPES.TSConditionalType:
  return [node.checkType, node.extendsType, node.falseType, node.trueType]
    .some(type => isDeeplyReferencingType(type, superVar, visited));
```

**Go Implementation:**
```go
func containsConditionalType(node *ast.Node) bool {
  // Correctly identifies conditional types but separate from circular reference logic
  // Used to skip conversion, but not integrated with self-reference detection
}
```

**Issue:** Conditional type detection is separated from circular reference logic, unlike TypeScript-ESLint.

**Impact:** May not properly handle edge cases involving conditional types in circular references.

### 7. **Mapped Type Key Usage Detection**

**TypeScript Implementation:**
```typescript
// Uses scope manager to check if key is used in value computation
if (scopeManagerKey.references.some(reference => reference.isTypeReference)) {
  return;
}
```

**Go Implementation:**
```go
if mappedType.Type != nil {
  var keyName string
  if name := typeParam.Name(); name != nil && name.Kind == ast.KindIdentifier {
    keyName = name.AsIdentifier().Text
  }
  valueText := getNodeText(ctx.SourceFile, mappedType.Type)
  if strings.Contains(valueText, keyName) {
    return; // Simple string-based check
  }
}
```

**Issue:** The Go implementation uses simple string containment rather than proper scope-aware reference checking.

**Impact:** May incorrectly detect key usage in string literals or comments, or miss usage in complex expressions.

## Recommendations

### Critical Fixes Needed:

1. **Implement Proper Circular Reference Detection:**
   - Port TypeScript-ESLint's `isDeeplyReferencingType` logic
   - Add scope-aware variable tracking
   - Remove the overly permissive approach that always returns false

2. **Fix SCC-Based Approach:**
   - Either properly implement TypeScript-ESLint's recursive approach OR
   - Adjust SCC logic to match TypeScript-ESLint's permissive behavior for valid circular references

3. **Add Missing Validations:**
   - Validate index signature parameters are Identifiers
   - Integrate conditional type detection with circular reference logic

4. **Improve Key Usage Detection:**
   - Replace string-based matching with proper AST traversal
   - Implement scope-aware reference checking for mapped type keys

### Test Cases Needing Review:

1. All circular reference test cases in the valid array
2. Complex interface dependency chains
3. Mapped types with key usage in value expressions
4. Qualified name references (A.Foo patterns)

### Priority Level: **HIGH**

The circular reference detection is a critical safety feature that prevents TypeScript compilation errors. The current Go implementation is dangerously permissive and will likely cause issues in real-world usage.

---