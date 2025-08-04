package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules/adjacent_overload_signatures"
	"github.com/web-infra-dev/rslint/internal/rules/array_type"
	"github.com/web-infra-dev/rslint/internal/rules/await_thenable"
	"github.com/web-infra-dev/rslint/internal/rules/class_literal_property_style"
	"github.com/web-infra-dev/rslint/internal/rules/no_array_delete"
	"github.com/web-infra-dev/rslint/internal/rules/no_base_to_string"
	"github.com/web-infra-dev/rslint/internal/rules/no_confusing_void_expression"
	"github.com/web-infra-dev/rslint/internal/rules/no_duplicate_type_constituents"
	"github.com/web-infra-dev/rslint/internal/rules/no_floating_promises"
	"github.com/web-infra-dev/rslint/internal/rules/no_for_in_array"
	"github.com/web-infra-dev/rslint/internal/rules/no_implied_eval"
	"github.com/web-infra-dev/rslint/internal/rules/no_meaningless_void_operator"
	"github.com/web-infra-dev/rslint/internal/rules/no_misused_promises"
	"github.com/web-infra-dev/rslint/internal/rules/no_misused_spread"
	"github.com/web-infra-dev/rslint/internal/rules/no_mixed_enums"
	"github.com/web-infra-dev/rslint/internal/rules/no_redundant_type_constituents"
	"github.com/web-infra-dev/rslint/internal/rules/no_unnecessary_boolean_literal_compare"
	"github.com/web-infra-dev/rslint/internal/rules/no_unnecessary_template_expression"
	"github.com/web-infra-dev/rslint/internal/rules/no_unnecessary_type_arguments"
	"github.com/web-infra-dev/rslint/internal/rules/no_unnecessary_type_assertion"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_argument"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_assignment"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_call"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_enum_comparison"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_member_access"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_return"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_type_assertion"
	"github.com/web-infra-dev/rslint/internal/rules/no_unsafe_unary_minus"
	"github.com/web-infra-dev/rslint/internal/rules/non_nullable_type_assertion_style"
	"github.com/web-infra-dev/rslint/internal/rules/only_throw_error"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_promise_reject_errors"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_reduce_type_parameter"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_return_this_type"
	"github.com/web-infra-dev/rslint/internal/rules/promise_function_async"
	"github.com/web-infra-dev/rslint/internal/rules/related_getter_setter_pairs"
	"github.com/web-infra-dev/rslint/internal/rules/require_array_sort_compare"
	"github.com/web-infra-dev/rslint/internal/rules/require_await"
	"github.com/web-infra-dev/rslint/internal/rules/restrict_plus_operands"
	"github.com/web-infra-dev/rslint/internal/rules/restrict_template_expressions"
	"github.com/web-infra-dev/rslint/internal/rules/return_await"
	"github.com/web-infra-dev/rslint/internal/rules/switch_exhaustiveness_check"
	"github.com/web-infra-dev/rslint/internal/rules/unbound_method"
	"github.com/web-infra-dev/rslint/internal/rules/use_unknown_in_catch_callback_variable"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IPCHandler implements the ipc.Handler interface
type IPCHandler struct{}

// HandleLint handles lint requests in IPC mode
func (h *IPCHandler) HandleLint(req api.LintRequest) (*api.LintResponse, error) {

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

	// Create rules
	var origin_rules = []rule.Rule{
		adjacent_overload_signatures.AdjacentOverloadSignaturesRule,
		array_type.ArrayTypeRule,
		await_thenable.AwaitThenableRule,
		class_literal_property_style.ClassLiteralPropertyStyleRule,
		no_array_delete.NoArrayDeleteRule,
		no_base_to_string.NoBaseToStringRule,
		no_confusing_void_expression.NoConfusingVoidExpressionRule,
		no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule,
		no_floating_promises.NoFloatingPromisesRule,
		no_for_in_array.NoForInArrayRule,
		no_implied_eval.NoImpliedEvalRule,
		no_meaningless_void_operator.NoMeaninglessVoidOperatorRule,
		no_misused_promises.NoMisusedPromisesRule,
		no_misused_spread.NoMisusedSpreadRule,
		no_mixed_enums.NoMixedEnumsRule,
		no_redundant_type_constituents.NoRedundantTypeConstituentsRule,
		no_unnecessary_boolean_literal_compare.NoUnnecessaryBooleanLiteralCompareRule,
		no_unnecessary_template_expression.NoUnnecessaryTemplateExpressionRule,
		no_unnecessary_type_arguments.NoUnnecessaryTypeArgumentsRule,
		no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule,
		no_unsafe_argument.NoUnsafeArgumentRule,
		no_unsafe_assignment.NoUnsafeAssignmentRule,
		no_unsafe_call.NoUnsafeCallRule,
		no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule,
		no_unsafe_member_access.NoUnsafeMemberAccessRule,
		no_unsafe_return.NoUnsafeReturnRule,
		no_unsafe_type_assertion.NoUnsafeTypeAssertionRule,
		no_unsafe_unary_minus.NoUnsafeUnaryMinusRule,
		non_nullable_type_assertion_style.NonNullableTypeAssertionStyleRule,
		only_throw_error.OnlyThrowErrorRule,
		prefer_promise_reject_errors.PreferPromiseRejectErrorsRule,
		prefer_reduce_type_parameter.PreferReduceTypeParameterRule,
		prefer_return_this_type.PreferReturnThisTypeRule,
		promise_function_async.PromiseFunctionAsyncRule,
		related_getter_setter_pairs.RelatedGetterSetterPairsRule,
		require_array_sort_compare.RequireArraySortCompareRule,
		require_await.RequireAwaitRule,
		restrict_plus_operands.RestrictPlusOperandsRule,
		restrict_template_expressions.RestrictTemplateExpressionsRule,
		return_await.ReturnAwaitRule,
		switch_exhaustiveness_check.SwitchExhaustivenessCheckRule,
		unbound_method.UnboundMethodRule,
		use_unknown_in_catch_callback_variable.UseUnknownInCatchCallbackVariableRule,
	}
	type RuleWithOption struct {
		rule   rule.Rule
		option interface{}
	}
	rulesWithOptions := []RuleWithOption{}
	// filter rule based on request.RuleOptions
	if len(req.RuleOptions) > 0 {
		for _, r := range origin_rules {
			if option, ok := req.RuleOptions[r.Name]; ok {
				rulesWithOptions = append(rulesWithOptions, RuleWithOption{
					rule:   r,
					option: option,
				})
			}
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

	// Collect diagnostics
	var diagnostics []api.Diagnostic
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

		diagnostic := api.Diagnostic{
			RuleName:  d.RuleName,
			MessageId: d.Message.Id,
			Message:   d.Message.Description,
			FilePath:  tspath.ConvertToRelativePath(d.SourceFile.FileName(), comparePathOptions),
			Range: api.Range{
				Start: api.Position{
					Line:   startLine + 1, // Convert to 1-based indexing
					Column: startColumn + 1,
				},
				End: api.Position{
					Line:   endLine + 1,
					Column: endColumn + 1,
				},
			},
		}

		diagnostics = append(diagnostics, diagnostic)
		errorsCount++
	}

	// Run linter
	lintedFilesCount, err := linter.RunLinter(
		programs,
		false, // Don't use single-threaded mode for IPC
		nil,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			return utils.Map(rulesWithOptions, func(r RuleWithOption) linter.ConfiguredRule {

				return linter.ConfiguredRule{
					Name: r.rule.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.rule.Run(ctx, r.option)
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
		diagnostics = []api.Diagnostic{}
	}
	// Create response
	return &api.LintResponse{
		Diagnostics: diagnostics,
		ErrorCount:  errorsCount,
		FileCount:   int(lintedFilesCount),
		RuleCount:   len(rulesWithOptions),
	}, nil
}

// runAPI runs the linter in IPC mode
func runAPI() int {
	handler := &IPCHandler{}
	service := api.NewService(os.Stdin, os.Stdout, handler)

	if err := service.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error in IPC mode: %v\n", err)
		return 1
	}
	return 0
}
