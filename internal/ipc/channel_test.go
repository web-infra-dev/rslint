package ipc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// newChannelPair wires two channels back-to-back over two io.Pipes so they
// can talk to each other in-process (A writes → B reads, and vice versa).
// Returned channels are NOT started; the caller sets handlers then Start()s.
func newChannelPair(t *testing.T) (a *Channel, b *Channel) {
	t.Helper()
	abR, abW := io.Pipe() // A → B
	baR, baW := io.Pipe() // B → A
	a = NewChannel(baR, abW)
	b = NewChannel(abR, baW)
	t.Cleanup(func() {
		_ = a.Close()
		_ = b.Close()
		_ = abW.Close()
		_ = baW.Close()
	})
	return a, b
}

type greetReq struct {
	Name string `json:"name"`
}
type greetResp struct {
	Greeting string `json:"greeting"`
}

func TestChannel_RequestResponse(t *testing.T) {
	a, b := newChannelPair(t)
	b.SetInboundHandler(func(_ context.Context, msg *Message) (any, error) {
		var req greetReq
		if err := msg.Decode(&req); err != nil {
			return nil, err
		}
		if msg.Kind != "greet" {
			t.Errorf("unexpected kind %q", msg.Kind)
		}
		return greetResp{Greeting: "hi " + req.Name}, nil
	})
	a.Start()
	b.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := a.SendRequest(ctx, "greet", greetReq{Name: "world"})
	if err != nil {
		t.Fatalf("SendRequest: %v", err)
	}
	if resp.Kind != KindResponse {
		t.Fatalf("expected response kind, got %q", resp.Kind)
	}
	var out greetResp
	if err := resp.Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Greeting != "hi world" {
		t.Fatalf("got %q", out.Greeting)
	}
}

func TestChannel_InboundError(t *testing.T) {
	a, b := newChannelPair(t)
	b.SetInboundHandler(func(_ context.Context, _ *Message) (any, error) {
		return nil, io.ErrUnexpectedEOF // arbitrary handler failure
	})
	a.Start()
	b.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := a.SendRequest(ctx, "boom", nil)
	if err == nil {
		t.Fatal("expected error from peer handler")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "peer error") {
		t.Fatalf("expected peer error, got %q", got)
	}
}

func TestChannel_Notification(t *testing.T) {
	a, b := newChannelPair(t)
	got := make(chan string, 1)
	b.RegisterNotification("log", func(msg *Message) {
		var s struct {
			Text string `json:"text"`
		}
		_ = msg.Decode(&s)
		got <- s.Text
	})
	a.Start()
	b.Start()

	if err := a.SendNotification("log", map[string]string{"text": "hello"}); err != nil {
		t.Fatalf("SendNotification: %v", err)
	}
	select {
	case s := <-got:
		if s != "hello" {
			t.Fatalf("got %q", s)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("notification not delivered")
	}
}

func TestChannel_ContextCancel(t *testing.T) {
	a, b := newChannelPair(t)
	// Handler never replies (blocks), so the request only ends via ctx.
	block := make(chan struct{})
	b.SetInboundHandler(func(_ context.Context, _ *Message) (any, error) {
		<-block
		return struct{}{}, nil
	})
	t.Cleanup(func() { close(block) })
	a.Start()
	b.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, err := a.SendRequest(ctx, "hang", nil)
	if err == nil {
		t.Fatal("expected ctx cancel error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestChannel_CloseRejectsPending(t *testing.T) {
	a, b := newChannelPair(t)
	block := make(chan struct{})
	b.SetInboundHandler(func(_ context.Context, _ *Message) (any, error) {
		<-block
		return struct{}{}, nil
	})
	t.Cleanup(func() { close(block) })
	a.Start()
	b.Start()

	errCh := make(chan error, 1)
	go func() {
		_, err := a.SendRequest(context.Background(), "hang", nil)
		errCh <- err
	}()
	time.Sleep(50 * time.Millisecond) // let the request register + write
	_ = a.Close()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected close error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pending request not rejected on close")
	}
}

func TestFrame_RoundTrip(t *testing.T) {
	src, err := NewMessage("hello", 7, map[string]int{"x": 42})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteFrame(&buf, src); err != nil {
		t.Fatalf("WriteFrame: %v", err)
	}
	got, err := ReadFrame(bufio.NewReader(&buf))
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}
	if got.Kind != "hello" || got.ID != 7 {
		t.Fatalf("got kind=%q id=%d", got.Kind, got.ID)
	}
	var v struct {
		X int `json:"x"`
	}
	if err := got.Decode(&v); err != nil || v.X != 42 {
		t.Fatalf("decode mismatch: v=%+v err=%v", v, err)
	}
}

type shortWriteWriter struct {
	failCall int
	failN    func(int) int
	failErr  error
	calls    int
}

func (w *shortWriteWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == w.failCall {
		return w.failN(len(p)), w.failErr
	}
	return len(p), nil
}

func TestWriteFrame_RejectsShortWrites(t *testing.T) {
	msg, err := NewMessage("hello", 7, map[string]int{"x": 42})
	if err != nil {
		t.Fatal(err)
	}
	writeBoom := errors.New("write boom")
	tests := []struct {
		name     string
		failCall int
		failN    func(int) int
		failErr  error
		want     error
	}{
		{"header partial nil", 1, func(n int) int { return n - 1 }, nil, io.ErrShortWrite},
		{"header zero nil", 1, func(int) int { return 0 }, nil, io.ErrShortWrite},
		{"body partial nil", 2, func(n int) int { return n - 1 }, nil, io.ErrShortWrite},
		{"body zero nil", 2, func(int) int { return 0 }, nil, io.ErrShortWrite},
		{"body full with error", 2, func(n int) int { return n }, writeBoom, writeBoom},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := &shortWriteWriter{
				failCall: test.failCall,
				failN:    test.failN,
				failErr:  test.failErr,
			}
			err := WriteFrame(writer, msg)
			if !errors.Is(err, test.want) {
				t.Fatalf("WriteFrame error = %v, want %v", err, test.want)
			}
		})
	}
}

// TestNewMessage_NilContainerPayloadOmitted pins the wire-parity fix: a nil
// map/slice/pointer (and untyped nil) is "no payload" → the data field is
// omitted, matching Node where `undefined` is dropped by JSON.stringify. An
// EMPTY but non-nil container still marshals to `{}` / `[]`.
func TestNewMessage_NilContainerPayloadOmitted(t *testing.T) {
	cases := []struct {
		name     string
		payload  any
		wantData bool
	}{
		{"untyped nil", nil, false},
		{"nil map", map[string]any(nil), false},
		{"nil slice", []int(nil), false},
		{"nil pointer", (*struct{ X int })(nil), false},
		{"empty map", map[string]any{}, true},
		{"empty slice", []int{}, true},
		{"non-empty map", map[string]any{"a": 1}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := NewMessage("k", 1, tc.payload)
			if err != nil {
				t.Fatalf("NewMessage: %v", err)
			}
			if gotData := len(msg.Data) > 0; gotData != tc.wantData {
				t.Errorf("data present = %v, want %v (Data=%q)", gotData, tc.wantData, msg.Data)
			}
		})
	}
}

func TestReadFrame_CapExceeded(t *testing.T) {
	var buf bytes.Buffer
	// A length header beyond the cap must be rejected before allocating.
	if err := binary.Write(&buf, binary.LittleEndian, uint32(maxFrameSize+1)); err != nil {
		t.Fatal(err)
	}
	_, err := ReadFrame(bufio.NewReader(&buf))
	if err == nil {
		t.Fatal("expected frame-cap error")
	}
	if !strings.Contains(err.Error(), "exceeds cap") {
		t.Fatalf("expected cap error, got %v", err)
	}
}

func TestReadFrame_EOF(t *testing.T) {
	// Empty reader → clean io.EOF (unwrapped, so callers can detect close).
	_, err := ReadFrame(bufio.NewReader(bytes.NewReader(nil)))
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

// ── robustness regressions (aligning with the Node IpcClient) ──

func TestChannel_RegisterAfterStartPanics(t *testing.T) {
	a, _ := newChannelPair(t)
	a.Start()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when registering a handler after Start")
		}
	}()
	a.SetInboundHandler(func(context.Context, *Message) (any, error) { return struct{}{}, nil })
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write boom") }

type gatedFrameWriter struct {
	mu          sync.Mutex
	buf         bytes.Buffer
	calls       int
	bodyEntered chan struct{}
	releaseBody chan struct{}
	releaseOnce sync.Once
}

func newGatedFrameWriter() *gatedFrameWriter {
	return &gatedFrameWriter{
		bodyEntered: make(chan struct{}),
		releaseBody: make(chan struct{}),
	}
}

func (w *gatedFrameWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	w.calls++
	call := w.calls
	w.mu.Unlock()
	if call == 2 {
		close(w.bodyEntered)
		<-w.releaseBody
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

func (w *gatedFrameWriter) release() {
	w.releaseOnce.Do(func() { close(w.releaseBody) })
}

func (w *gatedFrameWriter) bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return bytes.Clone(w.buf.Bytes())
}

func writeInboundRequest(t *testing.T, w io.Writer, kind MessageKind, id int) {
	t.Helper()
	msg, err := NewMessage(kind, id, struct{}{})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	if err := WriteFrame(w, msg); err != nil {
		t.Fatalf("WriteFrame: %v", err)
	}
}

func TestChannel_ResponseReceiptWaitsForCompleteFrame(t *testing.T) {
	inR, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close() })
	writer := newGatedFrameWriter()
	t.Cleanup(writer.release)

	receiptCh := make(chan *ResponseReceipt, 1)
	c := NewChannel(inR, writer)
	c.SetInboundHandler(func(context.Context, *Message) (any, error) {
		response, receipt := TrackResponse(map[string]any{"ok": true})
		receiptCh <- receipt
		return response, nil
	})
	c.Start()
	t.Cleanup(func() { _ = c.Close() })

	writeInboundRequest(t, inW, "init", 1)
	receipt := <-receiptCh
	select {
	case <-writer.bodyEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("response body write did not reach the gate")
	}
	select {
	case <-receipt.Done():
		t.Fatal("receipt completed before the response body write returned")
	default:
	}

	writer.release()
	select {
	case <-receipt.Done():
		if err := receipt.Err(); err != nil {
			t.Fatalf("response receipt: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("receipt did not complete after the full response write")
	}

	msg, err := ReadFrame(bufio.NewReader(bytes.NewReader(writer.bytes())))
	if err != nil {
		t.Fatalf("read committed response: %v", err)
	}
	if msg.Kind != KindResponse || msg.ID != 1 {
		t.Fatalf("committed frame = kind %q id %d, want response id 1", msg.Kind, msg.ID)
	}
}

func TestChannel_ResponseReceiptReportsMarshalFailure(t *testing.T) {
	inR, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close() })
	var output bytes.Buffer

	receiptCh := make(chan *ResponseReceipt, 1)
	c := NewChannel(inR, &output)
	c.SetInboundHandler(func(context.Context, *Message) (any, error) {
		response, receipt := TrackResponse(func() {})
		receiptCh <- receipt
		return response, nil
	})
	c.Start()
	t.Cleanup(func() { _ = c.Close() })

	writeInboundRequest(t, inW, "init", 1)
	receipt := <-receiptCh
	select {
	case <-receipt.Done():
		if err := receipt.Err(); err == nil || !strings.Contains(err.Error(), "unsupported type") {
			t.Fatalf("receipt error = %v, want marshal failure", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("marshal failure did not complete the response receipt")
	}

	msg, err := ReadFrame(bufio.NewReader(bytes.NewReader(output.Bytes())))
	if err != nil {
		t.Fatalf("read fallback error response: %v", err)
	}
	if msg.Kind != KindError || msg.ID != 1 {
		t.Fatalf("fallback frame = kind %q id %d, want error id 1", msg.Kind, msg.ID)
	}
	if err := c.SendNotification("late", nil); err == nil {
		t.Fatal("tracked marshal failure did not seal later writes")
	}
}

func TestChannel_TrackedHandlerErrorCompletesReceipt(t *testing.T) {
	inR, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close() })
	var output bytes.Buffer
	receiptCh := make(chan *ResponseReceipt, 1)

	c := NewChannel(inR, &output)
	c.SetInboundHandler(func(context.Context, *Message) (any, error) {
		response, receipt := TrackResponse(map[string]any{"ignored": true})
		receiptCh <- receipt
		return response, errors.New("handler failed")
	})
	c.Start()
	t.Cleanup(func() { _ = c.Close() })

	writeInboundRequest(t, inW, "init", 1)
	receipt := <-receiptCh
	select {
	case <-receipt.Done():
		if err := receipt.Err(); err == nil || !strings.Contains(err.Error(), "handler failed") {
			t.Fatalf("receipt error = %v, want handler failure", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("tracked handler error did not complete the response receipt")
	}

	msg, err := ReadFrame(bufio.NewReader(bytes.NewReader(output.Bytes())))
	if err != nil {
		t.Fatalf("read handler error response: %v", err)
	}
	if msg.Kind != KindError || msg.ID != 1 {
		t.Fatalf("handler error frame = kind %q id %d, want error id 1", msg.Kind, msg.ID)
	}
	var responseErr ErrorResponseData
	if err := msg.Decode(&responseErr); err != nil {
		t.Fatalf("decode handler error: %v", err)
	}
	if responseErr.Message != "handler failed" {
		t.Fatalf("handler error message = %q, want %q", responseErr.Message, "handler failed")
	}
	if err := c.SendNotification("late", nil); !errors.Is(err, errTerminalResponse) {
		t.Fatalf("late write error = %v, want %v", err, errTerminalResponse)
	}
}

func TestChannel_ResponseReceiptReportsCloseBeforeWrite(t *testing.T) {
	inR, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close() })
	var output bytes.Buffer
	handlerRelease := make(chan struct{})

	receiptCh := make(chan *ResponseReceipt, 1)
	c := NewChannel(inR, &output)
	c.SetInboundHandler(func(context.Context, *Message) (any, error) {
		response, receipt := TrackResponse(map[string]any{"ok": true})
		receiptCh <- receipt
		<-handlerRelease
		return response, nil
	})
	c.Start()
	t.Cleanup(func() { _ = c.Close() })

	writeInboundRequest(t, inW, "init", 1)
	receipt := <-receiptCh
	_ = c.Close()
	close(handlerRelease)
	select {
	case <-receipt.Done():
		if err := receipt.Err(); err == nil {
			t.Fatal("receipt reported success after the channel closed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("closed channel did not complete the response receipt")
	}
	if output.Len() != 0 {
		t.Fatalf("closed channel wrote %d unexpected bytes", output.Len())
	}
}

type panicWriter struct{}

func (panicWriter) Write([]byte) (int, error) { panic("writer boom") }

func TestChannel_ResponseReceiptReportsWriterPanic(t *testing.T) {
	inR, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close() })
	receiptCh := make(chan *ResponseReceipt, 1)

	c := NewChannel(inR, panicWriter{})
	c.SetInboundHandler(func(context.Context, *Message) (any, error) {
		response, receipt := TrackResponse(map[string]any{"ok": true})
		receiptCh <- receipt
		return response, nil
	})
	c.Start()
	t.Cleanup(func() { _ = c.Close() })

	writeInboundRequest(t, inW, "init", 1)
	receipt := <-receiptCh
	select {
	case <-receipt.Done():
		if err := receipt.Err(); err == nil || !strings.Contains(err.Error(), "writer panicked") {
			t.Fatalf("receipt error = %v, want writer panic", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("writer panic did not complete the response receipt")
	}
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("writer panic did not close the channel")
	}
}

func TestChannel_TerminalResponseSealsLaterWrites(t *testing.T) {
	inR, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close() })
	writer := newGatedFrameWriter()
	t.Cleanup(writer.release)
	receiptCh := make(chan *ResponseReceipt, 1)

	c := NewChannel(inR, writer)
	c.SetInboundHandler(func(context.Context, *Message) (any, error) {
		response, receipt := TrackTerminalResponse(struct{}{})
		receiptCh <- receipt
		return response, nil
	})
	c.Start()
	t.Cleanup(func() { _ = c.Close() })

	writeInboundRequest(t, inW, KindExit, 1)
	receipt := <-receiptCh
	select {
	case <-writer.bodyEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("terminal response body write did not reach the gate")
	}

	lateResult := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() {
		_, err := c.SendRequest(ctx, "late", nil)
		lateResult <- err
	}()

	pendingDeadline := time.After(2 * time.Second)
	for {
		c.mu.Lock()
		pending := len(c.pending)
		c.mu.Unlock()
		if pending == 1 {
			break
		}
		select {
		case <-pendingDeadline:
			t.Fatal("late writer did not queue behind the terminal response")
		default:
			runtime.Gosched()
		}
	}
	select {
	case err := <-lateResult:
		t.Fatalf("queued writer returned before terminal response was released: %v", err)
	default:
	}

	writer.release()
	select {
	case <-receipt.Done():
		if err := receipt.Err(); err != nil {
			t.Fatalf("terminal response receipt: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("terminal response receipt did not complete")
	}

	select {
	case err := <-lateResult:
		if !errors.Is(err, errTerminalResponse) {
			t.Fatalf("queued write error = %v, want %v", err, errTerminalResponse)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("queued writer did not observe the terminal write seal")
	}
	reader := bufio.NewReader(bytes.NewReader(writer.bytes()))
	if _, err := ReadFrame(reader); err != nil {
		t.Fatalf("read terminal response: %v", err)
	}
	if _, err := ReadFrame(reader); !errors.Is(err, io.EOF) {
		t.Fatalf("unexpected frame after terminal response: %v", err)
	}

	_ = c.Close()
	c.writeMu.Lock()
	admissionErr := c.writeAdmissionError()
	c.writeMu.Unlock()
	if !errors.Is(admissionErr, errTerminalResponse) {
		t.Fatalf("post-close admission error = %v, want first seal reason %v", admissionErr, errTerminalResponse)
	}
}

type panickingError struct{}

func (panickingError) Error() string { panic("error boom") }

func TestChannel_HandlerErrorPanicRecovered(t *testing.T) {
	a, b := newChannelPair(t)
	b.SetInboundHandler(func(context.Context, *Message) (any, error) {
		return nil, panickingError{}
	})
	a.Start()
	b.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for range 2 {
		_, err := a.SendRequest(ctx, "x", nil)
		if err == nil {
			t.Fatal("expected peer error")
		}
		if !strings.Contains(err.Error(), "inbound handler error panicked: error boom") {
			t.Fatalf("unexpected peer error: %v", err)
		}
		select {
		case <-b.Done():
			t.Fatal("panicking error stringer closed the channel")
		default:
		}
	}
}

func TestChannel_WriteErrorCascades(t *testing.T) {
	pr, pw := io.Pipe()
	t.Cleanup(func() { _ = pw.Close() })
	c := NewChannel(pr, failWriter{})
	c.Start()

	// A write failure is terminal: SendRequest returns the error AND the
	// channel closes, so other in-flight requests don't hang forever.
	if _, err := c.SendRequest(context.Background(), "x", "payload"); err == nil {
		t.Fatal("expected write error")
	}
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("write failure did not cascade-close the channel")
	}
}

func TestChannel_HandlerPanicRecovered(t *testing.T) {
	a, b := newChannelPair(t)
	b.SetInboundHandler(func(context.Context, *Message) (any, error) {
		panic("handler boom")
	})
	a.Start()
	b.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// A panicking handler must surface as an error frame, not crash the
	// process or hang the caller.
	_, err := a.SendRequest(ctx, "x", nil)
	if err == nil {
		t.Fatal("expected error from panicking handler")
	}
	if !strings.Contains(err.Error(), "peer error") {
		t.Fatalf("expected peer error frame, got %v", err)
	}
}

// TestChannel_WriteDeadline pins #1: a peer that never drains its read end
// makes a large Write block in the kernel; the per-write deadline turns that
// into a terminal error so SendRequest returns instead of hanging forever.
func TestChannel_WriteDeadline(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Windows anonymous pipes (os.Pipe) don't implement SetWriteDeadline, so
		// the write deadline degrades to a no-op there (by design — writeFrame
		// ignores the SetWriteDeadline error). The deadline path this test pins
		// is POSIX-only; on Windows the blocking Write never unblocks, so skip.
		t.Skip("os.Pipe does not support SetWriteDeadline on Windows")
	}
	inR, inW := io.Pipe() // inbound never produces a frame (readLoop parks, not EOF)
	defer inW.Close()
	pr, pw, err := os.Pipe() // outbound: reader end (pr) is never drained
	if err != nil {
		t.Fatal(err)
	}
	defer pr.Close()
	defer pw.Close()

	c := NewChannel(inR, pw)
	c.writeTimeout = 100 * time.Millisecond // short; set before Start
	c.Start()
	defer c.Close()

	// A frame far larger than the OS pipe buffer (~64 KiB) that pr never
	// drains: Write blocks → deadline fires → writeFrame errors → closeWith.
	big := strings.Repeat("x", 512*1024)
	done := make(chan error, 1)
	go func() {
		_, e := c.SendRequest(context.Background(), "big", map[string]string{"d": big})
		done <- e
	}()
	select {
	case e := <-done:
		if e == nil {
			t.Fatal("expected write-deadline error, got nil")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("SendRequest hung past the write deadline")
	}
}

// TestChannel_RequestTimeout pins #6: when the caller's ctx carries no
// deadline, the channel's requestTimeout bounds a request whose peer is alive
// but never replies.
func TestChannel_RequestTimeout(t *testing.T) {
	a, b := newChannelPair(t)
	b.SetInboundHandler(func(ctx context.Context, _ *Message) (any, error) {
		<-ctx.Done() // never reply until the channel closes
		return nil, ctx.Err()
	})
	a.requestTimeout = 100 * time.Millisecond // set before Start
	a.Start()
	b.Start()

	start := time.Now()
	_, err := a.SendRequest(context.Background(), "slow", nil) // ctx has no deadline
	if err == nil {
		t.Fatal("expected request-timeout error, got nil")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("request not bounded by requestTimeout (took %v)", elapsed)
	}
}

func TestChannel_SendAfterClose(t *testing.T) {
	a, _ := newChannelPair(t)
	a.Start()
	_ = a.Close()
	if _, err := a.SendRequest(context.Background(), "x", nil); err == nil {
		t.Fatal("expected SendRequest error after close")
	}
	if err := a.SendNotification("x", nil); err == nil {
		t.Fatal("expected SendNotification error after close")
	}
}

// blockingCloseReader models a Windows synchronous-I/O stdin pipe: Read parks
// (so readLoop blocks in ReadFrame) and Close blocks until released — mirroring
// (*os.File).Close on a fd whose in-flight ReadFile cannot be interrupted.
type blockingCloseReader struct {
	release     chan struct{} // test closes this to let Read and Close return
	closeCalled chan struct{} // closed iff Close is ever invoked
}

func (r *blockingCloseReader) Read([]byte) (int, error) {
	<-r.release
	return 0, io.EOF
}

func (r *blockingCloseReader) Close() error {
	close(r.closeCalled)
	<-r.release // uninterruptible, like a Windows FD.Close awaiting a live read
	return nil
}

// TestChannel_CloseDoesNotBlockOnReader pins the windows-latest hang fix:
// closeWith must NOT close (or otherwise block on) the underlying reader. When
// the channel closes, readLoop is typically parked in a blocked ReadFrame; on a
// Windows synchronous pipe, closing the reader to interrupt it makes
// (*os.File).Close block until that read returns — which it can't, since the
// peer is waiting for this process to exit — deadlocking the CLI's exit path.
// Reaping the parked readLoop is left to os.Exit (the sole caller is the CLI).
func TestChannel_CloseDoesNotBlockOnReader(t *testing.T) {
	r := &blockingCloseReader{
		release:     make(chan struct{}),
		closeCalled: make(chan struct{}),
	}
	t.Cleanup(func() { close(r.release) }) // unpark Read (and any Close) at the end
	_, pw := io.Pipe()
	t.Cleanup(func() { _ = pw.Close() })

	c := NewChannel(r, pw)
	c.Start() // readLoop parks in r.Read

	done := make(chan struct{})
	go func() { _ = c.Close(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close blocked while readLoop was parked in Read — closeWith must not close/await the reader")
	}

	// The channel is closed, and the reader was never touched.
	select {
	case <-c.Done():
	default:
		t.Fatal("Close did not shut the channel down")
	}
	select {
	case <-r.closeCalled:
		t.Fatal("closeWith closed the underlying reader; on Windows that deadlocks the exit path")
	default:
	}
}
