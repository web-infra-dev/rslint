// Package html_has_lang ports eslint-plugin-jsx-a11y's `html-has-lang` rule.
// The rule enforces that every `<html>` element carries a truthy `lang` prop
// so screen readers announce content in the correct language.
package html_has_lang

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const errorMessage = "<html> elements must have the lang prop."

var HtmlHasLangRule = rule.Rule{
	Name: "jsx-a11y/html-has-lang",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		elementType := func(node *ast.Node) string {
			return jsxa11yutil.GetElementType(node, ctx.Settings)
		}

		check := func(node *ast.Node) {
			// Upstream: `if (type && type !== 'html') return;`. A truthy
			// `type` other than "html" short-circuits — but a falsy type
			// (empty string, e.g. computed-member tag name unresolvable
			// in tsgo) falls through and is treated as if the element
			// could be `<html>`. We mirror the logical-AND semantics
			// verbatim so the listener gate matches upstream byte-for-
			// byte.
			t := elementType(node)
			if t != "" && t != "html" {
				return
			}

			// Upstream: `getPropValue(getProp(attrs, 'lang'))` followed by
			// `if (lang) return;`. PropValueIsTruthy mirrors `!!getPropValue`
			// via the full staticEval (extract) path, so:
			//   - `<html lang="en" />`         → "en"   → truthy → no report
			//   - `<html lang />`              → true   → truthy → no report
			//   - `<html lang={foo} />`        → "foo"  → truthy → no report
			//   - `<html lang={undefined} />`  → undef  → falsy  → REPORT
			//   - `<html />` / `<html {...p}/>`→ nil    → falsy  → REPORT
			attrs := reactutil.GetJsxElementAttributes(node)
			langAttr := jsxa11yutil.FindAttributeByName(attrs, "lang")
			if jsxa11yutil.PropValueIsTruthy(langAttr) {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "htmlHasLang",
				Description: errorMessage,
			})
		}

		// Upstream listens on `JSXOpeningElement` only — in ESTree this
		// covers both paired (`<html></html>`) and self-closing
		// (`<html />`) forms because both wrap a JSXOpeningElement node.
		// tsgo splits them into two distinct kinds, so we register both.
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
