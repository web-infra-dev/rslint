package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/ls"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/scanner"

	// "github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	util "github.com/web-infra-dev/rslint/internal/utils"
)

func ptrTo[T any](v T) *T {
	return &v
}

// LSP Server implementation
type LSPServer struct {
	conn        *jsonrpc2.Conn
	rootURI     string
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
		// Client finished initialization
		return nil, nil
	case "textDocument/didOpen":
		s.handleDidOpen(ctx, req)
		return nil, nil
	case "textDocument/didChange":
		s.handleDidChange(ctx, req)
		return nil, nil
	case "textDocument/didSave":
		s.handleDidSave(ctx, req)
		return nil, nil
	case "textDocument/diagnostic":
		s.handleDidSave(ctx, req)
		return nil, nil
	case "textDocument/codeAction":
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

func (s *LSPServer) handleInitialize(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	log.Printf("Handling initialize: %+v with pid %d", req, os.Getpid())
	// Check if params is nil
	if req.Params == nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "Initialize params cannot be nil",
		}
	}

	// Parse initialize params
	var params lsproto.InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		s.rootURI = "."
	} else {
		//nolint
		if params.RootUri.DocumentUri != nil {
			s.rootURI = uriToPath(*params.RootUri.DocumentUri)
		}
	}

	s.projectService = project.NewService(s, project.ServiceOptions{
		Logger:           project.NewLogger([]io.Writer{os.Stderr}, "tsgo.log", project.LogLevelVerbose),
		PositionEncoding: lsproto.PositionEncodingKindUTF8,
	})
	// Try to find rslint configuration files with multiple strategies
	var rslintConfigPath string
	var configFound bool

	// Use helper function to find config
	rslintConfigPath, configFound = findRslintConfig(s.fs, s.cwd)

	if !configFound {
		return nil, errors.New("config file not found")
	}

	// Load rslint configuration and extract tsconfig paths
	loader := config.NewConfigLoader(s.fs, s.cwd)
	rslintConfig, configDirectory, err := loader.LoadRslintConfig(rslintConfigPath)
	if err != nil {
		return nil, fmt.Errorf("could not load rslint config: %w", err)
	}
	s.rslintConfig = rslintConfig
	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, configDirectory)
	if err != nil {
		return nil, fmt.Errorf("could not load TypeScript configs from rslint config: %w", err)
	}

	if len(tsConfigs) == 0 {
		return nil, errors.New("no TypeScript configurations found in rslint config")
	}

	// Do not pre-create configured projects here. The service will create
	// configured or inferred projects on demand when files are opened.
	result := &lsproto.InitializeResult{
		Capabilities: &lsproto.ServerCapabilities{
			TextDocumentSync: &lsproto.TextDocumentSyncOptionsOrKind{
				Kind: ptrTo(lsproto.TextDocumentSyncKindFull),
			},
			CodeActionProvider: &lsproto.BooleanOrCodeActionOptions{
				Boolean: ptrTo(true),
			},
		},
	}

	return result, nil
}

func (s *LSPServer) handleDidOpen(ctx context.Context, req *jsonrpc2.Request) {
	log.Printf("Handling didOpen: %+v,%+v", req, ctx)
	var params lsproto.DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}

	uri := params.TextDocument.Uri
	content := params.TextDocument.Text

	s.documents[uri] = content
	s.runDiagnostics(ctx, uri, content)
}

func (s *LSPServer) handleDidChange(ctx context.Context, req *jsonrpc2.Request) {
	log.Printf("Handling didChange: %+v", req)
	var params lsproto.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}

	uri := params.TextDocument.Uri

	// For full document sync, we expect one change with the full text
	if len(params.ContentChanges) > 0 {
		content := params.ContentChanges[0].WholeDocument.Text
		s.documents[uri] = content
		s.runDiagnostics(ctx, uri, content)
	}
}

func (s *LSPServer) handleDidSave(ctx context.Context, req *jsonrpc2.Request) {
	log.Printf("Handling didSave: %+v", req)
	// Re-run diagnostics on save
	var params lsproto.DidSaveTextDocumentParams

	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}

	uri := params.TextDocument.Uri
	if content, exists := s.documents[uri]; exists {
		s.runDiagnostics(ctx, uri, content)
	}
}

func (s *LSPServer) handleShutdown(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	return nil, nil
}

func (s *LSPServer) handleCodeAction(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	log.Printf("Handling codeAction: %+v,%+v", req, ctx)
	var params lsproto.CodeActionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeParseError,
			Message: "Failed to parse code action params",
		}
	}

	uri := params.TextDocument.Uri

	// Get stored diagnostics for this document
	ruleDiagnostics, exists := s.diagnostics[uri]
	if !exists {
		// If no diagnostics exist for this document, try to generate them
		// This can happen if the document was opened without a proper didOpen event
		filePath := uriToPath(uri)
		if content, err := os.ReadFile(filePath); err == nil {
			s.documents[uri] = string(content)
			s.runDiagnostics(ctx, uri, string(content))

			// Try to get diagnostics again after running them
			ruleDiagnostics, exists = s.diagnostics[uri]
			if !exists {
				return []*lsproto.CodeAction{}, nil
			}
		} else {
			return []*lsproto.CodeAction{}, nil
		}
	}

	var codeActions []*lsproto.CodeAction

	// Find diagnostics that overlap with the requested range
	for _, ruleDiag := range ruleDiagnostics {
		// Check if diagnostic range overlaps with requested range
		diagStartLine, diagStartChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.Pos())
		diagEndLine, diagEndChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.End())

		diagRange := lsproto.Range{
			Start: lsproto.Position{Line: uint32(diagStartLine), Character: uint32(diagStartChar)},
			End:   lsproto.Position{Line: uint32(diagEndLine), Character: uint32(diagEndChar)},
		}

		if rangesOverlap(diagRange, params.Range) {
			// Add code action for fixes
			codeAction := createCodeActionFromRuleDiagnostic(ruleDiag, uri)
			if codeAction != nil {
				codeActions = append(codeActions, codeAction)
			}
			// add extract disable rule actions
			disableActions := createDisableRuleActions(ruleDiag, uri)
			codeActions = append(codeActions, disableActions...)

			// Add code actions for suggestions
			if ruleDiag.Suggestions != nil {
				for _, suggestion := range *ruleDiag.Suggestions {
					suggestionAction := createCodeActionFromSuggestion(ruleDiag, suggestion, uri)
					if suggestionAction != nil {
						codeActions = append(codeActions, suggestionAction)
					}
				}
			}
		}
	}

	return codeActions, nil
}

func (s *LSPServer) runDiagnostics(ctx context.Context, uri lsproto.DocumentUri, content string) {
	log.Printf("Running diagnostics for: %+v", uri)
	uriString := string(uri)

	// Only process TypeScript/JavaScript files
	if !isTypeScriptFile(uriString) {
		return
	}

	// Initialize rule registry with all available rules (ensure it's done once)
	config.RegisterAllRules()

	// Collect diagnostics
	var lsp_diagnostics []*lsproto.Diagnostic

	rule_diags, err := runLintWithProjectService(uri, s.projectService, ctx, s.rslintConfig)

	if err != nil {
		log.Printf("Error running lint: %v", err)
	}

	// Store rule diagnostics for code actions
	s.diagnostics[uri] = rule_diags

	for _, diagnostic := range rule_diags {
		lspDiag := convertRuleDiagnosticToLSP(diagnostic, content)
		lsp_diagnostics = append(lsp_diagnostics, lspDiag)
	}
	// Publish diagnostics
	params := lsproto.PublishDiagnosticsParams{
		Uri:         uri,
		Diagnostics: lsp_diagnostics,
	}

	var diagsBuilder strings.Builder
	for _, diag := range lsp_diagnostics {
		fmt.Fprintf(&diagsBuilder, "%v:%+v\n", diag.Message, diag.Range)
	}
	log.Printf("Publishing diagnostics for %s:\n%s", uri, diagsBuilder.String())
	s.conn.Notify(ctx, "textDocument/publishDiagnostics", params)
}

func convertRuleDiagnosticToLSP(ruleDiag rule.RuleDiagnostic, content string) *lsproto.Diagnostic {
	diagnosticStart := ruleDiag.Range.Pos()
	diagnosticEnd := ruleDiag.Range.End()
	startLine, startColumn := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, diagnosticStart)
	endLine, endColumn := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, diagnosticEnd)

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
	return ls.DocumentURIToFileName(uri)
}

// findRslintConfig searches for rslint configuration files using multiple strategies
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

func runLSP() int {
	log.SetOutput(os.Stderr) // Send logs to stderr so they don't interfere with LSP communication

	server := NewLSPServer()

	// Create a simple ReadWriteCloser from stdin/stdout
	stream := &struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: os.Stdin,
		Writer: os.Stdout,
		Closer: os.Stdin,
	}

	// Create connection using stdin/stdout
	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(stream, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(server.Handle),
	)

	// Wait for connection to close
	<-conn.DisconnectNotify()

	return 0
}

// LintResponse represents a lint response from Go to JS
type LintResponse struct {
	Diagnostics []lsproto.Diagnostic `json:"diagnostics"`
	ErrorCount  int                  `json:"errorCount"`
	FileCount   int                  `json:"fileCount"`
	RuleCount   int                  `json:"ruleCount"`
}

func runLintWithProjectService(uri lsproto.DocumentUri, service *project.Service, ctx context.Context, rslintConfig config.RslintConfig) ([]rule.RuleDiagnostic, error) {
	log.Printf("context: %v", ctx)
	// Initialize rule registry with all available rules
	config.RegisterAllRules()
	filename := uriToPath(uri)
	content, ok := service.FS().ReadFile(filename)
	if !ok {
		return nil, fmt.Errorf("failed to read file %s", filename)
	}
	service.OpenFile(filename, content, core.GetScriptKindFromFileName(filename), service.GetCurrentDirectory())
	project := service.EnsureDefaultProjectForURI(uri)
	languageService, done := project.GetLanguageServiceForRequest(ctx)
	program := languageService.GetProgram()
	defer done()
	// Collect diagnostics
	var diagnostics []rule.RuleDiagnostic
	var diagnosticsLock sync.Mutex

	// Create collector function
	diagnosticCollector := func(d rule.RuleDiagnostic) {
		diagnosticsLock.Lock()
		defer diagnosticsLock.Unlock()
		diagnostics = append(diagnostics, d)
	}

	linter.RunLinterInProgram(program, []string{filename}, util.ExcludePaths,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			activeRules := config.GlobalRuleRegistry.GetEnabledRules(rslintConfig, sourceFile.FileName())
			return activeRules
		}, diagnosticCollector)

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
		startLine, startChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, fix.Range.Pos())
		endLine, endChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, fix.Range.End())

		textEdit := &lsproto.TextEdit{
			Range: lsproto.Range{
				Start: lsproto.Position{Line: uint32(startLine), Character: uint32(startChar)},
				End:   lsproto.Position{Line: uint32(endLine), Character: uint32(endChar)},
			},
			NewText: fix.Text,
		}
		textEdits = append(textEdits, textEdit)
	}

	// Create workspace edit
	workspaceEdit := &lsproto.WorkspaceEdit{
		Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
			uri: textEdits,
		},
	}

	// Create the corresponding LSP diagnostic for reference
	diagStartLine, diagStartChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.Pos())
	diagEndLine, diagEndChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.End())

	lspDiagnostic := &lsproto.Diagnostic{
		Range: lsproto.Range{
			Start: lsproto.Position{Line: uint32(diagStartLine), Character: uint32(diagStartChar)},
			End:   lsproto.Position{Line: uint32(diagEndLine), Character: uint32(diagEndChar)},
		},
		Severity: ptrTo(lsproto.DiagnosticSeverity(ruleDiag.Severity.Int())),
		Source:   ptrTo("rslint"),
		Message:  fmt.Sprintf("[%s] %s", ruleDiag.RuleName, ruleDiag.Message.Description),
	}

	return &lsproto.CodeAction{
		Title:       "Fix: " + ruleDiag.Message.Description,
		Kind:        ptrTo(lsproto.CodeActionKind("quickfix")),
		Edit:        workspaceEdit,
		Diagnostics: &[]*lsproto.Diagnostic{lspDiagnostic},
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
		startLine, startChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, fix.Range.Pos())
		endLine, endChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, fix.Range.End())

		textEdit := &lsproto.TextEdit{
			Range: lsproto.Range{
				Start: lsproto.Position{Line: uint32(startLine), Character: uint32(startChar)},
				End:   lsproto.Position{Line: uint32(endLine), Character: uint32(endChar)},
			},
			NewText: fix.Text,
		}
		textEdits = append(textEdits, textEdit)
	}

	// Create workspace edit
	workspaceEdit := &lsproto.WorkspaceEdit{
		Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
			uri: textEdits,
		},
	}

	// Create the corresponding LSP diagnostic for reference
	diagStartLine, diagStartChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.Pos())
	diagEndLine, diagEndChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.End())

	lspDiagnostic := &lsproto.Diagnostic{
		Range: lsproto.Range{
			Start: lsproto.Position{Line: uint32(diagStartLine), Character: uint32(diagStartChar)},
			End:   lsproto.Position{Line: uint32(diagEndLine), Character: uint32(diagEndChar)},
		},
		Severity: ptrTo(lsproto.DiagnosticSeverity(ruleDiag.Severity.Int())),
		Source:   ptrTo("rslint"),
		Message:  fmt.Sprintf("[%s] %s", ruleDiag.RuleName, ruleDiag.Message.Description),
	}

	return &lsproto.CodeAction{
		Title:       "Suggestion: " + suggestion.Message.Description,
		Kind:        ptrTo(lsproto.CodeActionKind("quickfix")),
		Edit:        workspaceEdit,
		Diagnostics: &[]*lsproto.Diagnostic{lspDiagnostic},
		IsPreferred: ptrTo(false), // Mark suggestions as not preferred
	}
}

// Helper function to create disable rule actions for diagnostics without fixes
func createDisableRuleActions(ruleDiag rule.RuleDiagnostic, uri lsproto.DocumentUri) []*lsproto.CodeAction {
	var actions []*lsproto.CodeAction

	// Create the corresponding LSP diagnostic for reference
	diagStartLine, diagStartChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.Pos())
	diagEndLine, diagEndChar := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.End())

	lspDiagnostic := &lsproto.Diagnostic{
		Range: lsproto.Range{
			Start: lsproto.Position{Line: uint32(diagStartLine), Character: uint32(diagStartChar)},
			End:   lsproto.Position{Line: uint32(diagEndLine), Character: uint32(diagEndChar)},
		},
		Severity: ptrTo(lsproto.DiagnosticSeverity(ruleDiag.Severity.Int())),
		Source:   ptrTo("rslint"),
		Message:  fmt.Sprintf("[%s] %s", ruleDiag.RuleName, ruleDiag.Message.Description),
	}

	// Action 1: Disable rule for this line
	disableLineAction := createDisableRuleForLineAction(ruleDiag, uri, lspDiagnostic)
	if disableLineAction != nil {
		actions = append(actions, disableLineAction)
	}

	// Action 2: Disable rule for entire file
	disableFileAction := createDisableRuleForFileAction(ruleDiag, uri, lspDiagnostic)
	if disableFileAction != nil {
		actions = append(actions, disableFileAction)
	}

	return actions
}

// Helper function to create a "disable rule for this line" action
func createDisableRuleForLineAction(ruleDiag rule.RuleDiagnostic, uri lsproto.DocumentUri, lspDiagnostic *lsproto.Diagnostic) *lsproto.CodeAction {
	// Get the line where the diagnostic occurs
	lineStart := lspDiagnostic.Range.Start.Line

	// Create text edit to add eslint-disable-next-line comment
	disableComment := fmt.Sprintf("// eslint-disable-next-line %s\n", ruleDiag.RuleName)

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
	// Create text edit to add eslint-disable comment at the top of the file
	disableComment := fmt.Sprintf("/* eslint-disable %s */\n", ruleDiag.RuleName)

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
