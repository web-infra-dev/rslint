// Mirrors Jest v29.16.1 and Vitest v1.6.27; Rstest-only cases live in prefer_comparison_matcher_extras_test.go.
package prefer_comparison_matcher

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func comparisonError(matcher, output string) rule_tester.InvalidTestCaseError {
	expectedError := rule_tester.InvalidTestCaseError{
		MessageId: "useToBeComparison",
		Message:   "Prefer using `" + matcher + "` instead",
	}
	if output != "" {
		expectedError.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: "suggestComparisonMatcher", Output: output,
		}}
	}
	return expectedError
}

func TestPreferComparisonMatcherUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `expect.hasAssertions`}, {Code: `expect.hasAssertions()`}, {Code: `expect.assertions(1)`},
		{Code: `expect(true).toBe(...true)`}, {Code: `expect()`}, {Code: `expect().toStrictEqual({})`},
		{Code: `expect({}).toStrictEqual({})`}, {Code: `expect(a === b).toBe(true)`},
		{Code: `expect(a !== 2).toStrictEqual(true)`}, {Code: `expect(a === b).not.toEqual(true)`},
		{Code: `expect(a !== "string").toStrictEqual(true)`}, {Code: `expect(5 != a).toBe(true)`},
		{Code: `expect(a == "string").toBe(true)`}, {Code: `expect(a == "string").not.toBe(true)`},
		{Code: `expect(a).to.be.a("string");`}, {Code: `expect().fail('Should not succeed a HTTPS proxy request.');`},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, op := range []struct{ operator, matcher, inverse string }{
		{">", "toBeGreaterThan", "toBeLessThanOrEqual"}, {"<", "toBeLessThan", "toBeGreaterThanOrEqual"},
		{">=", "toBeGreaterThanOrEqual", "toBeLessThan"}, {"<=", "toBeLessThanOrEqual", "toBeGreaterThan"},
	} {
		for _, matcher := range []string{op.matcher, op.inverse} {
			valid = append(valid, rule_tester.ValidTestCase{Code: "expect(value)." + matcher + "(1);"}, rule_tester.ValidTestCase{Code: "expect(value).not." + matcher + "(1);"})
		}
		for _, equality := range []string{"toBe", "toEqual", "toStrictEqual"} {
			for _, literal := range []string{"'y'", "`y`", "`y${z}`"} {
				for _, pair := range [][2]string{{"x", literal}, {literal, "x"}} {
					for _, modifier := range []string{"", "not.", "resolves.", "resolves.not."} {
						for _, boolean := range []string{"true", "false"} {
							valid = append(valid, rule_tester.ValidTestCase{Code: fmt.Sprintf("expect(%s %s %s).%s%s(%s)", pair[0], op.operator, pair[1], modifier, equality, boolean)})
						}
					}
				}
			}
			for _, shape := range []struct {
				subjectSuffix, accessor, argument string
				negated                           bool
				column                            int
			}{
				{"", ".%s", "true", false, 18}, {",", ".%s", "true,", false, 19}, {"", "['%s']", "true", false, 18},
				{"", ".resolves.%s", "true", false, 27},
				{"", ".%s", "false", true, 18}, {"", "['%s']", "false", true, 18}, {"", ".resolves.%s", "false", true, 27},
				{"", ".not.%s", "true", true, 22}, {"", "['not'].%s", "true", true, 25}, {"", ".resolves.not.%s", "true", true, 31},
				{"", ".not.%s", "false", false, 22}, {"", ".resolves.not.%s", "false", false, 31},
				{"", `["resolves"].not.%s`, "false", false, 34}, {"", `["resolves"]["not"].%s`, "false", false, 37},
				{"", `["resolves"]["not"]['%s']`, "false", false, 37},
			} {
				code := fmt.Sprintf("expect(value %s 1%s)%s(%s);", op.operator, shape.subjectSuffix, fmt.Sprintf(shape.accessor, equality), shape.argument)
				// Relational subjects are booleans, so upstream promise-modifier cases are invalid Rstest assertions.
				if strings.Contains(shape.accessor, "resolves") {
					valid = append(valid, rule_tester.ValidTestCase{Code: code})
					continue
				}
				preferred := op.matcher
				if shape.negated {
					preferred = "not." + preferred
				}
				// Rstest keeps upstream's diagnostic but withholds the edit because
				// moving a non-constant operand after expect() can change evaluation.
				expectedError := comparisonError(preferred, "")
				expectedError.Line, expectedError.Column = 1, shape.column+len(op.operator)
				invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{expectedError}})
			}
		}
		// All seven Vitest invalid cases: four positive comparisons and three negated comparisons.
		for _, negated := range []bool{false, true} {
			if negated && op.operator == "<=" {
				continue
			}
			modifier := ""
			if negated {
				modifier = "not."
			}
			code := fmt.Sprintf("expect(a %s b).%stoBe(true)", op.operator, modifier)
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{comparisonError(modifier+op.matcher, "")}})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferComparisonMatcherRule, valid, invalid)
}
