package config

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
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

// Open() always returns vfsDirFile (the adapter is only used by fs.WalkDir
// which only opens directories). Verify this contract.
func TestVfsAdapter_OpenAlwaysReturnsDir(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "file.json"))

	adapter := &vfsAdapter{vfs: osvfs.FS(), root: tmpDir}

	// Open a file path — still returns vfsDirFile (no DirectoryExists check).
	f, err := adapter.Open("file.json")
	assert.NilError(t, err)
	defer f.Close()

	info, err := f.Stat()
	assert.NilError(t, err)
	// This returns isDir=true because Open always returns vfsDirFile.
	// This is correct for the fs.WalkDir use case where Open is never called for files.
	assert.Assert(t, info.IsDir(), "Open always returns vfsDirFile for fs.WalkDir usage")
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

// Default vfsAdapter (followSymlinks=false) skips symlinks entirely. This is
// what DiscoverGapFiles relies on for deterministic concurrent traversal.
func TestVfsAdapter_SymlinksSkippedByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	dirA := filepath.Join(tmpDir, "a")
	assert.NilError(t, os.MkdirAll(dirA, 0o755))
	// a/link -> tmpDir/a (any symlinked dir; cycle isn't required for this test)
	assert.NilError(t, os.Symlink(dirA, filepath.Join(dirA, "link")))
	createTestFile(t, filepath.Join(dirA, "real.ts"))

	adapter := &vfsAdapter{vfs: osvfs.FS(), root: tmpDir}

	f, err := adapter.Open("a")
	assert.NilError(t, err)
	defer f.Close()

	entries, err := f.(fs.ReadDirFile).ReadDir(-1)
	assert.NilError(t, err)

	hasLink := false
	hasReal := false
	for _, e := range entries {
		if e.Name() == "link" {
			hasLink = true
		}
		if e.Name() == "real.ts" {
			hasReal = true
		}
	}
	assert.Assert(t, !hasLink, "symlinks must be skipped when followSymlinks=false")
	assert.Assert(t, hasReal, "regular files must still be returned")
}

// Opt-in vfsAdapter (followSymlinks=true) follows symlinks but skips cycles.
// This is what loader.expandProjectGlob uses (single-threaded).
func TestVfsAdapter_SymlinkCycleFilteredWhenFollowing(t *testing.T) {
	tmpDir := t.TempDir()
	dirA := filepath.Join(tmpDir, "a")
	assert.NilError(t, os.MkdirAll(dirA, 0o755))
	// a/loop -> tmpDir creates a cycle: a/loop/a/loop/...
	assert.NilError(t, os.Symlink(tmpDir, filepath.Join(dirA, "loop")))

	adapter := &vfsAdapter{vfs: osvfs.FS(), root: tmpDir, followSymlinks: true}

	// The target is already an ancestor of "a", so entering the symlink would
	// create a cycle immediately.
	f, err := adapter.Open("a")
	assert.NilError(t, err)
	defer f.Close()

	entries, err := f.(fs.ReadDirFile).ReadDir(-1)
	assert.NilError(t, err)

	hasLoop := false
	for _, e := range entries {
		if e.Name() == "loop" {
			hasLoop = true
		}
	}
	assert.Assert(t, !hasLoop, "symlink target already in the ancestor chain must be skipped")
}

type vfsAdapterMockFS struct {
	vfs.FS
	entries         map[string]vfs.Entries
	realpaths       map[string]string
	realpathCalls   []string
	caseSensitiveFS bool
}

func (m *vfsAdapterMockFS) GetAccessibleEntries(path string) vfs.Entries {
	return m.entries[path]
}

func (m *vfsAdapterMockFS) Realpath(path string) string {
	m.realpathCalls = append(m.realpathCalls, path)
	if realpath, ok := m.realpaths[path]; ok {
		return realpath
	}
	return path
}

func (m *vfsAdapterMockFS) UseCaseSensitiveFileNames() bool {
	return m.caseSensitiveFS
}

func TestVfsAdapter_UsesSymlinkMetadataBeforeRealpath(t *testing.T) {
	for _, followSymlinks := range []bool{false, true} {
		t.Run(map[bool]string{false: "skip", true: "follow"}[followSymlinks], func(t *testing.T) {
			mock := &vfsAdapterMockFS{
				FS: osvfs.FS(),
				entries: map[string]vfs.Entries{
					"/repo": {
						Directories: []string{"link", "regular"},
						Symlinks:    map[string]struct{}{"link": {}},
					},
				},
				realpaths: map[string]string{"/repo/link": "/target"},
			}
			adapter := &vfsAdapter{vfs: mock, root: "/repo", followSymlinks: followSymlinks}

			file, err := adapter.Open(".")
			assert.NilError(t, err)
			entries, err := file.(fs.ReadDirFile).ReadDir(-1)
			assert.NilError(t, err)

			names := make([]string, 0, len(entries))
			for _, entry := range entries {
				names = append(names, entry.Name())
			}
			if followSymlinks {
				assert.DeepEqual(t, names, []string{"link", "regular"})
				assert.DeepEqual(t, mock.realpathCalls, []string{"/repo/link", "/repo"})
			} else {
				assert.DeepEqual(t, names, []string{"regular"})
				assert.Equal(t, len(mock.realpathCalls), 0)
			}
		})
	}
}

func TestVfsAdapter_FollowSymlinksKeepsDistinctAliasesToSameTarget(t *testing.T) {
	mock := &vfsAdapterMockFS{
		FS: osvfs.FS(),
		entries: map[string]vfs.Entries{
			"/repo": {
				Directories: []string{"alias-a", "alias-b"},
				Symlinks: map[string]struct{}{
					"alias-a": {},
					"alias-b": {},
				},
			},
		},
		realpaths: map[string]string{
			"/repo":         "/repo",
			"/repo/alias-a": "/shared",
			"/repo/alias-b": "/shared",
		},
	}
	adapter := &vfsAdapter{vfs: mock, root: "/repo", followSymlinks: true}

	file, err := adapter.Open(".")
	assert.NilError(t, err)
	entries, err := file.(fs.ReadDirFile).ReadDir(-1)
	assert.NilError(t, err)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	assert.DeepEqual(t, names, []string{"alias-a", "alias-b"})
}

func TestVfsAdapter_RealpathFallbackIsCached(t *testing.T) {
	mock := &vfsAdapterMockFS{
		FS: osvfs.FS(),
		entries: map[string]vfs.Entries{
			"/repo":       {Directories: []string{"child"}},
			"/repo/child": {Directories: []string{"grandchild"}},
		},
	}
	adapter := &vfsAdapter{vfs: mock, root: "/repo"}

	root, err := adapter.Open(".")
	assert.NilError(t, err)
	_, err = root.(fs.ReadDirFile).ReadDir(-1)
	assert.NilError(t, err)

	child, err := adapter.Open("child")
	assert.NilError(t, err)
	_, err = child.(fs.ReadDirFile).ReadDir(-1)
	assert.NilError(t, err)

	assert.DeepEqual(t, mock.realpathCalls, []string{
		"/repo",
		"/repo/child",
		"/repo/child/grandchild",
	})
}

func TestVfsAdapter_RealpathFallbackJoinsFilesystemRoots(t *testing.T) {
	tests := []struct {
		name      string
		root      string
		childPath string
	}{
		{name: "posix", root: "/", childPath: "/child"},
		{name: "drive", root: "C:/", childPath: "C:/child"},
		{name: "unc", root: "//server/share/", childPath: "//server/share/child"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mock := &vfsAdapterMockFS{
				FS: osvfs.FS(),
				entries: map[string]vfs.Entries{
					test.root: {Directories: []string{"child"}},
				},
			}
			adapter := &vfsAdapter{vfs: mock, root: test.root}

			file, err := adapter.Open(".")
			assert.NilError(t, err)
			entries, err := file.(fs.ReadDirFile).ReadDir(-1)
			assert.NilError(t, err)
			assert.Equal(t, len(entries), 1)
			assert.Equal(t, entries[0].Name(), "child")
			assert.DeepEqual(t, mock.realpathCalls, []string{test.root, test.childPath})
		})
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

// spyVFS wraps a real VFS and counts DirectoryExists calls.
type spyVFS struct {
	vfs.FS
	directoryExistsCalls int
}

func (s *spyVFS) DirectoryExists(path string) bool {
	s.directoryExistsCalls++
	return s.FS.DirectoryExists(path)
}

// Open() no longer calls DirectoryExists (always returns vfsDirFile).
// Verify with a spy VFS that DirectoryExists is never called.
func TestVfsAdapter_OpenNeverCallsDirectoryExists(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "sub/file.txt"))

	spy := &spyVFS{FS: osvfs.FS()}
	adapter := &vfsAdapter{vfs: spy, root: tmpDir}

	// Open root
	_, err := adapter.Open(".")
	assert.NilError(t, err)
	// Open subdirectory
	_, err = adapter.Open("sub")
	assert.NilError(t, err)

	assert.Equal(t, spy.directoryExistsCalls, 0, "Open should never call DirectoryExists")
}
