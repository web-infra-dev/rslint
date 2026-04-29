package linter

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestRunLinterInProgram_AllowFilesNil(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	lintedFiles := RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles < 2 {
		t.Errorf("Expected at least 2 files with nil allowFiles, got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_AllowFilesSingle(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
		"c.ts": "const c = 3;",
	})

	lintedFileNames := []string{}
	lintedFiles := RunLinterInProgram(program, []string{paths["a.ts"]}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file, got %d", lintedFiles)
	}
	for _, name := range lintedFileNames {
		if !strings.HasSuffix(name, "a.ts") {
			t.Errorf("Expected only a.ts, got %s", name)
		}
	}
}

func TestRunLinterInProgram_AllowFilesMultiple(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
		"c.ts": "const c = 3;",
	})

	lintedFileNames := []string{}
	lintedFiles := RunLinterInProgram(program, []string{paths["a.ts"], paths["c.ts"]}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 2 {
		t.Errorf("Expected 2 files, got %d", lintedFiles)
	}
	for _, name := range lintedFileNames {
		if strings.HasSuffix(name, "b.ts") {
			t.Errorf("b.ts should not be linted")
		}
	}
}

func TestRunLinterInProgram_AllowFilesEmpty(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	lintedFiles := RunLinterInProgram(program, []string{}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files with empty allowFiles, got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_AllowFilesNotInProgram(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	nonexistent := tspath.NormalizePath(filepath.Join(t.TempDir(), "nonexistent.ts"))
	lintedFiles := RunLinterInProgram(program, []string{nonexistent}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files for nonexistent, got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_AllowFilesNoRules(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	diagnostics := []rule.RuleDiagnostic{}
	lintedFiles := RunLinterInProgram(program, []string{paths["a.ts"]}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		false, func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file counted, got %d", lintedFiles)
	}
	if len(diagnostics) != 0 {
		t.Errorf("Expected 0 diagnostics, got %d", len(diagnostics))
	}
}

func TestRunLinterInProgram_AllowFilesPartialMatch(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	nonexistent := tspath.NormalizePath(filepath.Join(t.TempDir(), "nonexistent.ts"))
	lintedFileNames := []string{}
	lintedFiles := RunLinterInProgram(program, []string{paths["a.ts"], nonexistent}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file (partial match), got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_AllowFilesDuplicate(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	lintedFiles := RunLinterInProgram(program, []string{paths["a.ts"], paths["a.ts"]}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file (dedup), got %d", lintedFiles)
	}
}

func TestRunLinter_AllowFilesIntegration(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	lintedFileNames := []string{}
	result, err := runLinterPositional([]*compiler.Program{program}, true, []string{paths["b.ts"]}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if result.LintedFileCount != 1 {
		t.Errorf("Expected 1 file, got %d", result.LintedFileCount)
	}
}

func TestRunLinter_AllowFilesNilPassthrough(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	result, err := runLinterPositional([]*compiler.Program{program}, true, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if result.LintedFileCount < 2 {
		t.Errorf("Expected at least 2 files, got %d", result.LintedFileCount)
	}
}
