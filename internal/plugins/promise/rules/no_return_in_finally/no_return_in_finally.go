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
	return parent != nil && promiseutil.IsMemberCall(parent, "finally")
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
					ctx.ReportNode(node, buildMessage())
				}
			},
		}
	},
}
