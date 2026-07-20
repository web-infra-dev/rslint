// Package api is the programmatic IPC service used by `--api` mode (consumed
// by packages/rslint-wasm and packages/rslint-api). Most requests flow from
// the host to Go; lint may also dispatch plugin work back to a capable host.
//
// Framing is shared with internal/ipc (the single source of the
// length-prefixed-JSON wire format); this package only owns the
// application-level request/response types and the inbound dispatch.
package api

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

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
	// KindPluginLint is sent from Go to JS to run community ESLint plugin rules.
	KindPluginLint ipc.MessageKind = "pluginLint"
	// KindLoadConfigs asks a capable Node host to evaluate one Go-discovered
	// config frontier.
	KindLoadConfigs ipc.MessageKind = "loadConfigs"
	// KindActivateConfigs asks Node to validate and prepare the final effective
	// config/plugin generation before Go commits it.
	KindActivateConfigs ipc.MessageKind = "activateConfigs"
)

// Version is the IPC protocol version.
const Version = "2.0.0"

const CapabilityReversePluginLint = "reversePluginLint"
const CapabilityReverseConfigLoad = "reverseConfigLoadV1"

// HandshakeRequest represents a handshake request
type HandshakeRequest struct {
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// HandshakeResponse represents a handshake response
type HandshakeResponse struct {
	Version      string   `json:"version"`
	OK           bool     `json:"ok"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// LintRequest represents a lint request from JS to Go
type LintRequest struct {
	Files []string `json:"files,omitempty"`
	// CanonicalFiles is parallel to Files when the host already resolved physical
	// identity. Go uses these paths for this request instead of repeating realpath
	// calls; omitted by lower-level clients that have no pre-resolved identity.
	CanonicalFiles []string `json:"canonicalFiles,omitempty"`
	// Config is the low-level, already-resolved RslintConfig path used by WASM
	// and custom service clients. High-level native callers normally leave it
	// empty and use ConfigDiscovery: Go then discovers candidates/ownership and
	// asks the bidirectional Node host only to evaluate and normalize exact
	// JS/TS modules. Empty/absent without ConfigDiscovery means "no config"
	// (zero rules). Rules and languageOptions live in the entries; there is no
	// separate ruleOptions/languageOptions override surface.
	Config json.RawMessage `json:"config,omitempty"`
	// ConfigDiscovery enables the high-level host-filesystem path. It is
	// mutually exclusive with Config and intentionally unsupported by WASM.
	ConfigDiscovery *ConfigDiscoveryRequest `json:"configDiscovery,omitempty"`
	// Anchor directory for resolving the config's relative
	// files / ignores / parserOptions.project. Defaults to the working dir.
	ConfigDirectory string `json:"configDirectory,omitempty"`
	// PluginConfigDirectory is the opaque worker routing key for community
	// plugins. It can differ from ConfigDirectory when overrideConfig rebases
	// authored path patterns to the API cwd.
	PluginConfigDirectory string            `json:"pluginConfigDirectory,omitempty"`
	WorkingDirectory      string            `json:"workingDirectory,omitempty"`
	FileContents          map[string]string `json:"fileContents,omitempty"` // Map of file paths to their contents for VFS
	// EslintPlugins carries the names Go must register as Node-dispatched rule
	// placeholders. The live plugin implementations remain in the JS host.
	EslintPlugins []EslintPluginEntry `json:"eslintPlugins,omitempty"`
	// Fix, when true, applies rule auto-fixes in-band and returns the fixed
	// source per file in LintResponse.Output (ESLint's `fix: true`). The fix is
	// computed but NOT written to disk — the JS side (Rslint.outputFixes) writes
	// it. Diagnostics describe the original input; callers can lint Output again
	// when they need post-fix diagnostics.
	Fix                       bool `json:"fix,omitempty"`
	IncludeEncodedSourceFiles bool `json:"includeEncodedSourceFiles,omitempty"` // Whether to include encoded source files in response
}

// ConfigDiscoveryRequest is the API-facing scope for Go's shared staged
// discovery coordinator. Files themselves remain in LintRequest.Files.
type ConfigDiscoveryRequest struct {
	ExplicitConfigPath string `json:"explicitConfigPath,omitempty"`
	// Directories are static roots for the already-expanded Files set. Go limits
	// config discovery below them to branches that can govern those files.
	Directories    []string        `json:"directories,omitempty"`
	ExplicitFiles  []bool          `json:"explicitFiles,omitempty"`
	OverrideConfig json.RawMessage `json:"overrideConfig,omitempty"`
}

// EslintPluginEntry is the wire metadata for one object-form ESLint plugin.
// Prefix is the config mount name; RuleNames are names relative to the prefix.
type EslintPluginEntry struct {
	Prefix    string   `json:"prefix"`
	RuleNames []string `json:"ruleNames"`
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
	// using each caller-visible target path relative to the config directory.
	// This is the same path space as Diagnostic.FilePath and Output. The JS side
	// seeds one LintResult per entry, so ignored glob matches yield no phantom
	// results. Present for lintFiles; lintText seeds its own explicit path.
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

// Requester is the reverse-RPC capability exposed to a bidirectional lint
// handler. ipc.Channel implements it directly.
type Requester interface {
	SendRequest(ctx context.Context, kind ipc.MessageKind, payload any) (*ipc.Message, error)
}

// PeerCapabilityRequester augments Requester with the capabilities declared by
// the remote peer during the API handshake. Direct/embedded callers may keep
// implementing Requester alone; the API Service always supplies this richer
// view so a handler can validate capabilities learned only after a reverse
// operation (for example object-form plugins discovered from a config module).
type PeerCapabilityRequester interface {
	Requester
	PeerSupportsCapability(capability string) bool
}

type serviceRequester struct {
	Requester
	peerCapabilities map[string]struct{}
}

func (requester serviceRequester) PeerSupportsCapability(capability string) bool {
	_, ok := requester.peerCapabilities[capability]
	return ok
}

// BidirectionalHandler is an optional extension to Handler. Existing handlers
// continue to receive HandleLint; handlers that implement this interface also
// receive the request context and reverse-RPC transport.
type BidirectionalHandler interface {
	HandleLintWithContext(ctx context.Context, req LintRequest, requester Requester) (*LintResponse, error)
}

// Service manages bidirectional IPC communication for `--api` mode. Framing,
// request multiplexing, and the continuously running read loop are delegated
// to internal/ipc.Channel.
type Service struct {
	handler Handler
	channel *ipc.Channel
	reader  *observedReader

	// Preserve the old service's one-at-a-time inbound handling. Channel keeps
	// reading while this lock is held, so reverse responses are still routed and
	// a lint handler can synchronously await pluginLint without deadlocking.
	handlerMu        sync.Mutex
	handshakeOK      bool
	peerCapabilities map[string]struct{}

	exitMu        sync.Mutex
	exitRequestID int
	exitRequested bool
	exitAck       chan struct{}
	exitAckOnce   sync.Once
}

// NewService creates a new IPC service
func NewService(reader io.Reader, writer io.Writer, handler Handler) *Service {
	s := &Service{
		handler: handler,
		reader:  &observedReader{reader: reader},
		exitAck: make(chan struct{}),
	}
	observedWriter := &frameObserverWriter{
		writer:        writer,
		remaining:     -1,
		shouldCapture: s.exitHasBeenRequested,
		onFrame:       s.observeOutboundFrame,
	}
	s.channel = ipc.NewChannel(s.reader, observedWriter)
	s.channel.SetInboundHandler(s.handleInbound)
	return s
}

// Start starts the IPC service
func (s *Service) Start() error {
	s.channel.Start()
	select {
	case <-s.exitAck:
		// The complete exit response frame has reached the underlying writer.
		// Closing earlier can race Channel's asynchronous inbound handler and
		// suppress the ack that legacy Node/browser clients await.
		_ = s.channel.Close()
		return nil
	case <-s.channel.Done():
		err := s.reader.Err()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		return errors.New("api: IPC transport closed unexpectedly")
	}
}

func (s *Service) handleInbound(ctx context.Context, msg *ipc.Message) (any, error) {
	s.handlerMu.Lock()
	defer s.handlerMu.Unlock()

	switch msg.Kind {
	case ipc.KindHandshake:
		var req HandshakeRequest
		if err := msg.Decode(&req); err != nil {
			return nil, fmt.Errorf("failed to parse handshake request: %w", err)
		}
		s.handshakeOK = req.Version == Version
		s.peerCapabilities = make(map[string]struct{}, len(req.Capabilities))
		for _, capability := range req.Capabilities {
			s.peerCapabilities[capability] = struct{}{}
		}
		var capabilities []string
		if _, ok := s.handler.(BidirectionalHandler); ok {
			capabilities = []string{
				CapabilityReversePluginLint,
				CapabilityReverseConfigLoad,
			}
		}
		return HandshakeResponse{
			Version:      Version,
			OK:           s.handshakeOK,
			Capabilities: capabilities,
		}, nil

	case KindLint:
		if !s.handshakeOK {
			return nil, fmt.Errorf("API protocol handshake required (expected version %s)", Version)
		}
		var req LintRequest
		if err := msg.Decode(&req); err != nil {
			return nil, fmt.Errorf("failed to parse lint request: %w", err)
		}
		handler, bidirectional := s.handler.(BidirectionalHandler)
		if len(req.EslintPlugins) > 0 {
			if !bidirectional {
				return nil, errors.New("API handler does not support reversePluginLint requests")
			}
			if _, ok := s.peerCapabilities[CapabilityReversePluginLint]; !ok {
				return nil, errors.New("API peer does not advertise reversePluginLint capability")
			}
		}
		if req.ConfigDiscovery != nil {
			if !bidirectional {
				return nil, errors.New("API handler does not support reverseConfigLoad requests")
			}
			if _, ok := s.peerCapabilities[CapabilityReverseConfigLoad]; !ok {
				return nil, errors.New("API peer does not advertise reverseConfigLoadV1 capability")
			}
		}
		if bidirectional {
			return handler.HandleLintWithContext(ctx, req, serviceRequester{
				Requester:        s.channel,
				peerCapabilities: s.peerCapabilities,
			})
		}
		return s.handler.HandleLint(req)

	case KindGetAstInfo:
		if !s.handshakeOK {
			return nil, fmt.Errorf("API protocol handshake required (expected version %s)", Version)
		}
		var req GetAstInfoRequest
		if err := msg.Decode(&req); err != nil {
			return nil, fmt.Errorf("failed to parse get ast info request: %w", err)
		}
		return s.handler.HandleGetAstInfo(req)

	case ipc.KindExit:
		s.exitMu.Lock()
		if !s.exitRequested {
			s.exitRequested = true
			s.exitRequestID = msg.ID
		}
		s.exitMu.Unlock()
		return struct{}{}, nil

	default:
		return nil, fmt.Errorf("unknown message kind: %s", msg.Kind)
	}
}

func (s *Service) observeOutboundFrame(msg *ipc.Message) {
	if msg.Kind != ipc.KindResponse && msg.Kind != ipc.KindError {
		return
	}
	s.exitMu.Lock()
	isExitAck := s.exitRequested && msg.ID == s.exitRequestID
	s.exitMu.Unlock()
	if isExitAck {
		s.exitAckOnce.Do(func() { close(s.exitAck) })
	}
}

func (s *Service) exitHasBeenRequested() bool {
	s.exitMu.Lock()
	defer s.exitMu.Unlock()
	return s.exitRequested
}

// observedReader records the terminal read error so Service.Start can retain
// the legacy clean-EOF behavior while Channel owns frame decoding. It tracks
// only frame byte counts (never payload copies), allowing it to distinguish a
// clean frame-boundary EOF from a truncated frame.
type observedReader struct {
	reader    io.Reader
	mu        sync.Mutex
	err       error
	header    [4]byte
	headerLen int
	remaining uint32
	inBody    bool
}

func (r *observedReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.mu.Lock()
	r.observe(p[:n])
	if err != nil {
		if errors.Is(err, io.EOF) && (r.headerLen != 0 || r.inBody) {
			r.err = io.ErrUnexpectedEOF
		} else {
			r.err = err
		}
	}
	r.mu.Unlock()
	return n, err
}

func (r *observedReader) observe(data []byte) {
	for len(data) > 0 {
		if !r.inBody {
			n := min(4-r.headerLen, len(data))
			copy(r.header[r.headerLen:], data[:n])
			r.headerLen += n
			data = data[n:]
			if r.headerLen < 4 {
				continue
			}
			r.remaining = binary.LittleEndian.Uint32(r.header[:])
			r.headerLen = 0
			if r.remaining == 0 {
				continue
			}
			r.inBody = true
		}

		n := len(data)
		if uint64(n) > uint64(r.remaining) {
			n = int(r.remaining)
		}
		r.remaining -= uint32(n)
		data = data[n:]
		if r.remaining == 0 {
			r.inBody = false
		}
	}
}

func (r *observedReader) Err() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.err
}

// frameObserverWriter observes complete outbound frames after their bytes have
// been accepted by the real writer. It normally tracks lengths only; after an
// exit request arrives it captures response bodies until the matching ack is
// seen. Channel serializes all calls into it.
type frameObserverWriter struct {
	writer        io.Writer
	shouldCapture func() bool
	onFrame       func(*ipc.Message)
	mu            sync.Mutex
	header        [4]byte
	headerLen     int
	remaining     int
	capture       bool
	body          []byte
}

func (w *frameObserverWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	if n != len(p) && err == nil {
		err = io.ErrShortWrite
	}
	if n <= 0 {
		return n, err
	}

	w.mu.Lock()
	var completed []*ipc.Message
	data := p[:n]
	for len(data) > 0 {
		if w.remaining < 0 {
			headerBytes := min(4-w.headerLen, len(data))
			copy(w.header[w.headerLen:], data[:headerBytes])
			w.headerLen += headerBytes
			data = data[headerBytes:]
			if w.headerLen < 4 {
				continue
			}
			w.remaining = int(binary.LittleEndian.Uint32(w.header[:]))
			w.headerLen = 0
			w.capture = w.shouldCapture()
			if w.capture {
				w.body = make([]byte, 0, w.remaining)
			}
		}

		bodyBytes := min(w.remaining, len(data))
		if w.capture {
			w.body = append(w.body, data[:bodyBytes]...)
		}
		w.remaining -= bodyBytes
		data = data[bodyBytes:]
		if w.remaining == 0 {
			if w.capture {
				var msg ipc.Message
				if json.Unmarshal(w.body, &msg) == nil {
					completed = append(completed, &msg)
				}
			}
			w.remaining = -1
			w.capture = false
			w.body = nil
		}
	}
	w.mu.Unlock()

	if err == nil {
		for _, msg := range completed {
			w.onFrame(msg)
		}
	}
	return n, err
}

// Preserve Channel's write-deadline support when the underlying writer is an
// *os.File or net.Conn. Channel intentionally ignores unsupported deadlines.
func (w *frameObserverWriter) SetWriteDeadline(deadline time.Time) error {
	if writer, ok := w.writer.(interface {
		SetWriteDeadline(deadline time.Time) error
	}); ok {
		return writer.SetWriteDeadline(deadline)
	}
	return nil
}

// IsIPCMode returns true if the process is in IPC mode
func IsIPCMode() bool {
	return os.Getenv("RSLINT_IPC") == "1"
}

func EncodeAST(sourceFile *ast.SourceFile, id string) ([]byte, error) {
	data, _, err := encoder.EncodeSourceFile(sourceFile)
	return data, err
}
