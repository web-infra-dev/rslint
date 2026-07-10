package lsp

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/project/logging"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const codeActionKindSourceFixAllRslint = lsproto.CodeActionKind("source.fixAll.rslint")

type lintPassResult struct {
	Diagnostics     []rule.RuleDiagnostic
	HasSyntaxErrors bool
}

// ruleFixToTextEdit converts a rule fix into an LSP TextEdit using the
// source file's line map for position encoding.
func ruleFixToTextEdit(sourceFile ast.SourceFileLike, fix rule.RuleFix) *lsproto.TextEdit {
	startLine, startChar := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, fix.Range.Pos())
	endLine, endChar := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, fix.Range.End())
	return &lsproto.TextEdit{
		Range: lsproto.Range{
			Start: lsproto.Position{Line: uint32(startLine), Character: uint32(startChar)},
			End:   lsproto.Position{Line: uint32(endLine), Character: uint32(endChar)},
		},
		NewText: fix.Text,
	}
}

func (s *Server) handleInitialize(ctx context.Context, params *lsproto.InitializeParams) (lsproto.InitializeResponse, error) {
	log.Printf("handle initialize with pid: %d\n", os.Getpid())
	if s.initializeParams != nil {
		return nil, lsproto.ErrorCodeInvalidRequest
	}

	s.initializeParams = params

	s.positionEncoding = lsproto.PositionEncodingKindUTF16
	if genCapabilities := s.initializeParams.Capabilities.General; genCapabilities != nil && genCapabilities.PositionEncodings != nil {
		if slices.Contains(*genCapabilities.PositionEncodings, lsproto.PositionEncodingKindUTF8) {
			s.positionEncoding = lsproto.PositionEncodingKindUTF8
		}
	}

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
					Change:    ptrTo(lsproto.TextDocumentSyncKindFull),
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

	config.RegisterAllRules()

	s.session = project.NewSession(&project.SessionInit{
		BackgroundCtx: s.backgroundCtx,
		Options: &project.SessionOptions{
			CurrentDirectory:   s.cwd,
			DefaultLibraryPath: s.defaultLibraryPath,
			TypingsLocation:    s.typingsLocation,
			PositionEncoding:   lsproto.PositionEncodingKindUTF8,
			WatchEnabled:       s.watchEnabled,
		},
		FS:         s.fs,
		Client:     s,
		Logger:     logging.NewLogger(io.Discard),
		ParseCache: s.parseCache,
	})

	// Populate the global rule registry once per process; the LSP request path
	// resolves rule names against it after config merging.
	config.RegisterAllRules()

	// Try to load JSON config as fallback.
	// If JS/TS configs exist, the VS Code extension will send them via
	// rslint/configUpdate notification, which takes priority per-file.
	rslintConfigPath, configFound := findRslintConfig(s.fs, s.cwd)
	if configFound {
		s.rslintConfigPath = rslintConfigPath
		if err := s.reloadConfig(); err != nil {
			return err
		}
	}

	return nil
}

// reloadConfig loads (or reloads) the rslint JSON configuration from s.rslintConfigPath.
// The LSP reuses projects already loaded by project service and builds an
// isolated overlay Program on demand for a declared custom project. Resolving
// project paths here preserves declaration order and ensures type-aware rules
// run only when the governing config's first containing project supplies type
// information.
func (s *Server) reloadConfig() error {
	loader := config.NewConfigLoader(s.fs, s.cwd)
	rslintConfig, _, err := loader.LoadRslintConfig(s.rslintConfigPath)
	if err != nil {
		return fmt.Errorf("could not load rslint config: %w", err)
	}
	paths, err := s.resolveTsConfigPaths(rslintConfig, s.cwd)
	if err != nil {
		return fmt.Errorf("could not resolve tsconfig paths for %q: %w", s.rslintConfigPath, err)
	}
	s.jsonConfig = rslintConfig
	s.tsConfigPaths = paths
	return nil
}

// loadJSONConfigFallback resolves the complete JSON fallback without mutating
// live server state. Config-update transactions use it before committing an
// explicitly empty JS/TS config catalog.
func (s *Server) loadJSONConfigFallback() (config.RslintConfig, string, []string, error) {
	if s.fs == nil {
		return s.jsonConfig, "", nil, nil
	}

	configPath, found := findRslintConfig(s.fs, s.cwd)
	if !found {
		return config.RslintConfig{}, "", nil, nil
	}

	loader := config.NewConfigLoader(s.fs, s.cwd)
	rslintConfig, _, err := loader.LoadRslintConfig(configPath)
	if err != nil {
		return nil, "", nil, fmt.Errorf("load JSON fallback %q: %w", configPath, err)
	}
	paths, err := s.resolveTsConfigPaths(rslintConfig, s.cwd)
	if err != nil {
		return nil, "", nil, fmt.Errorf("resolve tsconfig paths for JSON fallback %q: %w", configPath, err)
	}
	return rslintConfig, configPath, paths, nil
}

func (s *Server) handleConfigUpdate(ctx context.Context, params any) error {
	// params is raw JSON from the custom notification
	data, err := stdjson.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal config update params: %w", err)
	}

	var payload struct {
		Generation string `json:"generation,omitempty"`
		Configs    []struct {
			ConfigDirectory string              `json:"configDirectory"`
			Entries         config.RslintConfig `json:"entries"`
		} `json:"configs"`
		// EslintPlugins carries the {prefix, ruleNames} metadata for every
		// ESLint plugin mounted across all configs, aggregated by the VS Code
		// extension (same shape the CLI sends as initPayload.EslintPlugins).
		// The live plugin objects stay in Node (the worker re-imports the
		// config); Go only needs the names to register placeholder rules that
		// make `<prefix>/<rule>` resolvable and route to the worker.
		EslintPlugins []config.EslintPluginEntry `json:"eslintPlugins,omitempty"`
	}
	if err := stdjson.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to parse config update: %w", err)
	}

	// Distinguish nil (malformed/missing "configs" field) from an explicitly
	// empty array (all JS configs were deleted — legitimate clear signal).
	// Go JSON: {"configs":[]} → non-nil empty slice; null/{}/missing → nil.
	if payload.Configs == nil {
		log.Printf("[rslint] Config update has no configs field; keeping existing JS configs intact")
		return nil
	}
	seenConfigDirs := make(map[string]string, len(payload.Configs))
	for _, cfg := range payload.Configs {
		if cfg.ConfigDirectory == "" {
			return errors.New("config update contains an empty configDirectory")
		}
		configID := lspFilesystemPathID(uriToPath(lsproto.DocumentUri(cfg.ConfigDirectory)), s.fs)
		if previous, exists := seenConfigDirs[configID]; exists {
			return fmt.Errorf(
				"config update contains duplicate directories %q and %q",
				previous,
				cfg.ConfigDirectory,
			)
		}
		seenConfigDirs[configID] = cfg.ConfigDirectory
		if err := config.ValidateConfig(cfg.Entries); err != nil {
			return fmt.Errorf("invalid config for %q: %w", cfg.ConfigDirectory, err)
		}
	}

	// Resolve every declared project before touching live config state. A bad
	// project path rejects the whole generation, matching the CLI/API and
	// preserving the previous config atomically.
	candidateConfigs := make(map[string]config.RslintConfig, len(payload.Configs))
	candidateTsConfigs := make(map[string][]string, len(payload.Configs))
	for _, cfg := range payload.Configs {
		configDir := uriToPath(lsproto.DocumentUri(cfg.ConfigDirectory))
		paths, err := s.resolveTsConfigPaths(cfg.Entries, configDir)
		if err != nil {
			return fmt.Errorf("resolve tsconfig paths for %q: %w", cfg.ConfigDirectory, err)
		}
		candidateConfigs[cfg.ConfigDirectory] = cfg.Entries
		candidateTsConfigs[cfg.ConfigDirectory] = paths
	}

	candidateJSONConfig := s.jsonConfig
	candidateJSONConfigPath := s.rslintConfigPath
	candidateJSONTsConfigs := s.tsConfigPaths
	if len(payload.Configs) == 0 {
		candidateJSONConfig, candidateJSONConfigPath, candidateJSONTsConfigs, err = s.loadJSONConfigFallback()
		if err != nil {
			return err
		}
	}

	// Invalidate work created from the previous config before committing the new
	// generation. Cancellation is best-effort; docGeneration also prevents an
	// already-completed old result from being merged after this update.
	for uri := range s.documents {
		s.docGeneration[uri]++
		s.cancelInflightPluginDispatch(uri)
		delete(s.diagnostics, uri)
	}

	// Replace all JS configs with the new set (may be empty when all deleted).
	// Keys are URI strings (e.g. "file:///project") sent from VS Code,
	// matching the URI format used throughout the LSP protocol.
	s.jsConfigs = candidateConfigs
	s.tsConfigPathsByConfig = candidateTsConfigs
	s.eslintPluginConfigGeneration = payload.Generation
	if len(payload.Configs) == 0 {
		s.jsonConfig = candidateJSONConfig
		s.rslintConfigPath = candidateJSONConfigPath
		s.tsConfigPaths = candidateJSONTsConfigs
	}
	log.Printf("[rslint] Config updated from JS/TS configs (%d config files)", len(payload.Configs))

	// Register placeholder rules for mounted ESLint plugins so their rule
	// names resolve (and route to the Node worker via IsEslintPluginRule)
	// instead of being silently dropped. Idempotent across config updates;
	// a same-named native rule always wins (RegisterEslintPluginRules skips it).
	// RegisterAllRules already ran in handleInitialized, so native rules exist.
	config.RegisterEslintPluginRules(payload.EslintPlugins)

	// Ask the client to re-pull diagnostics with the updated config.
	if err := s.RefreshDiagnostics(ctx); err != nil {
		log.Printf("[rslint] Failed to refresh diagnostics after config update: %v", err)
	}

	return nil
}

// handleDidChangeWatchedFiles handles file change notifications from the client.
func (s *Server) handleDidChangeWatchedFiles(ctx context.Context, params *lsproto.DidChangeWatchedFilesParams) error {
	if params == nil {
		return nil
	}

	// Forward all file change events to Session so it can detect tsconfig
	// changes, update its internal project state, and trigger
	// RefreshDiagnostics via its background queue.
	if s.session != nil {
		s.session.DidChangeWatchedFiles(ctx, params.Changes)
	}

	// Check for config file changes that affect rslint.
	needsTypeInfoRebuild := false
	for _, change := range params.Changes {
		uri := string(change.Uri)
		if isRslintConfigURI(uri) {
			// rslint config changed — reload config + typeInfoFiles + relint all.
			s.reloadConfigAndRelint()
			return nil
		}
		if isTsConfigURI(uri) {
			needsTypeInfoRebuild = true
		}
	}
	if needsTypeInfoRebuild {
		// tsconfig changed — rebuild tsConfigPaths so type-aware rule filtering
		// stays in sync. Session already handles the project state update and
		// triggers RefreshDiagnostics for relinting.
		if err := s.rebuildTsConfigPaths(); err != nil {
			log.Printf("[rslint] Failed to rebuild tsconfig paths: %v", err)
		}
	}

	return nil
}

// isRslintConfigURI returns true if the URI points to an rslint config file.
func isRslintConfigURI(uri string) bool {
	return strings.HasSuffix(uri, "/rslint.json") || strings.HasSuffix(uri, "/rslint.jsonc")
}

// isTsConfigURI returns true if the URI points to a tsconfig/jsconfig file,
// including variants like tsconfig.build.json, tsconfig.app.json, etc.
func isTsConfigURI(uri string) bool {
	idx := strings.LastIndex(uri, "/")
	if idx < 0 {
		return false
	}
	name := uri[idx+1:]
	return (strings.HasPrefix(name, "tsconfig") || strings.HasPrefix(name, "jsconfig")) &&
		strings.HasSuffix(name, ".json")
}

// resolveTsConfigPaths resolves parserOptions.project from a config and
// normalizes paths with realpath for cross-platform consistency.
func (s *Server) resolveTsConfigPaths(cfg config.RslintConfig, cwd string) ([]string, error) {
	paths, err := config.ResolveTsConfigPaths(cfg, cwd, s.fs)
	if err != nil {
		return nil, err
	}
	for i, p := range paths {
		if realPath := s.fs.Realpath(p); realPath != "" {
			p = realPath
		}
		paths[i] = tspath.NormalizePath(p)
	}
	return paths, nil
}

// rebuildTsConfigPaths resolves parserOptions.project from the current config.
// Called when a tsconfig or rslint config changes so that type-aware rule
// filtering stays in sync.
//
// For JS/TS configs we resolve per-config directory into tsConfigPathsByConfig.
// A config whose parserOptions.project is empty and has no auto-detected
// tsconfig resolves to nil. Files governed by that config have no type info,
// without affecting files governed by other configs. A nested template or
// fixture config without a tsconfig must not change sibling config behavior.
func (s *Server) rebuildTsConfigPaths() error {
	var tsConfigPaths []string
	if s.rslintConfigPath != "" {
		var err error
		tsConfigPaths, err = s.resolveTsConfigPaths(s.jsonConfig, s.cwd)
		if err != nil {
			return fmt.Errorf("resolve tsconfig paths for %q: %w", s.rslintConfigPath, err)
		}
	}

	var byConfig map[string][]string
	if len(s.jsConfigs) > 0 {
		byConfig = make(map[string][]string, len(s.jsConfigs))
		for dir, entries := range s.jsConfigs {
			configDir := uriToPath(lsproto.DocumentUri(dir))
			paths, err := s.resolveTsConfigPaths(entries, configDir)
			if err != nil {
				return fmt.Errorf("resolve tsconfig paths for %q: %w", dir, err)
			}
			byConfig[dir] = paths
		}
	}

	s.tsConfigPaths = tsConfigPaths
	s.tsConfigPathsByConfig = byConfig
	return nil
}

// reloadConfigAndRelint re-discovers and reloads the rslint JSON config, then
// re-lints all open documents. The JSON config remains a live fallback for
// files that have no JS/TS config ancestor, so it must stay current even while
// one or more JS/TS configs are active.
func (s *Server) reloadConfigAndRelint() {
	log.Printf("Reloading rslint config...")

	configPath, found := findRslintConfig(s.fs, s.cwd)
	if !found {
		log.Printf("rslint config file no longer exists, clearing config")
		s.jsonConfig = config.RslintConfig{}
		s.rslintConfigPath = ""
		s.tsConfigPaths = nil
	} else {
		previousPath := s.rslintConfigPath
		s.rslintConfigPath = configPath
		if err := s.reloadConfig(); err != nil {
			s.rslintConfigPath = previousPath
			log.Printf("Error reloading rslint config: %v", err)
			return
		}
	}

	for uri := range s.documents {
		s.pushDiagnostics(uri)
	}
}

// lintDebounceDelay is how long to wait after the last keystroke before
// running the linter. This avoids linting on every keystroke against
// incomplete/broken syntax that can cause panics or waste CPU.
const lintDebounceDelay = 200 * time.Millisecond

// scheduleLint marks a URI for deferred linting and resets the debounce timer.
// When the timer fires it signals debounceCh, which is consumed by the main
// dispatch loop. Must be called from the main dispatch loop goroutine.
func (s *Server) scheduleLint(uri lsproto.DocumentUri) {
	s.pendingLintURIs[uri] = struct{}{}
	if s.lintTimer != nil {
		s.lintTimer.Stop()
	}
	s.lintTimer = time.AfterFunc(lintDebounceDelay, func() {
		select {
		case s.debounceCh <- struct{}{}:
		default:
			// Already pending — no need to queue another signal
		}
	})
}

func (s *Server) handleDidOpen(ctx context.Context, params *lsproto.DidOpenTextDocumentParams) error {
	log.Printf("Handling didOpen: %s", params.TextDocument.Uri)

	uri := params.TextDocument.Uri
	content := params.TextDocument.Text

	s.documents[uri] = content

	// Notify session about the opened file so it creates the overlay
	if s.session != nil {
		s.session.DidOpenFile(ctx, uri, params.TextDocument.Version, content, params.TextDocument.LanguageId)
		s.pushDiagnostics(uri)
	}
	return nil
}

func (s *Server) handleDidChange(ctx context.Context, params *lsproto.DidChangeTextDocumentParams) error {
	log.Printf("Handling didChange: %s (version %d)", params.TextDocument.Uri, params.TextDocument.Version)

	uri := params.TextDocument.Uri

	// For full document sync, we expect one change with the full text
	if len(params.ContentChanges) > 0 {
		s.documents[uri] = params.ContentChanges[0].WholeDocument.Text
	}

	// Notify session immediately so tsgo's overlay stays up-to-date for
	// other LSP features (completions, hover, etc.).  Lint is deferred
	// via scheduleLint to avoid running the linter on every keystroke.
	if s.session != nil {
		s.session.DidChangeFile(ctx, uri, params.TextDocument.Version, params.ContentChanges)
		s.scheduleLint(uri)
	}
	return nil
}

func (s *Server) handleDidSave(ctx context.Context, params *lsproto.DidSaveTextDocumentParams) error {
	log.Printf("Handling didSave: %s", params.TextDocument.Uri)
	uri := params.TextDocument.Uri

	// didChange is authoritative for the current content of an open document.
	// didSave may include the text that reached disk, but carries no document
	// version, so a save for an older buffer can arrive after a newer didChange.
	// Never replace the versioned document mirror with this unversioned snapshot.
	currentContent, open := s.documents[uri]
	forwardSave := shouldForwardDidSave(currentContent, open, params.Text)
	if !forwardSave {
		log.Printf("Ignoring stale didSave for open document %s", uri)
	}

	// Clear pending debounce lint for this URI — pushDiagnostics below
	// will lint it immediately, so the debounce would be redundant.
	delete(s.pendingLintURIs, uri)

	// Notify session about the save event
	if s.session != nil {
		if forwardSave {
			s.session.DidSaveFile(ctx, uri)
		}
		s.pushDiagnostics(uri)
	}
	return nil
}

// shouldForwardDidSave suppresses only saves that are known to describe an
// older version of an open document. Saves without text and saves for documents
// not tracked as open are forwarded for LSP client compatibility and so tsgo
// can observe out-of-band disk changes.
func shouldForwardDidSave(currentContent string, open bool, savedText *string) bool {
	return savedText == nil || !open || currentContent == *savedText
}

func (s *Server) handleDidClose(ctx context.Context, params *lsproto.DidCloseTextDocumentParams) error {
	log.Printf("Handling didClose: %s", params.TextDocument.Uri)
	uri := params.TextDocument.Uri
	delete(s.documents, uri)
	delete(s.diagnostics, uri)
	delete(s.pendingLintURIs, uri)
	// Bump (do NOT delete) the generation on close so any in-flight plugin
	// result for this URI is stale. Keeping the counter monotonic — rather
	// than resetting it to 0 on a later reopen — prevents a generation collision
	// where a pre-close worker result could match a freshly reopened document.
	s.docGeneration[uri]++
	// Cancel an in-flight plugin dispatch for the closed doc so its Node worker
	// stops instead of running to completion — no superseding keystroke will.
	s.cancelInflightPluginDispatch(uri)

	if s.session != nil {
		// Push empty diagnostics to clear the client's display before closing
		if err := s.PublishDiagnostics(ctx, &lsproto.PublishDiagnosticsParams{
			Uri:         uri,
			Diagnostics: []*lsproto.Diagnostic{},
		}); err != nil {
			log.Printf("Error clearing diagnostics on close: %v", err)
		}
		s.session.DidCloseFile(ctx, uri)
	}
	return nil
}

func (s *Server) handleCodeAction(ctx context.Context, params *lsproto.CodeActionParams) (lsproto.CodeActionResponse, error) {
	log.Printf("Handling codeAction: %+v,%+v", params, ctx)
	uri := params.TextDocument.Uri

	// Handle source.fixAll requests (triggered by editor.codeActionsOnSave)
	if isFixAllRequest(params.Context) {
		return s.handleFixAllCodeAction(ctx, uri)
	}

	// Get stored diagnostics for this document
	ruleDiagnostics, exists := s.diagnostics[uri]
	if !exists {
		return lsproto.CodeActionResponse{
			CommandOrCodeActionArray: &[]lsproto.CommandOrCodeAction{},
		}, nil
	}

	var codeActions []lsproto.CommandOrCodeAction

	// Find diagnostics that overlap with the requested range
	for _, ruleDiag := range ruleDiagnostics {
		// Check if diagnostic range overlaps with requested range
		diagStartLine, diagStartChar := scanner.GetECMALineAndUTF16CharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.Pos())
		diagEndLine, diagEndChar := scanner.GetECMALineAndUTF16CharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.End())

		diagRange := lsproto.Range{
			Start: lsproto.Position{Line: uint32(diagStartLine), Character: uint32(diagStartChar)},
			End:   lsproto.Position{Line: uint32(diagEndLine), Character: uint32(diagEndChar)},
		}

		if rangesOverlap(diagRange, params.Range) {
			// Add code action for fixes
			codeAction := createCodeActionFromRuleDiagnostic(ruleDiag, uri)
			if codeAction != nil {
				codeActions = append(codeActions, lsproto.CommandOrCodeAction{
					Command:    nil,
					CodeAction: codeAction,
				})
			}
			// add extract disable rule actions
			disableActions := createDisableRuleActions(ruleDiag, uri)
			codeActions = append(codeActions, disableActions...)

			// Add code actions for suggestions
			if ruleDiag.Suggestions != nil {
				for _, suggestion := range *ruleDiag.Suggestions {
					suggestionAction := createCodeActionFromSuggestion(ruleDiag, suggestion, uri)
					if suggestionAction != nil {
						codeActions = append(codeActions, lsproto.CommandOrCodeAction{
							Command:    nil,
							CodeAction: suggestionAction,
						})
					}
				}
			}
		}
	}

	return lsproto.CodeActionResponse{
		CommandOrCodeActionArray: &codeActions,
	}, nil
}

// isFixAllRequest returns true if the code action context requests source.fixAll actions.
func isFixAllRequest(ctx *lsproto.CodeActionContext) bool {
	if ctx == nil || ctx.Only == nil {
		return false
	}
	for _, kind := range *ctx.Only {
		if kind == lsproto.CodeActionKindSourceFixAll || kind == codeActionKindSourceFixAllRslint {
			return true
		}
	}
	return false
}

// maxFixPasses is the maximum number of lint-fix cycles to prevent infinite loops
// when two rules produce fixes that undo each other.
const maxFixPasses = 10

// handleFixAllCodeAction computes all auto-fixes for the given URI using
// multi-pass fixing: each pass lints → applies fixes → updates the session
// overlay, repeating until no more fixes are found or maxFixPasses is reached.
// This handles cascading fixes (e.g. no-wrapper-object-types fix triggers no-inferrable-types).
// It does NOT push diagnostics or update s.diagnostics — that is left to the
// subsequent didSave handler in the normal save flow.
func (s *Server) handleFixAllCodeAction(ctx context.Context, uri lsproto.DocumentUri) (lsproto.CodeActionResponse, error) {
	empty := lsproto.CodeActionResponse{CommandOrCodeActionArray: &[]lsproto.CommandOrCodeAction{}}

	// Clear pending debounce for this URI — we are about to lint it fresh,
	// so any scheduled debounce lint for the same content is redundant.
	delete(s.pendingLintURIs, uri)

	if s.session == nil {
		return empty, nil
	}
	if !isLintableScriptFile(uri) {
		return empty, nil
	}

	rslintConfig, configCwd, isJSConfig := s.getConfigForURI(uri)
	tsConfigPaths := s.tsConfigPathsForURI(uri)
	originalContent := s.documents[uri]

	currentContent := s.computeFixAllContent(ctx, uri, originalContent, rslintConfig, configCwd, isJSConfig, tsConfigPaths)

	if currentContent == originalContent {
		return empty, nil
	}

	// Produce a single TextEdit that replaces the entire document content.
	// Individual per-fix TextEdits can't be composed across passes (offsets shift),
	// so we replace the whole document with the final result.
	lastLine, lastChar := computeEndPosition(originalContent)

	codeAction := &lsproto.CodeAction{
		Title: "Fix all rslint auto-fixable problems",
		Kind:  ptrTo(codeActionKindSourceFixAllRslint),
		Edit: &lsproto.WorkspaceEdit{
			Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
				uri: {
					{
						Range: lsproto.Range{
							Start: lsproto.Position{Line: 0, Character: 0},
							End:   lsproto.Position{Line: uint32(lastLine), Character: uint32(lastChar)},
						},
						NewText: currentContent,
					},
				},
			},
		},
	}

	return lsproto.CodeActionResponse{
		CommandOrCodeActionArray: &[]lsproto.CommandOrCodeAction{
			{CodeAction: codeAction},
		},
	}, nil
}

// computeFixAllContent runs the multi-pass lint→fix loop and returns the final
// fixed content (== originalContent when nothing changed). Each pass folds
// eslint-plugin fixes into the native fixes so source.fixAll applies both, on
// the SAME content (so their byte offsets align with ApplyRuleFixes's input).
// The per-pass native lint goes through s.fixAllNativeLint, which tests override
// to drive the fold loop without a real TS session.
func (s *Server) computeFixAllContent(ctx context.Context, uri lsproto.DocumentUri, originalContent string, rslintConfig config.RslintConfig, configCwd string, isJSConfig bool, tsConfigPaths []string) string {
	nativeLint := s.fixAllNativeLint
	if nativeLint == nil {
		nativeLint = s.defaultFixAllNativeLint
	}

	// Bound the eslint-plugin reverse requests across the WHOLE fixAll, not per
	// pass: source.fixAll runs inline on the dispatch loop, so a wedged or
	// mid-rebuild client that never answers rslint/pluginLint must not
	// freeze editor interaction — nor multiply the stall by maxFixPasses. Only
	// the plugin pass gets this deadline; the native pass keeps the original ctx
	// (it is in-process and does not depend on a client reply). Once the budget
	// expires lintPluginRulesSync returns nil and the remaining passes fold
	// native-only fixes.
	pluginTimeout := s.pluginReverseTimeout
	if pluginTimeout <= 0 {
		pluginTimeout = defaultPluginReverseTimeout
	}
	pluginCtx, cancelPlugin := context.WithTimeout(ctx, pluginTimeout)
	defer cancelPlugin()

	currentContent := originalContent
	for pass := range maxFixPasses {
		lintResult, err := nativeLint(ctx, uri, pass, currentContent, rslintConfig, configCwd, isJSConfig, tsConfigPaths)
		if err != nil {
			log.Printf("Error running lint for fixAll pass %d: %v", pass, err)
			break
		}
		if lintResult.HasSyntaxErrors {
			break
		}
		ruleDiags := lintResult.Diagnostics

		// Fold in eslint-plugin fixes so source.fixAll applies plugin rule fixes
		// too, not just native. The plugin pass lints the SAME currentContent, so
		// its fix byte offsets align with ApplyRuleFixes's input; suggestionsMode
		// is "off" because fixAll applies only autofixes.
		// Skip the plugin pass once the budget is spent: lintPluginRulesSync on an
		// already-expired pluginCtx would still enqueue a (wasted) reverse request
		// to the client before returning nil.
		if pluginCtx.Err() == nil {
			if pluginDiags := s.lintPluginRulesSync(pluginCtx, uri, currentContent, true, linter.SuggestionsModeOff); len(pluginDiags) > 0 {
				ruleDiags = append(ruleDiags, pluginDiags...)
			}
		}

		fixedContent, _, wasFixed := linter.ApplyRuleFixes(currentContent, ruleDiags)
		if !wasFixed {
			break
		}
		currentContent = fixedContent
		if currentContent == originalContent {
			break // cycle detected — fixes reverted to original content
		}
	}
	return currentContent
}

// defaultFixAllNativeLint builds each pass from an isolated editor overlay.
// The pass number is intentionally unused: speculative content never enters
// the real TypeScript Session, regardless of how many fix cycles run.
func (s *Server) defaultFixAllNativeLint(ctx context.Context, uri lsproto.DocumentUri, _ int, content string, rslintConfig config.RslintConfig, configCwd string, isJSConfig bool, tsConfigPaths []string) (lintPassResult, error) {
	return s.runConfiguredLintForContent(uri, ctx, content, rslintConfig, configCwd, isJSConfig, tsConfigPaths)
}

// computeEndPosition returns the line and UTF-16 character offset of the end
// of a text string, suitable for constructing an LSP Range that covers the
// entire document. Uses core.UTF16Len for correct UTF-16 code unit counting.
func computeEndPosition(text string) (int, int) {
	line := 0
	lastLineStart := 0
	for i := range len(text) {
		if text[i] == '\n' {
			line++
			lastLineStart = i + 1
		}
	}
	return line, int(core.UTF16Len(text[lastLineStart:]))
}

func convertRuleDiagnosticToLSP(ruleDiag rule.RuleDiagnostic) *lsproto.Diagnostic {
	diagnosticStart := ruleDiag.Range.Pos()
	diagnosticEnd := ruleDiag.Range.End()
	startLine, startColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(ruleDiag.SourceFile, diagnosticStart)
	endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(ruleDiag.SourceFile, diagnosticEnd)

	return &lsproto.Diagnostic{
		Range: lsproto.Range{
			Start: lsproto.Position{
				Line:      uint32(startLine),
				Character: uint32(startColumn),
			},
			End: lsproto.Position{
				Line:      uint32(endLine),
				Character: uint32(endColumn),
			},
		},
		Severity: ptrTo(lsproto.DiagnosticSeverity(ruleDiag.Severity.Int())),
		Source:   ptrTo("rslint"),
		Message:  lsproto.StringOrMarkupContent{String: ptrTo(fmt.Sprintf("[%s] %s", ruleDiag.RuleName, ruleDiag.Message.Description))},
	}
}

func isLintableScriptFile(uri lsproto.DocumentUri) bool {
	return config.IsSupportedLintFile(uriToPath(uri))
}

func uriToPath(uri lsproto.DocumentUri) string {
	// Convert file:// URI to file path using net/url for proper percent-decoding.
	// Handles spaces (%20), CJK characters, and other encoded chars in paths.
	// file:///home/user       → /home/user  (Unix)
	// file:///C:/Users        → C:/Users    (Windows — strip the leading slash)
	// file:///path%20name/f   → /path name/f
	uriStr := string(uri)
	if uriStr == "" {
		return ""
	}
	u, err := url.ParseRequestURI(uriStr)
	if err != nil {
		return uriStr // fallback: return as-is for non-URI strings
	}
	p := u.Path
	if u.Host != "" {
		return "//" + u.Host + p
	}
	// Windows drive letter: /C:/... → C:/...
	if len(p) >= 3 && p[0] == '/' && unicode.IsLetter(rune(p[1])) && p[2] == ':' {
		return p[1:]
	}
	return p
}

func findRslintConfig(fs vfs.FS, workingDir string) (string, bool) {
	defaultConfigs := []string{"rslint.json", "rslint.jsonc"}

	// Strategy 1: Try in the working directory
	for _, configName := range defaultConfigs {
		configPath := filepath.Join(workingDir, configName)
		if fs.FileExists(configPath) {
			return configPath, true
		}
	}
	return "", false
}

// LintResponse represents a lint response from Go to JS
type LintResponse struct {
	Diagnostics []lsproto.Diagnostic `json:"diagnostics"`
	ErrorCount  int                  `json:"errorCount"`
	FileCount   int                  `json:"fileCount"`
	RuleCount   int                  `json:"ruleCount"`
}

type lintProgramLoader func(tsConfigPath string) (*compiler.Program, error)

func runLintWithSession(uri lsproto.DocumentUri, session *project.Session, ctx context.Context, rslintConfig config.RslintConfig, cwd string, enforcePlugins bool, tsConfigPaths []string, fs vfs.FS) ([]rule.RuleDiagnostic, error) {
	result, err := runLintWithProgramLoader(uri, session, ctx, rslintConfig, cwd, enforcePlugins, tsConfigPaths, fs, nil)
	return result.Diagnostics, err
}

func runLintWithProgramLoader(
	uri lsproto.DocumentUri,
	session *project.Session,
	ctx context.Context,
	rslintConfig config.RslintConfig,
	cwd string,
	enforcePlugins bool,
	tsConfigPaths []string,
	fs vfs.FS,
	loadProgram lintProgramLoader,
) (lintPassResult, error) {
	filename := uriToPath(uri)

	// Files excluded by the config's `ignores` patterns produce no diagnostics,
	// matching CLI behavior. Return early before spinning up the language service.
	if rslintConfig.IsFileIgnored(filename, cwd) {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}}, nil
	}
	fileConfigResolver := config.NewFileConfigResolver(rslintConfig, cwd, enforcePlugins)
	// Files outside the config's `files` set should not spin up TypeScript LS.
	if fileConfigResolver.ConfigForFile(filename) == nil {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}}, nil
	}

	program, hasTypeInfo, err := selectLintProgram(uri, session, ctx, tsConfigPaths, fs, loadProgram)
	if err != nil {
		return lintPassResult{}, err
	}
	return lintSingleFile(program, filename, hasTypeInfo, fileConfigResolver, ctx, fs), nil
}

func lintSingleFile(
	program *compiler.Program,
	filename string,
	hasTypeInfo bool,
	fileConfigResolver *config.FileConfigResolver,
	ctx context.Context,
	fs vfs.FS,
) lintPassResult {
	sourceFile := sourceFileForPath(program, filename, fs)
	if sourceFile == nil {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}}
	}
	if len(program.GetSyntacticDiagnostics(ctx, sourceFile)) > 0 {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}, HasSyntaxErrors: true}
	}

	// Collect diagnostics
	var diagnostics []rule.RuleDiagnostic
	var diagnosticsLock sync.Mutex

	// Create collector function
	diagnosticCollector := func(d rule.RuleDiagnostic) {
		diagnosticsLock.Lock()
		defer diagnosticsLock.Unlock()
		diagnostics = append(diagnostics, d)
	}

	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     program,
		File:        sourceFile.FileName(),
		HasTypeInfo: hasTypeInfo,
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return fileConfigResolver.ActiveRulesForFileHasTypeInfo(filename, hasTypeInfo)
		},
		OnDiagnostic: diagnosticCollector,
	})

	if diagnostics == nil {
		diagnostics = []rule.RuleDiagnostic{}
	}
	return lintPassResult{Diagnostics: diagnostics}
}

func selectLintProgram(
	uri lsproto.DocumentUri,
	session *project.Session,
	ctx context.Context,
	tsConfigPaths []string,
	fs vfs.FS,
	loadProgram lintProgramLoader,
) (*compiler.Program, bool, error) {
	filename := uriToPath(uri)
	// Flush pending document changes and collect every already-loaded project
	// containing the file. The default language service remains the AST-only
	// fallback when none of the config's declared projects contains the file.
	_, languageService, loadedProjects, err := session.GetLanguageServiceAndProjectsForFile(ctx, uri)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get language service: %w", err)
	}
	program := languageService.GetProgram()

	// Type information follows parserOptions.project declaration order, not the
	// TypeScript session's default-project heuristic. Prefer an already-loaded
	// containing project. Custom config names that the project service has not
	// loaded are supplied by an isolated standalone Program; this avoids
	// mutating the Session's permanent API-open project set.
	loadedByConfig := make(map[string]*compiler.Program, len(loadedProjects))
	for _, candidate := range loadedProjects {
		if candidate == nil || candidate.GetProgram() == nil || !programContainsFile(candidate.GetProgram(), filename, fs) {
			continue
		}
		loadedByConfig[lspFilesystemPathID(string(candidate.Id()), fs)] = candidate.GetProgram()
	}
	for _, tsConfigPath := range tsConfigPaths {
		configID := lspFilesystemPathID(tsConfigPath, fs)
		if loadedProgram := loadedByConfig[configID]; loadedProgram != nil {
			return loadedProgram, true, nil
		}
		if loadProgram == nil {
			continue
		}
		candidate, loadErr := loadProgram(tsConfigPath)
		if loadErr != nil {
			return nil, false, fmt.Errorf("load configured project %q: %w", tsConfigPath, loadErr)
		}
		if programContainsFile(candidate, filename, fs) {
			return candidate, true, nil
		}
	}
	return program, false, nil
}

func programContainsFile(program *compiler.Program, filename string, fs vfs.FS) bool {
	return sourceFileForPath(program, filename, fs) != nil
}

func sourceFileForPath(program *compiler.Program, filename string, fs vfs.FS) *ast.SourceFile {
	if program == nil {
		return nil
	}
	if sourceFile := program.GetSourceFile(filename); sourceFile != nil {
		return sourceFile
	}
	if fs != nil {
		if realPath := fs.Realpath(filename); realPath != "" && realPath != filename {
			return program.GetSourceFile(realPath)
		}
	}
	return nil
}

func (s *Server) currentEditorOverlayFS() vfs.FS {
	files := make(map[string]string, len(s.documents))
	for uri, content := range s.documents {
		files[tspath.NormalizePath(uriToPath(uri))] = content
	}
	return utils.NewOverlayVFS(s.fs, files)
}

func (s *Server) newStandaloneLintProgramLoader() lintProgramLoader {
	var overlayFS vfs.FS
	return func(tsConfigPath string) (*compiler.Program, error) {
		if overlayFS == nil {
			overlayFS = s.currentEditorOverlayFS()
		}
		return createStandaloneLintProgram(tsConfigPath, overlayFS)
	}
}

func createStandaloneLintProgram(tsConfigPath string, fs vfs.FS) (*compiler.Program, error) {
	configDir := tspath.GetDirectoryPath(tspath.NormalizePath(tsConfigPath))
	host := utils.CreateCompilerHost(configDir, fs)
	return utils.CreateProgramLenient(true, fs, configDir, tsConfigPath, host)
}

func createStandaloneFallbackProgram(filename string, cwd string, fs vfs.FS) (*compiler.Program, error) {
	host := utils.CreateCompilerHost(cwd, fs)
	return utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Target:    core.ScriptTargetESNext,
		Module:    core.ModuleKindESNext,
		Jsx:       core.JsxEmitPreserve,
		AllowJs:   core.TSTrue,
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, []string{filename}, host)
}

func (s *Server) runConfiguredLint(
	uri lsproto.DocumentUri,
	ctx context.Context,
	rslintConfig config.RslintConfig,
	cwd string,
	enforcePlugins bool,
	tsConfigPaths []string,
) (lintPassResult, error) {
	return runLintWithProgramLoader(
		uri,
		s.session,
		ctx,
		rslintConfig,
		cwd,
		enforcePlugins,
		tsConfigPaths,
		s.fs,
		s.newStandaloneLintProgramLoader(),
	)
}

// runConfiguredLintForContent lints a speculative fix pass against an
// isolated overlay. It never mutates the TypeScript Session's open document,
// so cancelling or declining a code action cannot change later LSP results.
func (s *Server) runConfiguredLintForContent(
	uri lsproto.DocumentUri,
	ctx context.Context,
	content string,
	rslintConfig config.RslintConfig,
	cwd string,
	enforcePlugins bool,
	tsConfigPaths []string,
) (lintPassResult, error) {
	filename := tspath.NormalizePath(uriToPath(uri))
	if rslintConfig.IsFileIgnored(filename, cwd) {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}}, nil
	}
	resolver := config.NewFileConfigResolver(rslintConfig, cwd, enforcePlugins)
	if resolver.ConfigForFile(filename) == nil {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}}, nil
	}

	files := make(map[string]string, len(s.documents)+1)
	for documentURI, documentContent := range s.documents {
		files[tspath.NormalizePath(uriToPath(documentURI))] = documentContent
	}
	files[filename] = content
	overlayFS := utils.NewOverlayVFS(s.fs, files)

	for _, tsConfigPath := range tsConfigPaths {
		program, err := createStandaloneLintProgram(tsConfigPath, overlayFS)
		if err != nil {
			return lintPassResult{}, fmt.Errorf("load configured project %q: %w", tsConfigPath, err)
		}
		if programContainsFile(program, filename, overlayFS) {
			return lintSingleFile(program, filename, true, resolver, ctx, overlayFS), nil
		}
	}

	program, err := createStandaloneFallbackProgram(filename, cwd, overlayFS)
	if err != nil {
		return lintPassResult{}, fmt.Errorf("create fallback lint program: %w", err)
	}
	return lintSingleFile(program, filename, false, resolver, ctx, overlayFS), nil
}

func lspFilesystemPathID(filePath string, fs vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	useCaseSensitive := true
	if fs != nil {
		useCaseSensitive = fs.UseCaseSensitiveFileNames()
		if realPath := fs.Realpath(filePath); realPath != "" {
			filePath = tspath.NormalizePath(realPath)
		}
	}
	return string(tspath.ToPath(filePath, "", useCaseSensitive))
}

func lspActiveRulesForFile(rslintConfig config.RslintConfig, filePath string, cwd string, enforcePlugins bool, hasTypeInfo bool) []linter.ConfiguredRule {
	return config.NewFileConfigResolver(rslintConfig, cwd, enforcePlugins).
		ActiveRulesForFileHasTypeInfo(filePath, hasTypeInfo)
}

// Helper function to check if two ranges overlap
func rangesOverlap(a, b lsproto.Range) bool {
	// Ranges overlap if a starts before or at b's end AND b starts before or at a's end
	aStartsBefore := a.Start.Line < b.End.Line ||
		(a.Start.Line == b.End.Line && a.Start.Character <= b.End.Character)
	bStartsBefore := b.Start.Line < a.End.Line ||
		(b.Start.Line == a.End.Line && b.Start.Character <= a.End.Character)

	return aStartsBefore && bStartsBefore
}

// Helper function to create a code action from a rule diagnostic
func createCodeActionFromRuleDiagnostic(ruleDiag rule.RuleDiagnostic, uri lsproto.DocumentUri) *lsproto.CodeAction {
	fixes := ruleDiag.Fixes()
	if len(fixes) == 0 {
		return nil
	}

	// Convert rule fixes to LSP text edits
	var textEdits []*lsproto.TextEdit
	for _, fix := range fixes {
		textEdits = append(textEdits, ruleFixToTextEdit(ruleDiag.SourceFile, fix))
	}

	// Create workspace edit
	workspaceEdit := &lsproto.WorkspaceEdit{
		Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
			uri: textEdits,
		},
	}

	return &lsproto.CodeAction{
		Title:       "Fix: " + ruleDiag.Message.Description,
		Kind:        ptrTo(lsproto.CodeActionKind("quickfix")),
		Edit:        workspaceEdit,
		Diagnostics: &[]*lsproto.Diagnostic{convertRuleDiagnosticToLSP(ruleDiag)},
		IsPreferred: ptrTo(true), // Mark auto-fixes as preferred
	}
}

// Helper function to create a code action from a rule suggestion
func createCodeActionFromSuggestion(ruleDiag rule.RuleDiagnostic, suggestion rule.RuleSuggestion, uri lsproto.DocumentUri) *lsproto.CodeAction {
	fixes := suggestion.Fixes()
	if len(fixes) == 0 {
		return nil
	}

	// Convert rule fixes to LSP text edits
	var textEdits []*lsproto.TextEdit
	for _, fix := range fixes {
		textEdits = append(textEdits, ruleFixToTextEdit(ruleDiag.SourceFile, fix))
	}

	// Create workspace edit
	workspaceEdit := &lsproto.WorkspaceEdit{
		Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
			uri: textEdits,
		},
	}

	return &lsproto.CodeAction{
		Title:       "Suggestion: " + suggestion.Message.Description,
		Kind:        ptrTo(lsproto.CodeActionKind("quickfix")),
		Edit:        workspaceEdit,
		Diagnostics: &[]*lsproto.Diagnostic{convertRuleDiagnosticToLSP(ruleDiag)},
		IsPreferred: ptrTo(false), // Mark suggestions as not preferred
	}
}

// Helper function to create disable rule actions for diagnostics without fixes
func createDisableRuleActions(ruleDiag rule.RuleDiagnostic, uri lsproto.DocumentUri) []lsproto.CommandOrCodeAction {
	var actions []lsproto.CommandOrCodeAction

	lspDiagnostic := convertRuleDiagnosticToLSP(ruleDiag)

	// Action 1: Disable rule for this line
	disableLineAction := createDisableRuleForLineAction(ruleDiag, uri, lspDiagnostic)
	if disableLineAction != nil {
		actions = append(actions, lsproto.CommandOrCodeAction{
			Command:    nil,
			CodeAction: disableLineAction,
		})
	}

	// Action 2: Disable rule for entire file
	disableFileAction := createDisableRuleForFileAction(ruleDiag, uri, lspDiagnostic)
	if disableFileAction != nil {
		actions = append(actions, lsproto.CommandOrCodeAction{
			Command:    nil,
			CodeAction: disableFileAction,
		})
	}

	return actions
}

// Helper function to create a "disable rule for this line" action
func createDisableRuleForLineAction(ruleDiag rule.RuleDiagnostic, uri lsproto.DocumentUri, lspDiagnostic *lsproto.Diagnostic) *lsproto.CodeAction {
	// Get the line where the diagnostic occurs
	lineStart := lspDiagnostic.Range.Start.Line

	// Create text edit to add rslint-disable-next-line comment
	disableComment := fmt.Sprintf("// rslint-disable-next-line %s\n", ruleDiag.RuleName)

	// Find the start of the line to insert the comment
	lineStartPos := lsproto.Position{Line: lineStart, Character: 0}

	textEdit := &lsproto.TextEdit{
		Range: lsproto.Range{
			Start: lineStartPos,
			End:   lineStartPos,
		},
		NewText: disableComment,
	}

	workspaceEdit := &lsproto.WorkspaceEdit{
		Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
			uri: {textEdit},
		},
	}

	return &lsproto.CodeAction{
		Title:       fmt.Sprintf("Disable %s for this line", ruleDiag.RuleName),
		Kind:        ptrTo(lsproto.CodeActionKind("quickfix")),
		Edit:        workspaceEdit,
		Diagnostics: &[]*lsproto.Diagnostic{lspDiagnostic},
		IsPreferred: ptrTo(false),
	}
}

// Helper function to create a "disable rule for entire file" action
func createDisableRuleForFileAction(ruleDiag rule.RuleDiagnostic, uri lsproto.DocumentUri, lspDiagnostic *lsproto.Diagnostic) *lsproto.CodeAction {
	// Create text edit to add rslint-disable comment at the top of the file
	disableComment := fmt.Sprintf("/* rslint-disable %s */\n", ruleDiag.RuleName)

	// Insert at the very beginning of the file
	fileStartPos := lsproto.Position{Line: 0, Character: 0}

	textEdit := &lsproto.TextEdit{
		Range: lsproto.Range{
			Start: fileStartPos,
			End:   fileStartPos,
		},
		NewText: disableComment,
	}

	workspaceEdit := &lsproto.WorkspaceEdit{
		Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
			uri: {textEdit},
		},
	}

	return &lsproto.CodeAction{
		Title:       fmt.Sprintf("Disable %s for entire file", ruleDiag.RuleName),
		Kind:        ptrTo(lsproto.CodeActionKind("quickfix")),
		Edit:        workspaceEdit,
		Diagnostics: &[]*lsproto.Diagnostic{lspDiagnostic},
		IsPreferred: ptrTo(false),
	}
}

// getConfigForURI resolves the rslint config for a given file URI.
// It walks upward from the file's directory looking for the closest
// JS/TS config (matching ESLint v10 flat config behavior).
// Falls back to the JSON config if no JS/TS config matches.
// Returns the config entries, the directory to use as cwd for glob matching,
// and whether the config is from a JS/TS config (for plugin enforcement).
// For JS configs the cwd is the config's own directory (URI → path);
// for the JSON fallback it is s.cwd.
func (s *Server) getConfigForURI(uri lsproto.DocumentUri) (config.RslintConfig, string, bool) {
	if configKey, ok := s.nearestJSConfigKey(uri); ok {
		return s.jsConfigs[configKey], uriToPath(lsproto.DocumentUri(configKey)), true
	}
	return s.jsonConfig, s.cwd, false
}

// nearestJSConfigKey returns the deepest JS/TS config directory containing uri.
// Matching uses filesystem path semantics instead of URI string identity so
// Windows drive letters, UNC authorities, and path casing behave consistently
// with the compiler and config resolver.
func (s *Server) nearestJSConfigKey(uri lsproto.DocumentUri) (string, bool) {
	if len(s.jsConfigs) == 0 {
		return "", false
	}
	useCaseSensitive := true
	if s.fs != nil {
		useCaseSensitive = s.fs.UseCaseSensitiveFileNames()
	}
	filePath := tspath.NormalizePath(uriToPath(uri))
	compareOptions := tspath.ComparePathsOptions{UseCaseSensitiveFileNames: useCaseSensitive}
	bestKey := ""
	bestDir := ""
	for configKey := range s.jsConfigs {
		configDir := tspath.NormalizePath(uriToPath(lsproto.DocumentUri(configKey)))
		if configDir == "" {
			continue
		}
		contains := tspath.ComparePaths(filePath, configDir, compareOptions) == 0 ||
			tspath.StartsWithDirectory(filePath, configDir, useCaseSensitive)
		if contains && len(configDir) > len(bestDir) {
			bestKey = configKey
			bestDir = configDir
		}
	}
	return bestKey, bestKey != ""
}

// tsConfigPathsForURI returns the resolved parserOptions.project tsconfig
// paths for the rslint config that governs the given URI. It walks parents
// the same way getConfigForURI does so a nested config with no tsconfig
// does not affect type-info decisions for sibling configs.
//
// A nil return means the governing config has no resolved tsconfig, so callers
// must disable type-aware rules for this file.
func (s *Server) tsConfigPathsForURI(uri lsproto.DocumentUri) []string {
	if configKey, ok := s.nearestJSConfigKey(uri); ok {
		return s.tsConfigPathsByConfig[configKey]
	}
	return s.tsConfigPaths
}

// uriDirname returns the parent directory of a URI string.
// e.g. "file:///project/src/index.ts" → "file:///project/src"
func uriDirname(uri string) string {
	// Find the last '/' after the scheme (file://)
	idx := strings.LastIndex(uri, "/")
	if idx <= 0 {
		return uri
	}
	// Don't strip past the authority part (file:///)
	if strings.HasPrefix(uri, "file:///") && idx < len("file:///") {
		return uri
	}
	return uri[:idx]
}

// pushDiagnostics runs the linter for the given URI and pushes results to the client.
// Must be called synchronously from the LSP message loop (not from a goroutine)
// because session is not goroutine-safe.
func (s *Server) pushDiagnostics(uri lsproto.DocumentUri) {
	if s.session == nil {
		return
	}

	ctx := s.backgroundCtx

	if !isLintableScriptFile(uri) {
		return
	}

	// Bump this URI's generation BEFORE linting. Native diagnostics are
	// published synchronously below; the eslint-plugin pass runs in a
	// goroutine and stamps its result with this value so the main loop can
	// drop it if a newer keystroke relints in the meantime.
	s.docGeneration[uri]++
	generation := s.docGeneration[uri]

	rslintConfig, configCwd, isJSConfig := s.getConfigForURI(uri)
	tsConfigPaths := s.tsConfigPathsForURI(uri)
	lintResult, err := s.runConfiguredLint(uri, ctx, rslintConfig, configCwd, isJSConfig, tsConfigPaths)
	if err != nil {
		log.Printf("Error running lint for push diagnostics: %v", err)
		return
	}
	ruleDiags := lintResult.Diagnostics

	s.diagnostics[uri] = ruleDiags

	// Must use empty slice (not nil) so JSON serializes as [] instead of null
	lspDiags := make([]*lsproto.Diagnostic, 0, len(ruleDiags))
	for _, d := range ruleDiags {
		lspDiags = append(lspDiags, convertRuleDiagnosticToLSP(d))
	}

	if err := s.PublishDiagnostics(ctx, &lsproto.PublishDiagnosticsParams{
		Uri:         uri,
		Diagnostics: lspDiags,
	}); err != nil {
		log.Printf("Error publishing diagnostics: %v", err)
	}

	// Dispatch eslint-plugin rules off the main loop. The reverse request
	// MUST NOT run synchronously here — it would block the dispatch loop (and
	// thus all editor interaction) until the Node worker replies. Results merge
	// back via pluginResultCh on the main loop (s.diagnostics is lock-free).
	if !lintResult.HasSyntaxErrors {
		s.dispatchPluginLint(uri, generation)
	}
}
