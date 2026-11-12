package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsNonDOMNodeExpression mirrors Unicorn's syntax-only isNodeValueNotDomNode.
// Callers choose which wrappers to unwrap; identifiers other than `undefined`
// remain unknown, without type or scope inference.
func IsNonDOMNodeExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindArrayLiteralExpression, ast.KindArrowFunction, ast.KindClassExpression,
		ast.KindFunctionExpression, ast.KindObjectLiteralExpression, ast.KindTemplateExpression,
		ast.KindNoSubstitutionTemplateLiteral:
		return true
	}
	return utils.IsESTreeLiteralKind(node.Kind) || (ast.IsIdentifier(node) && node.Text() == "undefined")
}
