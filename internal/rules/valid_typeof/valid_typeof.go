package valid_typeof

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// validTypes is the set of strings that are valid results of the typeof operator.
var validTypes = map[string]bool{
	"undefined": true,
	"object":    true,
	"boolean":   true,
	"number":    true,
	"string":    true,
	"function":  true,
	"symbol":    true,
	"bigint":    true,
}

func invalidValueMsg() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidValue",
		Description: "Invalid typeof comparison value.",
	}
}

func notStringMsg() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notString",
		Description: "Typeof comparisons should be to string literals.",
	}
}

type validTypeofOptions struct {
	requireStringLiterals bool
}

func parseOptions(opts any) validTypeofOptions {
	result := validTypeofOptions{
		requireStringLiterals: false,
	}

	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if req, ok := optsMap["requireStringLiterals"].(bool); ok {
			result.requireStringLiterals = req
		}
	}

	return result
}

// isEqualityOperator checks if the operator kind is ==, ===, !=, or !==.
func isEqualityOperator(kind ast.Kind) bool {
	return kind == ast.KindEqualsEqualsToken ||
		kind == ast.KindEqualsEqualsEqualsToken ||
		kind == ast.KindExclamationEqualsToken ||
		kind == ast.KindExclamationEqualsEqualsToken
}

// https://eslint.org/docs/latest/rules/valid-typeof
var ValidTypeofRule = rule.Rule{
	Name: "valid-typeof",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindTypeOfExpression: func(node *ast.Node) {
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindBinaryExpression {
					return
				}

				bin := parent.AsBinaryExpression()
				if bin == nil || bin.OperatorToken == nil {
					return
				}

				// Only check equality/inequality operators
				if !isEqualityOperator(bin.OperatorToken.Kind) {
					return
				}

				// Get the sibling operand (the one that is not the typeof expression)
				var sibling *ast.Node
				if bin.Left == node {
					sibling = bin.Right
				} else {
					sibling = bin.Left
				}

				if sibling == nil {
					return
				}

				switch sibling.Kind {
				case ast.KindStringLiteral:
					// Check if the string value is a valid typeof result
					// Use AsStringLiteral().Text to get unquoted value (not .Text() which includes quotes)
					value := sibling.AsStringLiteral().Text
					if !validTypes[value] {
						ctx.ReportNode(sibling, invalidValueMsg())
					}

				case ast.KindIdentifier:
					// Bare `undefined` identifier is always invalid in typeof comparisons.
					// With requireStringLiterals, report as "notString";
					// without, report as "invalidValue" (since typeof never actually
					// returns the undefined *value*, only the string "undefined").
					if sibling.Text() == "undefined" {
						if opts.requireStringLiterals {
							ctx.ReportNode(sibling, notStringMsg())
						} else {
							ctx.ReportNode(sibling, invalidValueMsg())
						}
					} else if opts.requireStringLiterals {
						// Any non-undefined identifier is not a string literal
						ctx.ReportNode(sibling, notStringMsg())
					}

				case ast.KindTypeOfExpression:
					// typeof === typeof is always valid

				default:
					// For any other expression (template literals, variables, etc.)
					if opts.requireStringLiterals {
						ctx.ReportNode(sibling, notStringMsg())
					}
				}
			},
		}
	},
}
