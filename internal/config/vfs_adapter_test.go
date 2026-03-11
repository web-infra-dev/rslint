package config

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

func TestVfsAdapter_OpenRoot(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "a.txt"))

	adapter := &vfsAdapter{vfs: osvfs.FS(), root: tmpDir}
	f, err := adapter.Open(".")
	assert.NilError(t, err)
	defer f.Close()

	dirFile, ok := f.(fs.ReadDirFile)
	assert.Assert(t, ok, "root should be a ReadDirFile")

	entries, err := dirFile.ReadDir(-1)
	assert.NilError(t, err)
	assert.Assert(t, len(entries) >= 1)

	found := false
	for _, e := range entries {
		if e.Name() == "a.txt" {
			found = true
			assert.Assert(t, !e.IsDir())
		}
	}
	assert.Assert(t, found, "should find a.txt in root")
}

func TestVfsAdapter_OpenDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "sub/file.json"))

	adapter := &vfsAdapter{vfs: osvfs.FS(), root: tmpDir}
	f, err := adapter.Open("sub")
	assert.NilError(t, err)
	defer f.Close()

	info, err := f.Stat()
	assert.NilError(t, err)
	assert.Assert(t, info.IsDir())
}

func TestVfsAdapter_OpenFile(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "file.json"))

	adapter := &vfsAdapter{vfs: osvfs.FS(), root: tmpDir}
	f, err := adapter.Open("file.json")
	assert.NilError(t, err)
	defer f.Close()

	info, err := f.Stat()
	assert.NilError(t, err)
	assert.Assert(t, !info.IsDir())
}

func TestVfsAdapter_OpenNotExist(t *testing.T) {
	tmpDir := t.TempDir()

	adapter := &vfsAdapter{vfs: osvfs.FS(), root: tmpDir}
	_, err := adapter.Open("nonexistent")
	assert.Assert(t, err != nil)

	var pathErr *fs.PathError
	assert.Assert(t, errors.As(err, &pathErr))
	assert.Assert(t, errors.Is(pathErr.Err, fs.ErrNotExist))
}

func TestVfsDirFile_ReadDirAll(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "a.txt"))
	createTestFile(t, filepath.Join(tmpDir, "b.txt"))
	assert.NilError(t, os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0o755))

	adapter := &vfsAdapter{vfs: osvfs.FS(), root: tmpDir}
	f, err := adapter.Open(".")
	assert.NilError(t, err)
	defer f.Close()

	dirFile := f.(fs.ReadDirFile)
	entries, err := dirFile.ReadDir(-1)
	assert.NilError(t, err)
	assert.Equal(t, len(entries), 3)

	// Should be sorted
	assert.Equal(t, entries[0].Name(), "a.txt")
	assert.Assert(t, !entries[0].IsDir())
	assert.Equal(t, entries[1].Name(), "b.txt")
	assert.Assert(t, !entries[1].IsDir())
	assert.Equal(t, entries[2].Name(), "subdir")
	assert.Assert(t, entries[2].IsDir())

	// Second call should return nil, nil (exhausted)
	entries2, err := dirFile.ReadDir(-1)
	assert.NilError(t, err)
	assert.Assert(t, entries2 == nil)
}

func TestVfsDirFile_ReadDirPaginated(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "a.txt"))
	createTestFile(t, filepath.Join(tmpDir, "b.txt"))
	createTestFile(t, filepath.Join(tmpDir, "c.txt"))

	adapter := &vfsAdapter{vfs: osvfs.FS(), root: tmpDir}
	f, err := adapter.Open(".")
	assert.NilError(t, err)
	defer f.Close()

	dirFile := f.(fs.ReadDirFile)

	// Read 2 entries
	entries, err := dirFile.ReadDir(2)
	assert.NilError(t, err)
	assert.Equal(t, len(entries), 2)
	assert.Equal(t, entries[0].Name(), "a.txt")
	assert.Equal(t, entries[1].Name(), "b.txt")

	// Read remaining — should get 1 entry and io.EOF
	entries, err = dirFile.ReadDir(2)
	assert.Assert(t, errors.Is(err, io.EOF))
	assert.Equal(t, len(entries), 1)
	assert.Equal(t, entries[0].Name(), "c.txt")

	// Read again — should get 0 entries and io.EOF
	entries, err = dirFile.ReadDir(1)
	assert.Assert(t, errors.Is(err, io.EOF))
	assert.Assert(t, len(entries) == 0)
}

func TestVfsAdapter_SymlinkCycleFiltered(t *testing.T) {
	tmpDir := t.TempDir()
	dirA := filepath.Join(tmpDir, "a")
	assert.NilError(t, os.MkdirAll(dirA, 0o755))
	// a/loop -> tmpDir creates a cycle: a/loop/a/loop/...
	assert.NilError(t, os.Symlink(tmpDir, filepath.Join(dirA, "loop")))

	adapter := &vfsAdapter{vfs: osvfs.FS(), root: tmpDir}

	// Open "a" and read its entries
	f, err := adapter.Open("a")
	assert.NilError(t, err)
	defer f.Close()

	dirFile := f.(fs.ReadDirFile)
	entries, err := dirFile.ReadDir(-1)
	assert.NilError(t, err)

	// "loop" symlink target is tmpDir, which hasn't been visited yet → included
	hasLoop := false
	for _, e := range entries {
		if e.Name() == "loop" {
			hasLoop = true
		}
	}
	assert.Assert(t, hasLoop, "first encounter of symlink should be included")

	// Now open "a/loop" (which resolves to tmpDir) and read its entries
	f2, err := adapter.Open("a/loop")
	assert.NilError(t, err)
	defer f2.Close()

	dirFile2 := f2.(fs.ReadDirFile)
	entries2, err := dirFile2.ReadDir(-1)
	assert.NilError(t, err)

	// Inside a/loop (=tmpDir), there's "a" directory.
	// Inside "a", the symlink "loop" points to tmpDir again.
	// When ReadDir on a/loop lists "a", it's a regular dir (not a symlink) → included.
	// But when we later ReadDir on a/loop/a, "loop" symlink target (tmpDir) is
	// already in visitedSymTargets → filtered out, breaking the cycle.
	hasA := false
	for _, e := range entries2 {
		if e.Name() == "a" {
			hasA = true
		}
	}
	assert.Assert(t, hasA, "should see 'a' directory inside the symlink target")

	// Open a/loop/a and verify "loop" is now filtered
	f3, err := adapter.Open("a/loop/a")
	assert.NilError(t, err)
	defer f3.Close()

	dirFile3 := f3.(fs.ReadDirFile)
	entries3, err := dirFile3.ReadDir(-1)
	assert.NilError(t, err)

	for _, e := range entries3 {
		if e.Name() == "loop" {
			t.Fatal("symlink cycle should have been filtered out")
		}
	}
}

func TestVfsDirEntry_TypeAndInfo(t *testing.T) {
	entry := &vfsDirEntry{name: "test", isDir: true}
	assert.Equal(t, entry.Name(), "test")
	assert.Assert(t, entry.IsDir())
	assert.Equal(t, entry.Type(), fs.ModeDir)

	info, err := entry.Info()
	assert.NilError(t, err)
	assert.Assert(t, info.IsDir())
	assert.Equal(t, info.Name(), "test")

	fileEntry := &vfsDirEntry{name: "file.txt", isDir: false}
	assert.Assert(t, !fileEntry.IsDir())
	assert.Equal(t, fileEntry.Type(), fs.FileMode(0))
}
