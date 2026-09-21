package no_interpolation_in_snapshots

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNoInterpolationMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noInterpolation",
		Description: "Do not use string interpolation inside of snapshots",
	}
}

var NoInterpolationInSnapshotsRule = rule.Rule{
	Name:   "jest/no-interpolation-in-snapshots",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := utils.GetJestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := analysis.ParseExpectCall(node)

				if jestFnCall == nil ||
					!utils.INLINE_SNAPSHOT_MATCHERS[jestFnCall.Matcher] {
					return
				}

				for _, arg := range node.Arguments() {
					if arg := ast.SkipParentheses(arg); arg != nil && arg.Kind == ast.KindTemplateExpression {
						ctx.ReportNode(arg, buildNoInterpolationMessage())
					}
				}
			},
		}
	},
}
