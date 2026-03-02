package no_duplicate_case

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDuplicateCaseRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDuplicateCaseRule,
		[]rule_tester.ValidTestCase{
			{Code: `switch (a) { case 1: break; case 2: break; }`},
			{Code: `switch (a) { case 1: break; case "1": break; }`},
			{Code: `switch (a) { case 1: break; default: break; }`},
			{Code: `switch (a) { case "a": break; case "b": break; }`},
			{Code: `switch (a) { case a: break; case b: break; }`},
			// String literals with comment-like content should not be corrupted
			{Code: `switch (a) { case "http://example.com": break; case "other": break; }`},
			{Code: `switch (a) { case "a /* b */": break; case "a": break; }`},
			// String literals with different whitespace should not be collapsed
			{Code: `switch (a) { case "hello  world": break; case "hello world": break; }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `switch (a) { case 1: break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},
			{
				Code: `switch (a) { case "a": break; case "a": break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			{
				Code: `switch (a) { case 1: break; case 2: break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 44},
				},
			},
			{
				Code: `switch (a) { case a: break; case a: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},
			// Comments outside strings should still be stripped for comparison
			{
				Code: `switch (a) { case /*a*/ 1: break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
		},
	)
}
