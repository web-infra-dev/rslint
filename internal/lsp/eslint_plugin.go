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
			// it into the typed result.
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
// directory's catalog identity: the absolute filesystem path discovered by Go.
// The worker routes tasks by matching it byte-for-byte.
//
// textOverride forces the text the worker lints: the diagnostics path passes
// nil (use the s.documents overlay), while the multi-pass fixAll path passes
// the in-progress fixed content of the current pass so plugin fix byte offsets
// stay aligned with that content.
//
// Must be called from the main dispatch loop: it reads s.jsConfigs (lock-free)
// and the s.documents overlay.
func (s *Server) buildPluginFileInput(uri lsproto.DocumentUri, textOverride *string) (linter.EslintPluginFileInput, bool) {
	if s.isUnavailableConfigForURI(uri) {
		return linter.EslintPluginFileInput{}, false
	}
	rslintConfig, configCwd, isJSConfig := s.getLintConfigForURI(uri)
	return s.buildPluginFileInputWithConfig(uri, textOverride, rslintConfig, configCwd, isJSConfig)
}

func (s *Server) buildPluginFileInputWithConfig(
	uri lsproto.DocumentUri,
	textOverride *string,
	rslintConfig config.RslintConfig,
	configCwd string,
	isJSConfig bool,
) (linter.EslintPluginFileInput, bool) {
	configKey := s.pluginConfigKeyForURI(uri)
	filePath := uriToPath(uri)
	configFilePath, matchConfigDir := config.ResolveConfigPathSpace(filePath, configCwd, s.fs)
	if isDefaultExcludedLintPath(configFilePath, matchConfigDir, s.fs) {
		return linter.EslintPluginFileInput{}, false
	}

	fileConfigResolver := config.NewFileConfigResolver(rslintConfig, matchConfigDir, isJSConfig)
	enabledRules, merged := fileConfigResolver.EnabledRulesForFile(configFilePath)
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
	enabledRules = s.pluginRulesForCurrentGeneration(enabledRules)
	languageOptions, settings := config.PluginMergedMaps(merged)
	return linter.BuildEslintPluginFileInput(filePath, configKey, enabledRules, languageOptions, settings, text, nil)
}

// eslintPluginRuleSet expands activation metadata into the exact rule names
// the matching Node generation can execute. It always returns a non-nil map:
// an activated generation with no community plugins must block placeholders
// retained in the process-wide registry by older generations.
func eslintPluginRuleSet(entries []config.EslintPluginEntry) map[string]struct{} {
	rules := make(map[string]struct{})
	for _, entry := range entries {
		if entry.Prefix == "" {
			continue
		}
		for _, ruleName := range entry.RuleNames {
			if ruleName != "" {
				rules[entry.Prefix+"/"+ruleName] = struct{}{}
			}
		}
	}
	return rules
}

func (s *Server) pluginRulesForCurrentGeneration(rules []linter.ConfiguredRule) []linter.ConfiguredRule {
	if s.eslintPluginRules == nil {
		return rules
	}
	filtered := make([]linter.ConfiguredRule, 0, len(rules))
	for _, configuredRule := range rules {
		if !configuredRule.IsEslintPluginRule {
			filtered = append(filtered, configuredRule)
			continue
		}
		if _, ok := s.eslintPluginRules[configuredRule.Name]; ok {
			filtered = append(filtered, configuredRule)
		}
	}
	return filtered
}

// pluginConfigKeyForURI returns the owning config directory's absolute path.
// It uses the same resolver as getConfigForURI and preserves the catalog key
// byte-for-byte for the Node worker.
//
// For the JSON-config fallback there is no JS config directory, so the key is
// empty — the worker has no plugins registered for that path anyway (JSON
// configs cannot mount object-form plugins), and a file with no plugin rules never
// reaches dispatch.
func (s *Server) pluginConfigKeyForURI(uri lsproto.DocumentUri) string {
	if configKey, ok := s.nearestJSConfigKey(uri); ok {
		return configKey
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
func (s *Server) dispatchPluginLint(uri lsproto.DocumentUri, generation uint64) {
	if s.isUnavailableConfigForURI(uri) {
		s.cancelInflightPluginDispatch(uri)
		return
	}
	rslintConfig, configCwd, isJSConfig := s.getLintConfigForURI(uri)
	s.dispatchPluginLintWithConfig(uri, generation, rslintConfig, configCwd, isJSConfig)
}

func (s *Server) dispatchPluginLintWithConfig(
	uri lsproto.DocumentUri,
	generation uint64,
	rslintConfig config.RslintConfig,
	configCwd string,
	isJSConfig bool,
) {
	// Supersede any prior in-flight dispatch for this URI FIRST — before the
	// no-plugin-work early return below. Even a relint that yields no plugin
	// rules (the file became globally ignored, or its plugin-rule set dropped to
	// empty after a config refresh) must still cancel the prior dispatch so its Node
	// worker stops instead of running to completion. Go-side frees the goroutine;
	// a $/cancelRequest tells the client to stop the worker.
	s.cancelInflightPluginDispatch(uri)

	input, ok := s.buildPluginFileInputWithConfig(uri, nil, rslintConfig, configCwd, isJSConfig)
	if !ok {
		return
	}
	dispatch := s.pluginDispatchForGeneration(s.eslintPluginConfigGeneration)

	// Bound the reverse request as a backstop: even with supersede-cancel, a
	// client that neither answers nor is ever superseded (the user stops typing)
	// would otherwise leak this goroutine + its pendingServerRequests entry.
	timeout := s.pluginReverseTimeout
	if timeout <= 0 {
		timeout = defaultPluginReverseTimeout
	}
	ctx, cancel := context.WithTimeout(s.backgroundCtx, timeout)

	// Register so a later supersede or close can cancel the request. sendRequest
	// forwards that context cancellation to the client.
	handle := &pluginDispatchHandle{cancel: cancel, done: make(chan struct{})}
	s.inflightPluginDispatchMu.Lock()
	s.inflightPluginDispatch[uri] = handle
	s.inflightPluginDispatchMu.Unlock()

	go func() {
		defer close(handle.done)
		defer cancel()
		defer s.clearInflightPluginDispatch(uri, handle)
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
			nil,                         // timing — the LSP never collects it
			func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		)
		// Categorize like the fixAll sibling (lintPluginRulesSync): a superseded
		// batch (context.Canceled) is silent; a client that never answered within
		// pluginReverseTimeout (context.DeadlineExceeded) is benign and expected —
		// logging it at error severity would spam every debounced relint — so it
		// gets an info-level note, not an error. Only a genuine failure is an
		// error. Generation already guards staleness, so a non-delivered result
		// just leaves the pass native-only.
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
			case errors.Is(err, context.DeadlineExceeded):
				log.Printf("[rslint] eslint-plugin lint for %s timed out (client unresponsive); leaving it native-only", uri)
			default:
				log.Printf("[rslint] eslint-plugin lint error for %s: %v", uri, err)
			}
			return
		}
		// Deliver the freshly-computed result. Prefer the buffered send so a valid
		// result is never raced away by a deadline that expired in the gap between
		// the worker returning and this select; fall back to the ctx.Done() drop
		// only if the buffer is genuinely full (dispatch loop not draining).
		result := pluginLintResult{uri: uri, generation: generation, diags: diags}
		select {
		case s.pluginResultCh <- result:
		default:
			select {
			case s.pluginResultCh <- result:
			case <-ctx.Done():
			}
		}
	}()
}

// cancelInflightPluginDispatch cancels and $/cancelRequests the in-flight
// background plugin dispatch for uri, if any. Called when a newer keystroke
// supersedes it (dispatchPluginLint) or the document closes (handleDidClose).
func (s *Server) cancelInflightPluginDispatch(uri lsproto.DocumentUri) {
	s.inflightPluginDispatchMu.Lock()
	handle, ok := s.inflightPluginDispatch[uri]
	if ok {
		delete(s.inflightPluginDispatch, uri)
	}
	s.inflightPluginDispatchMu.Unlock()
	if !ok {
		return
	}
	handle.cancel()
	if handle.done != nil {
		<-handle.done
	}
}

// clearInflightPluginDispatch removes handle from the registry once its
// goroutine finishes, but only if a later dispatch has not already replaced it.
func (s *Server) clearInflightPluginDispatch(uri lsproto.DocumentUri, handle *pluginDispatchHandle) {
	s.inflightPluginDispatchMu.Lock()
	if s.inflightPluginDispatch[uri] == handle {
		delete(s.inflightPluginDispatch, uri)
	}
	s.inflightPluginDispatchMu.Unlock()
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
// (pluginReverseTimeout) so a wedged or mid-rebuild client that never answers
// cannot stall the dispatch loop: on expiry DispatchEslintPluginRules returns a
// context error and this returns nil, leaving the pass native-only.
//
// Must be called from the main dispatch loop (it reads jsConfigs + documents).
func (s *Server) lintPluginRulesSync(ctx context.Context, uri lsproto.DocumentUri, content string, fix bool, suggestionsMode string) []rule.RuleDiagnostic {
	if s.isUnavailableConfigForURI(uri) {
		return nil
	}
	rslintConfig, configCwd, isJSConfig := s.getLintConfigForURI(uri)
	return s.lintPluginRulesSyncWithConfig(ctx, uri, content, fix, suggestionsMode, rslintConfig, configCwd, isJSConfig)
}

func (s *Server) lintPluginRulesSyncWithConfig(
	ctx context.Context,
	uri lsproto.DocumentUri,
	content string,
	fix bool,
	suggestionsMode string,
	rslintConfig config.RslintConfig,
	configCwd string,
	isJSConfig bool,
) []rule.RuleDiagnostic {
	input, ok := s.buildPluginFileInputWithConfig(uri, &content, rslintConfig, configCwd, isJSConfig)
	if !ok {
		return nil
	}
	dispatch := s.pluginDispatchForGeneration(s.eslintPluginConfigGeneration)

	var diags []rule.RuleDiagnostic
	err := linter.DispatchEslintPluginRules(
		ctx,
		dispatch,
		[]linter.EslintPluginFileInput{input},
		fix,
		suggestionsMode,
		nil, // timing — the LSP never collects it
		func(d rule.RuleDiagnostic) { diags = append(diags, d) },
	)
	if err != nil {
		// context.Canceled means the editor aborted the fixAll request;
		// context.DeadlineExceeded means the pluginReverseTimeout budget elapsed
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

func (s *Server) pluginDispatchForGeneration(generation string) linter.EslintPluginDispatcher {
	dispatch := s.installEslintPluginDispatch()
	return func(ctx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		req.Generation = generation
		return dispatch(ctx, req)
	}
}
