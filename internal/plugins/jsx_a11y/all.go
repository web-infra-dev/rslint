package jsx_a11y_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/alt_text"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/anchor_has_content"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/aria_unsupported_elements"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/autocomplete_valid"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		alt_text.AltTextRule,
		anchor_has_content.AnchorHasContentRule,
		aria_unsupported_elements.AriaUnsupportedElementsRule,
		autocomplete_valid.AutocompleteValidRule,
	}
}
