# Rslint AST Patterns Guide

This document provides patterns and examples for working with the AST (Abstract Syntax Tree) when implementing lint rules.

> **Note**: This is a reference document for [PORT_RULE.md](./PORT_RULE.md). See that document for the complete rule porting workflow.

---

## Message Builder Pattern

Define message builders as functions for reusability and clarity:

```go
// Define all messages at package level
func messageNoDebugger() rule.RuleMessage {
    return rule.RuleMessage{
        Id:          "no-debugger",
        Description: "Unexpected 'debugger' statement.",
    }
}

func messageWithData(name string) rule.RuleMessage {
    return rule.RuleMessage{
        Id:          "unexpected",
        Description: fmt.Sprintf("Unexpected use of '%s'.", name),
    }
}

// Use in listener
ctx.ReportNode(node, messageNoDebugger())
```

---

## Listener Types

Besides basic `ast.Kind*` listeners, there are special listener types defined in `internal/rule/rule.go`:

### Basic Listener

Triggers when **entering** a node:

```go
rule.RuleListeners{
    ast.KindFunctionDeclaration: func(node *ast.Node) {
        // Executes when entering a function declaration node
    },
}
```

### Exit Listener

Triggers when **exiting** a node (after all children have been processed):

```go
rule.RuleListeners{
    rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
        // Executes when exiting a function declaration node
        // Useful when you need to collect child node info before reporting
    },
}
```

**Use Case**: When you need to traverse all child nodes first, then make a decision based on collected information.

### Pattern Listeners

For allow/deny pattern matching. **Note**: These require a `kind` argument:

```go
rule.RuleListeners{
    // Triggers when allow pattern matches for the specified node kind
    rule.ListenerOnAllowPattern(ast.KindCallExpression): func(node *ast.Node) {
        // Handle allowed pattern
    },

    // Triggers when allow pattern does NOT match for the specified node kind
    rule.ListenerOnNotAllowPattern(ast.KindCallExpression): func(node *ast.Node) {
        // Handle non-allowed pattern
    },
}
```

**Use Case**: When the rule involves allow/deny list matching.

---

## AST Shape Essentials

The tsgo AST differs from ESTree (ESLint's AST) in a few systematic ways. Most porting bugs trace back to one of these shape differences. Work through each section before and after implementing a rule.

### ParenthesizedExpression

tsgo keeps parentheses as an explicit `KindParenthesizedExpression` node; ESTree drops them during parsing. Any time a rule reads a child expression, parentheses may be sitting in between.

**Primary helpers** (from `shim/ast`, prefer these over hand-rolled loops):

- `ast.SkipParentheses(node)` — returns the innermost non-paren expression.
- `ast.WalkUpParenthesizedExpressions(node)` — returns the first non-paren ancestor.

```go
inner := ast.SkipParentheses(node.AsCallExpression().Expression)
// `inner` is the callee without any `( … )` wrapping
```

**Trap sites** — any expression-typed child can be parenthesised. The table below lists high-frequency offenders. The principle is universal: if you are about to read an expression-typed child and do anything with its kind/text/structure, unwrap it first.

| Kind                                                                                                                                                                 | Children to unwrap                        |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------- |
| `BinaryExpression`                                                                                                                                                   | `Left`, `Right`                           |
| `PrefixUnaryExpression` / `PostfixUnaryExpression`                                                                                                                   | `Operand`                                 |
| `CallExpression` / `NewExpression`                                                                                                                                   | `Expression` (callee)                     |
| `ElementAccessExpression`                                                                                                                                            | `Expression`, `ArgumentExpression`        |
| `PropertyAccessExpression`                                                                                                                                           | `Expression` (object)                     |
| `TemplateSpan`                                                                                                                                                       | `Expression`                              |
| `SpreadElement` / `AwaitExpression` / `YieldExpression` / `TypeOfExpression` / `VoidExpression` / `DeleteExpression`                                                 | `Expression`                              |
| `ConditionalExpression`                                                                                                                                              | `Condition`, `WhenTrue`, `WhenFalse`      |
| `VariableDeclaration` / `PropertyAssignment` / `BindingElement`                                                                                                      | `Initializer`                             |
| `IfStatement` / `WhileStatement` / `DoStatement` / `SwitchStatement` / `CaseClause` / `WithStatement` / `ReturnStatement` / `ThrowStatement` / `ExpressionStatement` | `Expression`                              |
| `ForStatement`                                                                                                                                                       | `Initializer`, `Condition`, `Incrementor` |
| `ForInStatement` / `ForOfStatement`                                                                                                                                  | `Initializer`, `Expression`               |

A helper that embeds the check (e.g. `isNumeric`, `isStringType`) should call `ast.SkipParentheses` at the top rather than require every call site to unwrap — otherwise one forgotten caller is a silent divergence.

### Optional Chain

tsgo does not have a `ChainExpression` wrapper. Instead, every link in an optional chain (`PropertyAccess` / `ElementAccess` / `Call` / `NonNullExpression`) carries `NodeFlagsOptionalChain`. Parentheses break the chain — `(foo?.bar)(x)` parses as an ordinary call whose callee happens to be a paren-wrapped optional chain.

- `ast.IsOptionalChain(node)` — true iff node is a link in an optional chain.
- `ast.IsOptionalChainRoot(node)` — true iff node introduces the chain (holds the leading `?.`).
- `ast.IsOutermostOptionalChain(node)` — true iff node is the top of its chain.

When porting a rule that switches on "is the argument a `ChainExpression`?", translate to `ast.IsOptionalChain(arg)` on the tsgo-side argument after `ast.SkipParentheses`.

### Literal Kinds

ESTree uses one `Literal` node carrying a typed `value`. tsgo splits the literal family across kinds, and booleans / null are keyword tokens rather than literal nodes:

| Value                   | tsgo kind                              | Text accessor                                                                                                                                                                                                                                                                                                          |
| ----------------------- | -------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Number                  | `KindNumericLiteral`                   | `node.AsNumericLiteral().Text` is **decimal-normalized at parse time** (`0x1`, `1e2`, `1.0` are all stored as their canonical decimal string — `"1"`, `"100"`, `"1"`). If you need the raw source form (e.g. for ESLint token-level parity), use `scanner.GetSourceTextOfNodeFromSourceFile(sf, node, false)` instead. |
| String                  | `KindStringLiteral`                    | `node.AsStringLiteral().Text` (cooked)                                                                                                                                                                                                                                                                                 |
| BigInt                  | `KindBigIntLiteral`                    | `node.AsBigIntLiteral().Text` retains the `n` suffix. Same caveat as NumericLiteral regarding base normalization — use `scanner.GetSourceTextOfNodeFromSourceFile` for raw form.                                                                                                                                       |
| Regex                   | `KindRegularExpressionLiteral`         | `node.Text()`                                                                                                                                                                                                                                                                                                          |
| `true` / `false`        | `KindTrueKeyword` / `KindFalseKeyword` | —                                                                                                                                                                                                                                                                                                                      |
| `null`                  | `KindNullKeyword`                      | —                                                                                                                                                                                                                                                                                                                      |
| `` `…` `` without `${}` | `KindNoSubstitutionTemplateLiteral`    | `node.AsNoSubstitutionTemplateLiteral().Text` (cooked)                                                                                                                                                                                                                                                                 |
| `` `…${x}…` ``          | `KindTemplateExpression`               | `Head.Text()` + each `TemplateSpan.Literal.Text()` (all cooked)                                                                                                                                                                                                                                                        |

Common translations:

- ESTree `node.type === "Literal" && typeof node.value === "number"` → `node.Kind == ast.KindNumericLiteral`.
- ESTree's `isStringLiteral` helper (StringLiteral + TemplateLiteral) → `ast.IsStringLiteralLike(node)` covers `StringLiteral` + `NoSubstitutionTemplateLiteral`. For TemplateExpression add it explicitly.
- Comparing numeric literal value (e.g. "is this `1`"): since `.Text` is already normalized, `node.AsNumericLiteral().Text == "1"` works — or call `utils.NormalizeNumericLiteral(text)` for an explicit, intent-revealing comparison.
- Distinguishing source forms (`0x1` vs `1` — rare, only when porting an ESLint rule that compares tokens by raw value): use `scanner.GetSourceTextOfNodeFromSourceFile(sf, node, false)`; `.Text` cannot be used for this.

### Binary Operator Kinds

tsgo uses `BinaryExpression` for the entire family of binary operators, including assignment and comma:

- `a + b`, `a * b`, `a && b`, `a ?? b`, … — plain arithmetic / logical
- `a, b` — sequence (ESTree `SequenceExpression`) via `OperatorToken.Kind == ast.KindCommaToken`
- `a = b`, `a += b`, `a **= b`, … — assignment (ESTree `AssignmentExpression`) via `ast.KindEqualsToken`, `ast.KindPlusEqualsToken`, etc.

For a rule that registers separate ESTree listeners for `AssignmentExpression` / `SequenceExpression`, collapse into one `BinaryExpression` listener and switch on `OperatorToken.Kind`. Do not rely on `IsBinaryExpression` alone to exclude assignments.

### Node Text and Positions

Raw `node.Pos()` and `node.End()` include leading trivia (whitespace, comments, line breaks). This is almost never what a rule wants — reading source text across `node.Pos()..node.End()` yields leading blanks, and reporting at `node.Pos()` positions the diagnostic on the trivia.

Prefer:

- `utils.TrimNodeTextRange(sourceFile, node)` — range with leading trivia skipped.
- `utils.TrimmedNodeText(sourceFile, node)` — source text over that trimmed range.
- `ctx.ReportNode(node, msg)` — diagnostic range already uses the trimmed span; no manual adjustment needed.

For fixes, construct these helpers inside the deferred report builder:

- `rule.RuleFixReplace(sf, node, text)` — replaces the trimmed span.
- `rule.RuleFixReplaceRange(range, text)` — replaces a specific range (useful for precision edits across multiple nodes).

### Object Literal Member Kinds

ESTree collapses several distinct object-literal members into a single `Property` node with `method` / `shorthand` / `computed` flags. tsgo uses **five different kinds**, and a rule that filters only on `KindPropertyAssignment` will silently miss the other four:

| Source                               | tsgo kind                             | ESTree equivalent                            |
| ------------------------------------ | ------------------------------------- | -------------------------------------------- |
| `{ a: b }`                           | `KindPropertyAssignment`              | `Property { kind: "init" }`                  |
| `{ a() {} }` (method shorthand)      | `KindMethodDeclaration`               | `Property { kind: "init", method: true }`    |
| `{ a }` (shorthand property)         | `KindShorthandPropertyAssignment`     | `Property { kind: "init", shorthand: true }` |
| `{ get a() {} }` / `{ set a(v) {} }` | `KindGetAccessor` / `KindSetAccessor` | `Property { kind: "get"/"set" }`             |
| `{ ...x }` (spread)                  | `KindSpreadAssignment`                | `SpreadElement`                              |

When porting a rule that uses `p.type === "Property" && p.kind === "init"` to collect "init-shape" members (e.g. `Object.defineProperty` descriptor detection), include all three of `KindPropertyAssignment`, `KindMethodDeclaration`, and `KindShorthandPropertyAssignment` — otherwise method-shorthand and shorthand-property forms silently go missing.

### ComputedPropertyName as a Wrapper

ESTree marks computed keys with `Property.computed = true` and puts the inner expression directly in `Property.key`. tsgo wraps the inner expression in a separate `KindComputedPropertyName` node:

```go
// { [x + 1]: ... }  or  class A { get [x + 1]() {} }
nameNode := node.Name()                          // → KindComputedPropertyName
if nameNode.Kind == ast.KindComputedPropertyName {
    expr := nameNode.AsComputedPropertyName().Expression  // the actual key expression
    expr = ast.SkipParentheses(expr)             // and don't forget parens
}
```

`utils.GetStaticPropertyName` already handles the wrapper and returns the static name for `[ 'a' ]` / `[ 1e2 ]` / `[ true ]` / `[ null ]` / `[ /re/ ]` / ``[`a`]`` forms — prefer it over hand-unwrapping.

---

## AST Traversal Patterns

### 1. ForEachChild - Iterate Direct Children

```go
node.ForEachChild(func(child *ast.Node) bool {
    if child.Kind == ast.KindIdentifier {
        id := child.AsIdentifier()
        fmt.Println(id.Text())
    }
    return false // false = continue, true = stop
})
```

### 2. Parent Chain - Find Specific Parent

```go
func findParentOfKind(node *ast.Node, kind ast.Kind) *ast.Node {
    current := node.Parent
    for current != nil {
        if current.Kind == kind {
            return current
        }
        current = current.Parent
    }
    return nil
}

// Example: Find the containing function
func findContainingFunction(node *ast.Node) *ast.Node {
    current := node.Parent
    for current != nil {
        switch current.Kind {
        case ast.KindFunctionDeclaration,
             ast.KindFunctionExpression,
             ast.KindArrowFunction,
             ast.KindMethodDeclaration:
            return current
        }
        current = current.Parent
    }
    return nil
}
```

### 3. NodeList Iteration

```go
// Iterate over function arguments
callExpr := node.AsCallExpression()
if args := callExpr.Arguments; args != nil {
    for _, arg := range args.Nodes {
        // Process each argument
        processArgument(arg)
    }
}

// Iterate over statement block
block := node.AsBlock()
for _, stmt := range block.Statements {
    // Process each statement
}
```

### 4. Type Casting and Checking

```go
// Safe type casting pattern
if node.Kind == ast.KindCallExpression {
    callExpr := node.AsCallExpression()
    if callExpr != nil {
        // Check if it's a specific function call
        if callExpr.Expression.Kind == ast.KindIdentifier {
            id := callExpr.Expression.AsIdentifier()
            if id.Text() == "eval" {
                // Handle eval call
            }
        }
    }
}

// Handle PropertyAccessExpression (e.g., console.log)
if node.Kind == ast.KindPropertyAccessExpression {
    propAccess := node.AsPropertyAccessExpression()
    objectName := ""
    propertyName := propAccess.Name.Text()

    if propAccess.Expression.Kind == ast.KindIdentifier {
        objectName = propAccess.Expression.AsIdentifier().Text
    }
    // objectName = "console", propertyName = "log"
}
```

---

## Using TypeChecker

For rules that need type information, access `TypeChecker` via `RuleContext`.

### RequiresTypeInfo vs nil check

How you handle TypeChecker availability depends on the rule type:

| Rule type                           | Approach                                          | Reason                                                                                                                             |
| ----------------------------------- | ------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| **@typescript-eslint** (type-aware) | Set `RequiresTypeInfo: true` in the `Rule` struct | The linter automatically skips the rule on files without a type checker. `ctx.TypeChecker` is guaranteed non-nil inside listeners. |
| **Core ESLint**                     | Check `ctx.TypeChecker == nil` at the call site   | Core rules must run on JS files (no type checker). They either degrade gracefully or return early. Do NOT set `RequiresTypeInfo`.  |

**Type-aware @typescript-eslint rule** — set `RequiresTypeInfo: true`, no nil check needed:

```go
var MyTSRule = rule.CreateRule(rule.Rule{
    Name:             "my-ts-rule",
    RequiresTypeInfo: true,
    Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
        return rule.RuleListeners{
            ast.KindCallExpression: func(node *ast.Node) {
                // ctx.TypeChecker is guaranteed non-nil
                exprType := ctx.TypeChecker.GetTypeAtLocation(node.AsCallExpression().Expression)
                // ...
            },
        }
    },
})
```

**Core ESLint rule** — nil check required, do NOT set `RequiresTypeInfo`:

```go
var MyCoreRule = rule.Rule{
    Name: "my-core-rule",
    Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
        return rule.RuleListeners{
            ast.KindCallExpression: func(node *ast.Node) {
                if ctx.TypeChecker == nil {
                    return // TypeChecker not available — degrade gracefully
                }

                exprType := ctx.TypeChecker.GetTypeAtLocation(node.AsCallExpression().Expression)
                // ...
            },
        }
    },
}
```

> **Why the distinction?** The linter uses `RequiresTypeInfo` to filter rules via `FilterNonTypeAwareRules` (see `internal/linter/linter.go`). Setting it on a core ESLint rule would prevent it from running on JS files entirely, which breaks `eslint:recommended` behavior.

### Common TypeChecker Methods

See `typescript-go/_packages/api/src/api.ts` for full API:

| Method                            | Description             |
| --------------------------------- | ----------------------- |
| `GetTypeAtLocation(node)`         | Get the type of a node  |
| `TypeToString(type)`              | Convert type to string  |
| `GetSignaturesOfType(type, kind)` | Get function signatures |

---

## Resolving Identifiers and Collecting References (ctx.Refs)

Reference: `internal/rule/ref_store.go`, consumer examples: `internal/rules/no_var/no_var.go` (References), `internal/rules/prefer_const/prefer_const.go` (Resolve).

`ctx.Refs` is a lazily built per-file identifier-reference index — rslint's stand-in for ESLint's scope manager (`variable.references`, `getScope().references`). It has two methods:

| Method                        | Direction                                                                                                                  | Touches TypeChecker?                                                            |
| ----------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `Resolve(node) *ast.Symbol`   | identifier → its declaring symbol, anywhere the checker can see (this file, cross-file, `.d.ts`, standard-library globals) | Only as a fallback, when the binder scope walk alone can't place the identifier |
| `References(sym) []*ast.Node` | symbol → every identifier in this file that references it                                                                  | Same fallback trigger as `Resolve`                                              |

Both resolve identifiers with the binder's scope walk first — the same scope
walk the checker performs, but without the checker itself — so the common
case (same-file locals, which is the overwhelming majority of what rules
query) never touches the TypeChecker and never triggers lazy type
computation. When the binder can't place an identifier — a symbol declared
outside this file (cross-file, `.d.ts`, standard-library globals) — and a
TypeChecker was supplied, `Resolve` falls back to it automatically, at the
cost of a real round-trip for that identifier; `References` picks up the same
fallback when queried with a symbol `Resolve` obtained that way. Without a
TypeChecker, that fallback is a no-op and both methods only ever see symbols
declared in this file.

Do **NOT** hand-roll any of this by walking the AST and calling
`ctx.TypeChecker.GetSymbolAtLocation` on every identifier, and do **NOT**
hand-roll a "try `ctx.Refs`, fall back to the checker" wrapper of your own —
`Resolve` already is that wrapper. Building your own copy repeats the
checker round-trip logic for no benefit.

To check whether an identifier is locally shadowed — resolved to a symbol
declared in this file, as opposed to a global, ambient, or cross-file one —
pair `Resolve` with `utils.IsSymbolDeclaredInFile(symbol, ctx.SourceFile)`:

```go
symbol := ctx.Refs.Resolve(node)
isLocal := symbol != nil && utils.IsSymbolDeclaredInFile(symbol, ctx.SourceFile)
```

When the question is specifically "does a local declaration shadow this
built-in _value_", use `utils.IsValueSymbolDeclaredInFile` instead: it counts
only the declarations that bind a value name in this file. Any
`namespace`/`module` declaration counts, whatever it happens to contain.
Interfaces and type aliases don't, and neither does anything inside a
`declare global` block or a module augmentation — those merge into the
ambient global's symbol (`interface Map {}` in a global script attaches to
the same symbol as lib.d.ts's `Map`) while the call site keeps reaching the
global value. See [UTILS_REFERENCE.md](./UTILS_REFERENCE.md#internalruleref_storego---reference-index-ctxrefs).

### Collecting references to a symbol

For **"every identifier that references this declared symbol"** (ESLint's
`variable.references`):

```go
refs := ctx.Refs.References(sym) // []*ast.Node, in source order
```

`sym` is usually a local declaration symbol (see below), but it can also be a
symbol obtained through `Resolve`'s checker fallback — `References` falls
back to the checker internally too, so it finds in-file identifiers
referencing an external symbol (e.g. every unshadowed use of `RegExp` in the
file) just as well as it finds references to a local variable.

### Semantics

- **Query key is a binder symbol.** For a same-file declaration, get it from the declaration node: `decl.Symbol()` (the binder attaches symbols to `VariableDeclaration` / `BindingElement` / etc., not to the name identifier). Do NOT query with `checker.GetSymbolAtLocation` results for a symbol you expect to be local — the checker may return merged symbols that compare unequal to binder symbols, and the lookup will miss. (A symbol obtained from `Resolve`'s checker fallback is the one exception: it's already the right key for `References`.)
- **Declaration names are not references**: the `a` of `var a = 1` is excluded; the `a` of a later `a = 2` is included. Reads and writes are both references — distinguish them positionally in your rule if needed.
- **Non-reference positions are pre-filtered**: property names (`x.a`), object-literal/destructuring keys, import/export binding names, labels, and lowercase JSX tag names never appear.
- **Scope- and meaning-aware**: shadowing identifiers in other scopes don't leak in, and type-position vs value-position identifiers resolve to the right symbol (mirrors the checker's meaning selection).
- **Always single-file search scope**: `References` only ever looks at identifiers _in this file_, regardless of where the queried symbol is declared. It answers "who references this symbol in this file," not "who references it anywhere in the program" — finding references in _other_ files still requires the checker's own whole-program search.
- Returned slice is **read-only**. `References(nil)` returns nil; a nil `*RefStore` receiver is also safe for both methods.

### Availability

`ctx.Refs` is nil when no program is available. Core ESLint rules must guard:

```go
if ctx.Refs == nil {
    return false // degrade gracefully, same as a ctx.TypeChecker nil check
}
```

For `RequiresTypeInfo: true` rules, a program always exists, so `ctx.Refs` is guaranteed non-nil. `Resolve`'s checker fallback additionally needs `ctx.TypeChecker` non-nil; when it's nil, `Resolve` stays binder-only.

### Example (from no-var)

```go
// Declaration side: collect binder symbols from the declaration nodes.
utils.CollectBindingNames(varDecl.Name(), func(ident *ast.Node, _ string) {
    if decl := ident.Parent; decl != nil && decl.Name() == ident {
        if sym := decl.Symbol(); sym != nil {
            vars = append(vars, varInfo{nameNode: ident, sym: sym})
        }
    }
})

// Reference side: one map lookup per symbol, no AST walking.
refs := make(map[*ast.Symbol][]*ast.Node, len(vars))
for _, v := range vars {
    refs[v.sym] = ctx.Refs.References(v.sym)
}
```

### Cost model

The index is lazy twice over: the single AST walk that buckets candidate identifiers runs on the first `References` call in the file, and name resolution runs once per **queried name** — asking about `foo` never pays for resolving unrelated identifiers like `console` or `Promise`. Files where no rule queries the index pay nothing. `Resolve` and `References` never touch the TypeChecker for a same-file symbol; the checker fallback only costs a round-trip for identifiers the binder can't place, which for most rules is a small minority.

---

## Complex Rule Patterns

### Control Flow Analysis

Reference: `internal/rules/constructor_super/constructor_super.go`

For rules that need to track state across multiple code paths (if/else, switch, try/catch):

```go
// Track super() calls across code paths
type CodePath struct {
    hasSuper    bool
    isReachable bool
}

func analyzeCodePaths(node *ast.Node) []CodePath {
    // Analyze if/else, switch, try/catch branches
    // Ensure all paths call super()
}
```

### Function Return Value Analysis

Reference: `internal/rules/array_callback_return/array_callback_return.go`

For rules that check if functions return values:

```go
// Check if a callback function has a return value
func checkCallbackReturn(funcNode *ast.Node) bool {
    hasReturn := false
    funcNode.ForEachChild(func(child *ast.Node) bool {
        if child.Kind == ast.KindReturnStatement {
            ret := child.AsReturnStatement()
            if ret.Expression != nil {
                hasReturn = true
                return true // Stop traversal
            }
        }
        return false // Continue traversal
    })
    return hasReturn
}
```

### Scope Stack Pattern

Reference: `internal/rules/no_extra_bind/no_extra_bind.go`

For rules that track state across nested scopes (e.g., `this` usage, variable declarations), use enter/exit listeners with a linked-list stack:

```go
type scopeInfo struct {
    // Per-scope state
    thisFound bool
    upper     *scopeInfo // Link to parent scope
}

var scope *scopeInfo

enterScope := func(node *ast.Node) {
    scope = &scopeInfo{upper: scope}
}

exitScope := func(node *ast.Node) {
    if scope != nil {
        // Check scope state before popping
        scope = scope.upper
    }
}

return rule.RuleListeners{
    ast.KindFunctionExpression:                       enterScope,
    rule.ListenerOnExit(ast.KindFunctionExpression):  exitScope,
    ast.KindFunctionDeclaration:                      enterScope,
    rule.ListenerOnExit(ast.KindFunctionDeclaration): exitScope,
    // Arrow functions do NOT create a new `this` scope — handle separately
}
```

**Key considerations**:

- Different node kinds may create different scope types (e.g., function vs method vs arrow)
- Arrow functions inherit `this` from the enclosing scope — typically should NOT push a new scope
- Class methods, getters, setters, and constructors create their own `this` scope

### Type Annotation Checking

Reference: `internal/plugins/typescript/rules/no_restricted_types/no_restricted_types.go`

For rules that check type annotations:

```go
// Check if a type annotation is any
func isAnyType(node *ast.Node) bool {
    if node.Kind == ast.KindAnyKeyword {
        return true
    }
    // Handle type references: type Foo = any
    if node.Kind == ast.KindTypeReference {
        typeRef := node.AsTypeReferenceNode()
        if typeRef.TypeName.Kind == ast.KindIdentifier {
            return typeRef.TypeName.AsIdentifier().Text == "any"
        }
    }
    return false
}
```

---

## Reporting Functions

`RuleContext` provides reporting methods in `internal/rule/context.go`.

### Basic Report

```go
ctx.ReportNode(node, rule.RuleMessage{
    Id:          "messageId",
    Description: "Error description",
})
```

### Report with Text Range

```go
ctx.ReportRange(core.TextRange{Pos: start, End: end}, rule.RuleMessage{...})
```

### Reports with Optional Edits

Autofixes and suggestions are optional artifacts. New native rules must defer
their construction so diagnostics-only consumers and suppressed diagnostics do
not pay for replacement text, token scans, edit ranges/slices, or suggestion
messages.

| Artifact                     | Node-keyed method                           | Range-keyed method                           |
| ---------------------------- | ------------------------------------------- | -------------------------------------------- |
| Autofixes                    | `ReportNodeWithDeferredFixes`               | `ReportRangeWithDeferredFixes`               |
| Suggestions                  | `ReportNodeWithDeferredSuggestions`         | `ReportRangeWithDeferredSuggestions`         |
| Both, independently demanded | `ReportNodeWithDeferredFixesAndSuggestions` | `ReportRangeWithDeferredFixesAndSuggestions` |

The diagnostic decision, message, and report range stay outside the builder and
must not depend on edit demand. Work needed only to decide whether an edit is
safe may be deferred too; return nil when the diagnostic has no artifact. Each
requested builder runs synchronously during the report call, after suppression,
and is never retained. It must not mutate detection state.

#### Autofix

```go
ctx.ReportNodeWithDeferredFixes(node, msg, func() []rule.RuleFix {
    if !canFix(node) {
        return nil
    }
    replacement := buildReplacement(ctx.SourceFile, node)
    return []rule.RuleFix{
        rule.RuleFixReplace(ctx.SourceFile, node, replacement),
    }
})
```

#### Suggestion

```go
ctx.ReportNodeWithDeferredSuggestions(node, msg, func() []rule.RuleSuggestion {
    replacement := buildReplacement(ctx.SourceFile, node)
    return []rule.RuleSuggestion{{
        Message: rule.RuleMessage{
            Id:          "suggestReplace",
            Description: "Replace this expression",
        },
        FixesArr: []rule.RuleFix{
            rule.RuleFixReplace(ctx.SourceFile, node, replacement),
        },
    }}
})
```

#### Autofix and Suggestion

When one diagnostic exposes both categories, use separate builders so the
consumer materializes only what it requested:

```go
ctx.ReportNodeWithDeferredFixesAndSuggestions(
    node,
    msg,
    func() []rule.RuleFix {
        return buildFixes(ctx.SourceFile, node)
    },
    func() []rule.RuleSuggestion {
        return buildSuggestions(ctx.SourceFile, node)
    },
)
```

---

## Fix Helper Functions

Defined in `internal/rule/diagnostic.go`:

```go
// Insert text before node (requires SourceFile)
rule.RuleFixInsertBefore(ctx.SourceFile, node, "text to insert")

// Insert text after node
rule.RuleFixInsertAfter(node, "text to insert")

// Replace node text (requires SourceFile)
rule.RuleFixReplace(ctx.SourceFile, node, "replacement text")

// Replace a specific range
rule.RuleFixReplaceRange(textRange, "replacement text")

// Remove node (requires SourceFile)
rule.RuleFixRemove(ctx.SourceFile, node)

// Remove a specific range
rule.RuleFixRemoveRange(textRange)
```

### Multi-Range Fixes

To remove or replace multiple non-contiguous ranges in a single fix, pass multiple `RuleFix` values:

```go
// Example: remove ".bind" and "(arg)" separately from "fn.bind(arg)"
ctx.ReportNodeWithDeferredFixes(node, msg, func() []rule.RuleFix {
    return []rule.RuleFix{
        rule.RuleFixRemoveRange(core.NewTextRange(dotStart, bindEnd)),    // removes ".bind"
        rule.RuleFixRemoveRange(core.NewTextRange(parenStart, parenEnd)), // removes "(arg)"
    }
})
```

---

## Token Scanning

When implementing auto-fixes that need to locate specific tokens (parentheses, brackets, operators), use the scanner utilities from `shim/scanner/` instead of manual character iteration.

### SkipTrivia — Quick Token Position Lookup

When you just need to find the start position of the next token (skipping whitespace and comments), use `scanner.SkipTrivia`:

```go
import "github.com/microsoft/typescript-go/shim/scanner"

// Skip whitespace, line/block comments, BOM, shebang, and conflict markers
sourceText := ctx.SourceFile.Text()
nextTokenPos := scanner.SkipTrivia(sourceText, startPos)
```

This is simpler and more efficient than creating a full scanner when you only need a position.

### Full Scanner — Token-by-Token Scanning

```go
import "github.com/microsoft/typescript-go/shim/scanner"

// Create scanner starting from a position
s := scanner.GetScannerForSourceFile(ctx.SourceFile, startPos)

// Scan tokens until reaching end position
for s.TokenStart() < endPos {
    switch s.Token() {
    case ast.KindOpenParenToken:
        openPos := s.TokenEnd()  // Position after '('
    case ast.KindCloseParenToken:
        closePos := s.TokenStart()  // Position before ')'
    }
    s.Scan()  // Move to next token
}
```

### Scanner Methods

| Method            | Description                                         |
| ----------------- | --------------------------------------------------- |
| `s.Token()`       | Current token kind (e.g., `ast.KindOpenParenToken`) |
| `s.TokenStart()`  | Start position of current token                     |
| `s.TokenEnd()`    | End position of current token                       |
| `s.TokenRange()`  | `core.TextRange` of current token                   |
| `s.Scan()`        | Advance to next token                               |
| `s.ResetPos(pos)` | Reset scanner to a new position                     |

### Why Use Scanner API

- **Handles whitespace and comments automatically** - No need to manually skip them
- **Consistent boundary handling** - Token positions are always accurate
- **Less error-prone** - Avoids off-by-one errors in manual scanning

---

## Nested Structure Handling

When extracting text within delimiters (parentheses, brackets, braces), track depth to handle nested structures correctly.

### Problem

```go
// Bug: Each '(' overwrites openParenEnd
for s.TokenStart() < node.End() {
    if s.Token() == ast.KindOpenParenToken {
        openParenEnd = s.TokenEnd()  // Wrong for Array((x), y)
    }
    // ...
}
```

For `Array((x), y)`, this incorrectly captures the inner `(` position, resulting in `[x)` instead of `[(x), y]`.

### Solution: Track Depth

```go
openParenEnd := -1
closeParenStart := -1
parenDepth := 0

for s.TokenStart() < node.End() {
    if s.Token() == ast.KindOpenParenToken {
        if openParenEnd == -1 {
            // Only capture the first open paren
            openParenEnd = s.TokenEnd()
        }
        parenDepth++
    } else if s.Token() == ast.KindCloseParenToken {
        parenDepth--
        if parenDepth == 0 {
            // This matches the first open paren
            closeParenStart = s.TokenStart()
            break
        }
    }
    s.Scan()
}

// Extract text between matched delimiters
text := ctx.SourceFile.Text()[openParenEnd:closeParenStart]
```

### Applies To

- Parentheses: `(`, `)`
- Brackets: `[`, `]`
- Braces: `{`, `}`
- Angle brackets (generics): `<`, `>`

---

## See Also

- [PORT_RULE.md](./PORT_RULE.md) - Main rule porting workflow
- [UTILS_REFERENCE.md](./UTILS_REFERENCE.md) - Utility functions reference
- [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - Commands and checklist
