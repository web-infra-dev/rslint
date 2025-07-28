## Rule: no-redeclare

### Test File: no-redeclare.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic redeclaration detection, TypeScript declaration merging logic, configuration options handling, core error messages
- ⚠️ **POTENTIAL ISSUES**: Global directive comment handling, ESLint scope manager integration, builtin globals detection in modules vs scripts
- ❌ **INCORRECT**: Missing ESLint global directive support, incomplete builtin globals list, scope traversal differences

### Discrepancies Found

#### 1. ESLint Global Directive Comments Not Supported
**TypeScript Implementation:**
```typescript
if (
  'eslintExplicitGlobalComments' in variable &&
  variable.eslintExplicitGlobalComments
) {
  for (const comment of variable.eslintExplicitGlobalComments) {
    yield {
      loc: getNameLocationInGlobalDirectiveComment(
        context.sourceCode,
        comment,
        variable.name,
      ),
      node: comment,
      type: 'comment',
    };
  }
}
```

**Go Implementation:**
```go
// No equivalent implementation found
```

**Issue:** The Go implementation completely lacks support for ESLint global directive comments like `/*global b:false*/`. These comments can declare global variables that should be checked for redeclaration.

**Impact:** Test case `/*global b:false*/ var b = 1;` should produce a `redeclaredBySyntax` error but won't be detected by the Go implementation.

**Test Coverage:** The test case with `/*global b:false*/ var b = 1;` will fail.

#### 2. ESLint Implicit Global Settings Missing
**TypeScript Implementation:**
```typescript
if (
  options.builtinGlobals &&
  'eslintImplicitGlobalSetting' in variable &&
  (variable.eslintImplicitGlobalSetting === 'readonly' ||
    variable.eslintImplicitGlobalSetting === 'writable')
) {
  yield { type: 'builtin' };
}
```

**Go Implementation:**
```go
// No equivalent - only handles hardcoded builtin globals
```

**Issue:** The Go implementation doesn't integrate with ESLint's scope manager to detect implicit global settings. It only uses a hardcoded list of builtin globals.

**Impact:** Test case with `languageOptions: { globals: { top: 'readonly' } }` expects `redeclaredAsBuiltin` error but may not work correctly.

**Test Coverage:** Test case `var top = 0;` with globals configuration will fail.

#### 3. TSDeclareFunction Filtering Missing
**TypeScript Implementation:**
```typescript
const identifiers = variable.identifiers
  .map(id => ({
    identifier: id,
    parent: id.parent,
  }))
  // ignore function declarations because TS will treat them as an overload
  .filter(
    ({ parent }) => parent.type !== AST_NODE_TYPES.TSDeclareFunction,
  );
```

**Go Implementation:**
```go
// No equivalent filtering for TSDeclareFunction
```

**Issue:** The Go implementation doesn't filter out `TSDeclareFunction` nodes, which should be ignored as TypeScript treats them as overloads.

**Impact:** May incorrectly report redeclarations for TypeScript declare function overloads.

**Test Coverage:** Test cases with function overloads `function a(): string; function a(): number; function a() {}` should pass but might fail.

#### 4. Scope Detection Logic Differences
**TypeScript Implementation:**
```typescript
function checkForBlock(node: TSESTree.Node): void {
  const scope = context.sourceCode.getScope(node);
  
  if (scope.block === node) {
    findVariablesInScope(scope);
  }
}
```

**Go Implementation:**
```go
// Manual scope tracking without ESLint scope manager integration
enterScope := func(node *ast.Node) {
  if _, processed := processedScopes[node]; processed {
    return
  }
  // ... manual scope creation
}
```

**Issue:** The Go implementation uses manual scope tracking instead of integrating with a proper scope manager, which may miss subtle scoping rules.

**Impact:** May incorrectly detect or miss redeclarations in complex scoping scenarios.

**Test Coverage:** Complex scoping test cases may behave differently.

#### 5. Module vs Script Context Detection
**TypeScript Implementation:**
```typescript
// Node.js or ES modules has a special scope.
if (
  scope.type === ScopeType.global &&
  scope.childScopes[0] &&
  // The special scope's block is the Program node.
  scope.block === scope.childScopes[0].block
) {
  findVariablesInScope(scope.childScopes[0]);
}
```

**Go Implementation:**
```go
// Check if this is a module by looking for exports or imports
isModuleScope := false
ctx.SourceFile.ForEachChild(func(node *ast.Node) bool {
  if ast.IsImportDeclaration(node) || ast.IsExportDeclaration(node) || ast.IsExportAssignment(node) {
    isModuleScope = true
    return true
  }
  return false
})
```

**Issue:** The Go implementation has a simplified module detection that may not match ESLint's sophisticated scope analysis.

**Impact:** Builtin global detection may work differently between module and script contexts.

**Test Coverage:** Tests with `sourceType: 'module'` vs `sourceType: 'script'` may behave differently.

#### 6. Declaration Merging Logic Edge Cases
**TypeScript Implementation:**
```typescript
// class + interface/namespace merging
if (
  identifiers.every(({ parent }) =>
    CLASS_DECLARATION_MERGE_NODES.has(parent.type),
  )
) {
  const classDecls = identifiers.filter(
    ({ parent }) => parent.type === AST_NODE_TYPES.ClassDeclaration,
  );
  if (classDecls.length === 1) {
    // safe declaration merging
    return;
  }
  
  // there's more than one class declaration, which needs to be reported
  for (const { identifier } of classDecls) {
    yield { loc: identifier.loc, node: identifier, type: 'syntax' };
  }
  return;
}
```

**Go Implementation:**
```go
case ast.KindClassDeclaration:
  // Report only when we encounter exactly the second class
  shouldReport = classCounts == 2
```

**Issue:** The Go implementation reports on the second class declaration immediately, while the TypeScript version reports all duplicate class declarations at once after analysis.

**Impact:** Error reporting timing and count may differ for multiple class declarations.

**Test Coverage:** Test case `class A {} class A {} namespace A {}` expects only one error but Go might report differently.

#### 7. Missing Variable Declaration List Processing
**TypeScript Implementation:**
```typescript
// Uses ESLint scope manager to automatically track variable declarations
```

**Go Implementation:**
```go
ast.KindVariableStatement: func(node *ast.Node) {
  varStmt := node.AsVariableStatement()
  if varStmt.DeclarationList == nil {
    return
  }
  declList := varStmt.DeclarationList.AsVariableDeclarationList()
  for _, decl := range declList.Declarations.Nodes {
    // Manual processing of each declaration
  }
}
```

**Issue:** The Go implementation manually processes variable declarations, which may miss edge cases that ESLint's scope manager handles automatically.

**Impact:** Complex variable declaration patterns may not be handled correctly.

**Test Coverage:** Destructuring and complex declaration patterns may behave differently.

### Recommendations
- Implement ESLint global directive comment parsing and processing
- Add integration with proper scope management system or enhance manual scope tracking
- Add filtering for TSDeclareFunction nodes to match TypeScript behavior
- Enhance module vs script context detection to match ESLint's behavior
- Review and test declaration merging logic for edge cases with multiple declarations
- Add comprehensive builtin globals list or integrate with ESLint's globals
- Test extensively with complex scoping scenarios to ensure compatibility
- Consider implementing iterator-based declaration processing to match TypeScript's approach

---