// Package prefer_mock_promise_shorthand implements the body shared by the
// test-framework rules that prefer `mockResolvedValue` / `mockRejectedValue` over
// a `mockReturnValue` or `mockImplementation` that hands back
// `Promise.resolve(...)` / `Promise.reject(...)`.
//
// The two methods move different things. `mockReturnValue(Promise.resolve(x))`
// already evaluates `x` once, when the mock is configured, and the shorthand
// evaluates it at the same moment, so the rewrite only changes when the promise is
// built: once per call instead of once overall. For a rejected promise that is a
// fix in itself — one built at configuration time is an unhandled rejection if the
// mock is never called. `mockImplementation(() => Promise.resolve(x))` evaluates
// `x` on every call, so the rewrite moves `x` to configuration time and has to
// prove that preserves what the callback did; the mock_shorthand package holds
// that analysis, which the return-shorthand rule shares.
package prefer_mock_promise_shorthand

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	"github.com/web-infra-dev/rslint/internal/utils/test_framework/mock_shorthand"
)

type Config struct {
	Name string
}

const (
	implementationMethod = "mockImplementation"
	returnValueMethod    = "mockReturnValue"
	resolvedValueMethod  = "mockResolvedValue"
	rejectedValueMethod  = "mockRejectedValue"
)

func useMockShorthandMessage(replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useMockShorthand",
		Description: "Prefer " + replacement,
	}
}

// isEvaluatedOnceSafe reports whether the value a callback returns can move out
// of the callback, where it is evaluated per call, into the shorthand's argument,
// where it is evaluated once.
//
// On top of what the return-shorthand rule checks, two shapes only this rule
// meets are excluded. An `async` callback qualifies here, because an async
// function that returns a promise settles the same way that promise does, but an
// `await` inside the value would leave the function it belongs to. And a spread
// argument, `Promise.resolve(...values)`, iterates `values` on every call: an
// iterator the first call exhausts gives the next one `undefined`, and moving the
// spread into the shorthand iterates it once. Both reference plugins rewrite both.
func isEvaluatedOnceSafe(ctx rule.RuleContext, callback, expression, promiseCall *ast.Node) bool {
	if !mock_shorthand.IsEvaluatedOnceSafe(ctx, expression) ||
		mock_shorthand.ReadsCallSiteBinding(callback, expression) ||
		mock_shorthand.ContainsAwait(expression) {
		return false
	}
	for _, argument := range promiseCall.Arguments() {
		if argument.Kind == ast.KindSpreadElement {
			return false
		}
	}
	return true
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					arguments := node.Arguments()
					if len(arguments) == 0 {
						return
					}

					accessor, name, once, ok := mock_shorthand.CalledMethod(node.AsCallExpression().Expression)
					if !ok {
						return
					}

					// replaced is what the rewrite overwrites with the promise's
					// argument: the `mockReturnValue` argument itself, or the whole
					// `mockImplementation` callback.
					replaced := ast.SkipParentheses(arguments[0])
					var callback, returned *ast.Node
					switch name {
					case mock_shorthand.WithOnce(returnValueMethod, once):
						returned = replaced
					case mock_shorthand.WithOnce(implementationMethod, once):
						callback = replaced
						if !mock_shorthand.IsCollapsibleCallback(callback, true) {
							return
						}
						returned = mock_shorthand.SingleReturnExpression(callback)
						if returned == nil {
							return
						}
					default:
						return
					}

					promiseCall, method, wrapped, ok := mock_shorthand.PromiseFactoryCall(ctx, returned)
					if !ok || callback != nil && !isEvaluatedOnceSafe(ctx, callback, returned, promiseCall) {
						return
					}

					replacement := mock_shorthand.WithOnce(resolvedValueMethod, once)
					if method == "reject" {
						replacement = mock_shorthand.WithOnce(rejectedValueMethod, once)
					}
					ctx.ReportNodeWithDeferredFixes(accessor, useMockShorthandMessage(replacement), func() []rule.RuleFix {
						return buildFix(ctx, node, accessor, replacement, arguments[0], replaced, callback, promiseCall, method, wrapped)
					})
				},
			}
		},
	}
}

// buildFix renames the method and replaces its first argument with the promise's
// argument, or with `undefined` when the promise was built without one: both
// Promise methods take the value optionally, the shorthands require it. Where a
// local binding shadows `undefined`, `void 0` stands in for it, since the
// identifier would read that binding instead.
//
// Parentheses around the first argument are kept, as upstream keeps them, except
// around a spread: `(...values)` is not an expression, so the spread replaces
// them too.
//
// The fix is withheld when the rewrite cannot keep the call's meaning:
//
//   - more than one promise argument, as upstream does, since the shorthand has
//     nowhere to put the rest;
//   - type arguments on the mock call, which describe the value the old method
//     took — a promise, or a function returning one — and not the settled value
//     the new method takes;
//   - a type assertion around a resolved promise, `Promise.resolve(x) as
//     Promise<T>`, which the rewrite would drop while the shorthand's parameter
//     is typed from the mock. A rejected promise's error is `unknown` in every
//     framework, so its assertion carries nothing and it stays fixable;
//   - a resolved value that is not a primitive literal; see
//     isTypeSafeResolvedValue;
//   - a comment or declaration the rewrite would delete.
func buildFix(
	ctx rule.RuleContext,
	call, accessor *ast.Node,
	replacement string,
	argument, replaced, callback, promiseCall *ast.Node,
	method string,
	wrapped bool,
) []rule.RuleFix {
	promiseArguments := promiseCall.Arguments()
	if len(promiseArguments) > 1 || call.AsCallExpression().TypeArguments != nil ||
		wrapped && method == "resolve" {
		return nil
	}

	var kept *ast.Node
	if len(promiseArguments) == 1 {
		kept = promiseArguments[0]
		if kept.Kind == ast.KindSpreadElement {
			replaced = argument
		}
	}
	if method == "resolve" && kept != nil && !isTypeSafeResolvedValue(kept) {
		return nil
	}
	if mock_shorthand.DropsComment(ctx, replaced, kept) {
		return nil
	}
	if callback != nil && mock_shorthand.DropsCallbackScope(ctx, callback, kept) {
		return nil
	}

	accessorRange, accessorText, ok := testFramework.AccessorReplacement(ctx.SourceFile, accessor, replacement)
	if !ok {
		return nil
	}
	argumentText := "undefined"
	if kept != nil {
		argumentText = mock_shorthand.ArgumentText(ctx, kept)
	} else if utils.IsShadowed(replaced, argumentText) {
		argumentText = "void 0"
	}
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(accessorRange, accessorText),
		rule.RuleFixReplace(ctx.SourceFile, replaced, argumentText),
	}
}

// isTypeSafeResolvedValue reports whether moving the value out of
// `Promise.resolve(value)` and into `mockResolvedValue(value)` keeps the call
// type-checking.
//
// The two check the value differently. `Promise.resolve` infers its own type
// argument from the value and only the resulting promise is compared with what
// the mock returns; the shorthand compares the value directly with the settled
// type. So a value that is itself a promise, one whose type is an unresolved type
// parameter that may be a promise, and a fresh object literal carrying a property
// the settled type lacks all type-check before the rewrite and fail after it,
// while the run-time behavior is the same. A primitive literal can be none of
// those, so only primitive literals are rewritten; everything else is reported
// without a fix. A rejected promise's reason is untyped in every framework and
// needs no such check.
func isTypeSafeResolvedValue(value *ast.Node) bool {
	value = ast.SkipParentheses(value)
	switch value.Kind {
	case ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword,
		ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindPrefixUnaryExpression:
		unary := value.AsPrefixUnaryExpression()
		operand := ast.SkipParentheses(unary.Operand)
		return (unary.Operator == ast.KindMinusToken || unary.Operator == ast.KindPlusToken) &&
			(operand.Kind == ast.KindNumericLiteral || operand.Kind == ast.KindBigIntLiteral)
	case ast.KindVoidExpression:
		return true
	case ast.KindIdentifier:
		return value.Text() == "undefined" && !utils.IsShadowed(value, "undefined")
	}
	return false
}
