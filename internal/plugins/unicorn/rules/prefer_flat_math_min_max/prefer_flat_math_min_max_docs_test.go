package prefer_flat_math_min_max_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_flat_math_min_max"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferFlatMathMinMaxDocs(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_flat_math_min_max.PreferFlatMathMinMaxRule,
		[]rule_tester.ValidTestCase{
			{Code: "const clamped = Math.max(Math.min(value, upper), lower);"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "const biggest = Math.max(Math.max(a, b), c);", Output: []string{"const biggest = Math.max(a, b, c);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max", Message: "Prefer a flat `Math.max()` call instead of nested calls.", Line: 1, Column: 17}}},
			{Code: "const smallest = Math.min(a, Math.min(b, c));", Output: []string{"const smallest = Math.min(a, b, c);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max", Message: "Prefer a flat `Math.min()` call instead of nested calls.", Line: 1, Column: 18}}},
		},
	)
}
