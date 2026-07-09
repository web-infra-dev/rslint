package void_use_memo

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const (
	missingReturnReason      = "useMemo() callbacks must return a value"
	missingReturnDescription = "This useMemo() callback doesn't return a value. useMemo() is for computing and caching values, not for arbitrary side effects"
	unusedResultReason       = "useMemo() result is unused"
	unusedResultDescription  = "This useMemo() value is unused. useMemo() is for computing and caching values, not for arbitrary side effects"
)

type voidUseMemoState struct {
	ctx rule.RuleContext
}

// VoidUseMemoRule is the rslint port of upstream `react-hooks/void-use-memo`.
//
// Upstream emits this rule from React Compiler's VoidUseMemo diagnostic
// category. This port validates the published local contract: useMemo
// callbacks must contain an explicit or implicit return, and a returned
// useMemo value must not be discarded as a bare expression statement.
var VoidUseMemoRule = rule.Rule{
	Name: "react-hooks/void-use-memo",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		state := &voidUseMemoState{ctx: ctx}
		return rule.RuleListeners{
			ast.KindCallExpression: state.processCallExpression,
		}
	},
}

func (state *voidUseMemoState) processCallExpression(node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil || !react_hooksutil.IsManualUseMemoCallee(call.Expression, state.ctx.TypeChecker) {
		return
	}
	args := call.Arguments
	if args == nil || len(args.Nodes) == 0 {
		return
	}

	callback := ast.SkipParentheses(args.Nodes[0])
	if callback == nil || callback.Kind == ast.KindSpreadElement || !ast.IsFunctionExpressionOrArrowFunction(callback) {
		return
	}

	if !hasUseMemoReturn(callback) {
		state.ctx.ReportNode(callback, buildVoidUseMemoMessage("missingReturn", missingReturnReason, missingReturnDescription))
		return
	}
	if isUnusedUseMemoResult(node) {
		state.ctx.ReportNode(reportableCallee(call.Expression), buildVoidUseMemoMessage("unusedResult", unusedResultReason, unusedResultDescription))
	}
}

func hasUseMemoReturn(callback *ast.Node) bool {
	if callback == nil {
		return false
	}
	body := react_hooksutil.GetFunctionBody(callback)
	if body == nil {
		return false
	}
	// React Compiler treats arrow expression bodies as implicit returns.
	if callback.Kind == ast.KindArrowFunction {
		if body.Kind != ast.KindBlock {
			return true
		}
	}

	// React Compiler's HIR-based terminal scan does not count returns from the
	// try block of a try/finally statement or from its finally block.
	return ast.ForEachReturnStatement(body, func(stmt *ast.Node) bool {
		return !isIgnoredTryFinallyReturn(stmt, callback)
	})
}

func isIgnoredTryFinallyReturn(ret *ast.Node, callback *ast.Node) bool {
	child := ret
	for parent := ret.Parent; parent != nil && parent != callback; parent = parent.Parent {
		if react_hooksutil.IsFunctionLikeContainer(parent) {
			return false
		}
		if parent.Kind == ast.KindTryStatement {
			tryStmt := parent.AsTryStatement()
			if tryStmt != nil && tryStmt.FinallyBlock != nil {
				if tryStmt.TryBlock != nil && tryStmt.TryBlock.AsNode() == child {
					return true
				}
				if tryStmt.FinallyBlock.AsNode() == child {
					return true
				}
			}
		}
		child = parent
	}
	return false
}

func isUnusedUseMemoResult(call *ast.Node) bool {
	current := call
	for {
		child, parent := walkUpResultWrappers(current, current == call)
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindExpressionStatement:
			return current == call
		case ast.KindBinaryExpression:
			// In a comma expression, every operand except the final one is
			// discarded even when the comma expression's final result is used.
			binary := parent.AsBinaryExpression()
			if binary == nil || binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindCommaToken {
				return false
			}
			if binary.Left == child {
				return true
			}
			if binary.Right == child {
				current = parent
				continue
			}
			return false
		default:
			return false
		}
	}
}

func walkUpResultWrappers(node *ast.Node, allowNonNull bool) (*ast.Node, *ast.Node) {
	child := node
	parent := child.Parent
	kinds := ast.OEKParentheses
	if allowNonNull {
		kinds |= ast.OEKNonNullAssertions
	}
	for parent != nil && ast.IsOuterExpression(parent, kinds) {
		child = parent
		parent = child.Parent
	}
	return child, parent
}

func reportableCallee(callee *ast.Node) *ast.Node {
	if unwrapped := ast.SkipParentheses(callee); unwrapped != nil {
		return unwrapped
	}
	return callee
}

func buildVoidUseMemoMessage(id, reason, description string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          id,
		Description: fmt.Sprintf("Error: %s\n\n%s.", reason, description),
		Data: map[string]string{
			"description": description,
			"reason":      reason,
		},
	}
}
