package no_undef

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUndefRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUndefRule,
		[]rule_tester.ValidTestCase{
			// === Variable / function / class declarations ===
			{Code: `var a = 1; a;`},
			{Code: `let b = 2; b;`},
			{Code: `const c = 3; c;`},
			{Code: `function f() { } f();`},
			{Code: `class MyClass {} new MyClass();`},
			{Code: `var f = function() {}; f();`},
			{Code: `var f = () => {}; f();`},
			{Code: `var a; a = 1;`},

			// === Parameters ===
			{Code: `function f(x: number) { return x; }`},
			{Code: `function foo(a: number, b: string) { return a + b; }`},

			// === typeof (default: no report) ===
			{Code: `typeof a`},
			{Code: `typeof a === 'string'`},
			// typeof with parentheses (ESTree has no ParenthesizedExpression)
			{Code: `typeof (a)`},
			{Code: `typeof ((a))`},

			// === Property access / object literal keys ===
			{Code: `var obj = { x: 1 }; obj.x;`},
			{Code: `var obj = { key: 1 };`},

			// === Labels ===
			{Code: `loop: for (var i = 0; i < 10; i++) { break loop; }`},

			// === Host globals must be configured explicitly ===
			{Code: `console.log("test");`, Globals: map[string]any{"console": "readonly"}},
			{Code: `var p = new Promise<void>((resolve) => resolve());`},
			{Code: `setTimeout(() => {}, 100);`, Globals: map[string]any{"setTimeout": "readonly"}},

			// === Type-only positions: type annotations ===
			{Code: `type MyType = string; var x: MyType;`},
			{Code: `interface MyInterface { x: number; } var y: MyInterface;`},
			{Code: `function f(x: string): string { return x; }`},

			// === Type-only positions: generic type arguments ===
			{Code: `function identity<T>(val: T): T { return val; }`},
			{Code: `function constrained<T extends object>(val: T) { return val; }`},
			{Code: `declare let stream: AsyncIterator<string>;`},

			// === Type-only positions: as / satisfies (type part is type-only) ===
			{Code: `var x = 1; var y = x as any;`},
			{Code: `var x = { a: 1 } satisfies Record<string, number>;`},

			// === Type-only positions: typeof in type position (TypeQuery) ===
			{Code: `var x = 1; type T = typeof x;`},

			// === Type-only positions: mapped / conditional / indexed types ===
			{Code: `interface I { a: number; } type Mapped = { [K in keyof I]: string };`},
			{Code: `type IsStr<T> = T extends string ? true : false;`},

			// === Type-only positions: interface extends (type-only) ===
			{Code: `interface Base { x: number; } interface Derived extends Base { y: number; }`},

			// === Type-only positions: class implements (type-only) ===
			{Code: `interface I { x: number; } class C implements I { x = 1; }`},

			// === Value positions: class extends (value!) ===
			{Code: `class Base {} class Child extends Base {}`},
			// Nested class extends
			{Code: `class A {} function f() { class B extends A {} return B; }`},

			// === Shorthand property with declared variable ===
			{Code: `var x = 1; var obj = { x };`},

			// === Destructuring ===
			{Code: `var { a, b } = { a: 1, b: 2 }; a; b;`},
			{Code: `var [x, y] = [1, 2]; x; y;`},

			// === For loop variables ===
			{Code: `for (var i = 0; i < 10; i++) { i; }`},
			{Code: `for (let i = 0; i < 10; i++) { i; }`},

			// === Catch clause variable ===
			{Code: `try {} catch (e) { e; }`},

			// === Class members ===
			{Code: `class Foo { bar() {} }; new Foo().bar();`},
			{Code: `class Foo { bar = 1; baz() { return this.bar; } }`},
			{Code: `var x = 1; x++;`},

			// === Enum ===
			{Code: `enum Direction { Up, Down }; Direction.Up;`},

			// === /*global*/ comments ===
			{Code: `/*global myVar*/ myVar = 1;`},
			{Code: `/*global a, b*/ a = 1; b = 2;`},
			{Code: `/*global myVar:writable*/ myVar = 1;`},
			{Code: `/*global myVar*/ /*global myVar:bogus*/ myVar;`},

			// === languageOptions.globals (config) ===
			{Code: `myConfiguredGlobal;`, Globals: map[string]any{"myConfiguredGlobal": "readonly"}},

			// === "off" global with a same-file declaration: the binding wins ===
			{Code: `var myOffGlobal123 = 1; myOffGlobal123;`, Globals: map[string]any{"myOffGlobal123": "off"}},
			{Code: `import Promise from "pkg"; Promise;`, Globals: map[string]any{"Promise": "off"}},

			// === ECMAScript language globals are selected by ecmaVersion ===
			{Code: `Object; toString; hasOwnProperty;`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 3}},
			{Code: `JSON;`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 5}},
			{Code: `Promise;`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{
				Code:            `Promise;`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 3},
				Globals:         map[string]any{"Promise": "readonly"},
			},
			{Code: `Atomics;`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2017}},
			{Code: `BigInt;`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
			{Code: `WeakRef;`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2021}},
			{Code: `Float16Array; Iterator;`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025}},
			{Code: `Temporal; AsyncDisposableStack;`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2026}},
			// Zero value means the moving `latest` edition.
			{Code: `Temporal;`},

			// === JSX: intrinsic (lowercase) tags are not identifier references ===
			{Code: `function C() { return <div />; }`, Tsx: true},
			{Code: `const el = <foo-bar />;`, Tsx: true},

			// === JSX: uppercase tags reference a declared component ===
			{Code: `function Foo() { return null; } const el = <Foo />;`, Tsx: true},

			// === languageOptions.globals + /*global*/ comment together ===
			{Code: `/*global fromComment*/ fromConfig; fromComment;`, Globals: map[string]any{"fromConfig": "readonly"}},

			// === Namespace ===
			{Code: `namespace MyNS { export var x = 1; } MyNS.x;`},

			// === Ambient declaration ===
			{Code: `declare var declaredAmbient: number; declaredAmbient;`},

			// === Type-only positions: union / intersection ===
			{Code: `type U = string | number;`},
			{Code: "type I2 = string & { tag: string };"},

			// === Type-only positions: tuple / array / function type ===
			{Code: `type Tup = [string, number];`},
			{Code: `type Arr = string[];`},
			{Code: `type Fn = (x: string) => number;`},

			// === Type-only positions: type predicate ===
			{Code: `function isStr(x: any): x is string { return typeof x === 'string'; }`},

			// === Type-only positions: nested generics ===
			{Code: `type Nested = Map<string, Array<Promise<number>>>;`},

			// === Type-only positions: index signature ===
			{Code: `interface Indexed { [key: string]: number; }`},

			// === Type-only positions: keyof ===
			{Code: `interface KI { a: number; } type Keys = keyof KI;`},

			// === Type-only positions: infer ===
			{Code: `type Unpack<T> = T extends Array<infer U> ? U : T;`},

			// === Type-only positions: template literal type ===
			{Code: "type EventName = `on${string}`;"},

			// === Generic type argument in class extends (base is value, arg is type) ===
			{Code: "class GenBase<T> { value!: T; } class GenChild extends GenBase<string> {}"},

			// === Both extends + implements ===
			{Code: `class BothBase {} interface IC { c(): void; } class Both extends BothBase implements IC { c() {} }`},

			// === Class expression extends (declared) ===
			{Code: `class ExprBase {} var CE = class extends ExprBase {};`},

			// === Multiple implements ===
			{Code: `interface IA { a(): void; } interface IB { b(): void; } class Multi implements IA, IB { a() {} b() {} }`},

			// === import.meta (meta property, not identifier reference) ===
			{Code: `var url = import.meta.url;`},

			// === new.target in constructor ===
			{Code: `class C { constructor() { var t = new.target; } }`},

			// === as const ===
			{Code: `var x = [1, 2, 3] as const;`},

			// === Assertion function return type (type-only) ===
			{Code: `function assertDefined(x: any): asserts x is string { if (!x) throw new Error(); }`},

			// === Conditional type with infer ===
			{Code: `type RetType<T> = T extends (...args: any[]) => infer R ? R : never;`},

			// === Index access type (type-only) ===
			{Code: `interface Obj { key: string; } type Val = Obj['key'];`},

			// === Destructuring rename (property key should not be reported) ===
			{Code: `var obj = { a: 1 }; var { a: renamed } = obj; renamed;`},

			// === Nested destructuring rename ===
			{Code: `var obj = { a: { b: 1 } } as any; var { a: { b: renamed } } = obj; renamed;`},

			// === Declared default parameter ===
			{Code: `var defaultVal = 1; function f(x = defaultVal) { return x; }`},

			// === Declared class property initializer ===
			{Code: `var propInit = 42; class C { x = propInit; }`},

			// === Rest element in destructuring ===
			{Code: `var [first, ...rest] = [1, 2, 3]; first; rest;`},
			{Code: `var obj = { a: 1, b: 2 }; var { a, ...others } = obj; a; others;`},

			// === Built-in globals ===
			{Code: `var g = globalThis;`},
			{Code: `var u = undefined;`},
			{Code: `var n = NaN;`},
			{Code: `var inf = Infinity;`},

			// === arguments inside function ===
			{Code: `function f() { return arguments; }`},
			{Code: `function f() { return arguments; }`, Globals: map[string]any{"arguments": "off"}},

			// === this in class method ===
			{Code: `class C { x = 1; m() { return this.x; } }`},

			// === Chained property access (only root is checked) ===
			{Code: `var obj = { a: { b: { c: 1 } } }; obj.a.b.c;`},

			// === Export declared ===
			{Code: `var exportedVar = 1; export { exportedVar };`},
			{Code: `var localVar = 2; export { localVar as renamedExport };`},
			{Code: `type ExportedType = string; export type { ExportedType };`},

			// === delete on declared ===
			{Code: `var obj = { prop: 1 } as any; delete obj.prop;`},

			// === Symbol property (built-in) ===
			{Code: `var iter = { [Symbol.iterator]() { return { next() { return { done: true, value: undefined }; } }; } };`},

			// === Computed destructuring key with declared ===
			{Code: `var key = 'a'; var { [key]: val } = { a: 1 } as any; val;`},

			// === Element access with declared ===
			{Code: `var arr = [1, 2, 3]; var idx = 1; var v = arr[idx];`},

			// === Getter / setter with declared ===
			{Code: `var getterVal = 1; var obj = { get x() { return getterVal; } };`},

			// === Private field with declared ===
			{Code: `var initVal = 10; class C { #x = initVal; m() { return this.#x; } }`},

			// === Parenthesized declared ===
			{Code: `var declared = 1; var v = (declared);`},

			// === keyof typeof (type-only) ===
			{Code: `var someObj = { a: 1 }; type Keys = keyof typeof someObj;`},

			// === Mapped type with as clause (type-only) ===
			{Code: `interface Src { a: number; b: string; } type OnlyStrings = { [K in keyof Src as Src[K] extends string ? K : never]: Src[K] };`},

			// === Multiple interface extends (type-only) ===
			{Code: `interface A { a: number; } interface B { b: string; } interface AB extends A, B {}`},

			// === Optional parameter type (type-only) ===
			{Code: `type MyType = string; function f(x?: MyType) { return x; }`},

			// === Class static block with declared ===
			{Code: `var staticInit = 42; class C { static { var v = staticInit; } }`},

			// === Import alias (import { Original as Alias }) ===
			{Code: `import { resolve as r } from "path"; r("/");`},
			{Code: `import { join as j, resolve as r } from "path"; j("a", "b"); r("/");`},

			// === Import type alias (import type { X as Y }) ===
			{Code: `import type { PlatformPath as PP } from "path";`},
			{Code: `import type { Configuration } from "pkg"; Configuration;`},

			// === Import types are type-only, including tsgo's JSDoc reparse ===
			{Code: `type Config = import("pkg").Configuration;`},
			{
				Code:     `/** @type {import("pkg").Configuration} */ const config = {};`,
				FileName: "jsdoc-import-type.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     `/** @type {import("pkg").Configuration} */ export default {};`,
				FileName: "jsdoc-import-type.mjs",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     `/** @param {import("pkg").Configuration} config */ function use(config) { return config; }`,
				FileName: "jsdoc-param-import-type.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\n/** @type {Configuration} */\nconst config = {};",
				FileName: "jsdoc-import-tag-type-only.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nconst Configuration = {};\nConfiguration;",
				FileName: "jsdoc-import-tag-local-value.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "const Configuration = {};\n/** @import { Configuration } from \"pkg\" */\nConfiguration;",
				FileName: "jsdoc-import-tag-earlier-local-value.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nfunction use(Configuration) { return Configuration; }",
				FileName: "jsdoc-import-tag-parameter-shadow.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "/** @import { Configuration } from \"types\" */\nimport { Configuration } from \"values\";\nConfiguration;",
				FileName: "jsdoc-import-tag-real-import.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "import { Configuration } from \"values\";\n/** @import { Configuration } from \"types\" */\nConfiguration;",
				FileName: "jsdoc-import-tag-earlier-real-import.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "/** @import { Configuration as Config } from \"types\" */\nimport Config from \"values\";\nConfig;",
				FileName: "jsdoc-import-tag-real-default-import.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "/** @import * as Types from \"types\" */\nimport * as Types from \"values\";\nTypes;",
				FileName: "jsdoc-import-tag-real-namespace-import.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "/** @import { Alias } from \"types\" */\nimport { Original as Alias } from \"values\";\nAlias;",
				FileName: "jsdoc-import-tag-real-aliased-import.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nConfiguration;",
				FileName: "jsdoc-import-tag-config-global.js",
				TSConfig: "tsconfig.allowJs.json",
				Globals:  map[string]any{"Configuration": "readonly"},
			},

			// === Re-export alias (export { X as Y } from 'module') ===
			{Code: `export { resolve as r } from "path";`},
			{Code: `export type { PlatformPath as PP } from "path";`},

			// === Parameter initializers resolve outside the function body ===
			{Code: `function Foo() {}
function f(x = new Foo()) { function Foo() {} }`},
			{Code: `function f(Foo, x = new Foo()) {}`},
			{Code: `const f = function Foo(x = new Foo()) { function Foo() {} };`},
		},
		[]rule_tester.InvalidTestCase{
			// === Basic undeclared references ===
			{
				Code: `a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},
			{
				Code: `var a = b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
				},
			},
			{
				Code: `undeclaredFunc();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},
			{
				Code:     "/** @type {import(\"pkg\").Configuration} */\nconst config = {};\nConfiguration;",
				FileName: "jsdoc-import-type-runtime-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 3, Column: 1},
				},
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nConfiguration;",
				FileName: "jsdoc-import-tag-runtime-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 1},
				},
			},
			{
				Code:     "/** @import { Configuration as Config } from \"pkg\" */\nConfig;",
				FileName: "jsdoc-import-tag-alias-runtime-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 1},
				},
			},
			{
				Code:     "/** @import * as Types from \"pkg\" */\nTypes;",
				FileName: "jsdoc-import-tag-namespace-runtime-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 1},
				},
			},
			// A parameter initializer is evaluated in a scope outside the body, so a
			// function declared there does not define the name.
			{
				Code: `function f(x = new Foo()) { function Foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 20},
				},
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\ntypeof Configuration;",
				FileName: "jsdoc-import-tag-typeof-runtime-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Options:  map[string]any{"typeof": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 8},
				},
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nfunction read() {\n  return Configuration;\n}",
				FileName: "jsdoc-import-tag-nested-runtime-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 3, Column: 10},
				},
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nfunction read(Configuration) { return Configuration; }\nConfiguration;",
				FileName: "jsdoc-import-tag-shadowed-and-unshadowed.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 3, Column: 1},
				},
			},
			{
				Code:     "function read() {\n  /** @import { Configuration } from \"pkg\" */\n  return Configuration;\n}",
				FileName: "nested-jsdoc-import-tag-runtime-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 3, Column: 10},
				},
			},
			{
				Code:     "/** @import * as Types from \"pkg\" */\nTypes.Configuration;",
				FileName: "jsdoc-import-tag-namespace-property-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 1},
				},
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nconst value = { Configuration };",
				FileName: "jsdoc-import-tag-shorthand-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 17},
				},
			},
			{
				Code:     "/** @import Configuration from \"pkg\" */\nConfiguration;",
				FileName: "jsdoc-import-tag-default-runtime-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 1},
				},
			},
			{
				Code:     "/** @import { Configuration as Config } from \"one\" */\n/** @import { Other as Config } from \"two\" */\nConfig;",
				FileName: "jsdoc-import-tag-duplicate-runtime-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 3, Column: 1},
				},
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nConfiguration;",
				FileName: "jsdoc-import-tag-off-global.js",
				TSConfig: "tsconfig.allowJs.json",
				Globals:  map[string]any{"Configuration": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 1},
				},
			},
			{
				Code:     "/** @import { Component } from \"pkg\" */\nconst el = <Component />;",
				FileName: "jsdoc-import-tag-jsx-component.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 13},
				},
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nconst value = { [Configuration]: true };",
				FileName: "jsdoc-import-tag-computed-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 18},
				},
			},
			{
				Code:     "/** @import { Original } from \"types\" */\nimport { Original as Alias } from \"values\";\nOriginal;",
				FileName: "jsdoc-import-tag-imported-property-name.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 3, Column: 1},
				},
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nConfiguration;",
				FileName: "jsdoc-import-tag-runtime-reference.mjs",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 1},
				},
			},
			{
				Code:     "/** @import { Configuration } from \"pkg\" */\nConfiguration;",
				FileName: "jsdoc-import-tag-runtime-reference.cjs",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 1},
				},
			},

			// === ecmaVersion boundary attacks ===
			{
				Code:            `JSON;`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 3},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:            `Promise;`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:            `Atomics;`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2016},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:            `BigInt;`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2019},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:            `WeakRef;`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:            `Float16Array;`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:            `Temporal;`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			// TypeScript exposes AsyncIterator only as a type; ESLint 10.8
			// does not provide it as a runtime language global.
			{
				Code:   `AsyncIterator;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},

			// === TypeChecker environment must not define host/ambient names ===
			{
				Code: "console;\nwindow;\nsetTimeout;\nprocess;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
					{MessageId: "undef", Line: 2, Column: 1},
					{MessageId: "undef", Line: 3, Column: 1},
					{MessageId: "undef", Line: 4, Column: 1},
				},
			},

			// === typeof with checkTypeof: true ===
			{
				Code:    `typeof anUndefinedVar === 'string'`,
				Options: map[string]interface{}{"typeof": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 8},
				},
			},
			// typeof with parentheses + checkTypeof: true
			{
				Code:    `typeof (anUndefinedVar)`,
				Options: map[string]interface{}{"typeof": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
				},
			},

			// === Multiple undeclared variables ===
			{
				Code: `var x = foo + bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
					{MessageId: "undef", Line: 1, Column: 15},
				},
			},

			// === Nested scope ===
			{
				Code: `function foo() { return undeclaredVar123; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 25},
				},
			},

			// === Various expression positions ===
			{
				Code: `var x = unknownFunc123();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
				},
			},
			{
				Code: `if (unknownCondition123) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 5},
				},
			},

			// === /*global*/ mismatch ===
			{
				Code: `/*global otherVar*/ unknownVar123 = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 21},
				},
			},

			// === languageOptions.globals explicitly "off" still reports ===
			{
				Code:    `myOffGlobal123;`,
				Globals: map[string]any{"myOffGlobal123": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},
			// Explicit globals:off replaces the selected language-global set.
			{
				Code:    `Promise;`,
				Globals: map[string]any{"Promise": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},

			// === /*global name:off*/ still reports (matches config "off" semantics) ===
			{
				Code: `/*global myOffComment123:off*/ myOffComment123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 32},
				},
			},

			// === "off" un-declares a global the checker knows from lib ===
			{
				Code: `/*global console:off*/ console.log(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 24},
				},
			},
			{
				Code:    `setTimeout(() => {}, 100);`,
				Globals: map[string]any{"setTimeout": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},

			// === JSX: undeclared uppercase tag reports on opening and closing
			// tags alike (typescript-eslint parity) ===
			{
				Code: `const el = <UndeclaredComp123 />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
				},
			},
			{
				Code: `const el = <UndeclaredComp456></UndeclaredComp456>;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
					{MessageId: "undef", Line: 1, Column: 33},
				},
			},

			// === JSX: hyphenated uppercase tag is still a component reference
			// (typescript-eslint parity) ===
			{
				Code: `const el = <Foo-bar />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
				},
			},

			// === Shorthand property with undeclared variable ===
			{
				Code: `var obj = { undeclaredShorthand123 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
				},
			},

			// === Class extends with undeclared base ===
			{
				Code: `class Child extends undeclaredBase123 {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 21},
				},
			},

			// === Template literal ===
			{
				Code: "var s = `${undeclaredTpl123}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 12},
				},
			},

			// === Array literal ===
			{
				Code: `var arr = [undeclaredArr123];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 12},
				},
			},

			// === Destructuring default value ===
			{
				Code: `var { d = undeclaredDefault123 } = {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 11},
				},
			},

			// === Ternary condition ===
			{
				Code: `var x = undeclaredTernary123 ? 1 : 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
				},
			},

			// === Optional chaining on undeclared ===
			{
				Code: `var x = undeclaredOptional123?.prop;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
				},
			},

			// === Computed property key ===
			{
				Code: `var obj = { [undeclaredComputed123]: 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 14},
				},
			},

			// === Spread element ===
			{
				Code: `var arr = [...undeclaredSpread123];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 15},
				},
			},

			// === As expression value side ===
			{
				Code: `var x = undeclaredAsVal123 as any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
				},
			},

			// === Nested class extends with undeclared base ===
			{
				Code: `function f() { class Inner extends undeclaredNested123 {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 36},
				},
			},

			// === Undeclared in enum value ===
			{
				Code: `enum E { A = undeclaredEnumVal123 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 14},
				},
			},

			// === Function argument ===
			{
				Code:    `console.log(undeclaredArg123);`,
				Globals: map[string]any{"console": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
				},
			},

			// === Nested shorthand property ===
			{
				Code: `var obj = { nested: { undeclaredNestedShort123 } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 23},
				},
			},

			// === Class expression extends undeclared ===
			{
				Code: `var CE = class extends undefClassExpr123 {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 24},
				},
			},

			// === new undeclared ===
			{
				Code: `new undefNew123();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 5},
				},
			},

			// === Arrow function body ===
			{
				Code: `var f1 = () => undefArrow123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 16},
				},
			},

			// === for-of iterable ===
			{
				Code: `for (var x1 of undefForOf123) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 16},
				},
			},

			// === for-in object ===
			{
				Code: `for (var x2 in undefForIn123) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 16},
				},
			},

			// === throw ===
			{
				Code: `function throwIt() { throw undefThrow123; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 28},
				},
			},

			// === Logical AND ===
			{
				Code: `var v1 = true && undefAnd123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 18},
				},
			},

			// === Nullish coalescing ===
			{
				Code: `var v2 = null ?? undefNullish123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 18},
				},
			},

			// === Unary NOT ===
			{
				Code: `var v3 = !undefNot123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 11},
				},
			},

			// === Unary negation ===
			{
				Code: `var v4 = -undefNeg123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 11},
				},
			},

			// === Tagged template ===
			{
				Code: "undefTag123`hello`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},

			// === void ===
			{
				Code: `void undefVoid123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 6},
				},
			},

			// === instanceof ===
			{
				Code: `var v5 = ({}) instanceof undefInst123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 26},
				},
			},

			// === Deeply nested arrow ===
			{
				Code: `var f2 = () => () => undefDeep123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 22},
				},
			},

			// === Generator yield ===
			{
				Code: `function* gen() { yield undefYield123; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 25},
				},
			},

			// === Async await ===
			{
				Code: `async function af() { await undefAwait123; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 29},
				},
			},

			// === satisfies value side ===
			{
				Code: `var v6 = undefSatisfies123 satisfies any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},

			// === Multiple shorthand undeclared ===
			{
				Code: `var obj1 = { undefShortA123, undefShortB123 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 14},
					{MessageId: "undef", Line: 1, Column: 30},
				},
			},

			// === Mixed declared/undeclared shorthand ===
			{
				Code: `var declared = 1; var obj2 = { declared, undefShortMix123 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 42},
				},
			},

			// === Shorthand in class method ===
			{
				Code: `class CS { m() { return { undefMethShort123 }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 27},
				},
			},

			// === Nested in if-else ===
			{
				Code: `if (true) { var v7 = undefIfElse123; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 22},
				},
			},

			// === Switch case ===
			{
				Code: `switch (1) { case 1: var v8 = undefSwitch123; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 31},
				},
			},

			// === Logical OR assignment ===
			{
				Code: `var v9: any; v9 ||= undefLogical123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 21},
				},
			},

			// === Export default undeclared ===
			{
				Code: `export default undefExportDefault123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 16},
				},
			},

			// === Computed method name ===
			{
				Code: `class C1 { [undefComputedMethod123]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
				},
			},

			// === Non-null assertion ===
			{
				Code: `var v1 = undefNonNull123!;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},

			// === Angle-bracket type assertion ===
			{
				Code: `var v2 = <any>undefAngleBracket123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 15},
				},
			},

			// === for-of assignment target (not declaration) ===
			{
				Code: `for (undefForOfTarget123 of [1, 2]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 6},
				},
			},

			// === Destructuring assignment (not declaration) ===
			{
				Code: `var arr: number[]; [undefAssignTarget123] = [1];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 21},
				},
			},

			// === Default parameter value ===
			{
				Code: `function f1(x = undefDefaultParam123) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 17},
				},
			},

			// === Class property initializer ===
			{
				Code: `class C2 { x = undefClassProp123; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 16},
				},
			},

			// === Class static property ===
			{
				Code: `class C3 { static x = undefStaticProp123; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 23},
				},
			},

			// === Object spread ===
			{
				Code: `var obj1 = { ...undefObjSpread123 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 17},
				},
			},

			// === Ternary both branches ===
			{
				Code: `var v1 = true ? undefTernaryA123 : undefTernaryB123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 17},
					{MessageId: "undef", Line: 1, Column: 36},
				},
			},

			// === Nested destructuring default (only default value, not property key) ===
			{
				Code: `var { a: { b = undefNestedDefault123 } = {} as any } = {} as any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 16},
				},
			},

			// === Comma expression ===
			{
				Code: `var v2 = (1, undefComma123);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 14},
				},
			},

			// === Exponentiation ===
			{
				Code: `var v3 = undefExponent123 ** 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},

			// === Assignment operator ===
			{
				Code: `var v4: any; v4 += undefPlusAssign123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 20},
				},
			},

			// === Nullish assignment ===
			{
				Code: `var v5: any; v5 ??= undefNullishAssign123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 21},
				},
			},

			// === in operator ===
			{
				Code: `var v6 = 'key' in undefInOperator123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 19},
				},
			},

			// === delete on undeclared ===
			{
				Code: `delete undefDelete123.prop;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 8},
				},
			},

			// === IIFE with undeclared ===
			{
				Code: `(function() { return undefIIFE123; })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 22},
				},
			},

			// === Arrow IIFE ===
			{
				Code: `(() => undefArrowIIFE123)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 8},
				},
			},

			// === Deep nested destructuring default ===
			{
				Code: `var { x: { y: { z = undefDeepDefault123 } = {} as any } = {} as any } = {} as any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 21},
				},
			},

			// === Undeclared in template literal conditional ===
			{
				Code: "var v7 = `${undefInTemplate123 ? 'a' : 'b'}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
				},
			},

			// === Logical OR ===
			{
				Code: `var v8 = false || undefLogicalOr123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 19},
				},
			},

			// === Computed destructuring key ===
			{
				Code: `var { [undefComputedKey123]: val } = {} as any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 8},
				},
			},

			// === Element access expression ===
			{
				Code: `var obj: any = {}; var v = obj[undefElementAccess123];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 32},
				},
			},

			// === Class static block ===
			{
				Code: `class C1 { static { var v = undefStaticBlock123; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 29},
				},
			},

			// === Getter body ===
			{
				Code: `var obj = { get x() { return undefGetter123; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 30},
				},
			},

			// === Setter body ===
			{
				Code: `var obj = { set x(v: any) { var t = undefSetter123; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 37},
				},
			},

			// === Private field initializer ===
			{
				Code: `class C2 { #x = undefPrivateField123; m() { return this.#x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 17},
				},
			},

			// === Parenthesized undeclared ===
			{
				Code: `var v = (undefParenthesized123);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},

			// === Double non-null ===
			{
				Code: `var v = undefDoubleNonNull123!!;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
				},
			},

			// === Nested as/satisfies ===
			{
				Code: `var v = (undefNestedAs123 as any) satisfies any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},

			// === Local export alias with undeclared (export { X as Y } without `from`) ===
			{
				Code: `export { undefLocalExport123 as aliased };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},

			// === Non-aliased local export of an undeclared name reads the
			// local binding (typescript-eslint parity) ===
			{
				Code: `export { undefPlainExport123 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},
			{
				Code: `export type { UndefTypeExport123 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 15},
				},
			},

			// === Optional call on undeclared ===
			{
				Code: `undefOptionalCall123?.();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},
		},
	)
}

func TestNoUndefTypeScriptReferenceParity(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUndefRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `type A = Record<string, AsyncIterator<number>>;
type B = IteratorObjectConstructor;
const value = {} as const;
type C = const;`,
				FileName: "default-esnext-type-globals.ts",
			},
			{
				Code:            `type A = Intl; type B = Reflect; type C = Temporal;`,
				FileName:        "default-esnext-namespace-globals.ts",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
			},
			{
				Code:     `export { IteratorObjectConstructor };`,
				FileName: "default-esnext-dual-export.ts",
			},
			{
				Code:     `type T = Record<string, unknown>;`,
				FileName: "type-global-survives-config-off.ts",
				Globals:  map[string]any{"Record": "off"},
			},
			{
				Code:            `type T = Promise<void>;`,
				FileName:        "type-global-independent-of-ecma-version.ts",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
			},
			{
				Code:     `const Record = 1; type T = Record<string, unknown>;`,
				FileName: "value-binding-does-not-shadow-type-global.ts",
			},
			{
				Code:     `type T = Record.Member;`,
				FileName: "qualified-type-global.ts",
			},
			{
				Code:     `export { Record }; export type { Record as RecordType }; export default Record;`,
				FileName: "dual-type-global-exports.ts",
			},
			{
				Code:     `export = Record;`,
				FileName: "type-global-export-assignment.ts",
			},
			{
				Code:     `type Foo = {}; type T = Foo.Member;`,
				FileName: "qualified-type-alias-reference.ts",
			},
			{
				Code:     `interface Foo {}; interface Bar extends Foo.Member {}`,
				FileName: "qualified-interface-heritage-reference.ts",
			},
			{
				Code:     `const outer = 1; declare function f(x: unknown): outer is string;`,
				FileName: "type-predicate-outer-value-reference.ts",
			},
			{
				Code:     `declare namespace N {}; N;`,
				FileName: "namespace-is-value-capable.ts",
			},
			{
				Code:     `declare namespace Foo {}; import X = Foo.Member; X;`,
				FileName: "namespace-import-equals-value-reference.ts",
			},
			{
				Code:     `type T = typeof import("pkg").Foo<Record>;`,
				FileName: "type-query-import-type-argument.ts",
			},
			{
				Code:     `type T = { [K in K]: K };`,
				FileName: "mapped-type-key-visible-in-constraint.ts",
			},
			{
				Code:     `type F = (x: typeof y, y: unknown) => void;`,
				FileName: "function-type-parameter-scope.ts",
			},
			{
				Code:     `type F = <T extends U, U>() => void;`,
				FileName: "function-type-parameter-declaration-scope.ts",
			},
			{
				Code:     `const node = <div<Record> />;`,
				FileName: "jsx-intrinsic-type-argument.tsx",
				Tsx:      true,
			},
			{
				Code:     "declare namespace NS {}\ntype T = NS.Missing;",
				FileName: "declared-qualified.ts",
			},
			{
				Code:     `type T = import("pkg").MissingNS.Member;`,
				FileName: "import-type-qualifier.ts",
			},
			{
				Code:     `export type { Missing as Alias } from "pkg";`,
				FileName: "type-reexport.ts",
			},
			{
				Code:     `type F = (x: unknown) => typeof x;`,
				FileName: "function-type-parameter-query.ts",
			},
			{
				Code:     `class Existing {} type T = Existing; type U = typeof Existing;`,
				FileName: "dual-space-class.ts",
			},
			{
				Code:     `/** @type {Missing} */ const x = {};`,
				FileName: "jsdoc-missing-type.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     `/** @param {Missing} x */ function f(x) { return x; }`,
				FileName: "jsdoc-missing-param.js",
				TSConfig: "tsconfig.allowJs.json",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `Record; AsyncIterator;`,
				FileName: "type-globals-do-not-leak-into-value-space.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
					{MessageId: "undef", Line: 1, Column: 9},
				},
			},
			{
				Code:     `type T = Record<string, unknown>; Record;`,
				FileName: "type-global-config-off-does-not-define-value.ts",
				Globals:  map[string]any{"Record": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 35},
				},
			},
			{
				Code:     `type T = typeof Record;`,
				FileName: "type-query-needs-value-global.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 17},
				},
			},
			{
				Code:     `type T = { [Record]: unknown };`,
				FileName: "computed-type-key-needs-value-global.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 13},
				},
			},
			{
				Code:     `class C extends Record {}`,
				FileName: "class-extends-needs-value-global.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 17},
				},
			},
			{
				Code:     `import X = Record; X;`,
				FileName: "import-equals-needs-value-global.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 12},
				},
			},
			{
				Code:     `declare function f(x: unknown): Record is string;`,
				FileName: "type-predicate-parameter-needs-value-global.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 33},
				},
			},
			{
				Code:     `type T = HTMLElement; type U = NodeJS.Process; type V = Buffer;`,
				FileName: "unsupported-host-type-globals.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
					{MessageId: "undef", Line: 1, Column: 32},
					{MessageId: "undef", Line: 1, Column: 57},
				},
			},
			{
				Code:     `type T = IteratorConstructor;`,
				FileName: "non-default-type-global.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},
			{
				Code:            `type T = globalThis;`,
				FileName:        "runtime-global-not-default-type-global.ts",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},
			{
				Code:     `IteratorObjectConstructor; type T = typeof IteratorObjectConstructor;`,
				FileName: "type-only-default-global-in-value-space.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
					{MessageId: "undef", Line: 1, Column: 44},
				},
			},
			{
				Code:     `type Foo = {}; import X = Foo.Member; X;`,
				FileName: "type-alias-import-equals-value-reference.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 27},
				},
			},
			{
				Code:     `interface Foo {}; Foo;`,
				FileName: "interface-does-not-define-value.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 19},
				},
			},
			{
				Code:     `function Foo() {}; type T = Foo;`,
				FileName: "function-does-not-define-type.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 29},
				},
			},
			{
				Code:     `const Foo = 1; type T = Foo.Member;`,
				FileName: "value-does-not-define-qualified-type.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 25},
				},
			},
			{
				Code:     `const Foo = 1; export type { Foo };`,
				FileName: "type-only-export-excludes-value.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 30},
				},
			},
			{
				Code:     `type Foo = {}; declare function f(x: unknown): Foo is string;`,
				FileName: "type-predicate-parameter-needs-value.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 48},
				},
			},
			{
				Code:     `type T = typeof Missing<Record>;`,
				FileName: "type-query-value-with-type-argument.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 17},
				},
			},
			{
				Code:     `type T<X> = X extends infer U ? U : U;`,
				FileName: "infer-type-not-visible-in-false-branch.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 37},
				},
			},
			{
				Code:     `type T = { [K in string]: K }; type U = K;`,
				FileName: "mapped-type-key-not-visible-outside.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 41},
				},
			},
			{
				Code:     `type F = (x: unknown) => typeof x; type T = typeof x;`,
				FileName: "function-type-parameter-not-visible-outside.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 52},
				},
			},
			{
				Code:     `const node = <MissingComponent<Record> />;`,
				FileName: "jsx-component-with-type-argument.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 15},
				},
			},
			{
				Code:     `const f = Missing<Record>;`,
				FileName: "instantiation-expression-type-argument.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 11},
				},
			},
			{
				Code:     `type T = { [key: MissingKey]: MissingValue };`,
				FileName: "index-signature-type-references.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 18},
					{MessageId: "undef", Line: 1, Column: 31},
				},
			},
			{
				Code:     `type T = new <X extends Missing>() => Missing2;`,
				FileName: "constructor-type-references.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 25},
					{MessageId: "undef", Line: 1, Column: 39},
				},
			},
			{
				Code:     `type T = Missing;`,
				FileName: "missing-type-reference.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},
			{
				Code:     `type T = MissingNS.Member.Deep;`,
				FileName: "missing-qualified-root.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},
			{
				Code:     `type T = typeof MissingValue;`,
				FileName: "missing-type-query-value.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 17},
				},
			},
			{
				Code:     `type T = typeof MissingNS.value.deep;`,
				FileName: "missing-qualified-type-query-root.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 17},
				},
			},
			{
				Code:     `type F<T> = typeof T;`,
				FileName: "type-parameter-does-not-define-value.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 20},
				},
			},
			{
				Code:     `type F = (x: unknown) => x;`,
				FileName: "value-parameter-does-not-define-type.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 26},
				},
			},
			{
				Code:     `const a = value as MissingA; const b = value2 satisfies MissingB;`,
				FileName: "missing-assertion-references.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 11},
					{MessageId: "undef", Line: 1, Column: 20},
					{MessageId: "undef", Line: 1, Column: 40},
					{MessageId: "undef", Line: 1, Column: 57},
				},
			},
			{
				Code:     `interface I extends MissingNS.Base<MissingArg> {}`,
				FileName: "missing-interface-heritage.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 21},
					{MessageId: "undef", Line: 1, Column: 36},
				},
			},
			{
				Code:     `class C implements MissingNS.Base<MissingArg> {}`,
				FileName: "missing-class-implements.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 20},
					{MessageId: "undef", Line: 1, Column: 35},
				},
			},
			{
				Code:     `type T = import("pkg").Box<MissingArg>;`,
				FileName: "missing-import-type-argument.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 28},
				},
			},
			{
				Code:     `type T = import(MissingSource).Box<MissingArg>;`,
				FileName: "invalid-import-type-source.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 36},
				},
			},
			{
				Code:     `type T = import("pkg", { with: { type: MissingOption } }).Box<MissingArg>;`,
				FileName: "invalid-import-type-options.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 63},
				},
			},
			{
				Code:     `type T = { plain: MissingType; [MissingKey]: MissingValue; [MissingMethod](): MissingReturn };`,
				FileName: "missing-type-member-references.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 19},
					{MessageId: "undef", Line: 1, Column: 33},
					{MessageId: "undef", Line: 1, Column: 46},
					{MessageId: "undef", Line: 1, Column: 61},
					{MessageId: "undef", Line: 1, Column: 79},
				},
			},
			{
				Code:     `type T = { [K in keyof MissingKeys as MissingRemap<K>]: MissingValue<K> };`,
				FileName: "missing-mapped-type-references.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 24},
					{MessageId: "undef", Line: 1, Column: 39},
					{MessageId: "undef", Line: 1, Column: 57},
				},
			},
			{
				Code:     `type T = MissingObj[MissingIndex];`,
				FileName: "missing-indexed-access-types.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 10},
					{MessageId: "undef", Line: 1, Column: 21},
				},
			},
			{
				Code:     `type T = keyof Missing;`,
				FileName: "missing-keyof-type.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 16},
				},
			},
			{
				Code:     `declare function f(x: unknown): missingParam is MissingType;`,
				FileName: "missing-type-predicate-references.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 33},
					{MessageId: "undef", Line: 1, Column: 49},
				},
			},
			{
				Code:     `type F = <T extends MissingConstraint = MissingDefault>(x: MissingParam) => MissingReturn;`,
				FileName: "missing-function-type-references.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 21},
					{MessageId: "undef", Line: 1, Column: 41},
					{MessageId: "undef", Line: 1, Column: 60},
					{MessageId: "undef", Line: 1, Column: 77},
				},
			},
			{
				Code:     `fn<MissingType>();`,
				FileName: "missing-call-type-argument.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
					{MessageId: "undef", Line: 1, Column: 4},
				},
			},
			{
				Code:     "tag<MissingType>`x`;",
				FileName: "missing-tagged-template-type-argument.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
					{MessageId: "undef", Line: 1, Column: 5},
				},
			},
			{
				Code:     `new Ctor<MissingType>();`,
				FileName: "missing-new-type-argument.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 5},
					{MessageId: "undef", Line: 1, Column: 10},
				},
			},
			{
				Code:     `type T = [label: Missing, optional?: Missing2];`,
				FileName: "missing-named-tuple-types.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 18},
					{MessageId: "undef", Line: 1, Column: 38},
				},
			},
			{
				Code:     `type T<X> = X extends infer U ? U | MissingTrue : MissingFalse;`,
				FileName: "missing-conditional-type-references.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 37},
					{MessageId: "undef", Line: 1, Column: 51},
				},
			},
			{
				Code:     `import X = MissingNS.Member; X;`,
				FileName: "missing-import-equals-root.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 12},
				},
			},
			{
				Code:     `export type { Missing as Alias };`,
				FileName: "missing-type-export.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 15},
				},
			},
			{
				Code:     `namespace N { export type T = Missing; }`,
				FileName: "missing-namespace-type.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 31},
				},
			},
			{
				Code:     `const Existing = 1; type T = Existing;`,
				FileName: "value-does-not-define-type.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 30},
				},
			},
			{
				Code:     `interface Existing {} type T = typeof Existing;`,
				FileName: "type-does-not-define-value.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 39},
				},
			},
			{
				Code:     `/** @type {Missing} */ const x = {}; Missing;`,
				FileName: "jsdoc-runtime-reference.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 38},
				},
			},
			{
				Code:     "/** @typedef {object} Foo */\nFoo;",
				FileName: "jsdoc-typedef-runtime.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 1},
				},
			},
			{
				Code:     "/** @typedef {object} Foo */\nexport { Foo };",
				FileName: "jsdoc-typedef-export.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 2, Column: 10},
				},
			},
		},
	)
}

func TestNoUndefLanguageDefaults(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.allow-js.json",
		t,
		&NoUndefRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     `exports; global; module; require; arguments;`,
				FileName: "file.cjs",
			},
			{
				Code:     `(() => arguments)();`,
				FileName: "arrow.cjs",
			},
			{
				Code:     `arguments;`,
				FileName: "arguments-off.cjs",
				Globals:  map[string]any{"arguments": "off"},
			},
			{
				Code:     "/** @import { arguments } from \"pkg\" */\narguments;",
				FileName: "jsdoc-arguments-off.cjs",
				TSConfig: "tsconfig.allowJs.json",
				Globals:  map[string]any{"arguments": "off"},
			},
			{
				Code:     `const require = 1; require;`,
				FileName: "local-require.cjs",
				Globals:  map[string]any{"require": "off"},
			},
			{
				Code:            `require('x'); global;`,
				FileName:        "a.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			{
				Code:            `arguments;`,
				FileName:        "a.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			{
				Code:            `require;`,
				FileName:        "authored-commonjs.ts",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			{
				Code:            `arguments;`,
				FileName:        "a.jsx",
				TSConfig:        "tsconfig.allow-js.json",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `require;`,
				FileName: "plain-script.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:            `require('x'); global; arguments;`,
				FileName:        "a.cjs",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
					{MessageId: "undef", Line: 1, Column: 15},
					{MessageId: "undef", Line: 1, Column: 23},
				},
			},
			{
				Code:     `arguments;`,
				FileName: "module-file.mjs",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:     "/** @import { arguments } from \"pkg\" */\narguments;",
				FileName: "jsdoc-arguments-off.mjs",
				TSConfig: "tsconfig.allowJs.json",
				Globals:  map[string]any{"arguments": "off"},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 2, Column: 1}},
			},
			{
				Code:     `require;`,
				FileName: "typed-commonjs.cts",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:            `arguments;`,
				FileName:        "authored-commonjs.ts",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:     `process;`,
				FileName: "node-name.cjs",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:     `__dirname;`,
				FileName: "node-path.cjs",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:     `require;`,
				FileName: "require-off.cjs",
				Globals:  map[string]any{"require": "off"},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
		},
	)
}
