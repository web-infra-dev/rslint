package loader

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type targetPlanRealpathCountingFS struct {
	vfs.FS
	mu    sync.Mutex
	calls map[string]int
}

type blockingProgramConfigFS struct {
	vfs.FS
	paths      map[string]struct{}
	waitFor    int
	mu         sync.Mutex
	active     int
	peak       int
	allStarted chan struct{}
	release    chan struct{}
	startOnce  sync.Once
}

func (f *blockingProgramConfigFS) ReadFile(filePath string) (string, bool) {
	if _, blocks := f.paths[tspath.NormalizePath(filePath)]; !blocks {
		return f.FS.ReadFile(filePath)
	}

	f.mu.Lock()
	f.active++
	if f.active > f.peak {
		f.peak = f.active
	}
	if f.active == f.waitFor {
		f.startOnce.Do(func() { close(f.allStarted) })
	}
	f.mu.Unlock()

	<-f.release
	content, ok := f.FS.ReadFile(filePath)
	f.mu.Lock()
	f.active--
	f.mu.Unlock()
	return content, ok
}

func (f *blockingProgramConfigFS) peakConcurrency() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.peak
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
	programs []*lintprogram.Program,
) []rule.RuleDiagnostic {
	t.Helper()

	var diags []rule.RuleDiagnostic
	_, err := linter.RunLinter(linter.RunLinterOptions{
		Programs:       programs,
		SingleThreaded: true,
		TypeCheck:      true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) {
				diags = append(diags, d)
			},
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
	set ProjectSet,
	cfg rslintconfig.RslintConfig,
	dir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	buildContext *buildContext,
) (rslintconfig.LintTargetPlan, LoadResult) {
	t.Helper()
	plan, err := resolveTargetPlanForTest(nil, cfg, dir, nil, fsys, allowFiles, allowDirs, true)
	if err != nil {
		t.Fatalf("resolveTargetPlanForTest: %v", err)
	}
	binding, err := loadAPIForTest(set, plan, dir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	return plan, binding
}

func TestTypeCheck_SkipsNoTsconfigSourceOnlyProgram(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"bad.ts": `const bad: number = "oops";
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fs)
	programSet, err := buildProjectsForConfig(
		dir,
		rslintconfig.RslintConfig{{Files: []string{"**/*.ts"}}},
		true,
		buildContext,
	)
	if err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	if len(programSet.compilerPrograms) != 0 {
		t.Fatalf("expected no tsconfig-backed programs, got %d", len(programSet.compilerPrograms))
	}

	_, binding := resolveAndBindTestTargets(
		t,
		programSet,
		rslintconfig.RslintConfig{{Files: []string{"**/*.ts"}}},
		dir,
		fs,
		nil,
		nil,
		buildContext,
	)
	programs := binding.compilerPrograms
	if len(programs) != 1 {
		t.Fatalf("expected one compatibility compiler Program, got %d", len(programs))
	}
	if len(binding.Programs) != 1 || binding.Programs[0].CanProvideProgramDiagnostics() {
		t.Fatalf("expected one source-only Program, got %+v", binding.Programs)
	}
	if got := programs[0].Options().ConfigFilePath; got != "" {
		t.Fatalf("expected compatibility Program to have no ConfigFilePath, got %q", got)
	}
	if !programs[0].Options().NoLib.IsTrue() || !programs[0].Options().NoResolve.IsTrue() {
		t.Fatalf("expected compatibility Program to stay non-project-backed, got options %+v", programs[0].Options())
	}
	if diags := collectProgramTypeDiagnostics(t, binding.Programs); containsTSDiagnostic(diags, "TS2322") {
		t.Fatalf("did not expect semantic diagnostics from a source-only Program: %+v", diags)
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
	programSet, err := buildProjectsForConfig(
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
		newBuildContext(fs),
	)
	if err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	programs := programSet.compilerPrograms
	if len(programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program, got %d", len(programs))
	}
	if got := programs[0].Options().ConfigFilePath; got == "" {
		t.Fatal("expected tsconfig-backed program to carry ConfigFilePath")
	}
	lintPrograms := lintprogram.NewFromCompilers(programs)
	if !lintPrograms[0].CanProvideProgramDiagnostics() {
		t.Fatal("expected tsconfig-backed Program to expose type-check capability")
	}

	diags := collectProgramTypeDiagnostics(t, lintPrograms)
	if !containsTSDiagnostic(diags, "TS2304") {
		var rendered []string
		for _, d := range diags {
			rendered = append(rendered, d.RuleName+": "+d.Message.Description)
		}
		t.Fatalf("expected TS2304 from the tsconfig-covered declaration graph, got:\n%s", strings.Join(rendered, "\n"))
	}
}

func TestTypeCheck_SkipsSourceOnlyPrograms(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"tsconfig.json": `{
  "compilerOptions": { "skipLibCheck": false },
  "include": ["src/in-project.ts"]
}
`,
		"src/in-project.ts": `export const bad: number = "oops";
`,
		"source-only.ts": `import type { Bad } from './bad';
export const value: Bad | null = null;
`,
		"bad.d.ts": `export type Bad = MissingGlobalType;
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fs)
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
			},
		},
	}}
	programSet, err := buildProjectsForConfig(
		dir,
		cfg,
		true,
		buildContext,
	)
	if err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	if len(programSet.compilerPrograms) != 1 {
		t.Fatalf("expected one tsconfig-backed program before source-only Program, got %d", len(programSet.compilerPrograms))
	}
	if !lintprogram.NewFromCompiler(programSet.compilerPrograms[0]).CanProvideProgramDiagnostics() {
		t.Fatal("expected tsconfig-backed Program to participate in type-check")
	}

	_, binding := resolveAndBindTestTargets(
		t,
		programSet,
		cfg,
		dir,
		fs,
		nil,
		nil,
		buildContext,
	)
	programs := binding.compilerPrograms
	if len(programs) != 2 {
		t.Fatalf("expected the source-only Program to be appended, got %d programs", len(programs))
	}
	sourceOnlyPath := filepath.ToSlash(filepath.Join(dir, "source-only.ts"))
	if !slices.Contains(binding.TargetsByProgram[1], sourceOnlyPath) {
		t.Fatalf("expected source-only.ts to target the source-only Program, got %v", binding.TargetsByProgram)
	}
	if got := programs[0].Options().ConfigFilePath; got == "" {
		t.Fatal("expected original tsconfig-backed program to carry ConfigFilePath")
	}
	if got := programs[1].Options().ConfigFilePath; got != "" {
		t.Fatalf("expected source-only Program to have no ConfigFilePath, got %q", got)
	}
	if !programs[1].Options().NoLib.IsTrue() || !programs[1].Options().NoResolve.IsTrue() {
		t.Fatalf("expected source-only Program to stay non-project-backed, got options %+v", programs[1].Options())
	}
	if len(binding.Programs) != 2 ||
		!binding.Programs[0].CanProvideProgramDiagnostics() ||
		binding.Programs[1].CanProvideProgramDiagnostics() {
		t.Fatalf("unexpected Program diagnostic capabilities: %+v", binding.Programs)
	}

	diags := collectProgramTypeDiagnostics(t, binding.Programs)
	if !containsTSDiagnostic(diags, "TS2322") {
		t.Fatalf("expected the tsconfig-backed Program to retain semantic diagnostics: %+v", diags)
	}
	if containsTSDiagnostic(diags, "TS2304") {
		t.Fatalf("did not expect declaration diagnostics from a source-only Program: %+v", diags)
	}
}

func TestLoadProgramsBindsImportedNonRootFile(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"main.ts":       "import { value } from './lib';\nconsole.log(value);\n",
		"lib.ts":        "export const value = 1;\n",
		"tsconfig.json": `{"files": ["main.ts"], "compilerOptions": {"module": "ESNext"}}`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fs)
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
			},
		},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}
	programSet, err := buildProjectsForConfig(dir, cfg, true, buildContext)
	if err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	if len(programSet.compilerPrograms) != 1 {
		t.Fatalf("expected one tsconfig-backed program, got %d", len(programSet.compilerPrograms))
	}
	if _, ok := programSet.configOrders[0][exactPathID(dir)]; !ok {
		t.Fatalf("config order was not keyed by normalized directory identity: %v", programSet.configOrders[0])
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
		buildContext,
	)
	programs := binding.compilerPrograms
	targetFiles := []string{plan.Targets[0].Path}
	targetsByProgram := binding.TargetsByProgram
	if len(programs) != 1 {
		t.Fatalf("imported non-root target should reuse existing Program, got %d programs", len(programs))
	}
	if len(binding.Programs) != 1 || !binding.Programs[0].CanProvideTypeChecker(binding.Programs[0].GetSourceFile(targetsByProgram[0][0])) {
		t.Fatal("imported non-root target lost its compiler-capable Program")
	}
	if len(targetFiles) != 1 || targetFiles[0] != libPath {
		t.Fatalf("expected lib.ts as the only target, got %v", targetFiles)
	}
	if len(targetsByProgram) != 1 || len(targetsByProgram[0]) != 1 ||
		canonicalPathID(targetsByProgram[0][0], fs) != canonicalPathID(libPath, fs) {
		t.Fatalf("expected lib.ts bound to the tsconfig Program, got %v", targetsByProgram)
	}
}

func TestOrderedProgramIndexesForConfig_NormalizesDirectorySeparators(t *testing.T) {
	set := ProjectSet{
		compilerPrograms: []*compiler.Program{nil},
		configOrders: []configOrders{{
			exactPathID(`C:\Repo`): 0,
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

func TestLoadProgramsBindsRealpathTargetToProgramSourceName(t *testing.T) {
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
	buildContext := newBuildContext(fs)
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
	programSet, err := buildProjectsForConfig(linkDir, cfg, true, buildContext)
	if err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	programs := programSet.compilerPrograms
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
		buildContext,
	)
	programs = binding.compilerPrograms
	targetFiles := []string{plan.Targets[0].Path}
	targetsByProgram := binding.TargetsByProgram
	targetPathBySourcePath := binding.TargetPathBySourcePath
	if len(programs) != 1 {
		t.Fatalf("realpath target should reuse existing Program, got %d programs", len(programs))
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

func TestLoadProgramsUsesPhysicalConfigSpaceForSymlinkedConfigRoot(t *testing.T) {
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
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfig(linkDir, cfg, true, buildContext)
	if err != nil || len(set.compilerPrograms) != 1 {
		t.Fatalf("create Program through symlinked config root: err=%v programs=%d", err, len(set.compilerPrograms))
	}
	plan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{testLintTarget(fsys, linkDir, realTarget)}}
	binding, err := loadAPIForTest(set, plan, linkDir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 1 {
		t.Fatalf("expected real target to bind to config Program, got %v", binding.TargetsByProgram)
	}
	sourcePath := binding.TargetsByProgram[0][0]
	configPath := binding.ConfigPathBySourcePath[sourcePath]
	if canonicalPathID(configPath, fsys) != canonicalPathID(realTarget, fsys) {
		t.Fatalf("config path must stay in the physical config-root space: source=%q config=%q target=%q", sourcePath, configPath, realTarget)
	}

	rslintconfig.RegisterAllRules()
	resolver := newLintConfigResolver(lintConfigResolverOptions{
		Config:                     cfg,
		CurrentDirectory:           linkDir,
		ConfigPathBySourcePath:     binding.ConfigPathBySourcePath,
		OwnerConfigDirBySourcePath: binding.OwnerConfigDirBySourcePath,
		FS:                         fsys,
	})
	rules := resolver.EnabledRulesForFile(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-debugger" {
		t.Fatalf("expected files selector to match in physical config space, got %v", configuredRuleNameSet(rules))
	}
}

func TestLoadProgramsConfigMatchingDoesNotDependOnProgramSourcePath(t *testing.T) {
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
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfig(rootDir, cfg, true, buildContext)
	if err != nil || len(set.compilerPrograms) != 1 {
		t.Fatalf("create Program: err=%v programs=%d", err, len(set.compilerPrograms))
	}
	plan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{testLintTarget(fsys, rootDir, linkPath)}}
	binding, err := loadAPIForTest(set, plan, rootDir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 1 {
		t.Fatalf("expected lexical target to bind to the physical Program source, got %v", binding.TargetsByProgram)
	}
	sourcePath := binding.TargetsByProgram[0][0]
	expectedSourcePath := authoritativePath(physicalPath, fsys)
	if canonicalPathID(sourcePath, fsys) != canonicalPathID(expectedSourcePath, fsys) {
		t.Fatalf("fixture must bind through physical Program source %q, got %q", expectedSourcePath, sourcePath)
	}
	expectedConfigPath := tspath.ResolvePath(authoritativePath(rootDir, fsys), "link.ts")
	if configPath := binding.ConfigPathBySourcePath[sourcePath]; configPath != expectedConfigPath {
		t.Fatalf("config matching must retain lexical target %q, got %q", expectedConfigPath, configPath)
	}

	rslintconfig.RegisterAllRules()
	resolver := newLintConfigResolver(lintConfigResolverOptions{
		Config:                     cfg,
		CurrentDirectory:           rootDir,
		ConfigPathBySourcePath:     binding.ConfigPathBySourcePath,
		OwnerConfigDirBySourcePath: binding.OwnerConfigDirBySourcePath,
		FS:                         fsys,
	})
	rules := resolver.EnabledRulesForFile(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-console" {
		t.Fatalf("Program membership changed the lexical files match: %v", configuredRuleNameSet(rules))
	}
}

func TestLoadProgramsBindsFileSymlinkOutsideProgramRoot(t *testing.T) {
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
	buildContext := newBuildContext(fsys)
	cfg := projectConfig("./tsconfig.json")
	set, err := buildProjectsForConfig(repoDir, cfg, true, buildContext)
	if err != nil || len(set.compilerPrograms) != 1 {
		t.Fatalf("expected one Program for file-symlink fixture, err=%v programs=%d", err, len(set.compilerPrograms))
	}
	var sourceName string
	for _, sourceFile := range set.compilerPrograms[0].GetSourceFiles() {
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

	plan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{testLintTarget(fsys, repoDir, realTarget)}}
	binding, err := loadAPIForTest(set, plan, repoDir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.compilerPrograms) != 1 || len(binding.TargetsByProgram[0]) != 1 {
		t.Fatalf("real target should bind through the Program's file symlink, targets=%v", binding.TargetsByProgram)
	}
	if len(binding.TargetsByProgram[0]) != 1 || binding.TargetsByProgram[0][0] != sourceName {
		t.Fatalf("expected target to bind to Program source %q, got %v", sourceName, binding.TargetsByProgram)
	}
	if owner := binding.OwnerConfigDirBySourcePath[sourceName]; owner != repoDir {
		t.Fatalf("expected bound source owner %q, got %q", repoDir, owner)
	}
}

func testLintTarget(fsys vfs.FS, ownerDir string, filePath string) rslintconfig.DiscoveredLintTarget {
	filePath = tspath.NormalizePath(filePath)
	canonicalPath := filePath
	if realPath := fsys.Realpath(filePath); realPath != "" {
		canonicalPath = tspath.NormalizePath(realPath)
	}
	return rslintconfig.DiscoveredLintTarget{
		Path:            filePath,
		CanonicalPath:   canonicalPath,
		ConfigDirectory: tspath.NormalizePath(ownerDir),
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

func TestLoadProgramsDoesNotBorrowParentConfigProgram(t *testing.T) {
	rootDir := t.TempDir()
	childDir := filepath.Join(rootDir, "child")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"tsconfig.json":   `{"include":["child/target.ts"]}`,
		"child/target.ts": `export const value = 1;`,
	})

	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fsys)
	configMap := map[string]rslintconfig.RslintConfig{
		tspath.NormalizePath(rootDir):  projectConfig("./tsconfig.json"),
		tspath.NormalizePath(childDir): rslintconfig.RslintConfig{{}},
	}
	set, err := buildProjectsForConfigs(configMap, true, buildContext)
	if err != nil {
		t.Fatalf("buildProjectsForConfigs: %v", err)
	}
	if len(set.compilerPrograms) != 1 {
		t.Fatalf("expected only the root tsconfig Program, got %d", len(set.compilerPrograms))
	}

	targetPath := filepath.Join(childDir, "target.ts")
	plan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{testLintTarget(fsys, childDir, targetPath)}}
	binding, err := loadAPIForTest(set, plan, rootDir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}

	if len(binding.compilerPrograms) != 2 || len(binding.TargetsByProgram[0]) != 0 || len(binding.TargetsByProgram[1]) != 1 {
		t.Fatalf("expected target only in a source-only Program, got targets=%v", binding.TargetsByProgram)
	}
	sourceOnlySource := binding.TargetsByProgram[1][0]
	if owner := binding.OwnerConfigDirBySourcePath[sourceOnlySource]; owner != tspath.NormalizePath(childDir) {
		t.Fatalf("expected source-only owner %q, got %q", tspath.NormalizePath(childDir), owner)
	}
	if binding.Programs[1].CanProvideTypeChecker(binding.Programs[1].SourceFiles()[0]) {
		t.Fatal("child-owned target unexpectedly received type services")
	}
}

func TestTypeCheckDeduplicatesSyntaxFromSourceOnlyAndParentProgram(t *testing.T) {
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
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfigs(configMap, true, buildContext)
	if err != nil {
		t.Fatalf("buildProjectsForConfigs: %v", err)
	}
	targetPath := tspath.NormalizePath(filepath.Join(childDir, "target.ts"))
	plan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{testLintTarget(fsys, childDir, targetPath)}}
	binding, err := loadAPIForTest(set, plan, rootDir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.compilerPrograms) != 2 || len(binding.TargetsByProgram[1]) != 1 {
		t.Fatalf("child-owned target must remain source-only, got targets %v", binding.TargetsByProgram)
	}

	diagnostics := collectTargetSyntacticDiagnostics(binding.Programs, binding.TargetsByProgram, true, false)
	if len(diagnostics) != 1 {
		t.Fatalf("expected one malformed source-only lint target, got %v", diagnostics)
	}
	diagnostics = append(diagnostics, collectProgramTypeDiagnostics(t, binding.Programs)...)
	remapDiagnosticTargetPaths(diagnostics, binding.TargetPathBySourcePath)
	if len(diagnostics) < 2 {
		t.Fatalf("fixture must exercise both source-only syntax and parent Program type-check paths, got %+v", diagnostics)
	}

	diagnostics = deduplicateTypeScriptDiagnostics(diagnostics, fsys)
	if len(diagnostics) != 1 || diagnostics[0].RuleName != "TypeScript(TS1110)" {
		t.Fatalf("expected one TS1110 diagnostic after cross-phase dedupe, got %+v", diagnostics)
	}
	if diagnostics[0].Origin != rule.DiagnosticOriginTypeScript {
		t.Fatalf("deduplicated TypeScript diagnostic lost its origin: %+v", diagnostics[0])
	}
}

func TestBuildProjectsDeduplicatesSharedTsconfigAndRetainsOwners(t *testing.T) {
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
	set, err := buildProjectsForConfigs(map[string]rslintconfig.RslintConfig{
		rootKey:  projectConfig("./tsconfig.json"),
		childKey: projectConfig("../tsconfig.json"),
	}, true, newBuildContext(fsys))
	if err != nil {
		t.Fatalf("buildProjectsForConfigs: %v", err)
	}
	if len(set.compilerPrograms) != 1 || len(set.configOrders) != 1 {
		t.Fatalf("shared tsconfig must produce one Program, got programs=%d orders=%d", len(set.compilerPrograms), len(set.configOrders))
	}
	if order, ok := set.configOrders[0][rootKey]; !ok || order != 0 {
		t.Fatalf("missing root config association: %v", set.configOrders[0])
	}
	if order, ok := set.configOrders[0][childKey]; !ok || order != 0 {
		t.Fatalf("missing child config association: %v", set.configOrders[0])
	}
}

func TestBuildProjectsBoundsParallelBuildsAndPreservesOrder(t *testing.T) {
	rootDir := t.TempDir()
	const (
		testGOMAXPROCS = 9 // Deliberately exceeds the former fixed worker limit.
		configCount    = testGOMAXPROCS + 3
	)
	previousGOMAXPROCS := runtime.GOMAXPROCS(testGOMAXPROCS)
	t.Cleanup(func() {
		runtime.GOMAXPROCS(previousGOMAXPROCS)
	})

	files := make(map[string]string, configCount)
	projects := make([]string, 0, configCount)
	configPaths := make(map[string]struct{}, configCount)
	for index := range configCount {
		name := "tsconfig-" + strconv.Itoa(index) + ".json"
		files[name] = `{"compilerOptions":{"noLib":true},"files":["./shared.ts"]}`
		projects = append(projects, "./"+name)
		configPaths[tspath.NormalizePath(filepath.Join(rootDir, name))] = struct{}{}
	}
	files["shared.ts"] = "export const shared = true;\n"
	writeProgramTestFiles(t, rootDir, files)

	expectedConcurrency := testGOMAXPROCS
	fsys := &blockingProgramConfigFS{
		FS:         bundled.WrapFS(osvfs.FS()),
		paths:      configPaths,
		waitFor:    expectedConcurrency,
		allStarted: make(chan struct{}),
		release:    make(chan struct{}),
	}
	type result struct {
		set ProjectSet
		err error
	}
	done := make(chan result, 1)
	go func() {
		set, err := buildProjectsForConfig(
			tspath.NormalizePath(rootDir),
			projectConfig(projects...),
			false,
			newBuildContext(fsys),
		)
		done <- result{set: set, err: err}
	}()

	select {
	case <-fsys.allStarted:
		close(fsys.release)
	case <-time.After(5 * time.Second):
		close(fsys.release)
		t.Fatalf("Program builds did not reach expected concurrency %d", expectedConcurrency)
	}
	got := <-done
	if got.err != nil {
		t.Fatalf("buildProjectsForConfig: %v", got.err)
	}
	if got := fsys.peakConcurrency(); got != expectedConcurrency {
		t.Fatalf("peak Program build concurrency = %d, want %d", got, expectedConcurrency)
	}
	if len(got.set.compilerPrograms) != configCount {
		t.Fatalf("Programs = %d, want %d", len(got.set.compilerPrograms), configCount)
	}
	for index, program := range got.set.compilerPrograms {
		want := tspath.ResolvePath(rootDir, projects[index])
		if got := tspath.NormalizePath(program.Options().ConfigFilePath); got != want {
			t.Fatalf("Program %d config path = %q, want %q", index, got, want)
		}
	}
}

func TestBuildProjectsSingleThreadedBuildsSerially(t *testing.T) {
	rootDir := t.TempDir()
	files := map[string]string{
		"tsconfig-a.json": `{"compilerOptions":{"noLib":true},"files":[]}`,
		"tsconfig-b.json": `{"compilerOptions":{"noLib":true},"files":[]}`,
	}
	writeProgramTestFiles(t, rootDir, files)
	configPaths := make(map[string]struct{}, len(files))
	for name := range files {
		configPaths[tspath.NormalizePath(filepath.Join(rootDir, name))] = struct{}{}
	}
	fsys := &blockingProgramConfigFS{
		FS:         bundled.WrapFS(osvfs.FS()),
		paths:      configPaths,
		waitFor:    1,
		allStarted: make(chan struct{}),
		release:    make(chan struct{}),
	}
	done := make(chan error, 1)
	go func() {
		_, err := buildProjectsForConfig(
			tspath.NormalizePath(rootDir),
			projectConfig("./tsconfig-a.json", "./tsconfig-b.json"),
			true,
			newBuildContext(fsys),
		)
		done <- err
	}()
	<-fsys.allStarted
	close(fsys.release)
	if err := <-done; err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	if got := fsys.peakConcurrency(); got != 1 {
		t.Fatalf("--singleThreaded peak Program build concurrency = %d, want 1", got)
	}
}

func TestExecuteProjectPlanScopesConcurrentProgramQueries(t *testing.T) {
	t.Setenv("RSLINT_DISABLE_PROGRAM_METADATA_CACHE", "")
	previousGOMAXPROCS := runtime.GOMAXPROCS(2)
	t.Cleanup(func() {
		runtime.GOMAXPROCS(previousGOMAXPROCS)
	})

	rootDir := tspath.NormalizePath(t.TempDir())
	tests := []struct {
		name           string
		configCount    int
		singleThreaded bool
		wantDerivedFS  bool
	}{
		{name: "one Program", configCount: 1},
		{name: "parallel Programs", configCount: 2, wantDerivedFS: true},
		{name: "single-threaded Programs", configCount: 2, singleThreaded: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := projectPlan{specs: make([]projectSpec, test.configCount)}
			for index := range plan.specs {
				plan.specs[index] = projectSpec{
					tsconfigPath: tspath.ResolvePath(rootDir, "missing-"+strconv.Itoa(index)+".json"),
					programCwd:   rootDir,
				}
			}
			buildContext := newBuildContext(bundled.WrapFS(osvfs.FS()))
			_, _ = executeProjectPlanForTest(plan, test.singleThreaded, buildContext)

			gotDerivedFS := buildContext.newCompilerHostWithCache(rootDir).FS() != buildContext.FS()
			if gotDerivedFS != test.wantDerivedFS {
				t.Fatalf("compiler host derived FS = %t, want %t", gotDerivedFS, test.wantDerivedFS)
			}
		})
	}
}

func TestExecuteProjectPlanPreservesFirstErrorPrecedence(t *testing.T) {
	rootDir := tspath.NormalizePath(t.TempDir())
	first := tspath.ResolvePath(rootDir, "missing-first.json")
	second := tspath.ResolvePath(rootDir, "missing-second.json")
	plan := projectPlan{
		specs: []projectSpec{
			{tsconfigPath: first, programCwd: rootDir},
			{tsconfigPath: second, programCwd: rootDir},
		},
		terminalErr: os.ErrInvalid,
	}
	_, err := executeProjectPlanForTest(
		plan,
		false,
		newBuildContext(bundled.WrapFS(osvfs.FS())),
	)
	if err == nil || !strings.Contains(err.Error(), first) {
		t.Fatalf("error = %v, want first Program path %q", err, first)
	}
	if strings.Contains(err.Error(), second) || strings.Contains(err.Error(), os.ErrInvalid.Error()) {
		t.Fatalf("later error won over the first Program failure: %v", err)
	}
}

func TestBuildProjectsPreservesSymlinkedTsconfigBase(t *testing.T) {
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
	set, err := buildProjectsForConfigs(map[string]rslintconfig.RslintConfig{
		aliasDir: projectConfig("./tsconfig.json"),
		realDir:  projectConfig("./tsconfig.json"),
	}, true, newBuildContext(fsys))
	if err != nil {
		t.Fatalf("buildProjectsForConfigs: %v", err)
	}
	if len(set.compilerPrograms) != 2 {
		t.Fatalf("distinct declared tsconfig paths must produce two Programs, got %d", len(set.compilerPrograms))
	}
	programByConfigPath := make(map[string]*compiler.Program, len(set.compilerPrograms))
	for _, program := range set.compilerPrograms {
		programByConfigPath[exactPathID(program.Options().ConfigFilePath)] = program
	}
	aliasConfigPath := tspath.ResolvePath(aliasDir, "tsconfig.json")
	realConfigPath := tspath.ResolvePath(realDir, "tsconfig.json")
	aliasProgram := programByConfigPath[exactPathID(aliasConfigPath)]
	realProgram := programByConfigPath[exactPathID(realConfigPath)]
	if aliasProgram == nil || realProgram == nil {
		t.Fatalf("missing lexical tsconfig Programs: %v", programByConfigPath)
		return
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

func TestLoadProgramsUsesGoverningConfigProjectOrder(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"shared.ts":       `export const value = 1;`,
		"tsconfig-a.json": `{"files":["shared.ts"]}`,
		"tsconfig-b.json": `{"files":["shared.ts"]}`,
	})

	dir = tspath.NormalizePath(dir)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfig(
		dir,
		projectConfig("./tsconfig-a.json", "./tsconfig-b.json"),
		true,
		buildContext,
	)
	if err != nil || len(set.compilerPrograms) != 2 {
		t.Fatalf("expected two ordered Programs, err=%v programs=%d", err, len(set.compilerPrograms))
	}

	targetPath := filepath.Join(dir, "shared.ts")
	plan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{testLintTarget(fsys, dir, targetPath)}}
	binding, err := loadAPIForTest(set, plan, dir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.TargetsByProgram[0]) != 1 || len(binding.TargetsByProgram[1]) != 0 {
		t.Fatalf("overlapping target must bind to the first declared project, got %v", binding.TargetsByProgram)
	}
}

func TestLoadProgramsRecomputesProgramMembershipAfterImportGraphChange(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"main.ts":       `import "./extra";`,
		"extra.ts":      `export const value = 1;`,
		"tsconfig.json": `{"files":["main.ts"]}`,
	})

	dir = tspath.NormalizePath(dir)
	config := projectConfig("./tsconfig.json")
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfig(dir, config, true, buildContext)
	if err != nil {
		t.Fatalf("initial buildProjectsForConfig: %v", err)
	}
	plan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{
		testLintTarget(fsys, dir, filepath.Join(dir, "extra.ts")),
	}}
	initial, err := loadAPIForTest(set, plan, dir, buildContext, true)
	if err != nil {
		t.Fatalf("initial loadAPIForTest: %v", err)
	}
	if len(initial.Programs) != 1 || len(initial.TargetsByProgram[0]) != 1 {
		t.Fatalf("imported target should initially use the configured Program, got targets=%v", initial.TargetsByProgram)
	}

	if err := os.WriteFile(filepath.Join(dir, "main.ts"), []byte(`export const main = 1;`), 0644); err != nil {
		t.Fatalf("rewrite main.ts: %v", err)
	}
	// Production fix passes reuse the run-scoped filesystem and parse caches.
	// Replacing the source-snapshot generation before rebuilding makes the new
	// text/hash visible while retaining content-keyed AST entries for unchanged
	// files.
	buildContext.invalidateSourceSnapshots()
	rebuilt, err := buildProjectsForConfig(dir, config, true, buildContext)
	if err != nil {
		t.Fatalf("rebuilt buildProjectsForConfig: %v", err)
	}
	afterFix, err := loadAPIForTest(rebuilt, plan, dir, buildContext, true)
	if err != nil {
		t.Fatalf("rebuilt loadAPIForTest: %v", err)
	}
	if len(afterFix.Programs) != 2 || len(afterFix.TargetsByProgram[1]) != 1 {
		t.Fatalf("target must move to a source-only Program after its importing edge is removed, got targets=%v", afterFix.TargetsByProgram)
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

	plan, err := resolveTargetPlanForTest(
		nil,
		rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
		dir,
		nil,
		fsys,
		nil,
		[]string{dir},
		true,
	)
	if err != nil {
		t.Fatalf("resolveTargetPlanForTest: %v", err)
	}
	if len(plan.Targets) != 2 {
		t.Fatalf("targets = %v, want two files", plan.Targets)
	}
	if counter.callCount(fileA) != 0 || counter.callCount(fileB) != 0 {
		t.Fatalf("regular targets performed realpath IO: a=%d b=%d", counter.callCount(fileA), counter.callCount(fileB))
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

	_, err := resolveTargetPlanForTest(
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
		return
	}
	if !strings.Contains(err.Error(), "resolve to the same file") || !strings.Contains(err.Error(), ownerA) || !strings.Contains(err.Error(), ownerB) {
		t.Fatalf("unexpected ownership conflict error: %v", err)
	}
}

func TestLintTargetPlanActiveConfigsSelectsOnlyGoverningConfigs(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/repo/a": {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
		"/repo/b": {{LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{Project: []string{"./missing.json"}},
		}}},
	}
	active := activeConfigsForTest(configMap, rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{{
		Path:            "/repo/a/index.ts",
		CanonicalPath:   "/repo/a/index.ts",
		ConfigDirectory: "/repo/a",
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
	plan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{{
		Path:            tspath.ResolvePath(activeDir, "index.ts"),
		CanonicalPath:   tspath.ResolvePath(activeDir, "index.ts"),
		ConfigDirectory: activeDir,
	}}}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	activeConfigMap := activeConfigsForTest(configMap, plan)
	set, err := buildProjectsForConfigs(activeConfigMap, true, newBuildContext(fsys))
	if err != nil || len(set.compilerPrograms) != 1 {
		t.Fatalf("plain lint should build only the active config Program: programs=%d err=%v", len(set.compilerPrograms), err)
	}
	if _, err := buildProjectsForConfigs(configMap, true, newBuildContext(fsys)); err == nil || !strings.Contains(err.Error(), "missing.json") {
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

	t.Run("same canonical path is deduplicated", func(t *testing.T) {
		fsys := &canonicalIdentityTestFS{
			FS: osvfs.FS(),
			realPaths: map[string]string{
				upper: "C:/Repo/Src/A.ts",
				lower: "C:/Repo/Src/A.ts",
			},
		}
		plan, err := resolveTargetPlanForTest(nil, config, configDir, nil, fsys, []string{upper, lower}, nil, true)
		if err != nil || len(plan.Targets) != 1 {
			t.Fatalf("same canonical target should be deduplicated: targets=%v err=%v", plan.Targets, err)
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
		plan, err := resolveTargetPlanForTest(nil, config, configDir, nil, fsys, []string{upper, lower}, nil, true)
		if err != nil || len(plan.Targets) != 2 {
			t.Fatalf("global case behavior must not merge distinct physical paths: targets=%v err=%v", plan.Targets, err)
		}
	})
}

func TestLoadProgramsRejectsCaseFoldedSourceWithDifferentCanonicalIdentity(t *testing.T) {
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

	set := ProjectSet{
		compilerPrograms: []*compiler.Program{program},
		configOrders:     []configOrders{{configDir: 0}},
	}
	plan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{{
		Path:            lower,
		CanonicalPath:   lower,
		ConfigDirectory: configDir,
	}}}
	binding, err := loadAPIForTest(set, plan, configDir, newBuildContext(fsys), true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.compilerPrograms) != 2 || len(binding.TargetsByProgram[0]) != 0 {
		t.Fatalf("lower-case target must not bind to the distinct upper-case source: %+v", binding.TargetsByProgram)
	}
	if got := binding.TargetsByProgram[1]; len(got) != 1 || got[0] != lower {
		t.Fatalf("lower-case target must bind to its exact compatibility source, got %v", got)
	}
}

func TestLoadProgramsSplitsCompatibilityProgramsForCaseFoldedPathCollisions(t *testing.T) {
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
	plan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{
		{Path: upper, CanonicalPath: upper, ConfigDirectory: configDir},
		{Path: lower, CanonicalPath: lower, ConfigDirectory: configDir},
	}}
	binding, err := loadAPIForTest(ProjectSet{}, plan, configDir, newBuildContext(fsys), true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.compilerPrograms) != 2 || len(binding.TargetsByProgram) != 2 {
		t.Fatalf("case-folded root names require separate compatibility Programs, got %d", len(binding.compilerPrograms))
	}
	bound := []string{binding.TargetsByProgram[0][0], binding.TargetsByProgram[1][0]}
	slices.Sort(bound)
	want := []string{upper, lower}
	slices.Sort(want)
	if !slices.Equal(bound, want) {
		t.Fatalf("compatibility Programs must preserve both exact source identities: got %v want %v", bound, want)
	}
}
