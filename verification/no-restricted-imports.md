## Rule: no-restricted-imports

### Test File: no-restricted-imports.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic import/export restriction functionality
  - Path and pattern restriction parsing
  - Type-only import allowances
  - Configuration option parsing (paths, patterns)
  - Message customization support
  - ImportEquals declaration handling
  - Export all declaration handling
  - Regex pattern matching with case sensitivity

- ⚠️ **POTENTIAL ISSUES**:
  - Inline type import detection (`import { type Bar }`)
  - Mixed import handling (value + type specifiers)
  - Export all declaration pattern matching
  - Complex glob pattern edge cases

- ❌ **INCORRECT**:
  - Missing export all declaration support in listeners
  - Incomplete type-only detection for mixed imports/exports
  - Message ID inconsistencies

### Discrepancies Found

#### 1. Missing Export All Declaration Handler
**TypeScript Implementation:**
```typescript
return {
  ExportAllDeclaration: rules.ExportAllDeclaration,
  // ... other handlers
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
  ast.KindImportDeclaration: func(node *ast.Node) { ... },
  ast.KindExportDeclaration: func(node *ast.Node) { ... },
  ast.KindImportEqualsDeclaration: func(node *ast.Node) { ... },
  // Missing: ExportAllDeclaration handler
}
```

**Issue:** The Go implementation doesn't handle `export * from 'module'` statements, which should be restricted according to the original rule.

**Impact:** Test case `export * from 'import1'` will not be caught by the Go implementation.

**Test Coverage:** The test case `"export * from 'import1';"` with error `messageId: 'path', type: 'ExportAllDeclaration'` will fail.

#### 2. Inline Type Import Detection
**TypeScript Implementation:**
```typescript
node.specifiers.every(
  specifier =>
    specifier.type === AST_NODE_TYPES.ImportSpecifier &&
    specifier.importKind === 'type',
)
```

**Go Implementation:**
```go
if namedImports.Elements != nil {
  allTypeOnly := true
  for _, elem := range namedImports.Elements.Nodes {
    if !elem.AsImportSpecifier().IsTypeOnly {
      allTypeOnly = false
      break
    }
  }
  return allTypeOnly
}
```

**Issue:** The Go implementation checks `IsTypeOnly` on import specifiers but may not correctly handle inline type imports like `import { type Bar }`.

**Impact:** Test cases with `import { type Bar }` or mixed `import { Bar, type Baz }` may not be handled correctly.

**Test Coverage:** Test cases using inline type syntax may produce incorrect results.

#### 3. Mixed Import/Export Handling
**TypeScript Implementation:**
```typescript
// Handles mixed imports like: import { Bar, type Baz } from 'foo'
// Each specifier is checked individually for type-only status
```

**Go Implementation:**
```go
// Uses allTypeOnly boolean - if ANY specifier is not type-only, 
// treats entire import as non-type-only
```

**Issue:** The Go implementation doesn't handle mixed imports where some specifiers are type-only and others are not. It should report errors for non-type specifiers while allowing type-only ones.

**Impact:** Test case `"import { Bar, type Baz } from 'import-foo';"` expects two separate errors but Go implementation may only report one or handle incorrectly.

**Test Coverage:** The test expecting two `importNameWithCustomMessage` errors for mixed imports will likely fail.

#### 4. Message ID Mapping
**TypeScript Implementation:**
```typescript
// Uses base rule message IDs:
// - 'path'
// - 'pathWithCustomMessage' 
// - 'patterns'
// - 'patternWithCustomMessage'
// - 'importNameWithCustomMessage'
```

**Go Implementation:**
```go
messageId := "path"
if pr.Message != "" {
  messageId = "pathWithCustomMessage"
}
// Similar for patterns and importNames
```

**Issue:** The Go implementation creates message IDs correctly but may not have the exact same message templates as the base ESLint rule.

**Impact:** Error messages may differ from expected TypeScript-ESLint output.

**Test Coverage:** Tests checking specific messageId values should pass, but actual message content may differ.

#### 5. Glob Pattern Exclude Logic
**TypeScript Implementation:**
```typescript
// Uses ignore library which handles negation patterns differently
ignore({
  allowRelativePaths: true,
  ignoreCase: !restrictedPattern.caseSensitive,
}).add(restrictedPattern.group)
```

**Go Implementation:**
```go
// Custom glob implementation handles excludes by checking prefix "!"
isExclude := strings.HasPrefix(pattern, "!")
if isExclude {
  pattern = pattern[1:] // Remove the !
}
```

**Issue:** The exclude logic implementation may not match the ignore library's behavior exactly, particularly for complex patterns.

**Impact:** Test cases with exclusion patterns like `'!import2/good'` may behave differently.

**Test Coverage:** Pattern tests with exclusions may not match expected behavior.

### Recommendations
- Add `ast.KindExportAllDeclaration` handler to the Go implementation listeners
- Improve type-only detection to handle inline type imports (`import { type Bar }`)
- Implement per-specifier checking for mixed imports/exports to match TypeScript behavior
- Verify glob pattern exclude logic matches ignore library behavior
- Add comprehensive test cases for mixed import scenarios
- Validate that message IDs and content match the base ESLint rule exactly
- Test edge cases with complex glob patterns and exclusions

---