//go:build windows

package ipc

import (
	"errors"
	"io"
	"strconv"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

type rpcMapping struct {
	data          []byte
	waitDoorbell  func() error
	ringGoToNode  func()
	closeDoorbell func()
	unmap         func() error
}

func parseWinHandle(raw string) (windows.Handle, error) {
	val, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || val == 0 || uint64(uintptr(val)) != val {
		return 0, errors.New("invalid Windows handle")
	}
	return windows.Handle(val), nil
}

func mapRpc(descriptor RpcMapping) (rpcMapping, error) {
	if descriptor.Version != RpcVersion ||
		descriptor.FD != 0 ||
		descriptor.DoorbellFD != 0 ||
		descriptor.ProcessID == 0 {
		return rpcMapping{}, errors.New("invalid RPC shared memory descriptor")
	}
	shmSrc, err := parseWinHandle(descriptor.Handle)
	if err != nil {
		return rpcMapping{}, err
	}
	nodeToGoSrc, err := parseWinHandle(descriptor.NodeToGoEvent)
	if err != nil {
		return rpcMapping{}, err
	}
	goToNodeSrc, err := parseWinHandle(descriptor.GoToNodeEvent)
	if err != nil {
		return rpcMapping{}, err
	}

	// Keep the parent process handle open with SYNCHRONIZE so abrupt parent death
	// wakes WaitForMultipleObjects immediately just like Unix socketpair EOF.
	process, err := windows.OpenProcess(windows.PROCESS_DUP_HANDLE|windows.SYNCHRONIZE, false, descriptor.ProcessID)
	if err != nil {
		return rpcMapping{}, err
	}
	var shmHandle, nodeToGo, goToNode, closeEvent windows.Handle
	var address uintptr
	success := false
	defer func() {
		if success {
			return
		}
		if address != 0 {
			_ = windows.UnmapViewOfFile(address)
		}
		if closeEvent != 0 {
			_ = windows.CloseHandle(closeEvent)
		}
		if goToNode != 0 {
			_ = windows.CloseHandle(goToNode)
		}
		if nodeToGo != 0 {
			_ = windows.CloseHandle(nodeToGo)
		}
		_ = windows.CloseHandle(process)
	}()

	self := windows.CurrentProcess()
	if err := windows.DuplicateHandle(process, shmSrc, self, &shmHandle, windows.FILE_MAP_WRITE, false, 0); err != nil {
		return rpcMapping{}, err
	}
	defer windows.CloseHandle(shmHandle)

	if err := windows.DuplicateHandle(process, nodeToGoSrc, self, &nodeToGo, windows.SYNCHRONIZE|windows.EVENT_MODIFY_STATE, false, 0); err != nil {
		return rpcMapping{}, err
	}
	if err := windows.DuplicateHandle(process, goToNodeSrc, self, &goToNode, windows.EVENT_MODIFY_STATE, false, 0); err != nil {
		return rpcMapping{}, err
	}

	closeEvent, err = windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return rpcMapping{}, err
	}

	address, err = windows.MapViewOfFile(shmHandle, windows.FILE_MAP_WRITE, 0, 0, RpcCapacity)
	if err != nil {
		return rpcMapping{}, err
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(address)), RpcCapacity)
	if atomic.LoadUint32((*uint32)(unsafe.Pointer(&data[rpcGlobalMagicOff]))) != RpcMagic {
		return rpcMapping{}, errors.New("invalid RPC shared memory magic")
	}

	success = true
	waitHandles := []windows.Handle{nodeToGo, closeEvent, process}
	return rpcMapping{
		data: data,
		waitDoorbell: func() error {
			event, err := windows.WaitForMultipleObjects(waitHandles, false, windows.INFINITE)
			if err != nil {
				return err
			}
			if event == windows.WAIT_OBJECT_0 {
				return nil
			}
			return io.EOF
		},
		ringGoToNode: func() {
			_ = windows.SetEvent(goToNode)
		},
		closeDoorbell: func() {
			_ = windows.SetEvent(closeEvent)
		},
		unmap: func() error {
			_ = windows.SetEvent(closeEvent)
			err := windows.UnmapViewOfFile(address)
			_ = windows.CloseHandle(closeEvent)
			_ = windows.CloseHandle(goToNode)
			_ = windows.CloseHandle(nodeToGo)
			_ = windows.CloseHandle(process)
			return err
		},
	}, nil
}
