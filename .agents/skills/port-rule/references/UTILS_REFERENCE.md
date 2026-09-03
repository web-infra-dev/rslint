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

Values a rule is handed were written for JavaScript, and Go's standard library answers a nearby but different question about each of them. These packages hold the readings that match — see [JavaScript Semantics](#javascript-semantics-ecmascript-minimatch3-isglob).

| Package                   | Description                                                      |
| ------------------------- | ---------------------------------------------------------------- |
| `utils/ecmascript`        | Trim, blank, whitespace, upper/lower case, number-to-string      |
| `utils/ecmascript/regexp` | Compiles and matches a JavaScript RegExp, and `/i` comparison    |
| `utils/unicode17`         | A general category — `\p{Lu}`, `\p{L}`, `\p{M}` — as Node has it |
| `utils/minimatch3`        | Glob matching the way minimatch 3 does — what plugins pin        |
| `utils/isglob`            | "Was this written as a glob, or is it a plain path?"             |

---

## `internal/utils/utils.go` - Basic Utilities

```go
import "github.com/web-infra-dev/rslint/internal/utils"
```

### Text Range & Comments

```go
// Get trimmed text range of a node (removes leading/trailing whitespace)
trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, node)

// Get all comments in a range (returns iterator). Scans locally from
// textRange.Pos() via the scanner — cheap even when called once per node.
for comment := range utils.GetCommentsInRange(ctx.SourceFile, textRange) {
    // comment is ast.CommentRange
}

// Check if there are comments in a range. Same cost profile as
// GetCommentsInRange above — safe to call per-node.
hasComments := utils.HasCommentsInRange(ctx.SourceFile, textRange)

// Get a read-only view of every comment overlapping an arbitrary [start, end)
// span, or just check whether one exists. Both binary-search the file's
// already-collected comment list — pass ctx.Comments.All(), don't rebuild it.
// Prefer this over GetCommentsInRange/HasCommentsInRange when you're calling
// it once per matching AST node across the whole file (e.g. once per boolean
// cast, once per label) — see the "Token and Comment Iteration" section below
// for why that distinction matters.
comments := utils.CommentsInSpan(ctx.Comments.All(), start, end)
hasComment := utils.HasCommentInSpan(ctx.Comments.All(), start, end)
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

Rule options do not require an `internal/utils` helper. The framework passes
ESLint's normalized `context.options` array to `Run`; guard the slice length
before reading the first positional option:

```go
func parseOptions(options []any) Options {
    opts := Options{/* defaults */}
    if len(options) == 0 {
        return opts
    }
    optsMap, _ := options[0].(map[string]any)
    if value, ok := optsMap["someOption"].(bool); ok {
        opts.SomeOption = value
    }
    return opts
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
// Whitespace, trimming, and every other question ECMAScript specifies live in
// utils/ecmascript — see the JavaScript Semantics section below.
isWhite := ecmascript.IsWhiteSpaceOrLineTerminator(r)
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

**Performance: don't call `ForEachComment`/`ForEachToken` on the whole file.**

`ForEachToken` walks `node`'s entire subtree down to the token level (recursive
`GetChildren`), and `ForEachComment` additionally re-scans trivia at every
token it visits. Called on a small subtree (a single function body, a single
expression) this is fine. Called on `ctx.SourceFile.AsNode()` — i.e. the whole
file — it re-tokenizes the entire file, and doing that once per enabled rule
adds up fast (this dominated a real profile: 17.4% cumulative time, larger
than several other perf fixes combined, from ~8 rules each independently
re-deriving the same file-wide comment list).

The linter already walks the whole file once per file — to build the
directive-comment list for `DisableManager` and inline `/* global */` parsing
— and hands you the result:

```go
// ctx.Comments.All() is every comment in the file, in source order, computed
// once per file by the linter. Use this instead of
// ForEachComment(ctx.SourceFile.AsNode(), ...).
for _, comment := range ctx.Comments.All() {
    // comment is *ast.CommentRange
}
```

Rules of thumb:

- Need every comment in the file → iterate `ctx.Comments.All()`. Don't call
  `ForEachComment` on `ctx.SourceFile.AsNode()` yourself.
- Need the comments in a bounded, non-adjacent span (e.g. between two tokens
  that aren't parent/child) → call `utils.CommentsInSpan` with
  `ctx.Comments.All()`, `start`, and `end`. If only existence matters, use
  `utils.HasCommentInSpan`. Both binary-search the sorted comment list (mirrors
  ESLint's `TokenStore#commentsExistBetween`) instead of scanning again. This
  matters most when the check runs once per matching AST node (once per
  boolean cast, once per label) — a per-node full-file scan is the
  expensive pattern to avoid.
- Need comments strictly inside one specific node's own span, and only call
  this a handful of times (not once per node across the file) →
  `utils.HasCommentInsideNode(sourceFile, node)` or `ForEachComment(node, ...)`
  scoped to that node are fine as-is; the walk is bounded by the node's size,
  not the file's.
- Need comments adjacent to a small range you already have the bounds for
  (not "the whole file") → `utils.GetCommentsInRange`/`HasCommentsInRange`
  scan locally from `range.Pos()` via the scanner and are cheap to call
  per-node.

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

// Access-expression helpers for `object.key` / `object[key]`
name, isStatic := utils.AccessExpressionStaticName(accessNode)
object := utils.AccessExpressionObject(accessNode)

// Whether an element access has a finite integer numeric-literal key.
// Parenthesized keys are transparent; unary signs and TS assertions are not.
isIntegerIndex := utils.IsIntegerElementAccess(accessNode)

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

// Get the declaration-identifier range exposed by TSESTree. For plain
// variable/parameter identifiers this includes TS `!`, `?`, and type
// annotations where TSESTree folds them into the Identifier range.
identifierRange := utils.GetESTreeBindingIdentifierRange(sourceFile, identifier)

// Collect active TypeScript default-library names from the type space.
utils.AddDefaultLibraryTypeGlobalNames(typeGlobals, program, typeChecker)
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

## `internal/rule/ref_store.go` - Reference Index (ctx.Refs)

The lazily built per-file identifier-reference index — rslint's stand-in for ESLint's `variable.references`. Two methods: `Resolve(node)` (identifier → symbol, binder scope walk first, falling back to the checker for symbols declared outside this file when a TypeChecker is available) and `References(sym)` (symbol → every referencing identifier in this file, same fallback trigger, plus one checker call per top-level symbol of a global script file to reconcile it with its checker-merged identity).

```go
// Guard first: ctx.Refs is nil when no program is available (JS-only runs).
if ctx.Refs == nil {
    return
}

// Query with the BINDER symbol from the declaration node —
// decl.Symbol(), never checker.GetSymbolAtLocation (merged symbols miss).
refs := ctx.Refs.References(decl.Symbol()) // []*ast.Node, source order, read-only
```

- Declaration names are excluded; reads and writes are both included.
- Property names, import/export bindings, labels, and lowercase JSX tags are pre-filtered out; shadowing and type-vs-value positions resolve correctly.
- `References` always searches this file only — cross-file references (uses in _other_ files) still need the checker — but the symbol it's queried with can be declared anywhere: pass a `Resolve` result obtained through its checker fallback to find in-file references to a global or ambient declaration too.
- Never hand-roll any of this with an AST walk + `GetSymbolAtLocation` per identifier, and never hand-roll your own "try ctx.Refs, then fall back to the checker" wrapper — `Resolve` already is that wrapper.
- To check locality (e.g. "is this callee shadowed by a local declaration") once `Resolve` returns a symbol, pair it with `utils.IsSymbolDeclaredInFile(symbol, ctx.SourceFile)` (`internal/utils/ts_eslint.go`): it walks `symbol.Declarations` for one that binds a name in this file. Declarations inside a `declare global` block or a module augmentation are not among them — they extend the ambient symbol, so `new Function()` in a file with `declare global { namespace Function {} }` still reaches the builtin.
- For built-in **value** shadowing specifically, use `utils.IsValueSymbolDeclaredInFile` instead: it narrows that to the declarations that bind a **value** name. Any `namespace`/`module` declaration counts, whatever it contains; interfaces and type aliases don't, because they merge into the ambient global's symbol (`interface Map {}` in a global script attaches to the same symbol as lib.d.ts's `Map`, and a call to `Map()` still reaches the global).

See [AST_PATTERNS.md — Resolving Identifiers and Collecting References](./AST_PATTERNS.md#resolving-identifiers-and-collecting-references-ctxrefs) for the full semantics and worked examples.

---

## `internal/utils/scope/` - ESLint Scope Model

An AST-derived reconstruction of ESLint's lexical scope tree — rslint's stand-in for `sourceCode.getScope()` / eslint-scope. Use it only when a rule's semantics are stated in terms of **scopes** (which scope owns a binding, what an inner scope shadows, whether two positions share a variable scope). For "which symbol is this identifier" and "where else is this symbol used", reach for `ctx.Refs` above instead — it is backed by the binder and understands globals, ambient, and cross-file declarations, which this package deliberately does not.

```go
manager := scope.Build(ctx.SourceFile)

for _, s := range manager.Scopes { // creation order == pre-order walk of the tree
    for _, v := range s.Vars {     // declaration order within the scope
        // v.Name, v.ID (identifier node), v.DefNode, v.Kind (scope.Def*)
    }
}

// Resolve a name outward, the way ESLint's scope chain does.
for s := start; s != nil; s = s.Parent {
    if declarations := s.Declarations(name); len(declarations) > 0 { /* ... */ }
}
```

- Scope kinds (`scope.Kind*`): `Global`, `Function`, `FunctionExprName`, `Block`, `Catch`, `Class`, `Module`, `Type`. ESLint's module scope is collapsed into `Global`.
- `Scope.VariableScope()` is eslint-scope's `Scope#variableScope`: the nearest enclosing function / namespace / global scope, i.e. the `var` hoist target.
- Definition kinds (`scope.Def*`) map to eslint-scope's `Definition#type`, including the TypeScript ones (`DefType`, `DefEnumName`, `DefNamespaceName`, `DefTypeParameter`, `DefImport`).
- eslint-scope merges all same-name declarations in a scope into one `Variable` with several `defs`; here each declaration is its own `scope.Variable` and `Scope.Declarations(name)` is the merged view, in source order — so `Declarations(name)[0]` is `defs[0]`.
- `Scope.GlobalAugmentation` marks scopes inside `declare global { ... }`; bindings there are conceptually global.
- Concepts rslint does not expose (`parserOptions.globalReturn`, `ecmaVersion`-dependent function-in-block hoisting) are unmodeled.

`internal/rules/no_shadow` is the reference consumer.

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

// Source-generation questions live on Program, not in internal/utils
isLibFile := program.IsSourceFileDefaultLibrary(sourceFile)
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

## JavaScript Semantics (`ecmascript`, `minimatch3`, `isglob`)

A ported rule is handed values that were written for JavaScript: a string to trim, a number to print back out, a regexp or a glob that came from a rule option or from the source under lint. Go's standard library answers a _nearby but different_ question about every one of them, and the gap is observable — it is the difference between a rule that reports what ESLint reports and one that does not.

**Reach for these before writing your own scan, and before reaching for `strings`, `regexp`, or a general-purpose glob library.** If one of them is missing a reading you need, add it there rather than open-coding it in the rule.

### `internal/utils/ecmascript` — the language's own semantics

```go
import "github.com/web-infra-dev/rslint/internal/utils/ecmascript"
```

A helper standing in for something JavaScript exposes carries that name with its receiver in front (`StringTrim` for `String.prototype.trim`, `StringToUpperCase` for `String.prototype.toUpperCase`, `NumberToString` for `Number::toString`). Everything else is named the way ECMAScript names it — a grammar production (`IsWhiteSpace`, `IsLineTerminator`), the two of them together (`IsWhiteSpaceOrLineTerminator`) — or, where the language spells out no name, the question being asked (`IsBlank`).

```go
// String.prototype.trim(). NOT strings.TrimSpace, which disagrees in BOTH
// directions: it strips U+0085 NEL (not ECMAScript whitespace) and keeps
// U+FEFF (which ECMAScript trims).
trimmed := ecmascript.StringTrim(title)

// `s.trim() === ""`
blank := ecmascript.IsBlank(line)

// Everything `\s` matches and `trim()` strips: ECMAScript WhiteSpace plus
// LineTerminator. NOT unicode.IsSpace.
white := ecmascript.IsWhiteSpaceOrLineTerminator(r)

// The two productions on their own, for a rule that reads them apart — JSX
// text, say, where a line break and a space are not the same thing.
ecmascript.IsWhiteSpace(r)     // tab, vertical tab, form feed, U+FEFF, Zs
ecmascript.IsLineTerminator(r) // \n \r U+2028 U+2029; Go knows only the first two
ecmascript.LineTerminators     // the four, as a string

// String.prototype.toUpperCase / toLowerCase. NOT strings.ToUpper or
// unicode.ToUpper, which map one character to one character on Go's edition of
// Unicode: `ß` uppercases to `SS` here as it does in JavaScript, `ﬁ` to `FI`,
// and a capital sigma lowercases to a final sigma when it ends a word. An
// all-ASCII string never leaves the ASCII path, so this is the default even
// for a tag name or an option.
upper := ecmascript.StringToUpperCase(name)
lower := ecmascript.StringToLowerCase(name)

// The locale-sensitive pair, which answers the same way: a linter must draw
// the same diagnostics on a Turkish machine as anywhere else.
ecmascript.StringToLocaleUpperCase(s)
ecmascript.StringToLocaleLowerCase(s)

// String(n) / string concatenation. strconv picks the same digits but leaves
// fixed notation at a different point and spells the infinities differently.
text := ecmascript.NumberToString(42) // "42", not "4.2e+01"

// Is this string a complete, valid regexp literal (slashes and flags included)?
// Call it before offering a fix that emits `/.../`.
ok := ecmascript.IsValidRegexLiteral("/abc/g")

// Byte-level trivia scans. SkipTrailingWhitespace is the reverse counterpart
// tsgo's scanner.SkipTriviaEx does not provide.
ecmascript.SkipLeadingWhitespace(text, low, high)
ecmascript.SkipTrailingWhitespace(text, low, high)
ecmascript.ContainsLineTerminator(text, low, high)
ecmascript.IsTriviaWhitespaceByte(b) // ASCII fast path
ecmascript.IsTriviaWhitespaceRune(r) // call only for r >= U+0080
```

> `forbidigo` denies `strings.ToLower`, `strings.ToUpper` and `strings.TrimSpace` under `internal/rules/**` and `internal/plugins/**`. There is no exception to argue about: for a string that holds nothing but ASCII the port returns the same answer, and past ASCII it returns the one JavaScript returns.

### `internal/utils/ecmascript/regexp` — a JavaScript RegExp

```go
import esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
```

Use this for **any** pattern that was written as a JavaScript regexp — a rule option, or a `new RegExp(...)` in the source under lint. The standard library's `regexp` is RE2 and will not even compile a pattern using lookahead, lookbehind, or a backreference; a rule that shrugs off that compile error silently drops whatever option it came from.

```go
re, err := esregexp.Compile("^\\d+$", "iu") // source, flags
if err != nil {
    // ErrUnsupportedFlag / ErrUnsupportedSyntax, or a syntax error
}
if re.Test(candidate) {
    // ...
}
re.Source() // "^\\d+$"
re.Flags()  // "iu"
```

A `/i` comparison is its own reading, and the pattern widening in this package is built on it:

```go
// The Canonicalize abstract operation — how `/x/i` compares two characters.
// The second argument says whether the pattern carries `u` or `v`, because the
// two readings disagree: without one, an uppercase mapping that never crosses
// into ASCII, so `k` and U+212A KELVIN SIGN are two characters; with one,
// simple case folding, so they are one. Σ ς σ are one character either way.
esregexp.Canonicalize(r, unicodeMode)
esregexp.CaseEquivalents(r, unicodeMode)      // everything `/i` accepts for r, or nil
esregexp.CaseEquivalenceGroups(unicodeMode)   // for widening a whole range
```

`Test` bounds how long one match may run (`esregexp.MatchTimeout`, 1s) and answers `false` on overrun — a backtracking engine on a user-written pattern is a way to hang the linter. Use `TestOrError` when a rule needs to distinguish "no match" from "gave up".

Not covered: the `v` flag's set syntax (`Compile` refuses it), and case-insensitive backreference comparison.

> `depguard` denies `github.com/dlclark/regexp2` under `internal/rules/**` and `internal/plugins/**`. Go through this package.

### `internal/utils/unicode17` — the edition of Unicode Node reads

```go
import "github.com/web-infra-dev/rslint/internal/utils/unicode17"
```

Go 1.26's `unicode` tables are Unicode 15.0; Node 26 carries ICU 78, which is Unicode 17.0. Two bicameral scripts, eight new case mappings, nine and a half thousand letters and ninety-three combining marks arrived in between, and a rule that asks Go about them gets an answer ESLint would not give. This package holds that difference, and answers the category questions outright:

```go
unicode17.IsUpper(r)  // \p{Lu}, as unicode.IsUpper would if it were on 17.0
unicode17.IsLower(r)  // \p{Ll}
unicode17.IsLetter(r) // \p{L}
unicode17.IsMark(r)   // \p{M} — the combining marks
```

Reach for it when the upstream rule reads a **character class**: change-case's `\p{Ll}` behind `unicorn/filename-case`, `\p{Uppercase_Letter}` behind `unicorn/prefer-array-flat`, `\p{M}` behind `no-misleading-character-class`. When upstream instead writes `c === c.toUpperCase()`, that is a case mapping and belongs to `ecmascript`, which reads this package on its own.

The package deletes itself: `TestDeltaStillNeeded` fails the moment the toolchain stops being Unicode 15.0, which on Go 1.27 is the signal to delete the package rather than maintain it.

> `depguard` denies the standard library's `unicode` under `internal/rules/**` and `internal/plugins/**`. A case question goes to `ecmascript`, a category question here, and an identifier question — is this character allowed to start or continue an identifier? — to tsgo's `scanner.IsIdentifierStart` / `scanner.IsIdentifierPart`, which reads TypeScript's own tables so that a rule and the parser never disagree.

### `internal/utils/minimatch3` — globs the way a plugin reads them

```go
import "github.com/web-infra-dev/rslint/internal/utils/minimatch3"
```

A port of minimatch 3.1.5 and the brace-expansion it pulls in, including the extended glob syntax (`!(a)`, `@(a|b)`, `+(a)`, `?(a)`, `*(a)`) that general-purpose Go glob libraries leave out.

```go
// One-shot
ok := minimatch3.Match("src/**/*.ts", "src/a/b.ts", minimatch3.Options{Dot: true})

// Compiled, when matching many paths against one pattern
m := minimatch3.New(pattern, minimatch3.Options{})
ok = m.Match(path)

// Brace expansion on its own
variants := minimatch3.BraceExpand("a{b,c}d") // ["abd", "acd"]
```

`Options` mirrors minimatch's own: `Dot`, `NoBrace`, `NoGlobStar`, `NoExt`, `NoCase`, `NoNegate`, `NoComment`.

> `depguard` denies `github.com/bmatcuk/doublestar` under `internal/rules/**` and `internal/plugins/**`. Go through this package.

### `internal/utils/isglob` — is this even a glob?

```go
import "github.com/web-infra-dev/rslint/internal/utils/isglob"
```

Ports is-glob 4.0.3 and the is-extglob 2.1.1 it depends on. A plugin asks this to decide _how to read a rule option_: a value that is a glob gets matched, a value that is not gets compared as a literal. Answering differently from upstream moves the line between the two.

```go
isglob.Is("src/*.ts")   // true
isglob.Is("src/a.ts")   // false — a plain path
isglob.IsExtglob("@(a|b)") // true, the extended-list question on its own
```

### Which package for which upstream dependency

Open the upstream rule's imports and its plugin's `package.json`, then match:

| Upstream reads the pattern with       | Use                                                         |
| ------------------------------------- | ----------------------------------------------------------- |
| a regexp literal or `new RegExp(...)` | `utils/ecmascript/regexp` (import as `esregexp`)            |
| `minimatch` at `^3.x`                 | `utils/minimatch3`                                          |
| `is-glob`                             | `utils/isglob`                                              |
| any other glob package                | **not supported — stop and report to the user (see below)** |

The stdlib `regexp` is not banned outright. A pattern written in this repository that RE2 and JavaScript read the same way, and that no user input reaches, can stay on it. Anything a user can influence — a rule option, a config file, the source under lint — takes `esregexp`, however plain the pattern looks: RE2 refuses syntax JavaScript accepts, and the caller usually swallows the compile error and reports nothing.

### ⚠️ Only minimatch 3 and is-glob are ported

**If the rule you are porting reads globs with anything else, stop and report it to the user. Do not substitute `minimatch3` or `doublestar` and carry on, and do not port the package yourself.**

`minimatch@10` is the one to expect. ESLint moved to it for the paths it matches on its own behalf (flat config `files` / `ignores`), and a plugin may follow. Only the 3.x reading is ported, because that is what the plugin ecosystem pins — `eslint-plugin-import` and `eslint-plugin-react` both depend on `minimatch@^3.1.2`.

The substitutions are not close enough to make quietly:

- **`minimatch3`** differs from 10 on POSIX character classes: `a[[:alpha:]]b` matches `aXb` under 10, and does not under 3.
- **`doublestar`** differs from 10 on 13 of 37 sampled patterns. Six are extended glob syntax, which it does not implement at all. The other seven are not exotic: `src/**` matches `src` itself under doublestar but not under minimatch, POSIX classes and `{1..3}` ranges are unsupported, a leading `!` is a literal rather than a negation, and the empty path and `a//b` are handled differently.

Report which upstream package and version the rule depends on, and which of its patterns would be read differently. The user decides whether to port it, accept a documented divergence, or skip the rule.

---

## See Also

- [PORT_RULE.md](./PORT_RULE.md) - Main rule porting workflow
- [AST_PATTERNS.md](./AST_PATTERNS.md) - AST traversal patterns and examples
- [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - Commands and checklist
