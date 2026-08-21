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

func enterTestCallback(
	testCallbackDepth *int,
	testCallbackFunctions map[*ast.Node]bool,
	node *ast.Node,
) {
	if node != nil && testCallbackFunctions[node] {
		*testCallbackDepth = *testCallbackDepth + 1
	}
}

func exitTestCallback(
	testCallbackDepth *int,
	testCallbackFunctions map[*ast.Node]bool,
	node *ast.Node,
) {
	if node != nil && testCallbackFunctions[node] {
		*testCallbackDepth = *testCallbackDepth - 1
	}
}

var NoConditionalInTestRule = rule.Rule{
	Name:   "rstest/no-conditional-in-test",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		testCallbackFunctions := analysis.Callbacks().Functions
		testCallbackDepth := 0

		reportConditional := func(node *ast.Node) {
			if testCallbackDepth > 0 {
				ctx.ReportNode(node, buildConditionalInTestMessage())
			}
		}

		reportOptionalChain := func(node *ast.Node) {
			if testCallbackDepth > 0 &&
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
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				enterTestCallback(&testCallbackDepth, testCallbackFunctions, node)
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				exitTestCallback(&testCallbackDepth, testCallbackFunctions, node)
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				enterTestCallback(&testCallbackDepth, testCallbackFunctions, node)
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				exitTestCallback(&testCallbackDepth, testCallbackFunctions, node)
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				enterTestCallback(&testCallbackDepth, testCallbackFunctions, node)
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				exitTestCallback(&testCallbackDepth, testCallbackFunctions, node)
			},

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
			ast.KindCallExpression:           reportOptionalChain,
			ast.KindNonNullExpression:        reportOptionalChain,
		}
	},
}
