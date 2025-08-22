package code_path_analyzer

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

type CodePathAnalyzer struct {
	currentNode *ast.Node
	codePath    *CodePath
	idGenerator *IdGenerator
}

func NewCodePathAnalyzer() *CodePathAnalyzer {
	return &CodePathAnalyzer{
		currentNode: nil,
		codePath:    nil,
		idGenerator: NewIdGenerator("s"),
	}
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
	analyzer.processCodePathToEnter(node)

	analyzer.currentNode = nil
}

// Does the process to leave a given AST node.
// This updates state of analysis and calls `leaveNode` of the wrapped.
func (analyzer *CodePathAnalyzer) leaveNode(node *ast.Node) {
	analyzer.currentNode = node

	analyzer.processCodePathTOExit(node)

	analyzer.postprocess(node)

	analyzer.currentNode = nil
}

// Updates the code path due to the position of a given node in the parent node thereof.
//
// For example, if the node is `parent.consequent`, this creates a fork from the current path.
func (analyzer *CodePathAnalyzer) preprocess(node *ast.Node) {
	codePath := analyzer.codePath
	state := codePath.state
	parent := node.Parent

	switch parent.Kind {
	// The `arguments.length == 0` case is in `postprocess` function.
	case ast.KindCallExpression:
		{
			if ast.IsOptionalChain(parent) && len(parent.Arguments()) >= 1 && parent.Arguments()[0] == node {
				state.MakeOptionalRight()
			}
		}
	}
}

func (analyzer *CodePathAnalyzer) processCodePathToEnter(node *ast.Node) {
	// !!!
}

func (analyzer *CodePathAnalyzer) postprocess(node *ast.Node) {
	// !!!
}

func (analyzer *CodePathAnalyzer) processCodePathTOExit(node *ast.Node) {
	// !!!
}
