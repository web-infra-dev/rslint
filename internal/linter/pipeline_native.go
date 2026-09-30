package linter

import (
	"context"
	"sync"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// runNativeObservation executes the native half of one already prepared lint
// generation and projects its diagnostics and selected files into target path
// space.
func runNativeObservation(
	ctx context.Context,
	generation Generation,
	plan *LintPlan,
	demand rule.EditDemand,
	lintedFiles []LintedFile,
) (NativeObservation, error) {
	if err := ctx.Err(); err != nil {
		return NativeObservation{}, err
	}
	// A nil plan is the explicit type-check-only/empty-generation shape: Phase 1
	// has no target projection, while RunLinter may still execute Phase 2 over
	// NativeGeneration.Programs.
	var diagnostics []rule.RuleDiagnostic
	if plan != nil {
		diagnostics = plan.SyntacticDiagnostics(generation.Native.TypeCheck)
	}
	runOptions := generation.runLinterOptions(plan)
	consumer := rule.DiagnosticConsumer{Demand: demand}
	var diagnosticsWait sync.WaitGroup
	finishDiagnostics := func() {}
	if runOptions.SingleThreaded {
		consumer.Report = func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}
	} else {
		diagnosticsChannel := make(chan rule.RuleDiagnostic, 4096)
		diagnosticsWait.Add(1)
		go func() {
			defer diagnosticsWait.Done()
			for diagnostic := range diagnosticsChannel {
				diagnostics = append(diagnostics, diagnostic)
			}
		}()
		consumer.Report = func(diagnostic rule.RuleDiagnostic) {
			diagnosticsChannel <- diagnostic
		}
		finishDiagnostics = func() {
			close(diagnosticsChannel)
			diagnosticsWait.Wait()
		}
	}
	var finishOnce sync.Once
	finish := func() { finishOnce.Do(finishDiagnostics) }
	defer finish()
	// Only fix-bearing diagnostics participate in source-identity validation.
	// Project all others before queueing so ordinary files can be collected
	// during the same observation, including an autofix observation.
	report := consumer.Report
	var projections sync.Map // weak.Pointer[ast.SourceFile] -> *diagnosticSource
	consumer.Report = func(diagnostic rule.RuleDiagnostic) {
		if source, ok := diagnostic.SourceFile.(*ast.SourceFile); ok && source != nil && len(diagnostic.Fixes()) == 0 {
			key := weak.Make(source)
			projection, found := projections.Load(key)
			if !found {
				projection, _ = projections.LoadOrStore(key, newDiagnosticSource(source))
			}
			if projected, ok := projection.(*diagnosticSource); ok {
				diagnostic.SourceFile = projected
			}
		}
		report(diagnostic)
	}
	runOptions.Consumer = consumer
	lintResult, err := RunLinter(runOptions)
	finish()

	for index := range diagnostics {
		diagnostics[index].FilePath = projectTargetPath(generation.Target.Path, diagnostics[index].FilePath)
	}
	result := NativeObservation{
		Diagnostics:           diagnostics,
		Lint:                  lintResult,
		Files:                 lintedFiles,
		HasTargetSyntaxErrors: plan != nil && plan.HasSyntacticDiagnostics(),
	}
	return result, joinContextError(err, ctx)
}

func (generation Generation) runLinterOptions(plan *LintPlan) RunLinterOptions {
	native := generation.Native
	options := RunLinterOptions{
		SingleThreaded: native.SingleThreaded,
		Cwd:            native.Cwd,
		TypeCheck:      native.TypeCheck,
		Timing:         native.Timing,
	}
	if plan == nil {
		options.TypeCheckOnlyPrograms = native.Programs
	} else {
		options.LintPlan = plan
	}
	return options
}
