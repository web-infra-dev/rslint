//go:build !darwin && !linux && !windows

package ipc

import "errors"

type rpcMapping struct {
	data          []byte
	waitDoorbell  func() error
	ringGoToNode  func()
	closeDoorbell func()
	unmap         func() error
}

func mapRpc(RpcMapping) (rpcMapping, error) {
	return rpcMapping{}, errors.New("shared-memory RPC unavailable on this platform")
}
