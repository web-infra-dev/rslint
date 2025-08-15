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

	"github.com/go-json-experiment/json"
	"github.com/microsoft/typescript-go/shim/collections"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rule"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/language"
)

type ServerOptions struct {
	In  Reader
	Out Writer
	Err io.Writer

	Cwd                string
	FS                 vfs.FS
	DefaultLibraryPath string
	TypingsLocation    string

	ParsedFileCache project.ParsedFileCache
}

func NewServer(opts *ServerOptions) *Server {
	if opts.Cwd == "" {
		panic("Cwd is required")
	}
	return &Server{
		r:                     opts.In,
		w:                     opts.Out,
		stderr:                opts.Err,
		requestQueue:          make(chan *lsproto.RequestMessage, 100),
		outgoingQueue:         make(chan *lsproto.Message, 100),
		pendingClientRequests: make(map[lsproto.ID]pendingClientRequest),
		pendingServerRequests: make(map[lsproto.ID]chan *lsproto.ResponseMessage),
		cwd:                   opts.Cwd,
		fs:                    opts.FS,
		defaultLibraryPath:    opts.DefaultLibraryPath,
		typingsLocation:       opts.TypingsLocation,
		parsedFileCache:       opts.ParsedFileCache,
		documents:             make(map[lsproto.DocumentUri]string),
		diagnostics:           make(map[lsproto.DocumentUri][]rule.RuleDiagnostic),
	}
}

var (
	_ project.ServiceHost = (*Server)(nil)
	_ project.Client      = (*Server)(nil)
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
		return nil, fmt.Errorf("%w: %w", lsproto.ErrInvalidRequest, err)
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
	pendingClientRequests   map[lsproto.ID]pendingClientRequest
	pendingClientRequestsMu sync.Mutex
	pendingServerRequests   map[lsproto.ID]chan *lsproto.ResponseMessage
	pendingServerRequestsMu sync.Mutex

	cwd                string
	fs                 vfs.FS
	defaultLibraryPath string
	typingsLocation    string

	initializeParams *lsproto.InitializeParams
	positionEncoding lsproto.PositionEncodingKind
	locale           language.Tag

	watchEnabled bool
	watcherID    atomic.Uint32
	watchers     collections.SyncSet[project.WatcherHandle]
	//nolint
	logger         *project.Logger
	projectService *project.Service

	// enables tests to share a cache of parsed source files
	parsedFileCache project.ParsedFileCache

	// !!! temporary; remove when we have `handleDidChangeConfiguration`/implicit project config support
	compilerOptionsForInferredProjects *core.CompilerOptions

	// rslint config
	rslintConfig config.RslintConfig
	documents    map[lsproto.DocumentUri]string                // URI -> content
	diagnostics  map[lsproto.DocumentUri][]rule.RuleDiagnostic // URI -> diagnostics

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
func (s *Server) WatchFiles(ctx context.Context, watchers []*lsproto.FileSystemWatcher) (project.WatcherHandle, error) {
	watcherId := fmt.Sprintf("watcher-%d", s.watcherID.Add(1))
	_, err := s.sendRequest(ctx, lsproto.MethodClientRegisterCapability, &lsproto.RegistrationParams{
		Registrations: []*lsproto.Registration{
			{
				Id:     watcherId,
				Method: string(lsproto.MethodWorkspaceDidChangeWatchedFiles),
				RegisterOptions: ptrTo(any(lsproto.DidChangeWatchedFilesRegistrationOptions{
					Watchers: watchers,
				})),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to register file watcher: %w", err)
	}

	handle := project.WatcherHandle(watcherId)
	s.watchers.Add(handle)
	return handle, nil
}

// UnwatchFiles implements project.Client.
func (s *Server) UnwatchFiles(ctx context.Context, handle project.WatcherHandle) error {
	if s.watchers.Has(handle) {
		_, err := s.sendRequest(ctx, lsproto.MethodClientUnregisterCapability, &lsproto.UnregistrationParams{
			Unregisterations: []*lsproto.Unregistration{
				{
					Id:     string(handle),
					Method: string(lsproto.MethodWorkspaceDidChangeWatchedFiles),
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to unregister file watcher: %w", err)
		}

		s.watchers.Delete(handle)
		return nil
	}

	return fmt.Errorf("no file watcher exists with ID %s", handle)
}

// RefreshDiagnostics implements project.Client.
func (s *Server) RefreshDiagnostics(ctx context.Context) error {
	if s.initializeParams.Capabilities == nil ||
		s.initializeParams.Capabilities.Workspace == nil ||
		s.initializeParams.Capabilities.Workspace.Diagnostics == nil ||
		!ptrIsTrue(s.initializeParams.Capabilities.Workspace.Diagnostics.RefreshSupport) {
		return nil
	}

	if _, err := s.sendRequest(ctx, lsproto.MethodWorkspaceDiagnosticRefresh, nil); err != nil {
		return fmt.Errorf("failed to refresh diagnostics: %w", err)
	}

	return nil
}

func (s *Server) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)
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
			if errors.Is(err, lsproto.ErrInvalidRequest) {
				s.sendError(nil, err)
				continue
			}
			return err
		}

		if s.initializeParams == nil && msg.Kind == lsproto.MessageKindRequest {
			req := msg.AsRequest()
			if req.Method == lsproto.MethodInitialize {
				resp, err := s.handleInitialize(ctx, req.Params.(*lsproto.InitializeParams))
				if err != nil {
					return err
				}
				s.sendResult(req.ID, resp)
			} else {
				s.sendError(req.ID, lsproto.ErrServerNotInitialized)
			}
			continue
		}

		if msg.Kind == lsproto.MessageKindResponse {
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
				s.cancelRequest(req.Params.(*lsproto.CancelParams).Id)
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
		case req := <-s.requestQueue:
			requestCtx := core.WithLocale(ctx, s.locale)
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
								s.sendError(req.ID, fmt.Errorf("%w: panic handling request %s: %v", lsproto.ErrInternalError, req.Method, r))
							} else {
								s.Log("unhandled panic in notification", req.Method, r)
							}
						}
					}
				}()
				if err := s.handleRequestOrNotification(requestCtx, req); err != nil {
					if errors.Is(err, context.Canceled) {
						s.sendError(req.ID, lsproto.ErrRequestCancelled)
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
	id := lsproto.NewIDString(fmt.Sprintf("ts%d", s.clientSeq.Add(1)))
	req := lsproto.NewRequestMessage(method, id, params)

	responseChan := make(chan *lsproto.ResponseMessage, 1)
	s.pendingServerRequestsMu.Lock()
	s.pendingServerRequests[*id] = responseChan
	s.pendingServerRequestsMu.Unlock()

	s.outgoingQueue <- req.Message()

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

func (s *Server) sendResult(id *lsproto.ID, result any) {
	s.sendResponse(&lsproto.ResponseMessage{
		ID:     id,
		Result: result,
	})
}

func (s *Server) sendError(id *lsproto.ID, err error) {
	code := lsproto.ErrInternalError.Code
	if errCode := (*lsproto.ErrorCode)(nil); errors.As(err, &errCode) {
		code = errCode.Code
	}
	// TODO(jakebailey): error data
	s.sendResponse(&lsproto.ResponseMessage{
		ID: id,
		Error: &lsproto.ResponseError{
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
		s.sendError(req.ID, lsproto.ErrInvalidRequest)
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
	registerRequestHandler(handlers, lsproto.TextDocumentDiagnosticInfo, (*Server).handleDocumentDiagnostic)
	registerRequestHandler(handlers, lsproto.TextDocumentCodeActionInfo, (*Server).handleCodeAction)
	return handlers
})

func registerNotificationHandler[Req any](handlers handlerMap, info lsproto.NotificationInfo[Req], fn func(*Server, context.Context, Req) error) {
	handlers[info.Method] = func(s *Server, ctx context.Context, req *lsproto.RequestMessage) error {
		var params Req
		// Ignore empty params; all generated params are either pointers or any.
		if req.Params != nil {
			params = req.Params.(Req)
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
			params = req.Params.(Req)
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

func (s *Server) handleShutdown(ctx context.Context, params any) (lsproto.ShutdownResponse, error) {
	s.projectService.Close()
	return lsproto.ShutdownResponse{}, nil
}

func (s *Server) handleExit(ctx context.Context, params any) error {
	return io.EOF
}

func (s *Server) Log(msg ...any) {
	fmt.Fprintln(s.stderr, msg...)
}

// !!! temporary; remove when we have `handleDidChangeConfiguration`/implicit project config support
func (s *Server) SetCompilerOptionsForInferredProjects(options *core.CompilerOptions) {
	s.compilerOptionsForInferredProjects = options
	if s.projectService != nil {
		s.projectService.SetCompilerOptionsForInferredProjects(options)
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
		lsproto.MethodWorkspaceDidChangeWatchedFiles:
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
