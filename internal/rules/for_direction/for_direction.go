package for_direction

import (
	"math/big"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder
func buildIncorrectDirection() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "incorrectDirection",
		Description: "The update clause in this loop moves the variable in the wrong direction.",
	}
}

// getUpdateDirection determines the direction of a unary update expression (++ or --)
// Returns: 1 for increment, -1 for decrement, 0 for unknown
func getUpdateDirection(updateExpr *ast.Node) int {
	if updateExpr == nil {
		return 0
	}

	switch updateExpr.Kind {
	case ast.KindPostfixUnaryExpression:
		postfix := updateExpr.AsPostfixUnaryExpression()
		if postfix != nil {
			switch postfix.Operator {
			case ast.KindPlusPlusToken:
				return 1
			case ast.KindMinusMinusToken:
				return -1
			}
		}
	case ast.KindPrefixUnaryExpression:
		prefix := updateExpr.AsPrefixUnaryExpression()
		if prefix != nil {
			switch prefix.Operator {
			case ast.KindPlusPlusToken:
				return 1
			case ast.KindMinusMinusToken:
				return -1
			}
		}
	}

	return 0
}

// getStaticNumberValue extracts static numeric value from an expression
// Returns the numeric value and true if it's a static number, 0 and false otherwise
func getStaticNumberValue(node *ast.Node, ctx *rule.RuleContext) (*big.Float, bool) {
	if node == nil {
		return nil, false
	}

	switch node.Kind {
	case ast.KindNumericLiteral:
		// Parse numeric literal
		text := node.Text()
		if text != "" {
			val := new(big.Float)
			if _, ok := val.SetString(text); ok {
				return val, true
			}
		}
		return nil, false

	case ast.KindPrefixUnaryExpression:
		// Handle unary minus/plus
		prefix := node.AsPrefixUnaryExpression()
		if prefix != nil && prefix.Operand != nil {
			if val, ok := getStaticNumberValue(prefix.Operand, ctx); ok {
				switch prefix.Operator {
				case ast.KindMinusToken:
					negVal := new(big.Float).Neg(val)
					return negVal, true
				case ast.KindPlusToken:
					return val, true
				}
			}
		}
		return nil, false

	case ast.KindParenthesizedExpression:
		// Unwrap parentheses
		expr := node.Expression()
		return getStaticNumberValue(expr, ctx)

	case ast.KindIdentifier:
		// Try to resolve constant variable values
		if ctx != nil && ctx.TypeChecker != nil {
			constValue := ctx.TypeChecker.GetConstantValue(node)
			if constValue != nil {
				// Try to convert the constant value to a float
				switch v := constValue.(type) {
				case float64:
					return big.NewFloat(v), true
				case int:
					return big.NewFloat(float64(v)), true
				case int64:
					return big.NewFloat(float64(v)), true
				}
			}
		}
		return nil, false
	}

	return nil, false
}

// getAssignmentDirection determines the direction of a compound assignment (+=, -=)
// Returns: 1 for positive direction, -1 for negative direction, 0 for unknown
func getAssignmentDirection(binaryExpr *ast.Node, ctx *rule.RuleContext) int {
	if binaryExpr == nil || binaryExpr.Kind != ast.KindBinaryExpression {
		return 0
	}

	binary := binaryExpr.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return 0
	}

	operator := binary.OperatorToken.Kind

	switch operator {
	case ast.KindPlusEqualsToken:
		// += direction depends on the right operand
		if val, ok := getStaticNumberValue(binary.Right, ctx); ok {
			if val.Sign() > 0 {
				return 1
			} else if val.Sign() < 0 {
				return -1
			}
		}
		return 0

	case ast.KindMinusEqualsToken:
		// -= direction is opposite of the right operand
		if val, ok := getStaticNumberValue(binary.Right, ctx); ok {
			if val.Sign() > 0 {
				return -1
			} else if val.Sign() < 0 {
				return 1
			}
		}
		return 0
	}

	return 0
}

// getVariableName extracts the variable name from an expression
func getVariableName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.Text()
	case ast.KindPostfixUnaryExpression:
		postfix := node.AsPostfixUnaryExpression()
		if postfix != nil {
			return getVariableName(postfix.Operand)
		}
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix != nil {
			return getVariableName(prefix.Operand)
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil {
			return getVariableName(binary.Left)
		}
	}

	return ""
}

// getTestVariableName extracts the variable name from a test condition
func getTestVariableName(testExpr *ast.Node, onLeft bool) string {
	if testExpr == nil || testExpr.Kind != ast.KindBinaryExpression {
		return ""
	}

	binary := testExpr.AsBinaryExpression()
	if binary == nil {
		return ""
	}

	if onLeft {
		return getVariableName(binary.Left)
	}
	return getVariableName(binary.Right)
}

// getExpectedDirection determines the expected direction based on the comparison operator
// Returns: 1 for increasing, -1 for decreasing, 0 for unknown
func getExpectedDirection(testExpr *ast.Node, counterOnLeft bool) int {
	if testExpr == nil || testExpr.Kind != ast.KindBinaryExpression {
		return 0
	}

	binary := testExpr.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return 0
	}

	operator := binary.OperatorToken.Kind

	// Determine expected direction based on operator and counter position
	if counterOnLeft {
		// Counter is on the left side (e.g., i < 10)
		switch operator {
		case ast.KindLessThanToken, ast.KindLessThanEqualsToken:
			// i < 10 or i <= 10 -> should increase
			return 1
		case ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
			// i > 10 or i >= 10 -> should decrease
			return -1
		}
	} else {
		// Counter is on the right side (e.g., 10 > i)
		// The logic is reversed: 10 > i means i < 10
		switch operator {
		case ast.KindLessThanToken, ast.KindLessThanEqualsToken:
			// 10 < i or 10 <= i is equivalent to i > 10 -> should decrease
			return -1
		case ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
			// 10 > i or 10 >= i is equivalent to i < 10 -> should increase
			return 1
		}
	}

	return 0
}

// ForDirectionRule enforces that for loop update clauses move the counter in the right direction
var ForDirectionRule = rule.CreateRule(rule.Rule{
	Name: "for-direction",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindForStatement: func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt == nil {
					return
				}

				// We need all three parts: initializer, condition, and incrementor
				if forStmt.Condition == nil || forStmt.Incrementor == nil {
					return
				}

				condition := forStmt.Condition
				incrementor := forStmt.Incrementor

				// Condition must be a binary expression with comparison operator
				if condition.Kind != ast.KindBinaryExpression {
					return
				}

				binary := condition.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil {
					return
				}

				operator := binary.OperatorToken.Kind

				// Only check comparison operators
				switch operator {
				case ast.KindLessThanToken, ast.KindLessThanEqualsToken,
					ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
					// These are the operators we care about
				default:
					// Not a comparison operator we handle
					return
				}

				// Try to determine which side has the counter variable
				// First, try assuming counter is on the left
				counterOnLeft := true
				counterName := getVariableName(binary.Left)

				// If left side doesn't give us a variable name, try right side
				if counterName == "" {
					counterOnLeft = false
					counterName = getVariableName(binary.Right)
				}

				// If we still don't have a counter name, we can't proceed
				if counterName == "" {
					return
				}

				// Get the name of the variable being updated
				updateVarName := getVariableName(incrementor)

				// If the update variable doesn't match the test variable, we can't determine direction
				if updateVarName != counterName {
					return
				}

				// Get expected direction from the test condition
				expectedDirection := getExpectedDirection(condition, counterOnLeft)
				if expectedDirection == 0 {
					return
				}

				// Get actual direction from the update expression
				actualDirection := getUpdateDirection(incrementor)

				// If we couldn't determine direction from unary operators, try compound assignment
				if actualDirection == 0 {
					actualDirection = getAssignmentDirection(incrementor, &ctx)
				}

				// If we still can't determine direction, it's ambiguous (e.g., i += unknown)
				if actualDirection == 0 {
					return
				}

				// Check if directions match
				if expectedDirection != actualDirection {
					// Report error on the entire for statement (from 'for' to opening brace)
					ctx.ReportNode(node, buildIncorrectDirection())
				}
			},
		}
	},
})
