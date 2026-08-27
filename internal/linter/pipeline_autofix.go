package linter

import (
	"context"
	"errors"
)

// runAutofixPipeline owns the complete bounded apply-and-observe state machine.
// Every round mutates autofixState only; an optional external commit happens
// once, after the final successful in-memory state has been decided.
func runAutofixPipeline(
	ctx context.Context,
	request PipelineRequest,
	result PipelineResult,
	initial observationExecution,
) (PipelineResult, error) {
	state := newAutofixState()
	initialObservation := initial.observation
	result.fix = fixResult{
		kind:     fixResultApplied,
		initial:  initialObservation,
		verified: true,
	}
	current := initial
	// Only current needs the initial generation's frozen fix text and plugin
	// task. Keep the observation needed by the public result separately so those
	// potentially large generation artifacts become unreachable after the first
	// re-observation.
	initial.fixTexts = nil
	initial.pluginOutcome = nil
	initial.deferredTask = nil
	if request.autofix.StopOnTargetSyntaxErrors && current.observation.Native.HasTargetSyntaxErrors {
		return commitFinalChanges(ctx, request, result, state)
	}

	for roundIndex := range request.autofix.MaxRounds {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		diagnostics, complete := current.observation.CompleteDiagnostics()
		if !complete {
			return result, errors.New("linter pipeline: autofix observation is incomplete")
		}
		changes, err := planFixes(diagnostics, current.fixTexts)
		if err = joinContextError(err, ctx); err != nil {
			return result, err
		}
		if len(changes) == 0 {
			return commitFinalChanges(ctx, request, result, state)
		}
		if err := ctx.Err(); err != nil {
			return result, err
		}

		round, applyErr := state.apply(changes)
		if applyErr != nil {
			return result, joinContextError(applyErr, ctx)
		}
		result.fix.verified = false
		result.fix.rounds = append(result.fix.rounds, round)
		result.fix.finalChanges = state.finalChanges()
		if round.RestoredInitial {
			result.Observation = initialObservation
			result.fix.verified = true
			if err := ctx.Err(); err != nil {
				return result, err
			}
			return commitFinalChanges(ctx, request, result, state)
		}
		if err := ctx.Err(); err != nil {
			return result, err
		}

		lastAllowedRound := roundIndex+1 == request.autofix.MaxRounds
		if lastAllowedRound && !request.autofix.VerifyAfterLastRound {
			return commitFinalChanges(ctx, request, result, state)
		}
		policy := request.policy
		planNextChanges := true
		if lastAllowedRound {
			policy.Demand = request.autofix.VerificationDemand
			planNextChanges = false
		}
		next, observeErr := executeObservation(
			ctx,
			request.provider,
			state.snapshot(),
			policy,
			request.dispatcher,
			current.observation.Index+1,
			planNextChanges,
			request.autofix.StopOnTargetSyntaxErrors,
		)
		recordPluginDispatch(&result, next)
		if observeErr != nil {
			return result, observeErr
		}
		result.Observation = next.observation
		mergeSuccessfulExecution(&result, next)
		current = next
		result.fix.verified = true
		if request.autofix.StopOnTargetSyntaxErrors && current.observation.Native.HasTargetSyntaxErrors {
			return commitFinalChanges(ctx, request, result, state)
		}
		if lastAllowedRound {
			return commitFinalChanges(ctx, request, result, state)
		}
	}
	return commitFinalChanges(ctx, request, result, state)
}

func commitFinalChanges(
	ctx context.Context,
	request PipelineRequest,
	result PipelineResult,
	state *autofixState,
) (PipelineResult, error) {
	result.fix.finalChanges = state.finalChanges()
	if request.commit == nil {
		return result, ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if len(result.fix.finalChanges) == 0 {
		result.fix.committed = true
		return result, nil
	}
	commitResult, commitErr := request.commit.committer.CommitFinalChanges(
		ctx,
		cloneFileChanges(result.fix.finalChanges),
	)
	confirmed, validationErr := validateCommitResult(result.fix.finalChanges, commitResult, commitErr)
	result.fix.committedPaths = confirmed
	result.fix.committed = validationErr == nil
	return result, joinContextError(validationErr, ctx)
}
