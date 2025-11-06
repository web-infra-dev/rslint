package constructor_super

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConstructorSuperRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConstructorSuperRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Non-derived classes
			{Code: `class A { }`},
			{Code: `class A { constructor() { } }`},

			// Derived classes with proper super() calls
			{Code: `class A extends B { }`},
			{Code: `class A extends B { constructor() { super(); } }`},

			// Classes extending null without constructor
			{Code: `class A extends null { }`},

			// Derived classes with super() in all paths
			{Code: `class A extends B { constructor() { if (true) super(); else super(); } }`},
			{Code: `class A extends B { constructor() { a ? super() : super(); } }`},

			// Early return substitutes for super()
			{Code: `class A extends B { constructor() { if (true) return a; super(); } }`},
			{Code: `class A extends B { constructor() { if (true) return; super(); } }`},
			{Code: `class A extends B { constructor() { if (a) return; if (b) return; super(); } }`},

			// Switch statements with super in all branches
			{Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; default: super(); break; } } }`},

			// Complex expressions as base class
			{Code: `class A extends (B = C) { constructor() { super(); } }`},
			{Code: `class A extends (B || C) { constructor() { super(); } }`},
			{Code: `class A extends (a ? B : C) { constructor() { super(); } }`},

			// Nested classes (outer has super)
			{Code: `class A extends B { constructor() { super(); class C extends D { constructor() { super(); } } } }`},
			{Code: `class A extends B { constructor() { super(); class C { } } }`},

			// No super needed - no extends
			{Code: `class A { constructor() { class B extends C { constructor() { super(); } } } }`},

			// Throw statements
			{Code: `class A extends B { constructor() { throw new Error(); } }`},

			// Multiple branches with returns
			{Code: `class A extends B { constructor() { if (a) { return; } else { return; } } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Non-derived classes calling super()
			{
				Code: `class A { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 27},
				},
			},

			// Classes extending null calling super()
			{
				Code: `class A extends null { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 40},
				},
			},

			// Classes extending literals calling super()
			{
				Code: `class A extends 100 { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 39},
				},
			},
			{
				Code: `class A extends "test" { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 42},
				},
			},

			// Derived classes missing super() entirely
			{
				Code: `class A extends B { constructor() { } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingAll", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { a; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingAll", Line: 1, Column: 21},
				},
			},

			// Nested classes where outer lacks super()
			{
				Code: `class A extends B { constructor() { class C extends D { constructor() { super(); } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingAll", Line: 1, Column: 21},
				},
			},

			// Missing super() in some code paths
			{
				Code: `class A extends B { constructor() { if (a) super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { a && super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { if (a) return; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; case 1: break; default: super(); break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},

			// Duplicate super() calls
			{
				Code: `class A extends B { constructor() { super(); super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 46},
				},
			},
			{
				Code: `class A extends B { constructor() { super(); if (a) super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 53},
				},
			},
			{
				Code: `class A extends B { constructor() { if (a) super(); else super(); super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 67},
				},
			},

			// Invalid extends expressions with super() call
			{
				Code: `class A extends (B = 5) { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 43},
				},
			},
			{
				Code: `class A extends (B += C) { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 44},
				},
			},
			{
				Code: `class A extends (B -= C) { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 44},
				},
			},
			{
				Code: `class A extends (B *= C) { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 44},
				},
			},
			{
				Code: `class A extends (B /= C) { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 44},
				},
			},

			// If-else without all paths having super
			{
				Code: `class A extends B { constructor() { if (a) { super(); } else { } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { if (a) { } else { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},

			// Ternary expression without super in one branch
			{
				Code: `class A extends B { constructor() { if (a) super(); else if (b) super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},

			// Switch without default
			{
				Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; case 1: super(); break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},

			// Multiple super calls in switch
			{
				Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; default: super(); break; } super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 101},
				},
			},

			// Combination of errors
			{
				Code: `class A extends B { constructor() { super(); } } class C extends null { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 89},
				},
			},
		},
	)
}
