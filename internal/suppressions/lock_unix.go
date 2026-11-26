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

func lockExistingBaseline(ctx context.Context, path string) (*lockedBaseline, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		f, err := os.Open(path)
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
		fdInfo, fdErr := f.Stat()
		pathInfo, pathErr := os.Stat(path)
		if fdErr == nil && pathErr == nil && os.SameFile(fdInfo, pathInfo) {
			if !fdInfo.Mode().IsRegular() {
				_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
				_ = f.Close()
				return nil, fmt.Errorf("suppressions file %q must be a regular file", path)
			}
			return &lockedBaseline{
				file: f,
				mode: fdInfo.Mode().Perm(),
				unlock: func() {
					_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
					_ = f.Close()
				},
			}, nil
		}
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
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

func publishNewBaseline(_ context.Context, tempName, path string) error {
	err := os.Link(tempName, path)
	if err == nil {
		_ = os.Remove(tempName)
		return nil
	}
	if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.ENOTSUP) || errors.Is(err, syscall.EXDEV) {
		return os.Rename(tempName, path)
	}
	return err
}

func replaceLockedBaseline(_ context.Context, tempName, path string, _ *lockedBaseline) error {
	// Require write permission on an existing baseline before replacing it
	// via rename, which checks directory permissions on Unix.
	probe, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_ = probe.Close()
	return os.Rename(tempName, path)
}
