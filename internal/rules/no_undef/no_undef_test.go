package no_undef

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUndefRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUndefRule,
		[]rule_tester.ValidTestCase{
			// Variable declarations
			{Code: `var a = 1; a;`},
			{Code: `let b = 2; b;`},
			{Code: `const c = 3; c;`},

			// Function declarations
			{Code: `function f() { } f();`},

			// Class declarations
			{Code: `class MyClass {} new MyClass();`},

			// Parameters
			{Code: `function f(x: number) { return x; }`},

			// typeof of undeclared is OK by default
			{Code: `typeof a`},
			{Code: `typeof a === 'string'`},

			// Property access - the property name should not be flagged
			{Code: `var obj = { x: 1 }; obj.x;`},

			// Object literal keys should not be flagged
			{Code: `var obj = { key: 1 };`},

			// Labels should not be flagged
			{Code: `loop: for (var i = 0; i < 10; i++) { break loop; }`},

			// Import (TypeScript resolves these)
			// Built-in globals available via lib
			{Code: `console.log("test");`},
			{Code: `var p = new Promise<void>((resolve) => resolve());`},
			{Code: `setTimeout(() => {}, 100);`},

			// Type annotations should not trigger (type-only position)
			{Code: `type MyType = string; var x: MyType;`},
			{Code: `interface MyInterface { x: number; } var y: MyInterface;`},

			// Destructuring
			{Code: `var { a, b } = { a: 1, b: 2 }; a; b;`},
			{Code: `var [x, y] = [1, 2]; x; y;`},

			// Assignment to declared variable
			{Code: `var a; a = 1;`},

			// Function expressions
			{Code: `var f = function() {}; f();`},

			// Arrow functions
			{Code: `var f = () => {}; f();`},

			// For loop variables
			{Code: `for (var i = 0; i < 10; i++) { i; }`},
			{Code: `for (let i = 0; i < 10; i++) { i; }`},

			// Catch clause variable
			{Code: `try {} catch (e) { e; }`},

			// Function parameter usage
			{Code: `function foo(x: number) { return x; }`},

			// Class with method usage
			{Code: `class Foo { bar() {} }; new Foo().bar();`},

			// Type reference
			{Code: `type X = string; var y: X;`},

			// Variable used with postfix operator
			{Code: `var x = 1; x++;`},

			// Multiple function parameters
			{Code: `function foo(a: number, b: string) { return a + b; }`},

			// Class with property and method using this
			{Code: `class Foo { bar = 1; baz() { return this.bar; } }`},

			// Enum declaration and usage
			{Code: `enum Direction { Up, Down }; Direction.Up;`},

			// /*global*/ comment declares a global variable
			{Code: `/*global myVar*/ myVar = 1;`},

			// /*global*/ comment with multiple names
			{Code: `/*global a, b*/ a = 1; b = 2;`},

			// /*global*/ comment with writable flag
			{Code: `/*global myVar:writable*/ myVar = 1;`},
		},
		[]rule_tester.InvalidTestCase{
			// Undeclared variable in assignment
			{
				Code: `a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},
			// Undeclared variable in initializer
			{
				Code: `var a = b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
				},
			},
			// typeof with checkTypeof: true
			{
				Code:    `typeof anUndefinedVar === 'string'`,
				Options: map[string]interface{}{"typeof": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 8},
				},
			},
			// Undeclared function call
			{
				Code: `undeclaredFunc();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},
			// Multiple undeclared variables
			{
				Code: `var x = foo + bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
					{MessageId: "undef", Line: 1, Column: 15},
				},
			},
			// Undeclared variable inside function
			{
				Code: `function foo() { return undeclaredVar123; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 25},
				},
			},
			// Undeclared function call in initializer
			{
				Code: `var x = unknownFunc123();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 9},
				},
			},
			// Undeclared variable in condition
			{
				Code: `if (unknownCondition123) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 5},
				},
			},
			// /*global*/ comment does not declare the used variable
			{
				Code: `/*global otherVar*/ unknownVar123 = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 21},
				},
			},
		},
	)
}
