package no_duplicate_logical_operands

import "github.com/web-infra-dev/rslint/internal/rule"

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/rules/no-duplicate-logical-operands.js
var NoDuplicateLogicalOperandsRule = rule.Rule{
	Name:   "unicorn/no-duplicate-logical-operands",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
