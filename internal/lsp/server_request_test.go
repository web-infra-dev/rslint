package lsp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/jsonrpc"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
)

func stringRequestID(value string) (*jsonrpc.ID, lsproto.IntegerOrString) {
	return jsonrpc.NewIDString(value), lsproto.IntegerOrString{String: &value}
}

func TestClientRequestCancelBeforeDispatch(t *testing.T) {
	s := &Server{}
	id, rawID := stringRequestID("client-1")
	req := &lsproto.RequestMessage{ID: id, Method: lsproto.Method("test/request")}
	requestCtx := s.registerClientRequest(context.Background(), req)
	s.cancelRequest(rawID)
	if !errors.Is(requestCtx.Err(), context.Canceled) {
		t.Fatalf("registered request context error = %v, want context.Canceled", requestCtx.Err())
	}

	s.pendingClientRequestsMu.Lock()
	_, pending := s.pendingClientRequests[*id]
	s.pendingClientRequestsMu.Unlock()
	if !pending {
		t.Fatal("canceled request was not registered for normal completion cleanup")
	}

	s.finishClientRequest(id)
	s.pendingClientRequestsMu.Lock()
	pendingCount := len(s.pendingClientRequests)
	s.pendingClientRequestsMu.Unlock()
	if pendingCount != 0 {
		t.Fatalf("finished request retained %d pending entries", pendingCount)
	}
}

func TestDispatchLoopHonorsCancelBeforeRequestRegistration(t *testing.T) {
	s, outgoing := newTestServerWithQueue()
	s.requestQueue = make(chan *lsproto.RequestMessage, 1)
	id, rawID := stringRequestID("queued-client-1")
	req := &lsproto.RequestMessage{
		ID:     id,
		Method: lsproto.MethodTextDocumentCodeAction,
		Params: &lsproto.CodeActionParams{
			TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///project/index.ts"},
		},
	}
	s.registerClientRequest(context.Background(), req)
	s.requestQueue <- req
	// readLoop registers before enqueueing, so cancellation cannot race ahead of
	// the request context even when dispatch has not started.
	s.cancelRequest(rawID)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.dispatchLoop(ctx) }()

	select {
	case msg := <-outgoing:
		resp := msg.AsResponse()
		if resp.ID == nil || *resp.ID != *id {
			t.Fatalf("response id = %v, want %s", resp.ID, id.String())
		}
		if resp.Error == nil || resp.Error.Code != int32(lsproto.ErrorCodeRequestCancelled) {
			t.Fatalf("response error = %+v, want request cancelled", resp.Error)
		}
	case <-time.After(time.Second):
		t.Fatal("dispatch loop did not cancel the request queued before registration")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("dispatch loop did not stop")
	}

	s.pendingClientRequestsMu.Lock()
	pendingCount := len(s.pendingClientRequests)
	s.pendingClientRequestsMu.Unlock()
	if pendingCount != 0 {
		t.Fatalf("request state leaked after dispatch: pending=%d", pendingCount)
	}
}

func TestSendRequestCancellationDoesNotBlockOnFullOutgoingQueue(t *testing.T) {
	queue := make(chan *lsproto.Message, 1)
	s := &Server{
		outgoingQueue:         queue,
		pendingServerRequests: make(map[jsonrpc.ID]chan *lsproto.ResponseMessage),
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		_, err := s.sendRequest(ctx, lsproto.Method("test/reverseRequest"), nil)
		result <- err
	}()

	var request *lsproto.Message
	select {
	case request = <-queue:
	case <-time.After(time.Second):
		t.Fatal("sendRequest did not queue the reverse request")
	}
	queue <- request
	cancel()

	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("sendRequest error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("sendRequest blocked while dropping cancellation on a full outgoing queue")
	}

	s.pendingServerRequestsMu.Lock()
	pendingCount := len(s.pendingServerRequests)
	s.pendingServerRequestsMu.Unlock()
	if pendingCount != 0 {
		t.Fatalf("canceled reverse request retained %d pending entries", pendingCount)
	}
	select {
	case got := <-queue:
		if got != request {
			t.Fatal("full outgoing queue was modified instead of dropping cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("full outgoing queue unexpectedly became empty")
	}
}
