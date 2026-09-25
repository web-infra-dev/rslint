package prefer_flat_math_min_max_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_flat_math_min_max"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferFlatMathMinMaxExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_flat_math_min_max.PreferFlatMathMinMaxRule,
		[]rule_tester.ValidTestCase{
			{Code: "Math.max(Math.min(a, b), Math.max?.(c, d));"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "Math.max((Math.max(a, b)), Math.min(c, d));", Output: []string{"Math.max(a, b, Math.min(c, d));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max", Message: "Prefer a flat `Math.max()` call instead of nested calls.", Line: 1, Column: 1, EndLine: 1, EndColumn: 43}}},
		},
	)
}
