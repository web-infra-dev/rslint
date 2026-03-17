package eqeqeq

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// operatorText converts an operator Kind to its string representation.
func operatorText(kind ast.Kind) string {
	switch kind {
	case ast.KindEqualsEqualsToken:
		return "=="
	case ast.KindExclamationEqualsToken:
		return "!="
	case ast.KindEqualsEqualsEqualsToken:
		return "==="
	case ast.KindExclamationEqualsEqualsToken:
		return "!=="
	default:
		return ""
	}
}

// strictEquivalent maps a loose operator to its strict counterpart.
func strictEquivalent(kind ast.Kind) string {
	switch kind {
	case ast.KindEqualsEqualsToken:
		return "==="
	case ast.KindExclamationEqualsToken:
		return "!=="
	default:
		return ""
	}
}

// looseEquivalent maps a strict operator to its loose counterpart.
func looseEquivalent(kind ast.Kind) string {
	switch kind {
	case ast.KindEqualsEqualsEqualsToken:
		return "=="
	case ast.KindExclamationEqualsEqualsToken:
		return "!="
	default:
		return ""
	}
}

// isLooseEquality checks if the operator is == or !=.
func isLooseEquality(kind ast.Kind) bool {
	return kind == ast.KindEqualsEqualsToken || kind == ast.KindExclamationEqualsToken
}

// isStrictEquality checks if the operator is === or !==.
func isStrictEquality(kind ast.Kind) bool {
	return kind == ast.KindEqualsEqualsEqualsToken || kind == ast.KindExclamationEqualsEqualsToken
}

// isNullLiteral checks if a node is the null keyword.
func isNullLiteral(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindNullKeyword
}

// isTypeOf checks if a node is a typeof expression.
func isTypeOf(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindTypeOfExpression
}

// areLiteralsAndSameType checks if both operands are literals of the same JS type.
func areLiteralsAndSameType(left, right *ast.Node) bool {
	if left == nil || right == nil {
		return false
	}

	// Both string literals
	if left.Kind == ast.KindStringLiteral && right.Kind == ast.KindStringLiteral {
		return true
	}

	// Both numeric literals
	if left.Kind == ast.KindNumericLiteral && right.Kind == ast.KindNumericLiteral {
		return true
	}

	// Both boolean literals (true/false)
	isBoolLeft := left.Kind == ast.KindTrueKeyword || left.Kind == ast.KindFalseKeyword
	isBoolRight := right.Kind == ast.KindTrueKeyword || right.Kind == ast.KindFalseKeyword
	if isBoolLeft && isBoolRight {
		return true
	}

	// Both null literals
	if left.Kind == ast.KindNullKeyword && right.Kind == ast.KindNullKeyword {
		return true
	}

	// Both regex literals
	if left.Kind == ast.KindRegularExpressionLiteral && right.Kind == ast.KindRegularExpressionLiteral {
		return true
	}

	return false
}

// buildUnexpectedMessage creates the error message for this rule.
func buildUnexpectedMessage(expectedOperator, actualOperator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Expected '" + expectedOperator + "' and instead saw '" + actualOperator + "'.",
	}
}

// eqeqeqOptions holds parsed options for the rule.
type eqeqeqOptions struct {
	mode       string // "always", "smart", "allow-null"
	nullOption string // "always", "never", "ignore" (only used in "always" mode)
}

// parseOptions extracts the mode and null sub-option from the rule options.
func parseOptions(opts any) eqeqeqOptions {
	result := eqeqeqOptions{
		mode:       "always",
		nullOption: "always",
	}

	if opts == nil {
		return result
	}

	// Options can be:
	// 1. A string: "always", "smart", "allow-null"
	// 2. An array: ["always", {"null": "ignore"}]
	switch v := opts.(type) {
	case string:
		result.mode = v
	case []interface{}:
		if len(v) > 0 {
			if modeStr, ok := v[0].(string); ok {
				result.mode = modeStr
			}
		}
		if len(v) > 1 {
			if optsMap, ok := v[1].(map[string]interface{}); ok {
				if nullVal, ok := optsMap["null"].(string); ok {
					result.nullOption = nullVal
				}
			}
		}
	}

	// "allow-null" is shorthand for ["always", {"null": "ignore"}]
	if result.mode == "allow-null" {
		result.mode = "always"
		result.nullOption = "ignore"
	}

	return result
}

// EqeqeqRule requires use of === and !==.
// https://eslint.org/docs/latest/rules/eqeqeq
var EqeqeqRule = rule.Rule{
	Name: "eqeqeq",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil {
					return
				}

				opKind := binary.OperatorToken.Kind
				left := binary.Left
				right := binary.Right

				if left == nil || right == nil {
					return
				}

				switch opts.mode {
				case "always":
					handleAlwaysMode(ctx, binary, opKind, left, right, opts.nullOption)
				case "smart":
					handleSmartMode(ctx, binary, opKind, left, right)
				}
			},
		}
	},
}

// handleAlwaysMode implements the "always" mode logic.
func handleAlwaysMode(ctx rule.RuleContext, binary *ast.BinaryExpression, opKind ast.Kind, left, right *ast.Node, nullOption string) {
	isNullCheck := isNullLiteral(left) || isNullLiteral(right)

	if isLooseEquality(opKind) {
		// Loose equality: == or !=
		switch nullOption {
		case "always":
			// Always require strict equality
			expected := strictEquivalent(opKind)
			actual := operatorText(opKind)
			ctx.ReportNode(binary.OperatorToken, buildUnexpectedMessage(expected, actual))
		case "never":
			// If it's a null comparison with loose equality, that's allowed
			// because we want == null for null checks, not === null
			if !isNullCheck {
				expected := strictEquivalent(opKind)
				actual := operatorText(opKind)
				ctx.ReportNode(binary.OperatorToken, buildUnexpectedMessage(expected, actual))
			}
		case "ignore":
			// If it's a null comparison, ignore it; otherwise flag it
			if !isNullCheck {
				expected := strictEquivalent(opKind)
				actual := operatorText(opKind)
				ctx.ReportNode(binary.OperatorToken, buildUnexpectedMessage(expected, actual))
			}
		}
	} else if isStrictEquality(opKind) && nullOption == "never" && isNullCheck {
		// In "never" mode, also flag === null and !== null
		// because we want == null instead
		expected := looseEquivalent(opKind)
		actual := operatorText(opKind)
		ctx.ReportNode(binary.OperatorToken, buildUnexpectedMessage(expected, actual))
	}
}

// handleSmartMode implements the "smart" mode logic.
// In smart mode, == is allowed for:
// 1. Comparing typeof expressions
// 2. Comparing two literals of the same type
// 3. Comparing against null
func handleSmartMode(ctx rule.RuleContext, binary *ast.BinaryExpression, opKind ast.Kind, left, right *ast.Node) {
	if !isLooseEquality(opKind) {
		return
	}

	// Allow typeof comparisons
	if isTypeOf(left) || isTypeOf(right) {
		return
	}

	// Allow same-type literal comparisons
	if areLiteralsAndSameType(left, right) {
		return
	}

	// Allow null comparisons
	if isNullLiteral(left) || isNullLiteral(right) {
		return
	}

	// None of the smart exceptions apply; flag it
	expected := strictEquivalent(opKind)
	actual := operatorText(opKind)
	ctx.ReportNode(binary.OperatorToken, buildUnexpectedMessage(expected, actual))
}
