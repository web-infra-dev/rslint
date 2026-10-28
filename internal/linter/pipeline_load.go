package linter

import (
	"context"
	"fmt"
	"runtime"
)

// executeObservation owns admission, lifetime and the result boundary for all
// integrations and rule backends. A provider publishes complete disjoint
// bindings; loading failures never trigger a second execution of rules.
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
	if err := ctx.Err(); err != nil {
		return observationExecution{}, err
	}
	plan, err := provider.PlanGeneration(ctx, snapshot)
	if err != nil {
		return observationExecution{}, &GenerationLoadError{Err: err}
	}
	loads := plan.Loads
	results := make([]observationExecution, len(loads))
	errs := make([]error, len(loads))
	resources := newPluginDispatchResources()
	prepare := func(i int, reservation *programReservation) (preparedObservation, error) {
		load := loads[i]
		loads[i] = GenerationLoad{}
		prepared, err := prepareObservation(ctx, load.Acquire, snapshot, policy, dispatcher,
			planChanges, stopOnTargetSyntaxErrors, reservation)
		if err == nil {
			prepared.plugin.resources = resources
			if len(prepared.plugin.inputs) != 0 {
				// Projection is another AST consumer. Native shard completion
				// alone must not return all of this binding's admission capacity.
				reservation.retainUntilClose(1)
			}
		}
		return prepared, err
	}
	// Syntax-gated operations decide whether any plugin work may run only
	// after every target is prepared. Ordinary observations stream units.
	var prepared []preparedObservation
	if stopOnTargetSyntaxErrors {
		prepared = make([]preparedObservation, len(loads))
		defer func() {
			for _, work := range prepared {
				if work.lease != nil {
					work.native.close()
					work.lease.close()
				}
			}
		}()
		blocked := false
		for i := range loads {
			prepared[i], err = prepare(i, nil)
			if err != nil {
				return observationExecution{}, err
			}
			blocked = blocked || prepared[i].targetSyntaxBlocked
		}
		if blocked {
			for i := range prepared {
				prepared[i].targetSyntaxBlocked = true
				prepared[i].plugin = pluginTask{}
			}
		}
	}
	estimates := make([]int, len(loads))
	for i, load := range loads {
		checkers := load.CheckerCount
		if checkers <= 0 {
			checkers = 4 // The compiler default, also used by programCheckerHint.
		}
		estimates[i] = max(1, min(load.TargetCount, checkers))
	}
	run := func(i int, reservation *programReservation) {
		var work preparedObservation
		if prepared != nil {
			work = prepared[i]
			prepared[i] = preparedObservation{}
		} else {
			work, errs[i] = prepare(i, reservation)
			if errs[i] != nil {
				return
			}
		}
		results[i], errs[i] = executePreparedObservation(ctx, work, snapshot, policy, dispatcher,
			index, planChanges, stopOnTargetSyntaxErrors)
	}
	if plan.SingleThreaded || len(loads) <= 1 {
		for i := range loads {
			run(i, nil)
		}
	} else {
		runWithProgramAdmission(estimates, runtime.GOMAXPROCS(0), run)
	}
	result, err := mergeObservationExecutions(results, index)
	if err != nil {
		return observationExecution{}, err
	}
	if result.observation.pluginKind == pluginObservationJoined {
		result.observation.pluginOutcome = applyPluginFailure(result.observation.pluginOutcome, policy.PluginFailure)
		result.pluginOutcome = &result.observation.pluginOutcome
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	for _, err := range errs {
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func mergeObservationExecutions(parts []observationExecution, index int) (observationExecution, error) {
	// A complete single binding already has the final diagnostic ordering.
	// Transfer its result without copying native and plugin output slices.
	if len(parts) == 1 && parts[0].observation.Native.Lint != nil {
		result := parts[0]
		result.observation.Index = index
		if result.deferredTask != nil {
			result.deferredTask.resources = nil
		}
		return result, nil
	}
	result := observationExecution{observation: ObservationResult{
		Index: index, Native: NativeObservation{Lint: &LintResult{ExecutedRules: make(map[string]struct{})}},
	}}
	var plugins []EslintPluginDispatchOutcome
	for _, part := range parts {
		native := part.observation.Native
		result.observation.Native.Diagnostics = append(result.observation.Native.Diagnostics, native.Diagnostics...)
		result.observation.Native.Files = append(result.observation.Native.Files, native.Files...)
		result.observation.Native.HasTargetSyntaxErrors = result.observation.Native.HasTargetSyntaxErrors || native.HasTargetSyntaxErrors
		if native.Lint != nil {
			result.observation.Native.Lint.LintedFileCount += native.Lint.LintedFileCount
			for name := range native.Lint.ExecutedRules {
				result.observation.Native.Lint.ExecutedRules[name] = struct{}{}
			}
		}
		if part.observation.pluginKind == pluginObservationJoined {
			result.observation.pluginKind = pluginObservationJoined
			plugins = append(plugins, part.observation.pluginOutcome)
		}
		if len(part.fixTexts) != 0 && result.fixTexts == nil {
			result.fixTexts = make(fixTextSnapshot)
		}
		for path, text := range part.fixTexts {
			if _, duplicate := result.fixTexts[path]; duplicate {
				return observationExecution{}, fmt.Errorf("linter pipeline: duplicate fix target %q", path)
			}
			result.fixTexts[path] = text
		}
		if part.deferredTask != nil {
			result.observation.pluginKind = pluginObservationProgressive
			if result.deferredTask == nil {
				result.deferredTask = part.deferredTask
				result.deferredTask.resources = nil
			} else {
				task := result.deferredTask
				task.inputs = append(task.inputs, part.deferredTask.inputs...)
				for wire, target := range part.deferredTask.targetPathByWire {
					task.targetPathByWire[wire] = target
				}
			}
		}
	}
	result.observation.pluginOutcome = mergePluginOutcomes(plugins)
	return result, nil
}
