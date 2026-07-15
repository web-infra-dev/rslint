package hostfs

import (
	"encoding/binary"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/microsoft/typescript-go/shim/vfs"
)

// nativePathFS preserves POSIX backslashes at the filesystem boundary. The
// TypeScript VFS normalizes both slash kinds before every operation, which is
// correct for ts-go identities but not for Node-compatible host paths where a
// backslash is a legal filename byte. Ordinary paths continue through the
// supplied VFS so cached, bundled, snapshot, and overlay behavior is retained.
type nativePathFS struct {
	vfs.FS
}

// NativeOS decorates the process OS filesystem before cache/bundle/overlay
// layers are applied. It is intentionally not a generic arbitrary-FS wrapper:
// an outer overlay must get the first opportunity to satisfy an exact virtual
// path before this bottom layer performs native POSIX I/O.
func NativeOS(fsys vfs.FS) vfs.FS {
	if fsys == nil || runtime.GOOS == "windows" {
		return fsys
	}
	if _, alreadyWrapped := fsys.(*nativePathFS); alreadyWrapped {
		return fsys
	}
	return &nativePathFS{FS: fsys}
}

func needsNativePOSIXOperation(path string) bool {
	return runtime.GOOS != "windows" && strings.ContainsRune(path, '\\')
}

func (fsys *nativePathFS) FileExists(path string) bool {
	if !needsNativePOSIXOperation(path) {
		return fsys.FS.FileExists(path)
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (fsys *nativePathFS) DirectoryExists(path string) bool {
	if !needsNativePOSIXOperation(path) {
		return fsys.FS.DirectoryExists(path)
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (fsys *nativePathFS) Stat(path string) vfs.FileInfo {
	if !needsNativePOSIXOperation(path) {
		return fsys.FS.Stat(path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	return info
}

func (fsys *nativePathFS) ReadFile(path string) (string, bool) {
	if !needsNativePOSIXOperation(path) {
		return fsys.FS.ReadFile(path)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return decodeBytes(contents), true
}

func decodeBytes(contents []byte) string {
	if len(contents) >= 2 {
		switch [2]byte{contents[0], contents[1]} {
		case [2]byte{0xff, 0xfe}:
			return decodeUTF16(contents[2:], binary.LittleEndian)
		case [2]byte{0xfe, 0xff}:
			return decodeUTF16(contents[2:], binary.BigEndian)
		}
	}
	if len(contents) >= 3 && contents[0] == 0xef && contents[1] == 0xbb && contents[2] == 0xbf {
		contents = contents[3:]
	}
	return string(contents)
}

func decodeUTF16(contents []byte, order binary.ByteOrder) string {
	units := make([]uint16, len(contents)/2)
	for index := range units {
		units[index] = order.Uint16(contents[index*2:])
	}
	return string(utf16.Decode(units))
}

func (fsys *nativePathFS) WriteFile(path string, data string) error {
	if !needsNativePOSIXOperation(path) {
		return fsys.FS.WriteFile(path, data)
	}
	return os.WriteFile(path, []byte(data), 0o666)
}

func (fsys *nativePathFS) AppendFile(path string, data string) error {
	if !needsNativePOSIXOperation(path) {
		return fsys.FS.AppendFile(path, data)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(data)
	return err
}

func (fsys *nativePathFS) Remove(path string) error {
	if !needsNativePOSIXOperation(path) {
		return fsys.FS.Remove(path)
	}
	return os.RemoveAll(path)
}

func (fsys *nativePathFS) Chtimes(path string, accessTime time.Time, modificationTime time.Time) error {
	if !needsNativePOSIXOperation(path) {
		return fsys.FS.Chtimes(path, accessTime, modificationTime)
	}
	return os.Chtimes(path, accessTime, modificationTime)
}

func (fsys *nativePathFS) Realpath(path string) string {
	if !needsNativePOSIXOperation(path) {
		return fsys.FS.Realpath(path)
	}
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	absolute, err := filepath.Abs(realPath)
	if err != nil {
		return path
	}
	return filepath.Clean(absolute)
}

func (fsys *nativePathFS) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	if !needsNativePOSIXOperation(root) {
		return fsys.FS.WalkDir(root, walkFn)
	}
	return filepath.WalkDir(root, walkFn)
}

func (fsys *nativePathFS) GetAccessibleEntries(path string) (result vfs.Entries) {
	if !needsNativePOSIXOperation(path) {
		return fsys.FS.GetAccessibleEntries(path)
	}
	result.Symlinks = make(map[string]struct{})
	entries, err := os.ReadDir(path)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		mode := entry.Type()
		isLink := mode&fs.ModeSymlink != 0
		if isLink || mode&fs.ModeType == 0 {
			if info, infoErr := os.Stat(filepath.Join(path, entry.Name())); infoErr == nil {
				mode = info.Mode()
			}
		}
		switch {
		case mode.IsDir():
			result.Directories = append(result.Directories, entry.Name())
		case mode.IsRegular():
			result.Files = append(result.Files, entry.Name())
		default:
			continue
		}
		if isLink {
			result.Symlinks[entry.Name()] = struct{}{}
		}
	}
	sort.Strings(result.Files)
	sort.Strings(result.Directories)
	return result
}
