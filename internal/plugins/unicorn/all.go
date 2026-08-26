package unicorn_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/catch_error_name"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/error_message"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/filename_case"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/new_for_builtins"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_fill_with_reference_type"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_await_in_promise_methods"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_instanceof_builtins"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_invalid_fetch_options"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_invalid_remove_event_listener"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_static_only_class"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_thenable"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unsafe_string_replacement"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_switch_case"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat_map"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_some"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_node_protocol"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_number_properties"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_set_has"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_ternary"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_array_join_separator"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_number_to_fixed_digits_argument"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		catch_error_name.CatchErrorNameRule,
		error_message.ErrorMessageRule,
		filename_case.FilenameCaseRule,
		new_for_builtins.NewForBuiltinsRule,
		no_array_fill_with_reference_type.NoArrayFillWithReferenceTypeRule,
		no_await_in_promise_methods.NoAwaitInPromiseMethodsRule,
		no_instanceof_builtins.NoInstanceofBuiltinsRule,
		no_invalid_fetch_options.NoInvalidFetchOptionsRule,
		no_invalid_remove_event_listener.NoInvalidRemoveEventListenerRule,
		no_static_only_class.NoStaticOnlyClassRule,
		no_thenable.NoThenableRule,
		no_unsafe_string_replacement.NoUnsafeStringReplacementRule,
		no_useless_switch_case.NoUselessSwitchCaseRule,
		prefer_array_flat.PreferArrayFlatRule,
		prefer_array_flat_map.PreferArrayFlatMapRule,
		prefer_array_some.PreferArraySomeRule,
		prefer_node_protocol.PreferNodeProtocolRule,
		prefer_number_properties.PreferNumberPropertiesRule,
		prefer_set_has.PreferSetHasRule,
	prefer_ternary.PreferTernaryRule,
		require_array_join_separator.RequireArrayJoinSeparatorRule,
		require_number_to_fixed_digits_argument.RequireNumberToFixedDigitsArgumentRule,
	}
}
