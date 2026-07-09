package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func writeProgramTestFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
}

func collectProgramTypeDiagnostics(
	t *testing.T,
	programs []*compiler.Program,
	skip []bool,
	typeInfoFiles map[string]struct{},
) []rule.RuleDiagnostic {
	t.Helper()

	var diags []rule.RuleDiagnostic
	_, err := linter.RunLinter(linter.RunLinterOptions{
		Programs:              programs,
		SingleThreaded:        true,
		TypeCheck:             true,
		SkipTypeCheckPrograms: skip,
		TypeInfoFiles:         typeInfoFiles,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			diags = append(diags, d)
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	return diags
}

func containsTSDiagnostic(diags []rule.RuleDiagnostic, code string) bool {
	needle := "TypeScript(" + code + ")"
	for _, d := range diags {
		if d.RuleName == needle {
			return true
		}
	}
	return false
}

func TestTypeCheck_SkipsNoTsconfigTargetFallbackProgram(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"bad.ts": `const bad: number = "oops";
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	programs, exitCode := createProgramsForConfig(
		dir,
		rslintconfig.RslintConfig{{Files: []string{"**/*.ts"}}},
		true,
		fs,
		nil,
		utils.NewParseCache(),
	)
	if exitCode != 0 {
		t.Fatalf("createProgramsForConfig exit code = %d", exitCode)
	}
	if len(programs) != 0 {
		t.Fatalf("expected no tsconfig-backed programs, got %d", len(programs))
	}

	programs, typeInfoFiles, gapFiles, _, _, _ := buildProgramsWithLintTargets(
		programs,
		nil,
		rslintconfig.RslintConfig{{Files: []string{"**/*.ts"}}},
		dir,
		nil,
		nil,
		fs,
		nil,
		nil,
		utils.NewParseCache(),
		true,
	)
	if len(programs) != 1 {
		t.Fatalf("expected one files-driven fallback program, got %d", len(programs))
	}
	if len(gapFiles) == 0 {
		t.Fatalf("expected files-driven fallback roots, got %v", gapFiles)
	}
	if typeInfoFiles == nil || len(typeInfoFiles) != 0 {
		t.Fatalf("expected empty type-info set for no-tsconfig fallback, got %v", typeInfoFiles)
	}
	if got := programs[0].Options().ConfigFilePath; got != "" {
		t.Fatalf("expected files-driven fallback program to have no ConfigFilePath, got %q", got)
	}
	if !programs[0].Options().NoLib.IsTrue() || !programs[0].Options().NoResolve.IsTrue() {
		t.Fatalf("expected files-driven fallback to stay AST-only, got options %+v", programs[0].Options())
	}
	skip := buildTypeCheckSkipMask(programs)
	if len(skip) != 1 || !skip[0] {
		t.Fatalf("expected no-tsconfig fallback program to be skipped for type-check, got %v", skip)
	}

	if diags := collectProgramTypeDiagnostics(t, programs, skip, typeInfoFiles); containsTSDiagnostic(diags, "TS2322") {
		t.Fatalf("did not expect semantic diagnostics from a no-tsconfig fallback program: %+v", diags)
	}
}

func TestTypeCheck_TsconfigBackedProgramReportsCoveredDeclarationErrors(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"tsconfig.json": `{
  "compilerOptions": { "skipLibCheck": false },
  "include": ["rslint.config.ts"]
}
`,
		"rslint.config.ts": `import type { Bad } from './bad';
export const value: Bad | null = null;
`,
		"bad.d.ts": `export type Bad = MissingGlobalType;
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	programs, exitCode := createProgramsForConfig(
		dir,
		rslintconfig.RslintConfig{{
			Files: []string{"**/*.ts"},
			LanguageOptions: &rslintconfig.LanguageOptions{
				ParserOptions: &rslintconfig.ParserOptions{
					Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
				},
			},
		}},
		true,
		fs,
		nil,
		utils.NewParseCache(),
	)
	if exitCode != 0 {
		t.Fatalf("createProgramsForConfig exit code = %d", exitCode)
	}
	if len(programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program, got %d", len(programs))
	}
	if got := programs[0].Options().ConfigFilePath; got == "" {
		t.Fatal("expected tsconfig-backed program to carry ConfigFilePath")
	}
	skip := buildTypeCheckSkipMask(programs)
	if skip != nil {
		t.Fatalf("expected tsconfig-backed program to participate in type-check, got %v", skip)
	}

	diags := collectProgramTypeDiagnostics(t, programs, skip, nil)
	if !containsTSDiagnostic(diags, "TS2304") {
		var rendered []string
		for _, d := range diags {
			rendered = append(rendered, d.RuleName+": "+d.Message.Description)
		}
		t.Fatalf("expected TS2304 from the tsconfig-covered declaration graph, got:\n%s", strings.Join(rendered, "\n"))
	}
}

func TestTypeCheck_SkipsGapFallbackPrograms(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"tsconfig.json": `{
  "compilerOptions": { "skipLibCheck": false },
  "include": ["src/in-project.ts"]
}
`,
		"src/in-project.ts": `export const ok = 1;
`,
		"gap.ts": `import type { Bad } from './bad';
export const value: Bad | null = null;
`,
		"bad.d.ts": `export type Bad = MissingGlobalType;
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
			},
		},
	}}
	programs, exitCode := createProgramsForConfig(
		dir,
		cfg,
		true,
		fs,
		nil,
		parseCache,
	)
	if exitCode != 0 {
		t.Fatalf("createProgramsForConfig exit code = %d", exitCode)
	}
	if len(programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program before gap fallback, got %d", len(programs))
	}
	if skip := buildTypeCheckSkipMask(programs); skip != nil {
		t.Fatalf("expected tsconfig-backed program to participate in type-check, got %v", skip)
	}

	programs, typeInfoFiles, gapFiles, _, _, _ := buildProgramsWithLintTargets(
		programs,
		nil,
		cfg,
		dir,
		nil,
		nil,
		fs,
		nil,
		nil,
		parseCache,
		true,
	)
	if len(programs) != 2 {
		t.Fatalf("expected the gap fallback program to be appended, got %d programs", len(programs))
	}
	gapPath := filepath.ToSlash(filepath.Join(dir, "gap.ts"))
	if !slices.Contains(gapFiles, gapPath) {
		t.Fatalf("expected gap.ts to be discovered as a gap file, got %v", gapFiles)
	}
	if got := programs[0].Options().ConfigFilePath; got == "" {
		t.Fatal("expected original tsconfig-backed program to carry ConfigFilePath")
	}
	if got := programs[1].Options().ConfigFilePath; got != "" {
		t.Fatalf("expected gap fallback program to have no ConfigFilePath, got %q", got)
	}
	if !programs[1].Options().NoLib.IsTrue() || !programs[1].Options().NoResolve.IsTrue() {
		t.Fatalf("expected gap fallback to stay AST-only, got options %+v", programs[1].Options())
	}
	skip := buildTypeCheckSkipMask(programs)
	if len(skip) != 2 || skip[0] || !skip[1] {
		t.Fatalf("expected only the gap fallback program to be skipped, got %v", skip)
	}

	if diags := collectProgramTypeDiagnostics(t, programs, skip, typeInfoFiles); containsTSDiagnostic(diags, "TS2304") {
		t.Fatalf("did not expect declaration diagnostics from a gap fallback program: %+v", diags)
	}
}

func TestBuildProgramsWithLintTargets_BindsImportedNonRootFile(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"main.ts":       "import { value } from './lib';\nconsole.log(value);\n",
		"lib.ts":        "export const value = 1;\n",
		"tsconfig.json": `{"files": ["main.ts"], "compilerOptions": {"module": "ESNext"}}`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
			},
		},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}
	programs, exitCode := createProgramsForConfig(dir, cfg, true, fs, nil, parseCache)
	if exitCode != 0 {
		t.Fatalf("createProgramsForConfig exit code = %d", exitCode)
	}
	if len(programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program, got %d", len(programs))
	}

	libPath := tspath.NormalizePath(filepath.Join(dir, "lib.ts"))
	programs, typeInfoFiles, gapFiles, targetFiles, targetsByProgram, _ := buildProgramsWithLintTargets(
		programs,
		nil,
		cfg,
		dir,
		nil,
		nil,
		fs,
		[]string{libPath},
		nil,
		parseCache,
		true,
	)
	if len(programs) != 1 {
		t.Fatalf("imported non-root target should reuse existing Program, got %d programs", len(programs))
	}
	if len(gapFiles) != 0 {
		t.Fatalf("imported non-root target should not become a gap file, got %v", gapFiles)
	}
	if typeInfoFiles != nil {
		t.Fatalf("no fallback appended, so typeInfoFiles should stay nil, got %v", typeInfoFiles)
	}
	if len(targetFiles) != 1 || targetFiles[0] != libPath {
		t.Fatalf("expected lib.ts as the only target, got %v", targetFiles)
	}
	if len(targetsByProgram) != 1 || len(targetsByProgram[0]) != 1 || targetsByProgram[0][0] != libPath {
		t.Fatalf("expected lib.ts bound to the tsconfig Program, got %v", targetsByProgram)
	}
}

func TestBuildProgramsWithLintTargets_BindsRealpathTargetToProgramSourceName(t *testing.T) {
	realDir := t.TempDir()
	linkDir := filepath.Join(filepath.Dir(realDir), filepath.Base(realDir)+"-link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	defer os.Remove(linkDir)

	writeProgramTestFiles(t, realDir, map[string]string{
		"src/a.ts":      "export const a = 1;\n",
		"tsconfig.json": `{"include": ["src/a.ts"]}`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
			},
		},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}

	linkDir = tspath.NormalizePath(linkDir)
	realTarget := tspath.NormalizePath(filepath.Join(realDir, "src/a.ts"))
	programs, exitCode := createProgramsForConfig(linkDir, cfg, true, fs, nil, parseCache)
	if exitCode != 0 {
		t.Fatalf("createProgramsForConfig exit code = %d", exitCode)
	}
	if len(programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program, got %d", len(programs))
	}

	var sourceName string
	for _, sf := range programs[0].GetSourceFiles() {
		if strings.HasSuffix(sf.FileName(), "/src/a.ts") {
			sourceName = sf.FileName()
			break
		}
	}
	if sourceName == "" {
		t.Fatal("expected program to include src/a.ts")
	}
	if sourceName == realTarget {
		t.Skip("compiler already canonicalized source file to realpath")
	}

	programs, typeInfoFiles, gapFiles, targetFiles, targetsByProgram, configPathBySourcePath := buildProgramsWithLintTargets(
		programs,
		nil,
		cfg,
		linkDir,
		nil,
		nil,
		fs,
		[]string{realTarget},
		nil,
		parseCache,
		true,
	)
	if len(programs) != 1 {
		t.Fatalf("realpath target should reuse existing Program, got %d programs", len(programs))
	}
	if len(gapFiles) != 0 {
		t.Fatalf("realpath target should not become a gap file, got %v", gapFiles)
	}
	if typeInfoFiles != nil {
		t.Fatalf("no fallback appended, so typeInfoFiles should stay nil, got %v", typeInfoFiles)
	}
	if len(targetFiles) != 1 || targetFiles[0] != realTarget {
		t.Fatalf("expected realpath target as the only discovered target, got %v", targetFiles)
	}
	if len(targetsByProgram) != 1 || len(targetsByProgram[0]) != 1 || targetsByProgram[0][0] != sourceName {
		t.Fatalf("expected realpath target to bind back to source name %q, got %v", sourceName, targetsByProgram)
	}
	if configPathBySourcePath[sourceName] != realTarget {
		t.Fatalf("expected source path %q to resolve config through target path %q, got %v", sourceName, realTarget, configPathBySourcePath)
	}
}
