// Package lsp / rslint custom LSP protocol extensions.
//
// rslint extends the standard Language Server Protocol with custom methods
// (all under the `rslint/` namespace) to coordinate the ESLint-plugin
// compatibility layer between the Go language server and a Node-hosted
// LSP client (typically the VS Code extension).
//
// Two custom methods exist:
//
//   - `rslint/configUpdate` (client → server, notification-shaped request)
//     The client side owns JS/TS rslint config loading because the Go
//     server can't import JS/TS modules. After the client loads
//     `rslint.config.{js,ts,mjs,mts}` it normalizes the entries and pushes
//     them here. The Go server uses the entries for rule-enabling /
//     ignore decisions; it does NOT execute ESLint plugins itself.
//     This method predates lintCompatBatch and is kept inline-registered
//     in server.go for now.
//
//   - `rslint/lintCompatBatch` (server → client, request)
//     Defined in this file. Used for every batch of ESLint-plugin rule
//     execution. See the doc on `MethodRslintLintCompatBatch` below.
//
// Both methods are rslint-specific and not part of LSP itself. They use
// the same JSON-RPC framing as standard LSP methods, so any LSP client
// that exposes a "raw" `onRequest(method, handler)` and `onNotification`
// API (vscode-languageclient does) can implement them.
//
// Cancellation flows through the standard `$/cancelRequest`. When the Go
// server's per-keystroke lint context is cancelled (LSP supersession,
// editor close, etc.) the dispatcher sends `$/cancelRequest` to the
// client; the client's vscode-jsonrpc layer marks the handler's
// `CancellationToken` cancelled and the handler's WorkerPool aborts
// in-flight tasks.

package lsp

import (
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"

	"github.com/web-infra-dev/rslint/internal/linter"
)

// MethodRslintLintCompatBatch is the LSP method name for the
// server-to-client request that delegates one batch of ESLint-plugin
// rule execution to the LSP client.
//
// Direction: server → client (this is unusual — most LSP requests are
// client → server — but the standard supports both via JSON-RPC).
//
// When invoked:
//
//	The Go server has finished native-rule execution for a file (or
//	cross-file Program) and has zero or more `IsEslintPluginRule` rules
//	left to run. It groups them per signature into a CompatBatch and
//	sends a `rslint/lintCompatBatch` request. The client (the VS Code
//	extension or any other Node-hosted LSP client) handles the request
//	by running the rules in its own WorkerPool — which is the only
//	process with live ESLint plugin instances — and replies with the
//	resulting diagnostics. The Go server merges the response into the
//	native-rule diagnostic stream and publishes via the usual
//	`textDocument/publishDiagnostics` channel.
//
// Why server-to-client:
//
//	ESLint plugins are JavaScript. The Go LSP server cannot host them.
//	The LSP client (extension host) already loads the user's
//	rslint.config and therefore already holds live plugin instances in
//	its Node process. Routing plugin execution back to the client
//	eliminates a third "sidecar" process and ~1000 lines of IPC glue.
//
// Cardinality:
//
//	One request per Program-level signature bucket — same as the
//	`CompatBatchHandler` contract (see linter.types.go). A typical
//	single-config lint produces one request per program; monorepos with
//	`overrides`-driven per-file rule variation produce more.
//
// Cancellation:
//
//	Standard `$/cancelRequest`. The client's vscode-jsonrpc layer
//	cancels the per-request CancellationToken; the handler then aborts
//	WorkerPool tasks via the per-task SharedArrayBuffer flag (same
//	mechanism the in-process CLI path already uses).
//
// Error response semantics:
//
//	A JSON-RPC error reply (`{ error: { code, message } }`) means the
//	whole batch failed — the linter treats the batch's files as
//	"compat-skipped" and surfaces a stderr log. Individual rule errors
//	are NOT carried as JSON-RPC errors; they appear inside
//	`CompatFileResult.RuleErrors` so one bad rule doesn't fail an
//	entire batch.
const MethodRslintLintCompatBatch lsproto.Method = "rslint/lintCompatBatch"

// LintCompatBatchParams is the wire-form of a single batch request.
// Identical in shape to linter.CompatBatch — we use the same type so
// there's only one place to evolve fields and one set of JSON tags. The
// alias keeps the LSP-protocol-facing name distinct from the
// linter-internal name for grep-ability.
type LintCompatBatchParams = linter.CompatBatch

// LintCompatBatchResult is the wire-form of one batch's results.
//
// Wrapping `Results` in a struct (rather than returning a bare array)
// is a forward-compatibility choice: future protocol versions may add
// sibling fields (statistics, partial-results flag, etc.) without
// breaking parsers. The Go-side LSP infrastructure unmarshals
// responses as `any` for unknown methods — the wrapping struct gives
// us a stable named type we can decode into.
type LintCompatBatchResult struct {
	Results []linter.CompatFileResult `json:"results"`
}
