package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsTypeScript(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
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
		// declare global with nested namespace and interface
		{Code: `
declare global {
  namespace jest {
    interface Matchers<R> {
      toBeSeven: () => R;
    }
  }
}
`},
		// TypeScript this parameter (type annotation only, not a real param)
		{Code: `export default function webpackLoader(this: any) {}`},
		// Constructor parameter property (promoted to class field)
		{Code: `
export class Foo {
  constructor(private readonly name: string) {}
}
`},

		// --- decorator argument usage ---
		{Code: `
declare function Component(opts: any): any;
declare class Vue {}
declare const HelloWorld: any;

@Component({
  components: {
    HelloWorld,
  },
})
export default class App extends Vue {}
`},

		// --- setter parameter: syntactically required, never reported ---
		{Code: `
export const obj = {
  set foo(a: number) {}
};
`},
		// setter in class
		{Code: `
export class Foo {
  set bar(value: string) {}
}
`},
		// setter with args: 'all' — setter param is syntactically required, still not reported
		{Code: `
export class Foo {
  set bar(value: string) {}
}
`, Options: map[string]interface{}{"args": "all"}},

		// --- conditional types with infer (type-level, never reported) ---
		{Code: `export type Test<U> = U extends (k: infer I) => void ? I : never;`},
		{Code: `export type Test<U> = U extends { [k: string]: infer I } ? I : never;`},

		// --- enum member access ---
		{Code: `
enum FormFieldIds {
  PHONE = 'phone',
  EMAIL = 'email',
}
export interface IFoo {
  fieldName: FormFieldIds.EMAIL;
}
`},
		// enum self-reference
		{Code: `
export enum Foo {
  A = 1,
  B = Foo.A,
}
`},

		// --- namespace: used externally ---
		{Code: `namespace Foo { export const Bar = 1; } console.log(Foo.Bar);`},

		// --- mapped types ---
		{Code: `
type Foo = 'a' | 'b';
type Bar = number;
export const map: { [name in Foo]: Bar } = { a: 1, b: 2 };
`},

		// --- template literal types ---
		{Code: `
type Color = 'red' | 'blue';
type Quantity = 'one' | 'two';
export type SeussFish = ` + "`${Quantity | Color} fish`" + `;
`},

		// --- export import (namespace re-export) ---
		{Code: `
namespace FooNS {
  export const fooVal = 1;
}
export namespace BarNS {
  export import TheFoo = FooNS;
}
`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
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
		// --- namespace self-reference: only used inside own body ---
		{
			Code: `
namespace Foo {
  export const Bar = 1;
  console.log(Foo.Bar);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2, Column: 11}},
		},
		// nested namespace error count
		{
			Code: `
export namespace Foo {
  namespace Bar {
    namespace Baz {
      namespace Bam {
        const x = 1;
      }
    }
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 13},
				{MessageId: "unusedVar", Line: 4, Column: 15},
				{MessageId: "unusedVar", Line: 5, Column: 17},
				{MessageId: "unusedVar", Line: 6, Column: 15},
			},
		},
		// declare module unused types
		{
			Code: `
declare module 'foo' {
  type Test = any;
  const x = 1;
  export = x;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 8}},
		},
		// merged interface line position
		{
			Code: `
interface Foo {
  a: string;
}
interface Foo {
  b: Foo;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2, Column: 11}},
		},
		// --- typeof property access: usedOnlyAsType ---
		{
			Code: `
const fooObj = {
  bar: { baz: 123 },
};
export type BarType = typeof fooObj.bar;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedOnlyAsType", Line: 2, Column: 7}},
		},
		// typeof with index access type
		{
			Code: `
const fooObj2 = { x: 1 };
export type X = (typeof fooObj2)['x'];
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedOnlyAsType", Line: 2, Column: 7}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}

func TestNoUnusedVarsTypeParameters(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// --- Type parameter USED in function return type ---
		{Code: `export function fn<T>(x: T): T { return x; }`},
		// --- Type parameter USED in parameter type ---
		{Code: `export function fn<T>(x: T): void { console.log(x); }`},
		// --- Type parameter USED in interface body ---
		{Code: `export interface I<T> { value: T; }`},
		// --- Type parameter USED in type alias body ---
		{Code: `export type A<T> = T[];`},
		// --- Type parameter USED in class member ---
		{Code: `export class C<T> { x!: T; }`},
		// --- Type parameter USED by another TP's constraint ---
		{Code: `export function fn<T, U extends T>(x: T, y: U): void { console.log(x, y); }`},
		// --- Type parameter USED by another TP's default ---
		{Code: `export interface I<T, U = T> { x?: U; }`},
		// --- Type parameter USED in conditional type ---
		{Code: `export type IsString<T> = T extends string ? true : false;`},
		// --- Type parameter USED in template literal type ---
		{Code: `export type Greeting<T extends string> = ` + "`Hello ${T}`" + `;`},
		// --- Type parameter USED in mapped type ---
		{Code: `export type MyRecord<K extends string> = { [P in K]: number };`},
		// --- infer type is not a declaration — P in mapped type not reported ---
		{Code: `export type ElementOf<T> = T extends (infer U)[] ? U : never;`},
		// --- arrow function type parameter ---
		{Code: `export const fn = <T,>(x: T): T => x;`},
		// --- method type parameter ---
		{Code: `
export class C {
  fn<T>(x: T): T { return x; }
}
`},
		// --- varsIgnorePattern applies to type parameters ---
		{
			Code:    `export interface I<_T> {}`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_"},
		},
		// --- call signature type parameter used ---
		{Code: `
export interface Factory {
  <T>(x: T): T;
}
`},
		// --- construct signature type parameter used ---
		{Code: `
export interface Constructable {
  new <T>(x: T): T;
}
`},
		// --- overloaded function with type parameter ---
		{Code: `
export function foo<T>(a: number): T;
export function foo<T>(a: string): T;
export function foo<T>(a: number | string): T { return a as unknown as T; }
`},
		// --- declare function type parameter (ambient) ---
		{Code: `declare function fn<T>(x: T): T; export { fn };`},
		// --- declare module type parameter skipped ---
		{Code: `
declare module 'foo' {
  function bar<T>(x: T): T;
}
`},
		// --- Type parameter used in typeof ---
		{Code: `
export function foo<T>(value: T): T { return value; }
export type Foo<T> = typeof foo<T>;
`},
		// --- Type parameter used in spread parameter ---
		{Code: `export type Fn<A extends unknown[]> = (...a: A) => unknown;`},
		// --- Type parameter used in nested generic ---
		{Code: `export type Wrapper<T> = Promise<Array<T>>;`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// --- Unused type parameter on interface ---
		{
			Code:   `export interface I<T> { x?: number; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 20}},
		},
		// --- Unused type parameter with default ---
		{
			Code:   `export interface I<T = unknown> { x?: number; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 20}},
		},
		// --- Unused type parameter with constraint ---
		{
			Code:   `export interface I<T extends string> { x?: number; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 20}},
		},
		// --- Unused type parameter on type alias ---
		{
			Code:   `export type A<T> = string;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		// --- Unused type parameter on type alias with default ---
		{
			Code:   `export type A<T = unknown> = string;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		// --- Unused type parameter on function ---
		{
			Code:   `export function fn<T>(): void {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 20}},
		},
		// --- Unused type parameter on class ---
		{
			Code:   `export class C<T> {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 16}},
		},
		// --- Unused type parameter on class with default ---
		{
			Code:   `export class C<T = unknown> {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 16}},
		},
		// --- Multiple type params: only unused one reported ---
		{
			Code:   `export interface I<T, U> { x: T; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 23}},
		},
		// --- CrossRef: T used by U's constraint, but U itself unused ---
		{
			Code:   `export interface I<T, U extends T> {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 23}},
		},
		// --- Arrow function unused type parameter ---
		{
			Code:   `export const fn = <T,>(): void => {};`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 20}},
		},
		// --- Method unused type parameter ---
		{
			Code: `
export class C {
  fn<T>(): void {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 6}},
		},
		// --- Call signature unused type parameter ---
		{
			Code: `
export interface Factory {
  <T>(): void;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 4}},
		},
		// --- Construct signature unused type parameter ---
		{
			Code: `
export interface Constructable {
  new <T>(): void;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 8}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}

