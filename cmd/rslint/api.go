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
	"github.com/typescript-eslint/rslint/internal/config"
	"github.com/typescript-eslint/rslint/internal/linter"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/rules/array_type"
	"github.com/typescript-eslint/rslint/internal/rules/await_thenable"
	"github.com/typescript-eslint/rslint/internal/rules/ban_ts_comment"
	"github.com/typescript-eslint/rslint/internal/rules/ban_tslint_comment"
	"github.com/typescript-eslint/rslint/internal/rules/class_literal_property_style"
	"github.com/typescript-eslint/rslint/internal/rules/class_methods_use_this"
	"github.com/typescript-eslint/rslint/internal/rules/consistent_generic_constructors"
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
	"github.com/typescript-eslint/rslint/internal/utils"
)

// IPCHandler implements the ipc.Handler interface
type IPCHandler struct{}

// HandleLint handles lint requests in IPC mode
func (h *IPCHandler) HandleLint(req ipc.LintRequest) (*ipc.LintResponse, error) {
	var tsconfig string
	if req.TSConfig != "" {
		tsconfig = req.TSConfig
	}

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

	// Handle tsconfig
	var configFileName string
	if tsconfig == "" {
		configFileName = tspath.ResolvePath(currentDirectory, "tsconfig.json")
		if !fs.FileExists(configFileName) {
			fs = utils.NewOverlayVFS(fs, map[string]string{
				configFileName: "{}",
			})
		}
	} else {
		configFileName = tspath.ResolvePath(currentDirectory, tsconfig)
		if !fs.FileExists(configFileName) {
			return nil, fmt.Errorf("error: tsconfig %q doesn't exist", tsconfig)
		}
	}
	currentDirectory = tspath.GetDirectoryPath(configFileName)

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
	}

	// Build rules with their configurations from request.RuleOptions
	var rulesWithConfig []config.EnabledRuleWithConfig
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
					rulesWithConfig = append(rulesWithConfig, config.EnabledRuleWithConfig{
						Rule: rule,
						Config: &config.RuleConfig{
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
			rulesWithConfig = append(rulesWithConfig, config.EnabledRuleWithConfig{
				Rule: rule,
				Config: &config.RuleConfig{
					Level: "error",
				},
			})
		}
	}

	// Create compiler host
	host := utils.CreateCompilerHost(currentDirectory, fs)
	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          host.GetCurrentDirectory(),
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
	}

	// Create program
	program, err := utils.CreateProgram(false, fs, currentDirectory, configFileName, host)
	if err != nil {
		return nil, fmt.Errorf("error creating TS program: %v", err)
	}

	// Find source files
	files := []*ast.SourceFile{}

	// If specific files are provided, use those
	if len(req.Files) > 0 {
		for _, filePath := range req.Files {
			absPath := tspath.ResolvePath(currentDirectory, filePath)
			sourceFile := program.GetSourceFile(absPath)
			if sourceFile != nil {
				files = append(files, sourceFile)
			}
		}
	} else {
		// Otherwise use all source files
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
		[]*compiler.Program{program},
		false, // Don't use single-threaded mode for IPC
		files,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			return utils.Map(rulesWithConfig, func(ruleWithConfig config.EnabledRuleWithConfig) linter.ConfiguredRule {
				return linter.ConfiguredRule{
					Name: ruleWithConfig.Rule.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return ruleWithConfig.Rule.Run(ctx, ruleWithConfig.Config.Options)
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
