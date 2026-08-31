package lsp

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

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

// deriveLSPRuleCatalog preserves the LSP's established generation gate: an
// empty object-form plugin rule name is not resolvable in Go, even though the
// CLI and API historically accepted the same metadata.
func deriveLSPRuleCatalog(
	base *rule.Catalog,
	plugins []config.EslintPluginEntry,
) (*rule.Catalog, []string) {
	filtered := make([]config.EslintPluginEntry, 0, len(plugins))
	for _, plugin := range plugins {
		ruleNames := make([]string, 0, len(plugin.RuleNames))
		for _, ruleName := range plugin.RuleNames {
			if ruleName != "" {
				ruleNames = append(ruleNames, ruleName)
			}
		}
		filtered = append(filtered, config.EslintPluginEntry{
			Prefix:    plugin.Prefix,
			RuleNames: ruleNames,
		})
	}
	return base.ForESLintPlugins(filtered)
}

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

func (p *documentProgressiveDiagnostics) Submit(
	parentCtx context.Context,
	run linter.DeferredPluginRun,
) {
	p.server.startDiagnosticEnrichment(
		parentCtx,
		p.uri,
		p.generation,
		p.pluginGeneration,
		run,
	)
}

// startDiagnosticEnrichment implements the progressive request's asynchronous
// executor port. RunPipeline alone decides when to submit work; this adapter
// supplies transport, deadline, cancellation, and stale-result admission.
func (s *Server) startDiagnosticEnrichment(
	parentCtx context.Context,
	uri lsproto.DocumentUri,
	generation uint64,
	pluginGeneration string,
	run linter.DeferredPluginRun,
) {
	if run == nil {
		return
	}
	dispatch := s.pluginDispatchForGeneration(pluginGeneration)

	// Bound the reverse request as a backstop: even with supersede-cancel, a
	// client that neither answers nor is ever superseded (the user stops typing)
	// would otherwise leak this goroutine + its pendingServerRequests entry.
	timeout := s.pluginReverseTimeout
	if timeout <= 0 {
		timeout = defaultPluginReverseTimeout
	}
	ctx, cancel := context.WithTimeout(parentCtx, timeout)

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
		outcome, runErr := run(ctx, dispatch)
		reportLSPPluginProtocolNotices(outcome.Notices)
		err := errors.Join(runErr, outcome.DispatchError)
		// Categorize like the fixAll sibling: a superseded
		// batch (context.Canceled) is silent; a client that never answered within
		// pluginReverseTimeout (context.DeadlineExceeded) is benign and expected —
		// it gets an info-level note, not an error. Only a genuine failure is an
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
		result := pluginLintResult{uri: uri, generation: generation, diags: outcome.Diagnostics}
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

func reportLSPPluginProtocolNotices(notices []linter.EslintPluginProtocolNotice) {
	writeLSPPluginProtocolNotices(os.Stderr, notices)
}

func writeLSPPluginProtocolNotices(w io.Writer, notices []linter.EslintPluginProtocolNotice) {
	for _, notice := range notices {
		switch notice.Kind {
		case linter.EslintPluginMissingFileResult:
			fmt.Fprintf(w, "rslint: plugin-lint returned no result for %q\n", notice.FilePath)
		case linter.EslintPluginUnconfiguredDiagnostic:
			fmt.Fprintf(w, "rslint: plugin diagnostic for unconfigured rule %q in %q\n", notice.RuleName, notice.FilePath)
		}
	}
}

// cancelInflightPluginDispatch cancels and $/cancelRequests the in-flight
// background plugin dispatch for uri, if any. Called when a newer prepared
// document generation supersedes it or the document closes (handleDidClose).
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

func (s *Server) pluginDispatchForGeneration(generation string) linter.EslintPluginDispatcher {
	dispatch := s.installEslintPluginDispatch()
	return func(ctx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		req.Generation = generation
		return dispatch(ctx, req)
	}
}

// pluginDispatchWithinBudget adapts the LSP reverse transport to one
// operation-wide budget. Each call is also canceled when its individual lint
// observation ends, while expiration prevents later observations from sending
// another request.
func (s *Server) pluginDispatchWithinBudget(
	parent context.Context,
	generation string,
) (linter.EslintPluginDispatcher, context.CancelFunc) {
	timeout := s.pluginReverseTimeout
	if timeout <= 0 {
		timeout = defaultPluginReverseTimeout
	}
	budgetCtx, cancelBudget := context.WithTimeout(parent, timeout)
	dispatch := s.pluginDispatchForGeneration(generation)
	return func(runCtx context.Context, request linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		if err := budgetCtx.Err(); err != nil {
			return nil, err
		}
		callCtx, cancelCall := context.WithCancel(budgetCtx)
		stop := context.AfterFunc(runCtx, cancelCall)
		defer func() {
			stop()
			cancelCall()
		}()
		return dispatch(callCtx, request)
	}, cancelBudget
}
