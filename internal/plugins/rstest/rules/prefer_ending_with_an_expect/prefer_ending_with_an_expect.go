package prefer_ending_with_an_expect

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_ending_with_an_expect"
)

// rstestCallbackArgument returns the inline callback of a registration.
//
// Rstest accepts two shapes, `(name, fn, timeout?)` and `(name, options, fn?)`,
// so the callback sits second or third depending on the call. The third
// argument is preferred when it is a function literal, matching the resolution
// rstest/expect-expect consumes; in `(name, fn, timeout)` that position holds a
// number and falls through to the second argument.
//
// Only function literals are returned. A callback passed by reference
// (`test('x', run)`) is left alone: this rule would have to decide whether the
// referenced function's last statement is the one the test ends with, and every
// wrong answer there is a false "Tests should end with an assertion" on working
// code.
func rstestCallbackArgument(call *ast.CallExpression) *ast.Node {
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
		return nil
	}
	arguments := call.Arguments.Nodes
	if len(arguments) >= 3 {
		if callback := functionLiteralArgument(arguments[2]); callback != nil {
			return callback
		}
	}
	return functionLiteralArgument(arguments[1])
}

func functionLiteralArgument(argument *ast.Node) *ast.Node {
	if argument == nil {
		return nil
	}
	argument = internalUtils.SkipAssertionsAndParens(argument)
	if argument == nil || !ast.IsFunctionExpressionOrArrowFunction(argument) {
		return nil
	}
	return argument
}

// sourceMayContainRstestAssertion reports whether the file mentions `expect` or
// `assert` at all. Every form the analysis can resolve — `ctx.expect`,
// `rstest.expect`, `import { expect as check }`, `import.meta.rstest.expect`,
// and the same shapes for `assert` — spells one of those identifiers
// somewhere, so a file without either can skip the resolving hook and avoid
// forcing the analysis to collect test callbacks.
//
// Both names have to be checked. Gating on `expect` alone switched the hook off
// for a file that asserts only through `assert`, which left every resolved
// `assert` binding unrecognized however well the analysis could resolve it.
func sourceMayContainRstestAssertion(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.AsNode().Kind != ast.KindSourceFile {
		return true
	}
	return sourceFile.HasIdentifier("expect") || sourceFile.HasIdentifier("assert")
}

// rstestPropertyAssertion reports whether node is a Chai assertion that asserts
// through a property getter rather than a call, such as
// `expect(cart).to.be.empty` or `context.expect(ok).to.be.true`. Rstest bundles
// Chai, so these assert even though no matcher is invoked, and the statement a
// test ends with is then a property access rather than a call.
//
// The chain is handed to the same parser every other Rstest expect consumer
// uses, so what counts as a finished assertion is decided in one place. A chain
// the author left unfinished resolves a Reason rather than a matcher:
// `expect(cart).not` is a bare modifier and `expect(cart).toBe` is an uncalled
// matcher, and neither asserts, so both stay reportable.
func rstestPropertyAssertion(
	analysis *rstestUtils.RstestCallAnalysis,
	node *ast.Node,
) bool {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil ||
		(node.Kind != ast.KindPropertyAccessExpression &&
			node.Kind != ast.KindElementAccessExpression) {
		return false
	}
	head := chainHeadCall(node)
	if head == nil {
		return false
	}
	parsed := analysis.ParseExpectCall(head)
	if parsed == nil ||
		parsed.Head == nil ||
		parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
		len(parsed.Matchers) == 0 {
		return false
	}
	// The statement has to be the whole chain. A prefix of one, as in
	// `expect(cart).to.be.empty.and`, is a different expression than the one
	// the parser resolved a matcher for.
	if parsed.Expression != node {
		return false
	}
	return parsed.Matchers[len(parsed.Matchers)-1].Kind == rstestUtils.RstestExpectMatcherProperty
}

// chainHeadCall returns the call a member chain starts from, so
// `expect(cart).to.be.empty` yields `expect(cart)`.
//
// Parentheses are skipped at every step, not just at the outer expression.
// They may wrap any link of the chain — `(expect(cart)).to.be.empty` and
// `(expect(cart).to.be).empty` are the same assertion as the unparenthesized
// one — and the expect parser already reads through them, so stopping here
// would reject a chain the parser accepts.
func chainHeadCall(node *ast.Node) *ast.Node {
	for node != nil {
		node = ast.SkipParentheses(node)
		if node == nil {
			return nil
		}
		switch node.Kind {
		case ast.KindCallExpression:
			return node
		case ast.KindPropertyAccessExpression:
			node = node.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			node = node.AsElementAccessExpression().Expression
		default:
			return nil
		}
	}
	return nil
}

var PreferEndingWithAnExpectRule = shared.NewRule(shared.Config{
	Name: "rstest/prefer-ending-with-an-expect",
	// Rstest ships both `expect` and `assert` (chai) as globals
	// (packages/core/src/utils/constants.ts globalApiList), so the default
	// asserting functions match rstest/expect-expect rather than jest.
	DefaultAssertFunctionNames: []string{"expect", "assert"},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		// Rstest reaches `expect` through the test context, namespace imports and
		// import aliases, and reaches Chai's `assert` through all but the test
		// context, none of which the callee-text patterns can match. Reuse the
		// same resolution rstest/expect-expect consumes so both rules agree on
		// what an assertion is.
		var isAssertion func(node *ast.Node) bool
		if sourceMayContainRstestAssertion(ctx.SourceFile) {
			isAssertion = func(node *ast.Node) bool {
				return analysis.IsExpectCall(node) || analysis.IsAssertCall(node)
			}
		}
		return shared.Runtime{
			IsAssertion: isAssertion,
			IsPropertyAssertion: func(node *ast.Node) bool {
				return rstestPropertyAssertion(analysis, node)
			},
			ClassifyTest: func(node *ast.Node) shared.TestBlock {
				parsed := analysis.ParseTestCall(node)
				if parsed == nil {
					return shared.TestBlock{}
				}
				// A `.todo` registration never runs the function it is given, so
				// requiring an assertion at its end would report code that cannot
				// fail. Todo is a semantic field that survives const aliases, so
				// never derive the exemption from the call-site Members.
				return shared.TestBlock{IsTest: true, IsExempt: parsed.Todo}
			},
			Callback: rstestCallbackArgument,
		}
	},
})
