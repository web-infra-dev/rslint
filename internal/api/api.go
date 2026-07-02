// Package api is the single-direction programmatic IPC service used by
// `--api` mode (consumed by packages/rslint-wasm and packages/rslint-api).
// The peer (a Node parent or a wasm host) sends lint/getAstInfo
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
	Files []string `json:"files,omitempty"`
	// Final resolved config — a serialized RslintConfig (RslintConfigEntry[]).
	// The JS side does ALL of overrideConfig / config-file / auto-discovery /
	// normalize and hands Go only this object; --api never reads config from
	// disk. Empty/absent means "no config" (zero rules). Rules AND
	// languageOptions live in the config entries — there is no separate
	// ruleOptions / languageOptions override surface.
	Config json.RawMessage `json:"config,omitempty"`
	// Anchor directory for resolving the config's relative
	// files / ignores / parserOptions.project. Defaults to the working dir.
	ConfigDirectory  string            `json:"configDirectory,omitempty"`
	WorkingDirectory string            `json:"workingDirectory,omitempty"`
	FileContents     map[string]string `json:"fileContents,omitempty"` // Map of file paths to their contents for VFS
	// Fix, when true, applies rule auto-fixes in-band and returns the fixed
	// source per file in LintResponse.Output (ESLint's `fix: true`). The fix is
	// computed but NOT written to disk — the JS side (Rslint.outputFixes) writes
	// it. Remaining (unfixed) diagnostics are still reported.
	Fix                       bool `json:"fix,omitempty"`
	IncludeEncodedSourceFiles bool `json:"includeEncodedSourceFiles,omitempty"` // Whether to include encoded source files in response
}
type ByteArray []byte

// LintResponse represents a lint response from Go to JS
type LintResponse struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	// ErrorCount / WarningCount are split by severity (ESLint semantics):
	// ErrorCount counts only error-severity diagnostics, NOT the total.
	ErrorCount   int `json:"errorCount"`
	WarningCount int `json:"warningCount"`
	// FixableErrorCount / FixableWarningCount count diagnostics that carry an
	// auto-fix, split by severity (ESLint LintResult.fixable*Count).
	FixableErrorCount   int `json:"fixableErrorCount"`
	FixableWarningCount int `json:"fixableWarningCount"`
	FileCount           int `json:"fileCount"`
	RuleCount           int `json:"ruleCount"`
	// LintedFiles lists the files actually linted (config `ignores` excluded),
	// each the program's canonical path relative to the config directory — the
	// same path space as Diagnostic.FilePath. The JS side seeds one LintResult
	// per entry, so a glob match the config ignores yields no phantom result,
	// and a symlinked glob path can't duplicate a result. Present for lintFiles;
	// lintText seeds from its own explicit path instead.
	//
	// MUST NOT be omitempty: an all-ignored lint produces an empty (non-nil)
	// slice that has to serialize as `[]`, distinct from an old binary that
	// omits the field entirely. The JS glob-fallback keys on the field's
	// ABSENCE, so collapsing empty→absent would re-seed phantom empty results.
	LintedFiles []string `json:"lintedFiles"`
	// Output holds the fixed source per file, present only when Fix was
	// requested and at least one fix applied (ESLint LintResult.output, but
	// keyed by file path since one lint covers many files). The JS side writes
	// these back via Rslint.outputFixes.
	Output             map[string]string    `json:"output,omitempty"`
	EncodedSourceFiles map[string]ByteArray `json:"encodedSourceFiles,omitempty"`
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
	RuleName    string       `json:"ruleName"`
	Message     string       `json:"message"`
	FilePath    string       `json:"filePath"`
	Range       Range        `json:"range"`
	Severity    string       `json:"severity,omitempty"`
	MessageId   string       `json:"messageId"`
	Fixes       []Fix        `json:"fixes,omitempty"`
	Suggestions []Suggestion `json:"suggestions,omitempty"`
}

// Fix represents a single fix that can be applied
type Fix struct {
	Text     string `json:"text"`
	StartPos int    `json:"startPos"` // Character position in the file content
	EndPos   int    `json:"endPos"`   // Character position in the file content
}

// Suggestion is an optional, user-selected fix (ESLint's "suggestions"):
// unlike Fixes, which `fix: true` applies automatically, a suggestion is
// surfaced for the editor/user to choose. Data exposes the messageId's
// placeholder values (ESLint v10 suggestion.data).
type Suggestion struct {
	MessageId string            `json:"messageId"`
	Message   string            `json:"message"`
	Data      map[string]string `json:"data,omitempty"`
	Fixes     []Fix             `json:"fixes,omitempty"`
}

// Handler defines the interface for handling IPC messages
type Handler interface {
	HandleLint(req LintRequest) (*LintResponse, error)
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
