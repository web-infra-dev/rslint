package no_commented_out_tests_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_commented_out_tests"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCommentedOutTests(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_commented_out_tests.NoCommentedOutTestsRule,
		[]rule_tester.ValidTestCase{
			{Code: `test("foo", () => {})`},
			{Code: `it.skip("foo", () => {})`},
			{Code: `describe.only("foo", () => {})`},
			{Code: `// fit("foo", () => {})`},
			{Code: `// xit("foo", () => {})`},
			{Code: `// xtest("foo", () => {})`},
			{Code: `// fdescribe("foo", () => {})`},
			{Code: `// xdescribe("foo", () => {})`},
			{Code: `// testSomething()`},
			{Code: `// latest(items)`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `// test("foo", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// it.skip("foo", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// describe.only("foo", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// test.for([{ value: 1 }])("$value", ({ value }) => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// describe['skip']("foo", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: "/*\n  describe(\"foo\", () => {})\n*/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
		},
	)
}
