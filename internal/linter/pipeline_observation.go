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

// preparedObservation is the ownership handoff between whole-generation
// validation and execution. Only actual fix planning retains a source reader;
// the executor retains neither the producer Generation nor its LintPlan.
type preparedObservation struct {
	native              *nativeObservationWork
	plugin              pluginTask
	lease               *releaseLease
	readFixText         sourceTextReader
	targetSyntaxBlocked bool
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
		pluginWork, err = materializePluginTask(ctx, plan, generation, snapshot, policy, detachedPlugin, planChanges)
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
		native:              nativeWork,
		plugin:              pluginWork,
		lease:               lease,
		targetSyntaxBlocked: targetSyntaxBlocked,
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
	index int,
	planChanges bool,
	stopOnTargetSyntaxErrors bool,
) (observationExecution, error) {
	var err error
	lease := prepared.lease
	defer lease.close()
	nativeWork := prepared.native
	defer nativeWork.close()
	pluginWork := prepared.plugin
	prepared.plugin = pluginTask{}
	targetSyntaxBlocked := prepared.targetSyntaxBlocked
	execution := observationExecution{
		observation: ObservationResult{Index: index},
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
			diagnostics := execution.fixDiagnostics(policy.PluginFailure)
			execution.fixTexts, runErr = freezeFixTextsForDiagnostics(
				ctx,
				prepared.readFixText,
				snapshot,
				diagnostics,
			)
		}
		execution.observation.detachDiagnosticSources()
		lease.close()
		return execution, joinContextError(runErr, ctx)
	case PluginAfterNativeJoined:
		native, nativeErr := runNativeObservation(ctx, nativeWork)
		execution.observation.Native = native
		if nativeErr != nil {
			execution.observation.detachDiagnosticSources()
			lease.close()
			return execution, nativeErr
		}
		if stopOnTargetSyntaxErrors && native.HasTargetSyntaxErrors {
			pluginWork.fixCandidates = nil
			execution.observation.detachDiagnosticSources()
			lease.close()
			execution.observation.pluginKind = pluginObservationNone
			return execution, ctx.Err()
		}
		if planChanges {
			candidateSources, sourceErr := fixSourcesFromDiagnostics(native.Diagnostics)
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
		execution.observation.detachDiagnosticSources()
		// Detached plugin inputs and frozen fix text no longer reference generation
		// state, so watcher/Program resources are released before a reverse request
		// can block.
		lease.close()
		if err := ctx.Err(); err != nil {
			return execution, err
		}
		if len(pluginWork.inputs) == 0 {
			execution.observation.pluginKind = pluginObservationNone
			return execution, nil
		}
		outcome := pluginWork.runRaw(ctx, dispatcher)
		execution.observation.pluginKind = pluginObservationJoined
		execution.observation.pluginOutcome = outcome
		execution.pluginOutcome = &execution.observation.pluginOutcome
		if err := ctx.Err(); err != nil {
			return execution, err
		}
		if planChanges {
			diagnostics := execution.fixDiagnostics(policy.PluginFailure)
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
		execution.observation.Native = native
		execution.observation.detachDiagnosticSources()
		lease.close()
		if nativeErr != nil {
			return execution, nativeErr
		}
		if err := ctx.Err(); err != nil {
			return execution, err
		}
		if stopOnTargetSyntaxErrors && native.HasTargetSyntaxErrors {
			execution.observation.pluginKind = pluginObservationNone
			return execution, nil
		}
		if len(pluginWork.inputs) == 0 {
			execution.observation.pluginKind = pluginObservationNone
			return execution, nil
		}
		execution.observation.pluginKind = pluginObservationProgressive
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
	execution observationExecution,
) (observationExecution, error) {
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
	execution.observation.Native = native
	if nativeErr != nil && cancelPlugin != nil {
		cancelPlugin()
	}
	if pluginCh != nil {
		outcome := <-pluginCh
		pluginJoined = true
		execution.observation.pluginKind = pluginObservationJoined
		execution.observation.pluginOutcome = outcome
		execution.pluginOutcome = &execution.observation.pluginOutcome
	} else {
		execution.observation.pluginKind = pluginObservationNone
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

// Fix validation uses the observation's failure policy without publishing it.
// Providers keep fix generations coherent because validation spans all targets.
func (execution observationExecution) fixDiagnostics(failure PluginFailurePolicy) []rule.RuleDiagnostic {
	observation := execution.observation
	observation.pluginOutcome = applyPluginFailure(observation.pluginOutcome, failure)
	diagnostics, _ := observation.CompleteDiagnostics()
	return diagnostics
}
