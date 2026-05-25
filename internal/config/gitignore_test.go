package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

// gitignoreSpyFS wraps a real VFS and tracks which directories GetAccessibleEntries was called on.
type gitignoreSpyFS struct {
	vfs.FS
	accessedDirs []string
}

func (s *gitignoreSpyFS) GetAccessibleEntries(path string) vfs.Entries {
	s.accessedDirs = append(s.accessedDirs, path)
	return s.FS.GetAccessibleEntries(path)
}

func setupGitignoreFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()
	for name, content := range files {
		fp := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(fp), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	return tspath.NormalizePath(tmpDir)
}

// --- Pattern conversion tests ---

func TestConvertGitignoreToGlobs_DirectoryPattern(t *testing.T) {
	globs := convertGitignoreToGlobs("dist/\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "**/dist/**/*")
}

func TestConvertGitignoreToGlobs_RootAnchored(t *testing.T) {
	globs := convertGitignoreToGlobs("/dist\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "dist")
}

func TestConvertGitignoreToGlobs_RootAnchoredDir(t *testing.T) {
	globs := convertGitignoreToGlobs("/dist/\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "dist/**/*")
}

func TestConvertGitignoreToGlobs_WildcardPattern(t *testing.T) {
	globs := convertGitignoreToGlobs("*.log\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "**/*.log")
}

func TestConvertGitignoreToGlobs_DoublestarPattern(t *testing.T) {
	// **/*.test.ts contains / → implicitly rooted, no extra **/ prefix
	globs := convertGitignoreToGlobs("**/*.test.ts\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "**/*.test.ts")
}

func TestConvertGitignoreToGlobs_PathWithSlash(t *testing.T) {
	// Contains / → implicitly rooted
	globs := convertGitignoreToGlobs("src/dist\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "src/dist")
}

func TestConvertGitignoreToGlobs_Negation(t *testing.T) {
	globs := convertGitignoreToGlobs("!dist/\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "!**/dist/**/*")
}

func TestConvertGitignoreToGlobs_CommentsAndBlanks(t *testing.T) {
	globs := convertGitignoreToGlobs("# comment\n\ndist/\n\n# another\n*.log\n", "")
	assert.Equal(t, len(globs), 2)
	assert.Equal(t, globs[0], "**/dist/**/*")
	assert.Equal(t, globs[1], "**/*.log")
}

func TestConvertGitignoreToGlobs_TrailingWhitespace(t *testing.T) {
	globs := convertGitignoreToGlobs("  dist/  \n  coverage/\n", "")
	assert.Equal(t, len(globs), 2)
	assert.Equal(t, globs[0], "**/dist/**/*")
	assert.Equal(t, globs[1], "**/coverage/**/*")
}

// --- Nested .gitignore with baseDir prefix ---

func TestConvertGitignoreToGlobs_NestedBaseDir(t *testing.T) {
	globs := convertGitignoreToGlobs("tmp/\n", "packages/app")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "packages/app/**/tmp/**/*")
}

func TestConvertGitignoreToGlobs_NestedNegation(t *testing.T) {
	globs := convertGitignoreToGlobs("!dist/\n", "packages/app")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "!packages/app/**/dist/**/*")
}

func TestConvertGitignoreToGlobs_NestedRootAnchored(t *testing.T) {
	// /src in packages/app/.gitignore → packages/app/src
	globs := convertGitignoreToGlobs("/src/\n", "packages/app")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "packages/app/src/**/*")
}

func TestConvertGitignoreToGlobs_NestedWildcard(t *testing.T) {
	globs := convertGitignoreToGlobs("*.generated.ts\n", "packages/app")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "packages/app/**/*.generated.ts")
}

// --- ReadGitignoreAsGlobs integration tests ---

func TestReadGitignoreAsGlobs_RootOnly(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore": "dist/\n*.log\n",
		"src/a.ts":   "x",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)
	assert.Assert(t, len(globs) >= 2)

	hasDistGlob := false
	hasLogGlob := false
	for _, g := range globs {
		if g == "**/dist/**/*" {
			hasDistGlob = true
		}
		if g == "**/*.log" {
			hasLogGlob = true
		}
	}
	assert.Assert(t, hasDistGlob, "should have dist glob")
	assert.Assert(t, hasLogGlob, "should have *.log glob")
}

func TestReadGitignoreAsGlobs_NoGitignore(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		"src/a.ts": "x",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)
	assert.Assert(t, globs == nil, "should return nil when no .gitignore")
}

func TestReadGitignoreAsGlobs_Nested(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":              "dist/\n",
		"packages/app/.gitignore": "tmp/\n",
		"src/a.ts":                "x",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)

	hasDistGlob := false
	hasTmpGlob := false
	for _, g := range globs {
		if g == "**/dist/**/*" {
			hasDistGlob = true
		}
		if g == "packages/app/**/tmp/**/*" {
			hasTmpGlob = true
		}
	}
	assert.Assert(t, hasDistGlob, "should have root dist glob")
	assert.Assert(t, hasTmpGlob, "should have nested tmp glob with packages/app prefix")
}

func TestReadGitignoreAsGlobs_NestedNegation(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":              "dist/\n",
		"packages/app/.gitignore": "!dist/\n",
		"src/a.ts":                "x",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)

	hasDistGlob := false
	hasNegation := false
	for _, g := range globs {
		if g == "**/dist/**/*" {
			hasDistGlob = true
		}
		if g == "!packages/app/**/dist/**/*" {
			hasNegation = true
		}
	}
	assert.Assert(t, hasDistGlob, "should have root dist glob")
	assert.Assert(t, hasNegation, "should have child negation glob with prefix")
}

func TestReadGitignoreAsGlobs_PrunesGitignoredDirs(t *testing.T) {
	// Root .gitignore ignores target/. Scanner should NOT enter target/
	// even to look for nested .gitignore files.
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                    "target/\n",
		"src/a.ts":                      "x",
		"target/deep/nested/.gitignore": "# should not be read\nfoo\n",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)

	// Positive: root .gitignore should be read
	assert.Assert(t, len(globs) >= 1, "should have at least root target/ glob")

	// Negative: target/deep/nested/.gitignore should NOT be read
	for _, g := range globs {
		if strings.Contains(g, "target/deep") {
			t.Errorf("should not read .gitignore inside gitignored target/ dir, but found: %s", g)
		}
	}
}

func TestReadGitignoreAsGlobs_SkipsNodeModules(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                  "dist/\n",
		"node_modules/pkg/.gitignore": "# should not be read\n",
		"src/a.ts":                    "x",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)

	// Only root .gitignore should be read
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "**/dist/**/*")
}

// --- Scope isolation tests ---

func TestConvertGitignoreToGlobs_ScopeIsolation(t *testing.T) {
	// packages/app/.gitignore has "src/" → should only affect packages/app/
	globs := convertGitignoreToGlobs("src/\n", "packages/app")
	assert.Equal(t, len(globs), 1)
	// Must be prefixed with packages/app, NOT a bare **/src/**/*
	assert.Equal(t, globs[0], "packages/app/**/src/**/*")
}

func TestReadGitignoreAsGlobs_ThreeLevelNested(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                  "dist/\n",
		"packages/app/.gitignore":     "tmp/\n",
		"packages/app/sub/.gitignore": "cache/\n",
		"src/a.ts":                    "x",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)

	hasRoot := false
	hasApp := false
	hasSub := false
	for _, g := range globs {
		if g == "**/dist/**/*" {
			hasRoot = true
		}
		if g == "packages/app/**/tmp/**/*" {
			hasApp = true
		}
		if g == "packages/app/sub/**/cache/**/*" {
			hasSub = true
		}
	}
	assert.Assert(t, hasRoot, "root dist")
	assert.Assert(t, hasApp, "app tmp")
	assert.Assert(t, hasSub, "sub cache")
}

// --- Additional edge case tests ---

func TestConvertGitignoreToGlobs_WindowsLineEndings(t *testing.T) {
	globs := convertGitignoreToGlobs("dist/\r\ncoverage/\r\n", "")
	assert.Equal(t, len(globs), 2)
	assert.Equal(t, globs[0], "**/dist/**/*")
	assert.Equal(t, globs[1], "**/coverage/**/*")
}

func TestConvertGitignoreToGlobs_MultiDotExtension(t *testing.T) {
	globs := convertGitignoreToGlobs("*.d.ts\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "**/*.d.ts")
}

func TestConvertGitignoreToGlobs_StarMatchAll(t *testing.T) {
	// Bare * matches everything at any depth
	globs := convertGitignoreToGlobs("*\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "**/*")
}

func TestConvertGitignoreToGlobs_BareNameNoSlash(t *testing.T) {
	// "build" without trailing / — in gitignore matches both file and dir
	// Our conversion: unrooted, no trailing / → **/build (no /**/* suffix)
	globs := convertGitignoreToGlobs("build\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "**/build")
}

func TestConvertGitignoreToGlobs_NestedStarMatchAll(t *testing.T) {
	globs := convertGitignoreToGlobs("*\n", "packages/app")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "packages/app/**/*")
}

// --- Ancestor inheritance tests ---

func TestReadGitignoreAsGlobs_AncestorInheritance(t *testing.T) {
	// Root has .gitignore with dist/. configDir is packages/app (child).
	// ReadGitignoreAsGlobs should walk UP to find root .gitignore.
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":            "dist/\n",
		"packages/app/src/a.ts": "x",
	})
	childDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(childDir, osvfs.FS(), nil)

	hasDistGlob := false
	for _, g := range globs {
		if g == "**/dist/**/*" {
			hasDistGlob = true
		}
	}
	assert.Assert(t, hasDistGlob, "child configDir should inherit root .gitignore dist/ pattern")
}

func TestReadGitignoreAsGlobs_AncestorPlusOwn(t *testing.T) {
	// Root has dist/, child has tmp/. Both should be in globs.
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":              "dist/\n",
		"packages/app/.gitignore": "tmp/\n",
		"packages/app/src/a.ts":   "x",
	})
	childDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(childDir, osvfs.FS(), nil)

	hasDist := false
	hasTmp := false
	for _, g := range globs {
		if g == "**/dist/**/*" {
			hasDist = true
		}
		if g == "**/tmp/**/*" {
			hasTmp = true
		}
	}
	assert.Assert(t, hasDist, "should inherit root dist/")
	assert.Assert(t, hasTmp, "should have own tmp/")
}

func TestReadGitignoreAsGlobs_SiblingIsolation(t *testing.T) {
	// packages/app has tmp/, packages/lib has cache/.
	// ReadGitignoreAsGlobs for app should NOT include lib's cache/.
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":              "dist/\n",
		"packages/app/.gitignore": "tmp/\n",
		"packages/lib/.gitignore": "cache/\n",
		"packages/app/src/a.ts":   "x",
		"packages/lib/src/b.ts":   "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	appGlobs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	for _, g := range appGlobs {
		if strings.Contains(g, "cache") {
			t.Errorf("app globs should NOT contain lib's cache pattern, found: %s", g)
		}
	}
	// Should have dist (inherited) and tmp (own)
	hasDist := false
	hasTmp := false
	for _, g := range appGlobs {
		if g == "**/dist/**/*" {
			hasDist = true
		}
		if g == "**/tmp/**/*" {
			hasTmp = true
		}
	}
	assert.Assert(t, hasDist, "should inherit root dist/")
	assert.Assert(t, hasTmp, "should have own tmp/")
}

func TestReadGitignoreAsGlobs_DeepAncestor(t *testing.T) {
	// .gitignore at root, configDir at packages/app/sub (3 levels deep)
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                "dist/\n",
		"packages/app/sub/src/a.ts": "x",
	})
	deepDir := tspath.NormalizePath(filepath.Join(dir, "packages/app/sub"))
	globs := ReadGitignoreAsGlobs(deepDir, osvfs.FS(), nil)

	hasDist := false
	for _, g := range globs {
		if g == "**/dist/**/*" {
			hasDist = true
		}
	}
	assert.Assert(t, hasDist, "deeply nested configDir should still inherit root .gitignore")
}

func TestReadGitignoreAsGlobs_MultipleAncestors(t *testing.T) {
	// Root has .gitignore (dist/), packages/ has .gitignore (vendor/),
	// configDir is packages/app/. Both ancestors should be collected.
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":            "dist/\n",
		"packages/.gitignore":   "vendor/\n",
		"packages/app/src/a.ts": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	hasDist := false
	hasVendor := false
	for _, g := range globs {
		if g == "**/dist/**/*" {
			hasDist = true
		}
		if g == "**/vendor/**/*" {
			hasVendor = true
		}
	}
	assert.Assert(t, hasDist, "should inherit root dist/")
	assert.Assert(t, hasVendor, "should inherit intermediate packages/ vendor/")
}

func TestReadGitignoreAsGlobs_AncestorNegationOverride(t *testing.T) {
	// Root ignores dist/, intermediate packages/.gitignore re-includes with !dist/.
	// configDir at packages/app should see both patterns (sequential evaluation).
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":            "dist/\n",
		"packages/.gitignore":   "!dist/\n",
		"packages/app/src/a.ts": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	hasDist := false
	hasNegation := false
	for _, g := range globs {
		if g == "**/dist/**/*" {
			hasDist = true
		}
		if g == "!**/dist/**/*" {
			hasNegation = true
		}
	}
	assert.Assert(t, hasDist, "should have root dist/")
	assert.Assert(t, hasNegation, "should have intermediate !dist/ negation")
}

func TestReadGitignoreAsGlobs_EmptyFile(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore": "",
		"src/a.ts":   "x",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)
	assert.Assert(t, globs == nil, "empty .gitignore → no globs")
}

func TestReadGitignoreAsGlobs_OnlyComments(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore": "# just a comment\n# another\n",
		"src/a.ts":   "x",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)
	assert.Assert(t, globs == nil, "comments-only .gitignore → no globs")
}

// =============================================================================
// Cross-matrix tests: configIgnores × .gitignore interaction
//
// These tests verify that passing configIgnores to ReadGitignoreAsGlobs
// correctly prunes config-ignored directories while preserving all .gitignore
// patterns from non-ignored directories.
// =============================================================================

// Matrix case: config ignores dir-level (**/tests/**) + nested .gitignore inside tests/
// Expected: tests/.gitignore NOT collected (pruned), root .gitignore collected.
// --- Cross-matrix: configIgnores × .gitignore ---
//
// Helper: globSet converts a glob slice to a set for precise assertions.
func globSet(globs []string) map[string]bool {
	m := make(map[string]bool, len(globs))
	for _, g := range globs {
		m[g] = true
	}
	return m
}

// A1: dir-level config ignore (**/tests/**) prunes tests/.gitignore.
func TestReadGitignoreAsGlobs_ConfigIgnoresPrunesDir(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":            "dist/\n",
		"tests/.gitignore":      "snapshots/\n",
		"tests/unit/.gitignore": "coverage/\n",
		"tests/unit/a.test.ts":  "x",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), []string{"**/tests/**"})
	gs := globSet(globs)

	// Exactly 1 glob: root .gitignore "dist/" → "**/dist/**/*"
	assert.Equal(t, len(globs), 1, "should have exactly 1 glob (root only), got: %v", globs)
	assert.Assert(t, gs["**/dist/**/*"], "should contain root dist pattern")
}

// A2: file-level config ignore (**/tests/**/*) does NOT prune tests/.gitignore.
func TestReadGitignoreAsGlobs_FileLevelConfigIgnoreDoesNotPrune(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":       "dist/\n",
		"tests/.gitignore": "snapshots/\n",
		"tests/a.test.ts":  "x",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), []string{"**/tests/**/*"})
	gs := globSet(globs)

	// 2 globs: root dist + tests snapshots (file-level ignore doesn't prune dirs)
	assert.Equal(t, len(globs), 2, "should have 2 globs, got: %v", globs)
	assert.Assert(t, gs["**/dist/**/*"], "root dist")
	assert.Assert(t, gs["tests/**/snapshots/**/*"], "tests snapshots should be collected")
}

// A3: dir-level + negation — isDirPathBlocked skips negation → still prunes.
func TestReadGitignoreAsGlobs_NegationInConfigIgnoreStillPrunes(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":           "dist/\n",
		"tests/e2e/.gitignore": "screenshots/\n",
		"tests/e2e/a.ts":       "x",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), []string{"**/tests/**", "!tests/e2e/**"})

	// Exactly 1 glob: root dist. tests/e2e/.gitignore NOT collected.
	assert.Equal(t, len(globs), 1, "should have exactly 1 glob, got: %v", globs)
	assert.Equal(t, globs[0], "**/dist/**/*")
}

// A4: sibling dirs — config ignores tests/, packages/ .gitignore preserved.
func TestReadGitignoreAsGlobs_SiblingDirPreservesGitignore(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":              "dist/\n",
		"packages/foo/.gitignore": "tmp/\n",
		"tests/.gitignore":        "snapshots/\n",
		"packages/foo/a.ts":       "x",
		"tests/a.test.ts":         "x",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), []string{"**/tests/**"})
	gs := globSet(globs)

	// Exactly 2 globs: root dist + packages/foo tmp. tests/ pruned.
	assert.Equal(t, len(globs), 2, "should have 2 globs, got: %v", globs)
	assert.Assert(t, gs["**/dist/**/*"], "root dist")
	assert.Assert(t, gs["packages/foo/**/tmp/**/*"], "packages/foo tmp")
}

// A5: nil configIgnores — backward compat, all .gitignore collected.
func TestReadGitignoreAsGlobs_NilConfigIgnores(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":       "dist/\n",
		"tests/.gitignore": "snapshots/\n",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)
	gs := globSet(globs)

	assert.Equal(t, len(globs), 2, "should have 2 globs (both collected), got: %v", globs)
	assert.Assert(t, gs["**/dist/**/*"], "root dist")
	assert.Assert(t, gs["tests/**/snapshots/**/*"], "tests snapshots")
}

// A6: deeply nested config-ignored dir — all nested .gitignore skipped.
func TestReadGitignoreAsGlobs_DeepNestedConfigIgnore(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                   "*.log\n",
		"vendor/.gitignore":            "cache/\n",
		"vendor/pkg/.gitignore":        "build/\n",
		"vendor/pkg/sub/.gitignore":    "temp/\n",
		"src/.gitignore":               "generated/\n",
		"vendor/pkg/sub/deep/file.txt": "x",
		"src/a.ts":                     "x",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), []string{"**/vendor/**"})
	gs := globSet(globs)

	// Exactly 2 globs: root *.log + src generated. All 3 vendor .gitignore pruned.
	assert.Equal(t, len(globs), 2, "should have 2 globs, got: %v", globs)
	assert.Assert(t, gs["**/*.log"], "root *.log")
	assert.Assert(t, gs["src/**/generated/**/*"], "src generated")
}

// A7: wildcard config ignore (crates/**) — multiple nested .gitignore skipped.
func TestReadGitignoreAsGlobs_WildcardConfigIgnoreSkipsNested(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                    "dist/\n",
		"crates/core/.gitignore":        "artifacts/\n",
		"crates/binding/.gitignore":     "generated/\n",
		"packages/rspack/.gitignore":    "tmp/\n",
		"crates/core/src/lib.rs":        "x",
		"crates/binding/src/lib.rs":     "x",
		"packages/rspack/src/index.ts":  "x",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), []string{"crates/**"})
	gs := globSet(globs)

	// Exactly 2: root dist + packages/rspack tmp. Both crates/ .gitignore pruned.
	assert.Equal(t, len(globs), 2, "should have 2 globs, got: %v", globs)
	assert.Assert(t, gs["**/dist/**/*"], "root dist")
	assert.Assert(t, gs["packages/rspack/**/tmp/**/*"], "packages/rspack tmp")
}

// A8: empty configIgnores slice (non-nil) — same as nil, no pruning.
func TestReadGitignoreAsGlobs_EmptyConfigIgnores(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":       "dist/\n",
		"tests/.gitignore": "snapshots/\n",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), []string{})
	gs := globSet(globs)

	assert.Equal(t, len(globs), 2, "empty slice = no pruning, got: %v", globs)
	assert.Assert(t, gs["**/dist/**/*"], "root dist")
	assert.Assert(t, gs["tests/**/snapshots/**/*"], "tests snapshots")
}

// A9: exact path config ignore — only that specific directory blocked.
func TestReadGitignoreAsGlobs_ExactPathConfigIgnore(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":              "dist/\n",
		"build/output/.gitignore": "cache/\n",
		"build/tools/.gitignore":  "tmp/\n",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), []string{"build/output/**"})
	gs := globSet(globs)

	// 2 globs: root dist + build/tools tmp. build/output cache pruned.
	assert.Equal(t, len(globs), 2, "should have 2 globs, got: %v", globs)
	assert.Assert(t, gs["**/dist/**/*"], "root dist")
	assert.Assert(t, gs["build/tools/**/tmp/**/*"], "build/tools tmp preserved")
}

// A10: ExtractConfigIgnores only extracts from global-ignore entries.
// An entry is "global ignore" when it has ONLY ignores — no files, rules, plugins, etc.
// Note: {Files: []string{}} (empty slice) still counts as len(Files)==0, so IS global.
func TestExtractConfigIgnores_OnlyGlobalEntries(t *testing.T) {
	config := RslintConfig{
		{Ignores: []string{"**/tests/**"}},                                             // global ignore ✓
		{Files: []string{"**/*.ts"}, Rules: Rules{"r": "error"}},                       // has files+rules → NOT global
		{Ignores: []string{"scripts/**"}},                                              // global ignore ✓
		{Files: []string{}, Ignores: []string{"also-global"}},                          // empty files + only ignores → IS global (len(Files)==0)
		{Ignores: []string{"vendor/**"}, Rules: Rules{"r": "error"}},                   // has rules → NOT global
		{Files: []string{"**/*.js"}, Ignores: []string{"not-global"}},                  // has non-empty files → NOT global
	}

	ignores := ExtractConfigIgnores(config)

	// Entries 0, 2, 3 are global ignores. Entries 1, 4, 5 are not.
	assert.Equal(t, len(ignores), 3, "should extract exactly 3 patterns, got: %v", ignores)
	assert.Equal(t, ignores[0], "**/tests/**")
	assert.Equal(t, ignores[1], "scripts/**")
	assert.Equal(t, ignores[2], "also-global")
}

// A11: multiple global ignore entries — all patterns combined.
func TestExtractConfigIgnores_MultipleEntries(t *testing.T) {
	config := RslintConfig{
		{Ignores: []string{"**/tests/**", "packages/rspack/compiled/**"}},
		{Ignores: []string{"crates/**"}},
	}

	ignores := ExtractConfigIgnores(config)
	assert.Equal(t, len(ignores), 3, "should combine all patterns, got: %v", ignores)
	assert.Equal(t, ignores[0], "**/tests/**")
	assert.Equal(t, ignores[1], "packages/rspack/compiled/**")
	assert.Equal(t, ignores[2], "crates/**")
}

// A12: bare config ignore without /** suffix (e.g., "tests" or "dist").
// isDirPathBlocked matches the directory itself and all parents, so this
// should still prune. Verify via exact glob count.
func TestReadGitignoreAsGlobs_BareConfigIgnorePattern(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":       "*.log\n",
		"tests/.gitignore": "snapshots/\n",
		"src/.gitignore":   "generated/\n",
		"tests/a.test.ts":  "x",
		"src/a.ts":         "x",
	})

	// Bare "tests" — no trailing /**. isDirPathBlocked("tests", ["tests"]) should
	// match via matchGlob("tests", "tests") → true → blocked.
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), []string{"tests"})
	gs := globSet(globs)

	// 2 globs: root *.log + src/generated. tests/ pruned by bare "tests" pattern.
	assert.Equal(t, len(globs), 2, "bare 'tests' should prune tests/.gitignore, got: %v", globs)
	assert.Assert(t, gs["**/*.log"], "root *.log")
	assert.Assert(t, gs["src/**/generated/**/*"], "src generated")
}

// A13: gitignore and config ignore both target the same directory.
// collectGitignoreGlobs checks gitignore patterns FIRST (isDirIgnoredByGlobs),
// then config ignores (isDirPathBlocked). If both match, the directory is pruned
// by whichever check comes first. Verify the result is identical to having only one.
func TestReadGitignoreAsGlobs_OverlappingGitignoreAndConfigIgnore(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":       "dist/\nbuild/\n",
		"dist/.gitignore":  "cache/\n",
		"build/.gitignore": "tmp/\n",
		"src/.gitignore":   "generated/\n",
		"dist/bundle.js":   "x",
		"build/output.js":  "x",
		"src/a.ts":         "x",
	})

	// dist/ is in both .gitignore and config ignores.
	// build/ is only in .gitignore.
	configIgnores := []string{"**/dist/**"}
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), configIgnores)
	gs := globSet(globs)

	// dist/ pruned by GITIGNORE (isDirIgnoredByGlobs runs first → **/dist/**/* matches).
	// build/ pruned by GITIGNORE.
	// Neither dist/.gitignore nor build/.gitignore collected.
	// Only root .gitignore (dist/ + build/) and src/.gitignore (generated/) remain.
	assert.Equal(t, len(globs), 3, "should have 3 globs, got: %v", globs)
	assert.Assert(t, gs["**/dist/**/*"], "root dist")
	assert.Assert(t, gs["**/build/**/*"], "root build")
	assert.Assert(t, gs["src/**/generated/**/*"], "src generated")

	// Compare: without config ignores, result should be identical (gitignore already prunes both).
	globsNoConfig := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)
	assert.Equal(t, len(globsNoConfig), len(globs), "overlapping config ignore should not change result vs nil configIgnores")
}

// Regression: same fixture, with vs without configIgnores.
// Proves the EXACT effect of configIgnores — the difference must be only patterns
// from the config-ignored directory (and nothing else changes).
func TestReadGitignoreAsGlobs_RegressionWithVsWithout(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                "dist/\n",
		"src/.gitignore":            "generated/\n",
		"tests/.gitignore":          "snapshots/\n",
		"tests/unit/.gitignore":     "coverage/\n",
		"tests/unit/a.test.ts":      "x",
		"src/a.ts":                  "x",
	})

	// WITHOUT configIgnores: all 4 .gitignore files collected → 4 patterns.
	globsWithout := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)
	gsWithout := globSet(globsWithout)
	assert.Equal(t, len(globsWithout), 4, "without configIgnores should collect all 4 patterns, got: %v", globsWithout)
	assert.Assert(t, gsWithout["**/dist/**/*"])
	assert.Assert(t, gsWithout["src/**/generated/**/*"])
	assert.Assert(t, gsWithout["tests/**/snapshots/**/*"])
	assert.Assert(t, gsWithout["tests/unit/**/coverage/**/*"])

	// WITH configIgnores: tests/ pruned → only 2 patterns (root + src).
	globsWith := ReadGitignoreAsGlobs(dir, osvfs.FS(), []string{"**/tests/**"})
	gsWith := globSet(globsWith)
	assert.Equal(t, len(globsWith), 2, "with configIgnores should collect 2 patterns (tests/ pruned), got: %v", globsWith)
	assert.Assert(t, gsWith["**/dist/**/*"])
	assert.Assert(t, gsWith["src/**/generated/**/*"])

	// The EXACT difference: the 2 missing patterns are from tests/ subtree.
	assert.Assert(t, !gsWith["tests/**/snapshots/**/*"], "tests/ pattern should be pruned")
	assert.Assert(t, !gsWith["tests/unit/**/coverage/**/*"], "tests/unit/ pattern should be pruned")
}

// Spy test: verify collectGitignoreGlobs does NOT call GetAccessibleEntries
// on config-ignored directories. This proves the directory is truly skipped
// at the walk level (not entered then filtered), which is the performance win.
func TestReadGitignoreAsGlobs_ConfigIgnoreSkipsWalkLevel(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                    "dist/\n",
		"src/a.ts":                      "x",
		"src/.gitignore":                "generated/\n",
		"tests/unit/a.test.ts":          "x",
		"tests/.gitignore":              "snapshots/\n",
		"tests/unit/.gitignore":         "coverage/\n",
		"tests/unit/deep/nested/a.ts":   "x",
	})

	spy := &gitignoreSpyFS{FS: osvfs.FS()}
	globs := ReadGitignoreAsGlobs(dir, spy, []string{"**/tests/**"})

	// Output correctness: only root + src patterns.
	assert.Equal(t, len(globs), 2, "got: %v", globs)

	// Walk correctness: tests/ and its children should NOT be accessed.
	for _, accessed := range spy.accessedDirs {
		rel := strings.TrimPrefix(accessed, dir)
		if strings.Contains(rel, "tests") {
			t.Errorf("collectGitignoreGlobs entered config-ignored tests/ directory: GetAccessibleEntries(%s)", accessed)
		}
	}

	// Positive check: root and src/ SHOULD be accessed.
	rootAccessed := false
	srcAccessed := false
	for _, accessed := range spy.accessedDirs {
		if accessed == dir {
			rootAccessed = true
		}
		if strings.HasSuffix(accessed, "/src") {
			srcAccessed = true
		}
	}
	assert.Assert(t, rootAccessed, "root directory should be accessed")
	assert.Assert(t, srcAccessed, "src/ should be accessed (not config-ignored)")
}

// Spy comparison: same fixture WITH vs WITHOUT configIgnores.
// Proves that configIgnores reduces the number of GetAccessibleEntries calls.
func TestReadGitignoreAsGlobs_ConfigIgnoreReducesWalkCount(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                  "*.log\n",
		"tests/unit/a.ts":             "x",
		"tests/unit/sub/b.ts":         "x",
		"tests/.gitignore":            "tmp/\n",
		"tests/unit/.gitignore":       "output/\n",
		"tests/unit/sub/.gitignore":   "cache/\n",
		"src/a.ts":                    "x",
	})

	// Without configIgnores: walks everything including tests/ subtree.
	spyWithout := &gitignoreSpyFS{FS: osvfs.FS()}
	ReadGitignoreAsGlobs(dir, spyWithout, nil)
	countWithout := len(spyWithout.accessedDirs)

	// With configIgnores: skips tests/ subtree.
	spyWith := &gitignoreSpyFS{FS: osvfs.FS()}
	ReadGitignoreAsGlobs(dir, spyWith, []string{"**/tests/**"})
	countWith := len(spyWith.accessedDirs)

	// With configIgnores should access FEWER directories.
	assert.Assert(t, countWith < countWithout,
		"configIgnores should reduce walk count: with=%d, without=%d", countWith, countWithout)

	// Specifically: tests/, tests/unit/, tests/unit/sub/ = 3 dirs skipped.
	assert.Equal(t, countWithout-countWith, 3,
		"should skip exactly 3 dirs (tests/, tests/unit/, tests/unit/sub/), but diff is %d", countWithout-countWith)
}
