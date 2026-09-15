// TestPreferEqualityMatcherUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 prefer-equality-matcher suite
// (tests/prefer-equality-matcher.test.ts) 1:1. Position assertions cover
// line/column for every invalid case. Rstest-specific sources, edge shapes,
// branch lock-ins and edit-demand coverage live in
// prefer_equality_matcher_extras_test.go.
package prefer_equality_matcher

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func equalitySuggestions(output func(string) string) []rule_tester.InvalidTestCaseSuggestion {
	matchers := []string{"toBe", "toEqual", "toStrictEqual"}
	suggestions := make([]rule_tester.InvalidTestCaseSuggestion, len(matchers))
	for index, matcher := range matchers {
		suggestions[index] = rule_tester.InvalidTestCaseSuggestion{
			MessageId: "suggestEqualityMatcher",
			Output:    output(matcher),
		}
	}
	return suggestions
}

func equalityError(column int, output func(string) string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId:   "useEqualityMatcher",
		Message:     "Prefer using one of the equality matchers instead",
		Line:        1,
		Column:      column,
		Suggestions: equalitySuggestions(output),
	}
}

func TestPreferEqualityMatcherUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferEqualityMatcherRule,
		[]rule_tester.ValidTestCase{
			// ---- === ----
			{Code: `expect.hasAssertions`},
			{Code: `expect.hasAssertions()`},
			{Code: `expect.assertions(1)`},
			{Code: `expect(true).toBe(...true)`},
			{Code: `expect(a).to.be.a("string");`},
			{Code: `expect(a == 1).toBe(true)`},
			{Code: `expect(1 == a).toBe(true)`},
			{Code: `expect(a == b).toBe(true)`},
			// ---- !== ----
			{Code: `expect.hasAssertions`},
			{Code: `expect.hasAssertions()`},
			{Code: `expect.assertions(1)`},
			{Code: `expect(true).toBe(...true)`},
			{Code: `expect(a != 1).toBe(true)`},
			{Code: `expect(1 != a).toBe(true)`},
			{Code: `expect(a != b).toBe(true)`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- === ----
			{
				Code: `expect(a === b).toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(17, func(matcher string) string {
					return `expect(a).` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a === b,).toBe(true,);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(18, func(matcher string) string {
					return `expect(a,).` + matcher + `(b,);`
				})},
			},
			{
				Code: `expect(a === b).toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(17, func(matcher string) string {
					return `expect(a).not.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a === b).resolves.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(26, func(matcher string) string {
					return `expect(a).resolves.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a === b).resolves.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(26, func(matcher string) string {
					return `expect(a).resolves.not.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a === b).not.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(21, func(matcher string) string {
					return `expect(a).not.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a === b).not.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(21, func(matcher string) string {
					return `expect(a).` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a === b).resolves.not.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(30, func(matcher string) string {
					return `expect(a).resolves.not.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a === b).resolves.not.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(30, func(matcher string) string {
					return `expect(a).resolves.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a === b)["resolves"].not.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(33, func(matcher string) string {
					return `expect(a)["resolves"].` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a === b)["resolves"]["not"]["toBe"](false);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(36, func(matcher string) string {
					return `expect(a)["resolves"]` + `["` + matcher + `"](b);`
				})},
			},
			// ---- !== ----
			{
				Code: `expect(a !== b).toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(17, func(matcher string) string {
					return `expect(a).not.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a !== b).toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(17, func(matcher string) string {
					return `expect(a).` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a !== b).resolves.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(26, func(matcher string) string {
					return `expect(a).resolves.not.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a !== b).resolves.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(26, func(matcher string) string {
					return `expect(a).resolves.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a !== b).not.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(21, func(matcher string) string {
					return `expect(a).` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a !== b).not.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(21, func(matcher string) string {
					return `expect(a).not.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a !== b).resolves.not.toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(30, func(matcher string) string {
					return `expect(a).resolves.` + matcher + `(b);`
				})},
			},
			{
				Code: `expect(a !== b).resolves.not.toBe(false);`,
				Errors: []rule_tester.InvalidTestCaseError{equalityError(30, func(matcher string) string {
					return `expect(a).resolves.not.` + matcher + `(b);`
				})},
			},
		},
	)
}
