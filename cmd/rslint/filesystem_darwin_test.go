// cspell:ignore APFS FSOPT FULLPATH Linkat STRASSE Straße

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
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

func TestCLIDarwinDirectoryEntryName(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeLintTargetContractFiles(t, root, map[string]string{"Straße/file": "", "Café/file": "", "Other/file": ""})
	path := func(name string) string { return filepath.Join(root, name) }
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
		if got := darwinDirectoryEntryName(input); got != want {
			t.Errorf("DirectoryEntryName(%q) = %q, want %q", input, got, want)
		}
	}
	check("Café", "")
	check(path("Café/file"), "")
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
	for _, input := range []string{path("Straße/WORK"), path("Other/work")} {
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
