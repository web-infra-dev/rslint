package prefer_ending_with_an_expect

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed prefer_ending_with_an_expect.schema.json
var schemaJSON []byte

func buildMustEndWithExpectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mustEndWithExpect",
		Description: "Tests should end with an assertion",
	}
}

func lastFunctionStatement(fn *ast.Node) *ast.Node {
	if fn == nil {
		return nil
	}
	body := fn.Body()
	if body == nil {
		return nil
	}
	if body.Kind != ast.KindBlock {
		return ast.SkipParentheses(body)
	}

	block := body.AsBlock()
	if block == nil || block.Statements == nil || len(block.Statements.Nodes) == 0 {
		return nil
	}
	last := block.Statements.Nodes[len(block.Statements.Nodes)-1]
	if last.Kind == ast.KindExpressionStatement {
		last = last.AsExpressionStatement().Expression
	}
	return ast.SkipParentheses(last)
}

func isAssertionCall(node *ast.Node, ctx rule.RuleContext, patterns []*esregexp.RegExp) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindAwaitExpression {
		node = ast.SkipParentheses(node.AsAwaitExpression().Expression)
	}
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	if jestUtils.IsTypeOfJestFnCall(node, ctx, jestUtils.JestFnTypeExpect) {
		return true
	}
	return testFramework.MatchesAssertName(
		testFramework.CalleeChainName(node.AsCallExpression().Expression),
		patterns,
	)
}

func isTestBlockCall(node *ast.Node, call *ast.CallExpression, ctx rule.RuleContext, additional []string) bool {
	if jestUtils.IsTypeOfJestFnCall(node, ctx, jestUtils.JestFnTypeTest) {
		return true
	}
	if len(additional) == 0 {
		return false
	}
	return slices.Contains(additional, testFramework.CalleeChainName(call.Expression))
}

var PreferEndingWithAnExpectRule = rule.Rule{
	Name:   "jest/prefer-ending-with-an-expect",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		parsedOptions := testFramework.ParseAssertionFunctionOptions(options, []string{"expect"})
		patterns := testFramework.CompileAssertPatterns(parsedOptions.AssertFunctionNames)
		additional := parsedOptions.AdditionalTestBlockFunctions

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil ||
					!isTestBlockCall(node, call, ctx, additional) ||
					call.Arguments == nil ||
					len(call.Arguments.Nodes) < 2 {
					return
				}

				callback := ast.SkipParentheses(call.Arguments.Nodes[1])
				if callback == nil ||
					!ast.IsFunctionExpressionOrArrowFunction(callback) ||
					isAssertionCall(lastFunctionStatement(callback), ctx, patterns) {
					return
				}

				ctx.ReportNode(call.Expression, buildMustEndWithExpectMessage())
			},
		}
	},
}
