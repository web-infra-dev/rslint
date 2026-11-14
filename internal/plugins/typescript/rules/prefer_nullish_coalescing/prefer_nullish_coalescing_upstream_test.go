// Cases from the typescript-eslint v8.72.2 prefer-nullish-coalescing test suite.
// Messages, complete ranges and suggestion outputs were recorded by running
// that suite with the same version. The rule has no autofixes.
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.72.2/packages/eslint-plugin/tests/rules/prefer-nullish-coalescing.test.ts

package prefer_nullish_coalescing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferNullishCoalescingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule, upstreamValid, upstreamInvalid)
}

var upstreamValid = []rule_tester.ValidTestCase{
	// valid[0]
	{
		Code: `
declare let x: string;
x || 'foo';
    `,
	},
	// valid[1]
	{
		Code: `
declare let x: number;
x || 'foo';
    `,
	},
	// valid[2]
	{
		Code: `
declare let x: boolean;
x || 'foo';
    `,
	},
	// valid[3]
	{
		Code: `
declare let x: object;
x || 'foo';
    `,
	},
	// valid[4]
	{
		Code: `
declare let x: string;
x ||= 'foo';
    `,
	},
	// valid[5]
	{
		Code: `
declare let x: number;
x ||= 'foo';
    `,
	},
	// valid[6]
	{
		Code: `
declare let x: boolean;
x ||= 'foo';
    `,
	},
	// valid[7]
	{
		Code: `
declare let x: object;
x ||= 'foo';
    `,
	},
	// valid[8]
	{
		Code: `
declare let x: string | null | undefined;
x ?? 'foo';
    `,
	},
	// valid[9]
	{
		Code: `
declare let x: string | null | undefined;
x ??= 'foo';
    `,
	},
	// valid[10]
	{
		Code: `
declare let x: number | null | undefined;
x ?? 'foo';
    `,
	},
	// valid[11]
	{
		Code: `
declare let x: number | null | undefined;
x ??= 'foo';
    `,
	},
	// valid[12]
	{
		Code: `
declare let x: boolean | null | undefined;
x ?? 'foo';
    `,
	},
	// valid[13]
	{
		Code: `
declare let x: boolean | null | undefined;
x ??= 'foo';
    `,
	},
	// valid[14]
	{
		Code: `
declare let x: object | null | undefined;
x ?? 'foo';
    `,
	},
	// valid[15]
	{
		Code: `
declare let x: object | null | undefined;
x ??= 'foo';
    `,
	},
	// valid[16]
	{
		Code:    `x !== undefined && x !== null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": true}},
	},
	// valid[17]
	{
		Code:    `x !== undefined && x !== null ? 'foo' : 'bar';`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[18]
	{
		Code:    `x !== null && x !== undefined && x !== 5 ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[19]
	{
		Code:    `x === null || x === undefined || x === 5 ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[20]
	{
		Code:    `x === undefined && x !== null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[21]
	{
		Code:    `x === undefined && x === null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[22]
	{
		Code:    `x !== undefined && x === null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[23]
	{
		Code:    `x === undefined || x !== null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[24]
	{
		Code:    `x === undefined || x === null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[25]
	{
		Code:    `x !== undefined || x === null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[26]
	{
		Code:    `x !== undefined || x === null ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[27]
	{
		Code:    `x === null || x === null ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[28]
	{
		Code:    `x === undefined || x === undefined ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[29]
	{
		Code:    `x == null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[30]
	{
		Code:    `undefined == null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[31]
	{
		Code:    `undefined != z ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[32]
	{
		Code:    `x == undefined ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[33]
	{
		Code:    `x != null ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[34]
	{
		Code:    `x != undefined ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[35]
	{
		Code:    `null == x ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[36]
	{
		Code:    `undefined == x ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[37]
	{
		Code:    `null != x ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[38]
	{
		Code:    `undefined != x ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[39]
	{
		Code: `
declare let x: number | undefined;
x !== 15 && x !== undefined ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[40]
	{
		Code: `
declare let x: number | undefined;
x !== undefined && x !== 15 ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[41]
	{
		Code: `
declare let x: number | undefined;
15 !== x && undefined !== x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[42]
	{
		Code: `
declare let x: number | undefined;
undefined !== x && 15 !== x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[43]
	{
		Code: `
declare let x: number | undefined;
15 !== x && x !== undefined ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[44]
	{
		Code: `
declare let x: number | undefined;
undefined !== x && x !== 15 ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[45]
	{
		Code: `
declare let x: string | undefined;
x !== 'foo' && x !== undefined ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[46]
	{
		Code: `
function test(value: number | undefined): number {
  return value !== foo() && value !== undefined ? value : 1;
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[47]
	{
		Code: `
const test = (value: boolean | undefined): boolean =>
  value !== undefined && value !== false ? value : false;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[48]
	{
		Code: `
declare let x: string;
x === null ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[49]
	{
		Code: `
declare let x: string | undefined | null;
x === null ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[50]
	{
		Code: `
declare let x: string | undefined | null;
x === undefined ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[51]
	{
		Code: `
declare let x: string | undefined | null;
x !== null ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[52]
	{
		Code: `
declare let x: string | undefined | null;
x !== undefined ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[53]
	{
		Code: `
function test<T>(val: T | undefined) {
  return val !== undefined ? val : 'foo';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[54]
	{
		Code: `
function test<T>(val: T | null) {
  return val !== null ? val : 'foo';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[55]
	{
		Code: `
function test<T extends string | null>(val: T | undefined) {
  return val !== undefined ? val : 'foo';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[56]
	{
		Code: `
class Property<T> {
  value: T | undefined;
  defaultValue!: T;

  get(): T {
    return this.value !== undefined ? this.value : this.defaultValue;
  }
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[57]
	{
		Code: `
function test<T>(val: T | undefined, fallback: T) {
  if (val === undefined) {
    val = fallback;
  }
  return val;
}
      `,
		Options: []any{map[string]any{"ignoreIfStatements": false}},
	},
	// valid[58]
	{
		Code: `
function test<T extends string | undefined>(val: T | null) {
  return val !== null ? val : 'foo';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[59]
	{
		Code: `
function test<T>(val: NoInfer<T> | undefined, fallback: T): T {
  if (val === undefined) {
    val = fallback;
  }
  return val;
}
      `,
		Options: []any{map[string]any{"ignoreIfStatements": false}},
	},
	// valid[60]
	{
		Code: `
function test<T extends Record<string, null>, K extends keyof T>(
  val: T[K] | undefined,
) {
  return val !== undefined ? val : 'fallback';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[61]
	{
		Code: `
function test<T>(val: T) {
  return val !== undefined ? val : 'fallback';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[62]
	{
		Code: `
function test<T extends unknown>(val: T | undefined) {
  return val !== undefined ? val : 'fallback';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[63]
	{
		Code: `
function test<T, U extends T>(val: U | undefined) {
  return val !== undefined ? val : 'fallback';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[64]
	{
		Code: `
declare let x: any;
x === null ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[65]
	{
		Code: `
declare let x: unknown;
x === null ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[66]
	{
		Code: `
declare let x: string;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[67]
	{
		Code: `
declare let x: string;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[68]
	{
		Code: `
declare let x: string | object;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[69]
	{
		Code: `
declare let x: string | object;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[70]
	{
		Code: `
declare let x: number;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[71]
	{
		Code: `
declare let x: number;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[72]
	{
		Code: `
declare let x: bigint;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[73]
	{
		Code: `
declare let x: bigint;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[74]
	{
		Code: `
declare let x: boolean;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[75]
	{
		Code: `
declare let x: boolean;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[76]
	{
		Code: `
declare let x: object;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[77]
	{
		Code: `
declare let x: object;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[78]
	{
		Code: `
declare let x: string[];
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[79]
	{
		Code: `
declare let x: string[];
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[80]
	{
		Code: `
declare let x: Function;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[81]
	{
		Code: `
declare let x: Function;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[82]
	{
		Code: `
declare let x: () => string;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[83]
	{
		Code: `
declare let x: () => string;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[84]
	{
		Code: `
declare let x: () => string | null | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[85]
	{
		Code: `
declare let x: () => string | null | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[86]
	{
		Code: `
declare let x: () => string | null;
x() ? x() : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[87]
	{
		Code: `
declare let x: () => string | null;
!x() ? y : x();
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[88]
	{
		Code: `
const a = 'foo';
declare let x: (a: string | null) => string | null;
x(a) ? x(a) : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[89]
	{
		Code: `
const a = 'foo';
declare let x: (a: string | null) => string | null;
!x(a) ? y : x(a);
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[90]
	{
		Code: `
declare let x: { n: string };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[91]
	{
		Code: `
declare let x: { n: string };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[92]
	{
		Code: `
declare let x: { n: string | object };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[93]
	{
		Code: `
declare let x: { n: string | object };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[94]
	{
		Code: `
declare let x: { n: number };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[95]
	{
		Code: `
declare let x: { n: number };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[96]
	{
		Code: `
declare let x: { n: bigint };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[97]
	{
		Code: `
declare let x: { n: bigint };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[98]
	{
		Code: `
declare let x: { n: boolean };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[99]
	{
		Code: `
declare let x: { n: boolean };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[100]
	{
		Code: `
declare let x: { n: object };
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[101]
	{
		Code: `
declare let x: { n: object };
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[102]
	{
		Code: `
declare let x: { n: string[] };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[103]
	{
		Code: `
declare let x: { n: string[] };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[104]
	{
		Code: `
declare let x: { n: Function };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[105]
	{
		Code: `
declare let x: { n: Function };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[106]
	{
		Code: `
declare let x: { n: () => string };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[107]
	{
		Code: `
declare let x: { n: () => string };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[108]
	{
		Code: `
declare let x: { n: () => string | null | undefined };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[109]
	{
		Code: `
declare let x: { n: () => string | null | undefined };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[110]
	{
		Code: `
declare let foo: string;
declare function makeFoo(): string;

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[111]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo) {
    foo = makeFoo();
  }
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[112]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo != null) {
    foo = makeFoo();
  }
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[113]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
    return foo;
  }
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[114]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  } else {
    return 'bar';
  }
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[115]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  } else if (foo.a) {
    return 'bar';
  }
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[116]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };
function shadowed() {
  if (foo == null) {
    const foo = makeFoo();
  }
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[117]
	{
		Code: `
declare let foo: { foo: string } | null;
declare function makeFoo(): { foo: { foo: string } };
function weirdDestructuringAssignment() {
  if (foo == null) {
    ({ foo } = makeFoo());
  }
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[118]
	{
		Code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject !== undefined && null !== null ? nullOrObject : 42;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[119]
	{
		Code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject !== undefined && null != null ? nullOrObject : 42;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[120]
	{
		Code: `
declare const nullOrObject: null | { a: string };

const test =
  nullOrObject !== undefined && null != undefined ? nullOrObject : 42;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[121]
	{
		Code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject === undefined || null === null ? 42 : nullOrObject;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[122]
	{
		Code: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject === undefined || null == null ? 42 : nullOrObject;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[123]
	{
		Code: `
declare const nullOrObject: null | { a: string };

const test =
  nullOrObject === undefined || null == undefined ? 42 : nullOrObject;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[124]
	{
		Code: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.a != null ? x[a] : 'foo';
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[125]
	{
		Code: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] != null ? x.a : 'foo';
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[126]
	{
		Code: `
declare let x: { a: string } | null;
declare let y: { a: string } | null;

x?.a ? y?.a : 'foo';
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
	},
	// valid[127]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `,
		Options: []any{map[string]any{"ignoreIfStatements": true}},
	},
	// valid[128]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) foo = makeFoo();
}
      `,
		Options: []any{map[string]any{"ignoreIfStatements": true}},
	},
	// ignoreConditionalTests
	// valid[129]
	{
		Code: `
declare let x: string | null | undefined;
x || 'foo' ? null : null;
    `,
	},
	// valid[130]
	{
		Code: `
declare let x: string | null | undefined;
(x ||= 'foo') ? null : null;
    `,
	},
	// valid[131]
	{
		Code: `
declare let x: number | null | undefined;
x || 'foo' ? null : null;
    `,
	},
	// valid[132]
	{
		Code: `
declare let x: number | null | undefined;
(x ||= 'foo') ? null : null;
    `,
	},
	// valid[133]
	{
		Code: `
declare let x: boolean | null | undefined;
x || 'foo' ? null : null;
    `,
	},
	// valid[134]
	{
		Code: `
declare let x: boolean | null | undefined;
(x ||= 'foo') ? null : null;
    `,
	},
	// valid[135]
	{
		Code: `
declare let x: object | null | undefined;
x || 'foo' ? null : null;
    `,
	},
	// valid[136]
	{
		Code: `
declare let x: object | null | undefined;
(x ||= 'foo') ? null : null;
    `,
	},
	// valid[137]
	{
		Code: `
declare let x: string | null | undefined;
if (x || 'foo') {
}
    `,
	},
	// valid[138]
	{
		Code: `
declare let x: string | null | undefined;
if ((x ||= 'foo')) {
}
    `,
	},
	// valid[139]
	{
		Code: `
declare let x: number | null | undefined;
if (x || 'foo') {
}
    `,
	},
	// valid[140]
	{
		Code: `
declare let x: number | null | undefined;
if ((x ||= 'foo')) {
}
    `,
	},
	// valid[141]
	{
		Code: `
declare let x: boolean | null | undefined;
if (x || 'foo') {
}
    `,
	},
	// valid[142]
	{
		Code: `
declare let x: boolean | null | undefined;
if ((x ||= 'foo')) {
}
    `,
	},
	// valid[143]
	{
		Code: `
declare let x: object | null | undefined;
if (x || 'foo') {
}
    `,
	},
	// valid[144]
	{
		Code: `
declare let x: object | null | undefined;
if ((x ||= 'foo')) {
}
    `,
	},
	// valid[145]
	{
		Code: `
declare let x: string | null | undefined;
do {} while (x || 'foo');
    `,
	},
	// valid[146]
	{
		Code: `
declare let x: string | null | undefined;
do {} while ((x ||= 'foo'));
    `,
	},
	// valid[147]
	{
		Code: `
declare let x: number | null | undefined;
do {} while (x || 'foo');
    `,
	},
	// valid[148]
	{
		Code: `
declare let x: number | null | undefined;
do {} while ((x ||= 'foo'));
    `,
	},
	// valid[149]
	{
		Code: `
declare let x: boolean | null | undefined;
do {} while (x || 'foo');
    `,
	},
	// valid[150]
	{
		Code: `
declare let x: boolean | null | undefined;
do {} while ((x ||= 'foo'));
    `,
	},
	// valid[151]
	{
		Code: `
declare let x: object | null | undefined;
do {} while (x || 'foo');
    `,
	},
	// valid[152]
	{
		Code: `
declare let x: object | null | undefined;
do {} while ((x ||= 'foo'));
    `,
	},
	// valid[153]
	{
		Code: `
declare let x: string | null | undefined;
for (; x || 'foo';) {}
    `,
	},
	// valid[154]
	{
		Code: `
declare let x: string | null | undefined;
for (; (x ||= 'foo');) {}
    `,
	},
	// valid[155]
	{
		Code: `
declare let x: number | null | undefined;
for (; x || 'foo';) {}
    `,
	},
	// valid[156]
	{
		Code: `
declare let x: number | null | undefined;
for (; (x ||= 'foo');) {}
    `,
	},
	// valid[157]
	{
		Code: `
declare let x: boolean | null | undefined;
for (; x || 'foo';) {}
    `,
	},
	// valid[158]
	{
		Code: `
declare let x: boolean | null | undefined;
for (; (x ||= 'foo');) {}
    `,
	},
	// valid[159]
	{
		Code: `
declare let x: object | null | undefined;
for (; x || 'foo';) {}
    `,
	},
	// valid[160]
	{
		Code: `
declare let x: object | null | undefined;
for (; (x ||= 'foo');) {}
    `,
	},
	// valid[161]
	{
		Code: `
declare let x: string | null | undefined;
while (x || 'foo') {}
    `,
	},
	// valid[162]
	{
		Code: `
declare let x: string | null | undefined;
while ((x ||= 'foo')) {}
    `,
	},
	// valid[163]
	{
		Code: `
declare let x: number | null | undefined;
while (x || 'foo') {}
    `,
	},
	// valid[164]
	{
		Code: `
declare let x: number | null | undefined;
while ((x ||= 'foo')) {}
    `,
	},
	// valid[165]
	{
		Code: `
declare let x: boolean | null | undefined;
while (x || 'foo') {}
    `,
	},
	// valid[166]
	{
		Code: `
declare let x: boolean | null | undefined;
while ((x ||= 'foo')) {}
    `,
	},
	// valid[167]
	{
		Code: `
declare let x: object | null | undefined;
while (x || 'foo') {}
    `,
	},
	// valid[168]
	{
		Code: `
declare let x: object | null | undefined;
while ((x ||= 'foo')) {}
    `,
	},
	// ignoreMixedLogicalExpressions
	// valid[169]
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
a || (b && c);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[170]
	{
		Code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
a || (b && c);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[171]
	{
		Code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
a || (b && c);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[172]
	{
		Code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
a || (b && c);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[173]
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || b || (c && d);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[174]
	{
		Code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
a || b || (c && d);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[175]
	{
		Code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
a || b || (c && d);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[176]
	{
		Code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
a || b || (c && d);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[177]
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && b) || c || d;
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[178]
	{
		Code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a && b) || c || d;
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[179]
	{
		Code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a && b) || c || d;
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// valid[180]
	{
		Code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a && b) || c || d;
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": true}},
	},
	// ignorePrimitives
	// valid[181]
	{
		Code: `
declare let x: string | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[182]
	{
		Code: `
declare let x: number | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[183]
	{
		Code: `
declare let x: boolean | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"boolean": true}}},
	},
	// valid[184]
	{
		Code: `
declare let x: bigint | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true}}},
	},
	// valid[185]
	{
		Code: `
declare let x: string | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[186]
	{
		Code: `
declare let x: number | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[187]
	{
		Code: `
declare let x: boolean | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[188]
	{
		Code: `
declare let x: bigint | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[189]
	{
		Code: `
declare let x: (string & { __brand?: any }) | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[190]
	{
		Code: `
declare let x: (number & { __brand?: any }) | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[191]
	{
		Code: `
declare let x: (boolean & { __brand?: any }) | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"boolean": true}}},
	},
	// valid[192]
	{
		Code: `
declare let x: (bigint & { __brand?: any }) | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true}}},
	},
	// valid[193]
	{
		Code: `
declare let x: (string & { __brand?: any }) | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[194]
	{
		Code: `
declare let x: (number & { __brand?: any }) | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[195]
	{
		Code: `
declare let x: (boolean & { __brand?: any }) | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[196]
	{
		Code: `
declare let x: (bigint & { __brand?: any }) | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[197]
	{
		Code: `
declare let x: string | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[198]
	{
		Code: `
declare let x: number | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[199]
	{
		Code: `
declare let x: boolean | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"boolean": true}}},
	},
	// valid[200]
	{
		Code: `
declare let x: bigint | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true}}},
	},
	// valid[201]
	{
		Code: `
declare let x: string | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[202]
	{
		Code: `
declare let x: number | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[203]
	{
		Code: `
declare let x: boolean | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"boolean": true}}},
	},
	// valid[204]
	{
		Code: `
declare let x: bigint | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true}}},
	},
	// valid[205]
	{
		Code: `
declare let x: string | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[206]
	{
		Code: `
declare let x: number | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[207]
	{
		Code: `
declare let x: boolean | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[208]
	{
		Code: `
declare let x: bigint | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[209]
	{
		Code: `
declare let x: string | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[210]
	{
		Code: `
declare let x: number | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[211]
	{
		Code: `
declare let x: boolean | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[212]
	{
		Code: `
declare let x: bigint | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[213]
	{
		Code: `
declare let x: (string & { __brand?: any }) | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[214]
	{
		Code: `
declare let x: (number & { __brand?: any }) | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[215]
	{
		Code: `
declare let x: (boolean & { __brand?: any }) | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"boolean": true}}},
	},
	// valid[216]
	{
		Code: `
declare let x: (bigint & { __brand?: any }) | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true}}},
	},
	// valid[217]
	{
		Code: `
declare let x: (string & { __brand?: any }) | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[218]
	{
		Code: `
declare let x: (number & { __brand?: any }) | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[219]
	{
		Code: `
declare let x: (boolean & { __brand?: any }) | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"boolean": true}}},
	},
	// valid[220]
	{
		Code: `
declare let x: (bigint & { __brand?: any }) | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true}}},
	},
	// valid[221]
	{
		Code: `
declare let x: (string & { __brand?: any }) | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[222]
	{
		Code: `
declare let x: (number & { __brand?: any }) | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[223]
	{
		Code: `
declare let x: (boolean & { __brand?: any }) | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[224]
	{
		Code: `
declare let x: (bigint & { __brand?: any }) | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[225]
	{
		Code: `
declare let x: (string & { __brand?: any }) | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[226]
	{
		Code: `
declare let x: (number & { __brand?: any }) | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[227]
	{
		Code: `
declare let x: (boolean & { __brand?: any }) | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[228]
	{
		Code: `
declare let x: (bigint & { __brand?: any }) | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[229]
	{
		Code: `
declare let x: never;
declare let y: number;
x || y;
    `,
	},
	// valid[230]
	{
		Code: `
declare let x: never;
declare let y: number;
x ? x : y;
    `,
	},
	// valid[231]
	{
		Code: `
declare let x: never;
declare let y: number;
!x ? y : x;
    `,
	},
	// valid[232]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== null ? defaultBoxOptional.a?.b : getFallbackBox();
    `,
	},
	// valid[233]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== null ? defaultBoxOptional.a?.b : getFallbackBox();
    `,
	},
	// valid[234]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
    `,
	},
	// valid[235]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a.b
  : getFallbackBox();
    `,
	},
	// valid[236]
	{
		Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
	},
	// valid[237]
	{
		Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
	},
	// valid[238]
	{
		Code: `
declare let x: 0 | 'foo' | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true, "string": true}}},
	},
	// valid[239]
	{
		Code: `
declare let x: 0 | 'foo' | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true, "string": false}}},
	},
	// valid[240]
	{
		Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[241]
	{
		Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[242]
	{
		Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[243]
	{
		Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[244]
	{
		Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
	},
	// valid[245]
	{
		Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
	},
	// valid[246]
	{
		Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
	},
	// valid[247]
	{
		Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
	},
	// valid[248]
	{
		Code: `
declare let x: 0 | 'foo' | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true, "string": true}}},
	},
	// valid[249]
	{
		Code: `
declare let x: 0 | 'foo' | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true, "string": true}}},
	},
	// valid[250]
	{
		Code: `
declare let x: 0 | 'foo' | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true, "string": false}}},
	},
	// valid[251]
	{
		Code: `
declare let x: 0 | 'foo' | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true, "string": false}}},
	},
	// valid[252]
	{
		Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[253]
	{
		Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[254]
	{
		Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[255]
	{
		Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[256]
	{
		Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[257]
	{
		Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[258]
	{
		Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[259]
	{
		Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"string": true}}},
	},
	// valid[260]
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a || b);
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
	},
	// valid[261]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a || b || c);
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
	},
	// valid[262]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a || (b && c));
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
	},
	// valid[263]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean((a || b) ?? c);
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
	},
	// valid[264]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ?? (b || c));
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
	},
	// valid[265]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ? b || c : 'fail');
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
	},
	// valid[266]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ? 'success' : b || c);
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
	},
	// valid[267]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(((a = b), b || c));
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
	},
	// valid[268]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean((a ? a : b) || c);
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
	},
	// valid[269]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(c || (!a ? b : a));
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
	},
	// valid[270]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a || b || c) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[271]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a || (b && c)) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[272]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if ((a || b) ?? c) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[273]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ?? (b || c)) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[274]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ? b || c : 'fail') {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[275]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ? 'success' : b || c) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[276]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (((a = b), b || c)) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[277]
	{
		Code: `
let a: string | undefined;
let b: string | undefined;

if (!(a || b)) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[278]
	{
		Code: `
let a: string | undefined;
let b: string | undefined;

if (!!(a || b)) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[279]
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

if (a ? a : b) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[280]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

if (!a ? b : a) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[281]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if ((a ? a : b) || c) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[282]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (c || (!a ? b : a)) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
	},
	// valid[283]
	{
		Code: `
declare const a: any;
declare const b: any;
a ? a : b;
      `,
		Options: []any{map[string]any{"ignorePrimitives": true}},
	},
	// valid[284]
	{
		Code: `
declare const a: any;
declare const b: any;
a ? a : b;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"number": true}}},
	},
	// valid[285]
	{
		Code: `
declare const a: unknown;
const b = a || 'bar';
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": false, "string": false}}},
	},
}

var upstreamInvalid = []rule_tester.InvalidTestCase{
	// invalid[0]
	{
		Code: `
declare let x: string | null | undefined;
x || 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
x ?? 'foo';
      `},
				},
			},
		},
	},
	// invalid[1]
	{
		Code: `
declare let x: string | null | undefined;
x ||= 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
x ??= 'foo';
      `},
				},
			},
		},
	},
	// invalid[2]
	{
		Code: `
declare let x: number | null | undefined;
x || 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
x ?? 'foo';
      `},
				},
			},
		},
	},
	// invalid[3]
	{
		Code: `
declare let x: number | null | undefined;
x ||= 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
x ??= 'foo';
      `},
				},
			},
		},
	},
	// invalid[4]
	{
		Code: `
declare let x: boolean | null | undefined;
x || 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
x ?? 'foo';
      `},
				},
			},
		},
	},
	// invalid[5]
	{
		Code: `
declare let x: boolean | null | undefined;
x ||= 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
x ??= 'foo';
      `},
				},
			},
		},
	},
	// invalid[6]
	{
		Code: `
declare let x: object | null | undefined;
x || 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
x ?? 'foo';
      `},
				},
			},
		},
	},
	// invalid[7]
	{
		Code: `
declare let x: object | null | undefined;
x ||= 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
x ??= 'foo';
      `},
				},
			},
		},
	},
	// invalid[8]
	{
		Code:    `x !== undefined && x !== null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[9]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[10]
	{
		Code:    `x !== undefined && x !== null ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[11]
	{
		Code:    `x !== null && x !== undefined ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[12]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] !== null &&
x.z[1][this[this.o]]['3'][a.b.c] !== undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[13]
	{
		Code:    `x !== null && x !== undefined ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[14]
	{
		Code:    `x === undefined || x === null ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[15]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] === undefined ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[16]
	{
		Code:    `x === undefined || x === null ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[17]
	{
		Code:    `x === null || x === undefined ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[18]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] === null ||
x.z[1][this[this.o]]['3'][a.b.c] === undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[19]
	{
		Code:    `x === null || x === undefined ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[20]
	{
		Code:    `undefined !== x && x !== null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[21]
	{
		Code: `
undefined !== x.z[1][this[this.o]]['3'][a.b.c] &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[22]
	{
		Code:    `undefined !== x && x !== null ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[23]
	{
		Code:    `null !== x && x !== undefined ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[24]
	{
		Code: `
null !== x.z[1][this[this.o]]['3'][a.b.c] &&
x.z[1][this[this.o]]['3'][a.b.c] !== undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[25]
	{
		Code:    `null !== x && x !== undefined ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[26]
	{
		Code:    `undefined === x || x === null ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[27]
	{
		Code: `
undefined === x.z[1][this[this.o]]['3'][a.b.c] ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[28]
	{
		Code:    `undefined === x || x === null ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[29]
	{
		Code:    `null === x || x === undefined ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[30]
	{
		Code: `
null === x.z[1][this[this.o]]['3'][a.b.c] ||
x.z[1][this[this.o]]['3'][a.b.c] === undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[31]
	{
		Code:    `null === x || x === undefined ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[32]
	{
		Code:    `x !== undefined && null !== x ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[33]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
null !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[34]
	{
		Code:    `x !== undefined && null !== x ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[35]
	{
		Code:    `x !== null && undefined !== x ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[36]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] !== null &&
undefined !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[37]
	{
		Code:    `x !== null && undefined !== x ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[38]
	{
		Code:    `x === undefined || null === x ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[39]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] === undefined ||
null === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[40]
	{
		Code:    `x === undefined || null === x ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[41]
	{
		Code:    `x === null || undefined === x ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[42]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] === null ||
undefined === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[43]
	{
		Code:    `x === null || undefined === x ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[44]
	{
		Code:    `undefined !== x && null !== x ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[45]
	{
		Code: `
undefined !== x.z[1][this[this.o]]['3'][a.b.c] &&
null !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[46]
	{
		Code:    `undefined !== x && null !== x ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[47]
	{
		Code:    `null !== x && undefined !== x ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[48]
	{
		Code: `
null !== x.z[1][this[this.o]]['3'][a.b.c] &&
undefined !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[49]
	{
		Code:    `null !== x && undefined !== x ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[50]
	{
		Code:    `undefined === x || null === x ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[51]
	{
		Code: `
undefined === x.z[1][this[this.o]]['3'][a.b.c] ||
null === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[52]
	{
		Code:    `undefined === x || null === x ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[53]
	{
		Code:    `null === x || undefined === x ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[54]
	{
		Code: `
null === x.z[1][this[this.o]]['3'][a.b.c] ||
undefined === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[55]
	{
		Code:    `null === x || undefined === x ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[56]
	{
		Code:    `x != undefined && x != null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[57]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] != undefined &&
x.z[1][this[this.o]]['3'][a.b.c] != null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[58]
	{
		Code:    `x != undefined && x != null ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 42,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[59]
	{
		Code:    `x == undefined || x == null ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[60]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] == undefined ||
x.z[1][this[this.o]]['3'][a.b.c] == null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[61]
	{
		Code:    `x == undefined || x == null ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 42,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[62]
	{
		Code:    `x != undefined && x !== null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[63]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] != undefined &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[64]
	{
		Code:    `x != undefined && x !== null ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 43,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[65]
	{
		Code:    `x == undefined || x === null ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[66]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] == undefined ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[67]
	{
		Code:    `x == undefined || x === null ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 43,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[68]
	{
		Code:    `x !== undefined && x != null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[69]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
x.z[1][this[this.o]]['3'][a.b.c] != null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 5, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[70]
	{
		Code:    `x !== undefined && x != null ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 43,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[71]
	{
		Code:    `undefined != x ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[72]
	{
		Code: `
undefined != x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 4, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[73]
	{
		Code:    `undefined != x ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[74]
	{
		Code:    `null != x ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[75]
	{
		Code:    `null != x.z[1][this[this.o]]['3'][a.b.c] ? x.z[1][this[this.o]]['3'][a.b.c] : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 80,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;`},
				},
			},
		},
	},
	// invalid[76]
	{
		Code:    `null != x ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[77]
	{
		Code:    `undefined == x ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[78]
	{
		Code: `
undefined == x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 4, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[79]
	{
		Code:    `undefined == x ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[80]
	{
		Code:    `null == x ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[81]
	{
		Code:    `null == x.z[1][this[this.o]]['3'][a.b.c] ? y : x.z[1][this[this.o]]['3'][a.b.c];`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 80,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;`},
				},
			},
		},
	},
	// invalid[82]
	{
		Code:    `null == x ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[83]
	{
		Code:    `x != undefined ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[84]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] != undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 4, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[85]
	{
		Code:    `x != undefined ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[86]
	{
		Code:    `x != null ? x : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[87]
	{
		Code:    `x.z[1][this[this.o]]['3'][a.b.c] != null ? x.z[1][this[this.o]]['3'][a.b.c] : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 80,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;`},
				},
			},
		},
	},
	// invalid[88]
	{
		Code:    `x != null ? x : (z = y);`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[89]
	{
		Code:    `x == undefined ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[90]
	{
		Code: `
x.z[1][this[this.o]]['3'][a.b.c] == undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 4, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
x.z[1][this[this.o]]['3'][a.b.c] ?? y;
      `},
				},
			},
		},
	},
	// invalid[91]
	{
		Code:    `x == undefined ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[92]
	{
		Code:    `x == null ? y : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	// invalid[93]
	{
		Code:    `x.z[1][this[this.o]]['3'][a.b.c] == null ? y : x.z[1][this[this.o]]['3'][a.b.c];`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 80,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;`},
				},
			},
		},
	},
	// invalid[94]
	{
		Code:    `x == null ? (z = y) : x;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? (z = y);`},
				},
			},
		},
	},
	// invalid[95]
	{
		Code:    `this != undefined ? this : y;`,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `this ?? y;`},
				},
			},
		},
	},
	// invalid[96]
	{
		Code: `
declare let x: string | null | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[97]
	{
		Code: `
declare let x: string | null | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[98]
	{
		Code: `
declare let x: number | null | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[99]
	{
		Code: `
declare let x: number | null | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[100]
	{
		Code: `
declare let x: boolean | null | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[101]
	{
		Code: `
declare let x: boolean | null | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[102]
	{
		Code: `
declare let x: object | null | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[103]
	{
		Code: `
declare let x: object | null | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[104]
	{
		Code: `
declare let x: { n: string | null | undefined };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n: string | null | undefined };
x.n ?? y;
      `},
				},
			},
		},
	},
	// invalid[105]
	{
		Code: `
declare let x: { n: string | null | undefined };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n: string | null | undefined };
x.n ?? y;
      `},
				},
			},
		},
	},
	// invalid[106]
	{
		Code: `
declare let x: { n: number | null | undefined };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n: number | null | undefined };
x.n ?? y;
      `},
				},
			},
		},
	},
	// invalid[107]
	{
		Code: `
declare let x: { n: number | null | undefined };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n: number | null | undefined };
x.n ?? y;
      `},
				},
			},
		},
	},
	// invalid[108]
	{
		Code: `
declare let x: { n: boolean | null | undefined };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n: boolean | null | undefined };
x.n ?? y;
      `},
				},
			},
		},
	},
	// invalid[109]
	{
		Code: `
declare let x: { n: boolean | null | undefined };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n: boolean | null | undefined };
x.n ?? y;
      `},
				},
			},
		},
	},
	// invalid[110]
	{
		Code: `
declare let x: { n: object | null | undefined };
x.n ? x.n : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n: object | null | undefined };
x.n ?? y;
      `},
				},
			},
		},
	},
	// invalid[111]
	{
		Code: `
declare let x: { n: object | null | undefined };
!x.n ? y : x.n;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n: object | null | undefined };
x.n ?? y;
      `},
				},
			},
		},
	},
	// invalid[112]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[113]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[114]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[115]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[116]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[117]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 33,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[118]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[119]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 33,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[120]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 32,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[121]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a != null ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[122]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a != null ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[123]
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a != null ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[124]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 53,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[125]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x.n?.a !== undefined && x.n.a !== null ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 51,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[126]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[127]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[128]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[129]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[130]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[131]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[132]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[133]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[134]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[135]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[136]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[137]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 33,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[138]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 30,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[139]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[140]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[141]
	{
		Code: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[142]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 54,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[143]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 53,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[144]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 53,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[145]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 52,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[146]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[147]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[148]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[149]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[150]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x?.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[151]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x.n?.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[152]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x?.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[153]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x.n.a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// invalid[154]
	{
		Code: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? (x?.n).a : y;
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
      `},
				},
			},
		},
	},
	// noStrictNullCheck
	// invalid[155]
	{
		Code: `
declare let x: string[] | null;
if (x) {
}
      `,
		TSConfig: "tsconfig.unstrict.json",
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "noStrictNullCheck",
				Message:   "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
				// Upstream reports `loc: { line: 0, column: 0 }`, outside ESLint's
				// 1-based lines, and expects line 0. rslint reports offset 0 of the
				// file, which is 1:1.
				Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
			},
		},
	},
	// ignoreConditionalTests
	// invalid[156]
	{
		Code: `
declare let x: string | null | undefined;
x || 'foo' ? null : null;
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
x ?? 'foo' ? null : null;
      `},
				},
			},
		},
	},
	// invalid[157]
	{
		Code: `
declare let x: string | null | undefined;
(x ||= 'foo') ? null : null;
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
(x ??= 'foo') ? null : null;
      `},
				},
			},
		},
	},
	// invalid[158]
	{
		Code: `
declare let x: number | null | undefined;
x || 'foo' ? null : null;
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
x ?? 'foo' ? null : null;
      `},
				},
			},
		},
	},
	// invalid[159]
	{
		Code: `
declare let x: number | null | undefined;
(x ||= 'foo') ? null : null;
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
(x ??= 'foo') ? null : null;
      `},
				},
			},
		},
	},
	// invalid[160]
	{
		Code: `
declare let x: boolean | null | undefined;
x || 'foo' ? null : null;
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
x ?? 'foo' ? null : null;
      `},
				},
			},
		},
	},
	// invalid[161]
	{
		Code: `
declare let x: boolean | null | undefined;
(x ||= 'foo') ? null : null;
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
(x ??= 'foo') ? null : null;
      `},
				},
			},
		},
	},
	// invalid[162]
	{
		Code: `
declare let x: object | null | undefined;
x || 'foo' ? null : null;
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
x ?? 'foo' ? null : null;
      `},
				},
			},
		},
	},
	// invalid[163]
	{
		Code: `
declare let x: object | null | undefined;
(x ||= 'foo') ? null : null;
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
(x ??= 'foo') ? null : null;
      `},
				},
			},
		},
	},
	// invalid[164]
	{
		Code: `
declare let x: string | null | undefined;
if (x || 'foo') {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
if (x ?? 'foo') {
}
      `},
				},
			},
		},
	},
	// invalid[165]
	{
		Code: `
declare let x: string | null | undefined;
if ((x ||= 'foo')) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 8, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
if ((x ??= 'foo')) {
}
      `},
				},
			},
		},
	},
	// invalid[166]
	{
		Code: `
declare let x: number | null | undefined;
if (x || 'foo') {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
if (x ?? 'foo') {
}
      `},
				},
			},
		},
	},
	// invalid[167]
	{
		Code: `
declare let x: number | null | undefined;
if ((x ||= 'foo')) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 8, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
if ((x ??= 'foo')) {
}
      `},
				},
			},
		},
	},
	// invalid[168]
	{
		Code: `
declare let x: boolean | null | undefined;
if (x || 'foo') {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
if (x ?? 'foo') {
}
      `},
				},
			},
		},
	},
	// invalid[169]
	{
		Code: `
declare let x: boolean | null | undefined;
if ((x ||= 'foo')) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 8, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
if ((x ??= 'foo')) {
}
      `},
				},
			},
		},
	},
	// invalid[170]
	{
		Code: `
declare let x: object | null | undefined;
if (x || 'foo') {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
if (x ?? 'foo') {
}
      `},
				},
			},
		},
	},
	// invalid[171]
	{
		Code: `
declare let x: object | null | undefined;
if ((x ||= 'foo')) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 8, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
if ((x ??= 'foo')) {
}
      `},
				},
			},
		},
	},
	// invalid[172]
	{
		Code: `
declare let x: string | null | undefined;
do {} while (x || 'foo');
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
do {} while (x ?? 'foo');
      `},
				},
			},
		},
	},
	// invalid[173]
	{
		Code: `
declare let x: string | null | undefined;
do {} while ((x ||= 'foo'));
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 17, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
do {} while ((x ??= 'foo'));
      `},
				},
			},
		},
	},
	// invalid[174]
	{
		Code: `
declare let x: number | null | undefined;
do {} while (x || 'foo');
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
do {} while (x ?? 'foo');
      `},
				},
			},
		},
	},
	// invalid[175]
	{
		Code: `
declare let x: number | null | undefined;
do {} while ((x ||= 'foo'));
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 17, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
do {} while ((x ??= 'foo'));
      `},
				},
			},
		},
	},
	// invalid[176]
	{
		Code: `
declare let x: boolean | null | undefined;
do {} while (x || 'foo');
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
do {} while (x ?? 'foo');
      `},
				},
			},
		},
	},
	// invalid[177]
	{
		Code: `
declare let x: boolean | null | undefined;
do {} while ((x ||= 'foo'));
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 17, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
do {} while ((x ??= 'foo'));
      `},
				},
			},
		},
	},
	// invalid[178]
	{
		Code: `
declare let x: object | null | undefined;
do {} while (x || 'foo');
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
do {} while (x ?? 'foo');
      `},
				},
			},
		},
	},
	// invalid[179]
	{
		Code: `
declare let x: object | null | undefined;
do {} while ((x ||= 'foo'));
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 17, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
do {} while ((x ??= 'foo'));
      `},
				},
			},
		},
	},
	// invalid[180]
	{
		Code: `
declare let x: string | null | undefined;
for (; x || 'foo';) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
for (; x ?? 'foo';) {}
      `},
				},
			},
		},
	},
	// invalid[181]
	{
		Code: `
declare let x: string | null | undefined;
for (; (x ||= 'foo');) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 11, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
for (; (x ??= 'foo');) {}
      `},
				},
			},
		},
	},
	// invalid[182]
	{
		Code: `
declare let x: number | null | undefined;
for (; x || 'foo';) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
for (; x ?? 'foo';) {}
      `},
				},
			},
		},
	},
	// invalid[183]
	{
		Code: `
declare let x: number | null | undefined;
for (; (x ||= 'foo');) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 11, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
for (; (x ??= 'foo');) {}
      `},
				},
			},
		},
	},
	// invalid[184]
	{
		Code: `
declare let x: boolean | null | undefined;
for (; x || 'foo';) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
for (; x ?? 'foo';) {}
      `},
				},
			},
		},
	},
	// invalid[185]
	{
		Code: `
declare let x: boolean | null | undefined;
for (; (x ||= 'foo');) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 11, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
for (; (x ??= 'foo');) {}
      `},
				},
			},
		},
	},
	// invalid[186]
	{
		Code: `
declare let x: object | null | undefined;
for (; x || 'foo';) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
for (; x ?? 'foo';) {}
      `},
				},
			},
		},
	},
	// invalid[187]
	{
		Code: `
declare let x: object | null | undefined;
for (; (x ||= 'foo');) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 11, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
for (; (x ??= 'foo');) {}
      `},
				},
			},
		},
	},
	// invalid[188]
	{
		Code: `
declare let x: string | null | undefined;
while (x || 'foo') {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
while (x ?? 'foo') {}
      `},
				},
			},
		},
	},
	// invalid[189]
	{
		Code: `
declare let x: string | null | undefined;
while ((x ||= 'foo')) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 11, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
while ((x ??= 'foo')) {}
      `},
				},
			},
		},
	},
	// invalid[190]
	{
		Code: `
declare let x: number | null | undefined;
while (x || 'foo') {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
while (x ?? 'foo') {}
      `},
				},
			},
		},
	},
	// invalid[191]
	{
		Code: `
declare let x: number | null | undefined;
while ((x ||= 'foo')) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 11, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
while ((x ??= 'foo')) {}
      `},
				},
			},
		},
	},
	// invalid[192]
	{
		Code: `
declare let x: boolean | null | undefined;
while (x || 'foo') {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
while (x ?? 'foo') {}
      `},
				},
			},
		},
	},
	// invalid[193]
	{
		Code: `
declare let x: boolean | null | undefined;
while ((x ||= 'foo')) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 11, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
while ((x ??= 'foo')) {}
      `},
				},
			},
		},
	},
	// invalid[194]
	{
		Code: `
declare let x: object | null | undefined;
while (x || 'foo') {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
while (x ?? 'foo') {}
      `},
				},
			},
		},
	},
	// invalid[195]
	{
		Code: `
declare let x: object | null | undefined;
while ((x ||= 'foo')) {}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 11, EndLine: 3, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
while ((x ??= 'foo')) {}
      `},
				},
			},
		},
	},
	// ignoreMixedLogicalExpressions
	// invalid[196]
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
a || (b && c);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 3, EndLine: 5, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
a ?? (b && c);
      `},
				},
			},
		},
	},
	// invalid[197]
	{
		Code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
a || (b && c);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 3, EndLine: 5, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
a ?? (b && c);
      `},
				},
			},
		},
	},
	// invalid[198]
	{
		Code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
a || (b && c);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 3, EndLine: 5, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
a ?? (b && c);
      `},
				},
			},
		},
	},
	// invalid[199]
	{
		Code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
a || (b && c);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 3, EndLine: 5, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
a ?? (b && c);
      `},
				},
			},
		},
	},
	// invalid[200]
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || b || (c && d);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 3, EndLine: 6, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a ?? b) || (c && d);
      `},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 8, EndLine: 6, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || b ?? (c && d);
      `},
				},
			},
		},
	},
	// invalid[201]
	{
		Code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
a || b || (c && d);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 3, EndLine: 6, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a ?? b) || (c && d);
      `},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 8, EndLine: 6, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
a || b ?? (c && d);
      `},
				},
			},
		},
	},
	// invalid[202]
	{
		Code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
a || b || (c && d);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 3, EndLine: 6, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a ?? b) || (c && d);
      `},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 8, EndLine: 6, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
a || b ?? (c && d);
      `},
				},
			},
		},
	},
	// invalid[203]
	{
		Code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
a || b || (c && d);
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 3, EndLine: 6, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a ?? b) || (c && d);
      `},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 8, EndLine: 6, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
a || b ?? (c && d);
      `},
				},
			},
		},
	},
	// invalid[204]
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && b) || c || d;
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 10, EndLine: 6, EndColumn: 12,
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
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 15, EndLine: 6, EndColumn: 17,
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
		},
	},
	// invalid[205]
	{
		Code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a && b) || c || d;
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 10, EndLine: 6, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a && (b) ?? c) || d;
      `},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 15, EndLine: 6, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a && b) || c ?? d;
      `},
				},
			},
		},
	},
	// invalid[206]
	{
		Code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a && b) || c || d;
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 10, EndLine: 6, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a && (b) ?? c) || d;
      `},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 15, EndLine: 6, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a && b) || c ?? d;
      `},
				},
			},
		},
	},
	// invalid[207]
	{
		Code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a && b) || c || d;
      `,
		Options: []any{map[string]any{"ignoreMixedLogicalExpressions": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 10, EndLine: 6, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a && (b) ?? c) || d;
      `},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 15, EndLine: 6, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a && b) || c ?? d;
      `},
				},
			},
		},
	},
	// should not false positive for functions inside conditional tests
	// invalid[208]
	{
		Code: `
declare let x: string | null | undefined;
if (() => x || 'foo') {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
if (() => x ?? 'foo') {
}
      `},
				},
			},
		},
	},
	// invalid[209]
	{
		Code: `
declare let x: string | null | undefined;
if (() => (x ||= 'foo')) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
if (() => (x ??= 'foo')) {
}
      `},
				},
			},
		},
	},
	// invalid[210]
	{
		Code: `
declare let x: number | null | undefined;
if (() => x || 'foo') {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
if (() => x ?? 'foo') {
}
      `},
				},
			},
		},
	},
	// invalid[211]
	{
		Code: `
declare let x: number | null | undefined;
if (() => (x ||= 'foo')) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
if (() => (x ??= 'foo')) {
}
      `},
				},
			},
		},
	},
	// invalid[212]
	{
		Code: `
declare let x: boolean | null | undefined;
if (() => x || 'foo') {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
if (() => x ?? 'foo') {
}
      `},
				},
			},
		},
	},
	// invalid[213]
	{
		Code: `
declare let x: boolean | null | undefined;
if (() => (x ||= 'foo')) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
if (() => (x ??= 'foo')) {
}
      `},
				},
			},
		},
	},
	// invalid[214]
	{
		Code: `
declare let x: object | null | undefined;
if (() => x || 'foo') {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
if (() => x ?? 'foo') {
}
      `},
				},
			},
		},
	},
	// invalid[215]
	{
		Code: `
declare let x: object | null | undefined;
if (() => (x ||= 'foo')) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
if (() => (x ??= 'foo')) {
}
      `},
				},
			},
		},
	},
	// invalid[216]
	{
		Code: `
declare let x: string | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 14, EndLine: 5, EndColumn: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
if (
  function weird() {
    return x ?? 'foo';
  }
) {
}
      `},
				},
			},
		},
	},
	// invalid[217]
	{
		Code: `
declare let x: string | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      5, Column: 15, EndLine: 5, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
if (
  function weird() {
    return (x ??= 'foo');
  }
) {
}
      `},
				},
			},
		},
	},
	// invalid[218]
	{
		Code: `
declare let x: number | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 14, EndLine: 5, EndColumn: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
if (
  function weird() {
    return x ?? 'foo';
  }
) {
}
      `},
				},
			},
		},
	},
	// invalid[219]
	{
		Code: `
declare let x: number | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      5, Column: 15, EndLine: 5, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
if (
  function weird() {
    return (x ??= 'foo');
  }
) {
}
      `},
				},
			},
		},
	},
	// invalid[220]
	{
		Code: `
declare let x: boolean | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 14, EndLine: 5, EndColumn: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
if (
  function weird() {
    return x ?? 'foo';
  }
) {
}
      `},
				},
			},
		},
	},
	// invalid[221]
	{
		Code: `
declare let x: boolean | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      5, Column: 15, EndLine: 5, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
if (
  function weird() {
    return (x ??= 'foo');
  }
) {
}
      `},
				},
			},
		},
	},
	// invalid[222]
	{
		Code: `
declare let x: object | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 14, EndLine: 5, EndColumn: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
if (
  function weird() {
    return x ?? 'foo';
  }
) {
}
      `},
				},
			},
		},
	},
	// invalid[223]
	{
		Code: `
declare let x: object | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      5, Column: 15, EndLine: 5, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
if (
  function weird() {
    return (x ??= 'foo');
  }
) {
}
      `},
				},
			},
		},
	},
	// https://github.com/typescript-eslint/typescript-eslint/issues/1290
	// invalid[224]
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;
a || b || c;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 3, EndLine: 5, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;
(a ?? b) || c;
      `},
				},
			},
		},
	},
	// invalid[225]
	{
		Code: `
declare let a: number | null | undefined;
declare let b: number;
declare let c: number;
a || b || c;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 3, EndLine: 5, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: number | null | undefined;
declare let b: number;
declare let c: number;
(a ?? b) || c;
      `},
				},
			},
		},
	},
	// invalid[226]
	{
		Code: `
declare let a: boolean | null | undefined;
declare let b: boolean;
declare let c: boolean;
a || b || c;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 3, EndLine: 5, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: boolean | null | undefined;
declare let b: boolean;
declare let c: boolean;
(a ?? b) || c;
      `},
				},
			},
		},
	},
	// invalid[227]
	{
		Code: `
declare let a: object | null | undefined;
declare let b: object;
declare let c: object;
a || b || c;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 3, EndLine: 5, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: object | null | undefined;
declare let b: object;
declare let c: object;
(a ?? b) || c;
      `},
				},
			},
		},
	},
	// default for missing option
	// invalid[228]
	{
		Code: `
declare let x: string | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[229]
	{
		Code: `
declare let x: number | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[230]
	{
		Code: `
declare let x: boolean | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[231]
	{
		Code: `
declare let x: bigint | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: bigint | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[232]
	{
		Code: `
declare let x: string | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[233]
	{
		Code: `
declare let x: number | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[234]
	{
		Code: `
declare let x: boolean | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[235]
	{
		Code: `
declare let x: bigint | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: bigint | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// falsy
	// invalid[236]
	{
		Code: `
declare let x: '' | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: '' | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[237]
	{
		Code:    "\ndeclare let x: `` | undefined;\nx || y;\n      ",
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "\ndeclare let x: `` | undefined;\nx ?? y;\n      "},
				},
			},
		},
	},
	// invalid[238]
	{
		Code: `
declare let x: 0 | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[239]
	{
		Code: `
declare let x: 0n | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[240]
	{
		Code: `
declare let x: false | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: false | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[241]
	{
		Code: `
declare let x: '' | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: '' | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[242]
	{
		Code:    "\ndeclare let x: `` | undefined;\nx ? x : y;\n      ",
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "\ndeclare let x: `` | undefined;\nx ?? y;\n      "},
				},
			},
		},
	},
	// invalid[243]
	{
		Code: `
declare let x: 0 | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[244]
	{
		Code: `
declare let x: 0n | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[245]
	{
		Code: `
declare let x: false | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: false | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// truthy
	// invalid[246]
	{
		Code: `
declare let x: 'a' | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 'a' | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[247]
	{
		Code:    "\ndeclare let x: `hello${'string'}` | undefined;\nx || y;\n      ",
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "\ndeclare let x: `hello${'string'}` | undefined;\nx ?? y;\n      "},
				},
			},
		},
	},
	// invalid[248]
	{
		Code: `
declare let x: 1 | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[249]
	{
		Code: `
declare let x: 1n | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[250]
	{
		Code: `
declare let x: true | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: true | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[251]
	{
		Code: `
declare let x: 'a' | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 'a' | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[252]
	{
		Code: `
declare let x: 'a' | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 'a' | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[253]
	{
		Code:    "\ndeclare let x: `hello${'string'}` | undefined;\nx ? x : y;\n      ",
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "\ndeclare let x: `hello${'string'}` | undefined;\nx ?? y;\n      "},
				},
			},
		},
	},
	// invalid[254]
	{
		Code:    "\ndeclare let x: `hello${'string'}` | undefined;\n!x ? y : x;\n      ",
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "\ndeclare let x: `hello${'string'}` | undefined;\nx ?? y;\n      "},
				},
			},
		},
	},
	// invalid[255]
	{
		Code: `
declare let x: 1 | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[256]
	{
		Code: `
declare let x: 1 | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[257]
	{
		Code: `
declare let x: 1n | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[258]
	{
		Code: `
declare let x: 1n | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[259]
	{
		Code: `
declare let x: true | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: true | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[260]
	{
		Code: `
declare let x: true | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: true | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// Unions of same primitive
	// invalid[261]
	{
		Code: `
declare let x: 'a' | 'b' | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[262]
	{
		Code:    "\ndeclare let x: 'a' | `b` | undefined;\nx || y;\n      ",
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "\ndeclare let x: 'a' | `b` | undefined;\nx ?? y;\n      "},
				},
			},
		},
	},
	// invalid[263]
	{
		Code: `
declare let x: 0 | 1 | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | 1 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[264]
	{
		Code: `
declare let x: 1 | 2 | 3 | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[265]
	{
		Code: `
declare let x: 0n | 1n | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0n | 1n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[266]
	{
		Code: `
declare let x: 1n | 2n | 3n | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[267]
	{
		Code: `
declare let x: true | false | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: true | false | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[268]
	{
		Code: `
declare let x: 'a' | 'b' | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[269]
	{
		Code: `
declare let x: 'a' | 'b' | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[270]
	{
		Code:    "\ndeclare let x: 'a' | `b` | undefined;\nx ? x : y;\n      ",
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "\ndeclare let x: 'a' | `b` | undefined;\nx ?? y;\n      "},
				},
			},
		},
	},
	// invalid[271]
	{
		Code:    "\ndeclare let x: 'a' | `b` | undefined;\n!x ? y : x;\n      ",
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": true, "string": false}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "\ndeclare let x: 'a' | `b` | undefined;\nx ?? y;\n      "},
				},
			},
		},
	},
	// invalid[272]
	{
		Code: `
declare let x: 0 | 1 | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | 1 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[273]
	{
		Code: `
declare let x: 0 | 1 | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | 1 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[274]
	{
		Code: `
declare let x: 1 | 2 | 3 | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[275]
	{
		Code: `
declare let x: 1 | 2 | 3 | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[276]
	{
		Code: `
declare let x: 0n | 1n | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0n | 1n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[277]
	{
		Code: `
declare let x: 0n | 1n | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0n | 1n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[278]
	{
		Code: `
declare let x: 1n | 2n | 3n | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[279]
	{
		Code: `
declare let x: 1n | 2n | 3n | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[280]
	{
		Code: `
declare let x: true | false | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: true | false | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[281]
	{
		Code: `
declare let x: true | false | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: true | false | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// Mixed unions
	// invalid[282]
	{
		Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[283]
	{
		Code: `
declare let x: true | false | null | undefined;
x || y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: true | false | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[284]
	{
		Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[285]
	{
		Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": false, "boolean": true, "number": false, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[286]
	{
		Code: `
declare let x: true | false | null | undefined;
x ? x : y;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: true | false | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[287]
	{
		Code: `
declare let x: true | false | null | undefined;
!x ? y : x;
      `,
		Options: []any{map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": false, "number": true, "string": true}}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: true | false | null | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[288]
	{
		Code: `
declare let x: null;
x || y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: null;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[289]
	{
		Code: `
const x = undefined;
x || y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
const x = undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[290]
	{
		Code: `
null || y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 6, EndLine: 2, EndColumn: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
null ?? y;
      `},
				},
			},
		},
	},
	// invalid[291]
	{
		Code: `
undefined || y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 11, EndLine: 2, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
undefined ?? y;
      `},
				},
			},
		},
	},
	// invalid[292]
	{
		Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x || y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      8, Column: 3, EndLine: 8, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[293]
	{
		Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      8, Column: 3, EndLine: 8, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[294]
	{
		Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x || y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      8, Column: 3, EndLine: 8, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[295]
	{
		Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      8, Column: 3, EndLine: 8, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[296]
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
let c: boolean | undefined;

const x = Boolean(a || b);
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 21, EndLine: 6, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;
let c: boolean | undefined;

const x = Boolean(a ?? b);
      `},
				},
			},
		},
	},
	// invalid[297]
	{
		Code: `
function outer() {
  const Boolean = (x: unknown) => x;

  return (a: string | null, b: string) => Boolean(a || b);
}
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 53, EndLine: 5, EndColumn: 55,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
function outer() {
  const Boolean = (x: unknown) => x;

  return (a: string | null, b: string) => Boolean(a ?? b);
}
      `},
				},
			},
		},
	},
	// invalid[298]
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = String(a || b);
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 20, EndLine: 5, EndColumn: 22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = String(a ?? b);
      `},
				},
			},
		},
	},
	// invalid[299]
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(() => a || b);
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 27, EndLine: 5, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(() => a ?? b);
      `},
				},
			},
		},
	},
	// invalid[300]
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(function weird() {
  return a || b;
});
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 12, EndLine: 6, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(function weird() {
  return a ?? b;
});
      `},
				},
			},
		},
	},
	// invalid[301]
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

const x = Boolean(f(a || b));
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      7, Column: 23, EndLine: 7, EndColumn: 25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

const x = Boolean(f(a ?? b));
      `},
				},
			},
		},
	},
	// invalid[302]
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(1 + (a || b));
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 26, EndLine: 5, EndColumn: 28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(1 + (a ?? b));
      `},
				},
			},
		},
	},
	// invalid[303]
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a ? a : b);
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      5, Column: 19, EndLine: 5, EndColumn: 28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a ?? b);
      `},
				},
			},
		},
	},
	// invalid[304]
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

const test = Boolean(!a ? b : a);
      `,
		Options: []any{map[string]any{"ignoreBooleanCoercion": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      5, Column: 22, EndLine: 5, EndColumn: 32,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

const test = Boolean(a ?? b);
      `},
				},
			},
		},
	},
	// invalid[305]
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

if (f(a || b)) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      7, Column: 9, EndLine: 7, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

if (f(a ?? b)) {
}
      `},
				},
			},
		},
	},
	// invalid[306]
	{
		Code: `
declare const a: string | undefined;
declare const b: string;

if (+(a || b)) {
}
      `,
		Options: []any{map[string]any{"ignoreConditionalTests": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 9, EndLine: 5, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const a: string | undefined;
declare const b: string;

if (+(a ?? b)) {
}
      `},
				},
			},
		},
	},
	// invalid[307]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox || getFallbackBox();
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      8, Column: 12, EndLine: 8, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ?? getFallbackBox();
      `},
				},
			},
		},
	},
	// invalid[308]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ? defaultBox : getFallbackBox();
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      8, Column: 1, EndLine: 8, EndColumn: 43,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ?? getFallbackBox();
      `},
				},
			},
		},
	},
	// invalid[309]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b != null ? defaultBoxOptional.a?.b : getFallbackBox();
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      8, Column: 1, EndLine: 8, EndColumn: 77,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `},
				},
			},
		},
	},
	// invalid[310]
	{
		Code: `
declare const x: any;
declare const y: any;
x || y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      4, Column: 3, EndLine: 4, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const x: any;
declare const y: any;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[311]
	{
		Code: `
declare const x: unknown;
declare const y: any;
x || y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      4, Column: 3, EndLine: 4, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const x: unknown;
declare const y: any;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[312]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b != null ? defaultBoxOptional.a.b : getFallbackBox();
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      8, Column: 1, EndLine: 8, EndColumn: 76,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `},
				},
			},
		},
	},
	// invalid[313]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ? defaultBoxOptional.a?.b : getFallbackBox();
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      8, Column: 1, EndLine: 8, EndColumn: 69,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `},
				},
			},
		},
	},
	// invalid[314]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ? defaultBoxOptional.a.b : getFallbackBox();
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      8, Column: 1, EndLine: 8, EndColumn: 68,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `},
				},
			},
		},
	},
	// invalid[315]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      8, Column: 1, EndLine: 10, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `},
				},
			},
		},
	},
	// invalid[316]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a.b
  : getFallbackBox();
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      8, Column: 1, EndLine: 10, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `},
				},
			},
		},
	},
	// invalid[317]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined && defaultBoxOptional.a?.b !== null
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      8, Column: 1, EndLine: 10, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `},
				},
			},
		},
	},
	// invalid[318]
	{
		Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined && defaultBoxOptional.a?.b !== null
  ? defaultBoxOptional.a.b
  : getFallbackBox();
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      8, Column: 1, EndLine: 10, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `},
				},
			},
		},
	},
	// invalid[319]
	{
		Code: `
declare let x: unknown;
declare let y: number;
!x ? y : x;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      4, Column: 1, EndLine: 4, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: unknown;
declare let y: number;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[320]
	{
		Code: `
declare let x: unknown;
declare let y: number;
x ? x : y;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      4, Column: 1, EndLine: 4, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: unknown;
declare let y: number;
x ?? y;
      `},
				},
			},
		},
	},
	// invalid[321]
	{
		Code: `
declare let x: { n: unknown };
!x.n ? y : x.n;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n: unknown };
x.n ?? y;
      `},
				},
			},
		},
	},
	// invalid[322]
	{
		Code: `
declare let x: { a: string } | null;

x?.['a'] != null ? x['a'] : 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      4, Column: 1, EndLine: 4, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { a: string } | null;

x?.['a'] ?? 'foo';
      `},
				},
			},
		},
	},
	// invalid[323]
	{
		Code: `
declare let x: { a: string } | null;

x?.['a'] != null ? x.a : 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      4, Column: 1, EndLine: 4, EndColumn: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { a: string } | null;

x?.['a'] ?? 'foo';
      `},
				},
			},
		},
	},
	// invalid[324]
	{
		Code: `
declare let x: { a: string } | null;

x?.a != null ? x['a'] : 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      4, Column: 1, EndLine: 4, EndColumn: 30,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { a: string } | null;

x?.a ?? 'foo';
      `},
				},
			},
		},
	},
	// invalid[325]
	{
		Code: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] != null ? x[a] : 'foo';
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      5, Column: 1, EndLine: 5, EndColumn: 30,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] ?? 'foo';
      `},
				},
			},
		},
	},
	// invalid[326]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `},
				},
			},
		},
	},
	// invalid[327]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `},
				},
			},
		},
	},
	// invalid[328]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ??= makeFoo();
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `},
				},
			},
		},
	},
	// Multiple errors
	/* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
	// invalid[329]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ||= makeFoo();
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      7, Column: 9, EndLine: 7, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ??= makeFoo();
  }
}
      `},
				},
			},
		},
	},
	/* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */
	// invalid[330]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === null) {
    foo = makeFoo();
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `},
				},
			},
		},
	},
	// invalid[331]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo = makeFoo();
  const bar = 42;
  return bar;
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 6, EndColumn: 36,
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
		},
	},
	// invalid[332]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 6, EndColumn: 38,
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
		},
	},
	// invalid[333]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ||= makeFoo();
  const bar = 42;
  return bar;
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 6, EndColumn: 38,
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
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      6, Column: 24, EndLine: 6, EndColumn: 27,
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
		},
	},
	// invalid[334]
	{
		Code: `
declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === undefined) {
    foo = makeFoo();
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `},
				},
			},
		},
	},
	// invalid[335]
	{
		Code: `
declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === undefined || foo === null) {
    foo = makeFoo();
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `},
				},
			},
		},
	},
	// invalid[336]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo.a == null) {
    foo.a = makeFoo();
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
      `},
				},
			},
		},
	},
	// invalid[337]
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo?.a == null) {
    foo.a = makeFoo();
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
      `},
				},
			},
		},
	},
	// invalid[338]
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    // comment
    foo = makeFoo();
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 9, EndColumn: 4,
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
			},
		},
	},
	// invalid[339]
	{
		Code: `
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
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      5, Column: 1, EndLine: 25, EndColumn: 2,
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
			},
		},
	},
	// invalid[340]
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) /* comment before 1 */ /* comment before 2 */ foo = makeFoo(); // comment inline
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      5, Column: 1, EndLine: 5, EndColumn: 80,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

/* comment before 1 */ /* comment before 2 */ foo ??= makeFoo(); // comment inline
      `},
				},
			},
		},
	},
	// invalid[341]
	{
		Code: `
declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  if (((((foo.a)) == null))) {
    ((((((((foo).a))))) = makeString()));
  }
}
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  ((foo).a) ??= makeString();
}
      `},
				},
			},
		},
	},
	// invalid[342]
	{
		Code: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ? a : b ? 1 : 2;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      5, Column: 13, EndLine: 5, EndColumn: 30,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
      `},
				},
			},
		},
	},
	// invalid[343]
	{
		Code: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ? a : (b ? 1 : 2);
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      5, Column: 13, EndLine: 5, EndColumn: 32,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
      `},
				},
			},
		},
	},
	// invalid[344]
	{
		Code: `
declare const c: string | null;
c !== null ? c : c ? 1 : 2;
      `,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const c: string | null;
c ?? (c ? 1 : 2);
      `},
				},
			},
		},
	},
	// invalid[345]
	{
		Code: `
function test<T extends string>(val: T | undefined) {
  return val !== undefined ? val : 'foo';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 41,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
function test<T extends string>(val: T | undefined) {
  return val ?? 'foo';
}
      `},
				},
			},
		},
	},
	// invalid[346]
	{
		Code: `
function test<T extends {}>(val: T | undefined) {
  return val !== undefined ? val : 'fallback';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 46,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
function test<T extends {}>(val: T | undefined) {
  return val ?? 'fallback';
}
      `},
				},
			},
		},
	},
	// invalid[347]
	{
		Code: `
function test<T>(val: T | undefined) {
  return val != undefined ? val : 'fallback';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 45,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
function test<T>(val: T | undefined) {
  return val ?? 'fallback';
}
      `},
				},
			},
		},
	},
	// invalid[348]
	{
		Code: `
function test<T>(val: T | undefined) {
  return val !== undefined && val !== null ? val : 'fallback';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 62,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
function test<T>(val: T | undefined) {
  return val ?? 'fallback';
}
      `},
				},
			},
		},
	},
	// invalid[349]
	{
		Code: `
function test<T extends string, U extends T>(val: U | undefined) {
  return val !== undefined ? val : 'fallback';
}
      `,
		Options: []any{map[string]any{"ignoreTernaryTests": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 46,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
function test<T extends string, U extends T>(val: U | undefined) {
  return val ?? 'fallback';
}
      `},
				},
			},
		},
	},
	// invalid[350]
	{
		Code: `
function test<T extends string>(val: T | undefined, fallback: T) {
  if (val === undefined) {
    val = fallback;
  }
  return val;
}
      `,
		Options: []any{map[string]any{"ignoreIfStatements": false}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      3, Column: 3, EndLine: 5, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
function test<T extends string>(val: T | undefined, fallback: T) {
  val ??= fallback;
  return val;
}
      `},
				},
			},
		},
	},
}
