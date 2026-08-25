package lsp

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/project/logging"
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
	// didChangeWatchedFiles. This allows Session to register tsconfig watchers
	// and call RefreshDiagnostics when project state changes.
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
		gitignoreWatchers := gitignoreFileWatchers(s.cwd, relativePatterns)
		ancestorConfigWatchers := ancestorJSConfigFileWatchers(s.cwd, relativePatterns)
		go func() {
			if err := s.WatchFiles(s.backgroundCtx, gitignoreWatcherID, gitignoreWatchers); err != nil {
				if s.backgroundCtx.Err() == nil {
					log.Printf("[rslint] Failed to register .gitignore watchers: %v", err)
				}
			}
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

	s.session = project.NewSession(&project.SessionInit{
		BackgroundCtx: s.backgroundCtx,
		Options: &project.SessionOptions{
			CurrentDirectory:   s.cwd,
			DefaultLibraryPath: s.defaultLibraryPath,
			TypingsLocation:    s.typingsLocation,
			PositionEncoding:   s.positionEncoding,
			WatchEnabled:       s.watchEnabled,
		},
		FS:         s.fs,
		Client:     s,
		Logger:     logging.NewLogger(io.Discard),
		ParseCache: s.parseCache,
	})
	if s.watchEnabled && s.outgoingQueue != nil {
		s.lintPrograms = newLintProgramStore(s)
	}

	// Load the JSON config used before the first JS/TS catalog transaction and
	// as fallback for files outside every discovered JS/TS config boundary.
	rslintConfigPath, configFound := findRslintConfig(s.fs, s.cwd)
	if configFound {
		s.rslintConfigPath = rslintConfigPath
		if err := s.reloadConfig(); err != nil {
			return err
		}
	}

	return nil
}
