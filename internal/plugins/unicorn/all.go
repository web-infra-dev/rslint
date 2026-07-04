package unicorn_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/filename_case"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_static_only_class"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_switch_case"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat_map"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		filename_case.FilenameCaseRule,
		no_static_only_class.NoStaticOnlyClassRule,
		no_useless_switch_case.NoUselessSwitchCaseRule,
		prefer_array_flat_map.PreferArrayFlatMapRule,
	}
}
