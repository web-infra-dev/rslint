# Rule Validation: no-redeclare

## Rule: no-redeclare

### Test File: no-redeclare.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic redeclaration detection, TypeScript declaration merging logic, configuration options handling, message formatting
- ⚠️ **POTENTIAL ISSUES**: ESLint directive comments handling, scope management complexity, builtin globals detection in different environments
- ❌ **INCORRECT**: Missing ESLint directive comment support, incomplete scope traversal, missing function overload handling

### Discrepancies Found

#### 1. ESLint Directive Comments Not Supported
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
// No equivalent handling for ESLint directive comments like /*global b:false*/
```

**Issue:** The Go implementation completely lacks support for ESLint directive comments (e.g., `/*global b:false*/ var b = 1;`). The TypeScript version detects these and reports `redeclaredBySyntax` errors.

**Impact:** Test case `/*global b:false*/ var b = 1;` will not trigger the expected `redeclaredBySyntax` error in the Go version.

**Test Coverage:** The test case with `/*global b:false*/ var b = 1;` will fail.

#### 2. Function Overload Handling Missing
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
// No special handling for TSDeclareFunction or function overloads
```

**Issue:** The Go implementation doesn't filter out `TSDeclareFunction` nodes, which should be ignored as they represent TypeScript function overloads.

**Impact:** Function overloads may incorrectly trigger redeclaration errors when they should be allowed.

**Test Coverage:** The test case with function overloads `function a(): string; function a(): number; function a() {}` might fail.

#### 3. Scope Management Complexity
**TypeScript Implementation:**
```typescript
function checkForBlock(node: TSESTree.Node): void {
  const scope = context.sourceCode.getScope(node);
  if (scope.block === node) {
    findVariablesInScope(scope);
  }
}

// Handles Program scope specially
if (
  scope.type === ScopeType.global &&
  scope.childScopes[0] &&
  scope.block === scope.childScopes[0].block
) {
  findVariablesInScope(scope.childScopes[0]);
}
```

**Go Implementation:**
```go
// Manual scope tracking with enterScope/exitScope
// May not handle all scope edge cases correctly
```

**Issue:** The Go implementation uses manual scope tracking rather than leveraging a proper scope analyzer. This may miss complex scoping scenarios and doesn't handle the special Node.js/ES module scope detection.

**Impact:** Variables may be incorrectly scoped, leading to false positives or missed redeclarations.

**Test Coverage:** Complex scoping scenarios may not work correctly.

#### 4. Incomplete Builtin Globals Environment Detection
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
// Only adds builtins to root scope if not a module
isModuleScope := false
ctx.SourceFile.ForEachChild(func(node *ast.Node) bool {
  if ast.IsImportDeclaration(node) || ast.IsExportDeclaration(node) || ast.IsExportAssignment(node) {
    isModuleScope = true
    return true
  }
  return false
})
```

**Issue:** The Go implementation uses a simple import/export detection for module scope, but doesn't properly integrate with ESLint's global variable system or handle different environments (browser vs Node.js).

**Impact:** Builtin global detection may not work correctly in all environments, leading to missed `redeclaredAsBuiltin` errors.

**Test Coverage:** Tests with `languageOptions.globals` may not work correctly.

#### 5. Declaration Merging Logic Timing
**TypeScript Implementation:**
```typescript
// Processes all declarations first, then applies merging logic
const [declaration, ...extraDeclarations] = iterateDeclarations(variable);
if (extraDeclarations.length === 0) {
  continue;
}
```

**Go Implementation:**
```go
// Reports immediately when second declaration is encountered
if len(varInfo.Declarations) > 1 {
  // Check merging logic here
}
```

**Issue:** The Go implementation checks for redeclarations immediately upon encountering each declaration, rather than collecting all declarations first and then applying the merging logic. This may lead to incorrect reporting order or missed merging scenarios.

**Impact:** The timing of error reporting may be different, and complex merging scenarios might not be handled correctly.

**Test Coverage:** Tests with multiple declarations of the same name may show different behavior.

#### 6. Missing Destructuring Pattern Support
**TypeScript Implementation:**
```typescript
// Automatically handled by scope manager's variable collection
```

**Go Implementation:**
```go
// Manual destructuring pattern handling
if ast.IsObjectBindingPattern(varDecl.Name()) || ast.IsArrayBindingPattern(varDecl.Name()) {
  // Handle destructuring patterns
  var processBindingPattern func(*ast.Node)
  // ... complex recursive logic
}
```

**Issue:** While the Go implementation attempts to handle destructuring, it may miss edge cases or nested patterns that the TypeScript scope manager handles automatically.

**Impact:** Complex destructuring patterns might not be correctly analyzed for redeclarations.

**Test Coverage:** Tests with complex destructuring like `var { a = 0, b: Object = 0 } = {};` need verification.

### Recommendations
- **Add ESLint directive comment parsing**: Implement support for `/*global*/` comments and report `redeclaredBySyntax` errors
- **Implement function overload filtering**: Filter out `TSDeclareFunction` nodes from redeclaration checking
- **Improve scope management**: Consider using a proper scope analyzer or more closely mirror ESLint's scope handling
- **Enhance builtin globals detection**: Properly integrate with environment-specific globals and ESLint's global variable system
- **Refactor declaration merging timing**: Collect all declarations first before applying merging logic
- **Test destructuring patterns thoroughly**: Ensure all destructuring edge cases are handled correctly
- **Add missing AST node types**: Ensure all declaration types (including export declarations) are properly handled

---