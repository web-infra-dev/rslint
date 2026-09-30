package no_test_return_statement

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_test_return_statement"
)

// jestTestCallback returns the function a Jest test registration runs. Jest's
// only overloads are `(name, fn)` and `(name, fn, timeout)`, so the callback
// is always the second argument. A name there is followed to its binding
// through the file's references rather than by source text, so a shadowed or
// reassigned name is never attributed to an unrelated function.
//
// bindings caches each resolved name by symbol, including names that resolve
// to no function. Resolving a binding scans every reference to it, and a
// callback shared by N tests is resolved once per registration and again for
// each of its references, which is quadratic without the cache.
func jestTestCallback(
	ctx rule.RuleContext,
	analysis *jestUtils.JestCallAnalysis,
	bindings map[*ast.Symbol]*ast.Node,
	node *ast.Node,
) *ast.Node {
	if node.Kind != ast.KindCallExpression || analysis.ParseTestCall(node) == nil {
		return nil
	}
	arguments := node.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) < 2 {
		return nil
	}
	callback := ast.SkipParentheses(arguments.Nodes[1])
	if ast.IsFunctionExpressionOrArrowFunction(callback) {
		return callback
	}
	if callback.Kind != ast.KindIdentifier || ctx.Refs == nil {
		return nil
	}
	symbol := ctx.Refs.Resolve(callback)
	if symbol == nil {
		return nil
	}
	if function, ok := bindings[symbol]; ok {
		return function
	}
	function := testFramework.LocalFunctionBinding(ctx.SourceFile, ctx.Refs, symbol)
	bindings[symbol] = function
	return function
}

var NoTestReturnStatementRule = shared.NewRule(shared.Config{
	Name: "jest/no-test-return-statement",
	Message: rule.RuleMessage{
		Id:          "noReturnValue",
		Description: "Jest tests should not return a value",
	},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		bindings := map[*ast.Symbol]*ast.Node{}
		return shared.Runtime{
			TestCallback: func(node *ast.Node) *ast.Node {
				return jestTestCallback(ctx, analysis, bindings, node)
			},
		}
	},
})
