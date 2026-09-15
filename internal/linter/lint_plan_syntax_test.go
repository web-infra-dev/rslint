package linter

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func syntaxDiagnosticTestProgram(t *testing.T) (*compiler.Program, string) {
	t.Helper()
	directory := t.TempDir()
	writeTestFiles(t, directory, map[string]string{
		"broken.ts": "const value = ;\n",
	})
	targetPath := norm(directory, "broken.ts")
	return gapProgram(t, directory, []string{targetPath}), targetPath
}

func preparedSyntaxDiagnostics(
	t *testing.T,
	programs []*lintprogram.Program,
	targets [][]string,
	programDiagnosticsIncluded bool,
) []rule.RuleDiagnostic {
	t.Helper()
	return preparedSyntaxPlan(t, programs, targets).SyntacticDiagnostics(programDiagnosticsIncluded)
}

func preparedSyntaxPlan(
	t *testing.T,
	programs []*lintprogram.Program,
	targets [][]string,
) *LintPlan {
	t.Helper()
	plan, err := PrepareLintPlan(PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: targets,
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func TestLintPlanSyntacticDiagnosticsRespectsProgramCoverage(t *testing.T) {
	compilerProgram, targetPath := syntaxDiagnosticTestProgram(t)
	compilerBacked := wrapTestPrograms(compilerProgram)
	coveredPlan := preparedSyntaxPlan(t, compilerBacked, [][]string{{targetPath}})
	if !coveredPlan.HasSyntacticDiagnostics() {
		t.Fatal("target syntax state was lost when program diagnostics own emission")
	}

	if diagnostics := coveredPlan.SyntacticDiagnostics(false); len(diagnostics) == 0 {
		t.Fatal("compiler-backed target lost its syntax diagnostic")
	}
	if diagnostics := coveredPlan.SyntacticDiagnostics(true); len(diagnostics) != 0 {
		t.Errorf("compiler-backed diagnostics were duplicated by the lint projection: %v", diagnostics)
	}

	sourceFile := compilerProgram.GetSourceFile(targetPath)
	if sourceFile == nil {
		t.Fatalf("compiler Program does not contain %q", targetPath)
	}
	sourceOnly, err := lintprogram.NewFromBoundSources(compilerProgram, []*ast.SourceFile{sourceFile})
	if err != nil {
		t.Fatalf("create source-only Program: %v", err)
	}
	if diagnostics := preparedSyntaxDiagnostics(t,
		[]*lintprogram.Program{sourceOnly},
		[][]string{{targetPath}},
		true,
	); len(diagnostics) == 0 {
		t.Fatal("source-only target must provide syntax diagnostics when no program-wide phase can")
	}

	mixedPlan := preparedSyntaxPlan(t,
		[]*lintprogram.Program{compilerBacked[0], sourceOnly},
		[][]string{{targetPath}, {targetPath}},
	)
	if diagnostics := mixedPlan.SyntacticDiagnostics(true); len(diagnostics) != 1 {
		t.Fatalf("mixed-capability shared syntax diagnostics = %d, want 1: %v", len(diagnostics), diagnostics)
	}
}

func TestLintPlanSyntacticDiagnosticsDeduplicatesSharedTargets(t *testing.T) {
	compilerProgram, targetPath := syntaxDiagnosticTestProgram(t)
	programs := wrapTestPrograms(compilerProgram, compilerProgram)
	diagnostics := preparedSyntaxDiagnostics(t,
		programs,
		[][]string{{targetPath}, {targetPath}},
		false,
	)
	if len(diagnostics) != 1 {
		t.Fatalf("shared syntax diagnostics = %d, want 1: %v", len(diagnostics), diagnostics)
	}
}

func TestLintPlanPublishesSyntacticStateAfterParallelPreparation(t *testing.T) {
	directory := t.TempDir()
	writeTestFiles(t, directory, map[string]string{
		"broken.ts": "const value = ;\n",
		"valid.ts":  "const value = 1;\n",
	})
	brokenPath := norm(directory, "broken.ts")
	validPath := norm(directory, "valid.ts")
	programs := wrapTestPrograms(gapProgram(t, directory, []string{brokenPath, validPath}))
	plan, err := PrepareLintPlan(PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{brokenPath, validPath}},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.HasSyntacticDiagnostics() {
		t.Fatal("parallel preparation did not publish target syntax state")
	}
	diagnostics := plan.SyntacticDiagnostics(false)
	if len(diagnostics) != 1 || diagnostics[0].FilePath != brokenPath {
		t.Fatalf("parallel syntax diagnostics = %+v", diagnostics)
	}
}

func TestLintPlanKeepsCleanSyntacticProjectionSparse(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	directory := t.TempDir()
	writeTestFiles(t, directory, map[string]string{
		"first.ts":  "const first = 1;\n",
		"second.ts": "const second = 2;\n",
	})
	firstPath := norm(directory, "first.ts")
	secondPath := norm(directory, "second.ts")
	plan, err := PrepareLintPlan(PrepareLintPlanOptions{
		Programs:         wrapTestPrograms(gapProgram(t, directory, []string{firstPath, secondPath})),
		TargetsByProgram: [][]string{{firstPath, secondPath}},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.syntacticDiagnosticGroups != nil || plan.HasSyntacticDiagnostics() {
		t.Fatalf("clean syntax projection = %+v, want nil", plan.syntacticDiagnosticGroups)
	}
}

func TestLintPlanParallelSyntacticDiagnosticsPreserveTargetOrder(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	directory := t.TempDir()
	writeTestFiles(t, directory, map[string]string{
		"early-valid.ts":  "const early = 1;\n",
		"early-broken.ts": "const early = ;\n",
		"late-broken.ts":  "const late = ;\n",
		"late-valid.ts":   "const late = 1;\n",
	})
	earlyValid := norm(directory, "early-valid.ts")
	earlyBroken := norm(directory, "early-broken.ts")
	lateBroken := norm(directory, "late-broken.ts")
	lateValid := norm(directory, "late-valid.ts")
	programs := []*lintprogram.Program{
		wrapTestPrograms(gapProgram(t, directory, []string{earlyValid, earlyBroken}))[0],
		wrapTestPrograms(gapProgram(t, directory, []string{lateBroken, lateValid}))[0],
	}

	earlyBlocked := make(chan struct{})
	laterChunkFinished := make(chan struct{})
	releaseEarly := make(chan struct{})
	defer func() {
		select {
		case <-releaseEarly:
		default:
			close(releaseEarly)
		}
	}()
	type preparationResult struct {
		plan *LintPlan
		err  error
	}
	resultCh := make(chan preparationResult, 1)
	go func() {
		plan, err := PrepareLintPlan(PrepareLintPlanOptions{
			Programs:         programs,
			TargetsByProgram: [][]string{{earlyValid, earlyBroken}, {lateBroken, lateValid}},
			GetRulesForFile: func(file *ast.SourceFile) []rule.ConfiguredRule {
				switch file.FileName() {
				case earlyValid:
					close(earlyBlocked)
					<-releaseEarly
				case lateValid:
					close(laterChunkFinished)
				}
				return nil
			},
		})
		resultCh <- preparationResult{plan: plan, err: err}
	}()

	waitFor := func(name string, signal <-chan struct{}) {
		t.Helper()
		select {
		case <-signal:
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for %s", name)
		}
	}
	waitFor("early chunk to block", earlyBlocked)
	waitFor("later chunk to finish", laterChunkFinished)
	close(releaseEarly)

	var result preparationResult
	select {
	case result = <-resultCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for lint plan preparation")
	}
	if result.err != nil {
		t.Fatal(result.err)
	}
	diagnostics := result.plan.SyntacticDiagnostics(false)
	if len(diagnostics) != 2 ||
		diagnostics[0].FilePath != earlyBroken ||
		diagnostics[1].FilePath != lateBroken {
		t.Fatalf("parallel syntax diagnostic order = %+v, want %q then %q", diagnostics, earlyBroken, lateBroken)
	}
}

func TestCollectFileSyntacticDiagnosticsProjectsTypeScriptFields(t *testing.T) {
	compilerProgram, targetPath := syntaxDiagnosticTestProgram(t)
	sourceFile := compilerProgram.GetSourceFile(targetPath)
	if sourceFile == nil {
		t.Fatalf("compiler Program does not contain %q", targetPath)
	}
	program := lintprogram.NewFromCompiler(compilerProgram)
	raw := program.SyntacticDiagnostics(context.Background(), sourceFile)
	diagnostics := CollectFileSyntacticDiagnostics(context.Background(), program, sourceFile)
	if len(raw) == 0 || len(diagnostics) != len(raw) {
		t.Fatalf("projected diagnostics = %d, raw diagnostics = %d", len(diagnostics), len(raw))
	}

	got := diagnostics[0]
	want := raw[0]
	if got.SourceFile != sourceFile || got.FilePath != targetPath {
		t.Fatalf("source identity = (%T, %q), want source file and %q", got.SourceFile, got.FilePath, targetPath)
	}
	if got.Range.Pos() != want.Loc().Pos() || got.Range.End() != want.Loc().End() {
		t.Fatalf("range = [%d,%d), want [%d,%d)", got.Range.Pos(), got.Range.End(), want.Loc().Pos(), want.Loc().End())
	}
	if !strings.HasPrefix(got.RuleName, "TypeScript(TS") || got.Message.Description != want.String() {
		t.Fatalf("identity/message = (%q, %q), want TypeScript code and %q", got.RuleName, got.Message.Description, want.String())
	}
	if got.Severity != rule.SeverityError || got.Origin != rule.DiagnosticOriginTypeScript || !got.PreFormatted {
		t.Fatalf("severity/origin/preformatted = (%v, %v, %v)", got.Severity, got.Origin, got.PreFormatted)
	}
}
