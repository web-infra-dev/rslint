//go:build darwin || linux

// cspell:words DONTWAIT NOSIGNAL NOSIGPIPE

package ipc

import (
	"errors"
	"io"
	"net"
	"os"
	"sync/atomic"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type rpcMapping struct {
	data          []byte
	waitDoorbell  func() error
	ringGoToNode  func()
	closeDoorbell func()
	unmap         func() error
}

func mapRpc(descriptor RpcMapping) (rpcMapping, error) {
	// Strictly validate protocol-owned inherited descriptors; never touch arbitrary FDs.
	if descriptor.Version != RpcVersion ||
		descriptor.FD != RpcMemInheritedFD ||
		descriptor.DoorbellFD != RpcDoorbellInheritedFD ||
		descriptor.Handle != "" ||
		descriptor.NodeToGoEvent != "" ||
		descriptor.GoToNodeEvent != "" ||
		descriptor.ProcessID != 0 {
		return rpcMapping{}, errors.New("invalid RPC shared memory descriptor")
	}
	defer unix.Close(RpcMemInheritedFD)

	var stat unix.Stat_t
	if err := unix.Fstat(RpcMemInheritedFD, &stat); err != nil {
		_ = unix.Close(RpcDoorbellInheritedFD)
		return rpcMapping{}, err
	}
	// Darwin rounds POSIX shared memory objects up to the host page size.
	if stat.Size < RpcCapacity {
		_ = unix.Close(RpcDoorbellInheritedFD)
		return rpcMapping{}, errors.New("invalid RPC shared memory mapping size")
	}
	data, err := unix.Mmap(RpcMemInheritedFD, 0, RpcCapacity, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		_ = unix.Close(RpcDoorbellInheritedFD)
		return rpcMapping{}, err
	}
	if atomic.LoadUint32((*uint32)(unsafe.Pointer(&data[rpcGlobalMagicOff]))) != RpcMagic {
		_ = unix.Munmap(data)
		_ = unix.Close(RpcDoorbellInheritedFD)
		return rpcMapping{}, errors.New("invalid RPC shared memory magic")
	}

	prepareRpcDoorbellSocket(RpcDoorbellInheritedFD)

	file := os.NewFile(uintptr(RpcDoorbellInheritedFD), "rslint-rpc-doorbell")
	conn, err := net.FileConn(file)
	_ = file.Close()
	if err != nil {
		_ = unix.Munmap(data)
		return rpcMapping{}, err
	}

	var rawConn syscall.RawConn
	if sc, ok := conn.(syscall.Conn); ok {
		rawConn, _ = sc.SyscallConn()
	}

	var wakeByte = [1]byte{1}
	var readBuf [256]byte

	return rpcMapping{
		data: data,
		waitDoorbell: func() error {
			n, err := conn.Read(readBuf[:])
			if n > 0 {
				if n == len(readBuf) && rawConn != nil {
					_ = rawConn.Control(func(fd uintptr) {
						for {
							more, _, recvErr := unix.Recvfrom(int(fd), readBuf[:], unix.MSG_DONTWAIT)
							if more <= 0 || recvErr != nil {
								break
							}
						}
					})
				}
				return nil
			}
			if err == nil {
				return io.EOF
			}
			return err
		},
		ringGoToNode: func() {
			if rawConn != nil {
				// Non-blocking send with MSG_NOSIGNAL (Linux) / SO_NOSIGPIPE (Darwin):
				// a full kernel buffer already has unread wake-ups so EAGAIN is safe to drop,
				// and a closed peer never triggers SIGPIPE.
				_ = rawConn.Control(func(fd uintptr) {
					_ = unix.Send(int(fd), wakeByte[:], rpcDoorbellSendFlags)
				})
				return
			}
			_, _ = conn.Write(wakeByte[:])
		},
		closeDoorbell: func() {
			_ = conn.Close()
		},
		unmap: func() error {
			_ = conn.Close()
			return unix.Munmap(data)
		},
	}, nil
}
