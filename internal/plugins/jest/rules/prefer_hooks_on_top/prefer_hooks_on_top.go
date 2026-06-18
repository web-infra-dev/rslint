package prefer_hooks_on_top

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNoHookOnTopMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noHookOnTop",
		Description: "Hooks should come before test cases",
	}
}

var PreferHooksOnTopRule = rule.Rule{
	Name: "jest/prefer-hooks-on-top",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		hooksContext := []bool{false}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeTest) {
					hooksContext[len(hooksContext)-1] = true
				}
				if hooksContext[len(hooksContext)-1] && utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeHook) {
					ctx.ReportNode(node, buildNoHookOnTopMessage())
				}
				hooksContext = append(hooksContext, false)
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				hooksContext = hooksContext[:len(hooksContext)-1]
			},
		}
	},
}
