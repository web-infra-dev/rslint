package no_shadow

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Cases follow the upstream ESLint test file
// (https://github.com/eslint/eslint/blob/main/tests/lib/rules/no-shadow.js),
// in roughly the same order so that a diff against that file is tractable.
// Framework-level concepts that rslint doesn't model
// (languageOptions.globals / env / parserOptions.globalReturn / script-vs-
// module sourceType distinction) are marked `Skip: true` with a comment
// pointing at the reason.
func TestNoShadowRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoShadowRule,

		[]rule_tester.ValidTestCase{
			// ---- Core JS: baseline (no shadow) ----
			{Code: `var a=3; function b(x) { a++; return x + a; }; setTimeout(function() { b(a); }, 0);`},

			// ---- Function-name initializer exception (function expression) ----
			{Code: `(function() { var doSomething = function doSomething() {}; doSomething() }())`},
			{Code: `(function() { var doSomething = foo || function doSomething() {}; doSomething() }())`},
			{Code: `(function() { var doSomething = function doSomething() {} || foo; doSomething() }())`},
			{Code: `(function() { var doSomething = foo && function doSomething() {}; doSomething() }())`},
			{Code: `(function() { var doSomething = foo ?? function doSomething() {}; doSomething() }())`},
			{Code: `(function() { var doSomething = foo || (bar || function doSomething() {}); doSomething() }())`},
			{Code: `(function() { var doSomething = foo || (bar && function doSomething() {}); doSomething() }())`},
			{Code: `(function() { var doSomething = foo ? function doSomething() {} : bar; doSomething() }())`},
			{Code: `(function() { var doSomething = foo ? bar: function doSomething() {}; doSomething() }())`},
			{Code: `(function() { var doSomething = foo ? bar: (baz || function doSomething() {}); doSomething() }())`},
			{Code: `(function() { var doSomething = (foo ? bar: function doSomething() {}) || baz; doSomething() }())`},
			{Code: `(function() { var { doSomething = function doSomething() {} } = obj; doSomething() }())`},
			{Code: `(function() { var { doSomething = function doSomething() {} || foo } = obj; doSomething() }())`},
			{Code: `(function() { var { doSomething = foo ? function doSomething() {} : bar } = obj; doSomething() }())`},
			{Code: `(function() { var { doSomething = foo ? bar : function doSomething() {} } = obj; doSomething() }())`},
			{Code: `(function() { var { doSomething = foo || (bar ? baz : (qux || function doSomething() {})) || quux } = obj; doSomething() }())`},
			{Code: `function foo(doSomething = function doSomething() {}) { doSomething(); }`},
			{Code: `function foo(doSomething = function doSomething() {} || foo) { doSomething(); }`},
			{Code: `function foo(doSomething = foo ? function doSomething() {} : bar) { doSomething(); }`},
			{Code: `function foo(doSomething = foo ? bar : function doSomething() {}) { doSomething(); }`},
			{Code: `function foo(doSomething = foo || (bar ? baz : (qux || function doSomething() {})) || quux) { doSomething(); }`},

			// ---- Miscellaneous ----
			{Code: "var arguments;\nfunction bar() { }"},
			{Code: `var a=3; var b = (x) => { a++; return x + a; }; setTimeout(() => { b(a); }, 0);`},

			// ---- Classes ----
			{Code: `class A {}`},
			{Code: `class A { constructor() { var a; } }`},
			// Function-name initializer exception (class expression)
			{Code: `(function() { var A = class A {}; })()`},
			{Code: `(function() { var A = foo || class A {}; })()`},
			{Code: `(function() { var A = class A {} || foo; })()`},
			{Code: `(function() { var A = foo && class A {} || foo; })()`},
			{Code: `(function() { var A = foo ?? class A {}; })()`},
			{Code: `(function() { var A = foo || (bar || class A {}); })()`},
			{Code: `(function() { var A = foo || (bar && class A {}); })()`},
			{Code: `(function() { var A = foo ? class A {} : bar; })()`},
			{Code: `(function() { var A = foo ? bar : class A {}; })()`},
			{Code: `(function() { var A = foo ? bar: (baz || class A {}); })()`},
			{Code: `(function() { var A = (foo ? bar: class A {}) || baz; })()`},
			{Code: `(function() { var { A = class A {} } = obj; }())`},
			{Code: `(function() { var { A = class A {} || foo } = obj; }())`},
			{Code: `(function() { var { A = foo ? class A {} : bar } = obj; }())`},
			{Code: `(function() { var { A = foo ? bar : class A {} } = obj; }())`},
			{Code: `(function() { var { A = foo || (bar ? baz : (qux || class A {})) || quux } = obj; }())`},
			{Code: `function foo(A = class A {}) { doSomething(); }`},
			{Code: `function foo(A = class A {} || foo) { doSomething(); }`},
			{Code: `function foo(A = foo ? class A {} : bar) { doSomething(); }`},
			{Code: `function foo(A = foo ? bar : class A {}) { doSomething(); }`},
			{Code: `function foo(A = foo || (bar ? baz : (qux || class A {})) || quux) { doSomething(); }`},

			// ---- Block shadowing a later var: not a shadow (var redeclares) ----
			{Code: `{ var a; } var a;`},

			// ---- hoist: default "functions" ----
			{Code: `{ let a; } let a;`},
			{Code: `{ let a; } var a;`},
			{Code: `{ const a = 0; } const a = 1;`},
			{Code: `{ const a = 0; } var a;`},

			// ---- hoist: "never" — outer declaration appears after inner, never report ----
			{Code: `{ let a; } let a;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `{ let a; } var a;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `{ let a; } function a() {}`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `{ const a = 0; } const a = 1;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `{ const a = 0; } var a;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `{ const a = 0; } function a() {}`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `function foo() { let a; } let a;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `function foo() { let a; } var a;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `function foo() { let a; } function a() {}`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `function foo() { var a; } let a;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `function foo() { var a; } var a;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `function foo() { var a; } function a() {}`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `function foo(a) { } let a;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `function foo(a) { } var a;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `function foo(a) { } function a() {}`, Options: map[string]interface{}{"hoist": "never"}},

			// ---- Implicit default hoist — after inner sibling already hidden in block ----
			{Code: `{ let a; } let a;`},
			{Code: `{ let a; } var a;`},
			{Code: `{ const a = 0; } const a = 1;`},
			{Code: `{ const a = 0; } var a;`},
			{Code: `function foo() { let a; } let a;`},
			{Code: `function foo() { let a; } var a;`},
			{Code: `function foo() { var a; } let a;`},
			{Code: `function foo() { var a; } var a;`},
			{Code: `function foo(a) { } let a;`},
			{Code: `function foo(a) { } var a;`},

			// ---- builtinGlobals off (default) ----
			{Code: `function foo() { var Object = 0; }`},
			// SKIP: `function foo() { var top = 0; }` with globals.browser — requires env.

			// Script-mode merging (VALID under script sourceType). Our implementation
			// always treats the file as a module, so `var Object = 0;` at top level
			// under `builtinGlobals: true` reports. Skip to preserve upstream intent.
			{Code: `var Object = 0;`, Options: map[string]interface{}{"builtinGlobals": true}, Skip: true},
			// SKIP: `var top = 0;` + browser globals — requires env.

			// ---- allow list ----
			{Code: `function foo(cb) { (function (cb) { cb(42); })(cb); }`, Options: map[string]interface{}{"allow": []interface{}{"cb"}}},

			// ---- Function-name initializer exception with arbitrary CallExpression
			// wrappers — ESLint's `outerScope === innerScope.upper` accepts any
			// non-scope-introducing wrapper, not just a fixed list of operators.
			{Code: `const a = wrap(function a() {});`},
			{Code: `const a = foo || wrap(function a() {});`},
			{Code: `const { a = wrap(function a() {}) } = obj;`},
			{Code: `const { a = foo || wrap(function a() {}) } = obj;`},
			{Code: `function foo(a = wrap(function a() {})) {}`},
			{Code: `function foo(a = foo || wrap(function a() {})) {}`},
			{Code: `const A = wrap(class A {});`},
			{Code: `const A = foo || wrap(class A {});`},
			{Code: `const { A = wrap(class A {}) } = obj;`},
			{Code: `const { A = foo || wrap(class A {}) } = obj;`},
			{Code: `function foo(A = wrap(class A {})) {}`},
			{Code: `function foo(A = foo || wrap(class A {})) {}`},
			// Sibling-init in same destructuring also exempted by the same rule.
			{Code: `const { a = foo, b = function a() {} } = {}`},
			{Code: `const { A = Foo, B = class A {} } = {}`},
			// FunctionExpression / ClassExpression at the test position of a
			// ternary in the initializer is also exempted (same scope/range).
			{Code: `var a = function a() {} ? foo : bar`},
			{Code: `var A = class A {} ? foo : bar`},
			// Wrap with side-effecting top-level `let` — still exempted.
			{Code: `let x = false; export const a = wrap(function a() { if (!x) { x = true; a(); } });`, Options: map[string]interface{}{"hoist": "all"}},

			// ---- Class fields / methods (not shadowing — different kinds) ----
			{Code: `class C { foo; foo() { let foo; } }`},

			// ---- Class static blocks ----
			{Code: `class C { static { var x; } static { var x; } }`},
			{Code: `class C { static { let x; } static { let x; } }`},
			{Code: `class C { static { var x; { var x; /* redeclaration */ } } }`},
			{Code: `class C { static { { var x; } { var x; /* redeclaration */ } } }`},
			{Code: `class C { static { { let x; } { let x; } } }`},

			// ---- ignoreOnInitialization (callback form) ----
			{Code: `const a = [].find(a => a)`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const a = [].find(function(a) { return a; })`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const [a = [].find(a => true)] = dummy`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const { a = [].find(a => true) } = dummy`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `function func(a = [].find(a => true)) {}`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `for (const a in [].find(a => true)) {}`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `for (const a of [].find(a => true)) {}`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const a = [].map(a => true).filter(a => a === 'b')`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const a = [].map(a => true).filter(a => a === 'b').find(a => a === 'c')`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const { a } = (({ a }) => ({ a }))();`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const person = people.find(item => {const person = item.name; return person === 'foo'})`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `var y = bar || foo(y => y);`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `var y = bar && foo(y => y);`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `var z = bar(foo(z => z));`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `var z = boo(bar(foo(z => z)));`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `var match = function (person) { return person.name === 'foo'; };` + "\n" + `const person = [].find(match);`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const a = foo(x || (a => {}))`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const { a = 1 } = foo(a => {})`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const person = {...people.find((person) => person.firstName.startsWith('s'))}`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const person = { firstName: people.filter((person) => person.firstName.startsWith('s')).map((person) => person.firstName)[0]}`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `() => { const y = foo(y => y); }`, Options: map[string]interface{}{"ignoreOnInitialization": true}},

			// ---- ignoreOnInitialization (IIFE form) ----
			{Code: `const x = (x => x)()`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `var y = bar || (y => y)();`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `var y = bar && (y => y)();`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `var x = (x => x)((y => y)());`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const { a = 1 } = (a => {})()`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `() => { const y = (y => y)(); }`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const [x = y => y] = [].map(y => y)`},

			// ========================================================
			// TypeScript-specific valid cases
			// ========================================================

			// ---- Function-type parameters (default ignoreFunctionTypeParameterNameValueShadow: true) ----
			{Code: `function foo<T = (arg: any) => any>(arg: T) {}`},
			{Code: `function foo<T = ([arg]: [any]) => any>(arg: T) {}`},
			{Code: `function foo<T = ({ args }: { args: any }) => any>(arg: T) {}`},
			{Code: `function foo<T = (...args: any[]) => any>(fn: T, args: any[]) {}`},
			{Code: `function foo<T extends (...args: any[]) => any>(fn: T, args: any[]) {}`},
			{Code: `function foo<T extends (...args: any[]) => any>(fn: T, ...args: any[]) {}`},
			{Code: `function foo<T extends ([args]: any[]) => any>(fn: T, args: any[]) {}`},
			{Code: `function foo<T extends ([...args]: any[]) => any>(fn: T, args: any[]) {}`},
			{Code: `function foo<T extends ({ args }: { args: any }) => any>(fn: T, args: any) {}`},
			{Code: `function foo<T extends (id: string, ...args: any[]) => any>(fn: T, ...args: any[]) {}`},
			{Code: `type Args = 1; function foo<T extends (Args: any) => void>(arg: T) {}`},

			// ---- Conditional types with infer T ----
			{Code: `export type ArrayInput<Func> = Func extends (arg0: Array<infer T>) => any ? T[] : Func extends (...args: infer T) => any ? T : never;`},

			// ---- Local Object vs global (builtinGlobals off) ----
			{Code: `function foo() { var Object = 0; }`},

			// ---- this params ----
			{Code: `function test(this: number) { function test2(this: number) {} }`},

			// ---- Declaration merging (value + namespace / value + interface) ----
			{Code: `class Foo { prop = 1; } namespace Foo { export const v = 2; }`},
			{Code: `function Foo() {} namespace Foo { export const v = 2; }`},
			{Code: `class Foo { prop = 1; } interface Foo { prop2: string }`},

			// ---- Module augmentation with type-only import of the same module ----
			{Code: `import type { Foo } from 'bar';` + "\n" + `declare module 'bar' { export interface Foo { x: string } }`},

			// ---- Type/value shadowing (default ignoreTypeValueShadow: true) ----
			{Code: `const x = 1; type x = string;`},
			{Code: `const x = 1; { type x = string; }`},
			// SKIP: `type Foo = 1;` + languageOptions.globals.Foo — requires globals.

			// ---- TS enum declaration (member order-independent initializer) ----
			{Code: `enum Direction { left = 'left', right = 'right' }`},

			// ---- ignoreFunctionTypeParameterNameValueShadow: true (default) — various TS signature positions ----
			{Code: `const test = 1; type Fn = (test: string) => typeof test;`},
			// SKIP: `type Fn = (Foo: string) => typeof Foo` with globals.Foo — requires globals.
			{Code: `const arg = 0; interface Test { (arg: string): typeof arg; }`},
			{Code: `const arg = 0; interface Test { p1(arg: string): typeof arg; }`},
			{Code: `const arg = 0; declare function test(arg: string): typeof arg;`},
			{Code: `const arg = 0; declare const test: (arg: string) => typeof arg;`},
			{Code: `const arg = 0; declare class Test { p1(arg: string): typeof arg; }`},
			{Code: `const arg = 0; declare const Test: { new (arg: string): typeof arg };`},
			{Code: `const arg = 0; type Bar = new (arg: number) => typeof arg;`},
			{Code: `const arg = 0; declare namespace Lib { function test(arg: string): typeof arg; }`},

			// ---- `declare global` is transparent ----
			{Code: `declare global { interface ArrayConstructor {} } export {};`, Options: map[string]interface{}{"builtinGlobals": true}},
			{Code: `declare global { const a: string; namespace Foo { const a: number; } } export {};`},
			{Code: `declare global { type A = 'foo'; namespace Foo { type A = 'bar'; } } export {};`, Options: map[string]interface{}{"ignoreTypeValueShadow": false}},
			// SKIP: `declare global { const foo; type Fn = (foo) => void }` under ignoreFunctionTypeParameterNameValueShadow: false — framework globals.

			// ---- Static method generic doesn't collide with class generic ----
			{Code: `export class Wrapper<Wrapped> { private constructor(private readonly wrapped: Wrapped) {} unwrap(): Wrapped { return this.wrapped; } static create<Wrapped>(wrapped: Wrapped) { return new Wrapper<Wrapped>(wrapped); } }`},
			{Code: `function makeA() { return class A<T> { constructor(public value: T) {} static make<T>(value: T) { return new A<T>(value); } }; }`},

			// ---- Real-code-discovered: any type-only specifier in an import treats the
			// whole declaration as a type import for shadow purposes (ESLint quirk). ----
			{Code: `import binding, { type AssetInfo } from 'm';
class Foo { static __from_binding(binding: any) { return binding; } }`},
			{Code: `import { foo, type Bar } from 'm';
function fn(foo: number) { return foo; }`},
			// (`import * as N, { type T } from 'm'` is not valid TS syntax, omitted)

			// ---- infer X in conditional types is type-level, doesn't leak to value scope ----
			{Code: `type X<T> = T extends infer U ? U : never;
const U = 1;`},
			{Code: `type X<T> = T extends string ? T : never;
const T = 1;`}, // T type param doesn't leak

			// ---- Object literal accessor (no shadow when distinct names) ----
			{Code: `const x = 1;
const o = { get y() { return 1; }, set y(v) {} };`},

			// ---- Class field with private name (#) — no shadow vs same plain name ----
			{Code: `class C { #priv = 1; m() { const priv = 2; return [this.#priv, priv]; } }`},

			// ---- Tuple labels don't introduce bindings ----
			{Code: `type Pair = [first: string, second: number];
const first = 1;`},

			// ---- Class field name same as class name — different namespace ----
			{Code: `class C6 { C6 = 1; }`},

			// ---- Method name same as outer var — methods are class-scoped ----
			{Code: `function g1() {}
class CG { g1() {} }`},

			// ---- Mapped type [K in keyof T] doesn't leak ----
			{Code: `type Map1<T> = { [K in keyof T]: K };
const K = 1;`},

			// ---- Multiple infer at same level (sibling, not nested) — not shadow ----
			{Code: `type X<T> = T extends { a: infer U } & { b: infer U } ? U : never;`},

			// ---- Default destructure with reference to outer of same name (shadow happens to be the inner) ----
			// `function f({ a = a }) {}` — `a` on right refers to outer. Inner a still shadows.
			// Treated as invalid but we want to not crash.

			// ---- Generic in arrow function does not shadow outer value ----
			{Code: `const T = 1; const arr = <T extends string>(x: T) => x;`},

			// ---- Import type + type alias — ignoreTypeValueShadow true ----
			{Code: `import type { foo } from './foo';` + "\n" + `type bar = number;` + "\n" + `function doThing(foo: number, bar: number) {}`},
			{Code: `import { type foo } from './foo';` + "\n" + `function doThing(foo: number) {}`},

			// ---- ignoreOnInitialization (TS flavor) ----
			{Code: `const a = [].find(a => a);`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const a = [].find(function (a) { return a; });`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const [a = [].find(a => true)] = dummy;`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const { a = [].find(a => true) } = dummy;`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `function func(a = [].find(a => true)) {}`, Options: map[string]interface{}{"ignoreOnInitialization": true}},

			// ---- hoist TS: "never" ----
			{Code: `type Foo<A> = 1; type A = 1;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `interface Foo<A> {} type A = 1;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `interface Foo<A> {} interface A {}`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `type Foo<A> = 1; interface A {}`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `{ type A = 1; } type A = 1;`, Options: map[string]interface{}{"hoist": "never"}},
			{Code: `{ interface Foo<A> {} } type A = 1;`, Options: map[string]interface{}{"hoist": "never"}},

			// ---- hoist TS: "functions" (default for type-only situations still suppresses) ----
			{Code: `type Foo<A> = 1; type A = 1;`, Options: map[string]interface{}{"hoist": "functions"}},
			{Code: `interface Foo<A> {} type A = 1;`, Options: map[string]interface{}{"hoist": "functions"}},
			{Code: `interface Foo<A> {} interface A {}`, Options: map[string]interface{}{"hoist": "functions"}},
			{Code: `type Foo<A> = 1; interface A {}`, Options: map[string]interface{}{"hoist": "functions"}},
			{Code: `{ type A = 1; } type A = 1;`, Options: map[string]interface{}{"hoist": "functions"}},
			{Code: `{ interface Foo<A> {} } type A = 1;`, Options: map[string]interface{}{"hoist": "functions"}},

			// ---- import-type + module augmentation of same module (valid forms) ----
			{Code: `import type { Foo } from 'bar';` + "\n" + `declare module 'bar' { export type Foo = string }`},
			{Code: `import type { Foo } from 'bar';` + "\n" + `declare module 'bar' { interface Foo { x: string } }`},
			{Code: `import { type Foo } from 'bar';` + "\n" + `declare module 'bar' { export type Foo = string }`},
			{Code: `import { type Foo } from 'bar';` + "\n" + `declare module 'bar' { export interface Foo { x: string } }`},
			{Code: `import { type Foo } from 'bar';` + "\n" + `declare module 'bar' { type Foo = string }`},
			{Code: `import { type Foo } from 'bar';` + "\n" + `declare module 'bar' { interface Foo { x: string } }`},

			// ---- .d.ts declare — under builtinGlobals with globals, valid ----
			// These rely on globals: languageOptions which we don't support. But
			// the `declare`-in-dts branch is independent of that and still applies,
			// so we also assert the subset that should be valid without globals:
			// ambient decls in .d.ts are always filtered regardless of shadowing.
			// (Full fixtures with a specific .d.ts filename are not trivial in the
			// unit tester; these rely on upstream fixture setup so we skip them.)

			// ==== Additional ESLint cases for full coverage ====

			{Code: `var match = function (person) { return person.name === 'foo'; };
const person = [].find(match);`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `
  const arg = 0;
  
  declare const Test: {
	new (arg: string): typeof arg;
  };
		`, Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": true}},
			{Code: `
		  declare global {
			const foo: string;
			type Fn = (foo: number) => void;
		  }
		  export {};
		`, Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false}},
			{Code: `
  import { type foo } from './foo';
  
  // 'foo' is already declared in the upper scope
  function doThing(foo: number) {}
		`, Options: map[string]interface{}{"ignoreTypeValueShadow": true}},
			{Code: `const a = [].map(a => true).filter(a => a === 'b');`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `
  const a = []
	.map(a => true)
	.filter(a => a === 'b')
	.find(a => a === 'c');
		`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `
  const person = people.find(item => {
	const person = item.name;
	return person === 'foo';
  });
		`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `
  var match = function (person) {
	return person.name === 'foo';
  };
  const person = [].find(match);
		`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const a = foo(x || (a => {}));`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const { a = 1 } = foo(a => {});`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const person = { ...people.find(person => person.firstName.startsWith('s')) };`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `
  const person = {
	firstName: people
	  .filter(person => person.firstName.startsWith('s'))
	  .map(person => person.firstName)[0],
  };
		`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `
  () => {
	const y = foo(y => y);
  };
		`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const x = (x => x)();`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `const { a = 1 } = (a => {})();`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `
  () => {
	const y = (y => y)();
  };
		`, Options: map[string]interface{}{"ignoreOnInitialization": true}},
			{Code: `var arguments;
function bar() { }`},
			{Code: `
  function test(this: Foo) {
	function test2(this: Bar) {}
  }
	  `},
			{Code: `
  class Foo {
	prop = 1;
  }
  interface Foo {
	prop2: string;
  }
	  `},
			{Code: `
  import type { Foo } from 'bar';
  
  declare module 'bar' {
	export interface Foo {
	  x: string;
	}
  }
	  `},
			{Code: `
  enum Direction {
	left = 'left',
	right = 'right',
  }
	  `},
			{Code: `const [x = y => y] = [].map(y => y);`},
			{Code: `
  import type { Foo } from 'bar';
  
  declare module 'bar' {
	export type Foo = string;
  }
		`},
			{Code: `
  import type { Foo } from 'bar';
  
  declare module 'bar' {
	interface Foo {
	  x: string;
	}
  }
		`},
			{Code: `
  import { type Foo } from 'bar';
  
  declare module 'bar' {
	export type Foo = string;
  }
		`},
			{Code: `
  import { type Foo } from 'bar';
  
  declare module 'bar' {
	export interface Foo {
	  x: string;
	}
  }
		`},
			{Code: `
  import { type Foo } from 'bar';
  
  declare module 'bar' {
	type Foo = string;
  }
		`},
			{Code: `
  import { type Foo } from 'bar';
  
  declare module 'bar' {
	interface Foo {
	  x: string;
	}
  }
		`},
		},

		[]rule_tester.InvalidTestCase{
			// ---- Basic function/arrow shadow (full line/col) ----
			{
				Code: `function a(x) { var b = function c() { var x = 'foo'; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Message: "'x' is already declared in the upper scope on line 1 column 12.", Line: 1, Column: 44},
				},
			},
			{
				Code: `var a = (x) => { var b = () => { var x = 'foo'; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Message: "'x' is already declared in the upper scope on line 1 column 10.", Line: 1, Column: 38},
				},
			},
			{
				Code: `function a(x) { var b = function () { var x = 'foo'; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Message: "'x' is already declared in the upper scope on line 1 column 12.", Line: 1, Column: 43},
				},
			},
			{
				Code: `var x = 1; function a(x) { return ++x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Message: "'x' is already declared in the upper scope on line 1 column 5.", Line: 1, Column: 23},
				},
			},
			{
				Code: `var a=3; function b() { var a=10; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Message: "'a' is already declared in the upper scope on line 1 column 5."},
				},
			},
			{
				Code: `var a=3; function b() { var a=10; }; setTimeout(function() { b(); }, 0);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Message: "'a' is already declared in the upper scope on line 1 column 5."},
				},
			},
			{
				Code: `var a=3; function b() { var a=10; var b=0; }; setTimeout(function() { b(); }, 0);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Message: "'a' is already declared in the upper scope on line 1 column 5."},
					{MessageId: "noShadow", Message: "'b' is already declared in the upper scope on line 1 column 19."},
				},
			},
			{
				Code: `var x = 1; { let x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code: `let x = 1; { const x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- hoist: default "functions" (outer function declared after inner) ----
			{Code: `{ let a; } function a() {}`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `{ const a = 0; } function a() {}`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo() { let a; } function a() {}`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo() { var a; } function a() {}`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo(a) { } function a() {}`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},

			// ---- hoist: "all" — all permutations ----
			{Code: `{ let a; } let a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `{ let a; } var a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `{ let a; } function a() {}`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `{ const a = 0; } const a = 1;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `{ const a = 0; } var a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `{ const a = 0; } function a() {}`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo() { let a; } let a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo() { let a; } var a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo() { let a; } function a() {}`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo() { var a; } let a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo() { var a; } var a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo() { var a; } function a() {}`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo(a) { } let a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo(a) { } var a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `function foo(a) { } function a() {}`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},

			// ---- Function/class expression self-name shadowing inner body ----
			{Code: `(function a() { function a(){} })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `(function a() { class a{} })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `(function a() { (function a(){}); })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `(function a() { (class a{}); })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `(function() { var a = function(a) {}; })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `(function() { var a = function() { function a() {} }; })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `(function() { var a = function() { class a{} }; })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `(function() { var a = function() { (function a() {}); }; })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `(function() { var a = function() { (class a{}); }; })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `(function() { var a = class { constructor() { class a {} } }; })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `class A { constructor() { var A; } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},

			// ---- Nested shadowing chain (multiple errors) ----
			{
				Code: `(function a() { function a(){ function a(){} } })()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 26},
					{MessageId: "noShadow", Line: 1, Column: 40},
				},
			},

			// ---- builtinGlobals ----
			{Code: `function foo() { var Object = 0; }`, Options: map[string]interface{}{"builtinGlobals": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadowGlobal", Message: "'Object' is already a global variable."}}},
			// SKIP: `function foo() { var top = 0; }` requires browser globals.
			{Code: `var Object = 0;`, Options: map[string]interface{}{"builtinGlobals": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadowGlobal"}}},
			// SKIP: `var top = 0;` browser globals.
			// SKIP: `var Object = 0;` with globalReturn — framework parserOptions.
			// SKIP: `var top = 0;` with globalReturn + browser — framework.
			{Code: `function foo(cb) { (function (cb) { cb(42); })(cb); }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 31}}},

			// ---- Class static blocks ----
			{Code: `class C { static { let a; { let a; } } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 33}}},
			{Code: `class C { static { var C; } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 24}}},
			{Code: `class C { static { let C; } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 24}}},
			{Code: `var a; class C { static { var a; } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 31}}},
			{Code: `class C { static { var a; } } var a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 24}}},
			{Code: `class C { static { let a; } } let a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 24}}},
			{Code: `class C { static { var a; } } let a;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 24}}},
			{Code: `class C { static { var a; class D { static { var a; } } } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 50}}},
			{Code: `class C { static { let a; class D { static { let a; } } } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 50}}},

			// ---- Hoist "all" inside IIFEs / arrows with param list ----
			{
				Code:    `let x = foo((x,y) => {});` + "\n" + `let y;`,
				Options: map[string]interface{}{"hoist": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    `let x = ((x,y) => {})();` + "\n" + `let y;`,
				Options: map[string]interface{}{"hoist": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
				},
			},

			// ---- ignoreOnInitialization: the function is inside the class, so shadow is still reported ----
			{
				Code:    `const a = fn(()=>{ class C { fn () { const a = 42; return a } } return new C() })`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 44},
				},
			},
			{
				Code:    `function a() {}` + "\n" + `foo(a => {});`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 2, Column: 5},
				},
			},
			{
				Code:    `const a = fn(()=>{ function C() { this.fn=function() { const a = 42; return a } } return new C() });`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 62},
				},
			},
			{
				Code:    `const x = foo(() => { const bar = () => { return x => {}; }; return bar; });`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 50},
				},
			},
			{
				Code:    `const x = foo(() => { return { bar(x) {} }; });`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 36},
				},
			},
			{
				Code:    `const x = () => { foo(x => x); }`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 23},
				},
			},
			{
				Code:    `const foo = () => { let x; bar(x => x); }`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 32},
				},
			},
			{
				Code:    `foo(() => { const x = x => x; });`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 23},
				},
			},
			{
				Code:    `const foo = (x) => { bar(x => {}) }`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 26},
				},
			},
			{
				Code:    `const a = (()=>{ class C { fn () { const a = 42; return a } } return new C() })()`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 42},
				},
			},
			{
				Code:    `const x = () => { (x => x)(); }`,
				Options: map[string]interface{}{"ignoreOnInitialization": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow", Line: 1, Column: 20},
				},
			},

			// SKIP: `(function Array() {})` + builtinGlobals — relies on env/sourceType=module.
			{Code: `(function Array() {})`, Options: map[string]interface{}{"builtinGlobals": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadowGlobal", Line: 1, Column: 11}}},
			{Code: `let a; { let b = (function a() {}) }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 28}}},
			{Code: `let a = foo; { let b = (function a() {}) }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow", Line: 1, Column: 34}}},

			// ========================================================
			// TypeScript-specific invalid cases
			// ========================================================

			{
				Code: "\n  type T = 1;\n  {\n\ttype T = 2;\n  }\n\t\t",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code: "\n  type T = 1;\n  function foo<T>(arg: T) {}\n\t\t",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code: "\n  function foo<T>() {\n\treturn function <T>() {};\n  }\n\t\t",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code: "\n  type T = string;\n  function foo<T extends (arg: any) => void>(arg: T) {}\n\t\t",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    "\n  const x = 1;\n  {\n\ttype x = string;\n  }\n\t\t",
				Options: map[string]interface{}{"ignoreTypeValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// SKIP: `type Foo = 1;` + globals.Foo — framework globals.

			{
				Code:    "\n  const test = 1;\n  type Fn = (test: string) => typeof test;\n\t\t",
				Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// SKIP: `type Fn = (Foo: string) => typeof Foo` + globals.Foo — framework globals.

			{
				Code:    "\n  const arg = 0;\n  interface Test {\n\t(arg: string): typeof arg;\n  }\n\t\t",
				Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    "\n  const arg = 0;\n  interface Test {\n\tp1(arg: string): typeof arg;\n  }\n\t\t",
				Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    "\n  const arg = 0;\n  declare function test(arg: string): typeof arg;\n\t\t",
				Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    "\n  const arg = 0;\n  declare const test: (arg: string) => typeof arg;\n\t\t",
				Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    "\n  const arg = 0;\n  declare class Test {\n\tp1(arg: string): typeof arg;\n  }\n\t\t",
				Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    "\n  const arg = 0;\n  declare const Test: {\n\tnew (arg: string): typeof arg;\n  };\n\t\t",
				Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    "\n  const arg = 0;\n  type Bar = new (arg: number) => typeof arg;\n\t\t",
				Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    "\n  const arg = 0;\n  declare namespace Lib {\n\tfunction test(arg: string): typeof arg;\n  }\n\t\t",
				Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Type imports with ignoreTypeValueShadow: false ----
			{
				Code:    "\nimport type { foo } from './foo';\nfunction doThing(foo: number) {}\n",
				Options: map[string]interface{}{"ignoreTypeValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    "\nimport { type foo } from './foo';\nfunction doThing(foo: number) {}\n",
				Options: map[string]interface{}{"ignoreTypeValueShadow": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code: "\nimport { foo } from './foo';\nfunction doThing(foo: number, bar: number) {}\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Module augmentation with interface shadowing ----
			{
				Code: "\ninterface Foo {}\ndeclare module 'bar' { export interface Foo { x: string } }\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code: "\nimport type { Foo } from 'bar';\ndeclare module 'baz' { export interface Foo { x: string } }\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code: "\nimport { type Foo } from 'bar';\ndeclare module 'baz' { export interface Foo { x: string } }\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- hoist: "all" with TS ----
			{
				Code:    "\n  let x = foo((x, y) => {});\n  let y;\n\t\t",
				Options: map[string]interface{}{"hoist": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
				},
			},
			{
				Code:    "\n  let x = foo((x, y) => {});\n  let y;\n\t\t",
				Options: map[string]interface{}{"hoist": "functions"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- hoist: "types" / "functions-and-types" / "all" — all type permutations ----
			{Code: "\n  type Foo<A> = 1;\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  interface Foo<A> {}\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  interface Foo<A> {}\n  interface A {}\n\t\t", Options: map[string]interface{}{"hoist": "types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  type Foo<A> = 1;\n  interface A {}\n\t\t", Options: map[string]interface{}{"hoist": "types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  {\n\ttype A = 1;\n  }\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  {\n\tinterface A {}\n  }\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  type Foo<A> = 1;\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  interface Foo<A> {}\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  interface Foo<A> {}\n  interface A {}\n\t\t", Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  type Foo<A> = 1;\n  interface A {}\n\t\t", Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  {\n\ttype A = 1;\n  }\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  {\n\tinterface A {}\n  }\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  type Foo<A> = 1;\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n\tif (true) {\n\t\tconst foo = 6;\n\t}\n\n\tfunction foo() { }\n\t\t", Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n\t// types\n\ttype Bar<Foo> = 1;\n\ttype Foo = 1;\n\n\t// functions\n\tif (true) {\n\t\tconst b = 6;\n\t}\n\n\tfunction b() { }\n\t\t", Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}, {MessageId: "noShadow"}}},
			{Code: "\n\t// types\n\ttype Bar<Foo> = 1;\n\ttype Foo = 1;\n\n\t// functions\n\tif (true) {\n\t\tconst b = 6;\n\t}\n\n\tfunction b() { }\n\t\t", Options: map[string]interface{}{"hoist": "types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  interface Foo<A> {}\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  interface Foo<A> {}\n  interface A {}\n\t\t", Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  type Foo<A> = 1;\n  interface A {}\n\t\t", Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  {\n\ttype A = 1;\n  }\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: "\n  {\n\tinterface A {}\n  }\n  type A = 1;\n\t\t", Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},

			// SKIP: TS cases relying on languageOptions.globals (args/has/Foo).

			// ---- Enum member shadowing outer const ----
			{
				Code: "\n\t\tconst A = 2;\n\t\tenum Test {\n\t\t\tA = 1,\n\t\t\tB = A,\n\t\t}\n\t",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- infer T shadowing infer T (nested in same conditional chain) ----
			{
				Code: `type X<T> = T extends (infer U) ? (U extends (infer U) ? U : never) : never;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Object literal getter/setter/method param shadow ----
			{
				Code: `const x = 1; const o = { foo(x: number) { return x; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code: `const x = 1; const o = { get foo() { const x = 2; return x; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code: `const x = 1; const o = { async foo(x: number) { return x; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			{
				Code: `const x = 1; const o = { *foo(x: number) { yield x; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- using declarations ----
			{
				Code: `using u = { [Symbol.dispose]() {} }; { using u = { [Symbol.dispose]() {} }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Class generic shadowed by instance method generic ----
			{
				Code: `class C<T> { static s<T>(): T { return null as any; } i<T>(): T { return null as any; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Object pattern rename shadow ----
			{
				Code: `const a = 1; function f({ x: a }: any) { return a; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Function expression name shadowed by its own parameter ----
			{
				Code: `const fn = function f(f: number) { return f; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Decorator: shadowing inside decorated method body ----
			{
				Code: `function dec(x: any) { return x; }
@dec class CD { method() { const dec = 1; return dec; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Computed property name with class member shadow ----
			{
				Code: `const k = 'x'; class C { foo() { const k = 1; return k; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Default parameter referencing outer var of same name (still reported) ----
			{
				Code: `const a = 1; function f({ a = a }: any) { return a; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Optional catch binding does NOT introduce binding (but declared var inside still checked) ----
			{
				Code: `const e = 1; try {} catch ({ message: e }) { return e; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Async generator with for-of ----
			{
				Code: `async function* g() { const v = 1; for await (const v of [Promise.resolve(1)]) yield v; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Multi-decl in single statement ----
			{
				Code: `const a = 1, b = 2; function f() { const a = 3, b = 4; return a + b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
				},
			},

			// ---- Switch case lexical scope shadow ----
			{
				Code: `function f() { const z = 1; switch (z) { case 1: { const z = 2; break; } case 2: { const z = 3; break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
				},
			},

			// ---- Namespace import param shadow ----
			{
				Code: `import * as Mods from 'fs'; function f(Mods: number) { return Mods; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Re-declared inside arrow returning class with the same name ----
			{
				Code: `const Z = 1; const make = () => class Z {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- typeof reference + parameter shadow ----
			{
				Code: `const t = 1; type T = typeof t; function f(t: number) { return t; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Default parameter with type reference ----
			{
				Code: `const opts = { x: 1 }; function f(opts: typeof opts = opts) { return opts; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Anonymous function expression with same name inside ----
			{
				Code: `const a = 1; const fn = function() { const a = 2; return a; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Nested class with shadowed name ----
			{
				Code: `const A = 1; class A_outer { x = class A {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Rest pattern shadow inner destructure ----
			{
				Code: `function f(...args: number[]) { function g({ args }: any) { return args; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Generator with for-of shadow ----
			{
				Code: `function* g() { const v = 1; for (const v of [1,2,3]) { yield v; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Real-code-discovered: function type inside parameter type
			// annotation carries its own generics that shadow outer generics ----
			{
				Code: `function f<A>(fn: <A>(x: A) => A): A { return fn as any; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Function type in return position with shadowing generic ----
			{
				Code: `function f<T>(): <T>(x: T) => T { return null as any; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Nested catch clause shadow ----
			{
				Code: `try {} catch (e) { try {} catch (e) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Method with same generic name as class generic ----
			{
				Code: `class C<T> { m<T>(x: T): T { return x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Double shadow chain ----
			{
				Code: `const x = 1; function f(x: number) { function g(x: number) { return x; } return g; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
				},
			},

			// ---- Array pattern with elision shadow ----
			{
				Code: `const a = 1; function f([, a]: any[]) { return a; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Optional-chain call with arrow param shadow ----
			{
				Code: `const a = 1; const fn = { call: (cb: any) => cb }; fn.call?.(a => a);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Parameter property with same name as outer ----
			{
				Code: `const x = 1; class C { constructor(public x: number) {} m() { const x = 1; return x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
				},
			},

			// ---- Dynamic import callback shadow ----
			{
				Code: `const mod = 1; import('m').then((mod) => mod);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- TS 4.7+: infer T extends U ----
			{
				Code: `type X<T> = T extends Array<infer U> ? (U extends Array<infer U> ? U : never) : never;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- TS 4.7 generic instantiation expression + builtinGlobals ----
			{
				Code:    `const g = Array<number>; function f(Array: any) { return Array; }`,
				Options: map[string]interface{}{"builtinGlobals": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadowGlobal"},
				},
			},

			// ---- TS 4.9 `accessor` field ----
			{
				Code: `const x = 1; class C { accessor x = 1; m() { const x = 2; return x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Async iteration with destructure ----
			{
				Code: `async function f() { const x = 1; for await (const { v: x } of [Promise.resolve({ v: 1 })]) { return x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- using + for-of ----
			{
				Code: `async function f() { using u = { [Symbol.dispose]() {} }; for (const u of [] as any[]) { void u; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Destructure with computed key + inner rebind ----
			{
				Code: `const k = 'a'; function f({ [k]: v }: any) { const k = 1; return [v, k]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- HOC pattern: component param shadowed by inner ----
			{
				Code: `function withTheme<P>(Component: any) { return function ThemedComponent(props: P) { const Component = 1; return Component; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Reducer pattern with case block ----
			{
				Code: `type A = { type: 'inc' } | { type: 'set'; value: number }; function reducer(state: number, action: A) { switch (action.type) { case 'inc': { const state = 1; return state + 1; } case 'set': return action.value; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Array chain with same callback name (3 separate shadows) ----
			{
				Code: `function chain() { const item = { id: 1 }; [1, 2, 3].map(item => item + 1).filter(item => item > 1).forEach(item => void item); void item; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
				},
			},

			// ---- Triply nested try/catch ----
			{
				Code: `const e = 1; try { try { throw new Error(); } catch (e) { try { throw e; } catch (e) {} } } catch (e) {} void e;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
					{MessageId: "noShadow"},
				},
			},

			// ---- Computed method name + param shadow ----
			{
				Code: `const methodName = 'foo'; const obj = { [methodName](methodName: string) { return methodName; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Nested for loops with same loop var ----
			{
				Code: `function f() { for (let i = 0; i < 1; i++) { for (let i = 0; i < 1; i++) {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Method overload: impl-signature param shadow ----
			{
				Code: `const a = 1; class C { m(x: string): void; m(x: number): void; m(a: any): void { void a; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Factory returning object method with generic shadow ----
			{
				Code: `function mk<T>(x: T) { return { get<T>(): T { return null as any; } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Empty block doesn't prevent later shadow ----
			{
				Code: `const c = 1; { /* empty */ } { const c = 2; void c; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- const enum + param shadow ----
			{
				Code: `const enum E { A, B } function f(E: number) { return E; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Abstract class with generic + abstract method generic shadow ----
			{
				Code: `abstract class C<T extends object> { abstract m<T>(): T; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Symbol.iterator method with shadowed param ----
			{
				Code: `const iter = 1; class C { [Symbol.iterator](iter: number) { return iter; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Constructor with optional param shadow ----
			{
				Code: `const x = 1; class C { constructor(x?: number) { void x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- typeof type query + same-name parameter ----
			{
				Code: `const val = { x: 1 }; function f(val: typeof val) { return val; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Async method param shadow ----
			{
				Code: `const x = 1; class C { async m(x: number) { return x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Generator method param shadow ----
			{
				Code: `const x = 1; class C { *m(x: number) { yield x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Ambient function shadowed by inner const (bot #1 / #3) ----
			{
				Code: `declare function foo(): void; function bar() { const foo = 1; return foo; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Heritage clause IIFE shadow (bot #2) ----
			{
				Code: `const h = 1; class C extends (function() { const h = 2; return class {}; })() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Heritage clause comma-expression arrow shadow ----
			{
				Code: `const h = 1; class C extends ((h => h)(1), Object) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Class decorator factory body shadow (bot #2) ----
			{
				Code: `function dec(x: any) { return x; }
@((target: any) => { const dec = 1; return dec && target; })
class D {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Method decorator body shadow ----
			{
				Code: `const md = 1; class C { @((t: any, k: string) => { const md = 1; void md; }) method() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Parameter decorator body shadow ----
			{
				Code: `const pd = 1; class C { method(@((t: any, k: any, i: number) => { const pd = 1; void pd; }) x: number) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Computed key on class method (self-discovered during audit) ----
			{
				Code: `const k = 1; class C { [((k) => k)(2)]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Computed key on class property ----
			{
				Code: `const k = 1; class C { [((k) => k)(2)] = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Computed key on getter ----
			{
				Code: `const g = 1; class C { get [((g) => g)(1)]() { return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Computed key on async generator method ----
			{
				Code: `const m = 1; class C { async *[((m) => m)(1)]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ---- Object-literal method with shadow inside computed key ----
			{
				Code: `const k = 1; const o = { [((k) => k)(2)]() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Object-literal getter with shadow inside computed key ----
			{
				Code: `const k = 1; const o = { get [((k) => k)(1)]() { return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},
			// ---- Object-literal setter with shadow inside computed key ----
			{
				Code: `const k = 1; const o = { set [((k) => k)(1)](v: number) {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadow"},
				},
			},

			// ==== Additional ESLint invalid cases for full coverage ====

			{Code: `let x = foo((x,y) => {});
let y;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}, {MessageId: "noShadow"}}},
			{Code: `function a() {}
foo(a => {});`, Options: map[string]interface{}{"ignoreOnInitialization": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `let x = ((x,y) => {})();
let y;`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}, {MessageId: "noShadow"}}},
			{Code: `
  type T = 1;
  {
	type T = 2;
  }
		`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  type T = 1;
  function foo<T>(arg: T) {}
		`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  function foo<T>() {
	return function <T>() {};
  }
		`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  type T = string;
  function foo<T extends (arg: any) => void>(arg: T) {}
		`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  const arg = 0;
  
  declare const Test: {
	new (arg: string): typeof arg;
  };
		`, Options: map[string]interface{}{"ignoreFunctionTypeParameterNameValueShadow": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  import type { foo } from './foo';
  function doThing(foo: number) {}
		`, Options: map[string]interface{}{"ignoreTypeValueShadow": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  import { type foo } from './foo';
  function doThing(foo: number) {}
		`, Options: map[string]interface{}{"ignoreTypeValueShadow": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  import { foo } from './foo';
  function doThing(foo: number, bar: number) {}
		`, Options: map[string]interface{}{"ignoreTypeValueShadow": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  interface Foo {}
  
  declare module 'bar' {
	export interface Foo {
	  x: string;
	}
  }
		`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  import type { Foo } from 'bar';
  
  declare module 'baz' {
	export interface Foo {
	  x: string;
	}
  }
		`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  import { type Foo } from 'bar';
  
  declare module 'baz' {
	export interface Foo {
	  x: string;
	}
  }
		`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  let x = foo((x, y) => {});
  let y;
		`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}, {MessageId: "noShadow"}}},
			{Code: `
  let x = foo((x, y) => {});
  let y;
		`, Options: map[string]interface{}{"hoist": "functions"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  {
	interface A {}
  }
  type A = 1;
		`, Options: map[string]interface{}{"hoist": "types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  {
	interface A {}
  }
  type A = 1;
		`, Options: map[string]interface{}{"hoist": "all"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
	if (true) {
		const foo = 6;
	}

	function foo() { }
		`, Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
	// types
	type Bar<Foo> = 1;
	type Foo = 1;

	// functions
	if (true) {
		const b = 6;
	}

	function b() { }
		`, Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}, {MessageId: "noShadow"}}},
			{Code: `
	// types
	type Bar<Foo> = 1;
	type Foo = 1;

	// functions
	if (true) {
		const b = 6;
	}

	function b() { }
		`, Options: map[string]interface{}{"hoist": "types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
  {
	interface A {}
  }
  type A = 1;
		`, Options: map[string]interface{}{"hoist": "functions-and-types"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
			{Code: `
			const A = 2;
			enum Test {
				A = 1,
				B = A,
			}
		`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noShadow"}}},
		},
	)
}
