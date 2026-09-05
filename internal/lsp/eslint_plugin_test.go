package lsp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsonrpc"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// textOnlySourceFile is a minimal ast.SourceFileLike (Text + ECMALineMap) — the
// same shape internal/linter.textSourceFile gives plugin diagnostics — so
// convertRuleDiagnosticToLSP can compute line/character positions in a unit
// test without spinning up a ts-go program.
type textOnlySourceFile struct{ text string }

func (f textOnlySourceFile) Text() string                { return f.text }
func (f textOnlySourceFile) ECMALineMap() []core.TextPos { return core.ComputeECMALineStarts(f.text) }

// pluginDiag builds a minimal plugin-style RuleDiagnostic backed by a text
// source file (the form applyEslintPluginResults produces).
func pluginDiag(text, ruleName, message string, start, end int) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		RuleName:   ruleName,
		Range:      core.NewTextRange(start, end),
		Message:    rule.RuleMessage{Description: message},
		SourceFile: textOnlySourceFile{text: text},
	}
}

type capturedProgressiveDiagnostics struct {
	baseline []rule.RuleDiagnostic
	run      linter.DeferredPluginRun
}

func (p *capturedProgressiveDiagnostics) PublishBaseline(
	_ context.Context,
	diagnostics []rule.RuleDiagnostic,
) {
	p.baseline = append([]rule.RuleDiagnostic(nil), diagnostics...)
}

func (p *capturedProgressiveDiagnostics) Submit(
	_ context.Context,
	run linter.DeferredPluginRun,
) {
	p.run = run
}

func preparedPluginRunForTest(
	t *testing.T,
	s *Server,
	uri lsproto.DocumentUri,
	content string,
	snapshot documentLintSnapshot,
) linter.DeferredPluginRun {
	t.Helper()
	run := pluginRunForProductionPassTest(t, s, uri, content, snapshot)
	if run == nil {
		t.Fatal("prepared plugin run is nil, want one task")
	}
	return run
}

func pluginRunForProductionPassTest(
	t *testing.T,
	s *Server,
	uri lsproto.DocumentUri,
	content string,
	snapshot documentLintSnapshot,
) linter.DeferredPluginRun {
	t.Helper()
	environment := s.freezeSpeculativeLintEnvironment(uri, snapshot.target)
	if environment.baseFS == nil {
		environment.baseFS = osvfs.FS()
	}
	generation, release, err := acquireSpeculativeGeneration(
		context.Background(), content, snapshot, environment,
	)
	if err != nil {
		t.Fatalf("prepare production plugin pass: %v", err)
	}
	presentation := &capturedProgressiveDiagnostics{}
	_, err = linter.RunPipeline(context.Background(), linter.NewProgressiveLintRequest(
		linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
			return generation, release, nil
		}),
		linter.ArtifactDemand{
			Native: rule.EditDemandNone,
			Plugin: rule.EditDemandAll,
		},
		presentation,
	))
	if err != nil {
		t.Fatalf("run production plugin pass: %v", err)
	}
	return presentation.run
}

func pluginRequestForProductionPassTest(
	t *testing.T,
	s *Server,
	uri lsproto.DocumentUri,
	content string,
	snapshot documentLintSnapshot,
) (linter.EslintPluginLintRequest, bool) {
	t.Helper()
	run := pluginRunForProductionPassTest(t, s, uri, content, snapshot)
	if run == nil {
		return linter.EslintPluginLintRequest{}, false
	}
	var request linter.EslintPluginLintRequest
	outcome, err := run(context.Background(), func(_ context.Context, got linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		request = got
		results := make([]linter.EslintPluginFileResult, len(got.Files))
		for index, file := range got.Files {
			results[index].FilePath = file.Path
		}
		return &linter.EslintPluginLintResult{Results: results}, nil
	})
	if err != nil || outcome.DispatchError != nil {
		t.Fatalf("run prepared plugin run: contract=%v dispatch=%v", err, outcome.DispatchError)
	}
	return request, true
}

func preparedPluginRequestForTest(
	t *testing.T,
	s *Server,
	uri lsproto.DocumentUri,
	content string,
	snapshot documentLintSnapshot,
) linter.EslintPluginLintRequest {
	t.Helper()
	request, ok := pluginRequestForProductionPassTest(t, s, uri, content, snapshot)
	if !ok {
		t.Fatal("prepared plugin request is empty, want one task")
	}
	return request
}

func fixAllGenerationWithNativeFixForTest(
	t *testing.T,
	calls *int,
	prepareSnapshot func(documentLintSnapshot) documentLintSnapshot,
	fix func(string) (core.TextRange, string, bool),
) speculativeGenerationAcquire {
	t.Helper()
	return func(
		ctx context.Context,
		content string,
		snapshot documentLintSnapshot,
		environment speculativeLintEnvironment,
	) (linter.Generation, linter.ReleaseFunc, error) {
		*calls = *calls + 1
		if prepareSnapshot != nil {
			snapshot = prepareSnapshot(snapshot)
		}
		generation, release, err := acquireSpeculativeGeneration(ctx, content, snapshot, environment)
		if err != nil {
			return linter.Generation{}, nil, err
		}
		originalRules := generation.Native.RulesForFile
		generation.Native.RulesForFile = func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
			var configured []rule.ConfiguredRule
			if originalRules != nil {
				configured = append(configured, originalRules(sourceFile)...)
			}
			configured = append(configured, rule.ConfiguredRule{
				Name:     "native/test-fix",
				Severity: rule.SeverityError,
				Run: func(ruleCtx rule.RuleContext) rule.RuleListeners {
					return rule.RuleListeners{
						ast.KindVariableStatement: func(*ast.Node) {
							textRange, replacement, ok := fix(ruleCtx.SourceFile.Text())
							if !ok {
								return
							}
							ruleCtx.ReportRangeWithFixes(
								textRange,
								rule.RuleMessage{Description: "test fix"},
								rule.RuleFix{Text: replacement, Range: textRange},
							)
						},
					}
				},
			})
			return configured
		}
		return generation, release, nil
	}
}

func TestDeriveLSPRuleCatalogSkipsEmptyRuleNames(t *testing.T) {
	catalog, _ := deriveLSPRuleCatalog(
		rule.NewCatalog(),
		[]config.EslintPluginEntry{{Prefix: "community", RuleNames: []string{"", "check"}}},
	)
	if _, ok := catalog.Lookup("community/"); ok {
		t.Fatal("empty plugin rule name became resolvable")
	}
	if _, ok := catalog.Lookup("community/check"); !ok {
		t.Fatal("non-empty plugin rule name was dropped")
	}
}

func TestWriteLSPPluginProtocolNotices(t *testing.T) {
	var output strings.Builder
	writeLSPPluginProtocolNotices(&output, []linter.EslintPluginProtocolNotice{
		{Kind: linter.EslintPluginMissingFileResult, FilePath: "/repo/a.ts"},
		{Kind: linter.EslintPluginUnconfiguredDiagnostic, FilePath: "/repo/b.ts", RuleName: "plugin/extra"},
	})
	want := "rslint: plugin-lint returned no result for \"/repo/a.ts\"\n" +
		"rslint: plugin diagnostic for unconfigured rule \"plugin/extra\" in \"/repo/b.ts\"\n"
	if output.String() != want {
		t.Fatalf("notice output = %q, want %q", output.String(), want)
	}
}

// ======== mergePluginDiagnostics tests ========

func TestMergePluginDiagnostics_MergesAndPublishes(t *testing.T) {
	s, queue := newTestServerWithQueue()
	uri := lsproto.DocumentUri("file:///project/a.ts")
	// Multi-line + multi-byte buffer: the 'é' (U+00E9 — 2 UTF-8 bytes but 1
	// UTF-16 code unit) on line 1 makes byte offsets diverge from UTF-16 char
	// offsets, so a correct byte→UTF-16 conversion is observable (not just
	// "something was published").
	text := "let y;\nconst café = 1;"
	s.documents[uri] = text
	// A prior native diagnostic already stored for the current generation:
	// "let" at bytes [0,3].
	s.diagnostics[uri] = []rule.RuleDiagnostic{pluginDiag(text, "native-rule", "native msg", 0, 3)}
	s.docGeneration[uri] = 7

	// Plugin diagnostic on the `1` literal at bytes [21,22]: line 1 starts at
	// byte 7, and within it the 2-byte 'é' shifts the `1` to UTF-16 char 13.
	s.mergePluginDiagnostics(pluginLintResult{
		uri:        uri,
		generation: 7,
		diags:      []rule.RuleDiagnostic{pluginDiag(text, "plug/no-foo", "plugin msg", 21, 22)},
	})

	// Native + plugin diagnostics must coexist in the stored slice, in order.
	if got := len(s.diagnostics[uri]); got != 2 {
		t.Fatalf("expected 2 merged diagnostics (native+plugin), got %d", got)
	}
	if s.diagnostics[uri][0].RuleName != "native-rule" || s.diagnostics[uri][1].RuleName != "plug/no-foo" {
		t.Errorf("merge order wrong: %q then %q", s.diagnostics[uri][0].RuleName, s.diagnostics[uri][1].RuleName)
	}

	// Decode the queued PublishDiagnostics and assert each diagnostic's
	// byte→UTF-16 line/char conversion to a concrete value.
	var msg *lsproto.Message
	select {
	case msg = <-queue:
	default:
		t.Fatal("expected a PublishDiagnostics notification to be queued")
	}
	params, ok := msg.AsRequest().Params.(*lsproto.PublishDiagnosticsParams)
	if !ok {
		t.Fatalf("queued message params is %T, want *PublishDiagnosticsParams", msg.AsRequest().Params)
	}
	if params.Uri != uri {
		t.Errorf("published for %q, want %q", params.Uri, uri)
	}
	if len(params.Diagnostics) != 2 {
		t.Fatalf("expected 2 published diagnostics, got %d", len(params.Diagnostics))
	}
	// native "let" at bytes [0,3] → line 0, char [0,3].
	if r := params.Diagnostics[0].Range; r.Start.Line != 0 || r.Start.Character != 0 || r.End.Line != 0 || r.End.Character != 3 {
		t.Errorf("native range = L%dC%d..L%dC%d, want L0C0..L0C3", r.Start.Line, r.Start.Character, r.End.Line, r.End.Character)
	}
	// plugin `1` at bytes [21,22] → line 1, char [13,14] (é is 2 bytes / 1 UTF-16 unit).
	if r := params.Diagnostics[1].Range; r.Start.Line != 1 || r.Start.Character != 13 || r.End.Line != 1 || r.End.Character != 14 {
		t.Errorf("plugin range = L%dC%d..L%dC%d, want L1C13..L1C14", r.Start.Line, r.Start.Character, r.End.Line, r.End.Character)
	}
}

func TestPluginDispatchForGeneration_StampsRequest(t *testing.T) {
	s := newTestServer()
	var received string
	s.eslintPluginDispatch = func(_ context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		received = req.Generation
		return &linter.EslintPluginLintResult{}, nil
	}

	_, err := s.pluginDispatchForGeneration("config-12")(context.Background(), linter.EslintPluginLintRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if received != "config-12" {
		t.Fatalf("request generation = %q, want config-12", received)
	}
}

func TestMergePluginDiagnostics_DropsStaleGeneration(t *testing.T) {
	s, queue := newTestServerWithQueue()
	uri := lsproto.DocumentUri("file:///project/a.ts")
	text := "const x = 1;"
	s.documents[uri] = text
	native := []rule.RuleDiagnostic{pluginDiag(text, "native-rule", "native msg", 0, 5)}
	s.diagnostics[uri] = native
	s.docGeneration[uri] = 9 // current generation

	// A plugin result stamped with an older generation (a newer keystroke
	// already re-linted) must be dropped — not merged, not published.
	s.mergePluginDiagnostics(pluginLintResult{
		uri:        uri,
		generation: 8,
		diags:      []rule.RuleDiagnostic{pluginDiag(text, "plug/no-foo", "plugin msg", 6, 7)},
	})

	if got := len(s.diagnostics[uri]); got != 1 {
		t.Errorf("stale result must not be merged; got %d diagnostics, want 1", got)
	}
	select {
	case <-queue:
		t.Fatal("stale plugin result must not publish diagnostics")
	default:
	}
}

func TestMergePluginDiagnostics_DropsClosedDocument(t *testing.T) {
	s, queue := newTestServerWithQueue()
	uri := lsproto.DocumentUri("file:///project/a.ts")
	// Document NOT in s.documents (closed). Generation matches but the buffer
	// is gone, so the result must be discarded.
	s.docGeneration[uri] = 3

	s.mergePluginDiagnostics(pluginLintResult{
		uri:        uri,
		generation: 3,
		diags:      []rule.RuleDiagnostic{pluginDiag("const x = 1;", "plug/no-foo", "plugin msg", 0, 1)},
	})

	if _, ok := s.diagnostics[uri]; ok {
		t.Error("closed document must not gain diagnostics")
	}
	select {
	case <-queue:
		t.Fatal("closed document must not publish diagnostics")
	default:
	}
}

// ======== dispatchLoop plugin-result case ========

// TestDispatchLoop_PluginResultMerged verifies the dispatch loop consumes
// pluginResultCh on the main goroutine and merges the result.
func TestDispatchLoop_PluginResultMerged(t *testing.T) {
	s, queue := newTestServerWithQueue()
	uri := lsproto.DocumentUri("file:///project/a.ts")
	text := "const x = 1;"
	s.documents[uri] = text
	s.diagnostics[uri] = []rule.RuleDiagnostic{}
	s.docGeneration[uri] = 1

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.dispatchLoop(ctx) }()

	s.pluginResultCh <- pluginLintResult{
		uri:        uri,
		generation: 1,
		diags:      []rule.RuleDiagnostic{pluginDiag(text, "plug/no-foo", "plugin msg", 0, 5)},
	}

	// Wait for the loop to consume the channel and publish.
	select {
	case <-queue:
		// good — merged + published on the main loop
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for plugin result to be merged + published")
	}
	cancel()
	if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("dispatchLoop returned unexpected error: %v", err)
	}
}

func newPushDiagnosticsPluginFixture(
	t *testing.T,
	source string,
	entries config.RslintConfig,
) (*Server, lsproto.DocumentUri, chan *lsproto.Message) {
	t.Helper()
	fixture := newLintProgramStoreFixture(t, source)
	s := fixture.server
	s.backgroundCtx = context.Background()
	queue := make(chan *lsproto.Message, 10)
	s.outgoingQueue = queue
	if err := s.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
		t.Fatal(err)
	}
	s.lintPrograms = fixture.store
	s.configSnapshotIncludesGitignore = true
	configDirectory := filepath.Dir(fixture.configPath)
	installJSConfigsForTest(s, map[string]config.RslintConfig{configDirectory: entries})
	s.tsConfigPathsByConfig = map[string][]string{
		configDirectory: {fixture.configPath},
	}
	s.eslintPluginConfigGeneration = "push-generation"
	return s, fixture.sourceURI, queue
}

func TestPushDiagnosticsPublishesBaselineBeforePluginAndAdmitsEnrichment(t *testing.T) {
	s, uri, queue := newPushDiagnosticsPluginFixture(
		t,
		"const value = 1;\n",
		config.RslintConfig{{
			Plugins: []string{"progressive"},
			Rules:   config.Rules{"progressive/check": "error"},
		}},
	)
	baselineAtDispatch := make(chan string, 1)
	s.eslintPluginDispatch = func(_ context.Context, request linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		select {
		case message := <-queue:
			params, ok := message.AsRequest().Params.(*lsproto.PublishDiagnosticsParams)
			if !ok {
				baselineAtDispatch <- fmt.Sprintf("unexpected baseline message %T", message.AsRequest().Params)
			} else {
				baselineAtDispatch <- fmt.Sprintf("%s:%d", params.Uri, len(params.Diagnostics))
			}
		default:
			baselineAtDispatch <- "plugin dispatch started before baseline publication"
		}
		return &linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{
			FilePath: request.Files[0].Path,
			Diagnostics: []linter.EslintPluginDiagnostic{{
				RuleName: "progressive/check",
				Message:  "enriched",
				StartPos: 0,
				EndPos:   5,
			}},
		}}}, nil
	}

	s.pushDiagnostics(uri)
	select {
	case observed := <-baselineAtDispatch:
		if want := fmt.Sprintf("%s:0", uri); observed != want {
			t.Fatalf("publication observed at dispatch = %q, want %q", observed, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("plugin enrichment did not start")
	}

	loopCtx, cancelLoop := context.WithCancel(context.Background())
	loopDone := make(chan error, 1)
	go func() { loopDone <- s.dispatchLoop(loopCtx) }()
	select {
	case message := <-queue:
		params, ok := message.AsRequest().Params.(*lsproto.PublishDiagnosticsParams)
		if !ok || params.Uri != uri || len(params.Diagnostics) != 1 {
			t.Fatalf("enriched publication = %+v", message.AsRequest().Params)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("admitted enrichment was not published")
	}
	cancelLoop()
	if err := <-loopDone; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("dispatch loop returned %v", err)
	}
	if diagnostics := s.diagnostics[uri]; len(diagnostics) != 1 || diagnostics[0].RuleName != "progressive/check" {
		t.Fatalf("admitted diagnostic cache = %+v", diagnostics)
	}
}

func TestPushDiagnosticsDoesNotSubmitIneligibleEnrichment(t *testing.T) {
	for _, test := range []struct {
		name    string
		source  string
		entries config.RslintConfig
	}{
		{
			name:   "syntax error",
			source: "const value = ;\n",
			entries: config.RslintConfig{{
				Plugins: []string{"progressive"},
				Rules:   config.Rules{"progressive/check": "error"},
			}},
		},
		{
			name:   "no plugin work",
			source: "const value = 1;\n",
			entries: config.RslintConfig{{
				Rules: config.Rules{"no-debugger": "error"},
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			s, uri, queue := newPushDiagnosticsPluginFixture(t, test.source, test.entries)
			dispatched := make(chan struct{}, 1)
			s.eslintPluginDispatch = func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
				dispatched <- struct{}{}
				return &linter.EslintPluginLintResult{}, nil
			}
			s.pushDiagnostics(uri)
			select {
			case <-queue:
			default:
				t.Fatal("baseline diagnostics were not published")
			}
			s.inflightPluginDispatchMu.Lock()
			_, inflight := s.inflightPluginDispatch[uri]
			s.inflightPluginDispatchMu.Unlock()
			if inflight {
				t.Fatal("ineligible enrichment registered an async lifecycle")
			}
			select {
			case <-dispatched:
				t.Fatal("ineligible enrichment dispatched plugin work")
			default:
			}
			select {
			case result := <-s.pluginResultCh:
				t.Fatalf("ineligible enrichment produced a result: %+v", result)
			default:
			}
		})
	}
}

// ======== production prepared plugin path ========

// TestPreparedPluginInputUsesExactPassContent pins the fixAll invariant: the
// prepared input must carry the in-progress content of this observation, not
// stale bytes from the live editor mirror.
func TestPreparedPluginInputUsesExactPassContent(t *testing.T) {
	s := newTestServer()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {
			{
				Plugins: []string{"tplsp"},
				Rules:   config.Rules{"tplsp/no-foo": "error"},
			},
		},
	})
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	s.documents[uri] = "const stale = 0;"

	request := preparedPluginRequestForTest(
		t,
		s,
		uri,
		"const current = 1;",
		s.documentLintSnapshot(uri),
	)
	input := request.Files[0]
	if input.Text == nil || *input.Text != "const current = 1;" {
		t.Errorf("prepared pass used stale text: %v", input.Text)
	}
	if input.ConfigKey != "/proj" {
		t.Errorf("configKey = %q, want /proj", input.ConfigKey)
	}
	if len(request.Rules) != 1 {
		t.Errorf("expected only the plugin rule forwarded, got %+v", request.Rules)
	} else if _, ok := request.Rules["tplsp/no-foo"]; !ok {
		t.Errorf("expected tplsp/no-foo, got %+v", request.Rules)
	}
}

func TestPreparedPluginInputRespectsFiles(t *testing.T) {
	s := newTestServer()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {
			{
				Files:   []string{"**/*.ts"},
				Plugins: []string{"tplfiles"},
				Rules:   config.Rules{"tplfiles/no-foo": "error"},
			},
		},
	})
	s.documents["file:///proj/matched.ts"] = "foo();"
	s.documents["file:///proj/outside.js"] = "foo();"

	matchedURI := lsproto.DocumentUri("file:///proj/matched.ts")
	request, ok := pluginRequestForProductionPassTest(t, s, matchedURI, s.documents[matchedURI], s.documentLintSnapshot(matchedURI))
	if !ok || len(request.Files) != 1 {
		t.Fatalf("matching TS file plugin request = %+v, want one file", request)
	} else if _, configured := request.Rules["tplfiles/no-foo"]; !configured || len(request.Rules) != 1 {
		t.Fatalf("expected tplfiles/no-foo for matching file, got %+v", request.Rules)
	}

	outsideURI := lsproto.DocumentUri("file:///proj/outside.js")
	if _, ok := pluginRequestForProductionPassTest(t, s, outsideURI, s.documents[outsideURI], s.documentLintSnapshot(outsideURI)); ok {
		t.Fatal("files-scope miss produced plugin run")
	}
}

func TestPreparedPluginInputPreservesTargetIdentityAcrossAuthoredConfigBases(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "workspace")
	configDir := filepath.Join(root, "physical", "package")
	physicalSourceDir := filepath.Join(configDir, "src")
	aliasSourceDir := filepath.Join(workspace, "linked-src")
	for _, directory := range []string{workspace, physicalSourceDir} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(physicalSourceDir, aliasSourceDir); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	aliasFile := filepath.Join(aliasSourceDir, "index.ts")
	if err := os.WriteFile(aliasFile, []byte("foo();\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	effective := config.ConfigWithAuthoredPathBase(config.RslintConfig{{
		Files:   []string{"linked-src/**/*.ts"},
		Plugins: []string{"tpauthoredbase"},
		Rules:   config.Rules{"tpauthoredbase/no-foo": "error"},
	}}, workspace)

	s := newTestServer()
	installRuleCatalogForTest(s, effective)
	s.cwd = workspace
	s.fs = osvfs.FS()
	uri := documentURIFromPath(aliasFile)
	s.documents[uri] = "foo();\n"
	request, ok := pluginRequestForProductionPassTest(
		t,
		s,
		uri,
		s.documents[uri],
		documentLintSnapshotForTest(s, uri, effective, configDir, true, nil),
	)
	if !ok || len(request.Files) != 1 {
		t.Fatal("workspace-authored plugin rule lost the lexical symlink target")
	}
	if _, configured := request.Rules["tpauthoredbase/no-foo"]; !configured || len(request.Rules) != 1 {
		t.Fatalf("unexpected plugin rules: %+v", request.Rules)
	}
}

func TestPreparedPluginInputExternalConfigPreservesGitignoreTargetIdentity(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "workspace")
	configDir := filepath.Join(root, "physical", "package")
	physicalSourceDir := filepath.Join(configDir, "src")
	aliasSourceDir := filepath.Join(workspace, "linked-src")
	for _, directory := range []string{workspace, physicalSourceDir} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(physicalSourceDir, aliasSourceDir); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	aliasFile := filepath.Join(aliasSourceDir, "ignored.ts")
	if err := os.WriteFile(aliasFile, []byte("foo();\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(workspace, ".gitignore"),
		[]byte("linked-src/ignored.ts\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	effective := config.ConfigWithGitignoreForTargetsFromRoot(
		config.RslintConfig{{
			Files:   []string{"src/**/*.ts"},
			Plugins: []string{"tpexternalgit"},
			Rules:   config.Rules{"tpexternalgit/no-foo": "error"},
		}},
		configDir,
		workspace,
		osvfs.FS(),
		[]string{aliasFile},
		nil,
	)

	s := newTestServer()
	installRuleCatalogForTest(s, effective)
	s.cwd = workspace
	s.fs = osvfs.FS()
	uri := documentURIFromPath(aliasFile)
	s.documents[uri] = "foo();\n"
	if _, ok := pluginRequestForProductionPassTest(
		t,
		s,
		uri,
		s.documents[uri],
		documentLintSnapshotForTest(s, uri, effective, configDir, true, nil),
	); ok {
		t.Fatal("external-config Git ignore produced plugin input")
	}
}

func TestPreparedPluginInputRespectsGitignore(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ignored.ts")
	if err := os.WriteFile(target, []byte("foo();\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("ignored.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := newTestServer()
	s.cwd = dir
	s.fs = osvfs.FS()
	effectiveConfig := config.RslintConfig{{
		Plugins: []string{"tpgitignore"},
		Rules:   config.Rules{"tpgitignore/no-foo": "error"},
	}}
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		tspath.NormalizePath(dir): effectiveConfig,
	})
	uri := documentURIFromPath(target)
	s.documents[uri] = "foo();\n"
	if _, ok := pluginRequestForProductionPassTest(t, s, uri, s.documents[uri], s.documentLintSnapshot(uri)); ok {
		t.Fatal("gitignored file produced plugin input")
	}
}

func TestPreparedPluginInputUsesEffectiveJSConfigSnapshot(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "source.ts")
	if err := os.WriteFile(targetPath, []byte("foo();\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newTestServer()
	s.cwd = dir
	s.fs = osvfs.FS()
	jsConfig := config.RslintConfig{{
		Plugins: []string{"tpsnapshot"},
		Rules:   config.Rules{"tpsnapshot/no-foo": "error"},
	}}
	s.jsConfigs = map[string]config.RslintConfig{tspath.NormalizePath(dir): jsConfig}
	s.jsConfigOwnerIndex = target.NewOwnerIndex(s.jsConfigs, s.fs)
	installRuleCatalogForTest(s, jsConfig)
	uri := documentURIFromPath(targetPath)
	s.documents[uri] = "foo();\n"

	frozen := s.documentLintSnapshot(uri)
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("source.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	request, ok := pluginRequestForProductionPassTest(t, s, uri, s.documents[uri], frozen)
	if !ok || len(request.Files) != 1 {
		t.Fatal("captured effective config changed after .gitignore update")
	} else if _, configured := request.Rules["tpsnapshot/no-foo"]; !configured || len(request.Rules) != 1 {
		t.Fatalf("captured effective config produced unexpected input: %+v", request)
	}
	if _, ok := pluginRequestForProductionPassTest(t, s, uri, s.documents[uri], s.documentLintSnapshot(uri)); ok {
		t.Fatal("fresh effective config did not observe .gitignore update")
	}
}

func TestDocumentLintSnapshotKeepsPluginRulesInsideDeclaredOwner(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	rootFile := filepath.Join(root, "root.ts")
	nestedFile := filepath.Join(nested, "nested.ts")
	for _, filePath := range []string{rootFile, nestedFile} {
		if err := os.WriteFile(filePath, []byte("debugger;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	s := newTestServer()
	s.cwd = root
	s.fs = osvfs.FS()
	s.configSnapshotIncludesGitignore = true
	rootOwnerDirectory := tspath.NormalizePath(root)
	nestedOwnerDirectory := tspath.NormalizePath(nested)
	rootConfig := config.RslintConfig{{Rules: config.Rules{
		"community/check": "error",
		"no-debugger":     "error",
	}}}
	nestedConfig := config.RslintConfig{{
		Plugins: []string{"community"},
		Rules:   config.Rules{"community/check": "error"},
	}}
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		rootOwnerDirectory:   rootConfig,
		nestedOwnerDirectory: nestedConfig,
	})

	rootURI := documentURIFromPath(rootFile)
	s.documents[rootURI] = "debugger;\n"
	rootSnapshot := s.documentLintSnapshot(rootURI)
	if rootSnapshot.configKey != rootOwnerDirectory {
		t.Fatalf("root config key = %q, want %q", rootSnapshot.configKey, rootOwnerDirectory)
	}
	if pluginRule, ok := rootSnapshot.ruleCatalog.Lookup("community/check"); !ok || !pluginRule.IsEslintPluginRule {
		t.Fatalf("catalog generation lost its object-plugin rule: %+v", pluginRule)
	}
	if len(rootSnapshot.resolvedConfig.EnabledRules) != 1 ||
		rootSnapshot.resolvedConfig.EnabledRules[0].Name != "no-debugger" {
		t.Fatalf("owner without a plugin declaration resolved plugin rules: %+v", rootSnapshot.resolvedConfig.EnabledRules)
	}
	if _, ok := pluginRequestForProductionPassTest(t, s, rootURI, s.documents[rootURI], rootSnapshot); ok {
		t.Fatal("owner without a plugin declaration produced an object-plugin request")
	}

	nestedSnapshot := s.documentLintSnapshot(documentURIFromPath(nestedFile))
	if nestedSnapshot.configKey != nestedOwnerDirectory {
		t.Fatalf("nested config key = %q, want %q", nestedSnapshot.configKey, nestedOwnerDirectory)
	}
	if pluginRule, ok := nestedSnapshot.ruleCatalog.Lookup("community/check"); !ok || !pluginRule.IsEslintPluginRule {
		t.Fatalf("JS owner lost its object-plugin rule: %+v", pluginRule)
	}
	if len(nestedSnapshot.resolvedConfig.EnabledRules) != 1 ||
		nestedSnapshot.resolvedConfig.EnabledRules[0].Name != "community/check" {
		t.Fatalf("JS owner did not resolve its object-plugin rule: %+v", nestedSnapshot.resolvedConfig.EnabledRules)
	}
}

func TestRuleCatalogGenerationDoesNotRetainRemovedPlugin(t *testing.T) {
	s := newTestServer()
	installRuleCatalogForTest(s, config.RslintConfig{{
		Plugins: []string{"current"},
		Rules:   config.Rules{"current/check": "error"},
	}})
	if _, ok := s.ruleCatalog.Lookup("current/check"); !ok {
		t.Fatal("plugin rule missing from current generation catalog")
	}

	installRuleCatalogForTest(s, config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}})
	if _, ok := s.ruleCatalog.Lookup("current/check"); ok {
		t.Fatal("removed plugin rule leaked into the next generation catalog")
	}
}

func TestPreparedPluginInputRespectsDefaultExcludedDirectories(t *testing.T) {
	s := newTestServer()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {{
			Plugins: []string{"tplexcluded"},
			Rules:   config.Rules{"tplexcluded/no-foo": "error"},
		}},
	})

	for _, uri := range []lsproto.DocumentUri{
		"file:///proj/node_modules/pkg/index.ts",
		"file:///proj/.git/hooks/pre-commit.ts",
	} {
		s.documents[uri] = "foo();"
		if _, ok := pluginRequestForProductionPassTest(t, s, uri, s.documents[uri], s.documentLintSnapshot(uri)); ok {
			t.Fatalf("default-excluded file %q produced plugin input", uri)
		}
	}
}

func TestPreparedPluginInputNestedEncodedConfigKeyAndCwd(t *testing.T) {
	s := newTestServer()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/Users/John Doe/my project": {
			{
				Files:   []string{"**/*.ts"},
				Plugins: []string{"tprootcfg"},
				Rules:   config.Rules{"tprootcfg/no-root": "error"},
			},
		},
		"/Users/John Doe/my project/packages/foo": {
			{
				Files:   []string{"src/**/*.ts"},
				Plugins: []string{"tpnestedcfg"},
				Rules:   config.Rules{"tpnestedcfg/no-foo": "error"},
			},
		},
	})
	uri := lsproto.DocumentUri("file:///Users/John%20Doe/my%20project/packages/foo/src/index.ts")
	s.documents[uri] = "foo();"

	request := preparedPluginRequestForTest(t, s, uri, s.documents[uri], s.documentLintSnapshot(uri))
	in := request.Files[0]
	if in.ConfigKey != "/Users/John Doe/my project/packages/foo" {
		t.Fatalf("configKey = %q, want decoded nested config path", in.ConfigKey)
	}
	if in.Path != "/Users/John Doe/my project/packages/foo/src/index.ts" {
		t.Fatalf("path = %q, want decoded filesystem path", in.Path)
	}
	if _, configured := request.Rules["tpnestedcfg/no-foo"]; !configured || len(request.Rules) != 1 {
		t.Fatalf("expected only the nested plugin rule, got %+v", request.Rules)
	}
}

// TestDeferredPluginRunRebuildsWithFixes verifies the opaque continuation turns a
// mocked worker result into a RuleDiagnostic carrying the fix, with the
// configured severity reattached.
func TestDeferredPluginRunRebuildsWithFixes(t *testing.T) {
	s := newTestServer()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {
			{
				Plugins: []string{"tplsync"},
				Rules:   config.Rules{"tplsync/no-bar": "error"},
			},
		},
	})
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	content := "const bar = 1;"

	// Mock the reverse dispatcher: one diagnostic with a fix, keyed by the
	// resolved file path (applyEslintPluginResults matches result→input on it).
	s.eslintPluginDispatch = func(_ context.Context, _ linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		return &linter.EslintPluginLintResult{
			Results: []linter.EslintPluginFileResult{
				{
					FilePath: "/proj/a.ts",
					Diagnostics: []linter.EslintPluginDiagnostic{
						{
							RuleName: "tplsync/no-bar",
							Message:  "no bar",
							StartPos: 6,
							EndPos:   9,
							Fixes:    []linter.EslintPluginFix{{Range: [2]int{6, 9}, Text: "baz"}},
						},
					},
				},
			},
		}, nil
	}

	snapshot := s.documentLintSnapshot(uri)
	work := preparedPluginRunForTest(t, s, uri, content, snapshot)
	outcome, err := work(context.Background(), s.pluginDispatchForGeneration(snapshot.pluginGeneration))
	if err != nil || outcome.DispatchError != nil {
		t.Fatalf("plugin run failed: contract=%v dispatch=%v", err, outcome.DispatchError)
	}
	diags := outcome.Diagnostics
	if len(diags) != 1 {
		t.Fatalf("expected 1 plugin diagnostic, got %d", len(diags))
	}
	if diags[0].RuleName != "tplsync/no-bar" {
		t.Errorf("ruleName = %q, want tplsync/no-bar", diags[0].RuleName)
	}
	if diags[0].Severity != rule.SeverityError {
		t.Errorf("severity = %v, want SeverityError (reattached from config)", diags[0].Severity)
	}
	fixes := diags[0].Fixes()
	if len(fixes) != 1 || fixes[0].Text != "baz" {
		t.Fatalf("expected one fix with text 'baz', got %+v", fixes)
	}
	if fixes[0].Range.Pos() != 6 || fixes[0].Range.End() != 9 {
		t.Errorf("fix range = [%d,%d], want [6,9]", fixes[0].Range.Pos(), fixes[0].Range.End())
	}

}

// ======== computeFixAllContent: native+plugin fold loop ========

// TestComputeFixAllContent_FoldsPluginFixes drives the fixAll multi-pass loop
// with an injected native lint (no TS session) and a mocked plugin dispatcher,
// asserting that BOTH a native fix and a plugin fix apply in the same pass (the
// fold), the plugin fix is not clobbered by the native one, and the loop
// converges across passes on the in-progress content.
func TestComputeFixAllContent_FoldsPluginFixes(t *testing.T) {
	s := newTestServer()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {
			{
				Plugins: []string{"tpfold"},
				Rules:   config.Rules{"tpfold/no-bar": "error"},
			},
		},
	})
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	const original = "const bar = 1;" // "bar" at [6,9], "1" at [12,13]
	s.documents[uri] = original

	// Mocked plugin dispatcher: fix "bar" → "baz" wherever it appears in the
	// content the worker was handed (req file text == the current pass content).
	var pluginPasses int
	s.eslintPluginDispatch = func(_ context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		pluginPasses++
		f := req.Files[0]
		content := ""
		if f.Text != nil {
			content = *f.Text
		}
		idx := strings.Index(content, "bar")
		if idx < 0 {
			return &linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{FilePath: f.Path}}}, nil
		}
		return &linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{
			FilePath: f.Path,
			Diagnostics: []linter.EslintPluginDiagnostic{{
				RuleName: "tpfold/no-bar", Message: "no bar", StartPos: idx, EndPos: idx + 3,
				Fixes: []linter.EslintPluginFix{{Range: [2]int{idx, idx + 3}, Text: "baz"}},
			}},
		}}}, nil
	}

	// Extend the production generation with one native test rule that fixes the
	// first "1" → "2" and then converges.
	var nativePasses int
	s.speculativeGeneration = fixAllGenerationWithNativeFixForTest(t, &nativePasses, nil, func(content string) (core.TextRange, string, bool) {
		idx := strings.Index(content, "1")
		if idx < 0 {
			return core.TextRange{}, "", false
		}
		return core.NewTextRange(idx, idx+1), "2", true
	})

	got := runSpeculativeFixAllForTest(t, s, context.Background(), uri, original, s.documentLintSnapshot(uri))

	// Both fixes applied (native "1"→"2" AND plugin "bar"→"baz") proves the
	// fold: the plugin fix survived alongside the native one in the same pass.
	if got != "const baz = 2;" {
		t.Fatalf("fix-all content = %q, want %q (native+plugin folded)", got, "const baz = 2;")
	}
	// Pass 0 fixes both; pass 1 sees neither "1" nor "bar" → no fix → converge.
	if nativePasses != 2 || pluginPasses != 2 {
		t.Errorf("expected 2 native + 2 plugin passes (1 fixing, 1 converging), got native=%d plugin=%d", nativePasses, pluginPasses)
	}
}

type retargetingDocumentFS struct {
	mockFS
	targetPath string
	targets    []string
	targetCall int
}

type retargetingConfigEvaluationFS struct {
	mockFS
	targetPath       string
	targetParentPath string
	configPath       string
	configCalls      int
}

func (fsys *retargetingConfigEvaluationFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	switch filePath {
	case fsys.targetPath:
		return "/owner-a/src/index.ts"
	case fsys.targetParentPath:
		return "/owner-a/src"
	case fsys.configPath:
		fsys.configCalls++
		if fsys.configCalls == 1 {
			return "/owner-a"
		}
		return "/owner-b"
	default:
		return filePath
	}
}

func TestDocumentLintSnapshotCollectsGitignoreFromFrozenTarget(t *testing.T) {
	uri := lsproto.DocumentUri("file:///alias/source.ts")
	filePath := tspath.NormalizePath(uriToPath(uri))
	fsys := &retargetingDocumentFS{
		targetPath: filePath,
		targets:    []string{"/owner-a/source.ts", "/owner-b/source.ts"},
	}
	s := newTestServer()
	s.cwd = "/alias"
	s.fs = fsys
	installJSConfigsForTest(s, map[string]config.RslintConfig{"/alias": {{}}})

	snapshot := s.documentLintSnapshot(uri)
	if snapshot.target.CanonicalPath != "/owner-a/source.ts" {
		t.Fatalf("snapshot target = %+v", snapshot.target)
	}
	if fsys.targetCall != 1 {
		t.Fatalf("target Realpath calls = %d, want one identity observation", fsys.targetCall)
	}
}

func TestDocumentLintSnapshotFreezesBootstrapConfigEvaluation(t *testing.T) {
	uri := lsproto.DocumentUri("file:///outside-link/src/index.ts")
	filePath := tspath.NormalizePath(uriToPath(uri))
	configPath := "/config-link"
	fsys := &retargetingConfigEvaluationFS{
		targetPath:       filePath,
		targetParentPath: tspath.GetDirectoryPath(filePath),
		configPath:       configPath,
	}
	s := newTestServer()
	s.cwd = configPath
	s.fs = fsys
	effectiveConfig := config.RslintConfig{{
		Files: []string{"src/*.ts"},
		Rules: config.Rules{"no-debugger": "error"},
	}}
	s.jsConfigs = map[string]config.RslintConfig{configPath: effectiveConfig}
	installRuleCatalogForTest(s, effectiveConfig)

	snapshot := s.documentLintSnapshot(uri)
	if snapshot.target.CanonicalPath != "/owner-a/src/index.ts" ||
		snapshot.resolvedConfig.MergedConfig == nil {
		t.Fatalf("bootstrap snapshot mixed config generations: %+v", snapshot)
	}
	if fsys.configCalls != 1 {
		t.Fatalf("config directory Realpath calls = %d, want one evaluation observation", fsys.configCalls)
	}
}

func TestNativeFixPassUsesFrozenTargetOverlay(t *testing.T) {
	uri := lsproto.DocumentUri("file:///alias/source.ts")
	filePath := tspath.NormalizePath(uriToPath(uri))
	fsys := &retargetingDocumentFS{
		targetPath: filePath,
		targets:    []string{"/owner-a/source.ts", "/owner-b/source.ts"},
	}
	s := newTestServer()
	s.cwd = "/alias"
	s.fs = fsys
	s.configSnapshotIncludesGitignore = true
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/alias": {{Rules: config.Rules{"no-debugger": "error"}}},
	})
	const content = "debugger;\n"
	s.documents[uri] = content
	snapshot := s.documentLintSnapshot(uri)

	environment := s.freezeSpeculativeLintEnvironment(uri, snapshot.target)
	generation, release, err := acquireSpeculativeGeneration(context.Background(), content, snapshot, environment)
	if err != nil {
		t.Fatal(err)
	}
	result, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(
		linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
			return generation, release, nil
		}),
		linter.ObservationPolicy{
			Demand: linter.ArtifactDemand{Native: rule.EditDemandAutofix},
		},
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := result.Observation.Native.Diagnostics
	if len(diagnostics) != 1 || diagnostics[0].RuleName != "no-debugger" {
		t.Fatalf("native diagnostics = %+v", diagnostics)
	}
	if fsys.targetCall != 1 {
		t.Fatalf("target Realpath calls = %d, want one frozen observation", fsys.targetCall)
	}
}

func (fsys *retargetingDocumentFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if filePath != fsys.targetPath {
		return filePath
	}
	index := fsys.targetCall
	fsys.targetCall++
	if index >= len(fsys.targets) {
		index = len(fsys.targets) - 1
	}
	return fsys.targets[index]
}

func TestComputeFixAllContentSharesFrozenTargetWithNativeAndPlugin(t *testing.T) {
	uri := lsproto.DocumentUri("file:///alias/source.ts")
	filePath := tspath.NormalizePath(uriToPath(uri))
	fsys := &retargetingDocumentFS{
		targetPath: filePath,
		targets:    []string{"/owner-a/source.ts", "/owner-b/source.ts"},
	}
	s := newTestServer()
	s.cwd = "/alias"
	s.fs = fsys
	s.configSnapshotIncludesGitignore = true
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/owner-a": {{
			Files:   []string{"**/*.ts"},
			Plugins: []string{"tpfrozentarget"},
			Rules:   config.Rules{"tpfrozentarget/owner-a": "error"},
		}},
		"/owner-b": {{
			Files:   []string{"**/*.ts"},
			Plugins: []string{"tpfrozentarget"},
			Rules:   config.Rules{"tpfrozentarget/owner-b": "error"},
		}},
	})
	s.tsConfigPathsByConfig = map[string][]string{
		"/owner-a": {"/owner-a/tsconfig.json"},
		"/owner-b": {"/owner-b/tsconfig.json"},
	}
	const source = "const value = 1;"
	s.documents[uri] = source
	s.eslintPluginConfigGeneration = "frozen-plugin-generation"

	snapshot := s.documentLintSnapshot(uri)
	s.eslintPluginConfigGeneration = "newer-live-generation"
	if snapshot.target.CanonicalPath != "/owner-a/source.ts" ||
		snapshot.target.ConfigDirectory != "/owner-a" ||
		len(snapshot.typeScriptConfigPaths) != 1 ||
		snapshot.typeScriptConfigPaths[0] != "/owner-a/tsconfig.json" {
		t.Fatalf("snapshot mixed target/config/project owners: %+v", snapshot)
	}

	nativeCalls := 0
	s.speculativeGeneration = func(
		ctx context.Context,
		content string,
		got documentLintSnapshot,
		environment speculativeLintEnvironment,
	) (linter.Generation, linter.ReleaseFunc, error) {
		nativeCalls++
		if got.target != snapshot.target ||
			len(got.typeScriptConfigPaths) != 1 ||
			got.typeScriptConfigPaths[0] != "/owner-a/tsconfig.json" {
			t.Fatalf("native pass received a different snapshot: %+v", got)
		}
		// Project ownership is asserted above; generation materialization uses a
		// source-only Program so this unit test needs no physical tsconfig fixture.
		pluginSnapshot := got
		pluginSnapshot.typeScriptConfigPaths = nil
		return acquireSpeculativeGeneration(ctx, content, pluginSnapshot, environment)
	}
	pluginCalls := 0
	s.eslintPluginDispatch = func(
		_ context.Context,
		req linter.EslintPluginLintRequest,
	) (*linter.EslintPluginLintResult, error) {
		pluginCalls++
		if req.Generation != "frozen-plugin-generation" {
			t.Fatalf("plugin generation = %q, want frozen snapshot token", req.Generation)
		}
		_, hasOwnerA := req.Rules["tpfrozentarget/owner-a"]
		if len(req.Files) != 1 || len(req.Rules) != 1 || !hasOwnerA ||
			req.Files[0].ConfigKey != "/owner-a" {
			t.Fatalf("plugin pass received a different owner: %+v", req.Files)
		}
		return &linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{
			FilePath: req.Files[0].Path,
		}}}, nil
	}

	if got := runSpeculativeFixAllForTest(t, s, context.Background(), uri, source, snapshot); got != source {
		t.Fatalf("fixAll changed source without fixes: %q", got)
	}
	if nativeCalls != 1 || pluginCalls != 1 {
		t.Fatalf("native/plugin calls = %d/%d, want 1/1", nativeCalls, pluginCalls)
	}
	if fsys.targetCall != 1 {
		t.Fatalf("target Realpath calls = %d, want one frozen observation", fsys.targetCall)
	}
}

// TestComputeFixAllContent_PluginTimeoutFallsBackNativeOnly asserts the
// source.fixAll plugin reverse requests are bounded by a deadline: a wedged
// client that never answers must not freeze the dispatch loop. On timeout the
// plugin pass is dropped and native fixes still apply.
func TestComputeFixAllContent_PluginTimeoutFallsBackNativeOnly(t *testing.T) {
	s := newTestServer()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {
			{
				Plugins: []string{"tptimeout"},
				Rules:   config.Rules{"tptimeout/no-bar": "error"},
			},
		},
	})
	s.pluginReverseTimeout = 20 * time.Millisecond
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	const original = "const bar = 1;"
	s.documents[uri] = original

	// Wedged dispatcher: never answers, blocks until the deadline cancels ctx.
	var dispatchCalls int
	s.eslintPluginDispatch = func(ctx context.Context, _ linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		dispatchCalls++
		<-ctx.Done()
		return nil, ctx.Err()
	}
	// Extend each production generation with one native test fix.
	var nativePasses int
	s.speculativeGeneration = fixAllGenerationWithNativeFixForTest(t, &nativePasses, nil, func(content string) (core.TextRange, string, bool) {
		idx := strings.Index(content, "1")
		if idx < 0 {
			return core.TextRange{}, "", false
		}
		return core.NewTextRange(idx, idx+1), "2", true
	})

	start := time.Now()
	got := runSpeculativeFixAllForTest(t, s, context.Background(), uri, original, s.documentLintSnapshot(uri))
	elapsed := time.Since(start)

	// Native fix applied; the wedged plugin pass timed out and was dropped
	// ("bar" stays, only "1"→"2").
	if got != "const bar = 2;" {
		t.Fatalf("fix-all content = %q, want %q (native-only after plugin timeout)", got, "const bar = 2;")
	}
	// The whole fixAll is bounded by the shared plugin budget — without the
	// deadline the wedged dispatch would hang the dispatch loop forever.
	if elapsed > 2*time.Second {
		t.Errorf("fixAll took %v; the plugin deadline should bound the stall", elapsed)
	}
	// The plugin pass is dispatched EXACTLY ONCE: pass 0 invokes it (and times
	// out); every later pass is skipped because pluginCtx is already expired. A
	// regression that re-dispatched on the expired ctx would send a wasted
	// reverse request to the client per remaining pass.
	if dispatchCalls != 1 {
		t.Errorf("expected exactly 1 plugin dispatch (pass 0; later passes skip on expiry), got %d", dispatchCalls)
	}
}

func TestComputeFixAllContent_SyntaxErrorSkipsPluginPass(t *testing.T) {
	s := newTestServer()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {{
			Plugins: []string{"tpsyntax"},
			Rules:   config.Rules{"tpsyntax/check": "error"},
		}},
	})
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	const malformed = "const value = ;"
	s.documents[uri] = malformed

	pluginCalls := 0
	s.eslintPluginDispatch = func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		pluginCalls++
		return &linter.EslintPluginLintResult{}, nil
	}
	got := runSpeculativeFixAllForTest(
		t, s, context.Background(), uri, malformed, s.documentLintSnapshot(uri),
	)
	if got != malformed {
		t.Fatalf("syntax-error fixAll changed content to %q", got)
	}
	if pluginCalls != 0 {
		t.Fatalf("syntax-error fixAll dispatched %d plugin passes, want 0", pluginCalls)
	}
}

func TestDeferredPluginRunExpiredContextReturnsPromptly(t *testing.T) {
	s := newTestServer()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {
			{
				Plugins: []string{"tpexpired"},
				Rules:   config.Rules{"tpexpired/no-bar": "error"},
			},
		},
	})
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	content := "const bar = 1;"
	snapshot := s.documentLintSnapshot(uri)
	work := preparedPluginRunForTest(t, s, uri, content, snapshot)

	s.eslintPluginDispatch = func(ctx context.Context, _ linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already expired before the call

	start := time.Now()
	outcome, err := work(ctx, s.pluginDispatchForGeneration(snapshot.pluginGeneration))
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("plugin run contract error: %v", err)
	}
	if !errors.Is(outcome.DispatchError, context.Canceled) || outcome.Diagnostics != nil {
		t.Errorf("expired ctx outcome = %+v, want canceled native-only result", outcome)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("expired ctx should return promptly, took %v", elapsed)
	}
}

// ======== codeAction coexistence (native fix + plugin fix/suggestion) ========

func codeActionsByTitle(resp lsproto.CodeActionResponse) map[string]*lsproto.CodeAction {
	out := map[string]*lsproto.CodeAction{}
	if resp.CommandOrCodeActionArray == nil {
		return out
	}
	for _, ca := range *resp.CommandOrCodeActionArray {
		if ca.CodeAction != nil {
			out[ca.CodeAction.Title] = ca.CodeAction
		}
	}
	return out
}

// TestHandleCodeAction_NativeAndPluginFixesCoexistOnSameRange pins the dominant
// per-line lightbulb path: a native diagnostic and a community-plugin
// diagnostic overlap the SAME range, both fixable. The quickfix-assembly body
// (run elsewhere only against an empty diagnostics map) must produce a "Fix:"
// action for EACH origin.
func TestHandleCodeAction_NativeAndPluginFixesCoexistOnSameRange(t *testing.T) {
	s := newTestServer()
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	text := "const bar = 1;"
	s.documents[uri] = text
	s.diagnostics[uri] = []rule.RuleDiagnostic{
		{
			RuleName:   "native/x",
			Range:      core.NewTextRange(6, 9),
			Message:    rule.RuleMessage{Description: "native msg"},
			SourceFile: textOnlySourceFile{text: text},
			FixesPtr:   &[]rule.RuleFix{{Range: core.NewTextRange(6, 9), Text: "NAT"}},
		},
		{
			RuleName:   "plug/y",
			Range:      core.NewTextRange(6, 9),
			Message:    rule.RuleMessage{Description: "plugin msg"},
			SourceFile: textOnlySourceFile{text: text},
			FixesPtr:   &[]rule.RuleFix{{Range: core.NewTextRange(6, 9), Text: "PLG"}},
		},
	}

	resp, err := s.handleCodeAction(context.Background(), &lsproto.CodeActionParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Range: lsproto.Range{
			Start: lsproto.Position{Line: 0, Character: 6},
			End:   lsproto.Position{Line: 0, Character: 9},
		},
		Context: &lsproto.CodeActionContext{},
	})
	if err != nil {
		t.Fatalf("handleCodeAction: %v", err)
	}
	byTitle := codeActionsByTitle(resp)
	if byTitle["Fix: native msg"] == nil {
		t.Errorf("missing native fix action; got titles %v", titleSet(byTitle))
	}
	if byTitle["Fix: plugin msg"] == nil {
		t.Errorf("missing plugin fix action; got titles %v", titleSet(byTitle))
	}
}

// TestHandleCodeAction_NativeFixAndPluginSuggestionCoexist pins that a native
// autofix and a plugin suggestion on the same file surface as distinct code
// actions, distinguished by preference: the native fix is IsPreferred, the
// plugin suggestion is not. createCodeActionFromSuggestion is otherwise
// uncovered.
func TestHandleCodeAction_NativeFixAndPluginSuggestionCoexist(t *testing.T) {
	s := newTestServer()
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	text := "const bar = 1;"
	s.documents[uri] = text
	s.diagnostics[uri] = []rule.RuleDiagnostic{
		{
			RuleName:   "native/x",
			Range:      core.NewTextRange(12, 13),
			Message:    rule.RuleMessage{Description: "native msg"},
			SourceFile: textOnlySourceFile{text: text},
			FixesPtr:   &[]rule.RuleFix{{Range: core.NewTextRange(12, 13), Text: "2"}},
		},
		{
			RuleName:   "plug/y",
			Range:      core.NewTextRange(6, 9),
			Message:    rule.RuleMessage{Description: "plugin msg"},
			SourceFile: textOnlySourceFile{text: text},
			Suggestions: &[]rule.RuleSuggestion{{
				Message:  rule.RuleMessage{Description: "use baz"},
				FixesArr: []rule.RuleFix{{Range: core.NewTextRange(6, 9), Text: "baz"}},
			}},
		},
	}

	resp, err := s.handleCodeAction(context.Background(), &lsproto.CodeActionParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Range: lsproto.Range{
			Start: lsproto.Position{Line: 0, Character: 0},
			End:   lsproto.Position{Line: 0, Character: 14},
		},
		Context: &lsproto.CodeActionContext{},
	})
	if err != nil {
		t.Fatalf("handleCodeAction: %v", err)
	}
	byTitle := codeActionsByTitle(resp)
	nat := byTitle["Fix: native msg"]
	if nat == nil {
		t.Fatalf("missing native fix action; got titles %v", titleSet(byTitle))
		return
	}
	if nat.IsPreferred == nil || !*nat.IsPreferred {
		t.Error("native autofix must be IsPreferred=true")
	}
	sug := byTitle["Suggestion: use baz"]
	if sug == nil {
		t.Fatalf("missing plugin suggestion action; got titles %v", titleSet(byTitle))
		return
	}
	if sug.IsPreferred == nil || *sug.IsPreferred {
		t.Error("plugin suggestion must be IsPreferred=false")
	}
}

func titleSet(m map[string]*lsproto.CodeAction) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestDispatchPluginLintTimesOutWedgedClient pins that the background
// diagnostics dispatch does not leak its goroutine when a registered-but-
// unresponsive client never answers: pluginReverseTimeout bounds it.
// backgroundCtx alone only cancels at shutdown, so without the deadline the
// goroutine + its pendingServerRequests entry would leak on every relint.
func TestDispatchPluginLintTimesOutWedgedClient(t *testing.T) {
	s := newTestServer()
	s.backgroundCtx = context.Background()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {
			{
				Plugins: []string{"tpleak"},
				Rules:   config.Rules{"tpleak/no-bar": "error"},
			},
		},
	})
	s.pluginReverseTimeout = 30 * time.Millisecond
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	s.documents[uri] = "const bar = 1;"
	s.docGeneration[uri] = 1
	snapshot := s.documentLintSnapshot(uri)
	work := preparedPluginRunForTest(t, s, uri, s.documents[uri], snapshot)

	var logBuf syncBuffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	released := make(chan error, 1)
	s.eslintPluginDispatch = func(ctx context.Context, _ linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		<-ctx.Done() // wedged client: only the deadline releases us
		released <- ctx.Err()
		return nil, ctx.Err()
	}

	s.startDiagnosticEnrichment(s.backgroundCtx, uri, 1, snapshot.pluginGeneration, work)

	select {
	case err := <-released:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("dispatch released with %v, want context.DeadlineExceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("dispatch goroutine never released — the reverse request leaked")
	}

	// The DeadlineExceeded must be logged as a BENIGN timeout, not an rslint
	// "lint error" — at error severity it would spam every debounced relint.
	deadline := time.Now().Add(time.Second)
	for !strings.Contains(logBuf.String(), "timed out") {
		if time.Now().After(deadline) {
			t.Fatalf("expected a benign timeout log line, got %q", logBuf.String())
		}
		time.Sleep(2 * time.Millisecond)
	}
	if strings.Contains(logBuf.String(), "lint error") {
		t.Errorf("DeadlineExceeded mislabeled as an rslint error: %q", logBuf.String())
	}
}

// syncBuffer is a goroutine-safe log sink for asserting what
// dispatchPluginLint's background goroutine logs.
type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// TestDispatchPluginLintDeliversSuccessResultNotRacedAway pins that a
// successful lint's result reaches pluginResultCh: the send is preferred over
// the ctx.Done() drop, so a freshly-computed result is never raced away by a
// deadline that expires in the gap before the send (the buffered channel has
// room).
func TestDispatchPluginLintDeliversSuccessResultNotRacedAway(t *testing.T) {
	s := newTestServer()
	s.backgroundCtx = context.Background()
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {
			{
				Plugins: []string{"tpok"},
				Rules:   config.Rules{"tpok/no-bar": "error"},
			},
		},
	})
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	s.documents[uri] = "const bar = 1;"
	s.docGeneration[uri] = 7
	s.eslintPluginConfigGeneration = "prepared-generation"
	snapshot := s.documentLintSnapshot(uri)
	work := preparedPluginRunForTest(t, s, uri, s.documents[uri], snapshot)
	s.eslintPluginConfigGeneration = "newer-live-generation"

	dispatchedGeneration := make(chan string, 1)
	s.eslintPluginDispatch = func(ctx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		dispatchedGeneration <- req.Generation
		return &linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{
			FilePath:    req.Files[0].Path,
			Diagnostics: []linter.EslintPluginDiagnostic{{RuleName: "tpok/no-bar", Message: "bad", StartPos: 6, EndPos: 9}},
		}}}, nil
	}

	s.startDiagnosticEnrichment(s.backgroundCtx, uri, 7, snapshot.pluginGeneration, work)

	select {
	case r := <-s.pluginResultCh:
		if r.generation != 7 {
			t.Errorf("delivered generation %d, want 7", r.generation)
		}
		if len(r.diags) != 1 || r.diags[0].RuleName != "tpok/no-bar" {
			t.Errorf("expected the plugin diagnostic delivered, got %+v", r.diags)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("successful plugin result was not delivered to pluginResultCh")
	}
	if got := <-dispatchedGeneration; got != "prepared-generation" {
		t.Fatalf("plugin generation = %q, want prepared-generation", got)
	}
}

// TestSendCancelRequest_QueuesCancelNotification pins that sendCancelRequest
// emits a $/cancelRequest notification carrying the reverse-request id.
func TestSendCancelRequest_QueuesCancelNotification(t *testing.T) {
	s, queue := newTestServerWithQueue()
	s.sendCancelRequest(jsonrpc.NewIDString("ts42"))
	select {
	case msg := <-queue:
		req := msg.AsRequest()
		if req.Method != lsproto.MethodCancelRequest {
			t.Fatalf("method = %q, want %q", req.Method, lsproto.MethodCancelRequest)
		}
		cp, ok := req.Params.(*lsproto.CancelParams)
		if !ok {
			t.Fatalf("params type = %T, want *lsproto.CancelParams", req.Params)
		}
		if cp.Id.String == nil || *cp.Id.String != "ts42" {
			t.Errorf("cancel id = %+v, want string \"ts42\"", cp.Id)
		}
	default:
		t.Fatal("no $/cancelRequest notification was queued")
	}
}

// TestDispatchPluginLintSupersedeCancelsPrior pins the full supersede path: a
// newer keystroke's dispatch cancels the prior in-flight one Go-side AND sends
// the client a $/cancelRequest for its reverse-request id (so the Node worker
// stops instead of running to completion). Uses the real sendRequest path so
// automatic context-to-$/cancelRequest forwarding is exercised end to end.
func TestDispatchPluginLintSupersedeCancelsPrior(t *testing.T) {
	s, queue := newTestServerWithQueue()
	s.pendingServerRequests = make(map[jsonrpc.ID]chan *lsproto.ResponseMessage)
	s.backgroundCtx = context.Background()
	s.pluginReverseTimeout = 500 * time.Millisecond // backstop so residual goroutines exit
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {{Plugins: []string{"tpsup"}, Rules: config.Rules{"tpsup/no-bar": "error"}}},
	})
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	s.documents[uri] = "const bar = 1;"
	s.docGeneration[uri] = 1
	snapshot := s.documentLintSnapshot(uri)
	work := preparedPluginRunForTest(t, s, uri, s.documents[uri], snapshot)

	// First dispatch queues a reverse request, then blocks on a response that
	// never comes.
	s.startDiagnosticEnrichment(s.backgroundCtx, uri, 1, snapshot.pluginGeneration, work)

	var firstID *jsonrpc.ID
	select {
	case msg := <-queue:
		request := msg.AsRequest()
		if request.Method != methodPluginLint {
			t.Fatalf("first message = %q, want %q", request.Method, methodPluginLint)
		}
		firstID = request.ID
	case <-time.After(time.Second):
		t.Fatal("first reverse request was not sent")
	}

	// Supersede: a newer keystroke dispatches again — must cancel the first.
	s.docGeneration[uri] = 2
	s.cancelInflightPluginDispatch(uri)
	nextWork := preparedPluginRunForTest(t, s, uri, s.documents[uri], snapshot)
	s.startDiagnosticEnrichment(s.backgroundCtx, uri, 2, snapshot.pluginGeneration, nextWork)

	// The supersede must $/cancelRequest the prior reverse request, and it must
	// be the FIRST thing queued after the supersede — cancel runs synchronously
	// before dispatchPluginLint starts the new goroutine, so the new reverse
	// request — so a refactor that moved cancel below the new send fails here.
	select {
	case msg := <-queue:
		req := msg.AsRequest()
		if req.Method != lsproto.MethodCancelRequest {
			t.Fatalf("first message after supersede = %q, want %q (cancel must precede the new request)", req.Method, lsproto.MethodCancelRequest)
		}
		cp, ok := req.Params.(*lsproto.CancelParams)
		if !ok || cp.Id.String == nil || *cp.Id.String != firstID.String() {
			t.Fatalf("cancel targeted %+v, want the prior id %s", req.Params, firstID.String())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("supersede did not $/cancelRequest the prior reverse request")
	}
}

func TestDispatchPluginLintEmptyPlanCancelsPriorWithoutNewRequest(t *testing.T) {
	s, queue := newTestServerWithQueue()
	s.pendingServerRequests = make(map[jsonrpc.ID]chan *lsproto.ResponseMessage)
	s.backgroundCtx = context.Background()
	s.pluginReverseTimeout = 500 * time.Millisecond
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {{Plugins: []string{"tpfilescancel"}, Rules: config.Rules{"tpfilescancel/no-bar": "error"}}},
	})
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	s.documents[uri] = "const bar = 1;"
	s.docGeneration[uri] = 1
	snapshot := s.documentLintSnapshot(uri)
	work := preparedPluginRunForTest(t, s, uri, s.documents[uri], snapshot)

	s.startDiagnosticEnrichment(s.backgroundCtx, uri, 1, snapshot.pluginGeneration, work)

	var firstID *jsonrpc.ID
	select {
	case msg := <-queue:
		req := msg.AsRequest()
		if req.Method != methodPluginLint {
			t.Fatalf("first message = %q, want %q", req.Method, methodPluginLint)
		}
		firstID = req.ID
	case <-time.After(time.Second):
		t.Fatal("first reverse request was not sent")
	}

	// A newer prepared plan with no plugin run must still cancel the stale
	// in-flight worker from the previous document generation.
	s.docGeneration[uri] = 2
	s.cancelInflightPluginDispatch(uri)
	s.startDiagnosticEnrichment(s.backgroundCtx, uri, 2, snapshot.pluginGeneration, nil)

	select {
	case msg := <-queue:
		req := msg.AsRequest()
		if req.Method != lsproto.MethodCancelRequest {
			t.Fatalf("files-miss dispatch queued %q, want only %q", req.Method, lsproto.MethodCancelRequest)
		}
		cp, ok := req.Params.(*lsproto.CancelParams)
		if !ok || cp.Id.String == nil || *cp.Id.String != firstID.String() {
			t.Fatalf("cancel targeted %+v, want the prior id %s", req.Params, firstID.String())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("files-miss dispatch did not cancel the prior reverse request")
	}

	select {
	case msg := <-queue:
		t.Fatalf("files-miss dispatch must not send a new pluginLint request, got %q", msg.AsRequest().Method)
	default:
	}
}

// TestHandleDidClose_CancelsInflightDispatch pins that closing a document with
// an in-flight plugin dispatch cancels it (Go-side + $/cancelRequest) — the
// close path has no superseding keystroke to do it.
func TestHandleDidClose_CancelsInflightDispatch(t *testing.T) {
	s, queue := newTestServerWithQueue()
	s.pendingServerRequests = make(map[jsonrpc.ID]chan *lsproto.ResponseMessage)
	s.backgroundCtx = context.Background()
	s.pluginReverseTimeout = 500 * time.Millisecond
	installJSConfigsForTest(s, map[string]config.RslintConfig{
		"/proj": {{Plugins: []string{"tpclose"}, Rules: config.Rules{"tpclose/no-bar": "error"}}},
	})
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	s.documents[uri] = "const bar = 1;"
	s.docGeneration[uri] = 1
	snapshot := s.documentLintSnapshot(uri)
	work := preparedPluginRunForTest(t, s, uri, s.documents[uri], snapshot)

	s.startDiagnosticEnrichment(s.backgroundCtx, uri, 1, snapshot.pluginGeneration, work)

	var firstID *jsonrpc.ID
	select {
	case msg := <-queue:
		request := msg.AsRequest()
		if request.Method != methodPluginLint {
			t.Fatalf("first message = %q, want %q", request.Method, methodPluginLint)
		}
		firstID = request.ID
	case <-time.After(time.Second):
		t.Fatal("plugin lint request was not sent")
	}

	if err := s.handleDidClose(context.Background(), &lsproto.DidCloseTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
	}); err != nil {
		t.Fatalf("handleDidClose: %v", err)
	}

	// Close must $/cancelRequest the in-flight dispatch.
	select {
	case msg := <-queue:
		req := msg.AsRequest()
		if req.Method != lsproto.MethodCancelRequest {
			t.Fatalf("first message after close = %q, want %q", req.Method, lsproto.MethodCancelRequest)
		}
		cp, ok := req.Params.(*lsproto.CancelParams)
		if !ok || cp.Id.String == nil || *cp.Id.String != firstID.String() {
			t.Fatalf("cancel targeted %+v, want the in-flight id %s", req.Params, firstID.String())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handleDidClose did not $/cancelRequest the in-flight dispatch")
	}

	// Registry entry cleared.
	s.inflightPluginDispatchMu.Lock()
	_, stillThere := s.inflightPluginDispatch[uri]
	s.inflightPluginDispatchMu.Unlock()
	if stillThere {
		t.Error("in-flight dispatch entry should be removed after close")
	}
}
