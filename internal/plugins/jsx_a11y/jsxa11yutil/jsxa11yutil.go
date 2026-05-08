// Package jsxa11yutil contains shared helpers for eslint-plugin-jsx-a11y rule
// ports. The functions here mirror jsx-ast-utils / jsx-a11y/util semantics that
// are common across many a11y rules: case-insensitive attribute lookup,
// "is the literal value extractable" predicates, polymorphic / componentMap
// element-type resolution, and presentation-role / accessible-child checks.
package jsxa11yutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// skipTransparent unwraps parentheses + TS assertion wrappers (`as`, `!`,
// `<T>`, `satisfies`). Upstream's jsx-ast-utils explicitly walks past
// `TSAsExpression` and TSNonNullExpression / TSSatisfies wrappers when
// extracting prop values; we mirror that with the standard rslint helper.
const skipTransparent = ast.OEKParentheses | ast.OEKAssertions

// FindAttributeByName returns the first JsxAttribute whose name matches `name`
// case-insensitively, mirroring jsx-ast-utils' `getProp` with its default
// `{ ignoreCase: true }`.
//
// JsxSpreadAttribute handling: when the spread argument is an ObjectLiteral,
// the property inside the literal is matched (matches upstream behavior for
// `{...{ alt: "x" }}`). Both PropertyAssignment (`alt: "x"`) and
// ShorthandPropertyAssignment (`alt`) shapes are supported — upstream's
// `property.type === 'Property'` covers both because ESTree unifies them
// under a single Property type with a `shorthand` flag.
//
// Spread of a non-literal (`{...this.props}`) is opaque — upstream returns
// undefined for these and we follow suit.
//
// Like upstream, the "key" prop is excluded from spread expansion.
//
// Returned node:
//   - For a regular JsxAttribute → the JsxAttribute node
//   - For a literal-spread match → the inner property/shorthand node
//
// Callers should branch on `.Kind` if they need to access the initializer; use
// AttributeInitializer below to abstract that.
func FindAttributeByName(attrs []*ast.Node, name string) *ast.Node {
	for _, attr := range attrs {
		switch attr.Kind {
		case ast.KindJsxAttribute:
			if strings.EqualFold(reactutil.GetJsxPropName(attr), name) {
				return attr
			}
		case ast.KindJsxSpreadAttribute:
			if strings.EqualFold(name, "key") {
				continue
			}
			spread := attr.AsJsxSpreadAttribute()
			if spread.Expression == nil {
				continue
			}
			expr := ast.SkipOuterExpressions(spread.Expression, skipTransparent)
			if expr.Kind != ast.KindObjectLiteralExpression {
				continue
			}
			obj := expr.AsObjectLiteralExpression()
			if obj.Properties == nil {
				continue
			}
			for _, prop := range obj.Properties.Nodes {
				var keyNode *ast.Node
				switch prop.Kind {
				case ast.KindPropertyAssignment:
					keyNode = prop.AsPropertyAssignment().Name()
				case ast.KindShorthandPropertyAssignment:
					keyNode = prop.AsShorthandPropertyAssignment().Name()
				default:
					continue
				}
				if keyNode == nil || keyNode.Kind != ast.KindIdentifier {
					continue
				}
				if strings.EqualFold(keyNode.AsIdentifier().Text, name) {
					return prop
				}
			}
		}
	}
	return nil
}

// AttributeInitializer returns the value-bearing child of an attribute-like
// node returned by FindAttributeByName. For a JsxAttribute this is its
// `Initializer` (StringLiteral or JsxExpression). For a PropertyAssignment
// inside a spread-object, it's the property's initializer. For a
// ShorthandPropertyAssignment (`{...{alt}}`), the value is the same Identifier
// node as the key — we return it so downstream extractors see "an identifier
// named alt", matching upstream's `propertyToJSXAttribute` synthesis where the
// shorthand value is the bound identifier. Returns nil for the boolean
// attribute form (`<img alt />`) where there is no initializer.
func AttributeInitializer(attr *ast.Node) *ast.Node {
	if attr == nil {
		return nil
	}
	switch attr.Kind {
	case ast.KindJsxAttribute:
		return attr.AsJsxAttribute().Initializer
	case ast.KindPropertyAssignment:
		return attr.AsPropertyAssignment().Initializer
	case ast.KindShorthandPropertyAssignment:
		return attr.AsShorthandPropertyAssignment().Name()
	}
	return nil
}

// AttributeIsBooleanForm reports the `<img alt />` form — a JsxAttribute with
// no initializer at all. PropertyAssignment shapes (from a spread-object
// match) always carry an initializer in legal source, so this returns false.
func AttributeIsBooleanForm(attr *ast.Node) bool {
	if attr == nil {
		return false
	}
	if attr.Kind != ast.KindJsxAttribute {
		return false
	}
	return attr.AsJsxAttribute().Initializer == nil
}

// LiteralStringValue returns the literal string value carried by an attribute
// initializer:
//
//   - direct StringLiteral (`attr="x"`)
//   - JsxExpression containing a StringLiteral (`attr={"x"}`)
//   - JsxExpression containing a NoSubstitutionTemplateLiteral (`attr={`x`}`)
//   - direct NoSubstitutionTemplateLiteral inside a property assignment
//   - PropertyAssignment / ShorthandPropertyAssignment from a literal-spread
//     match (the value is the bare expression — not wrapped in JsxExpression)
//
// Parentheses and TS assertion wrappers (`as`, `!`, `<T>`, `satisfies`) are
// transparently unwrapped on every layer — `attr={("x" as string)}` extracts
// "x", matching jsx-ast-utils' `extract`/`extractLiteral` walk past
// `TSAsExpression` / `TSNonNullExpression`.
//
// Returns ("", false) for every other shape — including JsxExpression with
// Identifier (`{undefined}`, `{someVar}`), TemplateExpression with
// substitutions, etc. This mirrors jsx-ast-utils' `getLiteralPropValue` for
// the string-typed cases the alt-text / role / type / title checks rely on.
func LiteralStringValue(attr *ast.Node) (string, bool) {
	inner := attributeInnerExpression(attr)
	if inner == nil {
		return "", false
	}
	switch inner.Kind {
	case ast.KindStringLiteral:
		return inner.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return inner.AsNoSubstitutionTemplateLiteral().Text, true
	}
	return "", false
}

// attributeInnerExpression returns the unwrapped value expression of an
// attribute, normalizing across:
//   - JsxAttribute with direct StringLiteral / Template (returns the literal)
//   - JsxAttribute with `{ … }` JsxExpression (returns the inner expression)
//   - PropertyAssignment from spread (returns the initializer expression)
//   - ShorthandPropertyAssignment from spread (returns the bound identifier)
//
// Parentheses + TS assertion wrappers are stripped on the result, so callers
// can pattern-match `.Kind` directly against semantic kinds (StringLiteral,
// Identifier, BinaryExpression, …) without re-walking wrappers.
//
// Returns nil for the boolean attribute form (`<img alt />`) and for
// `{ /* empty */ }` JsxExpression containers.
func attributeInnerExpression(attr *ast.Node) *ast.Node {
	init := AttributeInitializer(attr)
	if init == nil {
		return nil
	}
	if init.Kind == ast.KindJsxExpression {
		expr := init.AsJsxExpression().Expression
		if expr == nil {
			return nil
		}
		return ast.SkipOuterExpressions(expr, skipTransparent)
	}
	return ast.SkipOuterExpressions(init, skipTransparent)
}

// AttributeIsExplicitUndefined reports whether the attribute value is an
// explicit `undefined` reference — including TS-wrapped variants like
// `{undefined as any}` and `{(undefined)!}`, and shorthand-spread forms
// `{...{alt: undefined}}` where the initializer isn't wrapped in a
// JsxExpression. Mirrors jsx-ast-utils' Identifier extractor where the
// `undefined` reserved name evaluates to the actual `undefined` value,
// combined with its TSAsExpression unwrap loop. Used by ariaLabelHasValue /
// alt validity to distinguish `<img alt={undefined} />` from
// `<img alt={someVar} />`.
func AttributeIsExplicitUndefined(attr *ast.Node) bool {
	return utils.IsUndefinedIdentifier(attributeInnerExpression(attr))
}

// AltAttributeIsValid encodes the alt-text validity rule for a present `alt`
// (or area / input alt) attribute:
//
//	(altValue && !isNullValued) || altValue === ""
//
// where `isNullValued` is the boolean-attribute form (`<img alt />`).
//
// Because alt-text never inspects the alt value beyond the truthy / empty
// distinction, we don't need a full jsx-ast-utils value extractor here — the
// AST shape alone is enough:
//
//   - boolean form (`<img alt />`)                       → invalid
//   - empty string literal (the empty-string forms)      → valid (decorative)
//   - string literal non-empty                           → valid
//   - JsxExpression with Identifier `undefined`           → invalid
//   - JsxExpression with a literal `false` / `null`       → invalid
//   - JsxExpression with `false || false`-style constant-falsy LogicalExpression → invalid
//   - any other JsxExpression (Identifier, CallExpression, MemberExpression,
//     ArrowFunction, TemplateExpression with substitutions, ConditionalExpression,
//     non-zero NumericLiteral, BigIntLiteral, …) → valid (potentially truthy)
//
// "Valid" here means the rule should NOT report `altValueError`. A nil
// attribute is considered invalid because the caller is supposed to check
// "altProp === undefined" before calling this.
//
// Implementation: the upstream check is literally
//
//	(altValue && !isNullValued) || altValue === ''
//
// where altValue is jsx-ast-utils' `getPropValue(altProp)`. Three things
// follow:
//
//  1. The boolean attribute form (`<img alt />`) has altValue === true and
//     isNullValued === true, so the LHS is false and the RHS is `true === ''`
//     (false) → invalid.
//  2. An empty-string altValue is valid via the RHS regardless of how it's
//     produced — `alt=""`, `alt={""}`, `alt={"" && x}`, `alt={x && ""}` all
//     reach `=== ''` and pass. truthy/falsy heuristics are NOT enough; we
//     must compute the actual static value via [staticEval].
//  3. Anything else uses the LHS — truthy after JS coercion.
func AltAttributeIsValid(attr *ast.Node) bool {
	if attr == nil {
		return false
	}
	if AttributeIsBooleanForm(attr) {
		return false
	}
	inner := attributeInnerExpression(attr)
	if inner == nil {
		// `<img alt={} />` — empty JsxExpression. tsgo synthesizes this only
		// for malformed input; treat it as "not truthy" → invalid.
		return false
	}
	// Run through staticEval — note this also normalizes string-literal
	// "true" / "false" to booleans, so `<img alt="false" />` correctly fails
	// (matches jsx-ast-utils' Literal extractor).
	v := staticEval(inner)
	if jsValueIsExactlyEmptyString(v) {
		return true // matches the `altValue === ''` branch
	}
	return jsTruthy_(v)
}

// PropStaticStringValue mirrors `getPropValue(prop)` for callers that need
// to compare against a specific string. Returns ("", false) if the prop's
// static value isn't a string (e.g. boolean, undefined, unknown, function).
//
// Use this for upstream call sites that pass `getPropValue(...) === "x"` —
// e.g. `input[type="image"]`'s type check, where upstream applies real JS
// `===` semantics (case-sensitive, against the coerced static value).
//
// Differs from LiteralStringValue: this one runs the full staticEval, so
// `type={"image" + ""}` (BinaryExpression) and `type={cond ? "image" :
// "text"}` (ConditionalExpression) get statically resolved when possible.
func PropStaticStringValue(attr *ast.Node) (string, bool) {
	inner := attributeInnerExpression(attr)
	if inner == nil {
		return "", false
	}
	v := staticEval(inner)
	if v.Kind == jvString {
		return v.Str, true
	}
	return "", false
}

// LiteralPropTruthy mirrors `!!getLiteralPropValue(prop)`. Returns true when
// the attribute's value is a JS-truthy *literal*. Crucial differences from
// PropStaticStringValue / staticEval-based truthiness:
//
//   - Identifier (other than `undefined`) → null → falsy. Upstream
//     LITERAL_TYPES.Identifier maps non-undefined identifiers to null
//     intentionally — runtime variables aren't statically literal.
//   - null literal → "null" string → truthy (special upstream behavior!).
//   - Most expression kinds (Call, Member, Conditional, Logical, Binary,
//     Unary, etc.) → null → falsy under LITERAL_TYPES noop.
//   - String literal goes through the same `"true"`/`"false"` boolean
//     coercion as PropValue.
//
// Used for `<object>` title and `isPresentationRole` role checks where
// upstream uses `getLiteralPropValue`.
func LiteralPropTruthy(attr *ast.Node) bool {
	if attr == nil {
		return false
	}
	if AttributeIsBooleanForm(attr) {
		// `<… title />` — extractValue's null-attribute-value path returns
		// `true`, then `!!true` = true.
		return true
	}
	inner := attributeInnerExpression(attr)
	if inner == nil {
		return false
	}
	v := literalPropValue(inner)
	return jsTruthy_(v)
}

// AriaLabelHasValue mirrors upstream's `ariaLabelHasValue`:
//
//	const value = getPropValue(prop);
//	if (value === undefined) return false;
//	if (typeof value === 'string' && value.length === 0) return false;
//	return true;
//
// Note that this is NOT a simple "truthy" check — `null`, `false`, `0` all
// return true here even though they are falsy in JS, because they are
// neither `undefined` nor empty string. This matters: e.g. `aria-label={null}`
// counts as "has value" per upstream.
//
// Returns false when `attr` is nil — callers should test for the attribute's
// existence first; a nil attr is "no value" by definition.
func AriaLabelHasValue(attr *ast.Node) bool {
	if attr == nil {
		return false
	}
	if AttributeIsBooleanForm(attr) {
		// `<img aria-label />` — upstream's extractValue maps the null
		// initializer to `true`; not undefined, not string of length 0,
		// therefore "has value".
		return true
	}
	// Direct StringLiteral attribute init — `aria-label=""` or
	// `aria-label="x"`. Empty string → no value, anything else → has value.
	if init := AttributeInitializer(attr); init != nil && init.Kind == ast.KindStringLiteral {
		return init.AsStringLiteral().Text != ""
	}
	inner := attributeInnerExpression(attr)
	if inner == nil {
		// `<img aria-label={} />` — empty JsxExpression. tsgo synthesizes
		// this only for malformed source; treat as no-value to match the
		// stricter interpretation (upstream's getPropValue would also fail
		// here).
		return false
	}
	v := staticEval(inner)
	if jsValueIsExactlyUndefined(v) {
		return false
	}
	if v.Kind == jvString && v.Str == "" {
		return false
	}
	// jvUnknown defaults to "has value" because we don't know — same as
	// jsx-ast-utils returning a non-empty Identifier name string.
	return true
}

// GetElementType resolves the effective HTML element name for a JSX opening
// element, mirroring eslint-plugin-jsx-a11y's `getElementType(context)(node)`.
//
// Steps, in order:
//  1. Take the raw JSX tag-name string (`<Foo>` → "Foo", `<svg:path>` →
//     "svg:path", `<Foo.Bar>` → "Foo.Bar").
//  2. If `settings['jsx-a11y'].polymorphicPropName` is set, extract the
//     polymorphic prop's value via `getLiteralPropValue` semantics. Replace
//     rawType only when the value is truthy AND (no allow-list OR rawType
//     is in the allow-list). Non-string truthy values (number / boolean /
//     `null` literal which upstream maps to the "null" string) are
//     stringified — they typically won't match any element in
//     typesToValidate, mirroring upstream's behavior of replacing rawType
//     with a non-string then failing the Set.has() check.
//  3. If `settings['jsx-a11y'].components` is configured AND has rawType as
//     a key, replace rawType with the mapped value.
//
// Returns "" when the tag has no resolvable string form (e.g. computed
// member expressions, which aren't legal JSX anyway).
func GetElementType(node *ast.Node, settings map[string]interface{}) string {
	a11y := getJsxA11ySettings(settings)
	rawType := reactutil.GetJsxElementTypeString(node)
	polymorphicPropName, _ := a11y["polymorphicPropName"].(string)
	if polymorphicPropName != "" {
		attrs := reactutil.GetJsxElementAttributes(node)
		propAttr := FindAttributeByName(attrs, polymorphicPropName)
		if polyValue, ok := polymorphicPropValue(propAttr); ok {
			if allowList, ok := a11y["polymorphicAllowList"].([]interface{}); ok {
				for _, v := range allowList {
					if s, ok := v.(string); ok && s == rawType {
						rawType = polyValue
						break
					}
				}
			} else {
				rawType = polyValue
			}
		}
	}
	if components, ok := a11y["components"].(map[string]interface{}); ok {
		if mapped, ok := components[rawType].(string); ok {
			rawType = mapped
		}
	}
	return rawType
}

// polymorphicPropValue extracts the polymorphic-prop value via upstream's
// `getLiteralPropValue` semantics. Returns the truthy stringified value and
// `true` when the prop is set and resolves to a truthy literal; `("", false)`
// otherwise. For non-string truthy literals (number, bool, null), returns
// their JS String() coercion — these stringified forms ("123", "true",
// "null") generally aren't valid HTML element names and won't match
// typesToValidate, mirroring upstream's behavior where rawType becomes a
// non-string and the Set.has() check fails.
func polymorphicPropValue(propAttr *ast.Node) (string, bool) {
	if propAttr == nil {
		return "", false
	}
	if AttributeIsBooleanForm(propAttr) {
		// `<Foo as />` — upstream's extractValue maps null-attribute-value
		// to boolean true. polymorphicProp truthy → rawType = true. The
		// Set.has(true) check downstream fails, so this skips alt-text
		// entirely. Mirror by returning the string "true".
		return "true", true
	}
	inner := attributeInnerExpression(propAttr)
	if inner == nil {
		return "", false
	}
	v := literalPropValue(inner)
	if !jsTruthy_(v) {
		return "", false
	}
	if v.Kind == jvString {
		return v.Str, true
	}
	return jsToString(v), true
}

// IsPresentationRole mirrors `isPresentationRole`: the element has an
// explicit `role` attribute whose literal value is "presentation" or "none".
// Non-literal expressions and absent role attributes return false.
func IsPresentationRole(attrs []*ast.Node) bool {
	roleAttr := FindAttributeByName(attrs, "role")
	if roleAttr == nil {
		return false
	}
	value, ok := LiteralStringValue(roleAttr)
	if !ok {
		return false
	}
	return value == "presentation" || value == "none"
}

// HasAccessibleChild reports whether a JSX element provides a text
// alternative for assistive technology via children or the
// `dangerouslySetInnerHTML` / `children` attribute fallback. Mirrors
// upstream's `hasAccessibleChild(node.parent, elementType)`.
//
// `node` is the JSX element root — either a JsxElement (which carries both
// children and an opening element) or a JsxSelfClosingElement (which has
// only attributes). For other shapes (or nil), returns false.
//
// The check returns true when ANY of these hold:
//   - a non-empty JsxText / JsxTextAllWhiteSpaces child (`<object>Foo</object>`)
//   - a string-literal child (matches upstream's `case 'Literal'`)
//   - a JsxElement / JsxSelfClosingElement child whose tag is not hidden from
//     screen readers (`aria-hidden`, `<input type="hidden">`)
//   - a JsxExpression child whose payload is anything other than `{undefined}`
//   - the opening element declares a `dangerouslySetInnerHTML` or `children`
//     attribute (matches upstream's `hasAnyProp` fallback)
//
// JsxFragment children are NOT counted as accessible — upstream's switch has
// no `case 'JSXFragment'`, so they fall to the `default: return false`
// branch. `<object><>x</></object>` is therefore reported invalid even
// though the fragment contains text, matching upstream.
//
// `getElementType` is the per-context resolver — pass a closure that calls
// GetElementType with `ctx.Settings` already bound. Mirrors upstream's
// `hasAccessibleChild(node, elementType)` curry shape.
func HasAccessibleChild(node *ast.Node, getElementType func(*ast.Node) string) bool {
	if node == nil {
		return false
	}
	var children []*ast.Node
	var openingAttrs []*ast.Node
	switch node.Kind {
	case ast.KindJsxElement:
		jsx := node.AsJsxElement()
		if jsx.Children != nil {
			children = jsx.Children.Nodes
		}
		openingAttrs = reactutil.GetJsxElementAttributes(jsx.OpeningElement)
	case ast.KindJsxSelfClosingElement:
		// No children possible; the opening attributes are on the node
		// itself. Upstream's hasAccessibleChild walks `JSXElement.children`
		// (empty for self-closing) then falls back to
		// `node.openingElement.attributes` — same effect.
		openingAttrs = reactutil.GetJsxElementAttributes(node)
	default:
		return false
	}
	for _, child := range children {
		switch child.Kind {
		case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces:
			// Upstream's hasAccessibleChild uses `!!child.value` — any
			// non-empty text counts. tsgo splits JsxText into two kinds
			// (regular vs whitespace-only); both carry the raw text on
			// `.Text` and we check both for parity.
			if child.AsJsxText().Text != "" {
				return true
			}
		case ast.KindStringLiteral:
			// Bare string literals as children are uncommon, but mirror
			// upstream's `case 'Literal'`.
			if child.AsStringLiteral().Text != "" {
				return true
			}
		case ast.KindJsxElement:
			// Inspect the OPENING element of the paired JsxElement — that's
			// where the tag name and `aria-hidden` / `type="hidden"` live.
			// Matches upstream's `elementType(child.openingElement)` and
			// `child.openingElement.attributes`.
			opening := child.AsJsxElement().OpeningElement
			if opening != nil && !isHiddenFromScreenReader(opening, getElementType) {
				return true
			}
		case ast.KindJsxSelfClosingElement:
			if !isHiddenFromScreenReader(child, getElementType) {
				return true
			}
		case ast.KindJsxExpression:
			expr := child.AsJsxExpression().Expression
			if expr == nil {
				// `{}` empty container — matches upstream `default: false`.
				continue
			}
			inner := ast.SkipOuterExpressions(expr, skipTransparent)
			if utils.IsUndefinedIdentifier(inner) {
				continue
			}
			return true
		}
	}
	// Fallback: opening element declares dangerouslySetInnerHTML or children.
	for _, name := range []string{"dangerouslySetInnerHTML", "children"} {
		if FindAttributeByName(openingAttrs, name) != nil {
			return true
		}
	}
	return false
}

// isHiddenFromScreenReader mirrors upstream's `isHiddenFromScreenReader`:
//
//	if (type.toUpperCase() === 'INPUT') {
//	  const hidden = getLiteralPropValue(getProp(attrs, 'type'));
//	  if (hidden && hidden.toUpperCase() === 'HIDDEN') return true;
//	}
//	const ariaHidden = getPropValue(getProp(attrs, 'aria-hidden'));
//	return ariaHidden === true;
//
// Note the asymmetry: `type` uses getLiteralPropValue (literal-only),
// `aria-hidden` uses getPropValue (full static eval) and compares with
// JS `===` to boolean true. We mirror both — staticEval/literalPropValue
// handle the wrapper unwrapping and "true"/"false" string normalization
// transparently, so e.g. `aria-hidden="true"` and `aria-hidden={cond ? true : false}`
// both classify correctly.
func isHiddenFromScreenReader(child *ast.Node, getElementType func(*ast.Node) string) bool {
	tag := strings.ToUpper(getElementType(child))
	attrs := reactutil.GetJsxElementAttributes(child)
	if tag == "INPUT" {
		typeAttr := FindAttributeByName(attrs, "type")
		if typeAttr != nil {
			if inner := attributeInnerExpression(typeAttr); inner != nil {
				v := literalPropValue(inner)
				if v.Kind == jvString && strings.EqualFold(v.Str, "hidden") {
					return true
				}
			}
		}
	}
	ariaHidden := FindAttributeByName(attrs, "aria-hidden")
	if ariaHidden == nil {
		return false
	}
	if AttributeIsBooleanForm(ariaHidden) {
		// Boolean form maps to extractValue's null-attr-value → true; true === true → hidden.
		return true
	}
	inner := attributeInnerExpression(ariaHidden)
	if inner == nil {
		return false
	}
	// `getPropValue(...) === true` — only the actual boolean true matches.
	// `<div aria-hidden="true">` works because the Literal extractor maps the
	// case-insensitive string "true" to boolean true (jsxAstUtilsLiteralCoerce).
	v := staticEval(inner)
	return v.Kind == jvBool && v.Bool
}

func getJsxA11ySettings(settings map[string]interface{}) map[string]interface{} {
	if settings == nil {
		return nil
	}
	m, _ := settings["jsx-a11y"].(map[string]interface{})
	return m
}
