package jest

import (
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_disabled_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_focused_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_hooks"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_have_length"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_describe_callback"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		no_disabled_tests.NoDisabledTestsRule,
		no_focused_tests.NoFocusedTestsRule,
		no_hooks.NoHooksRule,
		prefer_to_have_length.PreferToHaveLengthRule,
		valid_describe_callback.ValidDescribeCallbackRule,
	}
}
