package no_shadow_restricted_names

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoShadowRestrictedNamesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoShadowRestrictedNamesRule,
		// Valid cases — ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `function foo(bar){ var baz; }`},
			{Code: `!function foo(bar){ var baz; }`},
			{Code: `!function(bar){ var baz; }`},
			{Code: `try {} catch(e) {}`},
			{Code: `export default function() {}`},
			{Code: `try {} catch {}`},
			{Code: `var undefined;`},
			{Code: `var undefined; doSomething(undefined);`},
			{Code: `var undefined; var undefined;`},
			{Code: `let undefined`},
			{Code: `import { undefined as undef } from 'foo';`},
			{
				Code:    `let globalThis;`,
				Options: map[string]interface{}{"reportGlobalThis": false},
			},
			{
				Code:    `class globalThis {}`,
				Options: map[string]interface{}{"reportGlobalThis": false},
			},
			{
				Code:    `import { baz as globalThis } from 'foo';`,
				Options: map[string]interface{}{"reportGlobalThis": false},
			},
			{Code: `globalThis.foo`},
			{Code: `const foo = globalThis`},
			{Code: `function foo() { return globalThis; }`},
			{Code: `import { globalThis as foo } from 'bar'`},
		},
		// Invalid cases — ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `function NaN(NaN) { var NaN; !function NaN(NaN) { try {} catch(NaN) {} }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Message: "Shadowing of global property 'NaN'.", Line: 1, Column: 10},
					{MessageId: "shadowingRestrictedName", Message: "Shadowing of global property 'NaN'.", Line: 1, Column: 14},
					{MessageId: "shadowingRestrictedName", Message: "Shadowing of global property 'NaN'.", Line: 1, Column: 25},
					{MessageId: "shadowingRestrictedName", Message: "Shadowing of global property 'NaN'.", Line: 1, Column: 40},
					{MessageId: "shadowingRestrictedName", Message: "Shadowing of global property 'NaN'.", Line: 1, Column: 44},
					{MessageId: "shadowingRestrictedName", Message: "Shadowing of global property 'NaN'.", Line: 1, Column: 64},
				},
			},
			{
				Code: `function undefined(undefined) { !function undefined(undefined) { try {} catch(undefined) {} }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 20},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 43},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 53},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 79},
				},
			},
			{
				Code: `function Infinity(Infinity) { var Infinity; !function Infinity(Infinity) { try {} catch(Infinity) {} }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 19},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 35},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 55},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 64},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 89},
				},
			},
			{
				Code: `function arguments(arguments) { var arguments; !function arguments(arguments) { try {} catch(arguments) {} }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 20},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 37},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 58},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 68},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 94},
				},
			},
			{
				Code: `function eval(eval) { var eval; !function eval(eval) { try {} catch(eval) {} }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 15},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 27},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 43},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 48},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 69},
				},
			},
			{
				Code: `var eval = (eval) => { var eval; !function eval(eval) { try {} catch(eval) {} }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 28},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 44},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 49},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 70},
				},
			},
			{
				Code: `var [undefined] = [1]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 6},
				},
			},
			{
				Code: `var {undefined} = obj; var {a: undefined} = obj; var {a: {b: {undefined}}} = obj; var {a, ...undefined} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 6},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 32},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 63},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 94},
				},
			},
			{
				Code: `var undefined; undefined = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `class undefined {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 7},
				},
			},
			{
				Code: `(class undefined {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},
			{
				Code: `import undefined from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},
			{
				Code: `import { undefined } from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
				},
			},
			{
				Code: `import { baz as undefined } from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 17},
				},
			},
			{
				Code: `import * as undefined from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},
			{
				Code: `function globalThis(globalThis) { var globalThis; !function globalThis(globalThis) { try {} catch(globalThis) {} }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 21},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 39},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 61},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 72},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 99},
				},
			},
			{
				Code: `const [globalThis] = [1]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},
			{
				Code: `var {globalThis} = obj; var {a: globalThis} = obj; var {a: {b: {globalThis}}} = obj; var {a, ...globalThis} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 6},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 33},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 65},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 97},
				},
			},
			{
				Code: `let globalThis; globalThis = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `class globalThis {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 7},
				},
			},
			{
				Code: `(class globalThis {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},
			{
				Code: `import globalThis from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},
			{
				Code: `import { globalThis } from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
				},
			},
			{
				Code: `import { baz as globalThis } from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 17},
				},
			},
			{
				Code: `import * as globalThis from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},
		},
	)
}

// Exhaustive coverage beyond the upstream ESLint test file, grouped by category.
// Each invalid case asserts the binding identifier's Line/Column so that any
// AST-shape or position regression is caught.
func TestNoShadowRestrictedNamesExtended(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoShadowRestrictedNamesRule,
		[]rule_tester.ValidTestCase{
			// ---- Safe var/let undefined (no writes anywhere) ----
			{Code: `var undefined;`},
			{Code: `let undefined;`},
			{Code: `var undefined; var undefined;`},
			{Code: `var undefined; doSomething(undefined);`},
			{Code: `var undefined; typeof undefined;`},
			{Code: `var undefined; undefined === null;`},
			{Code: `var undefined; delete undefined;`},
			{Code: `function f() { var undefined; return undefined; }`},
			{Code: `var undefined; function g() { var undefined; }`},
			// Outer write to global undefined; inner `var undefined` has distinct symbol.
			{Code: `function f() { var undefined; } undefined = 5;`},

			// ---- Non-binding contexts (property keys, member access) ----
			{Code: `({undefined: 1})`},
			{Code: `({undefined: 1, NaN: 2, Infinity: 3})`},
			{Code: `({[undefined]: 1})`},
			{Code: `obj.undefined`},
			{Code: `obj['undefined']`},
			{Code: `class C { undefined = 1; }`},
			{Code: `class C { undefined() {} }`},
			{Code: `class C { static undefined() {} }`},
			{Code: `class C { get undefined() { return 1; } }`},
			{Code: `class C { set undefined(v) {} }`},
			{Code: `class C { undefined: number = 1; }`},
			{Code: `({undefined() {}})`},
			{Code: `({get undefined() { return 1; }})`},
			{Code: `({set undefined(v) {}})`},

			// ---- Exports that are NOT shadowing declarations ----
			{Code: `export { foo as undefined }`, Tsx: false},
			{Code: `export default undefined;`},
			{Code: `import { globalThis as foo } from 'bar'`},
			{Code: `import { undefined as undef } from 'foo';`},

			// ---- Read references only ----
			{Code: `function foo() { return undefined; }`},
			{Code: `function foo() { return NaN; }`},
			{Code: `var x = Infinity;`},
			{Code: `typeof arguments;`},
			{Code: `function f() { return arguments; }`},

			// ---- globalThis read-only usages ----
			{Code: `globalThis.foo`},
			{Code: `const foo = globalThis;`},

			// ---- reportGlobalThis: false ----
			{
				Code:    `function globalThis(globalThis) { var globalThis; }`,
				Options: map[string]interface{}{"reportGlobalThis": false},
			},
			{
				Code:    `class globalThis {}`,
				Options: map[string]interface{}{"reportGlobalThis": false},
			},
			{
				Code:    `import globalThis from 'm'`,
				Options: map[string]interface{}{"reportGlobalThis": false},
			},
			{
				Code:    `for (var globalThis of arr) {}`,
				Options: map[string]interface{}{"reportGlobalThis": false},
			},
			// Other restricted names still reported even with the option off.

			// ---- TypeScript-only constructs that are NOT in the rule's listener set ----
			{Code: `enum undefined { A, B }`},
			{Code: `enum NaN { A, B }`},
			{Code: `interface undefined {}`},
			{Code: `interface NaN {}`},
			{Code: `type undefined = number`},
			{Code: `type NaN = number`},
			{Code: `namespace undefined {}`},
			{Code: `namespace NaN { export var x = 1; }`},
			{Code: `declare module 'undefined' {}`},
			{Code: `import undefined = require('m')`},
			{Code: `function f<T>(x: T): T { return x; }`},

			// ---- Type-position usages of restricted names (type reference, not binding) ----
			{Code: `function f(this: void) {}`},
			{Code: `const x: NaN = 1 as any;`},
			{Code: `type Foo = typeof undefined;`},

			// ---- Parameters in TYPE positions are not runtime bindings; ESLint's
			//      :function does not match these. They must NOT be reported. ----
			{Code: `type F = (undefined: number) => void;`},
			{Code: `type F = (NaN: number, Infinity: number) => void;`},
			{Code: `interface I { method(undefined: number): void }`},
			{Code: `interface I { (undefined: number): void }`},
			{Code: `interface I { new (undefined: number): any }`},
			{Code: `let x: { (undefined: number): void };`},
			{Code: `let x: { method(undefined: number): void };`},
			{Code: `type Ctor = new (undefined: number) => any;`},
			{Code: `interface I { [undefined: string]: any }`},

			// ---- Default values reference the global undefined (read, not shadow) ----
			{Code: `function foo(a = undefined) { return a; }`},
			{Code: `function foo(a = NaN) { return a; }`},
			{Code: `var x = undefined;`},

			// ---- Nested shadowed ----
			{Code: `var undefined; { let undefined; }`},
			{Code: `var undefined; function inner() { var undefined; }`},

			// ---- try/catch with no binding ----
			{Code: `try {} catch {}`},
			{Code: `try {} catch(e) { e; }`},

			// ---- Empty / anonymous ----
			{Code: `export default function() {}`},
			{Code: `export default class {}`},
			{Code: `(function() {})();`},
			{Code: `(() => {})()`},

			// ---- Private class fields use '#' prefix, distinct name ----
			{Code: `class C { #undefined = 1; m() { return this.#undefined; } }`},

			// ---- Static members with restricted names are property names, not bindings ----
			{Code: `class C { static undefined = 1; }`},
			{Code: `class C { static #undefined = 1; }`},
			{Code: `class C { static get undefined() { return 1; } }`},
			{Code: `class C { static set undefined(v) {} }`},
			{Code: `class C { static { let x = 1; } }`},

			// ---- Computed keys (read reference only) ----
			{Code: `class C { [undefined]() {} }`},
			{Code: `({[undefined]() {}})`},
			{Code: `({['undefined']() {}})`},
			{Code: `class C { [NaN]: number = 1; }`},

			// ---- Type parameters (generics) are type-level, not runtime bindings ----
			{Code: `function f<undefined>(x: any): any { return x; }`},
			{Code: `function f<NaN, Infinity>(): void {}`},
			{Code: `class C<undefined> {}`},
			{Code: `interface I<undefined> {}`},
			{Code: `type T<undefined> = any`},

			// ---- Ambient declarations without initializers / writes ----
			{Code: `declare var undefined: any;`},
			{Code: `declare const undefined: any;`},
			{Code: `declare let undefined: any;`},

			// ---- Re-export specifiers (no new local binding) ----
			{Code: `export { foo as NaN } from 'm';`},
			{Code: `export { undefined } from 'm';`},
			{Code: `export * as undefined from 'm';`},

			// ---- typeof / spread / template (reads only) ----
			{Code: "var x = [...undefined];"},
			{Code: "var x = {...undefined};"},
			{Code: "f(...undefined)"},
			{Code: "var x = `${undefined}`;"},
			{Code: "var x = tag`${undefined}`;"},

			// ---- Ternary / switch reads ----
			{Code: `function f(x) { return x === undefined ? NaN : Infinity; }`},
			{Code: `switch (x) { case undefined: break; case NaN: break; }`},

			// ---- Optional chaining as member access, not a binding ----
			{Code: `obj?.undefined`},
			{Code: `obj?.['undefined']`},

			// ---- Labels named like restricted identifiers are not bindings ----
			{Code: `undefined: for (;;) { break undefined; }`},

			// ---- Assignments to global restricted names (no new binding) ----
			{Code: `NaN = 5;`},
			{Code: `({NaN} = {})`},
			{Code: `[NaN] = [1]`},

			// ---- Read-only inside class field initializers ----
			{Code: `class C { x = undefined; y = NaN; z = Infinity; }`},

			// ---- Destructuring with computed key reading restricted (not binding) ----
			{Code: `var {[undefined]: x} = obj;`},

			// ---- Parameter typed as `undefined` (the param name, not shadow) ----
			{Code: `function f(x: undefined) {}`},
			{Code: `function f(x: NaN) {}`},

			// ---- Destructuring with default values reading restricted (read, not shadow) ----
			{Code: `function f({x = undefined}) {}`},
			{Code: `function f({x = NaN}) {}`},
			{Code: `function f([x = undefined]) {}`},
			{Code: `const {a = undefined} = obj;`},
			{Code: `const [b = NaN] = arr;`},

			// ---- Abstract accessor name (not a binding) ----
			{Code: `abstract class C { abstract get undefined(): number; }`},
			{Code: `abstract class C { abstract set undefined(v: number); }`},
			{Code: `abstract class C { abstract undefined(): void; }`},

			// ---- Class with index signature (type-level, not binding) ----
			{Code: `class C { [key: string]: any }`},
			{Code: `interface I { [undefined: string]: any }`},

			// ---- Function overloads (multiple declarations merge) ----
			{Code: `declare function f(x: number): number; declare function f(x: string): string;`},

			// ---- Global augmentation (type-level) ----
			{Code: `declare global { var globalThis: any; }`, Options: map[string]interface{}{"reportGlobalThis": false}},

			// ---- Dynamic import destructure uses read context for the module namespace ----
			{Code: `async function f() { const { x } = await import('m'); }`},

			// ---- Default-valued destructured parameter renamed from a restricted key ----
			{Code: `function f({undefined: x = 5}) {}`},
			{Code: `function f({NaN: n}) {}`},

			// ---- Rest param as plain identifier with safe name ----
			{Code: `function f(...args) {}`},
			{Code: `const f = (...xs) => xs;`},

			// ---- Readonly/public/private/protected modifiers do not bind restricted names ----
			{Code: `class C { constructor(readonly x: number) {} }`},
			{Code: `class C { constructor(private x: number, public y: number) {} }`},

			// ---- Method shorthand with restricted param name lives in type-only object type ----
			{Code: `let o: { m(undefined: any): void };`},

			// ---- Function body that only READS restricted identifiers ----
			{Code: `function f() { if (x === undefined) return NaN; return Infinity; }`},

			// ---- `as` casts and non-null assertions on read references ----
			{Code: `var x = undefined as any;`},
			{Code: `var x = (undefined)!;`},

			// ---- Computed class method body is read-only ----
			{Code: `class C { [Symbol.iterator]() { return undefined; } }`},

			// ---- In module / export-default expression forms ----
			{Code: `export default undefined;`},
			{Code: `export default NaN;`},
			{Code: `export default function() {}`},

			// ---- Nested namespace with safely-scoped undefined binding (no writes) ----
			{Code: `namespace N { var undefined; }`},

			// ---- React-like patterns with safe (non-restricted) names ----
			{Code: `function Component({ data, loading, error }) { return null; }`},
			{Code: `const handler = ({ type, payload }) => payload;`},

			// ---- async iteration reads only ----
			{Code: `async function f() { for await (const x of asyncIter()) { console.log(x, undefined); } }`},

			// ---- labeled statement with break/continue referencing label name ----
			{Code: `NaN: for (;;) { break NaN; continue NaN; }`},

			// ---- chain of optional member accesses reading restricted names ----
			{Code: `const x = obj?.NaN?.Infinity?.undefined;`},

			// ---- catch body mutating non-restricted name ----
			{Code: `try {} catch(e) { e = null; }`},

			// ---- jsx-like function component returning primitives ----
			{Code: `const C = () => undefined;`},

			// ---- Conditional / mapped / infer types (type-level, not runtime bindings) ----
			{Code: `type T<K> = { [undefined in K]: 1 };`},
			{Code: `type T<X> = X extends infer undefined ? 1 : 2;`},
			{Code: `type T<U> = U extends infer R ? R : never;`},

			// ---- Class implements clause with restricted name (type reference, not shadow) ----
			{Code: `class C implements undefined {}`},
			{Code: `class C implements I, undefined {}`},

			// ---- Function return type annotation is a type reference, not shadow ----
			{Code: `function f(): undefined { return undefined; }`},
			{Code: `const f = (): undefined => undefined;`},

			// ---- class / function expressions in extends position reading names, not declaring ----
			{Code: `class C extends Mixin(Base) {}`},

			// ---- async iterator / generator bodies with read-only references ----
			{Code: `async function* g() { yield undefined; yield NaN; yield Infinity; }`},

			// ---- Redux-like reducer default reads global undefined ----
			{Code: `const reducer = (state = undefined, action) => state;`},

			// ---- Express-like request handler using non-restricted names ----
			{Code: `app.get('/', (req, res, next) => res.send());`},

			// ---- Nested mapped type keys ----
			{Code: `type T = { [K in 'a' | 'b']: K extends 'a' ? undefined : NaN };`},

			// ---- Destructure with type annotation ----
			{Code: `const {a}: {a: number} = x;`},
			{Code: `const [a]: [number] = x;`},

			// ---- Array of function types ----
			{Code: `const arr: ((undefined: any) => void)[] = [];`},
			{Code: `const arr: Array<(NaN: any) => void> = [];`},

			// ---- Return type annotations (read-only type positions) ----
			{Code: `function f(): NaN { return null as any; }`},
			{Code: `const f = (): undefined => undefined;`},

			// ---- Type intersection / union references ----
			{Code: `type T = undefined | null;`},
			{Code: `type T = NaN & Infinity;`},

			// ---- Conditional types with restricted type reference ----
			{Code: `type T = X extends undefined ? null : X;`},

			// ---- Method with restricted name in interface (type-level) ----
			{Code: `interface I { NaN(): void; undefined(x: number): void; }`},

			// ---- Class field + same-named method (both are properties, not bindings) ----
			{Code: `class C { undefined: number = 1; undefined() {} }`},

			// ---- Paired accessors (both property names) ----
			{Code: `class C { get undefined() { return 1; } set undefined(v) {} }`},

			// ---- Type query in generic default ----
			{Code: `function f<T = undefined>(): void {}`},
			{Code: `function f<T extends undefined>(): void {}`},

			// ---- Nested function types in generic constraint ----
			{Code: `function f<T extends (undefined: number) => void>(): void {}`},

			// ---- Object method shorthand where property key is restricted (not a binding) ----
			{Code: `const o = { undefined: 1, NaN: 2, Infinity: 3, eval: 4, arguments: 5 };`},
			{Code: `const o = { ['undefined'](){}, ['NaN'](){} };`},

			// ---- Tuple with restricted type references ----
			{Code: `type T = [undefined, NaN, Infinity];`},

			// ---- JSDoc-style typedef inside comment (ignored, pure comment) ----
			{Code: "/** @type {any} */\nvar x;"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- 1. Function name: async / generator / async generator / export / default ----
			{
				Code: `async function NaN() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 16},
				},
			},
			{
				Code: `function* NaN() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 11},
				},
			},
			{
				Code: `async function* NaN() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 17},
				},
			},
			{
				Code: `export function NaN() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 17},
				},
			},
			{
				Code: `export default function NaN() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 25},
				},
			},

			// ---- 2. Parameter variations: rest / default / destructured / deep ----
			{
				Code: `function f(a, b, NaN) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 18},
				},
			},
			{
				Code: `function f(...NaN) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 15},
				},
			},
			{
				Code: `function f(NaN = 5) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
				},
			},
			{
				Code: `function f({NaN}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},
			{
				Code: `function f({NaN = 5}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},
			{
				Code: `function f({a: NaN}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 16},
				},
			},
			{
				Code: `function f([NaN]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},
			{
				Code: `function f({a, ...NaN}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 19},
				},
			},
			{
				Code: `function f({a: {b: [NaN]}}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 21},
				},
			},

			// ---- 3. Arrow function parameters ----
			{
				Code: `const f = (NaN) => NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
				},
			},
			{
				Code: `const f = NaN => NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 11},
				},
			},
			{
				Code: `const f = async (NaN) => NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 18},
				},
			},
			{
				Code: `const f = ({NaN}) => NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},

			// ---- 4. Method / constructor / accessor params (via inner FunctionExpression in ESLint; Parameter listener here) ----
			{
				Code: `class C { m(NaN) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},
			{
				Code: `class C { constructor(NaN) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 23},
				},
			},
			{
				Code: `class C { set p(NaN) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 17},
				},
			},
			{
				Code: `({m(NaN) {}})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `class C { constructor(public NaN: number) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 30},
				},
			},

			// ---- 5. Variable declarations: let / const / multi-declarator ----
			{
				Code: `let NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `const NaN = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 7},
				},
			},
			{
				Code: `var a, NaN, b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},
			{
				Code: `var NaN, Infinity;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
				},
			},
			{
				Code: `const undefined = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 7},
				},
			},
			{
				Code: `var undefined, undefined = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					// Both declarators merge into one variable in ESLint; any def with
					// init makes the whole variable unsafe, so both are reported.
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 16},
				},
			},

			// ---- 6. for-in / for-of loop-variable declarations (loop writes every iteration) ----
			{
				Code: `for (var undefined in obj) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
				},
			},
			{
				Code: `for (var undefined of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
				},
			},
			{
				Code: `for (let undefined of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
				},
			},
			{
				Code: `for (const undefined of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
				},
			},
			{
				Code: `for (var undefined = 0;;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
				},
			},
			{
				Code: `for (var NaN of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
				},
			},

			// ---- 7. Writes in various contexts make `var undefined` unsafe ----
			{
				Code: `var undefined; undefined = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; undefined++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; ++undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; undefined += 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; [undefined] = [1];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; ({undefined} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; for (undefined in obj) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; for (undefined of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; (undefined) = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			// Cross-scope write resolves to the outer symbol via TypeChecker.
			{
				Code: `var undefined; function inner() { undefined = 5; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 8. Catch with destructured bindings ----
			{
				Code: `try {} catch({NaN}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 15},
				},
			},
			{
				Code: `try {} catch([NaN]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 15},
				},
			},
			{
				Code: `try {} catch({a: NaN}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 18},
				},
			},
			{
				Code: `try {} catch({a: {b: NaN}}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 22},
				},
			},
			{
				Code: `try {} catch({...NaN}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 18},
				},
			},

			// ---- 9. Import combinations ----
			{
				Code: `import undefined, { x } from 'm';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},
			{
				Code: `import undefined, * as x from 'm';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},
			{
				Code: `import d, { NaN, Infinity as Inf } from 'm';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},
			{
				Code: `import { x as NaN, y as Infinity } from 'm';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 15},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 25},
				},
			},

			// ---- 10. Class declarations / expressions ----
			{
				Code: `class NaN extends Foo {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 7},
				},
			},
			{
				Code: `(class NaN extends Foo {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},
			{
				Code: `foo(class NaN {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 11},
				},
			},
			{
				Code: `export class NaN {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 14},
				},
			},
			{
				Code: `export default class NaN {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 22},
				},
			},

			// ---- 11. Deep destructuring in var ----
			{
				Code: `var [[[undefined]]] = x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},
			{
				Code: `var {a: {b: [{c: undefined}]}} = x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 18},
				},
			},
			{
				Code: `var [a, undefined, b] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 9},
				},
			},
			{
				Code: `var [undefined = 5] = [];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 6},
				},
			},
			{
				Code: `var {a: undefined = 5} = {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 9},
				},
			},

			// ---- 12. All 5 core restricted names in the same file ----
			{
				Code: `var NaN, Infinity, undefined = 0, eval, arguments;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5, Message: "Shadowing of global property 'NaN'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10, Message: "Shadowing of global property 'Infinity'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 20, Message: "Shadowing of global property 'undefined'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 35, Message: "Shadowing of global property 'eval'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 41, Message: "Shadowing of global property 'arguments'."},
				},
			},

			// ---- 13. Nested scope write makes outer unsafe ----
			{
				Code: `function outer() { var undefined; function inner() { undefined = 5; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 24},
				},
			},

			// ---- 14. reportGlobalThis default is true ----
			{
				Code: `let globalThis;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5, Message: "Shadowing of global property 'globalThis'."},
				},
			},
			{
				Code:    `let globalThis;`,
				Options: map[string]interface{}{"reportGlobalThis": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 15. reportGlobalThis: false still reports other restricted names ----
			{
				Code:    `function globalThis(NaN) {}`,
				Options: map[string]interface{}{"reportGlobalThis": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 21, Message: "Shadowing of global property 'NaN'."},
				},
			},

			// ---- 16. Multi-line positions ----
			{
				Code: "function outer() {\n  var undefined = 5;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 2, Column: 7},
				},
			},
			{
				Code: "class C {\n  m(NaN) {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 2, Column: 5},
				},
			},

			// ---- 17. typeof/delete/read do NOT make a simple var-undefined unsafe ----
			// (negative is in the valid block; here we confirm assignment via typeof-like
			// boundary constructs does make it unsafe)
			{
				Code: `var undefined; if (true) { undefined = 5; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 18. Named default-export function still named "undefined" ----
			{
				Code: `export default function undefined() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 25},
				},
			},

			// ---- 19. Named class expression on right-hand side of assignment ----
			{
				Code: `let C = class undefined {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 15},
				},
			},

			// ---- 20. Using declarations (TS / stage-3) still match VariableDeclaration ----
			{
				Code: `{ using undefined = {[Symbol.dispose]() {}} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 9},
				},
			},

			// ---- 21. Multiple restricted names in a single listener visit ----
			{
				Code: `function f(NaN, Infinity) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 17},
				},
			},
			{
				Code: `function f({a: NaN, b: Infinity}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 16},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 24},
				},
			},
			{
				Code: `import { NaN, Infinity, undefined } from 'm';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 15},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 25},
				},
			},

			// ---- 22. TypeScript type-only imports still produce a runtime-looking binding in ImportDeclaration ----
			{
				Code: `import type undefined from 'm';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},
			{
				Code: `import { type undefined } from 'm';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 15},
				},
			},

			// ---- 23. Writes inside catch/try bodies make outer var undefined unsafe ----
			{
				Code: `var undefined; try {} catch(e) { undefined = 5; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; try { undefined = 5; } catch(e) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 24. Named function expression used as default parameter value ----
			{
				Code: `function f(cb = function undefined() {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 26},
				},
			},

			// ---- 25. Multiple restricted names in one class/function container ----
			{
				Code: `class NaN { constructor(Infinity, undefined) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 7},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 25},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 35},
				},
			},

			// ---- 26. globalThis nested in destructure ----
			{
				Code: `var {a: {globalThis}} = x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
				},
			},

			// ---- 27. Common JSX-like / React-style (real-world) pattern ----
			{
				Code: `function Component({ undefined, ...props }) { return props; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 22},
				},
			},
			// Promise-style callback
			{
				Code: `new Promise((resolve, undefined) => { resolve(1); })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 23},
				},
			},

			// ---- 28. for-await-of with declaration ----
			{
				Code: `async function f() { for await (var undefined of arr) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 37},
				},
			},
			{
				Code: `async function f() { for await (const undefined of arr) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 39},
				},
			},

			// ---- 29. Async generator / generator function expressions ----
			{
				Code: `const f = async function* undefined() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 27},
				},
			},
			{
				Code: `const f = function* undefined() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 21},
				},
			},

			// ---- 30. IIFE with named function expression ----
			{
				Code: `(function undefined() {})();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 11},
				},
			},
			{
				Code: `!function undefined() {}();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 11},
				},
			},

			// ---- 31. new ClassExpression ----
			{
				Code: `new (class undefined {})();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
				},
			},

			// ---- 32. Ambient declarations that DO map to a runtime binding ----
			{
				Code: `declare function undefined(): void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 18},
				},
			},
			{
				Code: `declare class undefined {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 15},
				},
			},

			// ---- 33. Decorated class / parameters ----
			{
				Code: `@dec class undefined {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
				},
			},
			{
				Code: `class C { @dec method(undefined: any) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 23},
				},
			},
			{
				Code: `class C { constructor(@dec undefined: any) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 28},
				},
			},

			// ---- 34. Abstract methods (runtime function-like, no body) ----
			{
				Code: `abstract class C { abstract m(undefined: any): void; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 31},
				},
			},

			// ---- 35. Optional / typed parameters still report ----
			{
				Code: `function f(undefined?: number): void {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
				},
			},
			{
				Code: `function f(undefined: number = 5) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
				},
			},

			// ---- 36. satisfies / as const / type-assertion initializers still count as inits ----
			{
				Code: `const undefined = 5 as const;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 7},
				},
			},
			{
				Code: `const undefined = {} satisfies object;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 7},
				},
			},

			// ---- 37. class static block / field write makes outer var unsafe ----
			{
				Code: `var undefined; class C { static { undefined = 5; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; class C { x = (undefined = 5); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 38. Class field initializer creating a nested named class expression ----
			{
				Code: `class C { x = class undefined {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 21},
				},
			},
			{
				Code: `function f() { return class undefined {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 29},
				},
			},

			// ---- 39. Ternary / logical expressions with named class expression ----
			{
				Code: `const X = true ? class undefined {} : null;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 24},
				},
			},

			// ---- 40. Multiple errors concentrated in one function-like ----
			{
				Code: `function undefined(NaN, Infinity) { var arguments; var eval; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10, Message: "Shadowing of global property 'undefined'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 20, Message: "Shadowing of global property 'NaN'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 25, Message: "Shadowing of global property 'Infinity'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 41, Message: "Shadowing of global property 'arguments'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 56, Message: "Shadowing of global property 'eval'."},
				},
			},

			// ---- 41. Reserved-name shadowing via destructuring assignment (no declaration; only writes the existing global / var) ----
			// A bare destructuring assignment to globals is NOT shadowing (no new
			// binding introduced). It must not be reported on its own.
			// (This is the negative of the below "with outer var undefined" case.)
			// Outer var undefined + later destructuring assignment write → unsafe.
			{
				Code: `var undefined; [,, undefined] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 42. Write inside arrow body inside parameter default ----
			{
				Code: `var undefined; function f(x = () => undefined = 5) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 43. Empty-pattern default param with undefined: still reports
			//         the binding, defaults don't suppress the name check. ----
			{
				Code: `function f({} = undefined) { var undefined = 5; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 34},
				},
			},

			// ---- 44. Cross-file-like: writes inside nested classes ----
			{
				Code: `var undefined; class Outer { m() { class Inner { n() { undefined = 5; } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 45. Object shorthand in catch nested destructure ----
			{
				Code: `try {} catch({a: {b: [undefined]}}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 23},
				},
			},

			// ---- 46. Mixed multi-declarator: one has init, another is a safe undefined ----
			// The safe undefined must remain unreported even when siblings have inits.
			// Each sibling declarator is a separate tsgo VariableDeclaration / ESLint
			// VariableDeclarator, so cross-declarator "all-safe" is not a thing to worry
			// about here — but this test locks it in. The `NaN` sibling is always reported.
			{
				Code: `var a = 1, NaN, b = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
				},
			},

			// ---- 47. Array destructuring with holes ----
			{
				Code: `var [, undefined, , NaN] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 21},
				},
			},

			// ---- 48. JSX destructured prop (React pattern) ----
			{
				Code: `const C = ({undefined}) => null;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},

			// ---- 49. Function declaration nested inside method body ----
			{
				Code: `class C { m() { function undefined() {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 26},
				},
			},

			// ---- 50. Nested FE with restricted name inside an assignment RHS ----
			{
				Code: `const x = (function undefined() { return 1; })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 21},
				},
			},

			// ---- 51. Two `var undefined` with one having an initializer ----
			// ESLint merges both defs into one variable; any def with init poisons
			// the shared variable, reporting every def. rslint achieves the same via
			// symbol-level write tracking.
			{
				Code: `var undefined; var undefined = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 20},
				},
			},

			// ---- 52. Multi-init for-loop with restricted name among sibling declarators ----
			{
				Code: `for (var i = 0, undefined = 0; i < 10; i++) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 17},
				},
			},
			{
				Code: `for (let i = 0, NaN = 0; i < 10; i++) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 17},
				},
			},

			// ---- 53. TS parameter property with readonly + restricted name ----
			{
				Code: `class C { constructor(readonly undefined: number) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 32},
				},
			},
			{
				Code: `class C { constructor(protected NaN: number, public Infinity: number) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 33},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 53},
				},
			},

			// ---- 54. Renamed destructured binding where the new local name is restricted ----
			{
				Code: `const {x: undefined = 5} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 11},
				},
			},
			{
				Code: `const [[NaN = 0]] = [[]];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 9},
				},
			},

			// ---- 55. Mixin pattern — named class expression inside an arrow / factory ----
			{
				Code: `const Mixin = (Base) => class undefined extends Base {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 31},
				},
			},

			// ---- 56. Async arrow with destructured param ----
			{
				Code: `const f = async ({NaN}) => NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 19},
				},
			},

			// ---- 57. Rest binding in catch (9+ stage) ----
			{
				Code: `try {} catch({...undefined}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 18},
				},
			},

			// ---- 58. Chained destructuring assignment affecting outer var undefined ----
			{
				Code: `var undefined; ({a: undefined} = {a: 1});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; ({a: {b: undefined}} = {a: {b: 1}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; [[undefined]] = [[1]];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 59. Nullish-coalescing / logical-assignment writes ----
			{
				Code: `var undefined; undefined ||= 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; undefined ??= 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; undefined &&= 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 60. Class with named class expression inside method returning generator ----
			{
				Code: `class C { *m() { yield class undefined {}; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 30},
				},
			},

			// ---- 61. Decorator factory with restricted parameter ----
			{
				Code: `function dec(NaN: any) { return () => {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 14},
				},
			},

			// ---- 62. Namespace body var with initializer ----
			{
				Code: `namespace N { var undefined = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 19},
				},
			},

			// ---- 63. TS overloaded function declarations (each signature reports) ----
			{
				Code: `function undefined(x: number): number; function undefined(x: string): string; function undefined(x: any) { return x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 49},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 88},
				},
			},

			// ---- 64. Method names are NOT bindings — confirm negative even with async/static combos ----
			// (See valid block for the positives; here we ensure no false positive when
			// both method name AND nested element look restricted.)
			{
				Code: `class C { async NaN() { return function NaN() {}; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					// Only the nested named FE reports; the method name itself is a property key.
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 41},
				},
			},

			// ---- 65. Destructuring default reads don't suppress shadow on the binding ----
			{
				Code: `function f({undefined = NaN}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					// `undefined` is the binding name (reported); `NaN` is a read reference (not reported).
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},

			// ---- 66. Nested arrow / method closure writes ----
			{
				Code: `var undefined; class C { m() { const f = () => { undefined = 5; }; f(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 67. for-of with destructured variable binding ----
			{
				Code: `for (var [undefined] of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 11},
				},
			},
			{
				Code: `for (const {undefined} of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},
			{
				Code: `for (let {a: NaN} of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 14},
				},
			},

			// ---- 68. Class extends a named class expression with restricted name ----
			{
				Code: `class C extends class undefined {} {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 23},
				},
			},

			// ---- 69. Exported variable declarations ----
			{
				Code: `export const undefined = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 14},
				},
			},
			{
				Code: `export var NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
				},
			},
			{
				Code: `export let {NaN} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},

			// ---- 70. Global augmentation with a runtime declaration ----
			{
				Code: `declare global { function undefined(): void; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 27},
				},
			},
			{
				Code: `declare global { class NaN {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 24},
				},
			},

			// ---- 71. Shadowing that is both param-name AND inner FD-name ----
			{
				Code: `function f(undefined) { function undefined() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 34},
				},
			},

			// ---- 72. Dynamic import destructure binding restricted name ----
			{
				Code: `async function f() { const { undefined } = await import('m'); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 30},
				},
			},

			// ---- 73. Regression: parenthesized write with type assertion ----
			{
				Code: `var undefined; (undefined as any) = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; (<any>undefined) = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; undefined! = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 74. Destructure with type annotation whose local name is restricted ----
			{
				Code: `const {a: undefined}: {a: number} = x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 11},
				},
			},
			{
				Code: `const [undefined]: [number] = x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 8},
				},
			},

			// ---- 75. Object method value is a named function expression ----
			{
				Code: `const o = { m: function undefined() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 25},
				},
			},

			// ---- 76. Function declaration hoisted inside if-block body ----
			{
				Code: `if (true) { function undefined() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 22},
				},
			},

			// ---- 77. Static method with restricted parameter ----
			{
				Code: `class C { static m(undefined: any) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 20},
				},
			},

			// ---- 78. Generator method parameter ----
			{
				Code: `class C { *gen(undefined) { yield undefined; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 16},
				},
			},

			// ---- 79. Async method parameter ----
			{
				Code: `class C { async m(undefined) { return undefined; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 19},
				},
			},

			// ---- 80. Class expression name inside a JSX-like call argument context ----
			{
				Code: `register(class undefined {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 16},
				},
			},

			// ---- 81. Parameter property with default value ----
			{
				Code: `class C { constructor(public NaN: number = 5) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 30},
				},
			},

			// ---- 82. `var undefined` + assignment via compound indirect ----
			{
				Code: `var undefined; let x = (undefined = 5);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; throw (undefined = new Error());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},
			{
				Code: `var undefined; return (undefined = 5);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
				Skip: true, // SKIP: top-level return is a syntax/semantic error; keep only the annotation.
			},

			// ---- 83. Sequence expression containing a write ----
			{
				Code: `var undefined; (0, undefined = 5);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
				},
			},

			// ---- 84. Every restricted name as a catch param ----
			{
				Code: `try {} catch(undefined) {} try {} catch(NaN) {} try {} catch(Infinity) {} try {} catch(eval) {} try {} catch(arguments) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 14},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 41},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 62},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 88},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 110},
				},
			},

			// ---- 85. Generic method with restricted parameter ----
			{
				Code: `class C { m<T>(undefined: T): T { return undefined; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 16},
				},
			},

			// ---- 86. Method overload signatures (signature-only params are still runtime bindings in tsgo MethodDeclaration) ----
			{
				Code: `class C { m(undefined: number): number; m(x: string): string; m(x: any) { return x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
				},
			},

			// ---- 87. Ambient class with method param ----
			{
				Code: `declare class C { m(undefined: any): void; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 21},
				},
			},

			// ---- 88. Namespace exports with restricted name ----
			{
				Code: `namespace N { export function undefined() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 31},
				},
			},

			// ---- 89. Anonymous class expression with restricted method param ----
			{
				Code: `const X = class { foo(undefined) {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 23},
				},
			},

			// ---- 90. Nested class with restricted method param ----
			{
				Code: `class A { m() { class B { n(undefined) {} } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 29},
				},
			},

			// ---- 91. Multiple decorators on a parameter ----
			{
				Code: `class C { m(@dec1 @dec2 undefined: any) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 25},
				},
			},

			// ---- 92. Async arrow with aliased destructured rename producing restricted ----
			{
				Code: `const f = async ({a: undefined}) => undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 22},
				},
			},

			// ---- 93. Every restricted name as a function parameter in one signature ----
			{
				Code: `function f(undefined, NaN, Infinity, eval, arguments, globalThis) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12, Message: "Shadowing of global property 'undefined'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 23, Message: "Shadowing of global property 'NaN'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 28, Message: "Shadowing of global property 'Infinity'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 38, Message: "Shadowing of global property 'eval'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 44, Message: "Shadowing of global property 'arguments'."},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 55, Message: "Shadowing of global property 'globalThis'."},
				},
			},

			// ---- 94. Same-line multiple imports with assorted restricted names ----
			{
				Code: `import def, { a as NaN, b as Infinity, c as undefined, d as globalThis } from 'm';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 20},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 30},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 45},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 61},
				},
			},

			// ---- 95. Interactions between outer safe var and inner function param ----
			{
				Code: `var undefined; function f(undefined) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					// outer `var undefined` is safe (no writes), inner `undefined` param always reports.
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 27},
				},
			},

			// ---- 96. Computed key is a READ of restricted name (not shadowing);
			//          outer var undefined remains safe. Inner destructure binding reports. ----
			{
				Code: `var undefined; var {[undefined]: NaN} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 34},
				},
			},

			// ---- 97. Parameter + var redeclaration in same function scope.
			// ESLint merges defs into one variable; since one def is a Parameter
			// (not a VariableDeclarator), the whole variable is NOT safely shadowed
			// and EVERY def is reported, including the var-without-init. ----
			{
				Code: `function f(undefined) { var undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 29},
				},
			},
			{
				Code: `function f(undefined) { var undefined; var undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 29},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 44},
				},
			},

			// ---- 98. FunctionDeclaration name + param + inner var all named `undefined`.
			// The outer `function undefined` binding is in the enclosing scope (1 def).
			// The param + var merge in function scope (2 defs, one non-VariableDeclarator). ----
			{
				Code: `function undefined(undefined) { var undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 20},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 37},
				},
			},

			// ---- 99. Arrow param + inner var merge in same function scope ----
			{
				Code: `const f = (undefined) => { var undefined; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 32},
				},
			},

			// ---- 100. `let undefined;` in a nested block does NOT merge with outer param
			// (block scope is distinct). The let is safely shadowed if no inner writes. ----
			// (negative form; expressed as invalid only for the param) ----
			{
				Code: `function f(undefined) { { let undefined; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
				},
			},

			// ---- 101. Catch param + var-in-catch-body — var hoists OUT of catch scope.
			// Catch binding is catch-scoped (separate from enclosing module scope var). ----
			{
				Code: `try {} catch(undefined) { var undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 14},
				},
			},

			// ---- 102. Multiple `let` redeclarations in nested blocks are separate bindings ----
			{
				Code: `{ let undefined = 5; } { let undefined = 6; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 7},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 30},
				},
			},

			// ---- 103. Ambient `declare var undefined` merged with runtime `var undefined = 5` ----
			// The ambient decl itself has no init, but the runtime decl does; both must be reported. ----
			{
				Code: `declare var undefined: number; var undefined = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 13},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 36},
				},
			},

			// ---- 104. FunctionExpression name + param merge (named FE name is visible inside body) ----
			{
				Code: `(function undefined(undefined) { var undefined; })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 11},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 21},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 38},
				},
			},

			// ---- 105. Catch param scoped, var-in-catch-body-with-write in module scope ----
			// catch `undefined` reports; module-level `var undefined` (no write) remains safe. ----
			{
				Code: `var undefined; try {} catch(undefined) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 29},
				},
			},

			// ---- 106. Function declaration + assignment shadowing semantics ----
			// `function undefined` binds in its enclosing scope; assigning to `undefined`
			// inside the body writes to THAT same symbol (recursive self-ref-but-write).
			{
				Code: `function undefined() { undefined = 5; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
				},
			},

			// ---- 107. Two let-undefined in same block scope (TDZ / redeclaration error in TS/ESLint,
			// but if it parses we must still report). Typically a syntax error; keep as Skip. ----
			{
				Code: `{ let undefined; let undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 7},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 22},
				},
				Skip: true, // SKIP: Redeclaration of block-scoped `let` in the same block is a parse error under tsconfig strict; not worth testing alignment.
			},

			// ---- 108. FunctionDeclaration + var in the same scope — ESLint merges
			// them into one variable (2 reports); tsgo's TypeChecker keeps the
			// symbols distinct, so rslint reports only the function name. This is
			// a rare language-natural divergence tracked via scope-fallback below. ----
			{
				Code: `function undefined() {} var undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 10},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 29},
				},
			},
			{
				Code: `var undefined; function undefined() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 5},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 25},
				},
			},

			// ---- 109. Array destructuring rest element ----
			{
				Code: `var [...undefined] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 9},
				},
			},
			{
				Code: `var [a, b, ...undefined] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 15},
				},
			},

			// ---- 110. Deep rest inside nested destructure ----
			{
				Code: `var {a: {b: {...undefined}}} = x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 17},
				},
			},

			// ---- 111. Named FE's own name inside its body resolves to itself; writing it makes the name unsafe too (but FE name is always reported regardless). ----
			{
				Code: `const f = function undefined() { undefined = 5; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 20},
				},
			},

			// ---- 112. Parameter + inner const-with-init merge
			// const is block-scoped, different from function scope, so they are separate
			// bindings; only the param reports. ----
			{
				Code: `function f(undefined) { const undefined = 5; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 12},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 31},
				},
			},

			// ---- 113. Ambient overload-only declarations ----
			{
				Code: `declare function undefined(x: number): void; declare function undefined(x: string): void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 18},
					{MessageId: "shadowingRestrictedName", Line: 1, Column: 63},
				},
			},
		},
	)
}
