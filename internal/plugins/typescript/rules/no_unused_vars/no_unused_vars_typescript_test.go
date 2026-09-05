package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
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
		// computed property names are value references, even inside type declarations
		{Code: `
declare const registeredServiceBrand: unique symbol;

export interface RegisteredService {
  [registeredServiceBrand]: string;
}
`},
		{Code: `
declare const brand: unique symbol;
export type Branded = { [brand]: string };
`},
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
		// A used parameter property establishes the last used parameter.
		{Code: `
export class Foo {
  constructor(value: string, private readonly name: string) {
    console.log(this.name);
  }
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
		// Type predicate parameter references are type-only, matching typescript-eslint.
		{
			Code:   `export function isString(value: unknown): value is string { return true; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedOnlyAsType", Line: 1, Column: 26}},
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
		// --- used infer type and mapped-type bindings are not reported ---
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

// TestNoUnusedVarsExportedDirective covers the `/* exported */` comment, which
// marks a global-scope binding used for a separately loaded file's sake. The
// TypeScript scope manager puts type-only declarations in the same global scope
// as value ones, so the directive reaches every declaration form a script-mode
// file can bind at its top level.
func TestNoUnusedVarsExportedDirective(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		{Code: `/* exported publicValue */ var publicValue = 1;`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{Code: `/* exported publicFn */ function publicFn() {}`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{Code: `/* exported PublicClass */ class PublicClass {}`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{Code: `/* exported PublicInterface */ interface PublicInterface { a: string }`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{Code: `/* exported PublicType */ type PublicType = string;`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{Code: `/* exported PublicEnum */ enum PublicEnum { A }`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{Code: `/* exported PublicNS */ namespace PublicNS { export const a = 1; }`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{Code: `/* exported ambient */ declare var ambient: number;`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{Code: `/* exported external */ import external = require("external");`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		// A `var` reaches the global scope from inside a block.
		{Code: `/* exported hoisted */ { var hoisted = 1; }`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// A parameter, a block binding, and anything at all in a module are out
		// of the global scope the directive resolves against.
		{
			Code: `/* exported param */ function outer(param: number) {}
outer(1);`,
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 37}},
		},
		{
			Code:            `/* exported blockScoped */ { let blockScoped = 1; }`,
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 34}},
		},
		{
			Code: `/* exported moduleValue */ var moduleValue = 1;
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 32}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}

func TestNoUnusedVarsHeritage(t *testing.T) {
	// These member references are syntactically valid even when the checker
	// rejects the inherited type. ESLint still resolves a type-capable root.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, []rule_tester.ValidTestCase{
		{Code: `class Cls { static method() {} } export interface I extends Cls.method {}`},
		{Code: `class Cls { static method() {} } export class C implements Cls.method {}`},
		{Code: `interface Obj { method(): void } export interface I extends Obj.method {}`},
		{Code: `type Obj = { method(): void }; export class C implements Obj.method {}`},
		{Code: `class Cls {} export interface I extends Cls.nested.member {}`},
		{Code: `namespace NS { export interface Base {} } export interface I extends NS.Base {}`},
		{Code: `class Cls {} namespace Cls { export interface Base {} } export interface I extends Cls.Base {}`},
		{Code: `interface Base<T> { value: T } class Cls {} export interface I extends Base<Cls> {}`},
		{Code: `declare const obj: { Base: new () => object }; export class C extends obj.Base {}`},
		{Code: `import type * as NS from 'missing'; export interface I extends NS.Base {}`},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `declare const obj: { m(): void }; export interface I extends obj.m {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Message: "'obj' is defined but never used."}},
		},
		{
			Code:   `namespace NS { export interface Base {} } export function f() { let NS = 0; class C implements NS.Base {} return C; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Message: "'NS' is assigned a value but never used."}},
		},
		{
			Code:   `namespace NS { export interface Base {} } export function f() { class NS {} class C implements NS.Base {} return C; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Message: "'NS' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 13}},
		},
		{
			Code:   `const obj = { m() {} }; export type T = typeof obj.m;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedOnlyAsType", Message: "'obj' is assigned a value but only used as a type."}},
		},
		{
			Code:    `interface _Cls {} export interface I extends _Cls.member {}`,
			Options: map[string]any{"varsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar"}},
		},
	})
}
