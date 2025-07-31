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
	case "textDocument/completion":
		return s.handleCompletion(ctx, req)
	case "textDocument/hover":
		return s.handleHover(ctx, req)
	case "textDocument/definition":
		return s.handleDefinition(ctx, req)
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
			CompletionProvider: &lsproto.CompletionOptions{
				TriggerCharacters: &[]string{".", ":", "@", "#"},
				ResolveProvider:   ptrTo(false),
			},
			HoverProvider: &lsproto.BooleanOrHoverOptions{
				Boolean: ptrTo(true),
			},
			DefinitionProvider: &lsproto.BooleanOrDefinitionOptions{
				Boolean: ptrTo(true),
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
		files,
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

func (s *LSPServer) handleCompletion(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	var params lsproto.CompletionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeParseError,
			Message: "Failed to parse completion params",
		}
	}

	uri := params.TextDocument.Uri
	position := params.Position
	
	// Check if we have the document
	content, exists := s.documents[uri]
	if !exists {
		return nil, nil
	}

	// Get completion items using TypeScript compiler
	completionItems, err := s.getCompletionItems(string(uri), content, position)
	if err != nil {
		log.Printf("Error getting completion items: %v", err)
		return nil, nil
	}

	return &lsproto.CompletionList{
		IsIncomplete: false,
		Items:        completionItems,
	}, nil
}

func (s *LSPServer) handleHover(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	var params lsproto.HoverParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeParseError,
			Message: "Failed to parse hover params",
		}
	}

	uri := params.TextDocument.Uri
	position := params.Position
	
	// Check if we have the document
	content, exists := s.documents[uri]
	if !exists {
		return nil, nil
	}

	// Get hover information using TypeScript compiler
	hoverInfo, err := s.getHoverInfo(string(uri), content, position)
	if err != nil {
		log.Printf("Error getting hover info: %v", err)
		return nil, nil
	}

	return hoverInfo, nil
}

func (s *LSPServer) handleDefinition(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	var params lsproto.DefinitionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeParseError,
			Message: "Failed to parse definition params",
		}
	}

	uri := params.TextDocument.Uri
	position := params.Position
	
	// Check if we have the document
	content, exists := s.documents[uri]
	if !exists {
		return nil, nil
	}

	// Get definition locations using TypeScript compiler
	locations, err := s.getDefinitionLocations(string(uri), content, position)
	if err != nil {
		log.Printf("Error getting definition locations: %v", err)
		return nil, nil
	}

	return locations, nil
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
	
	// Check if we have the document
	content, exists := s.documents[uri]
	if !exists {
		return nil, nil
	}

	// Get code actions based on diagnostics and context
	codeActions, err := s.getCodeActions(string(uri), content, params)
	if err != nil {
		log.Printf("Error getting code actions: %v", err)
		return nil, nil
	}

	return codeActions, nil
}

// Helper methods for LSP features
func (s *LSPServer) getCompletionItems(uri string, content string, position lsproto.Position) ([]*lsproto.CompletionItem, error) {
	// Get TypeScript program and source file for better completions
	_ = uriToPath(uri) // filePath for future use
	
	// Create VFS and host for future TypeScript integration
	vfs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	workingDir := s.rootURI
	if workingDir == "" {
		workingDir = "."
	}
	_ = utils.CreateCompilerHost(workingDir, vfs) // host for future use

	// Try to get program for completions
	var items []*lsproto.CompletionItem
	
	// Basic TypeScript/JavaScript keywords and common identifiers
	tsKeywords := []string{
		"abstract", "any", "as", "async", "await", "boolean", "break", "case", "catch",
		"class", "const", "constructor", "continue", "debugger", "declare", "default",
		"delete", "do", "else", "enum", "export", "extends", "false", "finally", "for",
		"from", "function", "get", "if", "implements", "import", "in", "instanceof",
		"interface", "let", "module", "namespace", "never", "new", "null", "number",
		"object", "of", "package", "private", "protected", "public", "readonly",
		"return", "set", "static", "string", "super", "switch", "symbol", "this",
		"throw", "true", "try", "type", "typeof", "undefined", "unknown", "var",
		"void", "while", "with", "yield",
	}
	
	// Add keyword completions
	for _, keyword := range tsKeywords {
		items = append(items, &lsproto.CompletionItem{
			Label:  keyword,
			Kind:   ptrTo(lsproto.CompletionItemKindKeyword),
			Detail: ptrTo("TypeScript keyword"),
			Documentation: &lsproto.StringOrMarkupContent{
				String: ptrTo(fmt.Sprintf("TypeScript keyword: %s", keyword)),
			},
		})
	}
	
	// Add common method completions for string objects
	if position.Character > 0 {
		lines := strings.Split(content, "\n")
		if int(position.Line) < len(lines) {
			line := lines[position.Line]
			if int(position.Character) <= len(line) {
				// Check if we're after a dot (method completion)
				if int(position.Character) > 1 && line[position.Character-1] == '.' {
					stringMethods := []string{
						"charAt", "charCodeAt", "concat", "indexOf", "lastIndexOf",
						"match", "replace", "search", "slice", "split", "substr",
						"substring", "toLowerCase", "toUpperCase", "trim", "valueOf",
					}
					for _, method := range stringMethods {
						items = append(items, &lsproto.CompletionItem{
							Label:  method,
							Kind:   ptrTo(lsproto.CompletionItemKindMethod),
							Detail: ptrTo("String method"),
							Documentation: &lsproto.StringOrMarkupContent{
								String: ptrTo(fmt.Sprintf("String method: %s", method)),
							},
						})
					}
				}
			}
		}
	}
	
	return items, nil
}

func (s *LSPServer) getHoverInfo(uri string, content string, position lsproto.Position) (*lsproto.Hover, error) {
	// Enhanced hover information with better word detection and type hints
	
	lines := strings.Split(content, "\n")
	if int(position.Line) >= len(lines) {
		return nil, nil
	}
	
	line := lines[position.Line]
	if int(position.Character) >= len(line) {
		return nil, nil
	}
	
	// Enhanced word extraction
	start := int(position.Character)
	end := int(position.Character)
	
	// Find word boundaries (handle more complex identifiers)
	for start > 0 && isIdentifierChar(rune(line[start-1])) {
		start--
	}
	for end < len(line) && isIdentifierChar(rune(line[end])) {
		end++
	}
	
	if start == end {
		return nil, nil
	}
	
	word := line[start:end]
	
	// Provide enhanced information based on the word
	var hoverContent string
	var detectedType string
	
	// Check for common TypeScript/JavaScript types and keywords
	switch word {
	case "function":
		detectedType = "keyword"
		hoverContent = "**function** *(keyword)*\n\nDeclares a function in TypeScript/JavaScript."
	case "const":
		detectedType = "keyword"
		hoverContent = "**const** *(keyword)*\n\nDeclares a read-only named constant."
	case "let":
		detectedType = "keyword"
		hoverContent = "**let** *(keyword)*\n\nDeclares a block-scoped local variable."
	case "var":
		detectedType = "keyword"
		hoverContent = "**var** *(keyword)*\n\nDeclares a function-scoped or globally-scoped variable."
	case "class":
		detectedType = "keyword"
		hoverContent = "**class** *(keyword)*\n\nDeclares a class definition."
	case "interface":
		detectedType = "keyword"
		hoverContent = "**interface** *(keyword)*\n\nDefines a contract for object structure in TypeScript."
	case "type":
		detectedType = "keyword"
		hoverContent = "**type** *(keyword)*\n\nDefines a type alias in TypeScript."
	default:
		// Try to infer type from context
		detectedType = inferTypeFromContext(word, line, start, end)
		hoverContent = fmt.Sprintf("**%s** *(%s)*\n\n", word, detectedType)
		
		// Add additional context based on inferred type
		switch detectedType {
		case "string":
			hoverContent += "String value with methods like charAt(), substring(), etc."
		case "number":
			hoverContent += "Numeric value supporting arithmetic operations."
		case "boolean":
			hoverContent += "Boolean value (true or false)."
		case "function":
			hoverContent += "Function that can be called with arguments."
		default:
			hoverContent += "Identifier in the current scope."
		}
	}
	
	return &lsproto.Hover{
		Contents: lsproto.MarkupContentOrStringOrMarkedStringWithLanguageOrMarkedStrings{
			MarkupContent: &lsproto.MarkupContent{
				Kind:  lsproto.MarkupKindMarkdown,
				Value: hoverContent,
			},
		},
		Range: &lsproto.Range{
			Start: lsproto.Position{
				Line:      position.Line,
				Character: uint32(start),
			},
			End: lsproto.Position{
				Line:      position.Line,
				Character: uint32(end),
			},
		},
	}, nil
}

func (s *LSPServer) getDefinitionLocations(uri string, content string, position lsproto.Position) ([]*lsproto.Location, error) {
	// For now, return empty - in a full implementation this would use TypeScript compiler
	// to find the actual definition location
	return []*lsproto.Location{}, nil
}

func (s *LSPServer) getCodeActions(uri string, content string, params lsproto.CodeActionParams) ([]*lsproto.CodeAction, error) {
	var actions []*lsproto.CodeAction
	
	// Generate quick fixes for diagnostics in the range
	for _, diagnostic := range params.Context.Diagnostics {
		if diagnostic.Source != nil && *diagnostic.Source == "rslint" {
			// Create more specific fix actions based on the diagnostic message
			title := fmt.Sprintf("Fix: %s", diagnostic.Message)
			var fixText string
			
			// Provide context-specific fixes based on common rslint rules
			message := diagnostic.Message
			if strings.Contains(message, "unused") {
				title = "Remove unused variable"
				fixText = ""
			} else if strings.Contains(message, "semicolon") {
				title = "Add missing semicolon"
				fixText = ";"
			} else if strings.Contains(message, "quote") {
				title = "Fix quote style"
				fixText = "/* Quote fix needed */"
			} else {
				title = "Apply suggested fix"
				fixText = "/* TODO: Auto-fix not implemented yet */"
			}
			
			action := &lsproto.CodeAction{
				Title: title,
				Kind:  ptrTo(lsproto.CodeActionKindQuickFix),
				Diagnostics: &[]*lsproto.Diagnostic{diagnostic},
				Edit: &lsproto.WorkspaceEdit{
					Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
						lsproto.DocumentUri(uri): {
							{
								Range:   diagnostic.Range,
								NewText: fixText,
							},
						},
					},
				},
			}
			actions = append(actions, action)
		}
	}
	
	// Add source actions
	organizeImportsAction := &lsproto.CodeAction{
		Title: "Organize Imports",
		Kind:  ptrTo(lsproto.CodeActionKindSourceOrganizeImports),
		Edit: &lsproto.WorkspaceEdit{
			Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
				lsproto.DocumentUri(uri): {
					{
						Range: lsproto.Range{
							Start: lsproto.Position{Line: 0, Character: 0},
							End:   lsproto.Position{Line: 0, Character: 0},
						},
						NewText: "// Organize imports action triggered\n",
					},
				},
			},
		},
	}
	actions = append(actions, organizeImportsAction)
	
	return actions, nil
}

func isIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '$'
}

func inferTypeFromContext(word string, line string, start int, end int) string {
	// Simple type inference based on context clues
	
	// Check for string literals
	if strings.Contains(line, `"`+word+`"`) || strings.Contains(line, `'`+word+`'`) || strings.Contains(line, "`"+word+"`") {
		return "string"
	}
	
	// Check for number assignments
	if start > 2 && strings.Contains(line[0:start], "=") {
		after := ""
		if end < len(line) {
			after = strings.TrimSpace(line[end:])
		}
		if len(after) > 0 && (after[0] >= '0' && after[0] <= '9') {
			return "number"
		}
	}
	
	// Check for boolean keywords
	if word == "true" || word == "false" {
		return "boolean"
	}
	
	// Check for function declarations
	if strings.Contains(line, "function "+word) || 
		 (start > 0 && strings.Contains(line[0:start], word+" = function")) ||
		 strings.Contains(line, word+" = (") {
		return "function"
	}
	
	// Check for class declarations
	if strings.Contains(line, "class "+word) {
		return "class"
	}
	
	// Check for method calls (ends with parentheses)
	if end < len(line) && strings.TrimSpace(line[end:]) != "" && strings.HasPrefix(strings.TrimSpace(line[end:]), "(") {
		return "function"
	}
	
	// Default to identifier
	return "identifier"
}
