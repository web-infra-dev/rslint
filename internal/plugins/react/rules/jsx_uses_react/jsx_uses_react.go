package jsx_uses_react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxUsesReactRule marks the JSX pragma and fragment bindings as used.
var JsxUsesReactRule = rule.Rule{
	Name:   "react/jsx-uses-react",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		pragma := reactutil.GetReactPragmaFromContext(ctx)
		fragment := reactutil.GetReactFragmentPragma(ctx.Settings)
		mark := func(node *ast.Node) {
			ctx.MarkVariableAsUsed(pragma, node)
		}
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     mark,
			ast.KindJsxSelfClosingElement: mark,
			ast.KindJsxOpeningFragment:    mark,
			ast.KindJsxFragment: func(node *ast.Node) {
				ctx.MarkVariableAsUsed(fragment, node)
			},
		}
	},
}
