package n_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/n/rules/hashbang"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{hashbang.HashbangRule}
}
