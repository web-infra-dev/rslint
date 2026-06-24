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

// frameBinaryFlag is the high bit of the 4-byte length header. When set, the
// frame body is not bare JSON but a binary-carrying layout (see WriteFrame):
//
//	[4B jsonLen][JSON][4B binCount]( [4B blobLen][blob] )×binCount
//
// The low 31 bits hold the body length. maxFrameSize (256 MiB) occupies only
// 28 bits, so the high bit is always free to repurpose. A frame with no Binary
// attachments keeps the original bare-JSON format (flag clear), so every other
// frame kind — and the entire LSP path — stays byte-for-byte unchanged.
const frameBinaryFlag uint32 = 1 << 31

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
	// Binary carries opaque byte blobs out of band of the JSON body, as a
	// frame trailer. The transport never inspects them — application layers
	// (e.g. pluginLint moving a file's type snapshot here) use it to keep
	// large binary payloads out of the JSON, where they would otherwise be
	// base64-inflated by ~33%. Empty ⇒ a legacy bare-JSON frame.
	Binary [][]byte `json:"-"`
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
	var raw uint32
	if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
		// Propagate io.EOF unwrapped: a clean close is not an error.
		return nil, err
	}
	hasBinary := raw&frameBinaryFlag != 0
	length := raw &^ frameBinaryFlag // clear the flag bit to recover the body length
	if length > maxFrameSize {
		return nil, fmt.Errorf(
			"ipc: frame length %d exceeds cap %d (likely stream desync)",
			length, maxFrameSize)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("ipc: read frame body (len=%d): %w", length, err)
	}
	if hasBinary {
		return decodeBinaryFrame(body)
	}
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("ipc: decode frame (len=%d): %w", length, err)
	}
	return &msg, nil
}

// decodeBinaryFrame parses a binary-carrying frame body produced by WriteFrame:
//
//	[4B jsonLen][JSON][4B binCount]( [4B blobLen][blob] )×binCount
//
// Every field is bounds-checked against the remaining body so a truncated or
// malformed trailer degrades to an error rather than an out-of-range slice. Each
// blob is copied out of the body slab so a retained Message.Binary entry does not
// pin the whole frame's backing array.
func decodeBinaryFrame(body []byte) (*Message, error) {
	rest := body
	readU32 := func() (uint32, bool) {
		if len(rest) < 4 {
			return 0, false
		}
		v := binary.LittleEndian.Uint32(rest[:4])
		rest = rest[4:]
		return v, true
	}

	jsonLen, ok := readU32()
	if !ok || int(jsonLen) > len(rest) {
		return nil, fmt.Errorf("ipc: binary frame jsonLen out of bounds")
	}
	var msg Message
	if err := json.Unmarshal(rest[:jsonLen], &msg); err != nil {
		return nil, fmt.Errorf("ipc: decode binary frame JSON: %w", err)
	}
	rest = rest[jsonLen:]

	binCount, ok := readU32()
	if !ok {
		return nil, fmt.Errorf("ipc: binary frame missing binCount")
	}
	msg.Binary = make([][]byte, 0, binCount)
	for i := uint32(0); i < binCount; i++ {
		blobLen, ok := readU32()
		if !ok {
			return nil, fmt.Errorf("ipc: binary frame truncated at blob %d header", i)
		}
		if int(blobLen) > len(rest) {
			return nil, fmt.Errorf("ipc: binary frame truncated at blob %d body (want %d, have %d)", i, blobLen, len(rest))
		}
		blob := make([]byte, blobLen)
		copy(blob, rest[:blobLen])
		msg.Binary = append(msg.Binary, blob)
		rest = rest[blobLen:]
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

	// Legacy bare-JSON frame: [4B len][JSON]. Any frame with no Binary
	// attachments (every kind other than a binary-carrying pluginLint, plus
	// the entire LSP path) takes this branch and is byte-for-byte identical
	// to the pre-binary wire format.
	if len(msg.Binary) == 0 {
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

	// Binary-carrying frame: body = [4B jsonLen][JSON][4B binCount]( [4B blobLen][blob] )×N,
	// length header OR'd with frameBinaryFlag so ReadFrame knows to split it.
	var buf bytes.Buffer
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(body)))
	buf.Write(hdr[:])
	buf.Write(body)
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(msg.Binary)))
	buf.Write(hdr[:])
	for _, blob := range msg.Binary {
		binary.LittleEndian.PutUint32(hdr[:], uint32(len(blob)))
		buf.Write(hdr[:])
		buf.Write(blob)
	}
	if buf.Len() > maxFrameSize {
		return fmt.Errorf("ipc: frame body %d exceeds cap %d", buf.Len(), maxFrameSize)
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(buf.Len())|frameBinaryFlag); err != nil {
		return fmt.Errorf("ipc: write frame length: %w", err)
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("ipc: write frame body: %w", err)
	}
	return nil
}
