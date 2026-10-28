package target

import (
	"context"
	"runtime"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"gotest.tools/v3/assert"
)

// Integration tests for gap-directory pruning: the canPruneDir predicate
// (config.go) wired into the discoverFilesOutsideProgramsForTest walk. Predicate unit tests live
// in ignore_pattern_test.go.

// dirAccessed reports whether any walked directory path is, or sits under, a
// path segment named seg. Segment-anchored ("/seg" suffix or "/seg/" infix) to
// avoid matching siblings like "target-x".
func dirAccessed(dirs []string, seg string) bool {
	for _, d := range dirs {
		if strings.HasSuffix(d, "/"+seg) || strings.Contains(d, "/"+seg+"/") {
			return true
		}
	}
	return false
}

type globWalkRecorder struct {
	vfs.FS
	filesVisited int
	roots        []string
}

func (f *globWalkRecorder) WalkDir(root string, fn vfs.WalkDirFunc) error {
	f.roots = append(f.roots, root)
	return f.FS.WalkDir(root, func(path string, entry vfs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			f.filesVisited++
		}
		return fn(path, entry, err)
	})
}

func TestExpandGlobPrunesShallowDirectories(t *testing.T) {
	layout := []string{"src/a.ts"}
	for i := range 100 {
		layout = append(layout, "src/unrelated/deep/file"+strconvI(i)+".ts")
	}
	root, paths := setupDiscoveryFixture(t, layout)
	fsys := &globWalkRecorder{FS: osvfs.FS()}
	expansions, err := expandGlobs(context.Background(), fsys, root, []string{"src/*.ts"})
	assert.NilError(t, err)
	assert.Equal(t, expansions[0].Root, tspath.ResolvePath(root, "src"))
	assert.DeepEqual(t, expansions[0].Files, []string{paths["src/a.ts"]})
	assert.Equal(t, fsys.filesVisited, 1)
}

func TestSelectInputsPreservesLiteralAndScanRoles(t *testing.T) {
	// The selector needs an enumerable VFS, not a host filesystem or content map.
	fsys := iovfs.From(fstest.MapFS{
		"project/src/a.ts":      &fstest.MapFile{},
		"project/src/b.ts":      &fstest.MapFile{},
		"project/src/[ab].ts":   &fstest.MapFile{},
		"project/other/deep.ts": &fstest.MapFile{},
	}, true)
	for _, inputs := range [][]string{
		{"src/@(a|b).ts", "src/*.ts", "src/./*.ts", "src/[ab].ts", "other", "missing*.ts"},
		{"src/[ab].ts", "src/*.ts", "other", "src/@(a|b).ts", "missing*.ts"},
	} {
		selection, err := SelectInputs(context.Background(), fsys, "/project", inputs)
		assert.NilError(t, err)
		sort.Strings(selection.Files)
		assert.DeepEqual(t, selection.Files, []string{"/project/src/[ab].ts", "/project/src/a.ts", "/project/src/b.ts"})
		assert.DeepEqual(t, selection.Unmatched, []string{"/project/missing*.ts"})
		assert.DeepEqual(t, selection.Directories, []string{"/project/other"})
		assert.DeepEqual(t, selection.GlobRoots, []string{"/project/src"})
		assert.DeepEqual(t, selection.ScanDirectories, map[string][]string{
			"/project/src/a.ts": {"/project/src"},
			"/project/src/b.ts": {"/project/src"},
		})
	}
}

func TestExpandGlobsGroupsOnlyEqualRoots(t *testing.T) {
	root, paths := setupDiscoveryFixture(t, []string{"src/a.ts", "src/nested/b.ts", "hash/#group/a.ts"})
	for _, test := range []struct {
		name     string
		patterns []string
		roots    []string
		files    [][]string
	}{
		{
			name:     "overlap and normalized duplicate",
			patterns: []string{"src/**/*.ts", "src/*.ts", "src/./*.ts"},
			roots:    []string{"src"},
			files:    [][]string{{"src/a.ts", "src/nested/b.ts"}, {"src/a.ts"}, {"src/a.ts"}},
		},
		{
			name:     "unmatched input remains empty",
			patterns: []string{"src/*.ts", "src/missing*.ts"},
			roots:    []string{"src"},
			files:    [][]string{{"src/a.ts"}, nil},
		},
		{
			name:     "hash prefix is a filename",
			patterns: []string{"hash/#*/*.ts"},
			roots:    []string{"hash"},
			files:    [][]string{{"hash/#group/a.ts"}},
		},
		{
			name:     "ancestor roots stay independent",
			patterns: []string{"src/**/*.ts", "src/nested/*.ts"},
			roots:    []string{"src", "src/nested"},
			files:    [][]string{{"src/a.ts", "src/nested/b.ts"}, {"src/nested/b.ts"}},
		},
		{
			name:     "escaped and literal roots share a walk",
			patterns: []string{`src/\nested/*.ts`, "src/nested/*.ts"},
			roots:    []string{"src/nested"},
			files:    [][]string{{"src/nested/b.ts"}, {"src/nested/b.ts"}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && strings.Contains(test.patterns[0], `\`) {
				t.Skip("backslash glob escapes are POSIX-only")
			}
			fsys := &globWalkRecorder{FS: osvfs.FS()}
			expansions, err := expandGlobs(context.Background(), fsys, root, test.patterns)
			assert.NilError(t, err)
			assert.Equal(t, len(expansions), len(test.patterns))
			var wantRoots []string
			for _, name := range test.roots {
				wantRoots = append(wantRoots, tspath.ResolvePath(root, name))
			}
			assert.DeepEqual(t, fsys.roots, wantRoots)
			for index, names := range test.files {
				var want []string
				for _, name := range names {
					want = append(want, paths[name])
				}
				assert.DeepEqual(t, expansions[index].Files, want)
			}
		})
	}
}

func TestRealWorldTargetWalkMatchesConfigSelection(t *testing.T) {
	layout := []string{
		"packages/core/src/index.ts",
		"packages/core/dist/bundle.ts",
		"packages/core/node_modules/dep/i.ts",
		"target/build/a.ts",
		"tests/rspack-test/configCases/pkg/node_modules/d.ts",
		"tests/rspack-test/configCases/c.ts",
		"scripts/build.ts",
		"npm/darwin-arm64/index.ts",
		"npm/win32-x64-msvc/index.ts",
		"npm/util.ts",
		"src/app/main.ts",
		"src/util.tsx",
	}
	configDir, paths := setupDiscoveryFixture(t, layout)
	config := rslintconfig.RslintConfig{
		{Ignores: []string{
			"**/tests/**",
			"**/dist/**/*",
			"**/node_modules/**/*",
			"!tests/rspack-test/*/**/node_modules",
			"**/target/**/*",
			"npm/**/*.node",
			"npm/*",
			"!npm/darwin-arm64/**/*",
		}},
		{Files: []string{"**/*.ts", "**/*.tsx"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	var oracle []string
	for _, filePath := range paths {
		if config.GetConfigForFile(filePath, configDir) != nil {
			oracle = append(oracle, filePath)
		}
	}
	sort.Strings(oracle)

	got := discoverFilesOutsideProgramsForTest(
		config,
		configDir,
		osvfs.FS(),
		map[string]struct{}{},
		nil,
		nil,
		false,
	)
	sort.Strings(got)
	assert.DeepEqual(t, got, oracle)

	want := []string{
		paths["npm/darwin-arm64/index.ts"],
		paths["npm/win32-x64-msvc/index.ts"],
		paths["packages/core/src/index.ts"],
		paths["scripts/build.ts"],
		paths["src/app/main.ts"],
		paths["src/util.tsx"],
	}
	sort.Strings(want)
	assert.DeepEqual(t, got, want)
}

func setupDiscoveryTxtarFixture(t *testing.T, name string) (string, map[string]string) {
	t.Helper()
	archive := txtarfs.MustParseFile(t, "testdata/file_discovery.txtar")
	names, err := archive.FileNames(name)
	if err != nil {
		t.Fatal(err)
	}
	configDir := osvfs.FS().Realpath(tspath.NormalizePath(archive.Materialize(t, name)))
	paths := make(map[string]string, len(names))
	for _, relativeName := range names {
		paths[relativeName] = tspath.ResolvePath(configDir, relativeName)
	}
	return configDir, paths
}

// Core fix: a gitignore file-level dir (target/ → **/target/**/*) is pruned
// during the gap walk, and the gap-file set is unchanged.
func TestDiscoverFilesOutsidePrograms_PrunesGitignoreFileLevelDir(t *testing.T) {
	configDir, paths := setupDiscoveryTxtarFixture(t, "prunes-gitignore")
	// Simulate gitignore `target/` → file-level glob (what convertSinglePattern emits).
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/target/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

	// target/ must NOT be entered.
	for _, dir := range spy.snapshotAccessedDirs() {
		if strings.Contains(dir, "target") {
			t.Errorf("target was entered during walk: %s", dir)
		}
	}
	// gapFiles == exactly src/index.ts.
	assert.Equal(t, len(gapFiles), 1, "got %v", gapFiles)
	assert.Assert(t, toSet(gapFiles)[paths["src/index.ts"]])
}

// Negation re-includes a full path (rspack's !tests/.../target case): the
// top-level target is pruned, but the re-included path is walked.
func TestDiscoverFilesOutsidePrograms_NegationReincludeFullPath(t *testing.T) {
	configDir, paths := setupDiscoveryTxtarFixture(t, "negation-full-path")
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/target/**/*", "!sub/path/target/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)
	dirs := spy.snapshotAccessedDirs()

	// Top-level target NOT entered; the re-included sub/path/target IS entered.
	for _, d := range dirs {
		if strings.HasSuffix(d, "/target") && !strings.Contains(d, "sub/path") {
			t.Errorf("top-level target should be pruned, but entered: %s", d)
		}
	}
	assert.Assert(t, dirAccessed(dirs, "sub/path/target"), "re-included target must be walked")

	gapSet := toSet(gapFiles)
	assert.Assert(t, gapSet[paths["src/a.ts"]])
	assert.Assert(t, gapSet[paths["sub/path/target/y.ts"]], "re-included file must be a gap file")
	assert.Assert(t, !gapSet[paths["target/x.ts"]], "top-level target file must stay ignored")
}

// Negation re-includes a child of an excluded directory: the parent must NOT
// be pruned (rslint's file-level isFileIgnored re-includes the child).
func TestDiscoverFilesOutsidePrograms_NegationReincludeChildNotOverPruned(t *testing.T) {
	configDir, paths := setupDiscoveryTxtarFixture(t, "negation-child")
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"target/**/*", "!target/keep/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

	// Must reach target/keep.
	assert.Assert(t, dirAccessed(spy.snapshotAccessedDirs(), "target/keep"), "target/keep must be walked")

	gapSet := toSet(gapFiles)
	assert.Assert(t, gapSet[paths["src/a.ts"]])
	assert.Assert(t, gapSet[paths["target/keep/x.ts"]], "re-included child must be a gap file")
	assert.Assert(t, !gapSet[paths["target/other/y.ts"]], "non-negated sibling must stay ignored")
}

// Unrooted negation (!**/keep/) forces conservative behavior: the file-level
// directory is not pruned (a keep/ could appear at any depth inside it).
func TestDiscoverFilesOutsidePrograms_UnrootedNegationConservative(t *testing.T) {
	configDir, paths := setupDiscoveryTxtarFixture(t, "unrooted-negation")
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/build/**/*", "!**/keep/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

	assert.Assert(t, dirAccessed(spy.snapshotAccessedDirs(), "build"), "build must be walked (unrooted negation)")

	gapSet := toSet(gapFiles)
	assert.Assert(t, gapSet[paths["src/a.ts"]])
	assert.Assert(t, gapSet[paths["build/keep/x.ts"]], "unrooted negation re-includes build/keep")
	assert.Assert(t, !gapSet[paths["build/other/y.ts"]])
}

// A case-insensitive anchored negation protects only its own subtree. It must
// not disable pruning for unrelated file-level covers, which previously made
// macOS/Windows walks enter every ignored directory in large repositories.
func TestDiscoverFilesOutsidePrograms_CaseInsensitiveAnchoredNegationPrunesUnrelatedDir(t *testing.T) {
	configDir, paths := setupDiscoveryTxtarFixture(t, "case-insensitive")
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &caseInsensitiveSpyFS{spyFS: &spyFS{FS: osvfs.FS()}}
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, true)
	dirs := spy.snapshotAccessedDirs()

	assert.Assert(t, !dirAccessed(dirs, "target"), "unrelated target must be pruned")
	assert.Assert(t, dirAccessed(dirs, "scripts/debug"), "case-folded re-included subtree must be walked")

	gapSet := toSet(gapFiles)
	assert.Assert(t, gapSet[paths["src/a.ts"]])
	assert.Assert(t, gapSet[paths["scripts/debug/keep.ts"]], "case-folded negation must re-include debug")
	assert.Assert(t, !gapSet[paths["scripts/other/drop.ts"]], "non-negated sibling must remain ignored")
	assert.Assert(t, !gapSet[paths["target/deep/x.ts"]], "unrelated ignored file must stay excluded")
}

// Directory-level `dir/**` (absolute, not negatable) vs file-level `dir/**/*`
// (negation-aware): pruning behavior differs and stays aligned with ESLint v10.
func TestDiscoverFilesOutsidePrograms_DirLevelVsFileLevel(t *testing.T) {
	// 6a: dir-level — absolutely pruned; ! cannot re-include.
	t.Run("dir-level absolute", func(t *testing.T) {
		configDir, paths := setupDiscoveryTxtarFixture(t, "dir-level")
		config := rslintconfig.RslintConfig{
			{Ignores: []string{"dist/**", "!dist/keep.ts"}},
			{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
		}
		spy := &spyFS{FS: osvfs.FS()}
		gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)
		for _, d := range spy.snapshotAccessedDirs() {
			if strings.Contains(d, "dist") {
				t.Errorf("dir-level dist must be absolutely pruned, entered: %s", d)
			}
		}
		gapSet := toSet(gapFiles)
		assert.Assert(t, gapSet[paths["src/a.ts"]])
		assert.Assert(t, !gapSet[paths["dist/keep.ts"]], "dir-level ! cannot re-include")
	})

	// 6b: file-level — not pruned (negation protects keep.ts).
	t.Run("file-level negation-aware", func(t *testing.T) {
		configDir, paths := setupDiscoveryTxtarFixture(t, "dir-level")
		config := rslintconfig.RslintConfig{
			{Ignores: []string{"dist/**/*", "!dist/keep.ts"}},
			{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
		}
		spy := &spyFS{FS: osvfs.FS()}
		gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)
		assert.Assert(t, dirAccessed(spy.snapshotAccessedDirs(), "dist"), "file-level dist must be walked for negation")
		gapSet := toSet(gapFiles)
		assert.Assert(t, gapSet[paths["dist/keep.ts"]], "file-level ! re-includes keep.ts")
		assert.Assert(t, !gapSet[paths["dist/other.ts"]])
	})
}

// Real gitignore conversion path: `target/` in a .gitignore prunes target/.
// Uses a spy to assert the pruning actually happens (not just that the gap-file
// set is correct, which would pass even with pruning disabled).
func TestDiscoverFilesOutsidePrograms_GitignoreTargetPrunedE2E(t *testing.T) {
	dir, _ := setupDiscoveryTxtarFixture(t, "gitignore-target")
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := discoverFilesOutsideProgramsForTest(config, dir, spy, map[string]struct{}{}, nil, nil, false)

	// The actual optimization: target/ must NOT be entered.
	assert.Assert(t, !dirAccessed(spy.snapshotAccessedDirs(), "target"), "target should be pruned via gitignore")
	// gap-file set unchanged.
	assert.Equal(t, len(gapFiles), 1, "got %v", gapFiles)
	assert.Assert(t, toSet(gapFiles)[tspath.NormalizePath(dir+"/src/index.ts")])
	// Linter consistency: target files return nil.
	assert.Assert(t, configWithGitForFileForTest(config, tspath.NormalizePath(dir+"/target/a.ts"), dir) == nil)
}

// Nested .gitignore negation: root ignores build/, a sub/.gitignore re-includes
// it. The re-included subtree must be walked and discovered; the top-level
// build/ must still be pruned. Exercises the conversion → negPrefix → prune
// chain for nested-gitignore negations end to end.
func TestDiscoverFilesOutsidePrograms_NestedGitignoreNegationE2E(t *testing.T) {
	dir, _ := setupDiscoveryTxtarFixture(t, "nested-gitignore")
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := discoverFilesOutsideProgramsForTest(config, dir, spy, map[string]struct{}{}, nil, nil, false)
	gapSet := toSet(gapFiles)

	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/src/a.ts")])
	// Re-included nested build must be walked + discovered.
	assert.Assert(t, dirAccessed(spy.snapshotAccessedDirs(), "build"), "sub/build must be walked")
	if !gapSet[tspath.NormalizePath(dir+"/sub/build/keep.ts")] {
		// Consistency cross-check: linter must agree it is lintable.
		mc := configWithGitForFileForTest(config, tspath.NormalizePath(dir+"/sub/build/keep.ts"), dir)
		t.Errorf("sub/build/keep.ts should be a gap file; GetConfigForFile=%v", mc != nil)
	}
	// Top-level build/top.ts stays ignored.
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/build/top.ts")], "top-level build stays ignored")
}

// --- Strongest regression: pruning must not change the gap-file set ---
//
// Oracle = { f : f matches **/*.ts ∧ f∉programFiles ∧ GetConfigForFile(f)≠nil }.
// This is exactly the linter's per-file decision; discoverFilesOutsideProgramsForTest must equal
// it regardless of directory pruning.
func TestDiscoverFilesOutsidePrograms_PruningPreservesGapFiles(t *testing.T) {
	filesPatterns := []string{"**/*.ts", "**/*.tsx"}
	fixtures := []struct {
		name    string
		ignores []string
	}{
		{"gitignore target", []string{"**/target/**/*"}},
		{"negation full path", []string{"**/target/**/*", "!sub/path/target/**/*"}},
		{"negation child of excluded", []string{"target/**/*", "!target/keep/**/*"}},
		{"unrooted negation", []string{"**/build/**/*", "!**/keep/**/*"}},
		{"dir-level absolute", []string{"dist/**", "!dist/keep.ts"}},
		{"single-star file-level", []string{"build/*"}},                        // regression: 致命#1
		{"extension-filtered", []string{"target/**/*.log"}},                    // regression: ext filter must not prune
		{"dot slash negation", []string{"target/**/*", "!./target/keep/**/*"}}, // regression: 致命#2
		{"mixed", []string{"**/target/**/*", "**/tests/**", "!tests/e2e/**/*"}},
		{"bare rooted", []string{"dist"}},                            // /dist → "dist": no /**/* suffix, must not prune
		{"deep dir-only", []string{"a/b/target/**/*"}},               // deep-path positive cover
		{"brace extension filter", []string{"target/**/*.{js,jsx}"}}, // brace ext filter must not prune .ts
		{"multi-negation", []string{"**/target/**/*", "!target/keep/**/*", "!target/save/**/*"}},
		{"sequential re-ignore", []string{"target/**/*", "!target/keep/**/*", "target/keep/sub/**/*"}},
	}

	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			// The archive layout mixes shallow and deep files, .tsx, and a
			// non-matching .log so each pruning pattern is checked against the
			// linter's per-file decision over the same filesystem shape.
			configDir, paths := setupDiscoveryTxtarFixture(t, "pruning-oracle")
			config := rslintconfig.RslintConfig{
				{Ignores: fx.ignores},
				{Files: filesPatterns, Rules: rslintconfig.Rules{"test-rule": "error"}},
			}

			// Oracle = the linter's own per-file decision: matches a files
			// pattern AND GetConfigForFile != nil. discoverFilesOutsideProgramsForTest must equal
			// this set regardless of directory pruning.
			var oracle []string
			for _, abs := range paths {
				if configWithGitForFileForTest(config, abs, configDir) != nil {
					oracle = append(oracle, abs)
				}
			}
			sort.Strings(oracle)

			got := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, false)
			sort.Strings(got)

			assert.DeepEqual(t, got, oracle)
		})
	}
}

// A `!` negation inside a NON-global config entry (one carrying Files/rslintconfig.Rules)
// must not resurrect a globally-ignored file: GetConfigForFile evaluates global
// ignores first, so entry-level ignores can only narrow. canPruneDir sees
// only the global ignores and prunes target/ — which stays consistent with the
// linter (it also excludes target/keep). Locks down the "per-entry config
// cannot cause over-prune" invariant.
func TestDiscoverFilesOutsidePrograms_PerEntryNegationDoesNotResurrect(t *testing.T) {
	configDir, paths := setupDiscoveryTxtarFixture(t, "per-entry-negation")
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/target/**/*"}},
		{Files: []string{"**/*.ts"}, Ignores: []string{"!target/keep/**/*"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := discoverFilesOutsideProgramsForTest(config, configDir, spy, map[string]struct{}{}, nil, nil, false)
	gapSet := toSet(gapFiles)

	assert.Assert(t, gapSet[paths["src/a.ts"]])
	assert.Assert(t, !gapSet[paths["target/x.ts"]])
	assert.Assert(t, !gapSet[paths["target/keep/k.ts"]], "per-entry ! must not resurrect a global ignore")
	// Linter authority agrees, and target/ is pruned (consistent).
	assert.Assert(t, configWithGitForFileForTest(config, paths["target/keep/k.ts"], configDir) == nil)
	assert.Assert(t, !dirAccessed(spy.snapshotAccessedDirs(), "target"), "target pruned; per-entry ! does not protect")
}

// Pruning must produce identical gap files in parallel and single-threaded mode.
func TestDiscoverFilesOutsidePrograms_PruneSingleThreadedEquivalence(t *testing.T) {
	configDir, _ := setupDiscoveryTxtarFixture(t, "single-threaded")
	config := rslintconfig.RslintConfig{
		{Ignores: []string{"**/target/**/*", "!sub/path/target/**/*", "!target/keep/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: rslintconfig.Rules{"test-rule": "error"}},
	}

	par := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, false)
	seq := discoverFilesOutsideProgramsForTest(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, true)
	sort.Strings(par)
	sort.Strings(seq)
	assert.DeepEqual(t, par, seq)
}
