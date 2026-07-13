package prefer_nullish_coalescing

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Mirror upstream's two parametric generators —
//   types        = ['string', 'number', 'boolean', 'object']
//   nullishTypes = ['null', 'undefined', 'null | undefined']
//   equals       = ['', '=']
//
// `typeValidTest` produces all `(type, equals)` combinations.
// `nullishTypeTest` produces all `(nullish, type, equals)` combinations.

var primitiveTypes = []string{"string", "number", "boolean", "object"}
var nullishTypes = []string{"null", "undefined", "null | undefined"}

// typeValid expands a template `cb(type, equals)` over types × equals.
func typeValid(cb func(typ, eq string) string) []rule_tester.ValidTestCase {
	out := make([]rule_tester.ValidTestCase, 0, len(primitiveTypes)*2)
	for _, t := range primitiveTypes {
		out = append(out, rule_tester.ValidTestCase{Code: cb(t, "")})
	}
	for _, t := range primitiveTypes {
		out = append(out, rule_tester.ValidTestCase{Code: cb(t, "=")})
	}
	return out
}

// nullishTypeValid expands a template over nullish × type × equals.
func nullishTypeValid(cb func(nullish, typ, eq string) string) []rule_tester.ValidTestCase {
	out := make([]rule_tester.ValidTestCase, 0, len(nullishTypes)*len(primitiveTypes)*2)
	for _, n := range nullishTypes {
		for _, t := range primitiveTypes {
			for _, eq := range []string{"", "="} {
				out = append(out, rule_tester.ValidTestCase{Code: cb(n, t, eq)})
			}
		}
	}
	return out
}

// nullishTypeValidWithOpts wraps nullishTypeValid + per-case options.
func nullishTypeValidWithOpts(opts any, cb func(nullish, typ, eq string) string) []rule_tester.ValidTestCase {
	out := make([]rule_tester.ValidTestCase, 0, len(nullishTypes)*len(primitiveTypes)*2)
	for _, n := range nullishTypes {
		for _, t := range primitiveTypes {
			for _, eq := range []string{"", "="} {
				out = append(out, rule_tester.ValidTestCase{Code: cb(n, t, eq), Options: opts})
			}
		}
	}
	return out
}

// nullishTypeInvalid expands a template over nullish × type × equals where
// each case is an invalid case with a single suggestion. `codeFn(nullish, typ,
// eq)` produces the input code; `outputFn(nullish, typ, eq)` produces the
// suggestion output.
func nullishTypeInvalid(
	codeFn func(nullish, typ, eq string) string,
	outputFn func(nullish, typ, eq string) string,
	column, endColumn, line, endLine int,
	messageId string,
	options any,
) []rule_tester.InvalidTestCase {
	out := make([]rule_tester.InvalidTestCase, 0, len(nullishTypes)*len(primitiveTypes)*2)
	for _, n := range nullishTypes {
		for _, t := range primitiveTypes {
			for _, eq := range []string{"", "="} {
				ec := endColumn
				if ec > 0 {
					ec += len(eq)
				}
				out = append(out, rule_tester.InvalidTestCase{
					Code:    codeFn(n, t, eq),
					Options: options,
					Errors: []rule_tester.InvalidTestCaseError{{
						MessageId: messageId, Line: line, EndLine: endLine, Column: column, EndColumn: ec,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNullish", Output: outputFn(n, t, eq)},
						},
					}},
				})
			}
		}
	}
	return out
}

func mustConcat(slices ...[]rule_tester.ValidTestCase) []rule_tester.ValidTestCase {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	out := make([]rule_tester.ValidTestCase, 0, total)
	for _, s := range slices {
		out = append(out, s...)
	}
	return out
}

func mustConcatInvalid(slices ...[]rule_tester.InvalidTestCase) []rule_tester.InvalidTestCase {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	out := make([]rule_tester.InvalidTestCase, 0, total)
	for _, s := range slices {
		out = append(out, s...)
	}
	return out
}

func TestPreferNullishCoalescingRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule, buildValid(), buildInvalid())
}

func buildValid() []rule_tester.ValidTestCase {
	// ──────────────────────────────────────────────────────────────────────
	// Valid 1: non-nullable receiver — `||` and `||=` kept as-is.
	// ──────────────────────────────────────────────────────────────────────
	nonNullableOr := typeValid(func(typ, eq string) string {
		return fmt.Sprintf("\ndeclare let x: %s;\n(x ||%s 'foo');\n", typ, eq)
	})

	// Valid 2: nullable receiver with `??` / `??=` already in use.
	alreadyNullish := nullishTypeValid(func(nullish, typ, eq string) string {
		return fmt.Sprintf("\ndeclare let x: %s | %s;\nx ??%s 'foo';\n", typ, nullish, eq)
	})

	// Valid 3: ignoreTernaryTests permutations that don't form a clean check.
	ternaryNotClean := []rule_tester.ValidTestCase{
		{Code: `x !== undefined && x !== null ? x : y;`, Options: opt("ignoreTernaryTests", true)},
		{Code: `x !== undefined && x !== null ? 'foo' : 'bar';`, Options: optTern(false)},
		{Code: `x !== null && x !== undefined && x !== 5 ? x : y;`, Options: optTern(false)},
		{Code: `x === null || x === undefined || x === 5 ? x : y;`, Options: optTern(false)},
		{Code: `x === undefined && x !== null ? x : y;`, Options: optTern(false)},
		{Code: `x === undefined && x === null ? x : y;`, Options: optTern(false)},
		{Code: `x !== undefined && x === null ? x : y;`, Options: optTern(false)},
		{Code: `x === undefined || x !== null ? x : y;`, Options: optTern(false)},
		{Code: `x === undefined || x === null ? x : y;`, Options: optTern(false)},
		{Code: `x !== undefined || x === null ? x : y;`, Options: optTern(false)},
		{Code: `x !== undefined || x === null ? y : x;`, Options: optTern(false)},
		{Code: `x === null || x === null ? y : x;`, Options: optTern(false)},
		{Code: `x === undefined || x === undefined ? y : x;`, Options: optTern(false)},
		{Code: `x == null ? x : y;`, Options: optTern(false)},
		{Code: `undefined == null ? x : y;`, Options: optTern(false)},
		{Code: `undefined != z ? x : y;`, Options: optTern(false)},
		{Code: `x == undefined ? x : y;`, Options: optTern(false)},
		{Code: `x != null ? y : x;`, Options: optTern(false)},
		{Code: `x != undefined ? y : x;`, Options: optTern(false)},
		{Code: `null == x ? x : y;`, Options: optTern(false)},
		{Code: `undefined == x ? x : y;`, Options: optTern(false)},
		{Code: `null != x ? y : x;`, Options: optTern(false)},
		{Code: `undefined != x ? y : x;`, Options: optTern(false)},
		// Test contains a non-null/undefined sentinel.
		{Code: `
declare let x: number | undefined;
x !== 15 && x !== undefined ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: number | undefined;
x !== undefined && x !== 15 ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: number | undefined;
15 !== x && undefined !== x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: number | undefined;
undefined !== x && 15 !== x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: number | undefined;
15 !== x && x !== undefined ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: number | undefined;
undefined !== x && x !== 15 ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: string | undefined;
x !== 'foo' && x !== undefined ? x : y;
`, Options: optTern(false)},
		{Code: `
function test(value: number | undefined): number {
  return value !== foo() && value !== undefined ? value : 1;
}
`, Options: optTern(false)},
		{Code: `
const test = (value: boolean | undefined): boolean =>
  value !== undefined && value !== false ? value : false;
`, Options: optTern(false)},
		// non-nullable receiver — single-side checks ineligible.
		{Code: `
declare let x: string;
x === null ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: string | undefined | null;
x === null ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: string | undefined | null;
x === undefined ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: string | undefined | null;
x !== null ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: string | undefined | null;
x !== undefined ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: any;
x === null ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: unknown;
x === null ? x : y;
`, Options: optTern(false)},
		// Truthy ternary on non-nullable primitives — covers all primitives.
		{Code: `
declare let x: string;
x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: string;
!x ? y : x;
`, Options: optTern(false)},
		{Code: `
declare let x: string | object;
x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: string | object;
!x ? y : x;
`, Options: optTern(false)},
		{Code: `
declare let x: number;
x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: number;
!x ? y : x;
`, Options: optTern(false)},
		{Code: `
declare let x: bigint;
x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: bigint;
!x ? y : x;
`, Options: optTern(false)},
		{Code: `
declare let x: boolean;
x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: boolean;
!x ? y : x;
`, Options: optTern(false)},
		{Code: `
declare let x: object;
x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: object;
!x ? y : x;
`, Options: optTern(false)},
		{Code: `
declare let x: string[];
x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: string[];
!x ? y : x;
`, Options: optTern(false)},
		{Code: `
declare let x: Function;
x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: Function;
!x ? y : x;
`, Options: optTern(false)},
		{Code: `
declare let x: () => string;
x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: () => string;
!x ? y : x;
`, Options: optTern(false)},
		// Function call — both branches the same call expression.
		{Code: `
declare let x: () => string | null;
x() ? x() : y;
`, Options: optTern(false)},
		{Code: `
declare let x: () => string | null;
!x() ? y : x();
`, Options: optTern(false)},
		{Code: `
const a = 'foo';
declare let x: (a: string | null) => string | null;
x(a) ? x(a) : y;
`, Options: optTern(false)},
		{Code: `
const a = 'foo';
declare let x: (a: string | null) => string | null;
!x(a) ? y : x(a);
`, Options: optTern(false)},
		// Member access on non-nullable container.
		{Code: `
declare let x: { n: string };
x.n ? x.n : y;
`, Options: optTern(false)},
		{Code: `
declare let x: { n: string };
!x.n ? y : x.n;
`, Options: optTern(false)},
		{Code: `
declare let x: { n: string | object };
x.n ? x.n : y;
`, Options: optTern(false)},
		{Code: `
declare let x: { n: number };
x.n ? x.n : y;
`, Options: optTern(false)},
		// Function-typed member.
		{Code: `
declare let x: { n: () => string | null | undefined };
x.n ? x.n : y;
`, Options: optTern(false)},
	}

	// Valid 4: if-statements that don't fit the ??= pattern.
	ifNoFit := []rule_tester.ValidTestCase{
		{Code: `
declare let foo: string;
declare function makeFoo(): string;

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
`, Options: optTern(false)},
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo) {
    foo = makeFoo();
  }
}
`, Options: optTern(false)},
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo != null) {
    foo = makeFoo();
  }
}
`, Options: optTern(false)},
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
    return foo;
  }
}
`, Options: optTern(false)},
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  } else {
    return 'bar';
  }
}
`, Options: optTern(false)},
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  } else if (foo.a) {
    return 'bar';
  }
}
`, Options: optTern(false)},
		// Body is a block declaring a shadow — not an assignment.
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };
function shadowed() {
  if (foo == null) {
    const foo = makeFoo();
  }
}
`, Options: optTern(false)},
		// Destructuring assignment LHS — not member-access-like.
		{Code: `
declare let foo: { foo: string } | null;
declare function makeFoo(): { foo: { foo: string } };
function weirdDestructuringAssignment() {
  if (foo == null) {
    ({ foo } = makeFoo());
  }
}
`, Options: optTern(false)},
		// Optional chain / non-similar member compare.
		{Code: `
const a = 'b';
declare let x: { a: string; b: string } | null;
x?.a != null ? x[a] : 'foo';
`, Options: optTern(false)},
		{Code: `
const a = 'b';
declare let x: { a: string; b: string } | null;
x?.[a] != null ? x.a : 'foo';
`, Options: optTern(false)},
		{Code: `
declare let x: { a: string } | null;
declare let y: { a: string } | null;
x?.a ? y?.a : 'foo';
`, Options: optTern(false)},
		// Compound: `null !== null` etc. degenerate
		{Code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject !== undefined && null !== null ? nullOrObject : 42;
`, Options: optTern(false)},
		{Code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject !== undefined && null != null ? nullOrObject : 42;
`, Options: optTern(false)},
		{Code: `
declare const nullOrObject: null | { a: string };

const test =
  nullOrObject !== undefined && null != undefined ? nullOrObject : 42;
`, Options: optTern(false)},
		{Code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject === undefined || null === null ? 42 : nullOrObject;
`, Options: optTern(false)},
		{Code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject === undefined || null == null ? 42 : nullOrObject;
`, Options: optTern(false)},
		{Code: `
declare const nullOrObject: null | { a: string };

const test =
  nullOrObject === undefined || null == undefined ? 42 : nullOrObject;
`, Options: optTern(false)},
		// ignoreIfStatements: true
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
`, Options: optMap("ignoreIfStatements", true)},
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) foo = makeFoo();
}
`, Options: optMap("ignoreIfStatements", true)},
	}

	// Valid 5: ignoreConditionalTests — `||` inside test stays under default.
	// `x ||= 'foo'` must be parenthesized to land in the ternary test
	// position; `x || 'foo'` doesn't need parens.
	condTestsValid := mustConcat(
		nullishTypeValid(func(n, t, eq string) string {
			if eq == "=" {
				return fmt.Sprintf(`
declare let x: %s | %s;
(x ||= 'foo') ? null : null;
`, t, n)
			}
			return fmt.Sprintf(`
declare let x: %s | %s;
x || 'foo' ? null : null;
`, t, n)
		}),
		nullishTypeValid(func(n, t, eq string) string {
			return fmt.Sprintf(`
declare let x: %s | %s;
if (x ||%s 'foo') {
}
`, t, n, eq)
		}),
		nullishTypeValid(func(n, t, eq string) string {
			return fmt.Sprintf(`
declare let x: %s | %s;
do {} while (x ||%s 'foo');
`, t, n, eq)
		}),
		nullishTypeValid(func(n, t, eq string) string {
			return fmt.Sprintf(`
declare let x: %s | %s;
for (; x ||%s 'foo'; ) {}
`, t, n, eq)
		}),
		nullishTypeValid(func(n, t, eq string) string {
			return fmt.Sprintf(`
declare let x: %s | %s;
while (x ||%s 'foo') {}
`, t, n, eq)
		}),
		// Nested under `??`, `&&`, `!`, `,`
		[]rule_tester.ValidTestCase{
			{Code: `
let a: string | undefined;
let b: string | undefined;
if (!(a || b)) {}
`, Options: optMap("ignoreConditionalTests", true)},
			{Code: `
let a: string | undefined;
let b: string | undefined;
if (!!(a || b)) {}
`, Options: optMap("ignoreConditionalTests", true)},
			{Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
if (a ? a : b) {}
`, Options: optMap("ignoreConditionalTests", true)},
			{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
if (!a ? b : a) {}
`, Options: optMap("ignoreConditionalTests", true)},
			{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if ((a ? a : b) || c) {}
`, Options: optMap("ignoreConditionalTests", true)},
			{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if (c || (!a ? b : a)) {}
`, Options: optMap("ignoreConditionalTests", true)},
			{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if ((a || b) ?? c) {}
`, Options: optMap("ignoreConditionalTests", true)},
			{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if (a ?? (b || c)) {}
`, Options: optMap("ignoreConditionalTests", true)},
			{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if (a ? b || c : 'fail') {}
`, Options: optMap("ignoreConditionalTests", true)},
			{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if (a ? 'success' : b || c) {}
`, Options: optMap("ignoreConditionalTests", true)},
			{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if (((a = b), b || c)) {}
`, Options: optMap("ignoreConditionalTests", true)},
		},
	)

	// Valid 6: ignoreMixedLogicalExpressions
	mixedValid := nullishTypeValidWithOpts(optMap("ignoreMixedLogicalExpressions", true),
		func(n, t, eq string) string {
			return fmt.Sprintf(`
declare let a: %s | %s;
declare let b: %s | %s;
declare let c: %s | %s;
a ||%s (b && c);
`, t, n, t, n, t, n, eq)
		})

	mixedValid2 := []rule_tester.ValidTestCase{
		{Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || b || (c && d);
`, Options: optMap("ignoreMixedLogicalExpressions", true)},
		{Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || (b && c) || d;
`, Options: optMap("ignoreMixedLogicalExpressions", true)},
		{Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || (b && c && d);
`, Options: optMap("ignoreMixedLogicalExpressions", true)},
	}

	// Valid 7: ignorePrimitives matrix
	ignorePrim := []rule_tester.ValidTestCase{
		{Code: `
declare let x: string | null | undefined;
x || 'foo';
`, Options: optPrim(map[string]bool{"string": true})},
		{Code: `
declare let x: number | null | undefined;
x || 1;
`, Options: optPrim(map[string]bool{"number": true})},
		{Code: `
declare let x: bigint | null | undefined;
x || 1n;
`, Options: optPrim(map[string]bool{"bigint": true})},
		{Code: `
declare let x: boolean | null | undefined;
x || true;
`, Options: optPrim(map[string]bool{"boolean": true})},
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": true, "boolean": true, "number": false, "string": true})},
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": false, "boolean": true, "number": true, "string": true})},
		{Code: `
declare let x: 0 | 'foo' | undefined;
x || y;
`, Options: optPrim(map[string]bool{"number": true, "string": true})},
		{Code: `
declare let x: 0 | 'foo' | undefined;
x || y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare const a: any;
declare const b: any;
a ? a : b;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare const a: any;
declare const b: any;
a ? a : b;
`, Options: optPrim(map[string]bool{"number": true})},
		{Code: `
declare const a: unknown;
const b = a || 'bar';
`, Options: optPrim(map[string]bool{"bigint": true, "boolean": false, "number": false, "string": false})},
		// "never" types are ineligible.
		{Code: `
declare let x: never;
declare let y: number;
x ? x : y;
`, Options: optTern(false)},
		{Code: `
declare let x: never;
declare let y: number;
!x ? y : x;
`, Options: optTern(false)},
	}

	// Valid 8: ignoreBooleanCoercion
	booleanValid := []rule_tester.ValidTestCase{
		// `||` inside Boolean(...) — ignored.
		{Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
const x = Boolean(a || b);
`, Options: optMap("ignoreBooleanCoercion", true)},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(a || b || c);
`, Options: optMap("ignoreBooleanCoercion", true)},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(a || (b && c));
`, Options: optMap("ignoreBooleanCoercion", true)},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean((a || b) ?? c);
`, Options: optMap("ignoreBooleanCoercion", true)},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(a ?? (b || c));
`, Options: optMap("ignoreBooleanCoercion", true)},
		// Ternaries that aren't direct args of Boolean(...) — ignored.
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(a ? b || c : 'fail');
`, Options: optMap("ignoreBooleanCoercion", true)},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(a ? 'success' : b || c);
`, Options: optMap("ignoreBooleanCoercion", true)},
		// Comma sequence's last element is the Boolean arg — ignored.
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(((a = b), b || c));
`, Options: optMap("ignoreBooleanCoercion", true)},
	}

	// rslint-extra: locally shadowed `Boolean` — when the call target is not
	// the global Boolean, ignoreBooleanCoercion does NOT apply (a shadowed
	// Boolean is just a regular function call). The rule reports normally.
	// We keep this as an invalid test below.
	booleanShadow := []rule_tester.ValidTestCase{}

	// rslint-extra edges
	rslintExtraValid := []rule_tester.ValidTestCase{
		// `as`/`satisfies`/non-null assertion wrappers on test
		{Code: `
declare let x: string | null;
(x as any) !== undefined ? (x as any) : y;
`, Options: optTern(false)},
		// Same identifier under intersection of similar type
		{Code: `
type T = (string | null) & {};
declare let x: T;
x ?? 'foo';
`},
	}

	// ───────── ignorePrimitives — bulk matrix (mirrors upstream) ─────────
	ignorePrimBulk := []rule_tester.ValidTestCase{
		// per-primitive `||`
		{Code: `
declare let x: string | undefined;
x || y;
`, Options: optPrim(map[string]bool{"string": true})},
		{Code: `
declare let x: number | undefined;
x || y;
`, Options: optPrim(map[string]bool{"number": true})},
		{Code: `
declare let x: boolean | undefined;
x || y;
`, Options: optPrim(map[string]bool{"boolean": true})},
		{Code: `
declare let x: bigint | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": true})},
		// `ignorePrimitives: true` umbrella
		{Code: `
declare let x: string | undefined;
x || y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: number | undefined;
x || y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: boolean | undefined;
x || y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: bigint | undefined;
x || y;
`, Options: optMap("ignorePrimitives", true)},
		// Branded primitive intersections — upstream still respects
		// ignorePrimitives because intersection includes the primitive flag.
		{Code: `
declare let x: (string & { __brand?: any }) | undefined;
x || y;
`, Options: optPrim(map[string]bool{"string": true})},
		{Code: `
declare let x: (number & { __brand?: any }) | undefined;
x || y;
`, Options: optPrim(map[string]bool{"number": true})},
		{Code: `
declare let x: (boolean & { __brand?: any }) | undefined;
x || y;
`, Options: optPrim(map[string]bool{"boolean": true})},
		{Code: `
declare let x: (bigint & { __brand?: any }) | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": true})},
		{Code: `
declare let x: (string & { __brand?: any }) | undefined;
x || y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: (number & { __brand?: any }) | undefined;
x || y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: (boolean & { __brand?: any }) | undefined;
x || y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: (bigint & { __brand?: any }) | undefined;
x || y;
`, Options: optMap("ignorePrimitives", true)},
		// truthy ternary forms
		{Code: `
declare let x: string | undefined;
x ? x : y;
`, Options: optPrim(map[string]bool{"string": true})},
		{Code: `
declare let x: number | undefined;
x ? x : y;
`, Options: optPrim(map[string]bool{"number": true})},
		{Code: `
declare let x: boolean | undefined;
x ? x : y;
`, Options: optPrim(map[string]bool{"boolean": true})},
		{Code: `
declare let x: bigint | undefined;
x ? x : y;
`, Options: optPrim(map[string]bool{"bigint": true})},
		{Code: `
declare let x: string | undefined;
!x ? y : x;
`, Options: optPrim(map[string]bool{"string": true})},
		{Code: `
declare let x: number | undefined;
!x ? y : x;
`, Options: optPrim(map[string]bool{"number": true})},
		{Code: `
declare let x: boolean | undefined;
!x ? y : x;
`, Options: optPrim(map[string]bool{"boolean": true})},
		{Code: `
declare let x: bigint | undefined;
!x ? y : x;
`, Options: optPrim(map[string]bool{"bigint": true})},
		{Code: `
declare let x: string | undefined;
x ? x : y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: number | undefined;
x ? x : y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: boolean | undefined;
x ? x : y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: bigint | undefined;
x ? x : y;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: string | undefined;
!x ? y : x;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: number | undefined;
!x ? y : x;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: boolean | undefined;
!x ? y : x;
`, Options: optMap("ignorePrimitives", true)},
		{Code: `
declare let x: bigint | undefined;
!x ? y : x;
`, Options: optMap("ignorePrimitives", true)},
	}

	// ───────── real-world user scenarios ─────────
	realWorld := []rule_tester.ValidTestCase{
		// Already idiomatic `??`
		{Code: `
function getName(user?: { name?: string }) {
  return user?.name ?? 'anonymous';
}
`},
		{Code: `
type Config = { port?: number; host?: string };
declare const cfg: Config;
const port = cfg.port ?? 3000;
const host = cfg.host ?? 'localhost';
`},
		// Defaulting with non-nullable types — must NOT report.
		{Code: `
function paginate(limit: number, offset: number = 0) {
  return [limit, offset];
}
`},
		{Code: `
const role: 'admin' | 'user' | 'guest' = 'guest';
const permissions = role || 'guest';
`},
		// Logical assignment for caching / memo (already nullish form)
		{Code: `
const cache: Map<string, string | null> = new Map();
declare let key: string;
let cached: string | null | undefined;
cached = cache.get(key) ?? null;
`},
		// `??` with bigint default
		{Code: `
declare let bigCount: bigint | null;
const total = bigCount ?? 0n;
`},
		// Strictly nullable function arg + non-nullable receiver — `||` ok.
		{Code: `
function search(q: string) { return q.length; }
function caller(query: string) {
  return search(query || '');
}
`},
		// `Array.prototype.find` returns `T | undefined` — handled by `??` already.
		{Code: `
const items: string[] = [];
const first = items.find(x => x.length > 0) ?? 'none';
`},
		// JSX attribute (default for prop) using `??`
		{
			Code: `
type Props = { value?: string };
declare const props: Props;
declare function Foo(p: any): any;
const el = <Foo value={props.value ?? 'default'} />;
`,
			Tsx: true,
		},
		// Async/await — `await` result with `??`
		{Code: `
declare function fetchUser(): Promise<{ name: string } | null>;
async function main() {
  const u = (await fetchUser()) ?? { name: 'anonymous' };
  return u.name;
}
`},
		// Generator yielding union — non-nullable
		{Code: `
function* names(): Generator<string> {
  let names: string[] = ['a', 'b'];
  for (const n of names) yield n || 'default';
}
`},
		// Class + private fields, non-nullable.
		{Code: `
class Counter {
  #value: number = 0;
  inc() { this.#value = (this.#value || 0) + 1; }
}
`},
		// Tagged template with non-nullable tag.
		{Code: `
function tag(s: TemplateStringsArray, ...args: unknown[]): string { return s.join(''); }
const x = tag` + "`hello ${1 || 2}`" + `;
`},
	}

	return mustConcat(nonNullableOr, alreadyNullish, ternaryNotClean, ifNoFit, condTestsValid,
		mixedValid, mixedValid2, ignorePrim, booleanValid, booleanShadow,
		ignorePrimBulk, realWorld, rslintExtraValid)
}

func buildInvalid() []rule_tester.InvalidTestCase {
	// ── ||  /  ||= per (nullish, type, equals) — every base shape ─────────
	basicOr := nullishTypeInvalid(
		func(n, t, eq string) string {
			return fmt.Sprintf("\ndeclare let x: %s | %s;\nx ||%s 'foo';\n", t, n, eq)
		},
		func(n, t, eq string) string {
			return fmt.Sprintf("\ndeclare let x: %s | %s;\nx ??%s 'foo';\n", t, n, eq)
		},
		3, 5, 3, 3,
		"preferNullishOverOr", nil,
	)

	// ── Manual cases (single, position-asserted) ─────────────────────────
	manualOr := []rule_tester.InvalidTestCase{
		// Nullable receiver, `||` → `??`
		{Code: `
declare let x: object | null | undefined;
x ||= {};
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverOr", Line: 3, Column: 3, EndLine: 3, EndColumn: 6,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
x ??= {};
`},
			},
		}}},
	}

	// ── Ternary equivalences (clean nullish-equivalent test) ──────────────
	ternaryClean := []rule_tester.InvalidTestCase{
		ternFix(`x !== undefined && x !== null ? x : y;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`x !== null && x !== undefined ? x : y;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`x !== undefined && null !== x ? x : y;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`null !== x && x !== undefined ? x : y;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`undefined !== x && null !== x ? x : y;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`x !== null && undefined !== x ? x : y;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`undefined !== x && x !== null ? x : y;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`null !== x && undefined !== x ? x : y;`, `x ?? y;`, 1, 38, 1, 1),
		// === / null + undefined — alternate branch order.
		ternFix(`x === undefined || x === null ? y : x;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`x === null || x === undefined ? y : x;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`x === undefined || null === x ? y : x;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`null === x || x === undefined ? y : x;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`undefined === x || null === x ? y : x;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`x === null || undefined === x ? y : x;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`undefined === x || x === null ? y : x;`, `x ?? y;`, 1, 38, 1, 1),
		ternFix(`null === x || undefined === x ? y : x;`, `x ?? y;`, 1, 38, 1, 1),
		// `==` / `!=` mixed — collapses to !=/==.
		ternFix(`x != undefined && x !== null ? x : y;`, `x ?? y;`, 1, 37, 1, 1),
		ternFix(`x !== undefined && x != null ? x : y;`, `x ?? y;`, 1, 37, 1, 1),
		ternFix(`x != null && x != undefined ? x : y;`, `x ?? y;`, 1, 36, 1, 1),
		// Side-effect right branch needs parens.
		ternFix(`x !== undefined && x !== null ? x : (z = y);`, `x ?? (z = y);`, 1, 44, 1, 1),
		ternFix(`x !== null && undefined !== x ? x : (z = y);`, `x ?? (z = y);`, 1, 44, 1, 1),
		ternFix(`x === undefined || null === x ? (z = y) : x;`, `x ?? (z = y);`, 1, 44, 1, 1),
		ternFix(`null === x || x === undefined ? (z = y) : x;`, `x ?? (z = y);`, 1, 44, 1, 1),
		ternFix(`x === null || undefined === x ? (z = y) : x;`, `x ?? (z = y);`, 1, 44, 1, 1),
	}

	// Single-side undefined / null check.
	singleSide := []rule_tester.InvalidTestCase{
		{Code: `
declare let b: string | undefined;
b !== undefined ? b : 'a string';
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary", Line: 3, Column: 1, EndLine: 3, EndColumn: 33,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let b: string | undefined;
b ?? 'a string';
`},
			},
		}}},
		{Code: `
declare let b: string | undefined;
b === undefined ? 'a string' : b;
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary", Line: 3, Column: 1, EndLine: 3, EndColumn: 33,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let b: string | undefined;
b ?? 'a string';
`},
			},
		}}},
		{Code: `
declare let c: string | null;
c !== null ? c : 'a string';
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary", Line: 3, Column: 1, EndLine: 3, EndColumn: 28,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let c: string | null;
c ?? 'a string';
`},
			},
		}}},
		{Code: `
declare let c: string | null;
c === null ? 'a string' : c;
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary", Line: 3, Column: 1, EndLine: 3, EndColumn: 28,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let c: string | null;
c ?? 'a string';
`},
			},
		}}},
	}

	// Truthy ternary on nullable.
	truthyTern := []rule_tester.InvalidTestCase{
		{Code: `
declare let b: string | undefined;
b ? b : 'a string';
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary", Line: 3, Column: 1, EndLine: 3, EndColumn: 19,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let b: string | undefined;
b ?? 'a string';
`},
			},
		}}},
		{Code: `
declare let c: string | null;
!c ? 'a string' : c;
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary", Line: 3, Column: 1, EndLine: 3, EndColumn: 20,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let c: string | null;
c ?? 'a string';
`},
			},
		}}},
		{Code: `
declare let a: any;
a ? a : 'a string';
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary", Line: 3, Column: 1, EndLine: 3, EndColumn: 19,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let a: any;
a ?? 'a string';
`},
			},
		}}},
	}

	// Optional-chain similarity.
	optionalChain := []rule_tester.InvalidTestCase{
		{Code: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x.n?.a : y;
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary", Line: 3, Column: 1, EndLine: 3, EndColumn: 34,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
			},
		}}},
		{Code: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x?.n?.a : y;
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary", Line: 3, Column: 1, EndLine: 3, EndColumn: 22,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
			},
		}}},
		{Code: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x.n?.a : y;
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary", Line: 3, Column: 1, EndLine: 3, EndColumn: 35,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
			},
		}}},
	}

	// IfStatement → ??=
	ifBasic := []rule_tester.InvalidTestCase{
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitializeFoo1() {
  if (!foo) {
    foo = makeFoo();
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment", Line: 6, Column: 3, EndLine: 8, EndColumn: 4,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitializeFoo1() {
  foo ??= makeFoo();
}
`},
			},
		}}},
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitializeFoo2() {
  if (!foo) foo = makeFoo();
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment", Line: 6, Column: 3, EndLine: 6, EndColumn: 29,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitializeFoo2() {
  foo ??= makeFoo();
}
`},
			},
		}}},
		// `if (foo == null) foo = makeFoo()` — single-statement form.
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo = makeFoo();
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
			},
		}}},
		// `if (foo === undefined) foo = makeFoo()` on `T | undefined`.
		{Code: `
declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === undefined) {
    foo = makeFoo();
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
			},
		}}},
		// member-access LHS / RHS.
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo.a == null) {
    foo.a = makeFoo();
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
`},
			},
		}}},
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo?.a == null) {
    foo.a = makeFoo();
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
`},
			},
		}}},
		// Body already nullish-coalescing assignment — still emits.
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ??= makeFoo();
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
			},
		}}},
		// weirdParens.
		{Code: `
declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  if (((((foo.a)) == null))) {
    ((((((((foo).a))))) = makeString()));
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  ((foo).a) ??= makeString();
}
`},
			},
		}}},
		// `if (foo === undefined || foo === null) foo = makeFoo()` on `T |
		// null | undefined` — both checks → fixable.
		{Code: `
declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === undefined || foo === null) {
    foo = makeFoo();
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
			},
		}}},
	}

	// ignoreConditionalTests: false flips loop / if-test cases on.
	// Build the ternary-test case manually since the input shape differs by eq.
	condTernaryInvalid := func() []rule_tester.InvalidTestCase {
		out := []rule_tester.InvalidTestCase{}
		for _, n := range nullishTypes {
			for _, t := range primitiveTypes {
				// eq=''
				out = append(out, rule_tester.InvalidTestCase{
					Code: fmt.Sprintf(`
declare let x: %s | %s;
x || 'foo' ? null : null;
`, t, n),
					Options: optMap("ignoreConditionalTests", false),
					Errors: []rule_tester.InvalidTestCaseError{{
						MessageId: "preferNullishOverOr", Line: 3, EndLine: 3, Column: 3, EndColumn: 5,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNullish", Output: fmt.Sprintf(`
declare let x: %s | %s;
x ?? 'foo' ? null : null;
`, t, n)},
						},
					}},
				})
				// eq='='
				out = append(out, rule_tester.InvalidTestCase{
					Code: fmt.Sprintf(`
declare let x: %s | %s;
(x ||= 'foo') ? null : null;
`, t, n),
					Options: optMap("ignoreConditionalTests", false),
					Errors: []rule_tester.InvalidTestCaseError{{
						MessageId: "preferNullishOverOr", Line: 3, EndLine: 3, Column: 4, EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNullish", Output: fmt.Sprintf(`
declare let x: %s | %s;
(x ??= 'foo') ? null : null;
`, t, n)},
						},
					}},
				})
			}
		}
		return out
	}()

	condTestsInvalid := mustConcatInvalid(
		condTernaryInvalid,
		nullishTypeInvalid(
			func(n, t, eq string) string {
				return fmt.Sprintf(`
declare let x: %s | %s;
if (x ||%s 'foo') {
}
`, t, n, eq)
			},
			func(n, t, eq string) string {
				return fmt.Sprintf(`
declare let x: %s | %s;
if (x ??%s 'foo') {
}
`, t, n, eq)
			},
			7, 9, 3, 3,
			"preferNullishOverOr", optMap("ignoreConditionalTests", false),
		),
		nullishTypeInvalid(
			func(n, t, eq string) string {
				return fmt.Sprintf(`
declare let x: %s | %s;
do {} while (x ||%s 'foo');
`, t, n, eq)
			},
			func(n, t, eq string) string {
				return fmt.Sprintf(`
declare let x: %s | %s;
do {} while (x ??%s 'foo');
`, t, n, eq)
			},
			16, 18, 3, 3,
			"preferNullishOverOr", optMap("ignoreConditionalTests", false),
		),
		nullishTypeInvalid(
			func(n, t, eq string) string {
				return fmt.Sprintf(`
declare let x: %s | %s;
for (; x ||%s 'foo'; ) {}
`, t, n, eq)
			},
			func(n, t, eq string) string {
				return fmt.Sprintf(`
declare let x: %s | %s;
for (; x ??%s 'foo'; ) {}
`, t, n, eq)
			},
			10, 12, 3, 3,
			"preferNullishOverOr", optMap("ignoreConditionalTests", false),
		),
		nullishTypeInvalid(
			func(n, t, eq string) string {
				return fmt.Sprintf(`
declare let x: %s | %s;
while (x ||%s 'foo') {}
`, t, n, eq)
			},
			func(n, t, eq string) string {
				return fmt.Sprintf(`
declare let x: %s | %s;
while (x ??%s 'foo') {}
`, t, n, eq)
			},
			10, 12, 3, 3,
			"preferNullishOverOr", optMap("ignoreConditionalTests", false),
		),
	)

	// ignoreBooleanCoercion default false → reports inside Boolean(...)
	booleanInvalid := []rule_tester.InvalidTestCase{
		{Code: `
declare const a: string | true | undefined;
declare const b: string | boolean | undefined;
const x = Boolean(a || b);
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverOr", Line: 4, Column: 21, EndLine: 4, EndColumn: 23,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare const a: string | true | undefined;
declare const b: string | boolean | undefined;
const x = Boolean(a ?? b);
`},
			},
		}}},
		// ignoreBooleanCoercion: true but ternary is direct Boolean arg.
		{Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
const x = Boolean(a ? a : b);
`, Options: optMap("ignoreBooleanCoercion", true), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;
const x = Boolean(a ?? b);
`},
			},
		}}},
		// Same as above but with extra parens around the ternary —
		// `Boolean((a ? a : b))`. ESTree paren-transparency makes this still
		// match the carve-out (ternary IS the direct CallExpression argument).
		// tsgo keeps an explicit ParenthesizedExpression so we must skip
		// paren layers when matching `node.parent === CallExpression`.
		{Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
const x = Boolean((a ? a : b));
`, Options: optMap("ignoreBooleanCoercion", true), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;
const x = Boolean((a ?? b));
`},
			},
		}}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
const test = Boolean(!a ? b : a);
`, Options: optMap("ignoreBooleanCoercion", true), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
const test = Boolean(a ?? b);
`},
			},
		}}},
	}

	// ignoreIfStatements: false flips on
	ifFlippedOn := []rule_tester.InvalidTestCase{
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
`, Options: optMap("ignoreIfStatements", false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
			},
		}}},
	}

	// ignorePrimitives — literal-typed unions.
	primInvalid := []rule_tester.InvalidTestCase{
		{Code: `
declare let x: 0 | 1 | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": true, "boolean": true, "number": false, "string": true}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | 1 | undefined;
x ?? y;
`},
				},
			}}},
		{Code: `
declare let x: 1 | 2 | 3 | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": true, "boolean": true, "number": false, "string": true}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
`},
				},
			}}},
		{Code: `
declare let x: 'a' | 'b' | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": true, "boolean": true, "number": true, "string": false}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
`},
				},
			}}},
		{Code: `
declare let x: 0n | 1n | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": false, "boolean": true, "number": true, "string": true}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0n | 1n | undefined;
x ?? y;
`},
				},
			}}},
		{Code: `
declare let x: true | false | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": true, "boolean": false, "number": true, "string": true}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: true | false | undefined;
x ?? y;
`},
				},
			}}},
		// truthy ternary on literal
		{Code: `
declare let x: 'a' | 'b' | undefined;
x ? x : y;
`, Options: optPrim(map[string]bool{"bigint": true, "boolean": true, "number": true, "string": false}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverTernary",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
`},
				},
			}}},
		// default for missing option (no `string`/`number`/`bigint`/`boolean`)
		{Code: `
declare let x: string | undefined;
x || y;
`, Options: optMap("ignorePrimitives", map[string]interface{}{"bigint": true, "boolean": true, "number": true}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
x ?? y;
`},
				},
			}}},
		{Code: `
declare let x: number | undefined;
x || y;
`, Options: optMap("ignorePrimitives", map[string]interface{}{"bigint": true, "boolean": true, "string": true}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
x ?? y;
`},
				},
			}}},
		{Code: `
declare let x: boolean | undefined;
x || y;
`, Options: optMap("ignorePrimitives", map[string]interface{}{"bigint": true, "number": true, "string": true}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
x ?? y;
`},
				},
			}}},
		// Literal-typed truthy ternary on bigint
		{Code: `
declare let x: 1n | undefined;
x ? x : y;
`, Options: optPrim(map[string]bool{"bigint": false, "boolean": true, "number": true, "string": true}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverTernary",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1n | undefined;
x ?? y;
`},
				},
			}}},
	}

	// noStrictNullCheck — uses Go-side TSConfig override.
	//
	// Upstream reports `loc:{start:{line:0,column:0}, end:{line:0,column:0}}`
	// which ESLint surfaces as `column:1, line:0` (column gets +1, line is
	// passed through unchanged). rslint's framework normalizes both to
	// 1-based, producing line:1 column:1 — the underlying diagnostic
	// position (file start, pos 0..0) is identical; only the rendering
	// convention differs.
	noStrict := []rule_tester.InvalidTestCase{
		{
			Code: `
declare let x: string[] | null;
if (x) {
}
`,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStrictNullCheck", Line: 1, Column: 1, EndLine: 1, EndColumn: 1},
			},
		},
	}

	// Mixed `||` chains — emit one report per `||`.
	mixedOr := []rule_tester.InvalidTestCase{
		// `(a || b) || c` — chained `||`. Mirrors upstream byte-for-byte:
		// the inner suggestion produces `((a ?? b)) || c;` (extra parens are
		// redundant but syntactically valid; this is upstream's output).
		{Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string;

(a || b) || c;
`, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string;

((a ?? b)) || c;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string;

(a || b) ?? c;
`},
				},
			},
		}},
		// `a || b && c` — `&&` and `??` cannot be mixed without parens, but
		// upstream's autofix still produces `a ?? b && c` (a syntax error)
		// since the parent isn't `||`/`||=` so no parens are added. Mirror
		// upstream byte-for-byte; this is documented upstream behavior.
		{Code: `
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;

a || b && c;
`, Options: optMap("ignoreMixedLogicalExpressions", false), Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;

a ?? b && c;
`},
				},
			},
		}},
		// Aligned with upstream ordering (inner `||` first, then outer):
		// using exit-order listeners makes leaf-first emit order.
		{Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && b) || c || d;
`, Options: optMap("ignoreMixedLogicalExpressions", false), Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr", Line: 6, Column: 10, EndLine: 6, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && (b) ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr", Line: 6, Column: 15, EndLine: 6, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && b) || c ?? d;
`},
				},
			},
		}},
	}

	// If-with-non-= assignment operator (covers `||=`, `??=`).
	ifWithCompound := []rule_tester.InvalidTestCase{
		// `if (foo == null) foo ||= makeFoo()` — emits BOTH
		// preferNullishOverOr (inner ||=, leaf-first) AND
		// preferNullishOverAssignment (the if). Order: leaf → root via
		// exit listeners; matches upstream's depth-first emit order.
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ||= makeFoo();
  const bar = 42;
  return bar;
}
`, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ??= makeFoo();
  const bar = 42;
  return bar;
}
`},
				},
			},
			{
				MessageId: "preferNullishOverAssignment",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
`},
				},
			},
		}},
		// Body already `??=` — preferNullishOverAssignment still fires.
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ??= makeFoo();
  const bar = 42;
  return bar;
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
`},
			},
		}}},
	}

	// Issue #1290 — `a || b || c` where `b`/`c` are non-nullable.
	issue1290 := nullishTypeInvalid(
		func(n, t, eq string) string {
			_ = eq
			return fmt.Sprintf(`
declare let a: %s | null | undefined;
declare let b: %s;
declare let c: %s;
a || b || c;
`, t, t, t)
		},
		func(n, t, eq string) string {
			_ = eq
			return fmt.Sprintf(`
declare let a: %s | null | undefined;
declare let b: %s;
declare let c: %s;
(a ?? b) || c;
`, t, t, t)
		},
		3, 5, 5, 5,
		"preferNullishOverOr", nil,
	)
	// Filter: only the variant where eq=='' is meaningful (a || b || c, not a ||= b ||= c).
	issue1290Filtered := make([]rule_tester.InvalidTestCase, 0, len(issue1290)/2)
	for i, c := range issue1290 {
		// Keep only those whose code uses bare `||`. Indexing: cb is invoked
		// with eq='' for even loop positions (no — actually our generator
		// nests by nullish×type×eq, so eq cycles innermost). Filter by
		// substring presence instead.
		_ = i
		if strings.Contains(c.Code, "a || b || c;") {
			issue1290Filtered = append(issue1290Filtered, c)
		}
	}
	// Also collapse duplicates that vary only by `nullish` (since the code
	// template doesn't use `nullish`, all 3 are identical) — keep distinct.
	seen := map[string]bool{}
	dedup := make([]rule_tester.InvalidTestCase, 0, len(issue1290Filtered))
	for _, c := range issue1290Filtered {
		if seen[c.Code] {
			continue
		}
		seen[c.Code] = true
		dedup = append(dedup, c)
	}
	issue1290Final := dedup

	// Same-primitive unions (literal types) — invalid because nullable
	// suite of literal-typed primitives.
	literalUnions := []rule_tester.InvalidTestCase{
		{Code: `
declare let x: 'a' | 'b' | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": true, "boolean": true, "number": true, "string": false}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
`},
				},
			}}},
		{Code: `
declare let x: 0 | 1 | undefined;
x || y;
`, Options: optPrim(map[string]bool{"bigint": true, "boolean": true, "number": false, "string": true}),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | 1 | undefined;
x ?? y;
`},
				},
			}}},
	}

	// rslint-extra: tsgo-specific edge cases.
	rslintExtraInvalid := []rule_tester.InvalidTestCase{
		// Class member access via `this.foo`.
		{Code: `
class C {
  declare foo: string | null | undefined;
  use() {
    return this.foo || 'foo';
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverOr",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
class C {
  declare foo: string | null | undefined;
  use() {
    return this.foo ?? 'foo';
  }
}
`},
			},
		}}},
		// Nested ternary as right branch (precedence guard).
		{Code: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ? a : b ? 1 : 2;
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
`},
			},
		}}},
		// Side-effect in nullish branch with already-parenthesized RHS.
		{Code: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ? a : (b ? 1 : 2);
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
`},
			},
		}}},
		// `c !== null ? c : c ? 1 : 2` — operator-precedence guard.
		{Code: `
declare const c: string | null;
c !== null ? c : c ? 1 : 2;
`, Options: optTern(false), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare const c: string | null;
c ?? (c ? 1 : 2);
`},
			},
		}}},
		// Comments preserved in `if`-block body — block form (single line comment).
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    // comment
    foo = makeFoo();
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  // comment
foo ??= makeFoo();
}
`},
			},
		}}},
		// Comments preserved in `if`-block body — single-statement form, two block comments.
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) /* c1 */ /* c2 */ foo = makeFoo();
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

/* c1 */ /* c2 */ foo ??= makeFoo();
`},
			},
		}}},
		// Heavy comment mix — line + block + multi-line block + JSDoc, before
		// AND after the assignment, in a top-level `if` (no surrounding fn).
		// This is upstream's most aggressive comments suite case verbatim.
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) {
  // comment before 1
  /* comment before 2 */
  /* comment before 3
    which is multiline
  */
  /**
   * comment before 4
   * which is also multiline
   */
  foo = makeFoo(); // comment inline
  // comment after 1
  /* comment after 2 */
  /* comment after 3
    which is multiline
  */
  /**
   * comment after 4
   * which is also multiline
   */
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

// comment before 1
/* comment before 2 */
/* comment before 3
    which is multiline
  */
/**
   * comment before 4
   * which is also multiline
   */
foo ??= makeFoo(); // comment inline
// comment after 1
/* comment after 2 */
/* comment after 3
    which is multiline
  */
/**
   * comment after 4
   * which is also multiline
   */
`},
			},
		}}},
		// Single-statement form with leading inline `/* c */` AND trailing
		// `// inline` on the same line.
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) /* comment before 1 */ /* comment before 2 */ foo = makeFoo(); // comment inline
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

/* comment before 1 */ /* comment before 2 */ foo ??= makeFoo(); // comment inline
`},
			},
		}}},
		// Multiple consecutive line comments before the assignment in block.
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    // a
    // b
    // c
    foo = makeFoo();
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  // a
// b
// c
foo ??= makeFoo();
}
`},
			},
		}}},
		// Trailing comment after the (only) statement, before the closing `}`.
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
    // trailing
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo ??= makeFoo(); // trailing
}
`},
			},
		}}},
		// Both leading + trailing comments around the only statement.
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    /* lead */
    foo = makeFoo();
    /* tail */
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  /* lead */
foo ??= makeFoo(); /* tail */
}
`},
			},
		}}},
		// Single-statement form: a leading line comment between `if (...)` and
		// `stmt`. NOTE: upstream's separator is `' '` for non-block form,
		// which produces `// pre foo ??= makeFoo();` — the line comment
		// swallows the whole replacement on the same line. We mirror that
		// behavior byte-for-byte (verified against upstream ESLint).
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) // pre
foo = makeFoo();
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

// pre foo ??= makeFoo();
`},
			},
		}}},
		// Single-statement form: a multi-line block comment before stmt.
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) /* a
b */ foo = makeFoo();
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

/* a
b */ foo ??= makeFoo();
`},
			},
		}}},
		// Block body with NO comments — must not insert spurious whitespace.
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;
function f() {
  if (foo == null) {
    foo = makeFoo();
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;
function f() {
  foo ??= makeFoo();
}
`},
			},
		}}},
		// Single-statement form, NO comments.
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;
if (foo == null) foo = makeFoo();
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;
foo ??= makeFoo();
`},
			},
		}}},
		// Block body with multi-line block comment AFTER the statement.
		{Code: `
declare let foo: string | null;
declare function makeFoo(): string;
function f() {
  if (foo == null) {
    foo = makeFoo();
    /* multi
       line
    */
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverAssignment",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;
function f() {
  foo ??= makeFoo(); /* multi
       line
    */
}
`},
			},
		}}},
		// Locally shadowed Boolean — `||` is reported normally.
		{Code: `
function Boolean(x: any): boolean { return !!x; }
let a: string | boolean | undefined;
let b: string | boolean | undefined;
const x = Boolean(a || b);
`, Options: optMap("ignoreBooleanCoercion", true), Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverOr",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: `
function Boolean(x: any): boolean { return !!x; }
let a: string | boolean | undefined;
let b: string | boolean | undefined;
const x = Boolean(a ?? b);
`},
			},
		}}},
	}

	// ────────────────────── falsy literal types ──────────────────────────
	falsyLiterals := []rule_tester.InvalidTestCase{
		invalidWith(`
declare let x: '' | undefined;
x || y;
`, `
declare let x: '' | undefined;
x ?? y;
`, "preferNullishOverOr",
			optPrim(map[string]bool{"bigint": true, "boolean": true, "number": true, "string": false})),
		invalidWith(`
declare let x: `+"`"+`hello${'string'}`+"`"+` | undefined;
x || y;
`, `
declare let x: `+"`"+`hello${'string'}`+"`"+` | undefined;
x ?? y;
`, "preferNullishOverOr",
			optPrim(map[string]bool{"bigint": true, "boolean": true, "number": true, "string": false})),
		invalidWith(`
declare let x: 0 | undefined;
x || y;
`, `
declare let x: 0 | undefined;
x ?? y;
`, "preferNullishOverOr",
			optPrim(map[string]bool{"bigint": true, "boolean": true, "number": false, "string": true})),
		invalidWith(`
declare let x: 0n | undefined;
x || y;
`, `
declare let x: 0n | undefined;
x ?? y;
`, "preferNullishOverOr",
			optPrim(map[string]bool{"bigint": false, "boolean": true, "number": true, "string": true})),
		invalidWith(`
declare let x: false | undefined;
x || y;
`, `
declare let x: false | undefined;
x ?? y;
`, "preferNullishOverOr",
			optPrim(map[string]bool{"bigint": true, "boolean": false, "number": true, "string": true})),
		// truthy literal types in ternary form
		invalidWith(`
declare let x: '' | undefined;
x ? x : y;
`, `
declare let x: '' | undefined;
x ?? y;
`, "preferNullishOverTernary",
			optPrim(map[string]bool{"bigint": true, "boolean": true, "number": true, "string": false})),
		invalidWith(`
declare let x: 0 | undefined;
!x ? y : x;
`, `
declare let x: 0 | undefined;
x ?? y;
`, "preferNullishOverTernary",
			optPrim(map[string]bool{"bigint": true, "boolean": true, "number": false, "string": true})),
		invalidWith(`
declare let x: 0n | undefined;
x ? x : y;
`, `
declare let x: 0n | undefined;
x ?? y;
`, "preferNullishOverTernary",
			optPrim(map[string]bool{"bigint": false, "boolean": true, "number": true, "string": true})),
		invalidWith(`
declare let x: false | undefined;
x ? x : y;
`, `
declare let x: false | undefined;
x ?? y;
`, "preferNullishOverTernary",
			optPrim(map[string]bool{"bigint": true, "boolean": false, "number": true, "string": true})),
	}

	// ─────────────────────── mixed unions ────────────────────────────────
	mixedUnions := []rule_tester.InvalidTestCase{
		invalidWith(`
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
`, `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
`, "preferNullishOverOr",
			optPrim(map[string]bool{"bigint": false, "boolean": true, "number": false, "string": true})),
		invalidWith(`
declare let x: true | false | null | undefined;
x || y;
`, `
declare let x: true | false | null | undefined;
x ?? y;
`, "preferNullishOverOr",
			optPrim(map[string]bool{"bigint": true, "boolean": false, "number": true, "string": true})),
		invalidWith(`
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
`, `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
`, "preferNullishOverTernary",
			optPrim(map[string]bool{"bigint": false, "boolean": true, "number": false, "string": true})),
		invalidWith(`
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
`, `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
`, "preferNullishOverTernary",
			optPrim(map[string]bool{"bigint": false, "boolean": true, "number": false, "string": true})),
		invalidWith(`
declare let x: true | false | null | undefined;
!x ? y : x;
`, `
declare let x: true | false | null | undefined;
x ?? y;
`, "preferNullishOverTernary",
			optPrim(map[string]bool{"bigint": true, "boolean": false, "number": true, "string": true})),
	}

	// Null/undefined-only types and bare literals.
	nullishOnly := []rule_tester.InvalidTestCase{
		invalidWith(`
declare let x: null;
x || y;
`, `
declare let x: null;
x ?? y;
`, "preferNullishOverOr", nil),
		invalidWith(`
const x = undefined;
x || y;
`, `
const x = undefined;
x ?? y;
`, "preferNullishOverOr", nil),
		invalidWith(`
null || y;
`, `
null ?? y;
`, "preferNullishOverOr", nil),
		invalidWith(`
undefined || y;
`, `
undefined ?? y;
`, "preferNullishOverOr", nil),
		// enum union with undefined
		invalidWith(`
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x || y;
`, `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x ?? y;
`, "preferNullishOverOr", nil),
	}

	// Functions inside conditional tests — `() => x || 'foo'` is the test
	// expression; the inner `||` is NOT in conditional-test position
	// (function-body), so the rule reports.
	fnInsideCond := func() []rule_tester.InvalidTestCase {
		out := []rule_tester.InvalidTestCase{}
		for _, n := range nullishTypes {
			for _, t := range primitiveTypes {
				out = append(out, rule_tester.InvalidTestCase{
					Code: fmt.Sprintf(`
declare let x: %s | %s;
if (() => x || 'foo') {
}
`, t, n),
					Errors: []rule_tester.InvalidTestCaseError{{
						MessageId: "preferNullishOverOr", Line: 3, EndLine: 3, Column: 13, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNullish", Output: fmt.Sprintf(`
declare let x: %s | %s;
if (() => x ?? 'foo') {
}
`, t, n)},
						},
					}},
				})
				out = append(out, rule_tester.InvalidTestCase{
					Code: fmt.Sprintf(`
declare let x: %s | %s;
if (() => (x ||= 'foo')) {
}
`, t, n),
					Errors: []rule_tester.InvalidTestCaseError{{
						MessageId: "preferNullishOverOr", Line: 3, EndLine: 3, Column: 14, EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNullish", Output: fmt.Sprintf(`
declare let x: %s | %s;
if (() => (x ??= 'foo')) {
}
`, t, n)},
						},
					}},
				})
			}
		}
		return out
	}()

	// Extra tsgo edge cases:
	tsgoEdges := []rule_tester.InvalidTestCase{
		// Computed property access via element-access expressions —
		// member-access-like.
		invalidWith(`
declare let x: { foo: string } | null;
x ? x : { foo: '' };
`, `
declare let x: { foo: string } | null;
x ?? { foo: '' };
`, "preferNullishOverTernary", optTern(false)),
		// Class with `super.foo` access.
		invalidWith(`
class Base {
  declare foo: string | null;
}
class Derived extends Base {
  use() {
    return super.foo || 'foo';
  }
}
`, `
class Base {
  declare foo: string | null;
}
class Derived extends Base {
  use() {
    return super.foo ?? 'foo';
  }
}
`, "preferNullishOverOr", nil),
		// (Note) Generic type parameter constrained to a nullable union is NOT
		// reported — matches upstream's behavior because `getTypeAtLocation`
		// returns the type-parameter type whose flags don't include Null /
		// Undefined; ts-api-utils' `isNullableType` only checks union members.
		// Member of `this` inside an arrow class field.
		invalidWith(`
class C {
  declare foo: string | null;
  get value() {
    return this.foo || 'foo';
  }
}
`, `
class C {
  declare foo: string | null;
  get value() {
    return this.foo ?? 'foo';
  }
}
`, "preferNullishOverOr", nil),
		// Element access via numeric index.
		invalidWith(`
declare const arr: (string | null)[];
arr[0] || 'foo';
`, `
declare const arr: (string | null)[];
arr[0] ?? 'foo';
`, "preferNullishOverOr", nil),
		// Optional element access.
		invalidWith(`
declare const arr: ((string | null)[] | null);
arr?.[0] || 'foo';
`, `
declare const arr: ((string | null)[] | null);
arr?.[0] ?? 'foo';
`, "preferNullishOverOr", nil),
		// Returns of function call with nullable result.
		invalidWith(`
declare function getFoo(): string | null;
getFoo() || 'foo';
`, `
declare function getFoo(): string | null;
getFoo() ?? 'foo';
`, "preferNullishOverOr", nil),
		// Negative — `??` already in a chain doesn't cascade.
		invalidWith(`
declare let a: string | null;
declare let b: string | null;
declare let c: string;
a ?? b || c;
`, `
declare let a: string | null;
declare let b: string | null;
declare let c: string;
a ?? b ?? c;
`, "preferNullishOverOr", nil),
		// `as` and `satisfies` wrappers around the LHS — type-checking
		// still picks up the nullable wrapper underneath.
		invalidWith(`
declare let x: string | null;
(x as string | null) || 'foo';
`, `
declare let x: string | null;
(x as string | null) ?? 'foo';
`, "preferNullishOverOr", nil),
	}

	// ───────── real-world user scenarios (invalid) ─────────
	realWorldInvalid := []rule_tester.InvalidTestCase{
		// Express handler-style — `req.query` returns `string | undefined`.
		invalidWith(`
type Req = { query: { search?: string } };
declare const req: Req;
const term = req.query.search || 'default';
`, `
type Req = { query: { search?: string } };
declare const req: Req;
const term = req.query.search ?? 'default';
`, "preferNullishOverOr", nil),
		// Map.get
		invalidWith(`
const cache = new Map<string, string>();
const v = cache.get('k') || 'default';
`, `
const cache = new Map<string, string>();
const v = cache.get('k') ?? 'default';
`, "preferNullishOverOr", nil),
		// React-like prop fallback (parser handles tsx syntax)
		{
			Code: `
type Props = { name?: string };
declare const props: Props;
declare function Foo(p: any): any;
const el = <Foo name={props.name || 'default'} />;
`,
			Tsx:     true,
			Options: optMap("ignoreConditionalTests", false),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverOr",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
type Props = { name?: string };
declare const props: Props;
declare function Foo(p: any): any;
const el = <Foo name={props.name ?? 'default'} />;
`},
				},
			}},
		},
		// Async/await — `await result || fallback`
		invalidWith(`
declare function load(): Promise<string | null>;
async function main() {
  const value = (await load()) || 'default';
  return value;
}
`, `
declare function load(): Promise<string | null>;
async function main() {
  const value = (await load()) ?? 'default';
  return value;
}
`, "preferNullishOverOr", nil),
		// Destructuring with default + nullable union — when LHS is nullable.
		invalidWith(`
declare const cfg: { host?: string };
const host = cfg.host || 'localhost';
`, `
declare const cfg: { host?: string };
const host = cfg.host ?? 'localhost';
`, "preferNullishOverOr", nil),
		// Class private field with nullable storage — `||` form is still
		// reported (the `||` listener doesn't go through `isNodeEqual`).
		invalidWith(`
class Cache {
  #value: string | null = null;
  get(): string {
    return this.#value || 'fresh';
  }
}
`, `
class Cache {
  #value: string | null = null;
  get(): string {
    return this.#value ?? 'fresh';
  }
}
`, "preferNullishOverOr", nil),
		// Class field with `?: string`
		invalidWith(`
class User {
  declare name?: string;
  display(): string {
    return this.name || 'anonymous';
  }
}
`, `
class User {
  declare name?: string;
  display(): string {
    return this.name ?? 'anonymous';
  }
}
`, "preferNullishOverOr", nil),
		// Tagged template result fallback.
		invalidWith(`
declare function tag(s: TemplateStringsArray): string | null;
const x = tag` + "`hello`" + ` || 'default';
`, `
declare function tag(s: TemplateStringsArray): string | null;
const x = tag` + "`hello`" + ` ?? 'default';
`, "preferNullishOverOr", nil),
		// Function-call result that may be undefined (Array.find).
		invalidWith(`
const items: string[] = [];
const v = items.find(x => x.length > 0) || 'none';
`, `
const items: string[] = [];
const v = items.find(x => x.length > 0) ?? 'none';
`, "preferNullishOverOr", nil),
		// Generator-yielded value with nullable.
		invalidWith(`
declare function* gen(): Generator<string | null>;
declare const it: Iterator<string | null>;
const r = it.next().value || 'fallback';
`, `
declare function* gen(): Generator<string | null>;
declare const it: Iterator<string | null>;
const r = it.next().value ?? 'fallback';
`, "preferNullishOverOr", nil),
		// Index signature returning `T | undefined`
		invalidWith(`
type Dict = { [k: string]: string | undefined };
declare const d: Dict;
const v = d['key'] || 'fallback';
`, `
type Dict = { [k: string]: string | undefined };
declare const d: Dict;
const v = d['key'] ?? 'fallback';
`, "preferNullishOverOr", nil),
		// Tuple element typed `T | undefined`.
		invalidWith(`
declare const tup: [string | null, number | null];
const a = tup[0] || 'foo';
`, `
declare const tup: [string | null, number | null];
const a = tup[0] ?? 'foo';
`, "preferNullishOverOr", nil),
		// `env.FOO || default` — common bundler env-var fallback shape.
		invalidWith(`
declare const env: { FOO?: string };
const x = env.FOO || 'fallback';
`, `
declare const env: { FOO?: string };
const x = env.FOO ?? 'fallback';
`, "preferNullishOverOr", nil),
		// Nested object access with optional chain wrapped by `(...)`
		invalidWith(`
declare const config: { db?: { host?: string } };
const host = (config.db?.host) || 'localhost';
`, `
declare const config: { db?: { host?: string } };
const host = (config.db?.host) ?? 'localhost';
`, "preferNullishOverOr", nil),
		// Spread & rest argument — parens around `||` chain.
		invalidWith(`
declare function f(...args: string[]): void;
declare let opt: string | null;
f(opt || 'x');
`, `
declare function f(...args: string[]): void;
declare let opt: string | null;
f(opt ?? 'x');
`, "preferNullishOverOr", nil),
		// Array literal element fallback.
		invalidWith(`
declare let a: string | null;
const arr = [a || 'foo'];
`, `
declare let a: string | null;
const arr = [a ?? 'foo'];
`, "preferNullishOverOr", nil),
		// `.then` callback parameter fallback.
		invalidWith(`
declare function load(): Promise<string | null>;
load().then(v => v || 'default');
`, `
declare function load(): Promise<string | null>;
load().then(v => v ?? 'default');
`, "preferNullishOverOr", nil),
		// Default return in arrow with implicit body.
		invalidWith(`
declare let cache: Map<string, string | null>;
const get = (k: string) => cache.get(k) || 'fresh';
`, `
declare let cache: Map<string, string | null>;
const get = (k: string) => cache.get(k) ?? 'fresh';
`, "preferNullishOverOr", nil),
		// Computed access with string-literal index.
		invalidWith(`
declare const o: { [k: string]: string | undefined };
const v = o['k'] || 'fallback';
`, `
declare const o: { [k: string]: string | undefined };
const v = o['k'] ?? 'fallback';
`, "preferNullishOverOr", nil),
		// Branded primitive intersection — should still report when option
		// doesn't ignore that primitive.
		invalidWith(`
declare let x: (string & { __brand?: any }) | undefined;
x || y;
`, `
declare let x: (string & { __brand?: any }) | undefined;
x ?? y;
`, "preferNullishOverOr", nil),
	}

	// ───────── upstream optional-chain similarity batch ─────────
	// Covers `x.n?.a == null/undefined`, `x.n?.a != null/undefined` with
	// each branch reading a different chain shape (similar member-access).
	optionalChainBatch := []rule_tester.InvalidTestCase{
		invalidWith(`
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x.n.a : y;
`, `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x?.n?.a : y;
`, `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x.n.a : y;
`, `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x?.n?.a : y;
`, `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x.n?.a != null ? x.n.a : y;
`, `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x.n?.a != null ? x?.n.a : y;
`, `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x.n?.a != null ? x?.n?.a : y;
`, `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string | null } };
x.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y;
`, `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string | null } };
x.n?.a !== undefined && x.n.a !== null ? x.n.a : y;
`, `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x?.n?.a ? x?.n?.a : y;
`, `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x?.n?.a ? x.n?.a : y;
`, `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x?.n?.a ? x.n.a : y;
`, `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x?.n?.a : y;
`, `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x.n.a : y;
`, `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x?.n?.a : y;
`, `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
		invalidWith(`
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x.n.a : y;
`, `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`, "preferNullishOverTernary", optTern(false)),
	}

	// ───────── extra real-world variants ─────────
	moreRealWorld := []rule_tester.InvalidTestCase{
		// Tuple element fallback with two members.
		invalidWith(`
declare const tup: [string | null, number];
const [a, b] = tup;
const display = a || 'unknown';
`, `
declare const tup: [string | null, number];
const [a, b] = tup;
const display = a ?? 'unknown';
`, "preferNullishOverOr", nil),
		// JSX with conditional default — TSX path.
		{
			Code: `
type Props = { count?: number };
declare const props: Props;
declare function Counter(p: any): any;
const el = <Counter count={props.count !== undefined ? props.count : 0} />;
`,
			Tsx:     true,
			Options: optTern(false),
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNullishOverTernary",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
type Props = { count?: number };
declare const props: Props;
declare function Counter(p: any): any;
const el = <Counter count={props.count ?? 0} />;
`},
				},
			}},
		},
		// Nested JSON-like struct.
		invalidWith(`
type Cfg = { server?: { port?: number } };
declare const cfg: Cfg;
const port = cfg.server?.port || 3000;
`, `
type Cfg = { server?: { port?: number } };
declare const cfg: Cfg;
const port = cfg.server?.port ?? 3000;
`, "preferNullishOverOr", nil),
		// Generic function returning nullable.
		invalidWith(`
function getOr<T>(v: T | null, fallback: T): T {
  return v !== null ? v : fallback;
}
`, `
function getOr<T>(v: T | null, fallback: T): T {
  return v ?? fallback;
}
`, "preferNullishOverTernary", optTern(false)),
		// (Note) `lookup('id') ? lookup('id') : 'default'` is NOT reported —
		// CallExpression isn't `member-access-like` per upstream, so the
		// rule treats both branches as opaque calls and doesn't pair them
		// up. Matches upstream behavior.
		// String method result fallback (non-similar branches).
		invalidWith(`
declare let token: string | null;
const out = token != null ? token : '';
`, `
declare let token: string | null;
const out = token ?? '';
`, "preferNullishOverTernary", optTern(false)),
		// (Note) Class private field `this.#v` in a ternary is NOT reported
		// — matches upstream: `isNodeEqual` doesn't handle PrivateIdentifier
		// so the similarity check across branches fails. Real-world impact:
		// upstream rule misses this (arguably a bug in upstream); we mirror.
		// Array.prototype.at returns `T | undefined`.
		invalidWith(`
declare const arr: string[];
const v = arr.at(0) || 'none';
`, `
declare const arr: string[];
const v = arr.at(0) ?? 'none';
`, "preferNullishOverOr", nil),
		// Ternary inside a function call argument.
		invalidWith(`
declare function show(v: string): void;
declare let v: string | null;
show(v !== null ? v : 'fallback');
`, `
declare function show(v: string): void;
declare let v: string | null;
show(v ?? 'fallback');
`, "preferNullishOverTernary", optTern(false)),
		// `if`-statement with member access on this.
		invalidWith(`
class Thing {
  declare value: string | null;
  ensure() {
    if (this.value == null) {
      this.value = 'init';
    }
  }
}
`, `
class Thing {
  declare value: string | null;
  ensure() {
    this.value ??= 'init';
  }
}
`, "preferNullishOverAssignment", nil),
		// Ternary used as default argument.
		invalidWith(`
declare let v: string | null;
function f(arg: string = v !== null ? v : 'x') { return arg; }
`, `
declare let v: string | null;
function f(arg: string = v ?? 'x') { return arg; }
`, "preferNullishOverTernary", optTern(false)),
		// Computed key access with `null` member.
		invalidWith(`
declare const m: Map<string, string | null>;
const v = m.get('k') || 'fallback';
`, `
declare const m: Map<string, string | null>;
const v = m.get('k') ?? 'fallback';
`, "preferNullishOverOr", nil),
	}

	// Additional rare/extreme edges (non-deterministic positions skipped).
	extremeEdges := []rule_tester.InvalidTestCase{
		// Deeply nested optional chain similarity.
		invalidWith(`
declare let x: { a?: { b?: { c?: string } } };
x.a?.b?.c !== undefined ? x.a?.b?.c : 'default';
`, `
declare let x: { a?: { b?: { c?: string } } };
x.a?.b?.c ?? 'default';
`, "preferNullishOverTernary", optTern(false)),
		// Boolean(...) carve-out: ternary with truthy single-side check.
		invalidWith(`
declare let a: string | undefined;
const x = Boolean(!a ? 'fallback' : a);
`, `
declare let a: string | undefined;
const x = Boolean(a ?? 'fallback');
`, "preferNullishOverTernary", optMap("ignoreBooleanCoercion", true)),
		// Mixed `?.` chain + comparison (object then access).
		invalidWith(`
declare let foo: { bar?: string } | null;
foo?.bar !== null && foo?.bar !== undefined ? foo?.bar : 'fallback';
`, `
declare let foo: { bar?: string } | null;
foo?.bar ?? 'fallback';
`, "preferNullishOverTernary", optTern(false)),
		// `if`-block with prefix-unary `!` directly on member access.
		invalidWith(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
`, `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`, "preferNullishOverAssignment", nil),
		// Mixed comparisons with type that includes `void`.
		invalidWith(`
declare let foo: void | string;
foo !== undefined ? foo : 'fallback';
`, `
declare let foo: void | string;
foo ?? 'fallback';
`, "preferNullishOverTernary", optTern(false)),
		// Multiple chained `||` with nullable + non-nullable mix produces
		// per-`||` reports (issue #1290 family).
		invalidWith(`
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;
a || b || c;
`, `
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;
(a ?? b) || c;
`, "preferNullishOverOr", nil),
	}

	return mustConcatInvalid(
		basicOr, manualOr, ternaryClean, singleSide, truthyTern, optionalChain,
		ifBasic, condTestsInvalid, booleanInvalid, ifFlippedOn, primInvalid,
		noStrict, mixedOr, ifWithCompound, issue1290Final, literalUnions,
		falsyLiterals, mixedUnions, nullishOnly, fnInsideCond,
		tsgoEdges, rslintExtraInvalid, realWorldInvalid, extremeEdges,
		optionalChainBatch, moreRealWorld,
	)
}

// invalidWith builds a generic invalid case with one suggestion. Position
// fields are not asserted (zero values).
func invalidWith(code, output, messageId string, options any) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:    code,
		Options: options,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageId,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: output},
			},
		}},
	}
}

// ──────────────────────────────────────────────────────────────────────────
//                              option helpers
// ──────────────────────────────────────────────────────────────────────────

func opt(k string, v interface{}) map[string]interface{} {
	return map[string]interface{}{k: v}
}
func optTern(v bool) map[string]interface{} {
	return map[string]interface{}{"ignoreTernaryTests": v}
}
func optMap(k string, v interface{}) map[string]interface{} {
	return map[string]interface{}{k: v}
}
func optPrim(m map[string]bool) map[string]interface{} {
	mm := make(map[string]interface{}, len(m))
	for k, v := range m {
		mm[k] = v
	}
	return map[string]interface{}{"ignorePrimitives": mm}
}

// ternFix builds a minimal-suggestion ConditionalExpression invalid case
// where the input has `ignoreTernaryTests: false` and the expected single
// suggestion replaces the entire ternary with `<left> ?? <right>`.
func ternFix(code, output string, column, endColumn, line, endLine int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:    code,
		Options: optTern(false),
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "preferNullishOverTernary",
			Line:      line, EndLine: endLine, Column: column, EndColumn: endColumn,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestNullish", Output: output},
			},
		}},
	}
}
