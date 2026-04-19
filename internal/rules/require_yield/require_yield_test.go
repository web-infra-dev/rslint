package require_yield

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireYieldRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireYieldRule,
		[]rule_tester.ValidTestCase{
			// ---- Upstream ESLint suite ----
			{Code: `function foo() { return 0; }`},
			{Code: `function* foo() { yield 0; }`},
			{Code: `function* foo() { }`},
			{Code: `(function* foo() { yield 0; })();`},
			{Code: `(function* foo() { })();`},
			{Code: `var obj = { *foo() { yield 0; } };`},
			{Code: `var obj = { *foo() { } };`},
			{Code: `class A { *foo() { yield 0; } };`},
			{Code: `class A { *foo() { } };`},

			// ---- Yield variants ----
			// Bare yield (no argument)
			{Code: `function* foo() { yield; }`},
			// yield* delegation
			{Code: `function* foo() { yield* [1, 2, 3]; }`},

			// ---- Async generators ----
			{Code: `async function* foo() { yield 0; }`},
			{Code: `async function* foo() { }`},

			// ---- Yield inside control flow ----
			{Code: `function* foo() { if (true) { yield 1; } }`},
			{Code: `function* foo() { while (true) { yield 1; } }`},
			{Code: `function* foo() { try { yield 1; } catch (e) {} }`},
			{Code: `function* foo() { try { } catch (e) { yield 1; } }`},
			{Code: `function* foo() { try { } finally { yield 1; } }`},

			// ---- Export forms ----
			{Code: `export function* foo() { yield 1; }`, Tsx: true},
			{Code: `export default function*() { yield 1; }`, Tsx: true},

			// ---- Class method forms ----
			{Code: `class A { static *foo() { yield 0; } }`},
			{Code: `class A { *#foo() { yield 0; } }`},

			// ---- Computed / class expression / anonymous ----
			{Code: `var obj = { *['foo']() { yield 0; } };`},
			{Code: `const A = class { *foo() { yield 0; } };`},
			{Code: `(function*() { yield 0; })();`},

			// ---- Arrow inside generator (arrow can't be generator) ----
			{Code: `function* foo() { const x = () => 1; yield x; }`},

			// ---- Nested generators, both have yield ----
			{Code: `function* outer() { function* inner() { yield 1; } yield 2; }`},

			// ---- TS modifiers on class generator methods ----
			{Code: `class A { async *foo() { yield 0; } }`},
			{Code: `class A { public *foo() { yield 0; } }`},
			{Code: `class A { private *foo() { yield 0; } }`},
			{Code: `class A { protected *foo() { yield 0; } }`},
			{Code: `class A { public static async *foo() { yield 0; } }`},

			// ---- Function expression as object property value ----
			{Code: `var obj = { foo: function*() { yield 0; } };`},

			// ---- Control flow variants ----
			{Code: `function* foo() { for (const x of [1]) yield x; }`},
			{Code: `function* foo(x: number) { switch (x) { case 1: yield 1; break; default: yield 0; } }`},
			{Code: `function* foo() { do { yield 1; } while (false); }`},

			// ---- Generic generator ----
			{Code: `function* foo<T>(x: T): Generator<T> { yield x; }`},

			// ---- Nested FE inside arrow inside generator ----
			{Code: `function* foo() { const f = () => function*() { yield 1; }; yield 2; }`},

			// ---- Overload signatures (no body) + impl with yield ----
			{Code: `function* foo(x: string): Generator<string>; function* foo(x: number): Generator<number>; function* foo(x: any): Generator<any> { yield x; }`},

			// ---- Ambient declarations (no body) ----
			{Code: `declare function* foo(): Generator<number>;`},

			// ---- PropertyDefinition (class field) with generator FE ----
			{Code: `class A { foo = function*() { yield 0; }; }`},

			// ---- Multi-line function head ----
			{Code: "function*\n  foo() {\n  yield 0;\n}"},

			// ---- Decorator scenarios (with yield, so valid) ----
			// Single method decorator
			{Code: `function dec(t: any, k: any, d: any) {} class A { @dec *foo() { yield 0; } }`},
			// Multiple method decorators
			{Code: `function d1(t: any, k: any, d: any) {} function d2(t: any, k: any, d: any) {} class A { @d1 @d2 *foo() { yield 0; } }`},
			// Decorator factory (with args)
			{Code: `function dec() { return (t: any, k: any, d: any) => {}; } class A { @dec() *foo() { yield 0; } }`},
			// Decorator + TS modifiers
			{Code: `function dec(t: any, k: any, d: any) {} class A { @dec public static *foo() { yield 0; } }`},
			// Decorator + async generator
			{Code: `function dec(t: any, k: any, d: any) {} class A { @dec async *foo() { yield 0; } }`},
			// Class-level decorator (shouldn't affect method position)
			{Code: `function classDec(t: any) {} @classDec class A { *foo() { yield 0; } }`},
			// Class field decorator + generator FE
			{Code: `function dec(t: any, k: any) {} class A { @dec foo = function*() { yield 0; }; }`},
			// Multi-line decorator
			{Code: "function dec(t: any, k: any, d: any) {}\nclass A {\n  @dec\n  *foo() { yield 0; }\n}"},

			// ---- JSDoc scenarios (with yield, so valid) ----
			// JSDoc before function declaration
			{Code: `/** doc */ function* foo() { yield 0; }`},
			// Multi-line JSDoc
			{Code: "/**\n * doc\n */\nfunction* foo() { yield 0; }"},
			// JSDoc before class method
			{Code: `class A { /** doc */ *foo() { yield 0; } }`},
			// JSDoc before object method
			{Code: `var o = { /** doc */ *foo() { yield 0; } };`},
			// JSDoc + FunctionExpression as property value
			{Code: `var o = { /** doc */ foo: function*() { yield 0; } };`},
			// JSDoc before anonymous FE in IIFE
			{Code: `(/** doc */ function*() { yield 0; })();`},
			// Line comment before function
			{Code: "// comment\nfunction* foo() { yield 0; }"},
			// Block comment (non-JSDoc) before function
			{Code: `/* comment */ function* foo() { yield 0; }`},
			// JSDoc on class field with FE
			{Code: `class A { /** doc */ foo = function*() { yield 0; }; }`},

			// ---- Yield attribution boundary scenarios ----
			// Yield in computed property key of object method (belongs to outer generator).
			{Code: `function* foo() { const o = { [yield 1]() { return 0; } }; return 0; }`},
			// Yield in computed property key of class method.
			{Code: `function* foo() { class C { [yield 1]() { return 0; } } return 0; }`},
			// Yield in class heritage expression.
			{Code: `function* foo() { class C extends (yield 1) {} return 0; }`},
			// Nested yield expression.
			{Code: `function* foo() { yield yield 1; }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Upstream ESLint suite ----
			{
				Code: `function* foo() { return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `(function* foo() { return 0; })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 2, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `var obj = { *foo() { return 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: `class A { *foo() { return 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `function* foo() { function* bar() { yield 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `function* foo() { function* bar() { return 0; } yield 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 19, EndLine: 1, EndColumn: 32},
				},
			},

			// ---- Async generator ----
			{
				Code: `async function* foo() { return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},

			// ---- Export forms ----
			{
				Code: `export function* foo() { return 0; }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 8, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: `export default function*() { return 0; }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Class method forms ----
			{
				Code: `class A { static *foo() { return 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 11, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: `class A { *#foo() { return 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},

			// ---- Computed / class expression / anonymous FE ----
			{
				Code: `var obj = { *['foo']() { return 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 13, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: `const A = class { *foo() { return 0; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 19, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: `(function*() { return 0; })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},

			// ---- Control flow without yield in any branch ----
			{
				Code: `function* foo() { if (true) { return 1; } else { return 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},

			// ---- Arrow inside generator (arrow not a generator, body has no yield) ----
			{
				Code: `function* foo() { const x = () => 1; return x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},

			// ---- Multi-line ----
			{
				Code: "function* foo() {\n  return 0;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},

			// ---- Async class method generator ----
			{
				Code: `class A { async *foo() { return 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 11, EndLine: 1, EndColumn: 21},
				},
			},

			// ---- Public/static combined ----
			{
				Code: `class A { public static async *foo() { return 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 11, EndLine: 1, EndColumn: 35},
				},
			},

			// ---- Function expression as object property value ----
			{
				Code: `var obj = { foo: function*() { return 0; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 13, EndLine: 1, EndColumn: 27},
				},
			},

			// ---- Generic generator ----
			{
				Code: `function* foo<T>(x: T): Generator<T> { return x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},

			// ---- Overload signatures + impl with no yield ----
			{
				Code: `function* foo(x: string): Generator<string>; function* foo(x: number): Generator<number>; function* foo(x: any): Generator<any> { return x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 91, EndLine: 1, EndColumn: 104},
				},
			},

			// ---- Nested FE inside arrow: outer has no yield ----
			{
				Code: `function* foo() { const f = () => function*() { yield 1; }; return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},

			// ---- Class field with generator FE ----
			{
				Code: `class A { foo = function*() { return 0; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 11, EndLine: 1, EndColumn: 26},
				},
			},

			// ---- Multi-line function head ----
			{
				Code: "function*\n  foo() {\n  return 0;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},

			// ---- Decorator scenarios ----
			// Single method decorator (no args)
			{
				Code: "declare function dec(t: any, k: any, d: any): void;\nclass A { @dec *foo() { return 0; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 2, Column: 11, EndLine: 2, EndColumn: 20},
				},
			},
			// Decorator factory (with args) — stress test for findOpenParenPos
			{
				Code: "declare function dec(): (t: any, k: any, d: any) => void;\nclass A { @dec() *foo() { return 0; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 2, Column: 11, EndLine: 2, EndColumn: 22},
				},
			},
			// Multiple method decorators
			{
				Code: "declare function d1(t: any, k: any, d: any): void; declare function d2(t: any, k: any, d: any): void;\nclass A { @d1 @d2 *foo() { return 0; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 2, Column: 11, EndLine: 2, EndColumn: 23},
				},
			},
			// Decorator + TS modifiers
			{
				Code: "declare function dec(t: any, k: any, d: any): void;\nclass A { @dec public static *foo() { return 0; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 2, Column: 11, EndLine: 2, EndColumn: 34},
				},
			},
			// Decorator + async generator
			{
				Code: "declare function dec(t: any, k: any, d: any): void;\nclass A { @dec async *foo() { return 0; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 2, Column: 11, EndLine: 2, EndColumn: 26},
				},
			},
			// Class-level decorator (shouldn't affect method position)
			{
				Code: "declare function classDec(t: any): void;\n@classDec\nclass A { *foo() { return 0; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 3, Column: 11, EndLine: 3, EndColumn: 15},
				},
			},
			// Class field decorator + generator FE
			{
				Code: "declare function dec(t: any, k: any): void;\nclass A { @dec foo = function*() { return 0; }; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 2, Column: 11, EndLine: 2, EndColumn: 31},
				},
			},
			// Multi-line decorator
			{
				Code: "declare function dec(t: any, k: any, d: any): void;\nclass A {\n  @dec\n  *foo() { return 0; }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 3, Column: 3, EndLine: 4, EndColumn: 7},
				},
			},

			// ---- JSDoc / comment scenarios ----
			// JSDoc before function declaration
			{
				Code: `/** doc */ function* foo() { return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 12, EndLine: 1, EndColumn: 25},
				},
			},
			// Multi-line JSDoc
			{
				Code: "/**\n * doc\n */\nfunction* foo() { return 0; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 4, Column: 1, EndLine: 4, EndColumn: 14},
				},
			},
			// JSDoc before class method
			{
				Code: `class A { /** doc */ *foo() { return 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				},
			},
			// JSDoc before object method
			{
				Code: `var o = { /** doc */ *foo() { return 0; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				},
			},
			// JSDoc + FunctionExpression as property value
			{
				Code: `var o = { /** doc */ foo: function*() { return 0; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 22, EndLine: 1, EndColumn: 36},
				},
			},
			// JSDoc before anonymous FE in IIFE
			{
				Code: `(/** doc */ function*() { return 0; })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 13, EndLine: 1, EndColumn: 22},
				},
			},
			// Line comment before function
			{
				Code: "// comment\nfunction* foo() { return 0; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 2, Column: 1, EndLine: 2, EndColumn: 14},
				},
			},
			// Block comment (non-JSDoc) before function
			{
				Code: `/* comment */ function* foo() { return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 15, EndLine: 1, EndColumn: 28},
				},
			},
			// JSDoc on class field with FE
			{
				Code: `class A { /** doc */ foo = function*() { return 0; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 22, EndLine: 1, EndColumn: 37},
				},
			},

			// ---- Illegal yield in nested non-generator scope must NOT rescue outer generator ----
			// tsgo performs error-recovery parsing and still emits a KindYieldExpression
			// node for these illegal positions; attribution logic must isolate them.
			{
				Code: `function* foo() { function inner() { yield 1; } return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `function* foo() { const f = function() { yield 1; }; return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `function* foo() { const f = () => { yield 1; }; return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `function* foo() { const f = () => yield 1; return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `function* foo() { class C { m() { yield 1; } } return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `function* foo() { const o = { get x() { yield 1; return 0; } }; return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `function* foo() { const o = { set x(v) { yield 1; } }; return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `function* foo() { class C { constructor() { yield 1; } } return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `function* foo() { class C { x = yield 1; } return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			// Two-level: outer gen -> non-generator method -> nested generator with yield.
			// Outer gen has no yield; inner nested gen is a separate scope and does not count.
			{
				Code: `function* foo() { class C { m() { function* g() { yield 1; } } } return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},

			// ---- Illegal yield in parameter default values (nested non-gen) ----
			{
				Code: `function* outer() { function bar(x = yield 1) { return x; } return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `function* outer() { const f = (x = yield 1) => x; return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `function* outer() { class C { m(x = yield 1) { return x; } } return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `function* outer() { const o = { set x(v = yield 1) {} }; return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `function* outer() { class C { constructor(x = yield 1) {} } return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},

			// ---- Class static block (illegal yield) ----
			{
				Code: `function* outer() { class C { static { yield 1; } } return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// Static block with both illegal and legal yields in sub-scopes.
			{
				Code: `function* outer() { class C { static { function* g() { yield 1; } yield 2; } } return 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingYield", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
		},
	)
}
