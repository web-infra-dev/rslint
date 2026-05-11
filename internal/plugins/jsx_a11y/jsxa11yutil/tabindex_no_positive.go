package jsxa11yutil

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	jsxtx "github.com/microsoft/typescript-go/shim/transformers/jsxtransforms"
)

// LiteralPropToNumber mirrors upstream's `Number(getLiteralPropValue(prop))`
// pipeline — the exact pattern used by jsx-a11y/tabindex-no-positive's
// `Number(getLiteralPropValue(attribute))` extraction before the `> 0`
// gate. Returns (value, true) for a JS-Number-coerced result, or
// (0, false) when the coercion produces NaN (in which case upstream's
// `if (isNaN(value) || value <= 0) return;` short-circuits silently).
//
// Behavior table (validated against upstream eslint-plugin-jsx-a11y
// v6 via /tmp/tnp-verify):
//
//	Boolean form `<X tabIndex />`              → (1, true)   [extractValue null-attr → true → Number(true)=1]
//	`tabIndex={true}` / `tabIndex="true"`       → (1, true)   [jsxAstUtilsLiteralCoerce → bool]
//	`tabIndex={false}` / `tabIndex="false"`     → (0, true)
//	`tabIndex="5"` / `tabIndex="0x10"`          → ToNumber per JS StringToNumber
//	`tabIndex="abc"` / `tabIndex="-0x10"`       → (0, false)  [NaN]
//	`tabIndex=""` / `tabIndex="  "` / `{""}`    → (0, true)   [Number("")=0]
//	`tabIndex={1.589}` (non-integer)            → (1.589, true)  — NOT filtered, unlike no-noninteractive-tabindex
//	`tabIndex={undefined}`                      → (0, false)  [Number(undefined)=NaN]
//	`tabIndex={null}`                           → (0, false)  [LITERAL_TYPES.Literal: null→"null" → NaN]
//	`tabIndex={someVar}` (non-undefined Ident)  → (0, true)   [LITERAL_TYPES.Identifier=()=>null → Number(null)=0]
//	`tabIndex={[]}`                             → (0, true)   [empty array.join=""]
//	`tabIndex={[5]}` / `{[Infinity]}`           → 5 / +Inf — element extracted via TYPES path
//	`tabIndex={[5,6]}` / `{[null]}`             → (0, false) / (0, true)
//	`tabIndex={+true}` / `{!0}` / `{~-2}`       → unary applied per JS ToNumber semantics on operand
//	`tabIndex={delete a.b}`                     → (1, true)   [delete returns boolean true]
//	Other complex (Call / Member / Binary /
//	  Conditional / Logical / New / ...)         → (0, true)   [LITERAL_TYPES noop → null → Number(null)=0]
//
// Divergences vs jsxa11yutil.GetTabIndex (used by no-noninteractive-tabindex):
//
//   - boolean form here yields 1 (a positive number) so `tabIndex={true}`
//     reports; GetTabIndex returns "undefined" for the same shape so
//     no-noninteractive-tabindex skips it.
//   - non-integer numerics here pass through unfiltered (so `tabIndex={1.589}`
//     reports under tabindex-no-positive but is skipped by GetTabIndex's
//     `Number.isInteger` filter in step 1).
//   - this function does NOT have a `getPropValue` fallback path — upstream
//     tabindex-no-positive calls only `getLiteralPropValue`, so e.g.
//     `tabIndex={cond ? 1 : 2}` here yields 0 (LITERAL_TYPES.Conditional is
//     noop) and the rule skips; GetTabIndex's step 2 would resolve to 1 and
//     trigger no-noninteractive-tabindex's report.
func LiteralPropToNumber(attr *ast.Node, sourceText string) (float64, bool) {
	if attr == nil {
		return 0, false
	}
	if AttributeIsBooleanForm(attr) {
		// `<div tabIndex />` — extractValue null-attribute-value path returns
		// JS boolean true, then `Number(true) = 1`.
		return 1, true
	}
	// JsxAttribute with a *direct* StringLiteral initializer (the
	// `tabIndex="..."` shape, NOT `tabIndex={"..."}` which wraps in a
	// JsxExpression) is an HTML-style attribute string. babel's JSX parser
	// decodes HTML entities (`&#49;` → "1", `&nbsp;` → U+00A0, etc.) before
	// surfacing the value to lint rules. tsgo does NOT decode in its
	// StringLiteral.Text, so we apply jsxtransforms.DecodeEntities here to
	// realign with upstream `getLiteralPropValue` output for this shape.
	// Only the direct-StringLiteral path needs this — `tabIndex={"&#49;"}`
	// stays opaque (the inner StringLiteral is a JS expression, where
	// entities are not decoded).
	if attr.Kind == ast.KindJsxAttribute {
		if init := attr.AsJsxAttribute().Initializer; init != nil && init.Kind == ast.KindStringLiteral {
			decoded := jsxtx.DecodeEntities(init.AsStringLiteral().Text)
			return jsValueToNumber(jsxAstUtilsLiteralCoerce(decoded))
		}
	}
	inner := attributeInnerExpression(attr)
	if inner == nil {
		// `<div tabIndex={} />` — empty JsxExpression (only emitted for
		// malformed source). Upstream's getLiteralPropValue would route
		// through "type not in LITERAL_TYPES" → null → Number(null) = 0,
		// but the JSXEmptyExpression path produces undefined in some
		// builds. Treat as NaN to skip — the rule is permissive on
		// malformed input.
		return 0, false
	}
	return jsValueToNumber(literalPropToTabIndexJsValue(inner, sourceText))
}

// literalPropToTabIndexJsValue mirrors `getLiteralPropValue` for the cases
// where tabindex-no-positive's number-coercion semantics need finer
// resolution than the alt-text-shaped literalPropValue provides. Falls
// through to literalPropValue for everything else.
//
// Three kinds are re-routed:
//
//   - ArrayLiteralExpression: literalPropValue returns jsTruthy (alt-text /
//     role rules only care about truthiness). tabindex-no-positive needs
//     `Number([5]) = 5` precision via Array.join semantics, with element
//     extraction on the TYPES path (jsx-ast-utils' extractValueFromArrayExpression
//     imports the TYPES extractor, NOT LITERAL_TYPES, so `[Infinity]` resolves
//     to "Infinity" not to "").
//
//   - PrefixUnaryExpression: literalPropValue's staticEvalUnary only computes
//     `+`/`-`/`~` when the operand is already jvNumber. tabindex-no-positive
//     needs the full LITERAL_TYPES.UnaryExpression semantics, where the
//     operand is first ToNumber-coerced and then the unary is applied — so
//     `+true` yields 1 (reports), `+null` yields 0 (skips), etc. `!` is
//     fine via the generic path but we route it here for uniformity.
//
//   - DeleteExpression: literalPropValue falls through to jsNull, but
//     upstream's UnaryExpression extractor treats `delete x` as the boolean
//     `true` (the runtime semantic of the operator on a never-thrown path),
//     so `<div tabIndex={delete a.b} />` reports. typescript-go splits
//     DeleteExpression into its own kind, so we handle it explicitly.
//
// TypeOfExpression / VoidExpression / PostfixUnaryExpression (x++) all
// happen to produce upstream-compatible results through the generic
// literalPropValue → jsNull → Number(null) = 0 path:
//   - typeof x: upstream yields "number" / "string" / etc. → Number → NaN → skip;
//     ours yields 0 → 0 > 0 false → skip. Same observable outcome.
//   - void 0:   upstream yields undefined → NaN → skip; ours yields 0 → skip.
//   - x++:      upstream yields NaN (ToNumber on identifier) → skip; ours yields 0 → skip.
//
// Parentheses are stripped at the top so `<div tabIndex={(5)} />` and
// `<div tabIndex={((5))} />` route through the unwrapped literal. TS
// wrapper kinds (`as`, `!`, `satisfies`) are NOT stripped here because
// LITERAL_TYPES has no entry for them — upstream falls to noop → null
// → 0, and so do we via the default arm of literalPropValue.
func literalPropToTabIndexJsValue(inner *ast.Node, sourceText string) jsValue {
	inner = ast.SkipOuterExpressions(inner, ast.OEKParentheses)
	switch inner.Kind {
	case ast.KindArrayLiteralExpression:
		return arrayLiteralPropToTabIndexString(inner, sourceText)
	case ast.KindPrefixUnaryExpression:
		return unaryLiteralPropToTabIndexJsValue(inner.AsPrefixUnaryExpression())
	case ast.KindDeleteExpression:
		// jsx-ast-utils TYPES.UnaryExpression('delete', ...) returns `true`
		// — the runtime value of the operator on a deletable property. We
		// model it as boolean true so Number coerces to 1 and the rule
		// fires.
		return jsValue{Kind: jvBool, Bool: true}
	case ast.KindNoSubstitutionTemplateLiteral:
		// jsx-ast-utils' TemplateLiteral extractor reads `quasi.value.raw`,
		// NOT `cooked`. tsgo's NoSubstitutionTemplateLiteral only carries
		// the cooked `.Text` (no RawText field — the parser factory
		// constructor doesn't accept one). We synthesize raw from the
		// source slice between backticks. So `\`\t1\t\`` source-form is
		// `\t1\t` (6 chars, backslashes preserved) → Number = NaN → skip,
		// while the cooked Text is `	 1 	` → trim → 1 → would
		// false-positive without this raw extraction.
		return jsValue{Kind: jvString, Str: rawTemplateLiteralText(inner, sourceText)}
	case ast.KindTemplateExpression:
		// Same raw-text consideration applies to TemplateExpression with
		// substitutions. Head/Middle/Tail do carry RawText (parser
		// `getTemplateLiteralRawText` populates it), but for symmetry and
		// because Pos/End slicing is uniform across kinds, we route
		// everything through source-slice extraction.
		return templateExpressionRawText(inner.AsTemplateExpression(), sourceText)
	}
	return literalPropValue(inner)
}

// rawTemplateLiteralText extracts the raw source text between the backticks
// of a NoSubstitutionTemplateLiteral. tsgo's NoSubstitutionTemplateLiteral
// has no RawText field (the parser only populates cooked `.Text`), so we
// synthesize raw from the source slice. `node.Pos()` points at the leading
// backtick and `node.End()` at the byte after the trailing backtick — strip
// one byte on each end. Returns "" defensively if the slice would be empty
// or the source text is missing.
func rawTemplateLiteralText(node *ast.Node, sourceText string) string {
	pos := node.Pos()
	end := node.End()
	if sourceText == "" || pos+1 >= end-1 {
		return ""
	}
	if end-1 > len(sourceText) || pos+1 < 0 {
		return ""
	}
	return sourceText[pos+1 : end-1]
}

// templateExpressionRawText mirrors staticEvalTemplate but assembles the
// raw text from source slices for each quasi (Head/Middle/Tail), aligning
// with jsx-ast-utils' `quasi.value.raw` access pattern. Used only by the
// tabindex-no-positive `getLiteralPropValue` path — other a11y rules go
// through the shared staticEvalTemplate (which uses cooked `.Text`) since
// they only care about truthy / empty distinctions.
//
// Source slice boundaries:
//
//   - TemplateHead spans `\` … ${`, i.e. trim 1 byte from start (opening
//     backtick) and 2 bytes from end (`${`).
//   - TemplateMiddle spans `} … ${`, trim 1 + 2.
//   - TemplateTail spans `} … \``, trim 1 + 1.
func templateExpressionRawText(tpl *ast.TemplateExpression, sourceText string) jsValue {
	var sb strings.Builder
	if tpl.Head != nil {
		sb.WriteString(rawTemplateSegmentText(tpl.Head, sourceText, 1, 2))
	}
	if tpl.TemplateSpans != nil {
		for _, span := range tpl.TemplateSpans.Nodes {
			if span.Kind != ast.KindTemplateSpan {
				continue
			}
			sp := span.AsTemplateSpan()
			if sp.Expression != nil {
				// Substitution placeholder synthesis is identical to the
				// shared staticEvalTemplate path — escape-insensitive
				// because placeholders are either Identifier names or
				// ESTree-style type-name strings.
				expr := ast.SkipOuterExpressions(sp.Expression, ast.OEKParentheses)
				sb.WriteString(extractTemplateSubstitutionPlaceholder(expr))
			}
			if sp.Literal != nil {
				switch sp.Literal.Kind {
				case ast.KindTemplateMiddle:
					sb.WriteString(rawTemplateSegmentText(sp.Literal, sourceText, 1, 2))
				case ast.KindTemplateTail:
					sb.WriteString(rawTemplateSegmentText(sp.Literal, sourceText, 1, 1))
				}
			}
		}
	}
	return jsValue{Kind: jvString, Str: sb.String()}
}

// rawTemplateSegmentText extracts raw bytes from the source for a
// TemplateHead/Middle/Tail node, trimming `headTrim` bytes from the start
// and `tailTrim` from the end (the surrounding `\`` / `${` / `}` syntax).
// Returns "" defensively when the slice would be empty or out of range.
func rawTemplateSegmentText(node *ast.Node, sourceText string, headTrim, tailTrim int) string {
	pos := node.Pos() + headTrim
	end := node.End() - tailTrim
	if sourceText == "" || pos >= end || end > len(sourceText) || pos < 0 {
		return ""
	}
	return sourceText[pos:end]
}

// arrayLiteralPropToTabIndexString models jsx-ast-utils'
// extractValueFromArrayExpression + JS `[...].toString()` (Array.prototype.join
// with default ","). Element extraction goes through staticEval (TYPES path,
// matching upstream's `import extractValue from '../expressions'`), with the
// `Array.prototype.join` special-case where null / undefined elements
// stringify to "" rather than "null" / "undefined".
//
// Returns the joined string as a jsValue so the downstream jsValueToNumber
// applies the same StringToNumber coercion used everywhere else.
func arrayLiteralPropToTabIndexString(node *ast.Node, sourceText string) jsValue {
	arr := node.AsArrayLiteralExpression()
	var elements []*ast.Node
	if arr != nil && arr.Elements != nil {
		elements = arr.Elements.Nodes
	}
	if len(elements) == 0 {
		// `[].toString() === ""` → Number("") = 0.
		return jsValue{Kind: jvString, Str: ""}
	}
	parts := make([]string, 0, len(elements))
	for _, el := range elements {
		if el == nil {
			// Sparse elements `[,5]` → "" per Array.join spec.
			parts = append(parts, "")
			continue
		}
		parts = append(parts, arrayElementToString(el, sourceText))
	}
	return jsValue{Kind: jvString, Str: strings.Join(parts, ",")}
}

// arrayElementToString models the Array.join element-stringification step:
// take an array element node, walk it through the jsx-ast-utils TYPES path
// (the full extractor — not LITERAL_TYPES), then apply Array.prototype.join's
// null/undefined→"" special-case before ToString.
//
// Routes by element kind:
//
//   - Nested ArrayLiteralExpression: recurse so `[[5]].toString() === "5"`.
//     staticEval would only yield the coarse jvTruthy sentinel for arrays,
//     losing the join-able value.
//
//   - PrefixUnaryExpression (`+x`/`-x`/`~x`/`!x`): use the TYPES.UnaryExpression
//     ToNumber-on-operand protocol via unaryLiteralPropToTabIndexJsValue.
//     staticEvalUnary only computes `+`/`-`/`~` when the operand is already
//     jvNumber — so `[+true]` and `[+"5"]` would otherwise fall through to
//     jsNull and stringify to "" (skipping the rule when upstream reports).
//
// All other kinds go through staticEval (TYPES path), where the existing
// extractor handles ObjectExpression / RegExp / JsxElement (as jvTruthy →
// "" → NaN → upstream-compatible skip), Identifier (returns the name string
// per JS_RESERVED), DeleteExpression (returns true), etc. The
// null/undefined/unknown arm preserves Array.join's spec-mandated empty
// stringification.
func arrayElementToString(el *ast.Node, sourceText string) string {
	inner := ast.SkipOuterExpressions(el, ast.OEKParentheses)
	switch inner.Kind {
	case ast.KindArrayLiteralExpression:
		// jsx-ast-utils' extractValueFromArrayExpression recurses through
		// TYPES on each element — for nested arrays this loops back into
		// the same extractor. We mirror via a self-recurse here.
		return arrayLiteralPropToTabIndexString(inner, sourceText).Str
	case ast.KindPrefixUnaryExpression:
		ev := unaryLiteralPropToTabIndexJsValue(inner.AsPrefixUnaryExpression())
		switch ev.Kind {
		case jvNull, jvUndef, jvUnknown:
			return ""
		}
		return jsToString(ev)
	case ast.KindNoSubstitutionTemplateLiteral:
		// Same raw-text concern as the top-level template arm — use the
		// source slice between backticks, not the cooked text.
		return rawTemplateLiteralText(inner, sourceText)
	case ast.KindTemplateExpression:
		return templateExpressionRawText(inner.AsTemplateExpression(), sourceText).Str
	}
	ev := staticEval(el)
	switch ev.Kind {
	case jvNull, jvUndef, jvUnknown:
		// Array.prototype.join special-cases null/undefined to "" (NOT
		// "null"/"undefined"). Statically-unresolvable values fall through
		// the same way — upstream's TYPES path returns null for unknowns
		// and Array.join then writes "".
		return ""
	}
	return jsToString(ev)
}

// unaryLiteralPropToTabIndexJsValue mirrors jsx-ast-utils' UnaryExpression
// extractor (which LITERAL_TYPES inherits from TYPES): evaluate the operand
// via staticEval (TYPES path), then apply the JS unary operator.
//
// Unlike staticEvalUnary, this path applies ToNumber to the operand first
// for `+` / `-` / `~`, so `+true`, `~-2`, etc. resolve to numeric values
// instead of falling out to jsNull.
func unaryLiteralPropToTabIndexJsValue(un *ast.PrefixUnaryExpression) jsValue {
	operand := staticEval(un.Operand)
	switch un.Operator {
	case ast.KindExclamationToken:
		return jsValue{Kind: jvBool, Bool: !jsTruthy_(operand)}
	case ast.KindPlusToken:
		f, ok := jsValueToNumber(operand)
		if !ok {
			return jsValue{Kind: jvNumber, Num: math.NaN()}
		}
		return jsValue{Kind: jvNumber, Num: f}
	case ast.KindMinusToken:
		f, ok := jsValueToNumber(operand)
		if !ok {
			return jsValue{Kind: jvNumber, Num: math.NaN()}
		}
		return jsValue{Kind: jvNumber, Num: -f}
	case ast.KindTildeToken:
		f, ok := jsValueToNumber(operand)
		if !ok {
			return jsValue{Kind: jvNumber, Num: math.NaN()}
		}
		// JS ToInt32 → bitwise NOT → back to float64.
		return jsValue{Kind: jvNumber, Num: float64(^jsToInt32(f))}
	}
	// Unhandled unary token — upstream's extractor returns null.
	return jsNull
}

// jsValueToNumber implements JS ToNumber on the kinds produced by
// literalPropToTabIndexJsValue.
//
// Returns (0, false) when the result would be NaN (the rule's `isNaN(value)`
// guard skips, so we surface NaN via ok=false rather than returning a
// sentinel NaN value).
func jsValueToNumber(v jsValue) (float64, bool) {
	switch v.Kind {
	case jvNumber:
		if math.IsNaN(v.Num) {
			return 0, false
		}
		return v.Num, true
	case jvString:
		return tabIndexNoPositiveStringToNumber(v.Str)
	case jvBool:
		if v.Bool {
			return 1, true
		}
		return 0, true
	case jvBigInt:
		if v.Big == nil {
			return 0, true
		}
		// JS spec: `>` / `<` / `>=` / `<=` between a BigInt and a Number
		// coerces both via mathematical comparison without ToNumber (so
		// `2n > 0` is true), and `Number(bigint)` itself raises TypeError.
		// But upstream eslint-plugin-jsx-a11y calls JS `Number(...)` then
		// compares with `<= 0`, which for BigInt yields the float64 cast
		// (`Number(2n) === 2`). Mirror that — beyond-Float64-range BigInts
		// become ±Inf, which JS reports the same way (`Infinity > 0`).
		f, _ := new(big.Float).SetInt(v.Big).Float64()
		if math.IsNaN(f) {
			return 0, false
		}
		return f, true
	case jvNull:
		// jsx-ast-utils LITERAL_TYPES.Identifier (non-undefined) yields JS
		// null → Number(null) = 0. Note that the *null literal* `tabIndex={null}`
		// is routed through literalPropValue's KindNullKeyword arm, which
		// returns jvString "null" (LITERAL_TYPES.Literal special-case) and
		// lands on the jvString arm above with Number("null") = NaN.
		return 0, true
	case jvUndef:
		// Number(undefined) = NaN
		return 0, false
	}
	// jvFunction / jvTruthy / jvUnknown — upstream returns objects / arrays /
	// functions here, all of which ToPrimitive→ToNumber to NaN (with rare
	// exceptions like single-element arrays, but those are already routed
	// through the ArrayLiteralExpression arm of literalPropToTabIndexJsValue).
	return 0, false
}

// tabIndexNoPositiveStringToNumber implements ECMA-262 7.1.4.1.1 StringToNumber
// for the strings we encounter via getLiteralPropValue: trim whitespace,
// empty → 0, hex / oct / bin unsigned-integer prefixes (no signed-prefix
// support, per spec — `Number("-0x10")` is NaN), decimal otherwise. Returns
// (0, false) when the result is NaN.
//
// Note: signed decimals (`"+1"`, `"-5"`, `" -5 "`) work via ParseFloat
// natively. Signed hex/oct/bin are intentionally rejected to mirror JS.
func tabIndexNoPositiveStringToNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
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
}

// jsToInt32 implements ECMA-262 7.1.6 ToInt32: ToNumber → truncate → mod 2^32
// → wrap into signed-int32 range. NaN / ±Infinity all yield 0 per spec.
// Used by the `~` operator's pre-bitwise coercion.
func jsToInt32(f float64) int32 {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}
	n := math.Mod(math.Trunc(f), 4294967296.0)
	if n >= 2147483648.0 {
		n -= 4294967296.0
	} else if n < -2147483648.0 {
		n += 4294967296.0
	}
	return int32(n)
}
