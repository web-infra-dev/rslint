package txtarfs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRejectsUnsafeOrAmbiguousNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		archive string
		want    string
	}{
		{name: "absolute", archive: "-- /outside --\nx\n", want: "relative path"},
		{name: "parent traversal", archive: "-- ../outside --\nx\n", want: "relative path"},
		{name: "backslash", archive: "-- dir\\outside --\nx\n", want: "forward slashes"},
		{name: "volume", archive: "-- C:outside --\nx\n", want: "filesystem volume"},
		{name: "Windows punctuation", archive: "-- dir/file?.txt --\nx\n", want: "forbidden in Windows"},
		{name: "control character", archive: "-- dir/file\x00.txt --\nx\n", want: "control characters"},
		{name: "trailing dot", archive: "-- dir/file. --\nx\n", want: "non-portable suffix"},
		{name: "reserved device", archive: "-- dir/CON.txt --\nx\n", want: "reserved Windows device"},
		{
			name:    "duplicate",
			archive: "-- file.txt --\nfirst\n-- file.txt --\nsecond\n",
			want:    "duplicate archived file",
		},
		{
			name:    "case folded duplicate",
			archive: "-- File.txt --\nfirst\n-- file.txt --\nsecond\n",
			want:    "case-insensitive filesystems",
		},
		{
			name:    "file directory collision",
			archive: "-- dir --\nfile\n-- dir/child.txt --\nchild\n",
			want:    "also a parent directory",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse("unsafe.txtar", []byte(test.archive))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestArchiveReadFileReturnsCopy(t *testing.T) {
	t.Parallel()

	archive, err := Parse("copy.txtar", []byte("-- file.txt --\ncontent\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	first, err := archive.ReadFile("file.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	first[0] = 'X'
	second, err := archive.ReadFile("file.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if got, want := string(second), "content\n"; got != want {
		t.Fatalf("second ReadFile() = %q, want %q", got, want)
	}
	if _, err := archive.ReadFile("missing.txt"); err == nil {
		t.Fatal("ReadFile accepted a missing file")
	}
}

func TestExtractSelectsSubtreeAndPreservesBytes(t *testing.T) {
	t.Parallel()

	archive, err := Parse("suite.txtar", []byte(
		"-- first/a.txt --\nalpha\n"+
			"-- first/nested/empty.txt --\n"+
			"-- first-other/leak.txt --\nleak\n"+
			"-- second/b.txt --\nbeta\n",
	))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	names, err := archive.FileNames("first")
	if err != nil {
		t.Fatalf("FileNames: %v", err)
	}
	if got, want := strings.Join(names, ","), "a.txt,nested/empty.txt"; got != want {
		t.Fatalf("FileNames() = %q, want %q", got, want)
	}

	directory := archive.Materialize(t, "first")
	data, err := os.ReadFile(filepath.Join(directory, "a.txt"))
	if err != nil {
		t.Fatalf("read materialized file: %v", err)
	}
	if got, want := string(data), "alpha\n"; got != want {
		t.Fatalf("materialized content = %q, want %q", got, want)
	}
	data, err = os.ReadFile(filepath.Join(directory, "nested", "empty.txt"))
	if err != nil {
		t.Fatalf("read empty materialized file: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("empty materialized file contains %q", data)
	}
	for _, name := range []string{"b.txt", "leak.txt"} {
		if _, err := os.Stat(filepath.Join(directory, name)); !os.IsNotExist(err) {
			t.Fatalf("unselected subtree file %q was materialized: %v", name, err)
		}
	}
}

func TestExtractRejectsMissingPrefixAndExistingFile(t *testing.T) {
	t.Parallel()

	archive, err := Parse("suite.txtar", []byte("-- suite/file.txt --\nnew\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if err := archive.extract(t.TempDir(), "missing"); err == nil || !strings.Contains(err.Error(), "selects no files") {
		t.Fatalf("missing prefix error = %v", err)
	}

	directory := t.TempDir()
	fileName := filepath.Join(directory, "file.txt")
	if err := os.WriteFile(fileName, []byte("original\n"), 0o644); err != nil {
		t.Fatalf("seed existing file: %v", err)
	}
	if err := archive.extract(directory, "suite"); err == nil {
		t.Fatal("extract overwrote an existing file")
	}
	data, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatalf("read existing file: %v", err)
	}
	if got, want := string(data), "original\n"; got != want {
		t.Fatalf("existing file changed to %q, want %q", got, want)
	}
}

func TestExtractCannotEscapeThroughSymlink(t *testing.T) {
	archive, err := Parse("suite.txtar", []byte("-- suite/link/escaped.txt --\nescaped\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	directory := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(directory, "link")); err != nil {
		t.Skipf("cannot create symlink on this platform: %v", err)
	}
	if err := archive.extract(directory, "suite"); err == nil {
		t.Fatal("extract followed a symlink outside its root")
	}
	if _, err := os.Stat(filepath.Join(outside, "escaped.txt")); !os.IsNotExist(err) {
		t.Fatalf("archive escaped extraction root: %v", err)
	}
}

func FuzzParseDoesNotPanic(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("-- safe/file.txt --\ncontent\n"))
	f.Add([]byte("-- ../escape --\ncontent\n"))
	f.Add([]byte("-- File.txt --\na\n-- file.txt --\nb\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			t.Skip()
		}
		archive, err := Parse("fuzz.txtar", data)
		if err != nil {
			return
		}
		names, namesErr := archive.FileNames("")
		if len(names) == 0 {
			if namesErr == nil {
				t.Fatal("empty archive unexpectedly returned a successful file selection")
			}
			return
		}
		if namesErr != nil {
			t.Fatalf("validated archive could not list its files: %v", namesErr)
		}
		for _, name := range names {
			if _, err := archive.ReadFile(name); err != nil {
				t.Fatalf("validated archive could not read %q: %v", name, err)
			}
		}
	})
}
