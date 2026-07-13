// TestPreferFindExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
package prefer_find

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferFindExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferFindRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: Receiver wrappers ----

		// Non-array receiver behind type assertion — type of the receiver is
		// determined by the assertion, which is not arrayish.
		{Code: `
interface Box<T> { filter(p: (item: T) => boolean): Box<T> }
declare const box: Box<string>;
(box as Box<string>).filter(x => true)[0];
`},
		// `satisfies` keeps the value's apparent type; if it's not arrayish,
		// the rule must not fire.
		{Code: `
interface Box<T> { filter(p: (item: T) => boolean): Box<T> }
declare const box: Box<string>;
(box satisfies Box<string>).filter(x => true)[0];
`},

		// ---- Dimension 4: Access / key forms ----

		// `.at('1')` — Number('1') === 1, not zero.
		{Code: `[1, 2, 3].filter(x => x > 0).at('1');`},
		// `.at('')` — Number('') === 0, treated as zero. We DO fire here,
		// matching JS semantics. (Negative — included for symmetry.)
		// N/A: this row is actually invalid; covered below in invalid cases.

		// `[1]` numeric subscript — not zero.
		{Code: `[1, 2, 3].filter(x => x > 0)[1];`},
		// `[`1`]` template-1 subscript — not zero.
		{Code: "[1, 2, 3].filter(x => x > 0)[`1`];"},
		// `[` ` `]` whitespace string — String(' ') === ' ', not '0'.
		{Code: `[1, 2, 3].filter(x => x > 0)[' '];`},
		// Whole-line subscript with unresolvable identifier — neither
		// resolveStaticString nor resolveStaticNumber can ground it.
		{Code: `
declare const idx: number;
[1, 2, 3].filter(x => x > 0)[idx];
`},

		// ---- Dimension 4: Graceful degradation ----

		// `.at()` with zero arguments — args.length !== 1 short-circuits
		// getObjectIfArrayAtZero. (.at requires one arg at runtime; this is
		// a syntactic check.)
		{Code: `
declare const arr: string[];
(arr.filter(x => x.length > 0) as any).at();
`},
		// `.at(0, 1)` with two arguments — args.length !== 1.
		{Code: `
declare const arr: string[];
(arr.filter(x => x.length > 0) as any).at(0, 1);
`},
		// `.at(...spread)` — args.length is 1 (the SpreadElement counts as
		// one argument node), but the value can't be statically determined,
		// so resolveStaticNumber/String both fail → not zero.
		{Code: `
declare const arr: string[];
declare const idx: [number];
arr.filter(x => x.length > 0).at(...idx);
`},
		// `[at]` where `at` is unresolvable — locks in the symmetric
		// behavior for the [0] path: subscript that doesn't resolve to '0'
		// must not match.
		{Code: `
declare const arr: string[];
declare const at: any;
arr.filter(x => x.length > 0)[at];
`},

		// ---- Branch lock-ins ----

		// Locks in parseArrayFilterExpressions: callee is not a member
		// access at all (e.g. `filter(arr)[0]`) — must not match.
		{Code: `
declare function filter(arr: unknown[]): unknown[];
filter([1, 2, 3])[0];
`},
		// Locks in parseArrayFilterExpressions: same kind, but called as a
		// bare local function (not `<obj>.filter`).
		{Code: `
const filter = (arr: number[]) => arr.filter(x => x > 0);
filter([1, 2, 3])[0];
`},
		// Locks in isArrayish: intersection contains a non-array member,
		// so the type is rejected even though one side is array-shaped.
		{Code: `
interface Box<T> { filter(p: (x: T) => boolean): Box<T>; length: number }
declare const arr: { a: 1 }[] & Box<{ a: 1 }>;
arr.filter(x => true)[0];
`},
		// Locks in resolveConstInitializer: `let` binding, not `const` —
		// initializer cannot be statically inlined.
		{Code: `
declare const arr: string[];
let zero = 0;
arr.filter(item => item === 'aha').at(zero);
`},
		// Locks in resolveConstInitializer: parameter — has no const
		// initializer, can't be inlined.
		{Code: `
declare const arr: string[];
function check(zero: number) {
  return arr.filter(item => item === 'aha').at(zero);
}
check(0);
`},
		// Locks in resolveStaticString cycle guard: `const a = a` would loop
		// forever without the `seen` map. Use a const referring to itself
		// (legal at TDZ — won't compile-run, but the lint pass mustn't hang).
		{Code: `
declare const arr: string[];
// @ts-expect-error: TDZ self-reference
const a = a;
arr.filter(x => x.length > 0)[a];
`},
		// Locks in `NaN`/`Infinity` shadowing — if the program declares
		// its own `NaN`/`Infinity`, we must NOT treat the reference as the
		// global. Here `NaN` is shadowed by a function parameter; without
		// the shadow check we would (incorrectly) treat `.at(NaN)` as zero.
		{Code: `
declare const arr: string[];
function check(NaN: number) {
  return arr.filter(x => x.length > 0).at(NaN);
}
`},
		// Locks in the TypeChecker-driven Symbol short-circuit: any
		// argument whose static TYPE is `symbol` (or `unique symbol` or
		// `typeof Symbol.foo`) bypasses static-value resolution and is
		// treated as not-zero. Mirrors upstream's `typeof value === 'symbol'`
		// branch in `isTreatedAsZeroByArrayAt`. Without the type-info
		// shortcut, `Number(Symbol(...))` would throw at runtime, so the
		// rule must not rewrite.
		{Code: `
declare const arr: string[];
declare const s: symbol;
arr.filter(x => x.length > 0).at(s);
`},
		// Same shortcut for the `[s]` subscript path
		// (isTreatedAsZeroByMemberAccess).
		{Code: `
declare const arr: string[];
declare const s: symbol;
arr.filter(x => x.length > 0)[s];
`},
		// `unique symbol` typed via `typeof`-of-a-Symbol-call — still
		// symbol-like via TypeFlagsESSymbolLike.
		{Code: `
declare const arr: string[];
const sym = Symbol('zero');
arr.filter(x => x.length > 0).at(sym);
`},
		// ---- Primitive globals — non-zero canonicalization ----
		// `[null]` — String(null) === 'null' ≠ '0'.
		{Code: `
declare const arr: string[];
arr.filter(x => x.length > 0)[null as any];
`},
		// `[true]` — String(true) === 'true' ≠ '0'.
		{Code: `
declare const arr: string[];
arr.filter(x => x.length > 0)[true as any];
`},
		// `[false]` — String(false) === 'false' ≠ '0'.
		{Code: `
declare const arr: string[];
arr.filter(x => x.length > 0)[false as any];
`},
		// `[NaN]` — String(NaN) === 'NaN' ≠ '0'.
		{Code: `
declare const arr: string[];
arr.filter(x => x.length > 0)[NaN];
`},
		// `[Infinity]` / `[-Infinity]` — strings ≠ '0'.
		{Code: `
declare const arr: string[];
arr.filter(x => x.length > 0)[Infinity];
`},
		{Code: `
declare const arr: string[];
arr.filter(x => x.length > 0)[-Infinity];
`},
		// `.at(true)` — Number(true) === 1, not zero.
		{Code: `declare const arr: string[]; arr.filter(x => x.length > 0).at(true);`},
		// `.at(Infinity)` — Number(Infinity) === Infinity, trunc !== 0.
		{Code: `declare const arr: string[]; arr.filter(x => x.length > 0).at(Infinity);`},

		// Config `off` un-declares the builtin `undefined`/`NaN`/`Infinity` — should not resolve to zero.
		{
			Code: `
declare const arr: string[];
arr.filter(x => x.length > 0).at(undefined);
`,
			Globals: map[string]bool{"undefined": false},
		},
		{
			Code: `
declare const arr: string[];
arr.filter(x => x.length > 0).at(NaN);
`,
			Globals: map[string]bool{"NaN": false},
		},
		{
			Code: `
declare const arr: string[];
arr.filter(x => x.length > 0).at(Infinity);
`,
			Globals: map[string]bool{"Infinity": false},
		},
		// ---- jsNumberFromString lock-ins (hex / octal / binary STRING args) ----
		// `.at('0x1')` — JS Number('0x1') === 1 (hex prefix accepted by JS
		// string coercion). trunc=1 ≠ 0 → must NOT fire. Locks in the
		// strconv.ParseInt(_, 0, 64) hex/octal/binary fallback in
		// jsNumberFromString; without it, Go's ParseFloat would fail on
		// `0x1` (it only accepts hex floats with a `p` exponent) and the
		// resulting NaN would be (incorrectly) treated as zero.
		{Code: `declare const arr: string[]; arr.filter(x => x.length > 0).at('0x1');`},
		{Code: `declare const arr: string[]; arr.filter(x => x.length > 0).at('0o1');`},
		{Code: `declare const arr: string[]; arr.filter(x => x.length > 0).at('0b1');`},
		// `.at('1.0')` — Number('1.0') === 1, not zero (regression guard).
		{Code: `declare const arr: string[]; arr.filter(x => x.length > 0).at('1.0');`},
		// `.at('Infinity')` — Number('Infinity') === Infinity (parsed by
		// ParseFloat). trunc=Inf ≠ 0 → must NOT fire.
		{Code: `declare const arr: string[]; arr.filter(x => x.length > 0).at('Infinity');`},
		// `.at('-0x1')` — JS rejects sign on hex strings (Number('-0x1') =
		// NaN). Our hex fallback gates on trimmed[0]=='0', so signs are
		// rejected and we return NaN — which is treated as zero, so this
		// fires. Same outcome as JS via NaN propagation; documented for
		// future readers.
		// (Negative case below covers the NaN → fire path.)
		// ---- Non-zero const numeric — sanity check that we don't over-fire ----
		{Code: `
declare const arr: string[];
const one = 1;
arr.filter(x => x.length > 0).at(one);
`},
		{Code: `
declare const arr: string[];
const one = 1;
arr.filter(x => x.length > 0)[one];
`},
		// ---- Non-zero BigInt — Number(5n) === 5, trunc !== 0. ----
		{Code: `declare const arr: string[]; arr.filter(x => x.length > 0).at(5n);`},
		{Code: `declare const arr: string[]; arr.filter(x => x.length > 0)[5n];`},
		// ---- Unresolvable expression — runtime length, not statically zero ----
		{Code: `
declare const arr: string[];
arr.filter(x => x.length > 0)[arr.length - 1];
`},
		// ---- Iterable / non-array receiver (TypedArray) ----
		// Uint8Array is not an Array instance via Checker_isArrayOrTupleType.
		{Code: `
declare const ua: Uint8Array;
(ua as any).filter((x: number) => x > 0)[0];
`},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: Receiver wrappers ----

		// Locks in tsgo ParenthesizedExpression handling — upstream doesn't
		// need to test this because ESTree drops parens; tsgo keeps them as
		// explicit nodes and our SkipParentheses must unwrap them.
		{
			Code: `
declare const arr: string[];
(arr).filter(x => x.length > 0)[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
(arr).find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
		// Multi-level parens around the receiver.
		{
			Code: `
declare const arr: string[];
((arr)).filter(x => x.length > 0)[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
((arr)).find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
		// Non-null assertion on the receiver — `arr!` is arrayish if `arr` is.
		{
			Code: `
declare const arr: string[] | undefined;
arr!.filter(x => x.length > 0)[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[] | undefined;
arr!.find(x => x.length > 0);
`,
						},
					},
				},
			},
		},

		// ---- Dimension 4: Access / key forms ----

		// `.at('')` — JS Number('') === 0, treated as zero. Upstream's
		// helper resolves the empty string via Number(value) -> 0, so this
		// must fire. Locks in our jsNumberFromString empty-string branch.
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha').at('');
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// `.at('abc')` — Number('abc') is NaN, isNaN(NaN) === true → treated
		// as zero. JS semantics, not intuitive — lock it in.
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha').at('abc');
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// `[0x0]` — different numeric literal forms. tsgo's parser
		// normalizes NumericLiteral.Text to decimal at parse time
		// (`0x0`.Text == "0"), so the regular numeric path handles this.
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha')[0x0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// `.at('0x0')` — Number('0x0') === 0 → fire. Locks in
		// jsNumberFromString's hex fallback (ParseFloat rejects bare
		// `0x...` so we fall through to ParseInt with base 0).
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha').at('0x0');
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// `.at('0o0')` — same hex/octal/binary fallback path.
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha').at('0o0');
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// `.at('0b0')` — same hex/octal/binary fallback path.
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha').at('0b0');
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// `.at('1_000')` — JS Number() rejects `_` separators
		// (`Number('1_000') === NaN`), so isNaN → true → fire. Go's
		// ParseFloat accepts `_`, so we need an explicit guard for parity.
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha').at('1_000');
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// `[-0]` — String(-0) === '0' in JS. Lock in unary-minus-zero.
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha')[-0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// `.at(NaN)` shadowed by const `NaN = NaN` (still resolves to the
		// global since the const just aliases NaN). Sanity-check the const
		// resolver doesn't break on shadowing identifiers.
		// N/A: this is in the valid set above (the shadowing case).

		// ---- Branch lock-ins for parseArrayFilterExpressions ----

		// Ternary where one arm is a sub-ternary whose branches are filter
		// calls — confirms recursion through nested ConditionalExpression.
		{
			Code: `
declare const arr1: string[], arr2: string[], arr3: string[], pick: boolean;
(pick ? (pick ? arr1.filter(f) : arr2.filter(f)) : arr3.filter(f))[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr1: string[], arr2: string[], arr3: string[], pick: boolean;
(pick ? (pick ? arr1.find(f) : arr2.find(f)) : arr3.find(f));
`,
						},
					},
				},
			},
		},

		// ---- Branch lock-ins for getObjectIfArrayAtZeroExpression ----

		// `.at(0n)` — BigInt zero, Number(0n) === 0 → treated as zero. Lock
		// in the BigInt branch of resolveStaticNumber.
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha').at(0n);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// Locks in upstream getObjectIfArrayAtZeroExpression: it only gates
		// on `!callee.optional` (the member access `?.at` is excluded), NOT
		// on `node.optional` (the call's own `?.()`). So `<obj>.at?.(0)`
		// — call optional, member non-optional — still fires. Easy to
		// over-tighten by checking both, so a lock-in here.
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha').at?.(0);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// `.at(-0)` — unary minus on numeric zero, trunc(-0) === 0 → fires.
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha').at(-0);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},

		// ---- Real-user: tuple receiver ----
		// Tuple types are also array-ish via Checker_isTupleType. Closed
		// issues against upstream's prefer-find frequently involve tuples.
		{
			Code: `
declare const tuple: readonly [string, string, string];
tuple.filter(x => x.length > 0)[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const tuple: readonly [string, string, string];
tuple.find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
		// ---- Real-user: method-chain receiver ----
		// Common idiom: a getter / method returning the array, then filter.
		// Locks in that `<call>.filter(...)[0]` works as long as the call's
		// return type is arrayish.
		{
			Code: `
declare function getArr(): string[];
getArr().filter(x => x.length > 0)[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare function getArr(): string[];
getArr().find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
		// ---- Primitive globals: `.at(null)`, `.at(false)` — Number → 0 ----
		// Number(null) === 0 → treated as zero. ESLint's getStaticValue
		// recognizes null literals and reports; we mirror via the
		// KindNullKeyword resolver branch.
		{
			Code: `
declare const arr: string[];
arr.filter(x => x.length > 0).at(null as any);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
		// Number(false) === 0 → treated as zero.
		{
			Code: `
declare const arr: string[];
arr.filter(x => x.length > 0).at(false as any);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
		// `.at(undefined)` — Number(undefined) === NaN, isNaN(NaN) === true
		// → treated as zero. The `undefined` global identifier (when not
		// shadowed) is recognized by resolveStaticStringRec, falls through
		// to jsNumberFromString("undefined") → NaN.
		{
			Code: `
declare const arr: string[];
arr.filter(x => x.length > 0).at(undefined as any);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
		// ---- Decimal fractional `.at` — Math.trunc rounds toward zero ----
		// `.at(0.5)` — trunc(0.5) === 0 → zero. JS rounds-toward-zero on
		// fractional indices, so the rule fires. (Upstream's
		// `-0.12635678` case covers this, but adding a forward `0.5`
		// locks the positive-fractional path.)
		{
			Code: `
declare const arr: string[];
arr.filter(x => x.length > 0).at(0.5);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
		// ---- Type assertions transparently unwrapped ----
		// `.at(0 as number)` — SkipOuterExpressions unwraps the assertion.
		{
			Code: `
declare const arr: string[];
arr.filter(x => x.length > 0).at(0 as number);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
		// `[0 satisfies number]` — satisfies is transparent.
		{
			Code: `
declare const arr: string[];
arr.filter(x => x.length > 0)[0 satisfies number];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
		// Non-null assertion `x!` on a const-resolved zero — also transparent.
		{
			Code: `
declare const arr: string[];
const zero: number | undefined = 0;
arr.filter(x => x.length > 0).at(zero!);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      4,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
const zero: number | undefined = 0;
arr.find(x => x.length > 0);
`,
						},
					},
				},
			},
		},
	})
}
