// Package lsp / LSP-based eslint-plugin compat dispatcher.
//
// This file is the Go end of the `rslint/lintCompatBatch` LSP custom
// request defined in protocol.go. It produces a `linter.CompatBatchHandler`
// closure that the linter uses indistinguishably from the (now-deprecated)
// sidecar-process-backed implementation.
//
// Lifecycle: install once at server construction. The dispatcher itself
// is stateless — every call sends a fresh LSP request and decodes the
// response. The client (the VS Code extension) owns the WorkerPool state.
//
// Cancellation: relies on `sendRequestWithClientCancel`. When the per-file
// lint ctx is cancelled mid-flight (debounce supersession, file close,
// editor shutdown), the Go side sends `$/cancelRequest` to the client;
// the client's vscode-jsonrpc layer cancels the handler's
// CancellationToken; the handler's WorkerPool sets the per-task SAB flag;
// in-flight workers bail at the next per-node visit. The Go ctx then
// returns ctx.Err() through this dispatcher, which the linter surfaces
// as a per-batch failure (files in the batch are marked compat-skipped).
//
// Error handling: a JSON-RPC error response (`{ error: { code, message } }`)
// from the client means the batch failed catastrophically — return an
// error. Per-rule and per-file failures travel inside the result payload
// (`CompatFileResult.ParseError` / `RuleErrors`) and are NOT modeled as
// JSON-RPC errors, so one bad rule never fails a whole batch.

package lsp

import (
	"context"
	stdjson "encoding/json"
	"fmt"

	"github.com/web-infra-dev/rslint/internal/linter"
)

// newLintCompatLSPDispatcher returns a CompatBatchHandler closure that
// sends every batch as a `rslint/lintCompatBatch` LSP request to the
// connected client and decodes the response.
//
// The closure captures `s` (the server). It's safe to hold for the
// lifetime of the Server — there is no per-request mutable state, and
// `Server.sendRequestWithClientCancel` is goroutine-safe under
// `pendingServerRequestsMu`.
//
// Calling the closure when no client is connected (i.e. before LSP
// `initialize` completed, or after shutdown) will block on
// `outgoingQueue` until ctx cancels — but the linter never runs without
// a live client connection, so this isn't a practical concern.
//
// Per-batch deadline: removed. Two cheaper-and-more-precise mechanisms
// already cover every reachable failure mode:
//
//   - Worker hang / plugin infinite loop → Node-side WorkerPool's
//     `taskTimeoutMs` (default 30s) terminates the worker and returns
//     a `task_timeout`-stamped CompatFileResult per file.
//   - Client shutdown / editor restart → `ctx` is cancelled, which
//     `sendRequestWithClientCancel` translates to `$/cancelRequest`.
//   - Debounce supersession → linter cancels the in-flight ctx the
//     same way before issuing the next lint.
//
// The only case the previous Go-side `context.WithTimeout` caught
// that nothing else does is "Node host process alive but its event
// loop wedged" — extremely rare given the worker-pool / IPC paths
// have been audited for graceful shutdown. Removing this layer avoids
// the user-visible footgun of legitimately-large monorepo batches
// being killed at an arbitrary 30s ceiling.
func newLintCompatLSPDispatcher(s *Server) linter.CompatBatchHandler {
	return func(ctx context.Context, batch linter.CompatBatch) ([]linter.CompatFileResult, error) {
		// Empty batch fast path. The linter caller already filters down
		// to programs with at least one compat rule, but a config that
		// disables every plugin rule for a particular file's
		// `overrides` block can produce an empty `batch.Files` here. No
		// LSP round-trip needed.
		if len(batch.Files) == 0 {
			return nil, nil
		}

		// Send the request. sendRequestWithClientCancel handles
		// per-request bookkeeping AND fires `$/cancelRequest` if ctx
		// is cancelled before the response arrives.
		raw, err := s.sendRequestWithClientCancel(ctx, MethodRslintLintCompatBatch, batch)
		if err != nil {
			return nil, fmt.Errorf("rslint/lintCompatBatch: %w", err)
		}

		// `unmarshalResult` in the lsproto shim falls through to
		// `unmarshalAny` for unknown methods — `raw` here is whatever
		// `json.Value` (≈ []byte) the response carried, OR a
		// `map[string]interface{}` if the shim already decoded it. We
		// take a robust path: re-marshal to bytes, then unmarshal into
		// our typed struct. The double-hop is ~free (KB-scale payload)
		// and removes the type-discrimination burden from this code.
		blob, marshalErr := stdjson.Marshal(raw)
		if marshalErr != nil {
			return nil, fmt.Errorf("rslint/lintCompatBatch: re-marshal response: %w", marshalErr)
		}
		var decoded LintCompatBatchResult
		if err := stdjson.Unmarshal(blob, &decoded); err != nil {
			return nil, fmt.Errorf("rslint/lintCompatBatch: decode response: %w (raw: %s)", err, truncateForLog(blob))
		}

		// Validate the batch response: exactly one result per input file,
		// every result maps to a known input path (byte-equal then
		// normalize fallback, rewriting paths in place), and no duplicates.
		// Shared with the CLI runner — see
		// linter.ValidateAndNormalizeCompatResults (#17). A buggy client
		// that drops or duplicates results would otherwise look like "rule
		// simply found no issues" — indistinguishable from success and
		// very hard to debug after the fact.
		if err := linter.ValidateAndNormalizeCompatResults(decoded.Results, batch.Files); err != nil {
			return nil, fmt.Errorf("rslint/lintCompatBatch: %w", err)
		}
		return decoded.Results, nil
	}
}

// truncateForLog clips response payloads so an error message stays
// loggable when the response is huge (a batch over a 10k-file
// monorepo has results in the megabytes). 256 bytes is enough to spot
// shape issues — "expected object, got array of strings" — without
// flooding logs.
func truncateForLog(b []byte) string {
	const maxLogBytes = 256
	if len(b) <= maxLogBytes {
		return string(b)
	}
	return string(b[:maxLogBytes]) + "...(truncated)"
}
