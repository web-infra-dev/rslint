// TestPreferStrictEqualExtras preserves rslint's accessor and trivia behavior
// beyond the upstream Jest suite. Shared edit-demand behavior is covered by
// internal/utils/test_framework/rules/prefer_strict_equal.
package prefer_strict_equal_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_strict_equal"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStrictEqualExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_strict_equal.PreferStrictEqualRule,
		[]rule_tester.ValidTestCase{
			// The parsed matcher is the first invoked member. A method on the
			// assertion result is not another matcher in the same expect call.
			{Code: `expect(value).customMatcher().toEqual(expected);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(value).toEqual(expected).toEqual(other);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useToStrictEqual", Line: 1, Column: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: `expect(value).toStrictEqual(expected).toEqual(other);`}},
					},
					{
						MessageId: "useToStrictEqual", Line: 1, Column: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: `expect(value).toStrictEqual(expected).toEqual(other);`}},
					},
				},
			},
			{
				Code:   "expect(something)[`toEqual`](somethingElse);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToStrictEqual", Line: 1, Column: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: "expect(something)[`toStrictEqual`](somethingElse);"}}}},
			},
			{
				Code:   "expect(something).\n  toEqual(somethingElse);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToStrictEqual", Line: 2, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: "expect(something).\n  toStrictEqual(somethingElse);"}}}},
			},
			{
				Code:   "expect(something). /* c */ toEqual(somethingElse);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToStrictEqual", Line: 1, Column: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: "expect(something). /* c */ toStrictEqual(somethingElse);"}}}},
			},
			{
				Code:   "const toEqual = 'toEqual';\nexpect(something)[toEqual](somethingElse);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToStrictEqual", Line: 2, Column: 19}},
			},
			{
				Code:   "expect(something)[/* c */ 'toEqual'](somethingElse);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToStrictEqual", Line: 1, Column: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: "expect(something)[/* c */ 'toStrictEqual'](somethingElse);"}}}},
			},
		},
	)
}
