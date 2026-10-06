package jsx_uses_vars

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxUsesVarsRule marks JSX component bindings as used by no-unused-vars.
var JsxUsesVarsRule = rule.Rule{
	Name:   "react/jsx-uses-vars",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		mark := func(node *ast.Node) {
			tag := node.TagName()
			if tag == nil {
				return
			}
			if ast.IsIdentifier(tag) {
				name := tag.Text()
				// Upstream tests only ASCII lowercase, not TypeScript's broader
				// intrinsic-tag classification (which also includes hyphens).
				if name == "" || (name[0] >= 'a' && name[0] <= 'z') {
					return
				}
			} else {
				for tag.Kind == ast.KindPropertyAccessExpression {
					tag = tag.Expression()
				}
				if !ast.IsIdentifier(tag) {
					return
				}
			}
			ctx.MarkVariableAsUsed(tag.Text(), node)
		}
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     mark,
			ast.KindJsxSelfClosingElement: mark,
		}
	},
}
