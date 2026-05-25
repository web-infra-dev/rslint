package no_new_symbol

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-new-symbol
var NoNewSymbolRule = rule.Rule{
	Name: "no-new-symbol",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				expr := node.Expression()
				if expr == nil || expr.Kind != ast.KindIdentifier || expr.Text() != "Symbol" {
					return
				}

				// Use scope-based shadowing analysis (not TypeChecker) because
				// TypeChecker resolves `new Symbol()` to the global SymbolConstructor
				// even when Symbol is locally shadowed by let/const/function/class,
				// due to declaration merging with lib.d.ts types. IsShadowed
				// correctly mirrors ESLint's scope-based behavior.
				if !utils.IsShadowed(expr, "Symbol") {
					ctx.ReportNode(expr, rule.RuleMessage{
						Id:          "noNewSymbol",
						Description: "`Symbol` cannot be called as a constructor.",
					})
				}
			},
		}
	},
}
