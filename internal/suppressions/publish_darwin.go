package suppressions

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// cspell:ignore getattrlist setattrlist Attrlist Bitmapcount Commonattr FSOPT FULLSIZE Renamex

//nolint:staticcheck // The CLI builds without cgo. x/sys has no libSystem wrappers for these descriptor-based calls.
const (
	getFileAttributes = unix.SYS_FGETATTRLIST
	setFileAttributes = unix.SYS_FSETATTRLIST
)

func copyAccessACL(from, to *os.File) error {
	want, err := darwinFileACL(from)
	if err != nil {
		return err
	}
	inherited, err := darwinFileACL(to)
	if err != nil {
		return err
	}
	if bytes.Equal(inherited, want) {
		return nil
	}
	// macOS can grant inherited access even with zero mode. An open handle
	// would retain that access after chmod. Leave this empty inode unused.
	if len(inherited) != 0 {
		return errors.New("the directory grants a different ACL to replacement files")
	}
	return darwinACL(setFileAttributes, to, want, 0)
}

func darwinFileACL(file *os.File) ([]byte, error) {
	// getattrlist returns a length and an attribute reference before the ACL.
	// setattrlist accepts the same buffer without the leading length.
	const header = 12
	buf := make([]byte, header)
	for {
		if err := darwinACL(getFileAttributes, file, buf, unix.FSOPT_REPORT_FULLSIZE); err != nil {
			if errors.Is(err, unix.ENOTSUP) {
				return nil, nil
			}
			return nil, err
		}
		size := int(binary.NativeEndian.Uint32(buf))
		switch {
		case size < header:
			return nil, io.ErrUnexpectedEOF
		case size == header:
			return nil, nil
		case size <= len(buf):
			return buf[4:size], nil
		}
		buf = make([]byte, size)
	}
}

func darwinACL(op uintptr, file *os.File, buf []byte, options uintptr) error {
	attrs := unix.Attrlist{Bitmapcount: unix.ATTR_BIT_MAP_COUNT, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	_, _, errno := unix.Syscall6(op, file.Fd(), uintptr(unsafe.Pointer(&attrs)),
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), options, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

func renameNew(from, to string) error {
	err := unix.RenamexNp(from, to, unix.RENAME_EXCL)
	if err == unix.EINVAL || err == unix.ENOTSUP {
		return linkNew(from, to)
	}
	if err != nil {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: err}
	}
	return nil
}
