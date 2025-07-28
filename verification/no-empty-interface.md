## Rule: no-empty-interface

### Test File: no-empty-interface.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Basic empty interface detection (`interface Foo {}`)
  - Option parsing for `allowSingleExtends`
  - Detection of interfaces with single extends clause
  - Basic fix generation for interface-to-type conversion
  - Type parameter handling in fix generation
  - Class declaration merging detection using TypeChecker

- ⚠️ **POTENTIAL ISSUES**: 
  - Ambient declaration detection logic differs from TypeScript implementation
  - Missing suggestion functionality for cases where auto-fix is not appropriate
  - Error message IDs not defined as constants

- ❌ **INCORRECT**: 
  - Missing proper messageId field in RuleMessage structure
  - Incomplete ambient declaration detection (scope type checking missing)
  - No suggestion support for .d.ts files in declared modules

### Discrepancies Found

#### 1. Missing Message ID Constants
**TypeScript Implementation:**
```typescript
export type MessageIds = 'noEmpty' | 'noEmptyWithSuper';

messages: {
  noEmpty: 'An empty interface is equivalent to `{}`.',
  noEmptyWithSuper: 'An interface declaring no members is equivalent to its supertype.',
}
```

**Go Implementation:**
```go
ctx.ReportNode(interfaceDecl.Name(), rule.RuleMessage{
    Description: "An empty interface is equivalent to `{}`.",
})
```

**Issue:** The Go implementation hardcodes error messages instead of using messageId constants. This affects test compatibility since tests expect specific messageId values.

**Impact:** Test cases that check for specific messageId values (`noEmpty`, `noEmptyWithSuper`) will fail.

**Test Coverage:** All test cases use messageId assertions that won't match.

#### 2. Incomplete Ambient Declaration Detection
**TypeScript Implementation:**
```typescript
const isInAmbientDeclaration =
  isDefinitionFile(context.filename) &&
  scope.type === ScopeType.tsModule &&
  scope.block.declare;
```

**Go Implementation:**
```go
isInAmbientDeclaration := false
if strings.HasSuffix(ctx.SourceFile.FileName(), ".d.ts") {
    // Check if we're inside a declared module
    parent := node.Parent
    for parent != nil {
        if parent.Kind == ast.KindModuleDeclaration {
            moduleDecl := parent.AsModuleDeclaration()
            modifiers := moduleDecl.Modifiers()
            if modifiers != nil {
                for _, modifier := range modifiers.Nodes {
                    if modifier.Kind == ast.KindDeclareKeyword {
                        isInAmbientDeclaration = true
                        break
                    }
                }
            }
        }
        // ...
    }
}
```

**Issue:** The Go implementation only checks for declared modules but doesn't verify the scope type like the TypeScript version. The TypeScript version specifically checks `scope.type === ScopeType.tsModule` which is more precise.

**Impact:** May incorrectly identify ambient declarations, affecting whether fixes or suggestions are provided.

**Test Coverage:** The `.d.ts` test case with declared module should reveal this difference.

#### 3. Missing Suggestion Support
**TypeScript Implementation:**
```typescript
context.report({
  node: node.id,
  messageId: 'noEmptyWithSuper',
  ...(useAutoFix
    ? { fix }
    : !mergedWithClassDeclaration
      ? {
          suggest: [
            {
              messageId: 'noEmptyWithSuper',
              fix,
            },
          ],
        }
      : null),
});
```

**Go Implementation:**
```go
if isInAmbientDeclaration || mergedWithClassDeclaration {
    // Just report without fix or suggestion for ambient declarations or merged class declarations
    ctx.ReportNode(interfaceDecl.Name(), message)
} else {
    // Use auto-fix for non-ambient, non-merged cases
    ctx.ReportNodeWithFixes(interfaceDecl.Name(), message,
        rule.RuleFixReplace(ctx.SourceFile, node, replacement))
}
```

**Issue:** The Go implementation doesn't provide suggestions for cases where auto-fix is disabled but suggestions would be appropriate (ambient declarations without class merging).

**Impact:** Test cases expecting suggestions in `.d.ts` files will fail. The TypeScript version provides suggestions for ambient declarations that don't have merged class declarations.

**Test Coverage:** The declared module test case expects suggestions but Go implementation provides none.

#### 4. RuleMessage Structure Incompatibility
**TypeScript Implementation:**
```typescript
// Uses messageId for structured error reporting
messageId: 'noEmpty'
messageId: 'noEmptyWithSuper'
```

**Go Implementation:**
```go
rule.RuleMessage{
    Description: "An empty interface is equivalent to `{}`.",
}
```

**Issue:** The Go RuleMessage structure uses Description field instead of MessageId, which may not map correctly to the expected test output format.

**Impact:** Tests expecting specific messageId values in error objects will not match the Go implementation output.

**Test Coverage:** All test cases that verify messageId field will fail.

#### 5. Type Parameter Extraction Logic
**TypeScript Implementation:**
```typescript
let typeParam = '';
if (node.typeParameters) {
  typeParam = context.sourceCode.getText(node.typeParameters);
}
```

**Go Implementation:**
```go
var typeParamsText string
if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Nodes) > 0 {
    firstParam := interfaceDecl.TypeParameters.Nodes[0]
    lastParam := interfaceDecl.TypeParameters.Nodes[len(interfaceDecl.TypeParameters.Nodes)-1]
    firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
    lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)
    typeParamsRange := firstRange.WithEnd(lastRange.End())
    // Include the angle brackets
    typeParamsRange = typeParamsRange.WithPos(typeParamsRange.Pos() - 1).WithEnd(typeParamsRange.End() + 1)
    typeParamsText = string(ctx.SourceFile.Text()[typeParamsRange.Pos():typeParamsRange.End()])
}
```

**Issue:** The Go implementation manually reconstructs the type parameters text with angle brackets, while TypeScript uses `getText()` on the entire typeParameters node. This could lead to different text extraction results.

**Impact:** Generated fixes might have slightly different formatting or miss edge cases in type parameter extraction.

**Test Coverage:** Tests with generic interfaces like `interface Foo<T> extends Bar<T> {}` should reveal formatting differences.

### Recommendations
- Add messageId support to RuleMessage structure or ensure Description maps to expected messageId
- Implement proper scope type checking for ambient declaration detection
- Add suggestion support for appropriate cases (ambient declarations without class merging)
- Use more robust type parameter text extraction similar to TypeScript's approach
- Ensure the rule framework supports both fix and suggestion reporting mechanisms
- Add constants for message IDs to maintain consistency with TypeScript implementation

---