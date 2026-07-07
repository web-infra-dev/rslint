// Package no_redundant_roles ports eslint-plugin-jsx-a11y's
// `no-redundant-roles` rule. The rule flags a JSX element whose explicit
// `role` attribute duplicates the implicit (browser-provided) ARIA role of
// the underlying HTML element — e.g. `<button role="button" />`,
// `<body role="document" />`.
//
// Some HTML elements carry an implicit ARIA role that browsers already
// expose to assistive technology. Re-declaring the same role explicitly
// adds nothing and is a documented anti-pattern in the W3C ARIA guidance.
//
// Upstream signature:
//
//	options: { [tagName: string]: string[] }
//
// where each key/value pair allow-lists redundant implicit roles for a
// specific HTML element. The DEFAULT is `{ nav: ['navigation'] }` — i.e.
// `<nav role="navigation" />` is permitted out of the box, per the W3C
// advice that some screen readers historically required this echo to
// announce the landmark.
//
// Trigger: a JsxOpeningElement / JsxSelfClosingElement whose
//
//  1. element name resolves (via `getElementType` — `polymorphicPropName` +
//     `components` map) to an HTML element with a non-empty implicit ARIA
//     role,
//  2. carries an explicit literal `role` attribute whose lower-cased value
//     equals that implicit role, AND
//  3. is NOT in the user's allow-list for that element.
//
// The diagnostic message names the element and the implicit role
// (lower-cased), matching upstream's `errorMessage(element, implicitRole)`.
package no_redundant_roles

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Implicit-role lookup (`getImplicitRole`, `getExplicitRole`, and the
// per-element `implicitRoleFor*` helpers) lives in
// [jsxa11yutil] — see [jsxa11yutil.GetImplicitRole] / [jsxa11yutil.GetExplicitRole].
// `role-supports-aria-props` consumes the same helpers, so they are extracted
// to the plugin shared package per the porting skill's
// duplicate-across-rules rule.

// defaultRoleExceptions mirrors upstream's `DEFAULT_ROLE_EXCEPTIONS`. Only
// `nav: ['navigation']` is on by default — the W3C-recommended echo for
// older screen readers.
var defaultRoleExceptions = map[string][]string{
	"nav": {"navigation"},
}

func errorMessage(element, implicitRole string) string {
	return "The element " + element + " has an implicit role of " + implicitRole +
		". Defining this explicitly is redundant and should be avoided."
}

// parseOptions extracts the per-element allow-list map from the rule's
// JSON options. The shape is `{ [tagName]: string[] }` — anything else
// (non-string keys, non-array values, non-string array entries) is
// silently dropped, matching upstream's `additionalProperties` schema
// which permits but does not enforce string-array values.
//
// Returns nil when no options are provided. An EMPTY object (`[{}]`)
// returns a non-nil empty map; both shapes are observably equivalent
// because the downstream lookup (see [NoRedundantRolesRule]) does
// `hasOwn(opts, type)` per element, which is false for every key in
// both cases — so the default exceptions table (e.g. the built-in
// `nav: ['navigation']` allowance) applies. To disable a SPECIFIC
// default, the user must pass `{ <tag>: [] }`, e.g. `{ nav: [] }`.
//
// This mirrors upstream's `options[0] || {}` then per-key
// `hasOwn(allowedRedundantRoles, type)` lookup — `options = undefined`
// and `options = [{}]` both yield an empty allowedRedundantRoles
// object, and `hasOwn({}, 'nav')` is `false`, so DEFAULT_ROLE_EXCEPTIONS
// is consulted in both cases.
func parseOptions(raw any) map[string][]string {
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return nil
	}
	out := map[string][]string{}
	for k, v := range m {
		out[k] = jsxa11yutil.StringSliceOption(v)
	}
	return out
}

var NoRedundantRolesRule = rule.Rule{
	Name: "jsx-a11y/no-redundant-roles",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.UnwrapOptions(_rawOptions)
		opts := parseOptions(rawOptions)

		check := func(node *ast.Node) {
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)
			if elementType == "" {
				return
			}
			attrs := reactutil.GetJsxElementAttributes(node)

			implicitRole, ok := jsxa11yutil.GetImplicitRole(elementType, attrs)
			if !ok {
				return
			}
			explicitRole, ok := jsxa11yutil.GetExplicitRole(attrs)
			if !ok {
				return
			}
			if implicitRole != explicitRole {
				return
			}

			// Allow-list lookup: explicit user config takes priority via
			// `hasOwn(allowedRedundantRoles, type)`. An entry under `type`
			// — even an empty array — fully replaces the default. Only when
			// the entry is ABSENT do we fall through to the built-in
			// `defaultRoleExceptions` (which carries the nav-navigation
			// allowance).
			var allowed []string
			if opts != nil {
				if v, hasOwn := opts[elementType]; hasOwn {
					allowed = v
				} else {
					allowed = defaultRoleExceptions[elementType]
				}
			} else {
				allowed = defaultRoleExceptions[elementType]
			}
			for _, r := range allowed {
				if r == implicitRole {
					return
				}
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noRedundantRoles",
				Description: errorMessage(elementType, strings.ToLower(implicitRole)),
			})
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
