package lsp

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/web-infra-dev/rslint/internal/config"
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

// ======== pluginConfigKeyForURI tests ========

func TestPluginConfigKeyForURI(t *testing.T) {
	s := newTestServer()
	// JS config registered under a URI key (as configUpdate stores it).
	s.jsConfigs["file:///project"] = config.RslintConfig{{}}

	tests := []struct {
		name string
		uri  lsproto.DocumentUri
		want string
	}{
		{"file directly under config dir", "file:///project/a.ts", "file:///project"},
		{"file in nested subdir", "file:///project/src/deep/b.ts", "file:///project"},
		{"file outside any config", "file:///other/c.ts", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.pluginConfigKeyForURI(tt.uri); got != tt.want {
				t.Errorf("pluginConfigKeyForURI(%q) = %q, want %q", tt.uri, got, tt.want)
			}
		})
	}
}

func TestPluginConfigKeyForURI_NoJSConfigs(t *testing.T) {
	s := newTestServer() // no jsConfigs → JSON fallback path, empty key
	if got := s.pluginConfigKeyForURI("file:///project/a.ts"); got != "" {
		t.Errorf("expected empty configKey with no JS configs, got %q", got)
	}
}

// ======== buildPluginFileInput / lintPluginRulesSync (fixAll path) ========

// TestBuildPluginFileInput_TextOverridePrecedence pins the fixAll invariant:
// the explicit text override (the in-progress fixed content of the current
// pass) must win over the editor overlay, else multi-pass fixAll would lint
// stale bytes and misplace plugin fix offsets.
func TestBuildPluginFileInput_TextOverridePrecedence(t *testing.T) {
	config.RegisterEslintPluginRules([]config.EslintPluginEntry{
		{Prefix: "tplsp", RuleNames: []string{"no-foo"}},
	})

	s := newTestServer()
	s.jsConfigs["file:///proj"] = config.RslintConfig{
		{
			Plugins: []string{"tplsp"},
			Rules:   config.Rules{"tplsp/no-foo": "error"},
		},
	}
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	s.documents[uri] = "overlay buffer"

	// nil override → editor overlay (the diagnostics path).
	in, ok := s.buildPluginFileInput(uri, nil)
	if !ok {
		t.Fatal("expected ok=true (file has a plugin rule)")
	}
	if in.Text == nil || *in.Text != "overlay buffer" {
		t.Errorf("nil override should use overlay, got %v", in.Text)
	}
	if in.ConfigKey != "file:///proj" {
		t.Errorf("configKey = %q, want file:///proj", in.ConfigKey)
	}
	if len(in.Rules) != 1 || in.Rules[0].Name != "tplsp/no-foo" {
		t.Errorf("expected only the plugin rule forwarded, got %+v", in.Rules)
	}

	// Explicit override must win over the stale overlay.
	override := "fixed pass content"
	in2, ok := s.buildPluginFileInput(uri, &override)
	if !ok {
		t.Fatal("expected ok=true with override")
	}
	if in2.Text == nil || *in2.Text != "fixed pass content" {
		t.Errorf("override should win over overlay, got %v", in2.Text)
	}
}

// TestLintPluginRulesSync_RebuildsWithFixes verifies the synchronous fixAll
// helper turns a mocked worker result into a RuleDiagnostic carrying the fix
// (so ApplyRuleFixes can apply it) with the configured severity reattached.
func TestLintPluginRulesSync_RebuildsWithFixes(t *testing.T) {
	config.RegisterEslintPluginRules([]config.EslintPluginEntry{
		{Prefix: "tplsync", RuleNames: []string{"no-bar"}},
	})
	s := newTestServer()
	s.jsConfigs["file:///proj"] = config.RslintConfig{
		{
			Plugins: []string{"tplsync"},
			Rules:   config.Rules{"tplsync/no-bar": "error"},
		},
	}
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

	diags := s.lintPluginRulesSync(context.Background(), uri, content, true, "off")
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

	// A file with no plugin rules → nil (caller proceeds native-only).
	other := lsproto.DocumentUri("file:///elsewhere/x.ts")
	if got := s.lintPluginRulesSync(context.Background(), other, "x", true, "off"); got != nil {
		t.Errorf("expected nil for a file with no plugin rules, got %v", got)
	}
}

// ======== computeFixAllContent: native+plugin fold loop ========

// TestComputeFixAllContent_FoldsPluginFixes drives the fixAll multi-pass loop
// with an injected native lint (no TS session) and a mocked plugin dispatcher,
// asserting that BOTH a native fix and a plugin fix apply in the same pass (the
// fold), the plugin fix is not clobbered by the native one, and the loop
// converges across passes on the in-progress content.
func TestComputeFixAllContent_FoldsPluginFixes(t *testing.T) {
	config.RegisterEslintPluginRules([]config.EslintPluginEntry{
		{Prefix: "tpfold", RuleNames: []string{"no-bar"}},
	})
	s := newTestServer()
	s.jsConfigs["file:///proj"] = config.RslintConfig{
		{
			Plugins: []string{"tpfold"},
			Rules:   config.Rules{"tpfold/no-bar": "error"},
		},
	}
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

	// Injected native lint: fix the first "1" → "2" wherever it appears. Returns
	// no fix once the digit is gone (so the loop converges).
	var nativePasses int
	s.fixAllNativeLint = func(_ context.Context, _ lsproto.DocumentUri, _ int, content string, _ config.RslintConfig, _ string, _ bool, _ []string) ([]rule.RuleDiagnostic, error) {
		nativePasses++
		idx := strings.Index(content, "1")
		if idx < 0 {
			return nil, nil
		}
		return []rule.RuleDiagnostic{{
			RuleName:   "native/prefer-2",
			Range:      core.NewTextRange(idx, idx+1),
			Message:    rule.RuleMessage{Description: "use 2"},
			SourceFile: textOnlySourceFile{text: content},
			FixesPtr:   &[]rule.RuleFix{{Text: "2", Range: core.NewTextRange(idx, idx+1)}},
		}}, nil
	}

	got := s.computeFixAllContent(context.Background(), uri, original, config.RslintConfig{}, "", true, nil)

	// Both fixes applied (native "1"→"2" AND plugin "bar"→"baz") proves the
	// fold: the plugin fix survived alongside the native one in the same pass.
	if got != "const baz = 2;" {
		t.Fatalf("computeFixAllContent = %q, want %q (native+plugin folded)", got, "const baz = 2;")
	}
	// Pass 0 fixes both; pass 1 sees neither "1" nor "bar" → no fix → converge.
	if nativePasses != 2 || pluginPasses != 2 {
		t.Errorf("expected 2 native + 2 plugin passes (1 fixing, 1 converging), got native=%d plugin=%d", nativePasses, pluginPasses)
	}
}

// TestComputeFixAllContent_PluginTimeoutFallsBackNativeOnly asserts the
// source.fixAll plugin reverse requests are bounded by a deadline: a wedged
// client that never answers must not freeze the dispatch loop. On timeout the
// plugin pass is dropped and native fixes still apply.
func TestComputeFixAllContent_PluginTimeoutFallsBackNativeOnly(t *testing.T) {
	config.RegisterEslintPluginRules([]config.EslintPluginEntry{
		{Prefix: "tptimeout", RuleNames: []string{"no-bar"}},
	})
	s := newTestServer()
	s.jsConfigs["file:///proj"] = config.RslintConfig{
		{
			Plugins: []string{"tptimeout"},
			Rules:   config.Rules{"tptimeout/no-bar": "error"},
		},
	}
	s.fixAllPluginTimeout = 20 * time.Millisecond
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
	// Injected native lint: fix the first "1" → "2", nothing once it is gone.
	s.fixAllNativeLint = func(_ context.Context, _ lsproto.DocumentUri, _ int, content string, _ config.RslintConfig, _ string, _ bool, _ []string) ([]rule.RuleDiagnostic, error) {
		idx := strings.Index(content, "1")
		if idx < 0 {
			return nil, nil
		}
		return []rule.RuleDiagnostic{{
			RuleName:   "native/prefer-2",
			Range:      core.NewTextRange(idx, idx+1),
			Message:    rule.RuleMessage{Description: "use 2"},
			SourceFile: textOnlySourceFile{text: content},
			FixesPtr:   &[]rule.RuleFix{{Text: "2", Range: core.NewTextRange(idx, idx+1)}},
		}}, nil
	}

	start := time.Now()
	got := s.computeFixAllContent(context.Background(), uri, original, config.RslintConfig{}, "", true, nil)
	elapsed := time.Since(start)

	// Native fix applied; the wedged plugin pass timed out and was dropped
	// ("bar" stays, only "1"→"2").
	if got != "const bar = 2;" {
		t.Fatalf("computeFixAllContent = %q, want %q (native-only after plugin timeout)", got, "const bar = 2;")
	}
	// The whole fixAll is bounded by the shared plugin budget — without the
	// deadline the wedged dispatch would hang the dispatch loop forever.
	if elapsed > 2*time.Second {
		t.Errorf("fixAll took %v; the plugin deadline should bound the stall", elapsed)
	}
	if dispatchCalls == 0 {
		t.Error("expected the plugin dispatcher to be invoked (then time out)")
	}
}

// TestLintPluginRulesSync_ExpiredCtxReturnsNil pins the mechanism that keeps the
// shared fixAll plugin deadline from multiplying across passes: once the budget
// has expired, a later pass's call returns nil promptly instead of blocking
// another full timeout.
func TestLintPluginRulesSync_ExpiredCtxReturnsNil(t *testing.T) {
	config.RegisterEslintPluginRules([]config.EslintPluginEntry{
		{Prefix: "tpexpired", RuleNames: []string{"no-bar"}},
	})
	s := newTestServer()
	s.jsConfigs["file:///proj"] = config.RslintConfig{
		{
			Plugins: []string{"tpexpired"},
			Rules:   config.Rules{"tpexpired/no-bar": "error"},
		},
	}
	uri := lsproto.DocumentUri("file:///proj/a.ts")
	content := "const bar = 1;"

	s.eslintPluginDispatch = func(ctx context.Context, _ linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already expired before the call

	start := time.Now()
	diags := s.lintPluginRulesSync(ctx, uri, content, true, "off")
	elapsed := time.Since(start)

	if diags != nil {
		t.Errorf("expired ctx → native-only (nil diagnostics), got %v", diags)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("expired ctx should return promptly, took %v", elapsed)
	}
}
