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
		runProgram := func(index int, reservation *programReservation) {
			task := e.programs[index]
			e.programs[index] = nil
			defer func() { *task = programLintPlan{} }()
			options := e.options
			options.reservation = reservation
			programResults[index] = runLintRulesInProgram(task, options, consumer)
		}
		budget := runtime.GOMAXPROCS(0)
		if !e.options.SingleThreaded && len(e.programs) > budget {
			estimates := make([]int, len(e.programs))
			for index, task := range e.programs {
				estimates[index] = estimateProgramLintShards(task, budget)
			}
			runWithProgramAdmission(estimates, budget, runProgram)
		} else {
			work := newLintWorkGroup(e.options.SingleThreaded)
			for index := range e.programs {
				work.Queue(func() { runProgram(index, nil) })
			}
			work.RunAndWait()
		}
		e.programs = nil
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

// estimateProgramLintShards reads only the frozen native file projection. It
// must not acquire a checker: admission precedes binding/checker creation.
// The executor corrects this hint after discovering the actual associations.
func estimateProgramLintShards(plan *programLintPlan, maxWorkers int) int {
	capable, plain := 0, 0
	for _, file := range plan.files {
		if file.rules == nil || len(file.rules.native) == 0 {
			continue
		}
		if file.hasTypeChecker {
			capable++
		} else {
			plain++
		}
	}
	if capable == 0 {
		return max(1, checkerFreeLintWorkerCount(plain, maxWorkers))
	}
	// Match the compiler pool's default shape without making it a contract.
	checkers := 4
	if configured := plan.program.Options().Checkers; configured != nil {
		checkers = max(1, *configured)
	}
	shards := min(capable, checkers)
	if plain > 0 {
		shards++
	}
	return shards
}

// programAdmission reserves unfinished shards before a Program can create
// checkers. One dispatcher waits for capacity; workers only notify it, so the
// number of Program goroutines is bounded without waking a queue of waiters.
type programAdmission struct {
	mu       sync.Mutex
	changed  chan struct{}
	limit    int
	shards   int
	programs int
}

type programReservation struct {
	admission *programAdmission
	remaining int // protected by admission.mu
}

func runWithProgramAdmission(estimates []int, limit int, run func(int, *programReservation)) {
	admission := &programAdmission{limit: limit, changed: make(chan struct{}, 1)}
	var workers sync.WaitGroup
	var panics workerPanic
	for index, estimate := range estimates {
		reservation := admission.acquire(estimate)
		if panics.aborted.Load() {
			reservation.close()
			break
		}
		workers.Add(1)
		go func() {
			defer workers.Done()
			defer reservation.close()
			panics.run(func() { run(index, reservation) })
		}()
	}
	workers.Wait()
	panics.rethrow()
}

func (a *programAdmission) acquire(estimate int) *programReservation {
	for {
		a.mu.Lock()
		if a.shards < a.limit && a.programs < a.limit {
			// Admit a whole Program, including one larger than the budget.
			// This preserves its checker parallelism and guarantees progress.
			a.shards += estimate
			a.programs++
			a.mu.Unlock()
			return &programReservation{admission: a, remaining: estimate}
		}
		a.mu.Unlock()
		<-a.changed
	}
}

// notify is called with mu held. Only the dispatcher waits on changed, and a
// buffered notification makes releases before its next wait observable.
func (a *programAdmission) notify() {
	select {
	case a.changed <- struct{}{}:
	default:
	}
}

// ready replaces the estimate once the executor knows the actual shard groups.
// A nil reservation is the small-project and single-threaded execution path.
func (r *programReservation) ready(shards int) {
	if r == nil {
		return
	}
	a := r.admission
	a.mu.Lock()
	defer a.mu.Unlock()
	a.shards += shards - r.remaining
	r.remaining = shards
	a.notify()
}

func (r *programReservation) done() {
	if r == nil {
		return
	}
	a := r.admission
	a.mu.Lock()
	defer a.mu.Unlock()
	r.remaining--
	a.shards--
	a.notify()
}

// close runs after all entered shards join. It returns any reservation for
// work skipped by an abnormal exit, and only then releases the Program slot.
func (r *programReservation) close() {
	a := r.admission
	a.mu.Lock()
	defer a.mu.Unlock()
	a.shards -= r.remaining
	a.programs--
	a.notify()
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
