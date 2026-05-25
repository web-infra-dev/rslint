// Package no_distracting_elements ports eslint-plugin-jsx-a11y's
// `no-distracting-elements` rule. The rule flags JSX elements whose resolved
// type matches one of a configurable list of "distracting" tag names —
// `<marquee>` and `<blink>` by default. Both elements are deprecated and
// commonly cited for triggering motion sickness / reading difficulty.
//
// Upstream is a one-listener rule: read the (settings-resolved) element type
// for every JsxOpeningElement, look it up in `options.elements ||
// DEFAULT_ELEMENTS` via Array.prototype.find, and report when found. The
// element-type resolution (polymorphic prop, components map) lives in
// `jsxa11yutil.GetElementType`; this rule wires the listener and the find.
package no_distracting_elements

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// errorMessage mirrors upstream's `errorMessage(element)` template literal
// verbatim. The element name is interpolated into the diagnostic so a
// reviewer can grep for the exact tag.
func errorMessage(element string) string {
	return fmt.Sprintf("Do not use <%s> elements as they can create visual accessibility issues and are deprecated.", element)
}

// defaultElements mirrors upstream's `DEFAULT_ELEMENTS` constant. Order is
// preserved because upstream uses Array.prototype.find — when a configured
// list contains duplicates (legal but degenerate), the FIRST occurrence
// wins.
var defaultElements = []string{"marquee", "blink"}

type options struct {
	elements []string
}

func parseOptions(raw any) options {
	opts := options{elements: defaultElements}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	// Upstream: `const elementOptions = options.elements || DEFAULT_ELEMENTS`.
	// `||` only falls through when the LHS is JS-falsy (undefined / null /
	// "" / 0 / NaN). An EXPLICITLY empty array is JS-truthy so it replaces
	// the default and effectively disables the rule for everyone — mirror
	// that. StringSliceOption returns nil for an absent / non-array value
	// (we keep the default) and a non-nil possibly-empty []string for any
	// present array (we use as-is).
	if rawElements, ok := m["elements"]; ok {
		if elements := jsxa11yutil.StringSliceOption(rawElements); elements != nil {
			opts.elements = elements
		}
	}
	return opts
}

var NoDistractingElementsRule = rule.Rule{
	Name: "jsx-a11y/no-distracting-elements",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// Empty element list — rule is effectively disabled for this run.
		// Skip listener registration to avoid `GetElementType` + an empty
		// loop on every JsxOpeningElement in the file. Triggered by:
		//   - explicit `{ elements: [] }` (the user-facing way to disable)
		//   - rslint extension: `[123, true]` (StringSliceOption filters
		//     all non-string entries, leaving an empty slice)
		// Observable behavior matches upstream's `find` over an empty
		// array (always undefined → never reports).
		if len(opts.elements) == 0 {
			return rule.RuleListeners{}
		}

		check := func(node *ast.Node) {
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)
			// Upstream uses Array.prototype.find — first matching element
			// wins. We mirror by short-circuiting on the first hit. Strict
			// equality (`===`) is preserved by Go's `==` on string values.
			for _, e := range opts.elements {
				if elementType == e {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "distractingElement",
						Description: errorMessage(e),
					})
					return
				}
			}
		}

		// tsgo splits ESTree's `JSXOpeningElement` into KindJsxOpeningElement
		// (paired tags `<marquee>...</marquee>`) and KindJsxSelfClosingElement
		// (`<marquee />`). Upstream's single `JSXOpeningElement` listener
		// fires on both forms; we mirror by registering on both kinds.
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
