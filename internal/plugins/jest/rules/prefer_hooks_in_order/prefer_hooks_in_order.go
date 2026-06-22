package prefer_hooks_in_order

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildReorderHooksMessage(currentHook, previousHook string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "reorderHooks",
		Description: "`" + currentHook + "` hooks should be before any `" + previousHook + "` hooks",
		Data: map[string]string{
			"currentHook":  currentHook,
			"previousHook": previousHook,
		},
	}
}

var PreferHooksInOrderRule = rule.Rule{
	Name: "jest/prefer-hooks-in-order",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		previousHookIndex := -1
		inHook := false

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if inHook {
					return
				}

				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeHook {
					previousHookIndex = -1
					return
				}

				inHook = true
				currentHook := jestFnCall.Name
				currentHookIndex := utils.JestHookOrderIndex(currentHook)

				if currentHookIndex < previousHookIndex {
					ctx.ReportNode(node, buildReorderHooksMessage(currentHook, utils.JEST_HOOKS_ORDER[previousHookIndex]))
					return
				}

				previousHookIndex = currentHookIndex
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeHook) {
					inHook = false
					return
				}

				if inHook {
					return
				}

				previousHookIndex = -1
			},
		}
	},
}
