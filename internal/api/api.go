// Package ipc provides IPC communication between JS and Go using stdio
package ipc

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// Protocol implements a binary message protocol similar to esbuild:
// - First 4 bytes: message length (uint32 in little endian)
// - Next N bytes: JSON message content

// MessageKind represents the kind of IPC message
type MessageKind string

const (
	// KindLint is sent from JS to Go to request linting
	KindLint MessageKind = "lint"
	// KindResponse is sent from Go to JS with the lint results
	KindResponse MessageKind = "response"
	// KindError is sent when an error occurs
	KindError MessageKind = "error"
	// KindHandshake is sent for initial connection verification
	KindHandshake MessageKind = "handshake"
	// KindExit is sent to request termination
	KindExit MessageKind = "exit"
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
	Files            []string          `json:"files,omitempty"`
	Config           string            `json:"config,omitempty"`      // Path to rslint config file
	Format           string            `json:"format,omitempty"`
	WorkingDirectory string            `json:"workingDirectory,omitempty"`
	RuleOptions      map[string]string `json:"ruleOptions,omitempty"`
	FileContents     map[string]string `json:"fileContents,omitempty"` // Map of file paths to their contents for VFS
}

// LintResponse represents a lint response from Go to JS
type LintResponse struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	ErrorCount  int          `json:"errorCount"`
	FileCount   int          `json:"fileCount"`
	RuleCount   int          `json:"ruleCount"`
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
	RuleName string `json:"ruleName"`
	Message  string `json:"message"`
	FilePath string `json:"filePath"`
	Range    Range  `json:"range"`
	Severity string `json:"severity,omitempty"`
}

// Handler defines the interface for handling IPC messages
type Handler interface {
	HandleLint(req LintRequest) (*LintResponse, error)
}

// Service manages the IPC communication
type Service struct {
	reader  *bufio.Reader
	writer  io.Writer
	handler Handler
	mutex   sync.Mutex
}

// NewService creates a new IPC service
func NewService(reader io.Reader, writer io.Writer, handler Handler) *Service {
	return &Service{
		reader:  bufio.NewReader(reader),
		writer:  writer,
		handler: handler,
	}
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
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch msg.Kind {
		case KindHandshake:
			s.handleHandshake(msg)
		case KindLint:
			s.handleLint(msg)
		case KindExit:
			s.handleExit(msg)
			return nil
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
