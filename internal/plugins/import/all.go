package import_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		rules.NoSelfImportRule,
	}
}
