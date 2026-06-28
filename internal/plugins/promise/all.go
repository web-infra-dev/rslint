package promise_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/always_return"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/avoid_new"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/catch_or_return"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_callback_in_promise"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_multiple_resolved"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_nesting"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_new_statics"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_promise_in_callback"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_return_in_finally"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_return_wrap"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/param_names"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/prefer_catch"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/valid_params"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		always_return.AlwaysReturnRule,
		avoid_new.AvoidNewRule,
		catch_or_return.CatchOrReturnRule,
		no_callback_in_promise.NoCallbackInPromiseRule,
		no_multiple_resolved.NoMultipleResolvedRule,
		no_nesting.NoNestingRule,
		no_new_statics.NoNewStaticsRule,
		no_promise_in_callback.NoPromiseInCallbackRule,
		no_return_in_finally.NoReturnInFinallyRule,
		no_return_wrap.NoReturnWrapRule,
		param_names.ParamNamesRule,
		prefer_catch.PreferCatchRule,
		valid_params.ValidParamsRule,
	}
}
