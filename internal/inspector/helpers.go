package inspector

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
)

// FindNodeAtPosition finds a node at the specified position in a source file.
// If end > 0, it uses exact matching (nodePos == start && nodeEnd == end)
// If kind > 0, it also checks that the node kind matches
// Otherwise, it finds the deepest node containing the position
func FindNodeAtPosition(sourceFile *ast.SourceFile, start int, end int, kind int) *ast.Node {
	sfNode := sourceFile.AsNode()

	// If looking for SourceFile by kind, return it directly
	if kind > 0 && ast.Kind(kind) == ast.KindSourceFile {
		return sfNode
	}

	// Check if we're looking for the SourceFile itself by position
	if end > 0 && sfNode.Pos() == start && sfNode.End() == end {
		return sfNode
	}

	var result *ast.Node
	var kindMatch *ast.Node // Track the node matching the requested kind

	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}

		nodePos := node.Pos()
		nodeEnd := node.End()

		// Exact matching mode: when end is specified, find node with exact pos and end
		if end > 0 {
			if nodePos == start && nodeEnd == end {
				// If kind is specified, check it matches
				if kind > 0 {
					if int(node.Kind) == kind {
						result = node
						return true // Found exact match with kind, stop traversal
					}
					// Kind doesn't match, continue searching in children
				} else {
					result = node
					return true // Found exact match, stop traversal
				}
			}
			// Continue searching in children if this node contains the target range
			if nodePos <= start && nodeEnd >= end {
				node.ForEachChild(visit)
			}
			return false
		}

		// Range matching mode: find deepest node containing the position
		if nodePos <= start && start < nodeEnd {
			result = node
			// Check if this node matches the requested kind
			if kind > 0 && int(node.Kind) == kind {
				kindMatch = node
			}
			// Continue to find deeper nodes
			node.ForEachChild(visit)
		}
		return false
	}

	// Start traversal from source file's children
	sfNode.ForEachChild(visit)

	// If kind was specified and we found a match, return it instead of the deepest node
	if kind > 0 && kindMatch != nil {
		return kindMatch
	}

	return result
}

// GetTypeAtNode gets the type of a node using the type checker
func GetTypeAtNode(c *checker.Checker, node *ast.Node) *checker.Type {
	if node == nil {
		return nil
	}

	// Try to get type at location
	t := c.GetTypeAtLocation(node)
	if t != nil {
		return t
	}

	// For identifiers, try to get symbol first
	if node.Kind == ast.KindIdentifier {
		symbol := c.GetSymbolAtLocation(node)
		if symbol != nil {
			return c.GetTypeOfSymbol(symbol)
		}
	}

	return nil
}

// GetFlowNodeOfNode gets the FlowNode associated with an AST node
func GetFlowNodeOfNode(node *ast.Node) *ast.FlowNode {
	if node == nil {
		return nil
	}

	flowNodeData := node.FlowNodeData()
	if flowNodeData != nil {
		return flowNodeData.FlowNode
	}
	return nil
}

// GetSignatureOfNode gets the signature associated with a node
func GetSignatureOfNode(c *checker.Checker, node *ast.Node) *checker.Signature {
	if node == nil {
		return nil
	}

	switch node.Kind {
	// Call/New expressions - get resolved signature
	case ast.KindCallExpression, ast.KindNewExpression:
		return c.GetResolvedSignature(node)

	// Function-like declarations - get signature directly from declaration
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindMethodSignature,
		ast.KindConstructor,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindCallSignature,
		ast.KindConstructSignature,
		ast.KindFunctionType,
		ast.KindConstructorType,
		ast.KindIndexSignature:
		return c.GetSignatureFromDeclaration(node)
	}

	return nil
}
