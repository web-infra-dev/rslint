package no_unmodified_loop_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnmodifiedLoopConditionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnmodifiedLoopConditionRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Modified in loop body
			{Code: `var foo = 0; while (foo) { ++foo; }`},
			{Code: `var foo = 0; while (foo) { foo += 1; }`},
			{Code: `var foo = 0; while (foo) { foo = bar(); }`},
			{Code: `var foo = 0; while (foo < 10) { foo++; }`},
			{Code: `var foo = 0; while (foo < 10) { --foo; }`},
			{Code: `var foo = 0; while (foo < 10) { foo -= 1; }`},

			// Modified in do-while body
			{Code: `var foo = 0; do { foo++; } while (foo < 10)`},
			{Code: `var foo = 0; do { foo = next(); } while (foo)`},

			// Modified in for incrementor
			{Code: `for (var i = 0; i < 10; i++) {}`},
			{Code: `for (var i = 0; i < 10; ++i) {}`},
			{Code: `for (var i = 10; i > 0; i--) {}`},
			{Code: `for (var i = 0; i < 10; i += 1) {}`},

			// Modified in for body
			{Code: `for (var i = 0; i < 10; ) { i++; }`},

			// Dynamic expressions in condition - skip
			{Code: `while (ok(foo)) { }`},
			{Code: `while (foo.ok) { }`},
			{Code: `while (foo[0]) { }`},
			{Code: `while (new Foo()) { }`},
			{Code: `while (tag` + "`template`" + `) { }`},
			{Code: `while (a.b.c) { }`},
			{Code: `for (var i = 0; f(i) < 10; ) { }`},

			// Multiple variables, all modified
			{Code: `var a = 0, b = 0; while (a < b) { a++; b--; }`},

			// Boolean literal conditions (no identifiers)
			{Code: `while (true) { break; }`},
			{Code: `do { break; } while (true)`},

			// Empty condition for for loop
			{Code: `for (;;) { break; }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// While statement - identifier not modified
			{
				Code: `var foo = 0; while (foo) { } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 21},
				},
			},
			{
				Code: `var foo = 0; while (foo < 10) { } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 21},
				},
			},

			// Do-while statement - identifier not modified
			{
				Code: `var foo = 0; do { } while (foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 28},
				},
			},

			// For statement - no incrementor, not modified in body
			{
				Code: `for (var i = 0; i < 10; ) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 17},
				},
			},

			// Multiple identifiers - one not modified
			{
				Code: `var a = 0, b = 0; while (a < b) { a++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 26},
				},
			},

			// Modified outside loop but not inside
			{
				Code: `var foo = 0; while (foo) { } foo++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 21},
				},
			},

			// Modified in nested function (should not count)
			{
				Code: `var foo = 0; while (foo) { function f() { foo = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 21},
				},
			},
			{
				Code: `var foo = 0; while (foo) { var f = () => { foo = 1; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 21},
				},
			},
		},
	)
}
