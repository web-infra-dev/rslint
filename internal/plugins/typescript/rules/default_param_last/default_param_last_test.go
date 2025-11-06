package default_param_last

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDefaultParamLastRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DefaultParamLastRule, []rule_tester.ValidTestCase{
		// Valid: no parameters
		{Code: `function f() {}`},

		// Valid: only required parameters
		{Code: `function f(a: number) {}`},
		{Code: `function f(a: number, b: number) {}`},

		// Valid: default parameters at the end
		{Code: `function f(a = 0) {}`},
		{Code: `function f(a: number, b = 0) {}`},
		{Code: `function f(a: number, b: number, c = 0) {}`},

		// Valid: optional parameters at the end
		{Code: `function f(a: number, b?: number) {}`},
		{Code: `function f(a: number, b?: number, c?: number) {}`},

		// Valid: both optional and default at the end
		{Code: `function f(a: number, b?: number, c = 0) {}`},
		{Code: `function f(a: number, b = 0, c?: number) {}`},

		// Valid: rest parameter after defaults
		{Code: `function f(a: number, b = 0, ...c: number[]) {}`},

		// Valid: arrow functions
		{Code: `const f = (a: number, b = 0) => {}`},
		{Code: `const f = (a: number, b?: number) => {}`},

		// Valid: function expressions
		{Code: `const f = function(a: number, b = 0) {}`},
		{Code: `const f = function(a: number, b?: number) {}`},

		// Valid: methods
		{Code: `class A { method(a: number, b = 0) {} }`},
		{Code: `class A { method(a: number, b?: number) {} }`},

		// Valid: constructors
		{Code: `class A { constructor(a: number, b = 0) {} }`},
		{Code: `class A { constructor(a: number, b?: number) {} }`},

		// Valid: parameter properties
		{Code: `class A { constructor(public a: number, public b = 0) {} }`},
		{Code: `class A { constructor(private a: number, public b?: number) {} }`},

		// Valid: destructuring with defaults at the end
		{Code: `function f(a: number, { b } = { b: 0 }) {}`},
		{Code: `function f(a: number, [b] = [0]) {}`},

		// Valid: multiple defaults at the end
		{Code: `function f(a: number, b = 0, c = 1) {}`},
		{Code: `function f(a: number, b = 0, c = 1, d = 2) {}`},

		// Valid: all parameters have defaults
		{Code: `function f(a = 0, b = 1, c = 2) {}`},

		// Valid: all parameters are optional
		{Code: `function f(a?: number, b?: number, c?: number) {}`},
	}, []rule_tester.InvalidTestCase{
		// Invalid: default before required
		{
			Code: `function f(a = 0, b: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: optional before required
		{
			Code: `function f(a?: number, b: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: default in the middle
		{
			Code: `function f(a: number, b = 0, c: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: optional in the middle
		{
			Code: `function f(a: number, b?: number, c: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: multiple violations
		{
			Code: `function f(a = 0, b: number, c = 1, d: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: arrow function
		{
			Code: `const f = (a = 0, b: number) => {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: function expression
		{
			Code: `const f = function(a = 0, b: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: method
		{
			Code: `class A { method(a = 0, b: number) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: constructor
		{
			Code: `class A { constructor(a = 0, b: number) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: parameter properties
		{
			Code: `class A { constructor(public a = 0, private b: number) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: optional parameter property before required
		{
			Code: `class A { constructor(public a?: number, private b: number) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: destructuring with default before required
		{
			Code: `function f({ a } = { a: 0 }, b: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: array destructuring with default before required
		{
			Code: `function f([a] = [0], b: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: mixed defaults and optionals before required
		{
			Code: `function f(a = 0, b?: number, c: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: optional before required in arrow function
		{
			Code: `const f = (a?: number, b: number) => {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
			},
		},

		// Invalid: default after optional before required
		{
			Code: `function f(a?: number, b = 0, c: number) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldBeLast"},
				{MessageId: "shouldBeLast"},
			},
		},
	})
}
