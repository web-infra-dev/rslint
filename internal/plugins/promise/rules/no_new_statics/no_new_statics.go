package no_new_statics

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
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
	RunWithOptions: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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

				// Fix: remove "new" (3 chars) plus any trailing whitespace chars
				// (' ', '\t', '\r', '\n'). Comments and other non-whitespace
				// between `new` and the callee are preserved:
				//   new Promise.resolve()    → Promise.resolve()
				//   new  Promise.resolve()   → Promise.resolve()
				//   new/*c*/Promise.resolve() → /*c*/Promise.resolve()
				start := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
				end := start + 3 // len("new") == 3
				src := ctx.SourceFile.Text()
				for end < len(src) && (src[end] == ' ' || src[end] == '\t' || src[end] == '\r' || src[end] == '\n') {
					end++
				}
				ctx.ReportNodeWithFixes(node, buildMessage(methodName), rule.RuleFixRemoveRange(core.NewTextRange(start, end)))
			},
		}
	},
}
