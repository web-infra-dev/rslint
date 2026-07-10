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
	"github.com/microsoft/typescript-go/shim/vfs"
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

func resolveAndBindTestTargets(
	t *testing.T,
	set lintProgramSet,
	cfg rslintconfig.RslintConfig,
	dir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	parseCache *utils.ParseCache,
) (lintTargetPlan, lintTargetBinding) {
	t.Helper()
	plan, err := resolveLintTargetPlan(nil, cfg, dir, nil, fsys, allowFiles, allowDirs, true)
	if err != nil {
		t.Fatalf("resolveLintTargetPlan: %v", err)
	}
	binding, err := bindLintTargetPlan(set, plan, dir, fsys, parseCache, true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	return plan, binding
}

func TestTypeCheck_SkipsNoTsconfigTargetFallbackProgram(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"bad.ts": `const bad: number = "oops";
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	programSet, err := createProgramSetForConfig(
		dir,
		rslintconfig.RslintConfig{{Files: []string{"**/*.ts"}}},
		true,
		fs,
		utils.NewParseCache(),
	)
	if err != nil {
		t.Fatalf("createProgramSetForConfig: %v", err)
	}
	if len(programSet.Programs) != 0 {
		t.Fatalf("expected no tsconfig-backed programs, got %d", len(programSet.Programs))
	}

	_, binding := resolveAndBindTestTargets(
		t,
		programSet,
		rslintconfig.RslintConfig{{Files: []string{"**/*.ts"}}},
		dir,
		fs,
		nil,
		nil,
		utils.NewParseCache(),
	)
	programs := binding.Programs
	typeInfoFiles := binding.TypeInfoFiles
	gapFiles := binding.GapFiles
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
	programSet, err := createProgramSetForConfig(
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
		utils.NewParseCache(),
	)
	if err != nil {
		t.Fatalf("createProgramSetForConfig: %v", err)
	}
	programs := programSet.Programs
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
	programSet, err := createProgramSetForConfig(
		dir,
		cfg,
		true,
		fs,
		parseCache,
	)
	if err != nil {
		t.Fatalf("createProgramSetForConfig: %v", err)
	}
	if len(programSet.Programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program before gap fallback, got %d", len(programSet.Programs))
	}
	if skip := buildTypeCheckSkipMask(programSet.Programs); skip != nil {
		t.Fatalf("expected tsconfig-backed program to participate in type-check, got %v", skip)
	}

	_, binding := resolveAndBindTestTargets(
		t,
		programSet,
		cfg,
		dir,
		fs,
		nil,
		nil,
		parseCache,
	)
	programs := binding.Programs
	typeInfoFiles := binding.TypeInfoFiles
	gapFiles := binding.GapFiles
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
	programSet, err := createProgramSetForConfig(dir, cfg, true, fs, parseCache)
	if err != nil {
		t.Fatalf("createProgramSetForConfig: %v", err)
	}
	if len(programSet.Programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program, got %d", len(programSet.Programs))
	}

	libPath := tspath.NormalizePath(filepath.Join(dir, "lib.ts"))
	plan, binding := resolveAndBindTestTargets(
		t,
		programSet,
		cfg,
		dir,
		fs,
		[]string{libPath},
		nil,
		parseCache,
	)
	programs := binding.Programs
	typeInfoFiles := binding.TypeInfoFiles
	gapFiles := binding.GapFiles
	targetFiles := []string{plan.Targets[0].Path}
	targetsByProgram := binding.TargetsByProgram
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
	programSet, err := createProgramSetForConfig(linkDir, cfg, true, fs, parseCache)
	if err != nil {
		t.Fatalf("createProgramSetForConfig: %v", err)
	}
	programs := programSet.Programs
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

	plan, binding := resolveAndBindTestTargets(
		t,
		programSet,
		cfg,
		linkDir,
		fs,
		[]string{realTarget},
		nil,
		parseCache,
	)
	programs = binding.Programs
	typeInfoFiles := binding.TypeInfoFiles
	gapFiles := binding.GapFiles
	targetFiles := []string{plan.Targets[0].Path}
	targetsByProgram := binding.TargetsByProgram
	targetPathBySourcePath := binding.TargetPathBySourcePath
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
	if targetPathBySourcePath[sourceName] != realTarget {
		t.Fatalf("expected source path %q to resolve config through target path %q, got %v", sourceName, realTarget, targetPathBySourcePath)
	}
}

func TestBindLintTargetPlan_UsesPhysicalConfigSpaceForSymlinkedConfigRoot(t *testing.T) {
	realDir := t.TempDir()
	linkDir := filepath.Join(filepath.Dir(realDir), filepath.Base(realDir)+"-config-link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	defer os.Remove(linkDir)
	writeProgramTestFiles(t, realDir, map[string]string{
		"src/a.ts":      "debugger;\n",
		"tsconfig.json": `{"include":["src/a.ts"]}`,
	})

	linkDir = tspath.NormalizePath(linkDir)
	realTarget := tspath.NormalizePath(filepath.Join(realDir, "src/a.ts"))
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"src/**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{
			Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
		}},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}
	set, err := createProgramSetForConfig(linkDir, cfg, true, fsys, utils.NewParseCache())
	if err != nil || len(set.Programs) != 1 {
		t.Fatalf("create Program through symlinked config root: err=%v programs=%d", err, len(set.Programs))
	}
	plan := lintTargetPlan{Targets: []resolvedLintTarget{testLintTarget(fsys, linkDir, realTarget)}}
	binding, err := bindLintTargetPlan(set, plan, linkDir, fsys, utils.NewParseCache(), true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 1 {
		t.Fatalf("expected real target to bind to config Program, got %v", binding.TargetsByProgram)
	}
	sourcePath := binding.TargetsByProgram[0][0]
	configPath := binding.ConfigPathBySourcePath[sourcePath]
	if canonicalFilesystemPathID(configPath, fsys) != canonicalFilesystemPathID(realTarget, fsys) {
		t.Fatalf("config path must stay in the physical config-root space: source=%q config=%q target=%q", sourcePath, configPath, realTarget)
	}

	rslintconfig.RegisterAllRules()
	resolver := newLintConfigResolver(lintConfigResolverOptions{
		Config:                     cfg,
		CurrentDirectory:           linkDir,
		TypeInfoFiles:              binding.TypeInfoFiles,
		TargetPathBySourcePath:     binding.TargetPathBySourcePath,
		ConfigPathBySourcePath:     binding.ConfigPathBySourcePath,
		OwnerConfigDirBySourcePath: binding.OwnerConfigDirBySourcePath,
		FS:                         fsys,
	})
	rules := resolver.ActiveRulesForFile(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-debugger" {
		t.Fatalf("expected files selector to match in physical config space, got %v", configuredRuleNameSet(rules))
	}
}

func TestBindLintTargetPlan_BindsFileSymlinkOutsideProgramRoot(t *testing.T) {
	sharedDir := t.TempDir()
	writeProgramTestFiles(t, sharedDir, map[string]string{
		"shared.ts": `export const value = 1;`,
	})
	repoDir := t.TempDir()
	linkedPath := filepath.Join(repoDir, "linked.ts")
	realTarget := filepath.Join(sharedDir, "shared.ts")
	if err := os.Symlink(realTarget, linkedPath); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}
	writeProgramTestFiles(t, repoDir, map[string]string{
		"tsconfig.json": `{"files":["linked.ts"]}`,
	})

	repoDir = tspath.NormalizePath(repoDir)
	realTarget = tspath.NormalizePath(realTarget)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	cfg := projectConfig("./tsconfig.json")
	set, err := createProgramSetForConfig(repoDir, cfg, true, fsys, parseCache)
	if err != nil || len(set.Programs) != 1 {
		t.Fatalf("expected one Program for file-symlink fixture, err=%v programs=%d", err, len(set.Programs))
	}
	var sourceName string
	for _, sourceFile := range set.Programs[0].GetSourceFiles() {
		if strings.HasSuffix(sourceFile.FileName(), "/linked.ts") || sourceFile.FileName() == realTarget {
			sourceName = sourceFile.FileName()
			break
		}
	}
	if sourceName == "" {
		t.Fatal("expected Program to contain the symlinked source")
	}
	if sourceName == realTarget {
		t.Skip("compiler canonicalized the file symlink before Program lookup")
	}

	plan := lintTargetPlan{Targets: []resolvedLintTarget{testLintTarget(fsys, repoDir, realTarget)}}
	binding, err := bindLintTargetPlan(set, plan, repoDir, fsys, parseCache, true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	if len(binding.GapFiles) != 0 || len(binding.Programs) != 1 {
		t.Fatalf("real target should bind through the Program's file symlink, gaps=%v programs=%d", binding.GapFiles, len(binding.Programs))
	}
	if len(binding.TargetsByProgram[0]) != 1 || binding.TargetsByProgram[0][0] != sourceName {
		t.Fatalf("expected target to bind to Program source %q, got %v", sourceName, binding.TargetsByProgram)
	}
	if owner := binding.OwnerConfigDirBySourcePath[sourceName]; owner != repoDir {
		t.Fatalf("expected bound source owner %q, got %q", repoDir, owner)
	}
}

func testLintTarget(fsys vfs.FS, ownerDir string, filePath string) resolvedLintTarget {
	filePath = tspath.NormalizePath(filePath)
	canonicalPath := filePath
	if realPath := fsys.Realpath(filePath); realPath != "" {
		canonicalPath = tspath.NormalizePath(realPath)
	}
	return resolvedLintTarget{
		Path:           filePath,
		CanonicalPath:  canonicalPath,
		OwnerConfigDir: tspath.NormalizePath(ownerDir),
	}
}

func projectConfig(projects ...string) rslintconfig.RslintConfig {
	return rslintconfig.RslintConfig{{
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths(projects),
			},
		},
	}}
}

func TestBindLintTargetPlan_DoesNotBorrowParentConfigProgram(t *testing.T) {
	rootDir := t.TempDir()
	childDir := filepath.Join(rootDir, "child")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"tsconfig.json":   `{"include":["child/target.ts"]}`,
		"child/target.ts": `export const value = 1;`,
	})

	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	configMap := map[string]rslintconfig.RslintConfig{
		tspath.NormalizePath(rootDir):  projectConfig("./tsconfig.json"),
		tspath.NormalizePath(childDir): rslintconfig.RslintConfig{{}},
	}
	set, err := createProgramSetForConfigs(configMap, true, fsys, parseCache)
	if err != nil {
		t.Fatalf("createProgramSetForConfigs: %v", err)
	}
	if len(set.Programs) != 1 {
		t.Fatalf("expected only the root tsconfig Program, got %d", len(set.Programs))
	}

	targetPath := filepath.Join(childDir, "target.ts")
	plan := lintTargetPlan{Targets: []resolvedLintTarget{testLintTarget(fsys, childDir, targetPath)}}
	binding, err := bindLintTargetPlan(set, plan, rootDir, fsys, parseCache, true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}

	if len(binding.GapFiles) != 1 || binding.GapFiles[0] != tspath.NormalizePath(targetPath) {
		t.Fatalf("child-owned target must use fallback instead of borrowing the parent Program: %v", binding.GapFiles)
	}
	if len(binding.Programs) != 2 || len(binding.TargetsByProgram[0]) != 0 || len(binding.TargetsByProgram[1]) != 1 {
		t.Fatalf("expected target only in fallback Program, got targets=%v", binding.TargetsByProgram)
	}
	fallbackSource := binding.TargetsByProgram[1][0]
	if owner := binding.OwnerConfigDirBySourcePath[fallbackSource]; owner != tspath.NormalizePath(childDir) {
		t.Fatalf("expected fallback source owner %q, got %q", tspath.NormalizePath(childDir), owner)
	}
	if binding.TypeInfoFiles == nil || len(binding.TypeInfoFiles) != 0 {
		t.Fatalf("expected an explicit empty type-info set for the child gap target, got %v", binding.TypeInfoFiles)
	}
}

func TestTypeCheckDeduplicatesSyntaxFromGoverningFallbackAndParentProgram(t *testing.T) {
	rootDir := t.TempDir()
	childDir := filepath.Join(rootDir, "child")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"tsconfig.json":   `{"include":["child/target.ts"]}`,
		"child/target.ts": `let value: ;`,
	})

	rootDir = tspath.NormalizePath(rootDir)
	childDir = tspath.NormalizePath(childDir)
	configMap := map[string]rslintconfig.RslintConfig{
		rootDir:  projectConfig("./tsconfig.json"),
		childDir: rslintconfig.RslintConfig{{}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	set, err := createProgramSetForConfigs(configMap, true, fsys, parseCache)
	if err != nil {
		t.Fatalf("createProgramSetForConfigs: %v", err)
	}
	targetPath := tspath.NormalizePath(filepath.Join(childDir, "target.ts"))
	plan := lintTargetPlan{Targets: []resolvedLintTarget{testLintTarget(fsys, childDir, targetPath)}}
	binding, err := bindLintTargetPlan(set, plan, rootDir, fsys, parseCache, true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	if len(binding.GapFiles) != 1 {
		t.Fatalf("child-owned target must remain on fallback, got gaps %v", binding.GapFiles)
	}

	skip := buildTypeCheckSkipMask(binding.Programs)
	diagnostics, syntaxErrorFiles := collectTargetSyntacticDiagnostics(binding.Programs, binding.TargetsByProgram, skip, true, false)
	if len(syntaxErrorFiles) != 1 {
		t.Fatalf("expected one malformed lint target, got %v", syntaxErrorFiles)
	}
	diagnostics = append(diagnostics, collectProgramTypeDiagnostics(t, binding.Programs, skip, binding.TypeInfoFiles)...)
	remapDiagnosticTargetPaths(diagnostics, binding.TargetPathBySourcePath)
	if len(diagnostics) < 2 {
		t.Fatalf("fixture must exercise both fallback syntax and parent Program type-check paths, got %+v", diagnostics)
	}

	diagnostics = deduplicateTypeScriptDiagnostics(diagnostics, fsys)
	if len(diagnostics) != 1 || diagnostics[0].RuleName != "TypeScript(TS1110)" {
		t.Fatalf("expected one TS1110 diagnostic after cross-phase dedupe, got %+v", diagnostics)
	}
}

func TestCreateProgramSetForConfigs_DeduplicatesSharedTsconfigAndRetainsOwners(t *testing.T) {
	rootDir := t.TempDir()
	childDir := filepath.Join(rootDir, "child")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"tsconfig.json": `{"include":["src/**/*.ts"]}`,
		"src/a.ts":      `export const a = 1;`,
	})
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("mkdir child: %v", err)
	}

	rootKey := tspath.NormalizePath(rootDir)
	childKey := tspath.NormalizePath(childDir)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	set, err := createProgramSetForConfigs(map[string]rslintconfig.RslintConfig{
		rootKey:  projectConfig("./tsconfig.json"),
		childKey: projectConfig("../tsconfig.json"),
	}, true, fsys, utils.NewParseCache())
	if err != nil {
		t.Fatalf("createProgramSetForConfigs: %v", err)
	}
	if len(set.Programs) != 1 || len(set.ConfigOrders) != 1 {
		t.Fatalf("shared tsconfig must produce one Program, got programs=%d orders=%d", len(set.Programs), len(set.ConfigOrders))
	}
	if order, ok := set.ConfigOrders[0][rootKey]; !ok || order != 0 {
		t.Fatalf("missing root config association: %v", set.ConfigOrders[0])
	}
	if order, ok := set.ConfigOrders[0][childKey]; !ok || order != 0 {
		t.Fatalf("missing child config association: %v", set.ConfigOrders[0])
	}
}

func TestBindLintTargetPlan_UsesGoverningConfigProjectOrder(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"shared.ts":       `export const value = 1;`,
		"tsconfig-a.json": `{"files":["shared.ts"]}`,
		"tsconfig-b.json": `{"files":["shared.ts"]}`,
	})

	dir = tspath.NormalizePath(dir)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	set, err := createProgramSetForConfig(
		dir,
		projectConfig("./tsconfig-a.json", "./tsconfig-b.json"),
		true,
		fsys,
		parseCache,
	)
	if err != nil || len(set.Programs) != 2 {
		t.Fatalf("expected two ordered Programs, err=%v programs=%d", err, len(set.Programs))
	}

	targetPath := filepath.Join(dir, "shared.ts")
	plan := lintTargetPlan{Targets: []resolvedLintTarget{testLintTarget(fsys, dir, targetPath)}}
	binding, err := bindLintTargetPlan(set, plan, dir, fsys, parseCache, true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	if len(binding.TargetsByProgram[0]) != 1 || len(binding.TargetsByProgram[1]) != 0 {
		t.Fatalf("overlapping target must bind to the first declared project, got %v", binding.TargetsByProgram)
	}
}

func TestBindLintTargetPlan_RecomputesProgramMembershipAfterImportGraphChange(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"main.ts":       `import "./extra";`,
		"extra.ts":      `export const value = 1;`,
		"tsconfig.json": `{"files":["main.ts"]}`,
	})

	dir = tspath.NormalizePath(dir)
	config := projectConfig("./tsconfig.json")
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	set, err := createProgramSetForConfig(dir, config, true, fsys, parseCache)
	if err != nil {
		t.Fatalf("initial createProgramSetForConfig: %v", err)
	}
	plan := lintTargetPlan{Targets: []resolvedLintTarget{
		testLintTarget(fsys, dir, filepath.Join(dir, "extra.ts")),
	}}
	initial, err := bindLintTargetPlan(set, plan, dir, fsys, parseCache, true)
	if err != nil {
		t.Fatalf("initial bindLintTargetPlan: %v", err)
	}
	if len(initial.GapFiles) != 0 || len(initial.Programs) != 1 || len(initial.TargetsByProgram[0]) != 1 {
		t.Fatalf("imported target should initially use the real Program, got gaps=%v targets=%v", initial.GapFiles, initial.TargetsByProgram)
	}

	if err := os.WriteFile(filepath.Join(dir, "main.ts"), []byte(`export const main = 1;`), 0644); err != nil {
		t.Fatalf("rewrite main.ts: %v", err)
	}
	// Production fix passes reuse the run-scoped filesystem and parse caches.
	// ParseCache keys include the exact source text hash, so rewritten content
	// must still produce a fresh AST and updated import graph.
	rebuilt, err := createProgramSetForConfig(dir, config, true, fsys, parseCache)
	if err != nil {
		t.Fatalf("rebuilt createProgramSetForConfig: %v", err)
	}
	afterFix, err := bindLintTargetPlan(rebuilt, plan, dir, fsys, parseCache, true)
	if err != nil {
		t.Fatalf("rebuilt bindLintTargetPlan: %v", err)
	}
	if len(afterFix.GapFiles) != 1 || len(afterFix.Programs) != 2 || len(afterFix.TargetsByProgram[1]) != 1 {
		t.Fatalf("target must move to fallback after its importing edge is removed, got gaps=%v targets=%v", afterFix.GapFiles, afterFix.TargetsByProgram)
	}
}

func TestResolveLintTargetPlan_RejectsCanonicalTargetWithDifferentOwners(t *testing.T) {
	sharedDir := t.TempDir()
	writeProgramTestFiles(t, sharedDir, map[string]string{
		"target.ts": `export const value = 1;`,
	})
	ownersRoot := t.TempDir()
	ownerA := filepath.Join(ownersRoot, "owner-a")
	ownerB := filepath.Join(ownersRoot, "owner-b")
	if err := os.MkdirAll(ownerA, 0755); err != nil {
		t.Fatalf("mkdir owner A: %v", err)
	}
	if err := os.MkdirAll(ownerB, 0755); err != nil {
		t.Fatalf("mkdir owner B: %v", err)
	}
	sharedTarget := filepath.Join(sharedDir, "target.ts")
	targetA := filepath.Join(ownerA, "target.ts")
	targetB := filepath.Join(ownerB, "target.ts")
	if err := os.Symlink(sharedTarget, targetA); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := os.Symlink(sharedTarget, targetB); err != nil {
		t.Skipf("second symlink unavailable: %v", err)
	}

	ownerA = tspath.NormalizePath(ownerA)
	ownerB = tspath.NormalizePath(ownerB)
	targetA = tspath.NormalizePath(targetA)
	targetB = tspath.NormalizePath(targetB)
	configMap := map[string]rslintconfig.RslintConfig{
		ownerA: {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
		ownerB: {{Rules: rslintconfig.Rules{"no-console": "error"}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	_, err := resolveLintTargetPlan(
		configMap,
		nil,
		tspath.NormalizePath(ownersRoot),
		nil,
		fsys,
		[]string{targetA, targetB},
		nil,
		true,
	)
	if err == nil {
		t.Fatal("expected aliases governed by different configs to be rejected")
	}
	if !strings.Contains(err.Error(), "resolve to the same file") || !strings.Contains(err.Error(), ownerA) || !strings.Contains(err.Error(), ownerB) {
		t.Fatalf("unexpected ownership conflict error: %v", err)
	}
}
