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
// anywhere inside the given function node. var declarations are hoisted to function scope,
// so `var arguments` in any nested block still shadows the built-in arguments object.
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

	// Recursively check for var declarations named "arguments" in the function body.
	body := funcNode.Body()
	if body == nil {
		return false
	}

	found := false
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}

		// Don't recurse into nested functions (they have their own arguments)
		if n != funcNode {
			switch n.Kind {
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
				ast.KindArrowFunction, ast.KindMethodDeclaration,
				ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
				return
			}
		}

		// Only var declarations shadow the implicit arguments (hoisted to function scope).
		// let/const are block-scoped and do NOT shadow function-level arguments.
		if n.Kind == ast.KindVariableDeclaration {
			// Check parent VariableDeclarationList — only var (not let/const/using)
			parent := n.Parent
			if parent != nil && parent.Flags&ast.NodeFlagsBlockScoped == 0 {
				varDecl := n.AsVariableDeclaration()
				if varDecl != nil && varDecl.Name() != nil &&
					varDecl.Name().Kind == ast.KindIdentifier &&
					varDecl.Name().Text() == "arguments" {
					found = true
					return
				}
			}
		}

		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(body)
	return found
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
