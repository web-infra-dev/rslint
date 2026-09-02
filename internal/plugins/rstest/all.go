package rstest

import (
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/consistent_each_for"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/consistent_rstest_namespace"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/consistent_test_filename"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/expect_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/hoisted_apis_on_top"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/max_expects"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_alias_methods"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_commented_out_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_conditional_expect"
	no_conditional_in "github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_conditional_in_test"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_conditional_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_disabled_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_focused_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_hooks"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_identical_title"
	no_import_node "github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_import_node_test"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_interpolation_in_snapshots"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_mocks_import"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_standalone_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_called_exactly_once_with"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_called_once"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_called_times"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_each"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_expect_type_of"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_hooks_in_order"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_import_in_mock"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_rs_mocked"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_strict_boolean_matchers"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_to_be_falsy"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_to_be_truthy"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_todo"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_awaited_expect_poll"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_local_test_context_for_concurrent_snapshots"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_mock_type_parameters"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_test_timeout"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_expect_in_promise"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_title"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/warn_todo"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		consistent_each_for.ConsistentEachForRule,
		consistent_rstest_namespace.ConsistentRstestNamespaceRule,
		consistent_test_filename.ConsistentTestFilenameRule,
		expect_expect.ExpectExpectRule,
		hoisted_apis_on_top.HoistedApisOnTopRule,
		max_expects.MaxExpectsRule,
		no_alias_methods.NoAliasMethodsRule,
		no_commented_out_tests.NoCommentedOutTestsRule,
		no_conditional_expect.NoConditionalExpectRule,
		no_conditional_in.NoConditionalInTestRule,
		no_conditional_tests.NoConditionalTestsRule,
		no_disabled_tests.NoDisabledTestsRule,
		no_focused_tests.NoFocusedTestsRule,
		no_hooks.NoHooksRule,
		no_identical_title.NoIdenticalTitleRule,
		no_import_node.NoImportNodeTestRule,
		no_interpolation_in_snapshots.NoInterpolationInSnapshotsRule,
		no_mocks_import.NoMocksImportRule,
		no_standalone_expect.NoStandaloneExpectRule,
		prefer_called_exactly_once_with.PreferCalledExactlyOnceWithRule,
		prefer_called_once.PreferCalledOnceRule,
		prefer_called_times.PreferCalledTimesRule,
		prefer_each.PreferEachRule,
		prefer_expect_type_of.PreferExpectTypeOfRule,
		prefer_hooks_in_order.PreferHooksInOrderRule,
		prefer_import_in_mock.PreferImportInMockRule,
		prefer_rs_mocked.PreferRsMockedRule,
		prefer_strict_boolean_matchers.PreferStrictBooleanMatchersRule,
		prefer_to_be_falsy.PreferToBeFalsyRule,
		prefer_to_be_truthy.PreferToBeTruthyRule,
		prefer_todo.PreferTodoRule,
		require_awaited_expect_poll.RequireAwaitedExpectPollRule,
		require_local_test_context_for_concurrent_snapshots.RequireLocalTestContextForConcurrentSnapshotsRule,
		require_mock_type_parameters.RequireMockTypeParametersRule,
		require_test_timeout.RequireTestTimeoutRule,
		valid_expect.ValidExpectRule,
		valid_expect_in_promise.ValidExpectInPromiseRule,
		valid_title.ValidTitleRule,
		warn_todo.WarnTodoRule,
	}
}
