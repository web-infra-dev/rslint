package no_focused_tests

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var NoFocusedTestsRule = rule.Rule{
	Name: "jest/no-focused-tests",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					(jestFnCall.Kind != utils.JestFnTypeDescribe &&
						jestFnCall.Kind != utils.JestFnTypeTest) {
					return
				}
			},
		}
	},
}
