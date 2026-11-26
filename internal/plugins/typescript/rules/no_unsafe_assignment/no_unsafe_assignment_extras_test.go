package no_unsafe_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeAssignmentFastPathExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnsafeAssignmentRule, []rule_tester.ValidTestCase{
		// Guaranteed non-any / non-TypeReference expressions across all sender positions.
		{
			Code: `
declare const anyVal: any;
const nested = {
  str: 'hello',
  noSub: ` + "`plain`" + `,
  num: 42,
  big: 42n,
  t: true,
  f: false,
  n: null,
  typeOf: typeof anyVal,
  deleted: delete (anyVal as { a?: number }).a,
  voided: void anyVal,
  arrow: () => anyVal,
  fn: function () { return anyVal; },
  nested: { a: (1), b: (({ c: true })) },
};
const paren = (('hello'));
const spreadString = [...'abc', ...` + "`plain`" + `];
class Box {
  a = 'hello';
  b = { c: () => anyVal };
  d = void 0;
}
function withDefaults(a = 'x', b = { c: 1 }, c = () => anyVal) {
  return [a, b, c];
}
      `,
		},
		// JSX attributes with guaranteed safe expressions.
		{
			Code: `
type Props = {
  str: string;
  num: number;
  flag: boolean;
  cb: () => void;
  style: { color: string };
};
declare function Foo(props: Props): never;
<Foo str="ok" num={1} flag={true} cb={() => {}} style={{ color: 'red' }} />;
      `,
			Tsx: true,
		},
		// strictNullChecks: false — assigning null or void 0 directly is still safe.
		{
			Code: `
const a = null;
let b = null;
const c = void 0;
let d = void 0;
const obj: { a: string; b: string } = { a: (null), b: (void 0) };
class C {
  x = null;
  y = void 0;
}
      `,
			TSConfig: "tsconfig.unstrict.json",
		},
	}, []rule_tester.InvalidTestCase{
		// JSDoc @type {any} / @satisfies casts in JS files must never be skipped.
		{
			Code: `
const a = /** @type {any} */ ("foo");
const b = /** @type {any} */ ({ x: 1 });
const c = /** @type {any} */ (() => {});
const d = /** @type {any} */ (function () {});
const e = /** @satisfies {MissingType} */ ("foo");
const obj = {
  p1: /** @type {any} */ ("foo"),
  p2: /** @type {any} */ ({ x: 1 }),
  p3: (/** @type {any} */ ("bar")),
};
let assigned = '';
assigned = /** @type {any} */ ("baz");
class Holder {
  prop = /** @type {any} */ ("qux");
}
      `,
			FileName: "jsdoc-cast.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment", Line: 2},
				{MessageId: "anyAssignment", Line: 3},
				{MessageId: "anyAssignment", Line: 4},
				{MessageId: "anyAssignment", Line: 5},
				{MessageId: "anyAssignment", Line: 6},
				{MessageId: "anyAssignment", Line: 8},
				{MessageId: "anyAssignment", Line: 9},
				{MessageId: "anyAssignment", Line: 10},
				{MessageId: "anyAssignment", Line: 13},
				{MessageId: "anyAssignment", Line: 15},
			},
		},
		// Unparenthesized JSDoc @satisfies on variable declaration and object property
		// must also never be skipped.
		{
			Code: `
/** @satisfies {MissingType} */
const f = "foo";
const obj = {
  /** @satisfies {MissingType} */
  p: "foo",
};
      `,
			FileName: "jsdoc-satisfies.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment", Line: 3},
				{MessageId: "anyAssignment", Line: 6},
			},
		},
		// TS2590 template-literal cross-product overflow returns errorType (TypeFlagsAny).
		{
			Code: `
type Digit = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9';
declare const d: Digit;
const x: ` + "`${string}_`" + ` = ` + "`${d}${d}${d}${d}${d}_`" + `;
const obj: { a: ` + "`${string}_`" + ` } = { a: ` + "`${d}${d}${d}${d}${d}_`" + ` };
const spread = [...` + "`${d}${d}${d}${d}${d}_`" + `] as const;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment", Line: 4},
				{MessageId: "anyAssignment", Line: 5},
				{MessageId: "unsafeArraySpread", Line: 6},
			},
		},
		// Nested object literal containing any must still report on the inner property.
		{
			Code: `
declare const anyVal: any;
const x = { a: { b: anyVal } };
let y: { a: { b: number } };
y = { a: { b: anyVal } };
class C {
  z = { a: { b: anyVal } };
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment", Line: 3, Column: 18, EndColumn: 27},
				{MessageId: "anyAssignment", Line: 5, Column: 12, EndColumn: 21},
				{MessageId: "anyAssignment", Line: 7, Column: 14, EndColumn: 23},
			},
		},
		// Object literal containing SpreadAssignment must never be skipped.
		{
			Code: `
declare const anyVal: any;
const a = { a: { ...anyVal } };
const b = { ...anyVal };
let c: { x: number };
c = { ...anyVal };
class D {
  p = { ...anyVal };
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment", Line: 3, Column: 13, EndColumn: 29},
				{MessageId: "anyAssignment", Line: 4, Column: 7, EndColumn: 24},
				{MessageId: "anyAssignment", Line: 6, Column: 1, EndColumn: 18},
				{MessageId: "anyAssignment", Line: 8, Column: 3, EndColumn: 21},
			},
		},
		// ArrayLiteralExpression containing any produces Array<any> (TypeReference) and must never be skipped.
		{
			Code: `
declare const anyVal: any;
const a: { list: string[] } = { list: [anyVal] };
const b: string[] = [anyVal];
let c: string[];
c = [anyVal];
class D {
  list: string[] = [anyVal];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeAssignment", Line: 3},
				{MessageId: "unsafeAssignment", Line: 4},
				{MessageId: "unsafeAssignment", Line: 6},
				{MessageId: "unsafeAssignment", Line: 8},
			},
		},
		// JSX attributes with nested object any, spread any, or array spread any.
		{
			Code: `
type Props = {
  nested: { b: number };
  spread: { b: number };
  items: string[];
};
declare function Foo(props: Props): never;
declare const anyVal: any;
<Foo nested={{ b: anyVal }} spread={{ ...anyVal }} items={[...anyVal]} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment", Line: 9},
				{MessageId: "anyAssignment", Line: 9},
				{MessageId: "unsafeArraySpread", Line: 9},
			},
		},
		// Shadowed undefined is an Identifier and must never be skipped.
		{
			Code: `
function f(undefined: any) {
  const a = undefined;
  const obj = { a: undefined };
  return [a, obj];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment", Line: 3},
				{MessageId: "anyAssignment", Line: 4},
			},
		},
		// Under strictNullChecks: false, mutable variables initialized with null/void 0 widen to any on read.
		{
			Code: `
let a = null;
const b = a;
let c = void 0;
const d = c;
      `,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment", Line: 3},
				{MessageId: "anyAssignment", Line: 5},
			},
		},
		// Destructuring from non-spread object literals must still inspect senderType.
		{
			Code: `
declare const anyVal: any;
const { a } = { a: anyVal };
let b: number;
({ b } = { b: anyVal });
function f({ c } = { c: anyVal }) {
  return c;
}
const { d: { e } = { e: anyVal } } = {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeObjectPattern", Line: 3},
				{MessageId: "anyAssignment", Line: 3},
				{MessageId: "unsafeObjectPattern", Line: 5},
				{MessageId: "anyAssignment", Line: 5},
				{MessageId: "unsafeObjectPattern", Line: 6},
				{MessageId: "anyAssignment", Line: 6},
				{MessageId: "unsafeObjectPattern", Line: 9},
				{MessageId: "anyAssignment", Line: 9},
			},
		},
		// Inside a `with` statement, tsgo returns errorType for every expression.
		{
			Code: `
with ({}) {
  const x = 'foo';
  const obj = { a: 1 };
}
      `,
			FileName: "with-stmt.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment", Line: 3},
				{MessageId: "anyAssignment", Line: 4},
				{MessageId: "anyAssignment", Line: 4},
			},
		},
	})
}
