package no_constant_binary_expression

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builders
func buildConstantBinaryOperandMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "constantBinaryOperand",
		Description: "Unexpected constant binary expression. Comparisons will always evaluate the same.",
	}
}

func buildConstantShortCircuitMessage(property string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "constantShortCircuit",
		Description: "Unexpected constant " + property + " on the left-hand side of a `" + property + "` expression.",
	}
}

func buildAlwaysNewMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysNew",
		Description: "Unexpected comparison to newly constructed object. These two values can never be equal.",
	}
}

func buildBothAlwaysNewMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bothAlwaysNew",
		Description: "Unexpected comparison of two newly constructed objects. These two values can never be equal.",
	}
}

// isNullOrUndefined checks if a node represents null or undefined
func isNullOrUndefined(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		// Check for 'undefined' identifier
		return node.Text() == "undefined"
	case ast.KindVoidExpression:
		// void operator always produces undefined
		return true
	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && isNullOrUndefined(paren.Expression)
	}
	return false
}

// isStaticBoolean checks if a node is a static boolean value
func isStaticBoolean(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && isStaticBoolean(paren.Expression)
	case ast.KindPrefixUnaryExpression:
		// !value is always a boolean
		prefix := node.AsPrefixUnaryExpression()
		if prefix != nil && prefix.Operator == ast.KindExclamationToken {
			// Any value negated with ! produces a boolean
			return true
		}
	case ast.KindCallExpression:
		// Boolean() with constant argument is a static boolean
		call := node.AsCallExpression()
		if call != nil && call.Expression != nil && call.Expression.Kind == ast.KindIdentifier {
			if call.Expression.Text() == "Boolean" {
				if call.Arguments != nil && len(call.Arguments.Nodes) == 1 {
					return isConstant(call.Arguments.Nodes[0])
				}
			}
		}
	}
	return false
}

// isAlwaysNew checks if an expression creates a new object/array/function/regex
func isAlwaysNew(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindRegularExpressionLiteral,
		ast.KindNewExpression:
		return true
	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && isAlwaysNew(paren.Expression)
	}
	return false
}

// isConstant checks if a node is a constant value
func isConstant(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindNumericLiteral,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral,
		ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindNewExpression:
		return true
	case ast.KindIdentifier:
		// 'undefined' is a constant
		return node.Text() == "undefined"
	case ast.KindVoidExpression:
		// void operator
		return true
	case ast.KindParenthesizedExpression:
		// Parenthesized expressions - check the inner expression
		paren := node.AsParenthesizedExpression()
		return paren != nil && isConstant(paren.Expression)
	case ast.KindPrefixUnaryExpression:
		// Unary operators
		prefix := node.AsPrefixUnaryExpression()
		if prefix != nil {
			// ! (logical not) always produces a boolean, so it's constant behavior
			if prefix.Operator == ast.KindExclamationToken {
				return true
			}
			// Other unary operators (+, -, ~, etc.) are constant if operand is constant
			return isConstant(prefix.Operand)
		}
	case ast.KindCallExpression:
		// Built-in type conversions: Boolean(), String(), Number()
		// Only treat as constant if argument is constant (to avoid issues with shadowing)
		call := node.AsCallExpression()
		if call != nil && call.Expression != nil && call.Expression.Kind == ast.KindIdentifier {
			name := call.Expression.Text()
			if name == "Boolean" || name == "String" || name == "Number" {
				if call.Arguments != nil && len(call.Arguments.Nodes) == 1 {
					return isConstant(call.Arguments.Nodes[0])
				}
			}
		}
	}
	return false
}

// hasConstantNullishness checks if a node is always nullish or always non-nullish
func hasConstantNullishness(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Null or undefined is always nullish
	if isNullOrUndefined(node) {
		return true
	}

	// These are always non-nullish
	switch node.Kind {
	case ast.KindNumericLiteral,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindRegularExpressionLiteral,
		ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindNewExpression:
		return true
	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && hasConstantNullishness(paren.Expression)
	case ast.KindCallExpression:
		// Boolean(), String(), Number() with constant arguments are non-nullish
		call := node.AsCallExpression()
		if call != nil && call.Expression != nil && call.Expression.Kind == ast.KindIdentifier {
			name := call.Expression.Text()
			if name == "Boolean" || name == "String" || name == "Number" {
				if call.Arguments != nil && len(call.Arguments.Nodes) == 1 {
					return isConstant(call.Arguments.Nodes[0])
				}
			}
		}
	case ast.KindBinaryExpression:
		// Nullish coalescing operator
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil {
			if binary.OperatorToken.Kind == ast.KindQuestionQuestionToken {
				// left ?? right is constant if both sides have constant nullishness
				return hasConstantNullishness(binary.Left) && hasConstantNullishness(binary.Right)
			}
		}
	}
	return false
}

// hasConstantLooseBooleanComparison checks if a value has constant truthiness for loose equality
func hasConstantLooseBooleanComparison(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Static booleans have constant boolean comparison
	if isStaticBoolean(node) {
		return true
	}

	// These always coerce to specific boolean values in loose equality
	switch node.Kind {
	case ast.KindNumericLiteral:
		// Numbers always have constant boolean value
		return true
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		// Strings always have constant boolean value
		return true
	case ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		return node.Text() == "undefined"
	case ast.KindVoidExpression:
		return true
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		// Objects and arrays are always truthy
		return true
	case ast.KindRegularExpressionLiteral:
		return true
	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && hasConstantLooseBooleanComparison(paren.Expression)
	}
	return false
}

// hasConstantStrictBooleanComparison checks if a value has constant result for strict boolean equality
func hasConstantStrictBooleanComparison(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// For strict equality, only booleans can === booleans
	// Non-booleans always return false when compared with === to boolean
	if isStaticBoolean(node) {
		return true
	}

	// Non-boolean constants always return false with ===
	switch node.Kind {
	case ast.KindNumericLiteral,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNullKeyword,
		ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindRegularExpressionLiteral:
		return true
	case ast.KindIdentifier:
		return node.Text() == "undefined"
	case ast.KindVoidExpression:
		return true
	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && hasConstantStrictBooleanComparison(paren.Expression)
	}
	return false
}

// findBinaryExpressionConstantOperand finds which operand makes a comparison constant
func findBinaryExpressionConstantOperand(left, right *ast.Node, operator ast.Kind) *ast.Node {
	// For equality operators
	switch operator {
	case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
		// Loose equality - check for constant boolean comparison
		// Only report if comparing a boolean to another constant (not a variable)
		if isStaticBoolean(left) && hasConstantLooseBooleanComparison(right) {
			return left
		}
		if isStaticBoolean(right) && hasConstantLooseBooleanComparison(left) {
			return right
		}

	case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		// Strict equality - check for constant boolean comparison
		// Only report if comparing a boolean to another constant (not a variable)
		if isStaticBoolean(left) && hasConstantStrictBooleanComparison(right) {
			return left
		}
		if isStaticBoolean(right) && hasConstantStrictBooleanComparison(left) {
			return right
		}
	}

	// Check for nullish comparisons
	// Only report when BOTH sides have constant nullishness
	if operator == ast.KindEqualsEqualsToken || operator == ast.KindExclamationEqualsToken ||
		operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
		// Check if comparing null/undefined with another constant nullish value
		if isNullOrUndefined(left) && isNullOrUndefined(right) {
			return right
		}
		// Check if comparing constant non-nullish with null/undefined
		if isNullOrUndefined(left) && !isNullOrUndefined(right) && hasConstantNullishness(right) {
			return right
		}
		if isNullOrUndefined(right) && !isNullOrUndefined(left) && hasConstantNullishness(left) {
			return left
		}
	}

	return nil
}

// NoConstantBinaryExpressionRule detects constant binary expressions
var NoConstantBinaryExpressionRule = rule.CreateRule(rule.Rule{
	Name: "no-constant-binary-expression",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// Check logical expressions (&&, ||, ??)
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil {
					return
				}

				operator := binary.OperatorToken.Kind

				// Check for logical operators with constant short-circuit
				switch operator {
				case ast.KindAmpersandAmpersandToken: // &&
					if isConstant(binary.Left) {
						ctx.ReportNode(node, buildConstantShortCircuitMessage("&&"))
					}
				case ast.KindBarBarToken: // ||
					if isConstant(binary.Left) {
						ctx.ReportNode(node, buildConstantShortCircuitMessage("||"))
					}
				case ast.KindQuestionQuestionToken: // ??
					if hasConstantNullishness(binary.Left) {
						ctx.ReportNode(node, buildConstantShortCircuitMessage("??"))
					}
				}

				// Check for equality comparisons
				switch operator {
				case ast.KindEqualsEqualsToken,
					ast.KindExclamationEqualsToken,
					ast.KindEqualsEqualsEqualsToken,
					ast.KindExclamationEqualsEqualsToken:

					// Check if both operands are always new (can never be equal)
					if isAlwaysNew(binary.Left) && isAlwaysNew(binary.Right) {
						ctx.ReportNode(node, buildBothAlwaysNewMessage())
						return
					}

					// Check if one operand is always new and the other is a variable (not a constant)
					if isAlwaysNew(binary.Left) && !isConstant(binary.Right) {
						ctx.ReportNode(node, buildAlwaysNewMessage())
						return
					}
					if isAlwaysNew(binary.Right) && !isConstant(binary.Left) {
						ctx.ReportNode(node, buildAlwaysNewMessage())
						return
					}

					// Check for constant operands in comparisons
					constantOperand := findBinaryExpressionConstantOperand(binary.Left, binary.Right, operator)
					if constantOperand != nil {
						ctx.ReportNode(node, buildConstantBinaryOperandMessage())
					}
				}
			},
		}
	},
})
