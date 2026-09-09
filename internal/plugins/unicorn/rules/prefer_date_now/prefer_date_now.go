package prefer_date_now

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var preferDateMessage = rule.RuleMessage{
	Id:          "prefer-date",
	Description: "Prefer `Date.now()` over `new Date()`.",
}

func isNewDate(node *ast.Node) bool {
	if node == nil || !ast.IsNewExpression(node) || len(node.Arguments()) != 0 {
		return false
	}
	callee := utils.ESTreeRuntimeExpression(node.Expression())
	return callee != nil && ast.IsIdentifier(callee) && callee.Text() == "Date"
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-date-now.js
var PreferDateNowRule = rule.Rule{
	Name:   "unicorn/prefer-date-now",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		report := func(node, replacement *ast.Node, message rule.RuleMessage) {
			ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
				fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, replacement, "Date.now()")}
				if replacement.Kind == ast.KindPrefixUnaryExpression {
					fixes = append(fixes, unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, replacement)...)
				}
				return fixes
			})
		}
		reportDate := func(node *ast.Node) {
			node = utils.ESTreeRuntimeExpression(node)
			if isNewDate(node) {
				report(node, node, preferDateMessage)
			}
		}

		zeroArguments := 0
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if method, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods:         []string{"getTime", "valueOf"},
					ArgumentsLength: &zeroArguments,
				}); ok && isNewDate(utils.ESTreeRuntimeExpression(method.Object)) {
					methodName := method.Property.Text()
					report(method.Property, node, rule.RuleMessage{
						Id:          "prefer-date-now-over-methods",
						Description: "Prefer `Date.now()` over `Date#" + methodName + "()`.",
						Data:        map[string]string{"method": methodName},
					})
					return
				}

				if ast.IsOptionalChainRoot(node) || len(node.Arguments()) != 1 {
					return
				}
				callee := utils.ESTreeCallCallee(node.Expression())
				if callee == nil || !ast.IsIdentifier(callee) {
					return
				}
				date := utils.ESTreeRuntimeExpression(node.Arguments()[0])
				if !isNewDate(date) {
					return
				}
				switch callee.Text() {
				case "Number":
					report(node, node, rule.RuleMessage{
						// Preserve the upstream message ID, including "data".
						Id:          "prefer-date-now-over-number-data-object",
						Description: "Prefer `Date.now()` over `Number(new Date())`.",
					})
				case "BigInt":
					report(date, date, preferDateMessage)
				}
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				unary := node.AsPrefixUnaryExpression()
				if unary.Operator != ast.KindPlusToken && unary.Operator != ast.KindMinusToken {
					return
				}
				date := utils.ESTreeRuntimeExpression(unary.Operand)
				if !isNewDate(date) {
					return
				}
				if unary.Operator == ast.KindPlusToken {
					report(node, node, preferDateMessage)
				} else {
					report(date, date, preferDateMessage)
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				switch binary.OperatorToken.Kind {
				case ast.KindMinusToken, ast.KindAsteriskToken, ast.KindSlashToken,
					ast.KindPercentToken, ast.KindAsteriskAsteriskToken:
					reportDate(binary.Left)
					reportDate(binary.Right)
				case ast.KindMinusEqualsToken, ast.KindAsteriskEqualsToken, ast.KindSlashEqualsToken,
					ast.KindPercentEqualsToken, ast.KindAsteriskAsteriskEqualsToken:
					reportDate(binary.Right)
				}
			},
		}
	},
}
