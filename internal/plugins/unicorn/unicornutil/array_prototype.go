package unicornutil

import "github.com/microsoft/typescript-go/shim/ast"

func isIdentifierNamed(node *ast.Node, name string) bool {
	return node != nil && ast.IsIdentifier(node) && node.AsIdentifier().Text == name
}

// IsArrayPrototypeProperty mirrors unicorn's isArrayPrototypeProperty helper.
// It intentionally accepts only dotted, non-optional member access.
func IsArrayPrototypeProperty(node *ast.Node, property string) bool {
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsPropertyAccessExpression(node) {
		return false
	}

	propertyAccess := node.AsPropertyAccessExpression()
	if propertyAccess == nil || propertyAccess.QuestionDotToken != nil ||
		!isIdentifierNamed(propertyAccess.Name(), property) {
		return false
	}

	object := ast.SkipParentheses(propertyAccess.Expression)
	if object == nil {
		return false
	}
	if ast.IsEmptyArrayLiteral(object) {
		return true
	}

	if !ast.IsPropertyAccessExpression(object) {
		return false
	}
	prototypeAccess := object.AsPropertyAccessExpression()
	if prototypeAccess == nil || prototypeAccess.QuestionDotToken != nil ||
		!isIdentifierNamed(prototypeAccess.Name(), "prototype") {
		return false
	}

	return isIdentifierNamed(ast.SkipParentheses(prototypeAccess.Expression), "Array")
}
