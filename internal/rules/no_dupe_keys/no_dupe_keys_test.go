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
			// __proto__ as proto setter is allowed to appear multiple times
			{Code: `var x = { __proto__: foo, __proto__: bar };`},
			{Code: `var x = { "__proto__": foo, "__proto__": bar };`},
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
			// Computed __proto__ is a regular property, not a proto setter
			{
				Code: `var x = { ["__proto__"]: 1, ["__proto__"]: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},
			// Numeric literal equivalence: 0x1 and 1 are the same key
			{
				Code: `var x = { 0x1: "a", 1: "b" };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			// BigInt literal equivalence: 0x1n and 1n normalize to the same key
			{
				Code: `var x = { [0x1n]: "a", [1n]: "b" };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			// Template literal computed property
			{
				Code: "var x = { [`key`]: 1, [`key`]: 2 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			// Numeric overflow to Infinity: 1e309 and 1e999 both normalize to "Infinity"
			{
				Code: `var x = { [1e309]: "a", [1e999]: "b" };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
		},
	)
}
