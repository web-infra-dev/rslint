// Package no_noninteractive_tabindex ports eslint-plugin-jsx-a11y's
// `no-noninteractive-tabindex` rule. The rule discourages a `tabIndex` prop
// (with value ≥ 0) on elements that aren't keyboard-interactive — putting a
// non-interactive element in the tab order generally signals a misuse of the
// platform a11y semantics, and screen-reader / keyboard users get no
// productive interaction once focus arrives.
//
// Upstream signature:
//
//	options: {
//	  tags?:                  string[]   (default: undefined)
//	  roles?:                 string[]   (default: undefined)
//	  allowExpressionValues?: boolean    (default: undefined / false-ish)
//	}
//
// Trigger: a JsxOpeningElement / JsxSelfClosingElement whose `tabIndex` prop
// resolves to a usable integer ≥ 0, the element type is in aria-query's
// `dom` map, and the element is neither inherently interactive (per
// elementRoles / elementAXObjects) nor carrying an interactive `role`
// attribute. The diagnostic is reported on the `tabIndex` prop node.
//
// `tags` / `roles` are user escape hatches: a literal-string match against
// the resolved element name or `role` value short-circuits the report.
//
// `allowExpressionValues`: when true, `role={someExpression}` (anything that
// isn't a Literal-typed initializer or `undefined` JSX expression) skips the
// report. The `recommended` config sets this to true with
// `roles: ['tabpanel']` so the typical `<div role="tabpanel" tabIndex="0" />`
// pattern is allowed.
package no_noninteractive_tabindex

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// errorMessage mirrors upstream's `errorMessage` string verbatim.
const errorMessage = "`tabIndex` should only be declared on interactive elements."

// options holds the parsed configuration. nil slices distinguish "absent"
// from "explicit []": the absent case keeps the upstream `tags &&` truthy
// guard (no skip), while an explicit empty array also evaluates falsy under
// `tags &&`. Both shapes happen to produce the same observable behavior here
// — kept as one nil-able slice for simplicity.
type options struct {
	Tags                  []string
	Roles                 []string
	AllowExpressionValues bool
}

func parseOptions(raw any) options {
	opts := options{}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	opts.Tags = jsxa11yutil.StringSliceOption(m["tags"])
	opts.Roles = jsxa11yutil.StringSliceOption(m["roles"])
	if v, ok := m["allowExpressionValues"].(bool); ok {
		opts.AllowExpressionValues = v
	}
	return opts
}

var NoNoninteractiveTabindexRule = rule.Rule{
	Name: "jsx-a11y/no-noninteractive-tabindex",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.LegacyUnwrapOptions(_rawOptions)
		opts := parseOptions(rawOptions)
		// sourceText is required by GetTabIndexEx for raw-text template
		// literal extraction (NoSubstitutionTemplate has no RawText field).
		sourceText := ctx.SourceFile.Text()

		check := func(node *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(node)
			tabIndexProp := jsxa11yutil.FindAttributeByName(attrs, "tabIndex")
			if tabIndexProp == nil {
				return
			}
			// upstream `if (typeof tabIndex === 'undefined') return;` matches
			// strictly when getTabIndex's step-1 short-circuited (boolean
			// form, empty string, NaN, missing). The step-2 fallback returns
			// `null` (typeof object) for unrecognized expression types, which
			// passes the `typeof === 'undefined'` guard and reaches the
			// `tabIndex >= 0` check below — where `null >= 0` ToNumber-coerces
			// to `0 >= 0` = true → REPORT. GetTabIndexEx surfaces the
			// distinction via nullLike so we can mirror this exactly.
			tabIndex, hasTabIndex, nullLike := jsxa11yutil.GetTabIndexEx(tabIndexProp, sourceText)
			if !hasTabIndex && !nullLike {
				return
			}

			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)
			if !jsxa11yutil.IsDOMElement(elementType) {
				// Custom components — upstream skips so we don't second-guess
				// what low-level DOM the component renders.
				return
			}

			// User escape hatches.
			if opts.Tags != nil && slices.Contains(opts.Tags, elementType) {
				return
			}
			// Upstream uses `roles && includes(roles, role)` where `role` is
			// `getLiteralPropValue(getProp(attributes, 'role'))` — undefined
			// when the prop is absent or non-literal, in which case
			// `includes(roles, undefined)` is always false.
			if opts.Roles != nil {
				if roleAttr := jsxa11yutil.FindAttributeByName(attrs, "role"); roleAttr != nil {
					if roleVal, hasLiteral := jsxa11yutil.LiteralStringValue(roleAttr); hasLiteral {
						if slices.Contains(opts.Roles, roleVal) {
							return
						}
					}
				}
			}

			// allowExpressionValues + non-literal `role` → unconditional skip.
			// Note: the upstream code includes a "Special case if role is
			// assigned using ternary with literals on both side" branch, but
			// the if-block returns regardless of which arm matches. The
			// ternary check is therefore dead code — we mirror the OBSERVABLE
			// behavior (always skip) and elide the dead branch.
			if opts.AllowExpressionValues && jsxa11yutil.IsNonLiteralProperty(attrs, "role") {
				return
			}

			// Inherently / explicitly interactive — both forms exempt.
			if jsxa11yutil.IsInteractiveElement(elementType, attrs) ||
				jsxa11yutil.IsInteractiveRole(elementType, attrs) {
				return
			}

			// Upstream `tabIndex >= 0` ToNumber-coerces null to 0; the
			// nullLike arm therefore unconditionally reports (`0 >= 0` is
			// true). The hasTabIndex arm compares the resolved number.
			if (hasTabIndex && tabIndex >= 0) || nullLike {
				ctx.ReportNode(tabIndexProp, rule.RuleMessage{
					Id:          "noNoninteractiveTabindex",
					Description: errorMessage,
				})
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
