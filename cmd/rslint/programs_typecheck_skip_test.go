package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type targetPlanRealpathCountingFS struct {
	vfs.FS
	mu    sync.Mutex
	calls map[string]int
}

func (f *targetPlanRealpathCountingFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	f.mu.Lock()
	f.calls[filePath]++
	f.mu.Unlock()
	return f.FS.Realpath(filePath)
}

func (f *targetPlanRealpathCountingFS) callCount(filePath string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[tspath.NormalizePath(filePath)]
}

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

func bindingTargetsByProgram(binding lintTargetBinding) [][]string {
	targets := make([][]string, len(binding.Programs))
	for _, view := range binding.Views {
		if view.ProgramIndex < 0 || view.ProgramIndex >= len(targets) {
			continue
		}
		targets[view.ProgramIndex] = append(targets[view.ProgramIndex], view.TargetFiles...)
	}
	for index := range targets {
		slices.Sort(targets[index])
	}
	return targets
}

func bindingTypeInfoFiles(binding lintTargetBinding) map[string]struct{} {
	hasGapView := false
	for _, view := range binding.Views {
		if view.TypeInfoFiles != nil {
			hasGapView = true
			break
		}
	}
	if !hasGapView {
		return nil
	}
	typeInfoFiles := make(map[string]struct{})
	for _, view := range binding.Views {
		if view.TypeInfoFiles != nil {
			continue
		}
		for _, sourcePath := range view.TargetFiles {
			typeInfoFiles[sourcePath] = struct{}{}
		}
		for sourcePath, targetPath := range view.TargetPathBySourcePath {
			typeInfoFiles[sourcePath] = struct{}{}
			typeInfoFiles[targetPath] = struct{}{}
		}
	}
	return typeInfoFiles
}

func bindingSourcePathMap(binding lintTargetBinding, selectMap func(lintTargetView) map[string]string) map[string]string {
	result := make(map[string]string)
	for _, view := range binding.Views {
		for sourcePath, value := range selectMap(view) {
			if _, exists := result[sourcePath]; !exists {
				result[sourcePath] = value
			}
		}
	}
	return result
}

func bindingTargetPaths(binding lintTargetBinding) map[string]string {
	return bindingSourcePathMap(binding, func(view lintTargetView) map[string]string {
		return view.TargetPathBySourcePath
	})
}

func bindingConfigPaths(binding lintTargetBinding) map[string]string {
	return bindingSourcePathMap(binding, func(view lintTargetView) map[string]string {
		return view.ConfigPathBySourcePath
	})
}

func bindingOwnerPaths(binding lintTargetBinding) map[string]string {
	return bindingSourcePathMap(binding, func(view lintTargetView) map[string]string {
		return view.OwnerConfigDirBySourcePath
	})
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
	plan, err := resolveLintTargetPlan(cfg, dir, fsys, allowFiles, allowDirs, true)
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
	typeInfoFiles := bindingTypeInfoFiles(binding)
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
		t.Fatalf("expected files-driven fallback to stay non-project-backed, got options %+v", programs[0].Options())
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
		"src/in-project.ts": `export const bad: number = "oops";
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
	typeInfoFiles := bindingTypeInfoFiles(binding)
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
		t.Fatalf("expected gap fallback to stay non-project-backed, got options %+v", programs[1].Options())
	}
	skip := buildTypeCheckSkipMask(programs)
	if len(skip) != 2 || skip[0] || !skip[1] {
		t.Fatalf("expected only the gap fallback program to be skipped, got %v", skip)
	}

	diags := collectProgramTypeDiagnostics(t, programs, skip, typeInfoFiles)
	if !containsTSDiagnostic(diags, "TS2322") {
		t.Fatalf("expected the tsconfig-backed Program to retain semantic diagnostics: %+v", diags)
	}
	if containsTSDiagnostic(diags, "TS2304") {
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
	if _, ok := programSet.ConfigOrders[0][exactHostPathID(dir)]; !ok {
		t.Fatalf("config order was not keyed by normalized directory identity: %v", programSet.ConfigOrders[0])
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
	typeInfoFiles := bindingTypeInfoFiles(binding)
	gapFiles := binding.GapFiles
	targetFiles := []string{plan.Targets[0].Path}
	targetsByProgram := bindingTargetsByProgram(binding)
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
	if len(targetsByProgram) != 1 || len(targetsByProgram[0]) != 1 ||
		canonicalHostPathID(targetsByProgram[0][0], fs) != canonicalHostPathID(libPath, fs) {
		t.Fatalf("expected lib.ts bound to the tsconfig Program, got %v", targetsByProgram)
	}
}

func TestOrderedProgramIndexesForConfig_UsesExactHostDirectory(t *testing.T) {
	set := lintProgramSet{
		Programs: []*compiler.Program{nil},
		ConfigOrders: []programConfigOrders{{
			exactHostPathID(`C:/Repo`): 0,
		}},
	}
	indexes := orderedProgramIndexesForConfig(set, "C:/Repo")
	if len(indexes) != 1 || indexes[0] != 0 {
		t.Fatalf("normalized config directory did not find its Program: %v", indexes)
	}
	if indexes := orderedProgramIndexesForConfig(set, "c:/repo"); len(indexes) != 0 {
		t.Fatalf("config directory identity unexpectedly folded case: %v", indexes)
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
	typeInfoFiles := bindingTypeInfoFiles(binding)
	gapFiles := binding.GapFiles
	targetFiles := []string{plan.Targets[0].Path}
	targetsByProgram := bindingTargetsByProgram(binding)
	targetPathBySourcePath := bindingTargetPaths(binding)
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
	if len(bindingTargetsByProgram(binding)) != 1 || len(bindingTargetsByProgram(binding)[0]) != 1 {
		t.Fatalf("expected real target to bind to config Program, got %v", bindingTargetsByProgram(binding))
	}
	sourcePath := bindingTargetsByProgram(binding)[0][0]
	configPath := bindingConfigPaths(binding)[sourcePath]
	if canonicalHostPathID(configPath, fsys) != canonicalHostPathID(realTarget, fsys) {
		t.Fatalf("config path must stay in the physical config-root space: source=%q config=%q target=%q", sourcePath, configPath, realTarget)
	}

	rslintconfig.RegisterAllRules()
	resolver := mustNewLintConfigResolver(t, lintConfigResolverOptions{
		Config:                     cfg,
		CurrentDirectory:           linkDir,
		TypeInfoFiles:              bindingTypeInfoFiles(binding),
		ConfigPathBySourcePath:     bindingConfigPaths(binding),
		OwnerConfigDirBySourcePath: bindingOwnerPaths(binding),
		FS:                         fsys,
	})
	rules := resolver.ActiveRulesForFile(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-debugger" {
		t.Fatalf("expected files selector to match in physical config space, got %v", configuredRuleNameSet(rules))
	}
}

func TestBindLintTargetPlan_ConfigMatchingDoesNotDependOnProgramSourcePath(t *testing.T) {
	rootDir := t.TempDir()
	writeProgramTestFiles(t, rootDir, map[string]string{
		"physical/index.ts": "console.log('value');\n",
		"tsconfig.json":     `{"files":["physical/index.ts"]}`,
	})
	linkPath := filepath.Join(rootDir, "link.ts")
	physicalPath := filepath.Join(rootDir, "physical/index.ts")
	if err := os.Symlink(physicalPath, linkPath); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}

	rootDir = tspath.NormalizePath(rootDir)
	linkPath = tspath.NormalizePath(linkPath)
	physicalPath = tspath.NormalizePath(physicalPath)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"link.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{
			Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
		}},
		Rules: rslintconfig.Rules{"no-console": "error"},
	}}
	set, err := createProgramSetForConfig(rootDir, cfg, true, fsys, utils.NewParseCache())
	if err != nil || len(set.Programs) != 1 {
		t.Fatalf("create Program: err=%v programs=%d", err, len(set.Programs))
	}
	plan := lintTargetPlan{Targets: []resolvedLintTarget{testLintTarget(fsys, rootDir, linkPath)}}
	binding, err := bindLintTargetPlan(set, plan, rootDir, fsys, utils.NewParseCache(), true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	if len(bindingTargetsByProgram(binding)) != 1 || len(bindingTargetsByProgram(binding)[0]) != 1 {
		t.Fatalf("expected lexical target to bind to the physical Program source, got %v", bindingTargetsByProgram(binding))
	}
	sourcePath := bindingTargetsByProgram(binding)[0][0]
	expectedSourcePath := authoritativeHostPath(physicalPath, fsys)
	if canonicalHostPathID(sourcePath, fsys) != canonicalHostPathID(expectedSourcePath, fsys) {
		t.Fatalf("fixture must bind through physical Program source %q, got %q", expectedSourcePath, sourcePath)
	}
	expectedConfigPath := linkPath
	if configPath := bindingConfigPaths(binding)[sourcePath]; configPath != expectedConfigPath {
		t.Fatalf("config matching must retain lexical target %q, got %q", expectedConfigPath, configPath)
	}

	rslintconfig.RegisterAllRules()
	resolver := mustNewLintConfigResolver(t, lintConfigResolverOptions{
		Config:                     cfg,
		CurrentDirectory:           rootDir,
		TypeInfoFiles:              bindingTypeInfoFiles(binding),
		ConfigPathBySourcePath:     bindingConfigPaths(binding),
		OwnerConfigDirBySourcePath: bindingOwnerPaths(binding),
		FS:                         fsys,
	})
	rules := resolver.ActiveRulesForFile(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-console" {
		t.Fatalf("Program membership changed the lexical files match: %v", configuredRuleNameSet(rules))
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
	if len(bindingTargetsByProgram(binding)[0]) != 1 || bindingTargetsByProgram(binding)[0][0] != sourceName {
		t.Fatalf("expected target to bind to Program source %q, got %v", sourceName, bindingTargetsByProgram(binding))
	}
	if owner := bindingOwnerPaths(binding)[sourceName]; owner != repoDir {
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
	if len(binding.Programs) != 2 || len(bindingTargetsByProgram(binding)[0]) != 0 || len(bindingTargetsByProgram(binding)[1]) != 1 {
		t.Fatalf("expected target only in fallback Program, got targets=%v", bindingTargetsByProgram(binding))
	}
	fallbackSource := bindingTargetsByProgram(binding)[1][0]
	if owner := bindingOwnerPaths(binding)[fallbackSource]; owner != tspath.NormalizePath(childDir) {
		t.Fatalf("expected fallback source owner %q, got %q", tspath.NormalizePath(childDir), owner)
	}
	if bindingTypeInfoFiles(binding) == nil || len(bindingTypeInfoFiles(binding)) != 0 {
		t.Fatalf("expected an explicit empty type-info set for the child gap target, got %v", bindingTypeInfoFiles(binding))
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
	diagnostics, syntaxErrorFiles := collectTargetViewSyntacticDiagnostics(binding, true, false)
	if len(syntaxErrorFiles) != 1 {
		t.Fatalf("expected one malformed lint target, got %v", syntaxErrorFiles)
	}
	diagnostics = append(diagnostics, collectProgramTypeDiagnostics(t, binding.Programs, skip, bindingTypeInfoFiles(binding))...)
	if len(diagnostics) < 2 {
		t.Fatalf("fixture must exercise both fallback syntax and parent Program type-check paths, got %+v", diagnostics)
	}

	diagnostics = deduplicateTypeScriptDiagnostics(diagnostics, fsys)
	if len(diagnostics) != 1 || diagnostics[0].RuleName != "TypeScript(TS1110)" {
		t.Fatalf("expected one TS1110 diagnostic after cross-phase dedupe, got %+v", diagnostics)
	}
	if diagnostics[0].Origin != rule.DiagnosticOriginTypeScript {
		t.Fatalf("deduplicated TypeScript diagnostic lost its origin: %+v", diagnostics[0])
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

func TestProgramViewsPreserveLexicalConfigsForSharedPhysicalSource(t *testing.T) {
	rootDir := t.TempDir()
	ownerA := filepath.Join(rootDir, "owner-a")
	ownerB := filepath.Join(rootDir, "owner-b")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"physical/shared.ts": "debugger;\nconsole.log('shared');\n",
		"tsconfig.json":      `{"files":["physical/shared.ts"]}`,
	})
	for _, directory := range []string{ownerA, ownerB} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	physical := filepath.Join(rootDir, "physical", "shared.ts")
	aliasA := filepath.Join(ownerA, "shared.ts")
	aliasB := filepath.Join(ownerB, "shared.ts")
	if err := os.Symlink(physical, aliasA); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}
	if err := os.Symlink(physical, aliasB); err != nil {
		t.Skipf("second file symlink unavailable: %v", err)
	}

	rootDir = tspath.NormalizePath(rootDir)
	ownerA = tspath.NormalizePath(ownerA)
	ownerB = tspath.NormalizePath(ownerB)
	aliasA = tspath.NormalizePath(aliasA)
	aliasB = tspath.NormalizePath(aliasB)
	projectOptions := func() *rslintconfig.LanguageOptions {
		return &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{
			Project: rslintconfig.ProjectPaths{"../tsconfig.json"},
		}}
	}
	configMap := map[string]rslintconfig.RslintConfig{
		ownerA: {{
			Files:           []string{"**/*.ts"},
			LanguageOptions: projectOptions(),
			Plugins:         []string{"@typescript-eslint", "view-alpha"},
			Rules: rslintconfig.Rules{
				"no-debugger": "error",
				"@typescript-eslint/no-unnecessary-type-arguments": "error",
				"view-alpha/check": "error",
			},
		}},
		ownerB: {{
			Files:           []string{"**/*.ts"},
			LanguageOptions: projectOptions(),
			Plugins:         []string{"@typescript-eslint", "view-beta"},
			Rules: rslintconfig.Rules{
				"no-console": "error",
				"@typescript-eslint/no-unnecessary-type-arguments": "error",
				"view-beta/check": "error",
			},
		}},
	}

	rslintconfig.RegisterAllRules()
	rslintconfig.RegisterEslintPluginRules([]rslintconfig.EslintPluginEntry{
		{Prefix: "view-alpha", RuleNames: []string{"check"}},
		{Prefix: "view-beta", RuleNames: []string{"check"}},
	})
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	set, err := createProgramSetForConfigs(configMap, true, fsys, parseCache)
	if err != nil {
		t.Fatalf("createProgramSetForConfigs: %v", err)
	}
	if len(set.Programs) != 1 {
		t.Fatalf("shared tsconfig produced %d Programs, want 1", len(set.Programs))
	}
	targetA := testLintTarget(fsys, ownerA, aliasA)
	targetA.MergedConfig = rslintconfig.NewFileConfigResolver(configMap[ownerA], ownerA, true, fsys).ConfigForFile(aliasA)
	targetB := testLintTarget(fsys, ownerB, aliasB)
	targetB.MergedConfig = rslintconfig.NewFileConfigResolver(configMap[ownerB], ownerB, true, fsys).ConfigForFile(aliasB)
	plan := lintTargetPlan{Targets: []resolvedLintTarget{targetA, targetB}}
	binding, err := bindLintTargetPlan(set, plan, rootDir, fsys, parseCache, true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	if len(binding.Programs) != 1 || len(binding.Views) != 2 || len(binding.GapFiles) != 0 {
		t.Fatalf("binding = programs:%d views:%d gaps:%v, want one shared Program and two project views", len(binding.Programs), len(binding.Views), binding.GapFiles)
	}
	if binding.Views[0].ProgramIndex != 0 || binding.Views[1].ProgramIndex != 0 {
		t.Fatalf("view Program indexes = (%d,%d), want both 0", binding.Views[0].ProgramIndex, binding.Views[1].ProgramIndex)
	}

	var diagnostics []rule.RuleDiagnostic
	programViews, resolvers, err := buildBindingLintProgramViews(binding, lintConfigResolverOptions{
		ConfigMap:      configMap,
		EnforcePlugins: true,
		FS:             fsys,
	}, func(d rule.RuleDiagnostic) {
		diagnostics = append(diagnostics, d)
	})
	if err != nil {
		t.Fatalf("buildBindingLintProgramViews: %v", err)
	}
	if len(programViews) != 2 || programViews[0].Program != programViews[1].Program {
		t.Fatalf("ProgramViews did not share the project-backed Program: %+v", programViews)
	}
	runOptions := linter.RunLinterOptions{
		Programs:       binding.Programs,
		ProgramViews:   programViews,
		SingleThreaded: true,
		ExcludePaths:   []string{},
	}
	if _, err := linter.RunLinter(runOptions); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if len(diagnostics) != 2 {
		t.Fatalf("native diagnostics = %+v, want one per lexical view", diagnostics)
	}
	ruleByPath := make(map[string]string, len(diagnostics))
	for _, diagnostic := range diagnostics {
		ruleByPath[diagnostic.FilePath] = diagnostic.RuleName
	}
	if ruleByPath[aliasA] != "no-debugger" || ruleByPath[aliasB] != "no-console" {
		t.Fatalf("view-local native diagnostic routing = %v", ruleByPath)
	}

	targets := linter.CollectLintTargets(runOptions)
	if len(targets) != 2 || targets[0].File != targets[1].File || targets[0].ViewIndex == targets[1].ViewIndex {
		t.Fatalf("lint targets = %+v, want one SourceFile exposed through two distinct views", targets)
	}
	wantPathByView := map[int]string{0: aliasA, 1: aliasB}
	wantPluginByView := map[int]string{0: "view-alpha/check", 1: "view-beta/check"}
	for _, target := range targets {
		if target.Path != wantPathByView[target.ViewIndex] {
			t.Errorf("view %d target path = %q, want %q", target.ViewIndex, target.Path, wantPathByView[target.ViewIndex])
		}
		names := configuredRuleNameSet(target.Rules)
		if !names[wantPluginByView[target.ViewIndex]] {
			t.Errorf("view %d rules = %v, missing own plugin rule", target.ViewIndex, names)
		}
		hasTypeAware := false
		for _, configured := range target.Rules {
			if configured.Name == "@typescript-eslint/no-unnecessary-type-arguments" && configured.RequiresTypeInfo {
				hasTypeAware = true
			}
		}
		if !hasTypeAware {
			t.Errorf("view %d lost its project type-aware rule: %v", target.ViewIndex, names)
		}
	}

	pluginInputs := buildPluginFileInputs(runOptions, pluginConfigResolver{
		lintResolvers: resolvers,
		pluginConfigDirByOwner: map[string]string{
			ownerA: "wire-owner-a",
			ownerB: "wire-owner-b",
		},
	})
	if len(pluginInputs) != 2 {
		t.Fatalf("plugin inputs = %+v, want one per view", pluginInputs)
	}
	wireKeyByPath := map[string]string{}
	for _, input := range pluginInputs {
		wireKeyByPath[input.Path] = input.ConfigKey
	}
	if wireKeyByPath[aliasA] != "wire-owner-a" || wireKeyByPath[aliasB] != "wire-owner-b" {
		t.Fatalf("view-local plugin config routing = %v", wireKeyByPath)
	}
}

func TestCreateProgramSetForConfigs_PreservesSymlinkedTsconfigBase(t *testing.T) {
	rootDir := t.TempDir()
	realDir := filepath.Join(rootDir, "z-real")
	aliasDir := filepath.Join(rootDir, "a-alias")
	writeProgramTestFiles(t, realDir, map[string]string{
		"tsconfig.json": `{"include":["src/**/*.ts"]}`,
		"src/real.ts":   `export const source = "real";`,
	})
	writeProgramTestFiles(t, aliasDir, map[string]string{
		"src/alias.ts": `export const source = "alias";`,
	})
	aliasConfig := filepath.Join(aliasDir, "tsconfig.json")
	if err := os.Symlink(filepath.Join(realDir, "tsconfig.json"), aliasConfig); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	realDir = tspath.NormalizePath(realDir)
	aliasDir = tspath.NormalizePath(aliasDir)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	set, err := createProgramSetForConfigs(map[string]rslintconfig.RslintConfig{
		aliasDir: projectConfig("./tsconfig.json"),
		realDir:  projectConfig("./tsconfig.json"),
	}, true, fsys, utils.NewParseCache())
	if err != nil {
		t.Fatalf("createProgramSetForConfigs: %v", err)
	}
	if len(set.Programs) != 2 {
		t.Fatalf("distinct declared tsconfig paths must produce two Programs, got %d", len(set.Programs))
	}
	programByConfigPath := make(map[string]*compiler.Program, len(set.Programs))
	for _, program := range set.Programs {
		programByConfigPath[compilerPathID(program.Options().ConfigFilePath)] = program
	}
	aliasConfigPath := tspath.ResolvePath(aliasDir, "tsconfig.json")
	realConfigPath := tspath.ResolvePath(realDir, "tsconfig.json")
	aliasProgram := programByConfigPath[compilerPathID(aliasConfigPath)]
	realProgram := programByConfigPath[compilerPathID(realConfigPath)]
	if aliasProgram == nil || realProgram == nil {
		t.Fatalf("missing lexical tsconfig Programs: %v", programByConfigPath)
	}
	aliasSource := tspath.ResolvePath(aliasDir, "src/alias.ts")
	realSource := tspath.ResolvePath(realDir, "src/real.ts")
	if aliasProgram.GetSourceFile(aliasSource) == nil || aliasProgram.GetSourceFile(realSource) != nil {
		t.Fatalf("symlinked tsconfig must resolve includes from %q", aliasDir)
	}
	if realProgram.GetSourceFile(realSource) == nil || realProgram.GetSourceFile(aliasSource) != nil {
		t.Fatalf("real tsconfig must resolve includes from %q", realDir)
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
	if len(bindingTargetsByProgram(binding)[0]) != 1 || len(bindingTargetsByProgram(binding)[1]) != 0 {
		t.Fatalf("overlapping target must bind to the first declared project, got %v", bindingTargetsByProgram(binding))
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
	if len(initial.GapFiles) != 0 || len(initial.Programs) != 1 || len(bindingTargetsByProgram(initial)[0]) != 1 {
		t.Fatalf("imported target should initially use the real Program, got gaps=%v targets=%v", initial.GapFiles, bindingTargetsByProgram(initial))
	}

	if err := os.WriteFile(filepath.Join(dir, "main.ts"), []byte(`export const main = 1;`), 0644); err != nil {
		t.Fatalf("rewrite main.ts: %v", err)
	}
	// Production fix passes reuse the run-scoped filesystem and parse caches.
	// Replacing the source-snapshot generation before rebuilding makes the new
	// text/hash visible while retaining content-keyed AST entries for unchanged
	// files.
	parseCache.InvalidateSourceSnapshots()
	rebuilt, err := createProgramSetForConfig(dir, config, true, fsys, parseCache)
	if err != nil {
		t.Fatalf("rebuilt createProgramSetForConfig: %v", err)
	}
	afterFix, err := bindLintTargetPlan(rebuilt, plan, dir, fsys, parseCache, true)
	if err != nil {
		t.Fatalf("rebuilt bindLintTargetPlan: %v", err)
	}
	if len(afterFix.GapFiles) != 1 || len(afterFix.Programs) != 2 || len(bindingTargetsByProgram(afterFix)[1]) != 1 {
		t.Fatalf("target must move to fallback after its importing edge is removed, got gaps=%v targets=%v", afterFix.GapFiles, bindingTargetsByProgram(afterFix))
	}
}

func TestResolveLintTargetPlan_DirectoryWalkAvoidsPerTargetRealpath(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"src/a.ts": `export const a = 1;`,
		"src/b.ts": `export const b = 2;`,
	})
	dir = tspath.NormalizePath(dir)
	fileA := tspath.ResolvePath(dir, "src/a.ts")
	fileB := tspath.ResolvePath(dir, "src/b.ts")
	counter := &targetPlanRealpathCountingFS{FS: osvfs.FS(), calls: make(map[string]int)}
	fsys := bundled.WrapFS(cachedvfs.From(counter))

	plan, err := resolveLintTargetPlan(
		rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
		dir,
		fsys,
		nil,
		[]string{dir},
		true,
	)
	if err != nil {
		t.Fatalf("resolveLintTargetPlan: %v", err)
	}
	if len(plan.Targets) != 2 {
		t.Fatalf("targets = %v, want two files", plan.Targets)
	}
	if counter.callCount(fileA) != 0 || counter.callCount(fileB) != 0 {
		t.Fatalf("regular targets performed realpath IO: a=%d b=%d", counter.callCount(fileA), counter.callCount(fileB))
	}
}

func TestConfigsForLintTargetPlan_SelectsOnlyGoverningConfigs(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/repo/a": {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
		"/repo/b": {{LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{Project: []string{"./missing.json"}},
		}}},
	}
	active := configsForLintTargetPlan(configMap, lintTargetPlan{Targets: []resolvedLintTarget{{
		Path:           "/repo/a/index.ts",
		CanonicalPath:  "/repo/a/index.ts",
		OwnerConfigDir: "/repo/a",
	}}})
	if len(active) != 1 || active["/repo/a"] == nil {
		t.Fatalf("expected only the governing config, got %v", active)
	}
	if _, present := active["/repo/b"]; present {
		t.Fatalf("inactive config unexpectedly selected: %v", active)
	}
}

func TestPlainProgramSetSkipsInactiveConfigProjects(t *testing.T) {
	root := t.TempDir()
	activeDir := filepath.Join(root, "active")
	inactiveDir := filepath.Join(root, "inactive")
	writeProgramTestFiles(t, root, map[string]string{
		"active/index.ts":      "export const value = 1;\n",
		"active/tsconfig.json": `{"files":["index.ts"]}`,
	})
	if err := os.MkdirAll(inactiveDir, 0o755); err != nil {
		t.Fatalf("mkdir inactive config: %v", err)
	}
	activeDir = tspath.NormalizePath(activeDir)
	inactiveDir = tspath.NormalizePath(inactiveDir)
	configMap := map[string]rslintconfig.RslintConfig{
		activeDir:   projectConfig("./tsconfig.json"),
		inactiveDir: projectConfig("./missing.json"),
	}
	plan := lintTargetPlan{Targets: []resolvedLintTarget{{
		Path:           tspath.ResolvePath(activeDir, "index.ts"),
		CanonicalPath:  tspath.ResolvePath(activeDir, "index.ts"),
		OwnerConfigDir: activeDir,
	}}}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	activeConfigMap := configsForLintTargetPlan(configMap, plan)
	set, err := createProgramSetForConfigs(activeConfigMap, true, fsys, utils.NewParseCache())
	if err != nil || len(set.Programs) != 1 {
		t.Fatalf("plain lint should build only the active config Program: programs=%d err=%v", len(set.Programs), err)
	}
	if _, err := createProgramSetForConfigs(configMap, true, fsys, utils.NewParseCache()); err == nil || !strings.Contains(err.Error(), "missing.json") {
		t.Fatalf("the all-project type-check scope must still reject the inactive missing project, got %v", err)
	}
}

type canonicalIdentityTestFS struct {
	vfs.FS
	realPaths map[string]string
}

type exactCaseProgramFS struct {
	vfs.FS
	files map[string]string
}

func (fs *exactCaseProgramFS) UseCaseSensitiveFileNames() bool { return false }
func (fs *exactCaseProgramFS) FileExists(filePath string) bool {
	if _, ok := fs.files[tspath.NormalizePath(filePath)]; ok {
		return true
	}
	return fs.FS.FileExists(filePath)
}
func (fs *exactCaseProgramFS) ReadFile(filePath string) (string, bool) {
	if content, ok := fs.files[tspath.NormalizePath(filePath)]; ok {
		return content, true
	}
	return fs.FS.ReadFile(filePath)
}
func (fs *exactCaseProgramFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if _, ok := fs.files[filePath]; ok {
		return filePath
	}
	return fs.FS.Realpath(filePath)
}

func (fs *canonicalIdentityTestFS) UseCaseSensitiveFileNames() bool { return false }
func (fs *canonicalIdentityTestFS) FileExists(string) bool          { return true }
func (fs *canonicalIdentityTestFS) Realpath(filePath string) string {
	if realPath := fs.realPaths[tspath.NormalizePath(filePath)]; realPath != "" {
		return realPath
	}
	return tspath.NormalizePath(filePath)
}

func TestResolveLintTargetPlan_UsesCanonicalIdentityInsteadOfGlobalCaseFlag(t *testing.T) {
	configDir := "C:/Repo"
	upper := "C:/Repo/Src/A.ts"
	lower := "c:/repo/src/a.ts"
	config := rslintconfig.RslintConfig{{}}

	t.Run("same canonical path preserves lexical inputs", func(t *testing.T) {
		fsys := &canonicalIdentityTestFS{
			FS: osvfs.FS(),
			realPaths: map[string]string{
				upper: "C:/Repo/Src/A.ts",
				lower: "C:/Repo/Src/A.ts",
			},
		}
		plan, err := resolveLintTargetPlan(config, configDir, fsys, []string{upper, lower}, nil, true)
		if err != nil || len(plan.Targets) != 2 {
			t.Fatalf("distinct lexical targets must be preserved: targets=%v err=%v", plan.Targets, err)
		}
	})

	t.Run("distinct canonical paths remain distinct", func(t *testing.T) {
		fsys := &canonicalIdentityTestFS{
			FS: osvfs.FS(),
			realPaths: map[string]string{
				upper: upper,
				lower: lower,
			},
		}
		plan, err := resolveLintTargetPlan(config, configDir, fsys, []string{upper, lower}, nil, true)
		if err != nil || len(plan.Targets) != 2 {
			t.Fatalf("global case behavior must not merge distinct physical paths: targets=%v err=%v", plan.Targets, err)
		}
	})
}

func TestBindLintTargetPlan_RejectsCaseFoldedSourceWithDifferentCanonicalIdentity(t *testing.T) {
	configDir := "/repo"
	upper := "/repo/Source.ts"
	lower := "/repo/source.ts"
	fsys := &exactCaseProgramFS{
		FS: osvfs.FS(),
		files: map[string]string{
			upper: "export const upper = 1;\n",
			lower: "export const lower = 2;\n",
		},
	}
	host := utils.CreateCompilerHost(configDir, fsys)
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, []string{upper}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptionsLenient: %v", err)
	}
	if source := program.GetSourceFile(lower); source == nil || source.FileName() != upper {
		t.Fatalf("fixture must exercise case-folded Program lookup, got %v", source)
	}

	set := lintProgramSet{
		Programs:     []*compiler.Program{program},
		ConfigOrders: []programConfigOrders{{configDir: 0}},
	}
	plan := lintTargetPlan{Targets: []resolvedLintTarget{{
		Path:           lower,
		CanonicalPath:  lower,
		OwnerConfigDir: configDir,
	}}}
	binding, err := bindLintTargetPlan(set, plan, configDir, fsys, utils.NewParseCache(), true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	if len(binding.Programs) != 2 || len(bindingTargetsByProgram(binding)[0]) != 0 {
		t.Fatalf("lower-case target must not bind to the distinct upper-case source: %+v", bindingTargetsByProgram(binding))
	}
	if got := bindingTargetsByProgram(binding)[1]; len(got) != 1 || got[0] != lower {
		t.Fatalf("lower-case target must bind to its exact fallback source, got %v", got)
	}
}

func TestBindLintTargetPlan_SplitsFallbackProgramsForCaseFoldedPathCollisions(t *testing.T) {
	configDir := "/repo"
	upper := "/repo/Source.ts"
	lower := "/repo/source.ts"
	fsys := &exactCaseProgramFS{
		FS: osvfs.FS(),
		files: map[string]string{
			upper: "export const upper = 1;\n",
			lower: "export const lower = 2;\n",
		},
	}
	plan := lintTargetPlan{Targets: []resolvedLintTarget{
		{Path: upper, CanonicalPath: upper, OwnerConfigDir: configDir},
		{Path: lower, CanonicalPath: lower, OwnerConfigDir: configDir},
	}}
	binding, err := bindLintTargetPlan(lintProgramSet{}, plan, configDir, fsys, utils.NewParseCache(), true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	if len(binding.Programs) != 2 || len(bindingTargetsByProgram(binding)) != 2 {
		t.Fatalf("case-folded root names require separate fallback Programs, got %d", len(binding.Programs))
	}
	bound := []string{bindingTargetsByProgram(binding)[0][0], bindingTargetsByProgram(binding)[1][0]}
	slices.Sort(bound)
	want := []string{upper, lower}
	slices.Sort(want)
	if !slices.Equal(bound, want) {
		t.Fatalf("fallback Programs must preserve both exact source identities: got %v want %v", bound, want)
	}
}
