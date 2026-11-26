//go:build unix && !aix

package suppressions

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

type lockedBaseline struct {
	file   *os.File
	mode   fs.FileMode
	unlock func()
}

func readBaselineFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func flockExclusive(f *os.File) error {
	for {
		err := unix.Flock(int(f.Fd()), unix.LOCK_EX)
		if !errors.Is(err, unix.EINTR) {
			return err
		}
	}
}

func lockExistingBaseline(ctx context.Context, path string) (*lockedBaseline, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// Open O_RDWR so Linux NFS accepts exclusive locks and read-only
		// baselines fail with os.ErrPermission before staging a replacement.
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if err != nil {
			if errors.Is(err, syscall.EISDIR) {
				return nil, fmt.Errorf("suppressions file %q must be a regular file", path)
			}
			return nil, err
		}
		ofd, err := lockUnixFile(f)
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		fdInfo, fdErr := f.Stat()
		pathInfo, pathErr := os.Stat(path)
		if fdErr == nil && pathErr == nil && os.SameFile(fdInfo, pathInfo) {
			if !fdInfo.Mode().IsRegular() {
				unlockUnixFile(f, ofd)
				_ = f.Close()
				return nil, fmt.Errorf("suppressions file %q must be a regular file", path)
			}
			return &lockedBaseline{
				file: f,
				mode: fdInfo.Mode().Perm(),
				unlock: func() {
					unlockUnixFile(f, ofd)
					_ = f.Close()
				},
			}, nil
		}
		unlockUnixFile(f, ofd)
		_ = f.Close()
		if fdErr != nil {
			return nil, fdErr
		}
		if pathErr != nil {
			if errors.Is(pathErr, fs.ErrNotExist) {
				return nil, fs.ErrNotExist
			}
			return nil, pathErr
		}
	}
}

func replaceLockedBaseline(_ context.Context, tempName, path string, _ *lockedBaseline) error {
	return os.Rename(tempName, path)
}
