package unicorn_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/filename_case"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/new_for_builtins"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_instanceof_builtins"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_static_only_class"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_thenable"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_switch_case"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat_map"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_number_properties"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_array_join_separator"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_number_to_fixed_digits_argument"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		filename_case.FilenameCaseRule,
		new_for_builtins.NewForBuiltinsRule,
		no_instanceof_builtins.NoInstanceofBuiltinsRule,
		no_static_only_class.NoStaticOnlyClassRule,
		no_thenable.NoThenableRule,
		no_useless_switch_case.NoUselessSwitchCaseRule,
		prefer_array_flat.PreferArrayFlatRule,
		prefer_array_flat_map.PreferArrayFlatMapRule,
		prefer_number_properties.PreferNumberPropertiesRule,
		require_array_join_separator.RequireArrayJoinSeparatorRule,
		require_number_to_fixed_digits_argument.RequireNumberToFixedDigitsArgumentRule,
	}
}
