package no_cond_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCondAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCondAssignRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			// Default behavior (except-parens)
			{Code: `var x = 0; if (x == 0) { var b = 1; }`},
			{Code: `var x = 5; while (x < 5) { x = x + 1; }`},
			{Code: `x = 0;`},
			{Code: `var x; var b = (x === 0) ? 1 : 0;`},

			// With "except-parens" option - properly parenthesized assignments are allowed
			{Code: `if ((someNode = someNode.parentNode) !== null) { }`, Options: "except-parens"},
			{Code: `if ((a = b));`, Options: "except-parens"},
			{Code: `while ((a = b));`, Options: "except-parens"},
			{Code: `do {} while ((a = b));`, Options: "except-parens"},
			{Code: `for (;(a = b););`, Options: "except-parens"},
			{Code: `if (someNode || (someNode = parentNode)) { }`, Options: "except-parens"},
			{Code: `while (someNode || (someNode = parentNode)) { }`, Options: "except-parens"},
			{Code: `do { } while (someNode || (someNode = parentNode));`, Options: "except-parens"},
			{Code: `for (;someNode || (someNode = parentNode););`, Options: "except-parens"},

			// Arrow functions
			{Code: `if ((node => node = parentNode)(someNode)) { }`, Options: "except-parens"},
			{Code: `if ((function(node) { return node = parentNode; })(someNode)) { }`, Options: "except-parens"},

			// Switch statements - assignments in case clauses are not in test expressions
			{Code: `switch (foo) { case a = b: bar(); }`},
			{Code: `switch (foo) { case baz + (a = b): bar(); }`},

			// Assignments outside of conditionals
			{Code: `var x; x = 0;`},
			{Code: `var x = 1; x += 1;`},

			// Comparisons (not assignments)
			{Code: `if (x === 0) { }`},
			{Code: `while (x == 1) { }`},
			{Code: `do { } while (x != 0);`},
			{Code: `for (; x >= 0;) { }`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			// Missing parentheses (default "except-parens" mode)
			{
				Code: `var x; if (x = 0) { var b = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 12},
				},
			},
			{
				Code: `var x; while (x = 0) { var b = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},
			{
				Code: `var x = 0, y; do { y = x; } while (x = x + 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 36},
				},
			},
			{
				Code: `var x; for(; x+=1 ;){};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 14},
				},
			},
			{
				Code: `var x; if ((x) = (0));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 12},
				},
			},

			// With "always" option - all assignments in conditionals are forbidden
			{
				Code:    `if (x = 0) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5},
				},
			},
			{
				Code:    `while (x = 0) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
			{
				Code:    `do { } while (x = x + 1);`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			{
				Code:    `for(; x = y; ) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 7},
				},
			},
			{
				Code:    `if ((x = 0)) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			{
				Code:    `while ((x = 0)) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code:    `do { } while ((x = x + 1));`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			{
				Code:    `for(; (x = y); ) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
			{
				Code:    `if (someNode || (someNode = parentNode)) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			{
				Code:    `while (someNode || (someNode = parentNode)) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			{
				Code:    `do { } while (someNode || (someNode = parentNode));`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			{
				Code:    `for (; someNode || (someNode = parentNode); ) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},

			// Ternary conditionals
			{
				Code:    `var x = 0 ? (x = 1) : 2;`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
				},
			},

			// Compound assignment operators
			{
				Code: `if (x += 1) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 5},
				},
			},
			{
				Code: `while (x -= 1) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 8},
				},
			},
			{
				Code: `do { } while (x *= 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},
		},
	)
}
