package lsp

import (
	"context"
	stdjson "encoding/json"
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
)

const codeActionKindSourceFixAllRslint = lsproto.CodeActionKind("source.fixAll.rslint")

// ruleFixToTextEdit converts a rule fix into an LSP TextEdit using the
// source file's line map for position encoding.
func ruleFixToTextEdit(sourceFile *ast.SourceFile, fix rule.RuleFix) *lsproto.TextEdit {
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

	// Register all rules before loading config so that normalizeJSONConfig
	// can inject default core/plugin rules into the registry.
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
// The LSP session discovers tsconfig files on its own via projectService for
// providing type information. However, we still need to know which files are
// covered by parserOptions.project so that type-aware rules (e.g. require-await)
// are only enabled for files in the configured tsconfigs — matching CLI behavior.
func (s *Server) reloadConfig() error {
	loader := config.NewConfigLoader(s.fs, s.cwd)
	rslintConfig, _, err := loader.LoadRslintConfig(s.rslintConfigPath)
	if err != nil {
		return fmt.Errorf("could not load rslint config: %w", err)
	}
	s.jsonConfig = rslintConfig
	s.rebuildTsConfigPaths()
	return nil
}

func (s *Server) handleConfigUpdate(ctx context.Context, params any) error {
	// params is raw JSON from the custom notification
	data, err := stdjson.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal config update params: %w", err)
	}

	var payload struct {
		Configs []struct {
			ConfigDirectory string              `json:"configDirectory"`
			Entries         config.RslintConfig `json:"entries"`
		} `json:"configs"`
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

	// Replace all JS configs with the new set (may be empty when all deleted).
	// Keys are URI strings (e.g. "file:///project") sent from VS Code,
	// matching the URI format used throughout the LSP protocol.
	s.jsConfigs = make(map[string]config.RslintConfig, len(payload.Configs))
	for _, cfg := range payload.Configs {
		s.jsConfigs[cfg.ConfigDirectory] = cfg.Entries
	}
	// Clear the JSON config path so that a subsequent JSON file-watcher event
	// does not silently overwrite the JS/TS configs.
	s.rslintConfigPath = ""
	log.Printf("[rslint] Config updated from JS/TS configs (%d config files)", len(payload.Configs))

	s.rebuildTsConfigPaths()

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
		s.rebuildTsConfigPaths()
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
func (s *Server) resolveTsConfigPaths(cfg config.RslintConfig, cwd string) []string {
	paths, _ := config.ResolveTsConfigPaths(cfg, cwd, s.fs)
	for i, p := range paths {
		paths[i] = tspath.NormalizePath(s.fs.Realpath(p))
	}
	return paths
}

// rebuildTsConfigPaths resolves parserOptions.project from the current config.
// Called when a tsconfig or rslint config changes so that type-aware rule
// filtering stays in sync.
//
// For JS/TS configs we resolve per-config directory into tsConfigPathsByConfig.
// A config whose parserOptions.project is empty and has no auto-detected
// tsconfig resolves to nil — this disables type-aware-rule filtering only for
// files governed by that config, not globally across the workspace. A nested
// template / fixture config without a tsconfig must not relax filtering for
// other configs' files.
func (s *Server) rebuildTsConfigPaths() {
	if len(s.jsConfigs) > 0 {
		byConfig := make(map[string][]string, len(s.jsConfigs))
		for dir, entries := range s.jsConfigs {
			configDir := uriToPath(lsproto.DocumentUri(dir))
			byConfig[dir] = s.resolveTsConfigPaths(entries, configDir)
		}
		s.tsConfigPathsByConfig = byConfig
		s.tsConfigPaths = nil
	} else if s.rslintConfigPath != "" {
		s.tsConfigPaths = s.resolveTsConfigPaths(s.jsonConfig, s.cwd)
		s.tsConfigPathsByConfig = nil
	} else {
		s.tsConfigPaths = nil
		s.tsConfigPathsByConfig = nil
	}
}

// reloadConfigAndRelint re-discovers and reloads the rslint JSON config, then
// re-lints all open documents. Skips when JS/TS configs are active — those
// take priority and are managed by handleConfigUpdate.
func (s *Server) reloadConfigAndRelint() {
	if len(s.jsConfigs) > 0 {
		return
	}
	log.Printf("Reloading rslint config...")

	configPath, found := findRslintConfig(s.fs, s.cwd)
	if !found {
		log.Printf("rslint config file no longer exists, clearing config")
		s.jsonConfig = config.RslintConfig{}
		s.rslintConfigPath = ""
		s.tsConfigPaths = nil
	} else {
		s.rslintConfigPath = configPath
		if err := s.reloadConfig(); err != nil {
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
	if params.Text != nil {
		s.documents[uri] = *params.Text
	}

	// Clear pending debounce lint for this URI — pushDiagnostics below
	// will lint it immediately, so the debounce would be redundant.
	delete(s.pendingLintURIs, uri)

	// Notify session about the save event
	if s.session != nil {
		s.session.DidSaveFile(ctx, uri)
		s.pushDiagnostics(uri)
	}
	return nil
}

func (s *Server) handleDidClose(ctx context.Context, params *lsproto.DidCloseTextDocumentParams) error {
	log.Printf("Handling didClose: %s", params.TextDocument.Uri)
	uri := params.TextDocument.Uri
	delete(s.documents, uri)
	delete(s.diagnostics, uri)
	delete(s.pendingLintURIs, uri)

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
// This handles cascading fixes (e.g. ban-types fix triggers no-inferrable-types).
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
	if !isTypeScriptFile(string(uri)) {
		return empty, nil
	}

	rslintConfig, configCwd, isJSConfig := s.getConfigForURI(uri)
	tsConfigPaths := s.tsConfigPathsForURI(uri)
	originalContent := s.documents[uri]
	currentContent := originalContent

	for pass := range maxFixPasses {
		// For passes after the first, update the session overlay so that
		// runLintWithSession sees the fixed content from the previous pass.
		if pass > 0 {
			s.session.DidChangeFile(ctx, uri, int32(pass), []lsproto.TextDocumentContentChangePartialOrWholeDocument{
				{WholeDocument: &lsproto.TextDocumentContentChangeWholeDocument{Text: currentContent}},
			})
		}

		ruleDiags, err := runLintWithSession(uri, s.session, ctx, rslintConfig, configCwd, isJSConfig, tsConfigPaths, s.fs)
		if err != nil {
			log.Printf("Error running lint for fixAll pass %d: %v", pass, err)
			break
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
		Message:  fmt.Sprintf("[%s] %s", ruleDiag.RuleName, ruleDiag.Message.Description),
	}
}

func isTypeScriptFile(uri string) bool {
	path := strings.ToLower(uri)
	return strings.HasSuffix(path, ".ts") ||
		strings.HasSuffix(path, ".tsx") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".jsx")
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

func runLintWithSession(uri lsproto.DocumentUri, session *project.Session, ctx context.Context, rslintConfig config.RslintConfig, cwd string, enforcePlugins bool, tsConfigPaths []string, fs vfs.FS) ([]rule.RuleDiagnostic, error) {
	filename := uriToPath(uri)

	// Files excluded by the config's `ignores` patterns produce no diagnostics,
	// matching CLI behavior. Return early before spinning up the language service.
	if rslintConfig.IsFileIgnored(filename, cwd) {
		return []rule.RuleDiagnostic{}, nil
	}

	// GetLanguageService flushes any pending changes (from DidChangeFile) and
	// returns a language service whose program reflects the latest overlay content.
	languageService, err := session.GetLanguageService(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get language service: %w", err)
	}
	program := languageService.GetProgram()

	// Determine if this file has type information from the configured tsconfigs.
	// The session's program has a ConfigFilePath (the tsconfig it was created from).
	// If that tsconfig is NOT in parserOptions.project, type-aware rules should
	// be filtered out — matching CLI behavior.
	hasTypeInfo := true
	if tsConfigPaths != nil {
		configFilePath := program.Options().ConfigFilePath
		if configFilePath != "" {
			configFilePath = fs.Realpath(configFilePath)
		}
		programConfig := tspath.NormalizePath(configFilePath)
		hasTypeInfo = false
		for _, tc := range tsConfigPaths {
			if tc == programConfig {
				hasTypeInfo = true
				break
			}
		}
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
		Program: program,
		File:    filename,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			activeRules, _ := config.GlobalRuleRegistry.GetEnabledRules(rslintConfig, sourceFile.FileName(), cwd, enforcePlugins)
			if !hasTypeInfo {
				activeRules = linter.FilterNonTypeAwareRules(activeRules)
			}
			return activeRules
		},
		OnDiagnostic: diagnosticCollector,
	})

	if diagnostics == nil {
		diagnostics = []rule.RuleDiagnostic{}
	}
	return diagnostics, nil
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
	if len(s.jsConfigs) > 0 {
		// Both keys and lookups use URI strings (e.g. "file:///project"),
		// so path separators are always forward slashes — no platform issues.
		dir := uriDirname(string(uri))
		for {
			if cfg, ok := s.jsConfigs[dir]; ok {
				return cfg, uriToPath(lsproto.DocumentUri(dir)), true
			}
			parent := uriDirname(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return s.jsonConfig, s.cwd, false
}

// tsConfigPathsForURI returns the resolved parserOptions.project tsconfig
// paths for the rslint config that governs the given URI. It walks parents
// the same way getConfigForURI does so a nested config with no tsconfig
// does not leak its "allow-all" fallback into sibling configs.
//
// A nil return means the governing config has no resolved tsconfig; callers
// should treat this as "disable type-aware filtering for this file only".
func (s *Server) tsConfigPathsForURI(uri lsproto.DocumentUri) []string {
	if len(s.jsConfigs) > 0 {
		dir := uriDirname(string(uri))
		for {
			if _, ok := s.jsConfigs[dir]; ok {
				return s.tsConfigPathsByConfig[dir]
			}
			parent := uriDirname(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		return nil
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

	if !isTypeScriptFile(string(uri)) {
		return
	}

	rslintConfig, configCwd, isJSConfig := s.getConfigForURI(uri)
	tsConfigPaths := s.tsConfigPathsForURI(uri)
	ruleDiags, err := runLintWithSession(uri, s.session, ctx, rslintConfig, configCwd, isJSConfig, tsConfigPaths, s.fs)
	if err != nil {
		log.Printf("Error running lint for push diagnostics: %v", err)
		return
	}

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
}
