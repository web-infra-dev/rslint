package code_path_analyzer

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

type CodePathAnalyzer struct {
	currentNode *ast.Node
}

// Does the process to enter a given AST node.
// This updates state of analysis and calls `enterNode` of the wrapped.
func (analyzer *CodePathAnalyzer) enterNode(node *ast.Node) {
	analyzer.currentNode = node

	// Updates the code path due to node's position in its parent node.
	if node.Parent != nil {
		analyzer.preprocess(node)
	}

	// Updates the code path.
	// And emits onCodePathStart/onCodePathSegmentStart events.
	// analyzer.processCodePathToEnter(node)

	analyzer.currentNode = nil
}

// Does the process to leave a given AST node.
// This updates state of analysis and calls `leaveNode` of the wrapped.
func (analyzer *CodePathAnalyzer) leaveNode(node *ast.Node) {
	analyzer.currentNode = node

	// analyzer.processCodePathTOExit(node)

	// analyzer.postprocess(node)

	analyzer.currentNode = nil
}

func (analyzer *CodePathAnalyzer) preprocess(node *ast.Node) {
}
