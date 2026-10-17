package linter

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// lintExecution owns private, consumable task slots. The immutable plan and
// producer slices are borrowed only during preparation, never during execution.
// Closing a task drops our references without modifying its Program or ASTs.
type lintExecution struct {
	programs         []*programLintPlan
	typeCheckSources []*program.DiagnosticSource
	options          programRunOptions
}

func prepareLintExecution(opts RunLinterOptions) (*lintExecution, error) {
	if !opts.Consumer.Demand.IsValid() {
		return nil, errors.New("linter: invalid native edit demand")
	}
	if opts.LintPlan != nil && opts.TypeCheckOnlyPrograms != nil {
		return nil, errTypeCheckOnlyProgramsWithPlan
	}
	execution := &lintExecution{options: programRunOptions{
		Cwd:                  opts.Cwd,
		CollectExecutedRules: true,
		SingleThreaded:       opts.SingleThreaded,
		Timing:               opts.Timing,
	}}
	if opts.LintPlan != nil {
		if err := validateProgramsInPlan(opts.LintPlan); err != nil {
			return nil, err
		}
		execution.programs = make([]*programLintPlan, len(opts.LintPlan.programs))
		for index, plan := range opts.LintPlan.programs {
			// Only the task container is private. Its files and rules remain
			// immutable shared content, including when a caller reuses the plan.
			task := plan
			execution.programs[index] = &task
			if opts.TypeCheck {
				if source := plan.program.DiagnosticSource(); source != nil {
					execution.typeCheckSources = append(execution.typeCheckSources, source)
				}
			}
		}
	} else if opts.TypeCheck {
		if err := validatePrograms(opts.TypeCheckOnlyPrograms); err != nil {
			return nil, err
		}
		for _, sourceProgram := range opts.TypeCheckOnlyPrograms {
			if source := sourceProgram.DiagnosticSource(); source != nil {
				execution.typeCheckSources = append(execution.typeCheckSources, source)
			}
		}
	}
	return execution, nil
}

func validateProgramsInPlan(plan *LintPlan) error {
	for _, task := range plan.programs {
		if err := validateProgram(task.program); err != nil {
			return err
		}
	}
	return nil
}

func (e *lintExecution) close() {
	clear(e.programs)
	e.programs = nil
	clear(e.typeCheckSources)
	e.typeCheckSources = nil
}

func (e *lintExecution) run(consumer rule.DiagnosticConsumer) *LintResult {
	defer e.close()
	result := &LintResult{ExecutedRules: make(map[string]struct{})}
	programResults := make([]programLintResult, len(e.programs))
	if len(e.programs) > 0 {
		work := newProgramWorkGroup(e.options.SingleThreaded, len(e.programs))
		for index := range e.programs {
			task := e.programs[index]
			e.programs[index] = nil
			resultSlot := &programResults[index]
			options := e.options
			work.Queue(func() {
				defer func() { *task = programLintPlan{} }()
				*resultSlot = runLintRulesInProgram(task, options, consumer)
			})
		}
		e.programs = nil
		work.RunAndWait()
	}
	for _, programResult := range programResults {
		result.LintedFileCount += programResult.lintedFileCount
		for name := range programResult.executedRules {
			result.ExecutedRules[name] = struct{}{}
		}
	}

	// Keep the native/type-check phase barrier. Diagnostic sources borrow the
	// backend without retaining the lint facade or its owned derived caches.
	sources := e.typeCheckSources
	e.typeCheckSources = nil
	if len(sources) > 0 {
		runTypeCheckAcrossPrograms(typeCheckRequest{
			Sources:        sources,
			SingleThreaded: e.options.SingleThreaded,
			OnDiagnostic:   consumer.Report,
		})
	}
	return result
}

// lintWorkGroup joins all scheduled borrowers before re-raising an abnormal
// worker exit. Nested checker shards own their independent join barrier.
type lintWorkGroup struct {
	admission chan struct{}
	work      core.WorkGroup
	panics    workerPanic
}

// workerPanic preserves the original panic until all sibling borrowers join.
type workerPanic struct {
	panicOnce  sync.Once
	aborted    atomic.Bool
	panicValue any
}

func newLintWorkGroup(singleThreaded bool) *lintWorkGroup {
	return &lintWorkGroup{work: core.NewWorkGroup(singleThreaded)}
}

// newProgramWorkGroup bounds active projects to the execution CPU budget.
// Native lint and type-check own separate groups, and each project retains
// its independent checker-shard scheduling.
func newProgramWorkGroup(singleThreaded bool, programCount int) *lintWorkGroup {
	w := newLintWorkGroup(singleThreaded)
	if !singleThreaded && programCount > 0 {
		w.admission = make(chan struct{}, min(runtime.GOMAXPROCS(0), programCount))
	}
	return w
}

func (w *lintWorkGroup) Queue(run func()) {
	w.work.Queue(func() {
		if w.admission != nil {
			w.admission <- struct{}{}
			defer func() { <-w.admission }()
		}
		w.panics.run(run)
	})
}

func (w *lintWorkGroup) RunAndWait() {
	w.work.RunAndWait()
	w.panics.rethrow()
}

func (p *workerPanic) run(run func()) {
	if p.aborted.Load() {
		return
	}
	completed := false
	defer func() {
		if completed {
			return
		}
		value := recover()
		if value == nil {
			value = errors.New("linter: worker exited without completing")
		}
		p.panicOnce.Do(func() {
			p.panicValue = value
			p.aborted.Store(true)
		})
	}()
	run()
	completed = true
}

func (p *workerPanic) rethrow() {
	if p.aborted.Load() {
		panic(p.panicValue)
	}
}
