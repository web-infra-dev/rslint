package jsx_a11y_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/alt_text"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/anchor_has_content"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/anchor_is_valid"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/aria_unsupported_elements"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/autocomplete_valid"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/heading_has_content"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/html_has_lang"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/iframe_has_title"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/img_redundant_alt"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/media_has_caption"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/no_access_key"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/no_autofocus"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/no_distracting_elements"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		alt_text.AltTextRule,
		anchor_has_content.AnchorHasContentRule,
		anchor_is_valid.AnchorIsValidRule,
		aria_unsupported_elements.AriaUnsupportedElementsRule,
		autocomplete_valid.AutocompleteValidRule,
		heading_has_content.HeadingHasContentRule,
		html_has_lang.HtmlHasLangRule,
		iframe_has_title.IframeHasTitleRule,
		img_redundant_alt.ImgRedundantAltRule,
		media_has_caption.MediaHasCaptionRule,
		no_access_key.NoAccessKeyRule,
		no_autofocus.NoAutofocusRule,
		no_distracting_elements.NoDistractingElementsRule,
	}
}
