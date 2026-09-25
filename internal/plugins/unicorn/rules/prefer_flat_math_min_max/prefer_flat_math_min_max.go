package prefer_flat_math_min_max

import "github.com/web-infra-dev/rslint/internal/rule"

// PreferFlatMathMinMaxRule is intentionally inert for the tests-first red checkpoint.
// Port target: eslint-plugin-unicorn v75.0.0.
var PreferFlatMathMinMaxRule = rule.Rule{
	Name:   "unicorn/prefer-flat-math-min-max",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
