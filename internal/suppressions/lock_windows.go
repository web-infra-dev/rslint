package suppressions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// cspell:ignore DACL

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

func openBaselineFile(path string, flags int) (*os.File, error) {
	if flags == os.O_RDWR {
		return os.OpenFile(path, flags, 0)
	}
	native, err := nativePath(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(native, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(handle), path), nil
}

func openReplacementStage(name string, existing *os.File) (*os.File, error) {
	native, err := nativePath(name)
	if err != nil {
		return nil, err
	}
	src := windows.Handle(existing.Fd())
	var flags uint32
	volumeErr := windows.GetVolumeInformationByHandle(src, nil, 0, nil, nil, &flags, nil, 0)
	const access = windows.OWNER_SECURITY_INFORMATION | windows.GROUP_SECURITY_INFORMATION |
		windows.DACL_SECURITY_INFORMATION | windows.LABEL_SECURITY_INFORMATION
	var attrs *windows.SecurityAttributes
	// SMB shares can expose file security without supporting volume queries.
	if volumeErr != nil || flags&windows.FILE_PERSISTENT_ACLS != 0 {
		security, err := windows.GetSecurityInfo(src, windows.SE_FILE_OBJECT, access)
		if err != nil {
			return nil, err
		}
		attrs = &windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: security}
	}
	// Keep the empty file exclusive until its access permissions have been verified.
	handle, err := windows.CreateFile(native, windows.GENERIC_WRITE|windows.READ_CONTROL, 0, attrs,
		windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &os.PathError{Op: "create", Path: name, Err: err}
	}
	stage := os.NewFile(uintptr(handle), name)
	if attrs != nil {
		// Windows can combine the supplied descriptor with inheritable directory ACEs.
		actual, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, access)
		if err == nil {
			want := attrs.SecurityDescriptor.String()
			if want == "" || actual.String() != want {
				err = errors.New("the directory grants different access to replacement files")
			}
		}
		if err != nil {
			_ = stage.Close()
			_ = os.Remove(name)
			return nil, err
		}
	}
	return stage, nil
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
