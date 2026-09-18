package rslim_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/rslim/rules/require_dynamic_import_entry"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{require_dynamic_import_entry.RequireDynamicImportEntryRule}
}
