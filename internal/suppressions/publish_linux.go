package suppressions

import (
	"os"

	"golang.org/x/sys/unix"
)

// cspell:ignore Fgetxattr Fsetxattr Fremovexattr Renameat NOREPLACE

func copyAccessACL(from, to *os.File) error {
	const accessACL = "system.posix_acl_access"
	src, dst := int(from.Fd()), int(to.Fd())
	size, err := unix.Fgetxattr(src, accessACL, nil)
	switch err {
	case unix.ENOTSUP:
		return nil // Both files are on the same filesystem.
	case unix.ENODATA:
		// The destination may have inherited an ACL from the directory.
		err = unix.Fremovexattr(dst, accessACL)
		if err == unix.ENODATA {
			return nil
		}
		return err
	case nil:
		acl := make([]byte, size)
		n, err := unix.Fgetxattr(src, accessACL, acl)
		if err != nil {
			return err
		}
		return unix.Fsetxattr(dst, accessACL, acl[:n], 0)
	default:
		return err
	}
}

func renameNew(from, to string) error {
	err := unix.Renameat2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_NOREPLACE)
	if err == unix.ENOSYS || err == unix.EINVAL || err == unix.EOPNOTSUPP {
		return linkNew(from, to)
	}
	if err != nil {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: err}
	}
	return nil
}
