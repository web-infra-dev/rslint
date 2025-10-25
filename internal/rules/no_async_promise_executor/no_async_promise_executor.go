package no_async_promise_executor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildAsyncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "async",
		Description: "Promise executor functions should not be async.",
	}
}

// NoAsyncPromiseExecutorRule disallows using an async function as a Promise executor
var NoAsyncPromiseExecutorRule = rule.CreateRule(rule.Rule{
	Name: "no-async-promise-executor",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				if node == nil {
					return
				}

				// Check if it's a new Promise(...) expression
				expr := node.Expression()
				if expr == nil || expr.Kind != ast.KindIdentifier {
					return
				}

				if expr.Text() != "Promise" {
					return
				}

				// Get the arguments
				args := node.Arguments()
				if args == nil || len(args) == 0 {
					return
				}

				// Check the first argument (executor function)
				executor := args[0]
				if executor == nil {
					return
				}

				// Unwrap parentheses to get to the actual function
				actualExecutor := executor
				for actualExecutor != nil && actualExecutor.Kind == ast.KindParenthesizedExpression {
					actualExecutor = actualExecutor.Expression()
				}

				if actualExecutor == nil {
					return
				}

				// Check if the executor is an async function
				var isAsync bool
				var asyncKeywordNode *ast.Node

				switch actualExecutor.Kind {
				case ast.KindFunctionExpression:
					// Check for async keyword
					mods := actualExecutor.Modifiers()
					if mods != nil {
						for _, mod := range mods.Nodes {
							if mod != nil && mod.Kind == ast.KindAsyncKeyword {
								isAsync = true
								asyncKeywordNode = mod
								break
							}
						}
					}

				case ast.KindArrowFunction:
					// Arrow functions can be async
					mods := actualExecutor.Modifiers()
					if mods != nil {
						for _, mod := range mods.Nodes {
							if mod != nil && mod.Kind == ast.KindAsyncKeyword {
								isAsync = true
								asyncKeywordNode = mod
								break
							}
						}
					}
				}

				if isAsync && asyncKeywordNode != nil {
					ctx.ReportNode(asyncKeywordNode, buildAsyncMessage())
				}
			},
		}
	},
})
