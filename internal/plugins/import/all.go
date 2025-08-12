package import_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_self_import"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		no_self_import.NoSelfImportRule,
	}
}
