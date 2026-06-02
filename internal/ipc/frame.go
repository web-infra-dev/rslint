// Package ipc is the task-agnostic bidirectional IPC transport between the
// Go process and a Node peer.
//
// Wire format (mirrors the Node-side IpcClient in
// packages/rslint/src/ipc/client.ts):
//
//	[4 bytes u32 LE length][JSON body]
//	body = Message{kind, id, data}
//
// `data` is opaque to the transport (json.RawMessage). Application layers
// marshal/unmarshal their own typed payloads at the task boundary — the
// transport never inspects task content. The two ends are deliberately
// not import-coupled; the contract is pinned by the cross-language tests
// rather than a shared type module.
package ipc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// maxFrameSize caps a single frame body. A length header beyond this is
// treated as a stream desync (unframed bytes shifted the 4-byte header by
// N), not a real payload — without the cap a malformed uint32 (up to
// 4 GiB) makes ReadFrame allocate unboundedly and OOM the process. Matched
// to the Node side's MAX_FRAME_BYTES.
const maxFrameSize = 256 * 1024 * 1024 // 256 MiB

// MessageKind identifies a frame's purpose. The transport only owns the
// protocol-level kinds below; application kinds (e.g. task dispatch,
// output, log) are declared by the layers above and travel through the
// same opaque envelope.
type MessageKind string

const (
	// KindResponse replies to a request frame (carries the handler result).
	KindResponse MessageKind = "response"
	// KindError replies to a request frame with a failure (ErrorResponseData).
	KindError MessageKind = "error"
	// KindHandshake is the initial version-negotiation exchange.
	KindHandshake MessageKind = "handshake"
	// KindExit requests termination.
	KindExit MessageKind = "exit"
)

// Message is one decoded wire frame. `ID` is 0 for notifications and a
// positive monotonic integer for requests/responses. `Data` is the
// untyped payload — handlers decode it into a typed shape as needed.
type Message struct {
	Kind MessageKind     `json:"kind"`
	ID   int             `json:"id"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Decode unmarshals the message's Data into v.
func (m *Message) Decode(v any) error {
	if len(m.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(m.Data, v); err != nil {
		return fmt.Errorf("ipc: decode message data (kind=%s): %w", m.Kind, err)
	}
	return nil
}

// ErrorResponseData is the canonical body of an `error` frame. Mirrors the
// Node side's ErrorResponseData.
type ErrorResponseData struct {
	Message string `json:"message"`
}

// NewMessage marshals payload into a Message with the given kind and id. A
// nil payload (untyped nil OR a typed-nil pointer/interface) omits the data
// field entirely — matching the Node side, where `undefined` data is dropped
// by JSON.stringify — so "no payload" is wire-identical on both ends, not
// Go's `null` vs Node's omitted.
func NewMessage(kind MessageKind, id int, payload any) (*Message, error) {
	if isNilPayload(payload) {
		return &Message{Kind: kind, ID: id}, nil
	}
	raw, err := marshalJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("ipc: marshal payload (kind=%s): %w", kind, err)
	}
	return &Message{Kind: kind, ID: id, Data: raw}, nil
}

// isNilPayload reports whether payload should omit the data field: an untyped
// nil, or a typed-nil pointer/interface/map/slice/chan/func (e.g. a handler
// returning `(*LintResponse)(nil)` or a nil `map[string]any`). Without this a
// typed-nil value marshals to `data:null` instead of being omitted, diverging
// from Node, where `undefined` data is dropped by JSON.stringify. (An EMPTY
// but non-nil map/slice still marshals to `{}`/`[]` — only the nil case is a
// "no payload" omission; reflect.Value.IsNil distinguishes the two.)
func isNilPayload(payload any) bool {
	if payload == nil {
		return true
	}
	switch rv := reflect.ValueOf(payload); rv.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}

// marshalJSON encodes v WITHOUT Go's default HTML escaping (`<` `>` `&` →
// `<` …) so the bytes match Node's JSON.stringify, which emits those
// characters literally. Lint diagnostics routinely contain `<`/`>`/`&` (JSX,
// generics like `Foo<Bar>`, `&&`); escaping would diverge the wire from Node
// and break byte-level frame compatibility.
func marshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	// Encoder.Encode appends a trailing newline; drop it for framing.
	b := buf.Bytes()
	if n := len(b); n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}
	return b, nil
}

// ReadFrame reads one length-prefixed frame from r and decodes its
// Message. Returns io.EOF (unwrapped) on a clean stream close so callers
// can distinguish it from a transport fault.
func ReadFrame(r *bufio.Reader) (*Message, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		// Propagate io.EOF unwrapped: a clean close is not an error.
		return nil, err
	}
	if length > maxFrameSize {
		return nil, fmt.Errorf(
			"ipc: frame length %d exceeds cap %d (likely stream desync)",
			length, maxFrameSize)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("ipc: read frame body (len=%d): %w", length, err)
	}
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("ipc: decode frame (len=%d): %w", length, err)
	}
	return &msg, nil
}

// WriteFrame encodes msg into the wire format and writes it to w. Callers
// that share a writer across goroutines must serialize WriteFrame calls
// (Channel does this via its write mutex).
func WriteFrame(w io.Writer, msg *Message) error {
	body, err := marshalJSON(msg)
	if err != nil {
		return fmt.Errorf("ipc: encode frame (kind=%s): %w", msg.Kind, err)
	}
	if len(body) > maxFrameSize {
		return fmt.Errorf("ipc: frame body %d exceeds cap %d", len(body), maxFrameSize)
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(body))); err != nil {
		return fmt.Errorf("ipc: write frame length: %w", err)
	}
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("ipc: write frame body: %w", err)
	}
	return nil
}
