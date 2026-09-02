package unicornutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func isIdentifierNamed(node *ast.Node, name string) bool {
	return node != nil && ast.IsIdentifier(node) && node.AsIdentifier().Text == name
}

// IsArrayPrototypeProperty mirrors unicorn's isArrayPrototypeProperty helper.
// It intentionally accepts only dotted, non-optional member access.
// Parentheses and JavaScript JSDoc casts are transparent at every segment;
// authored TypeScript assertions remain visible.
func IsArrayPrototypeProperty(node *ast.Node, property string) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !ast.IsPropertyAccessExpression(node) {
		return false
	}

	propertyAccess := node.AsPropertyAccessExpression()
	if propertyAccess == nil || ast.IsOptionalChainRoot(node) ||
		!isIdentifierNamed(propertyAccess.Name(), property) {
		return false
	}

	object := utils.ESTreeRuntimeExpression(propertyAccess.Expression)
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
	if prototypeAccess == nil || ast.IsOptionalChainRoot(object) ||
		!isIdentifierNamed(prototypeAccess.Name(), "prototype") {
		return false
	}

	return isIdentifierNamed(utils.ESTreeRuntimeExpression(prototypeAccess.Expression), "Array")
}
