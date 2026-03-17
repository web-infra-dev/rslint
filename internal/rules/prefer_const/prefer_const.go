package prefer_const

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/prefer-const
var PreferConstRule = rule.Rule{
	Name: "prefer-const",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				declList := node.AsVariableDeclarationList()
				if declList == nil {
					return
				}

				// Only check `let` declarations
				if node.Flags&ast.NodeFlagsLet == 0 {
					return
				}

				if declList.Declarations == nil {
					return
				}

				// Check if this declaration list is the initializer of a for-in or for-of statement
				isForInOrOf := isInForInOrOf(node)

				for _, decl := range declList.Declarations.Nodes {
					varDecl := decl.AsVariableDeclaration()
					if varDecl == nil || varDecl.Name() == nil {
						continue
					}

					// Must have an initializer, OR be in a for-in/for-of loop
					if varDecl.Initializer == nil && !isForInOrOf {
						continue
					}

					// For simple identifier declarations
					if varDecl.Name().Kind == ast.KindIdentifier {
						checkIdentifier(varDecl.Name(), decl, &ctx)
					}
				}
			},
		}
	},
}

// isInForInOrOf checks if a VariableDeclarationList is the initializer of a for-in or for-of statement
func isInForInOrOf(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	return node.Parent.Kind == ast.KindForInStatement || node.Parent.Kind == ast.KindForOfStatement
}

// checkIdentifier checks a single identifier to see if it should be const
func checkIdentifier(nameNode *ast.Node, declNode *ast.Node, ctx *rule.RuleContext) {
	if ctx.TypeChecker == nil {
		return
	}

	sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	if sym == nil {
		return
	}

	if !isReassigned(sym, declNode, ctx) {
		name := nameNode.Text()
		ctx.ReportNode(nameNode, rule.RuleMessage{
			Id:          "useConst",
			Description: "'" + name + "' is never reassigned. Use 'const' instead.",
		})
	}
}

// isReassigned checks if a symbol is ever assigned to after its declaration
func isReassigned(sym *ast.Symbol, declNode *ast.Node, ctx *rule.RuleContext) bool {
	found := false
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}

		if n.Kind == ast.KindIdentifier && !isPartOfDeclaration(n, declNode) {
			refSym := ctx.TypeChecker.GetSymbolAtLocation(n)
			if refSym == sym && isWriteReference(n) {
				found = true
				return
			}
		}

		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(ctx.SourceFile.AsNode())
	return found
}

// isPartOfDeclaration checks if an identifier node is part of the variable declaration itself
// (i.e., it's the declaration name, not a reference usage)
func isPartOfDeclaration(identNode *ast.Node, declNode *ast.Node) bool {
	current := identNode
	for current != nil {
		if current == declNode {
			return true
		}
		current = current.Parent
	}
	return false
}

// isWriteReference checks if a reference is a write operation (assignment, increment, decrement, etc.)
func isWriteReference(node *ast.Node) bool {
	if node == nil {
		return false
	}

	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary == nil {
			return false
		}

		// Check if the node is on the left side of an assignment
		if binary.Left != node {
			return false
		}

		// Check for all assignment operators
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken,
			ast.KindPlusEqualsToken,
			ast.KindMinusEqualsToken,
			ast.KindAsteriskEqualsToken,
			ast.KindSlashEqualsToken,
			ast.KindPercentEqualsToken,
			ast.KindAsteriskAsteriskEqualsToken,
			ast.KindLessThanLessThanEqualsToken,
			ast.KindGreaterThanGreaterThanEqualsToken,
			ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
			ast.KindAmpersandEqualsToken,
			ast.KindBarEqualsToken,
			ast.KindCaretEqualsToken,
			ast.KindQuestionQuestionEqualsToken,
			ast.KindAmpersandAmpersandEqualsToken,
			ast.KindBarBarEqualsToken:
			return true
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

	case ast.KindPostfixUnaryExpression:
		postfix := parent.AsPostfixUnaryExpression()
		if postfix == nil {
			return false
		}
		switch postfix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return postfix.Operand == node
		}

	case ast.KindShorthandPropertyAssignment:
		return isInDestructuringAssignment(parent)

	case ast.KindPropertyAssignment:
		propAssignment := parent.AsPropertyAssignment()
		if propAssignment != nil && propAssignment.Initializer == node {
			return isInDestructuringAssignment(parent)
		}

	case ast.KindArrayLiteralExpression:
		return isInDestructuringAssignment(parent)

	case ast.KindObjectLiteralExpression:
		return isInDestructuringAssignment(parent)

	case ast.KindParenthesizedExpression:
		return isWriteReference(parent)

	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		return isWriteReference(parent)
	}

	return false
}

// isInDestructuringAssignment checks if a node is part of a destructuring assignment pattern
func isInDestructuringAssignment(node *ast.Node) bool {
	current := node
	for current != nil {
		if current.Kind == ast.KindObjectLiteralExpression || current.Kind == ast.KindArrayLiteralExpression {
			parent := current.Parent

			// Unwrap parentheses
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
			return false
		}
		current = current.Parent
	}
	return false
}
