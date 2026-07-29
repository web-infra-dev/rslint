// Package no_noninteractive_element_to_interactive_role ports
// eslint-plugin-jsx-a11y's `no-noninteractive-element-to-interactive-role`
// rule. It flags inherently non-interactive HTML elements (e.g. `<address>`,
// `<article>`, `<table>`, headings, lists, …) that are assigned an
// interactive ARIA role (`button`, `link`, `menuitem`, `checkbox`, …). The
// platform-supplied a11y semantics of structural / content elements do not
// support keyboard focus / activation; promoting them to widgets via the
// `role` attribute alone is a strong signal that the affordance is missing
// (no keyboard handler, no focus management, no announced state).
//
// Upstream signature:
//
//	options: {
//	  [tagName: string]: string[]     // per-element allowed-role override
//	}
//
// Each entry whitelists specific (tag, role) combinations under the
// `recommended` preset (`{ul: [...], ol: [...], li: [...], table: ["grid"],
// td: ["gridcell"], fieldset: ["radiogroup", "presentation"]}`). The
// `strict` preset omits the allow-list, so every (non-interactive element,
// interactive role) pair triggers a report.
//
// Trigger sequence — for every JsxAttribute named `role` on a JSX opening
// or self-closing element. Bail-outs return without reporting:
//
//  1. Attribute name (post-namespace serialization) is not literally `role`
//     → bail. Mirrors upstream `propName(attribute) !== 'role'`; namespaced
//     forms (`mynamespace:role`) serialize to `"mynamespace:role"` and miss
//     the equality check.
//  2. Tag isn't in aria-query's `dom` map → bail. Custom JSX components
//     (`<Button />`) and unknown tags are exempted because we can't know
//     what low-level DOM they render to.
//  3. `options[type]` exists AND contains the role returned by
//     [jsxa11yutil.GetExplicitRole] (lowercased + rolesMap-validated literal
//     value) → bail. The shape differs from
//     [no_interactive_element_to_noninteractive_role]: that sibling rule
//     uses `getLiteralPropValue` directly (raw, no rolesMap filter), while
//     this rule uses upstream's `getExplicitRole` helper. As a result a
//     `role` attribute that isn't a real ARIA role (e.g. `role="foobar"` or
//     `role="img button"` — a multi-role string) never matches the allow-
//     list and falls through to the interactive-role check, which itself
//     filters out non-roles.
//  4. Element is inherently non-interactive AND role is interactive (first
//     valid space-separated role wins via [jsxa11yutil.IsInteractiveRole])
//     → REPORT on the listener-current JsxAttribute node.
//
// Diagnostic text mirrors upstream verbatim:
//
//	"Non-interactive elements should not be assigned interactive roles."
//
// Reports happen at the JsxAttribute, not the opening element — upstream
// uses `node: attribute` in `context.report`. For duplicate `role`
// attributes on the same element (invalid React, parseable JSX), each
// listener invocation evaluates against the FIRST role attribute (via
// `getProp(attrs, 'role')`), so both attributes get reported if the FIRST
// classifies as interactive.
package no_noninteractive_element_to_interactive_role

import (
	_ "embed"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_noninteractive_element_to_interactive_role.schema.json
var schemaJSON []byte

// errorMessage mirrors upstream's `errorMessage` string verbatim.
const errorMessage = "Non-interactive elements should not be assigned interactive roles."

// options carries the per-element allowed-role overrides. Each key is a tag
// name and each value is a list of role strings that are exempt from the
// rule for that tag. Non-array values are silently dropped (upstream's JSON
// schema would reject them, so the filter is defensive); non-string entries
// within an array are likewise filtered out by [jsxa11yutil.StringSliceOption].
type options struct {
	allowedRoles map[string][]string
}

func parseOptions(raw []any) options {
	opts := options{}
	if len(raw) == 0 {
		return opts
	}
	m, _ := raw[0].(map[string]interface{})
	for key, v := range m {
		if parsed := jsxa11yutil.StringSliceOption(v); parsed != nil {
			if opts.allowedRoles == nil {
				opts.allowedRoles = make(map[string][]string)
			}
			opts.allowedRoles[key] = parsed
		}
	}
	return opts
}

var NoNoninteractiveElementToInteractiveRoleRule = rule.Rule{
	Name:   "jsx-a11y/no-noninteractive-element-to-interactive-role",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindJsxAttribute: func(attr *ast.Node) {
				// Upstream `propName(attribute) !== 'role'`. Case-sensitive
				// match — upstream's `propName` returns the JSXIdentifier
				// text verbatim, and `<X ROLE="…" />` therefore does NOT
				// trigger this rule (the `role` attribute on a tag with
				// uppercase ROLE is parsed as the literal name `ROLE`).
				// For JSXNamespacedName (`mynamespace:role`), `propName`
				// returns the colon-joined form (`"mynamespace:role"`),
				// which also fails the equality check. Mirror via the
				// same colon-joined serialization in
				// [reactutil.GetJsxPropName].
				if reactutil.GetJsxPropName(attr) != "role" {
					return
				}

				parent := reactutil.GetJsxParentElement(attr)
				if parent == nil {
					return
				}
				attrs := reactutil.GetJsxElementAttributes(parent)
				elementType := jsxa11yutil.GetElementType(parent, ctx.Settings)

				// Step 1: custom components / unknown tags — exempt.
				// Upstream `!dom.has(type)`.
				if !jsxa11yutil.IsDOMElement(elementType) {
					return
				}

				// Step 2: per-element override. Upstream computes role via
				// `getExplicitRole(type, attributes)` (lowercased + filtered
				// through `rolesMap.has`), then runs
				// `includes(allowedRoles[type], role)`. Mirror with
				// [jsxa11yutil.GetExplicitRole]; a non-literal or non-role
				// `role` attribute returns ok=false here and falls through
				// to the interactive-role check (which itself rejects
				// non-roles), matching upstream's `null` propagation
				// through `includes([…], null) → false`.
				if allowed, hasOwn := opts.allowedRoles[elementType]; hasOwn {
					if role, hasRole := jsxa11yutil.GetExplicitRole(attrs); hasRole {
						if slices.Contains(allowed, role) {
							return
						}
					}
				}

				// Step 3: non-interactive element + interactive role →
				// report. Both predicates share the SAME `attrs` view; the
				// role classifier internally extracts from the first role
				// attribute via `getLiteralPropValue(getProp(attrs, 'role'))`,
				// so duplicate-role attributes both classify against the
				// FIRST occurrence's value.
				if jsxa11yutil.IsNonInteractiveElement(elementType, attrs) &&
					jsxa11yutil.IsInteractiveRole(elementType, attrs) {
					ctx.ReportNode(attr, rule.RuleMessage{
						Id:          "noNoninteractiveElementToInteractiveRole",
						Description: errorMessage,
					})
				}
			},
		}
	},
}
