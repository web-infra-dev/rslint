package vue_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/vue/rules/no_duplicate_attributes"
	"github.com/web-infra-dev/rslint/internal/plugins/vue/rules/no_export_in_script_setup"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		no_duplicate_attributes.NoDuplicateAttributesRule,
		no_export_in_script_setup.NoExportInScriptSetupRule,
	}
}
