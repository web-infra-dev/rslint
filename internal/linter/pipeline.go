package linter

import (
	"context"
	"errors"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// RunPipeline is the single production lint orchestrator. It owns generation
// acquisition/release, preparation, native/plugin scheduling, path projection,
// fix planning, bounded in-memory rounds, re-observation, optional terminal
// commit, and rule aggregation.
// Integrations supply generation and transport behavior only through the sealed
// request's ports.
func RunPipeline(ctx context.Context, request PipelineRequest) (PipelineResult, error) {
	result := PipelineResult{executedRules: make(map[string]struct{})}
	if ctx == nil {
		return result, errors.New("linter pipeline: context must not be nil")
	}
	if err := request.validate(); err != nil {
		return result, err
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}

	planChanges := request.kind == pipelineRequestAutofix
	initial, err := executeObservation(
		ctx,
		request.provider,
		SourceSnapshot{},
		request.policy,
		request.dispatcher,
		0,
		planChanges,
		request.kind == pipelineRequestProgressiveLint ||
			(request.kind == pipelineRequestAutofix && request.autofix.StopOnTargetSyntaxErrors),
	)
	recordPluginDispatch(&result, initial)
	if err != nil {
		return result, err
	}
	result.Observation = initial.observation
	mergeSuccessfulExecution(&result, initial)

	switch request.kind {
	case pipelineRequestLint:
		result.fix = fixResult{kind: fixResultNone}
		return result, nil
	case pipelineRequestProgressiveLint:
		return presentProgressiveLint(ctx, request, result, initial)
	case pipelineRequestAutofix:
		return runAutofixPipeline(ctx, request, result, initial)
	default:
		return result, errors.New("linter pipeline: request kind is invalid")
	}
}

func presentProgressiveLint(
	ctx context.Context,
	request PipelineRequest,
	result PipelineResult,
	execution observationExecution,
) (PipelineResult, error) {
	result.fix = fixResult{kind: fixResultNone}
	var enrichment DeferredPluginRun
	if execution.deferredTask != nil {
		var err error
		enrichment, err = newDeferredPluginRun(*execution.deferredTask)
		if err != nil {
			return result, err
		}
	}
	baseline := append([]rule.RuleDiagnostic(nil), execution.observation.Native.Diagnostics...)
	request.presenter.PublishBaseline(ctx, baseline)
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if enrichment != nil {
		request.presenter.Submit(ctx, enrichment)
	}
	return result, nil
}

func mergeSuccessfulExecution(result *PipelineResult, execution observationExecution) {
	if result == nil {
		return
	}
	if lintResult := execution.observation.Native.Lint; lintResult != nil {
		for name := range lintResult.ExecutedRules {
			result.executedRules[name] = struct{}{}
		}
	}
}

func recordPluginDispatch(result *PipelineResult, execution observationExecution) {
	if result != nil && execution.pluginOutcome != nil {
		result.pluginOutcomes = append(result.pluginOutcomes, PluginDispatchRecord{
			Observation:   execution.observation.Index,
			Notices:       append([]EslintPluginProtocolNotice(nil), execution.pluginOutcome.Notices...),
			DispatchError: execution.pluginOutcome.DispatchError,
		})
	}
}
