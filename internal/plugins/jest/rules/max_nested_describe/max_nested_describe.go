package max_nested_describe

import (
	_ "embed"
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed max_nested_describe.schema.json
var schemaJSON []byte

const defaultMax = 5

type options struct {
	Max int
}

func parseOptions(rawOptions []any) options {
	opts := options{Max: defaultMax}
	if len(rawOptions) == 0 {
		return opts
	}

	optsMap, _ := rawOptions[0].(map[string]interface{})
	if maxVal, ok := internalUtils.CoerceIntegral(optsMap["max"]); ok && maxVal >= 0 {
		opts.Max = maxVal
	}

	return opts
}

func buildErrorExceededMaxDepthMessage(depth, maxAllowed int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "exceededMaxDepth",
		Description: fmt.Sprintf("Too many nested describe calls (%d) - maximum allowed is %d", depth, maxAllowed),
		Data: map[string]string{
			"depth": strconv.Itoa(depth),
			"max":   strconv.Itoa(maxAllowed),
		},
	}
}

var MaxNestedDescribeRule = rule.Rule{
	Name:   "jest/max-nested-describe",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		describes := make([]*ast.Node, 0, 8)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if !utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeDescribe) {
					return
				}

				describes = append(describes, node)
				if len(describes) > opts.Max {
					ctx.ReportNode(node, buildErrorExceededMaxDepthMessage(len(describes), opts.Max))
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if !utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeDescribe) {
					return
				}
				if len(describes) == 0 {
					return
				}
				if describes[len(describes)-1] == node {
					describes = describes[:len(describes)-1]
				}
			},
		}
	},
}
