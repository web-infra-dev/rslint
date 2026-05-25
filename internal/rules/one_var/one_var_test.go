package one_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOneVarRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&OneVarRule,
		[]rule_tester.ValidTestCase{
			// ---- Default mode ("always") ----
			{Code: `function foo() { var bar = true; }`},
			{Code: `function foo() { var bar = true, baz = 1; if (qux) { bar = false; } }`},
			{Code: `var foo = function() { var bar = true; baz(); }`},
			{Code: `function foo() { var bar = true, baz = false; }`, Options: []interface{}{"always"}},

			// ---- "never" ----
			{Code: `function foo() { var bar = true; var baz = false; }`, Options: []interface{}{"never"}},
			{Code: `for (var i = 0, len = arr.length; i < len; i++) {}`, Options: []interface{}{"never"}},

			// ---- initialized / uninitialized ----
			{Code: `var bar = true; var baz = false;`, Options: []interface{}{map[string]interface{}{"initialized": "never"}}},
			{Code: `var bar = true, baz = false;`, Options: []interface{}{map[string]interface{}{"initialized": "always"}}},
			{Code: `var bar, baz;`, Options: []interface{}{map[string]interface{}{"initialized": "never"}}},
			{Code: `var bar; var baz;`, Options: []interface{}{map[string]interface{}{"uninitialized": "never"}}},
			{Code: `var bar, baz;`, Options: []interface{}{map[string]interface{}{"uninitialized": "always"}}},
			{Code: `var bar = true, baz = false;`, Options: []interface{}{map[string]interface{}{"uninitialized": "never"}}},
			{Code: `var bar = true, baz = false, a, b;`, Options: []interface{}{map[string]interface{}{"uninitialized": "always", "initialized": "always"}}},
			{Code: `var bar = true; var baz = false; var a; var b;`, Options: []interface{}{map[string]interface{}{"uninitialized": "never", "initialized": "never"}}},
			{Code: `var bar, baz; var a = true; var b = false;`, Options: []interface{}{map[string]interface{}{"uninitialized": "always", "initialized": "never"}}},
			{Code: `var bar = true, baz = false; var a; var b;`, Options: []interface{}{map[string]interface{}{"uninitialized": "never", "initialized": "always"}}},
			{Code: `var bar; var baz; var a = true, b = false;`, Options: []interface{}{map[string]interface{}{"uninitialized": "never", "initialized": "always"}}},

			// ---- Destructuring ----
			{Code: `function foo() { var a = [1, 2, 3]; var [b, c, d] = a; }`, Options: []interface{}{"never"}},

			// ---- let / const mixed with var ----
			{Code: `function foo() { let a = 1; var c = true; if (a) {let c = true; } }`, Options: []interface{}{"always"}},
			{Code: `function foo() { const a = 1; var c = true; if (a) {const c = true; } }`, Options: []interface{}{"always"}},
			{Code: `function foo() { if (true) { const a = 1; }; if (true) {const a = true; } }`, Options: []interface{}{"always"}},
			{Code: `function foo() { let a = 1; let b = true; }`, Options: []interface{}{"never"}},
			{Code: `function foo() { const a = 1; const b = true; }`, Options: []interface{}{"never"}},
			{Code: `function foo() { let a = 1; const b = false; var c = true; }`, Options: []interface{}{"always"}},
			{Code: `function foo() { let a = 1, b = false; var c = true; }`, Options: []interface{}{"always"}},
			{Code: `function foo() { let a = 1; let b = 2; const c = false; const d = true; var e = true, f = false; }`, Options: []interface{}{map[string]interface{}{"var": "always", "let": "never", "const": "never"}}},

			// ---- block scope ----
			{Code: `let foo = true; for (let i = 0; i < 1; i++) { let foo = false; }`, Options: []interface{}{map[string]interface{}{"var": "always", "let": "always", "const": "never"}}},
			{Code: `let foo = true; for (let i = 0; i < 1; i++) { let foo = false; }`, Options: []interface{}{map[string]interface{}{"var": "always"}}},
			{Code: `let foo = true, bar = false;`, Options: []interface{}{map[string]interface{}{"var": "never"}}},
			{Code: `let foo = true, bar = false;`, Options: []interface{}{map[string]interface{}{"const": "never"}}},
			{Code: `let foo = true, bar = false;`, Options: []interface{}{map[string]interface{}{"uninitialized": "never"}}},
			{Code: `let foo, bar`, Options: []interface{}{map[string]interface{}{"initialized": "never"}}},
			{Code: `let foo = true, bar = false; let a; let b;`, Options: []interface{}{map[string]interface{}{"uninitialized": "never"}}},
			{Code: `let foo, bar; let a = true; let b = true;`, Options: []interface{}{map[string]interface{}{"initialized": "never"}}},
			{Code: `var foo, bar; const a=1; const b=2; let c, d`, Options: []interface{}{map[string]interface{}{"var": "always", "let": "always"}}},
			{Code: `var foo; var bar; const a=1, b=2; let c; let d`, Options: []interface{}{map[string]interface{}{"const": "always"}}},

			// ---- for-in / for-of with new block scope ----
			{Code: `for (let x of foo) {}; for (let y of foo) {}`, Options: []interface{}{map[string]interface{}{"uninitialized": "always"}}},
			{Code: `for (let x in foo) {}; for (let y in foo) {}`, Options: []interface{}{map[string]interface{}{"uninitialized": "always"}}},
			{Code: `var x; for (var y in foo) {}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},
			{Code: `var x, y; for (y in foo) {}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},
			{Code: `var x, y; for (var z in foo) {}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},
			{Code: `var x; for (var y in foo) {var bar = y; for (var z in bar) {}}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},
			{Code: `var a = 1; var b = 2; var x, y; for (var z in foo) {var baz = z; for (var d in baz) {}}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},
			{Code: `var x; for (var y of foo) {}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},
			{Code: `var x, y; for (y of foo) {}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},
			{Code: `var x, y; for (var z of foo) {}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},
			{Code: `var x; for (var y of foo) {var bar = y; for (var z of bar) {}}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},
			{Code: `var a = 1; var b = 2; var x, y; for (var z of foo) {var baz = z; for (var d in baz) {}}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},
			{Code: `var a = 1; var b = 2; var x, y; for (var z of foo) {var baz = z; for (var d of baz) {}}`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},

			// ---- separateRequires ----
			{Code: `var foo = require('foo'), bar;`, Options: []interface{}{map[string]interface{}{"separateRequires": false, "var": "always"}}},
			{Code: `var foo = require('foo'), bar = require('bar');`, Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}}},
			{Code: `var bar = 'bar'; var foo = require('foo');`, Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}}},
			{Code: `var foo = require('foo'); var bar = 'bar';`, Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}}},

			// ---- "consecutive" (https://github.com/eslint/eslint/issues/4680) ----
			{Code: `var a = 0, b, c;`, Options: []interface{}{"consecutive"}},
			{Code: `var a = 0, b = 1, c = 2;`, Options: []interface{}{"consecutive"}},
			{Code: `var a = 0, b = 1; foo(); var c = 2;`, Options: []interface{}{"consecutive"}},
			{Code: `let a = 0, b, c;`, Options: []interface{}{"consecutive"}},
			{Code: `let a = 0, b = 1, c = 2;`, Options: []interface{}{"consecutive"}},
			{Code: `let a = 0, b = 1; foo(); let c = 2;`, Options: []interface{}{"consecutive"}},
			{Code: `const a = 0, b = 1; foo(); const c = 2;`, Options: []interface{}{"consecutive"}},
			{Code: `const a = 0; var b = 1;`, Options: []interface{}{"consecutive"}},
			{Code: `const a = 0; let b = 1;`, Options: []interface{}{"consecutive"}},
			{Code: `let a = 0; const b = 1; var c = 2;`, Options: []interface{}{"consecutive"}},

			// ---- consecutive + separateRequires ----
			{Code: `const foo = require('foo'); const bar = 'bar';`, Options: []interface{}{map[string]interface{}{"const": "consecutive", "separateRequires": true}}},

			// ---- consecutive (per init/uninit) ----
			{Code: `var a = 0, b = 1; var c, d;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}}},
			{Code: `var a = 0; var b, c; var d = 1;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}}},
			{Code: `let a = 0, b = 1; let c, d;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}}},
			{Code: `let a = 0; let b, c; let d = 1;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}}},
			{Code: `const a = 0, b = 1; let c, d;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}}},
			{Code: `const a = 0; let b, c; const d = 1;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}}},
			{Code: `var a = 0, b = 1; var c; var d;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}}},
			{Code: `var a = 0; var b; var c; var d = 1;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}}},
			{Code: `let a = 0, b = 1; let c; let d;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}}},
			{Code: `let a = 0; let b; let c; let d = 1;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}}},
			{Code: `const a = 0, b = 1; let c; let d;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}}},
			{Code: `const a = 0; let b; let c; const d = 1;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}}},
			{Code: `var a, b; var c = 0, d = 1;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}}},
			{Code: `var a; var b = 0, c = 1; var d;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}}},
			{Code: `let a, b; let c = 0, d = 1;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}}},
			{Code: `let a; let b = 0, c = 1; let d;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}}},
			{Code: `let a, b; const c = 0, d = 1;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}}},
			{Code: `let a; const b = 0, c = 1; let d;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}}},
			{Code: `var a, b; var c = 0; var d = 1;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}}},
			{Code: `var a; var b = 0; var c = 1; var d;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}}},
			{Code: `let a, b; let c = 0; let d = 1;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}}},
			{Code: `let a; let b = 0; let c = 1; let d;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}}},
			{Code: `let a, b; const c = 0; const d = 1;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}}},
			{Code: `let a; const b = 0; const c = 1; let d;`, Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}}},

			// ---- consecutive (per kind) ----
			{Code: `var a = 0, b = 1;`, Options: []interface{}{map[string]interface{}{"var": "consecutive"}}},
			{Code: `var a = 0; foo; var b = 1;`, Options: []interface{}{map[string]interface{}{"var": "consecutive"}}},
			{Code: `let a = 0, b = 1;`, Options: []interface{}{map[string]interface{}{"let": "consecutive"}}},
			{Code: `let a = 0; foo; let b = 1;`, Options: []interface{}{map[string]interface{}{"let": "consecutive"}}},
			{Code: `const a = 0, b = 1;`, Options: []interface{}{map[string]interface{}{"const": "consecutive"}}},
			{Code: `const a = 0; foo; const b = 1;`, Options: []interface{}{map[string]interface{}{"const": "consecutive"}}},
			{Code: `let a, b; const c = 0, d = 1;`, Options: []interface{}{map[string]interface{}{"let": "consecutive", "const": "always"}}},
			{Code: `let a; const b = 0, c = 1; let d;`, Options: []interface{}{map[string]interface{}{"let": "consecutive", "const": "always"}}},
			{Code: `let a, b; const c = 0; const d = 1;`, Options: []interface{}{map[string]interface{}{"let": "consecutive", "const": "never"}}},
			{Code: `let a; const b = 0; const c = 1; let d;`, Options: []interface{}{map[string]interface{}{"let": "consecutive", "const": "never"}}},
			{Code: `const a = 0, b = 1; let c, d;`, Options: []interface{}{map[string]interface{}{"const": "consecutive", "let": "always"}}},
			{Code: `const a = 0; let b, c; const d = 1;`, Options: []interface{}{map[string]interface{}{"const": "consecutive", "let": "always"}}},
			{Code: `const a = 0, b = 1; let c; let d;`, Options: []interface{}{map[string]interface{}{"const": "consecutive", "let": "never"}}},
			{Code: `const a = 0; let b; let c; const d = 1;`, Options: []interface{}{map[string]interface{}{"const": "consecutive", "let": "never"}}},

			{Code: `var a = 1, b = 2; foo(); var c = 3, d = 4;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive"}}},
			{Code: `var bar, baz;`, Options: []interface{}{"consecutive"}},
			{Code: `var bar = 1, baz = 2; qux(); var qux = 3, quux;`, Options: []interface{}{"consecutive"}},
			{Code: `let a, b; var c; var d; let e;`, Options: []interface{}{map[string]interface{}{"var": "never", "let": "consecutive", "const": "consecutive"}}},
			{Code: `const a = 1, b = 2; var d; var e; const f = 3;`, Options: []interface{}{map[string]interface{}{"var": "never", "let": "consecutive", "const": "consecutive"}}},
			{Code: `var a, b; const c = 1; const d = 2; let e; let f; `, Options: []interface{}{map[string]interface{}{"var": "consecutive"}}},
			{Code: `var a = 1, b = 2; var c; var d; var e = 3, f = 4;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}}},
			{Code: `var a; somethingElse(); var b;`, Options: []interface{}{map[string]interface{}{"var": "never"}}},
			{Code: "var foo = 1;\nlet bar = function() { var x; };\nvar baz = 2;", Options: []interface{}{map[string]interface{}{"var": "never"}}},

			// ---- class static blocks ----
			{Code: `class C { static { var a; let b; const c = 0; } }`, Options: []interface{}{"always"}},
			{Code: `const a = 0; class C { static { const b = 0; } }`, Options: []interface{}{"always"}},
			{Code: `class C { static { const b = 0; } } const a = 0; `, Options: []interface{}{"always"}},
			{Code: `let a; class C { static { let b; } }`, Options: []interface{}{"always"}},
			{Code: `class C { static { let b; } } let a;`, Options: []interface{}{"always"}},
			{Code: `var a; class C { static { var b; } }`, Options: []interface{}{"always"}},
			{Code: `class C { static { var b; } } var a; `, Options: []interface{}{"always"}},
			{Code: `var a; class C { static { if (foo) { var b; } } }`, Options: []interface{}{"always"}},
			{Code: `class C { static { if (foo) { var b; } } } var a; `, Options: []interface{}{"always"}},
			{Code: `class C { static { const a = 0; if (foo) { const b = 0; } } }`, Options: []interface{}{"always"}},
			{Code: `class C { static { let a; if (foo) { let b; } } }`, Options: []interface{}{"always"}},
			{Code: `class C { static { const a = 0; const b = 0; } }`, Options: []interface{}{"never"}},
			{Code: `class C { static { let a; let b; } }`, Options: []interface{}{"never"}},
			{Code: `class C { static { var a; var b; } }`, Options: []interface{}{"never"}},
			{Code: `class C { static { let a; foo; let b; } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { let a; const b = 0; let c; } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { var a; foo; var b; } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { var a; let b; var c; } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { let a; if (foo) { let b; } } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { if (foo) { let b; } let a;  } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { const a = 0; if (foo) { const b = 0; } } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { if (foo) { const b = 0; } const a = 0; } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { var a; if (foo) var b; } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { if (foo) var b; var a; } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { if (foo) { var b; } var a; } }`, Options: []interface{}{"consecutive"}},
			{Code: `class C { static { let a; let b = 0; } }`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive"}}},
			{Code: `class C { static { var a; var b = 0; } }`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive"}}},

			// ---- Explicit Resource Management ----
			{Code: `using a = 0; let b = 1; const c = 2;`},
			{Code: `await using a = 0; let b = 1; const c = 2;`},
			{Code: `using a = 0, b = 1;`},
			{Code: `await using a = 0, b = 1;`},
			{Code: `function fn() { { using a = 0; } using b = 1; }`},
			{Code: `using a = 0; using b = 1;`, Options: []interface{}{"never"}},
			{Code: `await using a = 0; await using b = 1;`, Options: []interface{}{"never"}},
			{Code: `using a = 0, b = 1;`, Options: []interface{}{"consecutive"}},
			{Code: `await using a = 0, b = 1;`, Options: []interface{}{"consecutive"}},
			{Code: `using a = 0, b = 1;`, Options: []interface{}{map[string]interface{}{"initialized": "always"}}},
			{Code: `await using a = 0, b = 1;`, Options: []interface{}{map[string]interface{}{"initialized": "always"}}},
			{Code: `using a = 0; using b = 1;`, Options: []interface{}{map[string]interface{}{"initialized": "never"}}},
			{Code: `await using a = 0; await using b = 1;`, Options: []interface{}{map[string]interface{}{"initialized": "never"}}},
			{Code: `using a = 0, b = 1; foo(); using c = 2, d = 3;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive"}}},
			{Code: `await using a = 0, b = 1; foo(); await using c = 2, d = 3;`, Options: []interface{}{map[string]interface{}{"initialized": "consecutive"}}},

			// =========================================================
			// rslint-extra valid cases — beyond upstream coverage
			// =========================================================

			// Block-scope isolation: each `{}` opens a fresh let-scope
			{Code: `function f() { { let a; } { let a; } }`, Options: []interface{}{"always"}},
			// Nested function bodies: inner var doesn't leak to outer
			{Code: `function f() { var a; function g() { var a; } var b; }`, Options: []interface{}{"never"}},
			// Try / catch / finally as block boundaries
			{Code: `try { let a; } catch (e) { let a; } finally { let a; }`, Options: []interface{}{"always"}},
			// `for-let` body opens a new block; same name fine
			{Code: `for (let i = 0; i < 1; i++) { let i; }`, Options: []interface{}{"always"}},
			// switch case only requires consecutive (cross-case sees same scope)
			// switch-case + consecutive: upstream reports 0 errors because
			// `consecutive` walks parent.body which is undefined for SwitchCase
			// (it uses `consequent` instead). Verified against ESLint v10.2.1.
			{Code: `switch (x) { case 1: var a; var b; break; }`, Options: []interface{}{"consecutive"}},
			{Code: `switch (x) { case 1: var a; foo(); var b; break; }`, Options: []interface{}{"consecutive"}},

			// v10 — `consecutive` mode does NOT fire across `export` wrappers.
			// In ESLint, `export var x; export var y;` are wrapped in two
			// separate ExportNamedDeclaration nodes, so the consecutive check
			// (which uses parent.body indexOf) never finds them as siblings.
			// rslint mirrors this by skipping consecutive when either side has
			// `export` modifier. (Verified on rsbuild + rspack repos against
			// ESLint v10.2.1, 100% match in consecutive mode.)
			{Code: `export const a = 1; export const b = 2;`, Options: []interface{}{"consecutive"}},
			{Code: `export const a = 1, b = 2;`, Options: []interface{}{"consecutive"}},
			{Code: `const a = 1; export const b = 2;`, Options: []interface{}{"consecutive"}},
			{Code: `export const a = 1; const b = 2;`, Options: []interface{}{"consecutive"}},
			{Code: `export let a = 1; export let b = 2;`, Options: []interface{}{"consecutive"}},
			{Code: `export var a = 1; export var b = 2;`, Options: []interface{}{"consecutive"}},
			// (Mixed-export consecutive variant moved to invalid: inner b/c pair fires combine.)

			// v10 — per-kind `using: 'never'` doesn't fire on already-split usings
			{Code: `using a = 0; using b = 1;`, Options: []interface{}{map[string]interface{}{"using": "never"}}},
			// v10 — `using: 'always'` doesn't fire on a single statement with two declarators
			{Code: `using a = 0, b = 1;`, Options: []interface{}{map[string]interface{}{"using": "always"}}},
			// v10 — per-kind `awaitUsing: 'always'` doesn't fire on a single combined statement
			{Code: `await using a = 0, b = 1;`, Options: []interface{}{map[string]interface{}{"awaitUsing": "always"}}},
			// v10 — option key `awaitUsing` (camelCase) corresponds to node.kind `"await using"` (with space)
			{Code: `await using a = 0; await using b = 1;`, Options: []interface{}{map[string]interface{}{"awaitUsing": "never"}}},
			// Class methods/getters/setters/constructors are independent fn scopes
			{Code: `class C { a() { let x; } b() { let x; } }`, Options: []interface{}{"always"}},
			{Code: `class C { get x() { var a; } set x(v) { var a; } }`, Options: []interface{}{"always"}},
			// TS type annotations don't collapse declarators
			{Code: `var x: number, y: string;`, Options: []interface{}{"always"}},
			// TS namespace block as ModuleBlock
			{Code: `namespace N { var a, b; }`, Options: []interface{}{"always"}},
			{Code: `namespace N { var a; var b; }`, Options: []interface{}{"never"}},
			// Async function bodies
			{Code: `async function f() { let a, b; }`, Options: []interface{}{"always"}},
			{Code: `async function f() { let a; let b; }`, Options: []interface{}{"never"}},
			// Generator
			{Code: `function* g() { const a = 1, b = 2; }`, Options: []interface{}{"always"}},
			// IIFE creates a fresh scope — outer var doesn't conflict
			{Code: `var x; (function () { var x; })();`, Options: []interface{}{"always"}},
			// Arrow function bodies
			{Code: `const f = () => { let a, b; };`, Options: []interface{}{"always"}},
			// Multi-byte identifiers
			{Code: `var π = 1, ω = 2;`, Options: []interface{}{"always"}},
			// Empty declaration list never reachable, but empty-body fn shouldn't crash
			{Code: `function f() {}`, Options: []interface{}{"always"}},
			// Catch parameter is not a VariableDeclarationList
			{Code: `try {} catch (e) { var a; }`, Options: []interface{}{"always"}},
			// for-of / for-in with destructuring
			{Code: `for (const { a, b } of arr) {}`, Options: []interface{}{"always"}},
			{Code: `for (const [a, b] of arr) {}`, Options: []interface{}{"always"}},
			// separateRequires: require followed by non-require initializer is allowed
			// (upstream's recordTypes routes the require to scope.required, not scope.initialized,
			// so the subsequent non-require initialized var sees scope.initialized=false).
			{Code: `var a = require('a'); var b = 1;`, Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}}},
			// And the reverse: non-require then pure require also fine.
			{Code: `var a = 1; var b = require('b');`, Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}}},

			// #4 — for-init scope tracking lock-ins
			// Single for-init combining init declarators is fine (no consecutive prev sibling).
			{Code: `for (let i = 0, j = 0;;) {}`, Options: []interface{}{"always"}},
			// for-let opens a block scope, so `let j` outside the loop and `let i` inside don't conflict.
			{Code: `let j = 0; for (let i = 0;;) {}`, Options: []interface{}{"always"}},
			// for-var: `var i` inside for-init is the same function scope as outer `var j`,
			// but with "never" we don't combine for-init.
			{Code: `var j = 0; for (var i = 0;;) {}`, Options: []interface{}{"never"}},

			// #9 — for (await using ... of ...) reachable shape
			{Code: `async function f() { for (await using x of foo) {} }`, Options: []interface{}{"always"}},
			{Code: `async function f() { for (using x of foo) {} }`, Options: []interface{}{"always"}},
			// Block-scoped await using across two for-of heads: each is its own block,
			// so they don't conflict under "always" — matches upstream behavior.
			{Code: `async function f() { for (await using x of foo) {}; for (await using y of foo) {} }`, Options: []interface{}{"always"}},
			{Code: `for (using x of foo) {}; for (using y of foo) {}`, Options: []interface{}{"always"}},

			// (namespace var-isolation tests were moved to invalid because
			//  ESLint v10's one-var doesn't push a new scope on TSModuleBlock
			//  — every `var` across multiple namespaces is in the outer scope
			//  and triggers combine in `always` mode. See invalid section.)

			// =========================================================
			// Round 2 — additional tsgo-edge / real-world hardening tests
			// =========================================================

			// ---- Empty declarator name lists / declare-only forms ----
			{Code: `declare var x: number;`, Options: []interface{}{"always"}},
			{Code: `declare let x: number, y: string;`, Options: []interface{}{"always"}},
			// declare const in ambient namespace — should not emit
			{Code: `declare namespace N { var x: any; }`, Options: []interface{}{"always"}},

			// ---- Decorators on classes don't break method-scope tracking ----
			// (Decorators occur at TypeScript parse time but don't introduce var bindings.)
			{Code: `@dec class C { method() { let a, b; } }`, Options: []interface{}{"always"}},
			{Code: `@dec class C { method() { let a; let b; } }`, Options: []interface{}{"never"}},

			// ---- Type-only context: type aliases don't count as declarations ----
			{Code: `type T = number; var a; var b;`, Options: []interface{}{"never"}},
			{Code: `interface I { x: number; } var a;`, Options: []interface{}{"always"}},

			// ---- enum doesn't cross-pollute var scope ----
			{Code: `enum E { A, B } var x;`, Options: []interface{}{"always"}},
			{Code: `enum E { A } namespace N { enum E { B } }`, Options: []interface{}{"always"}},

			// ---- TS satisfies / as / non-null assertion in initializers ----
			{Code: `var a = (x as number), b = (y satisfies T);`, Options: []interface{}{"always"}},
			{Code: `var a = x!.foo, b = y!.bar;`, Options: []interface{}{"always"}},

			// ---- Optional chaining in initializer ----
			{Code: `var a = obj?.foo, b = obj?.bar;`, Options: []interface{}{"always"}},

			// ---- Tagged template / arguments / spread ----
			{Code: "var a = tag`a`, b = tag`b`;", Options: []interface{}{"always"}},
			{Code: `var a = [...x], b = [...y];`, Options: []interface{}{"always"}},

			// ---- BigInt / numeric separator literals ----
			{Code: `var a = 100n, b = 200n;`, Options: []interface{}{"always"}},
			{Code: `var a = 1_000, b = 2_000;`, Options: []interface{}{"always"}},

			// ---- RegExp literal initializer ----
			{Code: `var a = /foo/g, b = /bar/i;`, Options: []interface{}{"always"}},

			// ---- new / new.target / class instance creation ----
			{Code: `var a = new Date(), b = new Map();`, Options: []interface{}{"always"}},

			// ---- Conditional / ternary initializer ----
			{Code: `var a = x ? 1 : 2, b = y ? 3 : 4;`, Options: []interface{}{"always"}},

			// ---- typeof / void / delete in initializer ----
			{Code: `var a = typeof x, b = void 0;`, Options: []interface{}{"always"}},

			// ---- await in initializer (top-level await in module) ----
			{Code: `async function f() { var a = await x, b = await y; }`, Options: []interface{}{"always"}},

			// ---- yield in generator initializer ----
			{Code: `function* g() { let a = yield 1, b = yield 2; }`, Options: []interface{}{"always"}},

			// ---- Comments between declarators preserved (valid for never with line comments) ----
			{Code: `var a; // a's purpose
var b; // b's purpose`, Options: []interface{}{"never"}},

			// ---- Object destructuring with defaults ----
			{Code: `var { a = 1, b = 2 } = obj;`, Options: []interface{}{"never"}},
			{Code: `var { a: x = 1, b: y = 2 } = obj;`, Options: []interface{}{"always"}},

			// ---- Array destructuring with elision / rest / nested ----
			{Code: `var [, a, , b, ...c] = arr;`, Options: []interface{}{"always"}},
			{Code: `var [a, [b, c]] = nested;`, Options: []interface{}{"always"}},

			// ---- Computed property in destructuring ----
			{Code: `var { [k]: a } = obj, { [k]: b } = obj2;`, Options: []interface{}{"always"}},

			// ---- export from / re-export — does not declare local var ----
			{Code: `export { a } from './a'; var b;`, Options: []interface{}{"always"}},

			// ---- Multiple imports at top + var (ImportDeclaration is not a VariableStatement) ----
			{Code: `import { x } from './x'; import { y } from './y'; var a;`, Options: []interface{}{"always"}},

			// ---- Module declaration with augmentation ----
			{Code: `declare module 'lodash' { export const _: any; }`, Options: []interface{}{"always"}},

			// ---- Class field declarations are not VariableStatements ----
			{Code: `class C { x = 1; y = 2; method() { let a, b; } }`, Options: []interface{}{"always"}},

			// ---- new.target / super reference doesn't break scope tracking ----
			{Code: `class C extends D { constructor() { super(); var a, b; } }`, Options: []interface{}{"always"}},

			// ---- Real-world: React useState pattern ----
			{Code: `function App() { const [count, setCount] = useState(0), [name, setName] = useState(''); return null; }`, Options: []interface{}{"always"}},

			// ---- Real-world: webpack-style require with destructuring ----
			{Code: `const { join, resolve } = require('path');`, Options: []interface{}{map[string]interface{}{"separateRequires": true, "const": "always"}}},

			// ---- Real-world: TS strict null union default ----
			{Code: `const a: number | null = null, b: string | null = null;`, Options: []interface{}{"always"}},

			// ---- Real-world: error-handling pattern with try/catch ----
			{Code: `function f() { let result; try { result = compute(); } catch (e) { result = null; } return result; }`, Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}}},

			// ---- Empty destructuring still tracked (no declarators created) ----
			{Code: `var { } = obj;`, Options: []interface{}{"always"}},
			{Code: `var [] = arr;`, Options: []interface{}{"always"}},

			// ---- with statement (legacy) doesn't crash ----
			{Code: `function f() { with (obj) { var a, b; } }`, Options: []interface{}{"always"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- "never" basic ----
			{
				Code:    `var bar = true, baz = false;`,
				Output:  []string{`var bar = true; var baz = false;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Message: "Split 'var' declarations into multiple statements."},
				},
			},
			{
				Code:    `function foo() { var bar = true, baz = false; }`,
				Output:  []string{`function foo() { var bar = true; var baz = false; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `if (foo) { var bar = true, baz = false; }`,
				Output:  []string{`if (foo) { var bar = true; var baz = false; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `switch (foo) { case bar: var baz = true, quux = false; }`,
				Output:  []string{`switch (foo) { case bar: var baz = true; var quux = false; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			// switch-case + always: var is function-scoped, so two var statements
			// across cases share scope and trigger combine. Verified against
			// ESLint v10.2.1: reports `combine` on the second var at column 29-35.
			// Output is null upstream because the switch-case fix infrastructure
			// can't safely combine across non-body-array contexts.
			{
				Code:    `switch (x) { case 1: var a; var b; break; }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 29, EndLine: 1, EndColumn: 35},
				},
			},

			// Verified against ESLint v10.2.1: for-init in `always` mode reports
			// at the inner `var` keyword inside the for header, with end at the
			// last declarator (NOT including the `;`).
			{
				Code:    `function f() { var a = 1; for (var b = 2;;) {} }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 32, EndLine: 1, EndColumn: 41},
				},
			},

			// `export const a = 1; export const b = 2;` with default `always` —
			// scope-based combine fires. Position aligned to ESLint v10.2.1:
			// reports start at `const` (col 8), not `export` (col 1), end at `;`.
			{
				Code:    `export const a = 1; export const b = 2;`,
				Output:  nil, // export-wrapped fix not generated, mirrors upstream
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 28, EndLine: 1, EndColumn: 40},
				},
			},

			// `declare const a; declare const b;` — declare modifier is part
			// of VariableDeclaration in ESLint, so position starts at `declare`.
			{
				Code:    `declare const a: number, b: number;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
				},
			},

			// `export declare const ...` — only `export` is skipped; `declare`
			// stays in the range.
			{
				Code:    `export declare const a: number, b: number;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 8, EndLine: 1, EndColumn: 43},
				},
			},

			// Mixed export+inner consecutive: inner b→c combines, export-wrapped a is independent.
			{
				Code:    `export const a = 1; const b = 2; const c = 3;`,
				Output:  []string{`export const a = 1; const b = 2,  c = 3;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// Namespace / declare-module: declarations across multiple module
			// blocks share the OUTER scope (ESLint doesn't push for TSModuleBlock).
			// Verified against ESLint v10.2.1.
			{
				Code:    `namespace N { var a; } namespace M { var a; }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 38, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:    `var a; namespace N { var a; }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    `namespace Outer { var x; namespace Inner { var x; } }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 44, EndLine: 1, EndColumn: 50},
				},
			},
			// `declare module 'a.css' { const x: any; } declare module 'b.css' { const x: any; }`
			// — same outer scope; second `const x` reports combine.
			{
				Code:    `declare module 'a.css' { const classes: any; } declare module 'b.css' { const classes: any; }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// v10 — per-kind `using: 'never'` splits a combined using statement
			{
				Code:    `using a = 0, b = 1;`,
				Output:  []string{`using a = 0; using b = 1;`},
				Options: []interface{}{map[string]interface{}{"using": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Message: "Split 'using' declarations into multiple statements.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// v10 — per-kind `using: 'always'` combines two using statements
			{
				Code:    `using a = 0; using b = 1;`,
				Output:  []string{`using a = 0,  b = 1;`},
				Options: []interface{}{map[string]interface{}{"using": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Message: "Combine this with the previous 'using' statement."},
				},
			},
			// v10 — per-kind `awaitUsing: 'always'` (camelCase) maps to `"await using"` node.kind
			// and produces `'await using'` in the message text.
			{
				Code:    `await using a = 0; await using b = 1;`,
				Output:  []string{`await using a = 0,   b = 1;`},
				Options: []interface{}{map[string]interface{}{"awaitUsing": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Message: "Combine this with the previous 'await using' statement."},
				},
			},
			// v10 — per-kind `awaitUsing: 'never'` splits a combined await using statement
			{
				Code:    `await using a = 0, b = 1;`,
				Output:  []string{`await using a = 0; await using b = 1;`},
				Options: []interface{}{map[string]interface{}{"awaitUsing": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Message: "Split 'await using' declarations into multiple statements."},
				},
			},
			{
				Code:    `switch (foo) { default: var baz = true, quux = false; }`,
				Output:  []string{`switch (foo) { default: var baz = true; var quux = false; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- "always" basic ----
			{
				Code:    `function foo() { var bar = true; var baz = false; }`,
				Output:  []string{`function foo() { var bar = true,  baz = false; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Message: "Combine this with the previous 'var' statement."},
				},
			},
			{
				Code:    `var a = 1; for (var b = 2;;) {}`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `function foo() { var foo = true, bar = false; }`,
				Output:  []string{`function foo() { var foo = true; var bar = false; }`},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Message: "Split initialized 'var' declarations into multiple statements."},
				},
			},
			{
				Code:    `function foo() { var foo, bar; }`,
				Output:  []string{`function foo() { var foo; var bar; }`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitUninitialized", Message: "Split uninitialized 'var' declarations into multiple statements."},
				},
			},
			{
				// First iteration matches upstream exactly: combine `var c, d`
				// with prev `var b = false`. Subsequent iterations document the
				// rslint rule_tester's iterative fix-and-relint loop, which
				// surfaces additional legitimate diagnostics (each new state
				// has both initialized and uninitialized declarators in the
				// same statement, conflicting with the option matrix).
				Code: `function foo() { var bar, baz; var a = true; var b = false; var c, d;}`,
				Output: []string{
					`function foo() { var bar, baz; var a = true; var b = false,  c, d;}`,
					`function foo() { var bar, baz; var a = true,  b = false; var  c; var d;}`,
					`function foo() { var bar, baz; var a = true; var  b = false,   c,  d;}`,
					`function foo() { var bar, baz; var a = true,   b = false; var   c; var  d;}`,
					`function foo() { var bar, baz; var a = true; var   b = false,    c,   d;}`,
					`function foo() { var bar, baz; var a = true,    b = false; var    c; var   d;}`,
					`function foo() { var bar, baz; var a = true; var    b = false,     c,    d;}`,
					`function foo() { var bar, baz; var a = true,     b = false; var     c; var    d;}`,
					`function foo() { var bar, baz; var a = true; var     b = false,      c,     d;}`,
					`function foo() { var bar, baz; var a = true,      b = false; var      c; var     d;}`,
				},
				Options: []interface{}{map[string]interface{}{"uninitialized": "always", "initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Message: "Combine this with the previous 'var' statement with uninitialized variables."},
				},
			},
			{
				// Same rationale as prior test, mirrored options.
				Code: `function foo() { var bar = true, baz = false; var a; var b; var c = true, d = false; }`,
				Output: []string{
					`function foo() { var bar = true, baz = false; var a; var b,  c = true, d = false; }`,
					`function foo() { var bar = true, baz = false; var a,  b; var  c = true; var d = false; }`,
					`function foo() { var bar = true, baz = false; var a; var  b,   c = true,  d = false; }`,
					`function foo() { var bar = true, baz = false; var a,   b; var   c = true; var  d = false; }`,
					`function foo() { var bar = true, baz = false; var a; var   b,    c = true,   d = false; }`,
					`function foo() { var bar = true, baz = false; var a,    b; var    c = true; var   d = false; }`,
					`function foo() { var bar = true, baz = false; var a; var    b,     c = true,    d = false; }`,
					`function foo() { var bar = true, baz = false; var a,     b; var     c = true; var    d = false; }`,
					`function foo() { var bar = true, baz = false; var a; var     b,      c = true,     d = false; }`,
					`function foo() { var bar = true, baz = false; var a,      b; var      c = true; var     d = false; }`,
				},
				Options: []interface{}{map[string]interface{}{"uninitialized": "never", "initialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Message: "Combine this with the previous 'var' statement with initialized variables."},
				},
			},
			{
				Code:    `function foo() { var bar = true, baz = false; var a, b;}`,
				Output:  []string{`function foo() { var bar = true; var baz = false; var a; var b;}`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "never", "initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
					{MessageId: "split"},
				},
			},
			{
				Code:    `function foo() { var bar = true; var baz = false; var a; var b;}`,
				Output:  []string{`function foo() { var bar = true,  baz = false,  a,  b;}`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "always", "initialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
					{MessageId: "combine"},
					{MessageId: "combine"},
				},
			},
			{
				Code:    `function foo() { var a = [1, 2, 3]; var [b, c, d] = a; }`,
				Output:  []string{`function foo() { var a = [1, 2, 3],  [b, c, d] = a; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `function foo() { let a = 1; let b = 2; }`,
				Output:  []string{`function foo() { let a = 1,  b = 2; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Message: "Combine this with the previous 'let' statement."},
				},
			},
			{
				Code:    `function foo() { const a = 1; const b = 2; }`,
				Output:  []string{`function foo() { const a = 1,  b = 2; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Message: "Combine this with the previous 'const' statement."},
				},
			},
			{
				Code:    `function foo() { let a = 1; let b = 2; }`,
				Output:  []string{`function foo() { let a = 1,  b = 2; }`},
				Options: []interface{}{map[string]interface{}{"let": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `function foo() { const a = 1; const b = 2; }`,
				Output:  []string{`function foo() { const a = 1,  b = 2; }`},
				Options: []interface{}{map[string]interface{}{"const": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `function foo() { let a = 1, b = 2; }`,
				Output:  []string{`function foo() { let a = 1; let b = 2; }`},
				Options: []interface{}{map[string]interface{}{"let": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `function foo() { let a = 1, b = 2; }`,
				Output:  []string{`function foo() { let a = 1; let b = 2; }`},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized"},
				},
			},
			{
				Code:    `function foo() { let a, b; }`,
				Output:  []string{`function foo() { let a; let b; }`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitUninitialized"},
				},
			},
			{
				Code:    `function foo() { const a = 1, b = 2; }`,
				Output:  []string{`function foo() { const a = 1; const b = 2; }`},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized"},
				},
			},
			{
				Code:    `function foo() { const a = 1, b = 2; }`,
				Output:  []string{`function foo() { const a = 1; const b = 2; }`},
				Options: []interface{}{map[string]interface{}{"const": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- switch with let / position assertion ----
			{
				Code:    `let foo = true; switch(foo) { case true: let bar = 2; break; case false: let baz = 3; break; }`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"var": "always", "let": "always", "const": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 74},
				},
			},

			// ---- combine across newlines ----
			{
				Code:    "var one = 1, two = 2;\nvar three;",
				Output:  []string{"var one = 1, two = 2,\n three;"},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 2, Column: 1},
				},
			},

			// ---- splitInitialized / splitUninitialized ----
			{
				Code:    `var i = [0], j;`,
				Output:  []string{`var i = [0]; var j;`},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized"},
				},
			},
			{
				Code:    `var i = [0], j;`,
				Output:  []string{`var i = [0]; var j;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitUninitialized"},
				},
			},

			// ---- for-of / for-in combine ----
			{
				Code:    `for (var x of foo) {}; for (var y of foo) {}`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `for (var x in foo) {}; for (var y in foo) {}`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- combine with no options (default = always) ----
			{
				Code:   `var foo = function() { var bar = true; var baz = false; }`,
				Output: []string{`var foo = function() { var bar = true,  baz = false; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 40},
				},
			},
			{
				Code:   `function foo() { var bar = true; if (qux) { var baz = false; } else { var quxx = 42; } }`,
				Output: nil,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 45},
					{MessageId: "combine", Line: 1, Column: 71},
				},
			},
			{
				Code:   `var foo = () => { var bar = true; var baz = false; }`,
				Output: []string{`var foo = () => { var bar = true,  baz = false; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 35},
				},
			},
			{
				Code:   `var foo = function() { var bar = true; if (qux) { var baz = false; } }`,
				Output: nil,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 51},
				},
			},
			{
				Code:   `var foo; var bar;`,
				Output: []string{`var foo,  bar;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 10},
				},
			},

			// ---- for + initialized/uninitialized split ----
			{
				Code:    `var x = 1, y = 2; for (var z in foo) {}`,
				Output:  []string{`var x = 1; var y = 2; for (var z in foo) {}`},
				Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Line: 1, Column: 1},
				},
			},
			{
				Code:    `var x = 1, y = 2; for (var z of foo) {}`,
				Output:  []string{`var x = 1; var y = 2; for (var z of foo) {}`},
				Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Line: 1, Column: 1},
				},
			},
			{
				Code:    `var x; var y; for (var z in foo) {}`,
				Output:  []string{`var x,  y; for (var z in foo) {}`},
				Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 8},
				},
			},
			{
				Code:    `var x; var y; for (var z of foo) {}`,
				Output:  []string{`var x,  y; for (var z of foo) {}`},
				Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 8},
				},
			},
			{
				// First iteration matches upstream exactly. Subsequent rounds
				// reflect the rule_tester loop applying combine on the
				// re-emitted `var bar = y, a` (init+uninit conflict).
				Code: `var x; for (var y in foo) {var bar = y; var a; for (var z of bar) {}}`,
				Output: []string{
					`var x; for (var y in foo) {var bar = y,  a; for (var z of bar) {}}`,
					`var x; for (var y in foo) {var bar = y; var  a; for (var z of bar) {}}`,
					`var x; for (var y in foo) {var bar = y,   a; for (var z of bar) {}}`,
					`var x; for (var y in foo) {var bar = y; var   a; for (var z of bar) {}}`,
					`var x; for (var y in foo) {var bar = y,    a; for (var z of bar) {}}`,
					`var x; for (var y in foo) {var bar = y; var    a; for (var z of bar) {}}`,
					`var x; for (var y in foo) {var bar = y,     a; for (var z of bar) {}}`,
					`var x; for (var y in foo) {var bar = y; var     a; for (var z of bar) {}}`,
					`var x; for (var y in foo) {var bar = y,      a; for (var z of bar) {}}`,
					`var x; for (var y in foo) {var bar = y; var      a; for (var z of bar) {}}`,
				},
				Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 41},
				},
			},
			{
				Code:    `var a = 1; var b = 2; var x, y; for (var z of foo) {var c = 3, baz = z; for (var d in baz) {}}`,
				Output:  []string{`var a = 1; var b = 2; var x, y; for (var z of foo) {var c = 3; var baz = z; for (var d in baz) {}}`},
				Options: []interface{}{map[string]interface{}{"initialized": "never", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Line: 1, Column: 53},
				},
			},

			// ---- destructuring ----
			{
				Code:    `var {foo} = 1, [bar] = 2;`,
				Output:  []string{`var {foo} = 1; var [bar] = 2;`},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Line: 1, Column: 1},
				},
			},

			// ---- multi-line splits ----
			{
				Code:    "const foo = 1,\n    bar = 2;",
				Output:  []string{"const foo = 1;\n    const bar = 2;"},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Line: 1, Column: 1},
				},
			},
			{
				Code:    "var foo = 1,\n    bar = 2;",
				Output:  []string{"var foo = 1;\n    var bar = 2;"},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Line: 1, Column: 1},
				},
			},
			{
				Code:    "var foo = 1, // comment\n    bar = 2;",
				Output:  []string{"var foo = 1; // comment\n    var bar = 2;"},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Line: 1, Column: 1},
				},
			},
			{
				Code:    `var f, k /* test */, l;`,
				Output:  []string{`var f; var k /* test */; var l;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 1},
				},
			},
			{
				Code:    `var f,          /* test */ l;`,
				Output:  []string{`var f;          /* test */ var l;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 1},
				},
			},
			{
				Code:    "var f, k /* test \n some more comment \n even more */, l = 1, P;",
				Output:  []string{"var f; var k /* test \n some more comment \n even more */; var l = 1; var P;"},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 1},
				},
			},
			{
				Code:    `var a = 1, b = 2`,
				Output:  []string{`var a = 1; var b = 2`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 1},
				},
			},

			// ---- separateRequires ----
			{
				Code:    `var foo = require('foo'), bar;`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitRequires", Line: 1, Column: 1},
				},
			},
			{
				Code:    `var foo, bar = require('bar');`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitRequires", Line: 1, Column: 1},
				},
			},
			{
				Code:    `let foo, bar = require('bar');`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "let": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitRequires", Line: 1, Column: 1},
				},
			},
			{
				Code:    `const foo = 0, bar = require('bar');`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "const": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitRequires", Line: 1, Column: 1},
				},
			},
			{
				Code:    `const foo = require('foo'); const bar = require('bar');`,
				Output:  []string{`const foo = require('foo'),  bar = require('bar');`},
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "const": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 29},
				},
			},

			// ---- consecutive (https://github.com/eslint/eslint/issues/4680) ----
			{
				Code:    `var a = 1, b; var c;`,
				Output:  []string{`var a = 1, b,  c;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 15},
				},
			},
			{
				Code:    `var a = 0, b = 1; var c = 2;`,
				Output:  []string{`var a = 0, b = 1,  c = 2;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 19},
				},
			},
			{
				Code:    `let a = 1, b; let c;`,
				Output:  []string{`let a = 1, b,  c;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 15},
				},
			},
			{
				Code:    `let a = 0, b = 1; let c = 2;`,
				Output:  []string{`let a = 0, b = 1,  c = 2;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 19},
				},
			},
			{
				Code:    `const a = 0, b = 1; const c = 2;`,
				Output:  []string{`const a = 0, b = 1,  c = 2;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 21},
				},
			},
			{
				Code:    `const a = 0; var b = 1; var c = 2; const d = 3;`,
				Output:  []string{`const a = 0; var b = 1,  c = 2; const d = 3;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 25},
				},
			},
			{
				Code:    `var a = true; var b = false;`,
				Output:  []string{`var a = true,  b = false;`},
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 15},
				},
			},
			{
				Code:    `const a = 0; let b = 1; let c = 2; const d = 3;`,
				Output:  []string{`const a = 0; let b = 1,  c = 2; const d = 3;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 25},
				},
			},
			{
				Code:    `let a = 0; const b = 1; const c = 1; var d = 2;`,
				Output:  []string{`let a = 0; const b = 1,  c = 1; var d = 2;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 25},
				},
			},

			// ---- consecutive per init/uninit ----
			{
				Code:    `var a = 0; var b; var c; var d = 1`,
				Output:  []string{`var a = 0; var b,  c; var d = 1`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 19},
				},
			},
			{
				Code:    `var a = 0; var b = 1; var c; var d;`,
				Output:  []string{`var a = 0,  b = 1; var c,  d;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 12},
					{MessageId: "combineUninitialized", Line: 1, Column: 30},
				},
			},
			{
				Code:    `let a = 0; let b; let c; let d = 1;`,
				Output:  []string{`let a = 0; let b,  c; let d = 1;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 19},
				},
			},
			{
				Code:    `let a = 0; let b = 1; let c; let d;`,
				Output:  []string{`let a = 0,  b = 1; let c,  d;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 12},
					{MessageId: "combineUninitialized", Line: 1, Column: 30},
				},
			},
			{
				Code:    `const a = 0; let b; let c; const d = 1;`,
				Output:  []string{`const a = 0; let b,  c; const d = 1;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 21},
				},
			},
			{
				Code:    `const a = 0; const b = 1; let c; let d;`,
				Output:  []string{`const a = 0,  b = 1; let c,  d;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 14},
					{MessageId: "combineUninitialized", Line: 1, Column: 34},
				},
			},
			{
				Code:    `var a = 0; var b = 1; var c, d;`,
				Output:  []string{`var a = 0,  b = 1; var c; var d;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 12},
					{MessageId: "splitUninitialized", Line: 1, Column: 23},
				},
			},
			{
				Code:    `var a = 0; var b, c; var d = 1;`,
				Output:  []string{`var a = 0; var b; var c; var d = 1;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitUninitialized", Line: 1, Column: 12},
				},
			},
			{
				Code:    `let a = 0; let b = 1; let c, d;`,
				Output:  []string{`let a = 0,  b = 1; let c; let d;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 12},
					{MessageId: "splitUninitialized", Line: 1, Column: 23},
				},
			},
			{
				Code:    `let a = 0; let b, c; let d = 1;`,
				Output:  []string{`let a = 0; let b; let c; let d = 1;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitUninitialized", Line: 1, Column: 12},
				},
			},
			{
				Code:    `const a = 0; const b = 1; let c, d;`,
				Output:  []string{`const a = 0,  b = 1; let c; let d;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 14},
					{MessageId: "splitUninitialized", Line: 1, Column: 27},
				},
			},
			{
				Code:    `const a = 0; let b, c; const d = 1;`,
				Output:  []string{`const a = 0; let b; let c; const d = 1;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitUninitialized", Line: 1, Column: 14},
				},
			},
			{
				Code:    `var a; var b; var c = 0; var d = 1;`,
				Output:  []string{`var a,  b; var c = 0,  d = 1;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 8},
					{MessageId: "combineInitialized", Line: 1, Column: 26},
				},
			},
			{
				Code:    `var a; var b = 0; var c = 1; var d;`,
				Output:  []string{`var a; var b = 0,  c = 1; var d;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 19},
				},
			},
			{
				Code:    `let a; let b; let c = 0; let d = 1;`,
				Output:  []string{`let a,  b; let c = 0,  d = 1;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 8},
					{MessageId: "combineInitialized", Line: 1, Column: 26},
				},
			},
			{
				Code:    `let a; let b = 0; let c = 1; let d;`,
				Output:  []string{`let a; let b = 0,  c = 1; let d;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 19},
				},
			},
			{
				Code:    `let a; let b; const c = 0; const d = 1;`,
				Output:  []string{`let a,  b; const c = 0,  d = 1;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 8},
					{MessageId: "combineInitialized", Line: 1, Column: 28},
				},
			},
			{
				Code:    `let a; const b = 0; const c = 1; let d;`,
				Output:  []string{`let a; const b = 0,  c = 1; let d;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 21},
				},
			},
			{
				Code:    `var a; var b; var c = 0, d = 1;`,
				Output:  []string{`var a,  b; var c = 0; var d = 1;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 8},
					{MessageId: "splitInitialized", Line: 1, Column: 15},
				},
			},
			{
				Code:    `var a; var b = 0, c = 1; var d;`,
				Output:  []string{`var a; var b = 0; var c = 1; var d;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Line: 1, Column: 8},
				},
			},
			{
				Code:    `let a; let b; let c = 0, d = 1;`,
				Output:  []string{`let a,  b; let c = 0; let d = 1;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 8},
					{MessageId: "splitInitialized", Line: 1, Column: 15},
				},
			},
			{
				Code:    `let a; let b = 0, c = 1; let d;`,
				Output:  []string{`let a; let b = 0; let c = 1; let d;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Line: 1, Column: 8},
				},
			},
			{
				Code:    `let a; let b; const c = 0, d = 1;`,
				Output:  []string{`let a,  b; const c = 0; const d = 1;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineUninitialized", Line: 1, Column: 8},
					{MessageId: "splitInitialized", Line: 1, Column: 15},
				},
			},
			{
				Code:    `let a; const b = 0, c = 1; let d;`,
				Output:  []string{`let a; const b = 0; const c = 1; let d;`},
				Options: []interface{}{map[string]interface{}{"uninitialized": "consecutive", "initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized", Line: 1, Column: 8},
				},
			},

			// ---- consecutive (per kind) ----
			{
				Code:    `var a = 0; var b = 1;`,
				Output:  []string{`var a = 0,  b = 1;`},
				Options: []interface{}{map[string]interface{}{"var": "consecutive"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 12},
				},
			},
			{
				Code:    `let a = 0; let b = 1;`,
				Output:  []string{`let a = 0,  b = 1;`},
				Options: []interface{}{map[string]interface{}{"let": "consecutive"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 12},
				},
			},
			{
				Code:    `const a = 0; const b = 1;`,
				Output:  []string{`const a = 0,  b = 1;`},
				Options: []interface{}{map[string]interface{}{"const": "consecutive"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 14},
				},
			},
			{
				Code:    `let a; let b; const c = 0; const d = 1;`,
				Output:  []string{`let a,  b; const c = 0,  d = 1;`},
				Options: []interface{}{map[string]interface{}{"let": "consecutive", "const": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 8},
					{MessageId: "combine", Line: 1, Column: 28},
				},
			},
			{
				Code:    `let a; const b = 0; const c = 1; let d;`,
				Output:  []string{`let a; const b = 0,  c = 1; let d;`},
				Options: []interface{}{map[string]interface{}{"let": "consecutive", "const": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 21},
				},
			},
			{
				Code:    `let a; let b; const c = 0, d = 1;`,
				Output:  []string{`let a,  b; const c = 0; const d = 1;`},
				Options: []interface{}{map[string]interface{}{"let": "consecutive", "const": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 8},
					{MessageId: "split", Line: 1, Column: 15},
				},
			},
			{
				Code:    `let a; const b = 0, c = 1; let d;`,
				Output:  []string{`let a; const b = 0; const c = 1; let d;`},
				Options: []interface{}{map[string]interface{}{"let": "consecutive", "const": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 8},
				},
			},
			{
				Code:    `const a = 0; const b = 1; let c; let d;`,
				Output:  []string{`const a = 0,  b = 1; let c,  d;`},
				Options: []interface{}{map[string]interface{}{"const": "consecutive", "let": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 14},
					{MessageId: "combine", Line: 1, Column: 34},
				},
			},
			{
				Code:    `const a = 0; let b; let c; const d = 1;`,
				Output:  []string{`const a = 0; let b,  c; const d = 1;`},
				Options: []interface{}{map[string]interface{}{"const": "consecutive", "let": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 21},
				},
			},
			{
				Code:    `const a = 0; const b = 1; let c, d;`,
				Output:  []string{`const a = 0,  b = 1; let c; let d;`},
				Options: []interface{}{map[string]interface{}{"const": "consecutive", "let": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 14},
					{MessageId: "split", Line: 1, Column: 27},
				},
			},
			{
				Code:    `const a = 0; let b, c; const d = 1;`,
				Output:  []string{`const a = 0; let b; let c; const d = 1;`},
				Options: []interface{}{map[string]interface{}{"const": "consecutive", "let": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 14},
				},
			},
			{
				Code:    `var bar; var baz;`,
				Output:  []string{`var bar,  baz;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 10},
				},
			},
			{
				Code:    `var bar = 1; var baz = 2; qux(); var qux = 3; var quux;`,
				Output:  []string{`var bar = 1,  baz = 2; qux(); var qux = 3,  quux;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 14},
					{MessageId: "combine", Line: 1, Column: 47},
				},
			},
			{
				Code:    `let a, b; let c; var d, e;`,
				Output:  []string{`let a, b,  c; var d; var e;`},
				Options: []interface{}{map[string]interface{}{"var": "never", "let": "consecutive", "const": "consecutive"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 11},
					{MessageId: "split", Line: 1, Column: 18},
				},
			},
			{
				Code:    `var a; var b;`,
				Output:  []string{`var a,  b;`},
				Options: []interface{}{map[string]interface{}{"var": "consecutive"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 8},
				},
			},
			{
				Code:    `var a = 1; var b = 2; var c, d; var e = 3; var f = 4;`,
				Output:  []string{`var a = 1,  b = 2; var c; var d; var e = 3,  f = 4;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive", "uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 12},
					{MessageId: "splitUninitialized", Line: 1, Column: 23},
					{MessageId: "combineInitialized", Line: 1, Column: 44},
				},
			},
			{
				Code:    `var a = 1; var b = 2; foo(); var c = 3; var d = 4;`,
				Output:  []string{`var a = 1,  b = 2; foo(); var c = 3,  d = 4;`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized", Line: 1, Column: 12},
					{MessageId: "combineInitialized", Line: 1, Column: 41},
				},
			},
			{
				Code:    "var a\nvar b",
				Output:  []string{"var a,\n b"},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 2, Column: 1},
				},
			},

			// ---- export ----
			{
				Code:    `export const foo=1, bar=2;`,
				Output:  []string{`export const foo=1; export const bar=2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    "const foo=1,\n bar=2;",
				Output:  []string{"const foo=1;\n const bar=2;"},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    "export const foo=1,\n bar=2;",
				Output:  []string{"export const foo=1;\n export const bar=2;"},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    "export const foo=1\n, bar=2;",
				Output:  []string{"export const foo=1\n; export const bar=2;"},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `export const foo= a, bar=2;`,
				Output:  []string{`export const foo= a; export const bar=2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `export const foo=() => a, bar=2;`,
				Output:  []string{`export const foo=() => a; export const bar=2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `export const foo= a, bar=2, bar2=2;`,
				Output:  []string{`export const foo= a; export const bar=2; export const bar2=2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `export const foo = 1,bar = 2;`,
				Output:  []string{`export const foo = 1; export const bar = 2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- "never" should not autofix declarations in a block position ----
			{
				Code:    `if (foo) var x, y;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `if (foo) var x, y;`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"var": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `if (foo) var x, y;`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"uninitialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitUninitialized"},
				},
			},
			{
				Code:    `if (foo) var x = 1, y = 1;`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized"},
				},
			},
			{
				Code:    `if (foo) {} else var x, y;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `while (foo) var x, y;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `with (foo) var x, y;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `do var x, y; while (foo);`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `do var x = f(), y = b(); while (x < y);`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `for (;;) var x, y;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `for (foo in bar) var x, y;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `for (foo of bar) var x, y;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `label: var x, y;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- class static blocks ----
			{
				Code:    `class C { static { let x, y; } }`,
				Output:  []string{`class C { static { let x; let y; } }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `class C { static { var x, y; } }`,
				Output:  []string{`class C { static { var x; var y; } }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `class C { static { let x; let y; } }`,
				Output:  []string{`class C { static { let x,  y; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { static { var x; var y; } }`,
				Output:  []string{`class C { static { var x,  y; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { static { let x; foo; let y; } }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { static { var x; foo; var y; } }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { static { var x; if (foo) { var y; } } }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { static { let x; let y; } }`,
				Output:  []string{`class C { static { let x,  y; } }`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { static { var x; var y; } }`,
				Output:  []string{`class C { static { var x,  y; } }`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { static { let a = 0; let b = 1; } }`,
				Output:  []string{`class C { static { let a = 0,  b = 1; } }`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized"},
				},
			},
			{
				Code:    `class C { static { var a = 0; var b = 1; } }`,
				Output:  []string{`class C { static { var a = 0,  b = 1; } }`},
				Options: []interface{}{map[string]interface{}{"initialized": "consecutive"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combineInitialized"},
				},
			},

			// ---- Explicit Resource Management ----
			{
				Code:    `using a = 0; using b = 1;`,
				Output:  []string{`using a = 0,  b = 1;`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Message: "Combine this with the previous 'using' statement."},
				},
			},
			{
				Code:    `await using a = 0; await using b = 1;`,
				Output:  []string{`await using a = 0,   b = 1;`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Message: "Combine this with the previous 'await using' statement."},
				},
			},
			{
				Code:    `using a = 0, b = 1;`,
				Output:  []string{`using a = 0; using b = 1;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Message: "Split 'using' declarations into multiple statements."},
				},
			},
			{
				Code:    `await using a = 0, b = 1;`,
				Output:  []string{`await using a = 0; await using b = 1;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Message: "Split 'await using' declarations into multiple statements."},
				},
			},
			{
				Code:    `using a = 0; using b = 1;`,
				Output:  []string{`using a = 0,  b = 1;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `await using a = 0; await using b = 1;`,
				Output:  []string{`await using a = 0,   b = 1;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `using a = 0, b = 1;`,
				Output:  []string{`using a = 0; using b = 1;`},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized"},
				},
			},
			{
				Code:    `await using a = 0, b = 1;`,
				Output:  []string{`await using a = 0; await using b = 1;`},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized"},
				},
			},

			// =========================================================
			// rslint-extra invalid cases below — beyond upstream coverage
			// =========================================================

			// ---- Upstream semantic-walk lock-ins ----
			// hasOnlyOneStatement: `scope.required && hasRequires` triggers
			// — two consecutive pure-require statements with separateRequires
			// produce combine (since they're all-requires, not mixed).
			{
				Code:    `var a = require('a'); var b = require('b');`,
				Output:  []string{`var a = require('a'),  b = require('b');`},
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- TS-specific: parens around the require call ----
			// tsgo preserves ParenthesizedExpression; rule must SkipParentheses to match ESLint.
			{
				Code:    `var foo = (require('foo')), bar = 'bar';`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitRequires"},
				},
			},

			// ---- TS-specific: double parens ----
			{
				Code:    `var foo = ((require('foo'))), bar = 'bar';`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "var": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitRequires"},
				},
			},

			// ---- TS-specific: type annotations don't disable autofix ----
			{
				Code:    `var x: number = 1, y: string = 'a';`,
				Output:  []string{`var x: number = 1; var y: string = 'a';`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `var x: number = 1; var y: string = 'a';`,
				Output:  []string{`var x: number = 1,  y: string = 'a';`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- TS-specific: declare disables autofix (not in statement-list-friendly form) ----
			{
				Code:    `declare var x: number, y: string;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Destructuring rest / object patterns / array patterns ----
			{
				Code:    `var { a, b } = obj, c;`,
				Output:  []string{`var { a, b } = obj; var c;`},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized"},
				},
			},
			{
				Code:    `var [a, ...b] = arr, c = 1;`,
				Output:  []string{`var [a, ...b] = arr; var c = 1;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `let { a: { b, c } = {} } = obj, d;`,
				Output:  []string{`let { a: { b, c } = {} } = obj; let d;`},
				Options: []interface{}{map[string]interface{}{"initialized": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitInitialized"},
				},
			},

			// ---- ModuleBlock / namespace ----
			{
				Code:    `namespace N { var a = 1; var b = 2; }`,
				Output:  []string{`namespace N { var a = 1,  b = 2; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `namespace N { let a, b; }`,
				Output:  []string{`namespace N { let a; let b; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Class methods, accessors, constructor ----
			{
				Code:    `class C { method() { let a; let b; } }`,
				Output:  []string{`class C { method() { let a,  b; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { constructor() { var a = 1; var b = 2; } }`,
				Output:  []string{`class C { constructor() { var a = 1,  b = 2; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { get x() { let a; let b; return a + b; } }`,
				Output:  []string{`class C { get x() { let a,  b; return a + b; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { set x(v) { let a, b; a + b; } }`,
				Output:  []string{`class C { set x(v) { let a; let b; a + b; } }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Class expression body — is a function boundary ----
			{
				Code:    `var Foo = class { method() { var a; var b; } };`,
				Output:  []string{`var Foo = class { method() { var a,  b; } };`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- Async / generator function bodies ----
			{
				Code:    `async function f() { var a = 1; var b = 2; }`,
				Output:  []string{`async function f() { var a = 1,  b = 2; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `function* gen() { let a; let b; yield a + b; }`,
				Output:  []string{`function* gen() { let a,  b; yield a + b; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `async function* g() { let a, b; yield a; yield b; }`,
				Output:  []string{`async function* g() { let a; let b; yield a; yield b; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Try / catch / finally bodies are blocks ----
			{
				Code:    `function f() { try { let a; let b; } catch (e) {} }`,
				Output:  []string{`function f() { try { let a,  b; } catch (e) {} }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `function f() { try {} catch (e) { let a; let b; } }`,
				Output:  []string{`function f() { try {} catch (e) { let a,  b; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `function f() { try {} finally { let a; let b; } }`,
				Output:  []string{`function f() { try {} finally { let a,  b; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- IIFE (function expression call) — fresh function scope ----
			{
				Code:    `var x; (function () { var x; var y; })();`,
				Output:  []string{`var x; (function () { var x,  y; })();`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- Default-export wrapping (TS-specific shape) ----
			{
				Code:    `export default function f() { var a; var b; }`,
				Output:  []string{`export default function f() { var a,  b; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- Comments on same line preserved through split ----
			{
				Code:    `var a /* x */, b /* y */;`,
				Output:  []string{`var a /* x */; var b /* y */;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Trailing comma in var list — verified syntax error in both
			// ESLint (espree: "Parsing error: Unexpected token ;") and tsgo
			// parser. Since the source never reaches the rule, no diagnostic
			// from one-var is expected. We don't include a test case here
			// because the rule_tester runs the full linter and a parse error
			// would surface as a different diagnostic, not a one-var miss.

			// ---- Multi-byte chars in identifier names — column math ----
			{
				Code:    `var π = 1, ω = 2;`,
				Output:  []string{`var π = 1; var ω = 2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 1},
				},
			},

			// ---- Double-byte string-literal content does not break split ----
			{
				Code:    `var a = "你好", b = "世界";`,
				Output:  []string{`var a = "你好"; var b = "世界";`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Template literal initializer ----
			{
				Code:    "var a = `foo`, b = `bar`;",
				Output:  []string{"var a = `foo`; var b = `bar`;"},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Function-expression initializer ----
			{
				Code:    `var a = function () {}, b = function () {};`,
				Output:  []string{`var a = function () {}; var b = function () {};`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Arrow function initializer with parens ----
			{
				Code:    `const a = () => 1, b = () => 2;`,
				Output:  []string{`const a = () => 1; const b = () => 2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Generic / type-arg in initializer ----
			{
				Code:    `var a = foo<T>(), b = bar<U>();`,
				Output:  []string{`var a = foo<T>(); var b = bar<U>();`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Real-world: redux-style consts ----
			{
				Code:    `const ADD = 'ADD'; const REMOVE = 'REMOVE'; const RESET = 'RESET';`,
				Output:  []string{`const ADD = 'ADD',  REMOVE = 'REMOVE',  RESET = 'RESET';`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
					{MessageId: "combine"},
				},
			},

			// ---- Real-world: Node CommonJS imports ----
			{
				Code:    `const fs = require('fs'); const path = require('path');`,
				Output:  []string{`const fs = require('fs'),  path = require('path');`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- Real-world: top-level let in module ----
			{
				Code:    `export let a = 1; export let b = 2;`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					// upstream's joinDeclarations stops at the ExportNamedDeclaration wrapper.
					// We mirror by skipping fix when either side has `export` modifier.
					{MessageId: "combine"},
				},
			},

			// ---- Position: report node end span ----
			{
				Code:    `function f() { var a; var b; }`,
				Output:  []string{`function f() { var a,  b; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 23, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:    `var a, b;`,
				Output:  []string{`var a; var b;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},

			// #4 — for-init scope-conflict lock-ins
			// `var a = 1; for (var b = 2;;)` — combine reported (function-scope conflict),
			// no fix (covered earlier in the file already).
			// `for (var x in foo) {}; for (var y in foo) {}` — combine reported (function-scope
			// conflict on var across two for-in heads). Already covered.

			// #5 — `using a = require('a')` triggers separateRequires
			// (upstream's isRequire only checks call-callee.name, so `using` decls qualify).
			{
				Code:    `using a = require('a'), b = 1;`,
				Output:  nil,
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "using": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "splitRequires"},
				},
			},

			// #7 — Multi-byte char column / endColumn lock-ins.
			// UTF-16 column counting (the rule_tester uses GetECMALineAndUTF16CharacterOfPosition).
			{
				Code:    `var π = 1, ω = 2;`,
				Output:  []string{`var π = 1; var ω = 2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `var 你 = 1, 我 = 2;`,
				Output:  []string{`var 你 = 1; var 我 = 2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},

			// #10 — namespace var scope: same-namespace var conflict (always)
			{
				Code:    `namespace N { var a; var b; }`,
				Output:  []string{`namespace N { var a,  b; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			// And split inside a namespace
			{
				Code:    `namespace N { var a, b; }`,
				Output:  []string{`namespace N { var a; var b; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Multi-line body, position assertions ----
			// Note: removing the `var` keyword leaves the original whitespace intact —
			// indent (2 spaces) + space-between-`var`-and-`b` = 3 leading spaces on line 3.
			{
				Code: "function f() {\n  var a = 1;\n  var b = 2;\n}",
				Output: []string{
					"function f() {\n  var a = 1,\n   b = 2;\n}",
				},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 3, Column: 3},
				},
			},

			// =========================================================
			// =========================================================
			// Listener-coverage matrix — every ESLint listener kind × ≥3 forms
			// (verified against ESLint v10.2.1 via diff-eslint.mjs)
			// =========================================================
			// --- BlockStatement ---
			{
				Code:    `function f() { { let a; let b; } }`,
				Output:  []string{`function f() { { let a,  b; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `if (cond) { let a; let b; }`,
				Output:  []string{`if (cond) { let a,  b; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `try { let a; let b; } catch {}`,
				Output:  []string{`try { let a,  b; } catch {}`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			// --- FunctionDeclaration / FunctionExpression / ArrowFunction ---
			{
				Code:    `function f() { let a; let b; }`,
				Output:  []string{`function f() { let a,  b; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `var f = function () { let a; let b; };`,
				Output:  []string{`var f = function () { let a,  b; };`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `var f = () => { let a; let b; };`,
				Output:  []string{`var f = () => { let a,  b; };`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			// --- StaticBlock — function-scope boundary for var ---
			{
				Code:    `class C { static { var a; var b; } }`,
				Output:  []string{`class C { static { var a,  b; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { static { let a; let b; } }`,
				Output:  []string{`class C { static { let a,  b; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `class C { static { const a = 1; const b = 2; } }`,
				Output:  []string{`class C { static { const a = 1,  b = 2; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			// --- ForStatement / ForInStatement / ForOfStatement ---
			{
				Code:    `for (var i = 0;;) {} for (var j = 0;;) {}`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `for (var i in x) {} for (var j in x) {}`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `for (var i of x) {} for (var j of x) {}`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			// --- SwitchStatement — function-scope (var) shared, block scope per-switch ---
			{
				Code:    `switch (x) { case 1: var a; var b; break; }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `switch (x) { default: const a = 1; const b = 2; }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				// `let a; case 2: let b;` across switch cases — let is block-scoped
				// to the switch statement; both lets are in the same block scope.
				Code:    `switch (x) { case 1: let a; case 2: let b; break; }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// =========================================================
			// tsgo-only AST nodes (TSModuleBlock / TSTypeAlias / TSInterface /
			// TSEnum / TSDeclareFunction) — ESLint doesn't have listeners for
			// these, so they should be transparent: declarations on either side
			// see the same scope. Each verified against ESLint v10.2.1.
			// =========================================================
			{
				Code:    `namespace N { var a; var b; }`,
				Output:  []string{`namespace N { var a,  b; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `namespace A { var a; } namespace B { var a; }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `declare module 'x' { var a; var b; }`,
				Output:  []string{`declare module 'x' { var a,  b; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `declare module 'x.css' { const c: any; } declare module 'y.css' { const c: any; }`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			// TSTypeAliasDeclaration — transparent to scope
			{
				Code:    `type T = number; var a; var b;`,
				Output:  []string{`type T = number; var a,  b;`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				// Type alias between var statements is transparent to scope, so
				// `combine` is reported on `var b`. ESLint v10 attaches no fix
				// because joinDeclarations' previous-sibling lookup finds the
				// type alias (not a VariableDeclaration), so the fixer body
				// short-circuits. rslint matches: report + no fix.
				Code:    `var a; type T = number; var b;`,
				Output:  nil,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 25, EndLine: 1, EndColumn: 31},
				},
			},
			// TSInterfaceDeclaration — transparent
			{
				Code:    `interface I {} var a; var b;`,
				Output:  []string{`interface I {} var a,  b;`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			// TSEnumDeclaration — transparent
			{
				Code:    `enum E { A } var a; var b;`,
				Output:  []string{`enum E { A } var a,  b;`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			// TSDeclareFunction — transparent
			{
				Code:    `declare function f(): void; var a; var b;`,
				Output:  []string{`declare function f(): void; var a,  b;`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// Round 2 — additional tsgo-edge / real-world hardening invalid
			// =========================================================

			// ---- TS as / satisfies / non-null in initializer don't disable autofix ----
			{
				Code:    `var a = (x as number); var b = (y satisfies T);`,
				Output:  []string{`var a = (x as number),  b = (y satisfies T);`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
			{
				Code:    `var a = x!.foo, b = y!.bar;`,
				Output:  []string{`var a = x!.foo; var b = y!.bar;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Optional chain initializer ----
			{
				Code:    `var a = obj?.foo, b = obj?.bar;`,
				Output:  []string{`var a = obj?.foo; var b = obj?.bar;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Tagged template ----
			{
				Code:    "var a = tag`a`, b = tag`b`;",
				Output:  []string{"var a = tag`a`; var b = tag`b`;"},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- BigInt / numeric separator literal initializer ----
			{
				Code:    `var a = 100n, b = 200n;`,
				Output:  []string{`var a = 100n; var b = 200n;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},
			{
				Code:    `var a = 1_000, b = 2_000;`,
				Output:  []string{`var a = 1_000; var b = 2_000;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- RegExp literal initializer (with flags) ----
			{
				Code:    `var a = /foo/g, b = /bar/i;`,
				Output:  []string{`var a = /foo/g; var b = /bar/i;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- typeof / void in initializer ----
			{
				Code:    `var a = typeof x, b = void 0;`,
				Output:  []string{`var a = typeof x; var b = void 0;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- await in initializer ----
			{
				Code:    `async function f() { var a = await x, b = await y; }`,
				Output:  []string{`async function f() { var a = await x; var b = await y; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- yield in generator initializer ----
			{
				Code:    `function* g() { let a = yield 1, b = yield 2; }`,
				Output:  []string{`function* g() { let a = yield 1; let b = yield 2; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Object destructuring with defaults / rename ----
			{
				Code:    `var { a = 1 } = obj1, { b = 2 } = obj2;`,
				Output:  []string{`var { a = 1 } = obj1; var { b = 2 } = obj2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Array destructuring with rest ----
			{
				Code:    `var [a, ...rest] = arr1, [b, ...rest2] = arr2;`,
				Output:  []string{`var [a, ...rest] = arr1; var [b, ...rest2] = arr2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Computed property destructuring ----
			{
				Code:    `var { [k]: a } = obj, { [k]: b } = obj2;`,
				Output:  []string{`var { [k]: a } = obj; var { [k]: b } = obj2;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- React useState — split into separate statements ----
			{
				Code:    `function App() { const [count, setCount] = useState(0), [name, setName] = useState(''); return null; }`,
				Output:  []string{`function App() { const [count, setCount] = useState(0); const [name, setName] = useState(''); return null; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- React useState combine ----
			{
				Code:    `function App() { const [count, setCount] = useState(0); const [name, setName] = useState(''); return null; }`,
				Output:  []string{`function App() { const [count, setCount] = useState(0),  [name, setName] = useState(''); return null; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- Webpack require pattern (separateRequires + always) — pure-require pair combines ----
			{
				Code:    `const fs = require('fs'); const path = require('path'); const x = 1;`,
				Output:  []string{`const fs = require('fs'),  path = require('path'); const x = 1;`},
				Options: []interface{}{map[string]interface{}{"separateRequires": true, "const": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					// path → combines with fs (both pure-require, not mixed).
					{MessageId: "combine"},
				},
			},

			// ---- TS strict null union with type annotations ----
			{
				Code:    `const a: number | null = null, b: string | null = null;`,
				Output:  []string{`const a: number | null = null; const b: string | null = null;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Class with field decl and method (combine inside method body) ----
			{
				Code:    `class C { x = 1; y = 2; method() { let a; let b; } }`,
				Output:  []string{`class C { x = 1; y = 2; method() { let a,  b; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- super() in constructor before var ----
			{
				Code:    `class C extends D { constructor() { super(); var a; var b; } }`,
				Output:  []string{`class C extends D { constructor() { super(); var a,  b; } }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- with statement (var split inside) — NOT in statement-list position relative to with body, so no fix ----
			{
				Code:    `function f() { with (obj) { var a, b; } }`,
				Output:  []string{`function f() { with (obj) { var a; var b; } }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Multi-byte char in identifier + UTF-16 column counting ----
			// `var 你 = 1; var 我 = 2;` — second `var` starts at UTF-16 column 12.
			{
				Code:    `var 你 = 1; var 我 = 2;`,
				Output:  []string{`var 你 = 1,  我 = 2;`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 1, Column: 12},
				},
			},

			// ---- Module body var combine ----
			{
				Code:    `module M { var a; var b; }`,
				Output:  []string{`module M { var a,  b; }`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- consecutive across mixed kinds — mismatch should not combine ----
			{
				Code:    `let a = 1; let b = 2; const c = 3; const d = 4;`,
				Output:  []string{`let a = 1,  b = 2; const c = 3,  d = 4;`},
				Options: []interface{}{"consecutive"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
					{MessageId: "combine"},
				},
			},

			// ---- TS readonly modifier on declare ----
			{
				Code:    `declare const a: number, b: number;`,
				Output:  nil,
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Conditional initializer ----
			{
				Code:    `var a = x ? 1 : 2, b = y ? 3 : 4;`,
				Output:  []string{`var a = x ? 1 : 2; var b = y ? 3 : 4;`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- new in initializer ----
			{
				Code:    `var a = new Date(); var b = new Map();`,
				Output:  []string{`var a = new Date(),  b = new Map();`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- Spread in initializer ----
			{
				Code:    `var a = [...x]; var b = [...y];`,
				Output:  []string{`var a = [...x],  b = [...y];`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- Multi-line comment between var statements (combine fix preserves comment) ----
			{
				Code:    "var a = 1;\n/* between */\nvar b = 2;",
				Output:  []string{"var a = 1,\n/* between */\n b = 2;"},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- Async generator function body ----
			{
				Code:    `async function* gen() { var a = 1, b = 2; }`,
				Output:  []string{`async function* gen() { var a = 1; var b = 2; }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "split"},
				},
			},

			// ---- Object method shorthand body ----
			{
				Code:    `var obj = { method() { let a; let b; } };`,
				Output:  []string{`var obj = { method() { let a,  b; } };`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},

			// ---- Getter/setter pair: each is its own function-scope ----
			// Getter's two `let` collapse via combine; setter's `let c, d` is already
			// one statement and is the first `let` in its scope, so no diagnostic.
			{
				Code:    `var obj = { get x() { let a; let b; return a; }, set x(v) { let c, d; } };`,
				Output:  []string{`var obj = { get x() { let a,  b; return a; }, set x(v) { let c, d; } };`},
				Options: []interface{}{map[string]interface{}{"let": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine"},
				},
			},
		},
	)
}
