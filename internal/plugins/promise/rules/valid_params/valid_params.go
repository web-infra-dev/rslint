package valid_params

import (
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const skipTransparent = ast.OEKParentheses

type Options struct {
	Exclude map[string]bool
}

func buildRequireOneOptionalArgumentMessage(name string, numArgs int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requireOneOptionalArgument",
		Description: fmt.Sprintf("Promise.%s() requires 0 or 1 arguments, but received %d", name, numArgs),
		Data: map[string]string{
			"name":    name,
			"numArgs": strconv.Itoa(numArgs),
		},
	}
}

func buildRequireOneArgumentMessage(name string, numArgs int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requireOneArgument",
		Description: fmt.Sprintf("Promise.%s() requires 1 argument, but received %d", name, numArgs),
		Data: map[string]string{
			"name":    name,
			"numArgs": strconv.Itoa(numArgs),
		},
	}
}

func buildRequireTwoOptionalArgumentsMessage(name string, numArgs int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requireTwoOptionalArguments",
		Description: fmt.Sprintf("Promise.%s() requires 1 or 2 arguments, but received %d", name, numArgs),
		Data: map[string]string{
			"name":    name,
			"numArgs": strconv.Itoa(numArgs),
		},
	}
}

func calledMemberName(node *ast.Node) string {
	if node == nil || !ast.IsCallExpression(node) {
		return ""
	}
	callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) {
		return ""
	}
	name := callee.AsPropertyAccessExpression().Name()
	if name == nil || !ast.IsIdentifier(name) {
		return ""
	}
	return name.AsIdentifier().Text
}

var ValidParamsRule = rule.Rule{
	Name: "promise/valid-params",
	Schema: rule.Tuple(rule.Object(map[string]rule.Schema{
		"exclude": rule.Array(rule.String()),
	})),
	RunWithOptions: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		optsMap, _ := options[0].(map[string]any)
		excludeArr, _ := optsMap["exclude"].([]any)
		excludeMap := make(map[string]bool, len(excludeArr))
		for _, item := range excludeArr {
			s, _ := item.(string)
			excludeMap[s] = true
		}
		opts := Options{
			Exclude: excludeMap,
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if !promiseutil.IsPromiseLikeCall(node) {
					return
				}

				name := calledMemberName(node)
				if name == "" || opts.Exclude[name] {
					return
				}
				numArgs := len(node.AsCallExpression().Arguments.Nodes)

				switch name {
				case "resolve", "reject":
					if numArgs > 1 {
						ctx.ReportNode(node, buildRequireOneOptionalArgumentMessage(name, numArgs))
					}
				case "then":
					if numArgs < 1 || numArgs > 2 {
						ctx.ReportNode(node, buildRequireTwoOptionalArgumentsMessage(name, numArgs))
					}
				case "race", "all", "allSettled", "any", "catch", "finally":
					if numArgs != 1 {
						ctx.ReportNode(node, buildRequireOneArgumentMessage(name, numArgs))
					}
				}
			},
		}
	},
}
