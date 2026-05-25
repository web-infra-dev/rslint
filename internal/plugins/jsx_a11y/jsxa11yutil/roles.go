package jsxa11yutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
)

// AriaRoleNonAbstract is the list of every non-abstract ARIA role as defined
// by `aria-query`'s `rolesMap`. The order mirrors
// `roles.keys().filter(role => roles.get(role).abstract === false)`, so
// callers that need a deterministic iteration order (e.g. test enumeration
// against the upstream `createTests(validRoles)` snapshot) see the same
// order as upstream.
//
// Order observation: ARIA 1.x core roles appear alphabetically first,
// followed by DPUB-ARIA (`doc-*`) extensions, then Graphics-ARIA
// (`graphics-*`) extensions — this is `aria-query`'s key-insertion order,
// not pure alphabetic.
//
// Source: https://github.com/A11yance/aria-query/tree/main/src/etc/roles
var AriaRoleNonAbstract = []string{
	"alert", "alertdialog", "application", "article", "banner", "blockquote",
	"button", "caption", "cell", "checkbox", "code", "columnheader",
	"combobox", "complementary", "contentinfo", "definition", "deletion",
	"dialog", "directory", "document", "emphasis", "feed", "figure", "form",
	"generic", "grid", "gridcell", "group", "heading", "img", "insertion",
	"link", "list", "listbox", "listitem", "log", "main", "mark", "marquee",
	"math", "menu", "menubar", "menuitem", "menuitemcheckbox",
	"menuitemradio", "meter", "navigation", "none", "note", "option",
	"paragraph", "presentation", "progressbar", "radio", "radiogroup",
	"region", "row", "rowgroup", "rowheader", "scrollbar", "search",
	"searchbox", "separator", "slider", "spinbutton", "status", "strong",
	"subscript", "superscript", "switch", "tab", "table", "tablist",
	"tabpanel", "term", "textbox", "time", "timer", "toolbar", "tooltip",
	"tree", "treegrid", "treeitem",
	// DPUB-ARIA digital-publishing roles (https://www.w3.org/TR/dpub-aria-1.0/).
	"doc-abstract", "doc-acknowledgments", "doc-afterword", "doc-appendix",
	"doc-backlink", "doc-biblioentry", "doc-bibliography", "doc-biblioref",
	"doc-chapter", "doc-colophon", "doc-conclusion", "doc-cover",
	"doc-credit", "doc-credits", "doc-dedication", "doc-endnote",
	"doc-endnotes", "doc-epigraph", "doc-epilogue", "doc-errata",
	"doc-example", "doc-footnote", "doc-foreword", "doc-glossary",
	"doc-glossref", "doc-index", "doc-introduction", "doc-noteref",
	"doc-notice", "doc-pagebreak", "doc-pagefooter", "doc-pageheader",
	"doc-pagelist", "doc-part", "doc-preface", "doc-prologue",
	"doc-pullquote", "doc-qna", "doc-subtitle", "doc-tip", "doc-toc",
	// Graphics-ARIA roles (https://www.w3.org/TR/graphics-aria-1.0/).
	"graphics-document", "graphics-object", "graphics-symbol",
}

// AriaRoleAbstract is the list of abstract ARIA roles — values that
// `aria-query`'s `rolesMap` recognizes but that jsx-a11y/aria-role rejects
// (upstream's `validRoles` filter excludes `abstract === true`). The order
// mirrors `roles.keys().filter(role => roles.get(role).abstract === true)`.
//
// Source: https://github.com/A11yance/aria-query/tree/main/src/etc/roles/abstract
var AriaRoleAbstract = []string{
	"command", "composite", "input", "landmark", "range", "roletype",
	"section", "sectionhead", "select", "structure", "widget", "window",
}

// AriaRoleNonAbstractSet provides O(1) "is this a valid (non-abstract) ARIA
// role?" lookup. Used by jsx-a11y/aria-role to validate the literal `role`
// attribute values after splitting on spaces.
var AriaRoleNonAbstractSet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(AriaRoleNonAbstract))
	for _, r := range AriaRoleNonAbstract {
		set[r] = struct{}{}
	}
	return set
}()

// IsValidAriaRole reports whether `role` is a non-abstract ARIA role in
// `aria-query`'s `rolesMap`. Mirrors upstream's
// `validRoles.has(val)` lookup — comparison is exact (case-sensitive) and
// performs no whitespace trimming. `"Button"` is NOT a valid role; `"button"`
// is.
func IsValidAriaRole(role string) bool {
	_, ok := AriaRoleNonAbstractSet[role]
	return ok
}

// AriaRoleAllSet is the union of every name in aria-query's `rolesMap` —
// abstract + non-abstract + DPub + Graphics. Mirrors upstream's
// `rolesMap.has(role)` membership check used by jsx-a11y's `getExplicitRole`
// helper. Unlike [AriaRoleNonAbstractSet], abstract roles (`command`,
// `composite`, `widget`, …) ARE included because upstream's `rolesMap`
// exposes them under the same key space.
var AriaRoleAllSet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(AriaRoleNonAbstract)+len(AriaRoleAbstract))
	for _, r := range AriaRoleNonAbstract {
		set[r] = struct{}{}
	}
	for _, r := range AriaRoleAbstract {
		set[r] = struct{}{}
	}
	return set
}()

// GetExplicitRole mirrors upstream `util/getExplicitRole.js`:
//
//	const explicit = toLowerCase(getLiteralPropValue(getProp(attrs, 'role')));
//	return rolesMap.has(explicit) ? explicit : null;
//
// Steps:
//  1. Look up the `role` attribute via [FindAttributeByName] (mirrors
//     upstream's `getProp` with default `ignoreCase: true` — our shared
//     finder is case-insensitive).
//  2. Extract the literal value via [LiteralPropStringValue] (= upstream
//     `getLiteralPropValue`). Non-literal expressions (Identifier,
//     CallExpression, ConditionalExpression, TemplateExpression with
//     substitutions, …) yield (`""`, false), causing this helper to return
//     ("", false) as well — matches upstream's null return.
//  3. Lowercase the result and check membership in [AriaRoleAllSet]
//     (= upstream's `rolesMap.has`). The membership check filters out
//     non-role strings ("foobar", "button radio") and the magic "null"
//     literal that `getLiteralPropValue` synthesizes for `role={null}`.
//
// Returns the lowercased role and `true` only when the explicit role
// resolves to a real ARIA role name (including abstract roles, DPub
// `doc-*`, and Graphics `graphics-*`). Otherwise returns ("", false).
//
// Notable upstream quirks preserved here:
//
//   - The tag-name parameter that upstream's signature accepts is ignored
//     upstream and not exposed here.
//   - Space-separated multi-role attributes (`role="img button"`) DO NOT
//     match — upstream lowercases then does a whole-string lookup in
//     `rolesMap`. To match individual roles in a multi-role string, see
//     [IsInteractiveRole] / [IsNonInteractiveRole] which split on space.
//   - The "null" magic string from `getLiteralPropValue(role={null})`
//     fails the membership check (lowercased "null" isn't a real role).
func GetExplicitRole(attrs []*ast.Node) (string, bool) {
	roleAttr := FindAttributeByName(attrs, "role")
	if roleAttr == nil {
		return "", false
	}
	value, ok := LiteralPropStringValue(roleAttr)
	if !ok {
		return "", false
	}
	lower := strings.ToLower(value)
	if _, ok := AriaRoleAllSet[lower]; !ok {
		return "", false
	}
	return lower, true
}

// AriaRoleRequiredProps is the list of (role, required ARIA props) pairs for
// every non-abstract ARIA role whose `requiredProps` set is non-empty.
// Mirrors `aria-query`'s `roles.get(role).requiredProps` keys, with the keys
// preserved in insertion order so error messages match upstream's
// `String(Object.keys(requiredProps))` (= comma-joined) output byte-for-byte.
//
// Roles whose `requiredProps` map is empty (e.g. `button`, `link`, `cell`)
// are intentionally absent — the upstream rule's
// `if (requiredProps.length > 0)` gate filters them out, and the rule never
// emits diagnostics for those roles. Keeping them out of this table mirrors
// that gate as a data-structure rather than an extra runtime check.
//
// Source: https://github.com/A11yance/aria-query/tree/main/src/etc/roles
// (per-role `requiredProps` map; abstract roles excluded).
var AriaRoleRequiredProps = []struct {
	Role  string
	Props []string
}{
	{"checkbox", []string{"aria-checked"}},
	{"combobox", []string{"aria-controls", "aria-expanded"}},
	{"heading", []string{"aria-level"}},
	{"menuitemcheckbox", []string{"aria-checked"}},
	{"menuitemradio", []string{"aria-checked"}},
	{"meter", []string{"aria-valuenow"}},
	{"option", []string{"aria-selected"}},
	{"radio", []string{"aria-checked"}},
	{"scrollbar", []string{"aria-controls", "aria-valuenow"}},
	{"slider", []string{"aria-valuenow"}},
	{"switch", []string{"aria-checked"}},
	{"treeitem", []string{"aria-selected"}},
}

// ariaRoleRequiredPropsMap is the indexed view of AriaRoleRequiredProps —
// O(1) lookup of a role's required props slice. Returns (nil, false) for
// roles with no required props (absent from the table); callers should treat
// that as upstream's `requiredProps.length > 0` gate failing, i.e. skip the
// role.
var ariaRoleRequiredPropsMap = func() map[string][]string {
	m := make(map[string][]string, len(AriaRoleRequiredProps))
	for _, e := range AriaRoleRequiredProps {
		m[e.Role] = e.Props
	}
	return m
}()

// AriaRoleRequiredPropsFor returns the required ARIA props for `role`.
// Returns (nil, false) when the role has no required props or is unknown.
// Used by jsx-a11y/role-has-required-aria-props to gate on upstream's
// `if (requiredProps.length > 0)` condition.
func AriaRoleRequiredPropsFor(role string) ([]string, bool) {
	v, ok := ariaRoleRequiredPropsMap[role]
	return v, ok
}

// SemanticRoleConcept describes one (element-name, attributes) → roles
// mapping derived from `axobject-query`'s `elementAXObjects` joined with
// `AXObjectRoles`. The result is the set of ARIA roles that the given HTML
// element naturally implies — what jsx-a11y's `isSemanticRoleElement`
// upstream consults to skip the required-props check when the element
// already provides the role's semantics by virtue of its tag (e.g.
// `<input type="checkbox">` already provides the `checkbox` role).
//
// The table is filtered to entries where at least one mapped role appears
// in [AriaRoleRequiredProps]. Entries whose roles all have empty
// `requiredProps` are dropped — the rule's `validRoles.forEach(role => ...)`
// loop never fires on those roles anyway, so the semantic skip would have
// no observable effect.
//
// Source: https://github.com/A11yance/axobject-query/tree/main/src
type SemanticRoleConcept struct {
	// Name is the HTML element name (`input`, `select`, `h1`, ...).
	Name string
	// Attributes are the (name, value) pairs that ALL must match on the JSX
	// element for the concept to apply. An entry with Value == "" means the
	// attribute presence alone is enough; entries derived from
	// `axobject-query` always carry a value, so the empty-Value path is
	// dead code today but preserved to mirror upstream's
	// `cAttr.value !== undefined ? ... : true` branch.
	Attributes []SemanticRoleAttribute
	// Roles is the set of ARIA roles this concept implies.
	Roles []string
}

// SemanticRoleAttribute is one (name, value) pair in a SemanticRoleConcept's
// attribute requirement list. Comparisons against the JSX element's
// attributes are case-sensitive — mirrors upstream `cAttr.name === propName(attr)`
// and `cAttr.value === getLiteralPropValue(attr)`.
type SemanticRoleAttribute struct {
	Name  string
	Value string
}

// SemanticRoleConcepts is the table consulted by
// `jsx-a11y/role-has-required-aria-props` to skip elements that natively
// imply the role on their `role` attribute. See [SemanticRoleConcept] for
// the filter / derivation criterion.
//
// Generated from `axobject-query` v4 by joining `elementAXObjects` (concept
// → axObjects) with `AXObjectRoles` (axObject → roles), then keeping every
// (concept, joined-roles) entry where at least one role in the joined set
// has non-empty `requiredProps`.
//
// Order matches the live `axobject-query` iteration order so a future
// audit can diff against `axobject-query`'s `elementAXObjects` map.
var SemanticRoleConcepts = []SemanticRoleConcept{
	{
		Name:       "input",
		Attributes: []SemanticRoleAttribute{{Name: "type", Value: "checkbox"}},
		Roles:      []string{"checkbox", "switch"},
	},
	{
		Name:       "select",
		Attributes: nil,
		Roles:      []string{"combobox", "listbox"},
	},
	{Name: "h1", Roles: []string{"heading"}},
	{Name: "h2", Roles: []string{"heading"}},
	{Name: "h3", Roles: []string{"heading"}},
	{Name: "h4", Roles: []string{"heading"}},
	{Name: "h5", Roles: []string{"heading"}},
	{Name: "h6", Roles: []string{"heading"}},
	{Name: "option", Roles: []string{"option"}},
	{
		Name:       "input",
		Attributes: []SemanticRoleAttribute{{Name: "type", Value: "radio"}},
		Roles:      []string{"radio"},
	},
	{
		Name:       "input",
		Attributes: []SemanticRoleAttribute{{Name: "type", Value: "range"}},
		Roles:      []string{"slider"},
	},
}
