package linter

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func validateDeferredRoots(generation Generation, snapshot SourceSnapshot, policy ObservationPolicy, planChanges, stopOnSyntaxErrors bool) error {
	roots := generation.Native.DeferredRoots
	if roots == nil {
		return nil
	}
	if roots.Build == nil || generation.Native.RulesForFile == nil {
		return errors.New("linter pipeline: deferred roots require a builder and rule resolver")
	}
	if planChanges || stopOnSyntaxErrors || !snapshot.Empty() || generation.Native.TypeCheck || generation.Plugin != nil ||
		policy.Demand.LintedFiles || policy.Demand.Native != rule.EditDemandNone || policy.Demand.Plugin != rule.EditDemandNone {
		return errors.New("linter pipeline: deferred roots require ordinary native lint without retained source artifacts or edits")
	}
	return nil
}

// runDeferredRoots bounds the entire parse/bind/lint lifetime, not just AST
// traversal. A worker holds one root at a time and publishes only detached
// diagnostics. Complete Programs have already used their checker-owned shards.
func runDeferredRoots(ctx context.Context, native NativeGeneration) (NativeObservation, error) {
	files := native.DeferredRoots.FileNames
	diagnostics := make([][]rule.RuleDiagnostic, len(files))
	workerCount := min(runtime.GOMAXPROCS(0), len(files))
	if native.SingleThreaded {
		workerCount = min(1, workerCount)
	}
	type workerResult struct {
		rules        map[string]struct{}
		syntaxErrors bool
	}
	results := make([]workerResult, workerCount)
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var next atomic.Int64
	var workers sync.WaitGroup
	var errorOnce, panicOnce sync.Once
	var runErr error
	var panicValue any
	var panicked bool
	for worker := range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			completed := false
			defer func() {
				if completed {
					return
				}
				value := recover()
				if value == nil {
					value = errors.New("linter: source worker aborted without a panic value")
				}
				panicOnce.Do(func() { panicked = true; panicValue = value; cancel() })
			}()
			result := &results[worker]
			result.rules = make(map[string]struct{})
			registry := newListenerRegistry()
			for workerCtx.Err() == nil {
				index := int(next.Add(1)) - 1
				if index >= len(files) {
					break
				}
				reports, syntaxErrors, err := lintDeferredRoot(workerCtx, native, files[index], &registry, result.rules)
				if err != nil {
					errorOnce.Do(func() { runErr = err; cancel() })
					break
				}
				diagnostics[index] = reports
				result.syntaxErrors = result.syntaxErrors || syntaxErrors
			}
			completed = true
		}()
	}
	workers.Wait()
	if panicked {
		panic(panicValue)
	}
	if err := joinContextError(runErr, ctx); err != nil {
		return NativeObservation{}, err
	}
	observation := NativeObservation{Lint: &LintResult{
		LintedFileCount: int32(len(files)),
		ExecutedRules:   make(map[string]struct{}),
	}}
	for _, reports := range diagnostics {
		observation.Diagnostics = append(observation.Diagnostics, reports...)
	}
	for _, result := range results {
		observation.HasTargetSyntaxErrors = observation.HasTargetSyntaxErrors || result.syntaxErrors
		for name := range result.rules {
			observation.Lint.ExecutedRules[name] = struct{}{}
		}
	}
	return observation, nil
}

// Returning from this function ends the execution's references to this root's
// Program, plan, and AST before the worker claims another descriptor.
func lintDeferredRoot(ctx context.Context, native NativeGeneration, name string, registry *listenerRegistry, executed map[string]struct{}) ([]rule.RuleDiagnostic, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	sourceProgram, err := native.DeferredRoots.Build(ctx, name)
	if err != nil {
		return nil, false, fmt.Errorf("linter: build source %q: %w", name, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if err := validateProgram(sourceProgram); err != nil {
		return nil, false, err
	}
	files := sourceProgram.SourceFiles()
	if len(files) != 1 || files[0] == nil || files[0].FileName() != name || sourceProgram.CanProvideTypeChecker(files[0]) {
		return nil, false, fmt.Errorf("linter: deferred root %q requires its own single-source Program without a checker", name)
	}
	plan := programLintPlanFromFiles(sourceProgram, files)
	diagnostics := resolveProgramLintPlanFile(programRulePlanOptions{
		Program: sourceProgram, GetRulesForFile: native.RulesForFile,
	}, &plan, 0, ctx)
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	syntaxErrors := len(diagnostics) != 0
	filePlan := &plan.files[0]
	if !rule.CanIsolateSourceFile(filePlan.rules) {
		return nil, false, fmt.Errorf("linter: rules for deferred root %q require a complete Program", name)
	}
	if len(filePlan.rules) > 0 {
		for _, configured := range filePlan.rules {
			executed[configured.Name] = struct{}{}
		}
		lintFile(sourceProgram, programRunOptions{Cwd: native.Cwd, Timing: native.Timing}, rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		}, filePlan, filePlan.rules, nil, registry)
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	detachDiagnosticSources(diagnostics)
	return diagnostics, syntaxErrors, nil
}
