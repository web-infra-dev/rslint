// Package api / Bidirectional service.
//
// `Service` (in api.go) is server-only: a single goroutine reads frames and
// dispatches synchronously to a `Handler`. That model is used by `--api` mode
// (consumed by `packages/rslint-wasm` and `packages/rslint-api`) and is left
// alone here for backward compatibility.
//
// `BidirectionalService` is a parallel implementation that supports both
// inbound requests (Node → Go) AND outbound requests (Go → Node), with
// notifications in either direction. It powers the CLI runtime where
// the Node parent (engine.ts) drives the Go child binary over stdio IPC.
// The LSP path no longer uses this service — Go LSP server talks to its
// client (the VS Code extension) over the standard LSP JSON-RPC channel
// directly via custom methods like `rslint/lintCompatBatch`.
//
// Architecture:
//
//	┌─ reader goroutine ────────────────────────────────────┐
//	│  decode one frame                                     │
//	│  ├─ kind == response/error  → route to pending[reqID] │
//	│  └─ otherwise               → spawn handler goroutine │
//	└───────────────────────────────────────────────────────┘
//
//	┌─ writer goroutine ────────────────────────────────────┐
//	│  drain frames from a buffered channel and Flush()     │
//	└───────────────────────────────────────────────────────┘
//
//	┌─ outbound client (any goroutine) ─────────────────────┐
//	│  SendRequest(kind, data) → reqID, channel             │
//	│  enqueue frame to writer; select on reply / ctx       │
//	└───────────────────────────────────────────────────────┘
//
// Why two goroutines, not one shared loop:
//
//	The handler goroutine for an inbound request may itself call SendRequest
//	to ask the peer to do work. The peer's response arrives as a regular
//	frame on the same pipe. If the same goroutine that handles requests is
//	also the one reading the pipe, the SendRequest waiter would deadlock
//	waiting for a reply that the blocked goroutine could never read.
//	Splitting reader from handler removes that hazard.
//
// Why a buffered writer channel:
//
//	OS pipes have small kernel buffers (Linux default ~64 KiB). A blocking
//	write would stall whichever goroutine attempted it. The writer goroutine
//	owns the pipe; senders enqueue to a Go channel and proceed. Backpressure
//	is now between the senders and the channel buffer, never against the
//	pipe. (Spike 1 confirmed 200 KiB single frames pass cleanly.)
package ipc

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

// InboundHandler is the bidirectional-mode equivalent of Handler. It is
// invoked once per inbound request (Node → Go), in its own goroutine. It
// returns either a payload (which the runtime wraps as a `response` frame
// to the same reqID) or an error (wrapped as `error`).
//
// Multiple inbound requests may be in flight concurrently — the handler
// implementation is responsible for its own thread safety.
type InboundHandler interface {
	// Handle receives a fully decoded inbound request. The framework guarantees
	// msg.Kind is not "response" / "error" / a registered notification kind
	// (those are dispatched separately).
	//
	// Returning (data, nil) → the runtime sends a `response` frame with `data`
	// as payload. Returning (_, err) → the runtime sends an `error` frame.
	Handle(ctx context.Context, msg *Message) (interface{}, error)
}

// NotificationHandler is invoked for inbound notifications (id == 0,
// no reply expected). Errors returned are logged but not sent back over
// the wire.
type NotificationHandler func(ctx context.Context, msg *Message)

// BidirectionalService is the runtime for the new IPC model. Construct via
// NewBidirectionalService, then call Start.
//
// Lifecycle:
//
//	bs := NewBidirectionalService(stdin, stdout)
//	bs.SetInboundHandler(myHandler)         // optional
//	bs.RegisterNotification("log", myLog)   // optional
//	bs.Start()                              // spawns reader + writer goroutines
//	defer bs.Close()
//
//	resp, err := bs.SendRequest(ctx, "lintEslintPlugin", payload)
//	bs.SendNotification("output", payload)
//
// Close drains in-flight outbound requests (cancels their contexts via
// the closed stopCh) and waits for both goroutines to exit.
type BidirectionalService struct {
	r *bufio.Reader
	w *bufio.Writer

	writeCh chan []byte
	stopCh  chan struct{}

	pendingMu sync.Mutex
	pending   map[int64]chan *Message

	nextID atomic.Int64

	// handler / notification dispatch is set once before Start; reads after
	// Start are safe because we only swap pointers atomically. We use a
	// mutex anyway because changes are infrequent and clarity > micro-perf.
	dispatchMu       sync.RWMutex
	inboundHandler   InboundHandler
	notificationDisp map[MessageKind]NotificationHandler

	wg sync.WaitGroup

	// closed is set by Close() (and observed by senders) so SendRequest
	// after Close fails fast instead of hanging forever.
	closed atomic.Bool

	// serviceCtx is the context passed to inbound request handlers.
	// Cancelled by Close so handlers blocked in long-running work can
	// observe shutdown and abort instead of writing into a torn-down
	// transport. Previously each handler received `context.Background()`,
	// which meant Close + Wait could deadlock or, more commonly, let a
	// slow handler continue file I/O / external IPC long after the
	// caller had given up.
	serviceCtx    context.Context
	cancelService context.CancelFunc
}

// NewBidirectionalService wires the service to the given pipes. The pipes
// are typically os.Stdin / os.Stdout for a child-process IPC link.
func NewBidirectionalService(reader io.Reader, writer io.Writer) *BidirectionalService {
	ctx, cancel := context.WithCancel(context.Background())
	return &BidirectionalService{
		r:                bufio.NewReader(reader),
		w:                bufio.NewWriter(writer),
		writeCh:          make(chan []byte, 64), // small but enough for short bursts; senders block when full (intentional)
		stopCh:           make(chan struct{}),
		pending:          make(map[int64]chan *Message),
		notificationDisp: make(map[MessageKind]NotificationHandler),
		serviceCtx:       ctx,
		cancelService:    cancel,
	}
}

// SetInboundHandler installs the request handler. Set before Start.
// Setting after Start is allowed but races against in-flight dispatches —
// callers should restrict updates to startup.
func (bs *BidirectionalService) SetInboundHandler(h InboundHandler) {
	bs.dispatchMu.Lock()
	bs.inboundHandler = h
	bs.dispatchMu.Unlock()
}

// RegisterNotification installs a handler for a specific notification kind.
// Notifications are inbound messages with id=0 that expect no reply.
// Multiple kinds may be registered; the same kind registered twice
// overwrites the prior registration.
func (bs *BidirectionalService) RegisterNotification(kind MessageKind, h NotificationHandler) {
	bs.dispatchMu.Lock()
	bs.notificationDisp[kind] = h
	bs.dispatchMu.Unlock()
}

// Start spawns the reader and writer goroutines. Returns immediately;
// callers must Close to shut down.
func (bs *BidirectionalService) Start() {
	bs.wg.Add(2)
	go bs.writerLoop()
	go bs.readerLoop()
}

// Close triggers graceful shutdown:
//
//   - Marks the service closed so subsequent SendRequest / SendNotification
//     calls fail fast instead of hanging.
//   - Closes stopCh to wake up the writer goroutine and any senders parked
//     on writeCh.
//   - Calls failAllPending so outstanding outbound requests return an
//     error rather than waiting forever for a response that won't come.
//
// Close is **non-blocking**: it does NOT wait for the reader/writer
// goroutines to exit. The reader is typically blocked in a Read on the
// underlying stream, and the Service does not own the stream — only the
// caller can close the stream and let the reader unwind.
//
// Lifecycle pattern for a clean teardown:
//
//	bs.Close()         // unblock pending senders, mark closed
//	closeUnderlyingPipe()  // caller closes the input/output streams
//	bs.Wait()          // now the reader has hit EOF, wg drains cleanly
//
// Close is idempotent.
func (bs *BidirectionalService) Close() error {
	if !bs.closed.CompareAndSwap(false, true) {
		return nil
	}
	close(bs.stopCh)
	// Cancel the service-scoped ctx so in-flight inbound handlers can
	// observe shutdown via their ctx parameter and abort. Handlers that
	// ignore the ctx still complete normally; their response write will
	// hit the closed check at the top of SendResponse and fail fast.
	if bs.cancelService != nil {
		bs.cancelService()
	}
	bs.failAllPending()
	return nil
}

// Wait blocks until both the reader and writer goroutines have exited.
// Typically called after Close AND after the underlying input stream has
// been closed by the caller. Calling Wait without closing the input stream
// will hang if the reader is still blocked in Read.
//
// Safe to call multiple times — each call simply waits on the same
// WaitGroup.
func (bs *BidirectionalService) Wait() {
	bs.wg.Wait()
}

// SendRequest issues an outbound request and waits for the matching response.
// Returns the response Message on success, or an error on context expiry,
// service shutdown, or peer-side `error` reply.
//
// Multiple concurrent calls are safe and reqID-multiplexed.
func (bs *BidirectionalService) SendRequest(
	ctx context.Context,
	kind MessageKind,
	data interface{},
) (*Message, error) {
	if bs.closed.Load() {
		return nil, errors.New("api: BidirectionalService closed")
	}

	rawData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("api: marshal request data: %w", err)
	}

	id := bs.nextID.Add(1) // > 0 (notifications use 0)
	frame, err := encodeFrame(Message{Kind: kind, ID: int(id), Data: json.RawMessage(rawData)})
	if err != nil {
		return nil, fmt.Errorf("api: encode request: %w", err)
	}

	respCh := make(chan *Message, 1)
	bs.pendingMu.Lock()
	bs.pending[id] = respCh
	bs.pendingMu.Unlock()

	defer func() {
		bs.pendingMu.Lock()
		delete(bs.pending, id)
		bs.pendingMu.Unlock()
	}()

	// Enqueue to writer. Block on the channel (not the pipe) until the
	// writer drains it or shutdown intervenes.
	select {
	case bs.writeCh <- frame:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-bs.stopCh:
		return nil, errors.New("api: service stopped before send")
	}

	// Wait for the matching response (or error).
	select {
	case resp := <-respCh:
		if resp == nil {
			return nil, errors.New("api: nil response (peer closed)")
		}
		if resp.Kind == KindError {
			var er ErrorResponse
			if err := DecodeData(resp.Data, &er); err == nil {
				return resp, fmt.Errorf("api: peer error: %s", er.Message)
			}
			return resp, errors.New("api: peer error (undecodable)")
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-bs.stopCh:
		return nil, errors.New("api: service stopped while awaiting response")
	}
}

// SendNotification fires a unidirectional message (id=0). Returns when the
// frame has been enqueued to the writer (or shutdown intervened).
func (bs *BidirectionalService) SendNotification(kind MessageKind, data interface{}) error {
	if bs.closed.Load() {
		return errors.New("api: BidirectionalService closed")
	}
	rawData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("api: marshal notification data: %w", err)
	}
	frame, err := encodeFrame(Message{Kind: kind, ID: 0, Data: json.RawMessage(rawData)})
	if err != nil {
		return fmt.Errorf("api: encode notification: %w", err)
	}
	select {
	case bs.writeCh <- frame:
		return nil
	case <-bs.stopCh:
		return errors.New("api: service stopped before send")
	}
}

// SendResponse is invoked by the framework after an InboundHandler returns
// successfully. It is exported because tests / advanced users may want to
// reply asynchronously from outside the handler goroutine.
func (bs *BidirectionalService) SendResponse(reqID int, data interface{}) error {
	// Closed-check kept consistent with SendNotification / SendRequest:
	// once Close() has flipped the flag, the writer is racing teardown
	// and any frame we push here may be silently dropped by the drain
	// path. Returning early surfaces the problem to the caller (most
	// handlers can decide whether to log or just give up).
	if bs.closed.Load() {
		return errors.New("api: service closed")
	}
	rawData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("api: marshal response data: %w", err)
	}
	frame, err := encodeFrame(Message{Kind: KindResponse, ID: reqID, Data: json.RawMessage(rawData)})
	if err != nil {
		return fmt.Errorf("api: encode response: %w", err)
	}
	select {
	case bs.writeCh <- frame:
		return nil
	case <-bs.stopCh:
		return errors.New("api: service stopped before send")
	}
}

// SendErrorResponse delivers an `error` frame matching the given reqID.
func (bs *BidirectionalService) SendErrorResponse(reqID int, errMsg string) error {
	if bs.closed.Load() {
		return errors.New("api: service closed")
	}
	frame, err := encodeFrame(Message{Kind: KindError, ID: reqID, Data: ErrorResponse{Message: errMsg}})
	if err != nil {
		return fmt.Errorf("api: encode error response: %w", err)
	}
	select {
	case bs.writeCh <- frame:
		return nil
	case <-bs.stopCh:
		return errors.New("api: service stopped before send")
	}
}

// readerLoop owns the input pipe.
func (bs *BidirectionalService) readerLoop() {
	defer bs.wg.Done()
	for {
		// Quick stop check between frames; the I/O below blocks on the pipe.
		select {
		case <-bs.stopCh:
			bs.failAllPending()
			return
		default:
		}

		msg, err := readFrame(bs.r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				bs.failAllPending()
				// Peer closed the input stream. Trigger Close so the
				// writer goroutine exits and any caller waiting on
				// Wait() / a derived "done" signal unblocks instead of
				// hanging on the next outbound frame. Idempotent — if
				// Close was already called, this is a no-op.
				_ = bs.Close()
				return
			}
			// During shutdown, the underlying pipe may be closed by the
			// caller mid-Read; that surfaces as "read/write on closed
			// pipe" or similar. Suppressing the noise when we know we
			// are intentionally shutting down keeps logs tidy.
			if !bs.closed.Load() {
				fmt.Fprintf(stderrW(), "api: read error: %v\n", err)
			}
			bs.failAllPending()
			// Same rationale as above for non-EOF read errors: surface
			// the disconnect to senders / waiters via Close.
			_ = bs.Close()
			return
		}

		switch msg.Kind {
		case KindResponse, KindError:
			bs.routeResponse(msg)
		default:
			bs.dispatchInbound(msg)
		}
	}
}

// writerLoop owns the output pipe. On any write / flush error, the
// loop calls Close() so that:
//
//  1. bs.closed flips to true, gating subsequent senders (they get a
//     clean error instead of enqueueing into writeCh's buffer where
//     nobody will dequeue);
//  2. failAllPending fires, unblocking any SendRequest currently
//     waiting on respCh — without this they'd hang until ctx.Done() or
//     forever (if no ctx);
//  3. the reader loop's stopCh select wakes up, so the whole service
//     unwinds symmetrically with how readerLoop handles its own EOF.
//
// The previous behavior was to log and silently return, leaving the
// service in a half-alive state — `bs.closed` still false, pending
// requests still parked.
func (bs *BidirectionalService) writerLoop() {
	defer bs.wg.Done()
	for {
		select {
		case frame := <-bs.writeCh:
			if _, err := bs.w.Write(frame); err != nil {
				if !bs.closed.Load() {
					fmt.Fprintf(stderrW(), "api: write error: %v\n", err)
				}
				_ = bs.Close()
				return
			}
			// Flush every frame so the peer sees individual messages.
			if err := bs.w.Flush(); err != nil {
				if !bs.closed.Load() {
					fmt.Fprintf(stderrW(), "api: flush error: %v\n", err)
				}
				_ = bs.Close()
				return
			}
		case <-bs.stopCh:
			// Drain anything already enqueued for cleanliness, then exit.
			for {
				select {
				case frame := <-bs.writeCh:
					_, _ = bs.w.Write(frame)
				default:
					_ = bs.w.Flush()
					return
				}
			}
		}
	}
}

// routeResponse delivers a response/error frame to the matching pending waiter.
// If no waiter is registered (e.g. the request timed out and the entry was
// removed), the response is dropped with a stderr note.
func (bs *BidirectionalService) routeResponse(msg *Message) {
	id := int64(msg.ID)
	bs.pendingMu.Lock()
	ch, ok := bs.pending[id]
	if ok {
		// Claim the entry under the lock so a concurrent failAllPending
		// (shutdown / EOF) can't ALSO fill this channel with its nil
		// sentinel after we unlock — that race dropped the real response
		// (channel already full) and surfaced a false "peer closed" nil to
		// the waiter. The waiter's own deferred delete is then a no-op.
		delete(bs.pending, id)
	}
	bs.pendingMu.Unlock()
	if !ok {
		fmt.Fprintf(stderrW(), "api: orphan response id=%d kind=%s\n", msg.ID, msg.Kind)
		return
	}
	// Non-blocking send because the channel is buffered to 1 and the
	// waiter creates exactly one entry per reqID.
	select {
	case ch <- msg:
	default:
		fmt.Fprintf(stderrW(), "api: response channel full id=%d (waiter gone?)\n", msg.ID)
	}
}

// dispatchInbound routes a non-response frame to either the InboundHandler
// (request, id != 0) or a NotificationHandler (notification, id == 0).
//
// Each inbound request gets its own goroutine so that:
//
//   - The reader is never blocked by handler work (responses come back via
//     the same pipe; the reader must keep reading).
//   - Multiple in-flight inbound requests are allowed.
//   - The handler may itself call SendRequest without deadlock.
func (bs *BidirectionalService) dispatchInbound(msg *Message) {
	bs.dispatchMu.RLock()
	handler := bs.inboundHandler
	notification := bs.notificationDisp[msg.Kind]
	bs.dispatchMu.RUnlock()

	if msg.ID == 0 {
		// Notification path — no reply. Dispatch SYNCHRONOUSLY in the
		// reader goroutine to preserve wire order. The previous
		// goroutine-per-notification dispatch silently reordered
		// streaming outputs (e.g. lint diagnostics arriving via the
		// `output` notification) when handler runtime varied across
		// goroutines. Request dispatch stays async (line below) because
		// requests are id-multiplexed and reentrant; notifications are
		// unidirectional and order-bearing, so synchronous dispatch is
		// the correct trade. If a future notification handler must do
		// slow work, it's the handler's job to queue internally — the
		// transport must not silently reorder frames.
		if notification == nil {
			fmt.Fprintf(stderrW(), "api: unhandled notification kind=%s\n", msg.Kind)
			return
		}
		defer recoverAndLog(string(msg.Kind))
		notification(context.Background(), msg)
		return
	}

	// Request path.
	if handler == nil {
		_ = bs.SendErrorResponse(msg.ID, fmt.Sprintf("no inbound handler registered (kind=%s)", msg.Kind))
		return
	}
	// Track the handler goroutine in bs.wg so Wait() actually waits
	// for it. Previously each handler was a `go func() {...}()` with
	// no wg ticket, meaning Close + Wait could return while a slow
	// handler was still running — a goroutine leak from the caller's
	// perspective and a "handler still writes to the closed service"
	// hazard (writes silently fail, but the handler's downstream work
	// — file I/O, external IPC, etc. — continues).
	bs.wg.Add(1)
	go func() {
		defer bs.wg.Done()
		defer recoverAndLog(string(msg.Kind))
		// Service-scoped ctx — cancelled by Close() so a long-running
		// handler observes shutdown promptly instead of writing into
		// a torn-down transport. Per-request cancel (LSP-style
		// `$/cancelRequest`) is not wired yet; when it is, derive a
		// child ctx here and register the cancelFunc by msg.ID.
		resp, err := handler.Handle(bs.serviceCtx, msg)
		if err != nil {
			_ = bs.SendErrorResponse(msg.ID, err.Error())
			return
		}
		_ = bs.SendResponse(msg.ID, resp)
	}()
}

// failAllPending unblocks every outstanding SendRequest waiter (sending a
// nil sentinel the waiter converts into an error). Used during shutdown /
// EOF / pipe error so callers don't hang forever.
func (bs *BidirectionalService) failAllPending() {
	bs.pendingMu.Lock()
	defer bs.pendingMu.Unlock()
	for id, ch := range bs.pending {
		// Sentinel: nil msg signals "service down" to the waiter (which
		// converts it into an error). We can't synthesize a real Message
		// here without committing to a kind/id contract; nil is the
		// cleanest "channel closed-equivalent" signal.
		select {
		case ch <- nil:
		default:
		}
		delete(bs.pending, id)
	}
}

// ─── helpers ───────────────────────────────────────────────────────────

// encodeFrame produces the wire format `[4 bytes u32 LE length][JSON payload]`
// shared with internal/api.Service. Centralized here so any future protocol
// tweak is a single edit.
func encodeFrame(msg Message) ([]byte, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	frame := make([]byte, 4+len(body))
	binary.LittleEndian.PutUint32(frame[:4], uint32(len(body)))
	copy(frame[4:], body)
	return frame, nil
}

// maxFrameSize caps a single IPC frame's payload at 256 MiB. The peer is a
// trusted internal Node child, but stream desynchronization (e.g. a partial
// write before crash, a malformed test fixture, a non-IPC byte stream
// mistakenly fed in) would otherwise be interpreted as a giant length and
// trigger `make([]byte, 4 GiB)` → instant OOM. 256 MiB is far beyond any
// realistic single-batch lint payload yet bounded.
const maxFrameSize = 256 * 1024 * 1024

// readFrame reads one length-prefixed JSON frame from r. Returns a protocol
// error (rather than OOMing) if the header advertises a body larger than
// maxFrameSize; caller treats it as fatal and closes the connection.
func readFrame(r *bufio.Reader) (*Message, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(lenBuf[:])
	if length > maxFrameSize {
		return nil, fmt.Errorf("api: frame length %d exceeds cap %d (likely stream desync)", length, maxFrameSize)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("unmarshal frame: %w", err)
	}
	return &msg, nil
}

// DecodeData re-decodes a Message.Data field (which is interface{} — already
// decoded by json into map/string/etc.) into a typed Go value. Exported so
// the cmd/rslint IPC client reuses it instead of keeping a byte-identical
// private copy.
func DecodeData(data interface{}, target interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

// recoverAndLog catches panics in handler goroutines so a buggy handler
// doesn't crash the whole IPC service. The kind is logged for diagnosis.
func recoverAndLog(kind string) {
	if r := recover(); r != nil {
		fmt.Fprintf(stderrW(), "api: handler for kind=%s panicked: %v\n", kind, r)
	}
}

// stderrW returns the writer used by this package's internal diagnostics.
// Wrapped in a function so tests can override it (assign to overrideStderrW from a test
// init or helper). Default is os.Stderr.
var overrideStderrW io.Writer

func stderrW() io.Writer {
	if overrideStderrW != nil {
		return overrideStderrW
	}
	return os.Stderr
}
