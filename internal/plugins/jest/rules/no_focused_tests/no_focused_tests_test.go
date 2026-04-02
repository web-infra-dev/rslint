package no_focused_tests_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_focused_tests"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoFocusedTests(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_focused_tests.NoFocusedTestsRule,
		[]rule_tester.ValidTestCase{
			{Code: "import { describe as fdescribe } from '@jest/globals';\n\nfdescribe()"},
			{Code: "describe()"},
			{Code: "it()"},
			{Code: "describe.skip()"},
			{Code: "it.skip()"},
			{Code: "test()"},
			{Code: "test.skip()"},
			{Code: "var appliedOnly = describe.only; appliedOnly.apply(describe)"},
			{Code: "var calledOnly = it.only; calledOnly.call(it)"},
			{Code: "it.each()()"},
			{Code: "it.each`table`()"},
			{Code: "test.each()()"},
			{Code: "test.each`table`()"},
			{Code: "test.concurrent()"},
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

			// TODO: Fix this test case
			// Currently, rslint does not support pragma configuration
			// {
			// 	Code: "context.only()",
			// 	Errors: []rule_tester.InvalidTestCaseError{
			// 		{
			// 			MessageId: "focusedTest",
			// 			Line:      1,
			// 			Column:    9,
			// 			EndLine:   1,
			// 			EndColumn: 13,
			// 			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
			// 				{
			// 					MessageId: "suggestRemoveFocus",
			// 					Output:    "context()",
			// 				},
			// 			},
			// 		},
			// 	},
			// },

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
				Code: "describe.only.each`table`()",
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
								Output:    "describe.each`table`()",
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
				Code: "it.concurrent.only.each``()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "it.concurrent.each``()",
							},
						},
					},
				},
			},
			{
				Code: "it.only.each()()",
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
								Output:    "it.each()()",
							},
						},
					},
				},
			},
			{
				Code: "it.only.each`table`()",
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
								Output:    "it.each`table`()",
							},
						},
					},
				},
			},
			{
				Code: `it["only"]()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    4,
						EndLine:   1,
						EndColumn: 10,
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
				Code: "test.concurrent.only.each()()",
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
								Output:    "test.concurrent.each()()",
							},
						},
					},
				},
			},
			{
				Code: "test.only.each()()",
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
								Output:    "test.each()()",
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
				Code: `test["only"]()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    6,
						EndLine:   1,
						EndColumn: 12,
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
				Code: "fdescribe()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 10,
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
				Code: "fit()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 4,
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
			Code: "fit.each()()",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "focusedTest",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "it.each()()",
							},
						},
					},
				},
			},
			{
			Code: "fit.each`table`()",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "focusedTest",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "it.each`table`()",
							},
						},
					},
				},
			},
			{
				Code: "const { describe } = require('@jest/globals');\n\ndescribe.only()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      3,
						Column:    10,
						EndLine:   3,
						EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "const { describe } = require('@jest/globals');\n\ndescribe()",
							},
						},
					},
				},
			},
			{
				Code: "import { describe as describeThis } from '@jest/globals';\n\ndescribeThis.only()",
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
								Output:    "import { describe as describeThis } from '@jest/globals';\n\ndescribeThis()",
							},
						},
					},
				},
			},
			{
				Code: "const { fdescribe } = require('@jest/globals');\n\nfdescribe()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      3,
						Column:    1,
						EndLine:   3,
						EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "const { fdescribe } = require('@jest/globals');\n\ndescribe()",
							},
						},
					},
				},
			},
			{
				Code: "import { fdescribe as describeJustThis } from '@jest/globals';\n\ndescribeJustThis()\ndescribeJustThis.each()()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      3,
						Column:    1,
						EndLine:   3,
						EndColumn: 17,
					},
					{
						MessageId: "focusedTest",
						Line:      4,
						Column:    1,
						EndLine:   4,
						EndColumn: 24,
					},
				},
			},
			{
				Code: "import { describe as context } from '@jest/globals';\n\ncontext.only.each()()",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusedTest",
						Line:      3,
						Column:    9,
						EndLine:   3,
						EndColumn: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveFocus",
								Output:    "import { describe as context } from '@jest/globals';\n\ncontext.each()()",
							},
						},
					},
				},
			},
		},
	)
}
