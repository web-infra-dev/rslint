package node_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/hashbang"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_exports_assign"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		hashbang.HashbangRule,
		no_exports_assign.NoExportsAssignRule,
	}
}
