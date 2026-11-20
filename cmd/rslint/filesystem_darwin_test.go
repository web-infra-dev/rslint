// cspell:ignore APFS FSOPT FULLPATH Linkat STRASSE Straße Straß

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/suppressions"
	"golang.org/x/sys/unix"
)

func TestWorkingDirectoryWithHardLinksToSymlink(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(root, "project")
	work, other := filepath.Join(root, "work"), filepath.Join(root, "other")
	if err := os.Mkdir(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(project, work); err != nil {
		t.Fatal(err)
	}
	if err := unix.Linkat(unix.AT_FDCWD, work, unix.AT_FDCWD, other, 0); err != nil {
		t.Fatal(err)
	}
	upper := filepath.Join(root, "WORK")
	wantUpper := work
	if _, err := os.Lstat(upper); errors.Is(err, os.ErrNotExist) {
		// A case-sensitive volume can have a third, distinct directory entry.
		if err := unix.Linkat(unix.AT_FDCWD, work, unix.AT_FDCWD, upper, 0); err != nil {
			t.Fatal(err)
		}
		wantUpper = upper
	}
	t.Chdir(project)
	for input, want := range map[string]string{work: work, other: other, upper: wantUpper} {
		t.Setenv("PWD", input)
		got, err := getCLIWorkingDirectory()
		if err != nil || got != want {
			t.Errorf("cwd via %q = %q, %v, want %q", input, got, err, want)
		}
	}

	same, ok := newCLIFileSystem().(interface{ SameDirectoryEntry(left, right string) bool })
	if !ok {
		t.Fatal("CLI filesystem does not provide directory entry comparison")
	}
	linkName := func(path string) (alreadyExists bool) {
		t.Helper()
		err := unix.Linkat(unix.AT_FDCWD, work, unix.AT_FDCWD, path, 0)
		if err != nil && !errors.Is(err, os.ErrExist) {
			t.Fatal(err)
		}
		return err != nil
	}
	checks := make(map[[2]string]bool)
	for _, pair := range [][2]string{
		{"other", "OTHER"},
		{"Straße", "STRASSE"},
		{"Café", "Cafe\u0301"},
		{"\ufb03", "ffi"},
		{"\u13a0", "\uab70"},
	} {
		left, right := filepath.Join(root, pair[0]), filepath.Join(root, pair[1])
		linkName(left)
		checks[[2]string{left, right}] = linkName(right)
	}
	if err := os.Chmod(root, 0o111); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(root, 0o755); err != nil {
			t.Error(err)
		}
	})
	for pair, want := range checks {
		if got := same.SameDirectoryEntry(pair[0], pair[1]); got != want {
			t.Errorf("SameDirectoryEntry(%q, %q) in a search-only parent = %t, want %t", pair[0], pair[1], got, want)
		}
	}
	if same.SameDirectoryEntry(work, other) {
		t.Error("separate hard links in a search-only parent compare equal")
	}
	// Standard normalization and folding can collide for separate APFS entries.
	for _, pair := range [][2]string{
		{"a" + strings.Repeat("\u0345", 9), "a" + strings.Repeat("\u03b9", 9)},
		{"a" + strings.Repeat("\u0301", 31), "a" + strings.Repeat("\u0301", 30) + "\u034f\u0301"},
	} {
		if darwinEqualLookupNames(root, pair[0], pair[1]) {
			t.Errorf("hard link names %q and %q collide", pair[0], pair[1])
		}
	}
}

func TestCLIDarwinStoredEntryName(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeLintTargetContractFiles(t, root, map[string]string{"Straße/file": "", "Café/file": "", "Other/Straße.js": ""})
	path := func(name string) string { return filepath.Join(root, name) }
	if err := os.Symlink("Straße.js", path("Other/Café.js")); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(path("Straße/file"), path("Other/linked.js")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("Café", path("Shortcut")); err != nil {
		t.Fatal(err)
	}
	// The kernel can report Straße/WORK for a lookup of Other/work. Its
	// basename alone cannot establish the spelling in the requested parent.
	if err := os.Symlink("../Café", path("Straße/WORK")); err != nil {
		t.Fatal(err)
	}
	if err := unix.Linkat(unix.AT_FDCWD, path("Straße/WORK"), unix.AT_FDCWD, path("Other/work"), 0); err != nil {
		t.Fatal(err)
	}
	check := func(input, want string) {
		t.Helper()
		if got, _ := darwinStoredEntryName(input); got != want {
			t.Errorf("StoredEntryName(%q) = %q, want %q", input, got, want)
		}
	}
	check("Café", "")
	check(path("Café/file"), "file")
	check(path("missing"), "")
	check(path("Shortcut"), "Shortcut")
	check(path("Shortcut")+"/", "Shortcut")
	check(path("Straße/WORK"), "WORK")
	check(path("Other/work"), "work")
	if _, err := os.Stat(path("Other/WORK")); err == nil {
		check(path("Other/WORK"), "work")
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}

	// A searchable parent suffices for ordinary names. A hard link reported
	// in a different parent needs directory entries to disambiguate.
	for _, dir := range []string{root, path("Straße"), path("Other")} {
		if err := os.Chmod(dir, 0o111); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.Chmod(dir, 0o755); err != nil {
				t.Error(err)
			}
		})
	}
	check(path("Shortcut"), "Shortcut")
	for input, want := range map[string]string{"STRASSE": "Straße", "Cafe\u0301": "Café"} {
		if _, err := os.Stat(path(input)); err == nil {
			check(path(input), want)
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
	}
	fs := newCLIFileSystem()
	names := rslintconfig.NewInputPathNames(fs)
	for input, entry := range map[string]string{"Other/STRASSE.js": "Other/Straße.js", "Other/Cafe\u0301.js": "Other/Café.js"} {
		if _, err := os.Stat(path(input)); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			t.Fatal(err)
		}
		file := names.File(rslintconfig.FreezeFileIdentity(path(input), fs), false)
		if got := rslintconfig.GitPath(file); got != path(entry) || file.CanonicalPath != path("Other/Straße.js") {
			t.Errorf("observed %q as Git entry %q and physical target %q", input, got, file.CanonicalPath)
		}
	}
	for _, input := range []string{path("Straße/WORK"), path("Other/work"), path("Straße/file"), path("Other/linked.js")} {
		parent := filepath.Dir(input)
		if _, err := os.ReadDir(parent); !errors.Is(err, os.ErrPermission) {
			continue
		}
		reported, _ := darwinPathAttribute(input, unix.ATTR_CMN_FULLPATH, unix.FSOPT_NOFOLLOW)
		if reported != "" && filepath.Dir(reported) != parent {
			check(input, "")
		}
	}
}

func TestFixAndPrunePreserveUnlistedGitEntry(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const source = "var value = 1;\n"
	const ledger = `{"locked/STRASSE.js":{"no-var":{"count":1}}}`
	writeLintTargetContractFiles(t, root, map[string]string{
		".git": "", ".gitignore": "/locked/Straß[e/].js", "locked/placeholder": "",
		"else/file.js": source, "safe.js": source, suppressions.DefaultFilename: ledger,
	})
	path := func(name string) string { return tspath.ResolvePath(root, name) }
	if err := os.Symlink("../else/file.js", path("else/first.js")); err != nil {
		t.Fatal(err)
	}
	if err := unix.Linkat(unix.AT_FDCWD, path("else/first.js"), unix.AT_FDCWD, path("locked/Straße.js"), 0); err != nil {
		if errors.Is(err, unix.ENOTSUP) {
			t.Skip("filesystem does not support hard links to a symlink")
		}
		t.Fatal(err)
	}
	if err := os.Chmod(path("locked"), 0o111); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(path("locked"), 0o755); err != nil {
			t.Error(err)
		}
	})
	fsys := newCLIFileSystem()
	names := rslintconfig.NewInputPathNames(fsys)
	paths := []string{path("safe.js"), path("locked/STRASSE.js")}
	entries := rslintconfig.RslintConfig{{Files: []string{"**/*.js"}, Rules: rslintconfig.Rules{"no-var": "error"}}}
	catalog := explicitConfigCatalogForTest(root, entries)
	for _, p := range paths {
		catalog.TargetFiles = append(catalog.TargetFiles, names.File(rslintconfig.FreezeFileIdentity(p, fsys), false))
	}
	if _, err := darwinStoredEntryName(paths[1]); !errors.Is(err, os.ErrPermission) {
		t.Skip("this lookup does not produce an unlisted hard link")
	}
	if again := names.File(rslintconfig.FreezeFileIdentity(paths[1], fsys), false); again != catalog.TargetFiles[1] {
		t.Fatalf("repeated lookup changed file identity: %+v, then %+v", catalog.TargetFiles[1], again)
	}
	catalog.GitPolicy = rslintconfig.CollectGitPolicy(entries, root, fsys, catalog.TargetFiles, nil)
	code, _, stderr := runLintCommandForTest(t, root, lintArgs{
		FS: fsys, ConfigCatalog: catalog, Fix: true, Format: "jsonline", AllowFiles: paths,
		Suppressions: suppressions.Options{Mode: suppressions.Prune},
	})
	if code != 2 || !strings.Contains(stderr, "cannot be verified against .gitignore") {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	for name, want := range map[string]string{"safe.js": source, "else/file.js": source, suppressions.DefaultFilename: ledger} {
		if got, err := os.ReadFile(path(name)); err != nil || string(got) != want {
			t.Errorf("after --fix --prune-suppressions, %s = %q, %v, want %q", name, got, err, want)
		}
	}
}

func TestCLIDarwinRealpath(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeLintTargetContractFiles(t, root, map[string]string{"Src/Café.js": "", "Src/First.js": ""})
	path := func(name string) string { return tspath.ResolvePath(root, name) }
	if err := os.Symlink("Src", path("dir-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("dir-link/Café.js", path("entry.js")); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path("Src/Café.js"), 0); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(path("Src/First.js"), path("Second.js")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("Second.js", path("hard-link.js")); err != nil {
		t.Fatal(err)
	}
	fs := newCLIFileSystem()
	checks := map[string]string{
		"entry.js":     "Src/Café.js",
		"Src/First.js": "Src/First.js",
		"hard-link.js": "Second.js",
	}
	// Alternate case or normalization may be a distinct name on this volume.
	for name, want := range map[string]string{"src/cafe\u0301.JS": "Src/Café.js", "second.JS": "Second.js"} {
		if _, err := os.Stat(path(name)); err == nil {
			checks[name] = want
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
	}
	for name, want := range checks {
		if got := fs.Realpath(path(name)); got != path(want) {
			t.Errorf("Realpath(%q) = %q, want %q", name, got, path(want))
		}
	}
}
