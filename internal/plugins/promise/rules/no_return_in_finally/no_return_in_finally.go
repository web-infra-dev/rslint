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

// callbackBody returns the block body of a .finally() callback, or nil if fn
// isn't a function/arrow expression with a block body. Upstream only ever
// sees FunctionExpression/ArrowFunctionExpression here, since fn comes
// straight from the call's first argument.
func callbackBody(fn *ast.Node) *ast.Node {
	switch fn.Kind {
	case ast.KindFunctionExpression:
		return fn.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		return fn.AsArrowFunction().Body
	default:
		return nil
	}
}

var NoReturnInFinallyRule = rule.Rule{
	Name: "promise/no-return-in-finally",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if !promiseutil.IsMemberCall(node, "finally") {
					return
				}
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				fn := ast.SkipOuterExpressions(call.Arguments.Nodes[0], skipTransparent)
				body := callbackBody(fn)
				if body == nil || body.Kind != ast.KindBlock {
					return
				}
				// Mirrors upstream's body.body.some(...): only the callback's
				// direct top-level statements count; nested returns don't.
				for _, stmt := range body.AsBlock().Statements.Nodes {
					if stmt.Kind == ast.KindReturnStatement {
						callee := ast.SkipOuterExpressions(call.Expression, skipTransparent)
						ctx.ReportNode(callee.AsPropertyAccessExpression().Name(), buildMessage())
						return
					}
				}
			},
		}
	},
}
