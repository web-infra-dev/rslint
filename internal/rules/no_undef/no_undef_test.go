package no_undef

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
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

			// === Built-in globals via lib ===
			{Code: `console.log("test");`},
			{Code: `var p = new Promise<void>((resolve) => resolve());`},
			{Code: `setTimeout(() => {}, 100);`},

			// === Type-only positions: type annotations ===
			{Code: `type MyType = string; var x: MyType;`},
			{Code: `interface MyInterface { x: number; } var y: MyInterface;`},
			{Code: `function f(x: string): string { return x; }`},

			// === Type-only positions: generic type arguments ===
			{Code: `function identity<T>(val: T): T { return val; }`},
			{Code: `function constrained<T extends object>(val: T) { return val; }`},

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

			// === this in class method ===
			{Code: `class C { x = 1; m() { return this.x; } }`},

			// === Chained property access (only root is checked) ===
			{Code: `var obj = { a: { b: { c: 1 } } }; obj.a.b.c;`},

			// === Export declared ===
			{Code: `var exportedVar = 1; export { exportedVar };`},
			{Code: `var localVar = 2; export { localVar as renamedExport };`},

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
				Code: `console.log(undeclaredArg123);`,
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
