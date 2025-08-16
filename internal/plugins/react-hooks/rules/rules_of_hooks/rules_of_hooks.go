package rules_of_hooks

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var RulesOfHooksRule = rule.Rule{
	Name: "react-hooks/rules-of-hooks",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return utils.VisitModules(func(source, node *ast.Node) {

		}, utils.VisitModulesOptions{
			Commonjs: true,
			ESModule: true,
		})
	},
}
