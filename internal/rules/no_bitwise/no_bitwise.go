package no_bitwise

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// bitwiseOperatorKinds is the set of AST token kinds that represent bitwise
// operators (both binary forms like `|`, `<<` and their assignment variants,
// plus the unary `~`). Membership is all the rule needs — the string form
// of the operator is produced on demand via scanner.TokenToString.
var bitwiseOperatorKinds = map[ast.Kind]bool{
	ast.KindBarToken:                                     true, // |
	ast.KindAmpersandToken:                               true, // &
	ast.KindCaretToken:                                   true, // ^
	ast.KindLessThanLessThanToken:                        true, // <<
	ast.KindGreaterThanGreaterThanToken:                  true, // >>
	ast.KindGreaterThanGreaterThanGreaterThanToken:       true, // >>>
	ast.KindBarEqualsToken:                               true, // |=
	ast.KindAmpersandEqualsToken:                         true, // &=
	ast.KindCaretEqualsToken:                             true, // ^=
	ast.KindLessThanLessThanEqualsToken:                  true, // <<=
	ast.KindGreaterThanGreaterThanEqualsToken:            true, // >>=
	ast.KindGreaterThanGreaterThanGreaterThanEqualsToken: true, // >>>=
	ast.KindTildeToken:                                   true, // ~
}

type noBitwiseOptions struct {
	allow     map[string]bool
	int32Hint bool
}

func parseOptions(opts any) noBitwiseOptions {
	result := noBitwiseOptions{
		allow:     map[string]bool{},
		int32Hint: false,
	}

	optsMap := utils.GetOptionsMap(opts)
	if optsMap == nil {
		return result
	}

	if raw, ok := optsMap["allow"].([]interface{}); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				result.allow[s] = true
			}
		}
	}

	if b, ok := optsMap["int32Hint"].(bool); ok {
		result.int32Hint = b
	}

	return result
}

// isZeroNumericLiteral reports whether the node — after unwrapping parentheses
// — is a numeric literal whose value is 0. Used by the int32Hint branch to
// detect the `x | 0` typecasting idiom.
//
// Parens are unwrapped because ESTree (ESLint's AST) is transparent to them;
// `x | (0)` has `right.type === 'Literal'` in ESLint, so we must match.
//
// TS-go's parser stores NumericLiteral.Text in decimal form (e.g. `0x0`,
// `0b0`, `0o0`, `0.0`, `0e0` all surface as "0"), so a plain text compare
// suffices. BigInt literals (`0n`) are a distinct AST kind and remain
// excluded, matching ESLint (where `0n !== 0`).
func isZeroNumericLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	unwrapped := ast.SkipParentheses(node)
	if !ast.IsNumericLiteral(unwrapped) {
		return false
	}
	return unwrapped.AsNumericLiteral().Text == "0"
}

func buildUnexpectedMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected use of '" + operator + "'.",
	}
}

// NoBitwiseRule disallows bitwise operators.
// https://eslint.org/docs/latest/rules/no-bitwise
var NoBitwiseRule = rule.Rule{
	Name: "no-bitwise",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// report emits the diagnostic after shared allow / int32Hint filtering.
		report := func(node *ast.Node, opKind ast.Kind, rightOperand *ast.Node) {
			if !bitwiseOperatorKinds[opKind] {
				return
			}
			op := scanner.TokenToString(opKind)
			if opts.allow[op] {
				return
			}
			// int32Hint: allow the `x | 0` integer-cast idiom when enabled.
			if opts.int32Hint && opKind == ast.KindBarToken && isZeroNumericLiteral(rightOperand) {
				return
			}
			ctx.ReportNode(node, buildUnexpectedMessage(op))
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil {
					return
				}
				report(node, binary.OperatorToken.Kind, binary.Right)
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				prefix := node.AsPrefixUnaryExpression()
				if prefix == nil {
					return
				}
				// Unary bitwise (`~`) has no right operand for the int32Hint check.
				report(node, prefix.Operator, nil)
			},
		}
	},
}
