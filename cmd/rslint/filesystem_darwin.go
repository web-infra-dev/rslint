// cspell:ignore Attrlist Bitmapcount Commonattr Dirattr Fileattr FSOPT FULLPATH LINKCOUNT attrreference getattrlist

package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"golang.org/x/sys/unix"
)

func newCLIFileSystem() vfs.FS {
	return darwinCLIFileSystem{bundled.WrapFS(cachedvfs.From(darwinRealpathFS{osvfs.FS()}))}
}

type darwinCLIFileSystem struct{ vfs.FS }

// DirectoryEntryName returns the stored name in the input's parent, preserving
// a directory symlink's own name. An empty result means it could not be verified.
func (darwinCLIFileSystem) DirectoryEntryName(path string) string {
	return darwinDirectoryEntryName(path)
}

func darwinDirectoryEntryName(path string) string {
	path = strings.TrimRight(path, "/")
	if !filepath.IsAbs(path) {
		return ""
	}
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		return ""
	}
	stored, links := darwinPathAttribute(path, unix.ATTR_CMN_NAME, unix.FSOPT_NOFOLLOW)
	if links == 1 {
		return stored
	}
	if links > 1 {
		if info, err := os.Lstat(path); err == nil {
			reported, _ := darwinPathAttribute(path, unix.ATTR_CMN_FULLPATH, unix.FSOPT_NOFOLLOW)
			return darwinHardLinkEntryName(filepath.Dir(path), filepath.Base(path), reported, info)
		}
	}
	return ""
}

type darwinRealpathFS struct{ vfs.FS }

func (fs darwinRealpathFS) Realpath(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	physical, links := darwinPathAttribute(path, unix.ATTR_CMN_FULLPATH, 0)
	// An exact path also identifies the requested entry among its hard links.
	if links == 1 || physical == path {
		return tspath.NormalizeSlashes(physical)
	}
	resolved := fs.FS.Realpath(path)
	if links > 1 && physical != resolved {
		// The kernel can report another hard link's name. Follow the requested
		// path's symlinks and confirm the stored entry in that parent.
		if info, err := os.Lstat(resolved); err == nil {
			parent := filepath.Dir(resolved)
			if stored, parentLinks := darwinPathAttribute(parent, unix.ATTR_CMN_FULLPATH, 0); parentLinks == 1 {
				parent = stored
			}
			name := filepath.Base(resolved)
			if stored := darwinHardLinkEntryName(parent, name, physical, info); stored != "" {
				name = stored
			}
			return tspath.NormalizeSlashes(filepath.Join(parent, name))
		}
	}
	return resolved
}

func getCLIWorkingDirectory() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cwd = filepath.Clean(cwd)
	if native, err := syscall.Getwd(); err == nil && native == cwd {
		return cwd, nil
	}

	// PWD may contain deliberate symlinks. Look up the stored name of each
	// component without following its final link.
	parent := "/"
	for part := range strings.SplitSeq(strings.TrimPrefix(cwd, "/"), "/") {
		if stored := darwinDirectoryEntryName(filepath.Join(parent, part)); stored != "" {
			part = stored
		}
		parent = filepath.Join(parent, part)
	}
	return parent, nil
}

// darwinPathAttribute asks the kernel for a stored name and the hard link count
// in the same lookup. FSOPT_NOFOLLOW preserves the final symlink entry.
func darwinPathAttribute(path string, attribute uint32, flags uintptr) (string, uint32) {
	p, err := unix.BytePtrFromString(path)
	if err != nil {
		return "", 0
	}
	// Only the link count applicable to the object is packed.
	const directoryLinkCount = 1 // ATTR_DIR_LINKCOUNT from sys/attr.h.
	attrs := unix.Attrlist{
		Bitmapcount: unix.ATTR_BIT_MAP_COUNT, Commonattr: attribute,
		Dirattr: directoryLinkCount, Fileattr: unix.ATTR_FILE_LINKCOUNT,
	}
	var buf [unix.PathMax + 16]byte
	//nolint:staticcheck // Pure-Go builds lack a getattrlist wrapper.
	_, _, errno := unix.Syscall6(unix.SYS_GETATTRLIST, uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(&attrs)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), flags, 0)
	if errno != 0 {
		return "", 0
	}
	// getattrlist returns a byte count followed by an attrreference_t. Its
	// offset is relative to the reference itself and its length includes NUL.
	size := int(binary.NativeEndian.Uint32(buf[:]))
	start := 4 + int(binary.NativeEndian.Uint32(buf[4:]))
	end := start + int(binary.NativeEndian.Uint32(buf[8:]))
	if start != 16 || end <= start || end > size || size > len(buf) || buf[end-1] != 0 {
		return "", 0
	}
	if attribute == unix.ATTR_CMN_FULLPATH && buf[start] != '/' {
		return "", 0
	}
	return string(buf[start : end-1]), binary.NativeEndian.Uint32(buf[12:])
}
