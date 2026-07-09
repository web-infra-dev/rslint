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

			// If-else with super in both branches followed by other statements
			// (regression: trailing statements must not be required to also call super())
			{Code: `class A extends B { constructor() { if (a) { super(); } else { super(); } b(); } }`},
			{Code: `class A extends B { constructor() { if (a) { c(); super(); d(); } else { super(); } e(); } }`},

			// Switch with super in all cases followed by other statements
			{Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; default: super(); break; } b(); } }`},

			// Always-truthy loops that call super() before the only exit
			{Code: `class A extends B { constructor() { while (true) { super(); break; } } }`},
			{Code: `class A extends B { constructor() { for (;;) { super(); break; } } }`},
			{Code: `class A extends B { constructor() { label: while (true) { super(); break label; } } }`},

			// do-while always runs its body at least once, regardless of the condition
			{Code: `class A extends B { constructor() { do { super(); } while (false); } }`},

			// try/finally: finally always runs, so super() there satisfies the
			// requirement regardless of what the try/catch do
			{Code: `class A extends B { constructor() { try { super(); } finally {} } }`},
			{Code: `class A extends B { constructor() { try { a(); } catch (e) { } finally { super(); } } }`},

			// try without catch: an exception before super() propagates out of
			// the constructor, which is an acceptable terminating path
			{Code: `class A extends B { constructor() { try { super(); } finally { b(); } } }`},
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

			// Loops that may exit without ever calling super()
			{
				Code: `class A extends B { constructor() { while (true) { if (a) break; super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { for (var x of y) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { for (var i = 0; i < 10; i++) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { while (a) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},

			// try/catch without finally: catch must independently call super() too
			{
				Code: `class A extends B { constructor() { try { super(); } catch (e) {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
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
