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
// Returns (0, false) when:
//   - the prop is absent / boolean-form (getPropValue would yield true →
//     undefined per the step-1 boolean arm)
//   - the literal value is an empty string / non-integer number / NaN
//   - the literal value is a boolean (step-1 boolean arm)
//   - getPropValue cannot statically resolve the expression
func GetTabIndex(attr *ast.Node) (float64, bool) {
	if attr == nil {
		return 0, false
	}
	if AttributeIsBooleanForm(attr) {
		// `<div tabIndex />` — extractValue's null-attribute-value path
		// produces JS boolean `true`, which upstream's getTabIndex passes
		// through to its `=== true` branch → undefined.
		return 0, false
	}
	inner := attributeInnerExpression(attr)
	if inner == nil {
		return 0, false
	}

	// Step 1: getLiteralPropValue path — upstream applies the integer
	// requirement and the empty-string / boolean → undefined short-circuits.
	literalV := literalPropValue(inner)
	switch literalV.Kind {
	case jvString:
		if literalV.Str == "" {
			return 0, false
		}
		return parseLiteralTabIndexString(literalV.Str)
	case jvNumber:
		return parseLiteralTabIndexFloat(literalV.Num)
	case jvBool:
		return 0, false
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
		return arrayToTabIndex(inner)
	}
	if inner.Kind == ast.KindPrefixUnaryExpression {
		if val, ok, applicable := unaryStringToTabIndex(inner); applicable {
			return val, ok
		}
	}
	staticV := staticEval(inner)
	return staticEvalToTabIndex(staticV)
}

// arrayToTabIndex models JS `ToPrimitive(array, "default") → ToNumber` for
// an ArrayLiteralExpression as the tabIndex value. Mirrors upstream's
// getPropValue → ArrayExpression extractor + downstream `>= 0` coercion.
//
// Per spec (ECMA-262 Array.prototype.join):
//   - undefined / null / sparse elements stringify to "" (NOT "undefined" /
//     "null").
//   - other elements use String(element).
//
// Joined with "," then fed through ToNumber. Examples:
//
//	[]            → ""          → 0
//	[5]           → "5"         → 5
//	[1, 2]        → "1,2"       → NaN
//	[null]        → ""          → 0
//	[Infinity]    → "Infinity"  → +Inf (≥0 → reports)
//	[-1]          → "-1"        → -1  (<0 → skips)
func arrayToTabIndex(node *ast.Node) (float64, bool) {
	if node == nil || node.Kind != ast.KindArrayLiteralExpression {
		return 0, false
	}
	arr := node.AsArrayLiteralExpression()
	var elements []*ast.Node
	if arr != nil && arr.Elements != nil {
		elements = arr.Elements.Nodes
	}
	if len(elements) == 0 {
		// `[].toString() == ""` → ToNumber("") = 0.
		return 0, true
	}
	parts := make([]string, 0, len(elements))
	for _, el := range elements {
		if el == nil {
			parts = append(parts, "")
			continue
		}
		ev := staticEval(el)
		switch ev.Kind {
		case jvNull, jvUndef, jvUnknown:
			// Array.prototype.join special-cases null/undefined to "".
			// Unknown (statically unresolvable) values fall through the
			// same way — upstream's noop → null → "" via Array.join.
			parts = append(parts, "")
		default:
			parts = append(parts, jsToString(ev))
		}
	}
	return staticEvalToTabIndex(jsValue{Kind: jvString, Str: strings.Join(parts, ",")})
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
	}
	return 0, false
}
