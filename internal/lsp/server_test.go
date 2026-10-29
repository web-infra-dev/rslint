package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/jsonrpc"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config"
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
	s, uri, outgoing := newPushDiagnosticsPluginFixture(t, "var value = 1;\n",
		config.RslintConfig{{Rules: config.Rules{"no-var": "error"}}})
	s.requestQueue = make(chan *lsproto.RequestMessage, 3)
	s.requestQueue <- lsproto.TextDocumentDidOpenInfo.NewNotificationMessage(&lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "var value = 1;\n", Version: 1},
	})
	id, rawID := stringRequestID("queued-client-1")
	req := &lsproto.RequestMessage{
		ID:     id,
		Method: lsproto.MethodTextDocumentCodeAction,
		Params: &lsproto.CodeActionParams{
			TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		},
	}
	s.registerClientRequest(context.Background(), req)
	s.requestQueue <- req
	// readLoop registers before enqueueing, so cancellation cannot race ahead of
	// the request context even when dispatch has not started.
	s.cancelRequest(rawID)
	s.requestQueue <- lsproto.TextDocumentDidCloseInfo.NewNotificationMessage(&lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.dispatchLoop(ctx) }()

	select {
	case msg := <-outgoing:
		if msg.Kind != jsonrpc.MessageKindResponse {
			t.Fatal("canceled quick fix published diagnostics before the later close")
		}
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
	closed := nextTestLSPMessage(t, outgoing).AsRequest().Params.(*lsproto.PublishDiagnosticsParams)
	if closed.Uri != uri || len(closed.Diagnostics) != 0 {
		t.Fatal("close did not follow the canceled quick fix")
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

func startTestDispatchLoop(t *testing.T, s *Server) func() {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := s.dispatchLoop(ctx); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("dispatchLoop: %v", err)
		}
	}()
	stop := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("dispatch loop did not stop")
		}
	}
	t.Cleanup(stop)
	return stop
}

func nextTestLSPMessage(t *testing.T, queue <-chan *lsproto.Message) *lsproto.Message {
	t.Helper()
	select {
	case message := <-queue:
		return message
	case <-time.After(5 * time.Second):
		t.Fatal("no LSP message")
		return nil
	}
}

func TestDispatchLoopCoalescesDocumentNotifications(t *testing.T) {
	for _, scenario := range []string{"closed-preview", "incremental-changes", "reopen"} {
		t.Run(scenario, func(t *testing.T) {
			const original = "var value = 1;\n"
			const current = "const value = 1;\nvar latest = 2;\n"
			s, uri, outgoing := newPushDiagnosticsPluginFixture(t, original,
				config.RslintConfig{{Rules: config.Rules{"no-var": "error"}}})
			s.requestQueue = make(chan *lsproto.RequestMessage, 8)
			open := func(uri lsproto.DocumentUri, text string) {
				s.requestQueue <- lsproto.TextDocumentDidOpenInfo.NewNotificationMessage(&lsproto.DidOpenTextDocumentParams{
					TextDocument: &lsproto.TextDocumentItem{Uri: uri, LanguageId: "typescript", Version: 1, Text: text},
				})
			}
			closeDocument := func(uri lsproto.DocumentUri) {
				s.requestQueue <- lsproto.TextDocumentDidCloseInfo.NewNotificationMessage(&lsproto.DidCloseTextDocumentParams{
					TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
				})
			}
			switch scenario {
			case "closed-preview":
				preview := documentURIFromPath(tspath.ResolvePath(s.cwd, "src/preview.ts"))
				open(preview, original)
				closeDocument(preview)
				open(uri, current)
			case "incremental-changes":
				open(uri, original)
				s.requestQueue <- lsproto.TextDocumentDidChangeInfo.NewNotificationMessage(
					makePartialDidChangeParams(uri, 2, 0, 0, 0, 3, "const"))
				s.requestQueue <- lsproto.TextDocumentDidChangeInfo.NewNotificationMessage(
					makePartialDidChangeParams(uri, 3, 1, 0, 1, 0, "var latest = 2;\n"))
			case "reopen":
				open(uri, original)
				closeDocument(uri)
				open(uri, current)
			}
			stop := startTestDispatchLoop(t, s)
			for {
				message := nextTestLSPMessage(t, outgoing)
				params := message.AsRequest().Params.(*lsproto.PublishDiagnosticsParams)
				if len(params.Diagnostics) == 0 {
					continue // The close notification still clears the editor immediately.
				}
				if params.Uri != uri || len(params.Diagnostics) != 1 || params.Diagnostics[0].Range.Start.Line != 1 {
					t.Fatalf("published obsolete preview or intermediate content: %+v", params)
				}
				break
			}
			stop()
			if s.documents[uri] != current || len(s.documents) != 1 {
				t.Fatalf("document notifications were skipped or reordered: %v", s.documents)
			}
			if len(s.lintSchedule.pending) != 0 || len(outgoing) != 0 {
				t.Fatal("completed diagnostics retained duplicate work or publications")
			}
		})
	}
}

func TestDispatchLoopQuickFixPrecedesLaterClose(t *testing.T) {
	s, uri, outgoing := newPushDiagnosticsPluginFixture(t, "var value = 1;\n",
		config.RslintConfig{{Rules: config.Rules{"no-var": "error"}}})
	s.requestQueue = make(chan *lsproto.RequestMessage, 4)
	s.requestQueue <- lsproto.TextDocumentDidOpenInfo.NewNotificationMessage(&lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "var value = 1;\n", Version: 1},
	})
	id := jsonrpc.NewIDString("quick-fix")
	s.requestQueue <- &lsproto.RequestMessage{ID: id, Method: lsproto.MethodTextDocumentCodeAction,
		Params: &lsproto.CodeActionParams{
			TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
			Range:        lsproto.Range{End: lsproto.Position{Line: 1}},
			Context:      &lsproto.CodeActionContext{},
		},
	}
	s.requestQueue <- lsproto.TextDocumentDidCloseInfo.NewNotificationMessage(&lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
	})
	stop := startTestDispatchLoop(t, s)
	baseline := nextTestLSPMessage(t, outgoing).AsRequest().Params.(*lsproto.PublishDiagnosticsParams)
	if len(baseline.Diagnostics) != 1 {
		t.Fatal("quick fix did not flush current diagnostics")
	}
	response := nextTestLSPMessage(t, outgoing).AsResponse()
	if response.Error != nil || response.ID == nil || *response.ID != *id {
		t.Fatalf("quick fix response = %+v", response)
	}
	result := response.Result.(lsproto.CodeActionResponse)
	if result.CommandOrCodeActionArray == nil || len(*result.CommandOrCodeActionArray) == 0 {
		t.Fatal("later close suppressed the requested quick fix")
	}
	closed := nextTestLSPMessage(t, outgoing).AsRequest().Params.(*lsproto.PublishDiagnosticsParams)
	if closed.Uri != uri || len(closed.Diagnostics) != 0 {
		t.Fatal("close did not follow the quick fix")
	}
	stop()
}

func TestDispatchLoopFixAllDoesNotCancelInitialDiagnostics(t *testing.T) {
	s, uri, outgoing := newPushDiagnosticsPluginFixture(t, "const value = 1;\n",
		config.RslintConfig{{Rules: config.Rules{"no-var": "error"}}})
	s.requestQueue = make(chan *lsproto.RequestMessage, 4)
	s.requestQueue <- lsproto.TextDocumentDidOpenInfo.NewNotificationMessage(&lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "const value = 1;\n", Version: 1},
	})
	s.requestQueue <- &lsproto.RequestMessage{ID: jsonrpc.NewIDString("fix-all"), Method: lsproto.MethodTextDocumentCodeAction,
		Params: &lsproto.CodeActionParams{
			TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
			Context:      &lsproto.CodeActionContext{Only: &[]lsproto.CodeActionKind{codeActionKindSourceFixAllRslint}},
		},
	}
	stop := startTestDispatchLoop(t, s)
	response := nextTestLSPMessage(t, outgoing).AsResponse()
	if response.Error != nil {
		t.Fatalf("fix-all failed: %+v", response.Error)
	}
	result := response.Result.(lsproto.CodeActionResponse)
	if result.CommandOrCodeActionArray == nil || len(*result.CommandOrCodeActionArray) != 0 {
		t.Fatal("clean document unexpectedly needed edits")
	}
	baseline := nextTestLSPMessage(t, outgoing).AsRequest().Params.(*lsproto.PublishDiagnosticsParams)
	if baseline.Uri != uri || len(baseline.Diagnostics) != 0 {
		t.Fatalf("initial diagnostics = %+v", baseline)
	}
	stop()
}
