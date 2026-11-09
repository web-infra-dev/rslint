package linter

import (
	"context"
	"errors"
	"fmt"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// observationExecution is the pipeline-private result of one immutable source
// generation observation.
type observationExecution struct {
	observation   ObservationResult
	fixTexts      fixTextSnapshot
	pluginOutcome *EslintPluginDispatchOutcome
	deferredTask  *pluginTask
}

// observationUnitExecution owns one load's private output. Only the merge
// boundary can assemble its diagnostic groups into an ObservationResult.
type observationUnitExecution struct {
	native                nativeObservationUnit
	pluginKind            pluginObservationKind
	pluginOutcome         EslintPluginDispatchOutcome
	fixTexts              fixTextSnapshot
	deferredTask          *pluginTask
	programRootIdentities []string
}

// preparedObservation is the ownership handoff between whole-generation
// validation and execution. Only actual fix planning retains a source reader;
// the executor retains neither the producer Generation nor its LintPlan.
type preparedObservation struct {
	native                *nativeObservationWork
	plugin                pluginTask
	lease                 *releaseLease
	readFixText           sourceTextReader
	targetSyntaxBlocked   bool
	programRootIdentities []string
}

func prepareObservation(
	ctx context.Context,
	acquire func(context.Context) (Generation, ReleaseFunc, error),
	snapshot SourceSnapshot,
	policy ObservationPolicy,
	dispatcher EslintPluginDispatcher,
	planChanges bool,
	stopOnTargetSyntaxErrors bool,
	reservation *programReservation,
	sourceReads pluginSourceReadSlots,
) (preparedObservation, error) {
	if err := ctx.Err(); err != nil {
		return preparedObservation{}, err
	}
	generation, release, err := acquire(ctx)
	if err != nil {
		return preparedObservation{}, &GenerationLoadError{Err: err}
	}
	lease := &releaseLease{release: release}
	transferred := false
	defer func() {
		if !transferred {
			lease.close()
		}
	}()
	if err := ctx.Err(); err != nil {
		return preparedObservation{}, err
	}

	var plan *LintPlan
	if generation.Native.RulesForFile != nil {
		plan, err = PrepareLintPlanContext(ctx, PrepareLintPlanOptions{
			Programs:         generation.Native.Programs,
			TargetsByProgram: generation.Native.TargetsByProgram,
			SingleThreaded:   generation.Native.SingleThreaded,
			GetRulesForFile:  generation.Native.RulesForFile,
		})
		if err != nil {
			return preparedObservation{}, fmt.Errorf("linter pipeline: prepare lint plan: %w", err)
		}
	}
	lintedFiles, err := projectGenerationTargets(
		ctx,
		generation,
		plan,
		snapshot,
		policy.Demand,
	)
	if err != nil {
		return preparedObservation{}, err
	}
	detachedPlugin := policy.Plugin != PluginConcurrentJoined
	targetSyntaxBlocked := stopOnTargetSyntaxErrors && plan != nil && plan.HasSyntacticDiagnostics()
	pluginWork := pluginTask{failure: policy.PluginFailure}
	if !targetSyntaxBlocked {
		pluginWork, err = materializePluginTask(ctx, plan, generation, snapshot, policy, detachedPlugin, planChanges, sourceReads)
		if err != nil {
			return preparedObservation{}, err
		}
	}
	if len(pluginWork.inputs) > 0 && policy.Plugin != pluginProgressiveAfterNative && dispatcher == nil {
		return preparedObservation{}, errors.New("linter pipeline: joined plugin work requires a dispatcher")
	}
	nativeWork, err := prepareNativeObservation(generation, plan, policy.Demand.Native, lintedFiles, planChanges, policy.ProgramExecution)
	if err != nil {
		return preparedObservation{}, err
	}
	nativeWork.execution.options.reservation = reservation
	for index := range pluginWork.inputs {
		input := &pluginWork.inputs[index]
		if input.SourceFile != nil {
			input.DiagnosticSource = input.SourceFile
			if !planChanges {
				input.DiagnosticSource = diagnosticTextSource(input.SourceFile)
			}
		}
	}
	prepared := preparedObservation{
		native:                nativeWork,
		plugin:                pluginWork,
		lease:                 lease,
		targetSyntaxBlocked:   targetSyntaxBlocked,
		programRootIdentities: generation.ProgramRootIdentities,
	}
	if planChanges {
		prepared.readFixText = generation.Target.ReadText
	}
	transferred = true
	return prepared, nil
}

func executePreparedObservation(
	ctx context.Context,
	prepared preparedObservation,
	snapshot SourceSnapshot,
	policy ObservationPolicy,
	dispatcher EslintPluginDispatcher,
	planChanges bool,
	stopOnTargetSyntaxErrors bool,
) (observationUnitExecution, error) {
	var err error
	lease := prepared.lease
	defer lease.close()
	nativeWork := prepared.native
	defer nativeWork.close()
	pluginWork := prepared.plugin
	prepared.plugin = pluginTask{}
	targetSyntaxBlocked := prepared.targetSyntaxBlocked
	execution := observationUnitExecution{
		programRootIdentities: prepared.programRootIdentities,
	}
	switch policy.Plugin {
	case PluginConcurrentJoined:
		execution, runErr := executeConcurrentObservation(
			ctx,
			nativeWork,
			pluginWork,
			dispatcher,
			execution,
		)
		if runErr == nil && planChanges && !targetSyntaxBlocked {
			diagnostics := execution.fixableDiagnostics(policy.PluginFailure)
			execution.fixTexts, runErr = freezeFixTextsForDiagnostics(
				ctx,
				prepared.readFixText,
				snapshot,
				diagnostics,
			)
		}
		execution.detachDiagnosticSources()
		lease.close()
		return execution, joinContextError(runErr, ctx)
	case PluginAfterNativeJoined:
		native, nativeErr := runNativeObservation(ctx, nativeWork)
		execution.native = native
		if nativeErr != nil {
			execution.detachDiagnosticSources()
			lease.close()
			return execution, nativeErr
		}
		if stopOnTargetSyntaxErrors && native.hasTargetSyntaxErrors {
			pluginWork.fixCandidates = nil
			execution.detachDiagnosticSources()
			lease.close()
			execution.pluginKind = pluginObservationNone
			return execution, ctx.Err()
		}
		if planChanges {
			candidateSources, sourceErr := fixSourcesFromDiagnostics(native.diagnostics)
			if sourceErr != nil {
				return execution, sourceErr
			}
			if pluginWork.collectFixes {
				for _, candidate := range pluginWork.fixCandidates {
					if candidate.path == "" {
						return execution, errors.New("linter pipeline: plugin fix target path must not be empty")
					}
					if previous, duplicate := candidateSources[candidate.path]; duplicate && previous != candidate.source {
						return execution, fmt.Errorf("linter pipeline: duplicate fix target %q", candidate.path)
					}
					candidateSources[candidate.path] = candidate.source
				}
			}
			execution.fixTexts, err = freezeFixTexts(ctx, prepared.readFixText, snapshot, candidateSources)
			if err != nil {
				return execution, err
			}
		}
		pluginWork.fixCandidates = nil
		prepared.readFixText = nil
		execution.detachDiagnosticSources()
		// Detached plugin inputs and frozen fix text no longer reference generation
		// state, so watcher/Program resources are released before a reverse request
		// can block.
		lease.close()
		if err := ctx.Err(); err != nil {
			return execution, err
		}
		if len(pluginWork.inputs) == 0 {
			execution.pluginKind = pluginObservationNone
			return execution, nil
		}
		outcome := pluginWork.runRaw(ctx, dispatcher)
		execution.pluginKind = pluginObservationJoined
		execution.pluginOutcome = outcome
		if err := ctx.Err(); err != nil {
			return execution, err
		}
		if planChanges {
			diagnostics := execution.fixableDiagnostics(policy.PluginFailure)
			execution.fixTexts, err = retainFixTextsForDiagnostics(
				execution.fixTexts,
				diagnostics,
			)
			if err != nil {
				return execution, err
			}
		}
		return execution, nil
	case pluginProgressiveAfterNative:
		native, nativeErr := runNativeObservation(ctx, nativeWork)
		execution.native = native
		execution.detachDiagnosticSources()
		lease.close()
		if nativeErr != nil {
			return execution, nativeErr
		}
		if err := ctx.Err(); err != nil {
			return execution, err
		}
		if stopOnTargetSyntaxErrors && native.hasTargetSyntaxErrors {
			execution.pluginKind = pluginObservationNone
			return execution, nil
		}
		if len(pluginWork.inputs) == 0 {
			execution.pluginKind = pluginObservationNone
			return execution, nil
		}
		execution.pluginKind = pluginObservationProgressive
		execution.deferredTask = &pluginWork
		return execution, nil
	default:
		return execution, errors.New("linter pipeline: plugin execution policy is invalid")
	}
}

func executeConcurrentObservation(
	ctx context.Context,
	nativeWork *nativeObservationWork,
	pluginTask pluginTask,
	dispatcher EslintPluginDispatcher,
	execution observationUnitExecution,
) (observationUnitExecution, error) {
	var (
		pluginCh     <-chan EslintPluginDispatchOutcome
		cancelPlugin context.CancelFunc
		pluginJoined bool
	)
	var pluginPanic workerPanic
	if len(pluginTask.inputs) > 0 {
		pluginCtx, cancel := context.WithCancel(ctx)
		cancelPlugin = cancel
		ch := make(chan EslintPluginDispatchOutcome, 1)
		pluginCh = ch
		go func() {
			var outcome EslintPluginDispatchOutcome
			defer func() { ch <- outcome }()
			pluginPanic.run(func() { outcome = pluginTask.runRaw(pluginCtx, dispatcher) })
		}()
	}
	// This defer runs before executeObservation's lease defer, so panic,
	// cancellation, and native errors cancel and join plugin work before release.
	defer func() {
		if cancelPlugin != nil {
			cancelPlugin()
		}
		if pluginCh != nil && !pluginJoined {
			<-pluginCh
		}
	}()

	native, nativeErr := runNativeObservation(ctx, nativeWork)
	execution.native = native
	if nativeErr != nil && cancelPlugin != nil {
		cancelPlugin()
	}
	if pluginCh != nil {
		outcome := <-pluginCh
		pluginJoined = true
		execution.pluginKind = pluginObservationJoined
		execution.pluginOutcome = outcome
	} else {
		execution.pluginKind = pluginObservationNone
	}
	if cancelPlugin != nil {
		cancelPlugin()
	}
	pluginPanic.rethrow()
	return execution, joinContextError(nativeErr, ctx)
}

func joinContextError(err error, ctx context.Context) error {
	contextErr := ctx.Err()
	if contextErr == nil || errors.Is(err, contextErr) {
		return err
	}
	return errors.Join(err, contextErr)
}

// Fix validation needs only native/plugin reports that carry edits, not the
// unit's still-unmerged syntax/type groups or a complete observation result.
func (execution observationUnitExecution) fixableDiagnostics(failure PluginFailurePolicy) []rule.RuleDiagnostic {
	plugin := applyPluginFailure(execution.pluginOutcome, failure)
	var diagnostics []rule.RuleDiagnostic
	for _, group := range [][]rule.RuleDiagnostic{execution.native.diagnostics, plugin.Diagnostics} {
		for _, diagnostic := range group {
			if len(diagnostic.Fixes()) > 0 {
				diagnostics = append(diagnostics, diagnostic)
			}
		}
	}
	return diagnostics
}

// detachDiagnosticSources ends diagnostic ownership of compiler ASTs before
// releasing a unit's lease. Reuse one text projection per source object;
// distinct generations can share a path.
func (execution *observationUnitExecution) detachDiagnosticSources() {
	for _, diagnostics := range [][]rule.RuleDiagnostic{
		execution.native.syntacticDiagnostics, execution.native.diagnostics, execution.pluginOutcome.Diagnostics,
	} {
		for index := range diagnostics {
			diagnostic := &diagnostics[index]
			diagnostic.SourceFile = diagnosticTextSource(diagnostic.SourceFile)
		}
	}
	for _, diagnostics := range execution.native.typeDiagnostics {
		for index := range diagnostics {
			diagnostic := &diagnostics[index].ruleDiagnostic
			diagnostic.SourceFile = diagnosticTextSource(diagnostic.SourceFile)
		}
	}
}
