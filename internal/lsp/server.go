package lsp

import (
	"context"
	"log"
	"os"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"

	// "github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rule"
	util "github.com/web-infra-dev/rslint/internal/utils"
)

func ptrTo[T any](v T) *T {
	return &v
}

// LSP Server implementation
type LSPServer struct {
	conn        *jsonrpc2.Conn
	documents   map[lsproto.DocumentUri]string                // URI -> content
	diagnostics map[lsproto.DocumentUri][]rule.RuleDiagnostic // URI -> diagnostics
	// align with https://github.com/microsoft/typescript-go/blob/5cdf239b02006783231dd4da8ca125cef398cd27/internal/lsp/server.go#L147
	//nolint
	projectService *project.Service
	//nolint
	logger             *project.Logger
	fs                 vfs.FS
	defaultLibraryPath string
	typingsLocation    string
	cwd                string
	rslintConfig       config.RslintConfig
}

func (s *LSPServer) FS() vfs.FS {
	return s.fs
}
func (s *LSPServer) DefaultLibraryPath() string {
	return s.defaultLibraryPath
}
func (s *LSPServer) TypingsLocation() string {
	return s.typingsLocation
}
func (s *LSPServer) GetCurrentDirectory() string {
	return s.cwd
}
func (s *LSPServer) Client() project.Client {
	return nil
}

// FIXME: support watcher in the future
func (s *LSPServer) WatchFiles(ctx context.Context, watchers []*lsproto.FileSystemWatcher) (project.WatcherHandle, error) {
	return "", nil
}

func NewLSPServer() *LSPServer {
	log.Printf("cwd: %v", util.Must(os.Getwd()))
	return &LSPServer{
		documents:   make(map[lsproto.DocumentUri]string),
		diagnostics: make(map[lsproto.DocumentUri][]rule.RuleDiagnostic),
		fs:          bundled.WrapFS(osvfs.FS()),
		cwd:         util.Must(os.Getwd()),
	}
}

func (s *LSPServer) Handle(requestCtx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	// FIXME: implement cancel logic
	ctx := core.WithRequestID(requestCtx, req.ID.String())

	s.conn = conn
	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, req)
	case "initialized":
		return s.handleInitialized(ctx, req)
	case "textDocument/didOpen":
		if s.rslintConfig == nil {
			return nil, nil
		}
		s.handleDidOpen(ctx, req)
		return nil, nil
	case "textDocument/didChange":
		if s.rslintConfig == nil {
			return nil, nil
		}
		s.handleDidChange(ctx, req)
		return nil, nil
	case "textDocument/didSave":
		if s.rslintConfig == nil {
			return nil, nil
		}
		s.handleDidSave(ctx, req)
		return nil, nil
	case "textDocument/diagnostic":
		if s.rslintConfig == nil {
			return nil, nil
		}
		s.handleDidSave(ctx, req)
		return nil, nil
	case "textDocument/codeAction":
		if s.rslintConfig == nil {
			return nil, nil
		}
		return s.handleCodeAction(ctx, req)
	case "shutdown":
		return s.handleShutdown(ctx, req)

	case "exit":
		os.Exit(0)
		return nil, nil
	default:
		// Respond with method not found for unhandled methods
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: "method not found: " + req.Method,
		}
	}
}
