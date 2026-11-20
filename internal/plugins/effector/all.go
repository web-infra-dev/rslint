package effector_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/effector/rules/no_watch"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		no_watch.NoWatchRule,
	}
}
