package prefer_strict_equal_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_strict_equal"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStrictEqualRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_strict_equal.PreferStrictEqualRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(something).toStrictEqual(somethingElse);`},
			{Code: `a().toEqual('b')`},
			{Code: `expect(a);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(something).toEqual(somethingElse);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useToStrictEqual",
						Line:      1,
						Column:    19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestReplaceWithStrictEqual",
								Output:    `expect(something).toStrictEqual(somethingElse);`,
							},
						},
					},
				},
			},
			{
				Code: `expect(something).toEqual(somethingElse,);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useToStrictEqual",
						Line:      1,
						Column:    19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestReplaceWithStrictEqual",
								Output:    `expect(something).toStrictEqual(somethingElse,);`,
							},
						},
					},
				},
			},
			{
				// The suggestion keeps the double quotes. It used to rewrite
				// them to single quotes, which introduced a `quotes` violation
				// in a double-quoted codebase and disagreed with
				// jest/no-alias-methods, which has always preserved them.
				Code: `expect(something)["toEqual"](somethingElse);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useToStrictEqual",
						Line:      1,
						Column:    19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestReplaceWithStrictEqual",
								Output:    `expect(something)["toStrictEqual"](somethingElse);`,
							},
						},
					},
				},
			},
			{
				Code: "expect(something)[`toEqual`](somethingElse);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useToStrictEqual",
						Line:      1,
						Column:    19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestReplaceWithStrictEqual",
								Output:    "expect(something)[`toStrictEqual`](somethingElse);",
							},
						},
					},
				},
			},
			// Leading-trivia cases: the suggestion must not absorb whatever
			// separates the dot from the matcher name.
			{
				Code: "expect(something).\n  toEqual(somethingElse);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useToStrictEqual",
						Line:      2,
						Column:    3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestReplaceWithStrictEqual",
								Output:    "expect(something).\n  toStrictEqual(somethingElse);",
							},
						},
					},
				},
			},
			{
				Code: "expect(something). /* c */ toEqual(somethingElse);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useToStrictEqual",
						Line:      1,
						Column:    28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestReplaceWithStrictEqual",
								Output:    "expect(something). /* c */ toStrictEqual(somethingElse);",
							},
						},
					},
				},
			},
			{
				Code: "expect(something)[/* c */ 'toEqual'](somethingElse);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useToStrictEqual",
						Line:      1,
						Column:    27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestReplaceWithStrictEqual",
								Output:    "expect(something)[/* c */ 'toStrictEqual'](somethingElse);",
							},
						},
					},
				},
			},
		},
	)
}
