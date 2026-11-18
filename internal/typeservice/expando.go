package typeservice

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// An expando initializer contributes properties to the symbol of its assignment
// or declaration. The native binder owns that symbol, including merged names.
func expandoSymbol(c *checker.Checker, node *ast.Node, allowDeclaration bool) *ast.Symbol {
	parent := node.Parent
	if parent == nil {
		return nil
	}
	var name, decl *ast.Node
	switch {
	case ast.IsVariableDeclaration(parent) && parent.Initializer() == node:
		if !ast.IsInJSFile(node) && (!ast.IsVarConstLike(parent) || !ast.IsFunctionLikeDeclaration(node)) {
			return nil
		}
		name, decl = parent.Name(), parent
	case ast.IsBinaryExpression(parent):
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken:
			if allowDeclaration || binary.Right == node {
				name, decl = binary.Left, binary.Left
			}
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			outer := parent.Parent
			if outer == nil {
				return nil
			}
			if ast.IsVariableDeclaration(outer) && outer.Initializer() == parent {
				name, decl = outer.Name(), outer
			}
			if ast.IsBinaryExpression(outer) && outer.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && (allowDeclaration || outer.AsBinaryExpression().Right == parent) {
				name = outer.AsBinaryExpression().Left
				decl = name
			}
			if name == nil || !ast.IsBindableStaticNameExpression(name, false) || !sameExpandoName(name, binary.Left) {
				return nil
			}
		}
	case allowDeclaration && ast.IsFunctionDeclaration(node):
		name, decl = node.Name(), node
	}
	if name == nil || decl == nil {
		return nil
	}
	if !allowDeclaration {
		switch node.Kind {
		case ast.KindCallExpression:
			if !ast.IsFunctionExpressionOrArrowFunction(ast.SkipParentheses(node.Expression())) {
				return nil
			}
		case ast.KindFunctionExpression, ast.KindClassExpression, ast.KindArrowFunction:
		case ast.KindObjectLiteralExpression:
			if len(node.Properties()) != 0 && !ast.IsPrototypeAccess(name) {
				return nil
			}
		default:
			return nil
		}
	}
	return checker.Checker_getSymbolOfNode(c, decl)
}

func sameExpandoName(name, value *ast.Node) bool {
	if ast.IsPropertyNameLiteral(name) && ast.IsPropertyNameLiteral(value) {
		return name.Text() == value.Text()
	}
	if !ast.IsPropertyAccessExpression(value) && !ast.IsLiteralLikeElementAccess(value) {
		return false
	}
	right := expandoAccessName(value)
	if ast.IsMemberName(name) {
		base := value.Expression()
		if base.Kind == ast.KindThisKeyword || ast.IsIdentifier(base) && (base.Text() == "window" || base.Text() == "self" || base.Text() == "global") {
			return sameExpandoName(name, right)
		}
	}
	return (ast.IsPropertyAccessExpression(name) || ast.IsLiteralLikeElementAccess(name)) && expandoAccessName(name).Text() == right.Text() && sameExpandoName(name.Expression(), value.Expression())
}

func expandoAccessName(node *ast.Node) *ast.Node {
	if ast.IsElementAccessExpression(node) {
		return node.AsElementAccessExpression().ArgumentExpression
	}
	return node.Name()
}
