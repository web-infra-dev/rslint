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

func suggestStringMsg() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestString",
		Description: `Use "undefined" instead of undefined.`,
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
	return ast.GetBinaryOperatorPrecedence(kind) == ast.OperatorPrecedenceEquality
}

// https://eslint.org/docs/latest/rules/valid-typeof
var ValidTypeofRule = rule.Rule{
	Name: "valid-typeof",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindTypeOfExpression: func(node *ast.Node) {
				// Walk up through parenthesized expressions to find the enclosing
				// binary expression. The TS parser creates ParenthesizedExpression
				// nodes for parentheses, so we must skip them.
				parent := node.Parent
				for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
					parent = parent.Parent
				}
				if parent == nil || parent.Kind != ast.KindBinaryExpression {
					return
				}

				bin := parent.AsBinaryExpression()
				if bin == nil || bin.OperatorToken == nil {
					return
				}

				if !isEqualityOperator(bin.OperatorToken.Kind) {
					return
				}

				// Determine which operand is the sibling (not the typeof side).
				// Use SkipParentheses on both sides to match against the typeof node
				// since either side may be wrapped in parentheses.
				var sibling *ast.Node
				if ast.SkipParentheses(bin.Left) == node {
					sibling = ast.SkipParentheses(bin.Right)
				} else {
					sibling = ast.SkipParentheses(bin.Left)
				}

				if sibling == nil {
					return
				}

				switch {
				case ast.IsStringLiteralLike(sibling):
					// String literal or static template literal — check value.
					value := sibling.LiteralLikeData().Text
					if !validTypes[value] {
						ctx.ReportNode(sibling, invalidValueMsg())
					}

				case sibling.Kind == ast.KindNumericLiteral,
					sibling.Kind == ast.KindBigIntLiteral,
					sibling.Kind == ast.KindRegularExpressionLiteral,
					ast.IsBooleanLiteral(sibling),
					utils.IsNullLiteral(sibling):
					// Non-string literals can never be valid typeof results.
					ctx.ReportNode(sibling, invalidValueMsg())

				case sibling.Kind == ast.KindIdentifier:
					if sibling.Text() == "undefined" && !utils.IsShadowed(sibling, "undefined") {
						// Bare `undefined` referencing the global variable:
						// report with suggestion to use "undefined" string.
						msg := invalidValueMsg()
						if opts.requireStringLiterals {
							msg = notStringMsg()
						}
						ctx.ReportNodeWithSuggestions(sibling, msg, rule.RuleSuggestion{
							Message: suggestStringMsg(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplace(ctx.SourceFile, sibling, `"undefined"`),
							},
						})
					} else if opts.requireStringLiterals {
						// Any other identifier (including shadowed `undefined`)
						// is not a string literal.
						ctx.ReportNode(sibling, notStringMsg())
					}

				case sibling.Kind == ast.KindTypeOfExpression:
					// typeof === typeof is always valid

				default:
					if opts.requireStringLiterals {
						ctx.ReportNode(sibling, notStringMsg())
					}
				}
			},
		}
	},
}
