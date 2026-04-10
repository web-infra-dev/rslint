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

// CanBlockThrow checks if a block can throw before reaching a non-throwing
// terminal. Used to determine if a catch clause is reachable.
//
// A block "can throw" if it contains any statement that may raise an exception
// before control reaches a guaranteed non-throwing terminal (break, continue,
// or return without expression). Specifically:
//   - break / continue: non-throwing terminals → returns false
//   - return (no expression): non-throwing → returns false
//   - return (with expression): expression evaluation may throw → returns true
//   - throw: always throws → returns true
//   - empty statement: no effect → continues checking
//   - nested block: recurses
//   - try with finally that terminates: finally overrides → returns false
//   - any other statement (expression, if, for, etc.): may throw → returns true
func CanBlockThrow(block *ast.Node) bool {
	statements := block.Statements()
	if len(statements) == 0 {
		return false
	}
	for _, stmt := range statements {
		switch stmt.Kind {
		case ast.KindBreakStatement, ast.KindContinueStatement:
			return false
		case ast.KindReturnStatement:
			rs := stmt.AsReturnStatement()
			return rs != nil && rs.Expression != nil
		case ast.KindThrowStatement:
			return true
		case ast.KindEmptyStatement:
			continue
		case ast.KindBlock:
			return CanBlockThrow(stmt)
		case ast.KindTryStatement:
			ts := stmt.AsTryStatement()
			if ts != nil && ts.FinallyBlock != nil && BlockEndsWithTerminal(ts.FinallyBlock) {
				return false
			}
			return true
		default:
			return true
		}
	}
	return true
}

// BlockEndsWithTerminal checks if a block's last statement is a control flow
// terminal (break/return/throw/continue), possibly nested in inner blocks.
func BlockEndsWithTerminal(block *ast.Node) bool {
	nodes := block.Statements()
	if len(nodes) == 0 {
		return false
	}
	last := nodes[len(nodes)-1]
	switch last.Kind {
	case ast.KindBreakStatement, ast.KindContinueStatement,
		ast.KindReturnStatement, ast.KindThrowStatement:
		return true
	case ast.KindBlock:
		return BlockEndsWithTerminal(last)
	}
	return false
}
