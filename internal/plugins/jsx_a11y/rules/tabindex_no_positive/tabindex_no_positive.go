// Package tabindex_no_positive ports eslint-plugin-jsx-a11y's
// `tabindex-no-positive` rule. The rule discourages a positive `tabIndex`
// prop on any JSX element — positive tab-order values disrupt the natural
// keyboard navigation flow that screen-reader / keyboard-only users rely
// on (WCAG 2.4.3, Focus Order).
//
// Upstream signature: no options — the schema is `generateObjSchema()`
// (an empty object).
//
// Trigger: a JsxAttribute whose name resolves (case-insensitively, after
// `propName(attribute).toUpperCase() === 'TABINDEX'`) to `tabIndex`, and
// whose value, when coerced via `Number(getLiteralPropValue(attribute))`,
// is a real number greater than zero. The diagnostic is reported on the
// JsxAttribute node itself.
//
// Behavior differences from no-noninteractive-tabindex (which also reads
// `tabIndex`):
//
//   - This rule fires on EVERY JSX element, not just DOM elements —
//     `<MyButton tabIndex={5} />` reports here but is skipped by
//     no-noninteractive-tabindex.
//   - The element name, role attribute, and components-map settings are
//     IGNORED — only the tabIndex value matters.
//   - The Number-coercion is via `Number(literalValue)` not `Number.isInteger`,
//     so non-integers like `tabIndex={1.589}` and `tabIndex="0.5"` ALSO
//     report.
//   - The boolean attribute form `<div tabIndex />` reports here (extractValue
//     → true → Number(true) = 1 > 0) but is treated as "undefined" by
//     no-noninteractive-tabindex's getTabIndex helper.
//   - Only the literal extraction path runs — no `getPropValue` fallback —
//     so `tabIndex={cond ? 1 : 2}` does NOT report (LITERAL_TYPES noop on
//     ConditionalExpression).
//
// JsxSpreadAttribute is not visited by this rule's listener — upstream's
// listener is also `JSXAttribute`, so spread shapes like
// `<div {...{tabIndex: 5}} />` are intentionally not analyzed.
package tabindex_no_positive

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// errorMessage mirrors upstream's `errorMessage` string verbatim.
const errorMessage = "Avoid positive integer values for tabIndex."

var TabindexNoPositiveRule = rule.Rule{
	Name: "jsx-a11y/tabindex-no-positive",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		// Source text is needed for raw-string extraction on template
		// literals (NoSubstitutionTemplateLiteral has no RawText field in
		// tsgo). Captured once per file so the per-attribute callback
		// doesn't need to re-read it.
		sourceText := ctx.SourceFile.Text()
		return rule.RuleListeners{
			ast.KindJsxAttribute: func(attr *ast.Node) {
				// Upstream uses `propName(attribute).toUpperCase() === 'TABINDEX'`,
				// so the match is case-INsensitive — `Tabindex`, `TABINDEX`,
				// `tabindex` all qualify, matching jsx-ast-utils' propName
				// behavior on the namespace-stripped attribute name. Locked
				// by the case-variant tests below.
				name := reactutil.GetJsxPropName(attr)
				if !strings.EqualFold(name, "tabIndex") {
					return
				}
				value, ok := jsxa11yutil.LiteralPropToNumber(attr, sourceText)
				if !ok {
					// NaN — upstream `isNaN(value)` short-circuits.
					return
				}
				if value <= 0 {
					// 0 or negative — upstream `value <= 0` short-circuits. Includes
					// negative zero (JS `-0 <= 0` is true) so `tabIndex={-0}` is
					// also skipped here, mirroring upstream's check.
					return
				}
				ctx.ReportNode(attr, rule.RuleMessage{
					Id:          "tabIndexNoPositive",
					Description: errorMessage,
				})
			},
		}
	},
}
