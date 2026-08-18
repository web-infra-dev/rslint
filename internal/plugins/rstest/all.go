package rstest

import (
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/expect_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_commented_out_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_conditional_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_disabled_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_focused_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_identical_title"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_interpolation_in_snapshots"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_mocks_import"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_standalone_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_title"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		expect_expect.ExpectExpectRule,
		no_commented_out_tests.NoCommentedOutTestsRule,
		no_conditional_expect.NoConditionalExpectRule,
		no_disabled_tests.NoDisabledTestsRule,
		no_focused_tests.NoFocusedTestsRule,
		no_identical_title.NoIdenticalTitleRule,
		no_interpolation_in_snapshots.NoInterpolationInSnapshotsRule,
		no_mocks_import.NoMocksImportRule,
		no_standalone_expect.NoStandaloneExpectRule,
		valid_expect.ValidExpectRule,
		valid_title.ValidTitleRule,
	}
}
