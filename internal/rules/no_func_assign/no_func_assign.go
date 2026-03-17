package no_func_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder
func buildMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "isAFunction",
		Description: "'" + name + "' is a function.",
	}
}

// getIdentifierName extracts the name from an identifier node
func getIdentifierName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}
	return node.Text()
}

// isWriteReference checks if a node is a write reference (assignment target)
func isWriteReference(node *ast.Node) bool {
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

		// Check if it's an assignment operator and node is on the left side
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
		return isBindingPatternInAssignment(parent)

	case ast.KindArrayBindingPattern:
		return isBindingPatternInAssignment(parent)

	case ast.KindBindingElement:
		return isWriteReference(parent)

	case ast.KindShorthandPropertyAssignment:
		shorthand := parent.AsShorthandPropertyAssignment()
		if shorthand != nil && shorthand.Name() == node {
			return isInDestructuringAssignment(parent)
		}

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

// isBindingPatternInAssignment checks if a binding pattern is the left side of an assignment
func isBindingPatternInAssignment(node *ast.Node) bool {
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

// isInDestructuringAssignment checks if a node is part of a destructuring assignment pattern
func isInDestructuringAssignment(node *ast.Node) bool {
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
			return false
		}
		current = current.Parent
	}
	return false
}

// getFuncSymbol gets the symbol for a function node
func getFuncSymbol(funcNode *ast.Node, ctx *rule.RuleContext) *ast.Symbol {
	if ctx.TypeChecker == nil {
		return nil
	}

	switch funcNode.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := funcNode.AsFunctionDeclaration()
		if funcDecl != nil && funcDecl.Name() != nil {
			return ctx.TypeChecker.GetSymbolAtLocation(funcDecl.Name())
		}
	case ast.KindFunctionExpression:
		funcExpr := funcNode.AsFunctionExpression()
		if funcExpr != nil && funcExpr.Name() != nil {
			return ctx.TypeChecker.GetSymbolAtLocation(funcExpr.Name())
		}
	}

	return nil
}

// isNameShadowed checks if an identifier references a different variable due to shadowing
func isNameShadowed(node *ast.Node, funcName string, funcNode *ast.Node, ctx *rule.RuleContext) bool {
	if node == nil || ctx.TypeChecker == nil {
		return false
	}

	// Get the symbol at the identifier location
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}

	// Get the symbol of the function declaration
	funcSymbol := getFuncSymbol(funcNode, ctx)

	// If symbols are different, the name is shadowed
	if funcSymbol != nil {
		return symbol != funcSymbol
	}

	// Fallback: check if the identifier is within a scope that shadows the function name
	return isInShadowingScope(node, funcName, funcNode)
}

// isInShadowingScope checks if a node is within a scope that shadows the function name
func isInShadowingScope(node *ast.Node, funcName string, funcNode *ast.Node) bool {
	current := node.Parent
	for current != nil && current != funcNode {
		switch current.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindConstructor,
			ast.KindGetAccessor,
			ast.KindSetAccessor:
			if hasShadowingParameter(current, funcName) {
				return true
			}

		case ast.KindBlock:
			if hasShadowingVariable(current, funcName) {
				return true
			}

		case ast.KindCatchClause:
			catchClause := current.AsCatchClause()
			if catchClause != nil && catchClause.VariableDeclaration != nil {
				varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
				if varDecl != nil && varDecl.Name() != nil {
					if getIdentifierName(varDecl.Name()) == funcName {
						return true
					}
				}
			}
		}
		current = current.Parent
	}
	return false
}

// hasShadowingParameter checks if a function has a parameter with the given name
func hasShadowingParameter(node *ast.Node, name string) bool {
	var params []*ast.Node

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl != nil && funcDecl.Parameters != nil {
			params = funcDecl.Parameters.Nodes
		}
	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		if funcExpr != nil && funcExpr.Parameters != nil {
			params = funcExpr.Parameters.Nodes
		}
	case ast.KindArrowFunction:
		arrowFunc := node.AsArrowFunction()
		if arrowFunc != nil && arrowFunc.Parameters != nil {
			params = arrowFunc.Parameters.Nodes
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil && method.Parameters != nil {
			params = method.Parameters.Nodes
		}
	case ast.KindConstructor:
		constructor := node.AsConstructorDeclaration()
		if constructor != nil && constructor.Parameters != nil {
			params = constructor.Parameters.Nodes
		}
	case ast.KindGetAccessor:
		getter := node.AsGetAccessorDeclaration()
		if getter != nil && getter.Parameters != nil {
			params = getter.Parameters.Nodes
		}
	case ast.KindSetAccessor:
		setter := node.AsSetAccessorDeclaration()
		if setter != nil && setter.Parameters != nil {
			params = setter.Parameters.Nodes
		}
	}

	for _, param := range params {
		if param != nil && param.Kind == ast.KindParameter {
			paramDecl := param.AsParameterDeclaration()
			if paramDecl != nil && paramDecl.Name() != nil && getIdentifierName(paramDecl.Name()) == name {
				return true
			}
		}
	}

	return false
}

// hasShadowingVariable checks if a block contains a variable declaration with the given name
func hasShadowingVariable(node *ast.Node, name string) bool {
	if node.Kind != ast.KindBlock {
		return false
	}

	block := node.AsBlock()
	if block == nil || block.Statements == nil {
		return false
	}

	for _, stmt := range block.Statements.Nodes {
		if stmt != nil && stmt.Kind == ast.KindVariableStatement {
			varStmt := stmt.AsVariableStatement()
			if varStmt != nil && varStmt.DeclarationList != nil {
				declList := varStmt.DeclarationList.AsVariableDeclarationList()
				if declList != nil && declList.Declarations != nil {
					for _, decl := range declList.Declarations.Nodes {
						if decl != nil && decl.Kind == ast.KindVariableDeclaration {
							varDecl := decl.AsVariableDeclaration()
							if varDecl != nil && varDecl.Name() != nil {
								if getIdentifierName(varDecl.Name()) == name {
									return true
								}
							}
						}
					}
				}
			}
		}
	}

	return false
}

// findEnclosingScope finds the enclosing block or source file for a node
func findEnclosingScope(node *ast.Node) *ast.Node {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindSourceFile:
			return current
		case ast.KindBlock:
			return current
		case ast.KindModuleBlock:
			return current
		}
		current = current.Parent
	}
	return nil
}

// checkFuncReassignments finds all reassignments to the function name
func checkFuncReassignments(searchRoot *ast.Node, funcName string, funcNode *ast.Node, ctx *rule.RuleContext) {
	if funcName == "" {
		return
	}

	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		if node.Kind == ast.KindIdentifier && getIdentifierName(node) == funcName {
			// Skip if this is the function name declaration itself
			if node.Parent == funcNode {
				node.ForEachChild(func(child *ast.Node) bool {
					walk(child)
					return false
				})
				return
			}

			// Check if this is a write reference
			if isWriteReference(node) {
				// Check if the name is shadowed by a local variable
				if !isNameShadowed(node, funcName, funcNode, ctx) {
					ctx.ReportNode(node, buildMessage(funcName))
				}
			}
		}

		// Recursively check children
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}

	walk(searchRoot)
}

// NoFuncAssignRule disallows reassigning function declarations
var NoFuncAssignRule = rule.CreateRule(rule.Rule{
	Name: "no-func-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// Check function declarations
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl == nil || funcDecl.Name() == nil {
					return
				}

				funcName := getIdentifierName(funcDecl.Name())
				if funcName == "" {
					return
				}

				// Search enclosing scope for write references to funcName
				searchRoot := findEnclosingScope(node)
				if searchRoot == nil {
					return
				}

				checkFuncReassignments(searchRoot, funcName, node, &ctx)
			},

			// Check named function expressions
			ast.KindFunctionExpression: func(node *ast.Node) {
				funcExpr := node.AsFunctionExpression()
				if funcExpr == nil || funcExpr.Name() == nil {
					return
				}

				funcName := getIdentifierName(funcExpr.Name())
				if funcName == "" {
					return
				}

				// For named function expressions, only check within the function body
				checkFuncReassignments(node, funcName, node, &ctx)
			},
		}
	},
})
