package no_error_equal

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoErrorEqualRule = rule.Rule{
	Name:             "jest/no-error-equal",
	RequiresTypeInfo: true,
	Schema:           rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil || (parsed.Matcher != "toEqual" && parsed.Matcher != "toStrictEqual") {
					return
				}

				// The head must be called directly: expect(value).toEqual(...),
				// not a member such as expect.extend(...).toEqual(...).
				head := parsed.Head.Local.Node
				if head == nil {
					return
				}
				headCall := ast.WalkUpParenthesizedExpressions(head.Parent)
				if headCall == nil || headCall.Kind != ast.KindCallExpression ||
					ast.SkipParentheses(headCall.AsCallExpression().Expression) != head {
					return
				}

				// A missing argument has no type to inspect and is never an Error.
				arguments := headCall.AsCallExpression().Arguments
				if arguments == nil || len(arguments.Nodes) == 0 {
					return
				}

				argument := arguments.Nodes[0]
				if !utils.IsBuiltinSymbolLike(ctx.Program(), ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(argument), "Error") {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "equalError",
					Description: "Avoid using equality matchers to check errors",
				})
			},
		}
	},
}
