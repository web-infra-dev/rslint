## Rule: max-params

### Test File: max-params.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core parameter counting logic implementation
  - Void `this` parameter detection structure and logic
  - Configuration options parsing (max, maximum, countVoidThis)
  - Default max value of 3 parameters
  - Support for function declarations, expressions, arrow functions, methods, constructors, and accessors
  - Error message ID structure ("exceed")
  - Option precedence handling (maximum as deprecated alias for max)

- ⚠️ **POTENTIAL ISSUES**: 
  - Missing support for TypeScript-specific node types (TSFunctionType, TSDeclareFunction)
  - Different AST access patterns may affect void this parameter detection accuracy
  - Message format may not match ESLint base rule exactly

- ❌ **INCORRECT**: 
  - Missing AST node type listeners for `TSDeclareFunction` and `TSFunctionType`
  - Parameter extraction logic incomplete for TypeScript-specific constructs

### Discrepancies Found

#### 1. Missing TypeScript-Specific AST Node Support
**TypeScript Implementation:**
```typescript
type FunctionLike =
  | TSESTree.ArrowFunctionExpression
  | TSESTree.FunctionDeclaration
  | TSESTree.FunctionExpression
  | TSESTree.TSDeclareFunction
  | TSESTree.TSFunctionType;

return {
  ArrowFunctionExpression: wrapListener(baseRules.ArrowFunctionExpression),
  FunctionDeclaration: wrapListener(baseRules.FunctionDeclaration),
  FunctionExpression: wrapListener(baseRules.FunctionExpression),
  TSDeclareFunction: wrapListener(baseRules.FunctionDeclaration),
  TSFunctionType: wrapListener(baseRules.FunctionDeclaration),
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
  ast.KindFunctionDeclaration: checkFunction,
  ast.KindFunctionExpression:  checkFunction,
  ast.KindArrowFunction:       checkFunction,
  ast.KindMethodDeclaration:   checkFunction,
  ast.KindConstructor:         checkFunction,
  ast.KindGetAccessor:         checkFunction,
  ast.KindSetAccessor:         checkFunction,
}
```

**Issue:** The Go implementation is missing explicit support for `TSDeclareFunction` and `TSFunctionType` node types that are handled in the TypeScript version.

**Impact:** Critical test cases will fail:
- `declare function makeDate(m: number, d: number, y: number): Date;` (TSDeclareFunction)
- `type sum = (a: number, b: number) => number;` (TSFunctionType)

**Test Coverage:** Two explicit test cases fail without these node types:
1. Valid case with `declare function makeDate` (3 parameters, max: 3)
2. Invalid cases with `type sum` and `declare function makeDate` with restricted limits

#### 2. Parameter Extraction for Missing Node Types
**TypeScript Implementation:**
```typescript
// Base rule handles parameter extraction for all supported node types
// including TSDeclareFunction and TSFunctionType
```

**Go Implementation:**
```go
func getParameters(node *ast.Node) []*ast.Node {
  switch node.Kind {
  case ast.KindFunctionDeclaration:
    return node.AsFunctionDeclaration().Parameters.Nodes
  case ast.KindFunctionExpression:
    return node.AsFunctionExpression().Parameters.Nodes
  case ast.KindArrowFunction:
    return node.AsArrowFunction().Parameters.Nodes
  // ... missing cases for declare functions and function types
  default:
    return nil
  }
}
```

**Issue:** The `getParameters` function doesn't handle declare functions or function type aliases, so even if listeners were added, parameter extraction would fail.

**Impact:** Parameters cannot be counted for TypeScript-specific function constructs.

**Test Coverage:** All test cases involving `declare function` and `type` function signatures.

#### 3. Void This Parameter Type Annotation Structure
**TypeScript Implementation:**
```typescript
node.params[0].typeAnnotation?.typeAnnotation.type !== AST_NODE_TYPES.TSVoidKeyword
```

**Go Implementation:**
```go
// Check if it has a void type annotation
if paramNode.Type == nil {
  return false
}

return paramNode.Type.Kind == ast.KindVoidKeyword
```

**Issue:** The TypeScript version checks nested structure (`typeAnnotation.typeAnnotation.type`) while the Go version checks direct type (`paramNode.Type.Kind`). This suggests different AST structures between the two implementations.

**Impact:** The void `this` parameter detection may not work correctly, affecting the `countVoidThis: false` behavior.

**Test Coverage:** Multiple test cases depend on this:
- `method(this: void, a, b, c) {}` should count as 3 parameters when `countVoidThis: false`
- `method(this: void, a) {}` with `countVoidThis: true, max: 1` should fail

#### 4. Message Format and Structure
**TypeScript Implementation:**
```typescript
messages: baseRule.meta.messages,
// Uses ESLint core rule messages
```

**Go Implementation:**
```go
func buildExceedMessage(name string, count int, max int) rule.RuleMessage {
  return rule.RuleMessage{
    Id:          "exceed",
    Description: fmt.Sprintf("%s has too many parameters (%d). Maximum allowed is %d.", name, count, max),
  }
}
```

**Issue:** The Go implementation creates custom messages instead of using the base ESLint rule messages. The format and exact wording may differ.

**Impact:** Error message consistency with ESLint may be compromised, affecting user experience during migration.

**Test Coverage:** All invalid test cases expect `messageId: 'exceed'` which appears to match.

#### 5. Function Name Extraction Logic
**TypeScript Implementation:**
```typescript
// Base rule handles function name extraction for error messages
```

**Go Implementation:**
```go
func getFunctionName(node *ast.Node) string {
  switch node.Kind {
  case ast.KindFunctionDeclaration:
    // ... detailed name extraction logic for each type
  // ... extensive switch statement
  }
}
```

**Issue:** The Go implementation includes custom name extraction logic that may not match the base ESLint rule's approach. Additionally, it doesn't handle the missing node types.

**Impact:** Function names in error messages may differ from ESLint, and missing node types won't have appropriate names.

**Test Coverage:** Error message format in test failures would reveal naming inconsistencies.

### Recommendations

1. **Add TypeScript-specific AST node support:**
   ```go
   // Research correct AST kinds for:
   // - TSDeclareFunction (likely ast.KindDeclareFunction or similar)
   // - TSFunctionType (likely ast.KindFunctionType or ast.KindTypeReference)
   ```

2. **Extend parameter extraction logic:**
   ```go
   func getParameters(node *ast.Node) []*ast.Node {
     switch node.Kind {
     // ... existing cases
     case ast.KindDeclareFunction: // or correct kind
       return node.AsDeclareFunction().Parameters.Nodes
     case ast.KindFunctionType: // or correct kind  
       return node.AsFunctionType().Parameters.Nodes
     }
   }
   ```

3. **Verify void this parameter detection:**
   - Test the AST structure for `this: void` parameters in Go AST
   - Ensure the type annotation checking matches TypeScript's nested structure
   - Consider using TypeChecker for more accurate type detection

4. **Validate message format consistency:**
   - Compare actual error messages with ESLint base rule output
   - Ensure message ID "exceed" is correct
   - Verify function name format matches ESLint expectations

5. **Run comprehensive tests:**
   ```bash
   # Test the specific failing cases:
   # 1. declare function makeDate(m: number, d: number, y: number): Date;
   # 2. type sum = (a: number, b: number) => number;
   # 3. method(this: void, a, b, c) {} variations
   ```

---