package lsp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
)

// ======== isRslintConfigURI tests ========

func TestIsRslintConfigURI(t *testing.T) {
	tests := []struct {
		uri      string
		expected bool
	}{
		{"file:///project/rslint.json", true},
		{"file:///project/rslint.jsonc", true},
		{"file:///project/sub/rslint.json", true},
		{"file:///project/sub/rslint.jsonc", true},
		{"file:///project/tsconfig.json", false},
		{"file:///project/package.json", false},
		{"file:///project/src/some.ts", false},
		{"file:///project/not-rslint.json", false},
		{"file:///project/rslint.json.bak", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			if got := isRslintConfigURI(tt.uri); got != tt.expected {
				t.Errorf("isRslintConfigURI(%q) = %v, want %v", tt.uri, got, tt.expected)
			}
		})
	}
}

// ======== isTsConfigURI tests ========

func TestIsTsConfigURI(t *testing.T) {
	tests := []struct {
		uri      string
		expected bool
	}{
		{"file:///project/tsconfig.json", true},
		{"file:///project/jsconfig.json", true},
		{"file:///project/tsconfig.build.json", true},
		{"file:///project/tsconfig.app.json", true},
		{"file:///project/sub/tsconfig.json", true},
		{"file:///project/package.json", false},
		{"file:///project/rslint.json", false},
		{"file:///project/src/some.ts", false},
		{"file:///project/other-config.json", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			if got := isTsConfigURI(tt.uri); got != tt.expected {
				t.Errorf("isTsConfigURI(%q) = %v, want %v", tt.uri, got, tt.expected)
			}
		})
	}
}

// ======== reloadConfigAndRelint guard tests ========

func TestReloadConfigAndRelint_SkipsWhenJSConfigsActive(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{"/project/rslint.json": true}}
	s.cwd = "/project"
	// Simulate JS configs being active
	s.jsConfigs["file:///project"] = config.RslintConfig{{}}

	s.reloadConfigAndRelint()

	// Should NOT load JSON config because JS configs take priority
	if s.rslintConfigPath != "" {
		t.Errorf("expected empty config path (skipped), got %q", s.rslintConfigPath)
	}
}

func TestReloadConfigAndRelint_ClearsTypeInfoFiles(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"
	s.rslintConfigPath = "/project/rslint.json"
	// Set stale tsConfigPaths
	s.tsConfigPaths = []string{"/project/old-tsconfig.json"}

	s.reloadConfigAndRelint()

	// Config deleted → tsConfigPaths should be cleared
	if s.tsConfigPaths != nil {
		t.Errorf("expected tsConfigPaths cleared, got %v", s.tsConfigPaths)
	}
}

// ======== handleDidChangeWatchedFiles tests ========

func TestHandleDidChangeWatchedFiles_NilParams(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	err := s.handleDidChangeWatchedFiles(ctx, nil)
	if err != nil {
		t.Fatalf("should not error on nil params: %v", err)
	}
}

func TestHandleDidChangeWatchedFiles_EmptyChanges(t *testing.T) {
	s := newTestServer()
	s.rslintConfigPath = "/project/rslint.json"
	ctx := context.Background()

	err := s.handleDidChangeWatchedFiles(ctx, &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Nothing should change
	if s.rslintConfigPath != "/project/rslint.json" {
		t.Errorf("rslintConfigPath changed unexpectedly to %q", s.rslintConfigPath)
	}
}

func TestHandleDidChangeWatchedFiles_NonConfigFile(t *testing.T) {
	s := newTestServer()
	s.rslintConfigPath = "/project/rslint.json"
	ctx := context.Background()

	// Changes to non-config files should not trigger rslint config reload
	err := s.handleDidChangeWatchedFiles(ctx, &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{
			{Uri: "file:///project/src/index.ts", Type: lsproto.FileChangeTypeChanged},
			{Uri: "file:///project/src/utils.ts", Type: lsproto.FileChangeTypeCreated},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// rslintConfigPath should remain unchanged
	if s.rslintConfigPath != "/project/rslint.json" {
		t.Errorf("rslintConfigPath changed unexpectedly to %q", s.rslintConfigPath)
	}
}

func TestHandleDidChangeWatchedFiles_DetectsConfigFiles(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected bool
	}{
		{"rslint.json", "file:///project/rslint.json", true},
		{"rslint.jsonc", "file:///project/rslint.jsonc", true},
		{"nested rslint.json", "file:///project/sub/rslint.json", true},
		{"tsconfig.json", "file:///project/tsconfig.json", false},
		{"package.json", "file:///project/package.json", false},
		{"some.ts", "file:///project/src/some.ts", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer()
			// Use mockFS so findRslintConfig doesn't panic when config change is detected
			s.fs = &mockFS{files: map[string]bool{}}
			s.cwd = "/project"
			ctx := context.Background()

			err := s.handleDidChangeWatchedFiles(ctx, &lsproto.DidChangeWatchedFilesParams{
				Changes: []*lsproto.FileEvent{
					{Uri: lsproto.DocumentUri(tt.uri), Type: lsproto.FileChangeTypeChanged},
				},
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expected {
				// Config change detected → findRslintConfig ran → config path cleared
				// (because mockFS has no files)
				if s.rslintConfigPath != "" {
					t.Errorf("expected config path to be cleared (config file not found by mockFS), got %q", s.rslintConfigPath)
				}
			}
		})
	}
}

func TestHandleDidChangeWatchedFiles_TsConfigChangeRebuildsTypeInfoFiles(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"
	// Set stale tsConfigPaths
	s.tsConfigPaths = []string{"/project/old-tsconfig.json"}

	ctx := context.Background()
	err := s.handleDidChangeWatchedFiles(ctx, &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{
			{Uri: "file:///project/tsconfig.json", Type: lsproto.FileChangeTypeChanged},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// rebuildTsConfigPaths should have run → no config → tsConfigPaths cleared
	if s.tsConfigPaths != nil {
		t.Errorf("expected tsConfigPaths cleared after tsconfig change, got %v", s.tsConfigPaths)
	}
}

func TestHandleDidChangeWatchedFiles_TsConfigVariantDetected(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"
	s.tsConfigPaths = []string{"/project/old-tsconfig.json"}

	ctx := context.Background()
	// tsconfig.build.json should also trigger rebuild
	err := s.handleDidChangeWatchedFiles(ctx, &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{
			{Uri: "file:///project/tsconfig.build.json", Type: lsproto.FileChangeTypeChanged},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.tsConfigPaths != nil {
		t.Errorf("expected tsConfigPaths cleared after tsconfig.build.json change")
	}
}

func TestHandleDidChangeWatchedFiles_ConfigDeleted(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}} // empty — config file doesn't exist
	s.cwd = "/project"
	s.rslintConfigPath = "/project/rslint.json"
	s.jsonConfig = config.RslintConfig{{}} // non-empty config
	ctx := context.Background()

	err := s.handleDidChangeWatchedFiles(ctx, &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{
			{Uri: "file:///project/rslint.json", Type: lsproto.FileChangeTypeDeleted},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Config should be cleared
	if s.rslintConfigPath != "" {
		t.Errorf("rslintConfigPath should be empty after config deletion, got %q", s.rslintConfigPath)
	}
	if len(s.jsonConfig) != 0 {
		t.Errorf("rslintConfig should be empty after config deletion, got %d entries", len(s.jsonConfig))
	}
}

func TestHandleDidChangeWatchedFiles_MixedConfigAndNonConfig(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"
	s.rslintConfigPath = "/project/rslint.json"
	s.jsonConfig = config.RslintConfig{{}}
	ctx := context.Background()

	// A batch containing both non-config and config changes
	err := s.handleDidChangeWatchedFiles(ctx, &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{
			{Uri: "file:///project/tsconfig.json", Type: lsproto.FileChangeTypeChanged},
			{Uri: "file:///project/src/app.ts", Type: lsproto.FileChangeTypeChanged},
			{Uri: "file:///project/rslint.json", Type: lsproto.FileChangeTypeChanged},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// rslint config change should still be detected even among other changes
	if s.rslintConfigPath != "" {
		t.Errorf("expected config path to be cleared (config file not found by mockFS), got %q", s.rslintConfigPath)
	}
}

func TestHandleDidChangeWatchedFiles_SessionNil(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"
	// session is nil — should not panic
	ctx := context.Background()

	// Non-config change with nil session — should be safe
	err := s.handleDidChangeWatchedFiles(ctx, &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{
			{Uri: "file:///project/tsconfig.json", Type: lsproto.FileChangeTypeChanged},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error with nil session: %v", err)
	}

	// Config change with nil session — should still reload rslint config
	s.rslintConfigPath = "/project/rslint.json"
	err = s.handleDidChangeWatchedFiles(ctx, &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{
			{Uri: "file:///project/rslint.json", Type: lsproto.FileChangeTypeChanged},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error with nil session: %v", err)
	}

	// Config path should be cleared (mockFS has no files)
	if s.rslintConfigPath != "" {
		t.Errorf("expected config path cleared, got %q", s.rslintConfigPath)
	}
}

func TestHandleDidChangeWatchedFiles_AllChangeTypes(t *testing.T) {
	changeTypes := []struct {
		name       string
		changeType lsproto.FileChangeType
	}{
		{"created", lsproto.FileChangeTypeCreated},
		{"changed", lsproto.FileChangeTypeChanged},
		{"deleted", lsproto.FileChangeTypeDeleted},
	}

	for _, ct := range changeTypes {
		t.Run(ct.name, func(t *testing.T) {
			s := newTestServer()
			s.fs = &mockFS{files: map[string]bool{}}
			s.cwd = "/project"
			s.rslintConfigPath = "/project/rslint.json"
			s.jsonConfig = config.RslintConfig{{}}
			ctx := context.Background()

			err := s.handleDidChangeWatchedFiles(ctx, &lsproto.DidChangeWatchedFilesParams{
				Changes: []*lsproto.FileEvent{
					{Uri: "file:///project/rslint.json", Type: ct.changeType},
				},
			})
			if err != nil {
				t.Fatalf("unexpected error for change type %s: %v", ct.name, err)
			}

			// All change types should trigger config reload
			if s.rslintConfigPath != "" {
				t.Errorf("expected config path cleared for change type %s, got %q", ct.name, s.rslintConfigPath)
			}
		})
	}
}

// ======== RefreshDiagnostics tests ========

func TestRefreshDiagnostics_SignalsChannel(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	err := s.RefreshDiagnostics(ctx)
	if err != nil {
		t.Fatalf("RefreshDiagnostics failed: %v", err)
	}

	// Should have a signal in the channel
	select {
	case <-s.refreshCh:
		// good
	default:
		t.Fatal("expected a signal in refreshCh")
	}
}

func TestRefreshDiagnostics_NonBlocking(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	// Fill the buffered channel
	s.refreshCh <- struct{}{}

	// Second call should not block (drops the signal since one is already pending)
	err := s.RefreshDiagnostics(ctx)
	if err != nil {
		t.Fatalf("RefreshDiagnostics failed: %v", err)
	}

	// Channel should still have exactly one signal
	select {
	case <-s.refreshCh:
		// good — consumed the one signal
	default:
		t.Fatal("expected a signal in refreshCh")
	}

	// Channel should now be empty
	select {
	case <-s.refreshCh:
		t.Fatal("expected empty refreshCh after consuming one signal")
	default:
		// good
	}
}

func TestRefreshDiagnostics_MultipleCalls(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()

	// Call RefreshDiagnostics many times rapidly
	for range 10 {
		if err := s.RefreshDiagnostics(ctx); err != nil {
			t.Fatalf("RefreshDiagnostics failed: %v", err)
		}
	}

	// Should coalesce into a single signal
	select {
	case <-s.refreshCh:
		// good
	default:
		t.Fatal("expected a signal in refreshCh")
	}

	select {
	case <-s.refreshCh:
		t.Fatal("expected only one signal after multiple calls")
	default:
		// good
	}
}

// ======== reloadConfigAndRelint tests ========

// TestReloadConfig_NoTsConfigsIsAccepted verifies that a JSON config without
// parserOptions.project is accepted by the LSP (no "no TypeScript configurations" error).
// The LSP session discovers tsconfig files on its own via projectService.
func TestReloadConfig_NoTsConfigsIsAccepted(t *testing.T) {
	dir := t.TempDir()
	// JSON config with rules but no parserOptions.project → zero tsconfigs
	configContent := `[{"rules": {"no-console": "error"}}]`
	if err := os.WriteFile(filepath.Join(dir, "rslint.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	s := newTestServer()
	s.fs = osvfs.FS()
	s.cwd = dir
	s.rslintConfigPath = filepath.Join(dir, "rslint.json")

	config.RegisterAllRules()
	err := s.reloadConfig()
	if err != nil {
		t.Fatalf("reloadConfig should accept config without tsconfigs, got error: %v", err)
	}
	if len(s.jsonConfig) != 1 {
		t.Errorf("expected 1 config entry, got %d", len(s.jsonConfig))
	}
}

func TestReloadConfigAndRelint_RelintOpenDocuments(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"

	// Add non-TS documents — pushDiagnostics skips non-TS files,
	// so this tests the iteration without needing a real session.
	s.documents["file:///project/a.css"] = "body {}"
	s.documents["file:///project/b.json"] = "{}"

	s.reloadConfigAndRelint()

	// Config should be cleared (mockFS has no files)
	if s.rslintConfigPath != "" {
		t.Errorf("expected empty config path, got %q", s.rslintConfigPath)
	}
}

func TestReloadConfigAndRelint_ConfigCleared(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"
	s.rslintConfigPath = "/project/rslint.json"
	s.jsonConfig = config.RslintConfig{{}}

	s.reloadConfigAndRelint()

	if s.rslintConfigPath != "" {
		t.Errorf("expected config path cleared, got %q", s.rslintConfigPath)
	}
	if len(s.jsonConfig) != 0 {
		t.Errorf("expected empty config, got %d entries", len(s.jsonConfig))
	}
}

// ======== ptrIsTrue tests ========

func TestPtrIsTrue(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name     string
		input    *bool
		expected bool
	}{
		{"nil", nil, false},
		{"true", &trueVal, true},
		{"false", &falseVal, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ptrIsTrue(tt.input); got != tt.expected {
				t.Errorf("ptrIsTrue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ======== watchEnabled initialization tests ========

func TestHandleInitialized_WatchEnabledWhenClientSupports(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"
	s.backgroundCtx = context.Background()
	dynamicReg := true
	s.initializeParams = &lsproto.InitializeParams{
		Capabilities: &lsproto.ClientCapabilities{
			Workspace: &lsproto.WorkspaceClientCapabilities{
				DidChangeWatchedFiles: &lsproto.DidChangeWatchedFilesClientCapabilities{
					DynamicRegistration: &dynamicReg,
				},
			},
		},
	}

	// handleInitialized will fail (no real FS/project setup) but watchEnabled
	// should be set before the session is created.
	_ = s.handleInitialized(context.Background(), &lsproto.InitializedParams{})

	if !s.watchEnabled {
		t.Error("watchEnabled should be true when client supports dynamic registration")
	}
}

func TestHandleInitialized_WatchDisabledWhenClientDoesNotSupport(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"
	s.backgroundCtx = context.Background()
	dynamicReg := false
	s.initializeParams = &lsproto.InitializeParams{
		Capabilities: &lsproto.ClientCapabilities{
			Workspace: &lsproto.WorkspaceClientCapabilities{
				DidChangeWatchedFiles: &lsproto.DidChangeWatchedFilesClientCapabilities{
					DynamicRegistration: &dynamicReg,
				},
			},
		},
	}

	_ = s.handleInitialized(context.Background(), &lsproto.InitializedParams{})

	if s.watchEnabled {
		t.Error("watchEnabled should be false when client does not support dynamic registration")
	}
}

func TestHandleInitialized_WatchDisabledWhenCapabilitiesNil(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.cwd = "/project"
	s.backgroundCtx = context.Background()
	s.initializeParams = &lsproto.InitializeParams{
		Capabilities: nil,
	}

	_ = s.handleInitialized(context.Background(), &lsproto.InitializedParams{})

	if s.watchEnabled {
		t.Error("watchEnabled should be false when capabilities are nil")
	}
}

// ======== isBlockingMethod tests ========

func TestIsBlockingMethod_ConfigUpdate(t *testing.T) {
	if !isBlockingMethod(lsproto.Method("rslint/configUpdate")) {
		t.Error("rslint/configUpdate must be a blocking method to avoid data races on jsConfigs")
	}
}

func TestIsBlockingMethod_CodeAction(t *testing.T) {
	if !isBlockingMethod(lsproto.MethodTextDocumentCodeAction) {
		t.Error("textDocument/codeAction must be a blocking method to avoid data races on s.diagnostics")
	}
}

// TestDispatchLoop_ConfigUpdateBeforeDidOpen verifies that when a configUpdate
// notification and a didOpen notification are queued sequentially, the dispatch
// loop processes configUpdate first (both are blocking), so didOpen sees the
// updated jsConfigs.
func TestDispatchLoop_ConfigUpdateBeforeDidOpen(t *testing.T) {
	s, _ := newTestServerWithQueue()
	s.requestQueue = make(chan *lsproto.RequestMessage, 10)

	// Build a configUpdate notification with a test config
	configPayload := map[string]any{
		"configs": []map[string]any{
			{
				"configDirectory": "file:///project",
				"entries": []map[string]any{
					{
						"files": []string{"**/*.ts"},
						"rules": map[string]string{"no-console": "error"},
					},
				},
			},
		},
	}
	configReq := &lsproto.RequestMessage{
		Method: lsproto.Method("rslint/configUpdate"),
		Params: configPayload,
	}

	// Build a didOpen notification as a sentinel — once it's processed,
	// we know all prior blocking messages have completed.
	openReq := &lsproto.RequestMessage{
		Method: lsproto.MethodTextDocumentDidOpen,
		Params: &lsproto.DidOpenTextDocumentParams{
			TextDocument: &lsproto.TextDocumentItem{
				Uri:     "file:///project/test.ts",
				Text:    "const x = 1;",
				Version: 1,
			},
		},
	}

	// Queue both messages before starting the loop
	s.requestQueue <- configReq
	s.requestQueue <- openReq

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- s.dispatchLoop(ctx)
	}()

	// Wait for both messages to be processed, then cancel
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done

	// Verify configUpdate was applied before didOpen ran
	if len(s.jsConfigs) == 0 {
		t.Fatal("jsConfigs should have been populated by configUpdate before didOpen")
	}
	cfg, ok := s.jsConfigs["file:///project"]
	if !ok {
		t.Fatal("jsConfigs missing 'file:///project' key")
	}
	if len(cfg) == 0 {
		t.Fatal("config entries should not be empty")
	}

	// Verify didOpen also ran (the sentinel)
	if _, ok := s.documents["file:///project/test.ts"]; !ok {
		t.Fatal("didOpen should have stored the document")
	}
}

// ======== dispatchLoop tests ========

func TestDispatchLoop_DebounceLintsOnlyPending(t *testing.T) {
	s, queue := newTestServerWithQueue()

	// Add both a pending TS URI and a non-pending CSS URI.
	// pushDiagnostics skips non-TS files, so the pending TS URI would trigger
	// lint if session were non-nil. With session nil, pushDiagnostics returns
	// early, but we verify the pending set is cleared.
	s.documents["file:///project/a.ts"] = "const x = 1;"
	s.documents["file:///project/styles.css"] = "body {}"
	s.pendingLintURIs["file:///project/a.ts"] = struct{}{}

	ctx, cancel := context.WithCancel(context.Background())

	// Send debounce signal, then immediately send a request that will cancel
	// the loop — this ensures the debounce case runs first.
	s.debounceCh <- struct{}{}

	done := make(chan error, 1)
	go func() {
		done <- s.dispatchLoop(ctx)
	}()

	// Give the dispatch loop time to process the debounceCh signal
	time.Sleep(50 * time.Millisecond)
	cancel()

	err := <-done
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("dispatchLoop returned unexpected error: %v", err)
	}

	// Pending URIs should be cleared after processing
	if len(s.pendingLintURIs) != 0 {
		t.Errorf("pendingLintURIs should be empty after debounce, got %d", len(s.pendingLintURIs))
	}

	// No diagnostics published (session is nil → pushDiagnostics returns early)
	select {
	case <-queue:
		t.Fatal("unexpected diagnostics published with nil session")
	default:
		// good
	}
}

func TestDispatchLoop_RefreshRelintsDocs(t *testing.T) {
	s, queue := newTestServerWithQueue()

	// Add a non-TS document so pushDiagnostics skips it safely (no session needed)
	s.documents["file:///project/styles.css"] = "body {}"

	ctx, cancel := context.WithCancel(context.Background())

	// Send refresh signal
	s.refreshCh <- struct{}{}

	// Run dispatchLoop in a goroutine; it will process the refresh then block
	done := make(chan error, 1)
	go func() {
		done <- s.dispatchLoop(ctx)
	}()

	// Cancel after the refresh has been processed
	// The loop should have processed the refreshCh case (which is a no-op for
	// CSS files) and then block on the next select iteration.
	cancel()

	err := <-done
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("dispatchLoop returned unexpected error: %v", err)
	}

	// No diagnostics should be published for CSS file
	select {
	case <-queue:
		t.Fatal("unexpected diagnostics published for non-TS file")
	default:
		// good
	}
}
