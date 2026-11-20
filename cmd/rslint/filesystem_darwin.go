// cspell:ignore Attrlist Bitmapcount Commonattr Dirattr Fileattr FSOPT FULLPATH LINKCOUNT Nlink attrreference getattrlist

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
	return bundled.WrapFS(cachedvfs.From(darwinCLIFileSystem{osvfs.FS()}))
}

type darwinCLIFileSystem struct{ vfs.FS }

func (fs darwinCLIFileSystem) Realpath(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	physical, links := darwinPathAttribute(path, unix.ATTR_CMN_FULLPATH)
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
			if stored, parentLinks := darwinPathAttribute(parent, unix.ATTR_CMN_FULLPATH); parentLinks == 1 {
				parent = stored
			}
			name := darwinHardLinkEntryName(parent, filepath.Base(resolved), filepath.Base(physical), info)
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
		path := filepath.Join(parent, part)
		name := darwinEntryName(path)
		if name != "" && name != part {
			onDisk := filepath.Join(parent, name)
			original, err := os.Lstat(path)
			if err == nil {
				stat, ok := original.Sys().(*syscall.Stat_t)
				if ok && original.Mode()&os.ModeSymlink != 0 && stat.Nlink > 1 {
					path = filepath.Join(parent, darwinHardLinkEntryName(parent, part, name, original))
				} else if candidate, err := os.Lstat(onDisk); err == nil && os.SameFile(original, candidate) {
					path = onDisk
				}
			}
		}
		parent = path
	}
	return parent, nil
}

func darwinEntryName(path string) string {
	name, _ := darwinPathAttribute(path, unix.ATTR_CMN_NAME)
	return name
}

func darwinPathAttribute(path string, attribute uint32) (string, uint32) {
	p, err := unix.BytePtrFromString(path)
	if err != nil {
		return "", 0
	}
	attrs := unix.Attrlist{Bitmapcount: unix.ATTR_BIT_MAP_COUNT, Commonattr: attribute}
	flags := uintptr(unix.FSOPT_NOFOLLOW)
	if attribute == unix.ATTR_CMN_FULLPATH {
		// Only the link count applicable to the object is packed. Request it
		// with the full path so hard-linked names can be verified separately.
		const directoryLinkCount = 1 // ATTR_DIR_LINKCOUNT from sys/attr.h.
		attrs.Dirattr = directoryLinkCount
		attrs.Fileattr = unix.ATTR_FILE_LINKCOUNT
		flags = 0
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
	if start < 12 || end <= start || end > size || size > len(buf) || buf[end-1] != 0 {
		return "", 0
	}
	if attribute == unix.ATTR_CMN_FULLPATH {
		if start != 16 || buf[start] != '/' {
			return "", 0
		}
		return string(buf[start : end-1]), binary.NativeEndian.Uint32(buf[12:])
	}
	return string(buf[start : end-1]), 0
}
