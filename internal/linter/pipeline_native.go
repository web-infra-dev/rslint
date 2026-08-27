package linter

import (
	"context"
	"sync"

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
