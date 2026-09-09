// Regression and branch coverage beyond prefer_comparison_matcher_upstream_test.go.
package prefer_comparison_matcher_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_comparison_matcher"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferComparisonMatcherExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `expect(value > 1)[matcher].toBe(true);`}, {Code: `expect(value > 1)[foo()].toBe(true);`},
		{Code: `expect(a > b).toBe();`}, {Code: `expect(a > b).toBe(flag);`}, {Code: `expect(a > b).toBe(...flags);`},
		{Code: `expect(a + b).toBe(true);`}, {Code: `expect(a > b).toContain(true);`},
		{Code: `expect.soft(a > b).toBe(true);`}, {Code: `expect(a > b).toBe(true satisfies boolean);`},
		{Code: `expect(a > ('b' as string)).toBe(true);`},
		{Code: `function test(expect: any) { expect(a > b).toBe(true); }`},
		{Code: `import { expect } from '@rstest/core'; expect(a > b).toBe(true);`},
		{Code: `(expect(a > b) as any).toBe(true);`}, {Code: `expect(a > b)!.toBe(true);`},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct {
		code, output, matcher string
		column                int
	}{
		// ---- Dimension 4: accessor, receiver and literal wrappers ----
		{`expect(value > 1)[("toBe")](true);`, `expect(value).toBeGreaterThan(1);`, "toBeGreaterThan", 20},
		{`expect(value > 1).toBe(true as const);`, `expect(value).toBeGreaterThan(1);`, "toBeGreaterThan", 19},
		{`expect(value > 1).toBe((true));`, `expect(value).toBeGreaterThan((1));`, "toBeGreaterThan", 19},
		{`expect((a, b) > c).toBe(true);`, `expect((a, b)).toBeGreaterThan(c);`, "toBeGreaterThan", 20},
		{`expect(a > (b, c)).toBe(true);`, `expect(a).toBeGreaterThan((b, c));`, "toBeGreaterThan", 20},
		{`expect(((a) > (b))).toBe(true);`, `expect((a)).toBeGreaterThan(b);`, "toBeGreaterThan", 21},
		{`expect(value > 1).toBe(true).toString();`, `expect(value).toBeGreaterThan(1).toString();`, "toBeGreaterThan", 19},
		{`expect(value > 1).toBe(true).foo(false);`, `expect(value).toBeGreaterThan(1).foo(false);`, "toBeGreaterThan", 19},
		{`expect(a > b)?.toBe(true);`, `expect(a).toBeGreaterThan(b);`, "toBeGreaterThan", 16},
		{`expect(a > b).toBe?.(true);`, `expect(a).toBeGreaterThan?.(b);`, "toBeGreaterThan", 15},
		// Locks in upstream's inverted-operator policy, including its NaN limitation.
		{`expect(NaN > 1).toBe(false);`, `expect(NaN).toBeLessThanOrEqual(1);`, "toBeLessThanOrEqual", 17},
		{`expect(a <= b).rejects.not.toBe(true);`, `expect(a).rejects.toBeGreaterThan(b);`, "toBeGreaterThan", 28},
		// ---- Real-user: jest-community/eslint-plugin-jest#2000 empty subject ----
		{`expect().toBe(true); expect(a > b).toBe(true);`, `expect().toBe(true); expect(a).toBeGreaterThan(b);`, "toBeGreaterThan", 36},
		// ---- Real-user: trailing comma formatting in generated assertions ----
		{`expect(a > b,).toBe(true,);`, `expect(a,).toBeGreaterThan(b,);`, "toBeGreaterThan", 16},
		// ---- Dimension 4: comments and UTF-16 ----
		{`expect(a /* note */ > b).toBe(true);`, `expect(a).toBeGreaterThan(b);`, "toBeGreaterThan", 26},
		{`expect(a > b, "😀").toBe(true);`, `expect(a, "😀").toBeGreaterThan(b);`, "toBeGreaterThan", 21},
		// N/A: declaration names and private fields do not participate in matcher selection.
	} {
		expectedError := rule_tester.InvalidTestCaseError{MessageId: "useToBeComparison", Message: "Prefer using `" + test.matcher + "` instead", Line: 1, Column: test.column}
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: test.code, Output: []string{test.output}, Errors: []rule_tester.InvalidTestCaseError{expectedError}})
	}
	invalid = append(invalid,
		rule_tester.InvalidTestCase{Code: `(expect)(a > b).toBe(true);`, Output: []string{`(expect)(a).toBeGreaterThan(b);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeComparison"}}},
		rule_tester.InvalidTestCase{Code: `(expect(a > b)).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeComparison"}}},
		rule_tester.InvalidTestCase{Code: `((expect(a > b).not)).toBe(false);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeComparison"}}},
	)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_comparison_matcher.PreferComparisonMatcherRule, valid, invalid)
}
