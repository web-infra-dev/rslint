package promise_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/param_names"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		param_names.ParamNamesRule,
	}
}
