package max_nested_calls

import (
	_ "embed"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed max_nested_calls.schema.json
var schemaJSON []byte

const (
	messageID  = "max-nested-calls"
	defaultMax = 3
)

var MaxNestedCallsRule = rule.Rule{
	Name:   "unicorn/max-nested-calls",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		max := parseMax(options)
		check := func(node *ast.Node) {
			if nestedCallDepth(node) <= max {
				return
			}
			maxText := strconv.Itoa(max)
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          messageID,
				Description: "Call is nested too deeply. Maximum allowed is " + maxText + ".",
				Data:        map[string]string{"max": maxText},
			})
		}
		return rule.RuleListeners{
			ast.KindCallExpression: check,
			ast.KindNewExpression:  check,
		}
	},
}

func parseMax(options []any) int {
	if len(options) == 0 {
		return defaultMax
	}
	config, _ := options[0].(map[string]any)
	if value, ok := utils.CoerceIntegral(config["max"]); ok {
		return value
	}
	return defaultMax
}

func nestedCallDepth(node *ast.Node) int {
	depth := 1
	child := node

	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if isNestedCallBoundary(ancestor) {
			return depth
		}
		if isCallOrNewExpression(ancestor) && hasArgument(ancestor, child) {
			depth++
		}
		child = ancestor
	}

	return depth
}

func isCallOrNewExpression(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression)
}

func hasArgument(node, target *ast.Node) bool {
	for _, argument := range node.Arguments() {
		if argument == target {
			return true
		}
	}
	return false
}

func isNestedCallBoundary(node *ast.Node) bool {
	if utils.IsFunctionLikeContainer(node) {
		return true
	}
	switch node.Kind {
	case ast.KindClassDeclaration,
		ast.KindClassExpression,
		ast.KindJsxElement,
		ast.KindJsxSelfClosingElement,
		ast.KindJsxFragment,
		ast.KindClassStaticBlockDeclaration:
		return true
	default:
		return false
	}
}
