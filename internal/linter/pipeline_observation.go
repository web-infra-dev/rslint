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

func executeObservation(
	ctx context.Context,
	provider GenerationProvider,
	snapshot SourceSnapshot,
	policy ObservationPolicy,
	dispatcher EslintPluginDispatcher,
	index int,
	planChanges bool,
	stopOnTargetSyntaxErrors bool,
) (observationExecution, error) {
	generation, release, err := provider.AcquireGeneration(ctx, snapshot)
	if err != nil {
		return observationExecution{}, err
	}
	lease := &releaseLease{release: release}
	defer lease.close()
	if err := ctx.Err(); err != nil {
		return observationExecution{}, err
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
			return observationExecution{}, fmt.Errorf("linter pipeline: prepare lint plan: %w", err)
		}
	}
	lintedFiles, err := projectGenerationTargets(
		ctx,
		generation,
		plan,
		snapshot,
		policy.Demand.LintedFiles,
	)
	if err != nil {
		return observationExecution{}, err
	}
	detachedPlugin := policy.Plugin != PluginConcurrentJoined
	targetSyntaxBlocked := stopOnTargetSyntaxErrors && plan != nil && plan.HasSyntacticDiagnostics()
	pluginWork := pluginTask{failure: policy.PluginFailure}
	if !targetSyntaxBlocked {
		pluginWork, err = materializePluginTask(plan, generation, snapshot, policy, detachedPlugin)
		if err != nil {
			return observationExecution{}, err
		}
	}
	if len(pluginWork.inputs) > 0 && policy.Plugin != pluginProgressiveAfterNative && dispatcher == nil {
		return observationExecution{}, errors.New("linter pipeline: joined plugin work requires a dispatcher")
	}
	execution := observationExecution{
		observation: ObservationResult{Index: index},
	}
	switch policy.Plugin {
	case PluginConcurrentJoined:
		execution, runErr := executeConcurrentObservation(
			ctx,
			generation,
			plan,
			pluginWork,
			dispatcher,
			execution,
			policy.Demand.Native,
			lintedFiles,
		)
		if runErr == nil && planChanges && !targetSyntaxBlocked {
			diagnostics, _ := execution.observation.CompleteDiagnostics()
			execution.fixTexts, runErr = freezeFixTextsForDiagnostics(
				ctx,
				generation,
				snapshot,
				diagnostics,
			)
		}
		lease.close()
		return execution, joinContextError(runErr, ctx)
	case PluginAfterNativeJoined:
		native, nativeErr := runNativeObservation(ctx, generation, plan, policy.Demand.Native, lintedFiles)
		execution.observation.Native = native
		if nativeErr != nil {
			lease.close()
			return execution, nativeErr
		}
		if stopOnTargetSyntaxErrors && native.HasTargetSyntaxErrors {
			pluginWork.fixCandidates = nil
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
			execution.fixTexts, err = freezeFixTexts(ctx, generation, snapshot, candidateSources)
			if err != nil {
				return execution, err
			}
		}
		pluginWork.fixCandidates = nil
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
		outcome := pluginWork.run(ctx, dispatcher)
		execution.observation.pluginKind = pluginObservationJoined
		execution.observation.pluginOutcome = outcome
		execution.pluginOutcome = &execution.observation.pluginOutcome
		if err := ctx.Err(); err != nil {
			return execution, err
		}
		if planChanges {
			diagnostics, _ := execution.observation.CompleteDiagnostics()
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
		native, nativeErr := runNativeObservation(ctx, generation, plan, policy.Demand.Native, lintedFiles)
		execution.observation.Native = native
		// Clear the last SourceFile-bearing side channel before releasing the
		// generation. The detached input itself was already deep-frozen above.
		pluginWork.fixCandidates = nil
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
	generation Generation,
	plan *LintPlan,
	pluginTask pluginTask,
	dispatcher EslintPluginDispatcher,
	execution observationExecution,
	nativeDemand rule.EditDemand,
	lintedFiles []LintedFile,
) (observationExecution, error) {
	var (
		pluginCh     <-chan EslintPluginDispatchOutcome
		cancelPlugin context.CancelFunc
		pluginJoined bool
	)
	if len(pluginTask.inputs) > 0 {
		pluginCtx, cancel := context.WithCancel(ctx)
		cancelPlugin = cancel
		ch := make(chan EslintPluginDispatchOutcome, 1)
		pluginCh = ch
		go func() {
			ch <- pluginTask.run(pluginCtx, dispatcher)
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

	native, nativeErr := runNativeObservation(ctx, generation, plan, nativeDemand, lintedFiles)
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
	return execution, joinContextError(nativeErr, ctx)
}

func joinContextError(err error, ctx context.Context) error {
	contextErr := ctx.Err()
	if contextErr == nil || errors.Is(err, contextErr) {
		return err
	}
	return errors.Join(err, contextErr)
}
