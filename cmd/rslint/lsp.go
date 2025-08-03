package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
}

func NewLSPServer() *LSPServer {
	return &LSPServer{
		documents:   make(map[lsproto.DocumentUri]string),
		diagnostics: make(map[lsproto.DocumentUri][]rule.RuleDiagnostic),
	}
}

func (s *LSPServer) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	s.conn = conn

	requestJSON, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		return nil, err
	}

	log.Printf("Received request: %s", string(requestJSON))
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
	var params lsproto.InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeParseError,
			Message: "Failed to parse initialize params",
		}
	}

	// Set root URI from workspace folders if available, otherwise fall back to rootUri
	if params.WorkspaceFolders != nil && len(*params.WorkspaceFolders.WorkspaceFolders) > 0 {
		// Use the first workspace folder as the root
		s.rootURI = uriToPath(string((*params.WorkspaceFolders.WorkspaceFolders)[0].Uri))
	} else {
		return nil, errors.New("no workspace folders provided in initialize params")
	}

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
		return []*lsproto.CodeAction{}, nil
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
	uriString := string(uri)

	// Only process TypeScript/JavaScript files
	if !isTypeScriptFile(uriString) {
		return
	}

	// Initialize rule registry with all available rules (ensure it's done once)
	config.RegisterAllTypeSriptEslintPluginRules()

	// Convert URI to file path
	filePath := uriToPath(uriString)

	// Create a temporary file system with the content
	vfs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	// Create TypeScript program using utils
	// Use the directory containing the file as working directory
	workingDir := s.rootURI
	if workingDir == "" {
		workingDir = "."
	}

	host := utils.CreateCompilerHost(workingDir, vfs)

	// Try to find rslint.json in the working directory
	rslintConfigPath := workingDir + "/rslint.json"
	if !vfs.FileExists(rslintConfigPath) {
		// If no rslint.json found, skip diagnostics for now
		// In a real implementation, you'd create a default config
		log.Printf("No rslint.json found at %s", rslintConfigPath)
		return
	}

	// Load rslint configuration and extract tsconfig paths
	loader := config.NewConfigLoader(vfs, workingDir)
	rslintConfig, configDirectory, err := loader.LoadRslintConfig("rslint.json")
	if err != nil {
		log.Printf("Could not load rslint config: %v", err)
		return
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, configDirectory)
	if err != nil {
		log.Printf("Could not load TypeScript configs from rslint config: %v", err)
		return
	}

	if len(tsConfigs) == 0 {
		log.Printf("No TypeScript configurations found in rslint config")
		return
	}

	// Create multiple programs for all tsconfig files
	var programs []*compiler.Program
	var targetFile *ast.SourceFile

	for _, tsConfigPath := range tsConfigs {
		program, err := utils.CreateProgram(true, vfs, workingDir, tsConfigPath, host)
		if err != nil {
			log.Printf("Could not create program for %s: %v", tsConfigPath, err)
			continue
		}
		programs = append(programs, program)

		// Check if the current file is in this program
		sourceFiles := program.GetSourceFiles()

		if targetFile == nil {
			for _, sf := range sourceFiles {
				if strings.HasSuffix(sf.FileName(), filePath) || sf.FileName() == filePath {
					targetFile = sf
					break
				}
			}
		}
	}

	if len(programs) == 0 {
		log.Printf("Could not create any programs")
		return
	}

	if targetFile == nil {
		// If we can't find the file in any program, skip diagnostics
		log.Printf("Could not find file %s in any program", filePath)
		return
	}

	// Collect diagnostics
	var lsp_diagnostics []*lsproto.Diagnostic

	rule_diags, err := runLintWithPrograms(uri, programs, rslintConfig)

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

func uriToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		return strings.TrimPrefix(uri, "file://")
	}
	return uri
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

func runLintWithPrograms(uri lsproto.DocumentUri, programs []*compiler.Program, rslintConfig config.RslintConfig) ([]rule.RuleDiagnostic, error) {
	if len(programs) == 0 {
		return nil, errors.New("no programs provided")
	}

	// Initialize rule registry with all available rules
	config.RegisterAllTypeSriptEslintPluginRules()

	// Collect diagnostics
	var diagnostics []rule.RuleDiagnostic
	var diagnosticsLock sync.Mutex

	// Create collector function
	diagnosticCollector := func(d rule.RuleDiagnostic) {
		diagnosticsLock.Lock()
		defer diagnosticsLock.Unlock()
		diagnostics = append(diagnostics, d)
	}
	// Run linter with all programs using rule registry
	_, err := linter.RunLinter(
		programs,
		false, // Don't use single-threaded mode for LSP
		nil,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			activeRules := config.GlobalRuleRegistry.GetEnabledRules(rslintConfig, sourceFile.FileName())
			return activeRules
		},
		diagnosticCollector,
	)
	if err != nil {
		return nil, fmt.Errorf("error running linter: %w", err)
	}

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
