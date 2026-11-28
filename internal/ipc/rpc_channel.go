package ipc

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Wire response statuses shared with crates/rslint-native/src/rpc_transport.rs.
const (
	SyncStatusOK        uint32 = 0
	SyncStatusClosed    uint32 = 1
	SyncStatusCancelled uint32 = 2
	SyncStatusError     uint32 = 3
)

// Shared-memory header offsets (128-byte cache-line isolated).
const (
	rpcGlobalMagicOff          = 0
	rpcGlobalServerSleepingOff = 4
	rpcGlobalServerClosedOff   = 8
	rpcGlobalClientClosedOff   = 12
	rpcGlobalHeaderSize        = 128

	rpcSlotStateOff          = 0
	rpcSlotReqSeqOff         = 4
	rpcSlotRespSeqOff        = 8
	rpcSlotCancelSeqOff      = 12
	rpcSlotWorkerEpochOff    = 16
	rpcSlotReqLenOff         = 20
	rpcSlotRespLenOff        = 24
	rpcSlotRespStatusOff     = 28
	rpcSlotClientSleepingOff = 32
	rpcSlotTotalRespLenOff   = 36

	rpcStateIdle      uint32 = 0
	rpcStateReqReady  uint32 = 1
	rpcStateInFlight  uint32 = 2
	rpcStateRespReady uint32 = 3
	rpcStateRespChunk uint32 = 4
	rpcStateChunkAck  uint32 = 5
	rpcStateAbandoned uint32 = 6
)

// SyncRequest describes one binary RPC request claimed from a shared-memory slot.
// Payload is a Go-owned copy (never aliasing mapped shared memory).
type SyncRequest struct {
	Slot        int
	ReqSeq      uint32
	WorkerEpoch uint32
	Payload     []byte
	cancelCheck func(slot int, reqSeq uint32) bool
}

// IsCancelled reports whether the caller cancelled or abandoned this request.
func (r SyncRequest) IsCancelled() bool {
	return r.cancelCheck != nil && r.cancelCheck(r.Slot, r.ReqSeq)
}

// SyncReplyFunc publishes the reply for one claimed slot. Must be called at
// most once per SyncRequest; replies larger than RpcSlotSize stream in chunks
// over the same slot without re-executing the handler.
type SyncReplyFunc func(slot int, reqSeq uint32, status uint32, body []byte)

// SyncBatchHandler executes a batch of concurrently ready shared-memory RPC
// requests. Implementations may process independent targets concurrently and
// call reply for each request as soon as its result is ready.
type SyncBatchHandler func(ctx context.Context, reqs []SyncRequest, reply SyncReplyFunc)

type rpcServer struct {
	mapping rpcMapping
	handler SyncBatchHandler

	ctx         context.Context
	cancel      context.CancelFunc
	scannerDone chan struct{}
	wg          sync.WaitGroup
	closeOnce   sync.Once

	// Per-slot Go-owned request scratch buffers. Slot i is exclusively owned
	// by the single handler processing slot i until reply transitions the slot
	// out of IN_FLIGHT/ABANDONED.
	slotScratch   [RpcSlotCount][]byte
	chunkWaiting  [RpcSlotCount]atomic.Bool
	slotChunkWake [RpcSlotCount]chan struct{}
	cancelCheckFn func(slot int, reqSeq uint32) bool
}

func openRpcServer(parentCtx context.Context, descriptor RpcMapping, handler SyncBatchHandler) (*rpcServer, error) {
	mapping, err := mapRpc(descriptor)
	if err != nil {
		return nil, err
	}
	return newRpcServer(parentCtx, mapping, handler), nil
}

func newRpcServer(parentCtx context.Context, mapping rpcMapping, handler SyncBatchHandler) *rpcServer {
	ctx, cancel := context.WithCancel(parentCtx)
	s := &rpcServer{
		mapping:     mapping,
		handler:     handler,
		ctx:         ctx,
		cancel:      cancel,
		scannerDone: make(chan struct{}),
	}
	s.cancelCheckFn = s.isSlotCancelled
	for i := range s.slotChunkWake {
		s.slotChunkWake[i] = make(chan struct{}, 1)
	}
	return s
}

func (s *rpcServer) start() {
	go s.scanLoop()
}

func (s *rpcServer) globalWord(offset int) *uint32 {
	return (*uint32)(unsafe.Pointer(&s.mapping.data[offset]))
}

func (s *rpcServer) slotWord(slot int, fieldOff int) *uint32 {
	off := rpcGlobalHeaderSize + slot*RpcSlotHeaderStride + fieldOff
	return (*uint32)(unsafe.Pointer(&s.mapping.data[off]))
}

func (s *rpcServer) slotPayload(slot int) []byte {
	start := RpcHeaderSize + slot*RpcSlotSize
	return s.mapping.data[start : start+RpcSlotSize]
}

func (s *rpcServer) isSlotCancelled(slot int, reqSeq uint32) bool {
	if s.ctx.Err() != nil {
		return true
	}
	if atomic.LoadUint32(s.slotWord(slot, rpcSlotCancelSeqOff)) == reqSeq {
		return true
	}
	return atomic.LoadUint32(s.slotWord(slot, rpcSlotStateOff)) == rpcStateAbandoned
}

func (s *rpcServer) scanSlots(claimed []SyncRequest) []SyncRequest {
	claimed = claimed[:0]
	for i := range RpcSlotCount {
		stateWord := s.slotWord(i, rpcSlotStateOff)
		st := atomic.LoadUint32(stateWord)
		if st == rpcStateReqReady {
			if atomic.CompareAndSwapUint32(stateWord, rpcStateReqReady, rpcStateInFlight) {
				reqSeq := atomic.LoadUint32(s.slotWord(i, rpcSlotReqSeqOff))
				workerEpoch := atomic.LoadUint32(s.slotWord(i, rpcSlotWorkerEpochOff))
				reqLen := int(atomic.LoadUint32(s.slotWord(i, rpcSlotReqLenOff)))
				if reqLen < 0 || reqLen > RpcSlotSize {
					s.writeReply(i, reqSeq, SyncStatusError, []byte("ipc: invalid RPC request length"))
					continue
				}
				buf := s.slotScratch[i]
				if cap(buf) < reqLen {
					buf = make([]byte, reqLen)
				} else {
					buf = buf[:reqLen]
				}
				if reqLen > 0 {
					copy(buf, s.slotPayload(i)[:reqLen])
				}
				s.slotScratch[i] = buf
				claimed = append(claimed, SyncRequest{
					Slot:        i,
					ReqSeq:      reqSeq,
					WorkerEpoch: workerEpoch,
					Payload:     buf,
					cancelCheck: s.cancelCheckFn,
				})
			}
		} else if s.chunkWaiting[i].Load() && (st == rpcStateChunkAck || st == rpcStateAbandoned) {
			select {
			case s.slotChunkWake[i] <- struct{}{}:
			default:
			}
		}
	}
	return claimed
}

func (s *rpcServer) scanLoop() {
	defer close(s.scannerDone)

	// Only spin immediately after active work on >=4-core machines; 0% CPU while idle
	// and zero starvation on 2-core CI runners.
	var spinBudget time.Duration
	if runtime.GOMAXPROCS(0) >= 4 {
		spinBudget = 200 * time.Microsecond
	}
	serverSleeping := s.globalWord(rpcGlobalServerSleepingOff)
	clientClosed := s.globalWord(rpcGlobalClientClosedOff)

	claimed := make([]SyncRequest, 0, RpcSlotCount)
	lastActive := time.Time{}
	idleSpins := 0

	for {
		if s.ctx.Err() != nil || atomic.LoadUint32(clientClosed) != 0 {
			return
		}

		claimed = s.scanSlots(claimed)
		if len(claimed) > 0 {
			// Execute directly on the scanner goroutine: avoids per-query
			// goroutine scheduling/wake-up thrash and naturally batches concurrent
			// workers into one checker lease. Rare >64 KiB replies offload
			// chunk streaming to a background goroutine so scanLoop stays free
			// to pump slotChunkWake.
			s.runBatch(claimed)
			lastActive = time.Now()
			idleSpins = 0
			continue
		}

		// Tier 1: short adaptive spin immediately after handling work.
		// Poll slots in a tight batch before yielding so runtime.Gosched does
		// not dominate CPU during sub-10us worker gaps.
		if spinBudget > 0 && !lastActive.IsZero() {
			idleSpins++
			if idleSpins&31 == 0 {
				if time.Since(lastActive) >= spinBudget {
					lastActive = time.Time{}
				} else {
					runtime.Gosched()
				}
			}
			if !lastActive.IsZero() {
				continue
			}
		}

		// Tier 2: Dekker sleep (SeqCst store server_sleeping=1, then SeqCst scan).
		atomic.StoreUint32(serverSleeping, 1)
		if s.ctx.Err() != nil || atomic.LoadUint32(clientClosed) != 0 {
			atomic.StoreUint32(serverSleeping, 0)
			return
		}
		claimed = s.scanSlots(claimed)
		if len(claimed) > 0 {
			atomic.StoreUint32(serverSleeping, 0)
			s.runBatch(claimed)
			lastActive = time.Now()
			idleSpins = 0
			continue
		}

		err := s.mapping.waitDoorbell()
		atomic.StoreUint32(serverSleeping, 0)
		if err != nil {
			return
		}
		idleSpins = 0
	}
}

func (s *rpcServer) runBatch(batch []SyncRequest) {
	var repliedMask atomic.Uint64
	var panicErr error

	reply := func(slot int, reqSeq uint32, status uint32, body []byte) {
		if slot < 0 || slot >= RpcSlotCount {
			return
		}
		bit := uint64(1) << uint(slot)
		for {
			cur := repliedMask.Load()
			if cur&bit != 0 {
				return
			}
			if repliedMask.CompareAndSwap(cur, cur|bit) {
				break
			}
		}
		s.writeReply(slot, reqSeq, status, body)
	}

	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				panicErr = fmt.Errorf("ipc: sync RPC handler panicked: %v", recovered)
			}
		}()
		s.handler(s.ctx, batch, reply)
	}()

	// Guarantee every claimed slot transitions cleanly out of IN_FLIGHT even if
	// the handler panicked or omitted a reply.
	mask := repliedMask.Load()
	for _, req := range batch {
		bit := uint64(1) << uint(req.Slot)
		if mask&bit != 0 {
			continue
		}
		if panicErr != nil {
			s.writeReply(req.Slot, req.ReqSeq, SyncStatusError, []byte(panicErr.Error()))
		} else if s.ctx.Err() != nil {
			s.writeReply(req.Slot, req.ReqSeq, SyncStatusCancelled, nil)
		} else {
			s.writeReply(req.Slot, req.ReqSeq, SyncStatusError, []byte("ipc: sync RPC handler omitted reply"))
		}
	}
}

func (s *rpcServer) writeReply(slot int, reqSeq uint32, status uint32, body []byte) {
	stateWord := s.slotWord(slot, rpcSlotStateOff)
	respSeqWord := s.slotWord(slot, rpcSlotRespSeqOff)
	respLenWord := s.slotWord(slot, rpcSlotRespLenOff)
	respStatusWord := s.slotWord(slot, rpcSlotRespStatusOff)
	totalLenWord := s.slotWord(slot, rpcSlotTotalRespLenOff)
	clientSleepingWord := s.slotWord(slot, rpcSlotClientSleepingOff)
	payload := s.slotPayload(slot)

	// Fast single-slot path (covers >99.9% of queries with zero chunk overhead).
	if len(body) <= RpcSlotSize {
		if len(body) > 0 {
			copy(payload[:len(body)], body)
		}
		atomic.StoreUint32(respSeqWord, reqSeq)
		atomic.StoreUint32(respLenWord, uint32(len(body)))
		atomic.StoreUint32(totalLenWord, uint32(len(body)))
		atomic.StoreUint32(respStatusWord, status)
		if atomic.CompareAndSwapUint32(stateWord, rpcStateInFlight, rpcStateRespReady) {
			if atomic.LoadUint32(clientSleepingWord) != 0 {
				s.mapping.ringGoToNode()
			}
		} else {
			// Worker abandoned slot mid-flight; release slot cleanly back to IDLE via CAS.
			atomic.CompareAndSwapUint32(stateWord, rpcStateAbandoned, rpcStateIdle)
		}
		return
	}

	// Multi-chunk streaming path for replies > RpcSlotSize (zero handler re-execution).
	// Offload to a tracked goroutine so scanLoop remains free to pump slotChunkWake
	// and service other workers concurrently.
	owned := append([]byte(nil), body...)
	s.wg.Add(1)
	go s.writeChunkedReply(slot, reqSeq, status, owned)
}

func (s *rpcServer) writeChunkedReply(slot int, reqSeq uint32, status uint32, body []byte) {
	defer s.wg.Done()
	stateWord := s.slotWord(slot, rpcSlotStateOff)
	respSeqWord := s.slotWord(slot, rpcSlotRespSeqOff)
	respLenWord := s.slotWord(slot, rpcSlotRespLenOff)
	respStatusWord := s.slotWord(slot, rpcSlotRespStatusOff)
	totalLenWord := s.slotWord(slot, rpcSlotTotalRespLenOff)
	clientSleepingWord := s.slotWord(slot, rpcSlotClientSleepingOff)
	payload := s.slotPayload(slot)

	total := len(body)
	expectedState := rpcStateInFlight
	for off := 0; off < total; {
		end := min(off+RpcSlotSize, total)
		chunk := body[off:end]
		copy(payload[:len(chunk)], chunk)
		atomic.StoreUint32(respSeqWord, reqSeq)
		atomic.StoreUint32(respLenWord, uint32(len(chunk)))
		atomic.StoreUint32(totalLenWord, uint32(total))
		atomic.StoreUint32(respStatusWord, status)

		isFinal := end >= total
		nextState := rpcStateRespChunk
		if isFinal {
			nextState = rpcStateRespReady
		} else {
			// Drain any stale wake token and arm chunkWaiting strictly BEFORE
			// publishing RESP_CHUNK so scanSlots pumps slotChunkWake immediately,
			// and ALWAYS clear chunkWaiting BEFORE releasing the slot.
			select {
			case <-s.slotChunkWake[slot]:
			default:
			}
			s.chunkWaiting[slot].Store(true)
		}
		if !atomic.CompareAndSwapUint32(stateWord, expectedState, nextState) {
			if !isFinal {
				s.chunkWaiting[slot].Store(false)
			}
			atomic.CompareAndSwapUint32(stateWord, rpcStateAbandoned, rpcStateIdle)
			return
		}
		if atomic.LoadUint32(clientSleepingWord) != 0 {
			s.mapping.ringGoToNode()
		}
		if isFinal {
			return
		}
		acked := s.waitForChunkAck(slot, stateWord)
		s.chunkWaiting[slot].Store(false)
		if !acked {
			if !atomic.CompareAndSwapUint32(stateWord, rpcStateAbandoned, rpcStateIdle) {
				atomic.CompareAndSwapUint32(stateWord, rpcStateRespChunk, rpcStateIdle)
			}
			return
		}
		off = end
		expectedState = rpcStateChunkAck
	}
}

func (s *rpcServer) waitForChunkAck(slot int, stateWord *uint32) bool {
	clientClosed := s.globalWord(rpcGlobalClientClosedOff)
	deadline := time.Now().Add(25 * time.Microsecond)
	for {
		st := atomic.LoadUint32(stateWord)
		if st == rpcStateChunkAck {
			return true
		}
		if st == rpcStateAbandoned || st == rpcStateIdle || s.ctx.Err() != nil || atomic.LoadUint32(clientClosed) != 0 {
			return false
		}
		if time.Now().After(deadline) {
			break
		}
		runtime.Gosched()
	}

	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()
	for {
		st := atomic.LoadUint32(stateWord)
		if st == rpcStateChunkAck {
			return true
		}
		if st == rpcStateAbandoned || st == rpcStateIdle || s.ctx.Err() != nil || atomic.LoadUint32(clientClosed) != 0 {
			return false
		}
		select {
		case <-s.slotChunkWake[slot]:
		case <-s.ctx.Done():
			return false
		case <-ticker.C:
		}
	}
}

// close terminates the scanner and waits for every in-flight handler goroutine
// before releasing shared memory.
func (s *rpcServer) close() {
	s.closeOnce.Do(func() {
		s.cancel()
		atomic.StoreUint32(s.globalWord(rpcGlobalServerClosedOff), 1)
		s.mapping.ringGoToNode()
		s.mapping.closeDoorbell()
		<-s.scannerDone
		s.wg.Wait()
		_ = s.mapping.unmap()
	})
}
