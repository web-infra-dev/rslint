package jsxa11yutil

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
