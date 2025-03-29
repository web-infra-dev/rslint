package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"sync"

	"none.none/tsgolint/internal/linter"
	"none.none/tsgolint/internal/rule"
	"none.none/tsgolint/internal/utils"

	"none.none/tsgolint/internal/rules/await_thenable"
	"none.none/tsgolint/internal/rules/no_array_delete"
	"none.none/tsgolint/internal/rules/no_base_to_string"
	"none.none/tsgolint/internal/rules/no_duplicate_type_constituents"
	"none.none/tsgolint/internal/rules/no_floating_promises"
	"none.none/tsgolint/internal/rules/no_for_in_array"
	"none.none/tsgolint/internal/rules/no_implied_eval"
	"none.none/tsgolint/internal/rules/no_misused_promises"
	"none.none/tsgolint/internal/rules/no_redundant_type_constituents"
	"none.none/tsgolint/internal/rules/no_unnecessary_type_assertion"
	"none.none/tsgolint/internal/rules/no_unsafe_argument"
	"none.none/tsgolint/internal/rules/no_unsafe_assignment"
	"none.none/tsgolint/internal/rules/no_unsafe_call"
	"none.none/tsgolint/internal/rules/no_unsafe_enum_comparison"
	"none.none/tsgolint/internal/rules/no_unsafe_member_access"
	"none.none/tsgolint/internal/rules/no_unsafe_return"
	"none.none/tsgolint/internal/rules/no_unsafe_unary_minus"
	"none.none/tsgolint/internal/rules/only_throw_error"
	"none.none/tsgolint/internal/rules/prefer_promise_reject_errors"
	"none.none/tsgolint/internal/rules/require_await"
	"none.none/tsgolint/internal/rules/restrict_plus_operands"
	"none.none/tsgolint/internal/rules/restrict_template_expressions"
	"none.none/tsgolint/internal/rules/switch_exhaustiveness_check"
	"none.none/tsgolint/internal/rules/unbound_method"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

var traceOut = flag.String("trace", "", "File to put trace to")
var cpuprofOut = flag.String("cpuprof", "", "File to put cpu profiling to")

func main() {
	flag.Parse()

	if *traceOut != "" {
		f, err := os.Create(*traceOut)
		if err != nil {
		    panic(err)
		}
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
	}
	if *cpuprofOut != "" {
		f, err := os.Create(*cpuprofOut + ".pg.gz")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	configFileName := tspath.ResolvePath(currentDirectory, "tsconfig.eslint.json")

	var files []string
	// if args := os.Args[1:]; len(args) > 0 {
	// 	files = args 
	// }

		var diagnosticsMu sync.Mutex
		diagnostics := make([]rule.RuleDiagnostic, 0, 3)


		var rules = []rule.Rule{
			await_thenable.AwaitThenableRule,
			no_array_delete.NoArrayDeleteRule,
			no_base_to_string.NoBaseToStringRule,
			no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule,
			no_floating_promises.NoFloatingPromisesRule,
			no_for_in_array.NoForInArrayRule,
			no_implied_eval.NoImpliedEvalRule,
			no_misused_promises.NoMisusedPromisesRule,
			no_redundant_type_constituents.NoRedundantTypeConstituentsRule,
			no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule,
			no_unsafe_argument.NoUnsafeArgumentRule,
			no_unsafe_assignment.NoUnsafeAssignmentRule,
			no_unsafe_call.NoUnsafeCallRule,
			no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule,
			no_unsafe_member_access.NoUnsafeMemberAccessRule,
			no_unsafe_return.NoUnsafeReturnRule,
			no_unsafe_unary_minus.NoUnsafeUnaryMinusRule,
			only_throw_error.OnlyThrowErrorRule,
			prefer_promise_reject_errors.PreferPromiseRejectErrorsRule,
			require_await.RequireAwaitRule,
			restrict_plus_operands.RestrictPlusOperandsRule,
			restrict_template_expressions.RestrictTemplateExpressionsRule,
			switch_exhaustiveness_check.SwitchExhaustivenessCheckRule,
			unbound_method.UnboundMethodRule,
		}

		// rules = utils.Filter(rules, func(r rule.Rule) bool {
		// 	return r.Name == "restrict-template-expressions"
		// })


		fs := bundled.WrapFS(osvfs.FS())
		err = linter.RunLinter(
			false,
			fs,
			files,
			func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
				return utils.Map(rules, func(r rule.Rule) linter.ConfiguredRule {
					return linter.ConfiguredRule{
						Name: r.Name,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return r.Run(ctx, nil)
						},
					}
				})
			},
			currentDirectory,
			configFileName,
			func(diagnostic rule.RuleDiagnostic) {
				diagnosticsMu.Lock()
				defer diagnosticsMu.Unlock()

				diagnostics = append(diagnostics, diagnostic)
			},
		)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running linter: %v\n", err)
			return
			// os.Exit(1)
		}

	for _, d := range diagnostics {
		line, character := scanner.GetLineAndCharacterOfPosition(d.SourceFile, d.Range.Pos())
		fmt.Printf("%v\n   %v:%v   (error)   %v (%v)\n", tspath.GetRelativePathFromDirectory(currentDirectory, string(d.SourceFile.Path()), tspath.ComparePathsOptions{
			UseCaseSensitiveFileNames: true,
			CurrentDirectory: currentDirectory,
		}), line+1, character+1, d.Message, d.RuleName)
	}
	fmt.Printf("Total errors: %v\n", len(diagnostics))
}
