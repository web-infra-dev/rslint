// Package interactive_supports_focus ports eslint-plugin-jsx-a11y's
// `interactive-supports-focus` rule. The rule enforces that any DOM element
// carrying a mouse / keyboard event handler AND an interactive ARIA `role`
// also be keyboard-reachable: either inherently focusable (interactive
// element / inherently non-interactive element / non-interactive role
// already covers it) or via an explicit `tabIndex`.
//
// Upstream signature:
//
//	options: { tabbable?: string[] }   (default: [])
//
// `tabbable` enumerates roles that MUST be sequentially tabbable (tabIndex
// must be `0`, not `-1`). When the offending element's role is in `tabbable`
// the diagnostic carries a single suggestion (`tabIndex={0}`); otherwise the
// diagnostic carries two suggestions (`tabIndex={0}` or `tabIndex={-1}`).
//
// Trigger sequence — each predicate is checked in order against the JSX
// opening element. Bail-outs return without reporting:
//
//  1. Type isn't an aria-query DOM element → bail (custom components).
//  2. No mouse / keyboard interactive event handler attached → bail.
//  3. Element looks disabled (HTML5 `disabled` attribute is set to anything
//     other than `undefined`, or `aria-disabled` resolves to literal true)
//     → bail.
//  4. Element is hidden from screen readers (`aria-hidden={true}` or `<input
//     type="hidden">`) → bail.
//  5. Element role resolves to `presentation` / `none` → bail.
//  6. Element role is non-interactive (first valid space-separated role is
//     in the non-interactive set), OR element is inherently non-interactive,
//     OR element is inherently interactive (already focusable), OR element
//     has any tabIndex value (upstream `!== undefined`) → bail.
//  7. Element role resolves to an interactive role → REPORT. Message and
//     suggestion shape depends on whether the role is in `tabbable`.
//
// Diagnostic text mirrors upstream:
//
//	"Elements with the '<role>' interactive role must be tabbable."   (tabbable list)
//	"Elements with the '<role>' interactive role must be focusable."  (otherwise)
//
// Suggestions insert ` tabIndex={0}` or ` tabIndex={-1}` after the JSX
// element's name node (upstream `fixer.insertTextAfter(node.name, ...)`).
package interactive_supports_focus

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	// suggestionInsertTabIndexZeroDesc mirrors upstream's `tabIndex=0`
	// message verbatim.
	suggestionInsertTabIndexZeroDesc = "Add `tabIndex={0}` to make the element focusable in sequential keyboard navigation."
	// suggestionInsertTabIndexNegOneDesc mirrors upstream's `tabIndex=-1`
	// message verbatim.
	suggestionInsertTabIndexNegOneDesc = "Add `tabIndex={-1}` to make the element focusable but not reachable via sequential keyboard navigation."
)

// options captures the parsed configuration. Nil distinguishes "absent" from
// "explicit []" — both produce identical observable behavior (empty
// inclusion check), so we collapse them into a single nil-able slice.
type options struct {
	Tabbable []string
}

func parseOptions(raw any) options {
	opts := options{}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	opts.Tabbable = jsxa11yutil.StringSliceOption(m["tabbable"])
	return opts
}

var InteractiveSupportsFocusRule = rule.Rule{
	Name: "jsx-a11y/interactive-supports-focus",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		sourceText := ctx.SourceFile.Text()

		check := func(node *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(node)
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)

			if !jsxa11yutil.IsDOMElement(elementType) {
				return
			}

			hasInteractiveProps := jsxa11yutil.HasAnyJsxPropStrict(attrs, jsxa11yutil.InteractiveEventHandlerNames...)
			if !hasInteractiveProps {
				return
			}

			if jsxa11yutil.IsDisabledElement(attrs) {
				return
			}

			getElementType := func(child *ast.Node) string {
				return jsxa11yutil.GetElementType(child, ctx.Settings)
			}
			if jsxa11yutil.IsHiddenFromScreenReader(node, getElementType) {
				return
			}

			if jsxa11yutil.IsPresentationRole(attrs) {
				return
			}

			if !jsxa11yutil.IsInteractiveRole(elementType, attrs) {
				return
			}
			if jsxa11yutil.IsInteractiveElement(elementType, attrs) {
				return
			}
			if jsxa11yutil.IsNonInteractiveElement(elementType, attrs) {
				return
			}
			if jsxa11yutil.IsNonInteractiveRole(elementType, attrs) {
				return
			}
			if jsxa11yutil.HasUpstreamTabIndexValue(attrs, sourceText) {
				return
			}

			// Extract the role for the diagnostic message. Upstream's
			// `getLiteralPropValue(getProp(attributes, 'role'))` returns the
			// raw string for the LITERAL paths and synthesizes the magic
			// `"null"` / template-placeholder strings for the few other
			// shapes LITERAL_TYPES handles. For the shapes that survive the
			// preceding `isInteractiveRole` filter — i.e. role is a static
			// string whose first valid space-separated role is interactive —
			// LiteralPropStringValue suffices.
			role, _ := jsxa11yutil.LiteralPropStringValue(jsxa11yutil.FindAttributeByName(attrs, "role"))

			tagName := reactutil.GetJsxTagName(node)
			if tagName == nil {
				// Defensive: every JSX opening / self-closing element has a
				// tag-name node in legal source. Skip rather than crash.
				return
			}

			tabIndexZeroSuggestion := rule.RuleSuggestion{
				Message: rule.RuleMessage{
					Id:          "tabIndexZero",
					Description: suggestionInsertTabIndexZeroDesc,
				},
				FixesArr: []rule.RuleFix{rule.RuleFixInsertAfter(tagName, " tabIndex={0}")},
			}

			if slices.Contains(opts.Tabbable, role) {
				ctx.ReportNodeWithSuggestions(node, rule.RuleMessage{
					Id:          "tabbable",
					Description: fmt.Sprintf("Elements with the '%s' interactive role must be tabbable.", role),
				}, tabIndexZeroSuggestion)
				return
			}

			tabIndexNegOneSuggestion := rule.RuleSuggestion{
				Message: rule.RuleMessage{
					Id:          "tabIndexNegOne",
					Description: suggestionInsertTabIndexNegOneDesc,
				},
				FixesArr: []rule.RuleFix{rule.RuleFixInsertAfter(tagName, " tabIndex={-1}")},
			}

			ctx.ReportNodeWithSuggestions(node, rule.RuleMessage{
				Id:          "focusable",
				Description: fmt.Sprintf("Elements with the '%s' interactive role must be focusable.", role),
			}, tabIndexZeroSuggestion, tabIndexNegOneSuggestion)
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
