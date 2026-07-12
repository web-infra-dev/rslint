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

type gitignoreMockFS struct {
	vfs.FS
	entries         map[string]vfs.Entries
	files           map[string]string
	resolvedPaths   map[string]string
	readFileCalls   []string
	accessedDirs    []string
	realpathCalls   []string
	caseSensitiveFS bool
}

func (m *gitignoreMockFS) ReadFile(path string) (string, bool) {
	m.readFileCalls = append(m.readFileCalls, path)
	content, ok := m.files[path]
	return content, ok
}

func (m *gitignoreMockFS) GetAccessibleEntries(path string) vfs.Entries {
	m.accessedDirs = append(m.accessedDirs, path)
	return m.entries[path]
}

func (m *gitignoreMockFS) Realpath(path string) string {
	m.realpathCalls = append(m.realpathCalls, path)
	if realpath, ok := m.resolvedPaths[path]; ok {
		return realpath
	}
	return path
}

func (m *gitignoreMockFS) UseCaseSensitiveFileNames() bool {
	return m.caseSensitiveFS
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

func TestConfigWithGitignore_DefaultAndExplicitScopes(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":          "root.ts\n",
		"root.ts":             "debugger;\n",
		"nested/.gitignore":   "private.ts\n",
		"nested/private.ts":   "debugger;\n",
		"nested/unrelated.ts": "debugger;\n",
	})
	base := RslintConfig{{Rules: Rules{"no-debugger": "error"}}}

	full := ConfigWithGitignore(base, dir, osvfs.FS(), nil)
	assert.Assert(t, full.IsFileIgnored(tspath.ResolvePath(dir, "root.ts"), dir))
	assert.Assert(t, full.IsFileIgnored(tspath.ResolvePath(dir, "nested/private.ts"), dir))

	explicit := ConfigWithGitignore(
		base,
		dir,
		osvfs.FS(),
		[]string{tspath.ResolvePath(dir, "root.ts")},
	)
	assert.Assert(t, explicit.IsFileIgnored(tspath.ResolvePath(dir, "root.ts"), dir))
	assert.Assert(t, !explicit.IsFileIgnored(tspath.ResolvePath(dir, "nested/private.ts"), dir))

	empty := ConfigWithGitignore(base, dir, osvfs.FS(), []string{})
	assert.Equal(t, len(empty), len(base))
	assert.Assert(t, base[0].Ignores == nil, "ConfigWithGitignore mutated its input: %v", base)
}

func TestConfigWithGitignore_ExplicitMatchesDefaultPruning(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                   "dist/\n!dist/types/\n",
		"dist/types/.gitignore":        "private.ts\n",
		"dist/types/private.ts":        "debugger;\n",
		"dist/types/unrelated-file.ts": "debugger;\n",
	})
	target := tspath.ResolvePath(dir, "dist/types/private.ts")
	base := RslintConfig{{Rules: Rules{"no-debugger": "error"}}}

	full := ConfigWithGitignore(base, dir, osvfs.FS(), nil)
	explicit := ConfigWithGitignore(base, dir, osvfs.FS(), []string{target})
	assert.Equal(t, explicit.IsFileIgnored(target, dir), full.IsFileIgnored(target, dir))
	assert.Assert(t, !explicit.IsFileIgnored(target, dir))
}

func TestConfigWithGitignore_RejectsPatternsOutsideSourceScope(t *testing.T) {
	for _, test := range []struct {
		name    string
		pattern string
	}{
		{name: "forward slash", pattern: "../target.ts\n"},
		{name: "backslash", pattern: "/..\\target.ts\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := setupGitignoreFixture(t, map[string]string{
				"target.ts":            "debugger;\n",
				"unrelated/.gitignore": test.pattern,
			})
			target := tspath.ResolvePath(dir, "target.ts")
			base := RslintConfig{{Rules: Rules{"no-debugger": "error"}}}

			for _, effective := range []RslintConfig{
				ConfigWithGitignore(base, dir, osvfs.FS(), nil),
				ConfigWithGitignore(base, dir, osvfs.FS(), []string{target}),
			} {
				assert.Assert(t, !effective.IsFileIgnored(target, dir))
			}
		})
	}
}

func TestConfigWithGitignore_SymlinkedConfigPathSpace(t *testing.T) {
	realRoot := setupGitignoreFixture(t, map[string]string{
		".gitignore":    "src/source.ts\n",
		"src/source.ts": "debugger;\n",
	})
	aliasRoot := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(realRoot, aliasRoot); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	base := RslintConfig{{Rules: Rules{"no-debugger": "error"}}}

	for _, test := range []struct {
		name      string
		configDir string
		target    string
	}{
		{name: "aliased config and physical target", configDir: aliasRoot, target: filepath.Join(realRoot, "src/source.ts")},
		{name: "physical config and aliased target", configDir: realRoot, target: filepath.Join(aliasRoot, "src/source.ts")},
	} {
		t.Run(test.name, func(t *testing.T) {
			effective := ConfigWithGitignore(base, test.configDir, osvfs.FS(), []string{test.target})
			matchFile, matchDir := ResolveConfigPathSpace(test.target, test.configDir, osvfs.FS())
			assert.Assert(t, effective.IsFileIgnored(matchFile, matchDir))
		})
	}
}

func TestConfigWithGitignore_SymlinkedConfigKeepsLexicalAncestors(t *testing.T) {
	workspace := t.TempDir()
	externalParent := t.TempDir()
	realRoot := filepath.Join(externalParent, "real")
	if err := os.MkdirAll(filepath.Join(realRoot, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(realRoot, "src/source.ts")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	aliasRoot := filepath.Join(workspace, "alias")
	if err := os.Symlink(realRoot, aliasRoot); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	base := RslintConfig{{Rules: Rules{"no-debugger": "error"}}}
	matchFile, matchDir := ResolveConfigPathSpace(target, aliasRoot, osvfs.FS())

	t.Run("lexical ancestor applies", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(workspace, ".gitignore"), []byte("/alias/src/source.ts\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(filepath.Join(workspace, ".gitignore"))

		full := ConfigWithGitignore(base, aliasRoot, osvfs.FS(), nil)
		explicit := ConfigWithGitignore(base, aliasRoot, osvfs.FS(), []string{target})
		assert.Assert(t, full.IsFileIgnored(matchFile, matchDir))
		assert.Equal(t, explicit.IsFileIgnored(matchFile, matchDir), full.IsFileIgnored(matchFile, matchDir))
	})

	t.Run("physical-only ancestor does not apply", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(externalParent, ".gitignore"), []byte("/real/src/source.ts\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		full := ConfigWithGitignore(base, aliasRoot, osvfs.FS(), nil)
		explicit := ConfigWithGitignore(base, aliasRoot, osvfs.FS(), []string{target})
		assert.Assert(t, !full.IsFileIgnored(matchFile, matchDir))
		assert.Equal(t, explicit.IsFileIgnored(matchFile, matchDir), full.IsFileIgnored(matchFile, matchDir))
	})
}

func TestConfigWithGitignore_ExplicitSkipsDescendantSymlinkSource(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{})
	external := setupGitignoreFixture(t, map[string]string{
		".gitignore": "source.ts\n",
		"source.ts":  "debugger;\n",
	})
	if err := os.Symlink(external, filepath.Join(dir, "link")); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	target := filepath.Join(dir, "link/source.ts")
	matchFile, matchDir := ResolveConfigPathSpace(target, dir, osvfs.FS())
	base := RslintConfig{{Rules: Rules{"no-debugger": "error"}}}

	full := ConfigWithGitignore(base, dir, osvfs.FS(), nil)
	explicit := ConfigWithGitignore(base, dir, osvfs.FS(), []string{target})
	assert.Equal(t, explicit.IsFileIgnored(matchFile, matchDir), full.IsFileIgnored(matchFile, matchDir))
	assert.Assert(t, !explicit.IsFileIgnored(matchFile, matchDir))
}

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

func TestReadGitignoreAsGlobs_AncestorAnchoredOutsideConfigDoesNotApply(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                 "/dist/\n",
		"dist/root-build.js":         "x",
		"packages/app/dist/app.js":   "x",
		"packages/app/src/index.js":  "x",
		"packages/app/rslint.jsonc":  "[]",
		"packages/other/dist/app.js": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	for _, g := range globs {
		if strings.Contains(g, "dist") {
			t.Fatalf("root-anchored /dist/ must not become a config-relative dist ignore for nested configs, got %v", globs)
		}
	}
	cfg := RslintConfig{
		{Ignores: globs},
		{Rules: Rules{"no-debugger": "error"}},
	}
	appDist := tspath.NormalizePath(filepath.Join(appDir, "dist/app.js"))
	if cfg.GetConfigForFile(appDist, appDir) == nil {
		t.Fatalf("ancestor /dist/ should not ignore nested config dist file; globs=%v", globs)
	}
}

func TestReadGitignoreAsGlobs_AncestorAnchoredInsideConfigApplies(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                "/packages/app/dist/\n",
		"packages/app/dist/app.js":  "x",
		"packages/app/src/index.js": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	gs := globSet(globs)
	assert.Assert(t, gs["dist/**/*"], "ancestor pattern anchored into config should become config-relative dist/**/*, got: %v", globs)

	cfg := RslintConfig{
		{Ignores: globs},
		{Rules: Rules{"no-debugger": "error"}},
	}
	appDist := tspath.NormalizePath(filepath.Join(appDir, "dist/app.js"))
	if cfg.GetConfigForFile(appDist, appDir) != nil {
		t.Fatalf("ancestor /packages/app/dist/ should ignore app dist; globs=%v", globs)
	}
	srcFile := tspath.NormalizePath(filepath.Join(appDir, "src/index.js"))
	if cfg.GetConfigForFile(srcFile, appDir) == nil {
		t.Fatalf("ancestor /packages/app/dist/ must not ignore src; globs=%v", globs)
	}
}

func TestReadGitignoreAsGlobs_AncestorWildcardInsideConfigApplies(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                "/packages/*/dist/\n",
		"packages/app/dist/app.js":  "x",
		"packages/app/src/index.js": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	gs := globSet(globs)
	assert.Assert(t, gs["dist/**/*"], "ancestor wildcard path should be projected under configDir, got: %v", globs)
}

func TestReadGitignoreAsGlobs_AncestorAnchoredParentDirCoversConfig(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                "/packages/\n",
		"packages/app/src/index.js": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	gs := globSet(globs)
	assert.Assert(t, gs["**/*"], "ancestor pattern covering configDir should project to **/*, got: %v", globs)
	cfg := RslintConfig{
		{Ignores: globs},
		{Rules: Rules{"no-debugger": "error"}},
	}
	srcFile := tspath.NormalizePath(filepath.Join(appDir, "src/index.js"))
	if cfg.GetConfigForFile(srcFile, appDir) != nil {
		t.Fatalf("ancestor /packages/ should ignore all files under packages/app; globs=%v", globs)
	}
}

func TestReadGitignoreAsGlobs_AncestorUnrootedParentDirCoversConfig(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                "packages/\n",
		"packages/app/src/index.js": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	gs := globSet(globs)
	assert.Assert(t, gs["**/*"], "ancestor unrooted packages/ should cover configDir, got: %v", globs)
	cfg := RslintConfig{
		{Ignores: globs},
		{Rules: Rules{"no-debugger": "error"}},
	}
	srcFile := tspath.NormalizePath(filepath.Join(appDir, "src/index.js"))
	if cfg.GetConfigForFile(srcFile, appDir) != nil {
		t.Fatalf("ancestor packages/ should ignore all files under packages/app; globs=%v", globs)
	}
}

func TestReadGitignoreAsGlobs_AncestorIgnoredConfigRootSkipsOwnGitignore(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                "packages/\n",
		"packages/app/.gitignore":   "!src/index.js\n",
		"packages/app/src/index.js": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	for _, glob := range globs {
		if strings.HasPrefix(glob, "!") {
			t.Fatalf("config root ignored by ancestor must not read own negation .gitignore, got: %v", globs)
		}
	}
	cfg := RslintConfig{
		{Ignores: globs},
		{Rules: Rules{"no-debugger": "error"}},
	}
	srcFile := tspath.NormalizePath(filepath.Join(appDir, "src/index.js"))
	if cfg.GetConfigForFile(srcFile, appDir) != nil {
		t.Fatalf("ancestor packages/ should not be re-included by packages/app/.gitignore; globs=%v", globs)
	}
}

func TestReadGitignoreAsGlobs_AncestorIgnoredIntermediateSkipsGitignore(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                "packages/\n",
		"packages/.gitignore":       "!app/src/index.js\n",
		"packages/app/src/index.js": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	for _, glob := range globs {
		if strings.HasPrefix(glob, "!") {
			t.Fatalf("ignored intermediate ancestor must not re-include from its .gitignore, got: %v", globs)
		}
	}
	cfg := RslintConfig{
		{Ignores: globs},
		{Rules: Rules{"no-debugger": "error"}},
	}
	srcFile := tspath.NormalizePath(filepath.Join(appDir, "src/index.js"))
	if cfg.GetConfigForFile(srcFile, appDir) != nil {
		t.Fatalf("ancestor packages/ should not be re-included by packages/.gitignore; globs=%v", globs)
	}
}

func TestReadGitignoreAsGlobs_AncestorChildWildcardReadsIntermediateGitignore(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                                   "packages/*\n!packages/app/\n",
		"packages/.gitignore":                          "app/src/generated/\n",
		"packages/app/src/generated/file.js":           "x",
		"packages/app/src/generated/nested/other.js":   "x",
		"packages/app/src/not-generated/file.js":       "x",
		"packages/app/src/not-generated/nested/app.js": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	globs := ReadGitignoreAsGlobs(appDir, osvfs.FS(), nil)

	gs := globSet(globs)
	assert.Assert(t, gs["src/generated/**/*"], "packages/.gitignore should be collected, got: %v", globs)
	cfg := RslintConfig{
		{Ignores: globs},
		{Rules: Rules{"no-debugger": "error"}},
	}
	generated := tspath.NormalizePath(filepath.Join(appDir, "src/generated/file.js"))
	if cfg.GetConfigForFile(generated, appDir) != nil {
		t.Fatalf("intermediate .gitignore should ignore generated file; globs=%v", globs)
	}
	normal := tspath.NormalizePath(filepath.Join(appDir, "src/not-generated/file.js"))
	if cfg.GetConfigForFile(normal, appDir) == nil {
		t.Fatalf("root negation should keep non-generated file lintable; globs=%v", globs)
	}
}

func TestReadGitignoreAsGlobsForFiles_ChildWildcardReadsIntermediateGitignore(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                         "packages/*\n!packages/app/\n",
		"packages/.gitignore":                "app/src/generated/\n",
		"packages/app/src/generated/file.js": "x",
	})
	appDir := tspath.NormalizePath(filepath.Join(dir, "packages/app"))
	generated := tspath.NormalizePath(filepath.Join(appDir, "src/generated/file.js"))

	globs := ReadGitignoreAsGlobsForFiles(appDir, osvfs.FS(), []string{generated})

	gs := globSet(globs)
	assert.Assert(t, gs["src/generated/**/*"], "explicit-file gitignore chain should include packages/.gitignore, got: %v", globs)
	cfg := RslintConfig{
		{Ignores: globs},
		{Rules: Rules{"no-debugger": "error"}},
	}
	if cfg.GetConfigForFile(generated, appDir) != nil {
		t.Fatalf("intermediate .gitignore should ignore explicit generated file; globs=%v", globs)
	}
}

func TestReadGitignoreAsGlobs_DescendantChildWildcardReadsIntermediateGitignore(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                         "packages/*\n!packages/app/\n",
		"packages/.gitignore":                "app/src/generated/\n",
		"packages/app/src/generated/file.js": "x",
		"packages/app/src/index.js":          "x",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), nil)

	gs := globSet(globs)
	assert.Assert(t, gs["packages/app/src/generated/**/*"], "root scan should read packages/.gitignore, got: %v", globs)
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

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), ParseIgnorePatterns([]string{"**/tests/**"}))
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

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), ParseIgnorePatterns([]string{"**/tests/**/*"}))
	gs := globSet(globs)

	// 2 globs: root dist + tests snapshots (file-level ignore doesn't prune dirs)
	assert.Equal(t, len(globs), 2, "should have 2 globs, got: %v", globs)
	assert.Assert(t, gs["**/dist/**/*"], "root dist")
	assert.Assert(t, gs["tests/**/snapshots/**/*"], "tests snapshots should be collected")
}

// A3: dir-level + negation — isDirAbsolutelyBlocked skips negation → still prunes.
func TestReadGitignoreAsGlobs_NegationInConfigIgnoreStillPrunes(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":           "dist/\n",
		"tests/e2e/.gitignore": "screenshots/\n",
		"tests/e2e/a.ts":       "x",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), ParseIgnorePatterns([]string{"**/tests/**", "!tests/e2e/**"}))

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

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), ParseIgnorePatterns([]string{"**/tests/**"}))
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

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), ParseIgnorePatterns([]string{"**/vendor/**"}))
	gs := globSet(globs)

	// Exactly 2 globs: root *.log + src generated. All 3 vendor .gitignore pruned.
	assert.Equal(t, len(globs), 2, "should have 2 globs, got: %v", globs)
	assert.Assert(t, gs["**/*.log"], "root *.log")
	assert.Assert(t, gs["src/**/generated/**/*"], "src generated")
}

// A7: wildcard config ignore (crates/**) — multiple nested .gitignore skipped.
func TestReadGitignoreAsGlobs_WildcardConfigIgnoreSkipsNested(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                   "dist/\n",
		"crates/core/.gitignore":       "artifacts/\n",
		"crates/binding/.gitignore":    "generated/\n",
		"packages/rspack/.gitignore":   "tmp/\n",
		"crates/core/src/lib.rs":       "x",
		"crates/binding/src/lib.rs":    "x",
		"packages/rspack/src/index.ts": "x",
	})

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), ParseIgnorePatterns([]string{"crates/**"}))
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

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), ParseIgnorePatterns([]string{}))
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

	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), ParseIgnorePatterns([]string{"build/output/**"}))
	gs := globSet(globs)

	// 2 globs: root dist + build/tools tmp. build/output cache pruned.
	assert.Equal(t, len(globs), 2, "should have 2 globs, got: %v", globs)
	assert.Assert(t, gs["**/dist/**/*"], "root dist")
	assert.Assert(t, gs["build/tools/**/tmp/**/*"], "build/tools tmp preserved")
}

// A10: ExtractConfigIgnores only extracts from global-ignore entries.
// An entry is "global ignore" when it has ONLY ignores — no files, rules, plugins, etc.
func TestExtractConfigIgnores_OnlyGlobalEntries(t *testing.T) {
	config := RslintConfig{
		{Ignores: []string{"**/tests/**"}},                            // global ignore ✓
		{Files: []string{"**/*.ts"}, Rules: Rules{"r": "error"}},      // has files+rules → NOT global
		{Ignores: []string{"scripts/**"}},                             // global ignore ✓
		{Files: []string{}, Ignores: []string{"not-global-empty"}},    // explicit empty files → NOT global (invalid at ingress)
		{Ignores: []string{"vendor/**"}, Rules: Rules{"r": "error"}},  // has rules → NOT global
		{Files: []string{"**/*.js"}, Ignores: []string{"not-global"}}, // has non-empty files → NOT global
	}

	ignores := ExtractConfigIgnores(config)

	// Entries 0 and 2 are global ignores. Entries 1, 3, 4, 5 are not.
	assert.Equal(t, len(ignores), 2, "should extract exactly 2 patterns, got: %v", ignores)
	assert.Equal(t, ignores[0].Glob, "**/tests/**")
	assert.Equal(t, ignores[1].Glob, "scripts/**")
}

// A11: multiple global ignore entries — all patterns combined.
func TestExtractConfigIgnores_MultipleEntries(t *testing.T) {
	config := RslintConfig{
		{Ignores: []string{"**/tests/**", "packages/rspack/compiled/**"}},
		{Ignores: []string{"crates/**"}},
	}

	ignores := ExtractConfigIgnores(config)
	assert.Equal(t, len(ignores), 3, "should combine all patterns, got: %v", ignores)
	assert.Equal(t, ignores[0].Glob, "**/tests/**")
	assert.Equal(t, ignores[1].Glob, "packages/rspack/compiled/**")
	assert.Equal(t, ignores[2].Glob, "crates/**")
}

// A12: bare config ignore without /** suffix (e.g., "tests" or "dist").
// isDirAbsolutelyBlocked matches the directory itself and all parents, so this
// should still prune. Verify via exact glob count.
func TestReadGitignoreAsGlobs_BareConfigIgnorePattern(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":       "*.log\n",
		"tests/.gitignore": "snapshots/\n",
		"src/.gitignore":   "generated/\n",
		"tests/a.test.ts":  "x",
		"src/a.ts":         "x",
	})

	// Bare "tests" — no trailing /**. isDirAbsolutelyBlocked("tests", ["tests"]) should
	// match via matchGlob("tests", "tests") → true → blocked.
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), ParseIgnorePatterns([]string{"tests"}))
	gs := globSet(globs)

	// 2 globs: root *.log + src/generated. tests/ pruned by bare "tests" pattern.
	assert.Equal(t, len(globs), 2, "bare 'tests' should prune tests/.gitignore, got: %v", globs)
	assert.Assert(t, gs["**/*.log"], "root *.log")
	assert.Assert(t, gs["src/**/generated/**/*"], "src generated")
}

// A13: gitignore and config ignore both target the same directory.
// collectGitignoreGlobs checks config ignores FIRST (isDirAbsolutelyBlocked),
// then the collected gitignore patterns (isDirIgnoredByGlobs). If both match,
// the directory is pruned by whichever check comes first. Verify the result is
// identical to having only one.
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
	configIgnores := ParseIgnorePatterns([]string{"**/dist/**"})
	globs := ReadGitignoreAsGlobs(dir, osvfs.FS(), configIgnores)
	gs := globSet(globs)

	// dist/ pruned by CONFIG IGNORE (isDirAbsolutelyBlocked runs first → **/dist/** matches).
	// build/ pruned by the collected gitignore glob (isDirIgnoredByGlobs → **/build/**/* matches).
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
		".gitignore":            "dist/\n",
		"src/.gitignore":        "generated/\n",
		"tests/.gitignore":      "snapshots/\n",
		"tests/unit/.gitignore": "coverage/\n",
		"tests/unit/a.test.ts":  "x",
		"src/a.ts":              "x",
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
	globsWith := ReadGitignoreAsGlobs(dir, osvfs.FS(), ParseIgnorePatterns([]string{"**/tests/**"}))
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
		".gitignore":                  "dist/\n",
		"src/a.ts":                    "x",
		"src/.gitignore":              "generated/\n",
		"tests/unit/a.test.ts":        "x",
		"tests/.gitignore":            "snapshots/\n",
		"tests/unit/.gitignore":       "coverage/\n",
		"tests/unit/deep/nested/a.ts": "x",
	})

	spy := &gitignoreSpyFS{FS: osvfs.FS()}
	globs := ReadGitignoreAsGlobs(dir, spy, ParseIgnorePatterns([]string{"**/tests/**"}))

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
		".gitignore":                "*.log\n",
		"tests/unit/a.ts":           "x",
		"tests/unit/sub/b.ts":       "x",
		"tests/.gitignore":          "tmp/\n",
		"tests/unit/.gitignore":     "output/\n",
		"tests/unit/sub/.gitignore": "cache/\n",
		"src/a.ts":                  "x",
	})

	// Without configIgnores: walks everything including tests/ subtree.
	spyWithout := &gitignoreSpyFS{FS: osvfs.FS()}
	ReadGitignoreAsGlobs(dir, spyWithout, nil)
	countWithout := len(spyWithout.accessedDirs)

	// With configIgnores: skips tests/ subtree.
	spyWith := &gitignoreSpyFS{FS: osvfs.FS()}
	ReadGitignoreAsGlobs(dir, spyWith, ParseIgnorePatterns([]string{"**/tests/**"}))
	countWith := len(spyWith.accessedDirs)

	// With configIgnores should access FEWER directories.
	assert.Assert(t, countWith < countWithout,
		"configIgnores should reduce walk count: with=%d, without=%d", countWith, countWithout)

	// Specifically: tests/, tests/unit/, tests/unit/sub/ = 3 dirs skipped.
	assert.Equal(t, countWithout-countWith, 3,
		"should skip exactly 3 dirs (tests/, tests/unit/, tests/unit/sub/), but diff is %d", countWithout-countWith)
}
