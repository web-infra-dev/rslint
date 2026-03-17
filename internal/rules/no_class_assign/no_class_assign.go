package no_class_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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

// getClassSymbol gets the symbol for a class node
func getClassSymbol(classNode *ast.Node, ctx *rule.RuleContext) *ast.Symbol {
	if ctx.TypeChecker == nil {
		return nil
	}

	switch classNode.Kind {
	case ast.KindClassDeclaration:
		classDecl := classNode.AsClassDeclaration()
		if classDecl != nil && classDecl.Name() != nil {
			return ctx.TypeChecker.GetSymbolAtLocation(classDecl.Name())
		}
	case ast.KindClassExpression:
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
			if utils.IsWriteReference(node) {
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
