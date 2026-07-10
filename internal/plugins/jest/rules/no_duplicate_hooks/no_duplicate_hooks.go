package no_duplicate_hooks

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builder

func buildNoDuplicateHookMessage(hook string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noDuplicateHook",
		Description: "Duplicate " + hook + " in describe block",
		Data: map[string]string{
			"hook": hook,
		},
	}
}

var NoDuplicateHooksRule = rule.Rule{
	Name: "jest/no-duplicate-hooks",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		hookContexts := []map[string]int{{}}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}

				if jestFnCall.Kind == utils.JestFnTypeDescribe {
					hookContexts = append(hookContexts, map[string]int{})
					return
				}

				if jestFnCall.Kind != utils.JestFnTypeHook {
					return
				}

				currentLayer := hookContexts[len(hookContexts)-1]
				currentLayer[jestFnCall.Name]++
				if currentLayer[jestFnCall.Name] > 1 {
					ctx.ReportNode(node, buildNoDuplicateHookMessage(jestFnCall.Name))
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if !utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeDescribe) {
					return
				}
				if len(hookContexts) > 1 {
					hookContexts = hookContexts[:len(hookContexts)-1]
				}
			},
		}
	},
}
