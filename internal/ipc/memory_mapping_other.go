//go:build !darwin && !linux && !windows

package ipc

import "errors"

func mapMemory(MemoryMapping) (memoryMapping, error) {
	return memoryMapping{}, errors.New("shared memory unavailable on this platform")
}
