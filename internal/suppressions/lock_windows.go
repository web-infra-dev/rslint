//go:build windows

package suppressions

import (
	"context"
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

func lockBaselineUpdate(ctx context.Context, targetPath string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	lockPath := lockPathForBaseline(targetPath)
	// Go opens files on Windows without FILE_SHARE_DELETE, so no other process
	// can remove lockPath while any handle remains open. Retry briefly in case
	// a filter driver leaves a just-removed lock file delete-pending.
	var f *os.File
	err := retryWindowsTransient(ctx, func() error {
		var openErr error
		f, openErr = os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
		return openErr
	})
	if err != nil {
		return nil, err
	}
	var overlapped windows.Overlapped
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &overlapped); err != nil {
		_ = f.Close()
		return nil, err
	}
	return func() {
		var unlockOverlapped windows.Overlapped
		_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &unlockOverlapped)
		_ = f.Close()
		// Best-effort cleanup: succeeds for the last holder and harmlessly
		// returns ERROR_SHARING_VIOLATION when another writer holds a handle.
		_ = os.Remove(lockPath)
	}, nil
}

func replaceBaselineFile(ctx context.Context, tempName, path string) error {
	// A concurrent read-only check opens path without FILE_SHARE_DELETE on
	// Windows, which can briefly block MoveFileExW. Retry transiently once
	// write permission on path has already been verified.
	return retryWindowsTransient(ctx, func() error {
		return os.Rename(tempName, path)
	})
}

func retryWindowsTransient(ctx context.Context, op func() error) error {
	var err error
	for attempt := range 25 {
		err = op()
		if err == nil {
			return nil
		}
		if !errors.Is(err, windows.ERROR_SHARING_VIOLATION) && !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			return err
		}
		if attempt == 24 {
			return err
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return err
}
