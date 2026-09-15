package prefer_add_event_listener_options

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "prefer-add-event-listener-options"

func preferOptionsMessage(value, replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: "Prefer `" + replacement + "` over `" + value + "`.",
		Data: map[string]string{
			"replacement": replacement,
			"value":       value,
		},
	}
}

// PreferAddEventListenerOptionsRule prefers the options-object form of
// addEventListener over its legacy boolean useCapture argument.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-add-event-listener-options.js
var PreferAddEventListenerOptionsRule = rule.Rule{
	Name:   "unicorn/prefer-add-event-listener-options",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		argumentCount := 3
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "addEventListener",
					ArgumentsLength:     &argumentCount,
					RejectSpreadElement: true,
				})
				if !ok {
					return
				}

				// ESTree removes source parentheses and JavaScript JSDoc casts,
				// while authored TypeScript assertions remain visible to upstream.
				option := utils.ESTreeRuntimeExpression(call.Call.Arguments()[2])
				if option == nil {
					return
				}

				value := ""
				switch option.Kind {
				case ast.KindTrueKeyword:
					value = "true"
				case ast.KindFalseKeyword:
					value = "false"
				default:
					return
				}

				replacement := "{capture: " + value + "}"
				ctx.ReportNodeWithDeferredFixes(option, preferOptionsMessage(value, replacement), func() []rule.RuleFix {
					return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, option, replacement)}
				})
			},
		}
	},
}
