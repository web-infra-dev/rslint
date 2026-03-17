package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
	importPlugin "github.com/web-infra-dev/rslint/internal/plugins/import"
	reactPlugin "github.com/web-infra-dev/rslint/internal/plugins/react"
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
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_array_constructor"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_array_delete"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_base_to_string"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_confusing_void_expression"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_duplicate_enum_values"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_duplicate_type_constituents"
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
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_meaningless_void_operator"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_misused_new"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_misused_promises"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_misused_spread"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_mixed_enums"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_namespace"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_non_null_asserted_nullish_coalescing"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_non_null_asserted_optional_chain"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_non_null_assertion"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_redundant_type_constituents"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_require_imports"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_this_alias"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_boolean_literal_compare"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_template_expression"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_type_arguments"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_type_assertion"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_argument"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_assignment"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_call"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_enum_comparison"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_member_access"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_return"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_type_assertion"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_unary_minus"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unused_vars"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_useless_empty_export"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_var_requires"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/non_nullable_type_assertion_style"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/only_throw_error"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_as_const"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_includes"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_literal_enum_member"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_namespace_keyword"
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
	"github.com/web-infra-dev/rslint/internal/rules/array_callback_return"
	"github.com/web-infra-dev/rslint/internal/rules/constructor_super"
	"github.com/web-infra-dev/rslint/internal/rules/default_case"
	"github.com/web-infra-dev/rslint/internal/rules/for_direction"
	"github.com/web-infra-dev/rslint/internal/rules/getter_return"
	"github.com/web-infra-dev/rslint/internal/rules/no_async_promise_executor"
	"github.com/web-infra-dev/rslint/internal/rules/no_await_in_loop"
	"github.com/web-infra-dev/rslint/internal/rules/no_case_declarations"
	"github.com/web-infra-dev/rslint/internal/rules/no_class_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_compare_neg_zero"
	"github.com/web-infra-dev/rslint/internal/rules/no_cond_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_console"
	"github.com/web-infra-dev/rslint/internal/rules/no_const_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_constant_binary_expression"
	"github.com/web-infra-dev/rslint/internal/rules/no_constant_condition"
	"github.com/web-infra-dev/rslint/internal/rules/no_constructor_return"
	"github.com/web-infra-dev/rslint/internal/rules/no_debugger"
	"github.com/web-infra-dev/rslint/internal/rules/no_dupe_args"
	"github.com/web-infra-dev/rslint/internal/rules/no_dupe_keys"
	"github.com/web-infra-dev/rslint/internal/rules/no_duplicate_case"
	"github.com/web-infra-dev/rslint/internal/rules/no_empty"
	"github.com/web-infra-dev/rslint/internal/rules/no_empty_pattern"
	"github.com/web-infra-dev/rslint/internal/rules/no_import_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_loss_of_precision"
	"github.com/web-infra-dev/rslint/internal/rules/no_sparse_arrays"
	"github.com/web-infra-dev/rslint/internal/rules/no_template_curly_in_string"
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
	switch plugin {
	case "@typescript-eslint":
		return getAllTypeScriptEslintPluginRules()
	case "eslint-plugin-import":
		return importPlugin.GetAllRules()
	case "eslint-plugin-import/recommended":
		return importPlugin.GetRecommendedRules()
	case "react":
		return reactPlugin.GetAllRules()
	default:
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

var registerOnce sync.Once

func RegisterAllRules() {
	registerOnce.Do(func() {
		registerAllTypeScriptEslintPluginRules()
		registerAllEslintImportPluginRules()
		registerAllReactPluginRules()
		registerAllCoreEslintRules()
	})
}

func registerAllReactPluginRules() {
	for _, rule := range reactPlugin.GetAllRules() {
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
	GlobalRuleRegistry.Register("@typescript-eslint/no-array-constructor", no_array_constructor.NoArrayConstructorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-array-delete", no_array_delete.NoArrayDeleteRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-base-to-string", no_base_to_string.NoBaseToStringRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-confusing-void-expression", no_confusing_void_expression.NoConfusingVoidExpressionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-duplicate-enum-values", no_duplicate_enum_values.NoDuplicateEnumValuesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-duplicate-type-constituents", no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule)
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
	GlobalRuleRegistry.Register("@typescript-eslint/no-meaningless-void-operator", no_meaningless_void_operator.NoMeaninglessVoidOperatorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-new", no_misused_new.NoMisusedNewRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-promises", no_misused_promises.NoMisusedPromisesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-spread", no_misused_spread.NoMisusedSpreadRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-mixed-enums", no_mixed_enums.NoMixedEnumsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-namespace", no_namespace.NoNamespaceRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-non-null-asserted-nullish-coalescing", no_non_null_asserted_nullish_coalescing.NoNonNullAssertedNullishCoalescingRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-non-null-asserted-optional-chain", no_non_null_asserted_optional_chain.NoNonNullAssertedOptionalChainRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-non-null-assertion", no_non_null_assertion.NoNonNullAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-redundant-type-constituents", no_redundant_type_constituents.NoRedundantTypeConstituentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-this-alias", no_this_alias.NoThisAliasRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-require-imports", no_require_imports.NoRequireImportsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-boolean-literal-compare", no_unnecessary_boolean_literal_compare.NoUnnecessaryBooleanLiteralCompareRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-template-expression", no_unnecessary_template_expression.NoUnnecessaryTemplateExpressionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-arguments", no_unnecessary_type_arguments.NoUnnecessaryTypeArgumentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-assertion", no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-argument", no_unsafe_argument.NoUnsafeArgumentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-assignment", no_unsafe_assignment.NoUnsafeAssignmentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-call", no_unsafe_call.NoUnsafeCallRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-enum-comparison", no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-member-access", no_unsafe_member_access.NoUnsafeMemberAccessRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-return", no_unsafe_return.NoUnsafeReturnRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-type-assertion", no_unsafe_type_assertion.NoUnsafeTypeAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-unary-minus", no_unsafe_unary_minus.NoUnsafeUnaryMinusRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unused-vars", no_unused_vars.NoUnusedVarsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-useless-empty-export", no_useless_empty_export.NoUselessEmptyExportRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-var-requires", no_var_requires.NoVarRequiresRule)
	GlobalRuleRegistry.Register("@typescript-eslint/non-nullable-type-assertion-style", non_nullable_type_assertion_style.NonNullableTypeAssertionStyleRule)
	GlobalRuleRegistry.Register("@typescript-eslint/only-throw-error", only_throw_error.OnlyThrowErrorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-as-const", prefer_as_const.PreferAsConstRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-includes", prefer_includes.PreferIncludesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-literal-enum-member", prefer_literal_enum_member.PreferLiteralEnumMemberRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-namespace-keyword", prefer_namespace_keyword.PreferNamespaceKeywordRule)
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
	GlobalRuleRegistry.Register("array-callback-return", array_callback_return.ArrayCallbackReturnRule)
	GlobalRuleRegistry.Register("constructor-super", constructor_super.ConstructorSuperRule)
	GlobalRuleRegistry.Register("default-case", default_case.DefaultCaseRule)
	GlobalRuleRegistry.Register("for-direction", for_direction.ForDirectionRule)
	GlobalRuleRegistry.Register("getter-return", getter_return.GetterReturnRule)
	GlobalRuleRegistry.Register("no-async-promise-executor", no_async_promise_executor.NoAsyncPromiseExecutorRule)
	GlobalRuleRegistry.Register("no-await-in-loop", no_await_in_loop.NoAwaitInLoopRule)
	GlobalRuleRegistry.Register("no-case-declarations", no_case_declarations.NoCaseDeclarationsRule)
	GlobalRuleRegistry.Register("no-class-assign", no_class_assign.NoClassAssignRule)
	GlobalRuleRegistry.Register("no-compare-neg-zero", no_compare_neg_zero.NoCompareNegZeroRule)
	GlobalRuleRegistry.Register("no-cond-assign", no_cond_assign.NoCondAssignRule)
	GlobalRuleRegistry.Register("no-console", no_console.NoConsoleRule)
	GlobalRuleRegistry.Register("no-const-assign", no_const_assign.NoConstAssignRule)
	GlobalRuleRegistry.Register("no-constant-binary-expression", no_constant_binary_expression.NoConstantBinaryExpressionRule)
	GlobalRuleRegistry.Register("no-constant-condition", no_constant_condition.NoConstantConditionRule)
	GlobalRuleRegistry.Register("no-constructor-return", no_constructor_return.NoConstructorReturnRule)
	GlobalRuleRegistry.Register("no-debugger", no_debugger.NoDebuggerRule)
	GlobalRuleRegistry.Register("no-dupe-args", no_dupe_args.NoDupeArgsRule)
	GlobalRuleRegistry.Register("no-dupe-keys", no_dupe_keys.NoDupeKeysRule)
	GlobalRuleRegistry.Register("no-duplicate-case", no_duplicate_case.NoDuplicateCaseRule)
	GlobalRuleRegistry.Register("no-empty", no_empty.NoEmptyRule)
	GlobalRuleRegistry.Register("no-empty-pattern", no_empty_pattern.NoEmptyPatternRule)
	GlobalRuleRegistry.Register("no-import-assign", no_import_assign.NoImportAssignRule)
	GlobalRuleRegistry.Register("no-loss-of-precision", no_loss_of_precision.NoLossOfPrecisionRule)
	GlobalRuleRegistry.Register("no-template-curly-in-string", no_template_curly_in_string.NoTemplateCurlyInString)
	GlobalRuleRegistry.Register("no-sparse-arrays", no_sparse_arrays.NoSparseArraysRule)
}

// getAllTypeScriptEslintPluginRules returns all rules from the global registry.
func getAllTypeScriptEslintPluginRules() []rule.Rule {
	allRules := GlobalRuleRegistry.GetAllRules()
	var rules []rule.Rule
	for _, rule := range allRules {
		rules = append(rules, rule)
	}
	return rules
}

func isFileIgnored(filePath string, ignorePatterns []string, cwd string) bool {
	if cwd == "" {
		return isFileIgnoredSimple(filePath, ignorePatterns)
	}

	// Normalize the file path relative to cwd
	normalizedPath := normalizePath(filePath, cwd)

	for _, pattern := range ignorePatterns {
		// Try matching against normalized path
		if matched, err := doublestar.Match(pattern, normalizedPath); err == nil && matched {
			return true
		}

		// Also try matching against original path for absolute patterns
		if normalizedPath != filePath {
			if matched, err := doublestar.Match(pattern, filePath); err == nil && matched {
				return true
			}
		}

		// Try Unix-style path for cross-platform compatibility
		unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")
		if unixPath != normalizedPath {
			if matched, err := doublestar.Match(pattern, unixPath); err == nil && matched {
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
	for _, pattern := range ignorePatterns {
		if matched, err := doublestar.Match(pattern, filePath); err == nil && matched {
			return true
		}
	}
	return false
}

// MergedConfig is the final computed configuration for a single file
type MergedConfig struct {
	Rules           map[string]*RuleConfig
	Settings        Settings
	LanguageOptions *LanguageOptions
}

// GetConfigForFile computes the merged configuration for a file following ESLint flat config semantics.
// Returns nil if the file is globally ignored or no entry matches (should not be linted).
// Both JS and JSON configs are processed identically here — any differences in default rule
// behavior are handled during config loading (see normalizeJSONConfig).
// cwd is the directory the config lives in; file paths are resolved relative to it
// for files/ignores glob matching.
func (config RslintConfig) GetConfigForFile(filePath string, cwd string) *MergedConfig {
	merged := &MergedConfig{
		Rules: make(map[string]*RuleConfig),
	}

	// Track whether any non-global entry matched this file
	entryMatched := false

	for _, entry := range config {
		// 1. Global ignores: entry with only ignores means "skip this file entirely"
		if isGlobalIgnoreEntry(entry) {
			if isFileIgnored(filePath, entry.Ignores, cwd) {
				return nil
			}
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

		// 5. Settings: shallow merge
		if entry.Settings != nil {
			if merged.Settings == nil {
				merged.Settings = make(Settings)
			}
			for k, v := range entry.Settings {
				merged.Settings[k] = v
			}
		}

		// 6. LanguageOptions: deep merge
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
		if matched, err := doublestar.Match(pattern, normalizedPath); err == nil && matched {
			return true
		}
		if normalizedPath != filePath {
			if matched, err := doublestar.Match(pattern, filePath); err == nil && matched {
				return true
			}
		}
		unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")
		if unixPath != normalizedPath {
			if matched, err := doublestar.Match(pattern, unixPath); err == nil && matched {
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

const defaultTSConfig = `import { defineConfig, ts } from '@rslint/core';

export default defineConfig([
  ts.configs.recommended,
  {
    rules: {
      // customize rules here
    },
  },
]);
`

const defaultJSConfig = `import { defineConfig, js } from '@rslint/core';

export default defineConfig([
  js.configs.recommended,
  {
    rules: {
      // customize rules here
    },
  },
]);
`

// isESMPackage checks if the package.json in the given directory has "type": "module".
func isESMPackage(directory string) bool {
	data, err := os.ReadFile(filepath.Join(directory, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}
	return pkg.Type == "module"
}

// InitDefaultConfig initializes a default config file in the directory.
// - If tsconfig.json exists → rslint.config.ts (ESM syntax, handled by TS loaders)
// - Otherwise, follows the ESLint convention based on package.json "type" field:
//   - "type": "module" → rslint.config.js  (ESM syntax, .js is ESM in this context)
//   - no "type": "module" → rslint.config.mjs (ESM syntax, .mjs is always ESM)
func InitDefaultConfig(directory string) error {
	allConfigs := []string{
		"rslint.config.ts", "rslint.config.mts",
		"rslint.config.js", "rslint.config.mjs",
		"rslint.json", "rslint.jsonc",
	}
	for _, name := range allConfigs {
		p := filepath.Join(directory, name)
		if _, err := os.Stat(p); err == nil {
			return fmt.Errorf("config file already exists: %s", name)
		}
	}

	tsconfigPath := filepath.Join(directory, "tsconfig.json")
	if _, err := os.Stat(tsconfigPath); err == nil {
		configPath := filepath.Join(directory, "rslint.config.ts")
		if err := os.WriteFile(configPath, []byte(defaultTSConfig), 0644); err != nil {
			return fmt.Errorf("failed to create rslint.config.ts: %w", err)
		}
		fmt.Println("Created rslint.config.ts with TypeScript recommended config.")
	} else {
		// Use .js when the project is ESM ("type": "module" in package.json),
		// otherwise .mjs to ensure Node.js treats the file as ESM regardless.
		var configName, content string
		if isESMPackage(directory) {
			configName = "rslint.config.js"
			content = defaultJSConfig
		} else {
			configName = "rslint.config.mjs"
			content = defaultJSConfig
		}
		configPath := filepath.Join(directory, configName)
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", configName, err)
		}
		fmt.Printf("Created %s with JavaScript recommended config.\n", configName)
	}

	return nil
}
