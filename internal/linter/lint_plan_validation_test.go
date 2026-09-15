package linter

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestRunLinterRejectsTypeCheckOnlyProgramsWithLintPlan(t *testing.T) {
	programA, pathsA := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	programB, _ := createTestProgramWithFiles(t, map[string]string{
		"b.ts": "const b = 1;",
	})
	var runs atomic.Int32
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         wrapTestPrograms(programA),
		SingleThreaded:   true,
		TargetsByProgram: [][]string{{pathsA["a.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			configured := noopRule()
			configured[0].Run = func(rule.RuleContext) rule.RuleListeners {
				runs.Add(1)
				return nil
			}
			return configured
		},
	})
	_, err := RunLinter(RunLinterOptions{
		LintPlan:              plan,
		TypeCheckOnlyPrograms: wrapTestPrograms(programB),
		TypeCheck:             true,
	})
	if !errors.Is(err, errTypeCheckOnlyProgramsWithPlan) || runs.Load() != 0 {
		t.Fatalf("RunLinter conflict error = %v, rule runs = %d", err, runs.Load())
	}
}

func TestPrepareLintPlanRequiresTargetsForEveryProgram(t *testing.T) {
	raw, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	programs := wrapTestPrograms(raw)
	var ruleCalls atomic.Int32
	for _, targetsByProgram := range [][][]string{nil, {nil, nil}} {
		_, err := PrepareLintPlan(PrepareLintPlanOptions{
			Programs:         programs,
			TargetsByProgram: targetsByProgram,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				ruleCalls.Add(1)
				return noopRule()
			},
		})
		if !errors.Is(err, errTargetsByProgramLength) {
			t.Fatalf("TargetsByProgram length %d error = %v", len(targetsByProgram), err)
		}
	}
	if ruleCalls.Load() != 0 {
		t.Fatalf("invalid target projection resolved rules %d times", ruleCalls.Load())
	}
}

func TestPrepareLintPlanRequiresRuleHandler(t *testing.T) {
	raw, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	_, err := PrepareLintPlan(PrepareLintPlanOptions{
		Programs:         wrapTestPrograms(raw),
		TargetsByProgram: [][]string{nil},
	})
	if !errors.Is(err, errNilRuleHandler) {
		t.Fatalf("nil GetRulesForFile error = %v", err)
	}
}

func TestPrepareLintPlanContextDoesNotPublishCanceledPlan(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	secondResolverStarted := make(chan struct{})
	releaseSecondResolver := make(chan struct{})
	defer func() {
		select {
		case <-releaseSecondResolver:
		default:
			close(releaseSecondResolver)
		}
	}()
	cancelIssued := make(chan struct{})
	type preparationResult struct {
		plan *LintPlan
		err  error
	}
	resultCh := make(chan preparationResult, 1)
	go func() {
		plan, err := PrepareLintPlanContext(ctx, PrepareLintPlanOptions{
			Programs:         wrapTestPrograms(raw),
			TargetsByProgram: [][]string{{paths["a.ts"], paths["b.ts"]}},
			GetRulesForFile: func(source *ast.SourceFile) []rule.ConfiguredRule {
				switch source.FileName() {
				case paths["a.ts"]:
					<-secondResolverStarted
					cancel()
					close(cancelIssued)
				case paths["b.ts"]:
					close(secondResolverStarted)
					<-releaseSecondResolver
				}
				return noopRule()
			},
		})
		resultCh <- preparationResult{plan: plan, err: err}
	}()

	select {
	case <-cancelIssued:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for cancellation")
	}
	select {
	case result := <-resultCh:
		t.Fatalf("canceled preparation returned before workers joined: (%+v, %v)", result.plan, result.err)
	case <-time.After(50 * time.Millisecond):
	}
	close(releaseSecondResolver)

	select {
	case result := <-resultCh:
		if !errors.Is(result.err, context.Canceled) || result.plan != nil {
			t.Fatalf("canceled preparation = (%+v, %v), want (nil, context.Canceled)", result.plan, result.err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for canceled preparation to join")
	}
}

func TestPrepareLintPlanRejectsTargetOutsideBoundProgramBeforeRuleResolution(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	var ruleCalls atomic.Int32
	missing := paths["a.ts"] + ".missing"
	_, err := PrepareLintPlan(PrepareLintPlanOptions{
		Programs:         wrapTestPrograms(raw),
		TargetsByProgram: [][]string{{paths["a.ts"], missing}},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			ruleCalls.Add(1)
			return noopRule()
		},
	})
	if !errors.Is(err, errTargetNotInProgram) || !strings.Contains(err.Error(), missing) {
		t.Fatalf("missing target error = %v", err)
	}
	if ruleCalls.Load() != 0 {
		t.Fatalf("missing target resolved rules %d times", ruleCalls.Load())
	}
}

func TestPrepareLintPlanDoesNotReapplyDefaultExclusions(t *testing.T) {
	dir := t.TempDir()
	target := norm(dir, "node_modules/pkg/exact.ts")
	writeTestFiles(t, dir, map[string]string{"node_modules/pkg/exact.ts": "export const exact = 1;"})
	raw := gapProgram(t, dir, []string{target})
	if raw.GetSourceFile(target) == nil {
		t.Fatalf("fixture Program did not contain %q", target)
	}
	var ruleCalls atomic.Int32
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         wrapTestPrograms(raw),
		TargetsByProgram: [][]string{{target}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			ruleCalls.Add(1)
			return noopRule()
		},
	})
	result, err := RunLinter(RunLinterOptions{LintPlan: plan, SingleThreaded: true})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if result.LintedFileCount != 1 || ruleCalls.Load() != 1 {
		t.Fatalf("exact excluded-looking target: files=%d ruleCalls=%d", result.LintedFileCount, ruleCalls.Load())
	}
}

func TestSyntaxErrorTargetIsCountedWithoutResolvingOrRunningRules(t *testing.T) {
	dir := t.TempDir()
	writeTestFiles(t, dir, map[string]string{"broken.ts": "const = ;"})
	target := norm(dir, "broken.ts")
	raw := gapProgram(t, dir, []string{target})
	var ruleCalls atomic.Int32
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         wrapTestPrograms(raw),
		TargetsByProgram: [][]string{{target}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			ruleCalls.Add(1)
			return noopRule()
		},
	})
	result, err := RunLinter(RunLinterOptions{LintPlan: plan, SingleThreaded: true})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if result.LintedFileCount != 1 || ruleCalls.Load() != 0 || len(result.ExecutedRules) != 0 {
		t.Fatalf(
			"syntax target result: files=%d ruleCalls=%d executed=%v",
			result.LintedFileCount,
			ruleCalls.Load(),
			result.ExecutedRules,
		)
	}
}

func TestRunLinterRejectsNilProgramBeforeLintSideEffects(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	var ruleCalls atomic.Int32
	planOpts := PrepareLintPlanOptions{
		Programs:         append(wrapTestPrograms(raw), nil),
		TargetsByProgram: [][]string{{paths["a.ts"]}, nil},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			ruleCalls.Add(1)
			return noopRule()
		},
	}
	if _, err := PrepareLintPlan(planOpts); !errors.Is(err, errNilProgram) {
		t.Fatalf("PrepareLintPlan nil Program error = %v", err)
	}
	if ruleCalls.Load() != 0 {
		t.Fatalf("PrepareLintPlan resolved rules before rejecting nil Program: calls=%d", ruleCalls.Load())
	}
	typeCheckOnly := RunLinterOptions{
		TypeCheckOnlyPrograms: wrapTestPrograms(raw),
		TypeCheck:             true,
	}
	typeCheckOnly.TypeCheckOnlyPrograms = append(typeCheckOnly.TypeCheckOnlyPrograms, nil)
	if _, err := RunLinter(typeCheckOnly); !errors.Is(err, errNilProgram) {
		t.Fatalf("type-check-only RunLinter nil Program error = %v", err)
	}
	noOp, err := RunLinter(RunLinterOptions{TypeCheckOnlyPrograms: []*lintprogram.Program{nil}})
	if err != nil || noOp.LintedFileCount != 0 || len(noOp.ExecutedRules) != 0 {
		t.Fatalf("no-op RunLinter result = %+v, error = %v", noOp, err)
	}

	invalid := &lintprogram.Program{}
	invalidOpts := PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{invalid},
		TargetsByProgram: [][]string{nil},
		GetRulesForFile:  func(*ast.SourceFile) []rule.ConfiguredRule { return noopRule() },
	}
	if _, err := PrepareLintPlan(invalidOpts); !errors.Is(err, errInvalidProgram) {
		t.Fatalf("PrepareLintPlan zero Program error = %v", err)
	}
	if _, err := RunLinter(RunLinterOptions{
		LintPlan: &LintPlan{programs: []programLintPlan{{program: invalid}}},
	}); !errors.Is(err, errInvalidProgram) {
		t.Fatalf("RunLinter zero Program error = %v", err)
	}
}

func TestPreparedLintPlanFreezesProgramTypeCapability(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	file := raw.GetSourceFile(paths["a.ts"])
	if file == nil {
		t.Fatal("fixture Program did not contain a.ts")
	}
	sourceOnly := mustSourceOnlyTestProgram(t, raw, []*ast.SourceFile{file})
	tests := []struct {
		name     string
		program  *lintprogram.Program
		wantRuns int32
	}{
		{name: "compiler-capable", program: lintprogram.NewFromCompiler(raw), wantRuns: 1},
		{name: "source-only", program: sourceOnly},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var runs atomic.Int32
			programs := testPrograms(test.program)
			plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
				Programs:         programs,
				TargetsByProgram: [][]string{{paths["a.ts"]}},
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{
						Name:             "type-aware-probe",
						RequiresTypeInfo: true,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							if ctx.TypeChecker == nil || ctx.Program() != test.program {
								t.Fatal("prepared rule received incoherent Program capabilities")
							}
							runs.Add(1)
							return nil
						},
					}}
				},
			})
			if _, err := RunLinter(RunLinterOptions{LintPlan: plan}); err != nil {
				t.Fatalf("RunLinter: %v", err)
			}
			if got := runs.Load(); got != test.wantRuns {
				t.Fatalf("type-aware runs = %d, want %d", got, test.wantRuns)
			}
		})
	}
}
