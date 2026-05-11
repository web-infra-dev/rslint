package tabindex_no_positive

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestTabindexNoPositiveExtras locks in behavior NOT covered by upstream's
// own test file. Each block is a single semantic category — see comments
// for the upstream branch each case exercises.
//
// All cases here have been validated against the real eslint-plugin-jsx-a11y
// v6 rule via /tmp/tnp-verify (see the verification skill's notes).
func TestTabindexNoPositiveExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &TabindexNoPositiveRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Boolean / null / undefined extraction shapes
		// ============================================================
		// `tabIndex={false}` — JSXExpressionContainer.FalseKeyword → bool false
		// → Number(false) = 0 → `0 <= 0` skip.
		{Code: `<div tabIndex={false} />`, Tsx: true},
		// `tabIndex="false"` — StringLiteral → jsxAstUtilsLiteralCoerce → bool
		// false → Number(false) = 0 → skip.
		{Code: `<div tabIndex="false" />`, Tsx: true},
		// Case-insensitive coerce: `"FALSE"`, `"False"` → bool false → skip.
		{Code: `<div tabIndex="FALSE" />`, Tsx: true},
		{Code: `<div tabIndex="False" />`, Tsx: true},
		// `tabIndex={"false"}` — StringLiteral in JsxExpression also routes
		// through jsxAstUtilsLiteralCoerce → bool false → skip.
		{Code: `<div tabIndex={"false"} />`, Tsx: true},
		// NoSubstitutionTemplateLiteral `\`false\`` — does NOT route through
		// the "true"/"false" coercion (that's StringLiteral-only). Yields
		// the raw string "false"; Number("false") = NaN → skip.
		{Code: "<div tabIndex={`false`} />", Tsx: true},
		// Same for `\`true\``: raw string "true" → NaN → skip. Locks the
		// asymmetry between StringLiteral and NoSubstitutionTemplateLiteral.
		{Code: "<div tabIndex={`true`} />", Tsx: true},
		// `\`0\`` — raw string "0" → Number("0") = 0 → skip.
		{Code: "<div tabIndex={`0`} />", Tsx: true},
		// `\`-1\`` — raw string "-1" → Number("-1") = -1 → skip.
		{Code: "<div tabIndex={`-1`} />", Tsx: true},
		// `\`abc\`` — non-numeric → NaN → skip.
		{Code: "<div tabIndex={`abc`} />", Tsx: true},
		// Template with escape sequences — upstream's TemplateLiteral
		// extractor reads `quasi.value.raw` (NOT `cooked`), so `\t` / `\n`
		// stay as literal backslash-letter pairs (2 chars each).
		// `\`\t1\t\`` raw is `\t1\t` (6 chars including backslashes),
		// `Number("\t1\t")` would trim if interpreted as escape, but raw
		// keeps backslashes → NaN → skip. Lock this raw-vs-cooked semantics.
		{Code: "<div tabIndex={`\\t1\\t`} />", Tsx: true},
		{Code: "<div tabIndex={`\\n5`} />", Tsx: true},
		{Code: "<div tabIndex={`\\u0035`} />", Tsx: true},

		// ============================================================
		// String numeric edge cases — JS StringToNumber semantics
		// ============================================================
		// Empty string → Number("") = 0 → skip.
		{Code: `<div tabIndex="" />`, Tsx: true},
		{Code: `<div tabIndex={""} />`, Tsx: true},
		// Whitespace-only string → trim → "" → 0 → skip.
		{Code: `<div tabIndex=" " />`, Tsx: true},
		{Code: `<div tabIndex="   " />`, Tsx: true},
		// Signed-prefix hex/oct/bin — JS `Number("-0x10")` is NaN per spec
		// (signed prefixes aren't accepted for non-decimal bases).
		{Code: `<div tabIndex="-0x10" />`, Tsx: true},
		// Malformed hex / oct / bin — ParseUint fails → NaN → skip.
		{Code: `<div tabIndex="0x" />`, Tsx: true},
		{Code: `<div tabIndex="0xZZ" />`, Tsx: true},
		// "+0" / "-0" decimal strings → Number → 0 / -0 → skip via `<= 0`.
		{Code: `<div tabIndex="+0" />`, Tsx: true},
		{Code: `<div tabIndex="-0" />`, Tsx: true},
		// " -1 " — leading/trailing whitespace stripped by Number; -1 ≤ 0.
		{Code: `<div tabIndex=" -1 " />`, Tsx: true},
		// Plain non-numeric — NaN → skip.
		{Code: `<div tabIndex="abc" />`, Tsx: true},

		// ============================================================
		// Numeric edge cases — negative / non-integer / sentinel values
		// ============================================================
		// Negative integer / float — all `<= 0` skip.
		{Code: `<div tabIndex={-100} />`, Tsx: true},
		{Code: `<div tabIndex={-0.5} />`, Tsx: true},
		// Negative zero — JS `-0 <= 0` is true → skip.
		{Code: `<div tabIndex={-0} />`, Tsx: true},
		// `NaN` identifier — LITERAL_TYPES.Identifier = () => null → Number(null) = 0 → skip.
		{Code: `<div tabIndex={NaN} />`, Tsx: true},
		// `Infinity` identifier — same Identifier → null path; ours diverges
		// from staticEval (which returns +Inf) because LITERAL_TYPES strips
		// the JS_RESERVED special-cases.
		{Code: `<div tabIndex={Infinity} />`, Tsx: true},
		// `-Infinity` — PrefixUnary on Identifier; staticEval(Infinity)
		// returns "Infinity" string (it's a reserved name) → ToNumber →
		// +Inf → negate → -Inf → `-Inf <= 0` true → skip.
		{Code: `<div tabIndex={-Infinity} />`, Tsx: true},

		// ============================================================
		// BigInt — JS BigInt → Number coercion
		// ============================================================
		// 0n → Number(0n) = 0 → skip.
		{Code: `<div tabIndex={0n} />`, Tsx: true},
		// Negative BigInt → -1 → skip.
		{Code: `<div tabIndex={-1n} />`, Tsx: true},

		// ============================================================
		// Identifier / Call / Member / Conditional / Logical / Nullish
		// — all LITERAL_TYPES noop → null → Number(null) = 0 → skip.
		// ============================================================
		{Code: `<div tabIndex={obj.x} />`, Tsx: true},
		{Code: `<div tabIndex={obj?.x} />`, Tsx: true},
		// Optional-call — also CallExpression-like noop.
		{Code: `<div tabIndex={fn?.()} />`, Tsx: true},
		// Number() constructor call still noop under LITERAL_TYPES (it's a
		// CallExpression, regardless of what the callee is).
		{Code: `<div tabIndex={Number(5)} />`, Tsx: true},
		// Triple unary — `-(-(-1))` = -1 → skip.
		{Code: `<div tabIndex={-(-(-1))} />`, Tsx: true},
		{Code: `<div tabIndex={-(-(-2))} />`, Tsx: true},
		// Conditional with BigInt arms — Conditional is noop → null → 0 → skip.
		{Code: `<div tabIndex={true ? 1n : 0n} />`, Tsx: true},
		// ConditionalExpression with both arms positive — LITERAL_TYPES noop
		// → null → 0 → skip. This is the key divergence from
		// no-noninteractive-tabindex's getTabIndex (step-2 fallback would
		// resolve to 1 and report).
		{Code: `<div tabIndex={cond ? 1 : 2} />`, Tsx: true},
		{Code: `<div tabIndex={true ? 1 : 2} />`, Tsx: true},
		// LogicalExpression — noop → null → 0 → skip.
		{Code: `<div tabIndex={1 || 2} />`, Tsx: true},
		{Code: `<div tabIndex={null ?? 1} />`, Tsx: true},
		// BinaryExpression — noop → null → 0 → skip.
		{Code: `<div tabIndex={1 + 1} />`, Tsx: true},
		// NewExpression / ObjectExpression — noop → null → 0 → skip.
		{Code: `<div tabIndex={new X()} />`, Tsx: true},
		{Code: `<div tabIndex={{x:1}} />`, Tsx: true},
		// RegExp literal — TYPES.Literal regex branch returns null; here
		// also null → 0 → skip.
		{Code: `<div tabIndex={/foo/} />`, Tsx: true},
		// JSXElement / FunctionExpression / ArrowFunction — LITERAL_TYPES
		// has no overrides; ArrowFunction is FunctionExpression-like → null →
		// Number(null) = 0 → skip.
		{Code: `<div tabIndex={<X/>} />`, Tsx: true},
		{Code: `<div tabIndex={() => 5} />`, Tsx: true},
		{Code: `<div tabIndex={function(){}} />`, Tsx: true},
		// TYPES.SequenceExpression returns an array. LITERAL_TYPES doesn't
		// override; the noop fallback returns null → 0 → skip.
		{Code: `<div tabIndex={(0, 5)} />`, Tsx: true},
		// Math.PI — MemberExpression → null → 0 → skip.
		{Code: `<div tabIndex={Math.PI} />`, Tsx: true},

		// ============================================================
		// Template with substitutions — placeholder rendering
		// ============================================================
		// Substitution whose inner type is Literal (no "Expression" in the
		// name) yields "" → joined "" → Number("") = 0 → skip.
		{Code: "<div tabIndex={`${5}`} />", Tsx: true},
		{Code: "<div tabIndex={`${0}`} />", Tsx: true},
		{Code: "<div tabIndex={`${false}`} />", Tsx: true},
		{Code: "<div tabIndex={`${\"5\"}`} />", Tsx: true},
		// Substitution containing another template — the inner TemplateLiteral
		// is type `TemplateLiteral` (no "Expression"), so the substitution
		// renders to "" → outer joined "" → 0 → skip. This locks the
		// counter-intuitive upstream behavior where intuitively-positive
		// values get silently dropped.
		{Code: "<div tabIndex={`${`5`}`} />", Tsx: true},
		{Code: "<div tabIndex={`${`${5}`}`} />", Tsx: true},
		// Conditional in substitution: type is "ConditionalExpression" (contains
		// "Expression"), so substitution renders as "{ConditionalExpression}"
		// → Number("{ConditionalExpression}") = NaN → skip.
		{Code: "<div tabIndex={`${cond ? 1 : 2}`} />", Tsx: true},
		// Tagged template with substitution — same TemplateLiteral path inside
		// the tag's quasi yields "" for the Literal substitution → "" → 0 → skip.
		{Code: "<div tabIndex={tag`${5}`} />", Tsx: true},

		// ============================================================
		// ArrayLiteralExpression — Array.join semantics
		// ============================================================
		// Empty array → "" → 0 → skip.
		{Code: `<div tabIndex={[]} />`, Tsx: true},
		// [null] / [undefined] / [bar] — null/undefined join to ""; non-undefined
		// Identifier via TYPES = name "bar" → "bar" → NaN → skip.
		{Code: `<div tabIndex={[null]} />`, Tsx: true},
		{Code: `<div tabIndex={[undefined]} />`, Tsx: true},
		{Code: `<div tabIndex={[bar]} />`, Tsx: true},
		// Multi-element array joins with "," → NaN.
		{Code: `<div tabIndex={[5,6]} />`, Tsx: true},
		// Boolean / non-numeric string elements → ToNumber → NaN.
		{Code: `<div tabIndex={[true]} />`, Tsx: true},
		{Code: `<div tabIndex={[false]} />`, Tsx: true},
		{Code: `<div tabIndex={["abc"]} />`, Tsx: true},
		// Multi-element with trailing empty string — join "," → "1," → NaN → skip.
		{Code: `<div tabIndex={[1, ""]} />`, Tsx: true},
		// Sparse arrays — null-elements in slot positions stringify to "".
		// `[,5]` → ",5" → NaN; `[,,]` → "," → NaN; both skip.
		{Code: `<div tabIndex={[,5]} />`, Tsx: true},
		{Code: `<div tabIndex={[,,]} />`, Tsx: true},
		// Nested array yielding 0 → "0" → skip.
		{Code: `<div tabIndex={[[0]]} />`, Tsx: true},
		// Nested array with multiple elements → join "," → NaN → skip.
		{Code: `<div tabIndex={[[5,6]]} />`, Tsx: true},
		// Array of Unary on non-numeric operand — TYPES.UnaryExpression
		// applies ToNumber to operand, so `+null` is 0 ("0" → 0 → skip)
		// and `-true` is -1 ("-1" → -1 → skip).
		{Code: `<div tabIndex={[+null]} />`, Tsx: true},
		{Code: `<div tabIndex={[+undefined]} />`, Tsx: true},
		{Code: `<div tabIndex={[-true]} />`, Tsx: true},
		{Code: `<div tabIndex={[!0]} />`, Tsx: true}, // [true] → "true" → NaN → skip
		{Code: `<div tabIndex={[~0]} />`, Tsx: true}, // [-1] → "-1" → -1 → skip
		// `[typeof x]` — TYPES.UnaryExpression → typeof returns a type string;
		// my impl returns jsUndef → "" → 0 → skip. Both paths skip.
		{Code: `<div tabIndex={[typeof x]} />`, Tsx: true},
		// `[void 0]` — TYPES.UnaryExpression → undefined; my impl jsUndef → "" → 0 → skip.
		{Code: `<div tabIndex={[void 0]} />`, Tsx: true},
		// `[delete a.b]` — both paths yield boolean true → "true" → NaN → skip.
		{Code: `<div tabIndex={[delete a.b]} />`, Tsx: true},
		// `[NaN]` — TYPES.Identifier("NaN") falls to default → "NaN" string
		// (NaN is not in JS_RESERVED). Array.join → "NaN" → Number("NaN") = NaN → skip.
		// My impl matches (Identifier "NaN" → jvString "NaN" → "NaN" → NaN → skip).
		{Code: `<div tabIndex={[NaN]} />`, Tsx: true},
		// `[Math.PI]` — MemberExpression resolves to synthesized string; both paths skip.
		{Code: `<div tabIndex={[Math.PI]} />`, Tsx: true},
		// Array of single zero — "0" → 0 → skip; multi-zero joins to "0,0" → NaN → skip.
		{Code: `<div tabIndex={[0]} />`, Tsx: true},
		{Code: `<div tabIndex={[0,0,0]} />`, Tsx: true},
		{Code: `<div tabIndex={[5,0]} />`, Tsx: true},
		// Array of object / regex — staticEval yields jsTruthy → "" → 0 → skip.
		// Upstream TYPES.ObjectExpression yields {} → "[object Object]" → NaN → skip.
		// Both paths skip; observable behavior identical.
		{Code: `<div tabIndex={[{}]} />`, Tsx: true},
		{Code: `<div tabIndex={[/foo/]} />`, Tsx: true},
		// Postfix update — NaN → skip.
		{Code: `function F(){ let x = 0; return <div tabIndex={[x++]} />; }`, Tsx: true},
		// Array of JSX / function / arrow — synthesized truthy → "" → 0 → skip.
		{Code: `<div tabIndex={[<X/>]} />`, Tsx: true},
		{Code: `<div tabIndex={[function(){}]} />`, Tsx: true},
		{Code: `<div tabIndex={[() => 5]} />`, Tsx: true},
		// Array with NoSubstitutionTemplate yielding non-numeric.
		{Code: "<div tabIndex={[`abc`]} />", Tsx: true},
		// Array with TemplateExpression-substitution — inner substitution
		// renders to "" → [""] → "" → 0 → skip.
		{Code: "<div tabIndex={[`${5}`]} />", Tsx: true},
		// Zero in hex/oct/bin.
		{Code: `<div tabIndex={0o0} />`, Tsx: true},
		{Code: `<div tabIndex={0b0} />`, Tsx: true},
		{Code: `<div tabIndex={0x0} />`, Tsx: true},

		// ============================================================
		// Extreme small negative — `-0.0000001 < 0` → skip
		// ============================================================
		{Code: `<div tabIndex={-0.0000001} />`, Tsx: true},

		// ============================================================
		// MemberExpression-resolved numbers (Number.EPSILON, Number.MAX_VALUE)
		// — upstream LITERAL_TYPES.MemberExpression is noop → null → 0 → skip
		// ============================================================
		{Code: `<div tabIndex={Number.EPSILON} />`, Tsx: true},
		{Code: `<div tabIndex={Number.MAX_VALUE} />`, Tsx: true},

		// ============================================================
		// Non-Latin / fullwidth numerals — Number rejects them as NaN
		// ============================================================
		// U+FF15 fullwidth 5 → JS `Number("５")` is NaN.
		{Code: `<div tabIndex="５" />`, Tsx: true},

		// HTML named entities that don't resolve to a numeric string —
		// `&nbsp;` alone is just U+00A0 (no-break space) → trim → "" → 0 → skip.
		// `&amp;` is "&" → NaN → skip. Decoded by jsxtransforms.DecodeEntities
		// to match babel's JSX attribute string semantics.
		{Code: `<div tabIndex="&nbsp;" />`, Tsx: true},
		{Code: `<div tabIndex="&amp;" />`, Tsx: true},
		{Code: `<div tabIndex="&lt;1" />`, Tsx: true},
		// `&#48;` decodes to "0" → 0 ≤ 0 → skip.
		{Code: `<div tabIndex="&#48;" />`, Tsx: true},
		// Unknown entity stays as-is → NaN → skip.
		{Code: `<div tabIndex="&unknown;" />`, Tsx: true},
		// `&#0;` decodes to NUL char → NaN → skip.
		{Code: `<div tabIndex="&#0;" />`, Tsx: true},
		// Crucial negative: `{"&#49;"}` — the StringLiteral is inside a
		// JsxExpression, NOT a direct attribute string, so HTML entities
		// stay as the literal 5-char "&#49;" → Number = NaN → skip.
		// Mirrors babel's JSX parser which only decodes entities in
		// attribute strings (and JSX text), not inside JS expressions.
		{Code: `<div tabIndex={"&#49;"} />`, Tsx: true},
		{Code: "<div tabIndex={`&#49;`} />", Tsx: true},


		// ============================================================
		// JsxText "tabIndex=5" — pure text, not a JSX attribute, no report
		// ============================================================
		{Code: `<div>tabIndex=5</div>`, Tsx: true},
		// Element literally named tabIndex without prop — no JsxAttribute, no report.
		{Code: `<tabIndex />`, Tsx: true},

		// ============================================================
		// Spread of identifier — non-literal spread is opaque, no extraction
		// possible → no report even if runtime value is { tabIndex: 5 }.
		// Mirrors upstream's listener-only-on-JsxAttribute behavior.
		// ============================================================
		{Code: `const props = cond ? {tabIndex:1} : {}; const Q = () => <div {...props} />;`, Tsx: true},
		// Array containing SpreadElement — element is SpreadElement, not a
		// directly extractable value; both upstream and ours treat the
		// spread as "" in Array.join → joined value has no positive number
		// directly attributable to tabIndex → skip.
		{Code: `<div tabIndex={[...arr]} />`, Tsx: true},
		{Code: `<div tabIndex={[...arr, 5]} />`, Tsx: true},

		// ============================================================
		// JSX attribute with map+index (real-world pattern)
		// ============================================================
		// Identifier `i` → LITERAL_TYPES.Identifier = null → Number = 0 → skip.
		{Code: `function F(){return [1,2].map((x, i) => <li tabIndex={i} />);}`, Tsx: true},
		// `i + 1` BinaryExpression → LITERAL_TYPES.BinaryExpression noop →
		// null → 0 → skip. The rule cannot statically prove this is positive
		// at lint time, so it stays silent (acceptable false negative —
		// upstream behavior).
		{Code: `function F(){return [1,2].map((x, i) => <li tabIndex={i + 1} />);}`, Tsx: true},

		// ============================================================
		// Tagged template with substitution — TaggedTemplateExpression
		// delegates to TemplateLiteral on the inner quasi+expressions;
		// substitution renders to "" → "" → 0 → skip.
		// ============================================================
		{Code: "<div tabIndex={String.raw`${5}`} />", Tsx: true},

		// ============================================================
		// Unary — `+/-/!/~` operand coercion via LITERAL_TYPES.UnaryExpression
		// ============================================================
		// `+null` = 0; `-true` = -1; `+undefined` = NaN; `+"abc"` = NaN.
		{Code: `<div tabIndex={+null} />`, Tsx: true},
		{Code: `<div tabIndex={+undefined} />`, Tsx: true},
		{Code: `<div tabIndex={-true} />`, Tsx: true},
		{Code: `<div tabIndex={+"abc"} />`, Tsx: true},
		// `!1` = false → 0 → skip.
		{Code: `<div tabIndex={!1} />`, Tsx: true},
		// `~0` = -1 → skip; `~-1` = 0 → skip.
		{Code: `<div tabIndex={~0} />`, Tsx: true},
		{Code: `<div tabIndex={~-1} />`, Tsx: true},

		// ============================================================
		// typeof / void / postfix update — non-reporting unary kinds
		// ============================================================
		// Both produce values whose Number coercion happens to land on a
		// non-positive number or NaN, so they don't report. Locks the
		// observed behavior — the rule shouldn't fire on these even
		// though our code paths differ from upstream's UnaryExpression
		// extractor.
		{Code: `<div tabIndex={typeof x} />`, Tsx: true},
		{Code: `<div tabIndex={void 0} />`, Tsx: true},
		// x++ → operand "x" → ToNumber NaN → x++ value is NaN → skip.
		{Code: `function F(){ let x = 0; return <div tabIndex={x++} />; }`, Tsx: true},

		// ============================================================
		// Parentheses / TS wrappers
		// ============================================================
		// Bare parens are stripped — `(-1)` is still -1.
		{Code: `<div tabIndex={(-1)} />`, Tsx: true},
		{Code: `<div tabIndex={((-1))} />`, Tsx: true},
		// `(-1) as number` — LITERAL_TYPES doesn't recognize TSAsExpression
		// → noop → null → Number(null) = 0 → skip. This matches upstream
		// which only sees TSAsExpression via the typescript-eslint parser
		// and would route through extractValueForLiteral → null.
		{Code: `<div tabIndex={(-1) as number} />`, Tsx: true},
		{Code: `<div tabIndex={5 as any} />`, Tsx: true},
		{Code: `<div tabIndex={5 satisfies number} />`, Tsx: true},
		{Code: `<div tabIndex={(5)!} />`, Tsx: true},
		// AwaitExpression / YieldExpression — same upstream flow as
		// TSSatisfiesExpression: TYPES has no entry → noop → null →
		// Number(null) = 0 → 0 <= 0 → skip. Verified via differential
		// against eslint-plugin-jsx-a11y v6.10.2 — neither expression
		// triggers tabindex-no-positive. LiteralPropToNumber lands the
		// same way through literalPropValue's default jsNull arm followed
		// by jsValueToNumber's jvNull → (0, true) mapping. Locks the
		// alignment in case future LiteralPropToNumber refactors change
		// the jvNull handling.
		{Code: `async function f() { return <div tabIndex={await p} />; }`, Tsx: true},
		{Code: `function* g() { yield <div tabIndex={yield 0} />; }`, Tsx: true},

		// ============================================================
		// Spread attributes — listener fires on JsxAttribute only
		// ============================================================
		// `<div {...{tabIndex: 5}} />` — JsxSpreadAttribute is never visited
		// by the JsxAttribute listener, so this is silently skipped even
		// though the spread carries a positive tabIndex.
		{Code: `<div {...{tabIndex: 5}} />`, Tsx: true},
		{Code: `<div {...props} />`, Tsx: true},
		// Spread mixed with literal-object containing other props — still
		// no JsxAttribute named tabIndex, so silently skipped.
		{Code: `<div {...{tabIndex: 5, role: "x"}} />`, Tsx: true},

		// ============================================================
		// Attribute name variants — only an exact case-insensitive match
		// to "tabIndex" triggers; namespaced and hyphenated variants don't.
		// ============================================================
		// `xml:tabIndex` — GetJsxPropName returns "xml:tabIndex", which is
		// NOT case-insensitive-equal to "tabIndex". Skip.
		{Code: `<div xml:tabIndex="5" />`, Tsx: true},
		// `data-tabIndex` — propName is "data-tabIndex" (kebab), uppercased
		// is "DATA-TABINDEX" ≠ "TABINDEX". Skip. Upstream behavior locked.
		{Code: `<div data-tabIndex="5" />`, Tsx: true},

		// ============================================================
		// Comments around prop don't change extraction
		// ============================================================
		{Code: `<div /* before */ tabIndex={-1} /* after */ />`, Tsx: true},
		{Code: `<div tabIndex={/* note */ -1} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// String "true" → boolean coercion
		// ============================================================
		// Direct StringLiteral "true" — jsxAstUtilsLiteralCoerce → bool true
		// → Number(true) = 1 → report. Same for case-variants "True", "TRUE"
		// (the regex is case-insensitive).
		{Code: `<div tabIndex="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="True" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="TRUE" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// StringLiteral inside JsxExpression — also routed through the
		// boolean coercion.
		{Code: `<div tabIndex={"true"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Boolean attribute form / direct boolean true
		// ============================================================
		// `<div tabIndex />` — extractValue null-attribute-value path returns
		// JS boolean true. Number(true) = 1 → report. This is the key
		// divergence from no-noninteractive-tabindex's getTabIndex (which
		// returns undefined for booleans).
		{Code: `<div tabIndex />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `tabIndex={true}` — JSXExpressionContainer.TrueKeyword → bool true
		// → Number(true) = 1 → report.
		{Code: `<div tabIndex={true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Non-integer numerics — explicit divergence from no-noninteractive-tabindex
		// ============================================================
		{Code: `<div tabIndex={0.5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={1.5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="0.5" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="1.5" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Numeric variations — all positive numbers report
		// ============================================================
		{Code: `<div tabIndex={5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={100} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Hex / octal / binary numeric literals.
		{Code: `<div tabIndex={0x10} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={0o10} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={0b10} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Scientific notation.
		{Code: `<div tabIndex={1e2} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Numeric separator.
		{Code: `<div tabIndex={1_000} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `1e1000` overflows to +Infinity; Infinity > 0 → report.
		{Code: `<div tabIndex={1e1000} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Parenthesized numeric — parens stripped.
		{Code: `<div tabIndex={(5)} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={((5))} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// String numerics — JS StringToNumber on hex / oct / bin / decimal
		// ============================================================
		{Code: `<div tabIndex="5" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="100" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Hex / oct / bin STRING forms — Number("0x10") = 16 → report.
		{Code: `<div tabIndex="0x10" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="0X10" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="0o7" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="0b10" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Leading "+" decimal — Number("+1") = 1 → report.
		{Code: `<div tabIndex="+1" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Whitespace around decimal — trimmed.
		{Code: `<div tabIndex=" 1 " />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Tab / newline whitespace — JS `Number(...)` strips all
		// std-defined whitespace before parsing, so `"\t1\t"` and
		// `"\n1\n"` both coerce to 1 → report.
		{Code: "<div tabIndex=\"\t1\t\" />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<div tabIndex=\"\n1\n\" />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// HTML character entities in attribute strings — babel decodes
		// these at parse time, and rslint matches via
		// jsxtransforms.DecodeEntities. `&#49;` decimal → "1" → report;
		// `&#x31;` hex → "1" → report. nbsp-prefixed `&nbsp;5` decodes
		// to " 5" which JS `Number` trims to 5 → report.
		// ============================================================
		{Code: `<div tabIndex="&#49;" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="&#x31;" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="&nbsp;5" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Larger numeric entity values.
		{Code: `<div tabIndex="&#53;" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // "5"
		{Code: `<div tabIndex="&#x35;" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // "5"


		// ============================================================
		// NoSubstitutionTemplateLiteral and TaggedTemplate
		// ============================================================
		// `\`5\`` — raw string "5" → Number("5") = 5 → report.
		{Code: "<div tabIndex={`5`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `tag\`1\`` — TaggedTemplateExpression with inner NoSubstitution
		// template "1". LITERAL_TYPES.TaggedTemplateExpression inherits TYPES
		// which forwards to TemplateLiteral on `quasi`; literalPropValue's
		// arm digs into the inner template and yields "1" → 1 → report.
		{Code: "<div tabIndex={tag`1`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// BigInt — Number(2n) = 2 → report
		// ============================================================
		{Code: `<div tabIndex={2n} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={1n} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Large BigInt within Float64 representable range — Number(9007199254740992n)
		// = 9007199254740992 (= 2^53) → still > 0 → report. Locks the SetInt
		// → Float64 conversion path.
		{Code: `<div tabIndex={9007199254740992n} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// 1.0e2 — decimal exponent yielding 100 → report.
		{Code: `<div tabIndex={1.0e2} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Array of BinaryExpression — TYPES.BinaryExpression computes 1+1 = 2
		// → array [2] → "2" → 2 → report. staticEvalBinary in my impl handles
		// jvNumber+jvNumber identically.
		{Code: `<div tabIndex={[1 + 1]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Array of ConditionalExpression — TYPES path evaluates the condition
		// statically; truthy Identifier "cond" picks WhenTrue → [1] → "1" → 1 → report.
		{Code: `<div tabIndex={[cond ? 1 : 2]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Array with NoSubstitutionTemplate "5" — element = "5" → "5" → 5 → report.
		{Code: "<div tabIndex={[`5`]} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Extreme small positive numerics — `> 0` (not `>= 0`)
		// ============================================================
		{Code: `<div tabIndex={1e-10} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={0.0000001} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Numeric separator literal — 5,000,000.
		{Code: `<div tabIndex={5_000_000} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Small BigInt.
		{Code: `<div tabIndex={5n} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Real-world component / framework placement — lock listener
		// fires regardless of enclosing syntax
		// ============================================================
		{Code: `export default <div tabIndex={1} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `const f = () => <div tabIndex={1} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `const f = function() { return <div tabIndex={1} />; };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `const o = { render() { return <div tabIndex={1} />; } };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `async function* g() { yield <div tabIndex={1} />; await x; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Render-prop pattern — JSX passed as argument to a function call
		// is still inside the AST, so its attributes are visited.
		{Code: `const Comp = ({render}) => render(<div tabIndex={1} />);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Conditional rendering inside JSX children.
		{Code: `<div>{cond && <span tabIndex={1} />}</div>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div>{cond ? <span tabIndex={1} /> : null}</div>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// JSX tag shapes — JSXNamespacedName, JSXMemberExpression,
		// elements named "tabIndex" itself
		// ============================================================
		{Code: `<svg:foo tabIndex={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<A.B tabIndex={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<A.B.C tabIndex={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Element literally named tabIndex — still fires on its tabIndex prop.
		{Code: `<tabIndex tabIndex={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Coexisting attributes — key / required / spread literal next to tabIndex
		// ============================================================
		{Code: `<li key="k" tabIndex={1}>x</li>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<input required tabIndex={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div {...{key:"x"}} tabIndex={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Whitespace / newlines around attribute — unaffected
		// ============================================================
		{Code: `<div  tabIndex={1}  />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<div\n\ttabIndex={1}\n/>", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Array of JSX whose JSX has positive tabIndex — outer array
		// stringifies to "(jsx)" (NaN, skip), inner JsxAttribute reports.
		// ============================================================
		{Code: `<div tabIndex={[<X tabIndex={1} />]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Self-closing vs paired form — same listener
		// ============================================================
		{Code: `<div tabIndex={1}></div>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Cross-function / cross-element report counts
		// ============================================================
		{Code: `function A(){return <div tabIndex={1} />;} function B(){return <span tabIndex={5} />;}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError}},
		{Code: `<><div tabIndex={1} /><span tabIndex={2} /><p tabIndex={0} /></>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError}},

		// ============================================================
		// JSX text containing "tabIndex=5" — only the attribute reports;
		// JsxText is not visited by the JsxAttribute listener.
		// ============================================================
		{Code: `<div tabIndex={5}>tabIndex=5</div>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Report location — must span the entire JsxAttribute (name + value).
		// Upstream calls `context.report({ node: attribute })` where
		// `attribute` is the JSXAttribute AST node. Locking columns here
		// also catches any future refactor that accidentally narrows the
		// report range to just the value or just the name.
		// ============================================================
		// `<div tabIndex={5} />` — JsxAttribute spans cols 6..18 (= "tabIndex={5}").
		{
			Code: `<div tabIndex={5} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "tabIndexNoPositive",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 18,
			}},
		},
		// String form spans same range.
		{
			Code: `<div tabIndex="5" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "tabIndexNoPositive",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 18,
			}},
		},
		// Multi-line attribute — range spans from name to value's close brace.
		{
			Code: "<div\n  tabIndex={\n    5\n  } />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "tabIndexNoPositive",
				Message:   errorMessage,
				Line:      2, Column: 3, EndLine: 4, EndColumn: 4,
			}},
		},
		// tabIndex preceded by another attribute — range shifts to match position.
		{
			Code: `<div role="article" tabIndex="5" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "tabIndexNoPositive",
				Message:   errorMessage,
				Line:      1, Column: 21, EndLine: 1, EndColumn: 33,
			}},
		},
		// Boolean form `<div tabIndex />` — range covers just the name.
		{
			Code: `<div tabIndex />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "tabIndexNoPositive",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 14,
			}},
		},

		// ============================================================
		// ArrayLiteralExpression — `[5]` → "5" → 5 → report
		// ============================================================
		{Code: `<div tabIndex={[5]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={["5"]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `[Infinity]` — TYPES.Identifier resolves "Infinity" via JS_RESERVED
		// → +Inf → "Infinity" via Array.join → Number("Infinity") = +Inf → report.
		{Code: `<div tabIndex={[Infinity]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `[1n]` → BigInt → "1" → 1 → report.
		{Code: `<div tabIndex={[1n]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Nested `[[5]]` — outer toString → inner toString → "5" → 5 → report.
		{Code: `<div tabIndex={[[5]]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Triple-nested `[[[5]]]` — recursion depth ≥ 3. Each layer
		// toString-flattens to "5" → 5 → report.
		{Code: `<div tabIndex={[[[5]]]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Array with Unary on non-numeric operand — TYPES.UnaryExpression's
		// ToNumber-on-operand makes `+true` = 1 ("1" → 1 → report) and
		// `+"5"` = 5 ("5" → 5 → report). This is the fix lock for an
		// earlier regression where the Array element extraction was
		// using staticEvalUnary, which only handles numeric operands.
		{Code: `<div tabIndex={[+true]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={[+"5"]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `[~-2]` — `~-2` = 1 → "1" → 1 → report. Locks the bitwise-NOT
		// ToInt32 path inside Array element extraction.
		{Code: `<div tabIndex={[~-2]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Large BigInt in array — within Float64 precision. Number(9...n) =
		// 9007199254740992 → Float64 → "9007199254740992" → > 0 → report.
		{Code: `<div tabIndex={[9007199254740992n]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Unary on coerced operand — `+true` / `!0` / `~-2` / `-(-5)`
		// ============================================================
		// `+true` = +Number(true) = 1 → report.
		{Code: `<div tabIndex={+true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `+"5"` = 5; `-"-5"` = 5; `+"0x10"` = 16.
		{Code: `<div tabIndex={+"5"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={-"-5"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={+"0x10"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `+5` / `+0` — already-numeric operand reports.
		{Code: `<div tabIndex={+5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `-(-1)` — nested PrefixUnary; inner -1, outer -(-1) = 1 → report.
		{Code: `<div tabIndex={-(-1)} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `-(-5)` — same pattern, larger magnitude.
		{Code: `<div tabIndex={-(-5)} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Unary on whitespace-padded string — `+`/`-` apply ToNumber, which
		// trims; `+"  5  "` = 5 → report.
		{Code: `<div tabIndex={+"  5  "} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `!0` → true → 1 → report.
		{Code: `<div tabIndex={!0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `~-2` → ToInt32(-2) = -2, ^-2 = 1 → report.
		{Code: `<div tabIndex={~-2} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// `delete` operator — returns boolean true → 1 → report
		// ============================================================
		{Code: `<div tabIndex={delete a.b} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Custom components / namespaced / member tag names — rule
		// fires regardless of element type (unlike no-noninteractive-tabindex)
		// ============================================================
		{Code: `<MyButton tabIndex={5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<UX.Layout tabIndex={5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<svg:circle tabIndex={3} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<this.Foo tabIndex={5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Foo.Bar.Baz tabIndex={5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Case-insensitive name match — propName.toUpperCase() === 'TABINDEX'
		// ============================================================
		{Code: `<div TABINDEX="5" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabindex="5" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div TabIndex="5" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div Tabindex="5" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Random-case match.
		{Code: `<div TaBiNdEx="5" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Multiple tabIndex props — each visited independently; only the
		// invalid one reports.
		// ============================================================
		{Code: `<div tabIndex={0} tabIndex={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={1} tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Both invalid — two reports.
		{Code: `<div tabIndex={1} tabIndex={5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError}},

		// ============================================================
		// Comments don't suppress
		// ============================================================
		{Code: `<div /* a */ tabIndex={5} /* b */ />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={/* truthy */ 5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Listener boundary — nested invalid elements each report
		// ============================================================
		{Code: `<div tabIndex={5}><span tabIndex={1} /></div>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError}},
		{Code: `<><div tabIndex={5} /><span tabIndex={1} /></>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError}},

		// ============================================================
		// Real-world component patterns
		// ============================================================
		{Code: `function Outer() { return <div tabIndex={5}>focusable</div>; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `const items = arr.map(item => <li key={item.id} tabIndex={1}>{item.name}</li>);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `const Pane = React.forwardRef((props, ref) => <div ref={ref} tabIndex={3} {...props} />);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `class Form extends React.Component { render() { return <div tabIndex={5}>ready</div>; } }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Generic JSX — TS-generic tag still matches name
		// ============================================================
		{Code: `<Map<string, number> tabIndex={5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	})
}
