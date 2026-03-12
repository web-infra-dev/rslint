package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// FunctionReturnAnalysis holds the result of analyzing return behavior of a function.
type FunctionReturnAnalysis struct {
	EndReachable       bool // Whether the function's end is reachable (can fall through without return/throw)
	HasReturnWithValue bool // Whether any return statement returns a value
	HasEmptyReturn     bool // Whether any return statement is empty (return;)
}

// AnalyzeFunctionReturns analyzes a function node's return behavior using
// the binder's control flow graph and ForEachReturnStatement.
// The node must be a function-like node (FunctionDeclaration, FunctionExpression,
// ArrowFunction, Constructor, MethodDeclaration, GetAccessor, SetAccessor).
//
// The binder sets EndFlowNode on function bodies only when the function end is
// reachable (i.e., some code path falls through without returning or throwing).
// If EndFlowNode is nil and the body is present, all paths return or throw.
func AnalyzeFunctionReturns(node *ast.Node) FunctionReturnAnalysis {
	result := FunctionReturnAnalysis{
		EndReachable: true,
	}

	if node == nil {
		return result
	}

	body := node.Body()
	if body == nil {
		return result
	}

	// The binder sets NodeFlagsHasImplicitReturn when the function end is reachable.
	// This flag is set at the same time as EndFlowNode, but is simpler to check.
	result.EndReachable = node.Flags&ast.NodeFlagsHasImplicitReturn != 0

	// Scan return statements (ForEachReturnStatement skips nested functions)
	ast.ForEachReturnStatement(body, func(stmt *ast.Node) bool {
		if stmt.Expression() != nil {
			result.HasReturnWithValue = true
		} else {
			result.HasEmptyReturn = true
		}
		return false
	})

	return result
}

// IsFunctionEndReachable checks if a function's end is reachable (can fall through
// without a return or throw statement). Uses the binder's NodeFlagsHasImplicitReturn flag.
func IsFunctionEndReachable(node *ast.Node) bool {
	if node == nil {
		return true
	}
	return node.Flags&ast.NodeFlagsHasImplicitReturn != 0
}
