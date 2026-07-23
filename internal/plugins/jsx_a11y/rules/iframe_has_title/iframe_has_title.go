// Package iframe_has_title ports eslint-plugin-jsx-a11y's `iframe-has-title`
// rule. The rule enforces that every `<iframe>` element carries a unique,
// non-empty `title` attribute so screen-reader users can identify the
// embedded frame's purpose.
package iframe_has_title

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const errorMessage = "<iframe> elements must have a unique title property."

var IframeHasTitleRule = rule.Rule{
	Name:   "jsx-a11y/iframe-has-title",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		elementType := func(node *ast.Node) string {
			return jsxa11yutil.GetElementType(node, ctx.Settings)
		}

		check := func(node *ast.Node) {
			// Upstream: `if (type && type !== 'iframe') return;`. A truthy
			// `type` other than "iframe" short-circuits — but a falsy type
			// (empty string, e.g. unresolvable computed-tag name in tsgo)
			// falls through and is treated as if the element could be
			// `<iframe>`. We mirror the logical-AND semantics verbatim.
			t := elementType(node)
			if t != "" && t != "iframe" {
				return
			}

			// Upstream: `getPropValue(getProp(attrs, 'title'))` followed by
			// `if (title && typeof title === 'string') return;`. The
			// typeof-string gate is the key difference from html-has-lang —
			// boolean / number / object / function values fail even when
			// truthy. PropValueIsTruthyString encodes both halves of the
			// upstream guard:
			//   - `<iframe title="x" />`           → "x"     → truthy string → no report
			//   - `<iframe title={foo} />`         → "foo"   → Identifier name → string → no report
			//   - `<iframe title={undefined} />`   → undef   → falsy → REPORT
			//   - `<iframe title="" />`            → ""      → empty string → REPORT
			//   - `<iframe title={true} />`        → true    → typeof "boolean" → REPORT
			//   - `<iframe title={42} />`          → 42      → typeof "number" → REPORT
			//   - `<iframe title={``} />`          → ""      → empty string → REPORT
			//   - `<iframe title />`               → true    → typeof "boolean" → REPORT
			//   - `<iframe />` / `<iframe {...p}/>`→ nil     → no value → REPORT
			attrs := reactutil.GetJsxElementAttributes(node)
			titleAttr := jsxa11yutil.FindAttributeByName(attrs, "title")
			if jsxa11yutil.PropValueIsTruthyString(titleAttr) {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "iframeHasTitle",
				Description: errorMessage,
			})
		}

		// Upstream listens on `JSXOpeningElement` only — in ESTree this
		// covers both paired (`<iframe></iframe>`) and self-closing
		// (`<iframe />`) forms because both wrap a JSXOpeningElement.
		// tsgo splits them into two distinct kinds, so we register both.
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
