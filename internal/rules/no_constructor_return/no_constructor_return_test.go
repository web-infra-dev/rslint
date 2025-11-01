package no_constructor_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConstructorReturnRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConstructorReturnRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			// Regular functions with return statements are allowed
			{Code: `function fn() { return }`},
			{Code: `function fn(kumiko) { if (kumiko) { return kumiko } }`},
			{Code: `const fn = function () { return }`},
			{Code: `const fn = function () { if (kumiko) { return kumiko } }`},
			{Code: `const fn = () => { return }`},
			{Code: `const fn = () => { if (kumiko) { return kumiko } }`},

			// Classes without constructors or with empty constructors
			{Code: `class C {  }`},
			{Code: `class C { constructor() {} }`},
			{Code: `class C { constructor() { let v } }`},

			// Methods and getters can return values
			{Code: `class C { method() { return '' } }`},
			{Code: `class C { get value() { return '' } }`},

			// Constructors with bare return statements (no value) are allowed
			{Code: `class C { constructor(a) { if (!a) { return } else { a() } } }`},
			{Code: `class C { constructor() { return } }`},
			{Code: `class C { constructor() { { return } } }`},

			// Nested functions inside constructors can return values
			{Code: `class C { constructor() { function fn() { return true } } }`},
			{Code: `class C { constructor() { this.fn = function () { return true } } }`},
			{Code: `class C { constructor() { this.fn = () => { return true } } }`},

			// TypeScript: classes with multiple constructors
			{Code: `class C { constructor(); constructor(a?: string) {} }`},
			{Code: `class C { constructor(a: string); constructor(a?: string) { if (!a) { return } } }`},

			// TypeScript: constructors with type annotations
			{Code: `class C { constructor(private x: number) {} }`},
			{Code: `class C { constructor(public readonly y: string) { return } }`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			// Constructor returning a value
			{
				Code: `class C { constructor() { return '' } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    27,
					},
				},
			},
			// Constructor with conditional return of a value
			{
				Code: `class C { constructor(a) { if (!a) { return '' } else { a() } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    38,
					},
				},
			},

			// Additional test cases
			// Constructor returning a number
			{
				Code: `class C { constructor() { return 1 } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    27,
					},
				},
			},
			// Constructor returning an object
			{
				Code: `class C { constructor() { return {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    27,
					},
				},
			},
			// Constructor returning null
			{
				Code: `class C { constructor() { return null } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    27,
					},
				},
			},
			// Constructor returning undefined explicitly
			{
				Code: `class C { constructor() { return undefined } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    27,
					},
				},
			},
			// Constructor returning a boolean
			{
				Code: `class C { constructor() { return true } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    27,
					},
				},
			},
			// Constructor with multiple returns
			{
				Code: `class C { constructor(x) { if (x) return 'yes'; return 'no'; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    35,
					},
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    49,
					},
				},
			},
			// Constructor returning array
			{
				Code: `class C { constructor() { return [] } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    27,
					},
				},
			},
			// Constructor returning function call result
			{
				Code: `class C { constructor() { return foo() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    27,
					},
				},
			},
			// Constructor returning 'this' (which is the normal behavior but explicit return is flagged)
			{
				Code: `class C { constructor() { return this } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    27,
					},
				},
			},

			// TypeScript: constructor with type annotations still can't return values
			{
				Code: `class C { constructor(private x: number) { return 1 } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    44,
					},
				},
			},
		},
	)
}
