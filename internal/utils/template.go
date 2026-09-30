package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// TemplateElementRaw returns a template token's raw content and its range,
// excluding the backticks, closing brace, and interpolation delimiter.
func TemplateElementRaw(sourceFile *ast.SourceFile, node *ast.Node) (string, core.TextRange) {
	nodeRange := TrimNodeTextRange(sourceFile, node)
	endOffset := 1
	if node.Kind == ast.KindTemplateHead || node.Kind == ast.KindTemplateMiddle {
		endOffset = 2
	}
	contentRange := core.NewTextRange(nodeRange.Pos()+1, nodeRange.End()-endOffset)
	return sourceFile.Text()[contentRange.Pos():contentRange.End()], contentRange
}

// IsTaggedTemplateElement reports whether a template token belongs to a
// tagged template. Expressions inside interpolations are not tag contents.
func IsTaggedTemplateElement(node *ast.Node) bool {
	if node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression
	}
	if node.Kind == ast.KindTemplateHead {
		return node.Parent != nil &&
			node.Parent.Parent != nil &&
			node.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
	}
	return node.Parent != nil &&
		node.Parent.Parent != nil &&
		node.Parent.Parent.Parent != nil &&
		node.Parent.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
}
