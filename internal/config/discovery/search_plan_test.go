package discovery

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

type trackingStatFS struct {
	vfs.FS
	active atomic.Int32
	max    atomic.Int32
}

func (fsys *trackingStatFS) Stat(path string) vfs.FileInfo {
	active := fsys.active.Add(1)
	for {
		maximum := fsys.max.Load()
		if active <= maximum || fsys.max.CompareAndSwap(maximum, active) {
			break
		}
	}
	// Hold overlapping stat calls open long enough to make planner fan-out
	// deterministic without relying on scheduler luck.
	time.Sleep(2 * time.Millisecond)
	info := fsys.FS.Stat(path)
	fsys.active.Add(-1)
	return info
}

func TestBuildSearchPlanClassifiesExistingInputsBeforeGlobs(t *testing.T) {
	root := searchPlanFixture(t)
	writeSearchPlanFile(t, filepath.Join(root, "literal[1].js"))
	mustMkdirSearchPlan(t, filepath.Join(root, "packages", "app"))

	plan, err := BuildSearchPlan(
		osvfs.FS(),
		root,
		[]string{"literal[1].js", "packages/app"},
		SearchPlanOptions{GlobInputPaths: true, ErrorOnUnmatchedPattern: true},
	)
	if err != nil {
		t.Fatalf("BuildSearchPlan() error = %v", err)
	}

	if len(plan.ExplicitFiles) != 1 {
		t.Fatalf("ExplicitFiles = %#v, want one file", plan.ExplicitFiles)
	}
	assertSearchPlanPath(t, plan.ExplicitFiles[0].Path, filepath.Join(root, "literal[1].js"))
	if plan.ExplicitFiles[0].RawInput != "literal[1].js" || plan.ExplicitFiles[0].Order != 0 {
		t.Fatalf("ExplicitFiles[0] = %#v", plan.ExplicitFiles[0])
	}

	if len(plan.GlobSearches) != 1 {
		t.Fatalf("GlobSearches = %#v, want one search", plan.GlobSearches)
	}
	search := plan.GlobSearches[0]
	assertSearchPlanPath(t, search.BasePath, filepath.Join(root, "packages", "app"))
	assertSearchPlanPath(t, search.Patterns[0], filepath.Join(root, "packages", "app", "**"))
	if len(search.RawPatterns) != 1 || search.RawPatterns[0] != "packages/app" || search.Order != 1 {
		t.Fatalf("GlobSearches[0] = %#v", search)
	}
}

func TestBuildSearchPlanGroupsStaticAndDynamicGlobsByLexicalBase(t *testing.T) {
	root := searchPlanFixture(t)

	plan, err := BuildSearchPlan(
		osvfs.FS(),
		root,
		[]string{"src/*.js", "src/**/*.ts", "pkg*/index.js"},
		SearchPlanOptions{GlobInputPaths: true, ErrorOnUnmatchedPattern: true},
	)
	if err != nil {
		t.Fatalf("BuildSearchPlan() error = %v", err)
	}
	if len(plan.GlobSearches) != 2 {
		t.Fatalf("GlobSearches = %#v, want two searches", plan.GlobSearches)
	}

	dynamic := plan.GlobSearches[0]
	assertSearchPlanPath(t, dynamic.BasePath, root)
	if dynamic.ID != 0 || dynamic.Order != 2 || dynamic.RawPatterns[0] != "pkg*/index.js" {
		t.Fatalf("cwd search = %#v", dynamic)
	}

	static := plan.GlobSearches[1]
	assertSearchPlanPath(t, static.BasePath, filepath.Join(root, "src"))
	if static.ID != 1 || static.Order != 0 {
		t.Fatalf("static search identity = %#v", static)
	}
	if len(static.Patterns) != 2 || len(static.RawPatterns) != 2 {
		t.Fatalf("first search = %#v, want grouped patterns", static)
	}
	assertSearchPlanPath(t, static.Patterns[0], filepath.Join(root, "src", "*.js"))
	assertSearchPlanPath(t, static.Patterns[1], filepath.Join(root, "src", "**", "*.ts"))

}

func TestBuildSearchPlanDoesNotCompactNestedSearchRoots(t *testing.T) {
	root := searchPlanFixture(t)
	mustMkdirSearchPlan(t, filepath.Join(root, "packages"))
	mustMkdirSearchPlan(t, filepath.Join(root, "packages", "app"))

	plan, err := BuildSearchPlan(
		osvfs.FS(),
		root,
		[]string{"packages", "packages/app"},
		SearchPlanOptions{GlobInputPaths: true, ErrorOnUnmatchedPattern: true},
	)
	if err != nil {
		t.Fatalf("BuildSearchPlan() error = %v", err)
	}
	if len(plan.GlobSearches) != 2 {
		t.Fatalf("GlobSearches = %#v, want both nested roots", plan.GlobSearches)
	}
	assertSearchPlanPath(t, plan.GlobSearches[0].BasePath, filepath.Join(root, "packages"))
	assertSearchPlanPath(t, plan.GlobSearches[1].BasePath, filepath.Join(root, "packages", "app"))
}

func TestBuildSearchPlanPreservesMixedInputOrder(t *testing.T) {
	root := searchPlanFixture(t)
	writeSearchPlanFile(t, filepath.Join(root, "first.js"))
	writeSearchPlanFile(t, filepath.Join(root, "last.js"))

	plan, err := BuildSearchPlan(
		osvfs.FS(),
		root,
		[]string{"first.js", "src/*.js", "last.js", "src/**/*.ts"},
		SearchPlanOptions{GlobInputPaths: true, ErrorOnUnmatchedPattern: true},
	)
	if err != nil {
		t.Fatalf("BuildSearchPlan() error = %v", err)
	}
	if len(plan.ExplicitFiles) != 2 || plan.ExplicitFiles[0].Order != 0 || plan.ExplicitFiles[1].Order != 2 {
		t.Fatalf("ExplicitFiles = %#v", plan.ExplicitFiles)
	}
	if len(plan.GlobSearches) != 1 || plan.GlobSearches[0].Order != 1 {
		t.Fatalf("GlobSearches = %#v", plan.GlobSearches)
	}
	if got := plan.GlobSearches[0].RawPatterns; len(got) != 2 || got[0] != "src/*.js" || got[1] != "src/**/*.ts" {
		t.Fatalf("RawPatterns = %#v", got)
	}
}

func TestBuildSearchPlanSingleThreadedDisablesStatFanout(t *testing.T) {
	root := searchPlanFixture(t)
	inputs := make([]string, 0, 24)
	for index := range 24 {
		name := fmt.Sprintf("file-%02d.js", index)
		writeSearchPlanFile(t, filepath.Join(root, name))
		inputs = append(inputs, name)
	}
	run := func(singleThreaded bool) int32 {
		fsys := &trackingStatFS{FS: osvfs.FS()}
		_, err := BuildSearchPlan(fsys, root, inputs, SearchPlanOptions{
			GlobInputPaths:          true,
			ErrorOnUnmatchedPattern: true,
			SingleThreaded:          singleThreaded,
		})
		if err != nil {
			t.Fatalf("BuildSearchPlan(singleThreaded=%v): %v", singleThreaded, err)
		}
		return fsys.max.Load()
	}
	if maximum := run(true); maximum != 1 {
		t.Fatalf("single-threaded stat concurrency = %d, want 1", maximum)
	}
	if maximum := run(false); maximum <= 1 {
		t.Fatalf("parallel stat concurrency = %d, want >1", maximum)
	}
}

func TestBuildSearchPlanReportsMissingLiteralWithTypedError(t *testing.T) {
	root := searchPlanFixture(t)

	_, err := BuildSearchPlan(
		osvfs.FS(),
		root,
		[]string{"missing.js"},
		SearchPlanOptions{GlobInputPaths: true, ErrorOnUnmatchedPattern: true},
	)
	var notFound *NoFilesFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("error = %v, want *NoFilesFoundError", err)
	}
	if notFound.Pattern != "missing.js" || notFound.GlobDisabled {
		t.Fatalf("NoFilesFoundError = %#v", notFound)
	}
	if !errors.Is(err, ErrNoFilesFound) {
		t.Fatalf("errors.Is(%v, ErrNoFilesFound) = false", err)
	}

	_, err = BuildSearchPlan(
		osvfs.FS(),
		root,
		[]string{"src/*.js"},
		SearchPlanOptions{ErrorOnUnmatchedPattern: true},
	)
	if !errors.As(err, &notFound) || !notFound.GlobDisabled {
		t.Fatalf("glob-disabled error = %#v, want typed disabled-glob error", err)
	}
}

func TestBuildSearchPlanCanIgnoreMissingInputs(t *testing.T) {
	root := searchPlanFixture(t)
	plan, err := BuildSearchPlan(
		osvfs.FS(),
		root,
		[]string{"missing.js"},
		SearchPlanOptions{GlobInputPaths: true},
	)
	if err != nil {
		t.Fatalf("BuildSearchPlan() error = %v", err)
	}
	if len(plan.ExplicitFiles) != 0 || len(plan.GlobSearches) != 0 {
		t.Fatalf("plan = %#v, want empty plan", plan)
	}
}

func TestBuildSearchPlanNormalizesPOSIXBackslashLikeESLintGlobSearch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows uses backslash as a path separator")
	}
	root := searchPlanFixture(t)
	pattern := `literal\*/**/*.js`
	plan, err := BuildSearchPlan(
		osvfs.FS(),
		root,
		[]string{pattern},
		SearchPlanOptions{GlobInputPaths: true, ErrorOnUnmatchedPattern: true},
	)
	if err != nil {
		t.Fatalf("BuildSearchPlan() error = %v", err)
	}
	if len(plan.GlobSearches) != 1 || len(plan.GlobSearches[0].Patterns) != 1 {
		t.Fatalf("plan = %#v", plan)
	}
	wantPattern := tspath.NormalizeSlashes(filepath.Join(root, pattern))
	if got := plan.GlobSearches[0].Patterns[0]; got != wantPattern {
		t.Fatalf("planned pattern = %q, want ESLint normalizeToPosix spelling %q", got, wantPattern)
	}
	if got := plan.GlobSearches[0].BasePath; got != tspath.NormalizePath(filepath.Join(root, "literal")) {
		t.Fatalf("base path = %q, want normalized glob parent", got)
	}
}

func TestBuildSearchPlanKeepsNativeBackslashDirectoryBaseButNormalizesItsGlob(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows uses backslash as a path separator")
	}
	root := searchPlanFixture(t)
	directory := filepath.Join(root, `literal\*`)
	mustMkdirSearchPlan(t, directory)

	plan, err := BuildSearchPlan(
		discoveryTestFS(),
		root,
		[]string{`literal\*`},
		SearchPlanOptions{GlobInputPaths: true, ErrorOnUnmatchedPattern: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.GlobSearches) != 1 {
		t.Fatalf("plan = %#v", plan)
	}
	search := plan.GlobSearches[0]
	if search.BasePath != directory {
		t.Fatalf("base path = %q, want native spelling %q", search.BasePath, directory)
	}
	if got, want := search.Patterns[0], tspath.NormalizeSlashes(directory)+"/**"; got != want {
		t.Fatalf("pattern = %q, want slash-normalized %q", got, want)
	}
	if got := search.RawPatterns[0]; got != "literal/*" {
		t.Fatalf("raw pattern = %q, want ESLint diagnostic spelling", got)
	}
}

func TestBuildSearchPlanFollowsStatSymlinkWithoutChangingLexicalIdentity(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires additional privileges on Windows")
	}
	root := searchPlanFixture(t)
	realDirectory := filepath.Join(root, "real-directory")
	mustMkdirSearchPlan(t, realDirectory)
	writeSearchPlanFile(t, filepath.Join(realDirectory, "real.js"))
	if err := os.Symlink(realDirectory, filepath.Join(root, "directory-alias")); err != nil {
		t.Fatalf("Symlink(directory) error = %v", err)
	}
	if err := os.Symlink(filepath.Join(realDirectory, "real.js"), filepath.Join(root, "file-alias.js")); err != nil {
		t.Fatalf("Symlink(file) error = %v", err)
	}

	plan, err := BuildSearchPlan(
		osvfs.FS(),
		root,
		[]string{"directory-alias", "file-alias.js"},
		SearchPlanOptions{GlobInputPaths: true, ErrorOnUnmatchedPattern: true},
	)
	if err != nil {
		t.Fatalf("BuildSearchPlan() error = %v", err)
	}
	if len(plan.GlobSearches) != 1 || len(plan.ExplicitFiles) != 1 {
		t.Fatalf("plan = %#v", plan)
	}
	assertSearchPlanPath(t, plan.GlobSearches[0].BasePath, filepath.Join(root, "directory-alias"))
	assertSearchPlanPath(t, plan.ExplicitFiles[0].Path, filepath.Join(root, "file-alias.js"))
}

func searchPlanFixture(t *testing.T) string {
	t.Helper()
	return tspath.NormalizePath(t.TempDir())
}

func mustMkdirSearchPlan(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func writeSearchPlanFile(t *testing.T, path string) {
	t.Helper()
	mustMkdirSearchPlan(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte("export {};\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func assertSearchPlanPath(t *testing.T, got string, want string) {
	t.Helper()
	want = tspath.NormalizePath(want)
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}
