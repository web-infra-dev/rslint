package no_const_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder
func buildConstMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "const",
		Description: "'" + name + "' is constant.",
	}
}

// isConstBinding checks if a variable declaration is a const binding
func isConstBinding(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindVariableDeclarationList {
		return false
	}

	varDeclList := node.AsVariableDeclarationList()
	if varDeclList == nil {
		return false
	}

	// Check if the declaration is const (or using/await using in the future)
	// In TypeScript AST, const declarations have flags
	return (varDeclList.Flags & ast.NodeFlagsConst) != 0
}

// getIdentifierName gets the name of an identifier node
func getIdentifierName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}

	return node.Text()
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
		// Check if this is an assignment operation
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
		case ast.KindEqualsToken, // =
			ast.KindPlusEqualsToken,                              // +=
			ast.KindMinusEqualsToken,                             // -=
			ast.KindAsteriskEqualsToken,                          // *=
			ast.KindSlashEqualsToken,                             // /=
			ast.KindPercentEqualsToken,                           // %=
			ast.KindAsteriskAsteriskEqualsToken,                  // **=
			ast.KindLessThanLessThanEqualsToken,                  // <<=
			ast.KindGreaterThanGreaterThanEqualsToken,            // >>=
			ast.KindGreaterThanGreaterThanGreaterThanEqualsToken, // >>>=
			ast.KindAmpersandEqualsToken,                         // &=
			ast.KindBarEqualsToken,                               // |=
			ast.KindCaretEqualsToken,                             // ^=
			ast.KindQuestionQuestionEqualsToken,                  // ??=
			ast.KindAmpersandAmpersandEqualsToken,                // &&=
			ast.KindBarBarEqualsToken:                            // ||=
			return true
		}

	case ast.KindPrefixUnaryExpression:
		// Check for ++ and -- prefix operators
		prefix := parent.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}

		switch prefix.Operator {
		case ast.KindPlusPlusToken, // ++
			ast.KindMinusMinusToken: // --
			return prefix.Operand == node
		}

	case ast.KindPostfixUnaryExpression:
		// Check for ++ and -- postfix operators
		postfix := parent.AsPostfixUnaryExpression()
		if postfix == nil {
			return false
		}

		switch postfix.Operator {
		case ast.KindPlusPlusToken, // ++
			ast.KindMinusMinusToken: // --
			return postfix.Operand == node
		}

	case ast.KindObjectBindingPattern:
		// In destructuring like {x} = obj, x is a write reference
		return isBindingPatternInAssignment(parent)

	case ast.KindArrayBindingPattern:
		// In array destructuring like [x] = arr, x is a write reference
		return isBindingPatternInAssignment(parent)

	case ast.KindBindingElement:
		// Check if the binding element is part of a write context
		return isWriteReference(parent)

	case ast.KindShorthandPropertyAssignment:
		// In destructuring like {x} = obj or ({x} = obj), x is a write reference
		// Check if the parent shorthand property is in a destructuring assignment
		return isInDestructuringAssignment(parent)

	case ast.KindPropertyAssignment:
		// In destructuring like {b: x} = obj, x is a write reference
		propAssignment := parent.AsPropertyAssignment()
		if propAssignment != nil && propAssignment.Initializer == node {
			return isInDestructuringAssignment(parent)
		}

	case ast.KindObjectLiteralExpression:
		// In object destructuring like {x} = obj, x is a write reference
		return isInDestructuringAssignment(parent)

	case ast.KindArrayLiteralExpression:
		// In array destructuring like [x] = arr, x is a write reference
		return isInDestructuringAssignment(parent)

	case ast.KindParenthesizedExpression:
		// Unwrap parentheses and check the parent context
		return isWriteReference(parent)

	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		// Type assertions like (x as any) = 0
		// The type assertion wraps the identifier, check if the assertion is a write target
		return isWriteReference(parent)
	}

	return false
}

// isBindingPatternInAssignment checks if a binding pattern is the left side of an assignment
func isBindingPatternInAssignment(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// The binding pattern's parent might be wrapped in parentheses
	parent := node.Parent

	// Unwrap parentheses
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}

	// Check if the parent is a binary expression with = operator
	if parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			// Check if the binding pattern is on the left side
			leftNode := binary.Left
			// Unwrap parentheses on the left side
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

// isInDestructuringAssignment checks if a node is part of a destructuring assignment pattern
func isInDestructuringAssignment(node *ast.Node) bool {
	current := node
	for current != nil {
		if current.Kind == ast.KindObjectLiteralExpression || current.Kind == ast.KindArrayLiteralExpression {
			// Check if this literal is the left side of an assignment
			// May be wrapped in parentheses
			parent := current.Parent

			// Unwrap parentheses
			for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
				parent = parent.Parent
			}

			if parent != nil && parent.Kind == ast.KindBinaryExpression {
				binary := parent.AsBinaryExpression()
				if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
					// Check if the literal (or its parent parenthesized expression) is on the left
					leftNode := binary.Left
					// Unwrap parentheses on left side
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

// checkIdentifierWrite checks if an identifier is a write reference to a const variable
func checkIdentifierWrite(node *ast.Node, ctx *rule.RuleContext, constSymbols map[*ast.Symbol]bool) {
	// Check if this is a write reference (assignment, increment, etc.)
	if !isWriteReference(node) {
		return
	}

	// Get the symbol for this identifier
	if ctx.TypeChecker == nil {
		return
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return
	}

	// Check if this symbol refers to a const variable
	if !constSymbols[symbol] {
		return
	}

	// Report the violation
	identName := getIdentifierName(node)
	ctx.ReportNode(node, buildConstMessage(identName))
}

// NoConstAssignRule disallows reassigning const variables
var NoConstAssignRule = rule.CreateRule(rule.Rule{
	Name: "no-const-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Track const declarations by their symbol
		constSymbols := make(map[*ast.Symbol]bool)

		return rule.RuleListeners{
			// Track const variable declarations
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if !isConstBinding(node) {
					return
				}

				varDeclList := node.AsVariableDeclarationList()
				if varDeclList == nil || varDeclList.Declarations == nil {
					return
				}

				// Track all identifiers declared as const using their symbols
				for _, decl := range varDeclList.Declarations.Nodes {
					if decl.Kind != ast.KindVariableDeclaration {
						continue
					}

					varDecl := decl.AsVariableDeclaration()
					if varDecl == nil || varDecl.Name() == nil {
						continue
					}

					// Collect symbols for all identifiers in the binding name
					collectSymbols(varDecl.Name(), &ctx, constSymbols)
				}
			},

			// Check for reassignments to const variables
			ast.KindIdentifier: func(node *ast.Node) {
				checkIdentifierWrite(node, &ctx, constSymbols)
			},

			// Check shorthand property assignments in destructuring (e.g., {x} = obj)
			ast.KindShorthandPropertyAssignment: func(node *ast.Node) {
				shorthand := node.AsShorthandPropertyAssignment()
				if shorthand == nil || shorthand.Name() == nil {
					return
				}

				// Check if this shorthand is in a destructuring assignment
				if !isInDestructuringAssignment(node) {
					return
				}

				// This is a write reference, check if it refers to a const variable
				if ctx.TypeChecker == nil {
					return
				}

				symbol := ctx.TypeChecker.GetSymbolAtLocation(shorthand.Name())
				if symbol == nil {
					return
				}

				// Check if this symbol refers to a const variable
				if !constSymbols[symbol] {
					return
				}

				// Report the violation
				identName := getIdentifierName(shorthand.Name())
				ctx.ReportNode(shorthand.Name(), buildConstMessage(identName))
			},
		}
	},
})

// collectSymbols recursively collects symbols for all identifiers from a binding pattern
func collectSymbols(bindingName *ast.Node, ctx *rule.RuleContext, constSymbols map[*ast.Symbol]bool) {
	if bindingName == nil || ctx.TypeChecker == nil {
		return
	}

	switch bindingName.Kind {
	case ast.KindIdentifier:
		symbol := ctx.TypeChecker.GetSymbolAtLocation(bindingName)
		if symbol != nil {
			constSymbols[symbol] = true
		}

	case ast.KindObjectBindingPattern:
		// Walk through child nodes to find binding elements
		bindingName.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				bindingElem := child.AsBindingElement()
				if bindingElem != nil && bindingElem.Name() != nil {
					collectSymbols(bindingElem.Name(), ctx, constSymbols)
				}
			}
			return false
		})

	case ast.KindArrayBindingPattern:
		// Walk through child nodes to find binding elements
		bindingName.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				bindingElem := child.AsBindingElement()
				if bindingElem != nil && bindingElem.Name() != nil {
					collectSymbols(bindingElem.Name(), ctx, constSymbols)
				}
			}
			return false
		})
	}
}
