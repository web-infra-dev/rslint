package promise_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/catch_or_return"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_return_in_finally"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_return_wrap"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/param_names"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		catch_or_return.CatchOrReturnRule,
		no_return_in_finally.NoReturnInFinallyRule,
		no_return_wrap.NoReturnWrapRule,
		param_names.ParamNamesRule,
	}
}
