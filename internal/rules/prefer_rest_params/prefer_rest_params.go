package prefer_rest_params

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/prefer-rest-params
var PreferRestParamsRule = rule.Rule{
	Name: "prefer-rest-params",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if node.Text() != "arguments" {
					return
				}

				// Skip if this is a property name (e.g., obj.arguments)
				if isPropertyName(node) {
					return
				}

				// Skip if this is a declaration name (var arguments, function arguments(params))
				if isDeclarationName(node) {
					return
				}

				// Find the enclosing function (not arrow)
				enclosingFunc := findEnclosingFunction(node)
				if enclosingFunc == nil {
					// arguments at module/program level - not in a function
					return
				}

				// If enclosing function is arrow function, skip (arrows don't have their own arguments)
				if enclosingFunc.Kind == ast.KindArrowFunction {
					return
				}

				// Check if `arguments` is shadowed by a parameter or local var
				if isShadowedInFunction(enclosingFunc) {
					return
				}

				// Check if this is a non-computed member access (arguments.length is OK)
				if isNonComputedMemberAccess(node) {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "preferRestParams",
					Description: "Use the rest parameters instead of 'arguments'.",
				})
			},
		}
	},
}

// isPropertyName checks if the node is the property name part of a property access (e.g., obj.arguments).
func isPropertyName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindPropertyAccessExpression {
		propAccess := parent.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Name() == node {
			return true
		}
	}
	return false
}

// isDeclarationName checks if the node is the name of a declaration (parameter, variable, function).
func isDeclarationName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindParameter:
		param := parent.AsParameterDeclaration()
		if param != nil && param.Name() == node {
			return true
		}
	case ast.KindVariableDeclaration:
		varDecl := parent.AsVariableDeclaration()
		if varDecl != nil && varDecl.Name() == node {
			return true
		}
	case ast.KindFunctionDeclaration:
		funcDecl := parent.AsFunctionDeclaration()
		if funcDecl != nil && funcDecl.Name() == node {
			return true
		}
	}
	return false
}

// findEnclosingFunction walks up the AST to find the nearest enclosing function-like node.
// Returns nil if none is found.
func findEnclosingFunction(node *ast.Node) *ast.Node {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return current
		}
		current = current.Parent
	}
	return nil
}

// isShadowedInFunction checks if `arguments` is shadowed by a parameter or a var declaration
// inside the given function node.
func isShadowedInFunction(funcNode *ast.Node) bool {
	// Check parameters
	paramList := funcNode.ParameterList()
	if paramList != nil {
		for _, param := range paramList.Nodes {
			name := param.Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "arguments" {
				return true
			}
		}
	}

	// Check for var declarations in body
	body := funcNode.Body()
	if body != nil && body.Kind == ast.KindBlock {
		for _, stmt := range body.Statements() {
			if stmt.Kind == ast.KindVariableStatement {
				varStmt := stmt.AsVariableStatement()
				if varStmt != nil && varStmt.DeclarationList != nil {
					declList := varStmt.DeclarationList.AsVariableDeclarationList()
					if declList != nil && declList.Declarations != nil {
						for _, decl := range declList.Declarations.Nodes {
							varDecl := decl.AsVariableDeclaration()
							if varDecl != nil && varDecl.Name() != nil && varDecl.Name().Kind == ast.KindIdentifier {
								if varDecl.Name().Text() == "arguments" {
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

// isNonComputedMemberAccess checks if the node is the object of a non-computed property access
// (e.g., arguments.length). In that case, the usage is OK and should not be flagged.
func isNonComputedMemberAccess(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindPropertyAccessExpression {
		propAccess := parent.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Expression == node {
			return true
		}
	}
	return false
}
