package no_conditional_in

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_conditional_in_test.schema.json
var schemaJSON []byte

type options struct {
	allowOptionalChaining bool
}

func parseOptions(rawOptions []any) options {
	opts := options{allowOptionalChaining: true}
	if len(rawOptions) == 0 {
		return opts
	}

	optionMap, _ := rawOptions[0].(map[string]any)
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
	Name:   "rstest/no-conditional-in-test",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		// The scope is the test registration call itself, so every argument of
		// the call is inside a test and a function that merely happens to be
		// named as the callback is not. The counter, rather than a flag, keeps
		// an inner registration exiting from clearing the outer test's state.
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
				reportRange := internalUtils.TrimNodeTextRange(ctx.SourceFile, node)
				for parent := node.Parent; parent != nil &&
					parent.Kind == ast.KindNonNullExpression &&
					parent.AsNonNullExpression().Expression == node; parent = node.Parent {
					node = parent
					reportRange = reportRange.WithEnd(node.End())
				}
				ctx.ReportRange(reportRange, buildConditionalInTestMessage())
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
			ast.KindNonNullExpression:        reportOptionalChain,

			ast.KindCallExpression: func(node *ast.Node) {
				reportOptionalChain(node)
				if analysis.ParseTestCall(node) != nil {
					testCalls[node] = true
					testCallDepth++
				}
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
