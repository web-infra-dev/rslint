package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
	importPlugin "github.com/web-infra-dev/rslint/internal/plugins/import"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/adjacent_overload_signatures"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/array_type"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/await_thenable"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/ban_ts_comment"
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
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_literal_enum_member"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_namespace_keyword"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_promise_reject_errors"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_readonly"
	// "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_readonly_parameter_types" // Temporarily disabled - incomplete implementation
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_reduce_type_parameter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_return_this_type"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_string_starts_ends_with"
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
	"github.com/web-infra-dev/rslint/internal/rules/no_loss_of_precision"
	"github.com/web-infra-dev/rslint/internal/rules/no_sparse_arrays"
	"github.com/web-infra-dev/rslint/internal/rules/no_template_curly_in_string"
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

// ParserOptions contains parser-specific configuration
type ParserOptions struct {
	ProjectService bool         `json:"projectService"`
	Project        ProjectPaths `json:"project,omitempty"`
}

// Rules represents the rules configuration
// This can be extended to include specific rule configurations
type Rules map[string]interface{}

// Alternative: If you want type-safe rule configurations
type TypedRules struct {
	// Example rule configurations - extend as needed
	AdjacentOverloadSignatures         *RuleConfig `json:"@typescript-eslint/adjacent-overload-signatures,omitempty"`
	ArrayType                          *RuleConfig `json:"@typescript-eslint/array-type,omitempty"`
	ClassLiteralPropertyStyle          *RuleConfig `json:"@typescript-eslint/class-literal-property-style,omitempty"`
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

const defaultJsonc = `
[
  {
    // ignore files and folders for linting
    "ignores": [],
    "languageOptions": {
      "parserOptions": {
        // Rslint will lint all files included in your typescript projects defined here
        // support lint multi packages in monorepo
        "project": ["./tsconfig.json"]
      }
    },
    // same configuration as https://typescript-eslint.io/rules/
    "rules": {
      "@typescript-eslint/require-await": "off",
      "@typescript-eslint/no-unnecessary-type-assertion": "warn",
      "@typescript-eslint/array-type": ["warn", { "default": "array-simple" }]
    },
    "plugins": [
      "@typescript-eslint" // will enable all implemented @typescript-eslint rules by default
    ]
  }
]
`

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

			/// Merge rules from plugin
			for _, plugin := range entry.Plugins {

				for _, rule := range GetAllRulesForPlugin(plugin) {
					enabledRules[rule.Name] = &RuleConfig{Level: "error"} // Default level for plugin rules
				}
			}
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

func RegisterAllRules() {
	registerAllTypeScriptEslintPluginRules()
	registerAllEslintImportPluginRules()
	registerAllCoreEslintRules()
}

// registerAllTypeScriptEslintPluginRules registers all available rules in the global registry
func registerAllTypeScriptEslintPluginRules() {
	GlobalRuleRegistry.Register("@typescript-eslint/adjacent-overload-signatures", adjacent_overload_signatures.AdjacentOverloadSignaturesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/array-type", array_type.ArrayTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/await-thenable", await_thenable.AwaitThenableRule)
	GlobalRuleRegistry.Register("@typescript-eslint/ban-ts-comment", ban_ts_comment.BanTsCommentRule)
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
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-literal-enum-member", prefer_literal_enum_member.PreferLiteralEnumMemberRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-namespace-keyword", prefer_namespace_keyword.PreferNamespaceKeywordRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-promise-reject-errors", prefer_promise_reject_errors.PreferPromiseRejectErrorsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-readonly", prefer_readonly.PreferReadonlyRule)
	// TODO: prefer-readonly-parameter-types needs complete implementation for proper type checking
	// Temporarily disabled until the isReadonlyType function is fully implemented with proper
	// detection of readonly arrays, readonly objects, function types, and other edge cases
	// GlobalRuleRegistry.Register("@typescript-eslint/prefer-readonly-parameter-types", prefer_readonly_parameter_types.PreferReadonlyParameterTypesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-reduce-type-parameter", prefer_reduce_type_parameter.PreferReduceTypeParameterRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-return-this-type", prefer_return_this_type.PreferReturnThisTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-string-starts-ends-with", prefer_string_starts_ends_with.PreferStringStartsEndsWithRule)
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
	GlobalRuleRegistry.Register("no-loss-of-precision", no_loss_of_precision.NoLossOfPrecisionRule)
	GlobalRuleRegistry.Register("no-template-curly-in-string", no_template_curly_in_string.NoTemplateCurlyInString)
	GlobalRuleRegistry.Register("no-sparse-arrays", no_sparse_arrays.NoSparseArraysRule)
}

// getAllTypeScriptEslintPluginRules returns all registered rules (for backward compatibility when no config is provided)
func getAllTypeScriptEslintPluginRules() []rule.Rule {
	allRules := GlobalRuleRegistry.GetAllRules()
	var rules []rule.Rule
	for _, rule := range allRules {
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

// initialize a default config in the directory
func InitDefaultConfig(directory string) error {
	configPath := filepath.Join(directory, "rslint.jsonc")

	// if the config exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("rslint.json already exists in %s", directory)
	}

	// write file content
	err := os.WriteFile(configPath, []byte(defaultJsonc), 0644)
	if err != nil {
		return fmt.Errorf("failed to create rslint.json: %w", err)
	}

	return nil
}
