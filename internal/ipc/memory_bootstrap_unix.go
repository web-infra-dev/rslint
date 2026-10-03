//go:build darwin || linux

package ipc

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// This one-shot socket transfers only an anonymous descriptor. Source bytes
// stay in memory; all requests, publication and acknowledgements use Channel.
type memoryBootstrap struct {
	dir      string
	listener *net.UnixListener
	stop     func() bool
}

func listenMemoryBootstrap(ctx context.Context) (*memoryBootstrap, error) {
	dir, err := os.MkdirTemp("", "rslint-ipc-")
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: filepath.Join(dir, "fd"), Net: "unix"})
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}
	b := &memoryBootstrap{dir: dir, listener: listener}
	b.stop = context.AfterFunc(ctx, func() { _ = listener.Close() })
	return b, nil
}

func (b *memoryBootstrap) path() string { return b.listener.Addr().String() }

func (b *memoryBootstrap) close() {
	b.stop()
	_ = b.listener.Close()
	_ = os.RemoveAll(b.dir)
}

func (b *memoryBootstrap) receive(ctx context.Context, descriptor MemoryMapping) (memoryMapping, error) {
	if descriptor.FD != 0 || descriptor.Handle != "" || descriptor.ProcessID != 0 {
		return memoryMapping{}, errors.New("invalid transferred memory descriptor")
	}
	conn, err := b.listener.AcceptUnix()
	if err != nil {
		return memoryMapping{}, err
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	var data [1]byte
	oob := make([]byte, unix.CmsgSpace(4))
	n, controlSize, flags, _, err := conn.ReadMsgUnix(data[:], oob)
	if err != nil {
		return memoryMapping{}, err
	}
	messages, err := unix.ParseSocketControlMessage(oob[:controlSize])
	if err != nil {
		return memoryMapping{}, err
	}
	var fds []int
	defer func() {
		for _, fd := range fds {
			_ = unix.Close(fd)
		}
	}()
	for _, message := range messages {
		rights, err := unix.ParseUnixRights(&message)
		if err != nil {
			return memoryMapping{}, err
		}
		fds = append(fds, rights...)
	}
	if n != 1 || flags&unix.MSG_CTRUNC != 0 || len(fds) != 1 {
		return memoryMapping{}, errors.New("invalid shared memory fd transfer")
	}
	unix.CloseOnExec(fds[0])
	return mapMemoryFD(fds[0])
}
