package ipc

import (
	"context"
	"io"
	"testing"
	"time"
)

// Repro: writerLoop exits silently on write error. The next SendRequest
// enqueues into bs.writeCh (buffered, succeeds) but its frame never
// reaches the peer (writer gone). bs.closed stays false because Close
// isn't called from the writer error path, so failAllPending isn't
// fired either — SendRequest blocks on respCh forever.
//
// Expected outcome AFTER fix: SendRequest returns an error quickly
// (writer-side close cascades like reader-side does).
func TestWriterLoop_WriteError_DoesNotHangSenders(t *testing.T) {
	rA, wA := io.Pipe()
	rB, wB := io.Pipe()
	a := NewBidirectionalService(rA, wB)
	a.Start()

	// Close the OUTPUT pipe (the side the writer sees). First write
	// attempt will get io.ErrClosedPipe.
	_ = wB.Close()
	_ = rB.Close()

	// Drive a request. Use a 2s ctx — without the fix, even cancellation
	// works (ctx.Done() fires) but the bug is about NOT relying on the
	// caller's ctx; the transport should signal its own failure.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := a.SendRequest(ctx, KindLint, map[string]string{})
		done <- err
	}()

	select {
	case err := <-done:
		// Any error is fine — what matters is we didn't hang on respCh.
		// Specifically, the error should arrive WELL before the ctx
		// deadline (ctx is 5s; the transport should fail in tens of ms).
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("HANG: SendRequest still pending after 3s — writer-side close did not cascade to senders")
	}

	_ = a.Close()
	_ = rA.Close()
	_ = wA.Close()
}
