package no_new_statics

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// skipTransparent skips only parentheses — TS wrappers (as, !, satisfies) are
// intentionally left opaque so `new (Promise as any).resolve()` is not flagged,
// matching the behaviour a user sees in ESLint on a non-@typescript-eslint/parser
// run (where type-assertion wrappers are visible as distinct nodes).
const skipTransparent = ast.OEKParentheses

var promiseStatics = map[string]bool{
	"all":           true,
	"allSettled":    true,
	"any":           true,
	"race":          true,
	"reject":        true,
	"resolve":       true,
	"withResolvers": true,
}

func buildMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "avoidNewStatic",
		Description: fmt.Sprintf("Avoid calling 'new' on 'Promise.%s()'", name),
		Data:        map[string]string{"name": name},
	}
}

var NoNewStaticsRule = rule.Rule{
	Name: "promise/no-new-statics",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				callee := ast.SkipOuterExpressions(node.AsNewExpression().Expression, skipTransparent)
				if callee == nil || !ast.IsPropertyAccessExpression(callee) {
					return
				}
				prop := callee.AsPropertyAccessExpression()
				object := ast.SkipOuterExpressions(prop.Expression, skipTransparent)
				if object == nil || !ast.IsIdentifier(object) || object.AsIdentifier().Text != "Promise" {
					return
				}
				name := prop.Name()
				if name == nil || !ast.IsIdentifier(name) {
					return
				}
				methodName := name.AsIdentifier().Text
				if !promiseStatics[methodName] {
					return
				}

				// Fix: remove "new " (4 chars) from the start of the expression.
				nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				removeRange := nodeRange.WithEnd(nodeRange.Pos() + 4)
				ctx.ReportNodeWithFixes(node, buildMessage(methodName), rule.RuleFixRemoveRange(removeRange))
			},
		}
	},
}
