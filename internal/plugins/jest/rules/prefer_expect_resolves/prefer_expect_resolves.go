package prefer_expect_resolves

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func buildExpectResolvesErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectResolves",
		Description: "Use `await expect(...).resolves instead",
	}
}

var PreferExpectResolvesRule = rule.Rule{
	Name: "jest/prefer-expect-resolves",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != jestUtils.JestFnTypeExpect {
					return
				}

				expectCall := jestFnCall.Head.Local.Node.Parent
				if expectCall == nil || expectCall.Kind != ast.KindCallExpression {
					return
				}

				args := expectCall.Arguments()
				if len(args) == 0 {
					return
				}

				awaitNode := ast.SkipParentheses(args[0])
				if awaitNode == nil || awaitNode.Kind != ast.KindAwaitExpression {
					return
				}

				awaitExpr := awaitNode.AsAwaitExpression()
				if awaitExpr == nil || awaitExpr.Expression == nil {
					return
				}

				sourceFile := ctx.SourceFile
				awaitRange := rslintUtils.TrimNodeTextRange(sourceFile, awaitNode)
				argumentRange := rslintUtils.TrimNodeTextRange(sourceFile, awaitExpr.Expression)

				fixes := []rule.RuleFix{
					rule.RuleFixInsertBefore(sourceFile, expectCall, "await "),
					rule.RuleFixRemoveRange(core.NewTextRange(awaitRange.Pos(), argumentRange.Pos())),
				}
				if !slices.Contains(jestFnCall.Modifiers, "resolves") && !slices.Contains(jestFnCall.Modifiers, "rejects") {
					fixes = append(fixes, rule.RuleFixInsertAfter(expectCall, ".resolves"))
				}

				ctx.ReportNodeWithFixes(awaitNode, buildExpectResolvesErrorMessage(), fixes...)
			},
		}
	},
}
