package sharedsource

import (
	"errors"
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

func mapSources(descriptor Descriptor) ([]byte, func() error, error) {
	value, err := strconv.ParseUint(descriptor.Handle, 10, 64)
	if err != nil || value == 0 || uint64(uintptr(value)) != value || descriptor.ProcessID == 0 || descriptor.FD != 0 {
		return nil, nil, errors.New("invalid shared source descriptor")
	}
	process, err := windows.OpenProcess(windows.PROCESS_DUP_HANDLE, false, descriptor.ProcessID)
	if err != nil {
		return nil, nil, err
	}
	defer windows.CloseHandle(process)
	var handle windows.Handle
	if err := windows.DuplicateHandle(process, windows.Handle(value), windows.CurrentProcess(), &handle, windows.FILE_MAP_WRITE, false, 0); err != nil {
		return nil, nil, err
	}
	defer windows.CloseHandle(handle)
	address, err := windows.MapViewOfFile(handle, windows.FILE_MAP_WRITE, 0, 0, capacity)
	if err != nil {
		return nil, nil, err
	}
	data := unsafe.Slice((*byte)(unsafe.Pointer(address)), capacity)
	return data, func() error { return windows.UnmapViewOfFile(address) }, nil
}
