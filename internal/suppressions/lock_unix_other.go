//go:build unix && !aix && !darwin && !linux

package suppressions

import (
	"context"
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
	if handled, err := publishWithHardLink(tempName, path); handled {
		return err
	}
	return publishWithLockFile(ctx, tempName, path)
}
