package promise_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/always_return"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_return_wrap"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/param_names"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		always_return.AlwaysReturnRule,
		no_return_wrap.NoReturnWrapRule,
		param_names.ParamNamesRule,
	}
}
