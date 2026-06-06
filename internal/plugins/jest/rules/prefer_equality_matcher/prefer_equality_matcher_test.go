package prefer_equality_matcher_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_equality_matcher"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func expectSuggestions(output func(equalityMatcher string) string) []rule_tester.InvalidTestCaseSuggestion {
	matchers := []string{"toBe", "toEqual", "toStrictEqual"}
	suggestions := make([]rule_tester.InvalidTestCaseSuggestion, len(matchers))
	for i, matcher := range matchers {
		suggestions[i] = rule_tester.InvalidTestCaseSuggestion{
			MessageId: "suggestEqualityMatcher",
			Output:    output(matcher),
		}
	}
	return suggestions
}

func TestPreferEqualityMatcherRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_equality_matcher.PreferEqualityMatcherRule,
		[]rule_tester.ValidTestCase{
			// ===
			{Code: `expect.hasAssertions`},
			{Code: `expect.hasAssertions()`},
			{Code: `expect.assertions(1)`},
			{Code: `expect(a == 1).toBe(true)`},
			{Code: `expect(1 == a).toBe(true)`},
			{Code: `expect(a == b).toBe(true)`},
			{Code: `expect((a === b)).toBe(true)`},
			{Code: `expect((a === b) as boolean).toBe(true)`},
			{Code: `expect(a === b).toBe((true))`},
			{Code: `expect(a === b).toBe(true as boolean)`},
			// !==
			{Code: `expect(a != 1).toBe(true)`},
			{Code: `expect(1 != a).toBe(true)`},
			{Code: `expect(a != b).toBe(true)`},
			{Code: `expect((a !== b)).toBe(true)`},
		},
		[]rule_tester.InvalidTestCase{
			// ===
			{
				Code: `expect(a === b).toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    17,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a === b,).toBe(true,);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    18,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a,).` + equalityMatcher + `(b,);`
						}),
					},
				},
			},
			{
				Code: `expect(a === b).toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    17,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).not.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a === b).resolves.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    26,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).resolves.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a === b).resolves.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    26,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).resolves.not.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a === b).not.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    21,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).not.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a === b).not.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    21,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a === b).resolves.not.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    30,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).resolves.not.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a === b).resolves.not.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    30,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).resolves.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a === b)["resolves"].not.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    33,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).resolves.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a === b)["resolves"]["not"]["toBe"](false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    36,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).resolves.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			// !==
			{
				Code: `expect(a !== b).toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    17,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).not.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a !== b).toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    17,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a !== b).resolves.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    26,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).resolves.not.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a !== b).resolves.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    26,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).resolves.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a !== b).not.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    21,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a !== b).not.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    21,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).not.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a !== b).resolves.not.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    30,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).resolves.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
			{
				Code: `expect(a !== b).resolves.not.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useEqualityMatcher",
						Line:      1,
						Column:    30,
						Suggestions: expectSuggestions(func(equalityMatcher string) string {
							return `expect(a).resolves.not.` + equalityMatcher + `(b);`
						}),
					},
				},
			},
		},
	)
}
