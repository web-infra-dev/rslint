package jest

import (
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_disabled_tests"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_describe_callback"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetRecommendedRules() []rule.Rule {
	return []rule.Rule{
		no_disabled_tests.NoDisabledTestsRule,
		valid_describe_callback.ValidDescribeCallbackRule,
	}
}
