package suppressions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func nativePath(path string) (*uint16, error) {
	// CreateFile needs the extended form for long paths when Win32 long paths
	// are disabled. The CLI has already resolved this absolute native path.
	native := path
	if len(path) >= 248 {
		switch {
		case path[1] == ':':
			native = `\\?\` + path
		case strings.HasPrefix(path, `\\`) && !strings.HasPrefix(path, `\\?\`) && !strings.HasPrefix(path, `\\.\`):
			native = `\\?\UNC\` + path[2:]
		}
	}
	return windows.UTF16PtrFromString(native)
}

func lockBaseline(file *os.File) (*os.File, error) {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(file.Fd()), &info); err != nil {
		return nil, &os.PathError{Op: "identify", Path: file.Name(), Err: err}
	}
	name := fmt.Sprintf(".rslint-lock-%x-%x-%x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow)
	path := filepath.Join(filepath.Dir(file.Name()), name)
	native, err := nativePath(path)
	if err != nil {
		return nil, err
	}
	// The kernel removes this private guard on the last close, including after a crash.
	// CREATE_NEW leaves any pre-existing entry, including a reparse point, untouched.
	handle, err := windows.CreateFile(native, windows.DELETE, 0, nil, windows.CREATE_NEW,
		windows.FILE_FLAG_DELETE_ON_CLOSE|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, &os.PathError{Op: "lock", Path: path, Err: err}
	}
	return os.NewFile(uintptr(handle), path), nil
}

func renameNew(from, to string) error {
	source, err := nativePath(from)
	if err != nil {
		return err
	}
	destination, err := nativePath(to)
	if err != nil {
		return err
	}
	// With zero flags, MoveFileEx refuses to overwrite an existing file, including on FAT.
	if err := windows.MoveFileEx(source, destination, 0); err != nil {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: err}
	}
	return nil
}

func nativeUpdateError(err error) error {
	switch {
	case errors.Is(err, windows.ERROR_SHARING_VIOLATION), errors.Is(err, windows.ERROR_DELETE_PENDING):
		return fmt.Errorf("%w: %w", os.ErrExist, err)
	case errors.Is(err, windows.ERROR_ACCESS_DENIED):
		// Windows may report a delete-pending name or a genuine access restriction here.
		return fmt.Errorf("%w (if another writer is active, retry the command)", err)
	default:
		return err
	}
}
