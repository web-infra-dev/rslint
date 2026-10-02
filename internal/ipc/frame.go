// Package ipc is the task-agnostic bidirectional IPC transport between the
// Go process and a Node peer.
//
// Wire format (mirrors the Node-side IpcClient in
// packages/rslint/src/ipc/client.ts):
//
//	[4 bytes u32 LE length][JSON body]
//	body = Message{kind, id, data, attachments?, transport?}
//
// `data` is opaque to the transport (json.RawMessage). Application layers
// marshal/unmarshal their own typed payloads at the task boundary — the
// transport never inspects task content. Storage settings are exchanged at
// runtime; cross-language tests pin the fixed frame and envelope contract.
package ipc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"unicode/utf8"
)

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
	// KindTransportConfig returns storage settings before mapping bootstrap.
	KindTransportConfig MessageKind = "transportConfig"
	// KindTransportRelease acknowledges storage after a handler already replied.
	// Its ID identifies that request; it carries no application result.
	KindTransportRelease MessageKind = "transportRelease"
	// KindExit requests termination.
	KindExit MessageKind = "exit"
)

// Message is one decoded wire frame. `ID` is 0 for notifications and a
// positive monotonic integer for requests/responses; transport release frames
// reuse the owning request's ID without replying again. `Data` is the
// untyped payload — handlers decode it into a typed shape as needed.
type Message struct {
	Kind        MessageKind        `json:"kind"`
	ID          int                `json:"id"`
	Data        json.RawMessage    `json:"data,omitempty"`
	Attachments []AttachmentData   `json:"attachments,omitempty"`
	Transport   *TransportMetadata `json:"transport,omitempty"`
}

// Keep explicit null containers invalid rather than silently treating them as
// omitted attachments or transport metadata. Data remains application-owned.
func (m *Message) UnmarshalJSON(data []byte) error {
	if raw := bytes.TrimSpace(data); len(raw) == 0 || raw[0] != '{' {
		return errors.New("ipc: message envelope must be an object")
	}
	type fields Message
	*m = Message{}
	decoded := struct {
		*fields
		Attachments json.RawMessage `json:"attachments"`
		Transport   json.RawMessage `json:"transport"`
	}{fields: (*fields)(m)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if decoded.Attachments != nil {
		if raw := bytes.TrimSpace(decoded.Attachments); len(raw) == 0 || raw[0] != '[' {
			return errors.New("ipc: attachments must be an array")
		}
		if err := json.Unmarshal(decoded.Attachments, &m.Attachments); err != nil {
			return err
		}
	}
	if decoded.Transport != nil {
		if raw := bytes.TrimSpace(decoded.Transport); len(raw) == 0 || raw[0] != '{' {
			return errors.New("ipc: transport metadata must be an object")
		}
		if err := json.Unmarshal(decoded.Transport, &m.Transport); err != nil {
			return err
		}
	}
	return nil
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
	if length > MaxFrameSize {
		return nil, fmt.Errorf(
			"ipc: frame length %d exceeds cap %d (likely stream desync)",
			length, MaxFrameSize)
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
	body, err := encodeFrame(msg)
	if err != nil {
		return err
	}
	return writeEncodedFrame(w, body)
}

func encodeFrame(msg *Message) ([]byte, error) {
	if err := preflightAttachments(msg); err != nil {
		return nil, err
	}
	body, err := marshalJSON(msg)
	if err != nil {
		return nil, fmt.Errorf("ipc: encode frame (kind=%s): %w", msg.Kind, err)
	}
	if len(body) > MaxFrameSize {
		return nil, fmt.Errorf("ipc: frame body %d exceeds cap %d", len(body), MaxFrameSize)
	}
	return body, nil
}

// preflightAttachments rejects oversized inline values before JSON escaping or
// base64 encoding allocates a complete frame. The application payload is opaque
// and can contain JSON whitespace, so this is a lower bound excluding Data;
// encodeFrame checks the exact final size as well.
func preflightAttachments(msg *Message) error {
	if len(msg.Attachments) == 0 {
		return nil
	}
	skeleton := *msg
	skeleton.Data = nil
	skeleton.Attachments = append([]AttachmentData(nil), msg.Attachments...)
	emptyText, emptyBytes := "", []byte{}
	for i := range skeleton.Attachments {
		attachment := &skeleton.Attachments[i]
		if attachment.Text != nil {
			attachment.Text = &emptyText
		}
		if attachment.Bytes != nil {
			attachment.Bytes = &emptyBytes
		}
	}
	metadata, err := marshalJSON(&skeleton)
	if err != nil {
		return fmt.Errorf("ipc: encode attachment metadata: %w", err)
	}
	remaining := MaxFrameSize - len(metadata)
	for _, attachment := range msg.Attachments {
		if attachment.Text != nil {
			remaining = consumeJSONText(*attachment.Text, remaining)
		}
		if attachment.Bytes != nil {
			length := len(*attachment.Bytes)
			// Compare before multiplication so even a huge buffer cannot
			// overflow size arithmetic on a 32-bit platform.
			groups := length / 3
			if length%3 != 0 {
				groups++
			}
			if remaining < 0 || groups > remaining/4 {
				remaining = -1
			} else {
				remaining -= groups * 4
			}
		}
		if remaining < 0 {
			return fmt.Errorf("ipc: inline attachments exceed frame cap %d", MaxFrameSize)
		}
	}
	if remaining < 0 {
		return fmt.Errorf("ipc: attachment metadata exceeds frame cap %d", MaxFrameSize)
	}
	return nil
}

// consumeJSONText accounts for string contents as encoding/json emits them
// with HTML escaping disabled. The surrounding quotes are in the skeleton.
func consumeJSONText(value string, remaining int) int {
	for i := 0; i < len(value) && remaining >= 0; {
		c := value[i]
		if c < utf8.RuneSelf {
			switch c {
			case '\\', '"', '\n', '\r', '\t', '\b', '\f':
				remaining -= 2
			default:
				if c < 0x20 {
					remaining -= 6
				} else {
					remaining--
				}
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(value[i:])
		if r == '\u2028' || r == '\u2029' {
			remaining -= 6
		} else if r == utf8.RuneError && size == 1 {
			remaining -= len("�")
		} else {
			remaining -= size
		}
		i += size
	}
	return remaining
}

func writeEncodedFrame(w io.Writer, body []byte) error {
	var header [FrameHeaderSize]byte
	binary.LittleEndian.PutUint32(header[:], uint32(len(body)))
	if err := writeExact(w, header[:]); err != nil {
		return fmt.Errorf("ipc: write frame length: %w", err)
	}
	if err := writeExact(w, body); err != nil {
		return fmt.Errorf("ipc: write frame body: %w", err)
	}
	return nil
}

// writeExact treats any short write as terminal, even when a broken Writer
// returns a nil error. Retrying would be unsafe for framing: the peer has
// already received a prefix, and a Writer returning (0, nil) could otherwise
// spin forever.
func writeExact(w io.Writer, p []byte) error {
	n, err := w.Write(p)
	if err != nil {
		return err
	}
	if n != len(p) {
		return io.ErrShortWrite
	}
	return nil
}
