//go:build !darwin && !linux && !windows

package ipc

import (
	"context"
	"errors"
)

type memoryBootstrap struct{}

func listenMemoryBootstrap(context.Context) (*memoryBootstrap, error) {
	return nil, errors.New("shared memory unavailable on this platform")
}
func (*memoryBootstrap) path() string { return "" }
func (*memoryBootstrap) close()       {}
func (*memoryBootstrap) receive(context.Context, MemoryMapping) (memoryMapping, error) {
	return memoryMapping{}, errors.New("shared memory unavailable on this platform")
}
