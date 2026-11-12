package linter

import (
	"context"
	"sync"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// nativeObservationUnit keeps diagnostic groups private until every load has
// joined. It owns no Programs; fix planning may retain source identity until
// the unit freezes its diagnostic sources.
type nativeObservationUnit struct {
	syntacticDiagnostics  []rule.RuleDiagnostic
	diagnostics           []rule.RuleDiagnostic
	typeDiagnostics       [][]collectedTypeCheckDiagnostic
	lint                  *LintResult
	files                 []LintedFile
	hasTargetSyntaxErrors bool
}

// nativeObservationWork contains only consumable execution tasks and the
// requested output metadata. It never retains the producer Generation or plan.
type nativeObservationWork struct {
	execution              *lintExecution
	diagnostics            []rule.RuleDiagnostic
	files                  []LintedFile
	targetPath             func(string) string
	typeCheckPaths         map[string]string
	demand                 rule.EditDemand
	preserveSourceIdentity bool
	hasTargetSyntaxErrors  bool
}

func (work *nativeObservationWork) close() {
	if work.execution != nil {
		work.execution.close()
	}
}

func prepareNativeObservation(
	generation Generation,
	plan *LintPlan,
	demand rule.EditDemand,
	lintedFiles []LintedFile,
	preserveSourceIdentity bool,
	programExecution ProgramExecution,
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
	execution, err := prepareLintExecution(options, programExecution, !preserveSourceIdentity)
	if err != nil {
		return nil, err
	}
	work := &nativeObservationWork{
		execution:              execution,
		files:                  lintedFiles,
		targetPath:             generation.Target.Path,
		typeCheckPaths:         generation.Target.TypeCheckPaths,
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
func runNativeObservation(ctx context.Context, work *nativeObservationWork) (nativeObservationUnit, error) {
	if err := ctx.Err(); err != nil {
		return nativeObservationUnit{}, err
	}
	syntacticDiagnostics := work.diagnostics
	work.diagnostics = nil
	var diagnostics []rule.RuleDiagnostic
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
	if !work.preserveSourceIdentity || len(work.execution.options.directives) > 0 {
		report := consumer.Report
		consumer.Report = func(diagnostic rule.RuleDiagnostic) {
			if !work.preserveSourceIdentity || work.execution.options.directives[diagnostic.FilePath] != nil {
				diagnostic.SourceFile = diagnosticTextSource(diagnostic.SourceFile)
			}
			report(diagnostic)
		}
	}
	var finishOnce sync.Once
	finish := func() { finishOnce.Do(finishDiagnostics) }
	defer finish()
	lintResult, typeDiagnostics := work.execution.run(consumer)
	finish()
	for sourcePath, state := range work.execution.options.directives {
		diagnostics = append(diagnostics, state.complete(true, ctx.Err() != nil, sourcePath)...)
	}

	for _, group := range [][]rule.RuleDiagnostic{syntacticDiagnostics, diagnostics} {
		for index := range group {
			group[index].FilePath = projectTargetPath(work.targetPath, group[index].FilePath)
		}
	}
	for _, collected := range typeDiagnostics {
		for index := range collected {
			entry := &collected[index]
			path := work.typeCheckPaths[utils.NormalizeAbsoluteDrive(entry.key.path)]
			if path == "" {
				path = projectTargetPath(work.targetPath, entry.ruleDiagnostic.FilePath)
			}
			entry.ruleDiagnostic.FilePath = path
		}
	}
	result := nativeObservationUnit{
		syntacticDiagnostics:  syntacticDiagnostics,
		diagnostics:           diagnostics,
		typeDiagnostics:       typeDiagnostics,
		lint:                  lintResult,
		files:                 work.files,
		hasTargetSyntaxErrors: work.hasTargetSyntaxErrors,
	}
	return result, ctx.Err()
}
