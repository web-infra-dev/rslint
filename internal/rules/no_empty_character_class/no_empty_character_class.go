package no_empty_character_class

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-empty-character-class
//
// Implementation note: this rule rides on the shared regex character-class
// utilities in internal/utils. IterateRegexCharacterClasses fires once for
// every nesting level under the v flag, so an inner empty class is detected
// without any rule-side recursion.
//
// `[]` is empty (the rule fires); `[^]` is permitted (the leading `^` makes
// it match any character).
var NoEmptyCharacterClassRule = rule.Rule{
	Name: "no-empty-character-class",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				text := node.Text()
				pattern, flagsStr := utils.ExtractRegexPatternAndFlags(text)
				flags := utils.ParseRegexFlags(flagsStr)

				empty := false
				utils.IterateRegexCharacterClasses(pattern, flags, func(start, end int) {
					if empty {
						return
					}
					body := pattern[start+1 : end-1]
					if body == "" {
						empty = true
						return
					}
					if body == "^" {
						return
					}
				})

				if empty {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Empty class.",
					})
				}
			},
		}
	},
}
