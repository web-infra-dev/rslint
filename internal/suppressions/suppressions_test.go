package suppressions

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestScopedUpdate(t *testing.T) {
	cwd := t.TempDir()
	checked := filepath.ToSlash(filepath.Join(cwd, "a.ts"))
	for _, name := range []string{"a.ts", "untouched.ts"} {
		if err := os.WriteFile(filepath.Join(cwd, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(cwd, DefaultFilename)
	initial := `{"a.ts":{"x":{"count":3},"y":{"count":1}},"untouched.ts":{"x":{"count":1}},"gone.ts":{"x":{"count":1}}}`
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
		visible, stale, err := file.Apply(t.Context(), []string{checked}, append([]rule.RuleDiagnostic(nil), diagnostics...))
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
	want := entries{"a.ts": {"x": {Count: 1}}, "untouched.ts": {"x": {Count: 1}}}
	if !reflect.DeepEqual(pruned.entries, want) {
		t.Fatalf("prune: got %v, want %v", pruned.entries, want)
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
		_, _, err := baseline.Apply(ctx, []string{file}, []rule.RuleDiagnostic{
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
