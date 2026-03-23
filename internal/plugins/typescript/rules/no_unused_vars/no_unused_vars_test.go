package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// --- basic usage ---
		{Code: `const foo = 5; console.log(foo);`},
		// shorthand property: { stats } counts as usage of stats
		{Code: `function test(stats: string) { console.log({ stats }); } test("ok");`},
		{Code: `function foo() {} foo();`},
		{Code: `function foo(bar) { console.log(bar); } foo(1);`},
		{Code: `try {} catch (e) { console.log(e); }`},
		{Code: `const { foo, ...rest } = { foo: 1, bar: 2 }; console.log(rest);`, Options: map[string]interface{}{"ignoreRestSiblings": true}},
		{Code: `const _foo = 1;`, Options: map[string]interface{}{"varsIgnorePattern": "^_"}},
		{Code: `function foo(bar) {} foo(1);`, Options: map[string]interface{}{"args": "none"}},
		{Code: `try {} catch (e) {}`, Options: map[string]interface{}{"caughtErrors": "none"}},
		{Code: `export const foo = 1;`},
		// type-annotated variable that IS used
		{Code: `const bar: number = 1; console.log(bar);`},
		// argsIgnorePattern
		{Code: `function foo(_bar) {} foo(1);`, Options: map[string]interface{}{"argsIgnorePattern": "^_"}},
		// caughtErrorsIgnorePattern
		{Code: `try {} catch (_err) {}`, Options: map[string]interface{}{"caughtErrorsIgnorePattern": "^_"}},

		// --- declare function params ---
		{Code: `declare function doSomething(options: { a: string }): void; export { doSomething };`},
		{Code: `declare function foo(): void; foo();`},
		{Code: `
declare function getNormalizedConfig(): string;
declare function getNormalizedConfig(options: { env: string }): string;
getNormalizedConfig();
`},
		{Code: `
declare function getNormalizedConfig(): string;
declare function getNormalizedConfig(options: { env: string }): string;
export { getNormalizedConfig };
`},
		{Code: `declare function withRest(...args: any[]): void; export { withRest };`},
		{Code: `declare function multi(a: string, b: number): void; export { multi };`},
		{Code: `export declare function exportDeclare(x: number): void;`},
		{Code: `
declare function genericFunc<T>(input: T): T;
export { genericFunc };
`},

		// --- function overloads ---
		{Code: `
export function foo(a: number): number;
export function foo(a: string): string;
export function foo(a: number | string): number | string {
  return a;
}
`},
		{Code: `
function foo(): void;
function foo(): void {}
foo();
`},

		// --- declare namespace ---
		{Code: `
declare namespace MyNS {
  function nsFunc(param: string): void;
  var nsVar: string;
}
console.log(MyNS);
`},
		{Code: `export namespace ExportedNS { export const x = 1; }`},
		{Code: `
declare module 'some-module' {
  function moduleFunc(arg: string): void;
}
`},

		// --- constructor overloads ---
		{Code: `
export class MyClass {
  constructor(a: number);
  constructor(a: string);
  constructor(a: number | string) { console.log(a); }
}
`},

		// --- class/interface/type/enum: used ---
		{Code: `class Foo {} new Foo();`},
		{Code: `export class Foo {}`},
		{Code: `export interface Bar { x: number; }`},
		{Code: `export type Str = string;`},
		{Code: `enum Color { Red, Blue } console.log(Color.Red);`},
		{Code: `export enum Color { Red, Blue }`},

		// --- abstract/method/interface without body params ---
		{Code: `
abstract class AbstractBase {
  abstract doSomething(input: string): void;
}
export { AbstractBase };
`},
		{Code: `
class MyClass {
  method(a: number): number;
  method(a: string): string;
  method(a: number | string): number | string {
    return a;
  }
}
export { MyClass };
`},
		{Code: `
export interface IProcessor {
  process(input: string, options: { debug: boolean }): void;
}
`},

		// --- function type literal params (type-level, never reported) ---
		{Code: `
export interface Hot {
  on: <Data = any>(event: string, cb: (data: Data) => void) => void;
}
`},
		// call signature params
		{Code: `
export interface Callable {
  (x: number, y: string): boolean;
}
`},
		// construct signature params
		{Code: `
export interface Constructable {
  new (name: string): object;
}
`},
		// function type in type alias
		{Code: `
export type Handler = (event: string, data: unknown) => void;
`},
		// index signature param
		{Code: `export interface Dict { [key: string]: unknown; }`},
		// declare global (global scope augmentation, never reported)
		{Code: `declare global { const BUILD_HASH: string; }`},
		// Note: namespace augmentation (e.g., `declare namespace NodeJS`) is also
		// skipped when the symbol has declarations in other files, but this requires
		// @types/node or similar to be present, which can't be tested in fixtures.

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

		// --- destructuring: used ---
		{Code: `const { a } = { a: 1 }; console.log(a);`},
		{Code: `const [p] = [1]; console.log(p);`},
		// import: used
		{Code: `import type { Foo } from "./foo"; const bar: Foo = {} as any; console.log(bar);`},
		// namespace import: used
		{Code: `import * as path from "path"; console.log(path.join("a", "b"));`},
		// import equals: used
		{Code: `import path = require("path"); console.log(path.join("a", "b"));`},
		// parameter destructuring: used
		{Code: `function foo({ a }: { a: number }) { console.log(a); } foo({ a: 1 });`},
		// parameter destructuring: argsIgnorePattern applies
		{Code: `function foo({ _a, b }: { _a: number; b: number }) { console.log(b); } foo({ _a: 1, b: 2 });`, Options: map[string]interface{}{"argsIgnorePattern": "^_"}},
		// parameter destructuring: args "none" skips all
		{Code: `function foo({ a }: { a: number }) {} foo({ a: 1 });`, Options: map[string]interface{}{"args": "none"}},
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
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 5}},
		},
		// type-annotated but unused variable should still be reported
		{
			Code:   `const bar: number = 1;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// usedOnlyAsType
		{
			Code:    `const foo = 1; type Bar = typeof foo; export type { Bar };`,
			Options: map[string]interface{}{"vars": "all"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedOnlyAsType", Line: 1, Column: 7}},
		},
		// varsIgnorePattern no match
		{
			Code:    `const foo = 1;`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// reportUsedIgnorePattern
		{
			Code:    `const _foo = 1; console.log(_foo);`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 1, Column: 7}},
		},
		// reportUsedIgnorePattern applies to argsIgnorePattern too
		{
			Code:    `function foo(_x: number) { return _x; } foo(1);`,
			Options: map[string]interface{}{"argsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 1, Column: 14}},
		},
		// reportUsedIgnorePattern applies to caughtErrorsIgnorePattern too
		{
			Code:    `try { throw 1; } catch (_e) { console.log(_e); }`,
			Options: map[string]interface{}{"caughtErrorsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 1, Column: 25}},
		},
		// varsIgnorePattern should NOT apply to params
		{
			Code:    `function foo(_x: number) {} foo(1);`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 14}},
		},
		// varsIgnorePattern should NOT apply to catch
		{
			Code:    `try {} catch (_e) {}`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		// argsIgnorePattern no match
		{
			Code:    `export function foo(bar: number) {}`,
			Options: map[string]interface{}{"argsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 21}},
		},
		// caughtErrorsIgnorePattern no match
		{
			Code:    `try {} catch (err) {}`,
			Options: map[string]interface{}{"caughtErrorsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		// --- declare function ---
		{
			Code:   `declare function unusedFunc(): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 18}},
		},
		{
			Code: `
declare function unusedOverload(): void;
declare function unusedOverload(x: number): void;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2, Column: 18}},
		},
		{
			Code: `
declare function typedFunc(): void;
type FuncType = typeof typedFunc;
export type { FuncType };
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedOnlyAsType", Line: 2, Column: 18}},
		},
		// unused declare namespace (with members)
		{
			Code:   `declare namespace UnusedNS { export function inner(): void; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 19}},
		},
		// unused empty declare namespace
		{
			Code:   `declare namespace Rspack {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 19}},
		},
		// unused empty namespace (non-declare)
		{
			Code:   `namespace Rspack2 {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 11}},
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
		// --- destructuring: unused element ---
		{
			Code:   `const { a, b } = { a: 1, b: 2 }; console.log(a);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 12}},
		},
		{
			Code:   `const [p, q] = [1, 2]; console.log(p);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 11}},
		},
		// --- import: unused ---
		{
			Code:   `import { Foo } from "./foo";`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 10}},
		},
		// namespace import: unused
		{
			Code:   `import * as ns from "./foo";`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 13}},
		},
		// import equals: unused
		{
			Code:   `import path = require("path");`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 8}},
		},
		// parameter destructuring: unused element
		{
			Code: `
function foo({ a, b }: { a: number; b: string }) { console.log(a); }
foo({ a: 1, b: "x" });
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2, Column: 19}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}
