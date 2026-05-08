package jsx_a11y_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/alt_text"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		alt_text.AltTextRule,
	}
}
