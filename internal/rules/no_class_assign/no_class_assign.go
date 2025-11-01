package no_class_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder
func buildClassReassignmentMessage(className string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "classReassignment",
		Description: "'" + className + "' is a class.",
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
		// ++ and -- are write operations
		switch postfix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return postfix.Operand == node
		}

	case ast.KindPrefixUnaryExpression:
		prefix := parent.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}
		// ++ and -- are write operations
		switch prefix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return prefix.Operand == node
		}

	case ast.KindObjectBindingPattern:
		// In destructuring like {A} = obj, A is a write reference
		// ObjectBindingPattern is used in destructuring assignments
		// Check if this pattern is on the left side of an assignment
		return isBindingPatternInAssignment(parent)

	case ast.KindArrayBindingPattern:
		// In array destructuring like [A] = arr, A is a write reference
		return isBindingPatternInAssignment(parent)

	case ast.KindBindingElement:
		// Check if the binding element is part of a write context
		return isWriteReference(parent)

	case ast.KindShorthandPropertyAssignment:
		// In destructuring like {A} = obj or ({A} = obj), A is a write reference
		shorthand := parent.AsShorthandPropertyAssignment()
		if shorthand != nil && shorthand.Name() == node {
			return isInDestructuringAssignment(parent)
		}

	case ast.KindPropertyAssignment:
		// In destructuring like {b: A} = obj, A is a write reference
		propAssignment := parent.AsPropertyAssignment()
		if propAssignment != nil && propAssignment.Initializer == node {
			return isInDestructuringAssignment(parent)
		}

	case ast.KindArrayLiteralExpression:
		// In array destructuring like [A] = arr, A is a write reference
		return isInDestructuringAssignment(parent)

	case ast.KindObjectLiteralExpression:
		// In object destructuring like {A} = obj, A is a write reference
		return isInDestructuringAssignment(parent)

	case ast.KindParenthesizedExpression:
		// Unwrap parentheses and check the parent context
		return isWriteReference(parent)

	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		// Type assertions like (A as any) = 0
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

// getClassSymbol gets the symbol for a class node
func getClassSymbol(classNode *ast.Node, ctx *rule.RuleContext) *ast.Symbol {
	if ctx.TypeChecker == nil {
		return nil
	}

	if classNode.Kind == ast.KindClassDeclaration {
		classDecl := classNode.AsClassDeclaration()
		if classDecl != nil && classDecl.Name() != nil {
			return ctx.TypeChecker.GetSymbolAtLocation(classDecl.Name())
		}
	} else if classNode.Kind == ast.KindClassExpression {
		classExpr := classNode.AsClassExpression()
		if classExpr != nil && classExpr.Name() != nil {
			return ctx.TypeChecker.GetSymbolAtLocation(classExpr.Name())
		}
	}

	return nil
}

// isNameShadowed checks if an identifier references a different variable due to shadowing
func isNameShadowed(node *ast.Node, className string, classNode *ast.Node, ctx *rule.RuleContext) bool {
	if node == nil || ctx.TypeChecker == nil {
		return false
	}

	// Get the symbol at the identifier location
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}

	// Get the symbol of the class declaration
	var classSymbol = getClassSymbol(classNode, ctx)

	// If symbols are different, the name is shadowed
	if classSymbol != nil {
		return symbol != classSymbol
	}

	// Fallback: check if the identifier is within a scope that shadows the class name
	return isInShadowingScope(node, className, classNode)
}

// isInShadowingScope checks if a node is within a scope that shadows the class name
func isInShadowingScope(node *ast.Node, className string, classNode *ast.Node) bool {
	current := node.Parent
	for current != nil && current != classNode {
		switch current.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindConstructor,
			ast.KindGetAccessor,
			ast.KindSetAccessor:
			// Check if there's a parameter with the same name
			if hasShadowingParameter(current, className) {
				return true
			}

		case ast.KindBlock:
			// Check if there's a variable declaration with the same name
			if hasShadowingVariable(current, className) {
				return true
			}

		case ast.KindCatchClause:
			// Check if the catch variable has the same name
			catchClause := current.AsCatchClause()
			if catchClause != nil && catchClause.VariableDeclaration != nil {
				varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
				if varDecl != nil && varDecl.Name() != nil {
					if getIdentifierName(varDecl.Name()) == className {
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

// checkClassReassignments finds all reassignments to the class name
func checkClassReassignments(classNode *ast.Node, className string, ctx *rule.RuleContext) {
	if className == "" {
		return
	}

	// Find the scope to search - for class declarations, we need to search the parent scope
	// For class expressions, we only search within the class itself
	var searchRoot *ast.Node
	if classNode.Kind == ast.KindClassDeclaration {
		// For class declarations, search the enclosing block or source file
		searchRoot = findEnclosingScope(classNode)
	} else {
		// For class expressions, search within the class only
		searchRoot = classNode
	}

	if searchRoot == nil {
		return
	}

	// Walk the tree to find all identifiers with the class name
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		if node.Kind == ast.KindIdentifier && getIdentifierName(node) == className {
			// Skip if this is the class name declaration itself
			if node.Parent == classNode {
				// Continue walking children
				node.ForEachChild(func(child *ast.Node) bool {
					walk(child)
					return false
				})
				return
			}

			// Check if this is a write reference
			if isWriteReference(node) {
				// Check if the name is shadowed by a local variable
				if !isNameShadowed(node, className, classNode, ctx) {
					ctx.ReportNode(node, buildClassReassignmentMessage(className))
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

// NoClassAssignRule disallows reassigning class declarations
var NoClassAssignRule = rule.CreateRule(rule.Rule{
	Name: "no-class-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// Check class declarations
			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				if classDecl == nil || classDecl.Name() == nil {
					return
				}

				className := getIdentifierName(classDecl.Name())
				checkClassReassignments(node, className, &ctx)
			},

			// Check named class expressions
			ast.KindClassExpression: func(node *ast.Node) {
				classExpr := node.AsClassExpression()
				if classExpr == nil || classExpr.Name() == nil {
					return
				}

				// Only check named class expressions
				// For `let A = class A { ... }`, we need to check reassignments inside the class
				className := getIdentifierName(classExpr.Name())
				checkClassReassignments(node, className, &ctx)
			},
		}
	},
})
