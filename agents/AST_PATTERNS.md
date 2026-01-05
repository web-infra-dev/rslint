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

For rules that need type information, access `TypeChecker` via `RuleContext`:

```go
Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
    return rule.RuleListeners{
        ast.KindCallExpression: func(node *ast.Node) {
            if ctx.TypeChecker == nil {
                return // TypeChecker may not be available
            }

            // Get the type of an expression
            callExpr := node.AsCallExpression()
            exprType := ctx.TypeChecker.GetTypeAtLocation(callExpr.Expression)

            // Check if it's a Promise type
            if exprType != nil {
                typeStr := ctx.TypeChecker.TypeToString(exprType)
                // Use type information for analysis
            }
        },
    }
}
```

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

---

## See Also

- [PORT_RULE.md](./PORT_RULE.md) - Main rule porting workflow
- [UTILS_REFERENCE.md](./UTILS_REFERENCE.md) - Utility functions reference
- [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - Commands and checklist
