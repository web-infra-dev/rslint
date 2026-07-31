package require_to_throw_message

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var toThrowMatchers = map[string]bool{
	"toThrow":      true,
	"toThrowError": true,
}

func buildAddErrorMessage(matcherName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "addErrorMessage",
		Description: "Add an error message to " + matcherName + "()",
		Data: map[string]string{
			"matcherName": matcherName,
		},
	}
}

var RequireToThrowMessageRule = rule.Rule{
	Name:   "jest/require-to-throw-message",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					jestFnCall.Kind != utils.JestFnTypeExpect ||
					slices.Contains(jestFnCall.Modifiers, "not") {
					return
				}

				matcherName := jestFnCall.Matcher
				if !toThrowMatchers[matcherName] || len(node.Arguments()) > 0 {
					return
				}

				ctx.ReportNode(jestFnCall.MatcherEntry.Node, buildAddErrorMessage(matcherName))
			},
		}
	},
}
