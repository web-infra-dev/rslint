package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

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
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS())
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
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS())
	assert.Assert(t, globs == nil, "should return nil when no .gitignore")
}

func TestReadGitignoreAsGlobs_Nested(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":              "dist/\n",
		"packages/app/.gitignore": "tmp/\n",
		"src/a.ts":                "x",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS())

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
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS())

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
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS())

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
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS())

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
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS())

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
	globs := ReadGitignoreAsGlobs(childDir, osvfs.FS())

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
	globs := ReadGitignoreAsGlobs(childDir, osvfs.FS())

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
	appGlobs := ReadGitignoreAsGlobs(appDir, osvfs.FS())

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
	globs := ReadGitignoreAsGlobs(deepDir, osvfs.FS())

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
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS())

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
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS())

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
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS())
	assert.Assert(t, globs == nil, "empty .gitignore → no globs")
}

func TestReadGitignoreAsGlobs_OnlyComments(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore": "# just a comment\n# another\n",
		"src/a.ts":   "x",
	})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS())
	assert.Assert(t, globs == nil, "comments-only .gitignore → no globs")
}
