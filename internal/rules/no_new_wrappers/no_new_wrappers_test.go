package no_new_wrappers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewWrappersRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewWrappersRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Non-wrapper constructors
			{Code: `var a = new Object();`},
			{Code: `var a = new Map();`},
			{Code: `var a = new Date();`},
			// Function call (not constructor)
			{Code: `var a = String('test'), b = String.fromCharCode(32);`},
			// --- Shadowing: function parameters ---
			{Code: `function test(Number: any) { return new Number; }`},
			{Code: `function test({ Number }: any) { return new Number(); }`},
			{Code: `function test([Boolean]: any) { return new Boolean(); }`},
			{Code: `function test({ a: { String } }: any) { return new String(); }`},
			{Code: `function test(...Number: any) { return new Number(); }`},
			{Code: `function test(Boolean: any = true) { return new Boolean(); }`},
			{Code: `function test(a: any, Boolean: any, c: any) { return new Boolean(); }`},
			{Code: `var fn = (String: any) => new String();`},
			{Code: `function* gen(String: any) { yield new String(); }`},
			{Code: `async function af(Number: any) { return new Number(); }`},
			{Code: `var af = async (Boolean: any) => new Boolean();`},
			// --- Shadowing: var (hoisted) ---
			{Code: `function test() { var Boolean: any = function(){}; return new Boolean(true); }`},
			{Code: `function test() { var x = new String('hello'); var String: any = function() {}; }`},
			{Code: `function test() { if (true) { var Number: any = 42; } var v = new Number(); }`},
			{Code: `function test() { for (var Boolean: any = 0; Boolean < 1; Boolean++) {} new Boolean(); }`},
			{Code: `function test() { for (var String in {}) {} new String(); }`},
			{Code: `function test() { for (var Number of []) {} new Number(); }`},
			{Code: `function test() { switch (0) { case 0: var Number: any = 1; } new Number(); }`},
			// --- Shadowing: let/const (block-scoped) ---
			{Code: `function test() { let String: any = class {}; return new String('x'); }`},
			{Code: `function test() { const Boolean: any = class {}; return new Boolean(); }`},
			// --- Shadowing: function/class declaration ---
			{Code: `function test() { function Number() {} new Number(); }`},
			{Code: `function test() { class String { constructor() {} }; new String(); }`},
			// --- Shadowing: function expression name ---
			{Code: `var fn = function String() { return new String(); };`},
			// --- Shadowing: catch clause ---
			{Code: `try {} catch (Number) { new Number(); }`},
			// --- Shadowing: nested scopes ---
			{Code: `function test() { var Boolean: any = function() {}; function inner() { return new Boolean(); } }`},
			{Code: `function test() { var String: any = class {}; function mid() { function deep() { return new String(); } } }`},
			{Code: `function test() { var Number: any = class {}; var fn = () => new Number(); }`},
			// --- Shadowing: method/constructor parameters ---
			{Code: `var obj = { m(Boolean: any) { return new Boolean(); } };`},
			{Code: `class C { m(String: any) { return new String(); } }`},
			{Code: `class C { constructor(Number: any) { this.x = new Number(); } }`},
			// --- Shadowing: for-let/of (inside loop body) ---
			{Code: `function test() { for (let Boolean in {}) { new Boolean(); } }`},
			{Code: `function test() { for (let String of []) { new String(); } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// --- Basic ---
			{
				Code:   `var a = new String('hello');`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 9}},
			},
			{
				Code:   `var a = new Number(10);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 9}},
			},
			{
				Code:   `var a = new Boolean(false);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 9}},
			},
			// No parentheses
			{
				Code:   `var a = new String;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 9}},
			},
			// Multiple in one statement
			{
				Code: `var a = new String('a'); var b = new Number(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noConstructor", Line: 1, Column: 9},
					{MessageId: "noConstructor", Line: 1, Column: 34},
				},
			},
			// --- Nesting: inside various constructs ---
			{
				Code:   `function f() { return new String('x'); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 23}},
			},
			{
				Code:   `var f = () => new Number(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 15}},
			},
			{
				Code:   `if (true) { var x = new Boolean(false); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 21}},
			},
			{
				Code:   `class C { m() { return new Number(42); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 24}},
			},
			{
				Code:   `class C { constructor() { this.x = new Boolean(true); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 36}},
			},
			{
				Code:   `function outer() { function inner() { return new Number(1); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 46}},
			},
			// --- Scoping: shadow does NOT reach ---
			// let in sibling block
			{
				Code:   `function f() { { let String: any = class {}; } var x = new String('out'); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 56}},
			},
			// var in inner function does NOT hoist to outer
			{
				Code:   `function f() { var x = new Boolean(true); (function() { var Boolean: any = 1; })(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 24}},
			},
			// arrow param does not shadow outer scope
			{
				Code:   `function f() { var fn = (Number: any) => Number; var x = new Number(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 58}},
			},
			// catch variable does not shadow outside catch block
			{
				Code:   `function f() { try {} catch (String) {} var x = new String('after'); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 49}},
			},
			// for-let does not shadow outside the loop
			{
				Code:   `function f() { for (let Boolean of []) {} var x = new Boolean(false); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noConstructor", Line: 1, Column: 51}},
			},
			// --- Expressions ---
			{
				Code: `var x = true ? new String('a') : new Number(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noConstructor", Line: 1, Column: 16},
					{MessageId: "noConstructor", Line: 1, Column: 34},
				},
			},
		},
	)
}
