//go:build unix && !aix && !darwin && !linux

package suppressions

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func lockUnixFile(f *os.File) (bool, error) {
	return false, flockExclusive(f)
}

func unlockUnixFile(f *os.File, _ bool) {
	_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
}

func publishNewBaseline(_ context.Context, tempName, path string) error {
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
	if lockErr := flockExclusive(created); lockErr != nil {
		_ = created.Close()
		_ = os.Remove(path)
		return lockErr
	}
	renameErr := os.Rename(tempName, path)
	_ = unix.Flock(int(created.Fd()), unix.LOCK_UN)
	_ = created.Close()
	if renameErr != nil {
		_ = os.Remove(path)
		return renameErr
	}
	return nil
}
