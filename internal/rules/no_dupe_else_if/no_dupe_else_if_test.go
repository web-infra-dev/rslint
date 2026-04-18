package no_dupe_else_if

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeElseIfRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeElseIfRule,
		[]rule_tester.ValidTestCase{
			// different test conditions
			{Code: `if (a) {} else if (b) {}`},
			{Code: `if (a); else if (b); else if (c);`},
			{Code: `if (true) {} else if (false) {} else {}`},
			{Code: `if (1) {} else if (2) {}`},
			{Code: `if (f) {} else if (f()) {}`},
			{Code: `if (f(a)) {} else if (g(a)) {}`},
			{Code: `if (f(a)) {} else if (f(b)) {}`},
			{Code: `if (a === 1) {} else if (a === 2) {}`},
			{Code: `if (a === 1) {} else if (b === 1) {}`},

			// not an if-else-if chain
			{Code: `if (a) {}`},
			{Code: `if (a);`},
			{Code: `if (a) {} else {}`},
			{Code: `if (a) if (a) {}`},
			{Code: `if (a) if (a);`},
			{Code: `if (a) { if (a) {} }`},
			{Code: `if (a) {} else { if (a) {} }`},
			{Code: `if (a) {} if (a) {}`},
			{Code: `if (a); if (a);`},
			{Code: `while (a) if (a);`},
			{Code: `if (a); else a ? a : a;`},

			// not same conditions in the chain
			{Code: `if (a) { if (b) {} } else if (b) {}`},
			{Code: `if (a) if (b); else if (a);`},

			// not equal tokens
			{Code: `if (a) {} else if (!!a) {}`},
			{Code: `if (a === 1) {} else if (a === (1)) {}`},

			// more complex valid chains
			{Code: `if (a || b) {} else if (c || d) {}`},
			{Code: `if (a || b) {} else if (a || c) {}`},
			{Code: `if (a) {} else if (a || b) {}`},
			{Code: `if (a) {} else if (b) {} else if (a || b || c) {}`},
			{Code: `if (a && b) {} else if (a) {} else if (b) {}`},
			{Code: `if (a && b) {} else if (b && c) {} else if (a && c) {}`},
			{Code: `if (a && b) {} else if (b || c) {}`},
			{Code: `if (a) {} else if (b && (a || c)) {}`},
			{Code: `if (a) {} else if (b && (c || d && a)) {}`},
			{Code: `if (a && b && c) {} else if (a && b && (c || d)) {}`},
		},
		[]rule_tester.InvalidTestCase{
			// basic
			{
				Code: `if (a) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code: `if (a); else if (a);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			{
				Code: `if (a) {} else if (a) {} else {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (a) {} else if (c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (c) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 50},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},

			// multiple duplicates
			{
				Code: `if (a) {} else if (a) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (a) {} else if (b) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
					{MessageId: "unexpected", Line: 1, Column: 50},
					{MessageId: "unexpected", Line: 1, Column: 65},
				},
			},

			// inner if statements do not affect chain
			{
				Code: `if (a) { if (b) {} } else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},

			// various kinds of test conditions
			{
				Code: `if (a === 1) {} else if (a === 1) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26},
				},
			},
			{
				Code: `if (1 < a) {} else if (1 < a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			{
				Code: `if (true) {} else if (true) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			{
				Code: `if (a && b) {} else if (a && b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a && b || c)  {} else if (a && b || c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			{
				Code: `if (f(a)) {} else if (f(a)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},

			// spaces and comments do not affect comparison
			{
				Code: `if (a === 1) {} else if (a===1) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26},
				},
			},
			{
				Code: `if (a === 1) {} else if (a === /* comment */ 1) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26},
				},
			},

			// extra parens around the whole test condition
			{
				Code: `if (a === 1) {} else if ((a === 1)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},

			// more complex errors with || and &&
			{
				Code: `if (a || b) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a || b) {} else if (a) {} else if (b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
					{MessageId: "unexpected", Line: 1, Column: 40},
				},
			},
			{
				Code: `if (a || b) {} else if (b || a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (a || b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `if (a || b) {} else if (c || d) {} else if (a || d) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 45},
				},
			},
			{
				Code: `if ((a === b && fn(c)) || d) {} else if (fn(c) && a === b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 42},
				},
			},
			{
				Code: `if (a) {} else if (a && b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code: `if (a && b) {} else if (b && a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a && b) {} else if (a && b && c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a || c) {} else if (a && b || c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (c && a || b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (c && (a || b)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `if (a) {} else if (b && c) {} else if (d && (a || e && c && b)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 40},
				},
			},
			{
				Code: `if (a || b && c) {} else if (b && c && d) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 30},
				},
			},
			{
				Code: `if (a || b) {} else if (b && c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if ((a || b) && c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `if ((a && (b || c)) || d) {} else if ((c || b) && e && a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 39},
				},
			},
			{
				Code: `if (a && b || b && c) {} else if (a && b && c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `if (a) {} else if (b && c) {} else if (d && (c && e && b || a)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 40},
				},
			},
			{
				Code: `if (a || (b && (c || d))) {} else if ((d || c) && b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 39},
				},
			},
			{
				Code: `if (a || b) {} else if ((b || a) && c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a || b) {} else if (c) {} else if (d) {} else if (b && (a || c)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 55},
				},
			},
			{
				Code: `if (a || b || c) {} else if (a || (b && d) || (c && e)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 30},
				},
			},
			{
				Code: `if (a || (b || c)) {} else if (a || (b && c)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 32},
				},
			},
			{
				Code: `if (a || b) {} else if (c) {} else if (d) {} else if ((a || c) && (b || d)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 55},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (c && (a || d && b)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `if (a) {} else if (a || a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code: `if (a || a) {} else if (a || a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a || a) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a) {} else if (a && a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code: `if (a && a) {} else if (a && a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `if (a && a) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
		},
	)
}
