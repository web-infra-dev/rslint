// Package anchor_has_content ports eslint-plugin-jsx-a11y's
// `anchor-has-content` rule. The rule enforces that <a> elements (and
// configured custom anchor components) carry user-perceivable content,
// either via children, a `title` / `aria-label` attribute, or an
// accessible-child fallback (children prop / dangerouslySetInnerHTML).
package anchor_has_content

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const errorMessage = "Anchors must have content and the content must be accessible by a screen reader."

type options struct {
	// components mirrors the upstream `options.components` array — extra tag
	// names that should be treated as anchors in addition to the literal
	// "a". The values are checked AS-IS against the resolved nodeType (after
	// `getElementType`'s polymorphic / componentMap resolution).
	components []string
}

func parseOptions(raw any) options {
	opts := options{}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	opts.components = jsxa11yutil.StringSliceOption(m["components"])
	return opts
}

var AnchorHasContentRule = rule.Rule{
	Name: "jsx-a11y/anchor-has-content",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// Mirrors upstream `typeCheck = ['a'].concat(componentOptions)`.
		// The default anchor tag plus any rule-option-provided custom names.
		typeCheck := append([]string{"a"}, opts.components...)

		elementType := func(node *ast.Node) string {
			return jsxa11yutil.GetElementType(node, ctx.Settings)
		}

		check := func(node *ast.Node) {
			nodeType := elementType(node)
			if !slices.Contains(typeCheck, nodeType) {
				return
			}
			// Upstream calls `hasAccessibleChild(node.parent, elementType)`.
			// JsxAccessibleChildRoot normalizes the paired-vs-self-closing
			// AST split so HasAccessibleChild sees the same surface upstream
			// does, regardless of form.
			if jsxa11yutil.HasAccessibleChild(jsxa11yutil.JsxAccessibleChildRoot(node), elementType) {
				return
			}
			attrs := reactutil.GetJsxElementAttributes(node)
			// Upstream uses `hasAnyProp(attrs, ['title', 'aria-label'])` with
			// `hasProp`'s default `spreadStrict: true` — spread attributes
			// are opaque even when literal. HasAnyJsxPropStrict mirrors this.
			if jsxa11yutil.HasAnyJsxPropStrict(attrs, "title", "aria-label") {
				return
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "anchorHasContent",
				Description: errorMessage,
			})
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
