package lsp

import (
	"context"
	"log"
	"os"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
)

func (s *Server) handleInitialize(ctx context.Context, params *lsproto.InitializeParams) (lsproto.InitializeResponse, error) {
	log.Printf("handle initialize with pid: %d\n", os.Getpid())
	if s.initializeParams != nil {
		return nil, lsproto.ErrorCodeInvalidRequest
	}

	s.initializeParams = params

	// rslint diagnostics and edits use VS Code's native UTF-16 coordinates.
	// UTF-16 is mandatory for LSP clients, so selecting it unconditionally keeps
	// incoming incremental edits and outgoing ranges on one encoding contract.
	s.positionEncoding = lsproto.PositionEncodingKindUTF16

	response := &lsproto.InitializeResult{
		ServerInfo: &lsproto.ServerInfo{
			Name:    "typescript-go",
			Version: ptrTo(core.Version()),
		},
		Capabilities: &lsproto.ServerCapabilities{
			PositionEncoding: ptrTo(s.positionEncoding),
			TextDocumentSync: &lsproto.TextDocumentSyncOptionsOrKind{
				Options: &lsproto.TextDocumentSyncOptions{
					OpenClose: ptrTo(true),
					Change:    ptrTo(lsproto.TextDocumentSyncKindIncremental),
					Save: &lsproto.BooleanOrSaveOptions{
						SaveOptions: &lsproto.SaveOptions{
							IncludeText: ptrTo(true),
						},
					},
				},
			},
			CodeActionProvider: &lsproto.BooleanOrCodeActionOptions{
				CodeActionOptions: &lsproto.CodeActionOptions{
					CodeActionKinds: &[]lsproto.CodeActionKind{
						lsproto.CodeActionKindQuickFix,
						lsproto.CodeActionKindSourceFixAll,
						codeActionKindSourceFixAllRslint,
					},
				},
			},
		},
	}

	return response, nil
}
func (s *Server) handleInitialized(ctx context.Context, params *lsproto.InitializedParams) error {
	// Enable file watching if the client supports dynamic registration of
	// didChangeWatchedFiles. Resident Programs use these notifications to
	// invalidate source and project metadata.
	if s.initializeParams.Capabilities != nil &&
		s.initializeParams.Capabilities.Workspace != nil &&
		s.initializeParams.Capabilities.Workspace.DidChangeWatchedFiles != nil &&
		ptrIsTrue(s.initializeParams.Capabilities.Workspace.DidChangeWatchedFiles.DynamicRegistration) {
		s.watchEnabled = true
	}
	if s.watchEnabled && s.outgoingQueue != nil {
		relativePatterns := ptrIsTrue(
			s.initializeParams.Capabilities.Workspace.DidChangeWatchedFiles.RelativePatternSupport,
		)
		ancestorConfigWatchers := ancestorJSConfigFileWatchers(s.cwd, relativePatterns)
		go func() {
			if len(ancestorConfigWatchers) == 0 {
				return
			}
			if err := s.WatchFiles(s.backgroundCtx, ancestorJSConfigWatcherID, ancestorConfigWatchers); err != nil {
				if s.backgroundCtx.Err() == nil {
					log.Printf("[rslint] Failed to register ancestor JS config watchers: %v", err)
				}
			}
		}()
	}

	s.initialized = true
	if s.watchEnabled && s.outgoingQueue != nil {
		s.lintPrograms = newLintProgramStore(s)
	}

	return nil
}
