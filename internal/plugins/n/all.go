package n

import (
	"github.com/web-infra-dev/rslint/internal/plugins/n/rules/no_exports_assign"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		no_exports_assign.NoExportsAssignRule,
	}
}
