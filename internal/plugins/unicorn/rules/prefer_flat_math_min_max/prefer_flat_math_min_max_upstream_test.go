package prefer_flat_math_min_max_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_flat_math_min_max"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferFlatMathMinMaxUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_flat_math_min_max.PreferFlatMathMinMaxRule,
		[]rule_tester.ValidTestCase{
			{Code: "Math.max(a, b, c);"},
			{Code: "Math.min(a, b, c);"},
			{Code: "Math.max(Math.min(a, b), c);"},
			{Code: "Math.min(Math.max(a, b), c);"},
			{Code: "max(max(a, b), c);"},
			{Code: "Math.max.apply(Math, [Math.max(a, b), c]);"},
			{Code: "Math[\"max\"](Math[\"max\"](a, b), c);"},
			{Code: "Math.max?.(Math.max(a, b), c);"},
			{Code: "Math?.max(Math.max(a, b), c);"},
			{Code: "Math.max(Math.max?.(a, b), c);"},
			{Code: "Math.max(Math?.max(a, b), c);"},
			{Code: "globalThis.Math.max(Math.max(a, b), c);"},
			{Code: "Number.max(Number.max(a, b), c);"},
			{Code: "Math.hypot(Math.hypot(a, b), c);"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "Math.max(Math.max(a, b), c);", Output: []string{"Math.max(a, b, c);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "Math.min(a, Math.min(b, c));", Output: []string{"Math.min(a, b, c);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "Math.max(Math.max(a, b), Math.max(c, d), e);", Output: []string{"Math.max(a, b, c, d, e);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "Math.min(Math.min(Math.min(a, b), c), d);", Output: []string{"Math.min(a, b, c, d);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "Math.max(Math.max(a, b));", Output: []string{"Math.max(a, b);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "Math.min(Math.min());", Output: []string{"Math.min();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "Math.max(Math.max(...values), fallback);", Output: []string{"Math.max(...values, fallback);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "const value = Math.max(foo, Math.max(bar, baz)).toString();", Output: []string{"const value = Math.max(foo, bar, baz).toString();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "const value = Math.min((Math.min(a, b)), c);", Output: []string{"const value = Math.min(a, b, c);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "const value = Math.max(\n\tMath.max(a, b),\n\tc,\n);", Output: []string{"const value = Math.max(a, b, c);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "const value = Math.max(\n\tMath.max(\n\t\ta,\n\t\tb,\n\t),\n\tc,\n);", Output: []string{"const value = Math.max(a, b, c);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
			{Code: "const value = Math.max(\n\tMath.max(/* keep */ a, b),\n\tc,\n);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max"}}},
		},
	)
}
