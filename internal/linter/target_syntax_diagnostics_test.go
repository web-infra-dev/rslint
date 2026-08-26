package linter

import (
	"context"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
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

func TestCollectTargetSyntacticDiagnosticsRespectsProgramCoverage(t *testing.T) {
	compilerProgram, targetPath := syntaxDiagnosticTestProgram(t)
	compilerBacked := wrapTestPrograms(compilerProgram)

	if diagnostics := CollectTargetSyntacticDiagnostics(
		compilerBacked,
		[][]string{{targetPath}},
		false,
	); len(diagnostics) == 0 {
		t.Fatal("compiler-backed target lost its syntax diagnostic")
	}
	if diagnostics := CollectTargetSyntacticDiagnostics(
		compilerBacked,
		[][]string{{targetPath}},
		true,
	); len(diagnostics) != 0 {
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
	if diagnostics := CollectTargetSyntacticDiagnostics(
		[]*lintprogram.Program{sourceOnly},
		[][]string{{targetPath}},
		true,
	); len(diagnostics) == 0 {
		t.Fatal("source-only target must provide syntax diagnostics when no program-wide phase can")
	}
}

func TestCollectTargetSyntacticDiagnosticsDeduplicatesSharedTargets(t *testing.T) {
	compilerProgram, targetPath := syntaxDiagnosticTestProgram(t)
	programs := wrapTestPrograms(compilerProgram, compilerProgram)
	diagnostics := CollectTargetSyntacticDiagnostics(
		programs,
		[][]string{{targetPath}, {targetPath}},
		false,
	)
	if len(diagnostics) != 1 {
		t.Fatalf("shared syntax diagnostics = %d, want 1: %v", len(diagnostics), diagnostics)
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
