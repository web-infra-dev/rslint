package ipc

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

// Regression for M3: SendResponse / SendErrorResponse used to push frames
// straight onto writeCh without checking `closed`, so a handler that
// reached the reply line during shutdown would race the drain loop and
// see its frame silently dropped. After the fix, both functions fail
// fast with "api: service closed" the moment Close() has flipped the
// flag — same contract as SendRequest / SendNotification.
func TestSendResponse_AfterCloseReturnsClosedError(t *testing.T) {
	rPipe, wPipe := io.Pipe()
	defer wPipe.Close()
	defer rPipe.Close()
	bs := NewBidirectionalService(rPipe, wPipe)
	bs.Start()

	// Close — flips bs.closed.
	if err := bs.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Both reply paths must now reject immediately. Without the fix,
	// SendResponse would try writeCh <- frame and (a) succeed if the
	// channel still had room (frame dropped on drain) or (b) hit the
	// stopCh branch with a different message.
	if err := bs.SendResponse(1, map[string]string{"ok": "true"}); err == nil {
		t.Fatalf("SendResponse after Close: expected error, got nil")
	} else if !strings.Contains(err.Error(), "service closed") {
		t.Fatalf("SendResponse after Close: expected /service closed/, got %v", err)
	}
	if err := bs.SendErrorResponse(1, "boom"); err == nil {
		t.Fatalf("SendErrorResponse after Close: expected error, got nil")
	} else if !strings.Contains(err.Error(), "service closed") {
		t.Fatalf("SendErrorResponse after Close: expected /service closed/, got %v", err)
	}
}

// Regression for M4: Close() must cancel the service-scoped context so
// in-flight inbound handlers observe shutdown and can abort. Previously
// handlers received `context.Background()` and bs.Close + bs.Wait would
// either deadlock (waiting on bs.wg for the handler goroutine to finish)
// or let a slow handler continue file I/O after the caller had given up.
//
// Test wiring note: the service reads from `r` and writes to `w`. We
// drive the input by writing to `inW` (which feeds `r`). To let the
// reader goroutine exit cleanly at end-of-test, we close `inW` (which
// the reader observes as EOF) BEFORE awaiting bs.Wait(). Without that
// the reader stays blocked in Read forever and bs.Wait() deadlocks —
// nothing to do with the fix under test.
func TestInboundHandler_ServiceCtxCancelOnClose(t *testing.T) {
	// service reads from `r` (we write to `inW`); writes to `outW`
	// (we read from `outR`, though for this test we don't inspect it).
	r, inW := io.Pipe()
	outR, outW := io.Pipe()
	// Drain outR continuously so the service's writer never blocks.
	go func() {
		buf := make([]byte, 1024)
		for {
			if _, err := outR.Read(buf); err != nil {
				return
			}
		}
	}()
	bs := NewBidirectionalService(r, outW)

	// Handler that publishes its received ctx so the test can observe
	// cancellation. Blocks until ctx is cancelled or 2s elapses.
	gotCtx := make(chan context.Context, 1)
	finished := make(chan struct{})
	bs.SetInboundHandler(InboundHandlerFunc(func(ctx context.Context, _ *Message) (interface{}, error) {
		gotCtx <- ctx
		select {
		case <-ctx.Done():
			// expected — Close cancelled us
		case <-time.After(2 * time.Second):
			t.Errorf("handler ctx never cancelled within 2s after Close")
		}
		close(finished)
		return map[string]bool{"ok": true}, nil
	}))
	bs.Start()

	// Push an inbound request frame. Handler will fire on its arrival.
	frame, err := encodeFrame(Message{Kind: "lint", ID: 42, Data: nil})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	go func() { _, _ = inW.Write(frame) }()

	// Wait until handler actually started (received its ctx).
	var handlerCtx context.Context
	select {
	case handlerCtx = <-gotCtx:
	case <-time.After(2 * time.Second):
		t.Fatal("handler never started")
	}
	if handlerCtx.Err() != nil {
		t.Fatalf("handler ctx already cancelled before Close: %v", handlerCtx.Err())
	}

	// Close the service. The handler's ctx must be cancelled within a
	// small slack window — not blocked on the 2s timeout inside.
	closeAt := time.Now()
	if err := bs.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	select {
	case <-finished:
		elapsed := time.Since(closeAt)
		if elapsed > 500*time.Millisecond {
			t.Errorf("handler took %v to observe Close (expected <500ms)", elapsed)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("handler never finished after Close")
	}

	// Close input/output pipes so the reader/writer goroutines see EOF
	// and exit; otherwise bs.Wait() blocks forever on the reader stuck
	// in Read(). This is test plumbing, not part of the fix's contract.
	_ = inW.Close()
	_ = outW.Close()
	bs.Wait()
}

// InboundHandlerFunc adapts a plain function to InboundHandler.
type InboundHandlerFunc func(context.Context, *Message) (interface{}, error)

func (f InboundHandlerFunc) Handle(ctx context.Context, m *Message) (interface{}, error) {
	return f(ctx, m)
}
