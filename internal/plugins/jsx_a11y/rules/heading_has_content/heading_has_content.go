// Package heading_has_content ports eslint-plugin-jsx-a11y's
// `heading-has-content` rule. The rule enforces that heading elements
// (h1–h6, plus any user-configured custom heading components) carry
// content readable by a screen reader, either via children or via the
// accessible-child fallback (children prop / dangerouslySetInnerHTML).
package heading_has_content

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const errorMessage = "Headings must have content and the content must be accessible by a screen reader."

// headings mirrors upstream's hard-coded list of native HTML heading tags.
// Order matches upstream verbatim — slices.Contains short-circuits, so the
// concrete order only affects micro-perf, not semantics.
var headings = []string{"h1", "h2", "h3", "h4", "h5", "h6"}

type options struct {
	// components mirrors upstream's `options.components` array — extra tag
	// names that should be treated as headings in addition to h1–h6. Values
	// are checked AS-IS against the resolved nodeType (after `getElementType`'s
	// polymorphic / componentMap resolution).
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

var HeadingHasContentRule = rule.Rule{
	Name: "jsx-a11y/heading-has-content",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// Mirrors upstream `typeCheck = headings.concat(componentOptions)`.
		typeCheck := append(append([]string{}, headings...), opts.components...)

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
			// Upstream's third gate: `isHiddenFromScreenReader(nodeType, node.attributes)`.
			// Note this is checked against the OPENING element itself — i.e. an
			// `<h1 aria-hidden />` is treated as hidden and skipped (valid),
			// regardless of its lack of content. IsHiddenFromScreenReader
			// reads attributes off the passed node, so we pass the opening /
			// self-closing element directly (not jsxRoot).
			if jsxa11yutil.IsHiddenFromScreenReader(node, elementType) {
				return
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "headingHasContent",
				Description: errorMessage,
			})
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
