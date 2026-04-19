package react_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/button_has_type"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_boolean_value"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_closing_tag_location"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_equals_spacing"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_filename_extension"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_first_prop_new_line"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_max_props_per_line"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_no_bind"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_props_no_multi_spaces"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_uses_react"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_uses_vars"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_wrap_multilines"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/no_danger"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/no_unescaped_entities"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/react_in_jsx_scope"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/self_closing_comp"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/style_prop_object"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/void_dom_elements_no_children"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		button_has_type.ButtonHasTypeRule,
		jsx_boolean_value.JsxBooleanValueRule,
		jsx_closing_tag_location.JsxClosingTagLocationRule,
		jsx_equals_spacing.JsxEqualsSpacingRule,
		jsx_filename_extension.JsxFilenameExtensionRule,
		jsx_first_prop_new_line.JsxFirstPropNewLineRule,
		jsx_max_props_per_line.JsxMaxPropsPerLineRule,
		jsx_no_bind.JsxNoBindRule,
		jsx_props_no_multi_spaces.JsxPropsNoMultiSpacesRule,
		jsx_uses_react.JsxUsesReactRule,
		jsx_uses_vars.JsxUsesVarsRule,
		jsx_wrap_multilines.JsxWrapMultilinesRule,
		no_danger.NoDangerRule,
		no_unescaped_entities.NoUnescapedEntitiesRule,
		react_in_jsx_scope.ReactInJsxScopeRule,
		self_closing_comp.SelfClosingCompRule,
		style_prop_object.StylePropObjectRule,
		void_dom_elements_no_children.VoidDomElementsNoChildrenRule,
	}
}
