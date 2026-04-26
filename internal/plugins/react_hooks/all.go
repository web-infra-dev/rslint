package react_hooks

import (
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/exhaustive_deps"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/rules_of_hooks"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		rules_of_hooks.RulesOfHooksRule,
		exhaustive_deps.ExhaustiveDepsRule,
	}
}
