package valid_expect_with_promise

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/valid_expect_with_promise"
)

var ValidExpectWithPromiseRule = rule.Rule{
	Name:             "jest/valid-expect-with-promise",
	RequiresTypeInfo: true,
	Schema:           shared.Schema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkThenables := shared.ParseOptions(options).CheckThenables
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil || parsed.Head.Local.Node == nil {
				return
			}
			head := ast.WalkUpParenthesizedExpressions(parsed.Head.Local.Node.Parent)
			if head == nil || head.Kind != ast.KindCallExpression {
				return
			}

			// Without an argument, the subject type is the checker's error type,
			// which is neither a Promise nor a thenable.
			promise := false
			if arguments := head.AsCallExpression().Arguments.Nodes; len(arguments) > 0 {
				subject := arguments[0]
				typ := ctx.TypeChecker.GetTypeAtLocation(subject)
				// Unlike utils.IsPromiseLike, constructors of Promise subclasses
				// also count, since their declared type extends Promise.
				promise = utils.IsBuiltinSymbolLike(ctx.Program(), ctx.TypeChecker, typ, "Promise") ||
					(checkThenables && shared.IsStrictThenable(ctx.TypeChecker, subject, typ))
			}

			var modifier *jestUtils.ParsedJestFnMemberEntry
			for i := range parsed.ModifierEntries {
				if parsed.ModifierEntries[i].Name != "not" {
					modifier = &parsed.ModifierEntries[i]
					break
				}
			}

			if promise && modifier == nil {
				ctx.ReportNode(node, shared.PoorlyExpectedPromiseMessage)
				return
			}
			if !promise && modifier != nil {
				ctx.ReportNode(modifier.Node, shared.UnneededRejectResolveMessage(modifier.Name))
			}
		}}
	},
}
