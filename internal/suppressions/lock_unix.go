//go:build unix && !aix

package suppressions

import (
	"context"
	"errors"
	"io/fs"
	"os"

	"golang.org/x/sys/unix"
)

func lockBaselineUpdate(ctx context.Context, targetPath string) (func(), error) {
	lockPath := lockPathForBaseline(targetPath)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o666)
		if err != nil {
			return nil, err
		}
		for {
			err = unix.Flock(int(f.Fd()), unix.LOCK_EX)
			if !errors.Is(err, unix.EINTR) {
				break
			}
		}
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		// Verify lockPath still refers to the locked inode so a waiter cannot
		// proceed on a file unlinked by the previous lock holder.
		fdInfo, fdErr := f.Stat()
		pathInfo, pathErr := os.Stat(lockPath)
		if fdErr == nil && pathErr == nil && os.SameFile(fdInfo, pathInfo) {
			return func() {
				_ = os.Remove(lockPath)
				_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
				_ = f.Close()
			}, nil
		}
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
		_ = f.Close()
		if fdErr != nil {
			return nil, fdErr
		}
		if pathErr != nil && !errors.Is(pathErr, fs.ErrNotExist) {
			return nil, pathErr
		}
	}
}

func replaceBaselineFile(_ context.Context, tempName, path string) error {
	return os.Rename(tempName, path)
}
