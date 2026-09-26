//go:build darwin || linux

package sharedsource

import (
	"errors"

	"golang.org/x/sys/unix"
)

func mapSources(descriptor Descriptor) (mappedSources, error) {
	// fd 3 is the only extra inherited descriptor in the CLI spawn contract.
	// Never close an arbitrary descriptor supplied in malformed init data.
	if descriptor.FD != 3 || descriptor.Handle != "" || descriptor.ProcessID != 0 {
		return mappedSources{}, errors.New("invalid shared source descriptor")
	}
	defer unix.Close(3)
	var stat unix.Stat_t
	if err := unix.Fstat(3, &stat); err != nil {
		return mappedSources{}, err
	}
	// Darwin rounds POSIX shared objects up to the host page size. Only map
	// the negotiated capacity, but permit that unused extra tail.
	if stat.Size < capacity {
		return mappedSources{}, errors.New("invalid shared source mapping size")
	}
	data, err := unix.Mmap(3, 0, capacity, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return mappedSources{}, err
	}
	return mappedSources{data: data, unmap: func() error { return unix.Munmap(data) }}, nil
}
