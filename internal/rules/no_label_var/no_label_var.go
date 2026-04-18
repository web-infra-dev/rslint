package no_label_var

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-label-var
//
// Strategy: hybrid. ESLint's `getVariableByName` walks the scope chain all the
// way to the global scope, catching both file-local declarations and globals
// from `env`/`globals`. We approximate it with two complementary checks:
//
//  1. utils.IsShadowed — fast, works without type info; covers every binding
//     declared inside the current source file (var/let/const, function, class,
//     enum, namespace, import, parameter, catch, for-init, function-expression
//     name, hoisted vars).
//  2. ctx.TypeChecker.GetSymbolsInScope — when type info is available, also
//     catches globals provided by the tsconfig `lib` (e.g. `window`, `Promise`,
//     `console`). Differs from ESLint's `env`/`globals` config and does not
//     read `/* global foo */` comments — see the rule docs for details.
//
// On a JS file with no tsconfig only step 1 runs, so the rule still catches
// the dominant case (label clashing with a sibling declaration).
var NoLabelVarRule = rule.Rule{
	Name: "no-label-var",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
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

				if ctx.TypeChecker == nil {
					return
				}
				for _, sym := range ctx.TypeChecker.GetSymbolsInScope(node, ast.SymbolFlagsValue) {
					if sym != nil && sym.Name == name {
						report(node)
						return
					}
				}
			},
		}
	},
}
