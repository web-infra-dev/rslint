package lsp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// newTestServer creates a minimal Server for handler unit tests.
// Session is nil so session calls are safely skipped via nil guards.
func newTestServer() *Server {
	return &Server{
		jsConfigs:       make(map[string]config.RslintConfig),
		documents:       make(map[lsproto.DocumentUri]string),
		diagnostics:     make(map[lsproto.DocumentUri][]rule.RuleDiagnostic),
		refreshCh:       make(chan struct{}, 1),
		debounceCh:      make(chan struct{}, 1),
		pendingLintURIs: make(map[lsproto.DocumentUri]struct{}),
	}
}

// helper to build a didChange params for full-sync mode
func makeDidChangeParams(uri lsproto.DocumentUri, version int32, text string) *lsproto.DidChangeTextDocumentParams {
	return &lsproto.DidChangeTextDocumentParams{
		TextDocument: lsproto.VersionedTextDocumentIdentifier{
			Uri:     uri,
			Version: version,
		},
		ContentChanges: []lsproto.TextDocumentContentChangePartialOrWholeDocument{
			{WholeDocument: &lsproto.TextDocumentContentChangeWholeDocument{Text: text}},
		},
	}
}

// ======== handleDidOpen tests ========

func TestHandleDidOpen(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	err := s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{
			Uri:  uri,
			Text: "const x = 1;",
		},
	})
	if err != nil {
		t.Fatalf("handleDidOpen failed: %v", err)
	}

	content, ok := s.documents[uri]
	if !ok {
		t.Fatal("document not stored after didOpen")
	}
	if content != "const x = 1;" {
		t.Errorf("document content = %q, want %q", content, "const x = 1;")
	}
}

func TestHandleDidOpen_Reopen(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	// First open
	_ = s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "old", Version: 1},
	})

	// Re-open with new content (e.g. closed and opened again)
	_ = s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "new", Version: 1},
	})

	if s.documents[uri] != "new" {
		t.Errorf("re-open should overwrite content, got %q", s.documents[uri])
	}
}

// ======== handleDidChange tests ========

func TestHandleDidChange(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "const x = 1;"

	err := s.handleDidChange(ctx, makeDidChangeParams(uri, 2, "const x = 2;"))
	if err != nil {
		t.Fatalf("handleDidChange failed: %v", err)
	}

	if s.documents[uri] != "const x = 2;" {
		t.Errorf("document content = %q, want %q", s.documents[uri], "const x = 2;")
	}
}

func TestHandleDidChange_EmptyChanges(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "original"

	err := s.handleDidChange(ctx, &lsproto.DidChangeTextDocumentParams{
		TextDocument:   lsproto.VersionedTextDocumentIdentifier{Uri: uri},
		ContentChanges: []lsproto.TextDocumentContentChangePartialOrWholeDocument{},
	})
	if err != nil {
		t.Fatalf("handleDidChange failed: %v", err)
	}

	if s.documents[uri] != "original" {
		t.Errorf("content changed unexpectedly to %q", s.documents[uri])
	}
}

func TestHandleDidChange_RapidSuccessiveChanges(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "v0"

	// Simulate rapid typing — many didChange events in quick succession.
	// Each overwrites the previous content (full sync mode).
	// After all changes, only the last content should remain.
	for i := 1; i <= 20; i++ {
		text := fmt.Sprintf("version %d", i)
		err := s.handleDidChange(ctx, makeDidChangeParams(uri, int32(i), text))
		if err != nil {
			t.Fatalf("change %d failed: %v", i, err)
		}
	}

	if s.documents[uri] != "version 20" {
		t.Errorf("after rapid changes: content = %q, want %q", s.documents[uri], "version 20")
	}
}

func TestHandleDidChange_UnopenedDocument(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/unknown.ts")

	// didChange on a document that was never opened — should not panic,
	// should still store the content.
	err := s.handleDidChange(ctx, makeDidChangeParams(uri, 1, "new content"))
	if err != nil {
		t.Fatalf("handleDidChange failed: %v", err)
	}

	if s.documents[uri] != "new content" {
		t.Errorf("content = %q, want %q", s.documents[uri], "new content")
	}
}

// ======== handleDidSave tests ========

func TestHandleDidSave(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "old content"

	savedText := "saved content"
	err := s.handleDidSave(ctx, &lsproto.DidSaveTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Text:         &savedText,
	})
	if err != nil {
		t.Fatalf("handleDidSave failed: %v", err)
	}

	if s.documents[uri] != "saved content" {
		t.Errorf("document content = %q, want %q", s.documents[uri], "saved content")
	}
}

func TestHandleDidSave_NilText(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "existing content"

	err := s.handleDidSave(ctx, &lsproto.DidSaveTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Text:         nil,
	})
	if err != nil {
		t.Fatalf("handleDidSave failed: %v", err)
	}

	// Content should remain unchanged when Text is nil
	if s.documents[uri] != "existing content" {
		t.Errorf("content changed unexpectedly to %q", s.documents[uri])
	}
}

// ======== handleDidClose tests ========

func TestHandleDidClose(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	s.documents[uri] = "some content"
	s.diagnostics[uri] = []rule.RuleDiagnostic{{RuleName: "test-rule"}}

	err := s.handleDidClose(ctx, &lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
	})
	if err != nil {
		t.Fatalf("handleDidClose failed: %v", err)
	}

	if _, ok := s.documents[uri]; ok {
		t.Error("document should be removed after didClose")
	}
	if _, ok := s.diagnostics[uri]; ok {
		t.Error("diagnostics should be removed after didClose")
	}
}

func TestHandleDidClose_NonexistentDocument(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/never-opened.ts")

	// Closing a document that was never opened should not panic
	err := s.handleDidClose(ctx, &lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
	})
	if err != nil {
		t.Fatalf("handleDidClose failed: %v", err)
	}

	if len(s.documents) != 0 {
		t.Errorf("documents map should remain empty, got %d entries", len(s.documents))
	}
}

func TestHandleDidClose_OtherDocumentsUntouched(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri1 := lsproto.DocumentUri("file:///project/a.ts")
	uri2 := lsproto.DocumentUri("file:///project/b.ts")

	s.documents[uri1] = "content a"
	s.documents[uri2] = "content b"
	s.diagnostics[uri1] = []rule.RuleDiagnostic{{RuleName: "rule-a"}}
	s.diagnostics[uri2] = []rule.RuleDiagnostic{{RuleName: "rule-b"}}

	_ = s.handleDidClose(ctx, &lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri1},
	})

	if _, ok := s.documents[uri1]; ok {
		t.Error("uri1 document should be removed")
	}
	if s.documents[uri2] != "content b" {
		t.Error("uri2 document should be untouched")
	}
	if _, ok := s.diagnostics[uri1]; ok {
		t.Error("uri1 diagnostics should be removed")
	}
	if len(s.diagnostics[uri2]) != 1 {
		t.Error("uri2 diagnostics should be untouched")
	}
}

// ======== lifecycle / integration tests ========

func TestDocumentLifecycle_OpenChangeClose(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	// Open
	_ = s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "v1", Version: 1},
	})
	if s.documents[uri] != "v1" {
		t.Fatalf("after open: content = %q, want %q", s.documents[uri], "v1")
	}

	// Change
	_ = s.handleDidChange(ctx, makeDidChangeParams(uri, 2, "v2"))
	if s.documents[uri] != "v2" {
		t.Fatalf("after change: content = %q, want %q", s.documents[uri], "v2")
	}

	// Save
	saved := "v2"
	_ = s.handleDidSave(ctx, &lsproto.DidSaveTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Text:         &saved,
	})
	if s.documents[uri] != "v2" {
		t.Fatalf("after save: content = %q, want %q", s.documents[uri], "v2")
	}

	// Close
	_ = s.handleDidClose(ctx, &lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
	})
	if _, ok := s.documents[uri]; ok {
		t.Fatal("after close: document should be removed")
	}
}

func TestMultipleDocuments_IndependentLifecycles(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uriA := lsproto.DocumentUri("file:///project/a.ts")
	uriB := lsproto.DocumentUri("file:///project/b.ts")

	// Open both
	_ = s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uriA, Text: "a-v1", Version: 1},
	})
	_ = s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uriB, Text: "b-v1", Version: 1},
	})

	// Change A only
	_ = s.handleDidChange(ctx, makeDidChangeParams(uriA, 2, "a-v2"))

	if s.documents[uriA] != "a-v2" {
		t.Errorf("A should be updated, got %q", s.documents[uriA])
	}
	if s.documents[uriB] != "b-v1" {
		t.Errorf("B should be unchanged, got %q", s.documents[uriB])
	}

	// Change B only
	_ = s.handleDidChange(ctx, makeDidChangeParams(uriB, 2, "b-v2"))

	if s.documents[uriA] != "a-v2" {
		t.Errorf("A should still be a-v2, got %q", s.documents[uriA])
	}
	if s.documents[uriB] != "b-v2" {
		t.Errorf("B should be updated, got %q", s.documents[uriB])
	}

	// Close A, B should remain
	_ = s.handleDidClose(ctx, &lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uriA},
	})

	if _, ok := s.documents[uriA]; ok {
		t.Error("A should be removed")
	}
	if s.documents[uriB] != "b-v2" {
		t.Errorf("B should still be b-v2, got %q", s.documents[uriB])
	}
}

func TestRapidChanges_VersionTracking(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	// Open
	_ = s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "initial", Version: 1},
	})

	// Simulate typing "hello" character by character — 5 rapid didChange events
	texts := []string{"h", "he", "hel", "hell", "hello"}
	for i, text := range texts {
		_ = s.handleDidChange(ctx, makeDidChangeParams(uri, int32(i+2), text))
	}

	// Only the final content should matter
	if s.documents[uri] != "hello" {
		t.Errorf("after rapid typing: content = %q, want %q", s.documents[uri], "hello")
	}
}

// newTestServerWithQueue creates a Server with an outgoingQueue so that
// PublishDiagnostics calls can be verified via the returned channel.
func newTestServerWithQueue() (*Server, chan *lsproto.Message) {
	queue := make(chan *lsproto.Message, 10)
	return &Server{
		jsConfigs:       make(map[string]config.RslintConfig),
		documents:       make(map[lsproto.DocumentUri]string),
		diagnostics:     make(map[lsproto.DocumentUri][]rule.RuleDiagnostic),
		outgoingQueue:   queue,
		backgroundCtx:   context.Background(),
		refreshCh:       make(chan struct{}, 1),
		debounceCh:      make(chan struct{}, 1),
		pendingLintURIs: make(map[lsproto.DocumentUri]struct{}),
	}, queue
}

func TestHandleDidClose_NoPublishWhenSessionNil(t *testing.T) {
	s, queue := newTestServerWithQueue()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	s.documents[uri] = "some content"
	s.diagnostics[uri] = []rule.RuleDiagnostic{{RuleName: "test-rule"}}

	err := s.handleDidClose(ctx, &lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
	})
	if err != nil {
		t.Fatalf("handleDidClose failed: %v", err)
	}

	// With session == nil, no PublishDiagnostics should be sent
	select {
	case msg := <-queue:
		t.Fatalf("unexpected message when session is nil: %v", msg)
	default:
		// good — no message sent
	}

	// State should still be cleaned up
	if _, ok := s.documents[uri]; ok {
		t.Error("document should be removed after didClose")
	}
	if _, ok := s.diagnostics[uri]; ok {
		t.Error("diagnostics should be removed after didClose")
	}
}

func TestPublishDiagnostics_EmptySlice(t *testing.T) {
	s, queue := newTestServerWithQueue()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	err := s.PublishDiagnostics(ctx, &lsproto.PublishDiagnosticsParams{
		Uri:         uri,
		Diagnostics: []*lsproto.Diagnostic{},
	})
	if err != nil {
		t.Fatalf("PublishDiagnostics failed: %v", err)
	}

	// Verify a message was sent to the queue
	select {
	case msg := <-queue:
		if msg == nil {
			t.Fatal("expected non-nil message")
		}
	default:
		t.Fatal("expected a message in the outgoing queue")
	}
}

// ======== debounce tests ========

func TestScheduleLint_AddsPendingURI(t *testing.T) {
	s := newTestServer()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	s.scheduleLint(uri)

	if _, ok := s.pendingLintURIs[uri]; !ok {
		t.Fatal("scheduleLint should add URI to pendingLintURIs")
	}

	// Stop the timer to prevent it from firing during test cleanup
	if s.lintTimer != nil {
		s.lintTimer.Stop()
	}
}

func TestScheduleLint_MultipleURIs(t *testing.T) {
	s := newTestServer()
	uriA := lsproto.DocumentUri("file:///project/a.ts")
	uriB := lsproto.DocumentUri("file:///project/b.ts")

	s.scheduleLint(uriA)
	s.scheduleLint(uriB)

	if _, ok := s.pendingLintURIs[uriA]; !ok {
		t.Error("uriA should be pending")
	}
	if _, ok := s.pendingLintURIs[uriB]; !ok {
		t.Error("uriB should be pending")
	}

	if s.lintTimer != nil {
		s.lintTimer.Stop()
	}
}

func TestScheduleLint_SignalsDebounceCh(t *testing.T) {
	s := newTestServer()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	s.scheduleLint(uri)

	// Wait for the debounce timer to fire (200ms + small buffer)
	time.Sleep(300 * time.Millisecond)

	select {
	case <-s.debounceCh:
		// good — signal received
	default:
		t.Fatal("expected a signal in debounceCh after debounce delay")
	}
}

func TestScheduleLint_ResetsTimer(t *testing.T) {
	s := newTestServer()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	// Schedule, wait 150ms, schedule again — the timer should reset
	s.scheduleLint(uri)
	time.Sleep(150 * time.Millisecond)
	s.scheduleLint(uri)

	// At 200ms from the first call the original timer would have fired,
	// but we reset it at 150ms so it shouldn't fire until 150+200=350ms.
	time.Sleep(100 * time.Millisecond) // total 250ms from start
	select {
	case <-s.debounceCh:
		t.Fatal("debounce timer should have been reset — signal came too early")
	default:
		// good
	}

	// Wait for the reset timer to fire
	time.Sleep(200 * time.Millisecond) // total 450ms from start
	select {
	case <-s.debounceCh:
		// good
	default:
		t.Fatal("expected signal after reset timer fires")
	}
}

func TestHandleDidClose_CleansPendingLint(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	// Simulate a pending lint for this URI (as if scheduleLint was called)
	s.documents[uri] = "const x = 1;"
	s.pendingLintURIs[uri] = struct{}{}

	err := s.handleDidClose(ctx, &lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
	})
	if err != nil {
		t.Fatalf("handleDidClose failed: %v", err)
	}

	// Pending lint URI should be cleaned up
	if _, ok := s.pendingLintURIs[uri]; ok {
		t.Error("pendingLintURIs should be cleaned up after didClose")
	}
	// Document and diagnostics should also be cleaned up
	if _, ok := s.documents[uri]; ok {
		t.Error("document should be removed after didClose")
	}
}

func TestScheduleLint_DebounceCh_Full(t *testing.T) {
	s := newTestServer()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	// Pre-fill debounceCh so the timer callback hits the default branch
	s.debounceCh <- struct{}{}

	s.scheduleLint(uri)

	// Wait for the timer to fire
	time.Sleep(300 * time.Millisecond)

	// Channel should still have exactly one signal (the pre-filled one)
	select {
	case <-s.debounceCh:
		// good — consumed the pre-filled signal
	default:
		t.Fatal("expected a signal in debounceCh")
	}

	// Channel should now be empty — the timer's signal was dropped
	select {
	case <-s.debounceCh:
		t.Fatal("expected empty debounceCh; timer should have dropped the signal")
	default:
		// good
	}

	// URI should still be in pending set
	if _, ok := s.pendingLintURIs[uri]; !ok {
		t.Error("URI should remain in pendingLintURIs even if debounceCh was full")
	}
}

func TestDebounce_CloseRace(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "const x = 1;"

	// Step 1: scheduleLint adds URI to pending and starts a timer
	s.scheduleLint(uri)
	if _, ok := s.pendingLintURIs[uri]; !ok {
		t.Fatal("URI should be pending after scheduleLint")
	}

	// Step 2: close the document before timer fires — clears pending
	_ = s.handleDidClose(ctx, &lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
	})
	if _, ok := s.pendingLintURIs[uri]; ok {
		t.Fatal("URI should be removed from pendingLintURIs after close")
	}

	// Step 3: wait for the timer to fire
	time.Sleep(300 * time.Millisecond)

	// debounceCh should have a signal (timer still fires)
	select {
	case <-s.debounceCh:
		// good — signal received
	default:
		t.Fatal("timer should still fire even after close")
	}

	// Step 4: simulate what the dispatch loop does — iterate pending URIs
	// pendingLintURIs is empty, so nothing should be linted
	for lintURI := range s.pendingLintURIs {
		t.Errorf("should not lint any URI, but found %s", lintURI)
	}
	clear(s.pendingLintURIs)
}

func TestHandleDidChange_UsesDebounce(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "const x = 1;"

	_ = s.handleDidChange(ctx, makeDidChangeParams(uri, 2, "const x = 2;"))

	// Content should be updated immediately
	if s.documents[uri] != "const x = 2;" {
		t.Errorf("content = %q, want %q", s.documents[uri], "const x = 2;")
	}

	// With nil session, no debounce should be scheduled
	if len(s.pendingLintURIs) != 0 {
		t.Error("no pending URIs expected when session is nil")
	}

	if s.lintTimer != nil {
		s.lintTimer.Stop()
	}
}

// ======== handleConfigUpdate tests ========

func TestHandleConfigUpdate_ClearsConfigPath(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	// Simulate a previously loaded JSON config
	s.rslintConfigPath = "/project/rslint.json"

	// Send a JS config update — should clear the JSON config path
	err := s.handleConfigUpdate(ctx, map[string]any{
		"configs": []any{
			map[string]any{
				"configDirectory": "file:///project",
				"entries": []any{
					map[string]any{
						"files": []string{"**/*.ts"},
						"rules": map[string]any{"no-console": "error"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("handleConfigUpdate failed: %v", err)
	}

	if s.rslintConfigPath != "" {
		t.Errorf("rslintConfigPath should be cleared after JS config update, got %q", s.rslintConfigPath)
	}

	cfg, ok := s.jsConfigs["file:///project"]
	if !ok {
		t.Fatal("jsConfigs should have file:///project entry")
	}
	if len(cfg) != 1 {
		t.Errorf("config should have 1 entry, got %d", len(cfg))
	}
}

func TestHandleConfigUpdate_MultipleConfigs(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	err := s.handleConfigUpdate(ctx, map[string]any{
		"configs": []any{
			map[string]any{
				"configDirectory": "file:///project",
				"entries": []any{
					map[string]any{"ignores": []string{"dist/**"}},
					map[string]any{
						"files": []string{"**/*.ts"},
						"rules": map[string]any{"no-console": "error"},
					},
				},
			},
			map[string]any{
				"configDirectory": "file:///project/packages/foo",
				"entries": []any{
					map[string]any{
						"files": []string{"**/*.ts"},
						"rules": map[string]any{"no-console": "warn"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("handleConfigUpdate failed: %v", err)
	}

	if len(s.jsConfigs) != 2 {
		t.Errorf("jsConfigs should have 2 entries, got %d", len(s.jsConfigs))
	}
	rootCfg := s.jsConfigs["file:///project"]
	if len(rootCfg) != 2 {
		t.Errorf("root config should have 2 entries, got %d", len(rootCfg))
	}
	fooCfg := s.jsConfigs["file:///project/packages/foo"]
	if len(fooCfg) != 1 {
		t.Errorf("foo config should have 1 entry, got %d", len(fooCfg))
	}
}

func TestHandleConfigUpdate_EmptyConfigs(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	// Pre-populate with old configs
	s.jsConfigs["file:///old"] = config.RslintConfig{}
	s.rslintConfigPath = "/project/rslint.json"

	// An explicitly empty configs array ({"configs":[]}) is a legitimate
	// "all JS configs deleted" signal — it SHOULD clear existing state.
	err := s.handleConfigUpdate(ctx, map[string]any{
		"configs": []any{},
	})
	if err != nil {
		t.Fatalf("handleConfigUpdate failed: %v", err)
	}

	if len(s.jsConfigs) != 0 {
		t.Errorf("jsConfigs should be empty after explicit empty configs, got %v", s.jsConfigs)
	}
	if s.rslintConfigPath != "" {
		t.Errorf("rslintConfigPath should be cleared, got %q", s.rslintConfigPath)
	}
}

func TestHandleConfigUpdate_ReplacesOldConfigs(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	// First update with old config
	s.jsConfigs["file:///old-project"] = config.RslintConfig{{}}

	// New update should completely replace
	err := s.handleConfigUpdate(ctx, map[string]any{
		"configs": []any{
			map[string]any{
				"configDirectory": "file:///new-project",
				"entries":         []any{map[string]any{"files": []string{"**/*.ts"}}},
			},
		},
	})
	if err != nil {
		t.Fatalf("handleConfigUpdate failed: %v", err)
	}

	if _, ok := s.jsConfigs["file:///old-project"]; ok {
		t.Error("old config should be replaced")
	}
	if _, ok := s.jsConfigs["file:///new-project"]; !ok {
		t.Error("new config should exist")
	}
}

func TestHandleConfigUpdate_MalformedPayload(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	// Pre-populate to verify malformed update doesn't corrupt state
	origConfig := config.RslintConfig{{Rules: map[string]any{"r": "error"}}}
	s.jsConfigs["file:///old"] = origConfig
	s.rslintConfigPath = "/project/rslint.json"

	tests := []struct {
		name   string
		params any
	}{
		{"nil params", nil},
		{"string params", "not an object"},
		{"number params", 42},
		{"missing configs field", map[string]any{"other": "data"}},
		{"configs is not an array", map[string]any{"configs": "bad"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.handleConfigUpdate(ctx, tt.params)
			// Some payloads may succeed (e.g. missing configs field → empty Configs slice),
			// others may fail. Either way, the server should not panic and existing
			// configs must remain intact.
			_ = err

			cfg, ok := s.jsConfigs["file:///old"]
			if !ok {
				t.Error("existing jsConfigs entry should be preserved after malformed payload")
			} else if len(cfg) != 1 || cfg[0].Rules["r"] != "error" {
				t.Errorf("existing jsConfigs entry was corrupted: %v", cfg)
			}
			if s.rslintConfigPath != "/project/rslint.json" {
				t.Errorf("rslintConfigPath should be preserved, got %q", s.rslintConfigPath)
			}
		})
	}
}

func TestHandleConfigUpdate_MissingConfigDirectory(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	// Entry with empty configDirectory — should still be stored (keyed by "")
	err := s.handleConfigUpdate(ctx, map[string]any{
		"configs": []any{
			map[string]any{
				"entries": []any{
					map[string]any{"rules": map[string]any{"no-console": "error"}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("handleConfigUpdate failed: %v", err)
	}

	// Empty key should exist
	if _, ok := s.jsConfigs[""]; !ok {
		t.Error("config with empty configDirectory should still be stored")
	}
}

func TestHandleConfigUpdate_InvalidEntriesType(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	// entries is a string instead of an array — Go JSON unmarshal will fail
	err := s.handleConfigUpdate(ctx, map[string]any{
		"configs": []any{
			map[string]any{
				"configDirectory": "file:///project",
				"entries":         "not-an-array",
			},
		},
	})
	if err != nil {
		// Expected: unmarshal error for entries field
		t.Logf("got expected error: %v", err)
	}
}

func TestJSConfigDeletedFallsBackToJSON(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	// 1. Start with a JSON config
	jsonRule := config.ConfigEntry{Rules: map[string]any{"no-debugger": "error"}}
	s.jsonConfig = config.RslintConfig{jsonRule}
	s.rslintConfigPath = "/project/rslint.json"

	// Verify JSON config is active
	cfg, _, _ := s.getConfigForURI("file:///project/src/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-debugger"] != "error" {
		t.Fatalf("expected JSON config, got %v", cfg)
	}

	// 2. JS config arrives — overrides JSON
	err := s.handleConfigUpdate(ctx, map[string]any{
		"configs": []any{
			map[string]any{
				"configDirectory": "file:///project",
				"entries": []any{
					map[string]any{"rules": map[string]any{"no-console": "warn"}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("handleConfigUpdate failed: %v", err)
	}

	cfg, _, _ = s.getConfigForURI("file:///project/src/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-console"] != "warn" {
		t.Fatalf("expected JS config, got %v", cfg)
	}
	if s.rslintConfigPath != "" {
		t.Fatalf("rslintConfigPath should be cleared, got %q", s.rslintConfigPath)
	}

	// 3. All JS configs deleted — send explicitly empty configs array.
	// This is a legitimate "all configs removed" signal from the extension.
	err = s.handleConfigUpdate(ctx, map[string]any{
		"configs": []any{},
	})
	if err != nil {
		t.Fatalf("handleConfigUpdate (empty) failed: %v", err)
	}

	// 4. No JS configs remain → should fall back to JSON config
	cfg, _, _ = s.getConfigForURI("file:///project/src/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-debugger"] != "error" {
		t.Errorf("after all JS configs deleted, should fall back to JSON config, got %v", cfg)
	}
}

// ======== getConfigForURI tests ========

func TestGetConfigForURI_ClosestJSConfig(t *testing.T) {
	s := newTestServer()

	rootRule := config.ConfigEntry{Rules: map[string]any{"no-console": "error"}}
	fooRule := config.ConfigEntry{Rules: map[string]any{"no-console": "warn"}}
	s.jsConfigs["file:///project"] = config.RslintConfig{rootRule}
	s.jsConfigs["file:///project/packages/foo"] = config.RslintConfig{fooRule}

	// File in foo should use foo's config, cwd = foo's directory
	fooCfg, fooCwd, _ := s.getConfigForURI("file:///project/packages/foo/src/index.ts")
	if len(fooCfg) != 1 || fooCfg[0].Rules["no-console"] != "warn" {
		t.Errorf("foo file should use foo config, got %v", fooCfg)
	}
	if fooCwd != "/project/packages/foo" {
		t.Errorf("foo cwd should be /project/packages/foo, got %q", fooCwd)
	}

	// File in bar should use root config, cwd = root directory
	barCfg, barCwd, _ := s.getConfigForURI("file:///project/packages/bar/src/index.ts")
	if len(barCfg) != 1 || barCfg[0].Rules["no-console"] != "error" {
		t.Errorf("bar file should use root config, got %v", barCfg)
	}
	if barCwd != "/project" {
		t.Errorf("bar cwd should be /project, got %q", barCwd)
	}
}

func TestGetConfigForURI_FallbackToJSON(t *testing.T) {
	s := newTestServer()

	jsonRule := config.ConfigEntry{Rules: map[string]any{"no-debugger": "error"}}
	s.jsonConfig = config.RslintConfig{jsonRule}

	// No JS configs — should fall back to JSON config; cwd = s.cwd
	cfg, cwd, _ := s.getConfigForURI("file:///project/src/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-debugger"] != "error" {
		t.Errorf("should fall back to JSON config, got %v", cfg)
	}
	if cwd != s.cwd {
		t.Errorf("fallback cwd should be s.cwd %q, got %q", s.cwd, cwd)
	}
}

func TestGetConfigForURI_JSConfigOverridesJSON(t *testing.T) {
	s := newTestServer()

	jsonRule := config.ConfigEntry{Rules: map[string]any{"no-debugger": "error"}}
	s.jsonConfig = config.RslintConfig{jsonRule}

	jsRule := config.ConfigEntry{Rules: map[string]any{"no-console": "warn"}}
	s.jsConfigs["file:///project"] = config.RslintConfig{jsRule}

	// JS config should take priority over JSON; cwd = JS config's dir
	cfg, cwd, _ := s.getConfigForURI("file:///project/src/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-console"] != "warn" {
		t.Errorf("JS config should override JSON, got %v", cfg)
	}
	if cwd != "/project" {
		t.Errorf("JS config cwd should be /project, got %q", cwd)
	}
}

func TestGetConfigForURI_NoConfig(t *testing.T) {
	s := newTestServer()

	cfg, _, _ := s.getConfigForURI("file:///project/src/index.ts")
	if cfg != nil {
		t.Errorf("should return nil when no config exists, got %v", cfg)
	}
}

func TestGetConfigForURI_DeeplyNestedFile(t *testing.T) {
	s := newTestServer()

	rootRule := config.ConfigEntry{Rules: map[string]any{"no-console": "error"}}
	s.jsConfigs["file:///project"] = config.RslintConfig{rootRule}

	// Deeply nested file should still find root config
	cfg, _, _ := s.getConfigForURI("file:///project/a/b/c/d/e/f/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-console"] != "error" {
		t.Errorf("deeply nested file should find root config, got %v", cfg)
	}
}

func TestGetConfigForURI_FileAtConfigDir(t *testing.T) {
	s := newTestServer()

	rootRule := config.ConfigEntry{Rules: map[string]any{"no-console": "error"}}
	s.jsConfigs["file:///project"] = config.RslintConfig{rootRule}

	// File in the same directory as config
	cfg, _, _ := s.getConfigForURI("file:///project/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-console"] != "error" {
		t.Errorf("file at config dir should use that config, got %v", cfg)
	}
}

func TestGetConfigForURI_MonorepoMultiplePackages(t *testing.T) {
	s := newTestServer()

	// Simulate a monorepo with root config + 3 sub-package configs
	rootRule := config.ConfigEntry{Rules: map[string]any{"no-console": "error", "no-debugger": "error"}}
	fooRule := config.ConfigEntry{Rules: map[string]any{"no-console": "warn"}}
	barRule := config.ConfigEntry{Rules: map[string]any{"no-console": "off"}}
	bazRule := config.ConfigEntry{Rules: map[string]any{"no-console": "error", "no-eval": "error"}}

	s.jsConfigs["file:///monorepo"] = config.RslintConfig{rootRule}
	s.jsConfigs["file:///monorepo/packages/foo"] = config.RslintConfig{fooRule}
	s.jsConfigs["file:///monorepo/packages/bar"] = config.RslintConfig{barRule}
	s.jsConfigs["file:///monorepo/packages/baz"] = config.RslintConfig{bazRule}

	tests := []struct {
		name     string
		uri      lsproto.DocumentUri
		wantRule string // expected value of "no-console"
		wantLen  int    // expected number of entries
	}{
		{"foo file uses foo config", "file:///monorepo/packages/foo/src/index.ts", "warn", 1},
		{"foo nested file uses foo config", "file:///monorepo/packages/foo/src/utils/helper.ts", "warn", 1},
		{"bar file uses bar config", "file:///monorepo/packages/bar/src/index.ts", "off", 1},
		{"baz file uses baz config", "file:///monorepo/packages/baz/src/index.ts", "error", 1},
		{"root file uses root config", "file:///monorepo/src/index.ts", "error", 1},
		{"unknown package uses root config", "file:///monorepo/packages/qux/src/index.ts", "error", 1},
		{"file outside monorepo has no config", "file:///other-project/src/index.ts", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, _ := s.getConfigForURI(tt.uri)
			if tt.wantLen == 0 {
				if cfg != nil {
					t.Errorf("expected nil config, got %v", cfg)
				}
				return
			}
			if len(cfg) != tt.wantLen {
				t.Fatalf("expected %d entries, got %d", tt.wantLen, len(cfg))
			}
			if cfg[0].Rules["no-console"] != tt.wantRule {
				t.Errorf("no-console = %v, want %v", cfg[0].Rules["no-console"], tt.wantRule)
			}
		})
	}
}

func TestGetConfigForURI_NestedConfigs(t *testing.T) {
	s := newTestServer()

	// 3-level nesting: root → packages/foo → packages/foo/sub
	s.jsConfigs["file:///project"] = config.RslintConfig{
		{Rules: map[string]any{"level": "root"}},
	}
	s.jsConfigs["file:///project/packages/foo"] = config.RslintConfig{
		{Rules: map[string]any{"level": "foo"}},
	}
	s.jsConfigs["file:///project/packages/foo/sub"] = config.RslintConfig{
		{Rules: map[string]any{"level": "sub"}},
	}

	// File in sub should use sub's config (closest)
	cfg, cwd, _ := s.getConfigForURI("file:///project/packages/foo/sub/src/index.ts")
	if cfg[0].Rules["level"] != "sub" {
		t.Errorf("sub file should use sub config, got %v", cfg[0].Rules["level"])
	}
	if cwd != "/project/packages/foo/sub" {
		t.Errorf("sub cwd should be /project/packages/foo/sub, got %q", cwd)
	}

	// File in foo (but not in sub) should use foo's config
	cfg, cwd, _ = s.getConfigForURI("file:///project/packages/foo/src/index.ts")
	if cfg[0].Rules["level"] != "foo" {
		t.Errorf("foo file should use foo config, got %v", cfg[0].Rules["level"])
	}
	if cwd != "/project/packages/foo" {
		t.Errorf("foo cwd should be /project/packages/foo, got %q", cwd)
	}

	// File at root should use root config
	cfg, cwd, _ = s.getConfigForURI("file:///project/src/index.ts")
	if cfg[0].Rules["level"] != "root" {
		t.Errorf("root file should use root config, got %v", cfg[0].Rules["level"])
	}
	if cwd != "/project" {
		t.Errorf("root cwd should be /project, got %q", cwd)
	}
}

func TestGetConfigForURI_WindowsURI(t *testing.T) {
	s := newTestServer()

	// Windows file URIs use forward slashes: file:///C:/Users/project
	s.jsConfigs["file:///C:/Users/project"] = config.RslintConfig{
		{Rules: map[string]any{"no-console": "error"}},
	}

	cfg, cwd, _ := s.getConfigForURI("file:///C:/Users/project/src/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-console"] != "error" {
		t.Errorf("Windows URI should match, got %v", cfg)
	}
	if cwd != "C:/Users/project" {
		t.Errorf("Windows cwd should be C:/Users/project, got %q", cwd)
	}
}

func TestGetConfigForURI_PercentEncodedPaths(t *testing.T) {
	s := newTestServer()

	// Config directory with spaces — VS Code sends percent-encoded URIs
	s.jsConfigs["file:///Users/John%20Doe/my%20project"] = config.RslintConfig{
		{Rules: map[string]any{"no-console": "error"}},
	}

	// File in a subdirectory — walk up should match the encoded config key
	cfg, cwd, _ := s.getConfigForURI("file:///Users/John%20Doe/my%20project/src/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-console"] != "error" {
		t.Errorf("Percent-encoded URI should match config, got %v", cfg)
	}
	// cwd should be decoded for filesystem use
	if cwd != "/Users/John Doe/my project" {
		t.Errorf("cwd should be decoded path, got %q", cwd)
	}

	// Deeply nested file — walk must traverse multiple encoded segments
	cfg2, cwd2, _ := s.getConfigForURI("file:///Users/John%20Doe/my%20project/src/components/deep/file.ts")
	if len(cfg2) != 1 || cfg2[0].Rules["no-console"] != "error" {
		t.Errorf("Deeply nested file in encoded path should match config, got %v", cfg2)
	}
	if cwd2 != "/Users/John Doe/my project" {
		t.Errorf("deep file cwd should be decoded, got %q", cwd2)
	}

	// File outside the config dir should fallback
	cfg3, _, _ := s.getConfigForURI("file:///Users/John%20Doe/other%20project/src/file.ts")
	if len(cfg3) != 0 {
		t.Errorf("File outside config dir should fallback to empty JSON config, got %v", cfg3)
	}
}

func TestGetConfigForURI_IsJSConfig(t *testing.T) {
	s := newTestServer()

	jsonRule := config.ConfigEntry{Rules: map[string]any{"no-debugger": "error"}}
	s.jsonConfig = config.RslintConfig{jsonRule}

	jsRule := config.ConfigEntry{Rules: map[string]any{"no-console": "warn"}}
	s.jsConfigs["file:///project"] = config.RslintConfig{jsRule}

	// JS config should return isJSConfig=true
	_, _, isJS := s.getConfigForURI("file:///project/src/index.ts")
	if !isJS {
		t.Error("Expected isJSConfig=true when JS config matches")
	}

	// File outside JS config range falls back to JSON → isJSConfig=false
	_, _, isJS = s.getConfigForURI("file:///other/src/index.ts")
	if isJS {
		t.Error("Expected isJSConfig=false when falling back to JSON config")
	}
}

// ======== uriToPath tests ========

func TestUriToPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Unix
		{"file:///home/user/project", "/home/user/project"},
		{"file:///project/src/index.ts", "/project/src/index.ts"},
		// Windows (uppercase and lowercase drive letters)
		{"file:///C:/Users/project", "C:/Users/project"},
		{"file:///D:/src/index.ts", "D:/src/index.ts"},
		{"file:///c:/Users/project", "c:/Users/project"},
		// Percent-encoded paths (spaces, CJK, VS Code colon encoding)
		{"file:///path%20with%20spaces/file.ts", "/path with spaces/file.ts"},
		{"file:///C%3A/Users/project", "C:/Users/project"},
		{"file:///project/%E4%B8%AD%E6%96%87/file.ts", "/project/中文/file.ts"},
		{"file:///Users/John%20Doe/my%20project/src/index.ts", "/Users/John Doe/my project/src/index.ts"},
		// Edge cases
		{"file:///", "/"},
		{"file://host/share", "/share"}, // UNC: authority is host, path is /share
		{"not-a-uri", "not-a-uri"},
		{"", ""},
	}

	for _, tt := range tests {
		got := uriToPath(lsproto.DocumentUri(tt.input))
		if got != tt.want {
			t.Errorf("uriToPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ======== uriDirname tests ========

func TestUriDirname(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"file:///project/src/index.ts", "file:///project/src"},
		{"file:///project/index.ts", "file:///project"},
		{"file:///project", "file:///project"}, // stops at authority
		{"file:///C:/Users/project/src", "file:///C:/Users/project"},
		{"file:///C:/Users", "file:///C:"},
		{"file:///C:", "file:///C:"}, // stops at authority
		{"", ""},                     // empty string
		{"file:///", "file:///"},     // root URI
	}

	for _, tt := range tests {
		got := uriDirname(tt.input)
		if got != tt.want {
			t.Errorf("uriDirname(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCloseAndReopen(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	// Open → change → close
	_ = s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "v1", Version: 1},
	})
	_ = s.handleDidChange(ctx, makeDidChangeParams(uri, 2, "v2"))
	s.diagnostics[uri] = []rule.RuleDiagnostic{{RuleName: "some-rule"}}
	_ = s.handleDidClose(ctx, &lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
	})

	// Verify fully cleaned up
	if _, ok := s.documents[uri]; ok {
		t.Fatal("document should be gone after close")
	}
	if _, ok := s.diagnostics[uri]; ok {
		t.Fatal("diagnostics should be gone after close")
	}

	// Re-open with different content
	_ = s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, Text: "fresh", Version: 1},
	})

	if s.documents[uri] != "fresh" {
		t.Errorf("after reopen: content = %q, want %q", s.documents[uri], "fresh")
	}
	// Diagnostics should not have carried over from previous session
	if _, ok := s.diagnostics[uri]; ok {
		t.Error("stale diagnostics should not reappear after reopen")
	}
}


// ======== tsConfigPaths lifecycle tests ========

func TestHandleConfigUpdate_RebuildsTsConfigPaths(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	ctx := context.Background()

	// Set stale tsConfigPaths
	s.tsConfigPaths = []string{"/old/tsconfig.json"}

	// Config update with no parserOptions.project and no tsconfig.json (mockFS has no files)
	// → ResolveTsConfigPaths returns nil → rebuildTsConfigPaths clears tsConfigPaths
	err := s.handleConfigUpdate(ctx, map[string]any{
		"configs": []any{
			map[string]any{
				"configDirectory": "file:///project",
				"entries": []any{
					map[string]any{
						"rules": map[string]any{"no-console": "error"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("handleConfigUpdate failed: %v", err)
	}

	// tsConfigPaths should be nil (mockFS has no tsconfig.json to auto-detect)
	if s.tsConfigPaths != nil {
		t.Errorf("expected tsConfigPaths nil after config update with no project, got %v", s.tsConfigPaths)
	}
}

func TestHandleConfigUpdate_EmptyConfigs_ClearsTsConfigPaths(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	ctx := context.Background()

	s.tsConfigPaths = []string{"/project/tsconfig.json"}

	err := s.handleConfigUpdate(ctx, map[string]any{
		"configs": []any{},
	})
	if err != nil {
		t.Fatalf("handleConfigUpdate failed: %v", err)
	}

	if s.tsConfigPaths != nil {
		t.Errorf("expected tsConfigPaths nil after empty config update, got %v", s.tsConfigPaths)
	}
}

func TestRebuildTsConfigPaths_MixedConfigsWithAndWithoutProject(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}

	// Config A has project, Config B doesn't (and no tsconfig.json to auto-detect)
	// → ResolveTsConfigPaths returns nil for B → rebuildTsConfigPaths should set nil (allow all)
	s.jsConfigs = map[string]config.RslintConfig{
		"file:///project-a": {
			{
				LanguageOptions: &config.LanguageOptions{
					ParserOptions: &config.ParserOptions{
						Project: []string{"./tsconfig.json"},
					},
				},
			},
		},
		"file:///project-b": {
			{
				Rules: config.Rules{"no-console": "error"},
			},
		},
	}

	s.rebuildTsConfigPaths()

	// Should be nil because config B has no project and no auto-detected tsconfig
	if s.tsConfigPaths != nil {
		t.Errorf("expected tsConfigPaths nil for mixed configs (some without project), got %v", s.tsConfigPaths)
	}
}

func TestRebuildTsConfigPaths_AllConfigsHaveProject(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{
		"/project-a/tsconfig.json": true,
		"/project-b/tsconfig.json": true,
	}}

	s.jsConfigs = map[string]config.RslintConfig{
		"file:///project-a": {
			{
				LanguageOptions: &config.LanguageOptions{
					ParserOptions: &config.ParserOptions{
						Project: []string{"./tsconfig.json"},
					},
				},
			},
		},
		"file:///project-b": {
			{
				LanguageOptions: &config.LanguageOptions{
					ParserOptions: &config.ParserOptions{
						Project: []string{"./tsconfig.json"},
					},
				},
			},
		},
	}

	s.rebuildTsConfigPaths()

	if s.tsConfigPaths == nil {
		t.Fatal("expected tsConfigPaths non-nil when all configs have project")
	}
	if len(s.tsConfigPaths) != 2 {
		t.Fatalf("expected 2 tsconfig paths, got %d: %v", len(s.tsConfigPaths), s.tsConfigPaths)
	}
	// Verify actual paths contain the expected tsconfig locations
	pathSet := make(map[string]bool)
	for _, p := range s.tsConfigPaths {
		pathSet[p] = true
	}
	if !pathSet["/project-a/tsconfig.json"] {
		t.Errorf("expected /project-a/tsconfig.json in paths, got %v", s.tsConfigPaths)
	}
	if !pathSet["/project-b/tsconfig.json"] {
		t.Errorf("expected /project-b/tsconfig.json in paths, got %v", s.tsConfigPaths)
	}
}

func TestRebuildTsConfigPaths_NoConfig(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}

	// No jsConfigs, no rslintConfigPath
	s.rebuildTsConfigPaths()

	if s.tsConfigPaths != nil {
		t.Errorf("expected tsConfigPaths nil when no config, got %v", s.tsConfigPaths)
	}
}
