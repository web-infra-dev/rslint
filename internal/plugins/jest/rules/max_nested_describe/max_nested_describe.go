package max_nested_describe

import (
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const defaultMax = 5

type options struct {
	Max int
}

func parseOptions(raw any) options {
	opts := options{Max: defaultMax}
	if raw == nil {
		return opts
	}

	optArray := rule.NormalizeOptions(raw)
	if len(optArray) == 0 {
		return opts
	}

	optsMap, ok := optArray[0].(map[string]interface{})
	if !ok {
		return opts
	}

	switch v := optsMap["max"].(type) {
	case float64:
		if v == float64(int(v)) {
			opts.Max = int(v)
		}
	case int:
		opts.Max = v
	case int64:
		opts.Max = int(v)
	}

	if opts.Max < 0 {
		return options{Max: defaultMax}
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
	Name: "jest/max-nested-describe",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
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
