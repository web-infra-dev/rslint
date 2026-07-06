package jsx_props_no_spread_multi

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxPropsNoSpreadMultiRule disallows spreading the same identifier multiple
// times in one JSX opening element.
var JsxPropsNoSpreadMultiRule = rule.Rule{
	Name: "react/jsx-props-no-spread-multi",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		check := func(node *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(node)
			if len(attrs) == 0 {
				return
			}

			seen := map[string]struct{}{}
			for _, attr := range attrs {
				if !ast.IsJsxSpreadAttribute(attr) {
					continue
				}
				expr := attr.AsJsxSpreadAttribute().Expression
				// ESTree flattens parentheses, so `{...(props)}` still reaches
				// upstream as an Identifier. TS-specific wrappers (`as`, `!`,
				// `satisfies`) stay opaque because upstream sees distinct nodes.
				if expr != nil {
					expr = ast.SkipParentheses(expr)
				}
				if !ast.IsIdentifier(expr) {
					continue
				}
				name := expr.AsIdentifier().Text
				if _, ok := seen[name]; ok {
					ctx.ReportNode(attr, rule.RuleMessage{
						Id:          "noMultiSpreading",
						Description: "Spreading the same expression multiple times is forbidden",
					})
				}
				seen[name] = struct{}{}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
