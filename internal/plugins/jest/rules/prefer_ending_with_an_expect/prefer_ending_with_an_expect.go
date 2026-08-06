package prefer_ending_with_an_expect

import (
	_ "embed"
	"regexp"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed prefer_ending_with_an_expect.schema.json
var schemaJSON []byte

func mustEndWithExpectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mustEndWithExpect",
		Description: "Tests should end with an assertion",
	}
}

func inlineFunctionBody(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsFunctionExpressionOrArrowFunction(node) {
		return nil
	}
	if node.Kind == ast.KindArrowFunction {
		return node.AsArrowFunction().Body
	}
	return node.AsFunctionExpression().Body
}

func lastFunctionStatement(fn *ast.Node) *ast.Node {
	body := inlineFunctionBody(fn)
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

func isAssertionCall(node *ast.Node, ctx rule.RuleContext, patterns []*regexp.Regexp) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node != nil && node.Kind == ast.KindAwaitExpression {
		node = ast.SkipParentheses(node.AsAwaitExpression().Expression)
	}
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	if jestUtils.IsTypeOfJestFnCall(node, ctx, jestUtils.JestFnTypeExpect) {
		return true
	}
	return jestUtils.MatchesAssertFunctionName(
		jestUtils.CalleeChainName(node.AsCallExpression().Expression),
		patterns,
	)
}

var PreferEndingWithAnExpectRule = rule.Rule{
	Name:   "jest/prefer-ending-with-an-expect",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		parsedOptions := jestUtils.ParseAssertionFunctionOptions(options)
		patterns := jestUtils.CompileAssertFunctionNamePatterns(parsedOptions.AssertFunctionNames)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil {
					return
				}

				calleeName := jestUtils.CalleeChainName(call.Expression)
				parsedCall := jestUtils.ParseJestFnCall(node, ctx)
				isJestTest := parsedCall != nil && parsedCall.Kind == jestUtils.JestFnTypeTest
				isAdditionalTest := slices.Contains(parsedOptions.AdditionalTestBlockFunctions, calleeName)
				if !isJestTest && !isAdditionalTest {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
					return
				}

				callback := call.Arguments.Nodes[1]
				if callback == nil {
					return
				}
				callback = ast.SkipParentheses(callback)
				if !ast.IsFunctionExpressionOrArrowFunction(callback) {
					return
				}
				if isAssertionCall(lastFunctionStatement(callback), ctx, patterns) {
					return
				}

				ctx.ReportNode(call.Expression, mustEndWithExpectMessage())
			},
		}
	},
}
