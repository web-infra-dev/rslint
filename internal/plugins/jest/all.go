package jest

import (
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_alias_methods"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_disabled_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_focused_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_hooks"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_test_prefixes"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_describe_callback"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		no_alias_methods.NoAliasMethodsRule,
		no_disabled_tests.NoDisabledTestsRule,
		no_focused_tests.NoFocusedTestsRule,
		no_hooks.NoHooksRule,
		no_test_prefixes.NoTestPrefixesRule,
		valid_describe_callback.ValidDescribeCallbackRule,
	}
}
