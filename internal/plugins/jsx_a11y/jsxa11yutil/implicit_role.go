package jsxa11yutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
)

// GetImplicitRole mirrors eslint-plugin-jsx-a11y's `util/getImplicitRole.js`:
//
//	if (implicitRoles[type]) {
//	  implicitRole = implicitRoles[type](attrs);
//	}
//	return rolesMap.has(implicitRole) ? implicitRole : null;
//
// Returns the implicit ARIA role for `elementType` (a resolved HTML element
// name as produced by [GetElementType]) given the element's attributes, or
// ("", false) when the element has no implicit role.
//
// The dispatch table is CASE-SENSITIVE — `<BODY />` falls through to no
// implicit role, matching upstream's strict-equality table key. Per-element
// computation lives in [implicitRoleFor*]; this function dispatches by
// element name and then filters out roles that aren't in aria-query's `roles`
// map (which catches the "no implicit role" sentinel — an empty string from
// a/area/link without href, img with empty alt, img with svg src,
// menu/menuitem of unknown type — and would also catch any future divergence
// between implicitRoles output and the roles vocabulary).
//
// Notable upstream nuances locked in:
//   - `a` / `area` / `link` get the `link` implicit role ONLY when an `href`
//     attribute is PRESENT (regardless of value, including empty string,
//     boolean form, and `href={null}`).
//   - `form` and `section` return their roles UNCONDITIONALLY upstream,
//     even though the WAI-ARIA / HTML-ARIA spec scopes those to elements
//     WITH an accessible name. We mirror upstream, not the spec.
//   - `img` returns "" (= no role) when alt is the literal empty string OR
//     when src is a literal string containing ".svg".
//   - `input` defaults to "textbox" for absent / non-literal / unknown type.
//   - `menu` returns "toolbar" only when type is the literal string
//     "toolbar" (case-insensitive).
//   - `menuitem` returns "menuitem" / "menuitemcheckbox" / "menuitemradio"
//     based on type values "command" / "checkbox" / "radio" respectively.
//   - `select` returns "listbox" when `multiple` is truthy OR `size > 1`,
//     "combobox" otherwise.
func GetImplicitRole(elementType string, attrs []*ast.Node) (string, bool) {
	var role string
	switch elementType {
	case "a", "area", "link":
		if FindAttributeByName(attrs, "href") != nil {
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
	if _, ok := AriaRoleAllSet[role]; !ok {
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
// 'img'. Identifiers (`alt={foo}`) extract to null and likewise stay 'img'.
//
// The SVG arm uses optional chaining: only literal-string src values
// participate. Non-literal `src={someVariable}` returns null, and
// `null?.includes(...)` short-circuits to undefined → falsy → stays 'img'.
func implicitRoleForImg(attrs []*ast.Node) string {
	if alt := FindAttributeByName(attrs, "alt"); alt != nil {
		if v, ok := LiteralPropStringValue(alt); ok && v == "" {
			return ""
		}
	}
	if src := FindAttributeByName(attrs, "src"); src != nil {
		if v, ok := LiteralPropStringValue(src); ok && strings.Contains(v, ".svg") {
			return ""
		}
	}
	return "img"
}

// implicitRoleForInput mirrors upstream `implicitRoles/input.js`. The table is
// the HTML-ARIA mapping for input types:
//
//	button/image/reset/submit → button
//	checkbox                  → checkbox
//	radio                     → radio
//	range                     → slider
//	(anything else / absent)  → textbox
//
// Upstream coerces the type value with `getLiteralPropValue(type) || ''` then
// `.toUpperCase()`. The `|| ''` handles `<input type={null}>` (literal "null"),
// `<input type />` (boolean form → true), and `<input type={undefined}>` —
// non-string values short-circuit to the `default → textbox` arm.
func implicitRoleForInput(attrs []*ast.Node) string {
	typeAttr := FindAttributeByName(attrs, "type")
	if typeAttr == nil {
		return "textbox"
	}
	v, ok := LiteralPropStringValue(typeAttr)
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
	typeAttr := FindAttributeByName(attrs, "type")
	if typeAttr == nil {
		return ""
	}
	v, ok := LiteralPropStringValue(typeAttr)
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
	typeAttr := FindAttributeByName(attrs, "type")
	if typeAttr == nil {
		return ""
	}
	v, ok := LiteralPropStringValue(typeAttr)
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
// The `multiple` short-circuit uses [LiteralPropTruthy] — a direct port of
// `!!getLiteralPropValue(prop)` — which correctly classifies `multiple={null}`
// as truthy (LITERAL_TYPES.Literal maps null to the magic string "null"),
// `multiple=""` as falsy, etc.
//
// The `size > 1` check uses [LiteralPropJSNumber], which fully mirrors
// `Number(getLiteralPropValue(size))` including JS-style hex / octal / binary
// string prefixes (`<select size="0x10" />` → 16 → 16 > 1 → listbox) and
// whitespace trimming.
func implicitRoleForSelect(attrs []*ast.Node) string {
	if multiple := FindAttributeByName(attrs, "multiple"); multiple != nil {
		if LiteralPropTruthy(multiple) {
			return "listbox"
		}
	}
	if size := FindAttributeByName(attrs, "size"); size != nil {
		if n, ok := LiteralPropJSNumber(size); ok && n > 1 {
			return "listbox"
		}
	}
	return "combobox"
}

// GetExplicitRole lives in roles.go alongside [AriaRoleAllSet] —
// see [GetExplicitRole] there.
