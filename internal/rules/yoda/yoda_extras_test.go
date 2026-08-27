// TestYodaExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
// TestYodaExtrasEditDemand covers autofix/diagnostic-only edit-demand
// invariance for the rule's single deferred fix builder.
//
// Dimension 4 walk (rows that don't apply to yoda, with reasons):
//   - N/A declaration/container forms (class/function shapes): yoda only
//     inspects BinaryExpression comparisons, never functions or classes.
//   - N/A graceful degradation (SpreadAssignment/RestElement, empty bodies):
//     yoda never inspects object/array patterns or statement bodies.
//   - Access/key forms (identifier/string/numeric/computed/private key,
//     bracket forms) are NOT N/A — they are exercised extensively by the
//     upstream exceptRange suite via IsSameReference and are not repeated
//     here.
package yoda

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestYodaExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&YodaRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: single- and multi-level parenthesized receiver on the
			// non-literal side ----
			{Code: `if ((x) === "red") {}`, Options: "never"},
			{Code: `if (((x)) === "red") {}`, Options: "never"},

			// ---- Dimension 4: TS non-null assertion on the non-literal side ----
			{Code: `if (x! === "red") {}`, Options: "never"},

			// ---- Dimension 4: TS `as`/`satisfies` wrappers are not unwrapped, matching
			// ESLint (TSAsExpression/TSSatisfiesExpression are not an ESTree Literal) ----
			{Code: `if (("red" as const) === value) {}`, Options: "never"},
			{Code: `if (("red" satisfies string) === value) {}`, Options: "never"},

			// ---- Dimension 4: optional chain member as the non-literal side ----
			{Code: `if (foo?.bar === "red") {}`, Options: "never"},

			// ---- Documented divergence: a TypeScript-only wrapper has no runtime
			// effect, so the two comparisons still hold the same operand and the
			// range test stays exempt. ESLint's isSameReference has no case for
			// TSNonNullExpression/TSAsExpression/TSSatisfiesExpression, so it reads
			// none of these as a range test and reports every one of them ----
			{Code: `if (0 <= x! && x! < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= (x as number) && (x as number) < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= (x satisfies number) && (x satisfies number) < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= x! && x < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Dimension 4: PrivateIdentifier access in a range test ----
			{Code: `class C { #x = 0; m() { if (0 <= this.#x && this.#x < 1) {} } }`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Dimension 4: a literal shared between the two comparisons is still
			// the same reference, whatever kind of literal it is and however it is
			// spelled ----
			{Code: `if (0 < 1 && 1 < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < 1 && 0x1 < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < 0x10000000000000801 && 18446744073709552000 < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < 18446744073709552000 && 0x10000000000000801 < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < 0b10100010000111101000011111100111101111100110100011110110000010100 && 23363847825694777000 < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < 0o100000000000000001201 && 1152921504606847600 < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < 1_000 && 1000 < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < 1e3 && 1000 < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < 'a' && "a" < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < 1n && 1n < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < true && true < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < null && null < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < /a/ && /a/ < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if ('a' < 'b' && 'b' < c) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Dimension 4: a bare range test wrapped in genuinely redundant parens
			// (not an if/while condition) is still recognized as parenthesized ----
			{Code: `(0 <= x && x < 1);`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Dimension 4: redundant parens around one side of a range test are
			// walked through when finding the logical parent (tsgo ParenthesizedExpression,
			// unlike ESTree, would otherwise sit between the comparison and its `&&`) ----
			{Code: `if ((0 <= x) && x < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Docs: parenthesizing a range test keeps it exempt when it is joined
			// with a further condition ----
			{Code: `if ((0 <= rand && rand < 1) && count < 10) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `function isReddish(color) { return (color.hue < 60 || 300 < color.hue); }`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Real-user: `undefined` is a global Identifier, not an ESTree Literal,
			// so yoda does not enforce ordering against it (a common "why doesn't this
			// get flagged" surprise) ----
			{Code: `if (undefined === x) {}`, Options: "never"},

			// ---- Real-user: a namespaced/enum-style constant on the left is a
			// PropertyAccessExpression, not a literal, so it is never mistaken for a
			// Yoda condition ----
			{Code: `if (Color.Red === value) {}`, Options: "never"},

			// ---- Locks in upstream isEqualityOperator(): `!=`/`!==` are comparison
			// operators but not equality operators, so onlyEquality disables enforcement
			// for them too, not just `<`/`>`/`<=`/`>=` ----
			{Code: `if (x !== 5) {}`, Options: []any{"always", map[string]any{"onlyEquality": true}}},

			// ---- Combination matrix: exceptRange and onlyEquality both true; onlyEquality
			// alone already exempts the non-equality range comparisons ----
			{Code: `if (0 <= x && x < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true, "onlyEquality": true}}},

			// ---- Bug fix: a regex literal is a valid Literal bound for upstream's
			// getNormalizedLiteral (it only checks node.type === "Literal", not the
			// value's runtime type). JS's abstract relational comparison coerces a
			// RegExp via ToPrimitive to its source text, so "/a/" <= "/b/" compares
			// lexicographically and the range is ascending ----
			{Code: `if (/a/ <= x && x < /b/) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Bug fix: JavaScript orders strings by UTF-16 code units. An astral
			// character's leading surrogate sorts before U+E000 even though its UTF-8
			// encoding sorts after U+E000 ----
			{Code: `if ("\ud83d\ude00" <= x && x < "\ue000") {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Bug fix: RegExp.prototype.toString canonicalizes flags, so source
			// spellings with the same flags in different orders are equal bounds ----
			{Code: `if (/a/mi <= x && x < /a/im) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= a[/a/mi] && a[/a/im] < 2) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Bug fix: surrounding tokens come from the parsed token stream, so a
			// preceding interpolated template cannot hide the range's parentheses ----
			{Code: "`${seed}`; if (0 <= x && x < 1) {}", Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Bug fix: a string range bound is coerced by JS's ToNumber, which
			// reads an unsigned `0x`/`0o`/`0b` integer, unlike Go's own float parsing ----
			{Code: `if ('0x10' <= x && x < 20) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if ('0o17' <= x && x < 20) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Bug fix: ToNumber and StringToBigInt both trim a U+FEFF ZWNBSP off
			// a string bound, which is ECMAScript whitespace but not Go's ----
			{Code: "if ('\ufeff5' <= x && x < 10) {}", Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: "if (1n <= x && x < '\ufeff2') {}", Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Bug fix: static numeric member keys use JavaScript's NumberToString
			// notation before isSameReference compares them with a string key ----
			{Code: `if (0 <= a[1e-7] && a['1e-7'] < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= a[1e-6] && a['0.000001'] < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= a[1e20] && a['100000000000000000000'] < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= a[1e21] && a['1e+21'] < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Bug fix: explicit-radix numeric range bounds must recover the raw
			// token before applying JavaScript Number rounding ----
			{Code: `if (0x10000000000000801 <= x && x < 18446744073709552000) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- A sign is allowed before a decimal StringToBigInt input ----
			{Code: `if ('+1' <= x && x < 2n) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user: optional chaining under "always" mode ----
			{
				Code:    `if (data?.status === 200) {}`,
				Output:  []string{`if (200 === data?.status) {}`},
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected literal to be on the left side of ===.",
					Line:      1,
					Column:    5,
					EndLine:   1,
					EndColumn: 25,
				}},
			},

			// ---- Locks in upstream isRangeTestOperator(): `>`/`>=` never form a range
			// test, even in an `&&`-joined shape that otherwise looks like one ----
			{
				Code:    `if (0 >= x && x > 1) {}`,
				Output:  []string{`if (x <= 0 && x > 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- Locks in upstream isBetweenTest(): `||` never satisfies the `&&`-only
			// "between" test, so a superficially between-shaped `||` expression is not
			// exempted (isOutsideTest's own reference check also fails: 0 and 1 are
			// unrelated literals, not the same reference) ----
			{
				Code:    `if (0 <= x || x < 1) {}`,
				Output:  []string{`if (x >= 0 || x < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- A shared literal is the same reference only when both its kind and
			// its value match, so a number and the string spelling it are two
			// different bounds ----
			{
				Code:    `if (0 < 1 && "1" < x) {}`,
				Output:  []string{`if (0 < 1 && x > "1") {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 14}},
			},

			// ---- A substitution-free template is an ESTree TemplateLiteral rather
			// than a Literal, so two of them are never the same reference ----
			{
				Code:    "if (0 < `a` && `a` < x) {}",
				Output:  []string{"if (0 < `a` && x > `a`) {}"},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 16}},
			},

			// ---- Docs: a range test joined with a further condition needs its own
			// parentheses, otherwise the comparison's logical parent is the
			// `count < 10 && 0 <= rand` pair, which is not a range test ----
			{
				Code:    `if (count < 10 && 0 <= rand && rand < 1) {}`,
				Output:  []string{`if (count < 10 && rand >= 0 && rand < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 19}},
			},

			// ---- Locks in upstream isRangeTest(): the sibling of a comparison inside
			// the `&&`/`||` must itself be a BinaryExpression; a plain identifier
			// operand disqualifies the whole shape from being a range test ----
			{
				Code:    `if (flag && 1 > x) {}`,
				Output:  []string{`if (flag && x < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 13}},
			},

			// ---- Dimension 3: multi-line code — whitespace/newlines around the operator
			// are preserved verbatim by the fix, same as inline comments ----
			{
				Code:    "if (0 <=\n  x) {}",
				Output:  []string{"if (x >=\n  0) {}"},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5, EndLine: 2, EndColumn: 4}},
			},

			// ---- Bug fix: a swapped regexp operand must remain separated from a
			// following `instanceof` or `in` token so the fix still parses ----
			{
				Code:    `/a/ < x++instanceof C`,
				Output:  []string{`x++ > /a/ instanceof C`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `/a/ < f()in obj`,
				Output:  []string{`f() > /a/ in obj`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    "`${seed}`; function *f(){yield(1)<a}",
				Output:  []string{"`${seed}`; function *f(){yield a>(1)}"},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `/a/ < (x)as number`,
				Output:  []string{`(x) > /a/ as number`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `/a/ < (x)satisfies number`,
				Output:  []string{`(x) > /a/ satisfies number`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `1. > x++instanceof C`,
				Output:  []string{`x++ < 1. instanceof C`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `1. > f()in obj`,
				Output:  []string{`f() < 1. in obj`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `1. > (x)as number`,
				Output:  []string{`(x) < 1. as number`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `1. > (x)satisfies number`,
				Output:  []string{`(x) < 1. satisfies number`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `function f(){return(x) < /a/}`,
				Output:  []string{`function f(){return /a/ > (x)}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `function *f(){yield(x) < /a/}`,
				Output:  []string{`function *f(){yield /a/ > (x)}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `function f(){throw(x) < /a/}`,
				Output:  []string{`function f(){throw /a/ > (x)}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `/a/g > x++instanceof C`,
				Output:  []string{`x++ < /a/g instanceof C`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `"a" > x++instanceof C`,
				Output:  []string{`x++ < "a"instanceof C`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    "`a` > x++instanceof C",
				Output:  []string{"x++ < `a`instanceof C"},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `function f(){return(x) < "a"}`,
				Output:  []string{`function f(){return"a" > (x)}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    "function f(){return(x) < `a`}",
				Output:  []string{"function f(){return`a` > (x)}"},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},
			{
				Code:    `function f(){return(x) < 1.}`,
				Output:  []string{`function f(){return 1. > (x)}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},

			// ---- Bug fix: private access syntax never aliases an ordinary computed
			// string key in upstream's same-reference comparison ----
			{
				Code:    `class C { #x = 0; m() { if (0 <= this.#x && this['#x'] < 1) {} } }`,
				Output:  []string{`class C { #x = 0; m() { if (this.#x >= 0 && this['#x'] < 1) {} } }`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected"}},
			},

			// ---- Bug fix: explicit-radix numeric member keys must use JavaScript's
			// per-operation Number rounding before static-reference comparison ----
			{
				Code:    `if (0 <= a[0x10000000000000801] && a["18446744073709556000"] < 1) {}`,
				Output:  []string{`if (a[0x10000000000000801] >= 0 && a["18446744073709556000"] < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- Bug fix: reversing the UTF-16 bounds makes the range descend, so
			// exceptRange must not suppress the first Yoda comparison ----
			{
				Code:    `if ("\ue000" <= x && x < "\ud83d\ude00") {}`,
				Output:  []string{`if (x >= "\ue000" && x < "\ud83d\ude00") {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- Bug fix: mixed BigInt/Number range bounds must compare exactly, not
			// through a lossy float64 conversion. 9007199254740993n rounds to
			// 9007199254740992 in float64 (both exceed 2^53), which would make the
			// bounds look ascending when they are not; JS's actual BigInt-vs-Number
			// abstract relational comparison is exact, so this is not a valid range
			// test and exceptRange must not exempt it ----
			{
				Code:    `if (9007199254740993n <= x && x < 9007199254740992) {}`,
				Output:  []string{`if (x >= 9007199254740993n && x < 9007199254740992) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- Bug fix: mixed BigInt/String range bounds must also compare exactly.
			// JS's abstract relational comparison converts the String bound via
			// StringToBigInt for an exact BigInt comparison, never through a lossy
			// float64 conversion ----
			{
				Code:    `if (9007199254740993n <= x && x < "9007199254740992") {}`,
				Output:  []string{`if (x >= 9007199254740993n && x < "9007199254740992") {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if ("9007199254740993" <= x && x < 9007199254740992n) {}`,
				Output:  []string{`if (x >= "9007199254740993" && x < 9007199254740992n) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- Bug fix: ToNumber has no grammar for a digit separator, nor for any
			// spelling of Infinity but that one, so neither string is a bound the
			// range below it ascends to. Go's ParseFloat reads both ----
			{
				Code:    `if ('1_000' <= x && x < 2000) {}`,
				Output:  []string{`if (x >= '1_000' && x < 2000) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5, EndLine: 1, EndColumn: 17}},
			},
			{
				Code:    `if (5 <= x && x < 'INFINITY') {}`,
				Output:  []string{`if (x >= 5 && x < 'INFINITY') {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5, EndLine: 1, EndColumn: 11}},
			},
			{
				Code:    "if (2n <= x && x < '\ufeff1') {}",
				Output:  []string{"if (x >= 2n && x < '\ufeff1') {}"},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5, EndLine: 1, EndColumn: 12}},
			},

			// ---- Bug fix: StringToBigInt rejects either sign before a non-decimal
			// prefix, so these are not ascending ranges and must not be exempt ----
			{
				Code:    `if ('+0x1' <= x && x < 2n) {}`,
				Output:  []string{`if (x >= '+0x1' && x < 2n) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}},
			},
			{
				Code:    `if ('+0o1' <= x && x < 2n) {}`,
				Output:  []string{`if (x >= '+0o1' && x < 2n) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}},
			},
			{
				Code:    `if ('+0b1' <= x && x < 2n) {}`,
				Output:  []string{`if (x >= '+0b1' && x < 2n) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}},
			},
		},
	)
}

// TestYodaExtrasEditDemand verifies that diagnostic identity (range, message)
// is independent of edit demand, and that the autofix is only materialized
// when requested — yoda's only edit artifact is a single deferred fix.
func TestYodaExtrasEditDemand(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.ts",
		Path:     tspath.Path("/edit-demand.ts"),
	}, `1 < a`, core.ScriptKindTS)

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		comments := rule.NewCommentStore(sourceFile)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(YodaRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		})

		statement := sourceFile.Statements.Nodes[0]
		target := statement.AsExpressionStatement().Expression
		YodaRule.Run(ctx, nil)[target.Kind](target)
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
	}

	none := run(rule.EditDemandNone)
	autofix := run(rule.EditDemandAutofix)
	suggestion := run(rule.EditDemandSuggestion)
	all := run(rule.EditDemandAll)

	for _, got := range []rule.RuleDiagnostic{none, autofix, suggestion, all} {
		if got.Range != none.Range || !reflect.DeepEqual(got.Message, none.Message) {
			t.Fatalf("diagnostic identity changed with edit demand: %#v vs %#v", got, none)
		}
		if got.Suggestions != nil {
			t.Fatalf("yoda has no suggestions, but got some: %#v", got.Suggestions)
		}
	}

	if none.FixesPtr != nil {
		t.Fatalf("EditDemandNone consumer received fixes: %#v", none.Fixes())
	}
	if suggestion.FixesPtr != nil {
		t.Fatalf("EditDemandSuggestion consumer received fixes: %#v", suggestion.Fixes())
	}
	if len(autofix.Fixes()) != 1 || autofix.Fixes()[0].Text != "a > 1" {
		t.Fatalf("autofix = %#v, want a single fix replacing with %q", autofix.Fixes(), "a > 1")
	}
	if len(all.Fixes()) != 1 || all.Fixes()[0].Text != "a > 1" {
		t.Fatalf("all-edits fix = %#v, want a single fix replacing with %q", all.Fixes(), "a > 1")
	}
}
