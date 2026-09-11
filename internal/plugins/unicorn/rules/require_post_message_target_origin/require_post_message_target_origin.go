// Ported from eslint-plugin-unicorn v74.0.0 (MIT).
package require_post_message_target_origin

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/require-post-message-target-origin.js
var RequirePostMessageTargetOriginRule = rule.Rule{
	Name:   "unicorn/require-post-message-target-origin",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		oneArgument := 1
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "postMessage",
					ArgumentsLength:     &oneArgument,
					RejectSpreadElement: true,
					AllowOptionalMember: true,
				})
				if !ok {
					return
				}

				// Scan only the suffix after the argument. This preserves comments
				// and parentheses without re-scanning regex or template literals.
				start := node.Arguments()[0].End()
				s := scanner.GetScannerForSourceFile(ctx.SourceFile, start)
				trailingComma := s.Token() == ast.KindCommaToken
				if trailingComma {
					start = s.TokenEnd()
					s.Scan()
				}
				if s.Token() != ast.KindCloseParenToken || s.TokenEnd() != node.End() {
					return
				}
				closing := s.TokenRange()
				ctx.ReportRangeWithDeferredSuggestions(core.NewTextRange(start, closing.End()), rule.RuleMessage{
					Id:          "error",
					Description: "Missing the `targetOrigin` argument.",
				}, func() []rule.RuleSuggestion {
					replacements := []string{}
					target := utils.ESTreeRuntimeExpression(call.Object)
					if ast.IsIdentifier(target) {
						name := target.AsIdentifier().Text
						replacements = append(replacements, name+".location.origin")
						if name != "self" && name != "window" && name != "globalThis" {
							replacements = append(replacements, "self.location.origin")
						}
					} else {
						replacements = append(replacements, "self.location.origin")
					}
					replacements = append(replacements, "'*'")

					suggestions := make([]rule.RuleSuggestion, 0, len(replacements))
					for _, replacement := range replacements {
						text := ", " + replacement
						if trailingComma {
							text = " " + replacement + ","
						}
						suggestions = append(suggestions, rule.RuleSuggestion{
							Message: rule.RuleMessage{Id: "suggestion", Description: "Use `" + replacement + "`."},
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(core.NewTextRange(closing.Pos(), closing.Pos()), text),
							},
						})
					}
					return suggestions
				})
			},
		}
	},
}
