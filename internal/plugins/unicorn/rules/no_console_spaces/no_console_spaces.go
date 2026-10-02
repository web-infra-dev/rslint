// Ported from eslint-plugin-unicorn v76.0.0; see LICENSE.
package no_console_spaces

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoConsoleSpacesRule = rule.Rule{
	Name:   "unicorn/no-console-spaces",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		report := func(index int, method, position string) {
			textRange := core.NewTextRange(index, index+1)
			ctx.ReportRangeWithDeferredFixes(textRange, rule.RuleMessage{
				Id:          "no-console-spaces",
				Description: "Do not use " + position + " space between `console." + method + "` parameters.",
			}, func() []rule.RuleFix {
				// Removing only an escaped space would escape the closing quote
				// or backtick. Remove its escape too, preserving paired backslashes.
				prefix := ctx.SourceFile.Text()[:index]
				backslashes := len(prefix) - len(strings.TrimRight(prefix, "\\"))
				return []rule.RuleFix{rule.RuleFixRemoveRange(textRange.WithPos(index - backslashes%2))}
			})
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// A single argument has no boundary with another console parameter.
				minimumArguments := 2
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods:          []string{"log", "debug", "info", "warn", "error"},
					MinimumArguments: &minimumArguments,
				})
				if !ok {
					return
				}
				object := utils.ESTreeRuntimeExpression(call.Object)
				if !ast.IsIdentifier(object) || object.Text() != "console" {
					return
				}

				args := node.Arguments()
				for index, arg := range args {
					arg = utils.ESTreeRuntimeExpression(arg)
					if !ast.IsStringLiteralLike(arg) && !ast.IsTemplateExpression(arg) {
						continue
					}
					// The AST already gives the end; skip leading trivia without
					// scanning and decoding the entire literal again.
					start, end := scanner.GetTokenPosOfNode(arg, ctx.SourceFile, false), arg.End()
					if end-start < 4 {
						continue
					}
					// Upstream checks raw source, not the decoded value. Only ASCII
					// spaces next to the quotes/backticks matter, so byte checks also
					// preserve its behavior beside escapes and non-ASCII characters.
					raw := ctx.SourceFile.Text()[start+1 : end-1]
					method := call.Property.Text()
					if index != 0 && raw[0] == ' ' && raw[1] != ' ' {
						report(start+1, method, "leading")
					}
					if index != len(args)-1 && raw[len(raw)-1] == ' ' && raw[len(raw)-2] != ' ' {
						report(end-2, method, "trailing")
					}
				}
			},
		}
	},
}
