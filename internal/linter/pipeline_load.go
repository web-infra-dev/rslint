package linter

import (
	"context"
	"fmt"
	"runtime"

	"github.com/web-infra-dev/rslint/internal/rule"
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
	results := make([]observationUnitExecution, len(loads))
	errs := make([]error, len(loads))
	resources := newPluginDispatchResources()
	sourceReadLimit := runtime.GOMAXPROCS(0)
	if plan.SingleThreaded {
		sourceReadLimit = 1
	}
	sourceReads := make(pluginSourceReadSlots, sourceReadLimit)
	prepare := func(i int, reservation *programReservation) (preparedObservation, error) {
		load := loads[i]
		loads[i] = GenerationLoad{}
		// Loading owns a complete Program and its checker pool through the
		// last native/plugin borrower. Native shard progress returns CPU work,
		// but cannot admit another pool while this binding is still leased.
		reservation.retainUntilClose(load.CheckerCount)
		prepared, err := prepareObservation(ctx, load.Acquire, snapshot, policy, dispatcher,
			planChanges, stopOnTargetSyntaxErrors, reservation, sourceReads)
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
		estimates[i] = max(1, min(load.TargetCount, checkers), load.CheckerCount)
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
			planChanges, stopOnTargetSyntaxErrors)
	}
	if plan.SingleThreaded || len(loads) <= 1 {
		for i := range loads {
			run(i, nil)
		}
	} else {
		runWithProgramAdmission(estimates, runtime.GOMAXPROCS(0), run)
	}
	result, err := mergeObservationUnits(results, index)
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

func mergeObservationUnits(parts []observationUnitExecution, index int) (observationExecution, error) {
	var combined observationUnitExecution
	if len(parts) == 1 && parts[0].native.lint != nil {
		// Transfer a single binding's native/plugin slices and statistics.
		combined = parts[0]
	} else {
		combined.native.lint = &LintResult{ExecutedRules: make(map[string]struct{})}
		var plugins []EslintPluginDispatchOutcome
		for _, part := range parts {
			combined.programRootIdentities = append(combined.programRootIdentities, part.programRootIdentities...)
			native := part.native
			combined.native.syntacticDiagnostics = append(combined.native.syntacticDiagnostics, native.syntacticDiagnostics...)
			combined.native.diagnostics = append(combined.native.diagnostics, native.diagnostics...)
			combined.native.typeDiagnostics = append(combined.native.typeDiagnostics, native.typeDiagnostics...)
			combined.native.files = append(combined.native.files, native.files...)
			combined.native.hasTargetSyntaxErrors = combined.native.hasTargetSyntaxErrors || native.hasTargetSyntaxErrors
			if native.lint != nil {
				combined.native.lint.LintedFileCount += native.lint.LintedFileCount
				for name := range native.lint.ExecutedRules {
					combined.native.lint.ExecutedRules[name] = struct{}{}
				}
			}
			if part.pluginKind == pluginObservationJoined {
				combined.pluginKind = pluginObservationJoined
				plugins = append(plugins, part.pluginOutcome)
			}
			if len(part.fixTexts) != 0 && combined.fixTexts == nil {
				combined.fixTexts = make(fixTextSnapshot)
			}
			for path, text := range part.fixTexts {
				if _, duplicate := combined.fixTexts[path]; duplicate {
					return observationExecution{}, fmt.Errorf("linter pipeline: duplicate fix target %q", path)
				}
				combined.fixTexts[path] = text
			}
			if part.deferredTask != nil {
				combined.pluginKind = pluginObservationProgressive
				if combined.deferredTask == nil {
					combined.deferredTask = part.deferredTask
				} else {
					task := combined.deferredTask
					task.inputs = append(task.inputs, part.deferredTask.inputs...)
					for wire, target := range part.deferredTask.targetPathByWire {
						task.targetPathByWire[wire] = target
					}
				}
			}
		}
		combined.pluginOutcome = mergePluginOutcomes(plugins)
	}
	if combined.deferredTask != nil {
		combined.deferredTask.resources = nil
	}
	native := combined.native
	diagnostics := native.diagnostics
	if len(native.syntacticDiagnostics) > 0 {
		diagnostics = append(native.syntacticDiagnostics, diagnostics...)
	}
	// All groups retain load/Program order and their original physical keys.
	// Construct a public result only after syntax/native/type ordering and
	// cross-load type deduplication are complete.
	emitDeduplicatedTypeCheckDiagnostics(native.typeDiagnostics, func(diagnostic rule.RuleDiagnostic) {
		diagnostics = append(diagnostics, diagnostic)
	})
	return observationExecution{
		observation: ObservationResult{
			Index: index,
			Native: NativeObservation{
				Diagnostics: diagnostics, Lint: native.lint, Files: native.files,
				HasTargetSyntaxErrors: native.hasTargetSyntaxErrors,
			},
			ProgramRootIdentities: combined.programRootIdentities,
			pluginKind:            combined.pluginKind,
			pluginOutcome:         combined.pluginOutcome,
		},
		fixTexts:     combined.fixTexts,
		deferredTask: combined.deferredTask,
	}, nil
}
