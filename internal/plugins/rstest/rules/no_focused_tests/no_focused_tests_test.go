package no_focused_tests_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_focused_tests"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoFocusedTests(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_focused_tests.NoFocusedTestsRule,
		[]rule_tester.ValidTestCase{
			{Code: "describe()"},
			{Code: "it()"},
			{Code: "test()"},
			{Code: "describe.skip()"},
			{Code: "it.skip()"},
			{Code: "test.skip()"},
			{Code: "test.todo()"},
			{Code: "test.concurrent()"},
			{Code: "test.sequential()"},
			{Code: "test.fails()"},
			{Code: "test.each()()"},
			{Code: "test.each`table`()"},
			{Code: "test.for()()"},
			{Code: "describe.for()()"},
			{Code: "test.runIf(true)()"},
			{Code: "describe.skipIf(false)()"},
			{Code: "var appliedOnly = describe.only; appliedOnly.apply(describe)"},
			{Code: "var calledOnly = it.only; calledOnly.call(it)"},
			// Rstest has no fit/fdescribe aliases: these are plain unknown calls.
			{Code: "fit()"},
			{Code: "fdescribe()"},
			// Not imported from @rstest/core, so not a Rstest test call.
			{Code: "import { test } from 'node:test';\n\ntest.only()"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "describe.only()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "describe()",
							},
						},
					},
				},
			},
			{
				Code: "it.only()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    4,
						EndLine:   1,
						EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "it()",
							},
						},
					},
				},
			},
			{
				Code: "test.only()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    6,
						EndLine:   1,
						EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "test()",
							},
						},
					},
				},
			},
			{
				Code: `describe["only"]()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "describe()",
							},
						},
					},
				},
			},
			{
				Code: "test.concurrent.only()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    17,
						EndLine:   1,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "test.concurrent()",
							},
						},
					},
				},
			},
			{
				Code: "test.only.for()()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    6,
						EndLine:   1,
						EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "test.for()()",
							},
						},
					},
				},
			},
			{
				Code: "test.only.each`table`()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    6,
						EndLine:   1,
						EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "test.each`table`()",
							},
						},
					},
				},
			},
			{
				Code: "describe.only.each()()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "describe.each()()",
							},
						},
					},
				},
			},
			{
				Code: "test.runIf(true).only()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    18,
						EndLine:   1,
						EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "test.runIf(true)()",
							},
						},
					},
				},
			},
			{
				Code: "import { describe as describeThis } from '@rstest/core';\n\ndescribeThis.only()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      3,
						Column:    14,
						EndLine:   3,
						EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "import { describe as describeThis } from '@rstest/core';\n\ndescribeThis()",
							},
						},
					},
				},
			},
			{
				Code: "const { test } = require('@rstest/core');\n\ntest.only()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      3,
						Column:    6,
						EndLine:   3,
						EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "const { test } = require('@rstest/core');\n\ntest()",
							},
						},
					},
				},
			},
		},
	)
}
