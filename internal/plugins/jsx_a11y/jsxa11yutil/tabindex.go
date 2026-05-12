package jsxa11yutil

import (
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsNonLiteralProperty mirrors upstream's `isNonLiteralProperty(attributes,
// propName)`. Returns true iff the named JSX prop's value is something OTHER
// than a Literal-typed initializer or a JSXExpression containing
// `undefined` / a JSX text. The exact upstream branches:
//
//	if (!prop)                       return false;
//	if (!prop.value)                 return false;          // boolean form
//	if (prop.value.type === 'Literal') return false;        // direct string
//	if (prop.value.type === 'JSXExpressionContainer') {
//	  const e = prop.value.expression;
//	  if (e.type === 'Identifier' && e.name === 'undefined') return false;
//	  if (e.type === 'JSXText')                              return false;
//	  // every other expression — INCLUDING `Literal` inside
//	  // JSXExpressionContainer — falls through.
//	}
//	return true;
//
// Note: `<div role={"button"} />` returns TRUE — the inner StringLiteral is
// inside a JsxExpression, and upstream's `JSXExpressionContainer` branch only
// short-circuits on `undefined` / JSXText, not on inner Literals. This is
// load-bearing for the `allowExpressionValues` option in
// no-noninteractive-tabindex.
//
// tsgo↔ESTree note: ESTree flattens parens at parse time so
// `<div role={(undefined)} />` still hits the `Identifier 'undefined'` arm;
// tsgo preserves the ParenthesizedExpression. utils.IsUndefinedIdentifier
// strips parens internally so the behavior matches.
func IsNonLiteralProperty(attrs []*ast.Node, propName string) bool {
	prop := FindAttributeByName(attrs, propName)
	if prop == nil {
		return false
	}
	if prop.Kind != ast.KindJsxAttribute {
		// Spread-resolved PropertyAssignment / ShorthandPropertyAssignment.
		// Upstream's `getProp` synthesizes a JSXAttribute; treat any spread
		// resolution as non-literal so the `allowExpressionValues` skip
		// arm stays conservative.
		return true
	}
	init := prop.AsJsxAttribute().Initializer
	if init == nil {
		// Boolean form `<div role />` — upstream's `!propValue` returns false.
		return false
	}
	// Direct StringLiteral initializer maps to ESTree's `Literal` type.
	if init.Kind == ast.KindStringLiteral {
		return false
	}
	if init.Kind == ast.KindJsxExpression {
		expr := init.AsJsxExpression().Expression
		if expr == nil {
			// `<div role={} />` — empty JsxExpression. tsgo synthesizes this
			// only for malformed source. Treat as non-literal — upstream's
			// JSXExpressionContainer arm's payload is JSXEmptyExpression
			// which is neither Identifier 'undefined' nor JSXText, so it
			// falls through to `return true`.
			return true
		}
		// utils.IsUndefinedIdentifier internally strips parens, matching
		// the upstream `Identifier 'undefined'` arm regardless of paren
		// nesting (ESTree flattens parens at parse time; tsgo preserves
		// them).
		if utils.IsUndefinedIdentifier(expr) {
			return false
		}
		if ast.SkipParentheses(expr).Kind == ast.KindJsxText {
			return false
		}
	}
	return true
}

// GetTabIndex mirrors upstream's `getTabIndex(tabIndex)` exactly:
//
//	const literalValue = getLiteralPropValue(tabIndex);
//	if (typeof literalValue === 'string' || typeof literalValue === 'number') {
//	  if (typeof literalValue === 'string' && literalValue.length === 0) return undefined;
//	  const value = Number(literalValue);
//	  if (Number.isNaN(value)) return undefined;
//	  return Number.isInteger(value) ? value : undefined;
//	}
//	if (literalValue === true || literalValue === false) return undefined;
//	return getPropValue(tabIndex);
//
// Returned (val, true) means upstream's getTabIndex would yield a usable
// number (literal-numeric integer, or any value getPropValue can statically
// derive). The downstream `tabIndex >= 0` check is left to the caller —
// returning float64 so the caller's `>= 0` comparison mirrors JS exactly,
// including the ToNumber-coercion semantics of the step-2 fallback (where
// upstream skips the `Number.isInteger` filter and lets JS's loose `>=`
// handle string / boolean / number identically).
//
// Returns (0, false) when upstream would yield either `undefined` or `null`.
// Callers that need to distinguish the two — because they implement an
// upstream comparison like `tabIndex >= -1` whose result diverges between
// `undefined >= -1` (NaN, false) and `null >= -1` (ToNumber(null)=0, true) —
// must use [GetTabIndexEx] instead.
//
// Specifically, (0, false) covers:
//   - the prop is absent / boolean-form (getPropValue would yield true →
//     undefined per the step-1 boolean arm)
//   - the literal value is an empty string / non-integer number / NaN
//   - the literal value is a boolean (step-1 boolean arm)
//   - getPropValue cannot statically resolve the expression
func GetTabIndex(attr *ast.Node, sourceText string) (float64, bool) {
	val, resolved, _ := GetTabIndexEx(attr, sourceText)
	return val, resolved
}

// GetTabIndexEx mirrors upstream `getTabIndex(getProp(attrs, 'tabIndex'))`
// with explicit two-state classification of the unresolved case. Used by
// callers whose downstream comparison is loose JS `>=` / `>` / `<=` / `<`
// against a number, where `null` and `undefined` differ:
//
//	null >= -1      → ToNumber(null) = 0      → true
//	undefined >= -1 → ToNumber(undefined) = NaN → false
//
// Returns:
//   - (val, true,  false) — upstream resolved a usable Number. Callers
//     compare `val` directly.
//   - (0,   false, false) — upstream would yield `undefined`. Callers should
//     treat the comparison `tabIndex op N` as if `undefined op N` (NaN, so
//     all loose comparisons are false).
//   - (0,   false, true)  — upstream would yield `null`. Callers should
//     treat the comparison as if `null op N` (ToNumber(null)=0, so JS-loose
//     comparison applies).
//
// The undefined-vs-null split mirrors jsx-ast-utils' two extraction paths:
// step-1 `getLiteralPropValue` returns undefined for boolean form / empty
// string / NaN-coercing literal / boolean literal; step-2 `getPropValue`
// returns null for any ESTree expression type its TYPES table does not
// recognize (TSSatisfiesExpression, AwaitExpression, YieldExpression,
// ImportExpression, ...).
func GetTabIndexEx(attr *ast.Node, sourceText string) (val float64, resolved bool, nullLike bool) {
	if attr == nil {
		// upstream: getProp(attrs, 'tabIndex') === undefined → getTabIndex
		// passes undefined through unchanged.
		return 0, false, false
	}
	if AttributeIsBooleanForm(attr) {
		// `<div tabIndex />` — extractValue's null-attribute-value path
		// produces JS boolean `true`, which upstream's getTabIndex passes
		// through to its `=== true` branch → undefined.
		return 0, false, false
	}

	// Direct StringLiteral JsxAttribute (`tabIndex="…"`, NOT
	// `tabIndex={"…"}` which wraps in JsxExpression) carries HTML-style
	// attribute text. ESTree parsers decode HTML entities (`&#49;`→"1",
	// `&nbsp;`→U+00A0, …) before passing the value to lint rules; tsgo
	// keeps the raw `&…;` source in StringLiteral.Text, so
	// directAttributeStringValue applies jsxtransforms.DecodeEntities to
	// realign with upstream getLiteralPropValue output for this shape. The
	// decoded string then routes through the same step-1 classification as
	// a plain StringLiteral.
	if decoded, ok := directAttributeStringValue(attr); ok {
		return classifyTabIndexString(decoded)
	}

	inner := attributeInnerExpression(attr)
	if inner == nil {
		// `<div tabIndex={} />` — empty JsxExpression body / JSXEmptyExpression.
		// jsx-ast-utils' TYPES has no entry for JSXEmptyExpression → noop →
		// null. Downstream `null >= 0` ToNumber-coerces to true, so upstream
		// reports. Classify as null-like to mirror.
		return 0, false, true
	}

	// Template literals — upstream's TemplateLiteral extractor reads
	// `quasi.value.raw` (literal source between back-ticks), NOT the
	// cooked `.Text` field. Override here before the literalPropValue
	// dispatch (which uses cooked text and is shared with other a11y
	// rules that care about truthiness rather than ToNumber coercion).
	if inner.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return classifyTabIndexString(rawTemplateLiteralText(inner, sourceText))
	}
	if inner.Kind == ast.KindTemplateExpression {
		v := templateExpressionRawText(inner.AsTemplateExpression(), sourceText)
		return classifyTabIndexString(v.Str)
	}

	// Cluster A: TSNonNullExpression — upstream's TSNonNullExpression
	// extractor stringifies (`0!` → "0!", `(0)!` → "0!", `x!` → "x!"). The
	// resulting string always contains `!` and ToNumbers to NaN, so step-1
	// resolves to undefined. Without this branch staticEval's
	// OEKNonNullAssertions strip would resolve `0!` to the bare 0 — wrong.
	if inner.Kind == ast.KindNonNullExpression {
		return 0, false, false
	}

	// Step 1: getLiteralPropValue path — upstream applies the integer
	// requirement and the empty-string / boolean → undefined short-circuits.
	literalV := literalPropValue(inner)
	switch literalV.Kind {
	case jvString:
		if literalV.Str == "" {
			// upstream `length === 0` short-circuit → undefined.
			return 0, false, false
		}
		if v, ok := parseLiteralTabIndexString(literalV.Str); ok {
			return v, true, false
		}
		// upstream `Number.isNaN(value)` → undefined.
		return 0, false, false
	case jvNumber:
		if v, ok := parseLiteralTabIndexFloat(literalV.Num); ok {
			return v, true, false
		}
		// upstream `Number.isInteger(value) ? value : undefined`.
		return 0, false, false
	case jvBool:
		// upstream step-1 boolean arm → undefined.
		return 0, false, false
	case jvBigInt:
		// Cluster B: BigInt literal. Upstream getLiteralPropValue returns
		// the BigInt; getTabIndex falls through to step-2 (typeof bigint is
		// not 'string'/'number'/bool) → returns the BigInt → downstream
		// `>=` ToNumber-coerces (`Number(0n)=0`, `Number(-1n)=-1`,
		// out-of-range → ±Infinity). Match by computing Number(BigInt)
		// directly and surfacing it as resolved.
		if n, ok := jsValueToNumber(literalV); ok {
			return n, true, false
		}
		return 0, false, false
	}

	// Step 2: getPropValue fallback — covers expressions that LITERAL_TYPES
	// returns null for (BinaryExpression, ConditionalExpression, Logical /
	// Nullish, MemberExpression, CallExpression, etc.). Upstream returns the
	// raw getPropValue result and lets downstream `tabIndex >= 0` handle
	// JS-Number coercion — so we mirror that loose check here without the
	// step-1 integer requirement.
	//
	// Two specializations precede the default staticEval dispatch because
	// staticEval's coarse sentinels (jsTruthy for ArrayLiteralExpression,
	// jsNull for unary `+`/`-` on string operand) discard the precision the
	// downstream `>= 0` coercion needs. They are kept local to GetTabIndex
	// so the shared staticEval engine is not affected.
	if inner.Kind == ast.KindArrayLiteralExpression {
		// Cluster F: route through arrayLiteralPropToTabIndexString (same
		// engine tabindex-no-positive uses) — it handles nested arrays via
		// recursion, unary on non-number operands via the TYPES.UnaryExpression
		// protocol, and ObjectExpression / RegExp / function / JsxElement
		// via jsToString (which produces non-empty non-numeric strings →
		// downstream NaN). The pre-Ex inline arrayToTabIndex lost precision
		// on these by collapsing jsTruthy to "" → 0.
		arrV := arrayLiteralPropToTabIndexString(inner, sourceText)
		if n, ok := jsValueToNumber(arrV); ok && !math.IsNaN(n) {
			return n, true, false
		}
		return 0, false, false
	}
	if inner.Kind == ast.KindPrefixUnaryExpression {
		if v, ok, applicable := unaryStringToTabIndex(inner); applicable {
			if ok {
				return v, true, false
			}
			return 0, false, false
		}
		// Cluster E: jsx-ast-utils' UnaryExpression handler ToNumber-coerces
		// the operand before applying `+` / `-` / `~`. Our staticEvalUnary
		// only computes when operand is already jvNumber, so `+undefined`,
		// `-true`, `~null`, etc. fall through to its jsNull return → in
		// our previous Ex impl that landed on nullLike → REPORT (wrong).
		// Mirror upstream by running ToNumber on the operand here.
		un := inner.AsPrefixUnaryExpression()
		if un.Operator == ast.KindPlusToken || un.Operator == ast.KindMinusToken || un.Operator == ast.KindTildeToken || un.Operator == ast.KindExclamationToken {
			opV := unaryLiteralPropToTabIndexJsValue(un)
			if n, ok := jsValueToNumber(opV); ok {
				if !math.IsNaN(n) {
					return n, true, false
				}
			}
			return 0, false, false
		}
	}
	staticV := staticEval(inner)
	if v, ok := staticEvalToTabIndex(staticV); ok {
		return v, true, false
	}
	// staticEvalToTabIndex failed — classify by the inner sentinel:
	//   - jvNull, jvUnknown → upstream's "TYPES has no entry" / explicit
	//     AwaitExpression / YieldExpression / TSSatisfiesExpression / etc.
	//     → returns null. null-like.
	//   - jvUndef → upstream resolved to undefined (jvx Identifier
	//     `undefined`, void/typeof, ...) → undefined-like.
	//   - jvFunction / jvTruthy / jvBool / jvBigInt → upstream resolved to
	//     a non-numeric truthy that ToNumber-coerces to NaN (or, for
	//     BigInt, to a number our extractor doesn't yet model) → in either
	//     case the loose `>= -1` lands on the same arm as undefined. We
	//     classify as undefined-like to keep callers reporting on these
	//     until BigInt support is added (see TODO above arrayToTabIndex
	//     for the BigInt expansion path).
	switch staticV.Kind {
	case jvNull, jvUnknown:
		return 0, false, true
	}
	return 0, false, false
}

// HasUpstreamTabIndexValue mirrors upstream's
// `getTabIndex(getProp(attrs, 'tabIndex')) !== undefined` boolean check —
// is the tabIndex prop set to ANY value that upstream getTabIndex doesn't
// resolve to JS undefined?
//
// Differs from [GetTabIndex] / [GetTabIndexEx]:
//   - GetTabIndex returns true only when the value is a usable focus order
//     (integer-typed step-1 win, or step-2 numeric coercion).
//   - HasUpstreamTabIndexValue also returns true when upstream's step-2
//     getPropValue falls back to a non-numeric non-undefined value — e.g.
//     Identifier `someVar` (whose name string passes through TYPES.Identifier),
//     Call / Member / JsxElement (synthesized non-empty strings), Object /
//     Array / Function / BigInt (typeof "object"/"function"/"bigint" — all
//     `!== undefined`).
//
// Returns false only for the shapes that upstream getTabIndex outputs JS
// `undefined`:
//   - missing prop
//   - boolean form (`<X tabIndex />`) — getPropValue → true → step-1 boolean
//     arm → undefined
//   - empty string / non-integer numeric / NaN-coercing literal
//   - direct boolean (`{true}` / `{false}`)
//   - explicit `{undefined}` and TS-wrapped variants, `typeof` / `void`
//     expressions
//   - TSNonNullExpression (upstream stringifies to "x!" → NaN → undefined)
//
// Used by interactive-supports-focus's `hasTabindex` gate.
func HasUpstreamTabIndexValue(attrs []*ast.Node, sourceText string) bool {
	attr := FindAttributeByName(attrs, "tabIndex")
	if attr == nil {
		return false
	}
	if AttributeIsBooleanForm(attr) {
		return false
	}

	// Direct StringLiteral attribute carries HTML-style entity-encoded text
	// — directAttributeStringValue applies jsxtransforms.DecodeEntities so we
	// see the same value @typescript-eslint / @babel's JSX parser would expose.
	if decoded, ok := directAttributeStringValue(attr); ok {
		return tabIndexStringHasUpstreamValue(decoded)
	}

	inner := attributeInnerExpression(attr)
	if inner == nil {
		// empty `{}` — TYPES.JSXEmptyExpression missing → null → `null !== undefined` → true.
		return true
	}

	if inner.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return tabIndexStringHasUpstreamValue(rawTemplateLiteralText(inner, sourceText))
	}
	if inner.Kind == ast.KindTemplateExpression {
		v := templateExpressionRawText(inner.AsTemplateExpression(), sourceText)
		return tabIndexStringHasUpstreamValue(v.Str)
	}
	if inner.Kind == ast.KindNonNullExpression {
		// upstream TSNonNullExpression extractor stringifies the operand
		// with a trailing `!` → ParseFloat fails → step-1 undefined.
		return false
	}

	literalV := literalPropValue(inner)
	switch literalV.Kind {
	case jvString:
		return tabIndexStringHasUpstreamValue(literalV.Str)
	case jvNumber:
		if _, ok := parseLiteralTabIndexFloat(literalV.Num); ok {
			return true
		}
		return false
	case jvBool:
		return false
	case jvBigInt:
		// step-1 typeof bigint not in string/number/boolean → falls through
		// to step-2, which returns the bigint → `!== undefined`.
		return true
	}
	// step 2: getPropValue fallback. Anything that isn't JS undefined passes.
	return !jsValueIsExactlyUndefined(staticEval(inner))
}

// tabIndexStringHasUpstreamValue reduces getTabIndex's step-1 string
// classification to the binary "result !== undefined" question. Used by
// [HasUpstreamTabIndexValue] for direct StringLiteral / NoSubstitutionTemplateLiteral
// / TemplateExpression shapes; the same string semantics as
// classifyTabIndexString without the numeric value extraction.
func tabIndexStringHasUpstreamValue(s string) bool {
	if s == "" {
		return false
	}
	switch strings.ToLower(s) {
	case "true", "false":
		// jsxAstUtilsLiteralCoerce → boolean → step-1 boolean arm → undefined.
		return false
	}
	if _, ok := parseLiteralTabIndexString(s); ok {
		return true
	}
	return false
}

// classifyTabIndexString applies upstream getTabIndex's step-1 string
// semantics to an already-extracted string value (post-entity-decode for
// direct attribute strings; raw quasi text for templates). Returns the
// same three-state result as GetTabIndexEx.
func classifyTabIndexString(s string) (float64, bool, bool) {
	if s == "" {
		// upstream `length === 0` short-circuit → undefined.
		return 0, false, false
	}
	// jsx-ast-utils' Literal extractor coerces case-insensitive "true" /
	// "false" string values to actual booleans, which upstream's
	// getTabIndex then routes to its boolean arm → undefined.
	switch strings.ToLower(s) {
	case "true", "false":
		return 0, false, false
	}
	if v, ok := parseLiteralTabIndexString(s); ok {
		return v, true, false
	}
	// upstream `Number.isNaN(value)` → undefined.
	return 0, false, false
}

// unaryStringToTabIndex models JS unary `+x` / `-x` ToNumber coercion for
// the case where staticEvalUnary's number-only guard returns null — i.e.
// when the operand resolves to a string (Identifier / Member access / Call
// / nested expression). Mirrors jsx-ast-utils' UnaryExpression extractor.
//
// Returns (val, ok, true) when the node is `+expr`/`-expr` and the operand
// resolves to a coercible string. Returns (_, _, false) when the node is
// not a unary `+`/`-` (caller should fall through to the default staticEval
// path, which already handles unary on numeric operands).
func unaryStringToTabIndex(node *ast.Node) (float64, bool, bool) {
	un := node.AsPrefixUnaryExpression()
	if un.Operator != ast.KindPlusToken && un.Operator != ast.KindMinusToken {
		return 0, false, false
	}
	operandV := staticEval(un.Operand)
	if operandV.Kind != jvString {
		// Numeric / boolean operands are already covered by staticEvalUnary
		// + staticEvalToTabIndex; only string operands need this hop.
		return 0, false, false
	}
	val, ok := staticEvalToTabIndex(operandV)
	if !ok {
		// `+"abc"` → NaN → undefined → not reported.
		return 0, false, true
	}
	if un.Operator == ast.KindMinusToken {
		return -val, true, true
	}
	return val, true, true
}

// parseLiteralTabIndexString applies upstream's step-1 numeric coercion:
// JS `Number(string)` (ECMA-262 7.1.4.1.1 StringToNumber) on the trimmed
// input, then `Number.isInteger` check.
//
// Hex (`0x` / `0X`), octal (`0o` / `0O`), and binary (`0b` / `0B`) prefixes
// are accepted as unsigned-integer forms — JS `Number("0x10")` is 16. Go's
// `strconv.ParseFloat` doesn't recognize these prefixes, so they're handled
// explicitly. JS's StringToNumber rejects signed prefixes for these bases
// (`Number("-0x10")` is NaN), so we never strip a sign before the prefix
// dispatch.
func parseLiteralTabIndexString(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		// ECMA-262 StringToNumber("") → 0, and `0 >= 0` triggers the
		// rule's report. Important: literalPropValue's caller already
		// short-circuits on the *raw* (pre-trim) empty case to mirror
		// upstream's `literalValue.length === 0` early-return. We only
		// land here when the raw string had non-zero length but trimmed
		// to empty (`tabIndex="   "`), which upstream coerces to 0.
		return 0, true
	}
	// Hex / octal / binary unsigned-integer literals.
	if len(s) > 2 && s[0] == '0' {
		switch s[1] {
		case 'x', 'X':
			if n, err := strconv.ParseUint(s[2:], 16, 64); err == nil {
				return float64(n), true
			}
			return 0, false
		case 'o', 'O':
			if n, err := strconv.ParseUint(s[2:], 8, 64); err == nil {
				return float64(n), true
			}
			return 0, false
		case 'b', 'B':
			if n, err := strconv.ParseUint(s[2:], 2, 64); err == nil {
				return float64(n), true
			}
			return 0, false
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	if f != math.Trunc(f) {
		return 0, false
	}
	return f, true
}

// parseLiteralTabIndexFloat applies upstream's step-1 integer requirement
// to a numeric literal value.
func parseLiteralTabIndexFloat(f float64) (float64, bool) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	if f != math.Trunc(f) {
		return 0, false
	}
	return f, true
}

// staticEvalToTabIndex mirrors the step-2 `getPropValue → tabIndex >= 0`
// loose comparison. Unlike step 1, the integer requirement does NOT apply
// (upstream's downstream `>= 0` is just JS loose comparison and accepts
// non-integers, strings, booleans, and ±Infinity transparently via ToNumber).
//
//   - jvNumber  → return as-is; NaN excluded (NaN>=0 is false) but ±Infinity
//     kept (`Infinity>=0` is true → upstream reports;
//     `-Infinity>=0` is false → caller skips by comparison).
//   - jvString  → Trim + Number-coerce (incl. hex/octal/binary prefixes
//     and JS empty-string → 0). Non-numeric → false.
//   - jvBool    → false → 0, true → 1 (both ≥ 0 so both trigger report).
//   - else      → false. Upstream's `>= 0` on null/undefined/object/array
//     could in principle ToPrimitive-coerce to a numeric string in some
//     cases (e.g. `[5] >= 0` is true via "5" → 5). Those shapes are not
//     reachable in real tabIndex usage and supporting them would require
//     extending staticEval globally, so they are intentionally left
//     unhandled — see "Known divergences" in the package doc.
func staticEvalToTabIndex(v jsValue) (float64, bool) {
	switch v.Kind {
	case jvNumber:
		if math.IsNaN(v.Num) {
			return 0, false
		}
		return v.Num, true
	case jvString:
		// Mirror JS ToNumber(string): trim whitespace, return 0 for empty
		// (upstream's `"" >= 0` is true via ToNumber), accept hex / oct /
		// bin prefixes, parse decimal otherwise.
		s := strings.TrimSpace(v.Str)
		if s == "" {
			return 0, true
		}
		if len(s) > 2 && s[0] == '0' {
			switch s[1] {
			case 'x', 'X':
				if n, err := strconv.ParseUint(s[2:], 16, 64); err == nil {
					return float64(n), true
				}
				return 0, false
			case 'o', 'O':
				if n, err := strconv.ParseUint(s[2:], 8, 64); err == nil {
					return float64(n), true
				}
				return 0, false
			case 'b', 'B':
				if n, err := strconv.ParseUint(s[2:], 2, 64); err == nil {
					return float64(n), true
				}
				return 0, false
			}
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil || math.IsNaN(f) {
			return 0, false
		}
		return f, true
	case jvBool:
		// Mirror JS ToNumber(boolean): false → 0, true → 1. Both >= 0 so
		// both report. Upstream's step-1 boolean arm short-circuits to
		// undefined for *direct* booleans, but step-2 has no boolean
		// guard, so a `cond ? true : false` ternary lands here and reports.
		if v.Bool {
			return 1, true
		}
		return 0, true
	case jvBigInt:
		// Mirror JS Number(bigint): floors to Float64 (out-of-range → ±Inf).
		return jsValueToNumber(v)
	}
	return 0, false
}
