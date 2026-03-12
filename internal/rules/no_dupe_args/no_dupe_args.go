package no_dupe_args

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-dupe-args
var NoDupeArgsRule = rule.Rule{
	Name: "no-dupe-args",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// NOTE: Unlike ESLint which reports once per duplicated name on the function node,
		// we report on each individual duplicate occurrence for more precise diagnostics.
		checkParams := func(node *ast.Node) {
			params := node.Parameters()
			if params == nil {
				return
			}

			seen := make(map[string]bool)
			for _, param := range params {
				paramDecl := param.AsParameterDeclaration()
				if paramDecl == nil {
					continue
				}
				nameNode := param.Name()
				if nameNode == nil {
					continue
				}
				utils.CollectBindingNames(nameNode, func(ident *ast.Node, name string) {
					if seen[name] {
						ctx.ReportNode(ident, rule.RuleMessage{
							Id:          "unexpected",
							Description: fmt.Sprintf("Duplicate param '%s'.", name),
						})
					} else {
						seen[name] = true
					}
				})
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkParams,
			ast.KindFunctionExpression:  checkParams,
		}
	},
}
