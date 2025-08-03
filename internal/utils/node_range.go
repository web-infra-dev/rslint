package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
)

// SimpleNodeTextRange returns a text range directly from the node's position without trimming
func SimpleNodeTextRange(node *ast.Node) core.TextRange {
	return core.TextRange{}.WithPos(node.Pos()).WithEnd(node.End())
}
