package gitignore

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

func TestConvertRepresentativePatterns(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"node_modules/", "**/node_modules/**/*"},
		{"dist/", "**/dist/**/*"},
		{"dist-*", "**/dist-*"},
		{"*.log*", "**/*.log*"},
		{"*.css.d.ts", "**/*.css.d.ts"},
		{".vscode/**/*", ".vscode/**/*"},
		{"!.vscode/settings.json", "!.vscode/settings.json"},
		{"test-results/", "**/test-results/**/*"},
		{"output/", "**/output/**/*"},
		{"**/*.rs.bk", "**/*.rs.bk"},
		{"report.[0-9]*.[0-9]*.[0-9]*.[0-9]*.json", "**/report.[0-9]*.[0-9]*.[0-9]*.[0-9]*.json"},
		{"!scripts/node_modules/", "!scripts/node_modules/**/*"},
		{"!tests/fixtures/*/**/node_modules", "!tests/fixtures/*/**/node_modules"},
		{"!packages/tool/tests/**/node_modules", "!packages/tool/tests/**/node_modules"},
		{"!tests/fixtures/cases/output", "!tests/fixtures/cases/output"},
		{"packages/*/tests/js", "packages/*/tests/js"},
		{"/github/", "github/**/*"},
		{"/artifacts/", "artifacts/**/*"},
		{"npm/**/*.node", "npm/**/*.node"},
		{"/npm/*", "npm/*"},
		{"!npm/darwin-arm64/", "!npm/darwin-arm64/**/*"},
		{"/packages/*/temp", "packages/*/temp"},
		{"build/Release", "build/Release"},
		{".env.*", "**/.env.*"},
		{"!.env.example", "!**/.env.example"},
		{".vscode/*", ".vscode/*"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			assert.Equal(t, convertSinglePattern(test.line, ""), test.want)
		})
	}
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

// --- Collector integration tests ---

func TestReadGitignoreAsGlobs_RootOnly(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore": "dist/\n*.log\n",
		"src/a.ts":   "x",
	})
	globs := readGitignoreAsGlobs(dir, osvfs.FS(), nil)
	assert.Equal(t, len(globs), 2, "got: %v", globs)

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
	globs := readGitignoreAsGlobs(dir, osvfs.FS(), nil)
	assert.Assert(t, globs == nil, "should return nil when no .gitignore")
}

func TestReadGitignoreAsGlobs_Nested(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":              "dist/\n",
		"packages/app/.gitignore": "tmp/\n",
		"src/a.ts":                "x",
	})
	globs := readGitignoreAsGlobs(dir, osvfs.FS(), nil)

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
	globs := readGitignoreAsGlobs(dir, osvfs.FS(), nil)

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
	// The collector should not enter a directory excluded by a parent pattern.
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                     "ignored/\n",
		"src/a.ts":                       "x",
		"ignored/deep/nested/.gitignore": "# should not be read\nfoo\n",
	})
	globs := readGitignoreAsGlobs(dir, osvfs.FS(), nil)

	assert.Assert(t, len(globs) >= 1, "should contain the root ignore pattern")

	// The nested ignore source is unreachable and must not be collected.
	for _, g := range globs {
		if strings.Contains(g, "ignored/deep") {
			t.Errorf("collected a pattern from an excluded directory: %s", g)
		}
	}
}

func TestCollectors_ConfigRootPatternPrunesDescendantIgnoreSource(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		"packages/app/.gitignore":           "/ignored/\n",
		"packages/app/ignored/.gitignore":   "!keep.ts\n",
		"packages/app/ignored/keep.ts":      "x",
		"packages/app/ignored/unrelated.ts": "x",
	})
	configDir := tspath.ResolvePath(dir, "packages/app")
	target := tspath.ResolvePath(configDir, "ignored/keep.ts")

	full := Collect(configDir, osvfs.FS(), nil, nil)
	explicit := Collect(configDir, osvfs.FS(), []string{target}, nil)

	assert.DeepEqual(t, full, explicit)
	assert.DeepEqual(t, full, []string{"ignored/**/*"})
}

func TestReadGitignoreAsGlobs_SkipsNodeModules(t *testing.T) {
	dir := setupGitignoreFixture(t, map[string]string{
		".gitignore":                  "dist/\n",
		"node_modules/pkg/.gitignore": "# should not be read\n",
		"src/a.ts":                    "x",
	})
	globs := readGitignoreAsGlobs(dir, osvfs.FS(), nil)

	// Only root .gitignore should be read
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "**/dist/**/*")
}

func TestReadGitignoreAsGlobs_SkipsDefaultDirsCaseInsensitively(t *testing.T) {
	mock := &gitignoreMockFS{
		FS: osvfs.FS(),
		entries: map[string]vfs.Entries{
			"/repo": {
				Directories: []string{"NODE_MODULES"},
				Symlinks:    map[string]struct{}{},
			},
		},
		files:           map[string]string{},
		caseSensitiveFS: false,
	}

	readGitignoreAsGlobs("/repo", mock, nil)
	for _, accessed := range mock.accessedDirs {
		if strings.EqualFold(accessed, "/repo/node_modules") {
			t.Fatalf("case-insensitive filesystem must not scan default excluded directory %q", accessed)
		}
	}
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
	globs := readGitignoreAsGlobs(dir, osvfs.FS(), nil)

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

func TestReadGitignoreAsGlobs_SkipsDescendantSymlinkCycle(t *testing.T) {
	mock := &gitignoreMockFS{
		FS: osvfs.FS(),
		files: map[string]string{
			"/repo/.gitignore":        "root-cache/\n",
			"/repo/a/.gitignore":      "nested-cache/\n",
			"/repo/a/loop/.gitignore": "cycle-cache/\n",
		},
		entries: map[string]vfs.Entries{
			"/repo": {
				Directories: []string{"a"},
				Symlinks:    map[string]struct{}{},
			},
			"/repo/a": {
				Directories: []string{"loop"},
				Symlinks:    map[string]struct{}{"loop": {}},
			},
			"/repo/a/loop": {
				Directories: []string{"a"},
				Symlinks:    map[string]struct{}{},
			},
		},
		resolvedPaths: map[string]string{"/repo/a/loop": "/repo"},
	}

	globs := readGitignoreAsGlobs("/repo", mock, nil)

	assert.DeepEqual(t, globs, []string{"**/root-cache/**/*", "a/**/nested-cache/**/*"})
	assert.DeepEqual(t, mock.accessedDirs, []string{"/repo", "/repo/a"})
	assert.Equal(t, len(mock.realpathCalls), 0, "symlink metadata should avoid Realpath fallback")
}

func TestReadGitignoreAsGlobs_LegacySymlinkCycleUsesCachedRealPaths(t *testing.T) {
	mock := &gitignoreMockFS{
		FS: osvfs.FS(),
		files: map[string]string{
			"/repo/.gitignore":        "root-cache/\n",
			"/repo/a/.gitignore":      "nested-cache/\n",
			"/repo/a/loop/.gitignore": "cycle-cache/\n",
		},
		entries: map[string]vfs.Entries{
			"/repo":   {Directories: []string{"a"}},
			"/repo/a": {Directories: []string{"loop"}},
		},
		resolvedPaths: map[string]string{
			"/repo":        "/repo",
			"/repo/a":      "/repo/a",
			"/repo/a/loop": "/repo",
		},
	}

	globs := readGitignoreAsGlobs("/repo", mock, nil)

	assert.DeepEqual(t, globs, []string{"**/root-cache/**/*", "a/**/nested-cache/**/*"})
	assert.DeepEqual(t, mock.accessedDirs, []string{"/repo", "/repo/a"})
	assert.DeepEqual(t, mock.realpathCalls, []string{"/repo", "/repo/a", "/repo/a/loop"})
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

// The dirOnly suffix-append has a dedup guard: if the built glob already ends
// "/**/*", appending again would yield "…/**/*/**/*". A trailing-slash gitignore
// line whose body already ends "/**/*" (e.g. "src/**/*/") exercises it. Without
// the guard the result over-matches one level deeper. Pin the deduped output so
// the guard can't be silently dropped.
func TestConvertGitignoreToGlobs_DirOnlySuffixDedup(t *testing.T) {
	// Body "src/**/*" (rooted via the internal '/'), trailing '/' → dirOnly.
	// glob already ends "/**/*" → must NOT become "src/**/*/**/*".
	globs := convertGitignoreToGlobs("src/**/*/\n", "")
	assert.Equal(t, len(globs), 1)
	assert.Equal(t, globs[0], "src/**/*")

	// Same with a nested baseDir prefix.
	nested := convertGitignoreToGlobs("src/**/*/\n", "pkg/app")
	assert.Equal(t, len(nested), 1)
	assert.Equal(t, nested[0], "pkg/app/src/**/*")
}

// --- Path boundary helpers ---

func TestParentDirStopsAtFilesystemRoot(t *testing.T) {
	tests := []struct {
		path   string
		parent string
	}{
		{path: "/repo/pkg", parent: "/repo"},
		{path: "/repo", parent: "/"},
		{path: "/", parent: ""},
		{path: "C:/repo", parent: "C:/"},
		{path: "C:/", parent: ""},
		{path: "//server/share/repo", parent: "//server/share/"},
		{path: "//server/share/", parent: ""},
		{path: "pkg/src", parent: "pkg"},
		{path: "pkg", parent: ""},
	}

	for _, test := range tests {
		assert.Equal(t, parentDir(test.path), test.parent, "parentDir(%q)", test.path)
	}
}

func TestRelativeDirIsVolumeAndShareAware(t *testing.T) {
	tests := []struct {
		name string
		root string
		dir  string
		rel  string
		ok   bool
	}{
		{name: "posix", root: "/repo", dir: "/repo/pkg", rel: "pkg", ok: true},
		{name: "posix sibling", root: "/repo", dir: "/repository/pkg", ok: false},
		{name: "drive", root: "C:/repo", dir: "c:/repo/pkg", rel: "pkg", ok: true},
		{name: "other drive", root: "C:/repo", dir: "D:/repo/pkg", ok: false},
		{name: "unc", root: "//server/share/repo", dir: "//SERVER/SHARE/repo/pkg", rel: "pkg", ok: true},
		{name: "other share", root: "//server/share/repo", dir: "//server/other/repo/pkg", ok: false},
		{name: "share prefix", root: "//server/share/repo", dir: "//server/share-other/repo/pkg", ok: false},
		{name: "other server", root: "//server/share/repo", dir: "//other/share/repo/pkg", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rel, ok := relativeDir(test.root, test.dir, true)
			assert.Equal(t, ok, test.ok)
			assert.Equal(t, rel, test.rel)
		})
	}

	rel, ok := relativeDir("/Repo", "/repo/pkg", false)
	assert.Assert(t, ok)
	assert.Equal(t, rel, "pkg")
}

func TestSortByPathDepthTreatsUNCShareAsRoot(t *testing.T) {
	paths := []string{
		"//server/share/repo/pkg",
		"//server/share/repo",
		"//server/share/",
	}

	sortByPathDepth(paths)

	assert.DeepEqual(t, paths, []string{
		"//server/share/",
		"//server/share/repo",
		"//server/share/repo/pkg",
	})
}

func TestReadGitignoreAsGlobs_StopsAtConfigDirAcrossFilesystemRoots(t *testing.T) {
	tests := []struct {
		name         string
		configDir    string
		rootIgnore   string
		configIgnore string
	}{
		{name: "posix", configDir: "/repo", rootIgnore: "/.gitignore", configIgnore: "/repo/.gitignore"},
		{name: "drive", configDir: "C:/repo", rootIgnore: "C:/.gitignore", configIgnore: "C:/repo/.gitignore"},
		{name: "unc", configDir: "//server/share/repo", rootIgnore: "//server/share/.gitignore", configIgnore: "//server/share/repo/.gitignore"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mock := &gitignoreMockFS{
				FS: osvfs.FS(),
				files: map[string]string{
					test.rootIgnore:   "root-cache/\n",
					test.configIgnore: "local-cache/\n",
				},
				entries: map[string]vfs.Entries{
					test.configDir: {Symlinks: map[string]struct{}{}},
				},
			}

			globs := readGitignoreAsGlobs(test.configDir, mock, nil)

			assert.DeepEqual(t, globs, []string{"**/local-cache/**/*"})
			assert.DeepEqual(t, mock.readFileCalls, []string{test.configIgnore})
		})
	}
}

func TestExplicitCollectorNeverAccessesOutsideConfigDir(t *testing.T) {
	t.Run("outside target only", func(t *testing.T) {
		mock := &gitignoreMockFS{
			FS: osvfs.FS(),
			files: map[string]string{
				"/.gitignore":         "global-cache/\n",
				"/outside/.gitignore": "outside-cache/\n",
			},
			caseSensitiveFS: true,
		}

		globs := Collect("/repo", mock, []string{"/outside/source.ts"}, nil)
		assert.Assert(t, globs == nil, "outside target produced %v", globs)
		assert.Assert(t, len(mock.readFileCalls) == 0, "outside target triggered reads: %v", mock.readFileCalls)
		assert.Assert(t, len(mock.accessedDirs) == 0, "outside target triggered directory access: %v", mock.accessedDirs)
		assert.Assert(t, len(mock.realpathCalls) == 0, "outside target triggered realpath: %v", mock.realpathCalls)
	})

	t.Run("mixed targets collect only the valid chain", func(t *testing.T) {
		mock := &gitignoreMockFS{
			FS: osvfs.FS(),
			files: map[string]string{
				"/repo/.gitignore":     "root-cache/\n",
				"/repo/src/.gitignore": "generated.ts\n",
				"/outside/.gitignore":  "outside-cache/\n",
			},
			entries: map[string]vfs.Entries{
				"/repo": {Symlinks: map[string]struct{}{}},
			},
			caseSensitiveFS: true,
		}

		globs := Collect("/repo", mock, []string{
			"/outside/source.ts",
			"/repo/src/source.ts",
		}, nil)
		assert.DeepEqual(t, globs, []string{
			"**/root-cache/**/*",
			"src/**/generated.ts",
		})
		assert.DeepEqual(t, mock.readFileCalls, []string{
			"/repo/.gitignore",
			"/repo/src/.gitignore",
		})
		assert.DeepEqual(t, mock.accessedDirs, []string{"/repo"})
		assert.Assert(t, len(mock.realpathCalls) == 0, "mixed targets triggered unexpected realpath: %v", mock.realpathCalls)
	})
}

func TestCollectionBoundariesAreVolumeAndCaseAware(t *testing.T) {
	tests := []struct {
		name             string
		configDir        string
		stopDir          string
		candidate        string
		useCaseSensitive bool
	}{
		{name: "posix", configDir: "/repo", stopDir: "/repo/packages/app", candidate: "/repo/packages/app", useCaseSensitive: true},
		{name: "case insensitive", configDir: "/Repo", stopDir: "/repo/Packages/App", candidate: "/REPO/packages/app"},
		{name: "drive", configDir: "C:/Repo", stopDir: "c:/repo/Packages/App", candidate: "C:/REPO/packages/app"},
		{name: "unc", configDir: "//server/share/repo", stopDir: "//SERVER/SHARE/REPO/packages/app", candidate: "//server/share/repo/PACKAGES/APP"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			boundaries := normalizeCollectionBoundaries(test.configDir, []string{test.stopDir}, test.useCaseSensitive)
			assert.Equal(t, len(boundaries), 1)
			assert.Assert(t, isCollectionBoundary(test.candidate, boundaries, test.useCaseSensitive))
		})
	}

	assert.Assert(t, len(normalizeCollectionBoundaries("/repo", []string{"/other/app", "/repo"}, true)) == 0)
}

func TestCollectors_StopAtChildConfigBoundary(t *testing.T) {
	mock := &gitignoreMockFS{
		FS: osvfs.FS(),
		files: map[string]string{
			"/repo/.gitignore":                    "root-cache/\n",
			"/repo/packages/.gitignore":           "package-cache/\n",
			"/repo/packages/app/.gitignore":       "child-cache/\n",
			"/repo/packages/app/src/.gitignore":   "generated.ts\n",
			"/repo/packages/other/.gitignore":     "other-cache/\n",
			"/repo/packages/other/src/.gitignore": "other-generated.ts\n",
		},
		entries: map[string]vfs.Entries{
			"/repo":                    {Directories: []string{"packages"}, Symlinks: map[string]struct{}{}},
			"/repo/packages":           {Directories: []string{"app", "other"}, Symlinks: map[string]struct{}{}},
			"/repo/packages/app":       {Directories: []string{"src"}, Symlinks: map[string]struct{}{}},
			"/repo/packages/app/src":   {Symlinks: map[string]struct{}{}},
			"/repo/packages/other":     {Directories: []string{"src"}, Symlinks: map[string]struct{}{}},
			"/repo/packages/other/src": {Symlinks: map[string]struct{}{}},
		},
		caseSensitiveFS: true,
	}
	boundary := "/repo/packages/app"

	full := CollectWithBoundaries("/repo", mock, nil, nil, []string{boundary})
	assert.DeepEqual(t, full, []string{
		"**/root-cache/**/*",
		"packages/**/package-cache/**/*",
		"packages/other/**/other-cache/**/*",
		"packages/other/src/**/other-generated.ts",
	})
	for _, path := range append(append([]string(nil), mock.readFileCalls...), mock.accessedDirs...) {
		if _, ok := relativeDir(boundary, path, true); ok {
			t.Fatalf("collector crossed child config boundary: %q", path)
		}
	}

	mock.readFileCalls = nil
	mock.accessedDirs = nil
	mock.realpathCalls = nil
	explicit := CollectWithBoundaries(
		"/repo",
		mock,
		[]string{"/repo/packages/app/src/file.ts"},
		nil,
		[]string{boundary},
	)
	assert.DeepEqual(t, explicit, []string{
		"**/root-cache/**/*",
		"packages/**/package-cache/**/*",
	})
	for _, path := range append(append(append([]string(nil), mock.readFileCalls...), mock.accessedDirs...), mock.realpathCalls...) {
		if _, ok := relativeDir(boundary, path, true); ok {
			t.Fatalf("explicit collector crossed child config boundary: %q", path)
		}
	}
}
