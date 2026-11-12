// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_accidental_bitwise_operator

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoAccidentalBitwiseOperatorRule = rule.Rule{
	Name:   "unicorn/no-accidental-bitwise-operator",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{ast.KindBinaryExpression: func(node *ast.Node) {
			binary := node.AsBinaryExpression()
			operator := binary.OperatorToken
			var replacement ast.Kind
			switch operator.Kind {
			case ast.KindAmpersandToken:
				if !isShortCircuitGuard(binary.Left, binary.Right) {
					return
				}
				replacement = ast.KindAmpersandAmpersandToken
			case ast.KindBarToken, ast.KindBarEqualsToken:
				if !isFallbackOperand(binary.Right) {
					return
				}
				replacement = ast.KindBarBarToken
				if operator.Kind == ast.KindBarEqualsToken {
					replacement = ast.KindBarBarEqualsToken
				}
			default:
				return
			}

			bitwise := scanner.TokenToString(operator.Kind)
			logical := scanner.TokenToString(replacement)
			data := map[string]string{"bitwiseOperator": bitwise, "logicalOperator": logical}
			span := operator.Loc.WithPos(scanner.GetTokenPosOfNode(operator, ctx.SourceFile, false))
			ctx.ReportRangeWithDeferredSuggestions(span, rule.RuleMessage{
				Id:          "no-accidental-bitwise-operator/error",
				Description: "Unexpected bitwise operator `" + bitwise + "`. Did you mean the logical operator `" + logical + "`?",
				Data:        data,
			}, func() []rule.RuleSuggestion {
				fix := rule.RuleFixReplaceRange(span, logical)
				var fixes []rule.RuleFix
				if needsParentheses(node, replacement) {
					prefix := "("
					if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, prefix) {
						prefix = ";("
					}
					fixes = []rule.RuleFix{
						rule.RuleFixInsertBefore(ctx.SourceFile, node, prefix),
						fix,
						rule.RuleFixInsertAfter(node, ")"),
					}
				} else {
					fixes = []rule.RuleFix{fix}
				}
				return []rule.RuleSuggestion{{
					Message: rule.RuleMessage{
						Id:          "no-accidental-bitwise-operator/suggestion",
						Description: "Replace `" + bitwise + "` with `" + logical + "`.",
						Data:        data,
					},
					FixesArr: fixes,
				}}
			})
		}}
	},
}

func isShortCircuitGuard(left, right *ast.Node) bool {
	left = utils.ESTreeRuntimeExpression(left)
	if left == nil || !ast.IsIdentifier(left) {
		return false
	}
	right = utils.ESTreeRuntimeExpression(right)
	if right == nil || !ast.IsAccessExpression(right) || ast.IsOptionalChain(right) {
		return false
	}
	receiver := utils.ESTreeRuntimeExpression(right.Expression())
	return receiver != nil && ast.IsIdentifier(receiver) && receiver.Text() == left.Text()
}

func isFallbackOperand(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression, ast.KindClassExpression,
		ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	}
	return false
}

func needsParentheses(node *ast.Node, replacement ast.Kind) bool {
	// Use the immediate parent so existing parentheses are preserved.
	parent := node.Parent
	if parent == nil || !ast.IsBinaryExpression(parent) {
		return false
	}
	operator := parent.AsBinaryExpression().OperatorToken.Kind
	// JavaScript requires grouping when mixing && or || with ??.
	return operator == ast.KindQuestionQuestionToken ||
		ast.GetOperatorPrecedence(parent.Kind, operator, ast.OperatorPrecedenceFlagsNone) >
			ast.GetOperatorPrecedence(node.Kind, replacement, ast.OperatorPrecedenceFlagsNone)
}
