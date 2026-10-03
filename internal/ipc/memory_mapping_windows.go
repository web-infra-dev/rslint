// cspell:words READWRITE
package ipc

import (
	"errors"
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

func mapMemory(descriptor MemoryMapping) (memoryMapping, error) {
	value, err := strconv.ParseUint(descriptor.Handle, 10, 64)
	if err != nil || value == 0 || uint64(uintptr(value)) != value || descriptor.ProcessID == 0 || descriptor.FD != 0 {
		return memoryMapping{}, errors.New("invalid shared memory descriptor")
	}
	process, err := windows.OpenProcess(windows.PROCESS_DUP_HANDLE, false, descriptor.ProcessID)
	if err != nil {
		return memoryMapping{}, err
	}
	defer windows.CloseHandle(process)
	var handle windows.Handle
	if err := windows.DuplicateHandle(process, windows.Handle(value), windows.CurrentProcess(), &handle, windows.FILE_MAP_WRITE, false, 0); err != nil {
		return memoryMapping{}, err
	}
	defer windows.CloseHandle(handle)
	address, err := windows.MapViewOfFile(handle, windows.FILE_MAP_WRITE, 0, 0, MemoryCapacity)
	if err != nil {
		return memoryMapping{}, err
	}
	// The native arena uses SEC_RESERVE. Commit control words independently of
	// memory slots, including when an older peer did not prepare the header.
	if _, err := windows.VirtualAlloc(address, MemoryHeaderSize, windows.MEM_COMMIT, windows.PAGE_READWRITE); err != nil {
		_ = windows.UnmapViewOfFile(address)
		return memoryMapping{}, err
	}
	data := unsafe.Slice((*byte)(unsafe.Pointer(address)), MemoryCapacity)
	return memoryMapping{
		data: data,
		commitSlot: func(slot int) error {
			// Commit the whole slot: readers validate ranges against its capacity,
			// while the publication word carries only the generation, not length.
			start := address + uintptr(MemoryHeaderSize+slot*MemorySlotSize)
			_, err := windows.VirtualAlloc(start, MemorySlotSize, windows.MEM_COMMIT, windows.PAGE_READWRITE)
			return err
		},
		unmap: func() error { return windows.UnmapViewOfFile(address) },
	}, nil
}
