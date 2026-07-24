// Package no_interactive_element_to_noninteractive_role ports
// eslint-plugin-jsx-a11y's `no-interactive-element-to-noninteractive-role`
// rule. The rule flags inherently interactive HTML elements (e.g. `<a
// href>`, `<button>`, `<input>`, …) that are assigned a non-interactive or
// presentation ARIA role — collapsing the platform-supplied a11y semantics
// of the underlying control to a content container is a strong signal of
// misuse since screen-reader / keyboard users no longer recognize the
// affordance.
//
// Upstream signature:
//
//	options: {
//	  [tagName: string]: string[]     // per-element allowed-role override
//	}
//
// The single options object is an arbitrary tag-name → allowed-roles map.
// The recommended preset ships `{ tr: ['none', 'presentation'], canvas:
// ['img'] }` (so `<tr role="presentation" />` and `<canvas role="img" />`
// are exempt under recommended but flagged under strict).
//
// Trigger sequence — for every JsxAttribute named `role` on a JSX opening /
// self-closing element. Bail-outs return without reporting:
//
//  1. Attribute name (post-namespace serialization) is not literally `role`
//     → bail. Mirrors upstream `propName(attribute) !== 'role'`, which —
//     because `propName` stringifies `JSXNamespacedName` to `ns:name` —
//     skips namespaced forms like `<div mynamespace:role="term" />`.
//  2. Tag isn't in aria-query's `dom` map → bail. Custom JSX components
//     (`<Button />`) and unknown tags are exempted because we can't know
//     what low-level DOM they render to.
//  3. `options[type]` exists AND contains the FIRST role attribute's
//     LITERAL value → bail. Mirrors upstream's
//     `includes(allowedRoles[type], role)` — array.includes uses
//     SameValueZero so a `null` (non-literal) role never matches a
//     `string[]` allow-list, which collapses the check to a no-op for
//     non-literal `role` expressions.
//  4. Element is inherently interactive AND (role is non-interactive OR
//     role is presentation) → REPORT on the listener-current JsxAttribute
//     node.
//
// Diagnostic text mirrors upstream verbatim:
//
//	"Interactive elements should not be assigned non-interactive roles."
//
// Reports happen at the JsxAttribute, not the opening element — upstream
// uses `node: attribute` in `context.report`. For duplicate `role`
// attributes on the same element (invalid React, parseable JSX), each
// listener invocation evaluates against the FIRST role attribute (via
// `getProp(attrs, 'role')`), so both attributes get reported with the same
// underlying classification — mirror.
package no_interactive_element_to_noninteractive_role

import (
	_ "embed"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_interactive_element_to_noninteractive_role.schema.json
var schemaJSON []byte

// errorMessage mirrors upstream's `errorMessage` string verbatim.
const errorMessage = "Interactive elements should not be assigned non-interactive roles."

// options carries the per-element allowed-role overrides. Each key is a tag
// name and each value is a list of role strings that are exempt from the
// rule for that tag. Keys whose value is not an array of strings are
// silently ignored (JSON schema would reject them upstream, so the filter
// is defensive).
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

var NoInteractiveElementToNoninteractiveRoleRule = rule.Rule{
	Name:   "jsx-a11y/no-interactive-element-to-noninteractive-role",
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

				// Step 2: per-element override. Upstream uses
				// `getLiteralPropValue(getProp(attrs, 'role'))`, which
				// resolves to the FIRST role attribute on the element
				// (NOT necessarily the listener-current one — matters
				// only for duplicate-role forms). Routing through
				// `FindAttributeByName` + `LiteralPropStringValue`
				// mirrors that exactly, including literal-spread walks
				// (`{...{ role: "img" }}`).
				//
				// `includes(allowedRoles[type], role)` uses
				// SameValueZero, so a non-literal role (LiteralPropStringValue
				// returns ok=false) can never match a `string[]` allow-list
				// — the check collapses to no-op in that case, which is what
				// we want.
				if allowed, ok := opts.allowedRoles[elementType]; ok {
					firstRoleAttr := jsxa11yutil.FindAttributeByName(attrs, "role")
					if role, hasLiteral := jsxa11yutil.LiteralPropStringValue(firstRoleAttr); hasLiteral {
						if slices.Contains(allowed, role) {
							return
						}
					}
				}

				// Step 3: interactive element + non-interactive / presentation
				// role → report. All three predicates are tag-aware and
				// share the FILTERED `attrs` view; the role classifiers
				// internally extract from the first role attribute via
				// `getLiteralPropValue(getProp(attrs, 'role'))`.
				if jsxa11yutil.IsInteractiveElement(elementType, attrs) &&
					(jsxa11yutil.IsNonInteractiveRole(elementType, attrs) ||
						jsxa11yutil.IsPresentationRole(attrs)) {
					ctx.ReportNode(attr, rule.RuleMessage{
						Id:          "noInteractiveElementToNoninteractiveRole",
						Description: errorMessage,
					})
				}
			},
		}
	},
}
