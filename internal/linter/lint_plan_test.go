package linter

import (
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestPreparedLintPlanPreservesNativeSemanticsAndIsReused(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts":    "const a = 1;",
		"bad.ts":  "const bad = 1;",
		"gap.ts":  "const gap = 1;",
		"zero.ts": "const zero = 1;",
	})
	targets := [][]string{{paths["a.ts"], paths["bad.ts"], paths["gap.ts"], paths["zero.ts"]}}
	syntaxErrorFiles := map[string]struct{}{paths["bad.ts"]: {}}
	typeInfoFiles := map[string]struct{}{paths["a.ts"]: {}}

	newRuleHandler := func(calls map[string]int) RuleHandler {
		return func(file *ast.SourceFile) []ConfiguredRule {
			calls[file.FileName()]++
			switch file.FileName() {
			case paths["a.ts"]:
				rules := noopRule()
				return append(rules, ConfiguredRule{
					Name:               "community/plugin-rule",
					Severity:           rule.SeverityWarning,
					IsEslintPluginRule: true,
				})
			case paths["gap.ts"]:
				typeAwareRule := noopRule()[0]
				typeAwareRule.Name = "type-aware-rule"
				typeAwareRule.RequiresTypeInfo = true
				return []ConfiguredRule{typeAwareRule}
			default:
				return nil
			}
		}
	}

	directCalls := make(map[string]int)
	directResult, err := RunLinter(RunLinterOptions{
		Programs:         []*compiler.Program{program},
		SingleThreaded:   true,
		TargetFiles:      targets,
		GetRulesForFile:  newRuleHandler(directCalls),
		TypeInfoFiles:    typeInfoFiles,
		SyntaxErrorFiles: syntaxErrorFiles,
	})
	if err != nil {
		t.Fatalf("direct RunLinter failed: %v", err)
	}

	preparedCalls := make(map[string]int)
	preparedOpts := RunLinterOptions{
		Programs:         []*compiler.Program{program},
		SingleThreaded:   true,
		TargetFiles:      targets,
		GetRulesForFile:  newRuleHandler(preparedCalls),
		TypeInfoFiles:    typeInfoFiles,
		SyntaxErrorFiles: syntaxErrorFiles,
	}
	preparedOpts.PreparedPlan = PrepareLintPlan(preparedOpts)
	pluginTargets := preparedOpts.PreparedPlan.Targets()
	if len(pluginTargets) != 1 || pluginTargets[0].File.FileName() != paths["a.ts"] {
		t.Fatalf("prepared plugin projection = %+v, want only a.ts", pluginTargets)
	}
	if len(pluginTargets[0].Rules) != 2 {
		t.Fatalf("prepared a.ts rules = %d, want native and plugin rules", len(pluginTargets[0].Rules))
	}

	preparedResult, err := RunLinter(preparedOpts)
	if err != nil {
		t.Fatalf("prepared RunLinter failed: %v", err)
	}
	if !reflect.DeepEqual(preparedResult, directResult) {
		t.Fatalf("prepared result = %#v, direct result = %#v", preparedResult, directResult)
	}
	if preparedResult.LintedFileCount != 4 {
		t.Fatalf("prepared LintedFileCount = %d, want syntax and zero-rule files included", preparedResult.LintedFileCount)
	}
	if _, ok := preparedResult.ExecutedRules["community/plugin-rule"]; !ok {
		t.Fatal("prepared ExecutedRules omitted the configured plugin rule")
	}
	if _, ok := preparedResult.ExecutedRules["type-aware-rule"]; ok {
		t.Fatal("prepared ExecutedRules retained a type-aware rule for a gap file")
	}
	if !reflect.DeepEqual(preparedCalls, directCalls) {
		t.Fatalf("prepared callback calls = %v, direct callback calls = %v", preparedCalls, directCalls)
	}
	if preparedCalls[paths["a.ts"]] != 1 || preparedCalls[paths["gap.ts"]] != 1 || preparedCalls[paths["zero.ts"]] != 1 {
		t.Fatalf("prepared callback should run exactly once per eligible file, got %v", preparedCalls)
	}
	if preparedCalls[paths["bad.ts"]] != 0 {
		t.Fatalf("prepared callback ran for syntax-error file: %v", preparedCalls)
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
	done := make(chan *LintPlan, 1)
	var active atomic.Int32
	var signaled atomic.Bool

	opts := RunLinterOptions{
		Programs:    []*compiler.Program{program},
		TargetFiles: [][]string{{paths["a.ts"], paths["b.ts"], paths["c.ts"], paths["d.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			if active.Add(1) >= 2 && signaled.CompareAndSwap(false, true) {
				close(twoActive)
			}
			<-release
			active.Add(-1)
			return noopRule()
		},
		SyntaxErrorFiles: map[string]struct{}{},
	}
	go func() {
		done <- PrepareLintPlan(opts)
	}()

	select {
	case <-twoActive:
		close(release)
	case <-time.After(5 * time.Second):
		close(release)
		<-done
		t.Fatal("rule resolution did not overlap across workers")
	}
	if plan := <-done; len(plan.Targets()) != 4 {
		t.Fatalf("prepared targets = %d, want 4", len(plan.Targets()))
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
	plan := PrepareLintPlan(RunLinterOptions{
		Programs:         []*compiler.Program{program},
		SingleThreaded:   true,
		TargetFiles:      [][]string{wantOrder},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(file *ast.SourceFile) []ConfiguredRule {
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

func TestPreparedLintPlanPreservesSameFileAcrossPrograms(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"shared.ts": "const shared = 1;",
	})
	calls := 0
	opts := RunLinterOptions{
		Programs:         []*compiler.Program{program, program},
		SingleThreaded:   true,
		TargetFiles:      [][]string{{paths["shared.ts"]}, {paths["shared.ts"]}},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			calls++
			return noopRule()
		},
	}
	opts.PreparedPlan = PrepareLintPlan(opts)
	if targets := opts.PreparedPlan.Targets(); len(targets) != 2 {
		t.Fatalf("prepared targets = %d, want one entry per Program", len(targets))
	}
	result, err := RunLinter(opts)
	if err != nil {
		t.Fatalf("RunLinter failed: %v", err)
	}
	if result.LintedFileCount != 2 {
		t.Fatalf("LintedFileCount = %d, want one count per Program", result.LintedFileCount)
	}
	if calls != 2 {
		t.Fatalf("GetRulesForFile calls = %d, want one per Program and no execution-time repeats", calls)
	}
}

func TestRunLinterRejectsPreparedPlanForDifferentProgram(t *testing.T) {
	programA, pathsA := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	programB, pathsB := createTestProgramWithFiles(t, map[string]string{
		"b.ts": "const b = 1;",
	})
	plan := PrepareLintPlan(RunLinterOptions{
		Programs:         []*compiler.Program{programA},
		SingleThreaded:   true,
		TargetFiles:      [][]string{{pathsA["a.ts"]}},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile:  func(*ast.SourceFile) []ConfiguredRule { return noopRule() },
	})
	_, err := RunLinter(RunLinterOptions{
		Programs:         []*compiler.Program{programB},
		SingleThreaded:   true,
		TargetFiles:      [][]string{{pathsB["b.ts"]}},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile:  func(*ast.SourceFile) []ConfiguredRule { return noopRule() },
		PreparedPlan:     plan,
	})
	if err == nil {
		t.Fatal("RunLinter accepted a prepared plan bound to a different Program")
	}
}
