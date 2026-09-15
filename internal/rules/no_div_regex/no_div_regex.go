package no_div_regex

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-div-regex
var NoDivRegexRule = rule.Rule{
	Name:   "no-div-regex",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				text := node.Text()
				if len(text) < 2 || text[1] != '=' {
					return
				}

				ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: "A regular expression literal can be confused with '/='.",
				}, func() []rule.RuleFix {
					eqPos := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos() + 1
					return []rule.RuleFix{
						rule.RuleFixReplaceRange(core.NewTextRange(eqPos, eqPos+1), "[=]"),
					}
				})
			},
		}
	},
}
