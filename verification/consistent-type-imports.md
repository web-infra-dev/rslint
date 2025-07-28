## Rule: consistent-type-imports

### Test File: consistent-type-imports.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic import classification, type vs value usage detection, export handling, fixStyle and prefer options, messageIds structure, type queries with typeof
- ⚠️ **POTENTIAL ISSUES**: Type parameter shadowing logic, performance optimizations may affect completeness, decorator metadata handling incomplete, complex reference analysis
- ❌ **INCORRECT**: Missing export default handling, incomplete qualified name traversal, missing property signature detection, oversimplified fix generation

### Discrepancies Found

#### 1. Export Default Detection Missing
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.ExportDefaultDeclaration:
case AST_NODE_TYPES.TSExportAssignment:
  if (ref.isValueReference && ref.isTypeReference) {
    return node.importKind === 'type';
  }
```

**Go Implementation:**
```go
// TODO: Find the correct way to handle export default
// Currently commented out as KindExportDefault doesn't exist
// listeners[ast.KindExportDefault] = func(node *ast.Node) {
//   // This catches export default Foo
// }
```

**Issue:** The Go implementation is missing proper handling of `export default` statements, which should mark identifiers as value usage. Only `export =` is handled.

**Impact:** May incorrectly classify imports used in `export default` statements as type-only.

**Test Coverage:** Tests with `export default Type` patterns may fail incorrectly.

#### 2. Qualified Name Traversal Incomplete
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TSQualifiedName:
  // TSTypeQuery must have a TSESTree.EntityName as its child, so we can filter here and break early
  if (parent.left !== child) {
    return false;
  }
  child = parent;
  parent = parent.parent;
  continue;
```

**Go Implementation:**
```go
} else if typeRef.TypeName != nil && ast.IsQualifiedName(typeRef.TypeName) {
  // Handle qualified names like foo.Bar
  qualifiedName := typeRef.TypeName.AsQualifiedName()
  if qualifiedName.Left != nil && ast.IsIdentifier(qualifiedName.Left) {
    identifierName := qualifiedName.Left.AsIdentifier().Text
    allReferencedIdentifiers[identifierName] = true
  }
}
```

**Issue:** The Go version doesn't handle the recursive parent traversal logic for complex qualified names in type queries like `typeof foo.bar.baz`.

**Impact:** May misclassify imports used in deeply nested qualified type expressions.

**Test Coverage:** Test case `type Baz = (typeof foo.bar)['Baz']` requires proper qualified name handling.

#### 3. Property Signature Key Detection Missing
**TypeScript Implementation:**
```typescript
case AST_NODE_TYPES.TSPropertySignature:
  return parent.key === child;
case AST_NODE_TYPES.MemberExpression:
  if (parent.object !== child) {
    return false;
  }
  child = parent;
  parent = parent.parent;
  continue;
```

**Go Implementation:**
```go
// Missing: TSPropertySignature handling for computed property keys
// Only has basic property access detection
listeners[ast.KindPropertyAccessExpression] = func(node *ast.Node) {
  propAccess := node.AsPropertyAccessExpression()
  if propAccess.Expression != nil && ast.IsIdentifier(propAccess.Expression) {
    valueUsedIdentifiers[propAccess.Expression.AsIdentifier().Text] = true
  }
}
```

**Issue:** Missing detection of identifiers used as computed property keys in type signatures like `{ [constants.X]: ReadonlyArray<string> }`.

**Impact:** May misclassify imports used as computed property keys in type definitions as type-only when they should be value imports.

**Test Coverage:** Test case with `export type Y = { [constants.X]: ReadonlyArray<string>; }` tests this pattern.

#### 4. Decorator Metadata Compiler Options Missing
**TypeScript Implementation:**
```typescript
const emitDecoratorMetadata =
  getParserServices(context, true).emitDecoratorMetadata ?? false;
const experimentalDecorators =
  getParserServices(context, true).experimentalDecorators ?? false;
if (experimentalDecorators && emitDecoratorMetadata) {
  selectors.Decorator = (): void => {
    hasDecoratorMetadata = true;
  };
}
```

**Go Implementation:**
```go
// Check for decorator metadata compatibility
emitDecoratorMetadata := false
experimentalDecorators := false
// Note: For now, we'll skip the compiler options check as the API may not be available
// This is a simplification for the Go port
```

**Issue:** The Go version doesn't properly check TypeScript compiler options for decorator metadata settings, which affects when the rule should be disabled.

**Impact:** The special case where decorator metadata should disable the rule entirely won't work correctly.

**Test Coverage:** The separate test suite for `experimentalDecorators: true + emitDecoratorMetadata: true` won't behave as expected.

#### 5. Variable Reference Analysis Oversimplified
**TypeScript Implementation:**
```typescript
const [variable] = context.sourceCode.getDeclaredVariables(specifier);
if (variable.references.length === 0) {
  unusedSpecifiers.push(specifier);
} else {
  const onlyHasTypeReferences = variable.references.every(ref => {
    // Complex reference position analysis including:
    // - Export preservation logic
    // - Type query context checking
    // - Qualified name traversal
    // - Property signature key detection
    // - Member expression object checking
  });
}
```

**Go Implementation:**
```go
// Manual tracking with maps
if !allReferencedIdentifiers[identifierName] || shadowedIdentifiers[identifierName] {
  *unusedSpecifiers = append(*unusedSpecifiers, defaultImport)
} else if valueUsedIdentifiers[identifierName] {
  *valueSpecifiers = append(*valueSpecifiers, defaultImport)
} else {
  // Check if all references are shadowed by type parameters
  if areAllReferencesTypeParameterShadowed(identifierName, allReferencedNodes) {
    *unusedSpecifiers = append(*unusedSpecifiers, defaultImport)
  } else {
    *typeSpecifiers = append(*typeSpecifiers, defaultImport)
  }
}
```

**Issue:** The Go version uses simplified manual tracking instead of leveraging the comprehensive variable reference system that ESLint provides.

**Impact:** May miss complex reference patterns, especially the sophisticated logic for determining if references are in type-only positions.

**Test Coverage:** Complex scoping scenarios and edge cases around type vs value usage detection.

#### 6. Fix Generation Logic Oversimplified
**TypeScript Implementation:**
```typescript
function* fixToTypeImportDeclaration(
  fixer: TSESLint.RuleFixer,
  report: ReportValueImport,
  sourceImports: SourceImports,
): IterableIterator<TSESLint.RuleFix> {
  // 200+ lines of sophisticated fix generation handling:
  // - Token-based manipulation with nullThrows safety
  // - Complex import type classification (default, named, namespace)
  // - Comment preservation
  // - Existing type import merging
  // - Import attribute handling
  // - Multiple fix coordination
}
```

**Go Implementation:**
```go
func fixToTypeImportDeclaration(sourceFile *ast.SourceFile, report ReportValueImport, sourceImports *SourceImports, fixStyle string) []rule.RuleFix {
  // ~100 lines of basic fix generation with:
  // - Regex-based text manipulation
  // - Basic import categorization
  // - Limited comment handling
  // - Simplified merging logic
}
```

**Issue:** The Go fix generation is significantly simpler and uses regex instead of token-based manipulation, missing many edge cases.

**Impact:** Auto-fixes may not work correctly for complex import patterns, especially those with comments, mixed imports, or unusual formatting.

**Test Coverage:** Tests with complex mixed import scenarios and comments in import statements may produce incorrect fixes.

#### 7. Performance Limits May Affect Completeness
**Go Implementation:**
```go
// Multiple hardcoded limits throughout the code:
maxDecls := 50
maxExports := 50
maxChecks := 10
maxReferences := 50

// Example usage:
if processed >= maxDecls {
  break
}
if len(refs) < 50 {
  allReferencedNodes[identifierName] = append(refs, node)
}
```

**Issue:** The Go version includes arbitrary limits on processing declarations, exports, and references that may cause it to miss issues in large files.

**Impact:** May not detect all type-only imports in large files with many declarations, potentially missing optimization opportunities.

**Test Coverage:** Large files with many imports would reveal incomplete analysis.

#### 8. JSX and React Handling Missing
**TypeScript Implementation:**
```typescript
// Tests include JSX-specific scenarios:
{
  code: `
import React from 'react';
export const ComponentFoo: React.FC = () => {
  return <div>Foo Foo</div>;
};
  `,
  languageOptions: {
    parserOptions: {
      ecmaFeatures: { jsx: true },
    },
  },
},
```

**Go Implementation:**
```go
// No specific JSX or React pragma handling
// Missing jsx-related identifier tracking
```

**Issue:** The Go implementation doesn't have specific handling for JSX pragmas and React patterns that affect type vs value usage.

**Impact:** May incorrectly classify React imports or JSX-related identifiers.

**Test Coverage:** JSX test cases involving React, Fragment, and custom JSX pragmas.

### Recommendations
- Implement proper `export default` detection by finding the correct AST node type (likely `ExportDeclaration` with default export)
- Enhance qualified name handling to support recursive parent traversal for complex type expressions
- Add missing property signature key detection for computed property types like `{ [key]: type }`
- Implement proper decorator metadata checking by accessing TypeScript compiler options through the Go bindings
- Expand fix generation logic to use token-based manipulation instead of regex, matching the TypeScript implementation's sophistication
- Remove or make configurable the artificial processing limits that may cause incomplete analysis
- Add JSX and React-specific identifier tracking
- Implement the comprehensive reference position analysis that matches ESLint's variable reference system
- Add extensive test coverage for all identified edge cases
- Consider implementing a more sophisticated AST traversal system that matches the TypeScript implementation's parent-child relationship checking

---