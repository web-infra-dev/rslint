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
			// Upstream short-circuits on IsDOMComponent ONLY for a bare
			// JSXIdentifier tag (`<div>`, `<x-gif>`). Member-expression tags
			// with a lowercase base (`<appp.Foo>`) must still be reported on
			// the base — even though jsxUtil.isDOMComponent would call the
			// whole tag "DOM" by the leftmost-first-char rule.
			if tagName.Kind == ast.KindIdentifier && reactutil.IsDOMComponent(element) {
				return
			}
			identNode := reactutil.GetJsxTagBaseIdentifier(element)
			if identNode == nil {
				return
			}
			name := identNode.AsIdentifier().Text
			// `<this />` — ThisKeyword is not an Identifier so
			// GetJsxTagBaseIdentifier already returns nil for it. This guards
			// the pathological tsgo edge where a JSX parser recovery produces
			// an Identifier whose text happens to be `this`.
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
