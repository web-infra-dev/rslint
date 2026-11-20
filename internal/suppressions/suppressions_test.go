package suppressions

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
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
		"a.ts":{"x":{"count":3},"y":{"count":1}},
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
		visible, stale, err := file.Apply(t.Context(), []FileResult{{Path: checked, CanonicalPath: osvfs.FS().Realpath(checked)}}, append([]rule.RuleDiagnostic(nil), diagnostics...))
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
		"a.ts":                   {"x": {Count: 1}},
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

func TestCanonicalFileResult(t *testing.T) {
	root := filepath.FromSlash(osvfs.FS().Realpath(filepath.ToSlash(t.TempDir())))
	repo, link := filepath.Join(root, "repo"), filepath.Join(root, "deep", "link")
	for _, dir := range []string{filepath.Join(repo, "app"), filepath.Join(repo, "src"), filepath.Dir(link)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	makeDirectoryLink(t, repo, link)
	cwd := filepath.Join(link, "app")
	t.Chdir(cwd)
	file := filepath.Join(repo, "src", "File.js")
	for _, name := range []string{file, filepath.Join(root, "shared.js")} {
		if err := os.WriteFile(name, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// An older baseline can contain a caller-visible spelling and an entry
	// for an unselected file outside the canonical cwd.
	const aliasKey = "../../deep/link/src/File.js"
	const siblingKey = "../../shared.js"
	path := filepath.Join(root, "baseline.json")
	raw := fmt.Sprintf(`{%q:{"x":{"count":2}},%q:{"x":{"count":1}}}`, aliasKey, siblingKey)
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	canonical := filepath.ToSlash(file)
	alias := filepath.ToSlash(filepath.Join(link, "src", "File.js"))
	run := func(mode Mode, display string, found, visibleCount int) *File {
		t.Helper()
		baseline, err := Open(cwd, Options{Mode: mode, Location: path})
		if err != nil {
			t.Fatal(err)
		}
		var diagnostics []rule.RuleDiagnostic
		for range found {
			diagnostics = append(diagnostics, rule.RuleDiagnostic{FilePath: display, RuleName: "x", Severity: rule.SeverityError})
		}
		visible, unused, err := baseline.Apply(t.Context(), []FileResult{{Path: display, CanonicalPath: canonical}}, diagnostics)
		if err != nil || unused || len(visible) != visibleCount {
			t.Fatalf("mode %d: visible %v, unused %v, error %v", mode, visible, unused, err)
		}
		return baseline
	}
	run(Check, alias, 1, 1)
	recorded := run(Record, alias, 1, 0)
	if got := recorded.entries["../src/File.js"]["x"].Count; got != 1 {
		t.Fatalf("canonical count = %d, want 1", got)
	}
	run(Check, canonical, 1, 0)
	pruned := run(Prune, alias, 0, 0)
	want := entries{aliasKey: {"x": {Count: 2}}, siblingKey: {"x": {Count: 1}}}
	if !reflect.DeepEqual(pruned.entries, want) {
		t.Fatalf("pruned entries = %v, want %v", pruned.entries, want)
	}

	// ESLint resolves the storage location and hashes the native process cwd.
	native := filepath.Join(repo, "app")
	if runtime.GOOS == "windows" {
		native = cwd
	}
	for location, expected := range map[string]string{
		root:             filepath.Join(root, "rslint-suppressions_"+cwdHash(native)),
		"../shared.json": filepath.Join(filepath.Dir(native), "shared.json"),
	} {
		store, err := Open(cwd, Options{Mode: Record, Location: location})
		if err != nil || store.path != expected {
			t.Fatalf("location %q: got %v, error %v, want %q", location, store, err, expected)
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

func TestUpdatePreservesOtherWriter(t *testing.T) {
	parent := filepath.FromSlash(osvfs.FS().Realpath(filepath.ToSlash(t.TempDir())))
	cwd := filepath.Join(parent, "repo")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(parent, "link")
	makeDirectoryLink(t, cwd, link)
	t.Chdir(link)
	options := Options{Mode: Record}
	first, err := Open(link, options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Open(link, options)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.ToSlash(filepath.Join(cwd, "test.js"))
	update := func(ctx context.Context, baseline *File, id string) error {
		_, _, err := baseline.Apply(ctx, []FileResult{{Path: file, CanonicalPath: file}}, []rule.RuleDiagnostic{
			{FilePath: file, RuleName: id, Severity: rule.SeverityError},
		})
		return err
	}
	if err := update(t.Context(), first, "first"); err != nil {
		t.Fatal(err)
	}
	if err := update(t.Context(), second, "second"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := update(ctx, first, "cancelled"); err == nil {
		t.Fatal("cancelled update succeeded")
	}
	reopened, err := Open(cwd, Options{})
	if err != nil {
		t.Fatal(err)
	}
	want := entries{"test.js": {"first": {Count: 1}, "second": {Count: 1}}}
	if !reflect.DeepEqual(reopened.entries, want) {
		t.Fatalf("got %v, want %v", reopened.entries, want)
	}
	t.Run("preserve the baseline file symlink", func(t *testing.T) {
		location := filepath.Join(link, "linked.json")
		if err := os.Symlink(DefaultFilename, location); err != nil {
			t.Skipf("file symlink unavailable: %v", err)
		}
		linked, err := Open(link, Options{Mode: Record, Location: location})
		if err != nil {
			t.Fatal(err)
		}
		if err := update(t.Context(), linked, "linked"); err != nil {
			t.Fatal(err)
		}
		reopened, err := Open(cwd, Options{})
		if err != nil || reopened == nil || reopened.entries["test.js"]["linked"].Count != 1 {
			t.Fatalf("update through file symlink: %v, error %v", reopened, err)
		}
		if err := os.Remove(filepath.Join(cwd, DefaultFilename)); err != nil {
			t.Fatal(err)
		}
		if err := update(t.Context(), linked, "dangling"); err == nil {
			t.Fatal("updated through dangling baseline symlink")
		}
		if target, err := os.Readlink(location); err != nil || target != DefaultFilename {
			t.Fatalf("baseline symlink target = %q, error %v", target, err)
		}
	})
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
