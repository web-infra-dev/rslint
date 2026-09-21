package prefer_type_error

import (
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferTypeErrorRule is intentionally inert while the test-first port is red.
var PreferTypeErrorRule = rule.Rule{
	Name:   "unicorn/prefer-type-error",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
