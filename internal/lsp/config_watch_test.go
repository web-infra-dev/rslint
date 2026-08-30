package lsp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
)

func TestIsTsConfigURI(t *testing.T) {
	tests := []struct {
		uri  string
		want bool
	}{
		{uri: "file:///project/tsconfig.json", want: true},
		{uri: "file:///project/jsconfig.json", want: true},
		{uri: "file:///project/tsconfig.build.json", want: true},
		{uri: "file:///project/tsconfig.app.json", want: true},
		{uri: "file:///project/sub/tsconfig.json", want: true},
		{uri: "file:///project/package.json", want: false},
		{uri: "file:///project/rslint.json", want: false},
		{uri: "file:///project/src/some.ts", want: false},
		{uri: "file:///project/other-config.json", want: false},
		{uri: "", want: false},
	}
	for _, test := range tests {
		t.Run(test.uri, func(t *testing.T) {
			if got := isTsConfigURI(test.uri); got != test.want {
				t.Errorf("isTsConfigURI(%q) = %v, want %v", test.uri, got, test.want)
			}
		})
	}
}

func TestHandleDidChangeWatchedFilesNilParams(t *testing.T) {
	s := newTestServer()
	if err := s.handleDidChangeWatchedFiles(context.Background(), nil); err != nil {
		t.Fatalf("nil params: %v", err)
	}
}

func TestHandleDidChangeWatchedFilesIgnoresLegacyJSONConfig(t *testing.T) {
	s, outgoing := newTestServerWithQueue()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"
	s.configDiscoveryActive = true

	if err := s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{{
			Uri:  "file:///project/rslint.json",
			Type: lsproto.FileChangeTypeChanged,
		}},
	}); err != nil {
		t.Fatalf("legacy JSON event: %v", err)
	}
	select {
	case message := <-outgoing:
		t.Fatalf("legacy JSON event started config discovery: %+v", message)
	default:
	}
}

func TestHandleDidChangeWatchedFilesRebuildsTypeInfoForTSConfigVariants(t *testing.T) {
	for _, uri := range []lsproto.DocumentUri{
		"file:///project/tsconfig.json",
		"file:///project/tsconfig.build.json",
		"file:///project/jsconfig.json",
	} {
		t.Run(string(uri), func(t *testing.T) {
			s := newTestServer()
			s.fs = &mockFS{files: map[string]bool{}}
			s.cwd = "/project"
			s.tsConfigPathsByConfig = map[string][]string{
				"/project": {"/project/old-tsconfig.json"},
			}

			if err := s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
				Changes: []*lsproto.FileEvent{{Uri: uri, Type: lsproto.FileChangeTypeChanged}},
			}); err != nil {
				t.Fatalf("tsconfig event: %v", err)
			}
			if s.tsConfigPathsByConfig != nil {
				t.Fatalf("stale type-info paths survived rebuild: %+v", s.tsConfigPathsByConfig)
			}
		})
	}
}

func TestRefreshDiagnosticsCoalescesSignals(t *testing.T) {
	s := newTestServer()
	for range 10 {
		if err := s.RefreshDiagnostics(context.Background()); err != nil {
			t.Fatalf("RefreshDiagnostics: %v", err)
		}
	}
	select {
	case <-s.refreshCh:
	default:
		t.Fatal("expected a refresh signal")
	}
	select {
	case <-s.refreshCh:
		t.Fatal("expected refresh signals to coalesce")
	default:
	}
}

func TestPtrIsTrue(t *testing.T) {
	trueValue := true
	falseValue := false
	for _, test := range []struct {
		name  string
		value *bool
		want  bool
	}{
		{name: "nil", want: false},
		{name: "true", value: &trueValue, want: true},
		{name: "false", value: &falseValue, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := ptrIsTrue(test.value); got != test.want {
				t.Errorf("ptrIsTrue() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestHandleInitializedSetsWatchCapability(t *testing.T) {
	for _, test := range []struct {
		name         string
		capabilities *lsproto.ClientCapabilities
		want         bool
	}{
		{
			name: "supported",
			capabilities: &lsproto.ClientCapabilities{
				Workspace: &lsproto.WorkspaceClientCapabilities{
					DidChangeWatchedFiles: &lsproto.DidChangeWatchedFilesClientCapabilities{
						DynamicRegistration: boolPointer(true),
					},
				},
			},
			want: true,
		},
		{
			name: "unsupported",
			capabilities: &lsproto.ClientCapabilities{
				Workspace: &lsproto.WorkspaceClientCapabilities{
					DidChangeWatchedFiles: &lsproto.DidChangeWatchedFilesClientCapabilities{
						DynamicRegistration: boolPointer(false),
					},
				},
			},
		},
		{name: "nil capabilities"},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := newTestServer()
			s.fs = &mockFS{files: map[string]bool{}}
			s.cwd = "/project"
			s.backgroundCtx = context.Background()
			s.initializeParams = &lsproto.InitializeParams{Capabilities: test.capabilities}

			_ = s.handleInitialized(context.Background(), &lsproto.InitializedParams{})
			if s.watchEnabled != test.want {
				t.Errorf("watchEnabled = %v, want %v", s.watchEnabled, test.want)
			}
		})
	}
}

func boolPointer(value bool) *bool {
	return &value
}

func TestIsBlockingMethodCodeAction(t *testing.T) {
	if !isBlockingMethod(lsproto.MethodTextDocumentCodeAction) {
		t.Error("textDocument/codeAction must be blocking")
	}
}

func TestDispatchLoopDebounceLintsOnlyPending(t *testing.T) {
	s, queue := newTestServerWithQueue()
	s.documents["file:///project/a.ts"] = "const x = 1;"
	s.documents["file:///project/styles.css"] = "body {}"
	s.pendingLintURIs["file:///project/a.ts"] = struct{}{}
	s.debounceCh <- struct{}{}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.dispatchLoop(ctx) }()
	time.Sleep(50 * time.Millisecond)
	cancel()
	if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("dispatchLoop: %v", err)
	}
	if len(s.pendingLintURIs) != 0 {
		t.Errorf("pendingLintURIs = %d, want 0", len(s.pendingLintURIs))
	}
	select {
	case <-queue:
		t.Fatal("published diagnostics with nil session")
	default:
	}
}

func TestDispatchLoopRefreshRelintsDocuments(t *testing.T) {
	s, queue := newTestServerWithQueue()
	s.documents["file:///project/styles.css"] = "body {}"
	s.refreshCh <- struct{}{}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.dispatchLoop(ctx) }()
	cancel()
	if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("dispatchLoop: %v", err)
	}
	select {
	case <-queue:
		t.Fatal("published diagnostics for non-TS file")
	default:
	}
}
