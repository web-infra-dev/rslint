//go:build linux

// cspell:ignore SETLKW SETLK Renameat NOREPLACE

package suppressions

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"

	"golang.org/x/sys/unix"
)

const publishCreateExtraFlags = 0

func lockUnixFile(f *os.File) (bool, error) {
	lk := unix.Flock_t{
		Type:   unix.F_WRLCK,
		Whence: io.SeekStart,
		Start:  baselineLockOffset,
		Len:    baselineLockLength,
	}
	for {
		err := unix.FcntlFlock(f.Fd(), unix.F_OFD_SETLKW, &lk)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if err == nil {
			return true, nil
		}
		if errors.Is(err, unix.EINVAL) || errors.Is(err, unix.ENOSYS) {
			return false, flockExclusive(f)
		}
		return false, err
	}
}

func unlockUnixFile(f *os.File, ofd bool) {
	if ofd {
		lk := unix.Flock_t{
			Type:   unix.F_UNLCK,
			Whence: io.SeekStart,
			Start:  baselineLockOffset,
			Len:    baselineLockLength,
		}
		_ = unix.FcntlFlock(f.Fd(), unix.F_OFD_SETLK, &lk)
		return
	}
	_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
}

func publishNewBaseline(ctx context.Context, tempName, path string) error {
	if err := unix.Renameat2(unix.AT_FDCWD, tempName, unix.AT_FDCWD, path, unix.RENAME_NOREPLACE); err == nil {
		return nil
	} else if errors.Is(err, unix.EEXIST) {
		return fs.ErrExist
	} else if !errors.Is(err, unix.ENOSYS) && !errors.Is(err, unix.EINVAL) &&
		!errors.Is(err, unix.EOPNOTSUPP) && !errors.Is(err, unix.ENOTSUP) && !errors.Is(err, unix.EPERM) {
		return err
	}

	if handled, err := publishWithHardLink(tempName, path); handled {
		return err
	}
	return publishWithLockFile(ctx, tempName, path)
}
