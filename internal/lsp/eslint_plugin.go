package lsp

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// methodPluginLint is the server→client reverse request that asks the
// VS Code extension to run a batch of ESLint-plugin rules in its worker pool
// and return the diagnostics. It is the LSP equivalent of the CLI's
// `pluginLint` IPC request.
const methodPluginLint = lsproto.Method("rslint/pluginLint")

// installEslintPluginDispatch lazily builds the dispatcher closure once. It
// sends one plugin-lint batch over the reverse request and decodes the
// result. Reused across all files/keystrokes; only touches sendRequest
// (goroutine-safe), so the closure itself may run off the dispatch loop.
//
// Called from the main dispatch loop (pushDiagnostics) before spawning the
// plugin goroutine, so the lazy assignment never races.
func (s *Server) installEslintPluginDispatch() linter.EslintPluginDispatcher {
	if s.eslintPluginDispatch == nil {
		s.eslintPluginDispatch = func(ctx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
			raw, err := s.sendRequest(ctx, methodPluginLint, req)
			if err != nil {
				return nil, err
			}
			// raw is already-decoded JSON (map/slice); re-marshal then decode
			// into the typed result, mirroring handleConfigUpdate's handling
			// of an `any` params payload (service.go).
			data, err := stdjson.Marshal(raw)
			if err != nil {
				return nil, fmt.Errorf("marshal pluginLint result: %w", err)
			}
			var res linter.EslintPluginLintResult
			if err := stdjson.Unmarshal(data, &res); err != nil {
				return nil, fmt.Errorf("decode pluginLint result: %w", err)
			}
			return &res, nil
		}
	}
	return s.eslintPluginDispatch
}

// buildPluginFileInput assembles the single-file eslint-plugin dispatch input
// for uri, or returns ok=false when the file has no plugin rules (so the
// caller skips the reverse request entirely).
//
// The plugin rules are the IsEslintPluginRule subset of the file's enabled
// rules — exactly the rules the native pass treats as no-op placeholders.
// Per-file languageOptions / settings come from GetConfigForFile (the same
// merged config the native pass resolves). configKey is the owning config
// directory in URI form: that is the string the VS Code extension registered
// its worker-pool LoadedPlugins under (Uri.file(dir).toString()), and the
// worker routes tasks to plugins by matching it byte-for-byte.
//
// textOverride forces the text the worker lints: the diagnostics path passes
// nil (use the s.documents overlay), while the multi-pass fixAll path passes
// the in-progress fixed content of the current pass so plugin fix byte offsets
// stay aligned with that content.
//
// Must be called from the main dispatch loop: it reads s.jsConfigs (lock-free)
// and the s.documents overlay.
func (s *Server) buildPluginFileInput(uri lsproto.DocumentUri, textOverride *string) (linter.EslintPluginFileInput, bool) {
	rslintConfig, configCwd, isJSConfig := s.getConfigForURI(uri)
	configKey := s.pluginConfigKeyForURI(uri)
	filePath := uriToPath(uri)

	enabledRules, merged := config.GlobalRuleRegistry.GetEnabledRules(rslintConfig, filePath, configCwd, isJSConfig)
	if merged == nil {
		// File is globally ignored — no plugin (or native) diagnostics.
		return linter.EslintPluginFileInput{}, false
	}

	// Text is the content the worker lints. An explicit override (fixAll's
	// in-progress fixed content) wins; otherwise use the editor overlay
	// (unsaved buffer) so the worker lints the in-memory content, not the
	// stale on-disk copy. Fall back to nil (worker reads disk) only if we have
	// neither.
	text := textOverride
	if text == nil {
		if content, ok := s.documents[uri]; ok {
			text = &content
		}
	}

	// sourceFile=nil: the LSP rebuilds against the overlay Text (the worker
	// linted that same string). Shared filter/assembly with the CLI (F1).
	languageOptions, settings := config.PluginMergedMaps(merged)
	return linter.BuildEslintPluginFileInput(filePath, configKey, enabledRules, languageOptions, settings, text, nil)
}

// pluginConfigKeyForURI returns the owning config directory in URI form
// (e.g. "file:///project") — the configKey the Node worker pool routes on.
// It walks parents exactly like getConfigForURI but yields the URI key rather
// than the filesystem path getConfigForURI returns for cwd use.
//
// For the JSON-config fallback there is no JS config directory, so the key is
// empty — the worker has no plugins registered for that path anyway (JSON
// configs cannot mount object-form plugins), and a file with no plugin rules never
// reaches dispatch.
func (s *Server) pluginConfigKeyForURI(uri lsproto.DocumentUri) string {
	if len(s.jsConfigs) > 0 {
		dir := uriDirname(string(uri))
		for {
			if _, ok := s.jsConfigs[dir]; ok {
				return dir
			}
			parent := uriDirname(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return ""
}

// dispatchPluginLint runs the eslint-plugin lint for uri in a goroutine and
// delivers the rebuilt diagnostics back to the main dispatch loop via
// pluginResultCh, tagged with generation. It is the concurrent companion to
// pushDiagnostics's synchronous native pass.
//
// Concurrency (R1): the goroutine touches ONLY the dispatcher (sendRequest is
// goroutine-safe) and a local diagnostics slice. It NEVER writes s.diagnostics
// — that lock-free map is merged solely on the main loop when it consumes
// pluginResultCh. The generation stamp lets the main loop drop results that a
// newer keystroke has superseded.
//
// Must be called from the main dispatch loop (it reads jsConfigs + documents
// to build the input and the generation map).
//
// Follow-up (not needed for correctness): when a newer keystroke supersedes an
// in-flight dispatch we currently rely on generation to drop the stale result,
// but the Node worker keeps running until it finishes. Sending the client a
// `$/cancelRequest` for the superseded reverse-request id (s.sendRequest mints
// "ts%d" ids) would free that worker CPU sooner. It only saves work, never
// changes the published diagnostics, so it is deferred.
func (s *Server) dispatchPluginLint(uri lsproto.DocumentUri, generation uint64) {
	input, ok := s.buildPluginFileInput(uri, nil)
	if !ok {
		return
	}
	dispatch := s.installEslintPluginDispatch()
	ctx := s.backgroundCtx

	go func() {
		// onDiagnostic is invoked serially (DispatchEslintPluginRules emits
		// diagnostics single-threaded after its batches complete; here there is
		// only ever one batch), so the local slice needs no lock.
		var diags []rule.RuleDiagnostic
		err := linter.DispatchEslintPluginRules(
			ctx,
			dispatch,
			[]linter.EslintPluginFileInput{input},
			// Collect fixes + materialize suggestions so the stored plugin
			// diagnostics carry them for handleCodeAction's quickfix /
			// suggestion actions (the editor reads fixes off the already-
			// published diagnostics; it does not re-lint). The fixes are
			// collected, never applied here. Cost: the worker runs each plugin
			// rule's fix(fixer) per keystroke — small vs the lint itself.
			true,                        // fix → collectFixes
			linter.SuggestionsModeEager, // suggestionsMode
			func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		)
		// context.Canceled means the worker cooperatively dropped the batch
		// (newer request superseded it); generation already guards staleness,
		// so just don't deliver. Other errors are logged; the editor never
		// sees worker stderr, but a per-file crash is surfaced as a diagnostic
		// inside DispatchEslintPluginRules itself.
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[rslint] eslint-plugin lint error for %s: %v", uri, err)
			}
			return
		}
		select {
		case s.pluginResultCh <- pluginLintResult{uri: uri, generation: generation, diags: diags}:
		case <-ctx.Done():
		}
	}()
}

// mergePluginDiagnostics merges a plugin lint result into s.diagnostics and
// re-publishes. Runs ONLY on the main dispatch loop (it writes the lock-free
// s.diagnostics map). Stale results (a newer relint bumped the generation, or
// the document was closed) are dropped.
func (s *Server) mergePluginDiagnostics(r pluginLintResult) {
	if s.docGeneration[r.uri] != r.generation {
		return // superseded by a newer lint, or doc closed (generation cleared)
	}
	if _, open := s.documents[r.uri]; !open {
		return // document closed between dispatch and result
	}

	// Append plugin diagnostics to the native ones already stored for this
	// generation. handleCodeAction reads s.diagnostics[uri], so plugin
	// quick fixes / suggestions become available too.
	merged := append(s.diagnostics[r.uri], r.diags...)
	s.diagnostics[r.uri] = merged

	lspDiags := make([]*lsproto.Diagnostic, 0, len(merged))
	for _, d := range merged {
		lspDiags = append(lspDiags, convertRuleDiagnosticToLSP(d))
	}
	if err := s.PublishDiagnostics(s.backgroundCtx, &lsproto.PublishDiagnosticsParams{
		Uri:         r.uri,
		Diagnostics: lspDiags,
	}); err != nil {
		log.Printf("[rslint] Error publishing plugin diagnostics: %v", err)
	}
}

// lintPluginRulesSync runs uri's eslint-plugin rules synchronously against the
// given content and returns the rebuilt diagnostics (fixes collected when
// fix=true). It is the blocking companion to dispatchPluginLint, used by
// handleFixAllCodeAction to fold plugin fixes into each native fix pass.
//
// Blocking is safe even though handleCodeAction runs on the dispatch loop: the
// reverse-request response is routed by readLoop (server.go pendingServer-
// Requests), never by the dispatch loop, so awaiting our own request cannot
// deadlock — the same reason the native fixAll pass may block on the language
// service. onDiagnostic is invoked serially (DispatchEslintPluginRules emits
// diagnostics single-threaded after its batches complete; this path has only
// one batch), so the local slice needs no lock. Returns nil when the
// file has no plugin rules.
//
// The caller (computeFixAllContent) passes a ctx already bounded by a deadline
// (fixAllPluginTimeout) so a wedged or mid-rebuild client that never answers
// cannot stall the dispatch loop: on expiry DispatchEslintPluginRules returns a
// context error and this returns nil, leaving the pass native-only.
//
// Must be called from the main dispatch loop (it reads jsConfigs + documents).
func (s *Server) lintPluginRulesSync(ctx context.Context, uri lsproto.DocumentUri, content string, fix bool, suggestionsMode string) []rule.RuleDiagnostic {
	input, ok := s.buildPluginFileInput(uri, &content)
	if !ok {
		return nil
	}
	dispatch := s.installEslintPluginDispatch()

	var diags []rule.RuleDiagnostic
	err := linter.DispatchEslintPluginRules(
		ctx,
		dispatch,
		[]linter.EslintPluginFileInput{input},
		fix,
		suggestionsMode,
		func(d rule.RuleDiagnostic) { diags = append(diags, d) },
	)
	if err != nil {
		// context.Canceled means the editor aborted the fixAll request;
		// context.DeadlineExceeded means the fixAllPluginTimeout budget elapsed
		// (an unresponsive client) — both leave this pass native-only. Other
		// errors (worker crash, etc.) are logged but likewise leave the pass
		// native-only rather than failing the whole fixAll; a per-file plugin
		// crash is already surfaced on the diagnostics path.
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			log.Printf("[rslint] eslint-plugin fixAll for %s timed out (client unresponsive); applying native-only fixes", uri)
		case errors.Is(err, context.Canceled):
		default:
			log.Printf("[rslint] eslint-plugin fixAll lint error for %s: %v", uri, err)
		}
		return nil
	}
	return diags
}
