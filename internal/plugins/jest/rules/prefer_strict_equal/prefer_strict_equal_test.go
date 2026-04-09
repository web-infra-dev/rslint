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
				Code: `expect(something)["toEqual"](somethingElse);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useToStrictEqual",
						Line:      1,
						Column:    19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestReplaceWithStrictEqual",
								Output:    `expect(something)[toStrictEqual](somethingElse);`,
							},
						},
					},
				},
			},
		},
	)
}
