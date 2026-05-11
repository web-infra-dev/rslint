package jsxa11yutil

import "strings"

// AriaPropertyNames is the list of every ARIA state / property defined by
// `aria-query`'s `ariaPropsMap`. The order mirrors `aria.keys()`, so callers
// that need a deterministic iteration order (e.g. suggestion ranking for
// jsx-a11y/aria-props) get the same ordering as upstream.
//
// Source: https://github.com/A11yance/aria-query/blob/main/src/ariaPropsMap.js
var AriaPropertyNames = []string{
	"aria-activedescendant",
	"aria-atomic",
	"aria-autocomplete",
	"aria-braillelabel",
	"aria-brailleroledescription",
	"aria-busy",
	"aria-checked",
	"aria-colcount",
	"aria-colindex",
	"aria-colspan",
	"aria-controls",
	"aria-current",
	"aria-describedby",
	"aria-description",
	"aria-details",
	"aria-disabled",
	"aria-dropeffect",
	"aria-errormessage",
	"aria-expanded",
	"aria-flowto",
	"aria-grabbed",
	"aria-haspopup",
	"aria-hidden",
	"aria-invalid",
	"aria-keyshortcuts",
	"aria-label",
	"aria-labelledby",
	"aria-level",
	"aria-live",
	"aria-modal",
	"aria-multiline",
	"aria-multiselectable",
	"aria-orientation",
	"aria-owns",
	"aria-placeholder",
	"aria-posinset",
	"aria-pressed",
	"aria-readonly",
	"aria-relevant",
	"aria-required",
	"aria-roledescription",
	"aria-rowcount",
	"aria-rowindex",
	"aria-rowspan",
	"aria-selected",
	"aria-setsize",
	"aria-sort",
	"aria-valuemax",
	"aria-valuemin",
	"aria-valuenow",
	"aria-valuetext",
}

// AriaPropertySet mirrors `aria-query`'s `aria.has(key)` lookup as a Go set.
// Keys are the same lowercase canonical names from `AriaPropertyNames`. Used
// when a rule needs to test "is this attribute a recognized ARIA property?"
// without caring about iteration order.
var AriaPropertySet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(AriaPropertyNames))
	for _, name := range AriaPropertyNames {
		set[name] = struct{}{}
	}
	return set
}()

// AriaPropertyNamesUpper holds `AriaPropertyNames` pre-converted to upper
// case, 1:1 indexed. Suggestion ranking compares against an upper-cased
// candidate name, and computing the upper form once at init time avoids
// `strings.ToUpper` allocations on every suggestion query.
var AriaPropertyNamesUpper = func() []string {
	out := make([]string, len(AriaPropertyNames))
	for i, name := range AriaPropertyNames {
		out[i] = strings.ToUpper(name)
	}
	return out
}()
