package jsxa11yutil

// DomElements mirrors aria-query's `dom.keys()` ordered list — the set of
// HTML element names that aria-query's `dom` map contains. Used by rules
// (no-autofocus, …) that need to distinguish DOM elements from custom
// components without depending on aria-query at runtime.
//
// Source: https://github.com/A11yance/aria-query/blob/main/src/domMap.ts
var DomElements = []string{
	"a", "abbr", "acronym", "address", "applet", "area", "article", "aside",
	"audio", "b", "base", "bdi", "bdo", "big", "blink", "blockquote", "body",
	"br", "button", "canvas", "caption", "center", "cite", "code", "col",
	"colgroup", "content", "data", "datalist", "dd", "del", "details", "dfn",
	"dialog", "dir", "div", "dl", "dt", "em", "embed", "fieldset", "figcaption",
	"figure", "font", "footer", "form", "frame", "frameset", "h1", "h2", "h3",
	"h4", "h5", "h6", "head", "header", "hgroup", "hr", "html", "i", "iframe",
	"img", "input", "ins", "kbd", "keygen", "label", "legend", "li", "link",
	"main", "map", "mark", "marquee", "menu", "menuitem", "meta", "meter",
	"nav", "noembed", "noscript", "object", "ol", "optgroup", "option",
	"output", "p", "param", "picture", "pre", "progress", "q", "rp", "rt",
	"rtc", "ruby", "s", "samp", "script", "section", "select", "small",
	"source", "spacer", "span", "strike", "strong", "style", "sub", "summary",
	"sup", "table", "tbody", "td", "textarea", "tfoot", "th", "thead", "time",
	"title", "tr", "track", "tt", "u", "ul", "var", "video", "wbr", "xmp",
}

var domSet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(DomElements))
	for _, el := range DomElements {
		set[el] = struct{}{}
	}
	return set
}()

// IsDOMElement reports whether `name` is an HTML element name in
// aria-query's `dom` map. Mirrors upstream `dom.get(type)` truthy check used
// by `no-autofocus`'s `ignoreNonDOM` option to skip custom components.
// Comparison is exact (case-sensitive) — upstream's `dom.get(type)` is a
// Map.get with no normalization, and `getElementType(node)` already returns
// the resolved element name verbatim.
func IsDOMElement(name string) bool {
	_, ok := domSet[name]
	return ok
}
