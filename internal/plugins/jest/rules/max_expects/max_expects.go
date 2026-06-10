package max_expects

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

	optArray, ok := raw.([]interface{})
	if !ok || len(optArray) == 0 {
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

	if opts.Max < 1 {
		return options{Max: defaultMax}
	}

	return opts
}

func buildExceededMaxAssertionMessage(count, maxAllowed int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "exceededMaxAssertion",
		Description: fmt.Sprintf("Too many assertion calls (%d) - maximum allowed is %d", count, maxAllowed),
		Data: map[string]string{
			"count": strconv.Itoa(count),
			"max":   strconv.Itoa(maxAllowed),
		},
	}
}

func isTestCallbackFunction(fn *ast.Node, ctx rule.RuleContext) bool {
	parent := fn.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return true
	}
	return utils.IsTypeOfJestFnCall(parent, ctx, utils.JestFnTypeTest)
}

func shouldCountExpectCall(jestFnCall *utils.ParsedJestFnCall) bool {
	if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeExpect {
		return false
	}

	headNode := jestFnCall.Head.Local.Node
	if headNode != nil && utils.IsMemberAccessNode(headNode.Parent) {
		return false
	}

	return true
}

var MaxExpectsRule = rule.Rule{
	Name: "jest/max-expects",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		count := 0

		maybeResetCount := func(node *ast.Node) {
			if isTestCallbackFunction(node, ctx) {
				count = 0
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionExpression:                      maybeResetCount,
			rule.ListenerOnExit(ast.KindFunctionExpression): maybeResetCount,
			ast.KindArrowFunction:                           maybeResetCount,
			rule.ListenerOnExit(ast.KindArrowFunction):      maybeResetCount,
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}

				if jestFnCall.Kind == utils.JestFnTypeTest {
					count = 0
					return
				}

				if !shouldCountExpectCall(jestFnCall) {
					return
				}

				count++
				if count > opts.Max {
					ctx.ReportNode(node, buildExceededMaxAssertionMessage(count, opts.Max))
				}
			},
		}
	},
}
