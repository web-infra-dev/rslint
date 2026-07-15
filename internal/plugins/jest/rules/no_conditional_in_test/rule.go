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
	if !ast.IsOptionalChain(node) || !ast.IsOutermostOptionalChain(node) {
		return false
	}

	parent := node.Parent
	if parent == nil || !ast.IsOptionalChain(parent) {
		return true
	}

	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Expression != node
	case ast.KindElementAccessExpression:
		return parent.AsElementAccessExpression().Expression != node
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression != node
	default:
		return true
	}
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
				parsed := jestUtils.ParseJestFnCall(node, ctx)
				isUnsupportedFitConcurrent := parsed != nil &&
					parsed.Name == "fit" &&
					len(parsed.Members) == 1 &&
					parsed.Members[0] == "concurrent"
				if !isUnsupportedFitConcurrent && parsed != nil && parsed.Kind == jestUtils.JestFnTypeTest {
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
