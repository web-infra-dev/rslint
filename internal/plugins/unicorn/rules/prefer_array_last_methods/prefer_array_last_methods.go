package prefer_array_last_methods

import "github.com/web-infra-dev/rslint/internal/rule"

// PreferArrayLastMethodsRule is intentionally inert for the tests-first red checkpoint.
// Port target: eslint-plugin-unicorn v75.0.0.
var PreferArrayLastMethodsRule = rule.Rule{
	Name:   "unicorn/prefer-array-last-methods",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
