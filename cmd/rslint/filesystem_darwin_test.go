// cspell:ignore Linkat

package main

import (
	"errors"
	"os"
	"path/filepath"
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
