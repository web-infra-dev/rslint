package suppressions

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"testing"

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
	t.Run("directory location follows the Unix working directory", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows retains the process working directory spelling")
		}
		for _, pair := range [][2]string{{"Project", "PROJECT"}, {"Caf\u00e9", "Cafe\u0301"}} {
			names := []string{filepath.Join(parent, pair[0]), filepath.Join(parent, pair[1])}
			if err := os.Mkdir(names[0], 0o755); err != nil {
				t.Fatal(err)
			}
			err := os.Mkdir(names[1], 0o755)
			same := os.IsExist(err)
			if err != nil && !same {
				t.Fatal(err)
			}
			var filenames []string
			for _, name := range names {
				t.Chdir(name)
				file, err := Open(name, Options{Mode: Record, Location: parent})
				if err != nil {
					t.Fatal(err)
				}
				filenames = append(filenames, file.path)
				if key := file.fileKey(filepath.Join(name, "file.ts")); key != "file.ts" {
					t.Errorf("file key from %q = %q", name, key)
				}
			}
			if (filenames[0] == filenames[1]) != same {
				t.Errorf("directory baselines %q, same directory: %v", filenames, same)
			}
		}
	})
	repo := filepath.Join(parent, "projects", "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	stable := filepath.Join(parent, "stable")
	if err := os.Symlink(repo, stable); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	alias := filepath.Join(parent, "link")
	if err := os.Symlink("stable", alias); err != nil {
		t.Fatal(err)
	}
	t.Chdir(alias)
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
	shared, err := Open(alias, Options{Mode: Record, Location: "../shared.json"})
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(filepath.Dir(hashDirectory), "shared.json"); shared.path != want {
		t.Fatalf("relative location = %q, want %q", shared.path, want)
	}
	for _, dir := range []string{repo, alias} {
		if key := f.fileKey(filepath.ToSlash(filepath.Join(dir, "a.ts"))); key != "a.ts" {
			t.Errorf("file key under %s = %q, want a.ts", dir, key)
		}
	}
	t.Run("rules can report different counts for each filename", func(t *testing.T) {
		first := filepath.ToSlash(filepath.Join(repo, "name.js"))
		second := filepath.ToSlash(filepath.Join(alias, "name.js"))
		if err := os.WriteFile(first, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if f.operandKey(second) == f.fileKey(second) {
			first, second = second, first
		}
		firstKey, secondKey := f.operandKey(first), f.operandKey(second)
		selected := []FileResult{{Path: first}, {Path: second, InputPaths: []string{first}}}
		a := rule.RuleDiagnostic{FilePath: first, RuleName: "x", Severity: rule.SeverityError}
		b := rule.RuleDiagnostic{FilePath: second, RuleName: "x", Severity: rule.SeverityError}
		path := filepath.Join(parent, "counts.json")
		run := func(mode Mode, diagnostics ...rule.RuleDiagnostic) entries {
			t.Helper()
			store, err := Open(alias, Options{Mode: mode, Location: path})
			if err != nil {
				t.Fatal(err)
			}
			visible, unused, err := store.Apply(t.Context(), selected, diagnostics)
			if len(visible) != 0 || unused || err != nil {
				t.Fatalf("mode %d: visible %v, unused %v, error %v", mode, visible, unused, err)
			}
			return store.entries
		}
		saved := run(Record, a, b, b)
		if saved[firstKey]["x"].Count != 1 || saved[secondKey]["x"].Count != 2 {
			t.Fatalf("separate filenames: %v", saved)
		}
		selected[1].Unchecked = true
		saved = run(Prune)
		if len(saved) != 1 || saved[secondKey]["x"].Count != 2 {
			t.Fatalf("unchecked filename: %v", saved)
		}
		selected[0].Unchecked, selected[1].Unchecked = true, false
		selected[1].InputPaths = []string{first, second}
		run(Check, b, b)
	})
	sibling := filepath.Join(filepath.Dir(repo), "sibling.ts")
	if err := os.WriteFile(sibling, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	key := f.fileKey(filepath.ToSlash(sibling))
	want, err := filepath.Rel(hashDirectory, sibling)
	if err != nil || key != filepath.ToSlash(want) {
		t.Fatalf("sibling key = %q, want %q (error %v)", key, want, err)
	}
	data := entries{key: {"x": {Count: 1}}}
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
	t.Chdir(filepath.Join(lexical, "a"))
	f, err = Open(filepath.Join(lexical, "a"), Options{Mode: Record})
	if err != nil {
		t.Fatal(err)
	}
	data = make(entries)
	bases := []string{lexical, filepath.Join(stable, "packages", "view"), filepath.Join(packages, "view"), inner}
	upper := filepath.Join(parent, "INNER")
	if _, err := os.Stat(upper); err == nil {
		bases = append(bases, upper)
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, base := range bases {
		for _, name := range []string{"b", "linked"} {
			key := f.fileKey(filepath.Join(base, name, "file.ts"))
			if want := "../" + name + "/file.ts"; key != want {
				t.Errorf("sibling key under %s = %q, want %q", base, key, want)
			}
			data[key] = map[string]entry{"x": {Count: 1}}
		}
	}
	t.Run("repeated key keeps the single-input fallback", func(t *testing.T) {
		caller := filepath.Join(lexical, "b", "file.ts")
		operand := filepath.Join(stable, "packages", "view", "b", "file.ts")
		selected := FileResult{Path: caller, InputPaths: []string{operand}}
		duplicate := FileResult{Path: operand, InputPaths: []string{operand}, CoalescedInto: caller, Unchecked: true}
		file := f.matchFiles([]FileResult{selected, duplicate})[caller]
		if !slices.Equal(file.keys, []string{f.operandKey(operand)}) || file.portable != "../b/file.ts" {
			t.Fatalf("repeated operand: exact keys %v, fallback %q", file.keys, file.portable)
		}
	})
	t.Run("operand keeps the spelling of a link outside cwd", func(t *testing.T) {
		relative := filepath.Join(inner, "b", "file.ts")
		lower := filepath.Join(lexical, "b", "file.ts")
		upper := filepath.Join(lexical, "b", "FILE.ts")
		if !SamePathSpelling(lower, upper) {
			t.Skip("volume distinguishes these filenames")
		}
		files := f.matchFiles([]FileResult{
			{Path: upper, InputPaths: []string{relative, upper}},
			{Path: lower, InputPaths: []string{relative}, CoalescedInto: upper, Unchecked: true},
		})
		want := []string{f.operandKey(relative), f.operandKey(upper)}
		slices.Sort(want)
		if got := files[upper].keys; !slices.Equal(got, want) {
			t.Fatalf("selected link keys = %v, want %v", got, want)
		}
	})
	t.Run("case-only spelling of an ancestor link", func(t *testing.T) {
		alternate := filepath.Join(parent, "LINK")
		path := filepath.Join(alternate, "packages", "view", "b", "file.ts")
		want := "../b/file.ts"
		if err := os.Symlink(repo, alternate); err == nil {
			// A case-sensitive volume permits a second link with the same target.
			up, err := filepath.Rel(f.baseDir, parent)
			if err != nil {
				t.Fatal(err)
			}
			want = filepath.ToSlash(filepath.Join(up, "LINK", "packages", "view", "b", "file.ts"))
		} else if !os.IsExist(err) {
			t.Fatal(err)
		}
		if got := f.fileKey(path); got != want {
			t.Errorf("case-aware sibling key = %q, want %q", got, want)
		}
	})
	t.Chdir(filepath.Join(inner, "a"))
	if err := os.Remove(alias); err != nil {
		t.Fatal(err)
	}
	f, err = Open(filepath.Join(inner, "a"), Options{Mode: Prune})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.prune(data, nil); err != nil || len(data) != 2 {
		t.Fatalf("pruning after removing the working directory alias: entries %v, error %v", data, err)
	}
}

func TestDistinctLintResultsCannotWriteTheSameExactKey(t *testing.T) {
	const exact = "../../foo/file.js"
	direct := &checkedFile{literal: exact, portable: exact, keys: []string{exact}}
	alias := &checkedFile{literal: "../../deep/foo/file.js", portable: "../../FOO/file.js", keys: []string{exact}}
	separateFilenames(map[string][]*checkedFile{direct.portable: {direct}, alias.portable: {alias}})
	selectExistingKeys(nil, map[string]*checkedFile{direct.literal: direct, alias.literal: alias})
	if !slices.Equal(direct.keys, []string{exact}) || !slices.Equal(alias.keys, []string{"../../FOO/file.js"}) {
		t.Fatalf("keys for two measured filenames: %v and %v", direct.keys, alias.keys)
	}
}

func TestUpdatePreservesOtherWriter(t *testing.T) {
	parent := t.TempDir()
	cwd := filepath.Join(parent, "repo")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(parent, "link")
	if runtime.GOOS == "windows" {
		// cspell:ignore mklink
		if output, err := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", link, cwd).CombinedOutput(); err != nil {
			t.Fatalf("create directory junction: %v\n%s", err, output)
		}
	} else if err := os.Symlink(cwd, link); err != nil {
		t.Fatal(err)
	}
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
