//go:build windows

package suppressions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
)

type lockedBaseline struct {
	file   *os.File
	mode   fs.FileMode
	root   *os.Root
	unlock func()
}

func readBaselineFile(path string) ([]byte, error) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	// Grant FILE_SHARE_DELETE so read-only checks never block POSIX rename.
	handle, err := windows.CreateFile(
		ptr,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	f := os.NewFile(uintptr(handle), path)
	defer func() { _ = f.Close() }()
	return io.ReadAll(f)
}

func lockExistingBaseline(ctx context.Context, path string) (*lockedBaseline, error) {
	dir, base := filepath.Dir(path), filepath.Base(path)
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	for {
		if err := ctx.Err(); err != nil {
			_ = root.Close()
			return nil, err
		}
		// os.Root opens files with FILE_SHARE_READ|FILE_SHARE_WRITE|FILE_SHARE_DELETE,
		// so holding this lock does not prevent POSIX-semantics replacement.
		f, err := root.OpenFile(base, os.O_RDWR, 0)
		if err != nil {
			_ = root.Close()
			return nil, err
		}
		ov := windows.Overlapped{
			Offset:     uint32(baselineLockOffset),
			OffsetHigh: uint32(uint64(baselineLockOffset) >> 32),
		}
		if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, baselineLockLength, 0, &ov); err != nil {
			_ = f.Close()
			_ = root.Close()
			return nil, err
		}
		same, mode, checkErr := verifyLockedHandle(root, base, f)
		if checkErr == nil && same {
			if !mode.IsRegular() {
				unlockWindowsHandle(f)
				_ = root.Close()
				return nil, fmt.Errorf("suppressions file %q must be a regular file", path)
			}
			return &lockedBaseline{
				file: f,
				mode: mode.Perm(),
				root: root,
				unlock: func() {
					unlockWindowsHandle(f)
					_ = root.Close()
				},
			}, nil
		}
		unlockWindowsHandle(f)
		if checkErr != nil {
			_ = root.Close()
			return nil, checkErr
		}
	}
}

func unlockWindowsHandle(f *os.File) {
	ov := windows.Overlapped{
		Offset:     uint32(baselineLockOffset),
		OffsetHigh: uint32(uint64(baselineLockOffset) >> 32),
	}
	_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, baselineLockLength, 0, &ov)
	_ = f.Close()
}

func verifyLockedHandle(root *os.Root, base string, locked *os.File) (bool, fs.FileMode, error) {
	fdInfo, err := locked.Stat()
	if err != nil {
		return false, 0, err
	}
	var lockedID windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(locked.Fd()), &lockedID); err != nil {
		return false, 0, err
	}
	current, err := root.Open(base)
	if err != nil {
		return false, 0, err
	}
	var currentID windows.ByHandleFileInformation
	infoErr := windows.GetFileInformationByHandle(windows.Handle(current.Fd()), &currentID)
	_ = current.Close()
	if infoErr != nil {
		return false, 0, infoErr
	}
	same := lockedID.VolumeSerialNumber == currentID.VolumeSerialNumber &&
		lockedID.FileIndexHigh == currentID.FileIndexHigh &&
		lockedID.FileIndexLow == currentID.FileIndexLow
	return same, fdInfo.Mode(), nil
}

func publishNewBaseline(_ context.Context, tempName, path string) error {
	from, err := windows.UTF16PtrFromString(tempName)
	if err != nil {
		return err
	}
	to, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	// MoveFileEx with flags=0 atomically publishes only when path is absent on
	// every Windows filesystem (including exFAT and FAT32).
	err = windows.MoveFileEx(from, to, 0)
	if err == nil {
		return nil
	}
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) || errors.Is(err, windows.ERROR_FILE_EXISTS) {
		return fs.ErrExist
	}
	return &os.LinkError{Op: "rename", Old: tempName, New: path, Err: err}
}

func replaceLockedBaseline(ctx context.Context, tempName, path string, locked *lockedBaseline) error {
	tempBase, base := filepath.Base(tempName), filepath.Base(path)
	// os.Root.Rename uses FILE_RENAME_REPLACE_IF_EXISTS|FILE_RENAME_POSIX_SEMANTICS
	// on Windows. Retry briefly in case an external scanner holds path without FILE_SHARE_DELETE.
	var err error
	for attempt := range 25 {
		err = locked.root.Rename(tempBase, base)
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
