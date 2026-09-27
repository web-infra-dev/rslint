// cspell:words READWRITE
package ipc

import (
	"errors"
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

func mapSourceMapping(descriptor SourceDescriptor) (sourceMapping, error) {
	value, err := strconv.ParseUint(descriptor.Handle, 10, 64)
	if err != nil || value == 0 || uint64(uintptr(value)) != value || descriptor.ProcessID == 0 || descriptor.FD != 0 {
		return sourceMapping{}, errors.New("invalid shared source descriptor")
	}
	process, err := windows.OpenProcess(windows.PROCESS_DUP_HANDLE, false, descriptor.ProcessID)
	if err != nil {
		return sourceMapping{}, err
	}
	defer windows.CloseHandle(process)
	var handle windows.Handle
	if err := windows.DuplicateHandle(process, windows.Handle(value), windows.CurrentProcess(), &handle, windows.FILE_MAP_WRITE, false, 0); err != nil {
		return sourceMapping{}, err
	}
	defer windows.CloseHandle(handle)
	address, err := windows.MapViewOfFile(handle, windows.FILE_MAP_WRITE, 0, 0, SourceCapacity)
	if err != nil {
		return sourceMapping{}, err
	}
	// The native arena uses SEC_RESERVE. Commit control words independently of
	// source slots, including when an older peer did not prepare the header.
	if _, err := windows.VirtualAlloc(address, SourceHeaderSize, windows.MEM_COMMIT, windows.PAGE_READWRITE); err != nil {
		_ = windows.UnmapViewOfFile(address)
		return sourceMapping{}, err
	}
	data := unsafe.Slice((*byte)(unsafe.Pointer(address)), SourceCapacity)
	return sourceMapping{
		data: data,
		commitSlot: func(slot int) error {
			// Commit the whole slot: readers validate ranges against its capacity,
			// while the publication word carries only the generation, not length.
			start := address + uintptr(SourceHeaderSize+slot*SourceSlotSize)
			_, err := windows.VirtualAlloc(start, SourceSlotSize, windows.MEM_COMMIT, windows.PAGE_READWRITE)
			return err
		},
		unmap: func() error { return windows.UnmapViewOfFile(address) },
	}, nil
}
