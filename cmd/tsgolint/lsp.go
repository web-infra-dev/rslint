package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/rules/await_thenable"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_array_delete"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_base_to_string"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_confusing_void_expression"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_duplicate_type_constituents"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_floating_promises"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_for_in_array"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_implied_eval"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_meaningless_void_operator"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_misused_promises"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_misused_spread"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_mixed_enums"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_redundant_type_constituents"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unnecessary_boolean_literal_compare"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unnecessary_template_expression"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unnecessary_type_arguments"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unnecessary_type_assertion"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_argument"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_assignment"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_call"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_enum_comparison"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_member_access"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_return"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_type_assertion"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_unary_minus"
	"github.com/typescript-eslint/tsgolint/internal/rules/non_nullable_type_assertion_style"
	"github.com/typescript-eslint/tsgolint/internal/rules/only_throw_error"
	"github.com/typescript-eslint/tsgolint/internal/rules/prefer_promise_reject_errors"
	"github.com/typescript-eslint/tsgolint/internal/rules/prefer_reduce_type_parameter"
	"github.com/typescript-eslint/tsgolint/internal/rules/prefer_return_this_type"
	"github.com/typescript-eslint/tsgolint/internal/rules/promise_function_async"
	"github.com/typescript-eslint/tsgolint/internal/rules/related_getter_setter_pairs"
	"github.com/typescript-eslint/tsgolint/internal/rules/require_array_sort_compare"
	"github.com/typescript-eslint/tsgolint/internal/rules/require_await"
	"github.com/typescript-eslint/tsgolint/internal/rules/restrict_plus_operands"
	"github.com/typescript-eslint/tsgolint/internal/rules/restrict_template_expressions"
	"github.com/typescript-eslint/tsgolint/internal/rules/return_await"
	"github.com/typescript-eslint/tsgolint/internal/rules/switch_exhaustiveness_check"
	"github.com/typescript-eslint/tsgolint/internal/rules/unbound_method"
	"github.com/typescript-eslint/tsgolint/internal/rules/use_unknown_in_catch_callback_variable"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// LSP protocol structures
type InitializeParams struct {
	ProcessID    *int    `json:"processId"`
	RootPath     *string `json:"rootPath"`
	RootURI      *string `json:"rootUri"`
	Capabilities struct {
		TextDocument struct {
			PublishDiagnostics struct {
				RelatedInformation bool `json:"relatedInformation"`
				TagSupport         struct {
					ValueSet []int `json:"valueSet"`
				} `json:"tagSupport"`
			} `json:"publishDiagnostics"`
		} `json:"textDocument"`
	} `json:"capabilities"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

type ServerCapabilities struct {
	TextDocumentSync   int  `json:"textDocumentSync"`
	DiagnosticProvider bool `json:"diagnosticProvider"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

type LspDiagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Source   string `json:"source"`
	Message  string `json:"message"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type PublishDiagnosticsParams struct {
	URI         string          `json:"uri"`
	Diagnostics []LspDiagnostic `json:"diagnostics"`
}

// LSP Server implementation
type LSPServer struct {
	conn      *jsonrpc2.Conn
	rootURI   string
	documents map[string]string // URI -> content
}

func NewLSPServer() *LSPServer {
	return &LSPServer{
		documents: make(map[string]string),
	}
}

func (s *LSPServer) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	s.conn = conn
	log.Printf("Received request: %v", req)
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
	var params InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeParseError,
			Message: "Failed to parse initialize params",
		}
	}

	if params.RootURI != nil {
		s.rootURI = *params.RootURI
		// Remove file:// prefix if present
		s.rootURI = strings.TrimPrefix(s.rootURI, "file://")
	}

	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:   1, // Full document sync
			DiagnosticProvider: true,
		},
	}

	return result, nil
}

func (s *LSPServer) handleDidOpen(ctx context.Context, req *jsonrpc2.Request) {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}

	uri := params.TextDocument.URI
	content := params.TextDocument.Text

	s.documents[uri] = content
	s.runDiagnostics(ctx, uri, content)
}

func (s *LSPServer) handleDidChange(ctx context.Context, req *jsonrpc2.Request) {
	var params DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}

	uri := params.TextDocument.URI

	// For full document sync, we expect one change with the full text
	if len(params.ContentChanges) > 0 {
		content := params.ContentChanges[0].Text
		s.documents[uri] = content
		s.runDiagnostics(ctx, uri, content)
	}
}

func (s *LSPServer) handleDidSave(ctx context.Context, req *jsonrpc2.Request) {
	// Re-run diagnostics on save
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}

	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}

	uri := params.TextDocument.URI
	if content, exists := s.documents[uri]; exists {
		s.runDiagnostics(ctx, uri, content)
	}
}

func (s *LSPServer) handleShutdown(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	return nil, nil
}

func (s *LSPServer) runDiagnostics(ctx context.Context, uri, content string) {
	// Only process TypeScript/JavaScript files
	if !isTypeScriptFile(uri) {
		return
	}

	// Convert URI to file path
	filePath := uriToPath(uri)

	// Create a temporary file system with the content
	vfs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	// Create TypeScript program using utils
	// Use the directory containing the file as working directory
	workingDir := s.rootURI
	if workingDir == "" {
		workingDir = "."
	}

	host := utils.CreateCompilerHost(workingDir, vfs)

	// Try to find tsconfig.json in the working directory
	tsconfigPath := workingDir + "/tsconfig.json"
	if !vfs.FileExists(tsconfigPath) {
		// If no tsconfig found, skip diagnostics for now
		// In a real implementation, you'd create a default config
		log.Printf("No tsconfig.json found at %s", tsconfigPath)
		return
	}

	// For simplicity, we'll create a minimal tsconfig for single file analysis
	program, err := utils.CreateProgram(true, vfs, workingDir, "tsconfig.json", host)
	if err != nil {
		// If we can't create with tsconfig, skip for now
		log.Printf("Could not create program: %v", err)
		return
	}

	sourceFiles := program.GetSourceFiles()
	var targetFile *ast.SourceFile
	for _, sf := range sourceFiles {
		if strings.HasSuffix(sf.FileName(), filePath) || sf.FileName() == filePath {
			targetFile = sf
			break
		}
	}

	if targetFile == nil {
		// If we can't find the file in the program, skip diagnostics
		log.Printf("Could not find file %s in program", filePath)
		return
	}

	// Collect diagnostics
	var lsp_diagnostics []LspDiagnostic

	rule_diags, err := runLint(uri)
	for _, diagnostic := range rule_diags {
		lspDiag := convertRuleDiagnosticToLSP(diagnostic, content)
		lsp_diagnostics = append(lsp_diagnostics, lspDiag)
	}
	fmt.Fprintf(os.Stderr, "my diags: %v", lsp_diagnostics)
	// Publish diagnostics
	params := PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: lsp_diagnostics,
	}

	s.conn.Notify(ctx, "textDocument/publishDiagnostics", params)
}

func getAvailableRules() []rule.Rule {
	return []rule.Rule{
		no_array_delete.NoArrayDeleteRule,
		no_base_to_string.NoBaseToStringRule,
		no_for_in_array.NoForInArrayRule,
		no_implied_eval.NoImpliedEvalRule,
		only_throw_error.OnlyThrowErrorRule,
	}
}

func convertRuleDiagnosticToLSP(ruleDiag rule.RuleDiagnostic, content string) LspDiagnostic {
	diagnosticStart := ruleDiag.Range.Pos()
	diagnosticEnd := ruleDiag.Range.End()
	startLine, startColumn := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, diagnosticStart)
	endLine, endColumn := scanner.GetLineAndCharacterOfPosition(ruleDiag.SourceFile, diagnosticEnd)

	return LspDiagnostic{
		Range: Range{
			Start: Position{
				Line:      startLine,
				Character: startColumn,
			},
			End: Position{
				Line:      endLine,
				Character: endColumn,
			},
		},
		Severity: 1, // Error
		Source:   "rslint",
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
	Diagnostics []LspDiagnostic `json:"diagnostics"`
	ErrorCount  int             `json:"errorCount"`
	FileCount   int             `json:"fileCount"`
	RuleCount   int             `json:"ruleCount"`
}

// HandleLint handles lint requests in IPC mode
func runLint(uri string) ([]rule.RuleDiagnostic, error) {
	var tsconfig string
	// Get current directory
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)

	// Create filesystem
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	// Handle tsconfig
	var configFileName string
	if tsconfig == "" {
		configFileName = tspath.ResolvePath(currentDirectory, "tsconfig.json")
		if !fs.FileExists(configFileName) {
			fs = utils.NewOverlayVFS(fs, map[string]string{
				configFileName: "{}",
			})
		}
	} else {
		configFileName = tspath.ResolvePath(currentDirectory, tsconfig)
		if !fs.FileExists(configFileName) {
			return nil, fmt.Errorf("error: tsconfig %q doesn't exist", tsconfig)
		}
	}
	currentDirectory = tspath.GetDirectoryPath(configFileName)

	// Create rules
	var rules = []rule.Rule{
		await_thenable.AwaitThenableRule,
		no_array_delete.NoArrayDeleteRule,
		no_base_to_string.NoBaseToStringRule,
		no_confusing_void_expression.NoConfusingVoidExpressionRule,
		no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule,
		no_floating_promises.NoFloatingPromisesRule,
		no_for_in_array.NoForInArrayRule,
		no_implied_eval.NoImpliedEvalRule,
		no_meaningless_void_operator.NoMeaninglessVoidOperatorRule,
		no_misused_promises.NoMisusedPromisesRule,
		no_misused_spread.NoMisusedSpreadRule,
		no_mixed_enums.NoMixedEnumsRule,
		no_redundant_type_constituents.NoRedundantTypeConstituentsRule,
		no_unnecessary_boolean_literal_compare.NoUnnecessaryBooleanLiteralCompareRule,
		no_unnecessary_template_expression.NoUnnecessaryTemplateExpressionRule,
		no_unnecessary_type_arguments.NoUnnecessaryTypeArgumentsRule,
		no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule,
		no_unsafe_argument.NoUnsafeArgumentRule,
		no_unsafe_assignment.NoUnsafeAssignmentRule,
		no_unsafe_call.NoUnsafeCallRule,
		no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule,
		no_unsafe_member_access.NoUnsafeMemberAccessRule,
		no_unsafe_return.NoUnsafeReturnRule,
		no_unsafe_type_assertion.NoUnsafeTypeAssertionRule,
		no_unsafe_unary_minus.NoUnsafeUnaryMinusRule,
		non_nullable_type_assertion_style.NonNullableTypeAssertionStyleRule,
		only_throw_error.OnlyThrowErrorRule,
		prefer_promise_reject_errors.PreferPromiseRejectErrorsRule,
		prefer_reduce_type_parameter.PreferReduceTypeParameterRule,
		prefer_return_this_type.PreferReturnThisTypeRule,
		promise_function_async.PromiseFunctionAsyncRule,
		related_getter_setter_pairs.RelatedGetterSetterPairsRule,
		require_array_sort_compare.RequireArraySortCompareRule,
		require_await.RequireAwaitRule,
		restrict_plus_operands.RestrictPlusOperandsRule,
		restrict_template_expressions.RestrictTemplateExpressionsRule,
		return_await.ReturnAwaitRule,
		switch_exhaustiveness_check.SwitchExhaustivenessCheckRule,
		unbound_method.UnboundMethodRule,
		use_unknown_in_catch_callback_variable.UseUnknownInCatchCallbackVariableRule,
	}

	// Create compiler host
	host := utils.CreateCompilerHost(currentDirectory, fs)

	// Create program
	program, err := utils.CreateProgram(false, fs, currentDirectory, configFileName, host)
	if err != nil {
		return nil, fmt.Errorf("error creating TS program: %v", err)
	}

	// Find source files
	files := []*ast.SourceFile{}

	// If specific files are provided, use those

	// Otherwise use all source files
	fmt.Fprintf(os.Stderr, "uri: %v", uri)
	for _, file := range program.SourceFiles() {

		p := string(file.Path())
		fmt.Fprintf(os.Stderr, "file: %v ", p)
		// FIXME: should filter file
		files = append(files, file)

	}
	fmt.Fprintf(os.Stderr, "files: %v", files)

	slices.SortFunc(files, func(a *ast.SourceFile, b *ast.SourceFile) int {
		return len(b.Text()) - len(a.Text())
	})

	// Collect diagnostics
	var diagnostics []rule.RuleDiagnostic
	var diagnosticsLock sync.Mutex
	errorsCount := 0

	// Create collector function
	diagnosticCollector := func(d rule.RuleDiagnostic) {
		diagnosticsLock.Lock()
		defer diagnosticsLock.Unlock()
		diagnostics = append(diagnostics, d)
		errorsCount++
	}

	// Run linter
	err = linter.RunLinter(
		program,
		false, // Don't use single-threaded mode for IPC
		files,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			return utils.Map(rules, func(r rule.Rule) linter.ConfiguredRule {
				return linter.ConfiguredRule{
					Name: r.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.Run(ctx, nil)
					},
				}
			})
		},
		diagnosticCollector,
	)
	if err != nil {
		return nil, fmt.Errorf("error running linter: %v", err)
	}
	if diagnostics == nil {
		diagnostics = []rule.RuleDiagnostic{}
	}
	// Create response
	return diagnostics, nil
}
