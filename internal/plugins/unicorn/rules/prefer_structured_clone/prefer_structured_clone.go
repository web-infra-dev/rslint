package prefer_structured_clone

import "github.com/web-infra-dev/rslint/internal/rule"

// PreferStructuredCloneRule is intentionally inert while the test-first port is red.
var PreferStructuredCloneRule = rule.Rule{
	Name:   "unicorn/prefer-structured-clone",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
