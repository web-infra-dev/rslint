package no_promise_executor_return

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_promise_executor_return.schema.json
var schemaJSON []byte

const promiseName = "Promise"

// promiseMeaning is the set of declaration spaces an ESLint value reference
// searches. A `Promise` binding in any of them — including an empty
// `namespace Promise {}` or a type-only import alias — makes the callee local.
const promiseMeaning = ast.SymbolFlagsValue | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias

// returnKeywordLength is the width of the `return` keyword. A ReturnStatement
// always starts with that keyword, so its end position is the statement start
// plus this length — no token scan needed.
const returnKeywordLength = len("return")

// voidPrecedence mirrors ESLint's
// astUtils.getPrecedence({ type: "UnaryExpression", operator: "void" }).
const voidPrecedence = 16

var returnsValueMessage = rule.RuleMessage{
	Id:          "returnsValue",
	Description: "Return values from promise executor functions cannot be read.",
}

var prependVoidMessage = rule.RuleMessage{
	Id:          "prependVoid",
	Description: "Prepend `void` to the expression.",
}

var wrapBracesMessage = rule.RuleMessage{
	Id:          "wrapBraces",
	Description: "Wrap the expression in `{}`.",
}

type options struct {
	allowVoid bool
}

func parseOptions(rawOptions []any) options {
	var opts options
	if len(rawOptions) == 0 {
		return opts
	}
	optionsMap, _ := rawOptions[0].(map[string]any)
	opts.allowVoid, _ = optionsMap["allowVoid"].(bool)
	return opts
}

// sourceMayUsePromise reports whether the file spells `Promise` anywhere. The
// rule only fires inside a `new Promise(...)` executor, so a file whose parsed
// identifier table lacks that name can skip every listener. Stay conservative
// for direct callers that do not provide parser metadata.
func sourceMayUsePromise(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.AsNode().Kind != ast.KindSourceFile {
		return true
	}
	ok := sourceFile.HasIdentifier(promiseName)
	return ok
}

// isVoidExpression mirrors upstream expressionIsVoid. tsgo gives `void x` its
// own kind instead of ESTree's UnaryExpression + operator pair.
func isVoidExpression(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindVoidExpression
}

// enclosingFunction returns the nearest ancestor that owns node's ESLint code
// path — any function-like declaration or a class static block. A nil result
// means node belongs to the module's own code path (a global return).
func enclosingFunction(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(current) {
			return current
		}
	}
	return nil
}

// isUnnamedFunctionOrClassExpression reports whether wrapping expr in braces
// would turn it into an (illegal) anonymous declaration statement.
func isUnnamedFunctionOrClassExpression(expr *ast.Node) bool {
	if expr.Kind != ast.KindFunctionExpression && expr.Kind != ast.KindClassExpression {
		return false
	}
	return expr.Name() == nil
}

// voidPrependFixes mirrors upstream voidPrependFixer. `void ` goes in right
// after the `=>` or `return` token that introduces expr — not before expr
// itself — so any parentheses and comments in between stay outside the `void`.
func voidPrependFixes(ctx rule.RuleContext, expr *ast.Node, anchorEnd int, afterReturn bool) []rule.RuleFix {
	firstTokenStart := scanner.SkipTrivia(ctx.SourceFile.Text(), anchorEnd)

	requiresParens := utils.EslintLikePrecedence(expr) < voidPrecedence &&
		!ast.IsParenthesizedExpression(expr.Parent)

	inserted := "void "
	// In `return(x)` the keyword touches the value, so `void` needs a leading
	// space to stay a separate token; after `=>` it is already unambiguous.
	if afterReturn && anchorEnd == firstTokenStart {
		inserted = " " + inserted
	}
	if !requiresParens {
		return []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(firstTokenStart, firstTokenStart), inserted),
		}
	}
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(firstTokenStart, firstTokenStart), inserted+"("),
		rule.RuleFixInsertAfter(expr, ")"),
	}
}

// curlyWrapFixes mirrors upstream curlyWrapFixer: it turns a concise arrow body
// into a block body by bracketing everything after the `=>` token.
func curlyWrapFixes(ctx rule.RuleContext, arrow *ast.Node) []rule.RuleFix {
	arrowToken := arrow.AsArrowFunction().EqualsGreaterThanToken
	bodyStart := scanner.SkipTrivia(ctx.SourceFile.Text(), arrowToken.End())
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(bodyStart, bodyStart), "{"),
		rule.RuleFixInsertAfter(arrow, "}"),
	}
}

type noPromiseExecutorReturnState struct {
	ctx  rule.RuleContext
	opts options
}

// isPromiseExecutor reports whether fn is the first argument of a
// `new Promise(...)` expression whose callee is the global `Promise`.
func (state *noPromiseExecutorReturnState) isPromiseExecutor(fn *ast.Node) bool {
	// ESTree drops parentheses, so `new Promise((function () {}))` still has the
	// function itself as arguments[0]. Compare against the outermost wrapper
	// instead, which is the node tsgo stores in the argument list.
	argument := utils.OutermostParenthesizedExpression(fn)
	parent := argument.Parent
	if parent == nil || parent.Kind != ast.KindNewExpression {
		return false
	}

	newExpression := parent.AsNewExpression()
	if newExpression.Arguments == nil || len(newExpression.Arguments.Nodes) == 0 ||
		newExpression.Arguments.Nodes[0] != argument {
		return false
	}

	callee := ast.SkipParentheses(newExpression.Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier ||
		callee.AsIdentifier().Text != promiseName {
		return false
	}

	// Upstream requires sourceCode.isGlobalReference(callee): the name must
	// still resolve to the configured global rather than a source declaration.
	return state.ctx.Refs.IsGlobalNameReference(callee, promiseName, promiseMeaning)
}

// checkConciseArrowBody reports `new Promise(r => value)`, whose body is an
// implicit return.
func (state *noPromiseExecutorReturnState) checkConciseArrowBody(node *ast.Node) {
	arrow := node.AsArrowFunction()
	body := arrow.Body
	if body == nil || body.Kind == ast.KindBlock || arrow.EqualsGreaterThanToken == nil {
		return
	}
	if !state.isPromiseExecutor(node) {
		return
	}

	// NOTE: Unlike ESLint, tsgo keeps parentheses as real nodes. Upstream reads
	// and reports the paren-free expression, so unwrap before every check to
	// keep both the report range and the suggestion set aligned.
	expr := ast.SkipParentheses(body)
	if state.opts.allowVoid && isVoidExpression(expr) {
		return
	}

	state.ctx.ReportNodeWithDeferredSuggestions(expr, returnsValueMessage, func() []rule.RuleSuggestion {
		var suggestions []rule.RuleSuggestion
		// Without allowVoid, prepending `void` would not silence the rule, so
		// upstream deliberately withholds that suggestion.
		if state.opts.allowVoid {
			suggestions = append(suggestions, rule.RuleSuggestion{
				Message:  prependVoidMessage,
				FixesArr: voidPrependFixes(state.ctx, expr, arrow.EqualsGreaterThanToken.End(), false),
			})
		}
		if !isUnnamedFunctionOrClassExpression(expr) {
			suggestions = append(suggestions, rule.RuleSuggestion{
				Message:  wrapBracesMessage,
				FixesArr: curlyWrapFixes(state.ctx, node),
			})
		}
		return suggestions
	})
}

func (state *noPromiseExecutorReturnState) checkReturnStatement(node *ast.Node) {
	returnStatement := node.AsReturnStatement()
	if returnStatement.Expression == nil {
		return
	}

	executor := enclosingFunction(node)
	if executor == nil || !ast.IsFunctionExpressionOrArrowFunction(executor) ||
		!state.isPromiseExecutor(executor) {
		return
	}

	if !state.opts.allowVoid {
		state.ctx.ReportNode(node, returnsValueMessage)
		return
	}

	expr := ast.SkipParentheses(returnStatement.Expression)
	if isVoidExpression(expr) {
		return
	}

	state.ctx.ReportNodeWithDeferredSuggestions(node, returnsValueMessage, func() []rule.RuleSuggestion {
		returnKeywordEnd := scanner.SkipTrivia(state.ctx.SourceFile.Text(), node.Pos()) + returnKeywordLength
		return []rule.RuleSuggestion{{
			Message:  prependVoidMessage,
			FixesArr: voidPrependFixes(state.ctx, expr, returnKeywordEnd, true),
		}}
	})
}

// NoPromiseExecutorReturnRule disallows returning values from Promise executor
// functions.
var NoPromiseExecutorReturnRule = rule.Rule{
	Name:   "no-promise-executor-return",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		if !sourceMayUsePromise(ctx.SourceFile) {
			return nil
		}
		// `new Promise(...)` can only be the built-in executor call when the
		// global exists; `/* globals Promise:off */` removes it.
		if !ctx.Globals.Access(promiseName).IsDeclared() {
			return nil
		}

		state := &noPromiseExecutorReturnState{ctx: ctx, opts: parseOptions(rawOptions)}
		return rule.RuleListeners{
			ast.KindArrowFunction:   state.checkConciseArrowBody,
			ast.KindReturnStatement: state.checkReturnStatement,
		}
	},
}
