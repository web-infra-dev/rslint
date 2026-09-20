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

// sourceMayContainRstestExpect reports whether the file mentions `expect` at
// all. Every form the analysis can resolve — `ctx.expect`, `rstest.expect`,
// `import { expect as check }`, `import.meta.rstest.expect` — spells the
// identifier somewhere, so a file without it can skip the resolving hook and
// avoid forcing the analysis to collect test callbacks.
func sourceMayContainRstestExpect(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.AsNode().Kind != ast.KindSourceFile {
		return true
	}
	return sourceFile.HasIdentifier("expect")
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
		// import aliases, none of which the callee-text patterns can match. Reuse
		// the same resolution rstest/expect-expect consumes so both rules agree on
		// what an assertion is.
		var isAssertion func(node *ast.Node) bool
		if sourceMayContainRstestExpect(ctx.SourceFile) {
			isAssertion = func(node *ast.Node) bool {
				return analysis.IsExpectCall(node)
			}
		}
		return shared.Runtime{
			IsAssertion: isAssertion,
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
