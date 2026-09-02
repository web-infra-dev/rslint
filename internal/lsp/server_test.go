package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/jsonrpc"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// mockFS is a mock implementation of vfs.FS for testing. Only FileExists is properly implemented;
// other methods are stubbed with default implementations.
type mockFS struct {
	files map[string]bool // path -> exists
}

func (m *mockFS) FileExists(path string) bool {
	exists, found := m.files[path]
	if !found {
		return false
	}
	return exists
}

// Stubbed implementations of other vfs.FS interface methods for testing purposes
func (m *mockFS) UseCaseSensitiveFileNames() bool                             { return true }
func (m *mockFS) ReadFile(path string) (string, bool)                         { return "", false }
func (m *mockFS) WriteFile(path string, data string) error                    { return nil }
func (m *mockFS) AppendFile(path string, data string) error                   { return nil }
func (m *mockFS) Remove(path string) error                                    { return nil }
func (m *mockFS) Chtimes(path string, aTime time.Time, mTime time.Time) error { return nil }
func (m *mockFS) DirectoryExists(path string) bool                            { return false }
func (m *mockFS) GetAccessibleEntries(path string) vfs.Entries                { return vfs.Entries{} }
func (m *mockFS) Stat(path string) vfs.FileInfo                               { return nil }
func (m *mockFS) WalkDir(root string, walkFn vfs.WalkDirFunc) error           { return nil }
func (m *mockFS) Realpath(path string) string                                 { return path }

func TestDecodeParamsAcceptsRawJSONValue(t *testing.T) {
	data := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":null,"rootUri":"file:///tmp","capabilities":{}}}`)
	var msg lsproto.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	if msg.Kind != jsonrpc.MessageKindRequest {
		t.Fatalf("message kind = %v, want request", msg.Kind)
	}

	params, err := decodeParams[*lsproto.InitializeParams](msg.AsRequest())
	if err != nil {
		t.Fatalf("decode initialize params: %v", err)
	}
	if params == nil || params.Capabilities == nil {
		t.Fatalf("decoded params missing capabilities: %+v", params)
	}
}

func TestSetTraceNotificationIsAccepted(t *testing.T) {
	s := &Server{}
	for _, payload := range []string{
		`{"jsonrpc":"2.0","method":"$/setTrace","params":{"value":"off"}}`,
		`{"jsonrpc":"2.0","method":"$/setTrace","params":{"value":"messages"}}`,
		`{"jsonrpc":"2.0","method":"$/setTrace","params":{"value":"verbose"}}`,
	} {
		var msg lsproto.Message
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			t.Fatalf("unmarshal $/setTrace notification: %v", err)
		}
		if err := s.handleRequestOrNotification(context.Background(), msg.AsRequest()); err != nil {
			t.Fatalf("handle $/setTrace payload %s: %v", payload, err)
		}
	}
}

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
