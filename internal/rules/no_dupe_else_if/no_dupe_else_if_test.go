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
			{Code: `if (a) {} else if (b) {}`},
			{Code: `if (a) {} else if (b) {} else if (c) {}`},
			{Code: `if (a === 1) {} else if (a === 2) {}`},
			{Code: `if (a) {} else if (b) {} else {}`},
			{Code: `if (a) { if (a) {} }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `if (a) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			{
				Code: `if (a === 1) {} else if (a === 1) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
		},
	)
}
