package linter

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreparedRuleSetsPublishCompleteConcurrentViews(t *testing.T) {
	input := make([]rule.ConfiguredRule, 100)
	for index := range input {
		input[index] = rule.ConfiguredRule{Name: "rule", RequiresTypeInfo: index%2 == 0, IsEslintPluginRule: index%3 == 0}
	}
	var builder ruleSetBuilder
	var workers sync.WaitGroup
	start := make(chan struct{})
	results := make([]*preparedRuleSet, 32)
	for index := range results {
		workers.Go(func() {
			<-start
			results[index] = builder.prepare(input, index%2 == 0)
		})
	}
	close(start)
	workers.Wait()
	for index, got := range results {
		want := prepareRuleSet(input, index%2 == 0)
		if !reflect.DeepEqual(got, want) || got != results[index%2] {
			t.Fatal("concurrent cold readers did not receive one complete immutable result")
		}
	}
}

func TestCheckerFreeLintWorkerCountKeepsSmallSetsSerial(t *testing.T) {
	tests := []struct {
		files int
		procs int
		want  int
	}{
		{files: 0, procs: 10, want: 0},
		{files: 1, procs: 10, want: 1},
		{files: 255, procs: 10, want: 1},
		{files: 256, procs: 10, want: 2},
		{files: 1119, procs: 10, want: 8},
		{files: 10200, procs: 10, want: 10},
	}
	for _, test := range tests {
		if got := checkerFreeLintWorkerCount(test.files, test.procs); got != test.want {
			t.Fatalf("checkerFreeLintWorkerCount(%d, %d) = %d, want %d", test.files, test.procs, got, test.want)
		}
	}
}

func TestRunLinterParallelizesCheckerFreeFiles(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	release := make(chan struct{})
	twoActive := make(chan struct{})
	var active atomic.Int32
	var signaled atomic.Bool
	opts := checkerFreeExecutionTestOptions(t, false, func(rule.RuleContext) rule.RuleListeners {
		if active.Add(1) >= 2 && signaled.CompareAndSwap(false, true) {
			close(twoActive)
		}
		<-release
		active.Add(-1)
		return nil
	})

	type outcome struct {
		result *LintResult
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		result, err := RunLinter(opts)
		done <- outcome{result: result, err: err}
	}()

	select {
	case <-twoActive:
		close(release)
	case <-time.After(5 * time.Second):
		close(release)
		result := <-done
		if result.err != nil {
			t.Fatalf("RunLinter: %v", result.err)
		}
		t.Fatal("checker-free files did not overlap across workers")
	}
	result := <-done
	if result.err != nil {
		t.Fatalf("RunLinter: %v", result.err)
	}
	wantFiles := int32(minCheckerFreeFilesPerLintWorker * 2)
	if result.result.LintedFileCount != wantFiles {
		t.Fatalf("LintedFileCount = %d, want %d", result.result.LintedFileCount, wantFiles)
	}
}

func TestRunLinterSingleThreadedSerializesCheckerFreeFiles(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	var active atomic.Int32
	var maxActive atomic.Int32
	opts := checkerFreeExecutionTestOptions(t, true, func(rule.RuleContext) rule.RuleListeners {
		current := active.Add(1)
		for observed := maxActive.Load(); current > observed; observed = maxActive.Load() {
			if maxActive.CompareAndSwap(observed, current) {
				break
			}
		}
		time.Sleep(time.Millisecond)
		active.Add(-1)
		return nil
	})
	result, err := RunLinter(opts)
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	wantFiles := int32(minCheckerFreeFilesPerLintWorker * 2)
	if result.LintedFileCount != wantFiles {
		t.Fatalf("LintedFileCount = %d, want %d", result.LintedFileCount, wantFiles)
	}
	if maxActive.Load() != 1 {
		t.Fatalf("single-threaded checker-free execution had %d active files", maxActive.Load())
	}
}

func TestPrepareLintPlanParallelizesRuleResolution(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 1;",
		"c.ts": "const c = 1;",
		"d.ts": "const d = 1;",
	})
	release := make(chan struct{})
	twoActive := make(chan struct{})
	type prepareResult struct {
		plan *LintPlan
		err  error
	}
	done := make(chan prepareResult, 1)
	var active atomic.Int32
	var signaled atomic.Bool

	opts := PrepareLintPlanOptions{
		Programs:         wrapTestPrograms(program),
		TargetsByProgram: [][]string{{paths["a.ts"], paths["b.ts"], paths["c.ts"], paths["d.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			if active.Add(1) >= 2 && signaled.CompareAndSwap(false, true) {
				close(twoActive)
			}
			<-release
			active.Add(-1)
			return noopRule()
		},
	}
	go func() {
		plan, err := PrepareLintPlan(opts)
		done <- prepareResult{plan: plan, err: err}
	}()

	select {
	case <-twoActive:
		close(release)
	case <-time.After(5 * time.Second):
		close(release)
		<-done
		t.Fatal("rule resolution did not overlap across workers")
	}
	prepared := <-done
	if prepared.err != nil {
		t.Fatalf("PrepareLintPlan: %v", prepared.err)
	}
	if len(prepared.plan.Targets()) != 4 {
		t.Fatalf("prepared targets = %d, want 4", len(prepared.plan.Targets()))
	}
}

func TestPrepareLintPlanHonorsSingleThreadedOrder(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 1;",
		"c.ts": "const c = 1;",
	})
	wantOrder := []string{paths["c.ts"], paths["a.ts"], paths["b.ts"]}
	var gotOrder []string
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         wrapTestPrograms(program),
		SingleThreaded:   true,
		TargetsByProgram: [][]string{wantOrder},
		GetRulesForFile: func(file *ast.SourceFile) []rule.ConfiguredRule {
			gotOrder = append(gotOrder, file.FileName())
			return noopRule()
		},
	})
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("single-threaded resolution order = %v, want %v", gotOrder, wantOrder)
	}
	if gotTargets := plan.Targets(); len(gotTargets) != len(wantOrder) {
		t.Fatalf("single-threaded prepared targets = %d, want %d", len(gotTargets), len(wantOrder))
	}
}

func TestPreparedLintPlanPreservesSameFileAcrossProgramsInParallel(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"shared.ts": "const shared = 1;",
	})
	var calls atomic.Int32
	programs := wrapTestPrograms(program, program)
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{paths["shared.ts"]}, {paths["shared.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			calls.Add(1)
			return noopRule()
		},
	})
	if targets := plan.Targets(); len(targets) != 2 {
		t.Fatalf("prepared targets = %d, want one entry per Program", len(targets))
	}
	result, err := RunLinter(RunLinterOptions{
		LintPlan: plan,
	})
	if err != nil {
		t.Fatalf("RunLinter failed: %v", err)
	}
	if result.LintedFileCount != 2 {
		t.Fatalf("LintedFileCount = %d, want one count per Program", result.LintedFileCount)
	}
	if calls.Load() != 2 {
		t.Fatalf("GetRulesForFile calls = %d, want one per Program and no execution-time repeats", calls.Load())
	}
}

func TestPrepareLintPlanDeduplicatesTargetsInFirstOccurrenceOrder(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 1;",
	})
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         wrapTestPrograms(raw),
		TargetsByProgram: [][]string{{paths["b.ts"], paths["a.ts"], paths["b.ts"]}},
		SingleThreaded:   true,
		GetRulesForFile:  func(*ast.SourceFile) []rule.ConfiguredRule { return noopRule() },
	})
	targets := plan.Targets()
	if len(targets) != 2 || targets[0].File.FileName() != paths["b.ts"] || targets[1].File.FileName() != paths["a.ts"] {
		t.Fatalf("deduplicated target order = %v, want [%q %q]", targetFileNames(targets), paths["b.ts"], paths["a.ts"])
	}
}

func targetFileNames(targets []LintTarget) []string {
	fileNames := make([]string, len(targets))
	for index, target := range targets {
		fileNames[index] = target.File.FileName()
	}
	return fileNames
}

func TestRunLinterLimitsProgramAdmission(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	t.Cleanup(func() { runtime.GOMAXPROCS(previousProcs) })
	const programCount = 6
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	entered := make(chan struct{}, programCount)
	var active, peak, completed atomic.Int32
	opts, _ := programAdmissionTestOptions(t, programCount, false, func(rule.RuleContext) rule.RuleListeners {
		current := active.Add(1)
		updateProgramAdmissionPeak(&peak, current)
		entered <- struct{}{}
		<-release
		active.Add(-1)
		completed.Add(1)
		return nil
	})
	done := make(chan programAdmissionOutcome, 1)
	finished := make(chan struct{})
	t.Cleanup(func() {
		unblock()
		joinProgramAdmissionTestWorker(t, finished)
	})
	go func() {
		defer close(finished)
		result, err := RunLinter(opts)
		done <- programAdmissionOutcome{result: result, err: err}
	}()
	for range 2 {
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			t.Fatal("two independent Programs did not become active")
		}
	}
	select {
	case <-entered:
		t.Fatal("a third Program entered while two admitted Programs remained blocked")
	case <-time.After(30 * time.Millisecond):
	}
	unblock()
	select {
	case got := <-done:
		if got.err != nil || got.result == nil || got.result.LintedFileCount != programCount ||
			completed.Load() != programCount || peak.Load() != 2 || active.Load() != 0 {
			t.Fatalf("result/error/completed/peak/active = %+v/%v/%d/%d/%d", got.result, got.err, completed.Load(), peak.Load(), active.Load())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("admitted and pending Programs did not all complete")
	}
}

func TestRunLinterProgramAdmissionSingleWorker(t *testing.T) {
	for _, test := range []struct {
		name           string
		procs          int
		singleThreaded bool
	}{{name: "GOMAXPROCS-one", procs: 1}, {name: "single-threaded", procs: 2, singleThreaded: true}} {
		t.Run(test.name, func(t *testing.T) {
			previousProcs := runtime.GOMAXPROCS(test.procs)
			defer runtime.GOMAXPROCS(previousProcs)
			const programCount = 5
			var active, peak atomic.Int32
			var orderMu sync.Mutex
			var gotOrder []string
			opts, paths := programAdmissionTestOptions(t, programCount, test.singleThreaded, func(ctx rule.RuleContext) rule.RuleListeners {
				current := active.Add(1)
				updateProgramAdmissionPeak(&peak, current)
				orderMu.Lock()
				gotOrder = append(gotOrder, ctx.SourceFile.FileName())
				orderMu.Unlock()
				runtime.Gosched()
				active.Add(-1)
				return nil
			})
			done := make(chan programAdmissionOutcome, 1)
			go func() {
				result, err := RunLinter(opts)
				done <- programAdmissionOutcome{result: result, err: err}
			}()
			select {
			case got := <-done:
				if got.err != nil || got.result == nil || got.result.LintedFileCount != programCount || peak.Load() != 1 || active.Load() != 0 {
					t.Fatalf("result/error/peak/active = %+v/%v/%d/%d", got.result, got.err, peak.Load(), active.Load())
				}
			case <-time.After(5 * time.Second):
				t.Fatal("single-worker Program admission deadlocked")
			}
			if test.singleThreaded {
				// core's serial workgroup consumes its submitted callbacks in reverse order.
				// Program admission must keep that existing execution order.
				wantOrder := make([]string, len(paths))
				for index := range paths {
					wantOrder[index] = paths[len(paths)-1-index]
				}
				if !reflect.DeepEqual(gotOrder, wantOrder) {
					t.Fatalf("single-threaded Program order = %v, want %v", gotOrder, wantOrder)
				}
			}
		})
	}
}

func TestProgramLintWorkGroupQueuesBeforeAdmission(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	t.Cleanup(func() { runtime.GOMAXPROCS(previousProcs) })
	const programCount = 8
	work, queue := scheduledProjectTestQueue(2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	entered := make(chan struct{}, programCount)
	var completed atomic.Int32
	callback := func() {
		entered <- struct{}{}
		<-release
		completed.Add(1)
	}
	queued := make(chan struct{})
	submitPending := make(chan struct{})
	done := make(chan struct{})
	finished := make(chan struct{})
	t.Cleanup(func() {
		unblock()
		joinProgramAdmissionTestWorker(t, finished)
	})
	go func() {
		defer close(finished)
		queue(callback)
		queue(callback)
		select {
		case <-submitPending:
		case <-release:
		}
		for range programCount - 2 {
			queue(callback)
		}
		close(queued)
		work.RunAndWait()
		close(done)
	}()
	for range 2 {
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			t.Fatal("admitted callbacks did not start")
		}
	}
	close(submitPending)
	select {
	case <-queued:
	case <-time.After(5 * time.Second):
		t.Fatal("Queue blocked on Program admission instead of submitting the callback")
	}
	select {
	case <-entered:
		t.Fatal("pending Program callback bypassed the admission limit")
	default:
	}
	unblock()
	select {
	case <-done:
		if completed.Load() != programCount {
			t.Fatalf("completed callbacks = %d, want %d", completed.Load(), programCount)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("pending callbacks did not make progress as admission tokens were released")
	}
}

func TestProgramLintWorkGroupJoinsAbnormalWorkers(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)
	for _, exit := range []string{"panic", "Goexit"} {
		for _, siblingExits := range []bool{false, true} {
			name := fmt.Sprintf("%s/sibling-abnormal-%t", exit, siblingExits)
			t.Run(name, func(t *testing.T) {
				const programCount = 8
				const marker = "admitted Program failed"
				work, queue := scheduledProjectTestQueue(2)
				started := make(chan struct{}, 2)
				releaseFailure := make(chan struct{})
				releaseSibling := make(chan struct{})
				failureStopped := make(chan struct{})
				siblingStopped := make(chan struct{})
				var failureOnce, siblingOnce sync.Once
				unblockFailure := func() { failureOnce.Do(func() { close(releaseFailure) }) }
				unblockSibling := func() { siblingOnce.Do(func() { close(releaseSibling) }) }
				abort := func() {
					if exit == "Goexit" {
						runtime.Goexit()
					}
					panic(marker)
				}
				var pendingRan atomic.Int32
				queued := make(chan struct{})
				submitPending := make(chan struct{})
				done := make(chan any, 1)
				finished := make(chan struct{})
				t.Cleanup(func() {
					unblockFailure()
					unblockSibling()
					joinProgramAdmissionTestWorker(t, finished)
				})
				go func() {
					defer close(finished)
					defer func() { done <- recover() }()
					queue(func() {
						defer close(failureStopped)
						started <- struct{}{}
						<-releaseFailure
						abort()
					})
					queue(func() {
						defer close(siblingStopped)
						started <- struct{}{}
						<-releaseSibling
						if siblingExits {
							abort()
						}
					})
					select {
					case <-submitPending:
					case <-releaseFailure:
					}
					for range programCount - 2 {
						queue(func() { pendingRan.Add(1) })
					}
					close(queued)
					work.RunAndWait()
				}()
				for range 2 {
					select {
					case <-started:
					case <-time.After(5 * time.Second):
						t.Fatal("both admitted Program callbacks did not start")
					}
				}
				close(submitPending)
				select {
				case <-queued:
				case <-time.After(5 * time.Second):
					t.Fatal("pending Program callbacks could not be submitted")
				}
				unblockFailure()
				select {
				case <-failureStopped:
				case <-time.After(5 * time.Second):
					t.Fatal("abnormal Program callback did not exit")
				}
				// The callback's defer precedes workerPanic's abort publication.
				// Wait for that publication only to arrange the already-aborted scenario.
				abortDeadline := time.Now().Add(5 * time.Second)
				for !work.panics.aborted.Load() {
					if time.Now().After(abortDeadline) {
						t.Fatal("abnormal Program exit was not recorded")
					}
					runtime.Gosched()
				}
				select {
				case got := <-done:
					t.Fatalf("abnormal exit propagated before admitted sibling joined: %v", got)
				case <-time.After(30 * time.Millisecond):
				}
				unblockSibling()
				select {
				case got := <-done:
					if exit == "panic" && got != marker {
						t.Fatalf("panic = %v, want original %q", got, marker)
					}
					if exit == "Goexit" {
						err, ok := got.(error)
						if !ok || err == nil {
							t.Fatalf("Goexit outcome = %v, want abnormal-worker panic", got)
						}
					}
					select {
					case <-siblingStopped:
					default:
						t.Fatal("workgroup returned before the admitted sibling stopped")
					}
					if pendingRan.Load() != 0 {
						t.Fatalf("%d pending Program callbacks ran after the workgroup aborted", pendingRan.Load())
					}
				case <-time.After(5 * time.Second):
					t.Fatal("admitted or skipped callbacks retained admission tokens and blocked join")
				}
			})
		}
	}
}

type programAdmissionOutcome struct {
	result *LintResult
	err    error
}

func updateProgramAdmissionPeak(peak *atomic.Int32, current int32) {
	for observed := peak.Load(); current > observed; observed = peak.Load() {
		if peak.CompareAndSwap(observed, current) {
			return
		}
	}
}

func joinProgramAdmissionTestWorker(t *testing.T, finished <-chan struct{}) {
	t.Helper()
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Error("Program admission worker did not finish during cleanup")
	}
}

func programAdmissionTestOptions(t *testing.T, count int, singleThreaded bool, run func(rule.RuleContext) rule.RuleListeners) (RunLinterOptions, []string) {
	t.Helper()
	root := tspath.NormalizePath(t.TempDir())
	programs := make([]*program.Program, count)
	paths := make([]string, count)
	targets := make([][]string, count)
	for index := range programs {
		paths[index] = tspath.ResolvePath(root, fmt.Sprintf("project-%02d/file.ts", index))
		// This existing helper builds a distinct parser/binder Program for each root.
		programs[index] = pipelineTestProgram(t, root, paths[index], "const value = 1;")
		if programs[index].CanProvideProgramDiagnostics() {
			t.Fatal("Program admission fixture unexpectedly owns compiler checkers")
		}
		targets[index] = []string{paths[index]}
	}
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs: programs, TargetsByProgram: targets, SingleThreaded: singleThreaded,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: "native/program-admission", Run: run}}
		},
	})
	return RunLinterOptions{LintPlan: plan, SingleThreaded: singleThreaded}, paths
}

func scheduledProjectTestQueue(cpuCount int) (*lintWorkGroup, func(func())) {
	work := newLintWorkGroup(false)
	budget := newNativeWorkBudget(cpuCount)
	return work, func(run func()) {
		work.QueueProject(budget, func(lease *nativeWorkLease) {
			if lease.admit(1) {
				run()
			}
		})
	}
}

func TestRunLinterSharesCapacityAcrossProjectShards(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	t.Cleanup(func() { runtime.GOMAXPROCS(previousProcs) })
	const fileCount = minCheckerFreeFilesPerLintWorker * 4
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	entered := make(chan struct{}, fileCount)
	var active, peak, completed atomic.Int32
	run := func(rule.RuleContext) rule.RuleListeners {
		current := active.Add(1)
		updateProgramAdmissionPeak(&peak, current)
		entered <- struct{}{}
		<-release
		active.Add(-1)
		completed.Add(1)
		return nil
	}
	first := checkerFreeExecutionTestOptions(t, false, run)
	second := checkerFreeExecutionTestOptions(t, false, run)
	plan := &LintPlan{programs: append(append([]programLintPlan(nil), first.LintPlan.programs...), second.LintPlan.programs...)}
	done := make(chan programAdmissionOutcome, 1)
	finished := make(chan struct{})
	t.Cleanup(func() { unblock(); joinProgramAdmissionTestWorker(t, finished) })
	go func() {
		defer close(finished)
		result, err := RunLinter(RunLinterOptions{LintPlan: plan})
		done <- programAdmissionOutcome{result: result, err: err}
	}()
	for range 2 {
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			t.Fatal("admitted shards did not run in parallel")
		}
	}
	select {
	case <-entered:
		t.Fatal("independent projects multiplied the shared shard capacity")
	case <-time.After(30 * time.Millisecond):
	}
	unblock()
	select {
	case got := <-done:
		if got.err != nil || got.result == nil || got.result.LintedFileCount != fileCount ||
			completed.Load() != fileCount || peak.Load() != 2 || active.Load() != 0 {
			t.Fatalf("result/error/completed/peak/active = %+v/%v/%d/%d/%d", got.result, got.err, completed.Load(), peak.Load(), active.Load())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("pending project shards did not complete")
	}
}

func TestNativeWorkBudgetReservesPreparationCapacity(t *testing.T) {
	budget := newNativeWorkBudget(1)
	var aborted atomic.Bool
	first := budget.prepare(&aborted)
	if !first.admit(1) {
		t.Fatal("first project was not admitted")
	}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(first.close) }
	entered := make(chan struct{})
	finished := make(chan struct{})
	t.Cleanup(func() { release(); joinProgramAdmissionTestWorker(t, finished) })
	go func() {
		defer close(finished)
		second := budget.prepare(&aborted)
		defer second.close()
		close(entered)
		if !second.admit(0) {
			t.Error("zero-task project was not admitted")
		}
	}()
	select {
	case <-entered:
		t.Fatal("preparation started while the only CPU unit was occupied")
	case <-time.After(30 * time.Millisecond):
	}
	release()
	joinProgramAdmissionTestWorker(t, finished)
	if len(budget.workers) != 0 {
		t.Fatal("zero-task project retained preparation capacity")
	}
}

func TestNativeWorkBudgetHoldsWholeProjectWidth(t *testing.T) {
	budget := newNativeWorkBudget(3)
	var aborted atomic.Bool
	first := budget.prepare(&aborted)
	if !first.admit(2) {
		t.Fatal("first project was not admitted")
	}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(first.close) }
	prepared := make(chan struct{})
	admitted := make(chan struct{})
	finished := make(chan struct{})
	t.Cleanup(func() { release(); joinProgramAdmissionTestWorker(t, finished) })
	go func() {
		defer close(finished)
		second := budget.prepare(&aborted)
		defer second.close()
		close(prepared)
		if !second.admit(2) {
			t.Error("second project was not admitted")
			return
		}
		close(admitted)
	}()
	select {
	case <-prepared:
	case <-time.After(5 * time.Second):
		t.Fatal("available capacity could not prepare the next project")
	}
	select {
	case <-admitted:
		t.Fatal("project started without its complete shard width")
	case <-time.After(30 * time.Millisecond):
	}
	release()
	joinProgramAdmissionTestWorker(t, finished)
	select {
	case <-admitted:
	default:
		t.Fatal("project did not start after sibling completion")
	}
	if len(budget.workers) != 0 {
		t.Fatal("joined projects retained capacity")
	}
}

func TestNativeWorkBudgetOversizedProjectOccupiesCapacity(t *testing.T) {
	budget := newNativeWorkBudget(2)
	var aborted atomic.Bool
	lease := budget.prepare(&aborted)
	defer lease.close()
	if !lease.admit(10) || len(budget.workers) != 2 {
		t.Fatal("oversized shard set did not occupy the complete CPU capacity")
	}
	lease.close()
	if len(budget.workers) != 0 {
		t.Fatal("repeated close retained capacity")
	}
}

func TestNativeWorkBudgetPreparationAbnormalExit(t *testing.T) {
	for _, exit := range []string{"panic", "Goexit"} {
		t.Run(exit, func(t *testing.T) {
			work := newLintWorkGroup(false)
			budget := newNativeWorkBudget(2)
			started := make(chan struct{})
			release := make(chan struct{})
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(release) }) }
			var pendingRan atomic.Int32
			done := make(chan any, 1)
			finished := make(chan struct{})
			t.Cleanup(func() { unblock(); joinProgramAdmissionTestWorker(t, finished) })
			go func() {
				defer close(finished)
				defer func() { done <- recover() }()
				work.QueueProject(budget, func(*nativeWorkLease) {
					close(started)
					<-release
					if exit == "Goexit" {
						runtime.Goexit()
					}
					panic("preparation failed")
				})
				<-started
				for range 4 {
					work.QueueProject(budget, func(*nativeWorkLease) { pendingRan.Add(1) })
				}
				unblock()
				work.RunAndWait()
			}()
			select {
			case got := <-done:
				if exit == "panic" && got != "preparation failed" {
					t.Fatalf("panic = %v", got)
				}
				if exit == "Goexit" {
					if err, ok := got.(error); !ok || err == nil {
						t.Fatalf("Goexit outcome = %v", got)
					}
				}
			case <-time.After(5 * time.Second):
				t.Fatal("abnormal preparation retained a mutex or CPU unit")
			}
			if pendingRan.Load() != 0 || len(budget.workers) != 0 {
				t.Fatal("failed preparation started pending work or retained capacity")
			}
			assertNativeWorkBudgetReusable(t, budget)
		})
	}
}

func TestNativeWorkBudgetAbortsWaitingExpansion(t *testing.T) {
	for _, exit := range []string{"panic", "Goexit"} {
		t.Run(exit, func(t *testing.T) {
			work := newLintWorkGroup(false)
			budget := newNativeWorkBudget(3)
			started := make(chan struct{})
			expanding := make(chan struct{})
			release := make(chan struct{})
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(release) }) }
			var nativeRan, pendingRan atomic.Int32
			done := make(chan any, 1)
			finished := make(chan struct{})
			t.Cleanup(func() { unblock(); joinProgramAdmissionTestWorker(t, finished) })
			go func() {
				defer close(finished)
				defer func() { done <- recover() }()
				work.QueueProject(budget, func(lease *nativeWorkLease) {
					if !lease.admit(1) {
						return
					}
					close(started)
					<-release
					if exit == "Goexit" {
						runtime.Goexit()
					}
					panic("native failed during expansion")
				})
				<-started
				work.QueueProject(budget, func(lease *nativeWorkLease) {
					close(expanding)
					if lease.admit(3) {
						nativeRan.Add(1)
					}
				})
				<-expanding
				deadline := time.Now().Add(5 * time.Second)
				for len(budget.workers) != 3 && time.Now().Before(deadline) {
					time.Sleep(time.Millisecond)
				}
				if len(budget.workers) != 3 {
					t.Error("second project did not reach partial expansion")
				}
				work.QueueProject(budget, func(*nativeWorkLease) { pendingRan.Add(1) })
				unblock()
				work.RunAndWait()
			}()
			select {
			case got := <-done:
				if exit == "panic" && got != "native failed during expansion" {
					t.Fatalf("panic = %v", got)
				}
				if exit == "Goexit" {
					if err, ok := got.(error); !ok || err == nil {
						t.Fatalf("Goexit outcome = %v", got)
					}
				}
			case <-time.After(6 * time.Second):
				t.Fatal("aborted expansion did not join")
			}
			if nativeRan.Load() != 0 || pendingRan.Load() != 0 || len(budget.workers) != 0 {
				t.Fatal("aborted expansion dispatched native work or retained capacity")
			}
			assertNativeWorkBudgetReusable(t, budget)
		})
	}
}

func assertNativeWorkBudgetReusable(t *testing.T, budget *nativeWorkBudget) {
	t.Helper()
	finished := make(chan struct{})
	go func() {
		var fresh atomic.Bool
		lease := budget.prepare(&fresh)
		lease.close()
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("joined projects retained the preparation mutex")
	}
}

func TestRunLinterBudgetPreservesCompilerAffinityAndFixOrder(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(4)
	t.Cleanup(func() { runtime.GOMAXPROCS(previousProcs) })
	for _, singleThreaded := range []bool{false, true} {
		t.Run(fmt.Sprintf("singleThreaded=%t", singleThreaded), func(t *testing.T) {
			const filesPerProject = 16
			var programs []*program.Program
			var targets [][]string
			expected := make(map[*ast.SourceFile]*checker.Checker)
			active := make(map[*checker.Checker]*atomic.Int32)
			for range 2 {
				sources := make(map[string]string)
				for index := range filesPerProject {
					sources[fmt.Sprintf("file-%02d.ts", index)] = fmt.Sprintf("import { value as peer } from './file-%02d'; export const value: number = peer + 1;", index^1)
				}
				raw, paths := createTestProgramWithFilesAndCompilerOptions(t, sources, `{"noLib":true}`)
				fs := wrapTestPrograms(raw)[0].FS()
				root := tspath.GetDirectoryPath(paths["file-00.ts"])
				// The shared fixture defaults to a serial compiler. This case
				// needs the compiler's real parallel checker pool in both modes.
				raw, err := utils.CreateProgram(false, fs, root, "tsconfig.json", utils.CreateCompilerHost(root, fs))
				if err != nil {
					t.Fatalf("CreateProgram: %v", err)
				}
				project := wrapTestPrograms(raw)[0]
				programs = append(programs, project)
				var names []string
				projectCheckers := make(map[*checker.Checker]struct{})
				for index := range filesPerProject {
					name := paths[fmt.Sprintf("file-%02d.ts", index)]
					names = append(names, name)
					file := project.GetSourceFile(name)
					chk, release := project.TypeCheckerForFile(context.Background(), file)
					release()
					if chk == nil {
						t.Fatal("compiler fixture did not provide a checker")
					}
					expected[file] = chk
					projectCheckers[chk] = struct{}{}
					if active[chk] == nil {
						active[chk] = new(atomic.Int32)
					}
				}
				if len(projectCheckers) < 2 {
					t.Fatal("selected compiler files did not cover multiple real checker shards")
				}
				targets = append(targets, names)
			}
			if len(active) < 2 {
				t.Fatal("compiler fixture did not cover multiple real checkers")
			}
			textRange := core.NewTextRange(0, 1)
			configured := make([]rule.ConfiguredRule, 0, 2)
			for _, name := range []string{"first", "second"} {
				configured = append(configured, rule.ConfiguredRule{Name: "native/" + name, RequiresTypeInfo: true, Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != expected[ctx.SourceFile] {
						t.Error("native admission changed compiler checker affinity")
						return nil
					}
					counter := active[ctx.TypeChecker]
					if counter.Add(1) != 1 {
						t.Error("one checker was used concurrently")
					}
					defer counter.Add(-1)
					// Widen the overlap window across actual compiler checker tasks.
					time.Sleep(time.Millisecond)
					ctx.ReportRangeWithFixes(textRange, rule.RuleMessage{Id: name, Description: name}, rule.RuleFixReplaceRange(textRange, name))
					return nil
				}})
			}
			plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
				Programs: programs, TargetsByProgram: targets, SingleThreaded: singleThreaded,
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule { return configured },
			})
			var mu sync.Mutex
			diagnostics := make(map[string][]rule.RuleDiagnostic)
			result, err := RunLinter(RunLinterOptions{LintPlan: plan, SingleThreaded: singleThreaded, Consumer: rule.DiagnosticConsumer{
				Demand: rule.EditDemandAutofix,
				Report: func(d rule.RuleDiagnostic) {
					mu.Lock()
					defer mu.Unlock()
					diagnostics[d.FilePath] = append(diagnostics[d.FilePath], d)
				},
			}})
			if err != nil || result.LintedFileCount != filesPerProject*2 || len(diagnostics) != filesPerProject*2 {
				t.Fatalf("result/error/diagnostic files = %+v/%v/%d", result, err, len(diagnostics))
			}
			for name, got := range diagnostics {
				if len(got) != 2 {
					t.Fatalf("%s: diagnostics = %d", name, len(got))
				}
				for index, want := range []string{"first", "second"} {
					fixes := got[index].Fixes()
					if got[index].Message.Id != want || got[index].Range != textRange || len(fixes) != 1 || fixes[0] != rule.RuleFixReplaceRange(textRange, want) {
						t.Fatalf("%s: equal-position diagnostic/fix order changed: %+v", name, got)
					}
				}
			}
		})
	}
}
