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
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		check := func(node *ast.Node) {
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)
			if elementType == "" {
				return
			}
			attrs := reactutil.GetJsxElementAttributes(node)

			implicitRole, ok := getImplicitRole(elementType, attrs)
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

// getImplicitRole mirrors upstream `util/getImplicitRole.js`:
//
//	if (implicitRoles[type]) {
//	  implicitRole = implicitRoles[type](attrs);
//	}
//	return rolesMap.has(implicitRole) ? implicitRole : null;
//
// The map lookup is CASE-SENSITIVE — `<BODY />` falls through to no
// implicit role, matching upstream's strict-equality table key.
//
// Per-element computation lives in [implicitRoleFor*]; this function
// dispatches by element name and then filters out roles that aren't in
// aria-query's `roles` map (which catches the "no implicit role"
// sentinel — an empty string from a, area, link, img-with-empty-alt,
// img-with-svg-src, menu/menuitem of unknown type — and would also
// catch any future divergence between implicitRoles output and the
// roles vocabulary).
func getImplicitRole(elementType string, attrs []*ast.Node) (string, bool) {
	var role string
	switch elementType {
	case "a", "area", "link":
		// Upstream: only links with `href` get the `link` implicit role.
		// `getProp(attrs, 'href')` is truthy when the attribute is present
		// regardless of value — including empty-string, boolean form, and
		// `href={null}`. We mirror via FindAttributeByName.
		if jsxa11yutil.FindAttributeByName(attrs, "href") != nil {
			role = "link"
		}
	case "article":
		role = "article"
	case "aside":
		role = "complementary"
	case "body":
		role = "document"
	case "button":
		role = "button"
	case "datalist":
		role = "listbox"
	case "details":
		role = "group"
	case "dialog":
		role = "dialog"
	case "form":
		// UNCONDITIONAL. Upstream `implicitRoles/form.js` returns 'form'
		// without checking for an accessible name. The WAI-ARIA / HTML-ARIA
		// spec actually scopes the implicit `form` role to elements WITH an
		// accessible name (`aria-label`, `aria-labelledby`, `title`), but
		// eslint-plugin-jsx-a11y deliberately does not implement that
		// nuance — and our port mirrors the plugin's behavior, not the spec.
		role = "form"
	case "h1", "h2", "h3", "h4", "h5", "h6":
		role = "heading"
	case "hr":
		role = "separator"
	case "img":
		role = implicitRoleForImg(attrs)
	case "input":
		role = implicitRoleForInput(attrs)
	case "li":
		role = "listitem"
	case "menu":
		role = implicitRoleForMenu(attrs)
	case "menuitem":
		role = implicitRoleForMenuitem(attrs)
	case "meter":
		role = "progressbar"
	case "nav":
		role = "navigation"
	case "ol", "ul":
		role = "list"
	case "option":
		role = "option"
	case "output":
		role = "status"
	case "progress":
		role = "progressbar"
	case "section":
		// UNCONDITIONAL. Same as `form` above — upstream
		// `implicitRoles/section.js` returns 'region' without the
		// accessible-name precondition the HTML-ARIA spec defines. We
		// mirror upstream, not the spec.
		role = "region"
	case "select":
		role = implicitRoleForSelect(attrs)
	case "tbody", "tfoot", "thead":
		role = "rowgroup"
	case "textarea":
		role = "textbox"
	default:
		return "", false
	}
	if _, ok := jsxa11yutil.AriaRoleAllSet[role]; !ok {
		return "", false
	}
	return role, true
}

// implicitRoleForImg mirrors upstream `implicitRoles/img.js`:
//
//	const alt = getProp(attrs, 'alt');
//	if (alt && getLiteralPropValue(alt) === '') return '';
//	const src = getProp(attrs, 'src');
//	if (src && getLiteralPropValue(src)?.includes('.svg')) return '';
//	return 'img';
//
// The empty-alt arm fires only when alt is a LITERAL empty string — the
// boolean form `<img alt />` extracts as `true`, not `''`, so it stays
// 'img'. Identifiers (`alt={foo}`) extract to null and likewise stay
// 'img'.
//
// The SVG arm uses optional chaining: only literal-string src values
// participate. Non-literal `src={someVariable}` returns null, and
// `null?.includes(...)` short-circuits to undefined → falsy → stays
// 'img' (locked by upstream test `<img src={someVariable} role="img" />`).
func implicitRoleForImg(attrs []*ast.Node) string {
	if alt := jsxa11yutil.FindAttributeByName(attrs, "alt"); alt != nil {
		if v, ok := jsxa11yutil.LiteralPropStringValue(alt); ok && v == "" {
			return ""
		}
	}
	if src := jsxa11yutil.FindAttributeByName(attrs, "src"); src != nil {
		if v, ok := jsxa11yutil.LiteralPropStringValue(src); ok && strings.Contains(v, ".svg") {
			return ""
		}
	}
	return "img"
}

// implicitRoleForInput mirrors upstream `implicitRoles/input.js`. The
// table is the HTML-ARIA mapping for input types:
//
//	button/image/reset/submit → button
//	checkbox                  → checkbox
//	radio                     → radio
//	range                     → slider
//	(anything else / absent)  → textbox
//
// Upstream coerces the type value with `getLiteralPropValue(type) || ''`
// then `.toUpperCase()`. The trailing `|| ''` handles
// `<input type={null}>` (literalPropValue → "null"), `<input type />`
// (boolean form → true), and `<input type={undefined}>` — the upstream
// switch's `typeof value === 'string' && value.toUpperCase()` then short-
// circuits on non-string values, hitting the `default → textbox` arm.
func implicitRoleForInput(attrs []*ast.Node) string {
	typeAttr := jsxa11yutil.FindAttributeByName(attrs, "type")
	if typeAttr == nil {
		return "textbox"
	}
	v, ok := jsxa11yutil.LiteralPropStringValue(typeAttr)
	if !ok {
		return "textbox"
	}
	switch strings.ToUpper(v) {
	case "BUTTON", "IMAGE", "RESET", "SUBMIT":
		return "button"
	case "CHECKBOX":
		return "checkbox"
	case "RADIO":
		return "radio"
	case "RANGE":
		return "slider"
	default:
		return "textbox"
	}
}

// implicitRoleForMenu mirrors upstream `implicitRoles/menu.js`:
//
//	if `type === 'toolbar'` (case-insensitive) → toolbar
//	otherwise (absent type, non-literal type, other values) → ''
func implicitRoleForMenu(attrs []*ast.Node) string {
	typeAttr := jsxa11yutil.FindAttributeByName(attrs, "type")
	if typeAttr == nil {
		return ""
	}
	v, ok := jsxa11yutil.LiteralPropStringValue(typeAttr)
	if !ok {
		return ""
	}
	if strings.EqualFold(v, "toolbar") {
		return "toolbar"
	}
	return ""
}

// implicitRoleForMenuitem mirrors upstream `implicitRoles/menuitem.js`:
//
//	command  → menuitem
//	checkbox → menuitemcheckbox
//	radio    → menuitemradio
//	(other / absent / non-literal) → ''
func implicitRoleForMenuitem(attrs []*ast.Node) string {
	typeAttr := jsxa11yutil.FindAttributeByName(attrs, "type")
	if typeAttr == nil {
		return ""
	}
	v, ok := jsxa11yutil.LiteralPropStringValue(typeAttr)
	if !ok {
		return ""
	}
	switch strings.ToUpper(v) {
	case "COMMAND":
		return "menuitem"
	case "CHECKBOX":
		return "menuitemcheckbox"
	case "RADIO":
		return "menuitemradio"
	}
	return ""
}

// implicitRoleForSelect mirrors upstream `implicitRoles/select.js`:
//
//	const multiple = getProp(attrs, 'multiple');
//	if (multiple && getLiteralPropValue(multiple)) return 'listbox';
//	const size = getProp(attrs, 'size');
//	const sizeValue = size && getLiteralPropValue(size);
//	return sizeValue > 1 ? 'listbox' : 'combobox';
//
// The `multiple` short-circuit uses [jsxa11yutil.LiteralPropTruthy] —
// a direct port of `!!getLiteralPropValue(prop)` — which correctly
// classifies `multiple={null}` as truthy (LITERAL_TYPES.Literal maps
// null to the magic string "null"), `multiple=""` as falsy, etc.
//
// The `size > 1` check uses [jsxa11yutil.LiteralPropJSNumber], which
// fully mirrors `Number(getLiteralPropValue(size))` including JS-style
// hex / octal / binary string prefixes (`<select size="0x10" />` → 16
// → 16 > 1 → listbox) and whitespace trimming. Hand-rolled
// `strconv.ParseFloat` would miss these.
//
// Notable rows in the resulting truth table:
//
//	<select size />            → upstream true → ToNumber=1   → 1>1 false → combobox
//	<select size="" />         → ""  → ToNumber=0             → 0>1 false → combobox
//	<select size="0x10" />     → 16  → 16>1 true              →           listbox
//	<select size={null} />     → "null" → NaN                 → false     → combobox
//	<select size={1} />        → 1   → 1>1 false              → combobox
//	<select size={2} />        → 2   → 2>1 true               → listbox
//	<select size="3" />        → "3" → 3>1 true               → listbox
//	<select size={x} />        → null literal → not coercible → combobox
func implicitRoleForSelect(attrs []*ast.Node) string {
	if multiple := jsxa11yutil.FindAttributeByName(attrs, "multiple"); multiple != nil {
		if jsxa11yutil.LiteralPropTruthy(multiple) {
			return "listbox"
		}
	}
	if size := jsxa11yutil.FindAttributeByName(attrs, "size"); size != nil {
		if n, ok := jsxa11yutil.LiteralPropJSNumber(size); ok && n > 1 {
			return "listbox"
		}
	}
	return "combobox"
}
