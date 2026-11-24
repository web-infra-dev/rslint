package suppressions

import (
	"os"

	"golang.org/x/sys/unix"
)

func renameNew(from, to string) error {
	// cspell:ignore Renameat NOREPLACE
	err := unix.Renameat2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_NOREPLACE)
	if err == unix.ENOSYS || err == unix.EINVAL || err == unix.EOPNOTSUPP {
		return linkNew(from, to)
	}
	if err != nil {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: err}
	}
	return nil
}
