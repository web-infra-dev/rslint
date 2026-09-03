package no_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoVarRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoVarRule,
		[]rule_tester.ValidTestCase{
			{Code: `const JOE = 'schmoe';`},
			{Code: `let moo = 'car';`},
			{Code: `const JOE = 'schmoe'; let moo = 'car';`},
			{Code: `for (let i = 0; i < 10; i++) {}`},
			{Code: `for (const x of [1,2]) {}`},
			{Code: `declare global { var bar: string; }`},
			{Code: "declare global {\n  var g1: string;\n  var g2: number;\n}"},
		},
		[]rule_tester.InvalidTestCase{
			// ================================================================
			// Explicit TypeScript script mode (global scope → no fix)
			// ================================================================
			{
				Code:            `var foo = bar;`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
				},
			},
			{
				Code:            `var foo = bar, toast = most;`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
				},
			},
			{
				Code:            `if (true) { var x = 1; }`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 13},
				},
			},
			{
				Code:            `for (var i = 0; i < 10; i++) {}`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 6},
				},
			},
			{
				Code:            `var { a, b } = { a: 1, b: 2 };`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
				},
			},
			{
				Code:            `declare var declaredVar: number;`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Module mode: basic fixes (var → let)
			// ================================================================
			{
				Code:     `var defaultModule = 1;`,
				FileName: "default-module.js",
				TSConfig: "tsconfig.allow-js.json",
				Output:   []string{`let defaultModule = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			{
				Code:     `var commonJS = 1;`,
				FileName: "commonjs.cjs",
				TSConfig: "tsconfig.allow-js.json",
				Output:   []string{`let commonJS = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			{
				Code:     `var event = 1;`,
				FileName: "global-name.js",
				TSConfig: "tsconfig.allow-js.json",
				Output:   []string{`let event = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			{
				Code:   `export {}; var foo = 1;`,
				Output: []string{`export {}; let foo = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			{
				Code:   `export {}; var foo = 1, toast = 2;`,
				Output: []string{`export {}; let foo = 1, toast = 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			{
				Code:   `export {}; var foo = 1; let toast = 2;`,
				Output: []string{`export {}; let foo = 1; let toast = 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Type-only exports do not reference a value-only var, so they do
			// not create a false "used before declaration" blocker for the fix.
			{
				Code:   `export type { x }; var x = 1;`,
				Output: []string{`export type { x }; let x = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			{
				Code:   `export { type x }; var x = 1;`,
				Output: []string{`export { type x }; let x = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			{
				Code:   `export type { x as y }; var x = 1;`,
				Output: []string{`export type { x as y }; let x = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			{
				Code:   `export type { HTMLElement }; var HTMLElement = 1;`,
				Output: []string{`export type { HTMLElement }; let HTMLElement = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// A UMD namespace export names the module as a whole; it is not a
			// runtime read of a same-named local binding.
			{
				Code:   `export as namespace API; var API = 1; export = API;`,
				Output: []string{`export as namespace API; let API = 1; export = API;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Import attribute keys are syntax names, not variable references.
			{
				Code:   `import data from "pkg" with { type: "json" }; var type = 1;`,
				Output: []string{`import data from "pkg" with { type: "json" }; let type = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Multiple var statements, both fixable
			{
				Code:   `export {}; var foo = 1; var bar = 2;`,
				Output: []string{`export {}; let foo = 1; let bar = 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
					{MessageId: "unexpectedVar"},
				},
			},
			// Module-mode for-of
			{
				Code:   "export {}; for (var a of [1]) { console.log(a); }",
				Output: []string{"export {}; for (let a of [1]) { console.log(a); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Module-mode for-in
			{
				Code:   "export {}; for (var a in {x:1}) { console.log(a); }",
				Output: []string{"export {}; for (let a in {x:1}) { console.log(a); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Module-mode for-loop
			{
				Code:   `export {}; for (var i = 0; i < 10; i++) {}`,
				Output: []string{`export {}; for (let i = 0; i < 10; i++) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Initialized var in loop body → fix (condition 9 only blocks UNinitialized)
			{
				Code:   "export {}; for (let a of [1]) { var c = 1; console.log(c); }",
				Output: []string{"export {}; for (let a of [1]) { let c = 1; console.log(c); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Destructuring
			{
				Code:   `export {}; var { a, b } = { a: 1, b: 2 };`,
				Output: []string{`export {}; let { a, b } = { a: 1, b: 2 };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			{
				Code:   `export {}; var [c, d] = [1, 2];`,
				Output: []string{`export {}; let [c, d] = [1, 2];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// export var
			{
				Code: `export var exported = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 8, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: `export declare var exported: number;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 8, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code: `export /* gap */ declare var exported: number;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 18, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code: `export var exported = 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 8, EndLine: 1, EndColumn: 24},
				},
			},
			// declare var in module
			{
				Code:   `export {}; declare var x: number;`,
				Output: []string{`export {}; declare let x: number;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// declare namespace
			{
				Code:   `declare namespace NS { var nsVar: string; }`,
				Output: []string{`declare namespace NS { let nsVar: string; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 24},
				},
			},
			// Only declarations directly in `declare global` are ignored. A nested
			// namespace has its own TSModuleBlock and is reported by ESLint.
			{
				Code:   `declare global { namespace NS { var nested: string; } }`,
				Output: []string{`declare global { namespace NS { let nested: string; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 33, EndLine: 1, EndColumn: 52},
				},
			},
			// declare module
			{
				Code:   `declare module 'my-mod' { var modVar: string; }`,
				Output: []string{`declare module 'my-mod' { let modVar: string; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 27},
				},
			},
			// var in nested function
			{
				Code:   `function outer() { var nested = 1; }`,
				Output: []string{`function outer() { let nested = 1; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 1: switch case (no fix)
			// ================================================================
			{
				Code: "switch (0) { case 0: var sw = 1; break; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			{
				Code: "export {}; switch (0) { case 0: var sw = 1; break; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// A for initializer is not directly parented by SwitchCase and remains fixable.
			{
				Code:   "export {}; switch (0) { case 0: for (var i = 0; i < 1; i++) {} }",
				Output: []string{"export {}; switch (0) { case 0: for (let i = 0; i < 1; i++) {} }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 2: TDZ (no fix)
			// ================================================================
			// Self-reference in initializer
			{
				Code: `export {}; function f() { var a = a; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// IIFE — init is NOT a function, ref is inside init range
			{
				Code: "export {}; var foo = (function() { foo(); })();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Wrapped call — same reason
			{
				Code: "export {}; var foo = bar(function() { foo(); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Destructuring default self-ref: {a = a}
			{
				Code: "export {}; function f() { var {a = a} = {}; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Destructuring default self-ref: {foo = foo}
			{
				Code: "export {}; var { foo = foo } = {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Forward reference in same declaration: var a = b, b = 1
			// (caught by condition 7, not TDZ, but still no fix)
			{
				Code: "export {}; function f() { var a = b, b = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Forward reference in destructuring default: {a = b, b}
			// (caught by condition 7)
			{
				Code: "export {}; function f() { var {a = b, b} = {}; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Forward ref across statements: var bar = foo, foo = fn
			{
				Code: "export {}; var bar = foo, foo = function() { foo(); };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Destructuring forward ref: { bar = foo, foo }
			{
				Code: "export {}; var { bar = foo, foo } = {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 2: safe references (DO fix)
			// ================================================================
			// Function self-reference — deferred, safe
			{
				Code:   "export {}; var foo = function() { foo(); };",
				Output: []string{"export {}; let foo = function() { foo(); };"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Arrow self-reference — deferred, safe
			{
				Code:   "export {}; var foo = () => foo();",
				Output: []string{"export {}; let foo = () => foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Recursive function with default parameter — safe
			{
				Code:   "export {}; var fx = function(i = 0) { if (i < 5) return fx(i + 1); }; fx();",
				Output: []string{"export {}; let fx = function(i = 0) { if (i < 5) return fx(i + 1); }; fx();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Backward reference in destructuring default: {a, b = a} is safe
			{
				Code:   "export {}; function f() { var {a, b = a} = {}; }",
				Output: []string{"export {}; function f() { let {a, b = a} = {}; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// A previous declarator is initialized before a later initializer runs.
			{
				Code:     "var a = 1, b = a;",
				FileName: "backward-reference.js",
				TSConfig: "tsconfig.allow-js.json",
				Output:   []string{"let a = 1, b = a;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 4: redeclared (no fix)
			// ================================================================
			{
				Code: "export {}; function f() { var x = 1; var x = 2; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
					{MessageId: "unexpectedVar"},
				},
			},
			// Redeclared in same for-init
			{
				Code: "export {}; function f() { for (var i = 0, i = 0; false;); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Redeclared: var a; if (b) { var a; }
			{
				Code: "export {}; function f() { var a; if (true) { var a; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 5: used from outside block scope (no fix)
			// ================================================================
			{
				Code: "export {}; function f() { if (true) { var x = 1; } console.log(x); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Used after for-loop
			{
				Code: "export {}; function f() { for (var i = 0; i < 10; ++i) {} console.log(i); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Used after for-in
			{
				Code: "export {}; function f() { for (var a in {}) {} console.log(a); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Used after for-of
			{
				Code: "export {}; function f() { for (var a of []) {} console.log(a); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 6: variable name is `let` (no fix)
			// ================================================================
			{
				Code: "function f() { var let; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// `let` in destructuring
			{
				Code: "function f() { var { let } = {}; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 7: referenced before declaration (no fix)
			// ================================================================
			{
				Code: "export {}; function f() { console.log(x); var x = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Reference in nested block before declaration
			{
				Code: "export {}; function f() { if (true) { console.log(x); } var x = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Reference through function call (hoisting)
			{
				Code: "export {}; function foo() { a; } var a = 1; foo();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Uninitialized var referenced before declaration
			{
				Code: "export function f() { console.log(o); var o; return o; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
			// Reference in nested if before var in same block
			{
				Code: "export {}; function f() { if (true) { console.log(o); var o; return o; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 7: partial fix (first fixes, second doesn't)
			// ================================================================
			// Separate statements: var a = b; var b = 1 → first fixes
			{
				Code:   "export {}; var a = b; var b = 1;",
				Output: []string{"export {}; let a = b; var b = 1;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
					{MessageId: "unexpectedVar"},
				},
			},
			// var y = x; var x = 1 → first fixes
			{
				Code:   "export {}; function f() { var y = x; var x = 1; }",
				Output: []string{"export {}; function f() { let y = x; var x = 1; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
					{MessageId: "unexpectedVar"},
				},
			},
			// Cross-scope: outer a fixes, inner a (hoisted ref) doesn't
			{
				Code:   "export {}; var a = 1; function f() { console.log(a); var a = 2; }",
				Output: []string{"export {}; let a = 1; function f() { console.log(a); var a = 2; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
					{MessageId: "unexpectedVar"},
				},
			},
			// fn expression safe, separate statement with ref before decl not safe
			{
				Code:   "export {}; var bar = function() { foo(); }; var foo = function() {};",
				Output: []string{"export {}; let bar = function() { foo(); }; var foo = function() {};"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 8: closure in loop (no fix)
			// ================================================================
			{
				Code: "export {}; function f() { for (var a of [1]) { setTimeout(() => console.log(a)); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 9: uninitialized in loop (no fix)
			// ================================================================
			{
				Code: "export {}; function f() { for (let i of [1]) { var c; console.log(c); c = 'hello'; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},

			// ================================================================
			// Condition 10: statement position (no fix)
			// ================================================================
			{
				Code: "export {}; function f() { if (true) var bar = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar"},
				},
			},
		},
	)
}

func TestNoVarAutofixSafetyAndScopeParity(t *testing.T) {
	oneError := []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedVar"}}
	twoErrors := []rule_tester.InvalidTestCaseError{
		{MessageId: "unexpectedVar"},
		{MessageId: "unexpectedVar"},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoVarRule,
		nil,
		[]rule_tester.InvalidTestCase{
			// A catch parameter and var are distinct in scope analysis, but var
			// still writes that parameter. Let is either invalid or changes the
			// binding being assigned.
			{Code: `function f() { try {} catch (e) { var e = 1; } }`, Errors: oneError},
			{Code: `function f() { try {} catch (e) { var e; } }`, Errors: oneError},
			{Code: `function f() { try {} catch (e) { { var e = 1; } } }`, Errors: oneError},
			{Code: `function f() { try {} catch (e) { try {} catch (x) { var e = 1; } } }`, Errors: oneError},
			{Code: `function f() { try {} catch (e) { for (var e of []) {} } }`, Errors: oneError},
			{
				Code:   `function f() { try {} catch (e) { var x = 1; } }`,
				Output: []string{`function f() { try {} catch (e) { let x = 1; } }`},
				Errors: oneError,
			},
			{
				Code:   `function f() { try {} catch (e) {} var e = 1; }`,
				Output: []string{`function f() { try {} catch (e) {} let e = 1; }`},
				Errors: oneError,
			},
			{
				Code:   `function f() { try {} catch (e) { function g() { var e = 1; } } }`,
				Output: []string{`function f() { try {} catch (e) { function g() { let e = 1; } } }`},
				Errors: oneError,
			},
			{
				Code:   `function f() { try {} catch (e) { class C { static { var e = 1; } } } }`,
				Output: []string{`function f() { try {} catch (e) { class C { static { let e = 1; } } } }`},
				Errors: oneError,
			},

			// Calling a hoisted declaration before the var would enter the TDZ
			// after the replacement, even if its read is textually later.
			{Code: `function wrap() { foo(); var a = 1; function foo() { console.log(a); } }`, Errors: oneError},
			{
				Code:   `function wrap() { var a = 1; foo(); function foo() { console.log(a); } }`,
				Output: []string{`function wrap() { let a = 1; foo(); function foo() { console.log(a); } }`},
				Errors: oneError,
			},
			{Code: `function wrap() { bar(); var a = 1; function bar() { function inner() { console.log(a); } inner(); } }`, Errors: oneError},
			{
				Code:   `function wrap() { foo(); var a = 1; var b = 2; function foo() { console.log(a); } }`,
				Output: []string{`function wrap() { foo(); var a = 1; let b = 2; function foo() { console.log(a); } }`},
				Errors: twoErrors,
			},
			{
				Code:   `function wrap() { function foo() {} foo(); if (bar) { var a = 1; function foo() { console.log(a); } foo(); } }`,
				Output: []string{`function wrap() { function foo() {} foo(); if (bar) { let a = 1; function foo() { console.log(a); } foo(); } }`},
				Errors: oneError,
			},
			{Code: `function wrap() { new Foo(); var a = 1; function Foo() { console.log(a); } }`, Errors: oneError},
			{Code: "function wrap() { tag``; var a = 1; function tag() { console.log(a); } }", Errors: oneError},
			{Code: `function wrap() { (foo)(); var a = 1; function foo() { console.log(a); } }`, Errors: oneError},
			{Code: `function wrap() { foo.call(null); var a = 1; function foo() { console.log(a); } }`, Errors: oneError},
			{Code: `function wrap() { foo.apply(null, []); var a = 1; function foo() { console.log(a); } }`, Errors: oneError},
			{Code: `function wrap() { foo["call"](null); var a = 1; function foo() { console.log(a); } }`, Errors: oneError},
			{Code: `function wrap() { (0, foo)(); var a = 1; function foo() { console.log(a); } }`, Errors: oneError},
			{Code: `function wrap(flag) { (flag ? foo : other)(); var a = 1; function foo() { console.log(a); } }`, Errors: oneError},
			{Code: `function wrap() { (foo || other)(); var a = 1; function foo() { console.log(a); } }`, Errors: oneError},
			{Code: `function wrap() { var a = foo(); function foo() { return a; } }`, Errors: oneError},
			{Code: `function wrap(values) { for (var a of foo()) {} function foo() { return [a]; } }`, Errors: oneError},
			{Code: `function wrap() { var b = foo(), a = 1; function foo() { console.log(a); } }`, Errors: oneError},
			{
				Code:   `function wrap() { var a = 1, b = foo(); function foo() { console.log(a); } }`,
				Output: []string{`function wrap() { let a = 1, b = foo(); function foo() { console.log(a); } }`},
				Errors: oneError,
			},
			{
				Code:   `function wrap() { foo(); var a = 1; function foo(): typeof a { return 1; } }`,
				Output: []string{`function wrap() { foo(); let a = 1; function foo(): typeof a { return 1; } }`},
				Errors: oneError,
			},
			{
				Code:   `function wrap() { var a = () => foo(); function foo() { return a; } }`,
				Output: []string{`function wrap() { let a = () => foo(); function foo() { return a; } }`},
				Errors: oneError,
			},

			// Parentheses are not an execution boundary in ESTree. Empty binding
			// patterns declare no variable whose scope could make the fix unsafe.
			{
				Code:   `var foo = ((function () { foo(); }));`,
				Output: []string{`let foo = ((function () { foo(); }));`},
				Errors: oneError,
			},
			{
				Code:   `var foo = (((() => foo())));`,
				Output: []string{`let foo = (((() => foo())));`},
				Errors: oneError,
			},
			{
				Code:   `var foo = ((function () { foo(); }) as () => void);`,
				Output: []string{`let foo = ((function () { foo(); }) as () => void);`},
				Errors: oneError,
			},
			{
				Code:   `function f() { type T = typeof value; var value = null as typeof value; }`,
				Output: []string{`function f() { type T = typeof value; let value = null as typeof value; }`},
				Errors: oneError,
			},
			{Code: `function f() { var value = value as number; }`, Errors: oneError},
			{Code: `function f() { var value = typeof value; }`, Errors: oneError},
			{
				Code:   `function f() { var {} = {}; var [] = []; }`,
				Output: []string{`function f() { let {} = {}; let [] = []; }`},
				Errors: twoErrors,
			},
			{
				Code:   `function f(values) { for (var {} of values) {} }`,
				Output: []string{`function f(values) { for (let {} of values) {} }`},
				Errors: oneError,
			},
			{Code: `function f() { for (var x of x) {} }`, Errors: oneError},
			{Code: `function f() { for (var x in x) {} }`, Errors: oneError},
			{Code: `function f(values) { for (var [x = x] of values) {} }`, Errors: oneError},
			{Code: `function f(values) { for (var {x = x} of values) {} }`, Errors: oneError},

			// Function declarations and vars may legally share a binding, but a
			// let replacement may not. A nested function owns a separate binding.
			{Code: `function X() {} var X = 1;`, Errors: oneError},
			{Code: `var X = 1; function X() {}`, Errors: oneError},
			{Code: `function outer() { function X() {} var X = 1; }`, Errors: oneError},
			{Code: `function outer() { var X = 1; function X() {} }`, Errors: oneError},
			{Code: `function f(value: number) { var value = 1; }`, Errors: oneError},
			{Code: `function outer() { function X() {} { var X = 1; } }`, Errors: oneError},
			{Code: `declare namespace N { function X(): void; var X: unknown; }`, Errors: oneError},
			{
				Code:   `function outer() { function X() {} function inner() { var X = 1; } }`,
				Output: []string{`function outer() { function X() {} function inner() { let X = 1; } }`},
				Errors: oneError,
			},

			// These class/object/type positions execute in the current iteration.
			// Instance fields and methods are deferred and must keep var.
			{
				Code:   `function f(values) { for (var x of values) { class C { static { use(x); } } } }`,
				Output: []string{`function f(values) { for (let x of values) { class C { static { use(x); } } } }`},
				Errors: oneError,
			},
			{
				Code:   `function f(values) { for (var x of values) { class C { [x]() {} } } }`,
				Output: []string{`function f(values) { for (let x of values) { class C { [x]() {} } } }`},
				Errors: oneError,
			},
			{
				Code:   `function f(values) { for (var x of values) { class C { static field = x; } } }`,
				Output: []string{`function f(values) { for (let x of values) { class C { static field = x; } } }`},
				Errors: oneError,
			},
			{
				Code:   `function f(values) { for (var x of values) { ({ [x]: true }); } }`,
				Output: []string{`function f(values) { for (let x of values) { ({ [x]: true }); } }`},
				Errors: oneError,
			},
			{
				Code:   `function f(values) { for (var x of values) { type T = () => typeof x; } }`,
				Output: []string{`function f(values) { for (let x of values) { type T = () => typeof x; } }`},
				Errors: oneError,
			},
			{
				Code:   `function f(values) { for (var x of values) { @decorate(x) class C {} } }`,
				Output: []string{`function f(values) { for (let x of values) { @decorate(x) class C {} } }`},
				Errors: oneError,
			},
			{
				Code:   `function f(values) { for (var x of values) { class C { @decorate(x) method() {} } } }`,
				Output: []string{`function f(values) { for (let x of values) { class C { @decorate(x) method() {} } } }`},
				Errors: oneError,
			},
			{
				Code:   `function f(values) { for (var x of values) { class C { method(@decorate(x) value: unknown) {} } } }`,
				Output: []string{`function f(values) { for (let x of values) { class C { method(@decorate(x) value: unknown) {} } } }`},
				Errors: oneError,
			},
			{Code: `function f(values) { for (var x of values) { class C { method(@decorate(() => x) value: unknown) {} } } }`, Errors: oneError},
			{
				Code:   `function f(values) { for (var x of values) { class C { field: typeof x; } } }`,
				Output: []string{`function f(values) { for (let x of values) { class C { field: typeof x; } } }`},
				Errors: oneError,
			},
			{Code: `function f(values) { for (var x of values) { class C { field = x; } consume(C); } }`, Errors: oneError},
			{Code: `function f(values) { for (var x of values) { class C { accessor field = x; } consume(C); } }`, Errors: oneError},
			{
				Code:   `function f(values) { for (var x of values) { class C { static accessor field = x; } } }`,
				Output: []string{`function f(values) { for (let x of values) { class C { static accessor field = x; } } }`},
				Errors: oneError,
			},
			{Code: `function f(values) { for (var x of values) { class C { method() { return x; } } } }`, Errors: oneError},
			{Code: `function f(values) { for (var x of values) { class C { static { consume(() => x); } } } }`, Errors: oneError},
			{
				Code:   `function f(values) { for (const value of values) { class C { static { var x = value; consume(() => x); } } } }`,
				Output: []string{`function f(values) { for (const value of values) { class C { static { let x = value; consume(() => x); } } } }`},
				Errors: oneError,
			},
			{Code: `class C { static { for (var x of values) { consume(() => x); } } }`, Errors: oneError},

			// Ambient namespace declarations merge across separate ModuleBlocks.
			{
				Code:   `declare namespace N { var x: number; } declare namespace N { const y: typeof x; }`,
				Output: []string{`declare namespace N { let x: number; } declare namespace N { const y: typeof x; }`},
				Errors: oneError,
			},
			{
				Code:   `declare namespace N { const y: typeof x; } declare namespace N { var x: number; }`,
				Output: []string{`declare namespace N { const y: typeof x; } declare namespace N { let x: number; }`},
				Errors: oneError,
			},
			{
				Code:   `declare namespace N.M { var x: number; } declare namespace N.M { type Y = typeof x; }`,
				Output: []string{`declare namespace N.M { let x: number; } declare namespace N.M { type Y = typeof x; }`},
				Errors: oneError,
			},
			{
				Code:   `declare module "pkg" { var x: number; } declare module "pkg" { type Y = typeof x; }`,
				Output: []string{`declare module "pkg" { let x: number; } declare module "pkg" { type Y = typeof x; }`},
				Errors: oneError,
			},
			{
				Code:   `declare namespace N { var x: number; } declare namespace N { var x: number; }`,
				Errors: twoErrors,
			},
		},
	)
}
