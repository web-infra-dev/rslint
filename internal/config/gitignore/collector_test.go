// cspell:ignore gitdir

package gitignore

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
	"gotest.tools/v3/assert"
	"slices"
	"sync"
	"testing"
	"time"
)

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

func TestConvertGitignoreToPatterns_PreservesDirectoryNode(t *testing.T) {
	patterns := convertGitignoreToPatterns("!dist-path/\n!dist-path/**/*\n", "")
	assert.DeepEqual(t, patterns, []Pattern{
		{
			Glob:          "!**/dist-path/**/*",
			NodeGlob:      "**/dist-path",
			Negated:       true,
			DirectoryOnly: true,
		},
		{
			Glob:     "!dist-path/**/*",
			NodeGlob: "dist-path/**/*",
			Negated:  true,
		},
	})
}

func TestConvertGitignoreToPatterns_DistinguishesBareAndTrailingDoublestar(t *testing.T) {
	patterns := convertGitignoreToPatterns("**\nfoo/**\nfoo/**/\n", "")
	assert.DeepEqual(t, patterns, []Pattern{
		{
			Glob:     "**/**",
			NodeGlob: "**/**",
		},
		{
			Glob:         "foo/**",
			NodeGlob:     "foo/**",
			ContentsOnly: true,
		},
		{
			Glob:          "foo/**/*",
			NodeGlob:      "foo/**",
			DirectoryOnly: true,
			ContentsOnly:  true,
		},
	})
}

func TestConvertGitignoreToGlobs_CommentsAndBlanks(t *testing.T) {
	globs := convertGitignoreToGlobs("# comment\n\ndist/\n\n# another\n*.log\n", "")
	assert.Equal(t, len(globs), 2)
	assert.Equal(t, globs[0], "**/dist/**/*")
	assert.Equal(t, globs[1], "**/*.log")
}

func TestConvertGitignoreToGlobs_WhitespaceAndEscaping(t *testing.T) {
	globs := convertGitignoreToGlobs("  dist/  \n  coverage/\n", "")
	assert.Equal(t, len(globs), 2)
	assert.Equal(t, globs[0], "**/  dist/**/*")
	assert.Equal(t, globs[1], "**/  coverage/**/*")

	globs = convertGitignoreToGlobs("\\!important.ts\n\\#generated.ts\nname\\ with\\ spaces.ts\ntrailing\\ \nliteral\\*.ts\n{one,two}.ts\n", "")
	assert.DeepEqual(t, globs, []string{
		"**/!important.ts",
		"**/#generated.ts",
		"**/name with spaces.ts",
		"**/trailing ",
		"**/literal[*].ts",
		"**/[{]one,two[}].ts",
	})
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

func TestConvertGitignoreToPatterns_QuotesNestedFilesystemDirectory(t *testing.T) {
	patterns := convertGitignoreToPatterns("ignored.ts\n", "!pkg[1]{x}*")
	assert.DeepEqual(t, patterns, []Pattern{{
		Glob:     "./!pkg[[]1][{]x[}][*]/**/ignored.ts",
		NodeGlob: "!pkg[[]1][{]x[}][*]/**/ignored.ts",
	}})
}

// --- Collector integration tests ---

func TestConvertGitignoreToGlobs_ScopeIsolation(t *testing.T) {
	// packages/app/.gitignore has "src/" → should only affect packages/app/
	globs := convertGitignoreToGlobs("src/\n", "packages/app")
	assert.Equal(t, len(globs), 1)
	// Must be prefixed with packages/app, NOT a bare **/src/**/*
	assert.Equal(t, globs[0], "packages/app/**/src/**/*")
}

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

func TestCursorDirectoryReachability(t *testing.T) {
	for _, test := range []struct {
		name        string
		content     string
		directory   string
		wantBlocked bool
	}{
		{
			name:        "directory remains ignored",
			content:     "ignored/\n",
			directory:   "C:/repo/ignored",
			wantBlocked: true,
		},
		{
			name:        "exact directory negation reopens",
			content:     "ignored/\n!ignored/\n",
			directory:   "C:/repo/ignored",
			wantBlocked: false,
		},
		{
			name:        "file negation does not reopen directory",
			content:     "ignored/\n!ignored/keep.ts\n",
			directory:   "C:/repo/ignored",
			wantBlocked: true,
		},
		{
			name:        "descendant negation does not reopen directory",
			content:     "ignored/\n!ignored/**/*\n",
			directory:   "C:/repo/ignored",
			wantBlocked: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			cursor := NewCursor("c:/REPO", false)
			cursor, _ = cursor.AppendSource("C:/repo", test.content)
			_, blocked := cursor.Enter(test.directory)
			assert.Equal(t, blocked, test.wantBlocked)
		})
	}
}

func TestCursorTrailingDoublestarMatchesNestedRepeatedName(t *testing.T) {
	cursor := NewCursor("/repo", true)
	cursor, _ = cursor.AppendSource("/repo", "**/cache/**\n")

	cache, blocked := cursor.Enter("/repo/cache")
	assert.Assert(t, !blocked, "trailing /** must not match the directory before it")
	_, blocked = cache.Enter("/repo/cache/cache")
	assert.Assert(t, blocked, "a repeated name below the matched directory is still its content")
}

func TestCursorBlockSourceTraversalPreservesInheritedRules(t *testing.T) {
	cursor := NewCursor("/repo", true)
	cursor, _ = cursor.AppendSource("/repo", "ignored/\n")
	cursor = cursor.BlockSourceTraversal()

	next, blocked := cursor.Enter("/repo/ignored")
	assert.Assert(t, blocked)
	assert.Assert(t, !next.SourceReachable())

	next, globs := cursor.AppendSource("/repo", "must-not-apply/\n")
	assert.Assert(t, !next.SourceReachable())
	assert.Equal(t, len(globs), 0)
	_, blocked = next.Enter("/repo/must-not-apply")
	assert.Assert(t, !blocked)
}

func TestCursorRejectsPathsOutsideOwnerVolume(t *testing.T) {
	for _, test := range []struct {
		name      string
		root      string
		directory string
	}{
		{name: "different drive", root: "C:/repo", directory: "D:/repo"},
		{name: "different UNC share", root: "//server/share/repo", directory: "//server/other/repo"},
	} {
		t.Run(test.name, func(t *testing.T) {
			cursor := NewCursor(test.root, false)
			next, blocked := cursor.Enter(test.directory)
			assert.Assert(t, blocked)
			assert.Assert(t, !next.SourceReachable())
			next, globs := cursor.AppendSource(test.directory, "ignored/\n")
			assert.Equal(t, len(globs), 0)
			assert.Assert(t, !next.SourceReachable())
		})
	}
}

type readerTestFS struct {
	vfs.FS
	mu            sync.Mutex
	files         map[string]string
	directories   map[string]bool
	readCalls     map[string]int
	markerCalls   map[string]int
	caseSensitive bool
	beforeRead    func(string)
	beforeMarker  func(string)
}

func (fsys *readerTestFS) CaseSensitivity() tspath.CaseSensitivity {
	return pathutil.CaseSensitivity(fsys.caseSensitive)
}
func (fsys *readerTestFS) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	if fsys.beforeMarker != nil {
		fsys.beforeMarker(path.AsString())
	}
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	fsys.markerCalls[path.AsString()]++
	return fsys.directories[path.AsString()]
}
func (fsys *readerTestFS) FileExists(path tspath.RootedFilePath) bool {
	if fsys.beforeMarker != nil {
		fsys.beforeMarker(path.AsString())
	}
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	fsys.markerCalls[path.AsString()]++
	_, ok := fsys.files[path.AsString()]
	return ok
}
func (fsys *readerTestFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	if fsys.beforeRead != nil {
		fsys.beforeRead(path.AsString())
	}
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	fsys.readCalls[path.AsString()]++
	content, ok := fsys.files[path.AsString()]
	return content, ok
}

func newReaderTestFS(files map[string]string, directories ...string) *readerTestFS {
	fsys := &readerTestFS{files: files, directories: make(map[string]bool), readCalls: make(map[string]int), markerCalls: make(map[string]int), caseSensitive: true}
	for _, directory := range directories {
		fsys.directories[directory] = true
	}
	return fsys
}

func TestReaderNearestGitBoundaryIncludesDirectoryAndFileMarkers(t *testing.T) {
	for _, markerIsDirectory := range []bool{false, true} {
		fsys := newReaderTestFS(map[string]string{"/.gitignore": "outside/\n", "/work/.gitignore": "dist/\n"})
		if markerIsDirectory {
			fsys.directories["/work/.git"] = true
		} else {
			fsys.files["/work/.git"] = "gitdir: unrelated"
		}
		reader := NewReader(fsys)
		cursor := reader.Ancestors("/work/package")
		if cursor.RootDirectory() != "/work" {
			t.Fatalf("root=%q", cursor.RootDirectory())
		}
		if fsys.readCalls["/.gitignore"] != 0 || fsys.readCalls["/work/.git"] != 0 {
			t.Fatal("read above boundary or parsed marker contents")
		}
		if fsys.readCalls["/work/package/.gitignore"] != 0 {
			t.Fatal("Ancestors observed local source prematurely")
		}
		cursor = reader.Observe(cursor, "/work/package")
		_, blocked := reader.Enter(cursor, "/work/package/dist")
		if !blocked {
			t.Fatal("boundary local rules were not inherited")
		}
	}
}

func TestReaderNoGitBoundaryStopsAtVolumeAndUNCShareRoots(t *testing.T) {
	for _, test := range []struct{ directory, root string }{
		{"/work/pkg", "/"}, {"C:/work/pkg", "C:/"}, {"//server/share/work/pkg", "//server/share/"},
	} {
		fsys := newReaderTestFS(map[string]string{tspath.CombinePaths(test.root, ".gitignore"): "dist/\n"})
		reader := NewReader(fsys)
		cursor := reader.Ancestors(test.directory)
		if cursor.RootDirectory() != test.root {
			t.Errorf("%s root=%s want=%s", test.directory, cursor.RootDirectory(), test.root)
		}
		if fsys.readCalls[tspath.CombinePaths(test.root, ".gitignore")] != 1 {
			t.Errorf("root source missing for %s", test.directory)
		}
		for path := range fsys.markerCalls {
			if path == "//server/.git" {
				t.Fatal("escaped UNC share")
			}
		}
	}
}

func TestReaderIgnoredNestedMarkerDoesNotReopenEntry(t *testing.T) {
	fsys := newReaderTestFS(map[string]string{
		"/repo/.git": "", "/repo/.gitignore": "nested/\n",
		"/repo/nested/.git": "gitdir: elsewhere", "/repo/nested/.gitignore": "ignored.ts\n",
	})
	reader := NewReader(fsys)
	outer := reader.Observe(reader.Ancestors("/repo"), "/repo")
	_, blocked := reader.Enter(outer, "/repo/nested")
	if !blocked {
		t.Error("nested marker reopened an ignored directory entry")
	}
	if fsys.readCalls["/repo/nested/.gitignore"] != 0 {
		t.Error("ignored directory source was read")
	}
	inner := reader.Observe(reader.Ancestors("/repo/nested"), "/repo/nested")
	if !inner.SourceReachable() || inner.RootDirectory() != "/repo/nested" || fsys.readCalls["/repo/nested/.gitignore"] != 1 {
		t.Fatal("direct nested input did not establish its own Git entry point")
	}
}

func TestReaderNestedMarkerResetsUnreachableRawSources(t *testing.T) {
	fsys := newReaderTestFS(map[string]string{
		"/repo/.git": "", "/repo/.gitignore": "ignored/\n",
		"/repo/ignored/.gitignore": "*.ts\n", "/repo/ignored/inner/.git": "gitdir: elsewhere",
		"/repo/ignored/inner/.gitignore": "*.generated.ts\n",
	})
	reader := NewReader(fsys)
	cursor := reader.Observe(reader.Ancestors("/repo"), "/repo")
	blockedCursor, blocked := reader.Enter(cursor, "/repo/ignored")
	if !blocked || blockedCursor.SourceReachable() {
		t.Fatal("ignored source remained reachable")
	}
	reader.Observe(blockedCursor, "/repo/ignored")
	inner, blocked := reader.Enter(blockedCursor, "/repo/ignored/inner")
	if blocked || !inner.SourceReachable() || inner.RootDirectory() != "/repo/ignored/inner" {
		t.Fatal("nested marker did not establish fresh Git state")
	}
	reader.Observe(inner, "/repo/ignored/inner")
	if fsys.readCalls["/repo/ignored/.gitignore"] != 0 || fsys.readCalls["/repo/ignored/inner/.gitignore"] != 1 {
		t.Fatalf("source reads=%v", fsys.readCalls)
	}
	snapshot := reader.Snapshot()
	if len(snapshot.Roots) != 2 {
		t.Fatalf("roots=%v", snapshot.Roots)
	}
	if !slices.Contains(snapshot.CoveredDirectories, "/repo/ignored") {
		t.Fatal("proven unreachable source was not covered")
	}
}

func TestReaderSourceObservationsRemainFrozenForOneRequest(t *testing.T) {
	const directory = "/repo/src"
	const source = directory + "/.gitignore"
	for _, test := range []struct {
		name          string
		exists        bool
		content       string
		removeOnWrite bool
	}{
		{name: "missing"},
		{name: "empty", exists: true},
		{name: "replaced", exists: true, content: "before/\n"},
		{name: "deleted", exists: true, content: "before/\n", removeOnWrite: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			fsys := newReaderTestFS(map[string]string{"/repo/.git": ""})
			if test.exists {
				fsys.files[source] = test.content
			}
			reader := NewReader(fsys)
			reader.Observe(reader.Ancestors(directory), directory)
			first := reader.Snapshot()
			if test.removeOnWrite {
				delete(fsys.files, source)
			} else {
				fsys.files[source] = "after/\n"
			}

			cursor := reader.Observe(reader.Ancestors(directory), directory)
			assert.DeepEqual(t, reader.Snapshot(), first)
			_, beforeIgnored := reader.Enter(cursor, directory+"/before")
			_, afterIgnored := reader.Enter(cursor, directory+"/after")
			assert.Equal(t, beforeIgnored, test.content != "")
			assert.Assert(t, !afterIgnored)
			assert.Equal(t, fsys.readCalls[source], 1)

			fresh := NewReader(fsys)
			cursor = fresh.Observe(fresh.Ancestors(directory), directory)
			_, beforeIgnored = fresh.Enter(cursor, directory+"/before")
			_, afterIgnored = fresh.Enter(cursor, directory+"/after")
			assert.Assert(t, !beforeIgnored)
			assert.Equal(t, afterIgnored, !test.removeOnWrite)
			assert.Equal(t, fsys.readCalls[source], 2)
		})
	}
}

func TestReaderConcurrentSourceAndMarkerMissesAreObservedOnce(t *testing.T) {
	fsys := newReaderTestFS(map[string]string{"/repo/.git": "", "/repo/.gitignore": "dist/\n"})
	reader := NewReader(fsys)
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() { defer wait.Done(); reader.Observe(reader.Ancestors("/repo/src"), "/repo/src") }()
	}
	wait.Wait()
	for _, path := range []string{"/repo/.gitignore", "/repo/src/.gitignore"} {
		if fsys.readCalls[path] != 1 {
			t.Errorf("%s read %d times", path, fsys.readCalls[path])
		}
	}
	if fsys.markerCalls["/repo/src/.git"] != 2 {
		t.Fatalf("marker misses repeated: %v", fsys.markerCalls)
	}
	first := reader.Snapshot()
	first.Roots[0] = "/changed"
	first.Sources[0].Patterns[0].Glob = "changed"
	second := reader.Snapshot()
	if second.Roots[0] != "/repo" || second.Sources[0].Patterns[0].Glob == "changed" {
		t.Fatal("snapshot retained writable reader state")
	}
}

func TestReaderDifferentDirectoriesDoNotSerializeIO(t *testing.T) {
	for _, kind := range []string{"marker", "source"} {
		t.Run(kind, func(t *testing.T) {
			fsys := newReaderTestFS(map[string]string{"/repo/.git": "", "/repo/a/.gitignore": "dist/\n"})
			started, release := make(chan struct{}), make(chan struct{})
			var entered, released sync.Once
			defer released.Do(func() { close(release) })
			hook := func(path string) {
				if path == "/repo/a/.git" || path == "/repo/a/.gitignore" {
					entered.Do(func() { close(started) })
					<-release
				}
			}
			if kind == "marker" {
				fsys.beforeMarker = hook
			} else {
				fsys.beforeRead = hook
			}
			reader := NewReader(fsys)
			root := reader.Ancestors("/repo")
			observe := func(directory string) {
				if kind == "marker" {
					reader.Enter(root, directory)
				} else {
					reader.Observe(root, directory)
				}
			}
			firstDone := make(chan struct{})
			go func() { defer close(firstDone); observe("/repo/a") }()
			select {
			case <-started:
			case <-time.After(2 * time.Second):
				t.Fatal("first directory did not start IO")
			}
			secondDone := make(chan struct{})
			go func() { defer close(secondDone); observe("/repo/b") }()
			select {
			case <-secondDone:
			case <-time.After(2 * time.Second):
				t.Fatal("unrelated directory waited for blocked IO")
			}
			released.Do(func() { close(release) })
			<-firstDone
		})
	}
}

func TestReaderSharedSourceKeepsEachRoutesInheritedRules(t *testing.T) {
	fsys := newReaderTestFS(map[string]string{"/repo/.git": "", "/repo/shared/.gitignore": "/dist/\n"})
	reader := NewReader(fsys)
	root := reader.Ancestors("/repo")
	left, _ := root.AppendSourcePatterns("/repo", "left/\n")
	right, _ := root.AppendSourcePatterns("/repo", "right/\n")
	left = reader.Observe(left, "/repo/shared")
	right = reader.Observe(right, "/repo/shared")
	for _, test := range []struct {
		cursor       Cursor
		own, sibling string
	}{{left, "left", "right"}, {right, "right", "left"}} {
		_, ownIgnored := test.cursor.Enter("/repo/" + test.own)
		_, siblingIgnored := test.cursor.Enter("/repo/" + test.sibling)
		_, localIgnored := test.cursor.Enter("/repo/shared/dist")
		assert.Assert(t, ownIgnored && !siblingIgnored && localIgnored)
	}
	assert.Equal(t, fsys.readCalls["/repo/shared/.gitignore"], 1)
	snapshot := reader.Snapshot()
	assert.Equal(t, len(snapshot.Sources), 1)
	assert.Equal(t, len(snapshot.Sources[0].Patterns), 1)
	assert.Equal(t, snapshot.Sources[0].Patterns[0].Glob, "shared/dist/**/*")
}

// convertSinglePattern exposes the projection used by collector tests.
func convertSinglePattern(line string, baseDir string) string {
	patterns := convertGitignoreToPatterns(line, baseDir)
	if len(patterns) == 0 {
		return ""
	}
	return patterns[0].Glob
}
