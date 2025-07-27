## Rule: no-shadow

### Test File: no-shadow.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic shadowing detection, scope tracking, option parsing, built-in globals handling, message formatting
- ⚠️ **POTENTIAL ISSUES**: Function type parameter detection, type/value shadowing logic, complex AST pattern matching, temporal dead zone handling
- ❌ **INCORRECT**: Missing several critical TypeScript-specific features, incomplete scope analysis, missing advanced edge case handling

### Discrepancies Found

#### 1. Missing Type Import Detection
**TypeScript Implementation:**
```typescript
import { isTypeImport } from '../util/isTypeImport';

function isTypeValueShadow(
  variable: TSESLint.Scope.Variable,
  shadowed: TSESLint.Scope.Variable,
): boolean {
  const firstDefinition = shadowed.defs.at(0);
  const isShadowedValue =
    !('isValueVariable' in shadowed) ||
    !firstDefinition ||
    (!isTypeImport(firstDefinition) && shadowed.isValueVariable);
  return variable.isValueVariable !== isShadowedValue;
}
```

**Go Implementation:**
```go
isTypeValueShadow := func(variable *Variable, shadowed *Variable) bool {
  if !opts.IgnoreTypeValueShadow {
    return false
  }
  return variable.IsValue != shadowed.IsValue
}
```

**Issue:** The Go version oversimplifies type/value shadow detection and doesn't handle type imports properly.

**Impact:** May incorrectly flag or miss shadowing cases involving TypeScript type imports.

**Test Coverage:** Test case `import type { Foo } from "./foo"; function bar(Foo: string) {}` will fail.

#### 2. Incomplete Function Type Parameter Detection
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

function isFunctionTypeParameterNameValueShadow(
  variable: TSESLint.Scope.Variable,
  shadowed: TSESLint.Scope.Variable,
): boolean {
  return variable.defs.every(def =>
    allowedFunctionVariableDefTypes.has(def.node.type),
  );
}
```

**Go Implementation:**
```go
var allowedFunctionVariableDefTypes = map[ast.Kind]bool{
  ast.KindCallSignature:       true,
  ast.KindFunctionType:        true,
  ast.KindMethodSignature:     true,
  ast.KindEmptyStatement:      true, // TSEmptyBodyFunctionExpression
  ast.KindFunctionDeclaration: true, // TSDeclareFunction
  ast.KindConstructSignature:  true,
  ast.KindConstructorType:     true,
}

isFunctionTypeParameterNameValueShadow := func(variable *Variable, shadowed *Variable) bool {
  // Simplified check that excludes function declarations
  return allowedFunctionVariableDefTypes[parent.Kind] && parent.Kind != ast.KindFunctionDeclaration
}
```

**Issue:** The Go version has incorrect AST kind mappings and doesn't properly check all definitions like the TypeScript version.

**Impact:** Will incorrectly handle function type parameter shadowing cases.

**Test Coverage:** Test cases involving function type parameters will fail.

#### 3. Missing Complex Edge Case Handlers
**TypeScript Implementation:**
```typescript
function isGenericOfStaticMethod(variable: TSESLint.Scope.Variable): boolean {
  // Complex traversal logic for static method generics
}

function isGenericOfClass(variable: TSESLint.Scope.Variable): boolean {
  // Complex traversal logic for class generics
}

function isGenericOfAStaticMethodShadow(
  variable: TSESLint.Scope.Variable,
  shadowed: TSESLint.Scope.Variable,
): boolean {
  return isGenericOfStaticMethod(variable) && isGenericOfClass(shadowed);
}
```

**Go Implementation:**
```go
// These functions are completely missing
```

**Issue:** The Go version lacks handling for static method generic shadowing of class generics.

**Impact:** Test case `class Foo<T> { static method<T>() {} }` will incorrectly report shadowing.

**Test Coverage:** Static method generic test case will fail.

#### 4. Missing External Declaration Merging
**TypeScript Implementation:**
```typescript
function isExternalDeclarationMerging(
  scope: TSESLint.Scope.Scope,
  variable: TSESLint.Scope.Variable,
  shadowed: TSESLint.Scope.Variable,
): boolean {
  const [firstDefinition] = shadowed.defs;
  const [secondDefinition] = variable.defs;

  return (
    isTypeImport(firstDefinition) &&
    isImportDeclaration(firstDefinition.parent) &&
    isExternalModuleDeclarationWithName(
      scope,
      firstDefinition.parent.source.value,
    ) &&
    (secondDefinition.node.type === AST_NODE_TYPES.TSInterfaceDeclaration ||
     secondDefinition.node.type === AST_NODE_TYPES.TSTypeAliasDeclaration)
  );
}
```

**Go Implementation:**
```go
// This function is completely missing
```

**Issue:** The Go version doesn't handle external module declaration merging.

**Impact:** May incorrectly flag valid TypeScript declaration merging patterns as shadowing.

**Test Coverage:** Module declaration merging test cases will fail.

#### 5. Incomplete Initialization Pattern Detection
**TypeScript Implementation:**
```typescript
function isInitPatternNode(
  variable: TSESLint.Scope.Variable,
  shadowedVariable: TSESLint.Scope.Variable,
): boolean {
  // Complex logic to detect initialization patterns
  // Handles various assignment patterns and for-loop contexts
}

function isOnInitializer(
  variable: TSESLint.Scope.Variable,
  scopeVar: TSESLint.Scope.Variable,
): boolean {
  // Logic to detect if variable is in initializer of another variable
}
```

**Go Implementation:**
```go
// These functions are completely missing
```

**Issue:** The Go version lacks the `ignoreOnInitialization` option implementation.

**Impact:** Won't properly handle cases where shadowing should be ignored during initialization.

**Test Coverage:** Initialization-related test cases will fail.

#### 6. Incomplete Temporal Dead Zone Handling
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

  if (options.hoist === 'functions') {
    return !functionsHoistedNodes.has(outerDef.node.type);
  }
  // Additional hoist option handling
}
```

**Go Implementation:**
```go
// TDZ logic is partially implemented but incomplete
if opts.Hoist == "never" {
  if ast.IsFunctionDeclaration(shadowed.DeclaredAt) {
    return
  }
  // Simplified TDZ check
}
```

**Issue:** The Go version has incomplete temporal dead zone checking and doesn't handle all hoist options properly.

**Impact:** Will incorrectly report or miss shadowing in various hoisting scenarios.

**Test Coverage:** Hoisting-related test cases will have inconsistent behavior.

#### 7. Missing Global Augmentation Detection
**TypeScript Implementation:**
```typescript
function isGlobalAugmentation(scope: TSESLint.Scope.Scope): boolean {
  return (
    (scope.type === ScopeType.tsModule && scope.block.kind === 'global') ||
    (!!scope.upper && isGlobalAugmentation(scope.upper))
  );
}
```

**Go Implementation:**
```go
isGlobalAugmentation = func(scope *Scope) bool {
  if scope.Type == ScopeTypeTSModule && ast.IsModuleDeclaration(scope.Node) {
    moduleDecl := scope.Node.AsModuleDeclaration()
    nameNode := moduleDecl.Name()
    if ast.IsStringLiteral(nameNode) && nameNode.AsStringLiteral().Text == "global" {
      return true
    }
  }
  return scope.Parent != nil && isGlobalAugmentation(scope.Parent)
}
```

**Issue:** The Go version checks for string literal "global" but the TypeScript version checks for `kind === 'global'`, which may be different properties.

**Impact:** May not correctly identify global augmentation scopes.

**Test Coverage:** Global augmentation test cases may fail.

#### 8. Missing Definition File (.d.ts) Handling
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
      (def.type === DefinitionType.TSEnumName && def.node.declare) ||
      (def.type === DefinitionType.TSModuleName && def.node.declare)
    );
  });
}
```

**Go Implementation:**
```go
// This function is completely missing
```

**Issue:** The Go version doesn't handle TypeScript declaration files (.d.ts) with declare modifiers.

**Impact:** May incorrectly flag shadowing in declaration files where it should be ignored.

**Test Coverage:** Declaration file test cases will fail.

#### 9. Incorrect AST Kind Mappings
**TypeScript Implementation:**
```typescript
AST_NODE_TYPES.TSEmptyBodyFunctionExpression
```

**Go Implementation:**
```go
ast.KindEmptyStatement // Incorrect mapping
```

**Issue:** The Go version maps `TSEmptyBodyFunctionExpression` to `KindEmptyStatement`, which is incorrect.

**Impact:** Will not properly identify empty body function expressions.

**Test Coverage:** Function expression related test cases may fail.

#### 10. Missing Scope Exit Processing
**TypeScript Implementation:**
```typescript
return {
  'Program:exit'(node): void {
    const globalScope = context.sourceCode.getScope(node);
    const stack = [...globalScope.childScopes];

    while (stack.length) {
      const scope = stack.pop()!;
      stack.push(...scope.childScopes);
      checkForShadows(scope);
    }
  },
};
```

**Go Implementation:**
```go
// Uses individual node listeners but no comprehensive scope traversal
```

**Issue:** The Go version processes variables immediately upon declaration rather than doing a comprehensive scope analysis at the end.

**Impact:** May miss some shadowing relationships or process them in the wrong order.

**Test Coverage:** Complex nested shadowing cases may behave differently.

### Recommendations
- Implement missing TypeScript-specific edge case handlers (static method generics, external declaration merging)
- Add proper type import detection and type/value shadowing logic
- Implement complete temporal dead zone and hoisting behavior
- Add support for .d.ts file handling with declare modifiers
- Fix AST kind mappings to match TypeScript equivalents
- Implement missing initialization pattern detection for `ignoreOnInitialization` option
- Add comprehensive scope traversal similar to TypeScript implementation
- Enhance function type parameter detection logic
- Improve global augmentation detection to match TypeScript behavior
- Add missing utility functions for complex AST pattern matching

---