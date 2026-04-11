package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

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

func TestDiscoverGapFiles_Basic(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"scripts/b.ts",
	})

	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	// src/a.ts is in the program, scripts/b.ts is not
	programFiles := map[string]struct{}{
		paths["src/a.ts"]: {},
	}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

	assert.Assert(t, gapFiles != nil, "should not be nil when config has files")
	assert.Equal(t, len(gapFiles), 1)
	assert.Equal(t, gapFiles[0], paths["scripts/b.ts"])
}

func TestDiscoverGapFiles_GlobalIgnoresExclude(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"scripts/b.ts",
	})

	config := RslintConfig{
		// Global ignore
		{Ignores: []string{"scripts/**"}},
		// Rules entry
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{
		paths["src/a.ts"]: {},
	}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

	// scripts/b.ts should be excluded by global ignores
	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 0)
}

func TestDiscoverGapFiles_ProgramFilesSkipped(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/b.ts",
	})

	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	// Both files are in the program
	programFiles := map[string]struct{}{
		paths["src/a.ts"]: {},
		paths["src/b.ts"]: {},
	}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 0)
}

func TestDiscoverGapFiles_GetConfigForFilePreFilter(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"test/b.ts",
	})

	config := RslintConfig{
		// Entry with files AND entry-level ignores
		{
			Files:   []string{"**/*.ts"},
			Ignores: []string{"test/**"},
			Rules:   Rules{"test-rule": "error"},
		},
	}

	programFiles := map[string]struct{}{
		paths["src/a.ts"]: {},
	}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

	// test/b.ts matches **/*.ts but is excluded by entry-level ignores →
	// GetConfigForFile returns nil → not a gap file
	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 0)
}

func TestDiscoverGapFiles_NoFilesField_ReturnsNil(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/a.ts",
	})

	// JSON-style config: no files field
	config := RslintConfig{
		{Rules: Rules{"test-rule": "error"}},
	}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil)

	// Should return nil (backward compat signal)
	assert.Assert(t, gapFiles == nil, "should return nil when no entry has files field")
}

func TestDiscoverGapFiles_AllowDirsScope(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"scripts/b.ts",
		"tools/c.ts",
	})

	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	// Only allow scripts/ directory
	scriptsDir := tspath.NormalizePath(filepath.Join(configDir, "scripts"))
	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, []string{scriptsDir})

	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 1)
	assert.Equal(t, gapFiles[0], paths["scripts/b.ts"])
}

func TestDiscoverGapFiles_MultipleFilesPatterns(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/b.tsx",
		"src/c.js",
	})

	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"rule-a": "error"}},
		{Files: []string{"**/*.tsx"}, Rules: Rules{"rule-b": "error"}},
		// .js files have no matching entry
	}

	programFiles := map[string]struct{}{}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

	assert.Assert(t, gapFiles != nil)
	sort.Strings(gapFiles)

	expected := []string{paths["src/a.ts"], paths["src/b.tsx"]}
	sort.Strings(expected)

	assert.Equal(t, len(gapFiles), len(expected))
	for i := range expected {
		assert.Equal(t, gapFiles[i], expected[i])
	}
}

func TestDiscoverGapFiles_JsFilesNotDiscoveredWithoutPattern(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/b.js",
	})

	// Only **/*.ts in files, no **/*.js
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

	assert.Assert(t, gapFiles != nil)
	// b.js should NOT be discovered because no entry has files: ['**/*.js']
	assert.Equal(t, len(gapFiles), 1)
}

func TestDiscoverGapFiles_AllowFilesScope(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"scripts/b.ts",
		"tools/c.ts",
	})

	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	// Only allow scripts/b.ts via allowFiles
	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles,
		[]string{paths["scripts/b.ts"]}, nil)

	assert.Assert(t, gapFiles != nil)
	assert.Equal(t, len(gapFiles), 1)
	assert.Equal(t, gapFiles[0], paths["scripts/b.ts"])
}

func TestDiscoverGapFiles_AllExtensionsDiscoveredByPattern(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"src/readme.md",
		"src/data.json",
	})

	// files: ['**/*'] matches all files — no extension filtering
	config := RslintConfig{
		{Files: []string{"**/*"}, Rules: Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

	assert.Assert(t, gapFiles != nil)
	// All files matching the pattern should be discovered (no extension filter)
	assert.Equal(t, len(gapFiles), 3)
}

func TestDiscoverGapFilesMultiConfig(t *testing.T) {
	configDir1, paths1 := setupDiscoveryFixture(t, []string{
		"a.ts",
	})
	configDir2, paths2 := setupDiscoveryFixture(t, []string{
		"b.tsx",
	})

	configMap := map[string]RslintConfig{
		configDir1: {
			{Files: []string{"**/*.ts"}, Rules: Rules{"rule-a": "error"}},
		},
		configDir2: {
			{Files: []string{"**/*.tsx"}, Rules: Rules{"rule-b": "error"}},
		},
	}

	programFiles := map[string]struct{}{}

	gapFiles := DiscoverGapFilesMultiConfig(configMap, osvfs.FS(), programFiles, nil, nil)

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

func TestDiscoverGapFiles_EmptyFilesArray(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/a.ts",
	})

	// Entry with files: [] (empty array, not absent)
	config := RslintConfig{
		{Files: []string{}, Rules: Rules{"test-rule": "error"}},
	}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil)

	// Empty files array has no patterns to match → same as no files field → nil (legacy mode)
	assert.Assert(t, gapFiles == nil, "should return nil for empty files array")
}

func TestDiscoverGapFiles_FilesButNoRules(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/a.ts",
	})

	// Entry has files but no rules → GetConfigForFile returns MergedConfig
	// with empty rules → but entryMatched is true → not nil
	// However, linter will skip (len(rules) == 0)
	config := RslintConfig{
		{Files: []string{"**/*.ts"}},
	}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil)

	// GetConfigForFile returns non-nil (entry matched), so the file IS a gap file.
	// The linter will subsequently skip it because it has no rules, but that's
	// the linter's concern, not discovery's.
	assert.Assert(t, gapFiles != nil)
}

// --- Directory-level ignore blocking in DiscoverGapFiles ---

func TestDiscoverGapFiles_DirIgnoreBlocksTraversal(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"build/keep.ts",
		"build/other.ts",
	})

	// build/** is directory-level → blocks traversal → ! cannot re-include
	config := RslintConfig{
		{Ignores: []string{"build/**", "!build/keep.ts"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

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

func TestDiscoverGapFiles_FileIgnoreAllowsNegation(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"build/keep.ts",
		"build/other.ts",
	})

	// build/**/* is file-level → does NOT block traversal → ! CAN re-include
	config := RslintConfig{
		{Ignores: []string{"build/**/*", "!build/keep.ts"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

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

func TestDiscoverGapFiles_WildcardMiddleDirIgnoreBlocks(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"packages/app/dist/gen.ts",
		"packages/app/src/index.ts",
	})

	config := RslintConfig{
		{Ignores: []string{"packages/*/dist/**"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}
	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

	gapSet := make(map[string]struct{})
	for _, f := range gapFiles {
		gapSet[f] = struct{}{}
	}

	_, hasDist := gapSet[paths["packages/app/dist/gen.ts"]]
	assert.Assert(t, !hasDist, "packages/app/dist/ should be blocked by packages/*/dist/**")

	_, hasSrc := gapSet[paths["packages/app/src/index.ts"]]
	assert.Assert(t, hasSrc, "packages/app/src/ should not be blocked")
}

func TestDiscoverGapFiles_CrossEntryDirIgnoreAndNegation(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"build/keep.ts",
		"build/other.ts",
	})

	// Entry 1: dir-level ignore. Entry 2: negation.
	// dir/** blocks traversal → negation cannot re-include (even across entries,
	// because global ignores are merged and dir-level patterns block first).
	config := RslintConfig{
		{Ignores: []string{"build/**"}},
		{Ignores: []string{"!build/keep.ts"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}
	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

	for _, f := range gapFiles {
		if f == paths["build/keep.ts"] || f == paths["build/other.ts"] {
			t.Errorf("Expected build/ blocked by dir/** even with cross-entry negation, but found %s", f)
		}
	}
}

func TestDiscoverGapFiles_DoubleStarDirIgnoreBlocksNested(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"packages/app/dist/gen.ts",
	})

	// **/dist/** blocks any dist/ directory
	config := RslintConfig{
		{Ignores: []string{"**/dist/**"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	programFiles := map[string]struct{}{}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), programFiles, nil, nil)

	assert.Assert(t, gapFiles != nil)
	for _, f := range gapFiles {
		if f == paths["packages/app/dist/gen.ts"] {
			t.Error("Expected dist/ files to be blocked by **/dist/** ignore")
		}
	}
}

// --- Default excludes (node_modules, .git) ---

func TestDiscoverGapFiles_DefaultExcludesNodeModules(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"node_modules/pkg/index.ts",
	})

	// No global ignores at all — defaults should still exclude node_modules
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil)

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

func TestDiscoverGapFiles_DefaultExcludesGitDir(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		".git/hooks/pre-commit.ts",
	})

	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	gapFiles := DiscoverGapFiles(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil)

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
// GetAccessibleEntries is called by vfsAdapter.ReadDir only when fs.WalkDir
// actually enters a directory. If a directory is pruned (fs.SkipDir), this
// method is never called for it.
type spyFS struct {
	vfs.FS
	accessedDirs []string
}

func (s *spyFS) GetAccessibleEntries(path string) vfs.Entries {
	s.accessedDirs = append(s.accessedDirs, path)
	return s.FS.GetAccessibleEntries(path)
}

func TestDiscoverGapFiles_PrunesNodeModulesAtWalkLevel(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"node_modules/pkg/index.ts",
		"node_modules/pkg/nested/deep.ts",
	})

	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil)

	for _, dir := range spy.accessedDirs {
		if strings.Contains(dir, "node_modules") {
			t.Errorf("node_modules directory was entered during walk (GetAccessibleEntries called for %s)", dir)
		}
	}
}

func TestDiscoverGapFiles_PrunesGitDirAtWalkLevel(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		".git/objects/ab/file.ts",
	})

	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil)

	for _, dir := range spy.accessedDirs {
		if strings.Contains(dir, ".git") {
			t.Errorf(".git directory was entered during walk (GetAccessibleEntries called for %s)", dir)
		}
	}
}

func TestDiscoverGapFiles_PrunesUserIgnoredDirAtWalkLevel(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"vendor/lib/util.ts",
		"vendor/lib/nested/deep.ts",
	})

	config := RslintConfig{
		{Ignores: []string{"vendor/**"}}, // global ignore
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil)

	for _, dir := range spy.accessedDirs {
		if strings.Contains(dir, "vendor") {
			t.Errorf("vendor directory was entered during walk (GetAccessibleEntries called for %s)", dir)
		}
	}
}

func TestDiscoverGapFiles_PrunesNestedIgnoredDirButEntersParent(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"build/index.ts",
		"build/output/bundle.ts",
		"build/output/nested/deep.ts",
	})

	config := RslintConfig{
		{Ignores: []string{"build/output/**"}}, // only output/ ignored, not build/
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil)

	// build/ should be entered (not blocked)
	buildEntered := false
	outputEntered := false
	for _, dir := range spy.accessedDirs {
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

func TestDiscoverGapFiles_EntersNonExcludedDirs(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"src/components/button.ts",
		"lib/utils.ts",
	})

	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil)

	// src/ and lib/ should be entered
	srcEntered := false
	libEntered := false
	for _, dir := range spy.accessedDirs {
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
//   1. ReadGitignoreAsGlobs (with configIgnores for pruning)
//   2. Inject gitignore globs into config
//   3. DiscoverGapFiles
//   4. Verify GetConfigForFile (linter's per-file decision) is consistent
//
// The structural guarantee being tested: if isDirPathBlocked(dir, configIgnores)
// returns true in collectGitignoreGlobs (causing .gitignore skip), then
// GetConfigForFile also returns nil for any file in that dir.
// =============================================================================

// e2eSetup creates a fixture, runs ReadGitignoreAsGlobs + config injection + DiscoverGapFiles,
// and returns gap files + the final config (for GetConfigForFile verification).
func e2eSetup(t *testing.T, files map[string]string, config RslintConfig, programFiles map[string]struct{}) (string, []string, RslintConfig) {
	t.Helper()
	dir := setupGitignoreFixture(t, files)

	configIgnores := ExtractConfigIgnores(config)
	gitGlobs := ReadGitignoreAsGlobs(dir, osvfs.FS(), configIgnores)
	if len(gitGlobs) > 0 {
		config = append(RslintConfig{{Ignores: gitGlobs}}, config...)
	}

	if programFiles == nil {
		programFiles = map[string]struct{}{}
	}

	gapFiles := DiscoverGapFiles(config, dir, osvfs.FS(), programFiles, nil, nil)
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
	config := RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
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
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	// Build programFiles with both files (simulating tsconfig include: ["src"])
	dir := setupGitignoreFixture(t, files)
	indexPath := tspath.NormalizePath(dir + "/packages/foo/src/index.ts")
	genPathFull := tspath.NormalizePath(dir + "/packages/foo/src/generated/api.ts")

	configIgnores := ExtractConfigIgnores(config)
	gitGlobs := ReadGitignoreAsGlobs(dir, osvfs.FS(), configIgnores)
	if len(gitGlobs) > 0 {
		config = append(RslintConfig{{Ignores: gitGlobs}}, config...)
	}

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
		"src/index.ts":        "x",
		"tests/unit/a.ts":     "x",
		"tests/e2e/smoke.ts":  "x",
		"tests/.gitignore":    "tmp/\n",
		"tests/e2e/.gitignore": "screenshots/\n",
	}
	config := RslintConfig{
		{Ignores: []string{"**/tests/**/*", "!tests/e2e/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
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
	config := RslintConfig{
		{Ignores: []string{"**/vendor/**"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
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
	config := RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
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
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	// Create fixture once and use the same dir for programFiles and e2e flow.
	dir := setupGitignoreFixture(t, files)
	indexPath := tspath.NormalizePath(dir + "/src/index.ts")
	utilsPath := tspath.NormalizePath(dir + "/src/utils.ts")

	programFiles := map[string]struct{}{
		indexPath: {},
	}

	configIgnores := ExtractConfigIgnores(config)
	gitGlobs := ReadGitignoreAsGlobs(dir, osvfs.FS(), configIgnores)
	if len(gitGlobs) > 0 {
		config = append(RslintConfig{{Ignores: gitGlobs}}, config...)
	}

	gapFiles := DiscoverGapFiles(config, dir, osvfs.FS(), programFiles, nil, nil)
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
	config := RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Ignores: []string{"scripts/**"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
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
		"src/index.ts":             "x",
		"tests/helpers/setup.ts":   "x",
	}
	config := RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	dir := setupGitignoreFixture(t, files)
	testFile := tspath.NormalizePath(dir + "/tests/helpers/setup.ts")

	configIgnores := ExtractConfigIgnores(config)
	gitGlobs := ReadGitignoreAsGlobs(dir, osvfs.FS(), configIgnores)
	if len(gitGlobs) > 0 {
		config = append(RslintConfig{{Ignores: gitGlobs}}, config...)
	}

	// Even though setup.ts is in programFiles, GetConfigForFile should return nil
	// because tests/ is directory-blocked by config ignores.
	mc := config.GetConfigForFile(testFile, dir)
	assert.Assert(t, mc == nil, "GetConfigForFile should return nil for file in config-ignored dir, even if in TS Program")
}

// E2E case 9: gitignore and config ignore both target the same directory (dist/).
// Both mechanisms should work — the file should be excluded regardless of which one catches it.
func TestE2E_OverlappingGitignoreAndConfigIgnore(t *testing.T) {
	files := map[string]string{
		".gitignore":       "dist/\n",
		"src/index.ts":     "x",
		"dist/bundle.ts":   "x",
	}
	config := RslintConfig{
		{Ignores: []string{"**/dist/**"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
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
		".gitignore":              "dist/\n",
		"packages/foo/src/a.ts":  "x",
		"packages/foo/dist/b.ts": "x",
		"packages/bar/src/c.ts":  "x",
		"tests/unit/d.ts":        "x",
	}
	config := RslintConfig{
		{Ignores: []string{"**/tests/**"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	dir := setupGitignoreFixture(t, files)
	configIgnores := ExtractConfigIgnores(config)
	gitGlobs := ReadGitignoreAsGlobs(dir, osvfs.FS(), configIgnores)
	if len(gitGlobs) > 0 {
		config = append(RslintConfig{{Ignores: gitGlobs}}, config...)
	}

	// Only allow packages/foo/
	fooDir := tspath.NormalizePath(dir + "/packages/foo")
	gapFiles := DiscoverGapFiles(config, dir, osvfs.FS(), map[string]struct{}{}, nil, []string{fooDir})
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
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	dir := setupGitignoreFixture(t, files)
	configIgnores := ExtractConfigIgnores(config)
	gitGlobs := ReadGitignoreAsGlobs(dir, osvfs.FS(), configIgnores)
	if len(gitGlobs) > 0 {
		config = append(RslintConfig{{Ignores: gitGlobs}}, config...)
	}

	srcFile := tspath.NormalizePath(dir + "/src/index.ts")
	distFile := tspath.NormalizePath(dir + "/dist/bundle.ts")

	// Simulate lint-staged passing both files explicitly.
	gapFiles := DiscoverGapFiles(config, dir, osvfs.FS(), map[string]struct{}{}, []string{srcFile, distFile}, nil)
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
