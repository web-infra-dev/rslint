package ipc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// InboundHandler handles an inbound request frame (id > 0, not a
// response/error). Its return value is marshaled into a `response` frame;
// a returned error becomes an `error` frame. The ctx is cancelled when the
// channel closes.
type InboundHandler func(ctx context.Context, msg *Message) (any, error)

// ResponseReceipt is completed by Channel after an inbound response reaches a
// terminal write outcome. A nil Err means the complete intended response frame
// was accepted by the underlying writer. Call Err only after Done is closed.
//
// The receipt is framework-owned: completing it never calls application code,
// so Channel may safely publish the successful write commit point while holding
// its write lock.
type ResponseReceipt struct {
	done chan struct{}
	once sync.Once
	err  error
}

// Done closes exactly once when the response write succeeds or cannot succeed.
func (r *ResponseReceipt) Done() <-chan struct{} { return r.done }

// Err returns the response write outcome. It must be called after Done closes.
func (r *ResponseReceipt) Err() error { return r.err }

func (r *ResponseReceipt) complete(err error) {
	r.once.Do(func() {
		r.err = err
		close(r.done)
	})
}

type trackedResponse struct {
	payload  any
	receipt  *ResponseReceipt
	terminal bool
}

// TrackResponse wraps an inbound handler result and returns a receipt for its
// response write. The wire payload is unchanged; only Channel sees the wrapper.
func TrackResponse(payload any) (any, *ResponseReceipt) {
	return newTrackedResponse(payload, false)
}

// TrackTerminalResponse is TrackResponse plus a write-side seal: after the
// complete response frame is accepted, Channel rejects every later frame. This
// is used by terminal protocol requests whose acknowledgement must be the last
// outbound frame.
func TrackTerminalResponse(payload any) (any, *ResponseReceipt) {
	return newTrackedResponse(payload, true)
}

func newTrackedResponse(payload any, terminal bool) (any, *ResponseReceipt) {
	receipt := &ResponseReceipt{done: make(chan struct{})}
	return &trackedResponse{
		payload:  payload,
		receipt:  receipt,
		terminal: terminal,
	}, receipt
}

// NotificationHandler handles an inbound notification (id == 0). No reply.
type NotificationHandler func(msg *Message)

// Channel is a task-agnostic bidirectional length-prefixed-JSON RPC channel
// over a reader/writer pair (typically a child process's stdout/stdin). It
// is the Go counterpart to the Node IpcClient:
//
//   - SendRequest: reqID-multiplexed request → awaits its response/error.
//   - SendNotification: fire-and-forget (id 0).
//   - inbound requests/notifications dispatch to registered handlers.
//
// Handlers must be registered (SetInboundHandler / RegisterNotification)
// BEFORE Start — registering after Start panics, since the read loop reads
// the handler tables without locking. Multiple goroutines may call
// SendRequest/SendNotification concurrently; frame writes are serialized.
//
// A write failure or a read fault is terminal: the channel closes and every
// in-flight request fails with a stable error (mirrors the Node side).
type Channel struct {
	reader *bufio.Reader
	writer io.Writer

	// writeTimeout bounds a single frame write (0 = none; defaulted in
	// NewChannel). requestTimeout is the default deadline applied to a
	// SendRequest whose own ctx carries none (0 = none). Both are set before
	// Start and read-only afterward, so they need no lock.
	writeTimeout   time.Duration
	requestTimeout time.Duration

	writeMu sync.Mutex // serializes frame writes across goroutines

	mu       sync.Mutex // guards pending, nextID, closed, closeErr, started
	pending  map[int]chan *Message
	nextID   int
	closed   bool
	closeErr error
	// writeSealed rejects frames that have not yet passed write admission.
	// It is published while writeMu is held on a terminal response or write
	// fault, closing the unlock-before-close race for queued writers.
	writeSealed  bool
	writeSealErr error
	started      bool

	notifyHandlers map[MessageKind]NotificationHandler
	inbound        InboundHandler

	done     chan struct{}      // closed when the channel shuts down
	inCtx    context.Context    // ctx passed to inbound handlers
	inCancel context.CancelFunc // cancels inCtx on close
}

// defaultWriteTimeout bounds a single frame write so a wedged (alive but
// non-draining) peer can't block a write forever. Re-armed per write in
// writeFrame; only applies to writers that support SetWriteDeadline.
const defaultWriteTimeout = 30 * time.Second

var (
	errTerminalResponse = errors.New("ipc: writes sealed after terminal response")
	errWriteNotAdmitted = errors.New("ipc: write not admitted")
)

// NewChannel creates a channel over r (inbound frames) and w (outbound
// frames). Call SetInboundHandler/RegisterNotification before Start.
func NewChannel(r io.Reader, w io.Writer) *Channel {
	ctx, cancel := context.WithCancel(context.Background())
	return &Channel{
		reader:         bufio.NewReader(r),
		writer:         w,
		pending:        make(map[int]chan *Message),
		nextID:         1, // requests use id > 0; notifications use 0
		notifyHandlers: make(map[MessageKind]NotificationHandler),
		done:           make(chan struct{}),
		inCtx:          ctx,
		inCancel:       cancel,
		writeTimeout:   defaultWriteTimeout,
	}
}

// SetInboundHandler installs the handler for inbound request frames. Must be
// called before Start (panics otherwise — the read loop reads c.inbound
// without locking).
func (c *Channel) SetInboundHandler(h InboundHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		panic("ipc: SetInboundHandler must be called before Start")
	}
	c.inbound = h
}

// RegisterNotification registers a handler for inbound notifications of a
// given kind. Must be called before Start (panics otherwise). Registering
// the same kind twice overwrites the prior one.
func (c *Channel) RegisterNotification(kind MessageKind, h NotificationHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		panic("ipc: RegisterNotification must be called before Start")
	}
	c.notifyHandlers[kind] = h
}

// Start launches the reader loop in a goroutine. Returns immediately; the
// loop runs until EOF / transport error / Close.
func (c *Channel) Start() {
	c.mu.Lock()
	c.started = true
	c.mu.Unlock()
	go c.readLoop()
}

// Done returns a channel closed when the transport shuts down.
func (c *Channel) Done() <-chan struct{} { return c.done }

// SendRequest sends a request and blocks until the matching response/error
// arrives, ctx is cancelled, or the channel closes.
func (c *Channel) SendRequest(ctx context.Context, kind MessageKind, payload any) (*Message, error) {
	msg, err := NewMessage(kind, 0, payload)
	if err != nil {
		return nil, err
	}

	// Apply the default per-request timeout only when the caller's ctx carries
	// none, so a peer that is alive but never replies can't hang the request
	// forever. A caller-supplied deadline always wins.
	if c.requestTimeout > 0 {
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.requestTimeout)
			defer cancel()
		}
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, c.closeErr
	}
	id := c.nextID
	c.nextID++
	msg.ID = id
	ch := make(chan *Message, 1) // buffered so dispatch never blocks
	c.pending[id] = ch
	c.mu.Unlock()

	// Register pending BEFORE writing — a fast peer could respond before
	// the resolver is in the map otherwise. A write failure cascade-closes
	// (writeFrame), so the select below wakes via c.done.
	if err := c.writeFrame(msg); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}

	select {
	case resp := <-ch:
		if resp.Kind == KindError {
			var e ErrorResponseData
			_ = resp.Decode(&e)
			return nil, fmt.Errorf("ipc: peer error: %s", e.Message)
		}
		return resp, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	case <-c.done:
		// closeErr is set before close(c.done); the channel-close
		// happens-before edge makes this read safe without the mutex.
		return nil, c.closeErr
	}
}

// SendNotification fires a notification frame (id 0). No reply.
func (c *Channel) SendNotification(kind MessageKind, payload any) error {
	c.mu.Lock()
	if c.closed {
		err := c.closeErr
		c.mu.Unlock()
		return err
	}
	c.mu.Unlock()
	msg, err := NewMessage(kind, 0, payload)
	if err != nil {
		return err
	}
	return c.writeFrame(msg)
}

// deadlineWriter is the optional capability writeFrame uses to bound a write.
// *os.File (including pipes) and net.Conn satisfy it; writers that don't
// degrade gracefully to no write deadline.
type deadlineWriter interface{ SetWriteDeadline(t time.Time) error }

// writeFrame serializes a frame write. A write failure is terminal — the
// peer's read side is gone, so every in-flight request would otherwise hang;
// cascade-close so they all fail promptly (mirrors Node's onOutputError).
//
// The write is bounded by writeTimeout: a peer that stops draining its read
// end makes Write block in the kernel once the OS pipe buffer fills, and a
// cancelled ctx cannot interrupt an in-progress blocking Write. The deadline
// turns that wedge into a terminal error → closeWith, so every blocked
// SendRequest wakes via c.done instead of hanging forever. The deadline is
// re-armed on each write (a stale past-deadline can't linger to a later call).
func (c *Channel) writeFrame(msg *Message) error {
	return c.writeFrameTracked(msg, nil, false)
}

// writeFrameTracked writes one frame and optionally publishes its completion
// through receipt. Write admission and terminal/failure sealing are serialized
// by writeMu:
//
//   - queued writers check closed/sealed only after acquiring writeMu;
//   - a write fault seals admission before releasing writeMu, so a partially
//     written frame can never be followed by another frame;
//   - a terminal response seals admission before its successful receipt is
//     completed, so the waiter does not need to win a scheduling race to Close.
func (c *Channel) writeFrameTracked(msg *Message, receipt *ResponseReceipt, terminal bool) error {
	c.writeMu.Lock()

	if err := c.writeAdmissionError(); err != nil {
		c.writeMu.Unlock()
		if receipt != nil {
			receipt.complete(err)
		}
		return err
	}

	err := c.writeFrameBytes(msg)
	if err != nil {
		c.sealWrites(err)
	} else if terminal {
		c.sealWrites(errTerminalResponse)
	}
	if err == nil && receipt != nil {
		// Framework-owned and non-blocking; safe while writeMu establishes the
		// response-before-application commit point.
		receipt.complete(nil)
	}
	c.writeMu.Unlock()

	if err != nil {
		c.closeWith(err)
		if receipt != nil {
			receipt.complete(err)
		}
	}
	return err
}

func (c *Channel) writeFrameBytes(msg *Message) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("ipc: writer panicked: %v", recovered)
		}
	}()
	if c.writeTimeout > 0 {
		if dw, ok := c.writer.(deadlineWriter); ok {
			_ = dw.SetWriteDeadline(time.Now().Add(c.writeTimeout))
		}
	}
	return WriteFrame(c.writer, msg)
}

// writeAdmissionError is called only while writeMu is held.
func (c *Channel) writeAdmissionError() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.writeSealed {
		return fmt.Errorf("%w: %w", errWriteNotAdmitted, c.writeSealErr)
	}
	if c.closed {
		return fmt.Errorf("%w: %w", errWriteNotAdmitted, c.closeErr)
	}
	return nil
}

// sealWrites is called only while writeMu is held. The first seal reason wins,
// giving every queued writer a stable terminal error.
func (c *Channel) sealWrites(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.writeSealed {
		return
	}
	c.writeSealed = true
	c.writeSealErr = err
}

func (c *Channel) readLoop() {
	for {
		msg, err := ReadFrame(c.reader)
		if err != nil {
			c.closeWith(err)
			return
		}
		c.dispatch(msg)
	}
}

// dispatch routes one decoded frame. Two KNOWN LIMITATIONS are left as-is
// because current inbound request traffic is limited to initialization and no
// notifications are registered. Third-party plugin reverse RPC adds response
// frames, which are routed inline and do not use the asynchronous handler path
// described below. Revisit if this channel carries high-frequency or
// order-sensitive inbound requests or notifications:
//
//  1. Inbound concurrency is UNBOUNDED — every notification/request gets its
//     own goroutine (the async dispatch below is intentional: it lets an
//     in-handler reverse SendRequest receive its reply). A flooding peer thus
//     grows goroutines without bound. A bounded fix must NOT simply stop
//     reading under backpressure — a parked read loop can't deliver a reverse
//     reply, so a handler doing reverse-RPC would deadlock. Safe shape: keep
//     the read loop non-blocking (responses routed inline) plus a bounded
//     inbound worker that drops-oldest + warns past a soft cap.
//  2. Notifications are NOT order-preserving — each runs on its own goroutine,
//     so same-kind notifications can be handled out of arrival order. A
//     streamed/order-sensitive notification would need a single FIFO worker.
//     (The Node side preserves order on its single event loop; only Go side.)
func (c *Channel) dispatch(msg *Message) {
	// Response/error → route to the waiting SendRequest by id.
	if msg.Kind == KindResponse || msg.Kind == KindError {
		c.mu.Lock()
		ch, ok := c.pending[msg.ID]
		if ok {
			delete(c.pending, msg.ID)
		}
		c.mu.Unlock()
		if ok {
			ch <- msg
		} else {
			fmt.Fprintf(os.Stderr, "rslint: orphan response id=%d kind=%s\n", msg.ID, msg.Kind)
		}
		return
	}

	// Notification (id 0) → registered handler, run async with panic safety.
	if msg.ID == 0 {
		h, ok := c.notifyHandlers[msg.Kind]
		if !ok {
			fmt.Fprintf(os.Stderr, "rslint: unhandled notification kind=%s\n", msg.Kind)
			return
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr, "rslint: notification handler %s panicked: %v\n", msg.Kind, r)
				}
			}()
			h(msg)
		}()
		return
	}

	// Inbound request → handler, run async so the read loop keeps consuming
	// frames (lets an in-handler SendRequest receive its reply). A handler
	// panic is trapped and surfaced as an error frame, never crashing the
	// process (mirrors the Node side's runSafely).
	h := c.inbound
	if h == nil {
		c.sendError(msg.ID, fmt.Sprintf("no inbound handler registered (kind=%s)", msg.Kind))
		return
	}
	go c.handleInboundRequest(h, msg)
}

func (c *Channel) handleInboundRequest(h InboundHandler, msg *Message) {
	result, err := callInboundHandler(h, c.inCtx, msg)
	if err != nil {
		err = stableInboundError(err)
		if _, receipt, _, tracked := unwrapTrackedResponse(result); tracked {
			writeErr := c.sendErrorFrame(msg.ID, err.Error(), true)
			if writeErr != nil {
				err = errors.Join(err, writeErr)
				if !errors.Is(writeErr, errWriteNotAdmitted) {
					fmt.Fprintf(os.Stderr, "rslint: write error (id=%d): %v\n", msg.ID, writeErr)
				}
			}
			receipt.complete(err)
			return
		}
		c.sendError(msg.ID, err.Error())
		return
	}
	c.sendResponse(msg.ID, result)
}

// callInboundHandler scopes panic recovery to application code. Response
// writing happens afterward, so a post-handler transport failure can never be
// mistaken for a handler panic and produce a second response for the same ID.
func callInboundHandler(h InboundHandler, ctx context.Context, msg *Message) (result any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = nil
			err = fmt.Errorf("inbound handler panicked: %v", recovered)
		}
	}()
	return h(ctx, msg)
}

func stableInboundError(err error) (stable error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			stable = fmt.Errorf("inbound handler error panicked: %v", recovered)
		}
	}()
	return errors.New(err.Error())
}

func (c *Channel) sendResponse(id int, payload any) {
	payload, receipt, terminal, _ := unwrapTrackedResponse(payload)

	msg, err := newResponseMessage(id, payload)
	if err != nil {
		// The intended response was not written. A tracked lifecycle response
		// seals after its fallback error frame so its waiter can abort without
		// another writer overtaking the eventual Close.
		writeErr := c.sendErrorFrame(
			id,
			fmt.Sprintf("marshal response failed: %v", err),
			receipt != nil || terminal,
		)
		if receipt != nil {
			if writeErr != nil {
				err = errors.Join(err, writeErr)
			}
			receipt.complete(err)
		}
		return
	}
	if err := c.writeFrameTracked(msg, receipt, terminal); err != nil &&
		!errors.Is(err, errWriteNotAdmitted) {
		fmt.Fprintf(os.Stderr, "rslint: write response (id=%d): %v\n", id, err)
	}
}

func unwrapTrackedResponse(payload any) (any, *ResponseReceipt, bool, bool) {
	tracked, ok := payload.(*trackedResponse)
	if !ok {
		return payload, nil, false, false
	}
	return tracked.payload, tracked.receipt, tracked.terminal, true
}

func (c *Channel) sendError(id int, message string) {
	if err := c.sendErrorFrame(id, message, false); err != nil &&
		!errors.Is(err, errWriteNotAdmitted) {
		fmt.Fprintf(os.Stderr, "rslint: write error (id=%d): %v\n", id, err)
	}
}

func (c *Channel) sendErrorFrame(id int, message string, terminal bool) error {
	msg := &Message{Kind: KindError, ID: id}
	if raw, err := marshalJSON(ErrorResponseData{Message: message}); err == nil {
		msg.Data = raw
	}
	return c.writeFrameTracked(msg, nil, terminal)
}

func newResponseMessage(id int, payload any) (msg *Message, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			msg = nil
			err = fmt.Errorf("ipc: marshal response payload panicked: %v", recovered)
		}
	}()
	return NewMessage(KindResponse, id, payload)
}

// closeWith shuts the channel down: marks closed, records the cause, cancels
// inbound ctx, and unblocks every pending SendRequest (they select on c.done).
// Idempotent.
//
// It deliberately does NOT close the underlying reader. readLoop may be parked
// in a blocked ReadFrame on the peer's stdout pipe when the channel is closed
// (the normal CLI shutdown order: lint done → shutdown acked → Close, while the
// peer waits for THIS process to exit before closing its write end). Closing the
// reader to interrupt that read is only reliable on pollable fds; on a Windows
// synchronous-I/O stdin pipe (*os.File).Close blocks in semacquire until the
// in-flight ReadFile returns — which it never will, since the peer is waiting on
// us — so an inline close deadlocks the exit path and a backgrounded close just
// leaks a second goroutine wedged the same way.
//
// Interrupting readLoop is unnecessary. This Channel is used only by the CLI
// (cmd/rslint/ipc_cli.go), which always terminates via os.Exit (main.go); that
// reaps a parked readLoop goroutine instantly. Everything other code observes on
// close — closeErr, c.done, inbound ctx cancellation — is published here
// regardless, so SendRequest waiters and Done() consumers wake exactly as
// before. When readLoop itself originates the close (a ReadFrame error), its
// reader has already returned and the goroutine exits on its own.
func (c *Channel) closeWith(cause error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	if cause == nil || errors.Is(cause, io.EOF) {
		c.closeErr = errors.New("ipc: channel closed (peer EOF)")
	} else {
		c.closeErr = fmt.Errorf("ipc: channel closed: %w", cause)
	}
	if !c.writeSealed {
		c.writeSealed = true
		c.writeSealErr = c.closeErr
	}
	c.pending = make(map[int]chan *Message) // drop refs; waiters wake via c.done
	c.mu.Unlock()

	c.inCancel()
	close(c.done)
}

// Close shuts the channel down. Pending requests fail with a stable error.
func (c *Channel) Close() error {
	c.closeWith(errors.New("closed by caller"))
	return nil
}
