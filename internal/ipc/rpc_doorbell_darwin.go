//go:build darwin

package ipc

import "golang.org/x/sys/unix"

const rpcDoorbellSendFlags = unix.MSG_DONTWAIT

func prepareRpcDoorbellSocket(fd int) {
	_ = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_NOSIGPIPE, 1)
}
