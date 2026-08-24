package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
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
	utf8BOM := writeTestFile(t, dir, "utf8-bom.ts", []byte("\uFEFFlet a = 1;\n"))
	clean := writeTestFile(t, dir, "clean.ts", []byte("let a = 1;\n"))
	empty := writeTestFile(t, dir, "empty.ts", nil)
	utf16LE := writeTestFile(t, dir, "utf16-le.ts", []byte{0xFF, 0xFE, 'a', 0x00})
	utf16BE := writeTestFile(t, dir, "utf16-be.ts", []byte{0xFE, 0xFF, 0x00, 'a'})
	missing := tspath.NormalizePath(filepath.Join(dir, "missing.ts"))

	fs := osvfs.FS()
	for _, tt := range []struct {
		name string
		path string
		want bool
	}{
		{"utf-8 mark", utf8BOM, true},
		{"no mark", clean, false},
		{"empty file", empty, false},
		{"utf-16 little endian mark", utf16LE, true},
		{"utf-16 big endian mark", utf16BE, true},
		{"unreadable path", missing, false},
	} {
		if got := SourceHasBOM(fs, tt.path); got != tt.want {
			t.Errorf("SourceHasBOM(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

// TestSourceHasBOMFromOverlay locks in that caller-supplied content answers for
// itself. Consulting the file behind an overlay would report a mark on a buffer
// that does not have one, and miss a mark on one that does.
func TestSourceHasBOMFromOverlay(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	markedOnDisk := writeTestFile(t, dir, "marked.ts", []byte("\uFEFFlet a = 1;\n"))
	cleanOnDisk := writeTestFile(t, dir, "clean.ts", []byte("let a = 1;\n"))

	fs := NewOverlayVFS(osvfs.FS(), map[string]string{
		markedOnDisk: "let a = 1;\n",
		cleanOnDisk:  "\uFEFFlet a = 1;\n",
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
}
