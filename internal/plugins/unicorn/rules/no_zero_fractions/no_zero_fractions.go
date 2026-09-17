// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_zero_fractions

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var fractionPattern = regexp.MustCompile(`^([\d_]*)(\.[\d_]*)(.*)$`)

var NoZeroFractionsRule = rule.Rule{
	Name:   "unicorn/no-zero-fractions",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				raw := utils.TrimmedNodeText(ctx.SourceFile, node)
				parts := fractionPattern.FindStringSubmatch(raw)
				if parts == nil {
					return
				}
				fraction := strings.TrimRight(parts[2], ".0_")
				formatted := parts[1] + fraction
				if formatted == "" {
					formatted = "0"
				}
				formatted += parts[3]
				if formatted == raw {
					return
				}
				message := rule.RuleMessage{Id: "zero-fraction", Description: "Don't use a zero fraction in the number."}
				if parts[2] == "." {
					message = rule.RuleMessage{Id: "dangling-dot", Description: "Don't use a dangling dot in the number."}
				}
				literalRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				end := literalRange.Pos() + len(parts[1]) + len(parts[2])
				diagnosticRange := core.NewTextRange(end-(len(raw)-len(formatted)), end)
				ctx.ReportRangeWithDeferredFixes(diagnosticRange, message, func() []rule.RuleFix {
					fixed := formatted
					parent := node.Parent
					if fraction == "" && parts[3] == "" && parent != nil &&
						(ast.IsPropertyAccessExpression(parent) || ast.IsElementAccessExpression(parent)) && parent.Expression() == node {
						fixed = "(" + fixed + ")"
						if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, fixed) {
							fixed = ";" + fixed
						}
					}
					return append([]rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, fixed)}, unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)...)
				})
			},
		}
	},
}
