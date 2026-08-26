package no_invalid_remove_event_listener

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-invalid-remove-event-listener"

var invalidListenerMessage = rule.RuleMessage{
	Id:          messageID,
	Description: "The listener argument should be a function reference.",
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-invalid-remove-event-listener.md
var NoInvalidRemoveEventListenerRule = rule.Rule{
	Name:   "unicorn/no-invalid-remove-event-listener",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		minimumArguments := 2

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				_, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "removeEventListener",
					MinimumArguments:    &minimumArguments,
					AllowOptionalMember: true,
				})
				if !ok {
					return
				}

				arguments := node.Arguments()
				if arguments[0].Kind == ast.KindSpreadElement {
					return
				}

				listener := ast.SkipParentheses(arguments[1])
				switch listener.Kind {
				case ast.KindArrowFunction, ast.KindFunctionExpression:
					ctx.ReportRange(utils.GetFunctionHeadLoc(ctx.SourceFile, listener), invalidListenerMessage)
				case ast.KindCallExpression:
					if ast.IsOptionalChain(listener) {
						return
					}
					bindCall, isBindCall := unicornutil.MatchDotMethodCall(listener, unicornutil.DotMethodCallOptions{
						Method: "bind",
					})
					if isBindCall {
						ctx.ReportNode(bindCall.Property, invalidListenerMessage)
					}
				}
			},
		}
	},
}
