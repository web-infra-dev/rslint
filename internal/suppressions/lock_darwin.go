//go:build darwin

// cspell:ignore Renamex

package suppressions

import (
	"context"
	"errors"
	"io/fs"
	"os"

	"golang.org/x/sys/unix"
)

func lockUnixFile(f *os.File) (bool, error) {
	return false, flockExclusive(f)
}

func unlockUnixFile(f *os.File, _ bool) {
	_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
}

func publishNewBaseline(ctx context.Context, tempName, path string) error {
	if err := unix.RenamexNp(tempName, path, unix.RENAME_EXCL); err == nil {
		return nil
	} else if errors.Is(err, unix.EEXIST) {
		return fs.ErrExist
	} else if !errors.Is(err, unix.ENOTSUP) && !errors.Is(err, unix.EOPNOTSUPP) &&
		!errors.Is(err, unix.ENOSYS) && !errors.Is(err, unix.EPERM) && !errors.Is(err, unix.EINVAL) {
		return err
	}

	if handled, err := publishWithHardLink(tempName, path); handled {
		return err
	}
	return publishWithLockFile(ctx, tempName, path)
}
