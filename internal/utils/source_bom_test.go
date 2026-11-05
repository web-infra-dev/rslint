package utils

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// writeTestFile writes bytes to a file under dir and returns the path in the
// normalized form the VFS layers key on.
func writeTestFile(t *testing.T, dir string, name string, contents []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return tspath.NormalizePath(path)
}

// TestSourceHasBOMFromDisk covers the answer for text that came off disk. The
// file reader decodes the mark away before anyone sees the text, so the file's
// own bytes are the only witness left — including for UTF-16, which is decoded
// to UTF-8 and loses its mark the same way.
func TestSourceHasBOMFromDisk(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	utf8Marked := writeTestFile(t, dir, "utf8-bom.ts", []byte(BOM+"let a = 1;\n"))
	clean := writeTestFile(t, dir, "clean.ts", []byte("let a = 1;\n"))
	empty := writeTestFile(t, dir, "empty.ts", nil)
	short := writeTestFile(t, dir, "short.ts", []byte{0xFF})
	bomOnly := writeTestFile(t, dir, "bom-only.ts", []byte{0xFE, 0xFF})
	utf16LE := writeTestFile(t, dir, "utf16-le.ts", []byte{0xFF, 0xFE, 'a', 0x00})
	utf16BE := writeTestFile(t, dir, "utf16-be.ts", []byte{0xFE, 0xFF, 0x00, 'a'})
	missing := tspath.NormalizePath(filepath.Join(dir, "missing.ts"))

	fs := osvfs.FS()
	for _, tt := range []struct {
		name string
		path string
		want bool
	}{
		{"utf-8 mark", utf8Marked, true},
		{"no mark", clean, false},
		{"empty file", empty, false},
		{"one byte", short, false},
		{"two-byte UTF-16 mark", bomOnly, true},
		{"utf-16 little endian mark", utf16LE, true},
		{"utf-16 big endian mark", utf16BE, true},
		{"unreadable path", missing, false},
	} {
		if got, err := ReadSourceBOM(fs, tt.path); got != tt.want || err != nil {
			t.Errorf("ReadSourceBOM(%s) = (%v, %v), want (%v, nil)", tt.name, got, err, tt.want)
		}
	}
}

// TestSourceHasBOMFromOverlay locks in that caller-supplied content answers for
// itself. Consulting the file behind an overlay would report a mark on a buffer
// that does not have one, and miss a mark on one that does.
func TestSourceHasBOMFromOverlay(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	markedOnDisk := writeTestFile(t, dir, "marked.ts", []byte(BOM+"let a = 1;\n"))
	cleanOnDisk := writeTestFile(t, dir, "clean.ts", []byte("let a = 1;\n"))

	fs := NewOverlayVFS(osvfs.FS(), map[string]string{
		markedOnDisk: "let a = 1;\n",
		cleanOnDisk:  BOM + "let a = 1;\n",
	})

	if SourceHasBOM(fs, markedOnDisk) {
		t.Error("a clean overlay over a marked file reports a mark")
	}
	if !SourceHasBOM(fs, cleanOnDisk) {
		t.Error("a marked overlay over a clean file reports no mark")
	}

	// The text handed to the parser never carries the mark, whichever side it
	// came from, so offsets measured against it agree across both shapes.
	if text, _ := fs.ReadFile(cleanOnDisk); text != "let a = 1;\n" {
		t.Errorf("overlay ReadFile = %q, want the text without its mark", text)
	}
	if size := fs.Stat(cleanOnDisk).Size(); size != int64(len("let a = 1;\n")) {
		t.Errorf("overlay Stat size = %d, want the size without the mark", size)
	}

	t.Run("Windows realpath alias", func(t *testing.T) {
		t.Parallel()
		const alias = "C:/Link/source.ts"
		const canonical = "c:/Physical/source.ts"
		base := newMemoryReadFS(nil)
		base.realPaths[alias] = "C:/Physical/source.ts"
		files := map[string]string{canonical: BOM + "let value = 1;"}
		overlay := NewOverlayVFS(base, files)
		if !SourceHasBOM(overlay, alias) {
			t.Fatal("realpath drive spelling lost the overlay BOM")
		}
		if text, ok := overlay.ReadFile(alias); !ok || text != "let value = 1;" {
			t.Fatalf("realpath drive spelling lost the overlay text: %q, %v", text, ok)
		}
		files[canonical] = "let value = 20;"
		info := overlay.Stat(alias)
		if SourceHasBOM(overlay, alias) || info == nil || info.Size() != int64(len(files[canonical])) {
			t.Fatal("BOM and Stat did not observe the updated caller map")
		}
	})
}

func TestRestoreSourceBOM(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	markedPath := writeTestFile(t, dir, "marked.ts", []byte(BOM+"const value = 1;"))
	doubleMarkedPath := writeTestFile(t, dir, "double-marked.ts", []byte(BOM+BOM+"const value = 1;"))
	cleanPath := writeTestFile(t, dir, "clean.ts", []byte("const value = 1;"))
	fs := osvfs.FS()

	for _, tt := range []struct {
		name string
		path string
		text string
		want string
	}{
		{
			name: "restore stripped encoding mark",
			path: markedPath,
			text: "const value = 1;",
			want: BOM + "const value = 1;",
		},
		{
			name: "leave unmarked source unchanged",
			path: cleanPath,
			text: "const value = 1;",
			want: "const value = 1;",
		},
		{
			name: "preserve leading content mark separately",
			path: doubleMarkedPath,
			text: BOM + "const value = 1;",
			want: BOM + BOM + "const value = 1;",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got, err := RestoreSourceBOM(fs, tt.path, tt.text); got != tt.want || err != nil {
				t.Errorf("RestoreSourceBOM() = (%q, %v), want (%q, nil)", got, err, tt.want)
			}
		})
	}
}

func TestDiskBOMRetriesDescriptorExhaustion(t *testing.T) {
	t.Parallel()

	const path = "/source.ts"
	processLimit := syscall.EMFILE
	if runtime.GOOS == "windows" {
		processLimit = syscall.Errno(4) // ERROR_TOO_MANY_OPEN_FILES
	}
	for _, retryErr := range []error{nil, &os.PathError{Op: "open", Path: path, Err: syscall.ENFILE}} {
		attempts := 0
		has, err := readDiskBOM(path, func(path string) (bool, error) {
			attempts++
			if attempts == 1 {
				return false, &os.PathError{Op: "open", Path: path, Err: processLimit}
			}
			if attempts == 2 && retryErr != nil {
				return false, retryErr
			}
			return true, nil
		})
		if attempts != 2 || has != (retryErr == nil) || !errors.Is(err, retryErr) {
			t.Errorf("retry after descriptor exhaustion returned (%t, %v) in %d attempts, retry error %v", has, err, attempts, retryErr)
		}
	}
}

type failingBOMSource struct {
	vfs.FS
	err error
}

func (fs failingBOMSource) ReadSourceBOM(string) (bool, error) {
	return false, fs.err
}

func TestRestoreSourceBOMForwardsUnderlyingError(t *testing.T) {
	t.Parallel()

	dir := tspath.NormalizePath(t.TempDir())
	path := tspath.ResolvePath(dir, "source.ts")
	base := failingBOMSource{FS: osvfs.FS(), err: &os.PathError{Op: "open", Path: path, Err: syscall.EMFILE}}
	overlayPath := tspath.ResolvePath(dir, "overlay.ts")
	inner := NewOverlayVFS(base, map[string]string{overlayPath: BOM + "x;"})
	overlay := NewOverlayVFS(inner, nil)
	if text, err := RestoreSourceBOM(overlay, path, "x;"); text != "" || !errors.Is(err, syscall.EMFILE) {
		t.Errorf("unavailable BOM returned (%q, %v)", text, err)
	}
	if text, err := RestoreSourceBOM(overlay, overlayPath, "x;"); text != BOM+"x;" || err != nil {
		t.Errorf("overlay BOM returned (%q, %v)", text, err)
	}
}
