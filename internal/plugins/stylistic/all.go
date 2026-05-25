package stylistic_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/array_bracket_spacing"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/arrow_parens"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/arrow_spacing"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/block_spacing"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/brace_style"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/comma_dangle"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/comma_spacing"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/comma_style"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/computed_property_spacing"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/dot_location"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/eol_last"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/generator_star_spacing"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/jsx_closing_bracket_location"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/jsx_closing_tag_location"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		array_bracket_spacing.ArrayBracketSpacingRule,
		arrow_parens.ArrowParensRule,
		arrow_spacing.ArrowSpacingRule,
		block_spacing.BlockSpacingRule,
		brace_style.BraceStyleRule,
		comma_dangle.CommaDangleRule,
		comma_spacing.CommaSpacingRule,
		comma_style.CommaStyleRule,
		computed_property_spacing.ComputedPropertySpacingRule,
		dot_location.DotLocationRule,
		eol_last.EolLastRule,
		generator_star_spacing.GeneratorStarSpacingRule,
		jsx_closing_bracket_location.JsxClosingBracketLocationRule,
		jsx_closing_tag_location.JsxClosingTagLocationRule,
	}
}
