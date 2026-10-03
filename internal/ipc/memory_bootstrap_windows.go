package ipc

import "context"

// Windows duplicates the reader's anonymous mapping handle after preparation.
type memoryBootstrap struct{}

func listenMemoryBootstrap(context.Context) (*memoryBootstrap, error) {
	return &memoryBootstrap{}, nil
}
func (*memoryBootstrap) path() string { return "" }
func (*memoryBootstrap) close()       {}
func (*memoryBootstrap) receive(_ context.Context, descriptor MemoryMapping) (memoryMapping, error) {
	return mapMemory(descriptor)
}
