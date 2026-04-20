// cspell:ignore appp

package jsx_no_undef

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var JsxNoUndefRule = rule.Rule{
	Name: "react/jsx-no-undef",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		check := func(element *ast.Node) {
			tagName := reactutil.GetJsxTagName(element)
			if tagName == nil {
				return
			}
			// Upstream short-circuits intrinsic tags ONLY in the
			// `case 'JSXIdentifier'` branch — an Identifier whose first
			// character is lowercase (`/^[a-z]/` per jsxUtil.isDOMComponent).
			// Member-expression tags with a lowercase base (`<appp.Foo>`)
			// must still be checked against the base identifier, so this
			// guard deliberately does not apply to them.
			if tagName.Kind == ast.KindIdentifier {
				text := tagName.AsIdentifier().Text
				if text != "" && text[0] >= 'a' && text[0] <= 'z' {
					return
				}
			}
			identNode := reactutil.GetJsxTagBaseIdentifier(tagName)
			if identNode == nil {
				// ThisKeyword base, JsxNamespacedName, or any tsgo shape we
				// don't classify — upstream treats all of these as "skip".
				return
			}
			name := identNode.AsIdentifier().Text
			// Defensive: upstream returns early on `node.name === 'this'`.
			// tsgo normally emits ThisKeyword for `this` (already handled
			// via GetJsxTagBaseIdentifier returning nil), but a parser
			// recovery edge could still produce Identifier("this").
			if name == "this" {
				return
			}
			if utils.IsShadowed(identNode, name) {
				return
			}
			ctx.ReportNode(identNode, rule.RuleMessage{
				Id:          "undefined",
				Description: "'" + name + "' is not defined.",
			})
		}

		return rule.RuleListeners{
			// tsgo emits a JsxOpeningElement inside a JsxElement for the paired
			// `<Foo></Foo>` form, and a top-level JsxSelfClosingElement for
			// `<Foo />`. ESTree wraps both in a single JSXOpeningElement, so
			// we listen to both kinds. JsxFragment (`<></>`) has no tag name
			// and is never visited by these listeners — matching upstream.
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
