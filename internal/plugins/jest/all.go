package jest

import (
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/expect_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_alias_methods"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_commented_out_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_deprecated_functions"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_disabled_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_done_callback"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_duplicate_hooks"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_focused_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_hooks"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_identical_title"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_jasmine_globals"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_mocks_import"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_standalone_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_test_prefixes"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_called_with"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_each"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_strict_equal"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_be"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_contain"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_have_been_called"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_have_been_called_times"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_have_length"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_todo"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_describe_callback"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_title"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		expect_expect.ExpectExpectRule,
		no_alias_methods.NoAliasMethodsRule,
		no_commented_out_tests.NoCommentedOutTestsRule,
		no_deprecated_functions.NoDeprecatedFunctionsRule,
		no_disabled_tests.NoDisabledTestsRule,
		no_done_callback.NoDoneCallbackRule,
		no_duplicate_hooks.NoDuplicateHooksRule,
		no_focused_tests.NoFocusedTestsRule,
		no_hooks.NoHooksRule,
		no_identical_title.NoIdenticalTitleRule,
		no_jasmine_globals.NoJasmineGlobalsRule,
		no_mocks_import.NoMocksImportRule,
		no_standalone_expect.NoStandaloneExpectRule,
		no_test_prefixes.NoTestPrefixesRule,
		prefer_called_with.PreferCalledWithRule,
		prefer_each.PreferEachRule,
		prefer_strict_equal.PreferStrictEqualRule,
		prefer_to_be.PreferToBeRule,
		prefer_to_contain.PreferToContainRule,
		prefer_to_have_been_called_times.PreferToHaveBeenCalledTimesRule,
		prefer_to_have_been_called.PreferToHaveBeenCalledRule,
		prefer_to_have_length.PreferToHaveLengthRule,
		prefer_todo.PreferTodoRule,
		valid_describe_callback.ValidDescribeCallbackRule,
		valid_expect.ValidExpectRule,
		valid_title.ValidTitleRule,
	}
}
