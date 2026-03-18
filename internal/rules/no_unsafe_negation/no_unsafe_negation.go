package no_unsafe_negation

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-unsafe-negation
var NoUnsafeNegationRule = rule.Rule{
	Name: "no-unsafe-negation",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin == nil || bin.OperatorToken == nil {
					return
				}

				op := bin.OperatorToken.Kind

				// Check if operator is relational (always checked)
				isRelational := op == ast.KindInKeyword || op == ast.KindInstanceOfKeyword

				// Check if operator is ordering (only checked with option)
				isOrdering := op == ast.KindLessThanToken || op == ast.KindGreaterThanToken ||
					op == ast.KindLessThanEqualsToken || op == ast.KindGreaterThanEqualsToken

				if !isRelational && (!opts.enforceForOrderingRelations || !isOrdering) {
					return
				}

				left := bin.Left
				if left == nil || left.Kind != ast.KindPrefixUnaryExpression {
					return
				}

				prefix := left.AsPrefixUnaryExpression()
				if prefix == nil || prefix.Operator != ast.KindExclamationToken {
					return
				}

				// If left is directly a PrefixUnaryExpression (not wrapped in
				// ParenthesizedExpression), the negation is not parenthesized.
				// A parenthesized `(!a)` would make left.Kind == KindParenthesizedExpression,
				// so we would not reach this point.

				operatorText := getOperatorText(op)
				ctx.ReportNode(left, rule.RuleMessage{
					Id:          "unexpected",
					Description: fmt.Sprintf("Unexpected negating the left operand of '%s' operator.", operatorText),
				})
			},
		}
	},
}

// getOperatorText converts an operator Kind to its string representation
func getOperatorText(kind ast.Kind) string {
	switch kind {
	case ast.KindInKeyword:
		return "in"
	case ast.KindInstanceOfKeyword:
		return "instanceof"
	case ast.KindLessThanToken:
		return "<"
	case ast.KindGreaterThanToken:
		return ">"
	case ast.KindLessThanEqualsToken:
		return "<="
	case ast.KindGreaterThanEqualsToken:
		return ">="
	default:
		return ""
	}
}

type noUnsafeNegationOptions struct {
	enforceForOrderingRelations bool
}

func parseOptions(opts any) noUnsafeNegationOptions {
	result := noUnsafeNegationOptions{
		enforceForOrderingRelations: false,
	}

	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if enforce, ok := optsMap["enforceForOrderingRelations"].(bool); ok {
			result.enforceForOrderingRelations = enforce
		}
	}

	return result
}
