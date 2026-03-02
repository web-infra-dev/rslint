package no_dupe_args

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeArgsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeArgsRule,
		[]rule_tester.ValidTestCase{
			{Code: `function foo(a, b, c) {}`},
			{Code: `var foo = function(a, b, c) {}`},
			{Code: `function foo(a, b, c, d) {}`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `function foo(a, b, a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code: `function foo(a, a, a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code: `var foo = function(a, b, b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26},
				},
			},
		},
	)
}
