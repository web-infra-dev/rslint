package linter

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
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
