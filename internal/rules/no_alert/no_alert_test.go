package no_alert

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAlertRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAlertRule,
		[]rule_tester.ValidTestCase{
			// ================================================================
			// Non-prohibited calls
			// ================================================================
			{Code: `a[o.k](1)`},
			{Code: `foo.alert(foo)`},
			{Code: `foo.confirm(foo)`},
			{Code: `foo.prompt(foo)`},
			{Code: `console.alert()`},
			{Code: `window.scroll()`},
			{Code: `window.focus()`},

			// ================================================================
			// Shadowing: function declaration
			// ================================================================
			{Code: `function alert() {} alert();`},
			{Code: `function confirm() {} confirm();`},
			{Code: `function prompt() {} prompt();`},

			// ================================================================
			// Shadowing: variable declaration (var / let / const)
			// ================================================================
			{Code: `var alert = function() {}; alert();`},
			{Code: `let alert = 1; alert();`},
			{Code: `const alert = () => {}; alert();`},

			// ================================================================
			// Shadowing: destructuring
			// ================================================================
			{Code: `const { alert } = obj; alert();`},
			{Code: `const { a: alert } = obj; alert();`},
			{Code: `const [alert] = arr; alert();`},

			// ================================================================
			// Shadowing: function parameter
			// ================================================================
			{Code: `function foo(alert) { alert(); }`},
			{Code: `const foo = (alert) => { alert(); }`},
			{Code: `function foo({ alert }) { alert(); }`},
			{Code: `function foo([alert]) { alert(); }`},

			// ================================================================
			// Shadowing: inner scope (variable stays in scope for nested calls)
			// ================================================================
			{Code: `function foo() { var alert = bar; alert(); }`},
			{Code: `var alert = function() {}; function test() { alert(); }`},
			{Code: `function foo() { var alert = function() {}; function test() { alert(); } }`},

			// ================================================================
			// Shadowing: class declaration
			// ================================================================
			{Code: `class alert {} new alert();`},

			// ================================================================
			// Shadowing: catch clause
			// ================================================================
			{Code: `try {} catch(alert) { alert(); }`},

			// ================================================================
			// Shadowing: import
			// ================================================================
			{Code: `import alert from 'foo'; alert();`},
			{Code: `import { alert } from 'foo'; alert();`},

			// ================================================================
			// Shadowing: for-loop / for-of / for-in
			// ================================================================
			{Code: `for (let alert = 0; alert < 10; alert++) {}`},
			{Code: `for (const alert of arr) { alert(); }`},
			{Code: `for (const alert in obj) { alert(); }`},

			// ================================================================
			// Shadowing: enum
			// ================================================================
			{Code: `enum alert { A, B }`},

			// ================================================================
			// Dynamic property access (no static name)
			// ================================================================
			{Code: `window[alert]();`},
			{Code: `window[x + y]();`},

			// ================================================================
			// this: inside function (not global scope)
			// ================================================================
			{Code: `function foo() { this.alert(); }`},
			{Code: `const foo = function() { this.alert(); }`},

			// ================================================================
			// this: inside arrow function (ESLint scope = "function", not "global")
			// ================================================================
			{Code: `const foo = () => this.alert();`},
			{Code: `const foo = () => { this.alert(); }`},
			{Code: `const f = () => { const g = () => { this.alert(); } }`},

			// ================================================================
			// this: inside class (class body creates its own this binding)
			// ================================================================
			{Code: `class A { foo() { this.alert(); } }`},
			{Code: `class A { constructor() { this.alert(); } }`},
			{Code: `class A { get x() { return this.alert(); } }`},
			{Code: `class A { set x(v) { this.alert(); } }`},
			// class field initializer — this = instance
			{Code: `class A { x = this.alert(); }`},
			// class static field initializer — this = class
			{Code: `class A { static x = this.alert(); }`},
			// class static block — this = class
			{Code: `class A { static { this.alert(); } }`},
			// computed property name in class — this = outer scope (class context)
			{Code: `class A { [this.x]() {} }`},
			// nested: arrow inside class method
			{Code: `class A { foo() { const f = () => this.alert(); } }`},

			// ================================================================
			// Shadowed window / globalThis
			// ================================================================
			{Code: `function foo() { var window = bar; window.alert(); }`},
			{Code: `const window = {}; window.alert();`},
			{Code: `let window = {}; window.alert();`},
			{Code: `var globalThis = foo; globalThis.alert();`},
			{Code: `function foo() { var globalThis = foo; globalThis.alert(); }`},
			{Code: `import window from 'w'; window.alert();`},

			// ================================================================
			// Callee is a chained member expression (not direct prohibited call)
			// ================================================================
			{Code: `alert.call(null)`},
			{Code: `alert.apply(null, [])`},
			{Code: `window.alert.call(null)`},

			// ================================================================
			// Not a call (no CallExpression)
			// ================================================================
			{Code: `var x = alert;`},
			{Code: `var x = window.alert;`},
			{Code: `typeof alert;`},
			{Code: `new alert()`},
		},
		[]rule_tester.InvalidTestCase{
			// ================================================================
			// Direct calls: all three prohibited names
			// ================================================================
			{
				Code: `alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `confirm(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `prompt(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// window.* dot access
			// ================================================================
			{
				Code: `window.alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `window.confirm(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `window.prompt(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// window['*'] bracket access
			// ================================================================
			{
				Code: `window['alert'](foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `window['confirm'](foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `window['prompt'](foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Template literal bracket access
			// ================================================================
			{
				Code: "window[`alert`](foo)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// alert shadowed locally but window.alert still flagged
			// ================================================================
			{
				Code: `function alert() {} window.alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			{
				Code: "var alert = function() {};\nwindow.alert(foo)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 1},
				},
			},
			{
				Code: `function foo(alert) { window.alert(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},

			// ================================================================
			// Direct call inside function/arrow/class (not shadowed)
			// ================================================================
			{
				Code: `function foo() { alert(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			{
				Code: `const foo = () => alert()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			{
				Code: `const foo = () => { alert(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A { foo() { alert(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			{
				Code: `class A { constructor() { alert(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},

			// ================================================================
			// Deeply nested (3+ levels)
			// ================================================================
			{
				Code: `function a() { function b() { function c() { alert(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 46},
				},
			},
			{
				Code: `const a = () => { const b = () => { const c = () => { alert(); }; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 55},
				},
			},

			// ================================================================
			// Shadowed inside function but not at call site
			// ================================================================
			{
				Code: "function foo() { var alert = function() {}; }\nalert();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 1},
				},
			},

			// ================================================================
			// this.alert at global scope
			// ================================================================
			{
				Code: `this.alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `this['alert'](foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// window shadowed inside function but not at outer call site
			// ================================================================
			{
				Code: "function foo() { var window = bar; window.alert(); }\nwindow.alert();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 1},
				},
			},

			// ================================================================
			// globalThis access
			// ================================================================
			{
				Code: `globalThis['alert'](foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `globalThis.alert()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: "function foo() { var globalThis = bar; globalThis.alert(); }\nglobalThis.alert();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 1},
				},
			},

			// ================================================================
			// Optional chaining
			// ================================================================
			{
				Code: `window?.alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `(window?.alert)(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Parenthesized callee (SkipParentheses)
			// ================================================================
			{
				Code: `(alert)(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `((alert))(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// window/globalThis access inside nested functions (not shadowed)
			// ================================================================
			{
				Code: `function foo() { window.alert(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			{
				Code: `const foo = () => window.confirm()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			{
				Code: `class A { foo() { window.prompt(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},

			// ================================================================
			// Multiple errors in one file
			// ================================================================
			{
				Code: "alert();\nconfirm();\nprompt();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 2, Column: 1},
					{MessageId: "unexpected", Line: 3, Column: 1},
				},
			},

			// ================================================================
			// IIFE
			// ================================================================
			{
				Code: `(function() { alert(); })()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},

			// ================================================================
			// Parenthesized object expression: (window).alert()
			// ================================================================
			{
				Code: `(window).alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `(globalThis).alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `(this).alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Non-null assertion on object: window!.alert()
			// ================================================================
			{
				Code: `window!.alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Type assertion on object: (window as any).alert()
			// ================================================================
			{
				Code: `(window as any).alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Optional chaining on this/globalThis
			// ================================================================
			{
				Code: `this?.alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `globalThis?.confirm()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Direct call in class field / static block (alert is not shadowed)
			// ================================================================
			{
				Code: `class A { x = alert(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			{
				Code: `class A { static { alert(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},

			// ================================================================
			// window.alert in class field / static block (window not shadowed)
			// ================================================================
			{
				Code: `class A { x = window.alert(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			{
				Code: `class A { static { window.alert(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},

			// ================================================================
			// TypeScript outer expressions on callee
			// ================================================================
			// Non-null assertion on callee
			{
				Code: `alert!(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Type assertion (as) on callee
			{
				Code: `(alert as Function)(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Non-null assertion on member callee
			{
				Code: `(window.alert!)(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Type assertion on member callee
			{
				Code: `(window.alert as Function)(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
		},
	)
}
