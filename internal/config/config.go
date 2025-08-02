package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/rules/adjacent_overload_signatures"
	"github.com/typescript-eslint/rslint/internal/rules/array_type"
	"github.com/typescript-eslint/rslint/internal/rules/await_thenable"
	"github.com/typescript-eslint/rslint/internal/rules/ban_ts_comment"
	"github.com/typescript-eslint/rslint/internal/rules/ban_tslint_comment"
	"github.com/typescript-eslint/rslint/internal/rules/class_literal_property_style"
	"github.com/typescript-eslint/rslint/internal/rules/class_methods_use_this"
	"github.com/typescript-eslint/rslint/internal/rules/consistent_generic_constructors"
	"github.com/typescript-eslint/rslint/internal/rules/consistent_indexed_object_style"
	"github.com/typescript-eslint/rslint/internal/rules/consistent_return"
	"github.com/typescript-eslint/rslint/internal/rules/consistent_type_assertions"
	"github.com/typescript-eslint/rslint/internal/rules/consistent_type_definitions"
	"github.com/typescript-eslint/rslint/internal/rules/consistent_type_exports"
	"github.com/typescript-eslint/rslint/internal/rules/consistent_type_imports"
	"github.com/typescript-eslint/rslint/internal/rules/default_param_last"
	"github.com/typescript-eslint/rslint/internal/rules/explicit_member_accessibility"
	"github.com/typescript-eslint/rslint/internal/rules/explicit_module_boundary_types"
	"github.com/typescript-eslint/rslint/internal/rules/init_declarations"
	"github.com/typescript-eslint/rslint/internal/rules/no_array_delete"
	"github.com/typescript-eslint/rslint/internal/rules/no_base_to_string"
	"github.com/typescript-eslint/rslint/internal/rules/no_confusing_non_null_assertion"
	"github.com/typescript-eslint/rslint/internal/rules/no_confusing_void_expression"
	"github.com/typescript-eslint/rslint/internal/rules/no_duplicate_type_constituents"
	"github.com/typescript-eslint/rslint/internal/rules/no_duplicate_enum_values"
	"github.com/typescript-eslint/rslint/internal/rules/no_dupe_class_members"
	"github.com/typescript-eslint/rslint/internal/rules/no_dynamic_delete"
	"github.com/typescript-eslint/rslint/internal/rules/no_empty_function"
	"github.com/typescript-eslint/rslint/internal/rules/no_empty_interface"
	"github.com/typescript-eslint/rslint/internal/rules/no_empty_object_type"
	"github.com/typescript-eslint/rslint/internal/rules/no_floating_promises"
	"github.com/typescript-eslint/rslint/internal/rules/no_import_type_side_effects"
	"github.com/typescript-eslint/rslint/internal/rules/no_for_in_array"
	"github.com/typescript-eslint/rslint/internal/rules/no_implied_eval"
	"github.com/typescript-eslint/rslint/internal/rules/no_inferrable_types"
	"github.com/typescript-eslint/rslint/internal/rules/no_invalid_this"
	"github.com/typescript-eslint/rslint/internal/rules/no_invalid_void_type"
	"github.com/typescript-eslint/rslint/internal/rules/no_loop_func"
	"github.com/typescript-eslint/rslint/internal/rules/no_loss_of_precision"
	"github.com/typescript-eslint/rslint/internal/rules/no_magic_numbers"
	"github.com/typescript-eslint/rslint/internal/rules/max_params"
	"github.com/typescript-eslint/rslint/internal/rules/member_ordering"
	"github.com/typescript-eslint/rslint/internal/rules/no_meaningless_void_operator"
	"github.com/typescript-eslint/rslint/internal/rules/no_misused_new"
	"github.com/typescript-eslint/rslint/internal/rules/no_misused_promises"
	"github.com/typescript-eslint/rslint/internal/rules/no_misused_spread"
	"github.com/typescript-eslint/rslint/internal/rules/no_mixed_enums"
	"github.com/typescript-eslint/rslint/internal/rules/no_namespace"
	"github.com/typescript-eslint/rslint/internal/rules/no_non_null_asserted_nullish_coalescing"
	"github.com/typescript-eslint/rslint/internal/rules/no_non_null_asserted_optional_chain"
	"github.com/typescript-eslint/rslint/internal/rules/no_non_null_assertion"
	"github.com/typescript-eslint/rslint/internal/rules/no_redeclare"
	"github.com/typescript-eslint/rslint/internal/rules/no_redundant_type_constituents"
	"github.com/typescript-eslint/rslint/internal/rules/no_require_imports"
	"github.com/typescript-eslint/rslint/internal/rules/no_restricted_imports"
	"github.com/typescript-eslint/rslint/internal/rules/no_restricted_types"
	"github.com/typescript-eslint/rslint/internal/rules/no_shadow"
	"github.com/typescript-eslint/rslint/internal/rules/no_this_alias"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_declaration_merging"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_boolean_literal_compare"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_template_expression"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_type_arguments"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_type_assertion"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_type_constraint"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_type_conversion"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_type_parameters"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_argument"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_assignment"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_call"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_enum_comparison"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_function_type"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_member_access"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_return"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_type_assertion"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_unary_minus"
	"github.com/typescript-eslint/rslint/internal/rules/no_unused_expressions"
	"github.com/typescript-eslint/rslint/internal/rules/no_unused_vars"
	"github.com/typescript-eslint/rslint/internal/rules/no_use_before_define"
	"github.com/typescript-eslint/rslint/internal/rules/no_useless_constructor"
	"github.com/typescript-eslint/rslint/internal/rules/no_useless_empty_export"
	"github.com/typescript-eslint/rslint/internal/rules/non_nullable_type_assertion_style"
	"github.com/typescript-eslint/rslint/internal/rules/only_throw_error"
	"github.com/typescript-eslint/rslint/internal/rules/prefer_as_const"
	"github.com/typescript-eslint/rslint/internal/rules/prefer_promise_reject_errors"
	"github.com/typescript-eslint/rslint/internal/rules/prefer_reduce_type_parameter"
	"github.com/typescript-eslint/rslint/internal/rules/prefer_return_this_type"
	"github.com/typescript-eslint/rslint/internal/rules/promise_function_async"
	"github.com/typescript-eslint/rslint/internal/rules/related_getter_setter_pairs"
	"github.com/typescript-eslint/rslint/internal/rules/require_array_sort_compare"
	"github.com/typescript-eslint/rslint/internal/rules/require_await"
	"github.com/typescript-eslint/rslint/internal/rules/restrict_plus_operands"
	"github.com/typescript-eslint/rslint/internal/rules/restrict_template_expressions"
	"github.com/typescript-eslint/rslint/internal/rules/return_await"
	"github.com/typescript-eslint/rslint/internal/rules/switch_exhaustiveness_check"
	"github.com/typescript-eslint/rslint/internal/rules/unbound_method"
	"github.com/typescript-eslint/rslint/internal/rules/use_unknown_in_catch_callback_variable"
)

// RslintConfig represents the top-level configuration array
type RslintConfig []ConfigEntry

// ConfigEntry represents a single configuration entry in the rslint.json array
type ConfigEntry struct {
	Language        string           `json:"language"`
	Files           []string         `json:"files"`
	Ignores         []string         `json:"ignores,omitempty"` // List of file patterns to ignore
	LanguageOptions *LanguageOptions `json:"languageOptions,omitempty"`
	Rules           Rules            `json:"rules"`
	Plugins         []string         `json:"plugins,omitempty"` // List of plugin names
}

// LanguageOptions contains language-specific configuration options
type LanguageOptions struct {
	ParserOptions *ParserOptions `json:"parserOptions,omitempty"`
}

// ParserOptions contains parser-specific configuration
type ParserOptions struct {
	ProjectService bool     `json:"projectService"`
	Project        []string `json:"project,omitempty"`
}

// Rules represents the rules configuration
// This can be extended to include specific rule configurations
type Rules map[string]interface{}

// Alternative: If you want type-safe rule configurations
type TypedRules struct {
	// Example rule configurations - extend as needed
	NoArrayDelete                      *RuleConfig `json:"@typescript-eslint/no-array-delete,omitempty"`
	NoBaseToString                     *RuleConfig `json:"@typescript-eslint/no-base-to-string,omitempty"`
	NoForInArray                       *RuleConfig `json:"@typescript-eslint/no-for-in-array,omitempty"`
	NoImpliedEval                      *RuleConfig `json:"@typescript-eslint/no-implied-eval,omitempty"`
	OnlyThrowError                     *RuleConfig `json:"@typescript-eslint/only-throw-error,omitempty"`
	AwaitThenable                      *RuleConfig `json:"@typescript-eslint/await-thenable,omitempty"`
	NoConfusingVoidExpression          *RuleConfig `json:"@typescript-eslint/no-confusing-void-expression,omitempty"`
	NoDuplicateTypeConstituents        *RuleConfig `json:"@typescript-eslint/no-duplicate-type-constituents,omitempty"`
	NoFloatingPromises                 *RuleConfig `json:"@typescript-eslint/no-floating-promises,omitempty"`
	NoMeaninglessVoidOperator          *RuleConfig `json:"@typescript-eslint/no-meaningless-void-operator,omitempty"`
	NoMisusedPromises                  *RuleConfig `json:"@typescript-eslint/no-misused-promises,omitempty"`
	NoMisusedSpread                    *RuleConfig `json:"@typescript-eslint/no-misused-spread,omitempty"`
	NoMixedEnums                       *RuleConfig `json:"@typescript-eslint/no-mixed-enums,omitempty"`
	NoRedundantTypeConstituents        *RuleConfig `json:"@typescript-eslint/no-redundant-type-constituents,omitempty"`
	NoUnnecessaryBooleanLiteralCompare *RuleConfig `json:"@typescript-eslint/no-unnecessary-boolean-literal-compare,omitempty"`
	NoUnnecessaryTemplateExpression    *RuleConfig `json:"@typescript-eslint/no-unnecessary-template-expression,omitempty"`
	NoUnnecessaryTypeArguments         *RuleConfig `json:"@typescript-eslint/no-unnecessary-type-arguments,omitempty"`
	NoUnnecessaryTypeAssertion         *RuleConfig `json:"@typescript-eslint/no-unnecessary-type-assertion,omitempty"`
	NoUnsafeArgument                   *RuleConfig `json:"@typescript-eslint/no-unsafe-argument,omitempty"`
	NoUnsafeAssignment                 *RuleConfig `json:"@typescript-eslint/no-unsafe-assignment,omitempty"`
	NoUnsafeCall                       *RuleConfig `json:"@typescript-eslint/no-unsafe-call,omitempty"`
	NoUnsafeEnumComparison             *RuleConfig `json:"@typescript-eslint/no-unsafe-enum-comparison,omitempty"`
	NoUnsafeMemberAccess               *RuleConfig `json:"@typescript-eslint/no-unsafe-member-access,omitempty"`
	NoUnsafeReturn                     *RuleConfig `json:"@typescript-eslint/no-unsafe-return,omitempty"`
	NoUnsafeTypeAssertion              *RuleConfig `json:"@typescript-eslint/no-unsafe-type-assertion,omitempty"`
	NoUnsafeUnaryMinus                 *RuleConfig `json:"@typescript-eslint/no-unsafe-unary-minus,omitempty"`
}

// RuleConfig represents individual rule configuration
type RuleConfig struct {
	Level   string                 `json:"level,omitempty"`   // "error", "warn", "off"
	Options map[string]interface{} `json:"options,omitempty"` // Rule-specific options
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
func (rc *RuleConfig) GetOptions() map[string]interface{} {
	if rc == nil || rc.Options == nil {
		return make(map[string]interface{})
	}
	return rc.Options
}

// SetOptions sets the rule options
func (rc *RuleConfig) SetOptions(options map[string]interface{}) {
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

func GetAllRulesForPlugin(plugin string) []rule.Rule {
	if plugin == "@typescript-eslint" {
		return getAllTypeScriptEslintPluginRules()
	} else {
		return []rule.Rule{} // Return empty slice for unsupported plugins
	}
}

// parseArrayRuleConfig parses array-style rule configuration like ["error", {...options}]
// Supports ESLint-compatible formats:
// - ["off"] -> disabled rule
// - ["error"] -> enabled rule with error severity
// - ["warn"] -> enabled rule with warning severity
// - ["error", {...options}] -> enabled rule with error severity and options
// - ["warn", {...options}] -> enabled rule with warning severity and options
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

	// Second element (if present) should be the options object
	if len(ruleArray) > 1 {
		switch opts := ruleArray[1].(type) {
		case map[string]interface{}:
			ruleConfig.Options = opts
		case nil:
			// Explicitly null/nil options are valid
			ruleConfig.Options = make(map[string]interface{})
		default:
			// Invalid options type, but still create the rule config with just the level
			ruleConfig.Options = make(map[string]interface{})
		}
	}

	// Additional elements are ignored (following ESLint behavior)
	return ruleConfig
}

// GetRulesForFile returns enabled rules for a given file based on the configuration
func (config RslintConfig) GetRulesForFile(filePath string) map[string]*RuleConfig {
	enabledRules := make(map[string]*RuleConfig)

	for _, entry := range config {
		// First check if the file should be ignored
		if isFileIgnored(filePath, entry.Ignores) {
			continue // Skip this config entry for ignored files
		}

		// Check if the file matches the files pattern
		matches := true

		if matches {

			/// Make rules from plugin available (but don't auto-enable them)
			// Plugin specification should only make rules available, not automatically enable them.
			// Rules are only enabled if explicitly configured in the "rules" section.
			// Merge rules from this entry
			for ruleName, ruleValue := range entry.Rules {

				switch v := ruleValue.(type) {
				case string:
					// Handle simple string values like "error", "warn", "off"
					enabledRules[ruleName] = &RuleConfig{Level: v}
				case map[string]interface{}:
					// Handle object configuration
					ruleConfig := &RuleConfig{}
					if level, ok := v["level"].(string); ok {
						ruleConfig.Level = level
					}
					if options, ok := v["options"].(map[string]interface{}); ok {
						ruleConfig.Options = options
					}
					if ruleConfig.IsEnabled() {
						enabledRules[ruleName] = ruleConfig
					}
				case []interface{}:
					// Handle array format like ["error", {...options}] or ["warn"] or ["off"]
					ruleConfig := parseArrayRuleConfig(v)
					if ruleConfig != nil && ruleConfig.IsEnabled() {
						enabledRules[ruleName] = ruleConfig
					}
				}
			}
		}
	}
	return enabledRules
}

// RegisterAllTypeSriptEslintPluginRules registers all available rules in the global registry
func RegisterAllTypeSriptEslintPluginRules() {
	GlobalRuleRegistry.Register("@typescript-eslint/adjacent-overload-signatures", adjacent_overload_signatures.AdjacentOverloadSignaturesRule)
	GlobalRuleRegistry.Register("adjacent-overload-signatures", adjacent_overload_signatures.AdjacentOverloadSignaturesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/array-type", array_type.ArrayTypeRule)
	GlobalRuleRegistry.Register("array-type", array_type.ArrayTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/await-thenable", await_thenable.AwaitThenableRule)
	GlobalRuleRegistry.Register("await-thenable", await_thenable.AwaitThenableRule)
	GlobalRuleRegistry.Register("@typescript-eslint/ban-ts-comment", ban_ts_comment.BanTsCommentRule)
	GlobalRuleRegistry.Register("ban-ts-comment", ban_ts_comment.BanTsCommentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/ts-expect-error", ban_ts_comment.BanTsCommentRule)
	GlobalRuleRegistry.Register("ts-expect-error", ban_ts_comment.BanTsCommentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/ban-tslint-comment", ban_tslint_comment.BanTslintCommentRule)
	GlobalRuleRegistry.Register("ban-tslint-comment", ban_tslint_comment.BanTslintCommentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/class-literal-property-style", class_literal_property_style.ClassLiteralPropertyStyleRule)
	GlobalRuleRegistry.Register("class-literal-property-style", class_literal_property_style.ClassLiteralPropertyStyleRule)
	GlobalRuleRegistry.Register("@typescript-eslint/class-methods-use-this", class_methods_use_this.ClassMethodsUseThisRule)
	GlobalRuleRegistry.Register("class-methods-use-this", class_methods_use_this.ClassMethodsUseThisRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-generic-constructors", consistent_generic_constructors.ConsistentGenericConstructorsRule)
	GlobalRuleRegistry.Register("consistent-generic-constructors", consistent_generic_constructors.ConsistentGenericConstructorsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-indexed-object-style", consistent_indexed_object_style.ConsistentIndexedObjectStyleRule)
	GlobalRuleRegistry.Register("consistent-indexed-object-style", consistent_indexed_object_style.ConsistentIndexedObjectStyleRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-return", consistent_return.ConsistentReturnRule)
	GlobalRuleRegistry.Register("consistent-return", consistent_return.ConsistentReturnRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-type-assertions", consistent_type_assertions.ConsistentTypeAssertionsRule)
	GlobalRuleRegistry.Register("consistent-type-assertions", consistent_type_assertions.ConsistentTypeAssertionsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-type-definitions", consistent_type_definitions.ConsistentTypeDefinitionsRule)
	GlobalRuleRegistry.Register("consistent-type-definitions", consistent_type_definitions.ConsistentTypeDefinitionsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-type-exports", consistent_type_exports.ConsistentTypeExportsRule)
	GlobalRuleRegistry.Register("consistent-type-exports", consistent_type_exports.ConsistentTypeExportsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/consistent-type-imports", consistent_type_imports.ConsistentTypeImportsRule)
	GlobalRuleRegistry.Register("consistent-type-imports", consistent_type_imports.ConsistentTypeImportsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/default-param-last", default_param_last.DefaultParamLastRule)
	GlobalRuleRegistry.Register("default-param-last", default_param_last.DefaultParamLastRule)
	GlobalRuleRegistry.Register("@typescript-eslint/explicit-member-accessibility", explicit_member_accessibility.ExplicitMemberAccessibilityRule)
	GlobalRuleRegistry.Register("explicit-member-accessibility", explicit_member_accessibility.ExplicitMemberAccessibilityRule)
	GlobalRuleRegistry.Register("@typescript-eslint/explicit-module-boundary-types", explicit_module_boundary_types.ExplicitModuleBoundaryTypesRule)
	GlobalRuleRegistry.Register("explicit-module-boundary-types", explicit_module_boundary_types.ExplicitModuleBoundaryTypesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/init-declarations", init_declarations.InitDeclarationsRule)
	GlobalRuleRegistry.Register("init-declarations", init_declarations.InitDeclarationsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-array-delete", no_array_delete.NoArrayDeleteRule)
	GlobalRuleRegistry.Register("no-array-delete", no_array_delete.NoArrayDeleteRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-base-to-string", no_base_to_string.NoBaseToStringRule)
	GlobalRuleRegistry.Register("no-base-to-string", no_base_to_string.NoBaseToStringRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-confusing-non-null-assertion", no_confusing_non_null_assertion.NoConfusingNonNullAssertionRule)
	GlobalRuleRegistry.Register("no-confusing-non-null-assertion", no_confusing_non_null_assertion.NoConfusingNonNullAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-confusing-void-expression", no_confusing_void_expression.NoConfusingVoidExpressionRule)
	GlobalRuleRegistry.Register("no-confusing-void-expression", no_confusing_void_expression.NoConfusingVoidExpressionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-duplicate-type-constituents", no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule)
	GlobalRuleRegistry.Register("no-duplicate-type-constituents", no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-duplicate-enum-values", no_duplicate_enum_values.NoDuplicateEnumValuesRule)
	GlobalRuleRegistry.Register("no-duplicate-enum-values", no_duplicate_enum_values.NoDuplicateEnumValuesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-dupe-class-members", no_dupe_class_members.NoDupeClassMembersRule)
	GlobalRuleRegistry.Register("no-dupe-class-members", no_dupe_class_members.NoDupeClassMembersRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-dynamic-delete", no_dynamic_delete.NoDynamicDeleteRule)
	GlobalRuleRegistry.Register("no-dynamic-delete", no_dynamic_delete.NoDynamicDeleteRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-empty-function", no_empty_function.NoEmptyFunctionRule)
	GlobalRuleRegistry.Register("no-empty-function", no_empty_function.NoEmptyFunctionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-empty-interface", no_empty_interface.NoEmptyInterfaceRule)
	GlobalRuleRegistry.Register("no-empty-interface", no_empty_interface.NoEmptyInterfaceRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-empty-object-type", no_empty_object_type.NoEmptyObjectTypeRule)
	GlobalRuleRegistry.Register("no-empty-object-type", no_empty_object_type.NoEmptyObjectTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-floating-promises", no_floating_promises.NoFloatingPromisesRule)
	GlobalRuleRegistry.Register("no-floating-promises", no_floating_promises.NoFloatingPromisesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-for-in-array", no_for_in_array.NoForInArrayRule)
	GlobalRuleRegistry.Register("no-for-in-array", no_for_in_array.NoForInArrayRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-import-type-side-effects", no_import_type_side_effects.NoImportTypeSideEffectsRule)
	GlobalRuleRegistry.Register("no-import-type-side-effects", no_import_type_side_effects.NoImportTypeSideEffectsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-implied-eval", no_implied_eval.NoImpliedEvalRule)
	GlobalRuleRegistry.Register("no-implied-eval", no_implied_eval.NoImpliedEvalRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-inferrable-types", no_inferrable_types.NoInferrableTypesRule)
	GlobalRuleRegistry.Register("no-inferrable-types", no_inferrable_types.NoInferrableTypesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-invalid-this", no_invalid_this.NoInvalidThisRule)
	GlobalRuleRegistry.Register("no-invalid-this", no_invalid_this.NoInvalidThisRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-invalid-void-type", no_invalid_void_type.NoInvalidVoidTypeRule)
	GlobalRuleRegistry.Register("no-invalid-void-type", no_invalid_void_type.NoInvalidVoidTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-loop-func", no_loop_func.NoLoopFuncRule)
	GlobalRuleRegistry.Register("no-loop-func", no_loop_func.NoLoopFuncRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-loss-of-precision", no_loss_of_precision.NoLossOfPrecisionRule)
	GlobalRuleRegistry.Register("no-loss-of-precision", no_loss_of_precision.NoLossOfPrecisionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-magic-numbers", no_magic_numbers.NoMagicNumbersRule)
	GlobalRuleRegistry.Register("no-magic-numbers", no_magic_numbers.NoMagicNumbersRule)
	GlobalRuleRegistry.Register("@typescript-eslint/max-params", max_params.MaxParamsRule)
	GlobalRuleRegistry.Register("max-params", max_params.MaxParamsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/member-ordering", member_ordering.MemberOrderingRule)
	GlobalRuleRegistry.Register("member-ordering", member_ordering.MemberOrderingRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-meaningless-void-operator", no_meaningless_void_operator.NoMeaninglessVoidOperatorRule)
	GlobalRuleRegistry.Register("no-meaningless-void-operator", no_meaningless_void_operator.NoMeaninglessVoidOperatorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-new", no_misused_new.NoMisusedNewRule)
	GlobalRuleRegistry.Register("no-misused-new", no_misused_new.NoMisusedNewRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-promises", no_misused_promises.NoMisusedPromisesRule)
	GlobalRuleRegistry.Register("no-misused-promises", no_misused_promises.NoMisusedPromisesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-spread", no_misused_spread.NoMisusedSpreadRule)
	GlobalRuleRegistry.Register("no-misused-spread", no_misused_spread.NoMisusedSpreadRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-mixed-enums", no_mixed_enums.NoMixedEnumsRule)
	GlobalRuleRegistry.Register("no-mixed-enums", no_mixed_enums.NoMixedEnumsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-namespace", no_namespace.NoNamespaceRule)
	GlobalRuleRegistry.Register("no-namespace", no_namespace.NoNamespaceRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-non-null-asserted-nullish-coalescing", no_non_null_asserted_nullish_coalescing.NoNonNullAssertedNullishCoalescingRule)
	GlobalRuleRegistry.Register("no-non-null-asserted-nullish-coalescing", no_non_null_asserted_nullish_coalescing.NoNonNullAssertedNullishCoalescingRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-non-null-asserted-optional-chain", no_non_null_asserted_optional_chain.NoNonNullAssertedOptionalChainRule)
	GlobalRuleRegistry.Register("no-non-null-asserted-optional-chain", no_non_null_asserted_optional_chain.NoNonNullAssertedOptionalChainRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-non-null-assertion", no_non_null_assertion.NoNonNullAssertionRule)
	GlobalRuleRegistry.Register("no-non-null-assertion", no_non_null_assertion.NoNonNullAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-redeclare", no_redeclare.NoRedeclareRule)
	GlobalRuleRegistry.Register("no-redeclare", no_redeclare.NoRedeclareRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-redundant-type-constituents", no_redundant_type_constituents.NoRedundantTypeConstituentsRule)
	GlobalRuleRegistry.Register("no-redundant-type-constituents", no_redundant_type_constituents.NoRedundantTypeConstituentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-require-imports", no_require_imports.NoRequireImportsRule)
	GlobalRuleRegistry.Register("no-require-imports", no_require_imports.NoRequireImportsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-restricted-imports", no_restricted_imports.NoRestrictedImportsRule)
	GlobalRuleRegistry.Register("no-restricted-imports", no_restricted_imports.NoRestrictedImportsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-restricted-types", no_restricted_types.NoRestrictedTypesRule)
	GlobalRuleRegistry.Register("no-restricted-types", no_restricted_types.NoRestrictedTypesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-shadow", no_shadow.NoShadowRule)
	GlobalRuleRegistry.Register("no-shadow", no_shadow.NoShadowRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-this-alias", no_this_alias.NoThisAliasRule)
	GlobalRuleRegistry.Register("no-this-alias", no_this_alias.NoThisAliasRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-declaration-merging", no_unsafe_declaration_merging.NoUnsafeDeclarationMergingRule)
	GlobalRuleRegistry.Register("no-unsafe-declaration-merging", no_unsafe_declaration_merging.NoUnsafeDeclarationMergingRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-boolean-literal-compare", no_unnecessary_boolean_literal_compare.NoUnnecessaryBooleanLiteralCompareRule)
	GlobalRuleRegistry.Register("no-unnecessary-boolean-literal-compare", no_unnecessary_boolean_literal_compare.NoUnnecessaryBooleanLiteralCompareRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-template-expression", no_unnecessary_template_expression.NoUnnecessaryTemplateExpressionRule)
	GlobalRuleRegistry.Register("no-unnecessary-template-expression", no_unnecessary_template_expression.NoUnnecessaryTemplateExpressionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-arguments", no_unnecessary_type_arguments.NoUnnecessaryTypeArgumentsRule)
	GlobalRuleRegistry.Register("no-unnecessary-type-arguments", no_unnecessary_type_arguments.NoUnnecessaryTypeArgumentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-assertion", no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule)
	GlobalRuleRegistry.Register("no-unnecessary-type-assertion", no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-constraint", no_unnecessary_type_constraint.NoUnnecessaryTypeConstraintRule)
	GlobalRuleRegistry.Register("no-unnecessary-type-constraint", no_unnecessary_type_constraint.NoUnnecessaryTypeConstraintRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-conversion", no_unnecessary_type_conversion.NoUnnecessaryTypeConversionRule)
	GlobalRuleRegistry.Register("no-unnecessary-type-conversion", no_unnecessary_type_conversion.NoUnnecessaryTypeConversionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-parameters", no_unnecessary_type_parameters.NoUnnecessaryTypeParametersRule)
	GlobalRuleRegistry.Register("no-unnecessary-type-parameters", no_unnecessary_type_parameters.NoUnnecessaryTypeParametersRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-argument", no_unsafe_argument.NoUnsafeArgumentRule)
	GlobalRuleRegistry.Register("no-unsafe-argument", no_unsafe_argument.NoUnsafeArgumentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-assignment", no_unsafe_assignment.NoUnsafeAssignmentRule)
	GlobalRuleRegistry.Register("no-unsafe-assignment", no_unsafe_assignment.NoUnsafeAssignmentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-call", no_unsafe_call.NoUnsafeCallRule)
	GlobalRuleRegistry.Register("no-unsafe-call", no_unsafe_call.NoUnsafeCallRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-enum-comparison", no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule)
	GlobalRuleRegistry.Register("no-unsafe-enum-comparison", no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-function-type", no_unsafe_function_type.NoUnsafeFunctionTypeRule)
	GlobalRuleRegistry.Register("no-unsafe-function-type", no_unsafe_function_type.NoUnsafeFunctionTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-member-access", no_unsafe_member_access.NoUnsafeMemberAccessRule)
	GlobalRuleRegistry.Register("no-unsafe-member-access", no_unsafe_member_access.NoUnsafeMemberAccessRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-return", no_unsafe_return.NoUnsafeReturnRule)
	GlobalRuleRegistry.Register("no-unsafe-return", no_unsafe_return.NoUnsafeReturnRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-type-assertion", no_unsafe_type_assertion.NoUnsafeTypeAssertionRule)
	GlobalRuleRegistry.Register("no-unsafe-type-assertion", no_unsafe_type_assertion.NoUnsafeTypeAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-unary-minus", no_unsafe_unary_minus.NoUnsafeUnaryMinusRule)
	GlobalRuleRegistry.Register("no-unsafe-unary-minus", no_unsafe_unary_minus.NoUnsafeUnaryMinusRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unused-expressions", no_unused_expressions.NoUnusedExpressionsRule)
	GlobalRuleRegistry.Register("no-unused-expressions", no_unused_expressions.NoUnusedExpressionsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unused-vars", no_unused_vars.NoUnusedVarsRule)
	GlobalRuleRegistry.Register("no-unused-vars", no_unused_vars.NoUnusedVarsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-use-before-define", no_use_before_define.NoUseBeforeDefineRule)
	GlobalRuleRegistry.Register("no-use-before-define", no_use_before_define.NoUseBeforeDefineRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-useless-constructor", no_useless_constructor.NoUselessConstructorRule)
	GlobalRuleRegistry.Register("no-useless-constructor", no_useless_constructor.NoUselessConstructorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-useless-empty-export", no_useless_empty_export.NoUselessEmptyExportRule)
	GlobalRuleRegistry.Register("no-useless-empty-export", no_useless_empty_export.NoUselessEmptyExportRule)
	GlobalRuleRegistry.Register("@typescript-eslint/non-nullable-type-assertion-style", non_nullable_type_assertion_style.NonNullableTypeAssertionStyleRule)
	GlobalRuleRegistry.Register("non-nullable-type-assertion-style", non_nullable_type_assertion_style.NonNullableTypeAssertionStyleRule)
	GlobalRuleRegistry.Register("@typescript-eslint/only-throw-error", only_throw_error.OnlyThrowErrorRule)
	GlobalRuleRegistry.Register("only-throw-error", only_throw_error.OnlyThrowErrorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-as-const", prefer_as_const.PreferAsConstRule)
	GlobalRuleRegistry.Register("prefer-as-const", prefer_as_const.PreferAsConstRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-promise-reject-errors", prefer_promise_reject_errors.PreferPromiseRejectErrorsRule)
	GlobalRuleRegistry.Register("prefer-promise-reject-errors", prefer_promise_reject_errors.PreferPromiseRejectErrorsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-reduce-type-parameter", prefer_reduce_type_parameter.PreferReduceTypeParameterRule)
	GlobalRuleRegistry.Register("prefer-reduce-type-parameter", prefer_reduce_type_parameter.PreferReduceTypeParameterRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-return-this-type", prefer_return_this_type.PreferReturnThisTypeRule)
	GlobalRuleRegistry.Register("prefer-return-this-type", prefer_return_this_type.PreferReturnThisTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/promise-function-async", promise_function_async.PromiseFunctionAsyncRule)
	GlobalRuleRegistry.Register("promise-function-async", promise_function_async.PromiseFunctionAsyncRule)
	GlobalRuleRegistry.Register("@typescript-eslint/related-getter-setter-pairs", related_getter_setter_pairs.RelatedGetterSetterPairsRule)
	GlobalRuleRegistry.Register("related-getter-setter-pairs", related_getter_setter_pairs.RelatedGetterSetterPairsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/require-array-sort-compare", require_array_sort_compare.RequireArraySortCompareRule)
	GlobalRuleRegistry.Register("require-array-sort-compare", require_array_sort_compare.RequireArraySortCompareRule)
	GlobalRuleRegistry.Register("@typescript-eslint/require-await", require_await.RequireAwaitRule)
	GlobalRuleRegistry.Register("require-await", require_await.RequireAwaitRule)
	GlobalRuleRegistry.Register("@typescript-eslint/restrict-plus-operands", restrict_plus_operands.RestrictPlusOperandsRule)
	GlobalRuleRegistry.Register("restrict-plus-operands", restrict_plus_operands.RestrictPlusOperandsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/restrict-template-expressions", restrict_template_expressions.RestrictTemplateExpressionsRule)
	GlobalRuleRegistry.Register("restrict-template-expressions", restrict_template_expressions.RestrictTemplateExpressionsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/return-await", return_await.ReturnAwaitRule)
	GlobalRuleRegistry.Register("return-await", return_await.ReturnAwaitRule)
	GlobalRuleRegistry.Register("@typescript-eslint/switch-exhaustiveness-check", switch_exhaustiveness_check.SwitchExhaustivenessCheckRule)
	GlobalRuleRegistry.Register("switch-exhaustiveness-check", switch_exhaustiveness_check.SwitchExhaustivenessCheckRule)
	GlobalRuleRegistry.Register("@typescript-eslint/unbound-method", unbound_method.UnboundMethodRule)
	GlobalRuleRegistry.Register("unbound-method", unbound_method.UnboundMethodRule)
	GlobalRuleRegistry.Register("@typescript-eslint/use-unknown-in-catch-callback-variable", use_unknown_in_catch_callback_variable.UseUnknownInCatchCallbackVariableRule)
	GlobalRuleRegistry.Register("use-unknown-in-catch-callback-variable", use_unknown_in_catch_callback_variable.UseUnknownInCatchCallbackVariableRule)
}

// getAllTypeScriptEslintPluginRules returns all registered rules (for backward compatibility when no config is provided)
func getAllTypeScriptEslintPluginRules() []rule.Rule {
	allRules := GlobalRuleRegistry.GetAllRules()
	var rules []rule.Rule
	for _, rule := range allRules {
		// Don't add prefix - rules already have their correct names
		rules = append(rules, rule)
	}
	return rules
}

// isFileIgnored checks if a file should be ignored based on ignore patterns
func isFileIgnored(filePath string, ignorePatterns []string) bool {
	// Get current working directory for relative path resolution
	cwd, err := os.Getwd()
	if err != nil {
		// If we can't get cwd, fall back to simple matching
		return isFileIgnoredSimple(filePath, ignorePatterns)
	}

	// Normalize the file path relative to cwd
	normalizedPath := normalizePath(filePath, cwd)

	for _, pattern := range ignorePatterns {
		// Try matching against normalized path
		if matched, err := doublestar.PathMatch(pattern, normalizedPath); err == nil && matched {
			return true
		}

		// Also try matching against original path for absolute patterns
		if normalizedPath != filePath {
			if matched, err := doublestar.PathMatch(pattern, filePath); err == nil && matched {
				return true
			}
		}

		// Try Unix-style path for cross-platform compatibility
		unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")
		if unixPath != normalizedPath {
			if matched, err := doublestar.PathMatch(pattern, unixPath); err == nil && matched {
				return true
			}
		}
	}
	return false
}

// normalizePath converts file path to be relative to cwd for consistent matching
func normalizePath(filePath, cwd string) string {
	cleanPath := filepath.Clean(filePath)

	// If absolute path, try to make it relative to working directory
	if filepath.IsAbs(cleanPath) {
		if relPath, err := filepath.Rel(cwd, cleanPath); err == nil {
			// Only use relative path if it doesn't go outside the working directory
			if !strings.HasPrefix(relPath, "..") {
				return relPath
			}
		}
	}

	return cleanPath
}

// isFileIgnoredSimple provides fallback matching when cwd is unavailable
func isFileIgnoredSimple(filePath string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		if matched, err := doublestar.PathMatch(pattern, filePath); err == nil && matched {
			return true
		}
	}
	return false
}
