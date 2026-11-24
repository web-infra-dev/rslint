package suppressions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"gotest.tools/v3/assert"
)

func tempWorkingDirectory(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	assert.NilError(t, err)
	t.Chdir(dir)
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
		file, err := Open(options)
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

func TestMissingBaseline(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	if file, err := Open(Options{}); err != nil || file != nil {
		t.Fatalf("plain check: file=%v, err=%v", file, err)
	}
	if _, err := Open(Options{Location: "missing.json"}); err == nil {
		t.Fatal("explicit missing baseline was accepted")
	}
	f, err := Open(Options{Mode: Prune})
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
	// On case-insensitive volumes the process cwd also restores the stored case.
	alternate := filepath.Join(filepath.Dir(repo), "REPO")
	if _, err := os.Stat(alternate); err == nil && runtime.GOOS != "windows" {
		destination = alternate
	}
	makeDirectoryLink(t, destination, link)
	native := repo
	if runtime.GOOS == "windows" {
		native = link
	}
	t.Chdir(link)
	for location, want := range map[string]string{
		root:             filepath.Join(root, "rslint-suppressions_"+cwdHash(native)),
		"../shared.json": filepath.Join(filepath.Dir(link), "shared.json"),
	} {
		f, err := Open(Options{Mode: Record, Location: location})
		assert.NilError(t, err)
		assert.Equal(t, f.path, want)
		for name, want := range map[string]string{
			filepath.Join(src, "app.js"):         "src/app.js",
			filepath.Join(link, "src", "app.js"): "../../link/src/app.js",
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

func TestConcurrentUpdates(t *testing.T) {
	const holdLock = "RSLINT_TEST_HOLD_SUPPRESSIONS_LOCK"
	if os.Getenv(holdLock) == "1" {
		f, err := Open(Options{Mode: Record})
		assert.NilError(t, err)
		assert.NilError(t, f.update(t.Context(), func(entries) bool {
			fmt.Fprintln(os.Stdout, "locked")
			var reply string
			_, _ = fmt.Fscan(os.Stdin, &reply)
			return false
		}))
		return
	}
	cwd := tempWorkingDirectory(t)
	assert.NilError(t, os.WriteFile(DefaultFilename+".tmp", []byte("unrelated"), 0o600))
	var writers [2]*File
	for i := range writers {
		f, err := Open(Options{Mode: Record})
		assert.NilError(t, err)
		writers[i] = f
	}
	name := filepath.Join(cwd, "test.js")
	write := func(i int) error {
		_, _, err := reportError(t.Context(), writers[i], name, fmt.Sprintf("rule%d", i))
		return err
	}
	var group sync.WaitGroup
	var errs [2]error
	for i := range writers {
		group.Go(func() { errs[i] = write(i) })
	}
	group.Wait()
	contended := func(err error) bool {
		// Windows may report access denied while an earlier handle is closing.
		return errors.Is(err, os.ErrExist) || runtime.GOOS == "windows" && errors.Is(err, os.ErrPermission)
	}
	for i, err := range errs {
		if contended(err) {
			err = write(i)
		}
		assert.NilError(t, err)
	}
	if runtime.GOOS != "js" && runtime.GOOS != "wasip1" {
		program, err := os.Executable()
		assert.NilError(t, err)
		holder := exec.CommandContext(t.Context(), program, "-test.run=^TestConcurrentUpdates$")
		holder.Env = append(os.Environ(), holdLock+"=1")
		input, err := holder.StdinPipe()
		assert.NilError(t, err)
		defer input.Close()
		output, err := holder.StdoutPipe()
		assert.NilError(t, err)
		assert.NilError(t, holder.Start())
		var ready string
		_, err = fmt.Fscan(output, &ready)
		assert.NilError(t, err)
		assert.Equal(t, ready, "locked")
		_, _, err = reportError(t.Context(), writers[0], name, "contended")
		assert.Assert(t, contended(err), "contended update returned %v", err)
		assert.ErrorContains(t, err, "retry")
		// A separate baseline with a maximum-length basename can update while the first is locked.
		longName := strings.Repeat("x", 250) + ".json"
		other, err := Open(Options{Mode: Record, Location: longName})
		assert.NilError(t, err)
		for _, id := range []string{"first", "second"} {
			_, _, err := reportError(t.Context(), other, name, id)
			assert.NilError(t, err)
		}
		assert.NilError(t, holder.Process.Kill())
		assert.Assert(t, holder.Wait() != nil)
	}
	// The kernel released the interrupted writer's lock.
	_, _, err := reportError(t.Context(), writers[0], name, "after-crash")
	assert.NilError(t, err)
	leftover, err := os.ReadFile(DefaultFilename + ".tmp")
	assert.NilError(t, err)
	assert.Equal(t, string(leftover), "unrelated")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := reportError(ctx, writers[0], name, "cancelled"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled write returned %v", err)
	}
	f, err := Open(Options{})
	want := entries{"test.js": {"rule0": {Count: 1}, "rule1": {Count: 1}, "after-crash": {Count: 1}}}
	assert.NilError(t, err)
	assert.Assert(t, f != nil)
	assert.DeepEqual(t, f.entries, want)
}

func TestUpdateRequiresBaselineWriteAccess(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	const original = `{"a.js":{"x":{"count":1}}}`
	assert.NilError(t, os.WriteFile(DefaultFilename, []byte(original), 0o444))
	t.Cleanup(func() { _ = os.Chmod(DefaultFilename, 0o600) })
	if file, err := os.OpenFile(DefaultFilename, os.O_WRONLY, 0); err == nil {
		_ = file.Close()
		t.Skip("current user can write a read-only file")
	} else if !errors.Is(err, os.ErrPermission) {
		t.Fatal(err)
	}
	f, err := Open(Options{Mode: Record})
	assert.NilError(t, err)
	for _, id := range []string{"x", "y"} { // An unchanged count and a new count.
		_, _, err := reportError(t.Context(), f, filepath.Join(cwd, "a.js"), id)
		if !errors.Is(err, os.ErrPermission) {
			t.Fatalf("record %s: expected a permission error, got %v", id, err)
		}
	}
	if content, err := os.ReadFile(DefaultFilename); err != nil || string(content) != original {
		t.Fatalf("read-only baseline: %q, %v", content, err)
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
	f, err := Open(Options{Mode: Record, Location: link})
	assert.NilError(t, err)
	if _, _, err := reportError(t.Context(), f, filepath.Join(cwd, "test.js"), "x"); err != nil {
		t.Fatal(err)
	}
	defaultFile, err := Open(Options{})
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
	assert.NilError(t, os.Mkdir("cache", 0o755))
	for _, name := range []string{location, "cache/base.json"} {
		assert.NilError(t, os.WriteFile(name, []byte("{}"), 0o644))
	}
	f, err := Open(Options{Mode: Record, Location: location})
	assert.NilError(t, err)
	if _, _, err := reportError(t.Context(), f, filepath.Join(cwd, "test.js"), "x"); err != nil {
		t.Fatal(err)
	}
	f, err = Open(Options{Location: location})
	if err != nil || f == nil || f.entries["test.js"]["x"].Count != 1 {
		t.Fatalf("recorded baseline: file=%v err=%v", f, err)
	}
	if content, err := os.ReadFile("cache/base.json"); err != nil || string(content) != "{}" {
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
		assert.NilError(t, os.WriteFile(name, nil, 0o644))
	}
	const initial = `{
		"a.js":{"x":{"count":2,"reason":"deferred"}},
		"b.js":{"x":{"count":1,"ticket":9007199254740993}}
	}`
	assert.NilError(t, os.WriteFile(DefaultFilename, []byte(initial), 0o644))
	for _, step := range []struct {
		mode   Mode
		reason string
	}{{Prune, `"deferred"`}, {Record, ""}} {
		file, err := Open(Options{Mode: step.mode})
		assert.NilError(t, err)
		_, _, err = reportError(t.Context(), file, filepath.Join(cwd, "a.js"), "x")
		assert.NilError(t, err)
		encoded, err := os.ReadFile(DefaultFilename)
		assert.NilError(t, err)
		var saved map[string]map[string]map[string]json.RawMessage
		assert.NilError(t, json.Unmarshal(encoded, &saved))
		assert.Equal(t, string(saved["a.js"]["x"]["count"]), "1")
		assert.Equal(t, string(saved["a.js"]["x"]["reason"]), step.reason)
		assert.Equal(t, string(saved["b.js"]["x"]["ticket"]), "9007199254740993")
	}
}
