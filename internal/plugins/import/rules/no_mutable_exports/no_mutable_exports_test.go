package no_mutable_exports_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_mutable_exports"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMutableExportsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_mutable_exports.NoMutableExportsRule,
		[]rule_tester.ValidTestCase{
			// === Direct exports ===
			{Code: `export const count = 1`},
			{Code: `export function getCount() {}`},
			{Code: `export class Counter {}`},
			{Code: `export enum Direction { Up, Down }`},
			{Code: `export interface Foo {}`},
			{Code: `export type Foo = {}`},
			// Multiple const declarations
			{Code: `export const a = 1, b = 2`},
			// Const destructuring
			{Code: `export const { a, b } = obj`},
			{Code: `export const [a, b] = arr`},
			// using (const-like)
			{Code: `export using x = getResource()`},

			// === Default exports ===
			{Code: `export default function getCount() {}`},
			{Code: `export default class Counter {}`},
			// Default export of assignment expression (not a declaration lookup)
			{Code: `export default count = 1`},
			// Default export of literals/expressions
			{Code: `export default 42`},
			{Code: `export default { x: 1 }`},
			{Code: `export default [1, 2, 3]`},
			{Code: `export default () => {}`},

			// === Named exports with const ===
			{Code: "const count = 1\nexport { count }"},
			{Code: "const count = 1\nexport { count as counter }"},
			{Code: "const count = 1\nexport { count as default }"},
			// Const destructuring then named export
			{Code: "const { a } = obj\nexport { a }"},
			{Code: "const [a] = arr\nexport { a }"},
			{Code: "const { a: b } = obj\nexport { b }"},
			{Code: "const { a = 1 } = obj\nexport { a }"},
			{Code: "const { ...rest } = obj\nexport { rest }"},
			{Code: "const [, b] = arr\nexport { b }"},
			{Code: "const { a: { b } } = obj\nexport { b }"},

			// === Default export of const ===
			{Code: "const count = 1\nexport default count"},
			// Parenthesized const (parens stripped)
			{Code: "const x = 1\nexport default (x)"},

			// === Function/class then export ===
			{Code: "function getCount() {}\nexport { getCount }"},
			{Code: "function getCount() {}\nexport { getCount as getCounter }"},
			{Code: "function getCount() {}\nexport default getCount"},
			{Code: "function getCount() {}\nexport { getCount as default }"},
			{Code: "class Counter {}\nexport { Counter }"},
			{Code: "class Counter {}\nexport { Counter as Count }"},
			{Code: "class Counter {}\nexport default Counter"},
			{Code: "class Counter {}\nexport { Counter as default }"},

			// === Enum then export ===
			{Code: "enum Direction { Up, Down }\nexport { Direction }"},

			// === Type-only exports ===
			{Code: "type Foo = {}\nexport type { Foo }"},
			// Type-only specifier: valid even if x is let
			{Code: "let x = 1\nexport { type x }"},

			// === Re-exports (has module specifier, always valid) ===
			{Code: `export { foo } from './foo'`},
			{Code: `export { foo as bar } from './foo'`},
			{Code: `export * from './foo'`},
			{Code: `export * as ns from './foo'`},

			// === Import then re-export by name ===
			{Code: "import { x } from './first'\nexport { x }"},
			{Code: "import { x } from './first'\nexport default x"},

			// === Undeclared identifier (no matching declaration, no report) ===
			{Code: "export { undeclared }"},
			{Code: "export default undeclared"},

			// === Empty export ===
			{Code: "export {}"},

			// === TypeScript: export = ===
			{Code: "const x = 1\nexport = x"},

			// === Mixed specifiers: all const/func (valid) ===
			{Code: "const a = 1\nfunction b() {}\nexport { a, b }"},

			// === Namespace/module: export let inside namespace is NOT an ES module export ===
			{Code: "namespace Foo { export let x = 1 }"},
			{Code: "module Foo { export let x = 1 }"},
			{Code: "declare namespace Foo { export let x: number }"},
			{Code: "declare module 'foo' { export let x: number }"},
			{Code: "namespace A { namespace B { export let x = 1 } }"},

			// === Hoisting boundaries: let/const inside block does NOT hoist ===
			{Code: "{ let x = 1 }\nexport { x }"},
			{Code: "{ const x = 1 }\nexport { x }"},
			{Code: "if (true) { const x = 1 }\nexport { x }"},
			{Code: "switch (1) { case 1: const x = 1; break; }\nexport { x }"},
			// var inside function boundary does NOT hoist to module scope
			{Code: "function foo() { var x = 1 }\nexport { x }"},
			{Code: "const foo = () => { var x = 1 }\nexport { x }"},
			{Code: "(function() { var x = 1 })()\nexport { x }"},

			// === Import bindings (immutable, should NOT report) ===
			{Code: "import x from './first'\nexport { x }"},
			{Code: "import * as x from './first'\nexport { x }"},

			// === Declaration merging: all immutable ===
			{Code: "function Foo() {}\nnamespace Foo { export const bar = 1 }\nexport { Foo }"},
			{Code: "const Foo = 1\nnamespace Foo {}\nexport { Foo }"},
			{Code: "class Foo {}\nnamespace Foo {}\nexport { Foo }"},
			{Code: "enum Foo { A }\nnamespace Foo {}\nexport { Foo }"},

			// === Non-identifier default export expressions (not looked up) ===
			{Code: "let x = 1\nexport default x + 1"},
			{Code: "let x = 1\nexport default foo(x)"},
			{Code: "let x = 1\nexport default x!"},
			{Code: "let x = 1\nexport default x as number"},
			{Code: "let x = 1\nexport default x satisfies number"},

			// === CommonJS (not handled by this rule) ===
			{Code: "var x = 1\nmodule.exports = x"},
			{Code: "let x = 1\nmodule.exports = { x }"},

			// === String literal export name with const ===
			{Code: `const x = 1; export { x as "foo-bar" }`},
		},
		[]rule_tester.InvalidTestCase{
			// === Direct export let/var ===
			{
				Code: `export let count = 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 8},
				},
			},
			{
				Code: `export var count = 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 8},
				},
			},
			// Multiple declarations in one statement (single error on the declaration list)
			{
				Code: `export let a = 1, b = 2`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 8},
				},
			},
			{
				Code: `export var a = 1, b = 2`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 8},
				},
			},
			// Direct export with destructuring
			{
				Code: `export let { a, b } = obj`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 8},
				},
			},
			{
				Code: `export var { a, b } = obj`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 8},
				},
			},
			{
				Code: `export let [a, b] = arr`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 8},
				},
			},
			{
				Code: `export var [a, b] = arr`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 8},
				},
			},
			// TypeScript: let with type annotation
			{
				Code: `export let x: string = "hello"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 8},
				},
			},
			// TypeScript: declare let export
			{
				Code: `export declare let x: number`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 16},
				},
			},

			// === Named exports referencing let/var ===
			{
				Code: "let count = 1\nexport { count }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			{
				Code: "var count = 1\nexport { count }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Named export with alias
			{
				Code: "let count = 1\nexport { count as counter }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			{
				Code: "var count = 1\nexport { count as counter }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Named export as default
			{
				Code: "let count = 1\nexport { count as default }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},

			// === Default export of let/var ===
			{
				Code: "let count = 1\nexport default count",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			{
				Code: "var count = 1\nexport default count",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Uninitialized let/var
			{
				Code: "let x\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			{
				Code: "var x\nexport default x",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},

			// === Destructuring then named/default export ===
			// Object destructuring with let
			{
				Code: "let { a } = obj\nexport { a }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Array destructuring with let
			{
				Code: "let [a] = arr\nexport { a }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Nested destructuring with var
			{
				Code: "var { a: { b } } = obj\nexport { b }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Destructured var then default export
			{
				Code: "var { x } = obj\nexport default x",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Rest destructuring with let
			{
				Code: "let { ...rest } = obj\nexport { rest }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Renamed destructuring with let
			{
				Code: "let { a: b } = obj\nexport { b }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Default value destructuring with var
			{
				Code: "var { a = 1 } = obj\nexport { a }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Array destructuring with skip + let
			{
				Code: "let [, b] = arr\nexport { b }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Array rest destructuring with var
			{
				Code: "var [a, ...rest] = arr\nexport { rest }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Deeply nested destructuring
			{
				Code: "let { a: { b: [c] } } = obj\nexport { c }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},

			// === Multiple errors ===
			// Multiple specifiers from separate declarations
			{
				Code: "let a = 1\nlet b = 2\nexport { a, b }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
					{MessageId: "noMutableExports", Line: 2, Column: 1},
				},
			},
			// Mixed specifiers: only the let one is reported
			{
				Code: "let a = 1\nconst b = 2\nexport { a, b }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Same variable exported via direct + named → two errors on same declaration
			{
				Code: "export let x = 1\nexport { x as y }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 8},
					{MessageId: "noMutableExports", Line: 1, Column: 8},
				},
			},
			// Same variable exported as multiple aliases
			{
				Code: "let x = 1\nexport { x, x as y, x as z }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
					{MessageId: "noMutableExports", Line: 1, Column: 1},
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},

			// === Declaration after export (var hoists) ===
			{
				Code: "export { x }\nvar x = 1",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 2, Column: 1},
				},
			},

			// === ES2022: string literal export name with let ===
			{
				Code: `let x = 1; export { x as "foo-bar" }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			// Type assertion doesn't change mutability (still let)
			{
				Code: "let x = 1 as const\nexport default x",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},

			// === Hoisted var (var inside block hoists to module scope) ===
			{
				Code: "{ var x = 1 }\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports"},
				},
			},
			{
				Code: "if (true) { var x = 1 }\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports"},
				},
			},
			{
				Code: "for (var x = 0; x < 1; x++) {}\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports"},
				},
			},
			{
				Code: "for (var x of []) {}\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports"},
				},
			},
			{
				Code: "for (var x in {}) {}\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports"},
				},
			},
			{
				Code: "try { var x = 1 } catch {}\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports"},
				},
			},
			{
				Code: "{ { { var x = 1 } } }\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports"},
				},
			},
			{
				Code: "{ var x = 1 }\nexport default x",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports"},
				},
			},
			// switch / while / do-while / label / else / catch / finally
			{
				Code: "switch (1) { case 1: var x = 1; break; }\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},
			{
				Code: "while (false) { var x = 1 }\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},
			{
				Code: "do { var x = 1 } while (false)\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},
			{
				Code: "label: var x = 1\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},
			{
				Code: "if (false) {} else { var x = 1 }\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},
			{
				Code: "try {} catch (e) { var x = 1 }\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},
			{
				Code: "try {} finally { var x = 1 }\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},
			// Hoisted var with alias
			{
				Code: "if (true) { var x = 1 }\nexport { x as y }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},
			// Hoisted var with default export
			{
				Code: "switch (1) { default: var x = 1 }\nexport default x",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},

			// === Declaration merging: mutable binding ===
			{
				Code: "var Foo = 1\nnamespace Foo {}\nexport { Foo }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},
			{
				Code: "let Foo: any\ninterface Foo {}\nexport { Foo }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},

			// === declare var/let (ambient but mutable binding) ===
			{
				Code: "declare var x: number\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},
			{
				Code: "declare let x: number\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},

			// === var redeclaration ===
			{
				Code: "var x = 1\nvar x = 2\nexport { x }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMutableExports"}},
			},

			// === Parenthesized export default ===
			{
				Code: "let x = 1\nexport default (x)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
			{
				Code: "let x = 1\nexport default (((x)))",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noMutableExports", Line: 1, Column: 1},
				},
			},
		},
	)
}
