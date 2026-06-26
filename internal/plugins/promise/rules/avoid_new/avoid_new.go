package avoid_new

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// skipTransparent only unwraps parens — ESLint's ESTree parser drops
// parentheses, so `new (Promise)(...)` already has the identifier visible at
// the ESLint level. TS-only wrappers (non-null, as, satisfies) are
// intentionally NOT unwrapped: the original rule treats them as not-a-Promise-
// constructor and skips silently, and we mirror that.
const skipTransparent = ast.OEKParentheses

var AvoidNewRule = rule.Rule{
	Name: "promise/avoid-new",
	RunWithOptions: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				callee := ast.SkipOuterExpressions(node.AsNewExpression().Expression, skipTransparent)
				if callee == nil || callee.Kind != ast.KindIdentifier {
					return
				}
				if callee.AsIdentifier().Text == "Promise" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "avoidNew",
						Description: "Avoid creating new promises.",
					})
				}
			},
		}
	},
}
