//go:build darwin || linux

package ipc

import (
	"errors"

	"golang.org/x/sys/unix"
)

func mapMemory(descriptor MemoryMapping) (memoryMapping, error) {
	// Only the protocol's inherited descriptor belongs to this transport.
	// Never close an arbitrary descriptor supplied in malformed init data.
	if descriptor.FD != MemoryInheritedFD || descriptor.Handle != "" || descriptor.ProcessID != 0 {
		return memoryMapping{}, errors.New("invalid shared memory descriptor")
	}
	defer unix.Close(MemoryInheritedFD)
	return mapMemoryFD(MemoryInheritedFD)
}

// The caller owns fd, received either through inheritance or SCM_RIGHTS.
func mapMemoryFD(fd int) (memoryMapping, error) {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return memoryMapping{}, err
	}
	// Darwin rounds POSIX shared objects up to the host page size. Only map
	// the protocol capacity, but permit that unused extra tail.
	if stat.Size < MemoryCapacity {
		return memoryMapping{}, errors.New("invalid shared memory mapping size")
	}
	data, err := unix.Mmap(fd, 0, MemoryCapacity, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return memoryMapping{}, err
	}
	return memoryMapping{data: data, unmap: func() error { return unix.Munmap(data) }}, nil
}
