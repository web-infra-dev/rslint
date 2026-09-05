// Package prefer_then_catch ports eslint-plugin-unicorn's
// `prefer-then-catch` rule.
//
// It flags `.then(onFulfilled, onRejected)` calls where the rejection handler
// can be moved into a chained `.catch(...)`, and offers a suggestion that
// rewrites the call accordingly.
package prefer_then_catch

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageID           = "prefer-then-catch"
	suggestionMessageID = "prefer-then-catch/suggestion"
)

var (
	message = rule.RuleMessage{
		Id:          messageID,
		Description: "Prefer `.then(…).catch(…)` over passing a rejection handler to `.then()`.",
	}
	suggestionMessage = rule.RuleMessage{
		Id:          suggestionMessageID,
		Description: "Move the rejection handler to `.catch()`.",
	}
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-then-catch.js
var PreferThenCatchRule = rule.Rule{
	Name:   "unicorn/prefer-then-catch",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		twoArgs := 2
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "then",
					ArgumentsLength:     &twoArgs,
					RejectSpreadElement: true,
				})
				if !ok {
					return
				}

				args := node.Arguments()
				if len(args) != 2 {
					return
				}
				fulfillmentHandler := args[0]
				rejectionHandler := args[1]
				if isNullish(fulfillmentHandler, ctx) || isNullish(rejectionHandler, ctx) {
					return
				}

				if !canThenResultCatch(call, node, ctx) {
					return
				}

				ctx.ReportNodeWithDeferredSuggestions(
					call.Property,
					message,
					func() []rule.RuleSuggestion {
						return buildSuggestion(ctx, node, rejectionHandler)
					},
				)
			},
		}
	},
}

// isNullish mirrors upstream's isNullish: `null`, `void <expr>`, or the
// identifier `undefined` when it resolves to the global. Parentheses and
// TypeScript assertion wrappers are transparent.
func isNullish(node *ast.Node, ctx rule.RuleContext) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindVoidExpression:
		return true
	case ast.KindIdentifier:
		if node.AsIdentifier().Text != "undefined" {
			return false
		}
		return isGlobalUndefined(ctx, node)
	}
	return false
}

// isGlobalUndefined reports whether the identifier resolves to the global
// `undefined`, matching upstream's `isGlobalIdentifier` check.
func isGlobalUndefined(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.Resolve(node)
	if symbol == nil {
		return false
	}
	// A locally-declared `undefined` (parameter, var) shadows the global; only
	// an out-of-file resolution (lib / ambient) counts as global.
	return !utils.IsSymbolDeclaredInFile(symbol, ctx.SourceFile)
}

// canThenResultCatch mirrors upstream's canThenResultCatch. Without a
// type-aware program the rule reports by default; with a type checker it only
// reports when the receiver's `.catch` is callable.
func canThenResultCatch(call unicornutil.DotMethodCall, node *ast.Node, ctx rule.RuleContext) bool {
	if ctx.TypeChecker == nil {
		return true
	}

	defer func() {
		// TypeScript can throw while resolving incomplete projects; mirror
		// upstream's try/catch by reporting on failure.
		_ = recover()
	}()

	tc := ctx.TypeChecker
	receiverType := tc.GetNonNullableType(tc.GetTypeAtLocation(call.Object))
	if utils.IsTypeAnyType(receiverType) || utils.IsTypeUnknownType(receiverType) {
		return true
	}
	if !isNativePromiseType(ctx, receiverType) {
		return false
	}

	resultType := tc.GetTypeAtLocation(node)
	return hasCallableCatch(resultType, tc, node)
}

// isNativePromiseType mirrors upstream's isNativePromiseType: it accepts only
// a type whose direct symbol is the default-library Promise. In particular,
// this deliberately does not follow base types, intersection constituents, or
// type-parameter constraints.
func isNativePromiseType(ctx rule.RuleContext, t *checker.Type) bool {
	symbol := checker.Type_symbol(t)
	return symbol != nil && symbol.Name == "Promise" && utils.IsSymbolFromDefaultLibrary(ctx.Program(), symbol)
}

// hasCallableCatch mirrors upstream's hasCallableCatch. A union type is only
// considered catch-able when every constituent has a callable `catch`.
func hasCallableCatch(t *checker.Type, tc *checker.Checker, location *ast.Node) bool {
	if utils.IsUnionType(t) {
		for _, part := range t.Types() {
			if !hasCallableCatch(part, tc, location) {
				return false
			}
		}
		return true
	}

	catchMethod := checker.Checker_getPropertyOfType(tc, t, "catch")
	if catchMethod == nil {
		return false
	}
	catchType := tc.GetTypeOfSymbolAtLocation(catchMethod, location)
	return len(utils.GetCallSignatures(tc, catchType)) > 0
}

// isRejectionHandlerSafeToMove reports whether the rejection handler can be
// pulled out of `.then(...)` and re-attached as `.catch(...)`. Only identifiers
// and function-like nodes are safe; any call expression (which may have side
// effects) is rejected.
func isRejectionHandlerSafeToMove(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	return ast.IsIdentifier(node) || ast.IsFunctionLike(node)
}

// argumentRemovalRange returns the half-open source range to remove when
// dropping the rejection-handler argument. Mirrors upstream's
// getArgumentRemovalRange + getRejectionHandlerRemovalRange: the range starts
// at the comma BEFORE the rejection handler and ends just after the trailing
// comma (or just after the handler if no trailing comma).
func argumentRemovalRange(sourceFile *ast.SourceFile, rejectionHandler *ast.Node) (core.TextRange, bool) {
	handlerStart := scanner.SkipTrivia(sourceFile.Text(), rejectionHandler.Pos())
	start := previousComma(sourceFile, handlerStart)
	if start < 0 {
		return core.TextRange{}, false
	}

	end := trailingCommaEnd(sourceFile, rejectionHandler.End())

	return core.NewTextRange(start, end), true
}

// trailingCommaEnd returns rejectionHandlerEnd extended to just past a trailing
// comma if one is present before the next non-trivia token. Mirrors upstream's
// `getTokenAfter(lastToken)` followed by an `isCommaToken` check.
func trailingCommaEnd(sourceFile *ast.SourceFile, rejectionHandlerEnd int) int {
	text := sourceFile.Text()
	end := rejectionHandlerEnd
	for end < len(text) {
		ch := text[end]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			end++
			continue
		}
		if ch == ',' {
			return end + 1
		}
		break
	}
	return rejectionHandlerEnd
}

// previousComma walks backward from pos through trivia and returns the byte
// offset of the nearest preceding `,`, or -1 if none is found before any other
// non-trivia token. Mirrors the upstream `getTokenBefore(firstToken)` lookup
// without using a backwards scanner.
func previousComma(sourceFile *ast.SourceFile, pos int) int {
	text := sourceFile.Text()
	i := pos - 1
	for i >= 0 {
		ch := text[i]
		switch {
		case ch == ',':
			return i
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			i--
		case ch == '/' && i > 0 && text[i-1] == '/':
			// line comment — find its start
			j := i - 2
			for j >= 0 && text[j] != '\n' {
				j--
			}
			i = j
		case ch == '/' && i > 0 && text[i-1] == '*':
			// block comment — find its start
			j := i - 2
			for j >= 1 && (text[j] != '*' || text[j-1] != '/') {
				j--
			}
			i = j - 1
			if i < 0 {
				return -1
			}
		default:
			return -1
		}
	}
	return -1
}

// hasTrailingArgumentComment reports whether the rejection handler is followed
// by an inline or block comment before the closing paren. Upstream's
// `getParentheses` walks the paren chain, which our reparser already flattens
// at parse time; matching the same observable behavior only requires checking
// from the handler's last token up to the closing paren.
func hasTrailingArgumentComment(comments []*ast.CommentRange, rejectionHandler *ast.Node, argsCloseParen int) bool {
	for _, comment := range comments {
		if comment.Pos() < rejectionHandler.End() {
			continue
		}
		if comment.Pos() >= argsCloseParen {
			break
		}
		return true
	}
	return false
}

// buildSuggestion constructs the `.then(...).catch(...)` rewrite, or returns
// nil when the rejection handler cannot safely be moved (side effects, a
// comment in the discarded separator/trivia, or a comment between it and the
// closing paren). Comments within the handler move with its source text.
func buildSuggestion(ctx rule.RuleContext, callExpression *ast.Node, rejectionHandler *ast.Node) []rule.RuleSuggestion {
	if !isRejectionHandlerSafeToMove(rejectionHandler) {
		return nil
	}

	sourceFile := ctx.SourceFile
	removalRange, ok := argumentRemovalRange(sourceFile, rejectionHandler)
	if !ok {
		return nil
	}

	argsCloseParen := callExpression.End() - 1
	if sourceFile.Text()[argsCloseParen] != ')' {
		argsCloseParen = callExpression.End()
	}

	comments := ctx.Comments.All()
	// The removal range also covers the handler, whose source is inserted into
	// `.catch(...)`. Only comments in the separator before it are discarded.
	if utils.HasCommentInSpan(comments, removalRange.Pos(), rejectionHandler.Pos()) {
		return nil
	}
	if hasTrailingArgumentComment(comments, rejectionHandler, argsCloseParen) {
		return nil
	}

	rejectionHandlerText := strings.TrimLeft(sourceFile.Text()[rejectionHandler.Pos():rejectionHandler.End()], " \t\n\r")
	insertRange := core.NewTextRange(callExpression.End(), callExpression.End())

	return []rule.RuleSuggestion{{
		Message: suggestionMessage,
		FixesArr: []rule.RuleFix{
			rule.RuleFixRemoveRange(removalRange),
			rule.RuleFixReplaceRange(insertRange, ".catch("+rejectionHandlerText+")"),
		},
	}}
}
