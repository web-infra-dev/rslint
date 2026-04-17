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

| Kind                                                                                                                                                                                              | Children to unwrap                   |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ |
| `BinaryExpression`                                                                                                                                                                                | `Left`, `Right`                      |
| `PrefixUnaryExpression` / `PostfixUnaryExpression`                                                                                                                                                | `Operand`                            |
| `CallExpression` / `NewExpression`                                                                                                                                                                | `Expression` (callee)                |
| `ElementAccessExpression`                                                                                                                                                                         | `Expression`, `ArgumentExpression`   |
| `PropertyAccessExpression`                                                                                                                                                                        | `Expression` (object)                |
| `TemplateSpan`                                                                                                                                                                                    | `Expression`                         |
| `SpreadElement` / `AwaitExpression` / `YieldExpression` / `TypeOfExpression` / `VoidExpression` / `DeleteExpression`                                                                              | `Expression`                         |
| `ConditionalExpression`                                                                                                                                                                           | `Condition`, `WhenTrue`, `WhenFalse` |
| `VariableDeclaration` / `PropertyAssignment` / `BindingElement`                                                                                                                                   | `Initializer`                        |
| `IfStatement` / `WhileStatement` / `DoStatement` / `ForStatement` (condition) / `SwitchStatement` / `CaseClause` / `WithStatement` / `ReturnStatement` / `ThrowStatement` / `ExpressionStatement` | `Expression`                         |

A helper that embeds the check (e.g. `isNumeric`, `isStringType`) should call `ast.SkipParentheses` at the top rather than require every call site to unwrap — otherwise one forgotten caller is a silent divergence.

### Optional Chain

tsgo does not have a `ChainExpression` wrapper. Instead, every link in an optional chain (`PropertyAccess` / `ElementAccess` / `Call` / `NonNullExpression`) carries `NodeFlagsOptionalChain`. Parentheses break the chain — `(foo?.bar)(x)` parses as an ordinary call whose callee happens to be a paren-wrapped optional chain.

- `ast.IsOptionalChain(node)` — true iff node is a link in an optional chain.
- `ast.IsOptionalChainRoot(node)` — true iff node introduces the chain (holds the leading `?.`).
- `ast.IsOutermostOptionalChain(node)` — true iff node is the top of its chain.

When porting a rule that switches on "is the argument a `ChainExpression`?", translate to `ast.IsOptionalChain(arg)` on the tsgo-side argument after `ast.SkipParentheses`.

### Literal Kinds

ESTree uses one `Literal` node carrying a typed `value`. tsgo splits the literal family across kinds, and booleans / null are keyword tokens rather than literal nodes:

| Value                   | tsgo kind                              | Text accessor                                                        |
| ----------------------- | -------------------------------------- | -------------------------------------------------------------------- |
| Number                  | `KindNumericLiteral`                   | `node.AsNumericLiteral().Text` (raw source — may be `0x1` / `1_000`) |
| String                  | `KindStringLiteral`                    | `node.AsStringLiteral().Text` (cooked)                               |
| BigInt                  | `KindBigIntLiteral`                    | `node.AsBigIntLiteral().Text` (includes `n` suffix in source text)   |
| Regex                   | `KindRegularExpressionLiteral`         | `node.Text()`                                                        |
| `true` / `false`        | `KindTrueKeyword` / `KindFalseKeyword` | —                                                                    |
| `null`                  | `KindNullKeyword`                      | —                                                                    |
| `` `…` `` without `${}` | `KindNoSubstitutionTemplateLiteral`    | `node.AsNoSubstitutionTemplateLiteral().Text` (cooked)               |
| `` `…${x}…` ``          | `KindTemplateExpression`               | `Head.Text()` + each `TemplateSpan.Literal.Text()` (all cooked)      |

Common translations:

- ESTree `node.type === "Literal" && typeof node.value === "number"` → `node.Kind == ast.KindNumericLiteral`.
- ESTree's `isStringLiteral` helper (StringLiteral + TemplateLiteral) → `ast.IsStringLiteralLike(node)` covers `StringLiteral` + `NoSubstitutionTemplateLiteral`. For TemplateExpression add it explicitly.
- Comparing numeric literal value (e.g. "is this `1`"): `utils.NormalizeNumericLiteral(text) == "1"` — handles `1`, `1.0`, `0x1`, `1e0`, `1_000`, etc. with one comparison.

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

For fixes:

- `rule.RuleFixReplace(sf, node, text)` — replaces the trimmed span.
- `rule.RuleFixReplaceRange(range, text)` — replaces a specific range (useful for precision edits across multiple nodes).

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
        objectName = propAccess.Expression.AsIdentifier().Text()
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
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
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
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
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

| Method                            | Description              |
| --------------------------------- | ------------------------ |
| `GetTypeAtLocation(node)`         | Get the type of a node   |
| `GetSymbolAtLocation(node)`       | Get the symbol of a node |
| `TypeToString(type)`              | Convert type to string   |
| `GetSignaturesOfType(type, kind)` | Get function signatures  |

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

Reference: `internal/plugins/typescript/rules/no_explicit_any/no_explicit_any.go`

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
            return typeRef.TypeName.AsIdentifier().Text() == "any"
        }
    }
    return false
}
```

---

## Reporting Functions

`RuleContext` provides multiple reporting methods (defined in `internal/rule/rule.go`):

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

### Report with Autofix

```go
ctx.ReportNodeWithFixes(node, rule.RuleMessage{...},
    rule.RuleFix{
        Range: core.TextRange{Pos: start, End: end},
        Text:  "replacement text",  // Note: field is "Text", not "NewText"
    },
)
```

### Report with Suggestions

Suggestions are optional fixes that users can choose to apply:

```go
ctx.ReportNodeWithSuggestions(node, rule.RuleMessage{...},
    rule.RuleSuggestion{
        Message: rule.RuleMessage{
            Id:          "suggestRemove",
            Description: "Remove this statement",
        },
        FixesArr: []rule.RuleFix{  // Note: field is "FixesArr", not "Fixes"
            {Range: core.TextRange{...}, Text: ""},
        },
    },
)
```

---

## Fix Helper Functions

Defined in `internal/rule/rule.go`:

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
ctx.ReportNodeWithFixes(node, msg,
    rule.RuleFixRemoveRange(core.NewTextRange(dotStart, bindEnd)),   // removes ".bind"
    rule.RuleFixRemoveRange(core.NewTextRange(parenStart, parenEnd)), // removes "(arg)"
)
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
