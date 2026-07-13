// TestNoUnnecessaryTypeConversionExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
//
// Every assertion in this file was differential-validated against
// typescript-eslint 8.x's `no-unnecessary-type-conversion` running under
// ESLint + @typescript-eslint/parser on the same input snippets. Cases noted
// `// upstream: fires` / `// upstream: no fire` were each confirmed by the
// reference run; deviations are called out where the rule's surface diverges
// for documented reasons (and there are none in this file).
package no_unnecessary_type_conversion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeConversionExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeConversionRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver wrappers — TS non-null assertion routing to a non-string-like type ----
			// `foo!.toString()` where foo!'s type is `number` — number→toString stays valid.
			{Code: `
declare const x: number | null;
x!.toString();
`},
			// ---- Dimension 4: optional chain on .toString receiver ----
			// `s?.toString()` on possibly-undefined receiver: union with undefined is NOT all string-like.
			{Code: `
declare const s: string | undefined;
s?.toString();
`},
			// ---- Dimension 4: union with non-string constituent — should NOT trigger ----
			// `String(x)` where x is `string | number` — only every-constituent string-like fires.
			{Code: `
declare const x: string | number;
String(x);
`},
			// ---- Dimension 4: element-access key form for the conversion call ----
			// `globalThis['String']('asdf')` does not match an Identifier callee → not reported.
			{Code: `globalThis['String']('asdf');`},
			// ---- Dimension 4: nested-scope shadow (block-scoped function) ----
			{Code: `
{
  function String(value: unknown) {
    return value;
  }
  String('asdf');
}
export {};
`},
			// ---- Dimension 4: type alias declared in SAME scope shadows (matches upstream scope.set) ----
			// Differential-validated: upstream does NOT fire — type aliases enter ESLint's
			// scope.set even though they erase at runtime.
			{Code: `
type String = string;
String('asdf');
export {};
`},
			// ---- Dimension 4: interface declared in same scope shadows ----
			// Differential-validated: upstream does NOT fire.
			{Code: `
interface String {
  foo(): void;
}
String('asdf');
export {};
`},
			// ---- Dimension 4: class declared in same scope shadows ----
			{Code: `
class String {}
new String();
export {};
`},
			// ---- Dimension 4: NoSubstitutionTemplateLiteral is NOT a Literal upstream ----
			// `'asdf' + ``;` differential-validated: upstream does NOT fire because the
			// `'' === node.right.value` check requires the right operand to be a Literal
			// node, not a TemplateLiteral. tsgo treats backtick template as a literal kind;
			// we explicitly opt out to stay aligned with upstream.
			{Code: `
declare const s: string;
const r = s + ` + "``" + `;
`},
			{Code: `
declare const s: string;
const r = ` + "``" + ` + s;
`},
			{Code: `
let m: string = '';
m += ` + "``" + `;
`},
			// ---- Locks in enum-member short-circuit: a single enum member access ----
			{Code: `
enum E {
  A = 'a',
}
E.A.toString();
`},
			// ---- Locks in upstream handleUnaryOperator integer branch: float NumberLiteral must not fire ----
			{Code: `
const x = 1.5;
~~x;
`},
			// ---- Locks in upstream integer branch: NumberLiteralType union including a non-integer ----
			{Code: `
declare const x: 1 | 1.5;
~~x;
`},
			// ---- Locks in upstream integer branch: `number` (general) is not a NumberLiteral, so ~~ stays valid ----
			{Code: `
declare const x: number;
~~x;
`},
			// ---- Locks in `Boolean()` requiring booleanLike: any-type argument is not booleanLike ----
			{Code: `
declare const x: any;
Boolean(x);
`},
			// ---- Locks in `+x` numberLike check: bigint is NOT numberLike, so +x stays valid ----
			{Code: `
declare const x: bigint;
+x;
`},
			// ---- Locks in `'' + x` rightTypeMatches branch: x must be all string-like ----
			{Code: `
declare const x: string | number;
'' + x;
`},
			// ---- Locks in `x + ''` leftTypeMatches branch ----
			{Code: `
declare const x: string | number;
x + '';
`},
			// ---- Locks in `+=` listener: left non-string forbids reporting ----
			{Code: `
let n = 1;
n += '';
`},
			// ---- Real-user: `obj?.x.toString()` where `obj?.x` resolves to nullable ----
			{Code: `
declare const obj: { x: number } | undefined;
obj?.x.toString();
`},

			// ---- `String(...arr)` SpreadElement — never a no-op so the rule
			// must stay silent. tsgo's `getConstrainedTypeAtLocation` for a
			// SpreadElement unwraps to the element type (`string`), which
			// without an explicit short-circuit would let the rule over-fire
			// against upstream. ----
			{Code: `
declare const arr: string[];
String(...arr);
`},
			{Code: `
declare const ns: number[];
Number(...ns);
`},

			// =================================================================
			// Sanity non-trigger forms — rule MUST stay silent (no false-fires)
			// All differential-validated against typescript-eslint 8.x.
			// =================================================================

			// ---- Tagged template `String\`abc\`` is NOT a CallExpression ----
			{Code: "const x = String`abc`;"},
			// ---- `String.call(null, 'x')` — callee is MemberExpression, not Identifier ----
			{Code: `const x = String.call(null, 'asdf');`},
			// ---- `Symbol('x')` — name not in the 4 watched conversion fns ----
			{Code: `const x = Symbol('asdf');`},
			// ---- `parseInt('123')` — name not watched ----
			{Code: `const x = parseInt('123');`},

			// =================================================================
			// Type-system edges (differential-validated upstream stays silent)
			// =================================================================

			// ---- `String(undefined)` — undefined is not StringLike ----
			{Code: `String(undefined);`},
			// ---- `String(null)` — null is not StringLike ----
			{Code: `String(null);`},
			// ---- `String(0)` — number is not StringLike ----
			{Code: `String(0);`},
			// ---- `Number(prompt)` — string is not NumberLike ----
			{Code: `
declare const ps: string;
Number(ps);
`},
			// ---- Branded string `string & {__brand}` is Intersection, not String ----
			// Verified: upstream does NOT fire. UnionTypeParts only splits unions,
			// so an Intersection's flags are checked as-is (no StringLike).
			{Code: `
type Brand = string & { __brand: 'x' };
declare const b: Brand;
String(b);
`},
			// ---- `'asdf' + 'b'` — neither operand is the empty literal ----
			{Code: `const x = 'asdf' + 'b';`},

			// =================================================================
			// Shadow edges — differential-validated against upstream's
			// LOCAL-scope-only check (see isLocallyShadowed docstring).
			// =================================================================

			// ---- Destructure shadow in same block ----
			{Code: `
declare const obj: { String: (v: unknown) => unknown };
{
  const { String } = obj;
  String('asdf');
}
`},
			// ---- Function parameter shadow ----
			{Code: `
function outer(String: (v: unknown) => unknown) {
  String('asdf');
}
export {};
`},
			// ---- Arrow parameter shadow ----
			{Code: `
const fn = (String: (v: unknown) => unknown) => String('asdf');
export {};
`},
			// ---- Function expression's own name shadow ----
			{Code: `
const fn = function String(v: unknown) {
  return String('asdf');
};
export {};
`},
			// ---- Same-scope const variable shadow ----
			{Code: `
const String = (v: unknown) => v;
String('asdf');
export {};
`},
			// ---- Same-scope namespace declaration shadow (creates value binding) ----
			{Code: `
namespace String {
  export const x = 1;
}
String;
export {};
`},

			// Config `off` un-declares the builtin `String` — no report.
			{
				Code: `
declare const s: string;
String(s);
`,
				Globals: map[string]bool{"String": false},
			},
			// Config `off` un-declares the builtin `Number` — no report.
			{
				Code: `
declare const n: number;
Number(n);
`,
				Globals: map[string]bool{"Number": false},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized argument of String() ----
			// tsgo wraps argument in ParenthesizedExpression; rslint follows upstream
			// (ESTree has no paren nodes) and inlines the unwrapped inner text.
			{
				Code: `String(('asdf'));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `'asdf';`},
							{MessageId: "suggestSatisfies", Output: `'asdf' satisfies string;`},
						},
					},
				},
			},
			// ---- Dimension 4: parenthesized CALLEE — `(String)('asdf')` ----
			// Differential-validated: upstream fires (its ESTree parser implicitly
			// strips parens around the callee). rslint must SkipParentheses on the
			// callee to match — verified.
			{
				Code: `(String)('asdf');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    2,
						EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `'asdf';`},
							{MessageId: "suggestSatisfies", Output: `'asdf' satisfies string;`},
						},
					},
				},
			},
			// ---- Dimension 4: optional call form `String?.('asdf')` ----
			// Differential-validated: upstream fires.
			{
				Code: `String?.('asdf');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `'asdf';`},
							{MessageId: "suggestSatisfies", Output: `'asdf' satisfies string;`},
						},
					},
				},
			},
			// ---- Dimension 4: template literal with substitution receiver of .toString() ----
			// Differential-validated: upstream fires.
			{
				Code: "`pre${0}`.toString();",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    11,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							// TemplateExpression is not strong-precedence per utils.IsStrongPrecedenceNode,
							// so the wrapping fixer adds parens around the bare receiver — matches upstream.
							{MessageId: "suggestRemove", Output: "(`pre${0}`);"},
							{MessageId: "suggestSatisfies", Output: "(`pre${0}`) satisfies string;"},
						},
					},
				},
			},
			// ---- Dimension 4: NoSubstitutionTemplateLiteral receiver of .toString() ----
			// Differential-validated: upstream fires (receiver type is stringLike).
			{
				Code: "`asdf`.toString();",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    8,
						EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: "`asdf`;"},
							{MessageId: "suggestSatisfies", Output: "`asdf` satisfies string;"},
						},
					},
				},
			},
			// ---- Locks in: top-level function declaration does NOT propagate as shadow into nested function ----
			// Differential-validated: upstream's scope.set check is LOCAL — outer
			// declarations do not shadow inside nested function/arrow/block scopes,
			// even though TS semantics would resolve the bare name to the outer.
			{
				Code: `
function String(v: unknown) {
  return v;
}
function inner() {
  String('asdf');
}
export {};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      6,
						Column:    3,
						EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
function String(v: unknown) {
  return v;
}
function inner() {
  'asdf';
}
export {};
`},
							{MessageId: "suggestSatisfies", Output: `
function String(v: unknown) {
  return v;
}
function inner() {
  'asdf' satisfies string;
}
export {};
`},
						},
					},
				},
			},
			// ---- Locks in: top-level type alias does NOT propagate as shadow into nested block ----
			{
				Code: `
type String = string;
if (1) {
  String('asdf');
}
export {};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      4,
						Column:    3,
						EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
type String = string;
if (1) {
  'asdf';
}
export {};
`},
							{MessageId: "suggestSatisfies", Output: `
type String = string;
if (1) {
  'asdf' satisfies string;
}
export {};
`},
						},
					},
				},
			},
			// ---- Locks in handleUnaryOperator BooleanLike branch on a literal-union (true | false) ----
			{
				Code: `
declare const b: true | false;
!!b;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    1,
						EndColumn: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const b: true | false;
b;
`},
							{MessageId: "suggestSatisfies", Output: `
declare const b: true | false;
b satisfies boolean;
`},
						},
					},
				},
			},
			// ---- Locks in `Boolean()` reporting on a boolean variable (no shadow) ----
			{
				Code: `
declare const b: boolean;
Boolean(b);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    1,
						EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const b: boolean;
b;
`},
							{MessageId: "suggestSatisfies", Output: `
declare const b: boolean;
b satisfies boolean;
`},
						},
					},
				},
			},
			// ---- Locks in upstream "node.right === ''" branch when the LHS is a property access ----
			{
				Code: `
declare const obj: { s: string };
obj.s + '';
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    6,
						EndLine:   3,
						EndColumn: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const obj: { s: string };
obj.s;
`},
							{MessageId: "suggestSatisfies", Output: `
declare const obj: { s: string };
obj.s satisfies string;
`},
						},
					},
				},
			},
			// ---- Locks in `node.left === ''` branch with an Identifier RHS ----
			{
				Code: `
declare const s: string;
'' + s;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    1,
						EndLine:   3,
						EndColumn: 6,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const s: string;
s;
`},
							{MessageId: "suggestSatisfies", Output: `
declare const s: string;
s satisfies string;
`},
						},
					},
				},
			},
			// ---- `!(!b)` with paren between the two `!` operators — upstream
			// sees no paren node and fires; rslint must SkipParentheses on the
			// outer operand to match. ----
			{
				Code: `
declare const b: boolean;
!(!b);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    1,
						EndColumn: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const b: boolean;
b;
`},
							{MessageId: "suggestSatisfies", Output: `
declare const b: boolean;
b satisfies boolean;
`},
						},
					},
				},
			},
			// ---- `~(~1)` — same paren-between fix. ----
			{
				Code: `~(~1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `1;`},
							{MessageId: "suggestSatisfies", Output: `1 satisfies number;`},
						},
					},
				},
			},
			// ---- `s += ('')` — empty literal wrapped in parens still counts.
			// upstream's ESTree drops parens around the literal, so the rule
			// sees `s += ''` and fires. ----
			{
				Code: `
let m = '';
m += ('');
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    1,
						EndLine:   3,
						EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
let m = '';

`},
							{MessageId: "suggestSatisfies", Output: `
let m = '';
m satisfies string;
`},
						},
					},
				},
			},
			// ---- `((('asdf'))) + ''` — outer paren chain on the left operand
			// must not shift the report's start column off the inner literal.
			// upstream column 10 (right after the `'asdf'` literal); rslint
			// uses SkipParentheses to find the same position. ----
			{
				Code: `((('asdf'))) + '';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    10,
						EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `'asdf';`},
							{MessageId: "suggestSatisfies", Output: `'asdf' satisfies string;`},
						},
					},
				},
			},
			// ---- `String(...arr)` with spread argument — never a no-op
			// (spread iterates and forwards only the first element). upstream
			// stays silent because TypeScript reports the spread's type as the
			// iterable, not its element type; rslint must short-circuit on
			// SpreadElement to avoid an over-fire on `string[]`. ----
			// This is a NEGATIVE test — the entry below sits in the VALID set
			// already (search for `SpreadElement`), so we don't need another
			// invalid one here.

			// ---- Locks in upstream's `Number.isInteger((t as NumberLiteralType).value)`
			// check: large integers tsgo stringifies as scientific notation (e.g.
			// `1e+21`) MUST still be treated as integers — earlier string-pattern
			// implementations false-negatived these. ----
			{
				Code: `~~1e21;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `1e21;`},
							{MessageId: "suggestSatisfies", Output: `1e21 satisfies number;`},
						},
					},
				},
			},
			// ---- Locks in `~~` branch when operand is a negative literal preserved via unary minus ----
			{
				Code: `~~-42;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `(-42);`},
							{MessageId: "suggestSatisfies", Output: `(-42) satisfies number;`},
						},
					},
				},
			},
			// ---- Locks in triple-bang `!!!x`: 2 adjacent `!!` pairs → 2 reports ----
			// Differential-validated: upstream fires twice on `!!!b`.
			{
				Code: `
declare const b: boolean;
!!!b;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    1,
						EndColumn: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							// Outer report fixes !!! → (!b) because the inner `!b` is a
							// PrefixUnaryExpression (not strong-precedence).
							{MessageId: "suggestRemove", Output: `
declare const b: boolean;
(!b);
`},
							{MessageId: "suggestSatisfies", Output: `
declare const b: boolean;
(!b) satisfies boolean;
`},
						},
					},
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    2,
						EndColumn: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const b: boolean;
!b;
`},
							{MessageId: "suggestSatisfies", Output: `
declare const b: boolean;
!(b satisfies boolean);
`},
						},
					},
				},
			},
			// ---- Real-user: nested unnecessary conversions in the same expression must each report ----
			{
				Code: `String(String('asdf'));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `String('asdf');`},
							{MessageId: "suggestSatisfies", Output: `String('asdf') satisfies string;`},
						},
					},
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    8,
						EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `String('asdf');`},
							{MessageId: "suggestSatisfies", Output: `String('asdf' satisfies string);`},
						},
					},
				},
			},

			// =================================================================
			// Chained binary `+`: each `+` operator is its own potential match
			// (one report per matching pair). Differential-validated.
			// =================================================================

			// ---- `'' + s + ''` — parses `('' + s) + ''`. Both inner (left-empty)
			// and outer (right-empty) fire. rslint emits outer first, inner
			// second (top-down visit); upstream sorts by position so its
			// display order is opposite — the diagnostic SET is identical. ----
			{
				Code: `
declare const s: string;
const c = '' + s + '';
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						// Outer (right-empty branch): ` + ''` covers cols 17-21.
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    17,
						EndLine:   3,
						EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const s: string;
const c = ('' + s);
`},
							{MessageId: "suggestSatisfies", Output: `
declare const s: string;
const c = ('' + s) satisfies string;
`},
						},
					},
					{
						// Inner (left-empty branch): `'' + ` covers cols 11-15.
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    11,
						EndLine:   3,
						EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const s: string;
const c = s + '';
`},
							{MessageId: "suggestSatisfies", Output: `
declare const s: string;
const c = (s satisfies string) + '';
`},
						},
					},
				},
			},
			// ---- `'a' + '' + 'b'` — only the inner `'' + 'b'` middle? Actually
			// parses as `('a' + '') + 'b'`. Outer: right is 'b' (not empty) → no
			// fire. Inner: right is '' → fires.
			{
				Code: `const c = 'a' + '' + 'b';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    14,
						EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `const c = 'a' + 'b';`},
							{MessageId: "suggestSatisfies", Output: `const c = ('a' satisfies string) + 'b';`},
						},
					},
				},
			},
			// ---- `'' + ''` — right-empty branch fires first, only 1 report ----
			{
				Code: `const c = '' + '';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    13,
						EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `const c = '';`},
							{MessageId: "suggestSatisfies", Output: `const c = '' satisfies string;`},
						},
					},
				},
			},

			// =================================================================
			// Nested type-conversion calls
			// =================================================================

			// ---- `String(Number(n))` — only inner Number() fires (outer's arg
			// type is `number`, not StringLike). ----
			{
				Code: `
declare const n: number;
const c = String(Number(n));
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    18,
						EndLine:   3,
						EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const n: number;
const c = String(n);
`},
							{MessageId: "suggestSatisfies", Output: `
declare const n: number;
const c = String(n satisfies number);
`},
						},
					},
				},
			},

			// ---- `String(typeof 1)` — `typeof X` returns a string-literal-typed
			// expression which IS StringLike → fires. TypeOfExpression is not
			// strong-precedence per utils.IsStrongPrecedenceNode, so the
			// wrapping fixer adds parens (matches upstream output). ----
			{
				Code: `const c = String(typeof 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    11,
						EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `const c = (typeof 1);`},
							{MessageId: "suggestSatisfies", Output: `const c = (typeof 1) satisfies string;`},
						},
					},
				},
			},

			// =================================================================
			// Container forms — type-conversion call appearing in various
			// container shapes. Each is differential-validated and produces
			// the same diagnostic location as a bare call.
			// =================================================================

			// ---- Multi-line String call: trim collapses inner newlines ----
			{
				Code: `const c = String(
  'long string'
);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    11,
						EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `const c = 'long string';`},
							{MessageId: "suggestSatisfies", Output: `const c = 'long string' satisfies string;`},
						},
					},
				},
			},
			// ---- Class field initializer ----
			{
				Code: `
class K {
  x = String('asdf');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    7,
						EndColumn: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
class K {
  x = 'asdf';
}
`},
							{MessageId: "suggestSatisfies", Output: `
class K {
  x = 'asdf' satisfies string;
}
`},
						},
					},
				},
			},
			// ---- Class static field ----
			{
				Code: `
class K {
  static s = String('asdf');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    14,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
class K {
  static s = 'asdf';
}
`},
							{MessageId: "suggestSatisfies", Output: `
class K {
  static s = 'asdf' satisfies string;
}
`},
						},
					},
				},
			},
			// ---- Default parameter ----
			{
				Code: `function f(x: string = String('asdf')) { return x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    24,
						EndColumn: 30,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `function f(x: string = 'asdf') { return x; }`},
							{MessageId: "suggestSatisfies", Output: `function f(x: string = 'asdf' satisfies string) { return x; }`},
						},
					},
				},
			},
			// ---- Object literal value ----
			{
				Code: `
declare const s: string;
const c = { key: String(s) };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    18,
						EndLine:   3,
						EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const s: string;
const c = { key: s };
`},
							{MessageId: "suggestSatisfies", Output: `
declare const s: string;
const c = { key: s satisfies string };
`},
						},
					},
				},
			},
			// ---- Array literal element ----
			{
				Code: `
declare const s: string;
const c = [String(s)];
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    12,
						EndLine:   3,
						EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const s: string;
const c = [s];
`},
							{MessageId: "suggestSatisfies", Output: `
declare const s: string;
const c = [s satisfies string];
`},
						},
					},
				},
			},
			// ---- Template-string interpolation ----
			{
				Code: "declare const s: string;\nconst c = `pre${String(s)}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      2,
						Column:    17,
						EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: "declare const s: string;\nconst c = `pre${s}`;"},
							{MessageId: "suggestSatisfies", Output: "declare const s: string;\nconst c = `pre${s satisfies string}`;"},
						},
					},
				},
			},
			// ---- Deeply nested parens on callee `(((String)))('x')` ----
			{
				Code: `const c = (((String)))('asdf');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    14,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `const c = 'asdf';`},
							{MessageId: "suggestSatisfies", Output: `const c = 'asdf' satisfies string;`},
						},
					},
				},
			},
			// ---- catch-clause shadow does NOT enter body's local scope.set
			// (upstream local-only scope quirk — fires) ----
			{
				Code: `
try {} catch (String) {
  String('asdf');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    3,
						EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
try {} catch (String) {
  'asdf';
}
`},
							{MessageId: "suggestSatisfies", Output: `
try {} catch (String) {
  'asdf' satisfies string;
}
`},
						},
					},
				},
			},
			// ---- for-of binding does NOT enter body's local scope.set
			// (upstream local-only scope quirk — fires) ----
			{
				Code: `
declare const arr: any[];
for (const String of arr) {
  String('asdf');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      4,
						Column:    3,
						EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const arr: any[];
for (const String of arr) {
  'asdf';
}
`},
							{MessageId: "suggestSatisfies", Output: `
declare const arr: any[];
for (const String of arr) {
  'asdf' satisfies string;
}
`},
						},
					},
				},
			},

			// =================================================================
			// Type-system edges that DO fire
			// =================================================================

			// ---- Template literal type as argument — constrained type widens
			// to string → fires. ----
			{
				Code: `
declare const t: ` + "`pre${string}`" + `;
String(t);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    1,
						EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const t: ` + "`pre${string}`" + `;
t;
`},
							{MessageId: "suggestSatisfies", Output: `
declare const t: ` + "`pre${string}`" + `;
t satisfies string;
`},
						},
					},
				},
			},
		},
	)
}
