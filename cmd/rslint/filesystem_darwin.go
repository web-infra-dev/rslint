// cspell:ignore APFS Fstypename Attrlist Bitmapcount Commonattr Dirattr Fileattr FSOPT FULLPATH LINKCOUNT attrreference getattrlist

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
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
	"golang.org/x/sys/unix"
)

func newCLIFileSystem() vfs.FS {
	return darwinCLIFileSystem{bundled.WrapFS(cachedvfs.From(darwinRealpathFS{osvfs.FS()}))}
}

func newLSPFileSystem() vfs.FS {
	return darwinCLIFileSystem{bundled.WrapFS(osvfs.FS())}
}

type darwinCLIFileSystem struct{ vfs.FS }

var _ gitignore.StoredEntryNamer = darwinCLIFileSystem{}

// SameDirectoryEntry compares the final directory entries without merging
// distinct hard links to a symlink. The caller checks each earlier component.
func (darwinCLIFileSystem) SameDirectoryEntry(left, right string) bool {
	if !filepath.IsAbs(left) || !filepath.IsAbs(right) || !darwinSameParent(left, right) {
		return false
	}
	name, _ := darwinStoredEntryName(left)
	other, _ := darwinStoredEntryName(right)
	if name != "" && other != "" {
		return name == other
	}
	return darwinSameCaseLookup(left, right)
}

// APFS preserves literal ASCII name prefixes and suffixes. A case-sensitive
// parent additionally rules out spellings that differ only in ASCII case.
func (darwinCLIFileSystem) ASCIILookupAnchors(path string) (preserved, exactCase bool) {
	parent := filepath.Dir(path)
	var volume unix.Statfs_t
	if unix.Statfs(parent, &volume) != nil || unix.ByteSliceToString(volume.Fstypename[:]) != "apfs" {
		return false, false
	}
	sensitive, err := unix.Pathconf(parent, pcCaseSensitive)
	return true, err == nil && sensitive > 0
}

// StoredEntryName returns the actual leaf name without following its symlink.
// A search-only parent suffices when the kernel can identify one entry.
func (darwinCLIFileSystem) StoredEntryName(path string) (string, error) {
	return darwinStoredEntryName(path)
}

func darwinStoredEntryName(path string) (string, error) {
	path = strings.TrimRight(path, "/")
	if !filepath.IsAbs(path) {
		return "", nil
	}
	stored, links := darwinPathAttribute(path, unix.ATTR_CMN_NAME, unix.FSOPT_NOFOLLOW)
	if links == 1 {
		return stored, nil
	}
	if links > 1 {
		if info, err := os.Lstat(path); err == nil {
			reported, _ := darwinPathAttribute(path, unix.ATTR_CMN_FULLPATH, unix.FSOPT_NOFOLLOW)
			return darwinHardLinkEntryName(filepath.Dir(path), filepath.Base(path), reported, info)
		}
	}
	return "", nil
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
			if stored, _ := darwinHardLinkEntryName(parent, name, physical, info); stored != "" {
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
		if stored, _ := darwinStoredEntryName(filepath.Join(parent, part)); stored != "" {
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
