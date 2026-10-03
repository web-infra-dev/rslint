package ipc

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// Attachment is a complete value whose transport is chosen by Channel. Its
// contents must remain immutable until SendRequest returns. Strings retain
// their existing storage; byte buffers are borrowed rather than copied here.
type Attachment struct {
	text   string
	bytes  []byte
	binary bool
}

// Text supplies Unicode text. As with Go JSON string encoding, each invalid
// UTF-8 rune becomes U+FFFD. Valid text keeps its original string storage, and
// normalization happens once before choosing shared or inline transmission.
func Text(value string) Attachment {
	if !utf8.ValidString(value) {
		var normalized strings.Builder
		normalized.Grow(len(value))
		for _, r := range value {
			normalized.WriteRune(r)
		}
		value = normalized.String()
	}
	return Attachment{text: value}
}

// Bytes supplies arbitrary binary data, including malformed UTF-8 and NULs.
// The caller must not change the buffer until SendRequest returns.
func Bytes(value []byte) Attachment { return Attachment{bytes: value, binary: true} }

func (a Attachment) length() int {
	if a.binary {
		return len(a.bytes)
	}
	return len(a.text)
}

func (a Attachment) copyTo(dst []byte, offset int) int {
	if a.binary {
		return copy(dst, a.bytes[offset:])
	}
	return copy(dst, a.text[offset:])
}

func (a Attachment) inline() AttachmentData {
	if a.binary {
		value := a.bytes
		if value == nil {
			value = []byte{}
		}
		return AttachmentData{Bytes: &value}
	}
	return AttachmentData{Text: &a.text}
}

// The first application request fixes peer capabilities before its handler can
// send attachments. Legacy peers can still provide their inherited mapping;
// capable peers allocate nothing until an outbound attachment needs storage.
func (c *Channel) initializeMemory(transport *TransportMetadata) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return c.closeErr
	}
	if c.memoryInitialized {
		if transport != nil && transport.Mapping != nil {
			return errors.New("ipc: shared memory already initialized")
		}
		return nil
	}
	c.memoryInitialized = true
	c.memoryCapable = transport != nil && transport.SharedMemory == 1 && transport.Mapping == nil
	if transport != nil && transport.Mapping != nil {
		// Shared storage is optional. Unsupported layouts, platforms and
		// allocation failures retain the complete inline transport.
		c.memory, _ = openMemoryPool(*transport.Mapping)
	}
	return nil
}

func attach(msg *Message, memory *memoryPool, values []Attachment) []MemoryBatch {
	if len(values) == 0 {
		return nil
	}
	msg.Attachments = make([]AttachmentData, len(values))
	for i, value := range values {
		msg.Attachments[i] = value.inline()
	}
	if memory == nil {
		return nil
	}
	batches, ranges := memory.store(values)
	if len(batches) == 0 {
		return nil
	}
	for i, span := range ranges {
		if span != nil {
			msg.Attachments[i] = AttachmentData{Range: span}
		}
	}
	msg.Transport = &TransportMetadata{Batches: batches}
	return batches
}
