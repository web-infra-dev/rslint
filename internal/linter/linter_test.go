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

// noopRule returns a rule that reports on every identifier (for testing file filtering).
func noopRule() []ConfiguredRule {
	return []ConfiguredRule{
		{
			Name:     "test-rule",
			Severity: rule.SeverityWarning,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return rule.RuleListeners{
					ast.KindIdentifier: func(node *ast.Node) {
						ctx.ReportNode(node, rule.RuleMessage{Id: "test", Description: "test"})
					},
				}
			},
		},
	}
}

// createTestProgramWithFiles creates a TS program in a temp directory with the given files.
// Returns the program and a map of short filename -> normalized absolute path.
func createTestProgramWithFiles(t *testing.T, sourceFiles map[string]string) (*compiler.Program, map[string]string) {
	t.Helper()

	tmpDir := t.TempDir()

	includes := make([]string, 0, len(sourceFiles))
	normalizedPaths := make(map[string]string, len(sourceFiles))
	for name, content := range sourceFiles {
		filePath := filepath.Join(tmpDir, name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", name, err)
		}
		includes = append(includes, "./"+name)
		normalizedPaths[name] = tspath.NormalizePath(filePath)
	}

	includeJSON := `"` + strings.Join(includes, `","`) + `"`
	tsconfig := `{"include":[` + includeJSON + `]}`
	if err := os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(tsconfig), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig: %v", err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}

	return program, normalizedPaths
}

// TestRunLinterInProgram_AllowFilesNil verifies that nil allowFiles lints all project files.
func TestRunLinterInProgram_AllowFilesNil(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	lintedFiles := RunLinterInProgram(
		program,
		nil, // allowFiles = nil → lint all
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		func(d rule.RuleDiagnostic) {},
	)

	if lintedFiles < 2 {
		t.Errorf("Expected at least 2 files to be linted with nil allowFiles, got %d", lintedFiles)
	}
}

// TestRunLinterInProgram_AllowFilesSingle verifies that a single allowFiles entry limits linting.
func TestRunLinterInProgram_AllowFilesSingle(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
		"c.ts": "const c = 3;",
	})

	allowFiles := []string{paths["a.ts"]}
	lintedFileNames := []string{}

	lintedFiles := RunLinterInProgram(
		program,
		allowFiles,
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		func(d rule.RuleDiagnostic) {},
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file to be linted, got %d", lintedFiles)
	}
	for _, name := range lintedFileNames {
		if !strings.HasSuffix(name, "a.ts") {
			t.Errorf("Expected only a.ts to be linted, but got %s", name)
		}
	}
}

// TestRunLinterInProgram_AllowFilesMultiple verifies filtering to a subset of files.
func TestRunLinterInProgram_AllowFilesMultiple(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
		"c.ts": "const c = 3;",
	})

	// Allow a.ts and c.ts, skip b.ts
	allowFiles := []string{paths["a.ts"], paths["c.ts"]}
	lintedFileNames := []string{}

	lintedFiles := RunLinterInProgram(
		program,
		allowFiles,
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		func(d rule.RuleDiagnostic) {},
	)

	if lintedFiles != 2 {
		t.Errorf("Expected 2 files to be linted, got %d", lintedFiles)
	}
	for _, name := range lintedFileNames {
		if strings.HasSuffix(name, "b.ts") {
			t.Errorf("b.ts should not be linted, but was included")
		}
	}
}

// TestRunLinterInProgram_AllowFilesEmpty verifies that an empty (non-nil) slice lints nothing.
func TestRunLinterInProgram_AllowFilesEmpty(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	// Empty slice (not nil) means "allow nothing"
	lintedFiles := RunLinterInProgram(
		program,
		[]string{},
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		func(d rule.RuleDiagnostic) {},
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files to be linted with empty allowFiles, got %d", lintedFiles)
	}
}

// TestRunLinterInProgram_AllowFilesNotInProgram verifies behavior when specified files
// are not part of the TS program (e.g. not in tsconfig include).
func TestRunLinterInProgram_AllowFilesNotInProgram(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	// Use a path that doesn't exist in the program
	_ = paths
	nonexistent := tspath.NormalizePath(filepath.Join(t.TempDir(), "nonexistent.ts"))
	lintedFiles := RunLinterInProgram(
		program,
		[]string{nonexistent},
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		func(d rule.RuleDiagnostic) {},
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files to be linted for nonexistent file, got %d", lintedFiles)
	}
}

// TestRunLinterInProgram_AllowFilesNoRules verifies that a file with no matching config rules
// is still counted as linted but produces no diagnostics (matching ESLint behavior).
func TestRunLinterInProgram_AllowFilesNoRules(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	diagnostics := []rule.RuleDiagnostic{}
	lintedFiles := RunLinterInProgram(
		program,
		[]string{paths["a.ts"]},
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil }, // no rules
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
	)

	// File is counted as linted (lintedFileCount++ happens before rule check)
	if lintedFiles != 1 {
		t.Errorf("Expected 1 file counted as linted, got %d", lintedFiles)
	}
	if len(diagnostics) != 0 {
		t.Errorf("Expected 0 diagnostics when no rules match, got %d", len(diagnostics))
	}
}

// TestRunLinter_AllowFilesIntegration tests the multi-program RunLinter wrapper with allowFiles.
func TestRunLinter_AllowFilesIntegration(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	programs := []*compiler.Program{program}
	allowFiles := []string{paths["b.ts"]}

	lintedFileNames := []string{}
	lintedCount, err := RunLinter(
		programs,
		true, // singleThreaded for deterministic test
		allowFiles,
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		func(d rule.RuleDiagnostic) {},
	)

	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	if lintedCount != 1 {
		t.Errorf("Expected 1 file linted via RunLinter, got %d", lintedCount)
	}
	for _, name := range lintedFileNames {
		if !strings.HasSuffix(name, "b.ts") {
			t.Errorf("Expected only b.ts to be linted, got %s", name)
		}
	}
}

// TestRunLinter_AllowFilesNilPassthrough verifies RunLinter with nil allowFiles lints everything.
func TestRunLinter_AllowFilesNilPassthrough(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	lintedCount, err := RunLinter(
		[]*compiler.Program{program},
		true,
		nil, // lint all
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		func(d rule.RuleDiagnostic) {},
	)

	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	if lintedCount < 2 {
		t.Errorf("Expected at least 2 files linted with nil allowFiles, got %d", lintedCount)
	}
}

// TestRunLinterInProgram_AllowFilesPartialMatch verifies that when some allowFiles
// are in the program and some are not, only the matching files are linted.
func TestRunLinterInProgram_AllowFilesPartialMatch(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	// a.ts exists in program, nonexistent.ts does not
	nonexistent := tspath.NormalizePath(filepath.Join(t.TempDir(), "nonexistent.ts"))
	allowFiles := []string{paths["a.ts"], nonexistent}

	lintedFileNames := []string{}
	lintedFiles := RunLinterInProgram(
		program,
		allowFiles,
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		func(d rule.RuleDiagnostic) {},
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file to be linted (partial match), got %d", lintedFiles)
	}
	for _, name := range lintedFileNames {
		if !strings.HasSuffix(name, "a.ts") {
			t.Errorf("Expected only a.ts to be linted, got %s", name)
		}
	}
}

// TestRunLinterInProgram_AllowFilesDuplicate verifies that duplicate entries in
// allowFiles don't cause a file to be linted multiple times.
func TestRunLinterInProgram_AllowFilesDuplicate(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	// Same file twice
	allowFiles := []string{paths["a.ts"], paths["a.ts"]}

	diagnosticCount := 0
	lintedFiles := RunLinterInProgram(
		program,
		allowFiles,
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		func(d rule.RuleDiagnostic) { diagnosticCount++ },
	)

	// File should only be linted once (the loop iterates program files, not allowFiles)
	if lintedFiles != 1 {
		t.Errorf("Expected 1 file to be linted (dedup), got %d", lintedFiles)
	}
}
