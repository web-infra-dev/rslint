package lsp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// newTestServer creates a minimal Server for handler unit tests.
// Session is nil so session calls are safely skipped via nil guards.
func newTestServer() *Server {
	return &Server{
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
		TextDocument: lsproto.VersionedTextDocumentIdentifier{Uri: uri},
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
