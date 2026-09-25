// Package prefer_mock_return_shorthand implements the body shared by the
// test-framework rules that prefer `mockReturnValue` over a `mockImplementation`
// whose callback does nothing but return a value.
//
// The analysis deciding whether the callback can become a value evaluated once
// lives in the mock_shorthand package, which the promise-shorthand rule shares.
// Nothing framework-specific is left to configure beyond the rule's name.
package prefer_mock_return_shorthand

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	"github.com/web-infra-dev/rslint/internal/utils/test_framework/mock_shorthand"
)

type Config struct {
	Name string
}

const (
	implementationMethod = "mockImplementation"
	returnValueMethod    = "mockReturnValue"
)

func useMockShorthandMessage(replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useMockShorthand",
		Description: "Prefer " + replacement,
	}
}

// buildsRejectedPromise reports whether the expression constructs a rejected
// promise, which the shorthand would construct once, when the mock is configured,
// instead of once per call.
//
// A rejected promise nobody has awaited yet is an unhandled rejection. While the
// callback builds it, the promise only exists after the mock is called, and the
// call site that triggered it is also what awaits it. Moved into
// `mockReturnValue`, it exists from the moment the mock is set up, so a mock that
// ends up not being called — the mock configured in `beforeEach` for a branch this
// test does not take, say — leaves a rejection nobody handles, which a test runner
// reports as a run-level error even though every test passed.
//
// Both reference plugins rewrite this shape anyway, and one of them documents
// `mockReturnValue(Promise.reject(...))` as an acceptable result. It is not
// reported here at all, rather than reported without a fix,
// because the advice itself is the problem: the safe rewrite of a rejecting mock
// is `mockRejectedValue`, which builds the promise per call, and that belongs to
// the promise-shorthand rule rather than this one.
func buildsRejectedPromise(ctx rule.RuleContext, expression *ast.Node) bool {
	_, method, _, ok := mock_shorthand.PromiseFactoryCall(ctx, expression)
	return ok && method == "reject"
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
					if !ok || name != mock_shorthand.WithOnce(implementationMethod, once) {
						return
					}

					// An `async` callback's result is a promise the shorthand
					// would not create. `mockResolvedValue` is that rewrite, and
					// it belongs to the promise-shorthand rule.
					callback := ast.SkipParentheses(arguments[0])
					if !mock_shorthand.IsCollapsibleCallback(callback, false) {
						return
					}

					expression := mock_shorthand.SingleReturnExpression(callback)
					if expression == nil || !mock_shorthand.IsEvaluatedOnceSafe(ctx, expression) ||
						buildsRejectedPromise(ctx, expression) ||
						mock_shorthand.ReadsCallSiteBinding(callback, expression) {
						return
					}

					replacement := mock_shorthand.WithOnce(returnValueMethod, once)
					ctx.ReportNodeWithDeferredFixes(accessor, useMockShorthandMessage(replacement), func() []rule.RuleFix {
						accessorRange, accessorText, ok := testFramework.AccessorReplacement(ctx.SourceFile, accessor, replacement)
						if !ok || mock_shorthand.DropsComment(ctx, callback, expression) ||
							mock_shorthand.DropsCallbackScope(ctx, callback, expression) {
							return nil
						}
						return []rule.RuleFix{
							rule.RuleFixReplaceRange(accessorRange, accessorText),
							rule.RuleFixReplace(ctx.SourceFile, callback, mock_shorthand.ArgumentText(ctx, expression)),
						}
					})
				},
			}
		},
	}
}
