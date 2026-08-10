package no_label_var

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-label-var
//
// ESLint's `getVariableByName` walks the scope chain all the way to the global
// scope. The same-file binding check and framework globals view cover those
// two layers without allowing TypeScript libraries to change the result:
//
//  1. utils.IsShadowed — fast, works without type info; covers every binding
//     declared inside the current source file (var/let/const, function, class,
//     enum, namespace, import, parameter, catch, for-init, function-expression
//     name, hoisted vars).
//  2. ctx.Globals — catches the selected ECMAScript edition plus config and
//     inline globals, including explicit `off` overrides.
var NoLabelVarRule = rule.Rule{
	Name:   "no-label-var",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		report := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "identifierClashWithLabel",
				Description: "Found identifier with same name as label.",
			})
		}

		return rule.RuleListeners{
			ast.KindLabeledStatement: func(node *ast.Node) {
				ls := node.AsLabeledStatement()
				if ls == nil || ls.Label == nil {
					return
				}
				name := ls.Label.Text()

				if utils.IsShadowed(node, name) {
					report(node)
					return
				}

				if ctx.Globals.Access(name).IsDeclared() {
					report(node)
				}
			},
		}
	},
}
