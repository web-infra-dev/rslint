// Package no_aria_hidden_on_focusable ports eslint-plugin-jsx-a11y's
// `no-aria-hidden-on-focusable` rule. It enforces that `aria-hidden="true"`
// is not set on a focusable element — pairing the two confuses screen reader
// users when the element can still be reached by keyboard.
//
// Upstream signature:
//
//	options: {}   (no options)
//
// Trigger sequence on each JsxOpeningElement / JsxSelfClosingElement:
//
//  1. `aria-hidden` must resolve to JS boolean `true` via upstream's
//     `getPropValue(...) === true` (handled by [jsxa11yutil.IsAriaHiddenTrue]).
//     The boolean-attribute form (`<div aria-hidden />`) counts, as does the
//     case-insensitive string `"true"` and the explicit `{true}`.
//  2. The element must be focusable per upstream's `isFocusable(type, attrs)`:
//     - If `isInteractiveElement(type, attrs)` is true → focusable iff
//     `tabIndex === undefined || tabIndex >= 0`. Inherently interactive
//     elements (button, input, textarea, select, anchor-with-href, ...)
//     skip the rule only when an explicit `tabIndex={-2}`-or-smaller (or
//     any non-numeric resolved value that isn't `null`) removes them from
//     focus order.
//     - Otherwise → focusable iff `tabIndex >= 0`. Non-interactive tags
//     become focusable only when an explicit `tabIndex` of `0` or greater
//     is added.
//
// `aria-hidden=false` short-circuits step 1 and never reports, regardless of
// focusability. Nested elements are NOT walked — the rule operates per JSX
// element, mirroring upstream which intentionally leaves descendant focusable
// elements under an `aria-hidden` ancestor to other rules.
//
// Diagnostic text mirrors upstream verbatim: `aria-hidden="true" must not be
// set on focusable elements.`
package no_aria_hidden_on_focusable

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// errorMessage mirrors upstream's `message` field verbatim.
const errorMessage = `aria-hidden="true" must not be set on focusable elements.`

var NoAriaHiddenOnFocusableRule = rule.Rule{
	Name: "jsx-a11y/no-aria-hidden-on-focusable",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		sourceText := ctx.SourceFile.Text()

		check := func(node *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(node)
			if !jsxa11yutil.IsAriaHiddenTrue(attrs) {
				return
			}
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)
			if !isFocusable(elementType, attrs, sourceText) {
				return
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noAriaHiddenOnFocusable",
				Description: errorMessage,
			})
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

// isFocusable mirrors upstream's `isFocusable(type, attributes)`:
//
//	const tabIndex = getTabIndex(getProp(attrs, 'tabIndex'));
//	if (isInteractiveElement(type, attrs)) {
//	  return (tabIndex === undefined || tabIndex >= 0);
//	}
//	return tabIndex >= 0;
//
// upstream's `getTabIndex` returns three observable kinds:
//
//   - a Number (step-1 integer or step-2 Number-coercible) — both arms reduce
//     to `val >= 0`.
//   - JS `undefined` (step-1 short-circuits: boolean form, empty / NaN string,
//     boolean literal, non-integer numeric; or step-2 returns undefined via
//     explicit `{undefined}` / `typeof` / `void`). For interactive elements
//     this matches the `=== undefined` arm; for non-interactive `>= 0` is
//     `NaN >= 0` → false.
//   - JS `null` (step-2 fallthrough for shapes where `TYPES` has no extractor
//     — `JSXEmptyExpression`, `SatisfiesExpression`, etc.). `null === undefined`
//     is false but `null >= 0` ToNumber-coerces to `0 >= 0 → true`, so BOTH
//     arms accept null as focusable.
//   - A non-numeric resolved string (step-2 `Identifier` non-`undefined`,
//     `MemberExpression`, `CallExpression`, `TaggedTemplateExpression`, ...).
//     `string !== undefined` AND `Number(string) → NaN`, so BOTH arms reject.
//
// [jsxa11yutil.GetTabIndexEx]'s ambiguous `(0, false, false)` state collapses
// the last two non-null cases; [jsxa11yutil.HasUpstreamTabIndexValue]
// distinguishes them by asking `tabIndex !== undefined`. Combining both
// recovers the four-way classification.
func isFocusable(elementType string, attrs []*ast.Node, sourceText string) bool {
	tabIndexAttr := jsxa11yutil.FindAttributeByName(attrs, "tabIndex")
	val, resolved, nullLike := jsxa11yutil.GetTabIndexEx(tabIndexAttr, sourceText)
	interactive := jsxa11yutil.IsInteractiveElement(elementType, attrs)

	if resolved {
		// Numeric result — both arms reduce to `val >= 0`.
		return val >= 0
	}
	if nullLike {
		// null — `null >= 0` ToNumber-coerces to 0; both arms accept.
		return true
	}
	if !interactive {
		// Non-interactive arm: `>= 0`. Whether the value is undefined or a
		// non-numeric resolved string, ToNumber produces NaN → false.
		return false
	}
	// Interactive arm: `=== undefined || >= 0`. The `>= 0` half is already
	// excluded by resolved=false, nullLike=false (any value that would yield
	// true would have been resolved or null). So focusability turns purely
	// on `tabIndex === undefined`.
	return !jsxa11yutil.HasUpstreamTabIndexValue(attrs, sourceText)
}
