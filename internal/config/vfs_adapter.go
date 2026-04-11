package config

import (
	"io"
	"io/fs"
	"sort"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// vfsAdapter adapts a vfs.FS to a standard fs.FS rooted at a given directory,
// specifically for use with fs.WalkDir in DiscoverGapFiles. It is NOT a
// general-purpose fs.FS implementation — Open() always returns a directory
// handle (vfsDirFile) because fs.WalkDir only opens directories.
//
// It tracks visited symlink targets to prevent infinite recursion caused by
// symlink cycles, since the underlying VFS follows symlinks transparently
// and doublestar cannot detect them through this adapter.
type vfsAdapter struct {
	vfs               vfs.FS
	root              string
	visitedSymTargets map[string]struct{}
}

var _ fs.FS = (*vfsAdapter)(nil)

// Open implements fs.FS. This adapter is ONLY used by fs.WalkDir in
// DiscoverGapFiles, where Open() is only called on directories (via the
// internal readDir function). Therefore we always return a vfsDirFile
// without calling DirectoryExists — the parent's ReadDir already confirmed
// the entry is a directory, so the stat would be redundant.
func (a *vfsAdapter) Open(name string) (fs.File, error) {
	fullPath := a.fullPath(name)

	return &vfsDirFile{
		adapter: a,
		path:    fullPath,
		name:    name,
	}, nil
}

func (a *vfsAdapter) fullPath(name string) string {
	if name == "." {
		return a.root
	}
	return tspath.ResolvePath(a.root, name)
}

// vfsDirFile implements fs.ReadDirFile for directories.
type vfsDirFile struct {
	adapter *vfsAdapter
	path    string
	name    string
	entries []fs.DirEntry
	offset  int
}

var _ fs.ReadDirFile = (*vfsDirFile)(nil)

func (f *vfsDirFile) Stat() (fs.FileInfo, error) {
	return &vfsFileInfo{name: f.name, isDir: true}, nil
}

func (f *vfsDirFile) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: f.name, Err: fs.ErrInvalid}
}

func (f *vfsDirFile) Close() error { return nil }

func (f *vfsDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if f.entries == nil {
		accessible := f.adapter.vfs.GetAccessibleEntries(f.path)
		parentRealPath := f.adapter.vfs.Realpath(f.path)

		if f.adapter.visitedSymTargets == nil {
			f.adapter.visitedSymTargets = make(map[string]struct{})
		}

		f.entries = make([]fs.DirEntry, 0, len(accessible.Directories)+len(accessible.Files))

		for _, dir := range accessible.Directories {
			dirPath := tspath.ResolvePath(f.path, dir)
			dirRealPath := f.adapter.vfs.Realpath(dirPath)

			// A regular subdirectory's realpath equals parentRealPath + "/" + name.
			// If it differs, the entry is a symlink. Track symlink targets to
			// detect cycles: if the target was already visited, skip the entry.
			expectedRealPath := parentRealPath + "/" + dir
			if dirRealPath != expectedRealPath {
				if _, seen := f.adapter.visitedSymTargets[dirRealPath]; seen {
					continue
				}
				f.adapter.visitedSymTargets[dirRealPath] = struct{}{}
			}

			f.entries = append(f.entries, &vfsDirEntry{name: dir, isDir: true})
		}
		for _, file := range accessible.Files {
			f.entries = append(f.entries, &vfsDirEntry{name: file, isDir: false})
		}
		sort.Slice(f.entries, func(i, j int) bool {
			return f.entries[i].Name() < f.entries[j].Name()
		})
	}

	if n <= 0 {
		if f.offset >= len(f.entries) {
			return nil, nil
		}
		remaining := f.entries[f.offset:]
		f.offset = len(f.entries)
		return remaining, nil
	}

	if f.offset >= len(f.entries) {
		return nil, io.EOF
	}

	end := f.offset + n
	if end > len(f.entries) {
		end = len(f.entries)
	}
	result := f.entries[f.offset:end]
	f.offset = end
	if f.offset >= len(f.entries) {
		return result, io.EOF
	}
	return result, nil
}

// vfsDirEntry implements fs.DirEntry.
type vfsDirEntry struct {
	name  string
	isDir bool
}

func (e *vfsDirEntry) Name() string { return e.name }
func (e *vfsDirEntry) IsDir() bool  { return e.isDir }
func (e *vfsDirEntry) Type() fs.FileMode {
	if e.isDir {
		return fs.ModeDir
	}
	return 0
}
func (e *vfsDirEntry) Info() (fs.FileInfo, error) {
	return &vfsFileInfo{name: e.name, isDir: e.isDir}, nil
}

// vfsFileInfo implements fs.FileInfo with minimal information.
type vfsFileInfo struct {
	name  string
	isDir bool
}

func (i *vfsFileInfo) Name() string { return i.name }
func (i *vfsFileInfo) Size() int64  { return 0 }
func (i *vfsFileInfo) Mode() fs.FileMode {
	if i.isDir {
		return fs.ModeDir | 0o755
	}
	return 0o644
}
func (i *vfsFileInfo) ModTime() time.Time { return time.Time{} }
func (i *vfsFileInfo) IsDir() bool        { return i.isDir }
func (i *vfsFileInfo) Sys() any           { return nil }
