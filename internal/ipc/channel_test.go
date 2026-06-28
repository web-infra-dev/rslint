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
