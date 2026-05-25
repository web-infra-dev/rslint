package no_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
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
			// Script mode (no import/export → global scope → no fix)
			// ================================================================
			{
				Code: `var foo = bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
				},
			},
			{
				Code: `var foo = bar, toast = most;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
				},
			},
			{
				Code: `if (true) { var x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 13},
				},
			},
			{
				Code: `for (var i = 0; i < 10; i++) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 6},
				},
			},
			{
				Code: `var { a, b } = { a: 1, b: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
				},
			},
			{
				Code: `declare var declaredVar: number;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Module mode: basic fixes (var → let)
			// ================================================================
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
				Code:   `export var exported = 1;`,
				Output: []string{`export let exported = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 8},
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
