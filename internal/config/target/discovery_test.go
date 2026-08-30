package target

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rules"
	"gotest.tools/v3/assert"
)

func strconvI(i int) string { return strconv.Itoa(i) }

type discoveryMockFS struct {
	vfs.FS
	entries         map[string]vfs.Entries
	files           map[string]string
	resolvedPaths   map[string]string
	caseSensitiveFS bool
}

func (m *discoveryMockFS) ReadFile(path string) (string, bool) {
	content, ok := m.files[path]
	return content, ok
}

func (m *discoveryMockFS) GetAccessibleEntries(path string) vfs.Entries {
	return m.entries[path]
}

func (m *discoveryMockFS) Realpath(path string) string {
	if realpath, ok := m.resolvedPaths[path]; ok {
		return realpath
	}
	return path
}

func (m *discoveryMockFS) UseCaseSensitiveFileNames() bool {
	return m.caseSensitiveFS
}

func setupDiscoveryContentFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	return tspath.NormalizePath(tmpDir)
}

// setupDiscoveryFixture creates a temp dir with the given file paths and returns
// the normalized configDir and a map of short name → normalized absolute path.
func setupDiscoveryFixture(t *testing.T, files []string) (string, map[string]string) {
	t.Helper()
	tmpDir := t.TempDir()
	paths := make(map[string]string, len(files))
	for _, name := range files {
		fp := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(fp), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(fp, []byte("// "+name), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
		paths[name] = tspath.NormalizePath(fp)
	}
	return tspath.NormalizePath(tmpDir), paths
}

func TestDiscoverFilesOutsidePrograms_Basic(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"scripts/b.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	// src/a.ts is in the program, scripts/b.ts is not
	programFiles := map[string]struct{}{
		paths["src/a.ts"]: {},
	}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	assert.Assert(t, gapFiles != nil, "should not be nil when config has files")
	assert.Equal(t, len(gapFiles), 1)
	assert.Equal(t, gapFiles[0], paths["scripts/b.ts"])
}

func TestDiscoverFilesOutsidePrograms_GlobalIgnoresExclude(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"scripts/b.ts",
	})

	config := rslintconfig.RslintConfig{
		// Global ignore
		{Ignores: []string{"scripts/**"}},
		// rslintconfig.Rules entry
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{
		paths["src/a.ts"]: {},
	}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	// scripts/b.ts should be excluded by global ignores
	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 0)
}

func TestDiscoverFilesOutsidePrograms_ProgramFilesSkipped(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/b.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	// Both files are in the program
	programFiles := map[string]struct{}{
		paths["src/a.ts"]: {},
		paths["src/b.ts"]: {},
	}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 0)
}

func TestDiscoverFilesOutsidePrograms_EntryIgnoreDoesNotRemoveTarget(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"test/b.ts",
	})

	config := rslintconfig.RslintConfig{
		// Entry with files AND entry-level ignores
		{
			Files:   []string{"**/*.ts"},
			Ignores: []string{"test/**"},
			Rules:   rslintconfig.Rules{"test-rule": "error"},
		},
	}

	programFiles := map[string]struct{}{
		paths["src/a.ts"]: {},
	}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	// The default TypeScript baseline independently selects both files, so the
	// entry-level ignore only prevents this entry from contributing rules.
	assert.Assert(t, gapFiles != nil)
	assert.DeepEqual(t, gapFiles, []string{paths["test/b.ts"]})
}

func TestDiscoverLintFiles_EntryIgnorePreventsSelectorContribution(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/ignored.JS",
		"src/included.JS",
	})
	config := rslintconfig.RslintConfig{
		{Rules: rslintconfig.Rules{"no-debugger": "error"}},
		{
			Files:   []string{"**/*.JS"},
			Ignores: []string{"**/ignored.JS"},
			Rules:   rslintconfig.Rules{"no-console": "error"},
		},
	}

	targets := discoverLintFiles(config, configDir, osvfs.FS(), nil, nil, true)
	assert.DeepEqual(t, targets, []string{paths["src/included.JS"]})
	if merged := config.GetConfigForFile(paths["src/ignored.JS"], configDir); merged != nil {
		t.Fatalf("locally ignored selector made the non-default extension configurable: %#v", merged)
	}
	if merged := config.GetConfigForFile(paths["src/included.JS"], configDir); merged == nil || merged.Rules["no-debugger"] == nil {
		t.Fatalf("matching selector should make unscoped entries apply: %#v", merged)
	}
}

func TestDiscoverLintFiles_DefaultBaselineIsIndependentOfConfigEntries(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.js",
		"src/b.mjs",
		"src/c.cjs",
		"src/d.jsx",
		"src/e.ts",
		"src/f.tsx",
		"src/g.mts",
		"src/h.cts",
		"src/G.JS",
		"src/styles.css",
	})

	tests := []struct {
		name   string
		config rslintconfig.RslintConfig
	}{
		{name: "empty config"},
		{
			name: "global ignore only",
			config: rslintconfig.RslintConfig{
				{Ignores: []string{"generated/**"}},
			},
		},
		{
			name: "explicit TS entry does not remove defaults",
			config: rslintconfig.RslintConfig{
				{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
			},
		},
	}

	expected := []string{
		paths["src/a.js"],
		paths["src/b.mjs"],
		paths["src/c.cjs"],
		paths["src/d.jsx"],
		paths["src/e.ts"],
		paths["src/f.tsx"],
		paths["src/g.mts"],
		paths["src/h.cts"],
	}
	sort.Strings(expected)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets := discoverLintFiles(tt.config, configDir, osvfs.FS(), nil, nil, true)
			assert.DeepEqual(t, targets, expected)
		})
	}
}

func TestDiscoverLintFiles_PreservesUNCRoot(t *testing.T) {
	const configDir = "//server/share/repo"
	mock := &discoveryMockFS{
		FS: osvfs.FS(),
		entries: map[string]vfs.Entries{
			configDir: {
				Directories: []string{"src"},
				Symlinks:    map[string]struct{}{},
			},
			configDir + "/src": {
				Files:    []string{"a.ts"},
				Symlinks: map[string]struct{}{},
			},
		},
		files:         map[string]string{},
		resolvedPaths: map[string]string{},
	}

	targets := discoverLintFiles(nil, configDir, mock, nil, nil, true)
	assert.DeepEqual(t, targets, []string{"//server/share/repo/src/a.ts"})
}

type caseInsensitiveDiscoveryFS struct {
	vfs.FS
	files map[string]bool
}

func (f *caseInsensitiveDiscoveryFS) UseCaseSensitiveFileNames() bool { return false }
func (f *caseInsensitiveDiscoveryFS) FileExists(filePath string) bool {
	return f.files[strings.ToLower(tspath.NormalizePath(filePath))]
}
func (f *caseInsensitiveDiscoveryFS) Realpath(filePath string) string {
	return strings.ToLower(tspath.NormalizePath(filePath))
}

type fileExistsCountingFS struct {
	vfs.FS
	calls atomic.Int32
}

func (f *fileExistsCountingFS) FileExists(filePath string) bool {
	f.calls.Add(1)
	return f.FS.FileExists(filePath)
}

type realpathCountingFS struct {
	vfs.FS
	mu    sync.Mutex
	calls map[string]int
}

func (f *realpathCountingFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	f.mu.Lock()
	f.calls[filePath]++
	f.mu.Unlock()
	return f.FS.Realpath(filePath)
}

func (f *realpathCountingFS) callCount(filePath string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[tspath.NormalizePath(filePath)]
}

func TestDiscoverLintFiles_PreservesDistinctLexicalPathCasing(t *testing.T) {
	fsys := &caseInsensitiveDiscoveryFS{
		FS: osvfs.FS(),
		files: map[string]bool{
			"c:/repo/src/a.ts": true,
		},
	}
	targets := discoverLintFiles(
		nil,
		"C:/Repo",
		fsys,
		[]string{"C:/Repo/Src/A.ts", "c:/repo/src/a.ts"},
		nil,
		true,
	)
	assert.DeepEqual(t, targets, []string{"C:/Repo/Src/A.ts", "c:/repo/src/a.ts"})
}

func TestDiscoverLintFiles_ExplicitPatternCanExtendCaseSensitiveBaseline(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{"src/A.JS", "src/b.js"})

	withoutExplicitPattern := discoverLintFiles(nil, configDir, osvfs.FS(), nil, nil, true)
	assert.DeepEqual(t, withoutExplicitPattern, []string{paths["src/b.js"]})

	config := rslintconfig.RslintConfig{{Files: []string{"**/*.JS"}}}
	withExplicitPattern := discoverLintFiles(config, configDir, osvfs.FS(), nil, nil, true)
	expected := []string{paths["src/A.JS"], paths["src/b.js"]}
	sort.Strings(expected)
	assert.DeepEqual(t, withExplicitPattern, expected)
}

func TestDiscoverLintFiles_FilesAndGroupAppliesCandidatePostFilter(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/A.JS",
		"src/A.test.JS",
		"other/B.JS",
		"src/default.ts",
	})
	config := rslintconfig.RslintConfig{{
		FilePatternGroups: [][]string{
			{"src/**", "**/*.JS", "!**/*.test.JS"},
		},
	}}

	targets := discoverLintFiles(config, configDir, osvfs.FS(), nil, nil, true)
	expected := []string{paths["src/A.JS"], paths["src/default.ts"]}
	sort.Strings(expected)
	assert.DeepEqual(t, targets, expected)
}

func TestDiscoverLintFiles_EmptyFilesAndGroupMatchesSupportedBaseline(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.js",
		"src/b.ts",
		"src/readme.md",
	})
	config := rslintconfig.RslintConfig{{
		FilePatternGroups: [][]string{{}},
		Rules:             rslintconfig.Rules{"test-rule": "error"},
	}}

	targets := discoverLintFiles(config, configDir, osvfs.FS(), nil, nil, true)
	expected := []string{paths["src/a.js"], paths["src/b.ts"]}
	sort.Strings(expected)
	assert.DeepEqual(t, targets, expected)
}

func TestDiscoverFilesOutsidePrograms_NoFilesField_UsesDefaultExtensions(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/b.jsx",
		"src/c.cjs",
		"src/d.cts",
		"src/styles.css",
	})

	config := rslintconfig.RslintConfig{
		{Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, false)

	expected := []string{paths["src/a.ts"], paths["src/b.jsx"], paths["src/c.cjs"], paths["src/d.cts"]}
	sort.Strings(expected)
	assert.DeepEqual(t, gapFiles, expected)
}

func TestDiscoverFilesOutsidePrograms_AllowDirsScope(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"scripts/b.ts",
		"tools/c.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	// Only allow scripts/ directory
	scriptsDir := tspath.NormalizePath(filepath.Join(configDir, "scripts"))
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, []string{scriptsDir}, false)

	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 1)
	assert.Equal(t, gapFiles[0], paths["scripts/b.ts"])
}

func TestDiscoverLintFiles_AllowDirsStartsWalkAtScopedRoot(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"packages/app/src/a.ts",
		"packages/other/src/b.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}
	appDir := tspath.NormalizePath(filepath.Join(configDir, "packages/app"))
	spy := &spyFS{FS: osvfs.FS()}

	targets := discoverLintFiles(config, configDir, spy, nil, []string{appDir}, true)

	assert.DeepEqual(t, targets, []string{paths["packages/app/src/a.ts"]})

	rootDir := tspath.NormalizePath(configDir)
	otherDir := tspath.NormalizePath(filepath.Join(configDir, "packages/other"))
	for _, accessed := range spy.snapshotAccessedDirs() {
		if accessed == rootDir {
			t.Fatalf("scoped walk should not open config root %s", rootDir)
		}
		if strings.HasPrefix(accessed, otherDir) {
			t.Fatalf("scoped walk should not enter sibling %s", otherDir)
		}
	}
}

func TestDiscoverLintFiles_AllowDirsSkipsDefaultExcludedRoot(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"node_modules/pkg/a.ts",
		"src/a.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}
	nodeModulesDir := tspath.NormalizePath(filepath.Join(configDir, "node_modules"))
	spy := &spyFS{FS: osvfs.FS()}

	targets := discoverLintFiles(config, configDir, spy, nil, []string{nodeModulesDir}, true)

	assert.DeepEqual(t, targets, []string{})
	for _, accessed := range spy.snapshotAccessedDirs() {
		if strings.Contains(accessed, "node_modules") {
			t.Fatalf("default-excluded initial root should not be entered; accessed=%v", spy.snapshotAccessedDirs())
		}
	}
}

func TestDiscoverLintFiles_AllowDirsSkipsGloballyIgnoredRoot(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"dist/a.ts",
		"src/a.ts",
	})

	config := rslintconfig.RslintConfig{
		{Ignores: []string{"dist/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}
	distDir := tspath.NormalizePath(filepath.Join(configDir, "dist"))
	spy := &spyFS{FS: osvfs.FS()}

	targets := discoverLintFiles(config, configDir, spy, nil, []string{distDir}, true)

	assert.DeepEqual(t, targets, []string{})
	for _, accessed := range spy.snapshotAccessedDirs() {
		if accessed == distDir || strings.HasPrefix(accessed, distDir+"/") {
			t.Fatalf("globally ignored initial root should not be entered; accessed=%v", spy.snapshotAccessedDirs())
		}
	}
}

func TestDiscoverLintFiles_AllowDirsKeepsIgnoredRootWithNegation(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"dist/drop.ts",
		"dist/keep.ts",
	})

	config := rslintconfig.RslintConfig{
		{Ignores: []string{"dist/**/*", "!dist/keep.ts"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}
	distDir := tspath.NormalizePath(filepath.Join(configDir, "dist"))
	spy := &spyFS{FS: osvfs.FS()}

	targets := discoverLintFiles(config, configDir, spy, nil, []string{distDir}, true)

	assert.DeepEqual(t, targets, []string{paths["dist/keep.ts"]})
	enteredDist := false
	for _, accessed := range spy.snapshotAccessedDirs() {
		if accessed == distDir {
			enteredDist = true
			break
		}
	}
	assert.Assert(t, enteredDist, "negated ignored root must still be entered")
}

func TestDiscoverLintFiles_EmptyAllowDirsDoesNotWalk(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/a.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}
	spy := &spyFS{FS: osvfs.FS()}

	targets := discoverLintFiles(config, configDir, spy, nil, []string{}, true)

	assert.Equal(t, len(targets), 0)
	assert.Equal(t, len(spy.snapshotAccessedDirs()), 0)
}

func TestDiscoverLintFiles_ExplicitFileSkipsNestedDefaultExcludedDir(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"packages/app/node_modules/pkg/a.ts",
		"src/a.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	targets := discoverLintFiles(
		config,
		configDir,
		osvfs.FS(),
		[]string{paths["packages/app/node_modules/pkg/a.ts"]},
		nil,
		true,
	)

	assert.DeepEqual(t, targets, []string{})
}

func TestDiscoverLintTargetsFromRootExplicitFileUsesScanRootDefaultExcludes(t *testing.T) {
	root := t.TempDir()
	scanRoot := tspath.NormalizePath(filepath.Join(root, "workspace"))
	configDir := tspath.NormalizePath(filepath.Join(root, "physical", "package"))
	aliasDir := tspath.CombinePaths(scanRoot, "node_modules", "package")
	for _, directory := range []string{scanRoot, configDir, tspath.GetDirectoryPath(aliasDir)} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	physicalTarget := tspath.CombinePaths(configDir, "index.ts")
	if err := os.WriteFile(physicalTarget, []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(configDir, aliasDir); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	aliasTarget := tspath.CombinePaths(aliasDir, "index.ts")

	targets := discoverLintTargetsFromRoot(
		rslintconfig.RslintConfig{{
			Files: []string{"**/*.ts"},
			Rules: rslintconfig.Rules{"test-rule": "error"},
		}},
		configDir,
		scanRoot,
		osvfs.FS(),
		[]string{aliasTarget},
		nil,
		true,
	)
	assert.DeepEqual(t, targets, []File{})
}

func TestDiscoverLintFiles_OverlappingAllowDirsWalkChildOnce(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"packages/app/src/a.ts",
		"packages/app/src/nested/b.ts",
		"packages/other/src/c.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}
	appDir := tspath.NormalizePath(filepath.Join(configDir, "packages/app"))
	srcDir := tspath.NormalizePath(filepath.Join(configDir, "packages/app/src"))
	spy := &spyFS{FS: osvfs.FS()}

	targets := discoverLintFiles(config, configDir, spy, nil, []string{appDir, srcDir}, true)

	expected := []string{paths["packages/app/src/a.ts"], paths["packages/app/src/nested/b.ts"]}
	assert.DeepEqual(t, targets, expected)

	srcAccesses := 0
	for _, accessed := range spy.snapshotAccessedDirs() {
		if accessed == srcDir {
			srcAccesses++
		}
	}
	assert.Equal(t, srcAccesses, 1, "overlapping allowDirs should not walk child roots twice")
}

func TestDiscoverLintTargetsFromRoot_WalksEveryRequestedDirectory(t *testing.T) {
	root := t.TempDir()
	configDir := tspath.NormalizePath(filepath.Join(root, "config"))
	scanRoot := tspath.NormalizePath(filepath.Join(root, "workspace"))
	firstDir := tspath.NormalizePath(filepath.Join(root, "first"))
	secondDir := tspath.NormalizePath(filepath.Join(root, "second"))
	for _, directory := range []string{configDir, scanRoot, firstDir, secondDir} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	firstFile := tspath.CombinePaths(firstDir, "a.ts")
	secondFile := tspath.CombinePaths(secondDir, "b.ts")
	for _, filePath := range []string{firstFile, secondFile} {
		if err := os.WriteFile(filePath, []byte("export {};\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	targets := discoverLintTargetsFromRoot(
		rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"test-rule": "error"}}},
		configDir,
		scanRoot,
		osvfs.FS(),
		nil,
		[]string{firstDir, secondDir},
		true,
	)
	assert.DeepEqual(t, targets, []File{
		{PathIdentity: rslintconfig.PathIdentity{Path: firstFile, CanonicalPath: tspath.NormalizePath(osvfs.FS().Realpath(firstFile)), CanonicalParentPath: tspath.NormalizePath(osvfs.FS().Realpath(firstDir))}, ConfigDirectory: configDir},
		{PathIdentity: rslintconfig.PathIdentity{Path: secondFile, CanonicalPath: tspath.NormalizePath(osvfs.FS().Realpath(secondFile)), CanonicalParentPath: tspath.NormalizePath(osvfs.FS().Realpath(secondDir))}, ConfigDirectory: configDir},
	})
}

func TestDiscoverLintTargetsFromRoot_PreservesDistinctLexicalDirectorySelectors(t *testing.T) {
	configDir := tspath.NormalizePath(t.TempDir())
	physicalDir := tspath.NormalizePath(t.TempDir())
	physicalChild := tspath.CombinePaths(physicalDir, "sub")
	if err := os.MkdirAll(physicalChild, 0o755); err != nil {
		t.Fatal(err)
	}
	physicalFile := tspath.CombinePaths(physicalChild, "x.TS")
	if err := os.WriteFile(physicalFile, []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	aliasA := tspath.CombinePaths(configDir, "a")
	aliasB := tspath.CombinePaths(configDir, "b")
	if err := os.Symlink(physicalDir, aliasA); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	if err := os.Symlink(physicalChild, aliasB); err != nil {
		t.Skipf("second directory symlink unavailable: %v", err)
	}

	plan, err := Resolve(Request{
		Config: rslintconfig.RslintConfig{{
			Files: []string{"b/*.TS"},
			Rules: rslintconfig.Rules{"rule": "error"},
		}},
		ConfigDirectory: configDir,
		ScanRoot:        configDir,
		FS:              osvfs.FS(),
		Directories:     []string{aliasA, aliasB},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantPath := tspath.CombinePaths(aliasB, "x.TS")
	if len(plan.Files) != 1 || plan.Files[0].Path != wantPath {
		t.Fatalf("targets = %+v, want only %q", plan.Files, wantPath)
	}
}

func TestDiscoverLintTargetsFromRoot_DirectoryAliasDoesNotRewriteSiblingScope(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	physicalDir := tspath.CombinePaths(root, "real")
	if err := os.MkdirAll(physicalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	physicalFile := tspath.CombinePaths(physicalDir, "x.TS")
	if err := os.WriteFile(physicalFile, []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	aliasDir := tspath.CombinePaths(root, "alias-b")
	if err := os.Symlink(physicalDir, aliasDir); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}

	plan, err := Resolve(Request{
		Config: rslintconfig.RslintConfig{{
			Files: []string{"real/*.TS"},
			Rules: rslintconfig.Rules{"rule": "error"},
		}},
		ConfigDirectory: root,
		ScanRoot:        root,
		FS:              osvfs.FS(),
		Directories:     []string{root, aliasDir},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 1 || plan.Files[0].Path != physicalFile {
		t.Fatalf("targets = %+v, want only real-directory spelling %q", plan.Files, physicalFile)
	}
}

func TestDiscoverLintTargetsMultiConfig_PreservesProjectedCanonicalIdentity(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	physicalConfigDir := tspath.CombinePaths(root, "real")
	physicalTargetDir := tspath.CombinePaths(physicalConfigDir, "sub")
	if err := os.MkdirAll(physicalTargetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	physicalFile := tspath.CombinePaths(physicalTargetDir, "x.ts")
	if err := os.WriteFile(physicalFile, []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	configAlias := tspath.CombinePaths(root, "config-alias")
	targetAlias := tspath.CombinePaths(root, "target-alias")
	if err := os.Symlink(physicalConfigDir, configAlias); err != nil {
		t.Skipf("config directory symlink unavailable: %v", err)
	}
	if err := os.Symlink(physicalTargetDir, targetAlias); err != nil {
		t.Skipf("target directory symlink unavailable: %v", err)
	}

	targets := discoverLintTargetsMultiConfig(
		map[string]rslintconfig.RslintConfig{
			configAlias: {{Rules: rslintconfig.Rules{"rule": "error"}}},
		},
		nil,
		osvfs.FS(),
		nil,
		[]string{targetAlias},
		true,
	)
	wantPath := tspath.CombinePaths(targetAlias, "x.ts")
	if len(targets) != 1 || targets[0].Path != wantPath ||
		targets[0].CanonicalPath != tspath.NormalizePath(osvfs.FS().Realpath(physicalFile)) ||
		targets[0].ConfigDirectory != configAlias {
		t.Fatalf("projected target identity = %+v", targets)
	}
}

func TestDiscoverLintTargetsFromRoot_SharedAliasAncestorKeepsRelativeConfigMeaning(t *testing.T) {
	physicalRoot := tspath.NormalizePath(t.TempDir())
	physicalConfigDir := tspath.CombinePaths(physicalRoot, "config")
	physicalWorkspace := tspath.CombinePaths(physicalRoot, "workspace")
	for _, directory := range []string{physicalConfigDir, physicalWorkspace} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	visible := tspath.CombinePaths(physicalWorkspace, "visible.ts")
	ignored := tspath.CombinePaths(physicalWorkspace, "ignored.ts")
	for _, filePath := range []string{visible, ignored} {
		if err := os.WriteFile(filePath, []byte("export {};\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	aliasRoot := tspath.CombinePaths(tspath.NormalizePath(t.TempDir()), "root-alias")
	if err := os.Symlink(physicalRoot, aliasRoot); err != nil {
		t.Skipf("root symlink unavailable: %v", err)
	}
	configDir := tspath.CombinePaths(aliasRoot, "config")
	entries := rslintconfig.RslintConfig{
		{Ignores: []string{"../workspace/ignored.ts"}},
		{Files: []string{"../workspace/*.ts"}, Rules: rslintconfig.Rules{"rule": "error"}},
	}
	targets := discoverLintTargetsFromRoot(
		entries,
		configDir,
		physicalWorkspace,
		osvfs.FS(),
		nil,
		[]string{physicalWorkspace},
		true,
	)
	if len(targets) != 1 || targets[0].Path != visible {
		t.Fatalf("shared-alias targets = %+v, want only %q", targets, visible)
	}
	merged := rslintconfig.NewFileConfigResolverWithFS(entries, configDir, osvfs.FS(), rules.All()).ConfigForFile(visible)
	if merged == nil || merged.Rules["rule"] == nil {
		t.Fatalf("shared-alias effective config = %#v", merged)
	}
}

func TestDiscoverLintTargetsFromRoot_UnionsFilesAndDirectories(t *testing.T) {
	root := t.TempDir()
	configDir := tspath.NormalizePath(filepath.Join(root, "config"))
	scanRoot := tspath.NormalizePath(filepath.Join(root, "workspace"))
	directoryRoot := tspath.NormalizePath(filepath.Join(root, "directory"))
	exactRoot := tspath.NormalizePath(filepath.Join(root, "exact"))
	for _, directory := range []string{configDir, scanRoot, directoryRoot, exactRoot} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	directoryFile := tspath.CombinePaths(directoryRoot, "from-dir.ts")
	exactFile := tspath.CombinePaths(exactRoot, "exact.ts")
	for _, filePath := range []string{directoryFile, exactFile} {
		if err := os.WriteFile(filePath, []byte("export {};\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	targets := discoverLintTargetsFromRoot(
		nil,
		configDir,
		scanRoot,
		osvfs.FS(),
		[]string{exactFile},
		[]string{directoryRoot},
		true,
	)
	assert.DeepEqual(t, targets, []File{
		{PathIdentity: rslintconfig.PathIdentity{Path: directoryFile, CanonicalPath: tspath.NormalizePath(osvfs.FS().Realpath(directoryFile)), CanonicalParentPath: tspath.NormalizePath(osvfs.FS().Realpath(directoryRoot))}, ConfigDirectory: configDir},
		{PathIdentity: rslintconfig.PathIdentity{Path: exactFile, CanonicalPath: tspath.NormalizePath(osvfs.FS().Realpath(exactFile)), CanonicalParentPath: tspath.NormalizePath(osvfs.FS().Realpath(filepath.Dir(exactFile)))}, ConfigDirectory: configDir},
	})
}

func TestDiscoverLintTargetsFromRoot_RequestedAncestorWidensOnlyScanScope(t *testing.T) {
	root := t.TempDir()
	configDir := tspath.NormalizePath(filepath.Join(root, "config"))
	scanRoot := tspath.NormalizePath(filepath.Join(root, "workspace", "nested"))
	insideFile := tspath.CombinePaths(scanRoot, "inside.ts")
	siblingFile := tspath.CombinePaths(root, "sibling.ts")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(scanRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, filePath := range []string{insideFile, siblingFile} {
		if err := os.WriteFile(filePath, []byte("export {};\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	targets := discoverLintTargetsFromRoot(nil, configDir, scanRoot, osvfs.FS(), nil, []string{tspath.NormalizePath(root)}, true)
	paths := make([]string, 0, len(targets))
	for _, target := range targets {
		paths = append(paths, target.Path)
	}
	assert.DeepEqual(t, paths, []string{siblingFile, insideFile})
	for _, target := range targets {
		assert.Equal(t, target.ConfigDirectory, configDir)
	}
}

func TestDiscoverLintTargetsFromRoot_SkipsIgnoredExternalRootBeforeWalking(t *testing.T) {
	root := t.TempDir()
	configDir := tspath.NormalizePath(filepath.Join(root, "config"))
	scanRoot := tspath.NormalizePath(filepath.Join(root, "workspace"))
	externalDir := tspath.NormalizePath(filepath.Join(root, "external"))
	for _, directory := range []string{configDir, scanRoot, externalDir} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(tspath.CombinePaths(externalDir, "ignored.ts"), []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spy := &spyFS{FS: osvfs.FS()}
	targets := discoverLintTargetsFromRoot(
		rslintconfig.RslintConfig{
			{Ignores: []string{"../external/**"}},
			{Rules: rslintconfig.Rules{"test-rule": "error"}},
		},
		configDir,
		scanRoot,
		spy,
		nil,
		[]string{externalDir},
		true,
	)
	assert.DeepEqual(t, targets, []File{})
	for _, accessed := range spy.snapshotAccessedDirs() {
		if pathsEqual(accessed, externalDir, true) || tspath.StartsWithDirectory(accessed, externalDir, true) {
			t.Fatalf("ignored external root was entered: %v", spy.snapshotAccessedDirs())
		}
	}
}

func TestDiscoverLintTargetsFromRoot_SkipsDefaultExcludedScanRoot(t *testing.T) {
	root := t.TempDir()
	configDir := tspath.NormalizePath(root)
	scanRoot := tspath.NormalizePath(filepath.Join(root, "node_modules", "pkg"))
	if err := os.MkdirAll(scanRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tspath.CombinePaths(scanRoot, "index.ts"), []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name        string
		directories []string
	}{
		{name: "implicit scan root"},
		{name: "explicit scan root", directories: []string{scanRoot}},
	} {
		t.Run(test.name, func(t *testing.T) {
			spy := &spyFS{FS: osvfs.FS()}
			targets := discoverLintTargetsFromRoot(
				rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"test-rule": "error"}}},
				configDir,
				scanRoot,
				spy,
				nil,
				test.directories,
				true,
			)
			assert.DeepEqual(t, targets, []File{})
			for _, accessed := range spy.snapshotAccessedDirs() {
				if pathsEqual(accessed, scanRoot, true) || tspath.StartsWithDirectory(accessed, scanRoot, true) {
					t.Fatalf("default-excluded scan root was entered: %v", spy.snapshotAccessedDirs())
				}
			}
		})
	}
}

func TestDiscoverLintTargetsFromRoot_DistinctCanonicalRootsDoNotCaseFold(t *testing.T) {
	const (
		configDir = "/config"
		scanRoot  = "/workspace"
		upperDir  = "/repo/Foo"
		lowerDir  = "/repo/foo"
	)
	fake := &discoveryMockFS{
		FS: osvfs.FS(),
		entries: map[string]vfs.Entries{
			upperDir: {Files: []string{"upper.ts"}, Symlinks: map[string]struct{}{}},
			lowerDir: {Files: []string{"lower.ts"}, Symlinks: map[string]struct{}{}},
		},
		resolvedPaths: map[string]string{
			configDir: "/physical/config",
			scanRoot:  "/physical/workspace",
			upperDir:  "/physical/Foo",
			lowerDir:  "/physical/foo",
		},
		caseSensitiveFS: false,
	}
	targets := discoverLintTargetsFromRoot(nil, configDir, scanRoot, fake, nil, []string{upperDir, lowerDir}, true)
	paths := make([]string, 0, len(targets))
	for _, target := range targets {
		paths = append(paths, target.Path)
	}
	assert.DeepEqual(t, paths, []string{"/repo/Foo/upper.ts", "/repo/foo/lower.ts"})
}

func TestDiscoverLintTargets_DirectoryWalkAvoidsPerNodeRealpath(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/b.ts",
		"src/c.ts",
	})
	fsys := &realpathCountingFS{FS: osvfs.FS(), calls: make(map[string]int)}
	targets := discoverLintTargets(
		rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"test-rule": "error"}}},
		configDir,
		fsys,
		nil,
		[]string{configDir},
		true,
	)
	assert.Equal(t, len(targets), 3)
	for _, filePath := range paths {
		assert.Equal(t, fsys.callCount(filePath), 0, "regular walk target should use the config-root canonical hint")
	}
	assert.Equal(
		t,
		fsys.callCount(tspath.CombinePaths(configDir, "src")),
		0,
		"lexically contained walk directory should not resolve its physical path",
	)
}

func TestDiscoverLintTargets_ExplicitFileResolvesPhysicalIdentity(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{"src/a.ts"})
	fsys := &realpathCountingFS{FS: osvfs.FS(), calls: make(map[string]int)}
	targets := discoverLintTargets(nil, configDir, fsys, []string{paths["src/a.ts"]}, nil, true)
	assert.Equal(t, len(targets), 1)
	assert.Assert(t, fsys.callCount(paths["src/a.ts"]) > 0)
}

func TestDiscoverLintTargets_FileSymlinkResolvesPhysicalIdentity(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{"src/a.ts"})
	linkPath := tspath.NormalizePath(filepath.Join(configDir, "src/link.ts"))
	if err := os.Symlink(paths["src/a.ts"], linkPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	fsys := &realpathCountingFS{FS: osvfs.FS(), calls: make(map[string]int)}
	targets := discoverLintTargets(nil, configDir, fsys, nil, []string{configDir}, true)
	assert.Equal(t, len(targets), 2)

	canonicalByPath := make(map[string]string, len(targets))
	for _, target := range targets {
		canonicalByPath[target.Path] = target.CanonicalPath
	}
	assert.Equal(t, canonicalByPath[linkPath], tspath.NormalizePath(fsys.FS.Realpath(paths["src/a.ts"])))
	assert.Equal(t, fsys.callCount(paths["src/a.ts"]), 0)
	assert.Assert(t, fsys.callCount(linkPath) > 0)
}

func TestDiscoverLintTargets_MissingSymlinkMetadataResolvesFileIdentity(t *testing.T) {
	const configDir = "/repo"
	const filePath = "/repo/src/a.ts"
	const physicalPath = "/physical/src/a.ts"
	fsys := &discoveryMockFS{
		FS: osvfs.FS(),
		entries: map[string]vfs.Entries{
			configDir: {
				Directories: []string{"src"},
				// A nil Symlinks map means identity metadata is unavailable.
				Symlinks: nil,
			},
			"/repo/src": {
				Files:    []string{"a.ts"},
				Symlinks: nil,
			},
		},
		resolvedPaths: map[string]string{
			configDir:   configDir,
			"/repo/src": "/repo/src",
			filePath:    physicalPath,
		},
	}

	targets := discoverLintTargets(nil, configDir, fsys, nil, nil, true)
	assert.DeepEqual(t, targets, []File{{
		PathIdentity: rslintconfig.PathIdentity{
			Path:                filePath,
			CanonicalPath:       physicalPath,
			CanonicalParentPath: "/repo/src",
		},
		ConfigDirectory: configDir,
	}})
}

func TestDiscoverWalkRootsMapsCanonicalDirectoryAlias(t *testing.T) {
	fsys := &configPathSpaceFS{
		caseSensitive: true,
		realPaths: map[string]string{
			"/alias":    "/real",
			"/real/pkg": "/real/pkg",
		},
	}
	assert.DeepEqual(t, discoverWalkRoots("/alias", []string{"/real/pkg"}, fsys), []string{"pkg"})
}

func TestDiscoverFilesOutsidePrograms_MultipleFilesPatterns(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/b.tsx",
		"src/c.js",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"rule-a": "error"}},
		{Files: []string{"**/*.tsx"}, Rules: rslintconfig.Rules{"rule-b": "error"}},
		// .js remains in the default target baseline with zero matching rules.
	}

	programFiles := map[string]struct{}{}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	assert.Assert(t, gapFiles != nil)
	sort.Strings(gapFiles)

	expected := []string{paths["src/a.ts"], paths["src/b.tsx"], paths["src/c.js"]}
	sort.Strings(expected)

	assert.Equal(t, len(gapFiles), len(expected))
	for i := range expected {
		assert.Equal(t, gapFiles[i], expected[i])
	}
}

func TestDiscoverFilesOutsidePrograms_DefaultJsDiscoveredWithoutExplicitPattern(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/b.js",
	})

	// An explicit TypeScript selector does not remove the default JS target.
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 2)
}

func TestDiscoverFilesOutsidePrograms_AllowFilesScope(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"scripts/b.ts",
		"tools/c.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	// Only allow scripts/b.ts via allowFiles
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles,
		[]string{paths["scripts/b.ts"]}, nil,
		false,
	)

	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 1)
	assert.Equal(t, gapFiles[0], paths["scripts/b.ts"])
}

func TestDiscoverLintFiles_ExplicitFileBypassesFilesWithDirScope(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"explicit.js",
		"src/a.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	srcDir := tspath.NormalizePath(filepath.Join(configDir, "src"))
	targets := discoverLintFiles(
		config,
		configDir,
		osvfs.FS(),
		[]string{paths["explicit.js"]},
		[]string{srcDir},
		false,
	)

	expected := []string{paths["explicit.js"], paths["src/a.ts"]}
	sort.Strings(expected)
	assert.DeepEqual(t, targets, expected)
}

func TestDiscoverFilesOutsidePrograms_AllExtensionsDiscoveredByPattern(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/b.jsx",
		"src/c.cjs",
		"src/readme.md",
		"src/data.json",
	})

	// files: ['**/*'] matches the whole tree, but rslint still only lints
	// extensions it can parse.
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	assert.Assert(t, gapFiles != nil)
	expected := []string{paths["src/a.ts"], paths["src/b.jsx"], paths["src/c.cjs"]}
	sort.Strings(expected)
	assert.DeepEqual(t, gapFiles, expected)
}

func TestDiscoverFilesOutsideProgramsMultiConfig(t *testing.T) {
	configDir1, paths1 := setupDiscoveryFixture(t, []string{
		"a.ts",
	})
	configDir2, paths2 := setupDiscoveryFixture(t, []string{
		"b.tsx",
	})

	configMap := map[string]rslintconfig.RslintConfig{
		configDir1: {
			{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"rule-a": "error"}},
		},
		configDir2: {
			{Files: []string{"**/*.tsx"}, Rules: rslintconfig.Rules{"rule-b": "error"}},
		},
	}

	programFiles := map[string]struct{}{}

	gapFiles := discoverFilesOutsideProgramsMultiConfigForTest(configMap, osvfs.FS(), programFiles, nil, nil, false)

	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 2)

	gapSet := map[string]struct{}{}
	for _, f := range gapFiles {
		gapSet[f] = struct{}{}
	}
	_, hasA := gapSet[paths1["a.ts"]]
	_, hasB := gapSet[paths2["b.tsx"]]
	assert.Assert(t, hasA, "should find a.ts")
	assert.Assert(t, hasB, "should find b.tsx")
}

func TestDiscoverLintFilesMultiConfig_UsesNearestConfigOwner(t *testing.T) {
	rootDir, rootPaths := setupDiscoveryFixture(t, []string{
		"root.ts",
		"pkg/child.ts",
	})
	childDir := tspath.NormalizePath(filepath.Join(rootDir, "pkg"))

	configMap := map[string]rslintconfig.RslintConfig{
		rootDir: {
			{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"root-rule": "error"}},
		},
		childDir: {
			{Files: []string{"**/*.jsx"}, Rules: rslintconfig.Rules{"child-rule": "error"}},
		},
	}

	targets := discoverLintFilesMultiConfig(configMap, osvfs.FS(), nil, []string{rootDir}, false)

	expected := []string{rootPaths["root.ts"]}
	expected = append(expected, rootPaths["pkg/child.ts"])
	sort.Strings(expected)
	assert.DeepEqual(t, targets, expected)

	ownedTargets := discoverLintTargetsMultiConfig(configMap, nil, osvfs.FS(), nil, []string{rootDir}, false)
	owners := make(map[string]string, len(ownedTargets))
	for _, target := range ownedTargets {
		owners[target.Path] = target.ConfigDirectory
	}
	assert.Equal(t, owners[rootPaths["root.ts"]], rootDir)
	assert.Equal(t, owners[rootPaths["pkg/child.ts"]], childDir)
}

func TestDiscoverLintFilesMultiConfig_DoesNotWalkChildConfigFromParent(t *testing.T) {
	rootDir, rootPaths := setupDiscoveryFixture(t, []string{
		"root.ts",
		"pkg/child.ts",
	})
	childDir := tspath.NormalizePath(filepath.Join(rootDir, "pkg"))

	configMap := map[string]rslintconfig.RslintConfig{
		rootDir: {
			{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"root-rule": "error"}},
		},
		childDir: {
			{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"child-rule": "error"}},
		},
	}
	spy := &spyFS{FS: osvfs.FS()}

	targets := discoverLintFilesMultiConfig(configMap, spy, nil, []string{rootDir}, true)

	expected := []string{rootPaths["pkg/child.ts"], rootPaths["root.ts"]}
	sort.Strings(expected)
	assert.DeepEqual(t, targets, expected)

	childAccesses := 0
	rootAccesses := 0
	for _, accessed := range spy.snapshotAccessedDirs() {
		if accessed == rootDir {
			rootAccesses++
		}
		if accessed == childDir {
			childAccesses++
		}
	}
	assert.Equal(t, rootAccesses, 1, "each child config must not scan the invocation root again")
	assert.Equal(t, childAccesses, 1, "child config directory should be entered only by its owning config")
}

func TestConfigDirectoryIndex_UsesImmediateBoundariesAndNearestOwner(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/repo":          nil,
		"/repo/pkg":      nil,
		"/repo/pkg/deep": nil,
		"/other":         nil,
	}
	index := newConfigDirectoryIndex(configMap, nil)

	assert.DeepEqual(t, index.childConfigDirs("/repo"), []string{"/repo/pkg"})
	assert.DeepEqual(t, index.childConfigDirs("/repo/pkg"), []string{"/repo/pkg/deep"})
	owner, ok := NewOwnerIndex(configMap, nil).Resolve(
		FreezeFileIdentity("/repo/pkg/deep/src/index.ts", nil),
	)
	assert.Assert(t, ok)
	assert.Equal(t, owner, "/repo/pkg/deep")
}

func TestConfigDirectoryIndex_UsesVerifiedNativeCaseHierarchy(t *testing.T) {
	fsys := &caseInsensitiveDiscoveryFS{FS: osvfs.FS()}
	configMap := map[string]rslintconfig.RslintConfig{
		"C:/Repo":         nil,
		"c:/repo/Package": nil,
	}
	index := newConfigDirectoryIndex(configMap, fsys)

	assert.DeepEqual(t, index.childConfigDirs("C:/Repo"), []string{"c:/repo/Package"})
	for _, filePath := range []string{
		"C:/Repo/Package/src/index.ts",
		"c:/repo/package/src/index.ts",
	} {
		owner, ok := NewOwnerIndex(configMap, fsys).Resolve(
			FreezeFileIdentity(filePath, fsys),
		)
		assert.Assert(t, ok)
		assert.Equal(t, owner, "c:/repo/Package")
	}
}

func TestConfigDirectoryIndex_UsesPhysicalConfigRootForAliasedTarget(t *testing.T) {
	realDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(realDir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "src/index.ts"), []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(filepath.Dir(realDir), filepath.Base(realDir)+"-config-link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	defer os.Remove(linkDir)

	linkDir = tspath.NormalizePath(linkDir)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	configMap := map[string]rslintconfig.RslintConfig{linkDir: nil}
	owner, ok := NewOwnerIndex(configMap, fsys).Resolve(
		FreezeFileIdentity(filepath.Join(realDir, "src/index.ts"), fsys),
	)
	assert.Assert(t, ok)
	assert.Equal(t, owner, linkDir)
}

func TestDiscoverLintTargetsMultiConfig_MatchesIgnoresInPhysicalConfigSpace(t *testing.T) {
	realDir, paths := setupDiscoveryFixture(t, []string{"src/keep.ts", "src/ignored.ts"})
	linkDir := filepath.Join(filepath.Dir(realDir), filepath.Base(realDir)+"-config-link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	defer os.Remove(linkDir)
	linkDir = tspath.NormalizePath(linkDir)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	configMap := map[string]rslintconfig.RslintConfig{
		linkDir: {
			{Ignores: []string{"src/ignored.ts"}},
			{Files: []string{"src/**/*.ts"}, Rules: rslintconfig.Rules{"rule": "error"}},
		},
	}

	targets := discoverLintTargetsMultiConfig(
		configMap,
		nil,
		fsys,
		[]string{paths["src/ignored.ts"], paths["src/keep.ts"]},
		nil,
		true,
	)
	assert.DeepEqual(t, targets, []File{{
		PathIdentity: rslintconfig.PathIdentity{
			Path:                paths["src/keep.ts"],
			CanonicalPath:       tspath.NormalizePath(fsys.Realpath(paths["src/keep.ts"])),
			CanonicalParentPath: tspath.NormalizePath(fsys.Realpath(filepath.Dir(paths["src/keep.ts"]))),
		},
		ConfigDirectory: linkDir,
	}})
}

func TestDiscoverLintTargetsMultiConfig_AssignsExplicitFilesBeforeConfigProcessing(t *testing.T) {
	rootDir, paths := setupDiscoveryFixture(t, []string{"root.ts", "pkg/child.ts"})
	childDir := tspath.NormalizePath(filepath.Join(rootDir, "pkg"))
	fsys := &fileExistsCountingFS{FS: osvfs.FS()}
	configMap := map[string]rslintconfig.RslintConfig{
		rootDir:  {{Files: []string{"**/*.jsx"}, Rules: rslintconfig.Rules{"root": "error"}}},
		childDir: {{Files: []string{"**/*.jsx"}, Rules: rslintconfig.Rules{"child": "error"}}},
	}

	targets := discoverLintTargetsMultiConfig(
		configMap,
		nil,
		fsys,
		[]string{paths["pkg/child.ts"], paths["root.ts"]},
		nil,
		true,
	)
	assert.Equal(t, len(targets), 2)
	owners := map[string]string{}
	for _, target := range targets {
		owners[target.Path] = target.ConfigDirectory
	}
	assert.Equal(t, owners[paths["root.ts"]], rootDir)
	assert.Equal(t, owners[paths["pkg/child.ts"]], childDir)
	assert.Equal(t, fsys.calls.Load(), int32(2), "each explicit file should be checked only by its owner")
}

func TestDiscoverLintTargetsMultiConfig_PreservesHostAssignedExplicitOwner(t *testing.T) {
	rootDir, paths := setupDiscoveryFixture(t, []string{"pkg/index.ts"})
	childDir := tspath.NormalizePath(filepath.Join(rootDir, "pkg"))
	aliasPath := tspath.NormalizePath(filepath.Join(rootDir, "alias.ts"))
	if err := os.Symlink(paths["pkg/index.ts"], aliasPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	configMap := map[string]rslintconfig.RslintConfig{
		rootDir:  {{Rules: rslintconfig.Rules{"root": "error"}}},
		childDir: {{Rules: rslintconfig.Rules{"child": "error"}}},
	}

	targets := discoverLintTargetsMultiConfig(
		configMap,
		map[string]OwnerScope{
			rootDir: {ExplicitFiles: []string{aliasPath}},
		},
		fsys,
		[]string{aliasPath},
		nil,
		true,
	)
	if len(targets) != 1 {
		t.Fatalf("expected one explicitly assigned target, got %+v", targets)
	}
	assert.Equal(t, targets[0].Path, aliasPath)
	assert.Equal(t, targets[0].CanonicalPath, tspath.NormalizePath(fsys.Realpath(aliasPath)))
	assert.Equal(t, targets[0].ConfigDirectory, rootDir)
}

func TestDiscoverLintTargetsMultiConfig_MergesAutomaticAndHostAssignedFilesForSameOwner(t *testing.T) {
	rootDir, paths := setupDiscoveryFixture(t, []string{
		"pkg/automatic.ts",
		"pkg/explicit.ts",
	})
	childDir := tspath.NormalizePath(filepath.Join(rootDir, "pkg"))
	fsys := osvfs.FS()
	configMap := map[string]rslintconfig.RslintConfig{
		rootDir:  {{Rules: rslintconfig.Rules{"root": "error"}}},
		childDir: {{Rules: rslintconfig.Rules{"child": "error"}}},
	}

	targets := discoverLintTargetsMultiConfig(
		configMap,
		map[string]OwnerScope{
			childDir: {ExplicitFiles: []string{paths["pkg/explicit.ts"]}},
		},
		fsys,
		[]string{paths["pkg/automatic.ts"], paths["pkg/explicit.ts"]},
		nil,
		true,
	)

	assert.DeepEqual(t, targets, []File{
		{
			PathIdentity: rslintconfig.PathIdentity{
				Path:                paths["pkg/automatic.ts"],
				CanonicalPath:       tspath.NormalizePath(fsys.Realpath(paths["pkg/automatic.ts"])),
				CanonicalParentPath: tspath.NormalizePath(fsys.Realpath(filepath.Dir(paths["pkg/automatic.ts"]))),
			},
			ConfigDirectory: childDir,
		},
		{
			PathIdentity: rslintconfig.PathIdentity{
				Path:                paths["pkg/explicit.ts"],
				CanonicalPath:       tspath.NormalizePath(fsys.Realpath(paths["pkg/explicit.ts"])),
				CanonicalParentPath: tspath.NormalizePath(fsys.Realpath(filepath.Dir(paths["pkg/explicit.ts"]))),
			},
			ConfigDirectory: childDir,
		},
	})
}

func TestDiscoverLintTargetsMultiConfig_ExplicitOnlyConfigDoesNotOwnAutomaticFiles(t *testing.T) {
	rootDir, paths := setupDiscoveryFixture(t, []string{
		"ignored/automatic.ts",
		"ignored/explicit.ts",
	})
	ignoredDir := tspath.NormalizePath(filepath.Join(rootDir, "ignored"))
	configMap := map[string]rslintconfig.RslintConfig{
		rootDir: {
			{Ignores: []string{"ignored/**"}},
			{Rules: rslintconfig.Rules{"root": "error"}},
		},
		ignoredDir: {{Rules: rslintconfig.Rules{"nested": "error"}}},
	}
	fsys := osvfs.FS()

	targets := discoverLintTargetsMultiConfig(
		configMap,
		map[string]OwnerScope{
			ignoredDir: {
				ExplicitFiles: []string{paths["ignored/explicit.ts"]},
				ExplicitOnly:  true,
			},
		},
		fsys,
		[]string{paths["ignored/automatic.ts"], paths["ignored/explicit.ts"]},
		nil,
		true,
	)

	assert.DeepEqual(t, targets, []File{{
		PathIdentity: rslintconfig.PathIdentity{
			Path:                paths["ignored/explicit.ts"],
			CanonicalPath:       tspath.NormalizePath(fsys.Realpath(paths["ignored/explicit.ts"])),
			CanonicalParentPath: tspath.NormalizePath(fsys.Realpath(filepath.Dir(paths["ignored/explicit.ts"]))),
		},
		ConfigDirectory: ignoredDir,
	}})
}

func TestDiscoverLintTargetsMultiConfig_PrefersLexicalOwnerOverPhysicalConfig(t *testing.T) {
	rootDir, paths := setupDiscoveryFixture(t, []string{
		"real/other.ts",
		"real/sub/index.ts",
		"real/sub/nested/child.ts",
	})
	linkDir := tspath.NormalizePath(filepath.Join(rootDir, "link"))
	if err := os.Symlink(filepath.Join(rootDir, "real/sub"), linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	fSys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	realConfigDir := tspath.NormalizePath(filepath.Join(rootDir, "real"))
	nestedConfigDir := tspath.NormalizePath(filepath.Join(linkDir, "nested"))
	configMap := map[string]rslintconfig.RslintConfig{
		rootDir:         {{Rules: rslintconfig.Rules{"root": "error"}}},
		realConfigDir:   {{Rules: rslintconfig.Rules{"physical": "error"}}},
		nestedConfigDir: {{Rules: rslintconfig.Rules{"nested": "error"}}},
	}

	targets := discoverLintTargetsMultiConfig(
		configMap,
		map[string]OwnerScope{
			realConfigDir: {ExplicitFiles: []string{paths["real/other.ts"]}},
		},
		fSys,
		[]string{paths["real/other.ts"]},
		[]string{linkDir},
		true,
	)
	owners := make(map[string]string, len(targets))
	for _, target := range targets {
		owners[target.Path] = target.ConfigDirectory
	}
	assert.Equal(t, owners[tspath.ResolvePath(linkDir, "index.ts")], rootDir)
	assert.Equal(t, owners[tspath.ResolvePath(linkDir, "nested/child.ts")], nestedConfigDir)
	assert.Equal(t, owners[paths["real/other.ts"]], realConfigDir)
	assert.Equal(t, len(owners), 3)
}

func TestDiscoverLintTargets_DirectorySymlinkPreservesCanonicalIdentity(t *testing.T) {
	rootDir, paths := setupDiscoveryFixture(t, []string{"real/index.ts"})
	linkDir := tspath.NormalizePath(filepath.Join(rootDir, "link"))
	if err := os.Symlink(filepath.Join(rootDir, "real"), linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	targets := discoverLintTargets(
		rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"rule": "error"}}},
		rootDir,
		fsys,
		nil,
		[]string{linkDir},
		true,
	)
	if len(targets) != 1 {
		t.Fatalf("expected one target through directory symlink, got %+v", targets)
	}
	assert.Equal(t, targets[0].Path, tspath.ResolvePath(linkDir, "index.ts"))
	assert.Equal(t, targets[0].CanonicalPath, tspath.NormalizePath(fsys.Realpath(paths["real/index.ts"])))
}

func TestDiscoverLintTargets_PhysicalConfigFallbackPreservesDirectoryAliasPath(t *testing.T) {
	rootDir, paths := setupDiscoveryFixture(t, []string{"real/sub/index.ts"})
	linkDir := tspath.NormalizePath(filepath.Join(rootDir, "link"))
	if err := os.Symlink(filepath.Join(rootDir, "real/sub"), linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	fSys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	realConfigDir := tspath.NormalizePath(filepath.Join(rootDir, "real"))
	targets := discoverLintTargets(
		rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"rule": "error"}}},
		realConfigDir,
		fSys,
		nil,
		[]string{linkDir},
		true,
	)
	if len(targets) != 1 {
		t.Fatalf("expected one target through physical config fallback, got %+v", targets)
	}
	assert.Equal(t, targets[0].Path, tspath.ResolvePath(linkDir, "index.ts"))
	assert.Equal(t, targets[0].CanonicalPath, tspath.NormalizePath(fSys.Realpath(paths["real/sub/index.ts"])))
}

func TestDiscoverLintTargetsMultiConfig_AliasAuthoredEntrySelectsNonDefaultExtension(t *testing.T) {
	rootDir := tspath.NormalizePath(t.TempDir())
	realConfigDir := tspath.CombinePaths(rootDir, "real")
	realTargetDir := tspath.CombinePaths(realConfigDir, "sub")
	if err := os.MkdirAll(realTargetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	physicalTarget := tspath.CombinePaths(realTargetDir, "index.TS")
	if err := os.WriteFile(physicalTarget, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkDir := tspath.CombinePaths(rootDir, "link")
	if err := os.Symlink(realTargetDir, linkDir); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}

	inlineOverride := rslintconfig.ConfigWithAuthoredPathBase(rslintconfig.RslintConfig{{
		Files: []string{"link/**/*.TS"},
		Rules: rslintconfig.Rules{"inline": "error"},
	}}, rootDir)
	entries := append(rslintconfig.RslintConfig{{
		Files: []string{"sub/**/*.ts"},
		Rules: rslintconfig.Rules{"physical": "error"},
	}}, inlineOverride...)
	fSys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	targets := discoverLintTargetsMultiConfig(
		map[string]rslintconfig.RslintConfig{realConfigDir: entries},
		nil,
		fSys,
		nil,
		[]string{linkDir},
		true,
	)
	wantPath := tspath.CombinePaths(linkDir, "index.TS")
	if len(targets) != 1 ||
		targets[0].Path != wantPath ||
		targets[0].CanonicalPath != tspath.NormalizePath(fSys.Realpath(physicalTarget)) ||
		targets[0].ConfigDirectory != realConfigDir {
		t.Fatalf("alias-authored target = %+v, want %q owned by %q", targets, wantPath, realConfigDir)
	}
}

func TestDiscoverLintTargetsMultiConfig_PrunesUsingCallerVisibleDirectoryAlias(t *testing.T) {
	rootDir := tspath.NormalizePath(t.TempDir())
	realConfigDir := tspath.CombinePaths(rootDir, "real")
	realTargetDir := tspath.CombinePaths(realConfigDir, "sub")
	if err := os.MkdirAll(realTargetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	physicalTarget := tspath.CombinePaths(realTargetDir, "index.TS")
	if err := os.WriteFile(physicalTarget, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkDir := tspath.CombinePaths(rootDir, "link")
	if err := os.Symlink(realTargetDir, linkDir); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}

	inlineIgnore := rslintconfig.ConfigWithAuthoredPathBase(
		rslintconfig.RslintConfig{{Ignores: []string{"real/**"}}},
		rootDir,
	)
	entries := append(rslintconfig.RslintConfig{{
		Files: []string{"sub/**/*.TS"},
		Rules: rslintconfig.Rules{"physical": "error"},
	}}, inlineIgnore...)
	fSys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	targets := discoverLintTargetsMultiConfig(
		map[string]rslintconfig.RslintConfig{realConfigDir: entries},
		nil,
		fSys,
		nil,
		[]string{linkDir},
		true,
	)
	wantPath := tspath.CombinePaths(linkDir, "index.TS")
	if len(targets) != 1 || targets[0].Path != wantPath {
		t.Fatalf("directory alias was pruned in the physical path space: %+v", targets)
	}
}

func TestDiscoverLintTargetsMultiConfig_EvaluatesEveryRequestedDirectoryAlias(t *testing.T) {
	rootDir := tspath.NormalizePath(t.TempDir())
	realConfigDir := tspath.CombinePaths(rootDir, "real")
	realTargetDir := tspath.CombinePaths(realConfigDir, "sub")
	if err := os.MkdirAll(realTargetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	physicalTarget := tspath.CombinePaths(realTargetDir, "index.TS")
	if err := os.WriteFile(physicalTarget, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	aliasA := tspath.CombinePaths(rootDir, "a")
	aliasB := tspath.CombinePaths(rootDir, "b")
	for _, alias := range []string{aliasA, aliasB} {
		if err := os.Symlink(realTargetDir, alias); err != nil {
			t.Skipf("directory symlink unavailable: %v", err)
		}
	}
	entries := rslintconfig.ConfigWithAuthoredPathBase(rslintconfig.RslintConfig{{
		Files: []string{"b/**/*.TS"},
		Rules: rslintconfig.Rules{"inline": "error"},
	}}, rootDir)
	fSys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	wantPath := tspath.CombinePaths(aliasB, "index.TS")
	for _, directories := range [][]string{{aliasA, aliasB}, {aliasB, aliasA}} {
		targets := discoverLintTargetsMultiConfig(
			map[string]rslintconfig.RslintConfig{realConfigDir: entries},
			nil,
			fSys,
			nil,
			directories,
			true,
		)
		if len(targets) != 1 || targets[0].Path != wantPath {
			t.Fatalf("requested aliases %v produced %+v, want %q", directories, targets, wantPath)
		}
	}
}

func TestDiscoverLintTargetsMultiConfig_DefaultExcludesUseRequestedDirectoryAlias(t *testing.T) {
	for _, test := range []struct {
		name          string
		physicalParts []string
		aliasParts    []string
		wantTargets   int
	}{
		{
			name:          "lexical node_modules remains excluded",
			physicalParts: []string{"real", "sub"},
			aliasParts:    []string{"node_modules", "link"},
			wantTargets:   0,
		},
		{
			name:          "physical node_modules hidden by lexical alias remains visible",
			physicalParts: []string{"real", "node_modules", "pkg"},
			aliasParts:    []string{"link"},
			wantTargets:   1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			rootDir := tspath.NormalizePath(t.TempDir())
			realConfigDir := tspath.CombinePaths(rootDir, "real")
			physicalTargetDir := realConfigDir
			for _, part := range test.physicalParts[1:] {
				physicalTargetDir = tspath.CombinePaths(physicalTargetDir, part)
			}
			if err := os.MkdirAll(physicalTargetDir, 0o755); err != nil {
				t.Fatal(err)
			}
			physicalTarget := tspath.CombinePaths(physicalTargetDir, "index.TS")
			if err := os.WriteFile(physicalTarget, []byte("debugger;\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			aliasDir := rootDir
			for _, part := range test.aliasParts {
				aliasDir = tspath.CombinePaths(aliasDir, part)
			}
			if err := os.MkdirAll(tspath.GetDirectoryPath(aliasDir), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(physicalTargetDir, aliasDir); err != nil {
				t.Skipf("directory symlink unavailable: %v", err)
			}

			physicalRelative, ok := rslintconfig.RelativePathWithinConfigRoot(physicalTargetDir, realConfigDir, true)
			if !ok {
				t.Fatal("invalid physical fixture")
			}
			entries := rslintconfig.RslintConfig{{
				Files: []string{physicalRelative + "/**/*.TS"},
				Rules: rslintconfig.Rules{"physical": "error"},
			}}
			fSys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			targets := discoverLintTargetsMultiConfig(
				map[string]rslintconfig.RslintConfig{realConfigDir: entries},
				nil,
				fSys,
				nil,
				[]string{aliasDir},
				true,
			)
			if len(targets) != test.wantTargets {
				t.Fatalf("targets = %+v, want %d", targets, test.wantTargets)
			}
			if test.wantTargets == 1 && targets[0].Path != tspath.CombinePaths(aliasDir, "index.TS") {
				t.Fatalf("visible alias target = %+v", targets)
			}
		})
	}
}

func TestDiscoverLintTargetsMultiConfig_DefaultExcludeFiltersEachDirectoryAlias(t *testing.T) {
	rootDir := tspath.NormalizePath(t.TempDir())
	realConfigDir := tspath.CombinePaths(rootDir, "real")
	realTargetDir := tspath.CombinePaths(realConfigDir, "sub")
	if err := os.MkdirAll(realTargetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	physicalTarget := tspath.CombinePaths(realTargetDir, "index.TS")
	if err := os.WriteFile(physicalTarget, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	excludedAlias := tspath.CombinePaths(rootDir, "a", "node_modules", "link")
	visibleAlias := tspath.CombinePaths(rootDir, "z")
	if err := os.MkdirAll(tspath.GetDirectoryPath(excludedAlias), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, alias := range []string{excludedAlias, visibleAlias} {
		if err := os.Symlink(realTargetDir, alias); err != nil {
			t.Skipf("directory symlink unavailable: %v", err)
		}
	}
	entries := rslintconfig.RslintConfig{{
		Files: []string{"sub/**/*.TS"},
		Rules: rslintconfig.Rules{"physical": "error"},
	}}
	fSys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	targets := discoverLintTargetsMultiConfig(
		map[string]rslintconfig.RslintConfig{realConfigDir: entries},
		nil,
		fSys,
		nil,
		[]string{excludedAlias, visibleAlias},
		true,
	)
	wantPath := tspath.CombinePaths(visibleAlias, "index.TS")
	if len(targets) != 1 || targets[0].Path != wantPath {
		t.Fatalf("per-alias default exclusion produced %+v, want only %q", targets, wantPath)
	}
}

func TestProjectPathThroughRequestedDirectoriesPreservesCallerCasing(t *testing.T) {
	projections := projectPathThroughRequestedDirectories(
		"C:/Repo/src/a.TS",
		"C:/Repo/src/a.TS",
		[]resolvedAllowedDirectory{{
			LexicalPath:   "c:/repo/src",
			CanonicalPath: "C:/Repo/src",
		}},
		false,
	)
	if len(projections) != 1 ||
		projections[0].Path != "c:/repo/src/a.TS" ||
		projections[0].CanonicalPath != "C:/Repo/src/a.TS" {
		t.Fatalf("projections = %+v", projections)
	}

	projections = projectPathThroughRequestedDirectories(
		"C:/Repo/src/a.TS",
		"C:/Repo/src/a.TS",
		[]resolvedAllowedDirectory{
			{LexicalPath: "C:/Repo/src", CanonicalPath: "C:/Repo/src"},
			{LexicalPath: "c:/repo/src", CanonicalPath: "c:/repo/src"},
		},
		false,
	)
	if len(projections) != 1 ||
		projections[0].Path != "C:/Repo/src/a.TS" ||
		projections[0].CanonicalPath != "" {
		t.Fatalf("distinct case-only roots produced projections %+v", projections)
	}
}

func TestDiscoverFilesOutsideProgramsMultiConfig_UsesNearestConfigOwner(t *testing.T) {
	rootDir, rootPaths := setupDiscoveryFixture(t, []string{
		"root.ts",
		"pkg/child.ts",
		"pkg/child.jsx",
	})
	childDir := tspath.NormalizePath(filepath.Join(rootDir, "pkg"))

	configMap := map[string]rslintconfig.RslintConfig{
		rootDir: {
			{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"root-rule": "error"}},
		},
		childDir: {
			{Files: []string{"**/*.jsx"}, Rules: rslintconfig.Rules{"child-rule": "error"}},
		},
	}

	gapFiles := discoverFilesOutsideProgramsMultiConfigForTest(configMap, osvfs.FS(), map[string]struct{}{}, nil, []string{rootDir}, false)

	expected := []string{rootPaths["pkg/child.jsx"], rootPaths["pkg/child.ts"], rootPaths["root.ts"]}
	sort.Strings(expected)
	assert.DeepEqual(t, gapFiles, expected)
}

func TestDiscoverFilesOutsidePrograms_FilesButNoRules(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/a.ts",
	})

	// The selector contributes a target even though the entry has no rules.
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}},
	}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, false)

	// Discovery retains the gap target; the linter counts and parses it but does
	// not execute a rule traversal when the merged rule set is empty.
	assert.Assert(t, gapFiles != nil)
}

// --- Directory-level ignore blocking in discoverFilesOutsideProgramsForTest ---

func TestDiscoverFilesOutsidePrograms_DirIgnoreBlocksTraversal(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"build/keep.ts",
		"build/other.ts",
	})

	// build/** is directory-level → blocks traversal → ! cannot re-include
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"build/**", "!build/keep.ts"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	assert.Assert(t, gapFiles != nil)
	// build/ entirely blocked — neither keep.ts nor other.ts discovered
	for _, f := range gapFiles {
		if f == paths["build/keep.ts"] || f == paths["build/other.ts"] {
			t.Errorf("Expected build/ files to be blocked by dir/** ignore, but found %s", f)
		}
	}
	// src/index.ts should be discovered
	found := false
	for _, f := range gapFiles {
		if f == paths["src/index.ts"] {
			found = true
		}
	}
	assert.Assert(t, found, "src/index.ts should be discovered")
}

func TestDiscoverFilesOutsidePrograms_FileIgnoreAllowsNegation(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"build/keep.ts",
		"build/other.ts",
	})

	// build/**/* is file-level → does NOT block traversal → ! CAN re-include
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"build/**/*", "!build/keep.ts"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	assert.Assert(t, gapFiles != nil)

	gapSet := make(map[string]struct{})
	for _, f := range gapFiles {
		gapSet[f] = struct{}{}
	}

	// build/keep.ts re-included by ! → discovered
	_, hasKeep := gapSet[paths["build/keep.ts"]]
	assert.Assert(t, hasKeep, "build/keep.ts should be re-included by negation")

	// build/other.ts still ignored → not discovered
	_, hasOther := gapSet[paths["build/other.ts"]]
	assert.Assert(t, !hasOther, "build/other.ts should remain ignored")

	// src/index.ts always discovered
	_, hasSrc := gapSet[paths["src/index.ts"]]
	assert.Assert(t, hasSrc, "src/index.ts should be discovered")
}

func TestDiscoverFilesOutsidePrograms_WildcardMiddleDirIgnoreBlocks(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"packages/app/dist/gen.ts",
		"packages/app/src/index.ts",
	})

	config := rslintconfig.RslintConfig{
		{Ignores: []string{"packages/*/dist/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	gapSet := make(map[string]struct{})
	for _, f := range gapFiles {
		gapSet[f] = struct{}{}
	}

	_, hasDist := gapSet[paths["packages/app/dist/gen.ts"]]
	assert.Assert(t, !hasDist, "packages/app/dist/ should be blocked by packages/*/dist/**")

	_, hasSrc := gapSet[paths["packages/app/src/index.ts"]]
	assert.Assert(t, hasSrc, "packages/app/src/ should not be blocked")
}

func TestDiscoverFilesOutsidePrograms_CrossEntryDirIgnoreAndNegation(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"build/keep.ts",
		"build/other.ts",
	})

	// Entry 1: dir-level ignore. Entry 2: negation.
	// dir/** blocks traversal → negation cannot re-include (even across entries,
	// because global ignores are merged and dir-level patterns block first).
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"build/**"}},
		{Ignores: []string{"!build/keep.ts"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	for _, f := range gapFiles {
		if f == paths["build/keep.ts"] || f == paths["build/other.ts"] {
			t.Errorf("Expected build/ blocked by dir/** even with cross-entry negation, but found %s", f)
		}
	}
}

func TestDiscoverFilesOutsidePrograms_DoubleStarDirIgnoreBlocksNested(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"packages/app/dist/gen.ts",
	})

	// **/dist/** blocks any dist/ directory
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/dist/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), programFiles, nil, nil, false)

	assert.Assert(t, gapFiles != nil)
	for _, f := range gapFiles {
		if f == paths["packages/app/dist/gen.ts"] {
			t.Error("Expected dist/ files to be blocked by **/dist/** ignore")
		}
	}
}

// --- Default excludes (node_modules, .git) ---

func TestDiscoverFilesOutsidePrograms_DefaultExcludesNodeModules(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"node_modules/pkg/index.ts",
	})

	// No global ignores at all — defaults should still exclude node_modules
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, false)

	assert.Assert(t, gapFiles != nil)
	for _, f := range gapFiles {
		if f == paths["node_modules/pkg/index.ts"] {
			t.Error("node_modules should be excluded by default even without user ignores")
		}
	}
	// src/index.ts should still be discovered
	found := false
	for _, f := range gapFiles {
		if f == paths["src/index.ts"] {
			found = true
		}
	}
	assert.Assert(t, found, "src/index.ts should be discovered")
}

func TestDiscoverFilesOutsidePrograms_DefaultExcludesGitDir(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		".git/hooks/pre-commit.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, false)

	assert.Assert(t, gapFiles != nil)
	for _, f := range gapFiles {
		if strings.Contains(f, ".git") {
			t.Errorf(".git should be excluded by default, but found %s", f)
		}
	}
}

// --- Directory pruning verification (spy vfs) ---
// These tests verify that WalkDir actually skips excluded directories at the
// walk level (fs.SkipDir), not just filters files after entering them.
// This is critical for performance: entering node_modules with 10,000+ files
// is the difference between <100ms and 7s.

// spyFS wraps a vfs.FS and records which directories had their contents listed.
// GetAccessibleEntries is called by vfsAdapter.ReadDir only when the walker
// actually enters a directory. If a directory is pruned, this method is
// never called for it.
//
// discoverFilesOutsideProgramsForTest walks concurrently, so the recorder needs a lock.
type spyFS struct {
	vfs.FS
	mu           sync.Mutex
	accessedDirs []string
}

func (s *spyFS) GetAccessibleEntries(path string) vfs.Entries {
	s.mu.Lock()
	s.accessedDirs = append(s.accessedDirs, path)
	s.mu.Unlock()
	return s.FS.GetAccessibleEntries(path)
}

func (s *spyFS) snapshotAccessedDirs() []string {
	s.mu.Lock()
	out := append([]string(nil), s.accessedDirs...)
	s.mu.Unlock()
	return out
}

type caseInsensitiveSpyFS struct {
	*spyFS
}

func (s *caseInsensitiveSpyFS) UseCaseSensitiveFileNames() bool {
	return false
}

func TestDiscoverFilesOutsidePrograms_PrunesNodeModulesAtWalkLevel(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"node_modules/pkg/index.ts",
		"node_modules/pkg/nested/deep.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

	for _, dir := range spy.snapshotAccessedDirs() {
		if strings.Contains(dir, "node_modules") {
			t.Errorf("node_modules directory was entered during walk (GetAccessibleEntries called for %s)", dir)
		}
	}
}

func TestDiscoverFilesOutsidePrograms_PrunesDefaultExcludesCaseInsensitive(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"Node_Modules/pkg/index.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &caseInsensitiveSpyFS{spyFS: &spyFS{FS: osvfs.FS()}}
	discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, true)

	for _, dir := range spy.snapshotAccessedDirs() {
		if strings.Contains(dir, "Node_Modules") {
			t.Errorf("Node_Modules directory was entered on case-insensitive FS (GetAccessibleEntries called for %s)", dir)
		}
	}
}

func TestDiscoverFilesOutsidePrograms_PrunesGitDirAtWalkLevel(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		".git/objects/ab/file.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

	for _, dir := range spy.snapshotAccessedDirs() {
		if strings.Contains(dir, ".git") {
			t.Errorf(".git directory was entered during walk (GetAccessibleEntries called for %s)", dir)
		}
	}
}

func TestDiscoverFilesOutsidePrograms_PrunesUserIgnoredDirAtWalkLevel(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"vendor/lib/util.ts",
		"vendor/lib/nested/deep.ts",
	})

	config := rslintconfig.RslintConfig{
		{Ignores: []string{"vendor/**"}}, // global ignore
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

	for _, dir := range spy.snapshotAccessedDirs() {
		if strings.Contains(dir, "vendor") {
			t.Errorf("vendor directory was entered during walk (GetAccessibleEntries called for %s)", dir)
		}
	}
}

func TestDiscoverFilesOutsidePrograms_PrunesNestedIgnoredDirButEntersParent(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"build/index.ts",
		"build/output/bundle.ts",
		"build/output/nested/deep.ts",
	})

	config := rslintconfig.RslintConfig{
		{Ignores: []string{"build/output/**"}}, // only output/ ignored, not build/
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

	// build/ should be entered (not blocked)
	buildEntered := false
	outputEntered := false
	for _, dir := range spy.snapshotAccessedDirs() {
		if strings.HasSuffix(dir, "/build") || strings.HasSuffix(dir, "build") {
			buildEntered = true
		}
		if strings.Contains(dir, "build/output") {
			outputEntered = true
		}
	}
	assert.Assert(t, buildEntered, "build/ directory should be entered")
	assert.Assert(t, !outputEntered, "build/output/ directory should NOT be entered (pruned)")

	// build/index.ts should be in gap files
	found := false
	for _, f := range gapFiles {
		if f == paths["build/index.ts"] {
			found = true
		}
	}
	assert.Assert(t, found, "build/index.ts should be discovered")
}

func TestDiscoverFilesOutsidePrograms_EntersNonExcludedDirs(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"src/components/button.ts",
		"lib/utils.ts",
	})

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

	// src/ and lib/ should be entered
	srcEntered := false
	libEntered := false
	for _, dir := range spy.snapshotAccessedDirs() {
		if strings.HasSuffix(dir, "/src") {
			srcEntered = true
		}
		if strings.HasSuffix(dir, "/lib") {
			libEntered = true
		}
	}
	assert.Assert(t, srcEntered, "src/ should be entered")
	assert.Assert(t, libEntered, "lib/ should be entered")

	// All 3 files should be discovered
	assert.Equal(t, len(gapFiles), 3)
	_ = paths
}

// =============================================================================
// End-to-end cross-matrix tests: config ignores × .gitignore × gap files × linter
//
// These tests simulate the full flow:
//   1. rslintconfig.ConfigWithGitignore (with config ignores for pruning)
//   2. Inject gitignore globs into config
//   3. discoverFilesOutsideProgramsForTest
//   4. Verify GetConfigForFile (linter's per-file decision) is consistent
//
// The structural guarantee being tested: if isDirAbsolutelyBlocked(dir, configIgnores)
// returns true in collectGitignoreGlobs (causing .gitignore skip), then
// GetConfigForFile also returns nil for any file in that dir.
// =============================================================================

// e2eSetup creates a fixture, applies rslintconfig.ConfigWithGitignore, then runs discoverFilesOutsideProgramsForTest,
// and returns gap files + the final config (for GetConfigForFile verification).
func e2eSetup(t *testing.T, files map[string]string, config rslintconfig.RslintConfig, programFiles map[string]struct{}) (string, []string, rslintconfig.RslintConfig) {
	t.Helper()
	dir := setupDiscoveryContentFixture(t, files)

	config = rslintconfig.ConfigWithGitignore(config, dir, osvfs.FS(), nil)

	if programFiles == nil {
		programFiles = map[string]struct{}{}
	}

	gapFiles := discoverFilesOutsideProgramsForTest(config, dir, osvfs.FS(), programFiles, nil, nil, false)
	return dir, gapFiles, config
}

// E2E case 1: Standard rspack-like scenario.
// config ignores tests/, .gitignore ignores target/.
// Files in src/ should be gap files. Files in tests/ and target/ should not.
func TestE2E_StandardMonorepo(t *testing.T) {
	files := map[string]string{
		".gitignore":               "target/\n",
		"src/index.ts":             "x",
		"src/utils.ts":             "x",
		"tests/unit/a.test.ts":     "x",
		"tests/.gitignore":         "snapshots/\n",
		"target/build/output.ts":   "x",
		"packages/foo/src/main.ts": "x",
	}
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	dir, gapFiles, finalConfig := e2eSetup(t, files, config, nil)

	// Exactly 3 gap files: src/index.ts, src/utils.ts, packages/foo/src/main.ts.
	// tests/ excluded by config ignore, target/ excluded by gitignore.
	assert.Equal(t, len(gapFiles), 3, "should have exactly 3 gap files, got %d: %v", len(gapFiles), gapFiles)

	gapSet := toSet(gapFiles)
	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/src/index.ts")])
	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/src/utils.ts")])
	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/packages/foo/src/main.ts")])

	// Linter consistency: excluded files return nil from GetConfigForFile.
	for _, excluded := range []string{"/tests/unit/a.test.ts", "/target/build/output.ts"} {
		mc := finalConfig.GetConfigForFile(tspath.NormalizePath(dir+excluded), dir)
		assert.Assert(t, mc == nil, "GetConfigForFile(%s) should be nil", excluded)
	}

	// Linter consistency: included files return non-nil.
	for _, included := range []string{"/src/index.ts", "/packages/foo/src/main.ts"} {
		mc := finalConfig.GetConfigForFile(tspath.NormalizePath(dir+included), dir)
		assert.Assert(t, mc != nil, "GetConfigForFile(%s) should be non-nil", included)
	}
}

// E2E case 2: Nested .gitignore in non-ignored dir affects TS Program files.
// packages/foo/.gitignore ignores generated/. A file there is in programFiles
// (simulating tsconfig inclusion). GetConfigForFile should return nil for it.
func TestE2E_NestedGitignoreAffectsProgramFiles(t *testing.T) {
	files := map[string]string{
		".gitignore":                        "dist/\n",
		"packages/foo/.gitignore":           "generated/\n",
		"packages/foo/src/index.ts":         "x",
		"packages/foo/src/generated/api.ts": "x",
	}
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	// Build programFiles with both files (simulating tsconfig include: ["src"])
	dir := setupDiscoveryContentFixture(t, files)
	indexPath := tspath.NormalizePath(dir + "/packages/foo/src/index.ts")
	genPathFull := tspath.NormalizePath(dir + "/packages/foo/src/generated/api.ts")

	config = rslintconfig.ConfigWithGitignore(config, dir, osvfs.FS(), nil)

	// generated/api.ts is in programFiles but gitignored.
	// GetConfigForFile should return nil because gitignore patterns are in config.
	mc := config.GetConfigForFile(genPathFull, dir)
	assert.Assert(t, mc == nil, "GetConfigForFile should return nil for gitignored generated/ file (nested .gitignore collected)")

	// index.ts is in programFiles and NOT gitignored → should get rules.
	mc = config.GetConfigForFile(indexPath, dir)
	assert.Assert(t, mc != nil, "GetConfigForFile should return config for non-ignored file")
}

// E2E case 3: Config ignores with file-level pattern (**/tests/**/*).
// The negation !tests/e2e/**/* re-includes tests/e2e/ at file level.
// tests/e2e/ files should be discoverable as gap files.
func TestE2E_FileLevelIgnoreWithNegation(t *testing.T) {
	files := map[string]string{
		"src/index.ts":         "x",
		"tests/unit/a.ts":      "x",
		"tests/e2e/smoke.ts":   "x",
		"tests/.gitignore":     "tmp/\n",
		"tests/e2e/.gitignore": "screenshots/\n",
	}
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/tests/**/*", "!tests/e2e/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	dir, gapFiles, finalConfig := e2eSetup(t, files, config, nil)

	gapSet := toSet(gapFiles)

	// src/index.ts → discovered
	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/src/index.ts")], "src/index.ts should be gap file")

	// tests/unit/a.ts → file-level ignored, not negated → NOT discovered
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/tests/unit/a.ts")], "tests/unit/ should be excluded")

	// tests/e2e/smoke.ts → file-level ignored BUT negated → discovered!
	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/tests/e2e/smoke.ts")], "tests/e2e/ should be re-included by negation")

	// Verify linter: tests/e2e/smoke.ts gets rules
	mc := finalConfig.GetConfigForFile(tspath.NormalizePath(dir+"/tests/e2e/smoke.ts"), dir)
	assert.Assert(t, mc != nil, "GetConfigForFile should return config for negation-included tests/e2e/ file")

	// Verify linter: tests/unit/a.ts does NOT get rules
	mc = finalConfig.GetConfigForFile(tspath.NormalizePath(dir+"/tests/unit/a.ts"), dir)
	assert.Assert(t, mc == nil, "GetConfigForFile should return nil for ignored tests/unit/ file")
}

// E2E case 4: .gitignore + config ignores target different dirs.
// .gitignore ignores dist/, config ignores vendor/. Both should be excluded.
func TestE2E_GitignoreAndConfigIgnoreIndependent(t *testing.T) {
	files := map[string]string{
		".gitignore":         "dist/\n",
		"src/index.ts":       "x",
		"dist/bundle.ts":     "x",
		"vendor/lib/util.ts": "x",
	}
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/vendor/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	dir, gapFiles, finalConfig := e2eSetup(t, files, config, nil)

	gapSet := toSet(gapFiles)

	// src/ → discovered
	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/src/index.ts")], "src/index.ts should be gap file")
	// dist/ → excluded by gitignore
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/dist/bundle.ts")], "dist/ should be excluded by gitignore")
	// vendor/ → excluded by config ignore
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/vendor/lib/util.ts")], "vendor/ should be excluded by config ignore")

	// Verify linter
	mc := finalConfig.GetConfigForFile(tspath.NormalizePath(dir+"/dist/bundle.ts"), dir)
	assert.Assert(t, mc == nil, "dist/ file should be nil in linter (gitignored)")
	mc = finalConfig.GetConfigForFile(tspath.NormalizePath(dir+"/vendor/lib/util.ts"), dir)
	assert.Assert(t, mc == nil, "vendor/ file should be nil in linter (config-ignored)")
}

// E2E case 5: No .gitignore at all + config ignores.
// Only config ignores are active.
func TestE2E_NoGitignoreOnlyConfigIgnores(t *testing.T) {
	files := map[string]string{
		"src/index.ts":         "x",
		"tests/unit/a.test.ts": "x",
	}
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	dir, gapFiles, _ := e2eSetup(t, files, config, nil)

	gapSet := toSet(gapFiles)

	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/src/index.ts")], "src/index.ts should be gap file")
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/tests/unit/a.test.ts")], "tests/ should be excluded")
}

// E2E case 6: programFiles interaction — file in program is not a gap file.
func TestE2E_ProgramFilesExcluded(t *testing.T) {
	files := map[string]string{
		".gitignore":   "dist/\n",
		"src/index.ts": "x",
		"src/utils.ts": "x",
	}
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	// Create fixture once and use the same dir for programFiles and e2e flow.
	dir := setupDiscoveryContentFixture(t, files)
	indexPath := tspath.NormalizePath(dir + "/src/index.ts")
	utilsPath := tspath.NormalizePath(dir + "/src/utils.ts")

	programFiles := map[string]struct{}{
		indexPath: {},
	}

	config = rslintconfig.ConfigWithGitignore(config, dir, osvfs.FS(), nil)

	gapFiles := discoverFilesOutsideProgramsForTest(config, dir, osvfs.FS(), programFiles, nil, nil, false)
	gapSet := toSet(gapFiles)

	// index.ts in program → NOT a gap file
	assert.Assert(t, !gapSet[indexPath], "file in program should not be gap file")
	// utils.ts not in program → IS a gap file
	assert.Assert(t, gapSet[utilsPath], "file not in program should be gap file")
}

// E2E case 7: Multiple config ignore entries + gitignore — verify all patterns combine correctly.
func TestE2E_MultipleIgnoreEntries(t *testing.T) {
	files := map[string]string{
		".gitignore":            "target/\n",
		"src/index.ts":          "x",
		"tests/a.test.ts":       "x",
		"scripts/build.ts":      "x",
		"target/output.ts":      "x",
		"packages/foo/index.ts": "x",
	}
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Ignores: []string{"scripts/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	dir, gapFiles, finalConfig := e2eSetup(t, files, config, nil)
	gapSet := toSet(gapFiles)

	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/src/index.ts")], "src/ → discovered")
	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/packages/foo/index.ts")], "packages/ → discovered")
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/tests/a.test.ts")], "tests/ → config-ignored")
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/scripts/build.ts")], "scripts/ → config-ignored")
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/target/output.ts")], "target/ → gitignored")

	// Verify linter consistency for all excluded paths
	for _, excluded := range []string{"/tests/a.test.ts", "/scripts/build.ts", "/target/output.ts"} {
		mc := finalConfig.GetConfigForFile(tspath.NormalizePath(dir+excluded), dir)
		assert.Assert(t, mc == nil, "GetConfigForFile(%s) should return nil", excluded)
	}
}

// E2E case 8: config-ignored directory's file is in TS Program → GetConfigForFile returns nil.
// This verifies that even if tsconfig pulls in files from a config-ignored directory,
// the linter correctly skips them (isDirBlockedByIgnores blocks the directory).
func TestE2E_ConfigIgnoredDirInProgram(t *testing.T) {
	files := map[string]string{
		"src/index.ts":           "x",
		"tests/helpers/setup.ts": "x",
	}
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	dir := setupDiscoveryContentFixture(t, files)
	testFile := tspath.NormalizePath(dir + "/tests/helpers/setup.ts")

	config = rslintconfig.ConfigWithGitignore(config, dir, osvfs.FS(), nil)

	// Even though setup.ts is in programFiles, GetConfigForFile should return nil
	// because tests/ is directory-blocked by config ignores.
	mc := config.GetConfigForFile(testFile, dir)
	assert.Assert(t, mc == nil, "GetConfigForFile should return nil for file in config-ignored dir, even if in TS Program")
}

// E2E case 9: gitignore and config ignore both target the same directory (dist/).
// Both mechanisms should work — the file should be excluded regardless of which one catches it.
func TestE2E_OverlappingGitignoreAndConfigIgnore(t *testing.T) {
	files := map[string]string{
		".gitignore":     "dist/\n",
		"src/index.ts":   "x",
		"dist/bundle.ts": "x",
	}
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/dist/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	dir, gapFiles, finalConfig := e2eSetup(t, files, config, nil)
	gapSet := toSet(gapFiles)

	distFile := tspath.NormalizePath(dir + "/dist/bundle.ts")
	assert.Assert(t, !gapSet[distFile], "dist/ should be excluded (both gitignore and config ignore)")

	mc := finalConfig.GetConfigForFile(distFile, dir)
	assert.Assert(t, mc == nil, "GetConfigForFile should return nil for doubly-ignored dist/ file")
}

// E2E case 10: allowDirs scope combined with config ignores.
// Only files in the allowed directory should be discovered, config ignores still apply.
func TestE2E_AllowDirsWithConfigIgnores(t *testing.T) {
	files := map[string]string{
		".gitignore":             "dist/\n",
		"packages/foo/src/a.ts":  "x",
		"packages/foo/dist/b.ts": "x",
		"packages/bar/src/c.ts":  "x",
		"tests/unit/d.ts":        "x",
	}
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	dir := setupDiscoveryContentFixture(t, files)
	config = rslintconfig.ConfigWithGitignore(config, dir, osvfs.FS(), nil)

	// Only allow packages/foo/
	fooDir := tspath.NormalizePath(dir + "/packages/foo")
	gapFiles := discoverFilesOutsideProgramsForTest(config, dir, osvfs.FS(), map[string]struct{}{}, nil, []string{fooDir}, false)
	gapSet := toSet(gapFiles)

	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/packages/foo/src/a.ts")], "packages/foo/src/a.ts should be discovered (in allowDirs)")
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/packages/foo/dist/b.ts")], "packages/foo/dist/b.ts should be excluded (gitignored)")
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/packages/bar/src/c.ts")], "packages/bar/ should be excluded (not in allowDirs)")
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/tests/unit/d.ts")], "tests/ should be excluded (config-ignored)")
}

// E2E case 11: allowFiles (lint-staged fast path) combined with gitignore injection.
// Files passed explicitly should still be filtered by gitignore patterns.
func TestE2E_AllowFilesWithGitignore(t *testing.T) {
	files := map[string]string{
		".gitignore":     "dist/\n",
		"src/index.ts":   "x",
		"dist/bundle.ts": "x",
	}
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	dir := setupDiscoveryContentFixture(t, files)
	config = rslintconfig.ConfigWithGitignore(config, dir, osvfs.FS(), nil)

	srcFile := tspath.NormalizePath(dir + "/src/index.ts")
	distFile := tspath.NormalizePath(dir + "/dist/bundle.ts")

	// Simulate lint-staged passing both files explicitly.
	gapFiles := discoverFilesOutsideProgramsForTest(config, dir, osvfs.FS(), map[string]struct{}{}, []string{srcFile, distFile}, nil, false)
	gapSet := toSet(gapFiles)

	assert.Assert(t, gapSet[srcFile], "src/index.ts should be discovered (explicit allowFile)")
	assert.Assert(t, !gapSet[distFile], "dist/bundle.ts should be excluded (gitignored even when explicitly passed)")
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return m
}

// --- Concurrency/correctness regressions ---

// Symlinks are never followed in discoverFilesOutsideProgramsForTest, so symlinked subtrees do
// not contribute gap files even if their target would otherwise match the
// `files` pattern. This guarantees output determinism regardless of how the
// concurrent walker schedules sibling directories.
func TestDiscoverFilesOutsidePrograms_SkipsSymlinkedDirs(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := tspath.NormalizePath(tmpDir)

	realDir := filepath.Join(tmpDir, "real")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "in_real.ts"), []byte("// in_real"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Two symlinks to the same target — both must be skipped, regardless
	// of which a concurrent walker would visit first.
	if err := os.Symlink(realDir, filepath.Join(tmpDir, "link_a")); err != nil {
		t.Fatalf("symlink a: %v", err)
	}
	if err := os.Symlink(realDir, filepath.Join(tmpDir, "link_b")); err != nil {
		t.Fatalf("symlink b: %v", err)
	}

	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"r": "error"}},
	}

	for _, single := range []bool{false, true} {
		gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(),
			map[string]struct{}{}, nil, nil, single)
		assert.Equal(t, len(gapFiles), 1, "singleThreaded=%v: only the real path should produce a gap file", single)
		realPath := tspath.NormalizePath(filepath.Join(realDir, "in_real.ts"))
		assert.Equal(t, gapFiles[0], realPath, "singleThreaded=%v: gap file should be the canonical realpath", single)
	}
}

// singleThreaded=true and singleThreaded=false must produce the same gap-file
// set. The concurrent walker is correctness-preserving; --singleThreaded is a
// performance/debuggability knob, not a behavioral one.
func TestDiscoverFilesOutsidePrograms_SingleThreadedEquivalence(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/nested/deep/b.ts",
		"scripts/c.ts",
		"scripts/sub/d.ts",
		"tools/e.ts",
		"tools/skip/f.ts",
	})
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/skip/**"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"r": "error"}},
	}

	parallelGaps := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(),
		map[string]struct{}{}, nil, nil, false)
	serialGaps := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(),
		map[string]struct{}{}, nil, nil, true)

	// Both already sorted by discoverFilesOutsideProgramsForTest; equality is exact.
	assert.Equal(t, len(parallelGaps), len(serialGaps), "must produce same count")
	for i := range parallelGaps {
		assert.Equal(t, parallelGaps[i], serialGaps[i], "diverged at i=%d", i)
	}
}

// allowFiles fast path must produce a deterministic, sorted result. The
// implementation iterates a map, so without an explicit sort the output
// order is randomized across runs.
func TestDiscoverFilesOutsidePrograms_AllowFilesFastPathSorted(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"a.ts",
		"b.ts",
		"c.ts",
		"d.ts",
		"e.ts",
	})
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"r": "error"}},
	}
	allow := []string{paths["e.ts"], paths["a.ts"], paths["c.ts"], paths["b.ts"], paths["d.ts"]}
	expected := []string{paths["a.ts"], paths["b.ts"], paths["c.ts"], paths["d.ts"], paths["e.ts"]}
	sort.Strings(expected)

	// Run multiple times to give Go's randomized map iteration a chance to
	// surface non-determinism if the sort is missing.
	for i := range 8 {
		got := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(),
			map[string]struct{}{}, allow, nil, false)
		assert.Equal(t, len(got), len(expected), "run %d: count mismatch", i)
		for j := range expected {
			assert.Equal(t, got[j], expected[j], "run %d, idx %d: not in lexical order", i, j)
		}
	}
}

// walkPool must hold concurrent execution of `work` to at most `workers`
// at any moment. Tested with an atomic counter inside the work function —
// independent of process-wide goroutine count, so it is robust under
// `go test -race`, varied GOMAXPROCS, and CI background noise.
func TestWalkPool_BoundsConcurrency(t *testing.T) {
	const (
		workers  = 4
		dirCount = 200
		hold     = 2 * time.Millisecond // make work observable
		fanout   = 3
		maxDepth = 3 // 3^3 = 27 leaves per root × 200 roots ≈ thousands of jobs
	)

	pool := newWalkPool(workers)
	// Seed with `dirCount` independent root jobs so the queue has enough
	// fan-out to actually exercise concurrency.
	roots := make([]string, dirCount)
	for i := range roots {
		roots[i] = "r/" + strconvI(i)
	}
	pool.submitMany(roots)

	var (
		active    atomic.Int32
		maxActive atomic.Int32
		jobsRun   atomic.Int32
	)

	work := func(dir string) []string {
		cur := active.Add(1)
		// Track high-water mark.
		for {
			old := maxActive.Load()
			if cur <= old || maxActive.CompareAndSwap(old, cur) {
				break
			}
		}
		// Hold the slot briefly to let other workers pile up if they could.
		time.Sleep(hold)
		active.Add(-1)
		jobsRun.Add(1)

		// Fan out a few children up to maxDepth so the pool keeps having
		// work to dispatch.
		depth := strings.Count(dir, "/")
		if depth >= maxDepth+1 { // r/<i> already has 1 slash
			return nil
		}
		out := make([]string, fanout)
		for k := range fanout {
			out[k] = dir + "/" + strconvI(k)
		}
		return out
	}

	pool.run(work)

	if got := maxActive.Load(); got > workers {
		t.Fatalf("walkPool exceeded its concurrency cap: maxActive=%d > workers=%d", got, workers)
	}
	// Sanity: it actually did work.
	if got := jobsRun.Load(); got < int32(dirCount) {
		t.Fatalf("walkPool did not process all jobs: jobsRun=%d < dirCount=%d", got, dirCount)
	}
}

// Verifies the workers==1 specialization in walkPool.run: the pool must
// drive all work on the calling goroutine without spawning any helper.
// We assert by checking the caller goroutine ID is the only one observed
// inside the work function via runtime.Stack.
func TestWalkPool_SingleWorkerNoGoroutine(t *testing.T) {
	pool := newWalkPool(1)
	pool.submitMany([]string{"a", "b", "c"})

	caller := goroutineID()
	seen := make(map[int]struct{})
	work := func(dir string) []string {
		seen[goroutineID()] = struct{}{}
		return nil
	}
	pool.run(work)

	if _, ok := seen[caller]; !ok {
		t.Fatalf("work was not run on caller goroutine; seen=%v, caller=%d", seen, caller)
	}
	if len(seen) != 1 {
		t.Fatalf("workers=1 must run all work on a single goroutine; observed %d goroutines: %v", len(seen), seen)
	}
}

// goroutineID parses the current goroutine ID from runtime.Stack output.
// Used only by tests to verify the workers==1 specialization.
func goroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// Format: "goroutine 12 [running]:"
	s := string(buf[:n])
	const prefix = "goroutine "
	s = strings.TrimPrefix(s, prefix)
	end := strings.IndexByte(s, ' ')
	if end < 0 {
		return -1
	}
	id, err := strconv.Atoi(s[:end])
	if err != nil {
		return -1
	}
	return id
}
