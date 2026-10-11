package linter

import (
	"context"
	"sync"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// nativeObservationWork contains only consumable execution tasks and the
// requested output metadata. It never retains the producer Generation or plan.
type nativeObservationWork struct {
	execution              *lintExecution
	diagnostics            []rule.RuleDiagnostic
	files                  []LintedFile
	targetPath             func(string) string
	demand                 rule.EditDemand
	preserveSourceIdentity bool
	hasTargetSyntaxErrors  bool
}

func prepareNativeObservation(
	generation Generation,
	plan *LintPlan,
	demand rule.EditDemand,
	lintedFiles []LintedFile,
	preserveSourceIdentity bool,
) (*nativeObservationWork, error) {
	native := generation.Native
	options := RunLinterOptions{
		LintPlan:       plan,
		SingleThreaded: native.SingleThreaded,
		Cwd:            native.Cwd,
		TypeCheck:      native.TypeCheck,
		Timing:         native.Timing,
		Consumer:       rule.DiagnosticConsumer{Demand: demand},
	}
	if plan == nil {
		options.TypeCheckOnlyPrograms = native.Programs
	}
	execution, err := prepareLintExecution(options)
	if err != nil {
		return nil, err
	}
	work := &nativeObservationWork{
		execution:              execution,
		files:                  lintedFiles,
		targetPath:             generation.Target.Path,
		demand:                 demand,
		preserveSourceIdentity: preserveSourceIdentity,
		hasTargetSyntaxErrors:  plan.HasSyntacticDiagnostics(),
	}
	if plan != nil {
		work.diagnostics = plan.SyntacticDiagnostics(native.TypeCheck)
	}
	if !preserveSourceIdentity {
		for index := range work.diagnostics {
			work.diagnostics[index].SourceFile = diagnosticTextSource(work.diagnostics[index].SourceFile)
		}
	}
	return work, nil
}

// runNativeObservation projects sources before the result channel can become
// an AST owner. Fix planning keeps exact source identity until text freezing.
func runNativeObservation(ctx context.Context, work *nativeObservationWork) (NativeObservation, error) {
	if err := ctx.Err(); err != nil {
		return NativeObservation{}, err
	}
	diagnostics := work.diagnostics
	work.diagnostics = nil
	consumer := rule.DiagnosticConsumer{Demand: work.demand}
	var diagnosticsWait sync.WaitGroup
	finishDiagnostics := func() {}
	if work.execution.options.SingleThreaded {
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
	if !work.preserveSourceIdentity {
		report := consumer.Report
		consumer.Report = func(diagnostic rule.RuleDiagnostic) {
			diagnostic.SourceFile = diagnosticTextSource(diagnostic.SourceFile)
			report(diagnostic)
		}
	}
	var finishOnce sync.Once
	finish := func() { finishOnce.Do(finishDiagnostics) }
	defer finish()
	lintResult := work.execution.run(consumer)
	finish()

	for index := range diagnostics {
		diagnostics[index].FilePath = projectTargetPath(work.targetPath, diagnostics[index].FilePath)
	}
	result := NativeObservation{
		Diagnostics:           diagnostics,
		Lint:                  lintResult,
		Files:                 work.files,
		HasTargetSyntaxErrors: work.hasTargetSyntaxErrors,
	}
	return result, ctx.Err()
}
