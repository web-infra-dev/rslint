// Package api is the single-direction programmatic IPC service used by
// `--api` mode (consumed by packages/rslint-wasm and packages/rslint-api).
// The peer (a Node parent or a wasm host) sends lint/applyFixes/getAstInfo
// requests; this service answers them. It does NOT dispatch tasks back to
// the peer — that bidirectional path is internal/ipc.Channel.
//
// Framing is shared with internal/ipc (the single source of the
// length-prefixed-JSON wire format); this package only owns the
// application-level request/response types and the inbound dispatch.
package api

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/microsoft/typescript-go/shim/api/encoder"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/inspector"
	"github.com/web-infra-dev/rslint/internal/ipc"
)

// Re-export types from inspector package for backward compatibility
type (
	GetAstInfoRequest  = inspector.GetAstInfoRequest
	GetAstInfoResponse = inspector.GetAstInfoResponse
	NodeInfo           = inspector.NodeInfo
	NodeListMeta       = inspector.NodeListMeta
	TypeInfo           = inspector.TypeInfo
	IndexInfoType      = inspector.IndexInfoType
	SymbolInfo         = inspector.SymbolInfo
	SignatureInfo      = inspector.SignatureInfo
	TypePredicateInfo  = inspector.TypePredicateInfo
	FlowInfo           = inspector.FlowInfo
)

// AstInfoBuilder wraps inspector.Builder for backward compatibility
type AstInfoBuilder = inspector.Builder

// NewAstInfoBuilder creates a new AST info builder (backward compatible)
func NewAstInfoBuilder(c *checker.Checker, sf *ast.SourceFile) *AstInfoBuilder {
	return inspector.NewBuilder(c, sf)
}

// Application-level message kinds handled by this service. The transport-
// level kinds (response/error/handshake/exit) live in internal/ipc.
const (
	// KindLint is sent from JS to Go to request linting.
	KindLint ipc.MessageKind = "lint"
	// KindApplyFixes is sent from JS to Go to request applying fixes.
	KindApplyFixes ipc.MessageKind = "applyFixes"
	// KindGetAstInfo is sent from JS to Go to request AST info at a position.
	KindGetAstInfo ipc.MessageKind = "getAstInfo"
)

// Version is the IPC protocol version.
const Version = "1.0.0"

// HandshakeRequest represents a handshake request
type HandshakeRequest struct {
	Version string `json:"version"`
}

// HandshakeResponse represents a handshake response
type HandshakeResponse struct {
	Version string `json:"version"`
	OK      bool   `json:"ok"`
}

// LintRequest represents a lint request from JS to Go
type LintRequest struct {
	Files            []string `json:"files,omitempty"`
	Config           string   `json:"config,omitempty"` // Path to rslint.json config file
	Format           string   `json:"format,omitempty"`
	WorkingDirectory string   `json:"workingDirectory,omitempty"`
	// Supports both string level and array [level, options] format
	RuleOptions               map[string]interface{} `json:"ruleOptions,omitempty"`
	FileContents              map[string]string      `json:"fileContents,omitempty"`              // Map of file paths to their contents for VFS
	LanguageOptions           *LanguageOptions       `json:"languageOptions,omitempty"`           // Override languageOptions from config file
	IncludeEncodedSourceFiles bool                   `json:"includeEncodedSourceFiles,omitempty"` // Whether to include encoded source files in response
}

// LanguageOptions contains language-specific configuration options
type LanguageOptions struct {
	ParserOptions *ParserOptions `json:"parserOptions,omitempty"`
}

// ProjectPaths represents project paths that can be either a single string or an array of strings
type ProjectPaths []string

// UnmarshalJSON implements custom JSON unmarshaling to support both string and string[] formats
func (p *ProjectPaths) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var singlePath string
	if err := json.Unmarshal(data, &singlePath); err == nil {
		*p = []string{singlePath}
		return nil
	}

	// If that fails, try to unmarshal as array of strings
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return err
	}
	*p = paths
	return nil
}

// ParserOptions contains parser-specific configuration
type ParserOptions struct {
	ProjectService bool         `json:"projectService"`
	Project        ProjectPaths `json:"project,omitempty"`
}
type ByteArray []byte

// LintResponse represents a lint response from Go to JS
type LintResponse struct {
	Diagnostics        []Diagnostic         `json:"diagnostics"`
	ErrorCount         int                  `json:"errorCount"`
	FileCount          int                  `json:"fileCount"`
	RuleCount          int                  `json:"ruleCount"`
	EncodedSourceFiles map[string]ByteArray `json:"encodedSourceFiles,omitempty"`
}

// ApplyFixesRequest represents a request to apply fixes from JS to Go
type ApplyFixesRequest struct {
	FileContent string       `json:"fileContent"` // Current content of the file
	Diagnostics []Diagnostic `json:"diagnostics"` // Diagnostics with fixes to apply
}

// ApplyFixesResponse represents a response after applying fixes
type ApplyFixesResponse struct {
	FixedContent   []string `json:"fixedContent"`   // The content after applying fixes (array of intermediate versions)
	WasFixed       bool     `json:"wasFixed"`       // Whether any fixes were actually applied
	AppliedCount   int      `json:"appliedCount"`   // Number of fixes that were applied
	UnappliedCount int      `json:"unappliedCount"` // Number of fixes that couldn't be applied
}

// Position represents a position in a file
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Range represents a position range in a file
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Diagnostic represents a single lint diagnostic
type Diagnostic struct {
	RuleName  string `json:"ruleName"`
	Message   string `json:"message"`
	FilePath  string `json:"filePath"`
	Range     Range  `json:"range"`
	Severity  string `json:"severity,omitempty"`
	MessageId string `json:"messageId"`
	Fixes     []Fix  `json:"fixes,omitempty"`
}

// Fix represents a single fix that can be applied
type Fix struct {
	Text     string `json:"text"`
	StartPos int    `json:"startPos"` // Character position in the file content
	EndPos   int    `json:"endPos"`   // Character position in the file content
}

// Handler defines the interface for handling IPC messages
type Handler interface {
	HandleLint(req LintRequest) (*LintResponse, error)
	HandleApplyFixes(req ApplyFixesRequest) (*ApplyFixesResponse, error)
	HandleGetAstInfo(req GetAstInfoRequest) (*GetAstInfoResponse, error)
}

// Service manages the single-direction IPC communication for `--api` mode.
// Framing is delegated to internal/ipc (the shared wire format).
type Service struct {
	reader  *bufio.Reader
	writer  io.Writer
	handler Handler
	writeMu sync.Mutex
}

// NewService creates a new IPC service
func NewService(reader io.Reader, writer io.Writer, handler Handler) *Service {
	return &Service{
		reader:  bufio.NewReader(reader),
		writer:  writer,
		handler: handler,
	}
}

// Start starts the IPC service
func (s *Service) Start() error {
	for {
		msg, err := ipc.ReadFrame(s.reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch msg.Kind {
		case ipc.KindHandshake:
			s.handleHandshake(msg)
		case KindLint:
			s.handleLint(msg)
		case KindApplyFixes:
			s.handleApplyFixes(msg)
		case KindGetAstInfo:
			s.handleGetAstInfo(msg)
		case ipc.KindExit:
			s.handleExit(msg)
			return nil
		default:
			s.sendError(msg.ID, fmt.Sprintf("unknown message kind: %s", msg.Kind))
		}
	}
}

// handleHandshake handles handshake messages
func (s *Service) handleHandshake(msg *ipc.Message) {
	var req HandshakeRequest
	if err := msg.Decode(&req); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("failed to parse handshake request: %v", err))
		return
	}

	s.sendResponse(msg.ID, HandshakeResponse{
		Version: Version,
		OK:      true,
	})
}

// Handle exit message
func (s *Service) handleExit(msg *ipc.Message) {
	s.sendResponse(msg.ID, nil)
}

// handleLint handles lint messages
func (s *Service) handleLint(msg *ipc.Message) {
	var req LintRequest
	if err := msg.Decode(&req); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("failed to parse lint request: %v", err))
		return
	}
	resp, err := s.handler.HandleLint(req)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	s.sendResponse(msg.ID, resp)
}

// handleApplyFixes handles apply fixes messages
func (s *Service) handleApplyFixes(msg *ipc.Message) {
	var req ApplyFixesRequest
	if err := msg.Decode(&req); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("failed to parse apply fixes request: %v", err))
		return
	}

	resp, err := s.handler.HandleApplyFixes(req)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	s.sendResponse(msg.ID, resp)
}

// handleGetAstInfo handles get AST info messages
func (s *Service) handleGetAstInfo(msg *ipc.Message) {
	var req GetAstInfoRequest
	if err := msg.Decode(&req); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("failed to parse get ast info request: %v", err))
		return
	}

	resp, err := s.handler.HandleGetAstInfo(req)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	s.sendResponse(msg.ID, resp)
}

// sendResponse sends a response message
func (s *Service) sendResponse(id int, data interface{}) {
	msg, err := ipc.NewMessage(ipc.KindResponse, id, data)
	if err != nil {
		s.sendError(id, fmt.Sprintf("failed to marshal response: %v", err))
		return
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if err := ipc.WriteFrame(s.writer, msg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send response: %v\n", err)
	}
}

// sendError sends an error message
func (s *Service) sendError(id int, message string) {
	msg, _ := ipc.NewMessage(ipc.KindError, id, ipc.ErrorResponseData{Message: message})
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if err := ipc.WriteFrame(s.writer, msg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send error: %v\n", err)
	}
}

// IsIPCMode returns true if the process is in IPC mode
func IsIPCMode() bool {
	return os.Getenv("RSLINT_IPC") == "1"
}

func EncodeAST(sourceFile *ast.SourceFile, id string) ([]byte, error) {
	data, _, err := encoder.EncodeSourceFile(sourceFile)
	return data, err
}
