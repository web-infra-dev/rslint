package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func newTestRpcMapping() (rpcMapping, chan struct{}, chan struct{}, *atomic.Bool) {
	buf := make([]byte, RpcCapacity)
	atomic.StoreUint32((*uint32)(unsafe.Pointer(&buf[rpcGlobalMagicOff])), RpcMagic)
	nodeToGo := make(chan struct{}, 1)
	goToNode := make(chan struct{}, 1)
	done := make(chan struct{})
	var closeOnce sync.Once
	var unmapped atomic.Bool

	mapping := rpcMapping{
		data: buf,
		waitDoorbell: func() error {
			select {
			case <-nodeToGo:
				return nil
			case <-done:
				return io.EOF
			}
		},
		ringGoToNode: func() {
			select {
			case goToNode <- struct{}{}:
			default:
			}
		},
		closeDoorbell: func() {
			closeOnce.Do(func() { close(done) })
		},
		unmap: func() error {
			closeOnce.Do(func() { close(done) })
			unmapped.Store(true)
			return nil
		},
	}
	return mapping, nodeToGo, goToNode, &unmapped
}

func simCall(t *testing.T, srv *rpcServer, nodeToGo, goToNode chan struct{}, slot int, seq uint32, req []byte) (uint32, []byte) {
	t.Helper()
	copy(srv.slotPayload(slot), req)
	atomic.StoreUint32(srv.slotWord(slot, rpcSlotReqSeqOff), seq)
	atomic.StoreUint32(srv.slotWord(slot, rpcSlotWorkerEpochOff), 1)
	atomic.StoreUint32(srv.slotWord(slot, rpcSlotReqLenOff), uint32(len(req)))
	atomic.StoreUint32(srv.slotWord(slot, rpcSlotClientSleepingOff), 1)
	atomic.StoreUint32(srv.slotWord(slot, rpcSlotStateOff), rpcStateReqReady)
	if atomic.LoadUint32(srv.globalWord(rpcGlobalServerSleepingOff)) != 0 {
		select {
		case nodeToGo <- struct{}{}:
		default:
		}
	}

	var out []byte
	deadline := time.After(5 * time.Second)
	for {
		st := atomic.LoadUint32(srv.slotWord(slot, rpcSlotStateOff))
		switch st {
		case rpcStateRespChunk:
			n := int(atomic.LoadUint32(srv.slotWord(slot, rpcSlotRespLenOff)))
			out = append(out, srv.slotPayload(slot)[:n]...)
			atomic.StoreUint32(srv.slotWord(slot, rpcSlotStateOff), rpcStateChunkAck)
			select {
			case nodeToGo <- struct{}{}:
			default:
			}
		case rpcStateRespReady:
			n := int(atomic.LoadUint32(srv.slotWord(slot, rpcSlotRespLenOff)))
			status := atomic.LoadUint32(srv.slotWord(slot, rpcSlotRespStatusOff))
			if n > 0 {
				out = append(out, srv.slotPayload(slot)[:n]...)
			}
			atomic.StoreUint32(srv.slotWord(slot, rpcSlotClientSleepingOff), 0)
			atomic.StoreUint32(srv.slotWord(slot, rpcSlotStateOff), rpcStateIdle)
			return status, out
		default:
			select {
			case <-goToNode:
			case <-time.After(1 * time.Millisecond):
			case <-deadline:
				t.Fatalf("timed out waiting for slot %d response (state=%d)", slot, st)
			}
		}
	}
}

func TestRpcServerSingleChunkedPanicAbandonAndClose(t *testing.T) {
	mapping, nodeToGo, goToNode, unmapped := newTestRpcMapping()
	inHandler := make(chan struct{})
	releaseHandler := make(chan struct{})

	srv := newRpcServer(context.Background(), mapping, func(ctx context.Context, reqs []SyncRequest, reply SyncReplyFunc) {
		for _, r := range reqs {
			switch string(r.Payload) {
			case "echo":
				reply(r.Slot, r.ReqSeq, SyncStatusOK, []byte("pong"))
			case "big":
				big := bytes.Repeat([]byte{0xAB}, RpcSlotSize*2+1234)
				reply(r.Slot, r.ReqSeq, SyncStatusOK, big)
			case "panic":
				panic("boom")
			case "block":
				close(inHandler)
				select {
				case <-releaseHandler:
				case <-ctx.Done():
				}
				reply(r.Slot, r.ReqSeq, SyncStatusOK, []byte("late"))
			}
		}
	})
	srv.start()

	// 1. Single-slot reply
	st, body := simCall(t, srv, nodeToGo, goToNode, 0, 1, []byte("echo"))
	if st != SyncStatusOK || string(body) != "pong" {
		t.Fatalf("expected OK pong, got status=%d body=%q", st, body)
	}

	// 2. Multi-chunk reply (> 2 slots) twice in a row on same slot to verify chunkWaiting reset
	expectedBig := bytes.Repeat([]byte{0xAB}, RpcSlotSize*2+1234)
	for seq := uint32(2); seq <= 3; seq++ {
		st, body = simCall(t, srv, nodeToGo, goToNode, 1, seq, []byte("big"))
		if st != SyncStatusOK || !bytes.Equal(body, expectedBig) {
			t.Fatalf("seq=%d: expected %d-byte chunked reply, got status=%d len=%d", seq, len(expectedBig), st, len(body))
		}
	}

	// 3. Handler panic surfaces as SyncStatusError and does not crash server
	st, body = simCall(t, srv, nodeToGo, goToNode, 2, 4, []byte("panic"))
	if st != SyncStatusError || !bytes.Contains(body, []byte("boom")) {
		t.Fatalf("expected error containing boom, got status=%d body=%q", st, body)
	}

	// 4. Close waits for active handler before releasing shared memory
	copy(srv.slotPayload(3), []byte("block"))
	atomic.StoreUint32(srv.slotWord(3, rpcSlotReqSeqOff), 4)
	atomic.StoreUint32(srv.slotWord(3, rpcSlotReqLenOff), 5)
	atomic.StoreUint32(srv.slotWord(3, rpcSlotStateOff), rpcStateReqReady)
	select {
	case nodeToGo <- struct{}{}:
	default:
	}
	<-inHandler

	// Abandon slot 3 while IN_FLIGHT, then close server; server must wait and unmap cleanly.
	if !atomic.CompareAndSwapUint32(srv.slotWord(3, rpcSlotStateOff), rpcStateInFlight, rpcStateAbandoned) {
		t.Fatalf("expected slot 3 to be IN_FLIGHT")
	}
	closeDone := make(chan struct{})
	go func() {
		srv.close()
		close(closeDone)
	}()
	// Wait until close completes (ctx cancellation unblocks handler).
	select {
	case <-closeDone:
	case <-time.After(3 * time.Second):
		t.Fatal("srv.close() hung")
	}
	if !unmapped.Load() {
		t.Fatal("expected unmap to be called on close")
	}
	if atomic.LoadUint32(srv.slotWord(3, rpcSlotStateOff)) != rpcStateIdle {
		t.Fatalf("expected abandoned slot to return to IDLE, got %d", atomic.LoadUint32(srv.slotWord(3, rpcSlotStateOff)))
	}
	if atomic.LoadUint32(srv.globalWord(rpcGlobalServerClosedOff)) != 1 {
		t.Fatal("expected server_closed=1 after close")
	}
}

func TestTransportConfigAdvertisesRpcOnlyWhenSyncHandlerSet(t *testing.T) {
	check := func(withHandler bool) MemoryConfiguration {
		r1, w1 := io.Pipe()
		r2, w2 := io.Pipe()
		defer r1.Close()
		defer w1.Close()
		defer r2.Close()
		defer w2.Close()

		server := NewChannel(r1, w2)
		if withHandler {
			server.SetSyncHandler(func(ctx context.Context, reqs []SyncRequest, reply SyncReplyFunc) {})
		}
		server.Start()
		defer server.Close()

		client := NewChannel(r2, w1)
		client.Start()
		defer client.Close()

		resp, err := client.SendRequest(context.Background(), KindTransportConfig, nil)
		if err != nil {
			t.Fatalf("transportConfig failed: %v", err)
		}
		var cfg MemoryConfiguration
		if err := json.Unmarshal(resp.Data, &cfg); err != nil {
			t.Fatalf("unmarshal config: %v", err)
		}
		return cfg
	}

	without := check(false)
	if without.Rpc != nil {
		t.Fatalf("expected rpc=nil without SyncHandler, got %+v", without.Rpc)
	}
	with := check(true)
	if with.Rpc == nil || with.Rpc.SlotCount != RpcSlotCount || with.Rpc.SlotSize != RpcSlotSize {
		t.Fatalf("expected advertised rpc config, got %+v", with.Rpc)
	}
}
