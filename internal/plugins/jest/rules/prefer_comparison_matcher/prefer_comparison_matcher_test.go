package prefer_comparison_matcher_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_comparison_matcher"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func comparisonColumn(base int, operator string) int {
	return base + len(operator)
}

func generateInvalidCases(
	operator string,
	equalityMatcher string,
	preferredMatcher string,
	preferredMatcherWhenNegated string,
) []rule_tester.InvalidTestCase {
	useToBeComparison := func(column int) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{
			{MessageId: "useToBeComparison", Line: 1, Column: comparisonColumn(column, operator)},
		}
	}

	return []rule_tester.InvalidTestCase{
		{
			Code:   fmt.Sprintf(`expect(value %s 1).%s(true);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).%s(1);`, preferredMatcher)},
			Errors: useToBeComparison(18),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1,).%s(true,);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value,).%s(1,);`, preferredMatcher)},
			Errors: useToBeComparison(19),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1)['%s'](true);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).%s(1);`, preferredMatcher)},
			Errors: useToBeComparison(18),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1).resolves.%s(true);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).resolves.%s(1);`, preferredMatcher)},
			Errors: useToBeComparison(27),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1).%s(false);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).%s(1);`, preferredMatcherWhenNegated)},
			Errors: useToBeComparison(18),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1)['%s'](false);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).%s(1);`, preferredMatcherWhenNegated)},
			Errors: useToBeComparison(18),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1).resolves.%s(false);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).resolves.%s(1);`, preferredMatcherWhenNegated)},
			Errors: useToBeComparison(27),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1).not.%s(true);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).%s(1);`, preferredMatcherWhenNegated)},
			Errors: useToBeComparison(22),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1)['not'].%s(true);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).%s(1);`, preferredMatcherWhenNegated)},
			Errors: useToBeComparison(25),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1).resolves.not.%s(true);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).resolves.%s(1);`, preferredMatcherWhenNegated)},
			Errors: useToBeComparison(31),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1).not.%s(false);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).%s(1);`, preferredMatcher)},
			Errors: useToBeComparison(22),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1).resolves.not.%s(false);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).resolves.%s(1);`, preferredMatcher)},
			Errors: useToBeComparison(31),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1)["resolves"].not.%s(false);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).resolves.%s(1);`, preferredMatcher)},
			Errors: useToBeComparison(34),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1)["resolves"]["not"].%s(false);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).resolves.%s(1);`, preferredMatcher)},
			Errors: useToBeComparison(37),
		},
		{
			Code:   fmt.Sprintf(`expect(value %s 1)["resolves"]["not"]['%s'](false);`, operator, equalityMatcher),
			Output: []string{fmt.Sprintf(`expect(value).resolves.%s(1);`, preferredMatcher)},
			Errors: useToBeComparison(37),
		},
	}
}

func generateValidStringLiteralCases(operator string, matcher string) []rule_tester.ValidTestCase {
	pairs := [][2]string{
		{"x", "'y'"},
		{"x", "`y`"},
		{"x", "`y${z}`"},
	}

	var cases []rule_tester.ValidTestCase
	for _, pair := range pairs {
		a, b := pair[0], pair[1]
		codes := []string{
			fmt.Sprintf(`expect(%s %s %s).%s(true)`, a, operator, b, matcher),
			fmt.Sprintf(`expect(%s %s %s).%s(false)`, a, operator, b, matcher),
			fmt.Sprintf(`expect(%s %s %s).not.%s(true)`, a, operator, b, matcher),
			fmt.Sprintf(`expect(%s %s %s).not.%s(false)`, a, operator, b, matcher),
			fmt.Sprintf(`expect(%s %s %s).resolves.%s(true)`, a, operator, b, matcher),
			fmt.Sprintf(`expect(%s %s %s).resolves.%s(false)`, a, operator, b, matcher),
			fmt.Sprintf(`expect(%s %s %s).resolves.not.%s(true)`, a, operator, b, matcher),
			fmt.Sprintf(`expect(%s %s %s).resolves.not.%s(false)`, a, operator, b, matcher),
			fmt.Sprintf(`expect(%s %s %s).resolves.not.%s(false)`, b, operator, a, matcher),
			fmt.Sprintf(`expect(%s %s %s).resolves.not.%s(true)`, b, operator, a, matcher),
			fmt.Sprintf(`expect(%s %s %s).resolves.%s(false)`, b, operator, a, matcher),
			fmt.Sprintf(`expect(%s %s %s).resolves.%s(true)`, b, operator, a, matcher),
			fmt.Sprintf(`expect(%s %s %s).not.%s(false)`, b, operator, a, matcher),
			fmt.Sprintf(`expect(%s %s %s).not.%s(true)`, b, operator, a, matcher),
			fmt.Sprintf(`expect(%s %s %s).%s(false)`, b, operator, a, matcher),
			fmt.Sprintf(`expect(%s %s %s).%s(true)`, b, operator, a, matcher),
		}
		for _, code := range codes {
			cases = append(cases, rule_tester.ValidTestCase{Code: code})
		}
	}
	return cases
}

func comparisonOperatorCases(
	operator string,
	preferredMatcher string,
	preferredMatcherWhenNegated string,
) (valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	valid = []rule_tester.ValidTestCase{
		{Code: `expect()`},
		{Code: `expect({}).toStrictEqual({})`},
		{Code: fmt.Sprintf(`expect(value).%s(1);`, preferredMatcher)},
		{Code: fmt.Sprintf(`expect(value).%s(1);`, preferredMatcherWhenNegated)},
		{Code: fmt.Sprintf(`expect(value).not.%s(1);`, preferredMatcher)},
		{Code: fmt.Sprintf(`expect(value).not.%s(1);`, preferredMatcherWhenNegated)},
	}

	equalityMatchers := []string{"toBe", "toEqual", "toStrictEqual"}
	for _, equalityMatcher := range equalityMatchers {
		valid = append(valid, generateValidStringLiteralCases(operator, equalityMatcher)...)
		invalid = append(invalid, generateInvalidCases(operator, equalityMatcher, preferredMatcher, preferredMatcherWhenNegated)...)
	}

	return valid, invalid
}

func TestPreferComparisonMatcherRule(t *testing.T) {
	var validCases []rule_tester.ValidTestCase
	var invalidCases []rule_tester.InvalidTestCase

	operators := []struct {
		operator                    string
		preferredMatcher            string
		preferredMatcherWhenNegated string
	}{
		{">", "toBeGreaterThan", "toBeLessThanOrEqual"},
		{"<", "toBeLessThan", "toBeGreaterThanOrEqual"},
		{">=", "toBeGreaterThanOrEqual", "toBeLessThan"},
		{"<=", "toBeLessThanOrEqual", "toBeGreaterThan"},
	}

	for _, op := range operators {
		valid, invalid := comparisonOperatorCases(op.operator, op.preferredMatcher, op.preferredMatcherWhenNegated)
		validCases = append(validCases, valid...)
		invalidCases = append(invalidCases, invalid...)
	}

	validCases = append(validCases,
		rule_tester.ValidTestCase{Code: `expect.hasAssertions`},
		rule_tester.ValidTestCase{Code: `expect.hasAssertions()`},
		rule_tester.ValidTestCase{Code: `expect.assertions(1)`},
		rule_tester.ValidTestCase{Code: `expect(true).toBe(...true)`},
		rule_tester.ValidTestCase{Code: `expect()`},
		rule_tester.ValidTestCase{Code: `expect({}).toStrictEqual({})`},
		rule_tester.ValidTestCase{Code: `expect(a === b).toBe(true)`},
		rule_tester.ValidTestCase{Code: `expect(a !== 2).toStrictEqual(true)`},
		rule_tester.ValidTestCase{Code: `expect(a === b).not.toEqual(true)`},
		rule_tester.ValidTestCase{Code: `expect(a !== "string").toStrictEqual(true)`},
		rule_tester.ValidTestCase{Code: `expect(5 != a).toBe(true)`},
		rule_tester.ValidTestCase{Code: `expect(a == "string").toBe(true)`},
		rule_tester.ValidTestCase{Code: `expect(a == "string").not.toBe(true)`},
	)

	invalidCases = append(invalidCases,
		rule_tester.InvalidTestCase{
			Code:   `expect(value > 1).toBe(true as const);`,
			Output: []string{`expect(value).toBeGreaterThan(1);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useToBeComparison", Line: 1, Column: 19},
			},
		},
		rule_tester.InvalidTestCase{
			Code:   `expect((a, b) > c).toBe(true);`,
			Output: []string{`expect((a, b)).toBeGreaterThan(c);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useToBeComparison", Line: 1, Column: 20},
			},
		},
		rule_tester.InvalidTestCase{
			Code:   `expect(a > (b, c)).toBe(true);`,
			Output: []string{`expect(a).toBeGreaterThan((b, c));`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useToBeComparison", Line: 1, Column: 20},
			},
		},
	)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_comparison_matcher.PreferComparisonMatcherRule,
		validCases,
		invalidCases,
	)
}
