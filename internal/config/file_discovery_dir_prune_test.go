package config

import (
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

// Integration tests for gap-directory pruning: the canPruneDir predicate
// (config.go) wired into the DiscoverGapFiles walk. Predicate unit tests live
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

// Core fix: a gitignore file-level dir (target/ → **/target/**/*) is pruned
// during the gap walk, and the gap-file set is unchanged.
func TestDiscoverGapFiles_PrunesGitignoreFileLevelDir(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/index.ts",
		"target/build/a.ts",
		"target/build/deep/b.ts",
	})
	// Simulate gitignore `target/` → file-level glob (what convertSinglePattern emits).
	config := RslintConfig{
		{Ignores: []string{"**/target/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

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
func TestDiscoverGapFiles_NegationReincludeFullPath(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"target/x.ts",
		"sub/path/target/y.ts",
	})
	config := RslintConfig{
		{Ignores: []string{"**/target/**/*", "!sub/path/target/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil, false)
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
func TestDiscoverGapFiles_NegationReincludeChildNotOverPruned(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"target/keep/x.ts",
		"target/other/y.ts",
	})
	config := RslintConfig{
		{Ignores: []string{"target/**/*", "!target/keep/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

	// Must reach target/keep.
	assert.Assert(t, dirAccessed(spy.snapshotAccessedDirs(), "target/keep"), "target/keep must be walked")

	gapSet := toSet(gapFiles)
	assert.Assert(t, gapSet[paths["src/a.ts"]])
	assert.Assert(t, gapSet[paths["target/keep/x.ts"]], "re-included child must be a gap file")
	assert.Assert(t, !gapSet[paths["target/other/y.ts"]], "non-negated sibling must stay ignored")
}

// Unrooted negation (!**/keep/) forces conservative behavior: the file-level
// directory is not pruned (a keep/ could appear at any depth inside it).
func TestDiscoverGapFiles_UnrootedNegationConservative(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"build/keep/x.ts",
		"build/other/y.ts",
	})
	config := RslintConfig{
		{Ignores: []string{"**/build/**/*", "!**/keep/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil, false)

	assert.Assert(t, dirAccessed(spy.snapshotAccessedDirs(), "build"), "build must be walked (unrooted negation)")

	gapSet := toSet(gapFiles)
	assert.Assert(t, gapSet[paths["src/a.ts"]])
	assert.Assert(t, gapSet[paths["build/keep/x.ts"]], "unrooted negation re-includes build/keep")
	assert.Assert(t, !gapSet[paths["build/other/y.ts"]])
}

// Directory-level `dir/**` (absolute, not negatable) vs file-level `dir/**/*`
// (negation-aware): pruning behavior differs and stays aligned with ESLint v10.
func TestDiscoverGapFiles_DirLevelVsFileLevel(t *testing.T) {
	// 6a: dir-level — absolutely pruned; ! cannot re-include.
	t.Run("dir-level absolute", func(t *testing.T) {
		configDir, paths := setupDiscoveryFixture(t, []string{
			"src/a.ts", "dist/keep.ts", "dist/other.ts",
		})
		config := RslintConfig{
			{Ignores: []string{"dist/**", "!dist/keep.ts"}},
			{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
		}
		spy := &spyFS{FS: osvfs.FS()}
		gapFiles := DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil, false)
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
		configDir, paths := setupDiscoveryFixture(t, []string{
			"src/a.ts", "dist/keep.ts", "dist/other.ts",
		})
		config := RslintConfig{
			{Ignores: []string{"dist/**/*", "!dist/keep.ts"}},
			{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
		}
		spy := &spyFS{FS: osvfs.FS()}
		gapFiles := DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil, false)
		assert.Assert(t, dirAccessed(spy.snapshotAccessedDirs(), "dist"), "file-level dist must be walked for negation")
		gapSet := toSet(gapFiles)
		assert.Assert(t, gapSet[paths["dist/keep.ts"]], "file-level ! re-includes keep.ts")
		assert.Assert(t, !gapSet[paths["dist/other.ts"]])
	})
}

// Real gitignore conversion path: `target/` in a .gitignore prunes target/.
// Uses a spy to assert the pruning actually happens (not just that the gap-file
// set is correct, which would pass even with pruning disabled).
func TestDiscoverGapFiles_GitignoreTargetPrunedE2E(t *testing.T) {
	files := map[string]string{
		".gitignore":       "target/\n",
		"src/index.ts":     "x",
		"target/a.ts":      "x",
		"target/deep/b.ts": "x",
	}
	dir := setupDiscoveryContentFixture(t, files)
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}
	config = ConfigWithGitignore(config, dir, osvfs.FS(), nil)

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := DiscoverGapFiles(config, dir, spy, map[string]struct{}{}, nil, nil, false)

	// The actual optimization: target/ must NOT be entered.
	assert.Assert(t, !dirAccessed(spy.snapshotAccessedDirs(), "target"), "target should be pruned via gitignore")
	// gap-file set unchanged.
	assert.Equal(t, len(gapFiles), 1, "got %v", gapFiles)
	assert.Assert(t, toSet(gapFiles)[tspath.NormalizePath(dir+"/src/index.ts")])
	// Linter consistency: target files return nil.
	assert.Assert(t, config.GetConfigForFile(tspath.NormalizePath(dir+"/target/a.ts"), dir) == nil)
}

// Nested .gitignore negation: root ignores build/, a sub/.gitignore re-includes
// it. The re-included subtree must be walked and discovered; the top-level
// build/ must still be pruned. Exercises the conversion → negPrefix → prune
// chain for nested-gitignore negations end to end.
func TestDiscoverGapFiles_NestedGitignoreNegationE2E(t *testing.T) {
	files := map[string]string{
		".gitignore":        "build/\n",
		"sub/.gitignore":    "!build/\n",
		"src/a.ts":          "x",
		"build/top.ts":      "x",
		"sub/build/keep.ts": "x",
	}
	dir := setupDiscoveryContentFixture(t, files)
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}
	config = ConfigWithGitignore(config, dir, osvfs.FS(), nil)

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := DiscoverGapFiles(config, dir, spy, map[string]struct{}{}, nil, nil, false)
	gapSet := toSet(gapFiles)

	assert.Assert(t, gapSet[tspath.NormalizePath(dir+"/src/a.ts")])
	// Re-included nested build must be walked + discovered.
	assert.Assert(t, dirAccessed(spy.snapshotAccessedDirs(), "build"), "sub/build must be walked")
	if !gapSet[tspath.NormalizePath(dir+"/sub/build/keep.ts")] {
		// Consistency cross-check: linter must agree it is lintable.
		mc := config.GetConfigForFile(tspath.NormalizePath(dir+"/sub/build/keep.ts"), dir)
		t.Errorf("sub/build/keep.ts should be a gap file; GetConfigForFile=%v", mc != nil)
	}
	// Top-level build/top.ts stays ignored.
	assert.Assert(t, !gapSet[tspath.NormalizePath(dir+"/build/top.ts")], "top-level build stays ignored")
}

// --- Strongest regression: pruning must not change the gap-file set ---
//
// Oracle = { f : f matches **/*.ts ∧ f∉programFiles ∧ GetConfigForFile(f)≠nil }.
// This is exactly the linter's per-file decision; DiscoverGapFiles must equal
// it regardless of directory pruning.
func TestDiscoverGapFiles_PruningPreservesGapFiles(t *testing.T) {
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
		{"single-star file-level", []string{"build/*"}},                       // regression: 致命#1
		{"extension-filtered", []string{"target/**/*.log"}},                   // regression: ext filter must not prune
		{"dotslash negation", []string{"target/**/*", "!./target/keep/**/*"}}, // regression: 致命#2
		{"mixed", []string{"**/target/**/*", "**/tests/**", "!tests/e2e/**/*"}},
		{"bare rooted", []string{"dist"}},                            // /dist → "dist": no /**/* suffix, must not prune
		{"deep dir-only", []string{"a/b/target/**/*"}},               // deep-path positive cover
		{"brace extension filter", []string{"target/**/*.{js,jsx}"}}, // brace ext filter must not prune .ts
		{"multi-negation", []string{"**/target/**/*", "!target/keep/**/*", "!target/save/**/*"}},
		{"sequential re-ignore", []string{"target/**/*", "!target/keep/**/*", "target/keep/sub/**/*"}},
	}

	// Layout mixes shallow (depth-1) and deep (depth-2+) files plus a .tsx and a
	// non-matching .log, so single-star / extension-filter / pattern-match edges
	// are exercised by the oracle.
	layout := []string{
		"src/a.ts",
		"src/comp.tsx",
		"target/shallow.ts",
		"target/x.ts",
		"target/keep/k.ts",
		"target/other/o.ts",
		"target/log/skip.log",
		"sub/path/target/y.ts",
		"build/shallow.ts",
		"build/keep/bk.ts",
		"build/other/bo.ts",
		"dist/keep.ts",
		"dist/other.ts",
		"dist/sub/deep.ts", // bare-rooted "dist" witness: must not be pruned away
		"tests/unit/u.ts",
		"tests/e2e/e.ts",
		"a/b/target/deep/d.ts",    // deep dir-only witness (pruned)
		"a/b/other/o.ts",          // deep dir-only sibling witness (kept)
		"target/save/s.ts",        // multi-negation second re-include
		"target/keep/sub/deep.ts", // sequential re-ignore witness
	}

	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			configDir, paths := setupDiscoveryFixture(t, layout)
			config := RslintConfig{
				{Ignores: fx.ignores},
				{Files: filesPatterns, Rules: Rules{"test-rule": "error"}},
			}

			// Oracle = the linter's own per-file decision: matches a files
			// pattern AND GetConfigForFile != nil. DiscoverGapFiles must equal
			// this set regardless of directory pruning.
			var oracle []string
			for _, abs := range paths {
				if !isFileMatched(abs, filesPatterns, configDir) {
					continue
				}
				if config.GetConfigForFile(abs, configDir) != nil {
					oracle = append(oracle, abs)
				}
			}
			sort.Strings(oracle)

			got := DiscoverGapFiles(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, false)
			sort.Strings(got)

			assert.DeepEqual(t, got, oracle)
		})
	}
}

// A `!` negation inside a NON-global config entry (one carrying Files/Rules)
// must not resurrect a globally-ignored file: GetConfigForFile evaluates global
// ignores first, so entry-level ignores can only narrow. canPruneDir sees
// only the global ignores and prunes target/ — which stays consistent with the
// linter (it also excludes target/keep). Locks down the "per-entry config
// cannot cause over-prune" invariant.
func TestDiscoverGapFiles_PerEntryNegationDoesNotResurrect(t *testing.T) {
	configDir, paths := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"target/x.ts",
		"target/keep/k.ts",
	})
	config := RslintConfig{
		{Ignores: []string{"**/target/**/*"}},
		{Files: []string{"**/*.ts"}, Ignores: []string{"!target/keep/**/*"}, Rules: Rules{"test-rule": "error"}},
	}

	spy := &spyFS{FS: osvfs.FS()}
	gapFiles := DiscoverGapFiles(config, configDir, spy, map[string]struct{}{}, nil, nil, false)
	gapSet := toSet(gapFiles)

	assert.Assert(t, gapSet[paths["src/a.ts"]])
	assert.Assert(t, !gapSet[paths["target/x.ts"]])
	assert.Assert(t, !gapSet[paths["target/keep/k.ts"]], "per-entry ! must not resurrect a global ignore")
	// Linter authority agrees, and target/ is pruned (consistent).
	assert.Assert(t, config.GetConfigForFile(paths["target/keep/k.ts"], configDir) == nil)
	assert.Assert(t, !dirAccessed(spy.snapshotAccessedDirs(), "target"), "target pruned; per-entry ! does not protect")
}

// Pruning must produce identical gap files in parallel and single-threaded mode.
func TestDiscoverGapFiles_PruneSingleThreadedEquivalence(t *testing.T) {
	configDir, _ := setupDiscoveryFixture(t, []string{
		"src/a.ts",
		"target/x.ts",
		"target/keep/k.ts",
		"sub/path/target/y.ts",
	})
	config := RslintConfig{
		{Ignores: []string{"**/target/**/*", "!sub/path/target/**/*", "!target/keep/**/*"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"test-rule": "error"}},
	}

	par := DiscoverGapFiles(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, false)
	seq := DiscoverGapFiles(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, true)
	sort.Strings(par)
	sort.Strings(seq)
	assert.DeepEqual(t, par, seq)
}
