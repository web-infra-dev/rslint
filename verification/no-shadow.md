## Rule: no-shadow

### Test File: no-shadow.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic shadowing detection, scope management, hoisting behavior, type/value shadow ignoring, function type parameter handling, builtin globals support, class/enum duplicate name detection
- ⚠️ **POTENTIAL ISSUES**: Complex AST navigation patterns, initialization pattern detection, temporal dead zone handling, external declaration merging
- ❌ **INCORRECT**: Missing critical helper functions, incomplete scope analysis patterns, missing import type handling, static method generic shadowing logic

### Discrepancies Found

#### 1. Missing `isOnInitializer` Function
**TypeScript Implementation:**
```typescript
function isOnInitializer(
  variable: TSESLint.Scope.Variable,
  scopeVar: TSESLint.Scope.Variable,
): boolean {
  const outerScope = scopeVar.scope;
  const outerDef = scopeVar.defs.at(0);
  const outer = outerDef?.parent?.range;
  const innerScope = variable.scope;
  const innerDef = variable.defs.at(0);
  const inner = innerDef?.name.range;

  return !!(
    outer &&
    inner &&
    outer[0] < inner[0] &&
    inner[1] < outer[1] &&
    ((innerDef.type === DefinitionType.FunctionName &&
      innerDef.node.type === AST_NODE_TYPES.FunctionExpression) ||
      innerDef.node.type === AST_NODE_TYPES.ClassExpression) &&
    outerScope === innerScope.upper
  );
}
```

**Go Implementation:**
```go
// Function is completely missing
```

**Issue:** The Go implementation lacks the critical `isOnInitializer` function that prevents reporting shadowing in cases like `var a = function a() {};`

**Impact:** Will incorrectly report shadowing for legitimate function/class expressions that reference their own names.

**Test Coverage:** The valid test case `'var a = function a() { return a; };'` would fail.

#### 2. Missing `isInitPatternNode` Function
**TypeScript Implementation:**  
```typescript
function isInitPatternNode(
  variable: TSESLint.Scope.Variable,
  shadowedVariable: TSESLint.Scope.Variable,
): boolean {
  // Complex logic to detect initialization patterns
  // involving arrow functions, call expressions, and variable declarators
}
```

**Go Implementation:**
```go
// Function is completely missing
```

**Issue:** Missing complex initialization pattern detection that works with the `ignoreOnInitialization` option.

**Impact:** The `ignoreOnInitialization` option won't work correctly, potentially missing valid cases where shadowing should be ignored.

**Test Coverage:** Any test cases using `ignoreOnInitialization: true` would not behave correctly.

#### 3. Missing `isInTdz` (Temporal Dead Zone) Function
**TypeScript Implementation:**
```typescript
function isInTdz(
  variable: TSESLint.Scope.Variable,
  scopeVar: TSESLint.Scope.Variable,
): boolean {
  const outerDef = scopeVar.defs.at(0);
  const inner = getNameRange(variable);
  const outer = getNameRange(scopeVar);

  if (!inner || !outer || inner[1] >= outer[0]) {
    return false;
  }
  // Additional TDZ logic based on hoist options
}
```

**Go Implementation:**
```go
// Partial implementation in checkVariable, but not as a separate function
// TDZ logic is simplified and may not handle all cases correctly
```

**Issue:** The TDZ (Temporal Dead Zone) detection is incomplete and not properly separated into its own function.

**Impact:** Hoisting behavior may not work correctly for all variable types and positions.

**Test Coverage:** Complex hoisting test cases may fail.

#### 4. Missing Import Type Handling
**TypeScript Implementation:**
```typescript
import { isTypeImport } from '../util/isTypeImport';

function isExternalDeclarationMerging(
  scope: TSESLint.Scope.Scope,
  variable: TSESLint.Scope.Variable,
  shadowed: TSESLint.Scope.Variable,
): boolean {
  const [firstDefinition] = shadowed.defs;
  return (
    isTypeImport(firstDefinition) &&
    isImportDeclaration(firstDefinition.parent) &&
    // Additional external declaration merging logic
  );
}
```

**Go Implementation:**
```go
// No import type detection or external declaration merging logic
```

**Issue:** Missing import type detection and external declaration merging logic.

**Impact:** Will incorrectly report shadowing for legitimate type imports and module declaration merging.

**Test Coverage:** Test case `'import type { Foo } from "./foo"; function bar(Foo: string) {}'` may not work correctly.

#### 5. Incomplete Static Method Generic Shadowing
**TypeScript Implementation:**
```typescript
function isGenericOfStaticMethod(variable: TSESLint.Scope.Variable): boolean {
  // Complex AST traversal to detect static method generics
  const typeParameter = variable.identifiers[0].parent;
  const typeParameterDecl = typeParameter.parent;
  const functionExpr = typeParameterDecl.parent;
  const methodDefinition = functionExpr.parent;
  return methodDefinition.type === AST_NODE_TYPES.MethodDefinition &&
         methodDefinition.static;
}

function isGenericOfAStaticMethodShadow(variable, shadowed): boolean {
  return isGenericOfStaticMethod(variable) && isGenericOfClass(shadowed);
}
```

**Go Implementation:**
```go
// Logic is completely missing
```

**Issue:** Missing complex logic to handle static method generic type parameter shadowing class generics.

**Impact:** Will incorrectly report shadowing for legitimate static method generics that shadow class generics.

**Test Coverage:** Test case `'class Foo<T> { static method<T>() {} }'` may fail.

#### 6. Incomplete Function Type Parameter Detection
**TypeScript Implementation:**
```typescript
const allowedFunctionVariableDefTypes = new Set([
  AST_NODE_TYPES.TSCallSignatureDeclaration,
  AST_NODE_TYPES.TSFunctionType,
  AST_NODE_TYPES.TSMethodSignature,
  AST_NODE_TYPES.TSEmptyBodyFunctionExpression,
  AST_NODE_TYPES.TSDeclareFunction,
  AST_NODE_TYPES.TSConstructSignatureDeclaration,
  AST_NODE_TYPES.TSConstructorType,
]);

function isFunctionTypeParameterNameValueShadow(variable, shadowed): boolean {
  return variable.defs.every(def =>
    allowedFunctionVariableDefTypes.has(def.node.type),
  );
}
```

**Go Implementation:**
```go
isFunctionTypeParameterNameValueShadow := func(variable *Variable, shadowed *Variable) bool {
  // Simplified logic that doesn't properly check parent context
  return allowedFunctionVariableDefTypes[parent.Kind] && parent.Kind != ast.KindFunctionDeclaration
}
```

**Issue:** The Go implementation doesn't properly traverse the AST to determine if a parameter is truly in a function type context vs. a regular function parameter.

**Impact:** May not correctly distinguish between function type parameters and regular function parameters.

**Test Coverage:** Function type parameter test cases may not work correctly.

#### 7. Missing Definition File Detection
**TypeScript Implementation:**
```typescript
function isDeclareInDTSFile(variable: TSESLint.Scope.Variable): boolean {
  const fileName = context.filename;
  if (!isDefinitionFile(fileName)) {
    return false;
  }
  return variable.defs.some(def => {
    return (
      (def.type === DefinitionType.Variable && def.parent.declare) ||
      (def.type === DefinitionType.ClassName && def.node.declare) ||
      // Additional declare modifier checks
    );
  });
}
```

**Go Implementation:**
```go
// Logic is completely missing
```

**Issue:** Missing detection of declare modifiers in .d.ts files.

**Impact:** Will incorrectly report shadowing for legitimate declare statements in TypeScript definition files.

**Test Coverage:** Any .d.ts file test cases would be affected.

#### 8. Simplified Same-Scope Redeclaration Logic
**TypeScript Implementation:**
```typescript
// Integrated into the main checkForShadows function with complex scope traversal
function checkForShadows(scope: TSESLint.Scope.Scope): void {
  const variables = scope.variables;
  for (const variable of variables) {
    // Complex logic considering all definition types and scopes
  }
}
```

**Go Implementation:**
```go
// Simplified check in addVariable and checkVariable
if v, exists := scope.Variables[name]; exists {
  if v.IsValue && isValue {
    // Simple same-scope check
  }
}
```

**Issue:** The same-scope redeclaration detection is overly simplified and may not handle all TypeScript declaration merging cases correctly.

**Impact:** May report false positives for legitimate TypeScript declaration merging (interfaces, namespaces, etc.).

**Test Coverage:** Complex declaration merging scenarios may fail.

### Recommendations
- Implement missing `isOnInitializer` function to handle function/class expression self-references
- Add `isInitPatternNode` function to support `ignoreOnInitialization` option fully
- Implement proper `isInTdz` function with complete temporal dead zone logic
- Add import type detection and external declaration merging support
- Implement static method generic shadowing detection logic
- Improve function type parameter detection with proper AST traversal
- Add declare modifier detection for .d.ts files
- Enhance same-scope redeclaration logic to handle TypeScript declaration merging
- Add comprehensive AST node type mappings between TypeScript and Go AST types
- Implement proper scope analysis patterns that match TypeScript-ESLint's scope manager behavior

---