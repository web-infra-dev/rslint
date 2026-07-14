package lsp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// newTestServer creates a minimal Server for handler unit tests.
// Session is nil so session calls are safely skipped via nil guards.
func newTestServer() *Server {
	return &Server{
		jsConfigs:              make(map[string]config.RslintConfig),
		documents:              make(map[lsproto.DocumentUri]string),
		diagnostics:            make(map[lsproto.DocumentUri][]rule.RuleDiagnostic),
		refreshCh:              make(chan struct{}, 1),
		debounceCh:             make(chan struct{}, 1),
		pendingLintURIs:        make(map[lsproto.DocumentUri]struct{}),
		pluginResultCh:         make(chan pluginLintResult, 16),
		docGeneration:          make(map[lsproto.DocumentUri]uint64),
		inflightPluginDispatch: make(map[lsproto.DocumentUri]*pluginDispatchHandle),
	}
}

func installJSConfigsForTest(s *Server, configs map[string]config.RslintConfig) {
	s.jsConfigs = configs
	s.jsConfigOwnerResolver = config.NewConfigOwnerResolver(configs, s.fs)
	s.jsUnavailableConfigs = make(map[string]struct{})
}

func documentURIFromPath(filePath string) lsproto.DocumentUri {
	uriPath := filepath.ToSlash(filePath)
	if len(uriPath) >= 2 && uriPath[1] == ':' {
		uriPath = "/" + uriPath
	}
	return lsproto.DocumentUri((&url.URL{Scheme: "file", Path: uriPath}).String())
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

func TestHandleDidChangeImmediatelyInvalidatesPluginWork(t *testing.T) {
	s, queue := newTestServerWithQueue()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "const oldValue = 1;"
	s.docGeneration[uri] = 4
	s.diagnostics[uri] = []rule.RuleDiagnostic{{RuleName: "native/old"}}

	dispatchCtx, cancel := context.WithCancel(context.Background())
	s.inflightPluginDispatch[uri] = &pluginDispatchHandle{cancel: cancel}

	if err := s.handleDidChange(context.Background(), makeDidChangeParams(uri, 2, "const newValue = 2;")); err != nil {
		t.Fatalf("handleDidChange failed: %v", err)
	}
	if s.docGeneration[uri] != 5 {
		t.Fatalf("document generation = %d, want 5", s.docGeneration[uri])
	}
	select {
	case <-dispatchCtx.Done():
	default:
		t.Fatal("didChange did not cancel the previous plugin dispatch")
	}

	s.mergePluginDiagnostics(pluginLintResult{
		uri:        uri,
		generation: 4,
		diags:      []rule.RuleDiagnostic{{RuleName: "plugin/stale"}},
	})
	if _, cached := s.diagnostics[uri]; cached {
		t.Fatal("didChange retained diagnostics from the previous document version")
	}
	select {
	case <-queue:
		t.Fatal("stale plugin result published after didChange")
	default:
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

func TestHandleDidSave_DoesNotOverwriteNewerDidChange(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")

	if err := s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{
			Uri:     uri,
			Version: 1,
			Text:    "older saved content",
		},
	}); err != nil {
		t.Fatalf("handleDidOpen failed: %v", err)
	}
	if err := s.handleDidChange(ctx, makeDidChangeParams(uri, 2, "newer unsaved content")); err != nil {
		t.Fatalf("handleDidChange failed: %v", err)
	}

	savedText := "older saved content"
	err := s.handleDidSave(ctx, &lsproto.DidSaveTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Text:         &savedText,
	})
	if err != nil {
		t.Fatalf("handleDidSave failed: %v", err)
	}

	if s.documents[uri] != "newer unsaved content" {
		t.Errorf("document content = %q, want newer didChange content", s.documents[uri])
	}
}

func TestHandleDidSave_DoesNotOpenUntrackedDocument(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/untracked.ts")
	savedText := "saved content"

	if err := s.handleDidSave(ctx, &lsproto.DidSaveTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Text:         &savedText,
	}); err != nil {
		t.Fatalf("handleDidSave failed: %v", err)
	}

	if _, open := s.documents[uri]; open {
		t.Error("didSave must not add an untracked document to the open-document mirror")
	}
}

func TestShouldForwardDidSave(t *testing.T) {
	matchingText := "current content"
	staleText := "older content"
	tests := []struct {
		name           string
		currentContent string
		open           bool
		savedText      *string
		want           bool
	}{
		{
			name:           "matching open document",
			currentContent: matchingText,
			open:           true,
			savedText:      &matchingText,
			want:           true,
		},
		{
			name:           "stale open document",
			currentContent: matchingText,
			open:           true,
			savedText:      &staleText,
			want:           false,
		},
		{
			name:      "untracked document",
			open:      false,
			savedText: &staleText,
			want:      true,
		},
		{
			name:           "client omitted text",
			currentContent: matchingText,
			open:           true,
			savedText:      nil,
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldForwardDidSave(tt.currentContent, tt.open, tt.savedText); got != tt.want {
				t.Fatalf("shouldForwardDidSave() = %v, want %v", got, tt.want)
			}
		})
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
		jsConfigs:              make(map[string]config.RslintConfig),
		documents:              make(map[lsproto.DocumentUri]string),
		diagnostics:            make(map[lsproto.DocumentUri][]rule.RuleDiagnostic),
		outgoingQueue:          queue,
		backgroundCtx:          context.Background(),
		refreshCh:              make(chan struct{}, 1),
		debounceCh:             make(chan struct{}, 1),
		pendingLintURIs:        make(map[lsproto.DocumentUri]struct{}),
		pluginResultCh:         make(chan pluginLintResult, 16),
		docGeneration:          make(map[lsproto.DocumentUri]uint64),
		inflightPluginDispatch: make(map[lsproto.DocumentUri]*pluginDispatchHandle),
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

// ======== getConfigForURI tests ========

func TestNearestJSConfigKey_UsesFilesystemCaseSensitivity(t *testing.T) {
	s := newTestServer()
	s.fs = &caseInsensitiveLSPTestFS{mockFS: mockFS{files: map[string]bool{}}}
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"C:/Repo":              {{Rules: config.Rules{"root": "error"}}},
		"C:/Repo/Packages/App": {{Rules: config.Rules{"nested": "error"}}},
	})

	key, ok := s.nearestJSConfigKey("file:///c:/repo/packages/app/src/index.ts")
	if !ok || key != "C:/Repo/Packages/App" {
		t.Fatalf("nearest key = %q, %v; want nested config", key, ok)
	}
	cfg, _, isJS := s.getConfigForURI("file:///c:/repo/packages/app/src/index.ts")
	if !isJS || cfg[0].Rules["nested"] != "error" {
		t.Fatalf("getConfigForURI selected %v, isJS=%v", cfg, isJS)
	}
}

func TestNearestJSConfigKey_HandlesUNCAndFilesystemRoot(t *testing.T) {
	s := newTestServer()
	s.fs = &caseInsensitiveLSPTestFS{mockFS: mockFS{files: map[string]bool{}}}
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/":                         {{Rules: config.Rules{"root": "error"}}},
		"//Server/Share/Repository": {{Rules: config.Rules{"unc": "error"}}},
	})

	if key, ok := s.nearestJSConfigKey("file://server/share/repository/src/a.ts"); !ok || key != "//Server/Share/Repository" {
		t.Fatalf("UNC nearest key = %q, %v", key, ok)
	}
	if key, ok := s.nearestJSConfigKey("file:///outside.ts"); !ok || key != "/" {
		t.Fatalf("filesystem-root nearest key = %q, %v", key, ok)
	}
}

func TestGetConfigForURI_ClosestJSConfig(t *testing.T) {
	s := newTestServer()

	rootRule := config.ConfigEntry{Rules: map[string]any{"no-console": "error"}}
	fooRule := config.ConfigEntry{Rules: map[string]any{"no-console": "warn"}}
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/project":              {rootRule},
		"/project/packages/foo": {fooRule},
	})

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
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/project": {jsRule},
	})

	// JS config should take priority over JSON; cwd = JS config's dir
	cfg, cwd, _ := s.getConfigForURI("file:///project/src/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-console"] != "warn" {
		t.Errorf("JS config should override JSON, got %v", cfg)
	}
	if cwd != "/project" {
		t.Errorf("JS config cwd should be /project, got %q", cwd)
	}
}

func TestGetConfigForURI_EmptyRootBoundaryProtectsJSONFallback(t *testing.T) {
	s := newTestServer()
	s.jsonConfig = config.RslintConfig{{
		Rules: map[string]any{"no-debugger": "error"},
	}}
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/project": {},
		"/project/packages/app": {{
			Rules: map[string]any{"no-console": "error"},
		}},
	})

	rootCfg, rootCwd, rootIsJS := s.getConfigForURI("file:///project/src/index.ts")
	if len(rootCfg) != 0 || rootCwd != "/project" || !rootIsJS {
		t.Fatalf("root file must use empty JS boundary: cfg=%v cwd=%q isJS=%v", rootCfg, rootCwd, rootIsJS)
	}

	nestedCfg, nestedCwd, nestedIsJS := s.getConfigForURI("file:///project/packages/app/src/index.ts")
	if len(nestedCfg) != 1 || nestedCfg[0].Rules["no-console"] != "error" ||
		nestedCwd != "/project/packages/app" || !nestedIsJS {
		t.Fatalf("nested file must use nested JS config: cfg=%v cwd=%q isJS=%v", nestedCfg, nestedCwd, nestedIsJS)
	}

	outsideCfg, outsideCwd, outsideIsJS := s.getConfigForURI("file:///other/src/index.ts")
	if len(outsideCfg) != 1 || outsideCfg[0].Rules["no-debugger"] != "error" ||
		outsideCwd != s.cwd || outsideIsJS {
		t.Fatalf("file outside JS boundaries must retain JSON fallback: cfg=%v cwd=%q isJS=%v", outsideCfg, outsideCwd, outsideIsJS)
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
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/project": {rootRule},
	})

	// Deeply nested file should still find root config
	cfg, _, _ := s.getConfigForURI("file:///project/a/b/c/d/e/f/index.ts")
	if len(cfg) != 1 || cfg[0].Rules["no-console"] != "error" {
		t.Errorf("deeply nested file should find root config, got %v", cfg)
	}
}

func TestGetConfigForURI_FileAtConfigDir(t *testing.T) {
	s := newTestServer()

	rootRule := config.ConfigEntry{Rules: map[string]any{"no-console": "error"}}
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/project": {rootRule},
	})

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

	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/monorepo":              {rootRule},
		"/monorepo/packages/foo": {fooRule},
		"/monorepo/packages/bar": {barRule},
		"/monorepo/packages/baz": {bazRule},
	})

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
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/project": {
			{Rules: map[string]any{"level": "root"}},
		},
		"/project/packages/foo": {
			{Rules: map[string]any{"level": "foo"}},
		},
		"/project/packages/foo/sub": {
			{Rules: map[string]any{"level": "sub"}},
		},
	})

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

	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"C:/Users/project": {
			{Rules: map[string]any{"no-console": "error"}},
		},
	})

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

	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/Users/John Doe/my project": {
			{Rules: map[string]any{"no-console": "error"}},
		},
	})

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
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/project": {jsRule},
	})

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

func TestDocumentURIFromPath_WindowsDriveRoundTrip(t *testing.T) {
	const filePath = "C:/Users/Test User/project/index.ts"
	uri := documentURIFromPath(filePath)
	if uri != "file:///C:/Users/Test%20User/project/index.ts" {
		t.Fatalf("documentURIFromPath(%q) = %q", filePath, uri)
	}
	if got := uriToPath(uri); got != filePath {
		t.Fatalf("uriToPath(%q) = %q, want %q", uri, got, filePath)
	}
}

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
		{"file://host/share", "//host/share"},
		{"file://server/share/project/src/index.ts", "//server/share/project/src/index.ts"},
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

func TestRebuildTsConfigPaths_MixedConfigsWithAndWithoutProject(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{"/project-a/tsconfig.json": true}}

	// Config A has a project that resolves; Config B has neither a project
	// nor an auto-detectable tsconfig. The two must be tracked independently
	// so B's missing tsconfig does not disable filtering for A's files.
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/project-a": {
			{
				LanguageOptions: &config.LanguageOptions{
					ParserOptions: &config.ParserOptions{
						Project: []string{"./tsconfig.json"},
					},
				},
			},
		},
		"/project-b": {
			{
				Rules: config.Rules{"no-console": "error"},
			},
		},
	})

	s.rebuildTsConfigPaths()

	entryA := s.tsConfigPathsByConfig["/project-a"]
	if len(entryA) != 1 || entryA[0] != "/project-a/tsconfig.json" {
		t.Errorf("expected project-a to resolve to its tsconfig, got %v", entryA)
	}
	if entry, ok := s.tsConfigPathsByConfig["/project-b"]; !ok || entry != nil {
		t.Errorf("expected project-b entry present and nil (no tsconfig), got present=%v value=%v", ok, entry)
	}
	if s.tsConfigPaths != nil {
		t.Errorf("expected JSON fallback tsConfigPaths nil in JS-config mode, got %v", s.tsConfigPaths)
	}
}

func TestRebuildTsConfigPaths_AllConfigsHaveProject(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{
		"/project-a/tsconfig.json": true,
		"/project-b/tsconfig.json": true,
	}}

	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/project-a": {
			{
				LanguageOptions: &config.LanguageOptions{
					ParserOptions: &config.ParserOptions{
						Project: []string{"./tsconfig.json"},
					},
				},
			},
		},
		"/project-b": {
			{
				LanguageOptions: &config.LanguageOptions{
					ParserOptions: &config.ParserOptions{
						Project: []string{"./tsconfig.json"},
					},
				},
			},
		},
	})

	s.rebuildTsConfigPaths()

	entryA := s.tsConfigPathsByConfig["/project-a"]
	if len(entryA) != 1 || entryA[0] != "/project-a/tsconfig.json" {
		t.Errorf("expected project-a → /project-a/tsconfig.json, got %v", entryA)
	}
	entryB := s.tsConfigPathsByConfig["/project-b"]
	if len(entryB) != 1 || entryB[0] != "/project-b/tsconfig.json" {
		t.Errorf("expected project-b → /project-b/tsconfig.json, got %v", entryB)
	}
}

// Regression test: a nested config without any resolvable tsconfig must not
// affect type-info decisions for files under other configs.
// See https://github.com/web-infra-dev/rslint/issues/671 — the create-rstack
// workspace ships a `template-rslint/` starter directory with its own
// rslint.config.ts but no tsconfig.json, which used to flip the whole
// workspace's type-aware rules without checking the governing config's
// resolved tsconfigs.
func TestTsConfigPathsForURI_NestedConfigWithoutTsconfigDoesNotLeak(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{"/project/tsconfig.json": true}}

	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/project": {
			{
				LanguageOptions: &config.LanguageOptions{
					ParserOptions: &config.ParserOptions{
						Project: []string{"./tsconfig.json"},
					},
				},
			},
		},
		"/project/template-rslint": {
			{
				Rules: config.Rules{"no-console": "error"},
			},
		},
	})

	s.rebuildTsConfigPaths()

	// File under root config → root's resolved tsconfig.
	rootPaths := s.tsConfigPathsForURI("file:///project/test/skills.test.ts")
	if len(rootPaths) != 1 || rootPaths[0] != "/project/tsconfig.json" {
		t.Errorf("expected root-config file to see [/project/tsconfig.json], got %v", rootPaths)
	}

	// File under nested template config -> nil (no type info), scoped to this
	// config only; the root config's list above must remain unaffected.
	nestedPaths := s.tsConfigPathsForURI("file:///project/template-rslint/foo.ts")
	if nestedPaths != nil {
		t.Errorf("expected nested-config file to see nil tsconfig paths (no type info), got %v", nestedPaths)
	}
}

func TestTsConfigPathsForURI_JSONFallbackRemainsTypedWithNestedJSConfig(t *testing.T) {
	s := newTestServer()
	s.cwd = "/workspace"
	s.rslintConfigPath = "/workspace/rslint.json"
	s.fs = &mockFS{files: map[string]bool{
		"/workspace/tsconfig.json":              true,
		"/workspace/packages/app/tsconfig.json": true,
	}}
	s.jsonConfig = config.RslintConfig{{
		LanguageOptions: &config.LanguageOptions{
			ParserOptions: &config.ParserOptions{Project: []string{"./tsconfig.json"}},
		},
	}}
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/workspace/packages/app": {{
			LanguageOptions: &config.LanguageOptions{
				ParserOptions: &config.ParserOptions{Project: []string{"./tsconfig.json"}},
			},
		}},
	})

	s.rebuildTsConfigPaths()

	jsonPaths := s.tsConfigPathsForURI("file:///workspace/packages/lib/src/index.ts")
	if len(jsonPaths) != 1 || jsonPaths[0] != "/workspace/tsconfig.json" {
		t.Fatalf("expected JSON fallback tsconfig, got %v", jsonPaths)
	}
	jsPaths := s.tsConfigPathsForURI("file:///workspace/packages/app/src/index.ts")
	if len(jsPaths) != 1 || jsPaths[0] != "/workspace/packages/app/tsconfig.json" {
		t.Fatalf("expected nested JS config tsconfig, got %v", jsPaths)
	}
}

func TestReloadJSONFallbackWhileJSConfigIsActive(t *testing.T) {
	dir := t.TempDir()
	jsonConfigPath := filepath.Join(dir, "rslint.json")
	if err := os.WriteFile(jsonConfigPath, []byte(`[{"rules":{"no-debugger":"error"}}]`), 0o644); err != nil {
		t.Fatal(err)
	}

	s := newTestServer()
	s.cwd = dir
	s.fs = osvfs.FS()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/other": {{Rules: config.Rules{"no-console": "error"}}},
	})
	s.reloadConfigAndRelint()

	if s.rslintConfigPath != jsonConfigPath {
		t.Fatalf("expected JSON fallback path %q, got %q", jsonConfigPath, s.rslintConfigPath)
	}
	cfg, _, isJS := s.getConfigForURI("file:///workspace/src/index.ts")
	if isJS || len(cfg) != 1 || cfg[0].Rules["no-debugger"] != "error" {
		t.Fatalf("expected refreshed JSON fallback while JS config remains active, got isJS=%v config=%v", isJS, cfg)
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

// ======== runLintWithSession: ignored-file short-circuit ========

func TestIsLintableScriptFile_UsesDefaultLintExtensions(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"a.js", true},
		{"a.mjs", true},
		{"a.cjs", true},
		{"a.jsx", true},
		{"a.ts", true},
		{"a.mts", true},
		{"a.cts", true},
		{"a.tsx", true},
		{"style.css", false},
		{"data.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := lsproto.DocumentUri("file:///project/" + tt.name)
			if got := isLintableScriptFile(uri); got != tt.want {
				t.Fatalf("isLintableScriptFile(%q) = %v, want %v", uri, got, tt.want)
			}
		})
	}
}

func TestLSPActiveRulesForFile_RespectsFiles(t *testing.T) {
	config.RegisterAllRules()

	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	tsFile := tspath.NormalizePath(filepath.Join(srcDir, "matched.ts"))
	jsFile := tspath.NormalizePath(filepath.Join(srcDir, "outside.js"))
	for _, file := range []string{tsFile, jsFile} {
		if err := os.WriteFile(file, []byte("debugger;\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(dir, fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Target:  core.ScriptTargetESNext,
		Module:  core.ModuleKindESNext,
		AllowJs: core.TSTrue,
		CheckJs: core.TSTrue,
	}, []string{tsFile, jsFile}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptionsLenient: %v", err)
	}

	cfg := config.RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: config.Rules{"no-debugger": "error"},
		},
	}
	collect := func(file string) []rule.RuleDiagnostic {
		t.Helper()
		var diags []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: program,
			File:    file,
			GetRulesForFile: func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
				return lspActiveRulesForFile(cfg, sourceFile.FileName(), dir, false, true)
			},
			OnDiagnostic: func(d rule.RuleDiagnostic) {
				diags = append(diags, d)
			},
		})
		return diags
	}

	if got := collect(tsFile); len(got) != 1 {
		t.Fatalf("matching TS file should run no-debugger once, got %d diagnostics: %+v", len(got), got)
	}
	if got := collect(jsFile); len(got) != 0 {
		t.Fatalf("files-scope miss must not run LSP native rules, got %+v", got)
	}
}

func TestLSPFilesystemPathID_UsesCanonicalFilesystemIdentity(t *testing.T) {
	caseInsensitiveFS := &caseInsensitiveLSPTestFS{mockFS: mockFS{files: map[string]bool{}}}
	if lspFilesystemPathID("C:/Repo/TSConfig.json", caseInsensitiveFS) != lspFilesystemPathID("c:/repo/tsconfig.json", caseInsensitiveFS) {
		t.Fatal("case-insensitive filesystem path IDs must ignore casing")
	}
}

type exactCaseLSPProgramFS struct {
	vfs.FS
	files map[string]string
}

func (fs *exactCaseLSPProgramFS) UseCaseSensitiveFileNames() bool { return false }
func (fs *exactCaseLSPProgramFS) FileExists(filePath string) bool {
	if _, ok := fs.files[tspath.NormalizePath(filePath)]; ok {
		return true
	}
	return fs.FS.FileExists(filePath)
}
func (fs *exactCaseLSPProgramFS) ReadFile(filePath string) (string, bool) {
	if content, ok := fs.files[tspath.NormalizePath(filePath)]; ok {
		return content, true
	}
	return fs.FS.ReadFile(filePath)
}
func (fs *exactCaseLSPProgramFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if _, ok := fs.files[filePath]; ok {
		return filePath
	}
	return fs.FS.Realpath(filePath)
}

func TestSourceFileForPath_RejectsCaseFoldedDifferentFile(t *testing.T) {
	upper := "/repo/Source.ts"
	lower := "/repo/source.ts"
	fsys := &exactCaseLSPProgramFS{
		FS: osvfs.FS(),
		files: map[string]string{
			upper: "export const upper = 1;\n",
			lower: "export const lower = 2;\n",
		},
	}
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, []string{upper}, utils.CreateCompilerHost("/repo", fsys))
	if err != nil {
		t.Fatalf("CreateProgramFromOptionsLenient: %v", err)
	}
	if source := program.GetSourceFile(lower); source == nil || source.FileName() != upper {
		t.Fatalf("fixture must exercise case-folded Program lookup, got %v", source)
	}
	if source := sourceFileForPath(program, lower, fsys); source != nil {
		t.Fatalf("case-distinct target bound to %q", source.FileName())
	}
}

func TestSourceFileForPath_FindsProgramFileSymlinkFromRealTarget(t *testing.T) {
	sharedDir := t.TempDir()
	repoDir := t.TempDir()
	realTarget := filepath.Join(sharedDir, "shared.ts")
	linkedPath := filepath.Join(repoDir, "linked.ts")
	if err := os.WriteFile(realTarget, []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realTarget, linkedPath); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}

	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	linkedPath = tspath.NormalizePath(linkedPath)
	realTarget = tspath.NormalizePath(realTarget)
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, []string{linkedPath}, utils.CreateCompilerHost(repoDir, fsys))
	if err != nil {
		t.Fatalf("CreateProgramFromOptionsLenient: %v", err)
	}
	sourceName := ""
	for _, sourceFile := range program.GetSourceFiles() {
		if sourceFile.FileName() == linkedPath || sourceFile.FileName() == realTarget {
			sourceName = sourceFile.FileName()
			break
		}
	}
	if sourceName == "" {
		t.Fatal("Program does not contain the symlinked source")
	}
	if sourceName == realTarget {
		t.Skip("compiler canonicalized the file symlink before Program lookup")
	}
	if source := sourceFileForPath(program, realTarget, fsys); source == nil || source.FileName() != sourceName {
		t.Fatalf("real target did not bind to Program source %q: %v", sourceName, source)
	}
}

func TestSelectLintProgram_UsesDeclaredProjectOrderAndGapFallback(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "src", "index.ts")
	gapPath := filepath.Join(dir, "gap.ts")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{sourcePath, gapPath} {
		if err := os.WriteFile(file, []byte("export const value = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	firstConfig := filepath.Join(dir, "tsconfig.lint-first.json")
	secondConfig := filepath.Join(dir, "tsconfig.lint-second.json")
	for _, configPath := range []string{firstConfig, secondConfig} {
		if err := os.WriteFile(configPath, []byte(`{"files":["src/index.ts"]}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	ctx := context.Background()
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	s := NewServer(&ServerOptions{
		Cwd:                dir,
		FS:                 fsys,
		DefaultLibraryPath: bundled.LibPath(),
	})
	s.backgroundCtx = ctx
	s.initializeParams = &lsproto.InitializeParams{}
	if err := s.handleInitialized(ctx, &lsproto.InitializedParams{}); err != nil {
		t.Fatalf("initialize test session: %v", err)
	}
	defer s.session.Close()

	toURI := func(filePath string) lsproto.DocumentUri {
		return documentURIFromPath(filePath)
	}
	for _, file := range []string{sourcePath, gapPath} {
		uri := toURI(file)
		if err := s.handleDidOpen(ctx, &lsproto.DidOpenTextDocumentParams{
			TextDocument: &lsproto.TextDocumentItem{
				Uri:     uri,
				Version: 1,
				Text:    "export const value = 1;\n",
			},
		}); err != nil {
			t.Fatalf("open %s: %v", file, err)
		}
	}

	sourceURI := toURI(sourcePath)
	program, hasTypeInfo, err := selectLintProgram(
		sourceURI,
		s.session,
		ctx,
		[]string{secondConfig, firstConfig},
		fsys,
		s.newStandaloneLintProgramLoader(sourceURI),
	)
	if err != nil {
		t.Fatalf("select typed program: %v", err)
	}
	if !hasTypeInfo {
		t.Fatal("expected source to have type information")
	}
	if got := lspFilesystemPathID(program.Options().ConfigFilePath, fsys); got != lspFilesystemPathID(secondConfig, fsys) {
		t.Fatalf("expected first declared containing project %q, got %q", secondConfig, program.Options().ConfigFilePath)
	}
	secondConfigID := tspath.ToPath(fsys.Realpath(secondConfig), "", fsys.UseCaseSensitiveFileNames())
	if opened := s.session.Snapshot().ProjectCollection.ConfiguredProject(secondConfigID); opened != nil {
		t.Fatal("lint-only custom tsconfig must not be added to the Session's permanent API-open project set")
	}
	if _, err := s.defaultFixAllNativeLint(
		ctx,
		toURI(sourcePath),
		1,
		"export const value = 2;\n",
		config.RslintConfig{{}},
		dir,
		false,
		[]string{secondConfig},
	); err != nil {
		t.Fatalf("isolated fix pass: %v", err)
	}
	languageService, err := s.session.GetLanguageService(ctx, toURI(sourcePath))
	if err != nil {
		t.Fatalf("get language service after speculative fix: %v", err)
	}
	if got := languageService.GetProgram().GetSourceFile(tspath.NormalizePath(sourcePath)).Text(); got != "export const value = 1;\n" {
		t.Fatalf("speculative fix polluted Session overlay: got %q", got)
	}

	gapURI := toURI(gapPath)
	gapProgram, gapHasTypeInfo, err := selectLintProgram(
		gapURI,
		s.session,
		ctx,
		[]string{secondConfig, firstConfig},
		fsys,
		s.newStandaloneLintProgramLoader(gapURI),
	)
	if err != nil {
		t.Fatalf("select gap program: %v", err)
	}
	if gapHasTypeInfo {
		t.Fatal("file outside every declared project must not have type information")
	}
	if gapProgram.GetSourceFile(tspath.NormalizePath(gapPath)) == nil {
		t.Fatal("gap fallback program must contain the open file")
	}
}

func TestRunConfiguredLintForContent_SyntaxErrorSkipsRules(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "malformed.ts")
	const malformed = "debugger; const value = ;\n"
	if err := os.WriteFile(filePath, []byte(malformed), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newTestServer()
	s.cwd = dir
	s.fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	uri := documentURIFromPath(filePath)
	s.documents[uri] = malformed

	result, err := s.runConfiguredLintForContent(
		uri,
		context.Background(),
		malformed,
		config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}},
		dir,
		false,
		nil,
	)
	if err != nil {
		t.Fatalf("run lint: %v", err)
	}
	if !result.HasSyntaxErrors {
		t.Fatal("malformed file was not marked as having syntax errors")
	}
	if len(result.Diagnostics) == 0 || !strings.HasPrefix(result.Diagnostics[0].RuleName, "TypeScript(TS") {
		t.Fatalf("expected a TypeScript syntax diagnostic, got %+v", result.Diagnostics)
	}
	if result.Diagnostics[0].Origin != rule.DiagnosticOriginTypeScript {
		t.Fatalf("TypeScript syntax diagnostic has wrong origin: %+v", result.Diagnostics[0])
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.RuleName == "no-debugger" {
			t.Fatalf("rules ran for malformed file: %+v", result.Diagnostics)
		}
	}
	s.diagnostics[uri] = result.Diagnostics
	response, err := s.handleCodeAction(context.Background(), &lsproto.CodeActionParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Range: lsproto.Range{
			Start: lsproto.Position{Line: 0, Character: 0},
			End:   lsproto.Position{Line: 0, Character: uint32(len(malformed))},
		},
		Context: &lsproto.CodeActionContext{},
	})
	if err != nil {
		t.Fatalf("get syntax diagnostic code actions: %v", err)
	}
	if actions := response.CommandOrCodeActionArray; actions == nil || len(*actions) != 0 {
		t.Fatalf("syntax diagnostic offered an inapplicable rule action: %+v", actions)
	}
}

func TestRunConfiguredLintForContent_ZeroRuleTargetsReportSyntaxErrors(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "malformed.js")
	const malformed = "debugger; const value = ;\n"
	if err := os.WriteFile(filePath, []byte(malformed), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newTestServer()
	s.cwd = dir
	s.fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	uri := documentURIFromPath(filePath)
	s.documents[uri] = malformed

	tests := []struct {
		name   string
		config config.RslintConfig
	}{
		{
			name: "no matching config entry",
			config: config.RslintConfig{{
				Files: []string{"**/*.ts"},
				Rules: config.Rules{"no-debugger": "error"},
			}},
		},
		{name: "valid empty config", config: config.RslintConfig{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := s.runConfiguredLintForContent(
				uri,
				context.Background(),
				malformed,
				test.config,
				dir,
				false,
				nil,
			)
			if err != nil {
				t.Fatalf("run lint: %v", err)
			}
			if !result.HasSyntaxErrors || len(result.Diagnostics) == 0 || !strings.HasPrefix(result.Diagnostics[0].RuleName, "TypeScript(TS") {
				t.Fatalf("zero-rule target did not report its syntax error: %+v", result)
			}
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.RuleName == "no-debugger" {
					t.Fatalf("a rule ran outside its files selector: %+v", result.Diagnostics)
				}
			}
		})
	}
}

type realpathAliasLSPTestFS struct {
	vfs.FS
	aliasRoot string
	realRoot  string
}

func (fs *realpathAliasLSPTestFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	aliasRoot := tspath.NormalizePath(fs.aliasRoot)
	if filePath == aliasRoot {
		return tspath.NormalizePath(fs.realRoot)
	}
	if strings.HasPrefix(filePath, aliasRoot+"/") {
		return tspath.NormalizePath(fs.realRoot) + strings.TrimPrefix(filePath, aliasRoot)
	}
	return fs.FS.Realpath(filePath)
}

func TestRunConfiguredLintForContent_OverlaysLexicalAndRealpath(t *testing.T) {
	root := t.TempDir()
	realRoot := filepath.Join(root, "real-workspace")
	aliasRoot := filepath.Join(root, "alias-workspace")
	realFile := filepath.Join(realRoot, "src", "index.ts")
	if err := os.MkdirAll(filepath.Dir(realFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(realFile, []byte("const diskValue = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tsConfigPath := filepath.Join(realRoot, "tsconfig.json")
	if err := os.WriteFile(tsConfigPath, []byte(`{"compilerOptions":{"noLib":true},"files":["src/index.ts"]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	config.RegisterAllRules()
	s := newTestServer()
	s.cwd = aliasRoot
	s.fs = &realpathAliasLSPTestFS{
		FS:        bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		aliasRoot: aliasRoot,
		realRoot:  realRoot,
	}
	aliasFile := filepath.Join(aliasRoot, "src", "index.ts")
	uri := documentURIFromPath(aliasFile)
	realURI := documentURIFromPath(realFile)
	const openContent = "const editorValue = 2;\n"
	s.documents[realURI] = "const competingAliasValue = 3;\n"
	s.documents[uri] = openContent

	editorOverlay := s.currentEditorOverlayFS(uri)
	for _, filePath := range []string{aliasFile, realFile} {
		if got, ok := editorOverlay.ReadFile(tspath.NormalizePath(filePath)); !ok || got != openContent {
			t.Fatalf("editor overlay read %q = %q, %v; want open content", filePath, got, ok)
		}
	}

	const fixedContent = "debugger;\n"
	result, err := s.runConfiguredLintForContent(
		uri,
		context.Background(),
		fixedContent,
		config.RslintConfig{{
			Files: []string{"src/**/*.ts"},
			Rules: config.Rules{"no-debugger": "error"},
		}},
		aliasRoot,
		false,
		[]string{tsConfigPath},
	)
	if err != nil {
		t.Fatalf("runConfiguredLintForContent failed: %v", err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].RuleName != "no-debugger" {
		t.Fatalf("canonical program read stale disk content: %+v", result.Diagnostics)
	}
}

func TestRunConfiguredLintForContent_SymlinkedConfigRootKeepsRulePathSpace(t *testing.T) {
	parent := t.TempDir()
	realRoot := filepath.Join(parent, "real-root")
	aliasRoot := filepath.Join(parent, "alias-root")
	realFile := filepath.Join(realRoot, "src", "index.ts")
	if err := os.MkdirAll(filepath.Dir(realFile), 0o755); err != nil {
		t.Fatal(err)
	}
	const source = "debugger;\n"
	if err := os.WriteFile(realFile, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realRoot, "tsconfig.json"), []byte(`{"include":["src"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realRoot, aliasRoot); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	config.RegisterAllRules()
	s := newTestServer()
	s.cwd = aliasRoot
	s.fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	aliasFile := filepath.Join(aliasRoot, "src", "index.ts")
	uri := documentURIFromPath(aliasFile)
	s.documents[uri] = source
	result, err := s.runConfiguredLintForContent(
		uri,
		context.Background(),
		source,
		config.RslintConfig{{
			Files: []string{"src/**/*.ts"},
			Rules: config.Rules{"no-debugger": "error"},
		}},
		aliasRoot,
		false,
		[]string{filepath.Join(aliasRoot, "tsconfig.json")},
	)
	if err != nil {
		t.Fatalf("run lint through symlinked root: %v", err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].RuleName != "no-debugger" {
		t.Fatalf("symlinked config path lost scoped rules: %+v", result.Diagnostics)
	}
}

func TestConfigCatalog_SymlinkedOwnerMatchesPhysicalFile(t *testing.T) {
	parent := t.TempDir()
	realRoot := filepath.Join(parent, "real-root")
	aliasRoot := filepath.Join(parent, "alias-root")
	realFile := filepath.Join(realRoot, "src", "index.ts")
	if err := os.MkdirAll(filepath.Dir(realFile), 0o755); err != nil {
		t.Fatal(err)
	}
	const source = "debugger;\n"
	if err := os.WriteFile(realFile, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realRoot, aliasRoot); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	config.RegisterAllRules()
	s := newTestServer()
	s.cwd = realRoot
	s.fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	fileURI := documentURIFromPath(realFile)
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		tspath.NormalizePath(aliasRoot): {{
			Files: []string{"src/**/*.ts"},
			Rules: config.Rules{"no-debugger": "error"},
		}},
	})

	cfg, configCwd, isJSConfig := s.getConfigForURI(fileURI)
	if !isJSConfig || tspath.NormalizePath(configCwd) != tspath.NormalizePath(aliasRoot) {
		t.Fatalf("physical file resolved to config cwd %q, JS=%v", configCwd, isJSConfig)
	}
	result, err := s.runConfiguredLintForContent(
		fileURI,
		context.Background(),
		source,
		cfg,
		configCwd,
		isJSConfig,
		nil,
	)
	if err != nil {
		t.Fatalf("runConfiguredLintForContent failed: %v", err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].RuleName != "no-debugger" {
		t.Fatalf("physical file lost aliased files selector: %+v", result.Diagnostics)
	}

	if err := os.WriteFile(filepath.Join(realRoot, ".gitignore"), []byte("src/index.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	effective, configCwd, isJSConfig := s.getLintConfigForURI(fileURI)
	result, err = s.runConfiguredLintForContent(
		fileURI,
		context.Background(),
		source,
		effective,
		configCwd,
		isJSConfig,
		nil,
	)
	if err != nil {
		t.Fatalf("runConfiguredLintForContent with .gitignore failed: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("physical file did not inherit aliased config .gitignore: %+v", result.Diagnostics)
	}
}

func TestConfigCatalog_PrefersLexicalOwnerOverPhysicalConfig(t *testing.T) {
	root := t.TempDir()
	physicalDir := filepath.Join(root, "physical")
	physicalSubdir := filepath.Join(physicalDir, "sub")
	if err := os.MkdirAll(physicalSubdir, 0o755); err != nil {
		t.Fatal(err)
	}
	const source = "console.log('value');\ndebugger;\n"
	if err := os.WriteFile(filepath.Join(physicalSubdir, "index.ts"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	aliasDir := filepath.Join(root, "link")
	if err := os.Symlink(physicalSubdir, aliasDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	config.RegisterAllRules()
	s := newTestServer()
	s.cwd = root
	s.fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		tspath.NormalizePath(root): {{
			Rules: config.Rules{"no-console": "error"},
		}},
		tspath.NormalizePath(physicalDir): {{
			Rules: config.Rules{"no-debugger": "error"},
		}},
	})

	fileURI := documentURIFromPath(filepath.Join(aliasDir, "index.ts"))
	cfg, configCwd, isJSConfig := s.getConfigForURI(fileURI)
	if !isJSConfig || tspath.NormalizePath(configCwd) != tspath.NormalizePath(root) {
		t.Fatalf("lexical file resolved to config cwd %q, JS=%v", configCwd, isJSConfig)
	}
	result, err := s.runConfiguredLintForContent(
		fileURI,
		context.Background(),
		source,
		cfg,
		configCwd,
		isJSConfig,
		nil,
	)
	if err != nil {
		t.Fatalf("runConfiguredLintForContent failed: %v", err)
	}
	ruleNames := make(map[string]struct{}, len(result.Diagnostics))
	for _, diagnostic := range result.Diagnostics {
		ruleNames[diagnostic.RuleName] = struct{}{}
	}
	if _, ok := ruleNames["no-console"]; !ok {
		t.Fatalf("lexical owner rule missing: %+v", result.Diagnostics)
	}
	if _, ok := ruleNames["no-debugger"]; ok {
		t.Fatalf("physical config replaced lexical owner: %+v", result.Diagnostics)
	}
}

func TestComputeFixAllContent_DefaultExcludedFileIsUnchanged(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, ".git", "hooks", "check.ts")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatal(err)
	}
	const source = "var value = 1;\n"
	if err := os.WriteFile(filePath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	config.RegisterAllRules()
	s := newTestServer()
	s.cwd = root
	s.fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	uri := documentURIFromPath(filePath)
	s.documents[uri] = source
	got := s.computeFixAllContent(
		context.Background(),
		uri,
		source,
		config.RslintConfig{{Rules: config.Rules{"no-var": "error"}}},
		root,
		false,
		nil,
	)
	if got != source {
		t.Fatalf("fixAll modified a default-excluded file: %q", got)
	}
}

func TestComputeFixAllContent_NoTsconfigKeepsNativeFixes(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "template-nested")
	filePath := filepath.Join(configDir, "orphan.ts")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const source = "export const orphan = (() => { var output = 1; return output; })();\n"
	if err := os.WriteFile(filePath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	config.RegisterAllRules()
	s := newTestServer()
	s.cwd = root
	s.fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	uri := documentURIFromPath(filePath)
	s.documents[uri] = source
	cfg := config.RslintConfig{{
		Files: []string{"**/*.ts"},
		Rules: config.Rules{"no-var": "error"},
	}}
	result, err := s.runConfiguredLintForContent(uri, context.Background(), source, cfg, configDir, true, nil)
	if err != nil {
		t.Fatalf("lint without tsconfig: %v", err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].RuleName != "no-var" {
		t.Fatalf("lint without tsconfig lost native diagnostics: %+v", result.Diagnostics)
	}
	got := s.computeFixAllContent(
		context.Background(),
		uri,
		source,
		cfg,
		configDir,
		true,
		nil,
	)
	const want = "export const orphan = (() => { let output = 1; return output; })();\n"
	if got != want {
		t.Fatalf("fixAll without tsconfig = %q, want %q", got, want)
	}
}

type caseInsensitiveLSPTestFS struct {
	mockFS
}

func (f *caseInsensitiveLSPTestFS) UseCaseSensitiveFileNames() bool { return false }
func (f *caseInsensitiveLSPTestFS) Realpath(filePath string) string {
	return strings.ToLower(tspath.NormalizePath(filePath))
}

func TestLSPActiveRulesForFile_NoTsconfigFiltersTypeAwareNativeRules(t *testing.T) {
	config.RegisterAllRules()
	cfg := config.RslintConfig{{
		Rules: config.Rules{
			"no-debugger": "error",
			"@typescript-eslint/no-unsafe-member-access": "error",
		},
		Plugins: []string{"@typescript-eslint"},
	}}

	withoutTypeInfo := lspActiveRulesForFile(cfg, "/project/index.ts", "/project", true, false)
	if len(withoutTypeInfo) != 1 || withoutTypeInfo[0].Name != "no-debugger" {
		t.Fatalf("expected only non-type-aware native rule without tsconfig, got %+v", withoutTypeInfo)
	}

	withTypeInfo := lspActiveRulesForFile(cfg, "/project/index.ts", "/project", true, true)
	if len(withTypeInfo) != 2 {
		t.Fatalf("expected both native rules with type info, got %+v", withTypeInfo)
	}
	foundTypeAware := false
	for _, configuredRule := range withTypeInfo {
		if configuredRule.Name == "@typescript-eslint/no-unsafe-member-access" && configuredRule.RequiresTypeInfo {
			foundTypeAware = true
		}
	}
	if !foundTypeAware {
		t.Fatalf("expected configured type-aware rule, got %+v", withTypeInfo)
	}
}

// runLintWithSession must early-return for files matching the config's
// `ignores` patterns, WITHOUT touching the session. This test proves the
// guard semantically (not just by coincidence of a no-op session):
//
//  1. Positive: call with session=nil AND an ignored path. The call must
//     return empty diagnostics with no error. Passing a nil session is the
//     key trick — if the guard is removed, the very next line dereferences
//     session and panics, making the test fail loudly rather than silently.
//  2. Control: call with session=nil AND a non-ignored path. The call MUST
//     panic (runtime nil-pointer dereference). This proves the only thing
//     keeping the positive case alive is the ignore early-return, not some
//     accidental nil-session tolerance downstream.
func TestRunLintWithSession_IgnoredFileShortCircuits(t *testing.T) {
	ctx := context.Background()
	cwd := "/project"
	cfg := config.RslintConfig{
		// Global ignores entry: hides everything under lib/.
		{Ignores: []string{"lib/**"}},
		{Rules: config.Rules{"no-debugger": "error"}},
	}

	ignoredURI := lsproto.DocumentUri("file:///project/lib/util.ts")
	normalURI := lsproto.DocumentUri("file:///project/src/main.ts")

	t.Run("ignored file returns empty without touching session", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("runLintWithSession panicked on ignored file (early-return missing?): %v", r)
			}
		}()

		diags, err := runLintWithSession(ignoredURI, nil, ctx, cfg, cwd, false, nil, nil)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if diags == nil {
			t.Fatal("expected non-nil empty slice (LSP protocol expects [], not null)")
		}
		if len(diags) != 0 {
			t.Errorf("expected 0 diagnostics for ignored file, got %d: %+v", len(diags), diags)
		}
	})

	t.Run("non-ignored file falls through to session (nil-session → panic)", func(t *testing.T) {
		// This control test asserts the inverse: without a matching ignore,
		// the function proceeds to `session.GetLanguageService(...)` which
		// must nil-dereference. If this test stops panicking, it means some
		// other short-circuit has crept in and the positive test above may
		// be passing for the wrong reason.
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic when non-ignored file is given a nil session, got none — the ignore short-circuit may be matching too broadly")
			}
		}()
		_, _ = runLintWithSession(normalURI, nil, ctx, cfg, cwd, false, nil, nil)
	})
}

func TestRunLintWithSession_DefaultExcludedDirectoryShortCircuits(t *testing.T) {
	cfg := config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}}
	for _, uri := range []lsproto.DocumentUri{
		"file:///project/node_modules/pkg/index.ts",
		"file:///project/.git/hooks/pre-commit.ts",
	} {
		t.Run(string(uri), func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("default-excluded file reached the nil session: %v", recovered)
				}
			}()

			diagnostics, err := runLintWithSession(uri, nil, context.Background(), cfg, "/project", false, nil, nil)
			if err != nil {
				t.Fatalf("runLintWithSession returned an error: %v", err)
			}
			if diagnostics == nil || len(diagnostics) != 0 {
				t.Fatalf("default-excluded file diagnostics = %+v, want a non-nil empty slice", diagnostics)
			}
		})
	}
}

func TestLSPExplicitTargetIgnoreConformance(t *testing.T) {
	tests := []struct {
		name         string
		configDir    string
		relative     string
		config       config.RslintConfig
		gitignores   map[string]string
		symlinkDir   bool
		targetIgnore string
		wantLinted   bool
	}{
		{
			name:     "global config ignore suppresses explicit target",
			relative: "global.ts",
			config: config.RslintConfig{
				{Ignores: []string{"global.ts"}},
				{Files: []string{"**/*.ts"}, Rules: config.Rules{"no-debugger": "error"}},
			},
		},
		{
			name:     "entry ignore keeps target but removes rules",
			relative: "entry.ts",
			config: config.RslintConfig{{
				Files:   []string{"**/*.ts"},
				Ignores: []string{"entry.ts"},
				Rules:   config.Rules{"no-debugger": "error"},
			}},
			wantLinted: true,
		},
		{
			name:       "root gitignore suppresses explicit target",
			relative:   "ignored.ts",
			config:     config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}},
			gitignores: map[string]string{".gitignore": "ignored.ts\n"},
		},
		{
			name:     "config negation restores gitignored explicit target",
			relative: "dist/important.ts",
			config: config.RslintConfig{
				{Ignores: []string{"!dist/important.ts"}},
				{Rules: config.Rules{"no-debugger": "error"}},
			},
			gitignores: map[string]string{".gitignore": "dist/\n"},
			wantLinted: true,
		},
		{
			name:     "nested negation restores explicit target",
			relative: "nested/keep.ts",
			config:   config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}},
			gitignores: map[string]string{
				".gitignore":        "nested/*.ts\n",
				"nested/.gitignore": "!keep.ts\n",
			},
			wantLinted: true,
		},
		{
			name:     "ignored parent blocks nested source",
			relative: "blocked/keep.ts",
			config:   config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}},
			gitignores: map[string]string{
				".gitignore":         "blocked/\n",
				"blocked/.gitignore": "!keep.ts\n",
			},
		},
		{
			name:      "parent ignore does not affect nested config",
			configDir: "packages/app",
			relative:  "ignored/keep.ts",
			config:    config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}},
			gitignores: map[string]string{
				".gitignore":                      "/packages/app/ignored/\n",
				"packages/app/ignored/.gitignore": "!keep.ts\n",
			},
			wantLinted: true,
		},
		{
			name:     "pruned nested source does not override root negation",
			relative: "dist/types/private.ts",
			config:   config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}},
			gitignores: map[string]string{
				".gitignore":            "dist/\n!dist/types/\n",
				"dist/types/.gitignore": "private.ts\n",
			},
			wantLinted: true,
		},
		{
			name:       "directory symlink remains lintable without ignore",
			relative:   "link/source.ts",
			config:     config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}},
			symlinkDir: true,
			wantLinted: true,
		},
		{
			name:       "directory symlink obeys lexical root gitignore",
			relative:   "link/source.ts",
			config:     config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}},
			gitignores: map[string]string{".gitignore": "link/source.ts\n"},
			symlinkDir: true,
		},
		{
			name:         "directory symlink skips target gitignore source",
			relative:     "link/source.ts",
			config:       config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}},
			symlinkDir:   true,
			targetIgnore: "source.ts\n",
			wantLinted:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspace := t.TempDir()
			configDir := workspace
			if test.configDir != "" {
				configDir = filepath.Join(workspace, test.configDir)
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if test.symlinkDir {
				targetDir := t.TempDir()
				if test.targetIgnore != "" {
					if err := os.WriteFile(filepath.Join(targetDir, ".gitignore"), []byte(test.targetIgnore), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(targetDir, filepath.Join(configDir, "link")); err != nil {
					t.Skipf("directory symlink unavailable: %v", err)
				}
			}
			for relative, content := range test.gitignores {
				filePath := filepath.Join(workspace, relative)
				if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			const malformed = "debugger; const value = ;\n"
			target := filepath.Join(configDir, test.relative)
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, []byte(malformed), 0o644); err != nil {
				t.Fatal(err)
			}

			s := newTestServer()
			s.cwd = configDir
			s.fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			s.jsonConfig = test.config
			uri := documentURIFromPath(target)
			effective, configCwd, isJSConfig := s.getLintConfigForURI(uri)
			result, err := s.runConfiguredLintForContent(
				uri,
				context.Background(),
				malformed,
				effective,
				configCwd,
				isJSConfig,
				nil,
			)
			if err != nil {
				t.Fatalf("run lint: %v", err)
			}
			gotLinted := result.HasSyntaxErrors && len(result.Diagnostics) > 0 &&
				strings.HasPrefix(result.Diagnostics[0].RuleName, "TypeScript(TS")
			if gotLinted != test.wantLinted {
				t.Fatalf("linted=%v, want %v: diagnostics=%+v", gotLinted, test.wantLinted, result.Diagnostics)
			}
		})
	}
}
