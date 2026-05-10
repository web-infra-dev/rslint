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

// skipTransparent is the wrapper mask used by `staticEval` (the
// `getPropValue` / TYPES path). Strips parentheses, type assertions
// (`as` / `<T>x` / `TypeCastExpression`), and non-null assertions (`!`),
// because upstream's jsx-ast-utils does the equivalent:
//
//   - parens are flattened by ESTree's parser; tsgo preserves them, so
//     stripping is needed for parity.
//   - `TSAsExpression` is unwrapped via the while-loop in
//     `extractValueFromExpression`.
//   - `TSNonNullExpression` has its own TYPES extractor that recurses
//     into `.expression`, equivalent to stripping.
//
// `OEKSatisfies` is INTENTIONALLY EXCLUDED. Upstream's `TYPES` table has
// no entry for `TSSatisfiesExpression`, so it falls to the
// `TYPES[type] === undefined → return null` branch. Keeping satisfies
// opaque here makes it land on `staticEval`'s default `jsNull` arm,
// matching upstream's null exactly. Used only by `staticEval`; the
// `getLiteralPropValue` (`literalPropValue`) and `getProp` paths strip
// parens only — see those callers.
const skipTransparent = ast.OEKParentheses | ast.OEKTypeAssertions | ast.OEKNonNullAssertions

// StringSliceOption coerces a JSON-decoded option value into `[]string`,
// silently dropping any non-string entries. It is the standard helper for
// the `components: string[]` / `elements: string[]` / `specialLink:
// string[]` / `<element>: string[]` shapes that appear across jsx-a11y
// rule options (anchor-has-content, anchor-is-valid, alt-text, …).
//
// Returns:
//   - `nil` when `v` is not a `[]interface{}` (absent option, wrong type).
//     Callers should treat `nil` the same as upstream's `||` fallback —
//     i.e. apply the rule's default. An EXPLICIT empty array is a
//     deliberate "disable all" signal and is returned as a non-nil
//     zero-length slice; this matters for rules whose semantics differ
//     between "absent" and "explicit []".
//   - a freshly-allocated `[]string` containing only the string entries
//     of `v`, in order. Non-string entries (numbers, nested arrays,
//     objects) are dropped — upstream's options validate via JSON schema
//     so we only see well-typed values in practice; the filter is purely
//     defensive.
//
// Single source of truth for this pattern; do NOT inline-loop a fresh copy
// in a new rule.
func StringSliceOption(v interface{}) []string {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

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
			// jsx-ast-utils' getProp checks `attribute.argument.type ===
			// 'ObjectExpression'` strictly — no TS-wrapper unwrap. ESTree
			// folds parentheses at parse time so `({...})` shows up as a
			// bare ObjectExpression, but tsgo preserves the
			// ParenthesizedExpression node, so we must strip parens (and
			// only parens) to recover the same shape. TS-wrapper kinds
			// (`as`, `!`, `satisfies`) MUST stay opaque — upstream skips
			// `{...({alt: "x"})!}` and we mirror.
			expr := ast.SkipOuterExpressions(spread.Expression, ast.OEKParentheses)
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

// HasAnyJsxPropStrict mirrors jsx-ast-utils' `hasAnyProp` with default
// options (`spreadStrict: true`, `ignoreCase: true`). Returns true iff any
// of the given names is present as a DIRECT JsxAttribute (case-insensitive).
//
// Spread attributes — even literal ObjectLiteral spreads — are opaque under
// the strict default, so `<a {...{title: 'x'}} />` returns false here. This
// is the upstream-correct semantic for rules that explicitly use `hasProp`
// / `hasAnyProp` (e.g. anchor-has-content, hasAccessibleChild's fallback
// `dangerouslySetInnerHTML`/`children` lookup).
//
// Use this helper when porting upstream code that calls `hasProp` /
// `hasAnyProp`. Use FindAttributeByName instead when porting `getProp` /
// `getPropValue` / `getLiteralPropValue` call sites — those walk literal
// spreads.
func HasAnyJsxPropStrict(attrs []*ast.Node, names ...string) bool {
	for _, attr := range attrs {
		if attr.Kind != ast.KindJsxAttribute {
			// JsxSpreadAttribute is opaque under the strict default,
			// regardless of whether the spread argument is a literal
			// ObjectLiteralExpression.
			continue
		}
		propName := reactutil.GetJsxPropName(attr)
		for _, name := range names {
			if strings.EqualFold(propName, name) {
				return true
			}
		}
	}
	return false
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
// Only parentheses are stripped — TS assertion wrappers (`as`, `!`, `<T>`,
// `satisfies`) are LEFT IN PLACE so downstream extractors can decide whether
// to unwrap them. This matters because jsx-ast-utils' two extractors disagree
// on TS-wrapper handling:
//
//   - `extract()` / getPropValue: while-loops past TSAsExpression, returns
//     the inner value (so `{"x" as string}` → "x").
//   - `extractLiteral()` / getLiteralPropValue: maps TSAsExpression /
//     TSNonNullExpression / TypeCastExpression to noop → null (so
//     `{"x" as string}` → null).
//
// staticEval mirrors getPropValue and re-strips TS wrappers internally;
// literalPropValue mirrors getLiteralPropValue and DOES NOT.
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
		return ast.SkipOuterExpressions(expr, ast.OEKParentheses)
	}
	return ast.SkipOuterExpressions(init, ast.OEKParentheses)
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
//     isNullValued === true, so the LHS is false and the RHS is `true === ”`
//     (false) → invalid.
//  2. An empty-string altValue is valid via the RHS regardless of how it's
//     produced — `alt=""`, `alt={""}`, `alt={"" && x}`, `alt={x && ""}` all
//     reach `=== ”` and pass. truthy/falsy heuristics are NOT enough; we
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

// LiteralPropStringValue mirrors `getLiteralPropValue(prop)` filtered to the
// string-typed result. Returns ("", false) when the prop's literal-typed
// value isn't a string under jsx-ast-utils' LITERAL_TYPES rules:
//
//   - Boolean attribute form (`<input autocomplete />`, `<img alt />`) →
//     upstream returns boolean true, not a string → ("", false).
//   - Identifier (`<input autocomplete={x} />`, `<img alt={someAlt} />`) →
//     noop in LITERAL_TYPES → null → ("", false).
//   - LogicalExpression / ConditionalExpression / CallExpression /
//     MemberExpression / BinaryExpression — all noop in LITERAL_TYPES → null
//     → ("", false).
//   - String literals "true" / "false" coerce to booleans (NOT strings) →
//     ("", false).
//   - null literal returns the magic string "null" — upstream
//     LITERAL_TYPES.Literal special-case.
//   - TemplateExpression with substitutions becomes "head{Identifier}tail"
//     style placeholder strings — non-empty and matches upstream's
//     extractValueFromTemplateLiteral output exactly.
//
// Differs from LiteralStringValue (which only handles direct StringLiteral
// / NoSubstitutionTemplateLiteral) in two ways:
//  1. Routes through literalPropValue, which special-cases the `null`
//     literal to the string `"null"` (LITERAL_TYPES.Literal override).
//  2. Synthesizes a placeholder string for TemplateExpression with
//     substitutions (matches jsx-ast-utils' TemplateLiteral extractor).
//
// Used by rules whose upstream implementation calls `getLiteralPropValue` and
// gates on `typeof === 'string'` (autocomplete-valid, img-redundant-alt) —
// anything other than a literal-typed string makes the rule return early.
func LiteralPropStringValue(attr *ast.Node) (string, bool) {
	if attr == nil {
		return "", false
	}
	if AttributeIsBooleanForm(attr) {
		return "", false
	}
	inner := attributeInnerExpression(attr)
	if inner == nil {
		return "", false
	}
	v := literalPropValue(inner)
	if v.Kind == jvString {
		return v.Str, true
	}
	return "", false
}

// PropValueIsNullish mirrors the upstream `getPropValue(prop) == null` check
// (loose equality with `null`, true for both `null` and `undefined`). Used by
// rules that distinguish "no usable href" (absent prop, `prop={null}`,
// `prop={undefined}`) from "href provided" (anything else, including boolean
// `<a href />`, `prop={true}`, `prop={"foo"}`, `prop={someVar}`, calls,
// member access, etc.).
//
// Returns true when:
//   - `attr` is nil — `getProp` would have returned a missing prop, and
//     `getPropValue(undefined)` evaluates to `undefined`.
//   - the prop's value is an empty JsxExpression (`prop={}`) — tsgo accepts
//     this for error-recovery, but jsx-ast-utils' expressions extractor has
//     no entry for `JSXEmptyExpression` and falls through to its
//     `TYPES[type] === undefined → return null` path, which is `== null`.
//   - the prop's value statically resolves to `null` or `undefined`. This
//     covers `prop={null}`, `prop={undefined}`, and the TS-wrapped variants
//     `prop={null as any}` / `prop={undefined!}` (skipTransparent unwraps
//     parens + assertion wrappers per jsx-ast-utils' extract loop).
//   - staticEval cannot resolve the expression at all (`jvUnknown`). This
//     mirrors jsx-ast-utils returning `null` for unrecognized expression
//     types via the same fallback path. Producing this state is rare in
//     practice (most "I don't know" arms in staticEval fall through to
//     jsNull explicitly); kept here for defensive parity.
//
// Returns false for:
//   - boolean form `<a prop />` — upstream maps the null-attribute-value to
//     boolean `true`, which is `!= null`.
//   - any non-nullish static value (string, number, boolean, function,
//     truthy synthesized strings for member access / calls, etc.).
func PropValueIsNullish(attr *ast.Node) bool {
	if attr == nil {
		return true
	}
	if AttributeIsBooleanForm(attr) {
		return false
	}
	inner := attributeInnerExpression(attr)
	if inner == nil {
		// Empty `{}` JsxExpression (or expression-container holding only
		// trivia). jsx-ast-utils routes JSXEmptyExpression through its
		// "type not in TYPES" fallback → returns null. We mirror that —
		// `null != null` is false, so the prop contributes nothing to
		// `hasAnyHref` and the element correctly trips the noHref aspect.
		return true
	}
	v := staticEval(inner)
	return v.Kind == jvNull || v.Kind == jvUndef || v.Kind == jvUnknown
}

// PropValueIsTruthy mirrors `!!getPropValue(prop)` — the truthy classification
// of upstream's full TYPES extractor. Used by rules whose upstream form is
// `if (prop && getPropValue(prop)) ...` (e.g. no-access-key).
//
// Differs from LiteralPropTruthy (which uses the LITERAL_TYPES path):
//
//   - Identifier (other than `undefined`) here returns the identifier name as
//     a non-empty string → truthy. Under LiteralPropTruthy non-undefined
//     identifiers map to null → falsy.
//   - CallExpression / MemberExpression / Conditional / Logical / Binary
//     resolve via staticEval (truthy synthesis or genuine evaluation) here;
//     all noop → null (falsy) under LiteralPropTruthy.
//   - String literals "true" / "false" coerce to booleans on both paths;
//     `<div accessKey="false" />` is therefore falsy → not reported.
//
// Boolean attribute form (`<div accessKey />`) maps to upstream's
// null-attribute-value path → boolean `true` → truthy, matching
// jsx-ast-utils' getPropValue(`<div accessKey />`).
//
// Returns false when `attr` is nil — callers are expected to test for the
// attribute's existence first; a nil attr has no value to coerce.
func PropValueIsTruthy(attr *ast.Node) bool {
	if attr == nil {
		return false
	}
	if AttributeIsBooleanForm(attr) {
		// `<div accessKey />` — extractValue's null-attr-value path returns
		// boolean true; `!!true` == true.
		return true
	}
	inner := attributeInnerExpression(attr)
	if inner == nil {
		// Empty `{}` JsxExpression. tsgo synthesizes this only for malformed
		// source; getPropValue would route through "type not in TYPES" → null
		// → falsy.
		return false
	}
	return jsTruthy_(staticEval(inner))
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

// JsxAccessibleChildRoot returns the JSX root node to feed into
// HasAccessibleChild, mirroring upstream's `node.parent` access on a
// JSXOpeningElement.
//
// In ESTree, every JSXOpeningElement (paired or self-closing) is wrapped
// in a JSXElement that owns the children list, so upstream rules write
// `hasAccessibleChild(node.parent, …)` regardless of form. tsgo splits
// the two forms: paired `<a>x</a>` produces a KindJsxElement that
// contains a KindJsxOpeningElement, while self-closing `<a />` is a
// top-level KindJsxSelfClosingElement with no wrapper. To present the
// same "children + opening attributes" surface to HasAccessibleChild
// regardless of form, we return:
//
//   - the parent KindJsxElement when the listener fired on a
//     KindJsxOpeningElement (paired form)
//   - the node itself otherwise (self-closing form, where the node is a
//     KindJsxSelfClosingElement and HasAccessibleChild handles it via
//     its dedicated arm)
//
// Use this exactly once per listener invocation, before calling
// HasAccessibleChild. Without this normalization, a listener that fires
// on KindJsxOpeningElement would pass the opening element directly,
// which HasAccessibleChild's switch handles by falling to default-false
// — silently misreporting paired forms.
func JsxAccessibleChildRoot(node *ast.Node) *ast.Node {
	if node.Kind == ast.KindJsxOpeningElement && node.Parent != nil &&
		node.Parent.Kind == ast.KindJsxElement {
		return node.Parent
	}
	return node
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
			if opening != nil && !IsHiddenFromScreenReader(opening, getElementType) {
				return true
			}
		case ast.KindJsxSelfClosingElement:
			if !IsHiddenFromScreenReader(child, getElementType) {
				return true
			}
		case ast.KindJsxExpression:
			expr := child.AsJsxExpression().Expression
			if expr == nil {
				// `{}` and `{/* comment */}` — tsgo emits a JsxExpression
				// with no Expression. ESTree, by contrast, emits a
				// `JSXExpressionContainer` whose `expression` is a
				// `JSXEmptyExpression` node. Upstream's switch only
				// special-cases `child.expression.type === 'Identifier'`;
				// `JSXEmptyExpression` is NOT an Identifier, so it falls
				// through to the `return true` arm. Mirror that —
				// empty-container children count as accessible content.
				return true
			}
			// Upstream's switch checks `child.expression.type === 'Identifier'`
			// DIRECTLY without unwrapping. In typescript-eslint's AST, TS
			// wrappers (`as` / `!` / `satisfies` / `<T>x`) are exposed as
			// their own AST nodes whose `.type` is not 'Identifier' — so
			// they fall through to `return true` (= accessible). Parens are
			// auto-flattened by ESTree's parser, so `<a>{(undefined)}</a>`
			// is reported but `<a>{undefined as any}</a>` is not. Mirror
			// that by stripping parens only (NOT TS assertion wrappers).
			inner := ast.SkipParentheses(expr)
			if utils.IsUndefinedIdentifier(inner) {
				continue
			}
			return true
		}
	}
	// Fallback: opening element declares dangerouslySetInnerHTML or children.
	// Upstream calls `hasAnyProp(attrs, ['dangerouslySetInnerHTML', 'children'])`
	// — `hasAnyProp`'s default `spreadStrict: true` makes every spread
	// (literal or not) opaque, so we use HasAnyJsxPropStrict here. Walking
	// literal spreads via FindAttributeByName would diverge from upstream
	// for `<a {...{children: 'x'}} />`-style synthetic shapes.
	if HasAnyJsxPropStrict(openingAttrs, "dangerouslySetInnerHTML", "children") {
		return true
	}
	return false
}

// IsHiddenFromScreenReader mirrors upstream's `isHiddenFromScreenReader`:
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
//
// `child` is the JsxOpeningElement / JsxSelfClosingElement to inspect;
// `getElementType` is the per-context resolver (use a closure that calls
// GetElementType with `ctx.Settings` already bound).
func IsHiddenFromScreenReader(child *ast.Node, getElementType func(*ast.Node) string) bool {
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
