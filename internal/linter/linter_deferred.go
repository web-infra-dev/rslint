package linter

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var errDeferredWorkerAborted = errors.New("linter: deferred file worker aborted without a panic value")

func runDeferredFiles(ctx context.Context, files []deferredFilePlan, opts programRunOptions, consumer rule.DiagnosticConsumer) (programLintResult, error) {
	if len(files) == 0 {
		return programLintResult{}, ctx.Err()
	}
	if consumer.Demand != rule.EditDemandNone {
		return programLintResult{}, errors.New("linter: deferred files cannot retain edit artifacts")
	}
	workerCount := min(runtime.GOMAXPROCS(0), len(files))
	if opts.SingleThreaded {
		workerCount = 1
	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make([]programLintResult, workerCount)
	errorsByFile := make([]error, len(files))
	var next atomic.Int64
	var workers sync.WaitGroup
	var panicOnce sync.Once
	var panicValue any
	for worker := range workerCount {
		workers.Go(func() {
			completed := false
			defer func() {
				if !completed {
					value := recover()
					if value == nil {
						value = errDeferredWorkerAborted
					}
					panicOnce.Do(func() { panicValue = value; cancel() })
				}
			}()
			listeners := newListenerRegistry()
			defer listeners.reset()
			result := &results[worker]
			for workerCtx.Err() == nil {
				index := int(next.Add(1)) - 1
				if index >= len(files) {
					break
				}
				rules, syntax, err := runDeferredFile(workerCtx, files[index], opts, consumer, &listeners)
				if err != nil {
					errorsByFile[index] = err
					continue
				}
				result.lintedFileCount++
				result.hasSyntacticDiagnostics = result.hasSyntacticDiagnostics || syntax
				if len(rules) > 0 && result.executedRules == nil {
					result.executedRules = make(map[string]struct{})
				}
				for _, configured := range rules {
					result.executedRules[configured.Name] = struct{}{}
				}
			}
			completed = true
		})
	}
	workers.Wait()
	// Never unwind the generation or its diagnostic consumer while a worker
	// can still read or report. Propagate panics on the caller's goroutine.
	if panicValue != nil {
		panic(panicValue)
	}
	if err := ctx.Err(); err != nil {
		return programLintResult{}, err
	}
	for _, err := range errorsByFile {
		if err != nil {
			return programLintResult{}, err
		}
	}
	result := programLintResult{executedRules: make(map[string]struct{})}
	for _, worker := range results {
		result.lintedFileCount += worker.lintedFileCount
		result.hasSyntacticDiagnostics = result.hasSyntacticDiagnostics || worker.hasSyntacticDiagnostics
		for name := range worker.executedRules {
			result.executedRules[name] = struct{}{}
		}
	}
	return result, nil
}

// This scope owns the only AST references for one deferred input. Diagnostics
// leave it as text projections; the shared plan and worker registry retain none.
func runDeferredFile(
	ctx context.Context,
	input deferredFilePlan,
	opts programRunOptions,
	consumer rule.DiagnosticConsumer,
	listeners *listenerRegistry,
) ([]rule.ConfiguredRule, bool, error) {
	sourceProgram, err := input.sources.BuildFile(input.index)
	if err != nil {
		return nil, false, err
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	files, err := resolveExactProgramFiles(sourceProgram, []string{input.fileName()})
	if err != nil {
		return nil, false, err
	}
	plan := programLintPlanFromFiles(sourceProgram, files)
	diagnostics := resolveProgramLintPlanFile(programRulePlanOptions{
		Program: sourceProgram,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return input.rules
		},
	}, &plan, 0, ctx)
	syntax := len(diagnostics) > 0
	file := &plan.files[0]
	if len(file.rules) > 0 {
		runLintRulesInFile(sourceProgram, file, file.rules, nil, opts, rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		}, listeners)
	}
	detachDiagnosticSources(diagnostics)
	for _, diagnostic := range diagnostics {
		consumer.Report(diagnostic)
	}
	return file.rules, syntax, ctx.Err()
}
