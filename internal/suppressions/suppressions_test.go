package suppressions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestScopedUpdate(t *testing.T) {
	cwd := t.TempDir()
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
		visible, stale, err := file.Apply(t.Context(), []FileResult{{Path: checked}}, append([]rule.RuleDiagnostic(nil), diagnostics...))
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

func TestWorkingDirectoryAlias(t *testing.T) {
	parent, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(parent, "projects", "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(parent, "link")
	if err := os.Symlink(repo, alias); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	f, err := Open(alias, Options{Mode: Record, Location: alias})
	if err != nil {
		t.Fatal(err)
	}
	hashDirectory := repo
	if runtime.GOOS == "windows" {
		hashDirectory = alias
	}
	if got, want := filepath.Base(f.path), "rslint-suppressions_"+cwdHash(hashDirectory); got != want {
		t.Fatalf("directory baseline name = %q, want %q", got, want)
	}
	for _, dir := range []string{repo, alias} {
		if key := f.fileKey(filepath.ToSlash(filepath.Join(dir, "a.ts"))); key != "a.ts" {
			t.Errorf("file key under %s = %q, want a.ts", dir, key)
		}
	}
	sibling := filepath.Join(filepath.Dir(repo), "sibling.ts")
	if err := os.WriteFile(sibling, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if key := f.fileKey(filepath.ToSlash(sibling)); key != "../sibling.ts" {
		t.Fatalf("sibling key = %q, want ../sibling.ts", key)
	}
	data := entries{"../sibling.ts": {"x": {Count: 1}}}
	if err := f.prune(data, nil); err != nil || len(data) != 1 {
		t.Fatalf("pruning an existing sibling: entries %v, error %v", data, err)
	}

	packages := filepath.Join(repo, "packages")
	if err := os.MkdirAll(packages, 0o755); err != nil {
		t.Fatal(err)
	}
	inner := filepath.Join(parent, "inner")
	for _, name := range []string{"a", "b"} {
		if err := os.MkdirAll(filepath.Join(inner, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(inner, filepath.Join(packages, "view")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(sibling, filepath.Join(inner, "b", "file.ts")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("b", filepath.Join(inner, "linked")); err != nil {
		t.Fatal(err)
	}
	lexical := filepath.Join(alias, "packages", "view")
	f, err = Open(filepath.Join(lexical, "a"), Options{Mode: Record})
	if err != nil {
		t.Fatal(err)
	}
	data = make(entries)
	for _, base := range []string{lexical, filepath.Join(packages, "view"), inner} {
		for _, name := range []string{"b", "linked"} {
			key := f.fileKey(filepath.Join(base, name, "file.ts"))
			if want := "../" + name + "/file.ts"; key != want {
				t.Errorf("sibling key under %s = %q, want %q", base, key, want)
			}
			data[key] = map[string]entry{"x": {Count: 1}}
		}
	}
	t.Run("case-only spelling of an ancestor link", func(t *testing.T) {
		alternate := filepath.Join(parent, "LINK")
		path := filepath.Join(alternate, "packages", "view", "b", "file.ts")
		// Windows filepath.Rel folds link and LINK even when they can name
		// two independent links to the same directory. Supply that result on
		// every host so the fixture exercises the volume's actual behavior.
		folded := filepath.Join("b", "file.ts")
		want := folded
		if err := os.Symlink(repo, alternate); err == nil {
			want = filepath.Join("..", "..", "..", "LINK", "packages", "view", "b", "file.ts")
		} else if !os.IsExist(err) {
			t.Fatal(err)
		}
		if got := verifyRelativePathCase(lexical, path, folded); got != want {
			t.Errorf("case-aware relative path = %q, want %q", got, want)
		}
	})
	if err := os.Remove(alias); err != nil {
		t.Fatal(err)
	}
	if err := f.prune(data, nil); err != nil || len(data) != 2 {
		t.Fatalf("pruning after removing the working directory alias: entries %v, error %v", data, err)
	}
}

func TestUpdatePreservesOtherWriter(t *testing.T) {
	cwd := t.TempDir()
	options := Options{Mode: Record}
	first, err := Open(cwd, options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Open(cwd, options)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.ToSlash(filepath.Join(cwd, "test.js"))
	update := func(ctx context.Context, baseline *File, id string) error {
		_, _, err := baseline.Apply(ctx, []FileResult{{Path: file}}, []rule.RuleDiagnostic{
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
