package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// IsWriteReference checks if a node is a write reference (assignment target).
// This covers direct assignments, compound assignments, increment/decrement,
// destructuring patterns, and type assertion wrappers.
func IsWriteReference(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}

		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken,
			ast.KindPlusEqualsToken,
			ast.KindMinusEqualsToken,
			ast.KindAsteriskAsteriskEqualsToken,
			ast.KindAsteriskEqualsToken,
			ast.KindSlashEqualsToken,
			ast.KindPercentEqualsToken,
			ast.KindLessThanLessThanEqualsToken,
			ast.KindGreaterThanGreaterThanEqualsToken,
			ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
			ast.KindAmpersandEqualsToken,
			ast.KindBarEqualsToken,
			ast.KindCaretEqualsToken,
			ast.KindBarBarEqualsToken,
			ast.KindAmpersandAmpersandEqualsToken,
			ast.KindQuestionQuestionEqualsToken:
			return binary.Left == node
		}

	case ast.KindPostfixUnaryExpression:
		postfix := parent.AsPostfixUnaryExpression()
		if postfix == nil {
			return false
		}
		switch postfix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return postfix.Operand == node
		}

	case ast.KindPrefixUnaryExpression:
		prefix := parent.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}
		switch prefix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return prefix.Operand == node
		}

	case ast.KindObjectBindingPattern:
		return IsBindingPatternInAssignment(parent)

	case ast.KindArrayBindingPattern:
		return IsBindingPatternInAssignment(parent)

	case ast.KindBindingElement:
		return IsWriteReference(parent)

	case ast.KindShorthandPropertyAssignment:
		shorthand := parent.AsShorthandPropertyAssignment()
		if shorthand != nil && shorthand.Name() == node {
			return IsInDestructuringAssignment(parent)
		}

	case ast.KindPropertyAssignment:
		propAssignment := parent.AsPropertyAssignment()
		if propAssignment != nil && propAssignment.Initializer == node {
			return IsInDestructuringAssignment(parent)
		}

	case ast.KindArrayLiteralExpression:
		return IsInDestructuringAssignment(parent)

	case ast.KindObjectLiteralExpression:
		return IsInDestructuringAssignment(parent)

	case ast.KindSpreadElement:
		// ...x in array destructuring assignment context
		return IsWriteReference(parent)

	case ast.KindSpreadAssignment:
		// ...x in object destructuring assignment context
		return IsInDestructuringAssignment(parent)

	case ast.KindForInStatement, ast.KindForOfStatement:
		// for (x in obj) / for (x of arr) — x is a write target
		stmt := parent.AsForInOrOfStatement()
		if stmt != nil {
			return stmt.Initializer == node
		}

	case ast.KindParenthesizedExpression:
		return IsWriteReference(parent)

	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		return IsWriteReference(parent)

	case ast.KindNonNullExpression:
		return IsWriteReference(parent)
	}

	return false
}

// IsBindingPatternInAssignment checks if a binding pattern is the left side of an assignment.
func IsBindingPatternInAssignment(node *ast.Node) bool {
	if node == nil {
		return false
	}

	parent := node.Parent

	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}

	if parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			leftNode := binary.Left
			for leftNode != nil && leftNode.Kind == ast.KindParenthesizedExpression {
				parenExpr := leftNode.AsParenthesizedExpression()
				if parenExpr != nil {
					leftNode = parenExpr.Expression
				} else {
					break
				}
			}
			return leftNode == node
		}
	}

	return false
}

// IsInDestructuringAssignment checks if a node is part of a destructuring assignment pattern.
// Walks up from the node through nested array/object literals to find a top-level
// destructuring assignment (e.g. [{a}] = [...] or {x: [a]} = {...}).
func IsInDestructuringAssignment(node *ast.Node) bool {
	current := node
	for current != nil {
		if current.Kind == ast.KindObjectLiteralExpression || current.Kind == ast.KindArrayLiteralExpression {
			parent := current.Parent

			for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
				parent = parent.Parent
			}

			if parent != nil && parent.Kind == ast.KindBinaryExpression {
				binary := parent.AsBinaryExpression()
				if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
					leftNode := binary.Left
					for leftNode != nil && leftNode.Kind == ast.KindParenthesizedExpression {
						parenExpr := leftNode.AsParenthesizedExpression()
						if parenExpr != nil {
							leftNode = parenExpr.Expression
						} else {
							break
						}
					}
					if leftNode == current {
						return true
					}
				}
			}
			// Check if this is a destructuring target in for-in/for-of
			if parent != nil && (parent.Kind == ast.KindForInStatement || parent.Kind == ast.KindForOfStatement) {
				stmt := parent.AsForInOrOfStatement()
				if stmt != nil && stmt.Initializer == current {
					return true
				}
			}

			// Continue walking up — this array/object might be nested inside
			// another destructuring pattern (e.g. [{a}] = [...]).
		}
		current = current.Parent
	}
	return false
}
