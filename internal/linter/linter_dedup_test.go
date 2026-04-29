package linter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// --- helpers ---

func writeTestFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

func programFromTsconfig(t *testing.T, dir, tsconfig string) *compiler.Program {
	t.Helper()
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(dir, fs)
	prog, err := utils.CreateProgram(true, fs, dir, tsconfig, host)
	if err != nil {
		t.Fatalf("CreateProgram(%s): %v", tsconfig, err)
	}
	return prog
}

func gapProgram(t *testing.T, dir string, rootFiles []string) *compiler.Program {
	t.Helper()
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(dir, fs)
	prog, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Target: core.ScriptTargetESNext,
		Module: core.ModuleKindESNext,
	}, rootFiles, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptionsLenient: %v", err)
	}
	return prog
}

func norm(dir, rel string) string {
	return tspath.NormalizePath(filepath.Join(dir, filepath.FromSlash(rel)))
}

// programHasFile checks if a program's GetSourceFiles() includes a file.
func programHasFile(program *compiler.Program, fileName string) bool {
	for _, sf := range program.GetSourceFiles() {
		if sf.FileName() == fileName {
			return true
		}
	}
	return false
}

// requireProgramHasFile fatally fails if program's GetSourceFiles() doesn't contain fileName.
// This ensures the test actually exercises dedup by confirming the overlap exists.
func requireProgramHasFile(t *testing.T, program *compiler.Program, fileName string) {
	t.Helper()
	for _, sf := range program.GetSourceFiles() {
		if sf.FileName() == fileName {
			return
		}
	}
	t.Fatalf("precondition failed: program does not contain %s — overlap does not exist, test cannot verify dedup", filepath.Base(fileName))
}

// collectLintedFiles runs RunLinter and returns {fileName → lintCount}.
func collectLintedFiles(t *testing.T, programs []*compiler.Program) map[string]int {
	t.Helper()
	counts := make(map[string]int)
	_, err := runLinterPositional(
		programs, true, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			counts[sf.FileName()]++
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil, nil,
	)
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	return counts
}

// assertTotalLintCount checks that the exact set of expected files was linted, each exactly once.
func assertTotalLintCount(t *testing.T, counts map[string]int, expectedFiles []string) {
	t.Helper()
	for _, f := range expectedFiles {
		if counts[f] != 1 {
			t.Errorf("%s: linted %d times, want 1", filepath.Base(f), counts[f])
		}
	}
	// Check no unexpected files were linted
	expected := make(map[string]bool, len(expectedFiles))
	for _, f := range expectedFiles {
		expected[f] = true
	}
	for f, c := range counts {
		if !expected[f] && c > 0 {
			t.Errorf("unexpected file linted: %s (%d times)", filepath.Base(f), c)
		}
	}
}

// --- unit test: buildOwnedFileSet ---

func TestBuildOwnedFileSet_TsconfigProgram(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})
	owned := buildOwnedFileSet(program)
	if owned == nil {
		t.Fatal("expected non-nil owned set")
	}
	if _, ok := owned[paths["a.ts"]]; !ok {
		t.Error("a.ts should be owned")
	}
	if _, ok := owned[paths["b.ts"]]; !ok {
		t.Error("b.ts should be owned")
	}
	if len(owned) != 2 {
		t.Errorf("owned set size = %d, want 2 (exactly a.ts and b.ts)", len(owned))
	}
}

func TestBuildOwnedFileSet_ExcludesImportedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFiles(t, tmpDir, map[string]string{
		"app.ts":        "import { lib } from './lib'; export const x = lib;",
		"lib.ts":        "export const lib = 1;",
		"tsconfig.json": `{"include": ["./app.ts"]}`,
	})
	program := programFromTsconfig(t, tmpDir, "tsconfig.json")
	libPath := norm(tmpDir, "lib.ts")
	appPath := norm(tmpDir, "app.ts")

	// Hard precondition: lib.ts MUST be pulled in via import for the test to be meaningful
	requireProgramHasFile(t, program, libPath)

	owned := buildOwnedFileSet(program)
	if _, ok := owned[libPath]; ok {
		t.Error("lib.ts should NOT be owned (not in tsconfig include, only pulled in via import)")
	}
	if _, ok := owned[appPath]; !ok {
		t.Error("app.ts should be owned (in tsconfig include)")
	}
}

func TestBuildOwnedFileSet_GapProgram(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFiles(t, tmpDir, map[string]string{
		"gap.ts": "import { lib } from './lib'; export const x = lib;",
		"lib.ts": "export const lib = 1;",
	})
	gapPath := norm(tmpDir, "gap.ts")
	libPath := norm(tmpDir, "lib.ts")
	prog := gapProgram(t, tmpDir, []string{gapPath})

	// Hard precondition: lib.ts must be pulled into gap program via import
	requireProgramHasFile(t, prog, libPath)

	owned := buildOwnedFileSet(prog)
	if _, ok := owned[gapPath]; !ok {
		t.Error("gap.ts should be owned (root file of gap program)")
	}
	if _, ok := owned[libPath]; ok {
		t.Error("lib.ts should NOT be owned (pulled in via import, not a gap root file)")
	}
}

// --- integration: RunLinter cross-program dedup ---

// Two tsconfig programs, B imports A's file → A's file linted once by A, not by B.
func TestRunLinter_ImportedFileNotDuplicated(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFiles(t, tmpDir, map[string]string{
		"lib.ts":            "export const lib = 1;",
		"app.ts":            "import { lib } from './lib'; export const x = lib;",
		"tsconfig-lib.json": `{"include": ["./lib.ts"]}`,
		"tsconfig-app.json": `{"include": ["./app.ts"]}`,
	})

	programLib := programFromTsconfig(t, tmpDir, "tsconfig-lib.json")
	programApp := programFromTsconfig(t, tmpDir, "tsconfig-app.json")
	libPath := norm(tmpDir, "lib.ts")
	appPath := norm(tmpDir, "app.ts")

	// Hard precondition: app program MUST contain lib.ts (via import).
	// If this fails, the test cannot verify dedup.
	requireProgramHasFile(t, programApp, libPath)

	counts := collectLintedFiles(t, []*compiler.Program{programLib, programApp})
	assertTotalLintCount(t, counts, []string{libPath, appPath})
}

// Gap program imports a tsconfig file → tsconfig file linted once by tsconfig program.
func TestRunLinter_GapProgramDoesNotDuplicateTsconfigFiles(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFiles(t, tmpDir, map[string]string{
		"lib.ts":        "export const lib = 1;",
		"gap.ts":        "import { lib } from './lib'; export const x = lib;",
		"tsconfig.json": `{"include": ["./lib.ts"]}`,
	})

	programTs := programFromTsconfig(t, tmpDir, "tsconfig.json")
	gapPath := norm(tmpDir, "gap.ts")
	libPath := norm(tmpDir, "lib.ts")
	programGap := gapProgram(t, tmpDir, []string{gapPath})

	// Hard precondition: gap program MUST contain lib.ts (via import).
	requireProgramHasFile(t, programGap, libPath)

	counts := collectLintedFiles(t, []*compiler.Program{programTs, programGap})
	assertTotalLintCount(t, counts, []string{libPath, gapPath})
}

// Two programs with no overlap — all files linted exactly once.
func TestRunLinter_NoOverlapAllFilesLinted(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFiles(t, tmpDir, map[string]string{
		"a.ts":            "const a = 1;",
		"b.ts":            "const b = 2;",
		"tsconfig-a.json": `{"include": ["./a.ts"]}`,
		"tsconfig-b.json": `{"include": ["./b.ts"]}`,
	})

	progA := programFromTsconfig(t, tmpDir, "tsconfig-a.json")
	progB := programFromTsconfig(t, tmpDir, "tsconfig-b.json")

	counts := collectLintedFiles(t, []*compiler.Program{progA, progB})
	assertTotalLintCount(t, counts, []string{norm(tmpDir, "a.ts"), norm(tmpDir, "b.ts")})
}

// Single program — all root files linted.
func TestRunLinter_SingleProgramAllFilesLinted(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
		"c.ts": "const c = 3;",
	})

	counts := collectLintedFiles(t, []*compiler.Program{program})
	assertTotalLintCount(t, counts, []string{paths["a.ts"], paths["b.ts"], paths["c.ts"]})
}

// Diagnostics count matches single-program baseline (proves dedup doesn't double-report).
func TestRunLinter_DiagnosticsNotDuplicated(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFiles(t, tmpDir, map[string]string{
		"lib.ts":            "export const lib = 1;",
		"app.ts":            "import { lib } from './lib'; export const x = lib;",
		"tsconfig-lib.json": `{"include": ["./lib.ts"]}`,
		"tsconfig-app.json": `{"include": ["./app.ts"]}`,
	})

	programLib := programFromTsconfig(t, tmpDir, "tsconfig-lib.json")
	programApp := programFromTsconfig(t, tmpDir, "tsconfig-app.json")
	libPath := norm(tmpDir, "lib.ts")

	requireProgramHasFile(t, programApp, libPath)

	// Baseline: diagnostic count for lib.ts in single-program mode
	singleDiags := 0
	RunLinterInProgram(programLib, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {
			if d.SourceFile.FileName() == libPath {
				singleDiags++
			}
		}, nil, nil,
	)
	if singleDiags == 0 {
		t.Fatal("noopRule produced 0 diagnostics for lib.ts — test is broken")
	}

	// Multi-program mode
	multiDiags := 0
	runLinterPositional(
		[]*compiler.Program{programLib, programApp},
		true, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		false, func(d rule.RuleDiagnostic) {
			if d.SourceFile.FileName() == libPath {
				multiDiags++
			}
		}, nil, nil,
	)

	if multiDiags != singleDiags {
		t.Errorf("lib.ts: %d diagnostics in multi-program, want %d (single-program baseline)", multiDiags, singleDiags)
	}
}

// A → B → C chain, each owned by different programs. All linted once.
func TestRunLinter_TransitiveImportChain(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFiles(t, tmpDir, map[string]string{
		"c.ts":            "export const c = 1;",
		"b.ts":            "import { c } from './c'; export const b = c;",
		"a.ts":            "import { b } from './b'; export const a = b;",
		"tsconfig-c.json": `{"include": ["./c.ts"]}`,
		"tsconfig-b.json": `{"include": ["./b.ts"]}`,
		"tsconfig-a.json": `{"include": ["./a.ts"]}`,
	})

	progC := programFromTsconfig(t, tmpDir, "tsconfig-c.json")
	progB := programFromTsconfig(t, tmpDir, "tsconfig-b.json")
	progA := programFromTsconfig(t, tmpDir, "tsconfig-a.json")

	cPath := norm(tmpDir, "c.ts")
	bPath := norm(tmpDir, "b.ts")

	// Verify overlap exists: progA has b.ts+c.ts (transitive), progB has c.ts
	requireProgramHasFile(t, progA, bPath)
	requireProgramHasFile(t, progA, cPath) // transitive: a→b→c
	requireProgramHasFile(t, progB, cPath)

	counts := collectLintedFiles(t, []*compiler.Program{progC, progB, progA})
	assertTotalLintCount(t, counts, []string{norm(tmpDir, "a.ts"), norm(tmpDir, "b.ts"), norm(tmpDir, "c.ts")})
}

// Project references: plugin references core. Core files linted once.
// Whether the plugin program includes core's .ts source files depends on compiler
// behavior (may resolve to .d.ts instead). Either way, each file must be linted
// exactly once.
func TestRunLinter_ProjectReferences(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFiles(t, tmpDir, map[string]string{
		"core/src/lib.ts": "export const lib = 1;",
		"core/tsconfig.json": `{
			"compilerOptions": {"composite": true, "outDir": "./dist", "rootDir": "src"},
			"include": ["src"]
		}`,
		"plugin/src/index.ts": "import { lib } from '../../core/src/lib'; export const x = lib;",
		"plugin/tsconfig.json": `{
			"compilerOptions": {"outDir": "./dist", "rootDir": "src"},
			"include": ["src"],
			"references": [{"path": "../core"}]
		}`,
	})

	coreDir := filepath.Join(tmpDir, "core")
	pluginDir := filepath.Join(tmpDir, "plugin")
	progCore := programFromTsconfig(t, coreDir, "tsconfig.json")
	progPlugin := programFromTsconfig(t, pluginDir, "tsconfig.json")

	corePath := norm(coreDir, "src/lib.ts")
	pluginPath := norm(pluginDir, "src/index.ts")

	// Check whether real overlap exists (compiler may resolve to .d.ts instead).
	// Log the scenario for clarity, but verify correct behavior either way.
	pluginHasCoreSrc := programHasFile(progPlugin, corePath)
	if pluginHasCoreSrc {
		t.Log("plugin program includes core .ts source — verifying dedup prevents double lint")
	} else {
		t.Log("plugin program uses .d.ts for core — no overlap, verifying no files dropped")
	}

	counts := collectLintedFiles(t, []*compiler.Program{progCore, progPlugin})
	assertTotalLintCount(t, counts, []string{corePath, pluginPath})
}

// Gap program with files that import tsconfig files — gap files linted, tsconfig files not re-linted.
func TestRunLinter_GapProgramOnlyLintsOwnFiles(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFiles(t, tmpDir, map[string]string{
		"lib.ts":        "export const lib = 1;",
		"gap1.ts":       "import { lib } from './lib'; export const g1 = lib;",
		"gap2.ts":       "import { lib } from './lib'; export const g2 = lib;",
		"tsconfig.json": `{"include": ["./lib.ts"]}`,
	})

	progTs := programFromTsconfig(t, tmpDir, "tsconfig.json")
	gap1 := norm(tmpDir, "gap1.ts")
	gap2 := norm(tmpDir, "gap2.ts")
	libPath := norm(tmpDir, "lib.ts")
	progGap := gapProgram(t, tmpDir, []string{gap1, gap2})

	// Hard precondition: gap program must pull in lib.ts via import
	requireProgramHasFile(t, progGap, libPath)

	counts := collectLintedFiles(t, []*compiler.Program{progTs, progGap})
	assertTotalLintCount(t, counts, []string{libPath, gap1, gap2})
}

// LSP path: RunLinterInProgram direct call is NOT affected by ownedFiles filter.
// Imported file can still be linted when explicitly requested via allowFiles.
func TestRunLinterInProgram_DirectCallNotFiltered(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFiles(t, tmpDir, map[string]string{
		"app.ts":        "import { lib } from './lib'; export const x = lib;",
		"lib.ts":        "export const lib = 1;",
		"tsconfig.json": `{"include": ["./app.ts"]}`,
	})

	program := programFromTsconfig(t, tmpDir, "tsconfig.json")
	libPath := norm(tmpDir, "lib.ts")

	// lib.ts is NOT in tsconfig include, only pulled in via import
	requireProgramHasFile(t, program, libPath)

	owned := buildOwnedFileSet(program)
	if _, ok := owned[libPath]; ok {
		t.Fatal("precondition failed: lib.ts should NOT be in owned set")
	}

	// Direct RunLinterInProgram call (like LSP) — no ownedFiles filter applied
	lintedFiles := make(map[string]int)
	RunLinterInProgram(program, []string{libPath}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			lintedFiles[sf.FileName()]++
			return noopRule()
		},
		false, func(d rule.RuleDiagnostic) {}, nil, nil,
	)

	if lintedFiles[libPath] != 1 {
		t.Errorf("lib.ts: linted %d times via direct call, want 1 (should not be filtered)", lintedFiles[libPath])
	}
}
