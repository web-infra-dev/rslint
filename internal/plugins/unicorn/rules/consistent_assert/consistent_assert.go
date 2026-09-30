package consistent_assert

import (
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/consistent-assert.js
var ConsistentAssertRule = rule.Rule{
	Name:   "unicorn/consistent-assert",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
