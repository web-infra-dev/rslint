package linter

import (
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

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
