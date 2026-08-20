package no_negated_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoNegatedConditionUpstream migrates the full valid/invalid suite from
// upstream ESLint tests/lib/rules/no-negated-condition.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in the no_negated_condition_extras_test.go file.
func TestNoNegatedConditionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNegatedConditionRule,
		[]rule_tester.ValidTestCase{
			{Code: "if (a) {}"},
			{Code: "if (a) {} else {}"},
			{Code: "if (!a) {}"},
			{Code: "if (!a) {} else if (b) {}"},
			{Code: "if (!a) {} else if (b) {} else {}"},
			{Code: "if (a == b) {}"},
			{Code: "if (a == b) {} else {}"},
			{Code: "if (a != b) {}"},
			{Code: "if (a != b) {} else if (b) {}"},
			{Code: "if (a != b) {} else if (b) {} else {}"},
			{Code: "if (a !== b) {}"},
			{Code: "if (a === b) {} else {}"},
			{Code: "a ? b : c"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "if (!a) {;} else {;}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "if (a != b) {;} else {;}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: "if (a !== b) {;} else {;}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: "!a ? b : c",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code: "a != b ? c : d",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: "a !== b ? c : d",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
		},
	)
}
