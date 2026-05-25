package no_new_symbol

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewSymbolRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewSymbolRule,
		// Valid cases — Symbol is shadowed by a local declaration, no error expected
		[]rule_tester.ValidTestCase{
			// Basic non-constructor usage
			{Code: `var foo = Symbol('foo');`},
			{Code: `new foo(Symbol);`},
			{Code: `new foo(bar, Symbol);`},

			// Shadowing by declaration type
			{Code: `var Symbol = function() {}; new Symbol();`},
			{Code: `let Symbol = 1; new Symbol();`},
			{Code: `const Symbol = null; new Symbol();`},
			{Code: `function Symbol() {} new Symbol();`},
			{Code: `class Symbol {} new Symbol();`},

			// Shadowing by parameter type
			{Code: `function bar(Symbol) { var baz = new Symbol('baz'); }`},
			{Code: `const f = (Symbol) => { new Symbol(); };`},
			{Code: `function f(...Symbol) { new Symbol(); }`},
			{Code: `function f({ Symbol }) { new Symbol(); }`},
			{Code: `function f(Symbol = 1) { new Symbol(); }`},

			// Shadowing by destructuring
			{Code: `var { Symbol } = obj; new Symbol();`},
			{Code: `var [Symbol] = arr; new Symbol();`},
			{Code: `var { a: { Symbol } } = obj; new Symbol();`},
			{Code: `var { ["Symbol"]: Symbol } = obj; new Symbol();`},

			// Shadowing by loop variable
			{Code: `for (var Symbol = 0;;) { new Symbol(); }`},
			{Code: `for (var Symbol in obj) { new Symbol(); }`},
			{Code: `for (let Symbol of arr) { new Symbol(); }`},
			{Code: `for (const Symbol of arr) { new Symbol(); }`},

			// Shadowing by catch clause
			{Code: `try {} catch(Symbol) { new Symbol(); }`},

			// Top-level shadow affects all inner scopes
			{Code: `function Symbol() {} function f() { new Symbol(); }`},
			{Code: `var Symbol = 1; function f() { new Symbol(); }`},
			{Code: `var Symbol = 1; const f = () => new Symbol();`},
			{Code: `var Symbol = 1; class C { m() { new Symbol(); } }`},

			// Hoisting: var/function declarations hoist to top of scope
			{Code: `new Symbol(); var Symbol = 1;`},
			{Code: `new Symbol(); function Symbol() {}`},
		},
		// Invalid cases — Symbol refers to the global built-in
		[]rule_tester.InvalidTestCase{
			// Basic cases
			{
				Code: `var foo = new Symbol('foo');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewSymbol", Line: 1, Column: 15},
				},
			},
			{
				Code: `new Symbol()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewSymbol", Line: 1, Column: 5},
				},
			},

			// Nested function expression doesn't shadow global
			{
				Code: "function bar() { return function Symbol() {}; } var baz = new Symbol('baz');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewSymbol", Line: 1, Column: 63},
				},
			},

			// Block-scoped declarations don't shadow outside the block
			{
				Code:   "{ function Symbol() {} } new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "{ let Symbol = 1; } new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "{ const Symbol = 1; } new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "if (true) { let Symbol = 1; } new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "try { let Symbol = 1; } catch(e) {} new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},

			// Function/arrow scoped var doesn't shadow outside
			{
				Code:   "function foo() { var Symbol = 1; } new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "const f = () => { var Symbol = 1; }; new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "function a() { function b() { var Symbol = 1; } } new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},

			// IIFE parameters don't shadow outside
			{
				Code:   "(function(Symbol) {})(1); new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "((Symbol) => {})(1); new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},

			// Class static block scoped var doesn't shadow outside
			{
				Code:   "class C { static { var Symbol = 1; } } new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},

			// new Symbol() inside scopes without any shadow
			{
				Code:   "function f() { new Symbol(); }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "const f = () => new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "class C { m() { new Symbol(); } }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "if (true) { if (true) { new Symbol(); } }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},
			{
				Code:   "class C { static { new Symbol(); } }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNewSymbol"}},
			},

			// Mixed: shadow in one scope, global in another — only outer reports
			{
				Code: "function f(Symbol) { new Symbol(); } new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewSymbol", Line: 1, Column: 42},
				},
			},
			{
				Code: "{ let Symbol = 1; new Symbol(); } new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewSymbol", Line: 1, Column: 39},
				},
			},
		},
	)
}
