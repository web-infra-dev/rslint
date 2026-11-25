//go:build windows

package suppressions

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

// inProcessUpdateMu serializes goroutines in the same process so two
// goroutines never re-enter the same thread-owned Win32 mutex.
var inProcessUpdateMu sync.Mutex

func lockBaselineUpdate(ctx context.Context, targetPath string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	name, err := directoryMutexName(filepath.Dir(targetPath))
	if err != nil {
		return nil, err
	}
	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}

	inProcessUpdateMu.Lock()
	runtime.LockOSThread()
	handle, err := windows.CreateMutex(nil, false, namePtr)
	if err != nil {
		runtime.UnlockOSThread()
		inProcessUpdateMu.Unlock()
		return nil, err
	}
	for {
		status, waitErr := windows.WaitForSingleObject(handle, 50)
		switch status {
		case windows.WAIT_OBJECT_0, windows.WAIT_ABANDONED:
			return func() {
				_ = windows.ReleaseMutex(handle)
				_ = windows.CloseHandle(handle)
				runtime.UnlockOSThread()
				inProcessUpdateMu.Unlock()
			}, nil
		case uint32(windows.WAIT_TIMEOUT):
			if err := ctx.Err(); err != nil {
				_ = windows.CloseHandle(handle)
				runtime.UnlockOSThread()
				inProcessUpdateMu.Unlock()
				return nil, err
			}
		default:
			_ = windows.CloseHandle(handle)
			runtime.UnlockOSThread()
			inProcessUpdateMu.Unlock()
			if waitErr != nil {
				return nil, waitErr
			}
			return nil, fmt.Errorf("wait for suppressions lock failed: status %#x", status)
		}
	}
}

// directoryMutexName derives a kernel mutex name from the destination
// directory's volume and file index so no lock file is created on disk and
// case/short-name/junction aliases of the same directory share one mutex.
func directoryMutexName(dir string) (string, error) {
	if ptr, err := windows.UTF16PtrFromString(dir); err == nil {
		h, openErr := windows.CreateFile(
			ptr,
			0,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
			nil,
			windows.OPEN_EXISTING,
			windows.FILE_FLAG_BACKUP_SEMANTICS,
			0,
		)
		if openErr == nil {
			var info windows.ByHandleFileInformation
			infoErr := windows.GetFileInformationByHandle(h, &info)
			_ = windows.CloseHandle(h)
			if infoErr == nil && (info.VolumeSerialNumber != 0 || info.FileIndexHigh != 0 || info.FileIndexLow != 0) {
				return fmt.Sprintf("Local\\rslint-suppressions-%08x-%08x%08x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow), nil
			}
		}
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	sum := sha256.Sum256([]byte(strings.ToLower(abs)))
	return fmt.Sprintf("Local\\rslint-suppressions-%x", sum[:16]), nil
}

func replaceBaselineFile(ctx context.Context, tempName, path string) error {
	// A concurrent read-only check opens path without FILE_SHARE_DELETE on
	// Windows, which can briefly block MoveFileExW. Retry transiently once
	// write permission on path has already been verified.
	var err error
	for attempt := range 25 {
		err = os.Rename(tempName, path)
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
