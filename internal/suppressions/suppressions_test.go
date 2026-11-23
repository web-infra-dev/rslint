package suppressions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestScopedUpdate(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	checked := filepath.ToSlash(filepath.Join(cwd, "a.ts"))
	for _, name := range []string{"a.ts", "untouched.ts", "former-dir"} {
		if err := os.WriteFile(filepath.Join(cwd, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(cwd, DefaultFilename)
	driveKey := "D:/shared/file.ts"
	if runtime.GOOS == "windows" {
		driveKey = filepath.ToSlash(filepath.Join(cwd, "missing-on-mounted-drive.ts"))
	}
	initial := fmt.Sprintf(`{
		"a.ts":{"x":{"count":3},"y":{"count":1},"typed":{"count":2}},
		"untouched.ts":{"x":{"count":1}},
		"gone.ts":{"x":{"count":1}},
		"former-dir/old.ts":{"x":{"count":1}},
		"^/old.ts":{"x":{"count":1}},
		%q:{"x":{"count":1}},
		"//server/share/file.ts":{"x":{"count":1}}
	}`, driveKey)
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	diagnostics := []rule.RuleDiagnostic{
		{FilePath: checked, RuleName: "x", Severity: rule.SeverityError},
		{FilePath: checked, RuleName: "z", Severity: rule.SeverityWarning},
	}
	run := func(options Options, unused bool) *File {
		t.Helper()
		file, err := Open(cwd, options)
		if err != nil {
			t.Fatal(err)
		}
		visible, stale, err := file.Apply(t.Context(), []FileResult{{Path: checked, SkippedRules: []string{"typed"}}}, append([]rule.RuleDiagnostic(nil), diagnostics...))
		if err != nil || stale != unused || len(visible) != 1 || visible[0].RuleName != "z" {
			t.Fatalf("apply: %v, stale %v, error %v", visible, stale, err)
		}
		return file
	}
	run(Options{}, true)
	recorded := run(Options{Mode: Record, Rules: []string{"x"}}, true)
	if got := recorded.entries["a.ts"]; got["x"].Count != 1 || got["y"].Count != 1 {
		t.Fatalf("recorded counts: %v", got)
	}
	pruned := run(Options{Mode: Prune}, false)
	want := entries{
		"a.ts":                   {"x": {Count: 1}, "typed": {Count: 2}},
		"untouched.ts":           {"x": {Count: 1}},
		"//server/share/file.ts": {"x": {Count: 1}},
	}
	if runtime.GOOS != "windows" {
		want[driveKey] = map[string]entry{"x": {Count: 1}}
	}
	if !reflect.DeepEqual(pruned.entries, want) {
		t.Fatalf("prune: got %v, want %v", pruned.entries, want)
	}
}

func TestMissingBaseline(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	if file, err := Open(cwd, Options{}); err != nil || file != nil {
		t.Fatalf("plain check: file=%v, err=%v", file, err)
	}
	if _, err := Open(cwd, Options{Location: "missing.json"}); err == nil {
		t.Fatal("explicit missing baseline was accepted")
	}
	f, err := Open(cwd, Options{Mode: Prune})
	if err != nil {
		t.Fatal(err)
	}
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
	root := filepath.FromSlash(osvfs.FS().Realpath(filepath.ToSlash(t.TempDir())))
	repo, link := filepath.Join(root, "project", "repo"), filepath.Join(root, "link")
	src := filepath.Join(repo, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	destination := repo
	// On case-insensitive volumes the process cwd also restores the stored case.
	alternate := filepath.Join(filepath.Dir(repo), "REPO")
	if _, err := os.Stat(alternate); err == nil && runtime.GOOS != "windows" {
		destination = alternate
	}
	makeDirectoryLink(t, destination, link)
	makeDirectoryLink(t, src, filepath.Join(repo, "other"))
	run := func(cwd, name string, mode Mode, wantVisible int) *File {
		t.Helper()
		t.Chdir(cwd)
		f, err := Open(cwd, Options{Mode: mode})
		if err != nil {
			t.Fatal(err)
		}
		visible, unused, err := reportError(t.Context(), f, name, "x")
		if err != nil || unused || len(visible) != wantVisible {
			t.Fatalf("linting %s from %s: visible=%v unused=%v err=%v", name, cwd, visible, unused, err)
		}
		return f
	}
	run(link, filepath.Join(link, "src", "app.js"), Record, 0)
	run(repo, filepath.Join(src, "app.js"), Check, 0)
	run(repo, filepath.Join(repo, "other", "app.js"), Check, 1)
	run(repo, filepath.Join(repo, "other", "app.js"), Record, 0)
	f := run(link, filepath.Join(filepath.Dir(repo), "shared.js"), Record, 0)
	want := entries{"src/app.js": {"x": {Count: 1}}, "other/app.js": {"x": {Count: 1}}, "../shared.js": {"x": {Count: 1}}}
	if !reflect.DeepEqual(f.entries, want) {
		t.Fatalf("recorded %v, want %v", f.entries, want)
	}
	subdir := filepath.Join(link, "src")
	run(subdir, filepath.Join(repo, "peer.js"), Record, 0)
	run(subdir, filepath.Join(link, "peer.js"), Check, 1)

	native := repo
	if runtime.GOOS == "windows" {
		native = link
	}
	t.Chdir(link)
	for location, want := range map[string]string{
		root:             filepath.Join(root, "rslint-suppressions_"+cwdHash(native)),
		"../shared.json": filepath.Join(filepath.Dir(native), "shared.json"),
	} {
		f, err := Open(link, Options{Mode: Record, Location: location})
		if err != nil || f.path != want {
			t.Fatalf("location %q: file=%v err=%v, want %q", location, f, err, want)
		}
	}
}

func TestFileKeyPreservesCase(t *testing.T) {
	check := func(cwd, path, want string) {
		t.Helper()
		f := File{cwd: cwd}
		if got := f.fileKey(path); got != want {
			t.Errorf("fileKey(%q, %q) = %q, want %q", cwd, path, got, want)
		}
	}
	check("/tree/pkg/", "/tree/Pkg/File.js", "../Pkg/File.js")
	check("/tree/pkg/", "/tree/pkg-extra/File.js", "../pkg-extra/File.js")
	check("/", "/File.js", "File.js")
	if runtime.GOOS == "windows" {
		check("C:/tree/pkg/", "c:/tree/Pkg/File.js", "../Pkg/File.js")
		check("C:/tree/", "d:/Shared/File.js", "D:/Shared/File.js")
		check("//server/share/pkg/", "//server/share/Pkg/File.js", "../Pkg/File.js")
		check("//server/share/pkg/", "//SERVER/SHARE/Pkg/File.js", "//SERVER/SHARE/Pkg/File.js")
		check("//server/share/pkg/", "//server/other/File.js", "//server/other/File.js")
	}
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
	cwd := t.TempDir()
	t.Chdir(cwd)
	var writers [2]*File
	for i := range writers {
		f, err := Open(cwd, Options{Mode: Record})
		if err != nil {
			t.Fatal(err)
		}
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
	for i, err := range errs {
		if errors.Is(err, os.ErrExist) {
			err = write(i) // Retry after the overlapping writer completes.
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := reportError(ctx, writers[0], name, "cancelled"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled write returned %v", err)
	}
	f, err := Open(cwd, Options{})
	want := entries{"test.js": {"rule0": {Count: 1}, "rule1": {Count: 1}}}
	if err != nil || f == nil || !reflect.DeepEqual(f.entries, want) {
		t.Fatalf("concurrent updates: file=%v err=%v, want %v", f, err, want)
	}
}

func TestUpdateThroughFileSymlink(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	path := filepath.Join(cwd, DefaultFilename)
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(cwd, "linked.json")
	if err := os.Symlink(DefaultFilename, link); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}
	f, err := Open(cwd, Options{Mode: Record, Location: link})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := reportError(t.Context(), f, filepath.Join(cwd, "test.js"), "x"); err != nil {
		t.Fatal(err)
	}
	real, err := Open(cwd, Options{})
	if err != nil || real == nil || real.entries["test.js"]["x"].Count != 1 {
		t.Fatalf("write via symlink: file=%v err=%v", real, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := reportError(t.Context(), f, filepath.Join(cwd, "test.js"), "y"); err == nil {
		t.Fatal("replaced a dangling baseline link")
	}
	if name, err := os.Readlink(link); err != nil || name != DefaultFilename {
		t.Fatalf("stored symlink = %q, err=%v", name, err)
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
