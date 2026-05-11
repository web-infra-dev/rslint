package main

import (
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	api "github.com/web-infra-dev/rslint/internal/api"
)

// Regression for M7: shutdownPeer's doc comment promised "No-op if a
// SIGINT/SIGTERM already fired", but the previous implementation
// unconditionally called bs.SendRequest. With the fix, the explicit
// state.signalled.Load() check short-circuits before any frame is
// encoded — verified here by snooping the output stream while
// shutdownPeer runs against a state with signalled=true.
func TestShutdownPeer_SignalledStateSkipsRequest(t *testing.T) {
	rIn, wIn := io.Pipe()
	rOut, wOut := io.Pipe()

	bs := api.NewBidirectionalService(rIn, wOut)
	bs.Start()

	// Track whether any byte landed on the output stream while
	// shutdownPeer runs. We must drain rOut continuously (writer side
	// blocks otherwise), and use atomic to publish the observation.
	var observed atomic.Bool
	drainerDone := make(chan struct{})
	go func() {
		defer close(drainerDone)
		buf := make([]byte, 1024)
		for {
			n, err := rOut.Read(buf)
			if n > 0 {
				observed.Store(true)
			}
			if err != nil {
				return
			}
		}
	}()

	// Arrange the state as if SIGINT had fired.
	state := &runCLIState{
		payloadCh: make(chan *initPayload, 1),
		once:      sync.Once{},
		signalled: atomic.Bool{},
	}
	state.signalled.Store(true)

	// shutdownPeer must short-circuit synchronously — no frame at all.
	done := make(chan struct{})
	go func() {
		shutdownPeer(bs, state)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdownPeer hung instead of short-circuiting")
	}

	// Brief observation window: with M7's fix nothing should be written.
	time.Sleep(150 * time.Millisecond)
	if observed.Load() {
		t.Fatal("shutdownPeer emitted bytes despite state.signalled=true")
	}

	// Test plumbing: close pipes BEFORE bs.Wait() so the reader
	// goroutine sees EOF and exits.
	_ = wIn.Close()
	_ = rIn.Close()
	_ = bs.Close()
	_ = wOut.Close()
	_ = rOut.Close()
	bs.Wait()
	<-drainerDone
}

// Counterpart: when state is NOT signalled, shutdownPeer must still
// push a request — otherwise M7's fix could regress to "always skip"
// and silently break clean shutdowns.
func TestShutdownPeer_NotSignalledSendsRequest(t *testing.T) {
	rIn, wIn := io.Pipe()
	rOut, wOut := io.Pipe()

	bs := api.NewBidirectionalService(rIn, wOut)
	bs.Start()

	wroteFrame := make(chan struct{}, 1)
	drainerDone := make(chan struct{})
	go func() {
		defer close(drainerDone)
		buf := make([]byte, 4096)
		for {
			n, err := rOut.Read(buf)
			if n > 0 {
				select {
				case wroteFrame <- struct{}{}:
				default:
				}
			}
			if err != nil {
				return
			}
		}
	}()

	state := &runCLIState{
		payloadCh: make(chan *initPayload, 1),
		once:      sync.Once{},
		signalled: atomic.Bool{},
	}
	// signalled stays false.

	done := make(chan struct{})
	go func() {
		shutdownPeer(bs, state)
		close(done)
	}()

	// At minimum a length-prefixed frame must hit rOut. With no real
	// peer the SendRequest inside shutdownPeer will time out (5s),
	// which is fine — we only assert that a frame WAS emitted.
	select {
	case <-wroteFrame:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("shutdownPeer did not emit a frame when state.signalled=false")
	}

	// Close pipes to unblock the reader / SendRequest waiter; let
	// shutdownPeer finish via its ctx timeout or pipe-close cascade.
	_ = wIn.Close()
	_ = rIn.Close()
	_ = bs.Close()
	_ = wOut.Close()
	_ = rOut.Close()

	select {
	case <-done:
	case <-time.After(7 * time.Second):
		t.Fatal("shutdownPeer never returned")
	}
	bs.Wait()
	<-drainerDone
}
