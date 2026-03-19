package no_obj_calls

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var nonCallableGlobals = map[string]bool{
	"Math": true, "JSON": true, "Reflect": true, "Atomics": true, "Intl": true,
}

// https://eslint.org/docs/latest/rules/no-obj-calls
var NoObjCallsRule = rule.Rule{
	Name: "no-obj-calls",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkCallee := func(node *ast.Node, calleeNode *ast.Node) {
			if calleeNode.Kind == ast.KindIdentifier {
				name := calleeNode.AsIdentifier().Text
				if nonCallableGlobals[name] && !isShadowed(calleeNode, name) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedCall",
						Description: fmt.Sprintf("'%s' is not a function.", name),
					})
				}
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				checkCallee(node, callExpr.Expression)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				checkCallee(node, newExpr.Expression)
			},
		}
	},
}

// isShadowed checks if an identifier is shadowed by a local variable, parameter,
// or function declaration in an enclosing scope (not global).
func isShadowed(node *ast.Node, name string) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindBlock, ast.KindCaseClause, ast.KindDefaultClause:
			if blockDeclaresName(current, name) {
				return true
			}
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindArrowFunction, ast.KindMethodDeclaration:
			if functionParamDeclaresName(current, name) {
				return true
			}
		case ast.KindVariableStatement:
			if varStatementDeclaresName(current, name) {
				return true
			}
		case ast.KindSourceFile:
			// Reached the top-level scope — not shadowed
			return false
		}
		current = current.Parent
	}
	return false
}

// blockDeclaresName checks if a block contains a variable/function declaration with the given name.
func blockDeclaresName(block *ast.Node, name string) bool {
	var statements *ast.NodeList
	switch block.Kind {
	case ast.KindBlock:
		b := block.AsBlock()
		if b != nil {
			statements = b.Statements
		}
	case ast.KindCaseClause, ast.KindDefaultClause:
		c := block.AsCaseOrDefaultClause()
		if c != nil {
			statements = c.Statements
		}
	}
	if statements == nil {
		return false
	}
	for _, stmt := range statements.Nodes {
		switch stmt.Kind {
		case ast.KindVariableStatement:
			if varStatementDeclaresName(stmt, name) {
				return true
			}
		case ast.KindFunctionDeclaration:
			fd := stmt.AsFunctionDeclaration()
			if fd != nil && fd.Name() != nil && fd.Name().Text() == name {
				return true
			}
		}
	}
	return false
}

// varStatementDeclaresName checks if a variable statement declares the given name.
func varStatementDeclaresName(node *ast.Node, name string) bool {
	vs := node.AsVariableStatement()
	if vs == nil || vs.DeclarationList == nil {
		return false
	}
	dl := vs.DeclarationList.AsVariableDeclarationList()
	if dl == nil || dl.Declarations == nil {
		return false
	}
	for _, decl := range dl.Declarations.Nodes {
		vd := decl.AsVariableDeclaration()
		if vd != nil && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier && vd.Name().Text() == name {
			return true
		}
	}
	return false
}

// functionParamDeclaresName checks if a function's parameters include the given name.
func functionParamDeclaresName(node *ast.Node, name string) bool {
	var params *ast.NodeList
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fd := node.AsFunctionDeclaration()
		if fd != nil {
			params = fd.Parameters
			// Also check function name itself
			if fd.Name() != nil && fd.Name().Text() == name {
				return true
			}
		}
	case ast.KindFunctionExpression:
		fe := node.AsFunctionExpression()
		if fe != nil {
			params = fe.Parameters
		}
	case ast.KindArrowFunction:
		af := node.AsArrowFunction()
		if af != nil {
			params = af.Parameters
		}
	case ast.KindMethodDeclaration:
		md := node.AsMethodDeclaration()
		if md != nil {
			params = md.Parameters
		}
	}
	if params == nil {
		return false
	}
	for _, p := range params.Nodes {
		pd := p.AsParameterDeclaration()
		if pd != nil && pd.Name() != nil && pd.Name().Kind == ast.KindIdentifier && pd.Name().Text() == name {
			return true
		}
	}
	return false
}
