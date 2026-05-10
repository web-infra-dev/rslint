package jsxa11yutil

// staticEval ports jsx-ast-utils' `getPropValue` semantics so that jsx-a11y
// rules can distinguish "is the empty string" from "is truthy" precisely —
// alt-text's `(altValue && !isNullValued) || altValue === ''` check makes
// this distinction load-bearing.
//
// Why not reuse tsgo's evaluator.NewEvaluator (shim/evaluator/shim.go)?
// That evaluator is built for const-enum / readonly literal-type evaluation:
// it covers numeric arithmetic, bitwise ops, `+` string concat and template
// expressions, but DELIBERATELY OMITS the operators alt-text depends on —
// `&&` / `||` / `??` short-circuits, ConditionalExpression, PrefixUnary `!`,
// boolean / null literals. It also routes through `jsnum.Number` and
// `jsnum.PseudoBigInt`, which are typescript-go-internal types not exposed
// via the shim. So a partial reuse would still need this file's logic;
// keeping a single integrated evaluator is cleaner.
//
// Helpers we DO reuse from rslint utils / tsgo shim:
//
//	utils.NormalizeNumericLiteral / NormalizeBigIntLiteral — text → value
//	utils.IsUndefinedIdentifier                            — undefined check
//	utils.GetStaticStringValue                             — used by callers
//	ast.SkipOuterExpressions(OEKParentheses|OEKAssertions) — wrapper unwrap
//	ast.IsUndefinedIdentifier (TODO if needed)             — N/A
//
// The semantics modeled here mirror jsx-ast-utils v3 — see
// https://github.com/jsx-eslint/jsx-ast-utils/tree/main/src/values for the
// extractors per node type. Verified empirically against ESLint via probe
// scripts; see /tmp/jsx-a11y-verify if you need to re-check.

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// jsValue is a statically-evaluated JavaScript value, mirroring the shape of
// jsx-ast-utils' `getPropValue` output. Returning the actual value (instead of
// just a truthy / falsy bit) is required because alt-text and similar a11y
// rules distinguish three states: truthy, falsy, AND empty-string, and the
// `=== ''` branch alone takes precedence over truthiness.
//
// Discriminated by Kind. Only the field corresponding to Kind is meaningful.
//
//	Kind == jvString   → Str
//	Kind == jvNumber   → Num (NaN is allowed)
//	Kind == jvBigInt   → Big
//	Kind == jvBool     → Bool
//	Kind == jvNull     → (no field)
//	Kind == jvUndef    → (no field)
//	Kind == jvFunction → (truthy sentinel; arrow / function literal / known global ctor)
//	Kind == jvTruthy   → (sentinel for "statically truthy but specific value not modelled")
//	Kind == jvUnknown  → (cannot statically determine; rules should default-truthy)
type jsValue struct {
	Kind jsKind
	Str  string
	Num  float64
	Big  *big.Int
	Bool bool
}

type jsKind int

const (
	jvUnknown jsKind = iota
	jvString
	jvNumber
	jvBigInt
	jvBool
	jvNull
	jvUndef
	jvFunction
	jvTruthy
)

var (
	jsNull    = jsValue{Kind: jvNull}
	jsUndef   = jsValue{Kind: jvUndef}
	jsUnknown = jsValue{Kind: jvUnknown}
	jsFn      = jsValue{Kind: jvFunction}
	jsTruthy  = jsValue{Kind: jvTruthy}
)

// staticEval mirrors jsx-ast-utils' `getValue` (the engine behind getPropValue).
// Walks the expression tree and computes the static JS value, applying real
// JS semantics for `&&` / `||` / `??` / ternary / arithmetic so that callers
// can distinguish "altValue === ''" from "altValue is truthy" precisely.
//
// Parentheses and TS assertion wrappers are unwrapped on every recursion.
//
// Returns jsUnknown when the expression cannot be statically evaluated (rare
// in practice — see CallExpression / MemberExpression below for the
// jsx-ast-utils synthetic-string fallback that keeps unknown calls "truthy").
//
// Mirrors the relevant jsx-ast-utils Identifier reserved-word handling:
//
//	JS_RESERVED = { Array, Date, Infinity, Math, Number, Object, String, undefined }
//
// where `Infinity` evaluates to +Infinity, `undefined` to undefined, and the
// global constructor identifiers (Array / Date / Math / Number / Object /
// String) evaluate to their function values (truthy).
func staticEval(node *ast.Node) jsValue {
	if node == nil {
		return jsUnknown
	}
	node = ast.SkipOuterExpressions(node, skipTransparent)
	switch node.Kind {
	case ast.KindStringLiteral:
		// jsx-ast-utils' Literal extractor normalizes case-insensitive
		// "true" / "false" strings to actual booleans. This is load-bearing —
		// `<img alt="false" />` upstream evaluates altValue to `false`
		// (boolean), making it falsy → altValue error. Without this
		// normalization the rule would incorrectly accept `alt="false"`.
		return jsxAstUtilsLiteralCoerce(node.AsStringLiteral().Text)
	case ast.KindNoSubstitutionTemplateLiteral:
		// `` `text` `` parses as a TemplateLiteral with no expressions in
		// ESTree, so jsx-ast-utils routes it through extractValueFromTemplateLiteral
		// — which simply joins the quasi text and returns the raw string.
		// Crucially, the "true" / "false" → boolean coercion in
		// extractValueFromLiteral applies ONLY to ESTree's `Literal` (string)
		// type, NOT to `TemplateLiteral`. So `` `false` `` evaluates to the
		// non-empty string "false" (truthy), not boolean false. Confirmed via
		// differential against eslint-plugin-jsx-a11y v6.10.2 with
		// `<div accessKey={`false`} />`: ESLint reports, rslint pre-fix did
		// not. Do NOT pipe through jsxAstUtilsLiteralCoerce here.
		return jsValue{Kind: jvString, Str: node.AsNoSubstitutionTemplateLiteral().Text}
	case ast.KindTrueKeyword:
		return jsValue{Kind: jvBool, Bool: true}
	case ast.KindFalseKeyword:
		return jsValue{Kind: jvBool, Bool: false}
	case ast.KindNullKeyword:
		return jsNull
	case ast.KindNumericLiteral:
		// utils.NormalizeNumericLiteral returns "Infinity" / "-Infinity" for
		// overflow and a normalized decimal otherwise — feed both through
		// strconv.ParseFloat; failures fall to NaN (still a Number).
		text := utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text)
		f, err := strconv.ParseFloat(text, 64)
		if err != nil && text == "Infinity" {
			f = math.Inf(1)
		} else if err != nil && text == "-Infinity" {
			f = math.Inf(-1)
		} else if err != nil {
			f = math.NaN()
		}
		return jsValue{Kind: jvNumber, Num: f}
	case ast.KindBigIntLiteral:
		text := utils.NormalizeBigIntLiteral(node.AsBigIntLiteral().Text)
		bi := new(big.Int)
		if _, ok := bi.SetString(text, 10); !ok {
			return jsUnknown
		}
		return jsValue{Kind: jvBigInt, Big: bi}
	case ast.KindIdentifier:
		name := node.AsIdentifier().Text
		switch name {
		case "undefined":
			return jsUndef
		case "Infinity":
			return jsValue{Kind: jvNumber, Num: math.Inf(1)}
		case "Array", "Date", "Math", "Number", "Object", "String":
			return jsFn // upstream returns the actual constructor; we only need "truthy"
		}
		// Other identifiers — upstream returns the bare name string. We mirror
		// that so consumers see a non-empty truthy string.
		return jsValue{Kind: jvString, Str: name}
	case ast.KindThisKeyword:
		// Upstream's ThisExpression extractor returns the magic string "this"
		// — non-empty truthy.
		return jsValue{Kind: jvString, Str: "this"}
	case ast.KindBinaryExpression:
		return staticEvalBinary(node.AsBinaryExpression())
	case ast.KindConditionalExpression:
		ce := node.AsConditionalExpression()
		test := staticEval(ce.Condition)
		if jsTruthy_(test) {
			return staticEval(ce.WhenTrue)
		}
		return staticEval(ce.WhenFalse)
	case ast.KindPrefixUnaryExpression:
		return staticEvalUnary(node.AsPrefixUnaryExpression())
	case ast.KindTemplateExpression:
		return staticEvalTemplate(node.AsTemplateExpression())
	case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression:
		// Upstream FunctionExpression / ArrowFunctionExpression: returns a
		// function (truthy). ClassExpression has no upstream extractor →
		// null → falsy; but realistically `<img alt={class C{}}>` is never
		// a meaningful alt and we treat class literals like function
		// literals (truthy) for safer behavior. Documented divergence.
		return jsFn
	case ast.KindCallExpression:
		// Upstream's CallExpression / OptionalCallExpression extractor
		// synthesizes a non-empty string `${callee}${optional?"?.":""}(${args})`.
		// We model it as a jvString (typeof "string") rather than the
		// jvTruthy sentinel because rules like iframe-has-title gate on
		// `typeof === 'string'` and need to distinguish synthesized-string
		// shapes (Call / Member / JSX / TaggedTemplate) from genuine
		// non-string truthy values (Object / Array / New / RegExp).
		// The exact text isn't observed by any rule — only its non-emptiness
		// — so a stable placeholder suffices.
		return jsValue{Kind: jvString, Str: "(call)"}
	case ast.KindNewExpression:
		// Upstream NewExpression returns `new Object()` → empty object →
		// truthy but typeof "object". Keep as jvTruthy sentinel so it
		// classifies as truthy without tripping the typeof-string gate.
		return jsTruthy
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// Upstream's MemberExpression / OptionalMemberExpression extractor
		// synthesizes a non-empty string ("obj.prop" / "obj?.prop" /
		// "obj[key]"). Modeled as jvString (typeof "string") for the same
		// reason as CallExpression above. Optional access (`obj?.prop`)
		// is the same kind in tsgo with a flag — both forms upstream
		// return strings.
		return jsValue{Kind: jvString, Str: "(member)"}
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		// Upstream returns the actual object/array → typeof "object".
		// Truthy but not string.
		return jsTruthy
	case ast.KindRegularExpressionLiteral:
		// Upstream's Literal extractor returns the RegExp object → typeof
		// "object". Truthy but not string.
		return jsTruthy
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		// Upstream's JSXElement / JSXFragment extractor synthesizes a
		// non-empty string like "<Tag />" / "<></>". Modeled as jvString
		// so iframe-has-title's typeof-string check passes for
		// `<iframe title={<X />} />`-style inputs.
		return jsValue{Kind: jvString, Str: "(jsx)"}
	case ast.KindTaggedTemplateExpression:
		// Upstream's TaggedTemplateExpression extractor delegates to
		// TemplateLiteral on the inner quasi. Recurse into the inner
		// template to preserve the exact extracted text — including the
		// (rare) empty-tag-empty-template `` tag`` `` case which upstream
		// returns as "" and we must report.
		tt := node.AsTaggedTemplateExpression()
		if tt == nil || tt.Template == nil {
			return jsValue{Kind: jvString, Str: ""}
		}
		return staticEval(tt.Template)
	case ast.KindPostfixUnaryExpression:
		// `x++` / `x--` — UpdateExpression. Upstream computes `val++` /
		// `val--` after coercing the operand to a number; for non-numeric
		// operands (the common case `x++` where x is an identifier) the
		// result is NaN → falsy. We mirror that conservatively.
		return staticEvalUpdate(node.AsPostfixUnaryExpression().Operator,
			node.AsPostfixUnaryExpression().Operand)
	case ast.KindAwaitExpression, ast.KindYieldExpression:
		// Upstream has no extractor for these → console.error + return
		// null → falsy. Mirror the falsy classification.
		return jsNull
	case ast.KindTypeOfExpression, ast.KindVoidExpression:
		// Upstream UnaryExpression: typeof / void → undefined → falsy.
		return jsUndef
	case ast.KindDeleteExpression:
		// Upstream UnaryExpression: delete → true → truthy.
		return jsValue{Kind: jvBool, Bool: true}
	}
	// Default for anything we don't model: align with upstream's "unknown
	// type" path, which logs a console.error and returns null. A null falsy
	// fallback is safer than truthy because alt-text's `(altValue && …) ||
	// altValue === ''` check needs a definitive falsy on unrecognized kinds
	// to match upstream rejections.
	return jsNull
}

// staticEvalUpdate handles `x++` / `++x` / `x--` / `--x` for both Postfix
// and Prefix unary nodes. Mirrors jsx-ast-utils UpdateExpression.js: coerces
// operand to a number (NaN if non-numeric), then increments / decrements.
// Returns the result; for the prefix case the new value, for postfix the
// pre-update value (NaN in both branches when operand is non-numeric).
func staticEvalUpdate(op ast.Kind, operand *ast.Node) jsValue {
	v := staticEval(operand)
	if v.Kind != jvNumber {
		// JS coerces operand to number for ++/--; non-numeric coerces to
		// NaN, and the result is NaN (falsy).
		return jsValue{Kind: jvNumber, Num: math.NaN()}
	}
	switch op {
	case ast.KindPlusPlusToken:
		return jsValue{Kind: jvNumber, Num: v.Num + 1}
	case ast.KindMinusMinusToken:
		return jsValue{Kind: jvNumber, Num: v.Num - 1}
	}
	return jsUndef
}

// staticEvalBinary handles the operators alt-text actually depends on
// (`&&` / `||` / `??` short-circuits, plus `+` for string concat). Other
// operators (`-`, `*`, comparisons, …) are computed when both sides are
// numbers, otherwise fall back to jsUnknown.
func staticEvalBinary(bin *ast.BinaryExpression) jsValue {
	if bin.OperatorToken == nil {
		return jsUnknown
	}
	op := bin.OperatorToken.Kind
	left := staticEval(bin.Left)
	switch op {
	case ast.KindBarBarToken:
		// `a || b` → a if a truthy, else b
		if jsTruthy_(left) {
			return left
		}
		return staticEval(bin.Right)
	case ast.KindAmpersandAmpersandToken:
		// `a && b` → a if a falsy, else b
		if !jsTruthy_(left) {
			return left
		}
		return staticEval(bin.Right)
	case ast.KindQuestionQuestionToken:
		// `a ?? b` → a if a is neither null nor undefined, else b
		if left.Kind == jvNull || left.Kind == jvUndef {
			return staticEval(bin.Right)
		}
		return left
	case ast.KindCommaToken:
		// In ESTree this becomes SequenceExpression, whose extractor
		// returns an array of all values (truthy regardless of contents).
		// Mirror that — `(a, b, false)` is upstream-truthy, not the
		// rightmost-operand falsy.
		return jsTruthy
	}
	right := staticEval(bin.Right)
	switch op {
	case ast.KindPlusToken:
		// String concatenation has priority when EITHER side is a string —
		// matches JS's `+` rules for our purposes.
		if left.Kind == jvString || right.Kind == jvString {
			return jsValue{Kind: jvString, Str: jsToString(left) + jsToString(right)}
		}
		if left.Kind == jvNumber && right.Kind == jvNumber {
			return jsValue{Kind: jvNumber, Num: left.Num + right.Num}
		}
	case ast.KindMinusToken:
		if left.Kind == jvNumber && right.Kind == jvNumber {
			return jsValue{Kind: jvNumber, Num: left.Num - right.Num}
		}
	case ast.KindAsteriskToken:
		if left.Kind == jvNumber && right.Kind == jvNumber {
			return jsValue{Kind: jvNumber, Num: left.Num * right.Num}
		}
	case ast.KindSlashToken:
		if left.Kind == jvNumber && right.Kind == jvNumber {
			return jsValue{Kind: jvNumber, Num: left.Num / right.Num}
		}
	case ast.KindPercentToken:
		if left.Kind == jvNumber && right.Kind == jvNumber {
			return jsValue{Kind: jvNumber, Num: math.Mod(left.Num, right.Num)}
		}
	case ast.KindAsteriskAsteriskToken:
		if left.Kind == jvNumber && right.Kind == jvNumber {
			return jsValue{Kind: jvNumber, Num: math.Pow(left.Num, right.Num)}
		}
	case ast.KindLessThanToken:
		if l, r, ok := numericBoth(left, right); ok {
			return jsValue{Kind: jvBool, Bool: l < r}
		}
		if left.Kind == jvString && right.Kind == jvString {
			return jsValue{Kind: jvBool, Bool: left.Str < right.Str}
		}
	case ast.KindLessThanEqualsToken:
		if l, r, ok := numericBoth(left, right); ok {
			return jsValue{Kind: jvBool, Bool: l <= r}
		}
		if left.Kind == jvString && right.Kind == jvString {
			return jsValue{Kind: jvBool, Bool: left.Str <= right.Str}
		}
	case ast.KindGreaterThanToken:
		if l, r, ok := numericBoth(left, right); ok {
			return jsValue{Kind: jvBool, Bool: l > r}
		}
		if left.Kind == jvString && right.Kind == jvString {
			return jsValue{Kind: jvBool, Bool: left.Str > right.Str}
		}
	case ast.KindGreaterThanEqualsToken:
		if l, r, ok := numericBoth(left, right); ok {
			return jsValue{Kind: jvBool, Bool: l >= r}
		}
		if left.Kind == jvString && right.Kind == jvString {
			return jsValue{Kind: jvBool, Bool: left.Str >= right.Str}
		}
	case ast.KindEqualsEqualsToken:
		// JS `==` does loose equality (with type coercion). We only model the
		// strict-equal case where both operands have the same kind — looser
		// coercions (`1 == "1"`) fall through to unknown, since modeling the
		// full coercion table isn't worth the complexity for alt-text.
		return jsValue{Kind: jvBool, Bool: jsLooseEquals(left, right)}
	case ast.KindEqualsEqualsEqualsToken:
		return jsValue{Kind: jvBool, Bool: jsStrictEquals(left, right)}
	case ast.KindExclamationEqualsToken:
		return jsValue{Kind: jvBool, Bool: !jsLooseEquals(left, right)}
	case ast.KindExclamationEqualsEqualsToken:
		return jsValue{Kind: jvBool, Bool: !jsStrictEquals(left, right)}
	case ast.KindBarToken:
		if l, r, ok := numericBoth(left, right); ok {
			return jsValue{Kind: jvNumber, Num: float64(int32(l) | int32(r))}
		}
	case ast.KindAmpersandToken:
		if l, r, ok := numericBoth(left, right); ok {
			return jsValue{Kind: jvNumber, Num: float64(int32(l) & int32(r))}
		}
	case ast.KindCaretToken:
		if l, r, ok := numericBoth(left, right); ok {
			return jsValue{Kind: jvNumber, Num: float64(int32(l) ^ int32(r))}
		}
	case ast.KindLessThanLessThanToken:
		if l, r, ok := numericBoth(left, right); ok {
			return jsValue{Kind: jvNumber, Num: float64(int32(l) << (uint32(r) & 0x1F))}
		}
	case ast.KindGreaterThanGreaterThanToken:
		if l, r, ok := numericBoth(left, right); ok {
			return jsValue{Kind: jvNumber, Num: float64(int32(l) >> (uint32(r) & 0x1F))}
		}
	case ast.KindGreaterThanGreaterThanGreaterThanToken:
		if l, r, ok := numericBoth(left, right); ok {
			return jsValue{Kind: jvNumber, Num: float64(uint32(l) >> (uint32(r) & 0x1F))}
		}
	}
	// Assignment expressions (`=`, `+=`, `||=`, etc.). Upstream's
	// AssignmentExpression extractor returns `${left} ${op} ${right}` — a
	// non-empty string, hence truthy regardless of contents.
	switch op {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindQuestionQuestionEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
		return jsTruthy
	case ast.KindInKeyword, ast.KindInstanceOfKeyword:
		// Upstream computes these at the runtime level when both sides are
		// known. We can't know statically; conservatively report falsy
		// (matches upstream's fallback when the runtime check would throw).
		return jsNull
	}
	// Truly unknown operators — align with upstream's null fallback.
	return jsNull
}

// numericBoth coerces left/right to float64 if both are jvNumber. Returns
// (0, 0, false) otherwise. Doesn't try to coerce strings → numbers because
// alt-text never depends on those cases and the JS coercion rules
// (`"3" - 1`, `null * 2`) are subtle enough that being conservative is
// better than being slightly wrong.
func numericBoth(l, r jsValue) (float64, float64, bool) {
	if l.Kind == jvNumber && r.Kind == jvNumber {
		return l.Num, r.Num, true
	}
	return 0, 0, false
}

// jsStrictEquals models JS `===` for the kinds we statically evaluate.
// Different kinds → not strictly equal. Same kind → compare values, with NaN
// being not-equal-to-NaN per JS spec.
func jsStrictEquals(l, r jsValue) bool {
	if l.Kind == jvUnknown || r.Kind == jvUnknown ||
		l.Kind == jvTruthy || r.Kind == jvTruthy ||
		l.Kind == jvFunction || r.Kind == jvFunction {
		return false // can't compare unknowns
	}
	if l.Kind != r.Kind {
		return false
	}
	switch l.Kind {
	case jvString:
		return l.Str == r.Str
	case jvNumber:
		// NaN !== NaN per spec.
		if math.IsNaN(l.Num) || math.IsNaN(r.Num) {
			return false
		}
		return l.Num == r.Num
	case jvBigInt:
		if l.Big == nil || r.Big == nil {
			return l.Big == r.Big
		}
		return l.Big.Cmp(r.Big) == 0
	case jvBool:
		return l.Bool == r.Bool
	case jvNull, jvUndef:
		return true
	}
	return false
}

// jsLooseEquals models JS `==`. We only handle the two "easy" cases:
//   - strict-equal kinds → reuse jsStrictEquals
//   - null == undefined / undefined == null → true
//
// Full type coercion (`1 == "1"`, `null == 0`) is not modeled — callers see
// `false` for those, which is a conservative approximation. Alt-text never
// depends on `==` semantics; this keeps the surface small.
func jsLooseEquals(l, r jsValue) bool {
	if (l.Kind == jvNull && r.Kind == jvUndef) || (l.Kind == jvUndef && r.Kind == jvNull) {
		return true
	}
	return jsStrictEquals(l, r)
}

// staticEvalUnary mirrors jsx-ast-utils' UnaryExpression / UpdateExpression
// extractors for the operators that appear as PrefixUnary in tsgo:
//
//	`!x`       → !truthy(x)
//	`-x`       → -number(x)  (number-coerced)
//	`+x`       → +number(x)
//	`~x`       → ~int32(x)
//	`++x`      → ToNumber(x) + 1 (NaN if non-numeric)
//	`--x`      → ToNumber(x) - 1 (NaN if non-numeric)
//
// Note: `typeof x`, `void x`, `delete x` are NOT PrefixUnary in tsgo —
// they're separate kinds (KindTypeOfExpression / KindVoidExpression /
// KindDeleteExpression) handled in staticEval directly.
func staticEvalUnary(un *ast.PrefixUnaryExpression) jsValue {
	switch un.Operator {
	case ast.KindExclamationToken:
		return jsValue{Kind: jvBool, Bool: !jsTruthy_(staticEval(un.Operand))}
	case ast.KindMinusToken:
		v := staticEval(un.Operand)
		if v.Kind == jvNumber {
			return jsValue{Kind: jvNumber, Num: -v.Num}
		}
	case ast.KindPlusToken:
		v := staticEval(un.Operand)
		if v.Kind == jvNumber {
			return v
		}
	case ast.KindTildeToken:
		v := staticEval(un.Operand)
		if v.Kind == jvNumber {
			return jsValue{Kind: jvNumber, Num: float64(^int32(v.Num))}
		}
	case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
		return staticEvalUpdate(un.Operator, un.Operand)
	}
	return jsNull
}

// staticEvalTemplate mirrors jsx-ast-utils' TemplateLiteral.js extractor
// byte-for-byte. The upstream logic is, per substitution node:
//
//	type === 'TemplateElement'                → raw text
//	type === 'Identifier', name === 'undefined' → "undefined" (literal)
//	type === 'Identifier'                     → "{name}" (single curly)
//	type.indexOf('Expression') > -1           → "{TypeName}" (single curly)
//	otherwise                                 → ""  (load-bearing!)
//
// The `otherwise → ""` branch is the load-bearing quirk: jsx-ast-utils does
// NOT recursively extract substitution values. Anything whose ESTree type
// name is `Literal` / `JSXElement` / `JSXFragment` / `TemplateLiteral` /
// `Super` / `SpreadElement` / `MetaProperty` / `TSTypeAssertion` etc.
// contributes the empty string. So `\`${"en"}\`` → `""` (falsy), NOT `"en"`.
//
// We previously used `${name}` / `${Expression}` (with a `$` prefix) and
// always returned a truthy non-empty string for non-Identifier substitutions.
// That diverged from upstream for templates with literal-only substitutions,
// flipping `<html lang={\`${"en"}\`} />` from REPORT to no-report. This
// implementation uses single-curly wrapping and returns `""` for the
// non-`*Expression` branch to match upstream exactly.
func staticEvalTemplate(tpl *ast.TemplateExpression) jsValue {
	var sb strings.Builder
	if tpl.Head != nil {
		sb.WriteString(tpl.Head.AsTemplateHead().Text)
	}
	if tpl.TemplateSpans != nil {
		for _, span := range tpl.TemplateSpans.Nodes {
			if span.Kind != ast.KindTemplateSpan {
				continue
			}
			sp := span.AsTemplateSpan()
			if sp.Expression != nil {
				// ESTree flattens parens at parse time, so a parenthesized
				// substitution `${(x)}` reaches TemplateLiteral.js with the
				// inner expression directly. Strip parens here to match.
				// TS wrappers (TSAsExpression / TSNonNullExpression /
				// TSSatisfiesExpression / TSTypeAssertion) are PRESERVED —
				// they show up in ESTree as their own nodes and upstream
				// classifies each by its `.type` name string.
				expr := ast.SkipOuterExpressions(sp.Expression, ast.OEKParentheses)
				sb.WriteString(extractTemplateSubstitutionPlaceholder(expr))
			}
			if sp.Literal != nil {
				switch sp.Literal.Kind {
				case ast.KindTemplateMiddle:
					sb.WriteString(sp.Literal.AsTemplateMiddle().Text)
				case ast.KindTemplateTail:
					sb.WriteString(sp.Literal.AsTemplateTail().Text)
				}
			}
		}
	}
	return jsValue{Kind: jvString, Str: sb.String()}
}

// extractTemplateSubstitutionPlaceholder mirrors the inline substitution
// classification in jsx-ast-utils' TemplateLiteral.js. See
// staticEvalTemplate for the upstream algorithm and the load-bearing
// `otherwise → ""` quirk.
func extractTemplateSubstitutionPlaceholder(expr *ast.Node) string {
	if expr == nil {
		return ""
	}
	if expr.Kind == ast.KindIdentifier {
		name := expr.AsIdentifier().Text
		if name == "undefined" {
			// Upstream: `name === 'undefined' ? name : ...` → bare string.
			return "undefined"
		}
		return "{" + name + "}"
	}
	if t := tsgoKindToESTreeExpressionTypeName(expr.Kind); t != "" {
		return "{" + t + "}"
	}
	return ""
}

// tsgoKindToESTreeExpressionTypeName maps tsgo Kinds to their ESTree
// type-name string, but ONLY for kinds whose ESTree type name contains the
// substring "Expression" (the only branch in TemplateLiteral.js that
// produces a non-empty placeholder for non-Identifier substitutions).
//
// Returns "" for kinds whose ESTree type does NOT contain "Expression" —
// `Literal` (StringLiteral / NumericLiteral / BigIntLiteral / true / false /
// null / RegExp), `JSXElement` / `JSXFragment`, `TemplateLiteral`, `Super`,
// `MetaProperty`, `SpreadElement`, `TSTypeAssertion`, etc. Those callers
// receive `""` from extractTemplateSubstitutionPlaceholder.
func tsgoKindToESTreeExpressionTypeName(k ast.Kind) string {
	switch k {
	case ast.KindBinaryExpression:
		// Covers ESTree's BinaryExpression / LogicalExpression /
		// AssignmentExpression / SequenceExpression — all contain
		// "Expression". Upstream's TemplateLiteral.js doesn't differentiate
		// by operator; it just reads `.type`. Returning "BinaryExpression"
		// preserves the truthy-non-empty placeholder regardless of operator.
		return "BinaryExpression"
	case ast.KindCallExpression:
		return "CallExpression"
	case ast.KindNewExpression:
		return "NewExpression"
	case ast.KindConditionalExpression:
		return "ConditionalExpression"
	case ast.KindArrowFunction:
		return "ArrowFunctionExpression"
	case ast.KindFunctionExpression:
		return "FunctionExpression"
	case ast.KindClassExpression:
		return "ClassExpression"
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return "MemberExpression"
	case ast.KindObjectLiteralExpression:
		return "ObjectExpression"
	case ast.KindArrayLiteralExpression:
		return "ArrayExpression"
	case ast.KindTaggedTemplateExpression:
		return "TaggedTemplateExpression"
	case ast.KindThisKeyword:
		return "ThisExpression"
	case ast.KindPrefixUnaryExpression,
		ast.KindTypeOfExpression,
		ast.KindVoidExpression,
		ast.KindDeleteExpression:
		return "UnaryExpression"
	case ast.KindPostfixUnaryExpression:
		return "UpdateExpression"
	case ast.KindAwaitExpression:
		return "AwaitExpression"
	case ast.KindYieldExpression:
		return "YieldExpression"
	case ast.KindAsExpression:
		return "TSAsExpression"
	case ast.KindNonNullExpression:
		return "TSNonNullExpression"
	case ast.KindSatisfiesExpression:
		return "TSSatisfiesExpression"
	}
	return ""
}

// jsTruthy_ replicates JS truthiness on a jsValue.
func jsTruthy_(v jsValue) bool {
	switch v.Kind {
	case jvString:
		return v.Str != ""
	case jvNumber:
		return v.Num != 0 && !math.IsNaN(v.Num)
	case jvBigInt:
		return v.Big != nil && v.Big.Sign() != 0
	case jvBool:
		return v.Bool
	case jvNull, jvUndef:
		return false
	case jvFunction, jvTruthy:
		return true
	}
	// Unknown — default to truthy ("the developer wrote SOMETHING").
	return true
}

// jsToString approximates JS's String() coercion for the few kinds we produce.
// Used by the `+` operator's string-concat path.
func jsToString(v jsValue) string {
	switch v.Kind {
	case jvString:
		return v.Str
	case jvNumber:
		if math.IsNaN(v.Num) {
			return "NaN"
		}
		if math.IsInf(v.Num, 1) {
			return "Infinity"
		}
		if math.IsInf(v.Num, -1) {
			return "-Infinity"
		}
		return strconv.FormatFloat(v.Num, 'f', -1, 64)
	case jvBigInt:
		if v.Big == nil {
			return "0"
		}
		return v.Big.String()
	case jvBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case jvNull:
		return "null"
	case jvUndef:
		return "undefined"
	}
	return ""
}

// jsValueIsExactlyEmptyString reports whether the value is statically known
// to be the empty string. Used for the `altValue === ''` branch of the
// alt-text validity check.
func jsValueIsExactlyEmptyString(v jsValue) bool {
	return v.Kind == jvString && v.Str == ""
}

// literalPropValue mirrors jsx-ast-utils' `getLiteralPropValue` engine —
// LITERAL_TYPES — which is `getPropValue` minus the runtime-expression
// extractors (CallExpression, MemberExpression, etc. all become noop → null).
// Used by `<object title>` and `isPresentationRole(role)` since upstream
// reads those via `getLiteralPropValue`, not `getPropValue`.
//
// Critical differences vs staticEval:
//
//   - Identifier (non-undefined) returns jvNull instead of the name string.
//     Reason: upstream LITERAL_TYPES.Identifier returns `null` for any
//     non-undefined identifier — runtime variables aren't literal values.
//   - null literal returns the string "null" (truthy!). This is upstream's
//     special-case in LITERAL_TYPES.Literal: `extractedVal === null ?
//     'null' : extractedVal`.
//   - Most Expression kinds (Call, Member, Conditional, Logical, Binary,
//     Unary except `!`, …) are noop → jvNull → falsy. We model the few
//     that upstream does compute: TemplateLiteral, ArrayExpression filter,
//     and Literal coercion.
func literalPropValue(node *ast.Node) jsValue {
	if node == nil {
		return jsUnknown
	}
	// Only strip parentheses — TS-wrapper kinds (TSAsExpression /
	// TSNonNullExpression / TypeCastExpression / SatisfiesExpression) MUST
	// fall through to the default branch below and return jvNull, mirroring
	// jsx-ast-utils' LITERAL_TYPES which maps each of those to `noop` → null.
	// staticEval is the engine for the getPropValue path; this one is for the
	// getLiteralPropValue path, which deliberately rejects TS wrappers.
	node = ast.SkipOuterExpressions(node, ast.OEKParentheses)
	switch node.Kind {
	case ast.KindStringLiteral:
		return jsxAstUtilsLiteralCoerce(node.AsStringLiteral().Text)
	case ast.KindNoSubstitutionTemplateLiteral:
		// See the staticEval comment for the same case: jsx-ast-utils treats
		// `` `text` `` as a TemplateLiteral and does NOT apply the
		// "true"/"false" → boolean coercion that the Literal extractor uses.
		return jsValue{Kind: jvString, Str: node.AsNoSubstitutionTemplateLiteral().Text}
	case ast.KindTrueKeyword:
		return jsValue{Kind: jvBool, Bool: true}
	case ast.KindFalseKeyword:
		return jsValue{Kind: jvBool, Bool: false}
	case ast.KindNullKeyword:
		// Upstream LITERAL_TYPES.Literal: null → 'null' (truthy string).
		// This is intentional and load-bearing for `<object title={null} />`
		// being valid.
		return jsValue{Kind: jvString, Str: "null"}
	case ast.KindNumericLiteral:
		text := utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text)
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			f = math.NaN()
		}
		return jsValue{Kind: jvNumber, Num: f}
	case ast.KindBigIntLiteral:
		text := utils.NormalizeBigIntLiteral(node.AsBigIntLiteral().Text)
		bi := new(big.Int)
		if _, ok := bi.SetString(text, 10); !ok {
			return jsUnknown
		}
		return jsValue{Kind: jvBigInt, Big: bi}
	case ast.KindIdentifier:
		// LITERAL_TYPES.Identifier: undefined → undefined; everything else → null.
		if utils.IsUndefinedIdentifier(node) {
			return jsUndef
		}
		return jsNull
	case ast.KindTemplateExpression:
		// Template literals with substitutions ARE in LITERAL_TYPES via
		// TYPES (no override). We reuse staticEvalTemplate which produces
		// a string — non-empty since at least one quasi or placeholder
		// always renders.
		return staticEvalTemplate(node.AsTemplateExpression())
	case ast.KindTaggedTemplateExpression:
		// LITERAL_TYPES.TaggedTemplateExpression is NOT overridden — it
		// inherits TYPES.TaggedTemplateExpression, which forwards to
		// TemplateLiteral on `value.quasi`. So `tag` + redundant template
		// content (e.g. `tag`photo`` ) DOES get extracted as a string and
		// the rule reports. We mirror by digging into the inner Template.
		tt := node.AsTaggedTemplateExpression()
		if tt == nil || tt.Template == nil {
			return jsNull
		}
		switch tt.Template.Kind {
		case ast.KindNoSubstitutionTemplateLiteral:
			return jsxAstUtilsLiteralCoerce(tt.Template.AsNoSubstitutionTemplateLiteral().Text)
		case ast.KindTemplateExpression:
			return staticEvalTemplate(tt.Template.AsTemplateExpression())
		}
		return jsNull
	case ast.KindBinaryExpression:
		// LITERAL_TYPES.AssignmentExpression is NOT overridden — it inherits
		// TYPES.AssignmentExpression, which formats `${left} ${op} ${right}`
		// using TYPES (i.e. the getPropValue path) on each side. tsgo
		// collapses ESTree's AssignmentExpression into BinaryExpression with
		// an assignment operator, so we detect it here. Non-assignment
		// BinaryExpressions stay LITERAL_TYPES.BinaryExpression = noop →
		// null.
		bin := node.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return jsNull
		}
		switch bin.OperatorToken.Kind {
		case ast.KindEqualsToken,
			ast.KindPlusEqualsToken,
			ast.KindMinusEqualsToken,
			ast.KindAsteriskEqualsToken,
			ast.KindSlashEqualsToken,
			ast.KindPercentEqualsToken,
			ast.KindAsteriskAsteriskEqualsToken,
			ast.KindAmpersandAmpersandEqualsToken,
			ast.KindBarBarEqualsToken,
			ast.KindQuestionQuestionEqualsToken,
			ast.KindAmpersandEqualsToken,
			ast.KindBarEqualsToken,
			ast.KindCaretEqualsToken,
			ast.KindLessThanLessThanEqualsToken,
			ast.KindGreaterThanGreaterThanEqualsToken,
			ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
			// Format `${getValue(left)} ${operator} ${getValue(right)}`
			// using staticEval on each side (mirrors jsx-ast-utils'
			// `require('.').default` which is the TYPES extract, NOT the
			// literal path). Identifier sides stringify to their name;
			// literals stringify to their values.
			leftStr := jsToString(staticEval(bin.Left))
			rightStr := jsToString(staticEval(bin.Right))
			return jsValue{Kind: jvString, Str: leftStr + " " + assignmentOperatorText(bin.OperatorToken.Kind) + " " + rightStr}
		}
		return jsNull
	case ast.KindPrefixUnaryExpression:
		// LITERAL_TYPES.UnaryExpression: same as TYPES but undefined → null.
		v := staticEvalUnary(node.AsPrefixUnaryExpression())
		if v.Kind == jvUndef {
			return jsNull
		}
		return v
	case ast.KindArrayLiteralExpression:
		// LITERAL_TYPES.ArrayExpression filters out null entries. We don't
		// model the actual contents — for our purposes (just truthy?) any
		// array (empty or not) is truthy in JS, so jsTruthy is enough.
		return jsTruthy
	}
	// All other expressions (Call, Member, Conditional, Logical, Binary, etc.)
	// → noop → null per LITERAL_TYPES. Returning jvNull marks them as falsy
	// without claiming any specific identity.
	return jsNull
}

// assignmentOperatorText returns the textual form of an assignment-operator
// kind. Used by literalPropValue's AssignmentExpression branch to mirror
// jsx-ast-utils' `${left} ${operator} ${right}` synthesis.
func assignmentOperatorText(k ast.Kind) string {
	switch k {
	case ast.KindEqualsToken:
		return "="
	case ast.KindPlusEqualsToken:
		return "+="
	case ast.KindMinusEqualsToken:
		return "-="
	case ast.KindAsteriskEqualsToken:
		return "*="
	case ast.KindSlashEqualsToken:
		return "/="
	case ast.KindPercentEqualsToken:
		return "%="
	case ast.KindAsteriskAsteriskEqualsToken:
		return "**="
	case ast.KindAmpersandAmpersandEqualsToken:
		return "&&="
	case ast.KindBarBarEqualsToken:
		return "||="
	case ast.KindQuestionQuestionEqualsToken:
		return "??="
	case ast.KindAmpersandEqualsToken:
		return "&="
	case ast.KindBarEqualsToken:
		return "|="
	case ast.KindCaretEqualsToken:
		return "^="
	case ast.KindLessThanLessThanEqualsToken:
		return "<<="
	case ast.KindGreaterThanGreaterThanEqualsToken:
		return ">>="
	case ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
		return ">>>="
	}
	return "="
}

// jsxAstUtilsLiteralCoerce mirrors jsx-ast-utils' string-literal extractor,
// which normalizes the case-insensitive strings "true" / "false" to actual
// booleans. Any other string is returned unchanged.
//
// Why this matters: alt-text and other a11y rules read attribute values
// through this extractor. A literal `<img alt="false" />` evaluates to the
// boolean `false`, not the string "false" — making it falsy and failing the
// `(altValue && !isNullValued) || altValue === ''` check. Skipping this
// normalization would silently accept `alt="false"`.
func jsxAstUtilsLiteralCoerce(text string) jsValue {
	switch strings.ToLower(text) {
	case "true":
		return jsValue{Kind: jvBool, Bool: true}
	case "false":
		return jsValue{Kind: jvBool, Bool: false}
	}
	return jsValue{Kind: jvString, Str: text}
}

// jsValueIsExactlyUndefined reports whether the value is statically known
// to be `undefined`. Used by ariaLabelHasValue.
func jsValueIsExactlyUndefined(v jsValue) bool {
	return v.Kind == jvUndef
}
