package linter

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestRunLinterInProgram_AllowDirsBasic(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
		"lib/b.ts": "const b = 2;",
	})

	srcDir := tmpDirPath(t, paths, "src/a.ts")
	lintedFileNames := []string{}

	lintedFiles := RunLinterInProgram(program, nil, []string{srcDir}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file under src/, got %d", lintedFiles)
	}
	for _, name := range lintedFileNames {
		if !strings.Contains(name, "/src/") {
			t.Errorf("Expected only src/ files, got %s", name)
		}
	}
}

func TestRunLinterInProgram_AllowDirsNoFalsePrefix(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts":       "const a = 1;",
		"src-other/b.ts": "const b = 2;",
	})

	srcDir := tmpDirPath(t, paths, "src/a.ts")
	lintedFileNames := []string{}

	lintedFiles := RunLinterInProgram(program, nil, []string{srcDir}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file (only src/a.ts), got %d", lintedFiles)
	}
	for _, name := range lintedFileNames {
		if strings.Contains(name, "src-other") {
			t.Errorf("src-other/ should not match src/ allowDir, got %s", name)
		}
	}
}

func TestRunLinterInProgram_AllowDirsAndFilesOR(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
		"lib/b.ts": "const b = 2;",
		"other.ts": "const c = 3;",
	})

	srcDir := tmpDirPath(t, paths, "src/a.ts")
	lintedFileNames := []string{}

	lintedFiles := RunLinterInProgram(program,
		[]string{paths["lib/b.ts"]}, // allowFiles
		[]string{srcDir},            // allowDirs
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 2 {
		t.Errorf("Expected 2 files (OR logic), got %d", lintedFiles)
	}
	for _, name := range lintedFileNames {
		if strings.HasSuffix(name, "other.ts") {
			t.Errorf("other.ts should not be linted")
		}
	}
}

func TestRunLinterInProgram_AllowDirsEmpty(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	lintedFiles := RunLinterInProgram(program, nil, []string{}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files with empty allowDirs, got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_BothNilLintsAll(t *testing.T) {
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
		t.Errorf("Expected at least 2 files with both nil, got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_MultipleAllowDirs(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts":   "const a = 1;",
		"lib/b.ts":   "const b = 2;",
		"other/c.ts": "const c = 3;",
	})

	srcDir := tmpDirPath(t, paths, "src/a.ts")
	libDir := tmpDirPath(t, paths, "lib/b.ts")
	lintedFileNames := []string{}

	lintedFiles := RunLinterInProgram(program, nil, []string{srcDir, libDir}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 2 {
		t.Errorf("Expected 2 files (src + lib), got %d", lintedFiles)
	}
	for _, name := range lintedFileNames {
		if strings.Contains(name, "/other/") {
			t.Errorf("other/ should not be linted, got %s", name)
		}
	}
}

func TestRunLinterInProgram_AllowDirsWithEmptyAllowFiles(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
		"lib/b.ts": "const b = 2;",
	})

	srcDir := tmpDirPath(t, paths, "src/a.ts")

	lintedFiles := RunLinterInProgram(program, []string{}, []string{srcDir}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file (src/a.ts via allowDirs), got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_AllowDirsNoMatchInProgram(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
	})

	lintedFiles := RunLinterInProgram(program, nil, []string{"/nonexistent/dir"}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files for non-matching allowDirs, got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_NestedAllowDirs(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts":            "const a = 1;",
		"src/components/b.ts": "const b = 2;",
		"src/utils/c.ts":      "const c = 3;",
	})

	componentsDir := tmpDirPath(t, paths, "src/components/b.ts")
	lintedFileNames := []string{}

	lintedFiles := RunLinterInProgram(program, nil, []string{componentsDir}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFileNames = append(lintedFileNames, sf.FileName())
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file (only src/components/b.ts), got %d", lintedFiles)
	}
	for _, name := range lintedFileNames {
		if !strings.Contains(name, "/components/") {
			t.Errorf("Expected only components/ files, got %s", name)
		}
	}
}

func TestRunLinterInProgram_AllowDirsEmptyString(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	// Empty string as allowDir should not match anything
	lintedFiles := RunLinterInProgram(program, nil, []string{""}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files with empty string allowDir, got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_AllowDirsTrailingSlash(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
		"lib/b.ts": "const b = 2;",
	})

	// Trailing slash should still work
	srcDir := tmpDirPath(t, paths, "src/a.ts") + "/"
	lintedFiles := RunLinterInProgram(program, nil, []string{srcDir}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file with trailing slash allowDir, got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_AllowDirsSameAsFilePath(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
	})

	// Using the exact file path as allowDir should NOT match
	// (a file is not "inside" itself)
	lintedFiles := RunLinterInProgram(program, nil, []string{paths["src/a.ts"]}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files when allowDir equals file path, got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_AllowDirsAndFilesOverlap(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
	})

	srcDir := tmpDirPath(t, paths, "src/a.ts")
	diagnosticCount := 0

	// File matches both allowFiles AND allowDirs — should still be linted exactly once
	lintedFiles := RunLinterInProgram(program,
		[]string{paths["src/a.ts"]},
		[]string{srcDir},
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) { diagnosticCount++ }, nil,
		nil,
	)

	if lintedFiles != 1 {
		t.Errorf("Expected 1 file (overlap should not cause double lint), got %d", lintedFiles)
	}
}

// --- isDirAllowed direct unit tests ---

func TestIsDirAllowed_BasicMatch(t *testing.T) {
	if !isDirAllowed("/project/src/a.ts", []string{"/project/src"}) {
		t.Error("Expected /project/src/a.ts to be allowed under /project/src")
	}
}

func TestIsDirAllowed_NoMatch(t *testing.T) {
	if isDirAllowed("/project/lib/b.ts", []string{"/project/src"}) {
		t.Error("Expected /project/lib/b.ts NOT to be allowed under /project/src")
	}
}

func TestIsDirAllowed_NoPrefixFalsePositive(t *testing.T) {
	if isDirAllowed("/project/src-other/a.ts", []string{"/project/src"}) {
		t.Error("Expected src-other NOT to match src")
	}
}

func TestIsDirAllowed_EmptyString(t *testing.T) {
	if isDirAllowed("/project/a.ts", []string{""}) {
		t.Error("Expected empty string dir NOT to match")
	}
}

func TestIsDirAllowed_EmptySlice(t *testing.T) {
	if isDirAllowed("/project/a.ts", []string{}) {
		t.Error("Expected empty slice NOT to match")
	}
}

func TestIsDirAllowed_NilSlice(t *testing.T) {
	if isDirAllowed("/project/a.ts", nil) {
		t.Error("Expected nil slice NOT to match")
	}
}

func TestIsDirAllowed_TrailingSlash(t *testing.T) {
	if !isDirAllowed("/project/src/a.ts", []string{"/project/src/"}) {
		t.Error("Expected trailing slash to still match")
	}
}

func TestIsDirAllowed_ExactPathNotMatch(t *testing.T) {
	// A file path equal to the dir path is not "inside" the directory
	if isDirAllowed("/project/src", []string{"/project/src"}) {
		t.Error("Expected exact dir path NOT to match itself")
	}
}

func TestIsDirAllowed_MultipleMatches(t *testing.T) {
	if !isDirAllowed("/project/lib/b.ts", []string{"/project/src", "/project/lib"}) {
		t.Error("Expected /project/lib/b.ts to match second allowDir")
	}
}

func TestIsDirAllowed_RootDir(t *testing.T) {
	if !isDirAllowed("/any/file.ts", []string{"/"}) {
		t.Error("Expected root dir to match all absolute paths")
	}
}

func TestRunLinter_AllowDirsIntegration(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
		"lib/b.ts": "const b = 2;",
	})

	srcDir := tmpDirPath(t, paths, "src/a.ts")
	result, err := runLinterPositional([]*compiler.Program{program}, true, nil, []string{srcDir}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if result.LintedFileCount != 1 {
		t.Errorf("Expected 1 file under src/, got %d", result.LintedFileCount)
	}
}

func TestRunLinter_MultiplePrograms(t *testing.T) {
	// Two separate programs (simulating multi-config with different tsconfigs)
	programA, pathsA := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	programB, pathsB := createTestProgramWithFiles(t, map[string]string{
		"b.ts": "const b = 2;",
	})

	lintedFileNames := []string{}
	result, err := runLinterPositional(
		[]*compiler.Program{programA, programB},
		true,
		[]string{pathsA["a.ts"], pathsB["b.ts"]},
		nil,
		utils.ExcludePaths,
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
	if result.LintedFileCount != 2 {
		t.Errorf("Expected 2 files across 2 programs, got %d", result.LintedFileCount)
	}
}

func TestRunLinter_MultipleProgramsWithAllowDirs(t *testing.T) {
	programA, pathsA := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
		"lib/x.ts": "const x = 1;",
	})
	programB, pathsB := createTestProgramWithFiles(t, map[string]string{
		"src/b.ts": "const b = 2;",
		"lib/y.ts": "const y = 2;",
	})

	srcDirA := tmpDirPath(t, pathsA, "src/a.ts")
	srcDirB := tmpDirPath(t, pathsB, "src/b.ts")

	result, err := runLinterPositional(
		[]*compiler.Program{programA, programB},
		true,
		nil,
		[]string{srcDirA, srcDirB},
		utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	// Only src/ files from each program
	if result.LintedFileCount != 2 {
		t.Errorf("Expected 2 files (src from each program), got %d", result.LintedFileCount)
	}
}

func TestRunLinterInProgram_SkipTakesPriorityOverAllowDirs(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
	})

	srcDir := tmpDirPath(t, paths, "src/a.ts")

	// allowDirs would include src/a.ts, but skipFiles should take priority
	lintedFiles := RunLinterInProgram(program, nil, []string{srcDir},
		[]string{"src"}, // skip pattern matching "src" in path
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files (skip should override allowDirs), got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_SkipTakesPriorityOverAllowFiles(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const a = 1;",
	})

	lintedFiles := RunLinterInProgram(program, []string{paths["src/a.ts"]}, nil,
		[]string{"src"}, // skip pattern matching "src" in path
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files (skip should override allowFiles), got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_BothEmptyNonNil(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	// Both non-nil but empty → filter is active, nothing passes
	lintedFiles := RunLinterInProgram(program, []string{}, []string{}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if lintedFiles != 0 {
		t.Errorf("Expected 0 files with both empty non-nil, got %d", lintedFiles)
	}
}

func TestRunLinterInProgram_EmptyNonNilRules(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	diagnosticCount := 0
	lintedFiles := RunLinterInProgram(program, []string{paths["a.ts"]}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return []ConfiguredRule{} }, // empty non-nil
		false, func(d rule.RuleDiagnostic) { diagnosticCount++ }, nil,
		nil,
	)

	// Empty rules slice: len(rules) == 0 → skip (same as nil)
	if lintedFiles != 1 {
		t.Errorf("Expected 1 file counted, got %d", lintedFiles)
	}
	if diagnosticCount != 0 {
		t.Errorf("Expected 0 diagnostics with empty rules, got %d", diagnosticCount)
	}
}
