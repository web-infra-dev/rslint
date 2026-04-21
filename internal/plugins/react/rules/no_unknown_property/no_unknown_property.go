// cspell:disable — this file enumerates HTML / SVG / ARIA attribute names
// verbatim from React and the WHATWG / W3C specs, so it contains many
// attribute-name tokens (aria-*, SVG presentation attributes, popover /
// shadowroot attrs, …) that are not in a general English dictionary.
package no_unknown_property

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoUnknownPropertyRule = rule.Rule{
	Name: "react/no-unknown-property",
	Run:  runRule,
}

// domAttributeNames maps HTML-cased attribute names to the React DOM property
// name that should replace them. Mirrors eslint-plugin-react's
// DOM_ATTRIBUTE_NAMES table.
var domAttributeNames = map[string]string{
	"accept-charset": "acceptCharset",
	"class":          "className",
	"http-equiv":     "httpEquiv",
	"crossorigin":    "crossOrigin",
	"for":            "htmlFor",
	"nomodule":       "noModule",
}

// svgDomAttributeNames maps SVG hyphenated / colon attribute names to the
// React property name. Mirrors upstream's SVGDOM_ATTRIBUTE_NAMES.
var svgDomAttributeNames = map[string]string{
	"accent-height":                "accentHeight",
	"alignment-baseline":           "alignmentBaseline",
	"arabic-form":                  "arabicForm",
	"baseline-shift":               "baselineShift",
	"cap-height":                   "capHeight",
	"clip-path":                    "clipPath",
	"clip-rule":                    "clipRule",
	"color-interpolation":          "colorInterpolation",
	"color-interpolation-filters":  "colorInterpolationFilters",
	"color-profile":                "colorProfile",
	"color-rendering":              "colorRendering",
	"dominant-baseline":            "dominantBaseline",
	"enable-background":            "enableBackground",
	"fill-opacity":                 "fillOpacity",
	"fill-rule":                    "fillRule",
	"flood-color":                  "floodColor",
	"flood-opacity":                "floodOpacity",
	"font-family":                  "fontFamily",
	"font-size":                    "fontSize",
	"font-size-adjust":             "fontSizeAdjust",
	"font-stretch":                 "fontStretch",
	"font-style":                   "fontStyle",
	"font-variant":                 "fontVariant",
	"font-weight":                  "fontWeight",
	"glyph-name":                   "glyphName",
	"glyph-orientation-horizontal": "glyphOrientationHorizontal",
	"glyph-orientation-vertical":   "glyphOrientationVertical",
	"horiz-adv-x":                  "horizAdvX",
	"horiz-origin-x":               "horizOriginX",
	"image-rendering":              "imageRendering",
	"letter-spacing":               "letterSpacing",
	"lighting-color":               "lightingColor",
	"marker-end":                   "markerEnd",
	"marker-mid":                   "markerMid",
	"marker-start":                 "markerStart",
	"overline-position":            "overlinePosition",
	"overline-thickness":           "overlineThickness",
	"paint-order":                  "paintOrder",
	"panose-1":                     "panose1",
	"pointer-events":               "pointerEvents",
	"rendering-intent":             "renderingIntent",
	"shape-rendering":              "shapeRendering",
	"stop-color":                   "stopColor",
	"stop-opacity":                 "stopOpacity",
	"strikethrough-position":       "strikethroughPosition",
	"strikethrough-thickness":      "strikethroughThickness",
	"stroke-dasharray":             "strokeDasharray",
	"stroke-dashoffset":            "strokeDashoffset",
	"stroke-linecap":               "strokeLinecap",
	"stroke-linejoin":              "strokeLinejoin",
	"stroke-miterlimit":            "strokeMiterlimit",
	"stroke-opacity":               "strokeOpacity",
	"stroke-width":                 "strokeWidth",
	"text-anchor":                  "textAnchor",
	"text-decoration":              "textDecoration",
	"text-rendering":               "textRendering",
	"underline-position":           "underlinePosition",
	"underline-thickness":          "underlineThickness",
	"unicode-bidi":                 "unicodeBidi",
	"unicode-range":                "unicodeRange",
	"units-per-em":                 "unitsPerEm",
	"v-alphabetic":                 "vAlphabetic",
	"v-hanging":                    "vHanging",
	"v-ideographic":                "vIdeographic",
	"v-mathematical":               "vMathematical",
	"vector-effect":                "vectorEffect",
	"vert-adv-y":                   "vertAdvY",
	"vert-origin-x":                "vertOriginX",
	"vert-origin-y":                "vertOriginY",
	"word-spacing":                 "wordSpacing",
	"writing-mode":                 "writingMode",
	"x-height":                     "xHeight",
	"xlink:actuate":                "xlinkActuate",
	"xlink:arcrole":                "xlinkArcrole",
	"xlink:href":                   "xlinkHref",
	"xlink:role":                   "xlinkRole",
	"xlink:show":                   "xlinkShow",
	"xlink:title":                  "xlinkTitle",
	"xlink:type":                   "xlinkType",
	"xml:base":                     "xmlBase",
	"xml:lang":                     "xmlLang",
	"xml:space":                    "xmlSpace",
}

// attributeTagsMap lists the allowed tags for attributes whose usage is
// element-specific. The invalidPropOnTag message serializes the per-key
// []string with ", ", so each key's value-slice order mirrors upstream's
// ATTRIBUTE_TAGS_MAP declaration order.
var attributeTagsMap = map[string][]string{
	"abbr":         {"th", "td"},
	"charset":      {"meta"},
	"checked":      {"input"},
	"closedby":     {"dialog"},
	"crossOrigin":  {"script", "img", "video", "audio", "link", "image"},
	"displaystyle": {"math"},
	"download":     {"a", "area"},
	"fill": {
		"altGlyph", "circle", "ellipse", "g", "line", "marker", "mask",
		"path", "polygon", "polyline", "rect", "svg", "symbol", "text",
		"textPath", "tref", "tspan", "use",
		"animate", "animateColor", "animateMotion", "animateTransform", "set",
	},
	"focusable":   {"svg"},
	"imageSizes":  {"link"},
	"imageSrcSet": {"link"},
	"property":    {"meta"},
	"viewBox":     {"marker", "pattern", "svg", "symbol", "view"},
	"as":          {"link"},
	"align": {
		"applet", "caption", "col", "colgroup", "hr", "iframe", "img",
		"table", "tbody", "td", "tfoot", "th", "thead", "tr",
	},
	"valign":                   {"tr", "td", "th", "thead", "tbody", "tfoot", "colgroup", "col"},
	"noModule":                 {"script"},
	"onAbort":                  {"audio", "video"},
	"onCancel":                 {"dialog"},
	"onCanPlay":                {"audio", "video"},
	"onCanPlayThrough":         {"audio", "video"},
	"onClose":                  {"dialog"},
	"onDurationChange":         {"audio", "video"},
	"onEmptied":                {"audio", "video"},
	"onEncrypted":              {"audio", "video"},
	"onEnded":                  {"audio", "video"},
	"onError":                  {"audio", "video", "img", "link", "source", "script", "picture", "iframe"},
	"onLoad":                   {"script", "img", "link", "picture", "iframe", "object", "source", "body"},
	"onLoadedData":             {"audio", "video"},
	"onLoadedMetadata":         {"audio", "video"},
	"onLoadStart":              {"audio", "video"},
	"onPause":                  {"audio", "video"},
	"onPlay":                   {"audio", "video"},
	"onPlaying":                {"audio", "video"},
	"onProgress":               {"audio", "video"},
	"onRateChange":             {"audio", "video"},
	"onResize":                 {"audio", "video"},
	"onSeeked":                 {"audio", "video"},
	"onSeeking":                {"audio", "video"},
	"onStalled":                {"audio", "video"},
	"onSuspend":                {"audio", "video"},
	"onTimeUpdate":             {"audio", "video"},
	"onVolumeChange":           {"audio", "video"},
	"onWaiting":                {"audio", "video"},
	"autoPictureInPicture":     {"video"},
	"controls":                 {"audio", "video"},
	"controlsList":             {"audio", "video"},
	"disablePictureInPicture":  {"video"},
	"disableRemotePlayback":    {"audio", "video"},
	"loop":                     {"audio", "video"},
	"muted":                    {"audio", "video"},
	"playsInline":              {"video"},
	"allowFullScreen":          {"iframe", "video"},
	"webkitAllowFullScreen":    {"iframe", "video"},
	"mozAllowFullScreen":       {"iframe", "video"},
	"poster":                   {"video"},
	"preload":                  {"audio", "video"},
	"scrolling":                {"iframe"},
	"returnValue":              {"dialog"},
	"webkitDirectory":          {"input"},
	"shadowrootmode":           {"template"},
	"shadowrootclonable":       {"template"},
	"shadowrootdelegatesfocus": {"template"},
	"shadowrootserializable":   {"template"},
	"transform-origin":         {"rect"},
}

// domPropertyNamesOneWord is the set of single-word DOM property names the
// rule treats as always valid (case-insensitively). Mirrors upstream's
// DOM_PROPERTY_NAMES_ONE_WORD.
var domPropertyNamesOneWord = []string{
	"dir", "draggable", "hidden", "id", "lang", "nonce", "part", "slot", "style", "title", "translate", "inert",
	"accept", "action", "allow", "alt", "as", "async", "buffered", "capture", "challenge", "cite", "code", "cols",
	"content", "coords", "csp", "data", "decoding", "default", "defer", "disabled", "form",
	"headers", "height", "high", "href", "icon", "importance", "integrity", "kind", "label",
	"language", "loading", "list", "loop", "low", "manifest", "max", "media", "method", "min", "multiple", "muted",
	"name", "open", "optimum", "pattern", "ping", "placeholder", "poster", "preload", "profile",
	"rel", "required", "reversed", "role", "rows", "sandbox", "scope", "seamless", "selected", "shape", "size", "sizes",
	"span", "src", "start", "step", "summary", "target", "type", "value", "width", "wmode", "wrap",
	"accumulate", "additive", "alphabetic", "amplitude", "ascent", "azimuth", "bbox", "begin",
	"bias", "by", "clip", "color", "cursor", "cx", "cy", "d", "decelerate", "descent", "direction",
	"display", "divisor", "dur", "dx", "dy", "elevation", "end", "exponent", "fill", "filter",
	"format", "from", "fr", "fx", "fy", "g1", "g2", "hanging", "height", "hreflang", "ideographic",
	"in", "in2", "intercept", "k", "k1", "k2", "k3", "k4", "kerning", "local", "mask", "mode",
	"offset", "opacity", "operator", "order", "orient", "orientation", "origin", "overflow", "path",
	"ping", "points", "r", "radius", "rel", "restart", "result", "rotate", "rx", "ry", "scale",
	"seed", "slope", "spacing", "speed", "stemh", "stemv", "string", "stroke", "to", "transform",
	"u1", "u2", "unicode", "values", "version", "visibility", "widths", "x", "x1", "x2", "xmlns",
	"y", "y1", "y2", "z",
	"property",
	"ref", "key", "children",
	"results", "security",
	"controls",
	"popover", "popovertarget", "popovertargetaction",
}

// domPropertyNamesTwoWords is the camel-cased DOM property list. Mirrors
// upstream's DOM_PROPERTY_NAMES_TWO_WORDS. Lookups are case-insensitive.
var domPropertyNamesTwoWords = []string{
	"accessKey", "autoCapitalize", "autoFocus", "contentEditable", "enterKeyHint", "exportParts",
	"inputMode", "itemID", "itemRef", "itemProp", "itemScope", "itemType", "spellCheck", "tabIndex",
	"acceptCharset", "autoComplete", "autoPlay", "border", "cellPadding", "cellSpacing", "classID", "codeBase",
	"colSpan", "contextMenu", "dateTime", "encType", "formAction", "formEncType", "formMethod", "formNoValidate", "formTarget",
	"frameBorder", "hrefLang", "httpEquiv", "imageSizes", "imageSrcSet", "isMap", "keyParams", "keyType", "marginHeight", "marginWidth",
	"maxLength", "mediaGroup", "minLength", "noValidate", "onAnimationEnd", "onAnimationIteration", "onAnimationStart",
	"onBlur", "onChange", "onClick", "onContextMenu", "onCopy", "onCompositionEnd", "onCompositionStart",
	"onCompositionUpdate", "onCut", "onDoubleClick", "onDrag", "onDragEnd", "onDragEnter", "onDragExit", "onDragLeave",
	"onError", "onFocus", "onInput", "onKeyDown", "onKeyPress", "onKeyUp", "onLoad", "onWheel", "onDragOver",
	"onDragStart", "onDrop", "onMouseDown", "onMouseEnter", "onMouseLeave", "onMouseMove", "onMouseOut", "onMouseOver",
	"onMouseUp", "onPaste", "onScroll", "onScrollEnd", "onSelect", "onSubmit", "onBeforeToggle", "onToggle", "onTransitionEnd", "radioGroup",
	"readOnly", "referrerPolicy", "rowSpan", "srcDoc", "srcLang", "srcSet", "useMap", "fetchPriority",
	"crossOrigin", "accentHeight", "alignmentBaseline", "arabicForm", "attributeName",
	"attributeType", "baseFrequency", "baselineShift", "baseProfile", "calcMode", "capHeight",
	"clipPathUnits", "clipPath", "clipRule", "colorInterpolation", "colorInterpolationFilters",
	"colorProfile", "colorRendering", "contentScriptType", "contentStyleType", "diffuseConstant",
	"dominantBaseline", "edgeMode", "enableBackground", "fillOpacity", "fillRule", "filterRes",
	"filterUnits", "floodColor", "floodOpacity", "fontFamily", "fontSize", "fontSizeAdjust",
	"fontStretch", "fontStyle", "fontVariant", "fontWeight", "glyphName",
	"glyphOrientationHorizontal", "glyphOrientationVertical", "glyphRef", "gradientTransform",
	"gradientUnits", "horizAdvX", "horizOriginX", "imageRendering", "kernelMatrix",
	"kernelUnitLength", "keyPoints", "keySplines", "keyTimes", "lengthAdjust", "letterSpacing",
	"lightingColor", "limitingConeAngle", "markerEnd", "markerMid", "markerStart", "markerHeight",
	"markerUnits", "markerWidth", "maskContentUnits", "maskUnits", "mathematical", "numOctaves",
	"overlinePosition", "overlineThickness", "panose1", "paintOrder", "pathLength",
	"patternContentUnits", "patternTransform", "patternUnits", "pointerEvents", "pointsAtX",
	"pointsAtY", "pointsAtZ", "preserveAlpha", "preserveAspectRatio", "primitiveUnits",
	"referrerPolicy", "refX", "refY", "rendering-intent", "repeatCount", "repeatDur",
	"requiredExtensions", "requiredFeatures", "shapeRendering", "specularConstant",
	"specularExponent", "spreadMethod", "startOffset", "stdDeviation", "stitchTiles", "stopColor",
	"stopOpacity", "strikethroughPosition", "strikethroughThickness", "strokeDasharray",
	"strokeDashoffset", "strokeLinecap", "strokeLinejoin", "strokeMiterlimit", "strokeOpacity",
	"strokeWidth", "surfaceScale", "systemLanguage", "tableValues", "targetX", "targetY",
	"textAnchor", "textDecoration", "textRendering", "textLength", "transformOrigin",
	"underlinePosition", "underlineThickness", "unicodeBidi", "unicodeRange", "unitsPerEm",
	"vAlphabetic", "vHanging", "vIdeographic", "vMathematical", "vectorEffect", "vertAdvY",
	"vertOriginX", "vertOriginY", "viewBox", "viewTarget", "wordSpacing", "writingMode", "xHeight",
	"xChannelSelector", "xlinkActuate", "xlinkArcrole", "xlinkHref", "xlinkRole", "xlinkShow",
	"xlinkTitle", "xlinkType", "xmlBase", "xmlLang", "xmlnsXlink", "xmlSpace", "yChannelSelector",
	"zoomAndPan",
	"autoCorrect",
	"autoSave",
	"className", "dangerouslySetInnerHTML", "defaultValue", "defaultChecked", "htmlFor",
	"onBeforeInput", "onChange",
	"onInvalid", "onReset", "onTouchCancel", "onTouchEnd", "onTouchMove", "onTouchStart", "suppressContentEditableWarning", "suppressHydrationWarning",
	"onAbort", "onCanPlay", "onCanPlayThrough", "onDurationChange", "onEmptied", "onEncrypted", "onEnded",
	"onLoadedData", "onLoadedMetadata", "onLoadStart", "onPause", "onPlay", "onPlaying", "onProgress", "onRateChange", "onResize",
	"onSeeked", "onSeeking", "onStalled", "onSuspend", "onTimeUpdate", "onVolumeChange", "onWaiting",
	"onCopyCapture", "onCutCapture", "onPasteCapture", "onCompositionEndCapture", "onCompositionStartCapture", "onCompositionUpdateCapture",
	"onFocusCapture", "onBlurCapture", "onChangeCapture", "onBeforeInputCapture", "onInputCapture", "onResetCapture", "onSubmitCapture",
	"onInvalidCapture", "onLoadCapture", "onErrorCapture", "onKeyDownCapture", "onKeyPressCapture", "onKeyUpCapture",
	"onAbortCapture", "onCanPlayCapture", "onCanPlayThroughCapture", "onDurationChangeCapture", "onEmptiedCapture", "onEncryptedCapture",
	"onEndedCapture", "onLoadedDataCapture", "onLoadedMetadataCapture", "onLoadStartCapture", "onPauseCapture", "onPlayCapture",
	"onPlayingCapture", "onProgressCapture", "onRateChangeCapture", "onSeekedCapture", "onSeekingCapture", "onStalledCapture", "onSuspendCapture",
	"onTimeUpdateCapture", "onVolumeChangeCapture", "onWaitingCapture", "onSelectCapture", "onTouchCancelCapture", "onTouchEndCapture",
	"onTouchMoveCapture", "onTouchStartCapture", "onScrollCapture", "onScrollEndCapture", "onWheelCapture", "onAnimationEndCapture", "onAnimationIteration",
	"onAnimationStartCapture", "onTransitionEndCapture",
	"onAuxClick", "onAuxClickCapture", "onClickCapture", "onContextMenuCapture", "onDoubleClickCapture",
	"onDragCapture", "onDragEndCapture", "onDragEnterCapture", "onDragExitCapture", "onDragLeaveCapture",
	"onDragOverCapture", "onDragStartCapture", "onDropCapture", "onMouseDown", "onMouseDownCapture",
	"onMouseMoveCapture", "onMouseOutCapture", "onMouseOverCapture", "onMouseUpCapture",
	"autoPictureInPicture", "controlsList", "disablePictureInPicture", "disableRemotePlayback",
	"popoverTarget", "popoverTargetAction",
}

// domPropertiesIgnoreCase is the canonical casing for names whose React form
// differs only in letter case from the HTML form. normalizeAttributeCase uses
// this set to treat both forms identically.
var domPropertiesIgnoreCase = []string{
	"charset", "allowFullScreen", "webkitAllowFullScreen", "mozAllowFullScreen",
	"webkitDirectory", "popoverTarget", "popoverTargetAction",
}

// ariaProperties is the exhaustive set of standard aria-* attributes the rule
// recognizes as always valid. Mirrors upstream's ARIA_PROPERTIES.
var ariaProperties = map[string]bool{
	"aria-atomic": true, "aria-braillelabel": true, "aria-brailleroledescription": true, "aria-busy": true, "aria-controls": true, "aria-current": true,
	"aria-describedby": true, "aria-description": true, "aria-details": true,
	"aria-disabled": true, "aria-dropeffect": true, "aria-errormessage": true, "aria-flowto": true, "aria-grabbed": true, "aria-haspopup": true,
	"aria-hidden": true, "aria-invalid": true, "aria-keyshortcuts": true, "aria-label": true, "aria-labelledby": true, "aria-live": true,
	"aria-owns": true, "aria-relevant": true, "aria-roledescription": true,
	"aria-autocomplete": true, "aria-checked": true, "aria-expanded": true, "aria-level": true, "aria-modal": true, "aria-multiline": true, "aria-multiselectable": true,
	"aria-orientation": true, "aria-placeholder": true, "aria-pressed": true, "aria-readonly": true, "aria-required": true, "aria-selected": true,
	"aria-sort": true, "aria-valuemax": true, "aria-valuemin": true, "aria-valuenow": true, "aria-valuetext": true,
	"aria-activedescendant": true, "aria-colcount": true, "aria-colindex": true, "aria-colindextext": true, "aria-colspan": true,
	"aria-posinset": true, "aria-rowcount": true, "aria-rowindex": true, "aria-rowindextext": true, "aria-rowspan": true, "aria-setsize": true,
}

// reactOnProps is the set of pointer-event handlers added in React 16.4.
var reactOnProps = []string{
	"onGotPointerCapture",
	"onGotPointerCaptureCapture",
	"onLostPointerCapture",
	"onLostPointerCapture",
	"onLostPointerCaptureCapture",
	"onPointerCancel",
	"onPointerCancelCapture",
	"onPointerDown",
	"onPointerDownCapture",
	"onPointerEnter",
	"onPointerEnterCapture",
	"onPointerLeave",
	"onPointerLeaveCapture",
	"onPointerMove",
	"onPointerMoveCapture",
	"onPointerOut",
	"onPointerOutCapture",
	"onPointerOver",
	"onPointerOverCapture",
	"onPointerUp",
	"onPointerUpCapture",
}

// getDOMPropertyNames returns the version-dependent set of recognized DOM
// property names. Older React versions (< 16.1) recognize `allowTransparency`;
// React >= 16.4 adds pointer-event handlers; React >= 19 adds `precedence`.
func getDOMPropertyNames(settings map[string]interface{}) []string {
	names := make([]string, 0, len(domPropertyNamesTwoWords)+len(domPropertyNamesOneWord)+len(reactOnProps)+2)
	names = append(names, domPropertyNamesTwoWords...)
	names = append(names, domPropertyNamesOneWord...)

	// React < 16.1 still accepts `allowTransparency`; later versions drop it
	// (see facebook/react#10823).
	if reactutil.ReactVersionLessThan(settings, 16, 1, 0) {
		names = append(names, "allowTransparency")
		return names
	}
	// Pointer events arrived in React 16.4.
	if !reactutil.ReactVersionLessThan(settings, 16, 4, 0) {
		names = append(names, reactOnProps...)
		// `precedence` arrived in React 19 (stylesheet support).
		if !reactutil.ReactVersionLessThan(settings, 19, 0, 0) {
			names = append(names, "precedence")
		}
	}
	return names
}

// normalizeAttributeCase returns the canonical casing for `name` when its
// lowercased form matches a known case-insensitive entry, otherwise returns
// `name` unchanged. This lets `charset` / `charSet` / `CHARSET` compare equal.
func normalizeAttributeCase(name string) string {
	lower := strings.ToLower(name)
	for _, canonical := range domPropertiesIgnoreCase {
		if strings.ToLower(canonical) == lower {
			return canonical
		}
	}
	return name
}

// isValidDataAttribute reports whether `name` is a `data-*` attribute that
// React will accept: starts with `data-`, does not start with `data-xml`
// (case-insensitive), and does not contain a colon.
func isValidDataAttribute(name string) bool {
	if !strings.HasPrefix(strings.ToLower(name), "data-") {
		return false
	}
	// `data-xml*` is reserved and rejected by React.
	if len(name) >= 8 && strings.EqualFold(name[:8], "data-xml") {
		return false
	}
	// Colons split the attribute into a namespace; `data-*` must not contain
	// one (matches upstream `/^data-[^:]*$/`).
	return !strings.Contains(name[5:], ":")
}

// hasUpperCaseCharacter reports whether `name` contains any letter that
// differs from its lowercased form.
func hasUpperCaseCharacter(name string) bool {
	return strings.ToLower(name) != name
}

// getStandardName looks up `name` (case-insensitively, after the
// DOM_ATTRIBUTE_NAMES / SVGDOM_ATTRIBUTE_NAMES direct-maps) in the known React
// DOM property list. Returns the canonical property name or "" when no close
// match exists.
func getStandardName(name string, settings map[string]interface{}) string {
	if v, ok := domAttributeNames[name]; ok {
		return v
	}
	if v, ok := svgDomAttributeNames[name]; ok {
		return v
	}
	lower := strings.ToLower(name)
	for _, candidate := range getDOMPropertyNames(settings) {
		if strings.ToLower(candidate) == lower {
			return candidate
		}
	}
	return ""
}

// getAttributeNameText returns the display text of a JsxAttribute's name.
// Delegates to reactutil.GetJsxPropName for Identifier / JsxNamespacedName,
// which is the shape upstream `getText(context, node.name)` produces here
// (upstream's `node.name` is only one of those two for a valid JsxAttribute).
// Falls back to the raw source range for any unexpected shape so callers can
// still report a diagnostic instead of silently dropping it.
func getAttributeNameText(sf *ast.SourceFile, attr *ast.Node) string {
	if name := reactutil.GetJsxPropName(attr); name != "" && name != "spread" {
		return name
	}
	nameNode := attr.AsJsxAttribute().Name()
	if nameNode == nil {
		return ""
	}
	return utils.TrimmedNodeText(sf, nameNode)
}

// getJsxTagLocalName returns the text of a JSX tag name as a string for the
// purpose of this rule:
//   - Identifier                 → `<div>`        → "div"
//   - JsxNamespacedName          → `<a:b>`        → the localname ("b")
//
// Other shapes (PropertyAccessExpression, ThisKeyword) return "". Mirrors
// upstream's `childNode.parent.name.name` — for namespaced tags that's the
// JSXIdentifier inside the JSXNamespacedName (i.e. the local name).
//
// NOTE: this intentionally differs from reactutil.GetJsxElementTypeString,
// which returns "ns:name" for JsxNamespacedName — upstream eslint-plugin-react
// does not treat namespaced tags as a composite name in this rule.
func getJsxTagLocalName(tagName *ast.Node) string {
	if tagName == nil {
		return ""
	}
	switch tagName.Kind {
	case ast.KindIdentifier:
		return tagName.AsIdentifier().Text
	case ast.KindJsxNamespacedName:
		ns := tagName.AsJsxNamespacedName()
		if ns.Name() == nil || ns.Name().Kind != ast.KindIdentifier {
			return ""
		}
		return ns.Name().AsIdentifier().Text
	}
	return ""
}

// tagNameHasDot reports whether the JsxAttribute's parent tag uses the
// PropertyAccessExpression shape (e.g. `<Foo.Bar>`). Matches upstream's
// `node.parent.name.type === 'JSXMemberExpression'` check.
func tagNameHasDot(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	tagName := reactutil.GetJsxTagName(parent)
	if tagName == nil {
		return false
	}
	return tagName.Kind == ast.KindPropertyAccessExpression
}

// matchesTagConvention reports whether `name` looks like a valid HTML tag
// identifier: starts with a lowercase ASCII letter and contains no hyphen.
// Mirrors upstream's `/^[a-z][^-]*$/`.
func matchesTagConvention(name string) bool {
	if name == "" {
		return false
	}
	if name[0] < 'a' || name[0] > 'z' {
		return false
	}
	return !strings.Contains(name, "-")
}

// hasIsAttribute reports whether the element declares an `is` attribute
// (custom-elements extended built-in marker). Only JsxAttribute (never
// JsxSpreadAttribute) with an Identifier name are considered, matching
// upstream's `attrNode.type === 'JSXAttribute' && attrNode.name.type ===
// 'JSXIdentifier' && attrNode.name.name === 'is'`. The value is irrelevant.
func hasIsAttribute(parent *ast.Node) bool {
	for _, attr := range reactutil.GetJsxElementAttributes(parent) {
		if !ast.IsJsxAttribute(attr) {
			continue
		}
		nameNode := attr.AsJsxAttribute().Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			continue
		}
		if nameNode.AsIdentifier().Text == "is" {
			return true
		}
	}
	return false
}

// isValidHTMLTagInJSX reports whether the JsxAttribute's parent element is an
// HTML/DOM tag that should be subject to attribute validation — i.e. the tag
// name matches the lowercase-no-hyphen convention AND the element does not
// carry an `is="..."` attribute (which would make it a customized built-in).
func isValidHTMLTagInJSX(parent *ast.Node) bool {
	tagName := getJsxTagLocalName(reactutil.GetJsxTagName(parent))
	if !matchesTagConvention(tagName) {
		return false
	}
	return !hasIsAttribute(parent)
}

// runRule is the rule entrypoint. Registers a single JsxAttribute listener
// and walks the decision tree from upstream's `JSXAttribute(node)` handler.
func runRule(ctx rule.RuleContext, options any) rule.RuleListeners {
	ignore := map[string]bool{}
	requireDataLowercase := false
	if optsMap := utils.GetOptionsMap(options); optsMap != nil {
		if raw, ok := optsMap["ignore"].([]interface{}); ok {
			for _, v := range raw {
				if s, ok := v.(string); ok {
					ignore[s] = true
				}
			}
		}
		if v, ok := optsMap["requireDataLowercase"].(bool); ok {
			requireDataLowercase = v
		}
	}

	return rule.RuleListeners{
		ast.KindJsxAttribute: func(node *ast.Node) {
			attr := node.AsJsxAttribute()
			parent := reactutil.GetJsxParentElement(node)
			if parent == nil {
				return
			}

			actualName := getAttributeNameText(ctx.SourceFile, node)
			if actualName == "" {
				return
			}
			if ignore[actualName] {
				return
			}
			name := normalizeAttributeCase(actualName)

			// `<Foo.Bar foo />` — React components with a dotted tag name are
			// always skipped; their props are user-defined.
			if tagNameHasDot(parent) {
				return
			}

			if isValidDataAttribute(name) {
				if requireDataLowercase && hasUpperCaseCharacter(name) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id: "dataLowercaseRequired",
						Description: "React does not recognize data-* props with uppercase characters on a DOM element. Found '" +
							actualName + "', use '" + strings.ToLower(actualName) + "' instead",
					})
				}
				return
			}

			if ariaProperties[name] {
				return
			}

			tagName := getJsxTagLocalName(reactutil.GetJsxTagName(parent))
			// fbt / fbs JSX is an internal-i18n construct whose attribute set
			// upstream deliberately leaves unchecked.
			if tagName == "fbt" || tagName == "fbs" {
				return
			}
			if !isValidHTMLTagInJSX(parent) {
				return
			}

			// Element-specific attribute list: `crossOrigin`, `download`, etc.
			if allowed, ok := attributeTagsMap[name]; ok {
				if !slices.Contains(allowed, tagName) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id: "invalidPropOnTag",
						Description: "Invalid property '" + actualName + "' found on tag '" +
							tagName + "', but it is only allowed on: " + strings.Join(allowed, ", "),
					})
				}
				return
			}

			standardName := getStandardName(name, ctx.Settings)
			if standardName != "" && standardName == name {
				return
			}
			if standardName != "" && standardName != name {
				// Close match: report with suggestion + autofix that rewrites
				// the attribute name node directly.
				ctx.ReportNodeWithFixes(node, rule.RuleMessage{
					Id:          "unknownPropWithStandardName",
					Description: "Unknown property '" + actualName + "' found, use '" + standardName + "' instead",
				}, rule.RuleFixReplace(ctx.SourceFile, attr.Name(), standardName))
				return
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unknownProp",
				Description: "Unknown property '" + actualName + "' found",
			})
		},
	}
}
