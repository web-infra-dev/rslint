package jsx_a11y_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/alt_text"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/anchor_has_content"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/anchor_is_valid"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/aria_activedescendant_has_tabindex"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/aria_props"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/aria_proptypes"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/aria_role"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/aria_unsupported_elements"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/autocomplete_valid"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/click_events_have_key_events"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/control_has_associated_label"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/heading_has_content"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/html_has_lang"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/iframe_has_title"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/img_redundant_alt"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/interactive_supports_focus"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/label_has_associated_control"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/media_has_caption"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/mouse_events_have_key_events"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/no_access_key"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/no_autofocus"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/no_distracting_elements"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/no_noninteractive_element_interactions"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/no_noninteractive_tabindex"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/no_redundant_roles"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/no_static_element_interactions"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/scope"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/tabindex_no_positive"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		alt_text.AltTextRule,
		anchor_has_content.AnchorHasContentRule,
		anchor_is_valid.AnchorIsValidRule,
		aria_activedescendant_has_tabindex.AriaActivedescendantHasTabindexRule,
		aria_props.AriaPropsRule,
		aria_proptypes.AriaProptypesRule,
		aria_role.AriaRoleRule,
		aria_unsupported_elements.AriaUnsupportedElementsRule,
		autocomplete_valid.AutocompleteValidRule,
		click_events_have_key_events.ClickEventsHaveKeyEventsRule,
		control_has_associated_label.ControlHasAssociatedLabelRule,
		heading_has_content.HeadingHasContentRule,
		html_has_lang.HtmlHasLangRule,
		iframe_has_title.IframeHasTitleRule,
		img_redundant_alt.ImgRedundantAltRule,
		interactive_supports_focus.InteractiveSupportsFocusRule,
		label_has_associated_control.LabelHasAssociatedControlRule,
		media_has_caption.MediaHasCaptionRule,
		mouse_events_have_key_events.MouseEventsHaveKeyEventsRule,
		no_access_key.NoAccessKeyRule,
		no_autofocus.NoAutofocusRule,
		no_distracting_elements.NoDistractingElementsRule,
		no_noninteractive_element_interactions.NoNoninteractiveElementInteractionsRule,
		no_noninteractive_tabindex.NoNoninteractiveTabindexRule,
		no_redundant_roles.NoRedundantRolesRule,
		no_static_element_interactions.NoStaticElementInteractionsRule,
		scope.ScopeRule,
		tabindex_no_positive.TabindexNoPositiveRule,
	}
}
