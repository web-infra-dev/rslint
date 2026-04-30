package max_params

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// objectOption returns the array-wrapped single-option shape that matches
// rule_tester / multi-element CLI configs and exercises the JSON path through
// utils.GetOptionsMap (the typed-struct shortcut would silently bypass it).
func objectOption(opts map[string]interface{}) []interface{} {
	return []interface{}{opts}
}

// TestMaxParams covers four corpora:
//
//  1. Upstream parity (typescript-eslint) — every case from
//     typescript-eslint/packages/eslint-plugin/tests/rules/max-params.test.ts
//     migrated 1:1.
//  2. Upstream parity (eslint core, TS-meaningful cases) — the upstream eslint
//     core test cases that remain valid under the typescript-eslint schema
//     (object-only options; no `countThis`).
//  3. Universal edge shapes — every container kind ESLint's FunctionExpression
//     listener implicitly covers in ESTree but tsgo splits into distinct kinds
//     (constructor, method, getter, setter, class field arrow, object literal
//     method, class expression method, function type literal). Each kind is
//     exercised on both the valid and invalid side, plus modifier crosses
//     (static / private / async / generator / async-generator / abstract /
//     decorated).
//  4. Real-world & boundary scenarios — `this: void` x rest params, parameter
//     properties (constructor `public a`/`private b`/`readonly c`),
//     overload signatures, deeply-nested function-likes, decorated methods,
//     event-handler signatures, computed-property names, optional / default
//     parameters (each counts as 1).
func TestMaxParams(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxParamsRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// 1. Upstream parity — typescript-eslint
			//    packages/eslint-plugin/tests/rules/max-params.test.ts
			// ============================================================
			{Code: "function foo() {}"},
			{Code: "const foo = function () {};"},
			{Code: "const foo = () => {};"},
			{Code: "function foo(a) {}"},
			{Code: `
class Foo {
  constructor(a) {}
}
			`},
			{Code: `
class Foo {
  method(this: void, a, b, c) {}
}
			`},
			{Code: `
class Foo {
  method(this: Foo, a, b) {}
}
			`},
			{Code: "function foo(a, b, c, d) {}", Options: objectOption(map[string]interface{}{"max": 4})},
			{Code: "function foo(a, b, c, d) {}", Options: objectOption(map[string]interface{}{"maximum": 4})},
			{
				Code:    "\nclass Foo {\n  method(this: void) {}\n}\n",
				Options: objectOption(map[string]interface{}{"max": 0}),
			},
			{
				Code:    "\nclass Foo {\n  method(this: void, a) {}\n}\n",
				Options: objectOption(map[string]interface{}{"max": 1}),
			},
			{
				Code:    "\nclass Foo {\n  method(this: void, a) {}\n}\n",
				Options: objectOption(map[string]interface{}{"countVoidThis": true, "max": 2}),
			},
			{
				Code:    "function testD(this: void, a) {}",
				Options: objectOption(map[string]interface{}{"max": 1}),
			},
			{
				Code:    "function testD(this: void, a) {}",
				Options: objectOption(map[string]interface{}{"countVoidThis": true, "max": 2}),
			},
			{
				Code:    "const testE = function (this: void, a) {}",
				Options: objectOption(map[string]interface{}{"max": 1}),
			},
			{
				Code:    "const testE = function (this: void, a) {}",
				Options: objectOption(map[string]interface{}{"countVoidThis": true, "max": 2}),
			},
			{
				Code:    "\ndeclare function makeDate(m: number, d: number, y: number): Date;\n",
				Options: objectOption(map[string]interface{}{"max": 3}),
			},
			{
				Code:    "\ntype sum = (a: number, b: number) => number;\n",
				Options: objectOption(map[string]interface{}{"max": 2}),
			},

			// ============================================================
			// 2. Upstream parity — eslint core valid cases
			//    (rewritten from `[3]` to `[{max: 3}]` since the
			//    typescript-eslint schema is object-only; integer form is
			//    rejected upstream by schema validation.)
			// ============================================================
			{Code: "function test(d, e, f) {}"},
			{Code: "var test = function(a, b, c) {};", Options: objectOption(map[string]interface{}{"max": 3})},
			{Code: "var test = (a, b, c) => {};", Options: objectOption(map[string]interface{}{"max": 3})},
			{Code: "var test = function test(a, b, c) {};", Options: objectOption(map[string]interface{}{"max": 3})},

			// ============================================================
			// 3. Universal edge shapes — tsgo container kinds
			// ============================================================

			// --- Default option (max=3) on every container ---
			{Code: "function foo(a, b, c) {}"},
			{Code: "const foo = function (a, b, c) {};"},
			{Code: "const foo = (a, b, c) => {};"},
			{Code: "class C { method(a, b, c) {} }"},
			{Code: "class C { static method(a, b, c) {} }"},
			{Code: "class C { #method(a, b, c) {} }"},
			{Code: "class C { static #method(a, b, c) {} }"},
			{Code: "class C { constructor(a, b, c) {} }"},
			{Code: "class C { get x() { return 0; } }"},
			{Code: "class C { set x(v) {} }"},
			{Code: "class C { handler = (a, b, c) => {}; }"},
			{Code: "var X = class { method(a, b, c) {} }"},
			{Code: "var X = class { #m(a, b, c) {} }"},
			{Code: "var obj = { method(a, b, c) {} }"},
			{Code: "var obj = { get x() { return 0; } }"},
			{Code: "var obj = { set x(v) {} }"},
			{Code: "var obj = { ['computed'](a, b, c) {} }"},

			// --- async / generator / async-generator on each form ---
			{Code: "async function foo(a, b, c) {}"},
			{Code: "function* foo(a, b, c) {}"},
			{Code: "async function* foo(a, b, c) {}"},
			{Code: "const foo = async (a, b, c) => {};"},
			{Code: "const foo = async function (a, b, c) {};"},
			{Code: "const foo = async function* (a, b, c) {};"},
			{Code: "class C { async method(a, b, c) {} }"},
			{Code: "class C { *method(a, b, c) {} }"},
			{Code: "class C { async *method(a, b, c) {} }"},
			{Code: "class C { static async #m(a, b, c) {} }"},

			// --- Function-like types (TSFunctionType / declare) ---
			{Code: "type F = (a: number, b: number, c: number) => void;"},
			{Code: "declare function foo(a: number, b: number, c: number): void;"},

			// --- Empty bodies / edge param shapes ---
			{Code: "function foo() {}"},
			{Code: "function foo(...args) {}"},
			{Code: "function foo({a, b, c}) {}"},
			{Code: "function foo([a, b], {c}, d) {}"},
			{Code: "function foo(a?, b?, c?) {}"},
			{Code: "function foo(a = 1, b = 2, c = 3) {}"},
			{Code: "function foo(a: number, b: string, c: boolean) {}"},
			{
				// Generic type parameters do NOT count toward max.
				Code: "function foo<T, U, V, W, X>(a, b, c) {}",
			},

			// --- Options shapes — exercise the JSON path ---
			// Bare object (single-option CLI shape).
			{
				Code:    "function foo(a, b, c, d) {}",
				Options: map[string]interface{}{"max": 4},
			},
			// Array-wrapped (multi-element / rule_tester shape).
			{
				Code:    "function foo(a, b, c, d) {}",
				Options: objectOption(map[string]interface{}{"max": 4}),
			},
			// Empty options array → defaults to 3.
			{Code: "function foo(a, b, c) {}", Options: []interface{}{}},
			// Empty options object → defaults to 3.
			{
				Code:    "function foo(a, b, c) {}",
				Options: objectOption(map[string]interface{}{}),
			},
			// Legacy `maximum` key.
			{
				Code:    "function foo(a, b, c, d) {}",
				Options: map[string]interface{}{"maximum": 4},
			},
			// `maximum: 0` with no `max` → ESLint sets numParams=undefined,
			// effectively disabling the rule.
			{
				Code:    "function foo(a, b, c, d, e, f) {}",
				Options: objectOption(map[string]interface{}{"maximum": 0}),
			},
			// `maximum: 0, max: 4` → falls through to `max` per JS truthy.
			{
				Code:    "function foo(a, b, c, d) {}",
				Options: objectOption(map[string]interface{}{"maximum": 0, "max": 4}),
			},
			// `maximum: 4, max: 2` → maximum wins (truthy).
			{
				Code:    "function foo(a, b, c, d) {}",
				Options: objectOption(map[string]interface{}{"maximum": 4, "max": 2}),
			},
			// Bare integer / array-of-integer is REJECTED upstream by schema —
			// rslint has no schema layer, so we treat as "no option" and fall
			// back to defaults rather than guessing.
			{Code: "function foo(a, b, c) {}", Options: 7},
			{Code: "function foo(a, b, c) {}", Options: []interface{}{7}},

			// --- this parameter handling ---
			// `this` with no type annotation — never stripped (4 effective).
			{
				Code:    "function f(this, a, b, c) {}",
				Options: objectOption(map[string]interface{}{"max": 4}),
			},
			// `this: any` — not void, not stripped (4 effective).
			{
				Code:    "function f(this: any, a, b, c) {}",
				Options: objectOption(map[string]interface{}{"max": 4}),
			},
			// `this: Foo` — not void, not stripped (4 effective).
			{
				Code:    "function f(this: Foo, a, b, c) {}",
				Options: objectOption(map[string]interface{}{"max": 4}),
			},
			// `this: never` — not void, not stripped (4 effective).
			{
				Code:    "function f(this: never, a, b, c) {}",
				Options: objectOption(map[string]interface{}{"max": 4}),
			},
			// `this: unknown` — not void, not stripped (4 effective).
			{
				Code:    "function f(this: unknown, a, b, c) {}",
				Options: objectOption(map[string]interface{}{"max": 4}),
			},
			// `this: void` with countVoidThis: true — counted (4 with -> exceed unless max bumped).
			{
				Code:    "function f(this: void, a, b, c) {}",
				Options: objectOption(map[string]interface{}{"countVoidThis": true, "max": 4}),
			},
			// `this: void` + rest params: stripped, rest counts as 1.
			{Code: "function f(this: void, ...args) {}"},
			{Code: "function f(this: void, a, ...rest) {}"},
			// `this: void` + destructuring + default: still strips this, rest count regular.
			{Code: "function f(this: void, {a}, b = 1) {}"},
			// First param destructuring with key 'this' — NOT a real this param.
			{Code: "function f({ this: x }: any, a, b) {}"},
			// Method overload signatures — each declaration has ≤ 3 params.
			{Code: `
class Foo {
  method(a: number): void;
  method(a: number, b: string): void;
  method(...args: any[]): void {}
}
			`},

			// --- Parameter properties (TS class constructors) ---
			// Each parameter property counts as one param toward the limit.
			{Code: "class C { constructor(public a: number, private b: string, readonly c: boolean) {} }"},

			// --- Decorators on methods — head loc skips decorators, count is unaffected ---
			{Code: `
declare const dec: any;
class C {
  @dec
  method(a, b, c) {}
}
			`},

			// --- Abstract methods ---
			{Code: `
abstract class A {
  abstract foo(a, b, c): void;
}
			`},

			// --- Deeply nested function-likes — each gets its own counter ---
			{Code: "function outer(a) { function inner() { function deep(a, b, c) {} } }"},
			{Code: "class C { method(a, b, c) { return function inner(d, e, f) {}; } }"},

			// --- Real-world: callback signatures within max=3 ---
			{Code: "[1].forEach((value, index, array) => {});"},
			{Code: "const handler = (req, res, next) => {};"},

			// --- Object-property function expression / arrow ---
			{Code: "var obj = { foo: function (a, b, c) {} }"},
			{Code: "var obj = { foo: (a, b, c) => {} }"},
			{Code: "var obj = { foo: async function (a, b, c) {} }"},
			{Code: "var obj = { foo: function bar(a, b, c) {} }"},
			// Numeric / string property keys
			{Code: "var obj = { 0(a, b, c) {} }"},
			{Code: "var obj = { 'string-key'(a, b, c) {} }"},

			// --- Interface method signatures are NOT checked (upstream
			// typescript-eslint does not listen on TSMethodSignature). ---
			{Code: "interface I { foo(a, b, c, d, e): void }"},
			{Code: "interface I { (a, b, c, d, e): void }"},
			{Code: "interface I { new (a, b, c, d, e): void }"},
			// Index signatures: only 1 param syntactically; never exceeds.
			{Code: "interface I { [k: string]: any }"},

			// --- Function type as property value of interface IS checked
			// (it is a TSFunctionType, which IS in the listener set). ---
			{Code: "interface I { foo: (a, b, c) => void }"},

			// --- Class field initialized with function expression ---
			{Code: "class C { foo = function (a, b, c) {}; }"},
			{Code: "class C { foo = function bar(a, b, c) {}; }"},

			// --- Constructor type / call signature appearing inside other types ---
			// `new (...) => T` is KindConstructorType (not in listener set) ✓
			{Code: "type Ctor = new (a, b, c, d, e) => any;"},

			// --- Class with both static and instance methods ---
			{Code: `
class C {
  static a(a, b, c) {}
  b(a, b, c) {}
  static get g() { return 0; }
  set s(v) {}
}
			`},

			// --- TS namespaces / declare module ---
			{Code: `
declare module 'x' {
  function foo(a, b, c): void;
  export function bar(a, b, c): void;
}
			`},
		},

		[]rule_tester.InvalidTestCase{
			// ============================================================
			// 1. Upstream parity — typescript-eslint invalid
			// ============================================================
			{
				Code: "function foo(a, b, c, d) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'foo' has too many parameters (4). Maximum allowed is 3.",
						Line:      1, Column: 1, EndLine: 1, EndColumn: 13,
					},
				},
			},
			{
				Code:   "const foo = function (a, b, c, d) {};",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:   "const foo = (a, b, c, d) => {};",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:    "const foo = a => {};",
				Options: objectOption(map[string]interface{}{"max": 0}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code: "\nclass Foo {\n  method(this: void, a, b, c, d) {}\n}\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
				},
			},
			{
				Code:    "\nclass Foo {\n  method(this: void, a) {}\n}\n",
				Options: objectOption(map[string]interface{}{"countVoidThis": true, "max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:   "\nclass Foo {\n  method(this: Foo, a, b, c) {}\n}\n",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:    "\ndeclare function makeDate(m: number, d: number, y: number): Date;\n",
				Options: objectOption(map[string]interface{}{"max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:    "\ntype sum = (a: number, b: number) => number;\n",
				Options: objectOption(map[string]interface{}{"max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},

			// ============================================================
			// 2. Upstream parity — eslint core invalid (object-form options)
			// ============================================================
			{
				Code:    "function test(a, b, c) {}",
				Options: objectOption(map[string]interface{}{"max": 2}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'test' has too many parameters (3). Maximum allowed is 2.",
						Line:      1, Column: 1, EndLine: 1, EndColumn: 14,
					},
				},
			},
			{
				Code: "function test(a, b, c, d) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'test' has too many parameters (4). Maximum allowed is 3.",
					},
				},
			},
			{
				Code:    "var test = function(a, b, c, d) {};",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:    "var test = (a, b, c, d) => {};",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:    "(function(a, b, c, d) {});",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:    "var test = function test(a, b, c) {};",
				Options: objectOption(map[string]interface{}{"max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			// Object property options
			{
				Code:    "function test(a, b, c) {}",
				Options: objectOption(map[string]interface{}{"max": 2}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:    "function test(a, b, c, d) {}",
				Options: objectOption(map[string]interface{}{}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:    "function test(a) {}",
				Options: objectOption(map[string]interface{}{"max": 0}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			// Error location should not cover the entire function — just the head.
			{
				Code: `function test(a, b, c) {
              // Just to make it longer
            }`,
				Options: objectOption(map[string]interface{}{"max": 2}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},

			// ============================================================
			// 3. Universal edge shapes — every container kind reports
			// ============================================================

			// --- Class member kinds (each produces a distinct description) ---
			{
				Code:    "class C { method(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Method 'method' has too many parameters (4). Maximum allowed is 3.",
						Line:      1, Column: 11, EndLine: 1, EndColumn: 17,
					},
				},
			},
			{
				Code:    "class C { static method(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Static method 'method' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "class C { #method(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Private method '#method' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "class C { static #method(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Static private method '#method' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "class C { constructor(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Constructor has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "class C { set x(v) {} }",
				Options: objectOption(map[string]interface{}{"max": 0}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Setter 'x' has too many parameters (1). Maximum allowed is 0."},
				},
			},
			{
				Code:    "class C { static set x(v) {} }",
				Options: objectOption(map[string]interface{}{"max": 0}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Static setter 'x' has too many parameters (1). Maximum allowed is 0."},
				},
			},
			{
				Code:    "class C { handler = (a, b, c, d) => {}; }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			// Class expression
			{
				Code:    "var X = class { m(a, b, c, d) {} };",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},

			// --- Object literal forms ---
			{
				Code:    "var obj = { method(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Method 'method' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "var obj = { set x(v) {} }",
				Options: objectOption(map[string]interface{}{"max": 0}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:    "var obj = { ['m'](a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Method 'm' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- async / generator / async-generator (modifier ordering) ---
			{
				Code:    "async function foo(a, b, c, d) {}",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Async function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "function* foo(a, b, c, d) {}",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Generator function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "async function* foo(a, b, c, d) {}",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Async generator function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "const foo = async (a, b, c, d) => {};",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Async arrow function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "class C { async method(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Async method 'method' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "class C { async *method(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Async generator method 'method' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "class C { static async #m(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Static private async method '#m' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Function-like types (TSFunctionType / declare function) ---
			{
				Code:    "type F = (a: number, b: number, c: number, d: number) => void;",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				Code:    "declare function foo(a, b, c, d): void;",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Anonymous function expression / IIFE ---
			{
				Code: "(function(a, b, c, d) {})();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Function has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// ============================================================
			// 4. Real-world & boundary scenarios
			// ============================================================

			// --- countVoidThis: false (default) — strip `this: void` only ---
			{
				// `this: void` stripped (default), 4 remaining > 3.
				Code:    "function f(this: void, a, b, c, d) {}",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				// `this: any` is NOT stripped — 4 total > 3.
				Code:    "function f(this: any, a, b, c) {}",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},
			{
				// `this: void` + rest: stripped + rest counts as 1, 4 effective.
				Code:    "function f(this: void, a, b, c, ...rest) {}",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Function 'f' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				// countVoidThis:true + rest: this:void NOT stripped, 5 effective.
				Code:    "function f(this: void, a, b, c, ...rest) {}",
				Options: objectOption(map[string]interface{}{"max": 4, "countVoidThis": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Function 'f' has too many parameters (5). Maximum allowed is 4."},
				},
			},
			{
				// `this` without type — never stripped (4 effective).
				Code:    "function f(this, a, b, c) {}",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},

			// --- Parameter properties: each counts as one param ---
			{
				Code: "class C { constructor(public a: number, private b: string, readonly c: boolean, d: any) {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Constructor has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Optional / default parameters: each counts as one ---
			{Code: "function foo(a?, b?, c?, d?) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}}},
			{Code: "function foo(a = 1, b = 2, c = 3, d = 4) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}}},

			// --- Decorated method ---
			{
				Code: `
declare const dec: any;
class C {
  @dec
  method(a, b, c, d) {}
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Method 'method' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Abstract method ---
			{
				Code: `
abstract class A {
  abstract foo(a, b, c, d): void;
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Method 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Overload signatures: each declaration counted independently ---
			{
				Code: `
class Foo {
  method(a, b): void;
  method(a, b, c, d): void;
  method(...args: any[]): void {}
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 3, EndLine: 4, EndColumn: 9,
						Message: "Method 'method' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Inner function checked independently of outer ---
			{
				Code:    "function outer(a) { function inner(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Function 'inner' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "function outer(a, b, c, d) { function inner(x) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Function 'outer' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			// Both outer AND inner exceed → two separate diagnostics.
			{
				Code:    "function outer(a, b, c, d) { function inner(e, f, g, h) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Function 'outer' has too many parameters (4). Maximum allowed is 3."},
					{MessageId: "exceed",
						Message: "Function 'inner' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			// Class method body containing nested function — both reported.
			{
				Code: `
class C {
  outer(a, b, c, d) {
    return function inner(e, f, g, h) {};
  }
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 3, EndLine: 3, EndColumn: 8,
						Message: "Method 'outer' has too many parameters (4). Maximum allowed is 3."},
					{MessageId: "exceed",
						Message: "Function 'inner' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Multi-line position assertion ---
			{
				Code: `
function test(
  a,
  b,
  c,
  d
) {}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 14},
				},
			},

			// --- Real-world: Express handler signature ---
			{
				Code:    "const handler = (req, res, next, err) => {};",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Arrow function 'handler' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Real-world: forEach-like callback exceeding ---
			{
				Code:    "[].forEach(function (value, index, array, extra) {});",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed"}},
			},

			// --- countVoidThis combined with overloads (each declaration independent) ---
			{
				Code: `
class C {
  m(this: void, a, b, c): void;
  m(this: void, a, b, c, d): void;
  m(...args: any[]): any {}
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4,
						Message: "Method 'm' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Object-property FunctionExpression: described as "method", not "function" ---
			{
				Code:    "var obj = { foo: function (a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Method 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				// Named FunctionExpression as object property value: own name wins.
				Code:    "var obj = { foo: function bar(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Method 'bar' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				// Async object-property method (via FunctionExpression form).
				Code:    "var obj = { foo: async function (a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Async method 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				// Arrow as property value: still "Arrow function" (matches ESLint).
				Code:    "var obj = { foo: (a, b, c, d) => {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Arrow function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Class field initialized with function expression ---
			// Class fields are PropertyDeclaration in tsgo (not PropertyAssignment),
			// so the FunctionExpression-as-method branch does NOT apply — the
			// description is "function 'foo'" via parent walk for the
			// PropertyDeclaration name, matching ESLint (in ESTree the parent is
			// `PropertyDefinition`, which is not in the method-classifier branch).
			{
				Code:    "class C { foo = function (a, b, c, d) {}; }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "class C { foo = async function (a, b, c, d) {}; }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Async function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			// --- Class field with PrivateIdentifier or `static` modifier ---
			// ESLint v9's getFunctionNameWithKind picks up `parent.static`
			// and `parent.key.type === "PrivateIdentifier"` for PropertyDefinition
			// whose value is the function-like, so the description gets
			// "static" / "private" prefixes and `'#name'` from the field key.
			{
				// Private class field arrow.
				Code:    "class C { #foo = (a, b, c, d) => {}; }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Private arrow function '#foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				// Static private class field arrow.
				Code:    "class C { static #foo = (a, b, c, d) => {}; }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Static private arrow function '#foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				// Static class field arrow (non-private).
				Code:    "class C { static foo = (a, b, c, d) => {}; }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Static arrow function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				// Static async class field arrow — modifier ordering preserved.
				Code:    "class C { static foo = async (a, b, c, d) => {}; }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Static async arrow function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				// Private class field FunctionExpression value.
				Code:    "class C { #foo = function (a, b, c, d) {}; }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Private function '#foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				// Static private class field FunctionExpression with async.
				Code:    "class C { static #foo = async function (a, b, c, d) {}; }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Static private async function '#foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Function type inside an interface property signature ---
			// Now that getFunctionNameForDescription walks PropertySignature
			// parents, the description recovers `foo` from the property name.
			{
				Code:    "interface I { foo: (a, b, c, d) => void }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Function 'foo' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Function type as the value of a TypeAliasDeclaration ---
			// Walks TypeAliasDeclaration parent for the alias name.
			{
				Code:    "type F = (a: number, b: number, c: number, d: number) => void;",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Function 'F' has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- declare class methods (no body) ---
			{
				Code: `
declare class Foo {
  method(a, b, c, d): void;
}
				`,
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Method 'method' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code: `
declare class Foo {
  static method(a, b, c, d): void;
  constructor(a, b, c, d);
}
				`,
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Static method 'method' has too many parameters (4). Maximum allowed is 3."},
					{MessageId: "exceed",
						Message: "Constructor has too many parameters (4). Maximum allowed is 3."},
				},
			},

			// --- Numeric / string property key methods ---
			{
				Code:    "var obj = { 0(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Method '0' has too many parameters (4). Maximum allowed is 3."},
				},
			},
			{
				Code:    "var obj = { 'a-b'(a, b, c, d) {} }",
				Options: objectOption(map[string]interface{}{"max": 3}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed",
						Message: "Method 'a-b' has too many parameters (4). Maximum allowed is 3."},
				},
			},
		},
	)
}
