package require_array_sort_compare

import "github.com/web-infra-dev/rslint/internal/rule"

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/require-array-sort-compare.js
var RequireArraySortCompareRule = rule.Rule{
	Name:   "unicorn/require-array-sort-compare",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
