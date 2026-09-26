// cspell:words READWRITE
package sharedsource

import (
	"errors"
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

func mapSources(descriptor Descriptor) (mappedSources, error) {
	value, err := strconv.ParseUint(descriptor.Handle, 10, 64)
	if err != nil || value == 0 || uint64(uintptr(value)) != value || descriptor.ProcessID == 0 || descriptor.FD != 0 {
		return mappedSources{}, errors.New("invalid shared source descriptor")
	}
	process, err := windows.OpenProcess(windows.PROCESS_DUP_HANDLE, false, descriptor.ProcessID)
	if err != nil {
		return mappedSources{}, err
	}
	defer windows.CloseHandle(process)
	var handle windows.Handle
	if err := windows.DuplicateHandle(process, windows.Handle(value), windows.CurrentProcess(), &handle, windows.FILE_MAP_WRITE, false, 0); err != nil {
		return mappedSources{}, err
	}
	defer windows.CloseHandle(handle)
	address, err := windows.MapViewOfFile(handle, windows.FILE_MAP_WRITE, 0, 0, capacity)
	if err != nil {
		return mappedSources{}, err
	}
	// The native arena uses SEC_RESERVE. Commit control words independently of
	// source slots, including when an older peer did not prepare the header.
	if _, err := windows.VirtualAlloc(address, headerSize, windows.MEM_COMMIT, windows.PAGE_READWRITE); err != nil {
		_ = windows.UnmapViewOfFile(address)
		return mappedSources{}, err
	}
	data := unsafe.Slice((*byte)(unsafe.Pointer(address)), capacity)
	return mappedSources{
		data: data,
		commitSlot: func(slot int) error {
			// Commit the whole slot: readers validate ranges against its capacity,
			// while the publication word carries only the generation, not length.
			start := address + uintptr(headerSize+slot*SlotSize)
			_, err := windows.VirtualAlloc(start, SlotSize, windows.MEM_COMMIT, windows.PAGE_READWRITE)
			return err
		},
		unmap: func() error { return windows.UnmapViewOfFile(address) },
	}, nil
}
