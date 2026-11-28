//go:build linux

// cspell:words DONTWAIT NOSIGNAL

package ipc

import "golang.org/x/sys/unix"

const rpcDoorbellSendFlags = unix.MSG_DONTWAIT | unix.MSG_NOSIGNAL

func prepareRpcDoorbellSocket(fd int) {}
