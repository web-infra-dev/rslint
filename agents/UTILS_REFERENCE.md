# Rslint Utility Functions Reference

This document provides a comprehensive reference for utility functions available in `internal/utils/`. These utilities are commonly used when implementing lint rules.

> **Note**: This is a reference document for [PORT_RULE.md](./PORT_RULE.md). See that document for the complete rule porting workflow.

---

## Overview

| File                        | Description                                          |
| --------------------------- | ---------------------------------------------------- |
| `utils.go`                  | Basic utilities (text range, comments, collections)  |
| `ts_api_utils.go`           | TypeScript API utilities (type checking, signatures) |
| `ts_eslint.go`              | TypeScript-ESLint compatibility utilities            |
| `builtin_symbol_likes.go`   | Builtin symbol checking (Promise, Error, etc.)       |
| `type_matches_specifier.go` | Type specifier matching for rule options             |
| `set.go`                    | Generic Set data structure                           |

---

## `internal/utils/utils.go` - Basic Utilities

```go
import "github.com/web-infra-dev/rslint/internal/utils"
```

### Text Range & Comments

```go
// Get trimmed text range of a node (removes leading/trailing whitespace)
trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, node)

// Get all comments in a range (returns iterator)
for comment := range utils.GetCommentsInRange(ctx.SourceFile, textRange) {
    // comment is ast.CommentRange
}

// Check if there are comments in a range
hasComments := utils.HasCommentsInRange(ctx.SourceFile, textRange)
```

### Type Recursion

```go
// Recursively check each part of union/intersection types
utils.TypeRecurser(t, func(subType *checker.Type) bool {
    // Return true to stop recursion, false to continue
    return false
})
```

### AST Helpers

```go
// Get heritage clauses of a class/interface
heritageClauses := utils.GetHeritageClauses(classNode) // *ast.NodeList

// Check if a node has a specific modifier (e.g., async, static, public)
isAsync := utils.IncludesModifier(funcNode, ast.KindAsyncKeyword)

// Whether a node could plausibly evaluate to an Error object — mirrors
// ESLint's astUtils.couldBeError. Unwraps parens + TS assertions internally.
// Used by no-throw-literal, prefer-promise-reject-errors, etc.
mayBeError := utils.CouldBeError(node)

// Whether a node, after unwrapping parens + TS assertions, is the literal
// identifier `undefined`. Lexical check only — does not detect `void 0`.
isUndef := utils.IsUndefinedIdentifier(node)
```

### Collection Operations (Generic)

```go
// Filter, Map, Some, Every, Flatten - Similar to JS array methods
filtered := utils.Filter(items, func(item T) bool { return condition })
mapped := utils.Map(items, func(item T) U { return transform(item) })
hasAny := utils.Some(items, func(item T) bool { return condition })
allMatch := utils.Every(items, func(item T) bool { return condition })
```

### Rule Options Parsing

```go
// Extract options map from rule options (handles both array and object format)
optsMap := utils.GetOptionsMap(options)
if optsMap != nil {
    // Parse options from optsMap...
}
```

### String Comparison

```go
// Natural sort comparison: embedded numbers compared numerically (e.g., "a2" < "a10")
// Returns -1, 0, or 1
result := utils.NaturalCompare("item2", "item10") // -1
```

### Other Utilities

```go
// Create a pointer (useful for optional fields)
ptr := utils.Ref(value) // *T

// Check if a character is a JS whitespace character
isWhite := utils.IsStrWhiteSpace(r)
```

---

## `internal/utils/ts_api_utils.go` - TypeScript API Utilities

### Type Flag Checking

```go
// Check type flags
utils.IsTypeFlagSet(t, checker.TypeFlagsString)  // Check any TypeFlags
utils.IsUnionType(t)           // Is union type
utils.IsIntersectionType(t)    // Is intersection type
utils.IsTypeAnyType(t)         // Is any type
utils.IsTypeUnknownType(t)     // Is unknown type
utils.IsObjectType(t)          // Is object type
utils.IsTypeParameter(t)       // Is type parameter
utils.IsIntrinsicType(t)       // Is intrinsic type
utils.IsIntrinsicErrorType(t)  // Is error type
utils.IsIntrinsicVoidType(t)   // Is void type

// Boolean literal types
utils.IsBooleanLiteralType(t)
utils.IsTrueLiteralType(t)
utils.IsFalseLiteralType(t)
```

### Union/Intersection Type Splitting

```go
// Get all parts of a union type
parts := utils.UnionTypeParts(t)       // []*checker.Type
parts := utils.IntersectionTypeParts(t)
```

### Function Signatures

```go
// Get call signatures
sigs := utils.GetCallSignatures(typeChecker, t)
sigs := utils.GetConstructSignatures(typeChecker, t)
sigs := utils.CollectAllCallSignatures(typeChecker, t) // Including union types
```

### Symbol Checking

```go
utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsFunction)
```

### Promise/Callback Checking

```go
// Check if a parameter is a callback function
isCallback := utils.IsCallback(typeChecker, paramSymbol, node)

// Check if a type is thenable (has .then method)
isThen := utils.IsThenableType(typeChecker, node, t)
```

### Token and Comment Iteration

```go
// Iterate over all tokens of a node
utils.ForEachToken(node, func(token *ast.Node) {
    // Process each token
}, ctx.SourceFile)

// Iterate over all comments of a node
utils.ForEachComment(node, func(comment *ast.CommentRange) {
    // Process each comment
}, ctx.SourceFile)

// Get all children of a node (including tokens)
children := utils.GetChildren(node, ctx.SourceFile)
```

### Compiler Options

```go
// Check strict mode options
isEnabled := utils.IsStrictCompilerOptionEnabled(options, options.NoImplicitAny)
```

---

## `internal/utils/ts_eslint.go` - TypeScript-ESLint Compatibility

### Constraint Types

```go
// Get constraint info (handles generics)
constraintType, isTypeParameter := utils.GetConstraintInfo(typeChecker, t)

// Get constrained type at location
constrainedType := utils.GetConstrainedTypeAtLocation(typeChecker, node)
```

### Await Checking

```go
// Check if a type needs to be awaited
awaitable := utils.NeedsToBeAwaited(typeChecker, node, t)
// Returns: TypeAwaitableAlways | TypeAwaitableNever | TypeAwaitableMay
```

### Type Names

```go
// Get string name of a type
typeName := utils.GetTypeName(typeChecker, t)
```

### Array Method Checking

```go
// Check if it's an array method call with predicate (every, filter, find, etc.)
isArrayMethod := utils.IsArrayMethodCallWithPredicate(typeChecker, callExpr)
```

### Declaration Retrieval

```go
// Get the declaration of a variable
decl := utils.GetDeclaration(typeChecker, node)

// Check if it's a rest parameter declaration
isRest := utils.IsRestParameterDeclaration(decl)
```

### Any Type Checking

```go
// Check if type is any[]
isAnyArray := utils.IsTypeAnyArrayType(t, typeChecker)

// Check if type is unknown[]
isUnknownArray := utils.IsTypeUnknownArrayType(t, typeChecker)

// Discriminate any type
anyType := utils.DiscriminateAnyType(t, typeChecker, program, node)
// Returns: DiscriminatedAnyTypeAny | DiscriminatedAnyTypePromiseAny |
//          DiscriminatedAnyTypeAnyArray | DiscriminatedAnyTypeSafe
```

### Unsafe Assignment Checking

```go
// Check for unsafe any-to-non-any assignment
receiver, sender, unsafe := utils.IsUnsafeAssignment(t, receiverT, typeChecker, senderNode)
```

### Contextual Types

```go
// Get contextual type of a node
contextualType := utils.GetContextualType(typeChecker, node)
```

### Property Names & Binding Patterns

```go
// Get static property name from a property name node
// Handles Identifier, StringLiteral, NumericLiteral, ComputedPropertyName
// (with static string, numeric, BigInt, or template literal expressions)
name, isStatic := utils.GetStaticPropertyName(nameNode)

// Normalize numeric literal to its canonical form (matches ESLint's String(node.value))
// e.g., "0x1" -> "1", "1.0" -> "1", "1e2" -> "100"
normalized := utils.NormalizeNumericLiteral(text)

// Normalize BigInt literal to its decimal string representation
// e.g., "1n" -> "1", "0x1n" -> "1", "0o1n" -> "1", "0b1n" -> "1"
normalized := utils.NormalizeBigIntLiteral(text)

// Recursively collect all identifier names from a binding pattern
// Handles Identifier, ObjectBindingPattern, ArrayBindingPattern
utils.CollectBindingNames(nameNode, func(ident *ast.Node, name string) {
    // Process each identifier
})
```

### Other Helpers

```go
// Get this expression
thisExpr := utils.GetThisExpression(node)

// Get parent function node
parentFunc := utils.GetParentFunctionNode(node)

// Get for statement head location (for reporting)
headLoc := utils.GetForStatementHeadLoc(ctx.SourceFile, forNode)

// Check if expression precedence is higher than await
isHigher := utils.IsHigherPrecedenceThanAwait(node)

// Check if it's a strong precedence node (no parentheses needed)
isStrong := utils.IsStrongPrecedenceNode(innerNode)

// Check if it's a parenthesis-less arrow function
isParenless := utils.IsParenlessArrowFunction(node)

// Get member name
name, nameType := utils.GetNameFromMember(ctx.SourceFile, member)
// nameType: MemberNameTypePrivate | MemberNameTypeQuoted | MemberNameTypeNormal | MemberNameTypeExpression
```

### Enum Types

```go
// Get enum literals
enumLiterals := utils.GetEnumLiterals(t)

// Get enum types
enumTypes := utils.GetEnumTypes(typeChecker, t)
```

---

## `internal/utils/builtin_symbol_likes.go` - Builtin Symbol Checking

### Builtin Type Checking

```go
// Check if it's a Promise type (including derived classes)
isPromise := utils.IsPromiseLike(program, typeChecker, t)

// Check if it's a PromiseConstructor type
isPromiseCtor := utils.IsPromiseConstructorLike(program, typeChecker, t)

// Check if it's an Error type (including derived classes)
isError := utils.IsErrorLike(program, typeChecker, t)

// Check if it's a Readonly<Error> type
isReadonlyError := utils.IsReadonlyErrorLike(program, typeChecker, t)

// Check if it's a Readonly<T> type
isReadonly := utils.IsReadonlyTypeLike(program, typeChecker, t, predicate)
```

### Generic Builtin Symbol Checking

```go
// Check if it's a builtin symbol with specified names
isBuiltin := utils.IsBuiltinSymbolLike(program, typeChecker, t, "Promise", "Error")

// Check if it's any builtin symbol
isAnyBuiltin := utils.IsAnyBuiltinSymbolLike(program, typeChecker, t)

// Check if a symbol is from the default library
isFromLib := utils.IsSymbolFromDefaultLibrary(program, symbol)

// Check if a source file is the default library
isLibFile := utils.IsSourceFileDefaultLibrary(program, sourceFile)
```

---

## `internal/utils/type_matches_specifier.go` - Type Specifier Matching

Used for specifying type sources in rule options (e.g., allowing certain types from specific packages).

```go
// Define type specifier
specifier := utils.TypeOrValueSpecifier{
    From:    utils.TypeOrValueSpecifierFromPackage, // FromFile | FromLib | FromPackage
    Name:    utils.NameList{"SomeType"},
    Package: "some-package", // Used when From == FromPackage
    Path:    "./some-file",  // Used when From == FromFile
}

// Check if a type matches a specifier
matches := utils.TypeMatchesSomeSpecifier(t, specifiers, inlineSpecifiers, program)

// With callee names (for export aliases like `export { test as it }`)
matches := utils.TypeMatchesSomeSpecifierWithCalleeNames(t, specifiers, inlineSpecifiers, program, calleeNames)
```

---

## `internal/utils/set.go` - Set Data Structure

Generic Set implementation for efficient membership checking.

```go
// Create a Set
set := utils.NewSetFromItems("a", "b", "c")
set := utils.NewSetWithSizeHint[string](100)

// Set operations
set.Add("d")
set.Delete("a")
exists := set.Has("b")
length := set.Len()
set.Clear()
```

---

## `shim/ast/` - AST Utilities

```go
import "github.com/microsoft/typescript-go/shim/ast"
```

Reach for these **before** writing a helper of your own — the shim already covers a wide surface. This list is curated to the functions most commonly reused when porting rules; see `shim/ast/shim.go` for the full inventory.

### Navigation (paren-transparency)

Use these instead of hand-rolled loops. See [AST_PATTERNS.md § ParenthesizedExpression](./AST_PATTERNS.md#parenthesizedexpression).

- `ast.SkipParentheses(node)` — innermost non-paren expression
- `ast.WalkUpParenthesizedExpressions(node)` — first non-paren ancestor

### Optional chain

- `ast.IsOptionalChain(node)` — true iff node is in an optional chain
- `ast.IsOptionalChainRoot(node)` — true iff node introduces the chain
- `ast.IsOutermostOptionalChain(node)` — true iff node is the top of its chain
- Flag: `ast.NodeFlagsOptionalChain`

### Node-kind predicates (prefer over `node.Kind == ast.Kind*`)

- `ast.IsStringLiteralLike(node)` — covers `KindStringLiteral` + `KindNoSubstitutionTemplateLiteral`
- `ast.IsTemplateLiteralKind(kind)` — any of the `KindTemplate*` family
- `ast.IsNumericLiteral(node)`
- `ast.IsIdentifier(node)`
- `ast.IsCallExpression(node)`, `ast.IsNewExpression(node)`
- `ast.IsBinaryExpression(node)` (check `OperatorToken.Kind` separately for specific operator matching)
- `ast.IsPropertyAccessExpression(node)`, `ast.IsElementAccessExpression(node)`
- `ast.IsAssignmentExpression(node, excludeCompoundAssignment)`
- `ast.IsBindingPattern(node)` — object / array destructuring
- `ast.IsClassLike(node)` — class declaration or expression
- `ast.IsIterationStatement(node, lookInLabeledStatements)` — for/while variants

### Function-like

- `ast.IsFunctionLike(node)` / `ast.IsFunctionLikeDeclaration(node)` — covers all 7 function-like kinds
- `ast.IsFunctionLikeOrClassStaticBlockDeclaration(node)`
- `ast.GetThisContainer(node, …, …)` — **warning**: treats `PropertyDeclaration`, `ClassStaticBlockDeclaration`, `ModuleDeclaration`, etc. as `this` containers, which does NOT match ESLint's scope model; verify before reusing

### Modifiers

- `ast.HasSyntacticModifier(node, flags)` — bit-flag check (readonly, static, abstract, …)
- `ast.GetCombinedModifierFlags(node)`

---

## `shim/scanner/` - Scanner Utilities

```go
import "github.com/microsoft/typescript-go/shim/scanner"
```

### SkipTrivia

Skip whitespace, comments (line and block), BOM, shebang, and conflict markers to find the next token position:

```go
// Find the start position of the next meaningful token
sourceText := ctx.SourceFile.Text()
nextTokenPos := scanner.SkipTrivia(sourceText, startPos)
```

### Identifier character classification

Unicode-aware, matches the scanner's own lexing rules. Use these instead of hand-written `[A-Za-z_$]` checks.

- `scanner.IsIdentifierStart(ch rune) bool` — valid first character of an identifier
- `scanner.IsIdentifierPart(ch rune) bool` — valid non-first character
- `scanner.IsValidIdentifier(s string) bool` — whole-string check (start + parts + no reserved-word collision)

### GetScannerForSourceFile

Create a token-by-token scanner for more complex scanning needs. See [AST_PATTERNS.md](./AST_PATTERNS.md#token-scanning) for usage examples.

---

## `shim/checker/` - TypeChecker Native Methods

```go
import "github.com/microsoft/typescript-go/shim/checker"
```

For type-aware rules, **check `internal/utils/ts_api_utils.go` and `internal/utils/ts_eslint.go` first** — they wrap the common patterns with the correct invariants (e.g. `IsPromiseLike` handles subclass resolution, `NeedsToBeAwaited` handles generic constraints). Only fall through to the raw `Checker_*` functions below when no wrapper exists. Do **not** hand-roll type analysis on top of AST shape alone — the checker already answers those questions authoritatively.

### Resolution & Signatures

- `checker.Checker_getResolvedSignature(c, callLike)` — resolve the actual signature chosen for a call / new / decorator
- `checker.Checker_getSignaturesOfType(c, t, kind)` — all call/construct signatures of a type
- `checker.Checker_getReturnTypeOfSignature(c, sig)` — return type of a signature

### Type Structure

- `checker.Checker_getApparentType(c, t)` — apparent type (for member lookup; widens primitives to wrapper types)
- `checker.Checker_getWidenedType(c, t)` — widened type (e.g. literal → base)
- `checker.Checker_getBaseTypes(c, t)` — base types of an interface / class
- `checker.Checker_getTypeArguments(c, t)` — type arguments of a generic instance
- `checker.Checker_getBaseConstraintOfType(c, t)` — constraint of a type parameter (nil if unconstrained)

### Type / Symbol Navigation

- `checker.Checker_getTypeOfSymbol(c, sym)` — type of a symbol at its declaration
- `checker.Checker_getPropertyOfType(c, t, name)` — look up a property symbol by name
- `checker.Checker_getPropertiesOfType(c, t)` — all properties of a type
- `checker.Checker_getIndexInfosOfType(c, t)` — index signatures of a type
- `checker.Checker_getIndexTypeOfType(c, t, kind)` — value type for a given index kind

### From AST Nodes

- `checker.Checker_getTypeFromTypeNode(c, typeNode)` — type from a syntactic type annotation (`number`, `Foo<T>`, etc.)
- `checker.Checker_isArrayType(c, t)` — array-type classification
- `checker.IsTupleType(t)` — package-level tuple-type classification (does not need a Checker receiver)

### Type / Symbol Field Accessors

Internal struct fields that Go cannot expose via methods across packages are reachable through top-level accessors:

- `checker.Type_flags(t)` — `TypeFlags` bitset
- `checker.Type_symbol(t)` — associated symbol (for named types)

See `shim/checker/shim.go` for the full surface (~50 functions). If you find yourself reaching for a method that isn't exposed, add it to the shim rather than duplicating the logic.

---

## See Also

- [PORT_RULE.md](./PORT_RULE.md) - Main rule porting workflow
- [AST_PATTERNS.md](./AST_PATTERNS.md) - AST traversal patterns and examples
- [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - Commands and checklist
