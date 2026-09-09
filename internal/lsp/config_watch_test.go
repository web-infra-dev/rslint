package lsp

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
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

func TestHandleDidChangeWatchedFilesInvalidatesTypeInfoForTSConfigVariants(t *testing.T) {
	for _, uri := range []lsproto.DocumentUri{
		"file:///project/tsconfig.json",
		"file:///project/tsconfig.build.json",
		"file:///project/jsconfig.json",
	} {
		t.Run(string(uri), func(t *testing.T) {
			s := newTestServer()
			s.fs = &mockFS{files: map[string]bool{}}
			s.cwd = "/project"
			s.lintSessionRoots = newLintSessionProjectRootCache()
			s.lintSessionRoots.entries["/project/old-tsconfig.json"] = lintSessionProjectRootEntry{}

			if err := s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
				Changes: []*lsproto.FileEvent{{Uri: uri, Type: lsproto.FileChangeTypeChanged}},
			}); err != nil {
				t.Fatalf("tsconfig event: %v", err)
			}
			if len(s.lintSessionRoots.entries) != 0 {
				t.Fatalf("stale project metadata survived invalidation: %+v", s.lintSessionRoots.entries)
			}
			select {
			case <-s.refreshCh:
			default:
				t.Fatal("tsconfig change did not request new document snapshots")
			}
		})
	}
}

func TestHandleDidChangeWatchedFilesInvalidatesOrdinaryProjectPathsWithoutOpenDocuments(t *testing.T) {
	s := newTestServer()
	fsys := &mockFS{files: map[string]bool{"/project/custom.json": true}}
	s.fs = fsys
	s.cwd = "/project"
	installJSConfigsForTest(s, map[string]config.RslintConfig{s.cwd: {{
		LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{Project: config.ProjectPaths{"./custom.json"}}},
	}}})
	uri := lsproto.DocumentUri("file:///project/file.ts")
	initial := s.documentLintSnapshot(uri)
	if initial.projectPolicyError != nil || len(initial.typeScriptConfigPaths) != 1 {
		t.Fatalf("initial paths=%v error=%v", initial.typeScriptConfigPaths, initial.projectPolicyError)
	}
	delete(fsys.files, "/project/custom.json")
	if cached := s.documentLintSnapshot(uri); cached.projectPolicyError != nil || len(cached.typeScriptConfigPaths) != 1 {
		t.Fatalf("ordinary owner paths were expanded again before invalidation: %+v", cached)
	}
	if err := s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{{Uri: "file:///project/custom.json", Type: lsproto.FileChangeTypeDeleted}},
	}); err != nil {
		t.Fatal(err)
	}
	if missing := s.documentLintSnapshot(uri); missing.projectPolicyError == nil {
		t.Fatal("closed-document config deletion retained cached project paths")
	}
	if _, cached := s.tsConfigPathsByConfig[s.cwd]; cached {
		t.Fatal("failed path resolution was cached")
	}
	fsys.files["/project/custom.json"] = true
	if restored := s.documentLintSnapshot(uri); restored.projectPolicyError != nil || len(restored.typeScriptConfigPaths) != 1 {
		t.Fatalf("recreated config did not recover: %+v", restored)
	}
}

func TestHandleDidChangeWatchedFilesReevaluatesCustomProject(t *testing.T) {
	for _, pattern := range []string{"./custom.json", "./custom*.json"} {
		t.Run(pattern, func(t *testing.T) {
			fixture := newLintProgramStoreFixture(t, "export const value = 1;\n")
			s := fixture.server
			customPath := tspath.ResolvePath(s.cwd, "custom.json")
			if err := os.Rename(fixture.configPath, customPath); err != nil {
				t.Fatal(err)
			}
			s.backgroundCtx = context.Background()
			if err := s.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
				t.Fatal(err)
			}
			defer s.session.Close()
			s.lintPrograms = fixture.store
			s.session.DidOpenFile(context.Background(), fixture.sourceURI, 1, s.documents[fixture.sourceURI], "typescript")
			installJSConfigsForTest(s, map[string]config.RslintConfig{s.cwd: {{
				LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{Project: config.ProjectPaths{pattern}}},
			}}})
			select {
			case <-s.refreshCh:
			default:
			}
			for _, step := range []struct {
				name   string
				kind   lsproto.FileChangeType
				strict bool
			}{
				{name: "initial"},
				{name: "changed", kind: lsproto.FileChangeTypeChanged, strict: true},
				{name: "deleted", kind: lsproto.FileChangeTypeDeleted},
				{name: "recreated", kind: lsproto.FileChangeTypeCreated},
			} {
				if step.kind == lsproto.FileChangeTypeDeleted {
					if err := os.Remove(customPath); err != nil {
						t.Fatal(err)
					}
				} else if step.kind != 0 {
					strict := "false"
					if step.strict {
						strict = "true"
					}
					if err := os.WriteFile(customPath, []byte(`{"compilerOptions":{"noLib":true,"strict":`+strict+`},"files":["src/index.ts"]}`), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				if step.kind != 0 {
					if err := s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
						Changes: []*lsproto.FileEvent{{Uri: documentURIFromPath(customPath), Type: step.kind}},
					}); err != nil {
						t.Fatal(err)
					}
					select {
					case <-s.refreshCh:
					default:
						t.Fatalf("%s custom project did not schedule diagnostics", step.name)
					}
				}
				snapshot := s.documentLintSnapshot(fixture.sourceURI)
				if step.kind == lsproto.FileChangeTypeDeleted {
					if snapshot.projectPolicyError == nil {
						t.Fatal("deleted explicit project retained resolved paths")
					}
					continue
				}
				if snapshot.projectPolicyError != nil || len(snapshot.typeScriptConfigPaths) != 1 || snapshot.typeScriptConfigPaths[0] != customPath {
					t.Fatalf("%s paths=%v error=%v", step.name, snapshot.typeScriptConfigPaths, snapshot.projectPolicyError)
				}
				for _, speculative := range []bool{false, true} {
					var generation linter.Generation
					var release linter.ReleaseFunc
					var err error
					if speculative {
						generation, release, err = acquireSpeculativeGeneration(context.Background(), s.documents[fixture.sourceURI], snapshot,
							s.freezeSpeculativeLintEnvironment(fixture.sourceURI, snapshot.target))
					} else {
						provider := &documentGenerationProvider{server: s, uri: fixture.sourceURI, snapshot: snapshot}
						generation, release, err = provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
					}
					if err != nil || len(generation.Native.Programs) != 1 {
						t.Fatalf("%s speculative=%v generation=%+v error=%v", step.name, speculative, generation, err)
					}
					options := generation.Native.Programs[0].Options()
					if lintProgramLexicalPathID(options.ConfigFilePath, s.fs) != lintProgramLexicalPathID(customPath, s.fs) || options.Strict.IsTrue() != step.strict {
						t.Fatalf("%s speculative=%v selected stale project: %+v", step.name, speculative, options)
					}
					if release != nil {
						release()
					}
				}
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
