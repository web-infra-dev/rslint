package no_new_func

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewFuncRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewFuncRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// --- Not the global Function ---
			{Code: `var a = new _function("b", "c", "return b+c");`},
			{Code: `var a = _function("b", "c", "return b+c");`},

			// --- Function as a value reference, not invoked ---
			{Code: `call(Function)`},
			{Code: `new Class(Function)`},
			{Code: `var x = [Function]`},
			{Code: `var x = Function`},
			{Code: `typeof Function`},

			// --- Non-matching method calls ---
			{Code: `Function.toString()`},
			{Code: `Function.hasOwnProperty("call")`},
			{Code: `Function.prototype`},

			// --- Dynamic/computed property: not statically "call"/"apply"/"bind" ---
			{Code: `foo[Function]()`},
			{Code: `Function[call]()`},

			// --- Accessing but not calling .bind/.call/.apply ---
			{Code: `foo(Function.bind)`},
			{Code: `var x = Function.call`},
			{Code: `var x = Function.apply`},

			// --- Shadowing: class declaration ---
			{Code: `class Function {}; new Function()`},
			{Code: `const fn = () => { class Function {}; new Function() }`},

			// --- Shadowing: function declaration ---
			{Code: `function Function() {}; Function()`},
			{Code: `var fn = function () { function Function() {}; Function() }`},

			// --- Shadowing: function expression name ---
			{Code: `var x = function Function() { Function(); }`},

			// --- Shadowing: var (hoisted across blocks) ---
			{Code: `function test() { var Function = function(){}; return new Function(); }`},
			{Code: `function test() { var x = new Function("code"); var Function = function() {}; }`},
			{Code: `function test() { if (true) { var Function = 42; } new Function(); }`},
			{Code: `function test() { for (var Function = 0; Function < 1; Function++) {} new Function(); }`},
			{Code: `function test() { for (var Function in {}) {} new Function(); }`},
			{Code: `function test() { for (var Function of []) {} new Function(); }`},
			{Code: `function test() { switch (0) { case 0: var Function = 1; } new Function(); }`},

			// --- Shadowing: let/const ---
			{Code: `function test() { let Function = class {}; return new Function(); }`},
			{Code: `function test() { const Function = class {}; return Function(); }`},

			// --- Shadowing: parameter ---
			{Code: `function test(Function) { return new Function(); }`},
			{Code: `function test({ Function }) { return new Function(); }`},
			{Code: `function test([Function]) { return new Function(); }`},
			{Code: `function test(...Function) { return new Function(); }`},
			{Code: `function test(Function = class {}) { return new Function(); }`},
			{Code: `var fn = (Function) => Function();`},
			{Code: `function* gen(Function) { yield new Function(); }`},
			{Code: `async function af(Function) { return new Function(); }`},

			// --- Shadowing: catch clause ---
			{Code: `try {} catch (Function) { new Function(); }`},

			// --- Shadowing: nested scopes ---
			{Code: `function test() { var Function = class {}; function inner() { return new Function(); } }`},
			{Code: `function test() { var Function = class {}; var fn = () => new Function(); }`},

			// --- Shadowing: method/constructor parameters ---
			{Code: `var obj = { m(Function) { return new Function(); } };`},
			{Code: `class C { m(Function) { return new Function(); } }`},
			{Code: `class C { constructor(Function) { this.x = new Function(); } }`},

			// --- Shadowing: for-let/of (inside loop body) ---
			{Code: `function test() { for (let Function in {}) { new Function(); } }`},
			{Code: `function test() { for (let Function of []) { new Function(); } }`},

			// --- Shadowing applies to .call/.apply/.bind too ---
			{Code: `function test(Function) { return Function.call(null, "code"); }`},
			{Code: `function test() { var Function = class {}; Function.apply(null, ["code"]); }`},

			// --- Tagged template (not a call) ---
			{Code: "Function`code`"},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// === Direct: new Function(...) ===
			{
				Code:   `var a = new Function("b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},
			// No arguments
			{
				Code:   `new Function()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},

			// === Direct: Function(...) ===
			{
				Code:   `var a = Function("b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},

			// === Parenthesized callee ===
			{
				Code:   `(Function)("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `((Function))("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `new (Function)("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},

			// === Optional call ===
			{
				Code:   `Function?.("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},

			// === Method: .call / .apply / .bind (dot notation) ===
			{
				Code:   `var a = Function.call(null, "b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},
			{
				Code:   `var a = Function.apply(null, ["b", "c", "return b+c"]);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},
			{
				Code:   `var a = Function.bind(null, "b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},
			// .bind(...)() — only the inner Function.bind(...) call is reported
			{
				Code:   `var a = Function.bind(null, "b", "c", "return b+c")();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},

			// === Method: bracket notation ===
			{
				Code:   `var a = Function["call"](null, "b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},
			{
				Code:   `var a = Function["apply"](null, ["b", "c", "return b+c"]);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},
			{
				Code:   `var a = Function["bind"](null, "b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},
			// Template literal bracket notation
			{
				Code:   "var a = Function[`call`](null, \"code\")",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},

			// === Optional chaining on method ===
			{
				Code:   `(Function?.call)(null, "b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function?.call(null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function?.apply(null, ["code"])`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function?.bind(null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},

			// === Parenthesized object in method call ===
			{
				Code:   `(Function).call(null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `(Function).apply(null, ["code"])`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},

			// === TypeScript assertions on callee ===
			{
				Code:   `(Function as any)("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `(<any>Function)("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function!("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `(Function satisfies any)("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `new (Function as any)("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			// TypeScript assertion on object of method call
			{
				Code:   `(Function as any).call(null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},

			// === Nested new wrapping a Function call ===
			{
				Code:   `new (Function("code"))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 6}},
			},

			// === Nesting: inside various constructs ===
			{
				Code:   `function f() { return new Function("code"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 23}},
			},
			{
				Code:   `var f = () => new Function("code");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 15}},
			},
			{
				Code:   `if (true) { var x = new Function("code"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 21}},
			},
			{
				Code:   `class C { m() { return new Function("code"); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 24}},
			},
			{
				Code:   `class C { constructor() { this.x = new Function("code"); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 36}},
			},
			{
				Code:   `function outer() { function inner() { return new Function("code"); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 46}},
			},

			// === Multiple errors in one statement ===
			{
				Code: `var a = new Function("a"); var b = Function("b");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 9},
					{MessageId: "noFunctionConstructor", Line: 1, Column: 36},
				},
			},

			// === Expressions: ternary, logical, comma ===
			{
				Code: `var x = true ? new Function("a") : Function("b");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 16},
					{MessageId: "noFunctionConstructor", Line: 1, Column: 36},
				},
			},
			{
				Code:   `var x = foo || new Function("code");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 16}},
			},
			{
				Code:   `var x = foo ?? new Function("code");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 16}},
			},

			// === Scoping: inner shadow does NOT reach outer ===
			{
				Code:   `const fn = () => { class Function {} }; new Function('', '')`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 41}},
			},
			{
				Code:   `var fn = function () { function Function() {} }; Function('', '')`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 50}},
			},
			// let in sibling block does not shadow
			{
				Code:   `function f() { { let Function = class {}; } var x = new Function("code"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 53}},
			},
			// var in inner function does NOT hoist to outer
			{
				Code:   `function f() { var x = new Function("code"); (function() { var Function = 1; })(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 24}},
			},
			// arrow param does not shadow outer scope
			{
				Code:   `function f() { var fn = (Function) => Function; var x = new Function("code"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 57}},
			},
			// catch variable does not shadow outside catch block
			{
				Code:   `function f() { try {} catch (Function) {} var x = new Function("code"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 51}},
			},
			// for-let does not shadow outside the loop
			{
				Code:   `function f() { for (let Function of []) {} var x = new Function("code"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 52}},
			},
		},
	)
}
