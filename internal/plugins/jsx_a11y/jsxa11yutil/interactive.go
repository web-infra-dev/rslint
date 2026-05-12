package jsxa11yutil

// This file ports the data tables and predicates that
// eslint-plugin-jsx-a11y derives at module-load time from `aria-query` and
// `axobject-query`:
//
//   - InteractiveRoles / NonInteractiveRoles  ← roles whose superClass
//     contains 'widget' (excluding `progressbar`, plus `toolbar`) vs the
//     remainder.
//   - interactiveElementRoleSchemas / nonInteractiveElementRoleSchemas /
//     interactiveElementAXSchemas — `(elementName, attributePredicates)`
//     entries from aria-query's `elementRoles` / axobject-query's
//     `elementAXObjects`.
//
// The data tables snapshot the upstream output of the same filter pipeline
// — see the script in `agents/port-rule/notes` for the regenerator.
//
// Behavior parity caveat — schema-level `constraints` (e.g. "scoped to the
// body element", "ancestor table element has grid role") are NOT modeled.
// Upstream's `attributesComparator` doesn't read them either, so the parity
// is exact: a schema with constraints but no `attributes` matches the
// bare-element form unconditionally. This is a known imprecision in
// eslint-plugin-jsx-a11y; we mirror it for parity. Same for
// attribute-level `constraints` — `attributesComparator` looks only at
// `name` and `value`, so a predicate like `{name: "list", constraints:
// ["set"]}` and `{name: "list", constraints: ["undefined"]}` collapse to
// the same "the `list` attribute exists" check.

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
)

// elementSchema mirrors aria-query's elementSchema entries. `Attributes`
// is the predicate list `attributesComparator` walks; an empty / nil slice
// matches any tag of the same name (vacuous `every`).
type elementSchema struct {
	Name       string
	Attributes []elementAttrSchema
}

// elementAttrSchema is a single attribute predicate inside an elementSchema.
// `Value == ""` matches "attribute exists" only — value comparison is
// skipped when the schema's value is empty. Mirrors upstream's
// `baseAttr.value &&` truthy guard, which falls through for "", null,
// undefined, false, and 0.
type elementAttrSchema struct {
	Name  string
	Value string
}

// interactiveRolesSet is the set of roles whose `superClass` chain contains
// `widget`, PLUS `toolbar` (does not descend from widget but supports
// aria-activedescendant). Mirrors upstream's `interactiveRoles` Set, which
// is derived from aria-query and includes `progressbar` (a widget
// descendant via `range`).
var interactiveRolesSet = map[string]struct{}{
	"button":           {},
	"checkbox":         {},
	"columnheader":     {},
	"combobox":         {},
	"doc-backlink":     {},
	"doc-biblioref":    {},
	"doc-glossref":     {},
	"doc-noteref":      {},
	"grid":             {},
	"gridcell":         {},
	"link":             {},
	"listbox":          {},
	"menu":             {},
	"menubar":          {},
	"menuitem":         {},
	"menuitemcheckbox": {},
	"menuitemradio":    {},
	"option":           {},
	"progressbar":      {},
	"radio":            {},
	"radiogroup":       {},
	"row":              {},
	"rowheader":        {},
	"scrollbar":        {},
	"searchbox":        {},
	"slider":           {},
	"spinbutton":       {},
	"switch":           {},
	"tab":              {},
	"tablist":          {},
	"textbox":          {},
	"toolbar":          {},
	"tree":             {},
	"treegrid":         {},
	"treeitem":         {},
}

// allRolesSet contains every concrete role name in aria-query (lowercased).
// Used by IsInteractiveRole's "first valid role" walk — upstream filters the
// space-split role list down to entries that actually exist in the role
// keys before checking interactivity.
var allRolesSet = func() map[string]struct{} {
	out := make(map[string]struct{}, len(interactiveRolesSet)+len(nonInteractiveRoleNames))
	for k := range interactiveRolesSet {
		out[k] = struct{}{}
	}
	for _, k := range nonInteractiveRoleNames {
		out[k] = struct{}{}
	}
	return out
}()

// nonInteractiveRoleNames is the complement of interactiveRolesSet within
// the non-abstract role keys. Used only to populate allRolesSet.
var nonInteractiveRoleNames = []string{
	"alert", "alertdialog", "application", "article", "banner",
	"blockquote", "caption", "cell", "code", "complementary",
	"contentinfo", "definition", "deletion", "dialog", "directory",
	"doc-abstract", "doc-acknowledgments", "doc-afterword", "doc-appendix",
	"doc-biblioentry", "doc-bibliography", "doc-chapter", "doc-colophon",
	"doc-conclusion", "doc-cover", "doc-credit", "doc-credits",
	"doc-dedication", "doc-endnote", "doc-endnotes", "doc-epigraph",
	"doc-epilogue", "doc-errata", "doc-example", "doc-footnote",
	"doc-foreword", "doc-glossary", "doc-index", "doc-introduction",
	"doc-notice", "doc-pagebreak", "doc-pagefooter", "doc-pageheader",
	"doc-pagelist", "doc-part", "doc-preface", "doc-prologue",
	"doc-pullquote", "doc-qna", "doc-subtitle", "doc-tip", "doc-toc",
	"document", "emphasis", "feed", "figure", "form", "generic",
	"graphics-document", "graphics-object", "graphics-symbol",
	"group", "heading", "img", "insertion", "list", "listitem", "log",
	"main", "mark", "marquee", "math", "meter", "navigation", "none",
	"note", "paragraph", "presentation", "region",
	"rowgroup", "search", "separator", "status", "strong", "subscript",
	"superscript", "table", "tabpanel", "term", "time", "timer",
	"tooltip",
}

// interactiveElementRoleSchemas mirrors upstream's
// `interactiveElementRoleSchemas` — every entry in aria-query's
// `elementRoles` whose role list contains at least one interactive role.
// Order matches the upstream iteration; `attributesComparator` uses
// `Array.some`, so order does not affect correctness.
var interactiveElementRoleSchemas = []elementSchema{
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "button"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "image"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "reset"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "submit"}}},
	{Name: "button"},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "checkbox"}}},
	{Name: "th"},
	{Name: "th", Attributes: []elementAttrSchema{{Name: "scope", Value: "col"}}},
	{Name: "th", Attributes: []elementAttrSchema{{Name: "scope", Value: "colgroup"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "list"}, {Name: "type", Value: "email"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "list"}, {Name: "type", Value: "search"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "list"}, {Name: "type", Value: "tel"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "list"}, {Name: "type", Value: "text"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "list"}, {Name: "type", Value: "url"}}},
	{Name: "select", Attributes: []elementAttrSchema{{Name: "multiple"}, {Name: "size"}}},
	{Name: "td"},
	{Name: "a", Attributes: []elementAttrSchema{{Name: "href"}}},
	{Name: "area", Attributes: []elementAttrSchema{{Name: "href"}}},
	{Name: "select", Attributes: []elementAttrSchema{{Name: "size"}}},
	{Name: "datalist"},
	{Name: "select", Attributes: []elementAttrSchema{{Name: "multiple"}}},
	{Name: "option"},
	{Name: "th", Attributes: []elementAttrSchema{{Name: "scope", Value: "row"}}},
	{Name: "th", Attributes: []elementAttrSchema{{Name: "scope", Value: "rowgroup"}}},
	{Name: "tr"},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "radio"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "range"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "number"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type"}, {Name: "list"}}},
	{Name: "textarea"},
}

// nonInteractiveElementRoleSchemas mirrors upstream's
// `nonInteractiveElementRoleSchemas` — every entry in aria-query's
// `elementRoles` whose role list contains ONLY non-interactive roles.
var nonInteractiveElementRoleSchemas = []elementSchema{
	{Name: "article"},
	{Name: "header"},
	{Name: "blockquote"},
	{Name: "caption"},
	{Name: "td"},
	{Name: "code"},
	{Name: "aside"},
	{Name: "aside", Attributes: []elementAttrSchema{{Name: "aria-label"}}},
	{Name: "aside", Attributes: []elementAttrSchema{{Name: "aria-labelledby"}}},
	{Name: "footer"},
	{Name: "dd"},
	{Name: "del"},
	{Name: "dialog"},
	{Name: "html"},
	{Name: "em"},
	{Name: "figure"},
	{Name: "form", Attributes: []elementAttrSchema{{Name: "aria-label"}}},
	{Name: "form", Attributes: []elementAttrSchema{{Name: "aria-labelledby"}}},
	{Name: "form", Attributes: []elementAttrSchema{{Name: "name"}}},
	{Name: "a"},
	{Name: "area"},
	{Name: "b"},
	{Name: "bdo"},
	{Name: "body"},
	{Name: "data"},
	{Name: "div"},
	{Name: "hgroup"},
	{Name: "i"},
	{Name: "pre"},
	{Name: "q"},
	{Name: "samp"},
	{Name: "section"},
	{Name: "section", Attributes: []elementAttrSchema{{Name: "aria-label"}}},
	{Name: "section", Attributes: []elementAttrSchema{{Name: "aria-labelledby"}}},
	{Name: "small"},
	{Name: "span"},
	{Name: "u"},
	{Name: "details"},
	{Name: "fieldset"},
	{Name: "optgroup"},
	{Name: "address"},
	{Name: "h1"},
	{Name: "h2"},
	{Name: "h3"},
	{Name: "h4"},
	{Name: "h5"},
	{Name: "h6"},
	{Name: "img", Attributes: []elementAttrSchema{{Name: "alt"}}},
	{Name: "img", Attributes: []elementAttrSchema{{Name: "alt", Value: ""}}},
	{Name: "ins"},
	{Name: "menu"},
	{Name: "ol"},
	{Name: "ul"},
	{Name: "li"},
	{Name: "main"},
	{Name: "mark"},
	{Name: "math"},
	{Name: "meter"},
	{Name: "nav"},
	{Name: "p"},
	{Name: "progress"},
	{Name: "tbody"},
	{Name: "tfoot"},
	{Name: "thead"},
	{Name: "hr"},
	{Name: "output"},
	{Name: "strong"},
	{Name: "sub"},
	{Name: "sup"},
	{Name: "table"},
	{Name: "dfn"},
	{Name: "dt"},
	{Name: "time"},
}

// interactiveElementAXSchemas mirrors upstream's
// `interactiveElementAXObjectSchemas` — every entry in axobject-query's
// `elementAXObjects` whose AX-object list is ENTIRELY made of widget-type
// AX objects. Consulted only after interactive/non-interactive role
// schemas miss; the order matches upstream's iteration.
var interactiveElementAXSchemas = []elementSchema{
	{Name: "audio"},
	{Name: "button"},
	{Name: "canvas"},
	{Name: "td"},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "checkbox"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "color"}}},
	{Name: "th"},
	{Name: "select"},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "date"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "datetime"}}},
	{Name: "summary"},
	{Name: "embed"},
	{Name: "input"},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "time"}}},
	{Name: "a", Attributes: []elementAttrSchema{{Name: "href"}}},
	{Name: "option"},
	{Name: "datalist"},
	{Name: "menuitem"},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "radio"}}},
	{Name: "th", Attributes: []elementAttrSchema{{Name: "scope", Value: "row"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "search"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "range"}}},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "number"}}},
	{Name: "textarea"},
	{Name: "input", Attributes: []elementAttrSchema{{Name: "type", Value: "text"}}},
	{Name: "video"},
}

// schemaMatchesAttrs reports whether the given JSX-tag attribute list
// satisfies the elementSchema's attribute predicate list. Mirrors upstream's
// `attributesComparator`:
//
//   - empty Attributes ⇒ vacuous match
//   - each predicate must find at least one matching JsxAttribute
//   - matching: name (case-sensitive, mirroring `propName(attribute)` strict
//     equality) AND, when predicate.Value != "", literal value equality
//   - JsxSpreadAttribute is opaque (mirrors upstream's `attribute.type !==
//     'JSXAttribute'` guard)
func schemaMatchesAttrs(schema elementSchema, attrs []*ast.Node) bool {
	for _, baseAttr := range schema.Attributes {
		found := false
		for _, attr := range attrs {
			if attr.Kind != ast.KindJsxAttribute {
				continue
			}
			if reactutil.GetJsxPropName(attr) != baseAttr.Name {
				continue
			}
			if baseAttr.Value != "" {
				val, ok := LiteralStringValue(attr)
				if !ok || val != baseAttr.Value {
					continue
				}
			}
			found = true
			break
		}
		if !found {
			return false
		}
	}
	return true
}

// IsInteractiveElement reports whether `(tagName, attrs)` denotes an
// interactive HTML element by ARIA semantics. Mirrors upstream's
// `isInteractiveElement(tagName, attributes)`:
//
//  1. Tag must be in aria-query's `dom` map (custom components return
//     false unconditionally).
//  2. First match against interactive role schemas → true.
//  3. First match against non-interactive role schemas → false.
//  4. First match against interactive AX-object schemas → true.
//  5. Otherwise → false.
//
// Schema-level constraints are not modeled (see file-level note).
func IsInteractiveElement(tagName string, attrs []*ast.Node) bool {
	if !IsDOMElement(tagName) {
		return false
	}
	for _, s := range interactiveElementRoleSchemas {
		if s.Name == tagName && schemaMatchesAttrs(s, attrs) {
			return true
		}
	}
	for _, s := range nonInteractiveElementRoleSchemas {
		if s.Name == tagName && schemaMatchesAttrs(s, attrs) {
			return false
		}
	}
	for _, s := range interactiveElementAXSchemas {
		if s.Name == tagName && schemaMatchesAttrs(s, attrs) {
			return true
		}
	}
	return false
}

// IsInteractiveRole mirrors upstream's `isInteractiveRole(_, attributes)` —
// the tag-name parameter is unused. Returns true when the `role` attribute's
// LITERAL value (lowercased, space-split) has a first valid role that is in
// the interactive (widget) set.
//
// Non-literal expressions (`role={someVar}`), absent role, or first valid
// role that is non-interactive all return false. Empty / all-invalid
// space-split lists also return false.
//
// Routes through [LiteralPropStringValue] (= upstream `getLiteralPropValue`)
// rather than the simpler [LiteralStringValue]: this picks up upstream's
// TemplateExpression-with-substitutions synthesis (each substitution's
// LITERAL_TYPES extraction concatenated against the quasi text) and the
// `null` literal → "null" magic string. Matters for `` role={`button${''}`} ``
// and `role={null}` corner cases — uncommon, but covered by upstream's
// `getLiteralPropValue` and therefore in scope for parity.
func IsInteractiveRole(_ string, attrs []*ast.Node) bool {
	roleAttr := FindAttributeByName(attrs, "role")
	if roleAttr == nil {
		return false
	}
	value, ok := LiteralPropStringValue(roleAttr)
	if !ok {
		return false
	}
	// `String(value).toLowerCase().split(' ')` — single-space split, so
	// double-space input produces an empty entry that filters out at the
	// `allRolesSet` check below. Use raw Split rather than Fields so the
	// boundary cases match upstream byte-for-byte.
	for _, part := range strings.Split(strings.ToLower(value), " ") {
		if _, isRole := allRolesSet[part]; !isRole {
			continue
		}
		// First valid role wins per upstream's `validRoles[0]`.
		_, isInteractive := interactiveRolesSet[part]
		return isInteractive
	}
	return false
}

// nonInteractiveRolesSet mirrors upstream's `isNonInteractiveRole`
// `nonInteractiveRoles` Set — every non-abstract role whose `superClass`
// chain does NOT contain `widget`. Note the asymmetry with [interactiveRolesSet]:
//
//   - `toolbar` lives in BOTH sets upstream. interactiveRoles concat's it
//     in (does not descend from widget, but supports aria-activedescendant);
//     nonInteractiveRoles' filter (superClass !⊇ widget) ALSO matches it.
//   - `progressbar` lives only in interactiveRolesSet — it IS a widget
//     descendant, so the nonInteractiveRoles filter rejects it.
//
// Each individual upstream util reuses its OWN local filter pipeline; this
// duplication is load-bearing for the observable semantics of
// `interactive-supports-focus` and similar rules. Maintain the two sets
// separately rather than computing one as the complement of the other.
var nonInteractiveRolesSet = func() map[string]struct{} {
	out := make(map[string]struct{}, len(nonInteractiveRoleNames)+1)
	for _, k := range nonInteractiveRoleNames {
		out[k] = struct{}{}
	}
	// `toolbar` matches the upstream `!superClass.some(...widget)` filter.
	// nonInteractiveRoleNames intentionally omits it (it lives in
	// interactiveRolesSet for the isInteractiveRole path); replicate the
	// upstream membership explicitly here.
	out["toolbar"] = struct{}{}
	return out
}()

// IsNonInteractiveRole mirrors upstream's `isNonInteractiveRole(tagName,
// attributes)`. The function is tag-aware (unlike [IsInteractiveRole]) —
// upstream early-returns false for custom components, so the rule's caller
// can short-circuit without consulting role attributes on JSX-component tags.
//
// Returns true iff:
//   - tagName is an aria-query DOM element name, AND
//   - the `role` attribute resolves to a literal string whose first valid
//     space-separated role is in [nonInteractiveRolesSet].
//
// Returns false when the role expression is non-literal, when no valid role
// is found in the space-separated list, or when tagName is a custom
// component (matches upstream's `!dom.has(type)` guard).
func IsNonInteractiveRole(tagName string, attrs []*ast.Node) bool {
	if !IsDOMElement(tagName) {
		return false
	}
	roleAttr := FindAttributeByName(attrs, "role")
	if roleAttr == nil {
		return false
	}
	// Same extractor as IsInteractiveRole — see that comment for the
	// LiteralPropStringValue vs LiteralStringValue rationale.
	value, ok := LiteralPropStringValue(roleAttr)
	if !ok {
		return false
	}
	for _, part := range strings.Split(strings.ToLower(value), " ") {
		if _, isRole := allRolesSet[part]; !isRole {
			continue
		}
		_, isNonInteractive := nonInteractiveRolesSet[part]
		return isNonInteractive
	}
	return false
}

// abstractRolesSet mirrors upstream's `isAbstractRole.js` -
// `new Set(roles.keys().filter((role) => roles.get(role).abstract))` -
// every aria-query role whose definition carries `abstract: true`. These
// roles are not assignable to HTML elements at runtime; some rules (e.g.
// no-noninteractive-element-interactions) treat them as "no opinion"
// short-circuits.
//
// Source: https://github.com/A11yance/aria-query/blob/main/src/etc/roles/abstract/
var abstractRolesSet = map[string]struct{}{
	"command":     {},
	"composite":   {},
	"input":       {},
	"landmark":    {},
	"range":       {},
	"roletype":    {},
	"section":     {},
	"sectionhead": {},
	"select":      {},
	"structure":   {},
	"widget":      {},
	"window":      {},
}

// IsAbstractRole mirrors upstream `isAbstractRole(tagName, attributes)`:
//
//	if (!DOMElements.has(tagName)) return false;
//	const role = getLiteralPropValue(getProp(attributes, 'role'));
//	return abstractRoles.has(role);
//
// Notable upstream quirks we preserve:
//
//   - Custom components short-circuit to false (upstream `!DOMElements.has`).
//   - The role lookup uses `getLiteralPropValue` (= [LiteralPropStringValue]),
//     so non-literal `role` expressions (Identifier, CallExpression,
//     ConditionalExpression, …) silently produce a `false` here.
//   - Set membership is whole-string and case-sensitive — upstream does NOT
//     lowercase or split on whitespace, unlike the interactive/non-interactive
//     role predicates. `role=" widget "` (with surrounding whitespace) and
//     `role="WIDGET"` therefore do NOT match. Mirror.
func IsAbstractRole(tagName string, attrs []*ast.Node) bool {
	if !IsDOMElement(tagName) {
		return false
	}
	roleAttr := FindAttributeByName(attrs, "role")
	if roleAttr == nil {
		return false
	}
	value, ok := LiteralPropStringValue(roleAttr)
	if !ok {
		return false
	}
	_, isAbstract := abstractRolesSet[value]
	return isAbstract
}

// strictNonInteractiveElementRoleSchemas mirrors upstream
// `isNonInteractiveElement.js`'s view of `nonInteractiveElementRoleSchemas`
// — `elementRoles` entries whose role list is ENTIRELY non-interactive
// under that file's stricter `nonInteractiveRoles` set:
//
//	{role | !role.abstract
//	     && role !== 'toolbar'
//	     && role !== 'generic'
//	     && !role.superClass.some(...widget)}
//	∪ {'progressbar'}
//
// Differs from [nonInteractiveElementRoleSchemas]: that table is generated
// from `isInteractiveElement.js`'s `nonInteractiveRoles` set (which keeps
// `generic` in the non-interactive set), so generic-role HTML elements
// like `div`, `span`, `a`, `area`, `b`, `bdo`, `body`, `data`, `hgroup`,
// `i`, `pre`, `q`, `samp`, `small`, `u`, and `section` (no attrs) appear
// there but NOT here.
//
// `IsNonInteractiveElement` consults THIS table (not the one above) — a
// `<div role="radio" onClick={…} />` must trip `interactive-supports-focus`,
// which requires `IsNonInteractiveElement("div", …)` to return false. If
// we used the `isInteractiveElement.js`-view table, `div` would match the
// non-interactive schema and the rule would fall silent.
//
// Generated from `aria-query` v5 + the upstream filter pipeline.
// Element-level constraints are not modeled (see file-level note).
var strictNonInteractiveElementRoleSchemas = []elementSchema{
	{Name: "article"},
	{Name: "header"},
	{Name: "blockquote"},
	{Name: "caption"},
	{Name: "td"},
	{Name: "code"},
	{Name: "aside"},
	{Name: "aside", Attributes: []elementAttrSchema{{Name: "aria-label"}}},
	{Name: "aside", Attributes: []elementAttrSchema{{Name: "aria-labelledby"}}},
	{Name: "footer"},
	{Name: "dd"},
	{Name: "del"},
	{Name: "dialog"},
	{Name: "html"},
	{Name: "em"},
	{Name: "figure"},
	{Name: "form", Attributes: []elementAttrSchema{{Name: "aria-label"}}},
	{Name: "form", Attributes: []elementAttrSchema{{Name: "aria-labelledby"}}},
	{Name: "form", Attributes: []elementAttrSchema{{Name: "name"}}},
	{Name: "details"},
	{Name: "fieldset"},
	{Name: "optgroup"},
	{Name: "address"},
	{Name: "h1"},
	{Name: "h2"},
	{Name: "h3"},
	{Name: "h4"},
	{Name: "h5"},
	{Name: "h6"},
	{Name: "img", Attributes: []elementAttrSchema{{Name: "alt"}}},
	{Name: "img", Attributes: []elementAttrSchema{{Name: "alt", Value: ""}}},
	{Name: "ins"},
	{Name: "menu"},
	{Name: "ol"},
	{Name: "ul"},
	{Name: "li"},
	{Name: "main"},
	{Name: "mark"},
	{Name: "math"},
	{Name: "meter"},
	{Name: "nav"},
	{Name: "p"},
	{Name: "progress"},
	{Name: "section", Attributes: []elementAttrSchema{{Name: "aria-label"}}},
	{Name: "section", Attributes: []elementAttrSchema{{Name: "aria-labelledby"}}},
	{Name: "tbody"},
	{Name: "tfoot"},
	{Name: "thead"},
	{Name: "hr"},
	{Name: "output"},
	{Name: "strong"},
	{Name: "sub"},
	{Name: "sup"},
	{Name: "table"},
	{Name: "dfn"},
	{Name: "dt"},
	{Name: "time"},
}

// nonInteractiveElementAXSchemas mirrors upstream's
// `nonInteractiveElementAXObjectSchemas` — every entry in axobject-query's
// `elementAXObjects` whose AX-object list is ENTIRELY made of window/structure
// AX objects. Consulted only after the role-schema matchers miss; ordering
// matches upstream's iteration over `elementAXObjects`.
//
// Generated from `axobject-query` v4 + `AXObjects.type ∈ {window, structure}`
// using the same filter pipeline as upstream. Element-level constraints are
// not modeled (see file-level note).
var nonInteractiveElementAXSchemas = []elementSchema{
	{Name: "abbr"},
	{Name: "article"},
	{Name: "blockquote"},
	{Name: "caption"},
	{Name: "dfn"},
	{Name: "dd"},
	{Name: "dl"},
	{Name: "dt"},
	{Name: "details"},
	{Name: "dialog"},
	{Name: "dir"},
	{Name: "figcaption"},
	{Name: "figure"},
	{Name: "footer"},
	{Name: "form"},
	{Name: "h1"},
	{Name: "h2"},
	{Name: "h3"},
	{Name: "h4"},
	{Name: "h5"},
	{Name: "h6"},
	{Name: "iframe"},
	{Name: "img", Attributes: []elementAttrSchema{{Name: "usemap"}}},
	{Name: "img"},
	{Name: "label"},
	{Name: "legend"},
	{Name: "br"},
	{Name: "li"},
	{Name: "ul"},
	{Name: "ol"},
	{Name: "main"},
	{Name: "mark"},
	{Name: "marquee"},
	{Name: "menu"},
	{Name: "meter"},
	{Name: "nav"},
	{Name: "p"},
	{Name: "pre"},
	{Name: "progress"},
	{Name: "tr"},
	{Name: "ruby"},
	{Name: "table"},
	{Name: "time"},
}

// IsNonInteractiveElement reports whether `(tagName, attrs)` denotes a
// non-interactive HTML element by ARIA semantics. Mirrors upstream's
// `isNonInteractiveElement(tagName, attributes)`:
//
//  1. Tag must be in aria-query's `dom` map; otherwise false (custom
//     components — we don't know what DOM they render).
//  2. `<header>` always returns false — upstream comments call out that
//     header semantics depend on parent context (banner landmark inside
//     <body>), which static analysis cannot verify.
//  3. First match against non-interactive role schemas (with the `td`
//     exception below) → true.
//  4. First match against interactive role schemas (same `td` exception)
//     → false.
//  5. First match against non-interactive AX-object schemas → true.
//  6. Otherwise → false.
//
// `td` is excluded from the role-schema matchers via upstream's
// `tagName !== 'td'` guard — both `interactiveElementRoleSchemas` and
// `nonInteractiveElementRoleSchemas` contain a `{name: "td"}` entry, so
// without the guard the matchers would both fire. Upstream documents this
// as a TODO; we mirror byte-for-byte.
func IsNonInteractiveElement(tagName string, attrs []*ast.Node) bool {
	if !IsDOMElement(tagName) {
		return false
	}
	if tagName == "header" {
		return false
	}
	if tagName != "td" {
		for _, s := range strictNonInteractiveElementRoleSchemas {
			if s.Name == tagName && schemaMatchesAttrs(s, attrs) {
				return true
			}
		}
		for _, s := range interactiveElementRoleSchemas {
			if s.Name == tagName && schemaMatchesAttrs(s, attrs) {
				return false
			}
		}
	}
	for _, s := range nonInteractiveElementAXSchemas {
		if s.Name == tagName && schemaMatchesAttrs(s, attrs) {
			return true
		}
	}
	return false
}
