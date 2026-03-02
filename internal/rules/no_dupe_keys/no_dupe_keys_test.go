package no_dupe_keys

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeKeysRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeKeysRule,
		[]rule_tester.ValidTestCase{
			{Code: `var x = { a: 1, b: 2 };`},
			{Code: `var x = { a: 1, b: 2, c: 3 };`},
			{Code: `var x = { get a() {}, set a(v) {} };`},
			{Code: `var x = { [Symbol()]: 1, [Symbol()]: 2 };`},
			{Code: `var x = { "": 1, " ": 2 };`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var x = { a: 1, a: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			{
				Code: `var x = { a: 1, b: 2, a: 3 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			{
				Code: `var x = { a: 1, b: 2, a: 3, a: 4 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},
			{
				Code: `var x = { "a": 1, "a": 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			// Numeric literal equivalence: 0x1 and 1 are the same key
			{
				Code: `var x = { 0x1: "a", 1: "b" };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
		},
	)
}
