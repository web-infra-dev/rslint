package no_return_in_finally

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const skipTransparent = ast.OEKParentheses

func buildMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noReturnInFinally",
		Description: "No return in finally",
	}
}

// isFinallyCallback reports whether fn is directly passed as a callback to a
// .finally() call. Parentheses around fn are skipped.
func isFinallyCallback(fn *ast.Node) bool {
	parent := fn.Parent
	for parent != nil && ast.IsOuterExpression(parent, skipTransparent) {
		parent = parent.Parent
	}
	if parent == nil || !promiseutil.IsMemberCall(parent, "finally") {
		return false
	}
	callExpr := parent.AsCallExpression()
	if callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
		return false
	}
	firstArg := ast.SkipOuterExpressions(callExpr.Arguments.Nodes[0], skipTransparent)
	unwrappedFn := ast.SkipOuterExpressions(fn, skipTransparent)
	return unwrappedFn == firstArg
}

var NoReturnInFinallyRule = rule.Rule{
	Name: "promise/no-return-in-finally",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				fn := promiseutil.NearestFunctionBoundary(node)
				if fn == nil {
					return
				}
				if isFinallyCallback(fn) {
					var body *ast.Node
					switch fn.Kind {
					case ast.KindFunctionExpression:
						body = fn.AsFunctionExpression().Body
					case ast.KindFunctionDeclaration:
						body = fn.AsFunctionDeclaration().Body
					case ast.KindMethodDeclaration:
						body = fn.AsMethodDeclaration().Body
					case ast.KindGetAccessor:
						body = fn.AsGetAccessorDeclaration().Body
					case ast.KindSetAccessor:
						body = fn.AsSetAccessorDeclaration().Body
					case ast.KindConstructor:
						body = fn.AsConstructorDeclaration().Body
					case ast.KindArrowFunction:
						body = fn.AsArrowFunction().Body
					}
					
					// ESLint's promise/no-return-in-finally rule only checks top-level statements
					// in the function's block body. It ignores nested returns.
					if body != nil && body.Kind == ast.KindBlock {
						if node.Parent != body {
							return
						}
						ctx.ReportNode(node, buildMessage())
					}
				}
			},
		}
	},
}
