// modified based on https://github.com/microsoft/typescript-go/blob/cedc0cbe6c188f9bfe6a51af00c79be48c9ab74d/internal/lsp/server.go#L1
package lsp

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	return &Server{
		r:                      opts.In,
		w:                      opts.Out,
		stderr:                 opts.Err,
		requestQueue:           make(chan *lsproto.RequestMessage, 100),
		outgoingQueue:          make(chan *lsproto.Message, 100),
		pendingClientRequests:  make(map[jsonrpc.ID]pendingClientRequest),
		pendingServerRequests:  make(map[jsonrpc.ID]chan *lsproto.ResponseMessage),
		cwd:                    opts.Cwd,
		fs:                     opts.FS,
		defaultLibraryPath:     opts.DefaultLibraryPath,
		typingsLocation:        opts.TypingsLocation,
		parseCache:             opts.ParseCache,
		jsConfigs:              make(map[string]config.RslintConfig),
		jsUnavailableConfigs:   make(map[string]struct{}),
		documents:              make(map[lsproto.DocumentUri]string),
		diagnostics:            make(map[lsproto.DocumentUri][]rule.RuleDiagnostic),
		refreshCh:              make(chan struct{}, 1),
		debounceCh:             make(chan struct{}, 1),
		pendingLintURIs:        make(map[lsproto.DocumentUri]struct{}),
		pluginResultCh:         make(chan pluginLintResult, 16),
		docGeneration:          make(map[lsproto.DocumentUri]uint64),
		inflightPluginDispatch: make(map[lsproto.DocumentUri]*pluginDispatchHandle),
	}
}

var (
	_ project.Client = (*Server)(nil)
)

type pendingClientRequest struct {
	req    *lsproto.RequestMessage
	ctx    context.Context
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
	// jsConfigs is keyed by the catalog's absolute filesystem directory
	// byte-for-byte so Go ownership and Node plugin routing share one identity.
	jsConfigs map[string]config.RslintConfig
	// The resolver is rebuilt atomically with each config transaction. Its keys
	// are the same filesystem paths stored in jsConfigs.
	jsConfigOwnerResolver *config.ConfigOwnerResolver
	// configDiscoveryActive becomes true after the first structurally valid
	// configRefresh request. It lets Go's supplemental strict-ancestor JS and
	// config-scoped .gitignore watchers trigger a fresh transaction without
	// sending reverse requests before the client installs its handlers. The
	// extension remains the sole refresh owner for workspace/descendant JS and
	// JSON changes.
	configDiscoveryActive bool
	// configDiscoveryHasLastGood distinguishes a committed catalog with at
	// least one usable JS config from an empty catalog or the synthetic
	// unavailable boundaries used to keep LSP alive when every JS config is
	// broken. Refresh failures preserve only the usable JS catalog as last-good.
	configDiscoveryHasLastGood bool
	// configSnapshotIncludesGitignore means the current catalog already contains
	// the .gitignore view captured during its transaction. Before the first
	// committed snapshot, the JSON startup config still uses the live policy.
	configSnapshotIncludesGitignore bool
	// jsUnavailableConfigs contains absolute config-directory paths for failed
	// JS/TS config boundaries. They participate in ownership but suppress lint.
	jsUnavailableConfigs map[string]struct{}
	jsonConfig           config.RslintConfig // fallback JSON config (rslint.json/rslint.jsonc)
	rslintConfigPath     string              // path to rslint.json/rslint.jsonc, empty if not found
	// tsConfigPaths holds resolved parserOptions.project tsconfig paths.
	// For the JSON-config path this is a single global list.
	// For the JS-config path (multi-config monorepo) use tsConfigPathsByConfig
	// which keys per-config-directory so a nested config with no tsconfig
	// does not disable filtering for files under other configs.
	tsConfigPaths []string
	// A nil map value means the corresponding config has no type information.
	tsConfigPathsByConfig map[string][]string
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

	// eslintPluginDispatch sends one plugin-lint batch to the Node host over
	// the `rslint/pluginLint` reverse request and decodes its result.
	// nil until the first committed config transaction installs it. Safe to call from a
	// goroutine (it only touches sendRequest, which is goroutine-safe).
	eslintPluginDispatch linter.EslintPluginDispatcher
	// eslintPluginConfigGeneration identifies the JS config and Node worker
	// generation that must serve reverse plugin-lint requests. It changes only
	// on a serialized config transaction and is captured before dispatching work
	// to a goroutine.
	eslintPluginConfigGeneration string
	// eslintPluginRules contains exactly the object-form plugin rule names
	// activated for eslintPluginConfigGeneration. GlobalRuleRegistry keeps
	// placeholders process-wide, so this generation-local gate prevents a
	// placeholder left by an older config from being dispatched to the current
	// Node host. nil preserves the unscoped behavior used by isolated tests and
	// before the first config transaction; every committed generation installs
	// a non-nil set, including an empty one.
	eslintPluginRules map[string]struct{}

	// fixAllNativeLint, when non-nil, overrides the per-pass native lint used by
	// computeFixAllContent. Production leaves it nil (defaultFixAllNativeLint is
	// used, driving an isolated overlay Program); tests inject a mock to exercise the
	// plugin-fix fold loop without spinning up a language service.
	fixAllNativeLint func(ctx context.Context, uri lsproto.DocumentUri, pass int, content string, rslintConfig config.RslintConfig, configCwd string, isJSConfig bool, tsConfigPaths []string) (lintPassResult, error)

	// pluginReverseTimeout bounds each eslint-plugin reverse request to the
	// client (rslint/pluginLint) on BOTH paths: source.fixAll (summed across
	// passes, where it runs on the dispatch loop as a blocking method) and the
	// background diagnostics dispatch (per request). A wedged or mid-rebuild
	// client that never answers would otherwise stall editor interaction or leak
	// the dispatch goroutine + its pending-request entry. On expiry fixAll folds
	// native-only fixes and the diagnostics dispatch is dropped. Zero means use
	// defaultPluginReverseTimeout; tests set a small value to exercise the
	// deadline without a real client.
	pluginReverseTimeout time.Duration

	// pluginResultCh delivers eslint-plugin diagnostics computed off the
	// dispatch loop (in a goroutine that calls eslintPluginDispatch) back to
	// the main dispatch loop, which is the ONLY goroutine allowed to write
	// the lock-free s.diagnostics map and publish. Mirrors refreshCh/debounceCh.
	pluginResultCh chan pluginLintResult

	// inflightPluginDispatch tracks the in-flight background plugin dispatch per
	// URI so a superseding keystroke (or a document close) can cancel it — both
	// Go-side (free the goroutine + pending-request entry) and on the client (a
	// $/cancelRequest so the Node worker stops instead of running to completion).
	// Guarded by inflightPluginDispatchMu.
	inflightPluginDispatch   map[lsproto.DocumentUri]*pluginDispatchHandle
	inflightPluginDispatchMu sync.Mutex

	// docGeneration tracks a per-URI revision counter. pushDiagnostics bumps it
	// on every (re)lint and stamps the spawned plugin goroutine with the value.
	// A plugin result whose generation no longer matches is stale (a newer
	// keystroke already re-linted) and is dropped — cooperative ctx cancel does
	// not guarantee the in-flight worker stops, so generation guards correctness.
	// Only accessed from the main dispatch loop goroutine.
	docGeneration map[lsproto.DocumentUri]uint64
}

// pluginLintResult carries one URI's eslint-plugin diagnostics from the
// dispatch goroutine back to the main loop, tagged with the generation that
// was current when the lint was kicked off.
type pluginLintResult struct {
	uri        lsproto.DocumentUri
	generation uint64
	diags      []rule.RuleDiagnostic
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)
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

	if err := g.Wait(); err != nil && !errors.Is(err, io.EOF) && ctx.Err() != nil {
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
				initParams, err := decodeParams[*lsproto.InitializeParams](req)
				if err != nil {
					s.sendError(req.ID, err)
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
				if cancelParams, err := decodeParams[*lsproto.CancelParams](req); err == nil {
					s.cancelRequest(cancelParams.Id)
				}
			} else {
				if req.ID != nil {
					s.registerClientRequest(ctx, req)
				}
				s.requestQueue <- req
			}
		}
	}
}

func (s *Server) cancelRequest(rawID lsproto.IntegerOrString) {
	id := lsproto.NewID(rawID)
	s.pendingClientRequestsMu.Lock()
	if pendingReq, ok := s.pendingClientRequests[*id]; ok {
		s.pendingClientRequestsMu.Unlock()
		pendingReq.cancel()
		return
	}
	s.pendingClientRequestsMu.Unlock()
}

func (s *Server) registerClientRequest(ctx context.Context, req *lsproto.RequestMessage) context.Context {
	requestCtx, cancel := context.WithCancel(core.WithRequestID(ctx, req.ID.String()))
	s.pendingClientRequestsMu.Lock()
	if s.pendingClientRequests == nil {
		s.pendingClientRequests = make(map[jsonrpc.ID]pendingClientRequest)
	}
	if previous, exists := s.pendingClientRequests[*req.ID]; exists {
		previous.cancel()
	}
	s.pendingClientRequests[*req.ID] = pendingClientRequest{req: req, ctx: requestCtx, cancel: cancel}
	s.pendingClientRequestsMu.Unlock()
	return requestCtx
}

func (s *Server) clientRequestContext(ctx context.Context, req *lsproto.RequestMessage) context.Context {
	s.pendingClientRequestsMu.Lock()
	pending, ok := s.pendingClientRequests[*req.ID]
	s.pendingClientRequestsMu.Unlock()
	if ok {
		return pending.ctx
	}
	// Tests and embedded callers may inject directly into requestQueue.
	// Production requests are registered by readLoop before dispatch.
	return s.registerClientRequest(ctx, req)
}

func (s *Server) finishClientRequest(id *jsonrpc.ID) {
	s.pendingClientRequestsMu.Lock()
	pending, ok := s.pendingClientRequests[*id]
	delete(s.pendingClientRequests, *id)
	s.pendingClientRequestsMu.Unlock()
	if ok {
		pending.cancel()
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
		case r := <-s.pluginResultCh:
			// eslint-plugin diagnostics computed in a goroutine. Merge them
			// into the lock-free s.diagnostics map and re-publish — on the
			// main loop, the only goroutine permitted to touch s.diagnostics.
			s.mergePluginDiagnostics(r)
		case req := <-s.requestQueue:
			requestCtx := ctx
			if req.ID != nil {
				requestCtx = s.clientRequestContext(requestCtx, req)
			}

			handle := func() {
				if req.ID != nil {
					defer s.finishClientRequest(req.ID)
				}
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
		_, pending := s.pendingServerRequests[*id]
		if pending {
			delete(s.pendingServerRequests, *id)
		}
		s.pendingServerRequestsMu.Unlock()
		if pending {
			s.sendCancelRequest(id)
		}
		return nil, ctx.Err()
	case resp := <-responseChan:
		if resp.Error != nil {
			return nil, fmt.Errorf("request failed: %s", resp.Error.String())
		}
		return resp.Result, nil
	}
}

// pluginDispatchHandle is the cancel handle for an in-flight background plugin
// dispatch. sendRequest forwards context cancellation to the client after the
// reverse request has been queued.
type pluginDispatchHandle struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// sendCancelRequest asks the client to cancel a reverse request so its worker
// can stop instead of running to completion. It is best-effort: cancellation
// must never block the caller when the outgoing queue is saturated.
func (s *Server) sendCancelRequest(id *jsonrpc.ID) {
	var raw lsproto.IntegerOrString
	if n, ok := id.TryInt(); ok {
		raw.Integer = &n
	} else {
		str := id.String()
		raw.String = &str
	}
	select {
	case s.outgoingQueue <- lsproto.CancelRequestInfo.NewNotificationMessage(&lsproto.CancelParams{Id: raw}).Message():
	default:
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

	handlers[methodConfigRefresh] = func(s *Server, ctx context.Context, req *lsproto.RequestMessage) error {
		if req.ID == nil {
			return fmt.Errorf("%w: rslint/configRefresh must be a request", lsproto.ErrorCodeInvalidRequest)
		}
		params, err := decodeParams[configRefreshRequest](req)
		if err != nil {
			return err
		}
		response, err := s.handleConfigRefresh(ctx, params)
		if err != nil {
			return err
		}
		s.sendResult(req.ID, response)
		return nil
	}

	return handlers
})

func registerNotificationHandler[Req any](handlers handlerMap, info lsproto.NotificationInfo[Req], fn func(*Server, context.Context, Req) error) {
	handlers[info.Method] = func(s *Server, ctx context.Context, req *lsproto.RequestMessage) error {
		params, err := decodeParams[Req](req)
		if err != nil {
			return err
		}
		if err := fn(s, ctx, params); err != nil {
			return err
		}
		return ctx.Err()
	}
}

func registerRequestHandler[Req, Resp any](handlers handlerMap, info lsproto.RequestInfo[Req, Resp], fn func(*Server, context.Context, Req) (Resp, error)) {
	handlers[info.Method] = func(s *Server, ctx context.Context, req *lsproto.RequestMessage) error {
		params, err := decodeParams[Req](req)
		if err != nil {
			return err
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

// decodeParams accepts both raw params produced by the LSP reader and typed
// params used by embedded callers and tests.
func decodeParams[Req any](req *lsproto.RequestMessage) (Req, error) {
	if req.Params != nil {
		if params, ok := req.Params.(Req); ok {
			return params, nil
		}
	}
	return lsproto.UnmarshalParams[Req](req)
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

// defaultPluginReverseTimeout caps each eslint-plugin reverse request to the
// client — the source.fixAll passes (summed) and each background diagnostics
// dispatch. It is a generous BACKSTOP, not a precise budget: a superseded
// diagnostics dispatch is already discarded by the generation stamp, so this
// only has to bound a client that is genuinely wedged (never answers and is
// never superseded). 30s sits well above any legitimate single-file plugin lint
// — so a slow-but-valid lint is never cut off — while still freeing a dead
// client's goroutine and unblocking the fixAll dispatch loop in bounded time.
// (Fine-grained supersede cancellation via $/cancelRequest, which would let this
// be tightened, is a separate follow-up; see dispatchPluginLint.)
const defaultPluginReverseTimeout = 30 * time.Second

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
		// Config commits write maps that document handlers read lock-free, so
		// refresh transactions run on the same serialized dispatch loop.
		methodConfigRefresh:
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
