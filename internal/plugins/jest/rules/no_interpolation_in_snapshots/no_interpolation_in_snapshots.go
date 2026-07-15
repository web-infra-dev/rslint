package no_interpolation_in_snapshots

import (
	"github.com/microsoft/typescript-go/shim/ast"
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
	Name: "jest/no-interpolation-in-snapshots",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeExpect {
					return
				}

				if !utils.INLINE_SNAPSHOT_MATCHERS[jestFnCall.Matcher] {
					return
				}

				for _, arg := range node.Arguments() {
					if arg != nil && arg.Kind == ast.KindTemplateExpression {
						ctx.ReportNode(arg, buildNoInterpolationMessage())
					}
				}
			},
		}
	},
}
