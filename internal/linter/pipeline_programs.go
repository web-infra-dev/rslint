package linter

import (
	"context"
	"runtime"
)

// preparedProgramObservation owns only one Program's consumable native/plugin
// work. The complete binding's lease and requested artifacts stay with its
// parent observation; tasks never finalize a producer borrow independently.
type preparedProgramObservation struct {
	native *nativeObservationWork
	plugin pluginTask
}

func (p *preparedObservation) partitionPrograms(mode PluginExecution, planChanges, stopOnTargetSyntaxErrors bool) {
	execution := p.native.execution
	// Fix validation, syntax gating, type-diagnostic ordering/deduplication and
	// detached plugin release all consume the complete binding coherently.
	if planChanges || stopOnTargetSyntaxErrors || execution.typeCheck || mode != PluginConcurrentJoined {
		return
	}
	// In a serial native run, a blocked joined plugin must not prevent later
	// native Programs from running. Preserve that existing consumer ordering.
	if len(execution.programs) <= 1 || (execution.options.SingleThreaded && len(p.plugin.inputs) != 0) {
		return
	}
	p.programs = make([]preparedProgramObservation, len(execution.programs))
	inputOffset := 0
	for index, programPlan := range execution.programs {
		nativeExecution := *execution
		nativeExecution.programs = []*programLintPlan{programPlan}
		nativeExecution.options.reservation = nil
		native := *p.native
		native.execution = &nativeExecution
		// Syntax diagnostics have already been deduplicated over the complete
		// binding. Keep them and explicitly requested ASTs at the parent boundary.
		native.diagnostics = nil
		native.files = nil
		native.hasTargetSyntaxErrors = false
		plugin := p.plugin
		plugin.inputs = nil
		if len(p.plugin.inputs) != 0 {
			end := inputOffset
			for end < len(p.plugin.inputs) && p.plugin.inputs[end].programIndex == index {
				end++
			}
			// Own each container: a subslice of the complete input array would
			// keep completed Programs' ASTs alive through unrelated pending tasks.
			plugin.inputs = append([]EslintPluginFileInput(nil), p.plugin.inputs[inputOffset:end]...)
			inputOffset = end
		}
		p.programs[index] = preparedProgramObservation{native: &native, plugin: plugin}
	}
	execution.programs = nil
	p.plugin = pluginTask{}
}

func executePreparedPrograms(
	ctx context.Context,
	prepared *preparedObservation,
	dispatcher EslintPluginDispatcher,
	index int,
) (observationExecution, error) {
	parent := prepared.native
	budget := runtime.GOMAXPROCS(0)
	results := make([]observationExecution, len(prepared.programs))
	errs := make([]error, len(prepared.programs))
	estimates := make([]int, len(prepared.programs))
	for i, task := range prepared.programs {
		estimates[i] = estimateProgramLintShards(task.native.execution.programs[0], budget)
	}
	run := func(i int, reservation *programReservation) {
		task := prepared.programs[i]
		prepared.programs[i] = preparedProgramObservation{}
		defer task.native.close()
		task.native.execution.options.reservation = reservation
		if len(task.plugin.inputs) != 0 {
			reservation.retainUntilClose(1)
		}
		results[i], errs[i] = executeConcurrentObservation(ctx, task.native, task.plugin, dispatcher,
			observationExecution{observation: ObservationResult{Index: index}})
		results[i].observation.detachDiagnosticSources()
	}
	if parent.execution.options.SingleThreaded {
		// Preserve the native serial work group's reverse queue order.
		for i := len(prepared.programs) - 1; i >= 0; i-- {
			run(i, nil)
		}
	} else {
		runWithProgramAdmission(estimates, budget, run)
	}
	result, err := mergeObservationExecutions(results, index)
	if err != nil {
		return observationExecution{}, err
	}
	parent.projectDiagnosticPaths(parent.diagnostics)
	result.observation.Native.Diagnostics = append(parent.diagnostics, result.observation.Native.Diagnostics...)
	parent.diagnostics = nil
	result.observation.Native.Files = parent.files
	parent.files = nil
	result.observation.Native.HasTargetSyntaxErrors = parent.hasTargetSyntaxErrors
	for _, err := range errs {
		if err != nil {
			return result, err
		}
	}
	return result, ctx.Err()
}
