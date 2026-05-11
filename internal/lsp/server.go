// modified based on https://github.com/microsoft/typescript-go/blob/cedc0cbe6c188f9bfe6a51af00c79be48c9ab74d/internal/lsp/server.go#L1
package lsp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/microsoft/typescript-go/shim/collections"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/diagnostics"
	"github.com/microsoft/typescript-go/shim/jsonrpc"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"golang.org/x/sync/errgroup"
)

type ServerOptions struct {
	In  Reader
	Out Writer
	Err io.Writer

	Cwd                string
	FS                 vfs.FS
	DefaultLibraryPath string
	TypingsLocation    string

	ParseCache *project.ParseCache
}

func NewServer(opts *ServerOptions) *Server {
	if opts.Cwd == "" {
		panic("Cwd is required")
	}
	s := &Server{
		r:                     opts.In,
		w:                     opts.Out,
		stderr:                opts.Err,
		requestQueue:          make(chan *lsproto.RequestMessage, 100),
		outgoingQueue:         make(chan *lsproto.Message, 100),
		pendingClientRequests: make(map[jsonrpc.ID]pendingClientRequest),
		pendingServerRequests: make(map[jsonrpc.ID]chan *lsproto.ResponseMessage),
		cwd:                   opts.Cwd,
		fs:                    opts.FS,
		defaultLibraryPath:    opts.DefaultLibraryPath,
		typingsLocation:       opts.TypingsLocation,
		parseCache:            opts.ParseCache,
		jsConfigs:             make(map[string]config.RslintConfig),
		documents:             make(map[lsproto.DocumentUri]string),
		diagnostics:           make(map[lsproto.DocumentUri][]rule.RuleDiagnostic),
		refreshCh:             make(chan struct{}, 1),
		debounceCh:            make(chan struct{}, 1),
		pendingLintURIs:       make(map[lsproto.DocumentUri]struct{}),
	}
	// Install the LSP-based ESLint-plugin dispatcher. It sends every
	// compat batch back to the LSP client as a `rslint/lintCompatBatch`
	// request; the client (typically the VS Code extension) runs the
	// rules in its own WorkerPool and returns diagnostics. This
	// replaces the legacy sidecar-process architecture.
	//
	// Installed at construction time and never swapped out — the
	// dispatcher itself is stateless. Whether plugin rules actually
	// run is decided by the linter (whether any ConfiguredRule has
	// IsEslintPluginRule=true for the current file) and by the
	// client (whether it has any plugin entries configured); the Go
	// server doesn't need to know.
	s.compatDispatcher = newLintCompatLSPDispatcher(s)
	return s
}

var (
	_ project.Client = (*Server)(nil)
)

type pendingClientRequest struct {
	req    *lsproto.RequestMessage
	cancel context.CancelFunc
}

type Reader interface {
	Read() (*lsproto.Message, error)
}

type Writer interface {
	Write(msg *lsproto.Message) error
}

type lspReader struct {
	r *lsproto.BaseReader
}

type lspWriter struct {
	w *lsproto.BaseWriter
}

func (r *lspReader) Read() (*lsproto.Message, error) {
	data, err := r.r.Read()
	if err != nil {
		return nil, err
	}

	req := &lsproto.Message{}
	if err := json.Unmarshal(data, req); err != nil {
		return nil, fmt.Errorf("%w: %w", lsproto.ErrorCodeInvalidRequest, err)
	}

	return req, nil
}

func ToReader(r io.Reader) Reader {
	return &lspReader{r: lsproto.NewBaseReader(r)}
}

func (w *lspWriter) Write(msg *lsproto.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return w.w.Write(data)
}

func ToWriter(w io.Writer) Writer {
	return &lspWriter{w: lsproto.NewBaseWriter(w)}
}

var (
	_ Reader = (*lspReader)(nil)
	_ Writer = (*lspWriter)(nil)
)

type Server struct {
	r Reader
	w Writer

	stderr io.Writer

	clientSeq               atomic.Int32
	requestQueue            chan *lsproto.RequestMessage
	outgoingQueue           chan *lsproto.Message
	pendingClientRequests   map[jsonrpc.ID]pendingClientRequest
	pendingClientRequestsMu sync.Mutex
	pendingServerRequests   map[jsonrpc.ID]chan *lsproto.ResponseMessage
	pendingServerRequestsMu sync.Mutex

	cwd                string
	fs                 vfs.FS
	defaultLibraryPath string
	typingsLocation    string

	backgroundCtx    context.Context
	initializeParams *lsproto.InitializeParams
	positionEncoding lsproto.PositionEncodingKind
	watchEnabled     bool
	watchers         collections.SyncSet[project.WatcherID]

	session *project.Session

	// enables tests to share a cache of parsed source files
	parseCache *project.ParseCache

	// !!! temporary; remove when we have `handleDidChangeConfiguration`/implicit project config support
	compilerOptionsForInferredProjects *core.CompilerOptions

	// rslint config
	jsConfigs        map[string]config.RslintConfig // configDirectory URI -> config entries (from JS/TS configs)
	jsonConfig       config.RslintConfig            // fallback JSON config (rslint.json/rslint.jsonc)
	rslintConfigPath string                         // path to rslint.json/rslint.jsonc, empty if not found
	// tsConfigPaths holds resolved parserOptions.project tsconfig paths.
	// For the JSON-config path this is a single global list.
	// For the JS-config path (multi-config monorepo) use tsConfigPathsByConfig
	// which keys per-config-directory so a nested config with no tsconfig
	// does not disable filtering for files under other configs.
	tsConfigPaths         []string
	tsConfigPathsByConfig map[string][]string                           // configDirectory URI -> resolved tsconfig paths (nil value = allow-all for that config's files)
	documents             map[lsproto.DocumentUri]string                // URI -> content
	diagnostics           map[lsproto.DocumentUri][]rule.RuleDiagnostic // URI -> diagnostics

	// refreshCh receives signals from RefreshDiagnostics (called by Session's
	// background goroutine) and is consumed by the main dispatch loop so that
	// relinting runs on the main goroutine (session is not goroutine-safe).
	refreshCh chan struct{}

	// debounceCh receives a signal when the debounce timer fires, telling
	// the dispatch loop to lint only the URIs in pendingLintURIs.
	debounceCh      chan struct{}
	pendingLintURIs map[lsproto.DocumentUri]struct{}
	lintTimer       *time.Timer

	// compatDispatcher is set once in NewServer to the LSP-based
	// dispatcher (see lintcompat_dispatcher.go). Every batch travels
	// from here through `rslint/lintCompatBatch` to the client; the
	// client (extension) runs the rules in its own WorkerPool and
	// returns diagnostics. The Go server itself never spawns Node.
	//
	// Stateless: capture-only over `*Server`; no per-request mutable
	// state.
	compatDispatcher linter.CompatBatchHandler
}

// FS implements project.ServiceHost.
func (s *Server) FS() vfs.FS {
	return s.fs
}

// DefaultLibraryPath implements project.ServiceHost.
func (s *Server) DefaultLibraryPath() string {
	return s.defaultLibraryPath
}

// TypingsLocation implements project.ServiceHost.
func (s *Server) TypingsLocation() string {
	return s.typingsLocation
}

// GetCurrentDirectory implements project.ServiceHost.
func (s *Server) GetCurrentDirectory() string {
	return s.cwd
}

// Trace implements project.ServiceHost.
func (s *Server) Trace(msg string) {
	s.Log(msg)
}

// Client implements project.ServiceHost.
func (s *Server) Client() project.Client {
	if !s.watchEnabled {
		return nil
	}
	return s
}

// WatchFiles implements project.Client.
func (s *Server) WatchFiles(ctx context.Context, id project.WatcherID, watchers []*lsproto.FileSystemWatcher) error {
	_, err := s.sendRequest(ctx, lsproto.MethodClientRegisterCapability, &lsproto.RegistrationParams{
		Registrations: []*lsproto.Registration{
			{
				Id: string(id),
				RegisterOptions: &lsproto.RegisterOptions{
					WorkspaceDidChangeWatchedFiles: &lsproto.DidChangeWatchedFilesRegistrationOptions{
						Watchers: watchers,
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to register file watcher: %w", err)
	}

	s.watchers.Add(id)
	return nil
}

// UnwatchFiles implements project.Client.
func (s *Server) UnwatchFiles(ctx context.Context, id project.WatcherID) error {
	if s.watchers.Has(id) {
		_, err := s.sendRequest(ctx, lsproto.MethodClientUnregisterCapability, &lsproto.UnregistrationParams{
			Unregisterations: []*lsproto.Unregistration{
				{
					Id:     string(id),
					Method: string(lsproto.MethodWorkspaceDidChangeWatchedFiles),
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to unregister file watcher: %w", err)
		}

		s.watchers.Delete(id)
		return nil
	}

	return fmt.Errorf("no file watcher exists with ID %s", id)
}

// RefreshDiagnostics implements project.Client.
// Called from Session's background goroutine when project state changes
// (e.g. tsconfig reload). We signal the main dispatch loop via refreshCh
// so that relinting happens on the main goroutine (session is not goroutine-safe).
func (s *Server) RefreshDiagnostics(ctx context.Context) error {
	select {
	case s.refreshCh <- struct{}{}:
	default:
		// Already pending — no need to queue another signal
	}
	return nil
}

// PublishDiagnostics implements project.Client.
func (s *Server) PublishDiagnostics(ctx context.Context, params *lsproto.PublishDiagnosticsParams) error {
	s.outgoingQueue <- lsproto.TextDocumentPublishDiagnosticsInfo.NewNotificationMessage(params).Message()
	return nil
}

// RefreshInlayHints implements project.Client.
func (s *Server) RefreshInlayHints(ctx context.Context) error {
	// TODO: implement inlay hints refresh
	return nil
}

// RefreshCodeLens implements project.Client.
func (s *Server) RefreshCodeLens(ctx context.Context) error {
	// TODO: implement code lens refresh
	return nil
}

// ProgressStart implements project.Client.
func (s *Server) ProgressStart(message *diagnostics.Message, args ...any) {}

// ProgressFinish implements project.Client.
func (s *Server) ProgressFinish(message *diagnostics.Message, args ...any) {}

// SendTelemetry implements project.Client.
func (s *Server) SendTelemetry(ctx context.Context, telemetry lsproto.TelemetryEvent) error {
	return nil
}

// IsActive implements project.Client.
func (s *Server) IsActive() bool { return s.session != nil }

func (s *Server) Run() error {
	// Signals: os.Interrupt (SIGINT) + SIGTERM are the LSP supervisor's
	// standard "shut down" signals. SIGHUP catches the controlling-
	// terminal-disconnect case (rare for LSP, but if the editor exits
	// without cleanly closing stdin or sends SIGHUP — some unusual
	// terminal hosts do — we want graceful session teardown). On
	// Windows SIGHUP is a no-op for signal.Notify; the registration
	// is portable.
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	g, ctx := errgroup.WithContext(sigCtx)
	s.backgroundCtx = ctx
	g.Go(func() error { return s.dispatchLoop(ctx) })
	g.Go(func() error { return s.writeLoop(ctx) })

	// Don't run readLoop in the group, as it blocks on stdin read and cannot be cancelled.
	readLoopErr := make(chan error, 1)
	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-readLoopErr:
			return err
		}
	})
	go func() { readLoopErr <- s.readLoop(ctx) }()

	// Propagate a real error only when no shutdown signal fired. We MUST
	// test the SIGNAL context (sigCtx), not the errgroup-derived ctx:
	// errgroup cancels the derived ctx before Wait() returns, so ctx.Err()
	// is ALWAYS non-nil here — gating on `ctx.Err() != nil` (as the prior
	// code did) is therefore always true and propagates even a clean
	// shutdown's spurious context.Canceled as a fatal error (non-zero exit
	// on graceful stop). sigCtx is cancelled only by SIGINT/SIGTERM/SIGHUP,
	// so sigCtx.Err() != nil means the server was asked to stop and the
	// g.Wait() error is expected fallout. io.EOF is benign (client
	// disconnected).
	if err := runLoopError(g.Wait(), sigCtx.Err() != nil); err != nil {
		return err
	}
	return nil
}

// runLoopError decides whether g.Wait()'s result is a real failure to
// propagate from Run, vs expected shutdown fallout. `signalled` is the
// SIGNAL context's cancellation state — NOT the errgroup-derived ctx,
// which errgroup cancels before Wait() returns (so it is always cancelled
// here, and gating on it would always be true — propagating even a clean
// shutdown's spurious context.Canceled).
func runLoopError(err error, signalled bool) error {
	if err != nil && !errors.Is(err, io.EOF) && !signalled {
		return err
	}
	return nil
}

func (s *Server) readLoop(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		msg, err := s.read()
		if err != nil {
			if errors.Is(err, lsproto.ErrorCodeInvalidRequest) {
				s.sendError(nil, err)
				continue
			}
			return err
		}

		if s.initializeParams == nil && msg.Kind == jsonrpc.MessageKindRequest {
			req := msg.AsRequest()
			if req.Method == lsproto.MethodInitialize {
				initParams, ok := req.Params.(*lsproto.InitializeParams)
				if !ok {
					s.sendError(req.ID, lsproto.ErrorCodeInvalidParams)
					continue
				}
				resp, err := s.handleInitialize(ctx, initParams)
				if err != nil {
					return err
				}
				s.sendResult(req.ID, resp)
			} else {
				s.sendError(req.ID, lsproto.ErrorCodeServerNotInitialized)
			}
			continue
		}

		if msg.Kind == jsonrpc.MessageKindResponse {
			resp := msg.AsResponse()
			s.pendingServerRequestsMu.Lock()
			if respChan, ok := s.pendingServerRequests[*resp.ID]; ok {
				respChan <- resp
				close(respChan)
				delete(s.pendingServerRequests, *resp.ID)
			}
			s.pendingServerRequestsMu.Unlock()
		} else {
			req := msg.AsRequest()
			if req.Method == lsproto.MethodCancelRequest {
				if cancelParams, ok := req.Params.(*lsproto.CancelParams); ok {
					s.cancelRequest(cancelParams.Id)
				}
			} else {
				s.requestQueue <- req
			}
		}
	}
}

func (s *Server) cancelRequest(rawID lsproto.IntegerOrString) {
	id := lsproto.NewID(rawID)
	s.pendingClientRequestsMu.Lock()
	defer s.pendingClientRequestsMu.Unlock()
	if pendingReq, ok := s.pendingClientRequests[*id]; ok {
		pendingReq.cancel()
		delete(s.pendingClientRequests, *id)
	}
}

func (s *Server) read() (*lsproto.Message, error) {
	return s.r.Read()
}

func (s *Server) dispatchLoop(ctx context.Context) error {
	ctx, lspExit := context.WithCancel(ctx)
	defer lspExit()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.refreshCh:
			// Session detected a project change (e.g. tsconfig reload).
			// Re-lint all open documents on the main goroutine.
			for uri := range s.documents {
				s.pushDiagnostics(uri)
			}
		case <-s.debounceCh:
			// Debounce timer fired — lint only documents with pending changes.
			for uri := range s.pendingLintURIs {
				s.pushDiagnostics(uri)
			}
			clear(s.pendingLintURIs)
		case req := <-s.requestQueue:
			requestCtx := ctx
			if req.ID != nil {
				var cancel context.CancelFunc
				requestCtx, cancel = context.WithCancel(core.WithRequestID(requestCtx, req.ID.String()))
				s.pendingClientRequestsMu.Lock()
				s.pendingClientRequests[*req.ID] = pendingClientRequest{
					req:    req,
					cancel: cancel,
				}
				s.pendingClientRequestsMu.Unlock()
			}

			handle := func() {
				defer func() {
					if r := recover(); r != nil {
						stack := debug.Stack()
						s.Log("panic handling request", req.Method, r, string(stack))
						if isBlockingMethod(req.Method) {
							lspExit()
						} else {
							if req.ID != nil {
								s.sendError(req.ID, fmt.Errorf("%w: panic handling request %s: %v", lsproto.ErrorCodeInternalError, req.Method, r))
							} else {
								s.Log("unhandled panic in notification", req.Method, r)
							}
						}
					}
				}()
				if err := s.handleRequestOrNotification(requestCtx, req); err != nil {
					if errors.Is(err, context.Canceled) {
						s.sendError(req.ID, lsproto.ErrorCodeRequestCancelled)
					} else if errors.Is(err, io.EOF) {
						lspExit()
					} else {
						s.sendError(req.ID, err)
					}
				}

				if req.ID != nil {
					s.pendingClientRequestsMu.Lock()
					delete(s.pendingClientRequests, *req.ID)
					s.pendingClientRequestsMu.Unlock()
				}
			}

			if isBlockingMethod(req.Method) {
				handle()
			} else {
				go handle()
			}
		}
	}
}

func (s *Server) writeLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-s.outgoingQueue:
			if err := s.w.Write(msg); err != nil {
				return fmt.Errorf("failed to write message: %w", err)
			}
		}
	}
}

func (s *Server) sendRequest(ctx context.Context, method lsproto.Method, params any) (any, error) {
	id := jsonrpc.NewIDString(fmt.Sprintf("ts%d", s.clientSeq.Add(1)))
	req := &lsproto.RequestMessage{
		ID:     id,
		Method: method,
		Params: params,
	}

	responseChan := make(chan *lsproto.ResponseMessage, 1)
	s.pendingServerRequestsMu.Lock()
	s.pendingServerRequests[*id] = responseChan
	s.pendingServerRequestsMu.Unlock()

	// Enqueue with ctx-awareness. Without this, a full outgoingQueue
	// (writer wedged on a slow / unresponsive client) would block the
	// caller indefinitely even when ctx is already cancelled.
	select {
	case s.outgoingQueue <- req.Message():
	case <-ctx.Done():
		s.pendingServerRequestsMu.Lock()
		delete(s.pendingServerRequests, *id)
		s.pendingServerRequestsMu.Unlock()
		return nil, ctx.Err()
	}

	select {
	case <-ctx.Done():
		s.pendingServerRequestsMu.Lock()
		defer s.pendingServerRequestsMu.Unlock()
		if respChan, ok := s.pendingServerRequests[*id]; ok {
			close(respChan)
			delete(s.pendingServerRequests, *id)
		}
		return nil, ctx.Err()
	case resp := <-responseChan:
		if resp.Error != nil {
			return nil, fmt.Errorf("request failed: %s", resp.Error.String())
		}
		return resp.Result, nil
	}
}

// sendRequestWithClientCancel is sendRequest plus a `$/cancelRequest`
// notification on ctx cancellation. Use it for server→client requests
// that participate in lint-level cancellation (LSP supersession,
// per-file ctx). Standard `sendRequest` does NOT notify the client on
// ctx cancel — its caller only sees the local pending entry get cleared,
// which is fine for short capability-registration requests but wrong
// for long-running custom requests where the client should bail too.
//
// Behavior on ctx.Done:
//
//   - Delete the local pending entry and close the response channel
//     (same as sendRequest's existing behavior).
//   - Send a `$/cancelRequest` notification to the client carrying the
//     same id. The client's vscode-jsonrpc layer then cancels the
//     handler's CancellationToken, so a well-behaved handler aborts
//     its work and returns either an error or a partial result.
//
// We INTENTIONALLY do not wait for the client's response after sending
// the cancel. The LSP cancel protocol is fire-and-forget; the client
// MAY still reply, MAY reply with an error, or MAY drop the request.
// We surface ctx.Err() to our caller either way — a stale response
// arriving later finds no pending entry and is silently dropped.
func (s *Server) sendRequestWithClientCancel(ctx context.Context, method lsproto.Method, params any) (any, error) {
	id := jsonrpc.NewIDString(fmt.Sprintf("ts%d", s.clientSeq.Add(1)))
	req := &lsproto.RequestMessage{
		ID:     id,
		Method: method,
		Params: params,
	}

	responseChan := make(chan *lsproto.ResponseMessage, 1)
	s.pendingServerRequestsMu.Lock()
	s.pendingServerRequests[*id] = responseChan
	s.pendingServerRequestsMu.Unlock()

	// Enqueue to the writer with ctx-awareness. A wedged writer (slow
	// client, network stall) can fill outgoingQueue (cap 100); without
	// the ctx case here, an already-cancelled or about-to-be-cancelled
	// ctx couldn't unblock the caller — the send would wait
	// indefinitely. We also surface a clean ctx.Err() instead of the
	// previous "request silently hangs" symptom.
	select {
	case s.outgoingQueue <- req.Message():
	case <-ctx.Done():
		s.pendingServerRequestsMu.Lock()
		delete(s.pendingServerRequests, *id)
		s.pendingServerRequestsMu.Unlock()
		// No $/cancelRequest sent: the peer never observed our request,
		// so there's nothing on the client side to cancel.
		return nil, ctx.Err()
	}

	select {
	case <-ctx.Done():
		// Last-second arrival check: ctx and response can race when a
		// reply lands at almost the same moment ctx fires. Go's `select`
		// picks fairly between ready cases — it MAY pick ctx.Done even
		// when responseChan already has a value. To avoid dropping that
		// response (and incidentally sending a spurious $/cancelRequest
		// to the client for a request the client already finished), we
		// take the pending-map lock to serialize against readerLoop and
		// then attempt a non-blocking drain of responseChan.
		//
		// CRITICAL: drain `responseChan` (the local variable), NOT a
		// fresh map lookup. readerLoop's three-step sequence under the
		// same mutex is `respChan <- resp; close(respChan);
		// delete(map)`. If readerLoop wins the lock first, the map
		// entry is gone but the buffered value still sits in
		// responseChan — its memory survives map deletion because we
		// hold a stack reference. A `map[id]` lookup here would return
		// the channel zero-value (nil), and a `<-nil` non-blocking
		// recv silently selects `default:`, dropping the response. The
		// previous version had exactly that bug and surfaced as ESLint
		// diagnostic flicker under fast keystroke / debounce supersession.
		s.pendingServerRequestsMu.Lock()
		_, stillPending := s.pendingServerRequests[*id]
		if stillPending {
			delete(s.pendingServerRequests, *id)
		}
		s.pendingServerRequestsMu.Unlock()

		// Case A: response was already delivered while we held nothing.
		// readerLoop deleted the pending entry → `stillPending` is
		// false here, but the value is sitting in responseChan
		// (buffer=1, value persisted from `respChan <- resp`). Drain
		// it from the local variable.
		//
		// Case B: pending entry was still there → readerLoop hasn't
		// delivered. The non-blocking recv finds an empty channel and
		// falls through to the cancel-notify path below.
		select {
		case resp := <-responseChan:
			if resp != nil {
				if resp.Error != nil {
					return nil, fmt.Errorf("server request %s: %s", method, resp.Error.String())
				}
				return resp.Result, nil
			}
		default:
		}

		// No response was available — notify the client to cancel.
		// `$/cancelRequest` is a JSON-RPC notification (no ID);
		// vscode-jsonrpc handles cancel-token bookkeeping on the client
		// side. We tolerate a full outgoing queue: if the writer is
		// wedged we'd rather return ctx.Err() now than block here. The
		// client may not see the cancel — at worst it runs the handler
		// to completion and replies with results we silently drop.
		//
		// We always mint string-flavored IDs above (jsonrpc.NewIDString),
		// so `id.String()` round-trips losslessly into the String arm
		// of IntegerOrString. If the ID minting ever changes to integer
		// IDs, CancelParams.Id construction below must change too.
		idStr := id.String()
		cancelMsg := (&lsproto.RequestMessage{
			Method: lsproto.MethodCancelRequest,
			Params: &lsproto.CancelParams{
				Id: lsproto.IntegerOrString{String: &idStr},
			},
		}).Message()
		select {
		case s.outgoingQueue <- cancelMsg:
		default:
			// queue full — best-effort drop; client still survives,
			// we just don't preempt its handler. Logged so a wedged
			// outgoing path is visible.
			log.Printf("[rslint] sendRequestWithClientCancel: outgoingQueue full, dropping $/cancelRequest for id=%v", *id)
		}

		return nil, ctx.Err()
	case resp := <-responseChan:
		if resp == nil {
			return nil, fmt.Errorf("server request %s: response channel closed without value", method)
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("server request %s: %s", method, resp.Error.String())
		}
		return resp.Result, nil
	}
}

func (s *Server) sendResult(id *jsonrpc.ID, result any) {
	s.sendResponse(&lsproto.ResponseMessage{
		ID:     id,
		Result: result,
	})
}

func (s *Server) sendError(id *jsonrpc.ID, err error) {
	code := int32(lsproto.ErrorCodeInternalError)
	var errCode lsproto.ErrorCode
	if errors.As(err, &errCode) {
		code = int32(errCode)
	}
	// TODO(jakebailey): error data
	s.sendResponse(&lsproto.ResponseMessage{
		ID: id,
		Error: &jsonrpc.ResponseError{
			Code:    code,
			Message: err.Error(),
		},
	})
}

func (s *Server) sendResponse(resp *lsproto.ResponseMessage) {
	s.outgoingQueue <- resp.Message()
}

func (s *Server) handleRequestOrNotification(ctx context.Context, req *lsproto.RequestMessage) error {
	if handler := handlers()[req.Method]; handler != nil {
		return handler(s, ctx, req)
	}
	s.Log("unknown method", req.Method)
	if req.ID != nil {
		s.sendError(req.ID, lsproto.ErrorCodeInvalidRequest)
	}
	return nil
}

type handlerMap map[lsproto.Method]func(*Server, context.Context, *lsproto.RequestMessage) error

var handlers = sync.OnceValue(func() handlerMap {
	handlers := make(handlerMap)

	registerRequestHandler(handlers, lsproto.InitializeInfo, (*Server).handleInitialize)
	registerNotificationHandler(handlers, lsproto.InitializedInfo, (*Server).handleInitialized)
	registerRequestHandler(handlers, lsproto.ShutdownInfo, (*Server).handleShutdown)
	registerNotificationHandler(handlers, lsproto.ExitInfo, (*Server).handleExit)

	registerNotificationHandler(handlers, lsproto.TextDocumentDidOpenInfo, (*Server).handleDidOpen)
	registerNotificationHandler(handlers, lsproto.TextDocumentDidChangeInfo, (*Server).handleDidChange)
	registerNotificationHandler(handlers, lsproto.TextDocumentDidSaveInfo, (*Server).handleDidSave)
	registerNotificationHandler(handlers, lsproto.TextDocumentDidCloseInfo, (*Server).handleDidClose)
	registerNotificationHandler(handlers, lsproto.WorkspaceDidChangeWatchedFilesInfo, (*Server).handleDidChangeWatchedFiles)
	registerRequestHandler(handlers, lsproto.TextDocumentCodeActionInfo, (*Server).handleCodeAction)

	// Custom rslint notification
	handlers[lsproto.Method("rslint/configUpdate")] = func(s *Server, ctx context.Context, req *lsproto.RequestMessage) error {
		if err := s.handleConfigUpdate(ctx, req.Params); err != nil {
			log.Printf("[rslint] Error handling config update: %v", err)
		}
		return nil
	}

	return handlers
})

func registerNotificationHandler[Req any](handlers handlerMap, info lsproto.NotificationInfo[Req], fn func(*Server, context.Context, Req) error) {
	handlers[info.Method] = func(s *Server, ctx context.Context, req *lsproto.RequestMessage) error {
		var params Req
		// Ignore empty params; all generated params are either pointers or any.
		if req.Params != nil {
			p, ok := req.Params.(Req)
			if !ok {
				return fmt.Errorf("unexpected params type %T for %s", req.Params, info.Method)
			}
			params = p
		}
		if err := fn(s, ctx, params); err != nil {
			return err
		}
		return ctx.Err()
	}
}

func registerRequestHandler[Req, Resp any](handlers handlerMap, info lsproto.RequestInfo[Req, Resp], fn func(*Server, context.Context, Req) (Resp, error)) {
	handlers[info.Method] = func(s *Server, ctx context.Context, req *lsproto.RequestMessage) error {
		var params Req
		// Ignore empty params.
		if req.Params != nil {
			p, ok := req.Params.(Req)
			if !ok {
				return fmt.Errorf("unexpected params type %T for %s", req.Params, info.Method)
			}
			params = p
		}
		resp, err := fn(s, ctx, params)
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		s.sendResult(req.ID, resp)
		return nil
	}
}

func (s *Server) handleShutdown(ctx context.Context, params lsproto.NoParams) (lsproto.ShutdownResponse, error) {
	s.session.Close()
	return lsproto.ShutdownResponse{}, nil
}

func (s *Server) handleExit(ctx context.Context, params lsproto.NoParams) error {
	return io.EOF
}

func (s *Server) Log(msg ...any) {
	fmt.Fprintln(s.stderr, msg...)
}

// !!! temporary; remove when we have `handleDidChangeConfiguration`/implicit project config support
func (s *Server) SetCompilerOptionsForInferredProjects(options *core.CompilerOptions) {
	s.compilerOptionsForInferredProjects = options
	if s.session != nil {
		s.session.DidChangeCompilerOptionsForInferredProjects(context.Background(), options)
	}
}

func isBlockingMethod(method lsproto.Method) bool {
	switch method {
	case lsproto.MethodInitialize,
		lsproto.MethodInitialized,
		lsproto.MethodTextDocumentDidOpen,
		lsproto.MethodTextDocumentDidChange,
		lsproto.MethodTextDocumentDidSave,
		lsproto.MethodTextDocumentDidClose,
		lsproto.MethodWorkspaceDidChangeWatchedFiles,
		lsproto.MethodTextDocumentCodeAction,
		// rslint/configUpdate writes to s.jsConfigs, which is read by handlers
		// dispatched on the main loop (e.g. didOpen/didChange). Running it on
		// the same serialized dispatch loop avoids a data race without locks.
		lsproto.Method("rslint/configUpdate"):
		return true
	}
	return false
}

func ptrTo[T any](v T) *T {
	return &v
}

func ptrIsTrue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}
