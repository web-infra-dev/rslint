package prefer_called_with

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var preferCalledWithMatcherNames = map[string]bool{
	"toBeCalled":       true,
	"toHaveBeenCalled": true,
}

// Message builder

func buildPreferCalledWithMessage(matcherName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferCalledWith",
		Description: "Prefer " + matcherName + "With(/* expected args */)",
		Data: map[string]string{
			"matcherName": matcherName,
		},
	}
}

var PreferCalledWithRule = rule.Rule{
	Name: "jest/prefer-called-with",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					jestFnCall.Kind != jestUtils.JestFnTypeExpect ||
					slices.Contains(jestFnCall.Modifiers, "not") {
					return
				}

				matcherName := jestFnCall.Matcher
				if !preferCalledWithMatcherNames[matcherName] {
					return
				}

				reportNode := node
				if jestFnCall.MatcherEntry != nil && jestFnCall.MatcherEntry.Node != nil {
					reportNode = jestFnCall.MatcherEntry.Node
				}

				ctx.ReportNode(reportNode, buildPreferCalledWithMessage(matcherName))
			},
		}
	},
}
