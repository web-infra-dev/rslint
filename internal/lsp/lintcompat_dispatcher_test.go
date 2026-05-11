package lsp

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/jsonrpc"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"

	"github.com/web-infra-dev/rslint/internal/linter"
)

// fakeLSPClient drives a minimal `Server.outgoingQueue` / response
// pipeline so we can exercise the LSP-based compat dispatcher without a
// real LanguageClient. The Server itself is constructed with only the
// fields the dispatcher touches (outgoingQueue, pendingServerRequests*,
// clientSeq).
//
// The fake client runs as a goroutine inside the test. It reads from
// `outgoingQueue`, dispatches by method, and writes responses back via
// `Server.pendingServerRequests[id]` channels — mirroring the real
// flow exactly.
type fakeLSPClient struct {
	server *Server

	// requestHandler is invoked for every server→client request.
	// Return (result, nil) → a normal JSON-RPC response is delivered.
	// Return (_, err)      → an error response is delivered.
	// Return (nil, nil)    → no response is sent (used to simulate a
	//                       client that's processing slowly; pair with
	//                       cancellation expectations).
	requestHandler func(method lsproto.Method, params any, id *jsonrpc.ID) (any, error)

	// cancelObserved fires when the fake client reads a `$/cancelRequest`
	// notification from outgoingQueue. The value is the cancel target id
	// as a string (since we mint string IDs throughout).
	cancelObserved chan string

	stopped atomic.Bool
}

// errFakeNoReply, returned by a fake request handler, tells the fake
// client loop to send NO reply (simulating a slow/busy client). The
// dispatcher's ctx timeout / cancel is what rescues the caller.
var errFakeNoReply = errors.New("fake client: no reply")

func newFakeLSPClient(t *testing.T, handler func(method lsproto.Method, params any, id *jsonrpc.ID) (any, error)) (*Server, *fakeLSPClient) {
	t.Helper()
	s := &Server{
		outgoingQueue:         make(chan *lsproto.Message, 16),
		pendingServerRequests: make(map[jsonrpc.ID]chan *lsproto.ResponseMessage),
	}
	fc := &fakeLSPClient{
		server:         s,
		requestHandler: handler,
		cancelObserved: make(chan string, 4),
	}
	go fc.run(t)
	t.Cleanup(func() {
		fc.stopped.Store(true)
		// Drain a pending message slot so the goroutine's `<-` unblocks.
		select {
		case s.outgoingQueue <- nil:
		default:
		}
	})
	return s, fc
}

func (fc *fakeLSPClient) run(t *testing.T) {
	t.Helper()
	for {
		msg, ok := <-fc.server.outgoingQueue
		if !ok || fc.stopped.Load() {
			return
		}
		if msg == nil {
			// nil sentinel used by Cleanup to unblock this loop.
			return
		}
		// We only care about request- or notification-shaped outgoing
		// messages here; AsResponse panics if the message is a request,
		// so distinguish via the framework's Kind enum implicitly by
		// always trying AsRequest first.
		req := msg.AsRequest()
		if req.ID == nil {
			// Notification — typically $/cancelRequest for our purposes.
			if req.Method == lsproto.MethodCancelRequest {
				if cp, ok := req.Params.(*lsproto.CancelParams); ok && cp.Id.String != nil {
					select {
					case fc.cancelObserved <- *cp.Id.String:
					default:
					}
				}
			}
			continue
		}
		// Request — dispatch via handler.
		result, err := fc.requestHandler(req.Method, req.Params, req.ID)
		if errors.Is(err, errFakeNoReply) {
			// Explicit no-reply — simulate a slow/busy client. The
			// dispatcher's ctx timeout / cancel rescues the caller.
			continue
		}
		var resp *lsproto.ResponseMessage
		if err != nil {
			resp = &lsproto.ResponseMessage{
				ID: req.ID,
				Error: &jsonrpc.ResponseError{
					Code:    int32(lsproto.ErrorCodeInternalError),
					Message: err.Error(),
				},
			}
		} else {
			resp = &lsproto.ResponseMessage{ID: req.ID, Result: result}
		}
		fc.server.pendingServerRequestsMu.Lock()
		ch, ok := fc.server.pendingServerRequests[*req.ID]
		if ok {
			delete(fc.server.pendingServerRequests, *req.ID)
		}
		fc.server.pendingServerRequestsMu.Unlock()
		if ch != nil {
			ch <- resp
		}
	}
}

// TestLintCompatLSPDispatcher_RoundTripsResults verifies the happy
// path: a CompatBatch with one file produces a single LSP request, the
// fake client returns a typed result, and the dispatcher decodes it
// into the linter's CompatFileResult shape.
func TestLintCompatLSPDispatcher_RoundTripsResults(t *testing.T) {
	wantPath := "/proj/file.ts"
	s, _ := newFakeLSPClient(t, func(method lsproto.Method, params any, id *jsonrpc.ID) (any, error) {
		if method != MethodRslintLintCompatBatch {
			t.Errorf("unexpected method %q", method)
			return nil, fmt.Errorf("unexpected method %q", method)
		}
		// Encode a canned response — matches what a real client would
		// produce by running its WorkerPool. Returning a typed struct
		// also exercises the marshal/remarshal path in the dispatcher.
		return LintCompatBatchResult{
			Results: []linter.CompatFileResult{
				{
					FilePath: wantPath,
					Diagnostics: []linter.CompatDiagnostic{
						{
							RuleName: "foo/bar",
							Message:  "test diagnostic",
							StartPos: 0,
							EndPos:   4,
						},
					},
				},
			},
		}, nil
	})

	dispatch := newLintCompatLSPDispatcher(s)
	results, err := dispatch(context.Background(), linter.CompatBatch{
		Files: []linter.CompatLintFile{{Path: wantPath}},
		Rules: map[string]linter.CompatRuleConfig{"foo/bar": {Options: []interface{}{}}},
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].FilePath != wantPath {
		t.Errorf("filePath mismatch: got %q want %q", results[0].FilePath, wantPath)
	}
	if len(results[0].Diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(results[0].Diagnostics))
	}
	if results[0].Diagnostics[0].RuleName != "foo/bar" {
		t.Errorf("ruleName mismatch: got %q", results[0].Diagnostics[0].RuleName)
	}
}

// TestLintCompatLSPDispatcher_EmptyBatchSkipsRoundTrip verifies the
// fast path: an empty batch returns (nil, nil) without sending an LSP
// request. Important because LSP `initialize` runs before the dispatcher
// is wired and an empty batch from the linter must not deadlock.
func TestLintCompatLSPDispatcher_EmptyBatchSkipsRoundTrip(t *testing.T) {
	sent := atomic.Int32{}
	s, _ := newFakeLSPClient(t, func(_ lsproto.Method, _ any, _ *jsonrpc.ID) (any, error) {
		sent.Add(1)
		return nil, errors.New("handler must not be invoked for an empty batch")
	})
	dispatch := newLintCompatLSPDispatcher(s)
	results, err := dispatch(context.Background(), linter.CompatBatch{})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty batch, got %v", results)
	}
	if sent.Load() != 0 {
		t.Errorf("expected zero requests sent, got %d", sent.Load())
	}
}

// TestLintCompatLSPDispatcher_PropagatesClientError verifies that a
// JSON-RPC error reply (`{ error: ... }`) surfaces as a Go error.
// The linter treats this as a batch-level failure and the affected
// files are marked compat-skipped — silent passthrough would leak
// missing diagnostics into the editor's view.
func TestLintCompatLSPDispatcher_PropagatesClientError(t *testing.T) {
	s, _ := newFakeLSPClient(t, func(_ lsproto.Method, _ any, _ *jsonrpc.ID) (any, error) {
		return nil, &handlerError{message: "plugin import failed: missing eslint-plugin-foo"}
	})
	dispatch := newLintCompatLSPDispatcher(s)
	_, err := dispatch(context.Background(), linter.CompatBatch{
		Files: []linter.CompatLintFile{{Path: "/x.ts"}},
	})
	if err == nil {
		t.Fatal("expected error from client-side failure, got nil")
	}
}

// TestLintCompatLSPDispatcher_DetectsCardinalityMismatch verifies the
// "result count mismatch" guard. A client that returns N-1 results for
// N input files is buggy — we MUST reject rather than silently drop
// the last file's diagnostics. The previous architecture (sidecar)
// had no such guard; that was a class of bugs we shouldn't bring
// forward.
func TestLintCompatLSPDispatcher_DetectsCardinalityMismatch(t *testing.T) {
	s, _ := newFakeLSPClient(t, func(_ lsproto.Method, _ any, _ *jsonrpc.ID) (any, error) {
		// Return ONE result for a TWO-file batch.
		return LintCompatBatchResult{
			Results: []linter.CompatFileResult{
				{FilePath: "/a.ts"},
			},
		}, nil
	})
	dispatch := newLintCompatLSPDispatcher(s)
	_, err := dispatch(context.Background(), linter.CompatBatch{
		Files: []linter.CompatLintFile{
			{Path: "/a.ts"},
			{Path: "/b.ts"},
		},
	})
	if err == nil {
		t.Fatal("expected cardinality mismatch error, got nil")
	}
}

// TestLintCompatLSPDispatcher_CancellationSendsCancelRequest verifies
// the cancel propagation: when the dispatcher's ctx cancels mid-flight,
// the server fires `$/cancelRequest` with the matching id. This is the
// core LSP-cancel-flow contract that lets the client's WorkerPool bail
// on per-keystroke supersession.
func TestLintCompatLSPDispatcher_CancellationSendsCancelRequest(t *testing.T) {
	// Handler never returns — simulates a slow client. The dispatcher's
	// ctx cancel must rescue the test.
	requestObserved := make(chan string, 1)
	s, fc := newFakeLSPClient(t, func(_ lsproto.Method, _ any, id *jsonrpc.ID) (any, error) {
		select {
		case requestObserved <- id.String():
		default:
		}
		return nil, errFakeNoReply // no reply → dispatcher waits on ctx
	})

	dispatch := newLintCompatLSPDispatcher(s)
	ctx, cancel := context.WithCancel(context.Background())

	type dispatchResult struct {
		results []linter.CompatFileResult
		err     error
	}
	done := make(chan dispatchResult, 1)
	go func() {
		r, err := dispatch(ctx, linter.CompatBatch{
			Files: []linter.CompatLintFile{{Path: "/x.ts"}},
		})
		done <- dispatchResult{r, err}
	}()

	var sentID string
	select {
	case sentID = <-requestObserved:
	case <-time.After(2 * time.Second):
		t.Fatal("fake client never received the lintCompatBatch request")
	}
	cancel()

	select {
	case canceled := <-fc.cancelObserved:
		if canceled != sentID {
			t.Errorf("$/cancelRequest target id mismatch: got %q want %q", canceled, sentID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("$/cancelRequest never reached the client")
	}

	select {
	case r := <-done:
		if r.err == nil {
			t.Errorf("expected dispatcher to surface ctx.Err(), got results=%v", r.results)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("dispatcher did not return after ctx cancel")
	}
}

// handlerError lets test handlers return typed errors without
// pulling in errors.New repeatedly. Same shape the real client would
// produce when its onRequest handler throws.
type handlerError struct{ message string }

func (e *handlerError) Error() string { return e.message }

// Compile-time check that LintCompatBatchResult round-trips through
// stdjson.Marshal/Unmarshal preserving Results length and per-element
// fields. Catches accidental json-tag drift on the LSP-facing types.
func TestLintCompatBatchResult_JSONRoundTrip(t *testing.T) {
	in := LintCompatBatchResult{
		Results: []linter.CompatFileResult{
			{FilePath: "/a", Diagnostics: []linter.CompatDiagnostic{{RuleName: "r"}}},
			{FilePath: "/b", ParseError: "boom"},
		},
	}
	blob, err := stdjson.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out LintCompatBatchResult
	if err := stdjson.Unmarshal(blob, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Results) != len(in.Results) {
		t.Fatalf("results length: got %d want %d", len(out.Results), len(in.Results))
	}
	if out.Results[0].FilePath != "/a" || out.Results[0].Diagnostics[0].RuleName != "r" {
		t.Errorf("results[0] roundtrip mismatch: %+v", out.Results[0])
	}
	if out.Results[1].ParseError != "boom" {
		t.Errorf("results[1].ParseError roundtrip mismatch: %q", out.Results[1].ParseError)
	}
}

// TestSendRequestWithClientCancel_CtxCancelBeforeEnqueue verifies that
// when ctx is already cancelled at call time and the outgoingQueue is
// full (writer wedged), the call returns ctx.Err() promptly rather
// than blocking forever.
//
// Regression: previously the `s.outgoingQueue <- req.Message()` send
// was unguarded — a wedged writer on a cancelled ctx would deadlock.
// The fix wraps the send in a select with ctx.Done.
func TestSendRequestWithClientCancel_CtxCancelBeforeEnqueue(t *testing.T) {
	// Server with a TINY outgoingQueue (capacity 0 = unbuffered) and
	// NO writer goroutine — any send will block forever unless ctx
	// rescues it.
	s := &Server{
		outgoingQueue:         make(chan *lsproto.Message),
		pendingServerRequests: make(map[jsonrpc.ID]chan *lsproto.ResponseMessage),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancelled

	done := make(chan error, 1)
	go func() {
		_, err := s.sendRequestWithClientCancel(ctx, MethodRslintLintCompatBatch, struct{}{})
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("sendRequestWithClientCancel blocked despite cancelled ctx — outgoingQueue send was not ctx-aware")
	}

	// Pending entry must be cleaned up on the ctx-cancelled-before-enqueue
	// path — otherwise a late spurious response (or just stale state)
	// would leak forever.
	s.pendingServerRequestsMu.Lock()
	leak := len(s.pendingServerRequests)
	s.pendingServerRequestsMu.Unlock()
	if leak != 0 {
		t.Errorf("pendingServerRequests leak: %d entries, expected 0", leak)
	}
}

// TestSendRequestWithClientCancel_LastSecondResponseDrain verifies the
// M2 fix: when ctx fires at almost the same instant a response arrives,
// the dispatcher drains the response from respChan instead of dropping
// it and returning ctx.Err().
//
// The trick is to construct a deterministic race: we drive the response
// path manually (no readerLoop), so we know exactly when the response
// hits the channel. Then we cancel the ctx and observe the result.
//
// Without the M2 fix, Go's select fairness would let ctx.Done win
// some fraction of the time even when respChan has a value; with the
// fix, the ctx.Done handler explicitly drains respChan first.
func TestSendRequestWithClientCancel_LastSecondResponseDrain(t *testing.T) {
	// Drive the test entirely off-thread: the goroutine that runs the
	// request, and a separate one that simulates a "response just
	// arrived". We synchronize so the response is enqueued BEFORE
	// ctx.Cancel — but the dispatcher's select might still pick ctx
	// over respChan since both are ready by the time it gets there.
	// The M2 fix's drain logic is what catches the value.

	s := &Server{
		outgoingQueue:         make(chan *lsproto.Message, 1),
		pendingServerRequests: make(map[jsonrpc.ID]chan *lsproto.ResponseMessage),
	}

	// Drain outgoingQueue so the send doesn't block during the test
	// setup window.
	go func() {
		for range s.outgoingQueue {
			// discard
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// We need to know the pending id after sendRequestWithClientCancel
	// has registered it. Probe via a polling loop in a helper goroutine.
	type result struct {
		raw any
		err error
	}
	dispatchDone := make(chan result, 1)
	go func() {
		raw, err := s.sendRequestWithClientCancel(ctx, MethodRslintLintCompatBatch, struct{}{})
		dispatchDone <- result{raw, err}
	}()

	// Wait for the dispatcher to register its pending entry.
	var respChan chan *lsproto.ResponseMessage
	var pendingID jsonrpc.ID
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s.pendingServerRequestsMu.Lock()
		for id, ch := range s.pendingServerRequests {
			respChan = ch
			pendingID = id
			break
		}
		s.pendingServerRequestsMu.Unlock()
		if respChan != nil {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	if respChan == nil {
		t.Fatal("pending entry never appeared")
	}

	// Atomically: enqueue a response AND cancel the ctx, in that
	// order, holding the mutex so the dispatcher's ctx.Done handler
	// can't acquire it until both are done. This guarantees that when
	// the dispatcher takes the lock in its ctx.Done branch, the
	// response is already sitting in respChan.
	s.pendingServerRequestsMu.Lock()
	respChan <- &lsproto.ResponseMessage{ID: &pendingID, Result: map[string]any{"results": []any{}}}
	// Mimic the readerLoop's cleanup so the drain branch sees
	// `stillPending = false` and falls into the "value already in
	// respChan" case A.
	close(respChan)
	delete(s.pendingServerRequests, pendingID)
	s.pendingServerRequestsMu.Unlock()
	cancel()

	select {
	case r := <-dispatchDone:
		if r.err != nil {
			t.Fatalf("expected response to be drained despite ctx cancel, got error: %v", r.err)
		}
		if r.raw == nil {
			t.Fatal("expected non-nil response value drained from respChan")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("dispatcher did not return after response+cancel")
	}
}
