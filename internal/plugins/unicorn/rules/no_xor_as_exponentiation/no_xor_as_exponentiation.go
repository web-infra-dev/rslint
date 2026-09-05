package no_xor_as_exponentiation

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const (
	messageIDError      = "no-xor-as-exponentiation/error"
	messageIDSuggestion = "no-xor-as-exponentiation/suggestion"
)

var messageError = rule.RuleMessage{
	Id:          messageIDError,
	Description: "Unexpected bitwise XOR operator `^`. Did you mean the exponentiation operator `**`?",
}

var messageSuggestion = rule.RuleMessage{
	Id:          messageIDSuggestion,
	Description: "Replace `^` with `**`.",
}

// decimalIntegerPattern mirrors the upstream
// eslint-plugin-unicorn `isDecimalInteger` helper. It accepts:
//   - `0`
//   - legacy octal that happens to be a decimal integer shape (`0[0-7]*[89]\d*`)
//   - a non-zero digit followed by digits with optional numeric separators
var decimalIntegerPattern = regexp.MustCompile(`^(?:0|0[0-7]*[89]\d*|[1-9](?:_?\d)*)$`)

// isDecimalIntegerNumeric reports whether the given node is a numeric literal
// whose raw source text matches the upstream decimal-integer pattern. Mirrors
// `isDecimalIntegerNode` in the upstream plugin. Uses the raw source text
// (e.g. `0xFF` stays `0xFF`) rather than tsgo's decimal-normalized `Text`
// field, since the upstream regex is anchored to a literal-shape classification
// that the normalized value would silently fold into a decimal integer.
func isDecimalIntegerNumeric(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindNumericLiteral {
		return false
	}
	raw := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, node, false)
	return decimalIntegerPattern.MatchString(raw)
}

// NoXorAsExponentiationRule flags `^` between two decimal integer literals. In
// JavaScript `^` is bitwise XOR, not exponentiation; the rule mirrors the
// behavior of eslint-plugin-unicorn's `no-xor-as-exponentiation`. It attaches a
// single suggestion that rewrites the `^` token as `**`.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/no-xor-as-exponentiation.md
var NoXorAsExponentiationRule = rule.Rule{
	Name:   "unicorn/no-xor-as-exponentiation",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil ||
					binary.OperatorToken.Kind != ast.KindCaretToken {
					return
				}

				left := ast.SkipParentheses(binary.Left)
				right := ast.SkipParentheses(binary.Right)
				if !isDecimalIntegerNumeric(ctx, left) || !isDecimalIntegerNumeric(ctx, right) {
					return
				}

				operator := binary.OperatorToken
				ctx.ReportNodeWithDeferredSuggestions(operator, messageError, func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{{
						Message: messageSuggestion,
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, operator, "**"),
						},
					}}
				})
			},
		}
	},
}
