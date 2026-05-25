package jsxa11yutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
)

// IsContentEditable mirrors upstream `isContentEditable(_, attributes)`
// (eslint-plugin-jsx-a11y/src/util/isContentEditable.js) verbatim:
//
//	const prop = getProp(attributes, 'contentEditable');
//	return prop?.value?.raw === '"true"';
//
// The upstream check is deliberately restrictive: it gates on the RAW
// source text of the value (including surrounding quotes), so ONLY the bare
// `contentEditable="true"` form matches. Specifically:
//
//   - `<X contentEditable />`           — no `.value` → no match
//   - `<X contentEditable="true" />`    — raw `"true"` → MATCH
//   - `<X contentEditable="false" />`   — raw `"false"` → no match
//   - `<X contentEditable={true} />`    — raw is the JsxExpression text, not `"true"` → no match
//   - `<X contentEditable={"true"} />`  — value is a JsxExpression, not a string-literal attribute init → no match
//   - `<X contentEditable="&#116;rue" />` — raw `"&#116;rue"` (entity-encoded) → no match
//
// We mirror this by reading the JsxAttribute's StringLiteral initializer's
// `.Text` directly — tsgo preserves the unprocessed (entity-encoded)
// source text on StringLiteral.Text for JSX attributes, which is exactly
// what upstream's `.raw` (minus the surrounding quotes) carries. Entity
// decoding is intentionally NOT applied — upstream's raw comparison
// rejects entity-encoded forms.
//
// getProp is case-insensitive by default, so `contentEditable`,
// `contenteditable`, and `CONTENTEDITABLE` all match. JsxSpreadAttribute is
// walked by FindAttributeByName for literal-object spreads — but a
// PropertyAssignment / ShorthandPropertyAssignment inside a spread has no
// `.value.raw` equivalent (the value is the bare initializer expression,
// not a JSX-attribute string-literal wrapper), so spread matches fall
// through to "no match" here.
func IsContentEditable(attrs []*ast.Node) bool {
	for _, attr := range attrs {
		if attr.Kind != ast.KindJsxAttribute {
			// JsxSpreadAttribute — upstream's `prop?.value?.raw` reads the
			// JSXAttribute's literal value node, which spread synthesis does
			// not produce. Skip rather than walk into spread.
			continue
		}
		if !strings.EqualFold(reactutil.GetJsxPropName(attr), "contentEditable") {
			continue
		}
		// upstream returns immediately on the FIRST matching getProp result;
		// any subsequent contentEditable attrs are ignored. Mirror by
		// returning here regardless of the value shape.
		init := attr.AsJsxAttribute().Initializer
		if init == nil || init.Kind != ast.KindStringLiteral {
			return false
		}
		// StringLiteral.Text for JSX attributes preserves the raw source
		// text (entity-encoded, no surrounding quotes). Upstream's
		// `prop.value.raw === '"true"'` is equivalent to "the text between
		// the quotes equals `true` byte-for-byte".
		return init.AsStringLiteral().Text == "true"
	}
	return false
}
