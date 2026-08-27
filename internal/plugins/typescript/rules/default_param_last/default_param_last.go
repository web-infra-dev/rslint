package default_param_last

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	core "github.com/web-infra-dev/rslint/internal/rules/default_param_last"
)

// DefaultParamLastRule enforces default parameters to be last
var DefaultParamLastRule = rule.CreateRule(rule.Rule{
	Name:   "default-param-last",
	Schema: rule.EmptyArraySchema,
	Run:    core.DefaultParamLastRule.Run,
})
