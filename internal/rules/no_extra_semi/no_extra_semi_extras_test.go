// TestNoExtraSemiExtras locks in branches and edge shapes that the upstream test suite doesn't exercise.
// Each case carries an inline comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers.
package no_extra_semi

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtraSemiExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExtraSemiRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: Class expression inside variable declaration ----
			{Code: "const x = class { a() {} };"},
			{Code: "const x = class A { a() {} };"},

			// ---- Dimension 4: Correct TS abstract class and interfaces ----
			{Code: "abstract class A { abstract foo(): void; }"},
			{Code: "interface A { foo: string; }"},
			{Code: "type T = { foo: string; };"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: TS constructs with double semicolons ----
			{
				Code:   "(x as any);;",
				Output: []string{"(x as any);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code:   "x!;;",
				Output: []string{"x!;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 4},
				},
			},
			{
				Code:   "(x satisfies any);;",
				Output: []string{"(x satisfies any);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},

			// ---- Dimension 4: TS abstract class extra semicolons ----
			{
				Code:   "abstract class A { abstract foo();; }",
				Output: []string{"abstract class A { abstract foo(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},

			// ---- Real-user & Branch locks: Directive hazards ----
			{
				// Locks in: string literal directive checking branch when there's an existing prologue directive
				Code: "'use strict'; ; 'other directive'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			{
				// Locks in: double semicolon before a string literal that is not a directive
				Code:   "function foo() { ; ; 'bar'; }",
				Output: []string{"function foo() {  ; 'bar'; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				// Locks in: TSModuleBlock is a directive-prologue position, same as
				// Program and function bodies, so this must not be autofixed.
				Code: "namespace Foo { ; 'use strict'; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
		},
	)
}
