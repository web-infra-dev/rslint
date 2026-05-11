package ipc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// pipeServicePair stitches two BidirectionalService instances together via
// `io.Pipe`s so they exchange frames in-process. Returns both services and
// a cleanup that closes them in the right order.
//
// Naming: A is the "Go side" we're typically testing; B is the "peer".
type pipeServicePair struct {
	A *BidirectionalService
	B *BidirectionalService
}

func newPipeServicePair() (*pipeServicePair, func()) {
	// A writes to writeAToB; B reads from readBFromA. Symmetric for B→A.
	readBFromA, writeAToB := io.Pipe()
	readAFromB, writeBToA := io.Pipe()

	a := NewBidirectionalService(readAFromB, writeAToB)
	b := NewBidirectionalService(readBFromA, writeBToA)

	return &pipeServicePair{A: a, B: b}, func() {
		// Cleanup pattern (per BidirectionalService.Close docs):
		//   1) Close service → unblock senders, fail pending
		//   2) Close pipes  → reader goroutine hits EOF, wg drains
		//   3) Wait         → goroutines join cleanly
		// Order reversed (pipes first then service) also works; the only
		// hard requirement is that pipes get closed AT SOME POINT so
		// readerLoop exits and Wait returns.
		_ = a.Close()
		_ = b.Close()

		_ = writeAToB.Close()
		_ = writeBToA.Close()
		_ = readBFromA.Close()
		_ = readAFromB.Close()

		a.Wait()
		b.Wait()
	}
}

// ───────────────────────────────────────────────────────────────────
// Test helpers
// ───────────────────────────────────────────────────────────────────

// asMap unmarshals the (untyped) Message.Data into a generic map for
// assertion convenience.
func asMap(t *testing.T, data interface{}) map[string]interface{} {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return out
}

// stringHandler is a small InboundHandler that returns a fixed reply after
// optionally invoking an injection function (e.g. to issue a reverse RPC
// before responding).
type stringHandler struct {
	replyText string
	preReply  func(ctx context.Context) error
}

func (h *stringHandler) Handle(ctx context.Context, msg *Message) (interface{}, error) {
	if h.preReply != nil {
		if err := h.preReply(ctx); err != nil {
			return nil, err
		}
	}
	return map[string]interface{}{"reply": h.replyText, "kind": string(msg.Kind)}, nil
}

// ───────────────────────────────────────────────────────────────────
// Tests
// ───────────────────────────────────────────────────────────────────

// 1. Basic outbound request → response round-trip.
func TestBidirectional_BasicRoundTrip(t *testing.T) {
	pair, cleanup := newPipeServicePair()
	defer cleanup()

	pair.B.SetInboundHandler(&stringHandler{replyText: "hello"})

	pair.A.Start()
	pair.B.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := pair.A.SendRequest(ctx, KindLint, map[string]string{"q": "ping"})
	if err != nil {
		t.Fatalf("SendRequest: %v", err)
	}
	got := asMap(t, resp.Data)
	if got["reply"] != "hello" {
		t.Errorf("reply: got %v, want %q", got["reply"], "hello")
	}
}

//  2. Bidirectional: peer's handler issues a reverse SendRequest while
//     handling its own inbound — reqID multiplexing must keep both flows
//     distinct, no deadlock.
func TestBidirectional_ReverseRPCInsideHandler(t *testing.T) {
	pair, cleanup := newPipeServicePair()
	defer cleanup()

	// A is the linter side; B is the runner side.
	// B's handler for KindInit will issue a KindLintEslintPlugin reverse
	// request to A before replying.
	var reverseObserved atomic.Bool
	pair.A.SetInboundHandler(&stringHandler{replyText: "ack-from-A"})
	pair.B.SetInboundHandler(&stringHandler{
		replyText: "ack-from-B",
		preReply: func(ctx context.Context) error {
			resp, err := pair.B.SendRequest(ctx, KindLintEslintPlugin, map[string]string{"file": "x.ts"})
			if err != nil {
				return err
			}
			data := asMap(t, resp.Data)
			if data["reply"] != "ack-from-A" {
				return errors.New("unexpected reverse reply")
			}
			reverseObserved.Store(true)
			return nil
		},
	})

	pair.A.Start()
	pair.B.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// A → B init. B's handler will reverse-RPC to A while we wait here.
	// If reader/writer split isn't done correctly this deadlocks.
	resp, err := pair.A.SendRequest(ctx, KindInit, map[string]string{})
	if err != nil {
		t.Fatalf("SendRequest(init): %v", err)
	}
	if asMap(t, resp.Data)["reply"] != "ack-from-B" {
		t.Errorf("init reply: got %v", resp.Data)
	}
	if !reverseObserved.Load() {
		t.Errorf("expected reverse RPC to have completed")
	}
}

//  3. Concurrent outbound requests: many in-flight, reqID must keep
//     each waiter on its own response.
func TestBidirectional_ConcurrentReqIDMultiplexing(t *testing.T) {
	pair, cleanup := newPipeServicePair()
	defer cleanup()

	// B echoes the inbound's data verbatim, with kind information.
	pair.B.SetInboundHandler(&handlerEcho{})
	pair.A.Start()
	pair.B.Start()

	const N = 50
	var wg sync.WaitGroup
	wg.Add(N)

	results := make([]string, N)
	for i := range N {
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			resp, err := pair.A.SendRequest(ctx, KindLint, map[string]int{"i": i})
			if err != nil {
				results[i] = "err: " + err.Error()
				return
			}
			d := asMap(t, resp.Data)
			results[i] = jsonNumString(d["i"])
		}(i)
	}
	wg.Wait()

	for i, r := range results {
		want := jsonIntString(i)
		if r != want {
			t.Errorf("result[%d]: got %q, want %q", i, r, want)
		}
	}
}

// handlerEcho returns a {"i": data.i, "kind": ...} payload mirroring
// whatever 'i' the caller sent. Used by the multiplexing test.
type handlerEcho struct{}

func (handlerEcho) Handle(_ context.Context, msg *Message) (interface{}, error) {
	var d map[string]interface{}
	if err := json.Unmarshal(mustMarshal(msg.Data), &d); err != nil {
		return nil, err
	}
	return d, nil
}

func mustMarshal(v interface{}) []byte {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return raw
}

// JSON unmarshal turns numbers into float64; we want string repr for compare.
func jsonNumString(v interface{}) string {
	switch n := v.(type) {
	case float64:
		// float64 with no fractional part is the common case for JSON ints.
		if n == float64(int64(n)) {
			return jsonIntString(int(n))
		}
		return string(mustMarshal(n))
	default:
		return string(mustMarshal(n))
	}
}

func jsonIntString(i int) string {
	return strings.Trim(strings.TrimSpace(strings.Trim(jsonStr(i), `"`)), "")
}

func jsonStr(v interface{}) string {
	return string(mustMarshal(v))
}

//  4. Large frame: a single >64 KiB frame must round-trip cleanly. This
//     exercises the writer-goroutine + buffered channel decoupling that
//     prevents pipe-buffer deadlock.
func TestBidirectional_LargeFrame(t *testing.T) {
	pair, cleanup := newPipeServicePair()
	defer cleanup()

	pair.B.SetInboundHandler(&handlerEcho{})
	pair.A.Start()
	pair.B.Start()

	const size = 200 * 1024 // 200 KiB > Linux default 64 KiB pipe buffer
	big := strings.Repeat("a", size)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := pair.A.SendRequest(ctx, KindLint, map[string]string{"blob": big})
	if err != nil {
		t.Fatalf("large frame SendRequest: %v", err)
	}
	d := asMap(t, resp.Data)
	got, _ := d["blob"].(string)
	if len(got) != size {
		t.Errorf("large frame: got %d bytes, want %d", len(got), size)
	}
}

//  5. Notifications: id=0 frames should reach the registered notification
//     handler without consuming a pending entry.
func TestBidirectional_Notification(t *testing.T) {
	pair, cleanup := newPipeServicePair()
	defer cleanup()

	var got atomic.Value // string
	pair.B.RegisterNotification(KindLog, func(_ context.Context, msg *Message) {
		var p struct {
			Text string `json:"text"`
		}
		_ = json.Unmarshal(mustMarshal(msg.Data), &p)
		got.Store(p.Text)
	})

	pair.A.Start()
	pair.B.Start()

	if err := pair.A.SendNotification(KindLog, map[string]string{"text": "hello-log"}); err != nil {
		t.Fatalf("SendNotification: %v", err)
	}

	// Wait briefly for the notification to be processed.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if v, _ := got.Load().(string); v == "hello-log" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("notification was not received in time; got=%v", got.Load())
}

//  6. Unhandled inbound request: when no inbound handler is set, the
//     request must get a clear `error` reply rather than hanging.
func TestBidirectional_UnhandledInboundReturnsError(t *testing.T) {
	pair, cleanup := newPipeServicePair()
	defer cleanup()

	// B has NO inbound handler set.
	pair.A.Start()
	pair.B.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := pair.A.SendRequest(ctx, KindLint, map[string]string{})
	if err == nil {
		t.Fatalf("expected error reply, got nil")
	}
	if !strings.Contains(err.Error(), "no inbound handler registered") {
		t.Errorf("error message: %v", err)
	}
}

//  7. Close on the local service must wake up in-flight outbound SendRequest
//     waiters with an error, not hang. This is the property that lets the
//     LSP path cancel the runner cleanly when the editor disconnects.
func TestBidirectional_CloseFailsPending(t *testing.T) {
	pair, cleanup := newPipeServicePair()
	defer cleanup()

	// B handler hangs forever — A will never get a normal response.
	hold := make(chan struct{})
	pair.B.SetInboundHandler(holdingHandler{hold: hold})
	pair.A.Start()
	pair.B.Start()

	resp := make(chan error, 1)
	go func() {
		_, err := pair.A.SendRequest(context.Background(), KindLint, map[string]string{})
		resp <- err
	}()

	// Give the request time to register in pending, then close the LOCAL
	// service (A). Closing A.stopCh + failAllPending wakes the SendRequest
	// waiter with an error.
	time.Sleep(50 * time.Millisecond)
	_ = pair.A.Close()

	select {
	case err := <-resp:
		if err == nil {
			t.Errorf("expected error after Close, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Close did not unblock pending SendRequest")
	}
	close(hold) // unblock B's handler so its goroutine can return
}

type holdingHandler struct{ hold chan struct{} }

func (h holdingHandler) Handle(ctx context.Context, _ *Message) (interface{}, error) {
	select {
	case <-h.hold:
		return map[string]string{}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// 8. Encode/decode round-trip: catches accidental wire-format drift.
func TestEncodeDecodeFrame(t *testing.T) {
	cases := []Message{
		{Kind: KindInit, ID: 1, Data: json.RawMessage(`{"a":1}`)},
		{Kind: KindLog, ID: 0, Data: json.RawMessage(`{"text":"x"}`)},
		// Empty data
		{Kind: KindHandshake, ID: 99, Data: json.RawMessage(`null`)},
	}
	for _, m := range cases {
		frame, err := encodeFrame(m)
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		// Frame must start with a 4-byte u32 LE length matching body bytes.
		// We can recover the body length and sanity-check by reading.
		// reuse readFrame against a Reader wrapping the frame bytes
		decoded, err := readFrame(bufioReaderOf(frame))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if decoded.Kind != m.Kind || decoded.ID != m.ID {
			t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, m)
		}
	}
}

// bufioReaderOf wraps a byte slice in a *bufio.Reader for readFrame.
func bufioReaderOf(b []byte) *bufio.Reader {
	return bufio.NewReader(bytes.NewReader(b))
}

// 9. Notifications must be delivered to their handler IN WIRE ORDER.
// Order-bearing streams (the CLI's `output` redirect for lint
// diagnostics) get corrupted when consecutive notifications race their
// handler goroutines and interleave their side-effects. Regression
// test for the previous goroutine-per-notification dispatch; the
// current implementation runs handlers synchronously in the reader
// goroutine.
func TestNotifications_DispatchedInWireOrder(t *testing.T) {
	pair, cleanup := newPipeServicePair()
	defer cleanup()

	const n = 200
	var (
		mu       sync.Mutex
		received []int
	)
	pair.B.RegisterNotification(KindLog, func(_ context.Context, msg *Message) {
		var rec struct {
			Seq int `json:"seq"`
		}
		raw, _ := json.Marshal(msg.Data)
		_ = json.Unmarshal(raw, &rec)

		// A handler whose runtime varies per-notification — without
		// in-order dispatch this is precisely the kind of variance
		// that causes reordering on the receiver side. The test
		// asserts the recorded order matches the send order despite
		// the artificial jitter.
		if rec.Seq%17 == 0 {
			time.Sleep(time.Millisecond)
		}
		mu.Lock()
		received = append(received, rec.Seq)
		mu.Unlock()
	})

	pair.A.Start()
	pair.B.Start()

	for i := range n {
		if err := pair.A.SendNotification(KindLog, map[string]int{"seq": i}); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}

	// Wait until all are received.
	deadline := time.After(5 * time.Second)
	for {
		mu.Lock()
		got := len(received)
		mu.Unlock()
		if got >= n {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("only received %d/%d notifications", got, n)
		case <-time.After(10 * time.Millisecond):
		}
	}

	mu.Lock()
	defer mu.Unlock()
	for i, seq := range received {
		if seq != i {
			t.Fatalf("notification %d arrived out of order: got seq=%d, want %d", i, seq, i)
		}
	}
}

// Wait() must include inbound request handler goroutines. The previous
// dispatchInbound spawned `go func() { handler.Handle(...) }()` without
// a wg ticket, so Close + Wait would return while a slow handler was
// still running — a goroutine leak from the caller's perspective.
//
// We construct B's reader/writer manually so we can close them
// independently and isolate "Wait waits for handler" from "Wait waits
// for reader". A sends a request through the pipes; B's reader picks
// it up and dispatches to the blocking handler. We then close all
// pipes and B itself — at that point only the handler is still
// running. Wait must remain blocked until the handler releases.
func TestBidirectionalService_WaitJoinsInboundHandlers(t *testing.T) {
	readBFromA, writeAToB := io.Pipe()
	readAFromB, writeBToA := io.Pipe()
	a := NewBidirectionalService(readAFromB, writeAToB)
	b := NewBidirectionalService(readBFromA, writeBToA)

	holdReleased := make(chan struct{})
	handlerEntered := make(chan struct{}, 1)
	b.SetInboundHandler(blockingHandler{
		entered: handlerEntered,
		release: holdReleased,
	})

	a.Start()
	b.Start()

	// Fire a request from A. Don't await the response — we just need
	// B's handler to start. Attach a no-op error catcher because A's
	// SendRequest will eventually error once we close the pipes.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = a.SendRequest(ctx, KindLint, map[string]string{})
	}()

	// Wait for the handler to actually start.
	select {
	case <-handlerEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("handler never entered")
	}

	// Now close ALL the pipes and B. With pipes closed, B's reader
	// goroutine hits EOF and exits. The only outstanding wg ticket
	// should be the handler goroutine. Wait must therefore reflect
	// the handler's progress.
	_ = b.Close()
	_ = writeAToB.Close()
	_ = readBFromA.Close()

	waitDone := make(chan struct{})
	go func() {
		b.Wait()
		close(waitDone)
	}()

	// Wait must be blocked because the handler hasn't finished.
	select {
	case <-waitDone:
		t.Fatal("Wait returned BEFORE the handler completed — handler goroutines not tracked")
	case <-time.After(300 * time.Millisecond):
		// Expected: still blocked on handler.
	}

	// Release the handler. Now Wait should unblock promptly.
	close(holdReleased)

	select {
	case <-waitDone:
		// Expected.
	case <-time.After(3 * time.Second):
		t.Fatal("Wait did not return after handler completion")
	}

	// Cleanup remaining bits.
	_ = a.Close()
	_ = writeBToA.Close()
	_ = readAFromB.Close()
	a.Wait()
}

type blockingHandler struct {
	entered chan struct{}
	release chan struct{}
}

// Handle simulates a long-running unit of work that does NOT cooperate
// with cancellation (file I/O, external IPC, native call). The whole
// point of the wg-ticket invariant being tested is that Wait must
// reflect such uncooperative work — so we deliberately ignore ctx
// here.
//
// The prior version `select`ed on `ctx.Done()` too. That used to be a
// no-op because the framework passed `context.Background()` (never
// done), but after the M4 fix Close() cancels the service ctx — a
// cooperative handler would then return immediately, defeating the
// test's purpose (we'd be measuring how fast the handler bails on
// cancel, not whether Wait tracks the handler goroutine).
func (h blockingHandler) Handle(_ context.Context, _ *Message) (interface{}, error) {
	select {
	case h.entered <- struct{}{}:
	default:
	}
	<-h.release
	return map[string]string{}, nil
}

// 11. Oversized frame header is rejected without allocating the body.
// Guards against the OOM-on-stream-desync class of bug: a header
// advertising 4 GiB is interpreted as "protocol error", not as a
// `make([]byte, 4 GiB)` request.
func TestReadFrame_RejectsOversizedHeader(t *testing.T) {
	// Build a frame header advertising maxFrameSize+1 bytes. Don't bother
	// supplying a body — the cap check must reject before reading body.
	var header [4]byte
	binary.LittleEndian.PutUint32(header[:], maxFrameSize+1)
	r := bufio.NewReader(bytes.NewReader(header[:]))

	_, err := readFrame(r)
	if err == nil {
		t.Fatal("expected error for oversized frame header, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds cap") {
		t.Errorf("expected cap-exceeded error, got: %v", err)
	}
}

// readFrame at exactly maxFrameSize is the boundary: must still attempt
// to read (and fail on EOF since we didn't supply the body), NOT reject
// the header. This proves the cap is inclusive of valid frames and
// exclusive of bogus ones.
func TestReadFrame_AcceptsMaxSizedHeader(t *testing.T) {
	var header [4]byte
	binary.LittleEndian.PutUint32(header[:], maxFrameSize)
	r := bufio.NewReader(bytes.NewReader(header[:]))

	_, err := readFrame(r)
	// Expect a "read body" error (since we provided no body) — NOT the
	// "exceeds cap" error.
	if err == nil {
		t.Fatal("expected error reading missing body, got nil")
	}
	if strings.Contains(err.Error(), "exceeds cap") {
		t.Errorf("max-sized header should NOT be cap-rejected: %v", err)
	}
	if !strings.Contains(err.Error(), "read body") && !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("expected body-read error or ErrUnexpectedEOF, got: %v", err)
	}
}
