//go:build unix && !aix

package suppressions

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func lockBaselineFile(targetPath string, exclusive bool) (func(), error) {
	f, err := os.OpenFile(updateLockPath(targetPath), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	mode := unix.LOCK_SH
	if exclusive {
		mode = unix.LOCK_EX
	}
	for {
		err = unix.Flock(int(f.Fd()), mode)
		if !errors.Is(err, unix.EINTR) {
			break
		}
	}
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return func() {
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
		_ = f.Close()
	}, nil
}
