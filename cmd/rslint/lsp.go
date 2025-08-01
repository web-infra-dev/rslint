package main

import (
	"context"
	"encoding/json"
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
	"github.com/typescript-eslint/rslint/internal/config"
	"github.com/typescript-eslint/rslint/internal/linter"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

func ptrTo[T any](v T) *T {
	return &v
}

// LSP Server implementation
type LSPServer struct {
	conn      *jsonrpc2.Conn
	rootURI   string
	documents map[lsproto.DocumentUri]string // URI -> content
}

func NewLSPServer() *LSPServer {
	return &LSPServer{
		documents: make(map[lsproto.DocumentUri]string),
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
	case "shutdown":
		return s.handleShutdown(ctx, req)

	case "exit":
		os.Exit(0)
		return nil, nil
	default:
		// Respond with method not found for unhandled methods
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: fmt.Sprintf("method not found: %s", req.Method),
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

	if params.RootUri.DocumentUri != nil {
		s.rootURI = uriToPath(string(*params.RootUri.DocumentUri))
	}

	result := &lsproto.InitializeResult{
		Capabilities: &lsproto.ServerCapabilities{
			TextDocumentSync: &lsproto.TextDocumentSyncOptionsOrKind{
				Kind: ptrTo(lsproto.TextDocumentSyncKindFull),
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
	var allSourceFiles []*ast.SourceFile
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
		allSourceFiles = append(allSourceFiles, sourceFiles...)

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
		return nil, fmt.Errorf("no programs provided")
	}

	// Initialize rule registry with all available rules
	config.RegisterAllTypeSriptEslintPluginRules()

	// Find all source files from all programs, prioritizing the file matching the URI
	var targetFile *ast.SourceFile
	uriPath := uriToPath(string(uri))

	for _, program := range programs {
		for _, file := range program.SourceFiles() {

			if file.FileName() == uriPath {
				targetFile = file
			}
		}
	}

	// If we found a target file, prioritize it, otherwise use all files
	var files []*ast.SourceFile
	if targetFile != nil {
		files = []*ast.SourceFile{targetFile}
	} else {
		log.Printf("Target file not found for URI %s, processing all files", uri)
	}

	log.Printf("Processing %d files from %d programs", len(files), len(programs))

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
	err := linter.RunLinter(
		programs,
		false, // Don't use single-threaded mode for LSP
		&files,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			activeRules := config.GlobalRuleRegistry.GetEnabledRules(rslintConfig, sourceFile.FileName())
			return activeRules
		},
		diagnosticCollector,
	)
	if err != nil {
		return nil, fmt.Errorf("error running linter: %v", err)
	}

	if diagnostics == nil {
		diagnostics = []rule.RuleDiagnostic{}
	}
	return diagnostics, nil
}
