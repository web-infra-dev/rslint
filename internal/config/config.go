package config

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
	importPlugin "github.com/web-infra-dev/rslint/internal/plugins/import"
	jestPlugin "github.com/web-infra-dev/rslint/internal/plugins/jest"
	promisePlugin "github.com/web-infra-dev/rslint/internal/plugins/promise"
	reactPlugin "github.com/web-infra-dev/rslint/internal/plugins/react"
	reactHooksPlugin "github.com/web-infra-dev/rslint/internal/plugins/react_hooks"
	unicornPlugin "github.com/web-infra-dev/rslint/internal/plugins/unicorn"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/adjacent_overload_signatures"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/array_type"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/await_thenable"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/ban_ts_comment"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/ban_tslint_comment"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/ban_types"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/class_literal_property_style"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/consistent_generic_constructors"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/consistent_indexed_object_style"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/consistent_return"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/consistent_type_assertions"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/consistent_type_definitions"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/consistent_type_exports"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/consistent_type_imports"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/default_param_last"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/dot_notation"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/explicit_function_return_type"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/explicit_member_accessibility"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/member_ordering"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/method_signature_style"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/naming_convention"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_array_constructor"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_array_delete"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_base_to_string"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_confusing_void_expression"
	ts_no_dupe_class_members "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_dupe_class_members"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_duplicate_enum_values"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_duplicate_type_constituents"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_dynamic_delete"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_empty_function"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_empty_interface"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_explicit_any"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_extra_non_null_assertion"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_extraneous_class"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_floating_promises"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_for_in_array"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_implied_eval"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_inferrable_types"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_invalid_void_type"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_magic_numbers"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_meaningless_void_operator"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_misused_new"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_misused_promises"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_misused_spread"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_mixed_enums"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_namespace"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_non_null_asserted_nullish_coalescing"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_non_null_asserted_optional_chain"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_non_null_assertion"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_redeclare"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_redundant_type_constituents"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_require_imports"
	ts_no_restricted_imports "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_restricted_imports"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_this_alias"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_boolean_literal_compare"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_condition"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_template_expression"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_type_arguments"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_type_assertion"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_type_constraint"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_argument"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_assignment"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_call"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_enum_comparison"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_member_access"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_return"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_type_assertion"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_unary_minus"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unused_expressions"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unused_vars"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_use_before_define"
	ts_no_useless_constructor "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_useless_constructor"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_useless_empty_export"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_var_requires"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/non_nullable_type_assertion_style"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/only_throw_error"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/parameter_properties"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_as_const"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_destructuring"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_for_of"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_includes"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_literal_enum_member"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_namespace_keyword"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_optional_chain"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_promise_reject_errors"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_readonly"

	// "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_readonly_parameter_types" // Temporarily disabled - incomplete implementation
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_reduce_type_parameter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_regexp_exec"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_return_this_type"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_string_starts_ends_with"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_ts_expect_error"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/promise_function_async"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/related_getter_setter_pairs"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/require_array_sort_compare"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/require_await"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/restrict_plus_operands"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/restrict_template_expressions"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/return_await"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/switch_exhaustiveness_check"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/triple_slash_reference"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/unbound_method"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/unified_signatures"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/use_unknown_in_catch_callback_variable"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules/accessor_pairs"
	"github.com/web-infra-dev/rslint/internal/rules/array_callback_return"
	"github.com/web-infra-dev/rslint/internal/rules/constructor_super"
	"github.com/web-infra-dev/rslint/internal/rules/default_case"
	"github.com/web-infra-dev/rslint/internal/rules/default_case_last"
	"github.com/web-infra-dev/rslint/internal/rules/eqeqeq"
	"github.com/web-infra-dev/rslint/internal/rules/for_direction"
	"github.com/web-infra-dev/rslint/internal/rules/getter_return"
	"github.com/web-infra-dev/rslint/internal/rules/guard_for_in"
	"github.com/web-infra-dev/rslint/internal/rules/max_lines"
	"github.com/web-infra-dev/rslint/internal/rules/no_alert"
	"github.com/web-infra-dev/rslint/internal/rules/no_async_promise_executor"
	"github.com/web-infra-dev/rslint/internal/rules/no_await_in_loop"
	"github.com/web-infra-dev/rslint/internal/rules/no_bitwise"
	"github.com/web-infra-dev/rslint/internal/rules/no_caller"
	"github.com/web-infra-dev/rslint/internal/rules/no_case_declarations"
	"github.com/web-infra-dev/rslint/internal/rules/no_class_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_compare_neg_zero"
	"github.com/web-infra-dev/rslint/internal/rules/no_cond_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_console"
	"github.com/web-infra-dev/rslint/internal/rules/no_const_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_constant_binary_expression"
	"github.com/web-infra-dev/rslint/internal/rules/no_constant_condition"
	"github.com/web-infra-dev/rslint/internal/rules/no_constructor_return"
	"github.com/web-infra-dev/rslint/internal/rules/no_control_regex"
	"github.com/web-infra-dev/rslint/internal/rules/no_debugger"
	"github.com/web-infra-dev/rslint/internal/rules/no_delete_var"
	"github.com/web-infra-dev/rslint/internal/rules/no_dupe_args"
	"github.com/web-infra-dev/rslint/internal/rules/no_dupe_class_members"
	"github.com/web-infra-dev/rslint/internal/rules/no_dupe_else_if"
	"github.com/web-infra-dev/rslint/internal/rules/no_dupe_keys"
	"github.com/web-infra-dev/rslint/internal/rules/no_duplicate_case"
	"github.com/web-infra-dev/rslint/internal/rules/no_empty"
	"github.com/web-infra-dev/rslint/internal/rules/no_empty_character_class"
	"github.com/web-infra-dev/rslint/internal/rules/no_empty_pattern"
	"github.com/web-infra-dev/rslint/internal/rules/no_eval"
	"github.com/web-infra-dev/rslint/internal/rules/no_ex_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_extend_native"
	"github.com/web-infra-dev/rslint/internal/rules/no_extra_bind"
	"github.com/web-infra-dev/rslint/internal/rules/no_extra_boolean_cast"
	"github.com/web-infra-dev/rslint/internal/rules/no_extra_label"
	"github.com/web-infra-dev/rslint/internal/rules/no_fallthrough"
	"github.com/web-infra-dev/rslint/internal/rules/no_func_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_global_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_implicit_coercion"
	core_no_implied_eval "github.com/web-infra-dev/rslint/internal/rules/no_implied_eval"
	"github.com/web-infra-dev/rslint/internal/rules/no_import_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_inner_declarations"
	"github.com/web-infra-dev/rslint/internal/rules/no_invalid_regexp"
	"github.com/web-infra-dev/rslint/internal/rules/no_irregular_whitespace"
	"github.com/web-infra-dev/rslint/internal/rules/no_iterator"
	"github.com/web-infra-dev/rslint/internal/rules/no_label_var"
	"github.com/web-infra-dev/rslint/internal/rules/no_labels"
	"github.com/web-infra-dev/rslint/internal/rules/no_lone_blocks"
	"github.com/web-infra-dev/rslint/internal/rules/no_loop_func"
	"github.com/web-infra-dev/rslint/internal/rules/no_loss_of_precision"
	"github.com/web-infra-dev/rslint/internal/rules/no_misleading_character_class"
	"github.com/web-infra-dev/rslint/internal/rules/no_multi_str"
	"github.com/web-infra-dev/rslint/internal/rules/no_nested_ternary"
	"github.com/web-infra-dev/rslint/internal/rules/no_new"
	"github.com/web-infra-dev/rslint/internal/rules/no_new_func"
	"github.com/web-infra-dev/rslint/internal/rules/no_new_object"
	"github.com/web-infra-dev/rslint/internal/rules/no_new_symbol"
	"github.com/web-infra-dev/rslint/internal/rules/no_new_wrappers"
	"github.com/web-infra-dev/rslint/internal/rules/no_obj_calls"
	"github.com/web-infra-dev/rslint/internal/rules/no_octal"
	"github.com/web-infra-dev/rslint/internal/rules/no_octal_escape"
	"github.com/web-infra-dev/rslint/internal/rules/no_param_reassign"
	"github.com/web-infra-dev/rslint/internal/rules/no_proto"
	"github.com/web-infra-dev/rslint/internal/rules/no_prototype_builtins"
	"github.com/web-infra-dev/rslint/internal/rules/no_regex_spaces"
	"github.com/web-infra-dev/rslint/internal/rules/no_restricted_imports"
	"github.com/web-infra-dev/rslint/internal/rules/no_return_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_script_url"
	"github.com/web-infra-dev/rslint/internal/rules/no_self_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_self_compare"
	"github.com/web-infra-dev/rslint/internal/rules/no_sequences"
	"github.com/web-infra-dev/rslint/internal/rules/no_setter_return"
	"github.com/web-infra-dev/rslint/internal/rules/no_shadow"
	"github.com/web-infra-dev/rslint/internal/rules/no_shadow_restricted_names"
	"github.com/web-infra-dev/rslint/internal/rules/no_sparse_arrays"
	"github.com/web-infra-dev/rslint/internal/rules/no_template_curly_in_string"
	"github.com/web-infra-dev/rslint/internal/rules/no_this_before_super"
	"github.com/web-infra-dev/rslint/internal/rules/no_throw_literal"
	"github.com/web-infra-dev/rslint/internal/rules/no_undef"
	"github.com/web-infra-dev/rslint/internal/rules/no_undef_init"
	"github.com/web-infra-dev/rslint/internal/rules/no_unmodified_loop_condition"
	"github.com/web-infra-dev/rslint/internal/rules/no_unneeded_ternary"
	"github.com/web-infra-dev/rslint/internal/rules/no_unreachable"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_finally"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_negation"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_optional_chaining"
	"github.com/web-infra-dev/rslint/internal/rules/no_useless_call"
	"github.com/web-infra-dev/rslint/internal/rules/no_useless_catch"
	"github.com/web-infra-dev/rslint/internal/rules/no_useless_computed_key"
	"github.com/web-infra-dev/rslint/internal/rules/no_useless_concat"
	"github.com/web-infra-dev/rslint/internal/rules/no_useless_constructor"
	"github.com/web-infra-dev/rslint/internal/rules/no_useless_rename"
	"github.com/web-infra-dev/rslint/internal/rules/no_var"
	"github.com/web-infra-dev/rslint/internal/rules/no_with"
	"github.com/web-infra-dev/rslint/internal/rules/object_shorthand"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_const"
	core_prefer_promise_reject_errors "github.com/web-infra-dev/rslint/internal/rules/prefer_promise_reject_errors"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_rest_params"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_spread"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_template"
	"github.com/web-infra-dev/rslint/internal/rules/radix"
	"github.com/web-infra-dev/rslint/internal/rules/require_atomic_updates"
	"github.com/web-infra-dev/rslint/internal/rules/require_yield"
	"github.com/web-infra-dev/rslint/internal/rules/strict"
	"github.com/web-infra-dev/rslint/internal/rules/symbol_description"
	"github.com/web-infra-dev/rslint/internal/rules/use_isnan"
	"github.com/web-infra-dev/rslint/internal/rules/valid_typeof"
)

// RslintConfig represents the top-level configuration array
type RslintConfig []ConfigEntry

// ConfigEntry represents a single configuration entry in the config array
type ConfigEntry struct {
	Files           []string         `json:"files,omitempty"`
	Ignores         []string         `json:"ignores,omitempty"`
	LanguageOptions *LanguageOptions `json:"languageOptions,omitempty"`
	Rules           Rules            `json:"rules"`
	Plugins         []string         `json:"plugins,omitempty"`
	Settings        Settings         `json:"settings,omitempty"`
}

// Settings represents shared settings accessible to rules
type Settings map[string]interface{}

// LanguageOptions contains language-specific configuration options
type LanguageOptions struct {
	ParserOptions *ParserOptions `json:"parserOptions,omitempty"`
}

// ProjectPaths represents project paths that can be either a single string or an array of strings
type ProjectPaths []string

// UnmarshalJSON implements custom JSON unmarshaling to support both string and string[] formats
func (p *ProjectPaths) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var singlePath string
	if err := json.Unmarshal(data, &singlePath); err == nil {
		*p = []string{singlePath}
		return nil
	}

	// If that fails, try to unmarshal as array of strings
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return err
	}
	*p = paths
	return nil
}

// ParserOptions contains parser-specific configuration.
// ProjectService uses *bool to distinguish "not set" (nil) from "explicitly false".
type ParserOptions struct {
	ProjectService *bool        `json:"projectService,omitempty"`
	Project        ProjectPaths `json:"project,omitempty"`
}

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// Rules represents the rules configuration
// This can be extended to include specific rule configurations
type Rules map[string]interface{}

// RuleConfig represents individual rule configuration
type RuleConfig struct {
	Level   string      `json:"level,omitempty"`   // "error", "warn", "off"
	Options interface{} `json:"options,omitempty"` // Rule-specific options (string, map, array, etc.)
}

// IsEnabled returns true if the rule is enabled (not "off")
func (rc *RuleConfig) IsEnabled() bool {
	if rc == nil {
		return false
	}
	return rc.Level != "off" && rc.Level != ""
}

// GetLevel returns the rule level, defaulting to "error" if not specified
func (rc *RuleConfig) GetLevel() string {
	if rc == nil || rc.Level == "" {
		return "error"
	}
	return rc.Level
}

// GetOptions returns the rule options, ensuring we return a usable value
func (rc *RuleConfig) GetOptions() interface{} {
	if rc == nil || rc.Options == nil {
		return nil
	}
	return rc.Options
}

// SetOptions sets the rule options
func (rc *RuleConfig) SetOptions(options interface{}) {
	if rc != nil {
		rc.Options = options
	}
}

// GetSeverity returns the diagnostic severity for this rule configuration
func (rc *RuleConfig) GetSeverity() rule.DiagnosticSeverity {
	if rc == nil {
		return rule.SeverityError
	}
	return rule.ParseSeverity(rc.Level)
}

// PluginInfo defines a known plugin with its rule prefix and all accepted declaration names.
type PluginInfo struct {
	RulePrefix  string   // Rule name prefix, e.g. "import"
	DeclNames   []string // All accepted declaration names, e.g. ["eslint-plugin-import", "import"]
	getAllRules func() []rule.Rule
}

// KnownPlugins is the single source of truth for all supported plugins.
var KnownPlugins = []PluginInfo{
	{
		RulePrefix:  "@typescript-eslint",
		DeclNames:   []string{"@typescript-eslint"},
		getAllRules: func() []rule.Rule { return GetPluginRules("@typescript-eslint") },
	},
	{
		RulePrefix:  "import",
		DeclNames:   []string{"eslint-plugin-import", "import"},
		getAllRules: func() []rule.Rule { return importPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "jest",
		DeclNames:   []string{"eslint-plugin-jest", "jest"},
		getAllRules: func() []rule.Rule { return jestPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "promise",
		DeclNames:   []string{"eslint-plugin-promise", "promise"},
		getAllRules: func() []rule.Rule { return promisePlugin.GetAllRules() },
	},
	{
		RulePrefix:  "react",
		DeclNames:   []string{"react"},
		getAllRules: func() []rule.Rule { return reactPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "react-hooks",
		DeclNames:   []string{"eslint-plugin-react-hooks", "react-hooks"},
		getAllRules: func() []rule.Rule { return reactHooksPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "unicorn",
		DeclNames:   []string{"eslint-plugin-unicorn", "unicorn"},
		getAllRules: func() []rule.Rule { return unicornPlugin.GetAllRules() },
	},
}

// pluginByDeclName is a lookup table built from KnownPlugins: declaration name → *PluginInfo.
var pluginByDeclName map[string]*PluginInfo

func init() {
	pluginByDeclName = make(map[string]*PluginInfo)
	for i := range KnownPlugins {
		for _, name := range KnownPlugins[i].DeclNames {
			pluginByDeclName[name] = &KnownPlugins[i]
		}
	}
}

// NormalizePluginName converts a plugin declaration name to its rule prefix form.
// Looks up KnownPlugins; returns the input unchanged if not found.
func NormalizePluginName(pluginName string) string {
	if info, ok := pluginByDeclName[pluginName]; ok {
		return info.RulePrefix
	}
	return pluginName
}

// parseArrayRuleConfig parses array-style rule configuration like ["error", {...options}]
// Supports ESLint-compatible formats:
// - ["off"] -> disabled rule
// - ["error"] -> enabled rule with error severity
// - ["warn"] -> enabled rule with warning severity
// - ["error", {...options}] -> enabled rule with error severity and options
// - ["error", "both"] -> enabled rule with string option (e.g. no-inner-declarations)
// - ["error", "both", {...options}] -> enabled rule with string + object options
func parseArrayRuleConfig(ruleArray []interface{}) *RuleConfig {
	if len(ruleArray) == 0 {
		return nil
	}

	// First element should always be the severity level
	level, ok := ruleArray[0].(string)
	if !ok {
		return nil
	}

	ruleConfig := &RuleConfig{Level: level}

	// Remaining elements are rule options — pass them through to the rule's
	// option parser which knows how to interpret its own format.
	if len(ruleArray) > 1 {
		remaining := ruleArray[1:]
		if len(remaining) == 1 {
			// Single option element: pass directly (string, map, etc.)
			ruleConfig.Options = remaining[0]
		} else {
			// Multiple option elements: pass as array (e.g. ["both", {blockScopedFunctions: "disallow"}])
			ruleConfig.Options = remaining
		}
	}

	return ruleConfig
}

var registerOnce sync.Once

func RegisterAllRules() {
	registerOnce.Do(func() {
		registerAllTypeScriptEslintPluginRules()
		registerAllEslintImportPluginRules()
		registerAllReactPluginRules()
		registerAllReactHooksPluginRules()
		registerAllJestPluginRules()
		registerAllPromisePluginRules()
		registerAllUnicornPluginRules()
		registerAllCoreEslintRules()
	})
}

func registerAllReactPluginRules() {
	for _, rule := range reactPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllReactHooksPluginRules() {
	for _, rule := range reactHooksPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllJestPluginRules() {
	for _, rule := range jestPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllPromisePluginRules() {
	for _, rule := range promisePlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllUnicornPluginRules() {
	for _, rule := range unicornPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

// registerAllTypeScriptEslintPluginRules registers all available rules in the global registry
func registerAllTypeScriptEslintPluginRules() {
	GlobalRuleRegistry.Register("@typescript-eslint/adjacent-overload-signatures", adjacent_overload_signatures.AdjacentOverloadSignaturesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/array-type", array_type.ArrayTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/await-thenable", await_thenable.AwaitThenableRule)
	GlobalRuleRegistry.Register("@typescript-eslint/ban-ts-comment", ban_ts_comment.BanTsCommentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/ban-tslint-comment", ban_tslint_comment.BanTslintCommentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/ban-types", ban_types.BanTypesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/class-literal-property-style", class_literal_property_style.ClassLiteralPropertyStyleRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-generic-constructors", consistent_generic_constructors.ConsistentGenericConstructorsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-indexed-object-style", consistent_indexed_object_style.ConsistentIndexedObjectStyleRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-return", consistent_return.ConsistentReturnRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-type-assertions", consistent_type_assertions.ConsistentTypeAssertionsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-type-definitions", consistent_type_definitions.ConsistentTypeDefinitionsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-type-exports", consistent_type_exports.ConsistentTypeExportsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-type-imports", consistent_type_imports.ConsistentTypeImportsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/default-param-last", default_param_last.DefaultParamLastRule)
	GlobalRuleRegistry.Register("@typescript-eslint/dot-notation", dot_notation.DotNotationRule)
	GlobalRuleRegistry.Register("@typescript-eslint/explicit-function-return-type", explicit_function_return_type.ExplicitFunctionReturnTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/explicit-member-accessibility", explicit_member_accessibility.ExplicitMemberAccessibilityRule)
	GlobalRuleRegistry.Register("@typescript-eslint/member-ordering", member_ordering.MemberOrderingRule)
	GlobalRuleRegistry.Register("@typescript-eslint/method-signature-style", method_signature_style.MethodSignatureStyleRule)
	GlobalRuleRegistry.Register("@typescript-eslint/naming-convention", naming_convention.NamingConventionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-array-constructor", no_array_constructor.NoArrayConstructorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-array-delete", no_array_delete.NoArrayDeleteRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-base-to-string", no_base_to_string.NoBaseToStringRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-confusing-void-expression", no_confusing_void_expression.NoConfusingVoidExpressionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-dupe-class-members", ts_no_dupe_class_members.NoDupeClassMembersRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-duplicate-enum-values", no_duplicate_enum_values.NoDuplicateEnumValuesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-duplicate-type-constituents", no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-dynamic-delete", no_dynamic_delete.NoDynamicDeleteRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-explicit-any", no_explicit_any.NoExplicitAnyRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-empty-function", no_empty_function.NoEmptyFunctionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-empty-interface", no_empty_interface.NoEmptyInterfaceRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-extra-non-null-assertion", no_extra_non_null_assertion.NoExtraNonNullAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-extraneous-class", no_extraneous_class.NoExtraneousClassRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-invalid-void-type", no_invalid_void_type.NoInvalidVoidTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-floating-promises", no_floating_promises.NoFloatingPromisesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-for-in-array", no_for_in_array.NoForInArrayRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-implied-eval", no_implied_eval.NoImpliedEvalRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-inferrable-types", no_inferrable_types.NoInferrableTypesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-magic-numbers", no_magic_numbers.NoMagicNumbersRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-meaningless-void-operator", no_meaningless_void_operator.NoMeaninglessVoidOperatorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-new", no_misused_new.NoMisusedNewRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-promises", no_misused_promises.NoMisusedPromisesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-spread", no_misused_spread.NoMisusedSpreadRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-mixed-enums", no_mixed_enums.NoMixedEnumsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-namespace", no_namespace.NoNamespaceRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-non-null-asserted-nullish-coalescing", no_non_null_asserted_nullish_coalescing.NoNonNullAssertedNullishCoalescingRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-non-null-asserted-optional-chain", no_non_null_asserted_optional_chain.NoNonNullAssertedOptionalChainRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-non-null-assertion", no_non_null_assertion.NoNonNullAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-redeclare", no_redeclare.NoRedeclareRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-redundant-type-constituents", no_redundant_type_constituents.NoRedundantTypeConstituentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-this-alias", no_this_alias.NoThisAliasRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-require-imports", no_require_imports.NoRequireImportsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-restricted-imports", ts_no_restricted_imports.NoRestrictedImportsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-boolean-literal-compare", no_unnecessary_boolean_literal_compare.NoUnnecessaryBooleanLiteralCompareRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-condition", no_unnecessary_condition.NoUnnecessaryConditionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-template-expression", no_unnecessary_template_expression.NoUnnecessaryTemplateExpressionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-arguments", no_unnecessary_type_arguments.NoUnnecessaryTypeArgumentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-assertion", no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-constraint", no_unnecessary_type_constraint.NoUnnecessaryTypeConstraintRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-argument", no_unsafe_argument.NoUnsafeArgumentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-assignment", no_unsafe_assignment.NoUnsafeAssignmentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-call", no_unsafe_call.NoUnsafeCallRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-enum-comparison", no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-member-access", no_unsafe_member_access.NoUnsafeMemberAccessRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-return", no_unsafe_return.NoUnsafeReturnRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-type-assertion", no_unsafe_type_assertion.NoUnsafeTypeAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-unary-minus", no_unsafe_unary_minus.NoUnsafeUnaryMinusRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unused-expressions", no_unused_expressions.NoUnusedExpressionsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unused-vars", no_unused_vars.NoUnusedVarsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-use-before-define", no_use_before_define.NoUseBeforeDefineRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-useless-constructor", ts_no_useless_constructor.NoUselessConstructorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-useless-empty-export", no_useless_empty_export.NoUselessEmptyExportRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-var-requires", no_var_requires.NoVarRequiresRule)
	GlobalRuleRegistry.Register("@typescript-eslint/non-nullable-type-assertion-style", non_nullable_type_assertion_style.NonNullableTypeAssertionStyleRule)
	GlobalRuleRegistry.Register("@typescript-eslint/only-throw-error", only_throw_error.OnlyThrowErrorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/parameter-properties", parameter_properties.ParameterPropertiesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-as-const", prefer_as_const.PreferAsConstRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-destructuring", prefer_destructuring.PreferDestructuringRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-includes", prefer_includes.PreferIncludesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-literal-enum-member", prefer_literal_enum_member.PreferLiteralEnumMemberRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-namespace-keyword", prefer_namespace_keyword.PreferNamespaceKeywordRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-for-of", prefer_for_of.PreferForOfRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-optional-chain", prefer_optional_chain.PreferOptionalChainRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-promise-reject-errors", prefer_promise_reject_errors.PreferPromiseRejectErrorsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-readonly", prefer_readonly.PreferReadonlyRule)
	// TODO: prefer-readonly-parameter-types needs complete implementation for proper type checking
	// Temporarily disabled until the isReadonlyType function is fully implemented with proper
	// detection of readonly arrays, readonly objects, function types, and other edge cases
	// GlobalRuleRegistry.Register("@typescript-eslint/prefer-readonly-parameter-types", prefer_readonly_parameter_types.PreferReadonlyParameterTypesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-reduce-type-parameter", prefer_reduce_type_parameter.PreferReduceTypeParameterRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-regexp-exec", prefer_regexp_exec.PreferRegExpExecRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-return-this-type", prefer_return_this_type.PreferReturnThisTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-string-starts-ends-with", prefer_string_starts_ends_with.PreferStringStartsEndsWithRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-ts-expect-error", prefer_ts_expect_error.PreferTsExpectErrorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/promise-function-async", promise_function_async.PromiseFunctionAsyncRule)
	GlobalRuleRegistry.Register("@typescript-eslint/related-getter-setter-pairs", related_getter_setter_pairs.RelatedGetterSetterPairsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/require-array-sort-compare", require_array_sort_compare.RequireArraySortCompareRule)
	GlobalRuleRegistry.Register("@typescript-eslint/require-await", require_await.RequireAwaitRule)
	GlobalRuleRegistry.Register("@typescript-eslint/restrict-plus-operands", restrict_plus_operands.RestrictPlusOperandsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/restrict-template-expressions", restrict_template_expressions.RestrictTemplateExpressionsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/return-await", return_await.ReturnAwaitRule)
	GlobalRuleRegistry.Register("@typescript-eslint/switch-exhaustiveness-check", switch_exhaustiveness_check.SwitchExhaustivenessCheckRule)
	GlobalRuleRegistry.Register("@typescript-eslint/triple-slash-reference", triple_slash_reference.TripleSlashReferenceRule)
	GlobalRuleRegistry.Register("@typescript-eslint/unbound-method", unbound_method.UnboundMethodRule)
	GlobalRuleRegistry.Register("@typescript-eslint/unified-signatures", unified_signatures.UnifiedSignaturesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/use-unknown-in-catch-callback-variable", use_unknown_in_catch_callback_variable.UseUnknownInCatchCallbackVariableRule)
}

func registerAllEslintImportPluginRules() {
	for _, rule := range importPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

// registerAllCoreEslintRules registers core ESLint rules
func registerAllCoreEslintRules() {
	GlobalRuleRegistry.Register("accessor-pairs", accessor_pairs.AccessorPairsRule)
	GlobalRuleRegistry.Register("array-callback-return", array_callback_return.ArrayCallbackReturnRule)
	GlobalRuleRegistry.Register("constructor-super", constructor_super.ConstructorSuperRule)
	GlobalRuleRegistry.Register("default-case", default_case.DefaultCaseRule)
	GlobalRuleRegistry.Register("default-case-last", default_case_last.DefaultCaseLastRule)
	GlobalRuleRegistry.Register("for-direction", for_direction.ForDirectionRule)
	GlobalRuleRegistry.Register("getter-return", getter_return.GetterReturnRule)
	GlobalRuleRegistry.Register("guard-for-in", guard_for_in.GuardForInRule)
	GlobalRuleRegistry.Register("max-lines", max_lines.MaxLinesRule)
	GlobalRuleRegistry.Register("no-alert", no_alert.NoAlertRule)
	GlobalRuleRegistry.Register("no-async-promise-executor", no_async_promise_executor.NoAsyncPromiseExecutorRule)
	GlobalRuleRegistry.Register("no-await-in-loop", no_await_in_loop.NoAwaitInLoopRule)
	GlobalRuleRegistry.Register("no-bitwise", no_bitwise.NoBitwiseRule)
	GlobalRuleRegistry.Register("no-caller", no_caller.NoCallerRule)
	GlobalRuleRegistry.Register("no-case-declarations", no_case_declarations.NoCaseDeclarationsRule)
	GlobalRuleRegistry.Register("no-class-assign", no_class_assign.NoClassAssignRule)
	GlobalRuleRegistry.Register("no-compare-neg-zero", no_compare_neg_zero.NoCompareNegZeroRule)
	GlobalRuleRegistry.Register("no-cond-assign", no_cond_assign.NoCondAssignRule)
	GlobalRuleRegistry.Register("no-console", no_console.NoConsoleRule)
	GlobalRuleRegistry.Register("no-const-assign", no_const_assign.NoConstAssignRule)
	GlobalRuleRegistry.Register("no-constant-binary-expression", no_constant_binary_expression.NoConstantBinaryExpressionRule)
	GlobalRuleRegistry.Register("no-constant-condition", no_constant_condition.NoConstantConditionRule)
	GlobalRuleRegistry.Register("no-constructor-return", no_constructor_return.NoConstructorReturnRule)
	GlobalRuleRegistry.Register("no-control-regex", no_control_regex.NoControlRegexRule)
	GlobalRuleRegistry.Register("no-debugger", no_debugger.NoDebuggerRule)
	GlobalRuleRegistry.Register("no-delete-var", no_delete_var.NoDeleteVarRule)
	GlobalRuleRegistry.Register("no-dupe-args", no_dupe_args.NoDupeArgsRule)
	GlobalRuleRegistry.Register("no-dupe-class-members", no_dupe_class_members.NoDupeClassMembersRule)
	GlobalRuleRegistry.Register("no-dupe-keys", no_dupe_keys.NoDupeKeysRule)
	GlobalRuleRegistry.Register("no-duplicate-case", no_duplicate_case.NoDuplicateCaseRule)
	GlobalRuleRegistry.Register("no-empty", no_empty.NoEmptyRule)
	GlobalRuleRegistry.Register("no-empty-pattern", no_empty_pattern.NoEmptyPatternRule)
	GlobalRuleRegistry.Register("no-eval", no_eval.NoEvalRule)
	GlobalRuleRegistry.Register("no-ex-assign", no_ex_assign.NoExAssignRule)
	GlobalRuleRegistry.Register("no-extend-native", no_extend_native.NoExtendNativeRule)
	GlobalRuleRegistry.Register("no-extra-bind", no_extra_bind.NoExtraBindRule)
	GlobalRuleRegistry.Register("no-extra-label", no_extra_label.NoExtraLabelRule)
	GlobalRuleRegistry.Register("no-label-var", no_label_var.NoLabelVarRule)
	GlobalRuleRegistry.Register("no-labels", no_labels.NoLabelsRule)
	GlobalRuleRegistry.Register("no-func-assign", no_func_assign.NoFuncAssignRule)
	GlobalRuleRegistry.Register("no-global-assign", no_global_assign.NoGlobalAssignRule)
	GlobalRuleRegistry.Register("no-implicit-coercion", no_implicit_coercion.NoImplicitCoercionRule)
	GlobalRuleRegistry.Register("no-implied-eval", core_no_implied_eval.NoImpliedEvalRule)
	GlobalRuleRegistry.Register("no-import-assign", no_import_assign.NoImportAssignRule)
	GlobalRuleRegistry.Register("no-inner-declarations", no_inner_declarations.NoInnerDeclarationsRule)
	GlobalRuleRegistry.Register("no-irregular-whitespace", no_irregular_whitespace.NoIrregularWhitespaceRule)
	GlobalRuleRegistry.Register("no-lone-blocks", no_lone_blocks.NoLoneBlocksRule)
	GlobalRuleRegistry.Register("no-loop-func", no_loop_func.NoLoopFuncRule)
	GlobalRuleRegistry.Register("no-loss-of-precision", no_loss_of_precision.NoLossOfPrecisionRule)
	GlobalRuleRegistry.Register("no-misleading-character-class", no_misleading_character_class.NoMisleadingCharacterClassRule)
	GlobalRuleRegistry.Register("no-new", no_new.NoNewRule)
	GlobalRuleRegistry.Register("no-new-func", no_new_func.NoNewFuncRule)
	GlobalRuleRegistry.Register("no-new-wrappers", no_new_wrappers.NoNewWrappersRule)
	GlobalRuleRegistry.Register("no-restricted-imports", no_restricted_imports.NoRestrictedImportsRule)
	GlobalRuleRegistry.Register("no-multi-str", no_multi_str.NoMultiStrRule)
	GlobalRuleRegistry.Register("no-nested-ternary", no_nested_ternary.NoNestedTernaryRule)
	GlobalRuleRegistry.Register("no-octal", no_octal.NoOctalRule)
	GlobalRuleRegistry.Register("no-octal-escape", no_octal_escape.NoOctalEscapeRule)
	GlobalRuleRegistry.Register("no-param-reassign", no_param_reassign.NoParamReassignRule)
	GlobalRuleRegistry.Register("no-proto", no_proto.NoProtoRule)
	GlobalRuleRegistry.Register("radix", radix.RadixRule)
	GlobalRuleRegistry.Register("no-regex-spaces", no_regex_spaces.NoRegexSpacesRule)
	GlobalRuleRegistry.Register("no-return-assign", no_return_assign.NoReturnAssignRule)
	GlobalRuleRegistry.Register("no-script-url", no_script_url.NoScriptUrlRule)
	GlobalRuleRegistry.Register("no-self-assign", no_self_assign.NoSelfAssignRule)
	GlobalRuleRegistry.Register("no-self-compare", no_self_compare.NoSelfCompareRule)
	GlobalRuleRegistry.Register("no-sequences", no_sequences.NoSequencesRule)
	GlobalRuleRegistry.Register("no-shadow", no_shadow.NoShadowRule)
	GlobalRuleRegistry.Register("no-shadow-restricted-names", no_shadow_restricted_names.NoShadowRestrictedNamesRule)
	GlobalRuleRegistry.Register("strict", strict.StrictRule)
	GlobalRuleRegistry.Register("no-template-curly-in-string", no_template_curly_in_string.NoTemplateCurlyInString)
	GlobalRuleRegistry.Register("no-useless-computed-key", no_useless_computed_key.NoUselessComputedKeyRule)
	GlobalRuleRegistry.Register("no-useless-concat", no_useless_concat.NoUselessConcatRule)
	GlobalRuleRegistry.Register("no-sparse-arrays", no_sparse_arrays.NoSparseArraysRule)
	GlobalRuleRegistry.Register("no-extra-boolean-cast", no_extra_boolean_cast.NoExtraBooleanCastRule)
	GlobalRuleRegistry.Register("no-unneeded-ternary", no_unneeded_ternary.NoUnneededTernaryRule)
	GlobalRuleRegistry.Register("no-undef", no_undef.NoUndefRule)
	GlobalRuleRegistry.Register("no-undef-init", no_undef_init.NoUndefInitRule)
	GlobalRuleRegistry.Register("prefer-const", prefer_const.PreferConstRule)
	GlobalRuleRegistry.Register("prefer-promise-reject-errors", core_prefer_promise_reject_errors.PreferPromiseRejectErrorsRule)
	GlobalRuleRegistry.Register("prefer-template", prefer_template.PreferTemplateRule)
	GlobalRuleRegistry.Register("no-this-before-super", no_this_before_super.NoThisBeforeSuperRule)
	GlobalRuleRegistry.Register("no-var", no_var.NoVarRule)
	GlobalRuleRegistry.Register("no-with", no_with.NoWithRule)
	GlobalRuleRegistry.Register("prefer-rest-params", prefer_rest_params.PreferRestParamsRule)
	GlobalRuleRegistry.Register("prefer-spread", prefer_spread.PreferSpreadRule)
	GlobalRuleRegistry.Register("no-empty-character-class", no_empty_character_class.NoEmptyCharacterClassRule)
	GlobalRuleRegistry.Register("no-invalid-regexp", no_invalid_regexp.NoInvalidRegexpRule)
	GlobalRuleRegistry.Register("no-iterator", no_iterator.NoIteratorRule)
	GlobalRuleRegistry.Register("no-setter-return", no_setter_return.NoSetterReturnRule)
	GlobalRuleRegistry.Register("no-unsafe-negation", no_unsafe_negation.NoUnsafeNegationRule)
	GlobalRuleRegistry.Register("no-obj-calls", no_obj_calls.NoObjCallsRule)
	GlobalRuleRegistry.Register("no-new-object", no_new_object.NoNewObjectRule)
	GlobalRuleRegistry.Register("no-new-symbol", no_new_symbol.NoNewSymbolRule)
	GlobalRuleRegistry.Register("use-isnan", use_isnan.UseIsNaNRule)
	GlobalRuleRegistry.Register("eqeqeq", eqeqeq.EqeqeqRule)
	GlobalRuleRegistry.Register("no-fallthrough", no_fallthrough.NoFallthroughRule)
	GlobalRuleRegistry.Register("valid-typeof", valid_typeof.ValidTypeofRule)
	GlobalRuleRegistry.Register("no-unsafe-optional-chaining", no_unsafe_optional_chaining.NoUnsafeOptionalChainingRule)
	GlobalRuleRegistry.Register("no-unsafe-finally", no_unsafe_finally.NoUnsafeFinallyRule)
	GlobalRuleRegistry.Register("no-unmodified-loop-condition", no_unmodified_loop_condition.NoUnmodifiedLoopConditionRule)
	GlobalRuleRegistry.Register("no-unreachable", no_unreachable.NoUnreachableRule)
	GlobalRuleRegistry.Register("require-atomic-updates", require_atomic_updates.RequireAtomicUpdatesRule)
	GlobalRuleRegistry.Register("object-shorthand", object_shorthand.ObjectShorthandRule)
	GlobalRuleRegistry.Register("no-dupe-else-if", no_dupe_else_if.NoDupeElseIfRule)
	GlobalRuleRegistry.Register("no-throw-literal", no_throw_literal.NoThrowLiteralRule)
	GlobalRuleRegistry.Register("no-useless-call", no_useless_call.NoUselessCallRule)
	GlobalRuleRegistry.Register("no-useless-catch", no_useless_catch.NoUselessCatchRule)
	GlobalRuleRegistry.Register("no-useless-rename", no_useless_rename.NoUselessRenameRule)
	GlobalRuleRegistry.Register("no-useless-constructor", no_useless_constructor.NoUselessConstructorRule)
	GlobalRuleRegistry.Register("no-prototype-builtins", no_prototype_builtins.NoPrototypeBuiltinsRule)
	GlobalRuleRegistry.Register("require-yield", require_yield.RequireYieldRule)
	GlobalRuleRegistry.Register("symbol-description", symbol_description.SymbolDescriptionRule)
}

// isFileIgnored checks if a file is matched by ignore patterns, evaluated sequentially.
// Later patterns override earlier ones; a `!` prefix negates (re-includes) a previously
// ignored file. This aligns with ESLint v10's ignore semantics.
//
// For directory-level blocking (dir/** prevents traversal entirely), use isDirPathBlocked.
func isFileIgnored(filePath string, ignorePatterns []string, cwd string) bool {
	if cwd == "" {
		return isFileIgnoredSimple(filePath, ignorePatterns)
	}

	// Normalize the file path relative to cwd
	normalizedPath := normalizePath(filePath, cwd)
	unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")

	// Evaluate patterns sequentially. Later patterns override earlier ones.
	// A `!` prefix negates (re-includes) a previously ignored file.
	// This aligns with ESLint v10's ignore semantics.
	ignored := false
	for _, pattern := range ignorePatterns {
		negated := false
		if strings.HasPrefix(pattern, "!") {
			negated = true
			pattern = pattern[1:]
		}

		normalizedPattern := normalizePattern(pattern)

		// Match against the relative path only. Do NOT fall back to the
		// absolute filePath — patterns with **/ prefix (e.g., **/tmp/**/*)
		// would incorrectly match system directory names in the absolute path
		// (e.g., /tmp/ on Linux/macOS).
		matched := matchGlob(normalizedPattern, normalizedPath)
		// Windows path separator fallback.
		if !matched && unixPath != normalizedPath {
			matched = matchGlob(normalizedPattern, unixPath)
		}

		if matched {
			ignored = !negated
		}
	}
	return ignored
}

// normalizePattern cleans up a glob pattern to match paths produced by normalizePath.
// normalizePath uses tspath.NormalizePath on file paths (strips leading "./", collapses
// "/./", resolves ".."), so patterns must undergo the same transformation.
// matchGlob matches a glob pattern against a path using doublestar.
func matchGlob(pattern, path string) bool {
	m, err := doublestar.Match(pattern, path)
	return err == nil && m
}

// isFileLevelPattern returns true if the pattern only matches files (not directories).
// File-level patterns end with /**/* or /* (but not /**).
// These do NOT block directory traversal in ESLint v10's isDirectoryIgnored.
func isFileLevelPattern(pattern string) bool {
	return strings.HasSuffix(pattern, "/**/*") ||
		(strings.HasSuffix(pattern, "/*") && !strings.HasSuffix(pattern, "/**"))
}

func normalizePattern(pattern string) string {
	return tspath.NormalizePath(pattern)
}

// isDirBlockedByIgnores checks if the file's directory is blocked by a
// directory-level ignore pattern (e.g., `dir/**`). File-level patterns
// (`dir/**/*`, `dir/*`) and negation patterns are skipped.
// This aligns with ESLint v10: `dir/**` blocks directory traversal entirely,
// and `!` negation cannot undo it.
func isDirBlockedByIgnores(filePath string, ignorePatterns []string, cwd string) bool {
	var dirPath string
	if cwd != "" {
		dirPath = normalizePath(tspath.GetDirectoryPath(filePath), cwd)
	} else {
		dirPath = tspath.GetDirectoryPath(filePath)
	}
	dirPath = strings.ReplaceAll(dirPath, "\\", "/")
	dirPath = strings.TrimSuffix(dirPath, "/")
	if dirPath == "" || dirPath == "." {
		return false
	}
	return isDirPathBlocked(dirPath, ignorePatterns)
}

// isDirPathBlocked checks if a directory path is blocked by any directory-level ignore
// pattern. Shared between GetConfigForFile and DiscoverGapFiles.
//
// A directory is blocked if a pattern matches the path itself or any parent segment.
// For example, pattern "dir1/**" blocks "dir1", "dir1/sub", and "dir1/sub/deep".
// File-level patterns (ending with /**/* or /*) and negation (!) patterns are skipped —
// directory blocking is absolute and cannot be negated.
func isDirPathBlocked(dirPath string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		if pattern == "" || strings.HasPrefix(pattern, "!") {
			continue
		}
		if isFileLevelPattern(pattern) {
			continue
		}

		normalizedPattern := normalizePattern(pattern)

		if matchGlob(normalizedPattern, dirPath) || matchGlob(normalizedPattern, dirPath+"/x") {
			return true
		}
		segments := strings.Split(dirPath, "/")
		for i := 1; i < len(segments); i++ {
			partial := strings.Join(segments[:i], "/")
			if matchGlob(normalizedPattern, partial) || matchGlob(normalizedPattern, partial+"/x") {
				return true
			}
		}
	}
	return false
}

// normalizePath converts file path to be relative to cwd for consistent matching
func normalizePath(filePath, cwd string) string {
	return tspath.NormalizePath(tspath.ConvertToRelativePath(filePath, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: true,
		CurrentDirectory:          cwd,
	}))
}

// isFileIgnoredSimple provides fallback matching when cwd is unavailable
func isFileIgnoredSimple(filePath string, ignorePatterns []string) bool {
	ignored := false
	for _, pattern := range ignorePatterns {
		negated := false
		if strings.HasPrefix(pattern, "!") {
			negated = true
			pattern = pattern[1:]
		}
		normalizedPattern := normalizePattern(pattern)
		if matched, err := doublestar.Match(normalizedPattern, filePath); err == nil && matched {
			ignored = !negated
		}
	}
	return ignored
}

// MergedConfig is the final computed configuration for a single file
type MergedConfig struct {
	Rules           map[string]*RuleConfig
	Settings        Settings
	LanguageOptions *LanguageOptions
	Plugins         map[string]struct{}
}

// IsFileIgnored reports whether filePath is excluded by the config's global
// `ignores` patterns. It is distinct from GetConfigForFile returning nil,
// which also covers "no entry matched this file" — callers that need ESLint's
// "ignores hides the file from the linter entirely" semantics (including
// type-check diagnostics and file counts) should use this method.
func (config RslintConfig) IsFileIgnored(filePath string, cwd string) bool {
	patterns := ExtractConfigIgnores(config)
	if len(patterns) == 0 {
		return false
	}
	return isDirBlockedByIgnores(filePath, patterns, cwd) ||
		isFileIgnored(filePath, patterns, cwd)
}

// GetConfigForFile computes the merged configuration for a file following ESLint flat config semantics.
// Returns nil if the file is globally ignored or no entry matches (should not be linted).
//
// Global ignore evaluation happens in two phases:
//  1. Directory-level (isDirBlockedByIgnores): patterns like dir/** block entire directories.
//     Negation (!) cannot override directory-level blocking.
//  2. File-level (isFileIgnored): sequential evaluation with ! negation support for re-inclusion.
//
// After global ignore check, entries are merged in order if their files match and ignores don't.
// cwd is the directory the config lives in; file paths are resolved relative to it.
func (config RslintConfig) GetConfigForFile(filePath string, cwd string) *MergedConfig {
	merged := &MergedConfig{
		Rules:   make(map[string]*RuleConfig),
		Plugins: make(map[string]struct{}),
	}

	// 1. Collect all global ignore patterns and evaluate once.
	// This allows `!` negation patterns in separate entries to work correctly,
	// aligned with ESLint v10 which merges all global ignores before evaluating.
	globalIgnorePatterns := ExtractConfigIgnores(config)
	if len(globalIgnorePatterns) > 0 {
		// Phase 1: directory-level check. Patterns like `dir/**` block the
		// directory entirely — `!` negation cannot undo this. Aligned with
		// ESLint v10's isDirectoryIgnored behavior.
		if isDirBlockedByIgnores(filePath, globalIgnorePatterns, cwd) {
			return nil
		}
		// Phase 2: file-level check with sequential `!` negation support.
		if isFileIgnored(filePath, globalIgnorePatterns, cwd) {
			return nil
		}
	}

	// Track whether any non-global entry matched this file
	entryMatched := false

	for _, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			continue
		}

		// 2. files matching
		if len(entry.Files) > 0 && !isFileMatched(filePath, entry.Files, cwd) {
			continue
		}

		// 3. Entry-level ignores
		if isFileIgnored(filePath, entry.Ignores, cwd) {
			continue
		}

		entryMatched = true

		// 4. Rules: shallow merge, later entries override earlier ones
		for ruleName, ruleValue := range entry.Rules {
			switch v := ruleValue.(type) {
			case string:
				merged.Rules[ruleName] = &RuleConfig{Level: v}
			case []interface{}:
				if rc := parseArrayRuleConfig(v); rc != nil {
					merged.Rules[ruleName] = rc
				}
			case map[string]interface{}:
				ruleConfig := &RuleConfig{}
				if level, ok := v["level"].(string); ok {
					ruleConfig.Level = level
				}
				if options, ok := v["options"].(map[string]interface{}); ok {
					ruleConfig.Options = options
				}
				merged.Rules[ruleName] = ruleConfig
			}
		}

		// 5. Plugins: union from all matching entries (normalized to rule prefix form)
		for _, plugin := range entry.Plugins {
			merged.Plugins[NormalizePluginName(plugin)] = struct{}{}
		}

		// 6. Settings: shallow merge
		if entry.Settings != nil {
			if merged.Settings == nil {
				merged.Settings = make(Settings)
			}
			for k, v := range entry.Settings {
				merged.Settings[k] = v
			}
		}

		// 7. LanguageOptions: deep merge
		merged.LanguageOptions = mergeLanguageOptions(merged.LanguageOptions, entry.LanguageOptions)
	}

	// No entry matched this file — do not lint it
	if !entryMatched {
		return nil
	}

	return merged
}

// isGlobalIgnoreEntry returns true if the entry is a global ignore entry
// (has only ignores, no other fields).
func isGlobalIgnoreEntry(entry ConfigEntry) bool {
	return len(entry.Files) == 0 &&
		len(entry.Rules) == 0 &&
		len(entry.Plugins) == 0 &&
		entry.Settings == nil &&
		entry.LanguageOptions == nil &&
		len(entry.Ignores) > 0
}

// isFileMatched checks if a file matches any of the given glob patterns
func isFileMatched(filePath string, patterns []string, cwd string) bool {
	var normalizedPath string
	if cwd != "" {
		normalizedPath = normalizePath(filePath, cwd)
	} else {
		normalizedPath = filePath
	}

	for _, pattern := range patterns {
		normalizedPattern := normalizePattern(pattern)

		if matched, err := doublestar.Match(normalizedPattern, normalizedPath); err == nil && matched {
			return true
		}
		if normalizedPath != filePath {
			if matched, err := doublestar.Match(normalizedPattern, filePath); err == nil && matched {
				return true
			}
		}
		unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")
		if unixPath != normalizedPath {
			if matched, err := doublestar.Match(normalizedPattern, unixPath); err == nil && matched {
				return true
			}
		}
	}
	return false
}

// mergeLanguageOptions deep-merges two LanguageOptions, with override taking precedence
func mergeLanguageOptions(base, override *LanguageOptions) *LanguageOptions {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}
	merged := *base
	if override.ParserOptions != nil {
		if merged.ParserOptions == nil {
			merged.ParserOptions = override.ParserOptions
		} else {
			po := *merged.ParserOptions
			if override.ParserOptions.ProjectService != nil {
				po.ProjectService = override.ParserOptions.ProjectService
			}
			if len(override.ParserOptions.Project) > 0 {
				po.Project = override.ParserOptions.Project
			}
			merged.ParserOptions = &po
		}
	}
	return &merged
}

// RulePluginPrefix extracts the plugin prefix from a rule name.
// "@typescript-eslint/no-explicit-any" → "@typescript-eslint"
// "import/no-unresolved" → "import"
// "no-debugger" → "" (core rule)
func RulePluginPrefix(ruleName string) string {
	lastSlash := strings.LastIndex(ruleName, "/")
	if lastSlash < 0 {
		return ""
	}
	return ruleName[:lastSlash]
}

// GetPluginRules returns only rules under the given plugin namespace (prefix match).
func GetPluginRules(pluginName string) []rule.Rule {
	prefix := pluginName + "/"
	var rules []rule.Rule
	for name, r := range GlobalRuleRegistry.GetAllRules() {
		if strings.HasPrefix(name, prefix) {
			rules = append(rules, r)
		}
	}
	return rules
}

// GetCoreRules returns core ESLint rules (those without a "/" prefix).
func GetCoreRules() []rule.Rule {
	var rules []rule.Rule
	for name, r := range GlobalRuleRegistry.GetAllRules() {
		if !strings.Contains(name, "/") {
			rules = append(rules, r)
		}
	}
	return rules
}

// InitDefaultConfig, createDefaultConfig, migrateJSONConfig and related helpers
// are in config_init.go.
