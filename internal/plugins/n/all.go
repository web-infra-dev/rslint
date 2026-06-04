package n_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/n/rules/no_deprecated_api"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		no_deprecated_api.NoDeprecatedApiRule,
	}
}
