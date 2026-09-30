package no_accidental_bitwise_operator

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	errorMessageID      = "no-accidental-bitwise-operator/error"
	suggestionMessageID = "no-accidental-bitwise-operator/suggestion"
)

func logicalOperator(kind ast.Kind) (string, string, bool) {
	switch kind {
	case ast.KindAmpersandToken:
		return "&", "&&", true
	case ast.KindBarToken:
		return "|", "||", true
	case ast.KindBarEqualsToken:
		return "|=", "||=", true
	}
	return "", "", false
}

func isDefinitelyNonNumeric(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindClassExpression,
		ast.KindTemplateExpression,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindStringLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword:
		return true
	}
	return false
}

func isShortCircuitGuard(binary *ast.BinaryExpression) bool {
	if binary.OperatorToken.Kind != ast.KindAmpersandToken {
		return false
	}

	left := utils.ESTreeRuntimeExpression(binary.Left)
	right := utils.ESTreeRuntimeExpression(binary.Right)
	if left == nil || right == nil || left.Kind != ast.KindIdentifier || ast.IsOptionalChain(right) {
		return false
	}

	var object *ast.Node
	switch right.Kind {
	case ast.KindPropertyAccessExpression:
		object = right.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		object = right.AsElementAccessExpression().Expression
	default:
		return false
	}
	object = utils.ESTreeRuntimeExpression(object)
	return object != nil &&
		object.Kind == ast.KindIdentifier &&
		object.AsIdentifier().Text == left.AsIdentifier().Text
}

func isSuspicious(binary *ast.BinaryExpression) bool {
	if isShortCircuitGuard(binary) {
		return true
	}
	return (binary.OperatorToken.Kind == ast.KindBarToken ||
		binary.OperatorToken.Kind == ast.KindBarEqualsToken) &&
		isDefinitelyNonNumeric(binary.Right)
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-accidental-bitwise-operator.js
var NoAccidentalBitwiseOperatorRule = rule.Rule{
	Name:   "unicorn/no-accidental-bitwise-operator",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				bitwise, logical, supported := logicalOperator(binary.OperatorToken.Kind)
				if !supported || !isSuspicious(binary) {
					return
				}

				operatorRange := utils.TrimNodeTextRange(ctx.SourceFile, binary.OperatorToken)
				message := rule.RuleMessage{
					Id:          errorMessageID,
					Description: "Unexpected bitwise operator `" + bitwise + "`. Did you mean the logical operator `" + logical + "`?",
					Data: map[string]string{
						"bitwiseOperator": bitwise,
						"logicalOperator": logical,
					},
				}
				ctx.ReportRangeWithDeferredSuggestions(operatorRange, message, func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{{
						Message: rule.RuleMessage{
							Id:          suggestionMessageID,
							Description: "Replace `" + bitwise + "` with `" + logical + "`.",
							Data: map[string]string{
								"bitwiseOperator": bitwise,
								"logicalOperator": logical,
							},
						},
						FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(operatorRange, logical)},
					}}
				})
			},
		}
	},
}
