package suppressions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"gotest.tools/v3/assert"
)

func tempWorkingDirectory(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	assert.NilError(t, err)
	return dir
}

func TestScopedUpdate(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	checked := filepath.ToSlash(filepath.Join(cwd, "a.ts"))
	unchecked := filepath.ToSlash(filepath.Join(cwd, "fallback.ts"))
	private := filepath.Join(cwd, "private")
	assert.NilError(t, os.Mkdir(private, 0o755))
	for _, name := range []string{"a.ts", "fallback.ts", "private/untouched.ts", "former-dir"} {
		assert.NilError(t, os.WriteFile(filepath.Join(cwd, name), nil, 0o644))
	}
	linkErr := os.Symlink("gone.ts", filepath.Join(cwd, "linked.ts"))
	if linkErr != nil && runtime.GOOS != "windows" {
		t.Fatal(linkErr)
	}
	remote := filepath.Join(cwd, "remote")
	assert.NilError(t, os.Mkdir(remote, 0o755))
	makeDirectoryLink(t, remote, filepath.Join(cwd, "linked-dir"))
	assert.NilError(t, os.Remove(remote))
	assert.NilError(t, os.Chmod(private, 0))
	t.Cleanup(func() { _ = os.Chmod(private, 0o755) })
	path := filepath.Join(cwd, DefaultFilename)
	driveKey := "D:/shared/file.ts"
	if runtime.GOOS == "windows" {
		driveKey = filepath.ToSlash(filepath.Join(cwd, "missing-on-mounted-drive.ts"))
	}
	initial := fmt.Sprintf(`{
		"a.ts":{"x":{"count":3},"y":{"count":1},"typed":{"count":2}},
		"fallback.ts":{"x":{"count":1}},
		"private/untouched.ts":{"x":{"count":1}},
		"gone.ts":{"x":{"count":1}},
		"incomplete.ts":{"plugin/check":{"count":1}},
		"linked.ts":{"x":{"count":1}},
		"linked-dir/src/file.ts":{"x":{"count":1}},
		"former-dir/old.ts":{"x":{"count":1}},
		"^/old.ts":{"x":{"count":1}},
		%q:{"x":{"count":1}},
		"//server/share/file.ts":{"x":{"count":1}}
	}`, driveKey)
	assert.NilError(t, os.WriteFile(path, []byte(initial), 0o644))
	diagnostics := []rule.RuleDiagnostic{
		{FilePath: checked, RuleName: "x", Severity: rule.SeverityError},
		{FilePath: checked, RuleName: "z", Severity: rule.SeverityWarning},
		{FilePath: unchecked, RuleName: "x", Severity: rule.SeverityError},
	}
	run := func(options Options, unused bool) *File {
		t.Helper()
		file, err := Open(cwd, options)
		assert.NilError(t, err)
		results := []FileResult{
			{Path: checked, SkippedRules: []string{"typed"}},
			{Path: unchecked, Unchecked: true},
			// The plugin could not read a file that disappeared during lint.
			{Path: filepath.ToSlash(filepath.Join(cwd, "incomplete.ts")), SkippedRules: []string{"plugin/check"}, Incomplete: true},
		}
		visible, stale, err := file.Apply(t.Context(), results, append([]rule.RuleDiagnostic(nil), diagnostics...))
		if err != nil || stale != unused || len(visible) != 2 || visible[0].RuleName != "z" || visible[1].FilePath != unchecked {
			t.Fatalf("apply: %v, stale %v, error %v", visible, stale, err)
		}
		return file
	}
	run(Options{}, true)
	recorded := run(Options{Mode: Record, Rules: []string{"x"}}, true)
	if got := recorded.entries["a.ts"]; got["x"].Count != 1 || got["y"].Count != 1 {
		t.Fatalf("recorded counts: %v", got)
	}
	// Both files moved after the final observation. Its verified counts and
	// its skipped rules still govern pruning for their original paths.
	for _, name := range []string{"a.ts", "fallback.ts"} {
		assert.NilError(t, os.Rename(filepath.Join(cwd, name), filepath.Join(cwd, name+".saved")))
	}
	pruned := run(Options{Mode: Prune}, false)
	want := entries{
		"a.ts":                   {"x": {Count: 1}, "typed": {Count: 2}},
		"fallback.ts":            {"x": {Count: 1}},
		"incomplete.ts":          {"plugin/check": {Count: 1}},
		"private/untouched.ts":   {"x": {Count: 1}},
		"linked-dir/src/file.ts": {"x": {Count: 1}},
		"//server/share/file.ts": {"x": {Count: 1}},
	}
	if runtime.GOOS != "windows" {
		want[driveKey] = map[string]entry{"x": {Count: 1}}
	}
	if linkErr == nil {
		want["linked.ts"] = map[string]entry{"x": {Count: 1}}
	}
	assert.DeepEqual(t, pruned.entries, want)
}

func TestUncheckedOnlyTargetsDoNotMutateBaseline(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	broken := filepath.ToSlash(filepath.Join(cwd, "broken.ts"))
	f, err := Open(cwd, Options{Mode: Record})
	assert.NilError(t, err)
	_, _, err = f.Apply(t.Context(), []FileResult{{Path: broken, Unchecked: true}}, nil)
	assert.NilError(t, err)
	if _, err := os.Stat(filepath.Join(cwd, DefaultFilename)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no baseline to be created for unchecked-only targets, stat err=%v", err)
	}

	const initial = `{"gone.ts":{"x":{"count":1}}}`
	assert.NilError(t, os.WriteFile(filepath.Join(cwd, DefaultFilename), []byte(initial), 0o644))
	pruner, err := Open(cwd, Options{Mode: Prune})
	assert.NilError(t, err)
	_, _, err = pruner.Apply(t.Context(), []FileResult{{Path: broken, Unchecked: true}}, nil)
	assert.NilError(t, err)
	content, err := os.ReadFile(filepath.Join(cwd, DefaultFilename))
	assert.NilError(t, err)
	assert.Equal(t, string(content), initial)
}

func TestMissingBaseline(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	if file, err := Open(cwd, Options{}); err != nil || file != nil {
		t.Fatalf("plain check: file=%v, err=%v", file, err)
	}
	if _, err := Open(cwd, Options{Location: "missing.json"}); err == nil {
		t.Fatal("explicit missing baseline was accepted")
	}
	f, err := Open(cwd, Options{Mode: Prune})
	assert.NilError(t, err)
	if _, _, err := f.Apply(t.Context(), nil, nil); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(cwd, DefaultFilename)); err != nil || string(data) != "{}\n" {
		t.Fatalf("prune missing default: %q, %v", data, err)
	}
}

func reportError(ctx context.Context, f *File, path, id string) ([]rule.RuleDiagnostic, bool, error) {
	path = filepath.ToSlash(path)
	return f.Apply(ctx, []FileResult{{Path: path}}, []rule.RuleDiagnostic{
		{FilePath: path, RuleName: id, Severity: rule.SeverityError},
	})
}

func TestWorkingDirectoryPaths(t *testing.T) {
	root := tempWorkingDirectory(t)
	repo, link := filepath.Join(root, "project", "repo"), filepath.Join(root, "link")
	src := filepath.Join(repo, "src")
	assert.NilError(t, os.MkdirAll(src, 0o755))
	destination := repo
	// On case-insensitive volumes the physical path also restores the stored case.
	alternate := filepath.Join(filepath.Dir(repo), "REPO")
	if _, err := os.Stat(alternate); err == nil && runtime.GOOS != "windows" {
		destination = alternate
	}
	makeDirectoryLink(t, destination, link)
	t.Chdir(link)
	native := repo
	if runtime.GOOS == "windows" {
		native = link
	}
	for location, want := range map[string]string{
		root:             filepath.Join(root, "rslint-suppressions_"+cwdHash(native)),
		"../shared.json": filepath.Join(filepath.Dir(link), "shared.json"),
	} {
		f, err := Open(link, Options{Mode: Record, Location: location})
		assert.NilError(t, err)
		assert.Equal(t, f.path, want)
		for name, want := range map[string]string{
			filepath.Join(src, "app.js"):         "src/app.js",
			filepath.Join(link, "src", "app.js"): "src/app.js",
			filepath.Join(root, "outside.js"):    "../../outside.js",
		} {
			if got := f.fileKey(filepath.ToSlash(name)); got != want {
				t.Fatalf("fileKey(%q) = %q, want %q", name, got, want)
			}
		}
	}
}

func TestFileKeyUsesNativeRelativePaths(t *testing.T) {
	check := func(cwd, path, want string) {
		t.Helper()
		f := File{cwd: cwd}
		if got := f.fileKey(path); got != want {
			t.Errorf("fileKey(%q, %q) = %q, want %q", cwd, path, got, want)
		}
	}
	check("/tree/pkg/", "/tree/pkg-extra/File.js", "../pkg-extra/File.js")
	check("/", "/File.js", "File.js")
	check("/tree/pkg/", "/tree/Pkg/File.js", "../Pkg/File.js")
	if runtime.GOOS == "windows" {
		check("C:/tree/pkg/", "c:/tree/pkg/File.js", "File.js")
		check("C:/tree/", "d:/Shared/File.js", "d:/Shared/File.js")
		check("C:/tree/", "D:/Shared/File.js", "d:/Shared/File.js")
		check("//server/share/pkg/", "//SERVER/SHARE/pkg/File.js", "File.js")
		check(`\\server\share\pkg`, "//SERVER/SHARE/pkg/File.js", "File.js")
		check("//server/share/pkg/", "//server/other/File.js", "//server/other/File.js")
		check("//server/share", "//SERVER/SHARE/File.js", "File.js")
	}
}

func TestNormalizeDriveKeys(t *testing.T) {
	owner := json.RawMessage(`"team"`)
	ticket := json.RawMessage(`123`)
	data := entries{
		"D:/Other.js": {"x": {Count: 1}},
		"D:/File.js": {
			"x": {Count: 2, Extra: map[string]json.RawMessage{"owner": owner}},
			"y": {Count: 1},
		},
		"d:/File.js": {
			"x": {Count: 2, Extra: map[string]json.RawMessage{"ticket": ticket}},
			"z": {Count: 1},
		},
	}
	assert.NilError(t, normalizeDriveKeys(data))
	assert.DeepEqual(t, data, entries{
		"d:/Other.js": {"x": {Count: 1}},
		"d:/File.js": {
			"x": {Count: 2, Extra: map[string]json.RawMessage{"owner": owner, "ticket": ticket}},
			"y": {Count: 1}, "z": {Count: 1},
		},
	})
	data["D:/File.js"] = map[string]entry{"x": {Count: 1}}
	assert.ErrorContains(t, normalizeDriveKeys(data), `conflicting counts for rule "x"`)
}

func makeDirectoryLink(t *testing.T, destination, link string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		// cspell:ignore mklink
		if output, err := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", link, destination).CombinedOutput(); err != nil {
			t.Fatalf("create directory junction: %v\n%s", err, output)
		}
	} else if err := os.Symlink(destination, link); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateMergesLatestAndPreservesMode(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	baselinePath := filepath.Join(cwd, DefaultFilename)
	assert.NilError(t, os.WriteFile(baselinePath+".tmp", []byte("unrelated"), 0o600))
	assert.NilError(t, os.WriteFile(baselinePath, []byte(`{"test.js":{"rule0":{"count":1}}}`), 0o600))

	writer, err := Open(cwd, Options{Mode: Record})
	assert.NilError(t, err)
	// Simulate another update landing before this writer flushes.
	assert.NilError(t, os.WriteFile(baselinePath, []byte(`{"test.js":{"rule0":{"count":1},"rule1":{"count":1}}}`), 0o600))

	name := filepath.Join(cwd, "test.js")
	_, _, err = reportError(t.Context(), writer, name, "rule2")
	assert.NilError(t, err)

	if runtime.GOOS != "windows" {
		info, err := os.Stat(baselinePath)
		assert.NilError(t, err)
		assert.Equal(t, info.Mode().Perm(), os.FileMode(0o600))
	}

	leftover, err := os.ReadFile(baselinePath + ".tmp")
	assert.NilError(t, err)
	assert.Equal(t, string(leftover), "unrelated")

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := reportError(ctx, writer, name, "cancelled"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled write returned %v", err)
	}

	f, err := Open(cwd, Options{})
	want := entries{"test.js": {"rule0": {Count: 1}, "rule1": {Count: 1}, "rule2": {Count: 1}}}
	assert.NilError(t, err)
	assert.Assert(t, f != nil)
	assert.DeepEqual(t, f.entries, want)
}

func TestUpdateRequiresBaselineWriteAccess(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	baselinePath := filepath.Join(cwd, DefaultFilename)
	const original = `{"a.js":{"x":{"count":1}}}`
	assert.NilError(t, os.WriteFile(baselinePath, []byte(original), 0o444))
	t.Cleanup(func() { _ = os.Chmod(baselinePath, 0o600) })
	if file, err := os.OpenFile(baselinePath, os.O_WRONLY, 0); err == nil {
		_ = file.Close()
		t.Skip("current user can write a read-only file")
	} else if !errors.Is(err, os.ErrPermission) {
		t.Fatal(err)
	}
	f, err := Open(cwd, Options{Mode: Record})
	assert.NilError(t, err)
	// An unchanged count does not need to replace the read-only baseline.
	if _, _, err := reportError(t.Context(), f, filepath.Join(cwd, "a.js"), "x"); err != nil {
		t.Fatalf("record unchanged count on read-only baseline: got %v, want nil", err)
	}
	// A changed count must require write permission before replacing the file.
	if _, _, err := reportError(t.Context(), f, filepath.Join(cwd, "a.js"), "y"); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("record new count on read-only baseline: expected a permission error, got %v", err)
	}
	if content, err := os.ReadFile(baselinePath); err != nil || string(content) != original {
		t.Fatalf("read-only baseline: %q, %v", content, err)
	}

	// A writable existing baseline inside a read-only directory must still
	// succeed for unchanged record/prune runs without creating a lock file.
	if runtime.GOOS != "windows" {
		roDir := filepath.Join(cwd, "ro-dir")
		assert.NilError(t, os.Mkdir(roDir, 0o755))
		roBaseline := filepath.Join(roDir, "base.json")
		assert.NilError(t, os.WriteFile(roBaseline, []byte(original), 0o644))
		assert.NilError(t, os.Chmod(roDir, 0o555))
		t.Cleanup(func() { _ = os.Chmod(roDir, 0o755) })
		if err := probeDirectoryWritable(roDir); errors.Is(err, os.ErrPermission) {
			for _, mode := range []Mode{Record, Prune} {
				roFile, err := Open(cwd, Options{Mode: mode, Location: roBaseline})
				assert.NilError(t, err)
				if _, _, err := reportError(t.Context(), roFile, filepath.Join(cwd, "a.js"), "x"); err != nil {
					t.Fatalf("mode %v unchanged count in read-only dir: got %v, want nil", mode, err)
				}
			}
			recFile, err := Open(cwd, Options{Mode: Record, Location: roBaseline})
			assert.NilError(t, err)
			if _, _, err := reportError(t.Context(), recFile, filepath.Join(cwd, "a.js"), "y"); !errors.Is(err, os.ErrPermission) {
				t.Fatalf("record changed count in read-only dir: expected permission error, got %v", err)
			}
		}
	}
}

func TestUpdateThroughFileSymlink(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	path := filepath.Join(cwd, DefaultFilename)
	assert.NilError(t, os.WriteFile(path, []byte("{}"), 0o600))
	link := filepath.Join(cwd, "linked.json")
	if err := os.Symlink(DefaultFilename, link); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}
	f, err := Open(cwd, Options{Mode: Record, Location: link})
	assert.NilError(t, err)
	if _, _, err := reportError(t.Context(), f, filepath.Join(cwd, "test.js"), "x"); err != nil {
		t.Fatal(err)
	}
	defaultFile, err := Open(cwd, Options{})
	if err != nil || defaultFile == nil || defaultFile.entries["test.js"]["x"].Count != 1 {
		t.Fatalf("write via symlink: file=%v err=%v", defaultFile, err)
	}
	assert.NilError(t, os.Remove(path))
	_, _, err = reportError(t.Context(), f, filepath.Join(cwd, "test.js"), "y")
	assert.ErrorContains(t, err, "cannot resolve suppressions destination")
	if name, err := os.Readlink(link); err != nil || name != DefaultFilename {
		t.Fatalf("stored symlink = %q, err=%v", name, err)
	}
}

func TestUpdatePreservesLiteralBackslash(t *testing.T) {
	if filepath.Separator == '\\' {
		t.Skip("backslash is a path separator")
	}
	cwd := tempWorkingDirectory(t)
	const location = `cache\base.json`
	assert.NilError(t, os.Mkdir(filepath.Join(cwd, "cache"), 0o755))
	for _, name := range []string{location, "cache/base.json"} {
		assert.NilError(t, os.WriteFile(filepath.Join(cwd, name), []byte("{}"), 0o644))
	}
	f, err := Open(cwd, Options{Mode: Record, Location: location})
	assert.NilError(t, err)
	if _, _, err := reportError(t.Context(), f, filepath.Join(cwd, "test.js"), "x"); err != nil {
		t.Fatal(err)
	}
	f, err = Open(cwd, Options{Location: location})
	if err != nil || f == nil || f.entries["test.js"]["x"].Count != 1 {
		t.Fatalf("recorded baseline: file=%v err=%v", f, err)
	}
	if content, err := os.ReadFile(filepath.Join(cwd, "cache/base.json")); err != nil || string(content) != "{}" {
		t.Fatalf("other baseline: %q, %v", content, err)
	}
}

func TestDirectoryHash(t *testing.T) {
	// These values come from ESLint's imurmurhash, including a signed UTF-16 tail.
	for cwd, want := range map[string]string{
		// cspell:disable-next-line
		"/repo": "ej76rc", "/项目/😀": "1uwwl2n", `C:\dev\project`: "1w51rpt", "/x高": "115w3s7",
	} {
		if got := cwdHash(cwd); got != want {
			t.Errorf("cwdHash(%q) = %q, want %q", cwd, got, want)
		}
	}
}

func TestUpdatePreservesExtraMembers(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	for _, name := range []string{"a.js", "b.js"} {
		assert.NilError(t, os.WriteFile(filepath.Join(cwd, name), nil, 0o644))
	}
	const initial = `{
		"a.js":{"x":{"count":2,"reason":"deferred"}},
		"b.js":{"x":{"count":1,"ticket":9007199254740993}}
	}`
	baselinePath := filepath.Join(cwd, DefaultFilename)
	assert.NilError(t, os.WriteFile(baselinePath, []byte(initial), 0o644))
	for _, step := range []struct {
		mode   Mode
		reason string
	}{{Prune, `"deferred"`}, {Record, ""}} {
		file, err := Open(cwd, Options{Mode: step.mode})
		assert.NilError(t, err)
		_, _, err = reportError(t.Context(), file, filepath.Join(cwd, "a.js"), "x")
		assert.NilError(t, err)
		encoded, err := os.ReadFile(baselinePath)
		assert.NilError(t, err)
		var saved map[string]map[string]map[string]json.RawMessage
		assert.NilError(t, json.Unmarshal(encoded, &saved))
		assert.Equal(t, string(saved["a.js"]["x"]["count"]), "1")
		assert.Equal(t, string(saved["a.js"]["x"]["reason"]), step.reason)
		assert.Equal(t, string(saved["b.js"]["x"]["ticket"]), "9007199254740993")
	}
}

func TestRecordCreatesMissingParentDirectoryAndRejectsInvalidParent(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	assert.NilError(t, os.WriteFile(filepath.Join(cwd, "a.js"), nil, 0o644))

	for _, location := range []string{".cache/suppressions/", "config/sub/suppressions.json"} {
		f, err := Open(cwd, Options{Mode: Record, Location: location})
		assert.NilError(t, err)
		_, _, err = reportError(t.Context(), f, filepath.Join(cwd, "a.js"), "no-debugger")
		assert.NilError(t, err)
		reopened, err := Open(cwd, Options{Location: location})
		assert.NilError(t, err)
		assert.Assert(t, reopened != nil)
		assert.Equal(t, reopened.entries["a.js"]["no-debugger"].Count, 1)
	}

	for _, invalid := range []string{"a.js/base.json", "a.js/sub/"} {
		if _, err := Open(cwd, Options{Mode: Record, Location: invalid}); err == nil {
			t.Fatalf("Open(%q) unexpectedly succeeded for non-directory parent", invalid)
		}
	}

	if runtime.GOOS != "windows" {
		assert.NilError(t, os.Symlink("missing-target", filepath.Join(cwd, "dangling-dir")))
		if _, err := Open(cwd, Options{Mode: Record, Location: "dangling-dir/base.json"}); err == nil {
			t.Fatal("Open unexpectedly succeeded for dangling parent directory symlink")
		}
		readonlyDir := filepath.Join(cwd, "readonly")
		assert.NilError(t, os.Mkdir(readonlyDir, 0o555))
		t.Cleanup(func() { _ = os.Chmod(readonlyDir, 0o755) })
		if err := probeDirectoryWritable(readonlyDir); errors.Is(err, os.ErrPermission) {
			for _, loc := range []string{"readonly/base.json", "readonly/missing/base.json"} {
				if _, err := Open(cwd, Options{Mode: Record, Location: loc}); !errors.Is(err, os.ErrPermission) {
					t.Fatalf("Open(%q) = %v, want permission error", loc, err)
				}
			}
		}
	}
}

func TestOpenAndRecordIgnoreTempDirState(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	baselinePath := filepath.Join(cwd, DefaultFilename)
	missingTmp := filepath.Join(cwd, "no-such-tmp")
	readonlyTmp := filepath.Join(cwd, "readonly-tmp")
	assert.NilError(t, os.Mkdir(readonlyTmp, 0o555))
	t.Cleanup(func() { _ = os.Chmod(readonlyTmp, 0o755) })

	for _, tmp := range []string{missingTmp, readonlyTmp} {
		t.Setenv("TMPDIR", tmp)
		t.Setenv("TEMP", tmp)
		t.Setenv("TMP", tmp)

		assert.NilError(t, os.WriteFile(baselinePath, []byte("{oops"), 0o644))
		if _, err := Open(cwd, Options{}); err == nil {
			t.Fatalf("TMPDIR=%q: invalid baseline was silently skipped", tmp)
		}

		assert.NilError(t, os.WriteFile(baselinePath, []byte(`{"a.js":{"no-var":{"count":1}}}`), 0o644))
		f, err := Open(cwd, Options{})
		assert.NilError(t, err)
		assert.Assert(t, f != nil)
		assert.Equal(t, f.entries["a.js"]["no-var"].Count, 1)

		rec, err := Open(cwd, Options{Mode: Record})
		assert.NilError(t, err)
		_, _, err = reportError(t.Context(), rec, filepath.Join(cwd, "b.js"), "no-debugger")
		assert.NilError(t, err)
	}
}

func TestResolveLocationPathWindowsDriveRelative(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows drive-relative paths apply only on Windows")
	}
	cwd := tempWorkingDirectory(t)
	vol := filepath.VolumeName(cwd)
	assert.Assert(t, vol != "")
	assert.Equal(t, resolveLocationPath(cwd, vol+"base.json"), filepath.Join(cwd, "base.json"))
	assert.Equal(t, resolveLocationPath(cwd, strings.ToLower(vol)+"sub/base.json"), filepath.Join(cwd, "sub", "base.json"))
	assert.Equal(t, resolveLocationPath(cwd, "/root-base.json"), filepath.Clean(vol+"/root-base.json"))

	f, err := Open(cwd, Options{Mode: Record, Location: vol + "drive-dir/"})
	assert.NilError(t, err)
	assert.Equal(t, f.path, filepath.Join(cwd, "drive-dir", "rslint-suppressions_"+cwdHash(cwd)))
}

func TestConcurrentProcessUpdatesHelper(t *testing.T) {
	if os.Getenv("RSLINT_SUPPRESSIONS_CONCURRENT_WORKER") == "" {
		return
	}
	cwd := os.Getenv("RSLINT_SUPPRESSIONS_CONCURRENT_CWD")
	location := os.Getenv("RSLINT_SUPPRESSIONS_CONCURRENT_LOCATION")
	file := os.Getenv("RSLINT_SUPPRESSIONS_CONCURRENT_FILE")
	f, err := Open(cwd, Options{Mode: Record, Location: location})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, _, err := reportError(t.Context(), f, filepath.Join(cwd, file), "no-debugger"); err != nil {
		t.Fatalf("apply: %v", err)
	}
}

func TestConcurrentProcessUpdatesPreserveAllEntries(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	seed := make(entries, 2000)
	for i := range 2000 {
		seed[fmt.Sprintf("existing/file-%04d.js", i)] = map[string]entry{
			"no-debugger": {Count: 1},
		}
	}
	encoded, err := json.MarshalIndent(seed, "", "  ")
	assert.NilError(t, err)
	assert.NilError(t, os.WriteFile(filepath.Join(cwd, DefaultFilename), encoded, 0o644))

	exe, err := os.Executable()
	assert.NilError(t, err)
	const workers = 4
	var wg sync.WaitGroup
	errs := make([]error, workers)
	readerDone := make(chan struct{})
	var readerErr error
	var readerWG sync.WaitGroup
	readerWG.Add(1)
	go func() {
		defer readerWG.Done()
		for {
			select {
			case <-readerDone:
				return
			default:
				if _, openErr := Open(cwd, Options{}); openErr != nil {
					readerErr = openErr
					return
				}
			}
		}
	}()
	for i := range workers {
		workerTmp := t.TempDir()
		wg.Add(1)
		go func(idx int, tmpDir string) {
			defer wg.Done()
			cmd := exec.Command(exe, "-test.run=^TestConcurrentProcessUpdatesHelper$")
			cmd.Env = append(os.Environ(),
				"TMPDIR="+tmpDir,
				"TEMP="+tmpDir,
				"TMP="+tmpDir,
				"RSLINT_SUPPRESSIONS_CONCURRENT_WORKER=1",
				"RSLINT_SUPPRESSIONS_CONCURRENT_CWD="+cwd,
				fmt.Sprintf("RSLINT_SUPPRESSIONS_CONCURRENT_FILE=worker-%d.js", idx),
			)
			if out, runErr := cmd.CombinedOutput(); runErr != nil {
				errs[idx] = fmt.Errorf("worker %d failed: %w\n%s", idx, runErr, out)
			}
		}(i, workerTmp)
	}
	wg.Wait()
	close(readerDone)
	readerWG.Wait()
	assert.NilError(t, readerErr)
	for _, runErr := range errs {
		assert.NilError(t, runErr)
	}

	final, err := Open(cwd, Options{})
	assert.NilError(t, err)
	assert.Assert(t, final != nil)
	assert.Equal(t, len(final.entries), 2000+workers)
	for i := range workers {
		key := fmt.Sprintf("worker-%d.js", i)
		assert.Equal(t, final.entries[key]["no-debugger"].Count, 1)
	}
	dirEntries, err := os.ReadDir(cwd)
	assert.NilError(t, err)
	assert.Equal(t, len(dirEntries), 1)
	assert.Equal(t, dirEntries[0].Name(), DefaultFilename)

	// Concurrent first-time creation into a missing nested custom directory.
	const nestedLocation = ".cache/concurrent/"
	var nestedWG sync.WaitGroup
	nestedErrs := make([]error, workers)
	for i := range workers {
		nestedWG.Add(1)
		go func(idx int) {
			defer nestedWG.Done()
			f, openErr := Open(cwd, Options{Mode: Record, Location: nestedLocation})
			if openErr != nil {
				nestedErrs[idx] = openErr
				return
			}
			_, _, nestedErrs[idx] = reportError(t.Context(), f, filepath.Join(cwd, fmt.Sprintf("first-%d.js", idx)), "no-debugger")
		}(i)
	}
	nestedWG.Wait()
	for _, runErr := range nestedErrs {
		assert.NilError(t, runErr)
	}
	nested, err := Open(cwd, Options{Location: nestedLocation})
	assert.NilError(t, err)
	assert.Assert(t, nested != nil)
	assert.Equal(t, len(nested.entries), workers)
	for i := range workers {
		key := fmt.Sprintf("first-%d.js", i)
		assert.Equal(t, nested.entries[key]["no-debugger"].Count, 1)
	}
	nestedEntries, err := os.ReadDir(filepath.Join(cwd, ".cache", "concurrent"))
	assert.NilError(t, err)
	assert.Equal(t, len(nestedEntries), 1)
}

func TestConcurrentProcessUpdatesEquivalentUnicodeNames(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	// cspell:disable-next-line
	const nfc = "r\u00e9sum\u00e9.json"
	// cspell:disable-next-line
	const nfd = "re\u0301sume\u0301.json"

	seed := make(entries, 2000)
	for i := range 2000 {
		seed[fmt.Sprintf("existing/file-%04d.js", i)] = map[string]entry{
			"no-debugger": {Count: 1},
		}
	}
	encoded, err := json.MarshalIndent(seed, "", "  ")
	assert.NilError(t, err)
	assert.NilError(t, os.WriteFile(filepath.Join(cwd, nfc), encoded, 0o644))

	nfcInfo, err := os.Stat(filepath.Join(cwd, nfc))
	assert.NilError(t, err)
	nfdInfo, err := os.Stat(filepath.Join(cwd, nfd))
	if err != nil || !os.SameFile(nfcInfo, nfdInfo) {
		t.Skip("filesystem does not equate NFC and NFD filename spellings")
	}

	exe, err := os.Executable()
	assert.NilError(t, err)
	const workers = 4
	spellings := []string{nfc, nfd, nfc, nfd}
	var wg sync.WaitGroup
	errs := make([]error, workers)
	for i := range workers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cmd := exec.Command(exe, "-test.run=^TestConcurrentProcessUpdatesHelper$")
			cmd.Env = append(os.Environ(),
				"RSLINT_SUPPRESSIONS_CONCURRENT_WORKER=1",
				"RSLINT_SUPPRESSIONS_CONCURRENT_CWD="+cwd,
				"RSLINT_SUPPRESSIONS_CONCURRENT_LOCATION="+spellings[idx],
				fmt.Sprintf("RSLINT_SUPPRESSIONS_CONCURRENT_FILE=unicode-%d.js", idx),
			)
			if out, runErr := cmd.CombinedOutput(); runErr != nil {
				errs[idx] = fmt.Errorf("worker %d (%q) failed: %w\n%s", idx, spellings[idx], runErr, out)
			}
		}(i)
	}
	wg.Wait()
	for _, runErr := range errs {
		assert.NilError(t, runErr)
	}

	final, err := Open(cwd, Options{Location: nfc})
	assert.NilError(t, err)
	assert.Assert(t, final != nil)
	assert.Equal(t, len(final.entries), 2000+workers)
	for i := range workers {
		key := fmt.Sprintf("unicode-%d.js", i)
		assert.Equal(t, final.entries[key]["no-debugger"].Count, 1)
	}
	dirEntries, err := os.ReadDir(cwd)
	assert.NilError(t, err)
	assert.Equal(t, len(dirEntries), 1)
}

func TestLockDoesNotTouchCallerNamedLockFile(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	const lockNamed = ".rslint-suppressions.lock"

	// Fresh custom baseline literally named .rslint-suppressions.lock.
	fresh, err := Open(cwd, Options{Mode: Record, Location: lockNamed})
	assert.NilError(t, err)
	_, _, err = reportError(t.Context(), fresh, filepath.Join(cwd, "a.js"), "no-debugger")
	assert.NilError(t, err)

	// Updating an existing baseline of that name preserves old and new entries.
	existing, err := Open(cwd, Options{Mode: Record, Location: lockNamed})
	assert.NilError(t, err)
	_, _, err = reportError(t.Context(), existing, filepath.Join(cwd, "b.js"), "no-debugger")
	assert.NilError(t, err)

	// Recording to the default baseline in the same directory preserves the
	// separate .rslint-suppressions.lock file untouched.
	def, err := Open(cwd, Options{Mode: Record})
	assert.NilError(t, err)
	_, _, err = reportError(t.Context(), def, filepath.Join(cwd, "c.js"), "no-debugger")
	assert.NilError(t, err)

	reopenedCustom, err := Open(cwd, Options{Location: lockNamed})
	assert.NilError(t, err)
	assert.Assert(t, reopenedCustom != nil)
	assert.DeepEqual(t, reopenedCustom.entries, entries{
		"a.js": {"no-debugger": {Count: 1}},
		"b.js": {"no-debugger": {Count: 1}},
	})

	reopenedDefault, err := Open(cwd, Options{})
	assert.NilError(t, err)
	assert.Assert(t, reopenedDefault != nil)
	assert.DeepEqual(t, reopenedDefault.entries, entries{
		"c.js": {"no-debugger": {Count: 1}},
	})
}

func TestPublishNewBaselineExclusive(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	target := filepath.Join(cwd, "first-publish.json")

	const writers = 16
	start := make(chan struct{})
	var winners atomic.Int32
	var winnerContent atomic.Value
	errs := make([]error, writers)
	var wg sync.WaitGroup

	for i := range writers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tempName := filepath.Join(cwd, fmt.Sprintf(".publish-%02d.tmp", idx))
			content := fmt.Sprintf(`{"w-%02d.js":{"no-debugger":{"count":1}}}`, idx)
			if err := os.WriteFile(tempName, []byte(content), 0o666); err != nil {
				errs[idx] = err
				return
			}
			<-start
			err := publishNewBaseline(t.Context(), tempName, target)
			switch {
			case err == nil:
				winners.Add(1)
				winnerContent.Store(content)
			case errors.Is(err, fs.ErrExist):
				_ = os.Remove(tempName)
			default:
				_ = os.Remove(tempName)
				errs[idx] = fmt.Errorf("writer %d: %w", idx, err)
			}
		}(i)
	}
	close(start)
	wg.Wait()

	for _, err := range errs {
		assert.NilError(t, err)
	}
	assert.Equal(t, winners.Load(), int32(1))
	got, err := os.ReadFile(target)
	assert.NilError(t, err)
	assert.Equal(t, string(got), winnerContent.Load().(string))
}
