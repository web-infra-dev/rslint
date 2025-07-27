# Rule Validation: consistent-type-imports

## Rule: consistent-type-imports

### Test File: consistent-type-imports.test.ts

### Validation Summary
- ✅ **CORRECT**: Basic import classification, type-only detection, prefer option handling, export analysis, type query detection
- ⚠️ **POTENTIAL ISSUES**: Complex export patterns, decorator metadata handling, inline type import fixes, performance optimizations may affect behavior
- ❌ **INCORRECT**: Missing emitDecoratorMetadata integration, incomplete AST node coverage, simplified export analysis, missing advanced fix generation

### Discrepancies Found

#### 1. Decorator Metadata Handling
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

**Issue:** The Go implementation doesn't integrate with the TypeScript compiler options to check for `emitDecoratorMetadata` and `experimentalDecorators`, which is critical for proper rule behavior in decorator-heavy codebases.

**Impact:** The rule will incorrectly report type import violations in files with decorators when `emitDecoratorMetadata: true` and `experimentalDecorators: true` are set, leading to false positives.

**Test Coverage:** The dedicated decorator metadata test suite will fail.

#### 2. Export Pattern Analysis
**TypeScript Implementation:**
```typescript
/**
 * keep origin import kind when export
 * export { Type }
 * export default Type;
 * export = Type;
 */
if (
  (ref.identifier.parent.type ===
    AST_NODE_TYPES.ExportSpecifier ||
   ref.identifier.parent.type ===
     AST_NODE_TYPES.ExportDefaultDeclaration ||
   ref.identifier.parent.type ===
     AST_NODE_TYPES.TSExportAssignment) &&
  ref.isValueReference &&
  ref.isTypeReference
) {
  return node.importKind === 'type';
}
```

**Go Implementation:**
```go
listeners[ast.KindExportDeclaration] = func(node *ast.Node) {
    exportDecl := node.AsExportDeclaration()
    // Only track as value usage if it's not a type-only export
    if !exportDecl.IsTypeOnly && exportDecl.ExportClause != nil {
        // Basic export handling without reference analysis
    }
}
```

**Issue:** The Go implementation doesn't perform the sophisticated reference analysis to determine if exports maintain the original import kind. It lacks the dual reference checking (`ref.isValueReference && ref.isTypeReference`).

**Impact:** Export statements may be incorrectly classified, leading to wrong suggestions about type vs value imports.

**Test Coverage:** Export-related test cases will fail, particularly those testing mixed value/type exports.

#### 3. AST Node Coverage
**TypeScript Implementation:**
```typescript
// Comprehensive coverage including:
case AST_NODE_TYPES.TSTypeQuery:
case AST_NODE_TYPES.TSQualifiedName:
case AST_NODE_TYPES.TSPropertySignature:
case AST_NODE_TYPES.MemberExpression:
// Plus many more node types
```

**Go Implementation:**
```go
// Limited coverage:
listeners[ast.KindTypeReference] = func(node *ast.Node) { /* ... */ }
listeners[ast.KindTypeQuery] = func(node *ast.Node) { /* ... */ }
listeners[ast.KindExpressionStatement] = func(node *ast.Node) { /* ... */ }
// Missing many node types from TypeScript implementation
```

**Issue:** The Go implementation doesn't handle all the AST node types that the TypeScript version covers, particularly for complex type queries and property signatures.

**Impact:** Some type-only usage patterns will not be detected, leading to missed optimization opportunities.

**Test Coverage:** Complex type query test cases will fail.

#### 4. Variable Reference Analysis
**TypeScript Implementation:**
```typescript
const [variable] = context.sourceCode.getDeclaredVariables(specifier);
if (variable.references.length === 0) {
  unusedSpecifiers.push(specifier);
} else {
  const onlyHasTypeReferences = variable.references.every(ref => {
    // Sophisticated reference analysis
  });
}
```

**Go Implementation:**
```go
// Simplified approach using maps:
valueUsedIdentifiers := make(map[string]bool)
allReferencedIdentifiers := make(map[string]bool)
// Basic string-based tracking without full variable scope analysis
```

**Issue:** The Go implementation uses a simplified string-based approach rather than proper variable scope analysis, which can miss complex scoping scenarios.

**Impact:** Variables with the same name in different scopes may be incorrectly analyzed, leading to wrong classifications.

**Test Coverage:** Shadowing and scoping test cases will fail.

#### 5. Fix Generation Complexity
**TypeScript Implementation:**
```typescript
function* fixToTypeImportDeclaration(
  fixer: TSESLint.RuleFixer,
  report: ReportValueImport,
  sourceImports: SourceImports,
): IterableIterator<TSESLint.RuleFix> {
  // Complex fix generation with proper token handling
  // Handles comments, spacing, merging with existing imports
}
```

**Go Implementation:**
```go
func fixToTypeImportDeclaration(sourceFile *ast.SourceFile, report ReportValueImport, sourceImports *SourceImports, fixStyle string) []rule.RuleFix {
    // Simplified fix generation using regex patterns
    importPattern := regexp.MustCompile(`import\s+`)
}
```

**Issue:** The Go implementation uses regex-based fixes rather than proper AST token manipulation, which can break with complex import statements, comments, or unusual formatting.

**Impact:** Auto-fixes may produce malformed code or lose comments.

**Test Coverage:** Fix generation test cases will fail, particularly those with comments or complex import structures.

#### 6. Performance Limitations
**TypeScript Implementation:**
```typescript
// No explicit performance limits on analysis depth
```

**Go Implementation:**
```go
// Multiple performance safeguards:
maxDecls := 50
maxExports := 50
maxChecks := 10
maxDepth := 20
```

**Issue:** The Go implementation adds artificial limits to prevent performance issues, which may cause it to miss violations in large files.

**Impact:** In files with many imports or complex structures, some violations may go undetected.

**Test Coverage:** Large file test cases will show different behavior.

#### 7. Type Parameter Shadowing
**TypeScript Implementation:**
```typescript
// Implicit handling through proper variable scope analysis
```

**Go Implementation:**
```go
func isIdentifierShadowedByTypeParameter(node *ast.Node, identifierName string) bool {
    // Custom implementation that may not match TypeScript's scope resolution
    current := node.Parent
    maxDepth := 20 // Artificial limit
}
```

**Issue:** The Go implementation implements its own type parameter shadowing logic, which may not exactly match TypeScript's scope resolution rules.

**Impact:** Type parameter shadowing scenarios may be handled differently, affecting rule accuracy.

**Test Coverage:** Type parameter shadowing test cases will show discrepancies.

### Recommendations
- **Critical**: Implement proper integration with TypeScript compiler options for decorator metadata
- **High Priority**: Replace regex-based fix generation with proper AST token manipulation
- **High Priority**: Implement comprehensive variable scope analysis instead of string-based tracking
- **Medium Priority**: Add missing AST node type handlers for complete coverage
- **Medium Priority**: Improve export analysis to match TypeScript's reference tracking
- **Low Priority**: Remove or make optional the performance limits that could affect correctness
- **Enhancement**: Add proper comment preservation in fix generation
- **Testing**: Expand test coverage for complex scoping scenarios and decorator metadata cases

---