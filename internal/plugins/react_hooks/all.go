package react_hooks

import (
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/component_hook_factories"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/error_boundaries"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/exhaustive_deps"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/globals"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/immutability"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/incompatible_library"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/preserve_manual_memoization"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/rules_of_hooks"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		rules_of_hooks.RulesOfHooksRule,
		exhaustive_deps.ExhaustiveDepsRule,
		component_hook_factories.ComponentHookFactoriesRule,
		error_boundaries.ErrorBoundariesRule,
		globals.GlobalsRule,
		immutability.ImmutabilityRule,
		incompatible_library.IncompatibleLibraryRule,
		preserve_manual_memoization.PreserveManualMemoizationRule,
	}
}
