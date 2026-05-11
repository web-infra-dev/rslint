package ipc

import (
	"testing"
)

// #12 regression: routeResponse delivered the response AFTER unlocking
// pendingMu, leaving a window where a concurrent failAllPending (shutdown /
// EOF) could fill the (buffered-1) channel with its nil sentinel — dropping
// the real response and surfacing a false "peer closed" nil to the waiter.
// The fix claims (deletes) the pending entry under the lock so
// failAllPending can no longer touch it.
func TestRouteResponse_ClaimsEntryUnderLock(t *testing.T) {
	bs := &BidirectionalService{pending: make(map[int64]chan *Message)}
	id := int64(7)
	ch := make(chan *Message, 1)
	bs.pending[id] = ch

	msg := &Message{ID: 7}
	bs.routeResponse(msg)

	if _, stillPending := bs.pending[id]; stillPending {
		t.Errorf("routeResponse must delete the pending entry under the lock; still present")
	}
	select {
	case got := <-ch:
		if got != msg {
			t.Errorf("waiter received %v, want the routed message", got)
		}
	default:
		t.Errorf("routeResponse did not deliver the response to the waiter")
	}
}

// #12 (stronger, deterministic): routeResponse claims its pending entry
// UNDER the lock, so a subsequent failAllPending (shutdown) finds nothing
// for that id and cannot ALSO push its nil sentinel onto the same waiter.
//
// The waiter channel is buffered to 2 here purely so a (buggy)
// double-process is OBSERVABLE as two values, instead of being silently
// swallowed by production's buffered-1 non-blocking send. Running the two
// methods sequentially makes this deterministic: revert the delete out of
// the lock and the entry survives routeResponse, failAllPending pushes a
// second value, len(vals) == 2, and this test fails — no timing/race
// dependence.
func TestRouteResponse_FailAllPendingCannotDoubleProcess(t *testing.T) {
	bs := &BidirectionalService{pending: make(map[int64]chan *Message)}
	id := int64(1)
	ch := make(chan *Message, 2) // headroom to expose a double-process
	bs.pending[id] = ch
	msg := &Message{ID: 1}

	bs.routeResponse(msg)                           // claims + delivers under the lock
	bs.failAllPending() // must find nothing for id

	var vals []*Message
	for {
		select {
		case v := <-ch:
			vals = append(vals, v)
		default:
			goto done
		}
	}
done:
	if len(vals) != 1 {
		t.Fatalf("entry must be claimed exactly once; got %d values %v (a value beyond the first means failAllPending double-processed an entry routeResponse had already taken)", len(vals), vals)
	}
	if vals[0] != msg {
		t.Errorf("delivered value = %v, want the routed message", vals[0])
	}
}
