package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsCore(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// --- basic usage ---
		{Code: `const foo = 5; console.log(foo);`},
		// shorthand property: { stats } counts as usage of stats
		{Code: `function test(stats: string) { console.log({ stats }); } test("ok");`},
		{Code: `function foo() {} foo();`},
		{Code: `function foo(bar) { console.log(bar); } foo(1);`},
		{Code: `try {} catch (e) { console.log(e); }`},
		{Code: `export const foo = 1;`},
		// type-annotated variable that IS used
		{Code: `const bar: number = 1; console.log(bar);`},

		// --- class/interface/type/enum: used ---
		{Code: `class Foo {} new Foo();`},
		{Code: `export class Foo {}`},
		{Code: `export interface Bar { x: number; }`},
		{Code: `export type Str = string;`},
		{Code: `enum Color { Red, Blue } console.log(Color.Red);`},
		{Code: `export enum Color { Red, Blue }`},

		// --- export class/function: params used ---
		{Code: `export function bar(x: number) { return x; }`},
		{Code: `
export class Baz {
  value: number;
  constructor(a: number) { this.value = a; }
}
`},

		// --- scope ---
		{Code: `
const x = 1;
function foo(x: number) { return x; }
console.log(x);
foo(2);
`},

		// --- basic destructuring: used ---
		{Code: `const { a } = { a: 1 }; console.log(a);`},
		{Code: `const [p] = [1]; console.log(p);`},
		// parameter destructuring: used
		{Code: `function foo({ a }: { a: number }) { console.log(a); } foo({ a: 1 });`},
		// local function re-exported (not an import, still valid)
		{Code: `function foo() {} export { foo };`},

		// --- IIFE patterns ---
		{Code: `(function() { return 1; })();`},
		{Code: `(function foo() { return foo(); })();`},
		// named function expression: name used only inside body (self-reference)
		{Code: `const foo = function bar() { return bar(); }; foo();`},

		// --- self-referencing / recursive ---
		{Code: `function foo() { return foo(); } foo();`},
		{Code: `function foo(n: number): number { return n <= 1 ? 1 : n * foo(n - 1); } foo(5);`},
		// mutual recursion
		{Code: `
function isEven(n: number): boolean { return n === 0 ? true : isOdd(n - 1); }
function isOdd(n: number): boolean { return n === 0 ? false : isEven(n - 1); }
console.log(isEven(4));
`},

		// --- labeled statement ---
		{Code: `
var foo = 5;
label: while (true) {
  console.log(foo);
  break label;
}
`},

		// --- Function.bind / Function.toString ---
		{Code: `declare function myFunc(x: any): void; myFunc(function foo() {}.bind(undefined));`},
		{Code: `declare function myFunc(x: any): void; myFunc(function foo() {}.toString());`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// --- basic unused ---
		{
			Code:   `const foo = 5;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// unused function (no params) → report function name
		{
			Code:   `function foo() {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 10}},
		},
		// unused function WITH params → report function name AND params
		{
			Code: `function unused(a: number, b: string) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 1, Column: 10},
				{MessageId: "unusedVar", Line: 1, Column: 17},
				{MessageId: "unusedVar", Line: 1, Column: 28},
			},
		},
		// unused catch
		{
			Code:   `try {} catch (e) {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		// assignment without usage
		{
			Code:   `let foo = 5; foo = 10;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 14}},
		},
		// type-annotated but unused variable should still be reported
		{
			Code:   `const bar: number = 1;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// --- unused class/interface/type/enum ---
		{
			Code:   `class UnusedClass {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		{
			Code:   `interface UnusedInterface { x: number; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 11}},
		},
		{
			Code:   `type UnusedType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 6}},
		},
		{
			Code:   `enum UnusedEnum { A, B }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 6}},
		},
		// export function: unused param
		{
			Code:   `export function bar(x: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 21}},
		},
		// export class: unused constructor param
		{
			Code: `
export class UnusedCtorParam {
  constructor(a: number) {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 15}},
		},
		// scope: same-name param
		{
			Code: `
declare function other(x: number): void;
export { other };
export function bar(x: number) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 4, Column: 21}},
		},
		// after-used: only report params after last used
		{
			Code: `
export function foo(used: number, unused1: string, unused2: boolean) {
  return used;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 35},
				{MessageId: "unusedVar", Line: 2, Column: 52},
			},
		},
		// after-used: middle param used
		{
			Code: `
export function qux(a: number, b: string, c: boolean) {
  console.log(b);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 43},
			},
		},
		// after-used: all unused → report all
		{
			Code:   `export function bar(a: number, b: string, c: boolean) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 1, Column: 21},
				{MessageId: "unusedVar", Line: 1, Column: 32},
				{MessageId: "unusedVar", Line: 1, Column: 43},
			},
		},
		// args: "all" → report ALL unused regardless of position
		{
			Code: `
export function qux(a: number, b: string, c: boolean) {
  console.log(b);
}
`,
			Options: map[string]interface{}{"args": "all"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 21},
				{MessageId: "unusedVar", Line: 2, Column: 43},
			},
		},
		// after-used: default value param acts as boundary, only it is reported
		{
			Code:   `export const fn = (_a: string, _b: number, _c = {}) => {};`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 44}},
		},
		// --- basic destructuring: unused element ---
		{
			Code:   `const { a, b } = { a: 1, b: 2 }; console.log(a);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 12}},
		},
		{
			Code:   `const [p, q] = [1, 2]; console.log(p);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 11}},
		},
		// parameter destructuring: unused element
		{
			Code: `
function foo({ a, b }: { a: number; b: string }) { console.log(a); }
foo({ a: 1, b: "x" });
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2, Column: 19}},
		},
		// usedOnlyAsType
		{
			Code:    `const foo = 1; type Bar = typeof foo; export type { Bar };`,
			Options: map[string]interface{}{"vars": "all"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedOnlyAsType", Line: 1, Column: 7}},
		},
		// --- self-referencing function: unused externally ---
		{
			Code: `
function foox() {
  return foox();
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2, Column: 10}},
		},
		// nested self-referencing
		{
			Code: `
(function () {
  function foox() {
    if (true) {
      return foox();
    }
  }
})();
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 12}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}
