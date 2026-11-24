package suppressions

import (
	"os"

	"golang.org/x/sys/unix"
)

func renameNew(from, to string) error {
	// cspell:ignore Renamex
	err := unix.RenamexNp(from, to, unix.RENAME_EXCL)
	if err == unix.EINVAL || err == unix.ENOTSUP {
		return linkNew(from, to)
	}
	if err != nil {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: err}
	}
	return nil
}
