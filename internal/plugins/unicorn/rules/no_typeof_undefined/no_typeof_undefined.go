package no_typeof_undefined

import "github.com/web-infra-dev/rslint/internal/rule"

// NoTypeofUndefinedRule is intentionally inert while the test-first port is red.
var NoTypeofUndefinedRule = rule.Rule{
	Name:   "unicorn/no-typeof-undefined",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
