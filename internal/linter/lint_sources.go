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

func sourceWorkerCount(count int, singleThreaded bool) int {
	if singleThreaded {
		return min(1, count)
	}
	return min(runtime.GOMAXPROCS(0), count)
}

// runSourceTasks joins every admitted file before returning an error or
// propagating a panic. A worker claims its next file only after the previous
// callback has returned, so parsing cannot run ahead of lint completion.
func runSourceTasks(ctx context.Context, count, workerCount int, run func(context.Context, int, int) error) error {
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
			for workerCtx.Err() == nil {
				index := int(next.Add(1)) - 1
				if index >= count {
					break
				}
				if err := run(workerCtx, index, worker); err != nil {
					errorOnce.Do(func() { runErr = err; cancel() })
					break
				}
			}
			completed = true
		}()
	}
	workers.Wait()
	if panicked {
		panic(panicValue)
	}
	return joinContextError(runErr, ctx)
}

func runSourceFiles(ctx context.Context, files []sourceFilePlan, opts programRunOptions, consumer rule.DiagnosticConsumer) (programLintResult, bool, error) {
	type workerResult struct {
		registry     listenerRegistry
		rules        map[string]struct{}
		syntaxErrors bool
	}
	workers := make([]workerResult, sourceWorkerCount(len(files), opts.SingleThreaded))
	for index := range workers {
		workers[index].registry = newListenerRegistry()
		workers[index].rules = make(map[string]struct{})
	}
	diagnostics := make([][]rule.RuleDiagnostic, len(files))
	err := runSourceTasks(ctx, len(files), len(workers), func(ctx context.Context, index, worker int) error {
		result := &workers[worker]
		reports, syntaxErrors, err := lintSourceFile(ctx, files[index], opts, consumer.Demand, &result.registry, result.rules)
		diagnostics[index] = reports
		result.syntaxErrors = result.syntaxErrors || syntaxErrors
		return err
	})
	if err != nil {
		return programLintResult{}, false, err
	}
	result := programLintResult{lintedFileCount: int32(len(files)), executedRules: make(map[string]struct{})}
	var syntaxErrors bool
	for _, worker := range workers {
		syntaxErrors = syntaxErrors || worker.syntaxErrors
		for name := range worker.rules {
			result.executedRules[name] = struct{}{}
		}
	}
	for _, reports := range diagnostics {
		for _, diagnostic := range reports {
			consumer.Report(diagnostic)
		}
	}
	return result, syntaxErrors, nil
}

// Returning ends all execution references to this file's Program and AST.
// Only diagnostic text and edit values can survive into the next file.
func lintSourceFile(ctx context.Context, source sourceFilePlan, opts programRunOptions, demand rule.EditDemand, registry *listenerRegistry, executed map[string]struct{}) ([]rule.RuleDiagnostic, bool, error) {
	p, err := buildRootProgram(ctx, source.build, []string{source.name})
	if err != nil {
		return nil, false, err
	}
	file := p.SourceFiles()[0]
	diagnostics := CollectFileSyntacticDiagnostics(ctx, p, file)
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	syntaxErrors := len(diagnostics) > 0
	rules := rule.FilterNonTypeAwareRules(source.rules)
	if !syntaxErrors && len(rules) > 0 {
		filePlan := lintFilePlan{file: file, rules: rules, environment: firstNativeRuleEnvironment(rules)}
		for _, configured := range rules {
			executed[configured.Name] = struct{}{}
		}
		lintFile(p, opts, rule.DiagnosticConsumer{
			Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		}, &filePlan, rules, nil, registry)
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	detachSourceDiagnostics(diagnostics)
	return diagnostics, syntaxErrors, nil
}

func detachSourceDiagnostics(diagnostics []rule.RuleDiagnostic) {
	var sources map[*ast.SourceFile]*textSourceFile
	for index := range diagnostics {
		diagnostic := &diagnostics[index]
		source, ok := diagnostic.SourceFile.(*ast.SourceFile)
		if !ok || source == nil {
			continue
		}
		projection := sources[source]
		if projection == nil {
			if sources == nil {
				sources = make(map[*ast.SourceFile]*textSourceFile)
			}
			projection = newTextSourceFile(source.Text())
			sources[source] = projection
		}
		diagnostic.SourceFile = projection
	}
}
