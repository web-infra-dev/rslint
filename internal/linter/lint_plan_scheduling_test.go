package linter

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
	work := newProgramLintWorkGroup(false, programCount)
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
		work.Queue(callback)
		work.Queue(callback)
		select {
		case <-submitPending:
		case <-release:
		}
		for range programCount - 2 {
			work.Queue(callback)
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
				work := newProgramLintWorkGroup(false, programCount)
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
					work.Queue(func() {
						defer close(failureStopped)
						started <- struct{}{}
						<-releaseFailure
						abort()
					})
					work.Queue(func() {
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
						work.Queue(func() { pendingRan.Add(1) })
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

func TestDynamicProgramAdmissionReplenishesBeforeProgramCompletes(t *testing.T) {
	firstReady := make(chan *programReservation, 1)
	secondEntered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	finished := make(chan struct{})
	t.Cleanup(func() {
		unblock()
		joinProgramAdmissionTestWorker(t, finished)
	})
	go func() {
		defer close(finished)
		runWithProgramAdmission([]int{4, 7, 1}, 4, func(index int, reservation *programReservation) {
			switch index {
			case 0:
				reservation.ready(4)
				firstReady <- reservation
				<-release
				for range 3 {
					reservation.done()
				}
			case 1:
				// An oversized estimate may enter as a whole, then shrink to
				// the actual work. No shard needs to be artificially serialized.
				reservation.ready(2)
				reservation.done()
				reservation.done()
				close(secondEntered)
			case 2:
				reservation.ready(0)
			}
		})
	}()
	var first *programReservation
	select {
	case first = <-firstReady:
	case <-time.After(5 * time.Second):
		t.Fatal("first Program did not enter")
	}
	select {
	case <-secondEntered:
		t.Fatal("admitted another Program before any shard completed")
	case <-time.After(20 * time.Millisecond):
	}
	first.done()
	select {
	case <-secondEntered:
	case <-time.After(5 * time.Second):
		t.Fatal("did not replenish after one shard completed while its Program was still active")
	}
	unblock()
}

func TestDynamicProgramAdmissionReusesMixedPlan(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)
	var calls atomic.Int32
	run := func(rule.RuleContext) rule.RuleListeners {
		calls.Add(1)
		return nil
	}
	raw, paths := createTestProgramWithFilesAndCompilerOptions(t, map[string]string{
		"one.ts":    "export const one = 1;",
		"two.ts":    "export const two = 2;",
		"plugin.ts": "export {};",
		"empty.ts":  "export {};",
	}, `{"noLib":true}`)
	template := program.NewFromCompiler(raw)
	var programs []*program.Program
	var targets [][]string
	for _, count := range []int{1, 4, 8} {
		options := *raw.Options()
		options.Checkers = &count
		parallel, err := utils.CreateProgramFromOptions(false, &options, template.RootFileNames(),
			utils.CreateCompilerHost(template.CurrentDirectory(), template.FS()))
		if err != nil {
			t.Fatal(err)
		}
		programs = append(programs, program.NewFromCompiler(parallel))
		targets = append(targets, []string{paths["one.ts"], paths["two.ts"], paths["plugin.ts"], paths["empty.ts"]})
	}
	plain := checkerFreeExecutionTestOptions(t, false, run).LintPlan.programs[0]
	plainTargets := make([]string, len(plain.files))
	for index, file := range plain.files {
		plainTargets[index] = file.file.FileName()
	}
	// Keep a shared compiler, a checker-free Program large enough to split,
	// a plugin-only task and an empty task in the same reusable plan.
	programs = append(programs, plain.program, programs[1], programs[0], programs[0])
	targets = append(targets, plainTargets, targets[1], []string{paths["plugin.ts"]}, nil)
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs: programs, TargetsByProgram: targets,
		GetRulesForFile: func(file *ast.SourceFile) []rule.ConfiguredRule {
			switch file.FileName() {
			case paths["empty.ts"]:
				return nil
			case paths["plugin.ts"]:
				return []rule.ConfiguredRule{{Name: "plugin", IsEslintPluginRule: true}}
			default:
				return []rule.ConfiguredRule{{Name: "native/admission", Run: run}}
			}
		},
	})
	for _, single := range []bool{true, false} {
		for _, typeCheck := range []bool{false, true} {
			before := calls.Load()
			result, err := RunLinter(RunLinterOptions{
				LintPlan: plan, SingleThreaded: single, TypeCheck: typeCheck,
				Consumer: rule.DiagnosticConsumer{Report: func(rule.RuleDiagnostic) {}},
			})
			if err != nil || result == nil {
				t.Fatalf("RunLinter: result=%v error=%v", result, err)
			}
			if result.LintedFileCount != 273 || calls.Load()-before != 264 ||
				!reflect.DeepEqual(result.ExecutedRules, map[string]struct{}{"native/admission": {}, "plugin": {}}) {
				t.Fatalf("reused plan: result=%+v native calls=%d", result, calls.Load()-before)
			}
		}
	}
}

func TestDynamicProgramAdmissionJoinsAbnormalWorkers(t *testing.T) {
	for _, exit := range []string{"panic", "Goexit"} {
		t.Run(exit, func(t *testing.T) {
			release := make(chan struct{})
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(release) }) }
			started := make(chan struct{}, 2)
			failed := make(chan struct{})
			done := make(chan any, 1)
			finished := make(chan struct{})
			t.Cleanup(func() {
				unblock()
				joinProgramAdmissionTestWorker(t, finished)
			})
			go func() {
				defer close(finished)
				defer func() { done <- recover() }()
				runWithProgramAdmission([]int{1, 1, 4, 4, 4}, 2, func(index int, reservation *programReservation) {
					if index > 1 {
						return
					}
					started <- struct{}{}
					if index == 0 {
						<-started
						<-started
						defer close(failed)
						// Fail before ready, leaving the estimate to be returned.
						if exit == "Goexit" {
							runtime.Goexit()
						}
						panic("admission-marker")
					}
					reservation.ready(3)
					reservation.done()
					<-release // close must return the two unstarted shards.
				})
			}()
			select {
			case <-failed:
			case <-time.After(5 * time.Second):
				t.Fatal("first Program did not fail")
			}
			select {
			case failure := <-done:
				t.Fatalf("propagated abnormal exit before its sibling joined: %v", failure)
			case <-time.After(20 * time.Millisecond):
			}
			unblock()
			select {
			case failure := <-done:
				if failure == nil || (exit == "panic" && failure != "admission-marker") {
					t.Fatalf("lost original abnormal exit: %v", failure)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("abnormal worker stranded the dispatcher")
			}
		})
	}
}
