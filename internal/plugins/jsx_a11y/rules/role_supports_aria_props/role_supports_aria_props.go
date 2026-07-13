// Package role_supports_aria_props ports eslint-plugin-jsx-a11y's
// `role-supports-aria-props` rule. The rule listens on every JSX opening
// element, resolves the element's effective ARIA role (explicit `role`
// attribute, falling back to the implicit role per the HTML-ARIA mapping),
// and reports any ARIA prop on the element that is not in the role's
// `props` set per aria-query's rolesMap.
//
// Trigger order — mirrors upstream's `JSXOpeningElement` listener
// step-for-step:
//
//  1. Resolve the effective element type via [jsxa11yutil.GetElementType]
//     (honors the `polymorphicPropName` and `components` jsx-a11y settings).
//  2. Lookup the explicit `role` attribute via case-insensitive
//     [jsxa11yutil.FindAttributeByName].
//  3. Compute roleValue:
//     - If role attribute is present: extract via
//     [jsxa11yutil.LiteralPropStringValue] (= `getLiteralPropValue`).
//     Returns "" / no-string for non-literal expressions which fall
//     through to the typeof-string skip below.
//     - Otherwise: [jsxa11yutil.GetImplicitRole] returns the implicit role
//     per the HTML-ARIA table (or "" when the element has no implicit
//     role).
//  4. isImplicit = (no explicit role attribute AND implicit role found).
//     Used only to switch error message phrasing.
//  5. Skip if roleValue is not a string OR is not a key in
//     [jsxa11yutil.AriaRolePropsMap] (= `roles.get(roleValue) === undefined`).
//     The membership check is CASE-SENSITIVE on the raw extracted value —
//     `<div role="BUTTON" aria-checked />` skips silently because aria-query's
//     keys are lowercase.
//  6. Compute the role's supported-props set from
//     [jsxa11yutil.AriaRolePropsMap]; the invalid set is every ARIA name in
//     [jsxa11yutil.AriaPropertyNames] that is NOT in the supported set.
//  7. For each non-spread JsxAttribute on the element:
//     - Skip when getPropValue is nullish (= null or undefined).
//     - Compute the prop's name (raw, no case normalization — matches
//     upstream's `propName(prop)`).
//     - Report when the name is in the invalid set. The diagnostic is
//     attached to the JSXOpeningElement, not the attribute, mirroring
//     upstream's `context.report({ node, ... })`.
//
// Phase 1 Step 6 — observable divergences from upstream:
//   - None. The rule is a thin port. Two nuances worth flagging in case a
//     future maintainer is surprised:
//   - The membership check on roleValue is case-sensitive against
//     lowercase aria-query keys (so `role="BUTTON"` skips silently). This
//     matches upstream — `roles` is a Map<string, def> keyed by lowercase.
//   - Mixed-case ARIA prop names (e.g. `aria-Checked`) fail the
//     case-sensitive `invalidAriaPropsForRole.has(name)` lookup and are
//     silently NOT validated. This matches upstream's
//     `propName(prop)` (case-preserving) feeding into a lowercase Set.
package role_supports_aria_props

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// errorMessage mirrors upstream's `errorMessage(attr, role, tag, isImplicit)`
// template byte-for-byte.
func errorMessage(attr, role, tag string, isImplicit bool) string {
	if isImplicit {
		return "The attribute " + attr + " is not supported by the role " + role +
			". This role is implicit on the element " + tag + "."
	}
	return "The attribute " + attr + " is not supported by the role " + role + "."
}

var RoleSupportsAriaPropsRule = rule.Rule{
	Name: "jsx-a11y/role-supports-aria-props",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)
			if elementType == "" {
				return
			}
			attrs := reactutil.GetJsxElementAttributes(node)

			// Step 2: explicit role lookup.
			roleAttr := jsxa11yutil.FindAttributeByName(attrs, "role")

			// Step 3: compute roleValue. Note we DO NOT lowercase the
			// extracted value — upstream's `roles.get(roleValue)` uses the
			// raw literal value verbatim (case-sensitive), so e.g.
			// `role="BUTTON"` falls through to the membership check which
			// fails (since aria-query keys are lowercase) and the rule skips.
			var roleValue string
			var hasRoleValue bool
			if roleAttr != nil {
				roleValue, hasRoleValue = jsxa11yutil.LiteralPropStringValue(roleAttr)
			} else {
				roleValue, hasRoleValue = jsxa11yutil.GetImplicitRole(elementType, attrs)
			}

			// Step 4: isImplicit determines error-message phrasing.
			// Mirrors upstream `roleValue && role === undefined`.
			isImplicit := roleAttr == nil && hasRoleValue && roleValue != ""

			// Step 5: membership check. Skip when roleValue isn't a string
			// (covered by !hasRoleValue) OR is not a recognized ARIA role
			// (= `roles.get(roleValue) === undefined`).
			if !hasRoleValue {
				return
			}
			supportedProps, ok := jsxa11yutil.AriaRolePropsMap[roleValue]
			if !ok {
				return
			}

			// Step 7: walk every non-spread JsxAttribute. Spread attributes
			// are opaque per upstream's `prop.type !== 'JSXSpreadAttribute'`
			// filter — even literal-object spreads are NOT walked.
			for _, attr := range attrs {
				if attr.Kind != ast.KindJsxAttribute {
					continue
				}
				// Skip nullish values: `<div alt={null} />`,
				// `<div alt={undefined} />`. Upstream uses
				// `getPropValue(prop) != null` (loose equality, catches
				// both null and undefined).
				if jsxa11yutil.PropValueIsNullish(attr) {
					continue
				}
				name := reactutil.GetJsxPropName(attr)
				// Mirrors upstream's `invalidAriaPropsForRole.has(name)` —
				// the invalid set is every ARIA prop name not in the
				// role's supported set. Implemented as two lookups
				// (`AriaPropertySet` membership + supported-set absence)
				// to avoid materializing the inverted set per element.
				if _, isAria := jsxa11yutil.AriaPropertySet[name]; !isAria {
					continue
				}
				if _, isSupported := supportedProps[name]; isSupported {
					continue
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "role-supports-aria-props",
					Description: errorMessage(name, roleValue, elementType, isImplicit),
				})
			}
		}

		// Upstream listens on `JSXOpeningElement` only — ESTree wraps both
		// paired and self-closing forms under that node. tsgo splits them,
		// so we register both kinds.
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
