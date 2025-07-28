package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	ipc "github.com/typescript-eslint/rslint/internal/api"
	rslintconfig "github.com/typescript-eslint/rslint/internal/config"
	"github.com/typescript-eslint/rslint/internal/linter"
	"github.com/typescript-eslint/rslint/internal/rule"
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
	"github.com/typescript-eslint/rslint/internal/rules/no_array_delete"
	"github.com/typescript-eslint/rslint/internal/rules/no_base_to_string"
	"github.com/typescript-eslint/rslint/internal/rules/no_confusing_void_expression"
	"github.com/typescript-eslint/rslint/internal/rules/no_duplicate_type_constituents"
	"github.com/typescript-eslint/rslint/internal/rules/no_floating_promises"
	"github.com/typescript-eslint/rslint/internal/rules/no_for_in_array"
	"github.com/typescript-eslint/rslint/internal/rules/no_implied_eval"
	"github.com/typescript-eslint/rslint/internal/rules/no_meaningless_void_operator"
	"github.com/typescript-eslint/rslint/internal/rules/no_misused_promises"
	"github.com/typescript-eslint/rslint/internal/rules/no_misused_spread"
	"github.com/typescript-eslint/rslint/internal/rules/no_mixed_enums"
	"github.com/typescript-eslint/rslint/internal/rules/no_redundant_type_constituents"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_boolean_literal_compare"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_template_expression"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_type_arguments"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_type_assertion"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_argument"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_assignment"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_call"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_enum_comparison"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_member_access"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_return"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_type_assertion"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_unary_minus"
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
	"github.com/typescript-eslint/rslint/internal/rules/explicit_member_accessibility"
	"github.com/typescript-eslint/rslint/internal/rules/explicit_module_boundary_types"
	"github.com/typescript-eslint/rslint/internal/rules/init_declarations"
	"github.com/typescript-eslint/rslint/internal/rules/max_params"
	"github.com/typescript-eslint/rslint/internal/rules/no_confusing_non_null_assertion"
	"github.com/typescript-eslint/rslint/internal/rules/no_dupe_class_members"
	"github.com/typescript-eslint/rslint/internal/rules/no_duplicate_enum_values"
	"github.com/typescript-eslint/rslint/internal/rules/no_dynamic_delete"
	"github.com/typescript-eslint/rslint/internal/rules/no_empty_function"
	"github.com/typescript-eslint/rslint/internal/rules/no_empty_interface"
	"github.com/typescript-eslint/rslint/internal/rules/no_empty_object_type"
	"github.com/typescript-eslint/rslint/internal/rules/no_import_type_side_effects"
	"github.com/typescript-eslint/rslint/internal/rules/no_inferrable_types"
	"github.com/typescript-eslint/rslint/internal/rules/no_invalid_this"
	"github.com/typescript-eslint/rslint/internal/rules/no_invalid_void_type"
	"github.com/typescript-eslint/rslint/internal/rules/no_loop_func"
	"github.com/typescript-eslint/rslint/internal/rules/no_loss_of_precision"
	"github.com/typescript-eslint/rslint/internal/rules/no_magic_numbers"
	"github.com/typescript-eslint/rslint/internal/rules/no_misused_new"
	"github.com/typescript-eslint/rslint/internal/rules/no_namespace"
	"github.com/typescript-eslint/rslint/internal/rules/no_non_null_asserted_nullish_coalescing"
	"github.com/typescript-eslint/rslint/internal/rules/no_non_null_asserted_optional_chain"
	"github.com/typescript-eslint/rslint/internal/rules/no_non_null_assertion"
	"github.com/typescript-eslint/rslint/internal/rules/no_redeclare"
	"github.com/typescript-eslint/rslint/internal/rules/no_require_imports"
	"github.com/typescript-eslint/rslint/internal/rules/no_restricted_imports"
	"github.com/typescript-eslint/rslint/internal/rules/no_restricted_types"
	"github.com/typescript-eslint/rslint/internal/rules/no_shadow"
	"github.com/typescript-eslint/rslint/internal/rules/no_this_alias"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_type_constraint"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_type_conversion"
	"github.com/typescript-eslint/rslint/internal/rules/no_unnecessary_type_parameters"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_declaration_merging"
	"github.com/typescript-eslint/rslint/internal/rules/no_unsafe_function_type"
	"github.com/typescript-eslint/rslint/internal/rules/no_unused_expressions"
	"github.com/typescript-eslint/rslint/internal/rules/no_unused_vars"
	"github.com/typescript-eslint/rslint/internal/rules/no_use_before_define"
	"github.com/typescript-eslint/rslint/internal/rules/no_useless_constructor"
	"github.com/typescript-eslint/rslint/internal/rules/no_useless_empty_export"
	"github.com/typescript-eslint/rslint/internal/utils"
)

// IPCHandler implements the ipc.Handler interface
type IPCHandler struct{}

// HandleLint handles lint requests in IPC mode
func (h *IPCHandler) HandleLint(req ipc.LintRequest) (*ipc.LintResponse, error) {
	// Format is not used for IPC mode as we return structured data
	_ = req.Format

	// Set working directory if provided
	if req.WorkingDirectory != "" {
		if err := os.Chdir(req.WorkingDirectory); err != nil {
			return nil, fmt.Errorf("failed to change directory: %w", err)
		}
	}

	// Get current directory
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)

	// Create filesystem
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	// Apply file contents if provided
	if len(req.FileContents) > 0 {
		fs = utils.NewOverlayVFS(fs, req.FileContents)
	}

	// Initialize rule registry with all available rules
	rslintconfig.RegisterAllTypeSriptEslintPluginRules()

	// Load rslint configuration and determine which tsconfig files to use
	_, tsConfigs, configDirectory := rslintconfig.LoadConfigurationWithFallback(req.Config, currentDirectory, fs)

	// Create a map of all available rules
	allRules := map[string]rule.Rule{
		"array-type":                                  array_type.ArrayTypeRule,
		"await-thenable":                              await_thenable.AwaitThenableRule,
		"ban-ts-comment":                              ban_ts_comment.BanTsCommentRule,
		"ban-tslint-comment":                          ban_tslint_comment.BanTslintCommentRule,
		"class-literal-property-style":                class_literal_property_style.ClassLiteralPropertyStyleRule,
		"class-methods-use-this":                      class_methods_use_this.ClassMethodsUseThisRule,
		"consistent-generic-constructors":             consistent_generic_constructors.ConsistentGenericConstructorsRule,
		"@typescript-eslint/consistent-generic-constructors": consistent_generic_constructors.ConsistentGenericConstructorsRule,
		"consistent-indexed-object-style":             consistent_indexed_object_style.ConsistentIndexedObjectStyleRule,
		"@typescript-eslint/consistent-indexed-object-style": consistent_indexed_object_style.ConsistentIndexedObjectStyleRule,
		"consistent-return":                           consistent_return.ConsistentReturnRule,
		"@typescript-eslint/consistent-return":        consistent_return.ConsistentReturnRule,
		"consistent-type-assertions":                  consistent_type_assertions.ConsistentTypeAssertionsRule,
		"@typescript-eslint/consistent-type-assertions": consistent_type_assertions.ConsistentTypeAssertionsRule,
		"consistent-type-definitions":                 consistent_type_definitions.ConsistentTypeDefinitionsRule,
		"@typescript-eslint/consistent-type-definitions": consistent_type_definitions.ConsistentTypeDefinitionsRule,
		"consistent-type-exports":                     consistent_type_exports.ConsistentTypeExportsRule,
		"@typescript-eslint/consistent-type-exports":  consistent_type_exports.ConsistentTypeExportsRule,
		"consistent-type-imports":                     consistent_type_imports.ConsistentTypeImportsRule,
		"@typescript-eslint/consistent-type-imports":  consistent_type_imports.ConsistentTypeImportsRule,
		"default-param-last":                          default_param_last.DefaultParamLastRule,
		"@typescript-eslint/default-param-last":       default_param_last.DefaultParamLastRule,
		"no-array-delete":                             no_array_delete.NoArrayDeleteRule,
		"no-base-to-string":                           no_base_to_string.NoBaseToStringRule,
		"no-confusing-void-expression":                no_confusing_void_expression.NoConfusingVoidExpressionRule,
		"no-duplicate-type-constituents":              no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule,
		"no-floating-promises":                        no_floating_promises.NoFloatingPromisesRule,
		"no-for-in-array":                             no_for_in_array.NoForInArrayRule,
		"no-implied-eval":                             no_implied_eval.NoImpliedEvalRule,
		"no-meaningless-void-operator":                no_meaningless_void_operator.NoMeaninglessVoidOperatorRule,
		"no-misused-promises":                         no_misused_promises.NoMisusedPromisesRule,
		"no-misused-spread":                           no_misused_spread.NoMisusedSpreadRule,
		"no-mixed-enums":                              no_mixed_enums.NoMixedEnumsRule,
		"no-redundant-type-constituents":              no_redundant_type_constituents.NoRedundantTypeConstituentsRule,
		"no-unnecessary-boolean-literal-compare":      no_unnecessary_boolean_literal_compare.NoUnnecessaryBooleanLiteralCompareRule,
		"no-unnecessary-template-expression":          no_unnecessary_template_expression.NoUnnecessaryTemplateExpressionRule,
		"no-unnecessary-type-arguments":               no_unnecessary_type_arguments.NoUnnecessaryTypeArgumentsRule,
		"no-unnecessary-type-assertion":               no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule,
		"no-unsafe-argument":                          no_unsafe_argument.NoUnsafeArgumentRule,
		"no-unsafe-assignment":                        no_unsafe_assignment.NoUnsafeAssignmentRule,
		"no-unsafe-call":                              no_unsafe_call.NoUnsafeCallRule,
		"no-unsafe-enum-comparison":                   no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule,
		"no-unsafe-member-access":                     no_unsafe_member_access.NoUnsafeMemberAccessRule,
		"no-unsafe-return":                            no_unsafe_return.NoUnsafeReturnRule,
		"no-unsafe-type-assertion":                    no_unsafe_type_assertion.NoUnsafeTypeAssertionRule,
		"no-unsafe-unary-minus":                       no_unsafe_unary_minus.NoUnsafeUnaryMinusRule,
		"non-nullable-type-assertion-style":           non_nullable_type_assertion_style.NonNullableTypeAssertionStyleRule,
		"only-throw-error":                            only_throw_error.OnlyThrowErrorRule,
		"prefer-as-const":                             prefer_as_const.PreferAsConstRule,
		"prefer-promise-reject-errors":                prefer_promise_reject_errors.PreferPromiseRejectErrorsRule,
		"prefer-reduce-type-parameter":                prefer_reduce_type_parameter.PreferReduceTypeParameterRule,
		"prefer-return-this-type":                     prefer_return_this_type.PreferReturnThisTypeRule,
		"promise-function-async":                      promise_function_async.PromiseFunctionAsyncRule,
		"related-getter-setter-pairs":                 related_getter_setter_pairs.RelatedGetterSetterPairsRule,
		"require-array-sort-compare":                  require_array_sort_compare.RequireArraySortCompareRule,
		"require-await":                               require_await.RequireAwaitRule,
		"restrict-plus-operands":                      restrict_plus_operands.RestrictPlusOperandsRule,
		"restrict-template-expressions":               restrict_template_expressions.RestrictTemplateExpressionsRule,
		"return-await":                                return_await.ReturnAwaitRule,
		"switch-exhaustiveness-check":                 switch_exhaustiveness_check.SwitchExhaustivenessCheckRule,
		"unbound-method":                              unbound_method.UnboundMethodRule,
		"use-unknown-in-catch-callback-variable":      use_unknown_in_catch_callback_variable.UseUnknownInCatchCallbackVariableRule,
		"@typescript-eslint/explicit-member-accessibility": explicit_member_accessibility.ExplicitMemberAccessibilityRule,
		"explicit-member-accessibility":               explicit_member_accessibility.ExplicitMemberAccessibilityRule,
		"@typescript-eslint/explicit-module-boundary-types": explicit_module_boundary_types.ExplicitModuleBoundaryTypesRule,
		"explicit-module-boundary-types":              explicit_module_boundary_types.ExplicitModuleBoundaryTypesRule,
		"@typescript-eslint/init-declarations":        init_declarations.InitDeclarationsRule,
		"init-declarations":                           init_declarations.InitDeclarationsRule,
		"@typescript-eslint/max-params":               max_params.MaxParamsRule,
		"max-params":                                  max_params.MaxParamsRule,
		"@typescript-eslint/no-confusing-non-null-assertion": no_confusing_non_null_assertion.NoConfusingNonNullAssertionRule,
		"no-confusing-non-null-assertion":            no_confusing_non_null_assertion.NoConfusingNonNullAssertionRule,
		"@typescript-eslint/no-dupe-class-members":    no_dupe_class_members.NoDupeClassMembersRule,
		"no-dupe-class-members":                       no_dupe_class_members.NoDupeClassMembersRule,
		"@typescript-eslint/no-duplicate-enum-values": no_duplicate_enum_values.NoDuplicateEnumValuesRule,
		"no-duplicate-enum-values":                    no_duplicate_enum_values.NoDuplicateEnumValuesRule,
		"@typescript-eslint/no-dynamic-delete":        no_dynamic_delete.NoDynamicDeleteRule,
		"no-dynamic-delete":                           no_dynamic_delete.NoDynamicDeleteRule,
		"@typescript-eslint/no-empty-function":        no_empty_function.NoEmptyFunctionRule,
		"no-empty-function":                           no_empty_function.NoEmptyFunctionRule,
		"@typescript-eslint/no-empty-interface":       no_empty_interface.NoEmptyInterfaceRule,
		"no-empty-interface":                          no_empty_interface.NoEmptyInterfaceRule,
		"@typescript-eslint/no-empty-object-type":     no_empty_object_type.NoEmptyObjectTypeRule,
		"no-empty-object-type":                        no_empty_object_type.NoEmptyObjectTypeRule,
		"@typescript-eslint/no-import-type-side-effects": no_import_type_side_effects.NoImportTypeSideEffectsRule,
		"no-import-type-side-effects":                 no_import_type_side_effects.NoImportTypeSideEffectsRule,
		"@typescript-eslint/no-inferrable-types":      no_inferrable_types.NoInferrableTypesRule,
		"no-inferrable-types":                         no_inferrable_types.NoInferrableTypesRule,
		"@typescript-eslint/no-invalid-this":          no_invalid_this.NoInvalidThisRule,
		"no-invalid-this":                             no_invalid_this.NoInvalidThisRule,
		"@typescript-eslint/no-invalid-void-type":     no_invalid_void_type.NoInvalidVoidTypeRule,
		"no-invalid-void-type":                        no_invalid_void_type.NoInvalidVoidTypeRule,
		"@typescript-eslint/no-loop-func":             no_loop_func.NoLoopFuncRule,
		"no-loop-func":                                no_loop_func.NoLoopFuncRule,
		"@typescript-eslint/no-loss-of-precision":     no_loss_of_precision.NoLossOfPrecisionRule,
		"no-loss-of-precision":                        no_loss_of_precision.NoLossOfPrecisionRule,
		"@typescript-eslint/no-magic-numbers":         no_magic_numbers.NoMagicNumbersRule,
		"no-magic-numbers":                            no_magic_numbers.NoMagicNumbersRule,
		"@typescript-eslint/no-misused-new":           no_misused_new.NoMisusedNewRule,
		"no-misused-new":                              no_misused_new.NoMisusedNewRule,
		"@typescript-eslint/no-namespace":             no_namespace.NoNamespaceRule,
		"no-namespace":                                no_namespace.NoNamespaceRule,
		"@typescript-eslint/no-non-null-asserted-nullish-coalescing": no_non_null_asserted_nullish_coalescing.NoNonNullAssertedNullishCoalescingRule,
		"no-non-null-asserted-nullish-coalescing":    no_non_null_asserted_nullish_coalescing.NoNonNullAssertedNullishCoalescingRule,
		"@typescript-eslint/no-non-null-asserted-optional-chain": no_non_null_asserted_optional_chain.NoNonNullAssertedOptionalChainRule,
		"no-non-null-asserted-optional-chain":        no_non_null_asserted_optional_chain.NoNonNullAssertedOptionalChainRule,
		"@typescript-eslint/no-non-null-assertion":    no_non_null_assertion.NoNonNullAssertionRule,
		"no-non-null-assertion":                      no_non_null_assertion.NoNonNullAssertionRule,
		"@typescript-eslint/no-redeclare":             no_redeclare.NoRedeclareRule,
		"no-redeclare":                                no_redeclare.NoRedeclareRule,
		"@typescript-eslint/no-require-imports":       no_require_imports.NoRequireImportsRule,
		"no-require-imports":                          no_require_imports.NoRequireImportsRule,
		"@typescript-eslint/no-restricted-imports":    no_restricted_imports.NoRestrictedImportsRule,
		"no-restricted-imports":                       no_restricted_imports.NoRestrictedImportsRule,
		"@typescript-eslint/no-restricted-types":      no_restricted_types.NoRestrictedTypesRule,
		"no-restricted-types":                         no_restricted_types.NoRestrictedTypesRule,
		"@typescript-eslint/no-shadow":                no_shadow.NoShadowRule,
		"no-shadow":                                   no_shadow.NoShadowRule,
		"@typescript-eslint/no-this-alias":            no_this_alias.NoThisAliasRule,
		"no-this-alias":                               no_this_alias.NoThisAliasRule,
		"@typescript-eslint/no-unnecessary-type-constraint": no_unnecessary_type_constraint.NoUnnecessaryTypeConstraintRule,
		"no-unnecessary-type-constraint":             no_unnecessary_type_constraint.NoUnnecessaryTypeConstraintRule,
		"@typescript-eslint/no-unnecessary-type-conversion": no_unnecessary_type_conversion.NoUnnecessaryTypeConversionRule,
		"no-unnecessary-type-conversion":             no_unnecessary_type_conversion.NoUnnecessaryTypeConversionRule,
		"@typescript-eslint/no-unnecessary-type-parameters": no_unnecessary_type_parameters.NoUnnecessaryTypeParametersRule,
		"no-unnecessary-type-parameters":             no_unnecessary_type_parameters.NoUnnecessaryTypeParametersRule,
		"@typescript-eslint/no-unsafe-declaration-merging": no_unsafe_declaration_merging.NoUnsafeDeclarationMergingRule,
		"no-unsafe-declaration-merging":              no_unsafe_declaration_merging.NoUnsafeDeclarationMergingRule,
		"@typescript-eslint/no-unsafe-function-type":  no_unsafe_function_type.NoUnsafeFunctionTypeRule,
		"no-unsafe-function-type":                     no_unsafe_function_type.NoUnsafeFunctionTypeRule,
		"@typescript-eslint/no-unused-expressions":    no_unused_expressions.NoUnusedExpressionsRule,
		"no-unused-expressions":                       no_unused_expressions.NoUnusedExpressionsRule,
		"@typescript-eslint/no-unused-vars":           no_unused_vars.NoUnusedVarsRule,
		"no-unused-vars":                              no_unused_vars.NoUnusedVarsRule,
		"@typescript-eslint/no-use-before-define":     no_use_before_define.NoUseBeforeDefineRule,
		"no-use-before-define":                        no_use_before_define.NoUseBeforeDefineRule,
		"@typescript-eslint/no-useless-constructor":   no_useless_constructor.NoUselessConstructorRule,
		"no-useless-constructor":                      no_useless_constructor.NoUselessConstructorRule,
		"@typescript-eslint/no-useless-empty-export":  no_useless_empty_export.NoUselessEmptyExportRule,
		"no-useless-empty-export":                     no_useless_empty_export.NoUselessEmptyExportRule,
	}

	// Build rules with their configurations from request.RuleOptions
	var rulesWithConfig []rslintconfig.EnabledRuleWithConfig
	if len(req.RuleOptions) > 0 {
		for ruleName, ruleConfig := range req.RuleOptions {
			if rule, exists := allRules[ruleName]; exists {
				// Parse the rule config - can be just a string or an array with options
				var options map[string]interface{}
				var level string

				switch v := ruleConfig.(type) {
				case string:
					level = v
				case []interface{}:
					if len(v) > 0 {
						if levelStr, ok := v[0].(string); ok {
							level = levelStr
							if len(v) > 1 {
								if opts, ok := v[1].(map[string]interface{}); ok {
									options = opts
								} else {
									// For rules that expect a simple option value, pass it directly
									options = map[string]interface{}{"value": v[1]}
								}
							}
						}
					}
				}

				if level != "off" {
					rulesWithConfig = append(rulesWithConfig, rslintconfig.EnabledRuleWithConfig{
						Rule: rule,
						Config: &rslintconfig.RuleConfig{
							Level:   level,
							Options: options,
						},
					})
				}
			}
		}
	} else {
		// If no specific rules requested, use all rules
		for _, rule := range allRules {
			rulesWithConfig = append(rulesWithConfig, rslintconfig.EnabledRuleWithConfig{
				Rule: rule,
				Config: &rslintconfig.RuleConfig{
					Level: "error",
				},
			})
		}
	}

	// Create compiler host
	host := utils.CreateCompilerHost(configDirectory, fs)
	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          host.GetCurrentDirectory(),
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
	}

	// Create programs from all tsconfig files found in rslint config
	programs := []*compiler.Program{}
	for _, configFileName := range tsConfigs {
		program, err := utils.CreateProgram(false, fs, configDirectory, configFileName, host)
		if err != nil {
			return nil, fmt.Errorf("error creating TS program for %s: %v", configFileName, err)
		}
		programs = append(programs, program)
	}

	// Find source files from all programs
	files := []*ast.SourceFile{}

	// If specific files are provided, use those
	if len(req.Files) > 0 {
		for _, filePath := range req.Files {
			absPath := tspath.ResolvePath(configDirectory, filePath)
			// Try to find the file in any of the programs
			for _, program := range programs {
				sourceFile := program.GetSourceFile(absPath)
				if sourceFile != nil {
					files = append(files, sourceFile)
					break // Found in this program, no need to check others
				}
			}
		}
	} else {
		// Otherwise use all source files from all programs
		for _, program := range programs {
			for _, file := range program.SourceFiles() {
				p := string(file.Path())
				if strings.Contains(p, "/node_modules/") {
					continue
				}
				// skip bundled files
				if strings.Contains(p, "bundled:") {
					continue
				}
				files = append(files, file)
			}
		}
	}
	slices.SortFunc(files, func(a *ast.SourceFile, b *ast.SourceFile) int {
		return len(b.Text()) - len(a.Text())
	})

	// Collect diagnostics
	var diagnostics []ipc.Diagnostic
	var diagnosticsLock sync.Mutex
	errorsCount := 0

	// Create collector function
	diagnosticCollector := func(d rule.RuleDiagnostic) {
		diagnosticsLock.Lock()
		defer diagnosticsLock.Unlock()

		diagnosticStart := d.Range.Pos()
		diagnosticEnd := d.Range.End()

		startLine, startColumn := scanner.GetLineAndCharacterOfPosition(d.SourceFile, diagnosticStart)
		endLine, endColumn := scanner.GetLineAndCharacterOfPosition(d.SourceFile, diagnosticEnd)

		diagnostic := ipc.Diagnostic{
			RuleName:  d.RuleName,
			MessageID: d.Message.Id,
			Message:   d.Message.Description,
			FilePath:  tspath.ConvertToRelativePath(d.SourceFile.FileName(), comparePathOptions),
			Range: ipc.Range{
				Start: ipc.Position{
					Line:   startLine + 1, // Convert to 1-based indexing
					Column: startColumn + 1,
				},
				End: ipc.Position{
					Line:   endLine + 1,
					Column: endColumn + 1,
				},
			},
		}

		diagnostics = append(diagnostics, diagnostic)
		errorsCount++
	}

	// Run linter
	err = linter.RunLinter(
		programs,
		false, // Don't use single-threaded mode for IPC
		files,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			return utils.Map(rulesWithConfig, func(ruleWithConfig rslintconfig.EnabledRuleWithConfig) linter.ConfiguredRule {
				return linter.ConfiguredRule{
					Name: ruleWithConfig.Rule.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						// Pass options directly as received - if it's wrapped in {"value": X}, unwrap it
						options := ruleWithConfig.Config.Options
						var finalOptions interface{} = options
						if len(options) == 1 {
							if val, hasValue := options["value"]; hasValue {
								// This was a simple option that got wrapped, unwrap it
								finalOptions = val
							}
						}
						return ruleWithConfig.Rule.Run(ctx, finalOptions)
					},
				}
			})
		},
		diagnosticCollector,
	)
	if err != nil {
		return nil, fmt.Errorf("error running linter: %v", err)
	}
	if diagnostics == nil {
		diagnostics = []ipc.Diagnostic{}
	}
	// Create response
	return &ipc.LintResponse{
		Diagnostics: diagnostics,
		ErrorCount:  errorsCount,
		FileCount:   len(files),
		RuleCount:   len(rulesWithConfig),
	}, nil
}

// runAPI runs the linter in IPC mode
func runAPI() int {
	handler := &IPCHandler{}
	service := ipc.NewService(os.Stdin, os.Stdout, handler)

	if err := service.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error in IPC mode: %v\n", err)
		return 1
	}
	return 0
}
