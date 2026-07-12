package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
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

func TestRunLinter_TargetFilesEmptyDoesNotScanProgram(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	called := false
	result, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{nil},
		GetRulesForFile: func(sf *ast.SourceFile) []ConfiguredRule {
			called = true
			return noopRule()
		},
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if called {
		t.Fatal("GetRulesForFile should not be called for an empty target plan")
	}
	if result.LintedFileCount != 0 {
		t.Fatalf("expected zero linted files for an empty target plan, got %d", result.LintedFileCount)
	}

	targets := CollectLintTargets(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{nil},
		GetRulesForFile: func(sf *ast.SourceFile) []ConfiguredRule {
			return noopRule()
		},
	})
	if len(targets) != 0 {
		t.Fatalf("CollectLintTargets should also see zero files for an empty target plan, got %+v", targets)
	}
}

func TestRunLinter_TargetFilesCanSelectImportedNonRootFile(t *testing.T) {
	program, paths := createImportedNonRootProgram(t)
	target := paths["lib.ts"]

	var linted []string
	result, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{target}},
		GetRulesForFile: func(sf *ast.SourceFile) []ConfiguredRule {
			linted = append(linted, sf.FileName())
			return noopRule()
		},
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if result.LintedFileCount != 1 {
		t.Fatalf("expected exactly one imported non-root file to be linted, got %d", result.LintedFileCount)
	}
	if len(linted) != 1 || linted[0] != target {
		t.Fatalf("expected only lib.ts to be linted, got %v", linted)
	}

	targets := CollectLintTargets(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{target}},
		GetRulesForFile: func(sf *ast.SourceFile) []ConfiguredRule {
			return noopRule()
		},
	})
	if len(targets) != 1 || targets[0].File.FileName() != target || len(targets[0].Rules) == 0 {
		t.Fatalf("CollectLintTargets should mirror native exact targeting, got %+v", targets)
	}
}

func TestLintSingleFile_TargetsImportedNonRootFile(t *testing.T) {
	program, paths := createImportedNonRootProgram(t)
	target := paths["lib.ts"]

	var linted []string
	LintSingleFile(LintSingleFileOptions{
		Program: program,
		File:    target,
		GetRulesForFile: func(sf *ast.SourceFile) []ConfiguredRule {
			linted = append(linted, sf.FileName())
			return noopRule()
		},
	})

	if len(linted) != 1 || linted[0] != target {
		t.Fatalf("expected only imported lib.ts to be linted, got %v", linted)
	}
}

func createImportedNonRootProgram(t *testing.T) (*compiler.Program, map[string]string) {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"main.ts":       "import { value } from './lib';\nconsole.log(value);\n",
		"lib.ts":        "export const value = 1;\n",
		"tsconfig.json": `{"files": ["main.ts"], "compilerOptions": {"module": "ESNext"}}`,
	}
	paths := make(map[string]string, len(files))
	for name, content := range files {
		fullPath := filepath.Join(dir, name)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		paths[name] = tspath.NormalizePath(fullPath)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(dir, fs)
	program, err := utils.CreateProgram(true, fs, dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	return program, paths
}
