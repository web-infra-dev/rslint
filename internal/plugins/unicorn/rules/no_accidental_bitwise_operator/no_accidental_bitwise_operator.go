package no_accidental_bitwise_operator

import "github.com/web-infra-dev/rslint/internal/rule"

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-accidental-bitwise-operator.js
var NoAccidentalBitwiseOperatorRule = rule.Rule{
	Name:   "unicorn/no-accidental-bitwise-operator",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
