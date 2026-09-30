package rule

// Executor owns prepared runners for one serial lint task. Its zero value is
// ready to use. It is not safe for concurrent use; concurrent tasks use separate
// executors, even when they share the same configured rules.
//
// The caller must finish each file's listeners before running the next file and
// call Release after the task. No file context or listener is stored here.
type Executor struct {
	runners map[*preparedRule]FileRunner
}

// Run creates fresh listeners, borrowing a runner on the first use of a prepared
// configuration in this task. Ordinary rules retain their per-file Run path and
// do not allocate a runner map.
func (e *Executor) Run(r ConfiguredRule, ctx RuleContext) RuleListeners {
	if r.prepared == nil {
		return r.Run(ctx)
	}
	runner, ok := e.runners[r.prepared]
	if !ok {
		runner = r.prepared.runners.Get().(FileRunner) //nolint:forcetypeassert // The pool factory and Release only store FileRunner values.
		if e.runners == nil {
			e.runners = make(map[*preparedRule]FileRunner)
		}
		e.runners[r.prepared] = runner
	}
	return runner(ctx)
}

// Release returns this task's runners to their configurations. Call it only
// after all listeners have finished. The executor can then serve another task.
func (e *Executor) Release() {
	for prepared, runner := range e.runners {
		prepared.runners.Put(runner)
	}
	clear(e.runners)
}
