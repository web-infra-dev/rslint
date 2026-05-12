package jsxa11yutil

import (
	"math/big"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
)

// AriaPropertyDefinition describes the type and permitted values of a single
// ARIA state / property as defined by `aria-query`'s `ariaPropsMap`.
//
// Source: https://github.com/A11yance/aria-query/blob/main/src/ariaPropsMap.js
type AriaPropertyDefinition struct {
	// Type is one of: "boolean", "string", "integer", "number",
	// "tristate", "token", "tokenlist", "id", "idlist". The list is
	// closed — aria-query has no other type tags.
	Type string
	// Values is the list of permitted token values for "token" /
	// "tokenlist" types. Elements are either strings or booleans —
	// `aria-current`, `aria-haspopup`, and `aria-invalid` include the
	// JavaScript booleans `true` / `false` as valid token values, so the
	// list is heterogeneous. Empty for other types.
	//
	// Iteration order matches upstream's array order so the error-message
	// `${permittedValues}` (which hits Array.prototype.toString and joins
	// with a bare comma) renders byte-for-byte identical.
	Values []any
	// AllowUndefined mirrors upstream's `allowundefined` (lowercase!) field.
	// Set only for `aria-expanded`, `aria-grabbed`, `aria-hidden`, and
	// `aria-selected`, where ARIA explicitly permits the value `undefined`
	// as a third state distinct from absence.
	AllowUndefined bool
}

// AriaPropertyDefinitions mirrors `aria-query`'s `ariaPropsMap` entries
// keyed by lowercase canonical name. Source of truth for type / values /
// allow-undefined metadata used by jsx-a11y/aria-proptypes.
//
// AriaPropertyNames and AriaPropertySet remain the sources of truth for
// iteration order (suggestion ranking in jsx-a11y/aria-props) and existence
// checks; this map covers ONLY the value-validation metadata.
//
// Source: https://github.com/A11yance/aria-query/blob/main/src/ariaPropsMap.js
var AriaPropertyDefinitions = map[string]AriaPropertyDefinition{
	"aria-activedescendant":       {Type: "id"},
	"aria-atomic":                 {Type: "boolean"},
	"aria-autocomplete":           {Type: "token", Values: []any{"inline", "list", "both", "none"}},
	"aria-braillelabel":           {Type: "string"},
	"aria-brailleroledescription": {Type: "string"},
	"aria-busy":                   {Type: "boolean"},
	"aria-checked":                {Type: "tristate"},
	"aria-colcount":               {Type: "integer"},
	"aria-colindex":               {Type: "integer"},
	"aria-colspan":                {Type: "integer"},
	"aria-controls":               {Type: "idlist"},
	"aria-current":                {Type: "token", Values: []any{"page", "step", "location", "date", "time", true, false}},
	"aria-describedby":            {Type: "idlist"},
	"aria-description":            {Type: "string"},
	"aria-details":                {Type: "id"},
	"aria-disabled":               {Type: "boolean"},
	"aria-dropeffect":             {Type: "tokenlist", Values: []any{"copy", "execute", "link", "move", "none", "popup"}},
	"aria-errormessage":           {Type: "id"},
	"aria-expanded":               {Type: "boolean", AllowUndefined: true},
	"aria-flowto":                 {Type: "idlist"},
	"aria-grabbed":                {Type: "boolean", AllowUndefined: true},
	"aria-haspopup":               {Type: "token", Values: []any{false, true, "menu", "listbox", "tree", "grid", "dialog"}},
	"aria-hidden":                 {Type: "boolean", AllowUndefined: true},
	"aria-invalid":                {Type: "token", Values: []any{"grammar", false, "spelling", true}},
	"aria-keyshortcuts":           {Type: "string"},
	"aria-label":                  {Type: "string"},
	"aria-labelledby":             {Type: "idlist"},
	"aria-level":                  {Type: "integer"},
	"aria-live":                   {Type: "token", Values: []any{"assertive", "off", "polite"}},
	"aria-modal":                  {Type: "boolean"},
	"aria-multiline":              {Type: "boolean"},
	"aria-multiselectable":        {Type: "boolean"},
	"aria-orientation":            {Type: "token", Values: []any{"vertical", "undefined", "horizontal"}},
	"aria-owns":                   {Type: "idlist"},
	"aria-placeholder":            {Type: "string"},
	"aria-posinset":               {Type: "integer"},
	"aria-pressed":                {Type: "tristate"},
	"aria-readonly":               {Type: "boolean"},
	"aria-relevant":               {Type: "tokenlist", Values: []any{"additions", "all", "removals", "text"}},
	"aria-required":               {Type: "boolean"},
	"aria-roledescription":        {Type: "string"},
	"aria-rowcount":               {Type: "integer"},
	"aria-rowindex":               {Type: "integer"},
	"aria-rowspan":                {Type: "integer"},
	"aria-selected":               {Type: "boolean", AllowUndefined: true},
	"aria-setsize":                {Type: "integer"},
	"aria-sort":                   {Type: "token", Values: []any{"ascending", "descending", "none", "other"}},
	"aria-valuemax":               {Type: "number"},
	"aria-valuemin":               {Type: "number"},
	"aria-valuenow":               {Type: "number"},
	"aria-valuetext":              {Type: "string"},
}

// AriaLiteralValueKind discriminates the literal-typed value of a JSX
// attribute, mirroring jsx-ast-utils' `getLiteralPropValue` output.
//
// Used by jsx-a11y/aria-proptypes to branch validity-check logic on the
// runtime type of the literal value, including the boolean attribute form
// (`<div aria-hidden />` → `AriaLiteralBool` with `true`) and the
// upstream "not a literal expression" path (Identifier non-undefined /
// CallExpression / MemberExpression / Conditional / Logical / Binary /
// TSAsExpression / TSNonNullExpression / TSSatisfiesExpression — all noop
// in LITERAL_TYPES → null → `AriaLiteralNoLit`).
type AriaLiteralValueKind int

const (
	// AriaLiteralNoLit signals jsx-ast-utils' `getLiteralPropValue` returned
	// `null` — the LITERAL_TYPES noop path. Triggers upstream's
	// `if (value === null) return;` short-circuit.
	AriaLiteralNoLit AriaLiteralValueKind = iota
	// AriaLiteralUndef signals an explicit `undefined` identifier resolved
	// at parse time. Step-1 (`getPropValue == null`) normally catches this
	// before we get here, so the value is rare in practice — preserved for
	// upstream's `allowUndefined && value === undefined` branch parity.
	AriaLiteralUndef
	// AriaLiteralBool covers `{true}` / `{false}` / `<p />` boolean form
	// AND string-literal "true" / "false" (case-insensitive) which
	// jsx-ast-utils' Literal extractor coerces to actual booleans.
	AriaLiteralBool
	// AriaLiteralNumber covers numeric literals (including unary +/-/~)
	// and the JS NaN value for failed parses.
	AriaLiteralNumber
	// AriaLiteralBigInt covers BigInt literals (`123n`). jsx-ast-utils
	// has no special handling beyond returning the BigInt itself; ARIA
	// types don't recognize bigints so all kinds report invalid except
	// where the validityCheck explicitly inspects typeof.
	AriaLiteralBigInt
	// AriaLiteralString covers StringLiteral / NoSubstitutionTemplateLiteral
	// (verbatim, no boolean coerce) / TemplateExpression with substitutions
	// (synthesized placeholder string) / null-literal (LITERAL_TYPES.Literal
	// override returns the string "null") / assignment-expression
	// synthesized string.
	AriaLiteralString
)

// AriaLiteralValue is the literal-typed value of a JSX attribute, as seen by
// jsx-ast-utils' `getLiteralPropValue`. Discriminated by Kind — only the
// matching field is meaningful.
type AriaLiteralValue struct {
	Kind   AriaLiteralValueKind
	Bool   bool
	Num    float64
	BigInt *big.Int
	Str    string
}

// AriaLiteralValueAsJSNumber mirrors JS `Number(value)` semantics on an
// AriaLiteralValue:
//
//   - jvNumber: value passes through; NaN → (0, false).
//   - jvString: trim whitespace, empty → 0, hex (`0x` / `0X`), octal (`0o` /
//     `0O`), and binary (`0b` / `0B`) unsigned-integer prefixes handled
//     explicitly (Go's strconv.ParseFloat does not recognize them, JS
//     Number does), decimal via strconv.ParseFloat otherwise. Failed
//     parse → (0, false).
//   - jvBool: false → 0, true → 1 (mirrors JS `Number(true) = 1`).
//   - jvBigInt: float64 cast (out-of-range → ±Inf). Mirrors JS `Number(2n) = 2`.
//   - jvUndef: `Number(undefined) = NaN` → (0, false).
//   - Other kinds (NoLit) → (0, false).
//
// Returns (value, false) for any NaN-producing input — callers who need
// to distinguish "valid 0" from "no number" can use the boolean second
// return; callers who just need `isNaN(Number(value)) === false` can
// negate the boolean directly.
//
// Implementation reuses jsValueToNumber from tabindex_no_positive.go — the
// only difference is the input type (AriaLiteralValue vs the internal
// jsValue), so this is a thin adapter rather than a parallel implementation.
func AriaLiteralValueAsJSNumber(v AriaLiteralValue) (float64, bool) {
	return jsValueToNumber(ariaLiteralValueToInternal(v))
}

// ariaLiteralValueToInternal bridges the public AriaLiteralValue surface
// with the package-internal jsValue used by static_eval.go and
// tabindex_no_positive.go. Single hop, no semantics added.
func ariaLiteralValueToInternal(v AriaLiteralValue) jsValue {
	switch v.Kind {
	case AriaLiteralBool:
		return jsValue{Kind: jvBool, Bool: v.Bool}
	case AriaLiteralNumber:
		return jsValue{Kind: jvNumber, Num: v.Num}
	case AriaLiteralBigInt:
		return jsValue{Kind: jvBigInt, Big: v.BigInt}
	case AriaLiteralString:
		return jsValue{Kind: jvString, Str: v.Str}
	case AriaLiteralUndef:
		return jsUndef
	}
	// AriaLiteralNoLit — upstream's noop → null path.
	return jsNull
}

// LiteralPropAriaValue extracts the literal-typed value of a JSX attribute,
// mirroring jsx-ast-utils' `getLiteralPropValue` output kinds. Used by
// jsx-a11y/aria-proptypes to drive type-specific validity checks.
//
// Behavior table:
//
//	<p />                       → Bool{true}   (boolean attribute form)
//	<p={}>                      → NoLit        (empty JsxExpression)
//	<p="text">                  → String{text} (NOT "true"/"false" — those coerce to Bool)
//	<p="true"> / <p="false">    → Bool         (jsxAstUtilsLiteralCoerce)
//	<p={true}> / <p={false}>    → Bool
//	<p={`text`}>                → String{text} (NoSubstitutionTemplate — NO boolean coerce)
//	<p={`a${b}c`}>              → String       (synthesized placeholder via staticEvalTemplate)
//	<p={N}>                     → Number{N}
//	<p={Nn}>                    → BigInt{N}
//	<p={null}>                  → String{"null"} (LITERAL_TYPES.Literal override)
//	<p={undefined}>             → Undef
//	<p={someVar}>               → NoLit        (LITERAL_TYPES.Identifier non-undefined → null)
//	<p={foo.bar}>               → NoLit        (LITERAL_TYPES.MemberExpression noop → null)
//	<p={fn()}>                  → NoLit        (LITERAL_TYPES.CallExpression noop → null)
//	<p={cond ? a : b}>          → NoLit        (LITERAL_TYPES.ConditionalExpression noop → null)
//	<p={a && b}> / <p={a || b}> → NoLit        (LITERAL_TYPES.LogicalExpression noop → null)
//	<p={a + b}>                 → NoLit        (LITERAL_TYPES.BinaryExpression noop → null)
//	<p={x as any}>              → NoLit        (LITERAL_TYPES has no TSAsExpression — noop)
//	<p={x!}>                    → NoLit        (LITERAL_TYPES has no TSNonNullExpression — noop)
//	<p={x satisfies T}>         → NoLit        (LITERAL_TYPES has no TSSatisfiesExpression — noop)
//	<p={!x}>                    → Bool         (LITERAL_TYPES.UnaryExpression mirrors TYPES for `!`)
//	<p={-N}> / <p={+N}> / <p={~N}> → Number    (UnaryExpression numeric ops)
//	<p={<el />}>                → NoLit        (LITERAL_TYPES has no JSX — default noop)
//	<p={[a, b]}>                → NoLit        (we model arrays as truthy via jvTruthy; not a literal value)
//
// Parentheses are transparently unwrapped on every layer; TS-wrapper kinds
// (`as`, `!`, `satisfies`, `<T>`) are intentionally NOT stripped — upstream's
// LITERAL_TYPES rejects them, so we mirror.
func LiteralPropAriaValue(attr *ast.Node) AriaLiteralValue {
	if attr == nil {
		return AriaLiteralValue{Kind: AriaLiteralNoLit}
	}
	if AttributeIsBooleanForm(attr) {
		// `<div aria-hidden />` — upstream's extractValue maps the null
		// initializer to JS boolean true. The boolean form is the ONLY
		// path that produces a non-null `getPropValue` AND a non-null
		// `getLiteralPropValue` without an inner expression.
		return AriaLiteralValue{Kind: AriaLiteralBool, Bool: true}
	}
	inner := attributeInnerExpression(attr)
	if inner == nil {
		// `<div aria-hidden={} />` — empty JsxExpression. tsgo synthesizes
		// this only for malformed source; upstream falls through to noop
		// → null. Mirror as NoLit.
		return AriaLiteralValue{Kind: AriaLiteralNoLit}
	}
	v := literalPropValue(inner)
	switch v.Kind {
	case jvBool:
		return AriaLiteralValue{Kind: AriaLiteralBool, Bool: v.Bool}
	case jvNumber:
		return AriaLiteralValue{Kind: AriaLiteralNumber, Num: v.Num}
	case jvBigInt:
		return AriaLiteralValue{Kind: AriaLiteralBigInt, BigInt: v.Big}
	case jvString:
		return AriaLiteralValue{Kind: AriaLiteralString, Str: v.Str}
	case jvUndef:
		return AriaLiteralValue{Kind: AriaLiteralUndef}
	}
	// jvNull, jvUnknown, jvTruthy, jvFunction — all the upstream "noop →
	// null" / runtime-only paths. None are accepted as literal ARIA
	// values, so they fold into AriaLiteralNoLit which step-3 short-circuits.
	return AriaLiteralValue{Kind: AriaLiteralNoLit}
}

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
