package stylistic_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/array_bracket_spacing"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/arrow_parens"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		array_bracket_spacing.ArrayBracketSpacingRule,
		arrow_parens.ArrowParensRule,
	}
}
