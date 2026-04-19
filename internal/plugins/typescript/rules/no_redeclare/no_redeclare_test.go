package no_redeclare

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRedeclareRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRedeclareRule,
		[]rule_tester.ValidTestCase{
			// ====================================================================
			// Separate scopes — no redeclaration.
			// ====================================================================
			{Code: "var a = 3;\nvar b = function () {\n  var a = 10;\n};"},
			{Code: "var a = 3;\na = 10;"},
			{Code: "if (true) {\n  let b = 2;\n} else {\n  let b = 3;\n}"},
			{Code: "var a = 3;\n{\n  let a = 10;\n}"},
			{Code: "function a() {}\nfunction b() {\n  function a() {}\n}"},
			{Code: "class A {}\nfunction foo() {\n  class A {}\n}"},
			// Arrow function body is a separate scope.
			{Code: "const a = 1;\nconst fn = () => { const a = 2; };"},
			// Method body is a separate scope.
			{Code: "class C { method() { let x = 1; let y = 2; } }\nlet x = 1;"},
			// Constructor body is a separate scope.
			{Code: "class C { constructor() { let v = 1; } method() { let v = 2; } }"},
			// Static block is its own scope.
			{Code: "class C { static { let x = 1; } static { let x = 2; } }"},
			// Getter / setter — each a separate scope.
			{Code: "class C { get x() { let a = 1; return a; } set x(v) { let a = 2; } }"},
			// FunctionExpression with its own name — the name is only visible
			// inside the expression, so no conflict with outer `f`.
			{Code: "const f = function f() { return 0; };"},
			// for-let — the init has its own scope independent of the outer.
			{Code: "let i = 0;\nfor (let i = 0; i < 10; i++) {}"},
			// for-of / for-in let
			{Code: "let x = 0;\nfor (let x of []) {}\nfor (let x in {}) {}"},
			// Catch clause variable is scoped to the catch block.
			{Code: "let e = 1;\ntry {} catch (e) {}"},
			// Different branches of try / catch / finally.
			{Code: "try { let a = 1; } catch { let a = 2; } finally { let a = 3; }"},
			// Labeled statement does not introduce a scope for its inner statement.
			{Code: "let a = 1;\nlabel: {\n  let a = 2;\n}"},
			// Declaration file ambient declarations merge happily.
			{Code: "declare var a: number;\ndeclare var b: string;"},

			// ====================================================================
			// TypeScript overload signatures / generic parameters.
			// ====================================================================
			{Code: "function a(): string;\nfunction a(): number;\nfunction a() {}"},
			{Code: "function A<T>() {}\ninterface B<T> {}\ntype C<T> = Array<T>;\nclass D<T> {}"},
			// Ambient method overloads (bodyless declarations inside a class) do
			// not create duplicate entries.
			{Code: "class A {\n  foo(x: string): void;\n  foo(x: number): void;\n  foo(x: unknown) {}\n}"},

			// ====================================================================
			// builtinGlobals option.
			// ====================================================================
			{Code: "var Object = 0;", Options: map[string]interface{}{"builtinGlobals": false}},
			// `self` / `top` live in lib.dom; `builtinGlobals: false` suppresses.
			{Code: "var self = 1;", Options: map[string]interface{}{"builtinGlobals": false}},
			{Code: "var top = 0;", Options: map[string]interface{}{"builtinGlobals": false}},
			// Shadowing a builtin inside a function scope is fine: the function
			// introduces a new scope, and builtinGlobals only applies to the
			// global scope.
			{Code: "function f() { var Object = 0; }", Options: map[string]interface{}{"builtinGlobals": true}},
			// In a module (any file with a top-level `export`), top-level
			// declarations are module-scoped and do not merge with lib globals.
			{Code: "export {};\nvar Object = 0;", Options: map[string]interface{}{"builtinGlobals": true}},
			{Code: "import {} from './foo';\nvar Array = 0;", Options: map[string]interface{}{"builtinGlobals": true}},

			// ====================================================================
			// TypeScript declaration merging (default ignoreDeclarationMerge: true).
			// ====================================================================
			{Code: "interface A {}\ninterface A {}"},
			{Code: "interface A {}\nclass A {}"},
			{Code: "class A {}\nnamespace A {}"},
			{Code: "interface A {}\nclass A {}\nnamespace A {}"},
			{Code: "enum A {}\nnamespace A {}"},
			{Code: "function A() {}\nnamespace A {}"},
			{Code: "namespace A {}\nnamespace A {}"},
			// Order-insensitive merging.
			{Code: "namespace A {}\nclass A {}"},
			{Code: "namespace A {}\nfunction A() {}"},
			{Code: "namespace A {}\nenum A {}"},
			// Multiple non-conflicting interfaces.
			{Code: "interface A {}\ninterface A {}\ninterface A {}"},
			// Multiple non-conflicting namespaces.
			{Code: "namespace A {}\nnamespace A {}\nnamespace A {}"},
			// Explicit option echoes the default.
			{
				Code:    "interface A {}\ninterface A {}",
				Options: map[string]interface{}{"ignoreDeclarationMerge": true},
			},
			{
				Code:    "function A() {}\nnamespace A {}",
				Options: map[string]interface{}{"ignoreDeclarationMerge": true},
			},

			// ====================================================================
			// Namespaces isolate their own inner hoist scope.
			// ====================================================================
			{Code: "namespace A {\n  var x = 1;\n}\nnamespace B {\n  var x = 2;\n}"},

			// ====================================================================
			// Imports — no redeclaration across different names.
			// ====================================================================
			{Code: "import { a, b } from './foo';\nlet c = 1;"},
			{Code: "import a from './foo';\nimport type { b } from './bar';"},

			// ====================================================================
			// Function parameters with destructuring — no outer binding named
			// the same; the rule should not over-report.
			// ====================================================================
			{Code: "function foo({ bar }: { bar: string }) {\n  console.log(bar);\n}"},
			// Each function introduces an independent parameter scope.
			{Code: "function a(x: number) {}\nfunction b(x: number) {}"},
			// Nested destructuring — distinct binding names.
			{Code: "var { a: { b }, c } = { a: { b: 1 }, c: 2 };"},
			// Array destructuring — distinct names.
			{Code: "var [a, b] = [1, 2];"},
			// Default parameter value.
			{Code: "function f(x = 1, y = 2) { return x + y; }"},
			// Rest parameter is a distinct binding.
			{Code: "function f(x: number, ...rest: number[]) {}"},
			// Generator function body is a separate scope.
			{Code: "function* g() { var x = 1; }\nvar x = 2;"},
			// Async function body is a separate scope.
			{Code: "async function g() { var x = 1; }\nvar x = 2;"},
			// var inside catch does not conflict with the catch parameter in
			// ESLint's scope model either — ours just treats them as separate
			// scopes.
			{Code: "try {} catch (e) { var e = 1; }"},

			// ====================================================================
			// More scope boundaries.
			// ====================================================================
			// Setter body is a fresh scope — does not clash with outer bindings.
			{Code: "let a = 1;\nclass C { set x(v) { let a = 2; } }"},
			// Setter parameter is scoped to the setter.
			{Code: "let v = 1;\nclass C { set x(v) {} }"},
			// `import type` of a name, distinct from a later unrelated local.
			{Code: "import type { A } from './foo';\nlet B = 1;"},
			// Separate layers: outer var hoists, inner let shadows it.
			{Code: "for (var a = 0; a < 1; a++) {\n  let a = 1;\n}"},
			// Nested namespace bodies are independent scopes.
			{Code: "namespace N {\n  namespace Inner {\n    var x = 1;\n  }\n  namespace Inner2 {\n    var x = 2;\n  }\n}"},
			// Re-export does not declare a local binding.
			{Code: "export { foo } from './foo';\nlet foo = 1;"},
			// ClassExpression body is a separate scope.
			{Code: "let x = 1;\nconst C = class { m() { let x = 2; } };"},
			// Ambient declaration only — no redeclaration.
			{Code: "declare namespace N {\n  function foo(): void;\n  function foo(): string;\n}"},

			// ====================================================================
			// Module augmentation — string-literal module name does not create
			// a redeclaration. Multiple `declare module 'foo' { ... }` blocks
			// must be allowed (TS declaration merging for modules).
			// ====================================================================
			{Code: "declare module 'foo' { export const x: number; }\ndeclare module 'foo' { export const y: number; }"},
			{Code: "declare module '*.css' { const cls: Record<string, string>; export default cls; }\ndeclare module '*.png' { const url: string; export default url; }"},

			// ====================================================================
			// Type parameters introduce bindings that are scoped to their owner
			// and must never leak into the enclosing scope.
			// ====================================================================
			// Outer `type T` does not collide with a function's type parameter T.
			{Code: "type T = 1;\nfunction f<T>(): T { return null as any; }"},
			// Same type parameter name across sibling functions / classes.
			{Code: "function f<T>() {}\nfunction g<T>() {}"},
			{Code: "class A<T> {}\nclass B<T> {}"},
			// `infer U` inside a conditional type does not clash with an outer U.
			{Code: "type U = 1;\ntype Inner<X> = X extends infer U ? U : never;"},
			// Mapped-type key parameter does not clash with an outer K.
			{Code: "type K = 1;\ntype M<T> = { [K in keyof T]: T[K] };"},

			// ====================================================================
			// Label names live in a separate namespace — they do not collide
			// with variable bindings.
			// ====================================================================
			{Code: "var x = 1;\nx: while (true) { break x; }"},

			// ====================================================================
			// A class expression's internal name is scoped to the expression;
			// an outer class with the same name is not a redeclaration.
			// ====================================================================
			{Code: "const C = class C { m() { return C; } };"},
			{Code: "class C {}\nconst D = class C {};"},

			// ====================================================================
			// `this` parameter does not introduce a runtime binding.
			// ====================================================================
			{Code: "function f(this: unknown, y: number) {}\nfunction g(this: unknown, y: number) {}"},

			// ====================================================================
			// Declaration merging inside a namespace body behaves the same as
			// at program scope.
			// ====================================================================
			{Code: "namespace Outer {\n  interface A {}\n  interface A {}\n}"},
			{Code: "namespace Outer {\n  class A {}\n  namespace A {}\n}"},
			{Code: "namespace Outer {\n  function A() {}\n  namespace A {}\n}"},
			{Code: "namespace Outer {\n  enum A {}\n  namespace A {}\n}"},

			// ====================================================================
			// Decorator prefix on a class does not itself change the binding.
			// ====================================================================
			{Code: "function d(_: unknown) {}\n@d class A {}\nclass B {}"},

			// ====================================================================
			// `export` modifier does not by itself create a second binding —
			// declaration-merge rules still apply.
			// ====================================================================
			{Code: "export interface A {}\nexport interface A {}"},
			{Code: "export class A {}\nexport namespace A {}"},
			{Code: "export function A() {}\nexport namespace A {}"},
			{Code: "export enum A {}\nexport namespace A {}"},
			{Code: "export namespace A {}\nexport namespace A {}"},

			// ====================================================================
			// `export` of distinct names is always fine.
			// ====================================================================
			{Code: "export const a = 1, b = 2;\nexport const c = 3;"},

			// ====================================================================
			// `export { x }` / `export { x as y }` / `export * [as ns]` do NOT
			// create new local bindings, so they cannot conflict with local
			// declarations of the same (local) name.
			// ====================================================================
			{Code: "const foo = 1;\nexport { foo };"},
			{Code: "const foo = 1;\nexport { foo as bar };\nconst bar = 2;"},
			{Code: "export * from './a';\nexport * from './b';"},
			{Code: "export * as ns from './a';\nconst ns = 1;"},

			// ====================================================================
			// Anonymous `export default` produces no local binding.
			// ====================================================================
			{Code: "export default function () {};\nfunction foo() {}"},
			{Code: "export default class {};\nclass Foo {}"},
			{Code: "export default 42;\nconst foo = 1;"},

			// ====================================================================
			// Documented divergence #2 (see no_redeclare.md):
			// A user `type T = ...` never collides with a same-named lib
			// `interface T` because TypeScript keeps type aliases and
			// interfaces in different declaration spaces. ESLint, which does
			// name-level matching, flags these. We intentionally don't.
			// ====================================================================
			{Code: "type NodeListOf = 1;"},                      // lib.dom
			{Code: "type HTMLElement = 1;"},                     // lib.dom
			{Code: "type Array = 1;", Options: map[string]interface{}{"builtinGlobals": true}}, // lib.es5 core interface
			// Extending a lib interface with a same-named user interface is
			// the idiomatic TS pattern (ImportMeta augmentation, etc.) — it
			// must not be reported as a builtin redeclaration.
			{Code: "interface ImportMeta { foo: 1; }"},
			{Code: "interface Array<T> { custom(): T; }", Options: map[string]interface{}{"builtinGlobals": true}},
			// Sanity check: user-code `type T + interface T` still reports —
			// the divergence only covers collisions with *lib* interfaces.
		},
		[]rule_tester.InvalidTestCase{
			// ====================================================================
			// var — basic redeclarations.
			// ====================================================================
			{
				Code: "var a = 3;\nvar a = 10;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Message: "'a' is already defined.", Line: 2, Column: 5},
				},
			},
			{
				Code: "var a = {};\nvar a = [];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},
			{
				Code: "var a = 3;\nvar a = 10;\nvar a = 15;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
					{MessageId: "redeclared", Line: 3, Column: 5},
				},
			},
			// var in the same statement — two bindings in one VariableDeclarationList.
			{
				Code: "var a = 1, a = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 1, Column: 12},
				},
			},

			// ====================================================================
			// Hoisting across nested non-function structures.
			// ====================================================================
			// var hoists out of a switch.
			{
				Code: "switch (foo) {\n  case a:\n    var b = 3;\n  case b:\n    var b = 4;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 5, Column: 9},
				},
			},
			// var hoists out of if/else branches.
			{
				Code: "if (x) {\n  var a = 1;\n} else {\n  var a = 2;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 4, Column: 7},
				},
			},
			// var hoists out of a try/catch/finally.
			{
				Code: "try {\n  var a = 1;\n} catch (e) {\n  var a = 2;\n} finally {\n  var a = 3;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 4, Column: 7},
					{MessageId: "redeclared", Line: 6, Column: 7},
				},
			},
			// var hoists out of a while / do-while body.
			{
				Code: "while (x) {\n  var a = 1;\n}\nvar a = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 4, Column: 5},
				},
			},
			{
				Code: "do {\n  var a = 1;\n} while (x);\nvar a = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 4, Column: 5},
				},
			},
			// var in a for-statement initializer hoists.
			{
				Code: "for (var i = 0; i < 10; i++) {}\nvar i = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},
			// var in a for-of / for-in initializer hoists.
			{
				Code: "for (var x of []) {}\nvar x = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},
			{
				Code: "for (var x in {}) {}\nvar x = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},
			// var hoists past a labeled statement.
			{
				Code: "label: var a = 1;\nvar a = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},

			// ====================================================================
			// var vs function.
			// ====================================================================
			{
				Code: "var a;\nfunction a() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 10},
				},
			},
			{
				Code: "function a() {}\nfunction a() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 10},
				},
			},
			{
				Code: "var a = function () {};\nvar a = function () {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},
			// A class at program scope colliding with a var.
			{
				Code: "var A;\nclass A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},

			// ====================================================================
			// Block scope (let/const).
			// ====================================================================
			{
				Code: "{\n  let a = 1;\n  let a = 2;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 7},
				},
			},
			{
				Code: "{\n  const a = 1;\n  const a = 2;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 9},
				},
			},
			// Top-level let/const twice is also a redeclaration.
			{
				Code: "let a = 1;\nlet a = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},
			{
				Code: "const a = 1;\nconst a = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},
			// let + const → still a redeclaration.
			{
				Code: "let a = 1;\nconst a = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},
			// let + class.
			{
				Code: "let A = 1;\nclass A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},

			// ====================================================================
			// for-statement block scope.
			// ====================================================================
			// Two let in the same for init.
			{
				Code: "for (let i = 0, i = 1; ; ) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 1, Column: 17},
				},
			},

			// ====================================================================
			// Nested function / arrow scopes.
			// ====================================================================
			// Nested function body collects its own var set.
			{
				Code: "function f() {\n  var a;\n  var a;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 7},
				},
			},
			// Nested arrow function body.
			{
				Code: "const f = () => {\n  var a;\n  var a;\n};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 7},
				},
			},
			// Method body.
			{
				Code: "class C {\n  m() {\n    var a;\n    var a;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 4, Column: 9},
				},
			},
			// Static block.
			{
				Code: "class C {\n  static {\n    let a = 1;\n    let a = 2;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 4, Column: 9},
				},
			},
			// Getter body.
			{
				Code: "class C {\n  get x() {\n    let a = 1;\n    let a = 2;\n    return a;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 4, Column: 9},
				},
			},

			// ====================================================================
			// Namespace body — hoist scope like a function.
			// ====================================================================
			{
				Code: "namespace A {\n  var x = 1;\n  var x = 2;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 7},
				},
			},
			{
				Code: "namespace A {\n  type T = 1;\n  type T = 2;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 8},
				},
			},

			// ====================================================================
			// Builtin globals — unique coverage from lib.*.d.ts.
			// ====================================================================
			{
				Code: "var Object = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclaredAsBuiltin", Message: "'Object' is already defined as a built-in global variable.", Line: 1, Column: 5},
				},
				Options: map[string]interface{}{"builtinGlobals": true},
			},
			// Default options already enable builtinGlobals.
			{
				Code: "var Number = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclaredAsBuiltin", Line: 1, Column: 5},
				},
			},
			// `Promise` lives in lib.es2015.promise — covered by TS lib detection.
			{
				Code: "var Promise = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclaredAsBuiltin", Line: 1, Column: 5},
				},
			},
			// ====================================================================
			// Documented divergence #1 (see no_redeclare.md):
			// `builtinGlobals` resolves against the project's `lib` configuration
			// — DOM names and ES-extension additions count, not just the ES
			// core subset ESLint pre-declares. These cases lock in that we
			// detect non-ES-core lib globals the upstream rule would only see
			// with an explicit `globals` / `env` override.
			// ====================================================================
			// DOM value global (lib.dom).
			{
				Code: "var top = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclaredAsBuiltin", Line: 1, Column: 5},
				},
			},
			{
				Code: "var self = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclaredAsBuiltin", Line: 1, Column: 5},
				},
			},
			// Another DOM-only value binding.
			{
				Code: "var console = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclaredAsBuiltin", Line: 1, Column: 5},
				},
			},
			// Another DOM-only var binding (`declare var document: Document`).
			{
				Code: "var document = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclaredAsBuiltin", Line: 1, Column: 5},
				},
			},
			// Destructuring — one user+user conflict and one user+builtin.
			{
				Code: "var a;\nvar { a = 0, b: Object = 0 } = {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
					{MessageId: "redeclaredAsBuiltin", Line: 2, Column: 17},
				},
				Options: map[string]interface{}{"builtinGlobals": true},
			},

			// ====================================================================
			// Type aliases and mixed type-space / value-space redeclarations.
			// ====================================================================
			{
				Code: "type T = 1;\ntype T = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 6},
				},
			},
			{
				Code: "type something = string;\nconst something = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},
			// Type + interface with the same name — two entries in the same
			// "type space". Upstream treats these as redeclarations regardless
			// of ignoreDeclarationMerge.
			{
				Code: "type A = number;\ninterface A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 11},
				},
			},

			// ====================================================================
			// Declaration merging — ignoreDeclarationMerge: false reports each
			// combination as a redeclaration.
			// ====================================================================
			{
				Code: "interface A {}\ninterface A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 11},
				},
				Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			},
			{
				Code: "interface A {}\nclass A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
				Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			},
			{
				Code: "class A {}\nnamespace A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 11},
				},
				Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			},
			{
				Code: "interface A {}\nclass A {}\nnamespace A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
					{MessageId: "redeclared", Line: 3, Column: 11},
				},
				Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			},
			{
				Code: "function A() {}\nnamespace A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 11},
				},
				Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			},
			{
				Code: "function A() {}\nclass A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
				Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			},
			{
				Code: "function A() {}\nclass A {}\nnamespace A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
					{MessageId: "redeclared", Line: 3, Column: 11},
				},
				Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			},

			// ====================================================================
			// Declaration merging limits (ignoreDeclarationMerge: true) — merge
			// rules are violated when there's more than one of the "lead" kind.
			// ====================================================================
			// Two classes + namespace: only the two classes count as the redeclaration.
			{
				Code: "class A {}\nclass A {}\nnamespace A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
				Options: map[string]interface{}{"ignoreDeclarationMerge": true},
			},
			// Two functions + namespace: only two functions count.
			{
				Code: "function A() {}\nfunction A() {}\nnamespace A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 10},
				},
				Options: map[string]interface{}{"ignoreDeclarationMerge": true},
			},
			// Two enums + namespace: only enums count.
			{
				Code: "enum A {}\nnamespace A {}\nenum A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 6},
				},
				Options: map[string]interface{}{"ignoreDeclarationMerge": true},
			},

			// ====================================================================
			// Imports.
			// ====================================================================
			{
				Code: "import { a } from './foo';\nvar a;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},
			// Default import vs a later var.
			{
				Code: "import a from './foo';\nconst a = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},
			// Namespace import vs a later var.
			{
				Code: "import * as ns from './foo';\nconst ns = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},
			// import-equals redeclaration.
			{
				Code: "import eq = require('./foo');\nvar eq;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},

			// ====================================================================
			// Destructuring — bindings in a single declaration list.
			// ====================================================================
			// Two entries of the same name in the same destructuring pattern.
			{
				Code: "var { a, b: a } = { a: 1, b: 2 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 1, Column: 13},
				},
			},
			// Nested destructuring redeclaration.
			{
				Code: "var { a: { b } } = { a: { b: 1 } };\nvar b;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},
			// Array + object mixed destructuring.
			{
				Code: "var [{ a }, a] = [{ a: 1 }, 2];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 1, Column: 13},
				},
			},

			// ====================================================================
			// Setter body.
			// ====================================================================
			{
				Code: "class C {\n  set x(v) {\n    let a = 1;\n    let a = 2;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 4, Column: 9},
				},
			},

			// ====================================================================
			// Nested namespace — each body has its own scope.
			// ====================================================================
			{
				Code: "namespace N {\n  namespace Inner {\n    var x = 1;\n    var x = 2;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 4, Column: 9},
				},
			},

			// ====================================================================
			// Parameter vs body var — ESLint reports this as redeclaration.
			// ====================================================================
			{
				Code: "function f(x: number) {\n  var x = 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},

			// ====================================================================
			// Type collisions outside the interface-merge channel.
			// ====================================================================
			// type alias + class — no valid merge.
			{
				Code: "type A = number;\nclass A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},
			// Two enums in the same scope — never a valid merge.
			{
				Code: "enum E { a }\nenum E { b }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 6},
				},
			},

			// ====================================================================
			// import type collisions.
			// ====================================================================
			{
				Code: "import type { X } from './foo';\ntype X = number;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 6},
				},
			},

			// ====================================================================
			// export default function + later var.
			// ====================================================================
			{
				Code: "export default function foo() {}\nvar foo;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5},
				},
			},

			// ====================================================================
			// `export` modifier does not shield against redeclarations —
			// the underlying declarations still count.
			// ====================================================================
			{
				Code: "export var a = 1;\nexport var a = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 12},
				},
			},
			{
				Code: "export function f() {}\nexport function f() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 17},
				},
			},
			{
				Code: "export class A {}\nexport class A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 14},
				},
			},
			{
				Code: "export type T = 1;\nexport type T = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 13},
				},
			},
			{
				Code: "export enum E { a }\nexport enum E { b }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 13},
				},
			},
			// `export default` + explicit local of the same name.
			{
				Code: "export default function foo() {}\nfunction foo() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 10},
				},
			},
			{
				Code: "export default class A {}\nclass A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},

			// ====================================================================
			// Diagnostic range — lock that every messageId reports the
			// full identifier span (EndLine/EndColumn), not just the start.
			// ====================================================================
			{
				Code: "var hello = 1;\nvar hello = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 5, EndLine: 2, EndColumn: 10},
				},
			},
			{
				Code: "var Object = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclaredAsBuiltin", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
				},
				Options: map[string]interface{}{"builtinGlobals": true},
			},

			// ====================================================================
			// Decorator does not affect redeclaration detection.
			// ====================================================================
			{
				Code: "function d(_: unknown) {}\n@d class A {}\nclass A {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 7},
				},
			},

			// ====================================================================
			// Declaration merging is applied inside a namespace body the same
			// way as at program scope.
			// ====================================================================
			{
				Code: "namespace Outer {\n  class A {}\n  class A {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 9},
				},
			},

			// ====================================================================
			// Class expression + outer same-named declaration — the outer
			// declaration still counts (two program-scope bindings).
			// ====================================================================
			{
				Code: "class C {}\nconst C = class C {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 7},
				},
			},

			// ====================================================================
			// Ambient `declare var` repeated at the top level.
			// ====================================================================
			{
				Code: "declare var x: number;\ndeclare var x: string;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 2, Column: 13},
				},
			},

			// ====================================================================
			// `using` / `await using` redeclarations (block-scoped like let/const).
			// ====================================================================
			{
				Code: "{ using a = null as any; using a = null as any; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 1, Column: 32},
				},
			},
			{
				Code: "async function f() {\n  await using a = null as any;\n  await using a = null as any;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 15},
				},
			},

			// ====================================================================
			// Async / generator functions — still report inside their bodies.
			// ====================================================================
			{
				Code: "async function g() {\n  var x = 1;\n  var x = 2;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 7},
				},
			},
			{
				Code: "function* g() {\n  var x = 1;\n  var x = 2;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 3, Column: 7},
				},
			},
		},
	)
}
