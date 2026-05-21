package valid_expect

import (
	"github.com/web-infra-dev/rslint/internal/rule"
)

var ValidExpectRule = rule.Rule{
	Name: "jest/valid-expect",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return nil
	},
}
