//go:build linux

// cspell:ignore SETLKW SETLK Renameat NOREPLACE

package suppressions

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// Lock a single byte far beyond EOF using Linux OFD locks so writes serialize
// per open file description without overlapping [0, EOF) reads on SMB mounts
// with mandatory locking, while remaining compatible with NFS.
const linuxLockOffset = 0x7ffffffe

func lockUnixFile(f *os.File) (bool, error) {
	lk := unix.Flock_t{
		Type:   unix.F_WRLCK,
		Whence: io.SeekStart,
		Start:  linuxLockOffset,
		Len:    1,
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
			Start:  linuxLockOffset,
			Len:    1,
		}
		_ = unix.FcntlFlock(f.Fd(), unix.F_OFD_SETLK, &lk)
		return
	}
	_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
}

func publishNewBaseline(_ context.Context, tempName, path string) error {
	if err := unix.Renameat2(unix.AT_FDCWD, tempName, unix.AT_FDCWD, path, unix.RENAME_NOREPLACE); err == nil {
		return nil
	} else if errors.Is(err, unix.EEXIST) {
		return fs.ErrExist
	} else if !errors.Is(err, unix.ENOSYS) && !errors.Is(err, unix.EINVAL) && !errors.Is(err, unix.EOPNOTSUPP) && !errors.Is(err, unix.EPERM) {
		return err
	}

	if err := os.Link(tempName, path); err == nil {
		_ = os.Remove(tempName)
		return nil
	} else if errors.Is(err, fs.ErrExist) {
		return fs.ErrExist
	} else if !errors.Is(err, syscall.ENOTSUP) && !errors.Is(err, syscall.EPERM) && !errors.Is(err, syscall.EXDEV) {
		return err
	}

	created, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o666)
	if err != nil {
		return err
	}
	ofd, lockErr := lockUnixFile(created)
	if lockErr != nil {
		_ = created.Close()
		_ = os.Remove(path)
		return lockErr
	}
	renameErr := os.Rename(tempName, path)
	unlockUnixFile(created, ofd)
	_ = created.Close()
	if renameErr != nil {
		_ = os.Remove(path)
		return renameErr
	}
	return nil
}
