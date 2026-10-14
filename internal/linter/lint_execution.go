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
	programs          []*programLintPlan
	typeCheckPrograms []*program.Program
	options           programRunOptions
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
			if opts.TypeCheck && plan.program.CanProvideProgramDiagnostics() {
				execution.typeCheckPrograms = append(execution.typeCheckPrograms, plan.program)
			}
		}
	} else if opts.TypeCheck {
		if err := validatePrograms(opts.TypeCheckOnlyPrograms); err != nil {
			return nil, err
		}
		for _, sourceProgram := range opts.TypeCheckOnlyPrograms {
			if sourceProgram.CanProvideProgramDiagnostics() {
				execution.typeCheckPrograms = append(execution.typeCheckPrograms, sourceProgram)
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
	clear(e.typeCheckPrograms)
	e.typeCheckPrograms = nil
}

func (e *lintExecution) run(consumer rule.DiagnosticConsumer) *LintResult {
	defer e.close()
	result := &LintResult{ExecutedRules: make(map[string]struct{})}
	programResults := make([]programLintResult, len(e.programs))
	if len(e.programs) > 0 {
		work := newLintWorkGroup(e.options.SingleThreaded)
		var budget *nativeWorkBudget
		if !e.options.SingleThreaded && len(e.programs) > 1 {
			budget = newNativeWorkBudget(runtime.GOMAXPROCS(0))
		}
		for index := range e.programs {
			task := e.programs[index]
			e.programs[index] = nil
			resultSlot := &programResults[index]
			options := e.options
			work.QueueProject(budget, func(lease *nativeWorkLease) {
				defer func() { *task = programLintPlan{} }()
				options.nativeLease = lease
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

	// Keep the native/type-check phase barrier. Only capable Programs retain
	// this second borrow, and type-check consumes its own private input slots.
	programs := e.typeCheckPrograms
	e.typeCheckPrograms = nil
	if len(programs) > 0 {
		runTypeCheckAcrossPrograms(typeCheckRequest{
			Programs:       programs,
			SingleThreaded: e.options.SingleThreaded,
			OnDiagnostic:   consumer.Report,
		})
	}
	return result
}

// nativeWorkBudget bounds overlapping native graphs within one execution.
// The CPU-sized admission budget reserves capacity before checker preparation;
// it does not limit goroutines created internally by the compiler. Only one
// project widens its reservation at a time; its complete shard width remains
// reserved until all borrowers join. The budget owns no source or compiler data.
type nativeWorkBudget struct {
	preparation sync.Mutex
	workers     chan struct{}
}

func newNativeWorkBudget(cpuCount int) *nativeWorkBudget {
	return &nativeWorkBudget{workers: make(chan struct{}, cpuCount)}
}

// nativeWorkLease is confined to its project coordinator. Shard workers never
// change admission state. Abnormal-worker joins remain owned by lintWorkGroup.
type nativeWorkLease struct {
	budget    *nativeWorkBudget
	aborted   *atomic.Bool
	units     int
	preparing bool
}

func (b *nativeWorkBudget) prepare(aborted *atomic.Bool) *nativeWorkLease {
	b.preparation.Lock()
	lease := &nativeWorkLease{budget: b, aborted: aborted, preparing: true}
	if aborted.Load() {
		lease.close()
		return nil
	}
	b.workers <- struct{}{}
	lease.units = 1
	if aborted.Load() {
		lease.close()
		return nil
	}
	return lease
}

// admit uses the actual planned shard count. A wider project occupies the
// entire admission budget while preserving its existing checker/file topology.
func (l *nativeWorkLease) admit(shardCount int) bool {
	if l == nil {
		return true
	}
	if l.aborted.Load() {
		return false
	}
	width := min(shardCount, cap(l.budget.workers))
	if width == 0 {
		<-l.budget.workers
		l.units = 0
	}
	for l.units < width {
		l.budget.workers <- struct{}{}
		l.units++
		if l.aborted.Load() {
			return false
		}
	}
	l.finishPreparation()
	return !l.aborted.Load()
}

func (l *nativeWorkLease) finishPreparation() {
	if l.preparing {
		l.preparing = false
		l.budget.preparation.Unlock()
	}
}

func (l *nativeWorkLease) close() {
	if l == nil {
		return
	}
	for range l.units {
		<-l.budget.workers
	}
	l.units = 0
	l.finishPreparation()
}

// lintWorkGroup joins every admitted native borrower before re-raising an
// abnormal worker exit. Nested checker shards use the same barrier, so a
// project task cannot end while one of its file workers still uses it.
type lintWorkGroup struct {
	work   core.WorkGroup
	panics workerPanic
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

func (w *lintWorkGroup) Queue(run func()) {
	w.work.Queue(func() {
		w.panics.run(run)
	})
}

// QueueProject composes project admission with the same abnormal-worker join
// protocol as file tasks. Failure is published before its capacity is returned,
// so a queued project cannot enter preparation after a sibling aborts.
func (w *lintWorkGroup) QueueProject(budget *nativeWorkBudget, run func(*nativeWorkLease)) {
	w.work.Queue(func() {
		var lease *nativeWorkLease
		if budget != nil {
			lease = budget.prepare(&w.panics.aborted)
			if lease == nil {
				return
			}
		}
		defer lease.close()
		w.panics.run(func() { run(lease) })
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
