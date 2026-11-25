//go:build unix && !aix

package suppressions

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func lockBaselineUpdate(ctx context.Context, targetPath string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Lock the destination directory inode directly so no lock file is created
	// in the workspace and all equivalent filename spellings share one lock.
	dir, err := os.Open(filepath.Dir(targetPath))
	if err != nil {
		return nil, err
	}
	for {
		err = unix.Flock(int(dir.Fd()), unix.LOCK_EX)
		if !errors.Is(err, unix.EINTR) {
			break
		}
	}
	if err != nil {
		_ = dir.Close()
		return nil, err
	}
	return func() {
		_ = unix.Flock(int(dir.Fd()), unix.LOCK_UN)
		_ = dir.Close()
	}, nil
}

func replaceBaselineFile(_ context.Context, tempName, path string) error {
	return os.Rename(tempName, path)
}
