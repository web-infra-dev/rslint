// Package ipc provides IPC communication between JS and Go using stdio
package ipc

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/microsoft/typescript-go/shim/api/encoder"
	"github.com/microsoft/typescript-go/shim/ast"
)

// Protocol implements a binary message protocol similar to esbuild:
// - First 4 bytes: message length (uint32 in little endian)
// - Next N bytes: JSON message content

// MessageKind represents the kind of IPC message
type MessageKind string

const (
	// KindLint is sent from JS to Go to request linting
	KindLint MessageKind = "lint"
	// KindApplyFixes is sent from JS to Go to request applying fixes
	KindApplyFixes MessageKind = "applyFixes"
	// KindResponse is sent from Go to JS with the lint results
	KindResponse MessageKind = "response"
	// KindError is sent when an error occurs
	KindError MessageKind = "error"
	// KindHandshake is sent for initial connection verification
	KindHandshake MessageKind = "handshake"
	// KindExit is sent to request termination
	KindExit MessageKind = "exit"

	// Lazy loading APIs for type checker
	KindCheckerGetNodeType      MessageKind = "checker.getNodeType"
	KindCheckerGetNodeSymbol    MessageKind = "checker.getNodeSymbol"
	KindCheckerGetNodeSignature MessageKind = "checker.getNodeSignature"
	KindCheckerGetNodeFlowNode  MessageKind = "checker.getNodeFlowNode"
	KindCheckerGetNodeInfo      MessageKind = "checker.getNodeInfo"
)

// Version is the IPC protocol version
const Version = "1.0.0"

// Message represents an IPC message
type Message struct {
	Kind MessageKind `json:"kind"`
	ID   int         `json:"id"`
	Data interface{} `json:"data,omitempty"`
}

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
	IncludeTypeChecker        bool                   `json:"includeTypeChecker,omitempty"`        // Whether to create a type checker session
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
	HasTypeChecker     bool                 `json:"hasTypeChecker,omitempty"` // Whether type checker is available
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

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message"`
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
}

// CheckerHandler defines the interface for handling type checker IPC messages
type CheckerHandler interface {
	HandleCheckerGetNodeType(req CheckerNodeRequest) (*NodeTypeResponse, error)
	HandleCheckerGetNodeSymbol(req CheckerNodeRequest) (*NodeSymbolResponse, error)
	HandleCheckerGetNodeSignature(req CheckerNodeRequest) (*NodeSignatureResponse, error)
	HandleCheckerGetNodeFlowNode(req CheckerNodeRequest) (*NodeFlowNodeResponse, error)
	HandleCheckerGetNodeInfo(req CheckerNodeRequest) (*NodeInfoResponse, error)
}

// NodeLocation identifies a node in a source file using structured parameters
type NodeLocation struct {
	FilePath string `json:"filePath"` // Relative file path (e.g., "index.ts")
	Pos      int    `json:"pos"`      // Start position of the node
	Kind     int    `json:"kind"`     // SyntaxKind of the node
}

// CheckerNodeRequest represents a request to get node information from type checker
type CheckerNodeRequest struct {
	Node NodeLocation `json:"node"`
}

// TypeDetails represents detailed type information
type TypeDetails struct {
	Id                  uint32              `json:"Id"`
	Flags               uint32              `json:"Flags"`
	FlagNames           []string            `json:"FlagNames"`
	ObjectFlags         uint32              `json:"ObjectFlags"`
	ObjectFlagNames     []string            `json:"ObjectFlagNames"`
	Symbol              *uint64             `json:"Symbol,omitempty"`
	Target              *uint32             `json:"Target,omitempty"`
	Types               []uint32            `json:"Types,omitempty"`
	TypeString          string              `json:"TypeString"`
	IntrinsicName       string              `json:"IntrinsicName,omitempty"`
	Value               interface{}         `json:"Value,omitempty"`
	TypeParameters      []uint32            `json:"TypeParameters,omitempty"`
	FixedLength         *int                `json:"FixedLength,omitempty"`
	ElementInfos        []ElementInfoDetail `json:"ElementInfos,omitempty"`
	Properties          []uint64            `json:"Properties,omitempty"`
	CallSignatures      []string            `json:"CallSignatures,omitempty"`
	ConstructSignatures []string            `json:"ConstructSignatures,omitempty"`
}

// ElementInfoDetail represents tuple element info
type ElementInfoDetail struct {
	Flags uint32 `json:"Flags"`
}

// SymbolDetails represents detailed symbol information
type SymbolDetails struct {
	Id               uint64               `json:"Id"`
	Flags            uint32               `json:"Flags"`
	FlagNames        []string             `json:"FlagNames"`
	CheckFlags       uint32               `json:"CheckFlags"`
	CheckFlagNames   []string             `json:"CheckFlagNames"`
	Name             string               `json:"Name"`
	SymbolString     string               `json:"SymbolString"`
	Declarations     []NodeLocation       `json:"Declarations,omitempty"`
	ValueDeclaration *NodeLocation        `json:"ValueDeclaration,omitempty"`
	Members          map[string]uint64    `json:"Members,omitempty"`
	Exports          map[string]uint64    `json:"Exports,omitempty"`
	Parent           *uint64              `json:"Parent,omitempty"`
}

// SignatureDetails represents detailed signature information
// Note: Signature has no internal ID in typescript-go, so we don't expose one
type SignatureDetails struct {
	SignatureString  string            `json:"SignatureString"`
	TypeParameters   []uint32          `json:"TypeParameters,omitempty"`
	Parameters       []ParameterDetail `json:"Parameters"`
	ThisParameter    *ParameterDetail  `json:"ThisParameter,omitempty"`
	HasRestParameter bool              `json:"HasRestParameter"`
	ReturnType       *uint32           `json:"ReturnType,omitempty"`
	Declaration      *NodeLocation     `json:"Declaration,omitempty"`
}

// ParameterDetail represents a parameter in a signature
type ParameterDetail struct {
	Name     string `json:"Name"`
	SymbolId uint64 `json:"SymbolId"`
}

// FlowNodeDetails represents detailed flow node information
// Note: FlowNode has no internal ID in typescript-go, so we don't expose one
// Antecedent/Antecedents use internal indices for cycle prevention during collection
type FlowNodeDetails struct {
	Flags       uint32        `json:"Flags"`
	FlagNames   []string      `json:"FlagNames"`
	Node        *NodeLocation `json:"Node,omitempty"`
	Antecedent  *FlowNodeDetails `json:"Antecedent,omitempty"`
	Antecedents []*FlowNodeDetails `json:"Antecedents,omitempty"`
}

// NodeTypeResponse is the response for getNodeType
type NodeTypeResponse struct {
	Type           *TypeDetails            `json:"Type,omitempty"`
	ContextualType *TypeDetails            `json:"ContextualType,omitempty"`
	// Related objects collected during traversal (for reference lookup)
	RelatedTypes   map[uint32]TypeDetails  `json:"RelatedTypes,omitempty"`
	RelatedSymbols map[uint64]SymbolDetails `json:"RelatedSymbols,omitempty"`
}

// NodeSymbolResponse is the response for getNodeSymbol
type NodeSymbolResponse struct {
	Symbol         *SymbolDetails           `json:"Symbol,omitempty"`
	// Related objects collected during traversal (for reference lookup)
	RelatedTypes   map[uint32]TypeDetails   `json:"RelatedTypes,omitempty"`
	RelatedSymbols map[uint64]SymbolDetails `json:"RelatedSymbols,omitempty"`
}

// NodeSignatureResponse is the response for getNodeSignature
type NodeSignatureResponse struct {
	Signature      *SignatureDetails        `json:"Signature,omitempty"`
	// Related objects collected during traversal (for reference lookup)
	RelatedTypes   map[uint32]TypeDetails   `json:"RelatedTypes,omitempty"`
	RelatedSymbols map[uint64]SymbolDetails `json:"RelatedSymbols,omitempty"`
}

// NodeFlowNodeResponse is the response for getNodeFlowNode
type NodeFlowNodeResponse struct {
	FlowNode *FlowNodeDetails `json:"FlowNode,omitempty"`
}

// NodeInfoResponse is the response for getNodeInfo
type NodeInfoResponse struct {
	Kind              int      `json:"Kind"`
	KindName          string   `json:"KindName"`
	Flags             uint32   `json:"Flags"`
	FlagNames         []string `json:"FlagNames"`
	ModifierFlags     uint32   `json:"ModifierFlags"`
	ModifierFlagNames []string `json:"ModifierFlagNames"`
	Pos               int      `json:"Pos"`
	End               int      `json:"End"`
}

// Service manages the IPC communication
type Service struct {
	reader         *bufio.Reader
	writer         io.Writer
	handler        Handler
	checkerHandler CheckerHandler
	mutex          sync.Mutex
}

// NewService creates a new IPC service
func NewService(reader io.Reader, writer io.Writer, handler Handler) *Service {
	return &Service{
		reader:  bufio.NewReader(reader),
		writer:  writer,
		handler: handler,
	}
}

// SetCheckerHandler sets the checker handler for the service
func (s *Service) SetCheckerHandler(handler CheckerHandler) {
	s.checkerHandler = handler
}

// readMessage reads a message from the input
func (s *Service) readMessage() (*Message, error) {
	// Read message length (4 bytes)
	var length uint32
	if err := binary.Read(s.reader, binary.LittleEndian, &length); err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}

	// Read message content
	data := make([]byte, length)
	if _, err := io.ReadFull(s.reader, data); err != nil {
		return nil, fmt.Errorf("failed to read message content: %w", err)
	}

	// Unmarshal message
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// writeMessage writes a message to the output
func (s *Service) writeMessage(msg *Message) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Marshal message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write message length (4 bytes)
	if err := binary.Write(s.writer, binary.LittleEndian, uint32(len(data))); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// Write message content
	if _, err := s.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write message content: %w", err)
	}

	return nil
}

// Start starts the IPC service
func (s *Service) Start() error {
	for {
		msg, err := s.readMessage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch msg.Kind {
		case KindHandshake:
			s.handleHandshake(msg)
		case KindLint:
			s.handleLint(msg)
		case KindApplyFixes:
			s.handleApplyFixes(msg)
		case KindExit:
			s.handleExit(msg)
			return nil
		// Type checker messages
		case KindCheckerGetNodeType:
			s.handleCheckerGetNodeType(msg)
		case KindCheckerGetNodeSymbol:
			s.handleCheckerGetNodeSymbol(msg)
		case KindCheckerGetNodeSignature:
			s.handleCheckerGetNodeSignature(msg)
		case KindCheckerGetNodeFlowNode:
			s.handleCheckerGetNodeFlowNode(msg)
		case KindCheckerGetNodeInfo:
			s.handleCheckerGetNodeInfo(msg)
		default:
			s.sendError(msg.ID, fmt.Sprintf("unknown message kind: %s", msg.Kind))
		}
	}
}

// handleHandshake handles handshake messages
func (s *Service) handleHandshake(msg *Message) {
	var req HandshakeRequest
	data, err := json.Marshal(msg.Data)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("failed to marshal data: %v", err))
		return
	}

	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("failed to parse handshake request: %v", err))
		return
	}

	s.sendResponse(msg.ID, HandshakeResponse{
		Version: Version,
		OK:      true,
	})
}

// Handle exit message
func (s *Service) handleExit(msg *Message) {
	s.sendResponse(msg.ID, nil)
}

// handleLint handles lint messages
func (s *Service) handleLint(msg *Message) {
	var req LintRequest
	data, err := json.Marshal(msg.Data)

	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("failed to marshal data: %v", err))
		return
	}

	if err := json.Unmarshal(data, &req); err != nil {
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
func (s *Service) handleApplyFixes(msg *Message) {
	var req ApplyFixesRequest
	data, err := json.Marshal(msg.Data)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("failed to marshal data: %v", err))
		return
	}

	if err := json.Unmarshal(data, &req); err != nil {
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

// sendResponse sends a response message
func (s *Service) sendResponse(id int, data interface{}) {
	msg := &Message{
		ID:   id,
		Kind: KindResponse,
		Data: data,
	}
	if err := s.writeMessage(msg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send response: %v\n", err)
	}
}

// sendError sends an error message
func (s *Service) sendError(id int, message string) {
	msg := &Message{
		ID:   id,
		Kind: KindError,
		Data: ErrorResponse{Message: message},
	}
	if err := s.writeMessage(msg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send error: %v\n", err)
	}
}

// IsIPCMode returns true if the process is in IPC mode
func IsIPCMode() bool {
	return os.Getenv("RSLINT_IPC") == "1"
}

func EncodeAST(sourceFile *ast.SourceFile, id string) ([]byte, error) {
	return encoder.EncodeSourceFile(sourceFile, id)
}

// parseCheckerNodeRequest parses a checker node request from message data
func (s *Service) parseCheckerNodeRequest(msg *Message) (*CheckerNodeRequest, error) {
	data, err := json.Marshal(msg.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var req CheckerNodeRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse checker node request: %w", err)
	}
	return &req, nil
}

// handleCheckerGetNodeType handles checker.getNodeType messages
func (s *Service) handleCheckerGetNodeType(msg *Message) {
	if s.checkerHandler == nil {
		s.sendError(msg.ID, "checker handler not available")
		return
	}

	req, err := s.parseCheckerNodeRequest(msg)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	resp, err := s.checkerHandler.HandleCheckerGetNodeType(*req)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	s.sendResponse(msg.ID, resp)
}

// handleCheckerGetNodeSymbol handles checker.getNodeSymbol messages
func (s *Service) handleCheckerGetNodeSymbol(msg *Message) {
	if s.checkerHandler == nil {
		s.sendError(msg.ID, "checker handler not available")
		return
	}

	req, err := s.parseCheckerNodeRequest(msg)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	resp, err := s.checkerHandler.HandleCheckerGetNodeSymbol(*req)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	s.sendResponse(msg.ID, resp)
}

// handleCheckerGetNodeSignature handles checker.getNodeSignature messages
func (s *Service) handleCheckerGetNodeSignature(msg *Message) {
	if s.checkerHandler == nil {
		s.sendError(msg.ID, "checker handler not available")
		return
	}

	req, err := s.parseCheckerNodeRequest(msg)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	resp, err := s.checkerHandler.HandleCheckerGetNodeSignature(*req)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	s.sendResponse(msg.ID, resp)
}

// handleCheckerGetNodeFlowNode handles checker.getNodeFlowNode messages
func (s *Service) handleCheckerGetNodeFlowNode(msg *Message) {
	if s.checkerHandler == nil {
		s.sendError(msg.ID, "checker handler not available")
		return
	}

	req, err := s.parseCheckerNodeRequest(msg)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	resp, err := s.checkerHandler.HandleCheckerGetNodeFlowNode(*req)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	s.sendResponse(msg.ID, resp)
}

// handleCheckerGetNodeInfo handles checker.getNodeInfo messages
func (s *Service) handleCheckerGetNodeInfo(msg *Message) {
	if s.checkerHandler == nil {
		s.sendError(msg.ID, "checker handler not available")
		return
	}

	req, err := s.parseCheckerNodeRequest(msg)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	resp, err := s.checkerHandler.HandleCheckerGetNodeInfo(*req)
	if err != nil {
		s.sendError(msg.ID, err.Error())
		return
	}

	s.sendResponse(msg.ID, resp)
}
