// cspell:ignore Attrlist Bitmapcount Commonattr FSOPT Nlink attrreference getattrlist

package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

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
			candidate, _ := os.Lstat(onDisk)
			// A symlink with multiple hard links may report a different entry's name.
			if err == nil && os.SameFile(original, candidate) {
				stat, ok := original.Sys().(*syscall.Stat_t)
				if ok && (original.Mode()&os.ModeSymlink == 0 || stat.Nlink == 1) {
					path = onDisk
				}
			}
		}
		parent = path
	}
	return parent, nil
}

func darwinEntryName(path string) string {
	p, err := unix.BytePtrFromString(path)
	if err != nil {
		return ""
	}
	attrs := unix.Attrlist{Bitmapcount: unix.ATTR_BIT_MAP_COUNT, Commonattr: unix.ATTR_CMN_NAME}
	var buf [unix.PathMax + 12]byte
	//nolint:staticcheck // Pure-Go builds lack a getattrlist wrapper and keep the shell spelling if the syscall fails.
	_, _, errno := unix.Syscall6(unix.SYS_GETATTRLIST, uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(&attrs)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), unix.FSOPT_NOFOLLOW, 0)
	if errno != 0 {
		return ""
	}
	// getattrlist returns a byte count followed by an attrreference_t. Its
	// offset is relative to the reference itself and its length includes NUL.
	size := int(binary.NativeEndian.Uint32(buf[:]))
	start := 4 + int(binary.NativeEndian.Uint32(buf[4:]))
	end := start + int(binary.NativeEndian.Uint32(buf[8:]))
	if start < 12 || end <= start || end > size || size > len(buf) || buf[end-1] != 0 {
		return ""
	}
	return string(buf[start : end-1])
}
