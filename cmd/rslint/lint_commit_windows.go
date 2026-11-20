package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type fixFileID struct {
	// FILE_ID_INFO contains the volume serial and the full 128-bit file ID.
	full struct {
		volume uint64
		index  [16]byte
	}
	legacy struct{ volume, indexHigh, indexLow uint32 }
	path   string
}

func statFixFile(path string) (id fixFileID, err error) {
	path = filepath.FromSlash(path)
	// Frozen destinations are absolute. Keep long drive and UNC paths usable
	// on hosts where CreateFile still requires an extended-length prefix.
	if len(path) >= 248 && filepath.IsAbs(path) && !strings.HasPrefix(path, `\\?\`) && !strings.HasPrefix(path, `\\.\`) {
		if tail, unc := strings.CutPrefix(path, `\\`); unc {
			path = `\\?\UNC\` + tail
		} else {
			path = `\\?\` + path
		}
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return id, err
	}
	const share = windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE
	handle, err := windows.CreateFile(name, 0, share, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return id, err
	}
	defer func() { err = errors.Join(err, windows.CloseHandle(handle)) }()
	err = windows.GetFileInformationByHandleEx(handle, windows.FileIdInfo, (*byte)(unsafe.Pointer(&id.full)), uint32(unsafe.Sizeof(id.full)))
	if err == nil && id.full.index != [16]byte{} {
		return id, nil
	}
	// Older providers expose only the identity used by os.SameFile.
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return fixFileID{}, err
	}
	id = fixFileID{}
	if info.FileIndexHigh != 0 || info.FileIndexLow != 0 {
		id.legacy.volume = info.VolumeSerialNumber
		id.legacy.indexHigh, id.legacy.indexLow = info.FileIndexHigh, info.FileIndexLow
	} else if info.NumberOfLinks > 1 {
		return id, fmt.Errorf("filesystem did not return an ID for a file with %d hard links", info.NumberOfLinks)
	} else {
		// Providers without a file ID can still distinguish canonical paths.
		id.path = path
	}
	return id, nil
}
