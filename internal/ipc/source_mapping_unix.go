//go:build darwin || linux

package ipc

import (
	"errors"

	"golang.org/x/sys/unix"
)

func mapSourceMapping(descriptor SourceDescriptor) (sourceMapping, error) {
	// Only the protocol's inherited descriptor belongs to this transport.
	// Never close an arbitrary descriptor supplied in malformed init data.
	if descriptor.FD != SourceInheritedFD || descriptor.Handle != "" || descriptor.ProcessID != 0 {
		return sourceMapping{}, errors.New("invalid shared source descriptor")
	}
	defer unix.Close(SourceInheritedFD)
	var stat unix.Stat_t
	if err := unix.Fstat(SourceInheritedFD, &stat); err != nil {
		return sourceMapping{}, err
	}
	// Darwin rounds POSIX shared objects up to the host page size. Only map
	// the protocol capacity, but permit that unused extra tail.
	if stat.Size < SourceCapacity {
		return sourceMapping{}, errors.New("invalid shared source mapping size")
	}
	data, err := unix.Mmap(SourceInheritedFD, 0, SourceCapacity, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return sourceMapping{}, err
	}
	return sourceMapping{data: data, unmap: func() error { return unix.Munmap(data) }}, nil
}
