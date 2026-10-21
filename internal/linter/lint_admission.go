package linter

import "sync"

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
