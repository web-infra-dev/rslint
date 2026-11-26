//go:build darwin

// cspell:ignore Renamex

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
	if err := unix.RenamexNp(tempName, path, unix.RENAME_EXCL); err == nil {
		return nil
	} else if errors.Is(err, unix.EEXIST) {
		return fs.ErrExist
	} else if !errors.Is(err, unix.ENOTSUP) && !errors.Is(err, unix.EPERM) && !errors.Is(err, unix.EINVAL) {
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

	return publishWithExclusiveOpen(tempName, path)
}

func publishWithExclusiveOpen(tempName, path string) error {
	// On filesystems without RENAME_EXCL or hard links (such as exFAT),
	// O_CREATE|O_EXCL|O_EXLOCK atomically grants creation and an exclusive
	// lock to exactly one writer. Contenders get fs.ErrExist and block in
	// lockExistingBaseline until os.Rename replaces the placeholder inode.
	created, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR|unix.O_EXLOCK, 0o666)
	if err != nil {
		return err
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
