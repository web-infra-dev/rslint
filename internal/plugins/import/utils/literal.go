package utils

import "github.com/microsoft/typescript-go/shim/ast"

// IsESTreeStringLiteral reports whether node is a string-valued ESTree
// Literal. Parentheses are transparent, while tsgo's
// NoSubstitutionTemplateLiteral is deliberately excluded because ESTree
// represents it as TemplateLiteral instead.
func IsESTreeStringLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	return node != nil && node.Kind == ast.KindStringLiteral
}
