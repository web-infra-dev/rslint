package no_this_before_super

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoThisBeforeSuperRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoThisBeforeSuperRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Non-derived class - this is fine
			{Code: `class A { constructor() { this.b = 0; } }`},
			// Class with no constructor
			{Code: `class A extends B { }`},
			// Class extends null - no super needed
			{Code: `class A extends null { constructor() { } }`},
			// super() before this
			{Code: `class A extends B { constructor() { super(); this.c = 0; } }`},
			// super() before super.c()
			{Code: `class A extends B { constructor() { super(); super.c(); } }`},
			// super() in both branches of if/else, then this
			{Code: `class A extends B { constructor() { if (true) { super(); } else { super(); } this.c(); } }`},
			// Nested class has its own scope
			{Code: `class A extends B { constructor() { class C extends D { constructor() { super(); this.d = 0; } } super(); } }`},
			// this in nested function is fine
			{Code: `class A extends B { constructor() { function c() { this.d(); } super(); } }`},
			// this in nested arrow function (arrow functions capture this from enclosing scope,
			// but ESLint doesn't flag this case since the arrow isn't executed immediately)
			{Code: `class A extends B { constructor() { var c = () => this.d; super(); } }`},
			// super() in ternary on both branches
			{Code: `class A extends B { constructor() { a ? super() : super(); this.c = 0; } }`},
			// super() before this in nested expression
			{Code: `class A extends B { constructor() { super(); this.a = [this.b, this.c]; } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// this before super (no super at all)
			{
				Code: `class A extends B { constructor() { this.c = 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisBeforeSuper", Line: 1, Column: 37},
				},
			},
			// this before super (super comes later)
			{
				Code: `class A extends B { constructor() { this.c = 0; super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisBeforeSuper", Line: 1, Column: 37},
				},
			},
			// super.c() before super()
			{
				Code: `class A extends B { constructor() { super.c(); super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "superBeforeSuper", Line: 1, Column: 37},
				},
			},
			// this in super() arguments
			{
				Code: `class A extends B { constructor() { super(this.c); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisBeforeSuper", Line: 1, Column: 43},
				},
			},
			// super() only in if (no else), this after
			{
				Code: `class A extends B { constructor() { if (a) { super(); } this.c = 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisBeforeSuper", Line: 1, Column: 57},
				},
			},
			// this in if condition, super in body
			{
				Code: `class A extends B { constructor() { if (this.a) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisBeforeSuper", Line: 1, Column: 41},
				},
			},
		},
	)
}
