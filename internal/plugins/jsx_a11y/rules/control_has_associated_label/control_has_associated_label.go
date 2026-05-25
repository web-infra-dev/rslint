// Package control_has_associated_label ports eslint-plugin-jsx-a11y's
// `control-has-associated-label` rule. Enforces that an interactive control
// (DOM element, control component, or DOM element with an interactive role)
// has a discernible text label — either through visible text content, a
// labelling attribute (`alt` / `aria-label` / `aria-labelledby` / user-
// configured `labelAttributes`), or descendant content within a configured
// recursion depth.
//
// Upstream listener gate (`JSXElement`), checked in order:
//
//  1. `newIgnoreElements` (= user `ignoreElements` ∪ `['link']`) contains the
//     resolved tag → skip. The `link` exemption is hard-coded upstream and
//     cannot be disabled via config.
//  2. `getLiteralPropValue` of the `role` attribute is in `ignoreRoles` → skip.
//  3. Element is hidden from screen readers (`<input type="hidden">`,
//     `aria-hidden={true}`, `aria-hidden="true"`) → skip.
//  4. Trigger condition: `isInteractiveElement || (isDOMElement &&
//     isInteractiveRole) || controlComponents.indexOf(tag) > -1`.
//     - When false: no label requirement (rule does not apply).
//     - When true: run `mayHaveAccessibleLabel(root, min(options.depth ?? 2,
//     25), labelAttributes, getElementType, controlComponents)`.
//  5. If no label is found, report on the opening element.
//
// The message id is `controlHasAssociatedLabel`; the message text is taken
// verbatim from upstream's `errorMessage`.
package control_has_associated_label

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// errorMessage mirrors upstream's `errorMessage` constant verbatim.
const errorMessage = "A control must be associated with a text label."

// defaultDepth mirrors upstream's `options.depth === undefined ? 2 : options.depth`.
const defaultDepth = 2

// maxDepthCap mirrors upstream's `Math.min(..., 25)` ceiling on the recursion
// budget — protects against pathological JSX trees regardless of user config.
const maxDepthCap = 25

type options struct {
	labelAttributes   []string
	controlComponents []string
	ignoreElements    []string
	ignoreRoles       []string
	depth             int
}

func parseOptions(raw any) options {
	opts := options{depth: defaultDepth}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	opts.labelAttributes = jsxa11yutil.StringSliceOption(m["labelAttributes"])
	opts.controlComponents = jsxa11yutil.StringSliceOption(m["controlComponents"])
	opts.ignoreElements = jsxa11yutil.StringSliceOption(m["ignoreElements"])
	opts.ignoreRoles = jsxa11yutil.StringSliceOption(m["ignoreRoles"])
	// depth: number, default 2. JSON decodes numbers as float64.
	// Upstream truthiness: `options.depth === undefined ? 2 : options.depth`,
	// then `Math.min(depth, 25)`. We treat the absence of the key (`v == nil`)
	// as "undefined" and apply the default.
	if v, ok := m["depth"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			opts.depth = int(f)
		}
	}
	if opts.depth > maxDepthCap {
		opts.depth = maxDepthCap
	}
	return opts
}

var ControlHasAssociatedLabelRule = rule.Rule{
	Name: "jsx-a11y/control-has-associated-label",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		// Upstream `newIgnoreElements = new Set([].concat(ignoreElements,
		// ignoreList))` where `ignoreList = ['link']`. The `link` entry is
		// applied unconditionally — there is no way to opt out via config.
		newIgnoreElements := append([]string{}, opts.ignoreElements...)
		newIgnoreElements = append(newIgnoreElements, "link")

		getElementType := func(node *ast.Node) string {
			return jsxa11yutil.GetElementType(node, ctx.Settings)
		}

		check := func(elementNode *ast.Node) {
			opening, openingAttrs := jsxa11yutil.OpeningElementOf(elementNode)
			if opening == nil {
				return
			}
			tag := getElementType(opening)

			// Step 1: newIgnoreElements (user ignoreElements ∪ ['link']).
			// Exact match — matches upstream's `Set.has(tag)`.
			if slices.Contains(newIgnoreElements, tag) {
				return
			}

			// Step 2: `getLiteralPropValue(getProp(attributes, 'role'))` in
			// `ignoreRoles`. Routes through `LiteralPropStringValue` (=
			// upstream's `getLiteralPropValue` for string-typed results) so
			// non-literal role expressions (Identifier, Call, Conditional)
			// fall through to ok=false and never satisfy `ignoreRoles`.
			roleAttr := jsxa11yutil.FindAttributeByName(openingAttrs, "role")
			if roleAttr != nil {
				if roleValue, ok := jsxa11yutil.LiteralPropStringValue(roleAttr); ok {
					if slices.Contains(opts.ignoreRoles, roleValue) {
						return
					}
				}
			}

			// Step 3: hidden from screen reader — same helper used across
			// the plugin (e.g. click-events-have-key-events). Treats
			// `<input type="hidden">` and `aria-hidden={true}` /
			// `aria-hidden="true"` as hidden.
			if jsxa11yutil.IsHiddenFromScreenReader(opening, getElementType) {
				return
			}

			// Step 4: trigger condition.
			// Upstream: `isInteractiveElement || (isDOMElement &&
			// isInteractiveRole) || controlComponents.indexOf(tag) > -1`.
			// `controlComponents.indexOf` is EXACT match (case-sensitive)
			// and does NOT use minimatch — the minimatch comparison only
			// appears inside `mayHaveAccessibleLabel`'s React-component
			// fallback. Preserve this asymmetry.
			nodeIsInteractiveElement := jsxa11yutil.IsInteractiveElement(tag, openingAttrs)
			nodeIsDOMElement := jsxa11yutil.IsDOMElement(tag)
			nodeIsInteractiveRole := jsxa11yutil.IsInteractiveRole(tag, openingAttrs)
			nodeIsControlComponent := slices.Contains(opts.controlComponents, tag)
			shouldCheck := nodeIsInteractiveElement ||
				(nodeIsDOMElement && nodeIsInteractiveRole) ||
				nodeIsControlComponent
			if !shouldCheck {
				return
			}

			if jsxa11yutil.MayHaveAccessibleLabel(elementNode, 0, opts.depth, opts.labelAttributes, opts.controlComponents, getElementType) {
				return
			}

			ctx.ReportNode(opening, rule.RuleMessage{
				Id:          "controlHasAssociatedLabel",
				Description: errorMessage,
			})
		}

		// Listen on both paired (JsxElement) and self-closing
		// (JsxSelfClosingElement) — tsgo splits these into separate kinds
		// while ESTree (and therefore upstream) sees them as JSXElement with
		// `selfClosing` flag. Both forms must be classified independently.
		return rule.RuleListeners{
			ast.KindJsxElement:            check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
