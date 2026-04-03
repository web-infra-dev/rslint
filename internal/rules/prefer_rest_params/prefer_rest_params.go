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

				// Skip property keys: obj.arguments, { arguments: 1 }, { arguments: a } in destructuring
				if ast.IsIdentifierName(node) {
					return
				}

				// Skip declaration names: function arguments(), var arguments, { arguments } in binding
				if isBindingOrDeclarationName(node) {
					return
				}

				// Find the enclosing non-arrow function. Arrow functions don't have
				// their own `arguments`; the reference passes through to the outer function.
				enclosingFunc := ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
					return ast.IsFunctionLikeDeclaration(n) && n.Kind != ast.KindArrowFunction
				})
				if enclosingFunc == nil {
					return
				}

				// Check if `arguments` is shadowed at function scope by a simple parameter
				// or a hoisted `var` declaration named "arguments"
				if isShadowedInFunction(enclosingFunc) {
					return
				}

				// Check if `arguments` is shadowed at block scope by let/const/using,
				// a for-in/for-of loop variable, or a catch clause parameter
				if isBlockScopeShadowed(node, enclosingFunc) {
					return
				}

				// Allow non-computed member access: arguments.length, arguments.callee
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

// isBindingOrDeclarationName checks if the identifier is the name of a binding
// or declaration — not a reference to the implicit `arguments` object.
// We intentionally don't use ast.IsDeclarationName here because it treats
// ShorthandPropertyAssignment ({ arguments }) as a declaration, but in that
// context the identifier IS a reference to the arguments variable.
func isBindingOrDeclarationName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Name() == node
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Name() == node
	case ast.KindFunctionDeclaration:
		return parent.AsFunctionDeclaration().Name() == node
	case ast.KindFunctionExpression:
		return parent.AsFunctionExpression().Name() == node
	case ast.KindBindingElement:
		return parent.AsBindingElement().Name() == node
	}
	return false
}

// isShadowedInFunction checks if the implicit `arguments` is shadowed at
// function scope by a simple (non-destructuring) parameter named "arguments"
// or a hoisted `var arguments` declaration anywhere in the function body.
func isShadowedInFunction(funcNode *ast.Node) bool {
	paramList := funcNode.ParameterList()
	if paramList != nil {
		for _, param := range paramList.Nodes {
			name := param.Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "arguments" {
				return true
			}
		}
	}

	body := funcNode.Body()
	if body == nil {
		return false
	}

	// Walk the function body looking for `var arguments` declarations.
	// var is hoisted to function scope, so it shadows the implicit arguments
	// regardless of where in the function body it appears.
	found := false
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}
		// Don't recurse into nested functions (they have their own scope)
		if n != funcNode && ast.IsFunctionLikeDeclaration(n) {
			return
		}
		if n.Kind == ast.KindVariableDeclaration {
			parent := n.Parent
			// Must be inside a VariableDeclarationList (not a CatchClause) and non-block-scoped
			if parent != nil && parent.Kind == ast.KindVariableDeclarationList &&
				parent.Flags&ast.NodeFlagsBlockScoped == 0 {
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

// isBlockScopeShadowed checks if the identifier is inside a block-level scope
// that declares its own "arguments" binding, shadowing the implicit one.
// Covers: let/const/using in blocks, for-in/for-of loop variables, catch parameters.
func isBlockScopeShadowed(node *ast.Node, funcNode *ast.Node) bool {
	current := node.Parent
	for current != nil && current != funcNode {
		switch current.Kind {
		case ast.KindBlock, ast.KindModuleBlock:
			if blockHasBlockScopedArguments(current) {
				return true
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			if forLoopDeclaresArguments(current) {
				return true
			}
		case ast.KindCatchClause:
			if catchDeclaresArguments(current) {
				return true
			}
		}
		current = current.Parent
	}
	return false
}

// blockHasBlockScopedArguments checks if a block contains a let/const/using
// declaration named "arguments".
func blockHasBlockScopedArguments(block *ast.Node) bool {
	for _, stmt := range block.Statements() {
		if stmt == nil || stmt.Kind != ast.KindVariableStatement {
			continue
		}
		varStmt := stmt.AsVariableStatement()
		if varStmt == nil || varStmt.DeclarationList == nil {
			continue
		}
		if varStmt.DeclarationList.Flags&ast.NodeFlagsBlockScoped == 0 {
			continue
		}
		vdl := varStmt.DeclarationList.AsVariableDeclarationList()
		if vdl == nil || vdl.Declarations == nil {
			continue
		}
		for _, decl := range vdl.Declarations.Nodes {
			varDecl := decl.AsVariableDeclaration()
			if varDecl != nil && varDecl.Name() != nil &&
				varDecl.Name().Kind == ast.KindIdentifier &&
				varDecl.Name().Text() == "arguments" {
				return true
			}
		}
	}
	return false
}

// forLoopDeclaresArguments checks if a for-in/for-of loop declares a variable
// named "arguments" (e.g., `for (let arguments of [])`).
func forLoopDeclaresArguments(forNode *ast.Node) bool {
	forStmt := forNode.AsForInOrOfStatement()
	if forStmt == nil || forStmt.Initializer == nil ||
		forStmt.Initializer.Kind != ast.KindVariableDeclarationList {
		return false
	}
	declList := forStmt.Initializer.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return false
	}
	for _, decl := range declList.Declarations.Nodes {
		varDecl := decl.AsVariableDeclaration()
		if varDecl != nil && varDecl.Name() != nil &&
			varDecl.Name().Kind == ast.KindIdentifier &&
			varDecl.Name().Text() == "arguments" {
			return true
		}
	}
	return false
}

// catchDeclaresArguments checks if a catch clause has a parameter named "arguments".
func catchDeclaresArguments(catchNode *ast.Node) bool {
	catchClause := catchNode.AsCatchClause()
	if catchClause == nil || catchClause.VariableDeclaration == nil {
		return false
	}
	varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
	return varDecl != nil && varDecl.Name() != nil &&
		varDecl.Name().Kind == ast.KindIdentifier &&
		varDecl.Name().Text() == "arguments"
}

// isNonComputedMemberAccess checks if the node is the object of a property
// access (e.g., arguments.length). ESLint allows this pattern.
func isNonComputedMemberAccess(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindPropertyAccessExpression {
		return parent.AsPropertyAccessExpression().Expression == node
	}
	return false
}
