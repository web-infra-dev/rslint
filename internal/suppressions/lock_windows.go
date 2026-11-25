//go:build windows

package suppressions

import (
	"os"

	"golang.org/x/sys/windows"
)

func lockBaselineFile(targetPath string, exclusive bool) (func(), error) {
	f, err := os.OpenFile(updateLockPath(targetPath), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	var flags uint32
	if exclusive {
		flags = windows.LOCKFILE_EXCLUSIVE_LOCK
	}
	var overlapped windows.Overlapped
	if err := windows.LockFileEx(windows.Handle(f.Fd()), flags, 0, 1, 0, &overlapped); err != nil {
		_ = f.Close()
		return nil, err
	}
	return func() {
		var unlockOverlapped windows.Overlapped
		_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &unlockOverlapped)
		_ = f.Close()
	}, nil
}
