package no_conditional_in

import (
	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type options struct {
	allowOptionalChaining bool
}

func parseOptions(rawOptions []any) options {
	opts := options{allowOptionalChaining: true}
	raw := rule.LegacyUnwrapOptions(rawOptions)
	optionArray := rule.NormalizeOptions(raw)
	if len(optionArray) == 0 {
		return opts
	}

	optionMap, ok := optionArray[0].(map[string]interface{})
	if !ok {
		return opts
	}
	if allow, ok := optionMap["allowOptionalChaining"].(bool); ok {
		opts.allowOptionalChaining = allow
	}
	return opts
}

func buildConditionalInTestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionalInTest",
		Description: "Avoid having conditionals in tests",
	}
}

func isOutermostOptionalChain(node *ast.Node) bool {
	return ast.IsOptionalChain(node) &&
		ast.IsOutermostOptionalChain(node) &&
		(node.Parent == nil || !ast.IsOptionalChain(node.Parent))
}

var NoConditionalInTestRule = rule.Rule{
	Name: "jest/no-conditional-in-test",
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		testCallDepth := 0
		testCalls := map[*ast.Node]bool{}

		reportConditional := func(node *ast.Node) {
			if testCallDepth > 0 {
				ctx.ReportNode(node, buildConditionalInTestMessage())
			}
		}

		reportOptionalChain := func(node *ast.Node) {
			if testCallDepth > 0 &&
				!opts.allowOptionalChaining &&
				isOutermostOptionalChain(node) {
				ctx.ReportNode(node, buildConditionalInTestMessage())
			}
		}

		return rule.RuleListeners{
			ast.KindIfStatement:           reportConditional,
			ast.KindSwitchStatement:       reportConditional,
			ast.KindConditionalExpression: reportConditional,
			ast.KindBinaryExpression: func(node *ast.Node) {
				if ast.IsLogicalExpression(node) {
					reportConditional(node)
				}
			},

			ast.KindPropertyAccessExpression: reportOptionalChain,
			ast.KindElementAccessExpression:  reportOptionalChain,

			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				isUnsupportedFitConcurrent := call != nil &&
					jestUtils.CalleeChainName(call.Expression) == "fit.concurrent"
				if !isUnsupportedFitConcurrent &&
					jestUtils.IsTypeOfJestFnCall(node, ctx, jestUtils.JestFnTypeTest) {
					testCalls[node] = true
					testCallDepth++
				}
				reportOptionalChain(node)
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if !testCalls[node] {
					return
				}
				delete(testCalls, node)
				testCallDepth--
			},
		}
	},
}
